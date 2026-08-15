-- ROLLBACK for 420_site_planner_requires_backend_gate.sql
-- Restores load_available_components to its pre-420 text. No params key was ever added, so
-- none is removed.

BEGIN;

SELECT snapshot_agent('site-planner',
  '420_site_planner_requires_backend_gate_ROLLBACK: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true AND NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '420 ROLLBACK: expected exactly 1 live site-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_available_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '420 ROLLBACK: current query does not match 420''s post-state (concurrent edit since 420 applied?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_available_components,config,query}',
           to_jsonb('SELECT name, display_name, "function", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name'::text)
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
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '420 ROLLBACK: post-update, expected exactly 1 live site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}'
    INTO q
    FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) <> 0 THEN
    RAISE EXCEPTION '420 ROLLBACK: query still carries the exclusion: %', q;
  END IF;
END $$;

COMMIT;
