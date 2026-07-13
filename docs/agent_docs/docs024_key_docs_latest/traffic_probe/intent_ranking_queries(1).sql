-- ============================================================================
-- intent_events — ranking & inspection queries (P4)
-- Operator/report queries over the collected events. Read-only. Run with a
-- :domain or across all backend sites. lower() gives ASCII/most-Unicode
-- folding; full NFC normalisation is deferred to the collector (stdlib engine
-- can't NFC) — two byte-forms of the same accented term may rank separately
-- until then.
--
-- DEPENDENCY (events-per-1k): intent_events holds SUBMISSIONS only. The visit
-- DENOMINATOR lives in the engine's counters.json (exposed at /stats), NOT
-- here — so the rate metric needs visits collected too (see the harvest spec).
-- Until then, rank on absolute signal: volume, distinct terms, dominant
-- cluster share, recency, and inbound referer/landing-query.
-- ============================================================================

-- 1) Per-domain summary — the at-a-glance "is there demand here?" row.
--    events_per_1k uses the visit denominator from intent_site_stats (the
--    collector's /stats pull); NULL until the first stats pull for that host.
SELECT
    e.host,
    count(*)                                              AS events,
    count(*) FILTER (WHERE e.event_created_at > now() - interval '7 days')  AS events_7d,
    count(*) FILTER (WHERE e.event_created_at > now() - interval '30 days') AS events_30d,
    count(DISTINCT lower(btrim(e.value)))                 AS distinct_terms,
    count(DISTINCT NULLIF(e.ref_host, ''))                AS distinct_referers,
    count(*) FILTER (WHERE e.landing_query <> '')         AS with_landing_query,
    s.visits                                              AS visits,
    CASE WHEN COALESCE(s.visits, 0) > 0
         THEN round(1000.0 * count(*) / s.visits, 1) END  AS events_per_1k_visits,
    min(e.event_created_at)                               AS first_seen,
    max(e.event_created_at)                               AS last_seen
FROM intent_events e
LEFT JOIN intent_site_stats s ON s.host = e.host
GROUP BY e.host, s.visits
ORDER BY events_30d DESC, events DESC;

-- 2) Top terms for one domain (the actual demand signal).
--    :domain e.g. 'relojistas.com'
SELECT lower(btrim(value)) AS term, count(*) AS n,
       round(100.0 * count(*) / NULLIF(sum(count(*)) OVER (), 0), 1) AS pct_of_terms
FROM intent_events
WHERE host = :'domain' AND value <> ''
GROUP BY lower(btrim(value))
ORDER BY n DESC
LIMIT 50;

-- 3) Dominant-cluster share — feeds the graduation criterion (a clear cluster
--    ≥ ~30% of terms suggests a buildable intent). Crude single-term proxy;
--    real clustering (stemming/synonyms) is a later refinement.
SELECT host,
       max(term_n)                                        AS top_term_count,
       sum(term_n)                                        AS total,
       round(100.0 * max(term_n) / NULLIF(sum(term_n), 0), 1) AS top_term_pct
FROM (
    SELECT host, lower(btrim(value)) AS term, count(*) AS term_n
    FROM intent_events WHERE value <> ''
    GROUP BY host, lower(btrim(value))
) t
GROUP BY host
ORDER BY total DESC;

-- 4) Inbound referer breakdown for one domain (where visitors came from).
SELECT ref_host, count(*) AS n
FROM intent_events
WHERE host = :'domain' AND ref_host <> ''
GROUP BY ref_host
ORDER BY n DESC
LIMIT 25;

-- 5) Surviving inbound query params for one domain (campaign/search context).
SELECT landing_query, count(*) AS n
FROM intent_events
WHERE host = :'domain' AND landing_query <> ''
GROUP BY landing_query
ORDER BY n DESC
LIMIT 25;

-- 6) Recent raw submissions for one domain (eyeball the actual text).
SELECT event_created_at, kind, value, NULLIF(ref_host,'') AS ref_host, NULLIF(country,'') AS country
FROM intent_events
WHERE host = :'domain'
ORDER BY event_created_at DESC
LIMIT 50;
