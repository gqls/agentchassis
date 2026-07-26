# PLAN — 040-kafka-dial: instrument first, then remove the amplifiers

**Opened:** 2026-07-26. **Case:** `bugs_open/040_HANDOFF_2026-07-20_kafka_dial_timeouts_fleetwide_intermittent.md`
**Cite by slug — `040-kafka-dial`.** The number is shared with a closed
partial-build case (`bugs_closed/README.md` duplicate-numbers table).

---

## 1. The originating brief, and the correction to it

The handoff describes intermittent `dial tcp <broker-ip>:9092: i/o timeout` to all
three brokers from at least four of five nodes, classes it as **cluster network
infrastructure** rather than a chassis code bug, and offers six diagnostics in §4:
baseline the rate by log-grep, node-pinned probes, conntrack pressure, CNI/kubelet
logs, broker request-handler saturation, and the observation that a 10s in-cluster
dial timeout is pathological.

> **CORRECTION 2026-07-26 — five of the six §4 diagnostics point at causes that
> measurement rules out.** Not because the handoff was careless; the cluster state
> it described was real on 2026-07-20. But anyone picking this up and working §4
> top-to-bottom would spend the whole budget on conntrack and broker saturation.
> Measured today:
>
> | §4 item | Hypothesis | Measured | Verdict |
> |---|---|---|---|
> | 4.3 | conntrack exhaustion / SNAT pressure | 1,021 of 262,144; 28-day peak 113,891 | REFUTED |
> | 4.3 | listen-queue overflow | `ListenOverflows` increase 0, all 5 nodes, 8 days | REFUTED |
> | 4.5 | broker request-handler saturation / GC | brokers 60–65m CPU, node CPU 1–4% | REFUTED |
> | — | node packet pressure | `softnet_dropped_total` 0 over 28d; steal <0.5% | REFUTED |
> | — | CoreDNS slow | p99 2.69ms, 0 panics | REFUTED |
> | — | client-side DNS stalls | 1,200 probes, 3 nodes, 0 stalls ≥2s | NOT REPRODUCED |
>
> §4.6 (the 10s dial timeout) is the one that holds, and it is acted on here.

## 2. The real blocker, and the resizing it forces

**The root cause cannot be established because the bug cannot be measured.**

- No Kafka dial metric exists anywhere in the repo. Bug 003's F4 counters were
  planned and never built.
- The **primary** request lane (`platform/agentbase/agent.go`) counted nothing on
  a fetch failure — log, `sleep(1s)`, `continue`, forever.
- The pods that flake are **ephemeral spawned Jobs**. Their logs GC quickly, so
  §4.1's "baseline the rate by grepping pod logs" describes a method that cannot
  produce a durable baseline.
- **And the finding that reframes the whole case:** nothing in the fleet has ever
  served `/metrics`. `observability.NewMetricsServer` has zero callers;
  `cmd/agent-chassis/main.go` built its own mux with only `/health` and `/ready`;
  yet every spawned pod is annotated `prometheus.io/port: "9090"` +
  `prometheus.io/path: /metrics` and declares a matching containerPort. Verified
  against the live Prometheus: **zero `ai_persona_*` series exist.** Every counter
  in `platform/observability` has been dead since it was written.

So the plan is deliberately **not** a root-cause fix. It is: make the rate a
metric, then remove the confirmed amplifiers. Root-cause work resumes when there
is a number to watch.

**Decision (owner, 2026-07-26):** scope is instrumentation + client-side Go fixes;
broker-side changes go into the repo templates but are **not applied**; the case
**stays open**.

## 3. Design: one instrumented dial path

There were four independent dial configurations, none configurable, none counted:

| Site | Timeout | How |
|---|---|---|
| consumer | 10s | explicit inline `kafka.Dialer` |
| producer | **3s** | `Transport` left nil → kafka-go `DefaultTransport` |
| topic manager (8 sites) | 10s | bare `kafka.Dial` → `DefaultDialer` |
| health probe | 3s | explicit |

Collapsed into `platform/kafka/dialer.go`: `SharedDialer`,
`SharedDialerWithTimeout`, `SharedTransport`, `classifyDialErr`, `DialTimeout`.

**Decisions and why:**

- **Default 5s, not 10s.** §4.6's reasoning: a dial that cannot complete in 10s
  in-cluster is pathological, and a long timeout converts one lost SYN into a
  stalled orchestration. Env-tunable via `KAFKA_DIAL_TIMEOUT`.
- **`classifyDialErr` checks `*net.DNSError` BEFORE the generic
  `net.Error.Timeout()` test.** A resolution failure that happens to be a timeout
  is a DNS problem; folding it into `timeout` would erase precisely the
  distinction the metric exists to draw. `dns_timeout` is a separate label because
  *is the residual stall resolution or connect* is the open question.
- **Instrument inside `DialFunc`, which captures DNS + connect together.**
  kafka-go's `lookupHost` is a no-op when `Resolver` is nil, so the name is still
  unresolved on arrival and Go's `net.Dialer` resolves inside our call. The 10s
  budget the handoff saw consumed was the two combined, so measuring them together
  is correct; the `outcome` label is what separates them.
