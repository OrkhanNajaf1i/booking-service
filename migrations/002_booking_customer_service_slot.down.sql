-- File: migrations/002_booking_customer_service_slot.down.sql
ALTER TABLE slots DROP CONSTRAINT IF EXISTS fk_slots_bookingid;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS staffservices;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS customers;
