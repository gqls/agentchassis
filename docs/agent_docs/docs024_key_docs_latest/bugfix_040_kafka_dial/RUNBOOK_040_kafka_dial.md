# RUNBOOK — 040-kafka-dial

Every command here was needed to get something right. Gotchas attached.

Resolve this case **by slug** (`040-kafka-dial`): the number 040 is shared with
`bugs_closed/040_..._failed_page_build_leaves_page_deployed_and_partially_composed.md`.

---

## 1. The baseline metric (replaces the log grep)

The handoff's §4.1 said "baseline the rate" by grepping pod logs. **That method
cannot work** — the pods that flake are ephemeral spawned Jobs, and their logs are
gone before anyone looks. Once an image carrying `ai_persona_kafka_dial_total`
rolls, use this instead:

```bash
# Dial outcomes fleet-wide over 24h. This is THE metric for the case.
sum by (outcome) (increase(ai_persona_kafka_dial_total[24h]))

# Per-broker, timeouts only — is it one broker or all three?
sum by (broker) (increase(ai_persona_kafka_dial_total{outcome="timeout"}[24h]))

# Is the stall in resolution or in the connect? This is the open question.
sum by (outcome) (increase(ai_persona_kafka_dial_total{outcome=~"dns.*|timeout"}[24h]))

# Dial latency distribution — a healthy cluster should be entirely sub-100ms.
histogram_quantile(0.99, sum by (le) (rate(ai_persona_kafka_dial_duration_seconds_bucket[1h])))
```

**Gotcha: a zero here is meaningless until you have proven the metric is
scraped.** That is not paranoia — see §2.

### Querying Prometheus without a port-forward

There is no ingress. Exec into the Prometheus pod:

```bash
kubectl -n monitoring exec prometheus-kube-prometheus-stack-prometheus-0 -c prometheus \
  -- wget -qO- 'http://localhost:9090/api/v1/query?query=<URL-ENCODED-PROMQL>'
```

**Gotcha: `-c prometheus` is required** (the pod has two containers), and the
PromQL must be URL-encoded — `[24h]` and `{}` break the query otherwise. Helper
used throughout this investigation:

```bash
promq() {
  local q; q=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1]))" "$1")
  kubectl -n monitoring exec prometheus-kube-prometheus-stack-prometheus-0 -c prometheus \
    -- wget -qO- "http://localhost:9090/api/v1/query?query=$q" \
  | python3 -c "
import json,sys
for r in json.load(sys.stdin)['data']['result']:
    print({k:v for k,v in r['metric'].items() if k!='__name__'}, r['value'][1])"
}
```

## 2. Prove the metric exists before trusting any number from it

**This is the most important check in the file.** On 2026-07-26 the live
Prometheus held **zero** `ai_persona_*` series — nothing had ever served
`/metrics`, so every counter in `platform/observability` was dead. A metric that
reads zero because it is unscraped is indistinguishable from a fixed bug.

```bash
# Does ANY application metric exist? Expect a non-zero count post-roll.
kubectl -n monitoring exec prometheus-kube-prometheus-stack-prometheus-0 -c prometheus \
  -- wget -qO- 'http://localhost:9090/api/v1/label/__name__/values' \
| python3 -c "
import json,sys
n=json.load(sys.stdin)['data']
h=[x for x in n if x.startswith('ai_persona')]
print('ai_persona_* series:', len(h)); print(h[:20])"

# Is the port actually open in a pod? (was closed for the fleet's whole life)
kubectl -n ai-persona-system exec <pod> -- wget -qO- --timeout=3 http://localhost:9090/metrics | head
```

## 3. Post-roll deployment verification

Verify against the **running pod**, never git and never the tag. Grep a string
the change **creates**, plus a control string that must be absent:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)

# Expect >= 1 — this string exists only because of this change.
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "ai_persona_kafka_dial_total"'

# Positive control: expect 0. If this is non-zero, `strings | grep -c` is
# matching something other than what you think.
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "ai_persona_kafka_dial_nonexistent"'
```

## 4. The old log grep (keep one run for cross-check)

Run this **alongside** the metric once, to confirm the two agree before retiring
the grep:

```bash
kubectl -n ai-persona-system logs --since=12h --tail=3000 <pod> \
  | grep 'i/o timeout' | grep -o 'dial tcp [0-9.]*:9092' | sort | uniq -c
```

**Gotcha: normalise by pod uptime.** On 2026-07-26 a fleet-wide sweep with a 12h
window looked almost clean, but every pod had restarted ~100 minutes earlier — so
the window was 1.7h of real life, not 12h. Get the true denominator:

```bash
kubectl -n ai-persona-system get pod <pod> -o jsonpath='{.status.startTime}'
```

## 5. Reading the broker's real configuration

The Strimzi CR in the repo is a Terraform template; the rendered result is what
matters, and the checked-in state does not reflect it.

```bash
# What the brokers actually advertise (the name every client dials post-bootstrap).
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 \
  -- grep -i advertised /tmp/strimzi.properties

