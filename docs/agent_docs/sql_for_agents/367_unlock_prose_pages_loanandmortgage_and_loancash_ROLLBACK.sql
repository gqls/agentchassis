-- 367_..._ROLLBACK.sql — re-lock EXACTLY the pages migration 367 stamped.
--
-- Scoped to `_mig367_unlocked_prose_pages` rather than to the two domains, on
-- purpose: another thread may legitimately flip a page on these sites after 367
-- ran, and a domain-wide re-lock would silently undo their work as well. The
-- stamp table is the record of what THIS migration did.
--
-- The stamp table is left in place; it is the audit trail and re-running 367
-- after this is a clean no-op-then-redo.

BEGIN;

UPDATE pages p
   SET rebuild_policy = m.prev_policy,
       updated_at = now()
  FROM _mig367_unlocked_prose_pages m
 WHERE p.id = m.page_id
   AND COALESCE(p.rebuild_policy, 'generic') = 'generic';

DO $$
DECLARE n_left int;
BEGIN
    SELECT count(*) INTO n_left
      FROM pages p JOIN _mig367_unlocked_prose_pages m ON m.page_id = p.id
     WHERE COALESCE(p.rebuild_policy, 'generic') <> 'owned';
    IF n_left <> 0 THEN
        RAISE EXCEPTION 'mig367 rollback: % stamped pages did not return to owned', n_left;
    END IF;
    RAISE NOTICE 'mig367 rollback OK: all stamped pages are owned again.';
END $$;

COMMIT;
