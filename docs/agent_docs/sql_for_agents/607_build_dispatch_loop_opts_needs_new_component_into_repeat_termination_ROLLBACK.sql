-- ROLLBACK for 607: remove the opt-in key. The mechanism itself (the Go
-- ladder) is untouched — with the key gone, behaviour is byte-identical to
-- pre-607 on the next dispatch; no restart needed (config is read per call).

BEGIN;

DO $guard$
DECLARE
  c jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'mark_failed'->'config'
    INTO c
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF c IS NULL THEN
    RAISE EXCEPTION '607 ROLLBACK: mark_failed config not found at the expected path';
  END IF;
  IF NOT (c ? 'stop_on_repeat_failure_item_types') THEN
    RAISE EXCEPTION '607 ROLLBACK: the key is not present — 607 is not applied (or already rolled back)';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = default_config
      #- '{workflow,steps,process_item,config,sub_workflow,steps,mark_failed,config,stop_on_repeat_failure_item_types}'
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE
  c jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'mark_failed'->'config'
    INTO c
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF c ? 'stop_on_repeat_failure_item_types' THEN
    RAISE EXCEPTION '607 ROLLBACK VERIFY: the key survived the delete';
  END IF;
  IF NOT (c ? 'work_item_id' AND c ? 'error_message') THEN
    RAISE EXCEPTION '607 ROLLBACK VERIFY: an anchor key was lost';
  END IF;
  RAISE NOTICE '607 ROLLBACK OK: opt-in removed; behaviour is pre-607 on the next dispatch';
END
$verify$;

COMMIT;
