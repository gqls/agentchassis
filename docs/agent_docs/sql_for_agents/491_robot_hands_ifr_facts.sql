-- 491_robot_hands_ifr_facts.sql
-- Registers the IFR World Robotics figures on robot-hands.com's evidence base:
-- ONE series fact (annual global industrial robot installations, five
-- observations, 2020-2024, each with its own citation) and EIGHT metric facts
-- (2024 world total, 2025 forecast, 2024 operational stock, China/Japan 2024,
-- and the three regional shares).
--
-- Substrate for the first news editorial feature (news_editorial_features
-- lane, DESIGN_2026-08-19_starting_point.md section 2). Every quote below was
-- verified against the fetched source page on 2026-08-19, in-session, before
-- registration; the series takes each year AS FIRST REPORTED in that year's
-- IFR press release, one basis throughout, with the revision-state caveat
-- disclosed in the chart footnote (mig 265's Thames discipline - the 2024
-- edition restates 2022 as 552,946 against the 553,052 first reported; we do
-- not mix the two bases).
--
-- Follows 251's supersede shape exactly: supersede FIRST as its own statement
-- (a CTE alongside the INSERT sees one snapshot and trips the partial unique
-- index), then insert the new current row with the facts array extended.
--
-- ROLLBACK: UPDATE site_specs SET is_current=false WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
--   AND aspect='evidence_base' AND is_current;
--   then UPDATE ... SET is_current=true, superseded_at=NULL on the prior row.

\set ON_ERROR_STOP on
BEGIN;

-- Guard: the current row must exist and must not already carry these ids.
DO $$
DECLARE cur jsonb;
BEGIN
  SELECT data INTO cur FROM site_specs
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;
  IF cur IS NULL THEN RAISE EXCEPTION 'no current evidence_base row for robot-hands'; END IF;
  IF cur::text LIKE '%rh-ifr-installations-series%' THEN
    RAISE EXCEPTION 'rh-ifr facts already registered - refusing double apply';
  END IF;
