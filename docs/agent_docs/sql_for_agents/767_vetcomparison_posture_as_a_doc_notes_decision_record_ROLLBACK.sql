-- 767_vetcomparison_posture_as_a_doc_notes_decision_record_ROLLBACK.sql
--
-- Removes the doc_notes decision record 767 added. It does NOT touch the register's `posture`
-- key (that is 761's, and 761_..._ROLLBACK.sql removes it).
--
-- ⚠ Removing this re-opens reuse_agent's objection on the 761 round: the posture record would
-- then exist ONLY as a hand-rolled jsonb sub-object inside a shared config blob, rather than in
-- the platform's existing typed decision-record mechanism.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM doc_notes
   WHERE created_by = 'bugfix_414 register-programme lane (migration 767)';
  IF n IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '767 ROLLBACK ABORT: expected exactly 1 record to remove, found % - look before deleting', coalesce(n::text,'NULL');
  END IF;
END $$;

DELETE FROM doc_notes WHERE created_by = 'bugfix_414 register-programme lane (migration 767)';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM doc_notes
   WHERE created_by = 'bugfix_414 register-programme lane (migration 767)';
  IF n IS DISTINCT FROM 0 THEN
    RAISE EXCEPTION '767 ROLLBACK VERIFY: % record(s) remain', coalesce(n::text,'NULL');
  END IF;
  RAISE NOTICE '767 ROLLBACK OK: decision record removed; the register posture key is untouched (see 761 rollback)';
END $$;

COMMIT;
