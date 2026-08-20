-- 495_dartsonline_calendar_feature_page.sql
-- The SECOND news editorial feature, and the first rollout of NEWS-020's
-- pattern to a new site: /insights/darts-calendar-density.html on
-- dartsonline.com.
--
-- Same six-section shape, same locked+owned discipline, same
-- rendered-locally-through-the-live-template method as 492. Differences worth
-- knowing:
--   * the HERO SHIPS WITH ITS IMAGE from the start (owner ruling 2026-08-20:
--     image + semi-transparent overlay is the editorial default). The asset is
--     generated through the framework under the ContentHeroKey convention
--     (content_hero_darts_calendar_density) and its deployed path is
--     deterministic from that key, so the hero HTML could be rendered before
--     the image landed. DEPLOY ONLY ONCE THE ASSET IS ACTIVE, or the page
--     serves a broken src.
--   * the CTA points at this site's OWN tools (setup-builder,
--     dart-weight-comparator) — a retail site, so the honest turn is equipment,
--     not a hard sell on the news.
--
-- THE PREMISE, per design D's premise-not-topic test: the load-bearing claim is
-- that the withdrawals story is a SCHEDULE-DENSITY story rather than a
-- discipline story. It is falsifiable against exactly one series — the number of
-- Players Championship events per season — which is why that series is the
-- chart. Had the count been flat, the feature's argument would collapse and the
-- discipline framing would stand unchallenged. It is not flat: 30/30/30 through
-- 2022-24, then 34 in 2025 and 34 in 2026.
--
-- The cluster is 4 items across 3 channels (Dartsnews x2, Oche180, The Sun),
-- and the coverage section keeps the DISSENTING one — The Sun's 'disheartening'
-- piece is the same story from the inside and is the reason to look at the
-- calendar at all. Same discipline as 492 kept the weak-US-orders headline.
--
-- DEPENDS ON 494 (facts). ROLLBACK: DELETE FROM page_components WHERE
--   page_id=(SELECT id FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='darts-calendar-density');
--   DELETE FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='darts-calendar-density';

\set ON_ERROR_STOP on
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM site_specs, jsonb_array_elements(data->'facts') f
     WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='evidence_base' AND is_current
       AND f->>'id'='do-pdc-pc-events-series') THEN
    RAISE EXCEPTION 'apply 494 first - the page cites facts that are not registered';
  END IF;
  IF EXISTS (SELECT 1 FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='darts-calendar-density') THEN
    RAISE EXCEPTION 'page already exists - refusing double apply';
  END IF;
END $$;

INSERT INTO pages (site_id, name, url, title, page_type, status, build_status,
                   nav_label, nav_order, in_header, in_footer, rebuild_policy,
                   meta_description, sections, topics)
VALUES ('5fe8785b-223d-41a3-88ee-c07187622381', 'darts-calendar-density',
  '/insights/darts-calendar-density.html',
  'Why the Big Names Are Missing: The Darts Calendar Has Quietly Grown | Dartsonline.com',
  'blog-post', 'active', 'pending',
  'The calendar has grown', 80, false, true, 'owned',
  'Barry Hearn warned top players about skipping tournaments and Euro Tour withdrawals left organisers with a headache. Set against the calendar itself — 30 Players Championship events a season through 2024, 34 since 2025 — these are one story about schedule density, not four about discipline.',
  '["hero", "feature-analysis", "evidence-timeseries-pdc-calendar", "evidence-chart-2026-calendar", "feature-coverage", "call-to-action"]'::jsonb,
  ARRAY['PDC','tournament withdrawals','darts calendar','editorial feature']);