END $$;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs
  WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || $rhfacts$[
  {
    "id": "rh-ifr-installations-series",
    "kind": "series",
    "claim": "Annual global industrial robot installations, as first reported in each year's IFR World Robotics press release",
    "context_terms": ["installations", "installed", "robots"],
    "verified_at": "2026-08-19",
    "source": {"citation": {
      "publisher": "International Federation of Robotics",
      "title": "IFR press releases, World Robotics 2021 through 2025 editions",
      "url": "https://ifr.org/ifr-press-releases",
      "quote": "series parent; every observation below carries its own release",
      "accessed": "2026-08-19"}},
    "observations": [
      {"as_of": "2020", "value": 384000, "verified_at": "2026-08-19",
       "source": {"citation": {"publisher": "International Federation of Robotics", "title": "IFR presents World Robotics 2021 reports", "url": "https://ifr.org/ifr-press-releases/news/robot-sales-rise-again", "quote": "with 384,000 units shipped globally in 2020", "accessed": "2026-08-19", "published": "2021-10-28"}}},
      {"as_of": "2021", "value": 517385, "verified_at": "2026-08-19",
       "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics Report: \"All-Time High\" with Half a Million Robots Installed in one Year", "url": "https://ifr.org/ifr-press-releases/news/wr-report-all-time-high-with-half-a-million-robots-installed", "quote": "an all-time high of 517,385 new industrial robots installed in 2021 in factories around the world", "accessed": "2026-08-19", "published": "2022-10-13"}}},
      {"as_of": "2022", "value": 553052, "verified_at": "2026-08-19",
       "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2023 Report: Asia ahead of Europe and the Americas", "url": "https://ifr.org/ifr-press-releases/news/world-robotics-2023-report-asia-ahead-of-europe-and-the-americas", "quote": "recorded 553,052 industrial robot installations in factories around the world", "accessed": "2026-08-19", "published": "2023-09-26"}}},
      {"as_of": "2023", "value": 541302, "verified_at": "2026-08-19",
       "source": {"citation": {"publisher": "International Federation of Robotics", "title": "Record of 4 Million Robots Working in Factories Worldwide", "url": "https://ifr.org/ifr-press-releases/news/record-of-4-million-robots-working-in-factories-worldwide", "quote": "The annual installation figure of 541,302 units in 2023 is the second highest in history", "accessed": "2026-08-19", "published": "2024-09-24"}}},
      {"as_of": "2024", "value": 542000, "verified_at": "2026-08-19",
       "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years", "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years", "quote": "542,000 robots installed in 2024 - more than double the number 10 years ago", "accessed": "2026-08-19", "published": "2025-09-25"}}}
    ]
  },
  {
    "id": "rh-ifr-global-2024",
    "kind": "metric",
    "claim": "Industrial robots installed worldwide in 2024, per the IFR World Robotics 2025 press release",
    "value": 542000,
    "tolerance": "exact",
    "context_terms": ["installed", "installations", "world"],
    "writer_line": "{value} industrial robots were installed worldwide in 2024",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years", "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years", "quote": "542,000 robots installed in 2024 - more than double the number 10 years ago", "accessed": "2026-08-19", "published": "2025-09-25"}}
  },
  {
    "id": "rh-ifr-forecast-2025",
    "kind": "metric",
    "claim": "IFR forecast for global industrial robot installations in 2025",
    "value": 575000,
    "tolerance": "exact",
    "context_terms": ["forecast", "expected"],
    "writer_line": "robot installations are expected to grow to {value} units in 2025",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "The Robot Report", "title": "IFR: industrial robot deployments have doubled in 10 years", "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/", "quote": "Globally, robot installations are expected to grow by 6% to 575,000 units in 2025", "accessed": "2026-08-19"}}
  },
  {
    "id": "rh-ifr-stock-2024",
    "kind": "metric",
    "claim": "Industrial robots in operational use worldwide in 2024",
    "value": 4664000,
    "tolerance": "exact",
    "context_terms": ["operational", "in use"],
    "writer_line": "{value} industrial robots were in operational use worldwide in 2024",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "The Robot Report", "title": "IFR: industrial robot deployments have doubled in 10 years", "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/", "quote": "The total number of industrial robots in operational use worldwide was 4,664,000 units in 2024 - an increase of 9% compared to the previous year", "accessed": "2026-08-19"}}
  },
  {
    "id": "rh-ifr-china-2024",
    "kind": "metric",
    "claim": "Industrial robots installed in China in 2024",
    "value": 295000,
    "tolerance": "exact",
    "context_terms": ["china"],
    "writer_line": "China installed {value} industrial robots in 2024",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "The Robot Report", "title": "IFR: industrial robot deployments have doubled in 10 years", "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/", "quote": "295,000 industrial robots have been installed in the country, the highest annual total on record", "accessed": "2026-08-19"}}
  },
  {
    "id": "rh-ifr-japan-2024",
    "kind": "metric",
    "claim": "Industrial robots installed in Japan in 2024",
    "value": 44500,
    "tolerance": "exact",
    "context_terms": ["japan"],
    "writer_line": "Japan installed {value} industrial robots in 2024",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "The Robot Report", "title": "IFR: industrial robot deployments have doubled in 10 years", "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/", "quote": "44,500 units installed in 2024", "accessed": "2026-08-19"}}
  },
  {
    "id": "rh-ifr-asia-share-2024",
    "kind": "metric",
    "claim": "Asia's share of new industrial robot deployments in 2024",
    "value": 74,
    "unit": "percent",
    "tolerance": "exact",
    "context_terms": ["asia"],
    "writer_line": "Asia accounted for {value}% of new deployments in 2024",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years", "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years", "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas", "accessed": "2026-08-19", "published": "2025-09-25"}}
  },
  {
    "id": "rh-ifr-europe-share-2024",
    "kind": "metric",
    "claim": "Europe's share of new industrial robot deployments in 2024",
    "value": 16,
    "unit": "percent",
    "tolerance": "exact",
    "context_terms": ["europe"],
    "writer_line": "{value}% of new deployments in 2024 were in Europe",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years", "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years", "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas", "accessed": "2026-08-19", "published": "2025-09-25"}}
  },
  {
    "id": "rh-ifr-americas-share-2024",
    "kind": "metric",
    "claim": "The Americas' share of new industrial robot deployments in 2024",
    "value": 9,
    "unit": "percent",
    "tolerance": "exact",
    "context_terms": ["americas"],
    "writer_line": "{value}% of new deployments in 2024 were in the Americas",
    "verified_at": "2026-08-19",
    "source": {"citation": {"publisher": "International Federation of Robotics", "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years", "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years", "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas", "accessed": "2026-08-19", "published": "2025-09-25"}}
  }
]$rhfacts$::jsonb),
  cur.source, 'news_editorial_features-lane', 'news_editorial_features-lane', true, cur.pinned,
  'IFR World Robotics figures registered 2026-08-19 for the robot-demand editorial feature; quotes verified against fetched pages in-session'
FROM _cur cur;

-- Verify: the new current row carries the series with five observations,
-- each observation carrying its OWN source (the claims-gate rule).
DO $$
DECLARE obs jsonb; n int; missing int;
BEGIN
  SELECT f->'observations' INTO obs
    FROM site_specs, jsonb_array_elements(data->'facts') f
   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current
     AND f->>'id'='rh-ifr-installations-series';
  IF obs IS NULL THEN RAISE EXCEPTION 'series fact did not land'; END IF;
  n := jsonb_array_length(obs);
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 observations, got %', n; END IF;
  SELECT count(*) INTO missing FROM jsonb_array_elements(obs) o
   WHERE o->'source'->'citation'->>'url' IS NULL;
  IF missing > 0 THEN RAISE EXCEPTION '% observation(s) missing their own citation', missing; END IF;
END $$;

COMMIT;
