-- FILE: docs/agent_docs/sql_for_agents/380_bugfix_238_arm_dead_url_guard_HOLD.sql
--
-- bugs_open/238 — arm the dead-URL refusal on BOTH render_component steps of
-- page-content-writer (the only live agent that has any).
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
-- `render_section` AND `render_from_template` steps.
--
-- ⚠ CORRECTED 2026-08-11 (council 98852baa round 2, debug_historian): this file
-- originally armed ONE step and called it "full live coverage", on a census that
-- asked `default_config::text LIKE '%render_component%'`. That question counts
-- AGENTS (answer 1), not STEPS, and `_` is a SQL wildcard into the bargain. The
-- honest count, from a jsonb path over `$.**.steps`, is TWO — so the original
-- would have left `render_from_template` unguarded while the register and the
-- coverage report both said "armed". The verify block now counts every
-- render_component step and demands they ALL carry the flag, so a third step
-- fails this file loudly instead of shipping silently unguarded.
--
-- Reverting is the same keys set back to false (the ROLLBACK sidecar).
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

-- ⚠ THE DUPLICATE-ROW GUARD, added 2026-08-11 after the council's debug_historian
-- seat raised it (corr 98852baa). FOUR agent types on this estate carry TWO
-- active definition rows, and only the HIGHER VERSION is ever loaded at runtime —
-- so an `UPDATE ... WHERE type = '<x>'` can silently touch a stale duplicate, or
-- touch both, while a verify block that re-reads "the row for this type" happily
-- confirms whichever one it happens to find. Measured for THIS type 2026-08-11:
-- exactly 1 active row (id 5946a27b-…, version 2). The guard below is therefore
-- expected to pass today; it exists so that if a second row appears before this
-- HOLD is lifted, the file refuses instead of half-applying.
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
        RAISE EXCEPTION '238/380: expected exactly 1 live page-content-writer row, found % — this type now carries duplicates and only the HIGHEST VERSION is loaded at runtime; target that row by id, do not update by type', v_rows;
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

-- ⚠ PRE-FLIGHT: assert the paths EXIST and hold a render_component step, before
-- any jsonb_set runs. `jsonb_set(..., create_missing := true)` on a wrong path
-- inserts a whole new branch and reports success — arming nothing while every
-- downstream reader says "armed". The council's editquality seat gated round 2
-- on exactly this, and it was right that the original file asserted the nesting
-- rather than checking it.
DO $$
DECLARE
    v_a text;
    v_b text;
BEGIN
    SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}',
           default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_from_template,action}'
      INTO v_a, v_b
      FROM agent_definitions
     WHERE type = 'page-content-writer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_a IS DISTINCT FROM 'render_component' THEN
        RAISE EXCEPTION '238/380: render_section is % at the expected path (want render_component) — the workflow shape moved; re-derive the path before arming', COALESCE(v_a, '(absent)');
    END IF;
    IF v_b IS DISTINCT FROM 'render_component' THEN
        RAISE EXCEPTION '238/380: render_from_template is % at the expected path (want render_component) — the workflow shape moved; re-derive the path before arming', COALESCE(v_b, '(absent)');
    END IF;
END $$;

-- BOTH render_component steps, not one.
--
-- ⚠ THE ORIGINAL VERSION OF THIS FILE ARMED ONLY `render_section`, on a census
-- that said "exactly one live agent has a render_component step". That census
-- counted AGENTS with `default_config::text LIKE '%render_component%'` — which
-- also treats `_` as a wildcard — and the honest question was how many STEPS.
-- Re-measured with a jsonb path over `$.**.steps`: TWO, `render_section` and
-- `render_from_template`. Arming one would have left the second render path
-- unguarded while the coverage report and the register both said "armed".
UPDATE agent_definitions
   SET default_config = jsonb_set(
           jsonb_set(
               default_config,
               -- The steps live inside the process_sections_loop sub_workflow, not
               -- at the top level: a top-level jsonb_each finds nothing here and
               -- reads as "no such step" (the census trap this family keeps hitting).
               '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,refuse_dead_url_controls}',
               'true'::jsonb,
               true),
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_from_template,config,refuse_dead_url_controls}',
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

-- VERIFY BY COUNTING, not by reading one path back.
--
-- Two disciplines here, both learned the hard way and both flagged by the
-- council's debug_historian seat. (1) A jsonb path compared with `=`/`<>` sits
-- GREEN for ever when the key is ABSENT, because NULL <> 'true' is NULL, not
-- TRUE — so every comparison below is `IS DISTINCT FROM` / `IS NOT TRUE`, which
-- treat absence as failure. (2) Reading back the one path you just wrote proves
-- only that you wrote it; it cannot see a render_component step you never armed.
-- So this counts ALL render_component steps and asserts that ALL of them carry
-- the flag — which means a future THIRD step fails this migration loudly instead
-- of shipping silently unguarded.
DO $$
DECLARE
    v_steps  int;
    v_armed  int;
BEGIN
    SELECT count(*) FILTER (WHERE k.value->>'action' = 'render_component'),
           count(*) FILTER (WHERE k.value->>'action' = 'render_component'
                              AND (k.value->'config'->>'refuse_dead_url_controls')::boolean IS TRUE)
      INTO v_steps, v_armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') AS steps,
           LATERAL jsonb_each(steps) AS k
     WHERE ad.type = 'page-content-writer' AND ad.is_active
       AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL;

    IF v_steps = 0 THEN
        RAISE EXCEPTION '238/380: found ZERO render_component steps — the traversal is wrong, so a green result here would mean nothing; aborting';
    END IF;
    IF v_armed <> v_steps THEN
        RAISE EXCEPTION '238/380: % of % render_component step(s) armed — a step exists that this file does not know about; add it rather than shipping partial coverage the report will call complete', v_armed, v_steps;
    END IF;
    RAISE NOTICE '238/380: dead-URL refusal ARMED on ALL % render_component step(s) of page-content-writer', v_steps;
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
