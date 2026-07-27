-- 085b_palette_and_component_contrast_fix.sql
--
-- Measured defect: 101 WCAG-AA contrast failures across the 5 live pages of
-- fundamentallyai.com, in two mechanical families, plus one that only appears
-- once the first is fixed. Measured by rendering each page in headless Chromium
-- and computing the effective foreground/background pair for every text node
-- (scripts/render_audit.py). After this file + the CSS regeneration: 1 failure
-- left, and that one is a 12-site defect deferred to the renderer change.
--
-- FAMILY 1 — the palette defines only the 8 CORE slots, but layouts declare 17.
--   The 9 specialised slots therefore took `brochure-formal`'s own LIGHT
--   literals: card_bg #ffffff, header_bg #ffffff, cta_bg #1a365d. A white card
--   on a dark site, carrying text coloured for the dark site: 1.21:1.
--   Fleet-wide this is 12 of 31 palettes (all the generated dark ones).
--
-- FAMILY 2 — `primary` #0E1B2E scores 1.11:1 on its own background #090F1A.
--   The component library uses --color-primary as a FOREGROUND 53 times
--   (eyebrows, links, card titles) and as a background 26 times; one token
--   cannot hold both roles unless it contrasts with the page. Every eyebrow
--   and card title on the site was invisible.
--
-- FAMILY 3 (exposed by fixing 2) — three components paint themselves in
--   --color-primary and hard-code white ink over it. Fixed in the templates
--   below, role-correctly, so they work in both schemes.
--
-- Nothing here is cosmetic preference: every value is chosen to clear WCAG AA
-- against the surface it actually lands on, and the audit is the check.

BEGIN;

-- 1. The palette row the CSS renderer composes from.
UPDATE palettes
   SET colours = $pal${
  "text": "#E8EDF3",
  "accent": "#C8902A",
  "border": "#1E3050",
  "primary": "#86ADDE",
  "surface": "#132239",
  "secondary": "#4A6C99",
  "background": "#090F1A",
  "text_muted": "#8A9BB0",
  "primary_hover": "#A8C6EA",
  "primary_text": "#071019",
  "secondary_hover": "#5C80B0",
  "secondary_text": "#FFFFFF",
  "card_bg": "#132239",
  "header_bg": "#0B1424",
  "header_text": "#E8EDF3",
  "cta_bg": "#101E33",
  "cta_text": "#E8EDF3",
  "footer_bg": "#060B14",
  "footer_text": "rgba(232,237,243,0.88)",
  "heading": "#F2F6FA",
  "hero_title": "#FFFFFF",
  "hero_subtitle": "#C7D4E4"
}$pal$::jsonb,
       updated_at = NOW()
 WHERE id = 'c7c5435f-7cde-4cb7-9398-045bbb5be84a';

-- 2. The theme row carries its own copy; keep them identical or the next
--    composition run reintroduces the gap.
UPDATE css_themes
   SET color_palette = $pal${
  "text": "#E8EDF3",
  "accent": "#C8902A",
  "border": "#1E3050",
  "primary": "#86ADDE",
  "surface": "#132239",
  "secondary": "#4A6C99",
  "background": "#090F1A",
  "text_muted": "#8A9BB0",
  "primary_hover": "#A8C6EA",
  "primary_text": "#071019",
  "secondary_hover": "#5C80B0",
  "secondary_text": "#FFFFFF",
  "card_bg": "#132239",
  "header_bg": "#0B1424",
  "header_text": "#E8EDF3",
  "cta_bg": "#101E33",
  "cta_text": "#E8EDF3",
  "footer_bg": "#060B14",
  "footer_text": "rgba(232,237,243,0.88)",
  "heading": "#F2F6FA",
  "hero_title": "#FFFFFF",
  "hero_subtitle": "#C7D4E4"
}$pal$::jsonb,
       updated_at = NOW()
 WHERE id = 'b62650b3-d4df-4cb3-b586-b09b262dafa4';

-- 3. design_intent.palette.reference_values is the pin that stops a webdesign
--    run re-rolling the core colours (webdesign colour-churn landmine). The two
--    core values that changed must change here too, or the next design pass
--    restores the unreadable ones.
UPDATE site_specs
   SET data = jsonb_set(
                jsonb_set(data, '{palette,reference_values,primary}',   '"#86ADDE"'::jsonb, true),
                '{palette,reference_values,secondary}', '"#4A6C99"'::jsonb, true),
       updated_at = NOW()
 WHERE site_id = (SELECT id FROM sites WHERE domain = 'fundamentallyai.com')
   AND aspect = 'design_intent'
   AND is_current;

