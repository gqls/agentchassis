# NOTES — teaser-reveal-panel

**Travelling doc, per `037_TOOL_DOCS_convention(1).md`.** Append-only, newest at the
bottom. Every entry: date, what was observed, what was decided, and why — including dead
ends, because those are what the next maintainer cannot rederive. Tagged with the
convention's problem categories where one fits.

Intent and the do-not-undo list: `PLAN_teaser-reveal-panel.md`.
Fuller lane narrative (this component alongside everything else that day):
`../../NOTES_brochure_component_library.md`, entries from 2026-07-29 onward.

---

## 2026-07-29 — built, as the §4c carousel

Built to implement the pre-existing `experience_patterns` row
`teaser-detail-deeplink`. Registered as CLC-012. Shipped first on `index.html`, then
extended the same day with optional `image_url` / `image_alt` and rolled out to
`capabilities.html` (replacing `services-grid`), `multi-agent-review-council.html`
(replacing `info-card-grid`) and `model-fine-tuning.html`.

On `model-fine-tuning.html` the change **merged two card-grid sections** that were
independently restating four of the same six facts. That duplication is not a
component defect — it is the site-wide repeated-capability-list problem filed as
`bugs_open/151` (the per-section content writer gets the whole site's fact pool with no
record of what a sibling section already said). Worth knowing here because it is *why*
this component exists in the shape it does: one row of distinct teasers instead of five
paragraphs of overlap.

All images used already existed (~25 generated assets); none were newly generated.

## 2026-07-29 (evening) — owner feedback round 1

Four things, all accepted and all still load-bearing:

- padding widened;
- **the visible ellipsis must be CSS-drawn, never a stored character** — a literal
  trailing ellipsis is what a truncation checker reads as a cut-off generation, so
  storing one would make honest content look like damage. `content: "\2026"` on
  `::after`. Category: `content-vs-runtime-mismatch` (inverted — the *rendering* carries
  what the data must not);
- the open state was redesigned to **replace** the whole closed block with one merged
  paragraph, rather than appending the body under a continuation still showing its own
  cut-off line. The owner rejected the appended version explicitly;
- real carousel arrows, reusing `hero-card-carousel`'s exact `goTo` / `nearestIndex`
  pattern rather than writing a second implementation.

## 2026-07-30 — the bug that had been live since the first version

**Category: `js-not-extracted` / head-load ordering. The most expensive entry in this
file.**

Owner reported the arrows did not scroll and opening one card did not close another.

**Root cause: `<script src="/assets/js/snippets.js">` sits in `<head>`, plain, no
`defer`.** It executes synchronously at that point in parsing — *before* this
component's markup exists in `<body>`. `behaviour.js`'s first line queried
`[data-component="teaser-reveal-panel"]`, found **zero** panels, and
`if (!panels.length) return;` exited. **The whole file silently no-opped on every page
load, from the very first version.** Only the native `<details>` element (which needs no
JS) and the CSS had ever done anything; the deep-link and sibling-close logic had never
run client-side.

**Why nothing caught it for a day, which is the transferable part:** every check this
component had passed — the render harness, `probe_reveal_open_state.py`, and several
"verified live" claims — exercised either the static markup or forced `.open = true`
directly on DOM nodes. **None of them ever called `.click()`.** No amount of that class
of checking could have found this.

Fixed by wrapping per-panel init in `initAll()` gated on `document.readyState`, the same
shape `hero-card-carousel`'s snippet already had. Checked fleet-wide: 6 of 7 active
`js_snippets` already followed that convention, so this was **not a platform gap** — this
component simply missed it (`news-date-formatter` is the one other exception, unchecked).

Re-tested with real clicks after the fix: `scrollByCalled: true`, `scrollLeft` 0 → 272,
and opening card 1 correctly set card 0's `.open` to `false`. Two symptoms, one cause.

## 2026-07-31 — the padding was ZERO, not small; and one baseline for the control

Owner reported text touching the card edges and the "Read the rest" links sitting at
different heights. Two requests, one root cause plus one layout choice.

**Category: a new one worth naming — `undefined-css-var-drops-the-declaration`.**

