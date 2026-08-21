-- 547_split_shared_professional_dark_theme_per_site.sql
--
-- bugs_open/198 — OWNER DECISION 2026-08-21: give finetuning.uk and
-- gaswholesalers.com a theme row each.
--
-- THE PROBLEM THIS RESOLVES. Both sites pointed at ONE style_collection
-- (`3196d966`, "professional-dark") which pointed at ONE css_themes row
-- (`fecb962d`, origin='seed', 1,649 bytes holding a bare `:root` palette block and
-- no layout rules). They serve DIFFERENT stylesheets — 13,988 and 20,271 bytes —
-- so no backfill could ever make that single row true for both: seeding it from
-- either site's file would push that site's CSS onto the other. The fleet-wide
-- backfill of 2026-08-21 correctly skipped them for exactly this reason, and
-- migration 542's `site_count <= 1` gate has been REFUSING to patch either site
-- since it applied. This migration removes the cause rather than the symptom.
--
-- WHAT IT DOES. Creates a per-site collection and theme for each of the two sites
-- — the same shape every other live site already has (`collection-<domain>` /
-- `theme-<domain>`, origin='adopted', source_domain set) — and repoints
-- `sites.style_collection_id`. The seed `professional-dark` collection and theme
-- are left EXACTLY as they are: they are library assets, not site state, and a
-- future site may still adopt them. Nothing is deleted and nothing is mutated
-- except the two `sites` pointers.
--
-- THE DESIGN DOES NOT CHANGE, and that is the point of copying the FKs. What
-- actually renders `assets/css/styles.css` is `render_css_from_spec`, which reads
-- the palette / layout / typography_set FK'd rows — NOT `css_content`. Both new
-- themes carry the seed's exact three FKs (palette `3ce8a4e4`, layout `a9001f12`,
-- typography `31fc3a77`), so the next webdesign-agent run on either site produces
-- what it produces today. `forked_from_theme_id` records the lineage honestly.
--
-- The collections likewise copy the seed's `header_component_id` /
-- `footer_component_id` chrome pins. Omitting them would silently drop both sites'
-- header and footer — the `bugs_closed/170` trap ("a style_collection pins chrome,
-- and a fork copies it"), here in its mirror image: a fork that FAILS to copy.
--
-- SEEDING THE ROWS. Each new theme's `css_content` is that site's OWN currently
-- served stylesheet, fetched 2026-08-21 with a cache-buster and embedded below.
-- That is the right source because it is what the site actually serves, so the row
-- equals the file — the exact invariant that makes css-patch-agent's
-- "deploy the whole row" safe, and the one migration 543 now maintains at every
-- render. Both files were checked before embedding: 4 `:root` blocks each,
-- `--color-primary-ink` present, and the two-clause stale-ink test
-- (`<x>-ink == --color-text AND <x>-ink != --color-<x>`) clean on both slots of
-- both sites — so neither carries a pre-2026-08-14 derivation, which the bug
-- file's restore procedure requires checking before any seed.
--
-- AFTER THIS, BOTH SITES PASS THE 542 GATE: site_count becomes 1 and css_len goes
-- 1,649 -> 13,988 / 20,271, i.e. above the 4096-byte floor. Their parked contrast
-- items become unparkable via the RUNBOOK sweep.
--
-- CONFIG/DATA IS LIVE IMMEDIATELY ON APPLY.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM css_themes WHERE name IN ('theme-finetuning-uk','theme-gaswholesalers-com')) THEN
        RAISE EXCEPTION '198/547: already applied — a per-site theme already exists';
    END IF;
END $$;

BEGIN;

-- ── DRIFT GUARD: assert the exact shared state this migration is splitting ──────
DO $$
DECLARE
    v_sites int;
    v_theme_bytes int;
