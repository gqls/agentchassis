# NOTES — 040-kafka-dial

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-26 — session 1

### Ownership check first

`scripts/who-owns.py 040` returned **AMBIGUOUS** — two unrelated cases share the
number. The active workstream (`bugfix_003_spawn_loss`, 16 commits/14d) owns the
*other* one (`040-partial-build`, closed 07-25). Nobody was working
`040-kafka-dial`. Queue check for in-flight work items matching kafka/dial/timeout
returned only unrelated page-rerender rows. Clear to proceed.

### The measurements, in the order I made them

Started by re-running the handoff's own §2 grep across running fleet pods. Got
**one** dial timeout across six pods in a 12h window — against the handoff's
10–52 per pod per 12h. That looked like the bug had gone.

> **MISSTEP 1, caught within the hour: the 12h window was a lie.** Every pod in
> `ai-persona-system` had restarted ~100 minutes earlier, so "12h of logs" was
> 1.7h of actual pod life. I nearly recorded "rate has dropped" as a finding.
> **The check that catches it:** always pull `.status.startTime` and normalise by
> real uptime before comparing a log count to a historical one. A window longer
> than the pod's age silently reports a lower rate. Now in the RUNBOOK §4.

Then worked the §4 hypotheses. All clean: conntrack 1,021/262,144 (28-day peak
113,891 — never close); `ListenOverflows` increase 0 on all five nodes across 8
consecutive days; `softnet_dropped_total` 0 over 28d; CPU steal <0.5%; brokers at
60–65m CPU with node CPU 1–4%. So neither conntrack pressure (§4.3) nor broker
request-handler saturation (§4.5) survives.

The one event I *did* capture in full is worth recording precisely:

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

