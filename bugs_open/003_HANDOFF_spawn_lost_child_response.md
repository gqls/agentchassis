# HANDOFF — spawned child agents lose their response; parent orchestrations hang until reaped

**Purpose.** Start a fresh chat from exactly here to fix this. This is a **platform / infra
bug**, not a bug in any one site. It surfaced while rebuilding `leopardessconsulting.co.uk`
(see `docs/leopardessconsulting/RUNNING_NOTES.md` turn 17) but the evidence below shows it is
**fleet-wide**.

**Filed:** 2026-07-15
**Severity:** High — silently blocks any workflow that spawns a child agent and waits for its
response (image generation, page content builds, dispatch-loop item handlers). Work is lost or
delayed by 30–90 minutes per occurrence.

---

## 1. One-paragraph version

A parent orchestration spawns a child agent (e.g. `page-content-writer`, `image-generator`),
goes into `AWAITING_RESPONSES`, and **never receives the child's response** — even though the
child frequently *completes its work*. The parent then sits idle until a periodic SQL reaper
marks it `FAILED` at 30 min (dispatch loops) or 90 min (everything else). The proximate cause
is a **network path failure from certain worker nodes to Kafka broker-2**
(`personae-kafka-cluster-combined-pool-prod-2`, `10.20.99.93:9092`): child pods that land on a
bad node retry-loop forever on `dial tcp 10.20.99.93:9092: i/o timeout` and can neither consume
their job topic nor publish their response. Two platform gaps make it worse: (a) the child pod
never crashes, so k8s doesn't reschedule it onto a healthy node; (b) the parent's own
`timeout_seconds` never fires, so the only backstop is the slow reaper.

---

## 2. Symptom (what you'll see)

Orchestrations stuck/failed at a **spawn or call step**, with a reaper error. Fleet-wide, last
2 days (all sites):

```
current_step                        | count
------------------------------------+------
spawn_dispatch                      |   38
spawn_ingester                      |   10
call_dispatch                       |    8
process_item_iter_*_spawn_handler   |  ~19 (across iters)
process_item_iter_*_call_handler    |  ~12
spawn_image_gen_imagery             |    4
call_content_writer                 |    3
```

Query to reproduce that table:
```sql
SELECT current_step, count(*)
FROM orchestration_states
WHERE error LIKE 'reaper:%' AND created_at > now() - interval '2 days'
GROUP BY current_step ORDER BY count(*) DESC;
```

The two reaper error strings:
- `reaper: dispatch loop idle for >30 min`  (owner_agent_type = `build-dispatch-loop`)
- `reaper: stale AWAITING_RESPONSES for >90 min`  (everything else)

On leopardess (`site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`) alone I hit **11+** of these
in one session, across `spawn_content_writer`, `call_content_writer`, `spawn_image_gen_imagery`,
and `process_item_iter_*_(spawn|call)_handler`.

---

## 3. Root cause — evidence

**3a. The child pods can't reach Kafka broker-2.** A spawned `image-generator` pod's logs:
```
failed to dial: failed to open connection to
personae-kafka-cluster-combined-pool-prod-2.personae-kafka-cluster-kafka-brokers.kafka.svc:9092:
dial tcp 10.20.99.93:9092: i/o timeout
```
It then loops `"No activity for 5 minutes"` forever. It processed its `initialize` message and
sent an init response, but its **request/response consumers can't dial the broker**, so it never
does the actual work nor publishes back.

**3b. It's node-specific, not a broker outage.** Broker-2 has been `Running` for 17 days.
Broker topology:
| broker | pod IP | node |
|---|---|---|
| combined-pool-prod-0 | 10.20.161.217 | prod-instance-17744590808031336 |
| combined-pool-prod-1 | 10.20.0.21 | prod-instance-17722135234001149 |
| **combined-pool-prod-2** | **10.20.99.93** | **prod-instance-17735924839006832** |

A `page-content-writer` pod on a **different** node (`prod-instance-17735925437536833`) had
**0 dial errors** and successfully produced messages in the same window. So (some) nodes cannot
route to `10.20.99.93:9092` while others can. **The broker-0/1 vs broker-2 asymmetry is the
tell** — start the network investigation there.

