-- ROLLBACK for 774 (bugs_open/477 step B, part 1).
--
-- Dropping the column removes the follow-up's at-most-once claim. That is safe
-- ONLY while nothing writes it: the writer is delivery.ClaimFollowup, which is
-- unreachable until an image carrying it rolls AND 775's agent is enabled (it is
-- seeded disabled). If either has happened, DO NOT run this — dropping the
-- column while the sender is live turns "at most one follow-up per site" into
-- "one per scheduler tick", and it takes the record of what has already been
-- sent with it.
--
-- The guard below refuses rather than warns.

BEGIN;

DO $$
DECLARE
  stamped bigint;
  live_task int;
BEGIN
  SELECT count(*) INTO stamped FROM sites WHERE followup_sent_at IS NOT NULL;
  IF stamped > 0 THEN
    RAISE EXCEPTION '774 ROLLBACK REFUSED: % site(s) carry followup_sent_at. Dropping it discards the record of follow-ups already sent, and any enabled sender would then re-send to those customers.', stamped;
  END IF;

  SELECT count(*) INTO live_task FROM scheduled_tasks
   WHERE target_agent_type = 'delivery-followup-sender' AND enabled;
  IF live_task > 0 THEN
    RAISE EXCEPTION '774 ROLLBACK REFUSED: the delivery-followup-sender schedule is ENABLED. Disable it first, or this drop removes the only thing stopping a nightly email to every unconfirmed customer.';
  END IF;
END $$;

ALTER TABLE sites DROP COLUMN IF EXISTS followup_sent_at;

COMMIT;
