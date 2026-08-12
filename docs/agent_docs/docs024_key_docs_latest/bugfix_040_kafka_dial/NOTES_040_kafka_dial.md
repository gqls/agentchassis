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

### Round 3: REVISE — but by a DEAD SEAT, not an objection. Stopping here.

```
decision: revise
decided_by: unreadable reviewer(s): review_editquality.result
abstained: 7
```

**Read `decided_by` before reading the decision.** The gating seat died mid-run; it
did not object. On substance round 3 was **6 approve / 2 object**:

| seat | round 1 | round 3 |
|---|---|---|
| guardian | **hard veto** (2× HIGH) | object — 1 MEDIUM, 2 LOW ("noting for the record rather than blocking") |
| reuse_agent | object (MEDIUM) | **approve** |
| editquality | object (gating) | **unreadable — died** |
| prior_art_librarian | approve (2 LOW) | object (1 MEDIUM) |
| constitution, mission, render_guardian, debug_historian, tooling_provenance | approve | approve |

Same landmine as `bugs_closed/029`, where 3 of 5 rounds died this way. A REVISE
caused by a dead seat is harness noise; treating it as a verdict is reading noise
as signal.

**The guardian's one MEDIUM was right, and it is the third submission-fidelity
error in three rounds.** I claimed in `risks` and in a `grounded_in` check that
"both Writer sites move onto ProducerTransport wholesale, nothing left behind" —
while having moved `remote-job-spawner` into the NOT-LISTED-AS-EDITS section to fit
the 8-edit cap. So *as the plan reads*, only producer.go moves and the pool splits;
my safety argument contradicted my own edit list. **The code is fine** — verified at
HEAD, both `kafka.Writer{}` sites carry `ProducerTransport()`, and there are only
two. But the reviewers judge the plan, and the plan said two different things.

Three rounds, three variants of the same mistake: round 1 dropped `consumer.go`,
round 2 dropped `kafka_reachability.go`, round 3 kept a claim whose supporting edit
had been moved out. **The cap was never the problem — describing the code rather
than the submission was.**

`prior_art_librarian`'s MEDIUM was also fair: the "complete enumeration of Kafka
dial sites" claim shipped without a runnable query, while the two sibling checks in
the same submission each had one. Textbook asserted-absence. All four enumerations
are now in RUNBOOK §8c, none of them piped to `head`.

**Decision: stop at three rounds.** Neither remaining objection needs a code
change — one is a plan/code contradiction where the code is correct, the other is
"show your grep". Both are now recorded in the docs where the next thread will
actually read them. CLAUDE.md's guidance is one run per coherent task; three is
already well past that, and the gate has more than earned its keep on this case:

- it vetoed a fleet-wide timeout change I had argued for from an opinion,
- it forced the reuse question that turned up a hardcoded-200 `/health` waiting to
  be reintroduced,
- and its sibling-enumeration objection led to **two more phantom-broker sites**,
  one of which left core-manager with no route to Kafka at all.

**No `Council-Reviewed:` trailer on any commit, and none may be added.** It is
earned by APPROVED only; on a REVISE it is a permanent false claim of review. The
098 report will list this work as un-reviewed for ever — a known, accepted false
negative, with the trail recorded here instead.

### 2026-07-26 evening — LIVE. Three inert layers, not one.

The chassis rolled and the code is live (pod-grep with both controls, §8e of the
bug file). **The tag did not change** — still `v1.0.1167` — so the tag was worthless
as evidence and only the pod-grep settled it. Worth remembering: a same-tag rebuild
is explicitly a documented trap in CLAUDE.md, and this is what it looks like in
practice.

Then the interesting part. `/metrics` was open and Prometheus still collected
**nothing**. Two more layers underneath:

**Layer 2 — no discovery.** No PodMonitor existed; the operator ignores
`prometheus.io/*` annotations. Applied `podmonitor.yaml`. Targets appeared
immediately — 6 of them, all **DOWN**, `context deadline exceeded`.

