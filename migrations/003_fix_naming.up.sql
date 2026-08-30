-- 003_fix_naming.up.sql
-- 002 migration-u cedvel ve sutun adlarini altxettsiz yaratmisdi
-- (staffservices / businessid / starttime ...), repository SQL-leri ise
-- snake_case gozleyir (staff_services / business_id / start_time ...).
-- Bu migration adlari repo-larin gozledigi formaya getirir. Idempotentdir.

-- ============ CEDVEL ADLARI ============
DO $rename_tables$
BEGIN
    IF to_regclass('public.staffservices') IS NOT NULL
       AND to_regclass('public.staff_services') IS NULL THEN
        ALTER TABLE staffservices RENAME TO staff_services;
    END IF;

    IF to_regclass('public.staffworkinghours') IS NOT NULL
       AND to_regclass('public.working_hours') IS NULL THEN
        ALTER TABLE staffworkinghours RENAME TO working_hours;
    END IF;
END
$rename_tables$;

-- ============ SUTUN ADLARI ============
DO $rename_columns$
DECLARE
    rename_map CONSTANT text[][] := ARRAY[
        ['customers','businessid','business_id'],
        ['customers','userid','user_id'],
        ['customers','fullname','full_name'],
        ['customers','totalbookings','total_bookings'],
        ['customers','lastbookingat','last_booking_at'],
        ['customers','createdat','created_at'],
        ['customers','updatedat','updated_at'],

        ['services','businessid','business_id'],
        ['services','durationminutes','duration_minutes'],
        ['services','isactive','is_active'],
        ['services','createdat','created_at'],
        ['services','updatedat','updated_at'],

        ['staff_services','staffid','staff_id'],
        ['staff_services','businessid','business_id'],
        ['staff_services','serviceid','service_id'],
        ['staff_services','createdat','created_at'],

        ['working_hours','businessid','business_id'],
        ['working_hours','staffid','staff_id'],
        ['working_hours','dayofweek','day_of_week'],
        ['working_hours','starttime','start_time'],
        ['working_hours','endtime','end_time'],
        ['working_hours','isactive','is_active'],
        ['working_hours','createdat','created_at'],
        ['working_hours','updatedat','updated_at'],

        ['slots','businessid','business_id'],
        ['slots','staffid','staff_id'],
        ['slots','locationid','location_id'],
        ['slots','starttime','start_time'],
        ['slots','endtime','end_time'],
        ['slots','durationmins','duration_mins'],
        ['slots','bookingid','booking_id'],
        ['slots','createdat','created_at'],
        ['slots','updatedat','updated_at'],
        ['slots','deletedat','deleted_at'],

        ['bookings','businessid','business_id'],
        ['bookings','customerid','customer_id'],
        ['bookings','staffid','staff_id'],
        ['bookings','serviceid','service_id'],
        ['bookings','slotid','slot_id'],
        ['bookings','starttime','start_time'],
        ['bookings','endtime','end_time'],
        ['bookings','createdat','created_at'],
        ['bookings','updatedat','updated_at']
    ];
    i int;
    tbl text;
    old_col text;
    new_col text;
BEGIN
    FOR i IN 1 .. array_length(rename_map, 1) LOOP
        tbl     := rename_map[i][1];
        old_col := rename_map[i][2];
        new_col := rename_map[i][3];

        IF to_regclass('public.' || tbl) IS NULL THEN
            CONTINUE;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = tbl AND column_name = old_col
        ) AND NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = tbl AND column_name = new_col
        ) THEN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN %I TO %I', tbl, old_col, new_col);
        END IF;
    END LOOP;
END
$rename_columns$;
