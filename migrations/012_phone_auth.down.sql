-- 012_phone_auth.down.sql

DROP TABLE IF EXISTS phone_verifications;

DROP INDEX IF EXISTS users_phone_e164_verified_key;
DROP INDEX IF EXISTS idx_users_phone_e164;

ALTER TABLE users
    DROP COLUMN IF EXISTS phone_verified,
    DROP COLUMN IF EXISTS phone_e164;
