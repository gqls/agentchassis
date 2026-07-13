# PLAN — lobby-grid (build)

**function:** `lobby-grid`  •  **component id:** 9304f14d-e19b-4ce1-b3fd-f6a315aec6ed
**site:** vonc.com (Spark) — index, plan position 5
**Opened:** 2026-07-03  •  **Status:** DESIGN (inspect + confirm decisions before building)
**Related:** NOTES_lobby-grid.md, NOTES_provocation-card.md, PLAN_provocation-card.md,
RUNBOOK_section_assembly_drop.md, provocation_card_loader.js (the pattern to reuse).

---

## Aim
The Arena "lobby" — a grid of enterable cards representing today's live provocations/rooms.
Design (from the shell): 6-card grid — 1 `--featured` (spans 2 cols), 4 standard, 1 `--wide`
(full row); each card = icon, tag, title, description, a stat + a pulsing stat-dot; section
header (eyebrow, title with `<em>`, subtitle) + a CTA. Dark, arena-like.

## Current state
Mode-B empty shell (empty input_schema, `<no value>` template), inline hover `<script>`
(js_content=0, not extracted), NO fill mechanism. Correctly filtered out of the deploy by the
assembler (sectionHasVisibleContent) because its build-time text is empty — same reason
provocation-card was dropped before the marker fix.

## Approach (recommended): runtime-fill loader (Path 2)
Mirror provocation-card exactly — the proven, now-unblocked pattern:
1. A client-side loader (js_snippet `lobby-grid-loader`, applies_to ["lobby-grid"]) fetches the
   daily data and fills the grid's cards + header in the browser.
2. Add the `data-runtime-fill` marker to lobby-grid's `<section>` (template + index instance) so
   the patched assembler keeps the shell in the deployed page.
3. The loader populates the empty `.lobby-grid-section__card-*` fields at runtime; if data is
   missing it fails gracefully (leaves the shell / a neutral state), matching provocation-card.
Why not build-time static content: the cards are today's provocations (daily-changing) — static
build-time content would be stale. Runtime fill keeps it current from provocations data.

## Decisions to CONFIRM with the user (before building)
- **D-A (overlap).** provocation-card currently renders a 4-card mini-lobby (`.pc-card` × 4,
  filled from provocations.json `lobby[4]`). lobby-grid is a separate 6-card grid → overlap.
  RECOMMEND: lobby-grid becomes the primary "today's provocations grid"; TRIM provocation-card's
  mini-lobby so provocation-card focuses on the single headline provocation + stats + CTAs.
  Clearer separation, less redundancy. Alternative: keep both (redundant).
- **D-B (data source).** lobby-grid needs ~6 cards. RECOMMEND: define an `arena` array in
  provocations.json — `arena: [6× {icon, tag, title, desc, stat, url}]` — emitted by the Phase-3
  provocation pipeline; interim = extend provocations.sample.json. Alternative: reuse the
  existing `lobby[4]` array (only 4, and overlaps provocation-card's mini-lobby).

## Build steps (after decisions confirmed)
1. **Inspect** lobby-grid's html_template DOM: exact card block markup + the header/CTA
   selectors, confirm Mode-B (empty fields, selectors intact), confirm `data-component="lobby-grid"`.
   [structure-check SQL + full-template dump — issued]
2. **Build** `lobby_grid_loader.js` — IIFE; fetch `/data/provocations.json`; read `arena` (or
   chosen source); for each of the grid's card slots set icon/tag/title/desc/stat + the header
   (eyebrow/title/subtitle) + CTA; preserve structure + pulsing dot; fail gracefully; overwrite
   (not append). Reuse the provocation_card_loader structure/error-handling.
3. **Install** the loader as a js_snippet (applies_to ["lobby-grid"]) → trigger site-asset-renderer
   to bundle into /assets/js/snippets.js (Path-2 mechanism, as provocation-card).
4. **Mark** lobby-grid `data-runtime-fill="true"` on the `<section>` (template + current index
   rendered_html) so the assembler keeps it (guarded NOT LIKE '%data-runtime-fill%').
5. **Data** — extend provocations.sample.json with the `arena` array (interim) + fold into the
   Phase-3 pipeline so the live JSON carries it.
6. **(If D-A trim)** edit provocation-card template to drop the `.pc-card` mini-lobby + update
   provocation_card_loader to stop filling `lobby[]`.
7. **Rerender** the index (rerender-index-vonc.sh) → lobby-grid ships + fills; verify via curl
   (data-component="lobby-grid" present) + browser (cards fill from the loader).

## Dependencies
- provocations.json carrying the arena data (Phase 3 pipeline; interim sample).
- The assembler exemption (DONE) + site-asset-renderer (bundles snippets).

## Guardrails
- Reuse the provocation_card_loader pattern + the data-runtime-fill mechanism (don't invent new).
- Runtime fill only — no baked stale daily content.
- Confirm D-A before trimming provocation-card (it's live + working).
- Keep the loader's selectors matched EXACTLY to the template (get the template first).
