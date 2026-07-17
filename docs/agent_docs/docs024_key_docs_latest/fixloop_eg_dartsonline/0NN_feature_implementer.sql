-- 0NN_feature_implementer.sql — feature builder delta 2: the stage-loop implementer (v1 DRAFT).
-- 2026-07-17, "fixloop feature builder" thread. Renumber 0NN when filing. Applies to clients_db.
--
-- ██ STATUS: DRAFT, NOT APPLIED ██ — ships in the repo per the seed discipline.
-- Apply ONLY AFTER the chassis image carrying feature_stage_route + the delta-2
-- seams (read ref_field; prepare branch/message/expected-symbols fields; build
-- gate test_packages_field; feature-implementer on the isRepoCloningAgent spawn
-- gate) is live — commit c19b5d097 or later. The git-adapter needs nothing new
-- (same three verbs); rbac is the fix-implementer's (same spawn mechanism, same
-- service account for the gate Jobs).
--
-- WHAT THIS AGENT IS (design: DESIGN_stage_loop_delta2.md, E1-E5 approved
-- 2026-07-17; E1: a SEPARATE agent — fix-implementer stays frozen on single
-- plans). The feature builder's WRITE STEP: turns a council-APPROVED staged
-- plan (plan_format staged-v1) into feat/<short-corr> with ONE commit per
-- stage, a build gate per stage, a derived go-test END gate (D6), and ONE PR
-- whose body carries the owner's post-merge checklist as a task list. The
-- cage is the fix loop's, unchanged: dedicated pod via the orchestrator
-- (read token only), writes only via git-adapter, hard PER-STAGE allowlist
-- (to-be-created files enter via the stage's own add edits), PR is the human
-- terminal. E4: a pre-existing feat/* branch is a HARD refusal at seed time —
-- never silently reused.
--
-- THE LOOP. feature_stage_route emits each stage as a SINGLE-PLAN shape
-- (stage.stage_plan), so read_current_files and prepare run their proven
-- single-plan logic per stage. Stage 1 reads the base ref; stage N>1 reads the
-- feat/* branch (it must see earlier stages' commits). A stage whose edits are
-- all config_change (no repo files) skips read/commit/gate entirely — its
-- content rides in the PR body. A red stage gate ends the run: branch + log
-- left for inspection, NO PR, earlier stages' commits intact.
--
-- Input: {"fix_correlation_id": "<correlation of an APPROVED staged plan>",
--         "base_ref": "<optional; default main>"}
-- Refuses: council not approved; plan not staged-v1 (single plans belong to
-- fix-implementer); stale feat/* branch.

BEGIN;

SELECT snapshot_agent('feature-implementer', 'pre-update: feature builder delta 2 v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='feature-implementer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'feature-implementer',
    'Feature Implementer (FB delta 2)',
    'Turns a council-APPROVED staged plan (by correlation_id) into feat/<short-corr>: one commit per stage via the git-adapter (hard per-stage allowlist incl. to-be-created files, expected-symbols check), build gate per stage, derived go-test end gate, ONE PR carrying the owner''s post-merge checklist. Read token only; refuses non-approved/non-staged plans and stale feat/* branches; PR is the human terminal — never merges.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "feature-implementation"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_plan',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 3600,
      'steps', jsonb_build_object(

        'load_plan', jsonb_build_object(
          'action', 'query_database',
          'description', 'Latest staged plan for the correlation (in an approved run, the approved one).',
          'output_field', 'plan_row',
          'next_step', 'load_council',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id'),
            'query',
              'SELECT body FROM diagnosis_artifacts ' ||
              'WHERE correlation_id = $1 AND kind = ''fix_plan'' ' ||
              'ORDER BY created_at DESC LIMIT 1'
          )
        ),

        'load_council', jsonb_build_object(
          'action', 'query_database',
          'description', 'Latest council report + its raw decision — the gate input.',
          'output_field', 'council_row',
          'next_step', 'check_approved',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id'),
            'query',
              'SELECT body, metadata->>''decision'' AS decision FROM diagnosis_artifacts ' ||
              'WHERE correlation_id = $1 AND kind = ''council_report'' ' ||
              'ORDER BY created_at DESC LIMIT 1'
          )
        ),

        -- THE GATE: only a council-APPROVED plan may reach the write surface.
        -- (feature_stage_route additionally refuses non-staged-v1 plans.)
        'check_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Mirror of the designer''s spec gate: refuse anything the council did not approve.',
          'config', jsonb_build_object(
            'condition', 'council_row.decision == ''approved''',
            'then_step', 'route_init',
            'else_step', 'complete_refused'
          )
        ),

        -- Seed the loop: parses/validates the staged plan, derives feat/<short>,
        -- resolves base_ref (input_data.base_ref → main), enforces E4 freshness,
        -- emits stage 1.
        'route_init', jsonb_build_object(
          'action', 'feature_stage_route',
          'description', 'Stage-loop seed: E4 branch freshness + stage 1 emission (reads the BASE ref).',
          'output_field', 'stage',
          'next_step', 'create_branch',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'plan_row.body',
            'council_field', 'council_row.body'
          )
        ),

        'create_branch', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'feat/<short-corr> from the base ref, via the adapter — ONCE for the whole feature.',
          'output_field', 'branch_result',
          'next_step', 'check_has_edits',
          'config', jsonb_build_object(
            'adapter_action', 'create_branch',
            'data_literals', jsonb_build_object('repo_name', 'agentchassis'),
            'data_fields', jsonb_build_object('branch', 'stage.branch', 'from_branch', 'stage.base_ref')
          )
        ),

        -- ── the per-stage loop ──────────────────────────────────────────────
        'check_has_edits', jsonb_build_object(
          'action', 'conditional',
          'description', 'A config_change-only stage has no repo files: skip read/commit/gate, its content rides in the PR body.',
          'config', jsonb_build_object(
            'condition', 'stage.has_repo_edits == true',
            'then_step', 'stage_read',
            'else_step', 'stage_advance'
          )
        ),

        'stage_read', jsonb_build_object(
          'action', 'diagnose_read_repo_files',
          'description', 'Current bodies of THIS stage''s modify files at the routed ref (base for stage 1, the feat branch after — later stages see earlier commits). add files are expected absent.',
          'output_field', 'current_files',
          'next_step', 'stage_implement',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'plan_field', 'stage.stage_plan',
            'ref_field', 'stage.read_ref',
            'repo_owner', 'gqls',
            'repo_name', 'agentchassis'
          )
        ),

        'stage_implement', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Whole-file implementation of THIS stage only — the proven sketch_to_files contract, stage-scoped.',
          'output_field', 'implementation',
          'next_step', 'stage_prepare',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              -- Whole-file output: 16000 truncated a 41KB file on the fix
              -- loop's first live run; 32000 proven since (parity).
              'max_tokens', 32000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('stage', 'current_files'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — implement ONE STAGE of an approved staged plan (whole files)' || chr(10) || chr(10) ||
'A review council APPROVED this staged plan; you are implementing stage {{.stage.stage_index}} of {{.stage.stages_total}} ONLY. You output the COMPLETE NEW BODY of every modify/add file THIS STAGE names — and nothing else.' || chr(10) || chr(10) ||
'## Hard rules' || chr(10) ||
'1. ONLY this stage''s modify/add files. Any other file will be rejected by a deterministic per-stage allowlist after you.' || chr(10) ||
'2. Implement each edit EXACTLY as its sketch describes. If a sketch is ambiguous, choose the minimal reading and say so in notes.' || chr(10) ||
'3. NO drive-by changes: no reformatting untouched code, no renamed variables, no comment rewrites, no import reordering beyond what the change forces.' || chr(10) ||
'4. Match each file''s existing style: comment density, naming, error wording.' || chr(10) ||
'5. Output the WHOLE file — your output replaces the file byte-for-byte. Never elide with "... rest unchanged". For an add file (absent from current bodies), write the complete new file.' || chr(10) ||
'6. config_change edits are NOT yours — skip them (the PR body carries them for a human).' || chr(10) ||
'7. The stage''s expected symbols MUST appear in your output — they are checked deterministically: {{.stage.expected_symbols}}' || chr(10) || chr(10) ||
'## This stage''s goal' || chr(10) || '{{.stage.stage_goal}}' || chr(10) || chr(10) ||
'## What earlier stages already committed to this branch' || chr(10) || '{{.stage.prior_stages}}' || chr(10) || chr(10) ||
'## This stage''s plan' || chr(10) || '{{.stage.stage_plan}}' || chr(10) || chr(10) ||
'## Current file bodies at {{.stage.read_ref}} (your input — rewrite these; add files are absent by design)' || chr(10) || '{{.current_files.rendered}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"files": [{"path": "repo-relative/path", "content": "the complete new file body"}], "notes": "anything the reviewer must know"}'
          )
        ),

        'stage_prepare', jsonb_build_object(
          'action', 'diagnose_prepare_fix_commit',
          'description', 'THE PER-STAGE ALLOWLIST (deterministic): out-of-stage / incomplete / empty / no-op reject; expected symbols enforced; routed branch + per-stage commit message.',
          'output_field', 'commit_prep',
          'next_step', 'stage_commit',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'stage.stage_plan',
            'files_field', 'implementation.result',
            'originals_field', 'current_files.files',
            'council_field', 'council_row.body',
            'branch_field', 'stage.branch',
            'commit_message_field', 'stage.commit_message',
            'expected_symbols_field', 'stage.expected_symbols',
            'repo_name', 'agentchassis',
            'base_branch', 'main'
          )
        ),

        'stage_commit', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'Commit THIS stage''s validated files to the feat branch (repo-relative — empty domain).',
          'output_field', 'commit_result',
          'next_step', 'check_build_needed',
          'config', jsonb_build_object(
            'adapter_action', 'commit',
            'data_literals', jsonb_build_object('repo_name', 'agentchassis', 'domain', ''),
            'data_fields', jsonb_build_object(
              'files', 'commit_prep.files',
              'commit_message', 'commit_prep.commit_message',
              'branch', 'commit_prep.branch'
            )
          )
        ),

        'check_build_needed', jsonb_build_object(
          'action', 'conditional',
          'description', 'gate.build=false stages (all seed/doc — validated upstream) skip the build gate.',
          'config', jsonb_build_object(
            'condition', 'stage.gate_build == true',
            'then_step', 'stage_gate',
            'else_step', 'stage_advance'
          )
        ),

        'stage_gate', jsonb_build_object(
          'action', 'diagnose_build_gate',
          'description', 'Per-stage gate: gofmt (this stage''s files) + targeted go build in a golang k8s Job.',
          'output_field', 'gate',
          'next_step', 'check_gate',
          'config', jsonb_build_object(
            'branch', 'commit_prep.branch',
            'changed_files_field', 'commit_prep.files',
            'repo_owner', 'gqls',
            'repo_name', 'agentchassis',
            'timeout_seconds', 600
          )
        ),

        'check_gate', jsonb_build_object(
          'action', 'conditional',
          'description', 'Green → next stage. Red → NO PR; branch + earlier commits + log left for human inspection.',
          'config', jsonb_build_object(
            'condition', 'gate.passed == true',
            'then_step', 'stage_advance',
            'else_step', 'complete_gate_failed'
          )
        ),

        'stage_advance', jsonb_build_object(
          'action', 'feature_stage_route',
          'description', 'Mark the stage done; emit the next stage — or the terminal PR payload + derived test packages.',
          'output_field', 'stage',
          'next_step', 'check_more',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'plan_row.body',
            'council_field', 'council_row.body'
          )
        ),

        'check_more', jsonb_build_object(
          'action', 'conditional',
          'description', 'More stages → loop; exhausted → the end test gate.',
          'config', jsonb_build_object(
            'condition', 'stage.stage_ready == true',
            'then_step', 'check_has_edits',
            'else_step', 'test_gate'
          )
        ),

        -- ── the END gate (D6): go test over the derived packages ────────────
        'test_gate', jsonb_build_object(
          'action', 'diagnose_build_gate',
          'description', 'End gate: targeted go build + go test over the packages DERIVED from the plan''s .go edits (features add behaviour; build-only is not enough). No gofmt (per-stage gates covered it).',
          'output_field', 'gate',
          'next_step', 'check_tests',
          'config', jsonb_build_object(
            'branch', 'stage.branch',
            'changed_files_field', '',
            'test_packages_field', 'stage.test_packages',
            'repo_owner', 'gqls',
            'repo_name', 'agentchassis',
            'timeout_seconds', 900
          )
        ),

        'check_tests', jsonb_build_object(
          'action', 'conditional',
          'description', 'Tests green → the ONE PR. Red → no PR; branch + log are the hand-off.',
          'config', jsonb_build_object(
            'condition', 'gate.passed == true',
            'then_step', 'create_pr',
            'else_step', 'complete_tests_failed'
          )
        ),

        'create_pr', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'The human terminal: ONE PR for the whole feature; body carries stages, config_changes, the owner''s post-merge checklist (task list), and the council decision. Created, never merged.',
          'output_field', 'pr_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'adapter_action', 'create_pull_request',
            'data_literals', jsonb_build_object('repo_name', 'agentchassis'),
            'data_fields', jsonb_build_object(
              'title', 'stage.pr_title',
              'body', 'stage.pr_body',
              'head', 'stage.branch',
              'base', 'stage.base_ref'
            )
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'PR opened. The feature is NOT done: the PR body''s post-merge checklist (image THEN seed) is the owner''s, deliberately.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('pr_result', 'stage', 'gate'),
            'success_message', 'feature-implementer opened ONE PR for the staged build — human review terminal; merge + checklist are owner acts'
          )
        ),

        'complete_gate_failed', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'A stage''s build gate went RED: no PR. Earlier stages'' commits + the log are the hand-off.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('gate', 'stage', 'commit_prep'),
            'success_message', 'stage build gate failed — NO PR; feat branch left with earlier stages'' commits, build log in gate.log'
          )
        ),

        'complete_tests_failed', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'The END test gate went RED: no PR. All stage commits + the test log are the hand-off.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('gate', 'stage'),
            'success_message', 'end test gate failed — NO PR; feat branch left complete, test log in gate.log'
          )
        ),

        'complete_refused', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No write: plan not approved / not staged-v1, stale feat/* branch (E4), files unreadable, implementation failed a per-stage allowlist, or LLM refusal.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('council_row'),
            'success_message', 'feature-implementer wrote nothing: requires a council-APPROVED staged-v1 plan, a fresh feat/* branch, and valid per-stage implementations'
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

-- Rollback: restore the pre-update snapshot from agent_definitions_backup, or
-- DELETE FROM agent_definitions WHERE type='feature-implementer' AND version=1;
