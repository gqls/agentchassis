-- 715_HOLD: enable the cta_rank_anomaly discovery check (bugs_open/436, the
-- alarm half of 391 decision 3's pairing; lane bugfix_436_cta_eligibility).
--
-- ⛔ _HOLD — DO NOT let the runner apply this. APPLY BY HAND, AFTER the image
-- carrying check_cta_rank_anomaly.go has rolled. The discovery runner FAILS
-- THE WHOLE STEP on a check name its binary does not register
-- (discovery_checks.go: "failing the step … the run would have silently
-- omitted it"), and allow_unregistered_checks is deliberately not set here —
-- flipping that flag fleet-wide to smuggle one name through would tolerate
-- EVERY unregistered name. Config is live the moment this applies; the check
-- exists only after the roll. Image first, then seeds. A banner cannot hold
-- an ordering-critical file — the _HOLD suffix is what the runner's sidecar
-- regex excludes (see migration-runner practice).
--
-- HOW TO VERIFY BEFORE APPLYING: the running chassis must register the check.
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=2000 \
--     | grep -m1 'cta_rank_anomaly'   # appears in the runner's "registered" list on first discovery run
-- or induce: run any discovery step and read its logged `registered` array.
--
-- WHAT IT DOES: appends 'cta_rank_anomaly' to the completeness-discovery-agent
-- checks array — the same array that carries misdirected_cta. Guarded: no-op
-- if already present; aborts if the agent row or the checks array is not where
-- this expects it (the array moved once before; better loud than wrong).

BEGIN;

DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '715: expected exactly 1 live completeness-discovery-agent row, found %', n;
    END IF;
END $$;

-- Pre-mutation backup (council round 1, debug_historian, corr 9faa2a23): the
-- row governs every site's discovery run, so the guarded UPDATE alone is not
-- the discipline. TWO-ARG overload deliberately — it writes
-- agent_definitions_backup (the one-arg form writes an is_snapshot row into
-- agent_definitions itself; LANDMINES "snapshot_agent has TWO overloads").
-- Verify the snapshot holds the PRE-change config, not that one merely exists:
--   SELECT snapshot_taken_at,
--          NOT ((default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly') AS has_old
--   FROM agent_definitions_backup WHERE type='completeness-discovery-agent'
--   ORDER BY snapshot_taken_at DESC LIMIT 1;   -- has_old must be true
SELECT snapshot_agent('completeness-discovery-agent',
    '715_enable_cta_rank_anomaly_check_HOLD.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (default_config #> '{workflow,steps,run_checks,config,checks}') || '["cta_rank_anomaly"]'::jsonb
    )
WHERE type = 'completeness-discovery-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config #> '{workflow,steps,run_checks,config,checks}') IS NOT NULL
  AND NOT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly';

-- Verify: the name is now present exactly once, at the path this file wrote.
-- DO/RAISE so a miss aborts the COMMIT — including the "array was not at the
-- expected path so the UPDATE matched nothing" case, which a bare UPDATE
-- reports as success (0 rows) and this refuses to.
DO $$
DECLARE
    ok boolean;
BEGIN
    SELECT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly'
    INTO ok
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF NOT COALESCE(ok, false) THEN
        RAISE EXCEPTION '715: cta_rank_anomaly is not in the checks array after the update — is the array still at workflow.steps.run_checks.config.checks? Read the live row before editing this path.';
    END IF;
END $$;

COMMIT;
