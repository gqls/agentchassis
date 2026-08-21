-- 548_seed_webdesign_uk_theme_row_from_deploy_repo.sql
--
-- bugs_open/198 — the LAST armed site in the fleet.
--
-- `webdesign.uk` was recorded in this bug's THIRD WAVE section as "cleared as NOT
-- damage, so nobody re-investigates", on the grounds that it 302-redirects to
-- webdesign.co.uk "and has no stylesheet of its own". The redirect is real; the
-- conclusion is not. **A 302 means nothing is SERVED at that hostname — not that no
-- file exists.** `vm-sites` HEAD carries `webdesign.uk/assets/css/styles.css` at
-- **15582 bytes with 4 `:root` blocks**, while its linked `css_themes` row is
-- **0 bytes**: this bug's trap in its loaded form. The site is `status='deployed'`
-- with 8 pages (7 deployed) and `github_repo='vm-sites'`.
--
-- After migration 547 split the last shared theme, this was the only site left on
-- the refusing side of 542's gate — 21 PASS / 1 REFUSE — and that one was it.
--
-- WHY SEEDING IS THE RIGHT REPAIR HERE, rather than leaving the gate to hold it.
-- The gate makes the site SAFE (a patch is refused, so nothing can clobber the
-- file), but it does not make it PATCHABLE: every contrast finding for this site
-- would park for a human forever. Seeding the row from the deployed file makes
-- row == file, which is the invariant that makes "deploy the whole row" safe, and
-- migration 543 now maintains it at every webdesign-agent render.
--
-- THE SOURCE IS THE DEPLOY REPO, NOT THE URL — and that is the transferable part.
-- For a site behind a redirect, parking domain or CDN rule the served URL cannot
-- tell you what the artefact is; the repo can. Content below is
-- `git -C ~/projects/vm-sites show origin/master:webdesign.uk/assets/css/styles.css`
-- captured 2026-08-21, md5 `a582e515df3a31eeff30359c073205a9`, asserted in the verify block.
--
-- CHECKED BEFORE EMBEDDING, per this file's own restore procedure: 4 `:root` blocks,
-- `--color-primary-ink` present, and the two-clause stale-ink test
-- (`<x>-ink == --color-text AND <x>-ink != --color-<x>`) clean on both slots — so the
-- blob does not carry a pre-2026-08-14 derivation that a seed would reinstate.
--
-- This is a ONE-TIME data repair, like the fleet backfill of 2026-08-21. The guard
-- against it recurring is 543 (the producer now writes the row) plus 542 (the
-- consumer refuses an unsafe one).

DO $$
BEGIN
    IF (SELECT octet_length(ct.css_content)
          FROM sites s
          JOIN style_collections sc ON sc.id = s.style_collection_id
          JOIN css_themes ct ON ct.id = sc.css_theme_id
         WHERE s.domain = 'webdesign.uk') <> 0 THEN
        RAISE EXCEPTION '198/548: already applied (or somebody else seeded it) — the row is no longer 0 bytes';
    END IF;
END $$;

BEGIN;

-- Guarded on the row being EMPTY, so this can never overwrite content someone else
-- put there between writing and applying.
UPDATE css_themes ct
   SET css_content = $CSS_WEBDESIGN_UK$
