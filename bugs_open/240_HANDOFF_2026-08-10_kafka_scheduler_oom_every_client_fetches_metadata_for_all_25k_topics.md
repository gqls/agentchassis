# BUG 240 — kafka-scheduler OOMs every ~60s: every kafka-go client refetches metadata for ALL 25,042 topics every ~3s, and 24,131 of those topics are orphans nothing deletes

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

### 3. The cluster holds 25,042 topics, 24,131 of them orphaned `job.*`

`[MEASURED 2026-08-10]` via `kafka-topics.sh --list` on
`personae-kafka-cluster-combined-pool-prod-0`, **listed to a file inside the pod
and read back** — three consecutive reads returned identical counts. Do NOT pipe
`--list` down a `kubectl exec` stream; it truncates silently (§4 retraction):

| what | count |
|---|---|
| topics, total | **25,042** |
| `job.*` (ephemeral per-step topics) | **24,131** |
| `system.*` | 861 |

`[UNMEASURED]` **total partition count.** I originally recorded 50,100 here, from
`--describe`. Withdrawn — I could not obtain a figure I trust, for a reason worth
recording (see §4a). Sampling shows `job.*` topics are single-partition, so the
topic count is the right order-of-magnitude driver; if you need the partition
total, take it from the broker's own JMX metrics rather than from
`kafka-topics.sh` output.

`job.*` topics are created per orchestration step
(`platform/kafka/topic_manager.go:79`, `job.<corr>.<orch>.<agent>.<step>`). They
are ephemeral by design and **nothing deletes them** (see §4).

So every kafka-go process in the fleet fetches and unmarshals metadata for all
25,042 topics roughly every 3 seconds, in the background, forever. In a 128Mi
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
that idleness is not the common case. The 24,131 surviving topics are the
long-run evidence.)

**A live platform is rarely simultaneously idle, so the branch that deletes
topics almost never fires** — and it must win twice in a row to delete anything.

> **RETRACTED 2026-08-10, same session — and the retraction is the more useful
> finding.** I briefly corrected this line to say the reaper "fires
> occasionally", citing a fall from 24,087 (10:44Z) to 23,966 (10:54Z) and then
> 22,300. **Every one of those numbers was an artefact.** Piping
> `kafka-topics.sh --list` down a `kubectl exec` stream **truncates silently** at
> this topic count — no error, no non-zero exit, just a short list. Three reads
> 18 seconds apart returned **21,409 / 23,017 / 5,809**.
>
> `[MEASURED, reliable method]` listing to a file *inside* the pod and reading it
> back gives the same answer every time: **25,042 topics total, 24,131 `job.*`,
> 861 `system.*`** — identical across three consecutive reads.
>
> So the original claim stands, and my "correction" of it was noise. What I can
> say from reliable measurement is the standing backlog, not a rate: **~24,131
> `job.*` topics survive after the reaper has been running every 10 minutes for
> ~15 days.** Whether it ever fires is *unresolved* and now cheaply measurable —
> take two in-pod counts either side of a tick. Do not repeat my mistake of
> reading a trend out of the piped form.

### 4a. Two counting traps, and one of them I caused

Both cost me real time and both produce a **plausible wrong number with a zero
exit code**, which is the only reason they belong in a bug file at all.

**Trap 1 — piping `kafka-topics.sh` output truncates.** True both down a
`kubectl exec` stream *and* inside the pod. `--list | grep -c '^job\.'` returned
21,409 / 23,017 / 5,809 on three reads 18 s apart; `--list | grep | sort -u > f`
inside the pod produced **445** where the real figure was 24,131. **Always
redirect to a file first, then process the file.** The working incantation:

```bash
kubectl -n kafka exec <broker> -- bash -c \
  'bin/kafka-topics.sh --bootstrap-server localhost:9092 --list > /tmp/t.txt 2>/dev/null
   echo "total=$(wc -l < /tmp/t.txt) job=$(grep -c "^job\." /tmp/t.txt)"'
```
Three consecutive reads agreeing to the row is the check that it worked.

**Trap 2 — the broker's `/tmp` is a 5 MB tmpfs, and I filled it.** The topic list
is ~1.8 MB and a `--describe` dump is ~3 MB; together they exhausted it. **A full
`/tmp` makes `kafka-topics.sh --list > file` write ZERO BYTES and still exit 0.**
That is what produced my "445" and "0" readings, and — importantly — it is also
the real explanation for the `--describe` output "ending mid-record", which I had
started to write up as a property of the tool. It was ENOSPC of my own making.

