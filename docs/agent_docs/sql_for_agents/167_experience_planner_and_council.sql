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
              ' E''\n\n## Attached components — COMPLETE ground truth (page | slot | function | level | active)\n'' || ' ||
              ' COALESCE((SELECT string_agg(p.name||'' | ''||COALESCE(pc.slot_name,''-'')||'' | ''||COALESCE(cc.function,''(unlinked)'')||'' | ''||COALESCE(cc.component_level,''-'')||'' | ''||COALESCE(cc.is_active::text,''-''), E''\n'' ORDER BY p.name, pc.position) FROM page_components pc JOIN pages p ON p.id=pc.page_id LEFT JOIN content_components cc ON cc.id=pc.component_id WHERE p.site_id = $1::uuid), ''(none)'') || ' ||
              ' E''\n\n## Open work items (item_type | status | summary)\n'' || ' ||
              ' COALESCE((SELECT string_agg(item_type||'' | ''||status||'' | ''||left(summary,140), E''\n'') FROM site_work_items WHERE site_id = $1::uuid AND status NOT IN (''complete'',''verified'',''rejected'',''wont_fix'',''failed'',''cancelled'')), ''(none)'') || ' ||
              -- 174: the runtime loaders ARE the binding data contract. Without
              -- this section the council approved a §3 contract whose consumer
              -- it had never seen (archive[] vs data.archive.entries). Source,
              -- not a regex-extracted path list: the loaders alias their arg
              -- (var a = data.arena; … a.cards), so a regex silently misses
              -- paths and yields an authoritative-looking wrong answer.
              ' E''\n\n## Runtime JS loaders that hydrate this site (js_snippets) — THE BINDING DATA CONTRACT\n'' || ' ||
              ' E''Any data contract you specify MUST match the EXACT access paths in the source below. A field the loader never reads is dead weight; an access path the contract does not supply leaves that region SILENTLY EMPTY — these loaders fail gracefully by design, and runtime-fill shells are deliberately exempt from the dead-control and phantom-link checks, so NOTHING will flag it. Read the source. Never infer the shape from a component name.\n\n'' || ' ||
              ' COALESCE((SELECT string_agg(''### '' || js.name || E''\n'' || ''hydrates component(s): '' || js.applies_to::text || E''\n'' || COALESCE(js.description,'''') || E''\n```javascript\n'' || left(js.js_content, 8000) || CASE WHEN length(js.js_content) > 8000 THEN E''\n/* … TRUNCATED at 8000 chars — ask for the rest rather than guessing … */'' ELSE '''' END || E''\n```\n'', E''\n'' ORDER BY js.name) ' ||
              ' FROM js_snippets js WHERE js.is_active AND js.applies_to ?| (SELECT COALESCE(array_agg(DISTINCT cc.function), ARRAY[]::text[]) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id WHERE p.site_id = $1::uuid AND cc.function IS NOT NULL)), ' ||
              ' ''(no active runtime loader matches this site''''s attached components — if the plan assumes client-side hydration, that assumption is unverified)'') || ' ||
              -- 175: component-owned JS is a SECOND, distinct source of DOM
              -- contract (ships as /tools/assets/<function>.js). 174 surfaced
              -- only js_snippets, so gauntlet-interface's 3909 bytes of js_content
              -- stayed invisible and the contract critic could not verify
              -- Journey E. Third instance of this same omission class.
              ' E''\n\n## Component-owned JavaScript (content_components.js_content) — ALSO BINDING\n'' || ' ||
              ' E''Distinct from js_snippets above: this JS ships as /tools/assets/<function>.js and owns its component''''''''s DOM contract. If the plan names a selector or a computation for one of these components, it must appear in the source below. A selector the plan assumes and this source lacks does not exist.\n\n'' || ' ||
              ' COALESCE((SELECT string_agg(''### '' || cc.function || E'' (component-owned js)\n```javascript\n'' || left(cc.js_content, 8000) || CASE WHEN length(cc.js_content) > 8000 THEN E''\n/* … TRUNCATED at 8000 chars — ask for the rest rather than guessing … */'' ELSE '''' END || E''\n```\n'', E''\n'' ORDER BY cc.function) ' ||
              ' FROM (SELECT DISTINCT cc2.function, cc2.js_content FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc2 ON cc2.id=pc.component_id WHERE p.site_id = $1::uuid AND cc2.js_content IS NOT NULL AND cc2.js_content <> '''''''') cc), ' ||
              ' ''(no attached component carries its own js_content — any plan step assuming client-side behaviour from a component must say which script will provide it)'') ' ||
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
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',32000),
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
'D2 (REVISED 2026-07-18 by owner ruling — this supersedes any earlier "static detail pages" instruction) — per-provocation detail is rendered CLIENT-SIDE on the EXISTING archive page (/provocations/index.html), which is already a runtime-fill shell hydrating from /data/provocations.json. There are NO per-provocation static pages and NO new page_type in this MVP. Rationale, so you do not re-derive it: a prior council round established at HIGH severity that page_type=''provocation'' has zero prior rows and no proven build/render pipeline, so authoring one is not MVP work. Static per-provocation pages and the daily emitter both move to LATER.' || chr(10) ||
'D2 HARD CONSTRAINT — client-side detail must be a REAL, OBSERVABLE outcome, never a dead control. The original defect on this site was archive entries whose href was "#" with nothing behind it; you must not reproduce it in a new form. Opening an archive entry must produce an observable state change that an acceptance check can assert (for example: a deep-linkable URL fragment or query parameter, plus a detail region that becomes populated and visible with that entry''s real content from the feed). A control that only toggles a class, or that leads to an empty/placeholder region, is the same defect wearing a different coat. If a provocation in the feed has no detail content to show, the entry must not be presented as openable at all.' || chr(10) || chr(10) ||
'## The data contract shape (fixed by the existing client loader)' || chr(10) ||
'/data/provocations.json has keys: today {eyebrow, headline (may contain <em>), body, primary_cta{url,label}, secondary_cta{url,label}, stats[3]{value,label}}, lobby[≤4]{icon,title,desc,url}, arena {…reserved for the Arena widget}. Static site: no server; everything dynamic is client-side JS reading this JSON.' || chr(10) || chr(10) ||
'## HARD RULES' || chr(10) ||
'1. A not-yet feature is ABSENT or labelled coming-soon — NEVER simulated. No dead controls (href="#"/no-op), no fabricated numbers, no invented users, ever.' || chr(10) ||
'2. Every quantitative claim traces to the live feed or an evidence_base fact — vonc''s evidence_base currently has ZERO facts, so an MVP number must come from the feed/client-computed real state or not appear.' || chr(10) ||
'3. Reference REAL pages and selectors from the live context below; do not invent page names.' || chr(10) ||
'4. Tool-owned pages (rebuild_policy=owned) are rebuilt via the tool pipeline, never the generic page builder.' || chr(10) || chr(10) ||
'## Live site context' || chr(10) ||
'The "Attached components" list is the COMPLETE ground truth for this site: every component attached to every page, with its level and active flag. If a component is not in that list it is NOT attached. Do not claim a component exists, is active, or is "confirmed by query" unless the list shows it — and equally, do not object that something is unverifiable when the list settles it. Anything missing must be sequenced as a create/attach step.' || chr(10) || chr(10) ||
'{{.experience_context.text}}' || chr(10) || chr(10) ||
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
-- 176: compose renders the (now ~39KB, source-bearing) context inline and had
-- no length rule — only recompose did. It wrote past the 32000 output cap and
-- v1.0.1138's stop_reason decode turned that into a hard failure, killing the
-- run before the council. Cap NOT raised: approved plans are ~14KB.
'## LENGTH AND QUOTING DISCIPLINE (a run has already died here)' || chr(10) ||
'The site context above now includes REAL JAVASCRIPT SOURCE for the runtime loaders and component-owned scripts. That source is REFERENCE MATERIAL FOR YOU TO READ. It must NEVER be reproduced in the plan. Do not paste loader bodies, function definitions, or long code blocks into any section. When a data field or selector matters, name it and give the one-line access path (e.g. `data.archive.entries[]`, `.gauntlet-interface__timer`) — never the surrounding code.' || chr(10) || chr(10) ||
'The whole plan must come in WELL UNDER the output cap: aim for about 14,000 characters, and never exceed 20,000. A plan that hits the cap is DESTROYED, not shortened — the run fails outright with stop_reason=max_tokens and produces nothing. Approved plans for this experience have been ~14,000 characters, so this is comfortable, not tight.' || chr(10) || chr(10) ||
'Priority if you are running long: the ```criteria fence in §5 has ABSOLUTE priority and must always be complete, valid, closed JSON followed by the END trailer. Cut narrative, prose rationale and LATER-list detail first. A terse plan is fine; a cut-off plan is worthless.' || chr(10) || chr(10) ||
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
            -- ABSTENTION-TOLERANT (171): a flaky critic falls through to the NEXT
            -- critic, not to complete_refused. diagnose_council_decide reads an
            -- absent review field as an abstention, so this seat simply does not
            -- vote; the round still decides. Advisory seat — safe to lose.
            'error_step', 'review_feasibility',
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
'## Verdict discipline (IMPORTANT)' || chr(10) || 'Use "object" ONLY for a problem of MEDIUM or HIGH severity — something that would actually damage the experience if built as written. Record low-severity nits in "notes" and still verdict "approve". Any single objection forces a full revise round, so a low nit blocks the whole plan and burns a round; do not spend one on polish.' || chr(10) || chr(10) || '## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"journeys","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":["promise or control with no completing journey"],"checks":[{"sql":"SELECT ...","why":"..."}],"notes":"..."}'
          )
        ),

        'review_feasibility', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 2 — feasibility (veto).',
          'output_field', 'review_feasibility',
          'next_step', 'review_honesty',
          'config', jsonb_build_object(
            -- ABSTENTION-TOLERANT (171) — see review_journeys. This is the seat
            -- that actually died on bugs_open/008 item 5 in run 054b358a.
            'error_step', 'review_honesty',
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
'## Live site context' || chr(10) ||
'The "Attached components" list is the COMPLETE ground truth for this site: every component attached to every page, with its level and active flag. If a component is not in that list it is NOT attached. Do not claim a component exists, is active, or is "confirmed by query" unless the list shows it — and equally, do not object that something is unverifiable when the list settles it. Anything missing must be sequenced as a create/attach step.' || chr(10) || chr(10) ||
'{{.experience_context.text}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Verdict discipline (IMPORTANT)' || chr(10) || 'Use "object" ONLY for a problem of MEDIUM or HIGH severity — something that would actually damage the experience if built as written. Record low-severity nits in "notes" and still verdict "approve". Any single objection forces a full revise round, so a low nit blocks the whole plan and burns a round; do not spend one on polish.' || chr(10) || chr(10) || '## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"feasibility","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":["unbuilt dependency not sequenced"],"checks":[{"sql":"SELECT ...","why":"..."}],"notes":"..."}'
          )
        ),

        'review_honesty', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 3 — honesty auditor (HARD VETO).',
          'output_field', 'review_honesty',
          'next_step', 'review_mvp',
          'config', jsonb_build_object(
            -- DELIBERATELY NOT abstention-tolerant (171). This is the sole
            -- hard_veto seat. An absent review field reads as an abstention,
            -- so falling through would let a plan reach "approved" with the
            -- anti-fabrication gate never applied — precisely the failure the
            -- loop exists to catch (the Gauntlet's 12,847 fake competitors).
            -- A dead honesty auditor must refuse the run, not wave it through.
            -- Do NOT "make this consistent" with the other three.
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
'## Verdict discipline (IMPORTANT)' || chr(10) || 'Use "object" ONLY for a problem of MEDIUM or HIGH severity — something that would actually damage the experience if built as written. Record low-severity nits in "notes" and still verdict "approve". Any single objection forces a full revise round, so a low nit blocks the whole plan and burns a round; do not spend one on polish.' || chr(10) || chr(10) || '## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"honesty","verdict":"approve|object|veto","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":[],"checks":[],"notes":"the honest alternative if you veto"}'
          )
        ),

        'review_mvp', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 4 — MVP referee (advisory: approve|object only).',
          'output_field', 'review_mvp',
          'next_step', 'review_contracts',
          'config', jsonb_build_object(
            -- ABSTENTION-TOLERANT (171) — falls through to the NEXT critic,
            -- which since 174 is review_contracts, not council_decide.
            'error_step', 'review_contracts',
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
'## Verdict discipline (IMPORTANT)' || chr(10) || 'Use "object" ONLY for a problem of MEDIUM or HIGH severity — something that would actually damage the experience if built as written. Record low-severity nits in "notes" and still verdict "approve". Any single objection forces a full revise round, so a low nit blocks the whole plan and burns a round; do not spend one on polish.' || chr(10) || chr(10) || '## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never a section title, never quoted.' || chr(10) ||
'{"reviewer":"mvp","verdict":"approve|object","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":[],"checks":[],"notes":"..."}'
          )
        ),

        -- Critic 5, added by 174 after the council's first substantive escape:
        -- it unanimously approved a §3 data contract (archive[], arena{status}
        -- only) that the live loaders cannot read (data.archive.entries,
        -- a.cards), which would have silently blanked two runtime-fill regions.
        -- Its axis is deliberately NOT folded into feasibility — overlapping
        -- remits produce "each assumed the other checked it".
        'review_contracts', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Critic 5 — contract agreement / integration (approve|object; no veto, but objections block).',
          'output_field', 'review_contracts',
          'next_step', 'council_decide',
          'config', jsonb_build_object(
            'error_step', 'council_decide',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',8000),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('experience_context','proposal','schema_hint'),
            'output_format', 'json',
            'prompt_template',
'# Council critic: CONTRACT AGREEMENT (integration)' || chr(10) || chr(10) ||
'You judge ONE thing: wherever the plan names two artefacts that must agree, do they PROVABLY agree? You change nothing; you judge.' || chr(10) || chr(10) ||
'This platform''s most expensive defects are not bad code — they are two sides of one contract drifting apart, each side correct on its own. Real instances: a dedup index and a Go status list that had to match and did not (every keyed insert failed fleet-wide, invisibly); a component key required to be identical in three places (content_components.function, pages.name, doc subject_key) that was not (a tool''s acceptance criteria silently stopped covering it); two page-republish paths each assumed to publish a whole component.' || chr(10) || chr(10) ||
'Judge every producer/consumer pair the plan implies:' || chr(10) ||
'(a) DATA CONTRACT vs RUNTIME LOADER — for every field the plan defines, does a loader actually read it, and for every access path in the loader source, does the plan supply it? BOTH directions. A guard like `if (!data.x) return;` or `!Array.isArray(data.x.y)` is a HARD requirement: fail it and the region stays silently empty.' || chr(10) ||
'(b) COMPONENT vs BINDING — does every selector/field the plan names actually exist on the component it names?' || chr(10) ||
'(c) CONTROL vs DESTINATION — does every control the plan wires have a real destination?' || chr(10) ||
'(d) any other pair the plan asserts must match.' || chr(10) || chr(10) ||
'## THE RULE THAT MAKES THIS SEAT WORTH ITS COST' || chr(10) ||
'You may NOT approve a pair by reasoning about what a name implies. Quote the consumer''s own source from the context below. And if the source needed to settle a pair is NOT in your context, that is itself an OBJECTION — say exactly which source you needed and could not see. Never approve an unverifiable pair; an unseen consumer is how a mismatched contract reaches production.' || chr(10) || chr(10) ||
'## Live site context (includes the runtime loader sources — this is ground truth)' || chr(10) || '{{.experience_context.text}}' || chr(10) || chr(10) ||
'## Schema (the ONLY tables checks may use)' || chr(10) || '{{.schema_hint.text}}' || chr(10) || chr(10) ||
'## The plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'CHECKS: if a verdict hinges on a fact a read-only SQL query settles, put it in checks as {"sql":"SELECT ...","why":"..."} — SELECT/WITH only, ONLY the Schema tables.' || chr(10) || chr(10) ||
'Verdicts: approve (every pair provably agrees, and you quoted the proof), object (a pair does not agree, or you could not verify one — name it and quote both sides). You have NO veto, but the council decides by "ANY objection => revise", so your objection BLOCKS approval exactly as hard as a veto does. Do not spend one on style or naming preference — only on a pair that will not fit.' || chr(10) || chr(10) ||
'## Output — ONLY this JSON. Keep it COMPACT so it cannot truncate: at most 6 objections, each "problem" <= 240 chars, "notes" <= 400 chars, at most 3 checks. Close every brace. TYPE RULE: "edit" MUST be a bare INTEGER — the plan section number 1-5, or 0 for plan-wide. Never a string, never quoted.' || chr(10) ||
'{"reviewer":"contracts","verdict":"approve|object","objections":[{"edit":0,"problem":"...","severity":"low|medium|high"}],"missing":["contract half with no counterpart"],"checks":[{"sql":"SELECT ...","why":"..."}],"notes":"..."}'
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
            -- 174 added review_contracts. A seat missing from this list RUNS and
            -- is never read — its objections would vanish silently.
            'review_fields', jsonb_build_array('review_journeys.result','review_feasibility.result','review_honesty.result','review_mvp.result','review_contracts.result'),
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
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',32000),
            'temperature', 0.2,
            -- review_contracts added by 174. A critic absent from input_fields
            -- renders as <no value> and the reviser revises blind — bugs_open/016.
            'input_fields', jsonb_build_array('experience_context','proposal','review_journeys','review_feasibility','review_honesty','review_mvp','review_contracts','check_results','input_data'),
            'output_format', 'text',
            'prompt_template',
'# PROMPT — REVISE the EXPERIENCE_PLAN' || chr(10) || chr(10) ||
'The challenge council asked for revision. Produce a NEW full EXPERIENCE_PLAN (same five sections + criteria fence) that addresses EVERY objection and covers everything listed missing. Same hard rules: not-yet features absent or coming-soon never simulated; no dead controls; every number traces to the live feed or a real fact; reference real pages/selectors; tool pages via the tool pipeline.' || chr(10) || chr(10) ||
'## Your previous plan' || chr(10) || '{{.proposal}}' || chr(10) || chr(10) ||
'## Journeys critic said' || chr(10) || '{{.review_journeys}}' || chr(10) || chr(10) ||
'## Feasibility critic said' || chr(10) || '{{.review_feasibility}}' || chr(10) || chr(10) ||
'## Honesty auditor said (hard veto)' || chr(10) || '{{.review_honesty}}' || chr(10) || chr(10) ||
'## MVP referee said (NO veto, but its objections BLOCK approval just like any other seat — see SCOPE DISCIPLINE)' || chr(10) || '{{.review_mvp}}' || chr(10) || chr(10) ||
'## Contract-agreement critic said (NO veto, but its objections BLOCK — a mismatch here ships a silently empty region)' || chr(10) || '{{.review_contracts}}' || chr(10) || chr(10) ||
'## Verification results (the critics'' own read-only queries, now answered)' || chr(10) || '{{.check_results.results_text}}' || chr(10) || chr(10) ||
'SCOPE DISCIPLINE (run 7 escalated on exactly this): the council decides by "ANY objection => revise", so the MVP referee''s objection blocks approval as hard as a veto. Its scope cuts are NOT optional advice. For EVERY MVP-referee objection you must do one of two things, visibly: (a) MOVE the named item to the LATER list, or (b) keep it and state in §4, in one sentence, why the core loop genuinely cannot play without it. Silently keeping it is not an option and has already cost four rounds. Above all: answering one critic must NEVER enlarge the MVP cut to satisfy another — if fixing a feasibility or journeys objection requires adding scope, prefer CUTTING the feature that caused the objection over specifying it further. A smaller plan that all four seats approve beats a complete plan that never converges.' || chr(10) || chr(10) ||
'LENGTH DISCIPLINE (this loop has truncated before, which is itself an objection): revise by TIGHTENING and REPLACING text, never by appending. The plan must not grow round on round — if it is getting longer, cut narrative, prose rationale and LATER-list detail. The §5 criteria fence has ABSOLUTE priority: it must always be complete, valid, closed JSON followed by the END trailer, even if every other section has to be shortened to fit. A truncated plan is worse than a terse one.' || chr(10) || chr(10) || 'Use these to settle any objection that hinged on an unverified fact. If a result contradicts the plan, change the plan — do not argue with the data. Output the whole revised plan the same way: start with "# EXPERIENCE_PLAN — {{.experience_name}}", the five sections, the ```criteria fence, then a final line exactly <!-- END EXPERIENCE_PLAN --> after the closing ```. No preamble, no commentary.'
          )
        ),

        'reframe', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'After a veto (usually honesty): produce a plan that removes the fabrication/dead-end by DEMOTING that feature to honest coming-soon or absent — the smallest honest shape. One attempt, then escalate.',
          'output_field', 'proposal',
          'next_step', 'persist_plan',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object('model','claude-sonnet-5','provider','anthropic','api_key_env_var','ANTHROPIC_API_KEY','max_tokens',32000),
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