BEGIN
    SELECT count(*) INTO v_sites FROM sites WHERE style_collection_id = '3196d966-24ef-4415-9dc8-1afbc02166ca';
    IF v_sites <> 2 THEN
        RAISE EXCEPTION '198/547 drift: collection 3196d966 now serves % sites, expected exactly 2', v_sites;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM sites WHERE domain='finetuning.uk' AND style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca')
       OR NOT EXISTS (SELECT 1 FROM sites WHERE domain='gaswholesalers.com' AND style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca') THEN
        RAISE EXCEPTION '198/547 drift: the two sites are no longer both on collection 3196d966';
    END IF;
    SELECT octet_length(css_content) INTO v_theme_bytes FROM css_themes WHERE id='fecb962d-3ace-4c19-b08f-088eba46ea53';
    IF v_theme_bytes <> 1649 THEN
        RAISE EXCEPTION '198/547 drift: seed theme fecb962d is now % bytes, expected 1649 — somebody has written to it', v_theme_bytes;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM css_themes WHERE id='fecb962d-3ace-4c19-b08f-088eba46ea53'
                     AND palette_id='3ce8a4e4-8ae9-44cd-9360-ff690c437495'
                     AND layout_id='a9001f12-df09-4571-b04c-644553fe2c09'
                     AND typography_set_id='31fc3a77-2093-40b4-bfd2-96c161e1a7ff') THEN
        RAISE EXCEPTION '198/547 drift: seed theme composition FKs have moved — the copies below would change the sites design';
    END IF;
END $$;


-- ── finetuning.uk ──────────────────────────────────────────────────────────────
INSERT INTO css_themes (
    name, display_name, css_content, css_template,
    palette_id, layout_id, typography_set_id,
    origin, needs_review, forked_from_theme_id,
    source_site_id, source_domain, forked_at, version, is_active
) VALUES (
    'theme-finetuning-uk',
    'Composition for finetuning.uk',
    $CSS_FINETUNING_UK$
/* =====================================================================
 * LAYOUT: brochure-formal
 * Variables consumed via map-based template helpers:
 *   #hex     palette slot lookup with fallback
 *   default  typography slot lookup with fallback
 *   default  structure token lookup with fallback
 *
 * Renderer contract:
 *   - This template MUST NOT declare --section-* defaults; the renderer
 *     appends them after rendering based on palette luminance.
 *   - Renderer-managed surface section classes:
 *       .differentiators-section, .features-section, .faq-section,
 *       .services-section, .about-section
 *     These MUST be surface-coloured here.
 *   - Element rules use var(--section-*, var(--color-*)) so dark-section
 *     components can override per-container without restating rules.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        #1A1A2E;
  --color-primary-hover:  #1e3a8a;
  --color-primary-text:   #ffffff;
  --color-secondary:      #E8E4DC;
  --color-accent:         #C8873A;
  --color-background:     #F5F3EF;
  --color-surface:        #FFFFFF;
  --color-text:           #1A1A2E;
  --color-text-muted:     #6B6860;
  --color-border:         #D4CFC6;
  --color-card-bg:        #ffffff;
  --color-header-bg:      #0f172a;
  --color-header-text:    #f1f5f9;
  --color-cta-bg:         linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%);
  --color-cta-text:       #ffffff;
  --color-footer-bg:      #0f172a;
  --color-footer-text:    #cbd5e1;

  /* ── Typography ── */
  --font-body:        'Inter', 'DM Sans', system-ui, -apple-system, sans-serif;
  --font-heading:     'DM Sans', 'Inter', system-ui, sans-serif;
  --font-size-base:   16px;
  --line-height-base: 1.6;

  /* ── Structure ── */
  --container-max:    1200px;
  --container-pad-x:  1.5rem;
  --section-pad-y:    5rem;
  --section-pad-y-sm: 3rem;
  --radius:           0.375rem;
  --radius-sm:        0.25rem;
  --radius-lg:        0.5rem;
  --shadow-sm:        0 1px 2px rgba(0,0,0,0.05);
  --shadow-md:        0 4px 6px rgba(0,0,0,0.07);
  --shadow-lg:        0 10px 15px rgba(0,0,0,0.1);
  --transition:       200ms ease;
  --card-pad:         1.5rem;
  --grid-gap:         2rem;

  /* ── Optional palette-driven section overrides ──
   * If the palette declares a 'heading' slot, we emit --section-heading at
   * :root level. Per the Colour Inheritance Model contract, h1-h6 resolve
   * via var(--section-heading, var(--color-primary)). Setting it here makes
   * the palette's heading choice apply in light sections; dark-section
   * components still override it on their container per the Dark Section
   * Variable Contract. Palettes without a 'heading' slot fall through to
   * --color-primary as the contract specifies. */
  --section-heading: #0f172a;
}

