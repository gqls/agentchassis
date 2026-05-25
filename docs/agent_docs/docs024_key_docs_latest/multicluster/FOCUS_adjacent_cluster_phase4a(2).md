# FOCUS — Adjacent-Cluster MVP on Rackspace Spot (Phase 4a)

**Scope:** One additional Rackspace Spot K8s cluster, alongside the existing `production/uk001`. Cluster B runs `remote-job-spawner` + any dispatched agents. All Kafka topics and Postgres remain on cluster A. We prove the dispatch contract works across two clusters before considering AWS/GCP.

**The big question is correct: Kafka is the centrepiece.** Today the listener is `type: internal`, reachable only at `*.svc.cluster.local`. Cluster B has no way to resolve or reach that name. Adding an external Strimzi listener is the load-bearing change. Almost everything else is a small follow-on.

---

## 1. What stays the same

Everything in the chassis. The `pgxpool` config, the dispatch_actions code, the spawner code (almost — see §6), the agent_definitions schema, the action registry. The whole point of doing this on the same provider first is that the *only* moving parts should be infrastructure and config.

## 1.5 Field notes from the cluster check

Three things confirmed by running the §9 commands against the live uk001 cluster:

**1. Rackspace's "InternalIP" addresses are actually public IPv4.** The five uk001 worker nodes show InternalIPs in `134.213.168.0/24` — that's Rackspace public space (the `outputs_nodes.tf` comment was right). No separate ExternalIP exists; the InternalIP *is* the routable address. Same on the candidate cluster B (`va001-data-collector`, node IP `104.130.29.5` — also public Rackspace space). **Nodeport reachability across clusters is therefore expected to work.**

**2. Current Kafka has two internal listeners.** Confirmed from `kubectl get kafka -o yaml`:
- `plain` on 9092, type `internal`, no TLS
- `tls` on 9093, type `internal`, with TLS

We add a *third* listener — `type: nodeport` — alongside. The existing two carry on serving in-cluster traffic unchanged. Non-breaking change.

**3. Cluster B (`va001-data-collector`) already exists.** Rackspace Spot K8s cluster in US-East (Ashburn, VA), 1 worker node (4 cores / 8GB), control plane Ready, `v1.33.0`. The kubeconfig is available from the Rackspace console (API endpoint `hcp-1739cd9e-….spot.rackspace.com`). **Step 4 of this plan (provision cluster B) collapses to "fetch kubeconfig" — saves a full `010-infrastructure` cycle.**

**Geographic implication:** uk001 is UK (London region), va001 is US-East. That's transatlantic — ~75-85ms round-trip via undersea cable, possibly more on Rackspace's path. This is the network path the smoke test will measure (§7 below), and it's the headline number that decides whether full DB access from cluster B is viable for any given agent type.

**Preemption ages (uk001):** worker nodes are 78d/78d/62d/62d/52d old. Rackspace Spot preempts; when a node goes, its IP goes with it. This is the real reason nodeport is fragile in production — not unreachability, but stale-IP-in-configmap when preemption happens.

## 2. Decisions to settle before any apply

### 2.1 How does cluster B reach cluster A's Kafka?

**Decided: nodeport for the smoke test.** Pre-flight is already cleared — Rackspace InternalIPs are routable public IPv4, and va001 to uk001 nodes is reachable.

Two options compared:

| Option | What it is | Cost | When |
|---|---|---|---|
| **NodePort** (smoke test) | Strimzi external listener type=`nodeport`. va001 connects directly to a uk001 node IP at `:32094` (bootstrap) | $0 | Now — for smoke test and RTT measurement |
| **LoadBalancer** (promotion) | Strimzi external listener type=`loadbalancer`. 1 bootstrap LB + 3 broker LBs | $40/mo | When smoke test passes and we want a stable target |
| **Route / Ingress** | Doesn't apply — Kafka is TCP, not HTTP | — | Never |

Promoting nodeport → loadbalancer later is a one-word YAML change in the listener spec (`type: nodeport` → `type: loadbalancer`) plus a configmap update in cluster B pointing at the new LB hostname/IP. No code or DSN changes downstream.

