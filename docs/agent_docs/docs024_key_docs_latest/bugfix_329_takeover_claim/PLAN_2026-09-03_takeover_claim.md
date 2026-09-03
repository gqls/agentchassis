# PLAN — bugfix 329: a takeover is CLAIMED, never ASSUMED

Design, phasing, and the reasons. Corrections live here, marked, never edited away.

## What the thing IS, before any rule

`handleOrchestrationStatus` is the coordinator's front door: a message arrives for an
orchestration, and this function decides what to do with it based on the row's status. Two of its
branches exist to rescue work whose driver has died. Each decides "the driver is dead" by looking
at how long ago the row was last touched.

## The rule the fix asserts

**A takeover is claimed, never assumed.** A caller may resume a row only if the row *still* meets
the stale predicate at the moment of the write that takes it — not merely at the moment the caller
looked.

## How the case measured against it, before the fix

It did not. Both arms tested `time.Since(state.LastActivity) > StuckOrchestrationTimeout` against
the **caller's snapshot** and then proceeded. Nothing claimed the row.

## The mechanism, CORRECTED — and this is the whole design decision

> **CORRECTION 2026-09-03, to `bugs_open/329`'s own filed mechanism.** The bug file says
> `[MEASURED 2026-08-19]` that these writes are unversioned — "ends in `r.UpdateState(...)`, **not**
> `UpdateStateWithVersion`". **False.** `state.go:883-885` is a one-line delegating wrapper:
> `UpdateState` **is** `UpdateStateWithVersion`. Its fix candidate (2) is a **no-op**.

So the defect is **not** a missing guard on the write. It is a **check-then-act across two reads**:

1. the arm judges staleness on the **caller's snapshot**;
2. the write that follows does its **own fresh** `GetState` → mutate → version-CAS;
3. and step 2 **never re-tests the predicate from step 1**.

Two takers seconds apart therefore both win, because each CASes against the version it has just
read. The version CAS was always present — it was answering a different question.

⚠ **The corollary inverts the obvious test, and this is the single most useful thing in this file.**
Exactly **simultaneous** takers never double-executed: the loser's CAS fails and the arm returns the
error. So a `sync.WaitGroup` start-line test shows the **broken code passing**. The disconfirming
case is the **sequential** interleaving.

## The design

Move the predicate **inside** the version-guarded write. `ClaimStaleOrchestration` re-judges the
**fresh** row within `ExecuteWithOptimisticLocking` and writes only if it is still stale. The write
**is** the claim: `UpdateStateWithVersion` stamps `last_activity = now` and `version + 1`
unconditionally, so every later taker's fresh read declines with `ErrTakeoverLost`. A lost version
race is re-read and **re-judged**, never retried blind.

Built from existing machinery only — no new SQL, no new column, `processing_node` untouched.

## Why the alternatives lose, ranked by what becomes unrepresentable

