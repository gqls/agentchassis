-- 493_robot_hands_insights_index.sql
-- The Insights section index for robot-hands.com: /insights/index.html, the
-- hub page that puts editorial features in the TOP NAV (owner ask 2026-08-20).
--
-- WHY page_type='section-index' AND NOT 'blog-post'. Nav classification bars
-- every page whose URL sits under a child prefix (/insights/, /tools/, /blog/)
-- from the primary menu — EXCEPT a section-index, which is the section's
-- PARENT rather than one of its children. That exception is keyed on
-- page_type, not URL (populate_nav_tables_action.go:377-384, isSectionIndexType
-- at v3_site_actions.go:6179-6185 = blog-index|entity-directory|section-index|
-- news-index). It also sorts tier 2 (:469-471), so it sits with the other hub
-- pages. The feature pages themselves stay in_header=false and are reached
-- from here — which is the correct shape, not a limitation.
--
-- Header cap is max_header_items=8 (default, :111). robot-hands' live header
-- carries 6 items before this one [MEASURED 2026-08-20 by curl], so Insights
-- fits without displacing anything. CHECK THIS before repeating on a site with
-- a full header — tier 3 pages are dropped silently when the cap truncates.
--
-- Locked+owned for the same reason as 492: the listing is hand-authored today.
-- Automating it needs a rebuild_insights_listing sibling of
-- rebuild_blog_listing (which queries page_type='blog-post' ONLY,
-- rebuild_blog_listing_action.go:110) — deferred deliberately, one feature.
--
-- DEPENDS ON 492. After apply: assemble-only page-rerender for the new page,
-- then a nav_drift work item so nav-updater rebuilds site_nav_items and
-- re-renders chrome (a page row alone does not put a link in the header).
--
-- ROLLBACK: DELETE FROM page_components WHERE page_id=(SELECT id FROM pages
--   WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='insights-index');
--   DELETE FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='insights-index';

\set ON_ERROR_STOP on
BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='robot-demand-step-change') THEN
    RAISE EXCEPTION 'apply 492 first - the index would list a page that does not exist';
  END IF;
  IF EXISTS (SELECT 1 FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='insights-index') THEN
    RAISE EXCEPTION 'insights-index already exists - refusing double apply';
  END IF;
END $$;

INSERT INTO pages (site_id, name, url, title, page_type, status, build_status,
                   nav_label, nav_order, in_header, in_footer, rebuild_policy,
                   meta_description, sections, topics)
VALUES ('00ff3af5-dad8-4770-9f70-3edc267a3c92', 'insights-index', '/insights/index.html',
  'Insights | Editorial features on industrial automation | Robot-Hands.com',
  'section-index', 'active', 'pending',
  'Insights', 35, true, true, 'owned',
  'Editorial features on the stories moving industrial automation, each built from several sources and charted from cited figures.',
  '["hero", "info-card-grid"]'::jsonb,
  ARRAY['insights','editorial features','industrial automation']);

CREATE TEMP TABLE _pg ON COMMIT DROP AS
  SELECT id FROM pages WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND name='insights-index';

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data,
                             rendered_html, build_status, locked_at, locked_by, lock_type)
