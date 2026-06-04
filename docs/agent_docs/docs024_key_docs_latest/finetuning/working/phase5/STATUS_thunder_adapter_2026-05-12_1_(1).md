# STATUS — thunder-adapter Phase 3 (2026-05-12)

Live status for the Thunder Compute adapter build. Phase 2 (skeleton) deployed and verified. Phase 3 (lifecycle actions) — sub-phases 3.0–3.5 deployed and verified end-to-end. Phase 3.6 (gpu-provisioner real impl) delivered as code, awaiting deploy.

---

## 1. Sub-phase progress

| Phase | Scope | Status |
|---|---|---|
| **2** | Adapter skeleton: Kafka consumer/producer, DB connection, health endpoints, `not_implemented` responses for all actions | ✅ Deployed (`v1.0.1010`), smoke-tested end-to-end |
| **3.0** | Config URL fix: `https://api.thundercompute.com/v1` → `https://api.thundercompute.com:8443/v1` (correct port + path per Thunder docs) | ✅ Delivered, in-tree |
| **3.1** | Thunder Compute API client (`internal/adapters/thunder/api/`): `Client`, `CreateInstance`, `ListInstances`, `GetInstance` with 404→list fallback, `DeleteInstance` idempotent, `WaitForRunning` polling, `APIError` with classifier helpers | ✅ Delivered. Note: user had types.go = duplicate of client.go after file transfer; corrected. |
| **3.2** | SSH keypair generation + k8s Secret CRUD (`internal/adapters/thunder/ssh/`): ed25519 OpenSSH-format keypair, `SecretManager` wrapping kubernetes.Interface (testable), in-cluster constructor for prod, idempotent delete on 404. Plus `rbac.yaml` (Role + RoleBinding scoped to Secrets, bound to ai-persona-app SA) | ✅ Delivered. RBAC kustomize resource added to base/. |
| **3.3** | `provision_instance` action handler. Pre-check via `thunder_provision_check` view → ed25519 keypair → Thunder API create → k8s Secret persist → WaitForRunning poll → INSERT with retry (3 attempts, 1s/3s/5s backoff) → compensating cleanup on partial failure | ✅ Delivered |
| **3.4** | `decommission_instance` action handler. Lookup by `provisioning_id` (DB UUID) or `thunder_identifier` (numeric) → atomic state transition to `decommissioning` (idempotency anchor) → Thunder API delete → Secret delete → compute cost from `running_since` → mark `decommissioned` | ✅ Deployed and verified end-to-end via synthetic reaper test (2026-05-14): cost computed correctly at `$3.61` for 2h × $1.80, status transitioned cleanly, API 404 + Secret 404 both treated as success per idempotent design. |
| **3.5** | thunder-reaper scheduled task. Runs every 15 min. Pre-query finds `running` instances older than `max_uptime_hours`. Dispatches `decommission_instance` for each. | ✅ Deployed and verified end-to-end (2026-05-14): synthetic row with `running_since = NOW() - 2h`, `max_uptime_hours = 1` was picked up within 30s of the scheduler kick, decommission flow ran clean, `last_completed_at` set 3s before the decommission finished. |
| **3.6** | Replace `gpu-provisioner` stub (migration 022) with real implementation. New `dispatch_thunder_provision` action publishes `provision_instance` to thunder-adapter and awaits the response. Migration 029 UPDATEs the agent_definition's workflow. Caller contract preserved — model-trainer's `call_provisioner → call_launcher` chain works without changes. | ✅ Delivered as code, ⏳ awaiting deploy of chassis v1.0.1014 + migration 029. |

---

## 2. Files delivered

