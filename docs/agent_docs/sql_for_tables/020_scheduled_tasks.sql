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

-- vet cleanup

                                                                                                                                                                                                     -- vet_self_healing.sql
--
-- Replaces vet-task-reset with a broader self-healing task that:
--   1. Fails orchestrations stuck in AWAITING_RESPONSES for >20 min
--   2. Resets collection_tasks stuck in in_progress for >20 min
--   3. Only fires if it actually found something to fix
--
-- This breaks the stall chain:
--   stuck verifier → failed (by this task)
--   → batch processor's continue_on_error skips it (or batch processor
--     itself gets failed on next cycle)
--   → orphaned tasks reset to pending
--   → scheduler fires a new batch processor
--
-- Runs every 3 minutes in its own concurrency group.

-- Remove the old narrow task
DELETE FROM scheduled_tasks WHERE name = 'vet-task-reset';

-- Create the broader self-healing task
INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'stale-orchestration-reaper',
             'Fails stuck AWAITING_RESPONSES orchestrations and resets orphaned in_progress tasks. Self-healing loop that prevents pipeline stalls.',
             180,  -- every 3 minutes
             'generic',
             'system.agent.generic.requests',
             '{}',
             'reaper',
             1,
             60,
             '
         WITH failed_orchs AS (
             UPDATE orchestration_states
             SET status = ''FAILED'',
                 error = ''reaper: stale AWAITING_RESPONSES for >20 min'',
                 updated_at = NOW()
             WHERE status = ''AWAITING_RESPONSES''
               AND last_activity < NOW() - INTERVAL ''20 minutes''
             RETURNING orchestration_id, owner_agent_type, current_step
         ),
         reset_tasks AS (
             UPDATE business_intel.collection_tasks
             SET status = ''pending'',
                 started_at = NULL,
                 orchestration_id = NULL
             WHERE status = ''in_progress''
               AND started_at < NOW() - INTERVAL ''20 minutes''
             RETURNING id
         ),
         expired_awaited AS (
             UPDATE awaited_requests
             SET status = ''expired'',
                 processed_at = NOW()
             WHERE status = ''waiting''
               AND timeout_at < NOW() - INTERVAL ''5 minutes''
             RETURNING request_id
         )
         SELECT
             (SELECT COUNT(*) FROM failed_orchs)::text as orchs_failed,
             (SELECT COUNT(*) FROM reset_tasks)::text as tasks_reset,
             (SELECT COUNT(*) FROM expired_awaited)::text as awaited_expired
         HAVING
             (SELECT COUNT(*) FROM failed_orchs) > 0
             OR (SELECT COUNT(*) FROM reset_tasks) > 0
             OR (SELECT COUNT(*) FROM expired_awaited) > 0
         ',
             true
         );

-- Also fix the concurrency timeout issue on vet-batch-verify.
-- The scheduler considers a task "in flight" until last_completed_at > last_triggered_at
-- or timeout_seconds elapses. But nobody sets last_completed_at for fire-and-forget tasks.
--
-- The reaper now handles cleanup, so we can reduce the timeout windows.
-- This means the scheduler will consider previous runs as timed out sooner,
-- allowing re-firing.
UPDATE scheduled_tasks
SET timeout_seconds = 900  -- 15 min (was 1800)
WHERE name = 'vet-batch-verify';

UPDATE scheduled_tasks
SET timeout_seconds = 1800  -- 30 min (was 3600)
WHERE name = 'vet-sweep-continue';

-- Verify
SELECT name, interval_seconds, timeout_seconds, concurrency_group, enabled,
       CASE WHEN pre_query IS NOT NULL AND pre_query != '' THEN 'yes' ELSE 'no' END as has_pre_query
FROM scheduled_tasks
WHERE name IN ('stale-orchestration-reaper', 'vet-batch-verify', 'vet-sweep-continue', 'vet-task-reset')
ORDER BY name;

--

-- vet permanent pod

-- vet_intel_setup.sql
--
-- Point scheduled tasks at the dedicated vet-intel agent instead of generic.
-- The vet-intel agent listens on system.agent.vet-intel.requests.

-- 1. Update batch verify to target vet-intel
UPDATE scheduled_tasks
SET target_topic = 'system.agent.vet-intel.requests',
    target_agent_type = 'vet-batch-processor'
