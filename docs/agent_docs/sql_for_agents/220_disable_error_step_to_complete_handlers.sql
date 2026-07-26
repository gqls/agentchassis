-- ============================================================================
-- 220_disable_error_step_to_complete_handlers.sql — bugs_open/086 containment
--
-- WHY. Until v1.0.1169 the plan converter dropped step-level `error_step`, so
-- 55 declared handlers across 19 agents had never run. dca5649b3 fixed that and
-- shipped in v1.0.1169 (pod-verified), arming all 55 at once — which is exactly
-- what the council's guardian seat vetoed the change for.
--
-- Ten of the 55 route to `complete`, whose action is `complete_workflow`: a
-- failure at those steps would now end the orchestration GREEN. Five of them
-- would swallow the run's whole purpose (spec-updater.apply_update,
-- content-gap-planner.apply_plan, site-adoption-agent.write_design_intent,
-- blog-content-planner.create_post_pages, webdesign-agent.fork_theme); two are
-- borderline; three are defensible bookkeeping. None has fired in 30 days.
--
-- Owner ruling 2026-07-26: disable all ten, pending per-handler review. This
-- restores the behaviour the fleet has actually had for the last ten weeks —
-- these steps fail loudly — rather than inventing new routing.
--
-- HOW. Renames the key `error_step` -> `error_step_disabled_086` on exactly
-- those steps. The converter reads only `error_step`, so the renamed key is
-- inert; the author's declared target stays visible in the definition and
-- greppable, instead of being deleted. Untouched: every handler that routes
-- somewhere other than `complete`, and every `config.error_step`.
--
-- REVERT (per agent, or drop the type filter for all ten):
--   UPDATE agent_definitions d SET default_config = (
--     SELECT jsonb_set(d.default_config, ARRAY['workflow','steps',s.key,'error_step'],
--                      s.value->'error_step_disabled_086')
--            #- ARRAY['workflow','steps',s.key,'error_step_disabled_086']
--     FROM jsonb_each(d.default_config->'workflow'->'steps') s
--     WHERE s.value ? 'error_step_disabled_086' LIMIT 1)
--   WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
--     AND d.default_config::text LIKE '%error_step_disabled_086%';
--   -- (loop it: one step per pass, as the jsonb_set above rewrites one key)
-- ============================================================================

DO $$
DECLARE
  r          record;
  changed    int := 0;
  agents     text[] := '{}';
BEGIN
  -- snapshot every agent we are about to touch, before touching it
  FOR r IN
    SELECT DISTINCT type
    FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
    WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND value->>'error_step'='complete' AND value->'config'->>'error_step' IS NULL
  LOOP
    PERFORM snapshot_agent(r.type, '220_disable_error_step_to_complete_handlers.sql: pre-update');
    agents := agents || r.type;
  END LOOP;

  IF array_length(agents,1) IS NULL THEN
    RAISE NOTICE 'no step-level error_step->complete handlers found — already applied, or the set moved';
    RETURN;
  END IF;
  RAISE NOTICE 'snapshotted % agents: %', array_length(agents,1), agents;

  -- rename the key on each matching step, one step per statement
  LOOP
    SELECT d.id AS def_id, s.key AS step_name, d.type
      INTO r
    FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
      AND s.value->>'error_step'='complete'
      AND s.value->'config'->>'error_step' IS NULL
    LIMIT 1;

    EXIT WHEN NOT FOUND;

    UPDATE agent_definitions
    SET default_config = jsonb_set(
          default_config #- ARRAY['workflow','steps',r.step_name,'error_step'],
          ARRAY['workflow','steps',r.step_name,'error_step_disabled_086'],
          '"complete"'::jsonb)
    WHERE id = r.def_id;

    changed := changed + 1;
    RAISE NOTICE 'disabled %.% (was error_step -> complete)', r.type, r.step_name;

    EXIT WHEN changed > 20;  -- runaway guard; the known set is 10
  END LOOP;

  RAISE NOTICE 'disabled % handlers', changed;
  IF changed <> 10 THEN
    RAISE WARNING 'expected 10 handlers, changed % — read the post-check before trusting this', changed;
  END IF;
END $$;

-- post-check 1: nothing routes to complete at step level any more (want 0)
SELECT count(*) AS still_routing_to_complete
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step'='complete' AND value->'config'->>'error_step' IS NULL;

-- post-check 2: the ten are recorded, not deleted (want 10 rows)
SELECT type, key AS step, value->>'error_step_disabled_086' AS declared_target
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value ? 'error_step_disabled_086'
ORDER BY type, step;

-- post-check 3: the OTHER 45 handlers are untouched (want 45)
SELECT count(*) AS other_step_level_handlers_intact
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps')
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND value->>'error_step' IS NOT NULL AND value->'config'->>'error_step' IS NULL;
