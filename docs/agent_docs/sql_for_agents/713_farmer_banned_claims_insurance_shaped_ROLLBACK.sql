-- 713 ROLLBACK: delete the successor, restore the predecessor (the 698 row).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE n int;
BEGIN
  DELETE FROM site_specs WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base'
    AND created_by='loanzy_uk_example_site lane (migration 713)' AND is_current;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '713 rollback: expected 1 successor, deleted %', n; END IF;
  UPDATE site_specs SET is_current=true, superseded_at=NULL
   WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base'
     AND created_by='loanzy_uk_example_site lane (migration 698)';
END $$;
COMMIT;
