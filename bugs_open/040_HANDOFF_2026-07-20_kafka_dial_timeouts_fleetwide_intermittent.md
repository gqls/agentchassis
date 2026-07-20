# 040 — Kafka dial i/o timeouts are fleet-wide and intermittent: all three brokers, at least four of five nodes

**Filed:** 2026-07-20 ("bugfix 003" thread) · **Status:** OPEN, not started
**Class:** cluster network infrastructure. NOT a chassis code bug — split out of
`bugs_open/003` per its fix plan, which deliberately excludes the network layer.
**Severity:** medium on its own; multiplied by `bugs_open/003`'s platform gaps
(at-most-once consume, process-local timeout timers), each transient flake can
strand an orchestration for 30–90 min or permanently. Once 003's F2–F4 ship,
this degrades to visible retries — that is the intended division of labour.

---

## 1. What changed since 003 §3a/§3b was written (2026-07-15)

003's original network evidence was sharp: child pods on ONE node retry-looped
forever on `dial tcp 10.20.99.93:9092: i/o timeout` (broker-2 only), while pods
on other nodes were clean — a static node→broker route failure.

**That signature no longer reproduces.** Verified 2026-07-20 (~11:30Z, 12h log
window): dial timeouts now hit **all three brokers**, from **at least four of
the five nodes**, at low intermittent rates — and **broker-0 dominates**, not
broker-2. A static route fix would chase a moving target.

## 2. Evidence (2026-07-20, `--since=12h`, counts of `dial tcp <ip>:9092: i/o timeout`)

Broker topology (unchanged, all `Running` 22 days):

| broker | pod IP | node |
|---|---|---|
| combined-pool-prod-0 | 10.20.161.217 | prod-instance-17744590808031336 |
| combined-pool-prod-1 | 10.20.0.21 | prod-instance-17722135234001149 |
| combined-pool-prod-2 | 10.20.99.93 | prod-instance-17735924839006832 |

Per-pod dial-error counts by target broker:

| pod | node | b-0 (.217) | b-1 (.21) | b-2 (.93) |
|---|---|---|---|---|
| agent-build-dispatch-loop-3458d71d | …37536833 | 8 | 2 | 2 |
| agent-image-build-handler-537ecc0b | …90808031336 | 18 | 2 | 34 |
| agent-image-build-handler-47376bc9 | …35234001149 | 28 | 0 | 2 |

Totals across 8 sampled agent pods ranged 10–52 errors/pod/12h; the static
`agent-chassis` Deployment pod had 0 in the same window. Note the second row:
that pod runs **on broker-0's own node** and still times out dialling broker-0
(pod-network path, so this does not exonerate the fabric — but it does implicate
more than inter-node links).

Repro of the count:
```bash
kubectl -n ai-persona-system logs --since=12h --tail=3000 <pod> \
  | grep 'i/o timeout' | grep -o 'dial tcp [0-9.]*:9092' | sort | uniq -c
```

## 3. What this is NOT

- **Not a broker outage** — all three broker pods healthy, 22 days uptime.
- **Not the 003 §3a permanent wedge** (currently): sampled pods recover and
  produce/consume between flakes. The permanent-wedge state remains possible if
  a pod's luck runs bad, which is why 003 F4's fail-fast/self-crash matters.
- **Not consumer-group queueing** — that is `bugs_open/030` (one partition, one
  consumer, ~25–36 min dispatch latency). If your orchestration row is missing,
  read 030 before concluding drops.
- **Not the build-halt mechanism** — that is `bugs_open/029` (hung spawns
  saturate the dispatch concurrency group); this bug is one of the ways spawns
  come to hang.

## 4. Diagnostics for whoever picks this up

1. **Baseline the rate** before and after any change — the §2 grep across all
   agent pods, same window, is the metric. (003 F4 adds Prometheus counters for
   consumer dial errors; once that ships, use those instead of log-grepping.)
2. **Node-pinned probes** (003 §5a pattern): busybox `nc -vz <broker-ip> 9092`
   pinned to each node in turn, N attempts, record the failure rate — is it
   uniform per node or skewed?
3. **Conntrack pressure**: `conntrack -S` / `node_nf_conntrack_entries` vs
   limit on each node; dial timeouts that come and go under load smell of
   table exhaustion or SNAT port pressure.
4. **CNI + kubelet logs** on the worst node in the same minute as a logged
   dial timeout.
5. **Broker side**: request-handler saturation / GC pauses on broker-0
   (`kafka.network:type=RequestChannel` metrics via the Strimzi metrics
   ConfigMap if enabled) — brokers too busy to accept connections produce
   exactly this client signature.
6. Kafka client dial timeout is 10s (`platform/kafka/consumer.go:49`) — a
   dial that cannot complete in 10s on an in-cluster network is pathological
   regardless of cause.

## 5. Relationship to 003 (read this before "fixing" anything here)

003's F2 (DB-driven retry), F3 (at-least-once + completion dedupe) and F4
(honest health, self-crash, drain) are what convert this bug from
"silently strands work" to "causes visible retries". Do not block on the
network investigation to ship those; equally, do not close this bug because
they shipped — the flake rate is still an infrastructure defect, it just
stops being an outage class.
