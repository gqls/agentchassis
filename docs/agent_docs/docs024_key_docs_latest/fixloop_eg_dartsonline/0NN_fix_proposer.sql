-- 0NN_fix_proposer.sql — F1.1a + F1.1b(b) + F2.1 of the diagnosis→fix loop.
-- 2026-07-10 (v2: council wired in; proposer input widened; max_tokens inside ai_service). Renumber 0NN when filing. Applies to clients_db.
--
-- Creates: (1) diagnosis_artifacts gains kind='fix_plan' (+ iteration 0 for
--              run-level artifacts);
--          (2) agent_definitions row `fix-proposer` — a workflow that turns a
--              CONFIRMED diagnosis into a CONSTRAINED EDIT PLAN, persisted as
--              an artifact. It writes NO code: the git branch + PR is F1.1b,
--              behind the isolated write token (Q-C). An agent whose only
--              write surface is its own artifacts table needs no token yet.
--
-- The CONFIRMED gate is load-bearing: F1 was deliberately deferred until the
-- verdict guards (tier, symptom-closure, citation-backed coverage) made
-- CONFIRMED trustworthy — runs 1 and 2 of the benchmark produced CONFIRMED
-- verdicts a fixer must never have acted on. The workflow refuses anything
-- whose diagnosis status is not CONFIRMED.

BEGIN;

-- ── 1. artifacts table: new kind + run-level iteration 0 ─────────────────────
ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_kind_check;
ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_kind_check
    CHECK (kind IN ('bundle', 'iteration_note', 'fix_plan', 'council_report'));
ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_iteration_check;
ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_iteration_check
    CHECK (iteration >= 0);
COMMENT ON COLUMN diagnosis_artifacts.iteration IS
    '1-based loop iteration for bundle/iteration_note; 0 = a run-level artifact (fix_plan). Derived in assemble as route.diagnose_state.iteration + 1.';

-- ── 2. the fix-proposer agent ────────────────────────────────────────────────
-- Snapshot first if a live row exists (idempotent re-apply path).
SELECT snapshot_agent('fix-proposer', 'pre-update: F1.1a re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'fix-proposer',
    'Fix Proposer (F1.1a)',
    'Turns a CONFIRMED diagnosis (by correlation_id) into a constrained edit plan + a council review (edit-quality + guardian reviewers, deterministic decision), all persisted to diagnosis_artifacts. Writes no code; refuses non-CONFIRMED diagnoses.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "fix-planning"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_diagnosis',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 900,
      'steps', jsonb_build_object(

        'load_diagnosis', jsonb_build_object(
          'action', 'query_database',
          'description', 'Pull the diagnosis (status, conclusion incl. symptom coverage) for the given correlation_id.',
          'output_field', 'diagnosis_row',
          'next_step', 'check_confirmed',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id'),
            'query',
              'SELECT collected_data->''diagnosis''->>''status'' AS status, ' ||
              '       collected_data->''diagnosis''->>''conclusion'' AS conclusion ' ||
              'FROM orchestration_states ' ||
              'WHERE correlation_id = $1::uuid AND collected_data ? ''diagnosis'' ' ||
              'ORDER BY created_at DESC LIMIT 1'
          )
        ),

        -- THE GATE: only a CONFIRMED diagnosis may seed a fix plan.
        'check_confirmed', jsonb_build_object(
          'action', 'conditional',
          'description', 'Refuse anything not CONFIRMED — the whole reason F1 waited for the verdict guards.',
          'config', jsonb_build_object(
            'condition', 'diagnosis_row.status == ''CONFIRMED''',
            'then_step', 'load_last_bundle',
            'else_step', 'complete_refused'
          )
        ),

        'load_last_bundle', jsonb_build_object(
          'action', 'query_database',
          'description', 'The final iteration''s evidence bundle, for grounding the plan.',
          'output_field', 'last_bundle',
          'next_step', 'propose',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id'),
            'query',
              'SELECT string_agg(body, chr(10) || ''=== earlier iteration bundle ==='' || chr(10) ORDER BY iteration DESC) AS body ' ||
              'FROM (SELECT body, iteration FROM diagnosis_artifacts ' ||
              '      WHERE correlation_id = $1 AND kind = ''bundle'' ' ||
              '      ORDER BY iteration DESC LIMIT 2) last_two'
          )
        ),

        'propose', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Draft the constrained edit plan from the diagnosis + final bundle.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 8000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'last_bundle'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — constrained fix plan (F1.1a)' || chr(10) || chr(10) ||