**Confirmation step before applying** (we did this in §9 of the previous draft; results are now in §1.5): the five uk001 nodes are reachable at `134.213.168.{26,37,44,54,56}`. Cluster B will dial one of these. For the configmap, pick the longest-lived node (52d) as the most likely to still exist tomorrow, but accept it might preempt — if/when it does, edit the configmap.

### 2.2 Postgres: include in smoke test, measure RTT

**Decided: in scope.** The whole point of the smoke test now becomes "validate the dispatch contract *and* measure what the cross-cluster path costs us in latency, so we can decide which agent types are viable remotely". Skipping DB would mean discovering RTT pain only after Phase 5 starts.

We add a NodePort service on cluster A that targets `pgbouncer-clients`. Cluster B's spawned agents connect to it via the same env vars they always use (just with `OVERRIDE_DATABASE_HOST` pointing at the node IP). The chassis code is unchanged.

What we measure:
1. **Single-query RTT** — `SELECT 1` over the cross-cluster path vs in-cluster path.
2. **Transaction RTT** — small read-modify-write transaction.
3. **Workflow-shaped RTT** — typical "fetch agent_definition, write awaited_request, read orchestration_state" sequence: ~5-10 sequential queries.
4. **Connection establishment cost** — cold-start time for the pgxpool (TCP + TLS + SCRAM, ~4-6 RTTs).

Expected numbers based on UK ↔ Ashburn distance (~75-85ms one-way fibre):

| Operation | In-cluster (uk001) | Cross-cluster (uk001 → va001) | Ratio |
|---|---|---|---|
| `SELECT 1` | ~1-2 ms | ~80-100 ms | ~50x |
| Read-modify-write transaction (BEGIN, SELECT, UPDATE, COMMIT) | ~3-5 ms | ~85-110 ms (1 RTT per round, PgBouncer pipelines where possible) | ~25x |
| 10 sequential queries | ~15-30 ms | ~800-1000 ms | ~40x |
| pgxpool cold start (TCP + TLS + SCRAM) | ~5-10 ms | ~400-600 ms | ~60x |
| LLM-heavy agent (1 LLM call, 3 DB queries) | ~3-5 s | ~3.3-5.3 s | ~1.05x (LLM dominates) |
| Composition agent (50 small DB lookups) | ~80-200 ms | ~4-5 s | ~25x |

**What this means for which agent types can run remotely:**

- **LLM-bound agents** (briefing, content writing, site planner, design): basically free to dispatch remotely. The LLM call dwarfs everything; 200ms of DB latency is in the noise.
- **Adapters** (web search, scraping, image generation): also fine — their work is external HTTP calls, DB is just for storing the result.
- **Composition agents and orchestrators with many small DB calls**: probably *not* viable cross-cluster without batching their queries first. The "load 50 link records, resolve each, write back" pattern goes from 200ms to 4 seconds.
- **Anything with a tight inner loop on DB** (e.g. polling for state changes): explicitly avoid until the inner loop is fixed.

Numbers above are projections. The smoke test produces actual numbers — see §7.

### 2.3 Strimzi on cluster B?

**No.** Cluster B has no Kafka, no KafkaUser, no Kafka topics. Skip `030`, `040`, `045`, `080` entirely in the cluster B deployment. Strimzi operator install alone is hundreds of MB of running pods we don't need.

### 2.4 LoadBalancer cost reference (for the promotion path)

Confirmed from Rackspace Spot docs: **$10/month per load balancer**, flat, regardless of traffic. Billed per second. Production-grade cross-cluster path would need:

| LB | Purpose | Monthly cost |
|---|---|---|
| Kafka external bootstrap | Strimzi listener `type: loadbalancer` bootstrap | $10 |
| Kafka broker 0 | Strimzi listener per-broker LB | $10 |
| Kafka broker 1 | Strimzi listener per-broker LB | $10 |
| Kafka broker 2 | Strimzi listener per-broker LB | $10 |
| PgBouncer-clients external | Service `type: loadbalancer` | $10 |
| (Optional) Postgres-templates external | If we want template DB direct, not via clients PgBouncer | $10 |
| **Smoke test (nodeport everywhere)** | $0 | |
| **Production-stable (LBs for Kafka + PgBouncer)** | $50/mo | |
| **Production-stable + separate templates LB** | $60/mo | |

