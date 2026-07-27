-- 085d_imagery_wiring_and_component_roles.sql
--
-- The site had 21 generated line-illustration assets DEPLOYED and serving 200,
-- and referenced three of them. The brief asks for an imagery-rich consultancy
-- register; the platform planned it (23 site_plan_imagery rows), generated it,
-- deployed it, and then nothing wired it into the pages, because the five
-- "Re-render <page> after its image asset landed" work items were parked in
-- needs_human_review on 2026-07-20/21 and never drained (14 of 28 such items
-- fleet-wide are parked the same way).
--
-- Two classes of change:
--   TEMPLATES — three components fell back to /assets/images/hero.jpg, a
--   filename that exists on no site. A missing asset now renders NO <img>
--   rather than a broken one; and the three secondary heroes gain the image
--   support `hero` already had.
--   ROLE — components painting a full-bleed band in --color-primary now use
--   --color-cta-bg/--color-cta-text. primary is the library's FOREGROUND
--   colour in 53 places; a palette where it is light (every dark site) had
--   those bands inverted.

BEGIN;

-- hero-card-carousel — no asset -> no <img>. The fallback was /assets/images/hero.jpg, a filename that exists on NO site in the fleet; four cards rendered a broken-image icon with the alt text showing.
UPDATE content_components SET html_template = $tpl$<style>
  .hero-card-carousel {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .hero-card-carousel__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .hero-card-carousel__eyebrow {
    display: inline-block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary);
    margin-bottom: 0.75rem;
  }
  .hero-card-carousel__title {
    font-size: clamp(1.75rem, 3.5vw, 2.75rem);
    font-weight: 700;
    line-height: 1.15;
    color: var(--color-heading);
    margin: 0 0 1.75rem;
    max-width: 40ch;
  }
  .hero-card-carousel__head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1.5rem;
    flex-wrap: wrap;
  }
  .hero-card-carousel__pause {
    inline-size: 44px;
    block-size: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--color-border);
    background: var(--color-surface, transparent);
    color: var(--color-heading);
    border-radius: 999px;
    font-size: 1rem;
    line-height: 1;
    cursor: pointer;
    margin-bottom: 1.25rem;
    transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
  }
  .hero-card-carousel__pause:hover {
    background: var(--color-primary);
    border-color: var(--color-primary);
    color: var(--color-primary-text, #fff);
  }
  .hero-card-carousel__pause:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }

  /* Viewport is the positioning context for the overlaid arrows. */
  .hero-card-carousel__viewport { position: relative; }

  /* Track — native scroll-snap gives swipe-on-mobile with zero JS. align-items:
     start so cards size to their own content (no equal-height whitespace). */
  .hero-card-carousel__track {
    list-style: none;
    margin: 0;
    padding: 0.5rem 0 1.5rem;
    display: flex;
    align-items: flex-start;
    gap: 1.5rem;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    scroll-behavior: smooth;
    scrollbar-width: none;
  }
  .hero-card-carousel__track::-webkit-scrollbar { display: none; }
  .hero-card-carousel__slide {
    scroll-snap-align: start;
    flex: 0 0 min(85%, 620px);
  }
  @media (min-width: 900px) {
    .hero-card-carousel__slide { flex-basis: calc((100% - 3rem) / 2.15); }
  }

  /* Overlaid prev/next arrows — nudged up to sit over the image, not the text. */
  .hero-card-carousel__arrow {
    position: absolute;
    top: 34%;
    transform: translateY(-50%);
    z-index: 2;
    inline-size: 48px;
    block-size: 48px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--color-hairline, var(--color-border));
    background: var(--color-background);
    color: var(--color-heading);
    border-radius: 999px;
    font-size: 1.6rem;
    line-height: 1;
    cursor: pointer;
    box-shadow: var(--shadow, 0 2px 12px rgba(0,0,0,0.18));
    transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease;
  }
  .hero-card-carousel__arrow--prev { left: 0.5rem; }
  .hero-card-carousel__arrow--next { right: 0.5rem; }
  .hero-card-carousel__arrow:hover {
    background: var(--color-primary);
    border-color: var(--color-primary);
    color: var(--color-primary-text, #fff);
  }
  .hero-card-carousel__arrow:focus-visible { outline: 2px solid var(--color-primary); outline-offset: 2px; }

  /* The whole card is the click target. */
  .hero-card-carousel__card {
    display: flex;
    flex-direction: column;
    background: var(--color-surface, transparent);
    border: 1px solid var(--color-hairline, var(--color-border));
    border-radius: var(--border-radius, 0.75rem);
    overflow: hidden;
    text-decoration: none;
    color: inherit;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }
  a.hero-card-carousel__card:hover {
    border-color: var(--color-primary);
    box-shadow: var(--shadow, 0 6px 20px rgba(0,0,0,0.1));
  }
  a.hero-card-carousel__card:focus-visible {
    outline: 3px solid var(--color-primary);
    outline-offset: 2px;
  }

  /* Hover-zoom — the clip is required; the scaled image would otherwise spill. */
  .hero-card-carousel__media {
    overflow: hidden;
    aspect-ratio: 16 / 10;
    background: var(--color-surface-alt, var(--color-surface));
  }
  .hero-card-carousel__img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: scale 0.45s ease;
  }
  .hero-card-carousel__card:hover .hero-card-carousel__img,
  .hero-card-carousel__card:focus-visible .hero-card-carousel__img {
    scale: 1.08;
  }

  /* Compact body — content stacks tight, no filler gap below the text. */
  .hero-card-carousel__body {
    padding: 1.25rem 1.5rem 1.4rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .hero-card-carousel__card-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-heading);
    margin: 0;
    line-height: 1.25;
  }
  .hero-card-carousel__teaser {
    font-size: 0.9375rem;
    color: var(--color-text-muted);
    line-height: 1.6;
    margin: 0;
  }
  .hero-card-carousel__link {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-primary);
    margin-top: 0.15rem;
  }
  a.hero-card-carousel__card:hover .hero-card-carousel__link { text-decoration: underline; }

  .hero-card-carousel__live {
    position: absolute;
    width: 1px; height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }

  /* Accessibility: no motion for users who ask for none; no zoom on touch. */
  @media (prefers-reduced-motion: reduce) {
    .hero-card-carousel__track { scroll-behavior: auto; }
    .hero-card-carousel__img { transition: none; }
    .hero-card-carousel__card:hover .hero-card-carousel__img,
    .hero-card-carousel__card:focus-visible .hero-card-carousel__img { scale: 1; }
  }
  @media (hover: none) {
    .hero-card-carousel__card:hover .hero-card-carousel__img { scale: 1; }
  }
  @media (max-width: 768px) {
    .hero-card-carousel { padding: 3rem 1.25rem; }
    .hero-card-carousel__arrow { inline-size: 42px; block-size: 42px; font-size: 1.4rem; }
    .hero-card-carousel__arrow--prev { left: 0.25rem; }
    .hero-card-carousel__arrow--next { right: 0.25rem; }
  }
