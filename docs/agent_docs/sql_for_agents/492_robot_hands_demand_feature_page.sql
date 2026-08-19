-- 492_robot_hands_demand_feature_page.sql
-- The first news editorial feature: /insights/robot-demand-step-change.html on
-- robot-hands.com (news_editorial_features lane, DESIGN_2026-08-19 section 2;
-- design D's step 3, the hand-authored worked example).
--
-- Follows the Thames pattern (252/266) verbatim: content_data for the chart
-- sections copies the register EXACTLY (no display keys, no divergence),
-- rendered_html was produced by executing each component's LIVE template
-- through text/template with executeGoTemplate's funcmap and missingkey=zero
-- (call_agent.go:1168), and every row is locked permanent WITH rendered_html
-- written in the same statement - a locked row is preserved, never rendered,
-- so an empty one would render as nothing for ever.
--
-- Slot names are instance names matching pages.sections entry-for-entry (the
-- thames-water sections array is the precedent; bugs_open/095 is what happens
-- when they diverge). rebuild_policy='owned' + permanent locks keep every
-- generic rebuild path off the authored copy.
--
-- DEPENDS ON 491 (the facts must be registered before a page cites them).
-- Deploy after apply: assemble-only rerender (stitches stored rendered_html):
--   ./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_rerender_page.sh \
--     robot-demand-step-change robot-hands.com ""
--
-- ROLLBACK: DELETE FROM page_components WHERE page_id=(SELECT id FROM pages
--   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='robot-demand-step-change');
--   DELETE FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='robot-demand-step-change';

\set ON_ERROR_STOP on
BEGIN;

-- Guard: 491 applied, page not already present.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM site_specs, jsonb_array_elements(data->'facts') f
     WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND aspect='evidence_base' AND is_current
       AND f->>'id'='rh-ifr-installations-series') THEN
    RAISE EXCEPTION 'apply 491 first - the page cites facts that are not registered';
  END IF;
  IF EXISTS (SELECT 1 FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='robot-demand-step-change') THEN
    RAISE EXCEPTION 'page already exists - refusing double apply';
  END IF;
END $$;

INSERT INTO pages (site_id, name, url, title, page_type, status, build_status,
                   nav_label, nav_order, in_header, in_footer, rebuild_policy,
                   meta_description, sections, topics)
VALUES ('00ff3af5-dad8-4770-9f70-3edc267a3c92',
  'robot-demand-step-change',
  '/insights/robot-demand-step-change.html',
  'Half a Million Robots a Year: Reading the Demand Story Properly | Robot-Hands.com',
  'blog-post', 'active', 'pending',
  'Robot demand: the step change', 80, false, true, 'owned',
  'An editorial feature reading this week''s robot-demand coverage across several channels, charted against the IFR''s own five-year installation series - a step change that has held at altitude, and what that plateau means for end-of-arm tooling.',
  '["hero", "feature-analysis", "evidence-timeseries-ifr", "evidence-chart-2024", "feature-coverage", "call-to-action"]'::jsonb,
  ARRAY['industrial robots','robot demand','IFR','editorial feature']);

CREATE TEMP TABLE _pg ON COMMIT DROP AS
  SELECT id FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='robot-demand-step-change';

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 1, 'hero',
  $cd1${
  "headline": "Half a Million Robots a Year: Reading the Demand Story Properly",
  "subheadline": "The wires call it a record. The series underneath says something more useful: demand for factory robots stepped up once, sharply, and has since held at altitude. This feature reads the same story as it arrived from several channels, charts the background from the industry body's own figures, and says what a plateau at this height means if you specify end-of-arm tooling.",
  "cta_text": "Run MatchMatrix",
  "cta_url": "/tools/matchmatrix/index.html",
  "cta_target_title": "Run MatchMatrix | Gripper Selection Tool | Robot-Hands.com",
  "secondary_cta": "Browse the gripper index",
  "secondary_cta_url": "/gripper-catalog/index.html",
  "secondary_cta_target_title": "Browse Gripper Catalog | Filter by Technology, Payload, IP Rating | Robot-Hands.com"
}$cd1$::jsonb,
  $rh1$<section class="hero" data-component="hero" style="--hero-ink: var(--color-primary-text); background: var(--color-primary); background: linear-gradient(135deg, var(--color-primary) 0%, color-mix(in srgb, var(--color-primary) 85%, var(--color-primary-text)) 100%);">
        <div class="hero-content">
            <h1>Half a Million Robots a Year: Reading the Demand Story Properly</h1>
            <p class="hero-subheadline">The wires call it a record. The series underneath says something more useful: demand for factory robots stepped up once, sharply, and has since held at altitude. This feature reads the same story as it arrived from several channels, charts the background from the industry body's own figures, and says what a plateau at this height means if you specify end-of-arm tooling.</p>
            <a href="/tools/matchmatrix/index.html" class="btn btn-primary">Run MatchMatrix</a>
            <a href="/gripper-catalog/index.html" class="btn btn-secondary">Browse the gripper index</a>
        </div>
    </section>
