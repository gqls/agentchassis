-- 394_d004_covers_fence_narrows_to_the_copy_slot.sql
--
-- OWNER RULING 2026-08-11: narrow D-004's ```covers fence from "every slot on the
-- nine guide pages" to the one slot that actually holds the hand-authored copy.
--
-- WHY. RFC_015's rebuild-door gate went live in chassis v1.0.1288 and immediately
-- did exactly what the fence told it to: on the first guide-page rebuild
-- (2026-08-11 13:45) it preserved ALL THREE sections of guide-building-it and
-- filed three decision_blocked_change items — hero, generic-text-block and
-- call-to-action. An empty `slots` list means every slot, so all 27 sections
-- across the 9 guide pages were frozen against automated change.
--
-- That over-reaches D-004's OWN WORDS, which are: "Structure/styling may improve
-- freely; COPY regeneration requires superseding D-004 by name." The gate cannot
-- tell copy from structure — it preserves the whole row — so the fence has to
-- carry the distinction instead, by naming only the slot the copy lives in.
--
-- MEASURED before writing this (2026-08-11): all nine guide pages carry exactly
-- three slots, identically — hero, generic-text-block, call-to-action:
--   SELECT p.name, string_agg(pc.slot_name, ', ' ORDER BY pc.position)
--     FROM pages p JOIN page_components pc ON pc.page_id = p.id
--    WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
--      AND p.name LIKE 'guide-%' GROUP BY p.name;
-- generic-text-block is the prose body; hero and call-to-action are chrome the
-- pipeline is welcome to improve.
--
-- WHAT THIS DOES NOT DO. It does not weaken D-004: the hand-authored prose is
-- still protected, and changing it still requires naming the decision. It widens
-- what may IMPROVE, which is the owner's standing intent for this whole mechanism
-- ("we want the site to improve, so changes should be allowed, but not regress").
--
-- Targeted replace against the exact current fence, with before- and
-- after-assertions in DO blocks so a drifted body RAISEs rather than silently
-- no-opping. DO/RAISE deliberately, never a verify block of SELECTs: a non-empty
-- result does not stop the COMMIT even under ON_ERROR_STOP.
--
-- ROLLBACK: 394_..._ROLLBACK.sql — the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM doc_notes
     WHERE subject_key = 'D-004-guide-copy-hand-authored'
       AND body LIKE '%"guide-user-acceptance"],"slots":[]}%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 D-004 row carrying the 9-page/empty-slots fence, found % — the body has drifted (or 394 is already applied); read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE doc_notes
   SET body = replace(
           body,
           '"guide-user-acceptance"],"slots":[]}',
           '"guide-user-acceptance"],"slots":["generic-text-block"]}'
       )
 WHERE subject_key = 'D-004-guide-copy-hand-authored'
   AND body LIKE '%"guide-user-acceptance"],"slots":[]}%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM doc_notes
     WHERE subject_key = 'D-004-guide-copy-hand-authored'
       AND body LIKE '%"slots":["generic-text-block"]}%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '394 FAILED: expected exactly 1 D-004 row naming generic-text-block after update, found %', v_rows;
    END IF;
    -- And the empty-slots form must be gone, or a partial replace left both.
    SELECT count(*) INTO v_rows FROM doc_notes
     WHERE subject_key = 'D-004-guide-copy-hand-authored'
       AND body LIKE '%"slots":[]}%';
    IF v_rows <> 0 THEN
        RAISE EXCEPTION '394 FAILED: the empty-slots fence still present in % row(s)', v_rows;
    END IF;
END $post$;

COMMIT;
