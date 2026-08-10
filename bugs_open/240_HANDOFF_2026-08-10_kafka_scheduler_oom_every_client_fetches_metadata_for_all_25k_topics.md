# BUG 240 — kafka-scheduler OOMs every ~60s: every kafka-go client refetches metadata for ALL 24,998 topics every ~3s, and 24,087 of those topics are orphans nothing deletes

**Filed:** 2026-08-10 · found from the `bugfix_209_deploy_purpose_keyed_source` lane
(a work item filed correctly at `triaged` sat 20 minutes undispatched) ·
**Status:** OPEN, live incident, fleet-wide degradation.
**Severity:** critical — the whole scheduled layer runs at roughly a **14% duty
cycle**. Nothing alarms, because the scheduler is not down, it is *limping*.

## Symptom

`kafka-scheduler` is `CrashLoopBackOff`, **OOMKilled 132 times in 13 hours**
(memory limit `128Mi`). Each life is ~50–75 s, then the backoff climbs to its
5-minute ceiling.

`[MEASURED 2026-08-10]` one full cycle, from the pod's own `lastState`:

```
started 2026-08-10T10:20:27Z → finished 10:21:17Z  reason=OOMKilled exit=137
```

`[MEASURED]` fleet orchestration volume, `orchestration_states` by hour: **150–330
per hour** through 08-09, falling to **6–17 per hour** across 08-10 04:00–10:00Z.
That ratio is what a ~14% duty cycle predicts, and it is the damage: every timer
on the platform — the build pipeline, the three site-discovery rotations, every
reaper — fires only in the seconds the scheduler happens to be alive.

**This is not a scheduler bug.** The scheduler is merely the process with the
smallest memory limit, so it is the first to die.

## Root cause — three independent facts that only bite together

### 1. `producerTransport` leaves `MetadataTopics` blank, which means ALL topics

`platform/kafka/dialer.go:155-163` builds the shared producer transport. It sets
`Dial`, `DialTimeout`, `IdleTimeout`, `MetadataTTL` — and **not**
`MetadataTopics`. kafka-go v0.4.47's own doc on that field
(`transport.go:84-86`) states the consequence outright:

> Topic names for the metadata cached by this transport. **If this field is left
> blank, metadata information of all topics in the cluster will be retrieved.**

This is not an inference about what the library probably does; it is the
library's documented behaviour for the value we pass.

### 2. That refetch is a background loop averaging ~3 seconds, not 6

`transport.go:589-640`, `(*connPool).discover`:

```go
metadataTTL := func() time.Duration {
    return time.Duration(prng.Int63n(int64(p.metadataTTL)))   // rand in [0, 6s)
}
req := &meta.Request{TopicNames: p.metadataTopics}            // nil ⇒ ALL topics
```

It loops forever on that timer, in its own goroutine, whether or not the process
produces a single message. **The mean interval is ~3 s, not the 6 s the
`MetadataTTL` name suggests** — the TTL is the *upper bound* of a uniform random
draw. Anyone reasoning about refresh cost from the constant alone will be out by
2×.

### 3. The cluster holds 24,998 topics / 50,100 partitions, 24,087 of them orphans

`[MEASURED 2026-08-10]` via `kafka-topics.sh --list` / `--describe` on
`personae-kafka-cluster-combined-pool-prod-0`:

| what | count |
|---|---|
| topics, total | **24,998** |
| partitions, total | **50,100** |
| `job.*` (ephemeral per-step topics) | **24,087** |
| `system.*` | 861 |

`job.*` topics are created per orchestration step
(`platform/kafka/topic_manager.go:79`, `job.<corr>.<orch>.<agent>.<step>`). They
are ephemeral by design and **nothing deletes them** (see §4).

So every kafka-go process in the fleet fetches and unmarshals metadata for 50,100
partitions roughly every 3 seconds, in the background, forever. In a 128Mi
container with no `GOMEMLIMIT` and default `GOGC=100`, the heap target is twice
the live set — and the live set is now large enough that the target exceeds the
limit before a GC completes.

`[MEASURED]` cgroup trace of one life (`memory.current`, sampled from inside the
pod), against the process's own log:

```
10:26:29  process start
10:26:36  cur=44Mi  peak=65Mi     ← t=0 tick finishes; LAST log line until 10:26:59
10:26:43  cur=108Mi peak=109Mi    ← +64Mi in 7s with the process logging NOTHING
10:26:49  cur=111Mi peak=113Mi
10:26:56  cur=114Mi peak=115Mi
10:27:02  cur=117Mi peak=121Mi
10:27:08  POD GONE (OOMKilled)
```

The 64Mi step lands in a window where the scheduler emitted no log line at all —
it was idle between ticks. Nothing in `runTick` runs there. The only thing
executing is `discover()`.

### 4. Why the orphans accumulate: the cleanup only fires when the WHOLE fleet is idle

`deployments/kustomize/services/agent-job-cleanup/agent-job-cleanup-cronjob.yaml`
runs every 10 minutes and is **not** suspended (`[MEASURED]` last successful
`2026-08-10T10:40:24Z`). It deletes `job.*` topics only on an **idle tick** —
both of these empty:

```sh
kubectl get jobs -n ai-persona-system -l 'spawned-by in (orchestrator, remote-job-spawner)'
kubectl get pods -n ai-persona-system -l 'spawned-by in (orchestrator, remote-job-spawner)'
```

