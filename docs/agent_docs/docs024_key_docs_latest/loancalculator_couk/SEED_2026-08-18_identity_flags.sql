-- SEED — loancalculator.co.uk: the three identity-policy flags into the
-- structure spec aspect (bugs_open/241 plumbing; OWNER DECISION 2026-08-18,
-- "I would prefer /guides/ but I am happy to accept the most natural fix for
-- the code").
--
-- WHY THIS IS THE NATURAL FIX, not a workaround.
--   On 2026-08-17 a recompose fire wrote plan 9463e31d, which retyped all 14
--   guides to role `blog-post` and moved them from /guides/<slug>.html to
--   /blog/<slug>.html; 14 duplicate pages built and deployed. The framework
--   already has the control for exactly this hazard, written while planning
--   THIS site's rebuild (bugs_open/241): normaliseRealisedToPlanPage stamps a
--   realised-derived plan page `identity_authority: "realised"` and carries
--   `parent_section`, so CanonicalisePage keeps the page where it is SERVING
--   instead of re-deriving the role's default hub. Its own comment names our
--   case: "CanonicalisePage re-derives a blog-post's URL under /blog/, which
--   MOVES a live page that is serving from /guides/".
--   That control is opt-in per site with the unsafe default OFF (owner ruling
--   2026-08-02), and this site has never had it on:
--     honour_realised_identity / twin_identity_snap / stem_twin_snap were all
--     NULL on the current structure spec (measured 2026-08-18).
--
-- WHY ALL THREE AND NOT JUST honour_realised_identity.
--   `honour_realised_identity` is INERT unless a pairing layer has re-stamped
--   the page: the marker is stripped from every LLM-proposed page
--   (v3_site_actions.go:6476, "re-stamped only by a snap or a union"), so with
--   both snap layers off nothing re-stamps and realisedIdentityOf returns
--   ok=false. That precondition is NOT documented where the flag is documented;
--   it was established 2026-08-17 by the loanandmortgagecalculator lane
--   (commit 96c83ebff), which enabled the flag alone, measured the population
--   first, and got the twins anyway. Their conclusion for their own next round
--   was to seed all three together. `stem_twin_snap` is the layer that matches
--   this site's shape — a bare plan page against a prefixed realised one, in
--   either direction (`can-i-overpay` vs `guide-can-i-overpay`).
--
-- ⚠ WHAT THIS SEED DOES NOT DO, stated so nobody reads more into it.
--   It changes NOTHING until the next planner run for this site. It does not
--   retract the 14 duplicate pages already deployed, does not restore the
--   guides to the plan, and is NOT a substitute for the owner's separate
--   decision about those. It only makes the NEXT plan stop moving live URLs.
--
-- ⚠ NOT VERIFIED, and it needs a run to verify: which pass handled the 14
--   pages on 2026-08-17, and therefore whether honour_realised_identity alone
--   would have sufficed. Zero PLAN_PAGE_STEM_TWIN_OBSERVED rows exist
--   fleet-wide all-history, so the stem layer did not observe a pairing for
--   them — consistent with the plan entries having matched a realised page by
--   EXACT NAME at validate time and the identity being re-derived later at the
--   WRITE (which is 241's mechanism), but the run's own collected_data had
--   PURGED before it could be read (~2h, not the ~2 days the handoff assumes).
--   Treat the three-flag seeding as the LMC lane's measured recommendation
--   rather than as a claim about which layer will fire here.
--
-- ⚠ Do NOT rely on `pinned` to protect this row: write_site_spec drops pinned
--   (LANDMINES). A RE-ADOPTION of this site silently drops these opt-in keys
--   (LANDMINES 2026-08-11). Check after any adoption run:
--     SELECT data ? 'honour_realised_identity' FROM site_specs
--     WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
--       AND aspect='structure' AND is_current;

BEGIN;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND aspect = 'structure' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
  site_id,
  'structure',
  data || '{"honour_realised_identity": true, "twin_identity_snap": true, "stem_twin_snap": true}'::jsonb,
  'operator',
  'identity flags seeded 2026-08-18 (bugs_open/241; owner decision: keep /guides/, accept the natural code fix). Stops the planner re-deriving a live page''s identity from its role and moving it to the role default hub — the 2026-08-17 /blog/ incident. All three together per the precondition established by the loanandmortgagecalculator lane (96c83ebff): honour_realised_identity is inert unless a snap layer re-stamps the page. Carries forward url_shape:flat and the adoption-written pages/source/adopted_from unchanged.',
  true,
  'loancalculator_lane_20260818'
FROM site_specs
WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND aspect = 'structure'
ORDER BY created_at DESC
LIMIT 1;

-- Verify or ABORT. A block of SELECTs cannot stop a COMMIT (ON_ERROR_STOP
-- ignores a non-empty result), so this must RAISE — see CLAUDE.md / RFC_006.
DO $$
DECLARE n int; honour boolean; twin boolean; stem boolean; shape text; n_pages int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current structure spec, found %', n; END IF;

  SELECT (data->>'honour_realised_identity')::boolean,
         (data->>'twin_identity_snap')::boolean,
         (data->>'stem_twin_snap')::boolean,
         data->>'url_shape',
         jsonb_array_length(data->'pages')
    INTO honour, twin, stem, shape, n_pages
  FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;

  IF honour IS NOT TRUE THEN RAISE EXCEPTION 'honour_realised_identity not true after seed'; END IF;
  IF twin   IS NOT TRUE THEN RAISE EXCEPTION 'twin_identity_snap not true after seed';       END IF;
  IF stem   IS NOT TRUE THEN RAISE EXCEPTION 'stem_twin_snap not true after seed';           END IF;

  -- The carry-forward is the whole risk of a supersede-then-insert: url_shape
  -- and the adoption-written pages list must survive, or this seed has quietly
  -- undone the 2026-08-11 seed and the adoption record.
  IF shape IS DISTINCT FROM 'flat' THEN
    RAISE EXCEPTION 'url_shape lost or changed: expected flat, found %', COALESCE(shape,'(null)');
  END IF;
  IF n_pages <> 27 THEN
    RAISE EXCEPTION 'pages list changed: expected 27 entries, found %', COALESCE(n_pages,-1);
  END IF;
END $$;

COMMIT;
