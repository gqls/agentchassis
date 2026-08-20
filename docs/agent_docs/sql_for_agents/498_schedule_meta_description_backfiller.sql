-- 498 — put the meta-description backfiller on its own schedule
--       (bugs_open/320, SEO-004; OWNER INSTRUCTION 2026-08-20: "we can put the
--       backfiller on a separate schedule")
--
-- WHY NOW AND NOT BEFORE. The backfiller was deliberately hand-driven until a full
-- fleet run had been produced and read, because it writes public copy under the
-- owner's name and `bugs_closed/103` is what happens when that kind of generator
-- runs unattended before anyone looks. That run has now happened: 407 of 731 pages
-- empty -> 47 of 748, ~571 pages written, mean length 129 chars, copy inspected on
-- three sites. The caution has been paid for, so the schedule is the right next step.
--
-- "SEPARATE" IS LOAD-BEARING: its own `concurrency_group` with `max_concurrent = 1`,
-- not the shared `dispatch` group. Two concurrent runs would both pick the same site
-- from the pre-query and pay twice for the same LLM call. It could not corrupt
-- anything — `overwrite_existing` defaults false, so the loser writes nothing — but
-- paying twice for a no-op is still paying twice.
--
-- ── HOW IT PICKS WORK, AND WHY IT GOES QUIET BY ITSELF ──────────────────────
--
-- `cmd/scheduler/main.go` runs `pre_query`, takes THE FIRST ROW, and merges it into
-- `input_data` (`runPreQuery`, :442). So a pre-query returning `site_id` and
-- `domain` hands the agent exactly the input it takes by hand today.
--
-- **A pre-query returning NO ROWS is a clean no-op, not an error** (:205-220): the
-- scheduler stamps the timestamps and moves on, deliberately, so the task rotates to
-- the back of its queue instead of pinning the head of its group every tick
-- (`bugs_open/048`). That is what makes this safe to leave enabled for ever: when
-- every fillable page has a description the query returns nothing, the task no-ops,
-- and it costs one cheap SELECT an hour. **It does not need switching off when the
-- backlog is drained, and it wakes up on its own when new pages appear.**
--
-- The query picks the site with the MOST fillable empty pages, so the backlog drains
-- worst-first, and it requires the page to actually have rendered components —
-- `43` of the currently-empty pages have NONE, and a page with no content cannot be
-- described from its content. Those are a floor, not a backlog, and this query
-- correctly never selects them. **Do not "fix" that by relaxing the EXISTS clause;
-- the alternative to declining is inventing.**
--
-- INTERVAL 3600s. The workflow handles 25 pages per run in one LLM call, and new
-- pages arrive at roughly a dozen a day, so hourly clears any realistic inflow many
-- times over while costing nothing when idle. It is not a throughput knob — if a big
-- backlog ever needs draining fast, run the script by hand rather than dropping this.
--
-- SAFETY, unchanged and worth restating because a schedule is where people stop
-- reading: `overwrite_existing` is ABSENT from the workflow's save step, so it takes
-- the action's default of false and this task CANNOT replace existing copy. It fills
-- blanks. The action also runs the site's voice gate and the banned-claims sweep
-- before every write (the owner's condition for waiving the review pass), and an
-- unreadable gate REFUSES rather than passing.
--
-- ROLLBACK: 498_schedule_meta_description_backfiller_ROLLBACK.sql
--   (deletes the row; the agent stays and stays hand-runnable)

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'meta-description-backfill';
  IF n <> 0 THEN
    RAISE EXCEPTION '498: a meta-description-backfill task already exists — refusing to double-seed';
  END IF;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '498: expected exactly 1 live meta-description-backfiller agent, found % — scheduling a task at an agent that is not there would fire into the void every hour', n;
  END IF;
END $$;

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds)
VALUES (
  'meta-description-backfill',
  'Fills pages.meta_description on the site with the most fillable blanks, 25 pages per run (bugs_open/320, SEO-004). Cannot overwrite existing copy. Goes quiet by itself when the pre-query finds nothing.',
  3600,
  'meta-description-backfiller',
  'system.agent.generic.requests',
  '{}'::jsonb,
  'meta-description-backfill',   -- its own group: see the header
  1,
  $Q$
    SELECT s.id::text AS site_id, s.domain AS domain
    FROM sites s
    JOIN pages p ON p.site_id = s.id
    WHERE p.status = 'active'
      AND COALESCE(p.meta_description, '') = ''
      AND EXISTS (
        SELECT 1 FROM page_components pc
        WHERE pc.page_id = p.id
          AND pc.rendered_html IS NOT NULL
          AND COALESCE(pc.slot_name, '') NOT IN ('header','footer','head')
      )
    GROUP BY s.id, s.domain
    ORDER BY count(*) DESC, s.domain ASC
    LIMIT 1
  $Q$,
  true,
  900
);

DO $$
DECLARE r record; sid text;
BEGIN
  SELECT * INTO r FROM scheduled_tasks WHERE name = 'meta-description-backfill';

  IF r.target_agent_type IS DISTINCT FROM 'meta-description-backfiller' THEN
    RAISE EXCEPTION '498 VERIFY: target_agent_type is %', r.target_agent_type;
  END IF;
  IF r.max_concurrent <> 1 OR r.concurrency_group IS DISTINCT FROM 'meta-description-backfill' THEN
    RAISE EXCEPTION '498 VERIFY: the task is not in its own single-slot group';
  END IF;
  IF NOT r.enabled THEN
    RAISE EXCEPTION '498 VERIFY: task seeded disabled';
  END IF;

  -- The pre-query must actually RUN and return the shape the scheduler expects.
  -- A pre_query that errors is only discovered at the next tick, in a log line,
  -- as "Pre-query failed, skipping task" — so it is checked here instead.
  EXECUTE 'SELECT site_id FROM (' || r.pre_query || ') q' INTO sid;
  IF sid IS NULL THEN
    RAISE NOTICE '498: pre-query returned no rows — nothing to backfill right now. That is a valid steady state, not a fault.';
  ELSE
    RAISE NOTICE '498: pre-query resolves, next run would target site_id %', sid;
  END IF;

  RAISE NOTICE '498 OK: meta-description-backfill scheduled hourly, own group, fill-blanks-only';
END $$;

COMMIT;
