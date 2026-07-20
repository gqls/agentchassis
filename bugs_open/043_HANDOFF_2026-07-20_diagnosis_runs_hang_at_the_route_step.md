# 043 — diagnosis runs hang at the `route` step, so the loop returns no verdicts

**Filed 2026-07-20** (vetcomparison thread). **Status: OPEN.**
**Cause NOT determined** — this is an observation file. Deliberately not filed *into* the
diagnosis loop, because the diagnosis loop is the thing that is stalling; filing there would be
futile. That is also why it matters: CLAUDE.md now makes the loop the **default** for any durable
claim, and today it answered nothing.

## Symptom

Every diagnosis run filed from this thread on 2026-07-19/20 produced bundles and then stopped at
the same step. None reached a verdict.

| correlation | outcome | step |
|---|---|---|
| `be60b0d7-21c4-4e02-be95-2ec37387004f` | FAILED — `reaper: stale AWAITING_RESPONSES for >90 min` | `spawn_diagnoser` |
| `12ff5852-3950-4fec-ac68-1cc8a86d0521` | FAILED | `call_diagnoser` |
| `459fbdf3-5e48-4fe4-977a-10676d7437f6` | FAILED — `reaper: stale EXECUTING_STEP for >4h; step=route` | **`route`** |
| `f155b0c4-881b-4369-abe4-569d7b2ad4c8` | hung EXECUTING_STEP ~4 h at time of filing | **`route`** |
| `55dc0fa4-116c-40d6-90b2-bfad9ad73692` | hung EXECUTING_STEP ~4 h at time of filing | **`route`** |

Three of three runs that got as far as `route` hung there. Bundles were produced in each case
(`diagnosis_artifacts kind='bundle'`), so gathering works; it is the controller that does not
advance.

## What `route` is

`diagnose-agent`'s `route` step runs the `diagnose_route` action — the controller that loops
between `load_runtime` (gather) and `emit`, forwarding the verdict's data requests and overriding
`next_step` to either loop back or stop:

```json
{"action": "diagnose_route",
 "config": {"emit_step": "emit", "gather_step": "load_runtime",
            "state_field": "route.diagnose_state", "verdict_field": "verdict.result",
            "analysis_field": "repo_analysis", "max_iterations": 5}}
```

## A tidy hypothesis that turned out to be WRONG

`route`'s config carries `max_iterations: 5`, a **numeric** value — and `/bugs_open/042` proves
numeric step config never reaches actions that read it through `ExtractActionInputs`. A
controller silently receiving the wrong iteration bound would explain a loop that never
terminates, and it would have been an elegant story: the config defect disabling the loop that
would diagnose the config defect.

**It is not the cause.** `diagnose_route_action.go:103` reads
`datahelpers.GetIntField(config, "max_iterations", 5)` — straight off the raw config map,
bypassing `ExtractActionInputs` entirely. 042 does not touch this path.

Recorded because the next person will form the same hypothesis, and because a plausible story
that fits the symptom is not evidence. Check the call site before adopting it.

## Where to start

- Is `route` waiting on the verdict child (`verdict.result` never populated), or looping without
  a termination condition? `execution_path` / `orchestration_state_audit` for a hung correlation
  will distinguish these.
- `diagnose_route`'s advance logic and its guards, against a run whose bundles exist but whose
  `verdict` field is empty.
- Whether this correlates with the 2026-07-20 image roll — `459fbdf3` was filed and died before
  it, so probably not, but confirm rather than assume.

## Operational consequences, worth knowing regardless of cause

- **A filed diagnosis is not a delivered one.** Check for a terminal verdict before treating a
  correlation id as an answer. Three id's in this thread's docs would otherwise read as
  "diagnosed".
- **A dead run leaves its intake reading `awaiting_diagnosis`**, which makes the 090 trigger
  refuse a refile as a duplicate. Close the stale intake first (record why in
  `spec.failure_reason`), then refile — note `idx_swi_dedup` is partial over non-terminal
  statuses, so the closed row and the fresh one coexist correctly.
- **The `stale-orchestration-reaper` itself stalled** for ~2 h on 2026-07-20 (last_triggered
  17:01 against a 180 s interval) before recovering, so hung runs were not even being reaped
  during that window.
- **The council gate is not starting either — and this is broader than the deploy.** The
  orchestration layer created nothing at all between 18:53 and 19:06 UTC while the scheduler kept
  beating normally, and a council submission at 19:03 (`712be028`) produced zero
  `orchestration_state_audit` rows and never started. I attributed that to the quiet window and
  resubmitted at 19:12 once general orchestrations were flowing again (8 created in 5 min).
  > **The resubmission (`563462b8`, orchestration `44dcdbe3`) did not start either** — zero audit
  > rows, zero artifacts, 14 minutes on, while unrelated orchestrations continued to be created
  > throughout. So the quiet window is **not** a sufficient explanation, and my first reading of
  > it was wrong. Whatever is stopping `route` from advancing may be the same thing stopping
  > fix-loop work from being picked up at all; both the diagnosis loop and the council gate ride
  > the same machinery. **Tally for 2026-07-19/20: five diagnosis runs and two council
  > submissions, zero verdicts.**
  CLAUDE.md documents a ~300 s no-dispatch window after a chassis restart — this quiet began
  ~55 min after the roll, so that rule does not cover it. **Confirm orchestrations are actually
  being created before trusting any dispatch**, and note that zero audit rows means *queued*
  (~16 min under backlog) and cannot by itself be distinguished from *dropped* without watching.

## Related

- `/bugs_open/042` — the config-literal defect; its diagnosis run (`f155b0c4`) is one of the hung
  ones, which is why 042 carries a full case file rather than relying on a verdict.