> **CORRECTED 2026-07-20 ("bugfix 003" thread):** the broker-2/one-node signature no longer
> reproduces. 12h of live logs show dial i/o timeouts to **all three brokers from at least
> four different nodes** (broker-0 dominating), low-grade and intermittent. See the
> 2026-07-20 research pass at the end of this file — a static node→broker-2 network fix
> would chase a moving target; the durable fixes are the platform gaps in §4.

**3c. The reaper is a scheduled SQL sweep, not in-process.** The error strings come from
`docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql` (~lines 1060–1078), a `scheduled_tasks`
`pre_query` that bulk-`UPDATE`s `orchestration_states … SET status='FAILED'` where
`last_activity < NOW() - INTERVAL '90 minutes'` (or 30 for dispatch loops). That is why the
**workflow's own `timeout_seconds` never rescues it**: the in-code max-age check
(`platform/orchestration/coordinator.go:785`, `maxAge = TimeoutSeconds * 3`) only runs when the
orchestration is *re-processed*, and a lost child response never triggers a re-process. The SQL
reaper is the only backstop, and it's slow by design.

---

**3d. SECOND ROOT CAUSE — the consume loop is AT-MOST-ONCE, so any restart destroys
in-flight messages.** *(Added 2026-07-18, "session coordination" thread. Independent of
the broker-2 network path in 3a–3b: same symptom class, different mechanism. This is the
mechanism §4.3 was missing when it suspected "deploy churn is the spawn-killer".)*

`platform/kafka/consumer.go` `Consume()` (L74–108) fetches at **L81** and commits the
offset at **L103** — *before the message is processed*. The comment at L102 reads
"After successful processing, commit the offset", but nothing happens between the fetch
and the commit; the handler runs back in the caller, after `Consume()` has returned. The
intent was at-least-once; the code is **at-most-once**.

Both main chassis loops use it — `platform/agentbase/agent.go:468` (requests) and
**:528** (responses). So when a pod dies mid-work Kafka has already been told the message
was handled and **will never redeliver it**. That is the difference between "a restart
delays orchestrations" and "a restart annihilates them", and it is not deploy-specific:
OOM kills, evictions and node failures lose work identically. It is also the likely
mechanism behind the ~300s post-restart "spawn is silently dropped" rule in `CLAUDE.md`.

**The correct shape already exists in-tree**, which is the tell that this is an oversight
rather than a design choice: `platform/agentbase/client.go:60/105` and
`platform/agentbase/server.go:62/124` both use the manual `FetchMessage` … process …
`CommitMessages` pattern. Only the main agent loop takes the shortcut.

**The dedupe layer has the SAME defect — so fixing the offset commit alone changes nothing.**
*(Verified 2026-07-18. This CORRECTS the first version of this section, which claimed the
dedupe "makes the fix small". It does not. Read this paragraph before planning the fix.)*

A `processed_messages` dedupe does exist and is wired into both inbound paths — but **both
record RECEIPT, not COMPLETION**:

| path | seen-check | record | the work |
|---|---|---|---|
| `platform/agentbase/agent.go` | L801 | **L811** | L822 `processor.ProcessMessage` |
| `platform/messaging/processor.go` | L1296 | **L1317** | L1323+ (`NewMessageContext` → handler) |

So the message is marked processed *before* it is processed. If the pod dies mid-work and
the offset commit has been fixed to redeliver, the redelivered copy hits
`HasProcessedMessage` → `true` → **"Duplicate message ignored"** (`agent.go:805`) and is
dropped. The work is lost exactly as before, just through a different door. **This is the
same anti-pattern as §3d's offset commit, occurring a second time one layer up.**

Two further holes in the same layer:
- **No `request_id` ⇒ no dedupe at all, silently.** `HasProcessedMessage` returns `false`
  and `RecordMessageProcessing` returns `nil` when `RequestID == ""`
  (`platform/orchestration/state.go:163–165, 188–190`); the processor path is additionally
  gated on `execCtx.RequestID != ""` and `p.sqlDB != nil`.
- The agent-path gate `a.isStateless` is **not** a risk: hardcoded `true`
  (`agent.go:199`, `processor.go:100` "Always stateless now"). Checked, so nobody re-checks.

