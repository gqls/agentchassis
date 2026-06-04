-- 108_thunder_training_monitor_orchestrator.sql
-- DB: clients_db   (agent_definitions + scheduled_tasks both live in clients_db)
--
-- Completes the thunder-training-monitor (4b): the ORCHESTRATOR agent definition
-- that finds every running training instance and dispatches a per-instance worker
-- (107) for each, plus the scheduled_tasks row that fires it periodically.
--
-- WHY orchestrator-with-loop and NOT the reaper's scheduler-pre_query shape:
-- the scheduler merges only the FIRST pre_query row into input_data and fires once
-- per tick (010 §Pre-Queries). The reaper can use that because it decommissions
-- its target, so the next-oldest rotates to the top each tick. The monitor must
-- check EVERY running instance each tick (ALIVE ones stay running and would
-- otherwise sit at the top of the ordering forever, starving newer instances that
-- may have finished). So the orchestrator does the find+fan-out; the scheduler
-- just triggers it (gated so it skips when nothing is training).
--
-- Trigger path (no new deployment/topic wiring needed — 001 §Topics): the
-- scheduler posts to system.agent.generic.requests with config.agent_type =
-- 'thunder-training-monitor'; the orchestrator runs in the existing generic
-- chassis pods (pure coordination: find + spawn/call). Each worker is spawned as
-- its own Job pod with a per-spawn job.<id>.requests topic. Workers reply to the
-- orchestrator (their parent), per 001 §Agents respond to caller's topic.
--
-- Loop is SEQUENTIAL (call_agent awaits by default): same-step spawns reuse one
-- topic safely only because the prior worker Job has completed before the next is
-- spawned (001 §Topics). Do NOT make the worker call fire-and-forget.
--
-- DEPLOY ORDER (important): (1) build+deploy the 5 new chassis actions
-- (dispatch_thunder_ssh_get_status, classify_training_probe, record_probe_streak,
-- mark_training_run_terminal, find_active_training_instances) and register them;
-- (2) apply 106 (counter) and 107 (worker); (3) apply this 108; (4) manually test
-- the worker against a live instance; (5) ENABLE the schedule (it is inserted
-- DISABLED so it cannot fire workers before the actions exist). Enable with:
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'thunder-training-monitor';
--
-- Idempotent: ON CONFLICT DO UPDATE on both rows. category/image copied from
-- model-trainer (matches the Phase 5 family).

BEGIN;

-- ───────────────────────────────────────────────────────────────────────────
-- 1. thunder-training-monitor orchestrator definition.
--    find_active_training_instances -> loop(spawn_worker + call_worker) -> done
-- ───────────────────────────────────────────────────────────────────────────
INSERT INTO agent_definitions (
    type, version, display_name, description, category, status,
    default_config, input_contract, output_contract,
    image_repository, image_tag, is_active
)
VALUES (
    'thunder-training-monitor',
    1,
    'Thunder Training Monitor (orchestrator)',
    'Periodic orchestrator: finds every running Thunder instance with a training_run_id and, per instance, spawns a thunder-training-monitor-worker that probes the box and reconciles/decommissions on completion or failure. Fired by the scheduler; pure coordination (find + spawn/call), substantive per-instance work runs in the worker Job pods.',
    'training',
    'active',
    '{
      "workflow": {
        "processing_mode": "task",
        "timeout_seconds": 600,
        "start_step": "find_instances",
        "steps": {
          "find_instances": {
            "action": "find_active_training_instances",
            "description": "Load all running Thunder instances with a training_run_id and no decommission pending from clients_db; returns instances plus count.",
            "config": {},
            "output_field": "find_instances",
            "next_step": "monitor_loop"
          },
          "monitor_loop": {
            "action": "loop",
            "description": "For each running training instance, spawn a per-instance worker and call it. Sequential: a same-step spawn topic is reused safely only after the prior worker Job completes. continue_on_error so one failing instance does not abort the tick.",
            "config": {
              "items_field": "find_instances.instances",
              "item_variable": "current_instance",
              "max_iterations": 25,
              "continue_on_error": true,
              "sub_workflow": {
                "start_step": "spawn_worker",
                "steps": {
                  "spawn_worker": {
                    "action": "spawn_agent",
                    "description": "Spawn a per-instance worker Job (role monitor_worker).",
                    "config": { "role": "monitor_worker", "agent_type": "thunder-training-monitor-worker" },
                    "output_field": "spawn_worker",
                    "next_step": "call_worker"
                  },
                  "call_worker": {
                    "action": "call_agent",
                    "description": "Call the just-spawned worker with this instance ids and await its terminal result before the next iteration.",
                    "config": {
                      "agent_type": "thunder-training-monitor-worker",
                      "target_role": "monitor_worker",
                      "input_mapping": {
                        "provisioning_id": "current_instance.provisioning_id",
                        "training_run_id": "current_instance.training_run_id"
                      }
                    },
                    "output_field": "call_worker"
                  }
                }
              }
            },
            "output_field": "monitor_loop",
            "next_step": "done"
          },
          "done": {
            "action": "complete_workflow",
            "description": "Terminal. Surface the loop summary.",
            "config": { "output_fields": ["monitor_loop"] }
          }
        }
      }
    }'::jsonb,
    jsonb_build_object(
        'required', jsonb_build_array(),
        'optional', jsonb_build_array('pending')
    ),
    jsonb_build_object(
        'produces', jsonb_build_array('monitor_loop')
    ),
    COALESCE((SELECT image_repository FROM agent_definitions
               WHERE type = 'model-trainer' AND (is_snapshot IS NULL OR is_snapshot = false) AND deleted_at IS NULL
               ORDER BY version DESC LIMIT 1), 'docker.io/aqls/agent-chassis'),
    COALESCE((SELECT image_tag FROM agent_definitions
               WHERE type = 'model-trainer' AND (is_snapshot IS NULL OR is_snapshot = false) AND deleted_at IS NULL
               ORDER BY version DESC LIMIT 1), 'latest'),
    true
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name    = EXCLUDED.display_name,
    description     = EXCLUDED.description,
    category        = EXCLUDED.category,
    status          = EXCLUDED.status,
    default_config  = EXCLUDED.default_config,
    input_contract  = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    image_repository = EXCLUDED.image_repository,
    image_tag       = EXCLUDED.image_tag,
    is_active       = EXCLUDED.is_active,
    updated_at      = NOW();

