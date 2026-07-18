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