```
internal/adapters/thunder/
├── api/                                    [3.1]
│   ├── types.go                            CreateInstanceRequest/Response, Instance, status constants
│   ├── client.go                           HTTP client + WaitForRunning + APIError
│   └── client_test.go                      httptest-based unit tests
├── ssh/                                    [3.2]
│   ├── keypair.go                          ed25519 GenerateKeypair
│   ├── keypair_test.go
│   ├── secrets.go                          SecretManager (k8s Secret CRUD)
│   └── secrets_test.go                     fake clientset tests
├── store/
│   ├── config.go                           [3.3] LoadConfig + CheckCanProvision
│   └── instances.go                        [3.4] LookupByID/ByThunderIdentifier, MarkDecommissioning, MarkDecommissioned, ComputeCost
├── adapter.go                              [patched 3.3 + 3.4] +three fields, +three init blocks, +switch dispatch, +handleProvisionInstance, +handleDecommissionInstance, +sendSuccessResponse, +isProvisionDenial, +isInfrastructureError
├── provision_action.go                     [3.3] ProvisionAction.Execute with compensating cleanup
├── provision_action_test.go                [3.3] helper unit tests
└── decommission_action.go                  [3.4] DecommissionAction.Execute, idempotent end-to-end

platform/orchestration/actions/             [3.5 + 3.6 chassis-side]
├── thunder_decommission_dispatch.go        [3.5] DispatchThunderDecommissionAction — publishes decommission_instance to thunder-adapter, returns AwaitResponse:true
├── thunder_provision_dispatch.go           [3.6] DispatchThunderProvisionAction — publishes provision_instance, same pattern
└── registry.go                             [patched 3.5 + 3.6] Both actions registered under category "training", IsLocal: true

deployments/kustomize/services/thunder-adapter/base/
├── deployment.yaml                         [patched Phase 2] serviceAccountName, imagePullSecrets, command:
├── rbac.yaml                               [3.2] Role + RoleBinding for SSH Secret access
├── kustomization.yaml                      [patched 3.2] references rbac.yaml
├── service.yaml
└── thunder-adapter.yaml                    [patched 3.0] correct base URL

migrations/
├── 028_thunder_reaper.sql                  [3.5] agent_definition + scheduled_tasks row
└── 029_gpu_provisioner_real_impl.sql       [3.6] UPDATE agent_definitions for gpu-provisioner
```

---

## 3. Key design decisions

### Identifier strategy
- DB row has its own UUID (`id`), generated client-side before INSERT.
- Thunder API's numeric `identifier` stored as TEXT in `thunder_instance_id` column.
- SSH Secret name is deterministic: `thunder-ssh-<db-uuid>` — uses the DB UUID, not Thunder's UUID. Means the Secret name is known before the API call, so an orphan Secret (if INSERT fails) is reapable.

### Two interfaces for testability
`provision_action.go` defines small interfaces `thunderAPI` and `secretManager` covering only the methods the action calls. Tests inject mocks; production code passes the concrete `*api.Client` and `*ssh.SecretManager`. Same interfaces reused by `decommission_action.go`.

### Compensating cleanup uses fresh context
If `provision_instance` partially succeeds (Thunder create OK but later step fails), the compensating delete-instance + delete-Secret runs under `context.Background()` with a 30s timeout. Cleanup must run even if the parent context has expired.

### Idempotency anchors
- `provision`: pre-generated DB UUID makes the Secret name deterministic. If INSERT fails after Secret creation, cleanup knows which Secret to delete.
- `decommission`: `MarkDecommissioning` atomically transitions `running`/`provisioning` → `decommissioning`. Rows already in `decommissioned`/`failed`/`reaped` return immediately with cached cost/timestamp.

