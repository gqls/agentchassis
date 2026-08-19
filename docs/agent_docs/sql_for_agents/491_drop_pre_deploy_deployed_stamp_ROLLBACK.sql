-- 491_drop_pre_deploy_deployed_stamp_ROLLBACK.sql
--
-- Restores the pre-deploy update_status step on both agents and points
-- save_sections back at it.
--
-- SURGICAL BY DESIGN, not a whole-config restore from agent_definitions_backup.
-- 488_page_build_handler_refuses_owned_pages_before_the_writer.sql was PENDING
-- against page-build-handler from another lane when 491 was applied; a blob
-- restore would silently revert it if 488 lands in between. This touches only
-- the two keys 491 changed.
--
-- The step bodies below are the verbatim live config as measured 2026-08-19,
-- immediately before 491 was applied.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-build-handler',      '491_ROLLBACK: pre-restore');
SELECT snapshot_agent('tool-recreation-handler', '491_ROLLBACK: pre-restore');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(default_config, '{workflow,steps,update_status}', '{
             "action": "update_page_status",
             "config": {
                 "status": "deployed",
                 "site_id_field": "site_record.site_id",
                 "page_name_field": "input_data.spec.page_name"
             },
             "next_step": "spawn_rerender_agent",
             "description": "Mark page as deployed",
             "output_field": "status_updated"
         }'::jsonb),
         '{workflow,steps,save_sections,next_step}', '"update_status"'),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(default_config, '{workflow,steps,update_status}', '{
             "action": "update_page_status",
             "config": {
                 "status": "deployed",
                 "site_id_field": "site_record.site_id",
                 "page_name_field": "page_record.name"
             },
             "next_step": "spawn_rerender",
             "description": "Mark page as deployed",
             "output_field": "status_updated"
         }'::jsonb),
         '{workflow,steps,save_sections,next_step}', '"update_status"'),
       updated_at = NOW()
 WHERE type = 'tool-recreation-handler'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE bad int;
BEGIN
    SELECT count(*) INTO bad
      FROM agent_definitions
     WHERE type IN ('page-build-handler','tool-recreation-handler')
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND ( NOT (default_config->'workflow'->'steps' ? 'update_status')
             OR default_config->'workflow'->'steps'->'save_sections'->>'next_step' <> 'update_status' );
    IF bad <> 0 THEN
        RAISE EXCEPTION '491 ROLLBACK verify: % row(s) not restored', bad;
    END IF;
END $$;

COMMIT;
