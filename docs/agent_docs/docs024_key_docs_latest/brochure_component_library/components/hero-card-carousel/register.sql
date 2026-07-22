\set ON_ERROR_STOP on
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, js_content, input_schema)
VALUES (
  gen_random_uuid(),
  'hero-card-carousel', 'hero-card-carousel', 'Hero Card Carousel',
  'Auto-advancing, swipeable hero carousel of a few cards — each a hover-zoom image with a short title, one teaser line and a read-more link. Accessible: pausable, keyboard-safe, respects reduced-motion; first card carries the full message.',
  'hero', '["hero","carousel","interactive","cards","brochure"]'::jsonb,
  'hero-carousel', 'section', 'agent', false, true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","landing","about","capabilities"]'::jsonb,
  $HTML$
<style>
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

  /* Controls — placed before the track in DOM for tab order (WCAG APG). */
  .hero-card-carousel__controls {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.25rem;
  }
  .hero-card-carousel__btn {
    inline-size: 44px;
    block-size: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--color-border);
    background: var(--color-surface, transparent);
    color: var(--color-heading);
    border-radius: 999px;
    font-size: 1.25rem;
    line-height: 1;
    cursor: pointer;
    transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
  }
  .hero-card-carousel__btn:hover {
    background: var(--color-primary);
    border-color: var(--color-primary);
    color: var(--color-primary-text, #fff);
  }
  .hero-card-carousel__btn:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }

  /* Track — native scroll-snap gives swipe-on-mobile with zero JS; the JS only
     drives auto-advance and the buttons. */
  .hero-card-carousel__track {
    list-style: none;
    margin: 0;
    padding: 0.5rem 0 1.5rem;
    display: flex;
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

  .hero-card-carousel__card {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--color-surface, transparent);
    border: 1px solid var(--color-hairline, var(--color-border));
    border-radius: var(--border-radius, 0.75rem);
    overflow: hidden;
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
  .hero-card-carousel__card:focus-within .hero-card-carousel__img {
    scale: 1.08;
  }

  .hero-card-carousel__body {
    padding: 1.5rem 1.5rem 1.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    flex-grow: 1;
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
    flex-grow: 1;
  }
  .hero-card-carousel__link {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-primary);
    text-decoration: none;
    min-height: 44px;
    margin-top: auto;
  }
  .hero-card-carousel__link:hover { text-decoration: underline; }
  .hero-card-carousel__link:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
    border-radius: 2px;
  }

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
    .hero-card-carousel__card:focus-within .hero-card-carousel__img { scale: 1; }
  }
  @media (hover: none) {
    .hero-card-carousel__card:hover .hero-card-carousel__img { scale: 1; }
  }
  @media (max-width: 768px) {
    .hero-card-carousel { padding: 3rem 1.25rem; }
  }
</style>

<section class="hero-card-carousel" data-component="hero-card-carousel" role="region" aria-roledescription="carousel" aria-label="{{if .section_title}}{{.section_title}}{{else}}Featured{{end}}">
  <div class="hero-card-carousel__inner">
    <div class="hero-card-carousel__head">
      <div>
        {{if .section_eyebrow}}<span class="hero-card-carousel__eyebrow">{{.section_eyebrow}}</span>{{end}}
        {{if .section_title}}<h2 class="hero-card-carousel__title">{{.section_title}}</h2>{{end}}
      </div>
      <div class="hero-card-carousel__controls" data-hcc-controls>
        <button type="button" class="hero-card-carousel__btn" data-hcc-pause aria-label="Pause automatic rotation"><span data-hcc-pause-icon aria-hidden="true">&#10073;&#10073;</span></button>
        <button type="button" class="hero-card-carousel__btn" data-hcc-prev aria-label="Previous card"><span aria-hidden="true">&lsaquo;</span></button>
        <button type="button" class="hero-card-carousel__btn" data-hcc-next aria-label="Next card"><span aria-hidden="true">&rsaquo;</span></button>
      </div>
    </div>

    <ul class="hero-card-carousel__track" data-hcc-track>
      {{range $i, $card := .cards}}
      <li class="hero-card-carousel__slide" role="group" aria-roledescription="slide" aria-label="{{$card.title}}" data-hcc-slide>
        <article class="hero-card-carousel__card">
          <div class="hero-card-carousel__media">
            <img class="hero-card-carousel__img" src="{{if $card.image}}{{$card.image}}{{else}}/assets/images/hero.jpg{{end}}" alt="{{$card.image_alt}}" width="800" height="500" loading="{{if eq $i 0}}eager{{else}}lazy{{end}}">
          </div>
          <div class="hero-card-carousel__body">
            <h3 class="hero-card-carousel__card-title">{{$card.title}}</h3>
            <p class="hero-card-carousel__teaser">{{$card.teaser}}</p>
            {{if $card.link_url}}<a class="hero-card-carousel__link" href="{{$card.link_url}}">{{if $card.link_label}}{{$card.link_label}}{{else}}Read more{{end}}<span aria-hidden="true">&nbsp;&rarr;</span></a>{{end}}
          </div>
        </article>
      </li>
      {{end}}
    </ul>

    <div class="hero-card-carousel__live" aria-live="polite" data-hcc-live></div>
  </div>
</section>

$HTML$,
  $JS$
/* hero-card-carousel — accessible auto-advancing card carousel.
   WCAG 2.2.2: auto-rotation is pausable, stops on hover and keyboard focus,
   respects prefers-reduced-motion, and every control is a real <button>.
   Self-contained, supports multiple instances, no dependencies. */
