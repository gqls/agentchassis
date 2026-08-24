-- ROLLBACK for 590 — stop the sitemap rotation and retire its agent.
--
-- What this does NOT undo: any sitemap.xml already committed to a site. Those
-- are correct files describing real pages, and they do not become wrong because
-- the thing that generated them was switched off. They will simply stop being
-- refreshed — which is exactly the pre-590 situation the council objected to.
-- Remove them by hand if that is genuinely what you want, and say why.
--
-- It also leaves the site_discovery_rotation stamps for agent_type
-- 'sitemap-refresh'. They are harmless (nothing reads them once the task is
-- gone) and keeping them means a re-apply resumes the rotation where it stopped
-- rather than re-probing all 28 sites at once.

BEGIN;

DELETE FROM scheduled_tasks WHERE name = 'sitemap-refresh-rotation';

UPDATE agent_definitions SET is_active = false, deleted_at = now()
WHERE type = 'sitemap-refresh' AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DELETE FROM schema_migrations WHERE filename = '590_wire_render_sitemap_into_a_rotation.sql';

DO $$
DECLARE n_task int; n_agent int;
BEGIN
    SELECT count(*) INTO n_task FROM scheduled_tasks WHERE name='sitemap-refresh-rotation';
    IF n_task <> 0 THEN RAISE EXCEPTION 'rollback: % rotation task rows remain', n_task; END IF;

    SELECT count(*) INTO n_agent FROM agent_definitions
     WHERE type='sitemap-refresh' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_agent <> 0 THEN RAISE EXCEPTION 'rollback: % active sitemap-refresh rows remain', n_agent; END IF;

    RAISE NOTICE '590 rollback OK — rotation removed, agent retired.';
END $$;

COMMIT;
