# NOTES — provocation-card

Append-only history for the `provocation-card` component. Newest at the bottom.
Each entry: date · what was observed · decision/action · why. `Categories:` line
uses the shared taxonomy (see TOOL_DOCS_convention.md) so patterns roll up.

---

## 2026-06-24 — Discovered as an empty shell
**Observed:** provocation-card renders as blank card outlines on the index. Its
`input_schema` is `{}`, 0 template slots, and the stored `html_template` is full
of bare `<no value>` (field names lost). No `js_snippet` applied to it. So it had
**no content mechanism at all** — not content-writer-filled (no schema), not
JS-filled (no snippet).
**Why it matters:** a page rebuild cannot fix this — there is nothing to fill.
`Categories:` empty-shell, mode-b-template

## 2026-06-25 — Decision: runtime JS loader (Option 2), not build-time content
**Decision:** fill the card client-side from a daily JSON file, because the
content is **daily-regenerated** (Spark v1 product hypothesis). Build-time content
would freeze the provocation permanently. Confirmed on-doctrine (doc 022 Tier 1).
**Why:** the card is dynamic; static-content regeneration is the wrong tool here.
`Categories:` content-vs-runtime-mismatch

## 2026-06-25 — Loader written and inserted (Path 2)
**Action:** wrote `provocation-card-loader` (IIFE: fetch `/data/provocations.json`,
fill the DOM contract, fail gracefully, preserve the CTA SVGs). Inserted as a
`js_snippet` with `applies_to=["provocation-card"]`, is_active. Validated JS
syntax. Chose Path 2 (js_snippet) as the first route.
`Categories:` (none — normal change)

## 2026-06-29 — js bundle was stale; site-asset-renderer not auto-triggered
**Observed:** after inserting the snippet, the live site still didn't load it.
**Root cause:** `/assets/js/snippets.js` is only rendered by `site-asset-renderer`,
which is invoked at initial design (webdesign-agent) or full site rerender — NOT
when a `js_snippet` is added later. The normal page rerenders we ran touch the
per-component JS path (`/tools/assets/*.js`), never the snippet bundle.
**Action:** triggered `site-asset-renderer` manually for vonc via the generic
entry point. Ran clean (commit eb7f2ac); bundle deployed with the loader.
**Follow-up:** the durable fix (auto-trigger on snippet change) is FX-6 — but see
the Path-1 decision below, which makes it moot for this component.
`Categories:` js-bundle-stale

## 2026-06-29 — Loader proven working end-to-end
**Observed:** the provocation-card now fills on vonc.com/index.html — headline
(with em), AI take, stats, lobby icons, CTAs all render. Option-2 mechanism
validated.
`Categories:` (resolved)

## 2026-06-29 — Lobby titles/descriptions blank — data, not code
**Observed:** lobby card icons rendered but titles/descriptions were blank ("...").
**Root cause:** the committed `provocations.json` had stub `"title":"..."` /
`"desc":"..."`. The loader sets all three in one loop (icons proved the loop ran),
so it was a DATA gap, not a loader/CSS bug.
**Action:** re-committed the full JSON; all four lobby cards now populate.
**Lesson:** confirm test fixtures are complete before suspecting code.
`Categories:` (data fixture, not a code defect)

## 2026-06-29 — Path decision: Path 1 is the durable home (PD-3)
**Observed (PD-1):** `latest-news` ends with `<script src="/tools/assets/latest-news.js">`
and keeps its fetch in the component's extracted JS — i.e. **Path 1**. The
`news-date-formatter` snippet is only a helper.
**Observed (PD-2):** provocation-card `has_inline_script=TRUE` but `js_content`
is **empty** — its inline (card-activation) script was never extracted, so
`/tools/assets/provocation-card.js` isn't produced (and the card-hover behaviour
isn't deploying either).
**Decision (PD-3):** the architecturally-consistent home for the fetch-and-fill
loader is the component's own `js_content` (Path 1, auto-deploys on rerender),
matching news. Keep the Path-2 snippet as the working interim until the
extraction bug is fixed, then migrate and retire the snippet.
**Why:** consistency with the established pattern; removes the
site-asset-renderer/Gap-3 dependency for this component.
`Categories:` js-not-extracted

## 2026-06-29 — OPEN: extraction bug blocks Path-1 migration
**Open item:** investigate why provocation-card's inline `<script>` is not
extracted to `js_content` (separateInlineJS). The earlier template dump showed the
inline script truncated mid-function with a stray backslash — a possibly-malformed/
unterminated stored script may be why extraction skips it. This blocks both the
Path-1 loader migration AND the card-hover interactivity. Next in the work order.
`Categories:` js-not-extracted, mode-b-template
