# PLAN — teaser-reveal-panel

**Travelling doc, per `037_TOOL_DOCS_convention(1).md`.** Two docs per complex component,
keyed by the component `function`: this PLAN (intent; changes rarely) and
`NOTES_teaser-reveal-panel.md` (history; append-only).

> **Why these are FILES and not `doc_plans` / `doc_notes` rows.** The platform's
> travelling-doc tables refuse `subject_type='component'`. Migration
> `273_doc_subjects_component.sql` and its Go half (`c659e312b`) exist and would allow it;
> **the migration is not applied** (verified 2026-07-31: `doc_plans_subject_type_check`
> still lists only tool/pipeline/experience/action/experience-pattern). Convention 037 is
> a FILE convention and predates those tables, so it is available today and this is it.
> **When 273 lands, port these two files to the DB keyed by
> `subject_type='component', subject_key='teaser-reveal-panel'` and leave a pointer here** —
> do not maintain both. See NOTES, entry 2026-07-31, for what is blocking that.

## Aim

Present several short claims as a single row of cards a visitor can scan, where each card
shows a hook and the first clause of its explanation and can be opened in place to read the
whole thing. It replaces the pattern where a page restates the same handful of company
facts in section after section: one row of teasers, each opening to its own detail, instead
of five paragraphs saying overlapping things.

It implements the pre-existing `experience_patterns` row **`teaser-detail-deeplink`** (that
pattern has its own travelling doc under `subject_type='experience-pattern'` — this
component is one implementation of it, and the two documents are not interchangeable: the
pattern's doc describes the interaction, this one describes this build of it).

Concept register: **CLC-012**.

## Source spec

Owner request, 2026-07-29: a carousel for §4c of fundamentallyai.com, then extended the
same day to "carousels on almost every block, with images". Live on four pages:
`index.html`, `capabilities.html`, `multi-agent-review-council.html`,
`model-fine-tuning.html`.

## Behaviour contract

**Closed card:** image (optional), hook (bold), continuation (muted, with a CSS-drawn
ellipsis), and a control reading "Read the rest" with a down arrow.

**Open card:** the entire closed text block is **replaced** — not appended to — by one
flowing paragraph: the hook as a bold lead clause, then continuation and body concatenated
with no break. The image stays. There is deliberately no line break at the point the text
used to cut off, and no separate "chunk" sitting under a still-visible prompt.

**Row behaviour:** always one row, never two, at every width — `grid-auto-flow: column` at
all breakpoints; only the column width changes. Horizontal scroll with scroll-snap;
overlaid prev/next arrows on desktop, hidden under 40rem where native swipe covers it.

**Deep link:** `?open=<key>` opens the matching card (`data-trp-param="open"`,
`data-trp-key` per card). This is also the cheapest proof the component's JS is running at
all — see NOTES 2026-07-31.

**Opening one card closes its siblings.** Both text blocks stay permanently in the DOM
regardless of open state; only CSS decides which paints, so the claims gate and crawlers
always read the full text.

## Delivery mechanism

**Path 2 — library `js_snippet` bundled into `/assets/js/snippets.js`**, not an inline
`<script>`. Stated explicitly because it is the component's single biggest hazard: that
bundle is loaded from `<head>`, **not deferred**, so it executes before this component's
markup exists. The snippet therefore MUST gate its init on `document.readyState` and wait
for `DOMContentLoaded`. It did not, for the whole of its first day, and the entire file
silently no-opped on every page load (NOTES 2026-07-30).

Template lives in `content_components.html_template` (function `teaser-reveal-panel`, one
active row, currently used by fundamentallyai.com only — **check
`SELECT count(*) FROM page_components ... WHERE cc.function='teaser-reveal-panel'` before
editing; the row is fleet-visible even when one site uses it**). Behaviour lives in
`js_snippets`; changing it requires rebundling
(`scripts/rebundle_js_snippets_direct.sh`).

## Dependencies

- `content_components` row `teaser-reveal-panel`; `js_snippets` entry of the same name.
- `experience_patterns` row `teaser-detail-deeplink`.
- Theme custom properties. **Colours only** — see deliberate decision 5.
- Generated imagery under `/assets/images/`, referenced per item as `image_url`.
- Placement: `site_plan_sections` + `pages.sections` + `page_components.slot_name`. All
  three, or a re-render silently drops the section.

## Deliberate decisions — do NOT "fix" these

1. **The ellipsis on the continuation is drawn by CSS (`content: "\2026"`), never stored
   as a character.** A literal trailing ellipsis in stored content is what this platform's
   own truncation checks read as a cut-off generation, i.e. as damage. If you "tidy" this
   into the data you re-introduce that false signal. The visible dots exist only in the
   rendered pixel.

2. **Opening replaces the closed block entirely.** An earlier version appended the body
   under a continuation that still showed its own cut-off line; the owner rejected it. The
   merged paragraph is the requirement, not an implementation detail.

3. **The body is height-capped with internal scroll, open or closed.** If an open card
   could grow without limit the whole track's row height would grow, dragging the
   fixed-offset overlaid arrows out of alignment on every open and close. The cap is a
   safety bound, not an active constraint — no body written so far reaches it.

4. **Cards stretch to the tallest in the row (`align-items: stretch`), and the control is
   pinned to the card's bottom with `margin-top: auto`.** This is what keeps every "Read
   the rest" on one baseline and the coloured panels a uniform height when one card's
   summary wraps to fewer lines than its neighbours'. Reverting to `align-items: start`
   reintroduces both faults, and also breaks decision 3's promise: with `start`, the row
   height DID jump on open (measured 320 → 357).

5. **Spacing is a component-local scale with literal fallbacks
   (`--trp-card-x: var(--card-pad, 1.5rem)`), NOT the theme's.** The theme defines no
   `--spacing-*` family — `--spacing-section` is the only one that exists fleet-wide. An
   undefined `var()` with no fallback does not degrade to the initial value; it invalidates
   the declaration and the browser discards it. Eight declarations here computed to **zero**
   for that reason. Do not "simplify" these back to bare theme names. Fleet-wide landmine
   filed under footprint `content_components.html_template` / `--spacing-*`.

6. **Two text blocks stay in the DOM at all times.** Do not switch to
   building/destroying the open content in JS: the claims gate and crawlers must read the
   full text irrespective of open state.

## Verification contract

**Interaction, in a real browser, with real clicks — nothing less counts for this
component.** Every check it passed on its first day (render harness, contrast probe,
"verified live") exercised static markup or forced `.open = true` directly, and a bug that
made the JS never run survived all of them for a day.

- `scripts/probe_reveal_open_state.py <url>` — forces open state, measures contrast.
  Does NOT click; insufficient alone.
- The deep link is the cheap, injection-free proof the JS runs:
  `chromium --headless=new --dump-dom "…/model-fine-tuning.html?open=vector-search"` must
  yield exactly one `<details … open>` carrying that key, and zero with no parameter.
- **Layout claims must be read as COMPUTED values, never from the declaration.**
  `getComputedStyle(el).padding` returning `0px` on an element whose CSS plainly sets
  padding is decision 5's bug.
- **Any interaction probe of this component must be run as a before/after pair.** A
  cross-origin `https` script does not execute on a `file://` copy in this sandbox, so a
  harness reports the arrows and sibling-close as broken when they are fine — the exact
  signature of the real bug. Run the control; a failure you cannot reproduce on unchanged
  code is a harness fault.
