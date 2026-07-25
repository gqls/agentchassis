-- 205_bug003_awaited_retry_and_processed_lease_ROLLBACK.sql
--
-- ⚠ IMAGE FIRST, SCHEMA SECOND. The F2/F3 binary REQUIRES this schema
-- (writes status='retrying'; names status/lease_expires_at in the dedupe
-- INSERT). Running this rollback while that binary is deployed makes every
-- inbound message error at the dedupe INSERT and every ticker retry fail the
-- CHECK. Roll the image back to a pre-F2/F3 tag BEFORE running this file.
--
-- The processed_messages columns are deliberately LEFT IN PLACE: they are
-- additive, the old binary never references them, and dropping a column on a
-- 152k-row table for tidiness is all risk and no benefit. Drop them manually
-- later if ever required.

ROLLBACK;

BEGIN;

-- Refuse loudly if 205 was never applied (nothing to roll back).
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM schema_migrations
        WHERE filename = '205_bug003_awaited_retry_and_processed_lease.sql'
    ) THEN
        RAISE EXCEPTION '205 is not in schema_migrations — nothing to roll back';
    END IF;
END $$;

-- A. Drain 'retrying' rows, then restore the six-value CHECK.
UPDATE awaited_requests
SET status = 'expired'
WHERE status = 'retrying';

ALTER TABLE awaited_requests DROP CONSTRAINT IF EXISTS awaited_requests_status_check;
ALTER TABLE awaited_requests ADD CONSTRAINT awaited_requests_status_check
    CHECK (status IN ('waiting','processing','processed','expired','cancelled','error'));

DROP INDEX IF EXISTS idx_awaited_requests_retrying;

-- B. Restore the pre-205 cleanup function body (from
--    docs/agent_docs/sql_for_tables/001_awaited_requests.sql @ pre-205).
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

DELETE FROM awaited_requests
WHERE (status = 'processed' OR status = 'expired')
  AND processed_at < NOW() - INTERVAL '7 days';

RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- C. processed_messages: columns stay (see header); only the CHECK could
--    reject future writes, and 'complete' remains valid, so it stays too.

-- D. Revert the idle-timeout config only if it still holds our values
--    (do not clobber a later deliberate change).
UPDATE agent_definitions SET idle_timeout_seconds = 0
WHERE type = 'diagnose-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND idle_timeout_seconds = 600;

UPDATE agent_definitions SET idle_timeout_seconds = 0
WHERE type = 'image-generator'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND idle_timeout_seconds = 900;

DELETE FROM schema_migrations
WHERE filename = '205_bug003_awaited_retry_and_processed_lease.sql';

COMMIT;
