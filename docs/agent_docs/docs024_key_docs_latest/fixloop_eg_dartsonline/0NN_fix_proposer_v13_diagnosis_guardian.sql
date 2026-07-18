-- v13 (2026-07-18): adds an EIGHTH council reviewer, review_diagnosis_guardian
-- (candidate #4), GATED behind the relevance filter, added SURGICALLY (the v12
-- pattern: chained jsonb_set on the live config — never a full-config reapply,
-- which would clobber the co-edited guardian's code_checks + stability proviso).
--
-- Grounding (docs026 register, diagnosis-loop.md — the highest hot-concept
-- density category): DIAG-001 (read-only cite-or-abstain core), DIAG-008
-- (three-tier citation standard), DIAG-009 (read-only SQL enforcement),
-- DIAG-028 (persistence never fails a diagnosis; skip-never-guess), DIAG-030
-- (error_step is CONFIG-level; step-level is silently inert), DIAG-019/022
-- (spawn-wrapper keeps loop work + repo tokens off shared pods).
-- Lens: does the fix weaken the diagnosis machinery's honesty gates or safety
-- disciplines? Advisory only (approve|object, no veto).
--
-- Chain tail becomes:
--   ... -> gate_adoption -> [adoption_guardian?] -> gate_diagnosis
--       -> [diagnosis_guardian?] -> review_guardian -> council_decide

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v13 — diagnosis-loop guardian (gated, surgical)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions d
SET default_config =
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
  jsonb_set(
    d.default_config,
    '{workflow,steps,select_panel,config,footprints,diagnosis}',
    jsonb_build_array('diagnose_','diagnosis_artifacts','diagnose-agent','diagnose-orchestrator','pkg/diagnose','verdict','cite-or-abstain','data_request','sqlguard','symptom')
  ),
    '{workflow,steps,review_diagnosis_guardian}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #4) -- diagnosis-loop guardian: does the fix weaken the diagnosis machinery''s honesty gates or safety disciplines? Advisory only (no veto). Runs only when select_panel flags the fix as touching the diagnosis loop.',
      'output_field', 'review_diagnosis_guardian',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: DIAGNOSIS-LOOP GUARDIAN' || chr(10) || chr(10) ||
'You judge one thing: does this fix weaken the diagnosis machinery''s honesty gates or safety disciplines? You change nothing; you judge. (Yes -- the loop reviewing a fix to itself is exactly the point: its guards were earned from real failures and must not be quietly loosened.)' || chr(10) || chr(10) ||
'## The diagnosis loop''s load-bearing disciplines' || chr(10) ||
'- READ-ONLY, CITE-OR-ABSTAIN: the loop reads code and live state but never writes; every verdict claim cites evidence or the verdict abstains (UNVERIFIABLE is an honest terminal, not a failure). A fix that lets a verdict assert uncited claims, or gives the loop a write path, guts the core contract.' || chr(10) ||
'- THREE-TIER CITATIONS: a CONFIRMED verdict needs BOTH a static (code) citation AND a state/runtime citation. Weakening the tier guard turns plausible-but-unverified theories into CONFIRMED.' || chr(10) ||
'- READ-ONLY SQL, THREE GUARDS: model-authored data_requests run under layered enforcement (sqlguard lint, read-only role, EXPLAIN size pre-flight). Removing one layer because "the others catch it" is how layers die.' || chr(10) ||
'- OBSERVABILITY NEVER COSTS A DIAGNOSIS: persistence (diagnosis_artifacts, notes) degrades to a logged warning on failure -- it must never fail the run. And notes are skip-never-guess: a mis-filed note poisons history.' || chr(10) ||
'- CONFIG-LEVEL error_step: the workflow coordinator reads ONLY step.config.error_step -- a step-level error_step is parsed but silently inert (a real, recurring trap). Any plan adding error routing must place it inside config.' || chr(10) ||
'- TOKEN/POD ISOLATION: heavy loop work runs in a spawned diagnose-agent pod; shared chassis pods never hold the repo-read token. A fix that moves loop work (or the token) onto shared pods breaks the isolation.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit weaken a verdict guard (cite-or-abstain, tier coverage, symptom gate) or add an uncited-assertion path; (b) does it touch the read-only SQL enforcement layers; (c) does it make persistence/observability able to fail a diagnosis, or guess a note subject; (d) does it place error_step outside config (silently inert), or move loop work/tokens onto shared pods. If the fix does not touch the diagnosis machinery, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (disciplines intact, or does not touch the loop), object (weakens a discipline above -- name which). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (an artifact kind exists, a guard config value), put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "diagnosis_guardian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the discipline weakened", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_diagnosis}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate: run the diagnosis-loop guardian only if the fix touches the diagnosis machinery; else skip (it abstains).',
      'config', jsonb_build_object(
        'condition', 'panel.run_diagnosis == true',
        'then_step', 'review_diagnosis_guardian',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_adoption_guardian,next_step}', '"gate_diagnosis"'::jsonb
  ),
    '{workflow,steps,gate_adoption,config,else_step}', '"gate_diagnosis"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_diagnosis_guardian.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_diagnosis_guardian.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_diagnosis_guardian.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_diagnosis_guardian');

COMMIT;

-- Rollback (manual): restore the pre-update snapshot from agent_definitions_backup.
