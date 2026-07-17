-- 0NN_council_gate.sql — COUNCIL GATE v1 (advisory): the F2 council as a service.
-- 2026-07-17, "council gate" thread. Renumber 0NN when filing. Applies to clients_db.
--
-- ██ DO NOT APPLY YET ██ — owner decision 2026-07-17: the gate launches only
-- once more concept-register stage-3 seats are live (roster ruling; the other
-- three rulings: scope = platform/ + internal/ + pkg/; advisory mode first;
-- credits per submission = per task/commit). This file is complete and
-- apply-ready; a seed is LIVE config the moment it is applied, so applying
-- waits for the owner's named go, same as the bug-historian seat did.
--
-- WHAT THIS IS. Design: DESIGN_feature_builder_and_council_gate.md §2. The
-- fix-proposer's council already judges a fix_plan artifact by correlation_id
-- and does not care who authored it. This seed opens that council to ANY
-- author: a working session submits its change as a fix_plan-shaped artifact
-- (diff sketches + rationale, via 097_TRIGGER_council_review_v1.sh) and the
-- SAME three reviewers + deterministic decision judge it. Advisory: the
-- verdict is recorded (council_report artifact + doc_notes entry), never
-- enforced — enforcement (PR-mode) is build-order step 4, owner-gated.
--
-- NO NEW GO. Every action here is registered and live (>= v1.0.1127):
--   query_database, diagnose_persist_fix_plan (plan_field can read
--   input_data.plan — verified against DiagnosePersistFixPlanAction, which
--   resolves plan_field over CollectedData where input_data lives),
--   execute_llm_prompt, diagnose_council_decide (round counting is
--   orchestration-scoped since v1.0.1108+, so resubmissions on the same
--   correlation do not inherit stale rounds), diagnose_run_checks,
--   append_doc_note, conditional, complete_workflow.
--
-- LOCKSTEP WARNING. The SEVEN reviewer steps + relevance filter mirror the
-- live fix-proposer AS READ FROM THE DATABASE at apply time (2026-07-17
-- evening: v6 base + reuse-agent + guidelines + check_fields fix +
-- tooling-provenance + select_panel/gates + adoption-guardian) with ONE
-- change each: the "## The diagnosis" context section becomes the author's
-- stated rationale ({{.input_data.rationale}}), and select_panel's
-- extra_text_fields scans the rationale instead of a diagnosis conclusion.
-- The roster is growing FREQUENTLY (owner's warning) — before ANY re-apply,
-- diff this file's reviewer steps against the live fix-proposer row first,
-- and mirror any new seat/gate/footprint here; a seat or filter change in
-- fix-proposer MUST be mirrored here in the same migration, or the gate's
-- council silently lags the fix loop's (dedup-index/Go-list class of drift).
-- DELIBERATE divergence, not drift: fix-proposer's code_lookup step is NOT
-- mirrored — it answers reviewers' code-shaped questions for the blind
-- REPROPOSER, which the gate lacks (its authors are code-capable sessions
-- who read the objections themselves; same reason repropose/reframe are
-- absent). Requires image >= v1.0.1133 (select_review_panel +
-- council_decide abstention — pod-verified); image first, then this seed.
--
-- DIFFERENCES from the fix-proposer workflow, all deliberate:
--   * No CONFIRMED-diagnosis gate: intake is a submission, not a diagnosis.
--   * No repropose/reframe loop: the AUTHOR revises, not the council — a
--     revise verdict terminates with the objections + the reviewers' checks
--     answered (run_checks still runs, so fact-shaped objections come back
--     settled with evidence). Resubmit on the SAME correlation so the
--     artifact trail accumulates rounds in one place.
--   * run_checks includes ALL SEVEN reviewers' checks. (A v6-inherited
--     omission — only editquality + guardian — was flagged by this thread and
--     fixed the same day by the concept-register thread's v9, applied live.)
--   * Terminal always writes a doc_notes verdict entry (deterministic compose
--     by SQL — an awareness surface that could hallucinate would defeat
--     itself, same rule as the digest). Categories: council-gate + verdict.
--   * max_plan_bytes 65536 (double the fix-proposer's): submissions carry
--     real diff sketches. max_edits stays 8 — a wider change than 8 files is
--     architecture-shaped and should meet a human, not this gate.
--
-- APPROVED means: commit with the trailer line
--     Council-Reviewed: <fix_correlation_id>
-- which is what makes 098_REPORT_unreviewed_commits_v1.sh's join exact
-- (commit ↔ verdict by correlation, not by heuristic file overlap).

BEGIN;

-- Snapshot first if a live row exists (idempotent re-apply path).
SELECT snapshot_agent('council-gate', 'pre-update: council gate v1 (advisory)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='council-gate' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'council-gate',
    'Council Gate (advisory review service)',
    'The F2 council as a service: judges a SUBMITTED change (fix_plan-shaped artifact: diff sketches + rationale, any author, by correlation_id) through the same 7-reviewer council as the live fix-proposer: edit-quality + guardian (hard-veto) always-on, bug-historian / reuse-agent / guidelines / tooling-provenance / adoption-guardian advisory and relevance-gated by the select_review_panel filter. No repropose loop — the author revises and resubmits on the same correlation. Verdict = council_report artifact + doc_notes entry (categories council-gate+verdict). Advisory: records approval, never enforces it. Scope ruling 2026-07-17: platform/, internal/, pkg/.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "review"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_schema_hint',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 600,
      'steps', jsonb_build_object(

        -- F2.3b(a), verbatim from v6: the LIVE schema reviewers may write
        -- checks against. Loaded once; survives in collected_data.
        'load_schema_hint', jsonb_build_object(
          'action', 'query_database',
          'description', 'Live table/column list (information_schema) so reviewer checks stop hallucinating columns.',
          'output_field', 'schema_hint',
          'next_step', 'persist_submission',
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

        -- The submission wrapper's landing point: the plan arrives ready-made
        -- in input_data.plan (authored by the submitting session, validated by
        -- the trigger script) and is persisted through the SAME structural
        -- validation as a proposer plan — no-op edits, traversal paths, and
        -- empty rationales fail here, before any credits are spent on review.
        'persist_submission', jsonb_build_object(
          'action', 'diagnose_persist_fix_plan',
          'description', 'Validate the submitted plan structurally and persist it (kind=fix_plan) on the submission correlation. Failure = complete_invalid, no review spend.',
          'output_field', 'plan_persisted',
          'next_step', 'select_panel',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'plan_field', 'input_data.plan',
            'max_plan_bytes', 65536
          )
        ),

        -- v11-mirrored relevance filter: deterministic seat selection from the
        -- submission's edited paths + rationale text. Footprints identical to
        -- the live fix-proposer's; only the extra text field differs (a
        -- submission has a rationale where a fix has a diagnosis).
        'select_panel', jsonb_build_object(
          'action', 'select_review_panel',
          'description', 'Stage-3 relevance filter: deterministic (no-LLM) step that decides which OPTIONAL reviewer seats are relevant to THIS submission, from the edited file paths + rationale text vs a config footprint map. Emits panel.run_<seat> booleans the gates below check. Skipped seats abstain; council_decide tolerates absent fields.',
          'output_field', 'panel',
          'next_step', 'review_editquality',
          'config', jsonb_build_object(
            'plan_field', 'plan_persisted',
            'extra_text_fields', jsonb_build_array('input_data.rationale'),
            'footprints', jsonb_build_object(
              'bug_historian', jsonb_build_array('rerender','render','save_page_sections','sectionhasvisiblecontent','call_agent.go','missingkey','page_components','content_components'),
              'reuse_agent', jsonb_build_array('_action.go','.sql','migration','create table','new '),
              'guidelines', jsonb_build_array('input_contract','output_contract','idx_swi_dedup','site_work_items','agent_definitions','input_schema','save_page_sections'),
              'tooling_provenance', jsonb_build_array('contextkit','cmd/bundle','bundle','doc_plans','doc_notes','resolve_action','registry.go','docubundle','travelling','dedup','thin_versions'),
              'adoption', jsonb_build_array('apply_adoption_plan','site-adoption','adoption','needs_domain_research','domain-research-classifier','site_archetype','design_intent','design_reference','content_direction')
            )
          )
        ),

        'gate_bug_historian', jsonb_build_object(
          'action', 'conditional',
          'description', 'Relevance gate: run the bug-historian only if the submission touches a rebuild/rerender/render path; else skip (it abstains).',
          'config', jsonb_build_object(
            'condition', 'panel.run_bug_historian == true',
            'then_step', 'review_bug_historian',
            'else_step', 'gate_reuse_agent'
          )
        ),

        'gate_reuse_agent', jsonb_build_object(
          'action', 'conditional',
          'description', 'Relevance gate: run the reuse agent only if the submission adds new code/SQL; else skip.',
          'config', jsonb_build_object(
            'condition', 'panel.run_reuse_agent == true',
            'then_step', 'review_reuse_agent',
            'else_step', 'gate_guidelines'
          )
        ),

        'gate_guidelines', jsonb_build_object(
          'action', 'conditional',
          'description', 'Relevance gate: run the guidelines agent only if the submission touches contracts/work-items/agent defs/schemas; else skip.',
          'config', jsonb_build_object(
            'condition', 'panel.run_guidelines == true',
            'then_step', 'review_guidelines',
            'else_step', 'gate_tooling_provenance'
          )
        ),

        'gate_tooling_provenance', jsonb_build_object(
          'action', 'conditional',
          'description', 'Relevance gate: run the tooling & provenance reviewer only if the submission touches investigation/doc tooling; else skip.',
          'config', jsonb_build_object(
            'condition', 'panel.run_tooling_provenance == true',
            'then_step', 'review_tooling_provenance',
            'else_step', 'gate_adoption'
          )
        ),

        'gate_adoption', jsonb_build_object(
          'action', 'conditional',
          'description', 'Relevance gate: run the adoption-pipeline guardian only if the submission touches the adoption pipeline; else skip.',
          'config', jsonb_build_object(
            'condition', 'panel.run_adoption == true',
            'then_step', 'review_adoption_guardian',
            'else_step', 'review_guardian'
          )
        ),

        'review_editquality', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1 — edit quality (gate mode): real changes, rationale↔edit coverage both ways, right causal target, minimality.',
          'output_field', 'review_editquality',
          'next_step', 'gate_bug_historian',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: EDIT QUALITY' || chr(10) || chr(10) ||
'You review a SUBMITTED change plan against its author''s stated rationale. The author is a working session submitting its own change for advisory review before committing — there is no diagnosis behind it; the rationale is the claim you judge against. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) does every edit CHANGE something real (audits/comments are not edits); (b) coverage BOTH ways — quote any claim in the rationale with no covering edit into missing, and put any edit the rationale does not explain into objections; (c) does each edit target the problem the rationale states, not an adjacent one; (d) is the change minimal for what the rationale claims.' || chr(10) || chr(10) ||
'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: the edits do not implement the rationale, or all edits are no-ops).' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed and the results are recorded with the verdict, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The submitted plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "editquality", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": ["rationale claim with no covering edit"], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_bug_historian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.5 — bug historian: does this platform have a documented history of this failure shape? Advisory only (no veto — see PILOT_bug_historian_reviewer.md).',
          'output_field', 'review_bug_historian',
          'next_step', 'gate_reuse_agent',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: BUG HISTORIAN' || chr(10) || chr(10) ||
'You judge one thing: does this platform have a documented history of this exact failure shape, and does the submitted change account for it? You change nothing; you judge.' || chr(10) || chr(10) ||
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
'WHAT TO LOOK FOR in this submission: (a) does it touch a rebuild, rerender, regeneration, or template-render code path; (b) if so, does it introduce a NEW way for something required-but-missing to fail silently rather than loudly; (c) does it patch ONE call site of a shared underlying mechanism while leaving the mechanism itself generic and exploitable elsewhere; (d) is it fixing a SYMPTOM of this pattern without a corresponding fail-loud-not-silent guard.' || chr(10) || chr(10) ||
'Judge the submission against the pattern above: (a) does the proposed change risk reproducing this pattern; (b) does it fix only a SYMPTOM of the pattern, leaving the underlying mechanism exploitable elsewhere -- if so, name specifically what else is still exposed; (c) is there a simpler, already-proven fix shape from a past occurrence that this change should reuse instead of inventing a new one.' || chr(10) || chr(10) ||
'Verdicts: approve (no known-pattern concern), object (this change risks or incompletely addresses the documented recurring pattern above -- say which aspect and why, in objections). You do NOT have a veto -- if you see something severe enough that this change should not proceed at all, put it in objections at "high" severity and trust the router; if it is a true architecture-level concern, say so explicitly in notes so a human sees it either way.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The submitted plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "bug_historian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_reuse_agent', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.75 — reuse agent: does this platform already have something that does this? Advisory only (no veto — see PILOT_reuse_agent_reviewer.md).',
          'output_field', 'review_reuse_agent',
          'next_step', 'gate_guidelines',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: REUSE AGENT' || chr(10) || chr(10) ||
'You judge one thing: does this platform already have something that does what this plan is about to build, and did the plan check? You change nothing; you judge.' || chr(10) || chr(10) ||
'## Known discipline: reuse before create (STEP ZERO)' || chr(10) || chr(10) ||
'This platform has a standing, explicitly-named discipline: before creating any agent, action, function, or migration, search agent_definitions, the action registry, Go code, and existing workflows for an equivalent -- and never create without first demonstrating no existing coverage exists.' || chr(10) || chr(10) ||
'Documented successes of following it: reused ssh_get_status as a monitor probe rather than writing a new one; reused ListObjects for resume logic; reused datahelpers.GetIntField over a custom helper; reused snapshot_agent() instead of a new side-table migration for backups.' || chr(10) || chr(10) ||
'THE FOUNDING INCIDENT for this review seat: a prior session reinvented a trigger+triage SQL pair that already existed elsewhere in the codebase -- duplicated work a reuse check would have caught immediately.' || chr(10) || chr(10) ||
'A cautionary related case: this platform has at least one place (tool creation) where two different code paths independently solve overlapping problems (a "novel" path and a "fork" path for creating tool pages) with inconsistent side effects -- not because either was wrong alone, but because nobody unified them once both existed.' || chr(10) || chr(10) ||
'Judge the plan: (a) does any edit ADD a new function, action, table, or migration; (b) if so, is there evidence in the plan''s rationale/grounded_in that an existing-coverage search was done; (c) does the submission''s own rationale/grounded_in already name an existing mechanism this plan should extend instead of duplicating; (d) would this create a second way to do something the platform already has one way to do.' || chr(10) || chr(10) ||
'Verdicts: approve (no reuse concern, or additions are genuinely novel), object (this plan risks duplicating existing coverage -- name what already exists and where, in objections). You do NOT have a veto -- if you see something severe enough that this change should not proceed at all, put it in objections at "high" severity and trust the router; if it is a true architecture-level concern, say so explicitly in notes so a human sees it either way.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on whether a table, column, or action name already exists, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "reuse_agent", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_guidelines', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.9 — guidelines agent: does this follow the platform''s documented conventions/contracts, and if not, is the PLAN wrong or the RULE? Advisory only (no veto; a guideline-gap is a note not an objection — see PILOT_guidelines_agent_reviewer.md).',
          'output_field', 'review_guidelines',
          'next_step', 'gate_tooling_provenance',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: GUIDELINES AGENT' || chr(10) || chr(10) ||
'You judge two things about this submitted plan: (1) does it FOLLOW the platform''s documented conventions and contracts; (2) where it appears not to, is that because the PLAN is wrong, or because the RULE is? You change nothing; you judge.' || chr(10) || chr(10) ||
'## The platform''s load-bearing rules (the ones people keep relearning)' || chr(10) ||
'- WRAPPER-ORCHESTRATOR: anything doing substantive work (LLM calls, crawls, heavy DB, minutes of runtime) must run in a spawned pod via a parent (processing_mode:"orchestrator" + spawn_agent), never inline on a shared chassis slot; file writes from non-spawned actions die with a random pod.' || chr(10) ||
'- WORK-ITEM DEDUP: site_work_items dedup is idx_swi_dedup UNIQUE(site_id, item_key) over NON-TERMINAL statuses; the terminal-status set is a contract (drift between it and the Go list breaks every keyed insert fleet-wide); use DELETE+INSERT, not ON CONFLICT.' || chr(10) ||
'- TRUTHFUL PROVENANCE: hand-made work items copy the real owning path''s metadata, deviate only truthfully (source=''manual'', real created_by), and take URLs from pages.url -- never invent a path.' || chr(10) ||
'- DECLARED CONTRACTS: any input a workflow reads must be declared in the agent''s input_contract; a call site''s input_mapping must satisfy the callee''s contract.' || chr(10) ||
'- SCHEMA-SOURCE TIERS: a component field with required:true must set on_missing deliberately -- leaving it skip_field/empty hits the switch default and silently defers the whole section.' || chr(10) || chr(10) ||
'## The meta-rule for THIS seat (important)' || chr(10) ||
'A GUIDELINE-GAP is not a violation. If the submission''s evidence shows the change is correct but exposes a documented rule that is itself wrong or stale (this happens -- a runbook rule about max_tokens placement was recently found to be backwards), say so in notes as a recommended side-task / guideline amendment, and APPROVE. Do NOT object: forcing a correct change to revise because the underlying rule is bad is the wrong move. Object ONLY when the PLAN breaks a rule that is right.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (does a contract column exist, does an agent declare an input), put it in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "guidelines", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the specific rule violated", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "any guideline-gap goes HERE (approve + note), not in objections"}'
          )
        ),

        'review_tooling_provenance', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 1.95 — tooling & provenance: does this change use the platform''s own investigation (cmd/bundle/contextkit) and documentation (doc_plans/doc_notes travelling docs) machinery, or reinvent/work around it? Advisory only (no veto — see PILOT_tooling_provenance_reviewer.md).',
          'output_field', 'review_tooling_provenance',
          'next_step', 'gate_adoption',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: TOOLING & PROVENANCE' || chr(10) || chr(10) ||
'You judge one thing: does this change use the platform''s own investigation and documentation machinery, or reinvent / work around it? You change nothing; you judge.' || chr(10) || chr(10) ||
'## The platform''s own machinery (use it; don''t reinvent it)' || chr(10) ||
'- INVESTIGATION: cmd/bundle / contextkit is the platform''s tool for reading a change''s real supported method before touching it -- the standing rule is "that''s a code question -> a bundle." Actions resolve from the REGISTRY (key -> Handler symbol -> function), NEVER by filename convention (execute_llm_prompt lives in ai_actions.go; some actions lack the _action suffix) -- a change that assumes a file name from an action key is repeating a documented recurring mistake.' || chr(10) ||
'- TRAVELLING DOCS: every tool/pipeline/agent carries a living PLAN + NOTES in Postgres (doc_plans / doc_notes, keyed by subject_type + subject_key). Changes are supposed to load the subject''s prior decisions before changing it and leave a NOTES entry -- so the next change builds on this one instead of re-deriving lost context. (The fix loop itself uses exactly this pattern via diagnosis_artifacts / doc_notes.)' || chr(10) ||
'- DOC HYGIENE tooling (dedup / thin_versions / archiving) already exists.' || chr(10) || chr(10) ||
'Judge the plan: (a) does it add new ad-hoc context-gathering / bundling / source-parsing code where cmd/bundle/contextkit or an existing action already does it; (b) does it touch a tool/pipeline that has a travelling PLAN/NOTES without accounting for it; (c) does any edit resolve an action/handler by filename convention rather than the registry; (d) does it reinvent existing doc/context tooling. If none apply (most changes touch no tooling at all), approve.' || chr(10) || chr(10) ||
'Verdicts: approve (uses the platform''s machinery, or touches none of it), object (reinvents or works around existing tooling / ignores a subject''s travelling docs -- name the specific existing mechanism it should use). You do NOT have a veto -- put a severe concern in objections at "high" severity and trust the router; note a true architecture-level concern explicitly.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on whether a doc_plans/doc_notes row or a registry entry exists, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "tooling_provenance", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the existing tooling ignored/reinvented", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_adoption_guardian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer (gated, candidate #3) — adoption-pipeline guardian: does the change respect the adoption pipeline''s write-then-relay contract (ADO-006)? Advisory only (no veto). Runs only when select_panel matches the adoption footprint.',
          'output_field', 'review_adoption_guardian',
          'next_step', 'review_guardian',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
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
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "adoption_guardian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "names the adoption contract broken", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_guardian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 2 — pipeline guardian: surface ownership, blast radius, architecture-change signals. HARD VETO holder.',
          'output_field', 'review_guardian',
          'next_step', 'council_decide',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('input_data', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: PIPELINE GUARDIAN (hard-veto holder)' || chr(10) || chr(10) ||
'You protect the platform''s other pipelines from collateral damage. You review a SUBMITTED change plan (any author, advisory gate). You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) blast radius — which pipelines/workflows consume each edited file or workflow step; does the submission acknowledge them; (b) architecture-change signals — edits to shared contracts, wire formats, message shapes, exported signatures, or MANY packages at once mean this is not a constrained change: veto and say it needs an architecture review; (c) surface ownership — workflow-JSON edits must be operation config_change and name the owning pipeline.' || chr(10) || chr(10) ||
'Verdicts: approve, object (containable concerns), veto (cross-pipeline damage or architecture change dressed as a contained fix). Your veto BLOCKS. If you veto, name the safest contained alternative you can see in notes — the author reads it before resubmitting.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed and the results are recorded with the verdict, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The author''s stated rationale' || chr(10) || '{{.input_data.rationale}}' || chr(10) || chr(10) ||
'## The submitted plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "guardian", "verdict": "approve|object|veto", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'council_decide', jsonb_build_object(
          'action', 'diagnose_council_decide',
          'description', 'Deterministic aggregation: veto→rejected, object→revise, else approved. Guardian holds the hard veto. Persists kind=council_report on the submission correlation.',
          'output_field', 'council',
          'next_step', 'route_checks',
          'config', jsonb_build_object(
            'error_step', 'complete_invalid',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_bug_historian.result', 'review_reuse_agent.result', 'review_guidelines.result', 'review_tooling_provenance.result', 'review_guardian.result', 'review_adoption_guardian.result'),
            'hard_veto_from', jsonb_build_array('guardian'),
            'max_rounds', 3
          )
        ),

        -- GATE ROUTER — unlike the fix-proposer there is no repropose/reframe:
        -- the author revises. A revise verdict still runs the reviewers' checks
        -- first, so fact-shaped objections come back SETTLED with evidence
        -- instead of as homework (the F2.3 lesson, run aadd532a).
        'route_checks', jsonb_build_object(
          'action', 'conditional',
          'description', 'Revise verdict → answer the reviewers'' checks before surfacing it; anything else → surface the verdict directly.',
          'config', jsonb_build_object(
            'condition', 'council.should_revise == true',
            'then_step', 'run_checks',
            'else_step', 'compose_verdict'
          )
        ),

        'run_checks', jsonb_build_object(
          'action', 'diagnose_run_checks',
          'description', 'Run ALL three reviewers'' checks ([{sql,why}], SELECT/WITH only) under the data_request containment; results ride the verdict note back to the author.',
          'output_field', 'check_results',
          'next_step', 'compose_verdict_checked',
          'config', jsonb_build_object(
            'error_step', 'compose_verdict',
            'check_fields', jsonb_build_array('review_editquality.result.checks', 'review_bug_historian.result.checks', 'review_reuse_agent.result.checks', 'review_guidelines.result.checks', 'review_tooling_provenance.result.checks', 'review_guardian.result.checks', 'review_adoption_guardian.result.checks')
          )
        ),

        -- Deterministic verdict note (no LLM — an awareness surface that could
        -- hallucinate what the council decided would defeat itself).
        'compose_verdict', jsonb_build_object(
          'action', 'query_database',
          'description', 'Compose the verdict doc_note body by SQL (approved/rejected/exhausted path — no check results).',
          'output_field', 'gate_note',
          'next_step', 'append_verdict',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id', 'council.decision', 'council.decided_by', 'council.round', 'plan_persisted.summary'),
            'query',
              'SELECT ''COUNCIL GATE — '' || upper(COALESCE($2, ''unknown'')) || '' — '' || COALESCE($3, '''') || '' (round '' || COALESCE($4::text, ''?'') || '')'' || chr(10) || ' ||
              '       ''submission correlation: '' || $1 || chr(10) || ' ||
              '       ''plan summary: '' || COALESCE($5, ''(none)'') || chr(10) || ' ||
              '       ''full report: diagnosis_artifacts kind=council_report, correlation '' || $1 || '', latest row'' || ' ||
              '       CASE WHEN $2 = ''approved'' THEN chr(10) || ''author: commit with the trailer line  Council-Reviewed: '' || $1 ELSE '''' END AS body'
          )
        ),

        'compose_verdict_checked', jsonb_build_object(
          'action', 'query_database',
          'description', 'Compose the verdict doc_note body by SQL (revise path — includes the answered checks).',
          'output_field', 'gate_note',
          'next_step', 'append_verdict',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.fix_correlation_id', 'council.decision', 'council.decided_by', 'council.round', 'plan_persisted.summary', 'check_results.results_text'),
            'query',
              'SELECT ''COUNCIL GATE — '' || upper(COALESCE($2, ''unknown'')) || '' — '' || COALESCE($3, '''') || '' (round '' || COALESCE($4::text, ''?'') || '')'' || chr(10) || ' ||
              '       ''submission correlation: '' || $1 || chr(10) || ' ||
              '       ''plan summary: '' || COALESCE($5, ''(none)'') || chr(10) || ' ||
              '       ''reviewers'''' checks, answered:'' || chr(10) || COALESCE($6, ''(no checks were requested)'') || chr(10) || ' ||
              '       ''full report: diagnosis_artifacts kind=council_report, correlation '' || $1 || '', latest row'' || chr(10) || ' ||
              '       ''author: revise and resubmit on the SAME correlation so the trail accumulates'' AS body'
          )
        ),

        'append_verdict', jsonb_build_object(
          'action', 'append_doc_note',
          'description', 'Persist the verdict to doc_notes (pipeline/diagnose, categories council-gate+verdict) — the awareness channel. A note failure must not eat the verdict: council_report is already persisted, so route on to the terminal.',
          'output_field', 'verdict_note',
          'next_step', 'route_approved',
          'config', jsonb_build_object(
            'error_step', 'route_approved',
            'subject_type', 'pipeline',
            'subject_key', 'diagnose',
            'note_body_field', 'gate_note.body',
            'note_categories', jsonb_build_array('council-gate', 'verdict'),
            'note_source', 'council-gate',
            'created_by', 'council-gate'
          )
        ),

        'route_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Terminal router 1/2: approved gets its own terminal (the trailer instruction).',
          'config', jsonb_build_object(
            'condition', 'council.decision == ''approved''',
            'then_step', 'complete_approved',
            'else_step', 'route_rejected'
          )
        ),

        'route_rejected', jsonb_build_object(
          'action', 'conditional',
          'description', 'Terminal router 2/2: rejected (veto) vs revise/exhausted.',
          'config', jsonb_build_object(
            'condition', 'council.decision == ''rejected''',
            'then_step', 'complete_rejected',
            'else_step', 'complete_revise'
          )
        ),

        'complete_approved', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'APPROVED. Advisory: the author may commit, with the Council-Reviewed trailer so the 098 report can join commit↔verdict exactly.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('plan_persisted', 'council'),
            'success_message', 'council gate APPROVED — commit with trailer "Council-Reviewed: <fix_correlation_id>"; full report in diagnosis_artifacts kind=council_report'
          )
        ),

        'complete_revise', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'REVISE. The objections + answered checks are in the council_report and the doc_notes verdict entry; the author revises and resubmits on the same correlation.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('council', 'check_results'),
            'success_message', 'council gate says REVISE — read the objections and answered checks (council_report / doc_notes categories council-gate), revise, resubmit on the SAME correlation'
          )
        ),

        'complete_rejected', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'REJECTED (a veto). Advisory mode cannot block a commit — but the verdict is on the record, and the guardian''s notes name the safest contained alternative.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('council'),
            'success_message', 'council gate REJECTED this submission — do not ship as-is; the guardian''s notes name the safest contained alternative (council_report, latest row)'
          )
        ),

        'complete_invalid', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No verdict: the submission failed structural validation, or a reviewer/decide step errored. Nothing was decided; fix the submission and re-trigger.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('plan_persisted'),
            'success_message', 'council gate made NO verdict: submission failed validation (or a review step errored) — fix the submission JSON and re-trigger'
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

-- Post-apply verification (run these before announcing the gate live):
--   SELECT type, version, is_active, jsonb_object_keys(default_config->'workflow'->'steps')
--   FROM agent_definitions WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false;
--   -- expect 27 steps; review_fields must list all seven reviewers:
--   SELECT default_config->'workflow'->'steps'->'council_decide'->'config'->'review_fields'
--   FROM agent_definitions WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false;
--
-- Rollback: restore the pre-update snapshot from agent_definitions_backup
-- (snapshot_agent wrote it on apply), or
--   DELETE FROM agent_definitions WHERE type='council-gate' AND version=1;
