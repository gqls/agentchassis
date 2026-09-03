-- Reverses 726: restores the pre-ruling £200 wording in the delivery email body.
-- Only meaningful if the 2026-08-26 ruling is itself reversed.
BEGIN;
DO $$
DECLARE body text;
BEGIN
  SELECT default_config #>> '{workflow,steps,send_email,config,body_template}' INTO body
    FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('a one-off 59.99 pounds' IN body) = 0 THEN
    RAISE EXCEPTION '726 ROLLBACK: nothing to reverse — the body does not carry the 59.99 wording';
  END IF;
  PERFORM snapshot_agent('delivery-email-sender', '726 ROLLBACK: pre-restore');
END $$;
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,send_email,config,body_template}',
         to_jsonb(replace(default_config #>> '{workflow,steps,send_email,config,body_template}',
                          'a one-off 59.99 pounds', 'a one-off 200 pounds')), false),
       updated_at = now()
 WHERE type='delivery-email-sender' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
COMMIT;
