-- ============================================================================
-- MIGRATION: seed the diagnosis-loop agents
--   diagnose-orchestrator  (wrapper: spawn → call → complete)
--   diagnose-agent         (worker: the read-only loop)
-- ============================================================================
-- DRAFT for the agent-chassis repo. SQL written against the REAL agent_definitions
-- schema (\d verified): type+version UNIQUE; category NOT NULL varchar(50);
-- agent_category CHECK in {strategist,executor,analyst,integrator,coordinator,
-- specialist}; status CHECK in {active,experimental,deprecated,demo,template};
-- workflow lives in orchestrator_workflow (jsonb). Defaults exist for topics,
-- resources, health_config — omitted here so the table defaults apply.
--
-- The wrapper pattern is the dev guide's canonical med-export-orchestrator:
-- the diagnose loop does SUBSTANTIVE in-chassis work (LLM verdict calls,
-- multiple iterations, minutes), so per the guide's "does this agent need a
-- wrapper?" test it MUST be wrapped — the orchestrator spawns a dedicated Job
-- pod for diagnose-agent, which runs the loop and terminates. Clean logs, one
-- correlation per pod, no chassis-pod starvation.
--
-- Idempotency: ON CONFLICT (type, version) DO NOTHING — re-running won't dupe.
-- New status is 'experimental' (the table default + the CHECK) until the
-- real-bug evaluation gate passes; promote to 'active' after.

BEGIN;

-- ── 1. diagnose-agent (the worker: the read-only loop) ──────────────────────
-- agent_category 'analyst' (it diagnoses; it does not change anything).
-- The workflow is the gather → verdict → step loop, expressed thinly: the loop
-- CONTROL (guards, re-scope, iteration) stays in the Go engine via diagnose_run,
-- NOT re-expressed as workflow conditionals (dev guide: keep complexity in Go).
-- diagnose_run internally calls the verdict via execute_llm_prompt; the gather
-- steps run before it. We keep the gather as explicit steps so each read is
-- visible in logs, then hand the assembled context to the engine.
INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status,
   capabilities, input_contract, output_contract, orchestrator_workflow,
   default_config)
VALUES
  ('diagnose-agent',
   'Diagnosis Loop Agent',
   'Read-only diagnosis loop: hypothesise, gather scoped evidence (code + runtime), cite-or-abstain verdict, re-scope by following evidence. Emits a diagnosis + evidence trail for a human; never changes code or triggers a run.',
   'diagnose',
   'analyst',
   'experimental',
   '["code_analysis","read_only_diagnosis"]'::jsonb,
   -- input_contract: what a caller must provide (the symptom + seed scope).
   '{"required":["symptom"],"optional":["seed_scope","runtime_site","runtime_page","site_id","correlation_id","owner","repo","ref"]}'::jsonb,
   '{"provides":["diagnosis","evidence_trail"]}'::jsonb,
   -- orchestrator_workflow: gather (analyse → lookup → runtime → assemble) then
   -- the engine loop (diagnose_run) then emit then complete.
   '{
     "workflow": {
       "start_step": "analyse_repo",
       "steps": {
         "analyse_repo": {
           "action": "request_repo_analysis",
           "config": {
             "owner_field": "input_data.owner",
             "repo_field": "input_data.repo",
             "ref_field": "input_data.ref",
             "language": "go"
           },
           "output_field": "repo_analysis",
           "next_step": "lookup_symbols"
         },
         "lookup_symbols": {
           "action": "lookup_code_symbols",
           "config": {
             "query_field": "input_data.symptom",
             "repo_field": "repo_analysis.repo",
             "top_k": 12
           },
           "output_field": "code_lookup",
           "next_step": "load_runtime"
         },
         "load_runtime": {
           "action": "diagnose_load_runtime",
           "config": {
             "site_id_field": "input_data.site_id",
             "correlation_id_field": "input_data.correlation_id",
             "domain_field": "input_data.runtime_site"
           },
           "output_field": "runtime",
           "next_step": "assemble_bundle"
         },
         "assemble_bundle": {
           "action": "diagnose_assemble_bundle",
           "config": {
             "scope_field": "input_data.seed_scope",
             "code_results_field": "code_lookup.code_results",
             "analysis_field": "repo_analysis",
             "repo_root_field": "repo_analysis.root",
             "runtime_field": "runtime.runtime_evidence"
           },
           "output_field": "bundle",
           "next_step": "run_loop"
         },
         "run_loop": {
           "action": "diagnose_run",
           "config": {
             "seed_hypothesis_field": "input_data.symptom",
             "seed_scope_field": "input_data.seed_scope",
             "bundle_field": "bundle.bundle",
             "analysis_field": "repo_analysis",
             "max_iterations": 5,
             "verdict_prompt_ref": "diagnose-verdict-v1"
           },
           "output_field": "diagnosis",
           "next_step": "emit"
         },
         "emit": {
           "action": "diagnose_emit",
           "config": { "diagnosis_field": "diagnosis" },
           "next_step": "complete"
         },
         "complete": {
           "action": "complete_workflow",
           "config": { "result_from": "diagnosis" }
         }
       }
     }
   }'::jsonb,
   '{}'::jsonb)
