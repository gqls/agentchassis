-- 0NN_feature_designer.sql — feature builder delta 1: the feature-designer agent (v1 DRAFT).
-- 2026-07-17, "fixloop feature builder" thread. Renumber 0NN when filing. Applies to clients_db.
--
-- ██ STATUS: DRAFT, NOT APPLIED ██ — this file ships in the repo per the seed
-- discipline this very feature encodes (seeds are PR files, never executed by
-- the loop). The owner applies it AFTER the image carrying staged-v1 validation
-- in diagnose_persist_fix_plan (commit 4b3d50f4c or later) is live. A staged
-- plan hitting the OLD binary fails validation as an unknown shape.
--
-- WHAT THIS AGENT IS. The feature builder's DESIGN step (design §1,
-- DESIGN_feature_builder_and_council_gate.md; schema + owner decisions D1-D6 in
-- SCHEMA_staged_plan_v1.md): turns an owner-APPROVED capability spec into a
-- STAGED PLAN (plan_format staged-v1) and walks it through the SAME 3-seat
-- council + deterministic router the fix-proposer proved. Reuses the proposer's
-- machinery wholesale: diagnose_persist_fix_plan (staged path), the reviewers'
-- prompt seams, diagnose_council_decide, diagnose_run_checks,
-- diagnose_escalate. It writes NO code — the stage-loop implementer (delta 2)
-- turns an approved staged plan into a branch + PR.
--
-- INTAKE + THE GATE. Input: {"fix_correlation_id": "<fresh uuid for this run>",
-- "work_item_id": "<site_work_items.id>"}. The mirror of the proposer's
-- CONFIRMED gate is SPEC APPROVAL: the workflow refuses any work item that is
-- not item_type='capability_gap' carrying BOTH spec keys:
--   * owner_approval  — written by the owner, e.g.
--       UPDATE site_work_items SET spec = spec ||
--         '{"owner_approval": {"approved_by": "<name>", "date": "YYYY-MM-DD"}}'
--       WHERE id = '<work_item_id>';
--   * code_pointers   — repo paths (with a line of why) the design may build
--       on. The designer runs IN-CHASSIS with no repo read token (that token
--       is implementer-pod-only, owner ruling 2026-07-13), so it CANNOT
--       explore code: a spec without pointers would make it guess file paths.
--       It is instructed to use ONLY pointer paths and to-be-created paths;
--       a wrong pointer still fails deterministically at the implementer's
--       read step (modify must exist at ref). Upgrade path if pointer-curation
--       proves too costly: spawn the designer as a dedicated pod via an
--       orchestrator wrapper (mirror of fix-implementer-orchestrator) and add
--       it to the isRepoCloningAgent gate — a Go change, owner's call.
-- Deliberately NO new status values on site_work_items (the dedup-index/status
-- contract bit the fleet on 2026-07-16; approval lives in spec jsonb).
--
-- COUNCIL ROSTER: FIVE seats, mirroring the fix-proposer's live v8 chain
-- (the concept-register thread added review_reuse_agent and review_guidelines
-- on 2026-07-17 — PILOT_reuse_agent_reviewer.md,
-- PILOT_guidelines_agent_reviewer.md). Reuse is this builder's hard rule 1,
-- and the guidelines seat's wrapper-orchestrator/declared-contracts rules are
-- exactly what new workflows get wrong, so both bite hardest here.
-- Chain: editquality → bug_historian → reuse_agent → guidelines → guardian
-- (hard veto). Roster drift is expected — RUNBOOK_feature_builder.md A3:
-- re-mirror the live roster before applying this file.
--
-- DELIBERATE DELTA from fix-proposer: run_checks includes ALL reviewers'
-- checks (v6/v7 run only editquality+guardian). The advisory seats have no
-- veto — but their fact-shaped questions deserve answers like any other
-- reviewer's; answering checks feeds repropose evidence, it gates nothing.
--
-- MDL-039 GUARD (root ai_service SHADOWS step-level): this workflow carries NO
-- top-level ai_service key — every ai_service block is step-level. Keep it so.
-- Models: claude-sonnet-5 for design/repropose/reframe (the hardest generative
-- steps, matching diagnose-agent's move), claude-sonnet-4-6 for reviewers
-- (parity with the live fix-proposer council).

BEGIN;

-- Snapshot first if a live row exists (idempotent re-apply path).
SELECT snapshot_agent('feature-designer', 'pre-update: feature builder delta 1 v1 re-apply')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='feature-designer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'feature-designer',
    'Feature Designer (FB delta 1)',
    'Turns an owner-APPROVED capability spec (site_work_items capability_gap with owner_approval + code_pointers in spec) into a STAGED PLAN (plan_format staged-v1: ordered stages, per-stage allowlists incl. to-be-created files, image-before-seed checklist), reviewed by the same 5-seat council + deterministic router as the fix-proposer (editquality, bug-historian, reuse-agent, guidelines, guardian w/ hard veto). Writes no code; refuses unapproved or pointer-less specs. The stage-loop implementer (delta 2) takes approved staged plans to a branch + ONE PR.',
    'diagnose', 'coordinator', 'experimental',
    true, 1, '["diagnose", "feature-design"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_spec',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 1200,
      'steps', jsonb_build_object(

        'load_spec', jsonb_build_object(
          'action', 'query_database',
          'description', 'Pull the capability spec work item; approval computed in SQL so the conditional stays a string compare.',
          'output_field', 'spec_row',
          'next_step', 'check_spec_approved',
          'config', jsonb_build_object(
            'output_format', 'object',
            'params', jsonb_build_array('input_data.work_item_id'),
            'query',
              'SELECT id::text AS work_item_id, item_type, status, summary, ' ||
              '       spec::text AS spec_text, ' ||
              '       CASE WHEN item_type = ''capability_gap'' ' ||
              '                 AND spec ? ''owner_approval'' ' ||
              '                 AND spec ? ''code_pointers'' ' ||
              '            THEN ''approved'' ELSE ''unapproved'' END AS approval ' ||
              'FROM site_work_items WHERE id = $1::uuid'
          )
        ),

        -- THE GATE: only an owner-approved, pointer-carrying spec may seed a design.
        'check_spec_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Mirror of the proposer''s CONFIRMED gate: refuse anything without owner_approval AND code_pointers.',
          'config', jsonb_build_object(
            'condition', 'spec_row.approval == ''approved''',
            'then_step', 'load_schema_hint',
            'else_step', 'complete_refused'
          )
        ),

        -- F2.3b(a), reused verbatim: the LIVE schema the reviewers may write checks against.
        'load_schema_hint', jsonb_build_object(
          'action', 'query_database',
          'description', 'Live table/column list (information_schema) so reviewer checks stop hallucinating columns.',
          'output_field', 'schema_hint',
          'next_step', 'design',
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

        'design', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Draft the staged plan from the approved spec + its code pointers.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-5',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              -- Staged plans are sequences of edit plans; 8000 truncated single
              -- plans on the first proposer run, so a sequence gets 16000.
              'max_tokens', 16000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('spec_row'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — staged build plan (feature builder delta 1)' || chr(10) || chr(10) ||
'You are designing a STAGED BUILD PLAN from an owner-approved capability spec. You do NOT write code — you name an ordered sequence of stages, each a constrained edit plan a council can review and a caged implementer can execute one commit at a time.' || chr(10) || chr(10) ||
'## Hard rules' || chr(10) ||
'1. REUSE BEFORE RECREATE: every stage extends existing, named machinery. Name what each stage builds on; if the spec''s pointers show nothing to reuse, say so in risks rather than silently inventing a parallel mechanism.' || chr(10) ||
'2. ONLY THESE PATHS: edits may name (a) paths given in the spec''s code_pointers, (b) files this plan itself creates (operation "add"). You cannot read the repo — NEVER invent a path. If the pointers are insufficient to design the feature, output {"refusal": "<what pointer is missing and why>"} instead of a plan.' || chr(10) ||
'3. STAGES ARE COMMITS: array order is execution order; each stage must leave the tree buildable on its own (one branch-commit + build gate per stage). depends_on names earlier stages only.' || chr(10) ||
'4. CAPS: at most 6 stages, 8 edits per stage, 24 edits total. Beyond that the feature needs splitting — say so in risks and design the largest safe core.' || chr(10) ||
'5. SEEDS ARE FILES: anything DB-side (a new agent definition, a workflow config) ships as a .sql seed FILE in the PR — operation "add", artifact_role "seed" — NEVER as a live change. Every seed needs a post_merge_checklist entry. Whenever the plan ALSO ships code edits, an image_deploy entry must come strictly before any seed_apply (a seed naming an unregistered action fails at runtime); a plan with NO code edits needs no image_deploy entry — its truthful checklist is seed_apply then verify, do not invent an image step. A seed applied via ON CONFLICT (type, version) DO UPDATE already ACTIVATES the new definition atomically — do NOT add a separate stage or checklist step to "activate" or "bump the version"; applying the seed IS the activation.' || chr(10) ||
'5b. config_change is ONLY for editing an EXISTING live agent_definitions row that no seed re-inserts. Its "file" is a compact identifier of the form agent_definitions:<type> (e.g. agent_definitions:fix-implementer) — NO spaces, NO parentheses, NO repo path, NO prose. Its artifact_role is "code" or omit it — NEVER "seed"/"doc" (a config_change is not a repo file). config_change edits appear automatically in the PR body for the human; they are NOT post_merge_checklist entries (checklist acts are ONLY image_deploy | seed_apply | verify). If you find yourself wanting a config_change to "activate" a seeded version, you don''t — see rule 5.' || chr(10) ||
'6. Schema DDL only INSIDE a seed file, and any DDL a seed carries must be named in risks.' || chr(10) ||
'7. PLATFORM, not site data; MINIMAL; every edit CHANGES something (audits, comments and "no change required" are invalid); grounded_in quotes the spec verbatim.' || chr(10) ||
'8. expected_symbols per stage: names (functions, registry keys, SQL step names) that must literally appear in that stage''s produced files — the deterministic gate a reviewer can trust.' || chr(10) || chr(10) ||
'## The approved spec (work item ' || '{{.spec_row.work_item_id}}' || ')' || chr(10) || chr(10) ||
'{{.spec_row.summary}}' || chr(10) || chr(10) ||
'{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON (schema: SCHEMA_staged_plan_v1.md)' || chr(10) || chr(10) ||
'```json' || chr(10) ||
'{' || chr(10) ||
'  "plan_format": "staged-v1",' || chr(10) ||
'  "summary": "one paragraph: the capability and the shape of the build",' || chr(10) ||
'  "stages": [' || chr(10) ||
'    {"id": "s1", "title": "...", "goal": "what exists after this stage that did not before",' || chr(10) ||
'     "depends_on": [], ' || chr(10) ||
'     "edits": [{"file": "repo-relative/path", "symbol": "...", "operation": "modify|add|remove|config_change",' || chr(10) ||
'                "artifact_role": "code|seed|doc", "rationale": "...", "sketch": "..."}],' || chr(10) ||
'     "expected_symbols": ["..."], "gate": {"build": true}}' || chr(10) ||
'  ],' || chr(10) ||
'  "post_merge_checklist": [' || chr(10) ||
'    {"order": 1, "act": "image_deploy", "detail": "..."},' || chr(10) ||
'    {"order": 2, "act": "seed_apply", "file": "<seed file path>", "detail": "..."},' || chr(10) ||
'    {"order": 3, "act": "verify", "detail": "how the owner confirms it live"}' || chr(10) ||
'  ],' || chr(10) ||
'  "grounded_in": ["verbatim quotes from the spec"],' || chr(10) ||
'  "risks": "what could this break; what a reviewer should check"' || chr(10) ||
'}' || chr(10) ||
'```'
          )
        ),

        'persist_plan', jsonb_build_object(
          'action', 'diagnose_persist_fix_plan',
          'description', 'Structural validation (staged-v1 path: stage shape, cross-stage file discipline, seed/checklist contract) + write to diagnosis_artifacts (kind=fix_plan, plan_format staged-v1 — D1). A failed validation FAILS the run.',
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
          'description', 'Council reviewer 1/5 — edit quality, staged: real changes, minimality, stage boundaries, spec coverage.',
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
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: EDIT QUALITY (staged plan)' || chr(10) || chr(10) ||
'You review a proposed STAGED build plan against its approved spec. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) does every edit CHANGE something real (audits/comments are not edits); (b) does the stage list COVER the spec — quote any spec capability with no covering stage into missing; (c) are the stage BOUNDARIES right — each stage independently committable and buildable, dependencies declared, nothing that must land together split apart; (d) is the plan minimal — no stage that merely rearranges, no edit outside the spec''s scope; (e) do the edits use only the spec''s code_pointers paths or files the plan itself adds — flag any path that looks invented.' || chr(10) || chr(10) ||
'Verdicts: approve (sound), object (fixable problems — list them), veto (fundamentally wrong: builds a different capability, ignores the pointers, or all edits are no-ops).' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed before any revision and the results are fed back, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## The staged plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "editquality", "verdict": "approve|object|veto", "objections": [{"stage": "s1", "edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": ["spec capability with no covering stage"], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_bug_historian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 2/5 -- bug historian: documented failure-shape history vs this staged plan. Advisory only (no veto -- PILOT_bug_historian_reviewer.md).',
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
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'schema_hint'),
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
'This is a STAGED plan — a feature build, not a fix — so ALSO look for: a NEW rebuild/regeneration/render path being introduced without a fail-loud guard; a stage that regenerates or re-renders existing artifacts as a side effect; a seed that re-INSERTs or re-defines something live (DELETE+INSERT shape).' || chr(10) || chr(10) ||
'Judge the plan against the pattern above: (a) does any stage risk reproducing this pattern; (b) does the plan build a mechanism that will silently drop required-but-missing content rather than fail loudly; (c) is there a simpler, already-proven shape from a past occurrence a stage should reuse instead of inventing a new one.' || chr(10) || chr(10) ||
'Verdicts: approve (no known-pattern concern), object (this plan risks or incompletely addresses the documented recurring pattern above -- say which aspect and why, in objections). You do NOT have a veto -- if you see something severe enough that this build should not proceed at all, put it in objections at "high" severity and trust the router; if it is a true architecture-level concern, say so explicitly in notes so a human sees it either way.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} -- SELECT/WITH only, never writes. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## The staged plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "bug_historian", "verdict": "approve|object", "objections": [{"stage": "s1", "edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'review_reuse_agent', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 3/5 -- reuse agent: does the platform already have this, and did the plan check? Advisory only (no veto -- PILOT_reuse_agent_reviewer.md). Reuse is the feature builder''s hard rule 1, so this seat bites hardest here.',
          'output_field', 'review_reuse_agent',
          'next_step', 'review_guidelines',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: REUSE AGENT' || chr(10) || chr(10) ||
'You judge one thing: does this platform already have something that does what this plan is about to build, and did the plan check? You change nothing; you judge.' || chr(10) || chr(10) ||
'KNOWN DISCIPLINE: "Reuse before create" (STEP ZERO). This platform has a standing, explicitly-named discipline: before creating any agent, action, function, or migration, search agent_definitions, the action registry, Go code, and existing workflows for an equivalent -- and never create without first demonstrating no existing coverage exists.' || chr(10) || chr(10) ||
'Documented successes of following it: reused ssh_get_status as a monitor probe rather than writing a new one; reused ListObjects for resume logic; reused datahelpers.GetIntField over a custom helper; reused snapshot_agent() instead of a new side-table migration for backups.' || chr(10) || chr(10) ||
'THE FOUNDING INCIDENT for this review seat: a prior session reinvented a trigger+triage SQL pair that already existed elsewhere in the codebase -- duplicated work that a reuse check would have caught immediately.' || chr(10) || chr(10) ||
'A cautionary related case: this platform has at least one place (tool-lifecycle: "two divergent tool-creation paths") where two different code paths independently solve overlapping problems with inconsistent side effects -- not because either was wrong on its own, but because nobody unified them once both existed.' || chr(10) || chr(10) ||
'This is a STAGED plan -- a feature build, so creation is its whole business and the question sharpens: for EVERY stage that ADDS a function, action, agent, table, or workflow, (a) is there evidence in the stage''s rationale/grounded_in that an existing-coverage search was done; (b) do the spec''s own code_pointers already name an existing mechanism a stage should extend or call instead of duplicating; (c) would any stage create a second way to do something the platform already has one way to do.' || chr(10) || chr(10) ||
'Verdicts: approve (no reuse concern, or additions are genuinely novel), object (this plan risks duplicating existing coverage -- name what already exists and where, in objections). You do NOT have a veto.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on whether a table, column, action, or agent name already exists, put that query in checks as {"sql": "SELECT ...", "why": "..."} -- SELECT/WITH only. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## The staged plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "reuse_agent", "verdict": "approve|object", "objections": [{"stage": "s1", "edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "notes": "..."}'
          )
        ),

        'review_guidelines', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 4/5 -- guidelines agent: does the plan follow the platform''s documented conventions/contracts, and where it does not, is the plan wrong or the RULE? Advisory only (no veto -- PILOT_guidelines_agent_reviewer.md).',
          'output_field', 'review_guidelines',
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
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: GUIDELINES AGENT' || chr(10) || chr(10) ||
'You judge two things about this STAGED build plan: (1) does it FOLLOW the platform''s documented conventions and contracts; (2) where it appears not to, is that because the PLAN is wrong, or because the RULE is? You change nothing; you judge.' || chr(10) || chr(10) ||
'## The platform''s load-bearing rules (the ones people keep relearning)' || chr(10) ||
'- WRAPPER-ORCHESTRATOR: anything doing substantive work (LLM calls, crawls, heavy DB, minutes of runtime) must run in a spawned pod via a parent (processing_mode:"orchestrator" + spawn_agent), never inline on a shared chassis slot; file writes from non-spawned actions die with a random pod. A feature plan that seeds a NEW agent doing substantive work without this shape breaks the rule.' || chr(10) ||
'- WORK-ITEM DEDUP: site_work_items dedup is idx_swi_dedup UNIQUE(site_id, item_key) over NON-TERMINAL statuses; the terminal-status set is a contract (drift between it and the Go list breaks every keyed insert fleet-wide); use DELETE+INSERT, not ON CONFLICT.' || chr(10) ||
'- TRUTHFUL PROVENANCE: hand-made work items copy the real owning path''s metadata, deviate only truthfully (source=''manual'', real created_by), and take URLs from pages.url -- never invent a path.' || chr(10) ||
'- DECLARED CONTRACTS: any input a workflow reads must be declared in the agent''s input_contract; a call site''s input_mapping must satisfy the callee''s contract. A staged plan seeding a new workflow must declare what it reads.' || chr(10) ||
'- SCHEMA-SOURCE TIERS: a component field with required:true must set on_missing deliberately -- leaving it skip_field/empty hits the switch default and silently defers the whole section.' || chr(10) || chr(10) ||
'## The meta-rule for THIS seat (important)' || chr(10) ||
'A GUIDELINE-GAP is not a violation. If the spec/plan is correct but exposes a documented rule that is itself wrong or stale (this happens -- a runbook rule about max_tokens placement was recently found to be backwards), say so in notes as a recommended side-task / guideline amendment, and APPROVE. Do NOT object: forcing a correct plan to revise because the underlying rule is bad is the wrong move. Object ONLY when the PLAN breaks a rule that is right.' || chr(10) || chr(10) ||
'Verdicts: approve, object (a stage breaks a live rule -- name the rule and the stage in objections). You do NOT have a veto.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle, put that query in checks as {"sql": "SELECT ...", "why": "..."} -- SELECT/WITH only. Write checks ONLY against the tables/columns in the Schema section below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## The staged plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output -- ONLY this JSON' || chr(10) ||
'{"reviewer": "guidelines", "verdict": "approve|object", "objections": [{"stage": "s1", "edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "..."}], "notes": "..."}'
          )
        ),

        'review_guardian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer 5/5 — pipeline guardian, staged: per-stage blast radius, architecture signals, seed discipline. HARD VETO holder.',
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
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council reviewer: PIPELINE GUARDIAN (hard-veto holder, staged plan)' || chr(10) || chr(10) ||
'You protect the platform''s other pipelines from collateral damage. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) blast radius PER STAGE — which pipelines/workflows consume each edited file, workflow step, or seeded agent; does the plan acknowledge them; (b) architecture-change signals — edits to shared contracts, wire formats, message shapes, exported signatures, or MANY packages in one stage mean this is not a constrained build: veto and say it needs an architecture review; (c) seed discipline — anything DB-side must ship as a seed FILE with a checklist entry, image_deploy strictly before seed_apply; a plan that would change live state directly is a veto; (d) surface ownership — config_change edits must name the owning pipeline; (e) stage ordering — a stage must not depend on state (image or seed) that only exists after the owner''s post-merge acts.' || chr(10) || chr(10) ||
'Verdicts: approve, object (containable concerns), veto (cross-pipeline damage, live-state mutation, or architecture change dressed as a feature). Your veto BLOCKS. If you veto, name the safest contained alternative you can see in notes — it seeds the reframe.' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query could settle (a column''s type, whether a name exists in a table, a fleet-wide count), put that query in checks as {"sql": "SELECT ...", "why": "what this settles"} — SELECT/WITH only, never writes. Checks are executed before any revision and the results are fed back, so ask rather than assume. Write checks ONLY against the tables/columns in the Schema section below — a check against anything else fails and wastes the round. Two traps: workflow step definitions live in agent_definitions.default_config (jsonb — query with jsonb operators), there is NO steps table; a site''s domain lives on sites (join pages.site_id = sites.id). SQL cannot read Go source — do not ask code-shaped questions in checks; put those in objections for a human.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables available to checks)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## The staged plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON' || chr(10) ||
'{"reviewer": "guardian", "verdict": "approve|object|veto", "objections": [{"stage": "s1", "edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}'
          )
        ),

        'council_decide', jsonb_build_object(
          'action', 'diagnose_council_decide',
          'description', 'Deterministic aggregation, reused unchanged (it judges the artifact given — D1): veto→rejected, object→revise, else approved. Guardian holds the hard veto. Persists kind=council_report.',
          'output_field', 'council',
          'next_step', 'check_approved',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_bug_historian.result', 'review_reuse_agent.result', 'review_guidelines.result', 'review_guardian.result'),
            'hard_veto_from', jsonb_build_array('guardian'),
            'max_rounds', 3
          )
        ),

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

        'run_checks', jsonb_build_object(
          'action', 'diagnose_run_checks',
          'description', 'Run ALL FIVE reviewers'' checks (deliberate delta from fix-proposer, which runs two — see header) under the data_request containment; results feed repropose.',
          'output_field', 'check_results',
          'next_step', 'repropose',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'check_fields', jsonb_build_array('review_editquality.result.checks', 'review_bug_historian.result.checks', 'review_reuse_agent.result.checks', 'review_guidelines.result.checks', 'review_guardian.result.checks')
          )
        ),

        'repropose', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Revise the staged plan to address the council objections, then re-run persist + review + decide (the loop).',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-5',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 16000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'review_editquality', 'review_bug_historian', 'review_reuse_agent', 'review_guidelines', 'review_guardian', 'check_results'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — REVISE the staged build plan' || chr(10) || chr(10) ||
