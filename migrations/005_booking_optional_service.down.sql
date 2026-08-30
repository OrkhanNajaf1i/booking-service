-- 005_booking_optional_service.down.sql
--
-- Geri qaytarmadan evvel xidmetsiz bronlar temizlenmelidir,
-- eks halda NOT NULL qoyula bilmez.

DELETE FROM bookings WHERE service_id IS NULL;

ALTER TABLE bookings ALTER COLUMN service_id SET NOT NULL;
