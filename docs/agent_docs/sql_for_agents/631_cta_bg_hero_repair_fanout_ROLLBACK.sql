-- ROLLBACK for 631: cancel any fanned-out rerender still OPEN. Items already
-- claimed or complete are NOT undone — a completed re-render has published, and
-- the way back from that is 619's rollback plus a fresh fan-out.
BEGIN;
UPDATE site_work_items
   SET status = 'cancelled', updated_at = now()
 WHERE created_by = 'bugfix_398_cta_bg_hero_fanout'
   AND status IN ('detected','triaged');
COMMIT;
