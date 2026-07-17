-- v12 (2026-07-17): adds a SEVENTH council reviewer, review_adoption_guardian,
-- as the FIRST seat built BEHIND the relevance filter (candidate #3). Grounded
-- in ADO-006 (the other original stage-1 flagged rediscovered concept:
-- "adoption writes specs first, classifier consumes under fidelity rules") +
-- ADO-003. Lens: does the fix respect the adoption pipeline's event-driven,
-- write-then-relay contract, or break it? Advisory only (approve|object,
-- no veto).
--
-- ██ SURGICAL BY DESIGN ██ — the fix-proposer workflow is co-edited (the
-- guardian gained a code_checks mechanism, and a stability-preference proviso).
-- The v6-v11 full-config reapply pattern would clobber those. So this is a
-- chain of jsonb_set operations on the LIVE default_config, touching only the
-- specific paths it needs. Everything else (code_checks, the guardian proviso,
-- the other seats, the filter wiring) is left byte-identical. Idempotent
-- (refuses if review_adoption_guardian already present).
--
-- Gated, not always-on: adds an "adoption" footprint to select_panel and a
-- gate before the new seat, inserted at the tail of the gated chain:
--   ... -> gate_tooling_provenance -> [tooling_provenance?] -> gate_adoption
--       -> [adoption_guardian?] -> review_guardian -> council_decide
-- Requires the chassis image carrying select_review_panel (live since v1.0.1133).

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v12 — adoption-pipeline guardian (gated, surgical)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions d
SET default_config =
  -- 8. run_checks.check_fields += the new seat's checks
  jsonb_set(
  -- 7. escalate.review_fields += the new seat
  jsonb_set(
  -- 6. council_decide.review_fields += the new seat
  jsonb_set(
  -- 5. gate_tooling_provenance.else_step: review_guardian -> gate_adoption
  jsonb_set(
  -- 4. review_tooling_provenance.next_step: review_guardian -> gate_adoption
  jsonb_set(
  -- 3. add gate_adoption (conditional)
  jsonb_set(
  -- 2. add review_adoption_guardian (the reviewer)
  jsonb_set(
  -- 1. add the "adoption" footprint to select_panel
  jsonb_set(
    d.default_config,
    '{workflow,steps,select_panel,config,footprints,adoption}',
    jsonb_build_array('apply_adoption_plan','site-adoption','adoption','needs_domain_research','domain-research-classifier','site_archetype','design_intent','design_reference','content_direction')
  ),
    '{workflow,steps,review_adoption_guardian}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (gated, candidate #3) -- adoption-pipeline guardian: does the fix respect the adoption pipeline''s write-then-relay contract (ADO-006)? Advisory only (no veto). Runs only when select_panel flags the fix as touching adoption.',
      'output_field', 'review_adoption_guardian',
      'next_step', 'review_guardian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-4-6','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',3000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template',
'# Council reviewer: ADOPTION-PIPELINE GUARDIAN' || chr(10) || chr(10) ||
'You judge one thing: does this fix respect the adoption pipeline''s event-driven, write-then-relay contract, or break it? You change nothing; you judge.' || chr(10) || chr(10) ||
'## The adoption pipeline''s load-bearing contracts' || chr(10) ||
'- WRITE-THEN-RELAY: apply_adoption_plan writes the specs (site_archetype, design_reference, design_intent, content_direction, identity), pages, and work items ITSELF, then emits exactly ONE strategic item: needs_domain_research. It never calls the classifier/planner directly, and never emits build-stage items (needs_composition / needs_design) directly. A fix that makes adoption call a downstream stage directly, or emit a build-stage item, breaks the relay contract.' || chr(10) ||
'- ADOPTED SPECS ARE GROUND TRUTH: when the relay reaches the site, the domain-research-classifier treats the adopted identity/archetype/content_direction/design_intent as ground truth that OUTRANKS its own search+scrape -- it reads-and-extends, never overwrites. A fix that makes the classifier overwrite adopted specs breaks fidelity.' || chr(10) ||
'- NO BYPASS: adopted sites run the full strategist -> briefing -> planner chain exactly as fresh builds -- adoption routes THROUGH the planner, it does not replace it.' || chr(10) ||
'- LLM FOR REASONING, GO FOR EXTRACTION: never pay an LLM to transcribe what a regex/goquery can read (hex colours, fonts, CSS vars are extracted Go-side).' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit make adoption call a downstream stage directly, or emit a build-stage work item instead of the single needs_domain_research relay; (b) does it let the classifier overwrite (rather than read-and-extend) adopted specs; (c) does it bypass the strategist/briefing/planner chain for adopted sites; (d) does it move extraction work onto an LLM that Go should do. If the fix does not touch the adoption pipeline, approve.' || chr(10) || chr(10) ||
'Verdicts: approve (respects the contracts, or does not touch adoption), object (breaks a contract above -- name which). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (does a work_item item_type exist, what does an agent emit), put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "adoption_guardian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the adoption contract broken", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,gate_adoption}',
    jsonb_build_object(
      'action', 'conditional',
      'description', 'Relevance gate: run the adoption-pipeline guardian only if the fix touches the adoption/onboarding pipeline; else skip (it abstains).',
      'config', jsonb_build_object(
        'condition', 'panel.run_adoption == true',
        'then_step', 'review_adoption_guardian',
        'else_step', 'review_guardian'
      )
    )
  ),
    '{workflow,steps,review_tooling_provenance,next_step}', '"gate_adoption"'::jsonb
  ),
    '{workflow,steps,gate_tooling_provenance,config,else_step}', '"gate_adoption"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_adoption_guardian.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_adoption_guardian.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_adoption_guardian.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_adoption_guardian');

COMMIT;

-- Rollback (manual): restore the pre-update snapshot from agent_definitions_backup.