T+60s from pod start, **10.32s** elapsed (exactly kafka-go's `DefaultDialer` 10s),
and the retry succeeded 0.67s later. One event in ~2h of uptime.

### The DNS theory, and how far it actually got

The handoff's asymmetry — spawned Job pods 10–52 errors, the static Deployment pod
0 — plus that startup-time event made me look at DNS, since spawned pods do a
burst of fresh dials while long-lived pods hold pooled connections.

Found: `ndots:5`, three-domain search path, 2 CoreDNS replicas for 5 nodes, no
NodeLocal DNSCache. And the brokers advertise a **three-dot** name
(`...kafka.svc`, no `.cluster.local`), verified from the broker's own
`/tmp/strimzi.properties`. So every post-bootstrap connection walks three
NXDOMAIN rounds before the one that works, doubled by parallel AAAA.

The volume evidence is strong: **384,392 NXDOMAIN of 525,152 responses in 24h
(73%)**, A:AAAA exactly 1:1, ≈7.5 queries per useful answer against a predicted 8
for this name shape. Very tight match.

> **MISSTEP 2 — I was about to propose lowering `ndots`, which is strictly
> worse.** It is the obvious move and it backfires: the name genuinely needs
> `cluster.local` appended, so at `ndots:2` the resolver tries it absolute
> (NXDOMAIN) **and then still** walks all three search domains — four rounds
> instead of three. **The check:** write out the actual resolution order for the
> specific name before touching `ndots`. Logged in `WRONG_CALLS.md`. The real fix
> is fully-qualifying `advertisedHost` broker-side.

> **MISSTEP 3 — and this one is the important correction to my own theory.** I
> had the DNS story looking causal and went to confirm it with latency. It does
> not hold up as a *cause*:
> - CoreDNS p99 request duration is **2.69ms**, 0 panics. The server is fast.
> - 1,200 client-side probe lookups from three pods on three different nodes:
>   **zero stalls ≥2s.**
> - 300 short-name vs 300 FQDN lookups: **both under 1s total.** With healthy DNS
>   the search path costs no meaningful latency at all.
>
> So the search-path tax is a **volume** problem (3× the packets, 3× the exposure
> to loss), not a latency problem. It is risk reduction, **not** the cure, and
> saying otherwise would have been exactly the kind of confident-and-wrong claim
> the diagnosis-loop section of CLAUDE.md exists to prevent. Recorded as
> `[MEASURED]` in the PLAN so nobody reads it as the fix.

**Net position: root cause NOT established.** What I can say is what it is *not*,
and that the reason it stays unknown is that nobody can measure it.

### The finding that reframed the case

While wiring the dial metric I checked whether it would actually be collected.

**Zero `ai_persona_*` series exist in the live Prometheus.**
`observability.NewMetricsServer` has no callers anywhere.
`cmd/agent-chassis/main.go` built its own `ServeMux` with only `/health` and
`/ready`. Meanwhile `spawn_actions.go` annotates every spawned pod
`prometheus.io/port: "9090"` / `prometheus.io/path: /metrics` and declares a
matching containerPort. **Prometheus has been scraping a closed port for the life
of the fleet**, and every counter in `platform/observability` — including the
`fetch_message` error counter that was supposed to make broker trouble visible —
has been dead since it was written.

This nearly ate the whole change: I would have shipped a metric that goes
nowhere, then read its absence as good news. My own plan's "positive control
before trusting a zero" step is what forced the check. That step earned its place.

### Build and test in a shared tree

`go build ./platform/...` failed on `platform/orchestration/actions/` —
`undefined: maxItems`, plus a `NavVisibility` signature mismatch. Another
session's in-flight work, mid-edit, not compiling. Did **not** touch it. Overlaid
my files onto `git archive HEAD` in a scratch tree and built there.

That paid for itself immediately: it caught `a.agentType` → the field is
`a.AgentType`. A real compile error of mine that the broken shared tree was
masking.

Two smaller traps:
- `go build ./...` in the clean tree reports a package-name conflict from two
  stray `main` packages under `docs/.../traffic_probe/deploy_setup/working_dir`.
  Use `./platform/... ./cmd/... ./internal/...`.
- `prometheus/client_golang/prometheus/testutil` cannot be imported without
  editing the shared `go.mod`. Read the counter back through
  `prometheus.DefaultGatherer.Gather()` instead — better anyway, since that is the
  registry `promhttp.Handler()` serves, so the test also proves the series would
  appear on `/metrics`.
- busybox `date +%s%N` silently returns 0, so the first DNS timing loop reported
  `max=0ms` for all 150 samples. Second-granularity timing is enough for a 5s
  resolv.conf stall.

### Committed

- `95df64d63` — the change (11 files).
- `50cd1313a` — gofmt on two files the pre-commit hook flagged. The hook was
  right; I had not run gofmt before committing.

Council gate submitted: correlation `7abe1a57-e3db-4b71-9e3a-744fbf8c24b1`.
Two `grounded_in` quotes failed byte-exactness on the first pass and were fixed
before firing — a markdown line that is wrapped in the original, and a path I had
prefixed with the word "vendored" so it no longer resolved. Verifier in the
RUNBOOK §9.

### Left deliberately undone

- Broker `resources`/`jvmOptions` and the FQDN `advertisedHost`: written into the
  repo templates, **not applied**. Both roll the cluster; owner's call.
- The case stays in `bugs_open/`. Instruction was to close it; flagged the
  conflict with the repo's fixed-AND-live bar and the handoff's own §5, and the
  owner confirmed keep-open.

### Later the same day — the dead-metrics finding proven from the pod, and a roll that missed us

Chassis **v1.0.1167** rolled while I was writing docs (another session's 044+060
close). Checked whether it carried my change: pod started 14:57:15Z, my commit
landed 15:07:50Z, so **no**. Confirmed by pod-grep rather than by trusting the
timestamps, with both controls so the method itself is not what is under test:

```
strings /app/agent-chassis | grep -c ai_persona_kafka_dial_total                -> 0   (mine: NOT live)
strings /app/agent-chassis | grep -c ai_persona_kafka_messages_produced_total   -> 1   (positive control)
strings /app/agent-chassis | grep -c ai_persona_kafka_dial_nonexistent_xyz      -> 0   (negative control)
wget -qO- http://localhost:9090/metrics -> can't connect to remote host: Connection refused
```

The last line is the best evidence in the whole case. The metric strings **are** in
the shipped binary — they always were — and the annotated port is **closed**. The
counters exist, are maintained, and are unreachable. Stronger than the Prometheus
query alone, because it localises the fault to the process rather than to the
scrape config.

**Also: my 016b §9 append got swept into another session's commit** (`a9621ceb2`,
their 044+060 close) before I committed it. Content intact — this is the hazard
CLAUDE.md documents, and the answer is forward-only: the entry is in the file and
in history, just under someone else's message. Noted in my own commit rather than
attempting any history surgery. The lesson is the one already written down and
which I still got wrong: **commit each piece the moment it is coherent**, not after
writing five documents.

