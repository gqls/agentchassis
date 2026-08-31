-- 669 ROLLBACK: restore the pre-669 exemplar phrase from the snapshot's anchor.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_cfg text;
BEGIN
  SELECT default_config::text INTO v_cfg FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;
  IF v_cfg NOT LIKE '%a text-free mark%' THEN
    RAISE EXCEPTION '669 rollback probe: 669 does not appear applied';
  END IF;
  UPDATE agent_definitions
     SET default_config = replace(default_config::text,
           'no lettering or words of any kind (a text-free mark: the brand name is set in HTML beside the logo, never rendered in the image)',
           'no text outside the wordmark itself')::jsonb,
         updated_at = now()
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;
END $$;
COMMIT;
