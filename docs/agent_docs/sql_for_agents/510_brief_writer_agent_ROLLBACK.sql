-- ROLLBACK for 510 — remove the brief-writer agent.
--
-- What this does NOT undo: any mission_brief spec the agent already wrote, or any
-- needs_brief_review item it raised. Those are data about real domains and a
-- generated brief is not made wrong by removing its generator. Delete them by
-- hand if that is genuinely what you want, and say why.

BEGIN;

UPDATE agent_definitions SET is_active = false, deleted_at = now()
WHERE type = 'brief-writer' AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DELETE FROM schema_migrations WHERE filename = '510_brief_writer_agent.sql';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type='brief-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 0 THEN RAISE EXCEPTION 'rollback: % active brief-writer rows remain', n; END IF;
    RAISE NOTICE '510 rollback OK';
END $$;

COMMIT;
