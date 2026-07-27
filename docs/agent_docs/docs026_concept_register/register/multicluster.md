# Register — multicluster

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

15 concepts, consolidated from 42 raw extractions (21 unique blocks, each duplicated
once in the source cluster file) across units U13_docs024_small_dirs,
U17a_docs019_archive_discussions_and_main, U21_legacy_docs_b, U22_recent_small_docs,
U24c_docs_archive_traffic_probe. One concept (Cloudflare-in-front option) was
recognised as a mis-bucketed duplicate of a vm-backend-sites concept and relocated
there (merged with "Cloudflare-proxied-in-front option") — see register/vm-backend-sites.md.

### MCL-001 — Multi-cluster dispatch contract (dispatch_agent action + remote-job-spawner service)
- **status:** partial
- **status-evidence:** Built and ad-hoc tested 2026-03-02 ("successfully tested … Created K8s Job agent-copywriter-f8f34764 in ~640ms"; remote-job-spawner "deployed to the primary cluster and successfully tested"). But the later FOCUS/HANDOFF multi-cluster-dispatch-mvp docs explicitly flag it as still not wired into any real workflow ("No workflow currently uses dispatch_agent") and list the registry patch as outstanding. A still-later doc (PLAN_isolated_chat_environment(5)) describes a live second cluster (va001) running remote-job-spawner and dispatched agents connecting back to primary Kafka/Postgres as an established fact — suggesting the gap was closed by then, but no source explicitly confirms the intermediate gaps (A/B/C) were resolved, so status is kept at partial pending stage-2 verification.
- **what:** A parent-side Go action (`DispatchAgentAction`, `platform/orchestration/actions/dispatch_actions.go`) mirrors `SpawnAgentAction` but publishes a `DispatchRequest` to Kafka topic `system.dispatch.requests` instead of creating a local K8s Job. A separate `remote-job-spawner` service (renamed from agent-dispatcher, stateless, no Postgres dependency) on the target cluster consumes it, filters by `target_cluster`, calls `createAgentJob` (mirroring local spawn), and replies on `system.dispatch.responses`. The spawned agent talks back to its parent over the same shared Kafka via `parent_responses_topic` — no Kafka federation needed, one shared brain. Gives a dual-path spawn model (`spawn_agent` local / `dispatch_agent` remote) chosen per workflow step, parent unaware the child is remote. Provisioning can also target third-party-hosted clusters, with remote DB access via VPN tunnel + local PgBouncer; infrastructure-attributed failures are meant to be quarantined from the trust signal.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§1, multicluster/HANDOFF_multi_cluster_dispatch.md#§1, docs021.../014_multi_cluster_dispatch.md#1-2, docs021.../024_handoff_summary_2026_03_02.md, ED/MASTER_autonomous_build_and_operate(4).md#6.6,#7.5
- **relations:** Cluster-filter gap (MCL-003); Dispatch confirmation observability gap (MCL-004); Same-cluster loopback test requirement (MCL-005); Adjacent-cluster Phase 4a rollout (MCL-002); trust ledger; verification harness; isolated chat/satellite architecture (explicitly NOT reused for chat isolation, see MCL-002)
- **verify-later:** platform/orchestration/actions/dispatch_actions.go; cmd/remote-job-spawner/main.go; registry.go ~lines 95-100 (GlobalActionRegistry dispatch_agent entry); system.dispatch.requests/responses topics; presence of any agent_definitions row / workflow actually using dispatch_agent

### MCL-002 — Adjacent-cluster Phase 4a rollout: va001 second cluster (Rackspace Spot, US-East)
- **status:** aspirational
- **status-evidence:** Earlier FOCUS/HANDOFF docs describe this as entirely aspirational ("Nothing in this plan has been deployed or applied yet. This work happens on a new branch (multi-chassis)… every concrete IP, cluster name, and region key… is illustrative and must be re-discovered"). A later document (PLAN_isolated_chat_environment(5), Appendix) states in present tense that "A second K8s cluster (va001, Rackspace Spot, US-East) runs remote-job-spawner and dispatched agents that connect back to the primary cluster's Kafka … and primary Postgres" — treated here as the more recent, more specific evidence that the rollout completed.
- **stage2-verified (2026-07-14):** deployed → aspirational — va001 / config_production_va001 / va001-data-collector: 0 hits in .go/.yaml/.tf; found only in docs/ prose (FOCUS_adjacent_cluster_phase4a docs). Present-tense plan doc language misread as completion; matches the confirmed false-positive pattern.
- **what:** The concrete first-execution plan and (per later evidence) eventual reality: bring up a second Rackspace Spot cluster (`va001-data-collector`, US-East) as a pure dispatch target for `remote-job-spawner`, sharing cluster A's Kafka and primary Postgres by design ("no federation needed") to borrow trusted compute while keeping one brain. Deliberately identified elsewhere as the WRONG template to reuse for chat isolation, since chat's whole point is removing exactly that shared, synchronous, write-capable channel.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§1.5,§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§5, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#3,Appendix
- **relations:** Multi-cluster dispatch contract (MCL-001); Multi-cluster environment re-discovery handoff practice (MCL-012); RTT-based agent-type viability classification (MCL-010); Isolated chat/satellite architecture (Y-copy, the concept this is explicitly contrasted against)
- **verify-later:** kubeconfig config_production_uk001 / config_production_va001 validity on the multi-chassis branch; system.dispatch.requests/responses topic contract in production

