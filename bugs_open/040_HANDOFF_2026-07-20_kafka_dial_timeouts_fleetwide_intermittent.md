# 040 — Kafka dial i/o timeouts are fleet-wide and intermittent: all three brokers, at least four of five nodes

> **NUMBER COLLISION (2026-07-20, same day):** another thread filed a different
> `040` (`failed_page_build_leaves_page_deployed_and_partially_composed`, now in
> `/bugs_closed/`). Numbers are never reassigned — resolve by slug
> (`bugs_closed/README.md` duplicate-numbers table). Cite this case as
> **040-kafka-dial**.

**Filed:** 2026-07-20 ("bugfix 003" thread) · **Status:** **OPEN** — instrumentation
**LIVE and COLLECTING 2026-07-26** (v1.0.1167 + PodMonitor + NetworkPolicy fix, all
applied); root cause still NOT established; awaiting 7 days of metric.
**2026-07-27: the first timeout has been measured — §10.** One in ~22h, broker
prod-2, on a spawned pod. §10 also corrects §9's close condition, which used
`increase()` and returned **0** for that very event.
**2026-08-12: the week of data is in and does NOT close the case — and a second,
much bigger residual was found — see §11.** 32 `timeout` in 7d (condition 1 fails
outright) plus **71,832 `refused` in a single ~38h burst**, invisible to
`agent_error_log`, filed to the diagnosis loop (verdict **UNVERIFIABLE** — its
tools cannot reach Prometheus or the vendored kafka-go internals). One narrowly
scoped hardening shipped regardless (§11.4): `topic_manager.go`'s `getController`
no longer builds a dial target from an empty `Host`.
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

**§4.6 is the one that held** — but see the CORRECTION in §6: the timeout cut it
argued for was proposed, vetoed by the council gate, and **reverted**. §4.6 is a
remark in a bug file, not a measurement, and the histogram this change ships is
what will actually settle the right value.

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
~2h of uptime. Brief, self-healing, and it costs the full 10s each time.

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
   in-cluster network is pathological regardless of cause. **PROPOSED, VETOED,
   REVERTED — still 10s.** See the CORRECTION in §6. This remains the most
   plausible lead, and the histogram is what should choose the replacement value.

## 5. Repro of the original count

```bash
kubectl -n ai-persona-system logs --since=12h --tail=3000 <pod> \
  | grep 'i/o timeout' | grep -o 'dial tcp [0-9.]*:9092' | sort | uniq -c
```

## 6. Amplifiers found and removed (2026-07-26, commit `95df64d63`)

None of these is claimed to be the cause. Each made the fault worse or hid it.

- **Four inconsistent, uncounted dial configurations** — consumer 10s (explicit),
  producer **3s** (`Transport` left nil → kafka-go `DefaultTransport`), topic
  manager 10s × 8 sites (bare `kafka.Dial`), health probe 3s. **All four are now
  COUNTED and all four keep exactly the budget they had** — see the CORRECTION
  below. They are not yet unified.
- **`/metrics` never served** — now on the health mux *and* on `METRICS_PORT`
  (9090), the port the annotations already advertise.
- **Phantom fallback brokers at THREE sites** — `kafka-*.kafka-headless*:9092`
  names no Service that exists. Strimzi's headless service is
  `personae-kafka-cluster-kafka-brokers`; the only `kafka-headless` manifest in the
  repo (`deployments/kustomize/infrastructure/kafka/kafka.yaml`) belongs to a
  hand-rolled StatefulSet that **no kustomization references**, and
  `kubectl get svc -A | grep headless` is empty. Every such entry could only burn a
  full dial timeout before failing over. All three fixed, each fully-qualified:

  | site | what it had | how it was found |
  |---|---|---|
  | `platform/kafka/topic_manager.go` | 1 phantom + 1 unqualified duplicate | reading the file |
  | `platform/orchestration/actions/spawn_actions.go:1019` | 1 phantom + 1 unqualified duplicate | **the council gate** |
  | `internal/core-manager/admin/agent_handlers.go:766` | **3 phantoms, no valid entry at all** | an untruncated re-grep |

  The third is the worst and was nearly missed twice. With `KAFKA_BROKERS` unset,
  core-manager had **no route to Kafka whatsoever** — three consecutive dial
  timeouts and then failure. It also read only `KAFKA_BROKERS`, missing
  `SERVICE_INFRASTRUCTURE_KAFKA_BROKERS` (the variable spawned pods receive), so it
  reached that dead list more readily than the other two. It now uses
  `kafka.GetBrokers()`, the existing helper that checks both in the right order.

  > **Method note worth more than the fix:** site 2 was found only because the
  > council asked "did you enumerate the siblings?", and site 3 only because the
  > grep I ran to answer that was piped to `head` and silently truncated at 10
  > lines. **A capped enumeration is indistinguishable from a complete one.** If you
  > are establishing "these are all the sites", do not pipe to `head`.
