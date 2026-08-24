-- 591 — the planner is told what each component can EXPRESS (bugs_open/381, arm A)
--
-- WHY. `build-site-planner` composes each page by picking component names from a
-- menu that prints `name (display_name): function - description` and NOTHING about
-- what markup the component can produce. So the planner chooses blind on
-- expressiveness: `garden-tools.uk`'s seasonal-planner page promised "month by
-- month" under its own <h2> and was built from hero + generic-text-block +
-- info-card-grid + call-to-action, none of which can render a list. The writer had
-- nowhere to put twelve months and wrote four seasons of prose. Nothing downstream
-- reports it: the sections rendered, the page deployed, every check passed.
--
-- MEASURED 2026-08-24 (30d, fleet): 741 pages across 29 sites; 327 (44%) contain no
-- list, no table and no <strong> anywhere in their content; 1,863 of 1,980 section
-- placements (94%) used a template that cannot produce a list or a table.
--
-- WHY THE CAPABILITY IS DERIVED AND NOT DECLARED. `content_components.content_shape`
-- already exists for exactly this purpose (sql_for_tables/005_content_components.sql
-- :9118, COMMENT: "prose, structured_list, structured_card, key_value_pairs"). It is
-- DEAD: zero Go readers, omitted from the birth INSERT
-- (store_generated_component_action.go:634 sets 19 columns and not this one), NULL on
-- 128 of 151 active section rows, drifted to free text ('series', 'sequence',
-- 'mixed'), and 12 rows marked 'structured_list' have NO list markup in their
-- template. A hand-maintained capability column is the failure mode being fixed, so
-- this file adds no column and backfills nothing.
--
-- READ TIME, NOT STORED. `component_expresses()` is IMMUTABLE and called from the
-- menu query. No ALTER TABLE, no trigger, no generated column — deliberate: the
-- RFC_032/bugs_open/283 lane is rewriting html_template fleet-wide (140 of 297 active
-- templates now carry {{.InstanceID}}), and a stored capability would be stale the
-- moment a template changed. 151 rows; the cost is nil.
--
-- ⚠ A TEMPLATE TOKEN IS A BINDING, NOT CONTENT (raised by the 283 lane). The
-- function matches LITERAL markup (`<ul`, `<ol`, `<table`) and one template
-- construct (`{{range` over an llm array field, which is what a repeating card/FAQ
-- grid is). It never reads a `{{.Field}}` binding as evidence of anything.
--
-- THE SECOND ARM, AND WHY THE ORDER DOES NOT MATTER. Migration 594 retypes four
-- pass-through prose fields to `html`, which makes `component_expresses` report
-- {html-block,list,table} for them — including `generic-text-block`, the fleet's
-- default fallback (181 instances). Applied before this file, the menu shows the new
-- capability immediately; applied after, it starts prose-only and gains it. Both
-- converge and neither is wrong in the window.
--
-- THE EVIDENCE-BASE GATE, and why it is here rather than in the vocabulary. Surfacing
-- capability makes fact-fed components (evidence-chart, evidence-timeseries, the
-- timeline the editorial lane is building) reachable as generic list-expressers — and
-- on a site with no evidence base they can only fail. The vocabulary stays pure; the
-- gate lives in the menu row set, exactly as 419 gates `requires-backend`. MEASURED
-- by the editorial_design_uplift lane 2026-08-24: `data_sources` is EMPTY on both
-- evidence components and exactly ONE active component fleet-wide uses that column at
-- all, so a `requires-evidence-base` semantic tag is the ONLY available mechanism,
-- not belt-and-braces. NO ROW CARRIES THE TAG TODAY — the two-row tagging UPDATE is
-- owed by that lane — so this clause is INERT on apply and changes no menu by one
-- row. That is intended: the tag semantics should exist before the tags do.
--
-- WHAT IT DOES NOT DO. It does not add any component to the library. The library has
-- no generic checklist, steps, comparison-table or calendar component — enumerated
-- 2026-08-24, the 44 structural components are directories, trackers, calculators,
-- quizzes, spec sheets, one pricing table, two carousels and site-footer (chrome).
-- So this file converts a blind choice into an informed one; it cannot make a
-- seasonal planner possible. That gap is recorded in bugs_open/381 §3, not fixed here.
--
-- SCOPE. Config-only + one function. LIVE ON APPLY, no chassis roll. Scoped by id,
-- pre-state gated, DO/RAISE verify, snapshot first, rollback sidecar — 485's shape.
-- Sibling menus: 592 (site-planner), 593 (content-gap-planner). Council: migrations
-- have been in the gate's scope since 2026-08-19 (bugs_open/314).
--
-- ROLLBACK: 591_component_expresses_and_build_site_planner_menu_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '591_component_expresses_and_build_site_planner_menu: pre-update');

