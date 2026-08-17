-- 447_copy_editor_stage_two.sql
--
-- Seeds `copy-editor` — STAGE 2 of the two-stage copy pipeline
-- (`docs024_key_docs_latest/copy_quality_two_stage/PLAN_2026-08-12_two_stage_copy.md`).
--
-- WHAT STAGE 2 IS. Stage 1 (`page-content-writer`) is judged on facts, coverage,
-- structure, links and design classes — never on whether the prose reads like a
-- person wrote it, whether the most useful thing is first, or whether the page talks
-- about the site instead of the reader. `content-quality-auditor` raises that
-- judgement and nothing consumes it: design findings have an applier
-- (`css-patch-agent`), copy findings have none. This is that missing applier.
--
-- WHY IT IS CONFIG AND NOT GO. Every part already exists: `query_database` gives the
-- page-scoped read, `execute_llm_prompt` auto-injects the v2 house voice into any
-- template naming {{.voice_style}} (ai_actions.go:300-315), `checkpoint_for_review`
-- files the human-review item, and `section-editor` applies the edit. Nothing here
-- needs a binary, so nothing here waits for a roll.
--
-- THE FOUR RULES THAT MAKE IT SAFE, each measured rather than assumed:
--
--   1. PAGE-SCOPED READ, SECTION-SCOPED WRITE. A section-scoped reader cannot do this
--      job: the fleet lane's arm test wrote "Amortisation" in one section and
--      "amortization" in another because each section is written blind to its
--      siblings. `bugs_open/278` extends that to SELF-blindness — the same slot filled
--      twice, unable to see its own first pass. So load_page_components reads the
--      WHOLE page; the write goes back one component at a time.
--
--   2. LOCKS ARE STRUCTURAL, NOT INSTRUCTIONAL. `locked_at IS NULL` is in the SELECT,
--      so a locked component is never in scope to begin with. Proven three times: in
--      the 08-09 arm test BOTH prompt versions tried to overwrite the owner's
--      personally-approved opening and were stopped only by the lock — "not the
--      instructions, not the care taken writing them".
--
--   3. THE STAGE-1 BRIEF IS A FORBIDDEN INPUT (PLAN §1 corollary). An editor handed
--      the framing that produced the copy inherits it and re-writes the same page in
--      the same voice. Note what IS loaded: `content_direction.required_links`, the
--      declared SET — data, not framing. That distinction is the lane's own finding:
--      "the fix in the brief is to enumerate the set as data", because a prose
--      instruction to preserve a set is not reliably followed (measured 4×; the proof
--      case is a page that lost 6 links under exactly such an instruction, which is
--      still live in this component's own llm_guidance).
--
--   4. NO UNREVIEWED AUTO-REWRITE (owner decision D2, 2026-08-12). The workflow ENDS
--      at `checkpoint_for_review`. It cannot write to a page: no step in it can. The
--      approve path files a `section_edit` item for `section-editor`, which is the
--      only route that touches content_data. This preserves the guarantee already
--      written into `voice_tells` (HandlerAgent: "", "never an unreviewed
--      auto-rewrite") and is what keeps stage 2 out of architecture scope.
--
-- ⚠ THE KNOWN CONSEQUENCE, ACCEPTED WITH D2 AND NOT DISCOVERED AFTER IT: the human
-- review queue has no working surface (`bugs_open/033` — 735 items at
-- needs_human_review as at 2026-08-15, one type ever closed by a human). So this
-- agent's output parks until someone reads it. The owner ruled on 2026-08-15 to build
-- stage 2 in PARALLEL with the 033 thread rather than behind it, because stage 2's
-- first output is the committed proof case, which the owner reviews directly either
-- way. Until 033 lands, grade a proposal with the lane's gate and file the
-- `section_edit` by hand:
--   gate_stage2_edit.py --item <review_item_id>
--
-- ⚠ NOT YET EXERCISED. At seeding time this agent has never run. It is `experimental`
-- and nothing dispatches it: no item_type routes here, and it is deliberately absent
-- from the improvement sweep. Its first dispatch is a hand-filed canary (SCH-026
-- promotes born-`detected` items only for KNOWN-GOOD (type, handler) pairs; a new
-- pair is held for a human, and a hand-filed item must be born `triaged`).
--
-- ROLLBACK: 447_copy_editor_stage_two_ROLLBACK.sql (soft-deletes the row).

BEGIN;

INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status, is_active,
   idle_timeout_seconds, default_config)
VALUES (
  'copy-editor',
  'Copy Editor (stage 2)',
  'Stage 2 of the two-stage copy pipeline. Reads a WHOLE rendered page plus the site voice and the page''s declared link set, judges what stage 1 is not judged on (does this read like a person, is the most useful thing first, does it talk about the site instead of the reader), and proposes per-component field_updates. Never applies them: it ends at a human review checkpoint, and section-editor does the writing.',
  'specialist',
  'specialist',
  'experimental',
  true,
  600,
  jsonb_build_object(
    'workflow', jsonb_build_object(
      'start_step', 'ensure_site_record',
      'steps', jsonb_build_object(

        'ensure_site_record', jsonb_build_object(
          'action', 'ensure_site_record',
          'config', jsonb_build_object('store_brief_in_content_data', false),
          'output_field', 'site_record',
          'next_step', 'load_page_target'),

        -- The page and its DECLARED link set. Deliberately narrow: this selects
        -- required_links and nothing else from content_direction, because the rest of
        -- that column is the stage-1 brief (rule 3 above).
        'load_page_target', jsonb_build_object(
          'action', 'query_database',
          'config', jsonb_build_object(
            'query', 'SELECT p.id::text AS page_id, p.name AS page_name, s.domain, '
                  || 'COALESCE(p.content_direction->''required_links'', ''[]''::jsonb) AS required_links '
                  || 'FROM pages p JOIN sites s ON s.id = p.site_id WHERE p.id = $1',
            'params', jsonb_build_array('input_data.page_id'),
            'output_format', 'object'),
          'output_field', 'page_target',
          'next_step', 'load_page_components'),

        -- Rule 1 (page-scoped read) and rule 2 (the lock is in the SELECT).
        'load_page_components', jsonb_build_object(
          'action', 'query_database',
          'config', jsonb_build_object(
            'query', 'SELECT pc.id::text AS page_component_id, pc.slot_name, pc.position, '
                  || 'cc.name AS component_name, cc.input_schema AS declared_schema, '
                  || 'pc.content_data, pc.rendered_html '
                  || 'FROM page_components pc LEFT JOIN content_components cc ON cc.id = pc.component_id '
                  || 'WHERE pc.page_id = $1 AND pc.locked_at IS NULL '
                  || 'ORDER BY pc.position',
            'params', jsonb_build_array('input_data.page_id'),
            'output_format', 'rows'),
          'output_field', 'page_components',
          'next_step', 'run_copy_edit'),

        'run_copy_edit', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'config', jsonb_build_object(
            'ai_service', jsonb_build_object(
              'provider', 'anthropic',
              'model', 'claude-sonnet-5',
              'max_tokens', 16000,
              'api_key_env_var', 'ANTHROPIC_API_KEY'),
            'input_fields', jsonb_build_array('page_target', 'page_components'),
            'prompt',
'You are editing the copy of a page that is already live. It was written section by section, by a writer that could not see the other sections and was judged on facts, coverage and structure — never on whether it reads like a person wrote it. That is your job, and only yours.

{{.voice_style}}

PAGE: {{.page_target.domain}} /{{.page_target.page_name}}

THE PAGE''S COMPONENTS, in the order a reader meets them. Each carries its stored content_data (what you edit), its rendered HTML (what the reader sees) and its declared schema (what types you must return):
{{.page_components}}

LINKS THIS PAGE IS REQUIRED TO CARRY. This is the page''s own declaration, enumerated as data because a prose instruction to preserve a set is not reliably followed — this page lost six of them under exactly such an instruction:
{{.page_target.required_links}}

WHAT TO JUDGE, reading the whole page at once — these are the things a section-scoped writer structurally cannot see:
1. ORDER. Is the most useful thing to this reader first? A reader arrives with an errand. If the page opens by describing itself, or by explaining what it will explain, the useful part is buried.
2. THE READER, NOT THE SITE. Copy that inventories what the site contains talks about the site. Copy that answers what the reader came to do talks to the reader.
3. ONE NAME PER THING. Across the whole page, the same thing must be called the same thing.
4. REPETITION AND RESTATEMENT between sections — including a section restating its own neighbour, which the writer could not see.
5. THE TELLS. Defining by negation ("X, not Y"), stacked consequences, and openings that clear their throat before saying anything.

WHAT YOU MAY NOT DO. Every one of these is checked mechanically before your edit reaches a human, so an edit that breaks one is rejected, not discussed:
- You may not introduce a fact, a figure, a number, a claim or a link that is not already in the content you were given.
- You may not drop a link. Every href present must still be present, and every required link listed above must appear.
- You may not change markup: every class attribute and every structural element must survive. You are editing prose inside the existing HTML, not redesigning it.
- You may not change a field''s type. Return exactly the type the component''s declared schema names for that field — a declared array stays an array of the same shape; a declared html/text field stays a string.
- You may not touch a component that is not in the list above. Anything absent is locked and out of scope.
- If a section is already good, leave it out of your edits entirely. Rewriting for the sake of it is how the set-preservation failures happened.

Return ONLY JSON, no prose around it:
{"page_judgement":"two or three sentences on what is actually wrong with this page as a whole, or what is right if little is",
 "edits":[{"page_component_id":"<id from the list>","slot_name":"<slot>","field_updates":{"<field name>":"<the full new value for that field, same type as declared>"},"rationale":"what this changes and why a reader is better off"}],
 "no_change_needed":false}

If the page genuinely needs no editorial change, return an empty edits array and set no_change_needed to true. That is a real answer and a cheaper one than a rewrite nobody needed.'),
          'output_field', 'copy_edit',
          'next_step', 'request_review'),

        -- Rule 4. This is the LAST step that does anything: the workflow cannot write
        -- to a page, because no step in it can.
        'request_review', jsonb_build_object(
          'action', 'checkpoint_for_review',
          'config', jsonb_build_object(
            'item_type', 'copy_edit_proposed',
            'severity', 'medium',
            'site_id_from', 'site_record.site_id',
            'page_id_from', 'page_target.page_id',
            'review_fields_from', 'copy_edit.result',
            'summary_from', 'Stage-2 copy edit proposed for {{.page_target.domain}} /{{.page_target.page_name}} — review before it is applied',
            'on_approve', jsonb_build_object(
              'item_type', 'section_edit',
              'handler_agent', 'section-editor',
              'include_fields', jsonb_build_array('copy_edit', 'page_target'))),
          'output_field', 'review_item',
          'next_step', 'complete'),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'config', jsonb_build_object(),
          'next_step', NULL)
      ))));

