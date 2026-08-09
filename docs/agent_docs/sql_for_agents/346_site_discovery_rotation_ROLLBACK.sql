-- 346_site_discovery_rotation_ROLLBACK.sql
--
-- Removes the bugs_open/230 discovery rotation entirely: the three scheduled
-- tasks and the stamp table. To PAUSE rather than remove, prefer:
--   UPDATE scheduled_tasks SET enabled=false WHERE name LIKE 'site-discovery-rotation-%';
-- (pausing keeps the stamps, so a later re-enable resumes the rotation where it
-- stopped instead of re-examining everything from scratch).

BEGIN;

DELETE FROM scheduled_tasks WHERE name LIKE 'site-discovery-rotation-%';
DROP TABLE IF EXISTS site_discovery_rotation;
DELETE FROM doc_notes WHERE source = 'migration-346' AND subject_key = 'site-discovery-rotation';

DO $$
DECLARE remaining integer;
BEGIN
    SELECT count(*) INTO remaining FROM scheduled_tasks WHERE name LIKE 'site-discovery-rotation-%';
    IF remaining <> 0 OR to_regclass('public.site_discovery_rotation') IS NOT NULL THEN
        RAISE EXCEPTION 'rollback guard: rotation artefacts still present';
    END IF;
END $$;

COMMIT;
