-- 685 — vonc.com provocations: put the generator and the scheduler on a cadence,
--       and raise the shelf to 14 days.
--
-- OWNER INSTRUCTION 2026-09-02: "I'd like to make the challenges change every day
-- and not be restrained by needing my permission." Answers to the three sizing
-- questions the same day: over-generate to absorb the now-fatal readability rail;
-- keep 14 days of shelf; if it runs dry, "create a new set and carry on" (so this
-- is self-refilling by design, with no alerting step and no human in the loop).
--
-- WHY THE SITE NEEDS IT: vonc.com served the SAME provocation from 2026-08-22 to
-- 2026-09-02 — eleven days — under a heading that reads "Today's Provocation".
-- The publisher was never broken (6h schedule, ran clean throughout); it had
-- nothing to publish. Both producing agents are named `-manual` and have no
-- `scheduled_tasks` row at all, so nothing refilled the shelf.
--
-- ============================================================================
-- ⚠⚠  WHY THIS IS _HOLD AND NOT AN ORDINARY MIGRATION — READ BEFORE APPLYING
-- ============================================================================
-- This file MUST NOT be applied until the Go change of commit 326370d6c is LIVE
-- in the running chassis image. It is an ordering constraint, not caution:
--
--   Applied early, the generator starts writing drafts that the OLD binary gates
--   with the readability rail still ADVISORY. Those rows land `status='approved'`
--   without ever facing the rail. `loadGateCandidates` never re-gates an approved
--   row — by design, so a model's drift cannot retract a published provocation —
--   so once the new binary rolls, that batch is publishable FOR EVER without the
--   rail having applied to it. The site would then serve exactly the prose the
--   rail exists to stop, and nothing would ever flag it.
--
-- The guard below enforces this rather than trusting the operator to read: it
-- refuses to apply until at least one row carries `gate_version = '3'`, which
-- only the new binary can write. That is an ARTEFACT check, not a tag check —
-- it proves the new code actually gated something, not merely that a release
-- happened.
--
-- APPLY ORDER
--   1. Owner rolls the fleet (`make release`) so 326370d6c is live.
--   2. Fire ONE attended generator run and read it:
--        agent_type `provocation-generator-manual` on system.agent.generic.requests
--      Confirm the rail now REJECTS (look for a fatal `hard_to_read` reason) and
--      that new rows carry gate_version '3'.
--   3. Apply this file by hand, then `run-migrations.sh --record-only`.
--
-- SUPERSEDES TWO EARLIER GUARDS, deliberately and on the record:
--   * `321_provocation_scheduler_operator_handle.sql` RAISEs if a scheduled_tasks
--     row exists for the scheduler, on the ground that a schedule "re-automates
--     the step the owner took back on 2026-08-09". The owner has now reversed
--     that ruling (his third position on the question), so the premise is gone.
--     ⚠ 321 WILL THEREFORE FAIL IF RE-RUN. That is expected from today; it is a
--     record of a ruling that no longer holds, not a live control.
--   * `371_provocation_generator_operator_handle.sql` RAISEs on the same shape,
--     but its condition was "until one attended run has been read" — which was
--     satisfied by the attended runs of 2026-08-10 and 2026-08-12 that the owner
--     read and acted on (he binned eight and accepted eight).
-- Both agents' `description` fields also assert "OPERATOR-INVOKED BY DESIGN —
-- must never be given a scheduled_tasks row". Those are REWRITTEN below, because
-- a description that contradicts the live schedule is read as ground truth by
-- council seats and by the next session.

BEGIN;

-- ---------------------------------------------------------------------------
-- GUARD: refuse to apply against an image that predates the fatal rail.
-- ---------------------------------------------------------------------------
DO $$
DECLARE n_v3 int; n_gen int; n_sch int;
BEGIN
  SELECT count(*) INTO n_v3 FROM provocations
   WHERE gate_verdict->>'gate_version' = '3';
  IF n_v3 = 0 THEN
    RAISE EXCEPTION
      'REFUSING TO APPLY: no provocation has been gated by gate_version 3, so the '
      'binary carrying the fatal readability rail (commit 326370d6c) is not live yet. '
      'Applying now would let the generator bank approved-but-never-railed drafts that '
      'can never be re-gated. Roll the fleet, fire one attended generator run, then '
      'apply this file. See the header for the full ordering argument.';
  END IF;

  SELECT count(*) INTO n_gen FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF n_gen <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active generator agent, found %', n_gen; END IF;

  SELECT count(*) INTO n_sch FROM agent_definitions
   WHERE type='provocation-scheduler-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF n_sch <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active scheduler agent, found %', n_sch; END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. Shelf depth 6 -> 14 (owner's answer, 2026-09-02).
--    The 6 was his 2026-08-12 depth ruling, chosen so "one bad batch can never
--    fill a long stretch nobody is watching". That reasoning assumed he was
--    watching; unattended, the binding risk inverts — a thin shelf is how the
--    site went eleven days stale. 14 is still well inside `max_assign <= 30`,
--    which the action enforces.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,schedule,config,max_assign}',
         '14'::jsonb,
         true),
       description = 'Dates approved provocations, one per category per calendar day, forward only. '
                     'SCHEDULED DAILY since 2026-09-02 (owner instruction: no permission step). '
                     'Was operator-invoked between 2026-08-09 and 2026-09-02, when a human stamp '
                     'was required before dating; that stamp was removed from all three queries in '
                     'commit 326370d6c. Never dates today — earliest is tomorrow — which is what '
                     'bounds how fast anything can reach the site.'
 WHERE type='provocation-scheduler-manual' AND is_active
   AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