-- ── The derived vocabulary ──────────────────────────────────────────────────
-- html-block : an llm-sourced field declared `html` — the writer may emit subheads,
--              lists, tables and emphasis into it, so the component can express all
--              three (this is what 594 turns on for the four prose slots).
-- list       : the template contains literal <ul/<ol markup.
-- table      : the template contains literal <table markup.
-- items      : the template ranges over an llm array field — a repeating card, FAQ
--              or feature set. Structure, but fixed-shape structure, which is why it
--              is a separate token from `list` rather than folded into it.
-- {}         : prose only.
CREATE OR REPLACE FUNCTION component_expresses(p_html_template text, p_input_schema jsonb)
RETURNS text[]
LANGUAGE sql
IMMUTABLE
AS $fn$
  SELECT COALESCE(array_agg(x ORDER BY x), ARRAY[]::text[]) FROM (
    SELECT 'html-block'::text AS x WHERE EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'list' WHERE p_html_template ~* '<(ul|ol)[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'table' WHERE p_html_template ~* '<table[\s>]' OR EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' = 'html')
    UNION
    SELECT 'items' WHERE p_html_template ~* '\{\{[-\s]*range' AND EXISTS (
      SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
       WHERE f.value->>'source' = 'llm' AND f.value->>'type' IN ('array', 'list'))
  ) s;
$fn$;

COMMENT ON FUNCTION component_expresses(text, jsonb) IS
  'bugs_open/381: what markup a component can produce, DERIVED from its template and '
  'schema — never declared. Read by the planner menu queries (591/592/593). Supersedes '
  'the dead content_shape column in intent; that column has no readers and is wrong on '
  '12 rows. Regex note: a PostgreSQL \b is BACKSPACE, not a word boundary — the '
  'character classes here are deliberate.';

-- ── Pre-state ───────────────────────────────────────────────────────────────
DO $$
DECLARE n int; q text; p text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '591: expected exactly 1 live build-site-planner row, found % — a second active row would make the id-scoped UPDATE a silent no-op', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF q IS NULL OR p IS NULL THEN
    RAISE EXCEPTION '591: load_components.config.query or plan_site.config.prompt_template is NULL — the workflow has changed under me, refusing';
  END IF;

  -- The menu SELECT list, verbatim as 407 left it and 419 last gated it.
  IF position('SELECT name, display_name, "function", category, description FROM content_components' in q) = 0 THEN
    RAISE EXCEPTION '591: the menu SELECT list is not verbatim — another migration has edited it, refusing to splice blind';
  END IF;
  -- The requires-backend clause 419 installed, which the evidence gate sits beside.
  IF position($chk$? 'backend'))$chk$ in q) = 0 THEN
    RAISE EXCEPTION '591: the requires-backend clause (419) is missing from the menu query — refusing, the WHERE clause is not the shape this file was written against';
  END IF;
  IF position('component_expresses' in q) > 0 THEN
    RAISE EXCEPTION '591: already applied (the menu query already calls component_expresses) — refusing to double-apply';
  END IF;

  -- The listing line, and the rule anchor.
  IF (length(p) - length(replace(p, '- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}', '')))
     / length('- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}') <> 1 THEN
    RAISE EXCEPTION '591: the component listing line does not appear exactly once — refusing';
  END IF;
  IF (length(p) - length(replace(p, 'Return ONLY valid JSON.', '')))
     / length('Return ONLY valid JSON.') <> 1 THEN
    RAISE EXCEPTION '591: the "Return ONLY valid JSON." anchor does not appear exactly once — refusing to append rule 19 blind';
  END IF;
  IF position('expresses' in p) > 0 THEN
    RAISE EXCEPTION '591: already applied (the prompt already mentions expresses) — refusing to double-apply';
  END IF;
END $$;

