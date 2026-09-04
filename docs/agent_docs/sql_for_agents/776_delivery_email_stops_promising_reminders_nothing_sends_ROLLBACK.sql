-- ROLLBACK for 776 (bugs_open/477).
--
-- ⚠ THIS RESTORES A FALSE STATEMENT TO A CUSTOMER-FACING EMAIL. "so we stop
-- reminding you" is untrue until a follow-up sender exists. Run this only to
-- reach a known state while something else is diagnosed, never as a repair.
--
-- If you are here because the follow-up sender has SHIPPED and the stronger
-- wording is now true, do NOT use this file: write a forward migration that says
-- so and records why, so the record shows the clause becoming true rather than
-- an undo.

BEGIN;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
    INTO tpl FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN
    RAISE EXCEPTION '776 ROLLBACK: body_template absent — the step moved';
  END IF;
  IF position('press the button here to tell us you have moved:' in tpl) = 0 THEN
    RAISE EXCEPTION '776 ROLLBACK: 776''s line is not present — nothing to roll back, or another lane edited this template since';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,send_email,config,body_template}',
      to_jsonb(replace(
        default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template',
        'press the button here to tell us you have moved:',
        'press the button here so we stop reminding you:')),
      false),
    updated_at = NOW()
WHERE type='delivery-email-sender' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
    INTO tpl FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('so we stop reminding you' in tpl) = 0 THEN
    RAISE EXCEPTION '776 ROLLBACK VERIFY: the original clause was not restored';
  END IF;
  IF position('{{confirm_link}}' in tpl) = 0 THEN
    RAISE EXCEPTION '776 ROLLBACK VERIFY: {{confirm_link}} was disturbed';
  END IF;
  RAISE NOTICE '776 ROLLBACK: the false clause is back. It is still false.';
END $$;

COMMIT;
