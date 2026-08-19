-- ROLLBACK for 493_loop_nested_item_key_suffixes.sql (bugs_open/321)
--
-- Removes exactly the five leaves 493 added: the four item_key_suffix_field
-- keys and tool-suggester's loop-level continue_on_error. Gated on each value
-- being EXACTLY what 493 wrote — if a later migration changed one, this aborts
-- rather than deleting someone else's work.
--
-- ⚠ Rolling back reinstates the silent collision this migration exists to fix:
-- N loop-filed items per site collapse back to 1. Do this only to unblock a
-- defect the suffix itself causes (e.g. the hard-error arm firing on malformed
-- suggestions faster than continue_on_error can absorb).
--
-- Snapshots are not restored here — 493's snapshot_agent rows remain available
-- for a full-row restore if that is what is actually wanted.

BEGIN;

SELECT snapshot_agent('tool-suggester',
  '493 ROLLBACK: pre-removal');
SELECT snapshot_agent('component-quality-auditor',
  '493 ROLLBACK: pre-removal');
SELECT snapshot_agent('internal-linker',
  '493 ROLLBACK: pre-removal');

DO $$
DECLARE
  ts_id  uuid;
  cqa_id uuid;
  il_id  uuid;
  n      int;
  v      text;
BEGIN
  SELECT id INTO STRICT ts_id  FROM agent_definitions
   WHERE type='tool-suggester' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT id INTO STRICT cqa_id FROM agent_definitions
   WHERE type='component-quality-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT id INTO STRICT il_id  FROM agent_definitions
   WHERE type='internal-linker' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  -- Gate: only remove what 493 wrote, exactly as it wrote it.
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_novel_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'current_suggestion.function' THEN
    RAISE EXCEPTION '493 rollback gate: novel suffix is % (not 493''s value) — aborting', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_library_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'current_suggestion.function' THEN
    RAISE EXCEPTION '493 rollback gate: library suffix is % — aborting', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,continue_on_error}'
    INTO v FROM agent_definitions WHERE id = ts_id;
  IF v IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '493 rollback gate: continue_on_error is % — aborting', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = cqa_id;
  IF v IS DISTINCT FROM 'current_component.component_id' THEN
    RAISE EXCEPTION '493 rollback gate: CQA suffix is % — aborting', COALESCE(v,'<absent>');
  END IF;
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_suffix_field}'
    INTO v FROM agent_definitions WHERE id = il_id;
  IF v IS DISTINCT FROM 'current_link.source_page' THEN
    RAISE EXCEPTION '493 rollback gate: linker suffix is % — aborting', COALESCE(v,'<absent>');
  END IF;

  UPDATE agent_definitions SET
    default_config = ((default_config
      #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_novel_item,config,item_key_suffix_field}')
      #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_library_item,config,item_key_suffix_field}')
      #- '{workflow,steps,create_items_loop,config,continue_on_error}',
    updated_at = now()
  WHERE id = ts_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493 rollback: tool-suggester UPDATE touched % rows', n; END IF;

  UPDATE agent_definitions SET
    default_config = default_config
      #- '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config,item_key_suffix_field}',
    updated_at = now()
  WHERE id = cqa_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493 rollback: CQA UPDATE touched % rows', n; END IF;

  UPDATE agent_definitions SET
    default_config = default_config
      #- '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config,item_key_suffix_field}',
    updated_at = now()
  WHERE id = il_id;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '493 rollback: internal-linker UPDATE touched % rows', n; END IF;

  -- Post-verify: all five leaves gone.
  IF (SELECT default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_novel_item,config}' ? 'item_key_suffix_field'
        FROM agent_definitions WHERE id = ts_id) THEN
    RAISE EXCEPTION '493 rollback post-verify: novel suffix still present';
  END IF;
  IF (SELECT default_config #> '{workflow,steps,create_items_loop,config}' ? 'continue_on_error'
        FROM agent_definitions WHERE id = ts_id) THEN
    RAISE EXCEPTION '493 rollback post-verify: continue_on_error still present';
  END IF;
  IF (SELECT default_config #> '{workflow,steps,create_regen_items,config,sub_workflow,steps,create_work_item,config}' ? 'item_key_suffix_field'
        FROM agent_definitions WHERE id = cqa_id) THEN
    RAISE EXCEPTION '493 rollback post-verify: CQA suffix still present';
  END IF;
  IF (SELECT default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_rewrite_item,config}' ? 'item_key_suffix_field'
        FROM agent_definitions WHERE id = il_id) THEN
    RAISE EXCEPTION '493 rollback post-verify: linker suffix still present';
  END IF;

  RAISE NOTICE 'OK 493 ROLLBACK: five leaves removed';
END $$;

COMMIT;