-- ── The menu query: one derived column, and the evidence-base row gate ───────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,load_components,config,query}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,load_components,config,query}',
               $old$SELECT name, display_name, "function", category, description FROM content_components$old$,
               $new$SELECT name, display_name, "function", category, description, array_to_string(component_expresses(html_template, input_schema), ', ') AS expresses FROM content_components$new$
             ),
             $oldg$? 'backend'))$oldg$,
             $newg$? 'backend')) AND (NOT (COALESCE(semantic_tags, '[]'::jsonb) ? 'requires-evidence-base') OR EXISTS (SELECT 1 FROM site_specs ss_eb WHERE ss_eb.site_id = $1 AND ss_eb.aspect ILIKE '%evidence%' AND ss_eb.is_current))$newg$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
   AND type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── The prompt: print the capability, and one rule that uses it ─────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,plan_site,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,plan_site,config,prompt_template}',
               $oldl$- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}$oldl$,
               $newl$- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}{{if .expresses}} [expresses: {{.expresses}}]{{else}} [prose only]{{end}}$newl$
             ),
             $olda$Return ONLY valid JSON.$olda$,
             $newa$19. MATCH STRUCTURE TO PROMISE. Each component above carries what it can EXPRESS: `list` (renders a bulleted or numbered list), `table` (renders a table), `items` (renders a fixed set of repeating cards or entries), `html-block` (the writer may put subheadings, lists, emphasis and tables inside it), or `[prose only]` (paragraphs, and nothing else). A `[prose only]` section CANNOT render a list no matter what is written for it — the markup is not in its template — and the writer will silently flatten the promise into paragraphs. So: if a page or a section you are planning promises a month-by-month calendar, a step-by-step process, a checklist, a comparison, or a specification, at least one section on that page MUST express `list`, `table`, `items` or `html-block`. This is the difference between a page that keeps its own heading's promise and one that reads as padding. Do not pad the other way either: a page with nothing enumerable on it is right to be all prose.

Return ONLY valid JSON.$newa$
           )
         )
       ),
       updated_at = now()
 WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b'
   AND type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Verify. DO/RAISE, not bare SELECTs: a verify block of SELECTs cannot stop
--    the COMMIT (ON_ERROR_STOP ignores a non-empty result) ────────────────────
DO $$
DECLARE q text; p text; probe text[];
BEGIN
  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE id = 'f263eaa1-61e1-446e-9410-648e12b7875b';

  IF position('array_to_string(component_expresses(html_template, input_schema)' in q) = 0 THEN
    RAISE EXCEPTION '591 VERIFY: the menu query does not call component_expresses — the replace did not fire';
  END IF;
  IF position('requires-evidence-base' in q) = 0 THEN
    RAISE EXCEPTION '591 VERIFY: the evidence-base gate was not inserted';
  END IF;
  IF position($chk$? 'backend'))$chk$ in q) = 0 THEN
    RAISE EXCEPTION '591 VERIFY: the requires-backend clause was consumed rather than preserved';
  END IF;
  IF position('[expresses: {{.expresses}}]' in p) = 0 THEN
    RAISE EXCEPTION '591 VERIFY: the listing line does not print the capability';
  END IF;
  IF position('19. MATCH STRUCTURE TO PROMISE.' in p) = 0 THEN
    RAISE EXCEPTION '591 VERIFY: rule 19 was not inserted';
  END IF;
  IF (length(p) - length(replace(p, 'Return ONLY valid JSON.', '')))
     / length('Return ONLY valid JSON.') <> 1 THEN
    RAISE EXCEPTION '591 VERIFY: the closing anchor appears more than once — the prompt has been duplicated';
  END IF;

  -- The function itself, on a component whose answer is known by hand.
  SELECT component_expresses(html_template, input_schema) INTO probe
    FROM content_components WHERE function = 'ported-prose' AND is_active LIMIT 1;
  IF probe IS NULL OR NOT ('list' = ANY(probe) AND 'table' = ANY(probe) AND 'html-block' = ANY(probe)) THEN
    RAISE EXCEPTION '591 VERIFY: component_expresses(ported-prose) = % — expected html-block/list/table', probe;
  END IF;
  -- NEGATIVE control in the same breath: a prose-only component must come back empty.
  SELECT component_expresses(html_template, input_schema) INTO probe
    FROM content_components WHERE function = 'call-to-action' AND is_active LIMIT 1;
  IF probe IS NULL OR array_length(probe, 1) IS NOT NULL THEN
    RAISE EXCEPTION '591 VERIFY: component_expresses(call-to-action) = % — expected {} (a control that matches everything is not a control)', probe;
  END IF;

  RAISE NOTICE '591 OK: build-site-planner sees what each component can express, and rule 19 tells it to use that';
END $$;

COMMIT;