ON CONFLICT (type, version) DO NOTHING;

-- ── 2. diagnose-orchestrator (the wrapper: spawn → call → complete) ──────────
-- agent_category 'coordinator'. This is what a human/scheduler triggers via the
-- generic entry point; it spawns diagnose-agent as a dedicated Job pod and
-- forwards the result. Mirrors med-export-orchestrator exactly.
-- Optional input fields use the "?" suffix so a missing one doesn't fail the
-- whole call (ResolveInputMapping fails on any unresolved non-optional path).
INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status,
   capabilities, input_contract, output_contract, orchestrator_workflow,
   default_config)
VALUES
  ('diagnose-orchestrator',
   'Diagnosis Orchestrator',
   'Thin wrapper that spawns a dedicated diagnose-agent pod for one diagnosis and forwards the result. Keeps the loop''s substantive work off the shared chassis pods.',
   'diagnose',
   'coordinator',
   'experimental',
   '["orchestration"]'::jsonb,
   '{"required":["symptom"],"optional":["seed_scope","runtime_site","runtime_page","site_id","correlation_id","owner","repo","ref"]}'::jsonb,
   '{"provides":["diagnosis","evidence_trail"]}'::jsonb,
   '{
     "workflow": {
       "start_step": "spawn_diagnoser",
       "steps": {
         "spawn_diagnoser": {
           "action": "spawn_agent",
           "config": { "role": "diagnoser", "agent_type": "diagnose-agent" },
           "next_step": "call_diagnoser"
         },
         "call_diagnoser": {
           "action": "call_agent",
           "config": {
             "agent_type": "diagnose-agent",
             "target_role": "diagnoser",
             "input_mapping": {
               "symptom":         "input_data.symptom",
               "seed_scope?":     "input_data.seed_scope",
               "runtime_site?":   "input_data.runtime_site",
               "runtime_page?":   "input_data.runtime_page",
               "site_id?":        "input_data.site_id",
               "correlation_id?": "input_data.correlation_id",
               "owner?":          "input_data.owner",
               "repo?":           "input_data.repo",
               "ref?":            "input_data.ref"
             },
             "timeout_seconds": 600
           },
           "next_step": "complete"
         },
         "complete": {
           "action": "complete_workflow",
           "config": { "result_from": "diagnose-agent_result" }
         }
       }
     }
   }'::jsonb,
   '{}'::jsonb)
ON CONFLICT (type, version) DO NOTHING;

COMMIT;

-- ── Verify ──────────────────────────────────────────────────────────────────
-- SELECT type, category, agent_category, status, version,
--        orchestrator_workflow->'workflow'->>'start_step' AS start_step
-- FROM agent_definitions
-- WHERE type IN ('diagnose-agent','diagnose-orchestrator')
-- ORDER BY type;
--
-- Expect two rows, status 'experimental', start_step set.