and then only the topics that were *already* orphaned on the previous tick
(a two-pass tombstone). `[MEASURED]` sampling that exact predicate every ~7 s
across 2026-08-10 10:48:15–10:50:27Z: **18 of 18 samples BUSY**, 4–6 matching
jobs/pods at every sample, a mix of genuinely `Running` and `Succeeded`-but-unreaped.
(A 2-minute window cannot prove the fleet is *never* idle; what it establishes is
that idleness is not the common case. The 24,087 surviving topics are the
long-run evidence.)

**A live platform is never simultaneously idle, so the branch that deletes topics
is, in practice, dead** — and it needs to win twice in a row.

That guard is the fix for `bugs_closed/071`, where the *opposite* failure was
biting: the guard never matched a pod, so the cronjob deleted every `job.*` topic
every 10 minutes, killing live agents mid-run. **071 traded an over-aggressive
cleanup for one that never runs, and the second failure is silent** — a deleted
live topic kills a visible run, whereas an undeleted dead topic costs nothing
until 24,000 of them do.

`[INFERRED, not measured]` accumulation therefore starts at 071's fix
(2026-07-26). 24,087 topics over ~15 days ≈ 1,600/day ≈ 67/hour, which is the
right order for this platform's step volume — but I did not date the topics, and
Kafka does not record topic creation time, so treat the start date as consistent
rather than established.

## Why nobody noticed

`[MEASURED]` other kafka-go services carry the same cost and survive it, because
their limits are higher: `agent-chassis` 146–147Mi, `thunder-adapter` 135Mi,
`render-audit-adapter` 135Mi, `business-intel` 130Mi. **`kafka-scheduler` is the
only pod in the namespace with elevated restarts.**

That asymmetry is the positive control, and it is what rules out a
scheduler-specific cause: if the fault were in the scheduler's own queries or
tick logic, the other services would not be sitting at an elevated baseline. They
are. The scheduler is just the canary with a 128Mi limit.

## Fix candidates — ordered by what closes the door

1. **Set `MetadataTopics` on the shared transport, or stop using one transport
   for dynamic topics.** This is the door: it makes the cost independent of how
   many topics exist. ⚠ Needs care and a council round — writers here produce to
   *dynamic* `job.*` topics, and kafka-go fetches those on demand
   (`transport.go:418,438` `refreshMetadata(ctx, topicsToRefresh)`), so a static
   list may be viable, but **that on-demand path must be verified before anyone
   relies on it.** Do not ship this on the strength of this paragraph.
2. **Give the topic reaper a rule that does not require fleet-wide idleness** —
   e.g. delete a `job.*` topic whose correlation's orchestration is terminal, or
   simply one older than N hours with no live Job of that name. Closes the
   accumulation door without reintroducing 071 (which was caused by deleting
   topics of *running* agents; age/terminality is a per-topic test, not a
   fleet-wide one).
3. **Delete the existing 24,087 orphans once**, to end the live incident.
   One-off; without (1) or (2) it recurs. ⚠ Must not delete topics of live runs —
   use the same per-topic test as (2), not a bulk wipe.
4. **`GOMEMLIMIT` + a bigger limit on `kafka-scheduler`.** Mitigation only, and
   deliberately last: it makes the symptom go away while the metadata cost keeps
   growing with every new topic, so it buys time and hides the trend. Worth doing
   *alongside* 1–3, never instead.

## How to verify a fix

- `kafka-topics.sh --list | grep -c '^job\.'` falls and stays down.
- `kafka-scheduler` restart count stops advancing (it is the sensitive detector —
  it dies first).
- Fleet orchestration volume returns to 100+/hour:
  `SELECT date_trunc('hour',created_at), count(*) FROM orchestration_states
   WHERE created_at > now() - interval '6 hours' GROUP BY 1 ORDER BY 1;`
- ⚠ **Do not verify by "the scheduler is up"** — it is up for ~60 s at a time
  today. Check the restart count is not advancing, over at least 10 minutes.

## Diagnosis provenance (OWNER RULING 2026-07-31)

`090` was **not** run. Substituted first-hand verification, stated plainly as the
ruling requires:

- The deciding arm was **read, not inferred** — kafka-go's own field
  documentation for the value we pass, plus the `discover()` loop that issues it.
- The three quantities the claim rests on were **measured live** (topic count,
  partition count, per-life cgroup trace) and one was measured against the
  cronjob's **own predicate**, run verbatim.
- A **positive control** distinguishes the platform-wide cause from a
  scheduler-specific one (other kafka-go pods elevated; only the smallest-limit
  pod restarting).

The claim I am *not* making: I did not capture a Go heap profile — the binary
exposes no `pprof` endpoint (`cmd/scheduler/main.go:96-108` registers only
`/health` and `/ready`). The attribution of the 64Mi step to `discover()` rests
on it landing in a window where the process ran nothing else and logged nothing,
not on an allocation trace. If someone wants that last inch, add pprof and
profile at t=45s.

## Related

- `bugs_closed/071` — the cleanup guard whose fix created this accumulation. Read
  it before touching the cronjob; it documents exactly how deleting live topics
  breaks running agents.
- `bugs_open/040` — kafka dial timeouts fleet-wide. Same transport, same
  `ProducerTransport()`; a metadata refresh every ~3s over 25k topics is a dial
  and traffic source that any dial-rate analysis there must account for.
- `platform/kafka/dialer_test.go:195` `TestProducerTransportMatchesKafkaGoDefaults`
  pins the transport to kafka-go's defaults **including the blank
  `MetadataTopics`**. Fix candidate 1 must update that test deliberately — it is
  designed to stop exactly this field being changed by accident.