-- Validate the orchestrator workflow parsed and key steps + the loop sub_workflow exist.
DO $$
DECLARE
    steps jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps' INTO steps
      FROM agent_definitions
     WHERE type = 'thunder-training-monitor'
       AND (is_snapshot IS NULL OR is_snapshot = false) AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;

    IF steps IS NULL
       OR steps->'find_instances' IS NULL
       OR steps->'monitor_loop' IS NULL
       OR steps->'done' IS NULL
       OR steps#>'{monitor_loop,config,sub_workflow,steps,spawn_worker}' IS NULL
       OR steps#>'{monitor_loop,config,sub_workflow,steps,call_worker}' IS NULL THEN
        RAISE EXCEPTION 'migration 108: orchestrator workflow steps / loop sub_workflow missing or unparsed';
    END IF;
END $$;

-- ───────────────────────────────────────────────────────────────────────────
-- 2. scheduled_tasks row — fire the orchestrator every 5 min, gated so it skips
--    when nothing is training. Inserted DISABLED; enable after the deploy steps
--    above. concurrency_group keeps a single monitor tick in flight; the gating
--    pre_query also serves as dynamic input (harmless 'pending' column merged
--    into input_data — the orchestrator self-discovers and ignores it).
-- ───────────────────────────────────────────────────────────────────────────
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    concurrency_group, max_concurrent, timeout_seconds, pre_query, enabled
)
VALUES (
    'thunder-training-monitor',
    'Every 5 min: fire the thunder-training-monitor orchestrator to probe running training instances, reconcile finished/failed runs and decommission their boxes. Gated: skipped when no running training instance exists. INSERTED DISABLED — enable after the 5 monitor actions are deployed and 106/107/108 applied.',
    300,
    'thunder-training-monitor',
    'system.agent.generic.requests',
    'thunder-training-monitor',
    1,
    600,
    'SELECT 1 AS pending FROM public.thunder_instances WHERE status = ''running'' AND training_run_id IS NOT NULL AND decommission_requested_at IS NULL LIMIT 1',
    false
)
ON CONFLICT (name) DO UPDATE SET
    description       = EXCLUDED.description,
    interval_seconds  = EXCLUDED.interval_seconds,
    target_agent_type = EXCLUDED.target_agent_type,
    target_topic      = EXCLUDED.target_topic,
    concurrency_group = EXCLUDED.concurrency_group,
    max_concurrent    = EXCLUDED.max_concurrent,
    timeout_seconds   = EXCLUDED.timeout_seconds,
    pre_query         = EXCLUDED.pre_query,
    -- NOTE: deliberately NOT overwriting `enabled` on re-run, so a manual enable
    -- is not silently reverted by re-applying the migration.
    updated_at        = NOW();

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually):
-- ───────────────────────────────────────────────────────────────────────────
-- \echo orchestrator workflow:
-- SELECT jsonb_pretty(default_config) FROM agent_definitions
--  WHERE type='thunder-training-monitor' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
-- \echo schedule row (enabled should be false until you flip it):
-- SELECT name, interval_seconds, target_agent_type, enabled, concurrency_group, max_concurrent, pre_query
--   FROM scheduled_tasks WHERE name='thunder-training-monitor';
--
-- ENABLE once actions are deployed + 106/107 applied + a manual worker test passes:
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'thunder-training-monitor';
--
-- Manual one-off worker test (no schedule needed) — fire one worker at a known instance
-- by posting to the generic entry point with input_data {provisioning_id, training_run_id}:
--   (use your trigger-*.sh / kcat against system.agent.generic.requests with
--    config.agent_type='thunder-training-monitor-worker'); expect STATUS=ALIVE -> reset_streak -> done
--   on a still-training box, or complete/fail + decommission on a finished one.
