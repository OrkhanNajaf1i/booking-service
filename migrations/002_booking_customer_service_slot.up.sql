-- UUID generator
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================
-- Customers (multi-tenant)
-- =========================
CREATE TABLE IF NOT EXISTS customers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    businessid    UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    userid        UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    fullname      VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    phone         VARCHAR(50) NOT NULL,
    notes         TEXT NULL,
    status        VARCHAR(50) NOT NULL CHECK (status IN ('active', 'inactive', 'blocked')),
    totalbookings INTEGER NOT NULL DEFAULT 0,
    lastbookingat TIMESTAMPTZ NULL,
    createdat     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_customers_business_email ON customers(businessid, email);
CREATE INDEX IF NOT EXISTS idx_customers_businessid ON customers(businessid);
CREATE INDEX IF NOT EXISTS idx_customers_userid ON customers(userid);

-- =========================
-- Services (multi-tenant)
-- =========================
CREATE TABLE IF NOT EXISTS services (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    businessid       UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    durationminutes  INTEGER NOT NULL,
    price            NUMERIC(12,2) NOT NULL,
    isactive         BOOLEAN NOT NULL DEFAULT TRUE,
    createdat        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_services_businessid ON services(businessid);

-- =========================
-- Staff <-> Service M2M
-- =========================
CREATE TABLE IF NOT EXISTS staffservices (
    staffid    UUID NOT NULL REFERENCES staff_profiles(id) ON DELETE CASCADE,
    businessid UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    serviceid  UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    createdat  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (businessid, staffid, serviceid)
);

CREATE INDEX IF NOT EXISTS idx_staffservices_staffid ON staffservices(staffid);
CREATE INDEX IF NOT EXISTS idx_staffservices_serviceid ON staffservices(serviceid);

-- =========================
-- Staff Working Hours
-- =========================
CREATE TABLE IF NOT EXISTS staffworkinghours (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    businessid UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    staffid    UUID NOT NULL REFERENCES staff_profiles(id) ON DELETE CASCADE,
    dayofweek  INTEGER NOT NULL CHECK (dayofweek BETWEEN 0 AND 6),
    starttime  VARCHAR(5) NOT NULL, -- "HH:MM"
    endtime    VARCHAR(5) NOT NULL, -- "HH:MM"
    isactive   BOOLEAN NOT NULL DEFAULT TRUE,
    createdat  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_staffworkinghours_business_staff ON staffworkinghours(businessid, staffid);

-- =========================
-- Slots (Circular FK bookingid sonra elave olunur)
-- =========================
CREATE TABLE IF NOT EXISTS slots (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    businessid    UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    staffid       UUID NOT NULL REFERENCES staff_profiles(id) ON DELETE CASCADE,
    locationid    UUID NULL REFERENCES locations(id) ON DELETE SET NULL,
    starttime     TIMESTAMPTZ NOT NULL,
    endtime       TIMESTAMPTZ NOT NULL,
    durationmins  INTEGER NOT NULL,
    status        VARCHAR(50) NOT NULL CHECK (status IN ('available', 'booked', 'unavailable')),
    bookingid     UUID NULL, 
    notes         TEXT NULL,
    createdat     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deletedat     TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_slots_business_staff_time ON slots(businessid, staffid, starttime);
CREATE INDEX IF NOT EXISTS idx_slots_business_status_time ON slots(businessid, status, starttime);
CREATE INDEX IF NOT EXISTS idx_slots_locationid ON slots(locationid);

-- =========================
-- Bookings
-- =========================
CREATE TABLE IF NOT EXISTS bookings (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    businessid UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    customerid UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    staffid    UUID NOT NULL REFERENCES staff_profiles(id) ON DELETE RESTRICT,
    serviceid  UUID NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    slotid     UUID NOT NULL REFERENCES slots(id) ON DELETE RESTRICT,
    starttime  TIMESTAMPTZ NOT NULL,
    endtime    TIMESTAMPTZ NOT NULL,
    status     VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    notes      TEXT NULL,
    createdat  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_bookings_slotid ON bookings(slotid);
CREATE INDEX IF NOT EXISTS idx_bookings_business_status ON bookings(businessid, status);
CREATE INDEX IF NOT EXISTS idx_bookings_business_createdat ON bookings(businessid, createdat);


ALTER TABLE slots
    ADD CONSTRAINT fk_slots_bookingid
    FOREIGN KEY (bookingid) REFERENCES bookings(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_slots_bookingid ON slots(bookingid);
