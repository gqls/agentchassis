BEGIN;
-- Second evidence-chart on the Thames page: the regulatory negotiation.
-- Carries the same three readability corrections already proven on the sibling
-- chart (247/253) - eyebrow and bar fills must not use --color-primary, which on
-- this site equals the surface colour. Applied at INSERT rather than discovered
-- after shipping, which is the whole point of having found them once.
WITH pg AS (SELECT id FROM pages WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water'),
     comp AS (SELECT id FROM content_components WHERE name='evidence-chart')
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, comp.id, 4, 'evidence-chart-ofwat', $ofd${
 "section_eyebrow": "The regulated side",
 "section_title": "What the company asked for, and what the regulator allowed",
 "section_intro": "A restructuring plan argues over the debt. The revenue that services it is decided somewhere else entirely, by the regulator, in a process with its own negotiation. These are Ofwat's own figures for where the average household bill was, what Thames Water asked for, and where the final determination landed.",
 "facts": [
  {
   "id": "CIT-ofwat-bill-2024-25",
   "kind": "metric",
   "unit": "GBP",
   "claim": "Ofwat's final determination shows Thames Water's average household bill at \u00a3436 in 2024-25, before inflation, as the baseline in Figure 1.1",
   "value": 436,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "The average household bill stood at \u00a3{value} in 2024-25, before inflation",
   "context_terms": [
    "average household bill",
    "baseline",
    "thames"
   ]
  },
  {
   "id": "CIT-ofwat-bill-draft",
   "kind": "metric",
   "unit": "GBP",
   "claim": "Ofwat's draft decision would have set Thames Water's average household bill at \u00a3535 by 2029-30, before inflation, per Figure 1.1 of the final determination overview",
   "value": 535,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Ofwat's draft decision would have allowed \u00a3{value} by 2029-30",
   "context_terms": [
    "draft decision",
    "average household bill",
    "thames"
   ]
  },
  {
   "id": "CIT-ofwat-bill-proposal",
   "kind": "metric",
   "unit": "GBP",
   "claim": "Thames Water proposed an average household bill of \u00a3667 by 2029-30, before inflation, per Figure 1.1 of Ofwat's final determination overview",
   "value": 667,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "average bills will be lower than those proposed by Thames Water in response to our draft decision",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Thames Water asked for \u00a3{value} by 2029-30",
   "context_terms": [
    "proposal",
    "proposed",
    "average household bill",
    "thames"
   ]
  },
  {
   "id": "CIT-ofwat-bill-final",
   "kind": "metric",
   "unit": "GBP",
   "claim": "Ofwat's final decision sets Thames Water's average household bill at \u00a3588 by 2029-30, before inflation, per Figure 1.1 of the final determination overview",
   "value": 588,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Ofwat's final decision allows \u00a3{value} by 2029-30",
   "context_terms": [
    "final decision",
    "final determination",
    "average household bill",
    "thames"
   ]
  },
  {
   "id": "CIT-ofwat-bill-increase",
   "kind": "metric",
   "unit": "GBP",
   "claim": "Ofwat's final determination increases Thames Water's average household bill by \u00a3152 between 2024-25 and 2029-30, before inflation",
   "value": 152,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "this will increase average household bills by \u00a3152 from 2024-2025 to 2029-30 for Thames Water customers, before inflation",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "The determination raises the average household bill by \u00a3{value} across the period, before inflation",
   "context_terms": [
    "increase",
    "rise",
    "average household bill",
    "thames"
   ]
  },
  {
   "id": "CIT-ofwat-revenue-2025-30",
   "kind": "metric",
   "unit": "GBP_billion",
   "claim": "Ofwat's final determination allows Thames Water to collect \u00a316.4 billion through bills over the 2025-30 period",
   "value": 16.4,
   "source": {
    "citation": {
     "url": "https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf",
     "quote": "allows Thames Water to collect \u00a316.4 billion through bills",
     "title": "Overview of Thames Water's PR24 final determination",
     "accessed": "2026-07-28",
     "published": "2024-12",
     "publisher": "Ofwat"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "The determination allows \u00a3{value} billion to be collected through bills over 2025-30",
   "context_terms": [
    "collect",
    "through bills",
    "revenue",
    "thames"
   ]
  }
 ],
 "charts": [
  {
   "id": "ofwat-pr24-bills",
   "title": "Average household bill for Thames Water, before inflation",
   "caption": "The 2024-25 figure is the starting point; the other three are all the same year, 2029-30, under different decisions. Scaled against the company's own proposal, which is the largest registered figure here.",
   "unit": "",
   "max_fact_id": "CIT-ofwat-bill-proposal",
   "points": [
    {
     "fact_id": "CIT-ofwat-bill-2024-25",
     "label": "2024-25, where it started",
     "tone": "muted"
    },
    {
     "fact_id": "CIT-ofwat-bill-draft",
     "label": "2029-30, Ofwat's draft decision"
    },
    {
     "fact_id": "CIT-ofwat-bill-proposal",
     "label": "2029-30, what Thames Water asked for"
    },
    {
     "fact_id": "CIT-ofwat-bill-final",
     "label": "2029-30, Ofwat's final decision",
     "tone": "accent"
    }
   ],
   "source_note": "From Ofwat's final determination overview, published December 2024 and read on 28 July 2026. Three of these four bars describe the same year under different decisions, so this is a comparison of positions rather than a trend over time."
  }
 ]
}$ofd$::jsonb,
       $ofh$<style>
   
  .evidence-chart {
    padding: var(--spacing-xl, 4.5rem) var(--spacing-lg, 2rem);
    background: var(--color-background, #ffffff);
    color: var(--color-text, #1a1a1a);
  }
  .evidence-chart__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .evidence-chart__header {
    max-width: 62ch;
    margin: 0 0 2.75rem;
  }
  .evidence-chart__eyebrow {
    display: block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-accent, #c49a3c);
  }
  .evidence-chart__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem);
    font-weight: 700;
    line-height: 1.25;
    margin: 0.5rem 0 0;
  }
  .evidence-chart__intro {
    margin: 0.75rem 0 0;
    line-height: 1.6;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2.5rem;
  }
  .evidence-chart__figure {
    margin: 0;
    padding: 1.5rem 1.5rem 1.25rem;
    background: var(--color-card-bg, rgba(127, 127, 127, 0.08));
    border: 1px solid var(--color-border, rgba(127, 127, 127, 0.28));
    border-radius: var(--border-radius, 10px);
  }
  .evidence-chart__figcaption {
    display: block;
    margin: 0 0 1.25rem;
  }
  .evidence-chart__chart-title {
    display: block;
    font-size: 1.0625rem;
    font-weight: 700;
    line-height: 1.35;
  }
  .evidence-chart__chart-note {
    display: block;
    margin-top: 0.35rem;
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  .evidence-chart__row {
    display: grid;
    grid-template-columns: minmax(9ch, 30%) 1fr auto;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 0;
  }
  .evidence-chart__row + .evidence-chart__row {
    border-top: 1px solid var(--color-border, rgba(127, 127, 127, 0.22));
  }
  .evidence-chart__label {
    font-size: 0.9375rem;
    line-height: 1.35;
  }
  .evidence-chart__track {
    display: block;
    height: 1.35rem;
    border-radius: 3px;
    background: rgba(127, 127, 127, 0.22);
    overflow: hidden;
  }
   
  .evidence-chart__bar {
    display: block;
    height: 100%;
    width: calc(100% * var(--v, 0) / var(--m, 1));
    min-width: 2px;
    border-radius: 3px;
    background: var(--color-text-muted, #8a9bae);
  }
  .evidence-chart__bar--muted {
    background: #6B7F96;
  }
  .evidence-chart__bar--accent {
    background: var(--color-accent, var(--color-secondary, #0f766e));
  }
  .evidence-chart__value {
    font-size: 1.0625rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }
  .evidence-chart__verified {
    grid-column: 1 / -1;
    margin: 0;
    font-size: 0.75rem;
    color: var(--color-text-muted, #666);
  }
  .evidence-chart__source {
    margin: 1.1rem 0 0;
    padding-top: 0.85rem;
    border-top: 1px solid var(--color-border, rgba(127, 127, 127, 0.28));
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--color-text-muted, #555);
  }
  @media (max-width: 620px) {
    .evidence-chart { padding: 3.25rem 1.25rem; }
    .evidence-chart__row { grid-template-columns: 1fr auto; }
    .evidence-chart__track { grid-column: 1 / -1; order: 3; }
  }
</style><section class="evidence-chart" data-component="evidence-chart">
  <div class="evidence-chart__inner">
    
    <header class="evidence-chart__header">
      <span class="evidence-chart__eyebrow">The regulated side</span>
      <h2 class="evidence-chart__title">What the company asked for, and what the regulator allowed</h2>
      <p class="evidence-chart__intro">A restructuring plan argues over the debt. The revenue that services it is decided somewhere else entirely, by the regulator, in a process with its own negotiation. These are Ofwat&#39;s own figures for where the average household bill was, what Thames Water asked for, and where the final determination landed.</p>
    </header>
    
    <div class="evidence-chart__grid">
      <figure class="evidence-chart__figure" data-chart="ofwat-pr24-bills">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Average household bill for Thames Water, before inflation</span>
          <span class="evidence-chart__chart-note">The 2024-25 figure is the starting point; the other three are all the same year, 2029-30, under different decisions. Scaled against the company&#39;s own proposal, which is the largest registered figure here.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">2024-25, where it started</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--muted" style="--v:436.0000;--m:667.0000"></span></span>
          <span class="evidence-chart__value">436</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">2029-30, Ofwat&#39;s draft decision</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:535.0000;--m:667.0000"></span></span>
          <span class="evidence-chart__value">535</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">2029-30, what Thames Water asked for</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:667.0000;--m:667.0000"></span></span>
          <span class="evidence-chart__value">667</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">2029-30, Ofwat&#39;s final decision</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:588.0000;--m:667.0000"></span></span>
          <span class="evidence-chart__value">588</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <p class="evidence-chart__source">From Ofwat&#39;s final determination overview, published December 2024 and read on 28 July 2026. Three of these four bars describe the same year under different decisions, so this is a comparison of positions rather than a trend over time.</p>
      </figure>
      
      
    </div>
  </div>
</section>

$ofh$, 'deployed', now(), 'oufe-workstream', 'permanent'
FROM pg, comp
WHERE NOT EXISTS (SELECT 1 FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water'
    AND pc.slot_name='evidence-chart-ofwat');

UPDATE pages SET sections=(SELECT jsonb_agg(DISTINCT x)
  FROM jsonb_array_elements(COALESCE(sections,'[]'::jsonb) || '["evidence-chart-ofwat"]'::jsonb) x),
  build_status='needs_rebuild'
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water';
COMMIT;
SELECT slot_name, length(rendered_html) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water' ORDER BY pc.position;
