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
