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
no longer builds a dial target from an empty `Host`. **Council APPROVED, and the
fix is now PROVEN LIVE on `v1.0.1291`** (§11.6) — a follow-up diagnosis run
auditing the rest of `platform/kafka` for the same pattern was killed mid-flight
by that same chassis roll and produced no verdict; re-fired, see §11.6.
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

### 11.6 Fix proven live on a fresh chassis build; the follow-up audit run was killed by that same roll

`v1.0.1291` rolled (both `agent-chassis` pods, 2026-08-12). Proven live via the
OCI revision label (image cached locally): `docker image inspect
docker.io/aqls/agent-chassis:v1.0.1291 --format '{{index .Config.Labels
"org.opencontainers.image.revision"}}'` → `da5a7eb8f`; `git merge-base
--is-ancestor e1f960ac2 da5a7eb8f` → yes, zero intervening commits touched
`topic_manager.go`. Cross-checked against the running binary with controls
(`strings /app/agent-chassis`): the new log line `controller metadata invalid,
trying next broker` and the new error text `controller metadata returned an
empty host` both present (≥1); a nonsense negative control absent (0); a
known-good pre-existing string present (≥1).

A follow-up `090` diagnosis run was filed before the roll (correlation
`e91d71d6-058a-4902-a852-c6d54bc7411c`), asking whether the same
unvalidated-Host-from-a-kafka-go-response pattern this fix guards against
exists anywhere else in `platform/kafka` or its callers — a direct answer to
the council's `guardian` (enumerate `getController`'s callers exactly) and
`prior_art_librarian` (reuse vs. duplicate validator) objections, both
advisory and non-blocking on the APPROVED verdict. **That run was killed
mid-flight by the same chassis roll** — it was cycling normally through
several verdict rounds until ~14:38Z, then sat at `load_runtime` with no
`last_activity` movement for 100+ minutes; its `site_work_items` row is now
`status='failed'` (the documented 40-minute claim-timeout landing on a
terminal state per `max_attempts=1`, consistent with the owning pod being
replaced under it, not a content failure). **No verdict was produced; this is
not evidence for or against the audit question.** Re-fired as a fresh intake,
correlation recorded in the workstream NOTES once it lands.

### 11.7 Re-fired run came back UNVERIFIABLE (search capped, not refuted); first-hand grep settles both objections

The re-fire (correlation `58a0390c-33ec-4580-9697-3320b280475d`) completed
this time, but with outcome `UNVERIFIABLE`. Its own `needed_evidence` field
named the exact gap: the `fmt.Sprintf("%s:%d", ...)` content search that would
surface a sibling instance was capped at 40 rows ordered by path and cut off
before reaching `platform/orchestration/actions/spawn_actions.go`, so absence
there was UNKNOWN, not confirmed — this is the class in
`an-unverifiable-verdict-does-not-say-your-premise-was-false` (memory), and
here the verdict was explicit enough about *why* that filling the gap
first-hand was cheaper than a third diagnosis run.

Ran the two follow-ups the verdict itself asked for, uncapped:

- `grep -rn 'fmt.Sprintf("%s:%d"' --include=*.go platform/ internal/ cmd/` —
  **one hit, fleet-wide**: `topic_manager.go:322`, inside `controllerAddress`
  itself — the fixed function, which now rejects an empty `Host` (§11.4)
  before this line ever runs. No sibling instance of the idiom exists.
- `grep -rn 'kafka.Broker{' --include=*.go platform/ internal/ cmd/` — **zero
  hits**. No struct-literal construction of a `kafka.Broker` anywhere.
- Every live-metadata call site checked directly: `conn.Controller()` occurs
  exactly once (`topic_manager.go:292`, inside `getController`, the fixed
  path). `conn.Brokers()` occurs exactly once
  (`platform/health/kafka_reachability.go:115`, inside `dialAny`), and its
  result is **discarded** (`_, metaErr := conn.Brokers()`) — never formatted
  into a dial target. Nothing here shares the vulnerability class.
