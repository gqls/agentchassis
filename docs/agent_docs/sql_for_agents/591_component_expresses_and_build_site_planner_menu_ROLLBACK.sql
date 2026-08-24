-- ROLLBACK for 591 — surgical inverse. REFUSES rather than mis-splicing.
--
-- It removes exactly what 591 inserted: the derived column from the menu SELECT, the
-- evidence-base clause from the WHERE, the [expresses: …] token from the listing
-- line, and rule 19 from the prompt. It does NOT drop component_expresses() —
-- 592/593 also call it, and dropping a function three menus depend on in order to
-- undo one of them is how a rollback becomes an outage. Drop it by hand, after
-- 592 and 593 are also rolled back, with:
--     DROP FUNCTION IF EXISTS component_expresses(text, jsonb);
--
-- If any literal has moved, this file RAISEs and changes nothing. That is the
-- intended behaviour: a rollback that guesses is worse than one that refuses.

BEGIN;

DO $$
DECLARE n int; q text; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '591 ROLLBACK: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF position('array_to_string(component_expresses(html_template, input_schema), '', '') AS expresses' in q) = 0 THEN
    RAISE EXCEPTION '591 ROLLBACK: the derived column is not present verbatim — 591 was not applied, or something has edited the menu query since. Refusing.';
  END IF;
  IF position('requires-evidence-base' in q) = 0 THEN
    RAISE EXCEPTION '591 ROLLBACK: the evidence-base clause is not present — refusing rather than half-reverting';
  END IF;
  IF position('{{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}' in p) = 0 THEN
    RAISE EXCEPTION '591 ROLLBACK: the listing token is not present verbatim — refusing';
  END IF;
  IF position('19. MATCH STRUCTURE TO PROMISE.' in p) = 0 THEN
    RAISE EXCEPTION '591 ROLLBACK: rule 19 is not present — refusing';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_components,config,query}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,load_components,config,query}',
               $newg$ AND (NOT (COALESCE(semantic_tags, '[]'::jsonb) ? 'requires-evidence-base') OR EXISTS (SELECT 1 FROM site_specs ss_eb WHERE ss_eb.site_id = $1 AND ss_eb.aspect ILIKE '%evidence%' AND ss_eb.is_current))$newg$,
               $oldg$$oldg$
             ),
             $new$, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$,
             $old$ FROM content_components$old$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_site,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
               $newl${{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}$newl$,
               $oldl$$oldl$
             ),
             $newa$19. MATCH STRUCTURE TO PROMISE. Each component above carries what it can EXPRESS: `list` (renders a bulleted or numbered list), `table` (renders a table), `items` (renders a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` (paragraphs, and nothing else). A `[prose only]` section CANNOT render a list no matter what is written for it — the markup is not in its template — and the writer will silently flatten the promise into paragraphs. So: if a page or a section you are planning promises a month-by-month calendar, a step-by-step process, a checklist, a comparison, or a specification, at least one section on that page MUST express `list`, `table`, `items` or `html-block`. This is the difference between a page that keeps its own heading's promise and one that reads as padding. Do not pad the other way either: a page with nothing enumerable on it is right to be all prose.

Return ONLY valid JSON.$newa$,
             $olda$Return ONLY valid JSON.$olda$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';
  IF position('component_expresses' in q) > 0 OR position('requires-evidence-base' in q) > 0 THEN
    RAISE EXCEPTION '591 ROLLBACK VERIFY: the menu query still carries 591 content';
  END IF;
  IF position('expresses' in p) > 0 OR position('MATCH STRUCTURE TO PROMISE' in p) > 0 THEN
    RAISE EXCEPTION '591 ROLLBACK VERIFY: the prompt still carries 591 content';
  END IF;
  IF position('SELECT name, display_name, "function", category, description FROM content_components' in q) = 0 THEN
    RAISE EXCEPTION '591 ROLLBACK VERIFY: the original SELECT list was not restored';
  END IF;
  IF (length(p) - length(replace(p, 'Return ONLY valid JSON.', ''))) / length('Return ONLY valid JSON.') <> 1 THEN
    RAISE EXCEPTION '591 ROLLBACK VERIFY: the closing anchor is not present exactly once';
  END IF;
  RAISE NOTICE '591 ROLLBACK OK: build-site-planner menu and prompt restored';
END $$;

COMMIT;
