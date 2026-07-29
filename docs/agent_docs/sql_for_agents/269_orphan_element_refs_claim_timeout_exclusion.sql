-- 269_orphan_element_refs_claim_timeout_exclusion.sql
--
-- Adds 'orphan_element_refs' to the claimed-item-timeout sweep's item_type
-- exclusion list, so its Go verifier can actually do its job.
--
-- WHY THIS FILE HAS TO EXIST AT ALL
-- --------------------------------
-- Registering a verifier in Go is only half a contract. Migration 220's sweep
-- auto-completes a stuck claim on handler-orchestration evidence alone, and SQL
-- cannot run a Go verifier — so any item_type with a verifier MUST be excluded
-- here, or the sweep completes it 15 minutes later without the verifier ever
-- being consulted. That is the exact defect bugs_open/017 and /021 record.
--
-- The lockstep is pinned by
-- TestRegisteredVerifiersMatchClaimTimeoutExclusion, which is what caught this
-- one: the test failed the moment RegisterVerifier("orphan_element_refs", ...)
-- was added, naming this file and this line. Nothing about it was noticed by a
-- human first.
--
-- WHY A `replace()` AND NOT A REWRITE
-- ----------------------------------
-- 220's own header: "Do NOT hand-edit the column back: it is an 84-line SQL
-- string in a text column and a typo in it breaks the fleet's only claim
-- self-heal, silently, every 120 seconds." Retyping the whole pre_query to add
-- one string is precisely that risk. This changes the 84 lines by a targeted
-- substring replacement and then asserts the result, so a miss is loud rather
-- than silent.
--
-- PRE-CHECK, run 2026-07-29 before writing this: the live column matched 220's
-- file exactly —
--   SELECT substring(pre_query from 'item_type NOT IN \([^)]*\)')
--     FROM scheduled_tasks WHERE pre_query LIKE '%item_type NOT IN%';
--   -- item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section')
-- The verify block below re-asserts that, so if another session has changed the
-- column since, this refuses instead of half-applying.
--
-- ROLLBACK: the inverse replace, same shape —
--   UPDATE scheduled_tasks SET pre_query = replace(pre_query,
--     $n$..., 'empty_section', 'orphan_element_refs')$n$,
--     $o$..., 'empty_section')$o$) WHERE pre_query LIKE '%orphan_element_refs%';

\set ON_ERROR_STOP on

BEGIN;

DO $pre$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'')%';
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 scheduled_task carrying the known exclusion list, found % — the live pre_query has drifted from migration 220; read it before applying this', v_rows;
    END IF;
END $pre$;

UPDATE scheduled_tasks
   SET pre_query = replace(
         pre_query,
         $old$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section')$old$,
         $new$item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs')$new$),
       updated_at = NOW()
 WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'')%';

DO $verify$
DECLARE v_after int; v_old int;
BEGIN
    SELECT count(*) INTO v_after FROM scheduled_tasks
     WHERE pre_query LIKE '%''empty_section'', ''orphan_element_refs''%';
    IF v_after <> 1 THEN
        RAISE EXCEPTION 'exclusion list not extended: % rows carry the new list', v_after;
    END IF;

    -- The replacement must have consumed the old list, not appended a second one.
    SELECT count(*) INTO v_old FROM scheduled_tasks
     WHERE pre_query LIKE '%item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'')%';
    IF v_old <> 0 THEN
        RAISE EXCEPTION 'the old exclusion list is still present in % row(s)', v_old;
    END IF;

    RAISE NOTICE 'orphan_element_refs excluded from the claim-timeout auto-complete; its Go verifier now gates completion';
END $verify$;

COMMIT;
