# NOTES — `bugs_open/029`, the retry-kills-live-child lane

Append-only, newest at the bottom. Technical log: evidence, the exact queries, and
every misstep.

---

## 2026-08-18 — session start: is 029 still valid, and is anyone on it?

**Ownership.** `scripts/who-owns.py 029` warns the number is ambiguous (`bugs_closed/029`
is the tool-suggester phantom-links case — resolve by slug). It named
`bugfix_029_dispatch_gate` as the likely owner, **quiet 14 days**. Because that check
reads *commits* and is therefore lagging, I also grepped the live session transcripts:
session `0a093b4e` (the `site_ai_agent_orchestration` lane) had been told by the owner at
12:06 today *"I have started bugfix 029 in another thread"* — that thread is this one. So
029 is mine, and the orchestration lane is a **stakeholder**, not a competitor.

**Still valid?** Yes. `agent_error_log` timeouts per day, last 10 days:

```sql
SELECT date_trunc('day', occurred_at)::date, count(*) FROM agent_error_log
 WHERE error_message ILIKE '%timed out after%' AND occurred_at > now()-interval '10 days'
 GROUP BY 1 ORDER BY 1;
-- 08-08:4  08-09:17  08-10:7  08-11:19  08-12:32  08-14:13  08-15:23  08-16:17  08-17:67
```

**The shape has SHIFTED, and this is the finding that reframed the lane.** Historically
this bug was about `spawn_*` steps. Over the last 4 days `spawn_agent` accounts for **4**
timeouts and `call_agent` for **116**:

```
build-pipeline-trigger | call_dispatch                    | call_agent  | 83
build-dispatch-loop    | process_item_iter_N_call_handler | call_agent  | 31
build-pipeline-trigger | spawn_dispatch                   | spawn_agent |  2
build-dispatch-loop    | process_item_iter_0_spawn_handler| spawn_agent |  2
```

So the spawn handshake largely got fixed and **the call leg did not**. Anyone still
reading 029 as a spawn bug is reading a 2026-07 file.

**Closed loose end, checked rather than assumed.** The 2026-07-29 contribution flagged that
`CHASSIS_RESPONSES_START_AT=latest` — the setting that closed the response-deafness class —
existed only as cluster state and in no file. It is now in the repo
(`deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml:58`,
commit `871c24665`, an owner D4 decision) **and** read by `platform/agentbase/agent.go:441`.
Nothing to do; recorded so the next reader does not re-chase it.

---

## 2026-08-18 — the measurement that located the mechanism

### 1. The declared window is honoured once, then silently overridden

`call_dispatch` in the live `build-pipeline-trigger` definition declares
`"timeout_seconds": 900`. Observed windows straight from `awaited_requests`:

```sql
SELECT step_name, retry_version, count(*) AS n,
       percentile_cont(0.5) WITHIN GROUP (ORDER BY (timeout_at-sent_at))::interval(0) AS median_window
  FROM awaited_requests WHERE sent_at > now()-interval '4 days'
   AND step_name IN ('call_dispatch','process_item_iter_0_call_handler','process_item_iter_1_call_handler')
 GROUP BY 1,2 ORDER BY 1,2;
```

| step | retry_version | n | median window |
|---|---|---|---|
| `call_dispatch` | 0 | 1492 | **15:00** ← as declared |
| `call_dispatch` | 1 | 37 | **05:00** |
| `call_dispatch` | 2 | 25 | **05:00** |
| `call_dispatch` | 3 | 92 | **05:00** |
| `process_item_iter_0_call_handler` | 0 | 1474 | **20:00** ← as declared |
| `process_item_iter_0_call_handler` | 1 / 3 | 1 / 31 | **05:00** |
| `process_item_iter_1_call_handler` | 0 | 1030 | **20:00** |
| `process_item_iter_1_call_handler` | 3 | 24 | **05:00** |

The cause is in `platform/orchestration/coordinator.go`, `handleRecoverableError`
(~3186–3202), introduced by `7b10c9df2` as an improvement on a hardcoded 60s and never
revisited:

```go
originalDuration := awaited.TimeoutAt.Sub(awaited.SentAt)
newTimeout := originalDuration
if newTimeout <= 0 || newTimeout > 30*time.Minute {
    newTimeout = 3 * time.Minute        // <- a step declaring >30 min gets THREE minutes
}
if newTimeout > 5*time.Minute {
    newTimeout = 5 * time.Minute        // <- everything else is capped at FIVE
}
```

**The inversion is the part worth carrying: the longer a step declares, the shorter its
retry window becomes.**

### 2. What the truncation costs, in failure probability

Of callee runs that COMPLETED in the last 36h, the fraction exceeding each window:

| callee | completed runs | > 5 min (**the retry window**) | > its declared window |
|---|---|---|---|
| `build-dispatch-loop` (declared 900s) | 545 | **25.5%** | 5.9% |
| `page-build-handler` (declared 1200s) | 221 | **17.6%** | 0.5% |
| `page-rerender` | 1375 | 0.1% | 0.0% |

So for `build-dispatch-loop` the truncation raises a retry's chance of failing from ~6%
to ~26%; for `page-build-handler`, from 0.5% to 17.6%. `page-rerender` is fast and is
untouched — which is the control, and is why "the fleet is fine" readings are so easy to
get from the wrong agent.

**Blast radius fleet-wide** — steps declaring more than 300s, from live `agent_definitions`:
**33 steps across 25 agent types**, at 600 / 900 / 1200 / 1800 / 2100 / 3600 / 21600 /
43200 / **86400** seconds. The 86400 one is a human approval step.

### 3. A hypothesis I tested and had to DROP

I expected the replay to manufacture **duplicate work**, on the reasoning that the
callee-side dedup key includes `retry_version`
(`platform/agentbase/agent.go:1173`, `HasProcessedMessage(corr, requestID, agentID, retryVersion)`),
so a replay can never be seen as a duplicate of the in-flight original.

