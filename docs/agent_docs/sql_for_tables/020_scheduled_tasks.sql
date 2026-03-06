-- Migration: 066_kafka_scheduler.sql
-- Creates the scheduled_tasks table and seeds initial schedules.
-- The kafka-scheduler service reads this table and publishes trigger
-- messages to Kafka on the configured intervals.

-- ============================================================================
-- TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS scheduled_tasks (
                                               id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL UNIQUE,
    description       TEXT,

    -- Schedule: interval-based (cron support can be added later)
    interval_seconds  INT NOT NULL DEFAULT 300,

    -- Target: where to send the Kafka trigger message
    target_agent_type TEXT NOT NULL,
    target_topic      TEXT NOT NULL DEFAULT 'system.agent.generic.requests',
    input_data        JSONB DEFAULT '{}',

    -- Concurrency control
    -- Tasks in the same group respect max_concurrent across the group.
    -- NULL group = no concurrency constraints.
    concurrency_group TEXT,
    max_concurrent    INT DEFAULT 1,

    -- Optional: SQL query that returns a single JSON row to merge into input_data.
    -- Evaluated before each trigger. If it returns no rows, the task is skipped
    -- for this tick (nothing to do). Column names become input_data keys.
    -- Example: SELECT site_id::text, domain FROM sites WHERE ... LIMIT 1
    pre_query         TEXT,

    -- Lifecycle
    enabled           BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    last_completed_at TIMESTAMPTZ,
    timeout_seconds   INT DEFAULT 600,

    -- Metadata
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_st_enabled ON scheduled_tasks (enabled)
    WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_st_group ON scheduled_tasks (concurrency_group)
    WHERE concurrency_group IS NOT NULL;

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Build pipeline trigger: finds sites with triaged work items and dispatches
-- Runs every 2 minutes. Concurrency group 'dispatch' prevents overlapping
-- dispatch loops from piling up.
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, input_data, concurrency_group, max_concurrent, timeout_seconds)
VALUES (
           'build-pipeline-trigger',
           'Finds sites with pending triaged/approved work items and triggers the build-dispatch-loop for each. Seeds build_queue entries first.',
           120,
           'build-pipeline-trigger',
           '{}',
           'dispatch',
           2,
           900
       ) ON CONFLICT (name) DO NOTHING;

-- Improvement loop: runs discovery checks and dispatches fixes
-- Runs every 10 minutes per site. The trigger finds the next site needing checks.
-- Same concurrency group as dispatch — they share the pipeline.
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, input_data, concurrency_group, max_concurrent, timeout_seconds, pre_query)
VALUES (
           'improvement-sweep',
           'Triggers improvement-loop for the next site that has not been checked recently. Discovery agents find issues, triage promotes them, dispatch fixes them.',
           600,
           'improvement-loop',
           '{}',
           'dispatch',
           2,
           1800,
           'SELECT s.id::text as site_id, s.domain
            FROM sites s
            WHERE s.status IN (''active'', ''deployed'')
              AND NOT EXISTS (
                SELECT 1 FROM site_work_items wi
                WHERE wi.site_id = s.id
                  AND wi.status = ''claimed''
              )
            ORDER BY s.last_built_at ASC NULLS FIRST
            LIMIT 1'
       ) ON CONFLICT (name) DO NOTHING;


---


-- work items - lifecycle

-- ============================================================================
-- Migration: Work item lifecycle improvements
--
-- 1. Add 'blocked' status to site_work_items
-- 2. Add scheduled task for claimed item timeout
-- 3. Add scheduled task for feasibility re-check (blocked → triaged)
-- ============================================================================

-- 1. Add 'blocked' to allowed statuses
-- The check constraint on status (if one exists) needs updating.
-- Check current constraint:

-- First, see if there's a check constraint on status
SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'site_work_items'::regclass
  AND conname LIKE '%status%';

-- If a constraint exists, drop and recreate with 'blocked' added.
-- If no constraint exists, this is safe to skip.
-- Most implementations use application-level validation, so this may return 0 rows.

-- Add an index for blocked items (for the feasibility re-check query)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_work_items_blocked
    ON site_work_items (handler_agent)
    WHERE status = 'blocked';

