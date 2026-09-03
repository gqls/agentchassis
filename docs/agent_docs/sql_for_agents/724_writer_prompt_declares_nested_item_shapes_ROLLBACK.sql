-- 724_writer_prompt_declares_nested_item_shapes_ROLLBACK.sql
--
-- Restores the page-content-writer prompt to its pre-724 spelling: the exemplar renders a
-- field's elements from the flat item-name list only, and the field list carries no shape
-- sentences. Same anchoring discipline as the forward migration — exact-once guards, expected
-- balance deltas, DO/RAISE verify — so a partially-applied or concurrently-edited row refuses
-- rather than being silently rewritten.
--
-- SAFE IN BOTH DEPLOY STATES. A chassis carrying the Go half keeps emitting value_shape and
-- item_notes; the reverted template simply ignores both keys (they are never referenced), so
-- the prompt returns to exactly what it produced before 724 rather than to anything new. There
-- is therefore no need to roll the image back with it.
--
-- ⚠ WHAT REVERTING COSTS, so nobody applies this casually: the flat exemplar declares
-- mechanism-flow's `steps[].branches` a STRING, which is bugs_open/437 — 119 failed builds
-- across six sites in the fortnight to 2026-09-02, and pages that never publish while live
-- pages link them. The livespec declaration workflow.page-content-writer.prompt_item_shape
-- marks that spelling Forbidden, so the daily live-declaration-drift check will fire on this
-- row from the next run after it is applied. That is the mechanism working, not a false alarm.
--
-- Apply: psql -f THIS FILE ONLY.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '724 ROLLBACK REFUSED: expected exactly 1 active page-content-writer row, found %', n;
  END IF;
  PERFORM snapshot_agent('page-content-writer',
                         '724_..._ROLLBACK.sql: pre-revert');
END $$;

DO $do724r$
DECLARE
  tpl text; newtpl text; n int;
  ifs_before int; ends_before int; ranges_before int;
  anchor_A text := $a724r${{if .item_notes}}{{range $n := .item_notes}} {{$n}}{{end}}{{end}}$a724r$;
  anchor_B text := $b724r$"{{$f.name}}": {{if $f.value_shape}}{{$f.value_shape}}{{else if $f.item_fields}}$b724r$;
  repl_B   text := $rb724r$"{{$f.name}}": {{if $f.item_fields}}$rb724r$;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl
    FROM agent_definitions WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN
    RAISE EXCEPTION '724 ROLLBACK: generate_content.config.prompt_template not found';
  END IF;

  n := (length(tpl) - length(replace(tpl, anchor_A, ''))) / length(anchor_A);
  IF n <> 1 THEN RAISE EXCEPTION '724 ROLLBACK: item_notes tail found % times, expected 1 — is 724 applied?', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor_B, ''))) / length(anchor_B);
  IF n <> 1 THEN RAISE EXCEPTION '724 ROLLBACK: nested exemplar found % times, expected 1 — is 724 applied?', n; END IF;

  ifs_before    := (length(tpl) - length(replace(tpl, '{{if ',   ''))) / length('{{if ');
  ends_before   := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  ranges_before := (length(tpl) - length(replace(tpl, '{{range ',''))) / length('{{range ');

  newtpl := tpl;
  newtpl := replace(newtpl, anchor_A, '');
  newtpl := replace(newtpl, anchor_B, repl_B);

  IF length(newtpl) <> length(tpl) - length(anchor_A) + (length(repl_B) - length(anchor_B)) THEN
    RAISE EXCEPTION '724 ROLLBACK: unexpected length delta';
  END IF;
  -- The exact inverse of the forward migration's +1: dropping the item_notes tail removes one
  -- {{if }}, while '{{else if $f.item_fields}}' reverting to '{{if $f.item_fields}}' ADDS one
  -- back. Net -1.
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs_before - 1 THEN
    RAISE EXCEPTION '724 ROLLBACK: {{if}} count is not -1';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends_before - 2 THEN
    RAISE EXCEPTION '724 ROLLBACK: {{end}} count is not -2';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{range ', ''))) / length('{{range ') <> ranges_before - 1 THEN
    RAISE EXCEPTION '724 ROLLBACK: {{range}} count is not -1';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{else if ', ''))) / length('{{else if ') <> 0 THEN
    RAISE EXCEPTION '724 ROLLBACK: an {{else if}} survives the revert';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
           to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '724 ROLLBACK: updated % rows, expected exactly 1', n; END IF;
END $do724r$;

DO $$
DECLARE tpl text; n int;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl
    FROM agent_definitions WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('value_shape' in tpl) > 0 THEN
    RAISE EXCEPTION '724 ROLLBACK VERIFY: value_shape still referenced';
  END IF;
  IF position('item_notes' in tpl) > 0 THEN
    RAISE EXCEPTION '724 ROLLBACK VERIFY: item_notes still referenced';
  END IF;
  n := (length(tpl) - length(replace(tpl, '"{{$f.name}}": {{if $f.item_fields}}', '')))
       / length('"{{$f.name}}": {{if $f.item_fields}}');
  IF n <> 1 THEN RAISE EXCEPTION '724 ROLLBACK VERIFY: flat exemplar not restored exactly once (found %)', n; END IF;

  RAISE NOTICE '724 ROLLBACK OK: prompt restored to the pre-724 flat exemplar. bugs_open/437 is live again on this row.';
END $$;

COMMIT;
