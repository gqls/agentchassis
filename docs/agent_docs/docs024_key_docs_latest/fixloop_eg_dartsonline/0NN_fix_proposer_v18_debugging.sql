-- v18 (2026-07-18): adds council reviewer review_debug_historian (candidate #9),
-- GATED behind the relevance filter, added SURGICALLY (chained jsonb_set on the
-- live config — never a full-config reapply; preserves the co-edited guardian's
-- code_checks + stability proviso). 
-- Chain: ... -> gate_llm_reliability -> [review_llm_reliability?] -> gate_debugging -> [review_debug_historian?] -> review_guardian

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v18 — debugging/incident-lore historian (gated, surgical)')
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
    '{workflow,steps,select_panel,config,footprints,debugging}',
    jsonb_build_array('.go','platform/','internal/','cmd/','.sql')
  ),
    '{workflow,steps,review_debug_historian}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #9) -- debugging/incident-lore historian. Advisory only (no veto). Runs only when select_panel flags the fix as relevant.',
      'output_field', 'review_debug_historian',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: DEBUGGING / INCIDENT-LORE HISTORIAN' || chr(10) ||
'' || chr(10) ||
'You judge one thing: does this fix''s verification and data-surgery approach avoid the platform''s documented investigation traps? You change nothing; you judge. (Deliberately broad: you carry the largest body of hard-won debugging lore -- 74 documented lessons -- and fire on most code fixes; your job is the cheap question "we''ve been burned by this before, does the plan account for it?")' || chr(10) ||
'' || chr(10) ||
'## The lore that keeps being relearned' || chr(10) ||
'- NEEDLE-GATE SQL SURGERY: production text/jsonb mutations need: dump + backup first; a read-only needle-gate asserting every expected needle WITH a mechanically-derived occurrence count (never from memory); guarded idempotent UPDATE (exact-string replace or anchored regexp_replace, gated on a pre-state marker so re-runs are 0-row no-ops); RETURNING post-conditions; separate verify + rollback files.' || chr(10) ||
'- POSTGRES PITFALLS: replace() silently no-ops on a missed anchor while still reporting UPDATE 1; LIKE treats a needle''s % as a wildcard; bounded regex quantifiers cap at 255; substring(...from ''(p)'') returns only the FIRST capture group; an aborted transaction is sticky (open migrations with a defensive ROLLBACK, run via psql -f); a 0-rows result from your own verification query is not decisive either.' || chr(10) ||
'- INFORMATIONAL COLUMNS: sites.status is informational -- nothing in dispatch filters on ''active''; never scope blast-radius by it. Enumerate GROUP BY status before any blast-radius query; check pg_proc/pg_trigger before adding helpers that may already exist.' || chr(10) ||
'- VERIFY AGAINST THE POD: a deploy is verified by grepping the RUNNING pod''s binary for the change''s symbol -- never git, never the image tag (same-tag rebuilds ship stale binaries).' || chr(10) ||
'- REPAIR VS REGENERATE: broken stored templates split into Mode A (<no value>FIELD</no> -- field names survive, repairable in place) and Mode B (bare <no value> -- names irretrievably lost, only regeneration works). Attempting to repair Mode B is doomed; route it to regeneration.' || chr(10) || chr(10) ||
'Judge the plan: (a) does its verification/migration approach use the needle-gate discipline (backups, counted needles, guarded idempotent updates, separate verify/rollback), or naive replace/regex prone to the documented pitfalls; (b) does it scope blast radius by informational columns (the sites.status=active class) or skip enumerating real values first; (c) does its deploy-verification rely on git/tags instead of the running pod; (d) for stored-template/content surgery, does it respect the repair-vs-regenerate taxonomy. If the plans approach is sound on all four, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (contracts intact, or the fix does not touch this area), object (names the specific contract/lesson violated, in objections). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "debug_historian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_debugging}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate for review_debug_historian; skip = abstain.',
      'config', jsonb_build_object(
        'condition', 'panel.run_debugging == true',
        'then_step', 'review_debug_historian',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_llm_reliability,next_step}', '"gate_debugging"'::jsonb
  ),
    '{workflow,steps,gate_llm_reliability,config,else_step}', '"gate_debugging"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_debug_historian.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_debug_historian.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_debug_historian.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_debug_historian');

COMMIT;