<style>
.hero {
    min-height: 70vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    position: relative;

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--hero-ink) 95%, transparent);
    --section-text-muted: color-mix(in srgb, var(--hero-ink) 80%, transparent);
    --section-heading: var(--hero-ink);
    --section-surface: color-mix(in srgb, var(--hero-ink) 10%, transparent);
    --section-border: color-mix(in srgb, var(--hero-ink) 30%, transparent);
}
.hero-content {
    max-width: 900px;
    margin: 0 auto;
    color: var(--hero-ink);
    z-index: 1;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 700;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    line-height: 1.6;
}
.hero .btn {
    display: inline-block;
    padding: 0.875rem 2rem;
    margin: 0.5rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 600;
    font-size: 1rem;
    transition: all 0.2s ease;
}
.hero .btn-primary {
    background: var(--hero-ink);
    /* The inverted hero button's text must contrast with --hero-ink, not with
       the page. In the no-image branch --hero-ink IS --color-primary-text, so
       --color-primary is correct by construction and stays the fallback. The
       image branch hard-codes --hero-ink: #fff, where a light --color-primary
       is unreadable, so that branch supplies --hero-btn-ink alongside it. */
    color: var(--hero-btn-ink, var(--color-primary));
    border: 2px solid var(--hero-ink);
}
.hero .btn-primary:hover {
    background: transparent;
    color: var(--hero-ink);
}
.hero .btn-secondary {
    background: transparent;
    color: var(--hero-ink);
    border: 2px solid color-mix(in srgb, var(--hero-ink) 80%, transparent);
}
.hero .btn-secondary:hover {
    background: color-mix(in srgb, var(--hero-ink) 10%, transparent);
}
@media (max-width: 768px) {
    .hero {
        min-height: 60vh;
        padding: 3rem 1.5rem;
    }
    .hero .btn {
        display: block;
        width: 100%;
        max-width: 280px;
        margin: 0.5rem auto;
    }
}
</style>

$rh1$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 2, 'feature-analysis',
  $cd2${
 "ComponentID": "feature-analysis",
 "heading": "The story across the wires, and the number underneath it",
 "content": "<p>This week's trade coverage tells one story from several directions. A wire release on rising orders was carried, near verbatim, by more than one outlet our feed watches; a separate piece framed record robot adoption around the labour shortage that is driving it; earnings coverage read the same demand through vendors' revenue lines. One report pulled the other way: the North American trade body counted a weak quarter for orders in the United States even as revenues climbed. Four framings, one underlying question &mdash; is demand for factory robots actually stepping up, or is this a news cycle?</p>\n<p>There is a series that answers it, and it belongs to the industry's own statistics body, the International Federation of Robotics. In 2020, worldwide installations of industrial robots stood at 384,000 units. The following year they jumped to 517,385 installations &mdash; the step. Since then the figure has held above half a million every year: 553,052 installations in 2022, then 541,302 in 2023, then 542,000 installed in 2024.</p>\n<p>Read carefully, that is not a boom &mdash; it is a step change followed by a plateau at altitude. The last two years sit slightly <em>below</em> the 2022 peak. The genuinely record-breaking number in the headlines is the stock, not the flow: by 2024 there were 4,664,000 industrial robots in operational use worldwide, because installations at this rate accumulate faster than old cells retire. The IFR's own forecast points modestly upward &mdash; 575,000 units expected in 2025 &mdash; growth, but nothing like the 2021 discontinuity.</p>\n<p>Where the machines go explains the one dissenting headline. Asia took 74% of new deployments in 2024; Europe took 16%, and the Americas 9%. China alone installed 295,000 units &mdash; more than half the world's total &mdash; while Japan installed 44,500. A soft quarter for orders in the United States is entirely compatible with a global plateau at altitude, because the United States is a single-digit share of where robots are actually going. Neither headline is wrong; they are measuring different places.</p>\n<p>For anyone specifying end-of-arm tooling, the plateau is the useful fact. A demand step that has <em>held</em> for four consecutive years is not a cycle to wait out &mdash; it is the new baseline, and every one of those installations ends in a gripper, a vacuum cup, a magnetic pad or a tool changer that somebody had to specify against real payloads and real cycle times. That specification problem is the one this site exists for, and it is the same whether the arm went into Shenzhen or Stuttgart.</p>"
}$cd2$::jsonb,
  $rh2$<section id="feature-analysis" class="section section--generic">
  <div class="container">
    <h2 class="section__title">The story across the wires, and the number underneath it</h2>
    <div class="section__content"><p>This week's trade coverage tells one story from several directions. A wire release on rising orders was carried, near verbatim, by more than one outlet our feed watches; a separate piece framed record robot adoption around the labour shortage that is driving it; earnings coverage read the same demand through vendors' revenue lines. One report pulled the other way: the North American trade body counted a weak quarter for orders in the United States even as revenues climbed. Four framings, one underlying question &mdash; is demand for factory robots actually stepping up, or is this a news cycle?</p>
