# FOCUS — Adjacent-Cluster MVP on Rackspace Spot (Phase 4a)

**Scope:** One additional Rackspace Spot K8s cluster, alongside the existing `production/uk001`. Cluster B runs `remote-job-spawner` + any dispatched agents. All Kafka topics and Postgres remain on cluster A. We prove the dispatch contract works across two clusters before considering AWS/GCP.

**The big question is correct: Kafka is the centrepiece.** Today the listener is `type: internal`, reachable only at `*.svc.cluster.local`. Cluster B has no way to resolve or reach that name. Adding an external Strimzi listener is the load-bearing change. Almost everything else is a small follow-on.

---

## 1. What stays the same

Everything in the chassis. The `pgxpool` config, the dispatch_actions code, the spawner code (almost — see §6), the agent_definitions schema, the action registry. The whole point of doing this on the same provider first is that the *only* moving parts should be infrastructure and config.

## 2. Decisions to settle before any apply

### 2.1 How does cluster B reach cluster A's Kafka?

Three realistic options on Rackspace Spot:

| Option | What it is | Cost | When |
|---|---|---|---|
| **NodePort** | Strimzi external listener type=`nodeport`. Cluster B connects to `<any cluster-A node IP>:31xxx`. | Free | MVP only; brittle on node replacement |
| **LoadBalancer** | Strimzi external listener type=`loadbalancer`. 1 bootstrap LB + 3 broker LBs. | ~£20–30/mo | Production |
| **Route / Ingress** | Doesn't apply — Kafka is TCP, not HTTP | — | Never for this |

