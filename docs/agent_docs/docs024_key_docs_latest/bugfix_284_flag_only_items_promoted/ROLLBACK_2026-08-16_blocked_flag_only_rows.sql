-- ROLLBACK for REPAIR_2026-08-16_blocked_flag_only_rows.sql (bugs_open/284).
--
-- Restores the exact prior status/error of every row the repair touched, from the
-- row-level backup the repair took inside its own transaction. It restores ONLY
-- those two columns: `updated_at` is deliberately left at now(), because pretending
-- a row was never touched is a lie the next reader cannot detect, and because
-- `trg_site_work_items_updated_at` bumps it on any write anyway.
--
-- WHEN YOU WOULD RUN THIS: the repair landed, and then the rows turned out to be
-- wanted in `blocked` after all (e.g. an operator was using the blocked count as a
-- queue), or the guard turned out NOT to be live and the rows are re-blocking with
-- a mixed state you want to reset before retrying cleanly.
--
-- WHEN YOU WOULD NOT: rows re-blocking BECAUSE the guard is not live is not fixed
-- by this — it is fixed by rolling the binary. Rolling back only restores the
-- state the bug produces.

\set ON_ERROR_STOP on

BEGIN;

-- ── 0. Refuse if there is nothing to restore from. A rollback that silently
--       touches 0 rows reads exactly like a successful one.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM repair_284_backup;
    IF n = 0 THEN
        RAISE EXCEPTION 'REFUSING: repair_284_backup is empty — nothing to roll back to. '
                        'If the repair ran in a different session, its TEMP pre-state is gone '
                        'but repair_284_backup is a real table and should be here.';
    END IF;
    RAISE NOTICE 'restoring from % backed-up row(s)', n;
END $$;

\echo '── what will be restored ──'
SELECT b.item_type, w.status AS status_now, b.status AS status_restoring_to, count(*)
FROM repair_284_backup b JOIN site_work_items w USING (id)
GROUP BY 1,2,3 ORDER BY 4 DESC;

-- ── 1. Restore. Guarded on the row still being in the state the repair left it in,
--       so a row a HUMAN has since moved on purpose is not clobbered.
UPDATE site_work_items w
SET status = b.status, error = b.error, updated_at = now()
FROM repair_284_backup b
WHERE w.id = b.id
  AND w.status = CASE b.item_type WHEN 'capability_gap' THEN 'deferred'
                                  WHEN 'image_url_404'  THEN 'detected'
                                  ELSE w.status END;

-- ── 2. Verify, able to fail: every backed-up row is back to its recorded status.
DO $$
DECLARE mismatched int;
BEGIN
    SELECT count(*) INTO mismatched
    FROM repair_284_backup b JOIN site_work_items w USING (id)
    WHERE w.status IS DISTINCT FROM b.status;

    IF mismatched > 0 THEN
        RAISE EXCEPTION 'ROLLBACK VERIFY FAILED: % row(s) not restored — most likely a human '
                        'moved them after the repair, which the guard above deliberately skips. '
                        'Inspect before forcing.', mismatched;
    END IF;
    RAISE NOTICE 'ROLLBACK VERIFY PASSED: all backed-up rows restored';
END $$;

COMMIT;

-- The backup table is left in place on purpose. Drop it only when you are certain
-- neither direction is wanted again: DROP TABLE repair_284_backup;
