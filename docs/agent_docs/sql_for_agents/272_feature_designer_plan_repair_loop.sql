-- 272_feature_designer_plan_repair_loop.sql
--
-- bugs_open/099 — FIX CANDIDATE 2, the durable one. Candidate 1 (migration
-- `222_feature_designer_one_edit_per_file_per_stage.sql`) told the designer ONE of
-- the validator's rules. This closes the class instead.
--
-- WHY
-- ---
-- `diagnose_persist_fix_plan` returned a bare Go error whenever a plan failed
-- structural validation. That fails the step, `error_step` routes to
-- `complete_refused`, and a COMPLETED, GOOD design is discarded — with
-- `orchestration_states.error` NULL and the reason only in
-- `collected_data->>'__step_error'`, so a dashboard keyed on `error` reports the
-- run as clean. The validator has a dozen rules; candidate 1 fixed one of them by
-- stating it in the prompt, which has to be repeated per rule, per agent, and
-- drifts the moment the validator changes.
--
-- The Go half (shipped separately, and REQUIRED BEFORE THIS FILE — see ORDERING)
-- makes a structural refusal recoverable: the rejected plan and the exact problems
-- are persisted as a durable artefact and returned as a RESULT carrying
-- `plan_valid: false` and `should_repair_plan: true`, for a bounded number of
-- rounds, instead of the step failing. Nothing invalid is ever persisted as a
-- `fix_plan` on either path — the gate is not lowered, only made recoverable.
--
-- WHAT THIS FILE DOES — feature-designer only
-- -------------------------------------------
--   persist_plan.config  += repair_step='repair_plan', max_repair_attempts=1
--   persist_plan.next_step: review_editquality -> check_plan_valid
--   + check_plan_valid   conditional: plan_valid == true ? review_editquality : repair_plan
--   + repair_plan        execute_llm_prompt -> persist_plan
--
-- `fix-proposer` and `council-gate` (which uses the same action at step
-- `persist_submission`) are DELIBERATELY NOT CHANGED. With `repair_step` unset the
-- action behaves exactly as it did before, so their guarantee is unchanged. The
-- council gate is the gate this very change is reviewed by; coupling it to the
-- change it reviews buys nothing.
--
-- WHY THE CONDITION IS ORIENTED THIS WAY
-- --------------------------------------
-- `plan_valid == true` gates the path to the REVIEWERS. compareValues(nil,"true")
-- is false (conditional_branch_action.go:479-493), so if the field ever fails to
-- resolve the run goes to repair — visibly stalled — rather than carrying a plan
-- nothing validated to a council. The safe failure direction is the one that does
-- not reach a reviewer.
--
-- WHY NOT ROUTE INTO THE EXISTING `repropose`
-- -------------------------------------------
-- bugs_open/099's candidate 2 says to route into `repropose`, "which exists".
-- Checked, and it does not work as written: `persist_plan` runs BEFORE any council,
-- so on a first-pass refusal `repropose`'s prompt renders `{{.council_reviews.body}}`
-- and `{{.check_results.results_text}}` with nothing behind them, and it frames a
-- structural problem as a council objection. Hence a dedicated `repair_plan`.
-- The correction is recorded in the bug file too.
--
-- ORDERING — THE IMAGE MUST LAND FIRST
-- ------------------------------------
-- DB config is live immediately; Go is inert until a roll. Applied against a chassis
-- that does NOT yet carry the Go half, `persist_plan` still returns an error on a
-- bad plan, `error_step` still routes to `complete_refused`, and `check_plan_valid`
-- is simply never reached. So this file is INERT-BUT-HARMLESS early rather than
-- broken — but it does nothing until the image ships. Verify the binary first:
--   kubectl exec -n ai-persona-system <chassis pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c plan_validation_refusal'
--
-- IDEMPOTENT + NON-CLOBBERING
-- ---------------------------
-- Guarded: RAISEs if `check_plan_valid` already exists. Each step is written with
-- jsonb_set at its own path, so a concurrent edit to another step survives. The
-- verification block RAISEs rather than reporting a success it did not achieve.
--
-- ROLLBACK
-- --------
-- CORRECTED 2026-07-30 after the council's debug_historian seat objected that the
-- backup discipline here was unconfirmed against a documented double-overload trap.
-- It was right, and the first version of this block was WRONG. What was checked:
--
--   * snapshot_agent has TWO overloads — snapshot_agent(text) and
--     snapshot_agent(text, text). This file calls the 2-arg form, which writes to
--     **agent_definitions_backup**, NOT to an is_snapshot row in agent_definitions.
--     The original instruction here said "the newest is_snapshot row", which would
--     have sent a rollback looking in the wrong table entirely.
--   * That backup copies id/created_at/updated_at VERBATIM from the source row, so
--     ordering by created_at does NOT find the latest snapshot. Measured: all three
--     feature-designer backups share source created_at 2026-07-17 18:06:05, so
--     created_at is a three-way tie. Only **snapshot_taken_at** discriminates.
--
--   Working restore (verified to return the right row, 2026-07-30):
--     SELECT snapshot_taken_at, snapshot_reason FROM agent_definitions_backup
--      WHERE type = 'feature-designer' ORDER BY snapshot_taken_at DESC LIMIT 5;
--
--     UPDATE agent_definitions a
--        SET default_config = b.default_config, updated_at = now()
--       FROM (SELECT default_config FROM agent_definitions_backup
--              WHERE type = 'feature-designer'
--                AND snapshot_reason = '272_feature_designer_plan_repair_loop.sql: pre-update'
--              ORDER BY snapshot_taken_at DESC LIMIT 1) b
--      WHERE a.type = 'feature-designer' AND a.deleted_at IS NULL
--        AND COALESCE(a.is_snapshot, false) = false;
--
--   NOTE for whoever maintains migration 222 (candidate 1 of the same bug): its
--   rollback block carries the IDENTICAL wrong instruction ("the newest is_snapshot
--   row"). Not edited here — it is another lane's applied migration — but the trap
--   is recorded in LANDMINES.md so either file's reader is warned.
--
--   Or, sufficient on its own and needing no snapshot at all, put persist_plan back
--   on its old next_step:
--     UPDATE agent_definitions
--        SET default_config = jsonb_set(default_config,
--              '{workflow,steps,persist_plan,next_step}', '"review_editquality"')
--      WHERE type='feature-designer' AND deleted_at IS NULL
--        AND COALESCE(is_snapshot,false)=false;
--   -- which strands the two new steps unreachable, restoring the old behaviour.

BEGIN;

-- ---------------------------------------------------------------------------
-- Guard: refuse a second application.
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'feature-designer'
           AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
           AND (default_config->'workflow'->'steps') ? 'check_plan_valid'
    ) THEN
        RAISE EXCEPTION '272: already applied — feature-designer already has a check_plan_valid step';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'feature-designer'
           AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
           AND default_config->'workflow'->'steps'->'persist_plan'->>'next_step' = 'review_editquality'
    ) THEN
        RAISE EXCEPTION '272: persist_plan.next_step is not review_editquality — the step graph has changed under this migration; re-read it before applying';
    END IF;