/* ── Base reset ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; }
body {
  margin: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: var(--font-body);
  font-size: var(--font-size-base);
  line-height: var(--line-height-base);
  color: var(--color-text);
  background: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* ── Colour Inheritance Model ──
 * Element rules use var(--section-*, var(--color-*)) so dark-section
 * components can override per container without restating rules. */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 1rem;
  line-height: 1.25;
  font-weight: 700;
}
h1 { font-size: 2.5rem; }
h2 { font-size: 2rem; }
h3 { font-size: 1.5rem; font-weight: 600; }
h4 { font-size: 1.25rem; font-weight: 600; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
/* strong/em/cite/span: do NOT set color — they inherit from parent */
a {
  color: var(--color-accent-ink, var(--color-accent));
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-header-text);
  text-decoration: none;
}
.logo-img { max-height: 40px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 2rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.95rem;
  padding: 0.5rem 0;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-accent); }
.mobile-menu-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  min-width: 44px;
  min-height: 44px;
  color: var(--color-header-text);
}
.mobile-menu-toggle span {
  display: block;
  width: 22px;
  height: 2px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero (component-coloured) ── */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.2); }
.hero-section .container { text-align: center; max-width: 880px; }
.hero-section h1 { font-size: clamp(2rem, 4vw, 3rem); margin-bottom: 1rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.25rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 640px;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync
 * with buildSectionDefaults in render_css_from_spec_action.go, which
 * hardcodes the same list to emit dark-surface --section-* overrides
 * when surface is dark.
 *
 * LONG-TERM DIRECTION: move surface painting into the relevant
 * components (features, services, differentiators, about, faq), change
 * the renderer to emit overrides keyed on a data-section-bg attribute
 * instead of hardcoded class names, and remove this block from every
 * layout. Tracked as Phase 4.5 in 025_palette_layout_typography_migration.
 *
 * Until then, these 5 classes MUST be surface-coloured in every layout
 * so the renderer's assumption holds. Hero/CTA/testimonials/contact are
 * NOT in this list — their background is component-owned per the Dark
 * Section Variable Contract. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Generic section grids (3-col → 2 → 1) */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--grid-gap);
  margin-top: 3rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-md);
}
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 1rem;
  color: var(--color-accent);
}

/* ── About ── */
.about-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: center;
}

/* ── FAQ ── */
.faq-section .container { max-width: 820px; }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1.25rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-weight: 600;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; font-weight: 400; font-size: 1.25rem; }
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Call to action (component-coloured) ── */
.call-to-action-section { text-align: center; }

/* ── Testimonials (component-coloured) ── */
.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.125rem;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1rem;
  font-style: normal;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.95rem;
}

/* ── Contact ── */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

/* ── Forms ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.9rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.75rem;
  font: inherit;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 44px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 20%, transparent);
}

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-weight: 500;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition),
              color var(--transition), box-shadow var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-large { padding: 1rem 2rem; font-size: 1.05rem; }

/* ── Site footer ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding-top: 4rem;
  margin-top: auto;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 { color: #ffffff; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255,255,255,0.6);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  h1 { font-size: 2rem; }
  h2 { font-size: 1.625rem; }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }
}

/* ── Accessibility ── */
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}


/* === Component-specific styles === */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.fade-in-up { animation: fadeInUp 0.6s ease forwards; }