</style>

<section class="hero-card-carousel" data-component="hero-card-carousel" data-hcc-autoplay="{{if .autoplay}}true{{else}}false{{end}}" role="region" aria-roledescription="carousel" aria-label="{{if .section_title}}{{.section_title}}{{else}}Featured{{end}}">
  <div class="hero-card-carousel__inner">
    <div class="hero-card-carousel__head">
      <div>
        {{if .section_eyebrow}}<span class="hero-card-carousel__eyebrow">{{.section_eyebrow}}</span>{{end}}
        {{if .section_title}}<h2 class="hero-card-carousel__title">{{.section_title}}</h2>{{end}}
      </div>
      {{if .autoplay}}<button type="button" class="hero-card-carousel__pause" data-hcc-pause aria-label="Pause automatic rotation"><span data-hcc-pause-icon aria-hidden="true">&#10073;&#10073;</span></button>{{end}}
    </div>

    <div class="hero-card-carousel__viewport">
      <button type="button" class="hero-card-carousel__arrow hero-card-carousel__arrow--prev" data-hcc-prev aria-label="Previous card"><span aria-hidden="true">&lsaquo;</span></button>
      <ul class="hero-card-carousel__track" data-hcc-track>
        {{range $i, $card := .cards}}
        <li class="hero-card-carousel__slide" role="group" aria-roledescription="slide" aria-label="{{$card.title}}" data-hcc-slide>
          {{if $card.link_url}}<a class="hero-card-carousel__card" href="{{$card.link_url}}">{{else}}<div class="hero-card-carousel__card">{{end}}
            {{if $card.image}}<div class="hero-card-carousel__media">
              <img class="hero-card-carousel__img" src="{{$card.image}}" alt="{{$card.image_alt}}" width="800" height="500" loading="{{if eq $i 0}}eager{{else}}lazy{{end}}">
            </div>{{end}}
            <div class="hero-card-carousel__body">
              <h3 class="hero-card-carousel__card-title">{{$card.title}}</h3>
              <p class="hero-card-carousel__teaser">{{$card.teaser}}</p>
              {{if $card.link_url}}<span class="hero-card-carousel__link">{{if $card.link_label}}{{$card.link_label}}{{else}}Read more{{end}}<span aria-hidden="true">&nbsp;&rarr;</span></span>{{end}}
            </div>
          {{if $card.link_url}}</a>{{else}}</div>{{end}}
        </li>
        {{end}}
      </ul>
      <button type="button" class="hero-card-carousel__arrow hero-card-carousel__arrow--next" data-hcc-next aria-label="Next card"><span aria-hidden="true">&rsaquo;</span></button>
    </div>

    <div class="hero-card-carousel__live" aria-live="polite" data-hcc-live></div>
  </div>
