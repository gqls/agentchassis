# Pre-trigger runbook — iter_0 automated training launch

Goal: get the first automated `model-trainer → … → training-launcher` run to
start training on a Thunder VM. Two code changes must ship before triggering
(neither is in the currently-deployed build); the migration is already applied.

Namespace: `ai-persona-system`. Kafka cluster: `personae-kafka-cluster` (`-n kafka`).

---

## 0a. STALE FILES — do not re-apply

The `102_training_launcher_real.sql` and `HANDOFF_2026-05-24` in the repo/uploads
are the **pre-revision** drafts: that 102 still has `input_mapping` on the local
steps, a `nohup` (not `setsid`) command, `output_field` (singular), and no
`is_active` scoping. The **DB already has the revised version** (verified: no
local-step `input_mapping`, `setsid` command, `output_fields` plural). **Do not
re-run the uploaded 102** — it would revert the launcher to the broken shape. The
correct migration is the revised file produced this session. Likewise treat the
2026-05-24 handoff as superseded (it's the source of the `prepare_object_url`-added
and `__parent_responses_topic__` claims that turned out wrong). Confirm the deployed
chassis (`v1.0.1049`) was built from the **revised** action code (own-topic
derivation + `resolveTemplateToken`), not the handoff-era files — the §4 round-trip
plus the first run's `await_responses_topic` log line are the confirmations.

## 0. State going in (verified this session)

- Migration `102` is applied and correct: launcher steps carry no `input_mapping`;
  `presign_dataset={method}`, `ssh_exec_launch` has the `scripts_url`/`dataset_url`
  dot-paths + `setsid` command, `mark_running={training_run_id}`, `complete` uses
  `output_fields`. `model-trainer.call_launcher` maps `provisioning_id`. **Do not
  re-run 102.**
- Bucket resolved: adapter defaults `TRAINING_BUCKET=personae-model-training`;
  preparer writes there; dataset key aligns after the `s3://` strip.
- Two gaps that block the run, fixed in code but NOT yet deployed:
  1. **Adapter has no `prepare_object_url`** (deployed `thunder-adapter:v1.0.1048`
     is Phase-4: dataset/artefact-by-ID only) → `presign_dataset` returns
     `not_implemented`. Fix: `data_url_actions.go` + `adapter.go`.
  2. **Launcher dispatch actions route adapter replies to the parent topic**
     (`__parent_responses_topic__`) instead of the own topic that provision/
     decommission use → the await would hang. Fix: `thunder_ssh_exec_dispatch.go`
     + `thunder_prepare_object_url_dispatch.go`.

---

## 1. Apply the code

Adapter (thunder-adapter image):
- `internal/adapters/thunder/data_url_actions.go` — adds `ObjectURLRequest`,
  `ObjectURL` (presign by explicit key, GET/PUT), `handlePrepareObjectURL`;
  `DatasetURL`/`ArtefactURL` now delegate to `ObjectURL`.
- `internal/adapters/thunder/adapter.go` — adds `case "prepare_object_url"`.

Chassis (agent-chassis image):
- `platform/orchestration/actions/thunder_ssh_exec_dispatch.go` — reply topic
  now `ExecutionContext.ResponsesTopic` (own topic) + agent-type fallback.
- `platform/orchestration/actions/thunder_prepare_object_url_dispatch.go` — same
  topic fix (this file also already carries the `dataset_uri` fallback).

`coordinator.go` is unchanged — the earlier orchestrator change was withdrawn.

## 2. Build-time checks (catch what couldn't be checked here — no Go toolchain)

```
go build ./...        # must pass; confirms the four edits compile in-tree
go vet ./internal/adapters/thunder/... ./platform/orchestration/actions/...
```

Regression to watch for: `DatasetURL`/`ArtefactURL` now delegate to `ObjectURL`.
Behaviour is meant to be identical (same keys, same GET/PUT, same default
expiries). If there's a unit test for `prepare_dataset_url`, run it; otherwise
the post-deploy round-trip in step 4 covers it.

## 3. Deploy + confirm running images

Build/push both images with new tags (above `v1.0.1048` for the adapter; above
the current chassis tag), roll the deployments, then confirm what's actually
running:

```
kubectl -n ai-persona-system get deploy -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[0].image}{"\n"}{end}' \
  | grep -iE 'thunder|launcher|trainer|chassis'
```

Pass: `thunder-adapter` on the new tag; the `training-launcher` and
`model-trainer` pods (agent-chassis) on the new chassis tag. Confirm pods are
Ready:

