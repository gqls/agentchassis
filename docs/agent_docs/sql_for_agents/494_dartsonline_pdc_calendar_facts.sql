-- 494_dartsonline_pdc_calendar_facts.sql
-- Registers the PDC calendar substrate on dartsonline.com's evidence base: ONE
-- series fact (Players Championship events per season, 2022-2026, each
-- observation carrying its own citation) and FOUR metrics (2026 PC/ET/secondary
-- counts, plus the 2025 ET count so the chart can show the year-on-year move).
--
-- Substrate for the SECOND news editorial feature (news_editorial_features lane,
-- NEWS-020's pattern; rollout site 2 of the order in
-- DESIGN_2026-08-19_starting_point.md).
--
-- dartsonline's evidence_base existed with an EMPTY facts array
-- [MEASURED 2026-08-20: jsonb_array_length = 0], so these are the first
-- registered facts on the site. The append shape is identical to 491's.
--
-- SOURCE HONESTY, stated because it is weaker than 491's and must not be
-- glossed: these counts come from each season's own summary page on Wikipedia,
-- not from the PDC's own calendar listing. Every quote was fetched and verified
-- verbatim in-session on 2026-08-20, the publisher is named in every citation,
-- and the chart footnote says so on the page. The 2022/2023/2024 figures each
-- come from that season's dedicated Players Championship series page; the
-- 2025/2026 figures come from that season's Pro Tour page, whose one sentence
-- carries the PC, ET and secondary-tour counts together.
--
-- ROLLBACK: UPDATE site_specs SET is_current=false WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
--   AND aspect='evidence_base' AND is_current;  then restore the prior row.

\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE cur jsonb;
BEGIN
  SELECT data INTO cur FROM site_specs
   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='evidence_base' AND is_current;
  IF cur IS NULL THEN RAISE EXCEPTION 'no current evidence_base row for dartsonline'; END IF;
  IF cur::text LIKE '%do-pdc-pc-events-series%' THEN
    RAISE EXCEPTION 'do-pdc facts already registered - refusing double apply';
  END IF;
