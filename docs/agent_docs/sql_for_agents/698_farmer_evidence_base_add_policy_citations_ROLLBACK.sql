-- 698 ROLLBACK: delete the successor, restore the predecessor to current.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE n int;
BEGIN
  DELETE FROM site_specs WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0' AND aspect='evidence_base'
    AND created_by='loanzy_uk_example_site lane (migration 698)' AND is_current;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '698 rollback: expected 1 successor row, deleted %', n; END IF;
  UPDATE site_specs SET is_current=true, superseded_at=NULL
   WHERE id = (SELECT id FROM site_specs WHERE site_id='99cae989-2413-430d-b026-59dfeeb638c0'
               AND aspect='evidence_base' AND NOT is_current ORDER BY superseded_at DESC NULLS LAST LIMIT 1);
END $$;
COMMIT;