.grid-responsive {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

/* renderer-enforced compatibility aliases (component vocabulary bridge) */
:root {
  --border-radius: var(--radius, 8px);
  --shadow: var(--shadow-md, 0 2px 12px rgba(0, 0, 0, 0.15));
  --spacing-section: var(--section-pad-y, 4rem);
  --container-max-width: var(--container-max, 1200px);
  --primary-color: var(--color-primary);
  --secondary-color: var(--color-secondary);
  --accent-color: var(--color-accent);
  --color-heading: var(--color-text);
  --color-white: #ffffff;
  --color-error: #d64545;
  --hero-ink: var(--color-text);
}


/* renderer-owned legible-ink companions. Opt in with
   var(--color-primary-ink, var(--color-primary)) — never bare, so a
   stylesheet rendered before this block renders byte-identically. */
:root {
  --color-primary-ink: #1A1A2E;
  --color-accent-ink: #8a5c27;
  --color-accent-text: #1A1A2E;
}
$CSS_FINETUNING_UK$,
    '',
    '3ce8a4e4-8ae9-44cd-9360-ff690c437495',
    'a9001f12-df09-4571-b04c-644553fe2c09',
    '31fc3a77-2093-40b4-bfd2-96c161e1a7ff',
    'adopted', false,
    'fecb962d-3ace-4c19-b08f-088eba46ea53',
    (SELECT id FROM sites WHERE domain='finetuning.uk'),
    'finetuning.uk', NOW(), 1, true
);

INSERT INTO style_collections (
    name, display_name, description,
    header_component_id, header_home_component_id, footer_component_id,
    css_theme_id, color_palette, typography, category, industry_tags,
    is_active, origin, needs_review, forked_from_collection_id,
    source_site_id, source_domain, forked_at
)
SELECT
    'collection-finetuning-uk',
    'Composition for finetuning.uk',
    sc.description,
    sc.header_component_id, sc.header_home_component_id, sc.footer_component_id,
    (SELECT id FROM css_themes WHERE name='theme-finetuning-uk'),
    sc.color_palette, sc.typography, sc.category, sc.industry_tags,
    true, 'adopted', false, sc.id,
    (SELECT id FROM sites WHERE domain='finetuning.uk'),
    'finetuning.uk', NOW()
FROM style_collections sc
WHERE sc.id = '3196d966-24ef-4415-9dc8-1afbc02166ca';

UPDATE sites
   SET style_collection_id = (SELECT id FROM style_collections WHERE name='collection-finetuning-uk'),
       updated_at = NOW()
 WHERE domain = 'finetuning.uk'
   AND style_collection_id = '3196d966-24ef-4415-9dc8-1afbc02166ca';


-- ── gaswholesalers.com ──────────────────────────────────────────────────────────────
INSERT INTO css_themes (
    name, display_name, css_content, css_template,
    palette_id, layout_id, typography_set_id,
    origin, needs_review, forked_from_theme_id,
    source_site_id, source_domain, forked_at, version, is_active
) VALUES (
    'theme-gaswholesalers-com',
    'Composition for gaswholesalers.com',
    $CSS_GASWHOLESALERS_COM$
/* =====================================================================
 * LAYOUT: brochure-formal
 * Variables consumed via map-based template helpers:
 *   #hex     palette slot lookup with fallback
 *   default  typography slot lookup with fallback
 *   default  structure token lookup with fallback
 *
 * Renderer contract:
 *   - This template MUST NOT declare --section-* defaults; the renderer
 *     appends them after rendering based on palette luminance.
 *   - Renderer-managed surface section classes:
 *       .differentiators-section, .features-section, .faq-section,
 *       .services-section, .about-section
 *     These MUST be surface-coloured here.
 *   - Element rules use var(--section-*, var(--color-*)) so dark-section
 *     components can override per-container without restating rules.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        #1A1A2E;
  --color-primary-hover:  #1e3a8a;
  --color-primary-text:   #ffffff;
  --color-secondary:      #C8880A;
  --color-accent:         #E8A020;
  --color-background:     #F4F1EB;
  --color-surface:        #FFFFFF;
  --color-text:           #1C1C1C;
  --color-text-muted:     #5C5C5C;
  --color-border:         #D6CFC2;
  --color-card-bg:        #ffffff;
  --color-header-bg:      #0f172a;
  --color-header-text:    #f1f5f9;
  --color-cta-bg:         linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%);
  --color-cta-text:       #ffffff;
  --color-footer-bg:      #0f172a;
  --color-footer-text:    #cbd5e1;

  /* ── Typography ── */
  --font-body:        'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
  --font-heading:     'IBM Plex Sans Condensed', 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
  --font-size-base:   16px;
  --line-height-base: 1.6;

  /* ── Structure ── */
  --container-max:    1200px;
  --container-pad-x:  1.5rem;
  --section-pad-y:    5rem;
  --section-pad-y-sm: 3rem;
  --radius:           0.375rem;
  --radius-sm:        0.25rem;
  --radius-lg:        0.5rem;
  --shadow-sm:        0 1px 2px rgba(0,0,0,0.05);
  --shadow-md:        0 4px 6px rgba(0,0,0,0.07);
  --shadow-lg:        0 10px 15px rgba(0,0,0,0.1);
  --transition:       200ms ease;
  --card-pad:         1.5rem;
  --grid-gap:         2rem;

  /* ── Optional palette-driven section overrides ──
   * If the palette declares a 'heading' slot, we emit --section-heading at
   * :root level. Per the Colour Inheritance Model contract, h1-h6 resolve
   * via var(--section-heading, var(--color-primary)). Setting it here makes
   * the palette's heading choice apply in light sections; dark-section
   * components still override it on their container per the Dark Section
   * Variable Contract. Palettes without a 'heading' slot fall through to
   * --color-primary as the contract specifies. */
  --section-heading: #0f172a;
}

