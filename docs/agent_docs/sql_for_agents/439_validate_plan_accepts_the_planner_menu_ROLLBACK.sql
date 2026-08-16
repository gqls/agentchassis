-- ROLLBACK for 439 — remove validate_plan's menu_field key, restoring the
-- section/element-only acceptance surface (bugs_open/282 reopens on apply).
--
-- Removing the key is the whole rollback: the Go arm is a no-op without it, so
-- this reverts behaviour whether or not the image carrying the Go half has
-- rolled.

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '439_validate_plan_accepts_the_planner_menu: pre-rollback');

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,validate_plan,config,menu_field}',
       updated_at = now()
 WHERE type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  still text;
BEGIN
  SELECT default_config#>>'{workflow,steps,validate_plan,config,menu_field}'
    INTO still
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF still IS NOT NULL THEN
    RAISE EXCEPTION '439 rollback: menu_field is still present (%)', still;
  END IF;
END $$;

COMMIT;