-- Post-conditions, guarded. A verify block made of SELECTs cannot stop the COMMIT
-- (ON_ERROR_STOP ignores a non-empty result), so every assertion below RAISEs.
DO $$
DECLARE
  n int;
  cfg jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live copy-editor row, found %', n;
  END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'copy-editor' AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- Rule 4: no step may write to a page. If a later edit adds one, this fails loudly.
  IF EXISTS (SELECT 1 FROM jsonb_each(cfg->'workflow'->'steps') s(k, v)
              WHERE s.v->>'action' IN ('apply_section_edit','save_page_sections','render_component',
                                       'update_component_html','rerender_page_sections','git_commit')) THEN
    RAISE EXCEPTION 'copy-editor has a page-writing step — D2 forbids an unreviewed auto-rewrite';
  END IF;

  -- Rule 2: the lock must be in the SELECT, not in the prompt.
  IF (cfg->'workflow'->'steps'->'load_page_components'->'config'->>'query') NOT LIKE '%locked_at IS NULL%' THEN
    RAISE EXCEPTION 'load_page_components does not filter locked components structurally';
  END IF;

  -- The voice carrier must be referenced, or the agent silently writes in no voice
  -- at all (CQ-022: the block is injected ONLY into templates that name it).
  IF (cfg->'workflow'->'steps'->'run_copy_edit'->'config'->>'prompt') NOT LIKE '%{{.voice_style}}%' THEN
    RAISE EXCEPTION 'the prompt does not reference {{.voice_style}} — it would run with no house voice';
  END IF;

  -- Rule 3: the brief must not be loaded. Any select of content_direction other than
  -- the declared required_links array is the framing this agent is denied.
  IF EXISTS (SELECT 1 FROM jsonb_each(cfg->'workflow'->'steps') s(k, v)
              WHERE s.v->'config'->>'query' LIKE '%content_direction%'
                AND s.v->'config'->>'query' NOT LIKE '%content_direction->''required_links''%') THEN
    RAISE EXCEPTION 'a step loads the stage-1 brief — forbidden input (PLAN §1 corollary)';
  END IF;

  RAISE NOTICE 'copy-editor seeded: 4 rules asserted (no page write, structural lock, voice carrier, no brief)';
END $$;

COMMIT;