/* ── Base reset ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; }
body {
  margin: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: var(--font-body);
  font-size: var(--font-size-base);
  line-height: var(--line-height-base);
  color: var(--color-text);
  background: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* ── Colour Inheritance Model ──
 * Element rules use var(--section-*, var(--color-*)) so dark-section
 * components can override per container without restating rules. */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 1rem;
  line-height: 1.25;
  font-weight: 700;
}
h1 { font-size: 2.5rem; }
h2 { font-size: 2rem; }
h3 { font-size: 1.5rem; font-weight: 600; }
h4 { font-size: 1.25rem; font-weight: 600; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
/* strong/em/cite/span: do NOT set color — they inherit from parent */
a {
  color: var(--color-accent-ink, var(--color-accent));
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-header-text);
  text-decoration: none;
}
.logo-img { max-height: 40px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 2rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.95rem;
  padding: 0.5rem 0;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-accent); }
.mobile-menu-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  min-width: 44px;
  min-height: 44px;
  color: var(--color-header-text);
}
.mobile-menu-toggle span {
  display: block;
  width: 22px;
  height: 2px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero (component-coloured) ── */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.2); }
.hero-section .container { text-align: center; max-width: 880px; }
.hero-section h1 { font-size: clamp(2rem, 4vw, 3rem); margin-bottom: 1rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.25rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 640px;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync
 * with buildSectionDefaults in render_css_from_spec_action.go, which
 * hardcodes the same list to emit dark-surface --section-* overrides
 * when surface is dark.
 *
 * LONG-TERM DIRECTION: move surface painting into the relevant
 * components (features, services, differentiators, about, faq), change
 * the renderer to emit overrides keyed on a data-section-bg attribute
 * instead of hardcoded class names, and remove this block from every
 * layout. Tracked as Phase 4.5 in 025_palette_layout_typography_migration.
 *
 * Until then, these 5 classes MUST be surface-coloured in every layout
 * so the renderer's assumption holds. Hero/CTA/testimonials/contact are
 * NOT in this list — their background is component-owned per the Dark
 * Section Variable Contract. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Generic section grids (3-col → 2 → 1) */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--grid-gap);
  margin-top: 3rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-md);
}
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 1rem;
  color: var(--color-accent);
}

/* ── About ── */
.about-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: center;
}

/* ── FAQ ── */
.faq-section .container { max-width: 820px; }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1.25rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-weight: 600;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; font-weight: 400; font-size: 1.25rem; }
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Call to action (component-coloured) ── */
.call-to-action-section { text-align: center; }

/* ── Testimonials (component-coloured) ── */
.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.125rem;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1rem;
  font-style: normal;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.95rem;
}

/* ── Contact ── */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