### MCL-003 — Cluster-filter gap in remote-job-spawner (Gap A)
- **status:** partial
- **status-evidence:** Base FOCUS doc lists Gap A as unfixed; a later doc (phase4a(2)) corrects this: "already has the cluster filter — but uses logger.Debug on the skip path… This also corrects the earlier Phase 4 doc which described Gap A as un-fixed."
- **what:** Without a `target_cluster` filter, every deployed spawner would attempt to create the same Job when a second cluster comes online, causing duplicate creation or a K8s name-collision race. The filter guard already exists right after parsing the dispatch request (skips messages not addressed to `clusterID`); the remaining fix is switching the skip-path log from `logger.Debug` to `logger.Info` for observability.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§6, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.1
- **relations:** Multi-cluster dispatch contract (MCL-001); Adjacent-cluster Phase 4a rollout (MCL-002)
- **verify-later:** cmd/remote-job-spawner/main.go consume loop; DEFAULT_CLUSTER_ID env var; skip-path log level

### MCL-004 — Dispatch confirmation observability gap (Gap B / agent_dispatch_log)
- **status:** aspirational
- **status-evidence:** "Fix shape (MVP): persist a row in a new lightweight table… CREATE TABLE IF NOT EXISTS agent_dispatch_log" — proposed DDL only, not built.
- **what:** Today `DispatchAgentAction` fires-and-forgets to Kafka and sleeps ~12s; if the spawner is down, the parent only discovers failure via a generic ~30s+ init-timeout. Proposed fix: a new `agent_dispatch_log` table (agent_id, target_cluster, dispatched_at, confirmed_at, job_name, success, error, spawner_cluster) written on dispatch and updated by a response-consumer goroutine added to core-manager, giving near-real-time failure visibility.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.2, ED/MASTER_autonomous_build_and_operate(4).md#6.6
- **relations:** Multi-cluster dispatch contract (MCL-001); agent_instances (deliberately separate lifecycle)
- **verify-later:** proposed table agent_dispatch_log (existence); core-manager consumer goroutine (not yet written)

### MCL-005 — Same-cluster loopback test requirement (Gap C)
- **status:** aspirational
- **status-evidence:** "No workflow currently uses dispatch_agent… Before adding a second cluster, prove the round-trip works with target_cluster: 'default' against the local spawner."
- **what:** A deliberate pre-condition of the MVP plan: before any cross-cluster network work, dispatch to the same cluster to isolate whether failures originate in the dispatch contract itself versus cross-cluster networking. No such test workflow existed as of this documentation.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5
- **relations:** Multi-cluster dispatch contract (MCL-001)
- **verify-later:** presence of any agent_definitions row / workflow using dispatch_agent

### MCL-006 — Cross-cluster Kafka external listener (Strimzi nodeport→loadbalancer promotion)
- **status:** aspirational
- **status-evidence:** phase4a(2) §5 sequence is entirely prescriptive; HANDOFF: "Nothing in this plan has been deployed or applied yet."
- **what:** Adds a third Strimzi Kafka listener (nodeport, TLS+SCRAM, pinned ports) alongside existing internal listeners, letting a second K8s cluster reach the primary's Kafka without MirrorMaker/federation. Nodeport is the $0 smoke-test choice (brittle); promotion to loadbalancer (~$40/mo) is a one-word YAML change. Rackspace-specific gotcha: `preferredNodePortAddressType: InternalIP` required because Rackspace's "InternalIP" is actually routable public IPv4.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.1,§2.4,§3.1, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4b, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern (MCL-007); Kafka cluster-wide authorization gap (MCL-008); Adjacent-cluster Phase 4a rollout (MCL-002)
- **verify-later:** deployments/terraform kafka-cluster CR template

