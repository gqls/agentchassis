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