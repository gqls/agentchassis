-- c2_gtm_spec_key_for_artefact_only_sites.sql — write the SOURCE the head template reads, for every
-- site that today carries GTM-PQ3WCTBD in its stored head artefact ONLY.  bugs_open/397.
--
-- WHY: the 2026-08-24 fleet backfill inserted the snippet into site_components.rendered_html on 13
-- sites and never wrote site_specs.site_config.analytics.gtm_container_id — the key the template's
-- {{if .gtm_container_id}} gate actually reads. The next chrome render regenerates the head from
-- template+inputs, finds no key, and drops the tag (agritec.uk, 2026-08-24 19:20:53 — measured).
--
-- ⚠ OWNER-GATED. APPLYING THIS IS FIRING A REBUILD PER SITE, not a quiet config fix:
--   the stale_site_components check fingerprints site_config into render_inputs, so every targeted
--   site drifts at its next discovery run → needs_rerender (stale_chrome) → rerender-pages force-
--   rerenders ALL chrome slots and every page → one commit + one Actions run PER PAGE on a two-slot
--   runner. Sized 2026-08-25: 12 sites, 241 pages (~2 h of queue). Pick the moment.
--
-- HOW:
--   dry run (rolls back, prints what it WOULD do):
--     kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--       -v DRY=1 -v GO=yes -f - < this_file
--   apply:  same, without -v DRY=1.   Without -v GO=yes it refuses. Both variables are deliberate.
--   include the 4 sites that carry it NOWHERE (agritec.uk cv1.co.uk homegarden.uk lampenkap.com, as
--   of 2026-08-25 — "standard for new builds", owner 2026-08-24):  add -v UNTAGGED=1
--
-- APPLIED 2026-08-26 10:12:11Z (rows' created_at; owner go), 17 sites / 334 pages — kept for re-use if bucket B refills.
-- ~~EXPECTED FAILURE on apis.uk~~ CORRECTED 2026-08-26: 'refuses page re-renders' DISPROVEN — three
--   page_rerenders completed there overnight (the 11:19 failure was 383's re-walk). Expect COMPLETE.
--   The served page regains the tag with the wave. AFTER APPLYING, TELL
--   apis_uk_bees_homepage (CONTRIB in this dir, 2026-08-25) — they verify at the served bytes and settle
--   pages.build_status themselves. Tell the other lanes in bugs_open/397 §9 BEFORE applying.
--
-- MERGE, never replace: 10 of the 12 targets already hold a current site_config row (locale, chrome).
-- The 2026-07-31 rollout replaced wholesale and dropped relojistas.com's intent_probe key — do not
-- repeat that. Verify afterwards with scripts/check_gtm_state.sh --db (bucket B must read 0).
\set ON_ERROR_STOP on
\if :{?GO}
\else
  \echo 'REFUSED: pass -v GO=yes after reading the banner (and -v DRY=1 first).'
  DO $r$ BEGIN RAISE EXCEPTION 'refused without -v GO=yes'; END $r$;
\endif
\if :{?UNTAGGED}
\else
  \set UNTAGGED 0
\endif

BEGIN;

CREATE TEMP TABLE gtm_targets ON COMMIT DROP AS
SELECT s.id AS site_id, s.domain,
       (SELECT sc.rendered_html LIKE '%GTM-PQ3WCTBD%' FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head') AS head_has_tag,
       EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=s.id AND ss.aspect='site_config' AND ss.is_current) AS has_site_config
  FROM sites s
 WHERE s.status IN ('deployed','active')
   AND s.network_id = '00000000-0000-0000-0000-000000000002'          -- our own estate only; a third-party site never gets our container
   AND EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head')
   AND NOT EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=s.id AND ss.aspect='site_config' AND ss.is_current
                     AND ss.data->'analytics'->>'gtm_container_id' = 'GTM-PQ3WCTBD')
   -- an explicit opt-out (analytics.mode='none') outranks every backfill — added 2026-09-03 with
   -- the mode vocabulary (STY-061); without this line a future UNTAGGED=1 run would re-tag it
   AND NOT EXISTS (SELECT 1 FROM site_specs ss2 WHERE ss2.site_id=s.id AND ss2.aspect='site_config' AND ss2.is_current
                     AND ss2.data->'analytics'->>'mode' = 'none')
   AND ( :UNTAGGED = 1
         OR COALESCE((SELECT sc.rendered_html LIKE '%GTM-PQ3WCTBD%' FROM site_components sc WHERE sc.site_id=s.id AND sc.slot_name='head'), false) );

\echo '--- targets (domain | head already serves the tag | has a site_config row to MERGE into) ---'
SELECT domain, head_has_tag, has_site_config FROM gtm_targets ORDER BY domain;
SELECT count(*) AS target_sites,
       (SELECT count(*) FROM pages p WHERE p.site_id IN (SELECT site_id FROM gtm_targets)) AS pages_that_will_rerender
  FROM gtm_targets;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM gtm_targets;
  IF n = 0 THEN RAISE EXCEPTION 'nothing to do: no artefact-only site found (bucket B is already empty)'; END IF;
  IF n > 20 THEN RAISE EXCEPTION 'refusing: % targets is more than the 16 this file was sized for on 2026-08-25 — re-census first', n; END IF;
END $g$;

-- 1. supersede the current site_config row where one exists (history kept, never deleted)
UPDATE site_specs ss SET is_current=false, superseded_at=now()
  FROM gtm_targets t
 WHERE ss.site_id=t.site_id AND ss.aspect='site_config' AND ss.is_current;

-- 2. insert the merged row: previous data (if any) || analytics.gtm_container_id
INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current, pinned)
SELECT t.site_id, 'site_config',
       COALESCE(prev.data, '{}'::jsonb)
         || jsonb_build_object('analytics',
              COALESCE(prev.data->'analytics', '{}'::jsonb) || '{"gtm_container_id":"GTM-PQ3WCTBD"}'::jsonb),
       'operator', 'claude-session-google-2026-08-25',
       'Estate GTM container written to the key the head/header templates read ({{if .gtm_container_id}} via '
       'input_schema source config.analytics.gtm_container_id). The 2026-08-24 backfill wrote only the stored '
       'artefact; agritec.uk lost the tag on its next chrome render. bugs_open/397. Merged over the previous '
       'site_config row, which is superseded (not deleted) immediately above.',
       true, false
  FROM gtm_targets t
  -- only the row superseded by step 1 IN THIS TRANSACTION (now() is constant inside it) — never an
  -- older superseded row, which would resurrect keys that were deliberately retired
  LEFT JOIN LATERAL (SELECT data FROM site_specs p WHERE p.site_id=t.site_id AND p.aspect='site_config'
                       AND p.superseded_at = now() LIMIT 1) prev ON true;

