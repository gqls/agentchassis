-- ASSIGN — route the remaining open required_fields_missing rows at the router
-- (bugs_open/277, seed 410, register CQ-023). THE ACTUAL FIX for the parked
-- backlog — promoted to a reviewed edit at the council's direction (round 1
-- gating objection, corr 7b0e2833: "it should be an edit with its own
-- sketch/grounding, not a footnote executed outside review").
--
-- PRECONDITIONS (all four held when the canary ran, 2026-08-15 ~11:05Z):
--   1. Seed 410 applied; agent active (the claim gate otherwise blocks items).
--   2. Canary verified per-arm: 332bb3f6 stale->complete (orch 0177ce18),
--      4fa5b019 partial->converted (orch + content_rewrite:from_rfm: item),
--      e512af8a blob->parked, 483fb749 gas converter->parked. All four routes
--      matched the census prediction.
--   3. The census (CENSUS_2026-08-15_predicted_routes.sql, output saved beside
--      it) predicts every remaining row's route. Re-run it if more than a day
--      has passed — the queue moves daily.
--   4. Pre-image taken (STEP 0 below).
--
-- STEP 0 — PRE-IMAGE, before the write (needle-gate: the revert must not
-- depend on memory). Save beside this file:
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -t -A -F $'\t' -c \
--     "SELECT id, status, handler_agent, attempt_count, updated_at
--        FROM site_work_items
--       WHERE item_type='required_fields_missing'
--         AND status IN ('needs_human_review','unresolved')
--         AND COALESCE(handler_agent,'')=''" \
--     > DATA_<date>_assignment_preimage.tsv
--
-- STEP 1 — the guarded write (this file). The DO block re-derives the target
-- set AT EXECUTION TIME, refuses to run against a drifted queue shape, and
-- asserts the row count it changed.
--
-- REVERT (per row or wholesale; the pre-image holds the exact prior state —
-- every target row was needs_human_review/'' with its own attempt_count):
--
--   UPDATE site_work_items
--      SET handler_agent = '', status = 'needs_human_review', updated_at = NOW()
--    WHERE item_type = 'required_fields_missing'
--      AND handler_agent = 'required-fields-missing-handler'
--      AND status = 'triaged';   -- only rows the router has not yet processed
--
--   Rows the router already processed are not revertable by status flip alone —
--   complete/parked rows carry their outcome; consult the pre-image and the
--   orchestration trail (workflow_plan->>'start_step'='classify').
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < ASSIGN_….sql

BEGIN;

DO $$
DECLARE
    target_count  integer;
    claimed_count integer;
    agent_ok      boolean;
    changed       integer;
BEGIN
    -- The handler must exist and be active, or every row goes 'blocked' at claim.
    SELECT EXISTS (SELECT 1 FROM agent_definitions
                    WHERE type = 'required-fields-missing-handler' AND is_active
                      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL)
      INTO agent_ok;
    IF NOT agent_ok THEN
        RAISE EXCEPTION 'ASSIGN refused: required-fields-missing-handler is not an active agent definition (apply seed 410 first)';
    END IF;

    -- Target set, re-derived NOW (not from the census's day): parked, unrouted.
    SELECT count(*) INTO target_count FROM site_work_items
     WHERE item_type = 'required_fields_missing'
       AND status IN ('needs_human_review','unresolved')
       AND COALESCE(handler_agent,'') = '';
    IF target_count = 0 THEN
        RAISE EXCEPTION 'ASSIGN refused: 0 unrouted parked rows — either already assigned (re-running this file?) or the queue drifted; re-census before forcing anything';
    END IF;

    -- Nothing mid-flight may be swept: a claimed row belongs to a running loop.
    SELECT count(*) INTO claimed_count FROM site_work_items
     WHERE item_type = 'required_fields_missing' AND status = 'claimed';
    IF claimed_count <> 0 THEN
        RAISE EXCEPTION 'ASSIGN refused: % row(s) of this type are claimed mid-dispatch — wait a cadence and re-run', claimed_count;
    END IF;

    UPDATE site_work_items
       SET handler_agent = 'required-fields-missing-handler',
           status = 'triaged', attempt_count = 0, updated_at = NOW()
     WHERE item_type = 'required_fields_missing'
       AND status IN ('needs_human_review','unresolved')
       AND COALESCE(handler_agent,'') = '';
    GET DIAGNOSTICS changed = ROW_COUNT;

    IF changed <> target_count THEN
        RAISE EXCEPTION 'ASSIGN aborted: changed % rows but targeted % — concurrent write between census and update', changed, target_count;
    END IF;

    RAISE NOTICE 'ASSIGN: % row(s) routed at required-fields-missing-handler (attempt_count reset — the claim gate requires attempt_count < max_attempts)', changed;
END $$;

COMMIT;

-- POSTCONDITION (run after one dispatch cadence per site; the loop takes one
-- site per ~120s tick, so a fleet of N sites takes up to N cadences):
--
--   SELECT status,
--          COALESCE(result->>'route','(in orchestration trail)') AS route, count(*)
--   FROM site_work_items WHERE item_type='required_fields_missing'
--   GROUP BY 1,2 ORDER BY 3 DESC;
--
--   Expect: parked rows (needs_human_review) carry result.route + the router's
--   message in error; completed rows' trail is in orchestration_states
--   (workflow_plan->>'start_step'='classify'); zero rows left at triaged with
--   an empty handler; zero at 'blocked'.
