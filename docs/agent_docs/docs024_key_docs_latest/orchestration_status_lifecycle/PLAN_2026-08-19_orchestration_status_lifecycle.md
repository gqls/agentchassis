# PLAN — orchestration status lifecycle (2026-08-19)

> **Created late, and that is itself a correction.** CLAUDE.md says the standing five exist from
> the START. This lane began on 2026-08-17 as a single bug (`294`) and grew across three days into
> a class fix touching two scheduled tasks, seven Go files and four bug files before anyone wrote a
> plan. The record until now lived in commit messages and bug files. Written up here retrospectively
> and dated honestly rather than backdated.

## What we are trying to do

Make it structurally impossible for an orchestration to get stuck in a state that nothing ever
looks at again.

## How the problem was actually shaped (each layer found by fixing the one above)

1. **`294`** — a row stuck in `RUNNING` was unreachable by every recovery path, and pinned two
   Kafka topics for ever. Three independent locks: no reaper arm, `TimeoutMonitor` unreachable, and
   `handleOrchestrationStatus` had no `case` so a message **errored** rather than recovering it.
2. **`310`** — the identical gap one status over (`INITIALIZED`), found by a parallel lane while
   verifying 294's close. Both lanes produced **byte-identical** SQL independently.
3. **The class** — the reaper's coverage was a *list*, so any status nobody listed was never swept.
   And `database-cleanup` carried a **second** enumeration with the same blind spot.
4. **The vocabulary** — even after the invariant, "which statuses are terminal" was still literals
   in two `pre_query` texts. Adding a status still meant finding every copy by hand: the same shape,
   one level up.

## Decisions and their reasons

| # | decision | why |
|---|---|---|
| 1 | Reaper arm, not a Go fix, for `294` | if the pod dies mid-transition, no in-process recovery can exist — the process that would do it is gone. An external sweeper is the *only* thing that can reach the state |
| 2 | `RUNNING` licensed by CODE, not the census | the census the bug file mandated had **expired** — after the sweep it read 0 in every band, so it could not come out otherwise. One writer / one caller / next-write-flips-it-back does not expire |
| 3 | `INITIALIZED` licensed by MEASUREMENT, not code | **decision 2 does not transfer.** `INITIALIZED` *waits* on a queue, so its duration is load-dependent. Measured instead: 5,736 rows, max 6.31 s, zero over 5 min. The parallel lane reused the structural argument here and was ~1000× out |
| 4 | Invariant, not a third arm | two instances closed; the *defect* was the enumeration. Inverting which side is enumerated turns a silent failure (rows live for ever) into a loud one (rows get `FAILED`, visibly) |
| 5 | Vocabulary **table + FK**, not a CHECK | a CHECK is a second copy that can drift; the FK **is** the table. Precedent: `reaper_policies` (335, RFC_018). The consumers are DB-resident SQL that no Go constant list can reach |
| 6 | `NOT IN (terminal)`, not `IN (non-terminal)` | equivalent while the FK holds; they **fail differently** without it. `NOT IN` treats an unknown status as reapable — loud. `IN` would make it immortal, i.e. bug 294 again |
| 7 | Delete the pause vocabulary | declared in five places, implemented in none, halves disagreeing on the spelling. The table is now where a pause status gets declared, and `is_pausable` is already read by the invariant |
| 8 | Scope to `orchestration_states` only | 8+ other status columns keep their inline CHECKs. Owner decision; leaves the estate with two patterns, recorded as a follow-up rather than pretended away |

## Corrections to earlier belief in this lane

- **`294` root-cause item 4 was wrong.** `monitoring.go` did not miscount — it read `orchestrator_state`,
  a relation that does not exist. Then **my own correction was also wrong**: I said its endpoints
  returned 500, but `AddMonitoringEndpoints` had zero callers so they were never mounted. Both
  errors had the same cause: reading a function instead of its callers.
- **`TimeoutMonitor` is dormant, not merely blind.** Nothing constructs it.
- **A landmine I wrote contained a false claim**, caught by running its own check before commit: my
  grep was scoped to `platform/ internal/ pkg/` and missed a third spelling-user in `test/`.

## Where we are going

Open, deliberately not done: a compare-and-swap for **both** takeover arms (guardian's objection on
`465` — the resume is a heuristic, not a lock); the estate-wide status-vocabulary question; and
`240`'s banner, whose symptom is gone but whose root cause was never re-opened.