'A council reviewed your previous staged plan and asked for revision. Produce a NEW full plan (same staged-v1 JSON schema) that addresses every objection and covers every capability listed missing. You still write no code — you name stages and edits.' || chr(10) || chr(10) ||
'The SAME hard rules apply: only code_pointers paths or files this plan adds; stages are commits; caps (6 stages / 8 per stage / 24 total); seeds are files with an image-before-seed checklist; platform not site data; minimal; grounded in the spec; every edit CHANGES something.' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## Your previous plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Edit-quality reviewer said' || chr(10) || '{{.review_editquality.result}}' || chr(10) || chr(10) ||
'## Bug-historian reviewer said (advisory, no veto)' || chr(10) || '{{.review_bug_historian.result}}' || chr(10) || chr(10) ||
'## Reuse-agent reviewer said (advisory, no veto)' || chr(10) || '{{.review_reuse_agent.result}}' || chr(10) || chr(10) ||
'## Guidelines reviewer said (advisory, no veto)' || chr(10) || '{{.review_guidelines.result}}' || chr(10) || chr(10) ||
'## Guardian reviewer said (holds a hard veto)' || chr(10) || '{{.review_guardian.result}}' || chr(10) || chr(10) ||
'## Verification results (the reviewers'' own read-only queries, now answered)' || chr(10) || '{{.check_results.results_text}}' || chr(10) || chr(10) ||
'Use these results to SETTLE any objection that hinged on an unverified fact — cite them in grounded_in. If a result contradicts an edit, change or drop the edit; do not argue with the data.' || chr(10) || chr(10) ||
'## Output — ONLY the staged-v1 plan JSON. Address the objections; do not merely restate the old plan.'
          )
        ),

        'reframe', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'After a council VETO: produce a strictly narrower staged plan, or an explicit needs-architecture-review declaration plus the minimal safe preparatory stage. One attempt; then escalate.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-5',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 16000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('spec_row', 'plan_persisted', 'review_editquality', 'review_guardian'),
            'output_format', 'json',
            'prompt_template',
