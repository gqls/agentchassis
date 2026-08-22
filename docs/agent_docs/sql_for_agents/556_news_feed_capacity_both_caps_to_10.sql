-- 556_news_feed_capacity_both_caps_to_10.sql
--
-- bugs_open/316, capacity half. OWNER DECISION 2026-08-22: "increase the
-- capacity with both caps together."
--
-- BOTH literals move, or neither does — that is the whole point of this file.
-- `content-feed-trigger` gates the same fan-out TWICE:
--     find_news_sites  query ... LIMIT 5          (how many sites are SELECTED)
--     process_sites    config.max_iterations = 5  (how many are PROCESSED)
-- Raising only the query LIMIT changes throughput by NOTHING — the query returns
-- ten rows and the loop still processes five — while the cap-hit census (which
-- measures the step's own output length) flips from "5 of 5, at the cap" to
-- "10 of 10, under the cap" and STOPS REPORTING. The instrument would show
-- relief that never happened. That trap is why this migration exists as one file
-- touching two paths rather than two convenient edits.
--
-- WHY 10, derived rather than picked:
--   * there are 9 news-feed-eligible sites with a deployed page [MEASURED
--     2026-08-22], so any cap >= 9 stops binding entirely and every DUE site is
--     served on every run;
--   * 10 leaves exactly one slot of headroom, so a tenth site can be added
--     without silently re-entering the starvation regime;
--   * above 10 is inert today. And note the cap DOES still bind at 10 if an
--     eleventh site appears — which is wanted, because that is when LCO-009's
--     row-cap WARN should fire again and ask the question a second time.
--
-- ⚠ THIS DOES NOT MAKE EVERY SITE ON TIME, AND IT CANNOT. Read this before
-- reporting the capacity problem as solved.
-- The trigger fires every 6 hours (`scheduled_tasks.content-feed-refresh`,
-- interval_seconds = 21600), so NO site can be served more than 4x/day at ANY
-- cap. Against each site's own configured cadence [MEASURED 2026-08-22]:
--     7 sites want 4/day (6h cadence)  -> fully satisfied by this change
--     dartsonline.com wants 6/day (4h) -> ceiling 4/day, CAPPED BY FREQUENCY
--     relojistas.com  wants 8/day (3h) -> ceiling 4/day, CAPPED BY FREQUENCY
-- So the residual shortfall after this migration is exactly 6 fetches/day and it
-- belongs entirely to those two sites. It is a TRIGGER-FREQUENCY decision (or a
-- cadence decision), not a cap decision, and it is deliberately NOT taken here.
-- The bug file's headline "42 demanded vs 20 supplied, 2.10x" is a pooled figure
-- that hides this: the pool framing implies a bigger cap could close the gap, and
-- per-site it cannot.
--
-- COST, MEASURED NOT GUESSED. Supply goes from 4 runs x 5 = 20 site-refreshes/day
-- to 4 x 9 = 36 (cap 10, 9 eligible), i.e. +80%. Each refresh spawns a
-- `content-feed-orchestrator`, whose LLM component is `feed-triage`: 114 calls
-- over the last 7 days, avg 2,780 in / 1,992 out, 544k tokens total = ~78k
-- tokens/day. Scaling with site-refreshes puts it at ~140k tokens/day, an
-- increase of roughly 62k tokens/day. Non-LLM work (fetch, render, git commit)
-- scales the same way.
--
-- ⚠ DOES NOT TOUCH THE ORDERING migration 554 installed. The query text below is
-- 554's, byte-for-byte, with `LIMIT 5` -> `LIMIT 10` and nothing else; the guard
-- asserts the full pre-state and the verify block asserts the due-ordering
-- survived. Losing the fairness fix while "increasing capacity" would restore the
-- starvation this lane just removed.
--
-- LIVE ON APPLY. Config, no image roll.
-- Rollback: 556_..._ROLLBACK.sql returns both literals to 5.

SELECT snapshot_agent('content-feed-trigger',
       'migration 556: pre-update (bugs_open/316 capacity — both caps 5 -> 10)');

BEGIN;

DO $$
DECLARE
    live_rows int;
    q         text;
    iters     jsonb;
    versions  text;
BEGIN
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF live_rows <> 1 THEN
        SELECT string_agg(COALESCE(version::text,'(null)'), ', ' ORDER BY version) INTO versions
        FROM agent_definitions
        WHERE type = 'content-feed-trigger' AND is_active
          AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
        RAISE EXCEPTION 'MIGRATION 556: expected exactly 1 live content-feed-trigger row, found % (versions: %).', live_rows, COALESCE(versions,'(none)');
    END IF;

    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query',
           default_config->'workflow'->'steps'->'process_sites'->'config'->'max_iterations'
      INTO q, iters
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    -- Gated on the WHOLE pre-state, not the region being changed (council corr
    -- e6e8b923, debug_historian): a concurrent edit to any other clause leaves
    -- the tail intact, so a tail-only guard would pass and this rewrite would
    -- silently revert it. This literal is migration 554's installed query.
    IF q IS DISTINCT FROM $pre$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5$pre$ THEN
        RAISE EXCEPTION 'MIGRATION 556: find_news_sites is not byte-identical to migration 554''s installed query — another change landed first. Live length %, expected %. Live tail: %.',
            length(q), length($pre$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5$pre$), right(q, 70);
    END IF;

    IF iters IS DISTINCT FROM to_jsonb(5) THEN
        RAISE EXCEPTION 'MIGRATION 556: process_sites.max_iterations is %, expected 5 — re-derive against the live value. Raising the query LIMIT without this one is a no-op that also silences the cap census.', COALESCE(iters::text,'ABSENT');
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,find_news_sites,config,query}',
            to_jsonb($new$SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10$new$::text),
            false),
        '{workflow,steps,process_sites,config,max_iterations}',
        to_jsonb(10),
        false),
    updated_at = now()
