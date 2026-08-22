# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-22 ~13:00Z. Supersedes `HANDOFF_2026-08-21_continue_here.md`.

**STATE: 28 of 63 rebuilt and serve-graded (DB-corrected count — the old docs' "28" was 27; today's
blueprint-compiler makes it a true 28). Phase A complete. Phase C (13 external-script tools) is OPEN
with #1 DONE end to end. 35 remain: 12 Phase C + 23 Phase B (≥8 KB self-contained; five rich apps
LAST, one at a time, owner-reviewed). NOTHING IN FLIGHT — no open add_tool, no pending retires.**

Read: this file → `PLAN_2026-08-15_…` → `RUNBOOK_…` (note THREE new sections: the post-retire
tombstone re-read; the Phase C asset-half recipe; the retire-race rules all stand) → `NOTES_…`
(newest at bottom — the 2026-08-22 11:06Z and 12:35Z entries are today's two incidents) →
`SUMMARY_2026-08-21_phase_a_complete.md`.

## What happened on 2026-08-22 (all repaired; read before trusting any older claim)

1. **Four retired ported slots were RESURRECTED on 08-21** by a `literal_markdown` section-edit
   canary (the 277 lane's route): the check scans tombstones and the section-editor promoted them
   back to `approved`; sweeps published grid-generator, json-cleaner, noise-generator and
   text-extractor with BOTH tools stacked for ~19 h. Re-retired + corrective assembles + all four
   serve-graded PASS (NOTES 11:06Z). **Class bug `bugs_open/360`; the writer-door fix is COMMITTED
   (`1cd184f6e`, council APPROVED r1 corr `4007ce96`) but INERT until the next chassis roll** —
   until then the RUNBOOK's post-retire tombstone re-read rule is mandatory. Filer scoping and the
   486 batch-resurrect predicate remain with the 277/283 lanes (CONTRIBs delivered).
2. **The 435 `adopt_existing_page` flag had been REMOVED from tool-generator's live config** by an
   un-snapshotted write (window post-516 ~16:55Z 08-21 → 08:36:05Z 08-22; writer UNIDENTIFIED — no
   backup row, no ledger row). Every adopt-route add_tool fleet-wide died at `save_tool` (23505,
   item complete/error NULL). **RESTORED by migration 558** (snapshot + ledger + doc_note; council
   post-hoc corr `a367b63e`, verdict owed a read). One casualty (this lane's first #28 filing).
   **Check the flag before ANY filing** — the query is in NOTES 12:35Z; if it is gone again,
   somebody's process is re-removing it and THAT is the find.

## Phase C recipe (proven on #28 `tool-blueprint-compiler` today)

Six steps as ever, plus the Phase C extras:
- **The brief comes from the live tool's behaviour**: fetch the page AND its `<script src>` file(s)
  cache-busted; read the sidecar in full (blueprint-compiler's was 7 KB, self-describing). Expect
  the "asserts something untrue about itself" class — #28 was sighting #11 (a promise that
  placeholders are never invented, broken by its own empty-field fallbacks; the rebuild made the
  promise TRUE by refusing to compile with fields missing).
- **Serve-grade must include `src="<sidecar>"` = 0** — the decisive Phase C negative.
- **The sidecar FILE has no retirement mechanism** (`bugs_open/365`, filed today): dry-run
  `retract_asset_files` REFUSES non-/assets/ paths by design. Per tool: record the refusal, list
  the orphan in NOTES, move on. Cleanup is one batch when 365 ships. The shared
  `/tools/assets/webdesign-couk-header.js` goes with the LAST ported page only.
- **Post-retire tombstone re-read** at the end of the attendance window (360 is live until the roll).
- ⚠ ATTENDANCE unchanged: foreground poll loop, file only what you can attend.

## Phase C remainder (12), smallest first by the scope query

image-optimizer (8,281) · bayesian-rank (8,749) · community-growth (8,771) · head-architect (9,212 —
its ported slot took a literal_markdown edit 08-21, bytes differ from any older recorded md5) ·
seo-schema (9,495) · recommender-engine (9,513) · performance-budget (9,662) · smart-contrast
(11,104) · csp-builder (13,190) · fluid-typography (15,218) · vibe-equalizer (6,403, 2 ext scripts) ·
micro-cms (15,175, 4 ext scripts — Phase B-adjacent, consider owner sight).

## Owed / open

- **Verdict read owed: corr `a367b63e`** (558 post-hoc). If REVISE/REJECTED: the migration is
  APPLIED — act on the objections, do not pretend it can be held back.
- 360's fix rides the next chassis roll; after the roll, prove it (dispatch a section_edit at a
  removed row on a test page → expect `{skipped:true, tombstoned:true}` and the row still removed).
- `bugs_open/365` routed at the DGH-010/asset-retraction owner; orphan list so far:
  `/tools/blueprint-compiler/script.js`.
- The 098 report will show `6e8cef52c` (558's commit) as un-reviewed until `a367b63e` resolves —
  the submission commit `e5b…` carries the trailer; nothing further owed if approved.
- Sibling session [ac1f33] took `bugs_open/362` (link repair in the two tool writers) — committed,
  their verdict pending. Coordination notes in NOTES 12:35Z and their message trail.