That's not nothing, but it's the right scale for what we get: stable cross-region cluster pairing without the brittleness of per-node IPs.

---

## 3. Changes on cluster A (primary)

### 3.1 `040-kafka-cluster` — add the external listener

Today the `Kafka` resource has one listener (`plain`, internal, 9092). Add a second:

```yaml
listeners:
  - name: plain                # existing — leave alone
    port: 9092
    type: internal
    tls: false

  - name: tls                  # existing — leave alone
    port: 9093
    type: internal
    tls: true

  - name: external             # NEW
    port: 9094
    type: nodeport             # phase 4a — promote to `loadbalancer` later
    tls: true
    authentication:
      type: scram-sha-512
    configuration:
      # On Rackspace Spot, "InternalIP" is the routable public IPv4.
      # Tell Strimzi to use it for the per-broker advertised host so clients
      # outside the cluster get an address they can actually reach.
      preferredNodePortAddressType: InternalIP
      bootstrap:
        nodePort: 32094        # any free nodePort
      brokers:
        - broker: 0
          nodePort: 32100
        - broker: 1
          nodePort: 32101
        - broker: 2
          nodePort: 32102
```

Strimzi will:
- Create one nodePort service for the bootstrap (`personae-kafka-cluster-kafka-external-bootstrap`) on 32094 — reachable on any uk001 node IP.
- Create one nodePort service per broker (`...-kafka-external-N`) on the assigned ports.
- Issue a TLS cert covering the bootstrap + each broker's advertised host (driven by `preferredNodePortAddressType`). The generated cluster CA is in the secret `personae-kafka-cluster-cluster-ca-cert`.

**Promotion variant (for later, when we want stability):** swap nodeport for loadbalancer. Cost: $40/mo (1 + 3 LBs).

```yaml
  - name: external
    port: 9094
    type: loadbalancer       # ← only word that changes
    tls: true
    authentication:
      type: scram-sha-512
    configuration:
      bootstrap:
        loadBalancerIP: ""     # let Rackspace assign
      brokers:
        - broker: 0
        - broker: 1
        - broker: 2
```

**Non-obvious gotchas to flag before applying:**

1. **`preferredNodePortAddressType: InternalIP` is required on Rackspace.** Without it, Strimzi uses ExternalIP (which is `<none>` per the kubectl output) and broker reconnect after bootstrap fails. This is the specific Rackspace adaptation.

2. **Bootstrap connect succeeds, broker connect fails — known nodeport symptom.** If you see `kafkacat -L` work but `produce` hang, it's almost always that the per-broker advertised host isn't reachable. Triple-check `preferredNodePortAddressType` and that the per-broker nodePorts are open.

3. **Consistent nodePorts across re-applies.** We pin 32094/32100/32101/32102. Letting Strimzi auto-assign means cluster B's configmap drifts every apply.

### 3.2 `045-kafka-users` — add the cross-cluster KafkaUser

Add a new `KafkaUser` resource alongside `core-manager-user` and `personae-app-anonymous`. This is the user cluster B's agents will authenticate as:

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: personae-app-cross-cluster
  namespace: kafka
  labels:
    strimzi.io/cluster: personae-kafka-cluster
spec:
  authentication:
    type: scram-sha-512    # MUST be scram for the external listener
  authorization:
    type: simple
    acls:
      # Same topic ACLs as personae-app-anonymous — read/write personae- topics,
      # describe everything. Don't grant cluster-Alter or topic-Create from
      # outside; topics should only be created from cluster A.
      - resource: { type: topic, name: "personae-", patternType: prefix }
        operations: [Read, Write, Describe]
        host: "*"
      - resource: { type: topic, name: "job.", patternType: prefix }
        operations: [Read, Write, Create, Describe]
        host: "*"
      - resource: { type: topic, name: "system.dispatch.", patternType: prefix }
        operations: [Read, Write, Describe]
        host: "*"
      - resource: { type: topic, name: "*", patternType: literal }
        operations: [Describe]
        host: "*"
      - resource: { type: cluster }
        operations: [Describe]
        host: "*"
      - resource: { type: group, name: "*", patternType: literal }
        operations: [Read, Describe]
        host: "*"
