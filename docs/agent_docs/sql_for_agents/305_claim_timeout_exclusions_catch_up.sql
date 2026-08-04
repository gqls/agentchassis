-- 305_claim_timeout_exclusions_catch_up.sql
--
-- Adds 'content_duplication' AND 'page_canonical_collision' to the claimed-item
-- -timeout sweep's item_type exclusion list (scheduled_tasks.pre_query), so
-- their Go verifiers can actually gate completion.
--
-- WHY TWO TYPES IN ONE FILE — a declared catch-up, not a sweep-in
-- ---------------------------------------------------------------
-- 'page_canonical_collision' is this change's own type (bugs_open/080,
-- check_page_canonical_collision.go registers the verifier; the lockstep test
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion failed the moment it was
-- registered, naming 220 and this obligation — the designed catch).
--
-- 'content_duplication' belongs to the 151 lane: they added it to 220's
-- DECLARED list (commit ec8ad7959) but no targeted-replace ever reached the
-- LIVE column — verified 2026-08-03:
--   SELECT substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   -- ... 'empty_section', 'orphan_element_refs')   <- 4 entries, theirs absent
-- Since their check went live on completeness-discovery (seed 296, applied
-- 2026-08-03), the sweep could auto-complete their items past the verifier.
-- A replace() must name the exact current string, so applying only my entry
-- would re-encode their gap into the new string. Carrying their declared entry
-- is completing an application they already recorded in 220, not a decision
-- made for them. Flagged to the 151 lane in
-- bugfix_080_canonical_collisions/NOTES_canonical_collisions.md.
--
-- Precedent and mechanism: 269_orphan_element_refs_claim_timeout_exclusion.sql
-- (targeted replace + before/after assertions; never retype the 84-line column).
--
-- NOTE ON URGENCY for page_canonical_collision: its items are born
-- needs_human_review with no handler, so they are never 'claimed' and the sweep
-- cannot reach them today. The exclusion is the lockstep contract, not a live
-- hole — content_duplication's IS a live hole (its items dispatch to
-- deduplicate-sections).
--
-- ROLLBACK: the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 4-entry exclusion list, found % — the live pre_query has drifted; read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs')$old$,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision')$new$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'')%';

DO $verify$
DECLARE v_after int; v_old int;
BEGIN
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%''content_duplication'', ''page_canonical_collision''%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'exclusion list not extended: % rows carry the new list', v_after;
    END IF;

    -- The replacement must have consumed the old list, not appended a second
    -- one. The closing paren is load-bearing: the old 4-entry string ends
    -- ...'orphan_element_refs')  and the new one continues with a comma, so
    -- this pattern matches ONLY a surviving old list.
    SELECT count(*) INTO v_old FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'')%';
    IF v_old <> 0 THEN
        RAISE EXCEPTION 'the old exclusion list is still present in % row(s)', v_old;
    END IF;

    RAISE NOTICE 'content_duplication + page_canonical_collision excluded from the claim-timeout auto-complete; their Go verifiers now gate completion';
END $verify$;

COMMIT;