- **`kafka_brokers` with no port** in `deployments/kustomize/base/configmap-common.yaml`.
- **Hot spin** in `cmd/remote-job-spawner/main.go` — read errors `continue`d with
  no pause. Now the standard 1s backoff.
- ~~**Producer connection churn** — `IdleTimeout` 30s / `MetadataTTL` 6s made
  low-traffic agents re-dial almost every produce.~~ **REVERTED, still 30s / 6s.**
  Real, but retuning it changes failover reactivity fleet-wide and no measurement
  chose the new values. See the CORRECTION below.

> **CORRECTED 2026-07-26, same day — the council gate REJECTED the first version
> of this fix (hard veto from `guardian`, 2× HIGH, corr `7abe1a57`) and it was
> right.** The change as first written also cut the dial timeout **10s → 5s for
> every process in the fleet** and raised the producer's `IdleTimeout` 30s → 5m
> and `MetadataTTL` 6s → 30s. Those are behaviour changes to shared messaging
> plumbing and to failover reactivity across every pipeline, bundled into what was
> presented as instrumentation — and both were argued from a remark in this file
> and from the Java client's defaults, **not from measurement**, while building
> the instrument that would have measured them.
>
> **All of it is reverted.** `InstrumentedDialer(timeout)` takes the caller's own
> existing budget and the package has **no default at all**; `ProducerTransport`
> reproduces kafka-go's `DefaultTransport` byte-for-byte (3s dial / 5s DialTimeout
> / 30s IdleTimeout / 6s MetadataTTL) and stays package-level because a nil
> `Transport` already shared one pool per process. A test pins those three values.
>
> **Sequenced, not abandoned.** Once the counters ship,
> `ai_persona_kafka_dial_duration_seconds` says what the real dial latency
> distribution is, and a timeout can be chosen from that. That is part 2 and needs
> the architecture review the guardian asked for. Logged in `WRONG_CALLS.md`.

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

## 8b. COVERAGE LIMIT of the new metric — do not read it as fleet-wide

`cmd/agent-chassis` is the **only** binary that serves `/metrics`. Surveyed
2026-07-26, all 21 `cmd/*` binaries: agent-chassis YES, the other 20 no.

They all import `platform/kafka`, so their dial counters do increment in-process —
and are then thrown away, because nothing exposes them.

**What the metric therefore covers:** the chassis Deployment **and every spawned
agent pod** (spawned Jobs run `./agent-chassis`, verified from live pod specs).
That is deliberately the population 040 is about — the handoff's own §2 shows the
spawned pods carrying 10–52 errors per 12h while the static chassis Deployment had
**0** in the same window.

**What it does NOT cover:** the 13 adapter/service Deployments
(`web-scrape-adapter`, `git-adapter`, `reasoning-agent`, `core-manager`,
`kafka-scheduler`, `remote-job-spawner`, …). Their dials stay invisible.

So a zero from this metric means "no timeouts **on the chassis and spawned
agents**", not "none in the fleet". Say it that way when reporting, and cross-check
the adapters with the §5 log grep until they are wired too. Extending the pattern
to the other binaries is the obvious follow-up and was deliberately NOT bundled
here — bundling is what got round 1 vetoed.

## 8d. Council trail (advisory) — three rounds, no trailer

Correlation `7abe1a57-e3db-4b71-9e3a-744fbf8c24b1`, all three rounds under it:

| round | decision | `decided_by` |
|---|---|---|
| 1 | **REJECTED** | hard veto from `guardian` (2× HIGH) |
| 2 | **REVISE** | gating objection from `editquality` |
| 3 | **REVISE** | ⚠ `unreadable reviewer(s): review_editquality.result` |

**Round 3's REVISE was a dead seat, not an objection** — on substance it was 6
approve / 2 object, with the guardian down from a hard veto to one MEDIUM and two
LOWs it filed "rather than blocking". Read `decided_by` before the decision; the
same harness failure killed 3 of `bugs_closed/029`'s 5 rounds.

