-- 716 ROLLBACK: delete the successor, restore the predecessor (the 702 row).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE n int;
BEGIN
  DELETE FROM site_specs WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base'
    AND created_by='loanzy_uk_example_site lane (migration 716)' AND is_current;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '716 rollback: expected 1 successor, deleted %', n; END IF;
  UPDATE site_specs SET is_current=true, superseded_at=NULL
   WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='evidence_base'
     AND created_by='loanzy_uk_example_site lane (migration 702)';
END $$;
COMMIT;