**Layer 3 — the NetworkPolicy, and this one is the sharpest.** `allow-monitoring`
already named **port 9090**. Somebody wrote that rule intending exactly this scrape.
It has never matched a single packet, for two independent reasons:

- `from: [podSelector: {app: prometheus}]` with **no `namespaceSelector`** means
  "pods in *this* namespace" — Prometheus is in `monitoring`.
- `app: prometheus` **is not a label the pod has**; kube-prometheus-stack sets
  `app.kubernetes.io/name: prometheus`.

Either alone selects nothing. The discriminating test — and the reason I didn't
guess:

```
from ai-persona-system : pod:8080/health OK   pod:9090/metrics OK
from monitoring        : pod:8080        OK   pod:9090         TIMES OUT
```

Same-namespace fine, cross-namespace blocked, other port fine. That triangulates to
policy, not to pod networking and not to the application.

> **The pattern of this entire case, stated once: three separate layers, each of
> which alone produces a metric reading zero, and all three had legible intent.**
> The counters were written. The port was annotated. The firewall named the port.
> Every one of them was inert, and none of them could be distinguished from "the
> thing being measured is fine". That is why the bug survived six days of being
> unworked and why I nearly shipped a fix that changed nothing.

**Result:** targets 0 UP/6 DOWN → **6 UP/0 DOWN**; Prometheus **0 → 16
`ai_persona_*` series**. Not only mine — `agent_tasks_processed`,
`messages_processed`, `workflows_started`, `agent_health`. Years of counters,
incremented and thrown away, now landing.

**First fleet-wide dial baseline: 240 dials, 100% `ok`, p99 27.9ms**, spread
20/21/24 across the brokers with 175 to bootstrap.

Two things follow, and they pull in opposite directions:

1. **It does not close the case.** 20 minutes is not 7 days, and the handoff's own
   table has the *static* Deployment at **0** in a window where spawned pods carried
   10–52. A clean short sample from a mostly-idle fleet is exactly what this bug
   looks like when it is not firing. Reading this as "fixed" would be the same
   absence-as-evidence error the whole case is about.
2. **It settles the deferred timeout question with data.** p99 27.9ms against a 10s
   budget is ~360×. So the stall is not the distribution creeping toward the limit —
   it is a distinct, rare event. Part 2 can now pick a timeout from the histogram
   rather than from a remark in a bug file, which is exactly what the council's veto
   was protecting. The veto cost two rounds and bought a decision made on evidence.

---

## 2026-08-12 — session 2 (bug-backlog pickup)

### Ownership check first

`who-owns.py 040` returned OWNED, pointing at `bugfix_040_partial_build` —
**that is the OTHER 040** (the closed one), a false positive from the number
collision this file's own header warns about. The real workstream directory
(`bugfix_040_kafka_dial/`) had not been touched since 2026-07-26, 17 days —
past its own §9 "awaiting 7 days of metric" checkpoint and nobody had come back
to read it. `site_work_items` queue check clear (no open kafka/dial/refused work).

### The close condition does not close, and there is a second, bigger residual

Re-ran §9's condition live. `sum(max_over_time(ai_persona_kafka_dial_total{outcome="timeout"}[7d]))`
= **32** (not zero) — so condition 1 fails outright; the residual needs
diagnosing per condition 2, same as §10 already found once.

> **MISSTEP, caught within the session: `count by (outcome)(...)` is not
> `sum by (outcome)(...)`.** My first read of the full-window breakdown used
> `count by (outcome)`, which counts *distinct series* (i.e. how many pods hit
> each outcome), not the summed counter value. It happened to look plausible
> (`refused: 212`) and was wrong by 340x (`refused: 71,832`). Re-ran with `sum`
> before writing anything down. This is the exact `count`-vs-`sum` shape already
> named in the memory index (`a-count-you-kept-is-not-a-census.md` family) —
> should have reached for `sum` first, not discovered the gap by luck.

Corrected 17-day totals by outcome: `ok`=1,129,512 · `timeout`=48 ·
`dns_timeout`=300 · `dns`=3 · `error`=1 · **`refused`=71,832**. `refused` is new
since §10 (empty in the 07-27 snapshot) and is two orders of magnitude bigger
than every other non-ok outcome combined. This is a different, bigger finding
than the one-timeout-in-22h picture the bug file currently shows.

