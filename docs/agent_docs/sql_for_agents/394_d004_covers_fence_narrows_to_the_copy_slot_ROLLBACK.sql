-- 394_d004_covers_fence_narrows_to_the_copy_slot_ROLLBACK.sql
--
-- Inverse of 394: restores D-004's ```covers fence to every slot on the nine
-- guide pages. Same targeted-replace shape, assertions inverted.
--
-- Rolling this back re-freezes hero and call-to-action on all nine guide pages
-- against automated improvement, and the rebuild-door gate will file up to 27
-- decision_blocked_change items as those pages are rebuilt. That is the state
-- 394 was written to correct, so only roll back if the narrowing turns out to
-- have exposed copy that actually lives outside generic-text-block.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM doc_notes
     WHERE subject_key = 'D-004-guide-copy-hand-authored'
       AND body LIKE '%"slots":["generic-text-block"]}%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 D-004 row naming generic-text-block, found % — 394 may not be applied', v_rows;
    END IF;
END $pre$;

UPDATE doc_notes
   SET body = replace(
           body,
           '"guide-user-acceptance"],"slots":["generic-text-block"]}',
           '"guide-user-acceptance"],"slots":[]}'
       )
 WHERE subject_key = 'D-004-guide-copy-hand-authored'
   AND body LIKE '%"guide-user-acceptance"],"slots":["generic-text-block"]}%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM doc_notes
     WHERE subject_key = 'D-004-guide-copy-hand-authored'
       AND body LIKE '%"slots":[]}%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '394 ROLLBACK FAILED: expected the empty-slots fence restored on exactly 1 row, found %', v_rows;
    END IF;
END $post$;

COMMIT;
