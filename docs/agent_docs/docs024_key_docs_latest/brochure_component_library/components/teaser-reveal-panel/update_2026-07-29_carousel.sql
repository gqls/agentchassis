\set ON_ERROR_STOP on
-- Re-apply teaser-reveal-panel's template + input_schema to the LIVE
-- content_components row: carousel arrows (reusing hero-card-carousel's
-- goTo/nearestIndex pattern), a single row at every viewport width (removes
-- the desktop grid-wrap that produced 2 rows of cards), extra card padding, a
-- CSS-drawn (never stored) ellipsis on the closed continuation, and an open
-- state where continuation completes into body with no visible seam and the
-- "Read the rest" control disappears rather than dangling with nothing left
-- to invite. GENERATED from components/teaser-reveal-panel/{template.html,
-- input_schema.json} -- edit those files and regenerate, do not hand-edit.
--
-- input_schema is unchanged this round (no new content fields) -- only
-- template.html and behaviour.js changed, so no page's content_data needs
-- touching, only a page_rerender to pick up the new template.
BEGIN;

DROP TABLE IF EXISTS bak_cc_teaser_reveal_panel_pre_carousel_update;
CREATE TABLE bak_cc_teaser_reveal_panel_pre_carousel_update AS
SELECT * FROM content_components WHERE function = 'teaser-reveal-panel';

