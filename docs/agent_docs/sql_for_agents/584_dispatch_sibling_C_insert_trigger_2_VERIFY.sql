-- 584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql — RE-RUNNABLE, read-only, DO/RAISE.
--
-- Answers council round 2 (corr db9b7cbf, debug_historian MEDIUM): "the migration file itself
-- should carry a post-apply verify step … not rely solely on a manual one-off check". Run any
-- time (the migration runner's SIDECAR_RE excludes _VERIFY files, so this never auto-applies):
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql
--
-- Exit 0 = all five assertions hold; a RAISE names the one that failed. Proven to fire
-- 2026-08-25 by inverting assertion 1's predicate on a scratch copy (RUNBOOK).
--
-- ⚠ On this scheduler (fire-and-forget: runTick → fireTrigger → stampCompleted sets BOTH stamps
-- at fire, since 892a289e9 2026-03-17 — bugs_open/398, LANDMINES 2026-08-25) a row's stamps prove
-- only that it FIRED, never that a run completed. Liveness is therefore asserted from
-- orchestration_states, which the scheduler cannot fake; the double-handle census (5) is the
-- real safety property of N=2, and it is what the council asked to see.

DO $$
DECLARE n int; m int; v text; hardcoded int;
BEGIN
  -- 1/5 PARITY: every trigger row agrees on every column that shapes a dispatch turn.
  --     (the sibling was INSERT..SELECTed once; a by-name UPDATE on the original desyncs it silently)
  SELECT count(*) INTO m FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  SELECT count(DISTINCT (md5(coalesce(pre_query,'')), interval_seconds, target_agent_type, target_topic,
                         timeout_seconds, concurrency_group, max_concurrent, fire_message, enabled))
    INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF m >= 2 AND n <> 1 THEN
    RAISE EXCEPTION '584 VERIFY 1/5 PARITY: % trigger rows carry % distinct configs — a by-name UPDATE missed the sibling (LANDMINES 2026-08-25)', m, n;
  END IF;

  -- 2/5 IDENTITY: each row's input_data.task_name is its own name (582 / 584).
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name LIKE 'build-pipeline-trigger%' AND input_data->>'task_name' IS DISTINCT FROM name;
  IF n <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 2/5 IDENTITY: % row(s) whose input_data.task_name is not their own name', n;
  END IF;

  -- 3/5 STAMPS: no hardcoded notify stamp anywhere in live agent config — WHOLE-TEXT census,
  --     which sees sub_workflow/substep text (the round-1 guardian's "fourth writer" fear).
  SELECT count(*) INTO hardcoded FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''build-pipeline-trigger%';
  IF hardcoded <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 3/5 STAMPS: % active agent config(s) still carry a hardcoded notify stamp', hardcoded;
  END IF;
  -- positive control: the same scan must see substep-level text, or it is blind and this passes vacuously.
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%spawn_handler%';
  IF n = 0 THEN
    RAISE EXCEPTION '584 VERIFY 3/5 CONTROL: the whole-text census sees no substep text at all — the CHECK is broken, not the config';
  END IF;

  -- 4/5 LIVENESS: every ENABLED trigger row produced a trigger orchestration carrying its own
  --     task_name in the last 15 minutes (cadence is ~90 s; idle ticks still create a run).
  FOR v IN SELECT name FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled LOOP
    SELECT count(*) INTO n FROM orchestration_states
     WHERE owner_agent_type = 'build-pipeline-trigger'
       AND collected_data->'input_data'->>'task_name' = v
       AND created_at > now() - interval '15 minutes';
    IF n = 0 THEN
      RAISE EXCEPTION '584 VERIFY 4/5 LIVENESS: enabled row % has no trigger orchestration in 15 min (scheduler down, row disabled upstream, or task_name not delivered)', v;
    END IF;
  END LOOP;

  -- 5/5 NO DOUBLE-HANDLE: no work item had two handler orchestrations ALIVE AT ONCE in 24 h.
  --     (sequential retries — handlers = attempt_count — are legitimate and are not counted)
  WITH loops AS (
    SELECT orchestration_id FROM orchestration_states
     WHERE created_at > now() - interval '24 hours' AND owner_agent_type = 'build-dispatch-loop'),
  h AS (
    SELECT o.orchestration_id, o.collected_data->'input_data'->>'work_item_id' wi, o.created_at s, o.updated_at e
      FROM orchestration_states o JOIN loops l ON l.orchestration_id = o.parent_orchestration_id
     WHERE o.collected_data->'input_data'->>'work_item_id' IS NOT NULL)
  SELECT count(*) INTO n FROM h a JOIN h b
      ON a.wi = b.wi AND a.orchestration_id < b.orchestration_id AND a.s < b.e AND b.s < a.e;
  IF n <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 5/5 DOUBLE-HANDLE: % overlapping handler pair(s) on one work item in 24 h — the atomic claim did not hold', n;
  END IF;

  RAISE NOTICE '584 VERIFY: all 5 hold — parity across % row(s), identity, 0 hardcoded stamps (control passing), liveness, 0 double-handles in 24 h', m;
END $$;

-- Informational, never fails: the COST of the 1 s phase lock between the two rows over 24 h —
-- how often the second fire co-picks the first fire's site, and the share of claim attempts lost.
WITH l AS (
  SELECT orchestration_id, site_id, created_at s, updated_at e,
         collected_data->'input_data'->>'task_name' tn,
         (SELECT count(*) FROM jsonb_object_keys(collected_data) k
           WHERE k ~ '^claim_result_[0-9]+$' AND collected_data->k->>'claimed' = 'true')  claimed_ok,
         (SELECT count(*) FROM jsonb_object_keys(collected_data) k
           WHERE k ~ '^claim_result_[0-9]+$' AND collected_data->k->>'claimed' = 'false') claimed_lost,
         CASE WHEN jsonb_typeof(collected_data->'pending'->'items') = 'array'
              THEN jsonb_array_length(collected_data->'pending'->'items') ELSE 0 END loaded
    FROM orchestration_states
   WHERE created_at > now() - interval '24 hours' AND owner_agent_type = 'build-dispatch-loop')
SELECT tn AS row_task_name, count(*) AS loops,
       count(*) FILTER (WHERE loaded > 0) AS loaded,
       count(*) FILTER (WHERE loaded > 0 AND EXISTS (
         SELECT 1 FROM l b WHERE b.site_id = a.site_id AND b.tn <> a.tn
            AND a.s < b.e AND b.s < a.e AND abs(EXTRACT(epoch FROM a.s - b.s)) < 15)) AS co_picked_within_15s,
       sum(claimed_ok) AS items_claimed, sum(claimed_lost) AS claims_lost,
       round(100.0 * sum(claimed_lost) / NULLIF(sum(claimed_ok) + sum(claimed_lost), 0), 1) AS lost_pct
  FROM l a GROUP BY tn ORDER BY tn;
