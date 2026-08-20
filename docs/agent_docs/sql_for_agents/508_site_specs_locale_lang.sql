-- 508_site_specs_locale_lang.sql
-- bugs_open/252 (og/lang slug) §B — give each site a language to declare.
--
-- 507 gives the shared head templates a GATED lang attribute. This file sets the
-- value the gate reads: site_specs aspect `site_config`, key `locale.lang`,
-- resolved into the template by the existing schema-driven config path
-- (`source: config.locale.lang` -> resolveConfigPath). Second consumer of the
-- STY-050 mechanism; worked precedent migration 339.
--
-- ⚠⚠ HOLD RELEASED AND APPLIED 2026-08-20 17:2xZ, after 507 and after the binary probe.
-- ⚠ FIRST APPLY ATTEMPT ABORTED, correctly: the 'a real site this file does not name'
-- guard caught indoorplanters.co.uk, created the same day this file was authored. It was
-- added explicitly (evidence-thin, marked as such) and the file re-applied. That abort is
-- the guard earning its place — the alternative was a new site silently keeping `en`.
-- Original banner:
-- ⚠⚠ _HOLD — APPLY AFTER 507, AND ONLY ONCE THE BINARY CARRYING
-- head_assembly.go IS PROVEN RUNNING. Full statement of the ordering trap is in
-- 507's header; it applies identically here, and this file is the one that
-- actually changes served bytes.
--
-- OWNER DECISION 2026-08-20: opt the estate in NOW rather than shipping the
-- mechanism switched off. The reason is on record — a mechanism that ships with
-- zero consumers rots unexercised, and this platform has been bitten by that
-- before. The counter-risk (a site declaring the wrong language) is handled by
-- setting each domain EXPLICITLY below rather than deriving from the TLD, which
-- bugs_open/252 itself rejected: `.com` sites on this estate are mostly British,
-- so a TLD rule "guesses wrong in the direction that matters".
--
-- ⚠ ONE SITE IS NOT ENGLISH, AND A BLANKET RULE WOULD HAVE SHIPPED A FALSE
-- VALUE ON IT. relojistas.com is a Spanish-language publication: identity
-- location `España`, tagline "Portal de noticias de relojería en español", and
-- every heading on the served page is Spanish (checked 2026-08-20). It gets
-- `es-ES`. Its current `en` is false today; `en-GB` would have been false in a
-- newer, more confident way — the exact defect class this bug is about. Every
-- other real site reads `United Kingdom` in its identity spec, or has no location
-- recorded and English copy (fundamentallyai.com, gaswholesalers.com,
-- robot-hands.com, vonc.com — all four checked individually, all English).
--
-- POPULATION: the 26 real sites. All `*.internal` pool domains and
-- `system.internal` are EXCLUDED — they serve no visitor, so a language on them
-- is noise. 25 get `en-GB`; relojistas.com gets `es-ES`.
--
-- SAFETY:
--   · `site_config` is operator-owned, and update_site_spec_from_item merges
--     per-field, so the in-place merge below cannot be wholesale-reverted by a
--     pipeline writer (339's note, same mechanism).
--   · The merge PRESERVES any existing keys — `data || jsonb_build_object('locale', <existing locale> || {lang})`
--     rather than jsonb_set on '{locale}', which would replace a whole locale
--     object if one ever gains siblings. 15 sites already hold a current
--     site_config row (analytics.gtm_container_id, chrome.footer_note,
--     chrome.compliance_lines); none of those keys is touched.
--   · 10 sites have no current site_config row and get an INSERT.
--   · Nothing is served differently until each site's chrome re-renders. 507's
--     template edit and this file's spec edit BOTH move the render_inputs
--     digest (it hashes template and specs by value), so StaleSiteComponentsCheck
--     files a stale_chrome item per site and the existing detect->rebuild pipe
--     carries it out in per-site waves. That is the intended propagation — no new
--     sweep is built for this.
--   · webdesign.co.uk will hold `en-GB` and keep serving `en`: its head component
--     is a bare fragment with no <head> open tag to carry the attribute (see
--     507's "NOT COVERED"). The value is set anyway so that fixing that
--     component later needs no second migration. 117 pages, the most in the
--     fleet — do not read its unchanged lang as this migration failing.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then record: ./scripts/migration/run-migrations.sh --record-only <file> --note "..."
-- Rollback: 508_site_specs_locale_lang_ROLLBACK.sql

BEGIN;

CREATE TEMP TABLE _locale_targets(domain text PRIMARY KEY, lang text NOT NULL) ON COMMIT DROP;
INSERT INTO _locale_targets(domain, lang) VALUES
  -- .uk / .co.uk — 18
  ('adversecreditmortgage.co.uk',      'en-GB'),
  ('cookly.uk',                        'en-GB'),
  ('finetuning.uk',                    'en-GB'),
  ('gamesdesign.co.uk',                'en-GB'),
  ('idea.uk',                          'en-GB'),
  ('lendzy.co.uk',                     'en-GB'),
  ('leopardessconsulting.co.uk',       'en-GB'),
  ('loanandmortgagecalculator.co.uk',  'en-GB'),
  ('loancalculator.co.uk',             'en-GB'),
  ('loancash.co.uk',                   'en-GB'),
  ('loanzy.uk',                        'en-GB'),
  ('mortgagecalculator.co.uk',         'en-GB'),
  ('noted.co.uk',                      'en-GB'),
  ('remortgagecalculator.uk',          'en-GB'),
  ('vetcomparison.uk',                 'en-GB'),
  ('webdesign.co.uk',                  'en-GB'),
  ('webdesign.uk',                     'en-GB'),
  -- Added 2026-08-20 on the FIRST apply attempt, which ABORTED on this domain:
  -- it was created that same day, after this file was authored. It has no
  -- identity spec and no content yet, so unlike every other row here its
  -- evidence is the .co.uk registration plus estate context, NOT its own copy.
  -- [EVIDENCE-THIN, deliberately recorded as such.] Re-check when it has a
  -- mission or a page; the guard below is what forced the decision instead of
  -- letting it default to `en` in silence.
  ('indoorplanters.co.uk',             'en-GB'),
  -- .com, British — 7 (identity location "United Kingdom", or English copy
  -- confirmed by eye where no location is recorded)
  ('ai-agent-orchestration.com',       'en-GB'),
  ('dartsonline.com',                  'en-GB'),
  ('fundamentallyai.com',              'en-GB'),
  ('gaswholesalers.com',               'en-GB'),
  ('oufe.com',                         'en-GB'),
  ('robot-hands.com',                  'en-GB'),
  ('vonc.com',                         'en-GB'),
  -- NOT English — 1
  ('relojistas.com',                   'es-ES');

-- Every named domain must exist, or the list has drifted since authoring.
DO $$
DECLARE missing text;
BEGIN
  SELECT string_agg(t.domain, ', ') INTO missing
  FROM _locale_targets t LEFT JOIN sites s ON s.domain = t.domain WHERE s.id IS NULL;
  IF missing IS NOT NULL THEN
    RAISE EXCEPTION 'these domains are named in this migration but do not exist in sites: % — re-read the estate before applying', missing;
  END IF;
END $$;

-- And no real site may be silently omitted: if a new site has been added since
-- authoring, this ABORTS rather than quietly leaving it on the `en` default.
DO $$
DECLARE unlisted text;
BEGIN
  SELECT string_agg(s.domain, ', ') INTO unlisted
  FROM sites s LEFT JOIN _locale_targets t ON t.domain = s.domain
  WHERE s.domain IS NOT NULL AND s.domain <> ''
    AND s.domain NOT LIKE '%.internal'
    AND t.domain IS NULL;
  IF unlisted IS NOT NULL THEN
    RAISE EXCEPTION 'real sites exist that this migration does not name: % — decide each one''s language explicitly (do NOT derive it from the TLD) and add it, then re-apply', unlisted;
  END IF;
END $$;

-- A. Sites that already hold a current site_config row: merge, preserving
--    every existing key and any sibling keys under locale.
UPDATE site_specs ss SET
  data = ss.data || jsonb_build_object(
           'locale',
           COALESCE(ss.data -> 'locale', '{}'::jsonb) || jsonb_build_object('lang', t.lang)),
  updated_at = now()
FROM sites s, _locale_targets t
WHERE ss.site_id = s.id
  AND s.domain = t.domain
  AND ss.aspect = 'site_config'
  AND ss.is_current;

-- B. Sites with no current site_config row at all.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes, is_current)
SELECT s.id, 'site_config',
       jsonb_build_object('locale', jsonb_build_object('lang', t.lang)),
       'operator', NULL, 'migration-508-locale-lang',
       'bugs_open/252 §B — declared document language; owner decision 2026-08-20',
       true
FROM sites s JOIN _locale_targets t ON t.domain = s.domain
WHERE NOT EXISTS (
  SELECT 1 FROM site_specs x
  WHERE x.site_id = s.id AND x.aspect = 'site_config' AND x.is_current);

-- C. Assert the outcome: every named site now resolves a lang, the values are
--    the intended ones, and nothing else lost a key.
DO $$
DECLARE n_set int; n_expected int; n_es int; n_gtm_before int;
BEGIN
  SELECT count(*) INTO n_expected FROM _locale_targets;

  SELECT count(*) INTO n_set
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id JOIN _locale_targets t ON t.domain = s.domain
  WHERE ss.aspect = 'site_config' AND ss.is_current
    AND ss.data #>> '{locale,lang}' = t.lang;
  IF n_set <> n_expected THEN
    RAISE EXCEPTION 'expected % sites to carry the intended locale.lang, found % — aborting', n_expected, n_set;
  END IF;

  -- The one non-English site must NOT have been swept into en-GB.
  SELECT count(*) INTO n_es
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain = 'relojistas.com' AND ss.aspect = 'site_config' AND ss.is_current
    AND ss.data #>> '{locale,lang}' = 'es-ES';
  IF n_es <> 1 THEN
    RAISE EXCEPTION 'relojistas.com did not get es-ES — a blanket en-GB would be false metadata on a Spanish-language site';
  END IF;

  -- The pre-existing keys this file must not disturb.
  SELECT count(*) INTO n_gtm_before
  FROM site_specs WHERE aspect = 'site_config' AND is_current
    AND data #>> '{analytics,gtm_container_id}' IS NOT NULL;
  IF n_gtm_before <> 14 THEN
    RAISE EXCEPTION 'analytics.gtm_container_id count is %, expected 14 — this migration disturbed an unrelated key', n_gtm_before;
  END IF;
END $$;

COMMIT;

-- VERIFY (read-only, after apply):
--   SELECT s.domain, ss.data #>> '{locale,lang}' AS lang,
--          ss.data #>> '{analytics,gtm_container_id}' AS gtm_untouched
--     FROM site_specs ss JOIN sites s ON s.id = ss.site_id
--    WHERE ss.aspect = 'site_config' AND ss.is_current ORDER BY s.domain;
-- Expect 25 rows: 24 en-GB, relojistas.com es-ES, and every previously-set
-- gtm/chrome key still present.
--
-- Then at the ARTEFACT, once each site's chrome has re-rendered:
--   curl -s https://<domain>/<inner-page> | grep -oE '<html[^>]*>'
-- Expect lang="en-GB" (es-ES on relojistas.com); webdesign.co.uk stays "en"
-- for the reason named in this file's header, which is NOT a failure.
