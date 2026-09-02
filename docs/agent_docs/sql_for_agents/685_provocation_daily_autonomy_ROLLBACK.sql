-- ROLLBACK for 685 — removes the two schedules and restores the pre-2026-09-02
-- shelf depth and descriptions.
--
-- WHAT THIS DOES NOT DO, deliberately: it does not restore the human-approval
-- predicate. That lives in Go (commit 326370d6c), not in config, so reverting it
-- is a code change and a roll — not this file. Running this alone leaves the
-- pipeline exactly as it was between the roll and 685: capable of publishing
-- without a stamp, but with nothing driving it. That is the safe direction —
-- the site goes quiet rather than publishing unattended.
--
-- It also leaves any provocations already written or dated in place. Retire them
-- by hand if that is what you want; deleting rows here would destroy the record
-- of what the automated pipeline actually produced, which is the only evidence
-- of whether it worked.

BEGIN;

DELETE FROM scheduled_tasks
 WHERE name IN ('provocation-shelf-refill','provocation-date-assign');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,schedule,config,max_assign}',
         '6'::jsonb,
         true),
       description = 'Dates human-approved provocations, one per calendar day, forward only. '
                     'OPERATOR-INVOKED BY DESIGN — must never be given a scheduled_tasks row: '
                     'with an automated stamp that would reassemble the fully-automatic publish '
                     'path the owner removed on 2026-08-09.'
 WHERE type='provocation-scheduler-manual' AND is_active
   AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

UPDATE agent_definitions
   SET description = 'Writes candidate provocations into the pool as drafts. OPERATOR-INVOKED: '
                     'this seat is fired by hand until an attended run has been read.'
 WHERE type='provocation-generator-manual' AND is_active
   AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

DO $$
DECLARE n int; ma int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name IN ('provocation-shelf-refill','provocation-date-assign');
  IF n <> 0 THEN RAISE EXCEPTION 'rollback left % provocation task(s) behind', n; END IF;

  SELECT (default_config#>>'{workflow,steps,schedule,config,max_assign}')::int INTO ma
    FROM agent_definitions
   WHERE type='provocation-scheduler-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF ma <> 6 THEN RAISE EXCEPTION 'max_assign is % after rollback, expected 6', ma; END IF;

  RAISE NOTICE 'OK: both schedules removed, max_assign back to 6';
END $$;

COMMIT;
