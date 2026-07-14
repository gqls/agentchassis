# Cluster: multicluster-adapters-deploy
Categories included: multicluster, adapters, new:vm-backend-sites, storage-architecture, deployment-github, new:backend-service-deployment, new:persistent-service-deployment, site-snapshots-and-revert


<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster dispatch (Phase 4a) — the coupled model NOT reused for chat isolation
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** "A second K8s cluster (va001, Rackspace Spot, US-East) runs remote-job-spawner and dispatched agents that connect back to the primary cluster's Kafka ... and primary Postgres." (PLAN_isolated_chat_environment(5).md, Appendix)
- **what:** The platform already runs a second Kubernetes cluster whose remote-job-spawner/dispatch_agent mechanism deliberately shares cluster A's Kafka and primary Postgres ("no federation needed" by design), to borrow compute for trusted internal build work while keeping one brain. The isolated-chat-environment plan explicitly identifies this as the wrong template for chat isolation, since chat's whole point is removing exactly that shared, synchronous, write-capable channel.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#3,Appendix
- **relations:** Multi-cluster agent dispatch contract (see multicluster/ concepts below, the actual dispatch mechanism); Isolated chat/satellite architecture (Y-copy)
- **verify-later:** `system.dispatch.requests`/`system.dispatch.responses` topic contract

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster agent dispatch contract (dispatch_agent / remote-job-spawner)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** HANDOFF: "The parent-side action and the remote spawner service are already written and committed on the main codebase. Nothing in this plan has been deployed or applied yet."
- **what:** A parent-side Go action (`DispatchAgentAction`, `platform/orchestration/actions/dispatch_actions.go`) publishes a `DispatchRequest` to Kafka topic `system.dispatch.requests` instead of creating a local K8s Job; a separate `remote-job-spawner` service on the target cluster consumes it, calls `createAgentJob` (mirroring `spawnAgentKubernetesJobFromDefinition`), and replies on `system.dispatch.responses`. The spawned agent talks back to its parent over the same shared Kafka via `parent_responses_topic` — no Kafka federation needed. Registered as action `dispatch_agent` (`IsLocal: true`, category `agent`) in `registry.go`.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§1, multicluster/HANDOFF_multi_cluster_dispatch.md#§1, multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§1
- **relations:** Cluster-filter gap (Gap A); Dispatch confirmation observability gap (Gap B); Multi-cluster dispatch (Phase 4a) — the coupled model NOT reused for chat isolation
- **verify-later:** platform/orchestration/actions/dispatch_actions.go; cmd/remote-job-spawner/main.go; registry.go ~lines 95-100

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cluster-filter gap in remote-job-spawner (Gap A)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** Base FOCUS doc lists Gap A as unfixed; phase4a(2) corrects this: "already has the cluster filter — but uses logger.Debug on the skip path... This also corrects the earlier Phase 4 doc which described Gap A as un-fixed."
- **what:** Without a `target_cluster` filter, every deployed spawner would attempt to create the same Job when a second cluster comes online, causing duplicate creation or a K8s name-collision race. The fix is a guard right after parsing the dispatch request that skips messages not addressed to `clusterID`, and switching the skip-path log from `logger.Debug` to `logger.Info`.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§6, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.1
- **relations:** Multi-cluster agent dispatch contract; Adjacent-cluster Phase 4a execution plan
- **verify-later:** cmd/remote-job-spawner/main.go consume loop, DEFAULT_CLUSTER_ID env var

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dispatch confirmation observability gap (Gap B / agent_dispatch_log)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Fix shape (MVP): persist a row in a new lightweight table... CREATE TABLE IF NOT EXISTS agent_dispatch_log" — proposed DDL only, not built
- **what:** Today `DispatchAgentAction` fires-and-forgets to Kafka and sleeps ~12s; if the spawner is down, the parent only discovers failure via a generic ~30s+ init-timeout. Proposed fix: a new `agent_dispatch_log` table (agent_id, target_cluster, dispatched_at, confirmed_at, job_name, success, error, spawner_cluster) written on dispatch and updated by a response-consumer goroutine added to core-manager, giving near-real-time failure visibility.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.2
- **relations:** Multi-cluster agent dispatch contract; agent_instances (deliberately separate lifecycle)
- **verify-later:** proposed table agent_dispatch_log; core-manager consumer goroutine (not yet written)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Same-cluster loopback test requirement (Gap C)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "No workflow currently uses dispatch_agent... Before adding a second cluster, prove the round-trip works with target_cluster: 'default' against the local spawner."
- **what:** A deliberate pre-condition of the MVP plan: before any cross-cluster network work, dispatch to the same cluster to isolate whether failures originate in the dispatch contract itself versus cross-cluster networking. No such test workflow exists yet.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5
- **relations:** Multi-cluster agent dispatch contract
- **verify-later:** presence of any agent_definitions row / workflow using dispatch_agent

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster Kafka external listener (Strimzi nodeport→loadbalancer promotion)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** phase4a(2) §5 sequence is entirely prescriptive; HANDOFF: "Nothing in this plan has been deployed or applied yet."
- **what:** Adds a third Strimzi Kafka listener (nodeport, TLS+SCRAM, pinned ports) alongside existing internal listeners, letting a second K8s cluster reach the primary's Kafka without MirrorMaker/federation. Nodeport is the $0 smoke-test choice (brittle); promotion to loadbalancer (~$40/mo) is a one-word YAML change. Rackspace-specific gotcha: `preferredNodePortAddressType: InternalIP` required because Rackspace's "InternalIP" is actually routable public IPv4.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.1,§2.4,§3.1, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4b, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern; Kafka authorization gap; Adjacent-cluster Phase 4a execution plan
- **verify-later:** deployments/terraform kafka-cluster CR template

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster KafkaUser + secret replication pattern
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Manual `kubectl get | sed | kubectl apply` scripts described as the plan; not yet executed
- **what:** A new `KafkaUser` (scram-sha-512, ACLs mirroring the anonymous app user) is created on the primary so a remote cluster's agents can authenticate; its generated Secret plus the Kafka cluster CA cert, platform secrets, and default secrets are manually copied into the remote cluster's namespace. Flagged as manual/unautomated — External Secrets Operator is explicitly out of scope for the MVP.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2,§4.3, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4d
- **relations:** Cross-cluster Kafka external listener; Kafka authorization gap
- **verify-later:** secret names personae-platform-secrets, personae-default-secrets, personae-storage-secrets, personae-app-cross-cluster

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Kafka cluster-wide authorization gap (ACLs decorative)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** "The live Kafka cluster has no spec.kafka.authorization block. That means the ACLs declared above... are not actually enforced — they're decorative... everything connects as User:ANONYMOUS with full access."
- **what:** Because the Strimzi Kafka CR has no `authorization` block, KafkaUser ACLs are unenforced — SCRAM-SHA-512 gates the connection but an authenticated user is unrestricted once connected. Deemed acceptable for an internal smoke test but flagged as separate future work (`authorization: simple` cluster-wide).
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern
- **verify-later:** Kafka CR spec.kafka.authorization (absent); personae-app-anonymous ACL coverage

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster Postgres reachability strategy (Option C: local PgBouncer + tunnel)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Recommendation: Option C... For production." Phased adoption table marks this "Phase 5 — not current MVP."
- **what:** Five options (public Postgres LB; local PgBouncer→public primary LB; local PgBouncer→private tunnel→in-cluster Postgres; read replicas; API-driven) scored against reachability/auth/failure-mode/pool-sizing. Recommended production topology is Option C: a thin local PgBouncer per remote cluster resolving the same in-cluster DNS name as primary, tunnelled back to primary's Postgres — Postgres never publicly exposed, agent env vars identical across clusters, failures localized. No chassis code changes required.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§8, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** RTT-based agent-type viability classification; Cross-cluster Kafka external listener
- **verify-later:** platform/config DatabaseConfig struct; per-cluster PgBouncer pool_size sizing

<!-- SOURCE: U13_docs024_small_dirs.md -->
### RTT-based agent-type viability classification for remote dispatch
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** phase4a(2) §2.2 gives a projected numbers table explicitly labelled "Numbers above are projections. The smoke test produces actual numbers"
- **what:** A cost model classifying which agent types are safe to dispatch cross-cluster based on projected/measured Postgres RTT (UK↔Ashburn ~75-85ms one-way): LLM-bound agents and adapters are "basically free" to dispatch remotely (~1.05x slowdown); composition agents doing many small DB lookups suffer ~25-60x slowdown and are judged not viable without query batching; tight DB-polling inner loops should be avoided entirely.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.2,§5
- **relations:** Cross-cluster Postgres reachability strategy; Adjacent-cluster Phase 4a execution plan
- **verify-later:** actual measured ping/psql timings once phase4a smoke test is run

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cloud cluster expansion (Phase 4: AWS EKS / GCP GKE)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Entire section framed as design ("Two new directories, one per cloud", "Decision to take before writing") with no implementation claimed
- **what:** Extends the same-provider adjacent-cluster pattern to a genuinely remote cloud via a new `010-infrastructure` terraform module per cloud matching the existing kubeconfig-output contract. Requires exposing Kafka externally (TLS+SCRAM-SHA-512 loadbalancer, non-negotiable), restricting remote agent types to those that don't need DB for the MVP, replicating three secret bundles, and reusing the `OVERRIDE_KAFKA_BROKERS`/`OVERRIDE_DATABASE_*` env-var pattern with zero chassis code changes.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5,§6
- **relations:** Cross-cluster Kafka external listener; Cross-cluster Postgres reachability strategy
- **verify-later:** terraform-aws-modules/eks/aws, terraform-google-modules/kubernetes-engine/google usage (not present)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Adjacent-cluster Phase 4a execution plan (Rackspace Spot, va001)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** HANDOFF: "Nothing in this plan has been deployed or applied yet. This work happens on a new branch (multi-chassis) against a different set of clusters... every concrete IP, cluster name, and region key in the FOCUS docs is illustrative and must be re-discovered."
- **what:** The concrete first-execution plan: bring up (or reuse) a second Rackspace Spot cluster (`va001-data-collector`, US-East) as a pure dispatch target for `remote-job-spawner`, expose cluster A's Kafka (nodeport) and PgBouncer (nodeport, RTT measurement only) externally, and run two load-bearing checkpoints before the end-to-end dispatch/timing test. Superseded by the HANDOFF's caveat that this exact environment is now stale relative to the `multi-chassis` branch's actual target clusters.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§1.5,§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§5
- **relations:** Multi-cluster environment re-discovery handoff practice; RTT-based agent-type viability classification
- **verify-later:** kubeconfig config_production_uk001 / config_production_va001 validity on the multi-chassis branch

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster environment re-discovery handoff practice
- **category:** multicluster
- **status-signal:** unknown
- **status-evidence:** HANDOFF §6 "First moves in the new chat" prescribes re-running discovery rather than trusting the planning-chat's recorded facts
- **what:** A documented discipline for picking up multi-cluster work in a fresh session against a different concrete cluster set: treat all IPs, cluster names, region keys in prior FOCUS docs as illustrative only, and re-derive live facts before proceeding. Includes corrected environment facts already found to differ from plan-time assumptions (e.g. PgBouncer service actually named `pgbouncer`, not `pgbouncer-clients`).
- **sources:** multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§6,§8
- **relations:** Adjacent-cluster Phase 4a execution plan
- **verify-later:** kubectl -n ai-persona-system get svc pgbouncer; kubectl -n kafka get kafka personae-kafka-cluster

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Cross-cluster / multi-cluster dispatch (design reference within MASTER)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.6 "already designed and partly built (FOCUS_multi_cluster_dispatch_mvp): a dispatch_agent action + remote-job-spawner consuming system.dispatch.requests"
- **what:** Provisioning can target a third-party-hosted Kubernetes cluster with a `dispatch_agent` action and `remote-job-spawner`; remote agents reply on the same Kafka via `parent_responses_topic`; remote DB access uses a VPN tunnel + local PgBouncer at the same in-cluster DNS name. Forces one refinement: infrastructure-attributed failures are quarantined from the trust signal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.6, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** references FOCUS_multi_cluster_dispatch_mvp; trust ledger; verification harness
- **verify-later:** dispatch_agent action; remote-job-spawner; system.dispatch.requests; agent_dispatch_log

<!-- SOURCE: U21_legacy_docs_b.md -->
### Shared Kafka topic pools + worker pools for 1M agents
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** docs016/003c: "system.work.pool.{00-63} — 64 topics × 16 partitions = 1,024 partitions... Blast radius of a single bad agent is limited to ~1,000 co-located agents"; worker pod spec (embedded 3B via llama.cpp, SciSpacy, 5,000-10,000 goroutine workflows).
- **what:** Scaling architecture replacing per-agent topics with hashed shared pools carrying target_agent_id headers; long-running worker pods execute thousands of agent workflows as goroutines with local small models, routing bigger calls to shared inference servers; Redis/Valkey holds hot orchestration state (100K+ writes/sec) with Postgres persisting on completion; multi-cluster worker fleets via remote-job-spawner.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#Infrastructure-Design; docs016_dogs_medicine_pathways/002_project_outline.md#Infrastructure
- **relations:** multicluster docs021 (remote-job-spawner "proven working" per docs016/004); scheduler-and-tasks; canine project.
- **verify-later:** work pool topics in Kafka config; Redis state layer existence.

<!-- SOURCE: U22_recent_small_docs.md -->
### DispatchAgentAction (remote dispatch via Kafka)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** "successfully tested ... Created K8s Job agent-copywriter-f8f34764 in ~640ms"; but "Registry patch: Add dispatch_agent to GlobalActionRegistry" and full-workflow test listed as not done.
- **what:** An action mirroring `SpawnAgentAction` exactly (identical helpers/variable names) except step 7 publishes a `DispatchRequest` to `system.dispatch.requests` instead of creating a local K8s Job. Gives a dual-path spawn model — `spawn_agent` (local) and `dispatch_agent` (remote) — chosen per workflow step, with the parent unaware the child is remote. Longer startup/consumer waits for cross-cluster latency.
- **sources:** docs021.../014_multi_cluster_dispatch.md#1, docs021.../024_handoff_summary_2026_03_02.md
- **relations:** remote-job-spawner, SpawnAgentAction, vertical research/build cluster separation
- **verify-later:** actions/dispatch_actions.go; GlobalActionRegistry dispatch_agent entry

<!-- SOURCE: U22_recent_small_docs.md -->
### remote-job-spawner service
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** "The remote-job-spawner is deployed to the primary cluster and successfully tested" with logged Job creation (2026-03-02).
- **what:** A standalone stateless Go service (renamed from agent-dispatcher) consuming `system.dispatch.requests`, filtering by `target_cluster` header, and creating local K8s Jobs with the same spec as local spawn — no Postgres dependency (parent already wrote DB records). Confirms to `system.dispatch.responses`; scales horizontally via consumer groups; deployed per remote cluster with `CLUSTER_ID`.
- **sources:** docs021.../014_multi_cluster_dispatch.md#2, docs021.../015_scaling_analysis.md
- **relations:** DispatchAgentAction, system.dispatch.* topics, isolated chat environment (explicitly NOT reused for chat)
- **verify-later:** cmd/remote-job-spawner/main.go; system.dispatch.requests/responses topics; va001 cluster deployment

<!-- SOURCE: U22_recent_small_docs.md -->
### Multi-cluster scaling tiers (10K/100K/1M)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Explicit phased plan Phase 1-5 with "Current" only at Phase 1 stubbed stress test; per-tier "architectural change" tables.
- **what:** A scaling analysis mapping each agent-count tier to its primary bottleneck and the single architectural change that unlocks it: 10K = topic-creation churn (no change); 50-100K = K8s Job scheduling + Kafka partition count (multi-cluster dispatch + shared topic pools + 8-10 brokers); 1M = per-agent K8s overhead + LLM cost + cross-DC latency (worker pools, regional Kafka/MirrorMaker2, distributed DB, self-hosted GPU inference). Key principle: each jump is one change, agent code never changes.
- **sources:** docs021.../015_scaling_analysis.md, docs021.../021_2026-02-27-...-million-agent-scaling-plan.txt
- **relations:** shared topic pools, worker pools, self-hosted LLM inference
- **verify-later:** current Kafka broker/topic counts; orchestration_states partitioning

<!-- SOURCE: U22_recent_small_docs.md -->
### Shared topic pools (replace per-agent topics)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Shared topic pools (needed at this tier) ... Route messages by agent ID in headers instead of by topic isolation."
- **what:** A planned change at the 50-100K tier that replaces two dedicated Kafka topics per agent with a fixed set of partitioned pool topics (e.g. `system.agent-work.requests`, 50-100 partitions), routing by agent ID in headers with the chassis filtering. Eliminates topic-creation churn (the 10K-tier ceiling) entirely.
- **sources:** docs021.../015_scaling_analysis.md#50-100k
- **relations:** worker pools, multi-cluster scaling tiers
- **verify-later:** any shared-topic-pool routing in chassis message handling

