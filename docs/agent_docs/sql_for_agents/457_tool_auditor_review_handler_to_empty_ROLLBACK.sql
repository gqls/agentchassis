-- ROLLBACK for 457_tool_auditor_review_handler_to_empty.sql (bugs_open/291 phase 3)
--
-- Restores the phantom handler name on tool-auditor's create_review_item step and on
-- the rows 457 swept. ⚠ Only meaningful while the chassis binary carrying the relaxed
-- create_work_item validation is live — that is what makes an empty config handler
-- legal in the first place. Rolling this back does NOT reintroduce bugs_open/291's
-- bleed (migration 450's status key, applied separately, is what parks the items);
-- it only puts the inert phantom name back.

BEGIN;

SELECT snapshot_agent('tool-auditor',
  '457_tool_auditor_review_handler_to_empty_ROLLBACK: pre-rollback');

DO $$
DECLARE
  n int;
  h text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
     AND type = 'tool-auditor'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '457_ROLLBACK: expected exactly 1 live tool-auditor row at the pinned id, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}'
    INTO h FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';
  IF h IS DISTINCT FROM '' THEN
    RAISE EXCEPTION '457_ROLLBACK: handler_agent is %, not the empty string 457 wrote; refusing', h;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}',
         '"hitl-review"'),
       updated_at = now()
 WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
   AND type = 'tool-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Rows swept by 457 carry the handler_291 stamp; restore only those.
UPDATE site_work_items
   SET handler_agent = 'hitl-review',
       result = result - 'handler_291',
       updated_at = now()
 WHERE result ? 'handler_291';

DO $$
DECLARE
  h text;
  leftover int;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,handler_agent}'
    INTO h FROM agent_definitions WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';
  IF h IS DISTINCT FROM 'hitl-review' THEN
    RAISE EXCEPTION '457_ROLLBACK: post-state handler is %, expected hitl-review', h;
  END IF;
  SELECT count(*) INTO leftover FROM site_work_items WHERE result ? 'handler_291';
  IF leftover > 0 THEN
    RAISE EXCEPTION '457_ROLLBACK: % row(s) still carry the handler_291 stamp', leftover;
  END IF;
END $$;

COMMIT;
