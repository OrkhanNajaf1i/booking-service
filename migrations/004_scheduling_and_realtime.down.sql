-- 004_scheduling_and_realtime.down.sql

DROP TABLE IF EXISTS notification_outbox;
DROP TABLE IF EXISTS device_tokens;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS time_off;
DROP TABLE IF EXISTS schedule_settings;

ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_no_overlap;

DROP INDEX IF EXISTS idx_bookings_staff_time;
DROP INDEX IF EXISTS idx_bookings_customer_time;
DROP INDEX IF EXISTS idx_bookings_status_start;

ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed'));

ALTER TABLE bookings DROP COLUMN IF EXISTS confirmed_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS cancelled_by;
ALTER TABLE bookings DROP COLUMN IF EXISTS cancel_reason;
ALTER TABLE bookings DROP COLUMN IF EXISTS proposed_at;
ALTER TABLE bookings DROP COLUMN IF EXISTS proposal_note;
ALTER TABLE bookings DROP COLUMN IF EXISTS proposed_by;
ALTER TABLE bookings DROP COLUMN IF EXISTS proposed_end_time;
ALTER TABLE bookings DROP COLUMN IF EXISTS proposed_start_time;
ALTER TABLE bookings DROP COLUMN IF EXISTS duration_mins;
ALTER TABLE bookings DROP COLUMN IF EXISTS location_id;

DROP INDEX IF EXISTS ux_working_hours_staff_day;
ALTER TABLE working_hours DROP COLUMN IF EXISTS break_end;
ALTER TABLE working_hours DROP COLUMN IF EXISTS break_start;
ALTER TABLE working_hours DROP COLUMN IF EXISTS break_enabled;
