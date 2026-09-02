-- 706 ROLLBACK — restore predecessor 6a22f018, remove the successor. Refuses over a newer supersede.
BEGIN;
DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk') AND aspect='portfolio'
     AND is_current AND created_by='copy_quality_two_stage';
  IF n <> 1 THEN RAISE EXCEPTION '706 ROLLBACK: current portfolio row is not 706''s successor — re-census, do not roll back blind'; END IF;
  DELETE FROM site_specs
   WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk') AND aspect='portfolio'
     AND is_current AND created_by='copy_quality_two_stage';
  UPDATE site_specs SET is_current=true, superseded_at=NULL WHERE id='6a22f018-3148-4da9-aca6-0a9ee28ba60d';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '706 ROLLBACK: predecessor restore touched % rows', n; END IF;
END $r$;
DELETE FROM schema_migrations WHERE filename='706_finetuning_portfolio_register_wash.sql';
COMMIT;
