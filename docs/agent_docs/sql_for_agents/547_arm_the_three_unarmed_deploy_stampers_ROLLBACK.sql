-- 547_arm_the_three_unarmed_deploy_stampers_ROLLBACK.sql
--
-- Un-arms the three. Restores the exact prior behaviour: they stamp `deployed`
-- without reading the deploy step's result, and they leave `pages.content_hash`
-- untouched.
--
-- ⚠ RUNNING THIS RE-OPENS THE HAZARD 547 CLOSED, and migration 526's gate will
-- refuse to enable `page_content_divergence` again afterwards — which is the
-- intended interlock, not a fault. If the check is ALREADY enabled when you run
-- this, disable it too (526_..._ROLLBACK.sql) or it may begin reporting healthy
-- pages as diverged the first time one of these agents deploys.
--
-- WHEN TO RUN: page-publishing failures mentioning `deploy_result_field`, or a
-- flood of DEPLOY_EVIDENCE_UNREADABLE from these three agents. Restore service
-- first, investigate after.

BEGIN;

SELECT snapshot_agent('page-rebuild',           '547_ROLLBACK: pre-update');
SELECT snapshot_agent('pageflow-builder',       '547_ROLLBACK: pre-update');
SELECT snapshot_agent('site-work-orchestrator', '547_ROLLBACK: pre-update');

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config,deploy_result_field}',
       updated_at = NOW()
 WHERE type IN ('page-rebuild','pageflow-builder')
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,build_items_loop,config,sub_workflow,steps,update_page_status,config,deploy_result_field}',
       updated_at = NOW()
 WHERE type = 'site-work-orchestrator'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE still int; stamps int;
BEGIN
    SELECT count(*) INTO still
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator')
       AND (step->'config'->>'deploy_result_field') IS NOT NULL;
    IF still <> 0 THEN
        RAISE EXCEPTION '547 ROLLBACK verify: % of the three steps still name a deploy_result_field', still;
    END IF;

    -- The stamps themselves must SURVIVE. A rollback that removed the step, or
    -- the whole config object, would silently stop these agents marking pages
    -- deployed at all — far worse than the hazard being rolled back.
    SELECT count(*) INTO stamps
      FROM agent_definitions ad
      CROSS JOIN LATERAL jsonb_path_query(ad.default_config,
            '$.**{0 to 25} ? (@.action == "update_page_status" && @.config.status == "deployed")') AS step
     WHERE ad.is_active AND NOT COALESCE(ad.is_snapshot,false) AND ad.deleted_at IS NULL
       AND ad.type IN ('page-rebuild','pageflow-builder','site-work-orchestrator');
    IF stamps <> 3 THEN
        RAISE EXCEPTION '547 ROLLBACK verify: expected the 3 deployed-stamping steps to survive un-arming, found %', stamps;
    END IF;
END $$;

COMMIT;