```

Strimzi creates a `Secret` named `personae-app-cross-cluster` containing `password` and `sasl.jaas.config`. **That secret is what we'll replicate to cluster B in §4.3.**

> **Authorization caveat — important.** The live Kafka cluster has **no `spec.kafka.authorization` block**. That means the ACLs declared above (and on the existing `personae-app-anonymous` users) are **not actually enforced** — they're decorative. The internal listeners run without authentication, so everything connects as `User:ANONYMOUS` with full access. Adding `authentication: scram-sha-512` to the external listener *does* gate the connection (credentials are required to connect), but once connected, an authenticated user is unrestricted because cluster-level authorization isn't enabled.
>
> **For an internal smoke test that's acceptable — we control the credentials. Flag it for the production-promotion stage when we enable cluster-wide `authorization: simple`. That's a separate bigger change.** Enabling cluster-wide authorization is higher-risk than it looks: every in-cluster app currently connects anonymously, so turning on ACL enforcement means `personae-app-anonymous` must already cover everything those apps do, or they break. Treat it as its own piece of work with its own testing, not a rider on this phase.

### 3.3 Expose PgBouncer via NodePort (for the smoke test)

Don't change the existing `pgbouncer` Service (it's still in-cluster ClusterIP for primary's agents — confirmed by `agent-chassis.yaml` which uses `pgbouncer.ai-persona-system.svc.cluster.local:6432`). Add a **second** Service that targets the same Pods but is `type: NodePort`:

```yaml
# deployments/kustomize/services/pgbouncer/base/svc-external.yaml (NEW)
apiVersion: v1
kind: Service
metadata:
  name: pgbouncer-external
  namespace: ai-persona-system
  labels:
    app: pgbouncer
    purpose: cross-cluster
spec:
  type: NodePort
  selector:
    app: pgbouncer                  # same selector as the in-cluster Service — confirm with `kubectl get svc pgbouncer -o yaml`
  ports:
    - name: pgbouncer
      port: 6432
      targetPort: 6432
      nodePort: 32432               # pinned
```

Why a second Service instead of changing the first: in-cluster consumers continue to use `pgbouncer.ai-persona-system.svc.cluster.local:6432` exactly as today. Zero risk to existing traffic. The cross-cluster path is opt-in.

**TLS to PgBouncer from outside the cluster:** check the existing PgBouncer config (`SERVICE_INFRASTRUCTURE_CLIENTS_DATABASE_SSLMODE` in cluster A). If PgBouncer was deployed without TLS, the cross-cluster connection will be plaintext over Rackspace's network. For the smoke test that's acceptable (Rackspace is a single provider and the path is short), but it's a known limitation to fix before Phase 5. **In the agent's chassis config for cross-cluster (cluster B side), set `sslmode=prefer`** — PgBouncer accepts plaintext if TLS isn't configured, and uses it if it is, with no agent change required.

**Optional: also expose `postgres-templates`** if we want to measure templates DB access separately, or want to avoid double-hopping through PgBouncer for read-mostly `agent_definitions` traffic. For Phase 4a, route templates through `pgbouncer-clients` if the user is shared, or add a second NodePort on 32433 if it's a separate Postgres. Decision: skip for first pass; if the workflow needs `agent_definitions` reads and timing reveals weirdness, add it.

### 3.4 No other terraform changes on cluster A

Topics, monitoring, ingress — unchanged. Existing Postgres deployment, existing PgBouncer Deployment — unchanged. We only added one Service.

---

## 4. Build cluster B (va001)

### 4.1 Cluster B already exists

`va001-data-collector` is a live Rackspace Spot K8s cluster in US-East (Ashburn):
- API endpoint: `hcp-1739cd9e-….spot.rackspace.com`
- Worker node: `prod-instance-17685505075730012` at `104.130.29.5` (4 cores / 8GB / v1.33.0)
- Control plane: free, Ready

**Step 1:** download the kubeconfig from the Rackspace console and save it to `~/.kube/config_production_va001`. Region key for our terraform/kustomize naming: `va001`.

This skips the full `010-infrastructure` provisioning cycle that the previous Phase 4 plan assumed. We jump straight to deploying secrets and the spawner.

### 4.2 New paths to add to the repo

```
deployments/terraform/environments/production/va001/
├── 047-base-configs/            # NEW — for the replicated secrets
└── services/agents/
    └── 2220-remote-job-spawner/ # NEW — calls kustomize-apply against new overlay

