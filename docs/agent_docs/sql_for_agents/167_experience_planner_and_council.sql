-- 167_experience_planner_and_council.sql — Phase 3 of the Experience Loop
-- (RUNBOOK T3.1/T3.2). Creates the `experience-planner` agent: ONE workflow
-- that composes an EXPERIENCE_PLAN (doc_plans, subject_type='experience'),
-- then runs a four-critic challenge council over it and converges-or-escalates,
-- modelled verbatim on the fix-proposer v6 pattern (sequential execute_llm_prompt
-- reviewers → deterministic diagnose_council_decide → conditional router).
--
-- REUSE, not new Go: compose/critics = execute_llm_prompt; persist =
-- write_doc_plan (subject_type='experience', enabled by migration 163 + the
-- docResolveSubject gate fix); aggregation = diagnose_council_decide (verbatim);
-- verification = diagnose_run_checks; round record = doc_notes
-- (category experience-council) + diagnosis_artifacts kind=council_report.
--
-- Critics (PLAN §3-B), veto rights chosen against diagnose_council_decide's
-- rule "ANY veto rejects" (so a veto = a full gatekeeper):
--   journeys      — every control has a named destination + observable outcome;
--                   no step ends in '#'.                          [veto]
--   feasibility   — buildable with current components/tool-generator; data
--                   available at runtime; static-hosting constraint honoured. [veto]
--   honesty       — no invented stats/users/social proof anywhere. HARD VETO
--                   (fabrication is the cardinal rule; hard_veto_from=['honesty']).
--   mvp-referee   — cuts scope to the core playable loop. ADVISORY (approve|object
--                   only — the bug-historian precedent; an MVP opinion must not
--                   gate on its own).
--
-- ██ DEPLOY SEQUENCING ██ — apply ONLY AFTER the chassis image carrying the
-- docResolveSubject 'experience' fix (commit 66d32477d) is live. On an older
-- image, persist_plan fails "subject_type must be 'tool' or 'pipeline'" and the
-- run refuses. Order: image → this file → fire (092_TRIGGER_experience_plan.sh).
-- Everything else the workflow calls (execute_llm_prompt, write_doc_plan,
-- append_doc_note, query_database, diagnose_council_decide, diagnose_run_checks,
-- conditional, complete_workflow) is already registered fleet-wide.
--
-- No root ai_service (MDL-039: a root block would SHADOW every per-step model).
-- Applied out of band (psql -f + ledger row same sitting per
-- bugs_open/aaa_fails_to_mend/007).

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 167 initial experience-planner + council')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'experience-planner',
    'Experience Planner + Challenge Council',
    'Composes an EXPERIENCE_PLAN (doc_plans subject_type=experience: journeys, promise ledger, data contracts, MVP cut, journey criteria) for a named experience, then runs a four-critic challenge council (journeys/feasibility/honesty[hard-veto]/mvp-referee) and converges-or-escalates. Reuses diagnose_council_decide verbatim; writes travelling docs, not code.',
    'experience', 'coordinator', 'experimental',
    true, 1, '["experience-planning","council"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'load_context',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 1200,
      'steps', jsonb_build_object(

        'load_context', jsonb_build_object(
          'action', 'query_database',
          'description', 'Live site facts the plan must reference: pages (+rebuild_policy), tool components, open work items.',
          'output_field', 'experience_context',
          'next_step', 'load_schema_hint',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'output_format', 'object',
            'params', jsonb_build_array('input_data.site_id'),
            'query',
              'SELECT ' ||
              ' E''## Pages (name | type | url | rebuild_policy | build_status)\n'' || ' ||
              ' COALESCE((SELECT string_agg(name||'' | ''||COALESCE(page_type,''-'')||'' | ''||url||'' | ''||COALESCE(rebuild_policy,''generic'')||'' | ''||COALESCE(build_status,''-''), E''\n'' ORDER BY name) FROM pages WHERE site_id = $1::uuid), ''(none)'') || ' ||
              ' E''\n\n## Tool components (function | active)\n'' || ' ||
              ' COALESCE((SELECT string_agg(DISTINCT cc.function||'' | ''||cc.is_active::text, E''\n'') FROM content_components cc JOIN page_components pc ON pc.component_id=cc.id JOIN pages p ON p.id=pc.page_id WHERE p.site_id = $1::uuid AND cc.component_level=''tool''), ''(none)'') || ' ||
              ' E''\n\n## Open work items (item_type | status | summary)\n'' || ' ||
              ' COALESCE((SELECT string_agg(item_type||'' | ''||status||'' | ''||left(summary,140), E''\n'') FROM site_work_items WHERE site_id = $1::uuid AND status NOT IN (''complete'',''verified'',''rejected'',''wont_fix'',''failed'',''cancelled'')), ''(none)'') ' ||
              ' AS text'
          )
        ),

        'load_schema_hint', jsonb_build_object(
          'action', 'query_database',
          'description', 'Live table/column list so critic checks[] stop hallucinating columns.',
          'output_field', 'schema_hint',
          'next_step', 'compose',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'output_format', 'object',
            'params', jsonb_build_array(),
            'query',
              'SELECT string_agg(t.line, chr(10)) AS text FROM ( ' ||
              '  SELECT table_name || ''('' || string_agg(column_name || '' '' || data_type, '', '' ORDER BY ordinal_position) || '')'' AS line ' ||
              '  FROM information_schema.columns ' ||
              '  WHERE table_schema = ''public'' AND table_name IN ' ||
              '    (''pages'',''sites'',''site_plans'',''site_plan_pages'',''site_work_items'', ' ||
              '     ''content_components'',''page_components'',''doc_plans'',''doc_notes'',''site_specs'') ' ||
              '  GROUP BY table_name ORDER BY table_name ' ||
              ') t'
          )
        ),

        'compose', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Draft the EXPERIENCE_PLAN body (5 sections + a criteria fence) from the diagnosis, decisions, data contract and live context.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',16000),
            'temperature', 0.2,
            'input_fields', jsonb_build_array('experience_context','input_data'),
            'output_format', 'text',
            'prompt_template',
