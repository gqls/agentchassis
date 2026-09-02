-- 708_enable_unrendered_page_imagery_HOLD.sql
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  check_unrendered_page_imagery.go HAS ROLLED.                           ##
-- ##  The gate commit is the one that ships this file with the check.        ##
-- ############################################################################
--
-- bugs_open/114's closing detector: adds "unrendered_page_imagery" to
-- design-discovery-agent's checks array. The check files ONE flag-only rollup
-- per (site, state) — states unwired / fragment_slot / no_image_slot — for
-- pages holding an active content-hero asset the page renders nowhere, and
-- retracts a state's rollup when its census empties. Design rationale, the
-- census numbers behind each state, and the reason the generator was NOT gated:
-- docs/agent_docs/docs024_key_docs_latest/bugfix_114_imagery_wiring/
-- PLAN_2026-08-22_imagery_wiring.md (2026-09-02 revision).
--
-- WHY THE HOLD (648's reason, verbatim in effect). run_discovery_checks
-- resolves each configured name against the binary's registry, and an
-- unregistered name FAILS THE WHOLE STEP inside defer tx.Rollback() — taking
-- all 24 of this agent's live checks down on every design pass for as long as
-- the gap lasts. Config is live immediately; Go is not. IMAGE FIRST, THEN
-- CONFIG — 368, 494, 526 and 648 all carry this hold for this reason.
--
-- The `_HOLD` suffix is load-bearing: SIDECAR_RE excludes this file from
-- `--apply` while still listing it, so it is held back visibly. A banner alone
-- would not hold it — the runner does not read comments.
--
-- WHY design-discovery-agent. It already runs the imagery family this check
-- completes (content_image_missing generates; unfulfilled_imagery_plan tracks
-- the plan; undeployed_assets watches deploys; image_url_404 watches unbacked
-- references). This check asks the one question none of them can: the asset
-- exists, deployed, active — and the page it was made for renders it nowhere.
-- undeployed_assets in particular CANNOT absorb it: its evidence is
-- purpose-prefix and site-wide (one wired sibling vouches for every asset of
-- that purpose) and its remedy is a deploy, which is how its 1,651-row parked
-- backlog was built (bugs_open/114, NOTES 2026-09-02).
--
-- APPLY IT BY HAND, in this order (648's runbook, adapted). ⚠ PREFLIGHT FIRST
-- (round-3 guardian advisory): the duplicate-active-rows landmine — four agent
-- types carry TWO active definition rows and only the higher version loads.
-- The in-transaction needle-gate below already refuses on n<>1, but check
-- BEFORE spending the apply attempt:
--   SELECT count(*) FROM agent_definitions WHERE type='design-discovery-agent'
--    AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
-- Expect exactly 1 ([MEASURED 2026-09-02] design-discovery-agent is NOT among
-- the duplicated four — but "no duplicate exists" and "a duplicate cannot
-- cause damage" are different properties, and the four-type list grows).
--
--   1. CONFIRM THE RUNNING CHASSIS REGISTERS THE CHECK — probe the CAPABILITY
--      with a positive and a negative control:
--
--        SELECT count(DISTINCT pod_name) FILTER (WHERE name='unrendered_page_imagery')  AS have,
--               count(DISTINCT pod_name) FILTER (WHERE name='undeployed_assets')        AS positive_control,
--               count(*) FILTER (WHERE name='unrendered_page_imagery_NOTREAL')          AS negative_control
--          FROM service_binary_capabilities
--         WHERE service='agent-chassis' AND kind='discovery_check'
--           AND last_seen_at > now() - interval '10 minutes';
--
--      Require have = positive_control and negative_control = 0. Keep the
--      window NARROW — 648 measured a 45-minute window mixing pre- and
--      post-roll spawned pods and reading as a half-finished rollout.
--
--   2. THE FIRST QUERY AFTER APPLYING IS "WHAT DID I BREAK?":
--
--        SELECT current_step, status, count(*) FROM orchestration_states
--         WHERE owner_agent_type='design-discovery-agent'
--           AND created_at > now() - interval '30 minutes'
--         GROUP BY 1,2;
--
--   3. ONLY THEN look for the benefit — and EXPECT ROLLUPS. The 2026-09-02
--      census says the fleet holds all three states (231 no-slot tool pages,
--      16 fragment slots, and unwired populations on webdesign.co.uk,
--      gamesdesign.co.uk, loanandmortgagecalculator.co.uk among others). A
--      zero across the fleet is an unexercised detector, not a clean reading:
--
--        SELECT s.domain, wi.item_key, wi.spec->>'count', wi.spec->>'measured_at'
--          FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
--         WHERE wi.item_type='unrendered_page_imagery'
--         ORDER BY wi.created_at DESC;
--
--   4. COMPANION EDIT, IN THE SAME COMMIT THAT APPLIES THIS FILE: add
--      "unrendered_page_imagery" to liveConfiguredChecks in
--      discovery_checks_registration_test.go — that fixture asserts the LIVE
--      agents' roster, and this migration is what changes the roster.
--
--      Rollback: remove the name from the array (an absent check is simply not
--      run, so the rollback needs no hold of its own).

BEGIN;

-- Idempotency FIRST, snapshot second (648's guardian-caught ordering: a
-- snapshot taken after the refusal check cannot mislabel a post-update state
-- as 'pre-update' on a re-run).
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'design-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["unrendered_page_imagery"]'::jsonb;
    IF done > 0 THEN
        RAISE EXCEPTION '708: already applied — design-discovery-agent already runs unrendered_page_imagery';
    END IF;
END $$;

SELECT snapshot_agent('design-discovery-agent', '708_enable_unrendered_page_imagery: pre-update');

-- COUNTED NEEDLE-GATE. The step must still be the run_discovery_checks step,
-- carrying a checks ARRAY that already holds the two siblings this file's
-- "why this agent" argument rests on; and exactly ONE active non-snapshot row
-- must match (the duplicate-active-rows trap — [MEASURED 2026-09-02] the
-- checks array read 24 names on one row; "no duplicate exists" and "a
-- duplicate cannot cause damage" are different properties).
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'design-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->>'action' = 'run_discovery_checks'
       AND jsonb_typeof(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') = 'array'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["undeployed_assets","content_image_missing"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION
          '708 needle-gate: expected exactly 1 design-discovery-agent whose run_checks step is run_discovery_checks with a checks array containing BOTH undeployed_assets and content_image_missing, found % — re-derive against the live workflow', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
           || '["unrendered_page_imagery"]'::jsonb),
       updated_at = NOW()
 WHERE type = 'design-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- VERIFY INSIDE THE TRANSACTION, AS A RAISE — a SELECT cannot stop the COMMIT.
DO $$
DECLARE ok int; n_checks int;
BEGIN
    SELECT count(*) INTO ok FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'design-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["undeployed_assets","content_image_missing","unrendered_page_imagery"]'::jsonb;
    IF ok <> 1 THEN
        RAISE EXCEPTION '708 verify: design-discovery-agent does not carry the new check alongside its siblings after the update (matched %)', ok;
    END IF;

    -- Containment cannot tell an APPEND from a REPLACE; the LENGTH can.
    -- [MEASURED 2026-09-02] the array held 24 names; it must hold exactly 25
    -- after. If the live array has grown since that count, re-derive BOTH
    -- numbers before editing this arm — do not delete it.
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type='design-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks <> 25 THEN
        RAISE EXCEPTION '708 verify: checks array is % long, expected exactly 25 (24 as of 2026-09-02 + this one) — jsonb_set REPLACED where it should have APPENDED, or the array moved under this file', n_checks;
    END IF;
END $$;

COMMIT;