</section>

$tpl$, updated_at = NOW()
 WHERE id = '82274d36-9024-4727-a0e7-36e964c6767e';

-- people-feature-block — same guessed fallback, same result on about.
UPDATE content_components SET html_template = $tpl$
<style>
  .people-feature-block {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .people-feature-block__inner {
    max-width: var(--container-max-width, 1100px);
    margin: 0 auto;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: clamp(2rem, 5vw, 4rem);
    align-items: center;
  }
  .people-feature-block--reverse .people-feature-block__media { order: 2; }

  .people-feature-block__media {
    border-radius: var(--border-radius, 1rem);
    overflow: hidden;
    aspect-ratio: 4 / 3;
    background: var(--color-surface-alt, var(--color-surface));
    border: 1px solid var(--color-hairline, var(--color-border));
  }
  .people-feature-block__img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .people-feature-block__eyebrow {
    display: inline-block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary);
    margin-bottom: 0.75rem;
  }
  .people-feature-block__title {
    font-size: clamp(1.6rem, 3vw, 2.4rem);
    font-weight: 700;
    line-height: 1.2;
    color: var(--color-heading);
    margin: 0 0 1rem;
  }
  .people-feature-block__body {
    font-size: 1.0625rem;
    line-height: 1.7;
    color: var(--color-text);
    margin: 0 0 1.25rem;
  }
  .people-feature-block__quote {
    margin: 0 0 1.25rem;
    padding-inline-start: 1.1rem;
    border-inline-start: 3px solid var(--color-primary);
    font-size: 1.125rem;
    font-style: italic;
    line-height: 1.5;
    color: var(--color-heading);
  }
  .people-feature-block__attribution {
    display: block;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    font-style: normal;
    font-weight: 600;
    color: var(--color-text-muted);
  }
  .people-feature-block__link {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-primary);
    text-decoration: none;
    min-height: 44px;
  }
  .people-feature-block__link:hover { text-decoration: underline; }
  .people-feature-block__link:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
    border-radius: 2px;
  }

  @media (max-width: 820px) {
    .people-feature-block__inner { grid-template-columns: 1fr; }
    .people-feature-block--reverse .people-feature-block__media { order: 0; }
    .people-feature-block { padding: 3rem 1.25rem; }
  }
</style>

<section class="people-feature-block{{if .reverse}} people-feature-block--reverse{{end}}" data-component="people-feature-block">
  <div class="people-feature-block__inner">
    {{if .image}}<div class="people-feature-block__media">
      <img class="people-feature-block__img" src="{{.image}}" alt="{{.image_alt}}" loading="lazy" width="720" height="540">
    </div>{{end}}
    <div class="people-feature-block__content">
      {{if .eyebrow}}<span class="people-feature-block__eyebrow">{{.eyebrow}}</span>{{end}}
      {{if .title}}<h2 class="people-feature-block__title">{{.title}}</h2>{{end}}
      {{if .body}}<p class="people-feature-block__body">{{.body}}</p>{{end}}
      {{if .quote}}<blockquote class="people-feature-block__quote">{{.quote}}{{if .attribution}}<cite class="people-feature-block__attribution">{{.attribution}}</cite>{{end}}</blockquote>{{end}}
      {{if .link_url}}<a class="people-feature-block__link" href="{{.link_url}}">{{if .link_label}}{{.link_label}}{{else}}Read more{{end}}<span aria-hidden="true">&nbsp;&rarr;</span></a>{{end}}
    </div>
  </div>
