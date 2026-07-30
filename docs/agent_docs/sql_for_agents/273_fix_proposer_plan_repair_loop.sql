-- 273_fix_proposer_plan_repair_loop.sql
--
-- *** NOT APPLIED. This file exists to BE a visible trigger. ***
--
-- bugs_open/099 candidate 2, second consumer. `272_feature_designer_plan_repair_loop.sql`
-- opted `feature-designer` into the bounded plan-repair loop. `fix-proposer` uses the
-- SAME shared action (`diagnose_persist_fix_plan`, at the same step name
-- `persist_plan`) and therefore has the SAME defect: a structural validation failure
-- discards a completed, good plan, with `orchestration_states.error` NULL, so a
-- dashboard keyed on `error` reports the run as clean.
--
-- WHY IT IS WRITTEN BUT NOT APPLIED
-- ---------------------------------
-- Raised by the council gate's **bug_historian** seat on 272's review round
-- (correlation `f4a4628f-3b90-4054-a875-f2cf72b83e72`, verdict REVISE). The
-- objection, verbatim in substance: opting one consumer in while leaving the sibling
-- on the terminal path is *"the exact shape already in this platform's own index:
-- 'One call site of a shared judgement gets the rigorous fix; the sibling stays
-- heuristic.'"* Its `missing` item named the real gap — not that the sibling was
-- left, but that nothing would ever remind anyone:
--
--   "Whether fix-proposer's owning lane has an open ticket or plan tracking the same
--    migration — if not, bugs_open/099's failure mode persists there indefinitely
--    with no visible trigger to revisit it."
--
-- That is correct, and this file is the answer to it. `fix-proposer` belongs to
-- another lane; changing another lane's live agent from outside is the collision
-- CLAUDE.md warns about, and 099's own candidate 1 records that exact reasoning
-- (then reversed it, deliberately and with a stated reason). So the work is done and
-- left one command away, rather than either applied unilaterally or forgotten.
--
-- **To apply:** confirm the chassis carries the Go half (see ORDERING in 272), then
-- run this file. It is guarded and verified the same way 272 is.
--
-- WHAT DIFFERS FROM 272
-- ---------------------
--   * `persist_plan.next_step` here is `select_panel`, not `review_editquality`.
--   * There is no `spec_row` on this agent. The repair prompt uses
--     `{{.diagnosis_row.conclusion}}`, which is what `propose` and `repropose`
--     already read (`propose_inputs`: diagnosis_row, last_bundle).
--   * `max_tokens` is 8000, matching this agent's own `propose` step — NOT 32000.
--     A fix plan is a single ≤8-edit plan; a feature plan is up to 6 staged ones.
--     Sizing the repair call to the step that produces the same artefact is the
--     point, and this agent's propose demonstrably produces plans at 8000.
--   * `fix-proposer` has NO root `ai_service` block either (verified 2026-07-30),
--     so the step-level block is the whole effective config — and per
--     `resolveAIServiceConfig` (ai_actions.go:40-96) a step block wins per key over
--     a root one anyway.
--
-- ROLLBACK
-- --------
-- Same correction as 272's, and for the same reason: `snapshot_agent(text,text)`
-- writes to **agent_definitions_backup**, NOT an `is_snapshot` row, and that table
-- keeps the SOURCE row's `created_at`, so a restore MUST order by
-- **`snapshot_taken_at`**. See LANDMINES.md.
--
--   UPDATE agent_definitions a
--      SET default_config = b.default_config, updated_at = now()
--     FROM (SELECT default_config FROM agent_definitions_backup
--            WHERE type = 'fix-proposer'
--              AND snapshot_reason = '273_fix_proposer_plan_repair_loop.sql: pre-update'
--            ORDER BY snapshot_taken_at DESC LIMIT 1) b
--    WHERE a.type = 'fix-proposer' AND a.deleted_at IS NULL
--      AND COALESCE(a.is_snapshot, false) = false;
--
-- Or, sufficient alone, put persist_plan back on its old next_step:
--   UPDATE agent_definitions
--      SET default_config = jsonb_set(default_config,
--            '{workflow,steps,persist_plan,next_step}', '"select_panel"')
--    WHERE type='fix-proposer' AND deleted_at IS NULL
--      AND COALESCE(is_snapshot,false)=false;

BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'fix-proposer'
           AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
           AND (default_config->'workflow'->'steps') ? 'check_plan_valid'
    ) THEN
        RAISE EXCEPTION '273: already applied — fix-proposer already has a check_plan_valid step';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'fix-proposer'
           AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
           AND default_config->'workflow'->'steps'->'persist_plan'->>'next_step' = 'select_panel'
    ) THEN
        RAISE EXCEPTION '273: persist_plan.next_step is not select_panel — the step graph has changed under this migration; re-read it before applying';
    END IF;
END $$;

SELECT snapshot_agent('fix-proposer',
    '273_fix_proposer_plan_repair_loop.sql: pre-update');

-- 1. Opt persist_plan in, and route it through the router.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,persist_plan,config,repair_step}', '"repair_plan"'
           ),
           '{workflow,steps,persist_plan,config,max_repair_attempts}', '1'
         ),
         '{workflow,steps,persist_plan,next_step}', '"check_plan_valid"'
       ),
       updated_at = now()
 WHERE type = 'fix-proposer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- 2. The router. plan_valid == true is what reaches the reviewers; an unresolvable
