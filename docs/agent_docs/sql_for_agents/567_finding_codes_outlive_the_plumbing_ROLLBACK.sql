-- 567 ROLLBACK — restore the single 14/30-day clock on agent_error_log arm 1.
--
-- ⚠ THIS ROLLBACK IS LOSSY IN ONE DIRECTION AND THE LOSS IS THE POINT OF 567.
-- Restoring the old rule does not merely change future behaviour: the next hourly
-- `database-cleanup` run will delete EVERY finding row older than 30 days that 567 had
-- been keeping — irreversibly, and unread, which is exactly `bugs_open/358`. If you are
-- rolling back for a reason unrelated to arm 1 (a syntax problem, an unexpected sweep
-- failure), EXTRACT FIRST:
--
--   \copy (SELECT * FROM agent_error_log
--            WHERE occurred_at < now() - interval '30 days') TO 'findings_backup.csv' CSV HEADER
--
-- Order-independent with 566, the same way the forward migration is: it accepts the text
-- in either form and asserts arm 3 survives in whichever form it is in.

BEGIN;

DO $do$
DECLARE q text; n int; anchor text;
BEGIN
  anchor := 'WHERE (occurred_at < NOW() - INTERVAL ''30 days''' || chr(10) ||
            '                AND split_part(error_code, '':'', 1) = ANY (ARRAY[';

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '567 ROLLBACK REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;

  n := (length(q) - length(replace(q, anchor, ''))) / length(anchor);
  IF n <> 1 THEN
    RAISE EXCEPTION '567 ROLLBACK REFUSED: 567''s arm 1 is not present exactly once (found %) — it was never applied, or something else has since rewritten it', n;
  END IF;

  RAISE NOTICE '567 ROLLBACK: 567''s arm 1 found — reverting to the 14/30-day clock.';
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = regexp_replace(pre_query,
         'WHERE \(occurred_at < NOW\(\) - INTERVAL ''30 days''.*?OR occurred_at < NOW\(\) - INTERVAL ''365 days''',
         'WHERE (resolved = true AND occurred_at < NOW() - INTERVAL ''14 days'')' || chr(10) ||
         '           OR (resolved = false AND occurred_at < NOW() - INTERVAL ''30 days'')',
         'n'),
       updated_at = now()
 WHERE name = 'database-cleanup';

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';

  IF q LIKE '%split_part(error_code%' OR q LIKE '%INTERVAL ''365 days''%' THEN
    RAISE EXCEPTION '567 ROLLBACK: 567''s arm 1 was not fully removed';
  END IF;
  IF q NOT LIKE '%resolved = true AND occurred_at%' OR q NOT LIKE '%resolved = false AND occurred_at%' THEN
    RAISE EXCEPTION '567 ROLLBACK: the 14/30-day rule was not restored';
  END IF;

  -- negative controls: every other arm survives, arm 3 in whichever form it holds
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orchestrations AS (%' AND q LIKE '%deleted_stale AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '567 ROLLBACK: a pre-existing database-cleanup arm was lost';
  END IF;
  IF NOT (q LIKE '%''COMPLETED'', ''FAILED''%'
       OR q LIKE '%SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)%') THEN
    RAISE EXCEPTION '567 ROLLBACK: arm 3 is in neither its pre-566 nor its post-566 form — it was damaged';
  END IF;

  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '567 ROLLBACK: database-cleanup pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;

  RAISE NOTICE '567 ROLLBACK: applied — the single 14/30-day clock is back, and findings older than 30 days will be deleted on the next sweep.';
END
$do$;

COMMIT;
