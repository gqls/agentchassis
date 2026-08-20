-- ROLLBACK for 505. Drops the not-claimable-before stamp and the default policy
-- row (bugs_open/307).
--
-- ORDER MATTERS: apply 506's rollback FIRST. The read-side predicates reference
-- the column, and a pre_query that names a dropped column fails at every tick —
-- which would stop the build pipeline dispatching entirely, silently, with the
-- watchdog reporting a stall it cannot explain.
--
-- The chassis binary tolerates the column's absence by design (it latches the
-- pre-migration statement shape on SQLSTATE 42703 and logs it), so a rollback in
-- the wrong order is recoverable — but it costs one failure write per item to
-- discover, and the backoff is silently inert afterwards.

BEGIN;

DO $do$
BEGIN
  IF EXISTS (
    SELECT 1 FROM scheduled_tasks
     WHERE pre_query LIKE '%retry_after%'
  ) OR EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE default_config::text LIKE '%retry_after%'
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  ) THEN
    RAISE EXCEPTION 'refusing to drop retry_after while a live pre_query or agent config still reads it — apply 506''s rollback first';
  END IF;
END
$do$;

ALTER TABLE site_work_items DROP COLUMN IF EXISTS retry_after;

DELETE FROM reaper_policies
 WHERE queue = 'site_work_items' AND item_type = '__default__';

DO $do$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
              WHERE table_name = 'site_work_items' AND column_name = 'retry_after') THEN
    RAISE EXCEPTION 'rollback 505: retry_after still present';
  END IF;
END
$do$;

COMMIT;
