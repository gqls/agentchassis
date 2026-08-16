# HANDOFF — 2026-08-16, fresh chat starts here: 277 mechanism live and drained; 083 promoter live and drained; two council rounds owed; RFC_030 lane created, not started

**Written by the "bugfix 033" session at token-load hand-off, the morning after a fresh chassis
(`v1.0.1303`) rolled.** Everything below was measured 2026-08-16 ~09:45Z unless dated otherwise.
Read this file FROM DISK; then `PLAN_2026-08-15_required_fields_router.md`, then
`NOTES_required_fields_repair.md` from the bottom, then `RUNBOOK_required_fields_repair.md`.

## 1. What is LIVE (verify, do not trust)

| thing | state | how to re-verify |
|---|---|---|
| `required-fields-missing-handler` (seed 410 **v3**, CQ-023) | live, ledger-recorded; 8 routes; single active row | `SELECT count(*) FROM agent_definitions WHERE type='required-fields-missing-handler' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL` → 1 |
| producer `check_required_fields_missing.go` | born-`detected` at the router (revert `3c6354059`) — **LIVE on v1.0.1303** | binary probe per RUNBOOK; `git merge-base --is-ancestor 3c6354059 <stamp>` |
| `required_fields_missing` backlog | drained: 36 parked WITH route+facts (35 blob, 1 gas converter), rest closed; **0 triaged / 0 blocked / 0 unrouted** | RUNBOOK "fleet after-state" query |
| `detected-item-promoter` (seed 430, SCH-026) | live, 900s, ≤20/tick, known-good pairs; **pile 70 → 4** (held pair only) | `SELECT count(*) FROM site_work_items WHERE status='detected'`; `SELECT last_triggered_at FROM scheduled_tasks WHERE name='detected-item-promoter'` |
| RFC_030 (router engine) | **RULED + SCHEDULED**, lane created at `docs024_key_docs_latest/router_engine/`, **nothing built** | read its `HANDOFF_2026-08-15_continue_here.md` |

## 2. Owed work, in priority order

1. **Council round 2 for the promoter (corr `05a3d1c8`, REVISE).** Every objection is already
   MEASURED and written into `bugs_open/083`'s 2026-08-16 block — resubmit with those
   measurements in `grounded_in` (`RESUBMIT_CORR=05a3d1c8-…`, file
   `submission_083_promoter.json` here — add a `_round2`). Include the door-closer as an edit
   (see 3). Do NOT re-litigate the reuse_agent "invoke the Go action" point — state why the
   SQL mirror is the SCH-006 shape (the Go action is site-scoped + workflow-embedded).
2. **Council trail `7b0e2833` (the router) ended REVISE ×4 with seats disagreeing; the owner has
   since RULED both open questions** (083 cand-2 built; RFC_030 scheduled). A round 5 should be a
   short one: cite the two rulings, the live outcomes, and stop. If it REVISEs again on the
   same policy points, record and leave — the trail has done its job.
3. **Door-closer migration for the promoter** (guardian/editquality objection, cheap, real):
   add `AND wi.pipeline NOT IN ('diagnose','report')` to the candidates CTE. **New numbered
   file** (430 is ledger-recorded — never edit a recorded file); re-`ls` for the next free
   number (collisions at 408/409/429 this week). Verify block: same partition assert.
4. **083 verify criteria 2 + 3**: `phantom_internal_link` first-ever `complete` (waits for a
   re-raise); sample 3–5 of the 93 promoted-and-completed rows and verify at the LIVE PAGE.
   Then 083 can move to `bugs_closed/` (both paths on the commit — LANDMINE).
5. **Bug 277 close-out**: fixed AND live holds now. Remaining before moving to `bugs_closed/`:
   the +7-day churn guard (~0 new `unresolved` of the type after 2026-08-15 ~14:50Z) and the
   re-raise-then-park of the two cancelled conversions (`4fa5b019`'s and `7ed472ab`'s findings)
   via discovery rotation — if not re-raised by ~08-22, re-file by hand.
6. **The 4 held `page_component_status_drift` rows**: someone should canary ONE by hand
   (statement in seed 430's header) — that makes the pair known-good and the promoter takes the
   rest. Not this lane's type; tell `component-template-fixer`'s owner or do it with care.
7. **Two sibling born-triaged producers** (`check_integrity.go`, `check_tool_acceptance_due.go`)
   — their lanes' call whether to return to `detected` now the promoter exists; both pairs are
   known-good. Mention, don't do.
8. **Start the `router_engine` lane** (RFC_030) — its own cold-start handoff exists.

## 3. Landmines this lane hit (all in LANDMINES.md or WRONG_CALLS.md — grep before repeating)
- The loop's `mark_complete` REPLACES `result` on completed rows → close-evidence lives in
  `orchestration_states` (`workflow_plan->>'start_step'='classify'`), park-evidence on the row.
- `landmines-sync.py --apply` alone consumes the new-entry status; run
  `landmines-verify-dispatch.sh` (CLAUDE.md corrected 08-15).
- A code-comment figure is a dated snapshot ("exactly one ever terminal" was 07-25's; live = 50).
- The scheduler GATE requires `pipeline='build'`; the loop's loader does not filter — a
  content-only site never wakes the loop.
- The migration runner's dry run lists OTHER sessions' pending files — never `--apply`; use
  `--record-only <file> --note` for a hand-applied file.
- `handler_coverage_test`'s const resolver only sees `const ( … )` block form.

## 4. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py 277`, `083`,
and grep live `.jsonl` for `router_engine|RFC_030|05a3d1c8` · re-measure §1 · then §2 item 1.
