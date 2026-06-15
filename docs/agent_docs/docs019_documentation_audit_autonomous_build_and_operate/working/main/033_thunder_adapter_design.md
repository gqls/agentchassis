# 013 — Thunder Adapter Design

Proposal for routing all Thunder Compute interactions through a long-running cluster adapter, replacing the SSH-direct approach currently in our gpu-provisioner / training-launcher / status-checker / artefact-collector / decommissioner agents.

This document focuses on the three concerns raised: cost predictability, runaway-GPU prevention, and the credential boundary between Thunder VMs and our framework.

---

## TL;DR

- **Adapter pattern** wraps Thunder Compute API in a single cluster service (matching existing `ollama-adapter`, `image-generator-adapter` etc. pattern). Holds the Thunder API key. Generates ephemeral SSH keys per VM. Generates time-limited presigned URLs for any VM↔storage data flow.
- **Credential boundary preserved**: Thunder VMs receive no long-lived credentials. Dataset download and adapter upload both use presigned URLs that expire in 1-2 hours.
- **Runaway-GPU protection**: every VM has a hard 12-hour uptime cap enforced by Thunder's API on creation, plus a reaper that polls every 15 minutes to decommission anything orphaned, plus a daily cost cap that prevents new provisioning when exceeded.
- **Worst-case cost from a single failure**: ~$0.40 for the 15-minute reaper window. Worst-case from a complete-system failure where everything except Thunder's hard timeout breaks: $20 (one orphaned run's worth).
- **Trade-off**: ~2-3 days of upfront work to build the adapter and reaper. In exchange, every training-side agent shrinks from "implements SSH client + Thunder API + retry logic + SSH key management" to "sends one Kafka message".

---

## Why this beats the alternatives

We've considered three approaches:

### A. SSH-direct from each agent (the current planned path)

Each of `gpu-provisioner`, `training-launcher`, `training-status-checker`, `artefact-collector`, `gpu-decommissioner` independently:
- Holds Thunder API key (or reads it from env)
- Manages SSH keys
- Talks to Thunder API for its operation
- Talks to VM via SSH

**Pros**: simplest code path; stubs we already deployed fit this shape.

**Cons**:
- Thunder credentials spread across 5 agents — every restart/redeploy must propagate them
- SSH key lifecycle duplicated 5 ways (or kludged via shared k8s secret)
- No central state — if chassis pod crashes mid-workflow with VM provisioned, no clean record of what's running
- Reaper logic has to live somewhere. Either in every agent (bug-prone) or as a separate scheduler (basically: a stripped-down adapter)
- VM needs B2 credentials to download dataset → credential boundary breach

### B. HTTP job server on VM (the FOCUS doc's chosen Option B)

`POST /jobs` on a VM-side service. Bearer-token auth. ~200 lines of Python.

**Pros**: clean async API; chassis doesn't hold a 30-90 min connection.

**Cons**:
- **Designed for persistent VMs.** With ephemeral per-run instances, deploying systemd + Caddy + the Python service on every fresh VM is overhead we pay every run.
- Still doesn't solve the credential boundary — VM either has long-lived bearer token or chassis hands one out per call (which is what the adapter does anyway).
- Adds a new architectural pattern that doesn't exist elsewhere in the codebase.
- The `/jobs/{id}/adapter` download path: Thunder VM streams adapter binary out to chassis. Either chassis must be reachable from Thunder (bad credential shape) or VM uploads to S3 (needs B2 keys on VM).

The FOCUS doc made the right call **under its assumptions** (persistent VMs, high reuse). The actual operational model is per-run ephemeral VMs. Different assumptions, different optimal answer.

### C. thunder-adapter (proposed here)

One cluster service. Holds Thunder API key, B2 credentials, SSH keypair store. Listens on Kafka. Maintains state in postgres. Janitor cleans up orphans.

