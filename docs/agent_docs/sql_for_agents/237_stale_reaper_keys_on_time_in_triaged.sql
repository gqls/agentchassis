-- 237_stale_reaper_keys_on_time_in_triaged.sql
--
-- bugs_open/070 — `stale-work-item-reaper` parks fresh requests because its
-- pre-query keys on ROW AGE, not on time spent in `triaged`. Fix candidate 1
-- (the file's own recommendation, "minimal"): key on `updated_at`.
--
-- WHY
-- ---
-- The task's own description states the intent:
--   "Marks triaged work items as unresolved if they have been waiting 48h+
--    without being claimed."
-- The predicate measures something else — how long ago the ROW WAS CREATED:
--
--     WHERE status = 'triaged'
--       AND pipeline = 'build'
--       AND created_at < NOW() - INTERVAL '48 hours'   <-- row birth
--       AND claimed_at IS NULL
--
-- For a row created and triaged in one go the two coincide, which is why this
-- has never bitten the loader-created path. For a RE-QUEUED item they diverge
-- completely: a row created five days ago is *born* eligible, so any re-queue of
-- it is reapable on the reaper's very next tick.
--
-- Reproduced live, unprompted, 2026-07-25 (bugs_open/070 §"Reproduced live"):
-- row dd50beb1 was re-queued to `triaged` at ~15:20 and parked `unresolved` with
-- the prefix "[stale: triaged 48h+]" at 15:32 — TWELVE MINUTES after entering
-- `triaged`. The label is false twice over.
--
-- Damage measured on the live DB 2026-07-27 before this change:
--   SELECT count(*) FILTER (WHERE summary LIKE '[stale: triaged 48h+] [stale%') AS doubly,
--          count(*) AS prefixed
--   FROM site_work_items WHERE summary LIKE '[stale: triaged 48h+]%';
--   -> prefixed = 91, doubly = 9
-- (The prefix is prepended in place and accumulates; there is no way back to the
-- original text. That is candidate 3 in the bug file and is NOT fixed here.)
--
-- WHAT THIS CHANGES
-- -----------------
-- One clause: `created_at <` becomes `updated_at <`. An item sitting untouched in
-- `triaged` has `updated_at` = the moment it was set `triaged`, which IS the
-- stated intent. `site_work_items` carries `trg_site_work_items_updated_at`
-- (BEFORE UPDATE ... set_updated_at()), so `updated_at` is maintained by the
-- database on every write — verified 2026-07-27 against pg_trigger.
--
-- RISK, STATED RATHER THAN WAVED THROUGH
-- --------------------------------------
-- Anything that touches the row for unrelated reasons bumps `updated_at` and
-- postpones reaping, so a genuinely wedged item that is periodically written to
-- would never be parked — the queue-deadlock this task exists to prevent.
--
-- The direct measurement CANNOT settle this and the bug file says so: at any
-- moment there is essentially no `triaged` build backlog (0 rows, 2026-07-27),
-- because the claimer drains it in ~2 minutes. So there is no population to
-- measure. [UNMEASURABLE by that query — not "measured and safe".]
--
-- What can be argued without measurement, and is [REASONED, NOT MEASURED]: an
-- item is wedged in `triaged` precisely BECAUSE nothing is touching it — if a
-- handler were writing to the row it would not be stuck. So `updated_at` and
-- "time since last activity" coincide for exactly the population the reaper
-- exists to catch. If a future thread wants certainty rather than an argument,
-- candidate 2 (a dedicated `triaged_at` column) is the durable answer.
--
-- Blast radius if the argument is wrong: an item that SHOULD be parked stays
-- `triaged` instead of going `unresolved`. `unresolved` is a terminal park, so
-- not reaping leaves the item claimable by `build-pipeline-trigger` (every 120s)
-- rather than removing it from the queue. That is the milder of the two errors,
-- and it is the reverse of the error being fixed here, which parks items that
-- are actively being worked.
--
-- WHAT THIS DOES NOT CHANGE
-- -------------------------
-- The pre-query's SHAPE is untouched: it still returns exactly one row
-- (`affected_count`, possibly '0'). That matters because bugs_closed/048 is about
-- this same scheduled task. 048 is fixed and live on kafka-scheduler v1.0.1146 —
-- a no-op pre-query now stamps `last_triggered_at` and rotates to the back of its
-- group's queue instead of pinning the `maintenance` slot — so a zero-row result
-- is safe. Nothing here depends on that, because the shape is unchanged.
--
-- SEED-FILE CLOBBER: CHECKED, AND THE ANSWER IS ODDER THAN EXPECTED
-- -----------------------------------------------------------------
-- The canonical definition lives at
-- `docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql:1141` and still carries
-- `created_at`, so the "config re-seed clobber" landmine applies in principle.
-- In practice it does not, because THAT BLOCK OF THE SEED FILE DOES NOT PARSE.
-- Line 1116 is `'' ||` — real SQL, not a comment — and the INSERT that follows is
-- swallowed into a string literal. Verified 2026-07-27 by running lines 1113–1152
-- inside BEGIN/ROLLBACK:
--   psql:<stdin>:43: ERROR:  syntax error at or near "''"
--   LINE 1: '' ||
-- So a re-seed of `020_scheduled_tasks.sql` ERRORS at that point rather than
-- reinstating the old predicate — and, worse, everything below line 1116 in that
-- file never applies either. That is the same class as the broken `wi.domain`
-- pre_query at :1009 already recorded by the bugfix_029_dispatch_gate thread. It
-- is NOT fixed here (different owner, different bug); it is recorded in
-- bugs_open/070 so the next reader does not "fix" this migration by editing a
-- file that cannot run.
--
-- ROLLBACK
-- --------
--   UPDATE scheduled_tasks
--   SET pre_query = replace(pre_query,
--         'AND updated_at < NOW() - INTERVAL ''48 hours''',
--         'AND created_at < NOW() - INTERVAL ''48 hours''')
--   WHERE name = 'stale-work-item-reaper';
-- The exact pre-change text is also stored verbatim in `migration_backups`
-- (migration_name = '237_stale_reaper_keys_on_time_in_triaged.sql').
--
-- NOT AN agent_definitions CHANGE, so no snapshot_agent() call: this row lives in
-- `scheduled_tasks`. `migration_backups` is the equivalent before-image.

BEGIN;

-- Before-image, verbatim.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '237_stale_reaper_keys_on_time_in_triaged.sql',
       'scheduled_tasks',
       id::text,
       jsonb_build_object('name', name, 'pre_query', pre_query, 'description', description),
       'bugs_open/070 candidate 1: pre_query keyed on created_at before this change'
FROM scheduled_tasks
WHERE name = 'stale-work-item-reaper';

-- Refuse to run if the exact clause is not there: a silent no-op would leave the
-- bug in place while reporting success.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM scheduled_tasks
        WHERE name = 'stale-work-item-reaper'
          AND pre_query LIKE '%AND created_at < NOW() - INTERVAL ''48 hours''%'
    ) THEN
        RAISE EXCEPTION '237: the exact clause "AND created_at < NOW() - INTERVAL ''''48 hours''''" was not found in stale-work-item-reaper.pre_query — already fixed, or the text moved. Re-survey rather than forcing this.';
    END IF;