`[CORRECTED]` so: `--describe` is **not** known to truncate. I have no evidence
either way, because every observation I have of it was taken against a full disk.
The withdrawn 50,100 stays withdrawn — not because the tool is unreliable, but
because *my measurement* was, and I cannot separate the two after the fact.

Clean up after yourself on that broker; `scripts/kafka-orphan-topic-sweep.sh`
now checks free space before writing and removes its files on exit.

That guard is the fix for `bugs_closed/071`, where the *opposite* failure was
biting: the guard never matched a pod, so the cronjob deleted every `job.*` topic
every 10 minutes, killing live agents mid-run. **071 traded an over-aggressive
cleanup for one that never runs, and the second failure is silent** — a deleted
live topic kills a visible run, whereas an undeleted dead topic costs nothing
until 24,000 of them do.

`[INFERRED, not measured]` accumulation therefore starts at 071's fix
(2026-07-26). 24,131 topics over ~15 days ≈ 1,600/day ≈ 67/hour, which is the
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
3. **Delete the existing ~24,000 orphans once**, to end the live incident.
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

---

## OUTCOME 2026-08-10 12:02Z — incident ENDED, root cause CONFIRMED causally, bug stays OPEN

Fix candidate 3 run with owner authorisation
(`scripts/kafka-orphan-topic-sweep.sh --apply`): **23,781 topics deleted, 0 batch
failures.**

`[MEASURED, reliable method — three consecutive in-pod reads, identical]`

| | before | after |
|---|---|---|
| topics, total | 25,042 | **1,265** |
| `job.*` | 24,131 | **354** |

354 remaining = the 350 protected by the liveness test + a handful created since.

**The causal chain is now confirmed by dose–response, not just by reading code.**
Scheduler memory tracked the topic count down in three steps, with nothing else
changed — no deploy, no config edit, no restart of anything but the scheduler's
own crashloop:

| topic count | scheduler memory | behaviour |
|---|---|---|
| 24,131 | 121Mi peak | OOMKilled every ~50–75 s, 143 restarts |
| mid-sweep | 58Mi | survived 11 minutes |
| **354** | **15Mi** | **no OOM; restart count frozen at 143** |

Last OOM `2026-08-10T11:45:56Z`. `[MEASURED]` fleet orchestration volume recovered
17 → 37 → 43/hour while the sweep ran, and every short-interval scheduled task
returned to **≤1.0 intervals overdue** (they were at 4–10.5 during the incident).

This is the strongest form of confirmation available here and it is worth naming:
the original diagnosis was read out of the library's source and documentation,
and the prediction it made — *reduce the topic count and the memory falls with
it* — then held across three points without anything else being touched.

### Why this bug stays OPEN

**Nothing yet stops it recurring.** The sweep is a one-off. `MetadataTopics` is
still blank, the per-step topics still have no reliable reaper, and accumulation
restarts immediately. Fix candidates 1, 2 and 4 are all still owed; only 3 is
done. Re-measure the topic count in a week — if it is back in five figures, the
door is still open and this file will tell you why.

⚠ **Do not close this on the strength of "the scheduler is up".** That is exactly
the observation that would have been true for 60 seconds at a time all through
the incident.

### `[MEASURED]` the accumulation restarted immediately — a rate, not a guess

Two in-pod counts, same session, after the sweep:

| time | `job.*` topics |
|---|---|
| 12:02Z (sweep complete) | **354** |
| 12:39Z | **458** |

**+104 in 37 minutes ≈ 170/hour**, under the load of one lane re-rendering eight
sites. That is not a steady-state figure — it is what this platform produces when
something real is running, which is the condition that matters.

At that rate the cluster is back to five figures inside a week, which is exactly
the interval over which it got there the first time. **This is the concrete
argument for fix candidates 1 and 2**: the sweep bought roughly a week, and
nothing about the mechanism has changed. Anyone tempted to close this because the
scheduler looks healthy should re-run the count first — the number is the answer,
and it takes one command.

---

## 2026-08-10 ~17:10Z — fix candidate 1 (MetadataTopics) VERIFIED AND REFUTED in its naive form. Read this before writing that patch.

The owner chose the MetadataTopics fix (with GOMEMLIMIT and a scheduled sweep;
the per-topic reaper deliberately not taken). The prerequisite verification —
"does kafka-go fetch metadata on demand for topics outside a static list?" — is
now done, **from the library source, both arms**, and the answer forbids the
obvious patch:

**Setting `MetadataTopics` on the SHARED transport would break every `job.*`
producer — i.e. the whole orchestration fabric.**

The deciding arms, kafka-go v0.4.47:

1. **Metadata requests are served from the transport's cache, not the broker.**
   `transport.go:351-378` (`connPool.roundTrip`): a `meta.Request` is answered by
   `filterMetadataResponse(m, state.metadata)`; a topic absent from the cached
   state gets `UnknownTopicOrPartition` **without any broker round trip**, unless
   `AllowAutoTopicCreation` is set on the request.
2. **Our writers do not set auto-create.** `grep AllowAutoTopicCreation
   platform/kafka/*.go` → no hits, so it is false. `Writer.partitions`
   (`writer.go:744-768`) resolves every produce's topic through exactly this
   path. With a static list, **every produce to a topic outside it fails
   immediately**, and no request ever leaves the process.
3. **Auto-create would not save it either.** On that arm the broker IS asked, but
   the follow-up `refreshMetadata(ctx, topics)` (`transport.go:450-492`) waits
   for the topic to appear in `state.layout` — and state is fed by `discover()`,
   whose request is built **once** with the static `TopicNames`
   (`transport.go:600-603`). The dynamic topic can never appear, so the wait
   backs off exponentially until the produce context dies. Failure mode: every
   produce to a new topic burns its full deadline.

**What survives of the fix:**

- **A scoped transport is safe only for a service whose produce-topic set is
  closed.** `kafka-scheduler` qualifies (its topics = `scheduled_tasks.
  target_topic` values, readable at boot; guard the runtime-added-row case by
  recreating the producer on an UnknownTopicOrPartition). The chassis and every
  spawned agent do NOT qualify — they produce to per-step `job.*` topics.
- **The fleet-wide lever is therefore the TOPIC COUNT, not the fetch scope** —
  kafka-go v0.4.47 cannot scope dynamically. The reaper/sweep is not a stopgap
  for the shared transport; it is the only fleet-wide control available at this
  library version. (An upgrade may change this — unverified.)

**Corollary `[INFERRED, not proven]`:** blank list + 6s TTL + no auto-create
means a produce to a **just-created** topic fails with `UnknownTopicOrPartition`
until the next `discover` tick (mean ~3 s). That is a clean mechanism for the
known "spawn→call handshake fails ~half the time" intermittent
(`bugs_open/029`/`040` family) — the spawn creates the topics via a different
connection, and the caller's first produce races the discover tick. Flagged for
those lanes; not asserted here, and this file's fix must not silently change
that timing without them knowing (raising the shared `MetadataTTL` would WIDEN
that race window — do not "optimise" it as part of this fix).

---

## 2026-08-11 morning — C3 CONFIRMED LIVE; C4 has a blind spot nobody named; the overnight topic drop is unattributed

- **C3 is LIVE**: the overnight roll shipped v1.0.1284 with
  `GOMEMLIMIT=192MiB` + 256Mi limit visible in the deploy env; scheduler at
  10Mi, 0 restarts. `[MEASURED]` deploy env + `kubectl top` 09:45Z.
- **C4 HAS NEVER FIRED, and its 00:17 slot may NEVER fire.** The crontab entry
  is installed (RELOAD logged 18:05 BST 08-10) but `~/kafka-sweep-240.log`
  does not exist, and the journal has NO entries at all in the 00:15–00:20
  window — this machine was ASLEEP at 00:17 local. **User crontabs get no
  anacron catch-up**, so a slept-through slot is skipped silently, every
  night the machine sleeps. Practical consequence: only the 12:17 slot is
  real, so the effective cadence is ONCE daily, and "check the log" can read
  as delivery-failure when it is simply a slot that never happened. First
  possible real firing: 12:17 BST 2026-08-11 — verify the log exists after.
- **Topic count 106 at 09:45Z, down from 1,236 at 16:33Z 08-10 —
  `[UNVERIFIED]` what swept.** Not the C4 cron (above: never fired). No
  commit touched the sweep script; no session recorded a manual run in this
  file. At ~129/h steady creation, 106 is consistent with a sweep <1h before
  the reading — i.e. around the morning roll. Most likely a manual sweep by
  the owner or another session; whoever ran it, please record it here — an
  unattributed 1,130-topic deletion is exactly the kind of quiet fleet action
  the next diagnosis will trip over.