WHERE name = 'vet-batch-verify';

-- 2. Update sweep to target vet-intel
UPDATE scheduled_tasks
SET target_topic = 'system.agent.vet-intel.requests',
    target_agent_type = 'vet-pipeline-orchestrator'
WHERE name = 'vet-sweep-continue';

-- 3. Verify
SELECT name, target_agent_type, target_topic
FROM scheduled_tasks
WHERE name LIKE 'vet-%'
ORDER BY name;

---

-- database cleanup

-- ============================================================================
-- Database cleanup scheduled task
--
-- Cleans up accumulated data from:
--   1. agent_error_log         — resolved errors > 14 days, unresolved > 30 days
--   2. orchestration_state_audit — keeps last 100k rows
--   3. orchestration_states    — completed/failed > 7 days, stuck > 24 hours
--   4. orchestration_requests  — cascades from orchestration_states deletion
--
-- Runs as a fire_message=false scheduled task (CTE-only, no Kafka).
-- ============================================================================

-- First, fix the orchestration_requests FK to add CASCADE
-- (currently the only FK without it, which blocks orchestration cleanup)
ALTER TABLE orchestration_requests
DROP CONSTRAINT IF EXISTS fk_orch,
    ADD CONSTRAINT fk_orch
        FOREIGN KEY (orchestration_id)
        REFERENCES orchestration_states(orchestration_id)
        ON DELETE CASCADE;

-- ============================================================================
-- The cleanup task
-- ============================================================================
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type,
    target_topic, enabled, fire_message, timeout_seconds,
    concurrency_group, max_concurrent,
    pre_query
) VALUES (
             'database-cleanup',
             'Purges old errors, completed orchestrations, audit trails, and spawned agent records. Keeps recent data for debugging.',
             21600,  -- every 6 hours
             'database-cleanup',
             'system.agent.generic.requests',
             true,
             false,  -- CTE-only, no Kafka message
             120,
             'maintenance',
             1,
             '
             -- 1. Clean agent_error_log (resolved errors > 14 days, unresolved > 30 days)
             WITH deleted_errors AS (
                 DELETE FROM agent_error_log
                 WHERE (resolved = true AND occurred_at < NOW() - INTERVAL ''14 days'')
                    OR (resolved = false AND occurred_at < NOW() - INTERVAL ''30 days'')
                 RETURNING id
             ),

             -- 2. Clean orchestration_state_audit (keep last 100k rows)
             -- The audit tables timestamp column name varies by installation.
    -- Using sequence-based cleanup which is always safe.
    deleted_audit AS (
        DELETE FROM orchestration_state_audit
        WHERE id < (
            SELECT COALESCE(MAX(id) - 100000, 0)
            FROM orchestration_state_audit
        )
        RETURNING id
    ),

             -- 3. Clean completed/failed orchestration_states (> 7 days)
             -- CASCADE will clean orchestration_requests, input_requests, pending_requests
             deleted_orchestrations AS (
        DELETE FROM orchestration_states
        WHERE status IN (''COMPLETED'', ''FAILED'')
          AND updated_at < NOW() - INTERVAL ''7 days''
        RETURNING orchestration_id
    ),

             -- 4. Clean stale orchestrations stuck in EXECUTING_STEP (> 24 hours)
             -- These are leftovers from pod restarts that never completed
             deleted_stale AS (
        DELETE FROM orchestration_states
        WHERE status IN (''EXECUTING_STEP'', ''WAITING_FOR_RESPONSE'')
          AND updated_at < NOW() - INTERVAL ''24 hours''
        RETURNING orchestration_id
    )

    SELECT
        (SELECT COUNT(*) FROM deleted_errors)::text as errors_deleted,
             (SELECT COUNT(*) FROM deleted_audit)::text as audit_deleted,
             (SELECT COUNT(*) FROM deleted_orchestrations)::text as orchestrations_deleted,
             (SELECT COUNT(*) FROM deleted_stale)::text as stale_deleted
                 HAVING
                 (SELECT COUNT(*) FROM deleted_errors) > 0
                 OR (SELECT COUNT(*) FROM deleted_audit) > 0
                 OR (SELECT COUNT(*) FROM deleted_orchestrations) > 0
                 OR (SELECT COUNT(*) FROM deleted_stale) > 0
                 '
             ) ON CONFLICT (name) DO UPDATE SET
                 pre_query = EXCLUDED.pre_query,
                 interval_seconds = EXCLUDED.interval_seconds,
                 description = EXCLUDED.description,
                 updated_at = NOW();

             -- ============================================================================
             -- Retention policy summary
             --
             -- | Table                      | Retention    | Condition                    |
             -- |---------------------------|-------------|------------------------------|
             -- | agent_error_log           | 14d resolved | Resolved errors              |
             -- | agent_error_log           | 30d          | Unresolved errors            |
             -- | orchestration_state_audit | 100k rows    | Keep last 100k by sequence   |
             -- | orchestration_states      | 7d           | COMPLETED or FAILED          |
             -- | orchestration_states      | 24h          | Stuck EXECUTING/WAITING      |
             -- | orchestration_requests    | cascades     | Deleted with parent state    |
             -- | input_requests            | cascades     | Deleted with parent state    |
             -- | pending_requests          | cascades     | Deleted with parent state    |
             -- | site_work_items           | not cleaned  | Kept for analytics/history   |
             --
             -- site_work_items are NOT cleaned automatically. Completed items are
             -- small (no collected_data blob) and useful for tracking what was done.
             -- If this table grows too large, add a separate archival step that
             -- moves completed items to a site_work_items_archive table.
             -- ============================================================================

             -- ============================================================================
             -- Verify the task was created
             -- ============================================================================
             SELECT name, enabled, fire_message, interval_seconds,
                    timeout_seconds, last_triggered_at
             FROM scheduled_tasks
             WHERE name = 'database-cleanup';


