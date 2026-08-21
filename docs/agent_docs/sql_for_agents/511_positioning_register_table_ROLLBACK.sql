-- ROLLBACK for 511 — drop the positioning register table.
--
-- ⚠ THIS DESTROYS DATA THAT MAY NO LONGER EXIST ANYWHERE ELSE. The whole point
-- of 511 is that this table becomes the source of truth; once the markdown file
-- is retired or has drifted, dropping this loses the register. Before running
-- this, dump it:
--
--   \copy (SELECT * FROM positioning_register) TO 'register_backup.csv' CSV HEADER
--
-- Rolling back is safe ONLY while REGISTER_positioning.md is still complete and
-- authoritative — i.e. only very soon after 511 was applied.

BEGIN;

DROP TABLE IF EXISTS positioning_register;

DELETE FROM schema_migrations WHERE filename = '511_positioning_register_table.sql';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM information_schema.tables WHERE table_name='positioning_register';
    IF n <> 0 THEN RAISE EXCEPTION 'rollback: positioning_register still exists'; END IF;
    RAISE NOTICE '511 rollback OK';
END $$;

COMMIT;