### The council REJECTED round 1, and it was right

Hard veto from `guardian`, 2× HIGH. Verdict `rejected`, `abstained: 8`,
`unreadable: none` — so the decision was made on a full reading, not a dead seat.

**The two HIGH objections, both correct:**

1. The shared dialer lowered the default dial timeout **10s → 5s for every
   process in the fleet**. That is a behaviour change to shared messaging
   plumbing, not instrumentation, "bundled into a metrics PR".
2. Routing the producer through `SharedTransport` changed `IdleTimeout` 30s → 5m
   and `MetadataTTL` 6s → 30s **for every producer simultaneously** — "a real
   change to failover reactivity across all sites/pipelines".

I had justified both. Badly. The timeout from a *remark in the bug file* (§4.6),
the TTLs from *"the Java client's default is 300s"*. Neither is measurement, and
my own risks section had already admitted the timeout cut "makes it fail sooner
rather than hang" if the residual flake is multi-second-but-recoverable. I wrote
the counter-argument to my own change and shipped it anyway.

The guardian's alternative was better than arguing: **split it.** Land the
non-behaviour-changing half now — counters wrapped around the *existing*
per-call-site dialers, changing no default — and send the consolidation to
architecture review. Done. `SharedDialer`/`SharedTransport` are gone;
`InstrumentedDialer(timeout)` takes the caller's own budget and the package has
**no default at all**, so it cannot change anyone's behaviour. `ProducerTransport`
reproduces kafka-go's `DefaultTransport` byte-for-byte (3s/5s/30s/6s) and stays
package-level because a nil `Transport` *already* shared one pool per process —
a per-Writer Transport would have split it, which is itself a behaviour change I
nearly made while trying to make none.

This is strictly better than what I had, and not only for safety: choosing 5s
today would have been guesswork. Once the histogram exists the timeout can be
picked from the actual latency distribution. **The veto improved the change.**

**The reuse objection (editquality + reuse_agent, MEDIUM) — right in process,
wrong in conclusion, and the right answer was a third thing.** They said: you
cite "`observability.NewMetricsServer` has zero callers" as the defect, then
build a second metrics server beside it instead of calling it. Fair — I never
said why. On inspection it disqualified itself:

```go
func (m *MetricsServer) Start() error {
	http.Handle("/metrics", promhttp.Handler())   // GLOBAL DefaultServeMux — panics on duplicate
	http.HandleFunc("/health", ...)               // HARDCODED 200
	return http.ListenAndServe(":"+m.port, nil)
}
```

That `/health` is *precisely* the `bugs_open/003 §4.1` defect —
`kafka_reachability.go`'s own doc comment says pods "pointing at hardcoded-200
endpoints... reported healthy forever and w[ere] never rescheduled". So calling
it as written would have re-introduced a bug we deliberately fixed. But building
a rival was also wrong. **Fixed it and called it**: own mux, no fake `/health`,
one path, dead code revived rather than duplicated.

**`editquality` flagged `consumer.go` as missing.** The code had it all along —
I hit the 8-edit cap and dropped it from the *plan* while still quoting its
pre-change state in `grounded_in`. Reviewers only see the sketch, so from where
they sat the objection was correct. My error, and the same class as the trimmed
quotes: **the submission is the artefact under review, not the commit.**

**`prior_art_librarian` (MEDIUM)** asked that "nothing serves /metrics" be
confirmed from a running pod rather than taken on my word. Reasonable, and I'd
already done it by then — recorded above with both controls.

Round 2 resubmitted under the same correlation with `RESUBMIT_CORR`.

### One more thing found after round 1, and it would have sunk the whole fix

Opening the port is **necessary but not sufficient**. This cluster runs the
prometheus-operator, which discovers targets **only** via label-selected CRs:

```
spec.podMonitorSelector      = {matchLabels: {release: kube-prometheus-stack}}
spec.additionalScrapeConfigs = None
podmonitors: 0    scrapeconfigs: 0
```

