# Multi-Cluster Dispatch — Handoff Summary
## Date: 2026-03-02

## Context

Three sessions (2026-02-27 through 2026-03-02) designed and implemented multi-cluster agent dispatching for the hierarchical agent framework. The goal: allow agents to be spawned on remote Kubernetes clusters while the parent agent remains unaware of the physical location, communicating entirely through Kafka.

## What Was Built

### 1. `DispatchAgentAction` (chassis-side)

**File**: `platform/orchestration/actions/dispatch_actions.go`

Mirrors `SpawnAgentAction` exactly — same helper functions (`extractSpawnConfiguration`, `setupAgentTopics`, `createAgentInDBFromDefinition`, `sendInitializationMessage`, `buildSpawnResult`). The only difference is step 7: instead of calling `spawnAgentKubernetesJobFromDefinition`, it publishes a `DispatchRequest` to `system.dispatch.requests`.

Variable names are identical to `SpawnAgentAction` — `childRequestsTopic`, `parentResponsesTopic`, `stableIdentity`, etc. No changes to existing code.

Configurable per-step via workflow config: `startup_wait_seconds` (default 12s) and `consumer_wait_seconds` (default 8s), longer than local spawn to account for cross-cluster latency.

**Registry entry** needed in `platform/orchestration/actions/registry.go`:
```go
"dispatch_agent": {
Handler:     DispatchAgentAction,
Category:    "agent",
Description: "Dispatch agent creation to a remote cluster via Kafka",
IsLocal:     true,
},
```

### 2. `remote-job-spawner` Service

**File**: `cmd/remote-job-spawner/main.go`

Standalone Go service (was initially named `agent-dispatcher`, renamed to avoid conflict with `build-dispatch-loop`). Consumes from `system.dispatch.requests`, creates K8s Jobs locally. No Postgres dependency — the parent already created DB records.

Key features:
- Filters messages by `target_cluster` header (skips messages for other clusters)
- Three-tier config resolution: env override → dispatch message value → defaults
- Same Job spec structure as `spawnAgentKubernetesJobFromDefinition`
- Labels: `spawned-by: remote-job-spawner`, `spawner-cluster: uk_001`
- Consumer group: `remote-job-spawner-uk_001`
- Sends confirmations to `system.dispatch.responses`

### 3. Deployment Manifests

Full kustomize + terraform structure following the agent-chassis pattern:

```
kustomize/base/                          — deployment.yaml, kustomization.yaml
kustomize/overlays/production/uk_001/    — configmap, rbac, patch
kustomize/overlays/development/          — dev patch
terraform/                               — main.tf, variables.tf, outputs.tf, terraform.tfvars
docker/remote-job-spawner.dockerfile     → build/docker/backend/
```

- RBAC: same Job-creation permissions as agent-chassis (Role: `remote-job-spawner-role`)
- ServiceAccount: `ai-persona-app` (shared with chassis)
- Resources: 50m/64Mi requests, 200m/128Mi limits
- Secrets: references `personae-prod-config`, `personae-default-secrets`, `personae-platform-secrets`
- Terraform module number: 2220

### 4. Kafka Topics Created

```
system.dispatch.requests   — 2 partitions, replication 1
system.dispatch.responses  — 2 partitions, replication 1
```

### 5. Makefile Additions

- `build-remote-job-spawner` target
- Added to `build-agents` prereqs
- Push line added to `push-backend`
- Kustomize apply block in `deploy-agents`
- Added to `update-kustomization-images` loop
- Rollout restart in `redeploy-agents`
- Standalone: `deploy-remote-job-spawner`, `deploy-remote-job-spawner-tf`

## Current Status — Proven Working

The remote-job-spawner is deployed to the primary cluster and **successfully tested**. A kcat dispatch message was sent to `system.dispatch.requests` and the spawner:

1. Consumed the message (filtered by `target_cluster=uk_001`)
2. Created K8s Job `agent-copywriter-f8f34764` in ~640ms
3. Logged success confirmation

Spawner logs showing the working test:
```
{"msg":"Received dispatch request","agent_id":"f8f34764-...","agent_type":"copywriter","target_cluster":"uk_001"}
{"msg":"Created Kubernetes Job","job_name":"agent-copywriter-f8f34764","namespace":"ai-persona-system"}
{"msg":"Successfully created agent job","agent_id":"f8f34764-...","job_name":"agent-copywriter-f8f34764"}
```

### Test Command (kcat pattern)

