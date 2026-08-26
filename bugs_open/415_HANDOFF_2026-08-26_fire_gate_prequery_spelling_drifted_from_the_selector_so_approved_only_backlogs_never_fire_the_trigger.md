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
