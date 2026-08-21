-- 526_enable_page_content_divergence_HOLD.sql
--
-- bugs_open/315 candidate 4 / PLAN_2026-08-19 decision D5, phase 4.
-- Turns ON the divergence sweep by adding "page_content_divergence" to
-- availability-discovery-agent's checks array.
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  check_page_content_divergence.go HAS ROLLED.                           ##
-- ##  The gate commit is f715b8c1d.                                          ##
-- ############################################################################
--
-- WHY THE HOLD. run_discovery_checks resolves each name in this array against
-- the binary's own registry (discovery_checks/registry.go: Get(name)). A name
-- the binary does not register is not a no-op — check_site_unreachable's
-- migration 368 carries the same hold for the same reason, in its own words:
-- "the runner hard-fails on a name the binary does not register". Config is
-- live immediately; Go is not. Hence the hold, and it runs the same way round
-- as 494's: IMAGE FIRST, THEN CONFIG.
--
-- The `_HOLD` suffix is load-bearing: SIDECAR_RE (`_[A-Z][A-Z0-9_]*\.sql$`)
-- excludes this file from `--apply` while STILL listing it under "Sidecars
-- (hand-run only)", so it is held back visibly rather than silently. A banner
-- alone would not hold it — the runner does not read comments.
--
-- WHY availability-discovery-agent AND NOT ONE OF THE OTHER THREE. It is the
-- serving-side lane: it carries exactly one check today, site_unreachable,
-- which probes the public apex and files on AVAILABILITY while EXPLICITLY
-- declining the staleness class (its header records mortgagecalculator "serves
-- a divergent render today — a staleness defect, not an availability one" and
-- files nothing for it). That declined class is precisely this check's remit,
-- so the two sit in one lane without overlapping: a non-200 is site_unreachable's
-- finding and this check's skip.
--
-- APPLY IT BY HAND, in this order:
--
--   1. Confirm the running chassis registers the check. Ask the ARTEFACT, not
--      git, and per SERVICE rather than per fleet:
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--          | grep -m1 'build provenance'
--        git merge-base --is-ancestor f715b8c1d <the stamped sha> && echo SAFE-TO-ENABLE
--
--      An EMPTY grep means "the startup line has scrolled", NOT "unstamped" —
--      it is a startup line on a busy service. Fall back to the binary probe,
--      and ALWAYS run a control in the same breath (a sha that must be absent
--      and one that must be present); never `strings`, which is absent from
--      these images, and never a discovery grep for "some 40-hex string", which
--      matches Go's internal digit table and returns the same wrong answer on
--      every service.
--
--   2. THE FIRST QUERY AFTER APPLYING IS "WHAT DID I BREAK?", NOT "DID IT
--      WORK?". This is bugs_open/336's lesson, committed by this very lane at
--      the moment it was fixing this very bug, and it cost 33 minutes of every
--      page-publish in the estate:
--        SELECT current_step, status, count(*) FROM orchestration_states
--         WHERE agent_type='availability-discovery-agent'
--           AND created_at > now() - interval '30 minutes'
--         GROUP BY 1,2;
--      An unregistered check name fails the run_checks step for the WHOLE
--      agent, taking site_unreachable — a live, useful check — down with it.
--      That is the damage to look for, and it is not visible in this check's
--      own output.
--
--   3. ONLY THEN look for the benefit. Expect ZERO findings: [MEASURED
--      2026-08-21] all 228 active pages then carrying a content_hash served
--      bytes hashing exactly to their fingerprint. A finding on day one is more
--      likely to be a defect in the check than a divergence in the fleet —
--      read the item's stored_hash/served_hash pair and re-run the comparison
--      by hand before believing it:
--        SELECT summary, spec->>'stored_hash', spec->>'served_hash'
--          FROM site_work_items WHERE item_type='page_content_divergence'
--         ORDER BY created_at DESC LIMIT 10;
--        curl -s "https://<domain><url>?cb=$RANDOM$RANDOM" | sha256sum
--
--   3b. WHEN WILL IT ACTUALLY RUN? Sooner than the discovery-rotation folklore
--      suggests, and this is measured. availability-discovery-agent is driven by
--      the scheduled task `site-discovery-rotation-availability`, which fires
--      every 300s and picks ONE site whose last_selected_at is older than
--      `interval '4 hours'` — NOT the 7 days the quality rotation uses.
--      [MEASURED 2026-08-21 13:33Z] all 25 rotation rows were stamped within the
--      preceding four hours, so the whole fleet is swept roughly every 4-5 hours
--      and the first sites are examined within minutes of applying this.
--      Confirm the cadence yourself rather than trusting this comment:
--        SELECT name, interval_seconds,
--               substring(pre_query from 'interval ''[^'']+''') AS floor,
--               last_triggered_at
--          FROM scheduled_tasks WHERE name = 'site-discovery-rotation-availability';
--      And confirm a RUN at orchestration_states, never at last_triggered_at: a
--      rotation whose floor excludes every site advances its own timestamps while
--      dispatching nothing, and idle is indistinguishable from busy there.
--
--   4. The check is structurally inert on any page whose last deploy predates
--      the fingerprint — 588 of 816 pages on 2026-08-21 — so a quiet first day
--      is expected on that ground too, and the population grows as pages
--      redeploy. Rollback is 526_..._HOLD_ROLLBACK.sql; it removes the name and
--      restores today's behaviour exactly.

BEGIN;

SELECT snapshot_agent('availability-discovery-agent', '526_enable_page_content_divergence: pre-update');

-- Already-applied gate (the runner reads a RAISE containing 'already').
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_content_divergence"]'::jsonb;
    IF done > 0 THEN
        RAISE EXCEPTION '526: already applied — availability-discovery-agent already runs page_content_divergence';
    END IF;
END $$;

-- COUNTED NEEDLE-GATE. The step must still be the one this migration believes
-- it is: the run_discovery_checks step, carrying a checks ARRAY that already
-- holds site_unreachable. Anything else and the premise has moved — abort
-- rather than write a check name into a shape that will not read it.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->>'action' = 'run_discovery_checks'
       AND jsonb_typeof(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') = 'array'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["site_unreachable"]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION
          '526 needle-gate: expected exactly 1 availability-discovery-agent whose run_checks step is run_discovery_checks with a checks array containing site_unreachable, found % — re-derive against the live workflow', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
           || '["page_content_divergence"]'::jsonb),
       updated_at = NOW()
 WHERE type = 'availability-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- VERIFY INSIDE THE TRANSACTION, AND AS A RAISE RATHER THAN A SELECT.
-- A verify block made of SELECTs cannot stop the COMMIT: ON_ERROR_STOP does not
-- fire on a non-empty result, so the migration would report success while
-- having written the wrong thing. Only an exception rolls this back.
DO $$
DECLARE ok int; n_checks int;
BEGIN
    SELECT count(*) INTO ok FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'availability-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_content_divergence","site_unreachable"]'::jsonb;
    IF ok <> 1 THEN
        RAISE EXCEPTION '526 verify: availability-discovery-agent does not carry BOTH checks after the update (matched %)', ok;
    END IF;

    -- The append must not have dropped or duplicated anything: 1 check before,
    -- 2 after. A jsonb_set that replaced the array instead of appending would
    -- pass the containment test above with site_unreachable gone.
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type='availability-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks <> 2 THEN
        RAISE EXCEPTION '526 verify: checks array is % long, expected exactly 2 (site_unreachable + page_content_divergence)', n_checks;
    END IF;
END $$;

COMMIT;