'# PROMPT — compose the EXPERIENCE_PLAN' || chr(10) || chr(10) ||
'You own the WHOLE experience, not a page or a tool in isolation: the promise every button makes, the journey a visitor takes, the data a widget needs, the honesty of every number on the page. Write an EXPERIENCE_PLAN for the "{{.experience_name}}" experience on {{.experience_domain}}. This document travels; a build round and an acceptance ladder will execute it verbatim.' || chr(10) || chr(10) ||
'## The diagnosis you are fixing (three broken surfaces, artifact-verified 2026-07-17)' || chr(10) ||
'1. /provocations/index.html — archive entries are runtime-filled into a template whose href="#" was never given a destination; per-provocation detail pages were never planned (needs_page:provocation, now owned_page_review).' || chr(10) ||
'2. /tools/arena/index.html — the tool-arena widget does NOT fetch /data/provocations.json (the feed is live, 200/5.6KB) so it shows "Loading… DAY 0" forever.' || chr(10) ||
'3. /tools/gauntlet/index.html — a mock: both CTAs href="#", fabricated stats (12,847 competitors, 94,210 challenges, a "Live" leaderboard of invented users). build_status=needs_rebuild.' || chr(10) || chr(10) ||
'## Decisions already made (owner-accepted; do NOT relitigate)' || chr(10) ||
'D1 — the Gauntlet ships as a MINIMAL-REAL playable round: a timed round against today''s provocation, client-side scoring, NO leaderboard, zero fabricated numbers. Only demote it to an honest "coming soon" if it genuinely cannot be built both honest AND small — and then it must be ABSENT or labelled coming-soon, never simulated.' || chr(10) ||
'D2 — per-provocation detail pages are STATIC, emitted from the feed. NOTE: the daily emitter for /data/provocations.json is NOT built yet (the live file is a hand-committed sample). Your MVP cut MUST decide explicitly: either the emitter is in the MVP, or the MVP builds detail pages for the current sample feed and the emitter is the first LATER item. State which.' || chr(10) || chr(10) ||
'## The data contract shape (fixed by the existing client loader)' || chr(10) ||
'/data/provocations.json has keys: today {eyebrow, headline (may contain <em>), body, primary_cta{url,label}, secondary_cta{url,label}, stats[3]{value,label}}, lobby[≤4]{icon,title,desc,url}, arena {…reserved for the Arena widget}. Static site: no server; everything dynamic is client-side JS reading this JSON.' || chr(10) || chr(10) ||
'## HARD RULES' || chr(10) ||
'1. A not-yet feature is ABSENT or labelled coming-soon — NEVER simulated. No dead controls (href="#"/no-op), no fabricated numbers, no invented users, ever.' || chr(10) ||
'2. Every quantitative claim traces to the live feed or an evidence_base fact — vonc''s evidence_base currently has ZERO facts, so an MVP number must come from the feed/client-computed real state or not appear.' || chr(10) ||
'3. Reference REAL pages and selectors from the live context below; do not invent page names.' || chr(10) ||
'4. Tool-owned pages (rebuild_policy=owned) are rebuilt via the tool pipeline, never the generic page builder.' || chr(10) || chr(10) ||
'## Live site context' || chr(10) || '{{.experience_context.text}}' || chr(10) || chr(10) ||
'## Write the plan as markdown with EXACTLY these sections' || chr(10) ||
'### 1. Journeys' || chr(10) || 'Each journey is an ordered list of steps; every step names: page, control (a real CSS selector), action, and the OBSERVABLE outcome. No step may end at "#".' || chr(10) ||
'### 2. Promise ledger' || chr(10) || 'A table: CTA copy → the page/state the destination must deliver. (e.g. "Enter the Gauntlet" → a playable timed round actually starts.)' || chr(10) ||
'### 3. Data contracts' || chr(10) || 'What /data/provocations.json must contain for this experience, who writes it and when, and what is client-side-only. Name the emitter decision from D2. If the experience computes ANY number a visitor will read as a score/metric, define the EXACT computation and the honest meaning of its label here — an undefined "score" is a soft fabrication, and an acceptance check that only asserts "some digits appeared" cannot tell a real computation from an arbitrary one.' || chr(10) ||
'### 4. MVP cut + LATER' || chr(10) || 'The round-1 scope (the smallest honest playable loop) and an explicit LATER list. Restate the D1/D2 constraints as they apply to the cut. Write the MVP cut as an ORDERED, GATED step list: any prerequisite DATA step (e.g. committing /data/provocations.json with real today/lobby/arena content) is step 0 with an explicit gate ("do not proceed until it returns 200 with real content"), because a rebuild has nothing to fetch until it exists. Every later step names what must be true before it starts, and any claim that an existing work item is already resolved must be re-verified at build time, not merely cited. A dependency mentioned only in prose is NOT sequenced.' || chr(10) ||
'### 5. Acceptance criteria' || chr(10) || 'A fenced ```criteria block of JSON the runner executes, using ONLY these check types (multi-page journeys are described narratively in §1; the runner journey type is a later phase):' || chr(10) ||
'   {"profiles":["desktop","mobile"],"container":".tool-container","checks":[' || chr(10) ||
'     {"id":"...","type":"selector_exists","selector":".real-selector"},' || chr(10) ||
'     {"id":"...","type":"asset_loads","path":"/data/provocations.json"},' || chr(10) ||
'     {"id":"...","type":"interaction","steps":[{"action":"fill|click|select","selector":"#real","value":"x"}],"expect":{"selector":"#real","text_matches":"..."}},' || chr(10) ||
'     {"id":"...","type":"page_status_ok"},{"id":"...","type":"no_horizontal_overflow","profiles":["mobile"]}]}' || chr(10) ||
'   Every selector must be one you expect the built page to actually have.' || chr(10) || chr(10) ||
'## Output format (IMPORTANT)' || chr(10) ||
'Output the whole plan as markdown: start with "# EXPERIENCE_PLAN — {{.experience_name}}", the five sections, the ```criteria fence, and then — AFTER the closing ``` of that fence — one final line exactly: <!-- END EXPERIENCE_PLAN -->. No preamble before the "#", no commentary. (The trailer line is required so the criteria fence is preserved verbatim in storage.)'
          )
        ),

        'persist_plan', jsonb_build_object(
          'action', 'write_doc_plan',
          'description', 'Supersede-write the EXPERIENCE_PLAN (doc_plans, subject_type=experience). Each revise round supersedes the prior; the final approved plan is is_current.',
          'output_field', 'plan_persisted',
          'next_step', 'review_journeys',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'subject_type', 'experience',
            'subject_key_field', 'input_data.experience_key',
            'plan_body_field', 'proposal.result',
            'plan_source', 'experience-planner',
            'created_by', 'experience-planner'
          )
        ),

        'review_journeys', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 1 — journey completeness (veto).',
          'output_field', 'review_journeys',
          'next_step', 'review_feasibility',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('experience_context','proposal','schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council critic: JOURNEY COMPLETENESS (holds a veto)' || chr(10) || chr(10) ||
'You judge the EXPERIENCE_PLAN on ONE axis: does every journey actually complete? You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) does EVERY clickable control named in §1/§2 have a real named destination and an OBSERVABLE outcome — no step ends at "#"; (b) does every promise in the ledger (§2) have a journey step that delivers it; (c) are the acceptance criteria (§5) selectors plausibly present on the built page and do they actually exercise the journey, not just assert a container exists; (d) are there orphaned controls (a button with no journey) or orphaned journeys (a step with no control).' || chr(10) || chr(10) ||
'Verdicts: approve (every journey completes), object (fixable gaps — list them), veto (a core journey has no destination at all, i.e. the plan would ship another dead end).' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query settles (does page X exist, is component Y active), put it in checks as {"sql":"SELECT ...","why":"..."} — SELECT/WITH only. Write checks ONLY against the tables/columns in the Schema below. SQL cannot read the plan JSON or Go — judge those directly.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables checks may use)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"journeys","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":["promise or control with no completing journey"],"checks":[{"sql":"SELECT ...","why":"..."}],"notes":"..."}'
          )
        ),

        'review_feasibility', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 2 — feasibility (veto).',
          'output_field', 'review_feasibility',
          'next_step', 'review_honesty',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('experience_context','proposal','schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council critic: FEASIBILITY (holds a veto)' || chr(10) || chr(10) ||
'You judge whether the MVP cut can actually be BUILT with what this platform has, on a static host. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) can each MVP piece be built with the existing components / tool-generator / page pipeline (tool pages rebuilt via the tool pipeline, not the generic builder); (b) is every datum a widget needs available AT RUNTIME on a static site — i.e. in /data/provocations.json or client-computable — with NO server; (c) does the plan depend on anything unbuilt (e.g. the daily emitter) without saying so and sequencing it; (d) is the client-side scoring/timer for the minimal-real Gauntlet genuinely doable without a backend.' || chr(10) || chr(10) ||
'Verdicts: approve (buildable as scoped), object (buildable after fixable changes — list them), veto (the MVP as written cannot be built honestly on this stack — say what is impossible).' || chr(10) || chr(10) ||
'CHECKS: same rules — SELECT/WITH only, ONLY the Schema tables below.' || chr(10) || chr(10) ||
'## Schema (the ONLY tables checks may use)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## Live site context' || chr(10) || '{{.experience_context.text}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"feasibility","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":["unbuilt dependency not sequenced"],"checks":[{"sql":"SELECT ...","why":"..."}],"notes":"..."}'
          )
        ),

        'review_honesty', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 3 — honesty auditor (HARD VETO).',
          'output_field', 'review_honesty',
          'next_step', 'review_mvp',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('proposal','schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council critic: HONESTY AUDITOR (hard veto — fabrication is the cardinal sin)' || chr(10) || chr(10) ||
'You judge ONE thing: does the plan invent anything presented as real? You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) any quantitative claim (counts, rates, streaks, "N players") that does NOT trace to the live feed or an evidence_base fact — vonc''s evidence_base has ZERO facts, so ANY hard number that is not read from real client state is fabricated; (b) any invented user, name, leaderboard entry, testimonial, or social-proof ("everyone else has filed"); (c) any control or label that claims a capability the MVP will not actually deliver ("LIVE NOW" over a non-working widget); (d) any not-yet feature that is SIMULATED rather than absent or labelled coming-soon.' || chr(10) || chr(10) ||
'This is the anti-fabrication rule that the current Gauntlet violates (12,847 competitors, a fake Live leaderboard). The plan must not reproduce it in any form.' || chr(10) || chr(10) ||
'Verdicts: approve (nothing fabricated), object (a fixable honesty slip — name it), veto (the plan bakes in fabricated stats/users/social-proof, or simulates a not-yet feature). Your veto BLOCKS. If you veto, name the honest alternative in notes (e.g. "show real submitted-position count from the feed, or omit the stat").' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"honesty","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":[],"checks":[],"notes":"the honest alternative if you veto"}'
          )
        ),

        'review_mvp', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 4 — MVP referee (advisory: approve|object only).',
          'output_field', 'review_mvp',
          'next_step', 'council_decide',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('proposal'),
            'output_format', 'json',
            'prompt_template',
