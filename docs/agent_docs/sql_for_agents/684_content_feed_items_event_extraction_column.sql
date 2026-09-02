-- 684_content_feed_items_event_extraction_column.sql
--
-- WHAT THIS IS FOR (bugs_open/427, news_feed_ingestion lane, fix candidate #1).
--
-- Adds the idempotency marker for a new extraction step that will run
-- downstream of feed-triage: for each 'relevant' content_feed_items row, ask
-- whether it confirms a specific dated real-world event and, if so, register
-- it as an evidence_base fact (Go side: load_feed_items_for_event_extraction /
-- mark_feed_items_event_extracted actions, this lane's own commit).
--
-- `event_extracted_at IS NULL` means "not yet considered" — set on every item
-- the extraction pass looks at, whether or not it yielded a fact, so the same
-- non-event article isn't re-sent to the LLM every triage cycle forever. This
-- is deliberately a SEPARATE column from the existing `processed_at` (set by
-- apply_feed_scores, meaning "triaged") — conflating the two would make a
-- reader of `processed_at` unable to tell which stage actually ran.
--
-- Config change: additive column, no default, no backfill needed (absence IS
-- the correct value for every existing row — none of them have been through
-- extraction). Live immediately, no image roll required for the column itself;
-- the Go actions that read/write it do need their own image roll before the
-- workflow step referencing them is wired in (CLAUDE.md: image before seed).
-- Reversible — see 684_..._ROLLBACK.sql.

BEGIN;

-- DRIFT GUARD. Abort rather than silently no-op if another session already
-- added this column (or one of the same name for a different purpose).
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'content_feed_items'
           AND column_name = 'event_extracted_at'
    ) THEN
        RAISE EXCEPTION
            'content_feed_items.event_extracted_at already exists — this migration '
            'has already applied, or another session added a column of the same '
            'name for a different purpose. Re-read before re-running.';
    END IF;
END $$;

ALTER TABLE content_feed_items
    ADD COLUMN event_extracted_at timestamptz;

-- VERIFY. DO/RAISE, not a bare SELECT — ON_ERROR_STOP does not fire on a
-- non-empty result set, so a block of SELECTs cannot stop the COMMIT
-- (CLAUDE.md / RFC_006).
DO $$
DECLARE
    col_type text;
    col_nullable text;
BEGIN
    SELECT data_type, is_nullable INTO col_type, col_nullable
      FROM information_schema.columns
     WHERE table_name = 'content_feed_items'
       AND column_name = 'event_extracted_at';

    IF col_type IS NULL THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at was not created';
    END IF;
    IF col_type <> 'timestamp with time zone' THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at has wrong type: %', col_type;
    END IF;
    IF col_nullable <> 'YES' THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at must be nullable (absence = not yet considered), got NOT NULL';
    END IF;

    -- Positive control: every existing row must read NULL — this migration
    -- must not have backfilled a value it has no evidence for.
    IF (SELECT count(*) FROM content_feed_items WHERE event_extracted_at IS NOT NULL) <> 0 THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at was backfilled — it must start NULL on every existing row';
    END IF;

    RAISE NOTICE 'content_feed_items.event_extracted_at added: nullable timestamptz, 0 rows backfilled';
END $$;

COMMIT;
