-- 210_report_pipeline_scheduled_tasks.sql
-- Gripper dossier pilot — the two scheduled tasks that drive the lane.
-- Workstream: docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/
-- Design of record: DESIGN_2026-07-24_gripper_dossier_pilot.md §3 (tasks), §4.
--
-- POST-IMAGE, and AFTER seed 209. Both tasks ship DISABLED: seeding a task is
-- not the same as starting a pipeline, and the enable order is part of the
-- test plan (DESIGN §4) — enable report-dispatch FIRST, prove the handler on a
-- hand-inserted work item, and only then enable report-request-pull so real
-- visitor requests start arriving. Enabling both at once means the first live
-- request is also the first test.
--
-- INTERVAL ARITHMETIC. The scheduler ticks every 30s, so the effective period
-- is interval_seconds + 30 (bugs_open/030: 'effective period = interval +
-- TICK'). 270 -> fires about every 5 minutes; 90 -> about every 2 minutes.
-- Both are deliberately slower than they could be: a visitor is told at submit
-- time that the dossier takes a few minutes, and a tighter loop buys nothing
-- but load.
--
-- CONCURRENCY GROUPS are isolated per lane, so a stalled report build cannot
-- block the diagnose or build dispatch lanes (and vice versa). Do NOT put
-- these in the shared 'dispatch' group.
--
-- Both pre_queries GATE: they return no rows when there is nothing to do, and
-- the scheduler then does not fire. The dispatch gate deliberately also
-- matches stuck 'reporting' rows, so the loop still fires to run its own
-- reaper when the only outstanding item is one whose pod died — a gate that
-- looked only for awaiting_report would leave a dead build stuck forever.
--
-- IDEMPOTENT: ON CONFLICT (name) DO UPDATE. The enabled flag is deliberately
-- NOT overwritten on conflict — re-running this seed must never silently
-- re-disable a lane the operator has enabled.
--
-- CORRECTED 2026-07-27 — target_topic MUST be 'system.agent.scheduled.requests'.
-- This file originally named 'system.agent.generic.requests', which is the
-- scheduled_tasks column DEFAULT and is a topic NOTHING CONSUMES. Measured live
-- the day this was found: 18 of 18 enabled tasks use .scheduled.; the .generic.
-- topic held 7 tasks of which the only enabled one was this bug.
--
-- It fails SILENTLY and looks perfectly healthy from the producer side. The
-- scheduler logged "Successfully produced message" and "Triggered task", and
-- stamped last_triggered_at AND last_completed_at — which is the NORMAL
-- fire-and-forget path (cmd/scheduler/main.go:287-296), not a no-op marker, so
-- the timestamps cannot distinguish "fired into a live topic" from "fired into
-- a void". The only discriminating evidence is downstream: zero rows in
-- orchestration_states for the target agent_type, and zero mention of the
-- correlation_id in the chassis log. Check THAT, not the scheduler's own logs.

\set ON_ERROR_STOP on

BEGIN;

-- ============================================================================
-- 1. report-request-pull — pull visitor requests from the island.
-- ============================================================================
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, pre_query, enabled,
    timeout_seconds
) VALUES (
    'report-request-pull',
    'Pull pending gripper-dossier requests from each configured report island '
    || 'into site_work_items. Gated on at least one site having '
    || 'deploy_config.report_island, so it costs nothing until 208 is applied.',
    270,
    'report-request-collector',
    'system.agent.scheduled.requests',
    '{"action": "orchestrate", "config": {"agent_type": "report-request-collector"}, "input_data": {}}'::jsonb,
    'report-pull',
    1,
    'SELECT count(*)::text AS island_sites FROM sites WHERE status IN (''active'', ''deployed'') AND deploy_config ? ''report_island'' HAVING count(*) > 0',
    false,
    600
)
ON CONFLICT (name) DO UPDATE SET
    description       = EXCLUDED.description,
    interval_seconds  = EXCLUDED.interval_seconds,
    target_agent_type = EXCLUDED.target_agent_type,
    target_topic      = EXCLUDED.target_topic,
    input_data        = EXCLUDED.input_data,
    concurrency_group = EXCLUDED.concurrency_group,
    max_concurrent    = EXCLUDED.max_concurrent,
    pre_query         = EXCLUDED.pre_query,
    timeout_seconds   = EXCLUDED.timeout_seconds,
    updated_at        = NOW();
    -- enabled deliberately NOT updated — see header.

