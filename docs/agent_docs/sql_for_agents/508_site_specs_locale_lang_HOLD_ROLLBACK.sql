-- 508_site_specs_locale_lang_HOLD_ROLLBACK.sql
-- Reverses 508: removes `locale.lang` from every site_config row it set.
--
-- ORDER: run this BEFORE 507's rollback. The reverse leaves the templates
-- carrying a gate with no value, which is inert but reads as unfinished.
--
-- Deliberately surgical. It removes the `lang` KEY, and removes the `locale`
-- object only if that leaves it empty — so a sibling key added under locale
-- after this migration survives. It does not delete rows: the 10 rows 508
-- INSERTed are dropped only if `data` ends up an empty object, because a
-- pipeline writer may legitimately have added keys to them since.
--
-- Nothing served changes until each site's chrome re-renders. After that,
-- assemblePage's `en` default returns every page to the pre-2026-08-20 bytes.

BEGIN;

-- A. Drop the lang key wherever 508 put it.
UPDATE site_specs SET
  data = CASE
           WHEN ((data -> 'locale') - 'lang') = '{}'::jsonb
             THEN data - 'locale'
           ELSE jsonb_set(data, '{locale}', (data -> 'locale') - 'lang')
         END,
  updated_at = now()
WHERE aspect = 'site_config'
  AND is_current
  AND data #> '{locale,lang}' IS NOT NULL;

-- B. Remove the rows 508 created that now hold nothing at all. Anything with
--    surviving content is LEFT — deleting it would take another lane's key.
DELETE FROM site_specs
WHERE aspect = 'site_config'
  AND is_current
  AND created_by = 'migration-508-locale-lang'
  AND data = '{}'::jsonb;

-- C. Assert: no current site_config row declares a language, and the keys this
--    file must not touch are intact.
DO $$
DECLARE n_lang int; n_gtm int;
BEGIN
  SELECT count(*) INTO n_lang FROM site_specs
  WHERE aspect = 'site_config' AND is_current AND data #> '{locale,lang}' IS NOT NULL;
  IF n_lang <> 0 THEN
    RAISE EXCEPTION '% site_config rows still declare locale.lang after rollback', n_lang;
  END IF;

  SELECT count(*) INTO n_gtm FROM site_specs
  WHERE aspect = 'site_config' AND is_current
    AND data #>> '{analytics,gtm_container_id}' IS NOT NULL;
  IF n_gtm <> 14 THEN
    RAISE EXCEPTION 'analytics.gtm_container_id count is %, expected 14 — the rollback took an unrelated key', n_gtm;
  END IF;
END $$;

COMMIT;

-- VERIFY (read-only):
--   SELECT s.domain, ss.data #>> '{locale,lang}' AS lang_should_be_null,
--          ss.data #>> '{analytics,gtm_container_id}' AS gtm_intact
--     FROM site_specs ss JOIN sites s ON s.id = ss.site_id
--    WHERE ss.aspect = 'site_config' AND ss.is_current ORDER BY s.domain;
