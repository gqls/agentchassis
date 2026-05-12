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

**Out of scope (later)**

- Kafka MirrorMaker / federation.
- Placement strategy (which cluster gets which agent type). MVP routes by static `target_cluster` in workflow config or agent_definition.
- Cluster-local databases. MVP assumes the spawned agent can reach the primary `clients_db` and `templates_db` via the network (VPN, peering, or public LB with auth). Cross-cluster DB topology is a separate piece of work.
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