END $$;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs
  WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', COALESCE(cur.data->'facts','[]'::jsonb) || $dofacts$[
 {
  "id": "do-pdc-pc-events-series",
  "kind": "series",
  "claim": "Number of PDC Players Championship events per season, 2022-2026, as stated in each season's own summary",
  "context_terms": [
   "players championship",
   "events",
   "tournaments"
  ],
  "verified_at": "2026-08-20",
  "source": {
   "citation": {
    "publisher": "Wikipedia",
    "title": "PDC season summaries, 2022-2026",
    "url": "https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour",
    "quote": "series parent; every observation below carries its own page",
    "accessed": "2026-08-20"
   }
  },
  "observations": [
   {
    "as_of": "2022",
    "value": 30,
    "verified_at": "2026-08-20",
    "source": {
     "citation": {
      "publisher": "Wikipedia",
      "title": "2022 PDC Players Championship series",
      "url": "https://en.wikipedia.org/wiki/2022_PDC_Players_Championship_series",
      "quote": "The 2022 PDC Players Championship series consisted of 30 darts tournaments on the 2022 PDC Pro Tour.",
      "accessed": "2026-08-20"
     }
    }
   },
   {
    "as_of": "2023",
    "value": 30,
    "verified_at": "2026-08-20",
    "source": {
     "citation": {
      "publisher": "Wikipedia",
      "title": "2023 PDC Players Championship series",
      "url": "https://en.wikipedia.org/wiki/2023_PDC_Players_Championship_series",
      "quote": "The 2023 PDC Players Championship series consisted of 30 darts tournaments on the 2023 PDC Pro Tour.",
      "accessed": "2026-08-20"
     }
    }
   },
   {
    "as_of": "2024",
    "value": 30,
    "verified_at": "2026-08-20",
    "source": {
     "citation": {
      "publisher": "Wikipedia",
      "title": "2024 PDC Players Championship series",
      "url": "https://en.wikipedia.org/wiki/2024_PDC_Players_Championship_series",
      "quote": "The 2024 PDC Players Championship series consisted of 30 darts tournaments on the 2024 PDC Pro Tour.",
      "accessed": "2026-08-20"
     }
    }
   },
   {
    "as_of": "2025",
    "value": 34,
    "verified_at": "2026-08-20",
    "source": {
     "citation": {
      "publisher": "Wikipedia",
      "title": "2025 PDC Pro Tour",
      "url": "https://en.wikipedia.org/wiki/2025_PDC_Pro_Tour",
      "quote": "The 2025 calendar consisted of 34 Players Championship events, 14 European Tour events, as well as 24 events for each of the PDC's secondary tours",
      "accessed": "2026-08-20"
     }
    }
   },
   {
    "as_of": "2026",
    "value": 34,
    "verified_at": "2026-08-20",
    "source": {
     "citation": {
      "publisher": "Wikipedia",
      "title": "2026 PDC Pro Tour",
      "url": "https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour",
      "quote": "The 2026 calendar consists of 34 Players Championship events, 15 European Tour events, as well as 24 events for each of the PDC's secondary tours",
      "accessed": "2026-08-20"
     }
    }
   }
  ]
 },
 {
  "id": "do-pdc-pc-2026",
  "kind": "metric",
  "claim": "PDC Players Championship events in the 2026 calendar",
  "value": 34,
  "tolerance": "exact",
  "context_terms": [
   "players championship"
  ],
  "writer_line": "{value} Players Championship events",
  "verified_at": "2026-08-20",
  "source": {
   "citation": {
    "publisher": "Wikipedia",
    "title": "2026 PDC Pro Tour",
    "url": "https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour",
    "quote": "The 2026 calendar consists of 34 Players Championship events, 15 European Tour events, as well as 24 events for each of the PDC's secondary tours",
    "accessed": "2026-08-20"
   }
  }
 },
 {
  "id": "do-pdc-et-2026",
  "kind": "metric",
  "claim": "PDC European Tour events in the 2026 calendar",
  "value": 15,
  "tolerance": "exact",
  "context_terms": [
   "european tour"
  ],
  "writer_line": "{value} European Tour events",
  "verified_at": "2026-08-20",
  "source": {
   "citation": {
    "publisher": "Wikipedia",
    "title": "2026 PDC Pro Tour",
    "url": "https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour",
    "quote": "The 2026 calendar consists of 34 Players Championship events, 15 European Tour events, as well as 24 events for each of the PDC's secondary tours",
    "accessed": "2026-08-20"
   }
  }
 },
 {
  "id": "do-pdc-et-2025",
  "kind": "metric",
  "claim": "PDC European Tour events in the 2025 calendar",
  "value": 14,
  "tolerance": "exact",
  "context_terms": [
   "european tour"
  ],
  "writer_line": "{value} European Tour events in 2025",
  "verified_at": "2026-08-20",
  "source": {
   "citation": {
    "publisher": "Wikipedia",
    "title": "2025 PDC Pro Tour",
    "url": "https://en.wikipedia.org/wiki/2025_PDC_Pro_Tour",
    "quote": "The 2025 calendar consisted of 34 Players Championship events, 14 European Tour events, as well as 24 events for each of the PDC's secondary tours",
    "accessed": "2026-08-20"
   }
  }
 },
 {
  "id": "do-pdc-secondary-2026",
  "kind": "metric",
  "claim": "Events on each of the PDC's secondary tours in 2026 (Challenge Tour, Development Tour, Women's Series)",
  "value": 24,
  "tolerance": "exact",
  "context_terms": [
   "secondary",
   "challenge tour",
   "development tour",
   "women's series"
  ],
  "writer_line": "{value} events on each secondary tour",
  "verified_at": "2026-08-20",
  "source": {
   "citation": {
    "publisher": "Wikipedia",
    "title": "2026 PDC Pro Tour",
    "url": "https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour",
    "quote": "as well as 24 events for each of the PDC's secondary tours",
    "accessed": "2026-08-20"
   }
  }
 }
]$dofacts$::jsonb),
  cur.source, 'news_editorial_features-lane', 'news_editorial_features-lane', true, cur.pinned,
  'PDC calendar figures registered 2026-08-20 for the schedule-density editorial feature; quotes verified verbatim against fetched pages in-session; publisher is Wikipedia season summaries, named in every citation'
FROM _cur cur;

DO $$
DECLARE obs jsonb; n int; missing int;
BEGIN
  SELECT f->'observations' INTO obs
    FROM site_specs, jsonb_array_elements(data->'facts') f
   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='evidence_base' AND is_current
     AND f->>'id'='do-pdc-pc-events-series';
  IF obs IS NULL THEN RAISE EXCEPTION 'series fact did not land'; END IF;
  n := jsonb_array_length(obs);
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 observations, got %', n; END IF;
  SELECT count(*) INTO missing FROM jsonb_array_elements(obs) o
   WHERE o->'source'->'citation'->>'url' IS NULL;
  IF missing > 0 THEN RAISE EXCEPTION '% observation(s) missing their own citation', missing; END IF;
END $$;

COMMIT;
