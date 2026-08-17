-- ROLLBACK for 450_tool_auditor_review_items_park_at_needs_human_review.sql
-- Removes the status key from create_review_item.config, restoring the pre-450
-- shape (status absent → the binary's 'triaged' default → bugs_open/291's bleed).
-- Only run this if 450 itself has to come out; it does NOT touch handler_agent.

BEGIN;

SELECT snapshot_agent('tool-auditor',
  '450_tool_auditor_review_items_park_at_needs_human_review_ROLLBACK: pre-rollback');

DO $$
DECLARE
  n int;
  s text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
     AND type = 'tool-auditor'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '450_ROLLBACK: expected exactly 1 live tool-auditor row at the pinned id, found %', n;
  END IF;

  SELECT default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,status}'
    INTO s FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';
  IF s IS DISTINCT FROM 'needs_human_review' THEN
    RAISE EXCEPTION '450_ROLLBACK: status key is % — not the shape 450 wrote; refusing', s;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}',
         (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') - 'status'),
       updated_at = now()
 WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5'
   AND type = 'tool-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  has_status boolean;
BEGIN
  SELECT (default_config #> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config}') ? 'status'
    INTO has_status FROM agent_definitions
   WHERE id = '2ec19d27-1d77-4577-a7f8-6cbf9ef015f5';
  IF has_status THEN
    RAISE EXCEPTION '450_ROLLBACK: status key still present after removal';
  END IF;
END $$;

COMMIT;
