-- 0NN_fix_implementer.sql — F1.1b(c) part 2c: the fix-implementer agent (v1).
-- 2026-07-12. Renumber 0NN when filing. Applies to clients_db.
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER:
--   1. chassis image carrying: diagnose_prepare_fix_commit, diagnose_build_gate,
--      diagnose_read_repo_files, git_adapter_request, and fix-implementer on the
--      isRepoCloningAgent spawn gate (all > v1.0.1108);
--   2. git-adapter image carrying: create_branch, create_pull_request,
--      branch-aware commit;
--   3. kubectl apply of the agent-chassis rbac-job-spawner.yaml (adds pods/log
--      get — without it the build gate works but failure logs are lost).
-- A workflow naming unregistered actions fails at runtime.
--
-- WHAT THIS AGENT IS. The loop's WRITE STEP: turns a council-APPROVED plan into
-- a fix/* branch + a pull request. Constraints, all owner-ruled:
--   * the write credential lives in the git-adapter — this agent never holds
--     one (it carries only GITHUB_READ_TOKEN, to read current file bodies);
--   * the approved plan's file list is a HARD allowlist, enforced
--     deterministically (diagnose_prepare_fix_commit);
--   * the fix branch must survive gofmt + targeted go build in a golang k8s
--     Job BEFORE a PR is opened ("no PRs for broken code") — a red gate ends
--     the run with the build log, branch left for inspection, NO PR;
--   * the PR is the HUMAN TERMINAL: created, never merged by the platform.
--     Isolation model (owner, 2026-07-12): fix/* branches on THIS repo, owner
--     chooses what merges — no fork.
--
-- Input: {"fix_correlation_id": "<correlation of an APPROVED plan>"}
-- Refuses anything whose latest council decision is not 'approved' — the
-- mirror of the proposer's CONFIRMED gate.
--
-- v1 scope notes: the PR result lives in the run's collected_data (durable in
-- orchestration_states); a dedicated kind='fix_pr' artifact is deferred to
-- F1.2. config_change edits in a plan are NOT implemented by this agent (they
-- target agent_definitions rows, not the repo) — the PR body carries them for
-- the human.

BEGIN;

-- Snapshot first if a live row exists (idempotent re-apply path).
SELECT snapshot_agent('fix-implementer', 'pre-update: F1.1b(c) v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-implementer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'fix-implementer',
    'Fix Implementer (F1.1b(c))',
    'Turns a council-APPROVED fix plan (by correlation_id) into a fix/* branch + pull request via the git-adapter. Hard file allowlist; pre-PR gofmt+build gate in a golang k8s Job; holds only the read token; PR is the human terminal — never merges. Refuses non-approved plans.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "fix-implementation"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_plan',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 1800,
      'steps', jsonb_build_object(

        'load_plan', jsonb_build_object(
          'action', 'query_database',
          'description', 'Latest fix plan for the correlation (in an approved run, the approved one).',
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
          'next_step', 'load_diagnosis',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id'),
            'query',
              'SELECT body, metadata->>''decision'' AS decision FROM diagnosis_artifacts ' ||
              'WHERE correlation_id = $1 AND kind = ''council_report'' ' ||
              'ORDER BY created_at DESC LIMIT 1'
          )
        ),

        'load_diagnosis', jsonb_build_object(
          'action', 'query_database',
          'description', 'The diagnosis conclusion, for the PR body (Q-H package).',
          'output_field', 'diagnosis_row',
          'next_step', 'check_approved',
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

        -- THE GATE: only a council-APPROVED plan may reach the write surface.
        'check_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Mirror of the proposer''s CONFIRMED gate: refuse anything the council did not approve.',
          'config', jsonb_build_object(
            'condition', 'council_row.decision == ''approved''',
            'then_step', 'read_current_files',
            'else_step', 'complete_refused'
          )
        ),

        'read_current_files', jsonb_build_object(
          'action', 'diagnose_read_repo_files',
          'description', 'Current bodies of the plan''s modify/add files (contents API, read token) — whole-file rewrites need whole files.',
          'output_field', 'current_files',
          'next_step', 'sketch_to_files',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'plan_field', 'plan_row.body',
            'repo_owner', 'gqls',
            'repo_name', 'agentchassis',
            'ref', 'main'
          )
        ),

        'sketch_to_files', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Turn the approved plan''s sketches into COMPLETE new file bodies — only the plan''s files, no drive-by changes.',
          'output_field', 'implementation',
          'next_step', 'prepare',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 16000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_row', 'current_files'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — implement the approved fix plan (whole files)' || chr(10) || chr(10) ||
'A review council APPROVED this plan. Turn its sketches into code. You output the COMPLETE NEW BODY of every modify/add file the plan names — and nothing else.' || chr(10) || chr(10) ||
'## Hard rules' || chr(10) ||
'1. ONLY the plan''s modify/add files. Any other file will be rejected by a deterministic allowlist after you.' || chr(10) ||
'2. Implement each edit EXACTLY as its sketch describes. If a sketch is ambiguous, choose the minimal reading and say so in notes.' || chr(10) ||
'3. NO drive-by changes: no reformatting untouched code, no renamed variables, no comment rewrites, no import reordering beyond what the change forces. The diff a human reviews must contain ONLY the plan.' || chr(10) ||
'4. Match the file''s existing style: comment density, naming, error wording.' || chr(10) ||
'5. Output the WHOLE file — your output replaces the file byte-for-byte. Never elide with "... rest unchanged".' || chr(10) ||
'6. config_change edits are NOT yours — skip them (they are applied to agent_definitions by a human; the PR body carries them).' || chr(10) || chr(10) ||
'## The diagnosis (why this fix)' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The approved plan' || chr(10) || '{{.plan_row.body}}' || chr(10) || chr(10) ||
'## Current file bodies (your input — rewrite these)' || chr(10) || '{{.current_files.rendered}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"files": [{"path": "repo-relative/path.go", "content": "the complete new file body"}], "notes": "anything the reviewer must know"}'
          )
        ),

        'prepare', jsonb_build_object(
          'action', 'diagnose_prepare_fix_commit',
          'description', 'THE ALLOWLIST (deterministic): out-of-plan / incomplete / empty / no-op all reject. Assembles branch + commit + PR payloads.',
          'output_field', 'commit_prep',
          'next_step', 'create_branch',
          'config', jsonb_build_object(
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'plan_row.body',
            'files_field', 'implementation.result',
            'originals_field', 'current_files.files',
            'diagnosis_field', 'diagnosis_row.conclusion',
            'council_field', 'council_row.body',
            'repo_name', 'agentchassis',
            'base_branch', 'main'
          )
        ),

        'create_branch', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'fix/<short-corr> from main, via the adapter (idempotent on re-runs).',
          'output_field', 'branch_result',
          'next_step', 'commit_files',
          'config', jsonb_build_object(
            'adapter_action', 'create_branch',
            'data_literals', jsonb_build_object('repo_name', 'agentchassis', 'from_branch', 'main'),
            'data_fields', jsonb_build_object('branch', 'commit_prep.branch')
          )
        ),

        'commit_files', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'Commit the validated files to the fix branch (repo-relative — empty domain).',
          'output_field', 'commit_result',
          'next_step', 'build_gate',
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

        'build_gate', jsonb_build_object(
          'action', 'diagnose_build_gate',
          'description', 'Owner: no PRs for broken code. gofmt (changed files) + targeted go build in a golang k8s Job; passed/log is a RESULT.',
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
          'description', 'Green → open the PR. Red → NO PR; end with the build log, branch left for human inspection.',
          'config', jsonb_build_object(
            'condition', 'gate.passed == true',
            'then_step', 'create_pr',
            'else_step', 'complete_gate_failed'
          )
        ),

        'create_pr', jsonb_build_object(
          'action', 'git_adapter_request',
          'description', 'The human terminal: PR from the fix branch into main, body = the Q-H package. Created, never merged.',
          'output_field', 'pr_result',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'adapter_action', 'create_pull_request',
            'data_literals', jsonb_build_object('repo_name', 'agentchassis'),
            'data_fields', jsonb_build_object(
              'title', 'commit_prep.pr_title',
              'body', 'commit_prep.pr_body',
              'head', 'commit_prep.branch',
              'base', 'commit_prep.base_branch'
            )
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'PR opened. pr_result carries pr_url/pr_number; the whole trail is on the correlation in diagnosis_artifacts.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('pr_result', 'commit_prep', 'gate'),
            'success_message', 'fix-implementer opened a PR — human review terminal; nothing merges itself'
          )
        ),

        'complete_gate_failed', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Build gate RED: no PR. The branch and the build log are the human hand-off.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('gate', 'commit_prep'),
            'success_message', 'build gate failed — NO PR created; fix branch left for inspection, build log in gate.log'
          )
        ),

        'complete_refused', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No write: plan not approved, files unreadable, implementation failed validation, or LLM refusal.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('council_row'),
            'success_message', 'fix-implementer wrote nothing: requires a council-APPROVED plan and a valid, allowlisted implementation'
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

-- Rollback (manual): restore the pre-update snapshot from
-- agent_definitions_backup, or DELETE FROM agent_definitions WHERE
-- type='fix-implementer' AND version=1;
