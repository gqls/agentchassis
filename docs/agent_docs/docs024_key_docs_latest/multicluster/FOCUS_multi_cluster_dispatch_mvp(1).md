# FOCUS — Multi-Cluster Agent Dispatch MVP

**Goal:** Get an agent spawned on a *different* Kubernetes cluster, talking back to its parent on the primary cluster, reliably enough to run a real workflow against it. Use what's already built; close the gaps that block end-to-end use.

---

## 1. Snapshot — what exists now

### Built and committed

| Piece | Path | State |
|---|---|---|
| Parent-side action | `platform/orchestration/actions/dispatch_actions.go` | Written. Reuses `extractSpawnConfiguration`, `setupAgentTopics`, `createAgentInDBFromDefinition`, `sendInitializationMessage`, `buildSpawnResult`, `preRegisterAwaitedRequest` — matches the spawn-side helpers per guideline (reuse > recreate). |
| Action registry entry | `platform/orchestration/actions/registry.go` line 95-100 | `dispatch_agent` registered, `IsLocal: true`, category `agent`. |
| Spawner service | `cmd/remote-job-spawner/main.go` | Consumes `system.dispatch.requests`, calls `createAgentJob` (mirrors `spawnAgentKubernetesJobFromDefinition`), produces `DispatchResponse` to `system.dispatch.responses`. |
| Image build | makefile `build-remote-job-spawner` → `build/docker/backend/remote-job-spawner.dockerfile` | Wired into `build-agents` and `push-backend`. |
| Deploy | makefile `deploy-remote-job-spawner` (kustomize) and `deploy-remote-job-spawner-tf` (terraform) | Kustomize path `deployments/kustomize/services/remote-job-spawner/overlays/<overlay>`; terraform path `services/agents/2220-remote-job-spawner`. Need to confirm files actually live at those paths and apply cleanly. |
| Rollout-restart loop | makefile | Includes `kubectl rollout restart deployment remote-job-spawner -n ai-persona-system`. |

### Topic / wire contract

- **Dispatch in:** `system.dispatch.requests` (key = `agent_id`, value = `DispatchRequest` JSON).
- **Dispatch ack:** `system.dispatch.responses` (key = `agent_id`, value = `DispatchResponse` with `success`, `job_name`, `error`).
- **Agent talk-back:** the agent uses `parent_responses_topic` from the dispatch payload — same shared Kafka as the parent. No federation needed.
- **Topic creation:** parent creates `job.<stable_identity>.requests` and `job.<stable_identity>.responses` on the shared cluster *before* dispatching, via `setupAgentTopics`. The spawner does **not** create topics.

### Default config (where to point things)

| Setting | Default | Used by |
|---|---|---|
| `KAFKA_BROKERS` | `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092` | Both sides |
| `NAMESPACE` | `ai-persona-system` | Spawner job creation |
| `CLUSTER_ID` | `uk_001` | Spawner identity / label / consumer group suffix |
| `CONSUMER_GROUP` | `remote-job-spawner-{CLUSTER_ID}` | Spawner |
| `DispatchRequestsTopic` | `system.dispatch.requests` | Both sides (constant in code) |

### Reused (per guidelines — do not recreate)

`extractSpawnConfiguration`, `setupAgentTopics`, `createAgentInDBFromDefinition`, `sendInitializationMessage`, `buildSpawnResult`, `preRegisterAwaitedRequest`, `GenerateAgentName`, `createSubtreeInfo`, `getAgentDefinition`. The spawner's `createAgentJob` is the only K8s-side mirror of `spawnAgentKubernetesJobFromDefinition` — kept separate because the spawner must not import the chassis.

---

## 2. Structural gaps blocking MVP

These are fix-first items (per "structural issues above quick fixes"). Each is small.

### Gap A — Spawner does not filter by `target_cluster`

**File:** `cmd/remote-job-spawner/main.go`, in the consume loop.

