# 285 — the tool-improver rewrote a SHARED section template from a single-tool finding: one "fix" armed 115 pages fleet-wide to become the Asset Path Formatter on their next re-render, and reported success

Filed 2026-08-15 at the owner's direction, so the near-miss has its own case rather than
living inside `bugs_open/281`'s addendum (281 is the audit blind spot; THIS is the write-path
defect that the blind spot's machinery drove into). **Status: OPEN — fix committed at HEAD
and council-approved, NOT yet live (Go fence awaits the next roll). Damage fully restored
2026-08-15; zero pages ever served the bad markup (verified at the artefact).**

**First-hand verification substituted for a 090 run, per the 2026-07-31 ruling:** every claim
below is either a `component_versions` / `site_work_items` row quoted by id, a code line read
at HEAD, or a served-page check performed during the restoration. The fixing lane
(`bugfix_281_tool_audit_ported/`) independently derived the same mechanism first; this file
consolidates both sessions' evidence.

## What happened (both firings — this was not a one-off)

The shared `ported-page` wrapper (`content_components` `a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef`,
**115 page placements across sites** — every imported tool, learn-article and prose page uses
it; its whole job is a 77-char `{{.body}}` passthrough plus wrapper styles):

| when | what | evidence |
|---|---|---|
| 2026-08-05 00:35Z | **First firing.** `update_component_html` rewrote the shared template (pre-state snapshotted as v1, the 77-char passthrough) | `component_versions` v1, `changed_by='update_component_html'` |
| 2026-08-08 18:14Z | component-creator regenerated it (v2, 7,714 chars) — recovery-by-coincidence, not a fix | v2, `changed_by='component-creator:regen'` |
| 2026-08-14 18:48:38Z | **Second firing, the near-miss.** Tier-2 acceptance had failed `asset-formatter` on webdesign.co.uk ("seeded-rows: anchor `.asset-row` absent from deployed page") and filed `improve_tool` carrying the SHARED `component_id`. The tool-improver "fixed" the missing anchor by writing **8,864 chars of Asset Path Formatter markup into the shared wrapper's `html_template`** and flipping **all 115 placements to `build_status='pending'`** — a live re-render signal (`chrome_link_policy.go:133`). Item `1a0fa071-c7b3-4e7f-af10-a996333a16ce` completed 18:48:56 — **success reported 18 seconds after arming the fleet.** | v3 pre-edit snapshot (4,664 chars, the true pre-poison state); the item row; `LANDMINES`/281 addendum measurements |
| 2026-08-15 (≈21 h window) | Nothing re-rendered — the ONLY reason nothing broke. Assembly serves stored `page_components.rendered_html`; the template is consulted on re-render, and `content_data` on ported instances is a provenance stub, so any single re-render = that page becomes the Asset Path Formatter (or worse: stub body through a tool template). Discovered by the `bugfix_281` lane; verified unpropagated. | their Finding A |
| 2026-08-15 ~17:00Z | **Restored, owner-sanctioned:** template back to v3's content; poisoned state banked as v4 (`change_source='manual-restore'`); 114 `pending` placements un-flipped to `deployed` (guarded); the one already-`deployed` instance (`learn-ai-builders-content-first`, touched 18:51Z) checked clean. | 281 contribution block, commit `9dd7cea0a` |

## Root cause — three individually-reasonable defects composed

1. **The producer lost the instance identity.** Tier-2 acceptance (`check_tool_acceptance.go`)
   admits ported tools but filed `improve_tool` with the shared `component_id` — the only
   identity a ported tool has (its code lives per-instance in `page_components.rendered_html`,
   but the item's contract points at the component).
2. **The loader universalised it.** `load_tool` resolves `WHERE cc.id=$1 … LIMIT 1`, ignoring
   `spec.page_id` — on a 115-placement component it loads an ARBITRARY instance and presents
   it as "the tool source" (same shape as 281's mechanism 2).
3. **The writer had no census.** `update_component_html` wrote `html_template` keyed by
   `component_id` with no check of `component_level` or placement count. A fix scoped to one
   page's finding was applied to a fleet-wide shared artefact, silently, with a success status.

The fingerprint that makes this a class, not an incident: **the blast radius of a template
write is its PLACEMENTS, not its finding** — and nothing on the write path counted them.
It fired twice in nine days; only re-render latency kept it invisible both times.

## The fix (committed, awaiting the roll)

- `25f92a967` + council-approved follow-up `d7b2d9994` (Council-Reviewed: 360ae540):
  `sharedComponentWriteCheck` in `component_write_guard.go` — `update_component_html`
  **refuses** a write to a non-tool component placed on >1 page unless the caller sets
  `allow_shared_component_write` (opt-in, default OFF; refusal names the placement census).
  Fail-closed on the non-tool path; tool forks warn+proceed. Mutation-proven in the commits.
- Producers stop creating the poison item: ported findings from `tool_health` AND
  `check_tool_acceptance` now file handler-less `ported_tool_fix` (`needs_human_review`)
  instead of `improve_tool` at the shared component. `load_tool` pinned to
  component **and** page (seeds 425/426 — config, already applied).

## Residuals — named, not fixed here

- **`fix_component_template` is the one other page-aware `html_template` writer and is NOT
  fenced** (writer census in `d7b2d9994`; recorded open there).
- **Tier-4's judge** re-derives components with the same `LEFT JOIN … LIMIT 1` arbitrary-
  instance shape (281 addendum Finding B) — produces a verdict, files nothing; unfixed.
- The v1 firing's triggering item was not exhumed (the 08-05 write predates the current
  item retention on that path); the v1 snapshot is the evidence it happened.

## How to close (fixed AND live)

1. After the next roll, verify the fence at the artefact: stamp check
   (`git merge-base --is-ancestor d7b2d9994 <stamp>`), then **induce** a refusal — a synthetic
   `improve_tool` at the shared component must draw the refusal verdict and a
   `ported_tool_fix` row, not a template write. (A quiet fleet is not evidence; the guard must
   be seen refusing — the induced-failure discipline from RFC_006.)
2. Confirm the shared wrapper's template is still v3-content (4,664 chars, `{{.body}}`) and
   placements all `deployed` — i.e. no third firing between fix-commit and roll.
3. Then move this file to `bugs_closed/`, and strike the residuals that remain open into
   their own tracking (or this file stays open on the `fix_component_template` half).

## Relations

- `bugs_open/281` — the parent arc (audit blindness); its addendum Finding A is this incident's
  discovery record; its Track 1 commit carries this fix.
- `bugs_closed/012` — the same agent destroying a component by truncation (output-side); this
  is the addressing-side sibling: right content, wrong scope.
- `bugs_closed/021` — "durable write guard covers one path only": the writer-census discipline
  this fix now applies to `html_template` writers, with one writer still outstanding.
- 016b §9: "A fixer addressed by COMPONENT fixes the TEMPLATE — and a shared template's blast
  radius is its placements" (added 2026-08-15, cites this file).
