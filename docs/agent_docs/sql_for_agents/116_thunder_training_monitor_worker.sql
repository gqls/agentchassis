-- 107_thunder_training_monitor_worker.sql
-- DB: clients_db   (the chassis reads flywheel-C agent_definitions from clients_db)
--
-- Inserts the thunder-training-monitor-worker agent definition: the per-instance
-- worker the thunder-training-monitor orchestrator spawns once per running
-- training instance (one sub-agent per instance — NOT a subworkflow). It probes
-- one box and, depending on what it finds, reconciles the training_runs row and
-- releases the instance.
--
-- Pairs with (next migration): the thunder-training-monitor ORCHESTRATOR
-- (find_active_training_instances -> loop -> spawn+call this worker per instance)
-- and the scheduler entry that fires it periodically. Applying 107 alone is inert
-- (the worker exists but nothing calls it yet), so this is a safe standalone step.
--
-- Depends on these chassis actions being deployed FIRST (this migration only
-- writes the workflow that references them):
--   dispatch_thunder_ssh_get_status   (probe — publishes ssh_get_status, awaits)
--   classify_training_probe           (pure: STATUS= + reachable -> verdict, routes)
--   record_probe_streak               (consecutive-unreachable counter; needs mig 106)
--   mark_training_run_terminal        (training_runs running -> complete|failed)
--   dispatch_thunder_decommission     (existing; release the instance)
--
-- Worker input_data (set by the orchestrator's call_agent input_mapping at spawn):
--   provisioning_id (required) — thunder_instances.id; ssh_get_status + decommission
--                                resolve against it
--   training_run_id (required) — the model_lifecycle.training_runs row to reconcile
-- Every step reads provisioning_id/training_run_id from input_data (configOrInput /
-- ExtractActionInputs) or from a prior step's result by key — so NO per-step
-- input_mapping is needed here.
--
-- Step flow (classify and bump_streak override next_step at runtime):
--   probe -> classify --alive-----------> reset_streak --------------> done
--                     --unreachable------> bump_streak --{<thr}-------> done
--                                                       --{>=thr lost}-> mark_failed -> decommission -> done
--                     --done_ok----------> mark_complete -> decommission -> done
--                     --done_fail/gone---> mark_failed   -> decommission -> done
--
-- Await reply keying (verified, coordinator.go ~L1636/L2408): a step result is
-- stored under BOTH the step name AND its output_field, adapter body under
-- ".response". The probe step is named "probe" and classify's probe_step default
-- is "probe", so classify reads probe.stdout / probe.reachable robustly.
--
-- Idempotent: ON CONFLICT (type, version) DO UPDATE refreshes the row on re-run.
-- category/image_repository/image_tag are copied from model-trainer so the worker
-- matches the rest of Phase 5 (category was otherwise an unverified guess).
--
-- NOTE on numbering: confirm 107 is the next free SQL migration number in your
-- runner (a project doc 106_claude_anthropic_skill.md occupies 106 as a doc).

BEGIN;

INSERT INTO agent_definitions (
    type, version, display_name, description, category, status,
    default_config, input_contract, output_contract,
    image_repository, image_tag, is_active
)
VALUES (
           'thunder-training-monitor-worker',
           1,
           'Thunder Training Monitor (per-instance worker)',
           'Per-instance worker spawned by thunder-training-monitor. Probes one running training instance over SSH; on completion reconciles training_runs to complete and decommissions the box, on failure/crash/lost marks it failed and decommissions, and tolerates transient unreachability via a consecutive-unreachable counter (mig 106).',
           COALESCE((SELECT category FROM agent_definitions
                     WHERE type = 'model-trainer' AND (is_snapshot IS NULL OR is_snapshot = false) AND deleted_at IS NULL
                     ORDER BY version DESC LIMIT 1), 'training'),
           'active',
           '{
             "workflow": {
               "processing_mode": "task",
               "timeout_seconds": 300,
               "start_step": "probe",
               "steps": {
                 "probe": {
                   "action": "dispatch_thunder_ssh_get_status",
                   "description": "Probe the instance over SSH (provisioning_id from input_data): is run.sh/02_train alive, else what do run.sh markers say? Emits one of STATUS=ALIVE|DONE_OK|DONE_FAIL|GONE_UNKNOWN.",
                   "config": {
                     "status_command": "pgrep -af ''02_train_llama_3_3_70b|/workspace/run.sh'' >/dev/null 2>&1 && { echo STATUS=ALIVE; exit 0; }; if grep -q RUN_SH_DONE /workspace/train.log 2>/dev/null && [ -f /workspace/adapter_out/adapter_config.json ]; then echo STATUS=DONE_OK; elif grep -q RUN_SH_FATAL /workspace/train.log 2>/dev/null; then echo STATUS=DONE_FAIL; else echo STATUS=GONE_UNKNOWN; fi; tail -n 3 /workspace/train.log 2>/dev/null"
                   },
                   "output_field": "probe",
                   "next_step": "classify"
                 },
                 "classify": {
                   "action": "classify_training_probe",
                   "description": "Parse probe.stdout STATUS= and reachable into a verdict and route via next_step. Pure logic, no DB.",
                   "config": {
                     "probe_step": "probe",
                     "alive_step": "reset_streak",
                     "unreachable_step": "bump_streak",
                     "complete_step": "mark_complete",
                     "failed_step": "mark_failed"
                   },
                   "output_field": "classify_result",
                   "next_step": "done"
                 },
                 "reset_streak": {
                   "action": "record_probe_streak",
                   "description": "Reachable (ALIVE): zero the consecutive-unreachable counter, then leave the box running for the next tick.",
                   "config": { "mode": "reset", "ok_step": "done" },
                   "output_field": "streak_result",
                   "next_step": "done"
                 },
                 "bump_streak": {
                   "action": "record_probe_streak",
                   "description": "Unreachable/no-status: increment the consecutive-unreachable counter; at the threshold treat as lost (route mark_failed -> decommission), otherwise leave for the next tick.",
                   "config": { "mode": "bump", "unreachable_threshold": 3, "lost_step": "mark_failed", "ok_step": "done" },
                   "output_field": "streak_result",
                   "next_step": "done"
                 },
                 "mark_complete": {
                   "action": "mark_training_run_terminal",
                   "description": "DONE_OK: transition training_runs running -> complete (stamp completed_at), then release the box.",
                   "config": { "status": "complete" },
                   "output_field": "mark_result",
                   "next_step": "decommission"
                 },
                 "mark_failed": {
                   "action": "mark_training_run_terminal",
                   "description": "DONE_FAIL / GONE_UNKNOWN / lost: transition training_runs running -> failed (+error_message), then release the box.",
                   "config": { "status": "failed", "error_message": "training run ended without a success marker (crash/OOM/unreachable) - marked failed by thunder-training-monitor" },
                   "output_field": "mark_result",
                   "next_step": "decommission"
                 },
                 "decommission": {
                   "action": "dispatch_thunder_decommission",
                   "description": "Release the Thunder instance (provisioning_id from input_data) now that the run is terminal.",
                   "config": {},
                   "output_field": "decommission_result",
                   "next_step": "done"
                 },
                 "done": {
                   "action": "complete_workflow",
                   "description": "Terminal. Surface the classify verdict to the orchestrator (which ignores it).",
                   "config": { "output_fields": ["classify_result"] }
                 }
               }
             }
           }'::jsonb,
           jsonb_build_object(
                   'required', jsonb_build_array('provisioning_id', 'training_run_id'),
                   'optional', jsonb_build_array('thunder_instance_id', 'instance_ip')
           ),
           jsonb_build_object(
                   'produces', jsonb_build_array('classify_result')
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

-- Validate the workflow JSON parsed and the start_step + key steps exist.
DO $$
DECLARE
steps jsonb;
BEGIN
SELECT default_config->'workflow'->'steps' INTO steps
FROM agent_definitions
WHERE type = 'thunder-training-monitor-worker'
  AND (is_snapshot IS NULL OR is_snapshot = false) AND deleted_at IS NULL
ORDER BY version DESC LIMIT 1;

IF steps IS NULL
       OR steps->'probe' IS NULL
       OR steps->'classify' IS NULL
       OR steps->'mark_complete' IS NULL
       OR steps->'mark_failed' IS NULL
       OR steps->'decommission' IS NULL
       OR steps->'done' IS NULL THEN
        RAISE EXCEPTION 'migration 107: worker workflow steps missing/unparsed';
END IF;
END $$;

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually):
-- ───────────────────────────────────────────────────────────────────────────
-- \echo worker workflow:
-- SELECT jsonb_pretty(default_config) FROM agent_definitions
--  WHERE type='thunder-training-monitor-worker' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
-- \echo worker image (should match model-trainer):
-- SELECT type, image_repository, image_tag FROM agent_definitions
--  WHERE type IN ('thunder-training-monitor-worker','model-trainer') AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