**Bisected the timing** (`sum(max_over_time(...[Nh]))` at increasing N, see
RUNBOOK for the loop): zero `refused` in the last 20h; all 71,832 accumulated
between roughly 2026-08-10 00:47Z and 2026-08-11 16:47Z, in bursts (quiet
06:47–18:47 on 08-10, otherwise fairly continuous), then nothing for the 20h
since. Top-offending pods (`agent-content-feed-orchestrator`,
`agent-feed-ingester`, `agent-tool-acceptance-agent`, `agent-feed-triage`,
`agent-section-editor`, `agent-build-dispatch-loop`) each carried 850–1,300
refused dials in that window alone.

**Ruled out broker restarts as the cause first, before theorising code:** all
three broker pods are 45 days old with 0 recent restarts (`prod-0`'s restart
count of 6 is ancient, predates this metric entirely). So this is not "the
broker bounced during a roll" — the brokers were up throughout.

**Almost all of the 71,832 carry an EMPTY `broker` label**
(`sum by (broker)(...{outcome="refused"}...)` → `{}` 71,826 vs bootstrap-only 6).
Read `brokerLabel()` (dialer.go): it returns the input address unchanged when
`net.SplitHostPort` errors, but `net.SplitHostPort(":9092")` does **not**
error — it returns host="" successfully. So an empty broker label points at a
dial address of literally `:9092`, not a missing/malformed one.

**Found a live precedent for that exact string already in the code**:
`spawn_actions.go:1032` filters `broker != ":9092"` when validating a configured
broker list — i.e. someone already learned this string shows up and guarded one
call site against it. `topic_manager.go:299`'s `getController()` builds its
dial target as `fmt.Sprintf("%s:%d", controller.Host, controller.Port)` straight
from kafka-go's `Controller()` metadata response with no validation — if
`controller.Host` comes back empty, this produces exactly `:9092`, which as a
client dial resolves to the pod's own loopback, where nothing listens →
instant `ECONNREFUSED`, no timeout delay, matching both the outcome label and
the volume (fast failures, no 10s wait, so a retrying pod can rack up hundreds
in its lifetime).