- **`SharedTransport` is a `sync.Once` singleton**, because `kafka.Transport` owns
  the per-broker connection pool. One per process means producers share
  connections instead of each holding a set.
- **`IdleTimeout` 30s → 5m, `MetadataTTL` 6s → 30s.** kafka-go's defaults are
  aggressive enough that a low-traffic agent re-dials on nearly every produce —
  pure dial volume on a path that is not always reliable. 30s is still well inside
  the Java client's 300s metadata age.
- **`broker` label strips the port**, bounding cardinality to three brokers plus
  the bootstrap service.
- **The metrics listener logs `Error`, not `Fatal`.** Losing metrics must not take
  an agent down.

## 4. The DNS finding, and the fix that would have made it worse

Brokers advertise (verified live from the broker's own `strimzi.properties`):

```
PLAIN-9092://personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc:9092
```

**Three dots**, no `.cluster.local`. Pods run `ndots:5` with a three-domain search
path, so the resolver tries the search suffixes first and reaches the working name
on the fourth attempt — three NXDOMAIN round trips, each doubled by the parallel
AAAA. Measured fleet-wide over 24h: **384,392 NXDOMAIN of 525,152 responses
(73%)**, A:AAAA exactly 1:1, ≈**7.5 queries per useful answer** against a
predicted 8 for this name shape.

> **CORRECTION — I nearly proposed lowering `ndots`, and it is strictly worse.**
> The obvious reading is "ndots:5 is too high, drop it". But the name genuinely
> needs `cluster.local` appended, so at `ndots:2` the resolver tries it as
> absolute (NXDOMAIN) *and then still* walks all three search domains — **four
> rounds instead of three**. Checking the resolution order before proposing the
> change is what caught this. Logged in `WRONG_CALLS.md`.

The correct fix is broker-side: override `advertisedHost` with the FQDN (five
dots, so tried absolute first) — one query instead of six. Documented in the
template, **not applied**.

> **[MEASURED] This is a volume reduction, not a proven cure.** 300 short-name
> lookups and 300 FQDN lookups both completed in under a second, so with healthy
> DNS the search path costs no meaningful latency. It removes packets and
> therefore exposure to loss. It has **not** been shown to fix the timeouts, and
> must not be described as the fix.

## 5. Confirmed defects fixed alongside

- `topic_manager.go` fallback broker `kafka-0.kafka-headless.kafka:9092` — **no
  such Service** (Strimzi's headless service is
  `personae-kafka-cluster-kafka-brokers`). Could never connect; could only burn a
  dial timeout before failing. Removed, and the remaining bootstrap default
  fully-qualified.
- `deployments/kustomize/base/configmap-common.yaml` — `kafka_brokers` had **no
  `:9092`**, alone among every config in the repo.
- `cmd/remote-job-spawner/main.go` — read errors `continue`d with no pause, so a
  persistently unreachable broker became a hot spin. Now the same 1s backoff every
  other consume loop uses.
- Broker `resources: {}` and `KAFKA_HEAP_OPTS=-Xms128M` with **no `-Xmx`** (heap
  falls back to ~¼ node RAM, ~15GB). Repo templates corrected using the figures
  already in this module's unreferenced `config/kafka-nodepool-cr-prod.yaml`.
  **Not applied** — applying rolls the cluster, and the checked-in Terraform state
  is empty (`serial: 1`, zero resources) so `terraform apply` is not the safe route.

## 6. Why the case stays open

The user's instruction was to close it and move it to `bugs_closed`. Flagged and
overruled by the owner, on three grounds:

1. The repo's bar for `bugs_closed` is **fixed AND live**. These are Go changes —
   inert until an image roll.
2. The **root cause is not established**. What shipped is instrumentation plus
   amplifier removal.
3. The handoff's own §5 says explicitly: *"do not close this bug because they
   shipped — the flake rate is still an infrastructure defect."*

**Close condition**, written into the bug file: after an image roll carrying
`ai_persona_kafka_dial_total`, close when
`sum by (outcome) (increase(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))`
is zero fleet-wide — or, if non-zero, when the residual is diagnosed.

## 7. Sequencing

1. ~~Instrumented dial path + metrics.~~ **DONE**, committed `95df64d63`.
2. ~~Route all dial sites; count the primary lane; serve `/metrics`.~~ **DONE**.
3. ~~Confirmed defects.~~ **DONE**.
4. ~~Tests incl. the positive control.~~ **DONE** — clean-tree build + vet + tests green.
5. ~~Broker templates, not applied.~~ **DONE**.
6. Council gate — submitted, correlation `7abe1a57-e3db-4b71-9e3a-744fbf8c24b1`.
7. **Owner's next image roll** is the gate for everything above becoming live.
8. Post-roll: verify the metric is scraped (§2 of the RUNBOOK — a zero is
   meaningless until then), baseline it, then decide on the broker-side patches.
