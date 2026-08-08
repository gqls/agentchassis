# HANDOFF — 117 chrome staleness reference — cold start for a fresh chat

**Written** 2026-08-08 morning; **REWRITTEN 2026-08-08 afternoon — the earlier
version of this file said "RESEARCH COMPLETE, NO CODE CHANGED, NO PLAN YET" and
that is no longer true.** State now: **FIX BUILT + COMMITTED (`998bf4c9f`) +
MIGRATION 334 APPLIED + COUNCIL RUNNING. NOT YET LIVE — waiting on the owner's
next whole-fleet release.**

## Read these first, in this order

1. `bugs_open/117_HANDOFF_2026-07-27_...md` — the bug, the 2026-08-07/08
   measurement contribution, and the 2026-08-08 "fix is BUILT" contribution
2. `PLAN_2026-08-07_chrome_staleness_reference.md` — D1–D8; D6 (per-slot render
   inputs) and D7 (the mechanism) are the design
3. `NOTES_chrome_staleness_reference.md` — measurements + missteps, append-only
4. `RUNBOOK_chrome_staleness_reference.md` — **R10 is the deploy-verification
   recipe the next session must run**
5. Register entry **IMP-052** (`docs026_concept_register/register/improvement-loop.md`)

## What shipped (commit `998bf4c9f`, all tests green, guards mutation-proven)

One shared SQL expression — `datahelpers.ChromeRenderInputsSQL`
(`platform/orchestration/datahelpers/chrome_render_inputs.go`) — digests every
store chrome renders from. `render_site_components` stamps it into the new
`site_components.render_inputs` column in the same lock-guarded UPDATE that
stores `rendered_html`; the `stale_site_components` check recomputes the same
string and fires the existing `needs_rerender` → `rerender-pages` pipe on
`IS DISTINCT FROM`, one site-level item (`stale_chrome`), locked rows skipped.
`fix_harcoded_colours_action.go` now stamps `updated_at`. Four source-pinning
tests in `chrome_render_inputs_contract_test.go`.

## Live state right now

- **Migration 334 APPLIED + recorded** (`schema_migrations`,
  applied_by `bugfix_117_lane_hand_applied`, 2026-08-08 14:49Z). The column is
  live and inert — the running image (v1.0.1258 fleet) neither reads nor
  writes it. The ordering constraint (column before code) is therefore already
  satisfied for ANY future roll.
- **Council: APPROVED round 1** (14:57Z, correlation `f62e20ae-…`), 6 advisory
  objections, none high-severity — triaged with evidence in NOTES (the
  "content_hash already exists" objection from two seats is refuted by
  measurement: the column is populated on 0 of 1,884 live rows and hashes
  output, not inputs). The `998bf4c9f` commit carries `Council-Submitted:`;
  098 credits it automatically now the verdict is approved — no amend.
- **NOT deployed, deliberately.** Deploys are owner-run and whole-fleet
  (`make release redeploy-agents ENVIRONMENT=production REGION=uk001` — owner
  feedback 2026-08-03; a single-service roll at its own tag fragments the
  fleet). Another session has IMAGE_TAG at v1.0.1265 uncommitted in the tree —
  do not touch the makefile. The commit rides the next release.

## What the next session does, in order

1. ~~Read the council verdict~~ — DONE, APPROVED r1 (see above).
2. After the owner's next fleet release: run RUNBOOK **R10** — pod-grep
   positive `render_inputs` / negative `stale_sc_` on every replica, same exec.
3. Watch the one-time baseline wave: ~19 `stale_chrome` items (one per site,
   `item_key='stale_chrome'`), each rebuilds + restamps, then QUIET. Loud
   forever or silent from the start are the two opposite failure modes — check
   both directions. oufe.com/footer is the row that proves the fix (it was the
   false negative).
4. Then: bug 117 meets the fixed-AND-live bar → record the close in the bug
   file (it STAYS in `bugs_open/` per owner ruling 2026-08-06), update IMP-052's
   status line, and update this handoff.

## Rollout expectation — so nobody files the wave as a bug

Every site fires ONE `stale_chrome` item on its first post-roll discovery pass:
no stamp = stale, by design (backfilling stamps would have declared oufe's
known-stale footer fresh). Bounded, one wave, replaces 33 CONTINUOUS false
positives. The 3 `component_id IS NULL` rows (loanandmortgagecalculator.co.uk)
also fire and get healed by the rebuild's fallback-assign path.

## Constraints already honoured (do not re-litigate without new evidence)

- Work-item dedup: single producer, site-level key, stated in IMP-052
  (2026-08-02 ruling §1 satisfied).
- Handler satisfiability: `rerender-pages`' `check_refresh_components` step
  force-rerenders all three slots on `refresh_site_components: true` — read
  live 2026-08-08.
- Column not a `content_data` key: bug 190's guard documents "no automated
  content_data writer" on site_components as structural.
- The fingerprint correlates on an `sc` alias — un-aliased embedding silently
  compares each site with itself (test-pinned).
- Locked slots (6/57) unmonitored by this check; 069 owns that surface.

## Ownership

This lane owns 117 (`who-owns` + live-transcript symbol grep, re-verified
2026-08-08). Re-run RUNBOOK R8 before resuming — both checks lag.
