-- ============================================================================
-- 028_thunder_reaper.sql
--
-- Phase 3.5: thunder-reaper. Finds Thunder Compute instances that have run
-- longer than their max_uptime_hours and dispatches decommission_instance
-- to thunder-adapter for each.
--
-- Architecture:
--   scheduled_tasks (thunder-reaper, every 15 min)
--      → fires generic agent runner with thunder-reaper agent_definition
--      → workflow step dispatch_decommission (new chassis action)
--      → action publishes to system.adapter.thunder.requests
--      → thunder-adapter handleDecommissionInstance does the work
--      → response back to the reaper's responses_topic, workflow completes
--
-- The reaper handles ONE timed-out instance per tick (pre_query LIMIT 1).
-- With thunder_config.max_concurrent_instances = 2 this is sufficient:
-- worst case 2 stuck instances clear in 30 min. If the cap is raised, lower
-- interval_seconds or add a fan-out custom action later.
--
-- Idempotency: the underlying decommission_instance action is idempotent
-- (rows already in decommissioned/failed/reaped return cached cost without
-- re-calling APIs). Multiple reaper ticks on the same instance during the
-- decommission window are safe — Thunder API delete and Secret delete are
-- also idempotent on 404.
-- ============================================================================

BEGIN;

-- ── 1. thunder-reaper agent_definition ──
-- Two-step task workflow. processing_mode=task means runs in-chassis (no
-- dedicated pod spawn) — appropriate because the work is one Kafka publish
-- + one Kafka response, well under a second of chassis time per instance.

INSERT INTO agent_definitions (
    type, display_name, description,
    category, agent_category, status,
    default_config,
    capabilities,
    is_active, version
)
VALUES (
           'thunder-reaper',
           'Thunder Reaper',
           'Periodic sweep that finds Thunder Compute instances older than their '
               || 'max_uptime_hours and dispatches decommission_instance for each. '
               || 'Single-instance-per-tick; relies on scheduled_tasks pre_query '
               || 'to identify the next target.',
           'specialist',
           'executor',
           'experimental',
           jsonb_build_object(
                   'processing_mode', 'task',
                   'workflow', jsonb_build_object(
                           'start_step', 'dispatch_decommission',
                           'processing_mode', 'task',
                           'timeout_seconds', 120,
                           'steps', jsonb_build_object(
                                   'dispatch_decommission', jsonb_build_object(
                                   'action', 'dispatch_thunder_decommission',
                                   'description', 'Publishes decommission_instance to thunder-adapter '
                                       || 'using provisioning_id/thunder_identifier/reason from '
                                       || 'scheduler pre_query merged into input_data.',
                                   'config', jsonb_build_object(
                                           'output_field', 'decommission_dispatch',
                                           'timeout_seconds', 90
                                             ),
                                   'next_step', 'complete'
                                                            ),
                                   'complete', jsonb_build_object(
                                           'action', 'complete_workflow',
                                           'config', jsonb_build_object(
                                                   'output_field', 'reaper_summary'
                                                     )
                                               )
                                    )
                               )
           ),
           '["lifecycle", "thunder", "reaper"]'::jsonb,
           true,
           1
       )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       agent_category = EXCLUDED.agent_category,
                                       status = EXCLUDED.status,
                                       default_config = EXCLUDED.default_config,
                                       capabilities = EXCLUDED.capabilities,
                                       is_active = EXCLUDED.is_active,
                                       updated_at = NOW();


-- ── 2. scheduled_tasks row ──
-- Fires every 15 minutes IF the pre_query returns a row. The pre_query is
-- gating + dynamic input — returns the oldest timed-out instance's
-- provisioning_id, thunder_instance_id, and a human-readable reason string.
-- These columns get merged into the trigger message's input_data.
--
-- concurrency_group 'thunder-lifecycle' with max_concurrent=1 means we never
-- have two reaper ticks running simultaneously. If a previous tick is
-- mid-flight (e.g. decommission API call slow), the next tick is skipped.