Today every spawner consumes every `system.dispatch.requests` message regardless of `req.TargetCluster`. With one spawner running it works; the moment a second cluster's spawner comes up, **both** will create the Job (or race and both fail on the K8s name collision check, depending on timing).

**Fix shape:** right after parsing `req`, before `createAgentJob`:

```go
// Skip messages not addressed to this cluster.
// "default" is treated as the primary; map it via env if needed.
target := req.TargetCluster
if target == "" || target == "default" {
    target = getEnvOrDefault("DEFAULT_CLUSTER_ID", "uk_001")
}
if target != clusterID {
    logger.Info("Skipping dispatch for other cluster",
        zap.String("target_cluster", req.TargetCluster),
        zap.String("my_cluster_id", clusterID),
        zap.String("agent_id", req.AgentID))
    continue
}
```

Single-cluster behaviour unchanged (default cluster id resolves to its own id). Multi-cluster: each spawner picks up only its own messages. **No variable-name changes required.**

### Gap B — Parent does not observe dispatch confirmation

`DispatchAgentAction` fires-and-forgets to `system.dispatch.requests` and sleeps `DefaultRemoteStartupWait = 12s`. If the spawner is down, malfunctioning, or has bad RBAC, the parent finds out only when `sendInitializationMessage` times out — typically 30s+ later, and the failure mode is generic "init didn't respond" rather than "dispatch failed at the spawner".

**Fix shape (MVP — minimum viable):** persist a row in a new lightweight table when dispatching, and add a small consumer (in the existing core-manager or as a goroutine in the agent-chassis startup) that subscribes to `system.dispatch.responses` and updates that row. Don't block the workflow on it — just make the failure mode observable in the dashboard.

```sql
CREATE TABLE IF NOT EXISTS agent_dispatch_log (
    agent_id        UUID PRIMARY KEY,
    target_cluster  TEXT NOT NULL,
    agent_type      TEXT NOT NULL,
    correlation_id  TEXT,
    orchestration_id UUID,
    dispatched_at   TIMESTAMPTZ NOT NULL,
    confirmed_at    TIMESTAMPTZ,
    job_name        TEXT,
    success         BOOLEAN,
    error           TEXT,
    spawner_cluster TEXT  -- which spawner answered (helps diagnose Gap A in production)
);

CREATE INDEX idx_agent_dispatch_log_unconfirmed
    ON agent_dispatch_log (dispatched_at)
    WHERE confirmed_at IS NULL;
```

`DispatchAgentAction` inserts the row after producing. A response-consumer updates `confirmed_at`, `job_name`, `success`, `error`, `spawner_cluster`.

**Why a separate table not `agent_instances`:** `agent_instances` is per-client/per-spawn; the dispatch log is operational telemetry. Different lifecycle, different read patterns. Don't conflate.

**Where the consumer lives:** add a goroutine to the core-manager service (already listens on system topics, already has DB access). Don't create a new service for this.

### Gap C — No same-cluster loopback test exists

No workflow currently uses `dispatch_agent`. Before adding a second cluster, prove the round-trip works with `target_cluster: "default"` against the local spawner. That isolates whether failures are in the dispatch contract or the cross-cluster networking.

---

## 3. Things to verify before claiming the kustomize is real

The makefile references these paths. Confirm they exist and that what's in them matches the code:

```
deployments/kustomize/services/remote-job-spawner/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml          # 1 replica per cluster, env vars per §1
│   ├── rbac.yaml                # Role: batch/jobs CRUD, pods get/list
│   └── serviceaccount.yaml
└── overlays/
    └── production/uk_001/
        ├── kustomization.yaml   # newTag, namespace, configmap patch
        └── configmap.yaml       # CLUSTER_ID, OVERRIDE_* if needed

deployments/terraform/services/agents/2220-remote-job-spawner/
└── (analogous; whichever path is canonical for the environment)
```

