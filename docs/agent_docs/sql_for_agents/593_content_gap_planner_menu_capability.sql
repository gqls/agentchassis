-- 593 — content-gap-planner sees what each component can express (bugs_open/381, arm A)
--
-- WHY. The third of the three planner menus, and BY VOLUME THE ONE THAT MATTERS MOST:
-- `[MEASURED 2026-08-24, llm_call_log 30d]` content-gap-planner 749 calls,
-- build-site-planner 27, site-planner 5. The bug was found on a greenfield build, but
-- most component choices on this estate are made here. Rationale and the reason
-- capability is DERIVED rather than declared are in 591's header.
--
-- ⚠ REQUIRES 591 — it defines component_expresses(text, jsonb).
--
-- ⚠ THIS PROMPT HAS A PREFIX-CACHE BREAKPOINT AND THE LISTING SITS INSIDE IT.
-- `<!--CACHE_BREAKPOINT-->` is at character 897; the components listing at 853. So
-- both edits here fall in the CACHED PREFIX and will invalidate it once, on the next
-- call per site. That is a one-off cost and not a defect: the menu is already
-- site-varying (the query binds $1), so the prefix was never shared across sites in
-- the first place. The breakpoint marker itself is NOT moved, NOT duplicated and NOT
-- fragmented — the verify block asserts it still appears exactly once, because a
-- fragmented prefix is the failure mode migration 377 exists to prevent.
--
-- ⚠ THIS MENU SELECTS NO `description`, so the listing line is `- name (function)`
-- and stays that way — adding a description column is a bigger prompt change with its
-- own token cost, and is not this bug.
--
-- SCOPE. Config-only. LIVE ON APPLY, no chassis roll.
-- ROLLBACK: 593_content_gap_planner_menu_capability_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('content-gap-planner', '593_content_gap_planner_menu_capability: pre-update');

DO $$
DECLARE n int; q text; p text;
BEGIN
  IF to_regprocedure('component_expresses(text, jsonb)') IS NULL THEN
    RAISE EXCEPTION '593: component_expresses(text, jsonb) does not exist — apply 591 first';
  END IF;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '593: expected exactly 1 live content-gap-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';

  IF q IS NULL OR p IS NULL THEN
    RAISE EXCEPTION '593: menu query or plan_gaps.config.prompt_template is NULL — refusing';
  END IF;
  IF position('SELECT name, display_name, "function", category FROM content_components' in q) = 0 THEN
    RAISE EXCEPTION '593: the menu SELECT list is not verbatim — refusing to splice blind';
  END IF;
  IF position($chk$? 'backend'))$chk$ in q) = 0 THEN
    RAISE EXCEPTION '593: the requires-backend clause is missing from the menu query — refusing';
  END IF;
  IF position('component_expresses' in q) > 0 THEN
    RAISE EXCEPTION '593: already applied — refusing to double-apply';
  END IF;
  IF (length(p) - length(replace(p, '- {{.name}} ({{.function}})', '')))
     / length('- {{.name}} ({{.function}})') <> 1 THEN
    RAISE EXCEPTION '593: the component listing line does not appear exactly once — refusing';
  END IF;
  IF (length(p) - length(replace(p, '## Available Section Components', '')))
     / length('## Available Section Components') <> 1 THEN
    RAISE EXCEPTION '593: the components header does not appear exactly once — refusing';
  END IF;
  IF (length(p) - length(replace(p, '<!--CACHE_BREAKPOINT-->', '')))
     / length('<!--CACHE_BREAKPOINT-->') <> 1 THEN
    RAISE EXCEPTION '593: expected exactly one cache breakpoint marker — the prefix contract has changed, refusing';
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
               $old$SELECT name, display_name, "function", category FROM content_components$old$,
               $new$SELECT name, display_name, "function", category, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$
             ),
             $oldg$? 'backend'))$oldg$,
             $newg$? 'backend')) AND (NOT (COALESCE(semantic_tags, '[]'::jsonb) ? 'requires-evidence-base') OR EXISTS (SELECT 1 FROM site_specs ss_eb WHERE ss_eb.site_id = $1 AND ss_eb.aspect ILIKE '%evidence%' AND ss_eb.is_current))$newg$
           )
         )
       ),
       updated_at = now()
 WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc'
   AND type = 'content-gap-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_gaps,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}',
               $oldl$- {{.name}} ({{.function}})$oldl$,
               $newl$- {{.name}} ({{.function}}){{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}$newl$
             ),
             $oldh$## Available Section Components$oldh$,
             $newh$## Available Section Components
Each entry ends with what it can EXPRESS: `list` (a bulleted or numbered list), `table`, `items` (a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` — paragraphs and nothing else. A `[prose only]` section CANNOT render a list however it is written, and the writer will flatten the promise into paragraphs. If the gap you are filling is a calendar, a step-by-step process, a checklist, a comparison or a specification, choose a section that expresses `list`, `table`, `items` or `html-block`.$newh$
           )
         )
       ),
       updated_at = now()
 WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc'
   AND type = 'content-gap-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>>'{workflow,steps,plan_gaps,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = '637b750c-c460-489d-ba53-e5ebfb61adfc';
  IF position('array_to_string(component_expresses(html_template, input_schema)' in q) = 0 THEN
    RAISE EXCEPTION '593 VERIFY: the menu query does not call component_expresses';
  END IF;
  IF position('requires-evidence-base' in q) = 0 THEN
    RAISE EXCEPTION '593 VERIFY: the evidence-base gate was not inserted';
  END IF;
  IF position('[expresses: {{.expresses}}]' in p) = 0 THEN
    RAISE EXCEPTION '593 VERIFY: the listing line does not print the capability';
  END IF;
  IF position('Each entry ends with what it can EXPRESS' in p) = 0 THEN
    RAISE EXCEPTION '593 VERIFY: the header explanation was not inserted';
  END IF;
  IF (length(p) - length(replace(p, '<!--CACHE_BREAKPOINT-->', '')))
     / length('<!--CACHE_BREAKPOINT-->') <> 1 THEN
    RAISE EXCEPTION '593 VERIFY: the cache breakpoint no longer appears exactly once — the prefix has been fragmented';
  END IF;
  IF position('## Available Section Components' in p) > position('<!--CACHE_BREAKPOINT-->' in p) THEN
    RAISE EXCEPTION '593 VERIFY: the components block has moved past the cache breakpoint';
  END IF;
  RAISE NOTICE '593 OK: content-gap-planner menu carries capability (busiest planner: 749 calls/30d)';
END $$;

COMMIT;
