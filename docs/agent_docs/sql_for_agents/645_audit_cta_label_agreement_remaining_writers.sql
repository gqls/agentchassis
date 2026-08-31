-- 645 (_HOLD) — arm the CTA label/destination audit on the REMAINING FOUR
--                save_page_sections steps (bugs_open/399, sibling of 643).
--
-- ⚠⚠ HELD, and held on TWO conditions, not one:
--   1. The binary that reads audit_cta_label_agreement has rolled — probe the
--      POD exactly as 643's header specifies, with a control in the same breath.
--   2. **643 HAS BEEN APPLIED AND HAS RUN FOR AT LEAST ONE FULL SAVE CYCLE ON
--      BOTH PRIMARY WRITERS**, with CTA_LABEL_MISMATCH rows observed from
--      page-build-handler AND page-rerender. That is the canary the guardian
--      seat asked for (corr e9bda035): a new per-page DB read landing on every
--      content-writing pipeline at once is the risk, and two pipelines first is
--      how it is bounded.
--
-- WHY THIS FILE EXISTS AT ALL, rather than 643 simply arming all six: the
-- instrument argument and the rollout argument point in opposite directions and
-- both are right.
--   * ARM EVERYTHING: an instrument armed on half its writers reports a RATE
--     that reads fleet-wide and is silently biased by whichever writers it
--     missed. The rate is this mechanism's entire deliverable.
--   * ARM SLOWLY: six pipelines simultaneously acquiring a new DB read is a
--     fleet-wide blast radius for one unproven code path.
-- Staging resolves it only if the FIRST reading waits for the second apply.
-- ⚠⚠ SO: DO NOT READ OR PUBLISH THE MISMATCH RATE UNTIL THIS FILE HAS APPLIED.
-- Between 643 and 645 the record is a smoke test, not a measurement. The
-- RUNBOOK repeats this beside the query, because that is where someone will be
-- standing when they are tempted.
--
-- THE FOUR, and why a top-level path cannot reach them (census 2026-08-26):
--     pageflow-builder        workflow.steps.build_pages_loop.sub_workflow.steps.save_sections
--     page-rebuild            workflow.steps.build_pages_loop.sub_workflow.steps.save_sections
--     site-work-orchestrator  workflow.steps.build_items_loop.sub_workflow.steps.save_sections
--     tool-recreation-handler workflow.steps.save_sections   (top level)
-- tool-recreation-handler declares expects_no_sections_metadata and will persist
-- no sections, so the pass is inert there. Armed anyway: a hand-maintained
-- subset of a mechanically-derivable set is the drift class this estate keeps
-- filing, and "inert" is cheaper to carry than "exempt".

BEGIN;

SELECT snapshot_agent('pageflow-builder',        'pre-update: 645 arm audit_cta_label_agreement (bugs_open/399)');
SELECT snapshot_agent('page-rebuild',            'pre-update: 645 arm audit_cta_label_agreement (bugs_open/399)');
SELECT snapshot_agent('site-work-orchestrator',  'pre-update: 645 arm audit_cta_label_agreement (bugs_open/399)');
SELECT snapshot_agent('tool-recreation-handler', 'pre-update: 645 arm audit_cta_label_agreement (bugs_open/399)');

-- Refuse to run ahead of the canary: 643 must have armed both primary writers.
DO $$
DECLARE
    primary_armed integer;
BEGIN
    SELECT count(*) INTO primary_armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.steps.save_sections ? (@.action == "save_page_sections" && @.config.audit_cta_label_agreement == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL
      AND a.type IN ('page-build-handler', 'page-rerender');

    IF primary_armed <> 2 THEN
        RAISE EXCEPTION
            '643 has not armed both primary writers (found %) — apply 643 first and let it run a '
            'full save cycle on each. Arming the remaining four ahead of that skips the canary the '
            'guardian seat asked for.', primary_armed;
    END IF;
END $$;

-- build_pages_loop: pageflow-builder, page-rebuild.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
        'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'
        ->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- build_items_loop: site-work-orchestrator.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,audit_cta_label_agreement}',
        'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'build_items_loop'->'config'->'sub_workflow'
        ->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Top level: tool-recreation-handler (643 deliberately took only the two primary writers).
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_sections,config,audit_cta_label_agreement}', 'true'::jsonb, true),
    updated_at = NOW()
WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND type = 'tool-recreation-handler'
  AND default_config->'workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- VERIFY: the key must sit on a step whose ACTION is save_page_sections, never
-- merely exist somewhere in the document. jsonb_set with create_missing writes a
-- dead key at a path nothing reads if a step name ever differs, and a verify
-- counting the key alone would pass on that dead key and report success
-- (editquality, corr e9bda035).
DO $$
DECLARE
    armed integer;
BEGIN
    SELECT count(*) INTO armed
    FROM agent_definitions a,
         LATERAL jsonb_path_query(a.default_config,
                 'strict $.**.steps.save_sections ? (@.action == "save_page_sections" && @.config.audit_cta_label_agreement == true)') x
    WHERE a.is_active AND COALESCE(a.is_snapshot, false) = false AND a.deleted_at IS NULL;

    IF armed <> 6 THEN
        RAISE EXCEPTION
            'armed % of 6 save_page_sections steps — aborting. Either a step name differs from the '
            '2026-08-26 census (so the key landed on a path nothing reads), or the population moved.', armed;
    END IF;
    RAISE NOTICE 'audit_cta_label_agreement now armed on all 6 save_page_sections steps — the rate is readable from here';
END $$;

COMMIT;
