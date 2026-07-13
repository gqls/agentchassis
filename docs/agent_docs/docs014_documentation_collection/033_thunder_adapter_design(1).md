# 013 — Thunder Adapter Design

Canonical design for routing all Thunder Compute interactions through a long-running cluster adapter, replacing the SSH-direct approach in `gpu-provisioner` / `training-launcher` / `training-status-checker` / `artefact-collector` / `gpu-decommissioner`.

---

## Formal retraction of FOCUS doc Option B

**`FOCUS_finetuning_flywheel_and_service_13_.md` Option B (HTTP job server on the GPU VM) is superseded by this document.**

Reason: Option B was chosen under the assumption of long-lived persistent VMs that serve many training runs, where the cost of installing systemd + Caddy + the Python service amortises across many jobs. The actual operational model is per-run ephemeral instances (provision → train → decommission). Under that model, on-VM service installation is overhead paid every run, and credentials would have to live on each VM.

The thunder-adapter approach (this document) instead places the long-running service in our cluster, keeps Thunder VMs credential-free (presigned URLs only), and consolidates the Thunder API surface into one component. It matches the existing codebase pattern used by `ollama-adapter`, `image-generator-adapter`, etc.

This retraction is recorded here for the historical record. The FOCUS doc itself can be updated by the user when convenient; this note serves as the authoritative pointer in the meantime.

---

## TL;DR with confirmed decisions

