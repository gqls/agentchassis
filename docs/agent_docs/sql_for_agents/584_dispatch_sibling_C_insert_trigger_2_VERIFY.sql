-- 584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql — RE-RUNNABLE, read-only, DO/RAISE.
--
-- Answers council round 2 (corr db9b7cbf, debug_historian MEDIUM): "the migration file itself
-- should carry a post-apply verify step … not rely solely on a manual one-off check". Run any
-- time (the migration runner's SIDECAR_RE excludes _VERIFY files, so this never auto-applies).
-- Hardened 2026-08-25 post-approval per council r3 advisories: liveness no longer keys on
-- owner_agent_type (unreliable column), and assertion 6 blocks a third sibling clone:
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
--     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql
--
-- Exit 0 = all seven assertions hold; a RAISE names the one that failed. Proven to fire
-- 2026-08-25 by inverting assertion 1's predicate on a scratch copy (RUNBOOK).
--
-- ⚠ On this scheduler (fire-and-forget: runTick → fireTrigger → stampCompleted sets BOTH stamps
-- at fire, since 892a289e9 2026-03-17 — bugs_open/398, LANDMINES 2026-08-25) a row's stamps prove
-- only that it FIRED, never that a run completed. Liveness is therefore asserted from
-- orchestration_states, which the scheduler cannot fake; the double-handle census (5) is the
-- real safety property of N=2, and it is what the council asked to see.

DO $$
DECLARE n int; m int; z int; v text; hardcoded int;
BEGIN
  -- 1/7 GATE PARITY: every trigger row — disabled ones included, they are the rollback
  --     path — agrees on the dispatch GATE (pre_query/agent/topic/fire_message). The
  --     LEVER columns (enabled, interval_seconds) are asserted separately in 2/7: since
  --     ruling B (2026-08-26, migration 637) they legitimately differ between rows.
  SELECT count(*) INTO m FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  SELECT count(DISTINCT (md5(coalesce(pre_query,'')), target_agent_type, target_topic, fire_message))
    INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
  IF m >= 2 AND n <> 1 THEN
    RAISE EXCEPTION '584 VERIFY 1/7 GATE PARITY: % trigger rows carry % distinct gate configs — a by-name UPDATE missed a sibling (LANDMINES 2026-08-25)', m, n;
  END IF;

  -- 2/7 LEVER (owner ruling B, 2026-08-26): exactly ONE enabled trigger row, at
  --     interval_seconds >= 30. Option C (interval 25, ~3x) is GATED on the D4 spend
  --     governor — when C is ruled, edit this assertion in the same commit (lockstep).
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled;
  IF n <> 1 THEN
    RAISE EXCEPTION '584 VERIFY 2/7 LEVER: % enabled trigger rows, ruling B (2026-08-26) says exactly 1 — a rollback to the sibling state must re-edit this assertion in lockstep (637_..._ROLLBACK.sql)', n;
  END IF;
  SELECT interval_seconds INTO n FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled;
  IF n < 30 THEN
    RAISE EXCEPTION '584 VERIFY 2/7 LEVER: enabled row interval_seconds = % — below 30 is option C, gated on the D4 spend governor (owner ruling 2026-08-26)', n;
  END IF;

  -- 3/7 IDENTITY: each row's input_data.task_name is its own name (582 / 584).
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name LIKE 'build-pipeline-trigger%' AND input_data->>'task_name' IS DISTINCT FROM name;
  IF n <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 3/7 IDENTITY: % row(s) whose input_data.task_name is not their own name', n;
  END IF;

  -- 4/7 STAMPS: no hardcoded notify stamp anywhere in live agent config — WHOLE-TEXT census,
  --     which sees sub_workflow/substep text (the round-1 guardian's "fourth writer" fear).
  SELECT count(*) INTO hardcoded FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''build-pipeline-trigger%';
  IF hardcoded <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 4/7 STAMPS: % active agent config(s) still carry a hardcoded notify stamp', hardcoded;
  END IF;
  -- positive control: the same scan must see substep-level text, or it is blind and this passes vacuously.
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%spawn_handler%';
  IF n = 0 THEN
    RAISE EXCEPTION '584 VERIFY 4/7 CONTROL: the whole-text census sees no substep text at all — the CHECK is broken, not the config';
  END IF;

  -- 5/7 LIVENESS: every ENABLED trigger row produced a trigger orchestration carrying its own
  --     task_name in the last 15 minutes (cadence is ~90 s; idle ticks still create a run).
  FOR v IN SELECT name FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%' AND enabled LOOP
    -- ⚠ deliberately NOT filtered on owner_agent_type: that column reads ZERO for some
    -- demonstrably active agents (LANDMINES; council r3 editquality advisory, 2026-08-25).
    -- input_data.task_name is carried ONLY by the trigger's runs and the loops it spawns,
    -- so its presence alone proves the row's fire was delivered and became an orchestration.
    SELECT count(*) INTO n FROM orchestration_states
     WHERE collected_data->'input_data'->>'task_name' = v
       AND created_at > now() - interval '15 minutes';
    IF n = 0 THEN
      RAISE EXCEPTION '584 VERIFY 5/7 LIVENESS: enabled row % has no trigger orchestration in 15 min (scheduler down, row disabled upstream, or task_name not delivered)', v;
    END IF;
  END LOOP;

  -- 6/7 NO DOUBLE-HANDLE: no work item had two handler orchestrations ALIVE AT ONCE in 24 h.
  --     (sequential retries — handlers = attempt_count — are legitimate and are not counted)
  --     Narrowed 2026-08-26 (NOTES): for a STALE-REAPED handler, updated_at is the REAP stamp,
  --     not end-of-life — it sat as a zombie from the moment its request died until the reaper
  --     fired, so raw interval overlap counts a successor that legitimately re-claimed the
  --     released item (first live case: pair a52ac67f/d0f7ea9e, 2m overlap = the reap lag).
  --     Excluded shape: first-started member stale-reaped AND second started > 10 min later
  --     (a real claim race starts within seconds — first-claim p50 17.7 s, 2026-08-25); a
  --     stale-reaped member with a near-simultaneous partner still COUNTS. Excluded pairs are
  --     reported as a NOTICE, never silently dropped.
  --     Widened 2026-08-27 (NOTES): TWO reapers stamp stale handlers with different spellings —
  --     the workflow-timeout reaper writes 'Orchestration stale — running for …' and the
  --     step-level reaper writes 'reaper: stale EXECUTING_STEP for >4h; step=…'. Second live
  --     case (pair 0d699d65/fb7e9e0f, item 61265835): handler wedged at suggest_related_pages,
  --     claim released by the claim-level clock, successor re-claimed 2.6 h later via the normal
  --     atomic path and completed; the first member carried the step-reaper spelling and the
  --     exclusion missed it — a false RAISE, proven serial at the item's claim history.
  WITH loops AS (
    SELECT orchestration_id FROM orchestration_states
     WHERE created_at > now() - interval '24 hours' AND owner_agent_type = 'build-dispatch-loop'),
  h AS (
    SELECT o.orchestration_id, o.collected_data->'input_data'->>'work_item_id' wi, o.created_at s, o.updated_at e,
           o.status, o.error
      FROM orchestration_states o JOIN loops l ON l.orchestration_id = o.parent_orchestration_id
     WHERE o.collected_data->'input_data'->>'work_item_id' IS NOT NULL),
  pairs AS (
    SELECT (CASE WHEN a.s <= b.s THEN a.status ELSE b.status END = 'FAILED'
            AND (CASE WHEN a.s <= b.s THEN a.error ELSE b.error END LIKE 'Orchestration stale%'
                 OR CASE WHEN a.s <= b.s THEN a.error ELSE b.error END LIKE 'reaper: stale EXECUTING_STEP%')
            AND abs(EXTRACT(epoch FROM a.s - b.s)) > 600) AS zombie_tail
      FROM h a JOIN h b
        ON a.wi = b.wi AND a.orchestration_id < b.orchestration_id AND a.s < b.e AND b.s < a.e)
  SELECT count(*) FILTER (WHERE NOT zombie_tail), count(*) FILTER (WHERE zombie_tail)
    INTO n, z FROM pairs;
  IF n <> 0 THEN
    RAISE EXCEPTION '584 VERIFY 6/7 DOUBLE-HANDLE: % overlapping handler pair(s) on one work item in 24 h — the atomic claim did not hold', n;
  END IF;
  IF z <> 0 THEN
    RAISE NOTICE '584 VERIFY 6/7: % zombie-tail pair(s) excluded (first-started member stale-reaped; its updated_at is the reap stamp, not end-of-life — NOTES 2026-08-26)', z;
  END IF;

  -- 7/7 NO THIRD SIBLING (council r3 guardian/architecture advisory, 2026-08-25): the clone
  --     pattern must not repeat — a third trigger row means a session copied 584's shape.
  --     The sanctioned paths are interval_seconds or the D9 per-task executions fix (bugs_open/398).
  IF m > 2 THEN
    RAISE EXCEPTION '584 VERIFY 7/7 THIRD-SIBLING: % build-pipeline-trigger rows exist — 584 is a stopgap awaiting bugs_open/398, not a sanctioned pattern; use interval_seconds or D9', m;
  END IF;

  RAISE NOTICE '584 VERIFY: all 7 hold — gate parity across % row(s), lever = ruling B (1 enabled row, interval >= 30), identity, 0 hardcoded stamps (control passing), liveness, 0 double-handles in 24 h, no third sibling', m;
END $$;

-- Informational, never fails: co-pick cost table. Under ruling B (one enabled row) the cross-row
-- co-pick column reads 0 for post-637 loops by construction — the meaningful figure is then
-- claims_lost/(won+lost), which 2026-08-25 measured at 39% under the phase-locked sibling.
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
