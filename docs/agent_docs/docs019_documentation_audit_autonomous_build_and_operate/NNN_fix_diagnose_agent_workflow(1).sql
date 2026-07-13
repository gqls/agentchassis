-- NNN_fix_diagnose_agent_workflow.sql
--
-- Renumber NNN to the next number in your migration sequence (3-digit prefixes).
--
-- SUPERSEDES NNN_move_diagnose_workflow_to_default_config.sql for diagnose-agent:
-- that one MOVED the seeded workflow, but the seeded workflow references a
-- `diagnose_run` action THAT DOES NOT EXIST. The real built design is the
-- workflow-driven one (the diagnose_load_runtime / diagnose_assemble_bundle /
-- diagnose_route / diagnose_emit actions): gather -> verdict (execute_llm_prompt)
-- -> diagnose_route (engine guards + call-graph re-scope, next_step override) ->
-- back to assemble_bundle | emit. So this REWRITES diagnose-agent's workflow to
-- that shape, into default_config (the column the loader reads), with the verdict
-- prompt inline and an ai_service block (so the model is changeable in one place).
-- The diagnose-orchestrator workflow is correct (spawn+call) and is just moved to
-- default_config + given processing_mode/timeout_seconds.
--
-- Schema (verified): default_config jsonb NOT NULL; orchestration_workflow json.
-- This OVERWRITES diagnose-agent.default_config for the row (intended — it installs
-- the correct workflow). Re-running is safe (same value).

-- BACKUP FIRST (standing rule): snapshot each row before changing it.
SELECT snapshot_agent('diagnose-agent', 'rewrite workflow to diagnose_route in default_config');
SELECT snapshot_agent('diagnose-orchestrator', 'move workflow to default_config');