**Recommend: nodeport for MVP.** Promotes to loadbalancer with a small Strimzi YAML change later; no code or DSN changes downstream. The brittleness (broker IPs change when nodes get replaced — and on Rackspace Spot that's "regularly") is acceptable for a smoke-test phase that we expect to last days, not months.

**Pre-flight check (decides whether nodeport is even viable):** from a Pod on cluster B, can you reach `<cluster-A node InternalIP>:<anyNodePort>` directly? Run this as the first thing once cluster B exists:

```bash
# On cluster B, with cluster A's node IP picked from `kubectl get nodes -o wide`
kubectl run net-test --rm -it --image=busybox -- sh
# inside the pod:
nc -vz <cluster-A-node-ip> 22    # or any nodePort already exposed
```

If that times out, nodeport won't work — fall back to loadbalancer. If it connects (or refuses, which still means routable), nodeport is fine.

### 2.2 Cluster B's Postgres question

For Phase 4a MVP: **don't.** Pick a dispatch-test agent type that doesn't need DB. Goal at this stage is "the spawner on cluster B picks up the dispatch, creates the Job, the agent connects to cluster A's Kafka, talks back through the shared topics". That validates the contract. Real DB access via local-PgBouncer-and-tunnel is Phase 5 — the previous Postgres-strategy section covers it.

If we want even more confidence at Phase 4a, a NodePort service in front of `pgbouncer-clients` on cluster A gives us the same "works/doesn't" answer for DB reachability, at zero infra cost. Add this only if the no-DB smoke test exposes ambiguity.

### 2.3 Strimzi on cluster B?

**No.** Cluster B has no Kafka, no KafkaUser, no Kafka topics. Skip `030`, `040`, `045`, `080` entirely in the cluster B deployment. Strimzi operator install alone is hundreds of MB of running pods we don't need.

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

  - name: external             # NEW
    port: 9094
    type: nodeport             # phase 4a — promote to `loadbalancer` later
    tls: true
    authentication:
      type: scram-sha-512
    configuration:
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
- Create one nodePort service for the bootstrap (`personae-kafka-cluster-kafka-external-bootstrap`) on 32094 — load-balanced across the cluster's nodes.
- Create one nodePort service per broker (`...-kafka-external-N`) on the assigned ports.
- Issue a TLS cert covering the bootstrap + each broker's advertised host (which by default is the node external IP). The generated cluster CA is in the secret `personae-kafka-cluster-cluster-ca-cert`.

**Two non-obvious gotchas to flag before applying:**

1. **`useServiceDnsDomain` and advertised host.** With nodeport, each broker advertises a host that the client uses *after* bootstrap. Default = node external IP. On Rackspace, the "InternalIP" is what Strimzi reads — verify the address it picks is reachable from cluster B. If not, set `advertisedHost` per broker to whatever address actually works.

2. **`configuration.bootstrap.host` is for DNS, `nodePort` is for port.** Don't confuse them. We're using nodePort, so we don't need bootstrap.host. We do need consistent nodePorts so cluster B's config doesn't drift each time we re-apply.

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

### 3.3 No other terraform changes on cluster A

Topics, monitoring, Postgres — unchanged. PgBouncer remains in-cluster only at this phase.

---

## 4. Build cluster B

### 4.1 New region key

Pick one. Suggestion: `uk_002` for symmetry with `uk_001`. New paths:

```
deployments/terraform/environments/production/uk002/
├── 010-infrastructure/          # NEW — Rackspace Spot cluster
├── 020-ingress-nginx/           # optional, only if anything on B serves traffic
├── 047-base-configs/            # NEW — for the replicated secrets
└── services/agents/
    └── 2220-remote-job-spawner/ # NEW — calls kustomize-apply against new overlay

deployments/kustomize/services/remote-job-spawner/overlays/production/uk_002/
├── kustomization.yaml
└── configmap.yaml
```

### 4.2 What to deploy on cluster B (and what to skip)

| Module | On B? | Notes |
|---|---|---|
| `010-infrastructure` | YES | New Rackspace Spot cluster. Same module as uk001. Different `terraform.tfvars`. |
| `020-ingress-nginx` | OPTIONAL | Skip unless something on B will serve HTTP traffic externally. For pure dispatch target, skip. |
| `030-strimzi-operator` | NO | Cluster B has no Kafka. |
| `040-kafka-cluster` | NO | Same. |
| `045-kafka-users` | NO | Same. |
| `047-base-configs` | YES | Holds the replicated secrets. |
| `050-storage` | NO unless agents on B need PVCs | Most don't. |
| `060-databases` | NO | No DB on B. |
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

# Apply to cluster B
export KUBECONFIG=~/.kube/config_production_uk002

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

`deployments/kustomize/services/remote-job-spawner/overlays/production/uk_002/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: remote-job-spawner-config
  namespace: ai-persona-system
data:
  # Identity
  CLUSTER_ID: "uk_002"
  NAMESPACE: "ai-persona-system"

  # Kafka — point at cluster A's external listener via nodeport
  # Pick ANY cluster A worker node IP for bootstrap. Strimzi handles the rest.
  KAFKA_BROKERS: "<cluster-A-node-ip>:32094"
  KAFKA_SECURITY_PROTOCOL: "SASL_SSL"
  KAFKA_SASL_MECHANISM: "SCRAM-SHA-512"
  KAFKA_SASL_USERNAME: "personae-app-cross-cluster"
  # password comes from the replicated secret personae-app-cross-cluster

  # Tell the spawner to override what it passes to spawned agents.
  # These env vars are already supported by the spawner code:
  OVERRIDE_KAFKA_BROKERS: "<cluster-A-node-ip>:32094"

  # DB hosts: leave the defaults so the spawner still sets them on the Job.
  # They will be unreachable from cluster B at Phase 4a — that's expected
  # and only matters if a dispatched agent tries to use them.
```

The terraform wrapper at `deployments/terraform/environments/production/uk002/services/agents/2220-remote-job-spawner/main.tf` calls the existing `kustomize-apply` module against this overlay — same pattern as `2210-agent-chassis`.

---

## 5. Sequence of operations (the order matters)

1. **Cluster A — add external Kafka listener** (`040-kafka-cluster` apply). Wait for `kubectl -n kafka get svc | grep external` to show the nodeport services. Verify a kafkacat from a temporary Pod *within cluster A* still works on the internal listener (smoke check we haven't broken existing traffic).

2. **Cluster A — add cross-cluster KafkaUser** (`045-kafka-users` apply). `kubectl -n kafka get secret personae-app-cross-cluster` should appear within a few seconds.

3. **Cluster A — connectivity check from outside.** Easiest: temporarily port-forward, or just run kafkacat from your laptop against the external listener with the credentials. If this works, the listener is correctly configured. If it doesn't, fix here before touching cluster B.

4. **Provision cluster B** — `010-infrastructure` with new tfvars. Get kubeconfig at `~/.kube/config_production_uk002`.

5. **Cluster B — base configs** (`047-base-configs`) plus the secret replication of §4.3.

6. **Cluster B — pre-flight network test** (§2.1). From a Pod on B, confirm reachability to cluster A's nodeport. If this fails, stop and choose loadbalancer.

7. **Cluster B — kafkacat from a Pod** using the replicated `personae-app-cross-cluster` credentials. Confirm it can list topics on cluster A:
   ```bash
   kubectl -n ai-persona-system run kc --rm -it \
     --image=edenhill/kcat:1.7.1 -- \
     -b <cluster-A-node-ip>:32094 \
     -X security.protocol=SASL_SSL \
     -X sasl.mechanism=SCRAM-SHA-512 \
     -X sasl.username=personae-app-cross-cluster \
     -X sasl.password=<from secret> \
     -X ssl.ca.location=<from kafka-ca secret> \
     -L
   ```
   This is the moment to verify Kafka authentication works cross-cluster. If it does, everything else is straightforward.

8. **Cluster B — deploy `remote-job-spawner`** (`deploy-remote-job-spawner ENVIRONMENT=production REGION=uk002 REGION_PATH=uk_002`). Pod should come up, logs should show `"Remote Job Spawner ready, consuming from dispatch topic"`, consumer group should appear on cluster A's Kafka.

9. **Cluster A — dispatch test.** From the orchestration, trigger a `dispatch_agent` with `target_cluster: "uk_002"`. Expected observations:
   - Cluster A spawner logs: skip message (once §6 fix lands).
   - Cluster B spawner logs: "Received dispatch request" → "Successfully created agent job".
   - `kubectl --kubeconfig=~/.kube/config_production_uk002 -n ai-persona-system get jobs` shows the new Job.
   - The spawned Pod on B connects to the external Kafka listener and produces on its `responses_topic`.
   - Cluster A parent receives that response on the shared Kafka and the workflow continues.

If any step fails, the diagnostic order is: §3 (listener) → §4.3 (secrets in right place) → §7 (auth working) → spawner logs.

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

- No Postgres exposure. Phase 5.
- No loadbalancer Kafka listener — nodeport only. Promote when going to production traffic.
- No External Secrets Operator. Manual `kubectl apply` per §4.3 is fine for one cluster.
- No `agent_definitions.target_cluster` column. The dispatch still picks `target_cluster` from workflow config / step config. Per-agent-type routing comes in Phase 3 of the broader plan.
- No dispatch log table. That's Phase 1 (Gap B) and lives on the parent side — independent of whether cluster B exists. Worth doing in parallel.
- No monitoring stack on B. Logs via `kubectl logs` is sufficient at this stage.

---

## 8. Estimated effort

Assuming `010-infrastructure` for Rackspace Spot is a known quantity (it provisioned uk001 successfully), the actual work is small:

| Task | Rough size |
|---|---|
| Modify cluster A's `040-kafka-cluster` to add external listener | 20 lines YAML |
| Add cross-cluster KafkaUser in `045-kafka-users` | 30 lines YAML |
| Provision cluster B (`010-infrastructure` with new tfvars) | tfvars only |
| Cluster B `047-base-configs` and secret replication script | bash script ~30 lines |
| New spawner overlay for uk_002 | 2 short YAML files |
| New tf wrapper for `2220-remote-job-spawner` on uk002 | 1 short main.tf |
| Code change: `logger.Debug` → `logger.Info` in spawner | 4 lines |
| Test workflow that calls `dispatch_agent` with `target_cluster: "uk_002"` | 1 SQL insert |

The "first stage" is the §5 ordering — if step 3 (kafkacat from laptop) works, the rest is mechanical.

---

## 9. First concrete next step

Three commands to settle the unknowns:

```bash
# 1. Where does Strimzi currently say the brokers' external addresses would be?
#    (Tells us what address cluster B will need to dial.)
kubectl -n kafka get nodes -o wide
kubectl -n kafka get pods -l strimzi.io/cluster=personae-kafka-cluster -o wide

# 2. Confirm the existing Kafka resource on cluster A so we can patch it cleanly.
kubectl -n kafka get kafka personae-kafka-cluster -o yaml \
  | grep -A 30 'listeners:'

# 3. Does Rackspace Spot expose a separate "ExternalIP" we should prefer, or
#    is "InternalIP" actually the right address (per the outputs_nodes.tf comment)?
kubectl -n kafka get nodes -o json \
  | jq '.items[].status.addresses'
```

The output from those three commands tells us exactly what `advertisedHost` (if any) to set in §3.1 — and whether nodeport is viable or we go straight to loadbalancer.
