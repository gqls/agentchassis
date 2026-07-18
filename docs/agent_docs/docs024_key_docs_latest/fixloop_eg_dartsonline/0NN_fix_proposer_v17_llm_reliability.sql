-- v17 (2026-07-18): adds council reviewer review_llm_reliability (candidate #8),
-- GATED behind the relevance filter, added SURGICALLY (chained jsonb_set on the
-- live config — never a full-config reapply; preserves the co-edited guardian's
-- code_checks + stability proviso). 
-- Chain: ... -> gate_render -> [review_render_guardian?] -> gate_llm_reliability -> [review_llm_reliability?] -> review_guardian

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v17 — LLM-reliability specialist (gated, surgical)')
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
    '{workflow,steps,select_panel,config,footprints,llm_reliability}',
    jsonb_build_array('aiservice','ai_actions','ai_service','max_tokens','llm_call_log','anthropic','ollama','generatetext','stop_reason','budget_tokens','execute_llm_prompt','claude-')
  ),
    '{workflow,steps,review_llm_reliability}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #8) -- LLM-reliability specialist. Advisory only (no veto). Runs only when select_panel flags the fix as relevant.',
      'output_field', 'review_llm_reliability',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: LLM-RELIABILITY SPECIALIST' || chr(10) ||
'' || chr(10) ||
'You judge one thing: does this fix handle LLM call configuration and response semantics correctly? You change nothing; you judge. (This seat exists because two live bugs were found here in one week: a config-precedence trap that silently deadened 17 agents'' token budgets, and a truncation-detection gap that made cut-off responses look like successes.)' || chr(10) ||
'' || chr(10) ||
'## The LLM layer''s load-bearing facts' || chr(10) ||
'- ROOT SHADOWS STEP (BUG B / MDL-039): ExecuteLLMPromptAction reads the agent''s ROOT ai_service block FIRST; a step-level ai_service (including max_tokens) is completely DEAD when a root block exists. The old runbook rule was backwards. A fix placing config at step level for an agent with a root block changes nothing.' || chr(10) ||
'- TRUNCATION LOOKS LIKE SUCCESS (BUG A / MDL-038): GenerateText does not decode stop_reason, so a max_tokens-truncated HTTP 200 returns as a complete success. Signature: llm_call_log rows where output_tokens == max_tokens. A fix consuming LLM output must not assume completeness; a fix touching the client should decode stop_reason.' || chr(10) ||
'- THINKING SPEND: on newer models, omitting the thinking parameter runs ADAPTIVE thinking whose spend comes OUT OF max_tokens -- a small budget can be entirely consumed by thinking, yielding zero text blocks (a hard failure). Newer tokenizers also use ~30% more tokens. Budgets must account for both.' || chr(10) ||
'- SWAP DISCIPLINE: model swaps go through the snapshot_agent backup/rollback machinery and are verified in llm_call_log (model_resolved), never assumed from config.' || chr(10) ||
'- OBSERVABILITY: llm_call_log is the ground truth for what was actually sent/spent; a fix must not break its recording.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit place ai_service config where it will not actually be read (step level under an existing root block), or rely on the inverted old rule; (b) does it consume LLM output while assuming completeness, or set budgets ignoring thinking spend / tokenizer growth; (c) does it swap models without the snapshot/rollback + llm_call_log verification discipline; (d) does it degrade llm_call_log recording. If the fix does not touch LLM calls or their config, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (contracts intact, or the fix does not touch this area), object (names the specific contract/lesson violated, in objections). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "llm_reliability", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_llm_reliability}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate for review_llm_reliability; skip = abstain.',
      'config', jsonb_build_object(
        'condition', 'panel.run_llm_reliability == true',
        'then_step', 'review_llm_reliability',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_render_guardian,next_step}', '"gate_llm_reliability"'::jsonb
  ),
    '{workflow,steps,gate_render,config,else_step}', '"gate_llm_reliability"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_llm_reliability.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_llm_reliability.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_llm_reliability.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_llm_reliability');

COMMIT;
