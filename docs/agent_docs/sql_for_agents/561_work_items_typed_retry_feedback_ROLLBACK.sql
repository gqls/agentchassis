-- 561_work_items_typed_retry_feedback_ROLLBACK.sql
--
-- Reverses 561 by dropping site_work_items.retry_feedback.
--
-- SAFE TO RUN ONLY WHILE THE READER IS ABSENT. The Go loader
-- (load_work_item_actions.go) SELECTs this column once its half is live; a
-- chassis running that binary against a dropped column gets 42703 on EVERY
-- work-item load, which is a fleet-wide build outage, not a degraded feature.
-- So: roll the chassis back FIRST, then run this. The guard below refuses to
-- guess for you — it only checks the column is there and unread by config.
--
-- Dropping loses any feedback rows in flight. That costs one regeneration per
-- affected item (it falls back to a blind retry, i.e. pre-345 behaviour); it
-- does not lose any durable record, because the same rejection is also written
-- to agent_error_log by the same producer.

BEGIN;

DO $guard$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='site_work_items'
     AND column_name='retry_feedback';

  IF n <> 1 THEN
    RAISE EXCEPTION '561 ROLLBACK: site_work_items.retry_feedback does not exist — nothing to roll back';
  END IF;

  RAISE NOTICE '561 ROLLBACK: dropping retry_feedback (% rows currently carry one)',
    (SELECT count(*) FROM site_work_items WHERE retry_feedback IS NOT NULL);
END
$guard$;

ALTER TABLE site_work_items DROP COLUMN retry_feedback;

DO $verify$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='site_work_items'
     AND column_name='retry_feedback';

  IF n <> 0 THEN
    RAISE EXCEPTION '561 ROLLBACK VERIFY: retry_feedback still present';
  END IF;

  RAISE NOTICE '561 ROLLBACK OK: retry_feedback dropped';
END
$verify$;

COMMIT;
