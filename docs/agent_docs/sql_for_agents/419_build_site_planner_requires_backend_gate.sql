-- 419 — build-site-planner: requires-backend eligibility gate on the SECTION menu
-- (VMB-010's planner half, the call site bugs_open/276 explicitly names)
--
-- WHY. Sibling of 418 (content-gap-planner) and 406 (tool-suggester, the tool-level gate).
-- See 418's header for the full evidence base (fleet-wide sweep, dispatch volumes,
-- intent-probe's sole current placement already being backend-capable, why no fresh
-- architecture RFC is needed). This file is the minority-traffic call site (2 dispatches in
-- the last 30 days, measured 2026-08-15, vs content-gap-planner's 131) but is the one both
-- bugs_open/276 and the concept register (VMB-010) name directly.
--
-- SHAPE. build-site-planner's load_components step already binds $1 = site_record.site_id
-- (migration 407, 2026-08-14, for an unrelated reason — widening this same step to also show
-- a site's own already-placed TOOL components under a separate opt-in flag
-- plan_includes_tools). This migration reuses that existing binding — no params change.
-- The gate is added as a new top-level AND wrapping the existing
-- ( component_level IN ('section','element') OR ( component_level = 'tool' AND ... ) )
-- group, so it applies uniformly to BOTH branches. This is harmless to the tool branch: a
-- tool only reaches that branch if it is already placed on the planning site's own pages, so
-- a requires-backend tool already placed on a static site (which should not have happened,
-- and no such case exists today) would stop being re-offered for re-planning rather than
-- newly granted — never a widening.
--
-- ID-SCOPED + PRE-STATE GATED, same discipline as 418/406's addendum guidance.
--
-- CONSUMERS of load_components's output: the site-planning LLM step of build-site-planner
-- (output_field available_components), nothing else.
--
-- Config-only: no image dependency, live on apply.
--
-- ROLLBACK: 419_build_site_planner_requires_backend_gate_ROLLBACK.sql, or restore the
-- snapshot this file takes (snapshot_agent note
-- '419_build_site_planner_requires_backend_gate: pre-update').

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '419_build_site_planner_requires_backend_gate: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category, description FROM content_components WHERE is_active = true AND ( component_level IN (''section'',''element'') OR ( component_level = ''tool'' AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = ''structure'' AND ss.is_current AND ss.data->>''plan_includes_tools'' = ''true'') AND id IN (SELECT pc.component_id FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1) ) ) ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '419: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '419: load_components pre-state does not match expected text (concurrent edit?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_components,config,query}',
           to_jsonb('SELECT name, display_name, "function", category, description FROM content_components WHERE is_active = true AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) AND ( component_level IN (''section'',''element'') OR ( component_level = ''tool'' AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = ''structure'' AND ss.is_current AND ss.data->>''plan_includes_tools'' = ''true'') AND id IN (SELECT pc.component_id FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1) ) ) ORDER BY category, name'::text)
         ),
         updated_at = now()
   WHERE id = target_id;
END $$;

DO $$
DECLARE
  q text;
  p jsonb;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '419: post-update, expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>'{workflow,steps,load_components,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) = 0
     OR position('deploy_config' in q) = 0
     OR position('plan_includes_tools' in q) = 0 THEN
    RAISE EXCEPTION '419: load_components query does not carry both gates: %', q;
  END IF;
  IF p IS NULL OR jsonb_array_length(p) <> 1
     OR p->>0 <> 'site_record.site_id' THEN
    RAISE EXCEPTION '419: load_components params changed unexpectedly: %', p;
  END IF;
END $$;

COMMIT;
