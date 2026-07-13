# HANDOFF — Adoption interactivity routing fix deployed; recreation-loss defect still blocking

**Date:** 2026-05-26
**Site under investigation:** gamesdesign.co.uk (the adoption test bed)
**Companion docs:** `PLAN_tool_widget_clobber.md` (full investigation, all queries A–R3), `016_debugging_guide_addendum_adopted_tools_no_widget.md`

---

## 1. TL;DR

- **Deployed:** T1 (`apply_adoption_plan_action.go` — canonical-keyed `buildPageFeatureMap`) and T2 (`check_tool_recreation_needed.go` — new discovery check). Both are in production.
- **Confirmed:** the routing misroute that turned adopted tool pages into static description pages with no widget. It was a prefix desync between raw feature keys and the canonicalised lookup. Code-level fix verified.
- **Not confirmed:** that widgets now actually *deploy*. Query K (step 8) showed five games which had been routed correctly *all along* and whose recreation work items completed, but they had **no deployed widget**. So correct routing → completed recreation → no widget. **T1 alone may not be sufficient** if that downstream defect is real.
- **Not yet diagnosed:** *why* correctly-routed recreation didn't produce widgets. The state was reset under us before L/M/N could run (step 9), and the diagnosis must be re-pointed at whatever names currently exist.
- **Hold the trigger.** Do not mass-emit `needs_tool_recreation` yet — until R1–R3 + a re-pointed L confirm widgets actually land, bulk-triggering reproduces the games' empty result.

---

## 2. What is now in production

### T1 — `apply_adoption_plan_action.go` (one function changed)

**Function:** `buildPageFeatureMap` (≈ lines 718–793 in the merged file).

**Before:** keyed `featuresByPage` by the raw `fm["page"]` (e.g. `drop-rate-simulator`). The routing loop looks up `pageFeatures[pageName]` where `pageName` is canonicalised by `datahelpers.CanonicalisePage` — and the `tool` branch *adds* a `tool-` prefix. So tool lookups always missed (`tool-drop-rate-simulator` ≠ `drop-rate-simulator`); `adoptedPage.Features` came back empty; the page was routed `needs_content_page → page-build-handler` and rebuilt as a static description page with no widget. Games matched only by coincidence — their feature keys already carried `game-`, which the `game` branch preserves.

**After:** the map is keyed by the **canonical** name. The function resolves each feature's role from `plan["pages"]` and canonicalises via `CanonicalisePage`, so it lands in the same name space the routing loop looks up. Both `adoptedPage.Features` (line ≈ 551) and `contentData["interactive_features"]` (line ≈ 525) read `pageFeatures[pageName]`, so the single change fixes routing decision and content attachment together.

**Verified before deploy:** lines 1–717 and trailing helpers byte-identical to the source; only `buildPageFeatureMap` changed. One local rename flagged inline (`pageName` → `rawPage`). No import changes. `tool-recreation-handler` untouched.

### T2 — `check_tool_recreation_needed.go` (new discovery check)

A `discovery_checks` package check. Per site, it:
1. Finds pages where `page_type IN ('tool','game')`, `status='active'`, and no body component is a tool/game component *or* carries an inline `<script>`. The `<script>` arm makes detection robust to the actual `component_level` recreated widgets use.
2. Loads the latest adoption `findings`, builds a canonical-name → features map using the same `CanonicalisePage` transform as adoption (twin of `buildPageFeatureMap`, but read-only).
3. Emits `needs_tool_recreation` for widget-less pages where adoption captured `interactive_features`. Spec mirrors what `apply_adoption_plan` emits, so `tool-recreation-handler` receives identical input.
4. Pages with no captured features are surfaced as findings (visibility) but **not** auto-recreated — generation is the `tool-suggester` path, not this check.

**Item key:** `needs_tool_recreation:<page_name>` — deliberately distinct from adoption's `needs_page:<name>` to avoid colliding with completed content-page items for the same page.

**Cooldown:** 7-day per page (consistent with the other checks). Runner handles dedup via `idx_swi_dedup`.

**Consequence:** on its next scheduled discovery run, T2 picks up the existing widget-less interactive pages and emits recreation items automatically. That is the backfill mechanism — no separate emitter (deliberately, to keep one owner).

---

## 3. The blocking finding (must be resolved before the deploy is "done")

**Step 8 (queries J + K).** All five game pages were widget-less:

```
has_widget_component=f, has_script_section=f, for every game
```

Yet step 6 (query I) had already confirmed those same five games went `needs_tool_recreation → tool-recreation-handler` and the items were `complete`. **So correct routing + completed recreation did not yield a deployed widget for any game.**

If that's a real defect (not a detection artifact), then applying T1 to the *tools* will route them correctly but they will end up exactly where the games did — completed recreations, no widget. The visible bug would not change.

