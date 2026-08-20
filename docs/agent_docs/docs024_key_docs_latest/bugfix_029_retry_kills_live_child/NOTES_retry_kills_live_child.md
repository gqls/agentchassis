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

---

## 2026-08-18 16:55Z — LEAD 1 (the "initial wait has its own conversion gap") is **REFUTED**, and the instrument that nearly sold me a false positive

The handoff carried this forward `[UNVERIFIED]`: *"`call_diagnoser` declares 1800s and its
**rv0** window was 180s, while `call_dispatch` got its declared 900s on the same path shape."*
It came out of the approval round, not from NOTES — there is no measurement behind it in this
file. There is now, and it does not reproduce.

### The measurement

`call_diagnoser`, **entire table history, no time filter**:

| retry_version | window | n | span |
|---|---|---|---|
| **0** | **00:30:00** | **29** | 2026-08-11 18:21 → 2026-08-18 15:59 |
| 1 | 00:03:00 | 1 | |
| 2 | 00:03:00 | 1 | |
| 3 | 00:03:00 | 4 | |

**29 of 29 rv0 rows carry the declared 1800s.** Every 180s reading sits at rv1/rv2/rv3 — i.e.
on the *retry* path, which is the inversion this lane already fixed. The lead almost certainly
read a retry row as an initial one; `retry_version` is the column that separates them and it is
easy to drop from a `SELECT`.

Generalised beyond the one step — every rv0 window joined to **its own owning agent's**
declaration (3 days): **18 agent/step pairs, 18 OK, zero mismatches**, from `call_dispatch`
(900, n=511) through `call_handler` (2100), `call_scraper`/`call_indexer`/`call_diagnoser`
(1800) down to `dispatch_list` (60). Query in the RUNBOOK.

**The disconfirming result was available and named in advance:** any row where observed <
declared. None exists.

### The misstep: I manufactured two MISMATCHes with my own join, then had to un-manufacture them

My first pass aggregated the declaration with `max((s.value->'config'->>'timeout_seconds')::int)`
across **all agents** declaring a step of that name. It reported two mismatches, both false:

- `process_item_iter_N_call_handler`: observed 1200, "declared" 2100. **Two agents declare
  `call_handler`** — `diagnose-dispatch-loop` at 2100 and `report-dispatch-loop` at 1200. The
  ~3,500 iter-expanded rows belong to the latter and observe exactly its 1200. My `max()` chose
  the other agent's number.
- `trigger_deploy`: observed 120, "declared" 180. Same shape — `rerender-site` 180,
  `section-editor` 120; the rows are section-editor's.

**A step name is not a key.** The declaration lives at (agent, step), and an aggregate across
agents fabricates a disagreement between two things that were never compared. Had I written
this up one query earlier it would have gone in as "two steps are silently truncated at
registration" — a fresh false mechanism, in the file whose whole history is false mechanisms.
`[MEASURED]`, both the wrong version and the right one.

A second artefact from the same pass, worth stating so nobody re-runs it: steps my CTE labelled
"no declaration" (`call_verifier` 240s, `call_ingester` 300s, `call_section_editor` 600s,
`call_planner` 1100s) are **absent from `agent_definitions` altogether** — they come from
per-run plans in `orchestration_states.workflow_plan`, which is what `retryWindow` actually
reads. `agent_definitions` is not the whole universe of declarations, and treating it as one
turns "I did not look there" into "it is not declared".

### The bias this census HAS, which I cannot fully remove

**`awaited_requests` is keyed `PRIMARY KEY (request_id)` — one row per request, rewritten in
place on each retry.** `retry_version` is bumped and `sent_at`/`timeout_at` are *overwritten*.
So a retried request's earlier windows are **destroyed, not archived**, and:

> **Selecting `retry_version=0` silently selects requests that never retried.** It is a
> survivorship filter wearing the costume of a neutral one.

That matters here because it biases *towards* my conclusion: a step truncated to 180s at
registration would time out sooner, retry, and **leave the rv0 pool** — exactly the rows that
would refute me. So the census cannot be left to stand on its own count.

**Why the conclusion survives it anyway:** a truncated registration only becomes invisible if
its short window *always* expires. Where the callee answers inside 180s the row stays at rv0
and the 180 is visible. `call_dispatch` has ~1,000 rv0 rows with `min = max = 900` exactly —
no mixture, no tail. For the defect to hide there, every one of a thousand truncated requests
would have had to also time out, on a step whose normal completion is well inside three
minutes.

**Residual, stated rather than hidden:** this does not exclude a step whose truncated window
*always* expires. `[UNMEASURED]` — and unmeasurable from this table, because the evidence is
overwritten by the retry that would prove it. If that case is ever suspected, the instrument
has to be the emitted log line at registration, not this table.

### Status change

Lead 1 moves from `[UNVERIFIED]` **carried forward** to **CLOSED — refuted**. It is not a
suspicion to re-open on the strength of one 180s row; check `retry_version` first, and if the
row is rv≥1 it is Part A's territory and already fixed.

---

## 2026-08-18 16:55Z — the behavioural proof: still unproven, now WATCHED rather than polled

Post-roll (v1.0.1309, 15:45:31Z) at 16:43Z, **58 minutes in**: 237 → ~350 awaited rows, and
**still exactly the same three rv≥1 rows** recorded at 16:26Z (`spawn_rerender_agent`,
`process_item_iter_0_spawn_handler`, `spawn_dispatch`) — all on steps declaring nothing, all
non-discriminating for the reasons already in this file.

**The rv0 positive control continues to hold**, now on a bigger sample: `call_dispatch` 15:00
(n=30 post-roll), `process_item_iter_{0..4}_call_handler` 20:00, `call_diagnoser` 30:00.

**How long a wait is reasonable — measured, not guessed.** Arrival rate of the discriminating
event (a retry on a step declaring >300s), by hour over 5 days: typically **1–3/hour**, with
bursts (23 in the 08-17 15:00 hour). Zero in the first hour post-roll is ~a 20% outcome at that
rate — unremarkable, **not** yet evidence of anything, in either direction.

> ⚠ Note for whoever reads a *drop* in retries as proof the fix worked: **it is not.** The fix
> changes how long a retry waits, not whether one is sent. It does plausibly suppress
> *downstream* retries (rv2/rv3 were 117 of the 157 pre-roll `call_dispatch` retries — cascades
> off a 5-minute rv1), so the rate may genuinely fall. But a falling rate is also what a quiet
> fleet looks like, and the two are not separable by counting. **The window on a single rv≥1
> row is the test; the rate is not.**

**Armed, replacing the previous session's dead watch:** `scratchpad/watch_rsh010.sh` polls the
discriminating predicate every 3 min for 4 h and exits **0 = fired / 2 = query error / 3 =
window elapsed** — three distinct exits, so a broken kubeconfig can never be read as "no
retries". Foreground-tested before arming, against a pre-roll window where it **must** fire: it
returned 8 rows, every one truncated (05:00 on 900s declarations, 03:00 on 1800s). That control
doubles as the pre-roll baseline — **157 `call_dispatch` retries across 5 days, not one
untruncated.** The old code's distribution has zero variance, which is why a *single* post-roll
row at 15:00 will settle this.

---

## 2026-08-18 17:05Z — the WEDGE has a signature, and it is much sharper than "the continuation dies"

Went to file the 090 the handoff asks for, and censused the population first. The 18 wedged
rows are not a scatter — they are one shape, repeated.

```sql
SELECT owner_agent_type, current_step, count(*), min(last_activity), max(last_activity)
  FROM orchestration_states
 WHERE status='FAILED' AND error LIKE 'reaper: stale EXECUTING_STEP%' GROUP BY 1,2;
```

**All 18 are `build-dispatch-loop`, all at a `process_item_iter_N_spawn_handler` step.** Three
properties hold across the whole population, and each is tighter than I expected:

1. **`last_activity - created_at` is 25:14–25:22 in 17 of 18** (the 18th is 21:10). A uniform
   lifetime to within eight seconds is a **timer**, not a hang.
2. **Every single one has the PRECEDING iteration's `call_handler` in status `error`** — froze
   at `iter_1_spawn` ⇒ `iter_0_call_handler:error`; at `iter_2_spawn` ⇒ `iter_1_call_handler:error`;
   at `iter_3_spawn` ⇒ `iter_2_call_handler:error`; at `iter_4_spawn` ⇒ `iter_3_call_handler:error`.
   **18 of 18, no exceptions.** The freeze is on the ERROR path, never the happy path.
3. **The final spawn step is registered TWICE in 17 of 18** (gap up to 9m37s), and **the
   parent's last state write lands 9–16 seconds after that final send — in all 18.**

Sequence, read off `awaited_requests` ordered by `sent_at`: iteration N-1's `call_handler`
exhausts its retries and goes `error` → the loop advances → `iter_N_spawn_handler` is
registered and **processed** (the handshake SUCCEEDS) → the same spawn step is registered
again minutes later and is **also processed** → the parent writes state once more, ~12s later,
and never again. **`iter_N_call_handler` is never registered at all**, which is why the row
holds no `waiting` awaited request and is invisible to everything except the 4-hour reaper.

Worked example, `23eb0107` — iterations 0 and 1 complete normally (`call_handler` rv0, 20:00,
`processed`), iteration 2's `call_handler` ends **rv3 / 05:00 / error**, `iter_3_spawn_handler`
is sent 7s later, sent AGAIN 8m43s after that, and the row freezes 12s later.

**Where the 25 minutes comes from `[INFERRED]`, not measured:** `call_handler` declares 1200s
and the old retry cap was 300s — 20:00 + 5:00 = 25:00, against an observed 25:14–25:22. It fits
to within handshake overhead, but I **cannot** confirm it from this table: the row is rewritten
in place, so rv0/rv1/rv2's sends are destroyed (see the survivorship note above). The stored
rv3 window of 05:00 on a step declaring 1200s is the old inversion, and that much IS measured.

**This does not mean Part A fixes the wedge.** It means the wedge's ENTRY CONDITION — a
`call_handler` reaching terminal `error` — was being reached faster than the step's declaration
allowed. Whatever kills the continuation afterwards is untouched by Part A and still unexplained.
Do not let the arithmetic fitting become a claim that the wedge is closed.

### ⚠ CORRECTION to what I wrote 20 minutes ago in this same session

I read "all 18 wedges are on 08-17, none since" as evidence the wedge had **stopped**. Then I
checked the instrument:

```sql
SELECT min(created_at) FROM orchestration_states WHERE created_at > now() - interval '10 days';
--  2026-08-17 14:35:29
```