The chassis already has `rbac-job-spawner.yaml` granting `batch/jobs` CRUD to `ai-persona-app` SA in `ai-persona-system`. The spawner can reuse that SA, or have its own — either works as long as the binding is in the target cluster's namespace.

**Verification commands** (run against the primary cluster first):

```bash
kubectl -n ai-persona-system get deploy remote-job-spawner
kubectl -n ai-persona-system get pods -l app=remote-job-spawner
kubectl -n ai-persona-system logs deploy/remote-job-spawner --tail=100
```

If `remote-job-spawner` isn't running yet, that's the first thing to fix — it's the cheap end of the work.

---

## 4. MVP scope

**In scope**

1. Spawner deployed on the primary cluster and processing `system.dispatch.requests`.
2. Gap A fix (cluster filter) merged so the design supports >1 spawner without surprises.
3. Gap B fix (dispatch log + response consumer) so failures are visible within seconds, not via timeout.
4. A loopback test workflow that dispatches an agent with `target_cluster: "default"` (resolved to `uk_001`), receives the init response, and completes.
5. A second cluster brought up (or simulated by running a second spawner with a different `CLUSTER_ID` and dummy K8s API) and the same test passes with `target_cluster: "<second_id>"`.

**Out of scope for MVP, but tracked in Phase 4**

- Kafka MirrorMaker / federation — Phase 4 instead uses a Strimzi external listener (single Kafka, TLS+SCRAM, externally reachable). Mirroring becomes relevant only if we want cluster-local Kafka per region for resilience or data-residency reasons. Phase 5+.
- Placement strategy (which cluster gets which agent type). MVP routes by static `target_cluster` in workflow config or agent_definition.
- Cluster-local databases. MVP assumes the spawned agent can reach the primary `clients_db` and `templates_db` via the network. Phase 4 specifically restricts the remote cluster to agent types that don't need DB so we can validate the dispatch contract without exposing Postgres publicly. Exposing DB is Phase 5.
- Dispatch-time fuel budgets / quotas per cluster.

---

## 5. Phased plan

### Phase 0 — Same-cluster loopback (1–2 sessions)

