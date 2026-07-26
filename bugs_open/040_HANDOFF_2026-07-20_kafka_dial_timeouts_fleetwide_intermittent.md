# 040 — Kafka dial i/o timeouts are fleet-wide and intermittent: all three brokers, at least four of five nodes

> **NUMBER COLLISION (2026-07-20, same day):** another thread filed a different
> `040` (`failed_page_build_leaves_page_deployed_and_partially_composed`, now in
> `/bugs_closed/`). Numbers are never reassigned — resolve by slug
> (`bugs_closed/README.md` duplicate-numbers table). Cite this case as
> **040-kafka-dial**.

**Filed:** 2026-07-20 ("bugfix 003" thread) · **Status:** **OPEN** — instrumented
2026-07-26, root cause NOT established, changes INERT until an image roll
**Class:** was filed as cluster network infrastructure. **Reclassified 2026-07-26:**
the infrastructure hypotheses are refuted; what is proven is an *observability*
defect that made the bug unmeasurable. See §2.
**Severity:** medium on its own; multiplied by `bugs_open/003`'s platform gaps.
**Workstream:** `docs/agent_docs/docs024_key_docs_latest/bugfix_040_kafka_dial/`

---

## 0. Read this first (2026-07-26)

Two things will save you the budget this case has already burned twice.

1. **Do not work §4 top-to-bottom.** Five of its six diagnostics point at causes
   that measurement rules out. The refutation table is §2. §4 is kept below for
   the record, annotated.
2. **A zero from the new metric is not evidence of anything until you have proven
   the metric is scraped.** Until 2026-07-26 nothing in the fleet served
   `/metrics` at all — the live Prometheus held **zero** `ai_persona_*` series.
   `RUNBOOK_040_kafka_dial.md` §2 is the check. Do it first.

## 1. What changed since 003 §3a/§3b was written (2026-07-15)

003's original network evidence was sharp: child pods on ONE node retry-looped
forever on `dial tcp 10.20.99.93:9092: i/o timeout` (broker-2 only), while pods
on other nodes were clean — a static node→broker route failure.

**That signature no longer reproduces.** Verified 2026-07-20 (~11:30Z, 12h log
window): dial timeouts hit **all three brokers**, from **at least four of the
five nodes**, at low intermittent rates — and **broker-0 dominated**, not
broker-2. A static route fix would chase a moving target.

## 2. Refuted hypotheses (measured 2026-07-26 — do not re-walk these)

| Hypothesis | §4 ref | Measured 2026-07-26 | Verdict |
|---|---|---|---|
| conntrack exhaustion / SNAT port pressure | 4.3 | 1,021 of 262,144; 28-day max 113,891 | **REFUTED** |
| listen-queue overflow | 4.3 | `TcpExt_ListenOverflows` increase **0**, all 5 nodes, 8 consecutive days | **REFUTED** |
| broker request-handler saturation / GC | 4.5 | brokers at 60m/62m/65m CPU; node CPU 1–4% | **REFUTED** |
| node packet pressure | — | `node_softnet_dropped_total` **0** over 28d; CPU steal <0.5% | **REFUTED** |
| CoreDNS too slow | — | p99 request duration **2.69ms**; 0 panics | **REFUTED** |
| client-side DNS stalls | — | 1,200 probe lookups, 3 pods on 3 nodes: **0 stalls ≥2s** | **NOT REPRODUCED** |

Every query is in `RUNBOOK_040_kafka_dial.md` §6. Re-run them before suspecting
any of these again — they were true on 2026-07-26, not for ever.

**§4.6 is the one that held** and has been acted on: the dial timeout was 10s and
is now 5s, env-tunable.

### The one event captured in full

```
core-manager-7b6cd994b6-2z2bs   started 12:01:21Z
12:02:10.825Z  Creating topics for agent build-dispatch-loop
12:02:21.146Z  Failed to create topic system.agent.build-dispatch-loop.errors
               failed to connect to controller: failed to dial: failed to open
               connection to personae-kafka-cluster-combined-pool-prod-0.
               personae-kafka-cluster-kafka-brokers.kafka.svc:9092:
               dial tcp 10.20.161.217:9092: i/o timeout
12:02:21.820Z  Successfully created all topics for agent   topic_count 4
```

T+60s from pod start; **10.32s** elapsed — exactly kafka-go's `DefaultDialer` 10s,
so it waited the whole budget; the retry then succeeded in **0.67s**. One event in
~2h of uptime. Brief, self-healing, costs 10s (now 5s) each time.

## 3. ROOT CAUSE: not established. Here is why, and it is the actual finding

**The bug was unmeasurable, and that is now fixed.**

- **No Kafka dial metric existed anywhere.** Bug 003's F4 counters were planned
  and never built.
- **The primary request lane counted nothing.** `platform/agentbase/agent.go`
  logged, slept 1s, continued — forever. `server.go`'s *secondary* loop had always
  incremented `SystemErrors{error_type="fetch_message"}`; the primary never did.