**There is no `Council-Reviewed:` trailer on any commit and none may be added** —
it is earned by an APPROVED verdict only, and on a REVISE it would be a permanent
false claim of review. The `098` coverage report will therefore list this work as
un-reviewed for ever. That is a known, accepted false negative; this table is the
record instead.

What the gate actually bought, which is the reason to keep using it:

- **Vetoed a fleet-wide dial-timeout cut** (10s→5s) plus producer `IdleTimeout`
  30s→5m and `MetadataTTL` 6s→30s, all of which I had argued from a remark in this
  very file and from another library's defaults — while building the instrument
  that would have measured them. All reverted; sequenced behind the histogram.
- **Forced the reuse question** on `observability.NewMetricsServer`, which turned
  out to serve a hardcoded-200 `/health` — the `bugs_open/003 §4.1` defect, one
  call away from being reintroduced.
- **Its sibling-enumeration objection led to two more phantom-broker sites**, one
  of which left core-manager with no route to Kafka at all (§6).

Two objections stand unanswered by code, deliberately, because neither needs a code
change: the guardian's note that my *plan* contradicted itself about the producer
pool (the code is correct — both `kafka.Writer{}` sites carry `ProducerTransport`,
verified at HEAD), and `prior_art`'s request for a runnable enumeration query
(now RUNBOOK §8c).

## 8e. LIVE VERIFICATION 2026-07-26 — the metric works, and it took THREE fixes

The chassis rolled (`v1.0.1167`, pod started 17:11:30Z). **Note the tag did not
change**, so the tag proves nothing — only the pod-grep does.

**Step 1 — is the code live?** Yes, with both controls:

```
strings /app/agent-chassis | grep -c ai_persona_kafka_dial_total              -> 1   (mine)
strings /app/agent-chassis | grep -c ai_persona_kafka_messages_produced_total -> 1   (positive control)
strings /app/agent-chassis | grep -c ai_persona_kafka_dial_nonexistent_xyz    -> 0   (negative control)
wget -qO- http://localhost:9090/metrics                                        -> SERVES
```

Spawned agent pods carry it too (`app=dynamic-agent`, all on v1.0.1167).

**Step 2 — was it collected? No, and it took two more fixes.** Serving the port was
necessary and not remotely sufficient. Three layers, each of which alone produces
exactly the same symptom — a metric reading zero:

| # | layer | state before | fix |
|---|---|---|---|
| 1 | nothing served `/metrics` | port closed fleet-wide, forever | listener in `cmd/agent-chassis` (this roll) |
| 2 | Prometheus had no way to discover the pods | operator-driven, 0 PodMonitors; `prometheus.io/*` annotations inert | `podmonitor.yaml`, **APPLIED** |
| 3 | `allow-monitoring` NetworkPolicy matched nothing | every scrape `context deadline exceeded` | source selector fixed, **APPLIED** |

Layer 3 is the sharpest of the three. The policy already named **port 9090** — the
intent was unambiguous and someone wrote it deliberately — but its `from:` was a
bare `podSelector` (meaning *this* namespace, not `monitoring`) matching a label
(`app: prometheus`) the pod does not carry (it is `app.kubernetes.io/name`). Either
defect alone makes it select nothing. Discriminating evidence:

```
from ai-persona-system  -> curl pod:8080/health  OK    curl pod:9090/metrics  OK
from monitoring         -> curl pod:8080         OK    curl pod:9090          TIMES OUT
```

Result: targets went **0 UP / 6 DOWN → 6 UP / 0 DOWN**, and Prometheus went from
**0 `ai_persona_*` series to 16**. Not just this case's metric — `agent_tasks_processed`,
`messages_processed`, `workflows_started`, `agent_health` and the rest all came
alive. They had been incremented and discarded since the day they were written.

**Step 3 — the first fleet-wide dial baseline ever taken** (~20 min, 6 pods):

```
sum by (outcome) (ai_persona_kafka_dial_total)   -> ok = 240      (timeout = 0, dns = 0, error = 0)
histogram_quantile(0.99, ...dial_duration...)    -> 0.0279  (27.9 ms)
by broker: prod-0 = 20, prod-1 = 21, prod-2 = 24, bootstrap = 175
```

**240 dials, 100% `ok`, p99 27.9ms.** Label cardinality is the designed 4 series.

> **This is a baseline, NOT grounds to close.** 20 minutes is not 7 days, and the
> handoff's own evidence had the *static* Deployment at **0** errors in a window
> where spawned pods carried 10–52 — so a clean short sample from a mostly-static
> fleet is exactly what this bug looks like when it is not currently firing.

