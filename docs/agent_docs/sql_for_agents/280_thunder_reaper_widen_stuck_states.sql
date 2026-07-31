-- ============================================================================
-- 280_thunder_reaper_widen_stuck_states.sql
--
-- WHY: the thunder-reaper is enabled and firing every 15 min, but it has NEVER
-- reaped anything — 0 of 23 all-time rows in thunder_instances have reaped_at
-- set (measured 2026-07-31). That is not proof it works; it is the absence of
-- evidence either way, and the reason is its selection query.
--
-- The existing pre_query only matches:
--     status = 'running' AND running_since IS NOT NULL
--
-- so three classes of genuinely-BILLING instance are invisible to it:
--
--   1. status='provisioning' that never advanced — a box exists at Thunder and
--      is charging, but our row never reached 'running'. Provisioning normally
--      completes in ~1 min (measured: 42s await), so >1h is stuck by definition.
--   2. status='decommissioning' whose decommission dispatch failed — the row
--      sits there for ever and NOTHING retries it, because the reaper only
--      looks at 'running'.
--   3. status='running' with a NULL running_since — the age test is
--      `running_since < ...`, which is NULL-false, so the row is skipped
--      silently and bills until someone notices by hand.
--
-- MEASURED, not argued (2026-07-31, three synthetic rows in a rolled-back tx):
--     current pre_query  → 0 of 3 matched
--     query below        → 3 of 3 matched (6h prov, 9h deco, 30h null-clock)
--
-- SAFETY: this only ever WIDENS what gets reaped; no instance the current query
-- catches is excluded by the new one (for status='running' with running_since
-- set, COALESCE returns running_since and the predicate is identical).
-- `thunder_instance_id IS NOT NULL` is deliberate: a provisioning row with no
-- Thunder id means no box was ever created, so there is nothing to bill and
-- nothing to delete — reaping it would be a no-op dispatch.
--
-- STILL NOT COVERED BY THIS FILE — read this before believing you are safe:
-- the reaper can only see instances that exist in OUR table. An instance that
-- Thunder is billing but which we have no row for (a provision whose DB write
-- failed, a box started by hand, a row deleted) is invisible to every check we
-- have. `api.Client.ListInstances` (internal/adapters/thunder/api/client.go:91)
-- can enumerate Thunder's own truth and is unit-tested, but NO orchestration
-- action exposes it, so nothing reconciles the two. That orphan sweep is the
-- follow-up; until it exists, the manual check in the RUNBOOK is the only
-- thing that would catch this class.
--
-- ROLLBACK (the exact pre_query this replaces):
--     UPDATE scheduled_tasks SET pre_query = $q$
--         SELECT id::text AS provisioning_id,
--                thunder_instance_id,
--                'reaper:max_uptime_exceeded after ' ||
--                    ROUND(EXTRACT(EPOCH FROM (NOW() - running_since))/3600.0, 1) ||
--                    'h (cap=' || max_uptime_hours || 'h)' AS reason
--         FROM thunder_instances
--         WHERE status = 'running'
--           AND running_since IS NOT NULL
--           AND running_since < NOW() - (max_uptime_hours || ' hours')::interval
--         ORDER BY running_since ASC
--         LIMIT 1
--     $q$ WHERE name = 'thunder-reaper';
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db < docs/agent_docs/sql_for_agents/280_thunder_reaper_widen_stuck_states.sql
-- ============================================================================

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $q$
    SELECT id::text AS provisioning_id,
           thunder_instance_id,
           'reaper:' || CASE status
                WHEN 'running'         THEN 'max_uptime_exceeded'
                WHEN 'provisioning'    THEN 'stuck_provisioning'
                WHEN 'decommissioning' THEN 'stuck_decommissioning'
           END || ' after ' ||
           ROUND(EXTRACT(EPOCH FROM (NOW() - COALESCE(running_since, provisioned_at, created_at)))/3600.0, 1) ||
           'h (status=' || status || ', cap=' || max_uptime_hours || 'h)' AS reason
    FROM thunder_instances
    WHERE thunder_instance_id IS NOT NULL
      AND (
            (status = 'running'
               AND COALESCE(running_since, provisioned_at, created_at)
                   < NOW() - (max_uptime_hours || ' hours')::interval)
         OR (status = 'provisioning'
               AND COALESCE(provisioned_at, created_at) < NOW() - interval '1 hour')
         OR (status = 'decommissioning'
               AND COALESCE(decommission_requested_at, created_at) < NOW() - interval '1 hour')
          )
    ORDER BY COALESCE(running_since, provisioned_at, created_at) ASC
    LIMIT 1
$q$,
    updated_at = NOW()
WHERE name = 'thunder-reaper';

COMMIT;

-- Verify: all three branches present, task still enabled.
SELECT name,
       enabled,
       interval_seconds,
       (pre_query LIKE '%stuck_provisioning%')    AS covers_provisioning,
       (pre_query LIKE '%stuck_decommissioning%') AS covers_decommissioning,
       (pre_query LIKE '%COALESCE(running_since, provisioned_at, created_at)%') AS covers_null_clock
FROM scheduled_tasks
WHERE name = 'thunder-reaper';
