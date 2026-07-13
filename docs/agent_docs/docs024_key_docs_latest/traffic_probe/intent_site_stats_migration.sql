-- ============================================================================
-- 0NN_create_intent_site_stats.sql — visit-count snapshot (P4, Option A part 1)
-- ============================================================================
-- The events-per-1k DENOMINATOR (visits) lives in the engine's counters.json,
-- exposed at /stats — NOT in intent_events. The collector pulls /stats each run
-- and upserts the latest cumulative snapshot here, one row per host. Ranking
-- then LEFT JOINs this for a true events-per-1k rate.
--
-- One row per host (latest cumulative snapshot; counters.json persists across
-- engine restarts, so the totals are cumulative). Add a history table later if
-- a visits-over-time trend is wanted.
-- ============================================================================

CREATE TABLE IF NOT EXISTS intent_site_stats (
    host        text PRIMARY KEY,
    site_id     uuid REFERENCES sites(id) ON DELETE CASCADE,
    visits      bigint NOT NULL DEFAULT 0,
    events      bigint NOT NULL DEFAULT 0,
    observed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_intent_site_stats_site
    ON intent_site_stats (site_id) WHERE site_id IS NOT NULL;