**What it does settle:** the deferred timeout question now has data. **p99 is 27.9ms
against a 10s budget — roughly 360×.** Whatever the intermittent stall is, it is not
the normal distribution creeping up on the limit; it is a distinct, rare event. Part
2 can choose a timeout from this histogram instead of from an opinion, which is
precisely what the council's veto preserved.

## 9. CLOSE CONDITION (explicit, 2026-07-26)

The changes above are **Go code and are inert until an image roll**. Do not close
on them.

1. ~~Roll an image carrying `ai_persona_kafka_dial_total`; verify against the
   running pod with controls.~~ **DONE 2026-07-26** — §8e step 1.
2. ~~**Prove the metric is scraped.**~~ **DONE 2026-07-26** — §8e step 2. Took two
   further fixes (PodMonitor + NetworkPolicy), both applied. 16 `ai_persona_*`
   series now in Prometheus, up from zero.
3. Baseline it, and cross-check once against the §5 log grep before retiring the
   grep.
4. **Close when** `sum by (outcome) (increase(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))`
   is zero **across the chassis and spawned agents** (see §8b — that is not the
   whole fleet, and the claim must not be written as if it were) — **or**, if non-zero, when the residual is diagnosed
   (the `dns`/`dns_timeout` vs `timeout` split is designed to answer exactly that
   next question: resolution or connect).

   > **CORRECTED 2026-07-27 — condition 4's query is WRONG and would have closed
   > this bug on a day a timeout actually happened.** See §10. Use
   > `sum(max_over_time(...{outcome="timeout"}[7d]))`, not `increase(...)`.

Per handoff §8 and the owner's ruling of 2026-07-26, amplifier fixes alone are
**not** grounds to close.

## 10. FIRST MEASURED TIMEOUT — and the close condition that could not see it (2026-07-27)

Recorded by the bug-backlog triage sweep, 2026-07-27 ~15:53 UTC, the first read of
this metric since the 07-26 baseline. **The bug is not gone: the instrument caught
one.** All figures below are from the live Prometheus
(`prometheus-kube-prometheus-stack-prometheus-0`, container `prometheus`,
`/api/v1/query`).

```
sum(max_over_time(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))  -> 1
sum(max_over_time(ai_persona_kafka_dial_total{outcome="ok"}[7d]))       -> 22312
count by (outcome) (max_over_time(ai_persona_kafka_dial_total[7d]))     -> ok, timeout   (2 label values)
```

The single timeout, with its full label set:

```
outcome  = timeout
broker   = personae-kafka-cluster-combined-pool-prod-2.personae-kafka-cluster-kafka-brokers.kafka.svc
pod      = agent-page-rerender-af558880-9mwxw     (a SPAWNED agent, app=dynamic-agent)
job      = ai-persona-system/agent-chassis
```

Samples exist at exactly three scrapes, all at value `1`, `1785138799 /
1785138859 / 1785138919` epoch — **≈07:53 UTC on 2026-07-27**, then the pod died
and the series ended. `up` for `ai-persona-system/agent-chassis` = **5 targets**,
so scraping is healthy and this is a real observation, not a coverage hole.

**Rate, stated with its denominator and its window.** ~1 timeout against 22,312
successful dials — but the metric only goes back **~22 hours** (earliest sample
in Prometheus ≈2026-07-26 17:52 UTC, i.e. the roll that first served `/metrics`).
So: **1 dial timeout in ~22h across the chassis Deployment and every spawned
agent pod**, on a broker-side event lasting the full 10s budget. This is **not**
7 days of data and §9's condition still is not met. It does settle one thing —
`p99` dial duration over 24h is **9.98 ms** (down from the 27.9 ms first
baseline), against a 10s budget, so the intermittent stall remains a distinct
rare event and nothing like the normal distribution creeping up on the limit.

### The trap, which is the transferable half

**`increase()` returns 0 on this timeout. `max_over_time()` returns 1. Both are
correct; the close condition asks the wrong one.**

```
sum(increase(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))   -> 0      <-- §9.4 as written
sum(max_over_time(ai_persona_kafka_dial_total{outcome="timeout"}[7d])) -> 1   <-- the truth
```

