
-- ============================================================================
-- MAINTENANCE QUEUE TABLE (for future use)
-- The page-rebuild agent is designed to eventually pick work from this queue.
-- For now, pages are flagged manually and the agent is triggered via generic-agent.
--
-- Future flow:
--   maintenance-triage (orchestrator) scans for issues:
--     - stale pages (content older than N days)
--     - missing pages (in nav but no content)
--     - broken links
--     - CSS drift
--   → inserts tasks into maintenance_queue
--   → dispatches appropriate specialist (page-rebuild, link-fixer, etc.)
-- ============================================================================

CREATE TABLE IF NOT EXISTS maintenance_queue (
                                                 id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id),

    -- Task classification
    task_type       TEXT NOT NULL,          -- 'page_rebuild', 'css_update', 'nav_fix', 'link_repair', 'content_refresh'
    priority        INT NOT NULL DEFAULT 5, -- 1=urgent, 5=normal, 10=low
    reason          TEXT,                   -- 'stale_content', 'missing_page', 'broken_links', 'manual_request'

-- Task payload — what needs doing
-- For page_rebuild: { "pages": ["use-cases", "privacy"], "rebuild_all": false }
-- For css_update:   { "force_regenerate": true }
-- For nav_fix:      { "remove_stale": true, "add_missing": true }
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Execution tracking
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, claimed, in_progress, complete, failed, cancelled
    claimed_by      TEXT,                              -- agent_id that picked this up
    claimed_at      TIMESTAMPTZ,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,

    -- Result
    result          JSONB,                             -- { "pages_rebuilt": 3, "errors": [] }
    error_message   TEXT,
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,

    -- Metadata
    requested_by    TEXT,                               -- 'system', 'user:uuid', 'triage-agent'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Indexes for queue operations
CREATE INDEX IF NOT EXISTS idx_maintenance_queue_pending
    ON maintenance_queue (priority ASC, created_at ASC)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_maintenance_queue_site
    ON maintenance_queue (site_id, status);

CREATE INDEX IF NOT EXISTS idx_maintenance_queue_type
    ON maintenance_queue (task_type, status);

-- Claim function: atomically claim the next pending task
-- Usage: SELECT * FROM claim_maintenance_task('page_rebuild', 'agent-xyz');
CREATE OR REPLACE FUNCTION claim_maintenance_task(
    p_task_type TEXT,
    p_agent_id TEXT
) RETURNS maintenance_queue AS $$
DECLARE
claimed_task maintenance_queue;
BEGIN
UPDATE maintenance_queue
SET status = 'claimed',
    claimed_by = p_agent_id,
    claimed_at = NOW(),
    updated_at = NOW()
WHERE id = (
    SELECT id FROM maintenance_queue
    WHERE status = 'pending'
      AND task_type = p_task_type
      AND retry_count < max_retries
    ORDER BY priority ASC, created_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
            )
            RETURNING * INTO claimed_task;

RETURN claimed_task;
END;
$$ LANGUAGE plpgsql;

-- Complete function: mark task as done
CREATE OR REPLACE FUNCTION complete_maintenance_task(
    p_task_id UUID,
    p_result JSONB DEFAULT '{}'::jsonb
) RETURNS VOID AS $$
BEGIN
UPDATE maintenance_queue
SET status = 'complete',
    result = p_result,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = p_task_id;
END;
$$ LANGUAGE plpgsql;

-- Fail function: mark task as failed (with retry logic)
CREATE OR REPLACE FUNCTION fail_maintenance_task(
    p_task_id UUID,
    p_error TEXT
) RETURNS VOID AS $$
BEGIN
UPDATE maintenance_queue
SET retry_count = retry_count + 1,
    error_message = p_error,
    status = CASE
                 WHEN retry_count + 1 >= max_retries THEN 'failed'
                 ELSE 'pending'  -- back in queue for retry
        END,
    claimed_by = NULL,
    claimed_at = NULL,
    updated_at = NOW()
WHERE id = p_task_id;
END;
$$ LANGUAGE plpgsql;