**The oldest retained row in the table IS the first wedge row.** `orchestration_states` holds
only ~26 hours (08-17 and 08-18 only; 3,534 rows for 08-18, 2,099 for 08-17, and just 25 rows
older than 7 days in the entire table). So "the wedges begin at 14:35 on 08-17" is **the
retention boundary talking, not the fleet** — anything earlier was pruned, and the population
is bounded below by the instrument. A first occurrence sitting exactly on the edge of the
window is the classic shape of a phenomenon the window is CUTTING THROUGH, not one that started
there. `[MEASURED]` boundary, `[UNMEASURABLE]` prior extent.

**The half that survives is the other half:** 08-18 is fully retained, carries ~3,534
orchestrations, and contains **zero** wedges. That is a real absence over a full day of traffic
— and it predates the roll, so it is not evidence for the fix either. The 08-17 window also
overlaps the GitHub API incident already recorded in this file (13:49–16:12Z), and the retained
FAILED vocabulary shows `github API request failed` rows clustered 16:31–18:35 inside the same
window. **Correlation recorded, mechanism NOT claimed.**

### The wedge evidence has a shelf life — hours, not days

At ~26 hours of retention, **the entire 18-row population ages out during 2026-08-19.** Anyone
who wants a diagnosis run to read these rows must run it against them **now**; after that the
only copy is the dump in this lane's scratch and the tables above. This is not a reason to rush
a bad run — it is a reason not to assume the evidence will wait.

---

## 2026-08-18 17:05Z — the 090 is BLOCKED, and the blocker is bigger than the 090

`090` refused on its coverage check (four rows, all one **`failed`** item from 2026-08-12 about
`persistAwaitingStateWithRetry`, sharing only the file `coordinator.go`). `failed` is not in
the clause's exclusion list — `NOT IN ('complete','cancelled','rejected')` — so a terminal item
from six days ago reads as live coverage. `FORCE=1` is the documented override and this is what
it is for.

**But the script's other message is the one that matters, and it is not advisory in effect:**

```
local HEAD is 233 commit(s) ahead of origin/087_towards_multiple_domains —
the diagnosis reads origin/087_towards_multiple_domains
```

Verified rather than taken on trust:

| check | result |
|---|---|
| `f0117fb8b` (the commit the LIVE binary was built from) on any remote branch | **NO — none** |
| `retryWindow` present in `coordinator.go` on any remote branch | **NO — absent everywhere** |
| `bf7646a29` / `2a3d30ec3` (the Part A fix) on origin | **NO** |
| origin tip | `d3db49975`, **2026-08-18 12:39:51 BST** |
| the 233 unpushed commits | **all dated 2026-08-18**, 12:51 → 17:55 BST |

So this is a **same-day, ~5-hour push gap**, not chronic rot — origin was advanced at 12:39 and
the fix was committed at 13:46/14:19, after it. Two consequences, and the second is the one to
carry:

1. **A 090 filed now reads a tree with the OLD capping block in `handleRecoverableError`.** My
   symptom text says the wedge follows a `call_handler` that "exhausts its retries" — a
   diagnoser reading the unfixed tree has every reason to land on the truncation as the cause,
   which is Part A, already fixed and live. It would be a correct answer about the wrong tree,
   arriving with citations. **That is the exact failure the script's REF note warns about, and
   it is why I have not forced it.**
2. **The binary serving production was built from a commit that exists in no remote.** The
   estate's build path (`git archive HEAD`) makes an unpushed commit shippable, and nothing
   downstream requires it to be pushed. Not a defect of this lane, and not mine to fix silently
   — **routed to the owner as a decision**, because pushing 233 commits belonging to a dozen
   concurrent lanes is not a call one session should make on its own initiative.

**Status: 090 authored, seeded, coverage-read, and NOT dispatched.** The symptom text is in
this lane's scratch and reproduced above; it needs only the tree question settled.

---

## 2026-08-18 18:30Z — **PART A IS BEHAVIOURALLY PROVEN.** One row settled it, exactly as designed

The discriminating event arrived at **18:28:21Z**, 2h43m after the roll, on the 35th poll.

```
step_name     | rv | sent_at             | timeout_at          | window   | status
call_dispatch |  1 | 2026-08-18 18:28:21 | 2026-08-18 18:43:21 | 00:15:00 | processed
```

`call_dispatch` declares **900s**. The retry was granted **15:00 — its declaration, exactly.**

**Why one row is enough here, which is not usually true.** The disconfirming value was named in
advance and the old code's distribution has **zero variance**: every `call_dispatch` retry sent
before the roll, all history, is 05:00 —

| rv | window | n |
|---|---|---|
| 1 | 00:05:00 | 56 |
| 2 | 00:05:00 (+1 at 00:04:59) | 30 |
| 3 | 00:05:00 | 133 |

**219 pre-roll retries, not one above 05:00.** A single post-roll row at 15:00 is therefore not
a sample of a noisy distribution — it is an outcome the old code could not produce.

**Three controls, all passing:**

1. **Positive control — rv0 windows UNCHANGED**, on a much larger post-roll sample than the
   16:26Z check had: `call_dispatch` 15:00 (n=87), `process_item_iter_0_call_handler` 20:00
   (n=89), `call_diagnoser` 30:00 (n=4). Nothing else moved, so this is the retry path and not
   a global timeout change.
2. **`status = processed`.** The child answered *inside* the new window. This matters more than
   the number: a longer interval written to a column proves the arithmetic; a `processed` at
   rv1 proves the window was **used** — under the old code this request expires at 18:33:21 and
   becomes rv2. It is the first observation in this lane of a retry that **succeeded** where the
   old code would have escalated.
3. **Two independent watchers, same row.** The previous session's watch was still alive and
   fired on the same event with the same reading. ⚠ Instrument-independence, **not**
   evidence-independence — both read the same DB row, so this rules out a broken script, not a
   misread table.

### Status changes

**RSH-010: `SHIPPED, BEHAVIOURALLY UNPROVEN` → `SHIPPED AND BEHAVIOURALLY PROVEN`** (register
entry and index row updated; `verify-later` satisfied on its own stated terms).

**And what it does NOT change, which is the whole point of how this was written up:** `bugs_open/029`
stays **OPEN**. Part A was always one contributing defect. The wedge signature measured earlier
today — 18/18 on the error path, the final spawn registered twice, the parent's last write 9–16s
after it — is untouched by this and remains unexplained. **A proven fix to a contributing defect
is not a closed bug**, and the artefacts have said "part A only" in those words since the first
commit precisely so that this moment could not be misread as the bug closing.

---

## 2026-08-19 09:10Z — v1.0.1314 verified; the wedge evidence EXPIRED as predicted; and I talked myself into a suppression claim the data does not support

### 1. The new build, at the artefact

`v1.0.1314`, both replicas (`07:52:27Z` / `08:05:39Z`). The `build provenance` line was **not in
range** again (startup line, ~4 min retention) — so, binary probe with controls:

| probe | result |
|---|---|
| `d3590ca46` (08-18 22:17 BST) | **PRESENT on both replicas** — the build point |
| `8e3f29197`, `f110d0397` | absent |
| `deadbeef…` negative control | absent — control OK |

**Part A is still aboard**, by the ancestor test rather than a sha probe:
`bf7646a29` and `2a3d30ec3` are both ancestors of `d3590ca46`; a later commit (`d8065ea87`)
correctly is not. rv0 windows on the new build are unchanged (`call_dispatch` 15:00,
`process_item_iter_0_call_handler` 20:00). **No retries yet on 1314**, so RSH-010's proof still
rests on the 18:28:21Z row from 1309 — which is fine: the code is byte-identical by ancestry.

### 2. The wedge evidence is GONE, exactly as this file predicted

`orchestration_states` now starts at **2026-08-18 07:58:20**. Wedged rows retained: **0**. All 18
instances aged out overnight, on the schedule the 08-18 handoff named. Nothing was lost that was
not already transcribed here — but the diagnosis loop reads the LIVE DB, so **a `090` filed today
has no instances to cite.**

### 3. ⚠ CORRECTION, within ten minutes, to a claim I had already started writing up

I measured the wedge's entry condition — a `call_handler` reaching terminal `error` — and got:

| era | errored | rows | rate |
|---|---|---|---|
| pre-roll (08-12 → 08-18 15:45) | 31 | 6,260 | 0.50% |
| post-roll | **0** | 504 | 0% |

I wrote "**the entry condition has stopped occurring**", computed a Poisson tail
(expected 2.5, P(0) = 0.082), called it *suggestive but not decisive*, and went looking for the
sample size that would settle it (606 rows). **Every number there is correct and the framing is
worthless**, because I had not asked how those 31 errors are distributed:

| day | errored | call_handler rows |
|---|---|---|
| 08-12 | 0 | 448 |
| 08-13 | 0 | 232 |
| 08-14 | 0 | 1,241 |
| 08-15 | 1 | 1,302 |
| 08-16 | 0 | 481 |
| **08-17** | **30** | 1,436 |
| **08-18** | **0** | **1,603** |
| 08-19 | 0 | 20 |

**30 of the 31 are one day.** And the day that matters most is **08-18: 1,603 call_handler rows,
zero errors — the majority of them BEFORE the fix rolled at 15:45.** The entry condition was
already absent on the unfixed binary, for a full day, on more traffic than the post-roll sample.

**So the post-roll zero carries no information about Part A.** A quiet day is the baseline: six
of eight days have zero. My Poisson arithmetic assumed a uniform rate on a process that is
plainly **bursty**, and pooling a 30-error day with six zero-days to manufacture a "0.50%
expected rate" is what produced a p-value that felt like evidence. **The disconfirming result
was one `GROUP BY` away and I ran it only because I stopped to ask whether the burst was the
whole population.**

**What can honestly be said:** the wedge, and the `call_handler` errors that precede it, are
**EPISODIC** — one observed burst (08-17, 30 errors, 18 wedges, ~4 hours) against seven quiet
days. That burst overlaps the GitHub API incident already recorded in this file. Post-roll quiet
is consistent with the fix having suppressed it AND with the fix having done nothing, and this
data cannot separate those. `[MEASURED]` distribution; `[UNKNOWN]` effect of Part A on it.

**Consequence for closing 029:** we cannot say the wedge is fixed, and we cannot say it is still
biting. What we can say is that it is rare, bursty, unexplained, and that its one observed
occurrence coincided with an external outage. That is a materially different open state from
"unexplained and actively biting", and the bug file now says so.

---

## 2026-08-19 10:00Z — the capture is BUILT and PROVEN (RSH-011); the 090 ran and returned NOT CONFIRMED, for exactly the reason the capture exists

### 1. `wedge-evidence-capture` — RSH-011, live and induced