(function () {
  "use strict";
  var ROTATE_MS = 6000;
  var reduceMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  function initCarousel(root) {
    var track = root.querySelector("[data-hcc-track]");
    if (!track) return;
    var slides = Array.prototype.slice.call(root.querySelectorAll("[data-hcc-slide]"));
    if (slides.length < 2) {
      var c = root.querySelector("[data-hcc-controls]");
      if (c) c.style.display = "none";
      return;
    }
    var pauseBtn = root.querySelector("[data-hcc-pause]");
    var pauseIcon = root.querySelector("[data-hcc-pause-icon]");
    var prevBtn = root.querySelector("[data-hcc-prev]");
    var nextBtn = root.querySelector("[data-hcc-next]");
    var live = root.querySelector("[data-hcc-live]");

    var current = 0;
    var paused = false;      // user pressed pause
    var suspended = false;   // hover / focus (temporary)
    var timer = null;

    function behavior() { return reduceMotion ? "auto" : "smooth"; }

    function goTo(i, announce) {
      current = (i + slides.length) % slides.length;
      var trackRect = track.getBoundingClientRect();
      var slideRect = slides[current].getBoundingClientRect();
      track.scrollBy({ left: slideRect.left - trackRect.left, behavior: behavior() });
      if (announce && live) live.textContent = "Card " + (current + 1) + " of " + slides.length;
    }

    function nearestIndex() {
      var trackLeft = track.getBoundingClientRect().left;
      var best = 0, bestDist = Infinity;
      for (var i = 0; i < slides.length; i++) {
        var d = Math.abs(slides[i].getBoundingClientRect().left - trackLeft);
        if (d < bestDist) { bestDist = d; best = i; }
      }
      return best;
    }

    function stopTimer() { if (timer) { clearInterval(timer); timer = null; } }
    function startTimer() {
      stopTimer();
      if (reduceMotion || paused || suspended) return;
      timer = setInterval(function () { goTo(current + 1, false); }, ROTATE_MS);
    }

    function setPaused(p) {
      paused = p;
      if (pauseBtn) pauseBtn.setAttribute("aria-label", p ? "Start automatic rotation" : "Pause automatic rotation");
      if (pauseIcon) pauseIcon.innerHTML = p ? "&#9654;" : "&#10073;&#10073;";
      startTimer();
    }

    if (pauseBtn) pauseBtn.addEventListener("click", function () { setPaused(!paused); });
    if (prevBtn) prevBtn.addEventListener("click", function () { goTo(current - 1, true); });
    if (nextBtn) nextBtn.addEventListener("click", function () { goTo(current + 1, true); });

    // Pause on hover and on keyboard focus entering the carousel; resume on leave.
    root.addEventListener("pointerenter", function () { suspended = true; stopTimer(); });
    root.addEventListener("pointerleave", function () { suspended = false; startTimer(); });
    root.addEventListener("focusin", function () { suspended = true; stopTimer(); });
    root.addEventListener("focusout", function () {
      if (!root.contains(document.activeElement)) { suspended = false; startTimer(); }
    });

    // Keyboard arrows when the track (or a control) has focus.
    root.addEventListener("keydown", function (e) {
      if (e.key === "ArrowRight") { e.preventDefault(); goTo(current + 1, true); }
      else if (e.key === "ArrowLeft") { e.preventDefault(); goTo(current - 1, true); }
    });

    // Keep `current` in sync when the user swipes/scrolls manually.
    var scrollT = null;
    track.addEventListener("scroll", function () {
      if (scrollT) clearTimeout(scrollT);
      scrollT = setTimeout(function () { current = nearestIndex(); }, 120);
    }, { passive: true });

    // Pause when the carousel scrolls out of view (don't rotate off-screen).
    if ("IntersectionObserver" in window) {
      new IntersectionObserver(function (entries) {
        entries.forEach(function (en) {
          if (en.isIntersecting) { suspended = false; startTimer(); }
          else { suspended = true; stopTimer(); }
        });
      }, { threshold: 0.25 }).observe(root);
    } else {
      startTimer();
    }
  }

  function initAll() {
    Array.prototype.slice.call(document.querySelectorAll(".hero-card-carousel[data-component='hero-card-carousel']")).forEach(initCarousel);
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", initAll);
  else initAll();
})();

$JS$,
  $SCHEMA${
  "fields": {
    "section_eyebrow": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Short uppercase eyebrow label above the heading, e.g. 'What we do' or 'Featured work'. Under 4 words. Optional."
    },
    "section_title": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Section heading that frames the featured cards in one compelling phrase. Under 9 words."
    },
    "cards": {
      "type": "array",
      "source": "llm",
      "required": true,
      "items": {
        "image": { "type": "image" },
        "image_alt": { "type": "text" },
        "title": { "type": "text" },
        "teaser": { "type": "text" },
        "link_url": { "type": "url" },
        "link_label": { "type": "text" }
      },
      "llm_guidance": "3 to 5 hero cards. The FIRST card must carry the complete headline message on its own — a visitor who sees only the first card should still get the point (research: about 89% of carousel viewers only ever see slide one, so never hide anything essential on later cards). Each card: image_alt (a concise description of the line-illustration, for screen readers), title (a 3-6 word heading), teaser (ONE short line, max 16 words), link_url (relative or absolute URL for the read-more link), link_label (short CTA text, e.g. 'Read more'). Keep card copy minimal — a hero card is a title, one teaser line and a link, nothing more."
    }
  }
}
$SCHEMA$::jsonb
)
RETURNING function, section_type, component_level, render_mode, is_active, length(html_template) AS tpl_len, length(js_content) AS js_len;