</section>

$tpl$, updated_at = NOW()
 WHERE id = '57b96b01-591d-4a4c-86cd-64ca7beb1db5';

-- image-hover-card-grid — same guessed fallback; unexercised here but the same defect.
UPDATE content_components SET html_template = $tpl$
<style>
  .image-hover-card-grid {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .image-hover-card-grid__inner {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
  }
  .image-hover-card-grid__header {
    max-width: 60ch;
    margin: 0 auto 3rem;
    text-align: center;
  }
  .image-hover-card-grid__eyebrow {
    display: inline-block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary);
    margin-bottom: 0.75rem;
  }
  .image-hover-card-grid__title {
    font-size: clamp(1.6rem, 3vw, 2.4rem);
    font-weight: 700;
    line-height: 1.2;
    color: var(--color-heading);
    margin: 0;
  }
  .image-hover-card-grid__grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .image-hover-card-grid__card {
    position: relative;
    display: block;
    text-decoration: none;
    color: #fff;
    border-radius: var(--border-radius, 0.75rem);
    overflow: hidden;
    aspect-ratio: 4 / 5;
    background: var(--color-surface-alt, var(--color-surface));
    isolation: isolate;
  }
  .image-hover-card-grid__img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    z-index: -2;
    transition: scale 0.5s ease;
  }
  /* Legibility scrim; deepens on reveal. */
  .image-hover-card-grid__card::after {
    content: "";
    position: absolute;
    inset: 0;
    z-index: -1;
    background: linear-gradient(to top, rgba(0,0,0,0.78) 0%, rgba(0,0,0,0.35) 45%, rgba(0,0,0,0.12) 100%);
    transition: background 0.4s ease;
  }
  .image-hover-card-grid__body {
    position: absolute;
    inset-inline: 0;
    inset-block-end: 0;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .image-hover-card-grid__card-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin: 0;
    line-height: 1.25;
    text-shadow: 0 1px 2px rgba(0,0,0,0.4);
  }
  /* Description: present in the DOM for screen readers at all times; visually
     revealed on hover/focus. Uses opacity + transform (NOT display:none), so it
     is never removed from the accessibility tree. */
  .image-hover-card-grid__desc {
    font-size: 0.9375rem;
    line-height: 1.55;
    margin: 0;
    opacity: 0;
    max-height: 0;
    transform: translateY(0.5rem);
    overflow: hidden;
    transition: opacity 0.35s ease, max-height 0.35s ease, transform 0.35s ease;
  }
  .image-hover-card-grid__cue {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.8125rem;
    font-weight: 600;
    margin-top: 0.25rem;
  }
  .image-hover-card-grid__card:hover .image-hover-card-grid__img,
  .image-hover-card-grid__card:focus-visible .image-hover-card-grid__img {
    scale: 1.07;
  }
  .image-hover-card-grid__card:hover .image-hover-card-grid__desc,
  .image-hover-card-grid__card:focus-visible .image-hover-card-grid__desc {
    opacity: 1;
    max-height: 12rem;
    transform: translateY(0);
  }
  .image-hover-card-grid__card:focus-visible {
    outline: 3px solid var(--color-primary);
    outline-offset: 3px;
  }

  /* Touch devices can't hover — show the description permanently there. */
  @media (hover: none) {
    .image-hover-card-grid__desc { opacity: 1; max-height: 12rem; transform: none; }
    .image-hover-card-grid__cue { display: none; }
  }
  @media (prefers-reduced-motion: reduce) {
    .image-hover-card-grid__img,
    .image-hover-card-grid__desc,
    .image-hover-card-grid__card::after { transition: none; }
    .image-hover-card-grid__card:hover .image-hover-card-grid__img { scale: 1; }
  }
  @media (max-width: 768px) {
    .image-hover-card-grid { padding: 3rem 1.25rem; }
    .image-hover-card-grid__grid { grid-template-columns: 1fr; }
  }
</style>