---

-- vet-intel setup

-- vet_intel_setup.sql
--
-- Point all vet scheduled tasks at the dedicated vet-intel pod.
-- The vet-intel pod listens on system.agent.vet-intel.requests.

-- 1. Update batch verify
UPDATE scheduled_tasks
SET target_topic = 'system.agent.vet-intel.requests'
WHERE name = 'vet-batch-verify';

-- 2. Update sweep
UPDATE scheduled_tasks
SET target_topic = 'system.agent.vet-intel.requests'
WHERE name = 'vet-sweep-continue';

-- 3. Verify all vet tasks
SELECT name, target_agent_type, target_topic, enabled
FROM scheduled_tasks
WHERE name LIKE 'vet-%'
ORDER BY name;

---

-- companies house
-- =========================================================================
-- 3. SCHEDULED TASK
-- =========================================================================
-- Runs every 20 minutes. Pre-query checks if there are verified businesses
-- without CH data. Processes 20 per run at ~30s each = ~10 min per batch.
-- Targets the vet-intel pod.

INSERT INTO scheduled_tasks (
    id, name, description, interval_seconds,
    target_agent_type, target_topic, input_data,
    concurrency_group, max_concurrent, timeout_seconds,
    pre_query, enabled
) VALUES (
             gen_random_uuid(),
             'ch-enrichment',
             'Enriches verified businesses with Companies House data (financials, officers, ownership). Very gentle rate limiting.',
             1200,  -- every 20 minutes
             'ch-enricher',
             'system.agent.vet-intel.requests',
             '{"batch_size": 20, "vertical_slug": "veterinary"}',
             'ch-enrichment',
             1,      -- only 1 enricher at a time
             900,    -- 15 min timeout window
             '
         SELECT COUNT(*)::text as unenriched
         FROM business_intel.businesses b
         JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id
         LEFT JOIN business_intel.companies_house_data ch ON ch.business_id = b.id
         WHERE bv.slug = ''veterinary''
           AND b.verification_status = ''verified''
           AND ch.business_id IS NULL
         HAVING COUNT(*) > 0
         ',
             false  -- disabled until Go actions are built
         );


-- =========================================================================
-- 4. VERIFY
-- =========================================================================

SELECT name, target_agent_type, target_topic, enabled,
       interval_seconds, timeout_seconds
FROM scheduled_tasks
WHERE name = 'ch-enrichment';

SELECT COUNT(*) as table_exists
FROM information_schema.tables
WHERE table_schema = 'business_intel'
  AND table_name = 'companies_house_data';


--

-- timeouts