-- ============================================================================
-- 2. report-dispatch — claim and build one queued dossier per tick.
-- ============================================================================
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, pre_query, enabled,
    timeout_seconds
) VALUES (
    'report-dispatch',
    'Claim one awaiting_report work item and run report-builder. Also fires '
    || 'when a ''reporting'' item is stuck, so the loop can run its own reaper.',
    90,
    'report-dispatch-loop',
    'system.agent.scheduled.requests',
    '{"action": "orchestrate", "config": {"agent_type": "report-dispatch-loop"}, "input_data": {}}'::jsonb,
    'report-dispatch',
    1,
    'SELECT count(*)::text AS queued_reports FROM site_work_items WHERE pipeline = ''reports'' AND item_type = ''report_request'' AND ((status = ''awaiting_report'' AND attempt_count < max_attempts) OR (status = ''reporting'' AND claimed_at < NOW() - INTERVAL ''30 minutes'')) HAVING count(*) > 0',
    false,
    2400
)
ON CONFLICT (name) DO UPDATE SET
    description       = EXCLUDED.description,
    interval_seconds  = EXCLUDED.interval_seconds,
    target_agent_type = EXCLUDED.target_agent_type,
    target_topic      = EXCLUDED.target_topic,
    input_data        = EXCLUDED.input_data,
    concurrency_group = EXCLUDED.concurrency_group,
    max_concurrent    = EXCLUDED.max_concurrent,
    pre_query         = EXCLUDED.pre_query,
    timeout_seconds   = EXCLUDED.timeout_seconds,
    updated_at        = NOW();
    -- enabled deliberately NOT updated — see header.

-- ============================================================================
-- Assertions. The pre_queries are EXECUTED, not merely stored: a pre_query
-- with a typo fails silently at tick time (the task simply never fires), which
-- is the hardest kind of dead pipeline to notice. Running them here proves
-- they parse and return the gate shape.
-- ============================================================================
DO $$
DECLARE
    n INTEGER;
    q TEXT;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks
    WHERE name IN ('report-request-pull', 'report-dispatch');
    IF n <> 2 THEN
        RAISE EXCEPTION 'expected 2 report tasks, found %', n;
    END IF;

    SELECT count(*) INTO n FROM scheduled_tasks
    WHERE name IN ('report-request-pull', 'report-dispatch')
      AND concurrency_group IN ('report-pull', 'report-dispatch');
    IF n <> 2 THEN
        RAISE EXCEPTION 'report tasks must stay in their own concurrency groups, not the shared dispatch group';
    END IF;

    FOR q IN
        SELECT pre_query FROM scheduled_tasks
        WHERE name IN ('report-request-pull', 'report-dispatch')
    LOOP
        BEGIN
            EXECUTE q;
        EXCEPTION WHEN OTHERS THEN
            RAISE EXCEPTION 'pre_query does not execute (a task with a broken gate never fires and looks merely idle): % — %', q, SQLERRM;
        END;
    END LOOP;

    RAISE NOTICE 'report scheduled tasks seeded (disabled); both pre_queries execute';
END $$;

COMMIT;

-- ============================================================================
-- ENABLE (deliberately not run by this seed — DESIGN §4 order):
--
--   -- 1. dispatch first, and prove the handler on a hand-inserted item:
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'report-dispatch';
--
--   -- 2. only once a real dossier has been built end to end:
--   UPDATE scheduled_tasks SET enabled = true WHERE name = 'report-request-pull';
--
-- DISABLE (the kill switch — stops new work without touching data):
--   UPDATE scheduled_tasks SET enabled = false
--   WHERE name IN ('report-request-pull', 'report-dispatch');
-- ============================================================================
