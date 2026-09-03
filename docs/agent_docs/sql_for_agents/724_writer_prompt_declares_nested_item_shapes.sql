-- 724_writer_prompt_declares_nested_item_shapes.sql
--
-- bugs_open/437 — the DB half. The Go half (datahelpers.StructuredItemShape, carried on
-- llmFieldSpec as value_shape / item_notes) computes the nested element shape; this migration
-- teaches the page-content-writer prompt to SHOW it.
--
-- THE DEFECT, at the artefact. mechanism-flow declares `steps[].branches` as an array of
-- objects {body,label}. plan_sections projected that to a flat list of element NAMES, and this
-- prompt's Output Format exemplar rendered from those names:
--
--     "steps": [{ "body": "...", "branches": "...", "marker": "...", "note": "...", "title": "..." }]
--
-- i.e. the prompt itself declared `branches` a STRING. The writer copied the exemplar it was
-- shown (llm_call_log 34f25815-42d3-4057-b42a-b8b42189ae7e, 2026-09-02 19:07Z: prompt line 234
-- as above, reply "branches": "Broadcast ads follow the BCAP Code…"), the render type gate
-- refused it correctly, and the page never built. 119 such failures in the fortnight to
-- 2026-09-02 across six sites, deterministic — the model was obedient throughout.
-- Exemplars ship verbatim and demonstrations govern (LANDMINES:
-- a-quoted-exemplar-in-a-prompt-is-copied-verbatim), which is why this is a prompt fix and not
-- a writer fix.
--
-- TWO SITES, both anchored exactly-once against the live text (measured 2026-09-03):
--   A. "What To Write" field list — gains the per-property shape sentences, which also carry
--      each nested property's own schema description (the flat projection dropped those, so no
--      writer has ever seen "a decision point: two or more outcomes, rendered side by side").
--   B. Output Format exemplar — prefers the computed skeleton, falls back to today's flat
--      rendering, then to a scalar. `{{else if}}` keeps the existing {{end}} closing the chain.
--
-- DEPLOY ORDER IS FREE, and that is by construction rather than by luck: value_shape and
-- item_notes are `omitempty`, so a chassis WITHOUT the Go half emits neither key, and both new
-- directives are inside {{if}} guards. ⚠ This prompt renders under text/template's DEFAULT
-- missingkey (invalid) — datahelpers.RenderPromptTemplate sets no Option — so an absent key is
-- falsy in {{if}} but would print as a literal <no value> if ever used bare. Both keys appear
-- ONLY inside their guards, and platform/orchestration/actions/writer_prompt_item_shape_437_test.go
-- proves both deploy states render byte-identically to today for every component that has no
-- nested shape (1 component has one, measured 2026-09-03: mechanism-flow).
--
-- The contract is declared in platform/livespec as workflow.page-content-writer.prompt_item_shape
-- (fragments WriterPromptNestedExemplar / WriterPromptItemNotesTail, plus the pre-437 spelling
-- as Forbidden). The literals below are that declaration's, verbatim — keep the two in step.
-- Seed 023 still carries the pre-437 spelling; the seed is history, the live row is the system.
--
-- Apply: psql -f THIS FILE ONLY (never an unscoped runner --apply). Companion ROLLBACK alongside.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '724 REFUSED: expected exactly 1 active page-content-writer row, found %', n;
  END IF;
  PERFORM snapshot_agent('page-content-writer',
                         '724_writer_prompt_declares_nested_item_shapes.sql: pre-update');
END $$;

