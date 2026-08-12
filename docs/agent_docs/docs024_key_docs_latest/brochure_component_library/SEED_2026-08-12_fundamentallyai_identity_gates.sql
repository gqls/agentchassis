-- SEED — fundamentallyai.com: enable TWO of the three page-identity gates
-- (bugs_open/215 quiet mode, PLAN-048). Owner decision 2026-08-12.
--
-- Enables:  honour_realised_identity, twin_identity_snap
-- Leaves OFF (deliberately): stem_twin_snap
--
-- Read by siteIdentityPolicyFor (site_identity_policy.go), live in the chassis on
-- v1.0.1290 — artefact-verified 2026-08-12: literals "twin_identity_snap" and
-- "PLAN_PAGE_IDENTITY_TWIN_OBSERVED" present on BOTH replicas, one-letter
-- near-miss controls absent on both.
--
-- WHY stem_twin_snap STAYS OFF, and it is not mere caution. On THIS site the stem
-- layer is the one that would fire: both twin pairs carry the bare name in the plan
-- and the `tool-` spelling only as a realised page. With the layer ON, the snap
-- rewrites each bare plan entry onto the matched `tool-` page
-- (snapPlanPageOntoRealised -> normaliseRealisedToPlanPage), i.e. it moves future
-- builds onto the `tool-` side. Both pairs are 3 components against 3, so that is a
-- SURVIVOR CHOICE MADE BY MACHINE — reserved as a per-pair owner call (bugs_open/215
-- decision O2, RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md).
-- Left off, the layer still RECORDS what it would have done
-- (PLAN_PAGE_STEM_TWIN_OBSERVED), which is the evidence this pilot exists to
-- gather — and here that recording is harmless, because both sides of both pairs
-- are ALREADY realised, so no new page row can be minted by observing them.
--
-- ⚠ THIS SITE HAS NO STRUCTURE SPEC ROW AT ALL, so this is an INSERT, not the
--   supersede-then-insert carry-forward the sibling SEED_*.sql files perform. That
--   is safe and was checked, not assumed: the only other reader of this aspect is
--   siteUsesFlatURLs, whose own contract is "absent spec, absent key ... all mean
--   false" (site_url_shape.go:29-32), so a row carrying neither `url_shape` nor the
--   adoption keys is indistinguishable from today's no-row state for every reader
--   except mine. fundamentallyai was framework-built, never adopted, which is why
--   there are no pages/source/adopted_from keys to carry forward.
--
-- ⚠ ROLLBACK IS SAFE: setting these false (or deleting the row) moves no live URL.
--   The next replan simply reverts to minting twins — the pre-fix bug, no worse.
--
-- ⚠ RE-ADOPTION previously dropped these keys silently. That is FIXED and LIVE as of
--   v1.0.1290 (carryForwardStructureSpecKeys, 19acfc895 — verified present on both
--   replicas 2026-08-12 with a near-miss control). The check below is still the right
--   one after any adoption run, and note `?` (key presence) never `->>'k' = 'true'`:
--     SELECT data ? 'honour_realised_identity', data ? 'twin_identity_snap'
--     FROM site_specs WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
--       AND aspect='structure' AND is_current;
--
-- NOTHING HAPPENS UNTIL fundamentallyai's page list is next rebuilt. Owner decision
-- 2026-08-12 was to WAIT for a natural rebuild rather than trigger one (that site
-- has ~47 open work items and its sweep front owns its cleanup). Until then this row
-- is inert by design — an empty counter afterwards means "no replan yet", NOT "no
-- twins". Read the population with the classifying join in LANDMINES.md
-- ("The page-identity dark-launch counter is NOT a passive instrument"), never a
-- bare count.

BEGIN;

-- Guard: this seed assumes NO current structure row (measured 2026-08-12: zero).
-- If another session has created one since, ABORT — the correct action is then a
-- key-merge onto their row, not this INSERT, which would clobber their data.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='structure' AND is_current;
  IF n <> 0 THEN
    RAISE EXCEPTION 'expected 0 current structure specs for fundamentallyai, found % — another session created one; MERGE the two keys onto it instead of running this INSERT', n;
  END IF;
END $$;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
VALUES (
  '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd',
  'structure',
  '{"honour_realised_identity": true, "twin_identity_snap": true}'::jsonb,
  'operator',
  'bugs_open/215 PLAN-048 pilot, seeded 2026-08-12 on owner decision: honour_realised_identity + twin_identity_snap ON, stem_twin_snap deliberately OFF (it would pick the tool- side as survivor for both twin pairs, which is owner call O2). First structure spec for this site — it was framework-built, never adopted, so there are no adoption keys to carry forward.',
  true,
  'brochure_215_quiet_mode_thread'
);

-- Verify or abort. Checks BOTH what must be true and what must be a NO-OP: a seed
-- that silently enabled the third gate would look identical to a correct one if we
-- only asserted the two we meant to set.
DO $$
DECLARE n int; has_honour boolean; has_twin boolean; has_stem boolean;
        v_honour boolean; v_twin boolean;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='structure' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current structure spec after seed, found %', n; END IF;

  SELECT data ? 'honour_realised_identity',
         data ? 'twin_identity_snap',
         data ? 'stem_twin_snap',
         (data->>'honour_realised_identity')::boolean,
         (data->>'twin_identity_snap')::boolean
    INTO has_honour, has_twin, has_stem, v_honour, v_twin
  FROM site_specs
  WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='structure' AND is_current;

  IF NOT has_honour THEN RAISE EXCEPTION 'honour_realised_identity key absent after seed'; END IF;
  IF NOT has_twin   THEN RAISE EXCEPTION 'twin_identity_snap key absent after seed'; END IF;
  IF NOT v_honour   THEN RAISE EXCEPTION 'honour_realised_identity present but not true'; END IF;
  IF NOT v_twin     THEN RAISE EXCEPTION 'twin_identity_snap present but not true'; END IF;
  -- The no-op assertion: the third gate must NOT have been created, even as false.
  -- Absent and false behave identically today, but an explicit false would read to
  -- the next operator as "someone considered and disabled this", which is a
  -- different fact from "never set".
  IF has_stem THEN RAISE EXCEPTION 'stem_twin_snap must remain ABSENT on this site (owner call O2 pending), found the key present'; END IF;
END $$;

COMMIT;