UPDATE content_components
   SET html_template = $HTML$<style>
  /* Colour vocabulary verified against live css_themes before use (the
     --color-surface / --spacing-section class of error: those names are defined
     by NO active theme, so the fallback silently wins). Every var below was
     confirmed present in an active theme on 2026-07-29. */
  .trp { background: var(--color-background); color: var(--color-text); padding: var(--spacing-xl) 0; }
  .trp__inner { max-width: 1200px; margin: 0 auto; padding: 0 var(--spacing-lg); position: relative; }
  .trp__eyebrow { color: var(--color-accent); text-transform: uppercase; letter-spacing: .08em;
                  font-size: .78rem; font-weight: 700; margin: 0 0 .5rem; }
  .trp__title { color: var(--color-text); font-size: clamp(1.5rem, 3vw, 2.1rem); margin: 0 0 var(--spacing-lg); }

  /* Viewport wraps the track + the overlaid arrows. A carousel, not a wrapping
     grid, at every width: grid-auto-flow stays "column" always. Only the
     column width changes at wider viewports, so more cards sit in one row --
     this is what "always one row, never two" means structurally. */
  .trp__viewport { position: relative; }
  .trp__track { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(min(20rem, 82vw), 1fr);
                gap: var(--spacing-lg); align-items: start; overflow-x: auto; scroll-snap-type: x mandatory;
                padding-bottom: var(--spacing-md); margin: 0; list-style: none;
                scrollbar-width: thin; scroll-behavior: smooth; }
  .trp__slot { scroll-snap-align: start; min-width: 0; }
  @media (min-width: 60rem) {
    .trp__track { grid-auto-columns: minmax(17rem, 1fr); }
  }

  /* Overlaid prev/next arrows, same visual language as hero-card-carousel's.
     Fixed vertical position, not a percentage of the (variable-height once a
     card opens) track -- see .trp__body's max-height note below for why the
     height that position is measured against never actually changes. */
  .trp__arrow { position: absolute; top: 6.5rem; transform: translateY(-50%); z-index: 2;
                inline-size: 44px; block-size: 44px; display: inline-flex; align-items: center;
                justify-content: center; border: 1px solid var(--color-border);
                background: var(--color-background); color: var(--color-text); border-radius: 999px;
                font-size: 1.5rem; line-height: 1; cursor: pointer;
                box-shadow: var(--shadow); transition: background .18s ease, color .18s ease; }
  .trp__arrow--prev { left: -1.1rem; }
  .trp__arrow--next { right: -1.1rem; }
  .trp__arrow:hover { background: var(--color-accent); color: var(--color-card-bg); }
  .trp__arrow:focus-visible { outline: 2px solid var(--color-accent); outline-offset: 2px; }
  @media (max-width: 40rem) {
    .trp__arrow { display: none; } /* native swipe/scroll-snap covers touch; arrows would sit over content */
  }

  /* Card. box-sizing + min-width:0 travel with any max-width in a grid track.
     overflow:hidden so an edge-to-edge media image respects the rounded corners. */
  .trp__card { box-sizing: border-box; min-width: 0; background: var(--color-card-bg);
               border: 1px solid var(--color-border); border-radius: var(--border-radius);
               box-shadow: var(--shadow); overflow: hidden; }
  .trp__media { display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: cover;
                background: var(--color-background); }
  .trp__text { display: block; padding: var(--spacing-xl) var(--spacing-lg); }
  .trp__summary { list-style: none; cursor: pointer; display: block; }
  .trp__summary::-webkit-details-marker { display: none; }
  .trp__card--static { cursor: default; }
  .trp__hook { display: block; color: var(--color-text); font-weight: 700;
               font-size: 1.12rem; line-height: 1.4; }

  /* Continuation carries a DECORATIVE ellipsis, drawn by CSS, never stored as a
     real character. It exists only in the rendered pixel, not in the HTML text
     node -- the claims gate and any truncation-vs-damage check both read text
     nodes, so this is invisible to both, unlike a literal "..." in the data
     (which the platform's own no-ellipsis rule forbids for exactly that
     reason: a checker built on output_tokens==max_tokens reads a trailing
     ellipsis as a sign of a cut-off generation). Hidden once open, because an
     ellipsis on text that has just finished completing itself is a lie. */
  .trp__continuation { display: block; color: var(--color-text-muted); margin-top: .6rem; line-height: 1.6; }
  .trp__continuation::after { content: "\2026"; margin-left: .15em; }
  .trp__card[open] .trp__continuation::after { content: none; }
  /* Once open, continuation stops looking like a fragment and reads as the
     start of the same paragraph body completes -- full-strength text colour,
     not muted. */
  .trp__card[open] .trp__continuation { color: var(--color-text); }

  /* The control is visible ONLY while closed. Once open, body picks up
     exactly where continuation left off with no visible seam -- no border, no
     extra gap, no control sitting there with nothing left to invite. Closing
     is still one click away: the whole <summary> (image, hook, continuation)
     stays the native toggle target. */
  .trp__control { display: inline-flex; align-items: center; gap: .4rem; margin-top: .9rem;
                  color: var(--color-accent); font-weight: 600; font-size: .93rem; }
  .trp__control::after { content: "\2193"; }
  .trp__card[open] .trp__control { display: none; }
  .trp__summary:focus-visible { outline: 2px solid var(--color-accent); outline-offset: 3px; }

  /* Bounded height + internal scroll, ALWAYS, open or closed. This is the
     answer to "how does a dropdown behave inside a horizontally-scrolling
     carousel": if an open card could grow without limit, the whole track's
     row height would grow with it, dragging the overlaid arrows (positioned
     at a fixed offset from the top) out of alignment with every open/close.
     Capping it means the arrows never need to move and the track height never
     jumps. In practice every body written so far is 1-2 sentences and never
     reaches this cap -- it is a safety bound, not an active constraint. */
  .trp__body { padding: 0 var(--spacing-lg) var(--spacing-xl); margin-top: -.3rem;
               color: var(--color-text); line-height: 1.6; max-height: 12rem;
               overflow-y: auto; overscroll-behavior: contain; }
  .trp__body p { margin: 0; }
  @media (prefers-reduced-motion: reduce) {
    .trp__track { scroll-behavior: auto; }
  }

  .trp__live { position: absolute; width: 1px; height: 1px; overflow: hidden;
               clip: rect(0 0 0 0); white-space: nowrap; }
</style>

