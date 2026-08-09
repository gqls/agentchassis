-- 321_provocation_scheduler_operator_handle.sql
--
-- Seats `provocation-scheduler-manual`: the operator's handle for dating
-- human-approved provocations. **Deliberately has NO `scheduled_tasks` row, and
-- must never be given one.**
--
-- WHY IT IS MANUAL, AND WHY THAT IS A CONTROL RATHER THAN AN OMISSION
-- The publisher needs `status='approved' AND publish_on IS NOT NULL AND
-- human_approved_at IS NOT NULL`. Three components own those three facts, and no
-- single one can publish alone — that separation is the safety property, not the
-- gate.
--
-- Owner ruling 2026-08-09 put a human back in the publish path. Migration 320 gave
-- that its own column so the human's consent stopped being carried implicitly by
-- the date. Dating is therefore ordinary scheduling again — **but only while the
-- stamp is genuinely applied by a person.** Putting this agent on a cron alongside
-- an automated stamp would reassemble the fully-automatic path the ruling removed,
-- and it would look like plumbing while doing it.
--
--   gate_provocation       -> schedulable (it only ever sets a verdict)
--   generate_provocations  -> schedulable (it only ever writes undated drafts)
--   schedule_provocations  -> OPERATOR-INVOKED  <-- this file
--
-- An `agent_definitions` row with no `scheduled_tasks` row is exactly how you make
-- an action invokable-but-not-automatic on this platform. It is not "wiring to
-- publish"; nothing fires it but a person.
--
-- SCHEDULE DEPTH: 6 days (owner ruling 2026-08-09, "I think I decided 6 days ahead
-- was ok"). `max_assign` caps a single run at that, so one invocation cannot empty
-- a large approved backlog into a long unattended runway.
--
-- Idempotent; safe to re-run.

BEGIN;

INSERT INTO agent_definitions (type, display_name, category, description, default_config, is_active)
SELECT
    'provocation-scheduler-manual',
    'Provocation Scheduler (operator-invoked)',
    'content',
    'Dates human-approved provocations, one per calendar day, forward only. OPERATOR-INVOKED BY DESIGN — must never be given a scheduled_tasks row: with an automated stamp that would reassemble the fully-automatic publish path the owner removed on 2026-08-09.',
    jsonb_build_object(
        'processing_mode', 'task',
        'workflow', jsonb_build_object(
            'start_step', 'schedule',
            'steps', jsonb_build_object(
                'schedule', jsonb_build_object(
                    'action', 'schedule_provocations',
                    'description', 'Assign publish dates to approved, human-approved, undated provocations',
                    'next_step', 'complete',
                    'config', jsonb_build_object(
                        'domain', 'vonc.com',
                        'max_assign', 6
                    )
                ),
                'complete', jsonb_build_object(
                    'action', 'complete_workflow',
                    'description', 'Scheduling run finished'
                )
            )
        )
    ),
    true
WHERE NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'provocation-scheduler-manual'
       AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
);

DO $$
DECLARE n_agent int; n_sched int;
BEGIN
  SELECT count(*) INTO n_agent FROM agent_definitions
   WHERE type='provocation-scheduler-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF n_agent <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active scheduler agent, found %', n_agent; END IF;

  -- The control this file exists to assert. If a scheduled_tasks row ever appears
  -- for this agent, the human-approval step has been silently re-automated.
  SELECT count(*) INTO n_sched FROM scheduled_tasks WHERE target_agent_type='provocation-scheduler-manual';
  IF n_sched <> 0 THEN
    RAISE EXCEPTION 'a scheduled_tasks row exists for provocation-scheduler-manual (%). This agent must be operator-invoked; a schedule here re-automates the step the owner took back on 2026-08-09.', n_sched;
  END IF;

  RAISE NOTICE 'scheduler seated as operator-invoked; 0 scheduled_tasks rows, as required';
END $$;

COMMIT;

-- HOW TO INVOKE: dispatch agent_type 'provocation-scheduler-manual' on
-- system.agent.generic.requests (envelope: see 097_TRIGGER; one-line payload, since
-- kcat -P splits stdin on newlines into separate messages).
--
-- WHAT IT WILL AND WILL NOT TOUCH: only rows that are already `status='approved'`
-- AND `human_approved_at IS NOT NULL` AND `publish_on IS NULL`. It never dates a
-- draft, never re-dates a dated row, and never assigns a date in the past — the gap
-- behind today is deliberately not backfilled (PLAN §10.5).