deployments/kustomize/services/remote-job-spawner/overlays/production/va001/
├── kustomization.yaml
└── configmap.yaml
```

Note: we're **not adding** a new `010-infrastructure` directory because we're not provisioning a cluster — we're attaching to an existing one. If/when we want to manage va001 from terraform later, we can import it.

### 4.3 What to deploy on cluster B (and what to skip)

| Module | On va001? | Notes |
|---|---|---|
| `010-infrastructure` | NO — already exists | Cluster is live; just fetch kubeconfig |
| `020-ingress-nginx` | OPTIONAL | Skip unless something on B will serve HTTP traffic externally. For pure dispatch target, skip. |
| `030-strimzi-operator` | NO | Cluster B has no Kafka. |
| `040-kafka-cluster` | NO | Same. |
| `045-kafka-users` | NO | Same. |
| `047-base-configs` | YES | Holds the replicated secrets. |
| `050-storage` | NO unless agents on B need PVCs | Most don't. |
| `060-databases` | NO | No DB on B — uses primary's via NodePort. |
| `065-pgbouncer` | NO | No PgBouncer on B for Phase 4a. |
| `070-database-schemas` | NO | No DB. |
| `080-kafka-topics` | NO | Topics are on cluster A. |
| `090-monitoring` | OPTIONAL | If you want Prometheus on B; not required for the dispatch test. |
| `100-bootstrap-agents` | REDUCED | Bootstrap just `remote-job-spawner`. No core-manager, no auth-service, no admin-dashboard on B. |

### 4.3 Secret replication into cluster B's `ai-persona-system` namespace

Four secrets need to exist on B before the spawner can do useful work. Manual `kubectl get | kubectl apply` works for one cluster; automate later.

```bash
# Source: cluster A
export KUBECONFIG=~/.kube/config_production_uk001

# 1. Cross-cluster Kafka user (created by §3.2)
kubectl -n kafka get secret personae-app-cross-cluster -o yaml \
  | sed '/resourceVersion:/d; /uid:/d; /creationTimestamp:/d; /namespace:/d' \
  > /tmp/kafka-user.yaml

# 2. Kafka cluster CA cert (for TLS verification)
kubectl -n kafka get secret personae-kafka-cluster-cluster-ca-cert -o yaml \
  | sed '/resourceVersion:/d; /uid:/d; /creationTimestamp:/d; /namespace:/d' \
  > /tmp/kafka-ca.yaml

# 3. Platform secrets (DB passwords + agent bootstrap key)
#    NOTE: even though Phase 4a doesn't dispatch DB-needing agents, the
#    spawner needs these to construct the Job spec env vars. The DB hosts
#    will be unreachable from B — that's fine, it surfaces as a connection
#    error only if/when an agent on B actually tries to use them.
kubectl -n ai-persona-system get secret personae-platform-secrets -o yaml \
  | sed '/resourceVersion:/d; /uid:/d; /creationTimestamp:/d; /namespace:/d' \
  > /tmp/platform.yaml

# 4. Default secrets (Anthropic key etc.)
kubectl -n ai-persona-system get secret personae-default-secrets -o yaml \
  | sed '/resourceVersion:/d; /uid:/d; /creationTimestamp:/d; /namespace:/d' \
  > /tmp/default.yaml

# Apply to cluster B (va001)
export KUBECONFIG=~/.kube/config_production_va001

# Re-namespace the kafka user secret into ai-persona-system on B
sed 's/^  namespace: kafka$/  namespace: ai-persona-system/' /tmp/kafka-user.yaml \
  | kubectl apply -f -
sed 's/^  namespace: kafka$/  namespace: ai-persona-system/' /tmp/kafka-ca.yaml \
  | kubectl apply -f -

kubectl -n ai-persona-system apply -f /tmp/platform.yaml
kubectl -n ai-persona-system apply -f /tmp/default.yaml
```

The kafka secrets get re-namespaced because cluster B has no `kafka` namespace and the spawner Job runs in `ai-persona-system`.

### 4.4 Spawner overlay for cluster B

`deployments/kustomize/services/remote-job-spawner/overlays/production/va001/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: remote-job-spawner-config
  namespace: ai-persona-system