DO $do724$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; elses_before int; vars_before int; ranges_before int;
  anchor_A text := $a724$Each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}$a724$;
  repl_A   text := $ra724$Each item is an object with exactly these fields: {{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}{{if .item_notes}}{{range $n := .item_notes}} {{$n}}{{end}}{{end}}$ra724$;
  anchor_B text := $b724$"{{$f.name}}": {{if $f.item_fields}}$b724$;
  repl_B   text := $rb724$"{{$f.name}}": {{if $f.value_shape}}{{$f.value_shape}}{{else if $f.item_fields}}$rb724$;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl
    FROM agent_definitions WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN
    RAISE EXCEPTION '724: generate_content.config.prompt_template not found';
  END IF;

  n := (length(tpl) - length(replace(tpl, anchor_A, ''))) / length(anchor_A);
  IF n <> 1 THEN RAISE EXCEPTION '724: anchor A (field list) found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_B, ''))) / length(anchor_B);
  IF n <> 1 THEN RAISE EXCEPTION '724: anchor B (exemplar) found % times, expected 1', n; END IF;

  ifs_before    := (length(tpl) - length(replace(tpl, '{{if ',   ''))) / length('{{if ');
  ends_before   := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses_before  := (length(tpl) - length(replace(tpl, '{{else}}',''))) / length('{{else}}');
  vars_before   := (length(tpl) - length(replace(tpl, '{{.',     ''))) / length('{{.');
  ranges_before := (length(tpl) - length(replace(tpl, '{{range ',''))) / length('{{range ');

  newtpl := tpl;
  newtpl := replace(newtpl, anchor_A, repl_A);
  newtpl := replace(newtpl, anchor_B, repl_B);

  IF length(newtpl) <> length(tpl)
       + (length(repl_A) - length(anchor_A))
       + (length(repl_B) - length(anchor_B)) THEN
    RAISE EXCEPTION '724: unexpected length delta';
  END IF;

  -- Balance, with EXPECTED deltas rather than "unchanged": this edit adds directives, so an
  -- unchanged count would mean the splice did not happen.
  --   A: +1 {{if}} (item_notes), +1 {{range}}, +2 {{end}}.
  --   B: +1 {{if}} for value_shape, but the existing '{{if $f.item_fields}}' BECOMES
  --      '{{else if $f.item_fields}}' — so B's net {{if }} delta is ZERO, and it adds no
  --      {{end}} because it reuses the chain's existing one.
  -- Hence +1 overall, not +2. The rehearsal against the live text is what corrected this
  -- arithmetic; a guard that expected +2 would have refused a correct splice.
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before + 1 THEN
    RAISE EXCEPTION '724: {{if}} count is not +1';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before + 2 THEN
    RAISE EXCEPTION '724: {{end}} count is not +2';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{range ', ''))) / length('{{range ') <> ranges_before + 1 THEN
    RAISE EXCEPTION '724: {{range}} count is not +1';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses_before THEN
    RAISE EXCEPTION '724: {{else}} count changed — the flat and scalar arms must be untouched';
  END IF;
  -- A new bare {{.field}} would render <no value> under this path's default missingkey. The
  -- additions deliberately use {{if .x}} / {{range $n := .x}} / {{$n}}, none of which spell '{{.'.
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) / length('{{.') <> vars_before THEN
    RAISE EXCEPTION '724: replacement introduced a bare template variable';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{else if ', ''))) / length('{{else if ') <> 1 THEN
    RAISE EXCEPTION '724: expected exactly one {{else if}} after the splice';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
           to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '724: updated % rows, expected exactly 1', n; END IF;
END $do724$;

-- Verify (DO/RAISE — a block of SELECTs cannot stop the COMMIT): both new sites present exactly
-- once, the pre-437 spelling gone, and the flat/scalar arms still intact for every component
-- that has no nested shape.
DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl
    FROM agent_definitions WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  n := (length(tpl) - length(replace(tpl, '{{if $f.value_shape}}{{$f.value_shape}}{{else if $f.item_fields}}', '')))
       / length('{{if $f.value_shape}}{{$f.value_shape}}{{else if $f.item_fields}}');
  IF n <> 1 THEN RAISE EXCEPTION '724 VERIFY: nested exemplar not present exactly once (found %)', n; END IF;

  n := (length(tpl) - length(replace(tpl, '{{if .item_notes}}{{range $n := .item_notes}} {{$n}}{{end}}{{end}}', '')))
       / length('{{if .item_notes}}{{range $n := .item_notes}} {{$n}}{{end}}{{end}}');
  IF n <> 1 THEN RAISE EXCEPTION '724 VERIFY: item_notes tail not present exactly once (found %)', n; END IF;

  IF position('"{{$f.name}}": {{if $f.item_fields}}' in tpl) > 0 THEN
    RAISE EXCEPTION '724 VERIFY: the pre-437 exemplar spelling survives — it declares a nested array a string';
  END IF;

  -- The fallback arms must survive: they are what keeps every flat component's prompt
  -- byte-identical, which is this change's whole blast-radius claim.
  n := (length(tpl) - length(replace(tpl, '[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]', '')))
       / length('[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}"{{$k}}": "..."{{end}} }]');
  IF n <> 1 THEN RAISE EXCEPTION '724 VERIFY: flat item-fields exemplar arm lost (found %)', n; END IF;

  IF position('Each item is an object with exactly these fields:' in tpl) = 0 THEN
    RAISE EXCEPTION '724 VERIFY: field-list item-fields sentence lost';
  END IF;

  RAISE NOTICE '724 OK: the exemplar demonstrates nested element shapes; flat components unchanged.';
END $$;

COMMIT;
