-- 331_literal_markdown_claim_timeout_exclusion.sql
--
-- Adds 'literal_markdown' to the claimed-item-timeout sweep's item_type
-- exclusion list (scheduled_tasks.pre_query), so VerifyLiteralMarkdownResolved
-- can gate completion instead of being bypassed.
--
-- WHY THIS FILE EXISTS — the lockstep, not a judgement call
-- ---------------------------------------------------------
-- bugs_open/201 SYMPTOM 2: complete_work_item stamps an item 'complete' on the
-- handler saga's own word. check_literal_markdown.go now registers
-- VerifyLiteralMarkdownResolved; the moment it did,
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion failed and named this
-- obligation. The Go test reads the DECLARED list in
-- 220_claimed_item_timeout_generic_evidence.sql (updated in the same commit);
-- this file is what makes the LIVE column agree. Both halves are required —
-- 305's header records the 151 lane declaring an entry in 220 that never
-- reached the live column, leaving their verifier bypassable for two days.
--
-- Precedent and mechanism: 269, then 305, then 322 — targeted replace against
-- the exact current string, with a before-assertion and an after-assertion.
-- Never retype the pre_query column.
--
-- LIVE STATE READ BEFORE WRITING THIS (2026-08-06, per the LANDMINES entry that
-- requires reading the live column first, because a replace() must name the
-- exact current string and an undetected drift would re-encode someone else's
-- gap into the new one):
--   SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision', 'dead_fragment_link')
-- Exactly one row, seven entries, IDENTICAL to 220's declared list — no drift
-- to carry, same clean position 322 was in.
--
-- URGENCY, stated honestly, and it differs from 322's: items of this type DO
-- already exist (14 fleet-wide, and a fresh one filed on gaswholesalers.com on
-- 2026-08-06), and since bugs_open/201's fix-1 they route to page-build-handler
-- and so CAN be claimed. But the verifier this protects is a Go change and is
-- inert until an image carries it, so until that roll there is no verifier for
-- the sweep to bypass. Applying this before the roll is therefore safe and
-- slightly preferable — it closes the window rather than opening one.
--
-- ROLLBACK: the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 7-entry exclusion list, found % — the live pre_query has drifted; read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link')$old$,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link', 'literal_markdown')$new$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'')%';

DO $verify$
DECLARE v_after int; v_old int;
BEGIN
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%''dead_fragment_link'', ''literal_markdown''%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'exclusion list not extended: % rows carry the new list', v_after;
    END IF;

    -- The replacement must have CONSUMED the old list, not appended a second
    -- one. The closing paren is load-bearing: the old 7-entry string ends
    -- ...'dead_fragment_link')  and the new one continues with a comma, so this
    -- pattern matches ONLY a surviving old list.
    SELECT count(*) INTO v_old FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'')%';
    IF v_old <> 0 THEN
        RAISE EXCEPTION 'the old exclusion list is still present in % row(s)', v_old;
    END IF;

    RAISE NOTICE 'literal_markdown excluded from the claim-timeout auto-complete; VerifyLiteralMarkdownResolved will gate its completion once the image carrying it rolls';
END $verify$;

COMMIT;