<p>There is a series that answers it, and it belongs to the industry's own statistics body, the International Federation of Robotics. In 2020, worldwide installations of industrial robots stood at 384,000 units. The following year they jumped to 517,385 installations &mdash; the step. Since then the figure has held above half a million every year: 553,052 installations in 2022, then 541,302 in 2023, then 542,000 installed in 2024.</p>
<p>Read carefully, that is not a boom &mdash; it is a step change followed by a plateau at altitude. The last two years sit slightly <em>below</em> the 2022 peak. The genuinely record-breaking number in the headlines is the stock, not the flow: by 2024 there were 4,664,000 industrial robots in operational use worldwide, because installations at this rate accumulate faster than old cells retire. The IFR's own forecast points modestly upward &mdash; 575,000 units expected in 2025 &mdash; growth, but nothing like the 2021 discontinuity.</p>
<p>Where the machines go explains the one dissenting headline. Asia took 74% of new deployments in 2024; Europe took 16%, and the Americas 9%. China alone installed 295,000 units &mdash; more than half the world's total &mdash; while Japan installed 44,500. A soft quarter for orders in the United States is entirely compatible with a global plateau at altitude, because the United States is a single-digit share of where robots are actually going. Neither headline is wrong; they are measuring different places.</p>
<p>For anyone specifying end-of-arm tooling, the plateau is the useful fact. A demand step that has <em>held</em> for four consecutive years is not a cycle to wait out &mdash; it is the new baseline, and every one of those installations ends in a gripper, a vacuum cup, a magnetic pad or a tool changer that somebody had to specify against real payloads and real cycle times. That specification problem is the one this site exists for, and it is the same whether the arm went into Shenzhen or Stuttgart.</p></div>
  </div>
