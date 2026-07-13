# SCOPE — Phase I2: sprite-sheet bullets & list treatment

**What this is.** Phase I2 of the imagery best-in-class workstream (see
`PLAN_imagery_best_in_class.md`). Goal G4: themed graphic bullets, nav
accents, and section glyphs — visually coherent, cheap, and served as ONE
small download. The design was locked earlier in
`PLAN_imagery_sprite_sheet.md` / `CONTEXT_PACK_imagery_sprite_sheet.md`; this
document is the implementation scope after a FRESH schema/architecture check
(2026-07-11) and reconciles one deviation from the original plan.

**Core idea (locked).** A site generates ONE coherent source image — an N×M
grid of small glyphs in a single style (the diffusion model's "answer several
icons with one gridded image" tendency, harnessed rather than fought). The
browser slices it with CSS `background-position`; bullets/nav use
`::before { background: … }` + a `.sprite-<name>` class. No Go image cropping —
"slices" are CSS rules computed from grid geometry. One generation, one asset,
one stylesheet.

---

## Confirmed reality (fresh `\d` + code, 2026-07-11)

| Assumption in the locked plan | Reality | Impact |
|---|---|---|
| `site_plan_imagery.kind` is `text`+CHECK, extensible | ✅ `chk_kind` = logo/hero/illustration/icon/infographic; mirrored in `validImageryKinds` (write_site_plan_action.go:183) | Add `sprite_sheet` in BOTH (migration + Go), together — the standing rule. |
| Grid plan rides JSONB hint columns | ✅ `style_hints jsonb`, `constraints jsonb` exist | Put `{rows, cols, cell_names[], style}` in `style_hints`. |
| Adapter routes by kind | ✅ `dynamic_adapter.go` switch (icon/logo/illustration/infographic → Banana; else Stability) | One-line: add `sprite_sheet` to the Banana case. |
| `ImagePurposes` extensible | ✅ map in url_helpers.go | Add `sprite_sheet` (768×768 **jpg** — revised from png 2026-07-13: png exceeds Kafka commit msg-size + 80KB budget). |
| **Sprite CSS = a site `css_snippet`** | ❌ **`css_snippets` is a GLOBAL library** (name, css_content, `applies_to` matched to component lists) — NOT per-site. Site CSS is assembled by `render_css_from_spec` → deployed to `/assets/css/styles.css` → `<link>` in head. | **DEVIATION — resolved below:** deliver sprite CSS as a SEPARATE committed file `/assets/css/sprites.css` + a `<link>` injected into `<head>`, reusing the Turn-25/26 head-injection + git-commit patterns. Cleaner: site-specific, decoupled from theme CSS, no re-run of render_css_from_spec. |

**Delivery decision (the one real design change):** the sprite stylesheet is
site-specific (it references that site's sheet URL and computed positions), so
it does NOT belong in the global `css_snippets` table. Deploy it as
`/assets/css/sprites.css` (git-committed like styles.css) and inject
`<link rel="stylesheet" href="/assets/css/sprites.css">` into the head —
exactly the mechanism just built for favicon/OG (`injectBrandHeadTags` in
render_site_components, and asset-deployer's git-commit path). This is the
single biggest reuse win and removes the plan's only fuzzy piece.

---

## Build breakdown (each item is small; order = dependency order)

1. **Schema + Go kind (`sprite_sheet`).**
   - Migration: extend `chk_kind` to include `'sprite_sheet'` (drop/recreate
     CHECK; backup per doc 009). SQL artifact in this folder.
   - Go: add `"sprite_sheet": true` to `validImageryKinds`
     (write_site_plan_action.go). *Together with the migration — mismatched
     constraint vs mirror rejects plans.*
   - `ImagePurposes["sprite_sheet"] = {768, 768, 90, "png"}` (png — thin
     line-art glyphs; the jpg-muddies-lines lesson).

2. **Adapter routing.** `dynamic_adapter.go`: add `sprite_sheet` to the
   Banana `case`. (Banana/gemini-3-pro-image-preview — the coherent-grid model.)

3. **Planner prompt.** Teach build-site-planner to emit ONE
   `imagery.site[]` entry `kind: "sprite_sheet"`, key `sprite_sheet_main`,
   with `style_hints: {rows, cols, cell_names, style}`. Prompt: a clean N×M
   grid, ONE glyph per cell in reading order, single flat style/colour, flat
   selectable background (NOT transparent — the abandoned-transparency lesson),
   glyphs relevant to the site's verticals. Start 3×3, 9 cell_names.
   SQL prompt patch (mirrors the 053 imagery-block patches).

4. **Generation.** No new code if the item flows through image-build-handler
   as a normal `needs_imagery` row (store_asset with purpose/asset_key from
   spec). CONFIRM image-build-handler's kind→size path honours sprite_sheet's
   768² (it reads ImagePurposes/kindDefaults) — likely a config check only.

5. **Sprite-CSS emit action** (`emit_sprite_css` or reuse asset-deployer mode).
   Computes, from the stored sheet's grid plan + fixed dims:
   ```
   .sprite-<name>{background-image:url(/assets/images/sprite-sheet-main.jpg);
     background-position:-<c*cellW>px -<r*cellH>px;
     width:<cellW>px;height:<cellH>px;
     background-size:<sheetW>px <sheetH>px;display:inline-block}
   ```
   plus a `::before` bullet helper. Commits `/assets/css/sprites.css` to the
   site repo (git-adapter, same shape as favicon/OG). Pure string compute —
   no image processing. **Recommended home:** a new `brand_head`-sibling mode
   on asset-deployer, or a small standalone action; dispatched via a
   `needs_sprite_css` work item (the pattern proven in Turn 26).

6. **Head link injection.** Extend `injectBrandHeadTags` (or a sibling) to add
   the `sprites.css` `<link>` — idempotent, fleet-wide, no per-site template
   change. Guard: only when the site has an active `sprite_sheet` asset (avoid
   a 404 link on sites without a sheet).

7. **Consume.** Wire ONE real surface first — list bullets: a component/CSS
   rule `li::before { content:''; } li.sprite-check::before { … }` or a
   site-level list style. Keep it to one section on one page for the Phase gate.

8. **Fulfilment check.** Sibling of `check_unfulfilled_imagery_plan.go`:
   sprite_sheet planned but no asset → needs_imagery; asset exists but no
   sprites.css committed → needs_sprite_css. Register on design-discovery-agent.

---

## Phasing (stop and eyeball between phases)

- **I2.0 — schema + kind (no visible change).** Migration + Go kind + adapter
  route + ImagePurposes. Deploy. Gate: a hand-inserted `sprite_sheet` plan row
  validates and passes chk_kind.
- **I2.1 — sheet generation.** Planner emits the row (or hand-seed for
  robot-hands); generate via Banana; store. **GATE (human, the real risk):
  eyeball the sheet — clean coherent flat-background grid? Then ASSIGN cell
  meanings to what actually landed** (don't assume cell (r,c) == requested
  glyph). Record the true cell→name map back into style_hints.
- **I2.2 — sprite CSS.** Emit + commit sprites.css from the (verified) grid.
  **Gate: a test page shows the right cell for a given `.sprite-<name>`.**
- **I2.3 — consume + head link.** Inject the `<link>`; wire bullets on one
  section of robot-hands. **Gate: rendered live, readable at bullet size,
  accents land, one ≤80KB download.**
- **I2.4 — later (only if it earns it).** Per-page/section sheets (more rows
  at finer scope — the schema already allows it); a vision step to auto-map
  and re-request on mismatch; more cells.

---

## Risks / open decisions

- **Cell-content alignment (THE risk).** The model may not place the intended
  glyph in the intended cell, or drift in count/order. Mitigation (locked):
  ordered-grid prompt + eyeball-and-assign-after at the I2.1 gate; treat the
  sheet as "N coherent glyphs" and map names once generated. Vision
  auto-verify is I2.4.
- **Low-res legibility.** Glyphs shrunk to bullet size (~16–24px from 256px
  cells) must stay readable — tune cell count / display size at the I2.3 gate.
- **Coherence vs relevance.** One site sheet trades per-glyph relevance for
  visual coherence — accepted for v1; per-section sheets are the I2.4 valve.
- **Decision for user:** which surfaces first? Recommend **list bullets**
  (highest readability payoff, matches the brief's "relevant graphics as
  bullet points") over nav accents. Confirm 3×3 geometry and the starter
  cell vocabulary for robot-hands (technical: check, gauge, gripper, cog,
  chart, download, arrow, info, warning?).

## Reuse ledger (what already exists — do NOT rebuild)
- Head `<link>` injection + git-commit-to-site-repo: built Turns 25–26
  (`injectBrandHeadTags`, asset-deployer brand_head mode, dispatch-routed
  work item). Sprite CSS delivery is the SAME shape.
- Kind→provider routing, ImagePurposes, store_asset/asset_key, the
  needs_imagery → image-build-handler → asset-deployer chain: all reused as-is.
- Discovery-check pattern + registration: `check_unfulfilled_imagery_plan.go`
  is the template.

## Effort
Comparable to a single earlier phase. Mostly config + one CSS-emit action +
one planner prompt patch + one discovery check; the delivery mechanism is
already built. Biggest cost is the human eyeball gates (I2.1, I2.3), by design.
