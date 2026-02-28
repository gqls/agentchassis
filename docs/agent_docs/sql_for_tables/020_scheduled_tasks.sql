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

