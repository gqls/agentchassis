-- ROLLBACK for 630: restores the pre-630 default_config from migration_backups.
BEGIN;
UPDATE agent_definitions ad
   SET default_config = (SELECT mb.old_value->'default_config' FROM migration_backups mb
                          WHERE mb.migration_name='630_planner_mission_constraints_stay_inside_the_plan'
                          ORDER BY mb.applied_at DESC LIMIT 1)
 WHERE ad.type='build-site-planner' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
COMMIT;
