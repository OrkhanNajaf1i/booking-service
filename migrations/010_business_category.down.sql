-- 010_business_category.down.sql

DROP INDEX IF EXISTS idx_businesses_category_slug;

ALTER TABLE businesses
    DROP COLUMN IF EXISTS category_slug;
