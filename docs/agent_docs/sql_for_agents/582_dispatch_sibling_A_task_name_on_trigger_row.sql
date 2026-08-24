-- 582 — dispatch concurrency A of C: seed input_data.task_name on the trigger row (INERT)
--
-- First of the three-step config-only change that gives the dispatch pipeline a second
-- concurrent turn (dispatch_throughput PLAN §4, owner-authorised at N=2 — D1 default,
-- rulings recorded 2026-08-21; RESEARCH doc §10). The scheduler delivers a task row's
-- input_data into the fired orchestration (cmd/scheduler/main.go:192, fireTrigger:520),
-- so this key becomes readable as collected_data->input_data->task_name by every step.
-- NOTHING reads it until 583 applies — this migration is deliberately inert, and MUST
-- apply before 583 (a query_database param resolving to nil is a step ERROR).
--
-- ⚠ ORDER: 582 → 583 → 584. Numeric order preserves it under any runner; by-hand
-- application must keep it too. 584 is the activation; 582/583 are inert-forward and
-- need no rollback to restore old behaviour.
--
-- COUNCIL: migrations are in scope since 2026-08-19 (bugs_open/314 widening) —
-- submitted with the 583/584 set as one coherent change; commit carries
-- Council-Submitted. Register: WDS entry updated in the same commit.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
  WHERE name = 'build-pipeline-trigger' AND enabled
    AND concurrency_group = 'dispatch';
  IF n <> 1 THEN
    RAISE EXCEPTION '582 pre-flight: expected exactly 1 enabled build-pipeline-trigger row in dispatch group, found %', n;
  END IF;
  -- Refuse a drifted row: input_data must still be the empty object this was written against.
  SELECT count(*) INTO n FROM scheduled_tasks
  WHERE name = 'build-pipeline-trigger' AND COALESCE(input_data, '{}'::jsonb) = '{}'::jsonb;
  IF n <> 1 THEN
    RAISE EXCEPTION '582 pre-flight: build-pipeline-trigger input_data is no longer {} — re-read the row before applying';
  END IF;
END $$;

UPDATE scheduled_tasks
SET input_data = jsonb_set(COALESCE(input_data, '{}'::jsonb), '{task_name}', '"build-pipeline-trigger"'),
    updated_at = now()
WHERE name = 'build-pipeline-trigger';

DO $$
DECLARE v text;
BEGIN
  SELECT input_data->>'task_name' INTO v FROM scheduled_tasks WHERE name = 'build-pipeline-trigger';
  IF v IS DISTINCT FROM 'build-pipeline-trigger' THEN
    RAISE EXCEPTION '582 post-check: task_name is % — write did not land', v;
  END IF;
END $$;

COMMIT;

-- ROLLBACK (only needed for tidiness; the key is inert until 583):
--   UPDATE scheduled_tasks SET input_data = input_data - 'task_name'
--   WHERE name = 'build-pipeline-trigger';
