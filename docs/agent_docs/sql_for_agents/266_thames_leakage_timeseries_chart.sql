-- 266_thames_leakage_timeseries_chart.sql
-- The FIRST use of `evidence-timeseries` (mig 250, built 2026-07-28 and
-- deliberately unused until a legitimate series existed). The series is mig
-- 265's CIT-tw-leakage-series: five restated year-end actuals from one
-- document, corroborated three independent ways (see 265's header).
--
-- Follows 255's shape exactly: content_data carries a copy of the registered
-- facts (identical to the register - no display keys, no divergence), the
-- rendered_html was produced by executing the LIVE component template through
-- text/template with the same semantics as RenderTemplateWithMap
-- (rerender_pages_actions.go:790 - parse, execute, strip "<no value>"), and
-- the row is locked permanent WITH rendered_html written in the same
-- statement (the 182 pattern: a locked row is preserved, never rendered, so
-- an empty one would render as nothing for ever).
--
-- Scale denominator is max_fact_id = CIT-tw-leakage-target-2425 (20.5), a
-- registered fact per the component's own rule: a bar reaching the top of the
-- plot would have exactly met the final-year commitment. Every figure in the
-- rendered text resolves to a mig-265 fact: 5.4/10.4/11.1/12.0/13.2 (series
-- observations), 20.5 (target), 19.178 (penalty), 9.3 (sector).
--
-- ONE RENDERING QUIRK, deliberate: the 2023-24 value renders "12%" not
-- "12.0%" - printf "%.10g" drops the trailing zero. The component supports a
-- per-observation "display" override in content_data, but using it would make
-- content_data diverge from the register by a key the Go Observation struct
-- does not carry (and the 104 landmine says unknown keys can be silently
-- dropped on a re-marshal). Identical copies beat a cosmetic decimal.

BEGIN;

WITH pg AS (SELECT id FROM pages WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water'),
     comp AS (SELECT id FROM content_components WHERE name='evidence-timeseries')
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, comp.id, 5, 'evidence-timeseries-leakage', $ts${
  "ComponentID": "evidence-timeseries-leakage",
  "eyebrow": "The operational side",
  "section_title": "Five years of leakage, against the promise",
  "intro": "A restructuring argues over who bears losses that already exist. Operational performance decides whether new ones keep arriving: under Ofwat's incentive regime, missing a committed service level takes money back off the company's allowed revenue — the revenue everything else, including the debt, is paid from. Leakage is one of those commitments, and the company's own five-year record on it reads like this.",
  "series": [
    {
      "fact_id": "CIT-tw-leakage-series",
      "label": "Leakage reduction from the 2019-20 baseline, three-year average",
      "note": "Each bar is the position at 31 March. A taller bar means less water lost. Bars are scaled against the final-year commitment of a 20.5% reduction — a bar reaching the top would have met the 2024-25 target. The first two years beat their annual commitments; every year after fell short.",
      "unit": "%",
      "max_fact_id": "CIT-tw-leakage-target-2425"
    }
  ],
  "footnote": "Figures are as the company restated them in its 2024-25 Annual Performance Report — it notes that 'In 2023/24 we made improvements to our leakage reporting', so earlier years can differ from what was reported at the time. Missing the 2024-25 level carried a £19.178m underperformance penalty. Over the same five years the sector as a whole lowered leakage by 9.3% on this measure.",
  "facts": [
    {
      "id": "CIT-tw-leakage-series",
      "kind": "series",
      "claim": "Thames Water's leakage position at each year end of the 2020-25 period, measured as a three-year rolling average reduction from its 2019-20 baseline, as restated in the company's 2024-25 Annual Performance Report",
      "context_terms": ["leakage"],
      "verified_at": "2026-07-29",
      "observations": [
        {
          "as_of": "2021-03",
          "value": 5.4,
          "verified_at": "2026-07-29",
          "note": "Actual for 2020-21; ahead of the 4.1% commitment that year - the APR's ODI table shows a 2.971m reward for leakage in year 1",
          "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "% reduction in leakage using a 3-year average from the 2019/20", "accessed": "2026-07-29", "published": "2025-07"}}
        },
        {
          "as_of": "2022-03",
          "value": 10.4,
          "verified_at": "2026-07-29",
          "note": "Actual for 2021-22; marginally ahead of the 10.2% commitment - 0.400m reward in the ODI table",
          "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "% reduction in leakage using a 3-year average from the 2019/20", "accessed": "2026-07-29", "published": "2025-07"}}
        },
        {
          "as_of": "2023-03",
          "value": 11.1,
          "verified_at": "2026-07-29",
          "note": "Actual for 2022-23 AS RESTATED. Ofwat's contemporaneous WCPR 2022-23 recorded -10.7 against a -14.1 commitment, before the company's 2023-24 reporting-methodology restatement; ODI penalty 7.908m",
          "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "% reduction in leakage using a 3-year average from the 2019/20", "accessed": "2026-07-29", "published": "2025-07"}}
        },
        {
          "as_of": "2024-03",
          "value": 12.0,
          "verified_at": "2026-07-29",
          "note": "Actual for 2023-24; matches Ofwat WCPR 2023-24 (-12.0 actual, -17.4 commitment); ODI penalty 14.043m",
          "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "% reduction in leakage using a 3-year average from the 2019/20", "accessed": "2026-07-29", "published": "2025-07"}}
        },
        {
          "as_of": "2025-03",
          "value": 13.2,
          "verified_at": "2026-07-29",
          "note": "Actual for 2024-25; the BW04 page's own headline restates 13.2, Ofwat WCPR 2024-25 shows -13.2 actual against -20.5 commitment, and the ODI penalty is 19.178m",
          "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "% reduction in leakage using a 3-year average from the 2019/20", "accessed": "2026-07-29", "published": "2025-07"}}
        }
      ]
    },
    {
      "id": "CIT-tw-leakage-target-2425",
      "kind": "metric",
      "value": 20.5,
      "unit": "percent",
      "claim": "Thames Water's 2024-25 leakage performance commitment - the final-year target of the 2020-25 period - was a 20.5% reduction from the 2019-20 baseline on a three-year average, as updated in the company's 2024-25 Annual Performance Report",
      "context_terms": ["commitment", "target", "leakage"],
      "writer_line": "The final-year commitment asked for a {value}% reduction",
      "verified_at": "2026-07-29",
      "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "Now 249.3 20.5%", "accessed": "2026-07-29", "published": "2025-07"}}
    },
    {
      "id": "CIT-tw-leakage-penalty-2425",
      "kind": "metric",
      "value": 19.178,
      "unit": "GBP_million",
      "claim": "Missing the 2024-25 leakage commitment carried a £19.178 million underperformance penalty, as stated on the leakage page of Thames Water's 2024-25 Annual Performance Report",
      "context_terms": ["penalty", "leakage"],
      "writer_line": "Missing the final-year level carried a £{value}m underperformance penalty",
      "verified_at": "2026-07-29",
      "source": {"citation": {"publisher": "Thames Water", "title": "Annual Performance Report 2024-25", "url": "https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf", "quote": "Penalty: £19.178m (not met)", "accessed": "2026-07-29", "published": "2025-07"}}
    },
    {
      "id": "CIT-sector-leakage-2020-25",
      "kind": "metric",
      "value": 9.3,
      "unit": "percent",
      "claim": "Across the water sector as a whole, leakage fell by 9.3% over the 2020-25 period on the three-year-average measure, per Ofwat's Water company performance report 2024-25",
      "context_terms": ["sector"],
      "writer_line": "The sector as a whole lowered leakage by {value}% over the period",
      "verified_at": "2026-07-29",
      "source": {"citation": {"publisher": "Ofwat", "title": "Water company performance report 2024-25", "url": "https://www.ofwat.gov.uk/wp-content/uploads/2025/10/WCPR-24-25.pdf", "quote": "the sector has lowered leakage by 9.3%", "accessed": "2026-07-29", "published": "2025-10"}}
    }
  ]
}$ts$::jsonb,
       $th$
<style>
  .ev-ts { padding: var(--spacing-xl, 4.5rem) var(--spacing-lg, 2rem);
           background: var(--color-background, #101820); color: var(--color-text, #e8e2d9); }
  .ev-ts__inner { max-width: 62rem; margin: 0 auto; }
  .ev-ts__eyebrow { display: block; font-size: 0.8125rem; font-weight: 600;
    letter-spacing: 0.1em; text-transform: uppercase;
    color: var(--color-accent, #c49a3c); margin: 0 0 0.5rem; }
  .ev-ts__title { font-size: clamp(1.5rem, 2.6vw, 2.1rem); font-weight: 700;
    line-height: 1.25; margin: 0 0 0.75rem; }
  .ev-ts__intro { max-width: 62ch; margin: 0 0 2.5rem; line-height: 1.7;
    color: var(--color-text-muted, #8a9bae); }

  .ev-ts__figure { margin: 0 0 2.5rem; }
  .ev-ts__caption { display: block; margin: 0 0 1rem; }
  .ev-ts__label { display: block; font-size: 1.02rem; font-weight: 650; }
  .ev-ts__note { display: block; font-size: 0.9rem; line-height: 1.6;
    color: var(--color-text-muted, #8a9bae); margin-top: 0.2rem; }

  /* The plot. Heights come from --v (value) and --m (scale max); the browser
     does the division, because the template cannot do arithmetic. */
  .ev-ts__plot { display: flex; align-items: flex-end; gap: 0.5rem;
    min-height: 11rem; padding: 0 0 0.5rem;
    border-bottom: 2px solid var(--color-accent, #c49a3c);
    overflow-x: auto; }
  .ev-ts__col { flex: 1 1 0; min-width: 2.75rem; display: flex;
    flex-direction: column; justify-content: flex-end; align-items: center; gap: 0.35rem; }
  .ev-ts__bar { width: 100%;
    height: calc(10rem * var(--v) / var(--m));
    min-height: 2px;
    background: var(--color-accent, #c49a3c); opacity: 0.85; border-radius: 2px 2px 0 0; }
  .ev-ts__val { font-size: 0.8rem; font-variant-numeric: tabular-nums;
    color: var(--color-text, #e8e2d9); white-space: nowrap; }

  .ev-ts__axis { display: flex; gap: 0.5rem; margin: 0.4rem 0 0; }
  .ev-ts__tick { flex: 1 1 0; min-width: 2.75rem; text-align: center;
    font-size: 0.72rem; letter-spacing: 0.02em;
    color: var(--color-text-muted, #8a9bae); white-space: nowrap; }

  /* Provenance travels with the chart. A plotted point with no visible source is
     the thing this component exists to make impossible. */
  .ev-ts__sources { margin: 0.9rem 0 0; padding: 0.7rem 0.95rem;
    background: rgba(127,127,127,0.12);
    border-left: 3px solid var(--color-accent, #c49a3c);
    font-size: 0.85rem; line-height: 1.65; color: var(--color-text-muted, #8a9bae); }
  .ev-ts__sources ul { margin: 0.4rem 0 0; padding-left: 1.1rem; }
  .ev-ts__sources a { color: var(--color-accent, #c49a3c); }
  .ev-ts__footnote { margin: 2rem 0 0; padding-top: 1rem;
    border-top: 1px solid var(--color-accent, #c49a3c);
    font-size: 0.9rem; line-height: 1.65; max-width: 62ch;
    color: var(--color-text-muted, #8a9bae); }

  @media (max-width: 40rem) {
    .ev-ts { padding: 3rem 1.15rem; }
    .ev-ts__plot { min-height: 9rem; }
    .ev-ts__bar { height: calc(8rem * var(--v) / var(--m)); }
  }
</style><section id="evidence-timeseries-leakage" class="ev-ts" data-component="evidence-timeseries">
  <div class="ev-ts__inner">
    <span class="ev-ts__eyebrow">The operational side</span>
    <h2 class="ev-ts__title">Five years of leakage, against the promise</h2>
    <p class="ev-ts__intro">A restructuring argues over who bears losses that already exist. Operational performance decides whether new ones keep arriving: under Ofwat's incentive regime, missing a committed service level takes money back off the company's allowed revenue — the revenue everything else, including the debt, is paid from. Leakage is one of those commitments, and the company's own five-year record on it reads like this.</p>

    
      <figure class="ev-ts__figure" data-series="CIT-tw-leakage-series">
        <figcaption class="ev-ts__caption">
          <span class="ev-ts__label">Leakage reduction from the 2019-20 baseline, three-year average</span>
          <span class="ev-ts__note">Each bar is the position at 31 March. A taller bar means less water lost. Bars are scaled against the final-year commitment of a 20.5% reduction — a bar reaching the top would have met the 2024-25 target. The first two years beat their annual commitments; every year after fell short.</span>
        </figcaption>

        <div class="ev-ts__plot">
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">5.4%</span>
            <span class="ev-ts__bar" style="--v:5.4000;--m:20.5000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">10.4%</span>
            <span class="ev-ts__bar" style="--v:10.4000;--m:20.5000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">11.1%</span>
            <span class="ev-ts__bar" style="--v:11.1000;--m:20.5000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">12%</span>
            <span class="ev-ts__bar" style="--v:12.0000;--m:20.5000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">13.2%</span>
            <span class="ev-ts__bar" style="--v:13.2000;--m:20.5000" aria-hidden="true"></span>
          </div>
          
        </div>
        <div class="ev-ts__axis">
          <span class="ev-ts__tick">2021-03</span><span class="ev-ts__tick">2022-03</span><span class="ev-ts__tick">2023-03</span><span class="ev-ts__tick">2024-03</span><span class="ev-ts__tick">2025-03</span>
        </div>

        <div class="ev-ts__sources">
          Every point above is a separately sourced observation. Each carries the date the
          figure applies to, and where we read it:
          <ul>
            
            <li>2021-03 —
              <a href="https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf" rel="noopener noreferrer">Annual Performance Report 2024-25</a>, read 2026-07-29
              (last checked 2026-07-29)
            </li>
            
            <li>2022-03 —
              <a href="https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf" rel="noopener noreferrer">Annual Performance Report 2024-25</a>, read 2026-07-29
              (last checked 2026-07-29)
            </li>
            
            <li>2023-03 —
              <a href="https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf" rel="noopener noreferrer">Annual Performance Report 2024-25</a>, read 2026-07-29
              (last checked 2026-07-29)
            </li>
            
            <li>2024-03 —
              <a href="https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf" rel="noopener noreferrer">Annual Performance Report 2024-25</a>, read 2026-07-29
              (last checked 2026-07-29)
            </li>
            
            <li>2025-03 —
              <a href="https://www.thameswater.co.uk/media-library/oxpbdjgk/thames-water-annual-performance-report-2024-25.pdf" rel="noopener noreferrer">Annual Performance Report 2024-25</a>, read 2026-07-29
              (last checked 2026-07-29)
            </li>
            
          </ul>
        </div>
      </figure>
      
    

    <p class="ev-ts__footnote">Figures are as the company restated them in its 2024-25 Annual Performance Report — it notes that 'In 2023/24 we made improvements to our leakage reporting', so earlier years can differ from what was reported at the time. Missing the 2024-25 level carried a £19.178m underperformance penalty. Over the same five years the sector as a whole lowered leakage by 9.3% on this measure.</p>
  </div>
</section>

$th$, 'deployed', now(), 'oufe-workstream', 'permanent'
FROM pg, comp
WHERE NOT EXISTS (SELECT 1 FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water'
    AND pc.slot_name='evidence-timeseries-leakage');

UPDATE pages SET sections=(SELECT jsonb_agg(DISTINCT x)
  FROM jsonb_array_elements(COALESCE(sections,'[]'::jsonb) || '["evidence-timeseries-leakage"]'::jsonb) x),
  build_status='needs_rebuild'
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water';

COMMIT;

-- VERIFY: five slots, the new one locked permanent with non-empty HTML.
SELECT pc.position, pc.slot_name, pc.lock_type, length(pc.rendered_html) AS len
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water'
ORDER BY pc.position;