**Not yet established, and deliberately not asserted as fact:** whether
`getController` is what actually runs inside a spawned *agent* pod (topic
creation is normally the orchestrator's job, before spawn) versus kafka-go's
internal consumer-group-coordinator lookup (used by `consumer.go`'s `Reader`,
same `InstrumentedDialer`) hitting an analogous empty-Host response from the
broker. Both are plausible from reading the code; neither is confirmed from a
log line or a debugger. This is exactly the "cross-cutting, cause not
obviously where the symptom is" shape CLAUDE.md's diagnosis-before-debugging
section asks to file rather than guess at, so filed it instead of picking one
and writing a patch:

`090_TRIGGER_needs_diagnosis_v1.sh`, intake correlation
`04195fa7-28c2-410a-a8cb-15d42acf43c4`, claimed by the live dispatch loop,
run correlation `39bb6fe8-a55c-476e-8ffd-026bec4b57ca`. `site_work_items`
queue check was clear before filing (no duplicate).

### Council verdict and one more cross-check

Council: **APPROVED**, round 1, correlation `af5f74bc-5e6c-4a6c-a3fc-7ac27eab4b6f`
— 2 advisory objections (editquality medium, guardian low+medium), none
high-severity. Full review JSON archived below for reference; key point:
`editquality` flagged `LANDMINES.md`'s blank-`MetadataTopics` entry
(`bugs_open/240`) as an uncovered candidate mechanism. Checked the timing:
burst episode 1 overlaps 240's OOM window (partial fit, ~20,255 events);
episode 2 (72% of the burst) starts 7h after 240 resolved (does not fit).
Written up in bug 040 §11.5 and contributed to `bugs_open/240` directly rather
than duplicated. `guardian`'s two objections (name the 4th `getController`
caller explicitly; confirm the fail-fast-on-total-failure behaviour change is
intended) are noted here rather than actioned — both are advisory-only,
correctly assessed as low/medium, not blocking, and the fail-fast behaviour
*is* intended (stated in the submission's risk #4).

Full verdict JSON: `kubectl -n ai-persona-system exec postgres-clients-0 --
psql -U clients_user -d clients_db -c "SELECT body FROM diagnosis_artifacts
WHERE correlation_id='af5f74bc-5e6c-4a6c-a3fc-7ac27eab4b6f' AND
kind='council_report';"`

## 2026-08-12, later — fresh chassis build verified; the follow-up diagnosis run was killed by that same roll

**Fix proven live**, `v1.0.1291` (both `agent-chassis` pods, ~85min old at check
time): OCI revision label `da5a7eb8f` (`docker image inspect ... .Config.Labels`,
image cached locally so no log-tail or exact-match binary grep needed) —
`git merge-base --is-ancestor e1f960ac2 da5a7eb8f` → yes, and zero intervening
commits touched `topic_manager.go`. Cross-checked against the running binary
with controls: `controller metadata invalid, trying next broker` and
`controller metadata returned an empty host` both present (≥1), the nonsense
negative control absent (0), a known-good pre-existing string present (≥1).

**The follow-up `090` run (correlation `e91d71d6-058a-4902-a852-c6d54bc7411c`,
"does the same unvalidated-Host pattern exist elsewhere in platform/kafka")
did not complete — the chassis roll that shipped the build above killed it
mid-flight.** It was cycling normally (route → load_runtime → assemble_bundle →
verdict, multiple rounds) until ~14:38Z, then sat at `load_runtime` with zero
`last_activity` movement for 100+ minutes. `site_work_items` for this intake
is now `status='failed'` — the documented 40-minute claim-timeout landing on a
terminal state (`max_attempts=1`, per `090_TRIGGER`'s header note 5) rather
than resetting to retriable, which is consistent with the owning pod having
been replaced out from under it by the roll, not a content/logic failure.
**No verdict was produced; this is not evidence for or against the audit
question.** Re-fired as a fresh intake rather than resumed (the old one is
terminal and cannot be reclaimed).

**Re-fired** (`FORCE=1`, since the coverage probe's `status NOT IN
('complete','cancelled','rejected')` clause does not treat `failed` as
terminal, so it correctly flagged the dead run and had to be overridden with
the reason recorded above). New intake `needs_diagnosis:followup-controller-empty-host-audit-v2`,
run correlation `58a0390c-33ec-4580-9697-3320b280475d`. Confirmed no new
commits landed on `platform/kafka/` or `spawn_actions.go` since the push, so
the run's view of those files is current. **Left running in the background —
not polled to completion this session; check status with:**
```sql
SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s
FROM orchestration_states WHERE correlation_id = '58a0390c-33ec-4580-9697-3320b280475d'::uuid
ORDER BY created_at DESC LIMIT 1;
```
Verdict once done: `SELECT collected_data->'verdict' FROM orchestration_states
WHERE correlation_id = '58a0390c-33ec-4580-9697-3320b280475d'::uuid ORDER BY
created_at DESC LIMIT 1;` (pipe through `psql -t -A` to a file, not straight to
a pipeline — the payload is large enough that `kubectl exec -i` piped into
`python3 -c` has truncated mid-stream at least once this session).

## 2026-08-12, later still — re-fired run COMPLETED but UNVERIFIABLE; closed the loop by hand instead of firing a third run

`58a0390c-33ec-4580-9697-3320b280475d` finished (`status='COMPLETED'`,
`current_step='complete'`) — no third roll killed this one. Outcome:
**`UNVERIFIABLE`**, not REFUTED and not a hit. Per
`an-unverifiable-verdict-does-not-say-your-premise-was-false` (memory) this
means "the search didn't finish", not "the premise was wrong" — and this
verdict was unusually explicit about *why*: its `needed_evidence` field named
the exact gap verbatim — the `fmt.Sprintf("%s:%d", ...)` content search was
capped at 40 rows ordered by path and cut off before
`platform/orchestration/actions/spawn_actions.go`, so absence there was
UNKNOWN rather than confirmed, and `spawn_actions.go`'s own body (109,756
chars) hadn't fit in-context to be read directly either.

Rather than re-fire a third `090` run for two greps and two file reads, did
it first-hand:
- `grep -rn 'fmt.Sprintf("%s:%d"' --include=*.go platform/ internal/ cmd/` →
  **one hit fleet-wide**, `topic_manager.go:322`, inside `controllerAddress`
  — the already-fixed function. No sibling.
- `grep -rn 'kafka.Broker{' ...` → **zero hits**.
- Every `conn.Controller()` / `conn.Brokers()` call site read directly:
  `Controller()` used once (the fixed `getController`); `Brokers()` used once
  (`kafka_reachability.go:115`, `dialAny`) and its result is **discarded**
  (`_, metaErr := conn.Brokers()`) — never turned into a dial string.
- Read `CreateJobTopic` and `getConfiguredKafkaBrokers` in full: both
  `:9092`-literal filters operate on **statically-configured** strings (a
  function parameter and `os.Getenv`, respectively) — never on live kafka-go
  metadata. Different mechanism, different bug class; nothing to reuse or
  duplicate.

**Both council objections now have a first-hand, uncapped answer, written up
in `bugs_open/040` §11.7:** the unvalidated-Host-from-live-metadata pattern
exists nowhere else in the fleet (`getController`/`controllerAddress` is the
only site, already fixed), and the `:9092` filters `prior_art_librarian`
pointed at are an unrelated static-config guard, not a sibling instance.
`guardian`'s caller enumeration is confirmed exactly as the capped run had
already found: `CreateTopic`, `DeleteTopic`, `ListTopics`, `TopicExists`, no
fifth. This closes the follow-up-diagnosis thread from
`HANDOFF_2026-08-12b_040_build_verified_diagnosis_rerunning.md`. Bug 040
itself stays OPEN (§11.6/§11.7 — this only settles isolation, not root cause
of the `refused` burst).

## 2026-08-12, later again — another chassis roll (v1.0.1293); first post-fix `refused` read is silent but not yet meaningful

Owner rolled again: `v1.0.1293`, git commit `7a1887e31`, both pods restarted
~19:13Z. Fix (`e1f960ac2`) still an ancestor, confirmed. Waited out the ~300s
no-dispatch window before doing anything else — didn't fire any orchestration
this time, just PromQL, which isn't subject to that restriction.

Bisected `sum(max_over_time(ai_persona_kafka_dial_total{outcome="refused"}[Xh]))`
same way as §11.3: 0 at X=26, jumps to 19,616 at X=28, flat at 71,832 by X=72.
Cross-checked the zero isn't blind — `outcome="ok"`/`"timeout"` both nonzero
over the same window. So: genuinely zero `refused` events in the last ~26.5h.

**Went looking for exactly when the fix started running, and it matters:**
`kubectl get rs -l app=agent-chassis` still had the old (now-scaled-to-0)
`v1.0.1291` replicaset, creation timestamp `2026-08-12T14:55:10Z` — so the
fix has been continuously live since then. That's **22h08m after** the last
`refused` sample (~08-11 16:47Z, matches the already-known episode-2 end).
Of the ~26.5h of silence, only the last 4h24m happened with the fix actually
running; the other 22h of quiet came for free, on the unfixed binary, before
it ever shipped. **So this can't be read as the fix working yet** — the
metric was already capable of a day of silence with the bug still live
(that's exactly what the gap between episode 1 and episode 2 already showed,
on a shorter scale). Wrote the honest version of this into `bugs_open/040`
§11.8, including the explicit "what would actually be informative" note
(re-read once the fix has been live for several multiples of the ~12h
inter-episode gap already observed, not just once) so a future session
doesn't mistake today's quiet for confirmation.

Noted but flagged as weak, per `two-clean-runs-cannot-establish-stability`
(memory): the current 26.5h gap already exceeds the one inter-episode gap
observed so far (~12h). Two episodes is not a rate. Not relying on it.
