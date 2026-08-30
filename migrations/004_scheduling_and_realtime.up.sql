-- 004_scheduling_and_realtime.up.sql
--
-- Bu migration "onceden generasiya olunmus slot setirleri" modelinden
-- "qaydaya esasen runtime-da hesablanan bosluq" modeline kecidi qurur:
--
--   working_hours   -> hansi gun, saat necede-necede + nahar fasilesi
--   schedule_settings -> secim addimi (16/30/60 deq), bufer, min xeberdarliq
--   time_off        -> mezuniyyet / bloklanmis interval
--   bookings        -> artiq slot setirine baglı deyil (slot_id nullable)
--
-- Elave olaraq realtime ucun notifications ve device_tokens cedvelleri.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================================
-- WORKING HOURS: nahar fasilesi sutunlari
-- =====================================================================
ALTER TABLE working_hours ADD COLUMN IF NOT EXISTS break_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE working_hours ADD COLUMN IF NOT EXISTS break_start   VARCHAR(5);
ALTER TABLE working_hours ADD COLUMN IF NOT EXISTS break_end     VARCHAR(5);

-- Her staff ucun her gunden yalniz bir setir olsun
CREATE UNIQUE INDEX IF NOT EXISTS ux_working_hours_staff_day
    ON working_hours (business_id, staff_id, day_of_week);

-- =====================================================================
-- SCHEDULE SETTINGS
-- staff_id NULL  -> biznesin default ayari
-- staff_id dolu  -> hemin isciye mexsus override
-- =====================================================================
CREATE TABLE IF NOT EXISTS schedule_settings (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id               UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    staff_id                  UUID NULL REFERENCES staff_profiles(id) ON DELETE CASCADE,

    timezone                  VARCHAR(64)  NOT NULL DEFAULT 'Asia/Baku',

    -- Musteriye gosterilen vaxtlarin addimi. 16 deqiqe de ola biler.
    slot_step_mins            INTEGER NOT NULL DEFAULT 30 CHECK (slot_step_mins BETWEEN 5 AND 480),
    -- Xidmet secilmeyibse istifade olunan default mueddet
    default_duration_mins     INTEGER NOT NULL DEFAULT 30 CHECK (default_duration_mins BETWEEN 5 AND 1440),

    -- Randevudan evvel/sonra saxlanilan bos vaxt (temizlik, hazirliq)
    buffer_before_mins        INTEGER NOT NULL DEFAULT 0 CHECK (buffer_before_mins BETWEEN 0 AND 240),
    buffer_after_mins         INTEGER NOT NULL DEFAULT 0 CHECK (buffer_after_mins  BETWEEN 0 AND 240),

    -- En azi bu qeder deqiqe evvelceden bron edile biler
    min_notice_mins           INTEGER NOT NULL DEFAULT 60 CHECK (min_notice_mins BETWEEN 0 AND 43200),
    -- Ne qeder ireli tarixe bron acilsin
    max_advance_days          INTEGER NOT NULL DEFAULT 30 CHECK (max_advance_days BETWEEN 1 AND 365),

    -- TRUE olsa booking birbasa "confirmed" yaranir, hekim tesdiqi gozlenmir
    auto_confirm              BOOLEAN NOT NULL DEFAULT FALSE,
    -- Provider musteriye alternativ vaxt teklif ede bilsin?
    allow_reschedule_proposal BOOLEAN NOT NULL DEFAULT TRUE,

    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- staff_id dolu olanda (business_id, staff_id) tekrarlanmasin
CREATE UNIQUE INDEX IF NOT EXISTS ux_schedule_settings_staff
    ON schedule_settings (business_id, staff_id)
    WHERE staff_id IS NOT NULL;

-- Biznes default-u yalniz bir dene olsun
CREATE UNIQUE INDEX IF NOT EXISTS ux_schedule_settings_business_default
    ON schedule_settings (business_id)
    WHERE staff_id IS NULL;

-- =====================================================================
-- TIME OFF (mezuniyyet, xestelik, bloklanmis interval)
-- =====================================================================
CREATE TABLE IF NOT EXISTS time_off (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    staff_id    UUID NOT NULL REFERENCES staff_profiles(id) ON DELETE CASCADE,
    start_at    TIMESTAMPTZ NOT NULL,
    end_at      TIMESTAMPTZ NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT time_off_valid_range CHECK (end_at > start_at)
);

CREATE INDEX IF NOT EXISTS idx_time_off_staff_range
    ON time_off (business_id, staff_id, start_at, end_at);

-- =====================================================================
-- BOOKINGS: slot setirinden asililigi qirmaq + teklif (proposal) axini
-- =====================================================================
ALTER TABLE bookings ALTER COLUMN slot_id DROP NOT NULL;

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS location_id          UUID NULL REFERENCES locations(id) ON DELETE SET NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS duration_mins        INTEGER NOT NULL DEFAULT 30;

-- Provider alternativ vaxt teklif edende doldurulur
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS proposed_start_time  TIMESTAMPTZ NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS proposed_end_time    TIMESTAMPTZ NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS proposed_by          UUID NULL REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS proposal_note        TEXT NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS proposed_at          TIMESTAMPTZ NULL;

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancel_reason        TEXT NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancelled_by         UUID NULL REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS confirmed_at         TIMESTAMPTZ NULL;

-- Yeni statuslar: reschedule_proposed (provider basqa vaxt teklif etdi), no_show
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('pending', 'confirmed', 'reschedule_proposed', 'cancelled', 'completed', 'no_show'));