<section class="trp" data-component="teaser-reveal-panel" data-trp-param="open" data-experience-pattern="teaser-detail-deeplink">
  <div class="trp__inner">
    {{if .section_eyebrow}}<p class="trp__eyebrow">{{.section_eyebrow}}</p>{{end}}
    {{if .section_title}}<h2 class="trp__title">{{.section_title}}</h2>{{end}}
    <div class="trp__viewport">
      <button type="button" class="trp__arrow trp__arrow--prev" data-trp-prev aria-label="Previous card"><span aria-hidden="true">&lsaquo;</span></button>
      <ul class="trp__track" data-trp-track>
        {{range $i, $item := .items}}
        <li class="trp__slot" data-trp-slide>
          {{if $item.body}}
          <details class="trp__card" data-trp-key="{{$item.key}}">
            <summary class="trp__summary">
              {{if $item.image_url}}<img class="trp__media" src="{{$item.image_url}}" alt="{{$item.image_alt}}" loading="lazy">{{end}}
              <span class="trp__text">
                <span class="trp__hook">{{$item.hook}}</span>
                <span class="trp__continuation" data-continues="true">{{$item.continuation}}</span>
                <span class="trp__control">{{if $item.open_label}}{{$item.open_label}}{{else}}Read the rest{{end}}</span>
              </span>
            </summary>
            <div class="trp__body"><p>{{$item.body}}</p></div>
          </details>
          {{else}}
          <article class="trp__card trp__card--static" data-trp-key="{{$item.key}}">
            {{if $item.image_url}}<img class="trp__media" src="{{$item.image_url}}" alt="{{$item.image_alt}}" loading="lazy">{{end}}
            <span class="trp__text">
              <span class="trp__hook">{{$item.hook}}</span>
              <span class="trp__continuation">{{$item.continuation}}</span>
            </span>
          </article>
          {{end}}
        </li>
        {{end}}
      </ul>
      <button type="button" class="trp__arrow trp__arrow--next" data-trp-next aria-label="Next card"><span aria-hidden="true">&rsaquo;</span></button>
      <span class="trp__live" aria-live="polite" data-trp-live></span>
    </div>
  </div>
</section>
$HTML$,
       input_schema  = $SCHEMA${
  "fields": {
    "section_eyebrow": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Short uppercase eyebrow, 2-4 words. Optional."
    },
    "section_title": {
      "type": "text",
      "source": "llm",
      "required": false,
      "llm_guidance": "Short heading for the panel, under 10 words."
    },
    "items": {
      "type": "array",
      "source": "llm",
      "required": true,
      "on_missing": "skip_section",
      "missing_reason": "a teaser panel with no items is an empty shell; skip the section rather than render a heading over nothing",
      "llm_guidance": "3 to 6 items. Each item is a teaser that opens in place. Write hook and continuation as ONE thought split across two fields, and put the completion in body. HARD RULES: (1) Never write a figure, percentage, count or date in hook or continuation. Any number and the words that give it meaning must sit together inside body, because a checker reads a number and its surrounding context as one unit and a split makes a true figure look invented. (2) continuation must be an incomplete clause that body genuinely completes. Never end it with an ellipsis, three dots or any other punctuation trick; the incompleteness is carried in the data, not in the typography. (3) Only tease what you can deliver. If you have no body for an item, write continuation as a COMPLETE sentence and omit body entirely; the item then renders as a plain statement with no control. Never write a teaser whose promise the body does not answer. (4) image_url is optional per item, but if set, image_alt MUST also be set to a genuine description of what the image shows \u2014 never the hook restated, because a screen reader will read hook immediately after alt and a repeated phrase is redundant, not descriptive.",
      "items": {
        "key": {
          "type": "text",
          "llm_guidance": "Short lowercase-kebab identifier, unique within the panel. It appears in the URL when the item is open, so keep it readable."
        },
        "hook": {
          "type": "text",
          "llm_guidance": "One very short complete sentence. Under 12 words. It must stand alone."
        },
        "continuation": {
          "type": "text",
          "llm_guidance": "The deliberately unfinished second sentence, completed by body. Under 20 words. No ellipsis."
        },
        "body": {
          "type": "text",
          "llm_guidance": "The full text revealed on activation. Optional: an item with no body is legitimate and renders as a plain statement."
        },
        "open_label": {
          "type": "text",
          "llm_guidance": "Optional label for the control, e.g. 'Read the rest'. Defaults to 'Read the rest'."
        },
        "image_url": {
          "type": "text",
          "required": false,
          "llm_guidance": "Optional path to an existing site image, e.g. /assets/images/<name>.jpg. Never invent a filename; only reference an image known to exist."
        },
        "image_alt": {
          "type": "text",
          "required": false,
          "llm_guidance": "Required whenever image_url is set. A genuine description of the image's content, not a restatement of hook."
        }
      }
    }
  }
}$SCHEMA$::jsonb
 WHERE function = 'teaser-reveal-panel';

COMMIT;

SELECT function, is_active, length(html_template) AS template_bytes,
       html_template LIKE '%data-trp-prev%' AS has_arrows
  FROM content_components WHERE function = 'teaser-reveal-panel';
