# 461 — a live driver holds nothing, so the 5-minute stuck clock can judge a correctly-working orchestration dead, and a taker resumes its step alongside it

**Filed 2026-09-03 by the `bugfix_329_takeover_claim` lane**, as the named residual of
`bugs_open/329`'s fix (`b55f837ef`, register **WFA-025**). Filed rather than left in a risk block
because the `bug_historian` council seat objected at **medium** on exactly this and was right: an
incomplete fix of a documented recurring class should be owned, not noted.

**Status: OPEN, UNOWNED.** No known incident (see §4 — the same census that bounds 329 bounds this).

## 1. What it is, in plain terms

When a pod is driving an orchestration through a step, it holds **nothing** that says so. It stamps
`last_activity` when it moves *between* steps, and then goes quiet for as long as the step takes.
Another pod that sees a row untouched for five minutes is entitled to conclude the driver is dead
and resume it. The driver is not dead — it is working.

`bugs_open/329` closed **taker versus taker**: two would-be rescuers can no longer both resume one
row, because the staleness test now happens inside the version CAS that claims it. It could not
close **driver versus taker**, and no claim taken by the takeover side can: exclusion requires the
driver to hold something, and it holds nothing.

## 2. The mechanism, read not inferred

- `StuckOrchestrationTimeout = 5 * time.Minute` — `platform/orchestration/coordinator.go:38`.
- `defaultLocalActionTimeout = 7200 * time.Second` — `coordinator.go:1246`. **A step is permitted to
  run for two hours.**
- **Nothing refreshes `last_activity` during a local action.** Every refresh site in `coordinator.go`
  (`:1039`, `:2220`, `:2850`, `:2914`, `:3441`, `:3975`) is a step transition, and
  `UpdateStateWithVersion` stamps it (`state.go:1051`) only when something writes the row.

So the legitimate-silence ceiling is **7200 s** against a **300 s** "it must be dead" threshold — a
**24×** gap. Any step slower than five minutes is, by the clock, indistinguishable from a dead pod.
A nine-minute council seat is the everyday example.

⚠ **Widening the threshold is not the fix**, and this is why: to be *safe* it would have to exceed
7200 s, at which point recovery of genuinely abandoned rows takes two hours and is worthless. The
threshold cannot be both. That is the argument for a heartbeat rather than a bigger number.

## 3. Why this is worth a bug and not a footnote — the seat's argument, quoted

> *"This platform has a repeated documented shape of exactly this class — two actors mutating the
> same target with no mutual exclusion producing corrupted/overwritten/duplicated output (e.g.
> `bugs_closed/038` 'replan rebuilds every deployed page and regenerates its content',
> `bugs_closed/040` 'failed page build leaves page deployed and partially composed', `016b` §9 'a
> claim taken before a guard that can bail parks the row in a status nobody owns'). If the resumed
> orchestration is a content-writing workflow (page-rerender, section save), a genuine driver
> mid-write racing a taker's fresh resume is a plausible silent-content-loss vector this fix does
> not close, only shrinks."* — `prior_art`/`bug_historian` seat, council corr `3beb3f54`, round 1.

**Post-329 the bound is 2** concurrent actors on one orchestration (driver + exactly one taker),
down from unbounded. **2 is not 1**, and for a content-writing workflow the two are not equivalent.

## 4. What is measured, and what is not

`[MEASURED 2026-09-03 ~11:4xZ]`, all bounding 329 and this equally:

- Double-handle census, 24 h: **0** overlapping pairs across **3,044** handler orchestrations /
  **2,911** work items (the `dispatch_throughput` lane's meter — their `RUNBOOK_dispatch_throughput.md`
  §"Concurrency meters that actually measure concurrency").
- **0** rows with a repeated step name in one `execution_path` over 7 d.
- **0** takeover log lines in either chassis pod's reachable window.

⚠ **All three are bounded by guards in SERIES that do not cover every path** — the chassis intake
serialisation claim (`agent-chassis` only; `intake.go` refuses the mode when `a.spawned`) and
per-path CASes such as the work-item claim. On a path with external side effects and **no** claim
they say nothing. **[UNVERIFIED]** that a driver and a taker have ever in fact been inside one
orchestration together.

**A better instrument now exists and did not before.** 329's fix logs `STALE_TAKEOVER_CLAIMED` and
writes `processing_history` action `stale_takeover_claimed`. Once it rolls:

```sql
SELECT count(*) FROM orchestration_states
WHERE processing_history @> '[{"action":"stale_takeover_claimed"}]';
```
Every row that returns is a takeover that **happened**. Cross-check each against whether the previous
driver was genuinely dead — if the step later completes from the original driver too, that is this
bug, observed. **Do not size this bug before that query has data**; today it necessarily returns 0
because the fix is not yet rolled.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **A driver heartbeat.** The driver refreshes `last_activity` (or a dedicated lease column) on a
   timer while a step runs, so silence means death rather than slowness. Makes the bad state
   unrepresentable because the predicate finally measures what it claims to. ⚠ **This is a shared
   seam on `orchestration_states` and is architecture-scope** — a *third* guarded mechanism on a
   table whose council invariant (`state_locks_test.go`, corr `4a227ed9`) says there are exactly
   two, inherited by everything `platform/agentbase` builds. It needs an RFC, not a bug patch.
   ⚠ Prior art to read first: the intake pool already solved this shape one layer up —
   `intake_workers.go` heartbeats every `lease/3` and abandons the key on `INTAKE_CLAIM_LOST`. Its
   own limitation is instructive: it checks only **between** events because `processMessage` takes
   no context, so it cannot abort work in flight. A coordinator heartbeat inherits that limit.
2. **Give the takeover a second, independent liveness signal** — e.g. require the row to be stale
   AND the stamped pod to be absent from the live pod set. Cheaper, but ⚠ `processing_node` records
   the pod that **created** the row, not the one driving it (`bugs_open/075`, and `SetExecutingStep`'s
   assignment is documented-inert), so this needs a *new* correctly-maintained field and is candidate
   1 wearing a smaller hat.
3. **Shrink the ceiling instead of raising the floor:** bound local actions well under 5 minutes so
   the clock stops lying. Touches `bugs_open/169` (local-action deadline) and every long step —
   a scheduling change, not a concurrency one, and it does not close the class.
4. ~~Widen `StuckOrchestrationTimeout`.~~ **Rejected** — see §2. No value is both safe and useful.

## 6. Notes for whoever takes it

- **Read `bugs_open/329` first, including its two CORRECTIONS**: its filed mechanism was false
  (`UpdateState` **is** `UpdateStateWithVersion`), and the obvious concurrency test is **inverted**
  here — simultaneous actors are absorbed by the version CAS, so a `sync.WaitGroup` start-line test
  shows broken code passing. Both are in `LANDMINES.md`, footprinted.
- **Count the guards in series before believing any green.** Three of them, and only the middle one
  is this family's business.
- Consumers to tell, not merely measure: the `chassis_replica_scaling` lane (their cross-worker
  safety inventory), and the `dispatch_throughput` lane (whose census is the field meter on the path
  they own).
