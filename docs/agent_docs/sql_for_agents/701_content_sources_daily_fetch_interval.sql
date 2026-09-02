-- 701_content_sources_daily_fetch_interval.sql
--
-- OWNER DECISION 2026-09-02: news sources move from a 6-hour to a 24-hour
-- fetch interval. This is the answer to the capacity question left open by
-- bugs_closed/410 (the next_fetch_at phase lock — resolve by SLUG; 410 also
-- names an unrelated scan-loss case) and recorded in bugs_open/316.
--
-- WHY THIS SHAPE, AND NOT THE OTHER TWO. "Reduce the frequency to 24 h" could
-- mean the source interval or the TRIGGER cadence. They are not equivalent:
--
--   * THIS FILE: content_sources.fetch_interval 6h -> 24h, trigger cadence
--     UNCHANGED at 21600 s. The pass keeps checking four times a day, but each
--     site is due once a day, so ~3-4 sites contend per pass against the cap of
--     10 (migration 556's LIMIT). The cap therefore STOPS binding without being
--     touched — which is the point, the owner deferred the cap change.
--
--   * REJECTED — trigger cadence -> 86400 s: one pass a day, all 14 eligible
--     sites due at that single pass, cap 10 binds hard, and 4 sites get nothing
--     that day and wait 48 h. That is worse starvation than the bug we just
--     closed, and it would FORCE the cap change the owner deferred.
--
--   * REJECTED — both -> 24 h: same cap starvation, and it additionally
--     re-creates fetch_interval == trigger cadence, which is the EXACT
--     precondition of bugs_closed/410's phase lock. Survivable now (the
--     half-cadence look-ahead is live in both layers) but there is no reason
--     to re-arm it.
--
-- SO THE 24h != 6h INEQUALITY IS LOAD-BEARING, not incidental: it makes the
-- phase lock structurally impossible rather than merely fixed. Guard 1 asserts
-- the cadence is still 21600 s; if someone later changes the cadence to 24 h
-- this file's premise is void — re-read bugs_closed/410 §1 before proceeding.
--
-- CLASS FIX, NOT A ONE-OFF. Both INSERTs in
-- platform/orchestration/actions/seed_content_sources_action.go (:284, :355)
-- OMIT fetch_interval, so every new source inherits the COLUMN DEFAULT. This
-- file changes the default as well as the 73 existing rows; without that, every
-- news site seeded after today would silently come back at 6 hours. No Go
-- change is needed and no image roll is required — the interval lives entirely
-- in the DB (verified 2026-09-02: no Go file hardcodes 6h).
--
-- THE SUB-6h SOURCES GO TOO (owner decision, same conversation): dartsonline
-- (4h x2) and relojistas (3h x2, 4h x2) were set below the default
-- deliberately. Moving them is what makes the estate uniform and is where a
-- third of the remaining fetch volume was. NOTE FOR ANYONE RE-MEASURING 410:
-- those two sites were the CONTROL in every cadence census in that lane. After
-- this file there is no sub-cadence control site left.
--
-- WHY next_fetch_at IS RE-STAMPED (the third statement below). Changing only
-- the interval leaves every existing stamp at last_fetched_at + 6h, i.e. all 14
-- sites already due. They would then settle into whichever pass happened to
-- serve them: measured today that is 10 sites in one pass and 4 in another —
-- and 10 EXACTLY FILLS the cap, so the fifteenth news site would re-start cap
-- contention immediately. Spreading the first service across the four daily
-- passes (4/4/3/3) is what actually buys the headroom this change is for. Each
-- site then stays in its slot, because next_fetch_at is re-stamped to
-- fetch_time + 24h and the following day's same-slot trigger is ~24 h later.
--
-- EXPECTED AFTERWARDS, so nobody files any of it as a regression:
--   * Fetch volume falls from ~180 to ~73 source-fetches/day (~60%).
--   * Every news site refreshes once every 24 h instead of the ~9 h it has
--     been running since 410's fix (12 h before it).
--   * LCO-009 / --capped-schedule-ordering STOPS reporting cap hits. Migration
--     653's header predicted the opposite ("cap hits on most passes") and was
--     right for the estate as it stood; this file is why that stops. A silent
--     capped-schedule check after today is the expected result, NOT evidence
--     the check has broken — if you need to know it still works, give it a
--     demand control rather than reading the zero.

BEGIN;

-- Guard 1: the trigger cadence must still be 21600 s. Two things depend on it:
-- the 24h != cadence inequality above, and the live look-ahead being 3 h, which
-- is what the 6-hour slot spacing below assumes.
DO $$
DECLARE secs integer;
        last_fire timestamptz;
BEGIN
    SELECT interval_seconds, last_triggered_at INTO secs, last_fire
      FROM scheduled_tasks WHERE name = 'content-feed-refresh';

    IF secs IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 701: scheduled_tasks has no content-feed-refresh row — the cadence this file is sized against is gone. Do not apply blind.';
    END IF;
    IF secs IS DISTINCT FROM 21600 THEN
        RAISE EXCEPTION 'MIGRATION 701: content-feed-refresh cadence is % s, not 21600. This file assumes a 6h cadence (so that 24h != cadence, and the look-ahead is 3h). Re-derive the slot spacing and re-read bugs_closed/410 §1 before applying.', secs;
    END IF;

    -- Guard 2: the slot spread is computed from the NEXT expected pass, i.e.
    -- last_triggered_at + 6h. If the scheduler has not fired within a cadence
    -- plus slack, that base is wrong and the spread would land in the past.
    IF last_fire IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 701: content-feed-refresh has never fired (last_triggered_at IS NULL) — cannot compute the pass slots.';
    END IF;
    IF now() - last_fire > interval '7 hours' THEN
        RAISE EXCEPTION 'MIGRATION 701: content-feed-refresh last fired % ago (> one cadence + slack). The scheduler is not running; fix that first, or the slot spread below lands in the past and every site piles into one pass.', now() - last_fire;
    END IF;
