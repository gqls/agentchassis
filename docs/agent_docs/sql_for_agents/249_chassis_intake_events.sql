-- ============================================================================
-- 249_chassis_intake_events.sql
--
-- chassis_replica_scaling CS-2 (P1a): the intake side of "Kafka delivers,
-- Postgres decides". The chassis's consume loop becomes a thin ingest —
-- validate, persist one row here, commit the offset, milliseconds per message —
-- and a pool of claim-workers executes events with per-orchestration ordering
-- enforced by chassis_orchestration_claims. Cross-orchestration head-of-line
-- blocking (bugs_open/030's residual, bugs_open/096's mechanism) ends where
-- this table begins: a slow council occupies one worker's claim, not a lane.
--
-- LIVE IMMEDIATELY, USED ONLY WHEN THE FLAG FLIPS. This migration is inert
-- until a chassis image carrying the CHASSIS_INTAKE_MODE=worker_pool branch is
-- rolled AND the env var is set (kubectl set env — never apply -k, the overlays
-- carry other sessions' uncommitted edits). Rollout order mirrors the lane
-- rule (agentbase/agent.go:428-431): schema first, image second, flag last.
--
-- WHY (topic, partition, kafka_offset) IS UNIQUE — the transport idempotency
-- key. The ingest commits the Kafka offset AFTER the insert. A crash between
-- insert and commit redelivers the message; the redelivery hits ON CONFLICT
-- DO NOTHING and re-commits. Exactly-once intake without trusting header
-- hygiene — the DEDUPE_SKIPPED_NO_REQUEST_ID class that narrows
-- processed_messages (bugs_open/096) cannot narrow this.
--
-- WHY serialisation_key AND NOT orchestration_id ALONE. Requests carry the
-- orchestration id in headers; responses sometimes only carry the request id,
-- resolved via awaited_requests at ingest, and a resolvable-by-nothing message
-- gets a deterministic degenerate key (uuid5 of the request id) so it still
-- flows, alone in its own order domain. ClaimAwaitedRequest stays the semantic
-- duplicate arbiter — this key only decides WHO MAY RUN CONCURRENTLY.
--
-- WHY payload IS BYTEA. The worker reconstructs the original kafka.Message
-- (topic, headers, value) and hands it to the EXISTING processMessage
-- unchanged — ingest must not fork validation, bug-034 rejection recording, or
-- error routing. Raw bytes are the only representation that cannot drift from
-- what the wire carried.
--
-- WHY status='done' IS UNCONDITIONAL ON THE IN-PROCESS OUTCOME. It mirrors
-- commitConsumed's reviewed contract (agent.go:744-753, bugs_open/003 F3):
-- handler errors already route through handleProcessingError and the parent
-- drives the retry; re-running locally double-executes. 'failed' is reserved
-- for infrastructure-level inability to run the event at all (attempts cap).
--
-- WHY THE CLAIMS TABLE IS SEPARATE AND KEYED ON serialisation_key. The PK
-- makes double-claim UNREPRESENTABLE regardless of interleaving — the same
-- ON CONFLICT lease-takeover shape as processed_messages
-- (state.go RecordMessageProcessing), which survived bugs_open/003 review. A
-- single-query "claim next event whose key has no running peer" (NOT EXISTS +
-- SKIP LOCKED) was considered and REJECTED: under READ COMMITTED two workers
-- can skip each other's uncommitted locks onto two events of the SAME key.
--
-- Retention: done rows purge after 7 days (owner may re-rule; the purger reads
-- the interval from CHASSIS_INTAKE_RETENTION_DAYS). received_at is indexed for
-- the purge and for the "was my dispatch consumed?" query bugs 030/096 wanted:
--   SELECT received_at, status FROM chassis_intake_events
--   WHERE correlation_id = '<corr>'::uuid;
-- ============================================================================

CREATE TABLE IF NOT EXISTS chassis_intake_events (
    id                BIGSERIAL PRIMARY KEY,
    topic             TEXT        NOT NULL,
    partition         INT         NOT NULL,
    kafka_offset      BIGINT      NOT NULL,
    kind              TEXT        NOT NULL CHECK (kind IN ('request','response')),
    serialisation_key UUID        NOT NULL,
    orchestration_id  UUID,
    correlation_id    UUID,
    request_id        TEXT,
    headers           JSONB       NOT NULL,
    payload           BYTEA       NOT NULL,
    status            TEXT        NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','running','done','failed')),
    attempts          INT         NOT NULL DEFAULT 0,
    received_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at        TIMESTAMPTZ,
    finished_at       TIMESTAMPTZ,
    last_error        TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cie_exactly_once
    ON chassis_intake_events (topic, partition, kafka_offset);

-- The workers' scan: pending work grouped by key, oldest first within a key.
CREATE INDEX IF NOT EXISTS idx_cie_pending
    ON chassis_intake_events (serialisation_key, id)
    WHERE status IN ('pending','running');

-- The purge, and the operator's "did it arrive?" by time window.
CREATE INDEX IF NOT EXISTS idx_cie_received
    ON chassis_intake_events (received_at);

-- The operator's "did MY dispatch arrive?" by correlation.
CREATE INDEX IF NOT EXISTS idx_cie_correlation
    ON chassis_intake_events (correlation_id)
    WHERE correlation_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS chassis_orchestration_claims (
    serialisation_key UUID        PRIMARY KEY,
    claimed_by        TEXT        NOT NULL,   -- "<pod>/worker-<n>"
    claimed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_expires_at  TIMESTAMPTZ NOT NULL
);
