-- 509 — bugs_open/305: put the copy gate's REPAIR step into page-content-writer's
-- section loop, between generating a section and rendering it.
--
-- ⚠ HELD DELIBERATELY (_HOLD): APPLY ONLY AFTER THE IMAGE CARRYING
-- `rewrite_negations` IS LIVE.
--
-- This is not the usual "a new key is ignored by the old binary" case, and the
-- distinction is the whole reason for the suffix. This migration rewires the
-- step CHAIN: generate_content stops pointing at render_section and points at a
-- step whose action a pre-509 binary cannot resolve. Applied early, every
-- section build on the fleet fails at an unknown action instead of rendering.
-- `SIDECAR_RE` in scripts/migration/run-migrations.sh excludes `_HOLD` from
-- `--apply` while still listing it, so this file is visible and inert. Renaming
-- it (dropping `_HOLD`) is the deliberate act, and it has TWO preconditions, not
-- one (the second was added at the council's request, round 1, editquality and
-- bug_historian both on edit 7):
--
--   (1) THE IMAGE IS LIVE. Ask the binary, per service — not git, not the tag:
--       kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--       git merge-base --is-ancestor a5d5d0728 <that sha>
--       (an empty grep means the line SCROLLED, not that it is unstamped)
--
--   (2) THE PER-PAGE BUDGET CANARY HAS PASSED. The budget assumes CollectedData
--       persists across process_sections_loop iterations; that is read from the
--       loop design, not observed. Build ONE page of 3+ sections and read the
--       marker: page_hits must ACCUMULATE across sections rather than resetting.
--         SELECT collected_data->'__copy_gate'
--           FROM orchestration_states
--          WHERE collected_data ? '__copy_gate' ORDER BY updated_at DESC LIMIT 5;
--       If it resets, the gate still repairs every headline hit and simply
--       budgets per section — weaker than intended, not wrong. Apply anyway if
--       you accept that, but KNOW which one you shipped.
--
-- WHAT THE STEP DOES (full reasoning: platform/orchestration/actions/rewrite_negations_action.go):
-- scans the generated section for define-by-negation, leaves alone anything the
-- site's own brief supplied or the regulator requires, and asks the model ONCE
-- to rewrite the remaining sentences directly — beyond a budget of two per PAGE,
-- or any hit in a headline-class field. Each proposed rewrite is judged and
-- spliced individually. It never fails the step.
--
-- WHY page_budget = 2. The house voice's own standard, verbatim: "A very short
-- closing sentence or a matched contrasting pair is earned once or twice per
-- page at most". The count is per PAGE and rides in collected_data, because a
-- per-section threshold cannot express it — six sections at one construction
-- each is six on the page and every one passes.
--
-- WHY ONLY THIS AGENT. page-content-writer is 1,516 of 1,519 voice-carrier LLM
-- calls in the last 7 days [MEASURED 2026-08-19]. The COUNTING half needs no
-- migration at all: it is default-ON in the render and compile actions, so every
-- other writer's output is measured even though only this one is repaired.
--
-- ANCHORS, verified against the LIVE row 2026-08-20:
--   generate_content.action    = execute_llm_prompt
--   generate_content.next_step = render_section
--   render_section.action      = render_component
--   rewrite_negations          = ABSENT
-- Each UPDATE is anchored on those values, so a moved path means 0 rows and a
-- loud RAISE rather than an orphan key: jsonb_set with a missing parent returns
-- its input unchanged, and the verify below counts POSITIVE presence, so an
-- absent path yields NULL, is not counted, and raises. Needle-gated, so a re-run
-- is a legitimate 0-row no-op and the final-state check still passes.
--
-- Backup: snapshot_agent() (the standard idiom).

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '509_writer_rewrites_define_by_negation.sql: pre-update (bugs_open/305)');

-- 1. Add the step.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations}',
         '{
            "action": "rewrite_negations",
            "config": {
              "content_from": "generated_content.result",
              "page_budget": 2
            },
            "next_step": "render_section",
            "description": "Copy gate (bugs_open/305): rewrite define-by-negation sentences beyond the per-page budget, or any in a headline field. Brief-supplied and regulatory negations are counted, not rewritten. Never fails the step.",
            "output_field": "copy_gate"
          }'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,action}' = 'execute_llm_prompt'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}' = 'render_component'
   AND (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations}') IS NULL;

-- 2. Re-point the writer at it. Anchored on the CURRENT destination, so if
--    another session has already inserted a step here this does nothing rather
--    than stealing their edge.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}',
         '"rewrite_negations"'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}' = 'render_section'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations'
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,next_step}' = 'render_section'
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,next_step}' = 'rewrite_negations';
  IF n <> 1 THEN
    RAISE EXCEPTION '509 FAILED: expected exactly 1 page-content-writer with the chain generate_content -> rewrite_negations -> render_section, got % — read the live row and re-anchor before retrying', n;
  END IF;
  RAISE NOTICE '509 OK: generate_content -> rewrite_negations -> render_section';
END $$;

COMMIT;
