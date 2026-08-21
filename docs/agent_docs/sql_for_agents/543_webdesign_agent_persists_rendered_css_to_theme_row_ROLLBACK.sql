-- 543_webdesign_agent_persists_rendered_css_to_theme_row_ROLLBACK.sql
--
-- Reverses 543: removes persist_css_to_theme and restores generate_css → deploy_css.
--
-- ⚠ WHAT ROLLING BACK COSTS: the theme row stops tracking the deployed stylesheet,
-- so every backfilled row goes stale again at that site's next design run — with
-- no symptom until a css-patch dispatch arrives. Rows already persisted are left
-- as they are (they are correct; they simply stop being maintained). The 542 guard
-- still refuses to deploy an unsafe base, so the failure mode after a rollback is
-- refused patches, not clobbered sites.

BEGIN;

SELECT snapshot_agent('webdesign-agent', '543_ROLLBACK: pre-revert');

-- restore the edge first so no step is orphaned
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,generate_css,next_step}',
         to_jsonb('deploy_css'::text)
       ),
       updated_at = NOW()
 WHERE type = 'webdesign-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,persist_css_to_theme}',
       updated_at = NOW()
 WHERE type = 'webdesign-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'webdesign-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps ? 'persist_css_to_theme' THEN
        RAISE EXCEPTION '198/543 ROLLBACK: persist_css_to_theme survives';
    END IF;
    IF v_steps #>> '{generate_css,next_step}' <> 'deploy_css' THEN
        RAISE EXCEPTION '198/543 ROLLBACK: generate_css.next_step not restored — deploy is orphaned';
    END IF;

    RAISE NOTICE '198/543 ROLLBACK: verified — pre-543 shape restored (the row stops tracking the file)';
END $$;

COMMIT;
