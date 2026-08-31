-- 009_location_coordinates.up.sql
--
-- Filialin xerite uzerinde secilmesi ucun koordinatlar.
-- Musteri "harada yerlesir?" sualina cavab tapmalidir; ünvan metni
-- kifayet etmir, cunki adres yazilisi qeyri-dəqiq ola bilir.
--
-- NUMERIC secilib (float deyil): koordinat deqiqliyi itmemelidir.
-- 8,6 → ~11 sm deqiqlik, kifayetdir.

ALTER TABLE locations
    ADD COLUMN IF NOT EXISTS latitude  NUMERIC(10, 7),
    ADD COLUMN IF NOT EXISTS longitude NUMERIC(10, 7);

-- Koordinat ya tam verilir, ya da hec verilmir — yarimciq deyer
-- xeritede yanlis noqte gosterer.
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_coordinates_complete;
ALTER TABLE locations ADD CONSTRAINT locations_coordinates_complete
    CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude IS NOT NULL AND longitude IS NOT NULL)
    );

ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_coordinates_range;
ALTER TABLE locations ADD CONSTRAINT locations_coordinates_range
    CHECK (
        latitude IS NULL
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    );
