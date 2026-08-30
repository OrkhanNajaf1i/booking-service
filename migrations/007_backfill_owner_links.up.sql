-- 007_backfill_owner_links.up.sql
--
-- CreateBusiness biznesi yaradib owner_id-ni yazirdi, lakin iki seyi
-- etmirdi:
--
--   1) users.business_id-ni yeniləmirdi → JWT-ye business_id dusmurdu →
--      /bookings, /staff, /availability kimi butun biznes-kontekstli
--      endpoint-ler 400 qaytarirdi
--   2) sahib ucun staff_profiles setri yaratmirdi → "evvelce isci elave
--      edin" ekranindan kenara cixmaq mumkun olmurdu
--
-- Kod artiq duzeldilib; bu migration hemin sehvle yaradilmis movcud
-- hesablari berpa edir.

-- ============================================================
-- 1) Sahibi biznese bagla
-- ============================================================
UPDATE users u
SET business_id = b.id,
    is_owner    = TRUE,
    updated_at  = NOW()
FROM businesses b
WHERE b.owner_id = u.id
  AND u.business_id IS DISTINCT FROM b.id;

-- ============================================================
-- 2) Sahib ucun staff profili yarat (yoxdursa)
-- ============================================================
-- Basliq ucun service_category, o bos olsa industry islenir.
INSERT INTO staff_profiles (
    id, user_id, business_id, role, title, department, bio, status,
    joined_at, created_at, updated_at
)
SELECT
    gen_random_uuid(),
    b.owner_id,
    b.id,
    'admin',
    COALESCE(NULLIF(b.service_category, ''), NULLIF(b.industry, ''), 'Mutexessis'),
    '', '', 'active',
    NOW(), NOW(), NOW()
FROM businesses b
WHERE b.owner_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM staff_profiles sp
      WHERE sp.business_id = b.id AND sp.user_id = b.owner_id
  );
