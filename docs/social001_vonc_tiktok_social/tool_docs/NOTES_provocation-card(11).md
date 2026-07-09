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

## 2026-06-29 — Extraction bug root-caused: truncated template + path that bypassed extraction
**Observed:** `has_script_close = 0` for provocation-card — the stored
`html_template` has NO `</script>`; its tail ends mid-function with a stray
backslash. The inline script is genuinely **truncated in the DB**.
**Root cause (two layers):**
1. provocation-card was stored via a path that did not run `separateInlineJS`
   (confirmed by the sibling lobby-grid, whose script is INTACT yet still raw in
   the template with empty js_content — a working extraction would have replaced
   it). Most likely predates separateInlineJS's addition to the store action;
   never regenerated since.
2. provocation-card's template was ADDITIONALLY truncated at generation time
   (token limit), uncaught because the store-path validation checks unclosed
   `<style>` but not unclosed `<script>`.
**Consequence:** `/tools/assets/provocation-card.js` is never produced; the
card-activation interactivity doesn't deploy; and Path-1 migration is blocked
because there is no clean inline script to extract.
**Fix (decided):** REGENERATE provocation-card through the current store path with
a COMPLETE inline `<script>` that contains BOTH the card-activation behaviour AND
the data fetch-fill loader. separateInlineJS then extracts it to js_content →
`/tools/assets/provocation-card.js` deploys on rerender. Retire the Path-2 snippet
afterwards. (Cannot just re-extract — the stored source is truncated.)
**Also (system hardening, separate from this component):** add a `<script>`
open/close balance check to store-path validation; make separateInlineJS warn on
an unterminated `<script>`.
`Categories:` js-not-extracted, mode-b-template, schema-template-drift

## 2026-06-29 — Fix path: regenerate via component-creator (in-place) + Gap-2 dependency
**Generation path confirmed:** `component-creator` (only caller of
`store_generated_component`). Regeneration is **UPDATE-in-place** keyed by
`function`: preserves `component_id` (no FK relink), snapshots the old row to
`component_versions`, reactivates, and auto-raises a `needs_rerender` work item per
affected site. The pre-store validation now **rejects `<no value>` templates**, so a
regeneration can't re-store the broken shape.
**Gap-2 dependency (blocks a clean regenerate for THIS component):** the
component-creator prompt has no runtime-feed tier — regenerating as-is would
classify the provocation text as Tier A (LLM voice) and bake a build-time
provocation into the template, losing the daily loader. Fix = add a runtime-feed
tier (Tier E, `source: "feed.{name}"`) that makes the LLM emit the stable-selector
shell + an inline `<script>` loader (based on our proven loader) fetching
`/data/provocations.json`. Then `separateInlineJS` extracts it → `/tools/assets/
provocation-card.js` (Path 1); retire the Path-2 snippet.
**Determinism to confirm:** regeneration keys on the LLM's `function` output — pin
it to `provocation-card` so the UPDATE lands in place (else a duplicate INSERTs).
`Categories:` js-not-extracted, content-vs-runtime-mismatch, schema-template-drift

