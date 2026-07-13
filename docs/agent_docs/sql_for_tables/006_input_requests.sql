-- Schema for HITL input_requests table
-- This table persists human input requests for querying and UI display

CREATE TABLE IF NOT EXISTS input_requests (
                                              request_id          UUID PRIMARY KEY,
                                              orchestration_id    UUID NOT NULL,
                                              correlation_id      UUID NOT NULL,
                                              step_id             UUID,
                                              step_name           VARCHAR(255),
    request_type        VARCHAR(50) NOT NULL,  -- 'review', 'confirmation', 'questionnaire'
    agent_type          VARCHAR(100),
    agent_id            UUID,

    -- Request details
    title               TEXT,
    message             TEXT,
    data                JSONB,                 -- The data being reviewed/confirmed
    ui_config           JSONB,                 -- UI hints for display

-- Response routing
    reply_to_topic      VARCHAR(255) NOT NULL,

    -- Timing
    timeout_seconds     INT DEFAULT 3600,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at          TIMESTAMP WITH TIME ZONE,

                                      -- Status tracking
                                      status              VARCHAR(50) DEFAULT 'pending',  -- pending, completed, expired, cancelled

-- Response (when completed)
    response            JSONB,
    responded_by        VARCHAR(255),
    responded_at        TIMESTAMP WITH TIME ZONE,

                                      -- Indexes for common queries
                                      CONSTRAINT fk_orchestration FOREIGN KEY (orchestration_id)
    REFERENCES orchestration_states(orchestration_id) ON DELETE CASCADE
    );

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_input_requests_status ON input_requests(status);
CREATE INDEX IF NOT EXISTS idx_input_requests_correlation ON input_requests(correlation_id);
CREATE INDEX IF NOT EXISTS idx_input_requests_expires ON input_requests(expires_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_input_requests_created ON input_requests(created_at DESC);

-- View for pending requests (for UI)
CREATE OR REPLACE VIEW pending_input_requests AS
SELECT
    request_id,
    orchestration_id,
    correlation_id,
    step_name,
    request_type,
    agent_type,
    title,
    message,
    data,
    ui_config,
    reply_to_topic,
    timeout_seconds,
    created_at,
    expires_at,
    EXTRACT(EPOCH FROM (expires_at - NOW())) as seconds_remaining
FROM input_requests
WHERE status = 'pending'
  AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY created_at ASC;

-- Function to expire old requests (run periodically)
CREATE OR REPLACE FUNCTION expire_input_requests() RETURNS INTEGER AS $$
DECLARE
expired_count INTEGER;
BEGIN
UPDATE input_requests
SET status = 'expired'
WHERE status = 'pending'
  AND expires_at < NOW();

GET DIAGNOSTICS expired_count = ROW_COUNT;
RETURN expired_count;
END;
$$ LANGUAGE plpgsql;

-- Useful queries for debugging/monitoring

-- Get all pending HITL requests
-- SELECT * FROM pending_input_requests;

-- Get requests for a specific correlation
-- SELECT * FROM input_requests WHERE correlation_id = 'd6bc7920-6fad-47d4-a4ac-2777d193432a';

-- Get recent completed requests
-- SELECT * FROM input_requests WHERE status = 'completed' ORDER BY responded_at DESC LIMIT 10;

-- Count by status
-- SELECT status, COUNT(*) FROM input_requests GROUP BY status;