**Pros**:
- Matches existing codebase pattern (5+ adapters use this shape)
- Single source of truth for "what VMs are running"
- Single credential boundary (adapter) vs five (current plan)
- Reaper is a natural extension of the adapter (it's already polling Thunder API anyway)
- VMs get presigned URLs only — no long-lived credentials
- Each agent shrinks to ~50-80 lines
- Can swap Thunder for Lambda Labs / RunPod by swapping just the adapter

**Cons**:
- ~2-3 days to build vs ~1 week for full SSH-direct implementation across 5 agents (so adapter is actually faster total, but requires architectural agreement first)
- One more service to deploy and monitor

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

DATA FLOW DURING TRAINING:
─────────────────────────────────────────────────────────────────────────
1. Chassis exports postgres → uploads JSONL to s3://personae-model-training/
2. Adapter creates VM, generates SSH keypair, generates presigned download URL
3. Adapter SSHs to VM: `curl -o /tmp/data.jsonl '<presigned_url>'`
4. Adapter SSHs to VM: `nohup python3 train.py /tmp/data.jsonl &`
5. Adapter polls VM via SSH for completion
6. Adapter generates presigned upload URL for s3://...adapters/<run_id>/
7. Adapter SSHs to VM: `curl -X PUT --data-binary @adapter.tar.gz '<presigned_url>'`
8. Adapter calls Thunder API to terminate instance
9. Adapter records artefacts/training_runs status
```

**Key property**: arrows OUT of VM go to public endpoints (B2, signed). No arrow goes from VM into our cluster.

---

## Credential boundary: addressing the "Thunder access back to our framework" concern

Your concern: "I didn't really want thunder to have access credentials back to our framework."

Three things flow between cluster and VM. Each is handled to never give the VM long-lived framework credentials:

### Dataset download (chassis-side data → VM)

**Bad approach**: VM has B2 credentials, downloads from S3 directly.
**Adapter approach**: Adapter generates a presigned GET URL for the JSONL object on B2. URL expires in 1 hour. Adapter SSHs `curl -o /tmp/data.jsonl '<url>'` to the VM. After 1 hour, the URL is dead — even if the VM gets compromised post-training and someone copies that URL out, it doesn't work.

### Adapter binary upload (VM → S3)

**Bad approach**: VM has B2 credentials, uploads to S3 directly.
**Adapter approach**: Adapter generates a presigned PUT URL for the destination key. URL expires in 1 hour. Adapter SSHs `curl -X PUT --data-binary @adapter.tar.gz '<url>'` to the VM. Or — alternative — adapter SCPs the file out from VM to its own pod's memory, then uploads itself. The adapter approach allows either pattern.

### Status / log polling (cluster → VM)

The adapter SSHs into the VM to read the training process status (`ps`, `tail` of training log). This is one-way: the adapter pushes commands, reads output, closes the connection. The VM never initiates a connection to anything in our cluster. The SSH credentials are held only by the adapter and are unique per VM (ephemeral keypair).

### What the VM actually has

After provisioning:
- Authorized SSH key from adapter (gives the adapter shell access to the VM, not the reverse)
- Public internet access (Thunder allows by default)
- Two presigned URLs that expire within hours
- The training script and dataset on local disk
- Nothing else

If a Thunder VM is compromised, the blast radius is:
- Whatever's on its disk (just the training data and adapter binary for THIS run)
- Whatever the presigned URLs allow until they expire (download THIS dataset / upload to THIS upload key, both bucketed to a per-run key prefix)

**The compromised VM cannot:**
- Pull other training data from S3 (presigned URL is per-key)
- Read or write our database
- Read or write our Kafka
- Reach our chassis pods
- Reach other Thunder VMs we own

---

## Cost analysis (with specific numbers)

### Per-training-run base costs

From iter_0: 9h 12m runtime, $20 total. Implies ~$2.17/hr effective rate. Thunder pricing for an A100 80GB single instance is roughly $1.40-1.80/hr depending on availability. The $2.17 effective probably includes ~5 min provisioning + ~5 min teardown overhead.

| Run length | Per-run base cost (at $1.80/hr) |
|---|---|
| iter_0 (9h) | $16-20 |
| Future iter_N at 4h | $7-9 |
| Quick test run (1h) | $2-3 |

### Cluster overhead (static)

- thunder-adapter pod: ~256MB RAM, <100mCPU. **Marginal cost in a shared cluster: $0**.
- thunder-reaper scheduled task: runs every 15 min, takes <5 sec each time. **Marginal cost: $0**.
- Database rows: trivial.

### Risk scenarios — what does a "forgotten GPU" cost?

This is the catastrophic case. Let's bound it.

**Without any safeguards** (worst possible architecture):
- VM runs forever at $1.80/hr = **$43.20/day = $1,300/month**. This is the disaster.

**With ONLY a hard 12-hour Thunder API uptime cap** (built into provisioning):
- Even if every other safeguard fails, Thunder kills the instance after 12 hours.
- Worst case wasted spend per orphan: 12h × $1.80 = **$21.60 per orphan event**.

**With the reaper running every 15 minutes**:
- Reaper notices an orphan, decommissions it.
- Worst case wasted spend per orphan: ~17 minutes × $1.80 = **$0.51 per orphan event**.

**With the daily cost cap** (e.g. $200/day):
- Cumulative spend tracked in `thunder_instances` table.
- Adapter refuses new provision requests when over cap. Returns explicit error.
- Worst case escapes through cap only if existing instances continue running while cap is exceeded — but those are already capped at 12h by Thunder.

### Failure-mode catalogue

| Failure | Without safeguards | With adapter+reaper+hard-cap |
|---|---|---|
| Adapter crashes after provision, before decommission | VM runs forever | Reaper finds orphan in ≤15 min, $0.51 cost |
| Decommission Kafka message lost | VM runs forever | Reaper finds orphan, $0.51 cost |
| Network partition prevents decommission API call | VM runs forever | Thunder hard-cap at 12h, $21.60 cost |
| Bug in workflow skips decommission step | VM runs forever | Reaper finds orphan, $0.51 cost |
| Thunder API returns "deleted" but instance still running | Can happen, undetectable | Reaper double-checks via subsequent list calls; alarm if persists |
| Adapter restarted, loses in-memory state | All running VMs become orphans | Adapter rehydrates state from `thunder_instances` table on startup |
| Bug in reaper skips an instance | One orphan persists | 12h hard-cap fires, $21.60 cost |
| Cost cap bug, allows over-budget provisioning | Unbounded cost | Hard-cap still applies per-VM, $21.60 × N |

The catastrophic-cost scenario is now structurally impossible without simultaneous failure of (a) the reaper, (b) Thunder's own hard-uptime, and (c) the cost cap. Three independent safeguards, any one of which is sufficient.

### Summary cost comparison

| Scenario | Without adapter | With adapter |
|---|---|---|
| Per-run cost | $16-20 | $16-20 (same) |
| One-off mistake leaving VM up overnight | $43 | $0.51 |
| Persistent bug undetected for a week | $300 | $7-25 (varies by which safeguard catches it) |
| Mass cluster failure where chassis is down for hours | Multiple orphans, hundreds of $$$ | Adapter pod restart triggers state rehydration; reaper still runs; $5-15 |

---

## Preventing indefinite-running GPUs (defence in depth)

Three independent safeguards, in priority order:

### Layer 1: Thunder API hard uptime (enforced by Thunder)

When the adapter calls `POST /instances/create`, it includes a `max_uptime_hours` parameter (Thunder API supports this). After this elapses, Thunder unconditionally terminates the instance regardless of what we do. Default: **12 hours**.

Trust level: high. Thunder enforces it server-side.

### Layer 2: thunder-reaper (cluster-side reconciliation, every 15 min)

A scheduled task that runs every 15 minutes. For each iteration:

```
1. List all instances from Thunder API (one call)
2. List all rows in thunder_instances where status IN ('provisioning', 'running')
3. For each Thunder API instance:
     - If matches a thunder_instances row with status 'decommissioning' → re-issue terminate
     - If has no matching thunder_instances row → orphan, terminate it
     - If matching row but uptime > max_runtime_hours → terminate, mark as 'reaped'
4. For each thunder_instances row in 'running':
     - If no matching Thunder instance found → status went stale, mark 'lost'
5. Update cumulative cost-this-day in thunder_budget_state
6. If cumulative cost > daily_cap: set adapter status to 'cost_paused'
```

Trust level: medium. Depends on the reaper being scheduled and running, the Thunder API being reachable, and the `thunder_instances` table being correct.

### Layer 3: daily cost cap

`thunder_budget_state` table tracks cumulative cost in the current 24-hour window. Adapter checks this before every provision. Returns explicit `adapter_busy: cost_cap_exceeded` if over.

Default: **$200/day** (~10 training runs at $20 each). Configurable per-environment.

Trust level: low. Easy to bypass with a code bug. Acts as a final back-stop for catastrophic regressions.

### Combined effective max cost-per-failure

To leak more than $20:
- Thunder hard-cap must fail (Thunder API change, account error)
- AND reaper must fail (scheduler down, postgres down, code bug)
- AND no human notices for >12 hours

To leak more than $50:
- All of above
- AND cost cap bypassed
- AND no human notices for >24 hours

---

## Action vocabulary (what messages flow)

Adapter listens on `system.adapter.thunder.requests`. Each request is a single action with typed payload. All return correlation_id-tagged responses on the caller's responses_topic (standard chassis pattern).

### `provision_instance`

```json
{
  "action": "provision_instance",
  "training_run_id": "uuid",
  "instance_type": "a100_80gb_single",   // or "h100_single", etc.
  "max_uptime_hours": 12,
  "tags": {"purpose": "qlora_training", "iter": "iter_1"}
}
```

Adapter does:
1. Check cost cap. Fail if exceeded.
2. Generate ephemeral SSH keypair. Store private in cluster k8s secret.
3. Call Thunder API: create instance, attach public key, set max_uptime.
4. Wait for instance ready (poll Thunder API).
5. **INSERT thunder_instances row BEFORE returning** (status='running').
6. Return `{instance_id, instance_ip, ssh_user, ssh_key_secret_name, max_uptime_hours}`.

If step 4 fails or times out, decommission immediately, return error.

### `decommission_instance`

```json
{"action": "decommission_instance", "instance_id": "thunder-uuid"}
```

Adapter does:
1. UPDATE thunder_instances SET status='decommissioning'.
2. Call Thunder API delete.
3. Verify deletion via API list.
4. UPDATE thunder_instances SET status='decommissioned', cost_usd=...
5. Return `{decommissioned_at, final_uptime_seconds, cost_usd}`.

Idempotent: if instance already gone, return success with cached info.

### `prepare_dataset_url`

```json
{
  "action": "prepare_dataset_url",
  "s3_uri": "s3://personae-model-training/finetuning/datasets/.../training.jsonl",
  "method": "GET",
  "expires_in_seconds": 3600
}
```

Adapter generates a presigned URL. Returns `{presigned_url, expires_at}`.

(Used by `training-launcher` to get a URL it can push into the VM via SSH.)

### `prepare_artefact_url`

```json
{
  "action": "prepare_artefact_url",
  "training_run_id": "uuid",
  "method": "PUT",
  "expires_in_seconds": 3600,
  "key_template": "finetuning/adapters/{training_run_id}/adapter.tar.gz"
}
```

Same pattern, for upload.

### `ssh_exec`

```json
{
  "action": "ssh_exec",
  "instance_id": "thunder-uuid",
  "command": "nohup python3 /opt/train.py /tmp/data.jsonl > /tmp/train.log 2>&1 &",
  "timeout_seconds": 30,
  "background": true
}
```

Adapter SSHs into the VM, runs the command, returns stdout/stderr. For `background: true`, waits long enough to confirm the process started, returns the PID, then disconnects.

### `ssh_get_status`

```json
{"action": "ssh_get_status", "instance_id": "thunder-uuid", "pid": 12345, "log_path": "/tmp/train.log", "log_tail_lines": 200}
```

SSHes to VM, checks `ps -p $PID`, tails the log. Returns `{running: true/false, exit_code, log_tail}`.

### `list_instances` (admin/debug)

Returns all `thunder_instances` rows. For ops dashboards.

---

## New schema

```sql
-- ============================================================================
-- thunder_instances — adapter's source of truth for what VMs exist
-- ============================================================================
CREATE TABLE IF NOT EXISTS thunder_instances (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Thunder identity
    thunder_instance_id         TEXT NOT NULL UNIQUE,
    instance_type               TEXT NOT NULL,
    instance_ip                 TEXT,
    ssh_user                    TEXT NOT NULL DEFAULT 'ubuntu',
    ssh_key_secret_name         TEXT NOT NULL,        -- k8s secret with private key

    -- Lifecycle
    status                      TEXT NOT NULL CHECK (status IN
                                    ('provisioning','running','decommissioning',
                                     'decommissioned','reaped','lost','failed')),
    max_uptime_hours            INTEGER NOT NULL DEFAULT 12,

    -- Linkage
    training_run_id             UUID REFERENCES model_lifecycle.training_runs(id),
    requested_by                TEXT,                 -- agent_type that requested provision

    -- Cost
    hourly_rate_usd             NUMERIC,              -- snapshot at provision time
    cost_usd                    NUMERIC,              -- updated on decommission

    -- Timestamps
    provisioned_at              TIMESTAMPTZ,
    running_since               TIMESTAMPTZ,
    decommission_requested_at   TIMESTAMPTZ,
    decommissioned_at           TIMESTAMPTZ,
    reaped_at                   TIMESTAMPTZ,
    reaped_reason               TEXT,                 -- 'orphan', 'over_uptime', 'cost_cap'

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_thunder_instances_status_running
    ON thunder_instances(status) WHERE status IN ('provisioning','running');

CREATE INDEX IF NOT EXISTS idx_thunder_instances_training_run
    ON thunder_instances(training_run_id) WHERE training_run_id IS NOT NULL;

-- ============================================================================
-- thunder_budget_state — rolling 24-hour cost tracking
-- ============================================================================
CREATE TABLE IF NOT EXISTS thunder_budget_state (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    window_start                TIMESTAMPTZ NOT NULL,        -- start of 24h window
    cost_usd                    NUMERIC NOT NULL DEFAULT 0,
    daily_cap_usd               NUMERIC NOT NULL DEFAULT 200,
    is_paused                   BOOLEAN NOT NULL DEFAULT false,
    pause_reason                TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- View: deployable_provisioning_capacity — for adapter to check before provision
-- ============================================================================
CREATE OR REPLACE VIEW thunder_provision_check AS
SELECT
    bs.cost_usd,
    bs.daily_cap_usd,
    bs.is_paused,
    bs.pause_reason,
    (SELECT COUNT(*) FROM thunder_instances WHERE status IN ('provisioning','running')) AS active_count,
    bs.cost_usd + 25 < bs.daily_cap_usd AS can_provision
FROM thunder_budget_state bs
ORDER BY bs.created_at DESC LIMIT 1;
```

---

## Migration plan from current SSH-direct stubs

We currently have `gpu-provisioner` and `training-launcher` as stubs. Plan:

**Phase 1 (this session): No code yet, just decide.**
You read this doc, decide adapter vs SSH-direct. If adapter: continue. If SSH-direct: discard this doc, proceed to build real 022/023.

**Phase 2 (next): Build adapter foundation.**
- Schema migration (thunder_instances, thunder_budget_state)
- thunder-adapter Go service skeleton (Kafka consumer, postgres connection, health endpoint)
- Deploy as a Deployment alongside other adapters
- Verify it starts up cleanly with no actions yet

**Phase 3: Implement provisioning lifecycle.**
- `provision_instance` action (Thunder API client + SSH key generation)
- `decommission_instance` action
- Integration tests against Thunder (one provision-and-immediately-decommission run, ~$0.50 cost)
- Reaper scheduled task

**Phase 4: Implement data flow.**
- `prepare_dataset_url` and `prepare_artefact_url` actions
- `ssh_exec` and `ssh_get_status` actions
- Integration test: provision, exec a no-op command, decommission

**Phase 5: Replace stubs with real agents.**
- Real `gpu-provisioner`: sends `provision_instance` to adapter, updates `training_runs.thunder_instance_id`, returns instance details
- Real `training-launcher`: sends `prepare_dataset_url`, then `ssh_exec` for download, then `ssh_exec` for training launch
- Real `training-status-checker`: sends `ssh_get_status`
- Real `artefact-collector`: sends `prepare_artefact_url`, then `ssh_exec` for upload, registers row in `model_lifecycle.artefacts`
- Real `gpu-decommissioner`: sends `decommission_instance`, updates `training_runs.cost_usd`

**Phase 6: First end-to-end real training run.**
- Trigger model-trainer with the iter_0 export
- Watch the adapter logs
- Verify instance is decommissioned at the end
- Verify cost is recorded
- Should produce iter_1 adapter at total cost ~$15-25

---

## Time estimate

- Phase 2 (skeleton): 0.5 days
- Phase 3 (provision lifecycle + reaper): 1 day
- Phase 4 (data flow actions): 0.5 days
- Phase 5 (agent replacements): 0.5 days
- Phase 6 (verification): 0.5 days

**Total: 2.5-3 days** of focused work.

For comparison, the SSH-direct path requires:
- gpu-provisioner real implementation: 1 day (Thunder API + SSH key gen)
- training-launcher real implementation: 1 day (SSH client + SCP)
- training-status-checker real implementation: 0.5 days
- artefact-collector real implementation: 0.5 days (SCP + S3 upload)
- gpu-decommissioner real implementation: 0.5 days
- Reaper / orphan cleanup mechanism: 1 day (would need to be added either way)
- Verification: 0.5 days

**SSH-direct total: ~5 days**, plus duplicated logic across 5 agents that's harder to debug and modify.

The adapter is roughly **half the engineering effort** because the shared infrastructure (SSH client, Thunder API, key management, retry, circuit breaker) is written once instead of five times.

---

## Open questions for you

1. **Daily cost cap value.** Default proposed: $200/day. Comfortable, raise it, or lower it?

2. **Hard uptime cap.** Default proposed: 12 hours. iter_0 ran 9h 12m so 12h gives only ~30% headroom for iter_1 if something runs slightly longer. Maybe 18h is safer? Worst-case orphan cost goes from $21.60 to $32.40 — is the safety margin worth it?

3. **Reaper interval.** Proposed: every 15 minutes. Could be 5 minutes for tighter orphan capture (cost difference per orphan: $0.17 vs $0.51 — minor). Could be 30 minutes if cluster pressure matters (per-orphan: $1.02). 15 feels balanced.

4. **Provisioning concurrency limit.** Should the adapter cap concurrent active VMs (e.g. max 3 at once)? Useful for cost predictability and accidental loop protection.

5. **Daily cost cap reset behaviour.** Hard reset at UTC midnight, or rolling 24h window? Rolling is more cautious (prevents 23:59 + 00:01 double-spend). Hard reset is simpler.

6. **The FOCUS doc's Option B status.** Should we formally retract it in the doc? My recommendation: yes, with a one-paragraph note explaining the assumption change (persistent → ephemeral VMs) so the reasoning is preserved.

7. **Adapter as separate repo vs same repo.** All other adapters (web-search, image-generator, etc.) are in the agentchassis repo as `cmd/<adapter-name>/`. Same pattern for thunder-adapter? Yes, I assume.

Once you've decided on these — and on the broader yes-go vs no-go for the adapter — we can move into Phase 2.
