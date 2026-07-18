-- v14 (2026-07-18): adds a NINTH council reviewer, review_improvement_guardian
-- (candidate #5), GATED behind the relevance filter, added SURGICALLY (the
-- v12/v13 pattern: chained jsonb_set on the live config).
--
-- Grounding (docs026 register, improvement-loop.md): IMP-003 (the
-- discovery→triage→fix→rerender cycle with its pass cap), IMP-004 (discovery
-- check registry contracts), IMP-027 (the triage-drain fix — the loop once ran
-- UNBOUNDED, 845+ items in ~10 days, consuming most of the token budget; the
-- guards exist because of that incident). Lens: does the fix respect the
-- improvement loop's termination + enablement contracts? Advisory only.
--
-- Chain tail becomes:
--   ... -> gate_diagnosis -> [diagnosis_guardian?] -> gate_improvement
--       -> [improvement_guardian?] -> review_guardian -> council_decide

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v14 — improvement-loop guardian (gated, surgical)')
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
    '{workflow,steps,select_panel,config,footprints,improvement}',
    jsonb_build_array('improvement','discovery_check','run_discovery_checks','write_audit_findings','triage_detected','audit_pass','complete_clean','locked_at','acceptance_test','needs_rerender','maintenance_queue')
  ),
    '{workflow,steps,review_improvement_guardian}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #5) -- improvement-loop guardian: does the fix respect the improvement loop''s termination + enablement contracts? Advisory only (no veto). Runs only when select_panel flags the fix as touching the improvement/discovery machinery.',
      'output_field', 'review_improvement_guardian',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: IMPROVEMENT-LOOP GUARDIAN' || chr(10) || chr(10) ||
'You judge one thing: does this fix respect the improvement loop''s termination and enablement contracts? You change nothing; you judge. (These guards exist because the audit->fix->re-audit loop once ran UNBOUNDED -- 845+ findings across 4 domains in ~10 days, consuming most of the token budget. Do not let a fix quietly reopen that.)' || chr(10) || chr(10) ||
'## The improvement loop''s load-bearing contracts' || chr(10) ||
'- BOUNDED PASSES: audits are capped (pass-limit gate, >=3 -> complete_clean); sections that pass get locked_at and later audits SKIP them; unlock is always MANUAL. A fix that bypasses the pass cap, auto-unlocks sections, or re-audits locked sections reopens the unbounded drain.' || chr(10) ||
'- CONFIG-ONLY ENABLEMENT: discovery checks self-register via init() but run ONLY when named in the agent''s checks array; unknown names warn-and-skip. A check must only be ENABLED once its handler agent exists -- otherwise findings accumulate unconsumed. Findings insert at status=detected (unclaimable), so a check can run observe-only; the triager promotes detected->triaged.' || chr(10) ||
'- RUNNER OWNS INSERTION: checks append WorkItemSpecs; the RUNNER inserts them with dedup. A check/plugin that inserts its own rows bypasses dedup and the two-strike anti-churn machinery.' || chr(10) ||
'- CHEAP VERIFICATION: fixes are verified via the finding''s acceptance_test (a cheap targeted call), never a full re-audit; per-page sequential processing via depends_on prevents overlapping fixes.' || chr(10) ||
'- DISPATCH DISCIPLINE: the sweep cadence lives in scheduled_tasks under a shared dispatch concurrency group; discovery never floods an already-backed-up queue.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit bypass the pass cap, unlock sections automatically, or re-audit locked sections; (b) does it enable a discovery check whose handler agent does not exist, or make findings insert at a claimable status; (c) does it make a check insert its own work items instead of appending WorkItemSpecs for the runner; (d) does it replace acceptance-test verification with full re-audits, or break the depends_on sequencing / dispatch concurrency discipline. If the fix does not touch the improvement/discovery machinery, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (contracts intact, or does not touch the loop), object (breaks a contract above -- name which). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (is a check named in an agent''s checks array, does a handler agent exist), put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "improvement_guardian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the contract broken", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_improvement}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate: run the improvement-loop guardian only if the fix touches the improvement/discovery machinery; else skip (it abstains).',
      'config', jsonb_build_object(
        'condition', 'panel.run_improvement == true',
        'then_step', 'review_improvement_guardian',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_diagnosis_guardian,next_step}', '"gate_improvement"'::jsonb
  ),
    '{workflow,steps,gate_diagnosis,config,else_step}', '"gate_improvement"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_improvement_guardian.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_improvement_guardian.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_improvement_guardian.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_improvement_guardian');

COMMIT;

-- Rollback (manual): restore the pre-update snapshot from agent_definitions_backup.
