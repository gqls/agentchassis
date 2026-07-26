# 082 — the clients database runs BestEffort (drifted from its own manifest) and a 1s exec probe kills it whenever a neighbour saturates the node

**Filed:** 2026-07-26 · **By:** gauntlet_dead_cta (P4 front-end rebuild) — hit it
mid-delivery, did not cause it · **Severity:** HIGH — fleet-wide outage while it
lasts · **Status:** **CLOSED 2026-07-26 — FIXED AND LIVE on both databases**
(commit `0f87d30c8`, applied by terraform)

> **CLOSING NOTE — the title of this file is wrong and is kept only so the
> number resolves.** There was no "drift". See §0. The mechanism below (§1, §2
> Fault B, §3, §5) was exact and is what made the fix quick; the *location* of
> the cause was not.
>
> Workstream docs:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_082_postgres_qos/`

---

## 0. CORRECTION 2026-07-26 — Fault A is REFUTED. There was no drift.

**The claim:** §2 Fault A said the live StatefulSet *"has drifted from the
checked-in manifest, and the drift removed every resource guarantee"*, citing
`deployments/kustomize/infrastructure/postgres-clients/postgres-clients.yaml`,
and §6.1 prescribed restoring that block as *"reconciliation, not a new design —
the reviewed desired state has said this since day one"*.

**Why it is false:** that manifest has never been applied to anything. The
`kustomization.yaml` beside it is **0 bytes** and no kustomization in the repo
lists it (all six `infrastructure/*/kustomization.yaml` are 0 bytes). The live
object is built by **`deployments/terraform/modules/postgres-instance/main.tf`**,
instantiated at `environments/production/uk001/060-databases/main.tf`. That
module never specified `resources`. The database was not demoted to BestEffort —
it was **born** BestEffort at cluster build and had never been anything else.

The live object matches Terraform on **all seven** properties where the two
candidate sources disagree:

| property | live + terraform | the orphaned manifest |
|---|---|---|
| serviceName | `postgres-clients-headless` | `postgres-clients` |
| image | `pgvector/pgvector:pg15` | pg16 |
| containers | `postgres` | postgres + exporter |
| probe command | `pg_isready -U … -d …` | no `-d` flag |
| securityContext | fsGroup/runAsUser 999 | absent |
| password source | `envFrom postgres-clients-secret` | `db-secrets` secretKeyRef |
| storage | 100Gi ssd-large | 10Gi standard |

**What caught it:** this file's own evidence. §2 notes *"the live probe also
carries `-d clients_db`, which the manifest does not — the same drift, visible
twice."* That is not drift twice. **A live object cannot invent a command-line
argument its manifest never contained.** Drift subtracts; it does not add. One
unexplained *addition* is the signature of a different source.

**Why the correction mattered practically:** §6.1's `kubectl patch` would have
been correct for about a minute and then silently reverted by the next
`terraform apply` — leaving an intermittent bug and a manifest still lying to
the next reader.

**Cheap check for next time:** before asking *"has the live object drifted from
this file?"*, ask **"does anything apply this file?"**
`ls -la deployments/kustomize/infrastructure/*/kustomization.yaml` (0 bytes =
orphaned) and `grep -rn "<name>" --include="kustomization.yaml" deployments/`
(no hits = nobody applies it). Logged in `WRONG_CALLS.md`.

**Related decoy, NOT fixed here:** `scripts/deploy-system.sh:129` runs
`kubectl apply -f k8s/postgres-clients.yaml` — **that file does not exist**.
Three files in this repo are named for this database and two are dead. Left
alone as out of scope; recorded so it is not rediscovered as a mystery.

## 0b. What was actually fixed, and how it was verified

Both faults live in the Terraform module, so one edit fixed **both** databases —
`postgres-templates` had the identical defect and nobody had noticed, because it
is colder and had never been co-scheduled with a noisy neighbour.

- `resources`: requests `500m` / `512Mi`, limits `2000m` / `2Gi`, via four new
  module variables **with defaults**, so a future instance cannot silently be
  BestEffort again.
- `timeout_seconds: 5` on both probes (was the inherited 1s default).
- `failure_threshold: 3 → 6` on both. The readiness change was **not** in §6:
  `replicas = 1`, so there is no second backend to fail over to — dropping the
  only endpoint converts "slow" into "no such host" fleet-wide, which is exactly
  the `notReadyAddresses` symptom in §1.

Not council-reviewed: `deployments/` is outside the gate's scope
(`platform/`, `internal/`, `pkg/`) and `097` refuses it client-side.

**Verified in the kernel, not from the spec** — inside the running container:
`cpu.max = 200000 100000` (exactly 2 CPUs) and `memory.max = 2147483648`
(exactly 2 GiB) match the declared limits to the byte. `cpu.weight` measured
against a **positive control** — a still-BestEffort pod in another namespace:

```
BestEffort control (kafka/kcat-cgate-…) : cpu.weight = 1    (kernel minimum)
postgres-clients-0                      : cpu.weight = 59
```

**59× the contended CPU share, measured.** (Do not derive this from a shares
formula — the documented conversion predicts ~20 and this runtime produced 59.)
`terraform plan` afterwards: *"No changes. Your infrastructure matches the
configuration."* `ai-persona-system` now has **zero BestEffort pods**; all 65
are Burstable.

**[UNVERIFIED] — not proven under contention.** §7 warned that verifying during
a quiet period proves nothing. Both databases rescheduled **off** the ollama node
during the roll, so they are no longer co-tenants and the node is quiet. What is
proven: the structural cause is gone. What is not: that 500m is *sufficient*
under a full 8-core inference run (it is ~1.6 guaranteed cores against a database
idling at 30m — arithmetic, not observation).

**Anti-affinity (§6.3) deliberately deferred, with a reversal trigger:** if
`postgres-clients-0` is ever again co-scheduled with `ollama-adapter` **and** its
restart count moves, the floor was insufficient and anti-affinity becomes the fix
rather than an option. Nothing currently pins them apart — today's separation is
luck, not a guarantee.

**Residual — RESOLVED 2026-07-26, no open items.** A production `terraform apply`
was run behind a `timeout` which SIGTERMed it, leaving a **stale state lock**
(Lease `lock-tfstate-default-tfstate-databases` in the `default` namespace,
holder `d3e2fc63-c4c6-8586-45af-db70301eb9c1`). The change itself had landed and
state was converged; only the lock was orphaned. **Cleared by the owner with
`terraform force-unlock`.** Verified two ways: `terraform plan` with **default
locking** now succeeds where it previously failed to acquire — that is the
discriminating test, since a `-lock=false` plan would have passed either way —
and the lease's `holderIdentity` is now blank, matching its ten siblings. State
still reports "No changes. Your infrastructure matches the configuration."

---

## The original filing follows, unaltered except where marked.

> **I did not patch production.** The remedy below touches shared infrastructure
> that every session and every agent depends on, and the trigger (a neighbour pod
> legitimately using its full CPU limit) belongs to another workstream. It is
> written up so the owner — or whoever owns the ollama lane — can act in one
> command. See §6 before touching anything.

## 1. Symptom

`postgres-clients-0` crash-loops. `kubectl` shows the container repeatedly killed
and restarted, and the Service carries **no ready endpoints**, so every in-cluster
client connecting to `postgres-clients:5432` fails:

```
$ kubectl -n ai-persona-system get pod postgres-clients-0 \
    -o jsonpath='ready={.status.containerStatuses[0].ready} restarts={.status.containerStatuses[0].restartCount}'