<section class="image-hover-card-grid" data-component="image-hover-card-grid">
  <div class="image-hover-card-grid__inner">
    <header class="image-hover-card-grid__header">
      {{if .section_eyebrow}}<span class="image-hover-card-grid__eyebrow">{{.section_eyebrow}}</span>{{end}}
      {{if .section_title}}<h2 class="image-hover-card-grid__title">{{.section_title}}</h2>{{end}}
    </header>
    <div class="image-hover-card-grid__grid">
      {{range .cards}}
      <a class="image-hover-card-grid__card" href="{{if .link_url}}{{.link_url}}{{else}}#{{end}}">
        {{if .image}}<img class="image-hover-card-grid__img" src="{{.image}}" alt="{{.image_alt}}" loading="lazy" width="600" height="750">{{end}}
        <div class="image-hover-card-grid__body">
          <h3 class="image-hover-card-grid__card-title">{{.title}}</h3>
          <p class="image-hover-card-grid__desc">{{.description}}</p>
          {{if .link_label}}<span class="image-hover-card-grid__cue">{{.link_label}}<span aria-hidden="true">&nbsp;&rarr;</span></span>{{end}}
        </div>
      </a>
      {{end}}
    </div>
  </div>
</section>

$tpl$, updated_at = NOW()
 WHERE id = '2a6c8743-6efb-4bb0-8592-2240a3925760';

-- hero-services — takes hero_url/background_image like `hero` already does, and paints its gradient from the cta pair rather than --color-primary.
UPDATE content_components SET html_template = $tpl$<section class="hero hero-services{{if or .hero_url .background_image}} hero-services--imaged{{end}}" data-component="hero-services"{{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62), rgba(6,11,20,0.72)), url('{{or .hero_url .background_image}}'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-services {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    /* Emphasis-band role: --color-cta-bg / --color-cta-text are a curated
       pair in every palette. --color-primary is ALSO the library's foreground
       colour for links and eyebrows, so on a dark site it is a light value and
       painting a full-bleed hero with it inverts the site's register. */
    --hero-ink: var(--color-cta-text, var(--color-primary-text));
    background: var(--color-cta-bg, var(--color-primary));
    background: linear-gradient(135deg, var(--color-cta-bg, var(--color-primary)) 0%, color-mix(in srgb, var(--color-cta-bg, var(--color-primary)) 82%, var(--color-cta-text, var(--color-primary-text))) 100%);

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--hero-ink) 90%, transparent);
    --section-text-muted: color-mix(in srgb, var(--hero-ink) 70%, transparent);
    --section-heading: var(--hero-ink);
    --section-surface: color-mix(in srgb, var(--hero-ink) 5%, transparent);
    --section-border: color-mix(in srgb, var(--hero-ink) 20%, transparent);
}
.hero-services .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: var(--hero-ink);
}
.hero-services h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
}
.hero-services .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
.hero-services--imaged { background-image: inherit; }
</style>
$tpl$, updated_at = NOW()
 WHERE id = '10d65832-a60c-4435-bdf3-36f1857b271c';

-- hero-about — as hero-services.
UPDATE content_components SET html_template = $tpl$<section class="hero hero-about{{if or .hero_url .background_image}} hero-about--imaged{{end}}" data-component="hero-about"{{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62), rgba(6,11,20,0.72)), url('{{or .hero_url .background_image}}'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-about {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    /* Emphasis-band role: --color-cta-bg / --color-cta-text are a curated
       pair in every palette. --color-primary is ALSO the library's foreground
       colour for links and eyebrows, so on a dark site it is a light value and
       painting a full-bleed hero with it inverts the site's register. */
    --hero-ink: var(--color-cta-text, var(--color-primary-text));
    background: var(--color-cta-bg, var(--color-primary));
    background: linear-gradient(135deg, var(--color-cta-bg, var(--color-primary)) 0%, color-mix(in srgb, var(--color-cta-bg, var(--color-primary)) 82%, var(--color-cta-text, var(--color-primary-text))) 100%);

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--hero-ink) 90%, transparent);
    --section-text-muted: color-mix(in srgb, var(--hero-ink) 70%, transparent);
    --section-heading: var(--hero-ink);
    --section-surface: color-mix(in srgb, var(--hero-ink) 5%, transparent);
    --section-border: color-mix(in srgb, var(--hero-ink) 20%, transparent);
}
.hero-about .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: var(--hero-ink);
}
.hero-about h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
}
.hero-about .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
.hero-about--imaged { background-image: inherit; }
</style>
$tpl$, updated_at = NOW()
 WHERE id = 'e0db9a5b-d051-46e9-b4b3-16a434f9c07a';

-- hero-contact — as hero-services.
UPDATE content_components SET html_template = $tpl$<section class="hero hero-contact{{if or .hero_url .background_image}} hero-contact--imaged{{end}}" data-component="hero-contact"{{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62), rgba(6,11,20,0.72)), url('{{or .hero_url .background_image}}'); background-size: cover; background-position: center; --hero-ink: #fff; --hero-btn-ink: #0F1115;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
        </div>
    </section>
