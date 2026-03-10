to disable later
UPDATE scheduled_tasks SET enabled = false WHERE name LIKE 'vet-%';

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

---

-- added system user
-- Check if the function exists
SELECT proname FROM pg_proc WHERE proname = 'create_client_schema';

-- If it exists, use it
SELECT create_client_schema('system');

-- Insert the client record
INSERT INTO clients (external_id, name, settings)
VALUES ('system', 'System Scheduler', '{"type": "internal", "description": "Used by kafka-scheduler for automated triggers"}'::jsonb)
    ON CONFLICT DO NOTHING;

-- Verify both
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'client_system';
SELECT * FROM clients WHERE external_id = 'system';

--


--The scheduler already skips fireTrigger when the pre_query returns no rows (dynamicData == nil → continue). The problem is: when the pre_query DOES find items to reset/promote, it returns a row (the count), and the scheduler fires a message to generic, which runs the calculator.
--The pre_query already does all the work in the CTE's UPDATE. We just need to prevent it from returning rows to the scheduler:

-- claimed-item-timeout: do the work, never trigger a message
UPDATE scheduled_tasks
SET pre_query = '
WITH reset AS (
    UPDATE site_work_items
    SET status = ''triaged'',
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1
    WHERE status = ''claimed''
      AND claimed_at < NOW() - INTERVAL ''10 minutes''
      AND attempt_count < max_attempts
    RETURNING id, item_type, handler_agent
)
SELECT COUNT(*)::text as reset_count,
       string_agg(DISTINCT handler_agent, '', '') as agents
FROM reset
WHERE 1 = 0
'
WHERE name = 'claimed-item-timeout';

-- feasibility-recheck: same pattern
UPDATE scheduled_tasks
SET pre_query = '
WITH promoted AS (
    UPDATE site_work_items wi
    SET status = ''triaged'',
        error = NULL
    WHERE wi.status = ''blocked''
      AND EXISTS (
        SELECT 1 FROM agent_definitions ad
        WHERE ad.type = wi.handler_agent
          AND ad.deleted_at IS NULL
      )
    RETURNING wi.id, wi.item_type, wi.handler_agent
)
SELECT COUNT(*)::text as promoted,
       string_agg(DISTINCT handler_agent, '', '') as agents
FROM promoted
WHERE 1 = 0
'
WHERE name = 'feasibility-recheck';


--

-- Fix claimed-item-timeout: use HAVING to suppress zero-result rows
UPDATE scheduled_tasks
SET pre_query = '
WITH reset AS (
    UPDATE site_work_items
    SET status = ''triaged'',
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1
    WHERE status = ''claimed''
      AND claimed_at < NOW() - INTERVAL ''10 minutes''
      AND attempt_count < max_attempts
    RETURNING id, item_type, handler_agent
)
SELECT COUNT(*)::text as reset_count,
       string_agg(DISTINCT handler_agent, '', '') as agents
FROM reset
HAVING COUNT(*) > 0
'
WHERE name = 'claimed-item-timeout';

-- Fix feasibility-recheck: same HAVING pattern
UPDATE scheduled_tasks
SET pre_query = '
WITH promoted AS (
    UPDATE site_work_items wi
    SET status = ''triaged'',
        error = NULL
    WHERE wi.status = ''blocked''
      AND EXISTS (
        SELECT 1 FROM agent_definitions ad
        WHERE ad.type = wi.handler_agent
          AND ad.deleted_at IS NULL
      )
    RETURNING wi.id, wi.item_type, wi.handler_agent
)
SELECT COUNT(*)::text as promoted,
       string_agg(DISTINCT handler_agent, '', '') as agents
FROM promoted
HAVING COUNT(*) > 0
'
WHERE name = 'feasibility-recheck';

--


-- Add a pre_query to improvement-sweep that only fires when queue is small
UPDATE scheduled_tasks
SET pre_query = '
SELECT COUNT(*)::text as pending_items
FROM site_work_items
WHERE status IN (''triaged'', ''claimed'', ''detected'')
  AND domain = ''build''
HAVING COUNT(*) < 20
'
WHERE name = 'improvement-sweep';

--

-- added vet to scheduled tasks
-- vet_scheduled_tasks.sql
--
-- Three scheduled tasks for automated vet pipeline operation.
-- Disable or delete after the initial data collection campaign.
--
-- 1. vet-batch-verify:  Triggers batch processor every 15 min if pending tasks exist
-- 2. vet-task-reset:    Resets stuck in_progress tasks every 5 min (prevents orphans)
-- 3. vet-sweep-continue: Triggers sweep for unswept areas every 30 min
--