# Resource floor and heap. Returned "{}" and "-Xms128M" with no -Xmx on 2026-07-26.
kubectl -n kafka get pod personae-kafka-cluster-combined-pool-prod-0 \
  -o jsonpath='{.spec.containers[0].resources}'
kubectl -n kafka get pod personae-kafka-cluster-combined-pool-prod-0 \
  -o jsonpath='{range .spec.containers[0].env[*]}{.name}={.value}{"\n"}{end}' | grep -i heap
```

**Gotcha: brokers live in the `kafka` namespace, not `ai-persona-system`.** A
`kubectl get pods -n ai-persona-system | grep kafka` finds only `kafka-scheduler`
and looks like the brokers are missing.

## 6. Refutation checks (so nobody re-walks them)

All of these came back clean on 2026-07-26. Re-run before suspecting them again.

```bash
promq 'node_nf_conntrack_entries'                                  # 1021 …
promq 'node_nf_conntrack_entries_limit'                            # … of 262144
promq 'max_over_time(node_nf_conntrack_entries[28d])'              # peak 113891
promq 'increase(node_netstat_TcpExt_ListenOverflows[24h])'         # 0, all nodes, 8 days
promq 'increase(node_softnet_dropped_total[28d])'                  # 0
promq 'max_over_time(avg by (instance)(rate(node_cpu_seconds_total{mode="steal"}[5m]))[24h:5m])'  # <0.5%
promq 'histogram_quantile(0.99, sum by (le) (rate(coredns_dns_request_duration_seconds_bucket[24h])))'  # 0.00269
promq 'sum by (rcode) (increase(coredns_dns_responses_total[24h]))' # 384392 NXDOMAIN / 140760 NOERROR
kubectl top pods -n kafka                                          # brokers 60-65m CPU
```

### The client-side DNS stall probe

`nslookup` in these images is busybox: **`date +%s%N` silently yields 0**, so
nanosecond timing does not work. Time at second granularity, which is enough to
catch the 5s resolv.conf timeout that a lost UDP packet costs:

```bash
kubectl -n ai-persona-system exec <pod> -- sh -c '
N="personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc"
slow=0; worst=0; i=0
while [ $i -lt 400 ]; do
  s=$(date +%s); nslookup "$N" >/dev/null 2>&1; e=$(date +%s); d=$((e-s))
  [ $d -ge 1 ] && slow=$((slow+1)); [ $d -gt $worst ] && worst=$d
  i=$((i+1))
done
echo "slow_ge_1s=$slow worst_s=$worst"'
```

Run it from pods on **different nodes** in parallel. 1,200 lookups across three
nodes produced zero stalls ≥2s.

## 7. Building and testing in a shared tree

`platform/orchestration/actions/` was mid-edit by another session and did not
compile (`undefined: maxItems`), which makes `go build ./...` fail for reasons
that are not yours. **Do not fix their code.** Overlay your files onto a clean
`HEAD`:

```bash
W=/tmp/cleantree; rm -rf $W; mkdir -p $W
git archive HEAD | tar -x -C $W
for f in platform/kafka/dialer.go platform/kafka/consumer.go ... ; do cp "$f" "$W/$f"; done
cd $W && go build ./platform/... ./cmd/... ./internal/... && go test ./platform/kafka/
```

**Gotcha: use `./platform/... ./cmd/... ./internal/...`, not `./...`** — a bare
`./...` trips over two stray `main` packages in
`docs/agent_docs/.../traffic_probe/deploy_setup/working_dir` and reports a
package-name conflict instead of building.

This is also what caught a real bug of mine: the field is `a.AgentType`, not
`a.agentType`.

**Gotcha: `prometheus/client_golang/prometheus/testutil` is not resolvable**
without editing the shared `go.mod` (`go: updates to go.mod needed`). Read the
counter back through `prometheus.DefaultGatherer.Gather()` instead — which is
better anyway, because that is the registry `promhttp.Handler()` serves, so the
test also proves the series would appear on `/metrics`.

## 8. Broker-side changes — NOT APPLIED

Both are written into the repo templates and deliberately left unapplied
(owner's call, 2026-07-26). **`terraform apply` is not the safe route**: the
checked-in state for `040-kafka-cluster` has `serial: 1` and zero resources, so
it does not know about the live cluster.

Patch the live CRs directly, one at a time, and let Strimzi roll the brokers:

```bash
# (a) Resource floor + bounded heap. Matches the template.
kubectl -n kafka patch kafkanodepool combined-pool-prod --type merge -p '{
  "spec": {
    "resources": {"requests": {"memory": "2Gi", "cpu": "1"},
                  "limits":   {"memory": "4Gi", "cpu": "2"}},
    "jvmOptions": {"-Xms": "2G", "-Xmx": "2G"}
  }}'