<!-- SOURCE: U22_recent_small_docs.md -->
### Worker pool architecture (replace per-agent Jobs)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Worker pools replace per-agent Jobs ... This is the biggest code change in the scaling roadmap."
- **what:** The 1M-tier change: long-running chassis pods pull agent work from shared Kafka pools and run multiple workflows concurrently as goroutines, so scaling is a Deployment replica count instead of Job creation. Reuses the existing orchestration engine; the agent doesn't know whether it's a Job, a dispatched Job, or a goroutine.
- **sources:** docs021.../015_scaling_analysis.md#1m, docs021.../014_multi_cluster_dispatch.md#whats-not-done
- **relations:** shared topic pools, self-hosted LLM inference
- **verify-later:** any long-running worker/goroutine-pool mode in chassis

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Cloudflare-in-front option
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13(f) "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h, 4,416 attacks blocked)"; runbook(12) §8.
- **what:** Optional proxied (orange-cloud) Cloudflare record → VM origin (a reverse proxy, NOT a second Worker/copy). Adjustments: cache-bypass the API paths; set nginx `set_real_ip_from`/`real_ip_header CF-Connecting-IP` (else rate-limit throttles all CF IPs as one, and logs/digest show CF IPs); TLS Full(strict); bonus CF-IPCountry + instant relocation. setup.sh `CLOUDFLARE=true` writes cloudflare-realip.conf.
- **sources:** traffic_probe_runbook(12).md#8, traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** required for real client IPs in /access-digest + Thread-D blocklist
- **verify-later:** /etc/nginx/conf.d/cloudflare-realip.conf; setup.sh CLOUDFLARE branch

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster dispatch (Phase 4a) — the coupled model NOT reused for chat isolation
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** "A second K8s cluster (va001, Rackspace Spot, US-East) runs remote-job-spawner and dispatched agents that connect back to the primary cluster's Kafka ... and primary Postgres." (PLAN_isolated_chat_environment(5).md, Appendix)
- **what:** The platform already runs a second Kubernetes cluster whose remote-job-spawner/dispatch_agent mechanism deliberately shares cluster A's Kafka and primary Postgres ("no federation needed" by design), to borrow compute for trusted internal build work while keeping one brain. The isolated-chat-environment plan explicitly identifies this as the wrong template for chat isolation, since chat's whole point is removing exactly that shared, synchronous, write-capable channel.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#3,Appendix
- **relations:** Multi-cluster agent dispatch contract (see multicluster/ concepts below, the actual dispatch mechanism); Isolated chat/satellite architecture (Y-copy)
- **verify-later:** `system.dispatch.requests`/`system.dispatch.responses` topic contract

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster agent dispatch contract (dispatch_agent / remote-job-spawner)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** HANDOFF: "The parent-side action and the remote spawner service are already written and committed on the main codebase. Nothing in this plan has been deployed or applied yet."
- **what:** A parent-side Go action (`DispatchAgentAction`, `platform/orchestration/actions/dispatch_actions.go`) publishes a `DispatchRequest` to Kafka topic `system.dispatch.requests` instead of creating a local K8s Job; a separate `remote-job-spawner` service on the target cluster consumes it, calls `createAgentJob` (mirroring `spawnAgentKubernetesJobFromDefinition`), and replies on `system.dispatch.responses`. The spawned agent talks back to its parent over the same shared Kafka via `parent_responses_topic` — no Kafka federation needed. Registered as action `dispatch_agent` (`IsLocal: true`, category `agent`) in `registry.go`.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§1, multicluster/HANDOFF_multi_cluster_dispatch.md#§1, multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§1
- **relations:** Cluster-filter gap (Gap A); Dispatch confirmation observability gap (Gap B); Multi-cluster dispatch (Phase 4a) — the coupled model NOT reused for chat isolation
- **verify-later:** platform/orchestration/actions/dispatch_actions.go; cmd/remote-job-spawner/main.go; registry.go ~lines 95-100

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cluster-filter gap in remote-job-spawner (Gap A)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** Base FOCUS doc lists Gap A as unfixed; phase4a(2) corrects this: "already has the cluster filter — but uses logger.Debug on the skip path... This also corrects the earlier Phase 4 doc which described Gap A as un-fixed."
- **what:** Without a `target_cluster` filter, every deployed spawner would attempt to create the same Job when a second cluster comes online, causing duplicate creation or a K8s name-collision race. The fix is a guard right after parsing the dispatch request that skips messages not addressed to `clusterID`, and switching the skip-path log from `logger.Debug` to `logger.Info`.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§6, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.1
- **relations:** Multi-cluster agent dispatch contract; Adjacent-cluster Phase 4a execution plan
- **verify-later:** cmd/remote-job-spawner/main.go consume loop, DEFAULT_CLUSTER_ID env var

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dispatch confirmation observability gap (Gap B / agent_dispatch_log)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Fix shape (MVP): persist a row in a new lightweight table... CREATE TABLE IF NOT EXISTS agent_dispatch_log" — proposed DDL only, not built
- **what:** Today `DispatchAgentAction` fires-and-forgets to Kafka and sleeps ~12s; if the spawner is down, the parent only discovers failure via a generic ~30s+ init-timeout. Proposed fix: a new `agent_dispatch_log` table (agent_id, target_cluster, dispatched_at, confirmed_at, job_name, success, error, spawner_cluster) written on dispatch and updated by a response-consumer goroutine added to core-manager, giving near-real-time failure visibility.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§2.2
- **relations:** Multi-cluster agent dispatch contract; agent_instances (deliberately separate lifecycle)
- **verify-later:** proposed table agent_dispatch_log; core-manager consumer goroutine (not yet written)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Same-cluster loopback test requirement (Gap C)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "No workflow currently uses dispatch_agent... Before adding a second cluster, prove the round-trip works with target_cluster: 'default' against the local spawner."
- **what:** A deliberate pre-condition of the MVP plan: before any cross-cluster network work, dispatch to the same cluster to isolate whether failures originate in the dispatch contract itself versus cross-cluster networking. No such test workflow exists yet.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp.md#§2, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5
- **relations:** Multi-cluster agent dispatch contract
- **verify-later:** presence of any agent_definitions row / workflow using dispatch_agent

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster Kafka external listener (Strimzi nodeport→loadbalancer promotion)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** phase4a(2) §5 sequence is entirely prescriptive; HANDOFF: "Nothing in this plan has been deployed or applied yet."
- **what:** Adds a third Strimzi Kafka listener (nodeport, TLS+SCRAM, pinned ports) alongside existing internal listeners, letting a second K8s cluster reach the primary's Kafka without MirrorMaker/federation. Nodeport is the $0 smoke-test choice (brittle); promotion to loadbalancer (~$40/mo) is a one-word YAML change. Rackspace-specific gotcha: `preferredNodePortAddressType: InternalIP` required because Rackspace's "InternalIP" is actually routable public IPv4.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.1,§2.4,§3.1, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4b, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern; Kafka authorization gap; Adjacent-cluster Phase 4a execution plan
- **verify-later:** deployments/terraform kafka-cluster CR template

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster KafkaUser + secret replication pattern
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Manual `kubectl get | sed | kubectl apply` scripts described as the plan; not yet executed
- **what:** A new `KafkaUser` (scram-sha-512, ACLs mirroring the anonymous app user) is created on the primary so a remote cluster's agents can authenticate; its generated Secret plus the Kafka cluster CA cert, platform secrets, and default secrets are manually copied into the remote cluster's namespace. Flagged as manual/unautomated — External Secrets Operator is explicitly out of scope for the MVP.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2,§4.3, multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§4d
- **relations:** Cross-cluster Kafka external listener; Kafka authorization gap
- **verify-later:** secret names personae-platform-secrets, personae-default-secrets, personae-storage-secrets, personae-app-cross-cluster

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Kafka cluster-wide authorization gap (ACLs decorative)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** "The live Kafka cluster has no spec.kafka.authorization block. That means the ACLs declared above... are not actually enforced — they're decorative... everything connects as User:ANONYMOUS with full access."
- **what:** Because the Strimzi Kafka CR has no `authorization` block, KafkaUser ACLs are unenforced — SCRAM-SHA-512 gates the connection but an authenticated user is unrestricted once connected. Deemed acceptable for an internal smoke test but flagged as separate future work (`authorization: simple` cluster-wide).
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§3.2, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** Cross-cluster KafkaUser + secret replication pattern
- **verify-later:** Kafka CR spec.kafka.authorization (absent); personae-app-anonymous ACL coverage

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cluster Postgres reachability strategy (Option C: local PgBouncer + tunnel)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Recommendation: Option C... For production." Phased adoption table marks this "Phase 5 — not current MVP."
- **what:** Five options (public Postgres LB; local PgBouncer→public primary LB; local PgBouncer→private tunnel→in-cluster Postgres; read replicas; API-driven) scored against reachability/auth/failure-mode/pool-sizing. Recommended production topology is Option C: a thin local PgBouncer per remote cluster resolving the same in-cluster DNS name as primary, tunnelled back to primary's Postgres — Postgres never publicly exposed, agent env vars identical across clusters, failures localized. No chassis code changes required.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§8, multicluster/HANDOFF_multi_cluster_dispatch.md#§3
- **relations:** RTT-based agent-type viability classification; Cross-cluster Kafka external listener
- **verify-later:** platform/config DatabaseConfig struct; per-cluster PgBouncer pool_size sizing

<!-- SOURCE: U13_docs024_small_dirs.md -->
### RTT-based agent-type viability classification for remote dispatch
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** phase4a(2) §2.2 gives a projected numbers table explicitly labelled "Numbers above are projections. The smoke test produces actual numbers"
- **what:** A cost model classifying which agent types are safe to dispatch cross-cluster based on projected/measured Postgres RTT (UK↔Ashburn ~75-85ms one-way): LLM-bound agents and adapters are "basically free" to dispatch remotely (~1.05x slowdown); composition agents doing many small DB lookups suffer ~25-60x slowdown and are judged not viable without query batching; tight DB-polling inner loops should be avoided entirely.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§2.2,§5
- **relations:** Cross-cluster Postgres reachability strategy; Adjacent-cluster Phase 4a execution plan
- **verify-later:** actual measured ping/psql timings once phase4a smoke test is run

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Cross-cloud cluster expansion (Phase 4: AWS EKS / GCP GKE)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Entire section framed as design ("Two new directories, one per cloud", "Decision to take before writing") with no implementation claimed
- **what:** Extends the same-provider adjacent-cluster pattern to a genuinely remote cloud via a new `010-infrastructure` terraform module per cloud matching the existing kubeconfig-output contract. Requires exposing Kafka externally (TLS+SCRAM-SHA-512 loadbalancer, non-negotiable), restricting remote agent types to those that don't need DB for the MVP, replicating three secret bundles, and reusing the `OVERRIDE_KAFKA_BROKERS`/`OVERRIDE_DATABASE_*` env-var pattern with zero chassis code changes.
- **sources:** multicluster/FOCUS_multi_cluster_dispatch_mvp(3).md#§5,§6
- **relations:** Cross-cluster Kafka external listener; Cross-cluster Postgres reachability strategy
- **verify-later:** terraform-aws-modules/eks/aws, terraform-google-modules/kubernetes-engine/google usage (not present)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Adjacent-cluster Phase 4a execution plan (Rackspace Spot, va001)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** HANDOFF: "Nothing in this plan has been deployed or applied yet. This work happens on a new branch (multi-chassis) against a different set of clusters... every concrete IP, cluster name, and region key in the FOCUS docs is illustrative and must be re-discovered."
- **what:** The concrete first-execution plan: bring up (or reuse) a second Rackspace Spot cluster (`va001-data-collector`, US-East) as a pure dispatch target for `remote-job-spawner`, expose cluster A's Kafka (nodeport) and PgBouncer (nodeport, RTT measurement only) externally, and run two load-bearing checkpoints before the end-to-end dispatch/timing test. Superseded by the HANDOFF's caveat that this exact environment is now stale relative to the `multi-chassis` branch's actual target clusters.
- **sources:** multicluster/FOCUS_adjacent_cluster_phase4a(2).md#§1.5,§5, multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§5
- **relations:** Multi-cluster environment re-discovery handoff practice; RTT-based agent-type viability classification
- **verify-later:** kubeconfig config_production_uk001 / config_production_va001 validity on the multi-chassis branch

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Multi-cluster environment re-discovery handoff practice
- **category:** multicluster
- **status-signal:** unknown
- **status-evidence:** HANDOFF §6 "First moves in the new chat" prescribes re-running discovery rather than trusting the planning-chat's recorded facts
- **what:** A documented discipline for picking up multi-cluster work in a fresh session against a different concrete cluster set: treat all IPs, cluster names, region keys in prior FOCUS docs as illustrative only, and re-derive live facts before proceeding. Includes corrected environment facts already found to differ from plan-time assumptions (e.g. PgBouncer service actually named `pgbouncer`, not `pgbouncer-clients`).
- **sources:** multicluster/HANDOFF_multi_cluster_dispatch.md#§4,§6,§8
- **relations:** Adjacent-cluster Phase 4a execution plan
- **verify-later:** kubectl -n ai-persona-system get svc pgbouncer; kubectl -n kafka get kafka personae-kafka-cluster

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Cross-cluster / multi-cluster dispatch (design reference within MASTER)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.6 "already designed and partly built (FOCUS_multi_cluster_dispatch_mvp): a dispatch_agent action + remote-job-spawner consuming system.dispatch.requests"
- **what:** Provisioning can target a third-party-hosted Kubernetes cluster with a `dispatch_agent` action and `remote-job-spawner`; remote agents reply on the same Kafka via `parent_responses_topic`; remote DB access uses a VPN tunnel + local PgBouncer at the same in-cluster DNS name. Forces one refinement: infrastructure-attributed failures are quarantined from the trust signal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.6, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** references FOCUS_multi_cluster_dispatch_mvp; trust ledger; verification harness
- **verify-later:** dispatch_agent action; remote-job-spawner; system.dispatch.requests; agent_dispatch_log

<!-- SOURCE: U21_legacy_docs_b.md -->
### Shared Kafka topic pools + worker pools for 1M agents
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** docs016/003c: "system.work.pool.{00-63} — 64 topics × 16 partitions = 1,024 partitions... Blast radius of a single bad agent is limited to ~1,000 co-located agents"; worker pod spec (embedded 3B via llama.cpp, SciSpacy, 5,000-10,000 goroutine workflows).
- **what:** Scaling architecture replacing per-agent topics with hashed shared pools carrying target_agent_id headers; long-running worker pods execute thousands of agent workflows as goroutines with local small models, routing bigger calls to shared inference servers; Redis/Valkey holds hot orchestration state (100K+ writes/sec) with Postgres persisting on completion; multi-cluster worker fleets via remote-job-spawner.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#Infrastructure-Design; docs016_dogs_medicine_pathways/002_project_outline.md#Infrastructure
- **relations:** multicluster docs021 (remote-job-spawner "proven working" per docs016/004); scheduler-and-tasks; canine project.
- **verify-later:** work pool topics in Kafka config; Redis state layer existence.

<!-- SOURCE: U22_recent_small_docs.md -->
### DispatchAgentAction (remote dispatch via Kafka)
- **category:** multicluster
- **status-signal:** partial
- **status-evidence:** "successfully tested ... Created K8s Job agent-copywriter-f8f34764 in ~640ms"; but "Registry patch: Add dispatch_agent to GlobalActionRegistry" and full-workflow test listed as not done.
- **what:** An action mirroring `SpawnAgentAction` exactly (identical helpers/variable names) except step 7 publishes a `DispatchRequest` to `system.dispatch.requests` instead of creating a local K8s Job. Gives a dual-path spawn model — `spawn_agent` (local) and `dispatch_agent` (remote) — chosen per workflow step, with the parent unaware the child is remote. Longer startup/consumer waits for cross-cluster latency.
- **sources:** docs021.../014_multi_cluster_dispatch.md#1, docs021.../024_handoff_summary_2026_03_02.md
- **relations:** remote-job-spawner, SpawnAgentAction, vertical research/build cluster separation
- **verify-later:** actions/dispatch_actions.go; GlobalActionRegistry dispatch_agent entry

<!-- SOURCE: U22_recent_small_docs.md -->
### remote-job-spawner service
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** "The remote-job-spawner is deployed to the primary cluster and successfully tested" with logged Job creation (2026-03-02).
- **what:** A standalone stateless Go service (renamed from agent-dispatcher) consuming `system.dispatch.requests`, filtering by `target_cluster` header, and creating local K8s Jobs with the same spec as local spawn — no Postgres dependency (parent already wrote DB records). Confirms to `system.dispatch.responses`; scales horizontally via consumer groups; deployed per remote cluster with `CLUSTER_ID`.
- **sources:** docs021.../014_multi_cluster_dispatch.md#2, docs021.../015_scaling_analysis.md
- **relations:** DispatchAgentAction, system.dispatch.* topics, isolated chat environment (explicitly NOT reused for chat)
- **verify-later:** cmd/remote-job-spawner/main.go; system.dispatch.requests/responses topics; va001 cluster deployment

<!-- SOURCE: U22_recent_small_docs.md -->
### Multi-cluster scaling tiers (10K/100K/1M)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** Explicit phased plan Phase 1-5 with "Current" only at Phase 1 stubbed stress test; per-tier "architectural change" tables.
- **what:** A scaling analysis mapping each agent-count tier to its primary bottleneck and the single architectural change that unlocks it: 10K = topic-creation churn (no change); 50-100K = K8s Job scheduling + Kafka partition count (multi-cluster dispatch + shared topic pools + 8-10 brokers); 1M = per-agent K8s overhead + LLM cost + cross-DC latency (worker pools, regional Kafka/MirrorMaker2, distributed DB, self-hosted GPU inference). Key principle: each jump is one change, agent code never changes.
- **sources:** docs021.../015_scaling_analysis.md, docs021.../021_2026-02-27-...-million-agent-scaling-plan.txt
- **relations:** shared topic pools, worker pools, self-hosted LLM inference
- **verify-later:** current Kafka broker/topic counts; orchestration_states partitioning

<!-- SOURCE: U22_recent_small_docs.md -->
### Shared topic pools (replace per-agent topics)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Shared topic pools (needed at this tier) ... Route messages by agent ID in headers instead of by topic isolation."
- **what:** A planned change at the 50-100K tier that replaces two dedicated Kafka topics per agent with a fixed set of partitioned pool topics (e.g. `system.agent-work.requests`, 50-100 partitions), routing by agent ID in headers with the chassis filtering. Eliminates topic-creation churn (the 10K-tier ceiling) entirely.
- **sources:** docs021.../015_scaling_analysis.md#50-100k
- **relations:** worker pools, multi-cluster scaling tiers
- **verify-later:** any shared-topic-pool routing in chassis message handling