-- diagnose-agent: install the corrected workflow-driven loop in default_config.
UPDATE agent_definitions
SET default_config = '{
  "workflow": {
    "steps": {
      "analyse_repo": {
        "action": "request_repo_analysis",
        "config": {
          "language": "go",
          "ref_field": "input_data.ref",
          "repo_field": "input_data.repo",
          "owner_field": "input_data.owner"
        },
        "next_step": "lookup_symbols",
        "output_field": "repo_analysis",
        "description": "Analyse the repo at ref; awaits the analyser adapter (leaves the Output incl. root + commit_sha)"
      },
      "lookup_symbols": {
        "action": "lookup_code_symbols",
        "config": {
          "top_k": 12,
          "repo_field": "repo_analysis.repo",
          "query_field": "input_data.symptom"
        },
        "next_step": "load_runtime",
        "output_field": "code_lookup",
        "description": "Seed the FIRST scope from the symptom (B4a: retrieval seeds iter-1 only; runtime steers re-scopes)"
      },
      "load_runtime": {
        "action": "diagnose_load_runtime",
        "config": {
          "domain_field": "input_data.runtime_site",
          "site_id_field": "input_data.site_id",
          "correlation_id_field": "input_data.correlation_id",
          "data_requests_field": "route.data_requests"
        },
        "next_step": "assemble_bundle",
        "output_field": "runtime",
        "description": "Read-only runtime tier (agent_error_log, site_work_items, orchestration_states). Runs once unless the loop returns here."
      },
      "assemble_bundle": {
        "action": "diagnose_assemble_bundle",
        "config": {
          "loop_scope_field": "route.scope",
          "scope_field": "input_data.seed_scope",
          "code_results_field": "code_lookup.code_results",
          "hypothesis_field": "route.hypothesis",
          "seed_hypothesis_field": "input_data.symptom",
          "analysis_field": "repo_analysis",
          "repo_root_field": "repo_analysis.root",
          "runtime_field": "runtime.runtime_evidence"
        },
        "next_step": "verdict",
        "output_field": "bundle",
        "description": "Compose bundle: hypothesis + in-scope code bodies (analysis.ReadSymbolBody) + runtime. Loop returns HERE with route.scope."
      },
      "verdict": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {
            "model": "claude-sonnet-4-6",
            "provider": "anthropic",
            "api_key_env_var": "ANTHROPIC_API_KEY"
          },
          "max_tokens": 8000,
          "temperature": 0.0,
          "input_fields": [
            "bundle"
          ],
          "output_format": "json",
          "prompt_template": "# PROMPT \u2014 diagnosis verdict (cite-or-abstain)\n\nThis is the prompt the (chassis-side) model Verdicter runs once per loop iteration\n(`internal/diagnose`, design `DESIGN_diagnosis_loop.md` \u00a72). It is given the\ncurrent hypothesis and the assembled evidence bundle, and must return ONE verdict\nas JSON in the wire format below (`verdict_wire.go` parses it; the scaffold''s\nguards then act on it).\n\nThe output schema here MUST stay in lockstep with `verdict_wire.go` \u2014 that file''s\ntests are the seam check. The verdict-script used in testing is this exact format,\nso a script is a faithful stand-in for the model.\n\n---\n\n## System / role\n\nYou are a debugging analyst examining ONE hypothesis about a bug against a fixed\nbody of evidence (the bundle). The bundle contains: a constitution (project\nrules), the current hypothesis, in-scope source code in full, schema, and any\nruntime evidence (error logs, work-item state, DB rows). Your single job is to\njudge whether the evidence in the bundle **confirms**, **refutes**, or **does not\nsettle** the hypothesis \u2014 and to ground that judgement in verbatim quotes from the\nbundle.\n\nYou do NOT propose a fix. You do NOT speculate beyond the evidence. You judge and\ncite. A human acts on your verdict; the loop never changes code.\n\n## The three verdicts\n\nReturn exactly one `outcome`:\n\n- **CONFIRMED** \u2014 the evidence DIRECTLY supports the hypothesis. Not \"consistent\n  with\", not \"plausible given\" \u2014 direct. You must quote the specific evidence that\n  confirms it. If you cannot quote direct support, the outcome is NOT confirmed.\n\n- **REFUTED** \u2014 the evidence CONTRADICTS the hypothesis. **This is a correct and\n  expected outcome, not a failure.** If the bundle shows the hypothesis is wrong,\n  say so plainly and quote the contradicting evidence. Then state what the evidence\n  shows INSTEAD (`revised_hypothesis`) and where to look next (`next_scope`). The\n  most valuable thing you can do is abandon a wrong hypothesis the moment the\n  evidence breaks it \u2014 do not rescue a hypothesis the bundle contradicts.\n\n- **UNVERIFIABLE** \u2014 the bundle does not contain enough to confirm or refute. Name\n  the SPECIFIC evidence that would settle it (`needed_evidence`): a table to query,\n  a log to pull, a symbol to add to scope. Abstaining is correct when the evidence\n  is genuinely absent \u2014 far better than a confident guess.\n\n## Hard rules\n\n1. **Cite or abstain.** A CONFIRMED or REFUTED verdict MUST carry at least one\n   citation quoting the bundle. No citation \u2192 you may only return UNVERIFIABLE.\n   (The loop enforces this: a citation-less confirm/refute is coerced to\n   UNVERIFIABLE, so an un-grounded verdict is wasted \u2014 ground it.)\n\n2. **Quotes are verbatim.** Each citation''s `quote` is text copied from the bundle\n   exactly \u2014 a log line, a line of code, a schema row. Never paraphrase a quote;\n   the human verifies it against the bundle. Paraphrase belongs in the hypothesis\n   fields, never in a quote.\n\n3. **Confirm only on direct evidence (the asymmetry).** \"The logs are consistent\n   with X\" is UNVERIFIABLE, not CONFIRMED. Runtime evidence readily REFUTES (an\n   error that shouldn''t be there breaks a hypothesis) but CONFIRMS only when it\n   directly shows the mechanism, not merely a symptom compatible with it.\n\n4. **Follow the evidence to the next scope \u2014 do not re-search the symptom.** When\n   you REFUTE or abstain, `next_scope` should name the symbols/files the EVIDENCE\n   points at \u2014 the function the failing code calls, the action named in the error,\n   the symbol the trace implicates. (The loop then follows the call graph from\n   there.) The cause often lives in shared infrastructure named NOTHING like the\n   symptom; you reach it by following what the evidence names, not by re-describing\n   the symptom. If runtime evidence names a fault site (an agent, a step, a table),\n   put it in `runtime_site` so the next bundle re-gathers runtime there.\n\n5. **No fix.** Never output a code change, patch, or \"the fix is\u2026\". You diagnose;\n   the human fixes.\n\n6. **Tag each citation''s tier** \u2014 `static` (code/schema), `state` (a DB row at a\n   point in time), or `runtime` (a log/work-item from an actual run) \u2014 and for\n   state/runtime give `fresh` (when it was observed) so a verdict resting on stale\n   evidence is visible.\n\n7. **`data_requests` are READ-ONLY, and only SELECT.** When the bundle doesn''t\n   settle the hypothesis and a specific query would, you MAY ask for it in\n   `data_requests`. Each `sql` MUST be a single read-only `SELECT` (or\n   `WITH \u2026 SELECT`) written against the schema shown in the bundle \u2014 no `INSERT`,\n   `UPDATE`, `DELETE`, `MERGE`, DDL (`DROP`/`ALTER`/`CREATE`/`TRUNCATE`),\n   `GRANT`/`REVOKE`, `COPY`, `CALL`, or multiple statements. Anything else is\n   rejected and the request is dropped, so it only wastes an iteration. These run\n   under a read-only connection \u2014 you cannot change anything, so don''t try; ask\n   only for the narrowest read that would settle the question, and prefer naming\n   the table/columns you saw in the bundle''s schema section over guessing.\n\n## Output \u2014 return ONLY this JSON, nothing else\n\n```json\n{\n  \"outcome\": \"CONFIRMED | REFUTED | UNVERIFIABLE\",\n  \"citations\": [\n    {\"tier\": \"static|state|runtime\", \"where\": \"path:Symbol or table or log source\",\n     \"quote\": \"VERBATIM text from the bundle\", \"fresh\": \"when observed (state/runtime; omit for static)\"}\n  ],\n  \"revised_hypothesis\": \"REFUTED only: what the evidence shows instead\",\n  \"next_scope\": [\"REFUTED/UNVERIFIABLE: symbols or files the evidence points to\"],\n  \"needed_evidence\": \"UNVERIFIABLE only: the specific evidence that would settle it\",\n  \"runtime_site\": \"optional: a runtime fault site (agent/step/table) to gather next\",\n  \"data_requests\": [\n    {\"sql\": \"a SINGLE read-only SELECT or WITH \u2026 SELECT, against the schema shown in the bundle\",\n     \"why\": \"what this query would settle\"}\n  ]\n}\n```\nFields not relevant to the outcome may be omitted or left empty. Emit no prose\noutside the JSON object.\n\n## Worked example (the move that matters)\n\nHypothesis given: *\"the page rebuild reports success but the page is stale because\nthe writer''s sections never reach save_page_sections.\"*\n\nBundle (excerpt): `save_page_sections_action.go:SavePageSectionsAction` in full;\nruntime `agent_error_log` showing repeated rows:\n`step save_sections failed: content regression blocked: new content has 2854 chars\nvs 13040 existing`.\n\nCorrect verdict \u2014 the evidence REFUTES the hypothesis (the sections DO reach save;\nsave blocks them), and points the next scope upstream:\n\n```json\n{\n  \"outcome\": \"REFUTED\",\n  \"citations\": [\n    {\"tier\": \"runtime\", \"where\": \"agent_error_log\",\n     \"quote\": \"step save_sections failed: content regression blocked: new content has 2854 chars vs 13040 existing\",\n     \"fresh\": \"2026-06-14\"}\n  ],\n  \"revised_hypothesis\": \"the sections reach save but are far shorter than the existing page; the regeneration upstream is producing too little, and a guard blocks the overwrite\",\n  \"next_scope\": [\"plan_sections_action.go:PlanSectionsAction\"],\n  \"runtime_site\": \"page-build-handler\"\n}\n```\n\nThis is the behaviour the whole loop exists to produce: the hypothesis was stated\nconfidently and the evidence broke it; the right move is to REFUTE on the quoted\nlog line and re-point upstream \u2014 not to defend the original guess. (In the real\ncase this re-scope, followed across two more iterations, reached the actual cause\nin the coordinator''s result extraction \u2014 a symbol the symptom could never have\nnamed, reached only by following the evidence.)\n\n## A caution to apply to yourself\n\nTreat the bundle, and your own reading of it, with the same suspicion you apply to\nthe hypothesis. If a quote doesn''t actually say what you want it to, it is not\nsupport. If the bundle is missing the table or log you''d need, say UNVERIFIABLE and\nname it \u2014 do not infer the missing piece. A confident wrong verdict is worse than\nan honest abstention, because the loop and the human will trust your citation.\n\n---\n\n## Bundle under examination\n\nEverything you judge is in this bundle (the hypothesis under test is stated at its top; in-scope code and any runtime/DB evidence follow). Quote only from here.\n\n{{.bundle.bundle}}\n"
        },
        "next_step": "route",
        "output_field": "verdict",
        "description": "Cite-or-abstain verdict over the bundle (its own observable step). Output JSON at verdict.result."
      },
      "route": {
        "action": "diagnose_route",
        "config": {
          "verdict_field": "verdict.result",
          "state_field": "diagnose_state",
          "analysis_field": "repo_analysis",
          "gather_step": "assemble_bundle",
          "emit_step": "emit",
          "max_iterations": 5
        },
        "output_field": "route",
        "description": "Controller: run engine guards + call-graph re-scope (Advance), override next_step to loop back to assemble_bundle or stop at emit."
      },
      "emit": {
        "action": "diagnose_emit",
        "config": {
          "status_field": "route.status",
          "conclusion_field": "route.conclusion",
          "stopped_by_field": "route.stopped_by",
          "trail_field": "route.evidence_trail"
        },
        "next_step": "complete",
        "output_field": "diagnosis",
        "description": "Shape the human-facing diagnosis + evidence trail from the loop''s terminal state. Never a fix."
      },
      "complete": {
        "action": "complete_workflow",
        "config": {
          "result_from": "diagnosis"
        },
        "description": "Return the diagnosis to the caller (responses topic)"
      }
    },
    "start_step": "analyse_repo"
  },
  "ai_service": {
    "model": "claude-sonnet-4-6",
    "provider": "anthropic",
    "api_key_env_var": "ANTHROPIC_API_KEY"
  },
  "processing_mode": "orchestrator",
  "timeout_seconds": 1800
}'::jsonb,
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-agent';

-- diagnose-orchestrator: move its (correct) workflow to default_config + add the
-- top-level keys the working agents carry.
UPDATE agent_definitions
SET default_config = orchestration_workflow::jsonb
      || jsonb_build_object('processing_mode', 'orchestrator', 'timeout_seconds', 600),
    orchestration_workflow = NULL,
    updated_at = now()
WHERE type = 'diagnose-orchestrator'
  AND orchestration_workflow IS NOT NULL
  AND default_config = '{}'::jsonb;

-- Verify:
--   SELECT type,
--          (default_config ? 'workflow')                          AS has_workflow,
--          (default_config -> 'workflow' -> 'steps' ? 'route')     AS has_route_step,
--          (default_config -> 'workflow' -> 'steps' ? 'run_loop')  AS still_has_run_loop,
--          (default_config ->> 'processing_mode')                  AS mode,
--          (orchestration_workflow IS NULL)                        AS orch_null
--   FROM agent_definitions WHERE type IN ('diagnose-agent','diagnose-orchestrator');
-- diagnose-agent: has_workflow=t, has_route_step=t, still_has_run_loop=f, mode=orchestrator, orch_null=t