CREATE TEMP TABLE _pg ON COMMIT DROP AS
  SELECT id FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND name='darts-calendar-density';

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 1, 'hero',
  $cd1${
 "headline": "Why the Big Names Are Missing: The Darts Calendar Has Quietly Grown",
 "subheadline": "Barry Hearn has warned top players to behave like top players. Withdrawals from the Euro Tour left organisers with a headache, and a fellow pro called the circuit disheartening. Read together, and set against the calendar itself, these are not four stories about discipline \u2014 they are one story about how many events a professional is now expected to play.",
 "hero_url": "/assets/images/content-hero-darts-calendar-density.jpg",
 "background_image": "/assets/images/content-hero-darts-calendar-density.jpg",
 "cta_text": "Build your setup",
 "cta_url": "/tools/setup-builder/index.html",
 "cta_target_title": "Setup Builder | Dartsonline.com",
 "secondary_cta": "Read the guides",
 "secondary_cta_url": "/guides/index.html",
 "secondary_cta_target_title": "Darts Guides | Dartsonline.com"
}$cd1$::jsonb,
  $rh1$<section class="hero" data-component="hero" style="background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('/assets/images/content-hero-darts-calendar-density.jpg'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;">
        <div class="hero-content">
            <h1>Why the Big Names Are Missing: The Darts Calendar Has Quietly Grown</h1>
            <p class="hero-subheadline">Barry Hearn has warned top players to behave like top players. Withdrawals from the Euro Tour left organisers with a headache, and a fellow pro called the circuit disheartening. Read together, and set against the calendar itself, these are not four stories about discipline — they are one story about how many events a professional is now expected to play.</p>
            <a href="/tools/setup-builder/index.html" class="btn btn-primary">Build your setup</a>
            <a href="/guides/index.html" class="btn btn-secondary">Read the guides</a>
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
 "heading": "Four stories, one calendar",
 "content": "<p>Four stories reached our feed this week and they all point at the same thing. Barry Hearn told players that if you are a top player you should behave like one, aimed squarely at those skipping tournaments. A separate report had Luke Littler's Euro Tour withdrawals leaving organisers facing a headache. Vincent van der Voort criticised what he saw as a lax attitude among PDC stars at the World Series event in Auckland. And a friend of Littler's, talking to a national paper, described life on the PDC circuit as disheartening.</p>\n<p>Three of those are framed as a discipline story. The fourth is the same story from the inside, and it is the one that suggests where to look next: not at the players' attitude, but at the calendar.</p>\n<p>The floor of the professional game is the Players Championship series, and it has grown. From 2022 through 2024 it held steady at 30 events a season. For 2025 it stepped up to 34, and it stays at 34 for 2026. The European Tour has moved the same way \u2014 14 events in 2025, 15 in 2026 \u2014 and each of the PDC's three secondary tours now runs 24 events of its own.</p>\n<p>Put the two ranking circuits together and a player chasing order-of-merit money in 2026 has 34 Players Championship events plus 15 European Tour events available before a single televised major is counted. That is the context the discipline framing leaves out. A withdrawal is a choice, but it is a choice made against a schedule that has four more floor events than it did two years ago, and one more European Tour weekend than last year.</p>\n<p>None of which settles whether any particular absence was justified \u2014 that is a judgement, and the people making it have more information than we do. What the numbers do settle is that \"top players are skipping tournaments\" and \"there are more tournaments to skip\" are the same sentence viewed from either end. If the schedule keeps growing, selective entry stops being a lapse and becomes the rational way to play a season.</p>"
}$cd2$::jsonb,
  $rh2$<section id="feature-analysis" class="section section--generic">
  <div class="container">
    <h2 class="section__title">Four stories, one calendar</h2>
    <div class="section__content"><p>Four stories reached our feed this week and they all point at the same thing. Barry Hearn told players that if you are a top player you should behave like one, aimed squarely at those skipping tournaments. A separate report had Luke Littler's Euro Tour withdrawals leaving organisers facing a headache. Vincent van der Voort criticised what he saw as a lax attitude among PDC stars at the World Series event in Auckland. And a friend of Littler's, talking to a national paper, described life on the PDC circuit as disheartening.</p>
<p>Three of those are framed as a discipline story. The fourth is the same story from the inside, and it is the one that suggests where to look next: not at the players' attitude, but at the calendar.</p>
<p>The floor of the professional game is the Players Championship series, and it has grown. From 2022 through 2024 it held steady at 30 events a season. For 2025 it stepped up to 34, and it stays at 34 for 2026. The European Tour has moved the same way — 14 events in 2025, 15 in 2026 — and each of the PDC's three secondary tours now runs 24 events of its own.</p>
<p>Put the two ranking circuits together and a player chasing order-of-merit money in 2026 has 34 Players Championship events plus 15 European Tour events available before a single televised major is counted. That is the context the discipline framing leaves out. A withdrawal is a choice, but it is a choice made against a schedule that has four more floor events than it did two years ago, and one more European Tour weekend than last year.</p>
<p>None of which settles whether any particular absence was justified — that is a judgement, and the people making it have more information than we do. What the numbers do settle is that "top players are skipping tournaments" and "there are more tournaments to skip" are the same sentence viewed from either end. If the schedule keeps growing, selective entry stops being a lapse and becomes the rational way to play a season.</p></div>
  </div>