- **§4.1's method cannot produce a durable baseline.** It says grep pod logs, but
  the pods that flake are ephemeral spawned Jobs whose logs GC quickly.
- **And the finding that reframes the case: nothing in the fleet had ever served
  `/metrics`.** `observability.NewMetricsServer` has zero callers.
  `cmd/agent-chassis/main.go` built its own mux with only `/health` and `/ready`.
  Yet `spawn_actions.go` annotates every spawned pod `prometheus.io/port: "9090"`
  + `prometheus.io/path: /metrics` and declares a matching containerPort.
  **Prometheus has been scraping a closed port for the life of the fleet** —
  confirmed by querying the live server for metric names: **zero `ai_persona_*`
  series.** Every counter in `platform/observability` has been dead since written.

**Proven from the live pod, not just from Prometheus** (`agent-chassis-5645cb45d6-kpxtq`,
image `v1.0.1167`, 2026-07-26 — with both controls, so the method is not the
thing being tested):

```
strings /app/agent-chassis | grep -c ai_persona_kafka_messages_produced_total  -> 1   (positive control)
strings /app/agent-chassis | grep -c ai_persona_kafka_dial_nonexistent_xyz     -> 0   (negative control)
wget -qO- http://localhost:9090/metrics  -> wget: can't connect to remote host: Connection refused
```

The metric strings **are** compiled into the shipped binary. The port the pod is
annotated for is **closed**. That is the whole defect in three lines: the counters
exist, are maintained, and are unreachable.

### …and there is a SECOND layer. The annotations were never going to work either

Opening the port is necessary but **not sufficient**, and this nearly slipped
past. Every spawned pod carries `prometheus.io/scrape: "true"` +
`prometheus.io/port: "9090"` — but that is the **plain-Prometheus
`kubernetes_sd_configs` convention, and this cluster does not run plain
Prometheus.** It runs the prometheus-operator (kube-prometheus-stack), which
discovers targets **only** through label-selected CRs. Verified 2026-07-26:

```
spec.podMonitorSelector      = {matchLabels: {release: kube-prometheus-stack}}
spec.additionalScrapeConfigs = None
kubectl get podmonitors -A   -> 0
kubectl get scrapeconfigs -A -> 0
```

**So the annotations are inert and have never caused a single scrape.** Had only
the listener shipped, `/metrics` would have been open and still unscraped, and the
resulting zero would have looked exactly like a fixed bug — the same trap one
level up.

The missing half is committed as
`deployments/kustomize/services/agent-chassis/base/podmonitor.yaml`
(**NOT APPLIED**; selector verified against all three live pod shapes; the
`release: kube-prometheus-stack` label is what makes it visible at all). Apply it
with the roll — or before, as a pre-flight: the target will appear and read DOWN,
which confirms discovery works while the port is still closed.

So this case could never have been closed on evidence, because there was no
evidence channel. That, not the network, is what has been fixed.

## 4. Original diagnostics (kept for the record — see §2 before using)

1. ~~**Baseline the rate** via the §5 grep across all agent pods.~~ **Superseded:**
   use `ai_persona_kafka_dial_total` (RUNBOOK §1). The grep is unreliable — see the
   uptime trap in §7 below.
2. **Node-pinned probes** (003 §5a pattern): busybox `nc -vz <broker-ip> 9092`
   pinned to each node in turn, N attempts, record the failure rate. **Still
   unexercised and still worth doing** — this is the most promising untried lead.
3. ~~**Conntrack pressure.**~~ REFUTED, §2.
4. **CNI + kubelet logs** on the worst node in the same minute as a logged dial
   timeout. **Still untried.** Calico is the CNI (`calico-node`, 5 nodes).
5. ~~**Broker side:** request-handler saturation / GC pauses.~~ REFUTED, §2 —
   though note the brokers have no resource floor at all (§6), so this could
   become true under load even though it is not true now.
6. **Kafka client dial timeout is 10s** — a dial that cannot complete in 10s on an
   in-cluster network is pathological regardless of cause. **ACTED ON:** now 5s,
   `KAFKA_DIAL_TIMEOUT` (whole seconds).

## 5. Repro of the original count

```bash
kubectl -n ai-persona-system logs --since=12h --tail=3000 <pod> \
  | grep 'i/o timeout' | grep -o 'dial tcp [0-9.]*:9092' | sort | uniq -c
```

## 6. Amplifiers found and removed (2026-07-26, commit `95df64d63`)

None of these is claimed to be the cause. Each made the fault worse or hid it.

- **Four inconsistent, uncounted dial configurations** — consumer 10s (explicit),
  producer **3s** (`Transport` left nil → kafka-go `DefaultTransport`), topic
  manager 10s × 8 sites (bare `kafka.Dial`), health probe 3s. None configurable.
  Collapsed into `platform/kafka/dialer.go`.
- **`/metrics` never served** — now on the health mux *and* on `METRICS_PORT`
  (9090), the port the annotations already advertise.