<style>
.hero-contact {
    min-height: 50vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    /* Emphasis-band role: --color-cta-bg / --color-cta-text are a curated
       pair in every palette. --color-primary is ALSO the library's foreground
       colour for links and eyebrows, so on a dark site it is a light value and
       painting a full-bleed hero with it inverts the site's register. */
    --hero-ink: var(--color-cta-text, var(--color-primary-text));
    background: var(--color-cta-bg, var(--color-primary));
    background: linear-gradient(135deg, var(--color-cta-bg, var(--color-primary)) 0%, color-mix(in srgb, var(--color-cta-bg, var(--color-primary)) 82%, var(--color-cta-text, var(--color-primary-text))) 100%);

    /* Dark section context */
    --section-text: color-mix(in srgb, var(--hero-ink) 90%, transparent);
    --section-text-muted: color-mix(in srgb, var(--hero-ink) 70%, transparent);
    --section-heading: var(--hero-ink);
    --section-surface: color-mix(in srgb, var(--hero-ink) 5%, transparent);
    --section-border: color-mix(in srgb, var(--hero-ink) 20%, transparent);
}
.hero-contact .hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: var(--hero-ink);
}
.hero-contact h1 {
    font-size: clamp(1.75rem, 4vw, 2.75rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
}
.hero-contact .hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.2rem);
    line-height: 1.6;
}
.hero-contact--imaged { background-image: inherit; }
</style>
$tpl$, updated_at = NOW()
 WHERE id = '231098ee-085d-4005-ae50-cc9863d3af9d';

-- stat-band — emphasis band repainted from the cta pair; --section-* restated (already in 085b, re-applied here on top).
UPDATE content_components SET html_template = $tpl$
<style>
  .stat-band {
    padding: var(--spacing-section, 4.5rem 2rem);
    background: var(--color-cta-bg, var(--color-primary));
    color: var(--color-cta-text, var(--color-cta-text, var(--color-primary-text, #fff)));

    /* Dark Section Variable Contract: the layout gives every heading
       color: var(--section-heading, ...), and a matched rule beats the
       inherited colour above — so a band painted in --color-primary must
       restate its own --section-* or its headings take the page's. */
    --section-text: var(--color-cta-text, var(--color-primary-text, #fff));
    --section-text-muted: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text, #fff)) 78%, transparent);
    --section-heading: var(--color-cta-text, var(--color-primary-text, #fff));
    --section-surface: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text, #fff)) 10%, transparent);
    --section-border: color-mix(in srgb, var(--color-cta-text, var(--color-primary-text, #fff)) 30%, transparent);
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
    border-inline-start: 1px solid color-mix(in srgb, var(--color-cta-text, var(--color-primary-text, #fff)) 22%, transparent);
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

-- info-card-grid — the icon slot accepts icon_image; cards carrying only an emoji `icon` render exactly as before.
UPDATE content_components SET html_template = $tpl$<style>
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
    color: var(--color-primary);
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
    color: var(--color-primary);
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
      <span class="info-card-grid__eyebrow">{{.section_eyebrow}}</span>
      <h2 class="info-card-grid__title">{{.section_title}}</h2>
      <p class="info-card-grid__subtitle">{{.section_subtitle}}</p>
    </header>
    <div class="info-card-grid__grid">
      {{range .cards}}
      <article class="info-card-grid__card">
        <div class="info-card-grid__card-icon{{if .icon_image}} info-card-grid__card-icon--img{{end}}" aria-hidden="true">{{if .icon_image}}<img class="info-card-grid__card-icon-img" src="{{.icon_image}}" alt="" loading="lazy" width="44" height="44">{{else}}{{.icon}}{{end}}</div>
        <h3 class="info-card-grid__card-title">{{.title}}</h3>
        <p class="info-card-grid__card-body">{{.body}}</p>
        {{if .link_url}}<a class="info-card-grid__card-link" href="{{.link_url}}">
          {{.link_label}}
          <em class="info-card-grid__card-link-arrow" aria-hidden="true">&rarr;</em>
        </a>{{end}}
      </article>
      {{end}}
    </div>
  </div>
</section>
$tpl$, updated_at = NOW()
 WHERE id = 'fc56f085-8e9a-4f6b-8e8d-600f9a1381e2';

COMMIT;