```
kubectl -n ai-persona-system get pods | grep -iE 'thunder|launcher|trainer'
```

## 4. Upload the scripts bundle + adapter round-trip

### 4a. Upload the bundle to B2

The launcher presigns `finetuning/scripts/bundle.tar.gz` and the VM curls it; the
object must exist at that exact key in `personae-model-training`. Build the
tarball flat (so `run.sh` etc. sit at the archive root, matching the launch
command's `tar -xzf … -C /workspace`), then upload with the real `b2` CLI
(`pip install b2`, not the snap). Auth from `personae-storage-secrets` rather
than hardcoding keys.

```
# (build the bundle flat from the script dir)
tar -czf bundle.tar.gz run.sh 00_vm_setup.sh 02_train_llama_3_3_70b.py 03_inference_test.py

# auth (prefer pulling these from the k8s secret, not literals)
b2 account authorize "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY"

# upload to the EXACT key the migration's presign_scripts expects
b2 file upload personae-model-training ./bundle.tar.gz finetuning/scripts/bundle.tar.gz

# verify — NOTE this b2 CLI (v3+) wants a b2:// URI, not two bare args:
b2 ls --long "b2://personae-model-training/finetuning/scripts/"   # expect a bundle.tar.gz row
```

To update the launch chain later, edit `run.sh`, rebuild, re-upload to the SAME
key — no DB migration, no chassis redeploy.

### 4b. Adapter round-trip

Confirms three things without provisioning a VM: the deployed adapter has
`prepare_object_url` (not `not_implemented`), the reply key is `presigned_url`
(what the launcher tokens resolve), and the bundle is fetchable. Manual produce +
manual consume — does NOT depend on chassis await matching, so it isolates the
adapter from the topic question.

Envelope shape (verified against adapter.go handleMessage L309–314): the adapter
reads `action` and `reply_to_topic` from the **value's `body`**, and
`correlation_id`/`request_id` from the **Kafka `-H` headers**. So `action`,
`reply_to_topic`, and the B2 `key` all go inside `body`. (Same shape as the
`prepare_dataset_url` test; only `key` replaces `export_id`.)

```
# Terminal 1 — start the reply consumer FIRST, decode the URL through jq
# (kcat -f '%s' escapes & as \u0026; jq turns it back into & so curl works):
kubectl -n kafka run kcat-objurl-rx --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.responses -C -o end -q -f '%s\n' \
  | jq -r 'select(.body.action=="prepare_object_url") | .body.presigned_url'

# Terminal 2 — produce the prepare_object_url request (GET on the bundle key):
echo '{"body":{"action":"prepare_object_url","key":"finetuning/scripts/bundle.tar.gz","method":"GET","reply_to_topic":"system.agent.generic.responses"}}' | \
kubectl -n kafka run kcat-objurl-tx-$(date +%s) --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.thunder.requests -P \
  -H "request_id=objurl-test-$(date +%s)" \
  -H "correlation_id=objurl-test" \
  -H "message_type=request"
```

Then fetch it. IMPORTANT: the URL is signed for **GET** (`x-id=GetObject`).
`curl -I` sends a **HEAD**, which changes the SigV4 canonical request and returns
`403 SignatureDoesNotMatch` even though the URL is valid — use a GET:

```
curl -s -o /dev/null -w '%{http_code}\n' "<decoded URL>"        # expect 200 (file is ~7 KB)
# or a 1-byte range to avoid downloading: curl -r 0-0 -s -o /dev/null -w '%{http_code}\n' "<url>"  → 206
```

Failure reads:
- reply `status:"error"` `code:"not_implemented"` → the deployed adapter image
  doesn't carry `prepare_object_url`; rebuild/redeploy the adapter.
- GET → 403: get the real cause from the XML body (`curl -s "<url>"`):
  `SignatureDoesNotMatch` = signing/method/clock; `AccessDenied` = the B2 key
  lacks read on that key. (A `curl -I` 403 is just the HEAD-vs-GET mismatch — not
  a real failure; re-test with GET.)
- GET → 404 (`NoSuchKey`) → bundle not at that key / wrong bucket; re-check 4a.
- no reply at all → `reply_to_topic` wasn't in `body`, or the request never
  reached the adapter topic.

Bundle contents verified this session: flat layout (`run.sh`, `00_vm_setup.sh`,
`02_train_llama_3_3_70b.py`, `03_inference_test.py`); `run.sh` runs
`00_vm_setup.sh` in its `setup` step and both place the venv at
`${HOME}/unsloth_env`; `DATA=/workspace/training_iter0.jsonl` matches the launch
command's curl target; train args (`--data`/`--output`, `--epochs` default 3,
`--limit` 0=all) match the manifest.