-- ── OPEN QUESTIONS to confirm against the running chassis before relying on this
-- (flagged honestly — these depend on chassis internals I can't see from \d alone):
--   1. WHICH workflow column does the engine read? The schema has THREE:
--      orchestrator_workflow (jsonb), task_workflow (jsonb), orchestration_workflow
--      (json). I used orchestrator_workflow because the wrapper pattern is an
--      ORCHESTRATOR and the dev guide's examples are "orchestrator" agents — but
--      CONFIRM by checking an existing working row, e.g.:
--        SELECT type, (task_workflow IS NOT NULL) t, (orchestrator_workflow IS NOT NULL) o,
--               (orchestration_workflow IS NOT NULL) oo
--        FROM agent_definitions WHERE type IN ('med-export-orchestrator','intake-orchestrator','site-adoption-agent');
--      If the worker (diagnose-agent) should use task_workflow rather than
--      orchestrator_workflow, move its workflow JSON to that column.
--   2. The "workflow" top-level wrapper key: the dev guide example shows
--      {"workflow":{"start_step":...,"steps":{...}}}. CONFIRM the engine expects
--      that nesting (vs the steps at the top level) against the same existing row.
--   3. verdict_prompt_ref ("diagnose-verdict-v1"): how diagnose_run locates the
--      verdict prompt. If prompts live in their own table/store keyed by ref,
--      seed PROMPT_diagnosis_verdict.md there under that key; if the prompt is
--      passed inline, replace verdict_prompt_ref with the prompt text in config.
--   4. call_agent result key: I used "diagnose-agent_result" in the wrapper's
--      complete step. CONFIRM the actual key call_agent writes the child result
--      under (it may be role-based "diagnoser_result" or step-based). Check how
--      med-export-orchestrator's complete reads its child result.
-- NNN_move_diagnose_workflow_to_default_config.sql
--
-- Renumber NNN to the next number in your migration sequence (3-digit prefixes).
--
-- FIX: the seeded diagnose agents (diagnose-agent, diagnose-orchestrator, created
-- 2026-06-20) carry their workflow in `orchestration_workflow` (the `json` column),
-- but every WORKING agent (build-dispatch-loop, site-work-orchestrator,
-- domain-research-classifier, ...) carries it in `default_config` (jsonb). The
-- workflow loader reads default_config; a workflow left in orchestration_workflow
-- will not load. This moves it.
--
-- It copies the column across verbatim (no re-typing — preserves the exact seeded
-- workflow), nulls orchestration_workflow, and merges in the two top-level keys the
-- working agents have that the seeded workflow lacked: processing_mode +
-- timeout_seconds. (The seeded orchestration_workflow held ONLY {"workflow":{...}};
-- the reference agents put processing_mode/timeout_seconds at default_config top
-- level — e.g. site-work-orchestrator. Without processing_mode the agent may not
-- run as an orchestrator.)
--
-- Schema (verified from \d agent_definitions):
--   default_config jsonb NOT NULL DEFAULT '{}'::jsonb
--   orchestration_workflow json (nullable)
-- The cast orchestration_workflow::jsonb (json -> jsonb) is valid; `||` merges
-- top-level keys.
--
-- Idempotent + non-clobbering: runs only where orchestration_workflow IS NOT NULL
-- (so a second run is a no-op) and default_config = '{}' (so it never overwrites a
-- default_config that already holds config). default_config is NOT NULL, so the
-- guards also stop it ever writing NULL.

-- diagnose-agent (the worker that runs the loop)
UPDATE agent_definitions
SET default_config = orchestration_workflow::jsonb
      || jsonb_build_object('processing_mode', 'orchestrator', 'timeout_seconds', 1800),
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-agent'
  AND orchestration_workflow IS NOT NULL
  AND default_config = '{}'::jsonb;

-- diagnose-orchestrator (the thin spawn-and-forward wrapper)
UPDATE agent_definitions
SET default_config = orchestration_workflow::jsonb
      || jsonb_build_object('processing_mode', 'orchestrator', 'timeout_seconds', 600),
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-orchestrator'
  AND orchestration_workflow IS NOT NULL
  AND default_config = '{}'::jsonb;

-- Verify (expect has_workflow = t, orch_wf_null = t for both):
--   SELECT type,
--          (default_config ? 'workflow')          AS has_workflow,
--          (default_config ->> 'processing_mode')  AS processing_mode,
--          (orchestration_workflow IS NULL)        AS orch_wf_null
--   FROM agent_definitions
--   WHERE type IN ('diagnose-agent','diagnose-orchestrator');

-- FLAG (not fixed here — confirm, then a follow-up if needed):
--  1. ai_service: the moved default_config has NO ai_service block. If diagnose_run's
--     verdict needs agent-level model config (vs the verdict_prompt_ref 'diagnose-verdict-v1'
--     carrying it), add it — site-adoption-agent shows the shape:
--       default_config || '{"ai_service":{"model":"claude-sonnet-4-6","provider":"anthropic","api_key_env_var":"ANTHROPIC_API_KEY"}}'::jsonb
--  2. timeout alignment: diagnose-orchestrator's call_diagnoser step has timeout_seconds
--     600; the diagnose-agent loop (<=5 iterations, each a model verdict + gather) may
--     exceed that. If it times out, raise BOTH the orchestrator's call_diagnoser
--     timeout and the agent's timeout_seconds together.