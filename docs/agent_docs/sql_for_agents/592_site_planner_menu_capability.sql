-- 592 — site-planner sees what each component can express (bugs_open/381, arm A)
--
-- WHY. The sibling of 591 for the second of the three planners. Same defect, same
-- remedy: the menu prints identity and no capability, so the planner composes blind.
-- Rationale, measurements and the reason capability is DERIVED rather than declared
-- are in 591's header and are not repeated here.
--
-- ⚠ REQUIRES 591 — it defines component_expresses(text, jsonb). This file guards on
-- the function existing and refuses without it.
--
-- ⚠ NO EVIDENCE-BASE GATE HERE, AND IT IS NOT AN OVERSIGHT. 591 and 593 gate
-- `requires-evidence-base` rows out of the menu for sites with no evidence base.
-- site-planner's menu query takes NO site parameter (`params` is absent from its
-- load_available_components step), so the gate cannot be expressed without inventing
-- one and threading a site id that this workflow does not carry here. Rather than
-- paper over that, it is recorded: a fact-fed component IS reachable from this
-- planner's menu. It is a small hole today — site-planner ran 5 times in the 30 days
-- to 2026-08-24 against content-gap-planner's 749 — and closing it means giving this
-- step a site parameter, which is a separate change with its own blast radius.
--
-- SCOPE. Config-only. LIVE ON APPLY, no chassis roll.
-- ROLLBACK: 592_site_planner_menu_capability_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('site-planner', '592_site_planner_menu_capability: pre-update');

DO $$
DECLARE n int; q text; p text;
BEGIN
  IF to_regprocedure('component_expresses(text, jsonb)') IS NULL THEN
    RAISE EXCEPTION '592: component_expresses(text, jsonb) does not exist — apply 591 first';
  END IF;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '592: expected exactly 1 live site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';

  IF q IS NULL OR p IS NULL THEN
    RAISE EXCEPTION '592: load_available_components.config.query or plan_site.config.prompt_template is NULL — refusing';
  END IF;
  IF position('SELECT name, display_name, "function", category, description FROM content_components' in q) = 0 THEN
    RAISE EXCEPTION '592: the menu SELECT list is not verbatim — refusing to splice blind';
  END IF;
  IF position('component_expresses' in q) > 0 THEN
    RAISE EXCEPTION '592: already applied — refusing to double-apply';
  END IF;
  IF (length(p) - length(replace(p, '- [{{.function}}] {{.display_name}}: {{.description}}', '')))
     / length('- [{{.function}}] {{.display_name}}: {{.description}}') <> 1 THEN
    RAISE EXCEPTION '592: the component listing line does not appear exactly once — refusing';
  END IF;
  IF (length(p) - length(replace(p, '## Available Section Components', '')))
     / length('## Available Section Components') <> 1 THEN
    RAISE EXCEPTION '592: the components header does not appear exactly once — refusing';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_available_components,config,query}',
         to_jsonb(
           replace(
             default_config#>>'{workflow,steps,load_available_components,config,query}',
             $old$SELECT name, display_name, "function", category, description FROM content_components$old$,
             $new$SELECT name, display_name, "function", category, description, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57'
   AND type = 'site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_site,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
               $oldl$- [{{.function}}] {{.display_name}}: {{.description}}$oldl$,
               $newl$- [{{.function}}] {{.display_name}}: {{.description}}{{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}$newl$
             ),
             $oldh$## Available Section Components$oldh$,
             $newh$## Available Section Components
Each entry ends with what it can EXPRESS: `list` (a bulleted or numbered list), `table`, `items` (a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` — paragraphs and nothing else. A `[prose only]` section CANNOT render a list however it is written, and the writer will flatten the promise into paragraphs. If a page you plan promises a calendar, a step-by-step process, a checklist, a comparison or a specification, at least one of its sections must express `list`, `table`, `items` or `html-block`.$newh$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57'
   AND type = 'site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f7c8bee1-a845-4d5c-b136-761a844aba57';
  IF position('array_to_string(component_expresses(html_template, input_schema)' in q) = 0 THEN
    RAISE EXCEPTION '592 VERIFY: the menu query does not call component_expresses';
  END IF;
  IF position('[expresses: {{.expresses}}]' in p) = 0 THEN
    RAISE EXCEPTION '592 VERIFY: the listing line does not print the capability';
  END IF;
  IF position('Each entry ends with what it can EXPRESS' in p) = 0 THEN
    RAISE EXCEPTION '592 VERIFY: the header explanation was not inserted';
  END IF;
  IF (length(p) - length(replace(p, '## Available Section Components', '')))
     / length('## Available Section Components') <> 1 THEN
    RAISE EXCEPTION '592 VERIFY: the header appears more than once — the block has been duplicated';
  END IF;
  RAISE NOTICE '592 OK: site-planner menu carries capability';
END $$;

COMMIT;
