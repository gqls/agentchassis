-- ROLLBACK for 419_build_site_planner_requires_backend_gate.sql
-- Restores load_components to its pre-419 text (the 407 state). params untouched by 419,
-- so left as-is.

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '419_build_site_planner_requires_backend_gate_ROLLBACK: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category, description FROM content_components WHERE is_active = true AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) AND ( component_level IN (''section'',''element'') OR ( component_level = ''tool'' AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = ''structure'' AND ss.is_current AND ss.data->>''plan_includes_tools'' = ''true'') AND id IN (SELECT pc.component_id FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1) ) ) ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '419 ROLLBACK: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '419 ROLLBACK: current query does not match 419''s post-state (concurrent edit since 419 applied?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_components,config,query}',
           to_jsonb('SELECT name, display_name, "function", category, description FROM content_components WHERE is_active = true AND ( component_level IN (''section'',''element'') OR ( component_level = ''tool'' AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = ''structure'' AND ss.is_current AND ss.data->>''plan_includes_tools'' = ''true'') AND id IN (SELECT pc.component_id FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1) ) ) ORDER BY category, name'::text)
         ),
         updated_at = now()
   WHERE id = target_id;
END $$;

DO $$
DECLARE
  q text;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '419 ROLLBACK: post-update, expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_components,config,query}'
    INTO q
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) <> 0 THEN
    RAISE EXCEPTION '419 ROLLBACK: query still carries the gate: %', q;
  END IF;
  IF position('plan_includes_tools' in q) = 0 THEN
    RAISE EXCEPTION '419 ROLLBACK: 407''s tool-flag gate was lost: %', q;
  END IF;
END $$;

COMMIT;