/* ── Forms ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.9rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.75rem;
  font: inherit;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 44px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 20%, transparent);
}

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-weight: 500;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition),
              color var(--transition), box-shadow var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-large { padding: 1rem 2rem; font-size: 1.05rem; }

/* ── Site footer ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding-top: 4rem;
  margin-top: auto;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 { color: #ffffff; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255,255,255,0.6);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  h1 { font-size: 2rem; }
  h2 { font-size: 1.625rem; }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }
}

/* ── Accessibility ── */
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}


/* === Component-specific styles === */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.fade-in-up { animation: fadeInUp 0.6s ease forwards; }


/* Latest news section — homepage card grid.
   Uses theme CSS custom properties with fallbacks. */
 
.latest-news-section {
  padding: 5rem 2rem;
  background: var(--color-background, #f8fafc);
}
.latest-news-section .container { max-width: 1200px; margin: 0 auto; }
 
.latest-news-section .section-heading {
  font-size: clamp(1.75rem, 3.5vw, 2.5rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.15;
  color: var(--color-heading, #0f172a);
  margin: 0 0 1rem;
  position: relative;
  padding-top: 1.5rem;
}
.latest-news-section .section-heading::before {
  content: "";
  position: absolute;
  top: 0; left: 0;
  width: 2.5rem; height: 3px;
  background: var(--color-accent, #d97706);
  border-radius: 2px;
}
 
.latest-news-section .section-subheadline {
  font-size: 1.125rem;
  line-height: 1.6;
  color: var(--color-text-muted, #64748b);
  max-width: 64ch;
  margin: 0 0 3rem;
}
 
.latest-news-section .news-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.75rem;
}
.latest-news-section .news-empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 2rem;
  color: var(--color-text-muted, #64748b);
}
 
.news-card {
  background: var(--color-card-bg, #ffffff);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: 0.5rem;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.news-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  border-color: var(--color-accent, #d97706);
}
.news-card-content {
  padding: 1.5rem 1.5rem 1.25rem;
  display: flex; flex-direction: column;
  gap: 0.75rem; height: 100%;
}
 
.news-card-title { font-size: 1.125rem; font-weight: 600; line-height: 1.4; margin: 0; }
.news-card-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-card-title a:hover { color: var(--color-accent, #d97706); }
 
.news-card-summary {
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--color-text, #475569);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
 
.news-card-meta {
  display: flex; align-items: center;
  font-size: 0.8125rem;
  color: var(--color-text-muted, #64748b);
  margin-top: auto;
  padding-top: 0.5rem;
}
.news-card-meta .news-source { font-weight: 500; }
.news-card-meta .news-source::after {
  content: "·"; display: inline-block; margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1); font-weight: 400;
}
.news-card-meta .news-date { font-variant-numeric: tabular-nums; }
 
.news-section-footer { margin-top: 3rem; text-align: center; }
.news-more-link {
  display: inline-flex; align-items: center; gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 1rem; font-weight: 600;
  color: var(--color-accent, #d97706);
  text-decoration: none;
  border: 1.5px solid var(--color-accent, #d97706);
  border-radius: 0.375rem;
  transition: background 0.15s ease, color 0.15s ease;
}
.news-more-link:hover { background: var(--color-accent, #d97706); color: #ffffff; }
 
@media (max-width: 768px) {
  .latest-news-section { padding: 3.5rem 1.5rem; }
  .latest-news-section .news-grid { grid-template-columns: 1fr; gap: 1.25rem; }
  .latest-news-section .section-subheadline { margin-bottom: 2rem; }
}



/* News listing page — full archive, long-form reading. */
 
.news-listing-section {
  padding: 5rem 2rem;
  background: var(--color-background, #f8fafc);
}
.news-listing-container { max-width: 760px; margin: 0 auto; }
 
.news-listing-header {
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 2px solid var(--color-border, #e2e8f0);
}
.news-listing-title {
  font-size: clamp(2rem, 4vw, 2.75rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.15;
  color: var(--color-heading, #0f172a);
  margin: 0 0 1rem;
}
.news-listing-subtitle {
  font-size: 1.125rem;
  line-height: 1.6;
  color: var(--color-text-muted, #64748b);
  margin: 0;
  max-width: 64ch;
}
 
.news-listing-items { display: flex; flex-direction: column; gap: 0; }
.news-listing-loading,
.news-listing-empty {
  padding: 3rem 0;
  text-align: center;
  color: var(--color-text-muted, #64748b);
}
 
.news-list-item {
  padding: 2rem 0;
  border-bottom: 1px solid var(--color-border, #e2e8f0);
}
.news-list-item:last-child { border-bottom: none; }
.news-list-item-content { display: flex; flex-direction: column; gap: 0.75rem; }
 
.news-list-item-title { font-size: 1.375rem; font-weight: 600; line-height: 1.35; margin: 0; }
.news-list-item-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-list-item-title a:hover { color: var(--color-accent, #d97706); }
 
.news-list-item-summary {
  font-size: 1rem;
  line-height: 1.65;
  color: var(--color-text, #475569);
  margin: 0;
}
 
.news-list-item-meta {
  display: flex; align-items: center;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
  margin-top: 0.25rem;
}
.news-list-item-source { font-weight: 500; }
.news-list-item-source::after {
  content: "·"; display: inline-block; margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1); font-weight: 400;
}
.news-list-item-date { font-variant-numeric: tabular-nums; }
 
.news-list-item-topics {
  display: flex; flex-wrap: wrap; gap: 0.5rem;
  margin-top: 0.5rem;
}
.news-list-tag {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem; font-weight: 500;
  color: var(--color-text-muted, #64748b);
  background: var(--color-border, #e2e8f0);
  border-radius: 999px;
  letter-spacing: 0.02em;
}
 
.news-listing-footer {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid var(--color-border, #e2e8f0);
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 1rem;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
}
.news-listing-count, .news-listing-updated { margin: 0; }
 
@media (max-width: 768px) {
  .news-listing-section { padding: 3rem 1.25rem; }
  .news-listing-header { margin-bottom: 2rem; padding-bottom: 1.5rem; }
  .news-list-item { padding: 1.5rem 0; }
  .news-list-item-title { font-size: 1.1875rem; }
}


.grid-responsive {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

/* renderer-enforced compatibility aliases (component vocabulary bridge) */
:root {
  --border-radius: var(--radius, 8px);
  --shadow: var(--shadow-md, 0 2px 12px rgba(0, 0, 0, 0.15));
  --spacing-section: var(--section-pad-y, 4rem);
  --container-max-width: var(--container-max, 1200px);
  --primary-color: var(--color-primary);
  --secondary-color: var(--color-secondary);
  --accent-color: var(--color-accent);
  --color-heading: var(--color-text);
  --color-white: #ffffff;
  --color-error: #d64545;
  --hero-ink: var(--color-text);
}


/* renderer-owned legible-ink companions. Opt in with
   var(--color-primary-ink, var(--color-primary)) — never bare, so a
   stylesheet rendered before this block renders byte-identically. */
:root {
  --color-primary-ink: #1A1A2E;
  --color-accent-ink: #8a5d0e;
  --color-accent-text: #1C1C1C;
}
$CSS_GASWHOLESALERS_COM$,
    '',
    '3ce8a4e4-8ae9-44cd-9360-ff690c437495',
    'a9001f12-df09-4571-b04c-644553fe2c09',
    '31fc3a77-2093-40b4-bfd2-96c161e1a7ff',
    'adopted', false,
    'fecb962d-3ace-4c19-b08f-088eba46ea53',
    (SELECT id FROM sites WHERE domain='gaswholesalers.com'),
    'gaswholesalers.com', NOW(), 1, true
);

INSERT INTO style_collections (
    name, display_name, description,
    header_component_id, header_home_component_id, footer_component_id,
    css_theme_id, color_palette, typography, category, industry_tags,
    is_active, origin, needs_review, forked_from_collection_id,
    source_site_id, source_domain, forked_at
)
SELECT
    'collection-gaswholesalers-com',
    'Composition for gaswholesalers.com',
    sc.description,
    sc.header_component_id, sc.header_home_component_id, sc.footer_component_id,
    (SELECT id FROM css_themes WHERE name='theme-gaswholesalers-com'),
    sc.color_palette, sc.typography, sc.category, sc.industry_tags,
    true, 'adopted', false, sc.id,
    (SELECT id FROM sites WHERE domain='gaswholesalers.com'),
    'gaswholesalers.com', NOW()
FROM style_collections sc
WHERE sc.id = '3196d966-24ef-4415-9dc8-1afbc02166ca';

UPDATE sites
   SET style_collection_id = (SELECT id FROM style_collections WHERE name='collection-gaswholesalers-com'),
       updated_at = NOW()
 WHERE domain = 'gaswholesalers.com'
   AND style_collection_id = '3196d966-24ef-4415-9dc8-1afbc02166ca';


-- ── VERIFY ─────────────────────────────────────────────────────────────────────
DO $$
DECLARE
    r RECORD;
    v_expect_md5 text;
BEGIN
    FOR r IN
        SELECT s.domain, ct.id AS theme_id, ct.name AS theme, octet_length(ct.css_content) AS bytes,
               md5(ct.css_content) AS content_md5, sc.header_component_id, sc.footer_component_id,
               (SELECT count(*) FROM sites s2 JOIN style_collections sc2 ON s2.style_collection_id = sc2.id
                 WHERE sc2.css_theme_id = ct.id) AS site_count
        FROM sites s
        JOIN style_collections sc ON sc.id = s.style_collection_id
        JOIN css_themes ct ON ct.id = sc.css_theme_id
        WHERE s.domain IN ('finetuning.uk','gaswholesalers.com')
    LOOP
        v_expect_md5 := CASE r.domain
            WHEN 'finetuning.uk' THEN 'e3edb78e1ebe802ff1dbfe1f389cc3f5'
            WHEN 'gaswholesalers.com' THEN '2cdf5f76897ccafcc297513506fbd86d' END;

        -- the row must be byte-identical to the file the site serves; that identity
        -- is the whole safety argument for deploying the row wholesale
        IF r.content_md5 <> v_expect_md5 THEN
            RAISE EXCEPTION '198/547 verify: % theme content md5 is %, expected % — the row is NOT the served file',
                r.domain, r.content_md5, v_expect_md5;
        END IF;
        IF r.site_count <> 1 THEN
            RAISE EXCEPTION '198/547 verify: % still shares its theme with % sites', r.domain, r.site_count;
        END IF;
        IF r.bytes < 4096 THEN
            RAISE EXCEPTION '198/547 verify: % theme is % bytes, still below the 542 gate floor', r.domain, r.bytes;
        END IF;
        -- chrome pins must have come across, or both sites silently lose header/footer
        IF r.header_component_id IS DISTINCT FROM 'e99b0dfa-a3b3-4a29-ae83-73211cb3975e'::uuid
           OR r.footer_component_id IS DISTINCT FROM '09034086-a581-4bba-a5b4-760d863bb2df'::uuid THEN
            RAISE EXCEPTION '198/547 verify: % lost a chrome pin in the split (header=%, footer=%)',
                r.domain, r.header_component_id, r.footer_component_id;
        END IF;
        RAISE NOTICE '198/547: % -> % (% bytes, site_count 1, chrome pinned)', r.domain, r.theme, r.bytes;
    END LOOP;

    -- the seed library asset must be untouched and now unused by any site
    IF (SELECT octet_length(css_content) FROM css_themes WHERE id='fecb962d-3ace-4c19-b08f-088eba46ea53') <> 1649 THEN
        RAISE EXCEPTION '198/547 verify: the seed theme was modified — it must be left exactly as it was';
    END IF;
    IF (SELECT count(*) FROM sites WHERE style_collection_id='3196d966-24ef-4415-9dc8-1afbc02166ca') <> 0 THEN
        RAISE EXCEPTION '198/547 verify: a site still points at the shared seed collection';
    END IF;

    RAISE NOTICE '198/547: verified — both sites have their own composition, seed professional-dark intact and unused';
END $$;

COMMIT;