</section>
$rh2$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, 'fb870e82-2f01-46e4-9552-e764515e18d8', 3, 'evidence-timeseries-ifr',
  $cd3${
 "ComponentID": "evidence-timeseries-ifr",
 "eyebrow": "The background series",
 "section_title": "Five years of global robot installations",
 "intro": "One series decides whether the demand story is a cycle or a step change: how many industrial robots the world actually installs each year, as counted by the industry's own statistics body. Each point below is taken from the IFR World Robotics press release for that year, and each carries its own citation.",
 "series": [
  {
   "fact_id": "rh-ifr-installations-series",
   "max_fact_id": "rh-ifr-forecast-2025",
   "unit": "",
   "label": "Industrial robots installed worldwide per year (units)",
   "note": "As first reported in each year's IFR World Robotics press release."
  }
 ],
 "facts": [
  {
   "id": "rh-ifr-installations-series",
   "kind": "series",
   "claim": "Annual global industrial robot installations, as first reported in each year's IFR World Robotics press release",
   "context_terms": [
    "installations",
    "installed",
    "robots"
   ],
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "International Federation of Robotics",
     "title": "IFR press releases, World Robotics 2021 through 2025 editions",
     "url": "https://ifr.org/ifr-press-releases",
     "quote": "series parent; every observation below carries its own release",
     "accessed": "2026-08-19"
    }
   },
   "observations": [
    {
     "as_of": "2020",
     "value": 384000,
     "verified_at": "2026-08-19",
     "source": {
      "citation": {
       "publisher": "International Federation of Robotics",
       "title": "IFR presents World Robotics 2021 reports",
       "url": "https://ifr.org/ifr-press-releases/news/robot-sales-rise-again",
       "quote": "with 384,000 units shipped globally in 2020",
       "accessed": "2026-08-19",
       "published": "2021-10-28"
      }
     }
    },
    {
     "as_of": "2021",
     "value": 517385,
     "verified_at": "2026-08-19",
     "source": {
      "citation": {
       "publisher": "International Federation of Robotics",
       "title": "World Robotics Report: \"All-Time High\" with Half a Million Robots Installed in one Year",
       "url": "https://ifr.org/ifr-press-releases/news/wr-report-all-time-high-with-half-a-million-robots-installed",
       "quote": "an all-time high of 517,385 new industrial robots installed in 2021 in factories around the world",
       "accessed": "2026-08-19",
       "published": "2022-10-13"
      }
     }
    },
    {
     "as_of": "2022",
     "value": 553052,
     "verified_at": "2026-08-19",
     "source": {
      "citation": {
       "publisher": "International Federation of Robotics",
       "title": "World Robotics 2023 Report: Asia ahead of Europe and the Americas",
       "url": "https://ifr.org/ifr-press-releases/news/world-robotics-2023-report-asia-ahead-of-europe-and-the-americas",
       "quote": "recorded 553,052 industrial robot installations in factories around the world",
       "accessed": "2026-08-19",
       "published": "2023-09-26"
      }
     }
    },
    {
     "as_of": "2023",
     "value": 541302,
     "verified_at": "2026-08-19",
     "source": {
      "citation": {
       "publisher": "International Federation of Robotics",
       "title": "Record of 4 Million Robots Working in Factories Worldwide",
       "url": "https://ifr.org/ifr-press-releases/news/record-of-4-million-robots-working-in-factories-worldwide",
       "quote": "The annual installation figure of 541,302 units in 2023 is the second highest in history",
       "accessed": "2026-08-19",
       "published": "2024-09-24"
      }
     }
    },
    {
     "as_of": "2024",
     "value": 542000,
     "verified_at": "2026-08-19",
     "source": {
      "citation": {
       "publisher": "International Federation of Robotics",
       "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years",
       "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years",
       "quote": "542,000 robots installed in 2024 - more than double the number 10 years ago",
       "accessed": "2026-08-19",
       "published": "2025-09-25"
      }
     }
    }
   ]
  },
  {
   "id": "rh-ifr-forecast-2025",
   "kind": "metric",
   "claim": "IFR forecast for global industrial robot installations in 2025",
   "value": 575000,
   "tolerance": "exact",
   "context_terms": [
    "forecast",
    "expected"
   ],
   "writer_line": "robot installations are expected to grow to {value} units in 2025",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "The Robot Report",
     "title": "IFR: industrial robot deployments have doubled in 10 years",
     "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/",
     "quote": "Globally, robot installations are expected to grow by 6% to 575,000 units in 2025",
     "accessed": "2026-08-19"
    }
   }
  }
 ],
 "footnote": "Figures are as first reported in each year's release; the IFR revises figures in later editions, and the restatements are small but real \u2014 its 2024-edition release, for example, restates the 2022 figure slightly below the number first reported. One basis, stated, beats two bases mixed. The scale is itself a registered figure: a bar reaching the top of the plot would exactly meet the IFR's 2025 forecast of 575,000 units."
}$cd3$::jsonb,
  $rh3$
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
</style><section id="evidence-timeseries-ifr" class="ev-ts" data-component="evidence-timeseries">
  <div class="ev-ts__inner">
    <span class="ev-ts__eyebrow">The background series</span>
    <h2 class="ev-ts__title">Five years of global robot installations</h2>
    <p class="ev-ts__intro">One series decides whether the demand story is a cycle or a step change: how many industrial robots the world actually installs each year, as counted by the industry's own statistics body. Each point below is taken from the IFR World Robotics press release for that year, and each carries its own citation.</p>

    
      <figure class="ev-ts__figure" data-series="rh-ifr-installations-series">
        <figcaption class="ev-ts__caption">
          <span class="ev-ts__label">Industrial robots installed worldwide per year (units)</span>
          <span class="ev-ts__note">As first reported in each year's IFR World Robotics press release.</span>
        </figcaption>

        <div class="ev-ts__plot">
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">384000</span>
            <span class="ev-ts__bar" style="--v:384000.0000;--m:575000.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">517385</span>
            <span class="ev-ts__bar" style="--v:517385.0000;--m:575000.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">553052</span>
            <span class="ev-ts__bar" style="--v:553052.0000;--m:575000.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">541302</span>
            <span class="ev-ts__bar" style="--v:541302.0000;--m:575000.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">542000</span>
            <span class="ev-ts__bar" style="--v:542000.0000;--m:575000.0000" aria-hidden="true"></span>
          </div>
          
        </div>
        <div class="ev-ts__axis">
          <span class="ev-ts__tick">2020</span><span class="ev-ts__tick">2021</span><span class="ev-ts__tick">2022</span><span class="ev-ts__tick">2023</span><span class="ev-ts__tick">2024</span>
        </div>

        <div class="ev-ts__sources">
          Every point above is a separately sourced observation. Each carries the date the
          figure applies to, and where we read it:
          <ul>
            
            <li>2020 —
              <a href="https://ifr.org/ifr-press-releases/news/robot-sales-rise-again" rel="noopener noreferrer">IFR presents World Robotics 2021 reports</a>, read 2026-08-19
              (last checked 2026-08-19)
            </li>
            
            <li>2021 —
              <a href="https://ifr.org/ifr-press-releases/news/wr-report-all-time-high-with-half-a-million-robots-installed" rel="noopener noreferrer">World Robotics Report: "All-Time High" with Half a Million Robots Installed in one Year</a>, read 2026-08-19
              (last checked 2026-08-19)
            </li>
            
            <li>2022 —
              <a href="https://ifr.org/ifr-press-releases/news/world-robotics-2023-report-asia-ahead-of-europe-and-the-americas" rel="noopener noreferrer">World Robotics 2023 Report: Asia ahead of Europe and the Americas</a>, read 2026-08-19
              (last checked 2026-08-19)
            </li>
            
            <li>2023 —
              <a href="https://ifr.org/ifr-press-releases/news/record-of-4-million-robots-working-in-factories-worldwide" rel="noopener noreferrer">Record of 4 Million Robots Working in Factories Worldwide</a>, read 2026-08-19
              (last checked 2026-08-19)
            </li>
            
            <li>2024 —
              <a href="https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years" rel="noopener noreferrer">World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years</a>, read 2026-08-19
              (last checked 2026-08-19)
            </li>
            
          </ul>
        </div>
      </figure>
      
    

    <p class="ev-ts__footnote">Figures are as first reported in each year's release; the IFR revises figures in later editions, and the restatements are small but real — its 2024-edition release, for example, restates the 2022 figure slightly below the number first reported. One basis, stated, beats two bases mixed. The scale is itself a registered figure: a bar reaching the top of the plot would exactly meet the IFR's 2025 forecast of 575,000 units.</p>
  </div>
