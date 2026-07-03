-- NNN_fix_diagnose_complete_output_fields.sql
--
-- Run 17933a83 (§7E attempt 1) failed at the CHILD's complete step: Kafka
-- Message Size Too Large. MECHANISM (read from workflow_actions.go,
-- extractFinalResult): CompleteWorkflowAction honours `output_field` (string),
-- then `output_fields` (array), then process/aggregate_results — and otherwise
-- FALLS BACK TO SHIPPING THE ENTIRE non-system collected_data.
-- `result_from` IS NOT A KEY THE ACTION READS. Both diagnose agents carry it
-- (dead config), so every diagnose completion has ALWAYS shipped the full
-- context — masked while the 69-file-era collected_data sat under the ~1.05MB
-- cap, exposed at 1.27MB by the 515-file analysis. Empirical confirmation:
-- code-indexer (output_fields ["index_result"]) DELIVERED at cd 1,213,045;
-- diagnose-agent (result_from) FAILED at cd 1,270,781.
--
-- FIX: replace both complete configs with the key the action actually reads.
-- Standing rule honoured: workflow variable names in sync with what actions
-- expect. Parent payload becomes {"diagnose-agent_result": <child result>};
-- child payload becomes {"diagnosis": <emit output>} (~6KB measured).
-- PARKED (Go hardening, future build): the ship-everything fallback deserves a
-- size guard + warn so a config typo can never again ship megabytes silently.

BEGIN;

SELECT snapshot_agent('diagnose-agent',
  'complete: result_from (a key CompleteWorkflowAction never reads; fell back to shipping full collected_data, 1.27MB > kafka cap) -> output_fields ["diagnosis"]');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,complete,config}',
      '{"output_fields": ["diagnosis"]}'::jsonb),
    updated_at = now()
WHERE type = 'diagnose-agent'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

SELECT snapshot_agent('diagnose-orchestrator',
  'complete: result_from (dead key, same fallback) -> output_fields ["diagnose-agent_result"]');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,complete,config}',
      '{"output_fields": ["diagnose-agent_result"]}'::jsonb),
    updated_at = now()
WHERE type = 'diagnose-orchestrator'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- verify — expect output_fields on both, no result_from anywhere
SELECT type,
       default_config #> '{workflow,steps,complete,config}' AS complete_config
FROM agent_definitions
WHERE type IN ('diagnose-agent','diagnose-orchestrator')
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── REVERT ──────────────────────────────────────────────────────────────────
-- BEGIN;
-- SELECT snapshot_agent('diagnose-agent','revert complete to result_from');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--   '{workflow,steps,complete,config}', '{"result_from": "diagnosis"}'::jsonb),
--   updated_at = now()
-- WHERE type='diagnose-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- SELECT snapshot_agent('diagnose-orchestrator','revert complete to result_from');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--   '{workflow,steps,complete,config}', '{"result_from": "diagnose-agent_result"}'::jsonb),
--   updated_at = now()
-- WHERE type='diagnose-orchestrator' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- COMMIT;