- Read both `:9092`-literal filters in full —
  `topic_manager.go:CreateJobTopic` (lines 84-118) and
  `spawn_actions.go:getConfiguredKafkaBrokers` (lines 1017-1057+). Both
  operate on **statically-configured** broker strings (the `brokers []string`
  function parameter and `os.Getenv(...)` respectively, split on `,`) — never
  on a live kafka-go response. They guard a different failure (a phantom
  hostname baked into a fallback list, per the comments already at those
  sites) and do not participate in the Host-from-live-metadata pattern this
  fix closes.

**Answer to the council: no, the pattern does not exist anywhere else.**
`getController` (via `controllerAddress`) is the only site in the fleet that
builds a dial target from a live kafka-go `Controller`/`Brokers` response, and
it is the one already fixed. The `:9092` filters are a distinct, unrelated
mechanism (static-config hygiene, not live-metadata validation) — so
`prior_art_librarian`'s reuse-vs-duplicate question dissolves: there is
nothing of the same kind to reuse or duplicate. `guardian`'s caller
enumeration is confirmed exactly as the capped run already found:
`CreateTopic`, `DeleteTopic`, `ListTopics`, `TopicExists` — four call sites,
no fifth.

This closes the loop the two follow-up diagnosis runs were meant to close.
The bug itself (§11's `refused` burst) stays OPEN per §11.6 — this subsection
only answers "is the fixed pattern isolated", not "is this fix the confirmed
cause of the burst".

### 11.8 A fresh chassis roll (v1.0.1293), and the first post-fix `refused` read — silent, but not yet informative

Another owner-initiated chassis roll landed 2026-08-12 ~19:13Z: `v1.0.1293`,
git commit `7a1887e31`. `git merge-base --is-ancestor e1f960ac2 7a1887e31` →
yes, so the §11.4 fix is still live (it is also an ancestor of current `HEAD`,
5 commits further on — this tree moves fast, per usual). Both pods restarted,
so per the standing rule no orchestration/council run was fired for ~5 minutes
after.

**Read the `refused` counter fleet-wide, bisecting the same way as §11.3:**

```
sum(max_over_time(ai_persona_kafka_dial_total{outcome="refused"}[Xh]))
```
gives `0` for X≤26, jumps to 19,616 at X=28, 33,056 at X=34, 51,577 at X=48,
and flattens at 71,832 (the full known total) by X=72 — i.e. **the last
`refused` sample fleet-wide landed between 26h and 28h before this read
(2026-08-12 ~19:19Z), which lands inside the already-known episode 2 window
(~16:47Z on 08-11).** Confirmed the read itself isn't blind: `outcome="ok"`
and `outcome="timeout"` both carry real, nonzero counts over the same 12h/24h
windows, and `count(ai_persona_kafka_dial_total)` shows live series — the zero
is a genuine "no samples", not a broken query (the §11 runbook's own
documented trap).

**The honest reading of that zero, and why it is weaker than it looks:**
`agent-chassis:v1.0.1291` (carrying the fix) has been running continuously
since its replicaset was created at `2026-08-12T14:55:10Z` — confirmed via
`kubectl get rs -l app=agent-chassis` (old RS still present, scaled to 0,
timestamp intact even after the 1293 roll replaced it). **That is 22h08m
*after* the last known `refused` event (~08-11 16:47Z).** So of the ~26.5
hours this counter has now been silent, only the last **4h24m** of that
silence happened while the fix was actually running — the other ~22 hours of
quiet came *before* the fix ever shipped, on the unfixed binary. The metric
was already capable of a day-long quiet stretch with the bug still present
(episode 1 → episode 2 themselves were ~12h apart, but nothing says the gaps
are regular): **this read cannot yet distinguish "the fix stopped it" from
"it hadn't recurred yet regardless."** [UNVERIFIABLE-BY-THIS-METHOD, not
REFUTED-nor-CONFIRMED]

One weak, explicitly-flagged data point in the fix's favour: the current gap
since the last `refused` event (~26.5h) already exceeds the only interval
between two episodes observed so far (~12h, episode 1 end → episode 2 start).
With only two prior episodes this is not a rate, and treating it as one is
exactly the mistake named in `two-clean-runs-cannot-establish-stability`
(memory) — noted, not relied on.

**What would actually be informative:** a `refused` count taken once a
meaningful multiple of the observed inter-episode gap has elapsed *with the
fix continuously live* — say, re-read this same query no earlier than
2026-08-13 ~15:00Z (24h into the fix's runtime) and again a few days out. If
`refused` stays at 71,832 (no increase at all) across a span several times
longer than any gap seen between episodes 1 and 2, that becomes real evidence
the fix is the cause, not merely quiet timing. Recorded here so the next
session re-reading this bug doesn't mistake today's silence for that
evidence.

---

## CONTRIBUTION 2026-08-12 (dispatch/pool lane, not this bug's owner) — layer 2 is only HALF fixed: the PodMonitor discovers the spawned pods and has never discovered the static `agent-chassis` Deployment

**Filed here rather than fixed-and-owned, because `podmonitor.yaml` and this bug belong to
this lane** (`who-owns.py` names it, and this lane committed to it today). The one-line
manifest change is already committed — `889a7c055`, `base/deployment.yaml` — because a
metric that is live and unscraped was blocking my own lane's deliverable (SYS-091). Nothing
else here is actioned; the judgement calls below are yours.

**What the layer table above gets right, and the residue it does not cover.** Layers 1–3 are
sound and the fixes work. But layer 2's remedy discovers pods by a **numeric**
`targetPort: 9090`, and the operator compiles that to:

```
- action: keep
  source_labels: [__meta_kubernetes_pod_container_port_number]
  regex: "9090"
```

read verbatim out of `/etc/prometheus/config_out/prometheus.env.yaml` on the live
Prometheus, job `podMonitor/ai-persona-system/agent-chassis/0`. That arm matches the port a
pod **DECLARES in its spec**, not the port it serves. Spawned pods declare 9090
(`spawn_actions.go`, named `metrics`) and are matched. The static Deployment declared only
`containerPort: 8080` while serving `:9090` perfectly — so it passed the `app` label arm and
was dropped at the port arm. It therefore never became a target at all: no DOWN row, no
scrape error, nothing to alert on.

`[MEASURED 2026-08-12, v1.0.1293]`

```
active targets, all scrape pools ............................ 141
  ... whose pod name matches agent-chassis-* ................   0
both chassis pods, fetched directly on :9090 ................ SERVE (all nine go_sql_* series)
  go_sql_max_open_connections{db_name="clients_db"} .........  12
```

**Why this was invisible for 17 days, and why it is worth your attention specifically.**
The `job` label is the **PodMonitor's name**, so all ~108 spawned-agent targets carry
`job="ai-persona-system/agent-chassis"`. A target list read for that job — or any
`ai_persona_*` query — returns a healthy, plentiful, entirely real answer that contains zero
rows from the long-lived service. My lane checked the reader *before* building an instrument,
exactly as this bug teaches, and still recorded "both chassis pods are `health:"up"`" into a
handoff and into the concept register. That claim is now struck through in SYS-091; the
`WRONG_CALLS.md` entry is dated today.

**The bit you should decide on:** line 410 above reports the layer-2/3 fix taking targets to
**6 UP / 0 DOWN** and `ai_persona_*` series to 16. On the evidence above, those 6 cannot have
included the static chassis pods — the Deployment has declared only 8080 continuously until
today's commit. `[INFERRED — not retrospectively checkable]` they were spawned
`app=dynamic-agent` pods. If that is right, this bug's closing evidence, and the "19
`ai_persona_*` series" reading my lane quoted from it, describe **ephemeral agent Jobs only**,
and the fleet's longest-lived service has been contributing nothing to either figure. That
does not undo layers 1–3; it means the success criterion was measured on a population that
excluded the service most people picture when they read this bug. Worth re-reading whatever
you concluded from `ai_persona_*` counts.

**A second fact, unrecorded anywhere until now:** `podmonitor.yaml` is **not in the kustomize
build**. `base/kustomization.yaml` lists only `deployment.yaml`, and
`kubectl kustomize deployments/kustomize/services/agent-chassis/overlays/production/uk_001`
renders ConfigMap + Deployment + Role + RoleBinding and **no PodMonitor**. The object is live
because it was hand-applied per its own header comment ("NOT APPLIED. Apply it alongside the
image roll…"), which is now stale — it *is* applied. So it is committed, live, and reconciled
by nothing: an `apply -k` will not recreate it if it is ever deleted, and file-vs-live drift
is silent in both directions. Wiring it into `resources:` is the obvious repair and I have
deliberately **not** done it, because it changes what a whole-fleet release applies and that
is your call, not mine.

**Verify after the next roll** (the deployment change is inert until then):

```bash
kubectl -n monitoring exec prometheus-kube-prometheus-stack-prometheus-0 -c prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=go_sql_max_open_connections%7Bpod%3D~%22agent-chassis-.*%22%7D'
```

Two chassis pods reading `12`. **An empty vector is the disconfirming result** and is what
this returns today — so the check can come out either way, which is the only reason it is
worth running. Full trap in `LANDMINES.md` ("a PodMonitor's numeric `targetPort` keys on the
port a pod DECLARES").

---

## 2026-08-15 ~15:26–15:34 UTC — two "topic partition has no leader" produce failures in one sweep (contributed by the vigilant_designer lane)

Fresh occurrences, both inside improvement-sweep corr `12b85f92-7808-4f2b-9684-acd636ce43aa`
(gaswholesalers.com, hand-fired 15:05):

1. **15:26:48** — `page-content-writer` child FAILED at `complete_workflow`:
   `kafka.(*Client).Produce: fetch request error: topic partition has no leader`
   (topic=`job.12b85f92-012885b6-page-build-handler-process_item_iter_2_spawn_handler.responses`,
   partition=1). Took one `content_rewrite` work item to `failed` (attempt 1 of 3); the two
   sibling items dispatched minutes earlier completed normally.
2. **15:34:39** — the **head `improvement-loop` row itself** FAILED at its terminal
   `complete_workflow`: same error class (topic=`system.generic.responses`, partition=2).
   Every substantive step had already committed (audit ran, findings written, items promoted
   and dispatched, audit state recorded) — the FAILED status records only the final response
   write. A reader judging sweeps by head-row status alone will count this sweep as a failure;
   the work is all there.

Two different topics — one dynamic per-job, one long-lived — within eight minutes, both
`no leader` on Produce. Pattern-matches this bug's intermittent broker episodes rather than
anything in the sweep's own config (the same chain end-to-end had run clean at 15:05–15:12).
Evidence lives in `orchestration_states` error columns while retention lasts; quoted verbatim
in `vigilant_designer_offer_analysis/NOTES` 2026-08-15.

## CONTRIBUTION 2026-08-21 (bugs_open/343 lane, not this bug's owner) — the two error surfaces are DISJOINT, and this file names the one that sees <1%

Arrived here from `bugs_open/343`: two `availability-discovery-agent` orchestrations, 27 minutes
apart on 2026-08-20, both dying at the terminal step on a Kafka write —
`step complete failed: failed to execute action complete_workflow: failed to send response: failed
to write message to kafka: Kafka write errors (1/1)`. That is the same `complete_workflow` class this
file's 2026-08-15 section already describes, so **this is not a new mechanism** — it is a measurement
correction, and it bears on that section's closing line.

**The finding.** `[MEASURED 2026-08-21]` For Kafka errors, `agent_error_log` and
`orchestration_states.error` are **completely disjoint**:

| surface | orchestrations |
|---|---|
| `agent_error_log` only | **125** |
| `orchestration_states.error` only | **1** |
| **both** | **0** |

**Not a retention artefact** — re-run restricted to the window both tables cover gives the identical
125 / 1 / 0. So **this file's line "evidence lives in `orchestration_states` error columns while
retention lasts" names the surface that holds 1 of 126.** Any rate or blast-radius figure for this
bug should come from `agent_error_log`; a census on `orchestration_states.error` sees **under 1%** of
instances and will read as "rare and improving" regardless of what is happening.

**SHARPENED 2026-08-21 (029/343 lane, second session) — and the sharper form is the one to act on.**
The disjointness above is **specific to this error class**; the two sinks are *not* disjoint in
general. Whole retained window, `[MEASURED]`: **23,230** distinct orchestrations have
`agent_error_log` rows, **22** have `orchestration_states.error` set, and **9 of those 22 are in
both**. So they do overlap — and the real point is the denominator, not the overlap:

> **`orchestration_states.error` is populated for ~0.1% of what `agent_error_log` covers (22 vs
> 23,230). It is not a lossy mirror of the other sink; it is barely populated at all, and it is not a
> rate-measurement instrument for ANY class.**

That makes the 1-of-126 above a **worked consequence rather than a special case**, and it is the
sentence to put in front of anyone re-measuring this bug — because **this file's own
"How to verify" hands them the near-empty sink**, where a zero ends the investigation. Related
blindness in the same table, worth reading together: `agent_error_log` has **no column joining a
parent orchestration to its child** (see `bugs_open/343`), so "what actually failed here" is hard from
either direction.

**The worked pair, which is how it was found and why it is easy to get backwards.** The two
orchestrations recorded the *same* failure in *different* tables, and neither appears in both:

| orchestration | `orchestration_states.error` | `agent_error_log` |
|---|---|---|
| `efa24e6e` (15:41Z) | **names Kafka** | **0 rows** |
| `00525861` (16:08Z) | reaper only — `stale EXECUTING_STEP for >4h; step=complete` | **1 row, names Kafka** |

I first read this as "one failed loudly, one failed silently", because I checked
`orchestration_states.error` for both and inferred absence from the second. **Both logged it.** A
per-instance judgement about whether a failure was silent is unsafe unless both surfaces are checked.

⚠ **Retention trap met on the way, worth having here too.** `min(created_at)` on
`orchestration_states` reads **2026-07-19**, which suggests month-long retention. It is not:
**CANCELLED** rows (24) appear never to be pruned, while `COMPLETED` starts ~26 h back and `FAILED`
likewise. So the ~26 h figure this estate uses is right for the statuses that matter, and a naive
`min()` will tell you otherwise. Verified in the direction that mattered: **0 of 21** of 343's
08-17 wedged orchestrations survive there.

---

## 2026-08-21 — the WRITE side finally has an instrument, the empty-host dial is refused, and an opt-in retry stops a seconds-long blip terminating finished work

Picked up as an unowned past-due bug (§11.8's own "re-read no earlier than 08-13" is nine days stale).
**Nothing here closes §9.** Two of the four rounds below are instrumentation, one is behaviour, one is
classification; the `timeout` residual and the `refused` mechanism both stay open, and §9's close
condition is untouched.

### 12.1 The premises, re-measured today — and the bug is HOTTER than §11 left it

`[MEASURED 2026-08-21, live Prometheus, `max_over_time` not `increase()` per §10's trap]`

| reading | §11 (08-12) | today | note |
|---|---|---|---|
| `timeout` 7d | 32 | **146** | prod-0 dominant (74), then bootstrap 53, prod-1 10, prod-2 9 |
| `refused` 7d | 71,832 all-time | **94,419** | **it RECURRED after the §11.4 guard shipped** |
| …of which EMPTY broker label | 71,826 | **85,887** | the `:9092` self-dial signature |
| `ok` 7d | — | 965,438 | the denominator, so the rates above are ~0.015% and ~9.8% |

The last burst ended **~24 h before this read** (1,028 in `[24h]`, 14,104 in `[48h]`, **nothing** in
`[18h]`), carried by long-running spawned consumer pods — `agent-build-dispatch-loop`,
`agent-landmine-verifier`, `agent-feed-ingester`, ~690 events each.

**Why the recurrence matters more than the count.** §11.4's fix makes `getController` **structurally
unable** to emit an empty-host address, and §11.6 proved it live on `v1.0.1291`. So the bursts since
are **not** that producer. At least one more exists and it is **not in this repository**: kafka-go's
consumer-group coordinator lookup builds `net.JoinHostPort(out.Coordinator.Host, …)` straight from a
FindCoordinator response with **no validation** (`consumergroup.go`, verified in the module cache), and
every `Reader` in the fleet dials through `InstrumentedDialer`. That is mechanism (b) from §11.4,
now with a named line. `[INFERRED, not confirmed — the burst pods are GC'd, so §11.4's armed Warn-log
discriminator cannot be read retrospectively. §12.3 is what will settle it.]`

### 12.2 The finding §11 could not have made: the WRITE side was never instrumented, and writes are failing

The dial counters see a **connection** fail. Nothing had ever seen a **write** fail.
`[MEASURED 2026-08-21, `agent_error_log`, retained window 07-22..08-21]`

| | count |
|---|---|
| rows matching `Kafka write error` | **63** |
| rows matching `has no leader` | **40** |
| **distinct orchestrations affected** | **93** |

Recurring most days since 08-10 (08-11: 14, 08-14: 14, 08-15: 16, 08-18: 8, 08-20: 3). Steps hit:
`complete`/`complete_workflow` **11**, `process_message` 10, `orchestrate` 9, and ~20 `call_agent`
dispatch steps. **So this is not only the terminal step** — the 2026-08-15 section above saw one shape
of a wider class.

**And kafka-go will NOT retry it, which is the fact the whole round turns on.** Its writer loops to
`MaxAttempts` (default 10) with backoff but **breaks after ONE attempt** on
`!isTemporary(err) && !isTransientNetworkError(err)` (`writer.go`), and `protocol.ErrNoLeader` — the
client-side `"topic partition has no leader"` — is a **bare string type whose only method is
`Error() string`**. No `Temporary()`, so no retry. That is the `Kafka write errors (1/1)` fingerprint,
and **the `(1/1)` is messages-failed / messages-sent, not attempts** — two different counters and one
very plausible misreading. Full trap in `LANDMINES.md`.

**Compounded downstream:** `platform/errors`' nine transient needles matched **neither** string, so all
103 rows classified `error_unrecoverable` and terminated work permanently, on a condition usually over
in seconds.

### 12.3 What shipped, in four commits and three council rounds

| # | commit | what | state |
|---|---|---|---|
| 1 | `e4ce7073b` | `ai_persona_kafka_produce_total{topic_class,outcome}` + the `empty_host` dial refusal + PodMonitor wired into kustomize | **LIVE v1.0.1322** |
| 2 | `9b93af8a0` | round-2 fixes: closed `system.*` family set; `no_leader` split into `client_no_leader`/`broker_no_leader` | committed, rides the next roll |
| 3 | `<this round>` | opt-in bounded produce retry + 4 adopters + 2 classifier needles (SYS-093) | committed, rides the next roll |

**Proven live at the artefact, with controls that could have come out otherwise.** On the running
`v1.0.1322` pod: `ai_persona_kafka_produce_total` PRESENT, `refusing dial to structurally invalid
address` PRESENT, a pre-existing positive control PRESENT, a nonsense negative control ABSENT, **and a
literal from a commit made *after* the build ABSENT** — that last one is what shows the probe
discriminates rather than answering PRESENT to everything.

**And collecting:** `sum by (outcome) (max_over_time(ai_persona_kafka_produce_total[1h]))` →
**`ok = 99`** within the first hour. That is the **demand control**, not a nicety: a zero there would
mean the instrument is broken, not the fleet clean.

**The `empty_host` discriminator, and the disconfirming result named in advance.** On the next burst:
`sum(max_over_time(ai_persona_kafka_dial_total{outcome="empty_host"}[48h])) > 0` confirms the remaining
producer is library-internal. **`refused` carrying an EMPTY broker label must now be structurally
zero — a non-zero DISCONFIRMS**, meaning a third `:9092` constructor exists outside the instrumented
dial path, and that is where to look next. (Both read `0` in the 2 h after the roll, which is expected
and proves nothing: no burst has occurred in that window.)

### 12.4 The retry, and exactly what it does and does not promise

Opt-in (`kafka.WithRetry` / `ProduceWithRetry`), **default OFF**, adopted at **four named sites**:
`CompleteWorkflowAction`, `notifyParentOfFailure` (closing its asymmetry with `notifyParentOfSuccess`,
which has been on the shared reply seam since `bugs_open/133` while this stayed a bare fire-once
log-and-drop), `notifyParentOfSuccess`, and `processor.go`'s single response produce exit. The five
other `DeliverReply` callers are byte-identical and a test pins that.

- **Bounded and stated:** 4 attempts, jittered 500 ms→4 s, **~44 s worst case** before a reply is
  finally reported undeliverable. Long, and the right trade against losing a completed workflow.
- **Deterministic errors are never retried** — validation refusal (`bugs_open/274`), too-large (whose
  remedy is to DEGRADE), context-cancelled.
- **⚠ It CAN duplicate a reply** after a lost ack: kafka-go v0.4.47 has no idempotent producer. The
  premise is that the parent's two-phase `ClaimAwaitedRequest` absorbs it — which is why the adopters
  are restricted to sites whose consumer is a parent orchestration with that dedupe, and why the retry
  is **not** inside `Produce`, where 39 call sites with unaudited consumers would inherit it.
  **Named disconfirming signal: `DUPLICATE_SKIPPED` volume rising after the roll.**
- **Two needles admitted, a third refused.** `"kafka write error"` and `"no leader"` (the short form —
  `kafka.WriteErrors`' `Error()` embeds its members' texts, so the long form would miss the composite
  that actually arrives). **DELIBERATELY REFUSED: `"write message to kafka"`** — our own wrapper on
  *every* write failure including the deterministic ones, so admitting it would reclassify permanent
  failures as retryable, which is exactly `bugs_open/274`. A control test asserts both stay TERMINAL.
- **⚠ This changes what RE-RUNS**, and consumers are told rather than merely measured (owner ruling
  2026-07-29 §3): a child hitting either string may now be **redispatched by its parent** (capped at
  `retry_version >= 3`) instead of failing terminally. Head rows with no parent still terminate.

### 12.5 What the council caught, because two of three rounds found real defects

- **Round 1 (corr `a414d81b`) REVISE, and both code objections were RIGHT.** (HIGH) `topicClass`'s
  `case strings.HasPrefix(topic, "system."): return topic` **returned its raw input** — precisely the
  "substring of the input" case the plan's own cardinality rule forbids; I stated the rule and broke it
  in the next arm, then flagged it in the *risks* for a reviewer to confirm rather than resolving it.
  Measured afterwards: of **937** live `system.*` topics, 859 are caught by the `system.agent` arm and
  **78 would have reached that arm as distinct labels**, with `system.errors.<agent-type>` (18) and
  `system.responses.<agent-type>` (17) growing per agent type. Now a **closed** family set with an
  `system.other` bucket. (MEDIUM) `no_leader` collapsed the client-side and broker-side errors, which
  **behave oppositely** inside kafka-go — split, so a reader can tell *exhausted immediately* from
  *retried and still failed*. (MEDIUM) the PodMonitor wiring is orthogonal scope: fair, and **not
  reverted**, because it was an explicit owner decision and forward-only forbids an amend.
- **Round 2** resubmitted on the same correlation (env-var form — passing `RESUBMIT_CORR` positionally
  is what produced a malformed trail id earlier the same day; `WRONG_CALLS.md`).

### 12.6 What is STILL OPEN, stated so nobody reads this section as a close

1. **The `timeout` residual — 146 in 7 days, prod-0 dominant, undiagnosed.** §4.2's node-pinned `nc`
   probes remain **unexercised** and are still the most promising untried lead, now with a named broker
   to aim at. Remember §7's traps: normalise by pod uptime, brokers live in namespace `kafka`, and
   busybox `date +%s%N` returns 0.
2. **The `refused` mechanism is `[INFERRED]`, not confirmed.** `empty_host` on the next burst is the
   cheap answer; a `090` run is only worth firing if a burst arrives *before* that label is live, and
   then the Prometheus evidence must be handed to it inline (the loop cannot reach Prometheus — that is
   what made §11.4's verdict UNVERIFIABLE).
3. **The 13 adapter/service Deployments still serve no `/metrics`** (§8b). Every figure in §12 covers
   the chassis and spawned agents only. Extending them is its own round and was deliberately not
   bundled — bundling is what got round 1 vetoed back in July.
