-- 530_buytoletcalculator_locale_lang.sql
-- bugs_closed/252 follow-on — the FIRST catch of the new site-locale-unset-check.
--
-- buytoletcalculator.uk was created 2026-08-21, after migration 508 set the estate's
-- document languages. It therefore had no `site_config.locale.lang` and would have
-- declared `en` on every page it ever built. That is precisely the silent default
-- 252 removed, met one level up: 508 is a ONE-OFF, and new sites keep arriving.
--
-- It was found by the dry run of `site-locale-unset-check` before that job was even
-- deployed — which is the honest way to test a check: run its predicate against live
-- data and see whether it finds something you did not already know.
--
-- [EVIDENCE-THIN, recorded as such, exactly like indoorplanters.co.uk in 508.] This
-- site has no identity spec, no tagline, no location and no content yet. Its evidence
-- is the .uk registration plus estate context, NOT its own copy. Re-check when it has
-- a mission or a page. The owner's ruling of 2026-08-20 stands: a non-English site
-- must NOT be en-GB, and that generalises to future language sites — so if this turns
-- out to serve another language, change it here rather than treating en-GB as settled.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then record: ./scripts/migration/run-migrations.sh --record-only <file> --note "..."
-- Rollback: 530_buytoletcalculator_locale_lang_ROLLBACK.sql

BEGIN;

-- The site has no current site_config row at all, so this is an INSERT. Guarded so
-- it cannot double-insert against the partial unique index (site_id, aspect) WHERE
-- is_current, and so it is a no-op if another lane has since set one.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes, is_current)
SELECT s.id, 'site_config',
       jsonb_build_object('locale', jsonb_build_object('lang', 'en-GB')),
       'operator', NULL, 'migration-530-locale-lang',
       'bugs_closed/252 follow-on — declared document language; found by site-locale-unset-check. EVIDENCE-THIN: .uk registration only, no content yet.',
       true
FROM sites s
WHERE s.domain = 'buytoletcalculator.uk'
  AND NOT EXISTS (SELECT 1 FROM site_specs x
                   WHERE x.site_id = s.id AND x.aspect = 'site_config' AND x.is_current);

-- If a site_config row DID already exist, merge into it instead, preserving siblings.
UPDATE site_specs ss SET
  data = ss.data || jsonb_build_object('locale',
           COALESCE(ss.data -> 'locale', '{}'::jsonb) || jsonb_build_object('lang', 'en-GB')),
  updated_at = now()
FROM sites s
WHERE ss.site_id = s.id AND s.domain = 'buytoletcalculator.uk'
  AND ss.aspect = 'site_config' AND ss.is_current
  AND COALESCE(ss.data #>> '{locale,lang}', '') = '';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain = 'buytoletcalculator.uk' AND ss.aspect = 'site_config' AND ss.is_current
    AND ss.data #>> '{locale,lang}' = 'en-GB';
  IF n <> 1 THEN
    RAISE EXCEPTION 'buytoletcalculator.uk does not carry exactly one current site_config with locale.lang=en-GB (found %) — aborting', n;
  END IF;
END $$;

-- And no real site may be left without one, which is the whole point of this file.
DO $$
DECLARE unset text;
BEGIN
  SELECT string_agg(s.domain, ', ') INTO unset
  FROM sites s
  LEFT JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'site_config' AND ss.is_current
  WHERE s.domain IS NOT NULL AND s.domain <> '' AND s.domain NOT LIKE '%.internal'
    AND COALESCE(ss.data #>> '{locale,lang}', '') = '';
  IF unset IS NOT NULL THEN
    RAISE EXCEPTION 'real sites still declare no language: % — decide each one EXPLICITLY (never from the TLD) and add it', unset;
  END IF;
END $$;

COMMIT;

-- VERIFY: the site-locale-unset-check's own predicate should now return an empty
-- 'unset' array. Run the job, or its query, rather than trusting this file.
