-- NNN_restore_diagnose_orchestrator_workflow.sql
--
-- Renumber NNN to the next number in your migration sequence.
--
-- The diagnose-orchestrator lost its workflow: after the earlier migrations it
-- shows has_wf=f AND orchestration_workflow IS NULL — i.e. the move out of
-- orchestration_workflow nulled the source but default_config did not end up with
-- the workflow. Since the source column is now NULL, this RE-SEEDS the (correct,
-- unchanged) spawn->call->complete workflow explicitly into default_config, with
-- processing_mode + timeout, and the call_diagnoser timeout raised to 1800 to
-- match the diagnose-agent loop budget (so the orchestrator does not time out
-- waiting for the worker).
--
-- BACKUP FIRST (standing rule): snapshot the row before changing it.
SELECT snapshot_agent('diagnose-orchestrator', 'restore lost workflow into default_config');

UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "steps": {
      "spawn_diagnoser": {
        "action": "spawn_agent",
        "config": {
          "role": "diagnoser",
          "agent_type": "diagnose-agent"
        },
        "next_step": "call_diagnoser"
      },
      "call_diagnoser": {
        "action": "call_agent",
        "config": {
          "agent_type": "diagnose-agent",
          "target_role": "diagnoser",
          "input_mapping": {
            "ref?": "input_data.ref",
            "repo?": "input_data.repo",
            "owner?": "input_data.owner",
            "symptom": "input_data.symptom",
            "site_id?": "input_data.site_id",
            "seed_scope?": "input_data.seed_scope",
            "runtime_page?": "input_data.runtime_page",
            "runtime_site?": "input_data.runtime_site",
            "correlation_id?": "input_data.correlation_id"
          },
          "timeout_seconds": 1800
        },
        "next_step": "complete"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "result_from": "diagnose-agent_result"
        }
      }
    },
    "start_step": "spawn_diagnoser"
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 1860
}'::jsonb,
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-orchestrator';

-- Verify (expect has_wf=t, has_spawn=t, mode=orchestrator, orch_null=t):
--   SELECT type,
--          (default_config ? 'workflow') AS has_wf,
--          (default_config -> 'workflow' -> 'steps' ? 'spawn_diagnoser') AS has_spawn,
--          (default_config ->> 'processing_mode') AS mode,
--          (orchestration_workflow IS NULL) AS orch_null
--   FROM agent_definitions WHERE type = 'diagnose-orchestrator';
