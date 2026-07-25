-- Migration: Create awaited_requests table
-- Purpose: Global lookup of awaited requests across orchestrations
-- Uses EXISTING AwaitedRequest struct fields

-- ============================================================================
-- Create awaited_requests table matching AwaitedRequest struct
-- ============================================================================
CREATE TABLE IF NOT EXISTS awaited_requests (
    -- Primary key
                                                request_id VARCHAR(255) PRIMARY KEY,

    -- Orchestration context (from existing struct)
    orchestration_id UUID NOT NULL,
    correlation_id VARCHAR(255) NOT NULL,
    step_id VARCHAR(255),
    step_name VARCHAR(255),

    -- Retry tracking (from existing struct)
    retry_version INTEGER NOT NULL DEFAULT 0,

    -- Target information (from existing struct)
    target_agent_id VARCHAR(255),
    target_agent_type VARCHAR(255),

    -- Topic routing (from existing struct)
    responses_topic VARCHAR(255) NOT NULL,
    requests_topic VARCHAR(255),

    -- Timing (from existing struct)
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    timeout_at TIMESTAMP NOT NULL,

    -- Request chain tracking (from existing struct)
    reply_to_request_id VARCHAR(255),

    -- NEW: Lifecycle tracking (additions to existing struct)
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    processed_at TIMESTAMP,

    CONSTRAINT valid_status CHECK (status IN ('waiting', 'processed', 'expired', 'cancelled'))
    );

-- ============================================================================
-- Indexes for performance
-- ============================================================================

-- Primary lookup by request_id (used when response arrives)
CREATE INDEX IF NOT EXISTS idx_awaited_requests_request_id
    ON awaited_requests(request_id)
    WHERE status = 'waiting';

-- Lookup by orchestration_id (used for cleanup when orchestration completes)
CREATE INDEX IF NOT EXISTS idx_awaited_requests_orchestration
    ON awaited_requests(orchestration_id);

-- Cleanup by status and timeout (used by cleanup job)
CREATE INDEX IF NOT EXISTS idx_awaited_requests_cleanup
    ON awaited_requests(status, timeout_at)
    WHERE status = 'waiting';

-- Lookup by target agent (useful for debugging)
CREATE INDEX IF NOT EXISTS idx_awaited_requests_target
    ON awaited_requests(target_agent_type, target_agent_id);

-- ============================================================================
-- Cleanup function (called periodically)
-- ============================================================================
CREATE OR REPLACE FUNCTION cleanup_expired_awaited_requests()
RETURNS INTEGER AS $$
DECLARE
deleted_count INTEGER;
BEGIN
    -- Mark expired requests
UPDATE awaited_requests
SET status = 'expired',
    processed_at = NOW()
WHERE status = 'waiting'
  AND timeout_at < NOW();

GET DIAGNOSTICS deleted_count = ROW_COUNT;

-- Delete very old processed/expired requests (>7 days)
DELETE FROM awaited_requests
WHERE (status = 'processed' OR status = 'expired')
  AND processed_at < NOW() - INTERVAL '7 days';

RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Comments for documentation
-- ============================================================================
COMMENT ON TABLE awaited_requests IS 
    'Global registry of awaited requests across all orchestrations. Matches AwaitedRequest struct from state.go';

COMMENT ON COLUMN awaited_requests.request_id IS 
    'Unique identifier for the request. Used to match incoming responses.';

COMMENT ON COLUMN awaited_requests.retry_version IS 
    'Track retry attempts for this request';

COMMENT ON COLUMN awaited_requests.target_agent_id IS 
    'Specific agent instance that should handle this request';

COMMENT ON COLUMN awaited_requests.responses_topic IS 
    'Kafka topic where response will arrive';

COMMENT ON COLUMN awaited_requests.requests_topic IS 
    'Kafka topic where request was sent';

COMMENT ON COLUMN awaited_requests.reply_to_request_id IS 
    'For chained requests, tracks the parent request';

COMMENT ON COLUMN awaited_requests.status IS 
    'Current status: waiting (active), processed (response received), expired (timed out), cancelled (orchestration stopped)';

COMMENT ON COLUMN awaited_requests.sent_at IS 
    'When the request was sent/created';

COMMENT ON COLUMN awaited_requests.processed_at IS 
    'When the request was processed/completed';
        -- Add 'processing' status support
ALTER TABLE awaited_requests
DROP CONSTRAINT IF EXISTS awaited_requests_status_check;

ALTER TABLE awaited_requests
    ADD CONSTRAINT awaited_requests_status_check
        CHECK (status IN ('waiting', 'processing', 'processed', 'expired', 'cancelled', 'error'));

-- Add tracking columns
ALTER TABLE awaited_requests
    ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS processing_pod TEXT;

-- Index for faster claims
CREATE INDEX IF NOT EXISTS idx_awaited_requests_status_waiting
    ON awaited_requests(status) WHERE status = 'waiting';

---

-- ═══════════════════════════════════════════════════════════════════════════
-- 2026-07-24 — bugs_open/003 F2: 'retrying' status (mirror of migration 205)
-- ═══════════════════════════════════════════════════════════════════════════
-- Timeout retry used to live only in a process-local sleeping goroutine; a
-- chassis restart deleted every pending timer and nothing consumed
-- status='expired'. The per-minute ticker now atomically claims expired rows
-- into 'retrying' and drives them through the existing resend path. This
-- section keeps the seed in lockstep with the live schema — the authoritative
-- change is docs/agent_docs/sql_for_agents/205_bug003_awaited_retry_and_processed_lease.sql.

ALTER TABLE awaited_requests
DROP CONSTRAINT IF EXISTS awaited_requests_status_check;

ALTER TABLE awaited_requests
    ADD CONSTRAINT awaited_requests_status_check
        CHECK (status IN ('waiting', 'processing', 'retrying', 'processed', 'expired', 'cancelled', 'error'));

CREATE INDEX IF NOT EXISTS idx_awaited_requests_retrying
    ON awaited_requests (processing_started_at) WHERE status = 'retrying';

-- Cleanup function, current form: adds the wedged-'retrying' sweep and deletes
-- terminal 'cancelled'/'error' rows on the same 7-day schedule. REPLACES the
-- definition earlier in this file — seeds run top to bottom, last wins.
CREATE OR REPLACE FUNCTION cleanup_expired_awaited_requests()
RETURNS INTEGER AS $$
DECLARE
deleted_count INTEGER;
BEGIN
UPDATE awaited_requests
SET status = 'expired',
    processed_at = NOW()
WHERE status = 'waiting'
  AND timeout_at < NOW();

GET DIAGNOSTICS deleted_count = ROW_COUNT;

-- bugs_open/003 F2 backstop: cancel 'retrying' rows no actor is driving
-- (the ticker reclaims live-but-orphaned claims at 5 min; 60 min here means
-- the orchestration itself moved on).
UPDATE awaited_requests
SET status = 'cancelled',
    processed_at = NOW()
WHERE status = 'retrying'
  AND processing_started_at < NOW() - INTERVAL '60 minutes';

DELETE FROM awaited_requests
WHERE status IN ('processed','expired','cancelled','error')
  AND processed_at < NOW() - INTERVAL '7 days';

RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
