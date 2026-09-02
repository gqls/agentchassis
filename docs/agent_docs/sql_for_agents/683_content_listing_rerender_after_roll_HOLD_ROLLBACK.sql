-- 683_content_listing_rerender_after_roll_HOLD_ROLLBACK.sql
--
-- Removes the page_rerender items 683 filed, while they are still unstarted.
-- Deliberately refuses to delete an item that has already been picked up: a
-- rerender in flight must be allowed to finish or the page is left half-built.

BEGIN;

DO $$
DECLARE started int;
BEGIN
    SELECT count(*) INTO started
      FROM site_work_items
     WHERE batch_id = '00000000-0000-0000-0000-000000000683'
       AND status NOT IN ('triaged', 'detected');
    IF started > 0 THEN
        RAISE EXCEPTION 'REFUSING: % of 683''s items are no longer triaged — a rerender in '
                        'flight must finish. Cancel those individually if you really mean to.', started;
    END IF;
END $$;

DELETE FROM site_work_items
 WHERE batch_id = '00000000-0000-0000-0000-000000000683'
   AND status IN ('triaged', 'detected');

COMMIT;