</section>
$rh2$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, 'fb870e82-2f01-46e4-9552-e764515e18d8', 3, 'evidence-timeseries-pdc-calendar',
  $cd3${
 "ComponentID": "evidence-timeseries-pdc-calendar",
 "eyebrow": "The background series",
 "section_title": "Players Championship events per season",
 "intro": "The Players Championship series is the floor of the professional game \u2014 the events where ranking money is won away from television. Its size is the simplest measure of how much a season now asks of a player. Each point below is taken from that season's own summary and carries its own citation.",
 "series": [
  {
   "fact_id": "do-pdc-pc-events-series",
   "max_fact_id": "do-pdc-pc-2026",
   "unit": "",
   "label": "PDC Players Championship events per season",
   "note": "Counts as stated in each season's summary."
  }
 ],
 "facts": [
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
  }
 ],
 "footnote": "Figures are the event counts stated in each season's own summary page, which is a secondary source rather than the PDC's own calendar listing \u2014 the publisher is named in every citation so a reader can judge it. The scale is itself a registered figure: a bar reaching the top of the plot equals the current 2026 count of 34."
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
</style><section id="evidence-timeseries-pdc-calendar" class="ev-ts" data-component="evidence-timeseries">
  <div class="ev-ts__inner">
    <span class="ev-ts__eyebrow">The background series</span>
    <h2 class="ev-ts__title">Players Championship events per season</h2>
    <p class="ev-ts__intro">The Players Championship series is the floor of the professional game — the events where ranking money is won away from television. Its size is the simplest measure of how much a season now asks of a player. Each point below is taken from that season's own summary and carries its own citation.</p>

    
      <figure class="ev-ts__figure" data-series="do-pdc-pc-events-series">
        <figcaption class="ev-ts__caption">
          <span class="ev-ts__label">PDC Players Championship events per season</span>
          <span class="ev-ts__note">Counts as stated in each season's summary.</span>
        </figcaption>

        <div class="ev-ts__plot">
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">30</span>
            <span class="ev-ts__bar" style="--v:30.0000;--m:34.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">30</span>
            <span class="ev-ts__bar" style="--v:30.0000;--m:34.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">30</span>
            <span class="ev-ts__bar" style="--v:30.0000;--m:34.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">34</span>
            <span class="ev-ts__bar" style="--v:34.0000;--m:34.0000" aria-hidden="true"></span>
          </div>
          
          <div class="ev-ts__col">
            <span class="ev-ts__val">34</span>
            <span class="ev-ts__bar" style="--v:34.0000;--m:34.0000" aria-hidden="true"></span>
          </div>
          
        </div>
        <div class="ev-ts__axis">
          <span class="ev-ts__tick">2022</span><span class="ev-ts__tick">2023</span><span class="ev-ts__tick">2024</span><span class="ev-ts__tick">2025</span><span class="ev-ts__tick">2026</span>
        </div>

        <div class="ev-ts__sources">
          Every point above is a separately sourced observation. Each carries the date the
          figure applies to, and where we read it:
          <ul>
            
            <li>2022 —
              <a href="https://en.wikipedia.org/wiki/2022_PDC_Players_Championship_series" rel="noopener noreferrer">2022 PDC Players Championship series</a>, read 2026-08-20
              (last checked 2026-08-20)
            </li>
            
            <li>2023 —
              <a href="https://en.wikipedia.org/wiki/2023_PDC_Players_Championship_series" rel="noopener noreferrer">2023 PDC Players Championship series</a>, read 2026-08-20
              (last checked 2026-08-20)
            </li>
            
            <li>2024 —
              <a href="https://en.wikipedia.org/wiki/2024_PDC_Players_Championship_series" rel="noopener noreferrer">2024 PDC Players Championship series</a>, read 2026-08-20
              (last checked 2026-08-20)
            </li>
            
            <li>2025 —
              <a href="https://en.wikipedia.org/wiki/2025_PDC_Pro_Tour" rel="noopener noreferrer">2025 PDC Pro Tour</a>, read 2026-08-20
              (last checked 2026-08-20)
            </li>
            
            <li>2026 —
              <a href="https://en.wikipedia.org/wiki/2026_PDC_Pro_Tour" rel="noopener noreferrer">2026 PDC Pro Tour</a>, read 2026-08-20
              (last checked 2026-08-20)
            </li>
            
          </ul>
        </div>
      </figure>
      
    

    <p class="ev-ts__footnote">Figures are the event counts stated in each season's own summary page, which is a secondary source rather than the PDC's own calendar listing — the publisher is named in every citation so a reader can judge it. The scale is itself a registered figure: a bar reaching the top of the plot equals the current 2026 count of 34.</p>
  </div>
