-- 579_copy_editor_work_item_entry_path.sql
--
-- WHAT: give `copy-editor` (stage 2, CQ-024) a second entry path so it can be
-- DISPATCHED from a work item, not only hand-fired with an explicit page_id.
--
-- WHY NOW: this is the prerequisite half of "wiring". Nothing routes work to
-- copy-editor today — every run to date has been a hand-fired canary — and the
-- router half (audit `tone` findings -> handler_agent='copy-editor') is a Go
-- change on a shared action that goes to council separately. This migration is
-- deliberately INERT until that lands: it adds a branch nothing currently takes.
--
-- THE PROBLEM IT SOLVES: a dispatched handler receives only `work_item_id`
-- (claim_work_item_action.go returns nothing else), while copy-editor's
-- `load_page_target` binds `input_data.page_id`. A single dual-source step is
-- NOT possible: QueryDatabaseAction ERRORS on a param path that resolves to nil
-- ("query param path '%s' resolved to nil", database_actions.go), so a COALESCE
-- over both sources dies on whichever source is absent.
--
-- THE SHAPE: branch on entry, converge on one field.
--   route_entry (conditional_branch, truthy on input_data.work_item_id)
--     then -> load_work_item  (dispatched: read page_id off the item)
--     else -> echo_page_ref   (hand-fired: echo input_data.page_id)
--   both write `page_ref` {page_id}, and load_page_target now binds
--   `page_ref.page_id` instead of `input_data.page_id`.
-- A truthy condition is safe on an ABSENT field: for ==/!=/truthy a nil operand
-- is a legitimate null probe and routes to else_step rather than failing
-- (conditional_branch_action.go's header, and bugs_open/313 is why the numeric
-- operators are the ones that fail loudly instead).
--
-- SAFETY POSTURE IS UNCHANGED, and that is the point: no step here can write to
-- a page, migration 447's guard still RAISEs if one is added, and output still
-- parks at `copy_edit_proposed`/`needs_human_review`. Dispatching this agent
-- produces auto-PROPOSALS, never auto-edits, so owner decision D2 ("no
-- unreviewed auto-rewrite") is untouched.
--
-- ROLLBACK: 579_copy_editor_work_item_entry_path_ROLLBACK.sql

BEGIN;

-- Guard 1: the agent must exist, exactly once, in the shape we expect.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live copy-editor, found %', n;
  END IF;
END $$;

-- Guard 2: refuse if the entry path has already been changed (re-run safety),
-- and refuse if load_page_target is not the shape this migration rewrites.
DO $$
DECLARE start_step text; lpt_param text;
BEGIN
  SELECT default_config->'workflow'->>'start_step',
         default_config->'workflow'->'steps'->'load_page_target'->'config'->'params'->>0
    INTO start_step, lpt_param
    FROM agent_definitions
   WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF start_step = 'route_entry' THEN
    RAISE EXCEPTION '579 already applied (start_step is route_entry) — nothing to do';
  END IF;
  IF start_step <> 'ensure_site_record' THEN
    RAISE EXCEPTION 'unexpected start_step %, expected ensure_site_record', start_step;
  END IF;
  IF lpt_param <> 'input_data.page_id' THEN
    RAISE EXCEPTION 'load_page_target param is %, expected input_data.page_id — the shape moved, re-read before applying', lpt_param;
  END IF;
END $$;

-- Snapshot, so the rollback restores exactly what was here.
CREATE TABLE IF NOT EXISTS agent_definitions_backup (
  id uuid, type text, default_config jsonb, snapshot_reason text, taken_at timestamptz DEFAULT now());
INSERT INTO agent_definitions_backup (id, type, default_config, snapshot_reason)
SELECT id, type, default_config, '579_copy_editor_work_item_entry_path'
  FROM agent_definitions
 WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET
  default_config = jsonb_set(
    jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,start_step}', '"route_entry"'::jsonb),
      '{workflow,steps,route_entry}', jsonb_build_object(
        'action','conditional_branch',
        'description','Dispatched from a work item, or hand-fired with an explicit page_id? Truthy on an ABSENT field routes to else_step rather than failing.',
        'config', jsonb_build_object(
          'condition','input_data.work_item_id',
          'then_step','load_work_item',
          'else_step','echo_page_ref'),
        'output_field','entry_route')),
    '{workflow,steps,load_work_item}', jsonb_build_object(
      'action','query_database',
      'description','Dispatched path: the claim action hands over only work_item_id, so read the page off the item.',
      'config', jsonb_build_object(
        'query','SELECT page_id::text AS page_id, id::text AS work_item_id FROM site_work_items WHERE id = $1::uuid',
        'params', jsonb_build_array('input_data.work_item_id'),
        'output_format','object'),
      'output_field','page_ref',
      'next_step','ensure_site_record')),
  updated_at = now()
WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions SET
  default_config = jsonb_set(
    jsonb_set(
      default_config,
      '{workflow,steps,echo_page_ref}', jsonb_build_object(
        'action','query_database',
        'description','Hand-fired path: echo input_data.page_id into the same shape the dispatched path produces, so load_page_target has ONE binding.',
        'config', jsonb_build_object(
          'query','SELECT $1::text AS page_id',
          'params', jsonb_build_array('input_data.page_id'),
          'output_format','object'),
        'output_field','page_ref',
        'next_step','ensure_site_record')),
    '{workflow,steps,load_page_target,config,params}', jsonb_build_array('page_ref.page_id')),
  updated_at = now()
WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify: assert the result, and FAIL the transaction if it is not what we meant.
-- A block of SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
-- result) — it has to RAISE.
DO $$
DECLARE wf jsonb; bad text;
BEGIN
  SELECT default_config->'workflow' INTO wf FROM agent_definitions
   WHERE type='copy-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF wf->>'start_step' <> 'route_entry' THEN RAISE EXCEPTION 'start_step not set'; END IF;
  IF wf->'steps'->'route_entry' IS NULL THEN RAISE EXCEPTION 'route_entry missing'; END IF;
  IF wf->'steps'->'load_work_item' IS NULL THEN RAISE EXCEPTION 'load_work_item missing'; END IF;
  IF wf->'steps'->'echo_page_ref' IS NULL THEN RAISE EXCEPTION 'echo_page_ref missing'; END IF;
  IF wf->'steps'->'load_page_target'->'config'->'params'->>0 <> 'page_ref.page_id'
    THEN RAISE EXCEPTION 'load_page_target still binds the old param'; END IF;

  -- Both branches must land on the same next step, or one path silently skips setup.
  IF wf->'steps'->'load_work_item'->>'next_step' <> wf->'steps'->'echo_page_ref'->>'next_step'
    THEN RAISE EXCEPTION 'the two entry paths do not converge'; END IF;

  -- 447's rule, re-asserted here because this migration adds steps: no step in
  -- copy-editor may write to a page.
  SELECT string_agg(k, ', ') INTO bad
    FROM jsonb_object_keys(wf->'steps') k
   WHERE wf->'steps'->k->>'action' IN
     ('save_page_sections','update_component_html','apply_section_edit','render_component',
      'rerender_page_sections','deploy_page');
  IF bad IS NOT NULL THEN
    RAISE EXCEPTION 'copy-editor gained a page-writing step (%) — 447 forbids this', bad;
  END IF;
END $$;

COMMIT;