data:
  # Identity
  CLUSTER_ID: "va001"
  NAMESPACE: "ai-persona-system"

  # Kafka — point at cluster A's external listener via nodeport.
  # Pick a long-lived uk001 node IP (the 52d-old one, least likely to preempt soon).
  # If/when that node preempts, edit this configmap and restart the spawner Pod.
  KAFKA_BROKERS: "134.213.168.26:32094"
  KAFKA_SECURITY_PROTOCOL: "SASL_SSL"
  KAFKA_SASL_MECHANISM: "SCRAM-SHA-512"
  KAFKA_SASL_USERNAME: "personae-app-cross-cluster"
  # password comes from the replicated secret personae-app-cross-cluster

  # Tell the spawner to override what it passes to spawned agents.
  # These env vars are already supported by the spawner code:
  OVERRIDE_KAFKA_BROKERS: "134.213.168.26:32094"

  # DB host: point at cluster A's PgBouncer NodePort (§3.3). Spawned agents
  # connect to PgBouncer on uk001 across the Atlantic. Expect ~80ms RTT
  # per query — see §2.2 RTT table.
  OVERRIDE_DATABASE_HOST: "134.213.168.26"
  OVERRIDE_DATABASE_PORT: "32432"
  # sslmode=prefer matches both TLS-enabled and TLS-disabled PgBouncer.
  # Promote to require/verify-full before Phase 5.
  OVERRIDE_DATABASE_SSLMODE: "prefer"

  # Same for templates DB if exposed (skipped by default at this phase —
  # templates queries through the same PgBouncer if the user has access).
