-- ROLLBACK for 513 — removes the rendered_html_transform enablement from
-- section-editor's apply_edit step: deletes the allow_rendered_html_transform
-- flag and strips "transform_name" back out of input_fields.
--
-- After this, a literal_markdown item already minted with the section-editor
-- route is refused by the action's config gate (loud error → attempt ladder →
-- human review); it is NOT silently re-routed. The detector keeps minting
-- section-editor-routed items until ITS change is rolled back too (the code is
-- in the chassis image; this file only closes the config half).

BEGIN;

SELECT snapshot_agent('section-editor',
                      '513_..._ROLLBACK.sql: pre-rollback');

UPDATE agent_definitions
   SET default_config = default_config
         #- '{workflow,steps,apply_edit,config,allow_rendered_html_transform}',
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #> '{workflow,steps,apply_edit,config,allow_rendered_html_transform}' IS NOT NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,apply_edit,config,input_fields}',
         COALESCE((
           SELECT jsonb_agg(f)
             FROM jsonb_array_elements(default_config #> '{workflow,steps,apply_edit,config,input_fields}') f
            WHERE f <> '"transform_name"'::jsonb
         ), '[]'::jsonb)),
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND (default_config #> '{workflow,steps,apply_edit,config,input_fields}') @> '["transform_name"]'::jsonb;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'section-editor' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND ( default_config #> '{workflow,steps,apply_edit,config,allow_rendered_html_transform}' IS NOT NULL
        OR (default_config #> '{workflow,steps,apply_edit,config,input_fields}') @> '["transform_name"]'::jsonb );
  IF n <> 0 THEN
    RAISE EXCEPTION '513 ROLLBACK: % live section-editor rows still carry the flag or the input_fields entry', n;
  END IF;
  RAISE NOTICE '513 ROLLBACK OK: flag and input_fields entry removed.';
END $$;

COMMIT;
