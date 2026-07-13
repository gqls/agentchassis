-- 145_tool_acceptance_agent.sql — the Tier-4 acceptance orchestrator.
-- DB-only INSERT of a NEW agent type (no snapshot: nothing pre-existing to
-- restore; idempotent via WHERE NOT EXISTS). SAFE TO APPLY BEFORE the chassis
-- carrying its two actions deploys — nothing emits acceptance_run items yet
-- and nothing triggers the agent until 087 fires; DO NOT trigger until the
-- image with request_browser_run/judge_acceptance_results is live (an unknown
-- action fails the workflow).
--
-- WHAT (flow pinned in PLAN_tool_acceptance_runner; Stage 6):
--   ensure_site_record → load_docs (load_doc_context: current PLAN +
--   criteria_json for ('tool', spec.function)) → request_run
--   (request_browser_run: resolves the deployed page URL from pages, sends
--   run_checks to the browser-runner adapter, AWAITS; no-op skips when the
--   PLAN has no criteria — Tier 2 owns needs_criteria) → judge
--   (judge_acceptance_results: all pass → acceptance-run note; any fail →
--   acceptance-fail note + ONE improve_tool item carrying the criteria as
--   acceptance_test) → complete.
--
-- Workflows flat, complexity in Go: URL resolution, the no-criteria guard,
-- response-shape fallbacks, note/item writes all live in the two actions.
--
-- Handler contract (003): dispatchable as handler_agent for item_type
-- 'acceptance_run' (spec.function at input_data.spec.function); manual
-- triggers must ALSO send {"spec":{"function":...}} — one path, one contract.
-- Trigger points (runbook): after tool creation/recreation deploys; after any
-- improve_tool completes; periodic sweeps. None wired yet — 087 is manual.

BEGIN;

INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status,
   input_contract, output_contract, default_config)
SELECT
  'tool-acceptance-agent',
  'Tool Acceptance Agent (Tier 4)',
  'Drives a deployed tool in headless Chromium (browser-runner adapter) against the acceptance criteria in its travelling PLAN, then records the verdict: acceptance-run note on pass; acceptance-fail note plus an improve_tool item (criteria as acceptance_test) on failure. The tier that turns "deployed" into "works".',
  'tools',
  'analyst',
  'active',
  '{"required": ["site_id", "domain"], "optional": ["spec", "function", "page_name"]}'::jsonb,
  '{"acceptance_verdict": {"all_passed": "bool", "passed": "int", "failed": "int", "failing_checks": "[]", "improve_tool_created": "bool"}}'::jsonb,
  '{
    "workflow": {
      "start_step": "ensure_site_record",
      "processing_mode": "orchestrator",
      "timeout_seconds": 600,
      "steps": {
        "ensure_site_record": {
          "action": "ensure_site_record",
          "config": { "store_brief_in_content_data": false },
          "next_step": "load_docs",
          "description": "Load site record for context",
          "output_field": "site_record"
        },
        "load_docs": {
          "action": "load_doc_context",
          "config": {
            "subject_type": "tool",
            "subject_key_field": "input_data.spec.function",
            "error_step": "complete_error"
          },
          "next_step": "request_run",
          "description": "Current PLAN + latest NOTES + criteria_json for the tool",
          "output_field": "doc_context"
        },
        "request_run": {
          "action": "request_browser_run",
          "config": {
            "function_field": "input_data.spec.function",
            "criteria_field": "doc_context.criteria_json",
            "site_id_field": "site_record.site_id",
            "domain_field": "site_record.domain",
            "error_step": "complete_error"
          },
          "next_step": "judge",
          "description": "Resolve the deployed URL, send run_checks to the browser-runner adapter, AWAIT. No-op skip (no await) when the PLAN has no criteria.",
          "output_field": "browser_run"
        },
        "judge": {
          "action": "judge_acceptance_results",
          "config": {
            "results_field": "browser_run",
            "function_field": "input_data.spec.function",
            "criteria_field": "doc_context.criteria_json",
            "site_id_field": "site_record.site_id",
            "error_step": "complete_error"
          },
          "next_step": "complete",
          "description": "acceptance-run note on pass; acceptance-fail note + improve_tool item on failure",
          "output_field": "acceptance_verdict"
        },
        "complete": {
          "action": "complete_workflow",
          "config": { "multiple_output_fields": ["acceptance_verdict"] },
          "description": "Acceptance cycle complete"
        },
        "complete_error": {
          "action": "complete_workflow",
          "config": {
            "multiple_output_fields": ["acceptance_verdict", "browser_run"],
            "success_message": "Acceptance run completed with errors"
          },
          "description": "Acceptance run completed with errors"
        }
      }
    }
  }'::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL
);

-- Pipeline note (runbook §3: workflow-altering migrations leave number/what/why).
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 145: tool-acceptance-agent created (Stage 6 — Tier 4 goes self-driving)
Observed: the browser-runner adapter is live and smoke-proven, but nothing orchestrates criteria → browser → verdict; acceptance runs were hand-produced kcat requests.
Root cause: not-applicable (new capability).
Fix: tool-acceptance-agent — ensure_site → load_doc_context → request_browser_run (awaits the adapter; skips without faking when the PLAN has no criteria) → judge_acceptance_results (acceptance-run note on pass; acceptance-fail note + improve_tool item carrying the criteria as acceptance_test on failure).
Verified: unit tests on the response-shape fallbacks and verdicts; DO NOT trigger until the chassis image carrying request_browser_run/judge_acceptance_results deploys (unknown action fails the workflow). Trigger script: 087.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-acceptance-agent' AND deleted_at IS NULL
      AND default_config #>> '{workflow,start_step}' = 'ensure_site_record'
      AND default_config #>> '{workflow,steps,request_run,action}' = 'request_browser_run'
      AND default_config #>> '{workflow,steps,judge,action}' = 'judge_acceptance_results'
      AND default_config #>> '{workflow,steps,load_docs,config,error_step}' = 'complete_error'
      AND NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v) WHERE v ? 'error_step')
      AND input_contract->'optional' ? 'spec';
    IF n <> 1 THEN RAISE EXCEPTION '145: tool-acceptance-agent shape wrong (found %)', n; END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT type, status, default_config #>> '{workflow,start_step}' AS start
--   FROM agent_definitions WHERE type='tool-acceptance-agent' AND deleted_at IS NULL;
-- Rollback: UPDATE agent_definitions SET deleted_at = now()
--   WHERE type='tool-acceptance-agent' AND deleted_at IS NULL;