END $$;

SELECT snapshot_agent('feature-designer',
    '272_feature_designer_plan_repair_loop.sql: pre-update');

-- ---------------------------------------------------------------------------
-- 1. Opt persist_plan into the repair loop, and route it through the router.
-- ---------------------------------------------------------------------------
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
 WHERE type = 'feature-designer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- Keep the step's own description honest: it no longer always fails the run.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,persist_plan,description}',
         to_jsonb('Structural validation (staged-v1 path: stage shape, cross-stage file '
               || 'discipline, seed/checklist contract) + write to diagnosis_artifacts '
               || '(kind=fix_plan, plan_format staged-v1 — D1). A failed validation is '
               || 'RECOVERABLE (bugs_open/099 cand 2): the rejected plan and its problems '
               || 'are recorded as an iteration_note and routed to repair_plan for '
               || 'max_repair_attempts rounds, then the run fails as before. Nothing '
               || 'invalid is ever persisted as a fix_plan.'::text)
       ),
       updated_at = now()
 WHERE type = 'feature-designer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- ---------------------------------------------------------------------------
-- 2. The router. plan_valid == true is what reaches the reviewers.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,check_plan_valid}',
         jsonb_build_object(
           -- 'conditional_branch', NOT 'conditional'. The sibling check_* steps in
           -- this agent all use 'conditional', but the registry marks it
           -- Deprecated with DeprecatedBy 'conditional_branch' (registry.go:65-78);
           -- both resolve to the same ConditionalBranchAction handler, so this is
           -- identical in behaviour and does not add to the deprecation debt.
           'action', 'conditional_branch',
           'description', 'Plan router: structurally valid → the reviewers; refused → repair_plan. '
                       || 'Oriented so an unresolvable field routes to repair, never to a council.',
           'config', jsonb_build_object(
             'condition', 'plan_persisted.plan_valid == true',
             'then_step', 'review_editquality',
             'else_step', 'repair_plan'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'feature-designer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- ---------------------------------------------------------------------------
-- 3. The repair step. Names the STRUCTURAL problems — not council objections —
--    and asks for the same design back, minimally corrected.
-- ---------------------------------------------------------------------------
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
             'input_fields', jsonb_build_array('spec_row', 'plan_persisted'),
             'ai_service', jsonb_build_object(
               'provider', 'anthropic',
               'model', 'claude-sonnet-5',
               'max_tokens', 32000,
               'api_key_env_var', 'ANTHROPIC_API_KEY'
             ),
             'prompt_template',
               E'# PROMPT — REPAIR the plan''s STRUCTURE\n\n'
            || 'Your plan was NOT reviewed and NOT rejected on its merits. It failed the '
            || 'loop''s STRUCTURAL validator before any reviewer saw it, so it was never '
            || 'persisted. The design itself may be entirely sound.\n\n'
            || E'## The exact problems the validator reported\n{{.plan_persisted.validation_problems_text}}\n\n'
            || E'## Your plan, verbatim\n{{.plan_persisted.rejected_plan_json}}\n\n'
            || E'## What to do\n'
            || E'Return the SAME plan with ONLY those problems fixed. This is a structural '
            || E'repair, not a redesign:\n\n'
            || E'- Keep the intent, the stages and the edits. Do not drop scope, do not add '
            || E'scope, do not reorder work that had a reason to be ordered.\n'
            || E'- Fix each reported problem literally. "X appears in more than one edit of '
            || E'this stage" means COMBINE those edits into ONE whose sketch describes every '
            || E'change to that file, or move the later change to a later stage — not delete '
            || E'one of them.\n'
            || E'- Change nothing the validator did not complain about.\n'
            || E'- Every edit must still CHANGE something: no audits, no comment-only edits, '
            || E'no "no change required".\n\n'
            -- spec_text, NOT body: the design step's own prompt renders
            -- {{.spec_row.work_item_id}}, {{.spec_row.summary}} and
            -- {{.spec_row.spec_text}}, and there is no `body` field. A wrong path
            -- here renders an EMPTY section and says nothing about it.
            || E'## The spec this plan serves, for context only\n{{.spec_row.summary}}\n\n{{.spec_row.spec_text}}\n\n'
            || E'## Output — ONLY the plan JSON, same schema as before. No prose, no fences.'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'feature-designer'
   AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

-- ---------------------------------------------------------------------------
-- Verify. Every assertion, or RAISE — a partly-applied step graph is worse than
-- an unapplied one, because the run reaches a step that is not there.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    steps        jsonb;
    persist_next text;
    repair       text;
    attempts     text;
    cond         text;
BEGIN
    SELECT default_config->'workflow'->'steps'
      INTO steps
      FROM agent_definitions
     WHERE type = 'feature-designer'
       AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

    persist_next := steps->'persist_plan'->>'next_step';
    repair       := steps->'persist_plan'->'config'->>'repair_step';
    attempts     := steps->'persist_plan'->'config'->>'max_repair_attempts';
    cond         := steps->'check_plan_valid'->'config'->>'condition';

    IF persist_next IS DISTINCT FROM 'check_plan_valid' THEN
        RAISE EXCEPTION '272: persist_plan.next_step is % — expected check_plan_valid', persist_next;
    END IF;
    IF repair IS DISTINCT FROM 'repair_plan' THEN
        RAISE EXCEPTION '272: persist_plan.config.repair_step is % — expected repair_plan', repair;
    END IF;
    IF attempts IS DISTINCT FROM '1' THEN
        RAISE EXCEPTION '272: persist_plan.config.max_repair_attempts is % — expected 1', attempts;
    END IF;
    IF cond IS DISTINCT FROM 'plan_persisted.plan_valid == true' THEN
        RAISE EXCEPTION '272: check_plan_valid condition is % — expected plan_persisted.plan_valid == true', cond;
    END IF;

    -- Every step named as a destination must exist, or a refused plan walks off
    -- the end of the graph instead of being repaired.
    IF NOT (steps ? 'repair_plan') THEN
        RAISE EXCEPTION '272: repair_plan step is missing';
    END IF;
    IF NOT (steps ? (steps->'check_plan_valid'->'config'->>'then_step')) THEN
        RAISE EXCEPTION '272: check_plan_valid.then_step names a step that does not exist';
    END IF;
    IF NOT (steps ? (steps->'check_plan_valid'->'config'->>'else_step')) THEN
        RAISE EXCEPTION '272: check_plan_valid.else_step names a step that does not exist';
    END IF;
    IF steps->'repair_plan'->>'next_step' IS DISTINCT FROM 'persist_plan' THEN
        RAISE EXCEPTION '272: repair_plan must return to persist_plan, got %',
            steps->'repair_plan'->>'next_step';
    END IF;

    -- ai_service completeness, asked for by the council's llm_reliability seat.
    --
    -- Its objection was that a ROOT ai_service block might shadow this step-level
    -- one, citing MDL-039 ("a step-level block is completely dead when a root block
    -- exists"). Measured, and it does not hold, for two independent reasons:
    --   1. feature-designer has NO root ai_service block — its only root key is
    --      'workflow'. There is nothing to shadow.
    --   2. That behaviour was bugs_open/009 and is FIXED. resolveAIServiceConfig
    --      (ai_actions.go:40-96) is now a per-key overlay, root -> step -> runtime,
    --      "later wins PER KEY", and its own comment says it "replaces
    --      first-found-wins, under which the ENTIRE step block was dead config
    --      whenever a root block existed". So a step block now WINS per key.
    -- MDL-039 describes pre-fix behaviour and should be re-dated.
    --
    -- The residual risk is therefore the opposite one — this step's block missing a
    -- key that no root block supplies either, which under an overlay yields an
    -- incomplete effective config. That is what this asserts, and it keeps holding
    -- if someone later adds a root block.
    IF NOT (steps->'repair_plan'->'config'->'ai_service' ?& array['provider','model','max_tokens','api_key_env_var'] ) THEN
        RAISE EXCEPTION '272: repair_plan ai_service is incomplete — needs provider, model, max_tokens, api_key_env_var (got %)',
            steps->'repair_plan'->'config'->'ai_service';
    END IF;

    -- max_tokens is sized to the DESIGN step's, deliberately: repair_plan emits a
    -- plan of the same shape and size, on the same model, and design demonstrably
    -- produces staged plans today. The llm_reliability seat also asked for a
    -- 'thinking' key; there is none to set — 0 live steps fleet-wide carry
    -- ai_service.thinking and execute_llm_prompt does not read that key (the
    -- extended-thinking knob it reads is budget_tokens, ai_actions.go:359-360), so
    -- setting it would be dead config of exactly the class bugs_open/134 is about.
    IF (steps->'repair_plan'->'config'->'ai_service'->>'max_tokens')::int
       IS DISTINCT FROM (steps->'design'->'config'->'ai_service'->>'max_tokens')::int THEN
        RAISE WARNING '272: repair_plan max_tokens (%) differs from design (%) — intentional?',
            steps->'repair_plan'->'config'->'ai_service'->>'max_tokens',
            steps->'design'->'config'->'ai_service'->>'max_tokens';
    END IF;

    RAISE NOTICE '272: feature-designer plan repair loop wired — persist_plan -> check_plan_valid -> (review_editquality | repair_plan -> persist_plan), capped at 1 repair round';
END $$;

COMMIT;
