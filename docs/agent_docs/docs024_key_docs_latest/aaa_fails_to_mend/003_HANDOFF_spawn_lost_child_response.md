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

## 4. The two platform gaps to fix (independent of the network fix)

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
