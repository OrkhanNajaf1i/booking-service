-- 003_fix_naming.down.sql — adlari 002-deki altxettsiz formaya qaytarir.

DO $rename_tables_down$
BEGIN
    IF to_regclass('public.staff_services') IS NOT NULL
       AND to_regclass('public.staffservices') IS NULL THEN
        ALTER TABLE staff_services RENAME TO staffservices;
    END IF;

    IF to_regclass('public.working_hours') IS NOT NULL
       AND to_regclass('public.staffworkinghours') IS NULL THEN
        ALTER TABLE working_hours RENAME TO staffworkinghours;
    END IF;
END
$rename_tables_down$;

DO $rename_columns_down$
DECLARE
    rename_map CONSTANT text[][] := ARRAY[
        ['customers','business_id','businessid'],
        ['customers','user_id','userid'],
        ['customers','full_name','fullname'],
        ['customers','total_bookings','totalbookings'],
        ['customers','last_booking_at','lastbookingat'],
        ['customers','created_at','createdat'],
        ['customers','updated_at','updatedat'],

        ['services','business_id','businessid'],
        ['services','duration_minutes','durationminutes'],
        ['services','is_active','isactive'],
        ['services','created_at','createdat'],
        ['services','updated_at','updatedat'],

        ['staffservices','staff_id','staffid'],
        ['staffservices','business_id','businessid'],
        ['staffservices','service_id','serviceid'],
        ['staffservices','created_at','createdat'],

        ['staffworkinghours','business_id','businessid'],
        ['staffworkinghours','staff_id','staffid'],
        ['staffworkinghours','day_of_week','dayofweek'],
        ['staffworkinghours','start_time','starttime'],
        ['staffworkinghours','end_time','endtime'],
        ['staffworkinghours','is_active','isactive'],
        ['staffworkinghours','created_at','createdat'],
        ['staffworkinghours','updated_at','updatedat'],

        ['slots','business_id','businessid'],
        ['slots','staff_id','staffid'],
        ['slots','location_id','locationid'],
        ['slots','start_time','starttime'],
        ['slots','end_time','endtime'],
        ['slots','duration_mins','durationmins'],
        ['slots','booking_id','bookingid'],
        ['slots','created_at','createdat'],
        ['slots','updated_at','updatedat'],
        ['slots','deleted_at','deletedat'],

        ['bookings','business_id','businessid'],
        ['bookings','customer_id','customerid'],
        ['bookings','staff_id','staffid'],
        ['bookings','service_id','serviceid'],
        ['bookings','slot_id','slotid'],
        ['bookings','start_time','starttime'],
        ['bookings','end_time','endtime'],
        ['bookings','created_at','createdat'],
        ['bookings','updated_at','updatedat']
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
$rename_columns_down$;
