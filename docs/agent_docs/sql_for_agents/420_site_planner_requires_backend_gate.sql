-- 420 — site-planner: requires-backend SECTION exclusion, unconditional (no capability check)
-- Third and last call site for bugs_open/276 (see 418's header for the full evidence base).
--
-- WHY UNCONDITIONAL, NOT A CAPABILITY GATE LIKE 418/419. site-planner has ZERO dispatches
-- ever (checked with no time bound at all, not just a 30-day window, 2026-08-15) — it is
-- dead/legacy, apparently superseded by build-site-planner. Its workflow
-- (load_available_components -> load_style_collections -> plan_site -> validate_plan ->
-- complete) has NO ensure_site_record-equivalent step anywhere, so there is no proven
-- collected_data key holding a site id to bind as $1 — unlike 418/419, where a sibling step
-- in the same workflow chain was confirmed live-resolving before this migration was written.
-- A params path that resolves to nil hard-fails the step (an outage-class failure, per 407's
-- own header warning on the same class of change). Guessing an unproven binding on code that
-- has never executed risks planting exactly that landmine for zero present benefit.
--
-- Decisive call: gate this step too (closes the door completely, per house preference), but
-- by unconditionally excluding requires-backend-tagged components — no site/capability check,
-- no params change. This is strictly MORE conservative than the 418/419 gate (removes the tag
-- for every caller, not just callers without backend capability), which is the correct
-- direction for a call site with no site context to check against. If site-planner is ever
-- revived, whoever wires it up does the same site-id binding legwork 418/419 did and earns
-- the same opt-in-for-capable-sites behaviour properly, rather than inheriting a guess.
--
-- ID-SCOPED + PRE-STATE GATED, same discipline as 418/419.
--
-- Config-only: no image dependency, live on apply. Inert in practice (0 live dispatches to
-- affect) but removes a silent, never-checked default before it can ever fire.
--
-- ROLLBACK: 420_site_planner_requires_backend_gate_ROLLBACK.sql, or restore the snapshot
-- this file takes (snapshot_agent note '420_site_planner_requires_backend_gate: pre-update').

BEGIN;

SELECT snapshot_agent('site-planner',
  '420_site_planner_requires_backend_gate: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '420: expected exactly 1 live site-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_available_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '420: load_available_components pre-state does not match expected text (concurrent edit?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_available_components,config,query}',
           to_jsonb('SELECT name, display_name, "function", category, description FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true AND NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') ORDER BY category, name'::text)
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
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '420: post-update, expected exactly 1 live site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         (default_config#>'{workflow,steps,load_available_components,config,params}') IS NOT NULL
    INTO q, has_params
    FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) = 0 THEN
    RAISE EXCEPTION '420: load_available_components query does not carry the exclusion: %', q;
  END IF;
  IF position('EXISTS' in q) <> 0 OR position('$1' in q) <> 0 THEN
    RAISE EXCEPTION '420: query unexpectedly gained a site/capability check (should be unconditional): %', q;
  END IF;
  IF has_params THEN
    RAISE EXCEPTION '420: params key unexpectedly present (this migration adds none): found %', has_params;
  END IF;
END $$;

COMMIT;