### MCL-007 — Cross-cluster KafkaUser + secret replication pattern
- **status:** aspirational
- **status-evidence:** Manual `kubectl get | sed | kubectl apply` scripts described as the plan; not yet executed.
- **what:** A new `KafkaUser` (scram-sha-512, ACLs mirroring the anonymous app user) is created on the primary so a remote cluster's agents can authenticate; its generated Secret plus the Kafka cluster CA cert, platform secrets, and default secrets are manually copied into the remote cluster's namespace. Flagged as manual/unautomated — External Secrets Operator is explicitly out of scope for the MVP.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2,§4.3, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4d
- **relations:** Cross-cluster Kafka external listener (MCL-006); Kafka cluster-wide authorization gap (MCL-008)
- **verify-later:** secret names personae-platform-secrets, personae-default-secrets, personae-storage-secrets, personae-app-cross-cluster

### MCL-008 — Kafka cluster-wide authorization gap (ACLs decorative)
- **status:** partial
- **status-evidence:** "The live Kafka cluster has no spec.kafka.authorization block. That means the ACLs declared above… are not actually enforced — they're decorative… everything connects as User:ANONYMOUS with full access."
- **what:** Because the Strimzi Kafka CR has no `authorization` block, KafkaUser ACLs are unenforced — SCRAM-SHA-512 gates the connection but an authenticated user is unrestricted once connected. Deemed acceptable for an internal smoke test but flagged as separate future work (`authorization: simple` cluster-wide).
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern (MCL-007)
- **verify-later:** Kafka CR spec.kafka.authorization (absent); personae-app-anonymous ACL coverage

### MCL-009 — Cross-cluster Postgres reachability strategy (Option C: local PgBouncer + tunnel)
- **status:** aspirational
- **status-evidence:** "Recommendation: Option C… For production." Phased adoption table marks this "Phase 5 — not current MVP."
- **what:** Five options were scored against reachability/auth/failure-mode/pool-sizing (public Postgres LB; local PgBouncer→public primary LB; local PgBouncer→private tunnel→in-cluster Postgres; read replicas; API-driven). The recommended production topology is Option C: a thin local PgBouncer per remote cluster resolving the same in-cluster DNS name as primary, tunnelled back to primary's Postgres — Postgres never publicly exposed, agent env vars identical across clusters, failures localized. No chassis code changes required.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§8, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** RTT-based agent-type viability classification (MCL-010); Cross-cluster Kafka external listener (MCL-006)
- **verify-later:** platform/config DatabaseConfig struct; per-cluster PgBouncer pool_size sizing

### MCL-010 — RTT-based agent-type viability classification for remote dispatch
- **status:** aspirational
- **status-evidence:** phase4a(2) §2.2 gives a projected numbers table explicitly labelled "Numbers above are projections. The smoke test produces actual numbers."
- **stage2-verified (2026-07-14):** partial → aspirational — 0 Go-code hits for viability/RTT classification logic (grep '.go' for viability|RTT found only an unrelated file). The RTT numbers are explicitly labelled projections pending a smoke test that requires the va001 second cluster; per MCL-002 (verified in prior stage-2 batch: 006_VERIFICATION_stage2.md) va001 was never...
- **what:** A cost model classifying which agent types are safe to dispatch cross-cluster based on projected/measured Postgres RTT (UK↔Ashburn ~75-85ms one-way): LLM-bound agents and adapters are "basically free" to dispatch remotely (~1.05x slowdown); composition agents doing many small DB lookups suffer ~25-60x slowdown and are judged not viable without query batching; tight DB-polling inner loops should be avoided entirely.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.2,§5
- **relations:** Cross-cluster Postgres reachability strategy (MCL-009); Adjacent-cluster Phase 4a rollout (MCL-002)
- **verify-later:** actual measured ping/psql timings once the phase4a smoke test is run

### MCL-011 — Cross-cloud cluster expansion (Phase 4: AWS EKS / GCP GKE)
- **status:** aspirational
- **status-evidence:** Entire section framed as design ("Two new directories, one per cloud", "Decision to take before writing") with no implementation claimed.
- **what:** Extends the same-provider adjacent-cluster pattern to a genuinely remote cloud via a new `010-infrastructure` terraform module per cloud matching the existing kubeconfig-output contract. Requires exposing Kafka externally (TLS+SCRAM-SHA-512 loadbalancer, non-negotiable), restricting remote agent types to those that don't need DB for the MVP, replicating three secret bundles, and reusing the `OVERRIDE_KAFKA_BROKERS`/`OVERRIDE_DATABASE_*` env-var pattern with zero chassis code changes.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5,§6
- **relations:** Cross-cluster Kafka external listener (MCL-006); Cross-cluster Postgres reachability strategy (MCL-009)
- **verify-later:** terraform-aws-modules/eks/aws, terraform-google-modules/kubernetes-engine/google usage (not present)

