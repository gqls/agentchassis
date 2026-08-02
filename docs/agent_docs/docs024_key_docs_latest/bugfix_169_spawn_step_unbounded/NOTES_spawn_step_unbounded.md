# NOTES — bugs_open/169 part A — a local action has no deadline

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-02 — taking part A, and what the first pass established

**Part B is already fixed and live** (session "bugfix 19", `sql_for_agents/284`, plus
`bugs_closed/176`). Only **part A** — the spawn that sat at `EXECUTING_STEP` for 38+
minutes — is open, and the file explicitly instructs running `090` before committing
to a cause.

### Ownership, checked properly rather than by verdict

`who-owns.py 169` says **"OWNED or recently active"**. That verdict is lagging by
construction — it reads commits — so it was not taken at face value. The sharper checks:

- live `.jsonl` transcripts naming part A's own identifiers (`9da39de8`, `b286f2f5`):
  two sessions. One last wrote 12 hours ago; the other, `9de5c96a`, wrote **5 minutes
  ago** — but it is the **bugfix_154** lane writing a handoff, and its own note records
  *"`169` part A (spawn hang) untouched"*. So it is recording part A as open, not
  working it.
- nobody is mid-edit on the code a fix would touch: `git status` on
  `spawn_actions.go` / `coordinator.go` is clean, and there are **no commits to either
  in 3 days**.

**Conclusion: part A is genuinely unowned.** Recorded because "OWNED" from the script
would have been enough to wrongly stand down.

### Is it still valid?

The original artefacts are gone — the orchestration row is reaped (terminal rows are
kept ~24h) and item `b286f2f5` is now `failed` at `attempt_count=3`. So validity had to
be re-established from the class, not the instance:

- `orchestration_state_audit`, 3-day window (07-31 → 08-02): **165 distinct runs**
  entered a `*_spawn_handler` step across 2,283 transitions. **Exactly one** ended
  there, `FAILED`, at `process_item_iter_4_spawn_handler` — the same step name part A
  reports.

So the class still occurs, and it is rare.

### The mechanism, read first-hand

- `executeAction` (`coordinator.go:1534`) calls `handler(ctx, params)` with **no
  deadline**. It recovers panics; it does not bound duration.
- the chain above it — `continueExecution → executeStep → executeLocalAction →
  executeAction` — adds no deadline either.
- `timeout_seconds` IS parsed (`coordinator.go:1348-1353`) but only copied into
  `execCtx.TimeoutSeconds`, with the comment *"leave it for the action to set a
  default"*. **No action reads it as a deadline on its own execution** — the only hits
  across 271 action files are `spawn_actions.go` *setting* it on outgoing contexts
  (30/300/5/3), a different `IdleTimeoutSeconds`, and unrelated struct fields in
  `hitl_persistence.go` / `thunder_ssh_exec_dispatch.go`.
- `SpawnAgentAction` performs `setupAgentTopics`, `spawnAgentKubernetesJobFromDefinition`,
  `preRegisterAwaitedRequest` and `sendInitializationMessage` in sequence, plus two
  unconditional `time.Sleep(5s)` and a `sendMessageWithRetries` loop (10 × 8s ≈ 72s
  worst case). None of that explains 38 minutes; **the absence of a bound does.**

### The number that will decide the fix

`spawn_*` steps over the same 3-day window: **6,951 executions**, p50 **0s**, p95
**18s**, p99 **24s** — and **exactly one** above 300s, at 14,475s. The distribution is
bimodal with nothing in between: healthy spawns finish in seconds, the pathological one
ran four hours. A 5-minute bound would have caught it and touched none of the other
6,950.

### MISSTEP — I formed a restart hypothesis and it was wrong

Five orchestrations showed ~14,500s gaps. I assumed one stalled run, then a pod restart.
Both wrong: they are **five different orchestrations at five different times across
three days**, each stalling for almost exactly **4h01m** (14,475–14,593s, a 118s spread).
Five independent runs landing within two minutes of the same duration is a *timer*, not
a coincidence.

It is `coordinator.go:831` — `maxAge = WorkflowPlan.TimeoutSeconds × 3` (60 min fallback
when unset). At `timeout_seconds ≈ 4800` that is exactly 4h, and the overshoot matches
detection lag.