## 5. Cost gate

The gate is a view: `SELECT can_provision, denial_reason FROM thunder_provision_check;`
— run it AFTER cleaning up any orphaned `running` instance (an orphan counts
against `max_concurrent_instances=2` and against `thunder_spend_24h`). Expect
`t | NULL`.

`thunder_config` today: `daily_cap_usd=15`, `estimated_new_run_cost_usd=2`,
`default_hourly_rate_usd=1.80`, `default_hard_uptime_hours=18`. The manual iter_0
ran ~9.2h of GPU; with setup + the 70B-4bit download, instance uptime is ~10h ≈
**$18** — above the $15 cap, and `estimated_new_run_cost_usd=2` is unrealistic.
So check what `thunder_provision_check` actually compares before changing a value:
if it tests `spend_24h + estimated_new_run_cost_usd <= daily_cap_usd`, the current
`2` lets it pass but the run silently exceeds the cap; an honest `~18` would block
against `15`. Either way, make the estimate real and raise the cap above one run
with headroom — the 9.2h run is safely under the 18h reaper kill (~$32 ceiling
per instance). Provisional (confirm against the view first):

```
UPDATE thunder_config SET estimated_new_run_cost_usd = 20, daily_cap_usd = 30 WHERE singleton = 'X';
SELECT can_provision, denial_reason FROM thunder_provision_check;   -- re-check: expect t | NULL
```

DB cleanup of a failed run marks the row but does NOT terminate the VM: after
`UPDATE thunder_instances SET status='decommissioned'`, confirm the actual Thunder
instance is gone (console/API or a real `decommission_instance` dispatch), else it
keeps billing.

## 5b. Provision smoke test (validates the await/topic path — the D4 gate)

Fires `gpu-provisioner` standalone (one a100, prototyping) BEFORE the full run.
This is the real test of whether `v1.0.1049` carries the own-topic fix (D4): the
provision dispatch uses the same reply-topic derivation the launcher does, so if
its await resolves, the launcher's will too. The 2026-05-27/17:10 run hung here
(stuck `awaited_requests`), so read the topic line carefully.

Provisions a real billable instance — decommission it for real afterward (5d).

Preconditions: `SELECT can_provision, denial_reason FROM thunder_provision_check;`
→ `t | NULL`. Any earlier orphaned instance actually terminated (not just
DB-marked).

Open three watchers first:

```
# A — adapter:
kubectl -n ai-persona-system logs deploy/thunder-adapter -f --since=10s \
  | grep -vE "Failed to fetch message from Kafka|context deadline exceeded"

# B — chassis (where gpu-provisioner runs): the dispatch's await_responses_topic
kubectl -n ai-persona-system logs deploy/agent-chassis -f --since=10s \
  | grep -iE "Dispatched provision_instance|await_responses_topic|gpu-provisioner"

# C — DB poll (set ORCH after firing, bake it in — do NOT escape it):
#   ORCH=<the orch uuid printed by the fire>
#   watch -n 5 "kubectl -n ai-persona-system exec postgres-clients-0 -- \
#     psql -U clients_user -d clients_db -c \"
#   SELECT 'awaited' k, status, sent_at, processed_at, processed_at-sent_at latency
#     FROM awaited_requests WHERE orchestration_id='$ORCH'
#   UNION ALL
#   SELECT 'instance', status, created_at, decommissioned_at, NULL
#     FROM thunder_instances ORDER BY 3 DESC LIMIT 6;\""
```

Fire (corrected: no leading `kcat`, no `-c 1`, `-k` = correlation_id):

```
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ   (write these down)"

kubectl -n kafka run kcat-prov-$(date +%s) --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -k "$CORRELATION" \
  -H correlation_id=$CORRELATION \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_type=request \
  -H action=orchestrate \
  -H client_id=$CLIENT_ID \
  -H step_name=manual_provision_test \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H from_agent_type=user \
  -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
JSON
```

Read (in order):
- Watcher B, the decisive line — `Dispatched provision_instance to thunder-adapter
  … await_responses_topic=…`:
  - `system.responses.gpu-provisioner` → D4 is live. Good — proceed.
  - `system.agent.gpu-provisioner.responses` → the §6 fallback fired;
    `__my_responses_topic__` isn't seeded → the reply won't match → it will hang
    exactly like 17:10. Stop; the chassis needs the seeding/derivation sorted
    before iter_0.