SELECT pg.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 1, 'hero',
  $cd1${
 "headline": "Insights",
 "subheadline": "Editorial features on the stories moving industrial automation \u2014 each one built from several sources at once, with the background charted from figures we have checked and cited. Updated while a story is live; kept here afterwards.",
 "cta_text": "Run MatchMatrix",
 "cta_url": "/tools/matchmatrix/index.html",
 "cta_target_title": "Run MatchMatrix | Gripper Selection Tool | Robot-Hands.com"
}$cd1$::jsonb,
  $rh1$<section class="hero" data-component="hero" style="--hero-ink: var(--color-primary-text); background: var(--color-primary); background: linear-gradient(135deg, var(--color-primary) 0%, color-mix(in srgb, var(--color-primary) 85%, var(--color-primary-text)) 100%);">
        <div class="hero-content">
            <h1>Insights</h1>
            <p class="hero-subheadline">Editorial features on the stories moving industrial automation — each one built from several sources at once, with the background charted from figures we have checked and cited. Updated while a story is live; kept here afterwards.</p>
            <a href="/tools/matchmatrix/index.html" class="btn btn-primary">Run MatchMatrix</a>
            
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
SELECT pg.id, 'fc56f085-8e9a-4f6b-8e8d-600f9a1381e2', 2, 'info-card-grid',
  $cd2${
 "ComponentID": "insights-index-grid",
 "section_eyebrow": "Features",
 "section_title": "Current features",
 "section_subtitle": "Each feature reads one story across several channels and charts the background behind it. Figures resolve to cited sources; charts are drawn from the data, never sketched.",
 "cards": [
  {
   "icon": "\ud83d\udcc8",
   "title": "Half a Million Robots a Year: Reading the Demand Story Properly",
   "body": "Record installations, rising orders and one dissenting headline about weak US demand \u2014 all true at once. Five years of global installation figures show what the coverage does not: a single sharp step up, then a plateau at altitude.",
   "link_url": "/insights/robot-demand-step-change.html",
   "link_label": "Read the feature"
  }
 ]
}$cd2$::jsonb,
  $rh2$<style>
  .info-card-grid-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }

  .info-card-grid-section .info-card-grid__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }

  .info-card-grid-section .info-card-grid__header {
    text-align: center;
    margin-bottom: 3rem;
  }

  .info-card-grid-section .info-card-grid__eyebrow {
    display: inline-block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary-ink, var(--color-primary));
    margin-bottom: 0.75rem;
  }

  .info-card-grid-section .info-card-grid__title {
    font-size: clamp(1.75rem, 3vw, 2.5rem);
    font-weight: 700;
    color: var(--color-heading);
    margin: 0 0 1rem;
    line-height: 1.2;
  }

  .info-card-grid-section .info-card-grid__subtitle {
    font-size: 1.0625rem;
    color: var(--color-text-muted);
    max-width: 640px;
    margin: 0 auto;
    line-height: 1.6;
  }

  .info-card-grid-section .info-card-grid__grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .info-card-grid-section .info-card-grid__card {
    background: var(--color-card-bg, var(--color-surface));
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius, 0.5rem);
    padding: 2rem 1.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    box-shadow: var(--shadow, 0 2px 8px rgba(0,0,0,0.06));
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .info-card-grid-section .info-card-grid__card:hover {
    transform: translateY(-3px);
    box-shadow: var(--shadow, 0 6px 20px rgba(0,0,0,0.1));
  }

  .info-card-grid-section .info-card-grid__card-icon {
    width: 2.75rem;
    height: 2.75rem;
    background: var(--color-surface-alt, var(--color-surface));
    border-radius: var(--border-radius, 0.5rem);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.375rem;
    flex-shrink: 0;
    border: 1px solid var(--color-hairline, var(--color-border));
    overflow: hidden;
  }

  /* An icon IMAGE is line art drawn dark-on-light, so its chip is painted
     light and the artwork sits on it as a chip. This is a matched pair — the
     literal is chosen for the artwork it holds, not inherited from a theme
     that may be dark. Cards with only an emoji `icon` are untouched. */
  .info-card-grid-section .info-card-grid__card-icon--img {
    background: var(--color-icon-chip-bg, #EEF2F8);
    padding: 0.3rem;
  }
  .info-card-grid-section .info-card-grid__card-icon-img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    display: block;
  }

  .info-card-grid-section .info-card-grid__card-title {
    font-size: 1.0625rem;
    font-weight: 700;
    color: var(--color-heading);
    margin: 0;
    line-height: 1.3;
  }

  .info-card-grid-section .info-card-grid__card-body {
    font-size: 0.9375rem;
    color: var(--color-text-muted);
    line-height: 1.65;
    margin: 0;
    flex-grow: 1;
  }

  .info-card-grid-section .info-card-grid__card-link {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-primary-ink, var(--color-primary));
    text-decoration: none;
    min-height: 44px;
    padding: 0.25rem 0;
    transition: color 0.2s ease;
    margin-top: auto;
  }

  .info-card-grid-section .info-card-grid__card-link:hover {
    color: var(--color-primary-hover, var(--color-primary));
    text-decoration: underline;
  }

  .info-card-grid-section .info-card-grid__card-link:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
    border-radius: 2px;
  }

  .info-card-grid-section .info-card-grid__card-link-arrow {
    font-style: normal;
    transition: transform 0.2s ease;
  }

  .info-card-grid-section .info-card-grid__card-link:hover .info-card-grid__card-link-arrow {
    transform: translateX(3px);
  }

  @media (max-width: 768px) {
    .info-card-grid-section {
      padding: 3rem 1.25rem;
    }

    .info-card-grid-section .info-card-grid__grid {
      grid-template-columns: 1fr;
      gap: 1rem;
    }

    .info-card-grid-section .info-card-grid__card {
      padding: 1.5rem 1.25rem;
    }

    .info-card-grid-section .info-card-grid__header {
      margin-bottom: 2rem;
    }
  }
