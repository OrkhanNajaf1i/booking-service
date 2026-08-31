-- 012_phone_auth.up.sql
--
-- Musterinin telefon nomresi ile girisi.
--
-- Musteri e-poct/sifre yerine nomresini yazir, gelen 6 reqemli kodu
-- tesdiqleyir. Bu, tetbiq terefi ucundur — admin panel e-poct/sifre
-- ile qalir (bax: CLAUDE.md, rol ayrimi).

-- ── users ────────────────────────────────────────────────────
--
-- phone_e164: axtaris ucun normallasdirilmis forma. Istifadeci
-- "050 111 22 33", "+994 50 111 22 33" ve ya "994501112233" yaza
-- bilir — hamisi eyni setre dusmelidir.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone_e164     VARCHAR(20),
    ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Kohne setirlerin normallasdirilmasi: reqemlerden basqa her sey
-- atilir, sonra Azerbaycan qaydasina salinir.
UPDATE users
SET phone_e164 =
    CASE
        WHEN regexp_replace(phone, '\D', '', 'g') ~ '^994\d{9}$'
            THEN '+' || regexp_replace(phone, '\D', '', 'g')
        WHEN regexp_replace(phone, '\D', '', 'g') ~ '^0\d{9}$'
            THEN '+994' || substring(regexp_replace(phone, '\D', '', 'g') from 2)
        WHEN regexp_replace(phone, '\D', '', 'g') ~ '^\d{9}$'
            THEN '+994' || regexp_replace(phone, '\D', '', 'g')
        ELSE NULL
    END
WHERE phone IS NOT NULL AND phone <> '' AND phone_e164 IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_phone_e164
    ON users (phone_e164)
    WHERE phone_e164 IS NOT NULL;

-- Unikallıq YALNIZ tesdiqlenmis nomrelere aiddir.
--
-- Movcud data-da eyni nomre onlarla setirde tekrarlanir (kohne test
-- hesablari). Qlobal unikallıq miqrasiyani sindirardi. Tesdiqlenmis
-- setirler ise bu miqrasiyadan sonra yaranir, ona gore serti indeks
-- gelecekde iki hesabin eyni nomreye sahib olmasinin qarsisini alir.
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_e164_verified_key
    ON users (phone_e164)
    WHERE phone_verified AND phone_e164 IS NOT NULL;

-- ── phone_verifications ──────────────────────────────────────
--
-- Kod ACIQ saxlanilmir: baza sizsa, aktiv kodlar da sizerdi.
-- Yalniz SHA-256 hash-i yazilir.
--
-- 6 reqemli kodun fezasi kicikdir (10^6), ona gore esas mudafie
-- hash deyil, cehd limiti ve qisa omurdur.
CREATE TABLE IF NOT EXISTS phone_verifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_e164  VARCHAR(20)  NOT NULL,
    code_hash   CHAR(64)     NOT NULL,
    channel     VARCHAR(16)  NOT NULL,
    attempts    INTEGER      NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ  NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT phone_verifications_channel_check
        CHECK (channel IN ('sms', 'whatsapp', 'log'))
);

-- Ən son kodu tapmaq və saatlıq limiti saymaq üçün.
CREATE INDEX IF NOT EXISTS idx_phone_verifications_lookup
    ON phone_verifications (phone_e164, created_at DESC);
