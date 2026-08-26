-- 648_enable_archived_page_still_serving_HOLD.sql
--
-- bugs_open/359 fix candidate 2. Turns ON the retired-but-serving detector by
-- adding "archived_page_still_serving" to availability-discovery-agent's checks
-- array.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  check_archived_page_still_serving.go HAS ROLLED.                       ##
-- ##  The gate commit is 36b51a51b (this check's birth).                     ##
-- ############################################################################
--
-- WHY THE HOLD. run_discovery_checks resolves each name in this array against
-- the binary's own registry (discovery_checks/registry.go: Get(name)). A name
-- the binary does not register is NOT a no-op: discovery_checks.go returns an
-- error, and that return happens BEFORE tx.Commit() inside defer tx.Rollback(),
-- so it fails the WHOLE run_checks step and DISCARDS EVERY EARLIER CHECK'S
-- FINDINGS in the same run. Applying this before the roll would therefore take
-- site_unreachable AND page_content_divergence down on every availability pass
-- for as long as the gap lasted. Config is live immediately; Go is not. IMAGE
-- FIRST, THEN CONFIG — 368, 494 and 526 all carry this hold for this reason.
--
-- The `_HOLD` suffix is load-bearing: SIDECAR_RE (`_[A-Z][A-Z0-9_]*\.sql$`)
-- excludes this file from `--apply` while STILL listing it under "Sidecars
-- (hand-run only)", so it is held back visibly rather than silently. A banner
-- alone would not hold it — the runner does not read comments.
--
-- WHY availability-discovery-agent. It is the serving-side lane, and the three
-- checks answer three DISJOINT questions about the same wire:
--
--   site_unreachable            does the apex serve at all?          (an outage)
--   page_content_divergence     do the served bytes match the store? (staleness)
--   archived_page_still_serving does a page we RETIRED still answer? (governance)
--
-- They cannot double-file: for this check a non-200 is a skip or a RESOLVE and
-- never a finding, which is the exact inverse of the other two. And its
-- population is disjoint from every other page-fetching check in the package —
-- all of those are armed on the lifecycle axis, so archived pages are fetched by
-- NOTHING today. That is also the answer to register/operator-practice.md's
-- standing refusal of "a third fetcher of every page": this fetches the pages
-- nobody fetches.
--
-- APPLY IT BY HAND, in this order:
--
--   1. CONFIRM THE RUNNING CHASSIS REGISTERS THE CHECK — and probe the
--      CAPABILITY, not a commit that proxies it. The binary reports its own
--      registered checks (RFC_040 stage 2, migration 503), so this is a query
--      with a control rather than an exec-grep:
--
--        SELECT count(DISTINCT pod_name) FILTER (WHERE name='archived_page_still_serving') AS have,
--               count(DISTINCT pod_name) FILTER (WHERE name='site_unreachable')            AS positive_control,
--               count(*)                 FILTER (WHERE name='archived_page_still_serving_NOTREAL') AS negative_control
--          FROM service_binary_capabilities
--         WHERE service='agent-chassis' AND kind='discovery_check'
--           AND last_seen_at > now() - interval '30 minutes';
--
--      Require have = positive_control and negative_control = 0. A control that
--      comes out PRESENT means the probe matches everything and proves nothing;
--      a control that is ABSENT on both arms means the probe is broken, not that
--      the fleet is clean. [MEASURED 2026-08-26, before this check existed:
--      1288 pods, 1288 carrying site_unreachable, 0 carrying the fake name.]
--
--   2. THE FIRST QUERY AFTER APPLYING IS "WHAT DID I BREAK?", NOT "DID IT
--      WORK?" — bugs_open/336's lesson, and 526 records it costing 33 minutes
--      of every page-publish in the estate:
--
--        SELECT current_step, status, count(*) FROM orchestration_states
--         WHERE agent_type='availability-discovery-agent'
--           AND created_at > now() - interval '30 minutes'
--         GROUP BY 1,2;
--
--      An unregistered name fails the step for the WHOLE agent, taking two live,
--      useful checks down with it. That damage is not visible in this check's
--      own output.
--
--   3. ONLY THEN look for the benefit — and unlike 526, EXPECT FINDINGS. A zero
--      here is NOT a clean reading, it is an unexercised detector, and this
--      bug's §7 says so in terms: "a zero from the detector before it has ever
--      flagged the known-live cases" is not valid evidence.
--
--        SELECT s.domain, wi.summary, wi.spec->>'http_status'
--          FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
--         WHERE wi.item_type='archived_page_still_serving'
--         ORDER BY wi.created_at DESC;
--
--      [MEASURED 2026-08-26, scripts/audit-archived-still-serving.sh] the fleet
--      then held 39 archived-and-shipped pages, 7 of them serving 200 with both
--      per-domain controls holding: ai-agent-orchestration.com
--      /llm-cost-calculator.html · finetuning.uk /tools/password-entropy.html ·
--      fundamentallyai.com /blog/ai-readiness-checker-guide.html and
--      /tools/llm-cost-calculator/index.html · leopardessconsulting.co.uk
--      /our-approach.html · robot-hands.com /gripper-catalog.html and /news.html.
--      ⚠ RE-RUN THAT SCRIPT ON THE DAY. The population MOVES: both
--      loancalculator.co.uk pages the bug filed on 2026-08-22 had 404d by the
--      26th, and five of the seven above were not in that earlier sample. The
--      script's exit code is the expectation to compare against.
--
--   3b. WHEN WILL IT ACTUALLY RUN? availability-discovery-agent is driven by the
--      scheduled task `site-discovery-rotation-availability`, which fires every
--      300s and picks ONE site whose last_selected_at is older than
--      `interval '4 hours'`. [MEASURED 2026-08-26] enabled, last fired 15:01Z.
--      So the first sites are examined within minutes and the fleet within hours.
--      Confirm a RUN at orchestration_states, never at last_triggered_at: a
--      rotation whose floor excludes every site advances its own timestamps while
--      dispatching nothing, and idle is indistinguishable from busy there.
--
--   4. The item is FLAG-ONLY (handler_agent ''), so nothing acts on a finding.
--      That is deliberate: a page archived by MISTAKE and serving correctly is
--      indistinguishable on the wire from a page rightly archived and wrongly
--      serving, and un-publishing a good live page is the failure this estate
--      calls worse than the bug (bugs_open/359 §6.4). Each item's spec carries a
--      triage_hint naming both remedies and the evidence query.
--
--      Rollback is 648_..._HOLD_ROLLBACK.sql; it removes the name and restores
--      today's behaviour exactly. Removing a name is always safe — an absent
--      check is simply not run — so the rollback needs no hold of its own.
--
--   5. COMPANION EDIT, IN THE SAME COMMIT THAT APPLIES THIS FILE: add
--      "archived_page_still_serving" to liveConfiguredChecks in
--      discovery_checks_registration_test.go. That fixture is the build-enforced
--      safety proof for making an unregistered name fatal, and it asserts the
--      LIVE agents' roster — this migration is what changes the roster, so they
--      must move together or the test asserts a roster that no longer exists.
--      (This lane also refreshed that fixture by union on 2026-08-26 after
--      measuring it asserted 63 of the 82 names live agents configure.)

