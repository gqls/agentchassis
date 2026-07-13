# Design pipeline — two-track overlap investigation (baseline for the rewrite)

Status: investigation, as of 2026-06-20. Captures the architecture *as found* in the
deployed code + DB. No workflow changes made yet. This is the baseline the rewrite
decision should work from.

Confirmed against: agent_definitions (site-design-planner, webdesign-agent,
build-site-planner, pageflow-builder, site-work-orchestrator, domain-research-classifier,
domain-submitter); fork_theme_composition.go; render_css_from_spec_action.go;
render_css_composition_helpers.go; install_site_composition_action.go; a routing count on
site_work_items.

---

## TL;DR

`webdesign-agent` and `site-design-planner` are **not** two rival agents both writing the
palette and colliding. The `025_palette_layout_typography_migration` (Phase 4.3, which is
the **deployed** render code) deliberately split design into a **two-stage base+override
pipeline**:

- **Stage 1 — site-design-planner** (`needs_composition`): deterministic *selection +
  install*. Picks layout / typography_set / palette, installs them as a `css_themes` row
  (3 FKs) + `style_collections` row + `sites.style_collection_id`, and writes the
  `resolved_composition` spec. **Renders nothing, deploys nothing.**
- **Stage 2 — webdesign-agent** (`needs_design`): LLM *override + render + deploy*. The
  `analyze_design` LLM produces a `design_spec.color_scheme` + typography; then
  `render_css_from_spec` loads Stage 1's installed composition as the **base**, overlays
  the LLM's values per a fixed merge rule, renders the layout template, and `git_commit`s
  `/assets/css/styles.css`.

So the genuine remaining overlap is narrower than "both own design": **both decide the
core palette slots and typography**, and at render the LLM's values win the core slots.
That has a direct consequence for the idea.uk fix (below).

---

## Who triggers what

`build-site-planner.emit_design` (`emit_design_items`, guarded on
`style_collection_id IS NULL`) queues **both** `needs_composition` and `needs_design` from
one step. Routing confirmed on site_work_items:

| item_type | handler_agent | rows |
|---|---|---|
| needs_composition | site-design-planner | 2 |
| needs_design | webdesign-agent | 2 |

webdesign-agent is also called *directly* (not via a work item) by **pageflow-builder**
(`apply_site_design`) and **site-work-orchestrator** (`apply_site_design`). site-design-planner
is only reached via `needs_composition`.

---

## Do they share actions? Mostly disjoint; one shared file.

**Design-doing actions are disjoint** (different names, no sharing):

| Concern | Stage 1 (site-design-planner) | Stage 2 (webdesign-agent) |
|---|---|---|
| Layout | `resolve_composition_layout` | — |
| Typography | `resolve_composition_typography` (`resolveTypographySet`) | `analyze_design` (LLM) |
| Palette | `resolve_composition_palette` (`createPalette`) | `analyze_design` (LLM) |
| Theme + collection write | `install_site_composition` | `fork_theme_from_site` (only when `should_fork_theme`) |
| Render CSS | — | `render_css_from_spec` |
| Deploy styles.css | — | `git_commit` |

**Shared:** (a) generic utilities `read_site_spec`, `ensure_site_record` — not a conflict;
(b) one file, `fork_theme_composition.go` — Stage 1 calls `createPalette` /
`resolveTypographySet`, Stage 2's fork path calls the siblings `createPaletteForFork` /
`resolveTypographySetForFork`. Same low-level code, separate entry points. No duplicated
*responsibility* in that file — it's a shared library.

So the earlier worry ("actions duplicated in responsibility") is largely **no** — the
duplication is at the *decision* level (palette + typography decided in both stages), not
the action level.

---

## The render merge — the key reframe

`render_css_from_spec` (Phase 4.3 cutover, deployed) does, in order:

1. `loadThemeComposition(db, config, siteID, themeName)` — resolves the theme composition
   (palette + layout + typography in one query). Strongly indicated to read the site's
   **installed** composition via `sites.style_collection_id → style_collections.css_theme_id
   → css_themes(palette_id, layout_id, typography_set_id)` when `siteID` is present, else
   fall back to `theme_name` (default `standard-brochure`). *(One caveat: I read both
   callers and the install side, but not `loadThemeComposition`'s exact SQL — see Open
   questions.)*
2. `specPalette/specTypo/specSpacing = extractDesignSpecRawMaps(...)` — the LLM's
   `design_spec.color_scheme` / typography / spacing.
3. `mergedPalette = buildPaletteMap(comp.Palette, specPalette)`;
   `mergedTypo = buildTypographyMap(comp.Typography, specTypo)`.
4. Render the **layout template** with `{{palette}}`/`{{typo}}`/`{{token}}` helpers; append
   component CSS snippets; return the CSS (webdesign-agent then commits it).

Merge authority (from `render_css_composition_helpers.go`):

| Token group | Winner |
|---|---|
| Palette **core** slots: primary, secondary, accent, background, surface, text, text_muted, border | **design_spec (webdesign-agent LLM)** when non-empty |
| Palette **specialised** slots: primary_hover, heading, hero_title, cta_bg, footer_text, … | **composition theme (Stage 1)** |
| Typography (all) | **design_spec (webdesign-agent LLM)** |
| Structure tokens (container widths, paddings, radii) | **layout (Stage 1)** — spec doesn't contribute |

So: Stage 1 owns layout/structure + specialised palette slots + the *base* of everything;
Stage 2's LLM owns the **core palette slots + typography** at render time.

---

## What `install_site_composition` actually does (Stage 1 tail)

1. INSERT `css_themes` with all three FKs (palette_id, layout_id, typography_set_id);
   `css_content` left **empty** ("webdesign-agent populates these later when it renders").