UPDATE scheduled_tasks
SET pre_query = '
WITH reset AS (
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN ''failed''
            ELSE ''triaged''
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN ''Claim timed out (attempts exhausted)''
            ELSE ''Claim timed out — handler pod likely died''
        END
    WHERE status = ''claimed''
      AND claimed_at < NOW() - INTERVAL ''10 minutes''
    RETURNING id, item_type, handler_agent, status
)
SELECT COALESCE(COUNT(*)::text, ''0'') as reset_count
FROM reset
'
WHERE name = 'claimed-item-timeout';

-- Also verify the interval is reasonable
SELECT name, interval_seconds, timeout_seconds
FROM scheduled_tasks
WHERE name = 'claimed-item-timeout';

-- same prob
-- Set last_completed_at so the no-refire guard doesn't block
UPDATE scheduled_tasks
SET last_triggered_at = NOW() - INTERVAL '3 minutes',
    last_completed_at = NOW() - INTERVAL '3 minutes'
WHERE name = 'claimed-item-timeout';

-- Also fix the database-cleanup task's HAVING clause while we're at it
UPDATE scheduled_tasks
SET pre_query = REPLACE(
        pre_query,
        '    HAVING
            (SELECT COUNT(*) FROM deleted_errors) > 0
            OR (SELECT COUNT(*) FROM deleted_audit) > 0
            OR (SELECT COUNT(*) FROM deleted_orchestrations) > 0
            OR (SELECT COUNT(*) FROM deleted_stale) > 0',
        '    -- Always returns a row so scheduler marks task as executed'
                )
WHERE name = 'database-cleanup';

--
-- fix - couldn't find domain
-- Fixed: finds the next site to run discovery on
-- Skips sites that already have many open items (> 20)
-- Round-robins by picking the site least recently checked
UPDATE scheduled_tasks
SET pre_query = '
SELECT s.id::text as site_id, s.domain
FROM sites s
WHERE s.status IN (''active'', ''deployed'')
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = ''claimed''
      AND wi.domain = ''build''
  )
  AND (
    SELECT COUNT(*) FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status IN (''triaged'', ''detected'')
      AND wi.domain = ''build''
  ) < 20
ORDER BY s.updated_at ASC NULLS FIRST
LIMIT 1
'
WHERE name = 'improvement-sweep';


---

-- prequery to unstick scheduled tasks so the unsticker can work

INSERT INTO scheduled_tasks (
    name, description, target_topic, target_agent_type,
    interval_seconds, concurrency_group, max_concurrent,
    enabled, fire_message, pre_query
) VALUES (
             'stuck-task-reaper',
             'Auto-reset scheduled tasks stuck for over 1 hour',
             'system.internal.noop',
             'maintenance',
             300,
             'claim-management',
             1,
             true,
             false,
             $PRE$
                 WITH stuck AS (
        UPDATE scheduled_tasks
        SET last_completed_at = NOW()
        WHERE enabled = true
          AND last_triggered_at IS NOT NULL
          AND (
              last_completed_at IS NULL
              OR last_completed_at < last_triggered_at
          )
          AND last_triggered_at < NOW() - INTERVAL '3 hours'
          AND name != 'stuck-task-reaper'
        RETURNING name, last_triggered_at
    )
    SELECT COALESCE(COUNT(*)::text, '0') AS reset_count,
             COALESCE(string_agg(name, ', '), 'none') AS reset_tasks
    FROM stuck
    $PRE$
         );

---
-- site lock

UPDATE scheduled_tasks
SET pre_query = replace(
    pre_query,
    'WHERE s.status IN',
    'WHERE s.locked_at IS NULL AND s.status IN'
)
WHERE name = 'build-pipeline-trigger'
  AND pre_query NOT LIKE '%s.locked_at IS NULL%';

---
-- fixed site lock
UPDATE scheduled_tasks
SET pre_query = '
SELECT COUNT(*)::text as pending_sites
FROM sites s
WHERE s.locked_at IS NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items wi
    WHERE wi.site_id = s.id
      AND wi.status = ''triaged''
      AND wi.domain = ''build''
      AND wi.attempt_count < wi.max_attempts
)
HAVING COUNT(*) > 0
'
WHERE name = 'build-pipeline-trigger';

--

UPDATE scheduled_tasks
SET pre_query = REPLACE(pre_query, 'WAITING_FOR_RESPONSE', 'AWAITING_RESPONSES')
WHERE name = 'database-cleanup';