</section>

$rh3$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, 'f8c2393c-fc66-480d-93f7-738c68cd5f1b', 4, 'evidence-chart-2024',
  $cd4${
 "section_eyebrow": "Where the machines went",
 "section_title": "2024's installations, by place",
 "section_intro": "The regional split is what reconciles the week's contradictory headlines: a soft quarter in one country is compatible with a plateau at altitude worldwide, because deployment is concentrated where the dissenting headline was not looking.",
 "charts": [
  {
   "id": "regional-share-2024",
   "title": "Share of new deployments, 2024",
   "caption": "Scale: Asia's own share \u2014 the largest bar fills the track.",
   "unit": "%",
   "max_fact_id": "rh-ifr-asia-share-2024",
   "points": [
    {
     "fact_id": "rh-ifr-asia-share-2024",
     "label": "Asia",
     "tone": "accent"
    },
    {
     "fact_id": "rh-ifr-europe-share-2024",
     "label": "Europe"
    },
    {
     "fact_id": "rh-ifr-americas-share-2024",
     "label": "Americas"
    }
   ],
   "source_note": "IFR World Robotics 2025 press release: \u201cAsia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas.\u201d"
  },
  {
   "id": "national-markets-2024",
   "title": "The two largest national markets, against the world total (units, 2024)",
   "caption": "Scale: the world total \u2014 the top bar fills the track.",
   "unit": "",
   "max_fact_id": "rh-ifr-global-2024",
   "points": [
    {
     "fact_id": "rh-ifr-global-2024",
     "label": "World",
     "tone": "accent"
    },
    {
     "fact_id": "rh-ifr-china-2024",
     "label": "China"
    },
    {
     "fact_id": "rh-ifr-japan-2024",
     "label": "Japan"
    }
   ],
   "source_note": "IFR World Robotics 2025 data as reported by The Robot Report, cited in the register point by point."
  }
 ],
 "facts": [
  {
   "id": "rh-ifr-asia-share-2024",
   "kind": "metric",
   "claim": "Asia's share of new industrial robot deployments in 2024",
   "value": 74,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "asia"
   ],
   "writer_line": "Asia accounted for {value}% of new deployments in 2024",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "International Federation of Robotics",
     "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years",
     "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years",
     "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas",
     "accessed": "2026-08-19",
     "published": "2025-09-25"
    }
   }
  },
  {
   "id": "rh-ifr-europe-share-2024",
   "kind": "metric",
   "claim": "Europe's share of new industrial robot deployments in 2024",
   "value": 16,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "europe"
   ],
   "writer_line": "{value}% of new deployments in 2024 were in Europe",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "International Federation of Robotics",
     "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years",
     "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years",
     "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas",
     "accessed": "2026-08-19",
     "published": "2025-09-25"
    }
   }
  },
  {
   "id": "rh-ifr-americas-share-2024",
   "kind": "metric",
   "claim": "The Americas' share of new industrial robot deployments in 2024",
   "value": 9,
   "unit": "percent",
   "tolerance": "exact",
   "context_terms": [
    "americas"
   ],
   "writer_line": "{value}% of new deployments in 2024 were in the Americas",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "International Federation of Robotics",
     "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years",
     "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years",
     "quote": "Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas",
     "accessed": "2026-08-19",
     "published": "2025-09-25"
    }
   }
  },
  {
   "id": "rh-ifr-global-2024",
   "kind": "metric",
   "claim": "Industrial robots installed worldwide in 2024, per the IFR World Robotics 2025 press release",
   "value": 542000,
   "tolerance": "exact",
   "context_terms": [
    "installed",
    "installations",
    "world"
   ],
   "writer_line": "{value} industrial robots were installed worldwide in 2024",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "International Federation of Robotics",
     "title": "World Robotics 2025 report - Global Robot Demand in Factories Doubles Over 10 Years",
     "url": "https://ifr.org/ifr-press-releases/news/global-robot-demand-in-factories-doubles-over-10-years",
     "quote": "542,000 robots installed in 2024 - more than double the number 10 years ago",
     "accessed": "2026-08-19",
     "published": "2025-09-25"
    }
   }
  },
  {
   "id": "rh-ifr-china-2024",
   "kind": "metric",
   "claim": "Industrial robots installed in China in 2024",
   "value": 295000,
   "tolerance": "exact",
   "context_terms": [
    "china"
   ],
   "writer_line": "China installed {value} industrial robots in 2024",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "The Robot Report",
     "title": "IFR: industrial robot deployments have doubled in 10 years",
     "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/",
     "quote": "295,000 industrial robots have been installed in the country, the highest annual total on record",
     "accessed": "2026-08-19"
    }
   }
  },
  {
   "id": "rh-ifr-japan-2024",
   "kind": "metric",
   "claim": "Industrial robots installed in Japan in 2024",
   "value": 44500,
   "tolerance": "exact",
   "context_terms": [
    "japan"
   ],
   "writer_line": "Japan installed {value} industrial robots in 2024",
   "verified_at": "2026-08-19",
   "source": {
    "citation": {
     "publisher": "The Robot Report",
     "title": "IFR: industrial robot deployments have doubled in 10 years",
     "url": "https://www.therobotreport.com/ifr-industrial-robot-deployments-have-doubled-in-10-years/",
     "quote": "44,500 units installed in 2024",
     "accessed": "2026-08-19"
    }
   }
  }
 ]
}$cd4$::jsonb,
  $rh4$<style>
  /* Colour vocabulary checked against the live css_themes rather than assumed:
     --color-background / --color-text / --color-text-muted / --color-primary /
     --color-secondary / --color-accent / --color-card-bg / --color-border /
     --border-radius are the names the themes actually define (including the
     dark one). --color-surface, --spacing-section and --container-max-width are
     defined by NO theme, so anything resting on them renders its fallback on
     every site — which is how a light card ends up on a dark page. Where no
     variable exists, the fallback is a neutral translucent grey that reads
     correctly on light and dark alike, never a light literal. */
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
  /* Geometry is computed by the browser from the real value: --v is the
     figure itself and --m the chart's declared maximum. Nothing rounds it
     into the template, and there is no width to get wrong by hand. */
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
      <span class="evidence-chart__eyebrow">Where the machines went</span>
      <h2 class="evidence-chart__title">2024's installations, by place</h2>
      <p class="evidence-chart__intro">The regional split is what reconciles the week's contradictory headlines: a soft quarter in one country is compatible with a plateau at altitude worldwide, because deployment is concentrated where the dissenting headline was not looking.</p>
    </header>
    
    <div class="evidence-chart__grid">
      <figure class="evidence-chart__figure" data-chart="regional-share-2024">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Share of new deployments, 2024</span>
          <span class="evidence-chart__chart-note">Scale: Asia's own share — the largest bar fills the track.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Asia</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:74.0000;--m:74.0000"></span></span>
          <span class="evidence-chart__value">74%</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Europe</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:16.0000;--m:74.0000"></span></span>
          <span class="evidence-chart__value">16%</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Americas</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:9.0000;--m:74.0000"></span></span>
          <span class="evidence-chart__value">9%</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <p class="evidence-chart__source">IFR World Robotics 2025 press release: “Asia accounted for 74% of new deployments in 2024, compared with 16% in Europe and 9% in the Americas.”</p>
      </figure>
      
      <figure class="evidence-chart__figure" data-chart="national-markets-2024">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">The two largest national markets, against the world total (units, 2024)</span>
          <span class="evidence-chart__chart-note">Scale: the world total — the top bar fills the track.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">World</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:542000.0000;--m:542000.0000"></span></span>
          <span class="evidence-chart__value">542000</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">China</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:295000.0000;--m:542000.0000"></span></span>
          <span class="evidence-chart__value">295000</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Japan</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:44500.0000;--m:542000.0000"></span></span>
          <span class="evidence-chart__value">44500</span>
          <span class="evidence-chart__verified">verified 2026-08-19</span>
        </div>
        <p class="evidence-chart__source">IFR World Robotics 2025 data as reported by The Robot Report, cited in the register point by point.</p>
      </figure>
      
      
    </div>
  </div>
