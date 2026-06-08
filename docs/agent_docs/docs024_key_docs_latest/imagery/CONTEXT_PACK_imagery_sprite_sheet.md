# Context pack — sprite-sheet imagery (fresh thread)

Starting context for building the sprite-sheet imagery feature. Full plan: **`PLAN_imagery_sprite_sheet.md`** (attach it first).

---

## State + next action

Framing is done; decisions are locked (below). Nothing built yet.

**Next action — Phase 0 (no code):** confirm the schema and the renderer's CSS-inclusion path, and lock the grid geometry. Run:
```
dbcontext -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema site_plan_imagery,assets,css_snippets
```
Then confirm: where the renderer pulls a site's CSS snippets in, and the JSONB hint columns on `site_plan_imagery` (do they hold arbitrary plan data?). Lock geometry at 3×3 @ 768² (256 px cells) unless the schema suggests otherwise.

## Decisions (locked)

- One **source sheet per site** (scope column allows page/section later; emit more rows then).
- Generate as a known **N×M grid**, slice on fixed coordinates.
- **One sheet + CSS `background-position`** (web sprite) — so there is no image-cropping step to build.
- Model **Banana / gemini-3-pro-image-preview**.
- Bullets/nav via CSS **`::before { background: … }`** (simplest reliable).

## Standing rules (the constitution)

Reuse before recreate; **check the schema with a fresh `\d` before any SQL** (snapshots are stale); new enum values are `text` + `CHECK`, not native enums; complexity in Go actions, workflows thin; every agent is an orchestrator and replies on the caller's responses topic; no `logger.Debug`; don't rename vars/fields silently; structural fixes over patches; plain language. Deploy path: git → GitHub Actions → Backblaze. Namespaces `ai-persona-system` and `kafka`.

## Attach — docs

`PLAN_imagery_sprite_sheet.md`, `FOCUS_imagery_assessment.md`, `PLAN_imagery_phase_2g.md`, `TODO_imagery_followups.md`, and the imagery hard-lessons (SDXL can't do flat icons → Banana; transparency abandoned for flat-grey chips; the single-image-vs-grid note) — so the model-choice and transparency battles aren't re-litigated.

## Attach — code (by phase; scope to the phase you're on)

- **Phase 1 (sheet generation):** the plan-builder imagery prompt; `internal/adapters/imagegenerator/dynamic_adapter.go` (kind→provider routing); the Banana provider; `generate_image_actions.go`; `imagery_helpers.go`; the store-asset action.
- **Phase 2 (sprite CSS):** the CSS-emit step you add; the `css_snippets`/`css_themes` path; `url_helpers.go` (`ImagePurposes`).
- **Phase 3 (consume):** the renderer; the component templates / CSS where bullets and nav are produced.

## Reuse targets (extend, don't recreate)

`site_plan_imagery` (+ `kind='sprite_sheet'`, grid plan in JSONB); the adapter's kind→provider routing (+ `sprite_sheet→Banana`); `assets` (sheet = one row; derived-asset columns left for a possible later physical-slice route); `css_snippets`; `ImagePurposes`; the planner prompt; the icon-chip CSS; a fulfilment check beside `check_unfulfilled_imagery_plan.go`.

## Decision forks / risks

- **Cell-content alignment (main risk):** the model may not put the intended glyph in the intended cell. First cut: ordered-grid prompt, then **eyeball and assign meanings after** generating (treat the sheet as "N coherent glyphs"). Vision-verify is Phase 4.
- **Geometry:** start 3×3 @ 768²; revisit if glyphs are illegible when shrunk (Phase 3 gate).

## Minimum set to start fast

`PLAN_imagery_sprite_sheet.md` + the Phase 0 `dbcontext -schema …` output + the imagery assessment doc. That's enough to settle Phase 0 and start Phase 1.
