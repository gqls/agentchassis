-- 584 — dispatch concurrency C of C: the sibling row (THE ACTIVATION — N=2)
--
-- Inserts build-pipeline-trigger-2: same group/agent/topic/pre_query as the original,
-- its own name in input_data.task_name so (post-583) its completions stamp ITS row.
-- Each scheduled_tasks row is its own single-flight slot (loadDueTasks per-row guard),
-- so two rows = two concurrent dispatch turns. Per-site serialisation survives in
-- find_dispatchable_site's NOT EXISTS same-site claimed exclusion + the atomic
-- claim_work_item; two turns picking the same site in the same instant cost a wasted
-- spawn, not a double-dispatch (PLAN §4 safety argument — to be induced and verified,
-- recorded in NOTES).
--
-- AUTHORISATION: owner decision D1 default "stop at N=2" + rulings session 2026-08-21;
-- spend note: ~N× handler spawns while backlog exists; the LLM spend governor (D4) is
-- the standing prerequisite for N>2, NOT for N=2. ROLLBACK = one statement, instant:
--   UPDATE scheduled_tasks SET enabled = false WHERE name = 'build-pipeline-trigger-2';
--
-- VERIFY AT THE ARTEFACT (RUNBOOK; STARTER §7): within minutes,
--   SELECT name, last_triggered_at, last_completed_at FROM scheduled_tasks
--   WHERE name LIKE 'build-pipeline-trigger%';
-- must show trigger-2 firing AND its stamps landing on ITS OWN name; the per-MINUTE
-- concurrency meter must read 2 distinct sites in busy minutes (never trust a 5-min
-- bucket). If the meter still reads 1, the sibling never fired — check its stamps.

BEGIN;

DO $$
DECLARE n int; v text;
BEGIN
  -- 583 must be live (params on the loop's notify — the last edit in the chain).
  SELECT count(*) INTO n FROM agent_definitions
  WHERE type = 'build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config #>> '{workflow,steps,notify_scheduler,config,params,0}' = 'input_data.task_name';
  IF n <> 1 THEN
    RAISE EXCEPTION '584 pre-flight: 583 not applied — the sibling would fire but never be stamped (300s fallback) and would stamp the original row on completion';
  END IF;
  SELECT input_data->>'task_name' INTO v FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF v IS DISTINCT FROM 'build-pipeline-trigger' THEN
    RAISE EXCEPTION '584 pre-flight: 582 not applied';
  END IF;
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'build-pipeline-trigger-2';
  IF n <> 0 THEN
    RAISE EXCEPTION '584 pre-flight: build-pipeline-trigger-2 already exists';
  END IF;
END $$;

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic, input_data,
   concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds, fire_message)
SELECT
  'build-pipeline-trigger-2',
  'Sibling dispatch turn #2 (dispatch_throughput 582-584, N=2). Same selector/loop as build-pipeline-trigger; its own single-flight slot. Disable this row to return to N=1.',
  interval_seconds, target_agent_type, target_topic,
  '{"task_name": "build-pipeline-trigger-2"}'::jsonb,
  concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds, fire_message
FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
  WHERE name = 'build-pipeline-trigger-2' AND enabled AND concurrency_group = 'dispatch'
    AND input_data->>'task_name' = 'build-pipeline-trigger-2'
    AND pre_query IS NOT NULL AND pre_query <> '';
  IF n <> 1 THEN
    RAISE EXCEPTION '584 post-check: sibling row not in the expected shape';
  END IF;
END $$;

COMMIT;