## 2026-07-03 — Truncation repaired; but section DROPPED from the deploy
Template truncation (unclosed inline <script>, `});\`) was repaired via targeted REPLACE
(close forEach + IIFE + </script>); the index rebuild re-rendered provocation-card whole
(rendered_len 9994→10015, pc_script_closed=t) and saved it to page_components (pos 2,
deployed, @13:15).
BUT the deployed HTML (2026-07-03) does NOT contain provocation-card (nor lobby-grid) —
only hero, gauntlet-cta, brief-explanation, system-stats. So provocation-card is in
page_components but ABSENT from the live page. On 2026-07-02 it WAS in the deploy (as the
truncated shell). So the assembly/deploy (page-rerender deploy_page) is dropping the two
Mode-B / inline-<script> sections. This is the daily-provocation CENTERPIECE — investigate
the page-rerender assembly / rebuild deploy_result before anything else. HYPOTHESIS only
(interactive-section handling at assembly); do not assume.
Reminder: provocation-card is filled at RUNTIME by the Path-2 loader (snippets.js) — but
that only matters if the section's shell is IN the deployed HTML, which it currently isn't.
`Categories:` detool-on-rebuild, js-not-extracted (truncation fixed), assembly-drop (new)

---

## STATE as of 2026-07-03 (provenance snapshot)
**Working:**
- Template truncation repaired (inline `<script>` now closes: forEach + IIFE + `</script>`).
- Re-renders whole from the fixed template (rendered_len 9994→10015; script-closed check = t).
- `.pc-*` shell markup + selectors present in the rendered component; Path-2 loader
  (snippets.js, bundled by site-asset-renderer) is the runtime fill mechanism and is in
  the page `<head>`.
- Present in `page_components` for the index (position 2, build_status=deployed, @13:15).

**NOT working / open:**
- **Dropped from the deploy (PRIORITY).** provocation-card is in page_components but ABSENT
  from the deployed index.html (2026-07-03). It WAS in the 2026-07-02 deploy (as the
  truncated shell). This is the daily-provocation centerpiece missing from the live page.
  Investigation: PLAN_section_assembly_drop.md / RUNBOOK_section_assembly_drop.md.
- **Mode-B shell.** Empty input_schema, `<no value>` template — build-time content is empty
  by design (the loader fills it at runtime), but this may be what the assembler is
  dropping on. Durable target: Path-1 with a runtime-feed contract (Tier E), or keep the
  Path-2 loader if the assembler is fixed to keep shells.
- **Inline `<script>` not extracted (js_content=0)** — same extraction-bug class; the
  built-in card-hover script ships inline (now closed) rather than via `/tools/assets/...`.
- **Depends on Phase 3 data** — the loader needs `/data/provocations.json`
  (provocation-orchestrator + scheduled refresh) to fill real daily content.

`Categories:` js-not-extracted, empty-shell/mode-b-template, assembly-drop (new),
content-vs-runtime-mismatch

## 2026-07-03 (later) — Deploy-drop CAUSE FOUND: visible-content filter (not the extraction bug)
rerender_single_page_action.go getPageSections drops any section with <=10 chars of visible
text after stripping <style>/<script>/tags/entities/whitespace (sectionHasVisibleContent).
provocation-card is a Mode-B shell with empty build-time fields → <10 chars → dropped. The
inline <script> was a red herring (stripped before measuring). NOT the extraction bug.
provocation-card IS legitimately runtime-filled (Path-2 loader), so the drop is WRONG.
FIX (chosen): assembler exemption for `data-runtime-fill` (PATCH_section_visible_content.go)
+ add the marker to provocation-card's template + current rendered_html. Deploy code → mark →
needs_rerender. After: provocation-card ships as an empty shell that the loader fills live.
STATE update: still Mode-B + inline-script-not-extracted (separate), but once the marker fix
ships it will render on the live page again.
`Categories:` detool-on-rebuild (root-caused: visible-content filter), content-vs-runtime-mismatch

## 2026-07-03 (end of day) — RESTORED to the live page
Assembler patch (data-runtime-fill exemption in sectionHasVisibleContent) deployed + marker
added (template + index rendered_html) + index reassembled. curl confirms provocation-card is
back in the deployed index (5 sections; lobby-grid still absent, correct). The section-drop is
resolved structurally: runtime-filled shells are now first-class in the assembler.
WORKING now: shell ships to the live page (present in deployed HTML), template + instance carry
data-runtime-fill (future rebuilds keep it), inline hover <script> closed.
STILL OPEN: (a) confirm it visibly FILLS in a browser (loader is client-side; needs
/data/provocations.json — Phase 3); (b) inline <script> still not extracted to js_content
(separate extraction-bug class, cosmetic hover only); (c) durable Path-1 + runtime-feed contract
remains the longer-term target.
`Categories:` detool-on-rebuild (RESOLVED via assembler exemption), content-vs-runtime-mismatch (open: Phase-3 data)

## 2026-07-04 — Primary CTA destination is a 404 (provocations-index never built)
The card's primary CTA (and the hero/gauntlet CTAs via site_specs.cta.primary_url) point
at /provocations/index.html — planned 2026-06-22 with zero plan sections; seven
build/rerender items completed as silent no-ops; nothing ever deployed. See
NOTES_provocations-index.md + guide §9. The card fills correctly; its destination is the gap.
`Categories:` planning-gap (cross-ref)


## 2026-07-09 — Mini-lobby trim is the NEXT TASK (plan drafted); card otherwise stable
State: the card is live on the index and fills correctly at runtime (eyebrow, headline with `<em>`,
body, two CTAs, three stats) from `provocations.json` `today`. Its four-card mini-lobby STILL renders
and now duplicates lobby-grid's six-card arena, live on the index — the last redundancy there.
Trim plan drafted in PLAN_provocation-card ("Trim plan") and HANDOFF_2026-07-09 §4: template block
removal + loader `data.lobby` removal + the live index instance (rerender is assemble-only), with the
`pc-container` two-column media-query rule and the now-dead inline hover script handled in the same
edit. Method: dump → edit offline → UPDATE full text → verify by length delta; dated backups first.
Standing note: this card's primary CTA points at /provocations/index.html, which is now LIVE (the
archive shipped 2026-07-08) — the CTA no longer dead-ends. CTA-graph rework remains parked (Option B).
`Categories:` overlap/role-collision (being resolved), content-vs-runtime-mismatch, cta-graph (parked)


## 2026-07-09 (later) — Method corrected before any edit: HTML patching is the rejected mechanism
Reading 003/002 while priming the bundle: `content_data` is the source of truth; patched
`rendered_html` is lost on the next re-render ("HTML patching was rejected as an edit mechanism").
`fix_component_template_action` ALREADY implements `remove_element` (exactly this trim) but defers
page_components content changes to the section-editor workflow. Two re-render paths exist: light
(`rerender_page_sections`, a `page_rerender` item, from stored `content_data` via
`RenderComponentAction`, no LLM — escalates to a full rebuild when `content_data` is NULL) and full
(`needs_page`). `rerender_single_page` — what our index trigger runs — only ASSEMBLES pre-rendered
components, so a template-only edit never reaches the page. Since provocation-card is Mode-B, its
`content_data` may be NULL: that probe decides the sequence. `bundle_minilobby_trim.sh` is primed to
settle all of it before any edit. Incidental: `rerender_single_page` also injects `cc.js_content` per
component at assembly — the mechanism the inline-script extraction backlog would use.
`Categories:` method-correction, reuse, source-of-truth
