-- ╔══════════════════════════════════════════════════════════════════════════╗
-- ║ SUPERSEDED — DO NOT APPLY.                                                  ║
-- ║ You already have diagnose-agent + diagnose-orchestrator seeded (2026-06-20),║
-- ║ which is the better design: the loop runs inside the tested diagnose_run    ║
-- ║ engine action (complexity in Go, thin workflow) rather than re-expressed as ║
-- ║ workflow steps + re-invocation. Use that pair. Fix its workflow COLUMN with  ║
-- ║ NNN_move_diagnose_workflow_to_default_config.sql (orchestration_workflow →   ║
-- ║ default_config). This file is kept only as a record of the re-invocation     ║
-- ║ design that was considered and dropped.                                      ║
-- ╚══════════════════════════════════════════════════════════════════════════╝
--
-- NNN_seed_diagnose_agents.sql
--
-- Renumber NNN to the next number in your migration sequence (3-digit prefixes,
-- e.g. 002_, 003_, 018_).
--
-- Seeds the `diagnostician` agent definition — the READ-ONLY diagnosis loop.
-- This file did not exist; it is written here from the guidelines (001 §"Don't
-- create subworkflows in SQL", §"Spawn before call", §5/§9) and modelled on the
-- live reference agents (build-dispatch-loop, site-work-orchestrator,
-- domain-research-classifier) you supplied. The workflow lives in
-- `default_config` (confirmed: those agents put their workflow there; the
-- task_workflow / orchestrator_workflow columns are NULL).
--
-- ── DESIGN: why one iteration per orchestration (not a workflow cycle) ────────
-- The diagnosis loop is a WHILE loop: re-scope until CONFIRMED / abstain / cap,
-- each iteration's verdict choosing the next scope. The `loop` action does NOT
-- fit — it is a FOR-EACH over a collection fixed at loop entry (001 Appendix C),
-- and the next scope is not known until the verdict runs. A workflow-internal
-- cycle (route -> back to assemble) would re-run a step name and relies on cyclic
-- next_step the engine is not shown to support. So, following build-dispatch-loop
-- ("process one unit per orchestration, re-invoke for the next; separate
-- orchestrations = clean logs") and the rule "spawn sub-agents with their own
-- workflows", each diagnose orchestration runs ONE iteration:
--   load_runtime -> request_analysis -> lookup_symbols -> assemble -> verdict
--   -> route -> [continue: spawn_next -> call_next -> emit] | [terminal: emit]
--   -> complete
-- On "continue" it spawns+calls a fresh diagnostician (a sub-agent of the SAME
-- type) for the next iteration, passing the revised hypothesis + next_scope +
-- data_requests + iteration+1 in input_data. The final verdict bubbles up.
--
-- Consequence: cross-iteration state travels in input_data (not collected_data).
-- That is why diagnose_load_runtime reads data_requests from input_data.data_requests
-- and diagnose_assemble_bundle's scope/hypothesis fields are pointed at input_data.*
-- below (overriding the actions' route.* defaults, which reflected an earlier
-- cycle design).
--
-- ── CONFIRM BEFORE DEPLOY (flagged — not verifiable from here) ────────────────
--  1. request_repo_analysis and lookup_code_symbols config keys: the keys below
--     are best-effort; check each action's ActionInputSpec and adjust. request_repo_analysis
--     is REQUIRED (assemble needs the analyser Output for ReadSymbolBody + repo_root);
--     lookup_symbols is only the first-iteration scope fallback when no seed_scope.
--  2. diagnose_route and diagnose_emit must exist + be registered (the 4-action set).
--     route here PRODUCES {continue, iteration} (output_field "route") and a PROVEN
--     `conditional` branches on route.continue — rather than relying on an
--     action-driven next_step override. emit picks the child's verdict (if we
--     re-invoked) else this iteration's verdict.
--  3. Nesting: spawn_next+call_next makes iteration N+1 a CHILD of N (≤ max deep).
--     If you prefer flat, non-nested iterations, replace spawn_next/call_next with
--     a fire to the generic entry point (system.agent.generic.requests, config.agent_type
--     'diagnostician') + complete — the path the scheduler/manual triggers use
--     (001 §"Every pod-running agent needs a parent"). Confirm an action exists for
--     that emit before switching.
--  4. Paste PROMPT_diagnosis_verdict.md into the verdict step's prompt_template
--     (JSON-escaped). The placeholder below is NOT the real prompt.
--  5. agent_category must be in the CHECK list (strategist|executor|analyst|
--     integrator|coordinator|specialist) — 'analyst' here. 'diagnose' is only in
--     the free-text `category` column. status in (active|experimental|deprecated|
--     demo|template) — 'experimental'.
--
-- Idempotent on (type, version): the unique constraint is (type, version). Re-running
-- would conflict; guard with ON CONFLICT or bump version for a revision.

INSERT INTO agent_definitions (
    type, display_name, description,
    category, agent_category, status, is_active,
    capabilities, domain_tags,
    input_contract, output_contract,
    default_config
) VALUES (
    'diagnostician',
    'Diagnosis Loop',
    'Read-only diagnosis loop. Hypothesise -> gather scoped read-only evidence (code bodies + read-only DB rows) -> cite-or-abstain verdict -> re-scope by FOLLOWING evidence (call graph for code, vetted queries for data). Never edits, builds, or triggers a run. One iteration per orchestration; re-invokes itself for the next, carrying revised hypothesis + scope in input_data.',
    'diagnose',
    'analyst',
    'experimental',
    true,
    '["diagnosis", "read-only", "code-analysis", "llm"]'::jsonb,
    '["diagnosis", "debugging", "read-only"]'::jsonb,
    '{"required": ["symptom"], "optional": ["site_id", "correlation_id", "runtime_site", "seed_scope", "hypothesis", "scope", "data_requests", "iteration"], "description": "First call: symptom (+ optional site_id/correlation_id/runtime_site to scope runtime evidence, + optional seed_scope). Re-invocations carry hypothesis/scope/data_requests/iteration set by the previous diagnose_route."}'::jsonb,
    '{"produces": {"diagnosis": "The terminal verdict (CONFIRMED with citations, or UNVERIFIABLE), with the evidence trail across iterations"}}'::jsonb,
    '{
      "workflow": {
        "start_step": "load_runtime",
        "steps": {

          "load_runtime": {
            "action": "diagnose_load_runtime",
            "config": {
              "site_id_field": "input_data.site_id",
              "correlation_id_field": "input_data.correlation_id",
              "domain_field": "input_data.runtime_site",
              "data_requests_field": "input_data.data_requests"
            },
            "next_step": "request_analysis",
            "output_field": "runtime",
            "description": "Read-only runtime evidence (agent_error_log, site_work_items, orchestration_states) + execute the prior verdict''s data_requests read-only"
          },

          "request_analysis": {
            "action": "request_repo_analysis",
            "config": {
              "repo": "gqls/agentchassis"
            },
            "next_step": "lookup_symbols",
            "output_field": "repo_analysis",
            "description": "Analyse the repo (awaits the analyser adapter); leaves the analyser Output (incl. root + commit_sha) under repo_analysis. CONFIRM config keys against the action."
          },

          "lookup_symbols": {
            "action": "lookup_code_symbols",
            "config": {
              "query_field": "input_data.hypothesis"
            },
            "next_step": "assemble",
            "output_field": "code_lookup",
            "description": "First-iteration scope discovery when no seed_scope (returns code_results). CONFIRM config keys against the action."
          },

          "assemble": {
            "action": "diagnose_assemble_bundle",
            "config": {
              "loop_scope_field": "input_data.scope",
              "scope_field": "input_data.seed_scope",
              "code_results_field": "code_lookup.code_results",
              "hypothesis_field": "input_data.hypothesis",
              "seed_hypothesis_field": "input_data.symptom",
              "analysis_field": "repo_analysis",
              "repo_root_field": "repo_analysis.root",
              "runtime_field": "runtime.runtime_evidence"
            },
            "next_step": "verdict",
            "output_field": "bundle",
            "description": "Compose the bundle: hypothesis + in-scope code bodies (via analysis.ReadSymbolBody) + runtime evidence. Scope/hypothesis come from input_data (re-invocation)."
          },

          "verdict": {
            "action": "execute_llm_prompt",
            "config": {
              "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"},
              "max_tokens": 8000,
              "temperature": 0.0,
              "input_fields": ["bundle"],
              "output_format": "json",
              "prompt_template": "PASTE PROMPT_diagnosis_verdict.md HERE (JSON-escaped). It reads the assembled bundle at {{.bundle.bundle}} and returns ONE JSON object: outcome (CONFIRMED|REFUTED|UNVERIFIABLE), citations[], revised_hypothesis, next_scope[], data_requests[{sql,why}]. This placeholder is NOT the real prompt."
            },
            "next_step": "route",
            "output_field": "verdict",
            "description": "Cite-or-abstain verdict over the bundle. ParseVerdict coerces a citation-less or unknown-outcome response to UNVERIFIABLE."
          },

          "route": {
            "action": "diagnose_route",
            "config": {
              "verdict_field": "verdict",
              "iteration_field": "input_data.iteration",
              "max_iterations": 5
            },
            "next_step": "route_branch",
            "output_field": "route",
            "description": "Pure decision: continue = (outcome REFUTED AND next_scope present AND iteration < max). Outputs {continue: bool, iteration: <current+1>}. No DB, no writes."
          },

          "route_branch": {
            "action": "conditional",
            "config": {
              "condition": "route.continue == true",
              "then_step": "spawn_next",
              "else_step": "emit"
            },
            "description": "Continue to the next iteration, or emit the terminal verdict"
          },

          "spawn_next": {
            "action": "spawn_agent",
            "config": {
              "role": "next_iteration",
              "agent_type": "diagnostician"
            },
            "next_step": "call_next",
            "output_field": "spawn_next",
            "description": "Spawn a fresh diagnostician for the next iteration (a sub-agent of the same type)"
          },

          "call_next": {
            "action": "call_agent",
            "config": {
              "target_role": "next_iteration",
              "input_mapping": {
                "site_id?": "input_data.site_id",
                "correlation_id?": "input_data.correlation_id",
                "runtime_site?": "input_data.runtime_site",
                "symptom?": "input_data.symptom",
                "seed_scope?": "input_data.seed_scope",
                "hypothesis": "verdict.revised_hypothesis",
                "scope": "verdict.next_scope",
                "data_requests": "verdict.data_requests",
                "iteration": "route.iteration"
              },
              "timeout_seconds": 1800
            },
            "next_step": "emit",
            "output_field": "child_result",
            "description": "Run the next iteration and await its (eventually terminal) verdict"
          },

          "emit": {
            "action": "diagnose_emit",
            "config": {
              "verdict_field": "verdict",
              "child_field": "child_result"
            },
            "next_step": "complete",
            "output_field": "diagnosis",
            "description": "Emit the terminal verdict: the child''s bubbled-up verdict if we re-invoked, else this iteration''s verdict. Read-only; emits, never acts."
          },

          "complete": {
            "action": "complete_workflow",
            "config": {
              "output_fields": ["diagnosis"]
            },
            "description": "Diagnosis complete (read-only; no fix, no triggered run)"
          }
        }
      },
      "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"},
      "processing_mode": "orchestrator",
      "timeout_seconds": 1800
    }'::jsonb
);

-- Verify:
--   SELECT type, category, agent_category, status, is_active,
--          (default_config ? 'workflow') AS has_workflow
--   FROM agent_definitions WHERE type = 'diagnostician';
-- Expect one row, has_workflow = t.