--    field compares false and goes to repair, never onward to a council.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,check_plan_valid}',
         jsonb_build_object(
           'action', 'conditional_branch',
           'description', 'Plan router: structurally valid → the panel; refused → repair_plan. '
                       || 'Oriented so an unresolvable field routes to repair, never to a council.',
           'config', jsonb_build_object(
             'condition', 'plan_persisted.plan_valid == true',
             'then_step', 'select_panel',
             'else_step', 'repair_plan'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'fix-proposer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- 3. The repair step. Structural problems, NOT council objections.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,repair_plan}',
         jsonb_build_object(
           'action', 'execute_llm_prompt',
           'output_field', 'proposal',
           'next_step', 'persist_plan',
           'description', 'Repair a plan that failed STRUCTURAL validation (bugs_open/099 cand 2). '
                       || 'Bounded by persist_plan.config.max_repair_attempts, counted from the '
                       || 'durable refusal notes, so this cannot loop.',
           'config', jsonb_build_object(
             'error_step', 'complete_refused',
             'output_format', 'json',
             'temperature', 0.0,
             'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted'),
             'ai_service', jsonb_build_object(
               'provider', 'anthropic',
               'model', 'claude-sonnet-5',
               'max_tokens', 8000,
               'api_key_env_var', 'ANTHROPIC_API_KEY'
             ),
             'prompt_template',
               E'# PROMPT — REPAIR the plan''s STRUCTURE\n\n'
            || 'Your plan was NOT reviewed and NOT rejected on its merits. It failed the '
            || 'loop''s STRUCTURAL validator before any reviewer saw it, so it was never '
            || 'persisted. The plan itself may be entirely sound.\n\n'
            || E'## The exact problems the validator reported\n{{.plan_persisted.validation_problems_text}}\n\n'
            || E'## Your plan, verbatim\n{{.plan_persisted.rejected_plan_json}}\n\n'
            || E'## What to do\n'
            || E'Return the SAME plan with ONLY those problems fixed. This is a structural '
            || E'repair, not a redesign:\n\n'
            || E'- Keep the intent and the edits. Do not drop scope, do not add scope.\n'
            || E'- Fix each reported problem literally. "X appears in more than one edit" '
            || E'means COMBINE those edits into ONE whose sketch describes every change to '
            || E'that file — not delete one of them.\n'
            || E'- Change nothing the validator did not complain about.\n'
            || E'- Every edit must still CHANGE something: no audits, no comment-only edits, '
            || E'no "no change required".\n\n'
            || E'## The confirmed diagnosis this plan serves, for context only\n{{.diagnosis_row.conclusion}}\n\n'
            || E'## Output — ONLY the plan JSON, same schema as before. No prose, no fences.'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'fix-proposer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

DO $$
DECLARE
    steps jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps' INTO steps
      FROM agent_definitions
     WHERE type = 'fix-proposer'
       AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

    IF steps->'persist_plan'->>'next_step' IS DISTINCT FROM 'check_plan_valid' THEN
        RAISE EXCEPTION '273: persist_plan.next_step is %', steps->'persist_plan'->>'next_step';
    END IF;
    IF steps->'persist_plan'->'config'->>'repair_step' IS DISTINCT FROM 'repair_plan' THEN
        RAISE EXCEPTION '273: repair_step is %', steps->'persist_plan'->'config'->>'repair_step';
    END IF;
    IF steps->'check_plan_valid'->'config'->>'condition' IS DISTINCT FROM 'plan_persisted.plan_valid == true' THEN
        RAISE EXCEPTION '273: condition is %', steps->'check_plan_valid'->'config'->>'condition';
    END IF;
    IF NOT (steps ? (steps->'check_plan_valid'->'config'->>'then_step')) THEN
        RAISE EXCEPTION '273: then_step names a step that does not exist';
    END IF;
    IF NOT (steps ? (steps->'check_plan_valid'->'config'->>'else_step')) THEN
        RAISE EXCEPTION '273: else_step names a step that does not exist';
    END IF;
    IF steps->'repair_plan'->>'next_step' IS DISTINCT FROM 'persist_plan' THEN
        RAISE EXCEPTION '273: repair_plan must return to persist_plan';
    END IF;
    IF NOT (steps->'repair_plan'->'config'->'ai_service' ?& array['provider','model','max_tokens','api_key_env_var']) THEN
        RAISE EXCEPTION '273: repair_plan ai_service is incomplete';
    END IF;
    IF (steps->'repair_plan'->'config'->'ai_service'->>'max_tokens')::int
       IS DISTINCT FROM (steps->'propose'->'config'->'ai_service'->>'max_tokens')::int THEN
        RAISE WARNING '273: repair_plan max_tokens (%) differs from propose (%) — intentional?',
            steps->'repair_plan'->'config'->'ai_service'->>'max_tokens',
            steps->'propose'->'config'->'ai_service'->>'max_tokens';
    END IF;

    RAISE NOTICE '273: fix-proposer plan repair loop wired — persist_plan -> check_plan_valid -> (select_panel | repair_plan -> persist_plan), capped at 1 repair round';
END $$;

COMMIT;
