-- scripts/seed_dev.sql
--
-- Lokal test ucun numune data. YALNIZ dev bazasinda islet.
--
-- Yaradilanlar:
--   • 1 biznes         "Saglam Klinika"
--   • 1 owner/hekim    dr@test.az        / Test1234!
--   • 1 musteri        musteri@test.az   / Test1234!
--   • 2 xidmet         Konsultasiya 30dq, Muayine 60dq
--   • Heftelik qrafik  B.e-Cume 09:00-18:00, nahar 13:00-14:00
--   • Qrafik ayari     30 deq addim, auto_confirm = false
--
-- Parol hash-i bcrypt cost 10 ile "Test1234!" ucundur.

BEGIN;

-- Idempotent olsun deye evvelki seed silinir.
DELETE FROM bookings   WHERE business_id IN (SELECT id FROM businesses WHERE name = 'Saglam Klinika');
DELETE FROM businesses WHERE name = 'Saglam Klinika';
DELETE FROM users      WHERE email IN ('dr@test.az', 'musteri@test.az');

-- ============================================================
-- BIZNES + ISTIFADECILER
-- ============================================================

INSERT INTO businesses (id, name, phone, industry, service_category, business_type, is_active)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Saglam Klinika',
    '+994501112233',
    'healthcare',
    'Klinika',
    'multi_staff_business',
    TRUE
);

-- Owner (hekim). is_owner = TRUE, business-e baglidir.
INSERT INTO users (id, business_id, email, password_hash, full_name, phone, role, is_active, is_owner, email_verified)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'dr@test.az',
    '$2a$10$k3DY49Cdv574d7.i7AwiUORQ3ylKQ.8jHgDfHkulisSDJMCgIvoLC',
    'Dr. Elvin Memmedov',
    '+994501112233',
    'provider_owner',
    TRUE, TRUE, TRUE
);

UPDATE businesses
SET owner_id = '22222222-2222-2222-2222-222222222222'
WHERE id = '11111111-1111-1111-1111-111111111111';

-- Musteri. business_id NULL — musteri hec bir biznese aid deyil.
INSERT INTO users (id, business_id, email, password_hash, full_name, phone, role, is_active, is_owner, email_verified)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    NULL,
    'musteri@test.az',
    '$2a$10$k3DY49Cdv574d7.i7AwiUORQ3ylKQ.8jHgDfHkulisSDJMCgIvoLC',
    'Aysel Huseynova',
    '+994557778899',
    'customer',
    TRUE, FALSE, TRUE
);

-- ============================================================
-- LOCATION + STAFF PROFIL
-- ============================================================

INSERT INTO locations (id, business_id, name, address, city, is_active)
VALUES (
    '44444444-4444-4444-4444-444444444444',
    '11111111-1111-1111-1111-111111111111',
    'Merkez filial',
    'Nizami kuc. 12',
    'Baki',
    TRUE
);

-- Hekimin staff profili. Bron bu ID-ye baglanir.
INSERT INTO staff_profiles (id, user_id, business_id, location_id, role, title, department, bio, status)
VALUES (
    '55555555-5555-5555-5555-555555555555',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '44444444-4444-4444-4444-444444444444',
    'admin',
    'Kardioloq',
    'Terapiya',
    '15 illik tecrube',
    'active'
);

-- ============================================================
-- XIDMETLER
-- ============================================================

INSERT INTO services (id, business_id, name, description, duration_minutes, price, is_active)
VALUES
    ('66666666-6666-6666-6666-666666666661',
     '11111111-1111-1111-1111-111111111111',
     'Konsultasiya', 'Ilkin hekim konsultasiyasi', 30, 50.00, TRUE),
    ('66666666-6666-6666-6666-666666666662',
     '11111111-1111-1111-1111-111111111111',
     'Genis muayine', 'Tam kardioloji muayine', 60, 120.00, TRUE);

INSERT INTO staff_services (staff_id, business_id, service_id)
VALUES
    ('55555555-5555-5555-5555-555555555555',
     '11111111-1111-1111-1111-111111111111',
     '66666666-6666-6666-6666-666666666661'),
    ('55555555-5555-5555-5555-555555555555',
     '11111111-1111-1111-1111-111111111111',
     '66666666-6666-6666-6666-666666666662');

-- ============================================================
-- IS QRAFIKI — B.e-Cume 09:00-18:00, nahar 13:00-14:00
-- ============================================================

INSERT INTO working_hours (
    id, business_id, staff_id, day_of_week,
    start_time, end_time, break_enabled, break_start, break_end, is_active
)
SELECT
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    '55555555-5555-5555-5555-555555555555',
    day,
    '09:00', '18:00',
    TRUE, '13:00', '14:00',
    TRUE
FROM generate_series(1, 5) AS day;

-- Senbe qisa gun, nahar fasilesi yoxdur.
INSERT INTO working_hours (
    id, business_id, staff_id, day_of_week,
    start_time, end_time, break_enabled, is_active
)
VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    '55555555-5555-5555-5555-555555555555',
    6, '10:00', '14:00', FALSE, TRUE
);

-- ============================================================
-- QRAFIK AYARLARI
-- ============================================================

INSERT INTO schedule_settings (
    id, business_id, staff_id, timezone,
    slot_step_mins, default_duration_mins,
    buffer_before_mins, buffer_after_mins,
    min_notice_mins, max_advance_days,
    auto_confirm, allow_reschedule_proposal
)
VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',
    '55555555-5555-5555-5555-555555555555',
    'Asia/Baku',
    30,   -- secim addimi
    30,   -- default mueddet
    0, 10,-- randevudan sonra 10 deq bufer
    30,   -- en azi 30 deq evvelceden
    30,   -- 30 gun ireli
    FALSE,-- hekim tesdiqlemelidir (test axini ucun vacib)
    TRUE
);

-- ============================================================
-- MUSTERI KARTI
-- ============================================================
-- Tetbiq bunu POST /customers/self ile ozu de yaradir; burada
-- evvelceden qoyuruq ki, web panelinde de gorunsun.

INSERT INTO customers (id, business_id, user_id, full_name, email, phone, notes, status)
VALUES (
    '77777777-7777-7777-7777-777777777777',
    '11111111-1111-1111-1111-111111111111',
    '33333333-3333-3333-3333-333333333333',
    'Aysel Huseynova',
    'musteri@test.az',
    '+994557778899',
    'Test musterisi',
    'active'
);

COMMIT;

-- ============================================================
SELECT 'Seed hazirdir.' AS status;
SELECT 'Hekim  : dr@test.az / Test1234!' AS login;
SELECT 'Musteri: musteri@test.az / Test1234!' AS login;
