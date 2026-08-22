-- 564_dispatch_loop_passes_last_error_code_ROLLBACK.sql
--
-- Removes the single `last_error_code?` key from build-dispatch-loop's
-- call_handler input_mapping.
--
-- Effect: migration 563's prompt branch stops firing, because it gates on the
-- code and the code no longer crosses the allow-list. The retry then
-- regenerates blind — pre-345 behaviour, wasteful but not wrong. It does NOT
-- restore 533's misattributing block; that needs 563's own rollback.
--
-- Leaves 555's `last_error?` in place: the message is useful to any future
-- reader that does not need the classification.

BEGIN;

DO $guard$
DECLARE
  m jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF m IS NULL THEN
    RAISE EXCEPTION '564 ROLLBACK: call_handler input_mapping not found';
  END IF;
  IF NOT (m ? 'last_error_code?') THEN
    RAISE EXCEPTION '564 ROLLBACK: last_error_code? is not mapped — 564 is not applied';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,last_error_code?}'
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE
  m jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF m ? 'last_error_code?' THEN
    RAISE EXCEPTION '564 ROLLBACK VERIFY: the key is still mapped';
  END IF;
  IF NOT (m ? 'last_error?' AND m ? 'work_item_id' AND m ? 'spec' AND m ? 'site_id') THEN
    RAISE EXCEPTION '564 ROLLBACK VERIFY: the removal took a neighbouring key with it';
  END IF;

  RAISE NOTICE '564 ROLLBACK OK: last_error_code? removed, % keys remain',
    (SELECT count(*) FROM jsonb_object_keys(m));
END
$verify$;

COMMIT;
