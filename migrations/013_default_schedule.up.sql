-- 013_default_schedule.up.sql
--
-- Yeni biznesin qrafiki bos qalirdi.
--
-- Bos vaxtlar sorgu aninda `working_hours` + `schedule_settings`
-- uzerinden hesablanir. Biznes yaradilanda bu setirler yaranmirdi, ona
-- gore tetbiqde HER gun ucun "Bu gun is gunu deyil" cixirdi: yeni
-- yaradilmis biznes musteri terefinde sinmis gorunurdu.
--
-- Bu miqrasiya movcud setirleri duzeldir; yeni bizneslere default
-- qrafik kod terefinde yaradilir (ScheduleProvisioner).
--
-- Default: Bazar ertesi–Cume 09:00–18:00, nahar 13:00–14:00.
-- Hefte sonu setri de yaranir, amma `is_active = FALSE` — sahib
-- qrafik ekraninda yeddi gunu de gorup istediyini acmalidir.
--
-- Yalniz HEC bir is saati olmayan iscilere toxunur: sahibin oz
-- qurdugu qrafik ustunden yazilmamalidir.

INSERT INTO working_hours (
    business_id, staff_id, day_of_week,
    start_time, end_time,
    break_enabled, break_start, break_end,
    is_active
)
SELECT
    sp.business_id,
    sp.id,
    d.day_of_week,
    '09:00',
    '18:00',
    TRUE,
    '13:00',
    '14:00',
    -- 0 = Bazar, 6 = Senbe
    d.day_of_week BETWEEN 1 AND 5
FROM staff_profiles sp
CROSS JOIN generate_series(0, 6) AS d(day_of_week)
WHERE sp.status = 'active'
  AND NOT EXISTS (
      SELECT 1 FROM working_hours w WHERE w.staff_id = sp.id
  );

-- Ayarlar cedvelinin butun sutunlari default dasiyir, ona gore
-- yalniz setrin ozu lazimdir.
INSERT INTO schedule_settings (business_id, staff_id)
SELECT sp.business_id, sp.id
FROM staff_profiles sp
WHERE sp.status = 'active'
  AND NOT EXISTS (
      SELECT 1 FROM schedule_settings s WHERE s.staff_id = sp.id
  );