/* =====================================================================
 * LAYOUT: brochure-bold
 *
 * Character: gradients, scale, motion. Cousin of brochure-formal but
 * louder — taller hero, bigger typography, gradient accents on CTAs,
 * more pronounced hover motion.
 *
 * Renderer contract (same as all layouts):
 *   - No --section-* defaults declared on section containers
 *   - Element rules use var(--section-*, var(--color-*)) fallbacks
 *   - Five surface-managed classes coloured here until Phase 4.5
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        #1a1a1a;
  --color-primary-hover:  #ea580c;
  --color-primary-text:   #ffffff;
  --color-secondary:      #efe6d3;
  --color-accent:         #8a6410;
  --color-background:     #f7f1e6;
  --color-surface:        #efe6d3;
  --color-text:           #1a1a1a;
  --color-text-muted:     #6f6553;
  --color-border:         #d8c3a0;
  --color-card-bg:        #ffffff;
  --color-header-bg:      #09090b;
  --color-header-text:    #ffffff;
  --color-cta-bg:         #c2410c;
  --color-cta-text:       #ffffff;
  --color-footer-bg:      #18181b;
  --color-footer-text:    rgba(255,255,255,0.75);

  /* ── Typography ── */
  --font-body:        'Inter', system-ui, -apple-system, sans-serif;
  --font-heading:     'Playfair Display', Georgia, serif;
  --font-size-base:   16px;
  --line-height-base: 1.6;

  /* ── Structure ── */
  --container-max:    1200px;
  --container-pad-x:  1.5rem;
  --section-pad-y:    6rem;
  --section-pad-y-sm: 3.5rem;
  --hero-min-h:       70vh;
  --radius:           0.75rem;
  --radius-sm:        0.5rem;
  --radius-lg:        1rem;
  --shadow-sm:        0 2px 4px rgba(0,0,0,0.08);
  --shadow-md:        0 8px 20px rgba(0,0,0,0.12);
  --shadow-lg:        0 16px 40px rgba(0,0,0,0.18);
  --transition:       250ms cubic-bezier(0.4, 0, 0.2, 1);
  --card-pad:         2rem;
  --grid-gap:         2rem;
  --cta-pad:          1rem 2rem;
  --gradient-angle:   135deg;

  /* Composite gradient tokens (built from palette slots so they track
     the site's identity). Overridden wholesale by the cta_bg / hero_bg
     palette slots if themes declare them. */
  --gradient-primary: linear-gradient(var(--gradient-angle), var(--color-primary), var(--color-accent));
  --gradient-cta:     linear-gradient(var(--gradient-angle), var(--color-primary), var(--color-secondary));

  /* Per the Colour Inheritance Model contract: if the palette declares
     a 'heading' slot, emit --section-heading at :root level so it
     applies in light sections. Dark-section components override on
     their container. Palettes without a heading slot fall through to
     --color-primary via the h1-h6 fallback chain. */
  
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
  font-weight: 500;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* ── Colour Inheritance Model ──
 * h1-h6 use var(--section-heading, var(--color-primary)) exactly as
 * the contract specifies. Dark-section components override on container. */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 1rem;
  line-height: 1.1;
  font-weight: 800;
  letter-spacing: -0.02em;
}
h1 { font-size: clamp(2.5rem, 5vw, 4rem); }
h2 { font-size: clamp(2rem, 3.5vw, 2.75rem); }
h3 { font-size: 1.625rem; font-weight: 700; }
h4 { font-size: 1.25rem; font-weight: 700; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
/* strong/em/cite/span: do NOT set color — they inherit from parent */
a {
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary-hover); }

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
  position: sticky;
  top: 0;
  z-index: 1000;
  box-shadow: var(--shadow-sm);
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
  font-family: var(--font-heading);
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.02em;
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
  font-weight: 600;
  font-size: 0.95rem;
  padding: 0.5rem 0;
  position: relative;
}
.main-nav a::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: -2px;
  height: 2px;
  background: var(--gradient-primary);
  transform: scaleX(0);
  transform-origin: left;
  transition: transform var(--transition);
}
.main-nav a:hover::after,
.main-nav a.active::after { transform: scaleX(1); }
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
  width: 24px;
  height: 3px;
  background: currentColor;
  margin: 5px 0;
  border-radius: 2px;
}

/* ── Hero (component-coloured background; layout owns scale + layout) ── */
.hero-section {
  min-height: var(--hero-min-h);
  padding-block: calc(var(--section-pad-y) * 1.1);
  display: flex;
  align-items: center;
  position: relative;
  overflow: hidden;
}
.hero-section .container {
  text-align: center;
  max-width: 960px;
  position: relative;
  z-index: 1;
}
.hero-section h1 {
  margin-bottom: 1.5rem;
  letter-spacing: -0.03em;
}
.hero-subtitle, .hero-section .lead {
  font-size: clamp(1.125rem, 1.5vw, 1.375rem);
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2.5rem;
  max-width: 720px;
  font-weight: 400;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* Optional gradient-bar accent under hero heading. Components that
   want this opt in by including a span.hero-accent-bar child of h1. */
.hero-accent-bar {
  display: block;
  width: 80px;
  height: 4px;
  margin: 1.5rem auto 0;
  background: var(--gradient-primary);
  border-radius: 2px;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync
 * with buildSectionDefaults in render_css_from_spec_action.go, which
 * hardcodes the same list to emit dark-surface --section-* overrides
 * when surface is dark. Phase 4.5 of the migration moves this
 * surface-painting responsibility into the components themselves and
 * switches the renderer to a data-section-bg attribute selector;
 * until then, every layout must paint these 5 classes with surface. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── Section grids (3-col → 2 → 1) ── */
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
  border-radius: var(--radius);
  padding: var(--card-pad);
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition), box-shadow var(--transition);
  position: relative;
  overflow: hidden;
}
/* Gradient accent bar along card top */
.feature-card::before,
.service-card::before,
.differentiator-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: var(--gradient-primary);
  opacity: 0;
  transition: opacity var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}