</style>

<section class="info-card-grid-section" data-component="info-card-grid">
  <div class="info-card-grid__inner">
    <header class="info-card-grid__header">
      <span class="info-card-grid__eyebrow">Features</span>
      <h2 class="info-card-grid__title">Current features</h2>
      <p class="info-card-grid__subtitle">Each feature reads one story across several channels and charts the background behind it. Figures resolve to cited sources; charts are drawn from the data, never sketched.</p>
    </header>
    <div class="info-card-grid__grid">
      
      <article class="info-card-grid__card">
        <div class="info-card-grid__card-icon" aria-hidden="true">📈</div>
        <h3 class="info-card-grid__card-title">Half a Million Robots a Year: Reading the Demand Story Properly</h3>
        <p class="info-card-grid__card-body">Record installations, rising orders and one dissenting headline about weak US demand — all true at once. Five years of global installation figures show what the coverage does not: a single sharp step up, then a plateau at altitude.</p>
        <a class="info-card-grid__card-link" href="/insights/robot-demand-step-change.html">
          Read the feature
          <em class="info-card-grid__card-link-arrow" aria-hidden="true">&rarr;</em>
        </a>
      </article>
      
    </div>
  </div>
</section>


$rh2$,
  'deployed', now(), 'news_editorial_features-lane', 'permanent'
FROM _pg pg;

DO $$
DECLARE n int; short int; mismatch int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE coalesce(length(pc.rendered_html),0) < 500)
    INTO n, short FROM page_components pc JOIN _pg ON _pg.id=pc.page_id;
  IF n <> 2 THEN RAISE EXCEPTION 'expected 2 components, got %', n; END IF;
  IF short > 0 THEN RAISE EXCEPTION '% component(s) with suspiciously short rendered_html', short; END IF;
  SELECT count(*) INTO mismatch FROM pages p, jsonb_array_elements_text(p.sections) s
   WHERE p.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND p.name='insights-index'
     AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id AND pc.slot_name=s);
  IF mismatch > 0 THEN RAISE EXCEPTION '% sections entry(ies) with no matching slot_name', mismatch; END IF;
  -- the listing must actually link the feature, or the hub is decorative
  IF NOT EXISTS (SELECT 1 FROM page_components pc JOIN _pg ON _pg.id=pc.page_id
                  WHERE pc.rendered_html LIKE '%/insights/robot-demand-step-change.html%') THEN
    RAISE EXCEPTION 'listing does not link the feature page';
  END IF;
END $$;

COMMIT;