'# PROMPT — REFRAME after a council VETO' || chr(10) || chr(10) ||
'The council REJECTED your staged plan outright — the guardian judged it architecture-level change (or live-state mutation) dressed as a contained build. Do NOT resubmit the same shape: it will be vetoed again. Produce ONE of:' || chr(10) || chr(10) ||
'(a) a STRICTLY NARROWER staged plan the guardian could accept — prefer the reviewer''s own recommended alternative if its review names one; fewer stages, smaller blast radius, every DB-side artifact a seed file; or' || chr(10) || chr(10) ||
'(b) if no contained build exists at all, a plan whose only stages are the minimal safe preparatory steps, with risks stating plainly which decision the architecture review must take.' || chr(10) || chr(10) ||
'Same staged-v1 schema and remaining hard rules: only code_pointers paths or files this plan adds; caps; image-before-seed checklist; grounded in the spec; every edit CHANGES something.' || chr(10) || chr(10) ||
'## The approved spec' || chr(10) || '{{.spec_row.summary}}' || chr(10) || '{{.spec_row.spec_text}}' || chr(10) || chr(10) ||
'## Your VETOED plan' || chr(10) || '{{.plan_persisted.plan_json}}' || chr(10) || chr(10) ||
'## Edit-quality review' || chr(10) || '{{.review_editquality.result}}' || chr(10) || chr(10) ||
'## Guardian review (the veto — read its notes for the recommended alternative)' || chr(10) || '{{.review_guardian.result}}' || chr(10) || chr(10) ||
'## Output — ONLY the staged-v1 plan JSON.'
          )
        ),

        'escalate', jsonb_build_object(
          'action', 'diagnose_escalate',
          'description', 'Persist the human hand-off package: decision + spec + final staged plan + all five reviews.',
          'output_field', 'escalation',
          'next_step', 'complete_escalated',
          'config', jsonb_build_object(
            'fix_correlation_id', 'input_data.fix_correlation_id',
            'review_fields', jsonb_build_array('review_editquality.result', 'review_bug_historian.result', 'review_reuse_agent.result', 'review_guidelines.result', 'review_guardian.result')
          )
        ),

        'complete_escalated', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Escalated: no approvable staged plan. The hand-off package is in diagnosis_artifacts kind=escalation.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('escalation', 'council', 'plan_persisted'),
            'success_message', 'feature-designer escalated: no approvable staged plan; fetch the hand-off package from diagnosis_artifacts kind=escalation by correlation_id'
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'APPROVED. Staged plan + council report persisted; fetch both from diagnosis_artifacts by correlation_id. The stage-loop implementer (delta 2) takes an approved staged plan to a branch + ONE PR.',
          'config', jsonb_build_object('output_fields', jsonb_build_array('plan_persisted', 'council'))
        ),

        'complete_refused', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No plan: spec unapproved/pointer-less, designer refused or failed, or plan failed staged validation.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('spec_row'),
            'success_message', 'feature-designer made no plan: requires an owner-approved capability_gap spec carrying owner_approval + code_pointers, and a valid staged-v1 plan'
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
-- type='feature-designer' AND version=1;