-- ═══════════════════════════════════════════════════════════════════
-- 1. VET BATCH VERIFY — claims and verifies pending collection tasks
-- ═══════════════════════════════════════════════════════════════════
-- Pre-query: only fires if there are pending tasks and no batch processor
-- currently running (no in_progress tasks = nothing active)
INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'vet-batch-verify',
             'Triggers vet-batch-processor to verify pending businesses. Runs every 15 minutes if pending tasks exist and no batch is currently active.',
             900,  -- 15 minutes
             'vet-batch-processor',
             'system.agent.generic.requests',
             '{"batch_size": 100, "task_type": "initial_verification", "vertical_slug": "veterinary"}',
             'vet-verify',
             1,     -- only 1 batch processor at a time
             1800,  -- 30 min timeout (in-flight window)
             '
         SELECT COUNT(*)::text as pending_tasks
         FROM business_intel.collection_tasks
         WHERE status = ''pending''
           AND vertical_id = (SELECT id FROM business_intel.business_verticals WHERE slug = ''veterinary'')
         HAVING COUNT(*) > 0
         ',
             true
         );

-- ═══════════════════════════════════════════════════════════════════
-- 2. VET TASK RESET — resets orphaned in_progress tasks
-- ═══════════════════════════════════════════════════════════════════
-- Pure maintenance SQL via pre_query. The UPDATE runs in the pre_query.
-- If nothing was reset, returns no rows → task skipped (no message sent).
-- Uses the generic agent as a no-op target.
INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'vet-task-reset',
             'Resets collection_tasks stuck in in_progress for over 30 minutes. Prevents orphaned tasks from blocking the queue.',
             300,  -- every 5 minutes
             'generic',
             'system.agent.generic.requests',
             '{}',
             'maintenance',
             1,
             60,
             '
         WITH reset AS (
             UPDATE business_intel.collection_tasks
             SET status = ''pending'',
                 started_at = NULL,
                 orchestration_id = NULL
             WHERE status = ''in_progress''
               AND started_at < NOW() - INTERVAL ''30 minutes''
             RETURNING id
         )
         SELECT COUNT(*)::text as reset_count
         FROM reset
         HAVING COUNT(*) > 0
         ',
             true
         );

-- ═══════════════════════════════════════════════════════════════════
-- 3. VET SWEEP CONTINUE — sweeps unswept areas in batches
-- ═══════════════════════════════════════════════════════════════════
-- Pre-query: only fires if there are unswept areas remaining.
-- Sends limit=200 so each run does a manageable batch.
INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'vet-sweep-continue',
             'Triggers area-sweep-orchestrator to sweep the next batch of unswept areas. Runs every 30 minutes if unswept areas remain.',
             1800,  -- 30 minutes
             'vet-pipeline-orchestrator',
             'system.agent.generic.requests',
             '{"limit": 200, "promote_limit": 500, "verify_limit": 0, "delay_ms": 5000, "country": "GB", "business_type": "veterinary practice", "vertical_slug": "veterinary"}',
             'vet-sweep',
             1,     -- only 1 sweep at a time
             3600,  -- 1 hour timeout window
             '
         SELECT COUNT(*)::text as unswept_areas
         FROM business_intel.search_areas
         WHERE last_swept_at IS NULL
         HAVING COUNT(*) > 0
         ',
             true
         );

-- ═══════════════════════════════════════════════════════════════════
-- Verification: check all three were created
-- ═══════════════════════════════════════════════════════════════════
SELECT name, interval_seconds, target_agent_type, enabled,
       CASE WHEN pre_query IS NOT NULL THEN 'yes' ELSE 'no' END as has_pre_query
FROM scheduled_tasks
WHERE name LIKE 'vet-%'
ORDER BY name;

-- to disable later:
UPDATE scheduled_tasks SET enabled = false WHERE name LIKE 'vet-%';



---


-- adding a pre query that checks that there are items to be dispatched
UPDATE scheduled_tasks
SET pre_query = '
SELECT s.id::text as site_id, s.domain
FROM sites s
WHERE EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = ''triaged''
      AND wi.domain = ''build''
)
ORDER BY s.domain
LIMIT 1
HAVING COUNT(*) > 0
'
WHERE name = 'build-pipeline-trigger';


---
-- stop the scheduler sending messages to generic agent when it's finished - add a 'whether to send message' column

these three tasks (claimed-item-timeout, feasibility-recheck, vet-task-reset) are cron SQL jobs, not agent triggers. Their pre_query CTEs do all the work. The message sent afterward is meaningless — it exists only because the scheduler assumes every task triggers an agent.
The proper fix is a fire_message column on scheduled_tasks. When false, the scheduler runs the pre_query and stops. No Kafka message, no agent spawned.

