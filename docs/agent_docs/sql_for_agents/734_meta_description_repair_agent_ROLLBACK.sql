-- 734_..._ROLLBACK.sql — soft-delete the meta-description-repair agent.
--
-- ⚠ AFTER THIS, `meta_description_refused` ITEMS HAVE NO HANDLER. The Go producer
-- (save_page_meta_description_refusal_item.go) keeps filing them; writeWorkItem's
-- registration probe then DEMOTES each to `deferred` — durable and keyed, never a
-- dispatcher livelock (bugs_open/078), but nothing will repair them and they will
-- join the 17%-close flag-only population this agent exists to avoid.
-- If you are rolling this back, file a follow-up for the producer too.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type='meta-description-repair' AND is_active AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: expected 1 live meta-description-repair to retire, found %', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET is_active = false, deleted_at = now(), updated_at = now()
 WHERE type = 'meta-description-repair' AND deleted_at IS NULL;

DO $$
DECLARE live int; open_items int;
BEGIN
    SELECT count(*) INTO live FROM agent_definitions
     WHERE type='meta-description-repair' AND is_active AND deleted_at IS NULL;
    IF live <> 0 THEN
        RAISE EXCEPTION 'ABORT: the agent is still live after the retire';
    END IF;

    SELECT count(*) INTO open_items FROM site_work_items
     WHERE item_type='meta_description_refused'
       AND status NOT IN ('complete','cancelled','rejected','wont_fix');
    RAISE NOTICE '734 ROLLBACK: agent retired. % open meta_description_refused item(s) now have no handler.', open_items;
END $$;

COMMIT;