'# Council critic: MVP REFEREE (advisory — you have NO veto)' || chr(10) || chr(10) ||
'You cut scope. You judge whether the MVP cut is the SMALLEST honest playable loop, and challenge anything not needed for the core loop to be playable. You change nothing; you judge.' || chr(10) || chr(10) ||
'Judge: (a) is anything in the MVP cut that could be moved to LATER without breaking the core loop (land on a provocation → file a position → see the day''s record; enter a real timed Gauntlet round); (b) is the cut trying to build too much at once; (c) is the LATER list honest about what is deferred (esp. the daily emitter).' || chr(10) || chr(10) ||
'Verdicts: approve (tight and playable), object (over-scoped — say exactly what to defer). You do NOT have a veto; put anything severe in objections at high severity and trust the router.' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"mvp","verdict":"approve|object","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":[],"checks":[],"notes":"..."}'
          )
        ),

        'council_decide', jsonb_build_object(
          'action', 'diagnose_council_decide',
          'description', 'Deterministic aggregation (reused verbatim): any veto→rejected, any object→revise, else approved. Honesty holds the hard veto. Persists kind=council_report; sets should_revise/should_reframe.',
          'output_field', 'council',
          'next_step', 'append_council_note',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'fix_correlation_id', 'input_data.experience_correlation_id',
            'review_fields', jsonb_build_array('review_journeys.result','review_feasibility.result','review_honesty.result','review_mvp.result'),
            'hard_veto_from', jsonb_build_array('honesty'),
            'max_rounds', 5
          )
        ),

        'append_council_note', jsonb_build_object(
          'action', 'append_doc_note',
          'description', 'Record this round''s council verdict in the travelling docs (category experience-council). The full per-round record is diagnosis_artifacts kind=council_report.',
          'output_field', 'council_note',
          'next_step', 'check_approved',
          'config', jsonb_build_object(
            'error_step', 'check_approved',
            'subject_type', 'experience',
            'subject_key_field', 'input_data.experience_key',
            'note_body_field', 'council.decided_by',
            'note_site_id_field', 'input_data.site_id',
            'note_categories', jsonb_build_array('experience-council')
          )
        ),

        'check_approved', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 1/3: approved is the only path to complete.',
          'config', jsonb_build_object('condition', 'council.decision == ''approved''', 'then_step', 'complete', 'else_step', 'check_rejected')
        ),
        'check_rejected', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 2/3: a veto reframes once or escalates — never reproposes the same shape.',
          'config', jsonb_build_object('condition', 'council.decision == ''rejected''', 'then_step', 'check_reframe', 'else_step', 'check_revise')
        ),
        'check_reframe', jsonb_build_object(
          'action', 'conditional',
          'description', 'First veto with rounds left → one reframe; second veto / spent cap → escalate.',
          'config', jsonb_build_object('condition', 'council.should_reframe == true', 'then_step', 'reframe', 'else_step', 'complete_escalated')
        ),
        'check_revise', jsonb_build_object(
          'action', 'conditional',
          'description', 'Router 3/3: revise with rounds left → answer critic checks then recompose; exhausted → escalate.',
          'config', jsonb_build_object('condition', 'council.should_revise == true', 'then_step', 'run_checks', 'else_step', 'complete_escalated')
        ),

        'run_checks', jsonb_build_object(
          'action', 'diagnose_run_checks',
          'description', 'Answer the critics'' read-only checks (SELECT/WITH) before recompose, so a fact-shaped objection is settled with evidence.',
          'output_field', 'check_results',
          'next_step', 'recompose',
          'config', jsonb_build_object(
            'error_step', 'recompose',
            'check_fields', jsonb_build_array('review_journeys.result.checks', 'review_feasibility.result.checks')
          )
        ),

        'recompose', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Revise the plan to address every objection + missing item; loops back to persist + review + decide.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',16000),
            'temperature', 0.2,
            'input_fields', jsonb_build_array('experience_context','proposal','review_journeys','review_feasibility','review_honesty','review_mvp','check_results','input_data'),
            'output_format', 'text',
            'prompt_template',
