-- 566 ROLLBACK — restore `database-cleanup` arm 3's literal status pair.
--
-- ⚠ WHAT YOU ARE RESTORING IS A LEAK. Before 566, arm 3 deleted only
-- `('COMPLETED','FAILED')` while arm 4 skipped everything `is_terminal`, so any OTHER
-- terminal status was reaped by nothing. On 2026-08-22 that was 24 `CANCELLED` rows, the
-- oldest 34 days old against a 24-hour retention norm. Rolling back re-opens that.
--
-- Roll back only if 566 broke the sweep (the whole `pre_query` is one statement — if it
-- errors, NONE of the six arms run, including the agent_error_log and audit cleanups).
-- The symptom would be `scheduled_tasks` reporting a failed run for `database-cleanup`,
-- or all six deletion counts sitting at zero while rows accumulate.
--
-- The guard is on the POST-566 md5, so this refuses if anything landed after 566.

BEGIN;

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '566 ROLLBACK REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;
  IF md5(q) <> '7e9fe52d8f3d822e53fb8afd3628ccd7' THEN
    RAISE EXCEPTION '566 ROLLBACK REFUSED: pre_query is not 566''s output (md5 %). Something landed after 566 — read the live row and undo by hand rather than clobbering another change.', md5(q);
  END IF;
END
$do$;

UPDATE scheduled_tasks
   SET pre_query = replace(pre_query,
         'WHERE status IN (SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)',
         'WHERE status IN (''COMPLETED'', ''FAILED'')'),
       updated_at = now()
 WHERE name = 'database-cleanup'
   AND md5(pre_query) = '7e9fe52d8f3d822e53fb8afd3628ccd7';

UPDATE orchestration_status_vocabulary
   SET notes = replace(notes, ' Reaped at 24h by database-cleanup arm 3 since migration 566; before that arm 3 named only COMPLETED/FAILED literally and arm 4 skips is_terminal rows, so CANCELLED rows were never deleted (24 of them, oldest 34 days, measured 2026-08-22).', ''),
       updated_at = now()
 WHERE status = 'CANCELLED';

DO $do$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF md5(q) <> '7f4321d43784dd26b0a9b3ee27ec412d' THEN
    RAISE EXCEPTION '566 ROLLBACK: pre_query did not return to its pre-566 text (md5 %)', md5(q);
  END IF;
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '566 ROLLBACK: restored pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;
  RAISE NOTICE '566 ROLLBACK: arm 3 literal restored, query executes. The CANCELLED leak is back.';
END
$do$;

COMMIT;
