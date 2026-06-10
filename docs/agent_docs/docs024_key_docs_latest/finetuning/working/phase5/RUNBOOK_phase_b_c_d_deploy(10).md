# RUNBOOK — deploy & test the checkpoint/adapter upload path (steps 2–6)

Covers the upload-manifest path (Phases A/B/C) and the resume path (Phase D), in the
order they must go live. Companion to `PLAN_checkpoint_and_artefact_upload_b2.md`.

**Step 1 (you are doing this): `go build` the chassis with the four Phase B Go files +
the three registry entries + the 2026-06-08 register-before-send edit in
`thunder_prepare_object_url_dispatch.go` (pre-register the await before `ProduceWithValidation`
— the loop-await race fix; closes the intermittent presign_checkpoints stall), build the
adapter image (it now has `prepare_resume_url`), and deploy both. Bump the chassis tag from
`v1.0.1057`.** Everything below assumes that image is running.

## Staging principle (why the order matters)
Prove the **upload path in isolation first** (steps 2–4), then enable the monitor (step
6). The **resume test (step 5) is deliberately last and is blocked on code that is not
built yet** (`dispatch_thunder_prepare_resume_url` + migration `110`). Resume reads the
checkpoints the upload path writes, so it tells you nothing until the upload path is
green. Do **not** apply `110` before step 4 — it would put `check_resume` into every
launch, including the upload test, mixing the two.