ready=false restarts=4          # was 1 four minutes earlier — still climbing

$ kubectl -n ai-persona-system get endpoints postgres-clients -o jsonpath='{.subsets}'
[{"notReadyAddresses":[{"ip":"10.20.39.11", ... "name":"postgres-clients-0" ...}],
  "ports":[{"port":5432,"protocol":"TCP"}]}]
   # notReadyAddresses only — the Service has nowhere to send traffic
```

Kubelet's own account:

```
Warning  Unhealthy  Liveness probe failed: command timed out:
                    "pg_isready -U clients_user -d clients_db" timed out after 1s
Normal   Killing    Container postgres failed liveness probe, will be restarted
```

**The database itself is healthy throughout.** Reached directly with `kubectl exec`
(which bypasses the Service) it answers instantly, and its own log says it started
cleanly:

```
$ kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c "SELECT 'alive', count(*) FROM content_components;"
alive|194

2026-07-26 14:56:59.871 UTC [1] LOG:  database system is ready to accept connections
```

Postgres was using **30m CPU and 32Mi RSS** when it was killed for being unhealthy.

## 2. Root cause — two faults, both in our own manifests

> **REFUTED 2026-07-26 — see §0.** Fault A below is wrong. Nothing has ever
> applied that manifest; the live object is built by Terraform and never had
> resources to lose. Fault B (the 1s probe) and the trigger are correct.

**Fault A — the live StatefulSet has drifted from the checked-in manifest, and the
drift removed every resource guarantee.**

`deployments/kustomize/infrastructure/postgres-clients/postgres-clients.yaml:60-66`
has specified resources since the initial commit (`c442d6e77`):

```yaml
          resources:
            requests:
              memory: "512Mi"
              cpu: "500m"
            limits:
              memory: "1Gi"
              cpu: "1000m"
