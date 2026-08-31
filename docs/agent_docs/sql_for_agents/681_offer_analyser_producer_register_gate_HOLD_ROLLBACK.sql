-- ROLLBACK for 681. Restores the chain to cardinals -> write and removes the step.
-- Safe to run whether or not the image carries the action.
BEGIN;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps,write_offer_ordering,config,spec_data}',
         '"ordering_checked.object"'::jsonb, false)
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config, '{workflow,steps,verify_ordering_cardinals,next_step}',
         '"write_offer_ordering"'::jsonb, false)
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,repair_ordering_register}'
 WHERE type = 'offer-analyser' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE v_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO v_cfg FROM agent_definitions
   WHERE type = 'offer-analyser' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF v_cfg ? 'repair_ordering_register' THEN
    RAISE EXCEPTION 'rollback did not remove the step';
  END IF;
  IF v_cfg->'write_offer_ordering'->'config'->>'spec_data' <> 'ordering_checked.object' THEN
    RAISE EXCEPTION 'rollback did not restore the write source';
  END IF;
  RAISE NOTICE 'offer-analyser: producer register gate removed, chain restored';
END $$;

COMMIT;
