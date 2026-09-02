# 415 — the fire-gate `pre_query` spells eligibility differently from the selector, so an approved-only (or non-build) backlog never fires the build-pipeline trigger at all

**Filed 2026-08-26 by the bugs_open/413 fix session, on declared first-hand verification per
the 2026-07-31 owner ruling** (the substitute stated plainly: both texts below were read from
the live artefacts this evening — `scheduled_tasks.pre_query` and the `agent_definitions`
selector — and the divergence is a textual fact; the consequence is direct logic from the
gate's HAVING clause, not an inferred mechanism. No 090 run: there is no live damage to
diagnose today, only a door that closes under a reachable future shape.) Found while fixing
`bugs_open/413`; deliberately scoped OUT of migration 657 (agreed with the
dispatch_throughput lane 2026-08-26 ~20:4xZ) because it is a different clause-set on a
different artefact with a different failure shape.

## The drift (both texts read live 2026-08-26 ~20:1xZ)

There are now THREE spellings of "dispatchable work exists", in series:

1. **Fire gate** — `scheduled_tasks.pre_query` on `build-pipeline-trigger` (and the disabled
   `-2` sibling): counts sites having a row with
   `status = 'triaged' AND pipeline = 'build' AND attempt_count < max_attempts AND retry_after
   pass` under `s.locked_at IS NULL`, `HAVING COUNT(*) > 0` — no fire when zero.