- **Adapter pattern**, holding `THUNDER_COMPUTE_API_KEY`, B2 credentials, and per-VM ephemeral SSH keys.
- **Daily cost cap: $100/day** (rolling 24-hour window, never UTC midnight reset).
- **Hard uptime cap: 18 hours** per VM, enforced by Thunder API at provisioning time (~50% headroom over iter_0's 9h12m).
- **Reaper interval: 15 minutes** (scheduled task reconciling Thunder API ↔ our `thunder_instances` table).
- **Concurrency limit: 2** active VMs at any time. Structurally consistent with $100/day cap (2 × $1.80/hr × 24h = $86 max sustained).
- **Same repo** — `cmd/thunder-adapter/main.go` matching pattern of other adapters.
- **Worst-case cost from a single failure**: ~$0.51 for the reaper window. Worst-case from full safeguard failure: $32.40 (one orphaned 18h run).

---

## Why this beats the alternatives

(See "Alternatives considered" section below — three options compared. Adapter pattern wins on credential boundary, codebase consistency, total engineering effort.)

---

## Architecture (with credential boundaries marked)

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Our cluster (ai-persona-system)                  │
│                                                                       │
│   ┌─────────────────────┐                                            │
│   │  agent-chassis pods │                                            │
│   │  ┌─────────────────┐│                                            │
│   │  │ training-data-  ││──── (already done) ─┐                     │
│   │  │ preparer        ││                     │                     │
│   │  │ gpu-provisioner ││ ───┐                │                     │
│   │  │ training-       ││ ───┤                │                     │
│   │  │   launcher      ││ ───┤                │                     │
│   │  │ training-status-││ ───┤  Kafka         │                     │
│   │  │   checker       ││ ───┤  msgs          │                     │
│   │  │ artefact-       ││ ───┤                │                     │
│   │  │   collector     ││ ───┤                │                     │
│   │  │ gpu-            ││ ───┤                │                     │
│   │  │   decommissioner││ ───┘                │                     │
│   │  └─────────────────┘│                     │                     │
│   └─────────────────────┘                     │                     │
│                                                ▼                     │
│   ┌─────────────────────────────────────────────────┐                │
│   │  thunder-adapter (NEW Deployment)               │                │
│   │  ┌────────────────────────────────────────────┐ │                │
│   │  │ Holds: THUNDER_COMPUTE_API_KEY             │ │                │
│   │  │ Holds: B2_APPLICATION_KEY_ID/KEY           │ │                │
│   │  │ Holds: SSH keypair (cluster-internal)      │ │                │
│   │  │ Maintains: thunder_instances table state   │ │                │
│   │  └────────────────────────────────────────────┘ │                │
│   └─────────────────────────────────────────────────┘                │
│                          │                                            │
│   ┌─────────────────────────────────────────────────┐                │
│   │  thunder-reaper (NEW scheduled_task, every 15m) │                │
│   │  Reconciles: API instances ←→ thunder_instances │                │
│   │  Decommissions: orphans, over-budget, stuck     │                │
│   └─────────────────────────────────────────────────┘                │
└────────────────────────────│──────────────────────│──────────────────┘
                             │ Thunder REST API     │ presigned-URL
                             │ (out only)           │ (out only)
                             ▼                      ▼
                  ┌────────────────────┐    ┌──────────────────┐
                  │  Thunder Compute   │    │  Backblaze B2    │
                  │  ─────────────     │    │  bucket:         │
                  │  Ephemeral A100    │    │  personae-       │
                  │  VM (no creds!)    │    │  model-training  │
                  │                    │    │                  │
                  │  - SSH from cluster│    │                  │
                  │  - curl presigned  │    │                  │
                  │    download URL    │    │                  │
                  │  - curl presigned  │    │                  │
                  │    upload URL      │    │                  │
                  └────────────────────┘    └──────────────────┘
```

---

## Credential boundary (key concern)

Thunder VMs have NO long-lived framework credentials. Every cluster↔VM data flow uses time-limited presigned URLs.

**Dataset download** (chassis JSONL → VM): Adapter generates a presigned GET URL (1h expiry), SSHs `curl -o /tmp/data.jsonl '<url>'` to VM. URL becomes useless after 1h.

**Adapter binary upload** (VM → S3): Adapter generates presigned PUT URL (1h expiry), SSHs `curl -X PUT --data-binary @adapter.tar.gz '<url>'`. Or alternative: adapter SCPs binary out, uploads from its own pod.

**Status / log polling** (cluster → VM): Adapter SSHs in, runs `ps`/`tail`, disconnects. One-way: cluster pushes commands; VM never initiates anything.

**If a Thunder VM is compromised**, the blast radius is:
- Files on its disk (just THIS run's training data and adapter binary)
- The two presigned URLs until they expire (per-key, can't access other data)

The compromised VM cannot read our database, send Kafka messages, reach chassis pods, or access other VMs.

---

## Cost analysis (confirmed numbers)

### Per-training-run base costs

iter_0: 9h 12m, $20 total. Implies effective rate ~$2.17/hr (likely Thunder A100 80GB single instance at $1.40-1.80/hr + provisioning/teardown overhead).

| Run length | Per-run base cost (at $1.80/hr) |
|---|---|
| iter_0 (9h) | $16-20 |
| Future iter_N at 4h | $7-9 |
| Quick smoke test (1h) | $2-3 |

### Cluster overhead (static)

- thunder-adapter pod: ~256MB RAM, <100mCPU. **Marginal cost: $0** in shared cluster.
- thunder-reaper: scheduled task running every 15 min, <5 sec each. **Marginal cost: $0**.

### Failure-mode catalogue (with confirmed safeguards)

| Failure | Without safeguards | With adapter + reaper + 18h cap |
|---|---|---|
| Adapter crashes after provision, before decommission | VM runs forever ($43/day) | Reaper finds orphan in ≤15 min, **$0.51 cost** |
| Decommission Kafka message lost | VM runs forever | Reaper finds orphan, **$0.51 cost** |
| Network partition prevents decommission API call | VM runs forever | Thunder hard-cap fires at 18h, **$32.40 cost** |
| Bug in workflow skips decommission step | VM runs forever | Reaper finds orphan, **$0.51 cost** |
| Adapter restarted, loses in-memory state | All running VMs orphaned | Adapter rehydrates state from `thunder_instances` table on startup; reaper still operates |
| Bug in reaper skips an instance | One orphan persists | 18h hard-cap fires, **$32.40 cost** |
| Cost cap bug allows over-budget provisioning | Unbounded cost | Hard-cap and concurrency limit still apply per-VM |
| Concurrent provisioning attack/loop | Unbounded VM creation | **Concurrency limit (2) blocks all but 2** |

### Daily cost cap with rolling window

Rolling 24h prevents the UTC-midnight boundary bug ($99 spent at 23:59 + new run at 00:01 would total $198 in 2 minutes with hard reset). Rolling tracks: sum of `cost_usd` from instances decommissioned in last 24h + estimated cost for currently-running instances based on `running_since` and `hourly_rate_usd`.

When `current_24h_spend + estimated_new_run_cost > daily_cap`, adapter refuses provision requests with explicit `denial_reason='cost_cap_would_exceed'`.

---

## Defence in depth (three independent safeguards)

### Layer 1: Thunder API hard uptime (18 hours)

`POST /instances/create` with `max_uptime_hours=18`. Thunder unconditionally terminates the instance when reached. Server-side enforcement, doesn't depend on our cluster being up.

### Layer 2: thunder-reaper (every 15 minutes)

Scheduled task. Each run:
1. List all instances from Thunder API (one HTTP call)
2. List all `thunder_instances` rows in 'provisioning' or 'running'
3. For each Thunder API instance:
     - If matches a `decommissioning` row → re-issue terminate
     - If has no matching row → orphan, terminate it
     - If matching row but uptime > 18h → terminate, mark 'reaped'
4. For each `running` row:
     - If no matching Thunder instance found → status went stale, mark 'lost'
5. Recompute cumulative 24h cost

### Layer 3: cost cap + concurrency limit (per-call)

Every `provision_instance` call checks `thunder_provision_check` view:
- `is_paused = false`
- `active_count < max_concurrent_instances` (currently 2)
- `total_24h_spend + estimated_new_run_cost ≤ daily_cap_usd` ($100)

Returns explicit denial reason if any check fails.

---

## Action vocabulary (Kafka message shapes)

Adapter listens on `system.adapter.thunder.requests`. Standard chassis pattern: each request returns to the caller's `responses_topic`.

```
provision_instance         → {instance_id, instance_ip, ssh_user, ssh_key_secret_name, max_uptime_hours}
decommission_instance      → {decommissioned_at, final_uptime_seconds, cost_usd}
prepare_dataset_url        → {presigned_url, expires_at}    (GET)
prepare_artefact_url       → {presigned_url, expires_at}    (PUT)
ssh_exec                   → {stdout, stderr, exit_code, pid}
ssh_get_status             → {running, exit_code, log_tail}
list_instances             → {instances: [...]}             (admin/debug)
```

Detailed payloads for each are in the schema migration's documentation block.

---

## Alternatives considered (for the historical record)

### A. SSH-direct from each agent (rejected)

Each of 5 agents independently holds Thunder API key, manages SSH keys, talks to Thunder. Spreads credentials across 5 places, duplicates SSH lifecycle 5 ways, requires VM to have B2 credentials for dataset download (credential boundary breach).

### B. HTTP job server on VM (FOCUS doc Option B, retracted)

Per-VM systemd unit + Caddy + Python service. Designed for persistent VMs; overhead paid every run under ephemeral model. Doesn't solve credential boundary on its own.

### C. Adapter (selected) — see entire document above

---

## Migration plan

**Phase 1 (this session): decision + schema** ← currently here
- This document (decision recorded)
- `025_thunder_adapter_schema.sql` (schema migration)

**Phase 2: Adapter skeleton**
- `cmd/thunder-adapter/main.go` Go service skeleton
- Config YAML (`thunder-adapter.yaml`)
- Deployment manifest matching other adapters
- Verify clean startup with no actions yet

**Phase 3: Provisioning lifecycle**
- `provision_instance` action (Thunder API client + SSH key generation)
- `decommission_instance` action
- thunder-reaper scheduled task
- Integration test: one provision-and-immediately-decommission run (~$0.50 cost)

**Phase 4: Data flow**
- `prepare_dataset_url`, `prepare_artefact_url` actions
- `ssh_exec`, `ssh_get_status` actions
- Integration test: provision, exec a no-op, decommission

**Phase 5: Replace stubs with real agents**
- Real `gpu-provisioner` (sends `provision_instance` to adapter)
- Real `training-launcher` (sends `prepare_dataset_url` + `ssh_exec`)
- Real `training-status-checker` (sends `ssh_get_status`)
- Real `artefact-collector` (sends `prepare_artefact_url` + `ssh_exec`, registers row in `model_lifecycle.artefacts`)
- Real `gpu-decommissioner` (sends `decommission_instance`)

**Phase 6: First end-to-end real training run**
- Trigger model-trainer with iter_0 export
- Watch adapter logs
- Verify decommission at end
- Verify cost recorded
- Should produce iter_1 adapter at total cost ~$15-25

Total estimated effort: **2.5-3 days** of focused work.

---

## Open questions resolved by user

| Question | Decision |
|---|---|
| Daily cost cap | $100/day |
| Hard uptime | 18 hours |
| Reaper interval | 15 minutes |
| Concurrency limit | 2 (default), configurable |
| Cost cap reset | Rolling 24h |
| FOCUS Option B | Formally retracted (above) |
| Repo placement | Same repo, `cmd/thunder-adapter/` |