WHERE type = 'content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    q     text;
    iters jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query',
           default_config->'workflow'->'steps'->'process_sites'->'config'->'max_iterations'
      INTO q, iters
    FROM agent_definitions
    WHERE type = 'content-feed-trigger' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    -- BOTH, and the pairing is the assertion. Either alone is the defect.
    IF q NOT LIKE '%LIMIT 10' THEN
        RAISE EXCEPTION 'MIGRATION 556: the query cap did not move. Live tail: %', right(q, 70);
    END IF;
    IF iters IS DISTINCT FROM to_jsonb(10) THEN
        RAISE EXCEPTION 'MIGRATION 556: process_sites.max_iterations is %, expected 10 — the query cap moved and the loop cap did not, which is exactly the no-op this migration exists to avoid.', COALESCE(iters::text,'ABSENT');
    END IF;

    -- Migration 554's fairness fix must survive. "Increasing capacity" by
    -- reverting the ordering would restore the starvation this lane removed.
    IF q NOT LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10' THEN
        RAISE EXCEPTION 'MIGRATION 556: the due-ordering from migration 554 is gone. Live tail: %', right(q, 70);
    END IF;
    IF q LIKE '%ORDER BY s.domain%' THEN
        RAISE EXCEPTION 'MIGRATION 556: the alphabetical ordering has come back.';
    END IF;
    -- Eligibility predicate untouched, both arms.
    IF q NOT LIKE '%NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true)%'
       OR q NOT LIKE '%cs.next_fetch_at <= NOW()%' THEN
        RAISE EXCEPTION 'MIGRATION 556: an arm of the eligibility predicate was lost.';
    END IF;

    BEGIN
        EXECUTE 'EXPLAIN ' || q;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'MIGRATION 556: the new query does not plan (%).', SQLERRM;
    END;
END $$;

COMMIT;
