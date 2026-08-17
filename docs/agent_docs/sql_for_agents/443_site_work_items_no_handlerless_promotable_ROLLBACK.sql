-- 443_site_work_items_no_handlerless_promotable_ROLLBACK.sql
-- Hand-run sidecar. Drops the constraint. After this, a handler-less row can be
-- promoted again and will be stamped `blocked` with a routing error that
-- misdescribes it (bugs_open/284).
BEGIN;
ALTER TABLE site_work_items DROP CONSTRAINT IF EXISTS swi_no_handlerless_promotable;
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM pg_constraint
     WHERE conname='swi_no_handlerless_promotable' AND conrelid='site_work_items'::regclass;
    IF n <> 0 THEN RAISE EXCEPTION 'ROLLBACK 443: constraint still present'; END IF;
    RAISE NOTICE 'rollback 443 OK';
END $$;
COMMIT;