1. Verify `remote-job-spawner` kustomize files are present and apply cleanly to `uk_001`. Fix if not.
2. Deploy spawner. Check pod is `Running`, logs show "Remote Job Spawner starting" with the right cluster id.
3. Apply **Gap A** patch — add the `target_cluster` filter and a `DEFAULT_CLUSTER_ID` env var (default `uk_001`).
4. Write a thin loopback workflow in SQL (a one-step agent_definitions entry, or a manual orchestration trigger). Use `dispatch_agent` with `target_cluster: "default"`, `agent_type: <something cheap, e.g. "briefing-agent" or a test agent>`.
5. Run it. Confirm:
   - Spawner log: "Received dispatch request" → "Successfully created agent job".
   - K8s: `kubectl -n ai-persona-system get jobs -l spawned-by=remote-job-spawner` shows the job.
   - Parent orchestration completes (or the agent's normal failure mode kicks in — we just want it to *reach* the agent).

### Phase 1 — Observability and dispatch log (1 session)

1. Apply the `agent_dispatch_log` migration to `clients_db` (check schema first per guideline — confirm the DB and that we want it there vs templates_db; I'd say clients_db since dispatch is per-client work).
2. `DispatchAgentAction`: insert a row right after the `Producer.Produce` call. Fire-and-forget on failure — don't block.
3. Add a `DispatchResponseConsumer` goroutine in core-manager (or, if we'd rather keep core-manager untouched, in the agent-chassis startup — but core-manager is more natural since it already runs system-level consumers).
4. Smoke test: dispatch one agent, query `SELECT * FROM agent_dispatch_log WHERE agent_id = ...` — should show `confirmed_at` populated within a few seconds.

### Phase 2 — Second cluster (1–2 sessions, mostly ops)

1. Decide where the second cluster lives and how it reaches the shared Kafka. Network path is the hard part, not the code.
2. Replicate the secrets the spawner-created Job references:
   - `personae-platform-secrets` (`CLIENTS_DB_PASSWORD`, `TEMPLATES_DB_PASSWORD`, `AUTH_DB_PASSWORD`, `agent-bootstrap-key`)
   - `personae-default-secrets` (`ANTHROPIC_API_KEY`)
   - `personae-storage-secrets` (`B2_APPLICATION_KEY_ID`, `B2_APPLICATION_KEY`)
3. Deploy `remote-job-spawner` to the second cluster with its own `CLUSTER_ID` (e.g. `uk_002`).
4. Confirm both spawners are alive and the new spawner's logs show it skipping messages for other clusters (Gap A working as intended).
5. Run the same loopback test with `target_cluster: "uk_002"`. The Job should appear on cluster 2, the agent should talk back through the shared Kafka, the parent on cluster 1 should receive the response.

### Phase 3 — Production usage (later)

- Pick a real agent type that benefits from running elsewhere (image-heavy, GPU-adjacent, regional data residency, whatever the driver is).
- Add an optional `target_cluster` column to `agent_definitions` (`text NULL`) — when set, the chassis prefers `dispatch_agent` over `spawn_agent` for that type. **Do not change** the existing `spawn_agent` behaviour or any current variable names.
- Wire admin dashboard read-only view of `agent_dispatch_log`.

### Phase 4 — Cross-cloud cluster (AWS EKS or GCP GKE) talking to primary

This is where Phase 2's "same Kafka, different K8s cluster" stops working unmodified. The primary cluster's Kafka and Postgres are reachable only via in-cluster DNS (`*.svc.cluster.local`). A second cluster on AWS or GCP has no network path to those names. Three things have to be true before a remote-cluster agent can do useful work:

1. It can reach Kafka and produce/consume on shared topics.
2. It can reach `clients_db` and `templates_db` (or we restrict which agent types run remotely to ones that don't need DB — limited but possible for an MVP smoke test).
3. The cluster has the same secrets and image-pull access as primary.

The terraform/kustomize structure already has the shape needed for this — the layered modules (`010-infrastructure` → … → `100-bootstrap-agents`) are environment+region-keyed at `deployments/terraform/environments/<env>/<region>/`. Adding a cloud means a new region key and a swapped `010-infrastructure` module. The rest (Strimzi-on-K8s, Postgres-on-K8s, the spawner, the bootstrap agents) is cloud-agnostic.

#### 4a. New `010-infrastructure` modules

Two new directories, one per cloud:

```
deployments/terraform/environments/production/aws_eu_west_2/010-infrastructure/
deployments/terraform/environments/production/gcp_europe_west2/010-infrastructure/
```

These do **not** call the Hetzner/Kind module — they call their own:

- **AWS:** `terraform-aws-modules/eks/aws` (EKS) + VPC module. Output a kubeconfig the same way the existing module does so `deploy-infrastructure-from-ingress` continues to work for everything downstream of step 010.
- **GCP:** `terraform-google-modules/kubernetes-engine/google` (GKE). Same kubeconfig output contract.

**Variable contract to match the existing modules** (so `020`-`100` keep working unchanged):

| Var/output | Why | Existing module produces |
|---|---|---|
| `cluster_name` | Used by labels and Strimzi listener certs | yes |
| `kubeconfig_raw` (output) | Written to `~/.kube/config_<env>_<region>` by makefile | yes |
| `pod_cidr` / `service_cidr` | Needed if we want to firewall Kafka by source | yes |

Anything beyond that (node pools, IAM, OIDC) is cloud-specific and lives inside the module.

**Decision to take before writing:** do the new clusters run their own Strimzi+Postgres or rely entirely on primary's? Recommended for MVP: **run them lean** — no Strimzi, no Postgres, no monitoring stack on the remote cluster. Skip terraform modules `030-strimzi-operator`, `040-kafka-cluster`, `045-kafka-users`, `060-databases`, `065-pgbouncer`, `070-database-schemas`, `080-kafka-topics`, `090-monitoring` on the remote. The remote only needs `010`, `020-ingress` (optional, only if it serves anything), `047-base-configs` (for the replicated secrets), and `100-bootstrap-agents` reduced to just `remote-job-spawner`.

#### 4b. Expose Kafka externally (the actual hard part)

Strimzi supports this natively — add a second listener to the existing `Kafka` resource on the primary cluster:

```yaml
# In deployments/terraform/environments/production/uk001/040-kafka-cluster
listeners:
  - name: plain        # existing internal listener — leave it alone
    port: 9092
    type: internal
    tls: false
  - name: external     # NEW
    port: 9094
    type: loadbalancer
    tls: true
    authentication:
      type: scram-sha-512
    configuration:
      bootstrap:
        annotations:
          # Cloud-specific LB annotations; if Hetzner, set the LB type here
      brokers:
        - broker: 0
          advertisedHost: kafka-broker-0.<your-domain>
        - broker: 1
          advertisedHost: kafka-broker-1.<your-domain>
        - broker: 2
          advertisedHost: kafka-broker-2.<your-domain>
```

DNS records for `kafka-broker-{0,1,2}.<your-domain>` point at the per-broker LoadBalancer IPs Strimzi creates. **TLS+SCRAM-SHA-512** is non-negotiable for an external listener — there is no scenario where this should be plaintext on the public internet.

**What needs to change in the code:** the spawner sends `KafkaBrokers` to the agent it spawns. Today this is the in-cluster bootstrap. For remote spawns it must be the external bootstrap. Two ways to handle this without breaking primary behaviour:

1. **Per-cluster spawner config (simpler):** The remote spawner has `OVERRIDE_KAFKA_BROKERS=kafka-bootstrap.<your-domain>:9094` plus TLS env vars. Existing code already supports `OVERRIDE_*` taking precedence over what's in the dispatch message. Primary keeps using the in-cluster bootstrap; remote uses the external one. No chassis code changes.
2. **Per-target_cluster config in primary (later):** Look up the target's Kafka endpoint from an `agent_clusters` table. Cleaner but more work — defer.

For Phase 4 MVP, do option 1.

#### 4c. Postgres reachability

Two paths, ranked by simplicity:

1. **Don't.** Pick a remote-friendly agent type that doesn't need DB (e.g. a pure web-scrape adapter, or an image-generator). The spawner's env vars include DB hosts but if the agent never opens a DB connection it doesn't matter. This gets a working cross-cloud smoke test fastest.
2. **Expose Postgres via a TLS-terminating LB or a tunnel.** Options:
   - PgBouncer with TLS on an external LoadBalancer + IP allow-list per remote cluster's NAT egress.
   - A Tailscale/WireGuard sidecar pattern — primary runs a Tailscale endpoint, remote agents resolve via Tailscale DNS. Adds a daemon to each spawned Pod (sidecar) which complicates the chassis. Not recommended.
   - Cloud-native: AWS PrivateLink to a primary-side endpoint (only works if primary is also on AWS). GCP equivalent. Not portable across clouds.

For Phase 4 MVP, pick path 1 — restrict the remote cluster to agent types that don't need DB. Treat real cross-cloud DB as Phase 5.

#### 4d. Secret replication

The spawner's `createAgentJob` references three secrets that must already exist in the remote cluster's `ai-persona-system` namespace. Don't try to be clever — copy them:

```bash
# Pull from primary (run with primary kubeconfig)
for secret in personae-platform-secrets personae-default-secrets personae-storage-secrets; do
  kubectl -n ai-persona-system get secret $secret -o yaml \
    | grep -v '^\s*resourceVersion:' \
    | grep -v '^\s*uid:' \
    | grep -v '^\s*creationTimestamp:' \
    | grep -v '^\s*namespace:' \
    > /tmp/$secret.yaml
done

# Add the external Kafka credentials secret to the bundle
# (created by Strimzi as part of 045-kafka-users when you add the external user)
kubectl -n ai-persona-system get secret kafka-external-user -o yaml > /tmp/kafka-external-user.yaml

# Apply to remote (run with remote kubeconfig)
export KUBECONFIG=~/.kube/config_production_aws_eu_west_2
kubectl -n ai-persona-system apply -f /tmp/personae-platform-secrets.yaml
kubectl -n ai-persona-system apply -f /tmp/personae-default-secrets.yaml
kubectl -n ai-persona-system apply -f /tmp/personae-storage-secrets.yaml
kubectl -n ai-persona-system apply -f /tmp/kafka-external-user.yaml
```

Long-term, automate this with External Secrets Operator pointing at the same backing store (AWS Secrets Manager / GCP Secret Manager / Vault). Out of scope for the MVP — manual copy is fine for the first cluster.

#### 4e. Image pull

`docker.io/aqls/agent-chassis` is public. No registry pull-secret needed for the chassis image. If any agent image becomes private, this becomes a Phase 5 problem.

#### 4f. Spawner deployment on the remote

Same `deployments/kustomize/services/remote-job-spawner/` base. New overlay:

```
deployments/kustomize/services/remote-job-spawner/overlays/production/aws_eu_west_2/
├── kustomization.yaml
└── configmap.yaml   # CLUSTER_ID=aws_eu_west_2, OVERRIDE_KAFKA_BROKERS=..., OVERRIDE_DATABASE_HOST=... (or unset if DB not exposed)
```

And a terraform wrapper at `deployments/terraform/environments/production/aws_eu_west_2/services/agents/2220-remote-job-spawner/main.tf` that just calls the existing `kustomize-apply` module against the new overlay. This is the pattern the codebase already uses for `2210-agent-chassis` — reuse, don't recreate.

#### 4g. Makefile additions

`REGION_PATH` already drives `OVERLAY_PATH`. So:

```bash
make deploy-remote-job-spawner ENVIRONMENT=production REGION=aws_eu_west_2 REGION_PATH=aws_eu_west_2
```

should work once the overlay exists. No makefile changes strictly needed — that's why the existing structure is worth following.

#### 4h. End-to-end test for Phase 4

1. Primary already has the Phase 0–2 setup running.
2. Provision AWS EKS via the new `010-infrastructure` module. Verify nodes are ready.
3. Replicate secrets (4d) and apply the external Kafka listener on primary (4b).
4. Deploy `remote-job-spawner` to AWS cluster with `CLUSTER_ID=aws_eu_west_2`.
5. From admin dashboard / a manual orchestration, dispatch a simple agent (one that doesn't need DB) with `target_cluster: "aws_eu_west_2"`.
6. Observe:
   - Primary spawner logs: "Skipping dispatch for other cluster" with `my_cluster_id=uk_001`.
   - AWS spawner logs: "Received dispatch request" → "Successfully created agent job".
   - `kubectl --kubeconfig=~/.kube/config_production_aws_eu_west_2 -n ai-persona-system get jobs` shows the agent Job.
   - Agent Pod logs on AWS show it connecting to `kafka-bootstrap.<your-domain>:9094` (the external listener) over TLS+SCRAM.
   - Parent orchestration on primary receives the init response on the shared Kafka.
   - `agent_dispatch_log` shows `confirmed_at` populated with `spawner_cluster=aws_eu_west_2`.

If any of those fail, the diagnostic order is: spawner log → Kafka TLS handshake → DNS resolution → IAM/firewall → secret contents. Start at the top.

#### 4i. Cost and operational notes (worth knowing before the apply)

- An EKS control plane is ~$72/month even idle; GKE Autopilot is similar. Don't leave Phase 4 clusters running between sessions unless they're doing real work.
- The 3 broker LoadBalancers on primary are real cloud LB resources too. If primary is on Hetzner, that's ~€6/mo each. If primary is on a cloud LB pricier than that, costs add up.
- Each new cluster doubles the operational surface. Phase 4 is a real commitment, not just terraform-apply-and-forget.

---

## 6. Tests / acceptance

| Check | How |
|---|---|
| Spawner pod healthy | `kubectl -n ai-persona-system get pods -l app=remote-job-spawner` shows 1/1 Running, restarts <3. |
| Dispatch request consumed | Spawner log within 5s of dispatch contains "Received dispatch request" with the right `agent_id`. |
| Job created | `kubectl -n ai-persona-system get job agent-<type>-<id-prefix>` exists, owner labels include `spawned-by=remote-job-spawner` and `spawner-cluster=<cluster_id>`. |
| Cross-cluster filter works | Deploy a second spawner with different `CLUSTER_ID`. Dispatch with `target_cluster: "uk_001"`. The `uk_002` spawner log contains "Skipping dispatch for other cluster"; only `uk_001`'s spawner creates the Job. |
| Confirmation logged | `agent_dispatch_log.confirmed_at IS NOT NULL` within 10s of dispatch. |
| Failure surfaces | Stop the spawner. Dispatch. After 30s, `agent_dispatch_log` row still has `confirmed_at IS NULL` (visible in dashboard query) — does **not** require waiting for the orchestration timeout to know. |
| Agent talks back | Spawned agent's `responses_topic` produces messages consumed by the parent on the shared Kafka, regardless of which cluster the agent runs in. |
| **Phase 4 — TLS Kafka reachable** | From a temporary Pod on the AWS/GCP cluster: `kafkacat -b kafka-bootstrap.<your-domain>:9094 -X security.protocol=SASL_SSL -X sasl.mechanism=SCRAM-SHA-512 -X sasl.username=... -L` lists topics within 5s. |
| **Phase 4 — Cross-cloud dispatch** | Dispatch with `target_cluster: "aws_eu_west_2"`. Only that cluster's spawner creates a Job. Spawned agent connects to the external Kafka listener and replies on its `responses_topic`. Parent on primary receives the init response. `agent_dispatch_log.spawner_cluster = "aws_eu_west_2"`. |

---

## 7. Don't-forget list (operational)

- Spawner uses `IsLocal: true` in the registry — correct, the **dispatch action** runs locally on the parent. The agent being spawned runs remotely. Don't change this label.
- Kafka cluster names: `personae-kafka-cluster-*` in namespace `kafka`. The chassis namespace stays `ai-persona-system`.
- The spawner's `createAgentJob` already maps `spawner-cluster: <clusterID>` as a Job label. Useful for `kubectl get jobs -A -l spawner-cluster=uk_002`.
- The DB row in `agent_instances` is written by the **parent** in the primary `clients_db` before dispatch. The remote agent reads/writes via the same DB — so the remote cluster needs DB connectivity. If we ever want true cluster-local DBs, that's a much bigger change (separate plan).
- Don't use `logger.Debug` anywhere new — won't show up. Use `logger.Info` / `logger.Warn` / `logger.Error`.
- Every agent is an orchestrator — `dispatch_agent` doesn't change that; it just relocates where the Job runs. The spawned agent's workflow drives the same way it would locally.
- Don't put cross-cluster logic in workflow SQL. Spawn a sub-agent (via dispatch) with its own workflow if you need a clean log boundary.

---

## 8. First concrete next step

Run this on the primary cluster and paste the output — it tells us whether Phase 0 starts at step 1 (write kustomize) or step 4 (loopback test):

```bash
kubectl -n ai-persona-system get deploy remote-job-spawner -o wide 2>&1
kubectl -n ai-persona-system get pods -l app=remote-job-spawner 2>&1
ls -la deployments/kustomize/services/remote-job-spawner/ 2>&1
```
