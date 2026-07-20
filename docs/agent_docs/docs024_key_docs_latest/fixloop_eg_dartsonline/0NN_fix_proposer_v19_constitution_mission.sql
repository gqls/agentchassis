-- v19 (2026-07-20): adds TWO ALWAYS-ON council reviewers -- review_constitution
-- and review_mission -- to fix-proposer, SURGICALLY (chained jsonb_set on the live
-- config). Owner directive 2026-07-20: the platform has a constitution and a
-- mission that DECLARE themselves the fixed direction but were enforced by nobody;
-- these two seats make the council hold every fix to them. Always-on (NOT gated):
-- the constitution's own rules are "always-on, true regardless of task", and the
-- mission is "the frame every decision sits inside" -- gating either would
-- contradict what they are. review_constitution carries the root-cause /
-- anti-workaround rule as its headline (owner's first point).
--
-- Chain head becomes:
--   ... select_panel -> review_editquality -> review_constitution
--       -> review_mission -> gate_bug_historian -> [gated specialists]
--       -> review_guardian -> council_decide
--
-- MIRROR: after this applies, run 099_SYNC_gate_roster.py --apply to mirror both
-- seats onto council-gate (do NOT hand-patch the gate). The mirror copies every
-- review_* step verbatim and re-asserts editquality.next_step, so this is
-- mirror-safe.

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: v19 -- constitution + mission always-on seats')
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
    d.default_config,
    '{workflow,steps,review_constitution}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (ALWAYS-ON) -- constitution: does the fix uphold the always-on constitutional rules, root-cause-not-workaround first? Advisory (no veto).',
      'output_field', 'review_constitution',
      'next_step', 'review_mission',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template', '# Council reviewer: CONSTITUTION

You judge one thing: does this fix uphold the platform CONSTITUTION -- the always-on rules that hold true regardless of task? You change nothing; you judge. These are not this seat''s opinions; they are the hand-authored constitution (thin_slice_constitution.md; CTS-029/DEV-054), pasted into every task bundle and, until now, enforced by nobody.

## The always-on constitutional rules
- FIX THE CAUSE, NOT THE SYMPTOM. Prefer the fix that addresses the structural cause over the quick patch, even when the patch is faster. A fix must not work AROUND a defect -- adding a guard, a special-case, an avoidance, or routing past a known/listed bug so the symptom stops -- when the real cause is fixable. If a structural fix is being knowingly deferred, that deferral must be STATED and justified in the plan, never silent. A fix whose own rationale points at another bug it is stepping around, or names a symptom rather than a mechanism, fails this rule.
- REUSE BEFORE RECREATE. Before a new function, struct, component or action, an existing one that does the same or similar must be looked for and improved/altered. Recreating something the system already has is a defect, not a shortcut. If the plan adds new machinery, it must show it is not duplicating machinery that already exists.
- SCHEMA FIRST, PARAMETERISED ALWAYS. SQL is written against the real inspected schema, never an assumed shape; values pass as query parameters, never interpolated into a template string.
- NO SILENT RENAMES / NO NAME DRIFT. Variable, field and workflow-key names stay stable; a deliberate rename is stated. A workflow variable name must match the name the action reads.
- PLAIN, PRAGMATIC TONE. No hype or filler -- this governs generated content AND the commit message / plan rationale, not just chat.

(Detailed task-specific contracts -- wrapper-orchestrator, work-item dedup, input_contract, schema-source tiers -- are the guidelines seat''s beat. You own the ALWAYS-ON principles above, the root-cause rule first.)

Judge the plan: (a) does any edit patch a symptom or work around a known/listed bug instead of fixing the diagnosed cause, with no explicit justified deferral; (b) does it recreate machinery that already exists instead of reusing it; (c) does it break schema-first / parameterised-query / no-silent-rename / tone. If the fix addresses the real cause, reuses what exists, and engages none of these rules wrongly, approve.

Verdicts: approve (upholds the constitution, or does not engage these rules), object (name the rule broken -- root-cause / workaround first -- in objections). You do NOT have a veto; put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.

CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (does the "new" function already exist under another name; is the worked-around bug a real listed row), put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output -- ONLY this JSON
{"reviewer": "constitution", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the constitutional rule broken; root-cause/workaround first", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,review_mission}',
    jsonb_build_object(
      'action', 'execute_llm_prompt',
      'description', 'Council reviewer (ALWAYS-ON) -- mission: does the change serve, or at least not work against, the platform mission (best site per domain)? Advisory (no veto).',
      'output_field', 'review_mission',
      'next_step', 'gate_bug_historian',
      'config', jsonb_build_object(
        'error_step', 'complete_refused',
        'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
        'temperature', 0.0,
        'input_fields', jsonb_build_array('diagnosis_row','plan_persisted','schema_hint'),
        'output_format', 'json',
        'prompt_template', '# Council reviewer: MISSION

You judge one thing: does this change SERVE, or at least not work AGAINST, the platform MISSION? You change nothing; you judge. The mission (BIZ-001; doc 028_platform_mission_and_pipeline_direction) is the frame every architectural decision sits inside -- the doc says so itself: "when behaviour drifts, we check against this document."

## The mission
- BEST SITE PER DOMAIN, ONE PIPELINE. Given any domain -- blank, live, adopted, operator-briefed -- produce the best possible website end to end with minimal human input. "Best" = most useful to real visitors (measured by engagement) AND the best revenue via whatever model genuinely fits the domain. All inputs travel the SAME agent graph; what differs is evidence weight and the fidelity dial, not the pipeline.
- THE REVENUE MODEL SHAPES THE SITE, not the other way round. Consultancy, ad-supported tools, affiliate comparison, SaaS, publication -- each is a legitimate, completely different site. Defaulting to a consultancy shape when the signal is absent is a FAILURE MODE, not a safe fallback. Mixing shapes (a tools site with a "Start a Project" CTA) signals the classification is vague or a downstream agent is ignoring it.
- THE CLASSIFIER IS THE STRATEGIC BRAIN and always runs in full. Adoption and an operator mission WEIGHT its reasoning; they never SHORTCUT it. Its outputs are read downstream as DIRECTION, not suggestion.
- SILENT OVERRIDE IS THE FAILURE MODE WE ARE ELIMINATING. When a downstream agent (planner, composition, content, design) cannot implement a spec item, it builds what it can, MARKS the rest, and SURFACES the gap -- it never substitutes its own preference. Adoption is an INPUT ("what the site IS"), never a replacement for direction ("what it should become").

Most bug fixes are mission-neutral -- they restore intended behaviour -- and those APPROVE. Object only when a change actively pushes the platform AWAY from the mission.

Judge the plan: (a) does any edit let a downstream agent silently override, trim, or shortcut the classifier''s direction instead of marking + surfacing the gap; (b) does it bake in or default to one revenue/site shape (especially the consultancy default) where the mission says the shape must fit the domain; (c) does it treat adoption or an operator brief as a bypass of the full classifier run; (d) does it otherwise move the platform away from "best possible site per domain, minimal human input." If the change is mission-neutral or serves the mission, approve.

Verdicts: approve (serves or is neutral to the mission), object (name the mission principle it works against). You do NOT have a veto; put a severe concern in objections at "high" severity and trust the router; note a true direction-level concern explicitly.

CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output -- ONLY this JSON
{"reviewer": "mission", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the mission principle worked against", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
      )
    )
  ),
    '{workflow,steps,review_editquality,next_step}', '"review_constitution"'::jsonb
  ),
    '{workflow,steps,council_decide,config,review_fields}',
    (d.default_config #> '{workflow,steps,council_decide,config,review_fields}') || '["review_constitution.result","review_mission.result"]'::jsonb
  ),
    '{workflow,steps,escalate,config,review_fields}',
    (d.default_config #> '{workflow,steps,escalate,config,review_fields}') || '["review_constitution.result","review_mission.result"]'::jsonb
  ),
    '{workflow,steps,run_checks,config,check_fields}',
    (d.default_config #> '{workflow,steps,run_checks,config,check_fields}') || '["review_constitution.result.checks","review_mission.result.checks"]'::jsonb
  ),
  updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND NOT (d.default_config->'workflow'->'steps' ? 'review_constitution');

COMMIT;

-- Rollback (manual): restore the pre-update snapshot from agent_definitions_backup.