**Why this matters and is not a footnote:** that guard is an *orchestration-age* check
inside the message-handling path. It fires only when a message next arrives for that
orchestration. It therefore does **not** interrupt a blocked action — the goroutine
stays blocked; the row is marked failed later. So it is not a step timeout and does not
make part A benign.

**The cheap check that would have saved the wrong turn:** print the gap *boundaries*,
not just the durations. Five different start times killed the restart theory instantly;
the durations alone had looked like one event.

### Also caught: my own measurement's blind spot

The `lag()`-over-audit-rows method measures **wall-clock between recorded transitions**,
which for a local action approximates execution time but for an `AWAITING_RESPONSES`
step is mostly idle waiting. That is why the >600s tail is dominated by `call_*` steps
(`call_dispatch`, `call_scraper`, `call_diagnoser`) — those are **remote awaits and
legitimately slow**. Any bound must apply to **local** actions only. Marked here because
a fleet-wide timeout chosen off the unfiltered distribution would have been wrong.

`090` filed as instructed: `RUN_CORRELATION_ID=3ca53d45-4826-4935-96a3-a0af4d194d91`.

## 2026-08-02 (later) — fix shipped, and two missteps of my own

**The fix** is `fe34fd04f`: `executeLocalAction` derives a deadline via
`localActionContext` and passes it to the handler. Register **RSH-004**, council
`2c6800e6` submitted alongside. Default 600s, per-step `local_action_timeout_seconds`,
`<=0` = explicitly unbounded (Warn every time), `DISABLE_LOCAL_ACTION_TIMEOUT=true` as a
fleet-wide kill switch. Malformed value falls back to the **default**, never to
unbounded — a config typo must not silently restore the defect.

Seven tests, and the three risky behaviours **proven able to fail by mutation**:
removing the bound, wiring it to `timeout_seconds`, and making a malformed value mean
unbounded. First attempt at mutations 1 and 3 produced *build* failures rather than test
failures — which proves nothing about the guards — so they were redone as compiling
mutations. Worth remembering: **"the test suite went red" is not the same as "the guard
fired"**, and a mutation that fails to compile tells you only that your sed was wrong.

### MISSTEP 1 — I shipped the seam one commit ahead of its register entry

CLAUDE.md's ordering-exemption condition (2) is now the whole requirement and it says
**the same commit**. I committed `fe34fd04f` and registered RSH-004 in `91c82e36f`.
Logged in `WRONG_CALLS.md`. The cheap check is a question about the command already
being typed: *is the register entry in this pathspec?*

### MISSTEP 2 — backticks in `git commit -m` executed as shell

`claimed: command not found`, `timeout_seconds: command not found`, then
`git: Argument list too long`. Benign because it failed loudly; the dangerous version is
a backticked word that IS a valid command. This is already a LANDMINE I hold in memory
and I walked into it anyway. The fix is not better escaping — it is `git commit -F
<file>`, which removes the whole class. Used for the retry; worked first time.

### The 090 run returned no diagnosis, and my wait condition was wrong

I filed `090` as the bug file instructs (`3ca53d45`). It ran **five iterations**,
produced **five `bundle` artifacts and no diagnosis** — no `fix_plan`, no
`council_report` — and the work item completed.

Two things to separate here:

1. **My error.** I set a background wait on `kind IN ('verdict','diagnosis','final')`.
   **No such kind exists.** The kinds that exist are `fix_plan`, `council_report`,
   `bundle`, `escalation`, `iteration_note`. So my "still running" reading was partly my
   own watcher waiting for something that could never appear. Check the vocabulary
   before waiting on it.
2. **What is actually true.** Even with the right vocabulary, nothing under that
   correlation carries a conclusion. The bundles are auto-gathered *evidence* (they do
   corroborate the surroundings — every other `build-dispatch-loop` run in the window
   completed in 20–50s).

So the root cause rests on **first-hand verification, stated plainly** as CLAUDE.md's
2026-07-31 ruling requires when the loop is substituted: the call chain read end to end,
all 271 action files checked for a reader of `TimeoutSeconds`, the distribution measured
over all 96,047 recorded step executions.

**[UNMEASURED]** Whether the diagnosis loop routinely completes without a conclusion. One
run is not a rate, and there is no other diagnosis run in the last 10 days to compare
against — every other correlation in that window is a council run. Recorded as an
observation, deliberately **not** filed as a defect.
