# PLAN — component regeneration silently empties dependent pages

Phased plan for the next chat. Each phase has a goal, concrete steps, and a done-when. Gates are noted.
Use `RUNBOOK_component_regen_clobber.md` for the SQL/ops referenced here. Confirm code paths against the
checkout (see `BUNDLE_component_regen_clobber.md` §3) before reasoning about them as fact.

---

## Phase 0 — Confirm and reproduce (no writes)

Goal: independently confirm the mechanism before touching anything.

- RUNBOOK R1: inspect `content_components` `fdd92ad4` — its current `input_schema` field names and
  `html_template` placeholders; capture the real field list (don't trust the from-memory list).
- RUNBOOK R2: enumerate the live instances off `component_id` and resolve each to its site domain; confirm
  the count and that `gamesdesign.co.uk` is not among them.
- RUNBOOK R3: for one affected page, compare its `content_data` keys against the component's current
  `input_schema` fields; confirm the keys do not intersect (this is the bind failure, made concrete).
- RUNBOOK R6 (read-only form): confirm the affected pages' `rendered_html` is empty / the band is absent on
  the live site.

Done-when: the key mismatch and the empty render are both observed directly, and the affected set is known.

---

## Phase 1 — Locate the exact code path (no writes)

Goal: pin the two points the fix touches.

- Read `UpdateComponentHTMLAction` (`grep -rn "func UpdateComponentHTMLAction"`). Determine: does it re-render
  dependents inline after writing the new `html_template`? Where does it enumerate dependents (expect a
  `SELECT ... FROM page_components WHERE component_id = $1`)? Does it pass each page's existing `content_data`
  straight into the render, or does it remap anything first? The 15:06:12.940 → .956 gap predicts an inline loop.
- Read `RenderTemplate` + `RenderContext` in `component_library.go`. Confirm the bind is purely by exact field
  name (placeholder `{{.x}}` ↔ `ContentData["x"]`) with no alias/fallback, so a rename guarantees an empty fill.
- Read the regen trigger in `component_selector.go` (the "empty `input_schema` → needs regeneration" path) to
  understand what causes a regen in the first place and whether renames are intrinsic to it.

Done-when: you can point at the line that re-renders dependents and the line that binds by name, and explain
why a rename produces empty output.

---

## Phase 2 — Choose the route (decision, no writes)

Goal: pick a direction with the code in front of you. Treat the handoff's three routes as candidates, not an
either/or; they compose.

- Weigh: preserving field names across a regen (least disruptive to existing data, but constrains the
  regenerator); migrating dependents' `content_data` keys old→new at re-render time (fixes the data path, needs
  a reliable old→new map — note `stat_N_icon` has no new target and `cta_*` are new); a fail-loud guard on
  "re-render emptied a populated section" (cheap safety net regardless of which structural route is chosen).
- Prefer the structural cause over a symptom patch: a guard alone stops silent shipping but leaves regens
  breaking bindings. A name-preserving or migrating regen addresses the cause; the guard is complementary.
- Keep the fix in Go (in the regen/render path), not in a workflow and not in SQL. Reuse existing helpers
  (`GetComponentByID`, the existing rerender action) rather than adding parallel ones.

Done-when: a chosen direction (likely structural fix + guard) with a one-paragraph rationale tied to the code.

---

## Phase 3 — Implement (Go change)

Goal: implement the chosen route with minimal surface area.

- Make the change in the smallest set of functions identified in Phase 1. If migrating keys, derive the
  old→new map from the component's own schema delta where possible rather than hardcoding, and handle the
  non-1:1 cases (dropped/added fields) explicitly.
- Do not rename variables that other code depends on except where intended; note any intentional rename.
- Keep `logger.Debug` out; use a level that shows in logs for the new guard's failure path.
- Build (`go build ./...`), then ship via the normal path (GitHub Actions → Backblaze → new chassis image →
  bump `image_tag` on the affected agent rows). Back up any agent whose config/image you change.

Done-when: builds clean; a regen on a test component that renames a field no longer empties its dependents
(or fails loud if the guard route), verified on a throwaway component, not on a live shared one.

---

## Phase 4 — Recover the five affected pages (reuse existing rerender)

Goal: restore the blanked bands without an LLM.

- RUNBOOK R4: back up the five `page_components` rows (and the component row) first.
- RUNBOOK R5: align each affected page's `content_data` keys to the current `input_schema` (old→new), then
  trigger the no-LLM `rerender_page_sections` path — one `page_rerender` work item per page through the
  existing `rerender-pages` / `build-dispatch-loop` / `page-rerender` chain. Do not build a new rerender path.

Done-when: each of the five pages re-renders the band with visible content.

---

## Phase 5 — Verify and close

Goal: confirm recovery and that the regression can't recur silently.

- RUNBOOK R6: for each recovered page, confirm `rendered_html` is non-empty, `content_hash` changed, and the
  band is present on the live site.
- Confirm the Phase 3 guard would have caught this (e.g. force a rename on a test component and observe the
  loud failure rather than a silent drop).
- Brief end-of-work summary in chat (no summary document unless asked). Update whatever running notes the
  next chat keeps.

Done-when: five pages restored, the structural fix is deployed, and a deliberate rename now surfaces loudly.

---

## Gates and reuse reminders

- Phase 4 (recovery) can run independently of Phase 3 (the fix) — recovery only needs the current schema and the
  existing rerender path. But do not recover before Phase 0/1 confirm the mapping, or you may re-key wrongly.
- Reuse: `RenderTemplate`/`RenderContext`/`GetComponentByID` (render), the existing `rerender_page_sections`
  action and `rerender-pages` agent (recovery), `snapshot_agent`/`revert_agent` (agent backups). Adapt before
  adding.
- Concurrency: re-check component freshness immediately before any write (RUNBOOK R7); another chat may touch
  the same components.
