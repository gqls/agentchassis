-- Migration: HTTP request logging for centralised outbound HTTP visibility
--
-- Follows the same pattern as llm_call_log: captures request/response metadata
-- for all outbound HTTP calls from Go actions. Serves two purposes:
--   1. Operational visibility (which APIs are we hitting, how fast, what errors)
--   2. Rate limit monitoring (track calls per domain per window)
--
-- Run against clients_db as clients_user.

BEGIN;

-- ============================================================
-- 1. Table
-- ============================================================
CREATE TABLE IF NOT EXISTS http_request_log (
                                                id              BIGSERIAL PRIMARY KEY,
                                                created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Who made the call
    agent_type      TEXT,
    agent_id        TEXT,
    step_name       TEXT,
    orchestration_id TEXT,
    correlation_id  TEXT,
    action_name     TEXT,         -- e.g. 'ch_fetch_accounts', 'ch_detail_fetch'

-- What was called
    method          TEXT NOT NULL DEFAULT 'GET',
    url             TEXT NOT NULL,
    domain          TEXT,         -- extracted from URL for per-domain grouping
    path            TEXT,         -- URL path component

-- Response
    status_code     INTEGER,
    response_bytes  INTEGER,
    content_type    TEXT,
    latency_ms      INTEGER,
    success         BOOLEAN NOT NULL DEFAULT true,
    error_message   TEXT,

    -- Context
    metadata        JSONB DEFAULT '{}'::jsonb  -- any extra context (company_number, etc.)
    );

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_http_log_time
    ON http_request_log (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_http_log_domain
    ON http_request_log (domain, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_http_log_agent
    ON http_request_log (agent_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_http_log_action
    ON http_request_log (action_name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_http_log_errors
    ON http_request_log (created_at DESC)
    WHERE success = false;

-- ============================================================
-- 2. Stats view
-- ============================================================
CREATE OR REPLACE VIEW http_request_stats AS
SELECT
    domain,
    action_name,
    COUNT(*) AS total_calls,
    COUNT(*) FILTER (WHERE success = true) AS successes,
    COUNT(*) FILTER (WHERE success = false) AS failures,
    ROUND(AVG(latency_ms)) AS avg_latency_ms,
    ROUND(AVG(response_bytes)) AS avg_response_bytes,
    MAX(created_at) AS last_call,
    -- Calls in last 5 min (for rate limit monitoring)
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '5 minutes') AS calls_last_5min
FROM http_request_log
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY domain, action_name
ORDER BY total_calls DESC;

-- ============================================================
-- 3. Cleanup function (90 days success, 180 days errors)
-- ============================================================
CREATE OR REPLACE FUNCTION cleanup_old_http_logs()
RETURNS TABLE(deleted_success BIGINT, deleted_errors BIGINT) AS $$
DECLARE
d_success BIGINT;
    d_errors BIGINT;
BEGIN
DELETE FROM http_request_log
WHERE success = true AND created_at < NOW() - INTERVAL '90 days';
GET DIAGNOSTICS d_success = ROW_COUNT;

DELETE FROM http_request_log
WHERE success = false AND created_at < NOW() - INTERVAL '180 days';
GET DIAGNOSTICS d_errors = ROW_COUNT;

RETURN QUERY SELECT d_success, d_errors;
END;
$$ LANGUAGE plpgsql;

COMMIT;