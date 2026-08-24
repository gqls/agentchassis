-- ROLLBACK for 592 — surgical inverse; REFUSES rather than mis-splicing.
-- Does NOT drop component_expresses(): 591 and 593 also depend on it.
BEGIN;

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';
  IF q IS NULL OR position('array_to_string(component_expresses(html_template, input_schema), '', '') AS expresses' in q) = 0 THEN
    RAISE EXCEPTION '592 ROLLBACK: the derived column is not present verbatim — refusing';
  END IF;
  IF position('{{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}' in p) = 0 THEN
    RAISE EXCEPTION '592 ROLLBACK: the listing token is not present verbatim — refusing';
  END IF;
  IF position('Each entry ends with what it can EXPRESS' in p) = 0 THEN
    RAISE EXCEPTION '592 ROLLBACK: the header explanation is not present — refusing';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_available_components,config,query}',
         to_jsonb(replace(
           default_config#>>'{workflow,steps,load_available_components,config,query}',
           $new$, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$,
           $old$ FROM content_components$old$))),
       updated_at = now()
 WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';

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
             $newh$## Available Section Components
Each entry ends with what it can EXPRESS: `list` (a bulleted or numbered list), `table`, `items` (a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` — paragraphs and nothing else. A `[prose only]` section CANNOT render a list however it is written, and the writer will flatten the promise into paragraphs. If a page you plan promises a calendar, a step-by-step process, a checklist, a comparison or a specification, at least one of its sections must express `list`, `table`, `items` or `html-block`.$newh$,
             $oldh$## Available Section Components$oldh$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p FROM agent_definitions WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';
  IF position('component_expresses' in q) > 0 OR position('expresses' in p) > 0 THEN
    RAISE EXCEPTION '592 ROLLBACK VERIFY: 592 content still present';
  END IF;
  RAISE NOTICE '592 ROLLBACK OK';
END $$;

COMMIT;
