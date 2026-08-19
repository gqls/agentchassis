# 329 — both orchestration takeover arms decide by a CLOCK, not a lock, so two pods can run the same step

**Filed 2026-08-19** by the `orchestration_status_lifecycle` lane (`bugs_closed/294` → `310` → the
class fix). Status: **OPEN, UNOWNED.**

**Raised by the council, not by a symptom.** The `guardian` seat objected at *medium* on migration
`465`'s round (`1c212b15`, APPROVED) and was right. Filing it because an objection that lives only
inside a council verdict is effectively invisible — nobody greps `diagnosis_artifacts`.

**No known incident.** This heuristic has run for months and predates the change that surfaced it.
It is a correctness gap, not a fire.

---

## The one-paragraph version

When a pod sees an orchestration that looks stuck, it takes it over and resumes it. It decides
"stuck" by a **clock** — `time.Since(state.LastActivity) > StuckOrchestrationTimeout` (5 minutes) —
and then simply proceeds. Nothing claims the row. If the original pod is in fact still working and
merely has not refreshed `last_activity`, **two pods run the same step**, with whatever external
side effects that step has, twice. The estate already has the right primitive —
`TakeOverOrchestration` is a guarded compare-and-swap built for exactly this shape — and neither arm
uses it.

## The two arms

| arm | location | gate |
|---|---|---|
| `EXECUTING_STEP` takeover (pre-existing) | `coordinator.go` `case StatusExecutingStep` | `state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout` |
| `RUNNING` takeover (added by `465`) | `coordinator.go` `case StatusRunning` | `time.Since(state.LastActivity) > StuckOrchestrationTimeout` |

**The second deliberately mirrors the first.** `465` reused the estate's existing takeover guard
rather than inventing a weaker one, so this bug is **not a regression** — it is the pre-existing
shape, now present twice and therefore worth naming.

## Why there is no lock today `[MEASURED 2026-08-19]`

- `SetExecutingStep` (`state.go:1346`) ends in `r.UpdateState(ctx, state)` — **not**
  `UpdateStateWithVersion`. So the write that marks a step executing does **no** version check.
- `ClearExecutingStep` (`state.go:~1428`), the pre-existing takeover's own write, likewise ends in
  `r.UpdateState(ctx, state)`.
- `UpdateStateWithVersion` **exists** (`state.go:961`) and is used elsewhere, so optimistic locking
  is available and simply not on this path.
- `TakeOverOrchestration` (`state.go:1366+`) is a **guarded CAS**: it re-stamps `processing_node`
  from a named previous holder and returns whether the handover was won. Its own doc comment
  describes it as the fix for `bugs_open/075`.

⚠ **Do not build an ownership gate on `processing_node` naively.** `SetExecutingStep` assigns it but
`UpdateStateWithVersion`'s UPDATE does not list the column, so the assignment **never reaches the
database** (comment at `state.go:1354-1360`, `bugs_open/075`). The column therefore records the pod
that *created* the row, not the pod driving it. `TakeOverOrchestration` is the only writer that
moves it after creation — which is precisely why it, and not a hand-rolled check, is the primitive
to use.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Route both arms through `TakeOverOrchestration`.** A takeover becomes *claimed* rather than
   *assumed*: win the CAS, then resume; lose it, and return without acting. Uses the primitive that
   already exists and closes the door for both arms symmetrically.
2. **Version the write instead** — have `SetExecutingStep`/`ClearExecutingStep` use
   `UpdateStateWithVersion`, so a second pod's takeover fails the optimistic-lock check. Smaller
   diff, but it converts a race into a retry rather than into a refusal, and the existing
   `ExecuteWithOptimisticLocking` retry loop would then re-drive it.
3. **Leave the heuristic and widen the threshold.** Rejected as a fix — it changes the odds, not the
   representability, and the estate's own rule is to rank by what closes the door.

## How to verify a fix

- **Induce the race:** two concurrent callers against one orchestration whose `last_activity` is
  older than `StuckOrchestrationTimeout`. Exactly one must proceed; the other must return without
  executing the step. **The disconfirming result is that both proceed** — which is what today's code
  does, so run it against the unfixed path first and watch it fail.
- **Negative control in the same run:** a fresh row (`last_activity = NOW()`) must be taken over by
  *neither*, or a fix that refuses everything passes the test above.
- **Do not verify by reading the function** — this lane got the same class of thing wrong three
  times in two days by reading a body instead of its callers. Walk the caller graph, and confirm at
  the artefact.

## Notes for whoever takes it

- **Core dispatch plumbing.** The council's own history shows `coordinator.go` deflected upward six
  times in seven days; expect the architecture seat to look hard at anything beyond a point fix.
- **Needs a roll**, unlike this lane's other work — it is Go, inert until the next build.
- Related: `bugs_open/075` (why `processing_node` is inert), `bugs_closed/294`, migration `465`,
  council `1c212b15` for the objection verbatim, and
  `docs024_key_docs_latest/orchestration_status_lifecycle/` for the lane's runbook and notes.
