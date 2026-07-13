# UPDATE — corrections to FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md

Dated 2026-06-19. The work-plan is dated 2026-04-11; the items below are now stale or wrong against the live
state, verified from `agent_definitions` (`site-design-planner` id `8b3cb270-ee0d-4df2-abb6-e266758d2747`,
v1.0.1065, updated 2026-06-18) and the `resolved_composition` specs for idea.uk (fresh) and gamesdesign
(adoption). Fold these into the work-plan (or replace the named sections).

## 1. Phase 3 (Site-Design-Planner) is BUILT and LIVE — not "not started"

`site-design-planner` exists and is active (created 2026-04-19, updated 2026-06-18). It resolves composition
— palette + layout + typography — **deterministically, no LLM**, and installs it atomically into
`css_themes` + `style_collections` + `sites.style_collection_id` + a `resolved_composition` spec, **before**
webdesign-agent renders. It runs for BOTH fresh and adopted builds (confirmed: idea.uk and gamesdesign both
have `resolved_composition` written by it).

So in the Phase 3 table: 3b (create the agent) and 3c (wire into the pipeline) are **done**. 3a/3d/3e/3f/3g
(nav + layout spec schemas, `populate_nav_tables` reading the navigation spec, header/footer template selection
from a layout spec, `hero_nav_merged`, adoption nav extraction) are **not yet confirmed** — verify separately;
the agent and the composition install are live, the nav/layout-spec wiring may not be.

Workflow (live): `ensure_site_record → validate_composition_inputs → check_ready → resolve_composition_layout
→ resolve_composition_typography → resolve_composition_palette → install_site_composition → complete`.
If identity/classification specs are missing it branches to `complete_unready` and queues a classifier
recovery item.

## 2. The palette/layout/typography CHOICE has moved out of webdesign-agent's prompt

Decision #3 and Phase 2b describe the design choice living in **webdesign-agent's LLM prompt** as a three-way
priority `design_intent → design_reference → industry`. For palette/layout/typography that is **superseded**:
`site-design-planner` now makes that choice with a deterministic Go cascade. webdesign-agent, if still in the
flow, renders CSS from the installed composition rather than choosing the palette. **Confirm webdesign-agent's
current role** — it may now be vestigial, with `render_css_from_spec_action` producing `/assets/css/styles.css`
from the installed `css_themes.css_template`.

The live cascade order (per the agent description) is **`design_reference → mission → design_intent`** — note
this differs from the doc's `design_intent`-first webdesign order, and it adds a `mission` slot the doc doesn't
mention. The palette extraction lives in the `resolve_composition_palette` Go action; typography in
`resolve_composition_typography` (font-family match with spec cascade); layout in `resolve_composition_layout`
(industry-tag overlap against the layout library). The winning slot is recorded in
`resolved_composition.lineage.{palette_source, typography_source, layout_source}`.

## 3. The cascade reads STRUCTURED values — adopted `design_intent` has them, fresh `design_intent` does not

This is the operative finding and the reason fresh builds look generic.

- **Adoption (gamesdesign):** the adoption `generate_design_intent` step writes a `design_intent` that carries a
  **structured `palette.reference_values`** block (and the crawl gives a structured `font_family`). Result:
  `palette_source = "design_intent_values"`, `typography_source = "fingerprint_font_family_match"`. (NB: a
  `design_reference` was present too, with a clean `suggested_mapping`, but the `design_reference` cascade slot
  did **not** win — the `design_intent` slot did. Worth confirming why in `resolve_composition_palette.go`.)
  The palette resolved on-source (cyan-on-near-black) — the palette mechanism works when `design_intent` is
  structured.
- **Fresh (idea.uk):** the classifier writes a `design_intent` whose palette is **prose only** — the parchment
  / ink / rust hexes live inside the `colour_mood` sentence with no structured palette field, and the fonts
  (Fraunces / IBM Plex) live in the prose `typography_mood`. With no `design_reference` (nothing crawled) and no
  structured palette/typography in `design_intent`, every cascade slot misses and it falls to
  `palette_source = "archetype_default"` + `typography_source = "fallback_sans_modern"`. That is why idea.uk
  shipped a generic palette instead of its parchment/rust intent.

## 4. The gap to close for fresh builds (Decision #6, "every build conceptually an adoption")

- Fresh sites need their `design_intent` to carry a **structured palette** (the shape the cascade's
  `design_intent` slot reads — confirm the exact key in `resolve_composition_palette.go`) and a structured
  **`font_family`**, i.e. the same shape adoption's `generate_design_intent` produces. The values are already
  determined for idea.uk (the hexes and fonts are in its prose); they need to be emitted structured at the
  point the fresh `design_intent` is written (the classifier / a fresh `generate_design_intent` equivalent), or
  lifted from prose into structure by a small step before `site-design-planner` runs.
- **Layout selection ignores design character.** `resolve_composition_layout` matches on industry-tag overlap
  only. idea.uk's tags hit `tool-portal-dark` (a dark layout) on 3 matches, against a `design_intent` that
  explicitly wants light/editorial — so a light-editorial site landed on a dark tool layout, and the
  hardcoded-dark header/footer components followed. Consider weighting layout selection by the `design_intent`
  scheme (dark/light) and character, not tags alone — otherwise a correct palette still sits inside the wrong
  layout.

## 5. Minor

Decision #9 ("design agent write-back goes to `css_themes` metadata, not `site_specs`") is only partly true now:
`site-design-planner` writes a `resolved_composition` **spec** (`site_specs`) as well as installing
`css_themes` + `style_collections`. The composition lineage is in `site_specs`.