```

The live object has none:

```
$ kubectl -n ai-persona-system get statefulset postgres-clients \
    -o jsonpath='{.spec.template.spec.containers[0].resources}'
{}
$ kubectl -n ai-persona-system describe pod postgres-clients-0 | grep QoS
QoS Class:                   BestEffort
```

`resources: {}` means **BestEffort** — the lowest CPU share the kernel hands out
and the first thing evicted under memory pressure. The manifest's 500m CPU floor
has never applied to the running database. (The live probe also carries
`-d clients_db`, which the manifest does not — the same drift, visible twice.)

**Fault B — a 1-second `exec` probe cannot survive CPU starvation.**

Neither probe in the manifest sets `timeoutSeconds`, so both inherit the
Kubernetes default of **1s** (visible in the live spec). An `exec` probe has to
fork a process inside the container. A BestEffort container on a saturated node
cannot reliably fork, exec and return `pg_isready` within one second — so the
probe reports the database dead when it is merely descheduled.

**The trigger** is a neighbour doing nothing wrong. `ollama-adapter` is scheduled
on the same node and is licensed to take all of it:

```
$ kubectl top node prod-instance-17735925437536833
prod-instance-17735925437536833   7999m   106%   21144Mi   35%

$ kubectl top pod -n ai-persona-system --sort-by=cpu | head -3
ollama-adapter-57f5679794-8tw9b   7743m   15837Mi
postgres-clients-0                  30m      32Mi

$ kubectl -n ai-persona-system get pod -l app=ollama-adapter \
    -o jsonpath='{.items[0].spec.containers[0].resources}'
{"limits":{"cpu":"8","memory":"20Gi"},"requests":{"cpu":"2","memory":"20Gi"}}
```

`limits.cpu: 8` on an 8-core node, actively running inference (llama.cpp slot
logs). It is inside its limits. **The defect is not that ollama is busy — it is
that the database has no floor to stand on and a stopwatch too short to pass.**

So: *neighbour saturates node → BestEffort postgres gets no CPU → 1s exec probe
times out → kubelet kills a healthy database → Service loses its only endpoint →
every agent in the fleet fails its DB calls.* The restart does not help, because
the node is still saturated when the new container's probe runs.

## 3. Blast radius

Fleet-wide for as long as it lasts. Anything reaching Postgres through the Service
is cut off: the chassis, every spawned agent, the section-editor and page-rerender
paths, the diagnosis and council loops. Work already in flight can fail
mid-workflow. `kubectl exec` still works, which is why this can look survivable
from a terminal while every agent is failing.

## 4. What this cost, concretely

The gauntlet_dead_cta P4 front-end delivery stopped here. Component sources were
built and passed a full end-to-end run against the live API (65/65 gauntlet,
31/31 archive) but could not be delivered, because delivery runs through the
section-editor agent, which reaches the database through the Service.

## 5. Not-causes ruled out

- **Not disk/data corruption.** Postgres logs a clean start and a clean prior
  shutdown; queries return correct counts.
- **Not memory pressure.** The node is at 35% memory; postgres holds 32Mi.
- **Not a slow query or connection storm.** Postgres is at 30m CPU when killed.
- **Not caused by this workstream.** Our load in the window was a handful of
  single-row SELECTs; the AI traffic went to `tools.apis.uk` on the island VM,
  which calls Anthropic directly and never touches ollama or this cluster.

## 6. Fix candidates, safest first

> **SUPERSEDED 2026-07-26 — do NOT run candidate 1 as written.** The premise
> ("the reviewed desired state") is refuted in §0, and a `kubectl patch` here is
> reverted by the next `terraform apply`. The fix went into the Terraform module
> instead; §0b records what shipped. Candidate 2 was adopted (and extended to
> readiness), 3 deferred with a trigger, 4 answered: nothing had drifted, but
> nothing reconciles these manifests either — two of the three files named for
> this database are dead.

1. **Restore the resource block the manifest already specifies** (reconciliation,
   not a new design — the reviewed desired state has said this since day one):
   ```
   kubectl -n ai-persona-system patch statefulset postgres-clients --type=strategic -p \
     '{"spec":{"template":{"spec":{"containers":[{"name":"postgres",
       "resources":{"requests":{"cpu":"500m","memory":"512Mi"},
                    "limits":{"cpu":"2000m","memory":"1Gi"}}}]}}}}'
   ```
   This makes it Burstable with a guaranteed 500m floor. **It rolls the pod** —
   acceptable only because the pod is already being killed every ~90s, and
   strictly better after. Consider `limits.cpu: 2000m` rather than the manifest's
   1000m: the DB idles at 30m but should be able to burst.
2. **Give the probes room** — `timeoutSeconds: 5`, `failureThreshold: 6` on both.
   A database that is briefly descheduled is not a database that has died. Fix
   this in the manifest too, so the next apply keeps it.
3. **Keep the two apart.** `ollama-adapter` with `limits.cpu` equal to the whole
   node will starve any co-tenant. Either lower its limit, or add anti-affinity /
   a node selector so the stateful data plane and the inference plane never share
   a node.
4. **Then ask why the live object drifted at all**, because the same class of
   drift is invisible on every other service: nothing reconciles these manifests,
   so a live object can silently lose guarantees its reviewed manifest grants.
   That is the finding with the longest reach here, and it is not postgres-specific.

## 7. How to verify a fix

```
kubectl -n ai-persona-system get pod postgres-clients-0 \
  -o jsonpath='ready={.status.containerStatuses[0].ready} restarts={...restartCount}'
