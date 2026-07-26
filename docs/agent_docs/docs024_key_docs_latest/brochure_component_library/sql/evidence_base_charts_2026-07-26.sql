\set ON_ERROR_STOP on
-- fundamentallyai.com evidence_base — add the chart layer (owner green-light 2026-07-26).
--
-- What this does, in order (landmine L10: SUPERSEDE BEFORE INSERT, because
-- idx_site_specs_current is UNIQUE (site_id, aspect) WHERE is_current):
--   1. snapshot the current row into bak_site_specs_fai_evidence_20260726
--   2. correct F3/F3b's verified_at (see CORRECTION below)
--   3. add four new facts, each live-verified today with its query recorded
--   4. add the `charts` key the evidence-chart component reads
--   5. supersede the old row, insert the new one
--
-- CORRECTION (2026-07-26): F3 and F3b carried verified_at 2026-07-16 and a
-- source of "traffic_probe/ (measured 2026-07-16)". That date is impossible —
-- the relojistas feed cutover was on the 17th and the first full day of data was
-- the 18th; the measurement was taken and written up on the 19th
-- (traffic_probe/README_where_we_are.md, "## 2026-07-19 — the reactivation
-- number is in"). Corrected here rather than left, because the chart renders the
-- verified date on the page, and a date that cannot be true is precisely the
-- kind of claim this site exists to not make.
--
-- DESIGN RULES enforced by the VERIFY script alongside this file:
--   * A charted fact's `value` must be a JSON number (html/template's CSS filter
--     rejects a string under printf %.4f, which would silently kill the bar).
--   * A charted fact must NOT be `tolerance: gte` — F1/F2 say "state a FLOOR,
--     never the exact number", and a bar labelled with the exact value breaks
--     the fact's own rule.
--   * A SQL-sourced fact must NOT carry `display`: refresh_evidence_base
--     rewrites `value` and `verified_at` but never `display`, so a hand-written
--     display would silently drift away from the bar it sits next to.
--   * Chart definitions carry NO business figures — only ids, labels, prose and
--     a scale constant. Denominators come from `max_fact_id` wherever the
--     denominator is itself a measured quantity.
--   * Every point label must contain one of its fact's `context_terms`: the
--     claims gate reads a 70-character window around a number and needs a
--     context term inside it, or it reports our own charted figure as an
--     unregistered number.
--
-- New facts have NO `writer_line`, deliberately: refresh_evidence_base omits
-- such facts from the regenerated writer_block, so these figures stay owned by
-- the charts and never leak into hand-written prose that would then go stale.

\set site_id '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'

BEGIN;

DROP TABLE IF EXISTS bak_site_specs_fai_evidence_20260726;
CREATE TABLE bak_site_specs_fai_evidence_20260726 AS
SELECT * FROM site_specs
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;

-- Refuse to run against anything but exactly one current row.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_site_specs_fai_evidence_20260726;
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 current evidence_base row, found %', n;
  END IF;
END $$;

