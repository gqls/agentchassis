-- Migration: Create awaited_requests table
-- Purpose: Global lookup of awaited requests across orchestrations
-- Solves: Race condition where child creates request, parent receives response

-- ============================================================================
-- Create awaited_requests table
-- ============================================================================
CREATE TABLE IF NOT EXISTS awaited_requests (
    -- Primary key
                                                request_id UUID PRIMARY KEY,

    -- Orchestration context
                                                orchestration_id UUID NOT NULL,
                                                correlation_id VARCHAR(255) NOT NULL,

    -- Target information
    target_agent_type VARCHAR(100),
    target_role VARCHAR(100),

    -- Routing information
    reply_to_topic VARCHAR(255) NOT NULL,

    -- Timing
    timeout_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,

    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    -- Status values: 'waiting', 'processed', 'expired', 'cancelled'

    -- Metadata (for debugging)
    step_name VARCHAR(255),
    step_id UUID,

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
    'Global registry of awaited requests across all orchestrations. Used to match responses to waiting orchestrations.';

COMMENT ON COLUMN awaited_requests.request_id IS
    'Unique identifier for the request. Used to match incoming responses.';

COMMENT ON COLUMN awaited_requests.status IS
    'Current status: waiting (active), processed (response received), expired (timed out), cancelled (orchestration stopped)';

COMMENT ON COLUMN awaited_requests.reply_to_topic IS
    'Kafka topic where response should arrive. Used for verification.';