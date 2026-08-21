-- ROLLBACK for 512. Dropping the column loses the field-vs-prose distinction,
-- after which every row reads as equally certain — which is the thing 512 exists
-- to prevent. Re-running the loader restores it.
BEGIN;
DROP INDEX IF EXISTS idx_positioning_register_attribution;
ALTER TABLE positioning_register DROP COLUMN IF EXISTS attribution;
DELETE FROM schema_migrations WHERE filename = '512_positioning_register_attribution.sql';
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_name='positioning_register' AND column_name='attribution';
  IF n <> 0 THEN RAISE EXCEPTION 'rollback: attribution column still present'; END IF;
  RAISE NOTICE '512 rollback OK';
END $$;
COMMIT;
