-- 006_location_phone.up.sql
--
-- location domeni ve repository-si `phone` sutunu ile isleyir
-- (Location.Phone, CreateLocationRequest.Phone, SELECT ... phone),
-- lakin 001 migration-u onu yaratmamisdi. Neticede GET /locations
-- "column phone does not exist" ile 500 qaytarirdi.

ALTER TABLE locations ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
