# HANDOFF — Multi-Cluster Agent Dispatch (branch: `multi-chassis`)

**Purpose of this doc:** let a fresh chat pick up the multi-cluster work without re-deriving everything. Read this first, then the two FOCUS docs it references.

---

## 0. Status in one paragraph

We're adding the ability to dispatch agents to a *different* Kubernetes cluster while they keep talking to their parent through the shared Kafka. The parent-side action and the remote spawner service are **already written and committed** on the main codebase. **Nothing in this plan has been deployed or applied yet.** This work happens on a new branch (`multi-chassis`) against a **different set of clusters** from the ones referenced in the planning chat — so every concrete IP, cluster name, and region key in the FOCUS docs is illustrative and must be re-discovered against the new environment.

## 1. What's already built (in code, not deployed)

- **`platform/orchestration/actions/dispatch_actions.go`** — `DispatchAgentAction`. Mirrors `SpawnAgentAction` but publishes a `DispatchRequest` to Kafka topic `system.dispatch.requests` instead of creating a local K8s Job. Reuses `extractSpawnConfiguration`, `setupAgentTopics`, `createAgentInDBFromDefinition`, `sendInitializationMessage`, `buildSpawnResult`, `preRegisterAwaitedRequest`. Sends `target_cluster` in the **message headers**.
- **Registered** as `dispatch_agent` in `platform/orchestration/actions/registry.go` (`IsLocal: true`, category `agent`).
- **`cmd/remote-job-spawner/main.go`** — consumes `system.dispatch.requests`, filters by the `target_cluster` header (`"" | <my cluster> | "any"` are accepted, others skipped), calls `createAgentJob` (a K8s-side mirror of `spawnAgentKubernetesJobFromDefinition`), and replies on `system.dispatch.responses` with a `DispatchResponse`.
- **Build/deploy plumbing** exists: makefile `build-remote-job-spawner`, `deploy-remote-job-spawner` (kustomize), `deploy-remote-job-spawner-tf` (terraform). Kustomize base path `deployments/kustomize/services/remote-job-spawner/`. Confirm these files actually exist on the `multi-chassis` branch before relying on them.

## 2. Known issues to fix (small, structural)

