-- ROLLBACK for 684_content_feed_items_event_extraction_column.sql
--
-- Drops the column. Safe only once the Go actions that read/write it
-- (load_feed_items_for_event_extraction, mark_feed_items_event_extracted) and
-- the feed-triage workflow step referencing them have been rolled back or
-- were never wired in — dropping the column while a live workflow step still
-- references it will fail those runs at the DB, not silently.

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'content_feed_items'
           AND column_name = 'event_extracted_at'
    ) THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at does not exist — nothing to roll back, or already rolled back.';
    END IF;
END $$;

ALTER TABLE content_feed_items
    DROP COLUMN event_extracted_at;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'content_feed_items'
           AND column_name = 'event_extracted_at'
    ) THEN
        RAISE EXCEPTION 'content_feed_items.event_extracted_at still present after DROP COLUMN';
    END IF;
    RAISE NOTICE 'content_feed_items.event_extracted_at dropped';
END $$;

COMMIT;