So the `prometheus.io/scrape` annotations on every spawned pod are the **plain-
Prometheus convention and this is not plain Prometheus** — they are inert and
have never caused a single scrape. Had only the listener shipped, `/metrics`
would have been open and *still* unscraped, and the zero would have looked
exactly like a fixed bug. **The same trap, one level up, twice in one day.**

`deployments/kustomize/services/agent-chassis/base/podmonitor.yaml` is the
missing half (repo only, not applied). Selector verified against all three live
pod shapes; keyed on `app` not `spawned-by` because the two spawn paths stamp
different `spawned-by` values but the same `app`.

### Round 2: REVISE — and the gate found a bug I had missed

`decided_by: gating objection from editquality`. **8 of 11 seats approved**, and
the guardian moved **veto → object**. So the round-1 rewrite landed.

**The gate found a real second instance, and that is the most useful thing it has
done on this case.** Chasing `editquality`'s "the broker-list change is functional,
not instrumentation" objection, I looked for siblings and found the same phantom
entry `kafka-0.kafka-headless.kafka:9092` in
`spawn_actions.go:1019` — the fallback list **every spawned agent inherits**. I
had fixed the site I tripped over and never enumerated the others.
`WRONG_CALLS.md` already carries a tally row for exactly this ("enumerate the
SIBLING instances before quantifying") and I still did it.

Its neighbour went too: `personae-kafka-cluster-kafka-bootstrap.kafka:9092` is the
same host as the FQDN entry directly above it, merely unqualified — so it walks
the whole ndots:5 search path to arrive at a name already in the list.

Claim precision also corrected. I had written "no such Service exists". A
`kafka-headless` manifest **does** exist in-repo
(`deployments/kustomize/infrastructure/kafka/kafka.yaml`, a hand-rolled Kafka
StatefulSet) — it is just referenced by no kustomization and never applied, and no
such Service exists in the live cluster. Both checks now recorded rather than
asserted:

```
grep -rn 'infrastructure/kafka' --include=kustomization.yaml deployments/   -> empty
kubectl get svc -A | grep -i headless                                        -> empty
```

**I introduced a behaviour change while claiming I had made none.** Guardian
MEDIUM, and correct. Threading the caller's `ctx` into `topic_manager`'s dials
looked like a free improvement. It is not: six of those eight sites previously
called bare `kafka.Dial`, which uses `context.Background()`, so with a real ctx
any caller holding a deadline shorter than 10s silently gets a shorter dial.
That is precisely the class round 1 was vetoed for, reintroduced at smaller scale
in the round that was supposed to remove it.

> **MISSTEP 4, and the reverting of it nearly became MISSTEP 5.** My first fix was
> a blanket replace of `ctx` → `context.Background()` across all eight sites. But
> `WaitForTopic` and `WaitForTopicOld` **always** used `ctx` — blanket-reverting
> them would have been the identical error mirrored. **The check:** diff each site
> against the pre-change blob individually rather than trusting a global replace.
> Final state verified site-by-site: 6 background, 2 ctx, all eight targets
> identical to `95df64d63^`.

Checks the seats asked for, all answered by running them rather than arguing:

- `NewMetricsServer` callers: **zero** besides the one this change adds. So
  rewriting `Start()` cannot break an existing contract.
- `kafka.Writer{}` sites: **exactly two** (producer.go, remote-job-spawner) and
  **both** now use `ProducerTransport` — the pool moves wholesale rather than
  splitting, which was the guardian's LOW.
- Complete dial-site enumeration: consumer, producer, topic_manager ×8,
  kafka_reachability, remote-job-spawner reader+writer. That is all of them.

**And the gating objection was, again, an edit I dropped for the 8-edit cap** —
the health probe this time, `consumer.go` last time. Twice now, same cause. Round
3 lists the health probe explicitly and adds a **"NOT LISTED AS EDITS (8-edit cap)"**
section naming every remaining file, because the real lesson is not "remember the
cap" but **"if the plan does not fit in 8 edits, say which ones are missing"**.
An edit left out reads as an edit not made, and that has now cost two rounds.

Round 3 also drops the blanket "zero behaviour change" framing. Four small,
deliberate, separately-justified behaviour changes do ride along, and they are now
enumerated in the summary instead of being discoverable only inside individual
edit rationales.
