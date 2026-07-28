BEGIN;
-- The capital structure chart on the Thames page. evidence-chart is render_mode
-- 'agent', but content_data is supplied directly here so no model sits in the
-- path: every plotted value resolves through a fact_id into the register, and
-- the scale denominator is itself a registered fact (the WBS drawn total), which
-- is the rule that stops a chart restating a business figure of its own.
-- rendered_html is written here because save_page_sections PRESERVES locked rows
-- rather than rendering them; a row locked with empty rendered_html renders as
-- nothing for ever. Rendered by executing the component's OWN template.
WITH pg AS (
  SELECT id FROM pages WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water'
), comp AS (SELECT id FROM content_components WHERE name='evidence-chart')
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, comp.id, 3, 'evidence-chart', $ecdata${
 "section_eyebrow": "The capital structure",
 "section_title": "What the classes were actually worth, and why the split matters",
 "section_intro": "The waterfall tool above lets you move the assumptions. These are the real figures the plan was argued over, as they stood at the last financial year end before it was sanctioned. Class A sits ahead of Class B, which is the whole reason the two classes were treated differently in the vote.",
 "facts": [
  {
   "id": "CIT-tw-classa-2024",
   "kind": "metric",
   "unit": "GBP_billion",
   "claim": "Thames Water's Class A debt totalled around \u00a314.7 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum on the restructuring plan",
   "value": 14.7,
   "source": {
    "citation": {
     "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
     "quote": "the Class A debt totalling around \u00a314.7 billion",
     "title": "Three's a Crowd: The Thames Water Restructuring Plan",
     "accessed": "2026-07-28",
     "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Class A debt stood at around \u00a3{value} billion as at 31 March 2024",
   "context_terms": [
    "class a",
    "thames"
   ]
  },
  {
   "id": "CIT-tw-classb-2024",
   "kind": "metric",
   "unit": "GBP_billion",
   "claim": "Thames Water's Class B debt totalled \u00a31.4 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum on the restructuring plan",
   "value": 1.4,
   "source": {
    "citation": {
     "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
     "quote": "the Class B debt totalling \u00a31.4 billion",
     "title": "Three's a Crowd: The Thames Water Restructuring Plan",
     "accessed": "2026-07-28",
     "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Class B debt stood at \u00a3{value} billion as at 31 March 2024",
   "context_terms": [
    "class b",
    "thames"
   ]
  },
  {
   "id": "CIT-tw-wbs-total-2024",
   "kind": "metric",
   "unit": "GBP_billion",
   "claim": "The drawn facilities within Thames Water's whole-business securitisation totalled approximately \u00a316.3 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum",
   "value": 16.3,
   "source": {
    "citation": {
     "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
     "quote": "the various drawn facilities within the WBS totalled c.\u00a316.3 billion",
     "title": "Three's a Crowd: The Thames Water Restructuring Plan",
     "accessed": "2026-07-28",
     "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "Drawn facilities within the securitisation totalled about \u00a3{value} billion as at 31 March 2024",
   "context_terms": [
    "drawn facilities",
    "securitisation",
    "wbs",
    "thames"
   ]
  },
  {
   "id": "CIT-tw-mtm-2024",
   "kind": "metric",
   "unit": "GBP_billion",
   "claim": "Thames Water carried an additional mark-to-market exposure of approximately \u00a31.7 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum",
   "value": 1.7,
   "source": {
    "citation": {
     "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
     "quote": "an additional mark-to-market exposure of c.\u00a31.7 billion as at 31 March 2024",
     "title": "Three's a Crowd: The Thames Water Restructuring Plan",
     "accessed": "2026-07-28",
     "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP"
    }
   },
   "verified_at": "2026-07-28",
   "writer_line": "A further mark-to-market exposure of about \u00a3{value} billion sat alongside the drawn debt as at 31 March 2024",
   "context_terms": [
    "mark-to-market",
    "hedging",
    "thames"
   ]
  }
 ],
 "charts": [
  {
   "id": "tw-structure-2024",
   "title": "Thames Water drawn debt by class, as at 31 March 2024",
   "caption": "Scaled against the total drawn facilities within the securitisation, which is itself a registered figure rather than a number typed into this chart.",
   "unit": "bn",
   "max_fact_id": "CIT-tw-wbs-total-2024",
   "points": [
    {
     "fact_id": "CIT-tw-classa-2024",
     "label": "Class A debt",
     "tone": "accent"
    },
    {
     "fact_id": "CIT-tw-classb-2024",
     "label": "Class B debt"
    },
    {
     "fact_id": "CIT-tw-mtm-2024",
     "label": "Mark-to-market exposure",
     "tone": "muted"
    }
   ],
   "source_note": "All four figures are from the same source, read on 28 July 2026. They describe one date and are not a trend: nothing here says what happened before or after."
  }
 ]
}$ecdata$::jsonb,
       $echtml$<style>
   
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
    color: var(--color-primary, #1e40af);
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
    background: var(--color-primary, #1e40af);
  }
  .evidence-chart__bar--muted {
    background: var(--color-secondary, rgba(127, 127, 127, 0.6));
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
      <span class="evidence-chart__eyebrow">The capital structure</span>
      <h2 class="evidence-chart__title">What the classes were actually worth, and why the split matters</h2>
      <p class="evidence-chart__intro">The waterfall tool above lets you move the assumptions. These are the real figures the plan was argued over, as they stood at the last financial year end before it was sanctioned. Class A sits ahead of Class B, which is the whole reason the two classes were treated differently in the vote.</p>
    </header>
    
    <div class="evidence-chart__grid">
      <figure class="evidence-chart__figure" data-chart="tw-structure-2024">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Thames Water drawn debt by class, as at 31 March 2024</span>
          <span class="evidence-chart__chart-note">Scaled against the total drawn facilities within the securitisation, which is itself a registered figure rather than a number typed into this chart.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Class A debt</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:14.7000;--m:16.3000"></span></span>
          <span class="evidence-chart__value">14.7bn</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Class B debt</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:1.4000;--m:16.3000"></span></span>
          <span class="evidence-chart__value">1.4bn</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Mark-to-market exposure</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--muted" style="--v:1.7000;--m:16.3000"></span></span>
          <span class="evidence-chart__value">1.7bn</span>
          <span class="evidence-chart__verified">verified 2026-07-28</span>
        </div>
        <p class="evidence-chart__source">All four figures are from the same source, read on 28 July 2026. They describe one date and are not a trend: nothing here says what happened before or after.</p>
      </figure>
      
      
    </div>
  </div>
</section>

$echtml$, 'deployed', now(), 'oufe-workstream', 'permanent'
FROM pg, comp
WHERE NOT EXISTS (
  SELECT 1 FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water'
    AND pc.slot_name='evidence-chart');

UPDATE pages SET sections = (
    SELECT jsonb_agg(DISTINCT x)
    FROM jsonb_array_elements(COALESCE(sections,'[]'::jsonb) || '["evidence-chart"]'::jsonb) x),
  build_status='needs_rebuild'
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND name='thames-water';
COMMIT;
SELECT slot_name, COALESCE(lock_type,'-'), length(rendered_html)
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND p.name='thames-water' ORDER BY pc.position;