-- 4. hero — the inverted CTA in the image branch hard-codes --hero-ink: #fff
--    and colours its text --color-primary. Correct in the no-image branch
--    (where --hero-ink IS --color-primary-text, a designed pair) and broken in
--    the image branch on any site whose primary is light. The image branch now
--    supplies --hero-btn-ink alongside the white ink it already hard-codes;
--    sites without a hero image are byte-identical in effect.
UPDATE content_components SET html_template = $tpl$<section class="hero" data-component="hero" style="{{if or .hero_url .background_image}}background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('{{or .hero_url .background_image}}'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;{{else}}--hero-ink: var(--color-primary-text); background: var(--color-primary); background: linear-gradient(135deg, var(--color-primary) 0%, color-mix(in srgb, var(--color-primary) 85%, var(--color-primary-text)) 100%);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
            {{if and .secondary_cta .secondary_cta_url}}<a href="{{.secondary_cta_url}}" class="btn btn-secondary">{{.secondary_cta}}</a>{{end}}
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
$tpl$, updated_at = NOW()
 WHERE id = '23f95f00-f293-466e-b43a-81791ea0fc6c';

-- 5. stat-band — painted in --color-primary, but the layout gives every heading
--    `color: var(--section-heading, ...)`, and a matched rule beats the
--    inherited colour. Restates its own --section-* per the Dark Section
--    Variable Contract.
UPDATE content_components SET html_template = $tpl$
<style>
  .stat-band {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-primary);
    color: var(--color-primary-text, #fff);

    /* Dark Section Variable Contract: the layout gives every heading
       color: var(--section-heading, ...), and a matched rule beats the
       inherited colour above — so a band painted in --color-primary must
       restate its own --section-* or its headings take the page's. */
    --section-text: var(--color-primary-text, #fff);
    --section-text-muted: color-mix(in srgb, var(--color-primary-text, #fff) 78%, transparent);
    --section-heading: var(--color-primary-text, #fff);
    --section-surface: color-mix(in srgb, var(--color-primary-text, #fff) 10%, transparent);
    --section-border: color-mix(in srgb, var(--color-primary-text, #fff) 30%, transparent);
  }
  .stat-band__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .stat-band__header {
    text-align: center;
    max-width: 60ch;
    margin: 0 auto 2.5rem;
  }
  .stat-band__eyebrow {
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    opacity: 0.85;
  }
  .stat-band__title {
    font-size: clamp(1.5rem, 2.6vw, 2.1rem);
    font-weight: 700;
    line-height: 1.25;
    margin: 0.5rem 0 0;
  }
  .stat-band__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 2rem 1.5rem;
    text-align: center;
  }
  .stat-band__stat {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0 0.5rem;
  }
  .stat-band__stat + .stat-band__stat {
    border-inline-start: 1px solid color-mix(in srgb, var(--color-primary-text, #fff) 22%, transparent);
  }
  @media (max-width: 620px) {
    .stat-band__stat + .stat-band__stat { border-inline-start: 0; }
  }
  .stat-band__value {
    font-size: clamp(2.5rem, 5vw, 3.5rem);
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.02em;
    font-variant-numeric: tabular-nums;
  }
  .stat-band__label {
    font-size: 1rem;
    font-weight: 600;
  }
  .stat-band__caption {
    font-size: 0.8125rem;
    opacity: 0.8;
    line-height: 1.5;
  }
  @media (max-width: 768px) { .stat-band { padding: 3.25rem 1.25rem; } }
</style>

<section class="stat-band" data-component="stat-band">
  <div class="stat-band__inner">
    {{if or .section_title .section_eyebrow}}
    <header class="stat-band__header">
      {{if .section_eyebrow}}<span class="stat-band__eyebrow">{{.section_eyebrow}}</span>{{end}}
      {{if .section_title}}<h2 class="stat-band__title">{{.section_title}}</h2>{{end}}
    </header>
    {{end}}
    <div class="stat-band__grid">
      {{range .stats}}{{if .value}}
      <div class="stat-band__stat">
        <span class="stat-band__value" data-countup aria-label="{{.value}}">{{.value}}</span>
        <span class="stat-band__label">{{.label}}</span>
        {{if .caption}}<span class="stat-band__caption">{{.caption}}</span>{{end}}
      </div>
      {{end}}{{end}}
    </div>
  </div>
</section>

$tpl$, updated_at = NOW()
 WHERE id = '62859c4c-27ac-4859-8603-175930e1e325';

-- 6. portfolio-showcase — painted in --color-primary with nine hard-coded
--    rgba(255,255,255,x) inks. Repainted with --color-cta-bg / --color-cta-text,
--    which are a curated pair in every palette.
UPDATE content_components SET html_template = $tpl$<section class="portfolio-showcase-section" data-component="portfolio-showcase">
    <div class="portfolio-container">
        <h2>{{.headline}}</h2>
        {{if .intro}}<p class="portfolio-intro">{{.intro}}</p>{{end}}
        <div class="portfolio-grid">
            {{range .projects}}
            <div class="portfolio-item">
                <div class="portfolio-item-header">
                    <h3>{{.title}}</h3>
                    {{if .live_url}}<a href="{{.live_url}}" class="portfolio-link" target="_blank" rel="noopener">Visit Site &#8594;</a>{{end}}
                </div>
                {{if .domain}}<p class="portfolio-domain">{{.domain}}</p>{{end}}
                <p class="portfolio-description">{{.description}}</p>
                <div class="portfolio-meta">
                    {{if .built_with}}<span class="portfolio-tag">{{.built_with}}</span>{{end}}
                    {{if .build_time}}<span class="portfolio-tag portfolio-tag-time">{{.build_time}}</span>{{end}}
                </div>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Dark section to match the visual weight of the social-proof section it replaces.
   Painted with --color-cta-bg (a surface-role token) rather than --color-primary:
   primary is also the library's foreground colour for links and eyebrows, so on a
   dark site it is a LIGHT value and the hard-coded white text below vanished. Every
   ink here now derives from --color-cta-text, which is the pair --color-cta-bg was
   curated with. */
.portfolio-showcase-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-cta-bg, var(--color-primary, #1a1a2e));
    color: var(--color-cta-text, var(--color-white, #fff));
    --pf-ink: var(--color-cta-text, var(--color-white, #fff));
    --section-text: var(--pf-ink);
    --section-text-muted: color-mix(in srgb, var(--pf-ink) 78%, transparent);
    --section-heading: var(--pf-ink);
    --section-surface: color-mix(in srgb, var(--pf-ink) 10%, transparent);
    --section-border: color-mix(in srgb, var(--pf-ink) 30%, transparent);
}
.portfolio-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.portfolio-showcase-section h2 {
    text-align: center;
    margin-bottom: 1rem;
    color: var(--pf-ink);
}
.portfolio-intro {
    text-align: center;
    max-width: 700px;
    margin: 0 auto 3rem;
    color: color-mix(in srgb, var(--pf-ink) 85%, transparent);
    line-height: 1.7;
}
.portfolio-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.portfolio-item {
    padding: 2rem;
    background: color-mix(in srgb, var(--pf-ink) 7%, transparent);
    border-radius: 8px;
    border-left: 3px solid var(--color-accent, #0f3460);
    transition: transform 0.2s, box-shadow 0.2s;
}
.portfolio-item:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.2);
}
.portfolio-item-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 0.5rem;
}
.portfolio-item h3 {
    margin: 0;
    font-size: 1.2rem;
    color: var(--pf-ink);
}
.portfolio-link {
    color: var(--color-accent, #4da6ff);
    text-decoration: none;
    font-size: 0.85rem;
    font-weight: 500;
    white-space: nowrap;
    transition: color 0.2s;
}
.portfolio-link:hover {
    color: var(--pf-ink);
}
.portfolio-domain {
    font-family: monospace;
    font-size: 0.85rem;
    color: color-mix(in srgb, var(--pf-ink) 72%, transparent);
    margin-bottom: 1rem;
}
.portfolio-description {
    color: color-mix(in srgb, var(--pf-ink) 92%, transparent);
    line-height: 1.7;
    margin-bottom: 1.5rem;
}
.portfolio-meta {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
}
.portfolio-tag {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    background: color-mix(in srgb, var(--pf-ink) 14%, transparent);
    border-radius: 4px;
    font-size: 0.8rem;
    color: color-mix(in srgb, var(--pf-ink) 88%, transparent);
}
.portfolio-tag-time {
    background: color-mix(in srgb, var(--color-accent, #4da6ff) 18%, transparent);
    color: var(--color-accent, #4da6ff);
}
@media (max-width: 768px) {
    .portfolio-showcase-section { padding: 3rem 1.5rem; }
    .portfolio-grid { grid-template-columns: 1fr; }
    .portfolio-item-header { flex-direction: column; gap: 0.5rem; }
}
</style>
$tpl$, updated_at = NOW()
 WHERE id = '26c74966-19d4-4f3f-b61b-d77cc4876351';

COMMIT;

-- Verify:
--   SELECT jsonb_object_keys(colours) FROM palettes WHERE id='c7c5435f-7cde-4cb7-9398-045bbb5be84a';  -- 22 slots
--   then regenerate styles.css and re-run scripts/render_audit.py against the
--   SERVED pages. The DB is not the check; the rendered page is.
