-- NNN_fix_diagnose_route_state_threading.sql
--
-- Renumber NNN to the next number in your migration sequence.
--
-- Fixes the diagnose-agent loop's state threading. diagnose_route reads its prior
-- LoopState from config.state_field, but its OWN result lands under the route step's
-- output_field ("route"), so the state is at route.diagnose_state — NOT the bare
-- "diagnose_state" the config currently points at (which never exists at top level,
-- confirmed by collected_data_keys on run 8d488e01: there is a "route" key, no
-- "diagnose_state" key). Result: the loop RE-SEEDS every iteration instead of
-- threading, which silently breaks three things even though the loop still runs and
-- re-scopes (the scope flows through route.scope.Symbols, a separate field):
--   1. max_iterations is NEVER enforced (Iteration resets to 1 each pass);
--   2. the evidence_trail is truncated to the FINAL iteration (audit trail lost);
--   3. the cross-iteration guards (evidence-must-grow / no-thrash / scope-must-narrow)
--      reset each pass, so they cannot detect spinning across iterations.
--
-- Run 8d488e01 ended CONFIRMED only because the model happened to confirm on the 5th
-- pass; it was not the iteration cap stopping it, and its trail shows just 1 entry.
--
-- This is the operative LIVE fix and needs NO rebuild: the workflow config sets
-- state_field explicitly, so changing it here takes effect immediately. (The action's
-- code DEFAULT has also been corrected to route.diagnose_state for the next build, so
-- a workflow that omits state_field is correct too; that part is not urgent.)
--
-- Reads its state from route.diagnose_state — consistent with how diagnose_emit already
-- reads route.status / route.conclusion / route.evidence_trail. The route step's output
-- persists across the loop-back (diagnose_assemble_bundle already reads route.scope.Symbols
-- on loop-back and the re-scope works, so route.diagnose_state will thread the same way).

-- BACKUP FIRST.
SELECT snapshot_agent('diagnose-agent',
  'fix diagnose_route state threading: state_field diagnose_state -> route.diagnose_state (persist LoopState across loop-backs; restores iteration cap, full trail, guard memory)');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,route,config,state_field}',
      '"route.diagnose_state"'::jsonb,
      false),
    updated_at = now()
WHERE type = 'diagnose-agent';

-- Verify:
--   SELECT default_config #>> '{workflow,steps,route,config,state_field}' AS state_field
--   FROM agent_definitions WHERE type = 'diagnose-agent';
-- expect: route.diagnose_state

-- After this + a re-run, the diagnose-agent row's evidence_trail should carry ONE entry
-- PER iteration (not just the last), and a non-converging run should stop at
-- stopped_by = 'iteration-cap' rather than running on. Read by correlation_id:
--   SELECT collected_data #>> '{route,iteration}'                       AS iteration,
--          jsonb_array_length(collected_data #> '{route,evidence_trail}') AS trail_len,
--          collected_data #>> '{route,stopped_by}'                       AS stopped_by,
--          collected_data #>> '{emit,status}'                            AS status
--   FROM orchestration_states WHERE correlation_id = '<the-new-id>' ORDER BY created_at;

-- ─────────────────────────────────────────────────────────────────────────────
-- REVERT:
-- SELECT snapshot_agent('diagnose-agent','revert state_field route.diagnose_state -> diagnose_state');
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config,'{workflow,steps,route,config,state_field}','"diagnose_state"'::jsonb,false),
--     updated_at = now()
-- WHERE type = 'diagnose-agent';
