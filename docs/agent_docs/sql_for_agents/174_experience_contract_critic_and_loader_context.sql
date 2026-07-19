-- 174_experience_contract_critic_and_loader_context.sql
-- Experience Loop — close the council's first substantive escape.
--
-- WHAT ESCAPED (2026-07-19, found while starting T4 Step 0). The council
-- unanimously approved an EXPERIENCE_PLAN whose §3 data contract does not match
-- the live client loaders:
--   §3 says `archive[]` (top-level array); `fillArchive` guards on
--     `!Array.isArray(data.archive.entries)` → early return, archive never fills.
--   §3 says `arena` = `{status:"coming_soon"}` only; `fillLobbyGrid` reads
--     `Array.isArray(a.cards) ? a.cards : []` → zero cards filled.
--   §3 defines a top-level `lobby[≤4]`; nothing reads it.
-- Implementing it verbatim would silently blank two of three runtime-fill
-- regions — silently, because the loaders fail gracefully by design and G22
-- exempts runtime-fill shells from the dead-control and phantom-link checks.
--
-- ROOT CAUSE: the same class already fixed once. `load_context` was taught to
-- surface component ATTACHMENTS, but the loaders live in `js_snippets`, which
-- nothing surfaced. The council could see the components and not the JavaScript
-- that hydrates them, so feasibility approved a contract whose consumer it had
-- never been shown. Context lying by omission, one table over.
--
-- TWO CHANGES, SHIPPED TOGETHER ON PURPOSE. Context without a seat told to use
-- it repeats the escape (feasibility's remit already covered "is every datum
-- available at runtime" and it still approved). A seat without the context would
-- object to everything. Half-applying this is worse than not applying it.
--
--  (1) load_context gains a "Runtime JS loaders" section: for every ACTIVE
--      js_snippets row whose applies_to matches a component attached to this
--      site, the loader's real SOURCE. Deliberately the source, not a
--      regex-extracted list of access paths: the loaders alias their argument
--      (`var a = data.arena; … a.cards`), so a regex would silently miss paths
--      and produce an authoritative-looking, incomplete list — the very failure
--      being fixed. Capped at 8000 chars per loader with a visible marker.
--
--  (2) A fifth seat, `review_contracts`, whose single axis is: wherever the plan
--      names two artefacts that must agree, do they provably agree? This is a
--      distinct axis from feasibility (can ONE thing be built) and it is aimed
--      at this platform's dominant defect family — contract drift between two
--      sides that are each correct alone (the idx_swi_dedup/Go status-list
--      split, the function/pages.name/subject_key three-way key mismatch, the
--      two rerender paths). Deliberately NOT folded into feasibility as a fifth
--      sub-check: overlapping remits produce "each assumed the other checked it".
--
--      Its load-bearing instruction is that it may not approve a pair by
--      reasoning about what a name implies — it must quote the consumer's own
--      source — AND that a pair it cannot verify from context is itself an
--      objection. That second clause is what makes the council self-healing
--      about context gaps: a future omission surfaces as a visible objection
--      instead of a silent approval.
--
--      approve|object only (no veto): a contract mismatch is fixable by revising
--      the contract, so `revise` is the right route; `veto` routes to reframe,
--      which is for unsalvageable shape. Per migration 172's lesson the prompt
--      states plainly that its objections BLOCK, since decideCouncil returns
--      revise on ANY object — "advisory" must never again read as "optional".
--
-- Config-only: live on commit, no image roll. Seed 167 carries the same changes
-- in-place so a re-apply cannot clobber this.

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 174 contract critic + loader context')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

-- ── (1) load_context: append the runtime-loader section ──────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_context,config,query}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'load_context'->'config'->>'query',
           '''(none)'')  AS text',
           '''(none)'') || '
           || ' E''\n\n## Runtime JS loaders that hydrate this site (js_snippets) — THE BINDING DATA CONTRACT\n'' || '
           || ' E''Any data contract you specify MUST match the EXACT access paths in the source below. A field the loader never reads is dead weight; an access path the contract does not supply leaves that region SILENTLY EMPTY — these loaders fail gracefully by design, and runtime-fill shells are deliberately exempt from the dead-control and phantom-link checks, so NOTHING will flag it. Read the source. Never infer the shape from a component name.\n\n'' || '
           || ' COALESCE((SELECT string_agg(''### '' || js.name || E''\n'' || ''hydrates component(s): '' || js.applies_to::text || E''\n'' || COALESCE(js.description,'''') || E''\n```javascript\n'' || left(js.js_content, 8000) || CASE WHEN length(js.js_content) > 8000 THEN E''\n/* … TRUNCATED at 8000 chars — ask for the rest rather than guessing … */'' ELSE '''' END || E''\n```\n'', E''\n'' ORDER BY js.name) '
           || ' FROM js_snippets js WHERE js.is_active AND js.applies_to ?| (SELECT COALESCE(array_agg(DISTINCT cc.function), ARRAY[]::text[]) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id WHERE p.site_id = $1::uuid AND cc.function IS NOT NULL)), '
           || ' ''(no active runtime loader matches this site''''s attached components — if the plan assumes client-side hydration, that assumption is unverified)'')'
           || '  AS text'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── (2) the fifth seat ───────────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,review_contracts}',
         jsonb_build_object(
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
         true)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── (3) wire it in: review_mvp → review_contracts → council_decide ───────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,review_mvp,next_step}', '"review_contracts"'::jsonb, false),
         '{workflow,steps,review_mvp,config,error_step}', '"review_contracts"'::jsonb, false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── (4) council_decide must actually count the new seat ─────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,council_decide,config,review_fields}',
         '["review_journeys.result","review_feasibility.result","review_honesty.result","review_mvp.result","review_contracts.result"]'::jsonb,
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── (5) recompose must SEE the new critic, or it revises blind (bugs_open/016) ─
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,recompose,config,input_fields}',
           '["experience_context","proposal","review_journeys","review_feasibility","review_honesty","review_mvp","review_contracts","check_results","input_data"]'::jsonb,
           false),
         '{workflow,steps,recompose,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'recompose'->'config'->>'prompt_template',
           '## Verification results (the critics'' own read-only queries, now answered)',
           '## Contract-agreement critic said (NO veto, but its objections BLOCK — a mismatch here ships a silently empty region)' || chr(10) || '{{.review_contracts}}' || chr(10) || chr(10) ||
           '## Verification results (the critics'' own read-only queries, now answered)'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── Assert the whole wiring, including the parts easy to half-apply ─────────
DO $$
DECLARE
  w jsonb; q text; rp text;
BEGIN
  SELECT default_config->'workflow' INTO w
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  q  := w->'steps'->'load_context'->'config'->>'query';
  rp := w->'steps'->'recompose'->'config'->>'prompt_template';

  IF position('js_snippets' in q) = 0 THEN
    RAISE EXCEPTION 'load_context does not surface js_snippets';
  END IF;
  IF position('Attached components' in q) = 0 THEN
    RAISE EXCEPTION 'load_context lost its attached-components section';
  END IF;
  IF w->'steps'->'review_contracts' IS NULL THEN
    RAISE EXCEPTION 'review_contracts seat missing';
  END IF;
  IF w->'steps'->'review_mvp'->>'next_step' <> 'review_contracts' THEN
    RAISE EXCEPTION 'review_mvp does not chain into review_contracts';
  END IF;
  IF w->'steps'->'review_contracts'->>'next_step' <> 'council_decide' THEN
    RAISE EXCEPTION 'review_contracts does not chain into council_decide';
  END IF;
  IF NOT (w->'steps'->'council_decide'->'config'->'review_fields' @> '["review_contracts.result"]'::jsonb) THEN
    RAISE EXCEPTION 'council_decide does not count review_contracts — the seat would run and never be read';
  END IF;
  IF position('{{.review_contracts}}' in rp) = 0 THEN
    RAISE EXCEPTION 'recompose cannot see review_contracts output (bugs_open/016 shape)';
  END IF;
  IF position('SCOPE DISCIPLINE' in rp) = 0 OR position('LENGTH DISCIPLINE' in rp) = 0 THEN
    RAISE EXCEPTION 'recompose lost an existing discipline rule';
  END IF;
  -- honesty must STILL be the only seat that refuses on error (171)
  IF w->'steps'->'review_honesty'->'config'->>'error_step' <> 'complete_refused' THEN
    RAISE EXCEPTION 'review_honesty lost its fail-closed error_step (171)';
  END IF;
END $$;

COMMIT;

-- Rollback: restore the snapshot taken above. Partial rollback is NOT safe —
-- removing the seat without removing it from review_fields leaves council_decide
-- counting a field nothing writes (a permanent abstention).
