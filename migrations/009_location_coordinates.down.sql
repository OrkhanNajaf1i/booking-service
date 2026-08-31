-- 009_location_coordinates.down.sql
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_coordinates_range;
ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_coordinates_complete;
ALTER TABLE locations DROP COLUMN IF EXISTS longitude;
ALTER TABLE locations DROP COLUMN IF EXISTS latitude;
