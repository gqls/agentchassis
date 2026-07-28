-- SQL_p15_news_recency_window.sql — webdesign.co.uk
--
-- Set the news recency window now that bugs_closed/127's chassis half is live,
-- and re-arm so the change is exercised on the very next tick rather than in six
-- hours' time.
--
-- WHY NOW. 127 is CLOSED and live on BOTH sides — adapter and chassis are both
-- v1.0.1192, and the chassis binary carries the plumbing (pod-grep `time_range`
-- on agent-chassis-f757fcf65-bg9t7 → 1). Until this afternoon `time_range` could
-- not reach a provider at all; it can now.
--
-- WHY IT MATTERS, measured rather than assumed. `time_range` is OPTIONAL —
-- `FetchNewsSearchAction:166` only forwards it when the source config carries a
-- non-empty value, and the comment there is explicit that "absent means the news
-- vertical's own recency ranking applies". Absent is what we have had, and the
-- 13:55 fetch showed what that means in practice: the provider returned items
-- dated across a DECADE into a news feed —
--
--     2026-07-21  Meta Open-Sources Astryx …            (kept)
--     2026-07-07  The best new typefaces for July 2026  (kept)
--     2025-12-16  25 Best Sans Serif Fonts
--     2020-03-28  Open Source Fonts Are Love Letters …
--     2016-10-06  An open source font system for everyone   <- ten years old
--
-- 50 results came back and `WriteFeedItemsAction` wrote **4**, because its >30-day
-- age filter discarded 46. The feed worked, but it threw away 92% of what we paid
-- to fetch, after the fact.
--
-- WHY 'month' SPECIFICALLY, and not a tighter window. It matches
-- `WriteFeedItemsAction`'s own >30-day write filter exactly. That alignment is the
-- whole point: the provider now filters UPSTREAM to the same boundary the writer
-- enforces DOWNSTREAM, so nothing fetched is discarded for age and the two stages
-- stop disagreeing. A tighter window ('week', 'day') would start throwing away
-- items the writer would happily have kept — trading the current waste for a
-- different one — and this feed only ticks twice a day in practice, so a week is
-- already narrow. 'month' is the value that makes the two filters one decision.
--
-- WHY THE RE-ARM IS PART OF THE SAME FILE. `next_fetch_at` currently sits at
-- 19:54:50–19:55:09 while `content-feed-refresh` is due at **19:50:27** — the
-- sources would miss the imminent tick by ~4 minutes and defer to 01:50. That gap
-- is structural, not bad luck: `UpdateSourceTimestamps` stamps `NOW() + interval`
-- at INGESTION, minutes after the trigger that caused it, while the next tick is
-- `last_triggered_at + interval`. So a fetched source always comes due just AFTER
-- the following tick (measured elsewhere at 37 seconds). Clearing the timestamps
-- makes them due immediately, which is the only way to see this change today.
--
-- NOTHING ELSE CHANGES. Queries, names and `vertical_keywords` are untouched, so
-- the name↔keyword lockstep that stops `seed_content_sources` inventing a sixth
-- source still holds. This file only adds one key to each config.

\set ON_ERROR_STOP on

BEGIN;

UPDATE content_sources cs
   SET config = COALESCE(cs.config, '{}'::jsonb) || jsonb_build_object('time_range', 'month'),
       last_fetched_at = NULL,
       next_fetch_at   = NULL,
       error_count     = 0,
       last_error      = NULL,
       updated_at      = NOW()
  FROM sites s
 WHERE cs.site_id = s.id
   AND s.domain = 'webdesign.co.uk'
   AND cs.source_type = 'news_search';

DO $verify$
DECLARE v_n int; v_bad int; v_keys int; v_names int;
BEGIN
    SELECT count(*) FILTER (WHERE cs.config->>'time_range' = 'month'),
           count(*) FILTER (WHERE cs.next_fetch_at IS NOT NULL)
      INTO v_n, v_bad
      FROM content_sources cs JOIN sites s ON s.id = cs.site_id
     WHERE s.domain = 'webdesign.co.uk';

    IF v_n <> 5 THEN RAISE EXCEPTION 'expected 5 sources with time_range=month, got %', v_n; END IF;
    IF v_bad <> 0 THEN RAISE EXCEPTION '% source(s) still carry a next_fetch_at and will miss the tick', v_bad; END IF;

    -- the query must NOT have been disturbed: every source still has one
    IF EXISTS (SELECT 1 FROM content_sources cs JOIN sites s ON s.id = cs.site_id
                WHERE s.domain = 'webdesign.co.uk'
                  AND COALESCE(cs.config->>'query', '') = '') THEN
        RAISE EXCEPTION 'a source lost its query — the || merge overwrote instead of adding';
    END IF;

    -- name <-> vertical_keyword lockstep must still hold (drift summons a 6th source)
    SELECT count(*) INTO v_names FROM content_sources cs JOIN sites s ON s.id = cs.site_id
     WHERE s.domain = 'webdesign.co.uk';
    -- `jsonb_array_elements_text(...) kw` aliases the TABLE, not the value —
    -- the value column must be named explicitly (`AS kw(keyword)`). Getting this
    -- wrong failed this block's first run; the transaction rolled back and
    -- nothing was applied, which is the block working.
    SELECT count(*) INTO v_keys
      FROM site_specs sp
      JOIN sites s ON s.id = sp.site_id
      CROSS JOIN LATERAL jsonb_array_elements_text(
             sp.data->'content_features'->'news_feed'->'vertical_keywords') AS kw(keyword)
     WHERE s.domain = 'webdesign.co.uk' AND sp.aspect = 'classification' AND sp.is_current
       AND ('News Search: ' || kw.keyword) IN (
             SELECT c2.name FROM content_sources c2 JOIN sites s2 ON s2.id = c2.site_id
              WHERE s2.domain = 'webdesign.co.uk');
    IF v_keys <> v_names THEN
        RAISE EXCEPTION 'name/keyword lockstep broken: % names, % matching keywords', v_names, v_keys;
    END IF;

    RAISE NOTICE 'all 5 sources: time_range=month, re-armed, queries intact, lockstep holds.';
    RAISE NOTICE 'NOT VERIFIED YET: read source_published_at AND the count after the next tick. Expect MORE items than the 4 of 13:55, not fewer — the provider now filters to the same boundary the writer enforces.';
END
$verify$;

COMMIT;
