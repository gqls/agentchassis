clients_db=# \d processed_messages;
Table "public.processed_messages"
Column      |           Type           | Collation | Nullable | Default
------------------+--------------------------+-----------+----------+---------
 message_id       | uuid                     |           | not null |
 orchestration_id | uuid                     |           |          |
 processed_at     | timestamp with time zone |           | not null | now()
 node_id          | character varying(255)   |           |          |
 correlation_id   | uuid                     |           |          |
 processed_by     | character varying(255)   |           |          |
 request_id       | uuid                     |           |          |
 message_type     | character varying(50)    |           |          |
 agent_id         | character varying(255)   |           |          |
Indexes:
    "idx_processed_correlation" btree (correlation_id)
    "idx_processed_messages_orchestration_id" btree (orchestration_id)
    "idx_processed_request_id" btree (request_id)
    "processed_messages_unique" UNIQUE CONSTRAINT, btree (correlation_id, request_id, agent_id)


-- Add retry_version to processed_messages
ALTER TABLE processed_messages ADD COLUMN IF NOT EXISTS retry_version INT DEFAULT 0;

-- Update unique constraint to include retry_version
ALTER TABLE processed_messages DROP CONSTRAINT IF EXISTS processed_messages_pkey;
ALTER TABLE processed_messages ADD CONSTRAINT processed_messages_pkey
    PRIMARY KEY (correlation_id, request_id, agent_id, retry_version);
-- ═══════════════════════════════════════════════════════════════════════════
-- 2026-07-24 — bugs_open/003 F3: two-phase claim columns (mirror of mig 205)
-- ═══════════════════════════════════════════════════════════════════════════
-- Dedupe used to record RECEIPT, so a redelivered message whose first copy
-- died mid-work was dropped as a "duplicate". Rows are now claimed as
-- 'processing' with a lease and marked 'complete' after the handler returns.
-- DEFAULT 'complete' preserves receipt-time semantics for old binaries.
-- Authoritative change: sql_for_agents/205_bug003_awaited_retry_and_processed_lease.sql.

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

CREATE INDEX IF NOT EXISTS idx_processed_messages_processing
    ON processed_messages (processed_at) WHERE status = 'processing';
