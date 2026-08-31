-- 670 ROLLBACK: restore prompts from the backup table, by id.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_n int;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='bak_670_plan_imagery_wordmark') THEN
    RAISE EXCEPTION '670 rollback: backup table missing';
  END IF;
  UPDATE site_plan_imagery spi SET prompt = b.prompt
    FROM bak_670_plan_imagery_wordmark b WHERE b.id = spi.id;
  GET DIAGNOSTICS v_n = ROW_COUNT;
  RAISE NOTICE '670 rolled back: % prompts restored (backup table kept for audit)', v_n;
END $$;
COMMIT;