| candidate | verdict |
|---|---|
| **Chosen: predicate inside the version CAS** | Makes "N takers resume one stale row" unrepresentable. Only one write moves V→V+1, and it refreshes `last_activity`, so every other taker's fresh read fails the predicate. |
| Route both arms through `TakeOverOrchestration` (the bug file's candidate 1) | **Rejected.** Its CAS is `WHERE processing_node = $3` from the **observed** value and deliberately leaves `version` and `last_activity` alone ("bookkeeping, not a state transition", `state.go:1378-1380`). Where a row already carries the acting pod's own name, two callers in that pod both match and both report `rowsAffected = 1` — **no exclusion at all**. And sequential takers both win regardless: taker 2 observes taker 1's name, CASes it to itself, and the row is still stale by clock. |
| Version the write (candidate 2) | **A no-op — already the case.** See the correction above. |
| A lease column or advisory lock | The only option that also closes driver-versus-taker, and **deliberately not built.** It is a *third* guarded mechanism on `orchestration_states`, whose council invariant says there are exactly two; a session lock held across a 9-minute step pins a pool connection; and it duplicates one layer down what `chassis_orchestration_claims` does one layer up. **The measurements do not support carrying that seam** (see sizing). |
| Widen the threshold (candidate 3) | Changes the odds, not the representability — and the legitimate-silence ceiling is 7200 s, so no threshold under two hours is safe and one at two hours makes recovery worthless. |

## What the fix deliberately does NOT close

**Driver versus taker.** `defaultLocalActionTimeout = 7200 s` (`coordinator.go:1246`) and **nothing
refreshes `last_activity` during a local action** — every refresh site is a step transition. So a
live driver inside a nine-minute council seat is "stuck" by the five-minute clock while behaving
correctly. **A driver holds nothing, and no claim taken by the takeover side can exclude it.**

Post-fix bound per orchestration: **2** (driver + exactly one taker), down from unbounded. Closing
the rest needs a driver heartbeat — a separate seam, and one this fix should not smuggle in.

## Sizing, stated honestly because it argues for a SMALLER change

This is a correctness gap, not a fire, and the plan is sized as one.

- Double-handle census, 24 h `[MEASURED 2026-09-03 ~11:4xZ]`: **0 overlapping pairs** across 3,044
  handlers / 2,911 items.
- **0** rows with a repeated step name in one `execution_path` over 7 d.
- **0** "Found stuck orchestration" / "stalled between steps" lines in either chassis pod's
  reachable log window.

**What justifies fixing it anyway is what it REMOVES, not what it prevents today.** Three guards sit
in series — (i) the chassis intake serialisation claim, `agent-chassis` only; (ii) **these arms**;
(iii) per-path CASes such as the work-item claim. Guards (i) and (iii) absorb the outcome where they
apply, but that absorption is a **by-product**: (iii) exists to make work-item claiming exclusive,
and (i) is disabled structurally for spawned pods — which are the **majority** of orchestration
drivers. A refactor of (iii) could remove protection it never knew it was providing.

## Governance

- **Not architecture-scope** by the 2026-07-29 §1 test: the shared mechanism's guarantee is
  **narrowed** (an unconditional resume becomes a claimed one), not widened. ⚠ The commit hook's
  architecture signal fires anyway (exported symbol removed + ossified core site) — expected, and
  the council is the arbiter, not me.
- **No opt-in field** (2026-08-02 §2): the fix *removes* authority from a branch; no caller-supplied
  predicate licenses a widest branch; no action input spec changes, so RFC_022's N = 10 budget is
  untouched. A "revert to clock-only" switch would be a mechanism rotting unexercised, which
  2026-07-29 §2 declines to require.
- **Council:** `Council-Submitted: 3beb3f54-6d51-42fd-969f-78e4ea871659`, committed pre-verdict per
  the 2026-07-30 rule. **Still owed: read the verdict and act on a REVISE.**

## Verification, and it was watched to fail first

```
scripts/verify-head-builds.sh --with platform/orchestration/stale_takeover_claim_test.go --test ./platform/orchestration/
```
At HEAD `f9aa97a97` (unfixed) **all four fail**, and the failure messages name the write that must
not have happened — the EXECUTING_STEP arm caught issuing `status=RUNNING version=10` against a row
another actor had claimed at version 9, the RUNNING arm caught issuing `status=EXECUTING_STEP` on
the same. `TestSimultaneousTakersOnlyOneClaims` fails with the optimistic-lock error, which is the
inverted-test corollary demonstrated rather than asserted. With the fix, the package passes.

**Negative control in the same run:** `TestFreshRowsAreNeverTakenOver` — a fresh row must produce
**no database traffic at all** on either arm. It passes at HEAD too, which is correct: it does not
discriminate fixed from unfixed, it exists so that a fix which refuses *everything* cannot pass the
other four.

**Isolation from the guards in series:** the tests call `handleOrchestrationStatus` directly — below
the intake claim, above any action-level CAS — and sqlmock fails on any unexpected statement, so the
expectation list **is** the span's complete DB traffic. A test on the dispatch path would pass with
the fix reverted; that is why none of these lives there.

## Still open

1. **Post-roll verification at the artefact.** Go is inert until a roll. Needles:
   `STALE_TAKEOVER_CLAIMED` / `STALE_TAKEOVER_LOST` in the logs, and
   `processing_history @> '[{"action":"stale_takeover_claimed"}]'` in the DB — the first durable
   record this path has ever had. An induced fault must **bypass the intake claim** or it proves
   nothing; that needs the owner's sign-off on the vehicle.
2. **Should the `INITIALIZED` arm get the same treatment?** It has the identical check-then-act
   shape. Out of 329's scope; size it before proposing.
3. **`180 s` intake lease against a `300 s` stuck timeout** — recorded as a 329 finding
   (`bugs_open/329` §5). The `dispatch_throughput` lane was offered it and declined on the ground
   that they have never measured it. **[UNVERIFIED]** whether the ordering was chosen or defaulted.
