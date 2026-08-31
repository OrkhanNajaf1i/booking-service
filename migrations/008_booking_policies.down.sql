-- 008_booking_policies.down.sql

ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_customer_no_overlap;

ALTER TABLE schedule_settings
    DROP COLUMN IF EXISTS reschedule_window_mins,
    DROP COLUMN IF EXISTS allow_customer_reschedule,
    DROP COLUMN IF EXISTS cancellation_window_mins,
    DROP COLUMN IF EXISTS pending_expires_mins;
