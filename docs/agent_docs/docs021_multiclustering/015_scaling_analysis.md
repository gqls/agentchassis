# Multi-Cluster Scaling Analysis
## Date: 2026-03-02

## Overview

This document covers the scaling characteristics of the hierarchical agent framework across multiple remote clusters, from the current single-cluster setup through to the 1M agent target. Each scale tier has a different primary bottleneck and a specific architectural change that unlocks it.

## Current Architecture

Agents are spawned as K8s Jobs, each getting two dedicated Kafka topics (`job.{identity}.requests` and `.responses`). They communicate through Kafka, store state in Postgres (via PgBouncer), and run the agent-chassis image. The `remote-job-spawner` service (proven working 2026-03-02) enables dispatching agent creation to remote clusters via Kafka.

Infrastructure endpoints are currently in-cluster:
- Kafka: `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`
- Postgres: `pgbouncer.ai-persona-system.svc.cluster.local:6432`
- Agent image: `docker.io/aqls/agent-chassis` (public registry, accessible from any cluster)

## Scale Tiers

### 10K Agents — Current Architecture, No Changes

**Primary constraint**: Kafka topic creation churn.

Each agent creates two topics, so 10K agents means 20K topics with 40K partitions. The Kafka cluster handles this, but topic creation and propagation takes 10+ seconds per topic (observed in production logs). With agents executing over a period of hours and staggered starts, concurrent Pod count peaks at 300-500.

**What works fine at this scale**:
- K8s Job scheduling on a single cluster (300-500 concurrent Pods is comfortable)
- Postgres writes: 10K agents × 5 steps each ≈ ~15 writes/second sustained — trivial for a single instance
- PgBouncer connection pooling: 300-500 concurrent agents is well within pool limits
- Kafka message throughput: 10K agents × ~10 messages each = 100K messages over a few hours — negligible

**Resource requirements**: 3-5 nodes, single Kafka cluster (current setup). No architectural changes needed.

### 50-100K Agents — Multi-Cluster Dispatch

**Primary constraints**: K8s Job scheduling pressure, Kafka topic count.

100K agents on one cluster would overwhelm the K8s API server and etcd with Job creation/deletion churn. 200K topics with 400K partitions pushes Kafka controller limits (roughly 200K partitions per broker before metadata propagation degrades).

**What changes**:

1. **Multi-cluster dispatch** (built and proven): `dispatch_agent` action publishes to `system.dispatch.requests`, `remote-job-spawner` in each cluster creates local K8s Jobs. Each cluster sees 10-20K agents — comfortable territory. Agents don't know they're remote; they communicate through shared Kafka.

2. **Shared topic pools** (needed at this tier): Replace per-agent topics with a fixed set of partitioned topics like `system.agent-work.requests` with 50-100 partitions. Route messages by agent ID in headers instead of by topic isolation. The chassis filters messages by ID. This eliminates topic creation churn entirely.

3. **Kafka scaling**: 8-10 brokers for the partition count and consumer group coordination overhead.

4. **Postgres**: Write throughput becomes a consideration. 100K agents × 5 steps = ~140 writes/second sustained, with bursts of thousands during peak concurrent execution. PgBouncer handles connection pooling. NVMe storage recommended. Consider partitioning `orchestration_states` by date or client_id.

**Resource requirements**: 5-10 K8s clusters, 8-10 Kafka brokers, dedicated Postgres node with fast storage. Shared Kafka still works if all clusters are in the same region (Rackspace).

### 1M Agents — Worker Pools, Regional Kafka, Distributed DB

**Primary constraints**: Per-agent K8s overhead, LLM throughput, cross-region network latency.

At this scale, one Pod per agent is no longer viable. K8s overhead per Job (scheduling, image pull, container startup, etcd writes) dominates actual work time for agents that run 30 seconds to a few minutes.

**What changes**:

1. **Worker pools replace per-agent Jobs**: Long-running chassis Pods pull agent work from shared Kafka topics and execute multiple workflows concurrently as goroutines. The chassis already has the orchestration engine — it just runs multiple workflows per Pod instead of one. Scaling becomes adjusting replica count on a Deployment, not creating Jobs. This is the biggest code change in the scaling roadmap.

2. **Regional Kafka**: Cross-DC latency (50-200ms per message hop) × multiple workflow steps per agent adds up. Regional Kafka clusters with MirrorMaker 2 replicating `system.dispatch.*` topics keeps coordination global while agent-to-agent messaging stays local.

3. **Distributed database**: CockroachDB, Citus, or sharded Postgres. 1M agents writing state updates produces thousands of concurrent writes per second during peak. CockroachDB fits well — multi-region writes without manual sharding, and the query patterns (lookup by correlation_id/orchestration_id) map to its distributed key-value layer.

4. **Self-hosted LLM inference**: Per-token API costs dominate at this scale. Self-hosted 7B models (Mistral, Llama 3, Qwen 2.5) on GPU via vLLM with continuous batching handle 1,000-2,000 requests/minute per A100. For 1M agents each making one LLM call, 5-10 GPUs running for 48 hours covers the inference load.

**Resource requirements**: 50-100 K8s clusters, regional Kafka clusters (federated), distributed DB, GPU inference nodes. Estimated infrastructure cost for a 48-hour million-agent run: $1,000-3,000 on a hybrid GPU inference + CPU worker setup.

## Bottleneck Summary

