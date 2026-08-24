-- 595 — page-content-writer: RULE 10 is addressed to a type that exists (bugs_open/381, arm B2)
--
-- WHY. The writer's rulebook has exactly two sentences about markup, and between them
-- they instruct every prose slot on the estate into paragraphs:
--
--   RULE 9  "For fields of type `text`: return a plain string with no HTML wrapping."
--   RULE 10 "For fields of type `rich_text` or `content` that contain multiple
--            paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags"
--
-- `[MEASURED 2026-08-24]` across active section components: 940 llm fields are typed
-- `text` (135 components), 2 are `html`, and `rich_text` and `content` are declared
-- **ZERO** times. So the only rule that permits structure is addressed to nobody, the
-- rule that forbids it covers essentially everything, and the most structure RULE 10
-- would have permitted even if it were reachable is a <p>.
--
-- WHAT CHANGES. Migration 594 retypes four pass-through prose slots to `html`. This
-- file re-addresses RULE 10 to `html` and tells the writer what structure means, and
-- narrows RULE 9 so it stops contradicting a field's own guidance.
--
-- ⚠ RULE 9 IS NARROWED, NOT WEAKENED. Its real job — and 304's addition to it — is the
-- markdown ban, which stays absolute and is restated. What it stops asserting is that
-- a `text` field may never contain markup, because that was already false in practice:
-- `article-body.content` is typed `text`, carries guidance asking for headings and
-- lists, and renders a list in 76% of its instances. RULE 9 was being ignored where a
-- field's own guidance contradicted it, so the rulebook is being made to match what
-- the estate already does rather than the reverse.
--
-- THE EXAMPLES ARE DELIBERATELY NOT FIRST-PERSON PRACTICE CLAIMS. "Months of the year,
-- steps a reader takes, options being compared" — never "our testing process, step by
-- step". Coordinated with the bugs_open/380 lane, whose writer arm forbids exactly
-- that shape on a site with no operating history; an example inviting it here would
-- have pulled against their fix.
--
-- ⚠ 304's ROLLBACK WILL REFUSE AFTER THIS FILE, BY DESIGN. 304
-- (`304_forbid_markdown_in_text_fields.sql:65`) guards on the literal
-- "10. For fields of type `rich_text` or `content`", which this file replaces. That is
-- a refusal, not a corruption: 304's rollback declines rather than mis-splicing. This
-- file's own rollback restores both rules verbatim, so the pair is recoverable.
--
-- ⚠ TIMING, from the 305 lane (2026-08-24). Their define-by-negation scanner splits
-- sentences on `</p`, `<br`, `</li`, `</h`, `</div`, `</td` — and `</th` was MISSING
-- (it is not covered by the `</h` arm: the third character differs). A
-- define-by-negation construction inside a table HEADER cell therefore produced a
-- markup-bearing "sentence", and their repair splices over exactly that span, which
-- would have replaced the cell tags with prose and broken the table. Found by asking
-- them about this change; fixed by that lane in 714789d7b, mutation-proven — but it is
-- Go, so it is INERT UNTIL THE NEXT CHASSIS ROLL. PREFER APPLYING THIS FILE AFTER THAT
-- ROLL. The exposure is narrow (it needs the construction inside a <th> specifically),
-- lists and subheads were already safe, and nothing here is unsafe on its own.
--
-- PAIRS WITH 594. Guidance is the lever that moves behaviour (see 594's header); this
-- file removes the rulebook's contradiction of it. Neither ordering is unsafe.
--
-- SCOPE. Config-only, one prompt, two anchored replaces. LIVE ON APPLY, no roll.
-- Anchored replace() with exact-count prechecks — never a whole-prompt rewrite, which
-- would silently revert the bugs_open/380 lane's concurrent edits to the same prompt
-- (their anchors: the evidence_base writer_block block; mine: rules 9 and 10 — disjoint,
-- agreed 2026-08-24).
--
-- ROLLBACK: 595_writer_rules_9_and_10_name_html_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('page-content-writer', '595_writer_rules_9_and_10_name_html: pre-update');

DO $$
DECLARE n int; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '595: expected exactly 1 live page-content-writer row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';
  IF p IS NULL THEN
    RAISE EXCEPTION '595: generate_content.config.prompt_template is NULL — the loop sub-workflow has changed under me, refusing';
  END IF;

  IF (length(p) - length(replace(p, '9. For fields of type `text`: return a plain string with no HTML wrapping.', '')))
     / length('9. For fields of type `text`: return a plain string with no HTML wrapping.') <> 1 THEN
    RAISE EXCEPTION '595: RULE 9 does not appear exactly once verbatim — refusing to splice blind';
  END IF;
  IF (length(p) - length(replace(p, '10. For fields of type `rich_text` or `content` that contain multiple paragraphs:', '')))
     / length('10. For fields of type `rich_text` or `content` that contain multiple paragraphs:') <> 1 THEN
    RAISE EXCEPTION '595: RULE 10 does not appear exactly once verbatim — 304 or another migration has moved it, refusing';
  END IF;
  -- 304's markdown ban must still be inside rule 9; this file restates it and would
  -- otherwise silently drop it.
  IF position('Plain string also means NO markdown syntax' in p) = 0 THEN
    RAISE EXCEPTION '595: 304''s markdown ban is missing from RULE 9 — the prompt is not the shape this file was written against, refusing';
  END IF;
  IF position('10. For fields of type `html`' in p) > 0 THEN
    RAISE EXCEPTION '595: already applied — refusing to double-apply';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
               $old10$10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.$old10$,
               $new10$10. For fields of type `html`: write real structure, not a run of paragraphs. Use <h3> for a subheading each time the material turns to a new point (in a long block, roughly every 150 words); <p> for paragraphs; <ul> or <ol> with <li> wherever the content is genuinely enumerable — the months of the year, the steps a reader takes, the things to check before deciding, the options being compared; <strong> for the term a reader is scanning for; <table> only when the data really is tabular; <blockquote> for a quotation. Do NOT pad to earn a list: if the material is one idea, one paragraph is the right answer, and a three-item list of near-identical clauses is worse than the paragraph it replaced. Never use <h1> or <h2> — the section supplies its own heading — and never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>. Where this field's own description gives more specific guidance, that description wins.$new10$
             ),
             $old9$9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.$old9$,
             $new9$9. For fields of type `text`: return a plain string — a heading, a label, a sentence — with no HTML wrapping around it. The template handles the wrapping for these fields. If this field's own description asks for particular markup inside it, follow the description: it knows what its slot renders. (Structure belongs in `html` fields — see rule 10.)$new9$
           )
         )
       ),
       updated_at = now()
 WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8'
   AND type = 'page-content-writer'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';

  IF position('10. For fields of type `html`: write real structure' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: RULE 10 was not re-addressed to html';
  END IF;
  IF position('rich_text' in p) > 0 THEN
    RAISE EXCEPTION '595 VERIFY: the prompt still names rich_text — a type no component declares';
  END IF;
  IF position('9. For fields of type `text`: return a plain string — a heading' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: RULE 9 was not narrowed';
  END IF;
  -- 304's markdown ban must have SURVIVED: it is a separate sentence appended to
  -- rule 9 and the replace above must not have swallowed it.
  IF position('Plain string also means NO markdown syntax' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: 304''s markdown ban was destroyed by the rule 9 replace';
  END IF;
  IF position('<ul> or <ol> with <li> wherever the content is genuinely enumerable' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: the list instruction is absent';
  END IF;
  IF position('never emit <img>, <figure>, <iframe>' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: the forbidden-elements sentence is absent';
  END IF;
  -- The two rules must each appear exactly once and in order.
  IF position('9. For fields of type `text`' in p) > position('10. For fields of type `html`' in p) THEN
    RAISE EXCEPTION '595 VERIFY: rules 9 and 10 are out of order';
  END IF;
  -- Nothing else in the prompt was disturbed: three landmarks that belong to other
  -- lanes and must still be there.
  IF position('STRICT RULE -- NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE.' in p) = 0
     OR position('19. Never write the words' in p) = 0
     OR position('{{range .current_section.llm_field_specs}}' in p) = 0 THEN
    RAISE EXCEPTION '595 VERIFY: an unrelated part of the prompt is missing — this file must only touch rules 9 and 10';
  END IF;

  RAISE NOTICE '595 OK: rule 10 addresses html and asks for structure; rule 9 narrowed, markdown ban intact';
END $$;

COMMIT;
