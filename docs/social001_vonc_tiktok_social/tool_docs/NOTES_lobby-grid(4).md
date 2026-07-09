# NOTES — lobby-grid

Append-only provenance record for the `lobby-grid` component. Dated states of what works
and what doesn't. Newest at the bottom. `Categories:` uses the shared taxonomy
(TOOL_DOCS_convention.md). Read alongside the `lobby-grid` site_spec description.

**function:** `lobby-grid`
**component id:** 9304f14d-e19b-4ce1-b3fd-f6a315aec6ed
**site(s):** vonc.com (Spark) — index page, position 5
**aim:** the Arena "lobby" — a grid of enterable provocation/room cards. Design: 6-card
grid (1 `--featured` spanning 2 cols, 4 standard, 1 `--wide` full row); each card has an
icon, tag, title, description, a stat + a pulsing stat-dot; section header (eyebrow, title
with `<em>`, subtitle) + a CTA. Intended to be dynamic (today's provocations as enterable
cards) — a runtime-loader candidate like provocation-card.

---

## STATE as of 2026-07-03
**Working:**
- Component exists and renders its template (CSS + shell markup); rendered_html ~15,282
  bytes; its inline `<script>` (hover-dim / focus / IntersectionObserver entrance) is
  intact/closed (pc_script_closed-style check = t), i.e. NOT truncated like provocation-card.
- It is present in `page_components` for the index (position 5, build_status=deployed,
  re-rendered @2026-07-03 13:15).

**NOT working / open:**
- **Dropped from the deploy.** Despite being a deployed page_components row, lobby-grid is
  ABSENT from the deployed index.html (only hero/gauntlet-cta/brief-explanation/system-stats
  ship). Under active investigation — see PLAN_section_assembly_drop.md /
  RUNBOOK_section_assembly_drop.md. Hypothesis (unconfirmed): the assemble/deploy step
  drops Mode-B / `<no value>` / interactive sections.
- **Mode-B shell.** Empty input_schema, `<no value>` in the template — renders as an empty
  shell with no real card content. Needs proper build (real schema, or a runtime-feed
  contract) to show actual room/provocation cards.
- **Inline `<script>` not extracted (js_content = 0).** Same extraction-bug class as the
  other interactive components (has inline `<script>` but never separated to js_content →
  no `/tools/assets/lobby-grid.js`). Built-in interactivity not shipping via Path 1.
- **Overlap decision pending.** lobby-grid's card grid overlaps provocation-card's
  mini-lobby (the 4 `.pc-card`s). Open decision: make lobby-grid the "today's provocations
  grid" and trim provocation-card's mini-lobby (recommended), and settle v1 semantics
  (today's provocations as enterable cards; a richer ~6-card list in Phase 3).

`Categories:` empty-shell/mode-b-template, js-not-extracted, assembly-drop (new),
content-vs-runtime-mismatch

## 2026-07-03 (later) — Deploy-drop is CORRECT for lobby-grid (genuinely empty)
Same filter (sectionHasVisibleContent) drops lobby-grid — but lobby-grid is a Mode-B shell
with NO loader and NO content (its inline script is hover-only, fills nothing). It would ship
blank, so the drop is arguably CORRECT. lobby-grid is deliberately NOT given the
`data-runtime-fill` marker. It stays absent from the deploy until it's properly built:
either real static content, or (preferred, matching the "today's provocations grid" intent) a
runtime-fill loader like provocation-card's — at which point it gets the marker too. Tracked
as its own build task, not part of the section-drop fix.
`Categories:` empty-shell/mode-b-template, assembly-drop (correct-for-now)


## 2026-07-04 — Loader built; install pack ready (runtime-fill build underway)
Decisions confirmed: lobby-grid = the primary "today's provocations grid" (arena);
provocation-card keeps headline+stats+CTAs, its mini-lobby to be trimmed. Feed = `arena`
OBJECT in provocations.json {eyebrow,title,subtitle,cta_label,cta,cards[≤6]}.
BUILT: lobby_grid_loader.js (Path-2, mirrors provocation_card_loader; DOM contract in the
file header; icon=svg-inner-markup w/ emoji fallback; stat span selected by NOT-dot-class;
CTA trailing-text-node; enterable cards when url present; graceful failure).
READY: lobby_grid_install.sql (snippet insert + data-runtime-fill marker, guarded) +
provocations.sample.json v2 (arena, 6 cards). Run order: insert → site-asset-renderer →
marker → commit JSON → rerender index → verify.
This hand-build doubles as the REFERENCE for the future loader-builder agent
(PLAN_dynamic_sections_and_loaders.md gap 3).
STILL OPEN after this ships: inline hover <script> not extracted to js_content (extraction
class, cosmetic); Phase-3 pipeline to emit arena in the live JSON.
`Categories:` empty-shell/mode-b-template (being resolved via runtime-fill), content-vs-runtime-mismatch