</section>

$rh4$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 5, 'feature-coverage',
  $cd5${
 "ComponentID": "feature-coverage",
 "heading": "The coverage this feature draws on",
 "content": "<p>This feature was assembled from items our own news feed ingested &mdash; the same story arriving on more than one channel is the signal an editorial feature exists to read. Titles link to the original publishers; we quote headlines and link out, never republish.</p>\n<ul>\n<li><a href=\"https://investingnews.com/industrial-robot-installations-hit-record-highs-amid-labor-shortage-crisis/\" rel=\"noopener noreferrer\">Industrial Robot Installations Hit Record Highs Amid Labor Shortage Crisis</a> &mdash; Investing News Network</li>\n<li><a href=\"https://www.businesswire.com/news/home/20260811859922/en/Robot-Orders-Increase-in-Q2-as-Automation-Demand-Broadens-Across-Industries\" rel=\"noopener noreferrer\">Robot Orders Increase in Q2 as Automation Demand Broadens Across Industries</a> &mdash; Business Wire (the originating release)</li>\n<li><a href=\"https://theaiinsider.tech/2026/08/15/robot-orders-increase-in-q2-as-automation-demand-broadens-across-industries/\" rel=\"noopener noreferrer\">The same release, as carried by The AI Insider</a> &mdash; two channels, one story: this duplication across outlets is itself the evidence that the story is the week's biggest</li>\n<li><a href=\"https://www.scdigest.com/ontarget/26-08-12.php?cid=23188\" rel=\"noopener noreferrer\">US Robot Orders Weak in Q2, A3 Finds, while Revenues Much Stronger</a> &mdash; Supply Chain Digest: the divergent read, and the reason the regional chart above matters</li>\n</ul>\n<p>Background figures are drawn from the International Federation of Robotics' World Robotics press releases, cited point by point beneath the charts.</p>"
}$cd5$::jsonb,
  $rh5$<section id="feature-coverage" class="section section--generic">
  <div class="container">
    <h2 class="section__title">The coverage this feature draws on</h2>
    <div class="section__content"><p>This feature was assembled from items our own news feed ingested &mdash; the same story arriving on more than one channel is the signal an editorial feature exists to read. Titles link to the original publishers; we quote headlines and link out, never republish.</p>
