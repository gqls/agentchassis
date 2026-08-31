-- ROLLBACK for 645 — disarm the four non-primary save_page_sections steps.
-- Leaves 643's two primary writers armed. Safe at any time: the pass records
-- only, so disarming loses instrumentation and changes no content.
-- ⚠ After this, the rate is biased again — see 645's header.
BEGIN;

UPDATE agent_definitions
SET default_config = default_config
        #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}'
        #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND type IN ('pageflow-builder', 'page-rebuild', 'site-work-orchestrator');

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,save_sections,config,audit_cta_label_agreement}',
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND type = 'tool-recreation-handler';

DO $$
DECLARE armed integer;
BEGIN
    SELECT count(*) INTO armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.steps.save_sections ? (@.action == "save_page_sections" && @.config.audit_cta_label_agreement == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;
    IF armed <> 2 THEN
        RAISE EXCEPTION 'rollback left % armed, expected 643''s 2 primary writers', armed;
    END IF;
END $$;

COMMIT;
