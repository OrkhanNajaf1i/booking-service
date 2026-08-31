-- 011_backfill_business_category.up.sql
--
-- Kohne bizneslere kateqoriya slug-i yazir.
--
-- 010-dan sonra yeni biznes yaradarken kateqoriya secimi MECBURIDIR,
-- amma ondan evvelki setirlerde `category_slug` bosdur. Kod onlari
-- serbest metnden taniyir; taniya bilmediklerini "Diger"e yigir.
-- "Diger" ise pese deyil — orada biznes qalmamalidir.
--
-- Bu birdefelik backfill-dir. Acar sozlerin ESAS siyahisi Go terefdedir
-- (internal/domain/catalog/category.go); burada onun SQL kopyasi var.
-- Taksonomiyaya yeni sahe elave edende bura toxunmaq lazim deyil —
-- yalniz kohne setirler ucundur.
--
-- SIRA ONEMLIDIR: her UPDATE yalniz hele bos olan setirlere toxunur,
-- ona gore daha xususi sahe (dis hekimi) daha umumidan (hekim) evvel
-- gelir.

-- Muqayise ucun Azerbaycan herflerini ASCII-ye endiren komekci.
CREATE OR REPLACE FUNCTION pg_temp.norm(text) RETURNS text AS $$
    SELECT lower(translate(coalesce($1, ''),
        'əıışçğöüİŞÇĞÖÜƏ',
        'eiiscgouiscgoue'));
$$ LANGUAGE sql IMMUTABLE;

DO $$
DECLARE
    rule RECORD;
BEGIN
    FOR rule IN
        SELECT * FROM (VALUES
            ('dentist',   ARRAY['dis', 'stomatoloq', 'stomatologiya', 'dentist', 'ortodont', 'implant']),
            ('vet',       ARRAY['baytar', 'veterinar', 'heyvan']),
            ('hospital',  ARRAY['xestexana', 'hospital', 'klinika', 'clinic', 'tibb merkezi']),
            ('doctor',    ARRAY['hekim', 'doctor', 'terapevt', 'kardioloq', 'pediatr', 'nevroloq',
                                'ginekoloq', 'uroloq', 'dermatoloq', 'endokrinoloq', 'oftalmoloq',
                                'psixoloq', 'psixiatr', 'cerrah', 'travmatoloq', 'healthcare']),
            ('lab',       ARRAY['laboratoriya', 'analiz']),
            ('barber',    ARRAY['berber', 'barber', 'sac', 'hairdresser']),
            ('beauty',    ARRAY['gozellik', 'beauty', 'salon', 'kosmetoloq', 'manikur',
                                'pedikur', 'makiyaj', 'kirpik', 'epilyasiya']),
            ('spa',       ARRAY['spa', 'masaj', 'massage', 'hamam']),
            ('fitness',   ARRAY['fitnes', 'fitness', 'idman', 'gym', 'yoga', 'pilates', 'treyner']),
            ('education', ARRAY['repetitor', 'muellim', 'kurs', 'tehsil', 'dersler', 'hazirliq']),
            ('photo',     ARRAY['foto', 'photo', 'video', 'studiya']),
            ('nutrition', ARRAY['dietoloq', 'qidalanma', 'pehriz']),
            ('physio',    ARRAY['fizioterapiya', 'reabilitasiya']),
            ('tattoo',    ARRAY['tatu', 'tattoo', 'pirsinq']),
            ('auto',      ARRAY['avtoservis', 'avtoyuma', 'avto', 'diaqnostika']),
            ('master',    ARRAY['usta', 'temir', 'repair', 'santexnik', 'elektrik',
                                'kondisioner', 'mebel', 'qaynaq', 'kombi']),
            ('cleaning',  ARRAY['temizlik', 'cleaning', 'yigisdirma']),
            ('legal',     ARRAY['huquq', 'vekil', 'notarius', 'konsultasiya', 'mesletci']),
            ('event',     ARRAY['tedbir', 'senlik', 'toy', 'aparici'])
        ) AS t(slug, keywords)
    LOOP
        UPDATE businesses
        SET category_slug = rule.slug
        WHERE category_slug IS NULL
          AND EXISTS (
              SELECT 1 FROM unnest(rule.keywords) AS keyword
              WHERE pg_temp.norm(service_category) LIKE '%' || keyword || '%'
                 OR pg_temp.norm(industry) LIKE '%' || keyword || '%'
          );
    END LOOP;
END $$;

-- Qalanlar hec bir acar soza uymur. Onlari "Diger"e atmaq yerine
-- "Usta ve temir"e yazmaq da dogru olmazdi — hansi sahe oldugu
-- bilinmir. Bos qalirlar: kod onlari "Diger" kimi gosterir, sahib
-- ayarlardan duzeldir. Say az oldugundan bu qebul edilebilendir.