'# PROMPT — REVISE the EXPERIENCE_PLAN' || chr(10) || chr(10) ||
'The challenge council asked for revision. Produce a NEW full EXPERIENCE_PLAN (same five sections + criteria fence) that addresses EVERY objection and covers everything listed missing. Same hard rules: not-yet features absent or coming-soon never simulated; no dead controls; every number traces to the live feed or a real fact; reference real pages/selectors; tool pages via the tool pipeline.' || chr(10) || chr(10) ||
'## Your previous plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Journeys critic said' || chr(10) || '{{.review_journeys}}' || chr(10) || chr(10) ||
'## Feasibility critic said' || chr(10) || '{{.review_feasibility}}' || chr(10) || chr(10) ||
'## Honesty auditor said (hard veto)' || chr(10) || '{{.review_honesty}}' || chr(10) || chr(10) ||
'## MVP referee said (advisory)' || chr(10) || '{{.review_mvp}}' || chr(10) || chr(10) ||
'## Verification results (the critics'' own read-only queries, now answered)' || chr(10) || '{{.check_results.results_text}}' || chr(10) || chr(10) ||
'Use these to settle any objection that hinged on an unverified fact. If a result contradicts the plan, change the plan — do not argue with the data. Output the whole revised plan the same way: start with "# EXPERIENCE_PLAN — {{.experience_name}}", the five sections, the ```criteria fence, then a final line exactly <!-- END EXPERIENCE_PLAN --> after the closing ```. No preamble, no commentary.'
          )
        ),

        'reframe', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'After a veto (usually honesty): produce a plan that removes the fabrication/dead-end by DEMOTING that feature to honest coming-soon or absent — the smallest honest shape. One attempt, then escalate.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',16000),
            'temperature', 0.2,
            'input_fields', jsonb_build_array('experience_context','proposal','review_journeys','review_feasibility','review_honesty','input_data'),
            'output_format', 'text',
            'prompt_template',