INSERT INTO scheduled_tasks (
    name, description,
    interval_seconds, target_agent_type, target_topic,
    concurrency_group, max_concurrent, timeout_seconds,
    enabled, pre_query
)
VALUES (
           'thunder-reaper',
           'Every 15 min, finds the oldest Thunder Compute instance running longer '
               || 'than its max_uptime_hours and triggers thunder-reaper to '
               || 'dispatch a decommission_instance for it. One per tick.',
           900,                                      -- 15 minutes
           'thunder-reaper',
           'system.agent.generic.requests',          -- generic entry point (scheduler default)
           'thunder-lifecycle',                      -- concurrency group
           1,                                        -- one reaper run at a time
           180,                                      -- 3 min in-flight timeout (decommission max ~60s + buffer)
           true,
           $$
               SELECT id::text AS provisioning_id,
           thunder_instance_id,
           'reaper:max_uptime_exceeded after ' ||
           ROUND(EXTRACT(EPOCH FROM (NOW() - running_since))/3600.0, 1) ||
           'h (cap=' || max_uptime_hours || 'h)' AS reason
    FROM thunder_instances
    WHERE status = 'running'
      AND running_since IS NOT NULL
      AND running_since < NOW() - (max_uptime_hours || ' hours')::interval
    ORDER BY running_since ASC
    LIMIT 1
    $$
       )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              target_topic = EXCLUDED.target_topic,
                              concurrency_group = EXCLUDED.concurrency_group,
                              max_concurrent = EXCLUDED.max_concurrent,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              enabled = EXCLUDED.enabled,
                              pre_query = EXCLUDED.pre_query,
                              updated_at = NOW();


COMMIT;


-- ============================================================================
-- Verification queries (run after applying)
-- ============================================================================

-- Confirm agent_definition landed
-- SELECT type, status, is_active, version,
--        default_config->'workflow'->'start_step' AS start_step,
--        default_config->'workflow'->'steps'->'dispatch_decommission'->>'action' AS step_action
-- FROM agent_definitions
-- WHERE type = 'thunder-reaper';
--
-- Expect: thunder-reaper | experimental | t | 1 | "dispatch_decommission" | "dispatch_thunder_decommission"


-- Confirm scheduled_tasks landed
-- SELECT name, interval_seconds, enabled, target_agent_type, target_topic,
--        concurrency_group, max_concurrent, timeout_seconds,
--        last_triggered_at
-- FROM scheduled_tasks
-- WHERE name = 'thunder-reaper';
--
-- Expect: thunder-reaper | 900 | t | thunder-reaper | system.agent.generic.requests
--         | thunder-lifecycle | 1 | 180 | NULL (until first tick)


-- ── Manual smoke test ──
--
-- 1. Insert a synthetic "timed out" row to make sure the pre_query selects it
--    (the row references a fake thunder_instance_id; the reaper will dispatch
--    decommission, the adapter will hit Thunder's API which 404s on this fake,
--    treats 404 as success, marks the row decommissioned. Cost will be 0 if
--    running_since wasn't set sensibly — pick a value in the past).
--
-- INSERT INTO thunder_instances (
--     id, thunder_instance_id, instance_type, instance_ip, ssh_user,
--     ssh_key_secret_name, status, max_uptime_hours, requested_by,
--     hourly_rate_usd, provisioned_at, running_since
-- ) VALUES (
--     gen_random_uuid(),
--     '999999',
--     'synthetic_reaper_test',
--     '10.0.0.42',
--     'ubuntu',
--     'thunder-ssh-fake-uuid-for-test',
--     'running',
--     1,                                   -- 1h max — guaranteed timed out
--     'manual_smoke_test',
--     1.80,
--     NOW() - INTERVAL '2 hours',
--     NOW() - INTERVAL '2 hours'           -- running for 2h, cap is 1h
-- );
--
-- 2. Wait up to 15 min for the scheduler to tick, or kick it manually:
--    UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'thunder-reaper';
--    -- Next tick (within TICK_INTERVAL_SECONDS, default 30s) will fire it.
--
-- 3. Watch the chassis logs for the reaper agent execution and the
--    thunder-adapter logs for the decommission attempt.
--
-- 4. Verify the row was updated:
--    SELECT status, cost_usd, decommissioned_at FROM thunder_instances
--    WHERE thunder_instance_id = '999999';
--    -- Expect: decommissioned | 1.80 (1h overrun × 1.80) | <timestamp>
--
-- 5. Confirm the synthetic Secret would have been targeted (it doesn't exist,
--    so DeleteKeypairSecret returns success on 404 — idempotency working).