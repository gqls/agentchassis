-- 643 — arm the write-time CTA label/destination audit (bugs_open/399).
--
-- WHAT IT ARMS. `audit_cta_label_agreement` on every save_page_sections step.
-- When on, SavePageSectionsAction asks datahelpers.JudgeCTALabel — the SAME
-- question check_misdirected_cta asks of deployed HTML — of each CTA pair it is
-- about to persist, and writes one CTA_LABEL_MISMATCH row to agent_error_log
-- when the copy names a different page than the destination, or names two pages
-- equally well (recorded as ambiguous, never acted on: RFC_047 §10).
--
-- It RECORDS. It does not refuse a save, does not rewrite a label and does not
-- repoint a url. Opt-in with the unsafe default OFF in code, per the owner
-- ruling of 2026-08-02.
--
-- WHY A CHECK AND NOT A BETTER PROMPT — the measurement that decided it. The
-- writer is ALREADY told: `cta_text`'s own llm_guidance in content_components
-- says "the link destination is already fixed: write this CTA text FOR that
-- destination", and migration 476 + 477 put the resolved title in the prompt at
-- runtime as "Destination (fixed): <title>. … never promise a different one."
-- [MEASURED 2026-08-26] 781 of 2,297 page-content-writer prompts over three days
-- carry that literal, and of the CTA pairs written since 477 applied,
-- **155 of 1,060 (14.6%) still contradict their destination**. Told twice, and
-- still wrong one time in seven. Prompt text is not a control.
--
-- ⚠ SIX STEPS, NOT TWO, AND THE COUNT IS THE POINT. The obvious arming is
-- page-build-handler and page-rerender — the build writer and the repair
-- writer. A recursive census on 2026-08-26 found save_page_sections on SIX live
-- agents, four of them inside a loop's sub_workflow where a top-level
-- `{workflow,steps,save_sections}` path cannot reach them:
--
--     page-build-handler        workflow.steps.save_sections
--     page-rerender             workflow.steps.save_sections
--     tool-recreation-handler   workflow.steps.save_sections
--     pageflow-builder          workflow.steps.build_pages_loop.sub_workflow.steps.save_sections
--     page-rebuild              workflow.steps.build_pages_loop.sub_workflow.steps.save_sections
--     site-work-orchestrator    workflow.steps.build_items_loop.sub_workflow.steps.save_sections
--
-- This matters MORE for an instrument than for a guard. A guard armed on half
-- its writers protects half the estate, which is visibly partial. An instrument
-- armed on half its writers reports a RATE that reads as fleet-wide and is
-- silently biased by whichever writers were missed — and the whole deliverable
-- here is the rate, not the row. So all six are armed, uniformly, and the
-- assertion below fails the migration if the census ever stops being six.
-- (tool-recreation-handler declares expects_no_sections_metadata and will
-- persist no sections, so the pass is inert there. It is armed anyway: a
-- hand-maintained subset of a mechanically-derivable set is the drift class
-- this estate keeps filing, and "inert" is cheaper to carry than "exempt".)
--
-- REPLAY-SAFE: jsonb_set is idempotent, and the guard below tolerates a re-run.

BEGIN;

-- Fail loudly if the population has changed since the census above. A migration
-- that silently arms 5 of 7 is how a biased rate ships looking complete.
DO $$
DECLARE
    n integer;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.action ? (@ == "save_page_sections")') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

    IF n <> 6 THEN
        RAISE EXCEPTION
            'save_page_sections step census is %, expected 6 (censused 2026-08-26). '
            'A step was added or removed: re-run the census in this file''s header and '
            'extend the UPDATEs below before arming, or the recorded mismatch RATE is '
            'silently biased by the writers this migration did not reach.', n;
    END IF;
END $$;

-- Top-level save_sections: page-build-handler, page-rerender, tool-recreation-handler.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_sections,config,audit_cta_label_agreement}', 'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Loop sub_workflows: pageflow-builder, page-rebuild (build_pages_loop).
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
        'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'
        ->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Loop sub_workflow: site-work-orchestrator (build_items_loop).
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
        'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'build_items_loop'->'config'->'sub_workflow'
        ->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- VERIFY as a DO/RAISE, never a bare SELECT: ON_ERROR_STOP does not abort a
-- COMMIT on a non-empty result set, so a SELECT-shaped verify block cannot stop
-- a bad migration (the RFC_006 lesson).
DO $$
DECLARE
    armed integer;
BEGIN
    SELECT count(*) INTO armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.audit_cta_label_agreement ? (@ == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

    IF armed <> 6 THEN
        RAISE EXCEPTION 'armed % save_page_sections steps, expected 6 — aborting', armed;
    END IF;
    RAISE NOTICE 'audit_cta_label_agreement armed on 6 save_page_sections steps';
END $$;

COMMIT;

-- POST-APPLY READING, and the demand control that makes a zero informative.
-- The pass is INERT until an image carrying cta_label_audit.go rolls: an older
-- binary reads an unknown config key and ignores it. So a zero here means
-- "binary not rolled yet", NOT "no mismatches" — do not record a pre-roll zero
-- as evidence about this migration.
--
--   SELECT count(*) FILTER (WHERE (context->>'contradicts')::int > 0) AS pages_with_contradictions,
--          sum((context->>'contradicts')::int)                        AS contradictions,
--          sum((context->>'ambiguous')::int)                          AS ambiguous,
--          count(DISTINCT agent_type)                                 AS producing_agents
--   FROM agent_error_log
--   WHERE error_code = 'CTA_LABEL_MISMATCH' AND occurred_at > now() - interval '24 hours';
--
-- producing_agents MUST reach at least 2 (page-build-handler AND page-rerender)
-- once both paths have run. One producer means the coverage claim above is
-- failing silently, which is the failure this file's census exists to prevent.