Hourly CronJob at `:17` UTC, `deployments/kustomize/services/wedge-evidence-capture/`. It
snapshots frozen orchestrations **and their full `awaited_requests` set** into `doc_notes`
(`subject_type='pipeline'`, `categories ? 'wedge-evidence'`, key `wedge-evidence:<orch_id>`).

**Why it captures LIVE wedges and not just reaped ones — this is the whole design.** Reading
the live `database-cleanup` `pre_query` turned up a fact the lane did not have:

| arm | live threshold | repo seed says |
|---|---|---|
| `COMPLETED`/`FAILED` deleted | **24 hours** | 7 days |
| `EXECUTING_STEP`/`AWAITING_RESPONSES` deleted | **4 hours** | 24 hours |

The stale reaper terminates a frozen orchestration at **4 hours**. Cleanup **deletes** an
`EXECUTING_STEP` row at **4 hours**. **Same threshold, two processes, one row.** Where the
reaper wins you get a `FAILED` row carrying the stale marker that lives 24h more; **where
cleanup wins the orchestration is deleted having never been recorded at all.** So the 18
instances of 08-17 are the ones where the reaper happened to win, and **any rate computed from
reaped rows is a lower bound on a population we cannot enumerate.** `[MEASURED]` on the
thresholds, `[INFERRED]` on how often cleanup wins — that is unmeasurable for the same reason.
The capture takes the row at freeze+30min, ~3.5h before either process reaches it.

**Proven by induction, not by deployment.** A job that has only ever reported "nothing to
capture" has not shown it can capture, so `WEDGE_FROZEN_MINUTES` is env-overridable and was
driven to 0:

| run | result |
|---|---|
| clean (threshold 30) | exit 0, one summary row, 0 captured — correct, nothing is wedged |
| **induced (threshold 0)** | **5 real orchestrations captured**, each with its full iteration sequence; exit 1 |
| **immediate re-run** | **0 captured, no duplicate `subject_key`** — dedupe holds |
| cleanup | all 7 induced rows deleted; production starts clean |

⚠ **The first deploy FAILED in-cluster** on `doc_notes`'s `subject_type` CHECK constraint —
`orchestration` is not among the eight permitted values. The query was right, the data was
right, the report printed, and the write was thrown away. **Local SQL cannot catch it because
the constraint is only met where the job runs.** Both this and the retention discrepancy are
now in `LANDMINES.md`.

### 2. The 090 ran to completion — and returned **NOT CONFIRMED (UNVERIFIABLE)**

Run correlation `b346d0d4-bf9b-4068-9db7-5af18d719706`, 3 iterations, ~12 minutes, terminal
`complete`. **This is the first diagnosis run this lane has got a terminal answer from** — the
2026-08-18 attempt was killed at 12:54:27 by the very inversion Part A fixed.

**Verdict: `NOT CONFIRMED (stopped: scope-not-narrowing)`, last verdict REFUTED, no fix
proposed, handed to a human with the trail.** And the loop's own `data_requests` say precisely
why, in its words:

> *"Looks for a currently-stuck instance of the same shape (EXECUTING_STEP parked at a
> `*_spawn_handler` step) **to substitute fresh occurrence evidence for the 2026-08-17 burst
> that has already aged out of retention**."*

It then found a candidate — a `build-dispatch-loop` at `process_item_iter_4_spawn_handler` —
asked whether the preceding `iter_3_call_handler` had ended in error and whether the spawn was
registered twice, and **refuted it**. Checked afterwards: that orchestration
(`6f8c447b-c9e3-40ed-ac60-234fbfb109af`) **COMPLETED normally at 09:52:14**. It was in flight,
not wedged. **The loop was right to refute**, and the refutation is a transient in-flight row
being mistaken for a wedge — which is the exact false positive the capture's 30-minute
threshold exists to exclude, arrived at independently.

**So: a well-run refutation on absence of evidence, not a wrong answer.** It cost one run and it
established something worth having — the hypothesis cannot be tested against a table with no
instances in it, and the loop said so rather than confabulating a cause from the code alone.

**The two now compose:** the next burst is captured within 30 minutes and preserved
indefinitely; the 090 re-filed against those rows has the occurrence evidence this one went
looking for and could not find. Re-file the symptom verbatim from the 2026-08-18 17:05Z section
and add the captured `doc_notes` keys to the runtime tier.

---

## 2026-08-19 ~10:45Z — the evidence is NOT gone: it survives in `awaited_requests`, and it kills candidate 1

Picked the lane up cold from `bugs_open/029` and the 08-19 handoff. Re-verified the live state
first (below), then went at the open question — *what kills the parent after its spawn handshake* —
by reading source. Ended up somewhere better than expected, because **one of this lane's own
stated facts was wrong and I caught it by accident.**

### 0. Live state, re-verified before touching anything

| check | result |
|---|---|
| wedged rows retained | **0** (`build-dispatch-loop` / FAILED / 4 days → 0 rows) |
| `wedge-evidence-capture` (RSH-011) | ran **10:17Z**, `Complete`, *"newly captured this run: 0"* — the cron is alive and reporting |
| `agent_error_log` timeout census | last `build-dispatch-loop` entry **2026-08-17 18:52**; nothing since |

Not reproducing. Consistent with the handoff.

### 1. ⚠ MEASUREMENT TRAP I walked into first: `min(created_at)` says five weeks of retention

```sql
SELECT min(created_at), max(created_at), count(*) FROM orchestration_states;
--  2026-07-13 11:25:27 | 2026-08-19 10:43:41 | 3859
```

Read naively that is **five weeks** of retention and the handoff's "~26 hours" looks wrong. It is
not. Broken down by day and status, the table holds **08-18 and 08-19 only**, plus 24 `CANCELLED`
rows from 07-19…07-24 and one `FAILED` from 07-13 — statuses the cleanup never prunes. **The
aggregate is dominated by exactly the rows the retention rule does not apply to.**

The trap is that it errs in the *reassuring* direction: a session checking "is the evidence still
there?" with a `min()` gets told yes and stops looking. **Retention is a per-status property here;
measure it with `GROUP BY status`, never with a whole-table `min()`.**

### 2. Candidate 1 — the ticker's shared 60-second context — is **REFUTED**, two independent ways

The handoff's first candidate: `cleanupExpiredAwaitedRequests` (`coordinator.go:4348`) drives each
claimed retry under **one 60-second context shared across a batch of up to 25**, so a continuation
that spawns agents cannot finish inside it. The code is exactly as described — one
`context.WithTimeout(context.Background(), 60*time.Second)` per tick, `ClaimExpiredAwaitedRequestsForRetry(ctx, s.podName, 25)`,
then `for _, awaited := range claimed { s.retryExpiredAwaitedRequest(ctx, awaited) }` **serially**,
`cancel()` at the end. The claim is global (no pod scoping in the `WHERE`), so a rescuer really can
take 25 of anyone's rows.

**But the budget is never actually shared.** `ClaimExpiredAwaitedRequestsForRetry` stamps
`processing_started_at = NOW()` on every row of one batch, and `NOW()` is transaction-start time —
so rows of one batch carry a **byte-identical** timestamp and a batch is directly countable:

```sql
WITH b AS (SELECT processing_started_at, processing_pod, count(*) n
             FROM awaited_requests WHERE processing_started_at IS NOT NULL GROUP BY 1,2)
SELECT n AS batch_size, count(*) AS claims FROM b GROUP BY 1 ORDER BY 1 DESC;
--  batch_size | claims
--           1 |  31548      ← and NOTHING else
```

`[MEASURED 2026-08-19]` **31,548 claims over the full 7-day `awaited_requests` window, every one a
batch of exactly 1.** Never 2, never 25 — including across the 08-17 burst. Each claimed request
gets the whole 60 seconds to itself.

**And it does not use them.** Rebuilt from the burst population (§3): the next iteration's spawn is
sent **3–19 s** after the claim, and the parent's last write lands **9–16 s** after that. The
continuation dies **~12–35 s into a 60-second budget** — less than half way. A deadline that has not
expired cannot be what killed it.

> **⚠ My FIRST attempt at this measurement was wrong, and the way it was wrong is the lesson.**
> I first sized batches by bucketing `sent_at` by minute for `retry_version >= 1`, and got batches
> of 2–3 spanning **53–54 seconds** — which fits a 60-second budget so well I had already written
> the arithmetic down. **Two things were wrong with it.** (a) It is a **survivorship filter of
> exactly the kind this lane has been bitten by before**: a row that exhausts at rv=3 never resends,
> so `sent_at` never moves, so **the rows on the wedge path are the very rows the measure cannot
> see**. (b) The control killed it — co-timed rows have **different `processing_pod`s belonging to
> different agent deployments** (`agent-build-dispatch-loop-…`, `agent-chassis-…`,
> `agent-diagnose-orchestrator-…`), i.e. independent pods retrying independently inside the same
> minute, not one serial batch. `processing_started_at` has neither flaw: it is stamped at claim
> time, it is not reset on resend, and it is non-null on 31,548 of 31,961 rows (**98.7% coverage**).
> **The near-miss is the point: a 53-second span under a 60-second cap is a number I wanted, and I
> got it from a filter that had removed the population under test.**

### 3. The 08-17 evidence is **NOT gone.** `orchestration_states` lost it; `awaited_requests` kept it

This is the finding that matters, and it contradicts this lane's own handoff (mine to correct — the
08-19 handoff and the `bugs_open/029` entry both say the evidence expired and the 090 must wait for
the next burst).

`orchestration_states` retains ~26 h and **has** lost all 18 rows. `awaited_requests` retains
**7 days** — back to 2026-08-12 — and holds the complete per-step trace of the burst. The wedge
signature is fully reconstructible there **without `orchestration_states` at all**: an
`iter_N_call_handler` at `retry_version>=3 / status='error'`, the following
`iter_{N+1}_spawn_handler`, and the absence of any `iter_{N+1}_call_handler` row.

`[MEASURED 2026-08-19]` that query returns **20 instances** on 2026-08-17 — **two more than the 18
`orchestration_states` ever showed** — and `23eb0107`, the worked example in the 08-18 notes, is
among them. **20 of 20 have `next_call_registered = false`.** No exceptions.

| property | value across the 20 |
|---|---|
| claim (rv3 error) → next spawn sent | **3–19 s** |
| spawn `sent_at` → spawn `processed_at` | **9–16 s**, 37/37 rows |
| duplicate spawn registered | **17 of 20**, gap **06:54–09:37** |
| `iter_{N+1}_call_handler` ever registered | **0 of 20** |

**The 090 is therefore filable today.** The blocker the last run refuted on — *"no currently-stuck
instance … the 2026-08-17 burst has already aged out of retention"* — was true of the table it
looked in and false of the fleet. The symptom must point the loop at `awaited_requests`, and must
say plainly that `orchestration_states` is empty by design so the loop does not re-refute on the
same absence.

