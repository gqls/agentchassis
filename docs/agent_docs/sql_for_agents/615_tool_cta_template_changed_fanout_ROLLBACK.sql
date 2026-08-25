-- 615_tool_cta_template_changed_fanout_ROLLBACK.sql
--
-- Cancels the tool-cta template_changed re-render items that 615 filed and that
-- are still OPEN.
--
-- WHAT THIS CANNOT UNDO, and you should not expect it to: an item already
-- `claimed` or `complete` has, or is having, its page re-rendered and
-- published. Cancelling the row would not un-publish the artefact. The way back
-- from a published change is 614's ROLLBACK followed by a FRESH fan-out — the
-- deployed HTML only changes when something re-renders it.
--
-- So run this when you want to STOP the remaining work, not to reverse work
-- already done. It reports both numbers so the distinction is visible rather
-- than assumed.

BEGIN;

DO $$
DECLARE n_open int; n_gone int;
BEGIN
    SELECT count(*) FILTER (WHERE status IN ('detected','triaged')),
           count(*) FILTER (WHERE status NOT IN ('detected','triaged'))
      INTO n_open, n_gone
      FROM site_work_items WHERE created_by = 'bugfix_384_toolcta_fanout';
    RAISE NOTICE '615 rollback: % item(s) still open and cancellable; % already claimed/complete and NOT reversible here', n_open, n_gone;
END $$;

UPDATE site_work_items
   SET status = 'cancelled',
       updated_at = NOW()
 WHERE created_by = 'bugfix_384_toolcta_fanout'
   AND status IN ('detected','triaged');

DO $$
DECLARE n_left int;
BEGIN
    SELECT count(*) INTO n_left FROM site_work_items
     WHERE created_by = 'bugfix_384_toolcta_fanout' AND status IN ('detected','triaged');
    IF n_left > 0 THEN
        RAISE EXCEPTION '615 rollback verify: % item(s) still open after the cancel', n_left;
    END IF;
END $$;

COMMIT;