| Scale | Primary Bottleneck | Secondary | Architectural Change |
|-------|-------------------|-----------|---------------------|
| 10K | Topic creation churn | Nothing significant | None — current architecture works |
| 50K | K8s Job scheduling | Kafka topic count | Multi-cluster dispatch (done), shared topic pools |
| 100K | Kafka partition count | Postgres write throughput | 8-10 brokers, partitioned orchestration_states |
| 1M | Per-agent K8s overhead | LLM API costs, cross-DC latency | Worker pools, regional Kafka, distributed DB, self-hosted LLMs |

## Cost Estimates

### Stress Test at 10K (Stubbed LLMs)

| Item | Cost |
|------|------|
| LLM calls (stubbed for infrastructure test) | $0 |
| Extra Rackspace nodes for burst (few hours) | $10-50 |
| Ongoing monthly infrastructure | Already paying |

### Live Run at 10K (Real LLMs)

| Item | Cost |
|------|------|
| LLM API calls (~500-700 concurrent sites) | $250-700 |
| Infrastructure burst | $10-50 |

### 1M Agent Run (48 hours, self-hosted)

| Approach | Servers | Estimated Cost |
|----------|---------|---------------|
| CPU only (64-core bare metal) | 60-100 servers | $3,000-6,000 |
| GPU only (A100 instances) | 5-10 GPUs | $500-2,000 |
| Hybrid (GPU inference + CPU workers) | 5-8 GPUs + 10-20 CPU nodes | $1,000-3,000 |

The hybrid approach makes the most sense: K8s worker pools, Kafka brokers, and Postgres on CPU nodes; LLM inference on a small GPU cluster.

## Cluster Topology at Scale

```
Single Region (Rackspace)              Multi-Region
┌─────────────────────────┐           ┌─────────────────────────┐
│ Cluster A (primary)     │           │ Region A (London)       │
│  agent-chassis (3 pods) │           │  worker pool (20 pods)  │
│  remote-job-spawner     │           │  remote-job-spawner     │
│  Kafka (shared)         │           │  Kafka (regional)       │
│  Postgres (primary)     │           │  Postgres (primary)     │
│  PgBouncer              │           │  PgBouncer              │
├─────────────────────────┤           ├─────────────────────────┤
│ Cluster B (overflow)    │           │ Region B (Frankfurt)    │
│  remote-job-spawner     │           │  worker pool (20 pods)  │
│  spawned agent Jobs     │           │  remote-job-spawner     │
│  → connects to A's      │           │  Kafka (regional)       │
│    Kafka and Postgres   │           │  PgBouncer → A's PG     │
├─────────────────────────┤           ├─────────────────────────┤
│ Cluster C (AWS, future) │           │ Region C (US-East)      │
│  remote-job-spawner     │           │  worker pool (20 pods)  │
│  spawned agent Jobs     │           │  remote-job-spawner     │
│  → connects to A's      │           │  Kafka (regional)       │
│    Kafka and Postgres   │           │  PgBouncer → A's PG     │
│    via external endpoint│           │  GPU inference cluster  │
└─────────────────────────┘           └─────────────────────────┘

10K-50K agents                        100K-1M agents
Shared Kafka, single Postgres         Federated Kafka (MirrorMaker 2)
Per-agent Jobs                         Worker pools, distributed DB
```

## Implementation Phases

### Phase 1: Stubbed Stress Test at 10K (Current)
- Add `stub_llm` action to chassis (returns canned content after 2-5s delay)
- Push 10K agents through the system
- Find infrastructure ceilings: topic creation, K8s scheduling, PgBouncer connections
- Cost: near zero

### Phase 2: Self-Hosted LLM Validation
- Deploy vLLM or llama.cpp serving a 7B model on 1-2 machines
- Modify LLM adapter to point at local endpoint
- Run 100-500 real agents, validate content quality
- Cost: $50-100 in server rental

### Phase 3: Worker Pool Refactor
- Replace one-Job-per-agent with long-running worker Pods
- Workers pull from shared Kafka topic pools, execute workflows as goroutines
- This is the biggest code change and unlocks everything above 10K
- Test at 10K agents with local LLM

### Phase 4: Multi-Cluster at Scale
- Deploy remote-job-spawner to 2-3 additional clusters
- Expose Kafka and Postgres externally (or run PgBouncer per remote cluster)
- Run at 50-100K agents across clusters
- Implemented infrastructure: proven (remote-job-spawner working 2026-03-02)

### Phase 5: The Million-Agent Run
- Provision GPU inference cluster (5-10 A100s)
- Deploy regional Kafka with MirrorMaker 2
- Deploy CockroachDB or Citus for distributed state
- Run for 48 hours, record everything, tear down
- Budget: $2,000-5,000 all-in for burst infrastructure

## Key Design Principle

Each scale jump requires one architectural change, not a rewrite. The workflow engine, message format, agent definitions, and topic naming conventions stay the same at every scale. What changes is how agents are executed (Jobs → worker pools), how messages are routed (dedicated topics → shared pools), and where infrastructure runs (single cluster → multi-region).

The agent code itself never changes. An agent doesn't know whether it's running as a K8s Job on the primary cluster, a dispatched Job on a remote cluster, or a goroutine in a worker pool. It reads from its topics, executes its workflow, and responds to its parent. That abstraction is what makes the scaling path evolutionary rather than requiring a rewrite at each tier.