### 4. Where the parent actually dies — narrowed to one transition

The 9–16 s figure the 08-18 notes recorded as *"the parent's last state write lands 9–16 s after the
final send"* and the 9–16 s I measure as *spawn `sent_at` → `processed_at`* are **the same event**.
`handleCompleteResponse` stamps `processed_at` **before** `continueExecution` runs (stated in the
`claimRecoveryStaleness` comment at `coordinator.go:4409`). So:

1. spawn sent (claim + 3–19 s)
2. the `page-rerender` child answers the init handshake 9–16 s later — **37/37 spawn rows are
   `processed`**, so the child answered and the parent handled it
3. the parent marks the row processed and writes state — **this is the last write**
4. the parent must now execute `iter_{N+1}_call_handler`. **It never registers it, and never
   writes again.**

**So the death is inside `continueExecution` for the `call_handler` step, on the response-consumer
goroutine, immediately after the spawn reply was handled.** Not in the ticker, not in the spawn,
not in the retry path. That is a much smaller target than "somewhere after the handshake".

Two things about that goroutine, both `[VERIFIED at source]`, neither yet a cause:

- **The response consumer is strictly serial and synchronous** — `platform/agentbase/client.go:77`,
  `// Process synchronously to avoid race conditions / Don't use goroutine here`. Whatever
  `continueExecution` does inline, the pod processes **no further responses** until it returns.
- **It carries no deadline.** `c.ctx` is the agent-lifetime context, so nothing on this path can be
  killed by a context timeout — which is the third independent reason candidate 1 is dead.
- The message is **committed only on success**, so a death here leaves the response uncommitted.

`[UNVERIFIED]` and NOT claimed as the cause — the shapes worth testing next, in order:
1. **What is different about an iteration entered from the error path**, given `call_handler`
   registers normally on every healthy iteration of the same orchestration (`23eb0107` did it three
   times before wedging). `skipToNextLoopIteration` reaches the spawn through
   `createContinuationContext`, and the spawn's own registration goes through
   `persistAwaitingStateWithRetry`, which **reloads fresh state and copies only awaited entries +
   status onto it** — the discard documented in `LANDMINES.md`. Whether the `iter_N_error` /
   `error_count` keys it just wrote survive that park is a question with a definite answer I have
   not yet obtained.
2. Pod death at that instant (OOM/crash) would produce an identical trace. Untestable now — the
   pods are two days gone — but **cheap to settle on the next burst**, and worth adding to the
   capture.

### 5. ⚠ A check that CANNOT test this, and looks like it refutes it

"Did `build-dispatch-loop` ever log a `context deadline exceeded`?" — it never has (the fleet has
thousands, dominated by `vet-practice-verifier scrape_website`; `build-dispatch-loop` has **zero**).
That reads as a clean refutation of any context hypothesis. **It is blind.** `agenterrors.Write`
issues `db.ExecContext(ctx, …)` on the **caller's** context, so on any path where the context is the
thing that died, the error-log write dies with it. The absence is the *expected* reading under the
hypothesis, not evidence against it. Candidate 1 is refuted by §2, on measurements that could have
come out otherwise — not by this.

### 6. A source-level defect found while waiting on the 090 — the park's arrival check is keyed by STEP NAME, not request id

`persistAwaitingStateWithRetry` (`coordinator.go:2103`) opens each attempt with an "did the reply
beat the park?" check. It is keyed on the **step name**:

```go
for reqID := range state.AwaitedRequests {
    if existingData, exists := freshState.CollectedData[state.AwaitedRequests[reqID].StepName].(map[string]interface{}); exists {
        if _, hasResponse := existingData["response"]; hasResponse {
            return nil   // early — WITHOUT persisting status or the awaited entry
        }
    }
}
```

`response` is `awaitedResponseMarker`, and its own doc comment says it is what `applyResponseToState`
writes when a reply lands, read here "to decide a reply beat the park".

**It cannot distinguish "a reply to THIS request beat the park" from "a reply to a PREVIOUS request
on the same step name landed nine minutes ago."** So on any **re-registration of a step name that
has already been answered once** — exactly what the takeover does when it re-runs
`iter_N_spawn_handler` — the check fires spuriously and returns `nil` **without saving**. The caller
treats `nil` as success and goes on to `InsertAwaitedRequest`, so the row lands in the **table**
while the orchestration's JSONB never records the awaited entry and the status is never moved to
`AWAITING_RESPONSES`.

`[VERIFIED at source]` on the defect. `[UNVERIFIED]` as the wedge's cause, and **it does not fit as
the FIRST failure** — the ordering says the parent already failed to advance after spawn #1, whose
park has no prior marker to trip on. Recorded because it is real on its own terms, and because it is
a plausible reason the *second* attempt cannot recover what the first one lost. Whoever fixes it:
key the check on the **request id**, which is what the question actually asks.

Note this is adjacent to but distinct from `bugs_open/236` / RFC_012 (a), whose fix
(`carryCollectedDataOntoFreshState`, committed **2026-08-15**, `3ba384c63`) addressed the *discard*
in the same function and left this check untouched.

### 7. ⚠ Marker correction on my own §4, made before anyone else has to make it

I wrote that the spawn row's `processed_at` and "the parent's last state write" **are the same
event**. That is stronger than what I measured. What is `[MEASURED]`: the lane recorded a 9–16 s gap
between the final send and the last `orchestration_states` write; I measure a 9–16 s gap between
`sent_at` and `processed_at` on 37/37 spawn rows in `awaited_requests`. What is `[VERIFIED at
source]`: `handleCompleteResponse` stamps `processed_at` before `continueExecution` runs. The
identity of the two events is `[INFERRED]` from those three facts — well supported, not observed.
The narrowing in §4 survives either way, because both readings put the death after the reply was
handled; but the word "same" was doing work the evidence had not paid for.

### 8. The re-filed 090 came back **UNVERIFIABLE** too — but for a completely different reason, and it is OURS, not the symptom's

Run corr `d8af5f78-98bd-46fa-85b0-2a6899617db8`, filed 11:05Z, terminal ~11:10Z, **1 iteration**,
`status = UNVERIFIABLE`, no `stopped_reason`, `Citations: null`.

**The symptom correction WORKED.** The loop's own `RuntimeSite` reads
*"build-dispatch-loop orchestration — **awaited_requests** / response-consumer goroutine"*, and its
`NextScope` is a sensible eleven-symbol walk of the response path (`ProcessResponse`,
`applyResponseToState`, `advanceToNextStepWithRetry`, `saveStepResultWithRetry`,
`carryCollectedDataOntoFreshState`, …). It did **not** re-refute on absence of instances, which is
exactly what the morning's run did and what the correction was for.

**It stalled on a harness gap instead.** Its first `DataRequest`, in its own words:

> *"`awaited_requests` **is not in the bundle's schema listing**; its columns must be confirmed
> before a targeted row query can be written to check for the stuck `spawn_handler` awaited
> request."*

So the loop knew which table to read, could not discover its shape, asked for it, and the run ended
at iteration 1 without the answer. **The evidence bundle does not carry `awaited_requests`.** That
blocks *any* diagnosis of this class — every 029-shaped question lives in that table — and it is a
defect in our diagnosis harness, not in the bug, the symptom or the tree.

> **⚠ CORRECTION to §3 of this entry, and to what I put in the handoff and `bugs_open/029` an hour
> ago.** I wrote that the 090 "is filable today". That is true and it is not sufficient: it is
> filable and it will **stall**, because the loop cannot see the schema of the only table that holds
> the evidence. Two runs have now died on `awaited_requests` — the first because it was told the
> table did not matter, the second because it could not read it. **Fixing the symptom text was
> necessary and was not enough.**

**What actually unblocks this**, in order of cost: get `awaited_requests` into the diagnosis
bundle's schema listing (the durable fix — it is a first-class orchestration table and its absence
is arbitrary); or inline the `\d awaited_requests` output into the symptom text so iteration 1 does
not have to ask; or answer the data request and resume. **Do not re-file the same symptom unchanged
a third time** — it will stall in the same place, and that would be the third run spent on one
avoidable gap.

### 9. My bundle fix does NOT close the class — stated before anyone can read it as if it did

Having added `awaited_requests` to `schemaAlwaysTables` (`0132a3683`), I asked the loop's own
history which OTHER tables it has been unable to see. Two hits, both 2026-08-19 — mine, and a
different lane's run `dd61df1b`, which asked for:

```
workflow_templates, workflow_contract_chain, v_active_workflows, v_all_workflows
```

**None of those is covered by my change.** And the reason they are missing is worth more than the
list: the relevance include is applied as `table_name ILIKE $n` — a **prefix** match — so the
pattern `flow%` **does not match `workflow%`**. Whoever wrote `flow%` very likely meant workflow
tables to be in scope; they never have been. The two `v_` views miss for the same reason (the filter
has no `table_type` clause, so views are eligible once a pattern matches them — they simply never
match).

`[MEASURED 2026-08-19]`, and the count is a **floor set by the instrument**: `orchestration_states`
retains ~26 h, so this census structurally cannot see the `074beb8a`/236 runs that made the same
complaint, and it will not see today's by tomorrow. **Two distinct lanes hit it in one morning** —
that rate is the finding, not the total of two.

**So: one table fixed, class open.** The evidenced follow-up is to widen `defaultSchemaInclude` to
cover `workflow%` and rule on `v_%`. I have deliberately NOT folded that into the in-flight council
round (`e03f7122`) — changing the diff under a running review desyncs the submission from the code,
and the widening is a different lane's symptom that deserves its own rationale. Recorded in 016b §9
so it does not depend on this file being read.

### 9. The duplicate registration is at `retry_version=0` on BOTH rows — so the step BODY ran twice; and `allDone` is computed from a map that a known code path declines to write

`[MEASURED]` 2026-08-19 ~16:00Z, against the 20 retained instances in `awaited_requests`
(the 7-day table; `orchestration_states` holds none of them, by retention). Query: an
`iter_N_call_handler` row at `status='error'`, the following `iter_{N+1}_spawn_handler` rows for
the same `orchestration_id`, and a count of `iter_{N+1}_call_handler` rows.

| | |
|---|---|
| instances | **20**, all 2026-08-17, first 14:52:29Z, last 18:52:22Z |
| two `spawn_handler` rows for the same step | **17 of 20** |
| one `spawn_handler` row | 3 of 20 |
| both duplicate rows `status='processed'` | **17 of 17** |
| **`retry_version` on every spawn row, duplicate or not** | **0** |
| `iter_{N+1}_call_handler` rows | **0 of 20** |

