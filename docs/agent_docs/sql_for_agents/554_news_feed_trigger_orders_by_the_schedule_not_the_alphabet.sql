-- 554_news_feed_trigger_orders_by_the_schedule_not_the_alphabet.sql
--
-- bugs_open/316. `content-feed-trigger.find_news_sites` ends
-- `ORDER BY s.domain LIMIT 5` over the news-feed sites whose sources are due.
-- The runs are 6-hourly and the sort is STABLE, so whenever more than five sites
-- are in contention the same five names win — every time, for ever. Who absorbs
-- the shortfall is decided by the alphabet.
--
-- MEASURED 2026-08-22 (live clients_db; queries in the lane RUNBOOK):
--   * all five retained runs returned exactly 5 of 5 — every run at the cap;
--   * `webdesign.co.uk`, alphabetical rank 9 of 9, appears in ZERO of them while
--     continuously due since 08-21 08:45Z — last fetched 08-21 02:45Z, over 31
--     hours on a 6-hour cadence, i.e. 419% of its own configured interval;
--   * it was eligible by the trigger's OWN predicate at every run it missed
--     (news_feed recommended, 128 deployed pages, 5 of 5 sources due), while the
--     control `ai-agent-orchestration.com` reads 0 sources due at the same
--     instants — so the check discriminates rather than matching everything.
-- When bugs_open/316 was filed on 08-19 that same site was 7% late. The
-- starvation compounds; it does not sit still.
--
-- THIS IS NOT A NEW CONVENTION. The platform's own Go layer already orders this
-- exact work by its due-time — dispatch_feed_sources_action.go:101 and
-- feed_actions.go:1016 both use `ORDER BY next_fetch_at ASC NULLS FIRST`. Those
-- select SOURCES within a site. The SITE-selection query lives in config rather
-- than in Go, and config is the one layer that skipped it. This applies the
-- existing idiom to the layer that missed it.
--
-- WHAT CHANGES: the ORDER BY, and the derived column it needs. THE ELIGIBILITY
-- PREDICATE IS BYTE-FOR-BYTE UNCHANGED, deliberately — the set of sites the
-- trigger considers is provably identical before and after, so the blast radius
-- is exactly "who goes first".
--
-- ⚠ THE BUG FILE'S OWN FIX CANDIDATE 1 IS WRONG AS WRITTEN, and correcting it is
-- the substance of this migration. It proposes `ORDER BY min_next_fetch_at NULLS
-- FIRST`. The eligibility predicate has two arms:
--     arm A: NOT EXISTS (any active content_sources for this site)
--     arm B: EXISTS (an active source with next_fetch_at IS NULL OR <= NOW())
-- Arm A is the PROVISIONING path: `content-feed-orchestrator` carries
-- `check_has_sources` -> `seed_content_sources`, so a newly-classified news site
-- with no sources yet is picked up here and seeded. A site matching arm A has a
-- NULL aggregate AND is permanently eligible — nothing a fetch does can advance
-- a timestamp it does not have. Under NULLS FIRST it would win EVERY run for
-- ever if seeding ever failed to yield an active source: a silent, unbounded
-- head-of-queue squatter that starves all eight other sites while burning a slot
-- on a failing seed each time.
--
-- SO: NULLS LAST, and the reasoning is a comparison of failure modes rather than
-- a preference. NULLS FIRST risks a failure that is unbounded and silent; NULLS
-- LAST risks one that is bounded and obvious (a new site visibly has no news).
-- MEASURED 2026-08-22: the state has ZERO live instances — no news-feed site
-- with a deployed page lacks an active source, including none stuck with only
-- INACTIVE ones, so nothing changes today.
-- ⚠ IT IS STILL A BEHAVIOUR CHANGE for that case: today an unprovisioned site
-- sorts alphabetically among everyone, so an early-alphabet one would be served
-- FIRST and after this will be served LAST.
-- The real defect underneath is that provisioning and refresh share one capped
-- queue. A priority tweak papers over that; it is recorded as a follow-up in the
-- lane PLAN rather than smuggled in here.
--
-- COALESCE(next_fetch_at, '-infinity') INSIDE THE min(): a source with
-- next_fetch_at IS NULL has NEVER been fetched and is therefore maximally
-- overdue, but SQL min() SKIPS NULLs — a bare min(cs.next_fetch_at) would hide
-- that source behind its siblings' timestamps and rank the site as if its
-- never-fetched source did not exist.
--
-- `domain ASC` IS ONLY A TIE-BREAK, for determinism. Among arm-B sites exact key
-- ties are effectively measure-zero (they are timestamps), so this does not
-- reintroduce the alphabet.
--
-- WHY NOT `ORDER BY random()` (what the sibling model-directory-trigger uses,
-- and what the bug file calls the cheap version of the same idea): it makes
-- starvation UNBIASED rather than ABSENT. A site can still lose several draws
-- running, and nothing about the result is reproducible when you come to check
-- whether the fix worked.
--
-- ⚠ WHAT THIS MIGRATION DELIBERATELY DOES NOT DO: change any cap. Demand is 42
-- site-fetches/day against a supply of 20 (2.10x oversubscribed; removing the
-- cap entirely still leaves 36 vs 42). That is an owner SPEND decision, not a
-- defect fix. And there are TWO caps IN SERIES — this query's LIMIT 5 and
-- `process_sites.max_iterations` 5 — so raising only one changes throughput by
-- NOTHING while the cap-hit census stops reporting a cap hit, i.e. the
-- instrument would show relief that never happened. Both literals must move
-- together or neither moves.
--
-- LIVE ON APPLY. Config, no image roll, no ordering constraint in either
-- direction. The companion detector (cmd/config-key-audit
-- --capped-schedule-ordering) is Go and waits for the next fleet roll; its
-- positive control against this pre-fix query was captured first and is at
-- docs/agent_docs/docs024_key_docs_latest/bugfix_316_news_feed_ordering/
-- CONTROL_prefix_detector_run_2026-08-22.txt.
--
-- Rollback: 554_..._ROLLBACK.sql restores the alphabetical ORDER BY verbatim.

