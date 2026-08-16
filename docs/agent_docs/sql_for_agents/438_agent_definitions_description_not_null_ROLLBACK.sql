-- 438_agent_definitions_description_not_null_ROLLBACK.sql
-- Hand-run sidecar. Drops the NOT NULL and the DEFAULT on agent_definitions.description.
-- Does not re-NULL the three backfilled '' rows (nothing to restore). Be sure: after this,
-- a seed that omits description again produces an unspawnable agent on any binary that
-- predates the bugs_open/287 COALESCE fix.
BEGIN;
ALTER TABLE agent_definitions
    ALTER COLUMN description DROP NOT NULL,
    ALTER COLUMN description DROP DEFAULT;
DO $$
DECLARE nullable text;
BEGIN
    SELECT is_nullable INTO nullable FROM information_schema.columns
     WHERE table_schema='public' AND table_name='agent_definitions' AND column_name='description';
    IF nullable <> 'YES' THEN RAISE EXCEPTION 'ROLLBACK 438: column still NOT NULL'; END IF;
    RAISE NOTICE 'rollback 438 OK: description nullable again';
END $$;
COMMIT;
