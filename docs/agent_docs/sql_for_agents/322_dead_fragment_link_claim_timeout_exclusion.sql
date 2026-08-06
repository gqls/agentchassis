-- 322_dead_fragment_link_claim_timeout_exclusion.sql
--
-- Adds 'dead_fragment_link' to the claimed-item-timeout sweep's item_type
-- exclusion list (scheduled_tasks.pre_query), so its Go verifier can gate
-- completion instead of being bypassed.
--
-- WHY THIS FILE EXISTS AT ALL — the lockstep, not a judgement call
-- ---------------------------------------------------------------
-- bugs_open/071's fragment arm (check_phantom_internal_links_fragments.go)
-- registers VerifyDeadFragmentLinkResolved. The moment it did,
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion failed and named this
-- obligation: a verifier that the 15-minute auto-complete branch can walk past
-- is not a gate. The Go test reads the DECLARED list in
-- 220_claimed_item_timeout_generic_evidence.sql; this file is what makes the
-- LIVE column agree. Both halves are required — 305's own header records the
-- 151 lane declaring an entry in 220 that never reached the live column, which
-- left their verifier bypassable for two days.
--
-- Precedent and mechanism: 269_orphan_element_refs_claim_timeout_exclusion.sql,
-- then 305_claim_timeout_exclusions_catch_up.sql — targeted replace against the
-- exact current string, with a before-assertion and an after-assertion. Never
-- retype the 84-line pre_query column.
--
-- LIVE STATE READ BEFORE WRITING THIS (2026-08-06):
--   SELECT name, substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   claimed-item-timeout | item_type NOT IN ('truncated_component',
--     'hardcoded_section_colors', 'empty_section', 'orphan_element_refs',
--     'content_duplication', 'page_canonical_collision')
-- Exactly one row, six entries, and identical to 220's declared list — i.e. no
-- drift to carry this time, unlike 305.
--
-- URGENCY, stated honestly: LOW, and this is the lockstep contract rather than a
-- live hole today. dead_fragment_link items route to page-build-handler /
-- nav-link-fixer and so CAN be claimed — but no item of this type can exist
-- until the image carrying the arm rolls, so applying this before or after that
-- roll is equally safe. Applying it early is a no-op on a type nothing produces.
--
-- ROLLBACK: the inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known 6-entry exclusion list, found % — the live pre_query has drifted; read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision')$old$,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link')$new$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'')%';

DO $verify$
DECLARE v_after int; v_old int;
BEGIN
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%''page_canonical_collision'', ''dead_fragment_link''%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'exclusion list not extended: % rows carry the new list', v_after;
    END IF;

    -- The replacement must have CONSUMED the old list, not appended a second
    -- one. The closing paren is load-bearing: the old 6-entry string ends
    -- ...'page_canonical_collision')  and the new one continues with a comma,
    -- so this pattern matches ONLY a surviving old list.
    SELECT count(*) INTO v_old FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'')%';
    IF v_old <> 0 THEN
        RAISE EXCEPTION 'the old exclusion list is still present in % row(s)', v_old;
    END IF;

    RAISE NOTICE 'dead_fragment_link excluded from the claim-timeout auto-complete; VerifyDeadFragmentLinkResolved now gates its completion';
END $verify$;

COMMIT;