### MCL-012 — Multi-cluster environment re-discovery handoff practice
- **status:** convention
- **status-evidence:** HANDOFF §6 "First moves in the new chat" prescribes re-running discovery rather than trusting the planning-chat's recorded facts.
- **stage2-verified (2026-07-14):** unknown → convention — This is a documented operational discipline (re-derive facts each session), not a testable code artifact. Its cited example facts do check out: pgbouncer Service is actually named 'pgbouncer' (deployments/kustomize/services/pgbouncer/pgbouncer-deployment.yaml:5,67), not 'pgbouncer-clients', and the Kafka cluster is ...
- **what:** A documented discipline for picking up multi-cluster work in a fresh session against a different concrete cluster set: treat all IPs, cluster names, region keys in prior FOCUS docs as illustrative only, and re-derive live facts before proceeding. Includes corrected environment facts already found to differ from plan-time assumptions (e.g. PgBouncer service actually named `pgbouncer`, not `pgbouncer-clients`).
- **sources:** multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§6,§8
- **relations:** Adjacent-cluster Phase 4a rollout (MCL-002)
- **verify-later:** kubectl -n ai-persona-system get svc pgbouncer; kubectl -n kafka get kafka personae-kafka-cluster

### MCL-013 — Multi-cluster scaling tiers (10K/100K/1M agents)
- **status:** aspirational
- **status-evidence:** Explicit phased plan Phase 1-5 with "Current" only at Phase 1 stubbed stress test; per-tier "architectural change" tables in two independently-authored scaling docs (docs016 canine-project baseline and docs021 scaling analysis) that converge on the same roadmap.
- **what:** A scaling analysis mapping each agent-count tier to its primary bottleneck and the single architectural change that unlocks it: 10K = topic-creation churn (no change needed); 50-100K = K8s Job scheduling + Kafka partition count (multi-cluster dispatch + shared topic pools + 8-10 brokers); 1M = per-agent K8s overhead + LLM cost + cross-DC latency (worker pools, regional Kafka/MirrorMaker2, distributed DB, self-hosted GPU inference). Key principle: each jump is one change, agent code never changes. An earlier/parallel telling of the same roadmap (docs016) adds concrete numbers: `system.work.pool.{00-63}` (64 topics × 16 partitions = 1,024 partitions, blast radius ~1,000 co-located agents per bad actor), worker pods running embedded 3B models via llama.cpp + SciSpacy across 5,000-10,000 goroutine workflows, and Redis/Valkey holding hot orchestration state (100K+ writes/sec) with Postgres persisting on completion.
- **sources:** docs021.../015_scaling_analysis.md, docs021.../021_2026-02-27-...-million-agent-scaling-plan.txt, docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#Infrastructure-Design, docs016_dogs_medicine_pathways/002_project_outline.md#Infrastructure
- **relations:** Shared topic pools (MCL-014); Worker pool architecture (MCL-015); remote-job-spawner (MCL-001, cited as "proven working" per docs016/004)
- **verify-later:** current Kafka broker/topic counts; orchestration_states partitioning; Redis/Valkey state layer existence; work-pool topics in Kafka config

### MCL-014 — Shared topic pools (replace per-agent topics)
- **status:** aspirational
- **status-evidence:** "Shared topic pools (needed at this tier) … Route messages by agent ID in headers instead of by topic isolation."
- **what:** A planned change at the 50-100K tier that replaces two dedicated Kafka topics per agent with a fixed set of partitioned pool topics (e.g. `system.agent-work.requests`, 50-100 partitions), routing by agent ID in headers with the chassis filtering. Eliminates topic-creation churn (the 10K-tier ceiling) entirely.
- **sources:** docs021.../015_scaling_analysis.md#50-100k
- **relations:** Worker pool architecture (MCL-015); Multi-cluster scaling tiers (MCL-013)
- **verify-later:** any shared-topic-pool routing in chassis message handling

### MCL-015 — Worker pool architecture (replace per-agent Jobs)
- **status:** aspirational
- **status-evidence:** "Worker pools replace per-agent Jobs … This is the biggest code change in the scaling roadmap."
- **what:** The 1M-tier change: long-running chassis pods pull agent work from shared Kafka pools and run multiple workflows concurrently as goroutines, so scaling is a Deployment replica count instead of Job creation. Reuses the existing orchestration engine; the agent doesn't know whether it's a Job, a dispatched Job, or a goroutine.
- **sources:** docs021.../015_scaling_analysis.md#1m, docs021.../014_multi_cluster_dispatch.md#whats-not-done
- **relations:** Shared topic pools (MCL-014); Multi-cluster scaling tiers (MCL-013)
- **verify-later:** any long-running worker/goroutine-pool mode in chassis