<ul>
<li><a href="https://investingnews.com/industrial-robot-installations-hit-record-highs-amid-labor-shortage-crisis/" rel="noopener noreferrer">Industrial Robot Installations Hit Record Highs Amid Labor Shortage Crisis</a> &mdash; Investing News Network</li>
<li><a href="https://www.businesswire.com/news/home/20260811859922/en/Robot-Orders-Increase-in-Q2-as-Automation-Demand-Broadens-Across-Industries" rel="noopener noreferrer">Robot Orders Increase in Q2 as Automation Demand Broadens Across Industries</a> &mdash; Business Wire (the originating release)</li>
<li><a href="https://theaiinsider.tech/2026/08/15/robot-orders-increase-in-q2-as-automation-demand-broadens-across-industries/" rel="noopener noreferrer">The same release, as carried by The AI Insider</a> &mdash; two channels, one story: this duplication across outlets is itself the evidence that the story is the week's biggest</li>
<li><a href="https://www.scdigest.com/ontarget/26-08-12.php?cid=23188" rel="noopener noreferrer">US Robot Orders Weak in Q2, A3 Finds, while Revenues Much Stronger</a> &mdash; Supply Chain Digest: the divergent read, and the reason the regional chart above matters</li>
</ul>
<p>Background figures are drawn from the International Federation of Robotics' World Robotics press releases, cited point by point beneath the charts.</p></div>
  </div>
