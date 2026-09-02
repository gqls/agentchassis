-- ROLLBACK for 682. Removes the step's ai_service block. The gate then goes
-- LIVE-BUT-INERT again (records "no ai_service configuration resolvable" against
-- every violating point and repairs none) rather than failing — so prefer 681's
-- rollback if the intent is to remove the gate entirely.
BEGIN;
UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,repair_ordering_register,config,ai_service}'
 WHERE type='offer-analyser' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
DO $$
DECLARE v jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'repair_ordering_register'->'config' INTO v
    FROM agent_definitions WHERE type='offer-analyser' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF v ? 'ai_service' THEN RAISE EXCEPTION 'rollback did not remove the block'; END IF;
  RAISE NOTICE 'offer-analyser: register gate ai_service removed — gate is now live but inert';
END $$;
COMMIT;