- **Phantom fallback broker** `kafka-0.kafka-headless.kafka:9092` in
  `topic_manager.go` — **no such Service** (Strimzi's headless service is
  `personae-kafka-cluster-kafka-brokers`). Could only burn a dial timeout before
  failing. Removed; remaining bootstrap default fully-qualified.
- **`kafka_brokers` with no port** in `deployments/kustomize/base/configmap-common.yaml`.
- **Hot spin** in `cmd/remote-job-spawner/main.go` — read errors `continue`d with
  no pause. Now the standard 1s backoff.
- **Producer connection churn** — kafka-go's `IdleTimeout` 30s / `MetadataTTL` 6s
  made low-traffic agents re-dial almost every produce. Now 5m / 30s.

### The DNS search-path tax — real, large, and NOT the cause

Brokers advertise (live, from the broker's `/tmp/strimzi.properties`):

```
PLAIN-9092://personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc:9092
```

**Three dots**, no `.cluster.local`. Pods run `ndots:5` with a three-domain search
path, so every post-bootstrap connection walks three NXDOMAIN rounds before the
one that works, doubled by the parallel AAAA. Measured over 24h: **384,392
NXDOMAIN of 525,152 responses (73%)**, A:AAAA exactly 1:1, ≈**7.5 queries per
useful answer** (predicted 8 for this name shape).

> **[MEASURED] This is a volume problem, not a latency problem, and not the
> cure.** CoreDNS p99 is 2.69ms; 1,200 probes showed no stalls; 300 short-name vs
> 300 FQDN lookups both finished under 1s. It triples packet count and therefore
> exposure to loss. It has **not** been shown to fix the timeouts. Do not write it
> up as the fix.

> **TRAP — do NOT lower `ndots`.** It is the obvious move and it is strictly
> worse: the name genuinely needs `cluster.local` appended, so at `ndots:2` the
> resolver tries it absolute (NXDOMAIN) **and then still** walks all three search
> domains — four rounds instead of three. The fix is broker-side `advertisedHost`
> with the FQDN. Documented in
> `deployments/terraform/modules/kafka-cluster/templates/kafka-cluster-cr-prod.yaml.tpl`,
> **NOT APPLIED** (rolls the cluster). Logged in `WRONG_CALLS.md`.

### Broker has no resource floor — corrected in the repo, NOT APPLIED

Live: container `resources: {}` and `KAFKA_HEAP_OPTS=-Xms128M` with **no `-Xmx`**,
so max heap falls back to the JVM default of ~¼ of node RAM (~15GB on these
nodes). RSS is 4.8GB while idle, so latent rather than active — but the brokers
have **no scheduling guarantee at all** and are eviction-ranked with best-effort
workloads. Corrected in `templates/kafka-nodepool.yaml.tpl` using the figures
already in this module's unreferenced `config/kafka-nodepool-cr-prod.yaml`.
**Not applied** — `terraform apply` is unsafe here (checked-in state is `serial: 1`,
zero resources); use the `kubectl patch` in RUNBOOK §8.

## 7. Traps for whoever picks this up

- **Normalise log counts by pod uptime.** A `--since=12h` sweep on 2026-07-26
  looked almost clean, but every pod had restarted ~100 minutes earlier, so the
  real window was 1.7h. Pull `.status.startTime` first. A window longer than the
  pod's age silently understates the rate.
- **Brokers live in namespace `kafka`, not `ai-persona-system`.**
- **busybox `date +%s%N` returns 0** — nanosecond timing in these images silently
  yields all-zero results. Time at second granularity.
- **`go build ./platform/...` may fail on another session's WIP.** Overlay your
  files onto `git archive HEAD` (RUNBOOK §7). This caught a real compile error of
  mine that the broken tree was masking.

## 8. Relationship to 003 (read before "fixing" anything here)

003's F2 (DB-driven retry), F3 (at-least-once + completion dedupe) and F4 (honest
health, self-crash, drain) are what convert this bug from "silently strands work"
to "causes visible retries". F2+F3 shipped in v1.0.1159. **F4's dial counters were
never built — this case now supplies them.** Do not block on the network
investigation to ship 003's work; equally, do not close this bug because they
shipped.

## 9. CLOSE CONDITION (explicit, 2026-07-26)

The changes above are **Go code and are inert until an image roll**. Do not close
on them.

1. Roll an image carrying `ai_persona_kafka_dial_total`. Verify against the running
   pod, grepping a string the change **creates** plus an absent control string
   (RUNBOOK §3).
2. **Prove the metric is scraped** (RUNBOOK §2). A zero before this step means
   nothing.
3. Baseline it, and cross-check once against the §5 log grep before retiring the
   grep.
4. **Close when** `sum by (outcome) (increase(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))`
   is zero fleet-wide — **or**, if non-zero, when the residual is diagnosed
   (the `dns`/`dns_timeout` vs `timeout` split is designed to answer exactly that
   next question: resolution or connect).

Per handoff §8 and the owner's ruling of 2026-07-26, amplifier fixes alone are
**not** grounds to close.
