-- 010_business_category.up.sql
--
-- Biznesin sabit kateqoriyasi.
--
-- Indiyedek qruplasdirma `service_category` serbest metnine gore
-- gedirdi: "Dis Hekimi", "dis hekimi", "Stomatoloq" ucu de ayri
-- kateqoriya kimi gorunurdu. Biznes sayi artdiqca kesf ekrani eyni
-- sahenin onlarla variantina bolunecekdi.
--
-- Sutun NULLABLE-dir ve backfill edilmir: kohne setirler ucun kod
-- serbest metni acar sozlerle taniyir (domain/catalog.Resolve).
-- Sahib formada kateqoriya secen kimi burasi dolur ve tanima
-- tehmininden ustun tutulur.

ALTER TABLE businesses
    ADD COLUMN IF NOT EXISTS category_slug VARCHAR(32);

-- Kesf sorgusu bu sutun uzre suzur.
CREATE INDEX IF NOT EXISTS idx_businesses_category_slug
    ON businesses (category_slug)
    WHERE category_slug IS NOT NULL;