kubectl -n ai-persona-system describe pod postgres-clients-0 | grep QoS   # expect Burstable
kubectl -n ai-persona-system get endpoints postgres-clients -o jsonpath='{.subsets[0].addresses}'
```
Ready `true`, QoS `Burstable`, and a populated `addresses` (not
`notReadyAddresses`). Restart count must then stay flat **while the node is still
loaded** — verifying during a quiet period proves nothing, because the bug only
appears under contention. Re-check `kubectl top node` to confirm the neighbour is
still hot when you call it fixed.

---

## 8. POST-CLOSE ADDENDUM 2026-07-26 — the trigger has a NAME, a CADENCE, and a switch

Added by session "bugfix 61" (owner-authorised) while closing `bugs_closed/061`. **Nothing
here contradicts the analysis above** — §5's trigger account is exact, including the 7743m
measurement. This adds only what drives ollama hot, how often, and **how to make it happen on
demand**, which is what §7 says you need and what the `[UNVERIFIED]` 500m residual lacks.

### What was actually running

The load is the **med-price scraper's LLM fallback** (`bugs_closed/061`, vetcomparison). When
its regex parser finds no prices it sends the page to `ollama-adapter` for extraction. One such
call on 2026-07-26 ran **495 seconds** — and it is a *scheduled* job (`med-scrape-prices`, every
**21,600 s / 6 h**, batch of 20 listings), so this is a recurring, predictable load, not a
one-off. The fallback fires on retailer *category* pages, of which the batch contains several.

Measured from Prometheus across the incident window (node `prod-instance-17735925437536833`,
8 cores):

| time (UTC) | node CPU | `ollama-adapter` |
|---|---|---|
| 14:53 | 3.9 % | 0.04 cores |
| 14:55 | 40.5 % | 1.89 |
| 14:57 | **100.0 %** | **7.75** |
| 14:59 | **100.0 %** | 6.39 |
| 15:01 | 99.7 % | **7.70** |
| 15:03 | 4.1 % | 0.00 |

`kube_pod_info{pod="ollama-adapter-57f5679794-8tw9b"}` at 14:57 → node
`prod-instance-17735925437536833`, i.e. **co-tenant with `postgres-clients-0` at the time**
(measured, not inferred — the pods are unpinned, so this was luck, as §"Anti-affinity" says).

**Third-party confirmation of the blast radius:** an unrelated pod, the spawned
`agent-med-price-collector`, logged `Database ping failed: driver: bad connection` **every 30 s
from 14:55:04 to 15:00:04** — eleven consecutive health ticks — plus
`RETRY_TICKER_CLAIM_FAILED` and a failed awaited-request cleanup. This file's own quoted
postgres line, `14:56:59 database system is ready to accept connections`, falls inside that
window. **It fired twice in twelve minutes:** a second burn at 15:05–15:07 (ollama back to
15.6→7.8 cores) matches the second restart (terminated 15:05:35, container up 15:07:05).

### This makes your `[UNVERIFIED]` testable — a contention switch

§7 warns that verifying during a quiet period proves nothing, and the residual says 500m is
"arithmetic, not observation" against a full 8-core inference run. **You can now produce that
run deliberately**, ~8 minutes of ~7.75 cores, without touching ollama:

```sql
UPDATE business_intel.med_retailer_listings SET last_scraped_at = NULL
WHERE id = '0b50fd2d-a129-4edd-85a0-75843181fe0c';   -- petdrugsonline /advocate (category page)
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'med-scrape-prices';
```

The listing loader orders `last_scraped_at ASC NULLS FIRST`, so this puts the fallback-triggering
page at the head of the next batch. Watch `kubectl top node` and `postgres-clients-0`'s restart
count. **Caveat:** the two databases were rescheduled off the ollama node during the roll, so
today this only reproduces contention if something puts them back together — which is precisely
the reversal trigger already recorded above.

### Measurement trap, recorded because it cost me a wrong number

`sum(rate(container_cpu_usage_seconds_total{pod=~"..."}[2m]))` **double-counts**: the series set
includes both per-container samples and a pod-level cgroup total (empty `container` label). That
gave me **15.4 cores on an 8-core node** — an impossible figure I nearly wrote into this file.
Always filter `container!=""`. The corrected 7.75 agrees with §5's independent `kubectl top`
reading of 7743m, which is what caught it.