**Important caveat — K is suggestive, not proof.** `has_script_section=f` can be a false negative because:
- The snippets mechanism extracts inline `<script>` into `/assets/js/snippets.js`. A deployed widget can show no inline script in `rendered_html` while still being interactive.
- The recreated widget may sit at a `component_level` other than `tool`/`game`, defeating the component_level test.

Either could fully explain K without recreation actually being broken. Query **L** (page-component dump on a real interactive page) is decisive.

**Untested candidate mechanisms** (do not act on these without evidence):
- **a)** Recreation built a new `tool-game-*` page instead of populating the original `game-*` page (target/canonicalisation defect inside `tool-recreation-handler`). Step 8 showed five `tool-game-*` planned pages that fit this shape — but they vanished in the step-9 state reset, so origin is unknown.
- **b)** A widget did land but was clobbered by a subsequent `save_page_sections` rebuild (M1 — would leave a snapshot in `page_component_history` with `source='save_page_sections_overwrite'`).
- **c)** `tool-recreation-handler` completed without persisting any widget (handler-side defect — failed `check_tool_completeness` but completed the workflow anyway, or wrote to the wrong place).
- **d)** K is a false negative; widgets are present.

The five `tool-game-*` duplicate pages observed in step 8 are now unaccounted for and may or may not exist in current state.

---

## 4. State-of-play caveats

- **Parallel adoption chat is actively rewriting state.** Between step 8 and step 9, all five games, the five `tool-game-*` duplicates, and all five `needs_tool_recreation` items disappeared from the names we'd been querying. A `LEFT JOIN` cannot drop the page row, so these are genuine deletions/renames, not query artefacts. Almost certainly a re-adoption batch (or a reaper).
- **`pages.build_status` is drifting** between reads (e.g. `tool-drop-rate-simulator`: `needs_rebuild` → `deployed` between steps 3 and 8). Rebuilds are running.
- **Treat name-specific reads as time-sensitive.** Always re-baseline with R1 at the start of a new session before running any name-specific diagnostic.
- **Confirm the deployed code is current.** Since the adoption chat may have re-edited `apply_adoption_plan_action.go` after our merge was uploaded, re-verify in production that `buildPageFeatureMap` is the canonical-keyed version (look for the `canonByRawName` map and the `rawPage` local).

---

## 5. Next steps — exact order

Run in this order. Do not skip ahead.

### Phase A — Re-baseline (decides whether the rest of the work is even needed)

Run queries **R1, R2, R3** from `PLAN_tool_widget_clobber.md` §7.

- **R1** is the single most important read in this handoff. It lists every page on the site, its `page_type`, `build_status`, and a `has_widget` flag. **Read the tool/game rows first.** Possible outcomes:
  - All interactive pages show `has_widget=t` → the re-adoption ran with T1 deployed and the bug is gone for new adoption runs. Move to step 6 (verify), skip Phase B.
  - Interactive pages show `has_widget=f` → bug still present in current state; proceed to Phase B with current names.
  - Interactive pages renamed (e.g. games without the `game-` prefix) → adopt the new names for L/M/N; do not assume step-8 names.
- **R2** shows what work items exist now. Are there pending `needs_tool_recreation` items (T2 already fired)? `needs_content_page` items for tool pages (T1 not active on this run, or build path bypassed adoption)?
- **R3** confirms a fresh `adoption_crawl` row around the time of the reset and explains the wipe.

### Phase B — Diagnose recreation-loss (only if R1 shows widget-less interactive pages)

Re-point L/M/N from §7 at the current names from R1.

- **L (decisive):** dump `page_components` for one widget-less interactive page **and** any matching `tool-<name>` or `tool-game-<name>` page if it exists.
  - Description-shaped content on the original page (hero/generic-text-block/tool-list) → recreation didn't populate it.
  - Interactive markup on a separate `tool-*` twin → recreation mis-targeted to a new page (candidate a).
  - Interactive markup with no inline `<script>` on the original page → K was a snippets false-negative (candidate d); recreation is fine, T1 worked.
  - Nothing on either → handler never persisted (candidate c).
- **M:** if there are `tool-game-*` or `tool-<name>` duplicates, who created them (`handler_agent`, `created_by`, `source`)? Distinguishes recreation mis-target from a planner/reconciler divergence (029 family).
- **N1:** what page did each `needs_tool_recreation` target, and how did it terminate? If status=complete but target page is widget-less → handler-side defect or clobber.
- **N2:** clobber evidence — was a widget ever on the page then overwritten? A snapshot with `source='save_page_sections_overwrite'` is positive proof of M1.

### Phase C — Fix the right thing

The Phase B result decides what to fix next, not before:

- **Mis-target (a):** patch in `tool-recreation-handler`'s workflow / its recreation action where it resolves the target page. The handler should attach the widget to the original interactive page, not create a parallel `tool-*` page.
- **Clobber (b):** prioritise T4 (the M1 clobber fix). Make the tool widget a first-class section in `site_plan` (which `load_page_sections_from_spec` syncs into `pages.sections`), or guard `save_page_sections` from dropping `component_level='tool'`/`'game'` rows that the new section set doesn't include. See `PLAN_tool_widget_clobber.md` §5.
- **Handler never persisted (c):** inspect `tool-recreation-handler`'s `recreate_tool` → `check_tool_completeness` → `spawn_rerender` chain. Look at agent logs and the workflow's `output_field` plumbing. The check_tool_completeness pass/fail and the rerender's deploy result.
- **K was a false negative (d):** nothing to fix. Re-run K with a snippets-aware signal, or simply rely on R1's `has_widget` once T2 normalises `component_level`.

### Phase D — Canary, then backfill

Once Phase C resolves the recreation-loss:

1. Confirm `tool-recreation-handler` is `active` with the current chassis image: `kubectl -n ai-persona-system get pods | grep tool-recreation-handler`.
2. Pick **one** widget-less interactive page. Enqueue one `needs_tool_recreation` for it (either by triggering T2's discovery run, or by inserting a single work item matching T2's spec). Watch it complete.
3. Run L on that page. Confirm the widget actually deployed.
4. **Only then** let T2 backfill the rest (its scheduled run picks up the remaining pages), or trigger discovery for the site.
5. Re-run R1 + K-style checks: every interactive page should now show `has_widget=t`.

---

## 6. Open tasks (post-deploy)

| ID | Task | Status | Notes |
|---|---|---|---|
| T1 | Canonical-keyed `buildPageFeatureMap` | **Deployed** | Forward-looking: applies to new adoption runs. |
| T2 | `tool_recreation_needed` discovery check | **Deployed** | Backfills automatically on next discovery run, *if* recreation works. |
| T1b | Backfill emitter | **Resolved as T2** | No separate emitter — keeps one owner. |
| T4 | M1 clobber fix | **Pending diagnosis** | Phase B / N2 says whether it's active. |
| T5 | `tool-game-*` duplicate pages | **Pending re-observe** | May have been wiped by step-9 reset; R1/M will show. |
| T3 | Canonicalise `create_tool_component` + `deploy_tool` identity | **Open, independent** | The older surfaces were missed in `029` Phase 0. Same canonicalisation discipline as T1. Low risk; can land at any time. |
| Detection prompt review | Make sure adoption's `interactive_features` detection stays reliable | **Open** | b2 was ruled out for gamesdesign but the prompt is one place this whole chain could fail silently if it ever returns empty. |

---

## 7. Acceptance criteria

The deploy can be called complete when, for at least one site that goes through adoption end-to-end:

1. Every `page_type IN ('tool','game')` page on the site has `has_widget=t` in R1.
2. A deployed tool page in the browser actually renders an interactive widget (form/canvas/script), not a description.
3. T2 produces zero new `needs_tool_recreation` items on a subsequent discovery run (steady state — there's nothing left for it to find).
4. No `tool-<name>` or `tool-game-<name>` parallel pages exist for an adopted interactive page (the original page carries the widget, not a duplicate).

---

## 8. References

- `PLAN_tool_widget_clobber.md` — full investigation; queries A–R3; changelog steps 1–9.
- `016_debugging_guide_addendum_adopted_tools_no_widget.md` — assumption-checklist item, D/E/F diagnostic recipe, root cause.
- `apply_adoption_plan_action.go` (deployed) — T1; the new `buildPageFeatureMap` (≈ lines 718–793) is the fix.
- `check_tool_recreation_needed.go` (deployed) — T2; the discovery check + `loadAdoptionFeaturesByCanon` helper.
- `datahelpers/page_canonical.go` — the canonical join key (`CanonicalisePage`) both surfaces reuse.
- `029` — page-identity canonicalisation history; T3 lives here.

---

## 9. Risks and things to watch

- **Recreation-loss defect (§3) is the dominant risk.** Until Phase B/C resolves it, the user-visible symptom may not change after the deploy — bulk-triggering recreation now would just reproduce the games' completed-but-empty outcome and add noise.
- **Schema assumption in T2.** Detection treats `component_level IN ('tool','game')` as "has widget." If recreated widgets actually use a different level, the `<script>` arm still catches deployed widgets, but the field reported by R1's `has_widget` may underreport. Phase B / L confirms what level the recreated widget actually uses.
- **Adoption chat coordination.** State is moving; before any DB write or new deploy, confirm the latest `apply_adoption_plan_action.go` in production still contains T1's `buildPageFeatureMap` (look for `canonByRawName`), since the other chat may have re-edited the file.
- **Snippets extraction.** If `/assets/js/snippets.js` is broken or stale, even a correctly-deployed widget will not be interactive in the browser. Worth a separate cross-check once Phase D confirms widgets are landing.
- **`check_tool_health` blind spot persists.** Its INNER JOIN reports "no tools" when a tool page has no linked tool component. T2 partially covers this, but the existing check should be updated to flag widget-less interactive pages directly, not just sites with zero tools.
