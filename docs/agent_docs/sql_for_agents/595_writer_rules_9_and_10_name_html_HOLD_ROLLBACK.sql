-- ROLLBACK for 595 — surgical inverse; REFUSES rather than mis-splicing.
-- Restores rules 9 and 10 verbatim as they stood after 304. Note that once this has
-- run, 304's own rollback becomes applicable again.
BEGIN;

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';
  IF p IS NULL THEN
    RAISE EXCEPTION '595 ROLLBACK: prompt_template is NULL — refusing';
  END IF;
  IF position('10. For fields of type `html`: write real structure' in p) = 0 THEN
    RAISE EXCEPTION '595 ROLLBACK: rule 10 is not in its post-595 form — 595 was not applied, or something has edited it since. Refusing.';
  END IF;
  IF position('9. For fields of type `text`: return a plain string — a heading' in p) = 0 THEN
    RAISE EXCEPTION '595 ROLLBACK: rule 9 is not in its post-595 form — refusing rather than half-reverting';
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
               $new10$10. For fields of type `html`: write real structure, not a run of paragraphs. Use <h3> for a subheading each time the material turns to a new point (in a long block, roughly every 150 words); <p> for paragraphs; <ul> or <ol> with <li> wherever the content is genuinely enumerable — the months of the year, the steps a reader takes, the things to check before deciding, the options being compared; <strong> for the term a reader is scanning for; <table> only when the data really is tabular; <blockquote> for a quotation. Do NOT pad to earn a list: if the material is one idea, one paragraph is the right answer, and a three-item list of near-identical clauses is worse than the paragraph it replaced. Never use <h1> or <h2> — the section supplies its own heading — and never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>. Where this field's own description gives more specific guidance, that description wins.$new10$,
               $old10$10. For fields of type `rich_text` or `content` that contain multiple paragraphs: use proper HTML markup, wrapping each paragraph in <p> tags: <p>Paragraph 1</p><p>Paragraph 2</p>.$old10$
             ),
             $new9$9. For fields of type `text`: return a plain string — a heading, a label, a sentence — with no HTML wrapping around it. The template handles the wrapping for these fields. If this field's own description asks for particular markup inside it, follow the description: it knows what its slot renders. (Structure belongs in `html` fields — see rule 10.)$new9$,
             $old9$9. For fields of type `text`: return a plain string with no HTML wrapping. The template handles paragraph wrapping for these fields.$old9$
           )
         )
       ),
       updated_at = now()
 WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';

DO $$
DECLARE p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO p FROM agent_definitions WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';
  IF position('10. For fields of type `rich_text` or `content` that contain multiple paragraphs:' in p) = 0 THEN
    RAISE EXCEPTION '595 ROLLBACK VERIFY: rule 10 was not restored';
  END IF;
  IF position('9. For fields of type `text`: return a plain string with no HTML wrapping.' in p) = 0 THEN
    RAISE EXCEPTION '595 ROLLBACK VERIFY: rule 9 was not restored';
  END IF;
  IF position('Plain string also means NO markdown syntax' in p) = 0 THEN
    RAISE EXCEPTION '595 ROLLBACK VERIFY: 304''s markdown ban is missing';
  END IF;
  IF position('For fields of type `html`' in p) > 0 THEN
    RAISE EXCEPTION '595 ROLLBACK VERIFY: 595 content still present';
  END IF;
  RAISE NOTICE '595 ROLLBACK OK: rules 9 and 10 restored to their post-304 text';
END $$;

COMMIT;
