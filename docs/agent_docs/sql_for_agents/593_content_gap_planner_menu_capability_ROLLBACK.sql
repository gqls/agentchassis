-- ROLLBACK for 593 — surgical inverse; REFUSES rather than mis-splicing.
-- Does NOT drop component_expresses(): 591 and 592 also depend on it.
BEGIN;

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}'
    INTO q, p FROM agent_definitions WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';
  IF q IS NULL OR position('array_to_string(component_expresses(html_template, input_schema), '', '') AS expresses' in q) = 0 THEN
    RAISE EXCEPTION '593 ROLLBACK: the derived column is not present verbatim — refusing';
  END IF;
  IF position('requires-evidence-base' in q) = 0 THEN
    RAISE EXCEPTION '593 ROLLBACK: the evidence-base clause is not present — refusing rather than half-reverting';
  END IF;
  IF position('{{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}' in p) = 0 THEN
    RAISE EXCEPTION '593 ROLLBACK: the listing token is not present verbatim — refusing';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_available_components,config,query}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,load_available_components,config,query}',
               $newg$ AND (NOT (COALESCE(semantic_tags, '[]'::jsonb) ? 'requires-evidence-base') OR EXISTS (SELECT 1 FROM site_specs ss_eb WHERE ss_eb.site_id = $1 AND ss_eb.aspect ILIKE '%evidence%' AND ss_eb.is_current))$newg$,
               $oldg$$oldg$
             ),
             $new$, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$,
             $old$ FROM content_components$old$
           )
         )
       ),
       updated_at = now()
 WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_gaps,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}',
               $newl${{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}$newl$,
               $oldl$$oldl$
             ),
             $newh$## Available Section Components
Each entry ends with what it can EXPRESS: `list` (a bulleted or numbered list), `table`, `items` (a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` — paragraphs and nothing else. A `[prose only]` section CANNOT render a list however it is written, and the writer will flatten the promise into paragraphs. If the gap you are filling is a calendar, a step-by-step process, a checklist, a comparison or a specification, choose a section that expresses `list`, `table`, `items` or `html-block`.$newh$,
             $oldh$## Available Section Components$oldh$
           )
         )
       ),
       updated_at = now()
 WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}'
    INTO q, p FROM agent_definitions WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';
  IF position('component_expresses' in q) > 0 OR position('expresses' in p) > 0 THEN
    RAISE EXCEPTION '593 ROLLBACK VERIFY: 593 content still present';
  END IF;
  IF (length(p) - length(replace(p, '<!--CACHE_BREAKPOINT-->', ''))) / length('<!--CACHE_BREAKPOINT-->') <> 1 THEN
    RAISE EXCEPTION '593 ROLLBACK VERIFY: the cache breakpoint is not present exactly once';
  END IF;
  RAISE NOTICE '593 ROLLBACK OK';
END $$;

COMMIT;
