# NOTES — provocations-archive-list

Append-only provenance record for the `provocations-archive-list` component.
Intent lives in `SPEC_provocations-archive-list.md` (this component's PLAN).
`Categories:` uses the shared taxonomy (TOOL_DOCS_convention.md).

**component id:** 70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37  •  **function:** `provocations-archive-list`
**site:** vonc.com (`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`) — vonc-only, though `forked_from IS NULL`
**page:** provocations-index (`e4b3b195-919f-45ad-854e-201d3e846ea8`), its single section, ordering 0
**loader:** `provocations-archive-loader` (js_snippet, `applies_to ["provocations-archive-list"]`)
**feed:** `/data/provocations.json` → `archive.entries[]` `{date, title, teaser, stat, url}`

---

## 2026-07-06 — Created; both generation-time guards held
Built by component-creator via the 084 trigger (083 dual-placement pattern: `section_type` both
top-level and inside `spec`; description quote-free to survive the kcat/JSON pipeline). Result:
active, quality 100, 8 schema fields (7 `llm` + `cta_url` static with fallback `/index.html`),
`has_marker = t`, `has_inline_script = f`.

This was the **first live validation of baking the guards into generation** rather than repairing
afterwards: the `data-runtime-fill` marker was emitted in the section tag by the generator, so no
post-hoc marker SQL was needed (no "Part B" in the install), and the prohibition on `script` elements
means the inline-script truncation/extraction bug class cannot occur here. Header copy is `llm`, so
nothing in this component can be deferred by `plan_sections`.

DOM as generated: header (`__eyebrow` / `__title` with an `<em>` / `__subtitle`), a `__list`
containing exactly one hidden `a.provocations-archive__item[data-archive-template]` clone template
(`__item-date` / `__item-title` / `__item-teaser` / `__item-stat` with a dot span + an explicitly
classed `__item-stat-value`), a visible `__empty` line, and a CTA.
`Categories:` reference-build, generation-time-guards

## 2026-07-08 — Loader installed; archive fills; ghost-row defect found and fixed
Loader inserted as a js_snippet (`js_len` 3281 stored vs 3287 source — the familiar ~6-byte paste
drift; the bundle and the browser are the ground truth), bundled by site-asset-renderer (bundle header
reads `3 active snippet(s)`), and `archive` added to the committed `provocations.json`. The list fills:
eight rows (5 Jul → 28 Jun) with date, title, teaser and a pulsing-dot stat; the empty state hides once
anything renders; the loader fails gracefully and touches only the list, since the header and CTA are
build-time content written by the content-writer ("THE RECORD" / "Every *provocation* on record." /
"The positions are permanent. The splits don't lie.").

**Defect:** the hidden clone template rendered as an empty row with a lone dot. `hidden` maps to a
UA-stylesheet `display: none`, which loses to the component's own author rule
`.provocations-archive__item { display: grid; }`. Fixed by adding
`.provocations-archive__item[data-archive-template] { display: none; }` to the template **and** the
live instance, anchored on the base-rule opener; `rendered_len` 7455 → 7671 (+216 — the rule landed in
both the base selector and its mobile media-query copy, which is correct). Redeployed via the page's
own rerender trigger (085).
`Categories:` css-specificity (resolved), content-vs-runtime-mismatch (resolved via runtime-fill)

## 2026-07-09 — Live-confirmed; component closed
`curl … | grep -c 'data-archive-template] { display: none;'` returns 2 (was 0). Component is done and
stable. Open, but not owned here: the CTA's `cta_url` resolves to its static fallback `/index.html`
because no arena page exists (decision: leave until a real take-filing arena does — Option B); and the
`archive` feed is hand-committed until the Phase-3 pipeline emits `provocations.json`.

**Carry into the next dynamic component:** name the stat text span explicitly (this template's
`__item-stat-value` is cleaner than hunting for "the span without the dot class", as lobby-grid's
loader must); and any clone-template spec must require `[data-…-template] { display: none; }` in the
CSS, because `hidden` alone is not a reliable hiding mechanism once the item class carries a `display`.
`Categories:` reference-build, cta-graph (parked)
