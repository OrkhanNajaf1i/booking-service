-- 008_booking_policies.up.sql
--
-- Randevu siyasetleri. Bu deyerler hazira qeder kodda yox idi, ona gore:
--
--   • cavabsiz "pending" bron slotu ebedi bloklayirdi
--   • musteri randevudan 5 deqiqe evvel de legv ede bilirdi
--   • musteri oz vaxtini deyise bilmirdi (yalniz provider teklif edirdi)
--
-- Deyerler biznes uzre teyin olunur: dis hekimi ile berberin ehtiyaci
-- eyni deyil. Default-lar sahe standartidir (Booksy/Fresha: 24 saat).

ALTER TABLE schedule_settings
    -- Provider bu muddet erzinde cavab vermese bron avtomatik legv olunur
    -- ve slot azad olur. 0 = avtomatik legv sondurulub.
    ADD COLUMN IF NOT EXISTS pending_expires_mins INTEGER NOT NULL DEFAULT 1440
        CHECK (pending_expires_mins BETWEEN 0 AND 20160),

    -- Randevuya bu qeder vaxt qalanda musteri artiq legv ede bilmez.
    -- 0 = istenilen vaxt legv etmek olar.
    ADD COLUMN IF NOT EXISTS cancellation_window_mins INTEGER NOT NULL DEFAULT 1440
        CHECK (cancellation_window_mins BETWEEN 0 AND 20160),

    -- Musteri ozu vaxti deyise bilsin?
    ADD COLUMN IF NOT EXISTS allow_customer_reschedule BOOLEAN NOT NULL DEFAULT TRUE,

    -- Vaxt deyismek ucun son muddet (legv penceresi ile eyni mentiq).
    ADD COLUMN IF NOT EXISTS reschedule_window_mins INTEGER NOT NULL DEFAULT 1440
        CHECK (reschedule_window_mins BETWEEN 0 AND 20160);

-- ============================================================
-- Musterinin ust-uste dusen bronlarinin qarsisini alan indeks
-- ============================================================
-- Eyni musterinin eyni anda iki randevusu ola bilmez: bos yere gedilen
-- randevu hem biznesi, hem de novbede duran diger musterileri zerere
-- salir. Bu, no-show-un esas sebeblerindendir.
--
-- customer_id biznes uzre unikaldir, ona gore eyni sexsin ferqli
-- bizneslerdeki kartlari ayri-ayri sayilir; tam qorunma ucun users
-- seviyyesinde yoxlama servis qatinda aparilir.
DO $customer_overlap$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'btree_gist')
       AND NOT EXISTS (
           SELECT 1 FROM pg_constraint WHERE conname = 'bookings_customer_no_overlap'
       ) THEN
        ALTER TABLE bookings ADD CONSTRAINT bookings_customer_no_overlap
            EXCLUDE USING gist (
                customer_id WITH =,
                tstzrange(start_time, end_time) WITH &&
            ) WHERE (status IN ('pending', 'confirmed', 'reschedule_proposed'));
    END IF;
END
$customer_overlap$;