**The `retry_version=0` on BOTH halves of the pair is the new fact, and it changes the reading.**
The lane has been treating the duplicate as the retry/takeover path re-registering. It is not:
a retry bumps `retry_version`, and nothing here is above 0. Two rv0 registrations of one step name
mean **the step body executed twice**, which is a different question — loop expansion, the recursive
`continueExecution`, or a second consumer taking the same work — not the retry clock.

> ⚠ This **downgrades** §6's framing. §6 said the arrival check "fires spuriously on any
> re-registration of a step name, exactly what the takeover does when it re-runs
> `iter_N_spawn_handler`". The mechanism at source is unchanged and still `[VERIFIED at source]`,
> but **the takeover is not what re-runs it here** — `retry_version` says so on all 37 rows.
> Whatever runs the step twice is upstream of the retry machinery. Corrected here rather than in
> §6 so the sequence of belief survives.

**Two more source facts fetched today, both `[VERIFIED at source]`:**

- `coordinator.go:2671-2678`, `handleCompleteResponse` — `allDone := len(freshState.AwaitedRequests) == 0`,
  computed from the **state's JSONB map alone**. That single boolean decides whether `CurrentStep`
  advances to `NextStep`, whether `Status` goes back to `EXECUTING_STEP`, and whether
  `continueExecution` is called at all. **The `awaited_requests` TABLE is never consulted here.**
- `coordinator.go:848-852`, `continueExecution` — an unconditional early `return nil` when the
  loaded `state.Status == StatusAwaitingResponses`, logged at Info and otherwise silent.

**So the shape of the class, stated as a claim and not yet as a cause:** the set of outstanding
awaited requests has **two representations and nothing reconciles them** — the table (which the
response consumer processes rows from) and the JSONB map (which every advance/park decision is
computed from). §6's arrival check is a code path that makes them diverge *by returning success
without writing the map*; `allDone` is the path that turns a divergence into a wrong advance-or-park
decision; and the `continueExecution` early return is what makes the result silent instead of loud.

`[UNVERIFIED]` that this composition is what wedged the 20. It is filed, not concluded — 090 run
correlation **`be89750f-c2ab-4c0a-bc21-751e75d9b19b`**, dispatched 16:07Z with the
`awaited_requests` schema **inlined into the symptom text**, which is what the previous two runs
died for want of. It cleared `assemble_bundle` and reached `lookup_symbols`, so the harness gap
that killed run `d8af5f78` is not blocking this one.

**Misstep on the way in, now a LANDMINE entry:** two dispatches were lost before this one because
the 090 trigger interpolates the symptom into a `$json$` psql literal **unescaped** — a newline
fails with `0x0a must be escaped`, a double quote with `Token "…" is invalid` — and it prints
`SAVE: CORRELATION_ID=…` *before* the insert runs. Both times the banner looked like a successful
dispatch and no work item existed. Flatten and de-quote first; the honest positive signal is the
later, differently-named `SAVE: RUN_CORRELATION_ID=` line.

### 10. Third stall, third distinct reason — and this one names a REAL framework gap: the bundle lists `awaited_requests`'s schema but renders NONE of its rows

> **CORRECTED 2026-08-19 (§17c):** the gap is real, but the remedy named here — a `### awaited_requests`
> **rows** section — **would render empty**. Static row sections are scoped to the diagnosis TARGET,
> whose correlation has 0 such rows (control: 1,469 in the incident window). The schema half, already
> shipped, was the whole blocker: the loop now fetches the rows itself via `data_requests`. Read §17c
> before building it.

Run `be89750f-c2ab-4c0a-bc21-751e75d9b19b`, dispatched 16:07Z, terminal ~16:12Z, **UNVERIFIABLE**.

**First, what WORKED.** Inlining `\d awaited_requests` into the symptom did its job, and the proof is
in the loop's own output: its `data_requests` SQL is
`SELECT orchestration_id, step_name, retry_version, status, sent_at, timeout_at, processed_at,
target_agent_type FROM awaited_requests WHERE correlation_id = '…' ORDER BY sent_at` — **every column
name correct**, which it could only get from the inline. `d8af5f78` could not write that query at all.

> ⚠ **Correction to something I said an hour ago in this session.** I checked the bundle, found
> `awaited_requests` in it 8 times, and read that as "the schema reached the bundle". It did not —
> `0132a3683` is Go and is **inert until the next roll**. Those 8 occurrences were **my own symptom
> text echoed back**. The inline is what worked; the committed fix has still never been exercised.
> A string being present in an artefact does not tell you which input put it there.

**Why it stalled anyway,** in the loop's own words: *"The bundle contains zero rows for
build-dispatch-loop … Without an orchestration_id or correlation_id known to belong to a frozen
build-dispatch-loop instance, this cannot be pulled from the current bundle."*

**And it is right, structurally.** `diagnose_load_runtime_action.go` renders three row sections —
`### agent_error_log` (:274), `### site_work_items` (:309), `### orchestration_states` (:344). There
is **no `### awaited_requests` section**. So:

| | `orchestration_states` | `awaited_requests` |
|---|---|---|
| in `schemaAlwaysTables` (columns described) | yes | **yes, since `0132a3683`** |
| **rows rendered into the bundle** | **yes, :344** | **NO** |
| retention | ~26 hours | **7 days** |

**That table is the finding.** For any hang older than about a day, `orchestration_states` is empty
by retention and `awaited_requests` is the only table still holding the incident — and the bundle
renders rows from exactly the one that is empty. The `0132a3683` comment block (:784–795) says this
in terms: *"the per-request rows … that any orchestration-hang question is actually answered from …
routinely the ONLY table still holding the incident."* The author understood it and still fixed only
the schema half. **Describing a table's columns while carrying none of its rows is a bundle that can
write a perfect query and never run it** — necessary, not sufficient, for the third time on this
lane in one day.

**Fix candidate, framework-wide (not filed as a bug yet):** an `### awaited_requests` rows section in
`diagnose_load_runtime`, keyed as `orchestration_states` already is. Ranked by CLAUDE.md's
close-the-door rule this beats any per-symptom workaround, because the workaround (paste the ids)
has to be remembered by every future filer of an orchestration-hang symptom, and the ones who forget
get a bundle that looks complete.

**Interim, and it is only interim:** re-filed as `d52c3407-14e7-4b9e-be46-c8ee741b2532` (16:14Z) with
six frozen `orchestration_id`s **and a same-day healthy control** (`002141d6-8964-47a9-8a45-063da7994aed`,
1 spawn + 1 call per iteration) named in the symptom, plus the instruction to query by
`orchestration_id` and **not** by `correlation_id` — the mistake the previous run made unprompted,
because the only correlation it had was the diagnosis's own.

**Tally for `WRONG_CALLS`, three stalls, three causes, all ours:** (1) pointed at the wrong table;
(2) right table, no schema; (3) right table, right schema, **no rows and no ids**. Each fix was
necessary and none was sufficient, and each time the previous fix's success made the next gap look
like the same failure. The general shape: **when a run refutes on absence, establish WHICH absence
before re-filing** — the loop's `needed_evidence` says so precisely every time, and reading it as
"still not enough evidence" rather than as a specific unmet precondition is what cost the repeats.

### 11. The fourth 090 is the first one to READ the wedge — verdict UNVERIFIABLE, and it is a real result, not a stall

Run `d52c3407-14e7-4b9e-be46-c8ee741b2532`, filed 16:14Z with six frozen `orchestration_id`s and a
healthy control seeded into the symptom. **`UNVERIFIABLE`, with four citations — three of them
`tier: state`, quoting real rows.** The seeding worked: it pulled `04518118`'s rows, aggregated them,
and pulled the control for comparison.

**What it confirmed, in its own citation:**

> `04518118… | process_item_iter_1_call_handler | 1 | {3} | {error}`
> `04518118… | process_item_iter_2_spawn_handler | 2 | {0} | {processed}`
> control `002141d6… | process_item_iter_2_call_handler | 1 | {0} | {processed}`
> control `002141d6… | process_item_iter_2_spawn_handler | 1 | {0} | {processed}`

**Why it abstained, and it is right to.** The bundle carries no bodies for `handleCompleteResponse`
or `continueExecution`, and `orchestration_states` — the only place that could show the map missing
an entry the table has — is gone by retention for all six. So the row signature is *compatible with*
the hypothesis and is not an observation of the mechanism. **A cite-or-abstain verdict that abstains
on exactly that distinction is the loop working, not failing.**

#### 11a. It raised a real alternative against my own correction — and here is the check

Its `needed_evidence` asks for the retry-claim code *"to rule out a legitimate retry path resetting
`retry_version` to 0 on a fresh `request_id`, which would explain the duplicate spawn rows without
the map/table divergence."* That would falsify §9 and the `WRONG_CALLS` entry built on it. **Checked
at source rather than defended:**

| | |
|---|---|
| every `INSERT INTO awaited_requests` in the tree | **two**, `state.go:1611` (`InsertAwaitedRequest`) and `spawn_actions.go:166` (`preRegisterAwaitedRequest`) — **neither is on a retry path** |
| every `retry_version` writer | **three**, all `UPDATE … WHERE request_id = …`: `coordinator.go:3329` (commented out), `coordinator.go:3341` (`UpdateAwaitedRequestRetry`, live), `state.go:1965` (`UpdateAwaitedRequestForRetry`) |
| `UpdateAwaitedRequestForRetry` (`state.go:1962`) | **DEAD — zero callers in the tree.** The live writer is `UpdateAwaitedRequestRetry` at `coordinator.go:3337` |

**A retry mutates the existing row in place and cannot mint a second one.** So two distinct
`request_id`s at rv0 on one step name are not reachable from the retry driver.

**And the loop's own citation already proved it, inside a single orchestration:** `04518118`'s
`call_handler` is **one row at rv3** (a retry, bumped in place) while its `spawn_handler` is **two
rows at rv0 with distinct ids** (two registrations). The contrast is within one orchestration on one
day, so no cross-instance confound. **§9 and the `WRONG_CALLS` entry stand.**

> ⚠ Note for anyone reading Fable's plan doc alongside this: it reached the same conclusion but
> cited **`state.go:1962`**, the function with no callers. Right answer, wrong citation — the live
> writer is `coordinator.go:3337`. `[VERIFIED]` by `grep -rn 'UpdateAwaitedRequestForRetry'`
> returning only its own definition and doc comment.

#### 11b. What the loop wants next, and it is cheap