-- ---------------------------------------------------------------------------
-- 2. Generator description — the schedule makes the old text false.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
   SET description = 'Writes candidate provocations into the pool as drafts and gates them. '
                     'SCHEDULED every 12h since 2026-09-02 (owner instruction: no permission step), '
                     'but only fires when the shelf is short — see the pre_query on the '
                     'provocation-shelf-refill task, which skips the run entirely once 14 days of '
                     'inventory exist. Batch size is set by that pre_query and never exceeds 4, '
                     'the size proven against the 8000-token budget on 2026-08-10.'
 WHERE type='provocation-generator-manual' AND is_active
   AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

-- ---------------------------------------------------------------------------
-- 3. The two schedules.
--
-- THE GENERATOR IS DEMAND-DRIVEN, NOT PERIODIC. A plain periodic generator would
-- grow the pool without bound: the site consumes one provocation a day and the
-- scheduler dates up to 14, so unconditional generation would push publish dates
-- months out and spend model credits for ever. The pre_query returns NO ROWS when
-- the shelf is full, and `runPreQuery` treats no rows as "skip this task" — so a
-- full shelf costs one cheap SELECT and nothing else.
--
-- BATCH SIZE IS DELIBERATELY CAPPED AT 4, not raised to absorb the fatal rail in
-- one go. The owner asked to over-generate; that is done by CADENCE (twice daily)
-- rather than by batch, because 4-at-8000-tokens is the only combination this lane
-- has actually proven, and the failure mode of a bigger batch is a truncated
-- completion — which this pipeline has already been bitten by (stop_reason
-- max_tokens on 2026-08-10) and which presents as success.
-- Deficit x2 is the over-generation factor, so a shelf 1 short still asks for 2.
-- ---------------------------------------------------------------------------
INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds)
VALUES
  ('provocation-shelf-refill',
   'Writes and gates new vonc.com provocations, but only while the shelf holds fewer than 14 '
   'days of inventory. Skips entirely otherwise (pre_query returns no rows).',
   43200,                                   -- 12h
   'provocation-generator-manual',
   'system.agent.generic.requests',
   '{"domain": "vonc.com"}'::jsonb,
   'provocation-generation', 1,
   $q$
     SELECT LEAST(4, GREATEST(0, 14 - count(*)) * 2)::int AS count
       FROM provocations
      WHERE domain = 'vonc.com'
        AND status = 'approved'
        AND (publish_on IS NULL OR publish_on > current_date)
     HAVING GREATEST(0, 14 - count(*)) > 0
   $q$,
   true, 900),

  ('provocation-date-assign',
   'Dates approved, undated vonc.com provocations, up to 14 at a time, tomorrow onwards. '
   'No pre_query: the action already no-ops when nothing is undated.',
   86400,                                   -- 24h
   'provocation-scheduler-manual',
   'system.agent.generic.requests',
   '{"domain": "vonc.com", "max_assign": 14}'::jsonb,
   'provocation-generation', 1,
   NULL,
   true, 600);

-- ---------------------------------------------------------------------------
-- VERIFY — DO/RAISE, not a SELECT. A verify block made of SELECTs cannot stop
-- the COMMIT: ON_ERROR_STOP ignores a non-empty result set, so a "failing"
-- SELECT prints and commits anyway. Every assertion below can abort.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  n int; ma int; pq text; c int; inv int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name IN ('provocation-shelf-refill','provocation-date-assign') AND enabled;
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 enabled provocation tasks, found %', n; END IF;

  SELECT (default_config#>>'{workflow,steps,schedule,config,max_assign}')::int INTO ma
    FROM agent_definitions
   WHERE type='provocation-scheduler-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF ma <> 14 THEN RAISE EXCEPTION 'max_assign is %, expected 14', ma; END IF;

  SELECT pre_query INTO pq FROM scheduled_tasks WHERE name='provocation-shelf-refill';
  IF pq IS NULL OR pq = '' THEN
    RAISE EXCEPTION 'the refill task has no pre_query — it would generate unconditionally and grow the pool without bound';
  END IF;

  -- INDUCE THE PRE-QUERY. Asserting that a gate EXISTS is not asserting that it
  -- WORKS; this runs the real statement and checks the answer against the real
  -- inventory, so a typo that returns nothing (task never fires again) or returns
  -- a row unconditionally (pool grows for ever) both abort here.
  SELECT count(*) INTO inv FROM provocations
   WHERE domain='vonc.com' AND status='approved'
     AND (publish_on IS NULL OR publish_on > current_date);

  EXECUTE 'SELECT count(*) FROM (' || pq || ') s' INTO c;

  IF inv < 14 AND c <> 1 THEN
    RAISE EXCEPTION 'inventory is % (short of 14) but the pre_query returned % rows; it should return exactly 1 so the refill fires', inv, c;
  END IF;
  IF inv >= 14 AND c <> 0 THEN
    RAISE EXCEPTION 'inventory is % (>= 14) but the pre_query returned % rows; it should return 0 so the refill SKIPS', inv, c;
  END IF;

  RAISE NOTICE 'OK: 2 tasks enabled, max_assign=14, inventory=%, pre_query returned % row(s) as expected', inv, c;
END $$;

COMMIT;