**The theme defines no `--spacing-*` scale.** Measured against the live stylesheet,
`--spacing-section` is the only one that exists. This component used `--spacing-xl`,
`--spacing-lg` and `--spacing-md` in **eight** declarations, none with a fallback. An
undefined `var()` with no fallback does not fall back to the property's initial value — it
makes the declaration *invalid at computed-value time* and the browser discards it
outright. Measured in Chromium:

    .trp__text   padding: 0px 0px 0px 0px    (declared: var(--spacing-xl) var(--spacing-lg))
    .trp         padding: 0px 0px
    .trp__inner  padding-left: 0px
    .trp__track  column-gap: normal          (0 — the cards were touching)

The hook sat 1px from the card border and **that 1px was the border**. The open-state
`.trp__body` was equally dead, so an opened card had no padding either — not yet reported,
because it needs a click to see.

**The misstep worth recording is in the file's own comment.** This style block opened by
stating that every variable had been *confirmed present in an active theme on
2026-07-29*. That was true of the colours and never checked for the spacing. **A partial
audit reads exactly like a complete one.** The comment now says which half was checked.
Fixed by naming the scale locally with literal fallbacks (PLAN decision 5). Computed
after: `21.6px 24px 21.6px 24px`. Fleet landmine filed with a one-command `comm -23` audit
diffing a template's var names against the theme's.

**The alignment fault:** `align-items: start` sized each card to its own content, so a
card whose continuation ran to one line instead of two finished **278px against 304px**
and its control sat **26px higher** (y=1891 vs 1917). Changed to `align-items: stretch`,
card as a flex column, `.trp__control { margin-top: auto }`. Measured after, all four
pages: equal heights, one control position per page, 23px from the card bottom, font
unchanged at 14.88px.

**A claim this component made and did not keep, now kept:** the style block asserts the
track height never jumps on open because the body is capped. With `align-items: start` it
*did* jump — 320 → 357 measured. With stretch it is 425 → 425. The documented invariant is
now true; it was aspirational before.

### The dead end, in full, because it nearly caused a false report

My verification harness (live page + candidate stylesheet appended + real clicks)
reported **two failures: sibling-close not closing and the next arrow not scrolling** —
the exact two symptoms of the 2026-07-30 bug. It read unmistakably as "your CSS broke the
JS".

**It had not.** The control run — identical probe, stylesheet *not* injected — reproduced
the same two failures plus four more that the change fixes. So the two were the harness,
not the change. Confirmed positively on the real page with no injection at all, using the
component's own deep link:

    ?open=vector-search   -> 1 open <details>, key="vector-search"
    ?open=review-council  -> 1 open <details>, key="review-council"
    (no param)            -> 0 open

**Cause of the artifact: a cross-origin `https://` `<script>` does not execute on a
`file://` page in this sandbox.** The tag is present, the track is genuinely scrollable
(scrollWidth 1732 > clientWidth 1152), and `scrollLeft` never moves — indistinguishable
from broken behaviour. Fetching the bundle and inlining it did not fix it either. This is
the same trap as a landmine filed hours earlier the same day about probes reporting the
bug they exist to catch. **Always run the control.** Recorded in PLAN's verification
contract as a standing requirement for this component.

## 2026-07-31 — this file, and why it is a file

These two documents did not exist until now. The component had taken five rounds of owner
feedback and carried three separate hard-won decisions, and **none of it travelled with
it** — the history was in the lane's log and its commits, which the next person to touch
this template may never open. The owner asked directly whether the changes were in the
component's travelling docs. They were not. This is the fix.

They are files rather than `doc_plans`/`doc_notes` rows because those tables still refuse
`subject_type='component'`. Verified 2026-07-31: `doc_plans_subject_type_check` lists
tool/pipeline/experience/action/experience-pattern only. Migration
`273_doc_subjects_component.sql` and its Go half (`c659e312b`) are written and would allow
it; the migration is unapplied by its author's explicit sequencing ("image first, then the
migration"). Two findings handed to that lane rather than acted on here: whether the
running image now carries the Go half is **[UNVERIFIED]** (the change added no distinctive
string, so a pod-grep cannot settle it), and **two different migrations share the number
273** in `sql_for_agents/`, which matters for a runner that takes every pending file.

**When 273 lands, port both documents to the DB and leave a pointer — do not maintain
two copies.**