</section>
$rh5$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 6, 'call-to-action',
  $cd6${
 "headline": "Benchmark gripper specifications against your application requirements.",
 "subheadline": "MatchMatrix tests your application against 10 indexed gripper models. It filters gripping/holding force, jaw travel or grip range, rated payload, and IP rating across 6 actuation technologies. The database holds 59 published specification figures from 6 manufacturers. The tool calculates a shortlist based on these stated parameters, but you must verify the final selection against your own system constraints.",
 "primary_cta": "Run MatchMatrix",
 "primary_cta_url": "/tools/matchmatrix/index.html",
 "primary_cta_target_title": "Run MatchMatrix | Gripper Selection Tool | Robot-Hands.com",
 "secondary_cta": "View full specification index",
 "secondary_cta_url": "/gripper-catalog/index.html",
 "secondary_cta_target_title": "Browse Gripper Catalog | Filter by Technology, Payload, IP Rating | Robot-Hands.com"
}$cd6$::jsonb,
  $rh6$<section class="cta-section" data-component="call-to-action">
    <div class="cta-container">
        <h2>Benchmark gripper specifications against your application requirements.</h2>
        <p class="cta-subtitle">MatchMatrix tests your application against 10 indexed gripper models. It filters gripping/holding force, jaw travel or grip range, rated payload, and IP rating across 6 actuation technologies. The database holds 59 published specification figures from 6 manufacturers. The tool calculates a shortlist based on these stated parameters, but you must verify the final selection against your own system constraints.</p>
        <div class="cta-buttons">
            
            <a href="/tools/matchmatrix/index.html" class="cta-btn cta-btn-primary">Run MatchMatrix</a>
            
            
            <a href="/gripper-catalog/index.html" class="cta-btn cta-btn-secondary">View full specification index</a>
            
        </div>
    </div>
</section>
<style>
.cta-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-cta-bg, var(--color-primary));
    color: var(--color-cta-text, var(--color-primary-text));
    text-align: center;

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 90%, transparent);
    --section-text-muted: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 85%, transparent);
    --section-heading: var(--color-cta-text, var(--color-primary-text));
    --section-surface: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 5%, transparent);
    --section-border: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text)) 20%, transparent);
}
.cta-container {
    max-width: 800px;
    margin: 0 auto;
}
.cta-section h2 {
    margin-bottom: 1rem;
}
.cta-subtitle {
    margin-bottom: 2rem;
}
.cta-buttons {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
}
.cta-btn {
    display: inline-block;
    padding: 1rem 2rem;
    border-radius: 6px;
    text-decoration: none;
    font-weight: 600;
    transition: transform 0.2s, box-shadow 0.2s;
}
.cta-btn:hover {
    transform: translateY(-2px);
}
.cta-btn-primary {
    background: var(--color-cta-text, var(--color-primary-text));
    color: var(--color-cta-bg, var(--color-primary));
}
.cta-btn-secondary {
    background: transparent;
    border: 2px solid var(--color-cta-text, var(--color-primary-text));
    color: var(--color-cta-text, var(--color-primary-text));
}
@media (max-width: 768px) {
    .cta-section { padding: 3rem 1.5rem; }
    .cta-buttons { flex-direction: column; align-items: center; }
    .cta-btn { width: 100%; max-width: 280px; text-align: center; }
}
</style>
$rh6$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;

-- Verify: six locked rows, none with empty rendered_html, slots match sections.
DO $$
DECLARE n int; empty int; mismatch int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE coalesce(length(pc.rendered_html),0) < 500)
    INTO n, empty
    FROM page_components pc JOIN _pg ON _pg.id=pc.page_id;
  IF n <> 6 THEN RAISE EXCEPTION 'expected 6 components, got %', n; END IF;
  IF empty > 0 THEN RAISE EXCEPTION '% component(s) with suspiciously short rendered_html', empty; END IF;
  SELECT count(*) INTO mismatch
    FROM pages p, jsonb_array_elements_text(p.sections) s
   WHERE p.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND p.name='robot-demand-step-change'
     AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id AND pc.slot_name=s);
  IF mismatch > 0 THEN RAISE EXCEPTION '% sections entry(ies) with no matching slot_name', mismatch; END IF;
END $$;

COMMIT;