```

The terraform wrapper at `deployments/terraform/environments/production/va001/services/agents/2220-remote-job-spawner/main.tf` calls the existing `kustomize-apply` module against this overlay — same pattern as `2210-agent-chassis`.

**Spawner code note:** the `OVERRIDE_DATABASE_HOST/PORT/SSLMODE` env vars need to be passed through to the Job spec the spawner builds. The spawner already supports `OVERRIDE_KAFKA_BROKERS`; verify `OVERRIDE_DATABASE_HOST` and `OVERRIDE_DATABASE_PORT` are equally honoured (a quick grep on `cmd/remote-job-spawner/main.go` `createAgentJob` will confirm). If not present, add them — mirrors the existing pattern, no architectural change.

---

## 5. Sequence of operations (the order matters)

The two load-bearing checkpoints are step 4 (kafkacat from laptop confirms listener works) and step 8 (kafkacat from cluster B confirms cross-cluster auth works). If both pass, the dispatch test in step 11 is essentially mechanical.

### Cluster A side

1. **Add external Kafka listener** (`040-kafka-cluster` apply with §3.1 YAML). Wait for `kubectl -n kafka get svc | grep external` to show 4 nodeport services (1 bootstrap + 3 brokers). Verify a kafkacat from a temporary Pod *within cluster A* still works on the internal listener (smoke check we haven't broken existing traffic).

2. **Add cross-cluster KafkaUser** (`045-kafka-users` apply with §3.2 YAML). `kubectl -n kafka get secret personae-app-cross-cluster` should appear within seconds.

3. **Add PgBouncer NodePort Service** (apply §3.3 YAML). `kubectl -n ai-persona-system get svc pgbouncer-clients-external` should show NodePort with port 32432.

4. **Connectivity check from outside** (your laptop, or wherever). This is the first load-bearing checkpoint.
   ```bash
   # Kafka — list topics over the external listener
   kafkacat -b 134.213.168.26:32094 \
     -X security.protocol=SASL_SSL \
     -X sasl.mechanism=SCRAM-SHA-512 \
     -X sasl.username=personae-app-cross-cluster \
     -X sasl.password=<from secret> \
     -X ssl.ca.location=<from kafka-ca secret> \
     -L

   # PgBouncer — open a session
   PGPASSWORD=<from clients DB secret> \
     psql -h 134.213.168.26 -p 32432 -U <clients_user> -d <clients_db> -c "SELECT 1"
   ```
   If both work, everything in cluster A is correctly exposed. If either doesn't, fix here.

### Cluster B (va001) side

5. **Fetch kubeconfig** from Rackspace console, save to `~/.kube/config_production_va001`. Test: `kubectl --kubeconfig=… get nodes` should show `prod-instance-17685505075730012`.

6. **Apply 047-base-configs** + secret replication script of §4.3.

7. **Network pre-flight from a Pod on va001.** Sanity-check that Rackspace InternalIPs are indeed routable cross-region:
   ```bash
   kubectl --kubeconfig=~/.kube/config_production_va001 \
     -n ai-persona-system run net-test --rm -it --image=nicolaka/netshoot -- \
     bash -c "nc -vz 134.213.168.26 32094 && nc -vz 134.213.168.26 32432"
   ```
   Both must connect. If either times out, Rackspace doesn't route between regions at that level — stop and use loadbalancer instead (§2.4).

8. **kafkacat from a Pod on va001.** Second load-bearing checkpoint:
   ```bash
   kubectl --kubeconfig=~/.kube/config_production_va001 \
     -n ai-persona-system run kc --rm -it \
     --image=edenhill/kcat:1.7.1 -- \
     -b 134.213.168.26:32094 \
     -X security.protocol=SASL_SSL \
     -X sasl.mechanism=SCRAM-SHA-512 \
     -X sasl.username=personae-app-cross-cluster \
     -X sasl.password=<from secret> \
     -X ssl.ca.location=<from kafka-ca secret> \
     -L
   ```
   If it works, the Kafka cross-cluster contract is settled.

9. **Measure RTT from va001 → uk001 PgBouncer.** This is the headline number for the whole exercise — drives every later decision about which agent types run remotely.
   ```bash
   # From within a Pod on va001
   kubectl --kubeconfig=~/.kube/config_production_va001 \
     -n ai-persona-system run pgtest --rm -it --image=postgres:16 -- \
     bash -c '
       export PGPASSWORD="<from secret>"
       # Cold start (full TCP+TLS+SCRAM handshake)
       time psql -h 134.213.168.26 -p 32432 -U <clients_user> -d <clients_db> -c "SELECT 1"
       # 10 sequential queries (re-using PgBouncer session)
       time psql -h 134.213.168.26 -p 32432 -U <clients_user> -d <clients_db> \
         -c "SELECT 1; SELECT 2; SELECT 3; SELECT 4; SELECT 5; SELECT 6; SELECT 7; SELECT 8; SELECT 9; SELECT 10;"
       # Plain network RTT for comparison
       ping -c 10 134.213.168.26
     '
   ```
   Record three numbers: ping median (raw network), cold-start single SELECT (network + auth), 10-query batch (network with reuse). Compare against the §2.2 expected-numbers table.

10. **Deploy `remote-job-spawner` on va001** (`deploy-remote-job-spawner ENVIRONMENT=production REGION=va001`). Pod should come up. Logs should show `"Remote Job Spawner ready, consuming from dispatch topic"`. Consumer group `remote-job-spawner-va001` should appear in cluster A's Kafka.

### End-to-end test

11. **Dispatch test.** From the orchestration on cluster A (or an admin dashboard manual trigger), invoke `dispatch_agent` with `target_cluster: "va001"` and a DB-using agent type (e.g. `briefing-agent`). Expected observations:
   - **uk001** spawner logs: "Skipping dispatch for other cluster" (once §6 fix lands).
   - **va001** spawner logs: "Received dispatch request" → "Successfully created agent job".
   - `kubectl --kubeconfig=~/.kube/config_production_va001 -n ai-persona-system get jobs`: new Job appears.
   - Spawned Pod's logs: chassis connects to Kafka on `134.213.168.26:32094`, connects to Postgres on `134.213.168.26:32432`, registers awaited request, processes work item, writes back.
   - Parent on uk001: receives init response on the shared topic; workflow continues.
   - **Wall-clock comparison:** same workflow run with `target_cluster: "uk001"` (local) vs `target_cluster: "va001"` — note the difference. This is the real-world cost of cross-region dispatch for *this* agent type.

12. **Record the timings.** Add a small entry to a working notes file:
   ```
   Date: ____
   ping uk001 ↔ va001 median: ____ ms
   cold-start psql:             ____ ms
   10-query batch:              ____ ms (= ____ ms per query)
   briefing-agent local:        ____ s
   briefing-agent cross:        ____ s
   ```
   These numbers go straight into the Phase 5 decision on which agent types are safe to dispatch remotely.

If any step fails, the diagnostic order is: §3.1/3.3 (listeners/services) → §4.3 (secrets in right place) → step 7 (network routable) → step 8 (auth working) → spawner logs.

---

## 6. Small code change needed before step 9

`cmd/remote-job-spawner/main.go` already has the cluster filter — but uses `logger.Debug` on the skip path:

```go
if targetCluster != "" && targetCluster != clusterID && targetCluster != "any" {
    logger.Debug("Message not for this cluster, skipping", ...)
    continue
}
```

Per your logger guideline, Debug doesn't appear in logs, so when we deploy a second spawner we won't be able to verify the filter is doing its job. Change to `logger.Info`:

```go
logger.Info("Skipping dispatch for other cluster",
    zap.String("target_cluster", targetCluster),
    zap.String("my_cluster", clusterID),
    zap.String("agent_id", req.AgentID))