'# PROMPT — REFRAME after a council VETO' || chr(10) || chr(10) ||
'The council VETOED your plan — a critic judged it either dishonest (fabrication) or a dead end (a core journey with no destination). Do NOT resubmit the same shape. Produce a plan where the offending feature is made HONEST AND SMALL: the minimal-real version if one exists, otherwise ABSENT or an explicit "coming soon" label (never simulated). Prefer the alternative the vetoing critic named in its notes.' || chr(10) || chr(10) ||
'Same five sections + criteria fence, same hard rules. If the Gauntlet is what was vetoed and no honest minimal-real round is achievable, demote it to a labelled coming-soon panel and move the real round to the LATER list — that is an acceptable honest MVP.' || chr(10) || chr(10) ||
'## Your VETOED plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Honesty auditor (read its notes for the honest alternative)' || chr(10) || '{{.review_honesty}}' || chr(10) || chr(10) ||
'## Journeys critic' || chr(10) || '{{.review_journeys}}' || chr(10) || chr(10) ||
'## Feasibility critic' || chr(10) || '{{.review_feasibility}}' || chr(10) || chr(10) ||
'Output the whole reframed plan the same way: start with "# EXPERIENCE_PLAN — {{.experience_name}}", the five sections, the ```criteria fence, then a final line exactly <!-- END EXPERIENCE_PLAN --> after the closing ```. No preamble, no commentary.'
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'APPROVED: the EXPERIENCE_PLAN is is_current in doc_plans (subject_type=experience); the council trail is in doc_notes (experience-council) + diagnosis_artifacts (council_report). CP2. The MVP build round (T4) reads this plan.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('plan_persisted','council'),
            'success_message', 'experience-planner CONVERGED: EXPERIENCE_PLAN approved and current; proceed to the MVP build round.'
          )
        ),

        'complete_escalated', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'ESCALATED: no approvable plan within max_rounds. The current (rejected) plan stays is_current but MUST NOT be built until resolved; the disagreement is the round-boundary decision menu. Read council_report artifacts + doc_notes (experience-council).',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('plan_persisted','council'),
            'success_message', 'experience-planner ESCALATED: council did not converge; surface the disagreement (council_report by correlation_id) as the decision menu; do NOT build the current plan.'
          )
        ),

        'complete_refused', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'No plan: a step failed (compose/persist/critic/context). Nothing approved.',
          'config', jsonb_build_object(
            'output_fields', jsonb_build_array('experience_context'),
            'success_message', 'experience-planner made no plan: a compose/persist/critic step failed.'
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

-- Rollback: DELETE FROM agent_definitions WHERE type='experience-planner' AND version=1;
