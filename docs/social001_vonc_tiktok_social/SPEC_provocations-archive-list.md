# SPEC — provocations-archive-list (component)

**Opened:** 2026-07-05  •  **Status:** SPEC (component not yet created)
**Page:** provocations-index (e4b3b195-919f-45ad-854e-201d3e846ea8, /provocations/index.html,
"Provocations Archive | Spark") — this component is its single planned section (ordering 0).
**Related:** RUNBOOK_phase2_provocation_js "CURRENT STATUS" steps 1–4; NOTES_provocations-index.md;
PLAN_dynamic_sections_and_loaders.md (this build is the second hand-built reference for the
loader-builder agent); lobby_grid_loader.js (the pattern).

---

## Aim
The Provocations Archive — the destination of the site's primary CTAs. A self-contained section:
its own header (eyebrow, title with one emphasised word, subtitle) + a list of past provocations
as enterable cards, filled at runtime from the provocations feed. Dark, arena-adjacent, quieter
than the index (a record room, not the arena floor).

## Design decisions baked in at generation (each is a lesson from this thread)
1. **`data-runtime-fill="true"` emitted IN the template's `<section>` tag** — marker at generation,
   no post-hoc string surgery (guide §9 marker-anchoring entry).
2. **NO inline `<script>` anywhere in the template** — all behaviour lives in the external loader
   (js_snippet), so the inline-JS extraction-bug class cannot occur and the visible-content /
   assembly path stays trivial.
3. **Header copy = llm fields** (always resolvable → no plan_sections deferral risk). The list
   entries are pure markup (no schema fields) → nothing for the resolver to fail on.
4. **Clone-template list** (variable-length archive): the template contains ONE hidden item element
   the loader clones per entry — unlike lobby-grid's six fixed slots.
5. **Empty state**: a visible one-line element shown until the loader fills (and left showing if
   the feed is absent) — the page is shippable before the data lands.

## Component-creator input (084 trigger, 083 dual-placement pattern)
- `section_type`: `provocations-archive-list` (TOP-LEVEL for the contract AND inside spec).
- `site_type`: content  •  `page_context`: "provocations-index — the Provocations Archive page;
  the destination of the site's primary CTAs".
- `description` (pins the contract; function name MUST be provocations-archive-list):
  "A self-contained archive section for the Provocations Archive page. The `<section>` element MUST
  carry class `provocations-archive` and the attributes `data-component=\"provocations-archive-list\"`
  and `data-runtime-fill=\"true\"`. Header: a small eyebrow label; a heading with ONE emphasised word
  or phrase in an `<em>`; a one-or-two-sentence subtitle. Below the header, an archive list container
  `provocations-archive__list` containing EXACTLY ONE hidden item element that serves as a clone
  template: an `<a>` with class `provocations-archive__item` and the attribute `data-archive-template`
  and `hidden`, containing: a date element `provocations-archive__item-date`; a title
  `provocations-archive__item-title`; a one-line teaser `provocations-archive__item-teaser`; and a
  stat `provocations-archive__item-stat` with a small dot span `provocations-archive__item-dot`
  before the stat text span. After the list, an empty-state line with class
  `provocations-archive__empty` (visible by default). Then one call-to-action button linking back to
  today's provocation. Every copy field carrying the site voice (eyebrow, heading, subtitle, the
  empty-state line, the CTA label) must be a content-writer placeholder; the CTA url is a tunable
  with fallback `/index.html`. This is a DARK section using the site CSS variables only. Do NOT
  include any `<script>` element anywhere in this component — its list is populated at runtime by an
  external loader; the hidden template item and the empty state are the entire built-in behaviour."
- `design_direction`: "A record room adjacent to the arena — dark, composed, legible; generous
  vertical rhythm; cards as quiet rows that light on hover (CSS only); the featured energy belongs
  to the index, not here. Site CSS variables only; set the section --section-* variables for the
  dark surface."

## DOM contract (what the loader targets — must match the generated template; verify before
writing the loader)
```
[data-component="provocations-archive-list"]            (has data-runtime-fill)
  .provocations-archive__eyebrow                        textContent   (llm)
  .provocations-archive__title                          innerHTML     (llm, may carry <em>)
  .provocations-archive__subtitle                       textContent   (llm)
  .provocations-archive__list
    a.provocations-archive__item[data-archive-template][hidden]   ← clone per entry
      .provocations-archive__item-date                  textContent
      .provocations-archive__item-title                 textContent
      .provocations-archive__item-teaser                textContent
      .provocations-archive__item-stat  span (non-dot)  textContent  (dot span preserved)
    href on the clone                                    entry.url
  .provocations-archive__empty                          hidden once ≥1 entry rendered
  CTA button                                            label llm / url tunable fallback /index.html
```

## Data contract (provocations.json — extends the existing file)
```json
"archive": {
  "entries": [
    { "date": "3 Jul", "title": "…", "teaser": "…", "stat": "1,102 positions · 64% split",
      "url": "/provocations/index.html" }
  ]
}
```
Entries newest-first; loader renders up to 24, hides the empty state when ≥1 rendered, otherwise
leaves the shell + empty state as-is (graceful, same failure posture as the other loaders).
Interim: hand-extend provocations.sample.json; Phase-3 pipeline emits it later.

## Loader (Step 4 deliverable — mirror lobby_grid_loader.js + clone-template)
IIFE; bail if the section is absent; fetch `/data/provocations.json` (no-store); fill header
(eyebrow/title/subtitle; aria-label from title); find `[data-archive-template]`, for each
`archive.entries[i]` (cap 24) cloneNode(true) → remove `data-archive-template` + `hidden` → fill
date/title/teaser/stat(non-dot span)/href → append to `__list`; hide `__empty` if any rendered.
Installed as js_snippet `provocations-archive-loader`, `applies_to ["provocations-archive-list"]`
(dollar-quoted insert, then site-asset-renderer — App. H file destinations apply).

## Sequence (runbook CURRENT STATUS steps)
0. Confirms (pages.sections for e4b3b195; plan id; style-id pattern) →
1. 084 trigger creates the component; verify (has_marker=t, has_inline_script=f, schema>0) →
2. plan-section INSERT-SELECT (single row, ordering 0) →
3. needs_page build → 200 on /provocations/index.html (empty state visible) →
4. loader + `archive` data + asset renderer → cards render; CTAs across the site stop dead-ending.

## Guardrails
- If the generated template deviates from the DOM contract (class names differ), the DUMP is
  authoritative — adjust the loader to the real markup, as with lobby-grid; do not regenerate for
  cosmetic naming.
- Do not add plan rows or fire the build before Step 1 verifies (a plan row naming a nonexistent
  component re-enters unknown triage territory).
- One section only for v1; more (stats, CTA banner) can be added to the plan later.
