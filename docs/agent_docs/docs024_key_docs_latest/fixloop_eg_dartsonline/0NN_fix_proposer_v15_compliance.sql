-- v15 (2026-07-18): adds council reviewer review_compliance (candidate #6),
-- GATED behind the relevance filter, added SURGICALLY (chained jsonb_set on the
-- live config — never a full-config reapply; preserves the co-edited guardian's
-- code_checks + stability proviso). Justified by SEVERITY not frequency: two live fabrication incidents.
-- Chain: ... -> gate_improvement -> [review_improvement_guardian?] -> gate_compliance -> [review_compliance?] -> review_guardian

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v15 — compliance/legal eye (gated, surgical)')
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
    '{workflow,steps,select_panel,config,footprints,compliance}',
    jsonb_build_array('pricing','product_price','legal','compliance','testimonial','evidence_base','fabricat','banned_claim','claims_','vet_med','refund','privacy','disclaimer')
  ),
    '{workflow,steps,review_compliance}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #6) -- compliance/legal eye. Advisory only (no veto). Runs only when select_panel flags the fix as relevant.',
      'output_field', 'review_compliance',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: COMPLIANCE / LEGAL EYE' || chr(10) ||
'' || chr(10) ||
'You judge one thing: could this fix cause the platform to publish claims, prices, or content that is fabricated, unevidenced, or legally exposed? You change nothing; you judge. (This seat exists because of two LIVE incidents: fabricated veterinary prices that had to be stripped and legally recorded, and fabricated marketing claims -- including a poisoned writing rule in a site spec that INSTRUCTED the fabrication.)' || chr(10) ||
'' || chr(10) ||
'## The platform''s content-integrity contracts' || chr(10) ||
'- NO UNEVIDENCED CLAIMS: user-facing factual claims (prices, statistics, testimonials, capability claims) must be backed by evidence (the evidence_base / audit-row discipline: no claim ships without a source). Generated content follows smallest-true-claim, anti-hype voice rules.' || chr(10) ||
'- POISONED-SPEC CLASS: fabrication can be INSTRUCTED upstream -- a writing rule or spec that tells a writer to invent data. A fix touching prompts, specs, or writing rules must not introduce instructions that could cause invention of facts, prices, or reviews.' || chr(10) ||
'- LEGAL SURFACE: disclaimers must be conspicuous and proximate (in the deliverable, not just a footer); legal pages (terms/refund/privacy) state AI-can-be-wrong honestly; every claim cited and date-stamped; information-not-advice posture.' || chr(10) ||
'- SCANNERS AND GATES: claims-verification machinery (scans, banned-claim lists, evidence gates) must not be weakened or bypassed by a fix.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit generate or alter user-facing claims/prices/testimonials without evidence backing; (b) does it weaken or bypass a claims gate, scanner, or banned-claim list; (c) does it touch prompts/specs/writing-rules in a way that could instruct fabrication (the poisoned-spec class); (d) does it remove or obscure disclaimers or legal pages. If the fix does not touch user-facing content, claims machinery, or legal surface, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (contracts intact, or the fix does not touch this area), object (names the specific contract/lesson violated, in objections). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "compliance", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_compliance}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate for review_compliance; skip = abstain.',
      'config', jsonb_build_object(
        'condition', 'panel.run_compliance == true',
        'then_step', 'review_compliance',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_improvement_guardian,next_step}', '"gate_compliance"'::jsonb
  ),
    '{workflow,steps,gate_improvement,config,else_step}', '"gate_compliance"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_compliance.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_compliance.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_compliance.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_compliance');

COMMIT;