CREATE TEMP TABLE new_evidence AS
WITH cur AS (
  SELECT * FROM bak_site_specs_fai_evidence_20260726
),
-- 2. Corrected existing facts.
corrected AS (
  SELECT jsonb_agg(
           CASE WHEN f->>'id' IN ('F3-relojistas-feed', 'F3b-relojistas-baseline')
                THEN f
                     || jsonb_build_object('verified_at', '2026-07-19')
                     || jsonb_build_object('source', jsonb_build_object(
                          'artifact',
                          'docs/agent_docs/docs024_key_docs_latest/traffic_probe/README_where_we_are.md (measured and written up 2026-07-19; cutover 17 July, first full day 18 July)'))
                ELSE f
           END
           ORDER BY ord) AS facts
    FROM cur, LATERAL jsonb_array_elements(cur.data->'facts') WITH ORDINALITY AS t(f, ord)
),
-- 3. New facts. Values are the live results of the queries recorded in each
--    fact's source.sql, run against the cluster on 2026-07-26.
new_facts AS (
  SELECT $facts$[
    {
      "id": "F3c-relojistas-baseline-success",
      "kind": "metric",
      "value": 0,
      "tolerance": "exact",
      "claim": "before the relaunch, 0% of requests to relojistas.com's legacy feed URL succeeded — the URL returned an error to every request for as long as the logs go back (9-17 July 2026). The complement of F3b, recorded separately because a chart may only draw a figure that exists as a fact.",
      "source": {"artifact": "docs/agent_docs/docs024_key_docs_latest/traffic_probe/README_where_we_are.md (2026-07-19: 'it was between fifty and ninety requests a day and a 404 for every single one of them')"},
      "verified_at": "2026-07-19",
      "context_terms": ["feed", "relojistas", "before", "baseline", "relaunch"]
    },
    {
      "id": "F9-feed-items-collected",
      "kind": "metric",
      "value": 7321,
      "tolerance": "exact",
      "claim": "news and content items collected by the platform's own feed pipeline (all sites, all time)",
      "source": {"sql": "SELECT count(*) FROM content_feed_items"},
      "verified_at": "2026-07-26",
      "context_terms": ["feed", "items", "collected", "pipeline", "news"]
    },
    {
      "id": "F10-feed-items-scored",
      "kind": "metric",
      "value": 6166,
      "tolerance": "exact",
      "claim": "of those collected items, the number carrying a credibility assessment (content_feed_items.credibility IS NOT NULL)",
      "source": {"sql": "SELECT count(*) FROM content_feed_items WHERE credibility IS NOT NULL"},
      "verified_at": "2026-07-26",
      "context_terms": ["feed", "items", "credibility", "assessed", "scored"]
    },
    {
      "id": "F11-council-rounds-revise",
      "kind": "metric",
      "value": 108,
      "tolerance": "exact",
      "claim": "review rounds at the platform's own reviewer council that ended REVISE — the change was sent back. Rounds, not changes: one change is often reviewed more than once.",
      "source": {"sql": "SELECT count(*) FROM doc_notes WHERE categories ? 'council-gate' AND substring(body from 'COUNCIL GATE . ([A-Z]+)') = 'REVISE'"},
      "verified_at": "2026-07-26",
      "context_terms": ["council", "review", "revision", "revise", "sent back"]
    },
    {
      "id": "F12-council-rounds-approved",
      "kind": "metric",
      "value": 37,
      "tolerance": "exact",
      "claim": "review rounds at the platform's own reviewer council that ended APPROVED",
      "source": {"sql": "SELECT count(*) FROM doc_notes WHERE categories ? 'council-gate' AND substring(body from 'COUNCIL GATE . ([A-Z]+)') = 'APPROVED'"},
      "verified_at": "2026-07-26",
      "context_terms": ["council", "review", "approved"]
    },
    {
      "id": "F13-council-rounds-rejected",
      "kind": "metric",
      "value": 9,
      "tolerance": "exact",
      "claim": "review rounds at the platform's own reviewer council that ended REJECTED — a guardian seat vetoed the change outright",
      "source": {"sql": "SELECT count(*) FROM doc_notes WHERE categories ? 'council-gate' AND substring(body from 'COUNCIL GATE . ([A-Z]+)') = 'REJECTED'"},
      "verified_at": "2026-07-26",
      "context_terms": ["council", "review", "rejected", "veto"]
    }
  ]$facts$::jsonb AS facts
),
-- 4. The chart layer. No business figure appears anywhere in here.
charts AS (
  SELECT $charts$[
    {
      "id": "relojistas-feed-restoration",
      "title": "A dormant site's legacy feed, before and after relaunch",
      "caption": "Share of requests to the old feed URL that returned real content instead of an error. Most of this traffic is search-engine crawlers, and we say so rather than counting them as readers.",
      "unit": "%",
      "max": 100,
      "pages": ["index"],
      "source_note": "Read from the site's own request logs. Both figures are rows in this site's evidence register, and the date beside each is when it was last verified.",
      "points": [
        {"fact_id": "F3c-relojistas-baseline-success", "label": "Before the relaunch, feed requests succeeding", "tone": "muted"},
        {"fact_id": "F3-relojistas-feed", "label": "First full day after the relaunch, feed requests succeeding", "tone": "primary"}
      ]
    },
    {
      "id": "news-pipeline-credibility",
      "title": "What the news pipeline collects, and what carries a credibility assessment",
      "caption": "Collected items against those carrying a credibility assessment. Both counts move as the pipeline runs, and neither is a claim about why the difference exists.",
      "unit": "",
      "max_fact_id": "F9-feed-items-collected",
      "pages": ["capabilities"],
      "source_note": "Counted live from the platform's own database. These two figures move as the pipeline runs, and they are re-verified in place rather than retyped.",
      "points": [
        {"fact_id": "F9-feed-items-collected", "label": "Feed items collected", "tone": "muted"},
        {"fact_id": "F10-feed-items-scored", "label": "Feed items with a credibility assessment", "tone": "primary"}
      ]
    },
    {
      "id": "council-review-outcomes",
      "title": "What our own review council actually decides",
      "caption": "Every change to the platform goes to an independent reviewer council before it ships. These are individual review rounds, so one change may appear more than once — and most of them are sent back.",
      "unit": "",
      "max_fact_id": "F11-council-rounds-revise",
      "pages": ["capabilities"],
      "source_note": "Counted from the council's own decision notes. We publish the outcome mix, not the review text.",
      "points": [
        {"fact_id": "F11-council-rounds-revise", "label": "Council review rounds sent back for revision", "tone": "primary"},
        {"fact_id": "F12-council-rounds-approved", "label": "Council review rounds approved", "tone": "accent"},
        {"fact_id": "F13-council-rounds-rejected", "label": "Council review rounds rejected outright by a guardian veto", "tone": "muted"}
      ]
    }
  ]$charts$::jsonb AS charts
)
SELECT cur.site_id,
       cur.aspect,
       cur.data
         || jsonb_build_object('facts', corrected.facts || new_facts.facts)
         || jsonb_build_object('charts', charts.charts)
         || jsonb_build_object('schema_notes',
              (cur.data->>'schema_notes') ||
              ' | 2026-07-26 (brochure_component_library): added the charts layer for the evidence-chart component, plus F3c/F9-F13. F3/F3b verified_at corrected 2026-07-16 -> 2026-07-19 (the earlier date pre-dates the cutover it measures). Chart-only facts carry no writer_line, so they stay out of the regenerated writer_block.')
         AS data
  FROM cur, corrected, new_facts, charts;

-- 5. Supersede, then insert (L10 — this order is load-bearing).
UPDATE site_specs
   SET is_current = false, superseded_at = now()
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT site_id, aspect, data, 'manual', 'brochure_component_library', true, 'brochure_component_library',
       'chart layer + F3c/F9-F13; F3/F3b verified_at corrected (2026-07-26)'
  FROM new_evidence;

COMMIT;

\echo '--- charts installed ---'
SELECT jsonb_array_length(data->'charts') AS charts,
       jsonb_array_length(data->'facts')  AS facts
  FROM site_specs
 WHERE site_id = :'site_id'::uuid AND aspect = 'evidence_base' AND is_current;
