-- 555_dispatch_loop_passes_last_error_to_handlers_ROLLBACK.sql
--
-- Removes the single `last_error?` key from build-dispatch-loop's call_handler
-- input_mapping. Safe with the Go half and 533 still deployed: the prompt block
-- is {{if}}-guarded, so with the key unmapped it simply never renders — which
-- is exactly the pre-555 behaviour.

BEGIN;

DO $guard$
DECLARE m jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF m IS NULL OR NOT (m ? 'last_error?') THEN
    RAISE EXCEPTION '555 ROLLBACK: last_error? is not mapped — nothing to undo';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = #- '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,last_error?}',
    updated_at = NOW()
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE m jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF m ? 'last_error?' THEN
    RAISE EXCEPTION '555 ROLLBACK VERIFY: the key is still present';
  END IF;
  IF NOT (m ? 'work_item_id' AND m ? 'spec') THEN
    RAISE EXCEPTION '555 ROLLBACK VERIFY: pre-existing keys were damaged';
  END IF;
END
$verify$;

COMMIT;
