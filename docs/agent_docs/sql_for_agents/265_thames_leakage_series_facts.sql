-- 265_thames_leakage_series_facts.sql
-- The FIRST series fact: Thames Water's leakage position at each year end of
-- the 2020-25 period. This is the substrate `claims_series.go` was built for
-- (mig 250 shipped the renderer, deliberately unused until a legitimate series
-- existed) and it is the first data that genuinely qualified: five dated
-- actuals of one measure from one document, not a comparison wearing a trend's
-- clothes — the shape 251 and 254 both explicitly declined.
--
-- SOURCE AND WHY THIS ONE. Thames Water's Annual Performance Report 2024-25
-- (published 2025-07-15 per the PDF's own metadata), page "Leakage BW04",
-- charts Actual vs Target for 20/21..24/25 as "% reduction in leakage using a
-- 3-year average from the 2019/20 baseline survey". Ofwat's contemporaneous
-- Water Company Performance Reports state the same measure per year, BUT the
-- company RESTATED prior years after 2023-24 reporting-methodology
-- improvements ("In 2023/24 we made improvements to our leakage reporting" —
-- the APR's own note). Measured: Ofwat's 2022-23 report shows -10.7 for that
-- year where the APR's restated chart shows 11.1. Mixing contemporaneous
-- Ofwat values with restated APR values would therefore mix methodologies
-- inside one series, so ALL FIVE observations come from the APR's restated
-- chart, and the restatement is disclosed in the chart footnote (mig 266).
--
-- HOW THE CHART VALUES WERE VERIFIED (the 254 discipline — a column mapping
-- corroborated, not trusted). pdftotext -layout on the APR gives, per year
-- pair (Actual left, Target right, matching the legend order):
--   20/21: 5.4 / 4.1 · 21/22: 10.4 / 10.2 · 22/23: 11.1 / 14.1
--   23/24: 12.0 / 17.4 · 24/25: 13.2 / 20.5
-- Corroborated three independent ways:
--   1. The page's own sidebar headline restates the 24/25 actual: "13.2".
--   2. Ofwat WCPR 2023-24 states Thames actual -12.0 / commitment -17.4;
--      WCPR 2024-25 states -13.2 / -20.5 and baseline 672.9 Ml/d
--      ("Thames Water 672.9 -13.2 -20.5" in the normalised extraction).
--   3. The APR's own ODI table row "Leakage * 2.971 0.400 (7.908) (14.043)
--      (19.178) (37.758)": rewards in exactly the two years the pairing says
--      Thames beat its target (5.4>4.1, 10.4>10.2), penalties in exactly the
--      three it says Thames missed. The signs pin Actual vs Target for every
--      year at once.
--
-- as_of IS THE YEAR END, "YYYY-MM" form: the water-industry reporting year
-- ends 31 March, so 2020-21 -> 2021-03. validAsOf accepts YYYY-MM and lexical
-- order equals chronological order. The chart footnote explains the ticks.
--
-- Observation notes carry the corroboration per point. They are register
-- documentation: nothing renders an observation's note (the template renders
-- only the series-level note from the chart config).

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs
  WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || $new$[
    {"id":"CIT-tw-leakage-series","kind":"series",
     "claim":"Thames Water's leakage position at each year end of the 2020-25 period, measured as a three-year rolling average reduction from its 2019-20 baseline, as restated in the company's 2024-25 Annual Performance Report",
     "context_terms":["leakage"],
     "verified_at":"2026-07-29",
     "observations":[
       {"as_of":"2021-03","value":5.4,"verified_at":"2026-07-29",
        "note":"Actual for 2020-21; ahead of the 4.1% commitment that year - the APR's ODI table shows a 2.971m reward for leakage in year 1",
        "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"% reduction in leakage using a 3-year average from the 2019/20","accessed":"2026-07-29","published":"2025-07"}}},
       {"as_of":"2022-03","value":10.4,"verified_at":"2026-07-29",
        "note":"Actual for 2021-22; marginally ahead of the 10.2% commitment - 0.400m reward in the ODI table",
        "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"% reduction in leakage using a 3-year average from the 2019/20","accessed":"2026-07-29","published":"2025-07"}}},
       {"as_of":"2023-03","value":11.1,"verified_at":"2026-07-29",
        "note":"Actual for 2022-23 AS RESTATED. Ofwat's contemporaneous WCPR 2022-23 recorded -10.7 against a -14.1 commitment, before the company's 2023-24 reporting-methodology restatement; ODI penalty 7.908m",
        "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"% reduction in leakage using a 3-year average from the 2019/20","accessed":"2026-07-29","published":"2025-07"}}},
       {"as_of":"2024-03","value":12.0,"verified_at":"2026-07-29",
        "note":"Actual for 2023-24; matches Ofwat WCPR 2023-24 (-12.0 actual, -17.4 commitment); ODI penalty 14.043m",
        "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"% reduction in leakage using a 3-year average from the 2019/20","accessed":"2026-07-29","published":"2025-07"}}},
       {"as_of":"2025-03","value":13.2,"verified_at":"2026-07-29",
        "note":"Actual for 2024-25; the BW04 page's own headline restates 13.2, Ofwat WCPR 2024-25 shows -13.2 actual against -20.5 commitment, and the ODI penalty is 19.178m",
        "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"% reduction in leakage using a 3-year average from the 2019/20","accessed":"2026-07-29","published":"2025-07"}}}
     ]},
    {"id":"CIT-tw-leakage-target-2425","kind":"metric","value":20.5,"unit":"percent",
     "claim":"Thames Water's 2024-25 leakage performance commitment - the final-year target of the 2020-25 period - was a 20.5% reduction from the 2019-20 baseline on a three-year average, as updated in the company's 2024-25 Annual Performance Report",
     "context_terms":["commitment","target","leakage"],
     "writer_line":"The final-year commitment asked for a {value}% reduction",
     "verified_at":"2026-07-29",
     "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"Now 249.3 20.5%","accessed":"2026-07-29","published":"2025-07"}}},
    {"id":"CIT-tw-leakage-penalty-2425","kind":"metric","value":19.178,"unit":"GBP_million",
     "claim":"Missing the 2024-25 leakage commitment carried a £19.178 million underperformance penalty, as stated on the leakage page of Thames Water's 2024-25 Annual Performance Report",
     "context_terms":["penalty","leakage"],
     "writer_line":"Missing the final-year level carried a £{value}m underperformance penalty",
     "verified_at":"2026-07-29",
     "source":{"citation":{"publisher":"Thames Water","title":"Annual Performance Report 2024-25","url":"https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf","quote":"Penalty: £19.178m (not met)","accessed":"2026-07-29","published":"2025-07"}}},
    {"id":"CIT-sector-leakage-2020-25","kind":"metric","value":9.3,"unit":"percent",
     "claim":"Across the water sector as a whole, leakage fell by 9.3% over the 2020-25 period on the three-year-average measure, per Ofwat's Water company performance report 2024-25",
     "context_terms":["sector"],
     "writer_line":"The sector as a whole lowered leakage by {value}% over the period",
     "verified_at":"2026-07-29",
     "source":{"citation":{"publisher":"Ofwat","title":"Water company performance report 2024-25","url":"https://www.ofwat.gov.uk/wp-content/uploads/2025/10/WCPR-24-25.pdf","quote":"the sector has lowered leakage by 9.3%","accessed":"2026-07-29","published":"2025-10"}}}
  ]$new$::jsonb),
  'oufe-workstream','session','oufe-workstream', true, COALESCE(cur.pinned,false),
  'Thames Water leakage 2020-25: the register''s FIRST series fact (5 observations, each with its own citation), plus the final-year target, its penalty, and the sector comparator. All five observations from the restated BW04 chart in the company''s APR 2024-25 - NOT from the contemporaneous Ofwat reports, which predate a methodology restatement (Ofwat 2022-23 says -10.7 where the restated chart says 11.1). Column mapping corroborated three ways: the page''s own 13.2 headline, Ofwat WCPR 23-24/24-25 cross-checks, and the ODI reward/penalty signs per year.'
FROM _cur cur;

COMMIT;

-- VERIFY: count went 32 -> 36, and the series fact carries 5 observations,
-- every one with its own citation (the never-inherit rule, checked here the
-- way the gate checks it - per observation, not per fact).
SELECT jsonb_array_length(data->'facts') AS total_facts,
       (SELECT jsonb_array_length(f->'observations')
        FROM jsonb_array_elements(data->'facts') f
        WHERE f->>'id'='CIT-tw-leakage-series') AS observations,
       (SELECT count(*)
        FROM jsonb_array_elements(data->'facts') f,
             jsonb_array_elements(f->'observations') o
        WHERE f->>'id'='CIT-tw-leakage-series'
          AND o->'source'->'citation'->>'url' IS NOT NULL
          AND o->'source'->'citation'->>'quote' IS NOT NULL) AS observations_with_own_citation
FROM site_specs
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;