**Standing damage, measured 2026-07-18:** 22 wedged `EXECUTING_STEP` orchestrations, oldest
**1,224 hours (~51 days)** — every one a message that was acknowledged and then lost, with
§4.3's reaper blind spot ensuring it never gets cleaned up.

## 4. The platform gaps to fix (independent of the network fix)

1. **Child pods should fail fast on persistent Kafka dial failure.** Right now a pod that can't
   dial its broker logs `"No activity for 5 minutes"` indefinitely and stays `Running`, so k8s
   never reschedules it onto a healthy node. Add a liveness probe (or self-crash) when the
   Kafka consumer can't dial for N minutes. Entry point: `platform/agentbase/agent.go` around
   the idle-warn at line 1080 and `IDLE_TIMEOUT_SECONDS` handling at line 421.

   > **CORRECTED 2026-07-20:** spawned Jobs already HAVE liveness+readiness probes
   > (`cmd/remote-job-spawner/main.go:450–478` and `spawn_actions.go:2792–2812`) — but they
   > point at `/health` and `/ready`, which are **hardcoded 200s**
   > (`cmd/agent-chassis/main.go:141–150`). The fix is to make the endpoints honest, not to
   > add probes. Details + two adjacent gaps (idle_timeout_seconds=0 for exactly the wedging
   > child types; no probes at all on the chassis Deployment) in the 2026-07-20 research pass
   > at the end of this file.

2. **A parent in `AWAITING_RESPONSES` should honour its own `timeout_seconds`.** A lost child
   response shouldn't wait 30–90 min for a global SQL sweep. Either enforce the workflow
   timeout on the awaiting state, or shorten the reaper thresholds
   (`docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql` ~L1065/L1075). Prefer per-workflow
   timeout over a blanket threshold.

3. **The reaper has a blind spot: `EXECUTING_STEP` is never swept AT ALL.** *(Added 2026-07-17,
   "diagnosis fixloop 3" thread.)* The reaper's two error strings (§2) show its coverage:
   `AWAITING_RESPONSES` >90 min, dispatch loops >30 min. A parent that loses the spawn
   *itself* — the child orchestration row is never created, so the parent never reaches
   `AWAITING_RESPONSES` — wedges at **`EXECUTING_STEP` on the spawn step forever**. Nothing
   reaps it; the in-code max-age check (§3c) never fires because the orchestration is never
   re-processed.

   **Live specimen (preserved deliberately — inspect, do not tidy):** correlation
   `80c35dea-9488-46b1-97bf-7321af5c5af0` (2026-07-16 ~20:24Z) — a diagnose-orchestrator
   parent stuck at `spawn_diagnoser`, ZERO child rows, ZERO LLM calls, 13.7h stale when found.
   Deploy churn (v1.0.1126→1128 overnight) is the suspected spawn-killer, i.e. this gap also
   converts every deploy-window spawn loss into a permanent zombie.

   **Fleet scale of the blind spot** (query below): wedged `EXECUTING_STEP` rows going back
   **455–1,197 HOURS** — weeks-old zombies accumulating silently.
   ```sql
   SELECT correlation_id, current_step,
          EXTRACT(EPOCH FROM (NOW()-last_activity))::int/3600 AS stale_h
   FROM orchestration_states
   WHERE status='EXECUTING_STEP' AND last_activity < now() - interval '2 hours'
   ORDER BY last_activity LIMIT 20;
   ```
   **Fix shape:** extend the reaper with an `EXECUTING_STEP` clause (threshold can be generous —
   e.g. >3h — no legitimate step runs that long) OR add per-status thresholds; either way the
   reaper error string must name the step so triage's failure-pattern grouping (left(error,140))
   buckets these as one platform pattern. Note the diagnosis loop's own mitigation (an early
   child-row check ~3 min after dispatch) catches NEW losses at dispatch time but does nothing
   for the standing zombie population — the sweep is still needed.

