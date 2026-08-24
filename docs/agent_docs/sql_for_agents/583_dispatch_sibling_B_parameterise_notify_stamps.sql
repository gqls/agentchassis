-- 583 — dispatch concurrency B of C: the notify stamps become per-row (three queries, not two)
--
-- The trap this defuses (dispatch_throughput PLAN §4, STARTER §4-L1b): THREE steps —
-- build-pipeline-trigger's notify_scheduler AND notify_scheduler_idle, plus
-- build-dispatch-loop's notify_scheduler — all run the byte-identical hardcode
--   UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'build-pipeline-trigger'
-- (one md5, 64db3df8551b60a2098443ce00569604, read live 2026-08-24). A sibling task row
-- under that hardcode is never stamped (falls through to timeout_seconds, fires at 300s
-- not 60) while every sibling completion stamps the ORIGINAL row, releasing it early.
-- ⚠ The PLAN named only two queries; the idle path was found by counting occurrences in
-- the live config (2 in the trigger + 1 in the loop) — the census habit, dated.
--
-- The fix: WHERE name = $1 with params ["input_data.task_name"]. QueryDatabaseAction
-- (database_actions.go:31-73) resolves params from collected_data with an input_data.
-- fallback and errors on nil — which is why 582 MUST be applied first (induced before
-- applying: the DO-guard below RAISEs against a task_name-less row; verified 2026-08-24).
-- The loop learns its task_name via call_dispatch.input_mapping, edited here too.
-- In-flight orchestrations carry their snapshotted workflow_plan, so runs started
-- before this migration keep the old hardcode and stamp the original row — correct,
-- because they WERE the original row's execution.
--
-- CONSUMERS: the scheduler's per-row single-flight guard (loadDueTasks) — stamps now
-- land on the row that fired. No behaviour change while only one row exists (the
-- original row's task_name IS 'build-pipeline-trigger'); 584 is the activation.

BEGIN;

-- Cross-file ordering guard: 582 must have applied.
DO $$
DECLARE v text;
BEGIN
  SELECT input_data->>'task_name' INTO v FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF v IS DISTINCT FROM 'build-pipeline-trigger' THEN
    RAISE EXCEPTION '583 pre-flight: 582 not applied (task_name=%) — apply 582 first or the notify params resolve to nil and every notify step ERRORS', v;
  END IF;
END $$;

-- Drift guard: all three queries must still be the byte-identical hardcode.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type = 'build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND md5(default_config #>> '{workflow,steps,notify_scheduler,config,query}') = '64db3df8551b60a2098443ce00569604'
    AND md5(default_config #>> '{workflow,steps,notify_scheduler_idle,config,query}') = '64db3df8551b60a2098443ce00569604'
    AND md5(default_config #>> '{workflow,steps,call_dispatch,config,input_mapping}') = '956290a3aa2998c2c08c8f9336e040ed';
  IF n <> 1 THEN
    RAISE EXCEPTION '583 pre-flight: trigger row drifted from the shapes this was written against — re-read before applying';
  END IF;
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type = 'build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND md5(default_config #>> '{workflow,steps,notify_scheduler,config,query}') = '64db3df8551b60a2098443ce00569604';
  IF n <> 1 THEN
    RAISE EXCEPTION '583 pre-flight: loop row drifted — re-read before applying';
  END IF;
END $$;

SELECT snapshot_agent('build-pipeline-trigger', 'dispatch sibling B — parameterise notify stamps + map task_name (583)');
SELECT snapshot_agent('build-dispatch-loop',    'dispatch sibling B — parameterise notify stamp (583)');

-- Trigger: notify_scheduler
UPDATE agent_definitions
SET default_config =
    jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,notify_scheduler,config,query}',
        '"UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1"', false),
      '{workflow,steps,notify_scheduler,config,params}',
      '["input_data.task_name"]', true),
    updated_at = now()
WHERE type = 'build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Trigger: notify_scheduler_idle
UPDATE agent_definitions
SET default_config =
    jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,notify_scheduler_idle,config,query}',
        '"UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1"', false),
      '{workflow,steps,notify_scheduler_idle,config,params}',
      '["input_data.task_name"]', true),
    updated_at = now()
WHERE type = 'build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Trigger: call_dispatch passes task_name through to the loop
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,call_dispatch,config,input_mapping,task_name}',
      '"input_data.task_name"', true),
    updated_at = now()
WHERE type = 'build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Loop: notify_scheduler
UPDATE agent_definitions
SET default_config =
    jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,notify_scheduler,config,query}',
        '"UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = $1"', false),
      '{workflow,steps,notify_scheduler,config,params}',
      '["input_data.task_name"]', true),
    updated_at = now()
WHERE type = 'build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Post-check: no hardcoded stamp remains; params present on all three; mapping carries task_name.
DO $$
DECLARE bad int; good int;
BEGIN
  SELECT count(*) INTO bad FROM agent_definitions
  WHERE type IN ('build-pipeline-trigger','build-dispatch-loop')
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config::text LIKE '%WHERE name = ''build-pipeline-trigger''%';
  IF bad <> 0 THEN
    RAISE EXCEPTION '583 post-check: % row(s) still carry the hardcoded stamp', bad;
  END IF;
  SELECT count(*) INTO good FROM agent_definitions
  WHERE type = 'build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config #>> '{workflow,steps,notify_scheduler,config,params,0}' = 'input_data.task_name'
    AND default_config #>> '{workflow,steps,notify_scheduler_idle,config,params,0}' = 'input_data.task_name'
    AND default_config #>> '{workflow,steps,call_dispatch,config,input_mapping,task_name}' = 'input_data.task_name';
  IF good <> 1 THEN
    RAISE EXCEPTION '583 post-check: trigger edits did not all land';
  END IF;
END $$;

COMMIT;

-- ROLLBACK: restore the two snapshot_agent snapshots taken above (preferred), or
-- jsonb_set the three queries back to the hardcode and remove the params keys and
-- the input_mapping.task_name key.
