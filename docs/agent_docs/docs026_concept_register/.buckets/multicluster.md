
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