</section>

$rh3$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, 'f8c2393c-fc66-480d-93f7-738c68cd5f1b', 4, 'evidence-chart-2026-calendar',
  $cd4${
 "section_eyebrow": "The 2026 calendar",
 "section_title": "What a season is made of",
 "section_intro": "The ranking circuits a professional can enter in 2026, before any televised major. The secondary-tour figure is per tour \u2014 Challenge Tour, Development Tour and Women's Series each run that many.",
 "charts": [
  {
   "id": "calendar-2026",
   "title": "Events available in the 2026 season",
   "caption": "Scale: the Players Championship count, the largest of the three.",
   "unit": "",
   "max_fact_id": "do-pdc-pc-2026",
   "points": [
    {
     "fact_id": "do-pdc-pc-2026",
     "label": "Players Championship",
     "tone": "accent"
    },
    {
     "fact_id": "do-pdc-secondary-2026",
     "label": "Each secondary tour"
    },
    {
     "fact_id": "do-pdc-et-2026",
     "label": "European Tour"
    },
    {
     "fact_id": "do-pdc-et-2025",
     "label": "European Tour (2025)",
     "tone": "muted"
    }
   ],
   "source_note": "PDC season summaries for 2025 and 2026, cited in the register per figure."
  }
 ],
 "facts": [
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
      <span class="evidence-chart__eyebrow">The 2026 calendar</span>
      <h2 class="evidence-chart__title">What a season is made of</h2>
      <p class="evidence-chart__intro">The ranking circuits a professional can enter in 2026, before any televised major. The secondary-tour figure is per tour — Challenge Tour, Development Tour and Women's Series each run that many.</p>
    </header>
    
    <div class="evidence-chart__grid">
      <figure class="evidence-chart__figure" data-chart="calendar-2026">
        <figcaption class="evidence-chart__figcaption">
          <span class="evidence-chart__chart-title">Events available in the 2026 season</span>
          <span class="evidence-chart__chart-note">Scale: the Players Championship count, the largest of the three.</span>
        </figcaption>
        
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Players Championship</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--accent" style="--v:34.0000;--m:34.0000"></span></span>
          <span class="evidence-chart__value">34</span>
          <span class="evidence-chart__verified">verified 2026-08-20</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">Each secondary tour</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:24.0000;--m:34.0000"></span></span>
          <span class="evidence-chart__value">24</span>
          <span class="evidence-chart__verified">verified 2026-08-20</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">European Tour</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar" style="--v:15.0000;--m:34.0000"></span></span>
          <span class="evidence-chart__value">15</span>
          <span class="evidence-chart__verified">verified 2026-08-20</span>
        </div>
        <div class="evidence-chart__row">
          <span class="evidence-chart__label">European Tour (2025)</span>
          <span class="evidence-chart__track" aria-hidden="true"><span class="evidence-chart__bar evidence-chart__bar--muted" style="--v:14.0000;--m:34.0000"></span></span>
          <span class="evidence-chart__value">14</span>
          <span class="evidence-chart__verified">verified 2026-08-20</span>
        </div>
        <p class="evidence-chart__source">PDC season summaries for 2025 and 2026, cited in the register per figure.</p>
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
 "content": "<p>Assembled from items our own news feed ingested \u2014 the same story arriving on more than one channel is the signal a feature exists to read. We link to the original publishers and quote headlines only.</p>\n<ul>\n<li><a href=\"https://www.dartsnews.com/\" rel=\"noopener noreferrer\">&ldquo;If you're a top player, behave like a top player&rdquo; &mdash; Barry Hearn offers daunting warning to top players skipping tournaments</a> &mdash; Dartsnews.com</li>\n<li><a href=\"https://www.oche180.com/\" rel=\"noopener noreferrer\">Littler sparks PDC concern as Euro Tour withdrawals leave organisers facing headache</a> &mdash; Oche180</li>\n<li><a href=\"https://www.dartsnews.com/\" rel=\"noopener noreferrer\">&ldquo;Then I'd be fuming&rdquo; &mdash; Van der Voort criticises PDC stars' lax attitude at World Series event in Auckland</a> &mdash; Dartsnews.com</li>\n<li><a href=\"https://www.thesun.co.uk/\" rel=\"noopener noreferrer\">Luke Littler pal admits life of darts star on PDC circuit can be &lsquo;disheartening&rsquo; in brutally honest assessment</a> &mdash; The Sun: the same story from the inside, and the reason to look at the calendar</li>\n</ul>\n<p>Calendar figures are the PDC season summaries, cited per year beneath the chart.</p>"
}$cd5$::jsonb,
  $rh5$<section id="feature-coverage" class="section section--generic">
  <div class="container">
    <h2 class="section__title">The coverage this feature draws on</h2>
    <div class="section__content"><p>Assembled from items our own news feed ingested — the same story arriving on more than one channel is the signal a feature exists to read. We link to the original publishers and quote headlines only.</p>
<ul>
<li><a href="https://www.dartsnews.com/" rel="noopener noreferrer">&ldquo;If you're a top player, behave like a top player&rdquo; &mdash; Barry Hearn offers daunting warning to top players skipping tournaments</a> &mdash; Dartsnews.com</li>
<li><a href="https://www.oche180.com/" rel="noopener noreferrer">Littler sparks PDC concern as Euro Tour withdrawals leave organisers facing headache</a> &mdash; Oche180</li>
<li><a href="https://www.dartsnews.com/" rel="noopener noreferrer">&ldquo;Then I'd be fuming&rdquo; &mdash; Van der Voort criticises PDC stars' lax attitude at World Series event in Auckland</a> &mdash; Dartsnews.com</li>
<li><a href="https://www.thesun.co.uk/" rel="noopener noreferrer">Luke Littler pal admits life of darts star on PDC circuit can be &lsquo;disheartening&rsquo; in brutally honest assessment</a> &mdash; The Sun: the same story from the inside, and the reason to look at the calendar</li>
</ul>
<p>Calendar figures are the PDC season summaries, cited per year beneath the chart.</p></div>
  </div>
</section>
$rh5$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 6, 'call-to-action',
  $cd6${
 "headline": "Throwing more this season than last?",
 "subheadline": "The pros are playing a longer calendar than they were two years ago. If your own schedule has grown too, the setup builder walks through barrel weight, grip and flight combinations so you can settle on something you can repeat under pressure.",
 "primary_cta": "Build your setup",
 "primary_cta_url": "/tools/setup-builder/index.html",
 "primary_cta_target_title": "Setup Builder | Dartsonline.com",
 "secondary_cta": "Compare dart weights",
 "secondary_cta_url": "/tools/dart-weight-comparator/index.html",
 "secondary_cta_target_title": "Dart Weight Comparator | Dartsonline.com"
}$cd6$::jsonb,
  $rh6$<section class="cta-section" data-component="call-to-action">
    <div class="cta-container">
        <h2>Throwing more this season than last?</h2>
        <p class="cta-subtitle">The pros are playing a longer calendar than they were two years ago. If your own schedule has grown too, the setup builder walks through barrel weight, grip and flight combinations so you can settle on something you can repeat under pressure.</p>
        <div class="cta-buttons">
            
            <a href="/tools/setup-builder/index.html" class="cta-btn cta-btn-primary">Build your setup</a>
            
            
            <a href="/tools/dart-weight-comparator/index.html" class="cta-btn cta-btn-secondary">Compare dart weights</a>
            
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

DO $$
DECLARE n int; short int; mismatch int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE coalesce(length(pc.rendered_html),0) < 500)
    INTO n, short FROM page_components pc JOIN _pg ON _pg.id=pc.page_id;
  IF n <> 6 THEN RAISE EXCEPTION 'expected 6 components, got %', n; END IF;
  IF short > 0 THEN RAISE EXCEPTION '% component(s) with suspiciously short rendered_html', short; END IF;
  SELECT count(*) INTO mismatch FROM pages p, jsonb_array_elements_text(p.sections) s
   WHERE p.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND p.name='darts-calendar-density'
     AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id AND pc.slot_name=s);
  IF mismatch > 0 THEN RAISE EXCEPTION '% sections entry(ies) with no matching slot_name', mismatch; END IF;
  -- the hero must be on the IMAGE branch, not the gradient fallback (owner ruling)
  IF NOT EXISTS (SELECT 1 FROM page_components pc JOIN _pg ON _pg.id=pc.page_id
                  WHERE pc.slot_name='hero'
                    AND pc.rendered_html LIKE '%content-hero-darts-calendar-density.jpg%'
                    AND pc.rendered_html LIKE '%rgba(0,0,0,0.5)%') THEN
    RAISE EXCEPTION 'hero is not on the image+overlay branch';
  END IF;
END $$;

COMMIT;
