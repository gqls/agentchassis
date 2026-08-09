-- 342 ROLLBACK — restores save_sections.page_name_field to the spec path.
-- Only do this alongside rolling back 340: with the authoritative id live and
-- this reverted, an unbuilt_internal_link dispatch writes the TARGET's copy
-- onto the CONTAINER again (the 2026-08-09 dartsonline contamination).

\set ON_ERROR_STOP on

BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,save_sections,config,page_name_field}',
           '"input_data.spec.page_name"'::jsonb
       ),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE v_val text;
BEGIN
    SELECT default_config #>> '{workflow,steps,save_sections,config,page_name_field}' INTO v_val
      FROM agent_definitions
     WHERE type = 'page-build-handler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_val IS DISTINCT FROM 'input_data.spec.page_name' THEN
        RAISE EXCEPTION '342 ROLLBACK FAILED: page_name_field is %', v_val;
    END IF;
    RAISE NOTICE '342 ROLLBACK OK';
END $post$;

COMMIT;