## Environment quick-reference
- App pods: `kubectl -n ai-persona-system get pods`. Kafka: `kubectl -n kafka get pods`.
- DB: flywheel-C `agent_definitions`, `model_lifecycle.training_runs`, the `thunder_*`
  tables, and `scheduled_tasks` are in **`clients_db`**. Connect with your usual psql.
  (Confirm `scheduled_tasks`' DB if your scheduler is split out — see step 6.)
- B2: bucket `personae-model-training`, endpoint `https://s3.us-east-005.backblazeb2.com`,
  region `us-east-005`. Scripts bundle key: `finetuning/scripts/bundle.tar.gz` (confirm
  in step 3). Checkpoints: `finetuning/checkpoints/<run_id>/ckpt-<N>.tar.gz`. Final
  adapter: `finetuning/artefacts/<run_id>/adapter.tar.gz`.
- B2 access: use the native **`b2`** CLI (the `aws` CLI fails here with "Unable to locate
  credentials" unless you export the B2 key as AWS creds — see fallback below). Commands
  below show b2 **v4** (`b2 file upload`, `b2 ls b2://…`) with the **v3** form noted
  (`b2 upload-file`, `b2 ls <bucket> <path>`); check `b2 version`. If already authorized
  (you use b2 routinely) no auth step is needed; otherwise
  `b2 account authorize "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY"` (v4) /
  `b2 authorize-account …` (v3).
  - **aws fallback:** B2 is S3-compatible, so the `aws --endpoint-url
    https://s3.us-east-005.backblazeb2.com s3 …` form works *if* you first
    `export AWS_ACCESS_KEY_ID="$B2_APPLICATION_KEY_ID"
    AWS_SECRET_ACCESS_KEY="$B2_APPLICATION_KEY"`.

## Pre-flight (independent of these steps, do once)
Reconcile the stale run and sweep for orphans so the test run is the only thing live:
```sql
-- in clients_db. Confirm the full id first.
SELECT id, status, started_at FROM model_lifecycle.training_runs WHERE status='running';
UPDATE model_lifecycle.training_runs
   SET status='failed',
       error_message='superseded by iter_0 1cd65dd7; box gone — reconciled by hand',
       completed_at=now()
 WHERE id='e6ab9fad…' AND status='running';   -- expand the id

SELECT id, thunder_instance_id, status FROM thunder_instances
 WHERE status NOT IN ('decommissioned','reaped','failed');
-- decommission/reap anything stale via your normal path before testing.
```

---

## Step 2 — Verify the live launcher def, then apply `109`

**Goal:** confirm the live `training-launcher` matches what `109` assumes, then wire in the
six new steps. **Apply only after the chassis from step 1 is deployed** — if the workflow
references `compute_checkpoint_keys` / `flatten_presign_results` / `assemble_upload_manifest`
before the chassis has them, those steps fail "unknown action".

### 2a. Verify (from the `109` header)
```sql
-- in clients_db
-- presign_scripts must exist and carry scripts_url/dataset_url at TOP-LEVEL config
-- (not under input_mapping) — resolveTemplateToken reads top-level config[token].
SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_scripts}')
  FROM agent_definitions
 WHERE type='training-launcher'
   AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- informational (the run.sh wiring relies on ssh_exec_launch being untouched):
SELECT jsonb_pretty(default_config #> '{workflow,steps,ssh_exec_launch,config}')
  FROM agent_definitions
 WHERE type='training-launcher'
   AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
```
Pass criteria: `presign_scripts` exists; its `next_step` is whatever currently follows
it (we are about to repoint it at `compute_keys`); `scripts_url`/`dataset_url` are at the
top level of the relevant config. If `presign_scripts` is absent or named differently,
**stop** — `109` repoints `presign_scripts.next_step` and would no-op silently otherwise.

### 2b. Apply
```bash
psql "<clients_db connection>" -f 109_launcher_upload_manifest_wiring.sql
```
It is a single idempotent transaction; re-running it is safe.

### 2c. Verify (the `109` footer queries)
```sql
-- presign_scripts now enters the manifest path (expect "compute_keys"):
SELECT default_config #>> '{workflow,steps,presign_scripts,next_step}'
  FROM agent_definitions WHERE type='training-launcher'
   AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- the six new steps exist and chain to ssh_exec_launch:
SELECT s.key AS step, s.value->>'action' AS action, s.value->>'next_step' AS next_step
  FROM agent_definitions a,
       jsonb_each(a.default_config #> '{workflow,steps}') s
 WHERE a.type='training-launcher'
   AND (a.is_snapshot IS NULL OR a.is_snapshot=false) AND a.deleted_at IS NULL
   AND s.key IN ('compute_keys','presign_checkpoints','flatten_checkpoint_urls',
                 'presign_final','assemble_manifest','write_manifest')
 ORDER BY s.key;
```
Pass criteria: `presign_scripts.next_step = compute_keys`; the six rows present with
`write_manifest.next_step = ssh_exec_launch` and `assemble_manifest.next_step =
write_manifest`. Also confirm the loop's terminal exists:
```sql
SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints,config,sub_workflow,steps}')
  FROM agent_definitions WHERE type='training-launcher'
   AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
```
Expect two substeps: `presign_one` (`dispatch_thunder_prepare_object_url`,
`next_step: presign_done`) and `presign_done` (`loop_complete`).

**Rollback:** `109` only adds six keys + repoints one edge. To revert, repoint
`presign_scripts.next_step` back to its original value and delete the six keys (keep a
copy of the 2a dump first).

---

## Step 3 — Re-pack `bundle.tar.gz` and re-upload to B2

**Goal:** ship the edited `02_train_llama_3_3_70b.py` (Phase A) and `run.sh` (Phase C) to
the VM. The launcher curls this bundle and `tar -xzf` into `/workspace`, so the files
must be at the **bundle root** (extract to `/workspace/02_train_llama_3_3_70b.py`, etc.).
This is the "did it ship" check — analogous to a `go build` for the on-VM scripts.

### 3a. Confirm the bundle key the launcher actually fetches
```sql
-- in clients_db — find where presign_scripts gets its key/uri from:
SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_scripts,config}')
  FROM agent_definitions WHERE type='training-launcher'
   AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
```
Match the key below to whatever this signs (documented as
`finetuning/scripts/bundle.tar.gz`). If it differs, upload to *that* key.

### 3b. Re-pack (files at root, not in a subdir)
```bash
mkdir -p /tmp/bundle && cd /tmp/bundle
# copy in the THREE on-VM scripts (00_vm_setup.sh unchanged from the last bundle):
cp /path/to/00_vm_setup.sh .
cp /path/to/02_train_llama_3_3_70b.py .    # Phase A edited copy
cp /path/to/run.sh .                       # Phase C edited copy
chmod +x run.sh 00_vm_setup.sh
tar -czf bundle.tar.gz 00_vm_setup.sh 02_train_llama_3_3_70b.py run.sh
tar -tzf bundle.tar.gz    # MUST list the three files at the root, no leading dir
```

### 3c. Upload + verify
```bash
# b2 v4:
b2 file upload personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz
b2 ls --long "b2://personae-model-training/finetuning/scripts/"
# b2 v3 equivalents:
#   b2 upload-file personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz
#   b2 ls --long personae-model-training finetuning/scripts/
```
Pass criteria: the listing shows a fresh timestamp/size. Optional belt-and-braces: pull
it back (`b2 file download "b2://personae-model-training/finetuning/scripts/bundle.tar.gz" ./check.tar.gz`)
and confirm `run.sh` contains the `--upload-manifest` block and `02_train` has the
`--upload-manifest` flag.

---

## Step 4 — Tier-2 short launch (B+C integration test, and Tier-2 from the plan)

**Goal:** on one real box, confirm the manifest lands, a checkpoint uploads mid-run,
`RUN_SH_DONE` prints, and the final adapter object appears. This also closes Phase A's
Tier-2 (callback fires inside the Trainer loop).

### 4a. Make checkpoints appear quickly (test only)
`run.sh` hardcodes `SAVE_STEPS=50`; a short run may finish before step 50 and produce **no**
checkpoint. For the test, **temporarily** lower the cadence:
- Edit `run.sh` → `SAVE_STEPS=10` (or 5), re-pack + re-upload the bundle (step 3).
- **Run length is governed by the export's `rows_exported`, not by a time cap.** The
  standard ~1,958-row exports are ~245 steps/epoch ≈ 7–8h even at `epochs:1` (~110s/step) —
  so the full run is NOT short. But at `SAVE_STEPS=10` the **first checkpoint lands at
  ~step 10 ≈ 18 min**, which validates the new path (manifest delivered, callback firing in
  the real Trainer loop, a `ckpt-*.tar.gz` in B2) mid-run — you don't have to wait for the
  whole run. To also confirm `RUN_SH_DONE` + the final adapter quickly, use a genuinely
  small export (a few hundred rows → ~30–60 min); otherwise let a 1-epoch run finish (the
  final-upload PUT itself was already proven in Tier-1).
- **After the test, restore `SAVE_STEPS=50` and re-pack/re-upload** before any real run.

(Alternative: skip 4a and fold this into the next real `save_steps=50` run — the first
checkpoint then lands ~1.5h in. Slower to confirm, no dedicated spend.)

### 4b. Pre-flight, start log capture, then launch
The launch is a Kafka `orchestrate` message to **`model-trainer`** (it creates the
`training_runs` row and drives the `training-launcher` we wired). It is NOT a direct
call to `training-launcher`. Do the two pre-flight checks, start capturing logs, THEN
produce the message.

**(i) Cost gate + no orphan boxes.** Confirm provisioning is allowed and nothing is
already running (an orphan box burns spend and can trip the cost gate):
```sql
SELECT can_provision, denial_reason FROM thunder_provision_check;   -- expect: t | NULL
```
```
 can_provision | denial_reason
---------------+---------------
 t             |
(1 row)
```
```sql
SELECT id, status, training_run_id FROM thunder_instances
 WHERE status NOT IN ('decommissioned','reaped','failed') ORDER BY created_at DESC;
```
```
 id | status | training_run_id
----+--------+-----------------
(0 rows)
```
If `can_provision` is `f`, read `denial_reason` and clear it (cost cap / orphan box)
before launching. If the second query returns rows, decommission them first.

**(ii) Verify the export has real rows.** `export_id` is the PK of `training_exports.runs`.
`runs.rows_exported` can DISAGREE with the real `training_exports.rows` (an export row
existed with `rows_exported=1957` but 0 actual rows, which failed the run at
`prepare_training_data`), so check `actual_rows` directly and pick a non-zero one:
```sql
SELECT r.id, r.agent_type, r.rows_exported, count(x.*) AS actual_rows, r.created_at
  FROM training_exports.runs r
  LEFT JOIN training_exports.rows x ON x.export_id = r.id
 GROUP BY r.id ORDER BY r.created_at DESC;
```
```
                  id                  |     agent_type      | rows_exported | actual_rows |          created_at
--------------------------------------+---------------------+---------------+-------------+-------------------------------
 a8484922-4e3d-4c11-a2ed-4056b13a67a6 | page-content-writer |          1957 |           0 | 2026-04-23 15:04:29...   <- do NOT use (0 actual rows)
 146a9a12-c953-48eb-bf1f-c1856e5f13b7 | page-content-writer |          1958 |        1958 | 2026-04-23 ...           <- use this
```
The non-zero `actual_rows` id (iter_0 used `146a9a12-…`, 1958 rows) is what replaces
`<VALID_EXPORT_ID>` in the trigger. `actual_rows` governs run length (see 4a).

**(iii) Start capturing logs BEFORE you trigger** (so the first messages aren't missed),
in separate terminals:
```bash
# all chassis workers (where the launcher orchestration runs) — JSON to file + screen
kubectl -n ai-persona-system logs --tail=300 -l app=agent-chassis -f | tee logs-agent-chassis.json
# thunder-adapter (presign replies)
kubectl -n ai-persona-system logs --tail=500 -l app=thunder-adapter -f --max-log-requests 20 | tee logs-thunder-adapter.json
# the spawned launcher specifically
kubectl -n ai-persona-system logs --tail=500 -l agent-type=training-launcher -f --max-log-requests 20 | tee logs-training-launcher.json
```
Or capture chassis + dynamic-agent pods together in one timestamped file:
```bash
kubectl -n ai-persona-system logs -l 'app in (agent-chassis,dynamic-agent)' \
  -f --max-log-requests=20 | tee logs-all_chassis_logs_$(date +%H%M%S).log
```
After the run, grep any of these by `$CORRELATION`. (The focused per-step grep is in 4c.)

**(iv) Launch** — produce the `orchestrate` message with kcat from the `kafka` namespace:
```bash
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ   (write these down — grep the logs by CORRELATION)"

kubectl -n kafka run kcat-launch-$(date +%s) --rm -i --restart=Never \
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
  -H step_name=tier2_training_run \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H from_agent_type=user \
  -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"model-trainer"},"input_data":{"export_id":"<VALID_EXPORT_ID>","hyperparameters":{"epochs":1,"batch":1,"grad_accum":8,"lr":0.0002,"lora_r":16,"lora_alpha":16,"max_seq":4096}}}
JSON
```
- **`export_id`** (the `<VALID_EXPORT_ID>` placeholder above) — choose it via the
  `actual_rows` check in 4b(ii). `actual_rows` governs run length (see 4a).
- For the Tier-2 test, **`epochs:1`** (above) + a small export + `SAVE_STEPS` low (4a)
  keeps the run short while still producing checkpoints. `save_steps` is NOT a
  hyperparameter here — it lives in `run.sh`. (Real runs use `epochs:3`.)
- `model-trainer` creates a **new** `training_run` for this launch — it does not pick up
  the four `pending` rows. Capture the new `training_run_id` from the chassis log (4c) or:
  ```sql
  SELECT id, status, started_at FROM model_lifecycle.training_runs
   ORDER BY started_at DESC NULLS LAST LIMIT 3;   -- the freshly-created run
  ```
  Use that `run_id` for the B2 prefixes in 4e.

### 4c. Watch the launcher orchestration (chassis logs)

> **NOTE (2026-06-09): `presign_checkpoints` IS the single batch step now (confirmed in prod).**
> Migration 110 replaced the K-iteration loop + `flatten_checkpoint_urls` with one
> `dispatch_thunder_prepare_object_urls` step. Confirmed run (corr `07dc14fa`): the whole launcher
> completed in ~26s at state Version 30, the batch presign was ONE await (~2s) returning all 40
> URLs. Expected order below.

```bash
kubectl -n ai-persona-system get pods | grep -i chassis
kubectl -n ai-persona-system logs -f <chassis-pod> | \
  grep -Ei "compute_checkpoint_keys|prepare_object_url|assemble_upload_manifest|write_manifest|MANIFEST_WRITTEN|ssh_exec_launch|LAUNCH_PID"
```
Expect, in order: `presign_dataset` → `presign_scripts` → `compute_checkpoint_keys` (built K keys)
→ `presign_checkpoints` = ONE `prepare_object_urls` batch reply (all K urls, single await) →
`presign_final` → `assemble_upload_manifest` (logs `checkpoint_count`, `resume=false`) →
`write_manifest` ssh echoing `MANIFEST_WRITTEN` → `ssh_exec_launch` (`exit_code:0`, `LAUNCH_PID=…`)
→ `mark_running` → `complete`. No per-iteration `presign_checkpoints_iter_N`, no `flatten`.

### 4d. Confirm on the box (manifest + markers)

**First — has the launch sequence even reached the box?** Being able to SSH in only proves
*provision* succeeded; that happens early, before the presign loop, `write_manifest`, and
`ssh_exec_launch`. Everything the launcher puts on the VM is under **`/workspace`** (the
`/etc/thunder` dir is Thunder's own image, unrelated). Check state on the box:
```bash
ls -la /workspace/ 2>/dev/null || echo "no /workspace yet -> write_manifest NOT reached"
[ -f /workspace/upload_manifest.json ] && echo "MANIFEST PRESENT" \
  && python3 -m json.tool < /workspace/upload_manifest.json | head -40
ls -la /workspace/launch.log /workspace/train.log 2>/dev/null
pgrep -af '02_train_llama_3_3_70b|/workspace/run.sh|[p]ython'; echo "---"
tail -n 20 /workspace/train.log 2>/dev/null
```
Interpretation:
- **`/workspace` missing or empty (no `upload_manifest.json`)** → the launcher has NOT reached
  `write_manifest`; it is still in (or stuck in) the presign loop. Confirm from the launcher
  side (below) which iteration it's on and whether it's slowing.
- **manifest present, no `train.log`** → `write_manifest` done; `ssh_exec_launch` just fired or
  is next.
- **`train.log` with `RUN_SH_START`/`RUN_SH_STEP` + a live `python 02_train…`** → full chain
  completed, training running.

Launcher-side, check loop progress and per-iteration time (the `MANIFEST_WRITTEN` token is a
launcher log line, not a VM file):
```bash
kubectl -n ai-persona-system logs -l agent-type=training-launcher --tail=-1 \
  | grep -E "presign_checkpoints_iter_[0-9]+_presign_one|MANIFEST_WRITTEN|ssh_exec_launch|No activity"
# or from the capture: grep -E "iter_[0-9]+_presign_one|MANIFEST_WRITTEN" logs-training-launcher.json
```
**Watch the gap between successive `iter_N` lines.** If it grows each iteration (seconds early,
tens of seconds by iter_7–8, then minutes), the loop is hitting the state-bloat slowdown — every
step re-persists the full expanded 80-step workflow + a growing `collected_data`/`ProcessingHistory`
(O(n²)). At 40 iterations it will not finish in reasonable time and the box bills the whole while.
That is a loop-cost problem, **not** the await race (claims still succeed). Fix is structural:
collapse the loop into one batch `prepare_object_urls` adapter call (see PLAN — the documented
fallback). Kill the launcher job and decommission the box before re-running.

**Once the chain completes** — get the box IP/port and tail the run log:
```sql
-- in clients_db: the box for this run
SELECT thunder_instance_id, status, ip_address, ssh_port
  FROM thunder_instances ORDER BY created_at DESC LIMIT 3;
```
```bash
ssh -p <port> -i <your_thunder_key> ubuntu@<ip>
cat /workspace/upload_manifest.json | python3 -m json.tool   # checkpoints[] + final, NO resume
grep -E "RUN_SH_(START|STEP|SMOKE_OK|UPLOAD|FULL_OK|DONE|FATAL)" /workspace/train.log
```
Expect `RUN_SH_UPLOAD manifest=present save_steps=<N>` after the smoke pass, then later
`RUN_SH_FULL_OK` and `RUN_SH_DONE`. (`RUN_SH_DONE` only prints if the final upload
succeeded — that is the whole point of the `set -e` + hard-gate.)

### 4e. Confirm in B2 (the actual durability check)
```bash
# checkpoints appear DURING the run (re-run as it progresses) — b2 v4:
b2 ls --long "b2://personae-model-training/finetuning/checkpoints/<run_id>/"
# the final adapter appears at the end (after RUN_SH_DONE):
b2 ls --long "b2://personae-model-training/finetuning/artefacts/<run_id>/"
# b2 v3: b2 ls --long personae-model-training finetuning/checkpoints/<run_id>/   (and …/artefacts/<run_id>/)
```
**Pass criteria (all four):** manifest on the box with K+1 URLs; ≥1 `ckpt-<N>.tar.gz`
under `checkpoints/<run_id>/` mid-run; `RUN_SH_DONE` in `train.log`; `adapter.tar.gz`
under `artefacts/<run_id>/`. If `RUN_SH_DONE` printed but no `adapter.tar.gz` exists, the
hard-gate assumption is broken — investigate before step 6.

**If a checkpoint upload fails mid-run:** training continues (checkpoints are best-effort)
but the loop's `continue_on_error: false` is about *presigning*, not the in-loop PUT — a
PUT failure shows in `train.log`, not the chassis. The final upload is the gate.

---

## Step 5 — Resume end-to-end  ✅ WIRED (deploy + test): D3 built, migration `111` applied

**Phase D launcher wiring is now built and applied** (was the blocker). What landed:
- **D3:** chassis dispatch `dispatch_thunder_prepare_resume_url` (`thunder_prepare_resume_url_dispatch.go`,
  clone of the `prepare_object_url` dispatch — sends `prepare_resume_url` + `training_run_id`, reuses
  `ownResponsesTopic` + `preRegisterAwaitedRequest`) + its registry entry. Build into the chassis and
  bump the launcher def `image_tag` (same deploy as the batch route).
- **Migration `111`** (NOT 110 — 110 was the batch presign): inserts `check_resume`
  (`presign_final → check_resume → assemble_manifest`) and adds `resume_url`/`resume_key`/`resume_index`
  to `assemble_manifest.config` pointing at `resume_probe.*`. `check_resume` config is `{expiry_minutes:3000}`
  only — `training_run_id` resolves from `input_data` via `configOrInput` fallback, NOT a config dot-path.
  Apply 111 ONLY after the chassis carrying D3 is deployed (else `check_resume` hits unknown-action).
  (`assemble_upload_manifest` already emits `resume` only when `resume_url` is non-empty, so a
  `found=false` reply produces no resume block.)

The test (after the chassis with D3 is deployed and 111 applied):
1. Launch a run (save cadence low as in 4a so a checkpoint lands fast); wait for ≥1
   `ckpt-<N>.tar.gz` in `checkpoints/<run_id>/`.
2. Kill it mid-run (decommission the box, or kill the python on the box).
3. Relaunch for the **same `training_run_id`**.
4. Verify the chassis log shows `prepare_resume_url` returning `found=true` and the new
   `upload_manifest.json` on the box has a `resume` block pointing at the highest
   `ckpt-<N>`; `train.log` shows the resume download/extract and
   `resume_from_checkpoint=True` (training starts past step 0).
   - Transient B2 list hiccup ⇒ `check_resume` returns `error_recoverable` and the chassis
     retries the step (it should not fail the launch).
   - A truly fresh `run_id` (empty prefix) ⇒ `found=false`, no `resume` block, trains from
     scratch — confirm that path too.

---

## Step 6 — Enable the monitor schedule

**Goal:** let `thunder-training-monitor` decommission finished boxes automatically. The
PLAN's *build* gates are now met (Phase C makes `RUN_SH_DONE ⟹ adapter durable`; Phase D
resume is wired — Step 5). The one remaining *empirical* gate is step 4e proving
`RUN_SH_DONE ⟹ adapter.tar.gz in finetuning/artefacts/<run_id>/` on a real completing run —
that is what makes `DONE_OK → decommission` safe (decommission destroys the VM disk).
Safe-by-design even if the final upload is buggy: the hard gate means no `RUN_SH_DONE`
marker is written on a failed final PUT, so the probe sees process-gone-no-marker →
`GONE_UNKNOWN → mark_failed → decommission` (never a false `DONE_OK`). The terminal/
decommission branch has still **never fired live** (only the ALIVE path was verified
2026-06-05), so watch the first finishing box closely.

```sql
-- in the DB holding scheduled_tasks (clients_db unless your scheduler is split out):
SELECT name, enabled, schedule, last_completed_at
  FROM scheduled_tasks WHERE name='thunder-training-monitor';

UPDATE scheduled_tasks SET enabled=true WHERE name='thunder-training-monitor';
```

**Watch the first finishing box closely.** The monitor's terminal/decommission branch
(`mark_complete/mark_failed → decommission`) has **never run live**. On the next box that
reaches `RUN_SH_DONE`, confirm the monitor: probes the box, sees `DONE`, marks the
`training_runs` row complete, and decommissions the box — and that the `adapter.tar.gz`
was already in B2 *before* the decommission (it must be, given the gate, but verify once).
```bash
kubectl -n ai-persona-system logs -f <monitor-pod> | \
  grep -Ei "probe|DONE|decommission|mark_complete|mark_failed"
```
If anything looks wrong on that first cycle, disable again immediately:
```sql
UPDATE scheduled_tasks SET enabled=false WHERE name='thunder-training-monitor';
```

### 6a. Manual on-demand monitor triggers (test without the schedule)
You can exercise the monitor by hand (this is how it was verified live) instead of waiting
for the schedule — useful to watch the decommission branch on the first finishing box with
the schedule still disabled.

**Orchestrator sweep** (discovers active instances, spawns a worker per box):
```bash
CORRELATION=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid); REQ=$(cat /proc/sys/kernel/random/uuid)
kubectl -n kafka run kcat-monsweep-$(date +%s) --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -k "$CORRELATION" \
  -H correlation_id=$CORRELATION -H orchestration_id=$ORCH -H request_id=$REQ \
  -H message_type=request -H action=orchestrate -H client_id=demo_client \
  -H step_name=monitor_sweep -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H from_agent_type=user -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"thunder-training-monitor"},"input_data":{}}
JSON
```

**Worker probe of one box directly** (skip discovery — probe a known box). Get the ids
first:
```sql
SELECT id AS provisioning_id, training_run_id
  FROM thunder_instances WHERE status='running' AND training_run_id IS NOT NULL;
```
```bash
CORRELATION=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid); REQ=$(cat /proc/sys/kernel/random/uuid)
kubectl -n kafka run kcat-monprobe-$(date +%s) --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -k "$CORRELATION" \
  -H correlation_id=$CORRELATION -H orchestration_id=$ORCH -H request_id=$REQ \
  -H message_type=request -H action=orchestrate -H client_id=demo_client \
  -H step_name=monitor_probe -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H from_agent_type=user -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"thunder-training-monitor-worker"},"input_data":{"provisioning_id":"<PROVISIONING_ID>","training_run_id":"<TRAINING_RUN_ID>"}}
JSON
```
On a box that has reached `RUN_SH_DONE`, the worker should mark the run complete and trigger
decommission — confirm `adapter.tar.gz` is in B2 *before* the box is destroyed.

---

## One-line summary of the gates
- `109` (chain), `110` (batch presign — replaces the loop+flatten), and `111` (resume
  `check_resume`) all applied; apply each only after the chassis carrying its actions is deployed.
- Bundle re-uploaded with **both** edited scripts at the root.
- Step 4 upload path proven: `ckpt-0.tar.gz` confirmed in B2 (run `0ac806ab`). STILL TO SEE on a
  real completing run: the final `adapter.tar.gz` + `RUN_SH_DONE` before trusting auto-decommission.
- Step 5 (resume): D3 dispatch + `111` are built/applied — deploy the chassis with D3, then run the
  kill-and-relaunch test.
- Monitor enabled **last**; its terminal/decommission branch is built but has never fired live.