CREATE INDEX IF NOT EXISTS idx_bookings_staff_time      ON bookings (business_id, staff_id, start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_customer_time   ON bookings (business_id, customer_id, start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_status_start    ON bookings (status, start_time);

-- Eyni isci ucun ust-uste dusen aktiv booking-in DB seviyyesinde qarsisini alir.
-- Iki musteri eyni anda "Bron et" basdiqda ikincisi burada rədd olunur.
DO $overlap_guard$
BEGIN
    BEGIN
        CREATE EXTENSION IF NOT EXISTS btree_gist;
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE 'btree_gist elave edile bilmedi: %', SQLERRM;
    END;

    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'btree_gist')
       AND NOT EXISTS (
           SELECT 1 FROM pg_constraint WHERE conname = 'bookings_no_overlap'
       ) THEN
        ALTER TABLE bookings ADD CONSTRAINT bookings_no_overlap
            EXCLUDE USING gist (
                staff_id WITH =,
                tstzrange(start_time, end_time) WITH &&
            ) WHERE (status IN ('pending', 'confirmed', 'reschedule_proposed'));
    END IF;
END
$overlap_guard$;

-- =====================================================================
-- NOTIFICATIONS (in-app bildiris merkezi + realtime feed)
-- =====================================================================
CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NULL REFERENCES businesses(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id  UUID NULL REFERENCES bookings(id) ON DELETE CASCADE,
    type        VARCHAR(64) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    read_at     TIMESTAMPTZ NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications (user_id, is_read, created_at DESC);

-- =====================================================================
-- DEVICE TOKENS (Firebase Cloud Messaging push ucun)
-- =====================================================================
CREATE TABLE IF NOT EXISTS device_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    platform   VARCHAR(16) NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens (user_id, is_active);

-- =====================================================================
-- OUTBOX: worker bunu oxuyub push/email gonderir ve LISTEN/NOTIFY atir
-- =====================================================================
CREATE TABLE IF NOT EXISTS notification_outbox (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id   UUID NULL REFERENCES bookings(id) ON DELETE CASCADE,
    channel      VARCHAR(16) NOT NULL CHECK (channel IN ('push', 'email', 'ws')),
    type         VARCHAR(64) NOT NULL,
    payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
    status       VARCHAR(16) NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'sent', 'failed')),
    attempts     INTEGER NOT NULL DEFAULT 0,
    last_error   TEXT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at      TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON notification_outbox (status, scheduled_at)
    WHERE status = 'pending';