2. **Selector** — `find_dispatchable_site`: admits `status IN ('triaged','approved')`, has
   **no pipeline filter**, plus approval_mode / depends_on / lock-exception / busy-skip
   clauses (post-657: also windowed by the loader's ordering).
3. **Loader** — `LoadWorkItemsAction` as configured by `build-dispatch-loop > load_items`:
   same admission as the selector, **no pipeline filter** (the config sets no
   `item_pipeline`), site-scoped.

The gate is NARROWER than what it gates, in two independent ways: `'approved'` is missing,
and `pipeline='build'` is present.

## Consequence (the door, stated precisely)

The trigger fires only while at least one site holds a `triaged`, `pipeline='build'` row.
A backlog consisting entirely of `status='approved'` rows (the approval_mode path — items a
human approved, the exact rows the estate most wants dispatched promptly) or entirely of
non-build `pipeline` values (dormant today per WDS-003, but the column exists and has
misfired before — a discovery check once emitted `pipeline='design'` and stalled) is
**dispatchable by the selector and loadable by the loader, and the trigger never fires to
ask.** No error, no row anywhere: the damage would be an absence, same meter-blindness class
as 413.

**Severity honestly stated: theoretical at today's volume.** `[MEASURED 2026-08-26 ~20:1xZ]`
the fleet held ~1,100+ `triaged` build rows across ~25 sites, so the gate passes
continuously; no starvation attributable to THIS gap has been observed. It is filed because
the closed door is reachable (a quiet fleet whose only remaining work is approved rows —
exactly the end-of-backlog state Phase 3 and the throughput work drive toward), it is cheap
to fix, and the drift class has already bitten twice on the same seam (078 → 285: selector
narrower than loader; 413 → 657: selector ordering vs loader ordering).

## Prior art (grepped both dirs before filing)

- `docs/agent_docs/sql_for_agents/213_dispatch_gate_matches_dispatcher.sql` — written
  2026-08-12 for exactly this alignment, **UNTRACKED and never applied** (dispatch_throughput
  HANDOFF traps list). It also predates the sibling row and uses a by-name UPDATE (the
  LANDMINES sibling-parity trap), and predates 633's lock-exception arm and 657's window —
  **do not apply it as-is**; treat it as evidence the gap was seen, not as the fix.
- `bugs_closed/048` — a no-op pre_query starving a concurrency group: the gate-lies-quiet
  failure shape, different clause.
- `bugs_open/136` — the `domain`→`pipeline` key rename residue on the same column.

## Fix candidates (ranked by what closes the door)

1. **Make the gate's predicate a strict widening of "the selector could return a row"** —
   simplest honest form: `EXISTS` with `status IN ('triaged','approved')` and the
   attempt/retry arms, `pipeline` filter REMOVED (the selector and loader have none), lock
   arm kept in the cross-site spelling. A gate may be WIDER than its selector (a spare fire
   is one cheap no-op tick); it must never be narrower. One migration on
   `scheduled_tasks.pre_query`; **must update BOTH rows** (the disabled sibling too —
   LANDMINES parity trap) or the rollback row diverges.
2. Drop the pre_query entirely and let the trigger fire every interval (the selector already
   answers "nothing to do" with zero rows). Costs one selector execution per tick always —
   measured ~1.3 s server-side per run at current data; the gate exists to avoid exactly
   that; owner call.
3. Documentation only — rejected: the gap is invisible at the caller and the damage is an
   absence.

## How to verify

Induce, don't wait: on a test site, hold one row at `status='approved'` as the fleet's only
eligible work (or count fires across a window where triaged-build drops to zero — the
end-of-backlog state), and watch `orchestration_states` for `build-pipeline-trigger` fires.
Disconfirming result for the fix: with only approved rows pending, the trigger fires and the
site is served within one interval.

## Relations

`bugs_open/413` (sibling seam, fixed by 657) · migration `285` (the selector↔loader
eligibility agreement this gate was left out of) · `LANDMINES.md` 2026-08-26 "ONE ordering
contract" entry (the general class) · WDS-002/WDS-003 in the work-dispatch register.

## Addendum 2026-09-02 — re-verified, a THIRD narrowness found, and a fixing-session handoff cut

Still valid `[MEASURED 2026-09-02]`: both trigger rows share `md5(pre_query) =
200246f7ede3e33b14be2fc064efa7da`, text byte-identical to the filing read. **Third
narrowness confirmed**: the gate's bare `s.locked_at IS NULL` lacks the selector's
lock-EXCEPTION arm, so a fully-locked site whose excepted item is dispatchable is
selector-admissible but gate-invisible — same class as the other two. 413/657 closed
2026-09-02 (measured PASS), leaving this file the seam's only open drift.

**COLD-START for the fixing session:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_415_fire_gate/HANDOFF_2026-09-02_continue_here.md`
— chosen fix (candidate 1, gate ⊇ selector), the preflight md5 anchor, all traps (both-rows
update; 584 VERIFY 1/7 pins PARITY not text, so no lockstep owed; the stale staged 213 file;
the regexp_replace landmine), sidecar + council expectations, and the induced verification.

## CLOSED 2026-09-02 — fixed AND live AND proven (fixing session, same day as the handoff)

**Fix: candidate 1, migration `688_fire_gate_admits_what_the_selector_admits.sql`
(+`_ROLLBACK`/`_VERIFY`), committed `59e722812`, APPLIED 2026-09-02 13:28Z.** Whole-value
replacement on BOTH trigger rows in one statement (ROW_COUNT=2 asserted; both rows read back
`md5 = 2ebd918b33b36d1b55014bbe60cc2dcb`): `status IN ('triaged','approved')`, pipeline
filter removed, lock-exception arm added in the selector's cross-site spelling
(`work_items_common.go:851-870`). approval_mode/depends_on/busy-skip deliberately stay
selector-side — wider is the safe direction. Chain invariant now: **gate ⊇ selector ⊇ loader.**

**Proven at the artefact, 2026-09-02:**
- VERIFY sidecar mutation-proved: run against the pre-fix text it FAILED on the `'approved'`
  assertion (exit 3); post-apply it passes on both rows.
- Each narrowness proven load-bearing by read-only simulation over live data (the bug's
  sanctioned "simulate" arm — live data held no divergent population: 102 eligible rows, all
  triaged/build, 0 lock-exception entries): approved-only, non-build-only, and
  lock-excepted-only populations each left the OLD gate silent (`f`) and fire the NEW (`t`).
  Disconfirmable: any case could have read f/f or t/t.
- Live path: trigger fired 13:28:30Z, six seconds after the apply; new gate measures ~1.3 ms
  server-side (old: ~7.3 ms) — candidate 2's ~1.3 s/tick selector cost avoided.
- 584 daily VERIFY stays green by construction (1/7 pins PARITY, not text) — confirmed by the
  dispatch_throughput lane, pinged at commit and apply; they will hold 688's apply time
  beside their next cadence read.

**Housekeeping:** prior-art `213_dispatch_gate_matches_dispatcher.sql` renamed
`_SUPERSEDED` (`5cd756b99`) — it had become TRACKED (passenger in `0f3721c6e`), making it a
live whole-directory `--apply` hazard; the suffix puts it under the runner's SIDECAR_RE.
Register WDS-002 (new fire-gate bullet) + WDS-003 updated; LANDMINES corrected in three
places and re-synced. Council: `Council-Submitted: 5f0cb450-e40f-4ffd-ac8e-01534caeac25` —
**round 1 REVISE** (gating HIGH from prior_art_librarian: it quoted the `sites.locked_at`
landmine's HEADLINE, whose own first line is the 2026-08-03 "headline is HISTORY" correction —
answered with the correction text, a fresh live selector read, and round 1's own
`selector_has_lock_exception_arm: true` check; two MEDIUMs on the 213 rename being "deferred"
— in fact committed 6 s BEFORE the DB apply, evidenced with `git ls-files` + timestamps),
**round 2 APPROVED 2026-09-02 13:39Z, all reviewers, no SQL change in either round.**
098 credits the commit automatically from the Council-Submitted trailer (forward-only, no
amend). The REVISE round cost ~12 minutes and found no defect — but it did force the
before-the-apply timeline onto the record with checks attached, which round 1's prose had
left as stated intention.
