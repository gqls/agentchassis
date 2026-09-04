-- 780_content_listing_gets_its_own_carousel_HOLD.sql
--
-- OWNER RULING 2026-09-04, verbatim (relayed by the `boxingonline.com` lane from the owner's
-- own words): asked for the latest-articles section on boxingonline to be a carousel, he
-- ruled *"it can have its own carousel forked from the best one and made better."*
--
-- So `content-listing` gets its OWN carousel — not a re-point at `info-card-grid`'s, not a
-- shared field. Forked from the best existing implementation and improved. Four parts:
-- the schema field, the template, the JS snippet's `applies_to`, and the one instance the
-- owner actually asked about.
--
-- ══ WHY THIS EXISTS: 740 WOULD HAVE CAROUSELLED THE WRONG SECTION ═══════════════════════
-- [MEASURED 2026-09-04] `boxingonline.com/index` carries THREE components:
--     pos 1  content-listing  "Latest from the site"   <- what he means. NO carousel field
--     pos 2  info-card-grid   "A few places to start"  <- HAS a carousel field
--     pos 3  call-to-action
-- Migration `740` (approved, unapplied) defaults info-card-grid's carousel ON. Applying it
-- alone would have put a carousel on his index page in "A few places to start" while "Latest
-- from the site" stayed a grid — readable as the request being actioned. This migration is
-- what actually answers the request.
--
-- ══ WHICH IS "THE BEST ONE"? hero-card-carousel, AND THE REASON IS SEMANTICS ════════════
-- The owner delegated that judgement. Both existing implementations were read, not counted.
--
--   `hero-card-carousel` (8,975 chars) — the ARIA carousel pattern, properly:
--       role="region" aria-roledescription="carousel" aria-label="<section title>"
--       per slide: role="group" aria-roledescription="slide" aria-label="<card title>"
--       a pause control, and <ul>/<li> track semantics
--   `info-card-grid` (11,916 chars) — when its carousel is ON its section emits
--       `data-hcc-carousel data-hcc-autoplay="false"` and NOTHING ELSE. No role, no
--       roledescription, no label; its cards get `data-hcc-slide` and no slide semantics.
--       A screen reader is told nothing about it being a carousel.
--
-- **So the SEMANTIC base is hero-card-carousel.** What info-card-grid contributes is its
-- OPT-IN GATING — everything carousel-related inside `{{if $.carousel}}`, so the off case is
-- byte-identical — which hero-card-carousel does not need and does not have, because it is
-- always a carousel. This fork takes the semantics from one and the gating from the other.
--
-- ══ "MADE BETTER" — ONE REAL IMPROVEMENT OVER BOTH, AND IT IS NOT COSMETIC ══════════════
-- **The carousel track is FOCUSABLE (`tabindex="0"`). Neither existing implementation is.**
-- The shared snippet binds its keyboard handler on the root:
--     root.addEventListener("keydown", function (e) {
--       if (e.key === "ArrowRight") { ... } else if (e.key === "ArrowLeft") { ... } });
-- (`js_snippets.hero-card-carousel`, lines ~80-83). A keydown listener only fires when focus
-- is inside the element — and NEITHER template makes anything in the carousel focusable, so
-- today those arrow keys are reachable only after tabbing onto an arrow BUTTON. Giving the
-- track a tab stop means a keyboard user can reach the carousel itself and page through it.
-- ⚠ Stated precisely because I nearly claimed the opposite: **the keyboard handler EXISTS**;
-- what was missing is anything focusable to deliver it. I checked the JS before claiming.
--
-- ══ `show_load_more` — MY DECISION, NOT HIS, AND FLAGGED AS SUCH ════════════════════════
-- The owner did not address it. A horizontal carousel with a "Load more" button underneath
-- is incoherent: the button implies a list that grows downward, the carousel implies a track
-- that scrolls sideways. **DECISION: the carousel suppresses it** —
-- `{{if and .show_load_more (not $.carousel)}}`. It is a one-key change if he disagrees, and
-- the boxingonline lane is putting it to him in one line rather than blocking on it.
-- ⚠ It does not bite today anyway: [MEASURED 2026-09-04] boxingonline's instance already has
-- `show_load_more: false`. This is forward-looking, not load-bearing for his request.
--
-- ══ DEFAULT-OFF, DELIBERATELY — the ruling does NOT license a fleet change ══════════════
-- He said *"it CAN HAVE ITS OWN"*, singular, answering a request about one section on one
-- site. [MEASURED 2026-09-04] `content-listing` is **19 instances across 11 sites**.
-- Defaulting eleven sites' article listings into carousels off one request is not what he
-- said, so: **no `fallback` on the field** (unlike 740), and `carousel: true` is set on
-- boxingonline's ONE instance explicitly. Fleet-wide is a separate decision with a
-- before/after read.
-- ⚠ AND A REASON NOT TO GENERALISE YET, from the boxingonline lane: `bugs_open/444` —
-- [MEASURED 2026-09-04] of 82 active index-family pages, 20 have sections but name NO
-- listing, and SIX share the identical signature ["hero","generic-text-block",
-- "call-to-action"] across four sites. **On those pages there is no list to carousel.**
--
-- ══ BYTE-IDENTITY WHEN OFF IS PROVEN, NOT CLAIMED ══════════════════════════════════════
-- [MEASURED 2026-09-04] Both templates rendered through Go `html/template` with the same
-- data and diffed:
--     carousel ABSENT,  load_more true   -> IDENTICAL
--     carousel ABSENT,  load_more false  -> IDENTICAL
--     carousel FALSE,   load_more true   -> IDENTICAL
--     carousel TRUE vs FALSE (control)   -> DIFFERENT  (so the flag is not inert)
-- and by count on the ON render: data-hcc-carousel 1, -track 1, -slide 3, -prev 1, -next 1,
-- -live 1; aria-roledescription="carousel" 1, ="slide" 3; tabindex="0" 1; "Load more" 0
-- (against 1 when off). Every one reads 0 on the OFF render.
--
-- ⚠ TWO THINGS THE RENDER CHECK CAUGHT THAT REVIEW WOULD NOT HAVE:
--  1. **A CSS COMMENT BECAME EXECUTABLE.** My style block explained the gating with the
--     literal text `{{if $.carousel}}` inside a `/* … */` comment. Go's parser does not know
--     it is a comment — it parsed it as a real action, opening an `if` that never closed, and
--     the template failed with `unexpected EOF`. **Never write a template action inside a
--     template's own comment.** It would have been an invalid template in the live row.
--  2. **A 1-CHARACTER DIFF THAT WAS THE INSTRUMENT, NOT THE TEMPLATE.** The first diff came
--     back 1 char short and looked like a real defect. It was `psql -At` appending a newline
--     to the REFERENCE file; the DB value is 1,526 chars and `right(html_template,1) = E'\n'`
--     is FALSE. Stripping it: identical. **A one-byte difference is exactly the size that
--     gets waved through as noise or panicked over as a bug — check which it is.**
--
-- ══ GUARDS PROVEN BY INDUCED FAILURE — FIVE, AND ONE WAS VACUOUS UNTIL FIXED ══════════
-- [MEASURED 2026-09-04] every guard mutated and watched go red, plus a full forward-then-
-- rollback round-trip in one transaction with the live state re-checked afterwards (template
-- 1,526 chars, no carousel field, applies_to unchanged, 0 instances flagged — nothing leaked):
--   * a `fallback` added to the field  -> "would default ALL 19 instances ON, which the owner
--     did not rule"                                    (this is the guard protecting the fleet)
--   * the instance UPDATE widened to every site -> "19 instances carry carousel=true, expected 1"
--   * `aria-roledescription="slide"` stripped from the MARKUP -> "lost the ARIA carousel
--     semantics — that is the whole reason hero-card-carousel was the fork base"
--   * the live template edited under the migration -> "not the byte-for-byte text this fork was
--     rendered and diffed against"
--   * `tabindex="0"` stripped -> "lost the focusable track" — **BUT ONLY AFTER THE NEEDLE WAS
--     NARROWED. See below.**
--
-- ⚠⚠ **ONE GUARD WAS VACUOUS AND THE MUTATION IS THE ONLY REASON I KNOW.** The focusable-track
-- assertion was written as `position('tabindex="0"' in tpl) = 0`. Stripping the attribute from
-- the markup left it **GREEN** — because this template's own CSS comment explains the focusable
-- track and contains that literal, so the needle matched my own prose. **A source-scanning
-- assertion makes your COMMENTS load-bearing**, and a needle that matches the comment passes
-- whatever the code does. Narrowed to `data-hcc-track tabindex="0"`, a string only the markup
-- can produce, and it now fires. **Every one of these guards is a string search over a template
-- that contains prose — check each needle against the comments, not just the markup.**
--
-- ⚠ AND A MUTATION CAN ITSELF BE INVALID. My first attempt at the ARIA mutation removed the
-- FIRST occurrence of the string in the .sql file — which was in this header's prose, not in
-- the template. The guard correctly stayed green and I nearly recorded that as "the guard does
-- not fire". **A mutation that does not change the artefact proves nothing; confirm the mutation
-- landed where you meant before believing its result.**
--
-- ⚠ NO DISPATCH FROM THIS LANE. boxingonline is a paying customer's site and every dispatch
-- there belongs to `site_delivery_and_editor`. This migration sets CONFIG only. The page is
-- already `build_status='needs_rebuild'`, so it will pick this up on the rebuild THEY run.
--
-- Reversible: 780_..._ROLLBACK.sql restores the template verbatim, drops the field, restores
-- applies_to and clears the instance flag.
-- Source: owner ruling 2026-09-04 via the boxingonline lane; scoping in this lane's handoff.

