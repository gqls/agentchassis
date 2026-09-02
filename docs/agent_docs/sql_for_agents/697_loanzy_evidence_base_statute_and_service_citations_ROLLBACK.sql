-- 697 ROLLBACK: remove loanzy's evidence_base row (697 created it from nothing).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE n int;
BEGIN
  DELETE FROM site_specs WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base'
    AND created_by='loanzy_uk_example_site lane (migration 697)';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '697 rollback: expected to delete exactly 1 row, deleted %', n; END IF;
END $$;
COMMIT;
