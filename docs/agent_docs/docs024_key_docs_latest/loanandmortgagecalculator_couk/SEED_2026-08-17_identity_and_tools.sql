-- SEED — loanandmortgagecalculator.co.uk: honour_realised_identity + plan_includes_tools
-- into the structure spec aspect, ahead of the D6 planner loop
-- (PLAN_2026-08-17_site_plan_seed_and_planner_loop.md).
--
-- WHY honour_realised_identity, WITH THE MEASUREMENT ITS OWN AUTHOR ASKED FOR.
-- site_identity_policy.go says the switch is per-site precisely so it can be
-- turned on "where the population has actually been measured". Measured
-- 2026-08-17 by calling the real datahelpers.CanonicalisePage over this site's
-- 45 live active pages, through the descriptor write_site_plan_action.go:487
-- actually builds (Role=stored page_type, Slug=stored NAME, ParentSection=
-- parentSectionFromURL(url), FlatURLs=false):
--
--     45 active pages: 7 fixed points, 38 moved (name 17, url 38, type 0)
--
-- All 17 name moves are calculators: mortgages-stamp-duty ->
-- tool-mortgages-stamp-duty at /mortgages/mortgages-stamp-duty/index.html. A
-- name that collides with nothing on (site_id, name) is INSERTed, and
-- sync_pages_to_db is step 12 of 14 in the live build-site-planner workflow —
-- so without this key the first replan mints 17 phantom calculator pages and
-- strands the real ones (bugs_open/215, in that file's own words).
--
-- This site is the flag's FIRST live consumer. The sibling loancalculator.co.uk
-- replanned successfully without it because its pages were already canonical;
-- that is not evidence the flag is unnecessary here, it is evidence that site
-- had nothing to lose.
--
-- WHY plan_includes_tools, AND WHAT IT IS *NOT* WORTH HERE. Read migration
-- 407's live query before believing the sibling's rationale transfers: the
-- level test is
--   ( component_level IN ('section','element')      -- UNCONDITIONAL
--     OR ( component_level='tool' AND <this key> AND <placed on this site> ) )
-- and this site's 23 B2 calculator components are all component_level='section'
-- (measured 2026-08-17: 17 of 19 page_type='tool' pages carry section-level rows
-- only). So they are in the planner's menu either way. The key is worth exactly
-- three components today — loans-consolidation,
-- tool-affordability-complaint-checker, tool-overpayment-priority, the only
-- tool-level rows on the site, two of them created by the improvement loop on
-- 08-15 — plus every tool the generator makes from here. Seeded for that reason
-- and no larger one. (An earlier draft of the PLAN claimed this key was the
-- blocker for all 23; corrected in place, logged in WRONG_CALLS.md.)
--
-- DELIBERATELY NOT SEEDED: url_shape (this site is mixed — 39 flat / 6 nested;
-- absent = nested is right for new pages, and the identity key carries the
-- existing ones), twin_identity_snap and stem_twin_snap (separate question,
-- unmeasured collapse population, and one canary should answer one question).
--
-- Both values are the STRING "true": 407's query compares
-- data->>'plan_includes_tools' = 'true', and siteIdentityPolicyFor casts
-- (data->>'honour_realised_identity')::boolean, which accepts it.
--
-- ⚠ Do NOT rely on `pinned` to protect this row — write_site_spec ignores and
--   drops it (LANDMINES). After any adoption run, check the keys survived:
--     SELECT data ?& array['honour_realised_identity','plan_includes_tools']
--     FROM site_specs WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd'
--       AND aspect='structure' AND is_current;

BEGIN;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = 'ed633ada-f8af-424b-b4d4-8af79160dbcd'
  AND aspect = 'structure' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
  site_id,
  'structure',
  data || '{"honour_realised_identity": "true", "plan_includes_tools": "true"}'::jsonb,
  'operator',
  'honour_realised_identity + plan_includes_tools seeded 2026-08-17 for the D6 planner loop. Identity key justified by measurement: 38 of 45 live pages are not fixed points of CanonicalisePage, 17 of them calculators that would be re-minted under tool-<name> (bugs_open/215). Tools key is worth 3 tool-level components on this site, not 23 — the B2 calculators are section-level and unconditionally in the menu. Carries forward the adoption-written pages/source/adopted_from unchanged.',
  true,
  'claude-session-lmc-planner-20260817'
FROM site_specs
WHERE site_id = 'ed633ada-f8af-424b-b4d4-8af79160dbcd'
  AND aspect = 'structure'
ORDER BY created_at DESC
LIMIT 1;

-- Verify or abort. A verify block of plain SELECTs cannot stop the COMMIT
-- (ON_ERROR_STOP ignores a non-empty result set), so this is DO/RAISE.
DO $$
DECLARE n int; has_identity boolean; has_tools boolean; v_identity text; v_tools text; n_pages int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' AND aspect='structure' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current structure spec, found %', n; END IF;

  SELECT data ? 'honour_realised_identity', data ? 'plan_includes_tools',
         data->>'honour_realised_identity', data->>'plan_includes_tools',
         jsonb_array_length(data->'pages')
    INTO has_identity, has_tools, v_identity, v_tools, n_pages
  FROM site_specs
  WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' AND aspect='structure' AND is_current;

  IF NOT has_identity THEN RAISE EXCEPTION 'honour_realised_identity key absent after seed'; END IF;
  IF NOT has_tools THEN RAISE EXCEPTION 'plan_includes_tools key absent after seed'; END IF;
  IF v_identity <> 'true' THEN RAISE EXCEPTION 'honour_realised_identity is %, want the string true', v_identity; END IF;
  IF v_tools <> 'true' THEN RAISE EXCEPTION 'plan_includes_tools is %, not the string true 407 matches', v_tools; END IF;
  -- The adoption-written list is 41 entries against 45 live active pages; that
  -- staleness is recorded in the PLAN as an open question and must survive this
  -- seed untouched, so the canary can show what the planner does with it.
  IF n_pages <> 41 THEN RAISE EXCEPTION 'pages list changed: expected 41 entries, found %', n_pages; END IF;
  IF (SELECT data->>'source' FROM site_specs
      WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' AND aspect='structure' AND is_current)
     IS DISTINCT FROM 'adoption'
    THEN RAISE EXCEPTION 'source key lost or changed by the seed — carry-forward broken'; END IF;
END $$;

COMMIT;

-- Read back (run separately; this is not part of the transaction):
--   SELECT id, created_at, created_by,
--          data->>'honour_realised_identity' AS identity,
--          data->>'plan_includes_tools'      AS tools,
--          jsonb_array_length(data->'pages') AS n_pages
--   FROM site_specs
--   WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd'
--     AND aspect='structure' AND is_current;
--
-- To revert (forward-only: supersede this row and re-insert the previous data,
-- do not DELETE):
--   the previous current row is 4863c1a6-5d8f-4c97-b00d-f4a29ed8c255
--   (site-adoption-agent, 2026-07-31 19:33:30Z, neither key present).