END $$;

-- Guard 3: there is something to change. A zero here means another session has
-- already done this (or the table is empty) — either way, stop and look.
DO $$
DECLARE n integer;
BEGIN
    SELECT count(*) INTO n FROM content_sources
     WHERE is_active AND fetch_interval IS DISTINCT FROM interval '24 hours';
    IF n = 0 THEN
        RAISE EXCEPTION 'MIGRATION 701: no active source has a non-24h fetch_interval — already applied, or content_sources is empty. Nothing done.';
    END IF;
    RAISE NOTICE 'MIGRATION 701: % active source(s) will move to a 24h fetch_interval.', n;
END $$;

-- Backup, so the ROLLBACK sidecar can restore the ACTUAL previous values.
-- They are not uniform (3h x2, 4h x4, 6h x67 as of 2026-09-02), so a rollback
-- that just sets everything to the old default would be wrong.
CREATE TABLE bak_content_sources_fetch_interval_20260902 AS
SELECT id, site_id, name, fetch_interval, next_fetch_at, last_fetched_at, now() AS backed_up_at
  FROM content_sources;

-- 1. THE CLASS FIX: new sources inherit 24h, because both seeder INSERTs omit
--    the column. Without this line the change lasts until the next site build.
ALTER TABLE content_sources ALTER COLUMN fetch_interval SET DEFAULT '24:00:00'::interval;

-- 2. The existing rows.
UPDATE content_sources
   SET fetch_interval = interval '24 hours'
 WHERE fetch_interval IS DISTINCT FROM interval '24 hours';

-- 3. Spread the first service across the four daily passes, so the cap has real
--    headroom rather than being exactly filled. Slot k is served by the pass at
--    (next pass + k*6h): a stamp AT a trigger time is admitted by that pass
--    (look-ahead 3h) and not by the one before it (which reaches only +3h).
--    Whole sites move together — a site's sources share a slot, which is what
--    every cadence census in the 410 lane relies on.
WITH slots AS (
    SELECT s.id AS site_id,
           (row_number() OVER (ORDER BY s.domain) - 1) % 4 AS slot
      FROM sites s
     WHERE EXISTS (SELECT 1 FROM content_sources cs
                    WHERE cs.site_id = s.id AND cs.is_active)
),
base AS (
    SELECT last_triggered_at + interval '6 hours' AS next_pass
      FROM scheduled_tasks WHERE name = 'content-feed-refresh'
)
UPDATE content_sources cs
   SET next_fetch_at = base.next_pass + (slots.slot * interval '6 hours')
  FROM slots, base
 WHERE cs.site_id = slots.site_id
   AND cs.is_active;

-- VERIFY, in a DO block so a bad result ABORTS THE COMMIT. A verify section of
-- bare SELECTs cannot stop a COMMIT — ON_ERROR_STOP ignores a non-empty result
-- set — which is a trap this shop has already been bitten by.
DO $$
DECLARE bad integer; slots integer; biggest integer; dflt text; split integer;
BEGIN
    SELECT count(*) INTO bad FROM content_sources
     WHERE is_active AND fetch_interval IS DISTINCT FROM interval '24 hours';
    IF bad > 0 THEN
        RAISE EXCEPTION 'MIGRATION 701 VERIFY: % active source(s) are still not 24h.', bad;
    END IF;

    SELECT pg_get_expr(d.adbin, d.adrelid) INTO dflt
      FROM pg_attrdef d
      JOIN pg_attribute a ON a.attrelid = d.adrelid AND a.attnum = d.adnum
     WHERE d.adrelid = 'content_sources'::regclass AND a.attname = 'fetch_interval';
    IF dflt IS NULL OR dflt NOT LIKE '%24:00:00%' THEN
        RAISE EXCEPTION 'MIGRATION 701 VERIFY: column default is % — the class fix did not take, so every newly seeded source would still be 6h.', COALESCE(dflt,'(none)');
    END IF;

    -- The spread must actually spread, and no slot may reach the cap of 10.
    SELECT count(DISTINCT next_fetch_at) INTO slots FROM content_sources WHERE is_active;
    IF slots <> 4 THEN
        RAISE EXCEPTION 'MIGRATION 701 VERIFY: expected 4 distinct pass slots, found %.', slots;
    END IF;

    SELECT max(c) INTO biggest FROM (
        SELECT count(DISTINCT site_id) AS c FROM content_sources
         WHERE is_active GROUP BY next_fetch_at) x;
    IF biggest >= 10 THEN
        RAISE EXCEPTION 'MIGRATION 701 VERIFY: busiest slot holds % sites against a cap of 10 — the spread bought no headroom.', biggest;
    END IF;

    -- A site whose sources straddle two slots would be served twice a day.
    SELECT count(*) INTO split FROM (
        SELECT site_id FROM content_sources WHERE is_active
         GROUP BY site_id HAVING count(DISTINCT next_fetch_at) > 1) y;
    IF split > 0 THEN
        RAISE EXCEPTION 'MIGRATION 701 VERIFY: % site(s) have sources in more than one slot.', split;
    END IF;

    RAISE NOTICE 'MIGRATION 701 VERIFY OK: all active sources at 24h, column default 24h, 4 slots, busiest slot % sites (cap 10), no split sites.', biggest;
END $$;

COMMIT;