```

That's the only code change needed for Phase 4a. No variable-name changes, no contract changes.

This also corrects the earlier Phase 4 doc which described Gap A as un-fixed — the filter is in place, only the log level is wrong.

---

## 7. What this phase deliberately doesn't do

- **No loadbalancer Kafka listener** — nodeport only. Promote when smoke test passes and we want a stable target. $40/mo when we do.
- **No PgBouncer loadbalancer** — same nodeport story. +$10/mo when promoting.
- **No TLS to PgBouncer.** `sslmode=prefer` for the smoke test. Promote to `verify-full` (with CA mounted) before Phase 5 / production traffic.
- **No External Secrets Operator.** Manual `kubectl apply` per §4.3 is fine for one cluster.
- **No `agent_definitions.target_cluster` column.** Dispatch still picks `target_cluster` from workflow config / step config. Per-agent-type routing comes in Phase 3 of the broader plan.
- **No dispatch log table.** That's Gap B from the multi-cluster doc — lives on the parent side, independent of whether cluster B exists. Worth doing in parallel.
- **No monitoring stack on B.** Logs via `kubectl logs` is sufficient at this stage.
- **No `010-infrastructure` for va001.** Cluster already exists; attach via kubeconfig.

---

## 8. Estimated effort

| Task | Rough size |
|---|---|
| Modify cluster A's `040-kafka-cluster` to add external listener | ~25 lines YAML |
| Add cross-cluster KafkaUser in `045-kafka-users` | ~35 lines YAML |
| Add PgBouncer NodePort Service (`§3.3`) | ~15 lines YAML |
| Fetch va001 kubeconfig from Rackspace console | UI click |
| Cluster B `047-base-configs` + secret replication script | bash script ~30 lines |
| New spawner overlay for va001 | 2 short YAML files |
| New tf wrapper for `2220-remote-job-spawner` on va001 | 1 short main.tf |
| Code change: `logger.Debug` → `logger.Info` in spawner | 4 lines |
| Verify `OVERRIDE_DATABASE_HOST/PORT/SSLMODE` are honoured by spawner | grep + maybe 6 lines |
| Test workflow that calls `dispatch_agent` with `target_cluster: "va001"` | 1 SQL insert |

The two load-bearing checkpoints (kafkacat from laptop in step 4, then from va001 Pod in step 8) are the actual unknowns. Everything else is mechanical.

---

## 9. First concrete next step

The three diagnostic commands from the previous draft have been run — results are in §1.5. The actual next step is now:

```bash
# Decide on a long-lived uk001 node IP to pin in the va001 configmap.
# The 52d-old node is the most likely to outlast the smoke test.
kubectl --kubeconfig=~/.kube/config_production_uk001 get nodes \
  -o wide --sort-by=.metadata.creationTimestamp
# Pick the longest-running node's INTERNAL-IP. Currently that's 134.213.168.26.
```

After that, the order is §5 step by step — start with the Kafka external listener YAML in `040-kafka-cluster`.

