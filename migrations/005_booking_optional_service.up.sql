-- 005_booking_optional_service.up.sql
--
-- 002-de bookings.service_id NOT NULL yaradilmisdi, cunki hemin dizaynda
-- randevunun mueddeti yalniz xidmetden gelirdi.
--
-- Yeni modelde xidmet secmek mecburi deyil: secilmese schedule_settings
-- icindeki default_duration_mins islenir (mes. berbere "sac" demeden
-- sadece 30 deqiqelik vaxt tutmaq). Ona gore sutun nullable olur.

ALTER TABLE bookings ALTER COLUMN service_id DROP NOT NULL;