END $$;

UPDATE scheduled_tasks
SET pre_query = replace(pre_query,
        'AND created_at < NOW() - INTERVAL ''48 hours''',
        'AND updated_at < NOW() - INTERVAL ''48 hours'''),
    updated_at = NOW()
WHERE name = 'stale-work-item-reaper';

-- Post-conditions, inside the same transaction so a failure rolls the file back.
DO $$
DECLARE
    n   integer;
    pq  text;
BEGIN
    SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'stale-work-item-reaper';
    IF n <> 1 THEN
        RAISE EXCEPTION '237: expected exactly 1 stale-work-item-reaper row, found %', n;
    END IF;

    SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'stale-work-item-reaper';

    IF pq NOT LIKE '%AND updated_at < NOW() - INTERVAL ''48 hours''%' THEN
        RAISE EXCEPTION '237: post-condition failed — the updated_at clause is not present';
    END IF;
    IF pq LIKE '%AND created_at < NOW() - INTERVAL%' THEN
        RAISE EXCEPTION '237: post-condition failed — a created_at age clause is still present';
    END IF;
    -- The shape must be unchanged (see the 048 note above).
    IF pq NOT LIKE '%SELECT COUNT(*)::text as affected_count FROM stale%' THEN
        RAISE EXCEPTION '237: post-condition failed — the pre_query no longer returns affected_count';
    END IF;
    IF pq NOT LIKE '%AND claimed_at IS NULL%' OR pq NOT LIKE '%AND pipeline = ''build''%' THEN
        RAISE EXCEPTION '237: post-condition failed — an unrelated clause was lost';
    END IF;

    RAISE NOTICE '237: stale-work-item-reaper now keys on updated_at (time spent in triaged), not created_at (row birth).';
END $$;

COMMIT;