BEGIN;

-- ⚠ ORDER: THE IDEMPOTENCY CHECK COMES FIRST, THE SNAPSHOT SECOND.
-- 526 and 541 both snapshot immediately after BEGIN and check afterwards. That
-- is the wrong way round and the council's guardian seat caught it here
-- [medium]: on a re-run against an already-applied row, snapshot-first takes a
-- SECOND snapshot labelled 'pre-update' — which by then is a POST-update
-- snapshot wearing a pre-update label — and only then refuses. LANDMINES records
-- exactly that trap. A mislabelled snapshot is worse than no snapshot: it is the
-- artefact someone restores FROM.
--
-- Checking first costs nothing (the refusal path takes no snapshot, which is
-- correct — there is nothing to snapshot, the change is already in) and the
-- happy path is unchanged.
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["archived_page_still_serving"]'::jsonb;
    IF done > 0 THEN
        RAISE EXCEPTION '648: already applied — availability-discovery-agent already runs archived_page_still_serving';
    END IF;
END $$;

SELECT snapshot_agent('availability-discovery-agent', '648_enable_archived_page_still_serving: pre-update');

-- COUNTED NEEDLE-GATE. The step must still be the one this migration believes it
-- is: the run_discovery_checks step, carrying a checks ARRAY that already holds
-- BOTH of the names this file's "why this agent" argument rests on. Anything
-- else and the premise has moved — abort rather than write a check name into a
-- shape that will not read it.
--
-- The count=1 arm is the guard against the documented duplicate-active-rows trap
-- (a config change to the lower version silently does nothing). The council's
-- guardian seat asked, correctly, whether this agent is one of the duplicated
-- ones — asserting the guard is not the same as knowing the answer.
-- [MEASURED 2026-08-26] four types carry two active non-snapshot rows —
-- content-creator, content-creator-contact, chief-strategist,
-- site-component-architect — and `availability-discovery-agent` is NOT among
-- them. So the gate is expected to pass today; it ships because "no duplicate
-- exists" and "a duplicate cannot cause damage" are different properties.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->>'action' = 'run_discovery_checks'
       AND jsonb_typeof(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') = 'array'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["site_unreachable","page_content_divergence"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION
          '648 needle-gate: expected exactly 1 availability-discovery-agent whose run_checks step is run_discovery_checks with a checks array containing BOTH site_unreachable and page_content_divergence, found % — re-derive against the live workflow', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
           || '["archived_page_still_serving"]'::jsonb),
       updated_at = NOW()
 WHERE type = 'availability-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- VERIFY INSIDE THE TRANSACTION, AND AS A RAISE RATHER THAN A SELECT.
-- A verify block made of SELECTs cannot stop the COMMIT: ON_ERROR_STOP does not
-- fire on a non-empty result, so the migration would report success while having
-- written the wrong thing. Only an exception rolls this back.
DO $$
DECLARE ok int; n_checks int;
BEGIN
    SELECT count(*) INTO ok FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["site_unreachable","page_content_divergence","archived_page_still_serving"]'::jsonb;
    IF ok <> 1 THEN
        RAISE EXCEPTION '648 verify: availability-discovery-agent does not carry all THREE checks after the update (matched %)', ok;
    END IF;

    -- Containment cannot tell an APPEND from a REPLACE — a jsonb_set that
    -- replaced the array with just this one name would still satisfy the test
    -- above if the other two happened to be there, and would satisfy it
    -- vacuously if they were not. The LENGTH can tell them apart: 2 before, 3
    -- after. 526 shipped this same pair for the same reason.
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type='availability-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks <> 3 THEN
        RAISE EXCEPTION '648 verify: checks array is % long, expected exactly 3 (site_unreachable + page_content_divergence + archived_page_still_serving) — jsonb_set REPLACED where it should have APPENDED', n_checks;
    END IF;
END $$;

COMMIT;
