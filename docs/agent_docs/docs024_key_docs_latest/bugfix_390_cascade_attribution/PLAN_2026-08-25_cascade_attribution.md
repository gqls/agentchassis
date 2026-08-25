# PLAN — bugs_open/390: the repair is authored at a surface that cannot govern the pixel

**Opened 2026-08-25.** Lane for `bugs_open/390` (arm 2 of `bugs_closed/352`).

## 1. What the bug is, in plain terms

A `contrast_failure` work item records one unreadable text-on-background pairing on one page.
It is sent to `css-patch-agent`, which appends a CSS rule to the site's theme stylesheet and
then marks the item `complete`.

The item completes because a rule was **written**, not because the text became readable. When
the rule has no effect the text stays unreadable, the next audit measures the same pairing,
files it again, and another dead rule is appended.

## 2. Why the appended rule usually cannot win

Two independent reasons, both structural:

- **Source order.** The declaration that actually sets the colour normally lives in a per-page
  `<style>` block stored in `page_components.rendered_html`. `assemblePage`
  (`platform/orchestration/actions/rerender_single_page_action.go:560-720`, sections written at
  `:700`) emits those blocks inside `<main>` — always after the `<link>` to the theme in
  `<head>`. So the theme is always earlier, on every page, by construction.
- **Specificity.** The audit files an ancestor-derived selector (`.ancestorClass TAG`), and the
  agent's prompt tells it to repeat that selector verbatim. Against the page's own
  `.section-class .element-class` that is a lower specificity, so it loses before source order
  is even consulted.

## 3. Decisions, and their reasons

**D1 — fix the ROUTING before the RECORD (owner decision, 2026-08-25).** Two designs were on the
table: make the platform record where the winning declaration lives and repair at a surface that
can reach it; or stop `complete` meaning "a rule was written" and let only a fresh measurement
close the finding. The owner chose routing first. The gate is the right door-closer and is
deferred, not rejected.

**D2 — leave the historic rows alone (owner decision, 2026-08-25).** Arm 1 got migration 587,
which withdrew 73 unexecutable rows. Arm 2 will not get an equivalent. The next audit re-files
anything still broken, and with the repair aimed correctly those re-filings reach a handler that
can act. The damage figures stay quotable as evidence rather than being edited away.

**D3 — the cheap correction ships first, and it is the interim, not the destination.** The live
prompt states something false and instructs the losing move. Correcting it is one migration,
live on apply, no image roll. The measured requirement (D4) supersedes it.

**D4 — the attribution must be VERIFIED IN THE PAGE, not merely computed.** Walking
`document.styleSheets` cannot see everything (cascade layers, cross-origin sheets, duplicate
declarations). So the probe removes the declaration it believes wins, re-reads the computed
value, and restores it: if removal changed nothing, the attribution is wrong and is recorded as
unverified rather than acted on. This is arm 1's "prove it" invariant applied one level along —
the same reason its selector is asserted against the element it measured.

**D5 — the `item_key` does not change.** VIZ-016 records that the key embeds the selector and
that changing it already forced two alias-key retrofits in the retraction path. Every new fact
rides the spec.

## 4. Phasing

1. **Migration** — the prompt stops instructing the losing move (no image roll).
2. **Go, both images** — cascade attribution in the render-audit probe; routing and an override
   requirement in the filer.
3. **Migration** — the agent consumes the requirement; the one class no stylesheet can reach
   (an `!important` inline `style=` attribute) is parked before any LLM spend.

## 5. Corrections to the originating bug file

Recorded here rather than edited away, per the working-docs rules. Both are in
`NOTES_cascade_attribution.md` with their commands.

1. **390's worked case does not demonstrate the bug it is filed for.** `loancash.co.uk` has no
   linked style collection, so the agent parks there (`css_no_theme_198`) rather than completing.
   The mechanism is real; it is proven on `vonc.com`, not on loancash.
2. **A second way the repair goes inert — erasure — and it is designed behaviour.** Migration 543
   has webdesign-agent write the rendered CSS into the theme row byte-for-byte, so every
   theme-appended repair expires at that site's next design run. Filed separately.
