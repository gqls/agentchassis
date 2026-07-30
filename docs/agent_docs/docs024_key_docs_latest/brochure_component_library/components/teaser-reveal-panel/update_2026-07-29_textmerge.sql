\set ON_ERROR_STOP on
-- Re-apply teaser-reveal-panel's template to the LIVE content_components row:
-- fixes the padding mismatch between the closed hook and the opened body, and
-- replaces the whole closed text block (hook+continuation+control) with ONE
-- flowing paragraph (hook as bold lead + continuation + body, concatenated
-- with no line break at the old cut point) when a card opens, instead of
-- appending body below a continuation that still showed its own cut-off line.
-- GENERATED from components/teaser-reveal-panel/template.html -- edit that
-- file and regenerate, do not hand-edit this file. input_schema unchanged.
BEGIN;

DROP TABLE IF EXISTS bak_cc_teaser_reveal_panel_pre_textmerge_update;
CREATE TABLE bak_cc_teaser_reveal_panel_pre_textmerge_update AS
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
     ellipsis as a sign of a cut-off generation). */
  .trp__continuation { display: block; color: var(--color-text-muted); margin-top: .6rem; line-height: 1.6; }
  .trp__continuation::after { content: "\2026"; margin-left: .15em; }

  .trp__control { display: inline-flex; align-items: center; gap: .4rem; margin-top: .9rem;
                  color: var(--color-accent); font-weight: 600; font-size: .93rem; }
  .trp__control::after { content: "\2193"; }
  .trp__summary:focus-visible { outline: 2px solid var(--color-accent); outline-offset: 3px; }

  /* Opening a card REPLACES the whole closed block (image stays; hook,
     continuation and the control all disappear as one unit) rather than
     appending body below a continuation that still shows its own cut-off
     line. What replaces it is ONE flowing paragraph -- hook as a bold lead
     clause, then continuation and body concatenated with no break -- so
     there is no line break at the exact point the text used to be cut off,
     and no separate "chunk" sitting under the old prompt. Both blocks stay
     permanently in the DOM (the claims gate and crawlers must read the full
     text regardless of open state); only CSS decides which one paints. */
  .trp__card[open] .trp__text { display: none; }
  .trp__body-lead { font-weight: 700; color: var(--color-text); }

  /* Bounded height + internal scroll, ALWAYS, open or closed. This is the
     answer to "how does a dropdown behave inside a horizontally-scrolling
     carousel": if an open card could grow without limit, the whole track's
     row height would grow with it, dragging the overlaid arrows (positioned
     at a fixed offset from the top) out of alignment with every open/close.
     Capping it means the arrows never need to move and the track height never
     jumps. In practice every body written so far is a sentence or two and
     never reaches this cap -- it is a safety bound, not an active constraint.
     Horizontal padding matches .trp__text's exactly, so the paragraph that
     replaces the closed block sits the same distance from the card edge the
     hook did -- no inset mismatch between the two states. */
  .trp__body { padding: 0 var(--spacing-xl) var(--spacing-xl); margin-top: -.3rem;
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
            <div class="trp__body"><p><strong class="trp__body-lead">{{$item.hook}}</strong> {{$item.continuation}} {{$item.body}}</p></div>
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
$HTML$
 WHERE function = 'teaser-reveal-panel';

COMMIT;

SELECT function, is_active, length(html_template) AS template_bytes,
       html_template LIKE '%trp__body-lead%' AS has_merged_paragraph
  FROM content_components WHERE function = 'teaser-reveal-panel';
