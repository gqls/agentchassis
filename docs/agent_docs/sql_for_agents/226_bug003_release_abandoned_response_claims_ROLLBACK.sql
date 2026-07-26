-- ============================================================================
-- ROLLBACK for 226_bug003_release_abandoned_response_claims.sql
--
-- Restores cleanup_expired_awaited_requests() to its 205 definition — i.e.
-- WITHOUT the abandoned-claim release.
--
-- Understand what you are restoring before you run it: this reopens the hole
-- described in 226's header. Rows abandoned in 'processing' become invisible
-- to every recovery path again and their parents hang in AWAITING_RESPONSES
-- with no retry driver, backstopped only by the 90-minute reaper. 181 rows had
-- accumulated that way by 2026-07-26, two of them blocking live parents.
--
-- Rolling back does NOT re-park rows already released — that is not a defect,
-- it is the fix having worked. Nothing here undoes a retry that has run.
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

-- bugs_open/003 F2 backstop: cancel 'retrying' rows no actor is driving.
-- (The ticker reclaims live-but-orphaned claims at 5 min; 60 min here means
-- the orchestration itself moved on. processing_started_at is stamped by
-- every claim, so it cannot be NULL on a 'retrying' row.)
UPDATE awaited_requests
SET status = 'cancelled',
    processed_at = NOW()
WHERE status = 'retrying'
  AND processing_started_at < NOW() - INTERVAL '60 minutes';

-- Delete very old terminal requests (>7 days)
DELETE FROM awaited_requests
WHERE status IN ('processed','expired','cancelled','error')
  AND processed_at < NOW() - INTERVAL '7 days';

RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
