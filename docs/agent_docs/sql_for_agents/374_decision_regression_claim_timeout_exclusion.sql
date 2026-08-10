-- 374_decision_regression_claim_timeout_exclusion.sql
--
-- Adds 'decision_regression' to the claimed-item-timeout sweep's item_type
-- exclusion list (scheduled_tasks.pre_query), so VerifyDecisionRegressionResolved
-- (RFC_015) can gate completion instead of being bypassed.
--
-- The lockstep, same as 341/331/322/305/269: the Go guard
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion reads the DECLARED list in
-- 220_claimed_item_timeout_generic_evidence.sql (updated in the same commit as
-- the verifier); this file is what makes the LIVE column agree. Both halves are
-- required — 305's header records the 151 lane declaring an entry that never
-- reached the live column, leaving their verifier bypassable for two days.
--
-- WHY THIS ITEM TYPE PARTICULARLY NEEDS IT. decision_regression is filed at
-- status='needs_human_review', so the "handler" whose word would otherwise be
-- taken is a PERSON asserting they restored the decided outcome. The verifier
-- re-runs the guard predicate over the stored assembly and can contradict them.
-- On 2026-08-10 this lane completed such an item BY HAND after checking the
-- served page; with the verifier live, that completion would have been checked
-- mechanically instead of trusted.
--
-- LIVE STATE READ BEFORE WRITING THIS (2026-08-10 ~18:40Z):
--   SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision', 'dead_fragment_link',
--     'literal_markdown', 'unbuilt_internal_link', 'revenue_shape_cta',
--     'missing_conversion_path')
-- Exactly one row, ELEVEN entries, IDENTICAL to 220's declared list pre-edit —
-- no drift to carry. Read live rather than copied from 220, because 220 is the
-- declaration and this column is the fact.
--
-- URGENCY, stated honestly (same position as 341/331): decision_regression items
-- exist today (one filed and completed on 2026-08-10), but the verifier this
-- protects is a Go change, inert until an image carries it — until that roll
-- there is no verifier for the sweep to bypass. Applying before the roll closes
-- the window rather than opening one.
--
-- ROLLBACK: 374_..._ROLLBACK.sql — the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 11-entry exclusion list, found % — the live pre_query has drifted (or 374 is already applied); read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
           pre_query,
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'')',
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'', ''decision_regression'')'
       )
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'', ''revenue_shape_cta'', ''missing_conversion_path'')%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%''decision_regression''%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '374 FAILED: expected exactly 1 scheduled_task excluding decision_regression after update, found %', v_rows;
    END IF;
END $post$;

COMMIT;