SELECT snapshot_agent('content-feed-trigger',
       'migration 554: pre-update (bugs_open/316 — ORDER BY s.domain starves the alphabetical tail)');

BEGIN;

-- PRE-STATE GUARDS. DO/RAISE, not bare SELECTs: ON_ERROR_STOP ignores a
-- non-empty result set, so a verify block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    live_rows int;
    q         text;
BEGIN
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF live_rows <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 554: expected exactly 1 live content-feed-trigger row, found % — a second active row would make this UPDATE ambiguous. Resolve first.', live_rows;
    END IF;

    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
      INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF q IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 554: find_news_sites has no config.query — the step has been restructured. Re-derive this migration against the live definition.';
    END IF;

    -- ⚠ THIS GUARD IS THE POINT, not boilerplate. ~30 sessions share this tree
    -- and this row's updated_at moved at 08:36Z on the morning of this fix for
    -- reasons nothing in schema_migrations or either snapshot table accounts
    -- for. Gating on the exact pre-state means a concurrent edit ABORTS this
    -- migration instead of being silently overwritten under someone else's nose.
    IF q NOT LIKE '%ORDER BY s.domain LIMIT 5' THEN
        RAISE EXCEPTION 'MIGRATION 554: find_news_sites no longer ends "ORDER BY s.domain LIMIT 5" — another change landed first. Live tail is: %. Re-derive this migration against the live value.',
            right(q, 60);
    END IF;

    IF q LIKE '%due_at%' THEN
        RAISE EXCEPTION 'MIGRATION 554: the query already carries a due_at ordering — refusing to double-apply.';
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,find_news_sites,config,query}',
        to_jsonb($q$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5$q$::text),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- POST-STATE VERIFY.
DO $$
DECLARE
    q text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
      INTO q
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF q IS NULL THEN
        RAISE EXCEPTION 'MIGRATION 554: post-update query is NULL — jsonb_set wrote the wrong path.';
    END IF;
    IF q LIKE '%ORDER BY s.domain LIMIT%' THEN
        RAISE EXCEPTION 'MIGRATION 554: the alphabetical ORDER BY is still present after the update.';
    END IF;
    IF q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5' THEN
        RAISE EXCEPTION 'MIGRATION 554: the new ordering is not in place. Live tail is: %', right(q, 60);
    END IF;
    -- The eligibility predicate must be untouched: both arms still present.
    IF q NOT LIKE '%NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true)%' THEN
        RAISE EXCEPTION 'MIGRATION 554: arm A of the eligibility predicate (the provisioning path) was lost.';
    END IF;
    IF q NOT LIKE '%cs.next_fetch_at <= NOW()%' THEN
        RAISE EXCEPTION 'MIGRATION 554: arm B of the eligibility predicate (the due test) was lost.';
    END IF;
    IF q NOT LIKE '%build_status = ''deployed''%' THEN
        RAISE EXCEPTION 'MIGRATION 554: the deployed-page requirement was lost.';
    END IF;

    -- The rewritten SQL must actually PARSE AND PLAN. A migration that installs a
    -- syntactically broken query would apply cleanly and only fail 6 hours later,
    -- inside a trigger run nobody is watching. EXECUTE ... on a prepared plan is
    -- the cheapest way to make the database itself confirm it.
    BEGIN
        EXECUTE 'EXPLAIN ' || q;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'MIGRATION 554: the new query does not plan (%). Refusing to install a query that will fail at run time.', SQLERRM;
    END;
END $$;

COMMIT;