`next_scope` names `handleCompleteResponse`, `continueExecution`, `ClaimAwaitedRequestForRetry`,
`UpdateAwaitedRequestForRetry`, `retryExpiredAwaitedRequest`. Three of those are already read and
recorded here; the seed scope on the next run should carry the **function bodies**, which is what its
abstention actually turned on. **But note the ceiling:** the map-vs-table divergence it says it needs
is only observable in `orchestration_states`, which is purged for the 08-17 cohort and will be purged
for any future one within ~26 h. **RSH-011's hourly capture is what closes that** — it snapshots the
live wedge with its full `awaited_requests` set. So the honest sequence is: *the next burst is what
confirms this, and the capture job is already armed for it.* Nothing further is owed to the loop
until then.

### 12. C1/C2 survive the one test the table could have killed them with — 17/17 gaps clear the 5-minute takeover threshold

`[MEASURED]` 2026-08-19 ~16:50Z. All duplicate `iter_N_spawn_handler` pairs on 2026-08-17:

| | |
|---|---|
| duplicate pairs | **17** |
| gap `sent_at(2) − sent_at(1)` **> 300 s** | **17 of 17** |
| min / max gap | **414 s / 577 s** |
| pairs whose two rows share one `processing_pod` | **17 of 17** |

`StuckOrchestrationTimeout = 5 * time.Minute` (`coordinator.go:38`). The stuck-orchestration takeover
(**C1**, `handleOrchestrationStatus`'s `StatusExecutingStep` arm, `:761-775`) fires on an inbound
message when `LastActivity` is older than that, clears the executing step and re-runs `CurrentStep`
from scratch — a fresh execution, hence rv0 and a new request id. Its prediction is *gap = 300 s + the
trigger message's cadence*. **414–577 s is 300 s plus 114–277 s**, and **nothing lands under 300 s.**

> **This is a SURVIVAL, not a confirmation, and the difference matters.** A single pair under 300 s
> would have refuted C1 outright; none is, so C1 passed a test it could have failed — that is the
> disconfirmable-measurement bar this lane keeps failing to clear, cleared. But **C2** (the
> `StatusRunning` arm, `:782-801`) predicts an *identical* table signature, and nothing in
> `awaited_requests` separates them. So the honest state is **C1 and C2 both survive; C3 refuted at
> source (§11a); C4 and C5 disfavoured** (row 1 is `processed` in every pair; the gaps are minutes,
> not seconds).

⚠ **The same-pod figure is weaker than it looks — do not quote it as "one pod re-executed the step".**
`processing_pod` is stamped by the consumer that **claimed the reply**, not by whatever sent the
request. So 17/17 says both *replies* were claimed by one pod, which on a 2-replica deployment is
unremarkable and says nothing about who re-ran the step body. It grounds the draft plan's
`[UNVERIFIED]` 6/6 sample at 17/17 and is recorded for completeness, not as evidence for C1.

**Separating C1 from C2 needs logs**, and the 08-17 pods are two days gone — so it waits for the next
burst, where RSH-011's hourly capture is already armed. Nothing further is extractable from the table
on this question.

### 10. Blast radius of the `workflow%` follow-up, measured before proposing it — and a source figure I could NOT reconcile

`[MEASURED 2026-08-19]`, live DB:

| quantity | value |
|---|---|
| `schema_table_cap` (default) | **120** (`diagnose_load_runtime_action.go:440`) |
| tables the current include matches | **86** base tables + 1 view |
| `schemaAlwaysTables` | 7 (they sort FIRST, so truncation cannot reach them) |
| **added by `workflow%`** | **2** |
| added by `v\_%` | 11 views |

So the widening is safe on the cap: ~94 in use today, `workflow%` takes it to ~96, and even adding
every `v_` view lands ~107 — all under 120. **Worth stating because the obvious objection to
widening a relevance include is that it pushes something else out, and here it measurably does not.**

> **⚠ A figure in that file's own comment does NOT reconcile, and I am recording the discrepancy
> rather than quietly overwriting it.** The comment says *"Measured 2026-08-10 against the live DB:
> the default include (site%|page%|content%|flow%) selected **26 of 433** public tables"*. The same
> four patterns today select **86 of 457** base tables. Nine days is not obviously enough for a
> 3.3× jump in matches against a 5% growth in tables, so this is **either real growth in `site_*` /
> `page_*` tables or a method difference** (the original may have counted post-exclude, or counted
> only tables actually rendered). **I did not resolve it, and I am not calling the comment stale on
> a measurement I cannot show is like-for-like** — the honest state is "two numbers, one method
> unknown". Whoever widens the include should settle it in passing, because the cap headroom
> argument above depends on the larger number being the right one.

**Held deliberately, not forgotten.** I have not shipped the `workflow%` change in this session, for
one reason: run `d02a6958` is in flight, and if it needs further tables the right move is **one**
widening that covers everything rather than two council rounds a day apart. If the run comes back
clean, submit the widening on its own evidence — and per the 2026-07-29 ruling, tell the lane whose
run `dd61df1b` stalled, since it is their symptom being fixed.

### 11. The bundle fix is **BEHAVIOURALLY PROVEN** on `v1.0.1316` — before/after, with a positive control

`0132a3683` is an ancestor of build point `07eeba4a1` (present on both replicas, previous build
point absent), so the fix is aboard. That is *shipped*. This is *proven*:

```sql
SELECT left(correlation_id,8) AS run, iteration, created_at::timestamp(0),
       (body LIKE '%awaited_requests(%')     AS schema_renders_it,
       (body LIKE '%orchestration_states(%') AS control
  FROM diagnosis_artifacts WHERE kind='bundle' AND correlation_id IN (…three runs…) ORDER BY created_at;
```

| run | when | `awaited_requests(` | control `orchestration_states(` |
|---|---|---|---|
| `b346d0d4` iter 1–3 | 09:47–09:51 (pre-fix) | **f, f, f** | t, t, t |
| `d8af5f78` iter 1 | 11:06 (pre-fix) | **f** | t |
| **`d02a6958` iter 1** | **20:38 (post-fix)** | **t** | t |

And the rendered line is the Schema section's own format, carrying the columns the earlier run had
to ask for and could not get:

```
awaited_requests(request_id varchar, orchestration_id uuid, correlation_id varchar, step_id varchar,
 step_name varchar, retry_version integer, target_agent_id varchar, target_agent_type varchar,
 responses_topic varchar, requests_topic varchar, sent_at timestamp, timeout_at timestamp,
 reply_to_request_id varchar, status varchar, processed_at timestamp, processing_started_at timestamp,
 processing_pod text, request_payl…
```

**Why the control is load-bearing.** `orchestration_states(` is `t` in **all five** bundles, so the
Schema section was rendering perfectly well before — the absence was specific to my table, not a
broken section. Without that column the before/after would be equally consistent with "the section
was empty in those runs".

> **⚠ My FIRST check was blind and I nearly recorded it.** I began with
> `body LIKE '%awaited_requests%'` and `body LIKE '%awaited_requests%retry_version%'`, both of which
> returned **true** — and both are worthless, because **the symptom text I wrote names that table
> and those column names**, and the symptom is quoted inside the bundle. The discriminating form is
> `awaited_requests(` with the opening parenthesis: that is the *renderer's* syntax, which no prose
> in my symptom can produce. **A check that my own input can satisfy is not a check.**

`[MEASURED 2026-08-19 20:40Z]`. Register note: this is the behavioural half of `0132a3683`
(council-approved round 1, corr `e03f7122`).

**Runtime detail worth carrying:** the live bundle reports *"33 of 479 public tables are shown"* —
so the include actually in force at runtime is much narrower than the package default I measured
(86 base tables). Whoever proposes the `workflow%` widening should read the live agent config's
`schema_include_patterns`, not the Go default, before arguing about the cap.

### 12. The "~7 days / good until 08-24" figure, grounded at SOURCE instead of observed

The whole remaining plan rests on the 08-17 rows surviving long enough to diagnose, and until now
that rested on one observation (`min(sent_at)` = 08-12). The rule is DB-resident, in
`cleanup_expired_awaited_requests()` — the function `state.go:2055` calls every minute:

```sql
-- Delete very old terminal requests (>7 days)
DELETE FROM awaited_requests
WHERE status IN ('processed','expired','cancelled','error')
  AND processed_at < NOW() - INTERVAL '7 days';
```

`[VERIFIED at source, 2026-08-19]` via `pg_get_functiondef`. Three things this pins down that the
observation could not:

1. **It is keyed on `processed_at`, not `sent_at`.** My estimate used `sent_at`. For the wedge rows
   the two differ by 9–16 seconds, so the deadline is unchanged — but the *predicate* is now known,
   and anyone recomputing the window must use `processed_at` or they will be wrong on a slow row.
2. **Only TERMINAL rows are deleted.** The wedge population is all `processed` / `error`, so it does
   expire. A row that never reached a terminal status is kept indefinitely — so "7 days" is the rule
   for this evidence, not for the table.
3. **The deadline is therefore `processed_at + 7 days` ≈ 2026-08-24** for the 08-17 rows, and it is
   enforced continuously by a per-minute call, not by a nightly job — there is no "it might not have
   run yet" grace.

> **⚠ Do not confuse this with `cleanupExpiredAwaitedRequests` (Go, `coordinator.go:4348`).** The Go
> ticker is the *caller*; the retention is entirely inside the SQL function. Reading the Go and
> concluding "the cleanup only marks rows expired" is wrong — the DELETE is one statement further
> down, in a place `grep -rn "DELETE FROM awaited_requests" platform/ --include=*.go` returns
> **nothing** for. **That grep returning zero is why this was assumed rather than checked.**

### 13. Pre-registering how to read run `d02a6958`'s outcome — written BEFORE the verdict, deliberately

`[MEASURED]` the `diagnose-agent` cap is **5 iterations**
(`jsonb_path_query_first(default_config,'$.**.max_iterations')` = 5). At the time of writing this
run has produced bundles for iterations 1–3 (44,963 → 103,699 chars, both carrying the
`awaited_requests` schema) and has gone `verdict → route` twice, i.e. it is **entering iteration 4**.
So it may terminate at the cap. Three outcomes and what each is worth — fixed now, so I cannot
choose an interpretation after seeing the answer:

| outcome | what it means | what it does NOT mean |
|---|---|---|
| **CONFIRMED** with citations | a cause, gradable against the code | not a fix — the citations still need checking against the tree, per this lane's own history of confident-and-wrong |
| **REFUTED** on a named mechanism | genuinely useful; it eliminates a candidate the way the shared-60s-context measurement did | not "the bug is not real" |
| **UNVERIFIABLE at the iteration cap** | **the most likely, and it is NOT the same result as the two earlier ones.** Those two stopped because they could not SEE the evidence — one was told it had expired, one could not read the table. This one had four bundles' worth and still could not converge. That is a statement about the **difficulty of the question**, and its `NextScope` at the cap is a real finding: it is where four iterations of narrowing pointed | not a failure of the fix, and **not evidence the wedge has no cause** |

**The thing to read in every case is `NextScope` and the last `Verdict`, not `status`.** This lane
has now spent three runs learning that the status field is the least informative part of the record.

**And the fix's own result is already banked and independent of this**: the bundle renders the table
(NOTES §11), proven against four pre-fix bundles with a positive control. Whatever `d02a6958`
returns, that does not change.

### 14. Run `d02a6958`: UNVERIFIABLE, and it is the best result this lane has had — it ran out of ROAD, not out of evidence

> **CORRECTED 2026-08-19 (§17b):** it did **not** run out of iterations — it stopped at **3 of a
> permitted 5**, halted by the narrowing guard. The field is `route.stopped_by` (not `stopped_reason`)
> and it reads **`scope-not-narrowing`**; `iteration-cap` is a separate value in the same vocabulary.
> The "one query short / give it a head start" remedy therefore does not address the halt. §17b.

**Outcome: `UNVERIFIABLE`, 3 iterations, `stopped_reason` empty.** My §13 pre-registration guessed
it would hit the **5**-iteration cap; **it did not — it stopped at 3.** Recording that my prediction
was wrong before drawing anything from the result.

**What is genuinely new: it read the recovered evidence and cited it.** The final verdict carries a
**Tier 1 citation, `Fresh: 2026-08-17`**, quoting an actual row:

```
838f8c14-5d49-4cd3-9432-489121f538c2 | process_item_iter_0_call_handler | 3 | error |
2026-08-17 18:47:15 | 2026-08-17 18:52:15 | 2026-08-17 18:52:15
   — awaited_requests (data_request: candidate retry_version-3 error call_handlers)
```

That is one of the twenty. **Two runs ago the loop could not see this table at all; now it is citing
its rows by hand.** The bundle fix is doing exactly what it was built to do.

Its Tier 0 citations are the right code, too — `skipToNextLoopIterationForAsync` →
`skipToNextLoopIteration` → `createContinuationContext` + `continueExecution`, and
`handleCompleteResponse`'s own continuation. That is the transition this lane narrowed to
independently in §4, reached by the loop from the other direction.

**Why it stopped, in its own words:** it had established the **precondition** (a `call_handler` at
rv3/error) and not the **outcome** (the following `spawn_handler` processed with no
`iter_{N+1}_call_handler` ever created), and it wrote the exact SQL it still needed — a filtered
query over the ten candidate `orchestration_id`s it had found. **It ran out of iterations one query
short of the answer.**

#### The trap it hit and RECOVERED from — and whose fault each half is

Iteration 1 asked for the reconstruction with `ORDER BY orchestration_id, sent_at` and **no LIMIT**.
The harness capped the result at **200 rows** (`row_cap`, default 200 —
`diagnose_load_runtime_action.go:260` and `diagnose_run_checks_action.go:99`). Alphabetical order
plus a cap returns the **lexicographically first** orchestrations only, and every one of the ten
candidates sorts past the `03…` range that fitted. In its words: *"the dump is silent on the
mechanism, not confirming or refuting it."*

**Apportioning this honestly:** the **cap is ours**, the **`ORDER BY` was its**, and **the harness
behaved well** — it announced the truncation, which is why the loop knew to route around it and did,
with a properly filtered query in iteration 2 that found the candidates. This is the estate's own
rule arriving from outside: *a capped census cannot say WHO was cut — read the `ORDER BY`.*

> **So do NOT record this run as "blocked by the harness".** The previous two were. This one was
> not: it had the evidence, cited it, self-corrected a truncation, and stopped one question short.

#### What would settle it, and it is cheap

Hand the next run the reconstruction **as SQL it can execute in iteration 1**, so it starts where
this one ended. Note the tension with the runbook's symptom-authoring rule (*"assert neither rows
nor counts — the loop fetches and cites them"*): the resolution is to supply **the query, not the
answer**. A pointer preserves the loop's independent grading; a stated count would not.

`NextScope` at the stop — where three iterations of narrowing pointed — is worth carrying:
`executeStep`, `executeLocalAction`, **`processAwaitResponse`**, **`createContinuationContext`**,
`ProcessResponse`, `HandleResponse`, `handleRecoverableError`, `createAwaitedRequest`,
`extractRequestID`, `executeRemoteAction`.

### 15. Run `5d1d8f1c` — my "improved" re-file made it WORSE, and I confounded the experiment so I cannot say which change did it

**Outcome: `UNVERIFIABLE`, `stopped_by = "scope-not-narrowing"`, ONE iteration.** The run it was
meant to improve on (`d02a6958`) managed **three** iterations and produced real Tier-1 citations from
the 08-17 evidence. **This is a regression, and it is mine.**

What it did do: it read the instruction and **issued my reconstruction CTE as its first
`DataRequest`** — so the "put the query in the symptom" mechanism works. It then stopped before
iteration 2, which is where the results would have arrived. **It never saw the output of the query I
supplied.** Its Tier-0 citations were good (`handleCompleteResponse`'s `allDone` continuation,
`continueExecution`'s `SetExecutingStep`) and its Tier 1 was the expected
*"(no orchestration rows for this correlation/site)"*.

#### The methodological error, stated plainly

**I changed three things in one re-file:**
1. embedded the reconstruction SQL in the symptom,
2. named the 200-row cap and its remedy,
3. **widened the seed scope from 5 symbols to 6**, adding the previous run's own `NextScope` picks.

The result got worse. **I cannot attribute the regression to any one of them**, and (3) is the
obvious suspect on the stated stop reason — `scope-not-narrowing` compares the proposed scope against
what it started with, and I handed it a *wider* start whose natural next step (13 symbols) then read
as no narrowing. But that is a **hypothesis about my own change**, not a measurement, and the
one-run-per-condition design cannot separate it from the longer symptom text.

> **This is the same error this lane has logged in other people's work and I made it anyway: an
> experiment with three simultaneous interventions has no interpretable result.** The correct
> re-file changes ONE thing. `WRONG_CALLS.md` 2026-08-19.

#### What the next re-file should do — one change only

**Keep `d02a6958`'s seed scope EXACTLY** (the original five symbols; do NOT add the NextScope picks)
and add **only** the reconstruction SQL. If that run also stops at `scope-not-narrowing`, the SQL is
implicated; if it reaches iteration 2 and reads the results, the seed widening was the cause and the
lesson is *never seed with the previous run's NextScope — that is what the loop is for.*

**And note what is NOT in doubt:** `d02a6958` remains the high-water mark and its findings stand —
the bundle fix works, the loop can now read and cite `awaited_requests`, and the wedge's transition
is narrowed. Today's regression is about how I drove the tool, not about the evidence.

**Two blind checks caught in one evening, same shape.** I twice checked the bundle for a string I had
authored (`awaited_requests`, then `next_call_registered`) and both returned true on material that
was just **my own symptom text quoted back into the bundle**. The general rule: **any check against
the bundle body for a string you wrote is blind.** Discriminate on the *renderer's* syntax
(`awaited_requests(` with the parenthesis) or on the query's *output values*, never on its name.


### 16. ⚠ WITHDRAWN: my duplicate-spawn account, and the numbering collision in this file

**Two corrections, both from discovering at ~22:15Z that a second session worked this lane all day
and I never checked (`scripts/who-owns.py`, `git log` on the lane dir — neither run).**

**(a) My takeover account of the duplicate spawn is WITHDRAWN.** I wrote that the 06:54–09:37
duplicate-registration gap was "consistent with the already-established >5-min takeover sampled by a
5-min replay cycle". The other session measured the thing that settles it: **`retry_version` is 0 on
all 37 spawn rows, and a retry bumps `retry_version`** — so the duplicate is not the takeover
re-running the step, **the step BODY executed twice**. Their sections in
`HANDOFF_2026-08-19b_continue_here.md` are primary on this.

**The part worth keeping is how I got it wrong: the refuting data was in my own output.** Every spawn
row I printed in §3 shows `retry_version 0`, hours before I wrote the takeover line. I read past it
because a 5-min threshold sampled by a 5-min cycle *explains* a 5–10 minute gap, and the gap was in
range. **An explanation that fits is not evidence, and a fit cannot disconfirm.** Same failure shape
as the 53-second batch fit in §2 — which I caught, in the same session, and then repeated.

**(b) This file now has TWO §9, §10 and §11 sequences**, written by two sessions appending
concurrently without seeing each other. Neither set is wrong; the numbers are. **Resolve by date and
subject, not by number** — mine run from ~10:45Z (retention, the ticker refutation, the bundle fix,
the two 090s); theirs cover the rv0 finding, the framework gap and the `d52c3407` verdict. Do not
renumber either: both are cited by number from other documents already.

**(c) Their ranking beat mine on the framework gap.** I fixed the *schema* half and then worked
around the *rows* half by putting a reconstruction query in the symptom — a workaround that must be
remembered by every future filer. They ranked building the `### awaited_requests` rows section above
any such workaround, on the close-the-door rule. **My workaround was then tried and it regressed the
run** (§15). Their ranking was right before the evidence arrived, which is the stronger position.

> **CORRECTED 2026-08-19 ~21:1xZ (third session, §17b):** the rows-section ranking endorsed here
> **would have shipped an empty section.** Measured, not argued — see §17b before building it.

### 17. A third session's read of `v1.0.1316` — one corroboration and two corrections

Written by the session handed `HANDOFF_2026-08-19b_continue_here.md` (**note: that file says it
supersedes `HANDOFF_2026-08-19_continue_here.md`, but the owning session is appending to `19`, so a
fresh session is pointed at the stale one** — reconcile before trusting either). Filed as a
contribution, not a takeover: `bugs_open/029 hung spawns` was live and busy throughout.

**(a) The bundle-fix proof below is CORROBORATION of §11, not a finding.** Independently and
without seeing §11, from `orchestration_states` over ten runs rather than `diagnosis_artifacts` over
five: the `awaited_requests(` table entry renders in `d02a6958` and in **none** of the nine pre-fix
runs (10:50–16:16Z), with `orchestration_states(` present in all ten as a positive control and a
`zzz_no_such_table(` negative control absent from all ten. Schema length 12,219 → 12,650 (+431),
consistent with one added entry. §11 got there first and sourced it better. **One detail §11's
warning does not name:** a bare `awaited_requests` match is blind for a *second* reason beyond the
symptom text — `orchestration_states` has a **column** called `awaited_requests jsonb`, which is
rendered in every bundle ever produced. My first census returned `t` for all ten runs on that alone.
Both blind paths are closed by the same discriminator, the renderer's `(`.

**(b) CORRECTION to §14: run `d02a6958` did NOT run out of iterations.** §14 records it as *"ran out
of iterations one query short of the answer"* with *"`stopped_reason` empty"*. The populated field is
`route.stopped_by` and it reads **`scope-not-narrowing`** — a convergence guard, at **iteration 3 of
a permitted 5**. `iteration-cap` is a *separate* value in the same vocabulary
(`diagnose_emit_action.go:24`), so this is a distinction the harness itself draws.

`[MEASURED 2026-08-19]`, `collected_data->'route'` on the `diagnose-agent` row:

| run | `stopped_by` | iter | `prev_scope_size` | named next_scope | guard: `>prev+2`? |
|---|---|---|---|---|---|
| `d02a6958` | scope-not-narrowing | **3** of 5 | 5 | 8 | `8 > 7` ✓ reconciles |
| `5d1d8f1c` | scope-not-narrowing | 1 | 7 | 9 | `9 > 9` ✗ **does NOT reconcile** |

Guard: single exit at `pkg/diagnose/loop.go:432`, predicate `next.size() > prevSize+2`; `size()` is
`len(Symbols)` (`:205`); `named` is built from `v.NextScope` **alone** — it does *not* union the
previous scope (`namedScope`, `:398-417`). On a stop the state returns early *before*
`st.PrevScopeSize = d.NamedScopeSize` (`advance.go:104-111` vs `:120`), so the persisted
`prev_scope_size` **is** the pre-guard number the guard compared against.

**Why this matters for the next filing.** §14's remedy — hand the run the reconstruction SQL so it
starts where this one ended — addresses the *data* half, which the loop had already solved for itself
with a filtered query in iteration 2. It does not touch the halt.

> ~~**[UNRESOLVED — and it cuts BOTH ways.]** `5d1d8f1c`'s numbers do not satisfy the only predicate
> that can emit this stop reason, so *why* it halted at iteration 1 is **not established**. That
> means §15's suspicion (*"adding the previous run's NextScope symbols did it"*), which
> `NEXT_090_single_variable.sh` now encodes as a hard rule, is **equally unestablished** — as §15
> itself says. I tried to derive "seed a wider scope" as the remedy and **could not make the
> arithmetic support it either; recording that I failed rather than the tidy version.** Settle the
> reconciliation before the one-variable run's outcome is read as evidence about seed scope, or it
> will attribute to the seed whatever this unexplained trip does next.~~
>
> **WITHDRAWN 2026-08-19 ~22:3xZ — `5d1d8f1c` DOES reconcile; the error was mine and it was the
> OPERAND.** Caught by the owning session's §18 within minutes of my flagging it, and re-verified
> here from the trail. I read `collected_data->'verdict'->'result'->'next_scope'`; the guard acts on
> the **per-iteration** `NextScope`, which lives in `route.diagnose_state.trail[i].Verdict.NextScope`.
> The two disagree — for `d02a6958` the stored `verdict` field held **iteration 1's** value (8), not
> the iteration that actually tripped (12). `[MEASURED]` from the trail:
>
> | run | seed | init prev (`seed+1`) | per-iteration `named` | trips |
> |---|---|---|---|---|
> | `d02a6958` | 5 | 6 | 8 → 5 → **12** | `12 > 5+2` at **iter 3** |
> | `5d1d8f1c` | 6 | 7 | **13** | `13 > 7+2` at **iter 1** |
>
> **The consequence reverses §15's direction, and that is the useful part.** `prevSize` starts at
> `seed.size()+1`, so a **wider seed RAISES the allowance** — widening is *protective*, not harmful.
> `5d1d8f1c` had the wider seed AND the higher threshold and tripped anyway, so **seed width cannot
> be the cause**; what differed is the model naming **13** symbols in iteration 1 where the baseline
> named 8. `[UNVERIFIED]` whether that is symptom length or run-to-run variance — one run per
> condition cannot separate them. And the baseline's margin was **one symbol**: `8 > 8` is false.
>
> **My §17b conclusion survives its own broken arithmetic** — `d02a6958` was still guard-stopped at
> 3 of 5, not iteration-capped — which is exactly why the operand error was worth catching: a right
> conclusion resting on a wrong number is the shape that propagates. `WRONG_CALLS.md` 2026-08-19.

**(c) CORRECTION to §10 and to §16(c): the `### awaited_requests` rows section would render EMPTY.**
§10 calls the missing rows a REAL framework gap; §16(c) ranks building it above the query-in-symptom
workaround on the close-the-door rule. The ranking argument is sound and the measurement still kills
it — the bundle already contains the natural experiment.

- Every static row section is scoped to the **diagnosis target**: `agent_error_log` by
  `site_id`/`domain` (`diagnose_load_runtime_action.go:279-280`), `site_work_items` by `site_id`
  (`:314`), `orchestration_states` by `correlation_id`/`site_id` (`:349-350`). A new section built to
  that pattern inherits the scoping.
- `[MEASURED]` the diagnosis target's own correlation `ac075f27-…` has **0** `awaited_requests` rows;
  control — the 08-17 incident window — has **1,469**. The zero is specific to the target, not an
  empty table. (The loop stated this itself and it checks out.)
- The `orchestration_states` section is *already* scoped that way, already renders
  `(no orchestration rows for this correlation/site)` for this run — and **that empty string is the
  verdict's own first citation.**

For this class of diagnosis — a *historical* incident described in prose, where the target
correlation is the diagnosis run's own — the new section renders `(no rows…)`. It would be built,
council-gated, shipped, and inert for the case that motivated it.

**What actually unblocked the loop was the schema half, already shipped.** With the columns present
the model wrote its own valid `awaited_requests` query and `runDataRequests` (`:624`) returned real
rows — 215 lines over 37 orchestrations, including Tier-1 citations of the 08-17 incident. The rows
arrive **dynamically**; they never needed a static section. The residual defect is real but much
narrower than a new section: an unfiltered dump meets `row_cap=200` and an alphabetical `ORDER BY`,
which §14 already documents and which the harness already announces.

### 18. 2026-08-20 — the scope guard reconciles for BOTH runs, which corrects my §14, my §15 and the peer session's §17

A peer session (contribution, `5022305cf`, their §17) flagged that my §14 was wrong and that
`5d1d8f1c` does **not** reconcile against the convergence guard. **The first half is right; the
second is not, and getting the real numbers reverses my §15 remedy.**

**Guard, all `[VERIFIED at source]` and all as the peer stated:** single exit
`pkg/diagnose/loop.go:432`, `if next.size() > prevSize+2`; `size()` is `len(s.Symbols)` (:205);
`namedScope` (:398) builds `Symbols` from `v.NextScope` **alone** — it copies Tables/RuntimeSite from
prev but does **not** union prev's symbols; init `PrevScopeSize = seed.size()+1` (`advance.go:68`);
and on a stop `advance.go` returns at :104-111 **before** :120 assigns
`st.PrevScopeSize = d.NamedScopeSize`, so **the persisted `prev_scope_size` is the pre-guard number.**

`[MEASURED 2026-08-20]` — `route.stopped_by` and `route.diagnose_state.prev_scope_size`, with
per-iteration `NextScope` lengths from the evidence trail:

| run | seed | init prev | iter 1 | iter 2 | iter 3 | persisted prev | stopped_by |
|---|---|---|---|---|---|---|---|
| `d02a6958` | 5 | 6 | named **8**: 8 > 8 **false, passes** → prev 8 | named **5**: 5 > 10 false → prev 5 | named **12**: 12 > 7 **TRIPS** | **5** ✓ | scope-not-narrowing |
| `5d1d8f1c` | 6 | 7 | named **13**: 13 > 9 **TRIPS** | — | — | **7** ✓ | scope-not-narrowing |

**Both reconcile exactly.** The peer read `d02a6958` as "prev 5, named 8, 8 > 7" — right conclusion,
wrong operand: 8 is *iteration 1's* NextScope, and the trip was at iteration 3 with named **12**. For
`5d1d8f1c` they used named 9; it is **13**, and 13 > 9 trips. So "5d1d8f1c does not reconcile" is
withdrawn — nothing here is unexplained.

#### What this reverses

- **§14 is WRONG and withdrawn.** `d02a6958` did not "run out of iterations one query short": it
  stopped at 3 of a permitted 5 on the **convergence guard**. The peer was right, and the field to
  read is `route.stopped_by` — the `diagnosis.stopped_reason` I queried is a different, empty key.
- **§15's remedy is WRONG IN DIRECTION, and the peer could not derive the opposite either.** The
  threshold is `prevSize + 2`, and `prevSize` starts at `seed.size()+1` — so **a wider seed RAISES
  the allowance and is protective.** Seed width cannot have caused `5d1d8f1c`'s trip; it had the
  *wider* seed and a *higher* threshold. `NEXT_090_single_variable.sh` asserted the opposite as a hard
  rule and has been corrected in place. **My §15 suspicion is not merely unestablished — the
  arithmetic contradicts it.**
- **What actually differed** is the model naming **13** symbols in iteration 1 where the baseline
  named **8**. Attributable to the longer symptom or to variance; **`[UNVERIFIED]` either way**, and
  one run per condition cannot separate them. Note how close the baseline was: `8 > 8` is false, so
  `d02a6958` survived iteration 1 **by exactly one symbol**.

> **The lesson is about the operand, not the guard.** Three of us — the peer, and me twice — reasoned
> about this halt from a number we had not fetched. The guard was always deterministic and always
> printed its inputs. `route.stopped_by` and `route.diagnose_state.prev_scope_size` were one query away.

#### And my §16(c) endorsement is WITHDRAWN

I endorsed building a `### awaited_requests` **rows** section. The peer showed it would render
**empty**: every static row section is scoped to the diagnosis **target**
(`diagnose_load_runtime_action.go:279-280, :314, :349-350`), and the target correlation has **0**
`awaited_requests` rows against **1,469** in the 08-17 window. The natural experiment is already in
the bundle — `orchestration_states` is scoped the same way, renders
*"(no orchestration rows for this correlation/site)"*, and **that empty string is a verdict's own
first citation** (I quoted it from `5d1d8f1c` without drawing the inference). For a historical
incident described in prose, a new section renders `(no rows…)`.

**So the schema half I shipped was the whole blocker, and the residual is much narrower than a new
section:** unfiltered dump + `row_cap=200` + alphabetical `ORDER BY`. The peer measured the loop now
writing its own query and `runDataRequests` returning 215 lines over 37 orchestrations.

**One addition to my §15 blind-check warning, from the peer:** `orchestration_states` has a **column**
named `awaited_requests jsonb`, rendered in every bundle ever — so a bare `LIKE '%awaited_requests%'`
reads true for a *second* reason beyond my symptom text. Same remedy: match the renderer's `(`.