-- ============================================================================
-- 2. Claimed item timeout — scheduled task
--
-- Runs every 2 minutes. Resets items stuck in 'claimed' for >10 minutes.
-- Items that hit max_attempts go to 'failed'.
-- Items whose handler_agent doesn't exist in agent_definitions go to 'blocked'.
-- ============================================================================

INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent,
    enabled, timeout_seconds
) VALUES (
             'claimed-item-timeout',
             'Resets work items stuck in claimed status. Items exceeding max_attempts go to failed. Items with non-existent handlers go to blocked.',
             120,
             'generic',
             'system.agent.generic.requests',
             '{}'::jsonb,
             'maintenance',
             1,
             true,
             60
         ) ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    interval_seconds = EXCLUDED.interval_seconds;

-- The pre_query runs the cleanup directly and returns a summary.
-- If no items need cleanup, returns 0 rows → task is skipped.
UPDATE scheduled_tasks
SET pre_query = $PQ$
    WITH timeout_items AS (
    UPDATE site_work_items wi
    SET
        status = CASE
            WHEN wi.attempt_count >= wi.max_attempts THEN 'failed'
            WHEN NOT EXISTS (
                SELECT 1 FROM agent_definitions ad
                WHERE ad.type = wi.handler_agent AND ad.deleted_at IS NULL
            ) THEN 'blocked'
            ELSE 'triaged'
        END,
        claimed_at = NULL,
        claimed_by = NULL,
        error = CASE
            WHEN wi.attempt_count >= wi.max_attempts THEN 'Max attempts reached'
            WHEN NOT EXISTS (
                SELECT 1 FROM agent_definitions ad
                WHERE ad.type = wi.handler_agent AND ad.deleted_at IS NULL
            ) THEN 'Handler agent ''' || wi.handler_agent || ''' not registered'
            ELSE 'Claimed timeout — released for retry'
        END,
        updated_at = NOW()
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '10 minutes'
    RETURNING wi.id, wi.item_type, wi.status as new_status
),
summary AS (
    SELECT
        COUNT(*) FILTER (WHERE new_status = 'triaged') as released,
        COUNT(*) FILTER (WHERE new_status = 'failed') as failed,
        COUNT(*) FILTER (WHERE new_status = 'blocked') as blocked
    FROM timeout_items
)
SELECT released::text as released, failed::text as failed, blocked::text as blocked
FROM summary
WHERE released > 0 OR failed > 0 OR blocked > 0
    $PQ$
WHERE name = 'claimed-item-timeout';

-- NOTE: This pre_query does the work directly in SQL. The scheduled task
-- still triggers the generic agent, but the pre_query has already done the
-- cleanup. The agent just runs and completes with the summary in input_data.
-- If no items needed cleanup, pre_query returns 0 rows and the task is skipped.


-- ============================================================================
-- 3. Feasibility re-check — scheduled task
--
-- Runs every 10 minutes. Checks blocked items whose handler_agent now exists
-- in agent_definitions. Promotes them back to 'triaged'.
-- ============================================================================

INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent,
    enabled, timeout_seconds
) VALUES (
             'feasibility-recheck',
             'Promotes blocked work items back to triaged when their handler agent becomes available.',
             600,
             'generic',
             'system.agent.generic.requests',
             '{}'::jsonb,
             'maintenance',
             1,
             true,
             60
         ) ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    interval_seconds = EXCLUDED.interval_seconds;

UPDATE scheduled_tasks
SET pre_query = $PQ$
    WITH promoted AS (
    UPDATE site_work_items wi
    SET status = 'triaged',
        error = NULL,
        updated_at = NOW()
    WHERE wi.status = 'blocked'
      AND EXISTS (
        SELECT 1 FROM agent_definitions ad
        WHERE ad.type = wi.handler_agent
          AND ad.deleted_at IS NULL
      )
    RETURNING wi.id, wi.item_type, wi.handler_agent
)
SELECT COUNT(*)::text as promoted,
    string_agg(DISTINCT handler_agent, ', ') as agents
FROM promoted
WHERE (SELECT COUNT(*) FROM promoted) > 0
    $PQ$
WHERE name = 'feasibility-recheck';

-- ============================================================================
-- Verify
-- ============================================================================
SELECT name, interval_seconds, enabled, LEFT(pre_query, 80) as pre_query_preview
FROM scheduled_tasks
WHERE name IN ('claimed-item-timeout', 'feasibility-recheck');