-- 3. post-conditions: every target has the key; exactly one current site_config per target; no other
--    key was lost (the merged row's key set is a superset of the superseded row's)
DO $v$
DECLARE n int; bad text;
BEGIN
  SELECT count(*) INTO n FROM gtm_targets t
   WHERE NOT EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id=t.site_id AND ss.aspect='site_config' AND ss.is_current
                        AND ss.data->'analytics'->>'gtm_container_id'='GTM-PQ3WCTBD');
  IF n <> 0 THEN RAISE EXCEPTION 'post: % target(s) still lack the key', n; END IF;
  SELECT string_agg(t.domain, ' ') INTO bad FROM gtm_targets t
   WHERE (SELECT count(*) FROM site_specs ss WHERE ss.site_id=t.site_id AND ss.aspect='site_config' AND ss.is_current) <> 1;
  IF bad IS NOT NULL THEN RAISE EXCEPTION 'post: not exactly one current site_config on: %', bad; END IF;
  SELECT string_agg(t.domain, ' ') INTO bad FROM gtm_targets t
   JOIN site_specs cur ON cur.site_id=t.site_id AND cur.aspect='site_config' AND cur.is_current
   JOIN site_specs old ON old.site_id=t.site_id AND old.aspect='site_config' AND NOT old.is_current
                      AND old.superseded_at >= now() - interval '1 minute'
   WHERE NOT (cur.data ?& ARRAY(SELECT jsonb_object_keys(old.data)));
  IF bad IS NOT NULL THEN RAISE EXCEPTION 'post: a key was LOST on: %', bad; END IF;
END $v$;

\echo '--- after: current site_config rows on the targets ---'
SELECT t.domain, (SELECT string_agg(k, ',' ORDER BY k) FROM jsonb_object_keys(ss.data) k) AS keys
  FROM gtm_targets t JOIN site_specs ss ON ss.site_id=t.site_id AND ss.aspect='site_config' AND ss.is_current ORDER BY 1;

\if :{?DRY}
  ROLLBACK;
  \echo 'DRY RUN — rolled back. Nothing changed.'
\else
  COMMIT;
  \echo 'APPLIED. Expect one stale_chrome → needs_rerender item per target at its next discovery run; verify with scripts/check_gtm_state.sh --db'
\endif
