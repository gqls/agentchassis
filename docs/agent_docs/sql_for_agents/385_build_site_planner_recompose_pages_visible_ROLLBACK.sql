-- 385_build_site_planner_recompose_pages_visible_ROLLBACK.sql
-- Restores build-site-planner's default_config from the pre-apply copy taken by
-- the seed (agent_definitions_bak_385). Verifies the restore removed the marker.
\set ON_ERROR_STOP on
BEGIN;
UPDATE agent_definitions ad
   SET default_config = b.default_config, updated_at = NOW()
  FROM agent_definitions_bak_385 b
 WHERE ad.id = b.id;
DO $do$
DECLARE t text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_site'->'config'->>'prompt_template'
      INTO t FROM agent_definitions
     WHERE type = 'build-site-planner' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF position('REDESIGN REQUESTED' in t) > 0 THEN
        RAISE EXCEPTION '385-ROLLBACK: verify failed - the marker is still present';
    END IF;
    IF length(t) <> 19685 THEN
        RAISE EXCEPTION '385-ROLLBACK: verify failed - restored length % <> 19685', length(t);
    END IF;
END $do$;
COMMIT;