1. **`logger.Debug` in the spawner skip path.** In `cmd/remote-job-spawner/main.go`, the "message not for this cluster, skipping" line uses `logger.Debug`, which doesn't appear in our logs. Change to `logger.Info` so we can verify the cluster filter works once a second spawner exists. ~4 lines. No variable-name changes.
2. **No consumer for `system.dispatch.responses` (Gap B).** The parent fires the dispatch and sleeps ~12s, then sends the init message and sleeps ~8s. A failed dispatch is silent until the awaited-request times out (~30s+). Fix: an `agent_dispatch_log` table (separate from `agent_instances` — different lifecycle) + a small consumer goroutine in **core-manager** (it already runs system-level consumers; don't spin up a new service). This is parent-side and independent of whether a second cluster exists — can be done in parallel.
3. **Verify `OVERRIDE_DATABASE_HOST/PORT/SSLMODE` are honoured by the spawner.** `OVERRIDE_KAFKA_BROKERS` is confirmed supported; grep `createAgentJob` to confirm the DB overrides are passed through to the spawned Job's env. If not, add them mirroring the Kafka override pattern.

## 3. Architecture decisions already made

- **One shared Kafka, multiple K8s clusters.** Agents talk to each other through the shared Kafka regardless of which cluster they run in. No MirrorMaker / federation in scope.
- **Cross-cluster Kafka reachability = a new Strimzi external listener.** `type: nodeport` for the smoke test ($0, brittle on node preemption), promote to `type: loadbalancer` for production. The promotion is a one-word YAML change downstream of no code changes.
- **Postgres strategy:**
  - *Smoke test:* expose PgBouncer via a NodePort Service and connect cross-cluster, primarily to **measure round-trip time** so we know which agent types are viable remotely.
  - *Production (the recommended target topology, "Option C"):* local PgBouncer on each remote cluster + a private tunnel (Tailscale or cloud VPN) back to primary. Keeps Postgres in-cluster (no public exposure), keeps every agent's DSN identical across clusters (`pgbouncer.ai-persona-system.svc.cluster.local:6432`), localises failure modes. Full rationale in the multi-cluster FOCUS doc §8.
  - The **chassis code needs no changes** for cross-cluster Postgres — `postgres.go` already parameterises `sslmode`, has `pgxpool.HealthCheckPeriod=30s`, env-tunable pool sizes, retries, and `default_query_exec_mode=simple_protocol` (PgBouncer transaction-mode compatible).
- **Authorization caveat (carry this forward).** The live Kafka cluster has **no `spec.kafka.authorization`**, so KafkaUser ACLs are not enforced — everything connects as `User:ANONYMOUS` with full access. Adding `scram-sha-512` to the external listener gates the *connection* but an authenticated user is unrestricted. For an internal smoke test that's acceptable — we control the credentials. **Flag it for the production-promotion stage when we enable cluster-wide `authorization: simple`. That's a separate bigger change** — risky because every in-cluster app currently connects anonymously, so `personae-app-anonymous` must cover everything before enforcement is turned on.

## 4. Key environment facts (from the original cluster set — RE-VERIFY on the new one)

These were true of the planning environment. The new cluster set will differ; re-run discovery.

- **Provider:** Rackspace Spot. Nodes report addresses as `InternalIP` but those are **routable public IPv4** (e.g. the original set used `134.213.168.0/24`). This is why nodeport cross-cluster worked. **Re-check on the new clusters** — confirm node IPs are routable between them before committing to nodeport.
- **Kafka:** Strimzi, KRaft mode (no ZooKeeper), NodePools. Live prod CR is rendered from `deployments/terraform/modules/kafka-cluster/config/kafka-cluster-cr-prod_yaml.tpl` via `templatefile()` in the module `main.tf`. Two internal listeners: `plain` (9092), `tls` (9093). The cluster is `personae-kafka-cluster` in namespace `kafka`; node pool `combined-pool-prod` (3 replicas). **Stale/dev variants to ignore:** `kafka-cluster-cr.yaml` (has ZooKeeper — not live), `kafka-kraft-cluster.yaml`, `kafka-temp-fix.yaml`.
- **PgBouncer service is named `pgbouncer`** (not `pgbouncer-clients`). Clients DB DSN: `pgbouncer.ai-persona-system.svc.cluster.local:6432`, user `clients_user`, db `clients_db`. Templates DB is accessed **directly** (no PgBouncer): `postgres-templates.ai-persona-system.svc.cluster.local:5432`, user `templates_user`, db `templates_db`. Both currently `sslmode: disable`.
- **Namespaces:** app workloads in `ai-persona-system`, Kafka in `kafka`, Strimzi operator in `strimzi`.
- **Secrets the spawned Job references:** `personae-platform-secrets` (DB passwords + agent bootstrap key), `personae-default-secrets` (Anthropic key), `personae-storage-secrets` (Backblaze keys). Plus the cross-cluster Kafka user secret and the Kafka cluster CA cert. These must be replicated to any remote cluster's `ai-persona-system` namespace.
- **Deployment:** push to GitHub → GitHub Actions writes to Backblaze S3.

## 5. The two FOCUS docs (the actual working plans)

1. **`FOCUS_multi_cluster_dispatch_mvp.md`** — the broad multi-cluster plan. Phases 0-6: loopback test → dispatch observability (Gap B) → second same-network cluster → cross-cloud (AWS/GCP) → Postgres topology (the §8 deep-dive on Option C) → hardening. Read §8 (Postgres strategy) and §2 (structural gaps) carefully.
2. **`FOCUS_adjacent_cluster_phase4a.md`** — the focused execution plan for the *first* concrete step: one adjacent cluster on the same provider, Kafka external listener + PgBouncer NodePort, dispatch a real agent, measure RTT. This is the one to execute against first. Its §5 is the step-by-step sequence; §1.5 has the field notes (will need refreshing for the new clusters).

## 6. First moves in the new chat

1. **Confirm the branch and that nothing is applied.** `git branch` shows `multi-chassis`; `kubectl get` against the new clusters shows no `remote-job-spawner`, no external Kafka listener.
2. **Re-discover the new cluster set.** For each cluster: `kubectl get nodes -o wide` (get the routable IPs), `kubectl -n kafka get kafka personae-kafka-cluster -o yaml | grep -A30 listeners` (confirm listener state), `kubectl -n ai-persona-system get svc pgbouncer -o yaml` (confirm the selector and that the service exists). Verify cross-cluster node-IP reachability with a `nc -vz` from a Pod on cluster B to a node IP of cluster A.
3. **Pick the region keys / cluster IDs** for the new environment (the planning chat used `uk001` and `va001` — the new set will have its own names).
4. **Then follow `FOCUS_adjacent_cluster_phase4a.md` §5** in order. The two load-bearing checkpoints are kafkacat-from-laptop (confirms the listener) and kafkacat-from-remote-Pod (confirms cross-cluster auth). If both pass, the dispatch test is mechanical.
5. **In parallel**, the Gap B work (dispatch log + response consumer) is parent-side and can proceed without the second cluster.

## 7. Working preferences to respect (from the project owner)

- Every agent is an orchestrator. Keep workflows simple; put complexity in Go action code. Keep workflow variable names in sync with what the actions expect.
- Reuse and adapt existing functions/structs before writing new ones. Think hard about this each time.
- Don't create subworkflows in SQL — spawn sub-agents with their own workflows (keeps logs clear, responsibilities separate).
- Check DB schemas before writing SQL.
- Don't change existing variable names unless deliberate, and note it when you do.
- No `logger.Debug` (won't show in logs) — use `Info`/`Warn`/`Error`.
- Prioritise structural fixes over quick fixes. Don't jump to conclusions. Work in reasonable step sizes.
- Keep responses pragmatic. No "perfect/critical/excellent", no congratulations. Phrase responses with the flow of work in mind (mention the function calls that led to a problem where useful).
- Don't create summary documents unless asked; a brief end-of-response summary is fine.
- Namespaces: `kubectl -n ai-persona-system …`, `kubectl -n kafka …`. Kafka cluster is `personae-kafka-cluster`.
- Deployment: GitHub → GitHub Actions → Backblaze S3.

## 8. Open questions to settle in the new environment

- Are the new clusters' node IPs routable to each other (decides nodeport vs loadbalancer from the start)?
- Same Kafka topology on the new set (KRaft + NodePools, two internal listeners)? Or different?
- Will the new "primary" host Kafka + Postgres, with the other as a pure dispatch target? (That's the assumption — confirm.)
- Geographic separation between the new clusters? Drives the RTT expectations and which agent types are viable remotely.
- Do we want the Gap B dispatch-log work folded into this branch, or kept separate? It's independent of the cross-cluster networking.