BEGIN;

DO $mig$
DECLARE
    n         int;
    comp_id   uuid;
    old_tpl   text;
    new_tpl   text := $ct${{if $.carousel}}<style>
  /* OPT-IN CAROUSEL LAYOUT for content-listing — emitted ONLY when content_data
     sets carousel:true. Everything in this block, and every `data-hcc-*`
     attribute in the markup below, is inside a carousel guard, so an
     instance that does not set it renders BYTE-IDENTICALLY to the pre-carousel
     template. That gating pattern is taken from info-card-grid; the SEMANTICS
     below are taken from hero-card-carousel, which is the better base. */
  .section--articles .article-grid__viewport { position: relative; }

  .section--articles .article-grid--carousel {
    display: grid;
    grid-auto-flow: column;
    grid-template-columns: none;
    grid-auto-columns: minmax(min(20rem, 82vw), 1fr);
    gap: var(--cl-track-gap, 1.5rem);
    align-items: stretch;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    scroll-behavior: smooth;
    padding-bottom: 0.75rem;
    scrollbar-width: thin;
  }
  .section--articles .article-grid--carousel > .article-card {
    scroll-snap-align: start;
    min-width: 0;
  }
  /* hover-lift translates the card, which pulls it out of its snap position
     mid-scroll on a trackpad and reads as jitter. The grid layout keeps the
     lift; the carousel drops it. Same reasoning as info-card-grid. */
  .section--articles .article-grid--carousel > .article-card.hover-lift:hover {
    transform: none;
  }
  /* IMPROVEMENT over both existing carousels: the track is focusable. The
     shared snippet binds its ArrowLeft/ArrowRight handler with
     root.addEventListener("keydown", …), and NEITHER hero-card-carousel nor
     info-card-grid makes anything in the carousel focusable — so today a
     keyboard user reaches those keys only after tabbing onto an arrow button.
     tabindex="0" on the track gives the carousel itself a tab stop. */
  .section--articles .article-grid--carousel:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 4px;
  }

  @media (min-width: 60rem) {
    .section--articles .article-grid--carousel { grid-auto-columns: minmax(17rem, 1fr); }
  }
  /* The base grid--3 rule collapses to one column on mobile. With
     grid-auto-flow:column that would create one explicit column and size every
     other card off it, so the carousel restores `none` at the same breakpoint.
     Equal specificity, later in source order — do not move this above the
     base stylesheet. */
  @media (max-width: 768px) {
    .section--articles .article-grid--carousel {
      grid-template-columns: none;
      gap: var(--cl-track-gap, 1.5rem);
    }
  }

  .section--articles .article-grid__arrow {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    z-index: 2;
    inline-size: var(--cl-arrow-size, 44px);
    block-size: var(--cl-arrow-size, 44px);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--color-border);
    background: var(--color-background);
    color: var(--color-text);
    border-radius: 999px;
    font-size: 1.5rem;
    line-height: 1;
    cursor: pointer;
    box-shadow: var(--shadow, 0 4px 6px rgba(0,0,0,0.07));
    transition: background 0.18s ease, color 0.18s ease;
  }
  .section--articles .article-grid__arrow--prev { left: -1.1rem; }
  .section--articles .article-grid__arrow--next { right: -1.1rem; }
  .section--articles .article-grid__arrow:hover {
    background: var(--color-primary);
    color: var(--color-card-bg, #fff);
  }
  .section--articles .article-grid__arrow:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
  }
  /* Native swipe and scroll-snap cover touch, and at this width an overlaid
     arrow sits on the card text. A deliberate degradation, not a gap. */
  @media (max-width: 40rem) {
    .section--articles .article-grid__arrow { display: none; }
  }

  .section--articles .article-grid__live {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
  }

  @media (prefers-reduced-motion: reduce) {
    .section--articles .article-grid--carousel { scroll-behavior: auto; }
  }