.feature-card:hover::before,
.service-card:hover::before,
.differentiator-card:hover::before { opacity: 1; }
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 56px;
  height: 56px;
  margin-bottom: 1.25rem;
  color: var(--color-primary);
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
  background: var(--color-card-bg);
  border-radius: var(--radius-sm);
  padding: 1.25rem 1.5rem;
  margin-bottom: 0.75rem;
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--transition);
}
.faq-item:hover { box-shadow: var(--shadow-md); }
.faq-item summary {
  cursor: pointer;
  font-weight: 700;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after {
  content: "+";
  font-size: 1.5rem;
  font-weight: 400;
  color: var(--color-primary);
}
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Call to action (component-coloured) ── */
.call-to-action-section { text-align: center; }

/* ── Testimonials (component-coloured) ── */
.testimonials-section .testimonial {
  max-width: 760px;
  margin-inline: auto;
  text-align: center;
  font-size: clamp(1.125rem, 1.4vw, 1.375rem);
  font-weight: 400;
  line-height: 1.5;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1.25rem;
  font-style: normal;
  font-weight: 700;
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
  font-weight: 700;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.875rem 1rem;
  font: inherit;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 44px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 25%, transparent);
}

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: var(--cta-pad);
  font: inherit;
  font-weight: 700;
  font-size: 1rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: var(--radius);
  border: 2px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 48px;
  transition: transform var(--transition), box-shadow var(--transition),
              background var(--transition), color var(--transition),
              border-color var(--transition);
}
.btn-primary {
  background: var(--gradient-cta);
  color: var(--color-primary-text);
  box-shadow: var(--shadow-md);
}
.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
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
  transform: translateY(-2px);
}
.btn-large { padding: 1.25rem 2.5rem; font-size: 1.125rem; min-height: 56px; }

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
.site-footer h3, .site-footer h4 {
  color: #ffffff;
  font-family: var(--font-heading);
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-primary); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255,255,255,0.55);
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
  .hero-section { min-height: auto; padding-block: var(--section-pad-y-sm); }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }
  .btn { width: 100%; }
  .hero-actions { flex-direction: column; }
}

/* ── Accessibility ── */
:focus-visible {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    transform: none !important;
  }
}


/* === Component-specific styles === */
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.fade-in-up { animation: fadeInUp 0.6s ease forwards; }

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
  --color-primary-ink: #1a1a1a;
  --color-accent-ink: #7c5a0e;
  --color-accent-text: #ffffff;
}
$CSS_WEBDESIGN_UK$,
       version = version + 1,
       updated_at = NOW(),
       description = 'seeded 2026-08-21 from vm-sites HEAD (bugs_open/198): the deployed file this row must track'
  FROM style_collections sc, sites s
 WHERE sc.css_theme_id = ct.id
   AND s.style_collection_id = sc.id
   AND s.domain = 'webdesign.uk'
   AND octet_length(COALESCE(ct.css_content, '')) = 0;

DO $$
DECLARE
    v_md5 text;
    v_bytes int;
    v_sites int;
BEGIN
    SELECT md5(ct.css_content), octet_length(ct.css_content),
           (SELECT count(*) FROM sites s2 JOIN style_collections sc2 ON s2.style_collection_id = sc2.id
             WHERE sc2.css_theme_id = ct.id)
      INTO v_md5, v_bytes, v_sites
      FROM sites s
      JOIN style_collections sc ON sc.id = s.style_collection_id
      JOIN css_themes ct ON ct.id = sc.css_theme_id
     WHERE s.domain = 'webdesign.uk';

    IF v_md5 <> 'a582e515df3a31eeff30359c073205a9' THEN
        RAISE EXCEPTION '198/548 verify: row md5 is %, expected % — the row is NOT the deployed file', v_md5, 'a582e515df3a31eeff30359c073205a9';
    END IF;
    IF v_bytes < 4096 THEN
        RAISE EXCEPTION '198/548 verify: row is % bytes, still below the 542 gate floor', v_bytes;
    END IF;
    IF v_sites <> 1 THEN
        RAISE EXCEPTION '198/548 verify: theme is shared by % sites', v_sites;
    END IF;

    RAISE NOTICE '198/548: verified — webdesign.uk row seeded to % bytes, md5 matches the deploy repo blob, site_count 1', v_bytes;
END $$;

COMMIT;