### Error classification
- `isProvisionDenial` → `error_unrecoverable` (cap breach, paused — don't retry)
- `isInfrastructureError` → `error_recoverable` (5xx, rate-limit, ctx deadline — chassis retry policy applies)
- Other → `error_unrecoverable` (default safe)

### Cost computation
`hourly_rate_usd` snapshotted into the `thunder_instances` row at provision time. If `thunder_config.default_hourly_rate_usd` changes later, in-flight instances keep their original rate (accurate per-rate-at-provision attribution).

---

## 4. Deployment status

Phase 2 is the latest deployed version (`docker.io/aqls/thunder-adapter:v1.0.1010`). Phase 3.0–3.4 code is delivered but not yet built or pushed.

### Required before next deploy
- [ ] Verify go.mod has `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` (likely already present; check with `grep`).
- [ ] Confirm `THUNDER_COMPUTE_API_KEY` is set in `personae-default-secrets`. The patched `NewAdapter` returns an error at startup if it's empty.
- [ ] Run `go test ./internal/adapters/thunder/...` locally. All unit tests should pass; integration tests deferred.
- [ ] Build: `make build-thunder-adapter IMAGE_TAG=v1.0.1013 && docker push docker.io/aqls/thunder-adapter:v1.0.1013`
- [ ] Apply: `kubectl apply -k deployments/kustomize/services/thunder-adapter/overlays/production/uk_001/`
- [ ] Verify RBAC: `kubectl -n ai-persona-system get role,rolebinding | grep thunder-adapter`
- [ ] Verify startup logs show `thunder_api_url=https://api.thundercompute.com:8443/v1` and `ssh_namespace=ai-persona-system`.

---

## 5. Verification sequence (do this before relying on the adapter)

Before sending real provisioning traffic, do a deliberate end-to-end test with low blast radius.

### Step 1: Cap the spend temporarily
```sql
UPDATE thunder_config SET daily_cap_usd = 5;
-- 5 USD ≈ 2.5h of single A100 prototyping. Bounds the risk if anything misbehaves.
```

### Step 2: Provision a single instance via kcat
```bash
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

kubectl -n kafka run kcat-provision-$(date +%s) \
  --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.thunder.requests \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_type=request \
    -H client_id=demo_client \
    -H step_name=manual_provision_test \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<JSON
{"body":{"action":"provision_instance","reply_to_topic":"system.thunder.smoke.responses","training_run_id":"","gpu":"a100","mode":"prototyping"}}
JSON

# Watch adapter logs — expect ~3-5 min until "Provision complete"
kubectl -n ai-persona-system logs deploy/thunder-adapter --tail=0 -f
```

### Step 3: Verify the DB row
```sql
SELECT id, thunder_instance_id, status, instance_ip, ssh_user, ssh_key_secret_name,
       running_since, hourly_rate_usd, requested_by
FROM thunder_instances
ORDER BY created_at DESC
LIMIT 1;
-- Expect: status='running', instance_ip populated, ssh_user='ubuntu',
--         ssh_key_secret_name='thunder-ssh-<uuid>', requested_by='cli'
```

### Step 4: Verify the SSH Secret exists in k8s
```bash
kubectl -n ai-persona-system get secret -l app.kubernetes.io/managed-by=thunder-adapter
# Expect: thunder-ssh-<uuid> with the right labels
```

### Step 5: Decommission via kcat
```bash
PROVISIONING_ID="<uuid-from-step-3>"

# Same kcat shape as step 2, body changes:
... <<JSON
{"body":{"action":"decommission_instance","reply_to_topic":"system.thunder.smoke.responses","provisioning_id":"$PROVISIONING_ID","reason":"manual_verification"}}
JSON
```

### Step 6: Verify decommission completed
```sql
SELECT id, status, cost_usd, decommissioned_at,
       EXTRACT(EPOCH FROM (decommissioned_at - running_since))/3600 AS uptime_hours
FROM thunder_instances
WHERE id = '<provisioning-id>';
-- Expect: status='decommissioned', cost_usd > 0, decommissioned_at populated
```

```bash
kubectl -n ai-persona-system get secret thunder-ssh-<uuid>
# Expect: NotFound
```

### Step 7: Restore the cap
```sql
UPDATE thunder_config SET daily_cap_usd = 100;
```

### Manual cleanup fallback
If something gets stuck mid-way, manually clean up:
```bash
# Via Thunder CLI from a workstation with the token in ~/.thunder/token
tnr delete <numeric_identifier>

# Manually delete the orphan Secret
kubectl -n ai-persona-system delete secret thunder-ssh-<uuid>

# Update DB row to reflect manual cleanup
UPDATE thunder_instances
SET status='decommissioned', decommissioned_at=NOW(), cost_usd=0
WHERE id='<provisioning-id>';
```

---

## 6. Outstanding TODOs

### Within Phase 3
- ✅ **3.5 thunder-reaper** — delivered and verified 2026-05-14.
- ⏳ **3.6 gpu-provisioner real impl** — delivered as code; build chassis v1.0.1014, push, apply migration 029, then exercise.

### Verification-blocked (do after first real provision works)
- `api/types.go` has three TODOs flagged for verification on first real call: `ListInstances` response wrapper shape, `Instance` JSON field names, instance status string casing. All testable with `kubectl logs` after the first real provision — JSON decode errors will be obvious.
- Comprehensive integration tests for `ProvisionAction.Execute` and `DecommissionAction.Execute` using mock thunderAPI/secretManager and sqlmock. Deferred to keep test surface small until shape is confirmed.

### Architectural follow-ups (not blocking)
- **Observability filter gotcha (logged 2026-05-14)**: filtering `orchestration_states` by `owner_agent_type = '<agent>'` MISSES top-level chassis-resident workflows, which are owned by `'generic'`. The actual agent_type lives in `collected_data->'config'->>'agent_type'` and the orchestration_name follows `sched-<task-name>-<timestamp>`. Use one of these instead:
  ```sql
  WHERE collected_data->'config'->>'agent_type' = 'thunder-reaper'
     OR orchestration_name LIKE 'sched-thunder-reaper-%'
  ```
  Same pattern applies to any scheduler-fired chassis-resident agent (build-pipeline-trigger, work-item-archiver, thunder-reaper, etc.). Worth a section in the debugging guide if we hit it again.
- Per-error-type retry policy for the DB INSERT in `provisionAction.insertWithRetry`. Currently retries all errors with backoff — fine in practice (constraint violations on pre-generated UUID are impossible) but could be tighter.
- `thunder_instances.thunder_instance_id` is TEXT (per migration 025's schema comment "from Thunder API e.g. 'ti_abc123'") but we store integer-as-string. If we ever want to query by Thunder UUID rather than numeric identifier, we'd add a separate `thunder_uuid` column.
- Stub for `training-launcher` agent definition (migration 023) still returns mock data. After thunder-adapter is fully verified (provision + decommission against real Thunder API), replace with real implementation that SSH-execs the training script on the provisioned VM. Not blocking 3.6.

---

## 7. Docs updated this cycle

- `016_debugging_guide_v2.md` — new section 10 on adapter deployment failure modes (image-pull insufficient_scope, command-vs-args trap, Strimzi auto-create-off, deployment essentials checklist). Old section 10 renumbered to 11.
- `FOCUS_adapter_design.md` — rewrote "Deployment essentials" section with: full annotated manifest pattern, required cluster resources table, Kafka topic provisioning gotcha, scoped RBAC example, makefile integration notes, pre-deploy and post-deploy checklists. Cross-references debugging guide section 10.

---

## 8. Last touched

2026-05-14 — Phase 3.5 verified end-to-end via synthetic reaper test (decommission action ran clean, cost computed at $3.61 for 2h × $1.80, idempotent 404 handling for both API and Secret). Phase 3.6 (gpu-provisioner real impl) delivered as code — chassis v1.0.1014 build pending, then migration 029 application, then the real provision verification (Option 1 from the cycle plan).

---

## Update 2026-06-04 — thunder-training-monitor (built, verifying) + a reply-topic finding

A second periodic lifecycle agent now sits alongside the reaper: **`thunder-training-monitor`**
(orchestrator) + **`thunder-training-monitor-worker`**. It closes two gaps the reaper does not:
release of a *detached* training box on completion, and the `running → complete/failed`
reconcile that otherwise never happens (why `1cd65dd7` / `e6ab9fad` sat `running`).

Shape (deliberately NOT the reaper's scheduler-pre_query shape — the scheduler merges only
the first pre_query row and fires once per tick, which would starve newer instances because
ALIVE training boxes stay `running`): the orchestrator finds ALL running training instances
(`find_active_training_instances`) and loops `spawn_worker → call_worker` (sequential await).
Each worker probes the box via the adapter's **`ssh_get_status`**, classifies run.sh markers
(`ALIVE | DONE_OK | DONE_FAIL | GONE_UNKNOWN`), reconciles `training_runs`, and on a terminal
verdict releases the box via the existing **`dispatch_thunder_decommission`** — reusing the
idempotent decommission end-to-end.

Artifacts: migrations 106 (`thunder_instances.consecutive_unreachable_probes` + `last_probe_at`),
107 (worker def), 108 (orchestrator def + a `scheduled_tasks` row, **inserted DISABLED**), and 5
chassis actions (`dispatch_thunder_ssh_get_status`, `classify_training_probe`,
`record_probe_streak`, `mark_training_run_terminal`, `find_active_training_instances`).

**Deploy/verify state:** deployed, defs `active`. A manual worker test reached the `probe` step
and dispatched `ssh_get_status` to thunder-adapter, then the await **did not resolve** — the
coordinator registered the awaited request against the orchestration's OWN responses topic
(`determineResponsesTopic` → env `RESPONSES_TOPIC` / `execCtx.ResponsesTopic` =
`system.agent.generic.responses`), but `dispatch_thunder_ssh_get_status` (cloned from
`dispatch_thunder_ssh_exec`) put `__parent_responses_topic__` (= `system.generic.responses`,
the CLI's reply topic, unconsumed) into the adapter request envelope → reply routed to a topic
the coordinator was not watching → orphaned await. **Fix:** the dispatch action now prefers
`ExecutionContext.ResponsesTopic` (matches the coordinator in both the spawned and generic-entry
paths), `__parent` only as fallback. **NOTE for the adapter family:** `dispatch_thunder_ssh_exec`
(and the other thunder dispatches) still prefer `__parent_responses_topic__`; they work today only
because they are called from spawned children where the two topics coincide. If any is ever fired
top-level, it will orphan the same way — candidate for a shared `resolveAwaitResponsesTopic` helper.
Pending: confirm against thunder-adapter logs that the reply went to the dead topic, redeploy, re-test.

**Confirmed 2026-06-04 (full chassis state dump).** The reply-topic orphan above is
confirmed on the chassis side: the coordinator registered the awaited request on
`system.agent.generic.responses` (`determineResponsesTopic` #1 = env `RESPONSES_TOPIC`,
overriding the action's result topic) while the old dispatch put
`__parent_responses_topic__` = `system.generic.responses` (CLI-derived, unconsumed) in
the adapter envelope. Adapter-facing request id is `a9e722e8` (grep the adapter for
that, not the orchestration's `3e9254e3`). Also observed: the pod served a STALE cached
worker definition (an old reaper-style no-op stub) for the message's `agent_config`
while executing the full `WorkflowPlan` — i.e. 107 took effect; a redeploy clears the
cache. Remaining confirmation is purely the adapter log (did it reply, and where).

**VERIFIED 2026-06-04 18:21.** After the fix + redeploy, the worker's ALIVE path ran
end-to-end: probe → thunder-adapter `STATUS=ALIVE`/`reachable:true` (await resolved) →
classify `alive` → reset_streak → done. `thunder_instances.last_probe_at` updated,
counter 0, status running, no decommission. The adapter's `ssh_get_status` + the probe
`status_command` work against the live box. (Stub `agent_config` note above is amended:
it persisted across the redeploy, so it is a persistent source, not in-pod cache —
DB-level investigation, cosmetic only.)
