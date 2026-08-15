-- ROLLBACK for 418_content_gap_planner_requires_backend_gate.sql
-- Restores load_available_components to its pre-418 text and removes the params key that
-- 418 introduced (it did not exist before 418; #- removes it rather than setting it null).
-- Id-scoped + pre-state gated, same discipline as the forward migration.

BEGIN;

SELECT snapshot_agent('content-gap-planner',
  '418_content_gap_planner_requires_backend_gate_ROLLBACK: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '418 ROLLBACK: expected exactly 1 live content-gap-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_available_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '418 ROLLBACK: current query does not match 418''s post-state (concurrent edit since 418 applied?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = (
           jsonb_set(
             default_config,
             '{workflow,steps,load_available_components,config,query}',
             to_jsonb('SELECT name, display_name, "function", category FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name'::text)
           ) #- '{workflow,steps,load_available_components,config,params}'
         ),
         updated_at = now()
   WHERE id = target_id;
END $$;

DO $$
DECLARE
  q text;
  has_params boolean;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '418 ROLLBACK: post-update, expected exactly 1 live content-gap-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         (default_config#>'{workflow,steps,load_available_components,config,params}') IS NOT NULL
    INTO q, has_params
    FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) <> 0 THEN
    RAISE EXCEPTION '418 ROLLBACK: query still carries the gate: %', q;
  END IF;
  IF has_params THEN
    RAISE EXCEPTION '418 ROLLBACK: params key still present, expected removed';
  END IF;
END $$;

COMMIT;