Why: **the pods that carry this metric are ephemeral spawned Jobs, and a counter
on an ephemeral target is frequently born at its final value.** Prometheus never
scraped this series at `0` — its first sample was already `1` — so there is no
0→1 step for `increase()` to find, and `increase()` over a range whose samples
are all `1` is `0` by definition. The condition in §9.4 would therefore have
reported "no timeouts in 7 days" on a day one demonstrably occurred. This is the
same shape as §7's uptime trap (a window longer than the pod's life silently
understates) one level up: **a rate function longer than the target's life
silently understates to zero.**

Corrected close condition for §9.4:

```promql
sum(max_over_time(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))       # total, survives short-lived pods
count(max_over_time(ai_persona_kafka_dial_total{outcome!="ok"}[7d]))         # how many distinct pods saw a non-ok outcome
```

`increase()`/`rate()` stay correct for the long-lived chassis Deployment and are
still the right tool for the `ok` denominator on it — they are wrong for the
spawned-agent population, which §8b says is precisely the population this bug is
about. **Say which population any figure covers.**

### What this does NOT establish

- **It is one event.** It does not distinguish `timeout` from the `dns` /
  `dns_timeout` outcomes the labels were designed to separate — those buckets are
  empty, which is itself informative: this was a *connect* failure, not
  resolution. [MEASURED: `count by (outcome)` returns only `ok` and `timeout`.]
- **It does not name a cause.** Broker prod-2 with everything else clean matches
  the shape §1 already described (all brokers, low intermittent rate) and refutes
  nothing in §2.
- **It says nothing about the 13 adapter/service Deployments** (§8b). They still
  serve no `/metrics` and their dials remain invisible. A "1 in 22 hours" figure
  is the chassis + spawned agents only.
- **§4.2 (node-pinned `nc` probes) is still unexercised** and is still the most
  promising untried lead — now with a named broker to aim at.

## 11. The week is in (2026-08-12): §9 does not close, and a much bigger residual — `refused` — was found alongside the rare `timeout`

Picked up as an unowned, past-due bug (17 days since §10, past its own 7-day
checkpoint; `who-owns.py 040` matches only the *other* 040, a false positive from
the number collision at the top of this file). Re-ran §9's close condition live.

### 11.1 §9 does not close — condition 1 fails outright

```
sum(max_over_time(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))  -> 32
```

Non-zero. The rare, self-healing single-event stall §10 first caught is still
happening (32 in the most recent week), so this alone means §9 cannot close on
condition 1 and must fall to condition 2 (diagnose the residual).

### 11.2 A second outcome, two orders of magnitude bigger, that did not exist in §10's snapshot

[MEASURED 2026-08-12, Prometheus, `sum by (outcome)`, **not** `count by (outcome)`
— see the misstep below]:

| outcome | 17d total |
|---|---|
| `ok` | 1,129,512 |
| `timeout` | 48 |
| `dns_timeout` | 300 |
| `dns` | 3 |
| `error` | 1 |
| **`refused`** | **71,832** |

`refused` (`ECONNREFUSED`, per `classifyDialErr`) is empty in §10's 07-27
snapshot — it is new. It dwarfs every other non-`ok` outcome combined.

> **MISSTEP, caught same session before it was written down anywhere durable:**
> the first pass used `count by (outcome)(max_over_time(...))`, which counts
> DISTINCT SERIES (how many pods hit that outcome at least once), not the
> summed counter value. It returned `refused: 212` — plausible-looking, wrong by
> **340×**. Re-ran with `sum by (outcome)` before citing anything. Logged in
> `WRONG_CALLS.md` and in the RUNBOOK (§10) so the next session reaches for
> `sum` first.

### 11.3 Bisected timing, and broker restarts ruled out first

Widening `sum(max_over_time(...{outcome="refused"}[Nh]))` from the current time
(RUNBOOK §11) shows: **zero in the last 20h**; all 71,832 accumulated between
roughly **2026-08-10 00:47Z and 2026-08-11 16:47Z** (bursty, quiet
06:47–18:47 on 08-10), then nothing since. All three broker pods are 45 days old
with **0 recent restarts** — the brokers did not bounce during this window, so
this is not "the broker went away mid-roll".

### 11.4 The mechanism read from the code, and the fix that closes it regardless of whether it is THE cause

Almost all of the 71,832 carry an **empty `broker` label**
(`sum by (broker)` → `{}` 71,826 vs bootstrap-only 6). `brokerLabel()`
(`dialer.go`) returns the address unchanged only when `net.SplitHostPort`
**errors** — but `net.SplitHostPort(":9092")` succeeds with host=`""`. An empty
label therefore means the dial's *address* was `:9092`, not that the label was
missing.