4. **Make the consume loop at-least-once, and make a restart survivable.** *(Added
   2026-07-18; follows from §3d. This is the one that unlocks continuous deploy — see
   the note at the end of this section.)*

   a. **Commit after the handler succeeds, not on fetch.** Change the two call sites
      (`agent.go:468`, `:528`) to the `FetchMessage` … process … `CommitMessages` pattern
      already used by `client.go`/`server.go`, or fix `Consume()` itself to take a handler
      and commit on its success.

   a-bis. **…and move the dedupe record to completion, or (a) is INERT.** Per §3d, the
      redelivered message would be suppressed as a duplicate and the work lost anyway.
      These two must ship together; either alone is a no-op or worse. The naive form —
      just moving `RecordMessageProcessing` after the handler — reopens a window where two
      copies in flight *simultaneously* both pass the seen-check, so prefer a **two-phase
      claim**: insert on receipt with a lease/heartbeat (`processing`), mark `complete` on
      success, and treat a lease that expired without completing as reprocessable. That
      shape also gives the reaper in §4.3 something honest to sweep, and it is the same
      claim/lease pattern `site_work_items` already uses (`claimed_at` + the 40-minute
      claimed-item timeout) — reuse it rather than inventing a second one.
      **Also decide what happens with no `request_id`** (§3d): today those messages are
      undeduped and, post-fix, would be reprocessed on every redelivery. Either guarantee
      a `request_id` on every inbound path or make its absence loud rather than silent.
   b. **Fix the drain.** The chassis deployment (measured 2026-07-18) has
      `terminationGracePeriodSeconds=30` while `Agent.Shutdown()`
      (`platform/agentbase/agent.go:1088`) itself waits **up to 30s** for in-flight
      goroutines — SIGKILL races the graceful drain and can win. Grace must comfortably
      exceed the drain wait; add a `preStop` hook (there is none) so the pod stops
      accepting new messages and quiesces before dying.
   c. **Add a readiness probe that means it.** There is **none**, so a new pod counts as
      Ready the instant it starts and k8s kills the old one before the replacement has
      joined the consumer group and been assigned partitions. Report Ready only once it is
      actually consuming.
   d. **`replicas=1`** — a rollout leaves a window with no consumer at all. ≥2 (check the
      partition count first: with one partition you get failover, not parallelism).

   **Why this ordering matters for CD.** Continuous deploy on top of at-most-once delivery
   makes things *worse*, not better — it deploys more often, and every deploy destroys
   in-flight work. Fix (a) and restarts become redelivery rather than loss; (b)–(d) close
   the rollout gap; only then is automated deploy safe. A drain-before-deploy dance would
   paper over deploys specifically while doing nothing for crashes or evictions, which is
   why the delivery guarantee is the durable fix.

---

## 5. Diagnostics to run FIRST in the new thread

```bash
# (a) Confirm the node→broker-2 path. Pin a debug pod to the suspect node and probe the broker.
kubectl run netcheck --image=busybox --restart=Never -n ai-persona-system \
  --overrides='{"spec":{"nodeName":"prod-instance-17735924839006832"}}' \
  -- sh -c 'nc -vz 10.20.99.93 9092; echo rc=$?'
kubectl logs -n ai-persona-system netcheck; kubectl delete pod -n ai-persona-system netcheck
# Repeat pinned to a KNOWN-GOOD node (prod-instance-17735925437536833) — expect it to connect.

# (b) NetworkPolicy / routing around the kafka namespace and the ai-persona-system namespace.
kubectl get networkpolicy -A | grep -i kafka
kubectl -n kafka get pod personae-kafka-cluster-combined-pool-prod-2 -o yaml | grep -iA3 'listener\|advertised'

# (c) Which nodes are the bad ones? Correlate failed orchestrations to the node their child ran on.
#     (Child pod names carry the parent's short orchestration id; check pod → node placement.)
```

Also confirm scope with the SQL in §2 — if `spawn_dispatch` (business-intel ingestion) dominates,
this is hurting far more than the website fleet.

---

## 6. Standing reproduction / verification case

`ai-readiness-quiz` on leopardess is a clean repro: its content path is known-good (validation
passes, the `contact-block` validator blocker was fixed fleet-wide in turn 17), so **the only
thing stopping it building is this flake.** Fire it and watch whether the spawned
`page-content-writer` lands its response:

```bash
# site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732 ; page_name = ai-readiness-quiz
# (kcat trigger pattern: docs/leopardessconsulting/scripts/reassemble_pages.sh, or the
#  page-build-handler fire in RUNNING_NOTES turn 17.)
```
Success = the page gains `page_components` rows and the live page grows past its 12,425-byte
blank shell. Attempts 1–5 all FAILED here: none on content/validation, all on this infra flake
(take 5's error was literally `reaper: stale AWAITING_RESPONSES for >90 min`).

---

## 7. Useful facts / landmines

- **The 5-min "No activity" warning is a red herring for health** — the pod is alive, it just
  can't reach the broker. Don't treat `Running` as healthy for spawned agents.
- **Deleting the stuck child pods + re-firing works *sometimes*** — because the replacement may
  land on a good node. That's a workaround, not a fix, and it's why the failure looks
  intermittent.
- **`kubectl` auth in this environment expires** (saw `Unauthorized` mid-session). If trigger
  scripts end in `>/dev/null 2>&1` they publish nothing silently — don't swallow kcat stderr;
  assert an `orchestration_states` row appears after firing.
- **Broker-2 pod itself is fine** — do not restart it hoping to fix this; the problem is the
  path *to* it from certain nodes.

## 8. Key files
- `docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql` — the reaper (~L1060–1078)
- `platform/orchestration/coordinator.go:772–800` — in-code max-age check (only runs on re-process)
- `platform/agentbase/agent.go:421, 1080, 1410` — idle timeout / "No activity" warn / idle reaper
- `docs/leopardessconsulting/RUNNING_NOTES.md` turn 17 — the session where this was characterised
- **§3d/§4.4 (at-most-once) files:**
  - `platform/kafka/consumer.go:74–108` — `Consume()`; fetch L81, commit L103, misleading comment L102
  - `platform/agentbase/agent.go:468, 528` — the two call sites (request + response loops)
  - `platform/agentbase/client.go:60/105`, `platform/agentbase/server.go:62/124` — the CORRECT
    fetch/process/commit pattern to copy
  - `platform/orchestration/state.go:170, 207` — the existing `processed_messages` dedupe
    (verify its coverage before flipping the commit)

---

## Fresh occurrence — 2026-07-19, on a DIAGNOSIS spawn (claims-verification thread)

Recorded because this one is **not** a site-content pipeline: it killed a
`090_TRIGGER_needs_diagnosis` run, which widens the blast radius stated above
(image generation, page builds, dispatch-loop handlers) to include **the
diagnosis loop itself**. A thread following CLAUDE.md's "file it to the loop
before you assert" rule can therefore be silently denied the verdict the rule
tells it to obtain.

**Evidence.**

| fact | value |
|---|---|
| correlation | `46253496-f8e0-471f-9ae0-29c9e630ada5` |
| parent orchestration | `fa5ce58b-c46a-49be-a95e-fb15170f93b5` (owner `generic`) |
| parent state | `AWAITING_RESPONSES @ spawn_diagnoser`, unchanged for 45+ min |
| awaited request | step `spawn_diagnoser`, target `diagnose-agent`, **status `expired`**, `timeout_at 13:14:06`, **`retry_version = 0`** |
| child pod | `agent-diagnose-agent-98d8f09c-fdf7r`, phase Running, 63 log lines, **zero error-level lines** |
| child's last substantive log | `"Probably child agent. Initialization complete, now starting agent's own workflow"` — then only 5-minute idle warnings |
| child init response | `SendInitializationResponse` → `agent producer.go Successfully produced message` (the child DID reply) |
| artifacts written | none (`diagnosis_artifacts` empty for the correlation) |
| intake item | `needs_diagnosis:diagnose-council-decide-in-platform-orch`, still `awaiting_diagnosis` |

**Two things this occurrence adds to the case above.**

1. **The sweeper did not rescue it.** `retry_version = 0` on an expired request
   45 minutes past `timeout_at`. Whatever reclassification the stale-orchestration
   sweeper is supposed to perform (001 §"Stale Orchestration Sweeper": synthesize
   from the child's `final_result`, or re-send up to 3 times) did not fire here.
   Worth checking whether the sweeper skips `spawn_*` steps, or whether it requires
   a child orchestration row — there is none, because the child never got as far as
   creating one.

2. **It leaves a queue-blocking residue.** The intake item stays
   `awaiting_diagnosis`, and the 090 trigger's own pre-dispatch coverage check
   refuses to dispatch when open work exists on the target. So one lost spawn also
   blocks the *retry* of that same diagnosis unless the operator passes `FORCE=1`
   or closes the item by hand. A stall that self-blocks its own remedy is worth
   fixing above its raw frequency.

**Observability note for whoever picks this up.** The failure is completely silent
from the child's side: no error, no failed status, just a healthy-looking pod
logging idle warnings. The only positive signal is `awaited_requests.status =
'expired'` joined to the parent. Any monitor that watches pod health or log
severity will report this as fine.

---

## Research pass — 2026-07-20 ("bugfix 003" thread): third root cause, two corrections, fix plan

Every §3d/§4 code citation was re-verified against today's HEAD and still holds
(`consumer.go` fetch L81 / commit L103; call sites `agent.go:468/:528`;
receipt-time dedupe `agent.go:801/811` vs work at `:822`; the processor-path
mirror; the `request_id=""` silent-no-dedupe gates). Live reaper config matches
§3c plus one addition: its pre_query also marks `awaited_requests` rows
`expired` after a 5-min grace — **but nothing anywhere consumes `expired`**
(grep of the whole tree; the only writers are the reaper CTE and
`cleanup_expired_awaited_requests()`, which merely mark and, after 7 days,
delete).

**Live scale, 2026-07-20:** 70 reaper-failed orchestrations in the last 2 days
(`spawn_ingester` 27 — §5's business-intel fear confirmed; `spawn_diagnoser` 4 —
the diagnosis loop is being eaten). 24 wedged `EXECUTING_STEP` rows, oldest
`last_activity` **2026-05-28 (~53 days)** — §4.3's blind spot, still unswept.

### THIRD ROOT CAUSE — timeout enforcement is a process-local sleeping goroutine; a pod restart deletes every pending timer, and nothing rebuilds them

*(Direct code reading, each claim grep-verified. Filed to the diagnosis loop
before being asserted here — correlation `d971e8c2-0c41-4251-b46f-705b471f5dc1`,
item_key `needs_diagnosis:workflow-step-timeouts-on-awaited-child`; verdict
pending at the time of writing — check `diagnosis_artifacts` for that
correlation before building on this section.)*

The retry machinery §"Fresh occurrence" hoped for **does exist and is correct**:
`handleRequestTimeout` (`platform/orchestration/coordinator.go:2962`) checks the
DB row, and if still waiting resends with `retry_version++` (max 3, then
`routeToErrorStepOrFail` / loop-skip). What's broken is **how it is driven**:

- Its only drivers are `go s.handleRequestTimeout(...)` at
  `coordinator.go:1816` and `:2117` — goroutines spawned **at request-send
  time**, whose first line is `time.Sleep(time.Until(timeoutAt))`. The timer
  lives and dies with the pod that sent the request.
- `TimeoutMonitor` (`platform/orchestration/helpers.go:19–80`) looks like the
  durable answer but is **dead code** — `NewTimeoutMonitor` has zero callers.
- Nothing on startup scans `awaited_requests` to re-arm timers, and nothing
  consumes `status='expired'` (see above). So after a restart of the owning pod,
  a lost child response has **no rescue path at all** until the 90-min reaper.
- The owner of most parents is `generic` — the `agent-chassis` Deployment,
  `replicas=1`, rolled constantly (current pod born 2026-07-20 07:35Z,
  v1.0.1139). Every roll silently deletes every pending timer in that pod.

This fully explains the fresh occurrence's `retry_version = 0` at 45+ min past
`timeout_at` (attribution of that specific instance to a restart is plausible
rather than proven — the structural claim does not depend on it). It also
answers its open question: the "sweeper" performs no reclassification because
**there is no sweeper-driven retry** — retries fire only from (a) the in-memory
timer, or (b) an explicit recoverable-error *response* from the child
(`coordinator.go:291`). A child that never replies triggers neither.

> **NUANCE ADDED 2026-07-20b (owner's point, verified):** retry machinery DOES
> exist in `scheduled_tasks` — one layer up. `claimed-item-timeout` (every
> 120s) auto-completes a claimed `site_work_items` row when artifact evidence
> proves the work landed, and resets evidence-less claims >40 min old to
> `triaged` for re-dispatch, capped by `max_attempts`. So **work-item-backed
> flows do get an eventual retry**: reaper FAILs the orchestration (30–90 min)
> → claim times out (≤40 min) → a dispatch loop re-runs the whole item. The
> limits are what F2 addresses: latency is 70–130 min not ~1 min; the item is
> redone from scratch (in-orchestration progress lost); attempts are finite;
> and flows with **no work item or no running dispatch loop get nothing** —
> child spawn chains inside an orchestration, direct kcat/council triggers,
> and the diagnosis intake while `diagnose-pipeline-trigger` is disabled
> (enabled=f today). The awaited-request layer itself still has no consumer of
> `expired`.

Note how the three root causes compound on any restart: §3d loses the in-flight
messages (offset already committed), this section loses the timers that would
have noticed, and §4.3 means the wedged parent is never reaped. That is why
deploy windows are so destructive.

### Adjacent gaps confirmed while verifying §4

- **Probes exist but measure nothing** (correction to §4.1, marked inline
  above): `/health` and `/ready` return hardcoded 200s
  (`cmd/agent-chassis/main.go:141–150`); `checkHealth()` (`agent.go:1068`) pings
  only the DB and is wired to no probe. `platform/health/server.go` already has
  a `Checkers`-based server the chassis doesn't use.
- **The chassis Deployment has NO probes, no preStop, default 30s grace**
  (`deployments/kustomize/services/agent-chassis/base/deployment.yaml`),
  `replicas=1` in prod — §4.4b–d confirmed as filed.
- **`agent_definitions.idle_timeout_seconds = 0` for `diagnose-agent` and
  `image-generator`** — the idle monitor (`agent.go:421`) never starts for
  exactly the child types observed wedging; only the Job's
  `ActiveDeadlineSeconds=86400` (24h) bounds them. (`page-content-writer` is
  180s.)
- **The git adapter has the §3d defect too**: `internal/adapters/git/adapter.go:142`
  and `:272` are the only other `Consume()` callers in the tree — a restart
  mid-`git_commit` loses the message identically.
- **`system.agent.generic.requests` has PartitionCount=1** (RF 3), so §4.4d's
  `replicas≥2` buys failover only, not parallelism, unless partitions are
  raised.

### Proposed fix plan (2026-07-20) — ordered by leverage per risk

**F1 — reaper coverage (SQL/config, LIVE immediately, no image roll).** Add an
`EXECUTING_STEP` clause to the reaper pre_query (fixed prefix so triage's
`left(error,140)` grouping still buckets the pattern; step name appended):

```sql
failed_wedged AS (
    UPDATE orchestration_states
    SET status = 'FAILED',
        error = 'reaper: stale EXECUTING_STEP for >4h; step=' || COALESCE(current_step, '(none)'),
        updated_at = NOW()
    WHERE status = 'EXECUTING_STEP'
      AND last_activity < NOW() - INTERVAL '4 hours'
    RETURNING orchestration_id
)
```

First firing drains the standing zombies. Apply via `UPDATE scheduled_tasks`
AND mirror into `020_scheduled_tasks.sql` (the patch-style re-seed clobber
landmine).

> **CORRECTED & APPLIED 2026-07-20 12:43Z:** the draft said `>3h` / "no
> legitimate step runs 3h". Checked against `orchestration_state_audit`
> (7.5 weeks of history) before applying: exactly ONE healthy exit from an
> EXECUTING_STEP stint over 3h (3.72h, a `check_health` step, 2026-06-28),
> none over 4h — so the applied threshold is **>4h** (zero historical false
> positives). Also added `COALESCE(current_step,'(none)')` — the draft's bare
> `|| current_step` would NULL the entire error string on a NULL step.
> Live in `scheduled_tasks` and mirrored to `020_scheduled_tasks.sql`.

**F2 — durable, DB-driven retry of expired requests (Go).** Keep the sleeping
goroutine as the fast path; make the DB the guarantee. The per-pod 1-min ticker
`cleanupExpiredAwaitedRequests` (`coordinator.go:3649`) already exists: extend
it to atomically claim newly-expired rows
(`UPDATE … SET status='retrying' WHERE status='expired' … RETURNING` — atomic,
so concurrent pods can't double-drive; add `'retrying'` to the status CHECK)
and push each through the **same body as `handleRequestTimeout`** (factor it
out): `retry_version<3` → the existing resend at `coordinator.go:2760–2860`
(increments, re-arms `timeout_at`, persists, produces); `≥3` →
`routeToErrorStepOrFail`/loop-skip. Per-workflow timeouts then hold to ~1-min
granularity across restarts, and the 90-min reaper becomes a true backstop.
`EXECUTING_STEP` wedges (no request row exists) stay F1's job — the two layers
are complementary.

**F3 — at-least-once consume + completion-time dedupe (Go; §4.4a + a-bis, must
ship together).**
- Port all four `Consume()` call sites (`agent.go:468/:528`,
  `git/adapter.go:142/:272`) to `FetchMessage` → process → `CommitMessages`,
  then delete `Consume()` so the trap can't be re-adopted. Commit after
  `processMessage` *returns* (handler errors already route through
  `handleProcessingError` — the guarantee being bought is against pod death,
  not against handler error, so no poison-message loop is introduced).
- `processed_messages` two-phase claim (needs a migration — the table has no
  status column): add `status ('processing'|'complete')` + `lease_expires_at`;
  `RecordMessageProcessing` writes `processing` with a lease; new
  `MarkMessageComplete` after the handler succeeds (wire at `agent.go` ~849 and
  the `processor.go` mirror); `HasProcessedMessage` treats as duplicate only
  `complete` OR `processing` with a live lease. Same claim/lease shape as
  `site_work_items` — reuse, don't invent.
- `request_id=""` → make it loud (WARN + `MessagesDropped{reason="undeduped"}`
  metric); guaranteeing request_id on every inbound path is follow-on work.

**F4 — honest health, fail-fast, survivable rollout (Go + kustomize + config;
§4.1, §4.4b–d).**
- Consumer wrapper tracks last-successful-fetch / consecutive-dial-failure;
  chassis `main.go` swaps its hardcoded mux for `platform/health.Server` with
  real checkers (DB ping; Kafka broker dial). `/ready` reports ready only once
  consumers are constructed and polling.
- Self-crash backstop in agentbase: Kafka unreachable continuously for ~10 min
  → `os.Exit(1)`. Job `RestartPolicy=OnFailure` reschedules (possibly onto a
  healthy node); this is what actually un-wedges §3a-style pods even where
  probes are stale.
- Base deployment.yaml: liveness+readiness probes, `terminationGracePeriodSeconds:
  60` (must exceed `Shutdown()`'s 30s drain), `preStop` sleep; prod overlay
  `replicas: 2` (failover-only at 1 partition — raising partitions is a
  separate decision).
- Config (live immediately, no image): set `idle_timeout_seconds` for
  `diagnose-agent` (~600) and `image-generator` (~900).

**Sequencing.** (1) F1 + F4's idle_timeout config now — both config-only.
(2) F2+F3 together in one image roll — they interlock: F3's redelivery is only
safe with F3's completion-dedupe, and F2 gives a silent loss a driver. Council-
gate (097) before commit — `platform/` scope. (3) F4's Go + kustomize in the
next roll. Network flakiness is explicitly NOT fixed by any of this — after
F2–F4 it degrades from silent loss to visible retries; if dial-error rates stay
high, file a separate infra case (conntrack/CNI/broker load) with the new
health metrics as evidence.

**Verification for the fixing thread.**
- §6's `ai-readiness-quiz` repro, end-to-end.
- Pod-grep a literal the change **creates** (e.g. the new reaper error prefix,
  the `'retrying'` status string) plus a positive control — not a string the
  change merely uses.
- Deliberate chassis roll mid-orchestration: `awaited_requests.retry_version`
  must increment and the orchestration complete, instead of reaper-failing at
  90 min.
- `EXECUTING_STEP` zombie count drops to 0 and stays there.
- Re-run §2's SQL after a week: `spawn_*` reaper deaths should collapse.