<!-- SOURCE: U22_recent_small_docs.md -->
### Worker pool architecture (replace per-agent Jobs)
- **category:** multicluster
- **status-signal:** aspirational
- **status-evidence:** "Worker pools replace per-agent Jobs ... This is the biggest code change in the scaling roadmap."
- **what:** The 1M-tier change: long-running chassis pods pull agent work from shared Kafka pools and run multiple workflows concurrently as goroutines, so scaling is a Deployment replica count instead of Job creation. Reuses the existing orchestration engine; the agent doesn't know whether it's a Job, a dispatched Job, or a goroutine.
- **sources:** docs021.../015_scaling_analysis.md#1m, docs021.../014_multi_cluster_dispatch.md#whats-not-done
- **relations:** shared topic pools, self-hosted LLM inference
- **verify-later:** any long-running worker/goroutine-pool mode in chassis

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Cloudflare-in-front option
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13(f) "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h, 4,416 attacks blocked)"; runbook(12) §8.
- **what:** Optional proxied (orange-cloud) Cloudflare record → VM origin (a reverse proxy, NOT a second Worker/copy). Adjustments: cache-bypass the API paths; set nginx `set_real_ip_from`/`real_ip_header CF-Connecting-IP` (else rate-limit throttles all CF IPs as one, and logs/digest show CF IPs); TLS Full(strict); bonus CF-IPCountry + instant relocation. setup.sh `CLOUDFLARE=true` writes cloudflare-realip.conf.
- **sources:** traffic_probe_runbook(12).md#8, traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** required for real client IPs in /access-digest + Thread-D blocklist
- **verify-later:** /etc/nginx/conf.d/cloudflare-realip.conf; setup.sh CLOUDFLARE branch

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adapter/response message envelope contract (normative)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 035 §1 "last verified against code 2026-06-11"; 003(8) now points here
- **what:** Any reply to a chassis request must be recognised as an awaited response or it silently falls to process-as-work (row stuck waiting, ~10-min retries, no error). Load-bearing field: in_response_to_request_id = incoming request_id (request_id fallback only; git adapter's reuse pattern favoured). Three header tiers (validator-enforced / coordinator-needed-but-unvalidated / observability). Body headers MUST be a typed struct with real bools (map[string]string string-bools fail unmarshal and drop the reply pre-claim — the multi-day thunder outage). Send via ProduceWithValidation. Request parsing: action from body, payload at body.data, accept reply-topic from three keys. Sibling race: local dispatches must preRegisterAwaitedRequest before send (confirmed fixed in prod 06-09).
- **sources:** 035 §1; 016 §9 bool-trap + race entries
- **relations:** awaited_requests; O(K²) batch presign
- **verify-later:** ValidateOutgoingMessage field list

<!-- SOURCE: U06_finetuning.md -->
### Adapter design guide (adapter vs agent vs inline; canonical structure)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) opening: "Canonical guide for building long-running cluster services… Examples drawn from the working adapters in the repository."
- **what:** The decision rule (one external API + multiple internal callers → adapter; long per-orchestration work → spawned agent; short single-agent call → inline; shared infra like DB/Kafka → nothing) plus the canonical shape: struct fields, ordered NewAdapter with manual cleanup on every failure path, sequential fetch-handle-loop (no goroutine-per-message by default), handleMessage parse/dispatch/respond, health endpoints, sync.Once shutdown, topic conventions (`system.adapter.<name>.requests` for new work), config YAML field-name traps, credentials from env only.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md (whole)
- **relations:** thunder adapter (the guide's newest example); response header tiers; deployment essentials
- **verify-later:** consistency of existing adapters with the guide

<!-- SOURCE: U06_finetuning.md -->
### Adapter response-header tier taxonomy and the validator-coverage gap
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** FOCUS_adapter_design(3) "TODO — tighten validator coverage… Tier-2 fields are necessary for the orchestration to advance but not validated… Tracking issue: not yet filed."
- **what:** Response headers split into Tier 1 (five fields the platform Validator enforces; `is_error=true` bypasses), Tier 2 (what the chassis needs to route the reply to the awaiting orchestration — `in_response_to_request_id`, message_type, status vocabulary complete/error_recoverable/error_unrecoverable, is_complete/is_error, etc.; missing these means a silent AWAITING_RESPONSES hang the validator won't catch), Tier 3 (observability). Known live consequence of the gap: the matcher fix of 2026-05-22 (typed response-header struct so booleans serialise as real bools — a map[string]string sent string bools and the chassis dropped the reply). Proposal to extend the validator exists but is unfiled.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#sending-responses,#todo; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics (matcher fix)
- **relations:** reply-topic derivation; send-before-register race (same stuck-await symptom family)
- **verify-later:** platform/validation/Validator current coverage

<!-- SOURCE: U06_finetuning.md -->
### Adapter deployment essentials (manifest, cluster resources, RBAC, Makefile)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) "Deployment essentials — Real lessons from deploying thunder-adapter Phase 2. Every item below is something the deployment failed without."
- **what:** The complete pre-flight for shipping an adapter: serviceAccountName + imagePullSecrets + `command:` (not `args:` — Dockerfiles use CMD, so args replaces the binary path), required Secrets/SA/Docker-Hub grants, explicit KafkaTopic CRDs (Strimzi auto-create is off; missing reply topics fail only at first response), Recreate strategy, single replica, RBAC trap (resourceNames supports no globs — scope by verbs instead for dynamic names like thunder-ssh-<uuid>), four Makefile insertion points and the newName/newTag overlay split, pre/post-deploy checklists.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#deployment-essentials; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md (phase-2 deploy saga)
- **relations:** adapter design guide; wrong-binary image incident; debugging guide §10
- **verify-later:** thunder-adapter kustomize base vs the checklist

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-4 browser-runner adapter (headless Chromium over Kafka)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Stage 6 P0 DEPLOYED + SMOKE PASSED" 2026-07-11 (v1.0.1107; §2.15 smoke matched manual inspection; real bool headers; in_response_to_request_id matcher).
- **what:** A dedicated adapter deployment (image = debian-slim + Playwright + Chromium, playwright-go) consuming `system.adapter.browser-runner.requests` (035 Convention A) and replying on the caller's topic with `{results:[{check_id, profile, url, pass, detail}]}`. P0 scope: desktop 1366×900, three check types (`page_status_ok`, `selector_exists` asserted against the LIVE DOM after settle, `no_console_errors`); everything else honestly reported in `skipped[]`, never faked; browser launched per request so a crash poisons one run, not the pod; navigation failure is a check-fail, not an infra error. Contract pinned to the 035 Adapter Guide as normative after a compliance pass (typed header struct with real bools; `in_response_to_request_id` = incoming request_id is THE matcher; ProduceWithValidation never plain Produce). Build gotchas banked: playwright.azureedge.net CDN dead; v0.6100.0 must be required under its declared (pre-rename) module path; the driver ignores XDG_CACHE_HOME — set HOME in the image.
- **sources:** PLAN_tool_acceptance_runner(2).md (whole); RUNBOOK_travelling_docs(38).md#stage-6; RUNNING_NOTES_travelling_docs(39).md#stage-6-built,#2026-07-11; HANDOFF_2026-07-10…md#T3–T6
- **relations:** analyser-adapter mould (pattern source); 035 adapter guide; tool-acceptance-agent (caller).
- **verify-later:** `cmd/browser-runner-adapter/main.go`; `internal/adapters/browserrunner/`; KafkaTopic CR; dockerfile HOME=/pw-home.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter deployment troubleshooting (ImagePullBackOff / command-vs-args / Kafka topic provisioning)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** Appears only as "## 10. Adapter & Service Deployment Issues" in `debugging_old/016_debugging_guide_v2(1).md`; absent from every subsequent snapshot and from the live consolidated debugging guide (zero grep hits); the content lives instead in `035_adapter_guide.md`.
- **what:** Covers real thunder-adapter-era deployment failures: diagnosing Docker Hub `ImagePullBackOff`/`insufficient_scope`, the Kubernetes `command:` vs `args:` trap (args silently replaces the entire Dockerfile CMD), the Strimzi `auto.create.topics.enable=false` gotcha requiring an explicit KafkaTopic CRD, and a "deployment essentials checklist" required for every new adapter.
- **sources:** debugging_old/016_debugging_guide_v2(1).md#"10"; docs024_key_docs_latest/035_adapter_guide.md#"2.12-2.13"
- **relations:** adapters (035_adapter_guide.md), deployment-github, single-source relocation convention (below)
- **verify-later:** confirm `035_adapter_guide.md` §2.12/§2.13 still matches the checklist.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter Response Envelope Contract relocated from 003 to the adapter guide
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** `debugging_old/003_contracts_and_standards_v10/v11.md` contain the full section; live `003_contracts_and_standards(8).md` replaces it with one line: "Moved to 035_adapter_guide.md §1... now the single source for it."
- **what:** Defines how a long-lived adapter must shape its Kafka reply so the chassis recognises it as an awaited response: reuse the incoming `request_id`, fresh `message_id`, `ProduceWithValidation` (never plain `Produce`), and a typed Go struct for response `headers` (not `map[string]string`) so `is_complete`/`is_error` marshal as real JSON booleans. Motivated by a real production incident (thunder-adapter matcher failure, 2026-05-22).
- **sources:** debugging_old/003_contracts_and_standards_v11.md#"Adapter Response Envelope Contract"; docs024_key_docs_latest/035_adapter_guide.md#"1"
- **relations:** adapters, tool-pipeline, single-source relocation convention (below)
- **verify-later:** check adapter Go source for typed `ResponseHeaders` struct vs any remaining `map[string]string` header builders.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter as sole write credential holder
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "owner decision, 2026-07-12): the GitHub write credential stays in the git-adapter" (SUMMARY_write_step_position_2026-07-12.md)
- **what:** An architecture decision that the fix-implementer never holds a GitHub write token at all; it sends requests to the git-adapter for `create_branch`, `commit`, and `create_pull_request`, exactly as the site-deploy pipeline already does. Chosen over injecting a write token into the implementer's pod, keeping write credentials entirely out of LLM-driven pods. The implementer pod holds only a read-only token via the isRepoCloningAgent spawn gate.
- **sources:** fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** write step; git-adapter new actions; isRepoCloningAgent spawn gate; fix-implementer-orchestrator
- **verify-later:** grep/inspect `create_branch`; `commit`; `create_pull_request`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter new actions (create_branch, create_pull_request, branch-aware commit)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Part 1 BUILT (git-adapter, commit 89175383)... 4-test httptest suite green" (NOTES(10)#Turn 25)
- **what:** New git-adapter capabilities: `create_branch` is idempotent (existing branch returns its head rather than erroring); `create_pull_request` defaults its base to the repo's default branch and is the human review terminal; commit gains an optional `branch` parameter with domain-prefixing skipped for platform commits.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** git-adapter as sole write credential holder; git_adapter_request generic caller
- **verify-later:** grep/inspect `create_branch`; `create_pull_request`; `branch`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Agent-to-adapter capability maturation path (lanes: fast/slow/job)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** "Prove a slow-lane capability as a spawned agent first ... Promote the popular ones to warm adapters ... End-state: a resident chat-orchestrator adapter..." (PLAN_isolated_chat_environment(5).md §11)
- **what:** A general pattern for how agentic capability should mature over time to reduce latency: prove a capability first as a spawned agent, promote popular ones to warm long-running single-replica adapters, and converge on a resident orchestrator adapter that fans out to capability adapters without spawning per request. Framed for chat's three lanes — fast lane (bounded Q&A), slow lane (agentic, latency-warned), job lane (long-running submission + status).
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#11
- **relations:** Isolated chat/satellite architecture (Y-copy); Building-and-hosting as a service via chat
- **verify-later:** whether any chat capability has been promoted past "spawned agent" to a warm adapter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adapter response envelope contract
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "003-vs-FOCUS contradiction RESOLVED empirically (2026-06-11) from coordinator.go + git/thunder/dynamic/websearch adapters" (principles(59)).
- **what:** The chassis's normative contract for any Kafka message replier (adapter or agent): body must be a typed struct with real bools (never `map[string]string` string-bools — this exact bug caused a real multi-day thunder production fault), sent via `ProduceWithValidation` (never bare `Produce`), with `in_response_to_request_id` as the primary matcher the coordinator claims on (`request_id` is a fallback — "reuse both" is the safest pattern), and `action`/payload read from the message BODY (not headers). Originally duplicated and drifted between doc 003 and `FOCUS_adapter_design`; single-sourced into `035_adapter_guide.md` after empirical verification against `coordinator.go` and four live adapters (websearch was the deprecated outlier).
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 envelope-contract entries; NOTES_running_synthesis_v2(36).md/v3(32).md (analyser adapter build referencing the same contract).
- **relations:** Analyser adapter build; canonical-doc-home discipline; code-context retrieval infrastructure.
- **verify-later:** `platform/orchestration/types` `ResponseHeaders`/`ResponseMessage`; `platform/validation` `ValidateOutgoingMessage` (the still-open "promote from prose to validator enforcement" TODO).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Analyser adapter build (polyglot code-parsing service)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION (uk_001)" (principles(59)).
- **what:** A from-scratch Kafka-worker adapter modelled structurally on thunder/git (own image, dockerfile, kustomize base+overlay, config loader, graceful `Shutdown()` with `sync.Once`, health probes) whose one genuine difference is importing the shared chassis-root `internal/analysis` package and holding a dedicated least-privilege, read-only, repo-scoped GitHub token via `secretKeyRef` (never passed through the spawning pod). Fetches via a tarball GET (no git binary, no go-git), not a clone.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11/12 adapter-build entries; NOTES_running_synthesis_v3(32).md (turns 25-27, tarball-fetcher reuse into `internal/reposource`).
- **relations:** Adapter response envelope contract; code-context retrieval infrastructure; GitHub read-token scoping pattern.

<!-- SOURCE: U15_docs019_running_notes.md -->
### GitHub read-token scoping / least-privilege adapter secrets pattern
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25)."
- **what:** `spawn_actions.go` injects `GITHUB_READ_TOKEN` from a shared platform secret only for agent types flagged `isRepoCloningAgent` (currently just `diagnose-agent`), via `secretKeyRef` so the spawning pod itself never holds the token and no other agent type is granted it — the same read-only single-repo PAT the analyser adapter uses.
- **sources:** NOTES_running_synthesis_v3(32).md DECISIONS (turn 25).
- **relations:** Analyser adapter build; diagnose-agent self-contained repo fetch.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Thunder adapter (GPU provisioning, reaper, cost caps, credential boundary)
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** 033 header "Proposal for routing all Thunder Compute interactions through a long-running cluster adapter"; debugging guide v2_44 shows Phases 2–6 progressing
- **what:** A single long-lived `thunder-adapter` Deployment that holds the Thunder API key/B2 creds/SSH keypair store, provisions ephemeral GPU VMs via Kafka actions, and preserves a credential boundary: VMs get only ephemeral SSH keys + hours-expiring presigned URLs. Defence-in-depth: Thunder hard 12h uptime cap + a 15-min reaper + a daily cost cap.
- **sources:** WM/033_thunder_adapter_design.md#tldr, WM/033_thunder_adapter_design.md#preventing-indefinite-running-gpus-defence-in-depth, WM/033_thunder_adapter_design.md#new-schema, WM/016_debugging_guide_v2_44.md#9
- **relations:** fine-tuning flywheel; adapters pattern; multicluster provisioning
- **verify-later:** thunder_instances; thunder_budget_state; model_lifecycle.training_runs; system.adapter.thunder.requests

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adapter & service deployment debugging (rescued/dropped section)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** 016 base "Adapter & Service Deployment Issues … Rescued from an earlier guide revision … dropped from the main line"; absent from v2_44
- **what:** A family-delta section present in the base 016 but dropped from the v2_x main line: diagnosing adapter deployment failures (ImagePullBackOff/`insufficient_scope`, immediate crashes from `args:` replacing the whole CMD, `Unknown Topic Or Partition` on first message) and a deployment-essentials checklist. Built from the thunder-adapter Phase 2 debugging.
- **sources:** WM/016_debugging_guide.md#adapter-service-deployment-issues, WM/016_debugging_guide.md#imagepullbackoff-insufficient_scope-authorization-failed
- **relations:** dropped from 016 v2_44 (superseded main line); Thunder adapter; deployment-github
- **verify-later:** kustomize base/deployment.yaml; docker-hub-creds secret; ai-persona-app service account

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser-adapter deployment/migration plan
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** README.md marks most destinations "(NEW)"/"(EDIT)" against the real repo tree (tree -d, 2026-06-11), but notes "`NNN_create_code_symbols_index.sql` → workspace root (ALREADY APPLIED — commit for the record, your numbering)"
- **what:** A directory-by-directory migration map from a `chassis-drafts/analyser-adapter` staging area (which does not compile in this tree) to real agentchassis destinations: `cmd/analyser-adapter/main.go`, `internal/adapters/analyser/{adapter,analyse_action,github_source}.go`, `platform/orchestration/actions/{code_symbols_actions,analyser_request_action}.go` (+ registry.go insertion), the code-indexer migration, `configs/analyser-adapter.yaml`, a two-stage Dockerfile, and kustomize base/overlay scaffolding — all following the conventions already used for thunder-adapter. Also flags un-placeable items needing a human call: the `035_adapter_guide.md` doc home, the `system.adapter.analyser.requests` KafkaTopic CRD location, and the `analyser-github-read` Secret (never committed with a real token).
- **sources:** contextkit/README.md, contextkit/README(2).md
- **relations:** code-indexer agent, adapters (033/035 thunder/webscrape pattern), deployment-github (034)
- **verify-later:** build/docker/backend/, deployments/kustomize/services/analyser-adapter/, Makefile — confirm which of the four described insertions actually exist

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder adapter schema and provisioning gates
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Full schema with recorded user decisions ($100/day cap, 2 concurrent, 18h uptime, $1.80/hr A100, $25 estimated run); production fix dated 2026-05-22 for identifier recycling; ssh_port verification dated 2026-05-24.
- **what:** GPU VM lifecycle for training: thunder_instances (one row per VM ever provisioned — inserted BEFORE the API call so the reaper always has a record; status machine provisioning→running→decommissioning→decommissioned with reaped/lost/failed terminals; cost snapshot; reaper bookkeeping; FK to model_lifecycle.training_runs), thunder_config singleton (CHECK-enforced single row; caps and pause switch), and computed views thunder_spend_24h (rolling cost incl. running estimates, no drifting counter) and thunder_provision_check (can_provision + denial_reason evaluated at every provision request). Identifier recycling fixed by replacing global uniqueness with a partial unique index over live states only; ssh_port captured at provision so ssh_exec dials directly.
- **sources:** docs/agent_docs/sql_for_tables/042_thunder.sql
- **relations:** thunder unreachable streak; training_runs (flywheel C); 013_thunder_adapter_design.md.
- **verify-later:** adapter reading thunder_provision_check; reaper behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder consecutive-unreachable probe streak
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Migration 106 with in-transaction verification; rationale documented (single down-probe could be a transient SSH blip; each scheduler tick is a fresh sub-agent that can't hold state in memory).
- **what:** thunder-training-monitor durability: consecutive_unreachable_probes counter (+ last_probe_at) on thunder_instances, bumped/reset by the record_probe_streak action; only after the streak crosses a threshold is the instance treated as 'lost' (fail run + decommission). State lives on the row because monitor ticks are stateless.
- **sources:** docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql
- **relations:** thunder adapter; scheduler tick statelessness.
- **verify-later:** record_probe_streak action; threshold value in monitor config.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adapter microservice pattern (Kafka/HTTP adapters + secure external-API proxies)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 0123 codifies the pattern; image-generator, firecrawl/webscrape, playwright, git adapters all follow it; "We will use this exact same pattern for all our Python-based actions."
- **what:** Go agents never embed heavy dependencies: a workflow action produces a Kafka message to `system.adapter.<name>.requests` (or an internal HTTP call); a containerised worker service (Python or Go) in its own Deployment consumes via a shared consumer group, does the work (Playwright, Firecrawl, Stability, git), and replies to the reply_to topic. External GPU/API providers get a dedicated Go proxy adapter that holds the secret key and translates request formats — swap providers by changing one adapter, no workflow changes.
- **sources:** docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md; docs004_website_capture_project/playwright/implementation_roadmap.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md
- **relations:** adapters category anchor (docs 033/035 successor); image adapter; firecrawl adapter; thunder adapter (taxonomy) is a descendant of the "ThunderCompute LLaVA proxy" idea here.
- **verify-later:** internal/adapters/* inventory; which adapter topics exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Firecrawl scraping adapter and actions
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Agent definitions website-capture-firecrawl and webscrape-simple with image tags v1.0.407→v1.0.424 (iteration = real use); v1→v2 migration doc fixing "Unrecognized keys" errors and adding S3 ownership of screenshots/images.
- **what:** Firecrawl API adapter (Kafka consumer on system.adapter.firecrawl.requests) exposing scrape/crawl/extract actions to workflows (firecrawl_scrape, firecrawl_crawl, firecrawl_extract, plus a registered scrape_web action with upload_results to S3). v2 migration: formats array incl. screenshot+links, downloading Google-Cloud-hosted screenshots/images into own S3 (webscrape/client/date/id/ layout) for data ownership since Firecrawl assets expire in 30 days. Chosen over the half-built Playwright adapter to reduce MVP load.
- **sources:** docs003_firecrawl/README.0126.firecrawl_agent_definition.md; docs004_website_capture_project/firecrawl/001claude_initial.md; docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md; docs003_firecrawl/README.0129.testing_webscrape_message.md
- **relations:** adoption-pipeline crawling (live successor); playwright adapter (the road not taken); storage-architecture.
- **verify-later:** web-scrape-adapter deployment (referenced in initial_messages.txt scale-down list — so it was deployed); FIRECRAWL_API_KEY secret.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Adapter design pattern (canonical guide)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** context-pack FOCUS_adapter_design(3).md is a frozen copy; live docs advance to flywheel_docs/FOCUS_adapter_design(3).md and docs024_key_docs_latest/035_adapter_guide.md (38KB, Jun 2026)
- **what:** The canonical guide for building single-replica Kafka-consuming "adapter" services that wrap one external API and hold its credentials (git, web-scrape, image-generator, ollama, thunder). Covers the Adapter struct, NewAdapter cleanup convention, sequential Run loop, handleMessage dispatch, the three-tier response-header contract (Tier-1 validator-required, Tier-2 chassis-routing e.g. `in_response_to_request_id`, Tier-3 observability), health/shutdown, topic naming conventions A vs B, and deployment essentials (serviceAccountName, imagePullSecrets, `command:` not `args:`, Strimzi topic pre-creation).
- **sources:** docubundle/.../FOCUS_adapter_design(3).md#TL;DR, #Responsibilities, #Sending-responses, #Deployment-essentials
- **relations:** superseded by live 035_adapter_guide.md; instantiated by thunder-adapter; response-header contract underpins the reply-topic bugs
- **verify-later:** internal/adapters/*/adapter.go; platform/validation/Validator; 035_adapter_guide.md

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-adapter — Thunder Compute GPU provisioning adapter
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter_2026-06_04 §1 phases 3.0–3.6; FOCUS(21) §14 "Provision loop verified end-to-end (2026-05-22)"
- **what:** The adapter that wraps the Thunder Compute API to provision/decommission on-demand GPU VMs, holding the Thunder token and B2 keys. Actions: `provision_instance` (spend-check → ed25519 keypair → create → WaitForRunning → INSERT `thunder_instances` with compensating cleanup), `decommission_instance` (idempotent, computes cost from running_since), plus SSH (`ssh_exec`, `ssh_get_status`) and presign actions. Two matcher bugs that blocked it for days: response headers must be a typed struct (not map[string]string) so is_complete/is_error serialise as JSON bools; and `thunder_instance_id` uniqueness must be a partial index on live rows because Thunder recycles numeric ids.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1, #3; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics)
- **relations:** called by gpu-provisioner, training-launcher, thunder-reaper, thunder-training-monitor; credential boundary for presigned URLs
- **verify-later:** internal/adapters/thunder/api/types.go, provision_action.go, decommission_action.go; thunder_instances, thunder_config, thunder_provision_check

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder Compute API specifics (field/casing/template traps)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §14 "Request/response field shape is asymmetric (verified 2026-05-20 via tnr status --json)"
- **what:** Hard-won Thunder API facts: base URL `https://api.thundercompute.com:8443/v1`; CREATE uses snake_case ints (gpu_type, cpu_cores, num_gpus) but STATUS/LIST returns camelCase with numbers as JSON strings and UPPERCASE enums; real templates are `base`/`ollama`/`comfy-ui`/`forge-neo`/`unsloth` (the OpenAPI `ubuntu-22.04` example is rejected); the login user is `ubuntu` not `root`; SSH needs wait-for-sshd (RUNNING ≠ sshd ready); the SSH port from list is unreliable, use `tnr connect --json`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics; Phase 4 SSH item)
- **relations:** underpins thunder-adapter; `IdentifierInt()`/`IsReadyStatus` handling
- **verify-later:** internal/adapters/thunder/api/types.go (CreateInstanceRequest vs Instance)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P5 vmhost provisioning adapter (superseded service-deployer)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** plan(11) P5 "A vmhost adapter for what DOES need SSH … built to the analyser-adapter README skeleton"; earlier plan(1)/(4)/(5) P5 "registry + relocation (service_instances) and, eventually, the chassis service-deployer adapter".
- **what:** The eventual automation for what genuinely needs SSH: provision box, run setup.sh, onboard domain, ship engine, decommission — built as a `vmhost` adapter (cmd/vmhost-adapter, internal/adapters/vmhost, reuse thunder's ssh via shared/, kustomize, KafkaTopic system.adapter.vmhost.requests, 003 envelope), with a `service_instances` registry modelled on thunder_instances minus the reaper. Long-term it holds the deploy SSH credential, retiring the repo-secrets copy. Earlier versions named this the "service-deployer" adapter.
- **sources:** traffic_probe_plan(11).md#p5, traffic_probe_plan(1).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** future handler for backend_unreachable; supersedes "service-deployer" naming
- **verify-later:** analyser-adapter README; thunder_instances → service_instances

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### adapter response-envelope contract (request/response wiring conventions)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** adapter(4).go's header states the envelope is "grounded in the orchestrator (coordinator.go) and the three working adapters, not the docs — which disagree (003 vs FOCUS) and were resolved empirically." internal/adapters/analyser/ exists live in the repo, confirming the draft shipped.
- **what:** A reverse-engineered (from working adapters, not from docs) contract for how an adapter must shape its Kafka request/response envelope so the orchestrator actually routes the reply instead of timing out: action comes from `body.action` not headers; `in_response_to_request_id` (echoing the incoming `request_id`) is the load-bearing claim field in `coordinator.go`'s `ProcessResponse`, with `request_id` as fallback; the reply body must use canonical `types.ResponseHeaders` via `ToResponseHeaders` so `is_complete`/`is_error` marshal as real JSON bools (the "bool trap" — a `map[string]string` sending the string `"true"` fails the receiver's struct-bool unmarshal); sends must go via `ProduceWithValidation`, never plain `Produce`. websearch-adapter is flagged as the one adapter still on the deprecated string-bool/plain-Produce path.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go
- **relations:** contextkit toolchain (above); analyser-adapter service
- **verify-later:** platform/orchestration/coordinator.go ProcessResponse, internal/adapters/analyser/adapter.go, whether websearch-adapter has since been migrated off the string-bool map

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adapter/response message envelope contract (normative)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 035 §1 "last verified against code 2026-06-11"; 003(8) now points here
- **what:** Any reply to a chassis request must be recognised as an awaited response or it silently falls to process-as-work (row stuck waiting, ~10-min retries, no error). Load-bearing field: in_response_to_request_id = incoming request_id (request_id fallback only; git adapter's reuse pattern favoured). Three header tiers (validator-enforced / coordinator-needed-but-unvalidated / observability). Body headers MUST be a typed struct with real bools (map[string]string string-bools fail unmarshal and drop the reply pre-claim — the multi-day thunder outage). Send via ProduceWithValidation. Request parsing: action from body, payload at body.data, accept reply-topic from three keys. Sibling race: local dispatches must preRegisterAwaitedRequest before send (confirmed fixed in prod 06-09).
- **sources:** 035 §1; 016 §9 bool-trap + race entries
- **relations:** awaited_requests; O(K²) batch presign
- **verify-later:** ValidateOutgoingMessage field list

<!-- SOURCE: U06_finetuning.md -->
### Adapter design guide (adapter vs agent vs inline; canonical structure)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) opening: "Canonical guide for building long-running cluster services… Examples drawn from the working adapters in the repository."
- **what:** The decision rule (one external API + multiple internal callers → adapter; long per-orchestration work → spawned agent; short single-agent call → inline; shared infra like DB/Kafka → nothing) plus the canonical shape: struct fields, ordered NewAdapter with manual cleanup on every failure path, sequential fetch-handle-loop (no goroutine-per-message by default), handleMessage parse/dispatch/respond, health endpoints, sync.Once shutdown, topic conventions (`system.adapter.<name>.requests` for new work), config YAML field-name traps, credentials from env only.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md (whole)
- **relations:** thunder adapter (the guide's newest example); response header tiers; deployment essentials
- **verify-later:** consistency of existing adapters with the guide

<!-- SOURCE: U06_finetuning.md -->
### Adapter response-header tier taxonomy and the validator-coverage gap
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** FOCUS_adapter_design(3) "TODO — tighten validator coverage… Tier-2 fields are necessary for the orchestration to advance but not validated… Tracking issue: not yet filed."
- **what:** Response headers split into Tier 1 (five fields the platform Validator enforces; `is_error=true` bypasses), Tier 2 (what the chassis needs to route the reply to the awaiting orchestration — `in_response_to_request_id`, message_type, status vocabulary complete/error_recoverable/error_unrecoverable, is_complete/is_error, etc.; missing these means a silent AWAITING_RESPONSES hang the validator won't catch), Tier 3 (observability). Known live consequence of the gap: the matcher fix of 2026-05-22 (typed response-header struct so booleans serialise as real bools — a map[string]string sent string bools and the chassis dropped the reply). Proposal to extend the validator exists but is unfiled.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#sending-responses,#todo; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics (matcher fix)
- **relations:** reply-topic derivation; send-before-register race (same stuck-await symptom family)
- **verify-later:** platform/validation/Validator current coverage

<!-- SOURCE: U06_finetuning.md -->
### Adapter deployment essentials (manifest, cluster resources, RBAC, Makefile)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) "Deployment essentials — Real lessons from deploying thunder-adapter Phase 2. Every item below is something the deployment failed without."
- **what:** The complete pre-flight for shipping an adapter: serviceAccountName + imagePullSecrets + `command:` (not `args:` — Dockerfiles use CMD, so args replaces the binary path), required Secrets/SA/Docker-Hub grants, explicit KafkaTopic CRDs (Strimzi auto-create is off; missing reply topics fail only at first response), Recreate strategy, single replica, RBAC trap (resourceNames supports no globs — scope by verbs instead for dynamic names like thunder-ssh-<uuid>), four Makefile insertion points and the newName/newTag overlay split, pre/post-deploy checklists.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#deployment-essentials; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md (phase-2 deploy saga)
- **relations:** adapter design guide; wrong-binary image incident; debugging guide §10
- **verify-later:** thunder-adapter kustomize base vs the checklist

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-4 browser-runner adapter (headless Chromium over Kafka)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Stage 6 P0 DEPLOYED + SMOKE PASSED" 2026-07-11 (v1.0.1107; §2.15 smoke matched manual inspection; real bool headers; in_response_to_request_id matcher).
- **what:** A dedicated adapter deployment (image = debian-slim + Playwright + Chromium, playwright-go) consuming `system.adapter.browser-runner.requests` (035 Convention A) and replying on the caller's topic with `{results:[{check_id, profile, url, pass, detail}]}`. P0 scope: desktop 1366×900, three check types (`page_status_ok`, `selector_exists` asserted against the LIVE DOM after settle, `no_console_errors`); everything else honestly reported in `skipped[]`, never faked; browser launched per request so a crash poisons one run, not the pod; navigation failure is a check-fail, not an infra error. Contract pinned to the 035 Adapter Guide as normative after a compliance pass (typed header struct with real bools; `in_response_to_request_id` = incoming request_id is THE matcher; ProduceWithValidation never plain Produce). Build gotchas banked: playwright.azureedge.net CDN dead; v0.6100.0 must be required under its declared (pre-rename) module path; the driver ignores XDG_CACHE_HOME — set HOME in the image.
- **sources:** PLAN_tool_acceptance_runner(2).md (whole); RUNBOOK_travelling_docs(38).md#stage-6; RUNNING_NOTES_travelling_docs(39).md#stage-6-built,#2026-07-11; HANDOFF_2026-07-10…md#T3–T6
- **relations:** analyser-adapter mould (pattern source); 035 adapter guide; tool-acceptance-agent (caller).
- **verify-later:** `cmd/browser-runner-adapter/main.go`; `internal/adapters/browserrunner/`; KafkaTopic CR; dockerfile HOME=/pw-home.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter deployment troubleshooting (ImagePullBackOff / command-vs-args / Kafka topic provisioning)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** Appears only as "## 10. Adapter & Service Deployment Issues" in `debugging_old/016_debugging_guide_v2(1).md`; absent from every subsequent snapshot and from the live consolidated debugging guide (zero grep hits); the content lives instead in `035_adapter_guide.md`.
- **what:** Covers real thunder-adapter-era deployment failures: diagnosing Docker Hub `ImagePullBackOff`/`insufficient_scope`, the Kubernetes `command:` vs `args:` trap (args silently replaces the entire Dockerfile CMD), the Strimzi `auto.create.topics.enable=false` gotcha requiring an explicit KafkaTopic CRD, and a "deployment essentials checklist" required for every new adapter.
- **sources:** debugging_old/016_debugging_guide_v2(1).md#"10"; docs024_key_docs_latest/035_adapter_guide.md#"2.12-2.13"
- **relations:** adapters (035_adapter_guide.md), deployment-github, single-source relocation convention (below)
- **verify-later:** confirm `035_adapter_guide.md` §2.12/§2.13 still matches the checklist.

<!-- SOURCE: U12_docs024_archives.md -->
### Adapter Response Envelope Contract relocated from 003 to the adapter guide
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** `debugging_old/003_contracts_and_standards_v10/v11.md` contain the full section; live `003_contracts_and_standards(8).md` replaces it with one line: "Moved to 035_adapter_guide.md §1... now the single source for it."
- **what:** Defines how a long-lived adapter must shape its Kafka reply so the chassis recognises it as an awaited response: reuse the incoming `request_id`, fresh `message_id`, `ProduceWithValidation` (never plain `Produce`), and a typed Go struct for response `headers` (not `map[string]string`) so `is_complete`/`is_error` marshal as real JSON booleans. Motivated by a real production incident (thunder-adapter matcher failure, 2026-05-22).
- **sources:** debugging_old/003_contracts_and_standards_v11.md#"Adapter Response Envelope Contract"; docs024_key_docs_latest/035_adapter_guide.md#"1"
- **relations:** adapters, tool-pipeline, single-source relocation convention (below)
- **verify-later:** check adapter Go source for typed `ResponseHeaders` struct vs any remaining `map[string]string` header builders.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter as sole write credential holder
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "owner decision, 2026-07-12): the GitHub write credential stays in the git-adapter" (SUMMARY_write_step_position_2026-07-12.md)
- **what:** An architecture decision that the fix-implementer never holds a GitHub write token at all; it sends requests to the git-adapter for `create_branch`, `commit`, and `create_pull_request`, exactly as the site-deploy pipeline already does. Chosen over injecting a write token into the implementer's pod, keeping write credentials entirely out of LLM-driven pods. The implementer pod holds only a read-only token via the isRepoCloningAgent spawn gate.
- **sources:** fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** write step; git-adapter new actions; isRepoCloningAgent spawn gate; fix-implementer-orchestrator
- **verify-later:** grep/inspect `create_branch`; `commit`; `create_pull_request`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### git-adapter new actions (create_branch, create_pull_request, branch-aware commit)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "Part 1 BUILT (git-adapter, commit 89175383)... 4-test httptest suite green" (NOTES(10)#Turn 25)
- **what:** New git-adapter capabilities: `create_branch` is idempotent (existing branch returns its head rather than erroring); `create_pull_request` defaults its base to the repo's default branch and is the human review terminal; commit gains an optional `branch` parameter with domain-prefixing skipped for platform commits.
- **sources:** fixloop_eg_dartsonline/0NN_fix_implementer.sql#create_branch/commit_files/create_pr, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 25
- **relations:** git-adapter as sole write credential holder; git_adapter_request generic caller
- **verify-later:** grep/inspect `create_branch`; `create_pull_request`; `branch`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Agent-to-adapter capability maturation path (lanes: fast/slow/job)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** "Prove a slow-lane capability as a spawned agent first ... Promote the popular ones to warm adapters ... End-state: a resident chat-orchestrator adapter..." (PLAN_isolated_chat_environment(5).md §11)
- **what:** A general pattern for how agentic capability should mature over time to reduce latency: prove a capability first as a spawned agent, promote popular ones to warm long-running single-replica adapters, and converge on a resident orchestrator adapter that fans out to capability adapters without spawning per request. Framed for chat's three lanes — fast lane (bounded Q&A), slow lane (agentic, latency-warned), job lane (long-running submission + status).
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#11
- **relations:** Isolated chat/satellite architecture (Y-copy); Building-and-hosting as a service via chat
- **verify-later:** whether any chat capability has been promoted past "spawned agent" to a warm adapter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adapter response envelope contract
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "003-vs-FOCUS contradiction RESOLVED empirically (2026-06-11) from coordinator.go + git/thunder/dynamic/websearch adapters" (principles(59)).
- **what:** The chassis's normative contract for any Kafka message replier (adapter or agent): body must be a typed struct with real bools (never `map[string]string` string-bools — this exact bug caused a real multi-day thunder production fault), sent via `ProduceWithValidation` (never bare `Produce`), with `in_response_to_request_id` as the primary matcher the coordinator claims on (`request_id` is a fallback — "reuse both" is the safest pattern), and `action`/payload read from the message BODY (not headers). Originally duplicated and drifted between doc 003 and `FOCUS_adapter_design`; single-sourced into `035_adapter_guide.md` after empirical verification against `coordinator.go` and four live adapters (websearch was the deprecated outlier).
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 envelope-contract entries; NOTES_running_synthesis_v2(36).md/v3(32).md (analyser adapter build referencing the same contract).
- **relations:** Analyser adapter build; canonical-doc-home discipline; code-context retrieval infrastructure.
- **verify-later:** `platform/orchestration/types` `ResponseHeaders`/`ResponseMessage`; `platform/validation` `ValidateOutgoingMessage` (the still-open "promote from prose to validator enforcement" TODO).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Analyser adapter build (polyglot code-parsing service)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION (uk_001)" (principles(59)).
- **what:** A from-scratch Kafka-worker adapter modelled structurally on thunder/git (own image, dockerfile, kustomize base+overlay, config loader, graceful `Shutdown()` with `sync.Once`, health probes) whose one genuine difference is importing the shared chassis-root `internal/analysis` package and holding a dedicated least-privilege, read-only, repo-scoped GitHub token via `secretKeyRef` (never passed through the spawning pod). Fetches via a tarball GET (no git binary, no go-git), not a clone.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11/12 adapter-build entries; NOTES_running_synthesis_v3(32).md (turns 25-27, tarball-fetcher reuse into `internal/reposource`).
- **relations:** Adapter response envelope contract; code-context retrieval infrastructure; GitHub read-token scoping pattern.

<!-- SOURCE: U15_docs019_running_notes.md -->
### GitHub read-token scoping / least-privilege adapter secrets pattern
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "GitHub read token scoped to the diagnoser via secretKeyRef, not passthrough (turn 25)."
- **what:** `spawn_actions.go` injects `GITHUB_READ_TOKEN` from a shared platform secret only for agent types flagged `isRepoCloningAgent` (currently just `diagnose-agent`), via `secretKeyRef` so the spawning pod itself never holds the token and no other agent type is granted it — the same read-only single-repo PAT the analyser adapter uses.
- **sources:** NOTES_running_synthesis_v3(32).md DECISIONS (turn 25).
- **relations:** Analyser adapter build; diagnose-agent self-contained repo fetch.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Thunder adapter (GPU provisioning, reaper, cost caps, credential boundary)
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** 033 header "Proposal for routing all Thunder Compute interactions through a long-running cluster adapter"; debugging guide v2_44 shows Phases 2–6 progressing
- **what:** A single long-lived `thunder-adapter` Deployment that holds the Thunder API key/B2 creds/SSH keypair store, provisions ephemeral GPU VMs via Kafka actions, and preserves a credential boundary: VMs get only ephemeral SSH keys + hours-expiring presigned URLs. Defence-in-depth: Thunder hard 12h uptime cap + a 15-min reaper + a daily cost cap.
- **sources:** WM/033_thunder_adapter_design.md#tldr, WM/033_thunder_adapter_design.md#preventing-indefinite-running-gpus-defence-in-depth, WM/033_thunder_adapter_design.md#new-schema, WM/016_debugging_guide_v2_44.md#9
- **relations:** fine-tuning flywheel; adapters pattern; multicluster provisioning
- **verify-later:** thunder_instances; thunder_budget_state; model_lifecycle.training_runs; system.adapter.thunder.requests

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adapter & service deployment debugging (rescued/dropped section)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** 016 base "Adapter & Service Deployment Issues … Rescued from an earlier guide revision … dropped from the main line"; absent from v2_44
- **what:** A family-delta section present in the base 016 but dropped from the v2_x main line: diagnosing adapter deployment failures (ImagePullBackOff/`insufficient_scope`, immediate crashes from `args:` replacing the whole CMD, `Unknown Topic Or Partition` on first message) and a deployment-essentials checklist. Built from the thunder-adapter Phase 2 debugging.
- **sources:** WM/016_debugging_guide.md#adapter-service-deployment-issues, WM/016_debugging_guide.md#imagepullbackoff-insufficient_scope-authorization-failed
- **relations:** dropped from 016 v2_44 (superseded main line); Thunder adapter; deployment-github
- **verify-later:** kustomize base/deployment.yaml; docker-hub-creds secret; ai-persona-app service account

<!-- SOURCE: U17b_docs019_gofiles.md -->
### analyser-adapter deployment/migration plan
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** README.md marks most destinations "(NEW)"/"(EDIT)" against the real repo tree (tree -d, 2026-06-11), but notes "`NNN_create_code_symbols_index.sql` → workspace root (ALREADY APPLIED — commit for the record, your numbering)"
- **what:** A directory-by-directory migration map from a `chassis-drafts/analyser-adapter` staging area (which does not compile in this tree) to real agentchassis destinations: `cmd/analyser-adapter/main.go`, `internal/adapters/analyser/{adapter,analyse_action,github_source}.go`, `platform/orchestration/actions/{code_symbols_actions,analyser_request_action}.go` (+ registry.go insertion), the code-indexer migration, `configs/analyser-adapter.yaml`, a two-stage Dockerfile, and kustomize base/overlay scaffolding — all following the conventions already used for thunder-adapter. Also flags un-placeable items needing a human call: the `035_adapter_guide.md` doc home, the `system.adapter.analyser.requests` KafkaTopic CRD location, and the `analyser-github-read` Secret (never committed with a real token).
- **sources:** contextkit/README.md, contextkit/README(2).md
- **relations:** code-indexer agent, adapters (033/035 thunder/webscrape pattern), deployment-github (034)
- **verify-later:** build/docker/backend/, deployments/kustomize/services/analyser-adapter/, Makefile — confirm which of the four described insertions actually exist

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder adapter schema and provisioning gates
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Full schema with recorded user decisions ($100/day cap, 2 concurrent, 18h uptime, $1.80/hr A100, $25 estimated run); production fix dated 2026-05-22 for identifier recycling; ssh_port verification dated 2026-05-24.
- **what:** GPU VM lifecycle for training: thunder_instances (one row per VM ever provisioned — inserted BEFORE the API call so the reaper always has a record; status machine provisioning→running→decommissioning→decommissioned with reaped/lost/failed terminals; cost snapshot; reaper bookkeeping; FK to model_lifecycle.training_runs), thunder_config singleton (CHECK-enforced single row; caps and pause switch), and computed views thunder_spend_24h (rolling cost incl. running estimates, no drifting counter) and thunder_provision_check (can_provision + denial_reason evaluated at every provision request). Identifier recycling fixed by replacing global uniqueness with a partial unique index over live states only; ssh_port captured at provision so ssh_exec dials directly.
- **sources:** docs/agent_docs/sql_for_tables/042_thunder.sql
- **relations:** thunder unreachable streak; training_runs (flywheel C); 013_thunder_adapter_design.md.
- **verify-later:** adapter reading thunder_provision_check; reaper behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Thunder consecutive-unreachable probe streak
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Migration 106 with in-transaction verification; rationale documented (single down-probe could be a transient SSH blip; each scheduler tick is a fresh sub-agent that can't hold state in memory).
- **what:** thunder-training-monitor durability: consecutive_unreachable_probes counter (+ last_probe_at) on thunder_instances, bumped/reset by the record_probe_streak action; only after the streak crosses a threshold is the instance treated as 'lost' (fail run + decommission). State lives on the row because monitor ticks are stateless.
- **sources:** docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql
- **relations:** thunder adapter; scheduler tick statelessness.
- **verify-later:** record_probe_streak action; threshold value in monitor config.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adapter microservice pattern (Kafka/HTTP adapters + secure external-API proxies)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 0123 codifies the pattern; image-generator, firecrawl/webscrape, playwright, git adapters all follow it; "We will use this exact same pattern for all our Python-based actions."
- **what:** Go agents never embed heavy dependencies: a workflow action produces a Kafka message to `system.adapter.<name>.requests` (or an internal HTTP call); a containerised worker service (Python or Go) in its own Deployment consumes via a shared consumer group, does the work (Playwright, Firecrawl, Stability, git), and replies to the reply_to topic. External GPU/API providers get a dedicated Go proxy adapter that holds the secret key and translates request formats — swap providers by changing one adapter, no workflow changes.
- **sources:** docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md; docs004_website_capture_project/playwright/implementation_roadmap.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md
- **relations:** adapters category anchor (docs 033/035 successor); image adapter; firecrawl adapter; thunder adapter (taxonomy) is a descendant of the "ThunderCompute LLaVA proxy" idea here.
- **verify-later:** internal/adapters/* inventory; which adapter topics exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Firecrawl scraping adapter and actions
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Agent definitions website-capture-firecrawl and webscrape-simple with image tags v1.0.407→v1.0.424 (iteration = real use); v1→v2 migration doc fixing "Unrecognized keys" errors and adding S3 ownership of screenshots/images.
- **what:** Firecrawl API adapter (Kafka consumer on system.adapter.firecrawl.requests) exposing scrape/crawl/extract actions to workflows (firecrawl_scrape, firecrawl_crawl, firecrawl_extract, plus a registered scrape_web action with upload_results to S3). v2 migration: formats array incl. screenshot+links, downloading Google-Cloud-hosted screenshots/images into own S3 (webscrape/client/date/id/ layout) for data ownership since Firecrawl assets expire in 30 days. Chosen over the half-built Playwright adapter to reduce MVP load.
- **sources:** docs003_firecrawl/README.0126.firecrawl_agent_definition.md; docs004_website_capture_project/firecrawl/001claude_initial.md; docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md; docs003_firecrawl/README.0129.testing_webscrape_message.md
- **relations:** adoption-pipeline crawling (live successor); playwright adapter (the road not taken); storage-architecture.
- **verify-later:** web-scrape-adapter deployment (referenced in initial_messages.txt scale-down list — so it was deployed); FIRECRAWL_API_KEY secret.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Adapter design pattern (canonical guide)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** context-pack FOCUS_adapter_design(3).md is a frozen copy; live docs advance to flywheel_docs/FOCUS_adapter_design(3).md and docs024_key_docs_latest/035_adapter_guide.md (38KB, Jun 2026)
- **what:** The canonical guide for building single-replica Kafka-consuming "adapter" services that wrap one external API and hold its credentials (git, web-scrape, image-generator, ollama, thunder). Covers the Adapter struct, NewAdapter cleanup convention, sequential Run loop, handleMessage dispatch, the three-tier response-header contract (Tier-1 validator-required, Tier-2 chassis-routing e.g. `in_response_to_request_id`, Tier-3 observability), health/shutdown, topic naming conventions A vs B, and deployment essentials (serviceAccountName, imagePullSecrets, `command:` not `args:`, Strimzi topic pre-creation).
- **sources:** docubundle/.../FOCUS_adapter_design(3).md#TL;DR, #Responsibilities, #Sending-responses, #Deployment-essentials
- **relations:** superseded by live 035_adapter_guide.md; instantiated by thunder-adapter; response-header contract underpins the reply-topic bugs
- **verify-later:** internal/adapters/*/adapter.go; platform/validation/Validator; 035_adapter_guide.md

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-adapter — Thunder Compute GPU provisioning adapter
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter_2026-06_04 §1 phases 3.0–3.6; FOCUS(21) §14 "Provision loop verified end-to-end (2026-05-22)"
- **what:** The adapter that wraps the Thunder Compute API to provision/decommission on-demand GPU VMs, holding the Thunder token and B2 keys. Actions: `provision_instance` (spend-check → ed25519 keypair → create → WaitForRunning → INSERT `thunder_instances` with compensating cleanup), `decommission_instance` (idempotent, computes cost from running_since), plus SSH (`ssh_exec`, `ssh_get_status`) and presign actions. Two matcher bugs that blocked it for days: response headers must be a typed struct (not map[string]string) so is_complete/is_error serialise as JSON bools; and `thunder_instance_id` uniqueness must be a partial index on live rows because Thunder recycles numeric ids.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1, #3; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics)
- **relations:** called by gpu-provisioner, training-launcher, thunder-reaper, thunder-training-monitor; credential boundary for presigned URLs
- **verify-later:** internal/adapters/thunder/api/types.go, provision_action.go, decommission_action.go; thunder_instances, thunder_config, thunder_provision_check

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder Compute API specifics (field/casing/template traps)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §14 "Request/response field shape is asymmetric (verified 2026-05-20 via tnr status --json)"
- **what:** Hard-won Thunder API facts: base URL `https://api.thundercompute.com:8443/v1`; CREATE uses snake_case ints (gpu_type, cpu_cores, num_gpus) but STATUS/LIST returns camelCase with numbers as JSON strings and UPPERCASE enums; real templates are `base`/`ollama`/`comfy-ui`/`forge-neo`/`unsloth` (the OpenAPI `ubuntu-22.04` example is rejected); the login user is `ubuntu` not `root`; SSH needs wait-for-sshd (RUNNING ≠ sshd ready); the SSH port from list is unreliable, use `tnr connect --json`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics; Phase 4 SSH item)
- **relations:** underpins thunder-adapter; `IdentifierInt()`/`IsReadyStatus` handling
- **verify-later:** internal/adapters/thunder/api/types.go (CreateInstanceRequest vs Instance)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P5 vmhost provisioning adapter (superseded service-deployer)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** plan(11) P5 "A vmhost adapter for what DOES need SSH … built to the analyser-adapter README skeleton"; earlier plan(1)/(4)/(5) P5 "registry + relocation (service_instances) and, eventually, the chassis service-deployer adapter".
- **what:** The eventual automation for what genuinely needs SSH: provision box, run setup.sh, onboard domain, ship engine, decommission — built as a `vmhost` adapter (cmd/vmhost-adapter, internal/adapters/vmhost, reuse thunder's ssh via shared/, kustomize, KafkaTopic system.adapter.vmhost.requests, 003 envelope), with a `service_instances` registry modelled on thunder_instances minus the reaper. Long-term it holds the deploy SSH credential, retiring the repo-secrets copy. Earlier versions named this the "service-deployer" adapter.
- **sources:** traffic_probe_plan(11).md#p5, traffic_probe_plan(1).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** future handler for backend_unreachable; supersedes "service-deployer" naming
- **verify-later:** analyser-adapter README; thunder_instances → service_instances

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### adapter response-envelope contract (request/response wiring conventions)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** adapter(4).go's header states the envelope is "grounded in the orchestrator (coordinator.go) and the three working adapters, not the docs — which disagree (003 vs FOCUS) and were resolved empirically." internal/adapters/analyser/ exists live in the repo, confirming the draft shipped.
- **what:** A reverse-engineered (from working adapters, not from docs) contract for how an adapter must shape its Kafka request/response envelope so the orchestrator actually routes the reply instead of timing out: action comes from `body.action` not headers; `in_response_to_request_id` (echoing the incoming `request_id`) is the load-bearing claim field in `coordinator.go`'s `ProcessResponse`, with `request_id` as fallback; the reply body must use canonical `types.ResponseHeaders` via `ToResponseHeaders` so `is_complete`/`is_error` marshal as real JSON bools (the "bool trap" — a `map[string]string` sending the string `"true"` fails the receiver's struct-bool unmarshal); sends must go via `ProduceWithValidation`, never plain `Produce`. websearch-adapter is flagged as the one adapter still on the deprecated string-bool/plain-Produce path.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/adapter(4).go, docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/documentation_analysis_tool/analyser_request_action(1).go
- **relations:** contextkit toolchain (above); analyser-adapter service
- **verify-later:** platform/orchestration/coordinator.go ProcessResponse, internal/adapters/analyser/adapter.go, whether websearch-adapter has since been migrated off the string-bool map

<!-- SOURCE: U11_traffic_probe.md -->
### VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain; the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the probe project surfaced: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. Everything else was deliberately mapped onto existing mechanisms (adapter skeleton, thunder ssh, thunder_instances→service_instances, scheduled tasks, discovery checks, in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — the discovery agents scan live sites over HTTP regardless of hosting.
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** every concept below in this category; improvement-loop; adapters
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine — API-only capture backend for the class
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)".
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: POST /intent (capture + 303 to THANKS_PATH), GET /api/hit (visit beacon), GET /stats (key-gated summary), GET /health, GET /events (export), GET /access-digest (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with ACCEPT_HOSTS as optional defence-in-depth. Explicitly class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and per-domain content registry removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** engine store v2, /events endpoint, access-digest, setup.sh provisioning, idea.uk (fork origin)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

<!-- SOURCE: U11_traffic_probe.md -->
### Engine store v2 — daily JSONL events + debounced counters + on-box retention
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (relojistas_notes 2026-06-12 log).
- **what:** Two pre-launch storage cliffs were fixed structurally: v1 rewrote one ever-growing JSON file on every persist and held all events in RAM (superseded). v2 splits by access pattern — events append to daily JSONL (events-YYYYMMDD.jsonl, one line per submission, O(1) at any volume, rotation = the date, retention = delete old files); /stats counters live in a small counters.json flushed by a dirty-flag 5s debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. Retention enforced on-box by site-engine-prune.timer (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** deploy_setup/site-engine/store.go (header), traffic_probe_running_notes(28).md#2026-06-11 (store fix + store v2 + retention), relojistas_notes(8).md#decisions
- **relations:** /events export (tails these lines), intent event record, minimal-data privacy (90-day prune)
- **verify-later:** store.go in site-engine repo; systemctl list-timers site-engine-prune.timer on box

<!-- SOURCE: U11_traffic_probe.md -->
### Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1–D4)
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed the standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; sites.github_repo selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no needs_vm_deploy terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via deploy_config/service_instances only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved + #decision-1-4 analysis, traffic_probe_running_notes(28).md#2026-06-10 (decisions resolved)
- **relations:** vm-sites repo + Action, github_repo target selector, vmhost adapter (the later heavy path), development-guide (build pipeline)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

<!-- SOURCE: U11_traffic_probe.md -->
### vm-sites content repo and deploy-to-vm Action
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner."
- **what:** A standalone private repo (gqls/vm-sites; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested; the docs-tree copy is a reference snapshot only, contextkit pattern). Domain folders at repo ROOT — an assumption bug resolved 2026-06-11: the live sites repo keeps domain folders at root (the `sites/**` variant was a stale copy inside agentchassis/.git/workflows/, which GitHub never reads). The Action is a faithful sibling of the live B2 action: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks, rsync -az --delete over SSH into /var/www/vm-sites/<domain>; no CF purge; deploys content only for already-provisioned domains. Deletion-propagation gap shared with the B2 action — noted, not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11 (layout resolved; live b2 action learned), traffic_probe_runbook(13).md#3.1+5
- **relations:** deployment-github (B2 action sibling), setup.sh WEBROOT_OWNER, debugging lesson #24
- **verify-later:** gqls/vm-sites .github/workflows; Action run history

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine deploy Action and the narrow-sudo privilege model
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12 "the 3.9 engine-seam test now SHIPS the endpoint".
- **what:** On push of **.go/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook /usr/local/sbin/site-engine-deploy which atomically swaps the binary and restarts. Privilege model: no root key in CI; setup.sh (when DEPLOY_USER set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary runs as the unprivileged site-engine user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds GOARCH=amd64 (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10 (engine-deploy workflow + privilege model), traffic_probe_runbook(13).md#5
- **relations:** setup.sh (installs the hook), dedicated-vs-shared box policy (x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine

<!-- SOURCE: U11_traffic_probe.md -->
### setup.sh — idempotent multi-vhost box provisioning
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; add a domain = extend DOMAINS + re-run; existing domains untouched), parameterised, self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving /var/www/vm-sites/<domain> and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; deploy sudo hook; prune timer; MODE=full|update. Grown options: WEBROOT_OWNER (deploy-user rsync rights), WWW_ALIAS (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), CLOUDFLARE=true (CF real_ip conf), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2 deprecation found in the field). Warning captured: box-takeover semantics (ufw --force reset, removes nginx default site) — why sharing the idea.uk box was declined.
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10 (box artifact) + 2026-06-12 entries, traffic_probe_runbook(13).md#3.5+4
- **relations:** site-engine deploy hook, multi-domain multiplexing, vmhost adapter (automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

<!-- SOURCE: U11_traffic_probe.md -->
### Multi-domain single-binary hosting and domain onboarding/relocation
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx server_name blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend DOMAINS + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: THANKS_PATH is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standard /thanks.html, each domain shipping its own; relojistas keeps /gracias.html on its dedicated box.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13 (THANKS_PATH design point)
- **relations:** setup.sh, dedicated-vs-shared box policy, vmhost adapter (onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

<!-- SOURCE: U11_traffic_probe.md -->
### Dedicated vs shared box policy and VM sizing
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13).
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling for a ~€3.49/mo saving). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build). Policy hardened 2026-06-13: use EXISTING boxes only for new domains.
- **sources:** relojistas_notes(8).md#decisions+provenance (coordinates), traffic_probe_running_notes(28).md#2026-06-11 (sizing, bandwidth, box question), HANDOFF#where-things-stand
- **relations:** setup.sh takeover semantics, engine deploy Action x86 constraint
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

<!-- SOURCE: U11_traffic_probe.md -->
### Pull-not-push off-cluster data return
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled.
- **what:** The serving box only buffers (daily JSONL); the CLUSTER pulls over key-gated HTTPS on a schedule into clients_db. Rationale: pull keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or middle VM inverts that, adds an attack surface and a hop for no gain. B2 remains optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4). SSH is reserved for provisioning (P5).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11 (no collector VM; integration mapping)
- **relations:** /events endpoint, intent collection topology, vmhost adapter (the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

<!-- SOURCE: U11_traffic_probe.md -->
### requires-backend capability gate (Decision 5)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; component-side tag live (component inserted 2026-06-11); planner gate and audit check not applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): load_components gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap section_types only. Site side: deploy_config || {"target":"vm","capabilities":["backend"]} at onboarding. Later: an audit check comparing placed sections' requires-* tags against site capabilities → site_work_items findings. Supersedes the first design (an invented `intent-probe` site type in suitable_site_types + a suitable_site_types='[]' planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql (family-delta: the superseded layer-1 gate), traffic_probe_running_notes(28).md#2026-06-10 (naming correction)
- **relations:** intent-probe component, site-plan-and-reconciler (build-site-planner load_components), design-composition
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

<!-- SOURCE: U11_traffic_probe.md -->
### P5 vmhost provisioning adapter and service_instances registry
- **category:** NEW:vm-backend-sites
- **status-signal:** aspirational
- **status-evidence:** plan P5 is entirely future-tense; HANDOFF Thread B lists it as pending; "P5 — registry + provisioning adapter" never marked started.
- **what:** The SSH half of the class, automating what runbook §3 does by hand: a `vmhost` adapter (analyser-adapter README skeleton: cmd/vmhost-adapter, internal/adapters/vmhost/ reusing thunder's ssh package via the shared/ precedent, configs, dockerfile, kustomize overlays, Makefile ×4, KafkaTopic system.adapter.vmhost.requests, 003 envelope contract) for provision-box / run setup.sh / onboard-domain (extend DOMAINS + re-run) / ship engine / decommission. Tracked in a `service_instances` table modelled on thunder_instances MINUS the reaper/uptime cap (persistent boxes are never reaped). Thin request actions + a deployer-family agent. Long-term the adapter holds the deploy SSH credential, retiring the repo-secrets copy.
- **sources:** traffic_probe_plan(12).md#P5 + #framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** adapters (thunder precedent, 003 envelope), setup.sh (what it automates), backend_unreachable (future handler)
- **verify-later:** any vmhost-adapter code/kustomize; service_instances table existence

<!-- SOURCE: U11_traffic_probe.md -->
### Cloudflare-proxied-in-front option
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h …)"; the real_ip conf ("set CLOUDFLARE=true on its next setup.sh re-run") still pending at last entry.
- **what:** Optional per-domain posture: keep DNS on Cloudflare with a proxied record → VM origin. Explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem — avoid); the VM stays the single source of truth, CF just caches. Adjustments: cache-bypass the API paths; nginx set_real_ip_from CF ranges + real_ip_header CF-Connecting-IP (else rate-limiting throttles all of CF as one client and logs/digest/fail2ban see CF IPs); TLS Full (strict). Bonuses: CF-IPCountry populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound.
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10 (CF clarification) + 2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note
- **relations:** access-digest (real-IP dependency), setup.sh CLOUDFLARE param, multi-domain relocation
- **verify-later:** relojistas CF zone config; cloudflare-realip.conf on box

<!-- SOURCE: U11_traffic_probe.md -->
### VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain; the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the probe project surfaced: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. Everything else was deliberately mapped onto existing mechanisms (adapter skeleton, thunder ssh, thunder_instances→service_instances, scheduled tasks, discovery checks, in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — the discovery agents scan live sites over HTTP regardless of hosting.
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** every concept below in this category; improvement-loop; adapters
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine — API-only capture backend for the class
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)".
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: POST /intent (capture + 303 to THANKS_PATH), GET /api/hit (visit beacon), GET /stats (key-gated summary), GET /health, GET /events (export), GET /access-digest (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with ACCEPT_HOSTS as optional defence-in-depth. Explicitly class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and per-domain content registry removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** engine store v2, /events endpoint, access-digest, setup.sh provisioning, idea.uk (fork origin)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

<!-- SOURCE: U11_traffic_probe.md -->
### Engine store v2 — daily JSONL events + debounced counters + on-box retention
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (relojistas_notes 2026-06-12 log).
- **what:** Two pre-launch storage cliffs were fixed structurally: v1 rewrote one ever-growing JSON file on every persist and held all events in RAM (superseded). v2 splits by access pattern — events append to daily JSONL (events-YYYYMMDD.jsonl, one line per submission, O(1) at any volume, rotation = the date, retention = delete old files); /stats counters live in a small counters.json flushed by a dirty-flag 5s debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. Retention enforced on-box by site-engine-prune.timer (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** deploy_setup/site-engine/store.go (header), traffic_probe_running_notes(28).md#2026-06-11 (store fix + store v2 + retention), relojistas_notes(8).md#decisions
- **relations:** /events export (tails these lines), intent event record, minimal-data privacy (90-day prune)
- **verify-later:** store.go in site-engine repo; systemctl list-timers site-engine-prune.timer on box

<!-- SOURCE: U11_traffic_probe.md -->
### Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1–D4)
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed the standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; sites.github_repo selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no needs_vm_deploy terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via deploy_config/service_instances only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved + #decision-1-4 analysis, traffic_probe_running_notes(28).md#2026-06-10 (decisions resolved)
- **relations:** vm-sites repo + Action, github_repo target selector, vmhost adapter (the later heavy path), development-guide (build pipeline)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

<!-- SOURCE: U11_traffic_probe.md -->
### vm-sites content repo and deploy-to-vm Action
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner."
- **what:** A standalone private repo (gqls/vm-sites; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested; the docs-tree copy is a reference snapshot only, contextkit pattern). Domain folders at repo ROOT — an assumption bug resolved 2026-06-11: the live sites repo keeps domain folders at root (the `sites/**` variant was a stale copy inside agentchassis/.git/workflows/, which GitHub never reads). The Action is a faithful sibling of the live B2 action: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks, rsync -az --delete over SSH into /var/www/vm-sites/<domain>; no CF purge; deploys content only for already-provisioned domains. Deletion-propagation gap shared with the B2 action — noted, not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11 (layout resolved; live b2 action learned), traffic_probe_runbook(13).md#3.1+5
- **relations:** deployment-github (B2 action sibling), setup.sh WEBROOT_OWNER, debugging lesson #24
- **verify-later:** gqls/vm-sites .github/workflows; Action run history

<!-- SOURCE: U11_traffic_probe.md -->
### site-engine deploy Action and the narrow-sudo privilege model
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12 "the 3.9 engine-seam test now SHIPS the endpoint".
- **what:** On push of **.go/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook /usr/local/sbin/site-engine-deploy which atomically swaps the binary and restarts. Privilege model: no root key in CI; setup.sh (when DEPLOY_USER set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary runs as the unprivileged site-engine user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds GOARCH=amd64 (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10 (engine-deploy workflow + privilege model), traffic_probe_runbook(13).md#5
- **relations:** setup.sh (installs the hook), dedicated-vs-shared box policy (x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine

<!-- SOURCE: U11_traffic_probe.md -->
### setup.sh — idempotent multi-vhost box provisioning
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; add a domain = extend DOMAINS + re-run; existing domains untouched), parameterised, self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving /var/www/vm-sites/<domain> and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; deploy sudo hook; prune timer; MODE=full|update. Grown options: WEBROOT_OWNER (deploy-user rsync rights), WWW_ALIAS (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), CLOUDFLARE=true (CF real_ip conf), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2 deprecation found in the field). Warning captured: box-takeover semantics (ufw --force reset, removes nginx default site) — why sharing the idea.uk box was declined.
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10 (box artifact) + 2026-06-12 entries, traffic_probe_runbook(13).md#3.5+4
- **relations:** site-engine deploy hook, multi-domain multiplexing, vmhost adapter (automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

<!-- SOURCE: U11_traffic_probe.md -->
### Multi-domain single-binary hosting and domain onboarding/relocation
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx server_name blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend DOMAINS + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: THANKS_PATH is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standard /thanks.html, each domain shipping its own; relojistas keeps /gracias.html on its dedicated box.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13 (THANKS_PATH design point)
- **relations:** setup.sh, dedicated-vs-shared box policy, vmhost adapter (onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

<!-- SOURCE: U11_traffic_probe.md -->
### Dedicated vs shared box policy and VM sizing
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13).
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling for a ~€3.49/mo saving). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build). Policy hardened 2026-06-13: use EXISTING boxes only for new domains.
- **sources:** relojistas_notes(8).md#decisions+provenance (coordinates), traffic_probe_running_notes(28).md#2026-06-11 (sizing, bandwidth, box question), HANDOFF#where-things-stand
- **relations:** setup.sh takeover semantics, engine deploy Action x86 constraint
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

<!-- SOURCE: U11_traffic_probe.md -->
### Pull-not-push off-cluster data return
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled.
- **what:** The serving box only buffers (daily JSONL); the CLUSTER pulls over key-gated HTTPS on a schedule into clients_db. Rationale: pull keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or middle VM inverts that, adds an attack surface and a hop for no gain. B2 remains optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4). SSH is reserved for provisioning (P5).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11 (no collector VM; integration mapping)
- **relations:** /events endpoint, intent collection topology, vmhost adapter (the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

<!-- SOURCE: U11_traffic_probe.md -->
### requires-backend capability gate (Decision 5)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; component-side tag live (component inserted 2026-06-11); planner gate and audit check not applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): load_components gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap section_types only. Site side: deploy_config || {"target":"vm","capabilities":["backend"]} at onboarding. Later: an audit check comparing placed sections' requires-* tags against site capabilities → site_work_items findings. Supersedes the first design (an invented `intent-probe` site type in suitable_site_types + a suitable_site_types='[]' planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql (family-delta: the superseded layer-1 gate), traffic_probe_running_notes(28).md#2026-06-10 (naming correction)
- **relations:** intent-probe component, site-plan-and-reconciler (build-site-planner load_components), design-composition
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

<!-- SOURCE: U11_traffic_probe.md -->
### P5 vmhost provisioning adapter and service_instances registry
- **category:** NEW:vm-backend-sites
- **status-signal:** aspirational
- **status-evidence:** plan P5 is entirely future-tense; HANDOFF Thread B lists it as pending; "P5 — registry + provisioning adapter" never marked started.
- **what:** The SSH half of the class, automating what runbook §3 does by hand: a `vmhost` adapter (analyser-adapter README skeleton: cmd/vmhost-adapter, internal/adapters/vmhost/ reusing thunder's ssh package via the shared/ precedent, configs, dockerfile, kustomize overlays, Makefile ×4, KafkaTopic system.adapter.vmhost.requests, 003 envelope contract) for provision-box / run setup.sh / onboard-domain (extend DOMAINS + re-run) / ship engine / decommission. Tracked in a `service_instances` table modelled on thunder_instances MINUS the reaper/uptime cap (persistent boxes are never reaped). Thin request actions + a deployer-family agent. Long-term the adapter holds the deploy SSH credential, retiring the repo-secrets copy.
- **sources:** traffic_probe_plan(12).md#P5 + #framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** adapters (thunder precedent, 003 envelope), setup.sh (what it automates), backend_unreachable (future handler)
- **verify-later:** any vmhost-adapter code/kustomize; service_instances table existence

<!-- SOURCE: U11_traffic_probe.md -->
### Cloudflare-proxied-in-front option
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h …)"; the real_ip conf ("set CLOUDFLARE=true on its next setup.sh re-run") still pending at last entry.
- **what:** Optional per-domain posture: keep DNS on Cloudflare with a proxied record → VM origin. Explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem — avoid); the VM stays the single source of truth, CF just caches. Adjustments: cache-bypass the API paths; nginx set_real_ip_from CF ranges + real_ip_header CF-Connecting-IP (else rate-limiting throttles all of CF as one client and logs/digest/fail2ban see CF IPs); TLS Full (strict). Bonuses: CF-IPCountry populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound.
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10 (CF clarification) + 2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note
- **relations:** access-digest (real-IP dependency), setup.sh CLOUDFLARE param, multi-domain relocation
- **verify-later:** relojistas CF zone config; cloudflare-realip.conf on box

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods)
- **what:** All storage-touching actions construct their own client via storage.NewS3Client with env-var names in ObjectStorageConfig (B2_APPLICATION_KEY_ID/KEY from personae-platform-secrets); injected params.StorageClient is unreliable (nil when IMAGE_BUCKET absent at startup). Spawn-time env forwarding (Path C) is gated by isStorageEnabledAgent/orchestrator/code-driven — keep the gate maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 full
- **relations:** spawn env propagation; thunder presigned URLs
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

<!-- SOURCE: U06_finetuning.md -->
### Hostile-VM threat model for the training data plane
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14 "the VM holds no B2 credentials, just a time-limited URL".
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives: standing scoped B2 key on the box (prefix-wide bearer leak risk) and per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** presigned data plane; eval gate; storage credential architecture decision
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

<!-- SOURCE: U06_finetuning.md -->
### Presigned-URL data plane (adapter mints URLs; bytes never transit Kafka)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23".
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (note: `finetuning/` is a folder prefix, not a bucket; the preparer agent-def's `s3_bucket=finetuning` is stale/logical and cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (they now delegate to ObjectURL — one signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&` as &; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh
- **relations:** hostile-VM threat model; storage credential decision; batch presign; docubundle B2 notes
- **verify-later:** data_url_actions.go; TRAINING_BUCKET env; personae-storage-secrets wiring

<!-- SOURCE: U06_finetuning.md -->
### Storage credential architecture decision (no storage-adapter service)
- **category:** storage-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "Decision (2026-05-22): hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (max.message.bytes ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs. The acknowledged mess: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix: one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared storage.Client interface forced rebuilding every binary importing platform/storage.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture
- **relations:** presigned data plane; adapters category; debugging guide item 18
- **verify-later:** whether NewDefaultClient exists; secret sourcing per service

<!-- SOURCE: U18_sql_for_agents.md -->
### asset-deployer (S3 → optimize-by-purpose → git)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping deploy_image_asset: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler, undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Presigned-URL storage boundary (prepare_object_url / dataset / artefact)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §5a "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049 … signing against personae-model-training" (2026-06-02)
- **what:** The adapter mints presigned B2 URLs but never moves data — only the few-hundred-byte URL travels over Kafka, the actual bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h; DatasetURL/ArtefactURL delegate to it). Bucket is `personae-model-training` (the preparer's `s3_bucket=finetuning` agent_def value is stale/logical); B2 keys live in `personae-storage-secrets`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** enables the checkpoint/artefact upload; contrasted with rejected storage-adapter; verification gotcha: presigned GET fails HEAD (curl -I) with 403
- **verify-later:** internal/adapters/thunder/data_url_actions.go (ObjectURL/handlePrepareObjectURL); storage.Client.GetPresignedPutURL

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Storage-credential architecture decision (no storage-adapter; presign not blob-pipe)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §14 "Storage credential architecture — decision (2026-05-22) … do NOT build a storage-adapter service"
- **what:** A decision that routing multi-MB datasets / hundreds-of-MB artefacts through a storage-adapter over Kafka is a real failure (max.message.bytes ~1MB), so a storage-adapter is only safe for minting URLs, dangerous for moving data. Interim: hardcode the adapter's B2 env. Eventual (deferred, not built): one shared `storage.NewDefaultClient` reading `personae-storage-secrets` everywhere to kill the four-conventions credential mess; untrusted agents get presigned URLs, never a blob pipe.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Storage credential architecture)
- **relations:** justifies the presigned-URL boundary; GetPresignedPutURL added to storage.Client forced rebuild of every binary importing platform/storage
- **verify-later:** platform/storage/interface.go; personae-storage-secrets

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** PLAN(4) "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left"
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Threat model assumes the VM is hostile (holds no B2 key, only single-object URLs); a standing scoped key and per-save callback endpoint were both rejected. Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** makes the monitor's DONE_OK→decommission safe; still-LEFT Phase D3 dispatch_thunder_prepare_resume_url + migration for check_resume; corrects the "list-keys is the ONE adapter gap" claim
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … v2: events append to daily JSONL … O(1) at any volume"; burst-tested.
- **what:** v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff). v1 added a dirty-flag + 5s background flusher (AddVisit no longer persists per call; AddEvent still immediate). v2 replaced the monolithic file: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, bounded RAM), /stats counters live in a small `counters.json`; SIGTERM fsyncs. Removed the in-RAM `Store.Events` map and uncalled `Store.Snapshot()`.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action, traffic_probe_running_notes(27).md#2026-06-11-store-v2, deploy_setup/working_dir/main.go#header
- **relations:** abandoned Store.Events/Snapshot; drove the ENGINE_DATA_DIR rename
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Persistence design — tiered one-way data flow for exposed services (box → B2 → chassis)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task."
- **what:** A security-motivated pattern for any internet-facing satellite service (idea.uk being the first case) to get its data into the core chassis DB without opening an inbound path: (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reuses the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven ingest agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`), "chassis PULLS; box never connects in." Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only).
- **sources:** `running_notes(44).md` (`PERSISTENCE_design.md` summary, two checkpoints on 2026-06-04)
- **relations:** service-deployer pattern; Thunder adapter (B2 presigned-URL precedent); storage-architecture (032, S3/B2)
- **verify-later:** whether `business_intel.idea_orders` / `ecommerce.orders` / an `idea-ingest` scheduled task exist

<!-- SOURCE: U26_misc_dirs.md -->
### Result storage split (DB paper-trail + S3 artefacts)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** basic_usage/002 states it as fact: final_result column in orchestrator_state, a website_projects table per client schema with preview/live URLs, and site-publisher's s3_upload of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated final_result JSON + website_projects metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product".
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system (site-publisher s3_upload)
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods)
- **what:** All storage-touching actions construct their own client via storage.NewS3Client with env-var names in ObjectStorageConfig (B2_APPLICATION_KEY_ID/KEY from personae-platform-secrets); injected params.StorageClient is unreliable (nil when IMAGE_BUCKET absent at startup). Spawn-time env forwarding (Path C) is gated by isStorageEnabledAgent/orchestrator/code-driven — keep the gate maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 full
- **relations:** spawn env propagation; thunder presigned URLs
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

<!-- SOURCE: U06_finetuning.md -->
### Hostile-VM threat model for the training data plane
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14 "the VM holds no B2 credentials, just a time-limited URL".
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives: standing scoped B2 key on the box (prefix-wide bearer leak risk) and per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** presigned data plane; eval gate; storage credential architecture decision
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

<!-- SOURCE: U06_finetuning.md -->
### Presigned-URL data plane (adapter mints URLs; bytes never transit Kafka)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23".
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (note: `finetuning/` is a folder prefix, not a bucket; the preparer agent-def's `s3_bucket=finetuning` is stale/logical and cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (they now delegate to ObjectURL — one signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&` as &; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh
- **relations:** hostile-VM threat model; storage credential decision; batch presign; docubundle B2 notes
- **verify-later:** data_url_actions.go; TRAINING_BUCKET env; personae-storage-secrets wiring

<!-- SOURCE: U06_finetuning.md -->
### Storage credential architecture decision (no storage-adapter service)
- **category:** storage-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "Decision (2026-05-22): hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (max.message.bytes ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs. The acknowledged mess: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix: one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared storage.Client interface forced rebuilding every binary importing platform/storage.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture
- **relations:** presigned data plane; adapters category; debugging guide item 18
- **verify-later:** whether NewDefaultClient exists; secret sourcing per service

<!-- SOURCE: U18_sql_for_agents.md -->
### asset-deployer (S3 → optimize-by-purpose → git)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping deploy_image_asset: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler, undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Presigned-URL storage boundary (prepare_object_url / dataset / artefact)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §5a "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049 … signing against personae-model-training" (2026-06-02)
- **what:** The adapter mints presigned B2 URLs but never moves data — only the few-hundred-byte URL travels over Kafka, the actual bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h; DatasetURL/ArtefactURL delegate to it). Bucket is `personae-model-training` (the preparer's `s3_bucket=finetuning` agent_def value is stale/logical); B2 keys live in `personae-storage-secrets`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** enables the checkpoint/artefact upload; contrasted with rejected storage-adapter; verification gotcha: presigned GET fails HEAD (curl -I) with 403
- **verify-later:** internal/adapters/thunder/data_url_actions.go (ObjectURL/handlePrepareObjectURL); storage.Client.GetPresignedPutURL

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Storage-credential architecture decision (no storage-adapter; presign not blob-pipe)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §14 "Storage credential architecture — decision (2026-05-22) … do NOT build a storage-adapter service"
- **what:** A decision that routing multi-MB datasets / hundreds-of-MB artefacts through a storage-adapter over Kafka is a real failure (max.message.bytes ~1MB), so a storage-adapter is only safe for minting URLs, dangerous for moving data. Interim: hardcode the adapter's B2 env. Eventual (deferred, not built): one shared `storage.NewDefaultClient` reading `personae-storage-secrets` everywhere to kill the four-conventions credential mess; untrusted agents get presigned URLs, never a blob pipe.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Storage credential architecture)
- **relations:** justifies the presigned-URL boundary; GetPresignedPutURL added to storage.Client forced rebuild of every binary importing platform/storage
- **verify-later:** platform/storage/interface.go; personae-storage-secrets

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** PLAN(4) "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left"
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Threat model assumes the VM is hostile (holds no B2 key, only single-object URLs); a standing scoped key and per-save callback endpoint were both rejected. Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** makes the monitor's DONE_OK→decommission safe; still-LEFT Phase D3 dispatch_thunder_prepare_resume_url + migration for check_resume; corrects the "list-keys is the ONE adapter gap" claim
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … v2: events append to daily JSONL … O(1) at any volume"; burst-tested.
- **what:** v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff). v1 added a dirty-flag + 5s background flusher (AddVisit no longer persists per call; AddEvent still immediate). v2 replaced the monolithic file: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, bounded RAM), /stats counters live in a small `counters.json`; SIGTERM fsyncs. Removed the in-RAM `Store.Events` map and uncalled `Store.Snapshot()`.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action, traffic_probe_running_notes(27).md#2026-06-11-store-v2, deploy_setup/working_dir/main.go#header
- **relations:** abandoned Store.Events/Snapshot; drove the ENGINE_DATA_DIR rename
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Persistence design — tiered one-way data flow for exposed services (box → B2 → chassis)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task."
- **what:** A security-motivated pattern for any internet-facing satellite service (idea.uk being the first case) to get its data into the core chassis DB without opening an inbound path: (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reuses the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven ingest agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`), "chassis PULLS; box never connects in." Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only).
- **sources:** `running_notes(44).md` (`PERSISTENCE_design.md` summary, two checkpoints on 2026-06-04)
- **relations:** service-deployer pattern; Thunder adapter (B2 presigned-URL precedent); storage-architecture (032, S3/B2)
- **verify-later:** whether `business_intel.idea_orders` / `ecommerce.orders` / an `idea-ingest` scheduled task exist

<!-- SOURCE: U26_misc_dirs.md -->
### Result storage split (DB paper-trail + S3 artefacts)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** basic_usage/002 states it as fact: final_result column in orchestrator_state, a website_projects table per client schema with preview/live URLs, and site-publisher's s3_upload of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated final_result JSON + website_projects metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product".
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system (site-publisher s3_upload)
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Commit IS deploy (git → GitHub Actions → Backblaze B2, Cloudflare DNS-only)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decision 15; 034 shows the actual workflow
- **what:** Individual commits per work item; GH Actions fires per commit on a self-hosted runner, detects changed root-level domain directories, `b2 sync --delete --skip-newer` each to `b2://portfolio-sites/<domain>`, then purges Cloudflare cache per zone. No separate deploy step. The authoritative workflow lives in gqls/sites/.github/workflows — a stray copy under .git/workflows is a documented trap.
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; 016 §0 item 24
- **relations:** git-adapter; "git committed is not proof of new content"
- **verify-later:** gqls/sites .github workflow; B2 bucket layout

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chassis and site deploy model (single IMAGE_TAG; git → Actions → B2)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** README_001_flow_notes: "There's a single global IMAGE_TAG (currently v1.0.1066) and one agent-chassis binary that runs every dynamic agent … Rollback is symmetric and cheap"; checkpoint (uu) repeats the make targets.
- **what:** Two deploy surfaces: (1) chassis code ships as one image tag running every dynamic agent via agent_definitions config — targeted path `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; full `make release` bumps every service; rollback repoints to the old existing image without rebuild. (2) Site content deploys git → GitHub Actions → Backblaze B2. Operational wrinkle: full-build deploys commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** README_001_flow_notes.md; running_notes_checkpoint_uu.md#Deploy-rollback; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; running_notes_scheme_to_components(55).md#Th (hygiene)
- **relations:** deployed-binary-predates-disk; agent re-registration.
- **verify-later:** Makefile targets; agent_definitions.image_tag column use.

<!-- SOURCE: U05_content_quality_linking.md -->
### Git per-page deploy + non-fast-forward race
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes_14(26) Part 10: git-adapter "updateRef is force:false + no-retry, so a concurrent commit to the shared sites repo loses with a silent non-fast-forward (FOCUS_dispatch open item 4)".
- **what:** Pages deploy as one git commit each to the shared sites repo (gqls/sites, git→Cloudflare); concurrent commits during a cascade can silently lose a non-fast-forward push. Suspected (not confirmed) in the missing-homepage case before the auto-complete cause was pinned; remains an open reliability item. Minor cosmetic sibling: page-rerender's commit message template "Rerender: {{.filename}}" renders uninterpolated.
- **sources:** running_notes_14(26).md#part-10; NOTES(44) minor findings
- **relations:** deploy gate on sections_saved; dispatch throughput.
- **verify-later:** git-adapter updateRef retry behaviour; commit-message templating scope.

<!-- SOURCE: U09_adoption.md -->
### Git-adapter cross-site commit race (force:false, no retry, shared sites repo)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** "a latent bug today for concurrent multi-site builds… Fix: retry-on-non-fast-forward in updateRef… tracked as a guardrail" ; "(2026-06-04: the missing-homepage lead was NOT this race… remains a latent risk with no confirmed instance yet.)"
- **what:** `git_commit` publishes to the git-adapter (Kafka key = correlation_id), which does a GitHub Git Data API read-modify-write with `force:false` and no retry against a single shared `sites` repo. Same-site commits serialize via partition (git-safe for same-site parallelism); different sites can hit different replicas concurrently and the loser's deploy fails silently on non-fast-forward. Proposed optimistic-concurrency retry (re-read HEAD, rebuild tree, re-commit) not yet built.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#git-deploy-path
- **relations:** Lever A prerequisite; A4 homepage (initially suspected, exonerated)
- **verify-later:** git-adapter `github_client.go` CommitToRepo/updateRef; adapter.go consumer model (2 replicas, group git.adapter.group)

<!-- SOURCE: U11_traffic_probe.md -->
### sites.github_repo as deploy-target selector (resolveGitRepoName patch)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread B: "Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all"; plan P3 "Remaining: land the chassis patch".
- **what:** Tracing showed sites.github_repo is DORMANT end-to-end: git_commit reads config repo_name → default "sites"; upsertSite doesn't SELECT it; ensure_site_record doesn't return it. Specified 3-touch patch: (1) upsertSite RETURNING += COALESCE(github_repo,''), (2) EnsureSiteRecordAction return map += github_repo, (3) a private resolveGitRepoName(config, collected) helper (config repo_name → site_record.github_repo → "sites") used by BOTH git_commit and deploy_image_asset — the latter hardcodes "sites" at line 463 and would otherwise split-brain a probe site (pages → VM repo, images → sites). vet_med_export left alone (dedicated pipeline). Pre-flight confirmed github_repo empty on all 8 sites, so the fallback is safe. CommitToRepo already prefixes <domain>/ for any repo (shared root layout confirmed); createOrGetRepo auto-creates missing repos as PUBLIC — a deliberate-visibility trap.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_plan(12).md#P3, HANDOFF#thread-b
- **relations:** vm-sites repo, D1–D4, requires-backend gate (the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution; whether the patch ever landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Deployment-GitHub / self-hosted runner + deploy path
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 102_blog_handoff "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace"
- **what:** The publish path: agents commit generated site files via a git adapter; a self-hosted GitHub Actions runner runs the sites-repo workflow which `b2 sync`s to Backblaze. `needs_rerender` is the terminal build item that assembles pages and triggers deployment.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, WM/001_development_guide(0).md#every-pipeline-must-end-with-assembly-and-deployment, WM/007_adoption_pipeline_v3.md#data-flow-between-layers
- **relations:** blog-listing handoff; storage architecture (B2/S3); site plan reconciler terminal items
- **verify-later:** git-adapter; github-actions-runner dockerfile; needs_rerender handler

<!-- SOURCE: U20_legacy_docs_a.md -->
### Git deployment: commit_to_git + GitHub Action sync to B2
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** webbuild_pipeline/001pipeline: "Deployer commits sites/boxing-tickets.com/index.html to GitHub → GitHub Action automatically syncs that folder to B2 → Site is live."
- **what:** Deployment path: a git-adapter microservice (Kafka topic system.adapter.git.requests) commits generated site files to a repo (per-domain repos in the original design; a sites/<domain>/ folder in practice); a GitHub Action syncs to Backblaze B2 which serves the live site. GitCommitAction is the workflow-side action.
- **sources:** docs004_website_capture_project/webbuild_pipeline/001pipeline; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** deployment-github (docs 034 live successor: git-adapter deploy surface); storage-architecture B2.
- **verify-later:** git-adapter service; the GitHub Action workflow file; sites/ repo layout.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site manifest + external-edit desynchronisation detection
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** Design tables in 004 (git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) and manifest.json "winning genes" tracking in 008/014; no implementation evidence.
- **what:** Every generated site carries a manifest.json recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter watches all site repos; a human commit flags the manifest desynchronised, halting agent edit workflows and queueing HITL review — protecting human work from being overwritten and agents from stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism protecting human edits); deployment-github git-adapter.
- **verify-later:** any manifest.json in site repos; git webhook receivers.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Standardized per-page git deployment
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** docs017/002_standardising problem table ("Inconsistent deploy patterns → Standardize to per-page commits"); docs017/023 "Individual git commits per page → each goes live via GitHub Action → S3"; work items store commit_sha in result.
- **what:** Deployment converges on one mechanism: each page is committed individually via git_commit (with file_path override enabling CSS/asset commits), GitHub Actions deploy to hosting (Cloudflare, later S3), and commit SHAs are recorded on pages and work items for traceability. Removed redundant deployer steps whose data paths kept breaking.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy
- **relations:** git-adapter; deployment-github category; per-page loop.
- **verify-later:** git_commit action file_path config; GitHub Action workflows in site repos.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sites deployment chain (git → GitHub Actions → Backblaze B2) + image-tag chassis deploys
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** Used throughout ("after Actions propagates (a few minutes)"); 016b orientation: "Deployment is image-tag based... each agent's image_tag is bumped to adopt it; workflow (default_config) changes are DB-only and take effect immediately."
- **what:** Everything site-facing reaches production via git_commit to the 'sites' repo (files map keyed by repo-relative path — pages, tools/assets/*.js, assets/js/snippets.js, data/*.json) → GitHub Actions → Backblaze B2 (public), with the long-running git-adapter handling commits. Platform code ships as a chassis image (GitHub → Actions → image) adopted by bumping per-agent image_tag — so a source revert only reaches the cluster on the next build/push, while agent workflow changes (agent_definitions.default_config) are DB-only and instant. B2 404 NoSuchKey is the "page never deployed" signature.
- **sources:** docs/016b_debugging_guide_merged(3).md#orientation; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 (git_commit pattern); docs/RUNBOOK_phase2_provocation_js(29).md#4.2-4.3; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2
- **relations:** system-architecture; plan-storage revert note (pods keep old behaviour until push)
- **verify-later:** git-adapter; sites repo Actions workflow

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### github_repo target selection + resolveGitRepoName patch
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-10 "sites.github_repo is DORMANT end-to-end … The patch (guide's 'small patch' pattern)"; plan(11) P3 "Remaining … land the chassis patch (resolveGitRepoName helper …)".
- **what:** A site's `sites.github_repo` selects deploy target (vm-sites repo vs default "sites"), but was dormant (upsertSite didn't SELECT it, nothing read it). The fix: one `resolveGitRepoName(config, collected)` helper (config repo_name → site_record.github_repo → "sites") used by both `git_commit` and `deploy_image_asset`, plus upsertSite RETURNING + ensure_site_record map additions. `deploy_image_asset` hardcoded "sites" and would split-brain a probe site (pages→VM, logo/hero→sites) without the same fallback.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_plan(11).md#p3
- **relations:** enables P3 pipeline wiring; deploy_image_asset split-brain risk
- **verify-later:** git_deployer_actions.go, site_db_actions.go, upsertSite, EnsureSiteRecordAction, deploy_image_asset:463

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### "Commit is deploy" seam swapped B2→VM + two GitHub Actions
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** runbook(12) §5 "Two separate workflows"; running_notes 2026-06-11 both siblings "rewritten as faithful siblings … Validated".
- **what:** The static "commit is deploy" seam is preserved, only the destination moves. Content Action (`deploy-to-vm.yml` in vm-sites repo): on push, rsync -az --delete each changed root-level `<domain>/` over SSH to `/var/www/vm-sites/<domain>`; self-hosted runner, no CF purge. Engine Action (`deploy-engine-to-vm.yml` in site-engine repo): on push to `**.go`/go.mod, build static stripped linux/amd64, scp, run the narrow `site-engine-deploy` sudo hook (atomic swap + restart). Secrets VM_HOST/VM_USER/VM_SSH_KEY.
- **sources:** traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action, traffic_probe_running_notes(27).md#2026-06-11-live-b2-action
- **relations:** mirrors live deploy-to-b2.yml + Cloudflare Worker; target-agnostic terminal build item
- **verify-later:** vm-sites/.github/workflows/deploy-to-vm.yml, site-engine/.github/workflows/deploy-engine-to-vm.yml

<!-- SOURCE: U25_leopardess_social.md -->
### Chassis build/deploy practice (local Makefile builds, verify against the pod)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby build practice (2026-07-10): "images are built from the local filesystem via the Makefile; commits are at the user's discretion … Do not infer a deployed binary's contents from git history; verify against the running pod."
- **what:** The chassis deployment reality: images build from the local working tree, decoupled from git commits (image tag hand-recorded in commit messages), so the deployed binary can lead or lag the repo; verification is kubectl exec + grep -ac for symbols in /app/agent-chassis. Corollary for site work: a committed Go change (e.g. the A6 imagery routing) is inert until a Makefile build+push. Pod logs are ephemeral across rollouts; spawned agents log in their own pods, not agent-chassis.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 build practice, #2 log-hunting; docs/leopardessconsulting/HANDOFF.md#2-A6; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening)
- **relations:** imagery kind routing (A6 awaiting deploy); operator discipline
- **verify-later:** Makefile build targets; deploy pipeline for the chassis image

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Commit IS deploy (git → GitHub Actions → Backblaze B2, Cloudflare DNS-only)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decision 15; 034 shows the actual workflow
- **what:** Individual commits per work item; GH Actions fires per commit on a self-hosted runner, detects changed root-level domain directories, `b2 sync --delete --skip-newer` each to `b2://portfolio-sites/<domain>`, then purges Cloudflare cache per zone. No separate deploy step. The authoritative workflow lives in gqls/sites/.github/workflows — a stray copy under .git/workflows is a documented trap.
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; 016 §0 item 24
- **relations:** git-adapter; "git committed is not proof of new content"
- **verify-later:** gqls/sites .github workflow; B2 bucket layout

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Chassis and site deploy model (single IMAGE_TAG; git → Actions → B2)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** README_001_flow_notes: "There's a single global IMAGE_TAG (currently v1.0.1066) and one agent-chassis binary that runs every dynamic agent … Rollback is symmetric and cheap"; checkpoint (uu) repeats the make targets.
- **what:** Two deploy surfaces: (1) chassis code ships as one image tag running every dynamic agent via agent_definitions config — targeted path `make quick-agent-update IMAGE_TAG=…` (build → push → kustomize → DB image_tag → restart agent-chassis) plus `make update-and-restart-orchestrator` for the generic-orchestrator statefulset; full `make release` bumps every service; rollback repoints to the old existing image without rebuild. (2) Site content deploys git → GitHub Actions → Backblaze B2. Operational wrinkle: full-build deploys commit as "Rerender: <page>" — the shared message format no longer distinguishes build from rerender.
- **sources:** README_001_flow_notes.md; running_notes_checkpoint_uu.md#Deploy-rollback; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; running_notes_scheme_to_components(55).md#Th (hygiene)
- **relations:** deployed-binary-predates-disk; agent re-registration.
- **verify-later:** Makefile targets; agent_definitions.image_tag column use.

<!-- SOURCE: U05_content_quality_linking.md -->
### Git per-page deploy + non-fast-forward race
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes_14(26) Part 10: git-adapter "updateRef is force:false + no-retry, so a concurrent commit to the shared sites repo loses with a silent non-fast-forward (FOCUS_dispatch open item 4)".
- **what:** Pages deploy as one git commit each to the shared sites repo (gqls/sites, git→Cloudflare); concurrent commits during a cascade can silently lose a non-fast-forward push. Suspected (not confirmed) in the missing-homepage case before the auto-complete cause was pinned; remains an open reliability item. Minor cosmetic sibling: page-rerender's commit message template "Rerender: {{.filename}}" renders uninterpolated.
- **sources:** running_notes_14(26).md#part-10; NOTES(44) minor findings
- **relations:** deploy gate on sections_saved; dispatch throughput.
- **verify-later:** git-adapter updateRef retry behaviour; commit-message templating scope.

<!-- SOURCE: U09_adoption.md -->
### Git-adapter cross-site commit race (force:false, no retry, shared sites repo)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** "a latent bug today for concurrent multi-site builds… Fix: retry-on-non-fast-forward in updateRef… tracked as a guardrail" ; "(2026-06-04: the missing-homepage lead was NOT this race… remains a latent risk with no confirmed instance yet.)"
- **what:** `git_commit` publishes to the git-adapter (Kafka key = correlation_id), which does a GitHub Git Data API read-modify-write with `force:false` and no retry against a single shared `sites` repo. Same-site commits serialize via partition (git-safe for same-site parallelism); different sites can hit different replicas concurrently and the loser's deploy fails silently on non-fast-forward. Proposed optimistic-concurrency retry (re-read HEAD, rebuild tree, re-commit) not yet built.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#git-deploy-path
- **relations:** Lever A prerequisite; A4 homepage (initially suspected, exonerated)
- **verify-later:** git-adapter `github_client.go` CommitToRepo/updateRef; adapter.go consumer model (2 replicas, group git.adapter.group)

<!-- SOURCE: U11_traffic_probe.md -->
### sites.github_repo as deploy-target selector (resolveGitRepoName patch)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread B: "Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all"; plan P3 "Remaining: land the chassis patch".
- **what:** Tracing showed sites.github_repo is DORMANT end-to-end: git_commit reads config repo_name → default "sites"; upsertSite doesn't SELECT it; ensure_site_record doesn't return it. Specified 3-touch patch: (1) upsertSite RETURNING += COALESCE(github_repo,''), (2) EnsureSiteRecordAction return map += github_repo, (3) a private resolveGitRepoName(config, collected) helper (config repo_name → site_record.github_repo → "sites") used by BOTH git_commit and deploy_image_asset — the latter hardcodes "sites" at line 463 and would otherwise split-brain a probe site (pages → VM repo, images → sites). vet_med_export left alone (dedicated pipeline). Pre-flight confirmed github_repo empty on all 8 sites, so the fallback is safe. CommitToRepo already prefixes <domain>/ for any repo (shared root layout confirmed); createOrGetRepo auto-creates missing repos as PUBLIC — a deliberate-visibility trap.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_plan(12).md#P3, HANDOFF#thread-b
- **relations:** vm-sites repo, D1–D4, requires-backend gate (the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution; whether the patch ever landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Deployment-GitHub / self-hosted runner + deploy path
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 102_blog_handoff "Self-hosted GitHub Actions runner — deployed and running … Runner v2.333.1, pod in ai-persona-system namespace"
- **what:** The publish path: agents commit generated site files via a git adapter; a self-hosted GitHub Actions runner runs the sites-repo workflow which `b2 sync`s to Backblaze. `needs_rerender` is the terminal build item that assembles pages and triggers deployment.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, WM/001_development_guide(0).md#every-pipeline-must-end-with-assembly-and-deployment, WM/007_adoption_pipeline_v3.md#data-flow-between-layers
- **relations:** blog-listing handoff; storage architecture (B2/S3); site plan reconciler terminal items
- **verify-later:** git-adapter; github-actions-runner dockerfile; needs_rerender handler

<!-- SOURCE: U20_legacy_docs_a.md -->
### Git deployment: commit_to_git + GitHub Action sync to B2
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** webbuild_pipeline/001pipeline: "Deployer commits sites/boxing-tickets.com/index.html to GitHub → GitHub Action automatically syncs that folder to B2 → Site is live."
- **what:** Deployment path: a git-adapter microservice (Kafka topic system.adapter.git.requests) commits generated site files to a repo (per-domain repos in the original design; a sites/<domain>/ folder in practice); a GitHub Action syncs to Backblaze B2 which serves the live site. GitCommitAction is the workflow-side action.
- **sources:** docs004_website_capture_project/webbuild_pipeline/001pipeline; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** deployment-github (docs 034 live successor: git-adapter deploy surface); storage-architecture B2.
- **verify-later:** git-adapter service; the GitHub Action workflow file; sites/ repo layout.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site manifest + external-edit desynchronisation detection
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** Design tables in 004 (git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) and manifest.json "winning genes" tracking in 008/014; no implementation evidence.
- **what:** Every generated site carries a manifest.json recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter watches all site repos; a human commit flags the manifest desynchronised, halting agent edit workflows and queueing HITL review — protecting human work from being overwritten and agents from stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism protecting human edits); deployment-github git-adapter.
- **verify-later:** any manifest.json in site repos; git webhook receivers.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Standardized per-page git deployment
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** docs017/002_standardising problem table ("Inconsistent deploy patterns → Standardize to per-page commits"); docs017/023 "Individual git commits per page → each goes live via GitHub Action → S3"; work items store commit_sha in result.
- **what:** Deployment converges on one mechanism: each page is committed individually via git_commit (with file_path override enabling CSS/asset commits), GitHub Actions deploy to hosting (Cloudflare, later S3), and commit SHAs are recorded on pages and work items for traceability. Removed redundant deployer steps whose data paths kept breaking.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Git-Commit-Strategy
- **relations:** git-adapter; deployment-github category; per-page loop.
- **verify-later:** git_commit action file_path config; GitHub Action workflows in site repos.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sites deployment chain (git → GitHub Actions → Backblaze B2) + image-tag chassis deploys
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** Used throughout ("after Actions propagates (a few minutes)"); 016b orientation: "Deployment is image-tag based... each agent's image_tag is bumped to adopt it; workflow (default_config) changes are DB-only and take effect immediately."
- **what:** Everything site-facing reaches production via git_commit to the 'sites' repo (files map keyed by repo-relative path — pages, tools/assets/*.js, assets/js/snippets.js, data/*.json) → GitHub Actions → Backblaze B2 (public), with the long-running git-adapter handling commits. Platform code ships as a chassis image (GitHub → Actions → image) adopted by bumping per-agent image_tag — so a source revert only reaches the cluster on the next build/push, while agent workflow changes (agent_definitions.default_config) are DB-only and instant. B2 404 NoSuchKey is the "page never deployed" signature.
- **sources:** docs/016b_debugging_guide_merged(3).md#orientation; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 (git_commit pattern); docs/RUNBOOK_phase2_provocation_js(29).md#4.2-4.3; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2
- **relations:** system-architecture; plan-storage revert note (pods keep old behaviour until push)
- **verify-later:** git-adapter; sites repo Actions workflow

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### github_repo target selection + resolveGitRepoName patch
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-10 "sites.github_repo is DORMANT end-to-end … The patch (guide's 'small patch' pattern)"; plan(11) P3 "Remaining … land the chassis patch (resolveGitRepoName helper …)".
- **what:** A site's `sites.github_repo` selects deploy target (vm-sites repo vs default "sites"), but was dormant (upsertSite didn't SELECT it, nothing read it). The fix: one `resolveGitRepoName(config, collected)` helper (config repo_name → site_record.github_repo → "sites") used by both `git_commit` and `deploy_image_asset`, plus upsertSite RETURNING + ensure_site_record map additions. `deploy_image_asset` hardcoded "sites" and would split-brain a probe site (pages→VM, logo/hero→sites) without the same fallback.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_plan(11).md#p3
- **relations:** enables P3 pipeline wiring; deploy_image_asset split-brain risk
- **verify-later:** git_deployer_actions.go, site_db_actions.go, upsertSite, EnsureSiteRecordAction, deploy_image_asset:463

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### "Commit is deploy" seam swapped B2→VM + two GitHub Actions
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** runbook(12) §5 "Two separate workflows"; running_notes 2026-06-11 both siblings "rewritten as faithful siblings … Validated".
- **what:** The static "commit is deploy" seam is preserved, only the destination moves. Content Action (`deploy-to-vm.yml` in vm-sites repo): on push, rsync -az --delete each changed root-level `<domain>/` over SSH to `/var/www/vm-sites/<domain>`; self-hosted runner, no CF purge. Engine Action (`deploy-engine-to-vm.yml` in site-engine repo): on push to `**.go`/go.mod, build static stripped linux/amd64, scp, run the narrow `site-engine-deploy` sudo hook (atomic swap + restart). Secrets VM_HOST/VM_USER/VM_SSH_KEY.
- **sources:** traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action, traffic_probe_running_notes(27).md#2026-06-11-live-b2-action
- **relations:** mirrors live deploy-to-b2.yml + Cloudflare Worker; target-agnostic terminal build item
- **verify-later:** vm-sites/.github/workflows/deploy-to-vm.yml, site-engine/.github/workflows/deploy-engine-to-vm.yml

<!-- SOURCE: U25_leopardess_social.md -->
### Chassis build/deploy practice (local Makefile builds, verify against the pod)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby build practice (2026-07-10): "images are built from the local filesystem via the Makefile; commits are at the user's discretion … Do not infer a deployed binary's contents from git history; verify against the running pod."
- **what:** The chassis deployment reality: images build from the local working tree, decoupled from git commits (image tag hand-recorded in commit messages), so the deployed binary can lead or lag the repo; verification is kubectl exec + grep -ac for symbols in /app/agent-chassis. Corollary for site work: a committed Go change (e.g. the A6 imagery routing) is inert until a Makefile build+push. Pod logs are ephemeral across rollouts; spawned agents log in their own pods, not agent-chassis.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 build practice, #2 log-hunting; docs/leopardessconsulting/HANDOFF.md#2-A6; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening)
- **relations:** imagery kind routing (A6 awaiting deploy); operator discipline
- **verify-later:** Makefile build targets; deploy pipeline for the chassis image

<!-- SOURCE: U04_idea_uk.md -->
### Layer-5 gap = a persistent-service wrapper on already-deployed Thunder plumbing
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "The hard plumbing for Layer 5 already exists and is deployed in production… The remaining gap is a persistent-service wrapper — modest, and largely assembling existing pieces" (2026-06-04); no service-deployer built.
- **what:** The honest reassessment of automated backend deployment: provisioning, ssh_exec, presigned-B2 file transfer, and decommission all exist (Thunder adapter, verified in production), but they're built for **ephemeral** training VMs (18h cap, 15-min reaper, credential-free). A persistent service is the exact opposite shape, so the gap is: persistent-mode provisioning (reaper exemption), credential delivery to the box, DNS+TLS wiring, a `service_instances` table (sibling of thunder_instances), and a parameterised setup script — a `service-deployer` orchestrator modelled on model-trainer, with idea.uk as first consumer. Two distinct things kept clear: deploying the engine binary to a VM (infrastructure) vs expressing the engine as chassis actions (Phase D) — complementary, not alternatives.
- **sources:** idea.uk/PARALLEL_engine_deployment_and_layer5.md; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 5)
- **relations:** Thunder adapter (docs033); model-trainer pattern; Path A/setup.sh; 007 box recipe + site_api_routes.
- **verify-later:** thunder_instances table; absence of service_instances; cmd/thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Path A manual VM deploy — setup.sh as the future service-deployer payload
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box 2026-06-05 via this path; setup.sh iterated through real incidents (certbot abort, env comments).
- **what:** "Do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` that converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary — deliberately written so the chassis service-deployer can later `ssh_exec` the same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable). The single-binary model: landing page `go:embed`ded, env in /etc/idea/idea.env, atomic mv-based redeploys.
- **sources:** idea.uk/nginx/README.md; idea.uk/nginx/setup.sh.orig3 (header); idea.uk/nginx/README_setup_box.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A)
- **relations:** Layer-5 wrapper (Path B); page-serving gotchas; VM launch plan.
- **verify-later:** the live box's drift vs setup.sh (the doc's own rule: fold tweaks back in).

<!-- SOURCE: U04_idea_uk.md -->
### VM launch plan — dedicated hardened box, prior OVH reverse-proxy files audited
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old files' concrete bugs "all catalogued in the doc".
- **what:** Infrastructure-track decisions: a **dedicated** VM for idea.uk rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB–1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** Path A; 007 adoption-pipeline box recipe; Layer-5 wrapper.
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort).

<!-- SOURCE: U04_idea_uk.md -->
### B2 dead-drop persistence: one-way flow from the exposed box into the framework DB
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)" 2026-06-04 — design settled, but the live service still runs on orders.json only; no idea-ingest agent or idea_orders table evidenced.
- **what:** Standard tiered/DMZ design for internet-facing satellites: the exposed idea.uk box holds NO core-DB credentials and no network path to the cluster; it keeps a local operational store (JSON now; SQLite analysed — would break the stdlib-only property) and writes immutable terminal-event records to a scoped write-only B2 prefix (the dead-drop, reusing Thunder's presigned pattern); a scheduled in-cluster `idea-ingest` agent polls B2 and idempotently INSERTs into framework Postgres — the system of record. Kafka topic / narrow HTTPS ingest / direct PG all rejected (each is an inbound path in).
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md (persistence checkpoints)
- **relations:** storage-architecture (B2, presigned URLs); scheduler-and-tasks (the ingest schedule); checkpoint-upload plan (same threat model).
- **verify-later:** any idea-events B2 prefix or idea_orders table (expect absent).

<!-- SOURCE: U04_idea_uk.md -->
### VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool: because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, **every chassis build is invisible at the live domain — safe staging-in-place** — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. Monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); deployment-github (monorepo → Actions → B2); hybrid build_approach/hosting_trajectory classifier fields.
- **verify-later:** live nginx config on the box; whether cutover has since happened.

<!-- SOURCE: U04_idea_uk.md -->
### Layer-5 gap = a persistent-service wrapper on already-deployed Thunder plumbing
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "The hard plumbing for Layer 5 already exists and is deployed in production… The remaining gap is a persistent-service wrapper — modest, and largely assembling existing pieces" (2026-06-04); no service-deployer built.
- **what:** The honest reassessment of automated backend deployment: provisioning, ssh_exec, presigned-B2 file transfer, and decommission all exist (Thunder adapter, verified in production), but they're built for **ephemeral** training VMs (18h cap, 15-min reaper, credential-free). A persistent service is the exact opposite shape, so the gap is: persistent-mode provisioning (reaper exemption), credential delivery to the box, DNS+TLS wiring, a `service_instances` table (sibling of thunder_instances), and a parameterised setup script — a `service-deployer` orchestrator modelled on model-trainer, with idea.uk as first consumer. Two distinct things kept clear: deploying the engine binary to a VM (infrastructure) vs expressing the engine as chassis actions (Phase D) — complementary, not alternatives.
- **sources:** idea.uk/PARALLEL_engine_deployment_and_layer5.md; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 5)
- **relations:** Thunder adapter (docs033); model-trainer pattern; Path A/setup.sh; 007 box recipe + site_api_routes.
- **verify-later:** thunder_instances table; absence of service_instances; cmd/thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Path A manual VM deploy — setup.sh as the future service-deployer payload
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** idea.uk LIVE on the Hetzner box 2026-06-05 via this path; setup.sh iterated through real incidents (certbot abort, env comments).
- **what:** "Do it by hand once, and capture the steps as the automation artefact": a single idempotent, non-interactive, parameterised `setup.sh` that converges a fresh Ubuntu box to nginx+TLS+ufw+fail2ban+unattended-upgrades+hardened systemd unit+binary — deliberately written so the chassis service-deployer can later `ssh_exec` the same file (MODE=update = binary swap; re-run = rebuild; anti-lockout guard on SSH password disable). The single-binary model: landing page `go:embed`ded, env in /etc/idea/idea.env, atomic mv-based redeploys.
- **sources:** idea.uk/nginx/README.md; idea.uk/nginx/setup.sh.orig3 (header); idea.uk/nginx/README_setup_box.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md (Path A)
- **relations:** Layer-5 wrapper (Path B); page-serving gotchas; VM launch plan.
- **verify-later:** the live box's drift vs setup.sh (the doc's own rule: fold tweaks back in).

<!-- SOURCE: U04_idea_uk.md -->
### VM launch plan — dedicated hardened box, prior OVH reverse-proxy files audited
- **category:** NEW:backend-service-deployment
- **status-signal:** deployed
- **status-evidence:** Box provisioned 2026-06-04 (Hetzner CX, Nuremberg) following this plan; the year-old files' concrete bugs "all catalogued in the doc".
- **what:** Infrastructure-track decisions: a **dedicated** VM for idea.uk rather than the existing shared OVH multi-domain reverse proxy (blast-radius isolation; the proxy only knows how to reach k8s); reuse of the prior Terraform/nginx/fail2ban/logrotate/prometheus patterns with their specific year-old bugs fixed; secrets confirmed clean before reuse; VM sizing grounded in the engine being I/O-bound (1 vCPU / 512MB–1GB); search-grounded provider comparison (Hetzner vs Oracle vs spot).
- **sources:** idea.uk/nginx/VM_LAUNCH_PLAN.md; idea.uk/running_notes(63).md (2026-06-04 infra checkpoints)
- **relations:** Path A; 007 adoption-pipeline box recipe; Layer-5 wrapper.
- **verify-later:** the OVH proxy box's role for content sites (51.89.148.216 → k8s NodePort).

<!-- SOURCE: U04_idea_uk.md -->
### B2 dead-drop persistence: one-way flow from the exposed box into the framework DB
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** "Persistence decisions LOCKED (updated PERSISTENCE_design.md §10)" 2026-06-04 — design settled, but the live service still runs on orders.json only; no idea-ingest agent or idea_orders table evidenced.
- **what:** Standard tiered/DMZ design for internet-facing satellites: the exposed idea.uk box holds NO core-DB credentials and no network path to the cluster; it keeps a local operational store (JSON now; SQLite analysed — would break the stdlib-only property) and writes immutable terminal-event records to a scoped write-only B2 prefix (the dead-drop, reusing Thunder's presigned pattern); a scheduled in-cluster `idea-ingest` agent polls B2 and idempotently INSERTs into framework Postgres — the system of record. Kafka topic / narrow HTTPS ingest / direct PG all rejected (each is an inbound path in).
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md (persistence checkpoints)
- **relations:** storage-architecture (B2, presigned URLs); scheduler-and-tasks (the ingest schedule); checkpoint-upload plan (same threat model).
- **verify-later:** any idea-events B2 prefix or idea_orders table (expect absent).

<!-- SOURCE: U04_idea_uk.md -->
### VM cutover: nginx front door with reserved tool paths (staging-in-place via DNS)
- **category:** NEW:backend-service-deployment
- **status-signal:** aspirational
- **status-evidence:** Runbook delivered 2026-06-21; "gated on P0 + the site review… deliberate, not done" (TODO P1, 2026-06-26).
- **what:** The go-live mechanism for a chassis-built front site on a VM that already hosts a live paid tool: because idea.uk's DNS (Cloudflare) points at the VM while the chassis deploys to B2, **every chassis build is invisible at the live domain — safe staging-in-place** — and cutover is purely an nginx change: static root for general pages, `location` proxies for the reserved tool paths (/request /confirm /approve /decline /stripe/webhook /internal/* /order/* /op /health /capacity + policy pages), `try_files … =404` so a missed tool path fails loudly, no body rewrites on the webhook location (signature integrity), prove the webhook through nginx BEFORE cutover, rollback = restore one server block. Named biggest risk: reserved-path completeness. Monorepo stays authoritative; the VM is just one more consumer (pull-sync from B2/git or a path-conditional Action push).
- **sources:** idea.uk/RUNBOOK_idea_uk_vm_cutover.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 2); idea.uk/TODO_chassis_and_idea_uk(1).md#P1
- **relations:** scheme-to-components P0 (the gate); deployment-github (monorepo → Actions → B2); hybrid build_approach/hosting_trajectory classifier fields.
- **verify-later:** live nginx config on the box; whether cutover has since happened.

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk deployment topology — Docker/S3 plan superseded by systemd binary on a VM
- **category:** NEW:persistent-service-deployment
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "The 'Go-live checklist' above describes the original Docker/S3 plan. What's actually live differs and is the current truth: ... idea.uk runs as a single Go binary under systemd on a Hetzner box... — not Docker on a container host, and the landing page is embedded in the binary (`//go:embed page.html`), not a separate file on S3."
- **what:** The originally documented deploy plan (containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) was abandoned in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (page embedded via `go:embed`), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze."
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** deploy-from-context-packs guide (six deploy mechanisms); service-deployer pattern (Path B automation of this same shape)
- **verify-later:** the box at 116.203.204.115 (Hetzner, Nuremberg); `/etc/idea/idea.env`; systemd unit `idea`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### service-deployer pattern (persistent-VM automation, "Path B")
- **category:** NEW:persistent-service-deployment
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md` "PARALLEL THREAD — Layer 5 reassessed": "THE REAL GAP... A persistent service is the OPPOSITE [of Thunder's ephemeral VMs]: stays up, reaper-EXEMPT, holds its own credentials... So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS + a service_instances table + a parameterised setup script." Explicitly deferred: "Path A (manual now)... THEN build the service-deployer workflow around the proven script" — Path A was executed manually throughout this archive; Path B (the automated chassis workflow) was never built within it.
- **what:** A proposed chassis-native orchestrator, sibling of `model-trainer`, that would automate what was done by hand for idea.uk: provision a VM in *persistent* mode (reaper-exempt, unlike Thunder's ephemeral 18h-cap training VMs), ship the binary via the existing presigned-B2-URL mechanism, `ssh_exec` a parameterised `setup.sh`, deliver credentials, register in a new `service_instances` table, and health-check. The manual "Path A" run (deploying idea.uk by hand to a Hetzner box, iterating `setup.sh` against real-world failures — placeholder Let's Encrypt emails, systemd `EnvironmentFile` not stripping inline comments, etc.) was deliberately treated as *not throwaway* but as Path B's future payload/capture step.
- **sources:** `running_notes(44).md` ("PARALLEL_engine_deployment_and_layer5.md" summary, "CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted")
- **relations:** Thunder adapter (ephemeral VM precedent, explicitly contrasted); idea.uk deployment topology; deploy-from-context-packs guide (mechanism F)
- **verify-later:** whether `service_instances` table or a `service-deployer` agent definition exists in the live chassis

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk deployment topology — Docker/S3 plan superseded by systemd binary on a VM
- **category:** NEW:persistent-service-deployment
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "The 'Go-live checklist' above describes the original Docker/S3 plan. What's actually live differs and is the current truth: ... idea.uk runs as a single Go binary under systemd on a Hetzner box... — not Docker on a container host, and the landing page is embedded in the binary (`//go:embed page.html`), not a separate file on S3."
- **what:** The originally documented deploy plan (containerised `idea-svc` image + S3-hosted static landing page + separate deploy pipeline) was abandoned in favour of a much simpler shape once real deployment was attempted: one self-contained Go binary (page embedded via `go:embed`), deployed by build → scp → atomic `mv -f` swap → `systemctl restart`, behind nginx + Let's Encrypt on a single Hetzner VM. Explicitly flagged in `GUIDE_deploy_from_context_packs.md` as deploy-mechanism **F**, distinct from the chassis's k8s image path (A), DB/SQL path (B), work-items (C), orchestration triggers (D), and generated-static-sites-via-B2 path (E) — "Self-contained Go binary, file-based persistence, not k8s, not Backblaze."
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)"; `docubundle_idea_golive/GUIDE_deploy_from_context_packs.md` §F; `running_notes(44).md` (VM provisioning checkpoints, 2026-06-04/05)
- **relations:** deploy-from-context-packs guide (six deploy mechanisms); service-deployer pattern (Path B automation of this same shape)
- **verify-later:** the box at 116.203.204.115 (Hetzner, Nuremberg); `/etc/idea/idea.env`; systemd unit `idea`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### service-deployer pattern (persistent-VM automation, "Path B")
- **category:** NEW:persistent-service-deployment
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md` "PARALLEL THREAD — Layer 5 reassessed": "THE REAL GAP... A persistent service is the OPPOSITE [of Thunder's ephemeral VMs]: stays up, reaper-EXEMPT, holds its own credentials... So the gap = a persistent-service WRAPPER + credential delivery + DNS/TLS + a service_instances table + a parameterised setup script." Explicitly deferred: "Path A (manual now)... THEN build the service-deployer workflow around the proven script" — Path A was executed manually throughout this archive; Path B (the automated chassis workflow) was never built within it.
- **what:** A proposed chassis-native orchestrator, sibling of `model-trainer`, that would automate what was done by hand for idea.uk: provision a VM in *persistent* mode (reaper-exempt, unlike Thunder's ephemeral 18h-cap training VMs), ship the binary via the existing presigned-B2-URL mechanism, `ssh_exec` a parameterised `setup.sh`, deliver credentials, register in a new `service_instances` table, and health-check. The manual "Path A" run (deploying idea.uk by hand to a Hetzner box, iterating `setup.sh` against real-world failures — placeholder Let's Encrypt emails, systemd `EnvironmentFile` not stripping inline comments, etc.) was deliberately treated as *not throwaway* but as Path B's future payload/capture step.
- **sources:** `running_notes(44).md` ("PARALLEL_engine_deployment_and_layer5.md" summary, "CHECKPOINT 2026-06-04 (continued) — VM deploy artefacts drafted")
- **relations:** Thunder adapter (ephemeral VM precedent, explicitly contrasted); idea.uk deployment topology; deploy-from-context-packs guide (mechanism F)
- **verify-later:** whether `service_instances` table or a `service-deployer` agent definition exists in the live chassis

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec-supersede rollback pattern (and full snapshot revert as its big brother)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 002(4) rollback steps; 014 documents deployed snapshot system (migration 085)
- **what:** Targeted rollback = flip site_specs is_current to a previous aspect version + create rebuild work items; git history gives per-work-item revert of deployed HTML. Full point-in-time revert = site_snapshots (JSONB capture of site record/specs/pages+components/nav/site_components + git SHA), take_site_snapshot / revert_site_to_snapshot (always takes a pre_revert safety snapshot; does NOT git-revert or touch global content_components/research_results). Admin API + three workflow actions exist; recommended auto-triggers (post-deploy, pre-propagation, nightly) not yet wired.
- **sources:** 002(4)#Site Rollback Pattern; 014 full
- **relations:** component_versions; agent snapshots (different concern)
- **verify-later:** site_snapshots rows in prod; whether post-deploy snapshot step was added

<!-- SOURCE: U01_docs024_numbered_core.md -->
### component_versions population and change_source provenance
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 014 (April 2026 change_source column); 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift)
- **what:** Three best-effort writers: StoreGeneratedComponentAction create (v1) and regen (MAX+1, snapshot BEFORE update), UpdateComponentHTMLAction (tool improvements). change_source records originating work-item source. Unique (component_id, version_number). Lesson: best-effort operations need active monitoring — silent best-effort was "silent no-effort" for two years.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** component regeneration flow; snapshots
- **verify-later:** component_versions row counts by changed_by

<!-- SOURCE: U12_docs024_archives.md -->
### Milestone-tagged site-spec history with inline git-snapshot function
- **category:** site-snapshots-and-revert
- **status-signal:** superseded
- **status-evidence:** Archive `site_specs` schema carries `milestone`, `superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; live doc drops `milestone`/`superseded_by` entirely, replaces inline snapshotting with a work-item-triggered `snapshot-agent`, adds a bounded "last 5 rows" pruning policy, and drops the legacy `page_components.content_snapshot`/`schema_snapshot` columns.
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs", #"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns", #"Snapshots"
- **relations:** superseded by snapshot-agent + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether `page_components.content_snapshot`/`schema_snapshot` columns still exist.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Snapshots and revert (snapshot_agent/revert_agent)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.1 "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"
- **what:** Before patching an agent's `default_config`, call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup` kept as an audit trail. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (is_snapshot/version+1000), the source of several patch/revert footguns.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1, WM/016_debugging_guide_v2_44.md#9
- **relations:** deprecate-not-delete; component_versions history; debugging guide
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup

<!-- SOURCE: U19_sql_tables_components.md -->
### Site snapshots: point-in-time capture and revert
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 085 migration with take_site_snapshot / revert_site_to_snapshot plpgsql functions, iterated twice in-file with column-name fixes — indicating it was actually run and debugged against the live schema.
- **what:** Full site state captured into one self-contained JSONB row (survives row deletions): site record key fields, all current site_specs, all pages with their page_components (content_data + rendered_html), nav groups/items, site_components; git_commit_sha links DB state to deployed files. Revert takes a safety pre_revert snapshot first, then supersedes specs, delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert and does not touch global content_components templates. Triggers: deploy, manual, pre_edit, scheduled.
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql
- **relations:** page_component_history (finer grain); agent snapshot/revert (same philosophy for agents); deployment-github (file-side counterpart).
- **verify-later:** snapshot triggers actually firing on deploy; v_site_snapshots contents.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Site snapshots + dated-backup reversibility discipline
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** Pre-migration snapshot 044a0b57 taken 2026-06-23; take_site_snapshot call pattern in the migrations runbook; dated backup tables (_vonc_pc_backup_20260704/09 etc.) created before every risky UPDATE.
- **what:** Every significant change is preceded by reversibility: `take_site_snapshot(site_id, name, ..., 'manual')` for site state; `snapshot_agent('<type>','<reason>')`/`revert_agent` for agent definitions (never a hand-rolled agent_definitions_backup); ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` before direct row edits, with the explicit rule never to reuse an old backup name (CREATE TABLE IF NOT EXISTS silently no-ops while looking fresh); restore is UPDATE-in-place keyed on id, not delete+insert.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** debugging doctrine; direct-SQL-bypasses-guards caveat
- **verify-later:** take_site_snapshot / snapshot_agent SQL functions (doc 014)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec-supersede rollback pattern (and full snapshot revert as its big brother)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 002(4) rollback steps; 014 documents deployed snapshot system (migration 085)
- **what:** Targeted rollback = flip site_specs is_current to a previous aspect version + create rebuild work items; git history gives per-work-item revert of deployed HTML. Full point-in-time revert = site_snapshots (JSONB capture of site record/specs/pages+components/nav/site_components + git SHA), take_site_snapshot / revert_site_to_snapshot (always takes a pre_revert safety snapshot; does NOT git-revert or touch global content_components/research_results). Admin API + three workflow actions exist; recommended auto-triggers (post-deploy, pre-propagation, nightly) not yet wired.
- **sources:** 002(4)#Site Rollback Pattern; 014 full
- **relations:** component_versions; agent snapshots (different concern)
- **verify-later:** site_snapshots rows in prod; whether post-deploy snapshot step was added

<!-- SOURCE: U01_docs024_numbered_core.md -->
### component_versions population and change_source provenance
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 014 (April 2026 change_source column); 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift)
- **what:** Three best-effort writers: StoreGeneratedComponentAction create (v1) and regen (MAX+1, snapshot BEFORE update), UpdateComponentHTMLAction (tool improvements). change_source records originating work-item source. Unique (component_id, version_number). Lesson: best-effort operations need active monitoring — silent best-effort was "silent no-effort" for two years.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** component regeneration flow; snapshots
- **verify-later:** component_versions row counts by changed_by

<!-- SOURCE: U12_docs024_archives.md -->
### Milestone-tagged site-spec history with inline git-snapshot function
- **category:** site-snapshots-and-revert
- **status-signal:** superseded
- **status-evidence:** Archive `site_specs` schema carries `milestone`, `superseded_by` columns and a `CommitSpecSnapshot` Go function called inline; live doc drops `milestone`/`superseded_by` entirely, replaces inline snapshotting with a work-item-triggered `snapshot-agent`, adds a bounded "last 5 rows" pruning policy, and drops the legacy `page_components.content_snapshot`/`schema_snapshot` columns.
- **what:** The original design kept unbounded site-spec history in the DB, labelled key rows with a `milestone` string (`initial_research`, `post_build`, `rebrand_q2`...), and relied on a bare Go function invoked directly by completing actions to write a `.site-spec.json` git checkpoint. Content-level rollback used `content_snapshot` on `page_components`. This whole history/rollback substrate was replaced by a decoupled model: `site_specs` prunes to last-5-per-aspect, `page_component_history` is a dedicated append-only table for component rollback, and snapshotting became an ordinary dispatched work item (`needs_snapshot` → `snapshot-agent`) rather than an inline side-effect call.
- **sources:** old/older1/005_build_expand_plan.md#"Table: site_specs", #"Git Spec Snapshots"; docs024_key_docs_latest/P1_build_expand_plan.md#"Removing legacy columns", #"Snapshots"
- **relations:** superseded by snapshot-agent + page_component_history; content-governance locking model
- **verify-later:** confirm in DB whether `page_components.content_snapshot`/`schema_snapshot` columns still exist.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Snapshots and revert (snapshot_agent/revert_agent)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.1 "the convention is to call snapshot_agent('<agent-type>') first … revert_agent finds the most recent unrestored snapshot"
- **what:** Before patching an agent's `default_config`, call `snapshot_agent(type, reason)`; roll back with `revert_agent(type)`. Snapshots are rows in `agent_definitions_backup` kept as an audit trail. A legacy pre-migration pattern stored snapshots in `agent_definitions` itself (is_snapshot/version+1000), the source of several patch/revert footguns.
- **sources:** WM/016_debugging_guide_v2_44.md#6.1, WM/016_debugging_guide_v2_44.md#9
- **relations:** deprecate-not-delete; component_versions history; debugging guide
- **verify-later:** snapshot_agent/revert_agent functions; agent_definitions_backup

<!-- SOURCE: U19_sql_tables_components.md -->
### Site snapshots: point-in-time capture and revert
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 085 migration with take_site_snapshot / revert_site_to_snapshot plpgsql functions, iterated twice in-file with column-name fixes — indicating it was actually run and debugged against the live schema.
- **what:** Full site state captured into one self-contained JSONB row (survives row deletions): site record key fields, all current site_specs, all pages with their page_components (content_data + rendered_html), nav groups/items, site_components; git_commit_sha links DB state to deployed files. Revert takes a safety pre_revert snapshot first, then supersedes specs, delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert and does not touch global content_components templates. Triggers: deploy, manual, pre_edit, scheduled.
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql
- **relations:** page_component_history (finer grain); agent snapshot/revert (same philosophy for agents); deployment-github (file-side counterpart).
- **verify-later:** snapshot triggers actually firing on deploy; v_site_snapshots contents.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Site snapshots + dated-backup reversibility discipline
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** Pre-migration snapshot 044a0b57 taken 2026-06-23; take_site_snapshot call pattern in the migrations runbook; dated backup tables (_vonc_pc_backup_20260704/09 etc.) created before every risky UPDATE.
- **what:** Every significant change is preceded by reversibility: `take_site_snapshot(site_id, name, ..., 'manual')` for site state; `snapshot_agent('<type>','<reason>')`/`revert_agent` for agent definitions (never a hand-rolled agent_definitions_backup); ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` before direct row edits, with the explicit rule never to reuse an old backup name (CREATE TABLE IF NOT EXISTS silently no-ops while looking fresh); restore is UPDATE-in-place keyed on id, not delete+insert.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** debugging doctrine; direct-SQL-bypasses-guards caveat
- **verify-later:** take_site_snapshot / snapshot_agent SQL functions (doc 014)