2. INSERT `style_collections` pointing at the new theme via `css_theme_id` (collections have
   no palette/layout/typography FKs of their own — the renderer joins through `css_themes`).
3. UPDATE `sites.style_collection_id` — **guarded**, only if still NULL; **errors if already
   set** ("re-resolve not supported; clear it manually to force re-install"), with a
   race-loss check.
4. Supersede the old `resolved_composition` spec + INSERT the new one (lineage record).
5. **No render, no deploy, no `needs_design`/`needs_rerender` emit.**

---

## The genuine problems to fix (not "both write the palette")

1. **Core palette + typography are decided twice, and the LLM wins the core slots at
   render.** Stage 1's palette cascade resolves the core slots; Stage 2's LLM produces core
   slots that override them in `buildPaletteMap`. So Stage 1's core-palette work only shows
   through for slots the LLM leaves empty. (Stage 1's layout, structure, and specialised
   slots are *not* overridden — those genuinely belong to Stage 1.)

2. **Sequencing / no re-render after install.** Both work items are queued together. If
   `needs_design` (Stage 2) renders **before** `needs_composition` (Stage 1) installs,
   `loadThemeComposition` finds no installed composition and renders on the
   `standard-brochure` base. Nothing re-triggers a render after install
   (`install_site_composition` emits nothing), so a late install may never reach
   `styles.css`. Ordering depends on `emit_design_items` priorities + dispatch order — not
   yet confirmed (Open questions).

3. **Two write paths to `css_themes` + `style_collections`** (`install_site_composition`
   and `fork_theme_from_site`). They run in different scenarios (fresh/adopted composition
   vs webdesign-agent fork), but both grow the theme/collection library and both touch
   `sites.style_collection_id` semantics — worth confirming they can't both fire for one
   site and fight over `style_collection_id`.

---

## Direct consequence for the idea.uk colour fix

The classifier writes idea.uk's parchment palette into `design_intent.colour_mood` (prose),
which **nothing** reads. Both design stages read `design_intent.palette.reference_values`
(structured): Stage 1's cascade slot, and Stage 2's `analyze_design` prompt
(`{{if .design_intent.palette}}… reference_values …}}`). idea.uk has no such block, so:

- Stage 1 cascade → no palette → falls through (dark seed / default).
- Stage 2 LLM → no `design_intent.palette` shown, `colour_mood` ignored → invents core
  colours (blue) → **overrides** whatever Stage 1 produced for the core slots.

Therefore a composition-only fix (e.g. forcing Stage 1's palette to parchment) would be
**defeated by the Stage 2 LLM override** on the core slots. The fix that reaches **both
stages** is the classifier emitting structured `design_intent.palette.reference_values`
(+ `typography.reference_values`), mirroring adoption's `generate_design_intent`. With that
in place: Stage 1 base = parchment, Stage 2 LLM starts from parchment, merge = parchment.
This is why the classifier schema change is the correct idea.uk design fix, independent of
how the two-track overlap is resolved.

---

## Open questions (needed before/with the rewrite)

1. `loadThemeComposition`'s exact SQL — confirm it reads `sites.style_collection_id →
   style_collections → css_themes` for the site, and the precise condition under which it
   falls back to `standard-brochure`. (File: the helper defining `loadThemeComposition`,
   not in `render_css_from_spec_action.go`.)
2. `emit_design_items` — does it set priorities/ordering so `needs_composition` is dispatched
   before `needs_design`? And does the dispatch loop honour that strictly? (File:
   `emit_design_items` action; dispatch loop ordering.)
3. Can `install_site_composition` and `fork_theme_from_site` both fire for the same site in
   one build, and if so what happens to `style_collection_id`?

---

## Decision options for the rewrite (to choose — not yet decided)

The founding principle is "distinct responsibilities, minimal overlap." The remaining true
overlap is *core palette + typography decided in both stages*. Options:

- **(A) Keep the base+override split; make the boundary clean and the ordering guaranteed.**
  Declare it explicitly: Stage 1 owns layout + structure + full palette + typography
  selection (deterministic, from specs); Stage 2 owns render + deploy and only *fills slots
  Stage 1 left empty / handles legacy no-composition sites*. Flip/limit the `buildPaletteMap`
  core-slot rule so a Stage-1 palette that came from a structured source is **not**
  overridden by the LLM (LLM becomes fallback, not master, for core slots). Guarantee
  `needs_composition` before `needs_design`, and have `install_site_composition` trigger the
  render. Lowest risk; keeps the LLM for the no-spec case; removes the true overlap.

- **(B) Collapse to one design agent.** Either site-design-planner gains render+deploy and
  webdesign-agent's design role is retired (loses the LLM's expressive per-business palette),
  or webdesign-agent absorbs layout selection and the composition install is dropped (loses
  deterministic library composition + lineage + structure tokens). Each loses something the
  other provides.

- **(C) Make design fully deterministic.** Composition owns all of palette/typography/layout;
  the LLM no longer overrides core slots; webdesign-agent reduced to legacy/scrape-only
  (no installed composition). Most predictable; loses LLM expressiveness entirely.

Recommendation leans **(A)** — it matches "minimal overlap" without throwing away either
agent's distinct value, and the fixes are incremental (ordering guarantee + render trigger +
core-slot authority flip) rather than a rebuild. But this is a product-direction call.

---

## Note: architecture doc 002 is stale on this

`002_system_architecture.md` (Responsibility Boundaries) names webdesign-agent as the sole
design agent and does not mention `site-design-planner` / the composition path at all. It
should be updated to the two-stage model **once the rewrite direction is chosen** — updating
it now would bake in a model we're about to change.
