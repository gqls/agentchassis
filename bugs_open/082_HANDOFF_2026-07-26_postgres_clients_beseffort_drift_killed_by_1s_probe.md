# 082 — the clients database runs BestEffort (drifted from its own manifest) and a 1s exec probe kills it whenever a neighbour saturates the node

**Filed:** 2026-07-26 · **By:** gauntlet_dead_cta (P4 front-end rebuild) — hit it
mid-delivery, did not cause it · **Severity:** HIGH — fleet-wide outage while it
lasts · **Status:** OPEN, NOT FIXED, nothing applied

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
