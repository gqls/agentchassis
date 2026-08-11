-- FILE: docs/agent_docs/sql_for_agents/380_bugfix_238_arm_dead_url_guard_HOLD.sql
--
-- bugs_open/238 — arm the dead-URL refusal on the one live consumer of
-- render_component.
--
-- ⚠ _HOLD: ORDERING-CRITICAL, DO NOT LET THE RUNNER TAKE THIS.
-- The Go half (`shouldRefuseDeadURLControls`, dead_url_guard.go) is committed
-- but INERT until the chassis image rolls. This file only sets a config key, and
-- a config key that names behaviour the running binary does not have is a no-op
-- that LOOKS applied — the worst of both states, because the coverage report and
-- the register entry would then both say "armed" while nothing is guarded.
--
-- APPLY ONLY AFTER the running chassis is confirmed to carry the Go half. Ask
-- the service what it is running — do NOT reach for `strings`, which is absent
-- from the debian-slim images and whose failure is indistinguishable from "not
-- stamped" (three wrong readings in one day; bugs_open/249, CLAUDE.md rewritten
-- 2026-08-11):
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--     | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the commit that added dead_url_guard.go> <that sha>
-- Read the stamp of THIS service, not the fleet tag: until `release` pins one
-- REF, a single tag can straddle several revisions (v1.0.1284 shipped three).
--
-- WHAT IT DOES. Sets `refuse_dead_url_controls: true` on page-content-writer's
-- `render_section` step. Measured 2026-08-10, that agent is the ONLY live
-- definition with a render_component step, so this one key is full live
-- coverage — and reverting is the same key set back to false.
--
-- WHAT IT COSTS, stated plainly because it is the point of the owner decision:
-- with this armed, a section whose template has an UNGATED {{.field}} inside
-- src=/href= that resolves to nothing will REFUSE to render. The step fails,
-- save_page_sections never runs, and the stored row plus the live page are left
-- exactly as they were. A `dead_url_control` work item is filed naming page,
-- slot and fields. That means a page in the damaged state cannot be rebuilt
-- until its data is fixed — which is the intent (a queued item beats a silent
-- regression), but it IS a live blocking behaviour and it is why the Go half
-- defaults OFF (owner ruling 2026-08-02: new authority on a shared seam ships
-- as an opt-in field with the unsafe default off).
--
-- KNOWN POPULATION AT WRITING (measured 2026-08-10, re-measure before applying):
-- 5 deployed rows across 4 sites lack at least one *_image_url. finetuning.uk's
-- was repaired by 378/379. The other four — ai-agent-orchestration.com
-- /index.html, leopardessconsulting.co.uk /blog.html and its
-- automation-savings-estimator tool page, oufe.com's recovery-waterfall tool
-- page — would refuse on their next rebuild. Three of those four are the
-- "never had the value" class (no candidate assets exist), so the refusal is
-- correct and the item is the right outcome, but nobody should be surprised by it.
--
-- The gated fields are NOT affected: missingBareFields walks root-scope actions
-- only, so a {{if .card1_link_url}} guard keeps its field out of the report and
-- out of this refusal.
--
-- Rollback: 380_bugfix_238_arm_dead_url_guard_ROLLBACK.sql (sets it back to false).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('page-content-writer', '380_bugfix_238_arm_dead_url_guard: pre-update');

\echo '=== BEFORE: the render_section step config ==='
SELECT jsonb_pretty(jsonb_path_query_first(default_config, '$.**.steps.render_section.config'))
  FROM agent_definitions
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_rows int;
    v_armed boolean;
BEGIN
    SELECT count(*) INTO v_rows
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '238/380: expected exactly 1 live page-content-writer row, found % — resolve before editing config', v_rows;
    END IF;

    SELECT (jsonb_path_query_first(default_config,
              '$.**.steps.render_section.config.refuse_dead_url_controls'))::text::boolean
      INTO v_armed
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_armed IS TRUE THEN
        RAISE EXCEPTION '238/380: already applied — refuse_dead_url_controls is already true';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           -- The step lives inside the process_sections_loop sub_workflow, not at
           -- the top level: a top-level jsonb_each finds nothing here and reads as
           -- "no such step" (the census trap this file's family keeps hitting).
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,refuse_dead_url_controls}',
           'true'::jsonb,
           true),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

\echo '=== AFTER ==='
SELECT jsonb_pretty(jsonb_path_query_first(default_config, '$.**.steps.render_section.config'))
  FROM agent_definitions
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    v_armed boolean;
    v_action text;
BEGIN
    SELECT (jsonb_path_query_first(default_config,
              '$.**.steps.render_section.config.refuse_dead_url_controls'))::text::boolean,
           jsonb_path_query_first(default_config, '$.**.steps.render_section.action') #>> '{}'
      INTO v_armed, v_action
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_armed IS NOT TRUE THEN
        RAISE EXCEPTION '238/380: the key did not land — the jsonb path is wrong for this row shape; aborting rather than reporting a flip that did not happen';
    END IF;
    -- The key must sit on a render_component step or it guards nothing at all.
    IF v_action IS DISTINCT FROM 'render_component' THEN
        RAISE EXCEPTION '238/380: render_section.action is %, not render_component — the key landed on the wrong step; aborting', v_action;
    END IF;
    RAISE NOTICE '238/380: dead-URL refusal ARMED on page-content-writer.render_section';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- WATCH after arming — the refusal is a step failure, so it surfaces as one:
--   SELECT count(*), max(created_at) FROM site_work_items
--    WHERE item_type = 'dead_url_control';
-- The first non-zero count is also the first real frequency measurement of this
-- class. A sustained zero means either nothing is regenerating a damaged section
-- (plausible — the improvement loop is paused) or the flag did not take; the
-- pod-grep above discriminates.
-- ---------------------------------------------------------------------------
