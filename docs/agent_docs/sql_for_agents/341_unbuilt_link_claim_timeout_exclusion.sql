-- 341_unbuilt_link_claim_timeout_exclusion.sql
--
-- Adds 'unbuilt_internal_link' to the claimed-item-timeout sweep's item_type
-- exclusion list (scheduled_tasks.pre_query), so VerifyUnbuiltInternalLinkResolved
-- (bugs_open/220) can gate completion instead of being bypassed.
--
-- The lockstep, same as 331/322/305/269: the Go guard
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion reads the DECLARED list in
-- 220_claimed_item_timeout_generic_evidence.sql (updated in the same commit as
-- the verifier); this file is what makes the LIVE column agree. Both halves are
-- required — 305's header records the 151 lane declaring an entry that never
-- reached the live column, leaving their verifier bypassable for two days.
--
-- Mechanism: targeted replace against the exact current string, with a
-- before-assertion and an after-assertion. Never retype the pre_query column.
--
-- LIVE STATE READ BEFORE WRITING THIS (2026-08-08 ~19:20Z):
--   SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision', 'dead_fragment_link',
--     'literal_markdown')
-- Exactly one row, eight entries, IDENTICAL to 220's declared list pre-edit —
-- no drift to carry.
--
-- URGENCY, stated honestly (same position as 331): unbuilt_internal_link items
-- exist and are claimable today, but the verifier this protects is a Go change,
-- inert until an image carries it — until that roll there is no verifier for
-- the sweep to bypass. Applying before the roll closes the window rather than
-- opening one.
--
-- ROLLBACK: 341_..._ROLLBACK.sql — the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 8-entry exclusion list, found % — the live pre_query has drifted (or 341 is already applied); read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
           pre_query,
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')',
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'')'
       )
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%''unbuilt_internal_link''%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '341 FAILED: expected exactly 1 scheduled_task excluding unbuilt_internal_link after update, found %', v_rows;
    END IF;
END $post$;

COMMIT;
