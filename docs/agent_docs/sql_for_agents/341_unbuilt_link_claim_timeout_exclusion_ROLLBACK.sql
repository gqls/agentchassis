-- 341 ROLLBACK — inverse replace: drops 'unbuilt_internal_link' from the
-- claimed-item-timeout exclusion list. Only do this if the verifier is also
-- being withdrawn: an excluded-with-no-verifier type churns on the 40-minute
-- reset forever (bugs_open/006 §C), and a verified-but-not-excluded type is
-- bypassable — the Go guard holds the two in lockstep.

\set ON_ERROR_STOP on

BEGIN;

UPDATE scheduled_tasks
   SET pre_query = replace(
           pre_query,
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'', ''unbuilt_internal_link'')',
           'item_type NOT IN (''truncated_component'', ''hardcoded_section_colors'', ''empty_section'', ''orphan_element_refs'', ''content_duplication'', ''page_canonical_collision'', ''dead_fragment_link'', ''literal_markdown'')'
       )
 WHERE pre_query LIKE '%''unbuilt_internal_link''%';

DO $post$
DECLARE v_rows int;
BEGIN
    SELECT count(*) INTO v_rows FROM scheduled_tasks
     WHERE pre_query LIKE '%''unbuilt_internal_link''%';
    IF v_rows <> 0 THEN
        RAISE EXCEPTION '341 ROLLBACK FAILED: % row(s) still exclude unbuilt_internal_link', v_rows;
    END IF;
    RAISE NOTICE '341 ROLLBACK OK';
END $post$;

COMMIT;