Measured, and it came out **negative**:

```
bucket          | requests | callee_runs | runs_per_request
never retried   |      515 |         515 |             1.00
retried (v1)    |       21 |          21 |             1.00
retried (v2)    |        8 |           8 |             1.00
retried (v3)    |       23 |          23 |             1.00
```

Exactly 1.00 callee orchestrations per request in every bucket. **No duplicate runs.**
The reason is in the replay itself: it carries the *original child's* orchestration id
(that is `bugs_closed/129`'s fix — replay, never reconstruct), so it resolves the existing
child row rather than starting a new one. Recorded because I would otherwise have written
"the retry manufactures duplicate work" into the bug file, and it is false.

### 4. What the exhausted retries actually did to the callee

```sql
-- fate of the callee when the parent exhausted its budget (v3) on call_dispatch
FAILED    | 20 | median callee duration 04:26:46
COMPLETED |  3 | median callee duration 00:26:41
```

The 20 FAILED are all the reaper:

```
process_item_iter_1_spawn_handler | reaper: stale EXECUTING_STEP for >4h | 9
process_item_iter_2_spawn_handler | reaper: stale EXECUTING_STEP for >4h | 4
process_item_iter_3_spawn_handler | reaper: stale EXECUTING_STEP for >4h | 3
process_item_iter_4_spawn_handler | reaper: stale EXECUTING_STEP for >4h | 2
process_item_iter_{2,3}_spawn_handler | reaper: dispatch loop idle for >30 min | 1 each
```

**Measurement trap I walked into and caught.** My first read used `updated_at` as the
freeze time and got a suspiciously uniform `alive_for` of ~4h26m for every row. That is
not the freeze — `updated_at` is when the **reaper wrote FAILED**. The true freeze time is
`last_activity`. Same class as this bug file's own 2026-07-20 warning about `updated_at`
on `site_work_items`; different table, same mistake. **Use `last_activity` for the freeze,
`updated_at` only for the reap.**

### 5. The decisive correlation

With `last_activity` as the freeze time, every wedged loop froze at **~25m15s** after its
own creation — 17 of 18 inside an **11-second band** (25:11–25:22). That is a timer, not
load and not a roll (the rows straddle the 17:05Z chassis roll on 08-17 and are unaffected
by it).

Joining each frozen child to its parent's `awaited_requests` row for `call_dispatch`:

| child step | child ran for | parent retry_version | last window | **freeze − last send** |
|---|---|---|---|---|
| `process_item_iter_4_spawn_handler` | 25:17 | 3 | 05:00 | **+00:16** |
| `process_item_iter_4_spawn_handler` | 25:11 | 3 | 05:00 | **+00:11** |
| `process_item_iter_1_spawn_handler` | 25:14 | 3 | 05:00 | **+00:14** |
| `process_item_iter_1_spawn_handler` | 25:14 | 3 | 05:00 | **+00:14** |
| `process_item_iter_3_spawn_handler` | 25:14 | 3 | 05:00 | **+00:14** |
| `process_item_iter_2_spawn_handler` | 25:13 | 3 | 05:00 | **+00:13** |
| `process_item_iter_2_spawn_handler` | 21:10 | 3 | 05:00 | −03:50 ← **outlier, stated** |
| `process_item_iter_3_spawn_handler` | 25:17 | 3 | 05:00 | **+00:17** |
| `process_item_iter_1_spawn_handler` | 25:19 | 3 | 05:00 | **+00:19** |
| `process_item_iter_3_spawn_handler` | 25:14 | 3 | 05:00 | **+00:14** |
| `process_item_iter_1_spawn_handler` | 25:16 | 3 | 05:00 | **+00:15** |
| `process_item_iter_2_spawn_handler` | 25:22 | 3 | 05:00 | **+00:22** |

**11 of 12 froze 11–22 seconds after the parent sent its final replay.** One outlier froze
before it, and is reported rather than dropped.

`[MEASURED]` — the freeze, the send time, the windows, the fractions above.
`[UNVERIFIED]` — **why** the replay wedges the child. The leading candidate is two
concurrent drivers of one orchestration row racing on the optimistic lock introduced by
`7b10c9df2`, with the replay winning and doing nothing useful while the real worker's
update is lost. I have **not** proved that, and it is not part of the claim above.

### 6. Two candidate mechanisms I looked at and ruled OUT as the 4-hour block

- `platform/kafka/topic_manager.go` `WaitForTopic` — bounded at
  `KAFKA_TOPIC_WAIT_ATTEMPTS` (default 15) × ~3s jitter ≈ 45s. Not a 4-hour block.
- `coordinator.go:831`'s `maxAge = WorkflowPlan.TimeoutSeconds × 3` — real, but it
  **fails** the workflow with a distinctive "Orchestration stale" error. Our rows carry
  the reaper's error, not that one, so it did not fire.

### 7. Diagnosis loop filed before asserting

Per CLAUDE.md, a cross-cutting structural claim goes through `090` **before** it is
committed to. Queue was checked first and was empty (no duplicate filing).

```
intake correlation 0e4f89fb-fb46-4bb3-a658-aa939713fd88
run correlation    c8312dce-db45-4554-b2ab-5ac50e7e0c8a   <- artifacts are under THIS
```

Chassis uptime checked before dispatch (4h+, well clear of the ~300s post-roll drop
window) — the 2026-07-26 contribution to this bug file records what skipping that costs.

---

## 2026-08-18 — the `[UNVERIFIED]` half is now CODE-GROUNDED: the replay triggers a "take over"

Above I recorded, honestly, that I had proved *that* the replay wedges the child but not
*why*, and named an untested hypothesis (two drivers racing on the optimistic lock). I
have now found the path in the code, and the hypothesis was **right in shape and wrong in
detail** — the second driver is not an accident of concurrency, it is **deliberate, and the
coordinator invites it**.

`platform/orchestration/coordinator.go`, the status switch (~line 758):

```go
case StatusExecutingStep:
    // Check if stuck
    if state.CurrentlyExecuting != nil && time.Since(state.LastActivity) > StuckOrchestrationTimeout {
        s.logger.Warn("Found stuck orchestration, taking over",
            zap.String("stuck_step", *state.CurrentlyExecuting))
        if err := repo.ClearExecutingStep(ctx, state.OrchestrationID); err != nil { return err }
        state, err := repo.GetState(ctx, state.OrchestrationID)
        if err != nil { return fmt.Errorf("failed to reload state: %w", err) }
        return s.continueExecution(ctx, state, execCtx)     // <- concurrent second driver
    }
    s.logger.Info("Orchestration is actively executing")
    return nil
```

and `coordinator.go:38`:

```go
StuckOrchestrationTimeout = 5 * time.Minute
```

### The complete chain

1. The child is legitimately inside a long step. `process_item_iter_N_spawn_handler` runs
   `SpawnAgentAction`, which makes K8s API calls, creates and waits on two Kafka topics,
   and contains **two hardcoded `time.Sleep(5 * time.Second)`** — a step that routinely
   runs minutes without the coordinator writing anything.
2. **`last_activity` is not a heartbeat.** Grepped: it is written on insert and on
   `UpdateState` (`state.go:1033/1052`) and **nowhere else**. There is no mid-step touch.
   So a healthy child inside a long step goes quiet on that column by construction.
3. Past five minutes of that quiet, the child is **indistinguishable from a dead pod** to
   the code above.
4. The parent's replay arrives (at ~25 min, because of the window truncation in §1). The
   child's coordinator loads the row, sees `EXECUTING_STEP`, evaluates
   `time.Since(LastActivity) > 5m` as **true**, declares it stuck, **clears the executing
   step**, and calls `continueExecution` — **while the original worker is still running.**
5. Two drivers, one row, an optimistic version column. One loses and abandons its write.
   The row stops advancing and is left in `EXECUTING_STEP` with no live driver.
6. Nothing re-drives it. `TimeoutMonitor` and the retry driver both key on awaited
   requests, and it has none. The reaper's `EXECUTING_STEP > 4h` arm gets it, four hours
   later.

The observed **+11 to +22 seconds** between the parent's final send and the child's freeze
is consistent with Kafka delivery plus this takeover path running and the race resolving.

### Why this is the right place to call it a root cause

The takeover heuristic is **not stupid** — it exists to recover an orchestration whose pod
died mid-step, which is a real failure this estate has. The defect is narrower and more
interesting than "bad code":

> **`last_activity` is being used as a LIVENESS signal, but it is not maintained as one.**

Liveness needs something a live worker keeps refreshing. A timestamp that only moves when
the coordinator happens to write is a record of *progress*, and progress and liveness are
different things — a worker can be alive and making no writes for twenty minutes, which is
precisely what a spawn step does. Every long step is therefore a false positive waiting
for any message to arrive and trip it. **The retry replay is just the commonest such
message; it is the trigger, not the cause.**

That also predicts something worth stating because it is disconfirmable: **any** message
delivered to a child mid-long-step should be able to wedge it, not only a retry replay. I
have **not** tested that. `[UNVERIFIED]`

### Reuse before building: the estate already has the right mechanism

Do not invent a lease. `platform/orchestration/intake_repo.go:154` already has
`HeartbeatClaim(ctx, key, claimedBy, lease)` — *"extends the lease; reports false when the
claim is no longer held"* — used for intake claims. The orchestration takeover path uses a
bare timestamp comparison instead. A lease held by an identified holder, refreshed while
work is in flight, makes the bad state **unrepresentable**: a second driver cannot take
over while the first holds a valid lease, and a genuinely dead pod's lease expires on its
own. `orchestration_states` already carries `processing_node` and `currently_executing`,
so the holder identity is half-present already.

Ranking by the estate's own rule — *what makes the bad state unrepresentable, not what is
smallest*:

| | change | what it does | leaves what |
|---|---|---|---|
| 1 | **lease + heartbeat on step execution** | a live worker cannot be taken over at all | genuine pod death still recovered, on lease expiry |
| 2 | honour the declared retry window | fewer premature replays, so fewer triggers | a slow step is still takeover-able by any message |
| 3 | re-drive / faster reaper for `EXECUTING_STEP` | cuts dwell from 4h | the wedge still happens |

**2 alone would have made the measured outage rarer without making it safe** — which is
the point the PLAN makes about not conflating A, B and C. 1 is the one that closes the
door.

---

## 2026-08-18 — I tried to prove the takeover at RUNTIME and the check came back BLIND. Recording the blindness, not a zero.

The takeover path logs a distinctive line — `s.logger.Warn("Found stuck orchestration,
taking over", ...)`. If it fires in production, that is direct runtime proof of the
mechanism rather than a code read plus a correlation. So I went looking.

**Result: zero hits — and zero on the control too, which is the only reason this is not
now recorded as evidence of absence.**

| where | lines available | `Found stuck orchestration, taking over` | control: `Orchestration is actively executing` |
|---|---|---|---|
| both `agent-chassis` pods, `--since=24h` | 829 | 0 | **0** |
| 13 live `agent-build-dispatch-loop` pods | 3,811 | 0 | **0** |

**The control is the whole point.** `Orchestration is actively executing` is the *other*
arm of the same `switch` — any pod that evaluated `StatusExecutingStep` at all must emit
one line or the other. Zero on both arms means those pods never reached that code in the
window I can see, so the query cannot discriminate "the takeover does not happen" from
"I am looking at the wrong minutes". Without the control I would have written *"the
takeover never fires in production"*, which the data does not support in either direction.

**Why it is blind, and it is not fixable by trying harder:**

- **Chassis log retention here is ~4 minutes.** The earliest line `--since=24h` returns is
  `12:20:54Z` against a query at ~12:25Z and a pod that started at `07:57Z`. `--since=24h`
  returning four minutes of log is exactly the shape that makes a grep look thorough and
  be worthless. (MEMORY's *a "fresh build" can ship no new code* warns about controls that
  match everything; this is the mirror — a control that matches nothing.)
- **The population is not occurring right now.** Today has had **2** timeout rows fleet-wide,
  both `diagnose-*`. The wedges I measured are from **08-17** (67 timeouts that day), and
  the job pods that would hold those lines were reaped hours ago.
- This bug file already records the general form of this trap, from 2026-07-27: *"a hung
  spawned pod is evidence on a clock. Nothing in this fleet keeps one for you."* I hit the
  same wall from the log side.

**So the mechanism rests on two legs, not three, and I am saying so:** (1) the DB
correlation — 11 of 12 children freezing 11–22 s after the parent's final replay; (2) the
code path, read at HEAD, which is deterministic given `last_activity > 5 min`. The runtime
log confirmation is **not available** and I have not obtained it.

**What would obtain it, for whoever is here next.** Do not wait for a natural instance —
catch one live. `build-pipeline-trigger` fires every 60 s and is the free reproducer this
bug file has recommended since 2026-07-28. Watch for a `build-dispatch-loop` whose
`last_activity` has been static for >5 min while its status is `EXECUTING_STEP`, then
`kubectl logs -f` **that pod specifically** before its parent's next replay is due, and
capture `kubectl get pod -o yaml` in the same breath. The window between "identifiable"
and "reaped" is minutes.

> **The transferable rule, since it caught me twice today in different clothes:** a zero
> is only a finding when something in the same query could have come out non-zero. This
> morning it was `updated_at` (a number that could not have come out other than ~4h26m);
> this afternoon it is a log grep whose control is also zero. Same failure, opposite
> direction — one produced a false positive, the other a false negative.

---

## 2026-08-18 — the estate has TWO liveness proxies and MAINTAINS NEITHER; and the guarded takeover it built already exists but is not on this path

Two findings, and together they settle what the fix has to be.

### 1. There are two takeover paths, and only one is guarded

```
coordinator.go:290  ->  repo.TakeOverOrchestration(...)   # CAS on processing_node, guarded on the previous holder
coordinator.go:765  ->  repo.ClearExecutingStep(...)      # unguarded; THIS is the path that wedges children
```

`TakeOverOrchestration` was built for `bugs_open/075`, carries a long WHY comment, and is
protected by a source-scanning invariant test (`state_locks_test.go`) asserting that the
pod-name CAS and `UpdateStateWithVersion`'s version-CAS never govern the same column. It is
careful, reviewed work — and the `StatusExecutingStep` arm does not call it. It calls
`ClearExecutingStep`, which does a plain read-modify-write with no CAS and no notion of who
is holding the step.

So the *shape* of the answer is already in the tree. "Reuse existing machinery before
building new" points straight at it.

### 2. But `processing_node` is not a liveness signal either — and the code says so itself

I was about to propose "gate the takeover on `processing_node`" and stopped, because
`state.go`'s own comment on `SetExecutingStep` rules it out in advance:

> *"UPDATE does not list processing_node among its columns, so this assignment never
> reaches the database. Consequence: **processing_node records the pod that CREATED the
> row, not the pod driving it — which is why an ownership gate built on it could never
> distinguish a dead owner from a live one.**"*

**That is the same defect as `last_activity`, in a second costume.** The estate has two
columns that look like liveness and neither is maintained as such:

| column | written when | why it is not liveness |
|---|---|---|
| `last_activity` | on insert, and on `UpdateState` only | a healthy worker inside a long step writes nothing for minutes |
| `processing_node` | at row creation, and only ever moved by `TakeOverOrchestration` | names the creating pod, not the driving one — a dead pod's name persists for ever |

A takeover decision needs something a **live worker keeps refreshing**. Neither of these is,
so no combination of them can be made correct — which is why this cannot be fixed by
choosing the better column or tuning the 5-minute threshold. **Tuning the threshold is the
trap here**: any value is simultaneously too short for a legitimate spawn step and too long
for a dead pod, because the signal does not carry the distinction at all.

### 3. What that leaves

The estate already has a correct instance of the pattern that *is* missing:
`platform/orchestration/intake_repo.go:154` — `HeartbeatClaim(ctx, key, claimedBy, lease)`,
*"extends the lease; reports false when the claim is no longer held"* — with the
lease-expiry semantics spelled out at `:100`. Intake claims are held safely this way today.
Step execution is not.

So the fix direction is: **a lease on step execution, refreshed by the worker while the step
is in flight, with takeover permitted only on an EXPIRED lease** — modelled on the intake
lease rather than invented, and keeping `TakeOverOrchestration`'s column-separation
invariant intact (add to one side, never both — the council objection recorded at
corr `4a227ed9`).

That makes the bad state unrepresentable rather than unlikely: a second driver cannot take
over a live worker at all, and a genuinely dead pod is still recovered when its lease
expires — which is the behaviour the `:765` arm was written to provide and currently
provides unsafely.

**Consequence for the retry-window defect (A):** it is still worth fixing, but this reframes
it. With a lease in place, an early replay stops being destructive — it becomes merely
wasteful. So A drops from "causes the outage" to "causes avoidable load", and the ordering
in the PLAN stands: **the lease is the fix; honouring the declared window is the
efficiency.**

---

## 2026-08-18 — **CORRECTED: the replay does NOT kill a live child. It re-executes one that was already dead.** My headline claim was wrong.

> **This corrects the claim made throughout this file above, in `PLAN_2026-08-18`, in
> `README_where_we_are`, in the commit messages `55137dc2f` and `c2fdd2590`, and in two
> messages I sent to the `site_ai_agent_orchestration` lane.** The measurements above are
> all still correct. The *interpretation* of the central one was not.

**What caught it:** the Fable design pass, which I had briefed to contradict me if the code
did not support the diagnosis. It did exactly that, and then I grounded its claim myself
rather than taking it — a subagent's report is another document.

### The claim that failed

I wrote: *"the parent's own retry kills the child that was still legitimately working."*
The evidence was that 11 of 12 wedged children's `last_activity` froze 11–22 seconds after
the parent sent its final (rv=3) replay.

### Why it is wrong

**The takeover arm cannot fire on a healthy child.** Every `UpdateStateWithVersion` bumps
`last_activity` (`state.go:1052`), so `time.Since(last_activity) > 5m` on an
`EXECUTING_STEP` row *means the driving goroutine was already gone*. The precondition for
the takeover is that the child is already dead. I had read the guard and still asserted a
mechanism the guard forbids.

### The disconfirming measurement, which I ran because it could have come out either way

If Fable were right, each wedged child should carry **two** spawn-init `awaited_requests`
rows for the same step (`preRegisterAwaitedRequest`, `spawn_actions.go:162`) — one per
execution of that spawn. If I were right, there should be **one**.

```sql
SELECT ar.orchestration_id, ar.step_name, count(*) AS init_rows,
       min(ar.sent_at)::timestamp(0) AS first_spawn,
       max(ar.sent_at)::timestamp(0) AS last_spawn,
       max(os.last_activity)::timestamp(0) AS froze
  FROM awaited_requests ar
  JOIN orchestration_states os ON os.orchestration_id = ar.orchestration_id
 WHERE os.owner_agent_type='build-dispatch-loop' AND os.status='FAILED'
   AND os.error LIKE 'reaper: stale EXECUTING_STEP%' AND os.created_at > now()-interval '4 days'
   AND ar.step_name = os.current_step
 GROUP BY 1,2 ORDER BY 3 DESC, 4;
```

**17 of 18 returned `init_rows = 2`.** Sample:

| child | first_spawn | last_spawn | froze | gap 1→2 |
|---|---|---|---|---|
| `4b0e2854` | 18:26:39 | 18:36:09 | 18:36:22 | 9m30s |
| `838f8c14` | 18:52:22 | 19:01:40 | 19:01:53 | 9m18s |
| `c634317c` | 18:05:21 | 18:12:15 | 18:12:25 | 6m54s |

**And the 18th returned `init_rows = 1` — `2186f4b1`, which is EXACTLY the outlier from my
own correlation table** (the one at −3m50s that I reported rather than dropped). One spawn
at 16:53:19, froze 16:53:30, eleven seconds later, and no takeover — because at rv=3 its
staleness was under five minutes, so the arm declined to fire. Fable predicted that row's
behaviour before I looked. **The outlier I kept for honesty turned out to be the control
that decides the question**, which is the best argument I have yet met for not tidying
outliers away.

### The corrected sequence

1. The child's **own** `iter_N_call_handler` await — to its `page-rerender` grandchild — is
   retried on the same truncated 5-minute windows and exhausts. **The truncation bites
   inside the child too, and real page work is abandoned there.**
2. Exhaustion routes to `skipToNextLoopIterationForAsync`; the next iteration's
   `spawn_handler` runs, completes its init handshake — and then the continuation dies
   with no further state write. **This is the first freeze, and it is the actual "hung
   spawn" of 029.**
3. The parent's rv=1 and rv=2 replays land while the child is awaiting or <5m stale and
   are swallowed benignly (`ErrWaitingForResponse` → `nil`). **Budget burnt for nothing.**
4. The rv=3 replay finds >5m staleness → **takeover** → re-runs the spawn (a *duplicate*
   `page-rerender` agent and K8s job — a real, non-idempotent side effect) → wedges
   identically → **and re-stamps `last_activity`, resetting the 4-hour reaper clock.** The
   reaper fired at freeze₂+4h, not freeze₁+4h.

So what I measured as "the freeze" was the **second** freeze, and the +11–22 s was the gap
between the takeover re-running the spawn and that re-execution dying the same way.

### What survives, what dies, and what is now open

| claim | status |
|---|---|
| retry window truncated to 5 min / 3 min above a 30-min declaration | **STANDS** — measured and code-read; and it is now *worse* than I thought, because it also fires one level down, inside the child |
| 33 steps across 25 agents affected | **STANDS** |
| the frozen row is unreachable for 4h and starves the site for 40 min | **STANDS**, and is worse: each takeover *extends* the 4h |
| the replay is destructive | **STANDS, but the damage is different** — duplicated non-idempotent side effects and a reset reaper clock, not the killing of live work |
| **"the retry kills a child that was still working"** | **WITHDRAWN** |
| what kills the continuation after the first spawn handshake | **OPEN — this is now the centre of 029, and I do not know the answer** |

### Consequence for the 090 run in flight

The symptom I filed asserts the withdrawn mechanism. Whatever it returns, it is answering a
question that is now partly wrong — so I will **not** cite a CONFIRMED verdict from it as
support for the replay-kills-live-work claim, and a REFUTED verdict would be correct rather
than surprising. The narrow symptom worth filing next is Fable's: *what kills the
post-handshake continuation?* Fable's own candidate — that the durable ticker's recovery
runs under a shared **60-second** context (`cleanupExpiredAwaitedRequests`,
`coordinator.go:4264`), so a continuation that spawns agents cannot finish inside it and
every error-path write dies with the deadline — is plausible and **[UNVERIFIED]**; the
fast-path timer uses `context.Background()`, so it cannot be the whole story.

---

## 2026-08-18 — the 090 run was KILLED BY THE BUG IT WAS FILED ON, and it is the sharpest evidence in this file

The diagnosis run finished terminal with **no diagnosis**: 5 `bundle` artifacts, no verdict.
The reason is this bug.

```
diagnose-dispatch-loop | COMPLETED | last_activity 12:54:30
diagnose-orchestrator  | FAILED    | 12:54:27   error: "Request cdf1a95b… timed out after 3 retries"
diagnose-agent         | COMPLETED | 12:56:58
```

**The child COMPLETED at 12:56:58. The parent gave up at 12:54:27 — 2 minutes 31 seconds
before its child finished successfully.** A 42-minute diagnosis run was done, and the answer
was thrown away because the parent stopped listening.

`call_diagnoser` **declares `timeout_seconds: 1800`**, and the stored `workflow_plan` carries
it. Its `awaited_requests` row at `retry_version 3` had a window of **00:03:00**. A **tenfold**
shortfall.

### Why 180s and not the 300s I predicted — the boundary is unreachable as an equality

I expected 1800s to be capped to 300s (`> 30*time.Minute` is false for exactly 1800). The row
says 180s, so I read the construction again:

```go
SentAt:    time.Now(),
TimeoutAt: time.Now().Add(getTimeout(step)),
```

**Two separate clock reads.** So the stored `TimeoutAt - SentAt` is always *marginally more*
than the declared value — and a step declaring **exactly** 30 minutes therefore trips
`newTimeout > 30*time.Minute` and falls into the **3-minute** arm. **Six live steps declare
exactly 1800.** My round-1 test annotation said the old code gave them 5 minutes; it gave 3.
Corrected in the test with the live case cited beside it.

This is a nice instance of the estate's own rule: the live row disagreed with my arithmetic,
and the row was right. I had reasoned about the constant and not about how the two operands
are produced.

### Does the fix I committed actually repair this case? Yes, and here is why it is not circular

`retryWindow` calls `datahelpers.ConvertStepTimeout` itself and then `GetStepTimeout` on the
step **read from the stored plan** — which I verified carries `config.timeout_seconds: 1800`
for this very step. So on the next roll this request's retries get **1800s**, not 180s, and a
parent whose child finishes at +42 min is still listening. `[INFERRED]` — this is the code
path read against a verified plan row, **not** an observation of the fixed binary, which
cannot exist until the roll. The disconfirming result is named in RSH-010's `verify-later`:
if `rv>=1` windows stay at 05:00/03:00 on a binary the provenance stamp says carries the fix,
the plan lookup is failing and the fallback is being taken.

---

## 2026-08-18 — COUNCIL ROUND 1: **REVISE**, and the gating objection found a real defect in my own submission

Corr `7c92389a-617f-4abc-b03b-0ef84ca2239f`, gated by **`editquality`**, 5 seats abstained.

**The objection, and it is correct:** my rationale ranked a lease/CAS guard as *"(1) the fix"*
and stated *"shipping only (2) [the retry-window fix] would be the classic mistake here and
the plan does not do that"* — **and then the plan contained only the retry-window edit.**

That is a straight self-contradiction and I put it there. **The cause is a FOSSIL:** the
ranking paragraph was drafted *before* the afternoon's correction, when I still believed the
replay destroyed live work and a lease was therefore the thing that would have prevented the
outage. When the mechanism was withdrawn I updated `grounded_in` — round 1's evidence array
already carried the withdrawal — **and did not update the ranking sitting above it.** A
document can carry its own refutation and its stale conclusion in the same breath, and the
seat read both.

**How I answered it: by withdrawing the claim, NOT by widening the plan.** The lease cannot be
"the fix" any more, because the mechanism that ranked it is gone — there is no live worker for
it to protect. It is still worth building for a narrower reason (it stops the replay
re-executing a corpse, duplicating a spawn and resetting the reaper clock), but that is a
bounded waste-and-dwell argument. **Bundling it in now, to make the plan look proportionate to
a ranking I no longer believe, would be shipping a change on a justification I know to be
stale — which is the same error one level up, and exactly what the seat had just caught.**

Round 2 resubmitted on the same correlation (`RESUBMIT_CORR`) so the trail accumulates: the
ranking is replaced by an explicit statement that **nothing here closes 029**, the live
specimen above is added as grounding, the 1800s arithmetic is sharpened, and a scope risk is
declared naming the counter-argument (33 steps are being silently under-waited today
regardless of what 029 turns out to be).

**Worth recording plainly: this is the second time today an outside reviewer caught something
I had the evidence to catch myself.** Fable caught the mechanism; the council caught the
fossil. Both were cheap. The common factor is that each was asked to disagree — and in both
cases the disconfirming material was already sitting in my own document.

---

## 2026-08-18 — COUNCIL ROUND 3: **APPROVED** (2 advisory objections, none high). Three rounds, two real defects found.

Corr `7c92389a-617f-4abc-b03b-0ef84ca2239f`, *"approved with 2 advisory objection(s) — none
high-severity"*, 3 abstained. **Verdict READ, so `Council-Reviewed:` is now writable.**

| round | verdict | gated by | what it found |
|---|---|---|---|
| 1 | REVISE | `editquality` | **a real self-contradiction** — the rationale ranked a lease as "the fix" and said shipping only the window fix "would be the classic mistake", then shipped only the window fix. A fossil of the withdrawn mechanism. |
| 2 | REVISE | `prior_art_librarian` (HIGH) | **a real asserted absence** — "no live caller reachable from the chassis" with no caller graph, used to justify not fixing a named twin. |
| 3 | **APPROVED** | — | `editquality` and `prior_art_librarian` both moved to approve. |

**Both gating objections were things I had the material to catch myself**, and neither was a
matter of care: the first was a stale paragraph sitting above its own refutation in the same
document; the second was a claim I had simply never checked. **The instruction to disagree is
what bought them** — the same property as the Fable brief earlier in the day.

### The advisories, and what I did about each

- **`reuse_agent` (medium) — ACTED ON.** `getTimeout(step)` already exists and sets the
  *initial* awaited window at both registration sites; `retryWindow` re-implemented the
  lookup instead of going through it. Fair, and now fixed: `retryWindow` calls `getTimeout`
  after conversion, so the declared timeout is read **one way rather than two**. That is this
  bug's own defect class (two implementations of one judgement) and the seat was right to
  flag it in a change whose whole point is that a second reader of a timeout disagreed.
- **`editquality` (low) — accepted, not fixed.** The *fallback* branch still derives from
  `TimeoutAt - SentAt`, the poisoned source. True. It fires only when the plan has lost the
  step, and the alternative — failing the retry outright — is worse than a slightly wrong
  window. Disclosed rather than silently narrowed.
- **`bug_historian` (medium) — already actioned.** The dormant twin as a recurring shape. It
  now has a `LANDMINES.md` entry with the caller-graph check and the data-side tell, and it
  is in RSH-010. Not fixed, deliberately and on the record.
- **`guardian` (medium) — accepted.** Stability preference on core dispatch plumbing, and the
  concession that long-declared steps now wait longer. That is the declared intent of the
  config; the counter-argument is in the submission's `risks` and stands.
- **`guardian` / `prior_art_librarian` (low) — noted, not answerable.** The caller graph is
  prose in `grounded_in` rather than a `code_check` the seats could run themselves. They are
  right that this is weaker, and it is a limit of the tool: `code_checks` is
  declaration-only, so a call-site search cannot be expressed there. **Worth carrying: an
  absence claim is exactly the kind a reviewer cannot verify from the submission, which is
  why the LANDMINES entry gives the reader the grep rather than the conclusion.**

### The one thing in the advisories worth chasing later, recorded so it is not lost

`reuse_agent` noted that `getTimeout(step)` sets the **initial** window at
`TimeoutAt: time.Now().Add(getTimeout(step))`. But `call_diagnoser` declares 1800s and its
rv0 window was **180s** — so on *that* path the declared value did not reach the initial
await either, while `call_dispatch` got its declared 900s on the same path shape. **So the
initial-window path may have its own conversion gap, separate from the retry truncation this
change fixes.** `[UNVERIFIED]` — I have not traced which registration site each takes.
Deliberately not widened into this round; it is a fresh question, not a loose end of this fix.

---

## 2026-08-18 — a handed-over lead CHECKED and CLOSED: the `page_rerender` failure cluster is NOT 029

The `site_ai_agent_orchestration` lane handed over ~107 failed `page_rerender` rows, 72%
concentrated in three sites, webdesign.co.uk still failing today — offered as the place the
suspected initial-wait conversion gap would show. They were right to offer it and right not
to diagnose it. **It is not 029, and it is not a standing defect at all.**

```sql
SELECT count(*) FILTER (WHERE error ILIKE '%timed out after%')    AS the_029_signature,
       count(*) FILTER (WHERE error ILIKE '%claim timed out%')    AS claim_timeouts,
       count(*) FILTER (WHERE error ILIKE '%save_page_sections%') AS content_gating,
       count(*) FILTER (WHERE error ILIKE '%commit%' OR error ILIKE '%github%') AS git_api,
       count(*) AS total
  FROM site_work_items WHERE item_type='page_rerender' AND status='failed'
   AND updated_at > now()-interval '5 days';
--  7 | 2 | 17 | 76 | 108
```

**76 of 108 are GitHub API / git-adapter failures** (`failed to get latest commit/base tree`,
`failed to create commit: github API request…`). Only **7** carry this bug's signature.

And they are **one incident, not a pattern** — every git failure falls inside a single
**2h35m window on 2026-08-17, 13:37 → 16:12 UTC**, across six sites:

| site | git fails | first | last |
|---|---|---|---|
| webdesign.co.uk | 28 | 13:37:35 | 15:35:44 |
| gamesdesign.co.uk | 22 | 13:49:34 | 16:12:34 |
| robot-hands.com | 16 | 14:45:41 | 15:59:36 |
| loancalculator.co.uk | 6 | 15:59:50 | 16:10:54 |
| cookly.uk | 2 | 14:49:16 | 14:49:48 |
| relojistas.com | 1 | 15:31:17 | 15:31:17 |

Fleet-wide, time-bounded, self-limiting, already over — the shape of a GitHub API outage or
rate-limit window, not a defect anyone needs to fix. **The "three sites carry 72%" reading is
an artefact of which sites had the most rerender traffic during those 2½ hours**, not of
anything wrong with those sites. `[MEASURED]`.

**The disconfirming result was available and is why this is worth recording:** if the
initial-wait gap were showing in this population, the `timed out after` count would dominate.
It is 7 of 108. So this lead is closed rather than parked — I am not carrying it forward as a
suspicion, and nobody should re-open it on the row count alone.

**Not filed as a bug.** A dated, finished incident with an external cause is not a
`bugs_open/` entry, and filing one would put a permanent-looking record against six sites for
a two-hour API wobble. If it recurs, the query above is the check.

> **ADDENDUM, same day — the closure above holds and my ACCOUNT of it was incomplete.**
> Caught by the `site_ai_agent_orchestration` lane, whose "webdesign.co.uk is still failing
> today" I had quietly treated as covered. It was not: my incident window closes at
> **2026-08-17 16:12Z**, and there are failed `page_rerender` rows created on **08-18** at
> 00:38, 11:58, 11:59 and 12:29Z. **The one-incident explanation does not reach them**, and I
> framed the whole population as though it did.
>
> Re-measured here rather than taken on their word — the population created after my window
> closed:
>
> ```sql
> SELECT count(*) AS total_after_window,
>        count(*) FILTER (WHERE error ILIKE '%timed out after%') AS the_029_signature,
>        count(*) FILTER (WHERE error ILIKE '%commit%' OR error ILIKE '%github%') AS git_api,
>        count(*) FILTER (WHERE error ILIKE '%save_page_sections%' OR error ILIKE '%rebuild_policy%'
>                           OR error ILIKE '%claims floor%' OR error ILIKE '%REFUSED%'
>                           OR error ILIKE '%rerender_page_sections%') AS content_gating
>   FROM site_work_items WHERE item_type='page_rerender' AND status='failed'
>    AND created_at > '2026-08-17 16:12:34+00';
> --  15 | 0 | 0 | 15
> ```
>
> **15 rows, 0 with this bug's signature, 0 git, all 15 content gating** — `overwrite:
> REFUSED`, `rebuild_policy=owned`, `claims floor blocked: 19 banned claim(s)`. So there are
> **three** populations here, not two: the 08-17 git incident, older content gating, and a
> live content-gating population that is still producing rows. **The negative for 029 is
> stronger than I stated (0 of 15 in the freshest slice), and my "one incident" framing was
> doing work it could not support for part of the range.**
>
> **The transferable bit is theirs and it is good:** their "still failing today" was a **true
> fact doing no work** — accurate, offered in good faith, and unconnected to the conclusion it
> sat next to. I did not test it against my own window before writing the closure. A claim
> being true is not the same as a claim being *load-bearing*, and an untested true one reads
> exactly like a tested one.

---

## 2026-08-18 16:26Z — POST-ROLL VERIFICATION on v1.0.1309: **the fix SHIPPED. The behavioural claim is NOT yet proven, and here is exactly why.**

The fleet rolled to **v1.0.1309** (both replicas started `15:45:31Z` / `15:45:53Z`). Running
RSH-010's `verify-later`, in the order it specifies.

### 1. Did it ship? YES — proven at the artefact, with controls in both directions

**The provenance log line was NOT in range**, exactly as CLAUDE.md warns: `--tail=3000`
returned no `build provenance` on either pod, because it is a STARTUP line and retention here
is ~4 minutes. **That is "not in range", not "unstamped"** — so I used the binary probe, which
has no shelf life, and probed for **known** values rather than discovering one.

| probe | result |
|---|---|
| `f0117fb8b93ea3e1f32298daeb9751bcff4b90c7` | **PRESENT** on replica 1 |
| same sha, replica 2 (`…-wpk5w`) | **PRESENT** — both replicas carry one build |
| `9fa90ea38…`, `3f1426a8d…`, `06a1ccfcb…`, `5ce70ed38…` (near-neighbour commits) | absent |
| `deadbeefdeadbeef…` (negative control) | **absent — control OK** |

So the build point is `f0117fb8b` (committed 16:26:18 BST, pod up 16:45:31 BST — consistent).

**"Did my fix ship?" is then a git question, not an inference:**

```
git merge-base --is-ancestor bf7646a29 f0117fb8b   -> SHIPPED
git merge-base --is-ancestor 2a3d30ec3 f0117fb8b   -> SHIPPED
git merge-base --is-ancestor 6e23dabb7 f0117fb8b   -> correctly NOT an ancestor (control)
```

⚠ **Note for the next reader: probing the binary for MY OWN commit sha returned absent, and
that is correct.** The binary carries the sha it was **built from**, not its ancestors. A
probe for your own commit is the wrong test and will tell you your fix did not ship when it
did. The ancestor test is the right one.

### 2. Does it work? **UNPROVEN — the discriminating population has not occurred.**

Post-roll `awaited_requests`: **237 rows, 234 at rv0, 3 retried.** So there is demand, and the
demand control passes — this is not a dead path.

**The positive control passes exactly as designed** — rv0 windows are UNTOUCHED:

| step | rv0 window | declared |
|---|---|---|
| `call_dispatch` | **15:00** | 900s ✓ |
| `process_item_iter_{0..4}_call_handler` | **20:00** | 1200s ✓ |
| `spawn_dispatch`, `…_spawn_handler` | 02:00 | — |
| `deploy_page` | 03:00 | — |

**But all three retries are on steps that declare NO timeout at all**, verified in the stored
plan (`step_in_plan = t`, `cfg_tmo` and `step_tmo` both EMPTY):

| step | rv1 window | what it means |
|---|---|---|
| `spawn_rerender_agent` | 03:00 | `retryWindow` correctly returned the 180s system default |
| `process_item_iter_0_spawn_handler` | 03:00 | same |
| `spawn_dispatch` | 02:00 | 120s — set by the spawn path's own `preRegisterAwaitedRequest`, not by `retryWindow` `[INFERRED]` |

**A step declaring nothing yields 180s under BOTH the old code and the new one.** The old
truncation only bit declarations *above* 300s. **So these three retries cannot distinguish the
fix from its absence, and I am not counting them as evidence.** No step declaring >300s has
retried since the roll.

> **This is "a zero with no traffic is a dead path" wearing a new costume, and it is the reason
> the check was built with a rv0 control.** I *have* traffic — 237 rows — just not on the path
> under test. Traffic on the agent is not demand on the mechanism. Had I reported "verified,
> windows correct" off these three rows, every number in the report would have been true and
> the conclusion unearned.

**The 03:00 readings are the fix behaving correctly, not a defect** — but they are also the
corner `editquality` flagged at LOW in the approval round ("the fallback branch still derives
from the poisoned source"). Here the plan lookup *succeeded* and the step simply declares
nothing, so the default is the right answer. Recorded so nobody reads 03:00 as the old
3-minute inversion: **the inversion produced 03:00 for steps declaring 1800s+; this is 03:00
for a step declaring nothing.** Same number, opposite meaning — check the declaration before
concluding.

### 3. What would settle it, and it is now armed

A retry on any step declaring >300s: `call_dispatch` (900s), `process_item_iter_N_call_handler`
(1200s), `call_diagnoser` (1800s). **Expected: the declared value (15:00 / 20:00 / 30:00).
Under the old code: 05:00, or 03:00 for the 1800s case.** A background watch is running on
exactly that predicate. Until it fires, RSH-010's status is **SHIPPED, BEHAVIOURALLY
UNPROVEN** and says so.
