-- v20 (2026-07-20): adds ALWAYS-ON council reviewer review_prior_art (the
-- "prior-art librarian"; owner suggestion 2026-07-20) to fix-proposer,
-- SURGICALLY. Class: asserted-absence / dormant-machinery -- no existing seat
-- verifies a rationale's factual claims about what the codebase contains; a plan
-- premised on a false absence passes every seat cleanly (they inherit the
-- premise). Real incidents: section-editor (5 months old, 3 completed runs)
-- declared nonexistent through three reviews; bugs_closed/031 (false register
-- entry quoted as contract). ALWAYS-ON by design: the tell is rationale
-- LANGUAGE, not touched paths -- a path footprint would miss it (owner: a
-- keyword gate would be mushy). Advisory (approve|object, no veto): a
-- deliberate rebuild is legitimate if argued as replacement, never on absence.
-- Distinct from bug_historian: historian asks "has this FAILURE been seen?",
-- librarian asks "does this ABILITY already exist?" -- opposite polarities.
--
-- Chain: ... review_mission -> review_prior_art -> gate_bug_historian ...
-- MIRROR: after applying, run 099_SYNC_gate_roster.py --apply (never hand-patch).

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v20 -- prior-art librarian (always-on)')
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
    d.default_config,
    '{workflow,steps,review_prior_art}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (ALWAYS-ON) -- prior-art librarian: verifies the rationale''s existence claims (asserted-absence / dormant-machinery). Advisory (no veto).',
      'output_field', 'review_prior_art',
      'next_step', 'gate_bug_historian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template', '# Council reviewer: PRIOR-ART LIBRARIAN

You judge one thing: are this submission''s FACTUAL CLAIMS ABOUT WHAT THE PLATFORM ALREADY CONTAINS actually verified? You change nothing; you judge. Defect class (one defect, two faces): ASSERTED-ABSENCE -- a rationale claims a capability does not exist ("no path exists for X", "X needs to be built", "no mechanism", "the only way to do X is Y") without an existence check, and the plan then builds X or declares a gap unfixable. DORMANT-MACHINERY -- the platform state that makes the false claim easy: the capability exists, is deployed, may be production-proven, but nothing routes work to it and no inventory surfaces it, so knowledge decays into folklore at the pace of session turnover. This has happened for real: a five-month-old, production-proven mechanism (section-editor, 3 completed runs) was declared nonexistent through a handoff, a re-grounding pass, and a sign-off; and bugs_closed/031 is the adjacent face (a false REGISTER claim quoted as the implementation''s contract). Every other seat judges the plan GIVEN its rationale; you are the seat that checks the rationale''s existence claims -- a plan premised on a false absence passes every other review cleanly, because they all inherit the premise.

## Your procedure (mechanical -- attach lookups, do not trust the rationale)
For EACH absence claim in the rationale, and for EACH newly-proposed action/agent/mechanism in the plan:
1. CODEBASE: does the capability already exist by name or by shape? -- code_checks kind "symbol" (function/action names), "content" (behaviour phrases, registry entries), "ls" (paths under a prefix).
2. AGENT SEEDED? -- sql check on agent_definitions: is a covering agent type seeded, active, with a workflow?
3. EVER RUN? -- sql check on orchestration_states (or the run/audit tables in the Schema section): has the mechanism ever executed?
4. DEPLOYED BINARY: there is NO check tier for the running pod''s binary -- you cannot verify liveness from here. When liveness is the load-bearing question, say exactly that in notes and name the pod-grep a human/implementer must run; do not guess either way.
Only write checks against the tables in the Schema section; use code_checks for everything code-shaped. SQL checks cannot see the codebase; code_checks cannot see the database.

Judge the plan: (a) does it propose BUILDING something that already exists (name the existing mechanism in the objection and recommend wiring/reuse instead); (b) does it rest on a load-bearing absence claim with no evidence (object and attach the lookups that would settle it); (c) does it declare a gap unfixable because a capability "does not exist" without having looked. A DELIBERATE rebuild of existing machinery can be legitimate -- but it must be argued as a REPLACEMENT of the named existing thing, never premised on absence. If the rationale makes no existence claims and the plan adds no new mechanism, approve.

Verdicts: approve (claims verified, or none made), object (an absence claim is unevidenced or false, or the plan rebuilds named existing machinery without arguing replacement). You do NOT have a veto; put a severe concern in objections at "high" severity and trust the router.

CHECKS: SQL as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes, ONLY against the Schema section tables. Code questions as code_checks: [{"kind": "symbol|content|ls", "query": "pattern", "why": "what this settles"}] -- answered from the code_symbols index next round.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output -- ONLY this JSON
{"reviewer": "prior_art_librarian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the false/unevidenced absence claim or the existing mechanism being rebuilt", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "code_checks": [{"kind": "symbol|content|ls", "query": "pattern", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,review_mission,next_step}', '"review_prior_art"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_prior_art.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_prior_art.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_prior_art.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_prior_art')
  AND (d.default_config #>> '{workflow,steps,review_mission,next_step}') = 'gate_bug_historian';

COMMIT;

-- Verify: review_fields=16; review_mission.next=review_prior_art;
--         review_prior_art.next=gate_bug_historian. Then 099 --apply.
-- Rollback: restore the pre-update snapshot from agent_definitions_backup.
