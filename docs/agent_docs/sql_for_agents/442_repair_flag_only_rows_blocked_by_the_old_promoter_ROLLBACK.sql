-- 442_repair_flag_only_rows_blocked_by_the_old_promoter_ROLLBACK.sql
-- Hand-run sidecar. Restores exactly the rows 442 stamped — keyed on
-- result.repair_284, never on a re-derived predicate (which would sweep rows
-- this file never touched). Be sure you want this: it puts 60 rows back into a
-- state that misdescribes them as routing failures.
BEGIN;
WITH reverted AS (
    UPDATE site_work_items
       SET status = 'blocked',
           error  = result->'repair_284'->>'from_error',
           result = result - 'repair_284',
           updated_at = now()
     WHERE result ? 'repair_284'
    RETURNING 1
) SELECT count(*) AS reverted FROM reverted;
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM site_work_items WHERE result ? 'repair_284';
    IF n <> 0 THEN RAISE EXCEPTION 'ROLLBACK 442: % row(s) still carry repair_284', n; END IF;
    RAISE NOTICE 'rollback 442 OK';
END $$;
COMMIT;
