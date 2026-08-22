# PLAN — bugs 235 / 155 / 071 close-out (2026-08-22)

**Workstream:** finish the residuals on three heavily-worked open bugs, robustly
(framework-wide over per-site), and move the files that meet the fixed-AND-live bar.
Full approved plan (verified state, owner decisions, phases):
`/home/ant/.claude/plans/please-can-you-look-sharded-hippo.md` — this file carries the
decisions and their reasons; NOTES carries what actually happened.

## Owner decisions (taken 2026-08-22, in-session)

1. **235**: DELETE the stale `logo.jpg` objects via the framework (`retract_asset_files`,
   dry-run first, guards read). The bug file reserved this as the owner's call; it is now taken.
2. **071 scope**: take the persistence-hole fix and the section_editor CTA fabrication fix.
   NOT taken now (owner-ruled): NormalizePagePath dir-index class; fragment-capability RFC.
   Both stay recorded in 071.
3. **155**: full depth — behavioural proof + migration-324 hygiene + retire the readerless
   `{purpose}_uri` writers (contributes to 209 Phase 3).

## Why each piece, in one line

- **071 persistence fix**: a valid build whose warnings produce no repair currently writes
  NOTHING durable (issues die with collected_data ~24h) — 071's candidate 1, never fully done.
- **section_editor CTA fix**: `section_editor_actions.go:783-785` still fabricates
  `cta_url:"/contact.html"` upstream of the template guard — bugs_open/203's exact class,
  on a path 203 never names; correct-or-absent is the ruled pattern (owner 2026-07-27 family).
- **155 writer retirement**: zero readers remain (Go grep + live agent_definitions +
  workflow_templates + active content_components censuses, 2026-08-22); a readerless
  last-write-wins cache is the exact defect 155 was filed about — delete the class, not scope it.
- **324 hygiene**: the migration is applied live but UNTRACKED in git and unrecorded; the
  runner halts at its guard on every `--apply`. Commit + `--record-only`.
- **155 contract migration**: asset-deployer `input_contract` still requires `s3_uri` —
  stale since 324/91dda3243; it blocks the file's own closure recipe (asset_id-only dispatch).
- **235 deletion**: 0 references fleet-wide re-measured 2026-08-22; the running chassis
  (v1.0.1323) carries `retract_asset_files` (marker probe); guards refuse anything unsafe.

## Key design decisions (from the design round, 2026-08-22)

- 071 fix hooks the existing recorder family via `LogActionEntry` (RFC_012 seam), new code
  `CONTENT_VALIDATION_WARNING_DETAIL`, one row per build, dedupe against repaired hrefs via a
  shared `repairedHrefSet` helper extracted from `annotateLinkRepairs` (no drift possible).
- section_editor fix is a pure deletion; `contextToInterfaceMap` (LNK-005) is already
  correct-or-absent, 29/30 active cta_url templates carry the guard, the 30th supplies its
  own content_data.
- 155 deletions keep the `_url` writes and `updateSiteContentField` itself (live caller);
  only the four `_uri` writes go. Post-roll marker pair: `"Failed to store URI"` ABSENT /
  `"Failed to store URL"` PRESENT.
- Ordering: baseline 209 proof runs on the CURRENT binary BEFORE the writer retirement rolls,
  so pre/post proofs bracket the deletion.

## Out of scope (deliberately)

- RFC_029 Phase 2 / `findFieldRecursive` — `staged_component_build`'s lane.
- NormalizePagePath resolvability split — full design is in the plan file for whoever takes
  it (identity vs servable-form comparison; unanchored TrimSuffix fix). Wants its own bug
  file + 090 + council round.
- Fragment capability (stable section ids) — 071's own conclusion: architecture round, not
  a bug patch.
