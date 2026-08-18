-- 462_copy_editor_edit_budget.sql
--
-- Fixes what stage 2's SECOND run found, on the first page that was not easy.
--
-- WHAT HAPPENED `[MEASURED 2026-08-18]`. `copy-editor` (CQ-024, seeded 447) was run
-- against `ai-agent-orchestration.com/index` — 8 components, 78,302 chars of payload,
-- and a DIFFUSE register fault: 15 define-by-negation constructions ("X, not Y" ×8,
-- "rather than" ×7) spread across the page, where the v2 house voice allows a matched
-- contrasting pair "once or twice per page at most".
--
-- The run FAILED at `run_copy_edit`:
--   AI call failed: response truncated: stop_reason=max_tokens
--   (output_tokens=16000 reached the configured cap, 9687 chars recovered)
--
-- TWO SEPARATE FAULTS, and only one of them is the token cap:
--
-- 1. **The output cap was too low for the shape of the task.** Raised 16,000 → 32,000.
--
-- 2. **The agent tried to rewrite the WHOLE PAGE, which the design forbids.** This is
--    the real finding. On the proof case (one component, one obviously missing link
--    set) it was surgical — six added lines, nothing else. Given a fault spread thinly
--    across eight components it attempted everything at once. So the restraint
--    observed on 2026-08-17 was a property of a LEGIBLE DEFECT, not of the design, and
--    the honest reading of run 1 is now narrower than it looked that evening.
--    PLAN §3 rule 3 already said the write is "one component at a time"; the config
--    asked for page-scoped read AND every write in a single completion, which is not
--    the same thing and is what broke.
--
-- THE FIX IS A BUDGET, NOT JUST A BIGGER CAP. The prompt now names an explicit ceiling
-- of THREE edits per run, ordered by what most improves the page, and says plainly that
-- a page needing more than three is a page to run again rather than to rewrite wholesale.
-- Rationale: an edit set bounded at the SOURCE cannot truncate, and a truncated
-- completion is the one failure this estate has repeatedly seen persist as a fragment
-- while reporting success (`bugs_open/012`). Here it failed loudly instead — the
-- platform's `stop_reason` check did its job — and the budget is what stops it arising.
--
-- ⚠ The read stays FULL. The page-scoped read is the whole point of stage 2 (a
-- section-scoped writer cannot judge order, one-name-per-thing, or restatement between
-- sections), so nothing here trims what the agent SEES. Input was never the problem:
-- 78,302 chars is roughly 20k tokens and the call was nowhere near an input limit.
--
-- ROLLBACK: 462_copy_editor_edit_budget_ROLLBACK.sql (restores 16,000 + the old prompt).

BEGIN;

SELECT snapshot_agent('copy-editor', '462_copy_editor_edit_budget.sql: pre-update');

-- 1. the cap
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_copy_edit,config,ai_service,max_tokens}', '32000'::jsonb, false),
       updated_at = now()
 WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 2. the budget, inserted verbatim ahead of the output contract so it is read as a
--    constraint on WHAT TO EMIT rather than as advice about style.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_copy_edit,config,prompt}',
         to_jsonb(replace(
           default_config #>> '{workflow,steps,run_copy_edit,config,prompt}',
           'Return ONLY JSON, no prose around it:',
'HOW MANY EDITS. At most THREE, and fewer is better. If more than three components need
work, edit the three that most improve the page for a reader and say so in
page_judgement — this page will be run again. Do not attempt a whole-page rewrite: it is
not what this stage is for, the edits are applied one component at a time, and an
oversized answer is truncated rather than accepted. Rank by what a reader loses if it
stays as it is, not by how easy the change would be to make.

Return ONLY JSON, no prose around it:')),
         false),
       updated_at = now()
 WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF (cfg #>> '{workflow,steps,run_copy_edit,config,ai_service,max_tokens}') <> '32000' THEN
    RAISE EXCEPTION 'max_tokens not raised (found %)', cfg #>> '{workflow,steps,run_copy_edit,config,ai_service,max_tokens}';
  END IF;

  IF (cfg #>> '{workflow,steps,run_copy_edit,config,prompt}') NOT LIKE '%At most THREE%' THEN
    RAISE EXCEPTION 'the edit budget did not land in the prompt — the anchor line must have drifted';
  END IF;

  -- The 447 invariants must survive this edit. A prompt rewrite that dropped the voice
  -- carrier would degrade silently to no house voice at all (CQ-022).
  IF (cfg #>> '{workflow,steps,run_copy_edit,config,prompt}') NOT LIKE '%{{.voice_style}}%' THEN
    RAISE EXCEPTION '447 invariant broken: the prompt no longer references {{.voice_style}}';
  END IF;

  IF EXISTS (SELECT 1 FROM jsonb_each(cfg->'workflow'->'steps') s(k, v)
              WHERE s.v->>'action' IN ('apply_section_edit','save_page_sections','render_component',
                                       'update_component_html','rerender_page_sections','git_commit')) THEN
    RAISE EXCEPTION '447 invariant broken: copy-editor gained a page-writing step';
  END IF;

  RAISE NOTICE 'copy-editor: max_tokens 32000, 3-edit budget live, 447 invariants intact';
END $$;

COMMIT;
