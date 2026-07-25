-- 205_bug003_awaited_retry_and_processed_lease.sql
--
-- bugs_open/003 F2+F3 — schema half. Two code causes lose spawned-child work:
--
--   F2 (durable retry): timeout retry of an awaited child response is driven
--   ONLY by a process-local sleeping goroutine (coordinator.go
--   handleRequestTimeout); a chassis restart deletes every pending timer and
--   NOTHING consumes awaited_requests.status='expired' — 365 expired rows with
--   retry_version=0 measured 2026-07-23. The Go fix extends the per-minute
--   cleanup ticker to atomically claim newly-expired rows into a new
--   'retrying' status and drive them through the existing resend path. This
--   file adds 'retrying' to the status CHECK and gives the wedged-claim
--   backstop a sweep.
--
--   F3 (at-least-once): the consume loop commits Kafka offsets before
--   processing, and processed_messages records RECEIPT, not COMPLETION — so a
--   pod death both loses the in-flight message AND (post-offset-fix) would
--   drop the redelivery as a "duplicate". The Go fix records a two-phase
--   claim: INSERT status='processing' with a lease on receipt, UPDATE to
--   'complete' after the handler returns. This file adds the status +
--   lease_expires_at columns.
--
--   F4 config (same bug's fix plan): diagnose-agent and image-generator have
--   idle_timeout_seconds=0, so the agentbase idle monitor never starts for
--   exactly the child types observed wedging on Kafka dial failure; only the
--   Job's 24h ActiveDeadlineSeconds bounds them. Set to 600/900.
--
-- ORDERING IS SAFE EITHER WAY. Old binaries: never write 'retrying', ignore
-- the new columns (INSERT names its columns; status defaults to 'complete',
-- which IS today's receipt-time semantics), and the replaced cleanup function
-- only adds a sweep of a status old binaries never set. New binaries REQUIRE
-- this migration: they write status='retrying' (CHECK would reject) and name
-- status/lease_expires_at in the dedupe INSERT (would error). So: apply this
-- BEFORE rolling the F2/F3 image. Applying early is harmless.
--
-- MIRRORS. The cleanup function is also defined in
-- docs/agent_docs/sql_for_tables/001_awaited_requests.sql — that seed is
-- updated in the same commit (re-seed clobber landmine, CLAUDE.md).
--
-- ROLLBACK: 205_bug003_awaited_retry_and_processed_lease_ROLLBACK.sql
-- VERIFY:   205_bug003_awaited_retry_and_processed_lease_VERIFY.sql

-- Defensive: a prior aborted statement in this session would leave a sticky
-- failed-transaction state; clear it before we begin.
ROLLBACK;

BEGIN;

-- ============================================================================
-- A. awaited_requests: admit 'retrying'
-- ============================================================================

ALTER TABLE awaited_requests DROP CONSTRAINT IF EXISTS awaited_requests_status_check;
-- valid_status is the CREATE TABLE-era name (001_awaited_requests.sql:40);
-- live prod has already replaced it, but a fresh seed would not have.
ALTER TABLE awaited_requests DROP CONSTRAINT IF EXISTS valid_status;
ALTER TABLE awaited_requests ADD CONSTRAINT awaited_requests_status_check
    CHECK (status IN ('waiting','processing','retrying','processed','expired','cancelled','error'));

-- The ticker reclaims stale 'retrying' rows (claiming pod died) by
-- processing_started_at age; keep that scan off the main indexes.
CREATE INDEX IF NOT EXISTS idx_awaited_requests_retrying
    ON awaited_requests (processing_started_at) WHERE status = 'retrying';

-- ============================================================================
-- B. cleanup function: sweep wedged 'retrying' rows; delete terminal 'error'
--    and 'cancelled' rows on the same 7-day schedule
-- ============================================================================
-- A 'retrying' row is normally transient (seconds): claimed by one actor,
-- driven through resend ('waiting'), terminal fail ('error'), or release
-- ('processed'). The ticker reclaims a dead claimer at 5 min. A row still
-- 'retrying' after 60 min therefore belongs to an orchestration that left
-- AWAITING_RESPONSES between claim and drive — cancel it so nothing wedges.

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

-- ============================================================================
-- C. processed_messages: two-phase claim columns
-- ============================================================================
-- DEFAULT 'complete' does double duty: it backfills all existing rows without
-- a table rewrite (fast default), and every row an OLD binary inserts during
-- the mixed-fleet roll window lands as 'complete' — exactly the receipt-time
-- semantics that binary implements. Only the new binary writes 'processing'.

ALTER TABLE processed_messages
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'complete';
ALTER TABLE processed_messages
    ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'processed_messages_status_check'
    ) THEN
        ALTER TABLE processed_messages ADD CONSTRAINT processed_messages_status_check
            CHECK (status IN ('processing','complete'));
    END IF;
END $$;

-- The verification probe ("any 'processing' row older than its lease?") and
-- the takeover WHERE clause both scan only leased rows.
CREATE INDEX IF NOT EXISTS idx_processed_messages_processing
    ON processed_messages (processed_at) WHERE status = 'processing';

-- ============================================================================
-- D. F4 config: start the idle monitor for the child types that wedge
-- ============================================================================
-- agent_definitions config is live immediately (no image roll). Values:
-- diagnose-agent 600s (diagnosis steps are minutes, not hours);
-- image-generator 900s (generation + upload headroom). page-content-writer
-- already runs at 180s and is untouched.

UPDATE agent_definitions
SET idle_timeout_seconds = 600
WHERE type = 'diagnose-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND idle_timeout_seconds = 0;

UPDATE agent_definitions
SET idle_timeout_seconds = 900
WHERE type = 'image-generator'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND idle_timeout_seconds = 0;

INSERT INTO schema_migrations (filename)
VALUES ('205_bug003_awaited_retry_and_processed_lease.sql')
ON CONFLICT DO NOTHING;

COMMIT;