</style>
{{end}}<section class="section section--articles"{{if $.carousel}} data-hcc-carousel data-hcc-autoplay="false" role="region" aria-roledescription="carousel" aria-label="{{if .section_title}}{{.section_title}}{{else}}Latest articles{{end}}"{{end}}>
      <div class="container">
        {{if or .section_title .section_subtitle}}<div class="section__header">
          {{if .section_title}}<h2 class="section__title">{{.section_title}}</h2>{{end}}
          {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
        </div>{{end}}
        {{if $.carousel}}<div class="article-grid__viewport">
          <button type="button" class="article-grid__arrow article-grid__arrow--prev" data-hcc-prev aria-label="Previous article"><span aria-hidden="true">&lsaquo;</span></button>
        {{end}}<div class="article-grid grid grid--3{{if $.carousel}} article-grid--carousel{{end}}"{{if $.carousel}} data-hcc-track tabindex="0"{{end}}>
          {{range .articles}}
          <article class="article-card hover-lift"{{if $.carousel}} data-hcc-slide role="group" aria-roledescription="slide" aria-label="{{.title}}"{{end}}>
            {{if .image}}<div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
              {{if .category}}<span class="article-card__category">{{.category}}</span>{{end}}
            </div>{{end}}
            <div class="article-card__content">
              <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
              {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
              {{if or .date .read_time}}<div class="article-card__meta">
                {{if .date}}<span class="article-card__date">{{.date}}</span>{{end}}
                {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
              </div>{{end}}
            </div>
          </article>
          {{end}}
        </div>{{if $.carousel}}
          <button type="button" class="article-grid__arrow article-grid__arrow--next" data-hcc-next aria-label="Next article"><span aria-hidden="true">&rsaquo;</span></button>
          <span class="article-grid__live" aria-live="polite" data-hcc-live></span>
        </div>{{end}}
        {{if and .show_load_more (not $.carousel)}}
        <div class="section__actions">
          <button class="button button--secondary">{{.load_more_text}}</button>
        </div>
        {{end}}
      </div>
    </section>$ct$;
BEGIN
    -- DRIFT GUARDS. Abort rather than clobber.
    SELECT count(*) INTO n FROM content_components WHERE is_active AND name = 'content-listing';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: expected exactly 1 active content-listing, found %', n;
    END IF;

    SELECT id, html_template INTO comp_id, old_tpl
      FROM content_components WHERE is_active AND name = 'content-listing';

    IF old_tpl IS DISTINCT FROM $ct$<section class="section section--articles">
      <div class="container">
        {{if or .section_title .section_subtitle}}<div class="section__header">
          {{if .section_title}}<h2 class="section__title">{{.section_title}}</h2>{{end}}
          {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
        </div>{{end}}
        <div class="article-grid grid grid--3">
          {{range .articles}}
          <article class="article-card hover-lift">
            {{if .image}}<div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
              {{if .category}}<span class="article-card__category">{{.category}}</span>{{end}}
            </div>{{end}}
            <div class="article-card__content">
              <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
              {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
              {{if or .date .read_time}}<div class="article-card__meta">
                {{if .date}}<span class="article-card__date">{{.date}}</span>{{end}}
                {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
              </div>{{end}}
            </div>
          </article>
          {{end}}
        </div>
        {{if .show_load_more}}
        <div class="section__actions">
          <button class="button button--secondary">{{.load_more_text}}</button>
        </div>
        {{end}}
      </div>
    </section>$ct$ THEN
        RAISE EXCEPTION 'ABORT: content-listing template is not the byte-for-byte text this fork '
                        'was rendered and diffed against (live % chars). Someone has edited it; '
                        're-run the render diff before applying.', length(old_tpl);
    END IF;

    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name = 'content-listing' AND input_schema->'fields' ? 'carousel';
    IF n <> 0 THEN
        RAISE EXCEPTION 'ABORT: content-listing already declares a carousel field — already applied?';
    END IF;

    -- PART 1: the schema field. NO `fallback` — default OFF is deliberate (see header).
    UPDATE content_components
       SET input_schema = jsonb_set(input_schema, '{fields,carousel}',
             jsonb_build_object(
               'type', 'boolean',
               'source', 'static',
               'required', false,
               'llm_guidance',
               'Optional. Set true to lay the article cards out as a single-row horizontal '
               'carousel with prev/next arrows and a keyboard-focusable track, instead of a '
               'wrapping grid. Absent or false renders byte-identically to the grid layout, so '
               'an existing instance is never changed by this field existing. When true the '
               'load-more button is suppressed: a carousel that scrolls sideways and a button '
               'that grows a list downward are incoherent together. Requires the '
               'hero-card-carousel js_snippet, which applies_to this component.'),
             true),
           html_template = new_tpl,
           updated_at = now()
     WHERE id = comp_id;

    -- PART 2: the JS snippet must reach this component. Its root selector already exposes
    -- `[data-hcc-carousel]` as a deliberate opt-in hook for other components, so no JS change
    -- is needed — only the applies_to that decides which sites get the bundle.
    SELECT count(*) INTO n FROM js_snippets
     WHERE is_active AND name = 'hero-card-carousel'
       AND applies_to @> '["content-listing"]'::jsonb;
    IF n <> 0 THEN
        RAISE EXCEPTION 'ABORT: applies_to already names content-listing';
    END IF;
    UPDATE js_snippets
       SET applies_to = applies_to || '["content-listing"]'::jsonb
     WHERE is_active AND name = 'hero-card-carousel';

    RAISE NOTICE '780: content-listing has its own carousel (opt-in, default OFF). '
                 'Template % -> % chars.', length(old_tpl), length(new_tpl);
END $mig$;

-- PART 3: turn it on for the ONE instance the owner asked about, and only that one.
UPDATE page_components pc
   SET content_data = jsonb_set(pc.content_data, '{carousel}', 'true'::jsonb, true),
       updated_at = now()
  FROM pages p, sites s, content_components cc
 WHERE p.id = pc.page_id AND s.id = p.site_id AND cc.id = pc.component_id
   AND s.domain = 'boxingonline.com' AND p.name = 'index'
   AND cc.name = 'content-listing' AND p.status = 'active';

-- VERIFY. Separate block, re-reading the LIVE rows — not the variables above.
DO $$
DECLARE n int; tpl text;
BEGIN
    SELECT html_template INTO tpl FROM content_components
     WHERE is_active AND name='content-listing';

    IF position('data-hcc-carousel' in tpl) = 0 THEN
        RAISE EXCEPTION 'ABORT: the live template does not carry data-hcc-carousel';
    END IF;
    IF position('aria-roledescription="carousel"' in tpl) = 0
       OR position('aria-roledescription="slide"' in tpl) = 0 THEN
        RAISE EXCEPTION 'ABORT: the live template lost the ARIA carousel semantics — that is '
                        'the whole reason hero-card-carousel was the fork base';
    END IF;
    -- ⚠ NEEDLE CHOSEN TO BE NON-VACUOUS. The obvious assertion — position('tabindex="0"') — 
    -- PASSES with the attribute deleted, because this template's own CSS comment explains the
    -- focusable track and contains that literal. Proven by mutation: stripping the attribute
    -- left the guard green. The needle must be a string only the MARKUP can produce.
    IF position('data-hcc-track tabindex="0"' in tpl) = 0 THEN
        RAISE EXCEPTION 'ABORT: the live template lost the focusable track — that is the '
                        '"made better" the owner asked for';
    END IF;
    IF position('{{if and .show_load_more (not $.carousel)}}' in tpl) = 0 THEN
        RAISE EXCEPTION 'ABORT: the load-more suppression is not in the live template';
    END IF;

    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name='content-listing'
       AND input_schema->'fields'->'carousel'->>'source' = 'static'
       AND input_schema->'fields'->'carousel'->>'type' = 'boolean';
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: the carousel field did not land as a static boolean'; END IF;

    -- Default-OFF is load-bearing: a fallback here would flip all 19 instances.
    IF EXISTS (SELECT 1 FROM content_components WHERE is_active AND name='content-listing'
                AND input_schema->'fields'->'carousel' ? 'fallback') THEN
        RAISE EXCEPTION 'ABORT: the carousel field has a fallback — that would default ALL '
                        '19 instances ON, which the owner did not rule';
    END IF;

    SELECT count(*) INTO n FROM js_snippets
     WHERE is_active AND name='hero-card-carousel' AND applies_to @> '["content-listing"]'::jsonb;
    IF n <> 1 THEN RAISE EXCEPTION 'ABORT: applies_to does not name content-listing'; END IF;

    -- Exactly ONE instance turned on, and it is the right one.
    SELECT count(*) INTO n FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
     WHERE cc.name='content-listing' AND (pc.content_data->>'carousel') = 'true';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: % content-listing instances carry carousel=true, expected exactly 1', n;
    END IF;
    SELECT count(*) INTO n FROM page_components pc
      JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
      JOIN content_components cc ON cc.id=pc.component_id
     WHERE cc.name='content-listing' AND (pc.content_data->>'carousel')='true'
       AND s.domain='boxingonline.com' AND p.name='index';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: the one enabled instance is NOT boxingonline.com/index';
    END IF;

    RAISE NOTICE '780 VERIFY: template forked with ARIA semantics + focusable track, field is '
                 'static boolean with NO fallback, applies_to names content-listing, and exactly '
                 'one instance (boxingonline.com/index) is switched on.';
END $$;

COMMIT;