```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-dispatch-test-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.dispatch.requests \
  -H correlation_id=$CORRELATION_ID \
  -H agent_id=$AGENT_ID \
  -H agent_type=copywriter \
  -H target_cluster=uk_001 \
  -H message_type=dispatch_request <<JSON
{"agent_id":"$AGENT_ID","agent_type":"copywriter","agent_name":"copywriter-dispatch-test","role":"writer","client_id":"default","image_repository":"docker.io/aqls/agent-chassis","image_tag":"v1.0.813","command":null,"resources":{"requests":{"cpu":"100m","memory":"256Mi"},"limits":{"cpu":"500m","memory":"1Gi"}},"health_config":{"port":8080,"liveness_path":"/health","readiness_path":"/ready","initial_delay_seconds":30},"env_vars":[],"category":"data-driven","requests_topic":"job.test-dispatch.requests","responses_topic":"job.test-dispatch.responses","parent_responses_topic":"system.agent.generic.responses","target_cluster":"uk_001","kafka_brokers":"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092","database_host":"pgbouncer.ai-persona-system.svc.cluster.local","database_port":"6432","database_user":"clients_user","database_name":"clients_db","dispatched_at":"$TIMESTAMP"}
JSON
```

Note: Kafka console producer headers have separator conflicts — use kcat instead. The `-H key=value` syntax works cleanly.

## Multi-Cluster Infrastructure Design

### Shared Kafka (current approach, right for single-region Rackspace + one remote)

All clusters connect to cluster A's Kafka via external listener. No local Kafka elsewhere. Requires exposing Kafka externally via Strimzi listener config:

```yaml
listeners:
  - name: plain
    port: 9092
    type: internal
  - name: external
    port: 9094
    type: loadbalancer
    tls: true
    authentication:
      type: scram-sha-512
```

Same for Postgres — expose PgBouncer externally, or run a PgBouncer per remote cluster proxying to cluster A.

### Cluster C (e.g. AWS) — What Changes

When adding a third cluster in a different provider/region:

**Kafka**: Shared Kafka still works if latency is acceptable (50-200ms cross-region adds to message round-trips but agents run for 30s+ so it's negligible for the workload). The remote-job-spawner in cluster C gets `OVERRIDE_KAFKA_BROKERS` pointing to cluster A's external Kafka endpoint.

**Postgres**: Same pattern — agents in cluster C connect to cluster A's Postgres via external endpoint or a local PgBouncer proxy. At scale, read replicas in each region reduce latency for workflow reads.

**What a cluster C agent needs**:
- K8s cluster with `ai-persona-system` namespace
- ServiceAccount `ai-persona-app` with Job create/delete permissions
- Secrets: `personae-platform-secrets`, `personae-default-secrets`, `docker-hub-creds`
- ConfigMap: `personae-prod-config`
- Network access to cluster A's Kafka and Postgres
- `remote-job-spawner` deployment with `CLUSTER_ID=aws_us_east_1` (or similar)
- Agent chassis image accessible (`docker.io/aqls/agent-chassis` already public)

**Agent spawning from cluster C**: A cluster C agent that calls `SpawnAgentAction` (local spawn) creates topics on the shared Kafka, writes DB records to the shared Postgres, and spawns child Jobs locally via the K8s API — all works because the agent's env vars point to the shared infrastructure. The child it spawns in cluster C also gets the same external endpoints. No code changes needed.

**When shared Kafka stops scaling** (100K+ agents or high cross-region latency): Move to Kafka-per-cluster with MirrorMaker 2 replicating `system.dispatch.*` topics. Agent-to-agent topics only need to exist on Kafka reachable by both parent and child — for cross-cluster parent-child pairs this pushes toward federated Kafka or a shared coordination layer. This is a later concern.

### Topology Diagram

```
Cluster A (Rackspace, primary)         Cluster B (Rackspace, overflow)
┌────────────────────────┐            ┌────────────────────────┐
│ agent-chassis (3 pods) │            │ remote-job-spawner     │
│ remote-job-spawner     │            │ spawned agent Jobs     │
│ core-manager           │            │                        │
│ Kafka ◄────────────────┼── ext ─────┼── agents connect here  │
│ Postgres ◄─────────────┼── ext ─────┼── agents connect here  │
│ PgBouncer              │            │ (or local PgBouncer)   │
└────────────────────────┘            └────────────────────────┘
                                      
Cluster C (AWS, future)
┌────────────────────────┐
│ remote-job-spawner     │
│ spawned agent Jobs     │
│ (local PgBouncer       │
│  proxying to A)        │
│                        │
│ connects to A's Kafka  │
│ and Postgres via ext   │
│ endpoints              │
└────────────────────────┘
```

## Files Delivered

All in `/mnt/user-data/outputs/` from the implementation session:

| File | Destination in repo |
|------|-------------------|
| `remote_job_spawner_main.go` | `cmd/remote-job-spawner/main.go` |
| `dispatch_actions.go` | `platform/orchestration/actions/dispatch_actions.go` |
| `remote-job-spawner-deploy.tar.gz` | Kustomize, terraform, dockerfile, makefile additions |
| `FILE_PLACEMENT.txt` | Directory mapping guide |

## Key Design Decisions

1. **Dual-path spawning**: `spawn_agent` (local K8s) and `dispatch_agent` (remote via Kafka) coexist. Workflows choose which to use per step.
2. **Parent doesn't know child is remote**: Same Kafka topic communication, same response format, same workflow engine behaviour.
3. **Spawner is stateless**: Horizontal scaling via Kafka consumer groups. No DB dependency.
4. **No code changes to existing agents**: Spawned agent doesn't know it was dispatched remotely. It reads from its topics and responds.
5. **Variable names kept in sync**: `DispatchAgentAction` and `SpawnAgentAction` use identical variable names for maintainability.
6. **`DispatchRequest` struct defined independently** in both `dispatch_actions.go` and `main.go` — no import dependency between chassis and spawner. Can extract to shared package later.
7. **Rename from agent-dispatcher to remote-job-spawner**: Avoids naming conflict with `build-dispatch-loop`. Kafka topics (`system.dispatch.*`) kept as-is since they describe the protocol.

## What's Not Done Yet

### Immediate next steps
1. **Full workflow test**: Run an actual workflow using `dispatch_agent` instead of `spawn_agent` — validates DB record creation, topic setup, init message, and response routing through the dispatch path end-to-end
2. **Registry patch**: Add `dispatch_agent` to `GlobalActionRegistry` in `registry.go`
3. **Commit**: `dispatch_actions.go`, `remote-job-spawner` binary, deployment manifests

### Cluster B deployment
4. **Expose Kafka externally**: Strimzi listener config change (LoadBalancer on port 9094, TLS + SCRAM-SHA-512)
5. **Expose Postgres externally**: LoadBalancer for PgBouncer, or deploy PgBouncer in cluster B proxying to cluster A
6. **Deploy remote-job-spawner to cluster B**: With `OVERRIDE_KAFKA_BROKERS` and `OVERRIDE_DATABASE_HOST` pointing to cluster A's external endpoints
7. **Copy secrets/configmaps to cluster B**: `personae-platform-secrets`, `personae-default-secrets`, `docker-hub-creds`, `personae-prod-config`
8. **Test cross-cluster dispatch**: Workflow in cluster A dispatches agent to cluster B, agent responds back

### Scaling roadmap (from earlier sessions)
9. **Stress test at 10K agents** with stubbed LLMs on current single cluster
10. **Tree navigation API and UI** for exploring agent hierarchies
11. **Worker pool architecture** for 100K+ scale (shared topic pools instead of per-agent topics)
12. **Self-hosted LLM inference** (vLLM/llama.cpp on GPU nodes)
13. **1M agent demonstration scenario**

## Key Architecture References

- Development guide: `/mnt/project/001e_development_guide_new_agents_v5.md`
- System architecture: `/mnt/project/002c_system_architecture_v3.md`
- Contracts/standards: `/mnt/project/003b_contracts_and_standards_v2.md`
- Build pipeline plan: `/mnt/project/005b_build_expand_plan_v2.md`
- Previous handoff: `/mnt/project/006b_useful_notes_handoff_summary.md`
- Tool library guide: `/mnt/project/010_tool_library_guide.md`
- Kafka scheduler guide: `/mnt/project/011_kafka_scheduler_guide.md`
- Chassis code: `/mnt/project/production_agent-chassis-full_context.txt`
- API adapters: `/mnt/project/production_agent-adapters_context.txt`

## Transcripts

- `2026-02-27-18-19-36-million-agent-scaling-plan.txt` — Strategic planning, use cases, infrastructure cost estimates
- `2026-02-28-20-03-32-multi-cluster-dispatch-design.txt` — Architecture analysis, dispatch design, infrastructure requirements
- `2026-03-02-14-58-40-multi-cluster-dispatch-implementation.txt` — Implementation, deployment, testing, rename to remote-job-spawner