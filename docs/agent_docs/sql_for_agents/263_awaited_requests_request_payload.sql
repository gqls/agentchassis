-- 263 — awaited_requests.request_payload: the exact message that was produced
-- for an awaited request, so a RETRY can be a REPLAY of it.
--
-- bugs_open/129. The retry path (coordinator.go handleRecoverableError) does not
-- resend the original request; it SYNTHESISES a new one from the *awaiting*
-- orchestration's own state, which gives the child the PARENT's orchestration_id
-- (the child then loads the parent's row, sees AWAITING_RESPONSES, and declines
-- the work while logging success), an empty body, and Action:"execute" regardless
-- of the original action.
--
-- Measured 2026-07-28 on this database: of 430 retried awaited_requests in 14
-- days, 430 took that path and 294 (68%) exhausted the retry budget. All-history
-- the distribution is 93 at retry_version 1, 45 at 2 and 294 at 3 — a retry that
-- recovers decays; this one accumulates at the cap.
--
-- ORDERING (stated, not assumed): this column MUST exist before the binary that
-- writes it. It is nullable and additive, so the current binary ignores it
-- completely, and the new binary tolerates NULL — with no payload stored it
-- refuses to emit a retry rather than emitting a poisoned one. There is no
-- window in which either half breaks the other.
--
-- Shape: {"topic":"…","key":"…","headers":{…},"body":{…}} — the arguments of the
-- original producer.Produce call, so the replay needs no reconstruction.

BEGIN;

ALTER TABLE awaited_requests
    ADD COLUMN IF NOT EXISTS request_payload jsonb;

-- Guard: the column exists and is the type the replay path expects. Additive
-- and idempotent, so a re-run is a no-op rather than an error — but a WRONG
-- type would be silent at DDL time and fatal at retry time, so assert it.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'awaited_requests'
          AND column_name = 'request_payload'
          AND data_type = 'jsonb'
    ) THEN
        RAISE EXCEPTION '263: awaited_requests.request_payload is missing or not jsonb — the retry replay path cannot store what it sent';
    END IF;
END $$;

COMMENT ON COLUMN awaited_requests.request_payload IS
    'The exact message produced for this request {topic,key,headers,body}. A retry is a REPLAY of it: only retry_version, message_id and timestamp may differ. NULL means the sending action did not record one — the coordinator then refuses to retry rather than synthesising a message carrying the awaiting orchestration''s own orchestration_id (bugs_open/129).';

COMMIT;
