# PLAN — 2026-08-02 — bound a local action's execution, so a blocked call cannot park an orchestration

`bugs_open/169` **part A** only. Part B (UUID-ordered site starvation) is fixed and live
— `sql_for_agents/284` plus `bugs_closed/176` — by the "bugfix 19" session.

## The defect, stated as a mechanism

A local action runs with **no deadline anywhere in the chain**:

```
continueExecution → executeStep → executeLocalAction → executeAction → handler(ctx, params)
                                                        (coordinator.go:1563)
```

`executeAction` recovers panics and bounds nothing. No caller wraps the context. So any
action that blocks on a network call parks its orchestration at `EXECUTING_STEP`
indefinitely, and — for `build-dispatch-loop` — holds its `site_work_items` row in
`claimed` until the 120s `claimed-item-timeout` reaper eventually expires it.

`timeout_seconds` **looks** like the bound and is not. It is parsed
(`coordinator.go:1348-1353`) into `execCtx.TimeoutSeconds` with the comment *"leave it
for the action to set a default"*, and **no action reads it as a deadline on its own
execution** (checked across 271 action files).

## What the evidence supports, and what it does not

`orchestration_state_audit`, 3-day window (2026-07-31 → 08-02):

| measure | value |
|---|---|
| `spawn_*` step executions | **6,951** |
| p50 / p95 / p99 | **0s / 18s / 24s** |
| executions over 300s | **exactly 1**, at 14,475s |
| distinct runs entering `*_spawn_handler` | 165 |
| runs that ENDED at a `*_spawn_handler` step | 1 (status `FAILED`) |

The distribution is bimodal with nothing in between: healthy spawns finish in seconds;
the pathological one ran four hours. **A bound anywhere between a minute and an hour
would have caught it and touched none of the other 6,950.**

**[CORRECTED]** Five orchestrations showed ~14,500s gaps and I first read them as one
stalled run, then as a pod restart. Both wrong — five *different* orchestrations at five
*different* times, each stalling almost exactly **4h01m**. That is
`coordinator.go:831`'s `maxAge = WorkflowPlan.TimeoutSeconds × 3` stale-orchestration
guard. **It does not make part A benign**: it is an orchestration-*age* check inside the
message-handling path, so it fires only when a message next arrives, marks the row
failed, and never interrupts the blocked goroutine.

**[MARKED]** The `lag()`-over-audit-rows method measures wall-clock between recorded
transitions. For a local action that approximates execution time; for an
`AWAITING_RESPONSES` step it is mostly idle waiting. That is why the >600s tail is
dominated by `call_*` steps — **remote awaits, legitimately slow**. Any bound must apply
to **local actions only**.

## The fix

**Bound the local action call with a context deadline**, in `executeLocalAction`, and
route a timeout to the normal step-failure path so the work item is released rather than
left `claimed`.

### Why a context deadline and NOT "run it in a goroutine and abandon it"

Abandoning the goroutine guarantees the coordinator regains control even against an
action that ignores `ctx` — but it leaks the goroutine, and a late write from an
abandoned action into a step the coordinator has already failed is a data hazard worse
than the hang. The plausible four-hour blockers here are all **ctx-aware** (`database/sql`,
the Kafka producer's `ProduceWithValidation(ctx, …)`, the K8s client), so a deadline
should actually cut them. Start with the safe mechanism; escalate only on evidence that
it is insufficient.

### Why NOT reuse `timeout_seconds` — measured, not assumed

Of the live steps carrying `timeout_seconds`, **53 of 64 are `call_agent`**, and the rest
are mostly waiting semantics too (`await_approval` at 86400, `request_human_input`,
`dispatch_thunder_*`). The key means **"how long to wait for something external"**.
Repurposing it as "how long this action may execute" would make one shared word mean two
things — the exact defect RFC 006 has just been decided on. So: a **new, explicitly named
key**, and `timeout_seconds` is left alone.

### Shape

- `local_action_timeout_seconds` on a step overrides; **`0` or negative disables** the
  bound for that step (a deliberate escape hatch).
- Default **600s**. That is ~25× the observed all-step p99.9 (214s) and 25× the spawn p99
  (24s), and no local action legitimately runs longer — `await_approval` and
  `request_human_input` send a notification and return, they do not block.
- On expiry: a clear error naming the action, the step and the elapsed time, routed to
  the existing failure path (so `error_step` still works).

## Risk, and the honest statement of it

This changes what the framework **guarantees for every action** — from "may run forever"
to "is cancelled at N". Per CLAUDE.md's 2026-07-29 ruling §1 that is a guarantee change,
so it needs the council and is written up here as such. The failure mode to fear is a
too-tight bound causing fleet-wide breakage, which is worse than the rare hang it fixes.
Mitigations: a generous default, a per-step override, an explicit off switch, and the
measured distribution above showing 6,950 of 6,951 spawns finishing inside 24 seconds.

## Verification

- unit tests: an action that blocks past the deadline returns a timeout error; one that
  finishes inside it is untouched; `0` disables the bound; the default applies when the
  key is absent. **Each proven by mutation**, not just by passing.
- the bound must be shown to *fire* — a passing test suite that never exercised the
  timeout would be the same "inert by omission" trap RFC 006 just closed.

## What this does NOT do

It does not explain *which* call blocked for four hours — the orchestration row is
reaped and only one audit row survives. It makes that class survivable rather than
diagnosing the instance. `090` was filed as the bug file instructs
(`RUN_CORRELATION_ID=3ca53d45-4826-4935-96a3-a0af4d194d91`); its verdict is reconciled
against these findings in NOTES before anything is claimed as root cause.
