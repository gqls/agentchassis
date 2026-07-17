-- 0NN_fix_proposer.sql — F1.1a + F1.1b(b) + F2.1/F2.2/F2.3 of the diagnosis→fix loop.
-- 2026-07-11 (v5: live schema hint for reviewer checks — F2.3b(a)).
-- 2026-07-10 (v4: decision router + verify step + reframe + escalation).
-- Renumber 0NN when filing. Applies to clients_db.
--
-- v7 (2026-07-17): adds a fourth council reviewer, review_reuse_agent --
-- ADVISORY (approve|object, no veto), grounded in DEV-001 (reuse-before-create
-- discipline) + FIX-036's founding incident (a reinvented trigger+triage SQL
-- pair). Slots between review_bug_historian and review_guardian. Full design:
-- docs026_concept_register/PILOT_reuse_agent_reviewer.md.
--
-- v6 (2026-07-16): adds a third council reviewer, review_bug_historian --
-- an ADVISORY seat (approve|object only, no veto option -- confirmed via
-- diagnose_council_decide_action.go that ANY reviewer's veto rejects
-- regardless of hard_veto_from, so a veto option here would make this a
-- second full gatekeeper by accident, not a historian). Seeded with a
-- curated digest of 7 independently-recurring "silent content loss during
-- rebuild/rerender" incidents (concept register docs026, STY-049 + 6
-- relations). Slots between review_editquality and review_guardian in the
-- existing sequential chain; council_decide's review_fields and repropose's
-- input_fields both extended to include it. Full design:
-- docs026_concept_register/PILOT_bug_historian_reviewer.md.
--
-- v5 (F2.3b(a)), from run 823b539f: 5 of 7 reviewer checks failed on
-- HALLUCINATED SCHEMA (p.domain, calling_context, agent_workflow_steps) —
-- reviewers write SQL blind because their prompts carry no schema. Fix: a thin
-- load_schema_hint query_database step pulls the LIVE table/column list from
-- information_schema at run time (no drift, no Go, no image dependency) and
-- both reviewer prompts get it plus the two traps that actually bit: workflow
-- steps live in agent_definitions.default_config (jsonb), not a steps table;
-- domain lives on sites (join pages.site_id = sites.id). v5 needs no image
-- beyond v4's (>= v1.0.1108).
--
-- ██ DEPLOY SEQUENCING (v4) ██ — apply this file ONLY AFTER the chassis image
-- carrying diagnose_run_checks + diagnose_escalate + the should_reframe council
-- action is live (> v1.0.1107). A v4 workflow on the old binary fails at the
-- first unknown action; the old workflow on the new binary is harmless. Order:
-- image → this file → fire. (The §1 constraint block alone is safe any time.)
--
-- Creates: (1) diagnosis_artifacts gains kind='fix_plan' (+ iteration 0 for
--              run-level artifacts) + kind='escalation' (v4);
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
--
-- v4 (F2.3), from the 2026-07-10 benchmark evidence:
--   * DECISION ROUTER — approved→complete; revise(rounds left)→run_checks→
--     repropose; rejected(first)→reframe; rejected(again)/exhausted→escalate.
--   * VERIFY STEP — reviewers attach checks:[{sql,why}] (SELECT/WITH only);
--     diagnose_run_checks answers them under the data_request containment and
--     feeds results to repropose. (Run aadd532a exhausted one verification
--     short of approval; its guardian's blockers were all "containable by
--     pre-deploy audit queries".)
--   * REFRAME — a veto means "wrong shape", so reproposing the same shape just
--     gets vetoed again (run 8c770fd5). One reframe: strictly narrower, or an
--     explicit "needs architecture review" + minimal safe interim step.
--   * ESCALATION — a first-class success terminal (kind='escalation'): the
--     human hand-off package (decision + diagnosis + final plan + reviews).

BEGIN;

-- ── 1. artifacts table: new kind + run-level iteration 0 ─────────────────────
ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_kind_check;
ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_kind_check
    CHECK (kind IN ('bundle', 'iteration_note', 'fix_plan', 'council_report', 'escalation'));
ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_iteration_check;
ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_iteration_check
    CHECK (iteration >= 0);
COMMENT ON COLUMN diagnosis_artifacts.iteration IS
    '1-based loop iteration for bundle/iteration_note; 0 = a run-level artifact (fix_plan). Derived in assemble as route.diagnose_state.iteration + 1.';

-- ── 2. the fix-proposer agent ────────────────────────────────────────────────
-- Snapshot first if a live row exists (idempotent re-apply path).
SELECT snapshot_agent('fix-proposer', 'pre-update: v7 — adding review_reuse_agent advisory council seat')
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
    'Turns a CONFIRMED diagnosis (by correlation_id) into a constrained edit plan + a council review (edit-quality + bug-historian + reuse-agent + guardian reviewers, deterministic decision). v4 router: approved completes; revise runs the reviewers'' verification checks then reproposes (cap 3 rounds); a first veto reframes once (narrower plan or explicit needs-architecture-review); rejected/exhausted escalate as a first-class hand-off artifact. Writes no code; refuses non-CONFIRMED diagnoses.',
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
            'then_step', 'load_schema_hint',
            'else_step', 'complete_refused'
          )
        ),

        -- F2.3b(a): the LIVE schema the reviewers may write checks against.
        -- Loaded once; survives in collected_data across every review round.
        'load_schema_hint', jsonb_build_object(
          'action', 'query_database',
          'description', 'Live table/column list (information_schema) so reviewer checks stop hallucinating columns.',
          'output_field', 'schema_hint',
          'next_step', 'load_last_bundle',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array(),
            'query',
              'SELECT string_agg(t.line, chr(10)) AS text FROM ( ' ||
              '  SELECT table_name || ''('' || string_agg(column_name || '' '' || data_type, '', '' ORDER BY ordinal_position) || '')'' AS line ' ||
              '  FROM information_schema.columns ' ||
              '  WHERE table_schema = ''public'' AND table_name IN ' ||
              '    (''pages'',''sites'',''site_plans'',''site_plan_pages'',''site_work_items'', ' ||
              '     ''content_components'',''page_components'',''agent_definitions'', ' ||
              '     ''diagnosis_artifacts'',''agent_error_log'') ' ||
              '  GROUP BY table_name ORDER BY table_name ' ||
              ') t'
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
          'next_step', 'review_bug_historian',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: EDIT QUALITY' || chr(10) || chr(10) ||
'You review a proposed fix plan against its diagnosis. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) does every edit CHANGE something real (audits/comments are not edits); (b) does the plan address every mechanism the diagnosis cites — quote any cited mechanism with no covering edit into missing; (c) does each edit target the causal path the diagnosis established, not an adjacent one; (d) is the plan minimal.' || chr(10) || chr(10) ||
'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: fixes a different bug, or all edits are no-ops).' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed before any revision and the results are fed back, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "editquality", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": ["cited mechanism with no covering edit"], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_bug_historian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.5 -- bug historian: does this platform have a documented history of this failure shape? Advisory only (no veto -- see design note, PILOT_bug_historian_reviewer.md).',
          'output_field', 'review_bug_historian',
          'next_step', 'review_reuse_agent',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: BUG HISTORIAN' || chr(10) || chr(10) ||
'You judge one thing: does this platform have a documented history of this exact failure shape, and does the plan account for it? You change nothing; you judge.' || chr(10) || chr(10) ||
'## Known recurring pattern: silent content loss during rebuild/rerender' || chr(10) || chr(10) ||
'This exact shape has independently recurred at least 7 times across this platform''s history, in different subsystems, built by different people, unaware of each other:' || chr(10) ||
'1. An interactive tool/game (raw HTML/JS in rendered_html) was silently deleted by a routine content rebuild -- the prose-based regression guard could not see script-heavy content was missing.' || chr(10) ||
'2. The identical DELETE+INSERT rebuild pattern destroyed a working A* pathfinding game; recurred independently on a second site months later.' || chr(10) ||
'3. The page assembler''s own visible-content filter is a SECOND, independent silent-drop path for the same class of content.' || chr(10) ||
'4. LLM-generated sections were silently rendering broken/empty; an early defence layer (schema validation, empty-section filter) was built specifically for this.' || chr(10) ||
'5. That same visible-content filter needed a second fix because it silently dropped INTENTIONALLY empty runtime-filled sections too.' || chr(10) ||
'6. A component regeneration silently renamed schema fields, breaking every dependent instance in one 16ms batch.' || chr(10) ||
'7. MOST RECENT: Go''s template engine renders any missing required field as empty with NO ERROR (missingkey=zero); this silently blanked article bodies platform-wide. Only ONE call site is guarded so far -- the root behaviour itself remains generic and unpatched, so ANY other unaudited call site has the identical exposure.' || chr(10) || chr(10) ||
'THE PATTERN: something is silently dropped, overwritten, or rendered empty -- no error, no warning, no failed work item -- because a rebuild/regeneration/render path did not check whether what it was about to discard or skip was actually required or actually present. Every instance was caught only after real content was already lost.' || chr(10) || chr(10) ||
'WHAT TO LOOK FOR in this plan: (a) does this plan touch a rebuild, rerender, regeneration, or template-render code path; (b) if so, does it introduce a NEW way for something required-but-missing to fail silently rather than loudly; (c) does the plan''s fix patch ONE call site of a shared underlying mechanism while leaving the mechanism itself generic and exploitable elsewhere; (d) is the plan fixing a SYMPTOM of this pattern without a corresponding fail-loud-not-silent guard.' || chr(10) || chr(10) ||
'Judge the plan against the pattern above: (a) does the plan''s own proposed change risk reproducing this pattern; (b) does the plan fix only a SYMPTOM of the pattern, leaving the underlying mechanism exploitable elsewhere -- if so, name specifically what else is still exposed; (c) is there a simpler, already-proven fix shape from a past occurrence that this plan should reuse instead of inventing a new one.' || chr(10) || chr(10) ||
'Verdicts: approve (no known-pattern concern), object (this plan risks or incompletely addresses the documented recurring pattern above -- say which aspect and why, in objections). You do NOT have a veto -- if you see something severe enough that this fix should not proceed at all, put it in objections at "high" severity and trust the router; if it is a true architecture-level concern, say so explicitly in notes so a human sees it either way.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "bug_historian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_reuse_agent', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.75 -- reuse agent: does this platform already have something that does this? Advisory only (no veto -- see PILOT_reuse_agent_reviewer.md).',
          'output_field', 'review_reuse_agent',
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
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: REUSE AGENT' || chr(10) || chr(10) ||
'You judge one thing: does this platform already have something that does what this plan is about to build, and did the plan check? You change nothing; you judge.' || chr(10) || chr(10) ||
'## Known discipline: reuse before create (STEP ZERO)' || chr(10) || chr(10) ||
'This platform has a standing, explicitly-named discipline: before creating any agent, action, function, or migration, search agent_definitions, the action registry, Go code, and existing workflows for an equivalent -- and never create without first demonstrating no existing coverage exists.' || chr(10) || chr(10) ||
'Documented successes of following it: reused ssh_get_status as a monitor probe rather than writing a new one; reused ListObjects for resume logic; reused datahelpers.GetIntField over a custom helper; reused snapshot_agent() instead of a new side-table migration for backups.' || chr(10) || chr(10) ||
'THE FOUNDING INCIDENT for this review seat: a prior session reinvented a trigger+triage SQL pair that already existed elsewhere in the codebase -- duplicated work a reuse check would have caught immediately.' || chr(10) || chr(10) ||
'A cautionary related case: this platform has at least one place (tool creation) where two different code paths independently solve overlapping problems (a "novel" path and a "fork" path for creating tool pages) with inconsistent side effects -- not because either was wrong alone, but because nobody unified them once both existed.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit ADD a new function, action, table, or migration; (b) if so, is there evidence in the plan''s rationale/grounded_in that an existing-coverage search was done; (c) does the diagnosis''s own evidence already name an existing mechanism this plan should extend instead of duplicating; (d) would this create a second way to do something the platform already has one way to do.' || chr(10) || chr(10) ||
'Verdicts: approve (no reuse concern, or additions are genuinely novel), object (this plan risks duplicating existing coverage -- name what already exists and where, in objections). You do NOT have a veto -- if you see something severe enough that this fix should not proceed at all, put it in objections at "high" severity and trust the router; if it is a true architecture-level concern, say so explicitly in notes so a human sees it either way.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on whether a table, column, or action name already exists, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "reuse_agent", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
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
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: PIPELINE GUARDIAN (hard-veto holder)' || chr(10) || chr(10) ||
'You protect the platform''s other pipelines from collateral damage. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) blast radius — which pipelines/workflows consume each edited file or workflow step; does the plan acknowledge them; (b) architecture-change signals — edits to shared contracts, wire formats, message shapes, exported signatures, or MANY packages at once mean this is not a constrained fix: veto and say it needs an architecture review; (c) surface ownership — workflow-JSON edits must be operation config_change and name the owning pipeline.' || chr(10) || chr(10) ||
'Verdicts: approve, object (containable concerns), veto (cross-pipeline damage or architecture change dressed as a fix). Your veto BLOCKS. If you veto, name the safest contained alternative you can see in notes — it seeds the reframe.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed before any revision and the results are fed back, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "guardian", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'council_decide', jsonb_build_object(
          'action', 'diagnose_council_decide',
          'description', 'Deterministic aggregation: veto→rejected, object→revise, else approved. Guardian holds the hard veto. Persists kind=council_report. Sets should_revise + should_reframe for the router.',
          'output_field', 'council',
          'next_step', 'check_approved',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_bug_historian.result', 'review_reuse_agent.result', 'review_guardian.result'),
            'hard_veto_from', jsonb_build_array('guardian'),
            'max_rounds', 3
          )
        ),

        -- ── F2.3 DECISION ROUTER — a chain of thin conditionals; the logic
        --    that computes the flags is deterministic Go (applyCouncilCaps).
        'check_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 1/3: approved is the ONLY path to the plain complete terminal.',
          'config', jsonb_build_object(
            'condition', 'council.decision == ''approved''',
            'then_step', 'complete',
            'else_step', 'check_rejected'
          )
        ),

        'check_rejected', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 2/3: a rejection (veto) either reframes once or escalates — never reproposes the same shape.',
          'config', jsonb_build_object(
            'condition', 'council.decision == ''rejected''',
            'then_step', 'check_reframe',
            'else_step', 'check_revise'
          )
        ),

        'check_reframe', jsonb_build_object(
          'action', 'conditional',
          'description', 'First rejection with rounds left → one reframe; a second veto (or spent cap) → escalate.',
          'config', jsonb_build_object(
            'condition', 'council.should_reframe == true',
            'then_step', 'reframe',
            'else_step', 'escalate'
          )
        ),

        'check_revise', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 3/3: revise with rounds left → answer the reviewers'' checks, then repropose; exhausted → escalate.',
          'config', jsonb_build_object(
            'condition', 'council.should_revise == true',
            'then_step', 'run_checks',
            'else_step', 'escalate'
          )
        ),

        -- ── F2.3 VERIFY STEP — answer the reviewers' read-only checks before
        --    the next repropose, so a fact-shaped objection is settled with
        --    evidence instead of another blind revision (run aadd532a).
        'run_checks', jsonb_build_object(
          'action', 'diagnose_run_checks',
          'description', 'Run the reviewers'' checks ([{sql,why}], SELECT/WITH only) under the data_request containment; results feed repropose.',
          'output_field', 'check_results',
          'next_step', 'repropose',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'check_fields', jsonb_build_array('review_editquality.result.checks', 'review_guardian.result.checks')
          )
        ),

        'repropose', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Revise the plan to address the council objections, then re-run persist + review + decide (the loop).',
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
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'review_editquality', 'review_bug_historian', 'review_reuse_agent', 'review_guardian', 'check_results'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — REVISE the constrained fix plan' || chr(10) || chr(10) ||
'A council reviewed your previous plan and asked for revision. Produce a NEW full plan (same JSON schema) that addresses every objection and covers every mechanism listed missing. You still write no code — you name edits.' || chr(10) || chr(10) ||
'The SAME hard rules apply: platform not site data; minimal; grounded; no new deps/DDL; workflow-JSON edits are config_change and name the owning pipeline; cover EVERY cited mechanism; every edit CHANGES something (no audits, no comment-only, no "no change required").' || chr(10) || chr(10) ||
'## The confirmed diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## Your previous plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Edit-quality reviewer said' || chr(10) || '{{.review_editquality.result}}' || chr(10) || chr(10) ||
'## Bug-historian reviewer said (advisory, no veto)' || chr(10) || '{{.review_bug_historian.result}}' || chr(10) || chr(10) ||
'## Reuse-agent reviewer said (advisory, no veto)' || chr(10) || '{{.review_reuse_agent.result}}' || chr(10) || chr(10) ||
'## Guardian reviewer said (holds a hard veto)' || chr(10) || '{{.review_guardian.result}}' || chr(10) || chr(10) ||
'## Verification results (the reviewers'' own read-only queries, now answered)' || chr(10) || '{{.check_results.results_text}}' || chr(10) || chr(10) ||
'Use these results to SETTLE any objection that hinged on an unverified fact — cite them in grounded_in. If a result contradicts an edit, change or drop the edit; do not argue with the data.' || chr(10) || chr(10) ||
'## Output — ONLY the plan JSON (summary, edits[], grounded_in[], risks). Address the objections; do not merely restate the old plan.'
          )
        ),

        -- ── F2.3 REFRAME — a veto means the SHAPE is wrong (run 8c770fd5:
        --    architecture change dressed as a point fix). One chance to change
        --    shape; a second veto escalates.
        'reframe', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'After a council VETO: produce a strictly narrower remediation, or an explicit needs-architecture-review declaration plus the minimal safe interim step. One attempt; then escalate.',
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
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'review_editquality', 'review_guardian'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — REFRAME after a council VETO' || chr(10) || chr(10) ||
'The council REJECTED your plan outright — the guardian judged it an architecture-level change dressed as a contained fix. Do NOT resubmit the same shape: it will be vetoed again. Produce ONE of:' || chr(10) || chr(10) ||
'(a) a STRICTLY NARROWER remediation the guardian could accept — prefer the reviewer''s own recommended alternative if its review names one; smallest possible blast radius; a site-scoped interim step is acceptable HERE (an exception to the platform rule) PROVIDED risks names the deferred structural fix explicitly ("interim only; the platform fix — <name it> — needs an architecture review"); or' || chr(10) || chr(10) ||
'(b) if no contained remediation exists at all, a plan whose only edits are the minimal safe preparatory step, with risks stating plainly which decision the architecture review must take.' || chr(10) || chr(10) ||
'Same JSON schema and remaining hard rules: grounded; no new deps/DDL; workflow-JSON edits are config_change naming the owning pipeline; every edit CHANGES something.' || chr(10) || chr(10) ||
'## The confirmed diagnosis' || chr(10) || '{{.diagnosis_row.conclusion}}' || chr(10) || chr(10) ||
'## Your VETOED plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Edit-quality review' || chr(10) || '{{.review_editquality.result}}' || chr(10) || chr(10) ||
'## Guardian review (the veto — read its notes for the recommended alternative)' || chr(10) || '{{.review_guardian.result}}' || chr(10) || chr(10) ||
'## Output — ONLY the plan JSON (summary, edits[], grounded_in[], risks).'
          )
        ),

        -- ── F2.3 ESCALATION — a first-class success terminal: the hand-off
        --    package a human needs to decide (kind=escalation).
        'escalate', jsonb_build_object(
          'action', 'diagnose_escalate',
          'description', 'Persist the human hand-off package: decision + diagnosis + final plan + all three reviews (their notes carry the recommended alternative / unrun checklist).',
          'output_field', 'escalation',
          'next_step', 'complete_escalated',
          'config', jsonb_build_object(
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_bug_historian.result', 'review_reuse_agent.result', 'review_guardian.result')
          )
        ),

        'complete_escalated', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Escalated: no approvable constrained plan. The hand-off package is in diagnosis_artifacts kind=escalation.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('escalation', 'council', 'plan_persisted'),
            'success_message', 'fix-proposer escalated: no approvable constrained plan; fetch the hand-off package from diagnosis_artifacts kind=escalation by correlation_id'
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'APPROVED (v4: the only route here). Plan + council report persisted; fetch both from diagnosis_artifacts by correlation_id. F1.1b(c) takes an approved plan to a branch + PR.',
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
--   v4 → v3: restore the pre-update snapshot from agent_definitions_backup
--   (snapshot_agent wrote it on apply); or DELETE the row and re-apply the v3
--   file from git history.
--   DELETE FROM agent_definitions WHERE type='fix-proposer' AND version=1;
--   ALTER TABLE diagnosis_artifacts DROP CONSTRAINT diagnosis_artifacts_kind_check;
--   ALTER TABLE diagnosis_artifacts ADD CONSTRAINT diagnosis_artifacts_kind_check
--       CHECK (kind IN ('bundle','iteration_note'));
--   (leave iteration_check at >= 0; 0-rows only exist if a fix_plan was written;
--    drop 'escalation' from the kind list only after deleting any escalation rows)
