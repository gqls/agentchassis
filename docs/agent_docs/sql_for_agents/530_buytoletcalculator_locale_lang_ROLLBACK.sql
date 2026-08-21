-- 530_buytoletcalculator_locale_lang_ROLLBACK.sql
-- Reverses 530. Removes locale.lang from buytoletcalculator.uk, and removes the
-- row entirely only if 530 created it AND it now holds nothing else — a pipeline
-- writer may legitimately have added keys since.
BEGIN;

UPDATE site_specs ss SET
  data = CASE WHEN ((ss.data -> 'locale') - 'lang') = '{}'::jsonb
              THEN ss.data - 'locale'
              ELSE jsonb_set(ss.data, '{locale}', (ss.data -> 'locale') - 'lang') END,
  updated_at = now()
FROM sites s
WHERE ss.site_id = s.id AND s.domain = 'buytoletcalculator.uk'
  AND ss.aspect = 'site_config' AND ss.is_current
  AND ss.data #> '{locale,lang}' IS NOT NULL;

DELETE FROM site_specs ss USING sites s
WHERE ss.site_id = s.id AND s.domain = 'buytoletcalculator.uk'
  AND ss.aspect = 'site_config' AND ss.is_current
  AND ss.created_by = 'migration-530-locale-lang'
  AND ss.data = '{}'::jsonb;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain = 'buytoletcalculator.uk' AND ss.aspect = 'site_config'
    AND ss.is_current AND ss.data #> '{locale,lang}' IS NOT NULL;
  IF n <> 0 THEN RAISE EXCEPTION 'rollback left a locale.lang in place'; END IF;
END $$;

COMMIT;