# (b) Fully-qualified advertised listeners. Rewrites advertised.listeners.
#     Read the warnings in kafka-cluster-cr-prod.yaml.tpl first — in particular,
#     do NOT reach for lowering ndots instead; it is strictly worse.
kubectl -n kafka patch kafka personae-kafka-cluster --type json -p '[{
  "op": "add",
  "path": "/spec/kafka/listeners/0/configuration",
  "value": {"brokers": [
    {"broker": 0, "advertisedHost": "personae-kafka-cluster-combined-pool-prod-0.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local"},
    {"broker": 1, "advertisedHost": "personae-kafka-cluster-combined-pool-prod-1.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local"},
    {"broker": 2, "advertisedHost": "personae-kafka-cluster-combined-pool-prod-2.personae-kafka-cluster-kafka-brokers.kafka.svc.cluster.local"}
  ]}}]'
```

Pre-flight for either: confirm all three brokers are `Running` and in-sync
first, and watch the roll rather than firing and leaving.

```bash
kubectl -n kafka get pods -l strimzi.io/cluster=personae-kafka-cluster -w
kubectl -n kafka logs -f deployment/strimzi-cluster-operator -n strimzi | grep -i personae
```

Baseline the dial metric **before** and after, or the change cannot be judged.

## 8c. Reproducible enumerations (asserted absence needs a command)

The council's prior-art seat objected — correctly — that the "complete enumeration
of Kafka dial sites" claim shipped without a runnable query, unlike the other two
checks in the same submission. Absence claims need the command attached. These are
the three, and **none of them is piped to `head`** (see §7's trap: a capped
enumeration is indistinguishable from a complete one, and that is how a third
phantom-broker site was nearly missed).

```bash
# Every Kafka dial site in the repo. Expect: consumer.go (NewReader),
# producer.go (Writer), topic_manager.go x8, kafka_reachability.go (probe),
# remote-job-spawner (reader + writer). Anything else under test/ is test-only.
grep -rn "kafka\.NewReader\|kafka\.Dial(\|kafka\.DialContext(\|kafka\.Writer{\|kafka\.DialLeader" \
  --include=*.go . | grep -v "_test\|/test/"

# Anything still on kafka-go's DefaultTransport (nil Transport). Expect: none —
# both Writer sites must carry a Transport, or the connection pool is split.
grep -rn -A6 "kafka\.Writer{" --include=*.go . | grep -v "_test\|/test/" \
  | grep -E "Writer\{|Transport"

# Callers of the metrics server. Expect exactly one (cmd/agent-chassis).
grep -rn "NewMetricsServer" --include=*.go .

# Phantom broker literals. Expect only comments.
grep -rn "kafka-headless" --include=*.go .
```

## 9. Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <submission.json>
```

Submission for this case: correlation `7abe1a57-e3db-4b71-9e3a-744fbf8c24b1`.
Three rounds, all under that one correlation: **REJECTED** (guardian hard veto) ->
**REVISE** (gating objection) -> **REVISE** (an *unreadable* seat, not an objection).

**Read `decided_by` before reading the decision.** Round 3's REVISE was
`decided_by: unreadable reviewer(s): review_editquality.result` — the gating seat
died mid-run. On the substance that round was 6 approve / 2 object, with the
guardian down from a hard veto to one MEDIUM and two LOWs it explicitly filed
"rather than blocking". A REVISE caused by a dead seat is a harness failure and
says nothing about the change; treating it as a verdict would be reading noise as
signal. (Same landmine as `bugs_closed/029`: 3 of its 5 rounds died this way.)

**No `Council-Reviewed:` trailer is on any of these commits, and none may be
added.** The trailer is earned by an APPROVED verdict only; on a REVISE it would
be a permanent false claim of review. The 098 coverage report will therefore list
this work as un-reviewed for ever, which is a known and accepted false negative —
the verdict trail is recorded here instead.

**Gotcha: verify every `grounded_in` quote is byte-exact before firing.**
Reviewers cannot open files, so a trimmed quote is a different claim and draws
objections against code that is actually fine. Two of mine were wrong on the
first pass — a markdown line that is wrapped in the original, and a path prefixed
with the word "vendored" so it no longer resolved. Checker:

```bash
python3 - <<'PY'
import json, re, subprocess
d=json.load(open('<submission.json>'))
for q in d['plan']['grounded_in']:
    m=re.search(r'`(.+)`\s*$', q, re.S)
    if not m: continue
    path=q.split(' ')[0].split(':')[0]
    src=subprocess.run(['git','show',f'HEAD:{path}'],capture_output=True,text=True).stdout
    print('EXACT' if m.group(1) in src else 'MISMATCH', path)
PY
```

Verdict:

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;

SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

**A missing orchestration row is queue latency, not a dropped dispatch.** Do not
re-fire; a duplicate spends the same credits and lands further back in the same
lane. Check depth by re-running the trigger script.