'You are drafting a CONSTRAINED EDIT PLAN from a CONFIRMED diagnosis. You do NOT write code, patches, or diffs to apply — you name the smallest set of edits a human (and a later automated slice) can review, each grounded in the diagnosis evidence.' || chr(10) || chr(10) ||
'## Hard rules' || chr(10) ||
'1. PLATFORM, not site data: fix the mechanism in code/workflow definitions, never one site''s rows (owner ruling, 2026-07-09).' || chr(10) ||
'2. MINIMAL: the fewest edits that remove the confirmed cause. If you need more than a handful, the fix is architecture change — say so in risks and keep the plan to the safe core.' || chr(10) ||
'3. GROUNDED: every edit''s rationale must trace to the diagnosis conclusion or the bundle; quote the evidence in grounded_in.' || chr(10) ||
'4. NO new dependencies, no schema DDL, no deletes of files.' || chr(10) ||
'5. Respect surface ownership: an edit to a workflow JSON in agent_definitions is operation "config_change" and must say so.' || chr(10) ||
'6. COVER EVERY MECHANISM the diagnosis cites: a workflow step quoted in the citations (e.g. a success-labelled error terminal) and any generation code cited must each have a covering edit or an explicit line in risks saying why not.' || chr(10) ||
'7. Every edit CHANGES something. "No code change required", audits, and comment-only edits are invalid and will be rejected by validation — put observations in risks, not edits.' || chr(10) || chr(10) ||
'## The confirmed diagnosis' || chr(10) || chr(10) ||
'{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## Final evidence bundle' || chr(10) || chr(10) ||
'{{.last_bundle.body}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) || chr(10) ||
'```json' || chr(10) ||
'{' || chr(10) ||
'  "summary": "one paragraph: what is broken and what the plan changes",' || chr(10) ||
'  "edits": [' || chr(10) ||
'    {"file": "repo-relative/path.go", "symbol": "FunctionOrStep", "operation": "modify|add|remove|config_change",' || chr(10) ||
'     "rationale": "why THIS edit, tracing to the diagnosis", "sketch": "the intended change, described precisely"}' || chr(10) ||
'  ],' || chr(10) ||
'  "grounded_in": ["verbatim quotes from the diagnosis/bundle this plan rests on"],' || chr(10) ||
'  "risks": "what could this break; what a reviewer should check"' || chr(10) ||
'}' || chr(10) ||
'```'
          )
        ),

        'persist_plan', jsonb_build_object(
          'action', 'diagnose_persist_fix_plan',
          'description', 'Structural validation + write to diagnosis_artifacts (kind=fix_plan). A failed validation FAILS the run.',
          'output_field', 'plan_persisted',
          'next_step', 'review_editquality',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'proposal.result'
          )
        ),

        'review_editquality', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1 — edit quality: real changes, minimality, right causal path, missing mechanisms.',
          'output_field', 'review_editquality',
          'next_step', 'review_guardian',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: EDIT QUALITY' || chr(10) || chr(10) ||
'You review a proposed fix plan against its diagnosis. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) does every edit CHANGE something real (audits/comments are not edits); (b) does the plan address every mechanism the diagnosis cites — quote any cited mechanism with no covering edit into missing; (c) does each edit target the causal path the diagnosis established, not an adjacent one; (d) is the plan minimal.' || chr(10) || chr(10) ||
'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: fixes a different bug, or all edits are no-ops).' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "editquality", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": ["cited mechanism with no covering edit"], "notes": "..."}'
          )
        ),

        'review_guardian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 2 — pipeline guardian: surface ownership, blast radius, architecture-change signals. HARD VETO holder.',
          'output_field', 'review_guardian',
          'next_step', 'council_decide',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: PIPELINE GUARDIAN (hard-veto holder)' || chr(10) || chr(10) ||
'You protect the platform''s other pipelines from collateral damage. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) blast radius — which pipelines/workflows consume each edited file or workflow step; does the plan acknowledge them; (b) architecture-change signals — edits to shared contracts, wire formats, message shapes, exported signatures, or MANY packages at once mean this is not a constrained fix: veto and say it needs an architecture review; (c) surface ownership — workflow-JSON edits must be operation config_change and name the owning pipeline.' || chr(10) || chr(10) ||
'Verdicts: approve, object (containable concerns), veto (cross-pipeline damage or architecture change dressed as a fix). Your veto BLOCKS.' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "guardian", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "notes": "..."}'
          )
        ),

        'council_decide', jsonb_build_object(
          'action', 'diagnose_council_decide',
          'description', 'Deterministic aggregation: veto→rejected, object→revise, else approved. Guardian holds the hard veto. Persists kind=council_report.',
          'output_field', 'council',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_guardian.result'),
            'hard_veto_from', jsonb_build_array('guardian')
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Plan + council report persisted; fetch both from diagnosis_artifacts by correlation_id.',
          'config', jsonb_build_object('output_fields', jsonb_build_array('plan_persisted', 'council'))
        ),

        'complete_refused', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No plan: diagnosis not CONFIRMED, proposer failed, or plan failed validation.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('diagnosis_row'),
            'success_message', 'fix-proposer made no plan: requires a CONFIRMED diagnosis and a valid constrained edit plan'
          )
        )
      )
    ))
FROM agent_definitions d
WHERE d.type = 'diagnose-orchestrator'
  AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
ON CONFLICT (type, version) DO UPDATE
   SET default_config = EXCLUDED.default_config,
       description    = EXCLUDED.description,
       updated_at     = now();

COMMIT;

-- Rollback (manual):
--   DELETE FROM agent_definitions WHERE type='fix-proposer' AND version=1;
--   ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_kind_check;
--   ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_kind_check
--       CHECK (kind IN ('bundle','iteration_note'));
--   (leave iteration_check at >= 0; 0-rows only exist if a fix_plan was written)
