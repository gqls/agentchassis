-- ROLLBACK for 775 (bugs_open/477 step B, part 2).
--
-- Removes the follow-up sender and its schedule. Safe while the schedule is
-- disabled and nothing has been sent; the guard below refuses otherwise, because
-- deleting the agent while sites carry followup_sent_at leaves stamps whose
-- meaning has no definition anywhere.

BEGIN;

DO $$
DECLARE stamped bigint; enabled_now boolean;
BEGIN
  SELECT enabled INTO enabled_now FROM scheduled_tasks WHERE name = 'delivery-followup-send';
  IF enabled_now THEN
    RAISE EXCEPTION '775 ROLLBACK REFUSED: delivery-followup-send is ENABLED. Disable it first and confirm no run is in flight.';
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.columns
              WHERE table_schema='public' AND table_name='sites' AND column_name='followup_sent_at') THEN
    SELECT count(*) INTO stamped FROM sites WHERE followup_sent_at IS NOT NULL;
    IF stamped > 0 THEN
      RAISE EXCEPTION '775 ROLLBACK REFUSED: % site(s) have already been followed up. Removing the agent leaves those stamps with nothing that defines them, and a re-seed would not know they are spent.', stamped;
    END IF;
  END IF;
END $$;

DELETE FROM scheduled_tasks WHERE name = 'delivery-followup-send';
UPDATE agent_definitions SET deleted_at = now(), is_active = false
 WHERE type = 'delivery-followup-sender' AND deleted_at IS NULL;

COMMIT;
