-- ROLLBACK for 407 — restore build-site-planner's load_components to the
-- pre-widening menu (section/element only, no params).
--
-- This restores the exact query text the live row carried before 407 (also
-- preserved in the snapshot 407 takes: snapshot_agent note
-- '407_build_site_planner_sees_the_sites_own_tools: pre-update'). It REMOVES
-- the params key entirely rather than emptying it, matching the pre-407
-- config shape.

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '407 ROLLBACK: pre-rollback');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_components,config}',
         (default_config#>'{workflow,steps,load_components,config}')
           - 'params'
           || jsonb_build_object('query',
                'SELECT name, display_name, "function", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name')
       ),
       updated_at = now()
 WHERE type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  q text;
  p jsonb;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>'{workflow,steps,load_components,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('plan_includes_tools' in q) > 0 THEN
    RAISE EXCEPTION '407 ROLLBACK: query still carries the widening: %', q;
  END IF;
  IF p IS NOT NULL THEN
    RAISE EXCEPTION '407 ROLLBACK: params key still present: %', p;
  END IF;
END $$;

COMMIT;