`platform/kafka/topic_manager.go`'s `getController()` (pre-fix, line 299) built
its dial target as `fmt.Sprintf("%s:%d", controller.Host, controller.Port)`
straight from kafka-go's `Controller()` metadata response with no validation. If
`controller.Host` comes back empty, this produces the literal string `:9092`,
which as a client-side TCP dial resolves to the pod's **own loopback**, where
nothing listens — an instant `ECONNREFUSED` (no 10s wait, which is why a single
pod can rack up 800–1,300 of them). `platform/orchestration/actions/spawn_actions.go:1032`
already filters the literal string `:9092` out of a *different* broker list —
someone already learned this string appears, without connecting it to this bug.

**Filed to the diagnosis loop rather than shipped as a guessed fix** (per
CLAUDE.md's diagnosis-before-debugging: cross-cutting, cause not obviously where
the symptom is). `090_TRIGGER_needs_diagnosis_v1.sh`, intake correlation
`04195fa7-28c2-410a-a8cb-15d42acf43c4`, run correlation
`39bb6fe8-a55c-476e-8ffd-026bec4b57ca`. **Verdict: UNVERIFIABLE** — not a
refutation. The loop independently confirmed, from real `agent_error_log` rows,
that `getController`/`CreateTopic` genuinely does produce `failed to connect to
controller: ... i/o timeout` failures fleet-wide (matching this bug's original,
already-known signature) — but it has no tool to query Prometheus and no
visibility into kafka-go's vendored internals, so it could not confirm or rule
out either candidate mechanism for the *specific* `refused`/empty-broker burst:
(a) `getController` itself, or (b) kafka-go's internal consumer-group
coordinator lookup (used by `consumer.go`'s `Reader`, same `InstrumentedDialer`),
which is not in this repo's code index. Full verdict JSON in the workstream
`NOTES`.

**Own follow-up check, first-hand:** searched `agent_error_log` for `refused`
fleet-wide across the burst window, and for the exact pod names Prometheus named
as top offenders. **Zero rows mention Kafka, dial, or connection-refused at
all** — this failure is completely invisible to application logs; only the
metric ever saw it. (The same pods' only log lines in that window are an
unrelated `message validation failed` delivery error at 02:25–02:26Z on
2026-08-10, inside the first burst episode — noted as a time correlation, not
claimed as the same defect.)

**Shipped anyway, narrowly scoped:** `getController` now rejects an empty
`Host` (`controllerAddress` helper, unit-tested) instead of silently formatting
it into `:9092`, and logs a `Warn` when it does — closing the "silent" half of
this finding regardless of which candidate mechanism turns out to be the real
one, and adding the log visibility that was missing either way. **This is not
claimed to close the bug** — mechanism (b) is untouched and unconfirmed, and the
burst has not recurred in 20h to test against. `go build`/`vet`/`test`/`test
-race` all clean on `platform/kafka/...`. Commit: pending council submission.

**Left for whoever picks this up next:** if `refused` recurs, the new `Warn` log
will say whether it came through `getController` (this fix's guard will fire and
the pod will retry a *different* broker instead of dialing garbage) — settling
mechanism (a) vs (b) the next time it happens, which the diagnosis loop could
not settle from history alone.

### 11.5 A third candidate the council surfaced, checked and only partially fits

The council submission's `editquality` seat flagged a directly-relevant landmine
this session had not checked: `platform/kafka/dialer.go`'s `producerTransport`
leaves `MetadataTopics` blank, so every kafka-go client in the fleet fetches
metadata for **every topic in the cluster** (~25,000 as of 2026-08-10) roughly
every 3s — `bugs_open/240` (kafka-scheduler OOM), same transport. A metadata
storm at that scale is a plausible contributor to broker-side load that could
produce malformed/empty controller responses under pressure.

**Checked the timing rather than assuming a shared cause.** 240's incident window
runs through its last recorded OOM at `2026-08-10T11:45:56Z`, resolved by
`12:02Z` (topic count 24,131 → 354). Burst episode 1 here (~00:47–06:47Z on
08-10, ~20,255 of the 71,832 `refused`) **does overlap** — plausible partial
contributor. Burst episode 2 (~18:47Z 08-10–16:47Z 08-11, ~51,577 events, **72%
of the total**) **starts 7 hours after 240's incident was already resolved**,
topic count already down at 354. The metadata-storm mechanism cannot explain
the majority of this burst. Contribution recorded in `bugs_open/240` rather than
duplicated here; both bugs stay open.