- Watcher A: adapter logs `provision_instance`, creates the instance (3–5 min:
  Thunder create + WaitForRunning), sends a success reply.
- Watcher C: instance row goes `…→running`; the `awaited` row goes
  `waiting → done/processed` with a finite latency. A `waiting` row that never
  clears (while the adapter logged a reply) = the topic mismatch above.

Clean up — decommission for REAL (this actually calls Thunder DeleteInstance,
unlike a DB-only UPDATE). Use the new instance's `provisioning_id` (the
`thunder_instances.id` from watcher C):

```
echo '{"body":{"action":"decommission_instance","provisioning_id":"<instance id>","reply_to_topic":"system.agent.generic.responses","reason":"provision smoke test cleanup"}}' | \
kubectl -n kafka run kcat-decomm-$(date +%s) --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.thunder.requests -P \
  -H "request_id=decomm-$(date +%s)" -H "correlation_id=decomm-test" -H "message_type=request"
# then confirm on Thunder (console/API) the VM is actually gone.
```

Only once watcher B shows `system.responses.gpu-provisioner` and the await
resolved is it worth triggering the full iter_0 below.

## 6. Trigger iter_0

Trigger `model-trainer` with your existing mechanism (the one used to script
flywheel C), with `input_data` containing at least:
- `export_id`: `146a9a12-…`
- `hyperparameters`: the iter_0 set (read at `input_data.hyperparameters` by the
  preparer; the launcher passes it through)

The orchestrator chain is: `spawn_data_preparer → spawn_provisioner →
spawn_launcher → call_data_preparer → call_provisioner → call_launcher →
complete`.

## 7. Watch (markers + failure signatures)

Chassis logs (model-trainer, then launcher):
```
kubectl -n ai-persona-system logs -f deploy/model-trainer
kubectl -n ai-persona-system logs -f deploy/training-launcher
```
Expect, in order:
- preparer: dataset uploaded, `training_runs` row INSERTed `pending`.
- provisioner: 3–5 min, returns `provisioning_id` (+ ip/user/secret).
- launcher steps: `presign_dataset → presign_scripts → ssh_exec_launch →
  mark_running → complete`, each "Dispatched … to thunder-adapter" then resuming
  on the reply.

Adapter logs:
```
kubectl -n ai-persona-system logs -f deploy/thunder-adapter
```
Expect two `prepare_object_url` presigns (dataset, bundle) and one `ssh_exec`
that returns quickly with `LAUNCH_PID=` in stdout.

On the VM (via an `ssh_get_status` or your access), then training:
- `/workspace/launch.log` — curl/tar of bundle + dataset (empty/clean on success).
- `/workspace/train.log` — `RUN_SH_START` → `RUN_SH_STEP step=setup` →
  `RUN_SH_SMOKE_OK` → `RUN_SH_STEP step=full_train` → … → `RUN_SH_FULL_OK` →
  `RUN_SH_DONE`.

Failure signatures and where they point:
- Adapter reply `not_implemented` on `prepare_object_url` → adapter image didn't
  pick up the new action (step 3).
- `presign_dataset` dispatched, adapter logged a successful presign, but the
  launcher never advances and times out (~600s) → reply-topic mismatch (step 1
  fix not deployed).
- `RUN_SH_FATAL missing_required_file=…` in train.log → bundle/path issue.
- curl error in `launch.log` (no train.log) → presigned URL 404: bucket/key
  mismatch or bundle not uploaded (step 4).
- `ssh_exec` blocks ~5 min then errors → the launch command didn't detach
  (setsid hardening not deployed).
- smoke fails (`RUN_SH_SMOKE_OK` absent, error in train.log) → data format / OOM;
  the smoke pass did its job of stopping before the full run.

## 8. After launch

`ssh_exec` returns fast; `mark_running` flips `training_runs` to `running`; the
launcher `complete`s and returns `launch_result` to model-trainer. The real
train then runs **detached** for ~9h (1958 examples × 3 epochs, per the
manifest), outside any await window. There is no training-monitor yet, so the
`running → complete/failed` transition won't happen automatically — watch
`train.log` for `RUN_SH_FULL_OK`, and note the artefact-upload path
(`prepare_artefact_url` PUT) is still untested until `adapter_out` exists.
