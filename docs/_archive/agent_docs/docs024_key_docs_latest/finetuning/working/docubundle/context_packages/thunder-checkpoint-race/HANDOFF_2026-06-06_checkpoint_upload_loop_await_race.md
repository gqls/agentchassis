# HANDOFF — 2026-06-06 — Flywheel-C checkpoint/adapter upload: loop-await race

## Where this sits
Flywheel C fine-tunes Llama 3.3 70B on Thunder A100s; checkpoints + the final adapter upload
to Backblaze B2 via **presigned PUT URLs** carried in an upload manifest. The launcher
(`training-launcher` def, type `training-launcher`, id `1223bdc1`) builds the manifest, writes
it onto the VM, then SSH-launches training. The launcher is spawned + called by the
`model-trainer` orchestrator (id `94f5a069`): spawn data-preparer/provisioner/launcher → call
each in order; `call_launcher` runs the 109 chain in the spawned launcher pod.

- Chassis image: `docker.io/aqls/agent-chassis:v1.0.1057` (has the Phase B actions).
- DB `clients_db`: `agent_definitions`, `model_lifecycle.*`, `training_exports.*`.
- k8s ns `ai-persona-system`; Kafka `-n kafka`. B2 bucket `personae-model-training`,
  endpoint `https://s3.us-east-005.backblazeb2.com`. Use the **b2** CLI (not aws).

## Done + verified this session
- Phase A (02_train upload/resume), Phase B (launcher manifest actions), Phase C (run.sh),
  Phase D adapter side (`prepare_resume_url`) — built. See PLAN/NOTES.
- **109** (launcher manifest wiring) — applied + verified live.
- **109a** (write_manifest creates `/workspace` with `sudo mkdir` + `sudo chown`) — applied +
  verified live. It's the launcher's first VM-FS touch; the non-sudo `mkdir` failed because
  `/` isn't ssh-user-writable.
- **109b** (presign_one reads the loop item via `key_path:"ckpt_key"`, NOT `input_mapping`) —
  applied + verified live, and **CONFIRMED WORKING IN PRODUCTION**: the adapter now presigns
  the correct per-iteration key (logged `…/ckpt-9.tar.gz`, method PUT). The dataset-key bug is
  closed. (Rule learned: a loop substep that is a *local* action reads the item via a config
  dot-path the action resolves; `input_mapping` is the `call_agent` caller mechanism and is
  dead for local loop substeps.)
- Export verification: **use `146a9a12` or `fef7be6b`** (1958 actual rows each). **Do NOT use
  `a8484922`** — 0 actual rows despite `rows_exported=1957`. Verify with the LEFT JOIN count in
  RUNBOOK §4b before any launch.

## CURRENT BLOCKER (OPEN): loop-await race
After 109b the `presign_checkpoints` loop presigns correct keys but **stalls intermittently at
a later iteration** (iter_6 in run `cdba2808`, iter_9 in run `a2a41ae2` — moving point ⇒ race).
The adapter replies (~1s, even twice for the stuck iteration), but the launcher never clears
the await; the timeout handler re-dispatches every ~3 min with a fresh request_id, so
RetryVersion stays 0 and it never hits the max-retries fail path → effectively infinite retry.

**Root cause (verified against `production_agent-chassis-actions-current_context.txt`):**
our loop substep is a LOCAL dispatch (`dispatch_thunder_prepare_object_url`). It PRODUCES the
adapter request, returns `await_response:true`, THEN the coordinator registers the awaited
request (`processAwaitResponse` L1677 → persist state → `InsertAwaitedRequest`).
**Send-BEFORE-register.** For a ~1s reply the response can arrive before the `awaited_requests`
row is inserted → `ClaimAwaitedRequest` (L396, `WHERE status='waiting'`) finds nothing → the
reply is dropped → timeout → retry → same race. `presign_dataset` (first dispatch) wins the
race; later loop iterations lose it more often. CONTRAST: `spawn_agent`/`call_agent` call
`preRegisterAwaitedRequest` (L57855) to register BEFORE sending — which is why the working
production loops (vet-batch-processor `process_batch`, content-feed `process_sites`, both
**call_agent** substeps) don't stall and ours (local dispatch) does.

## FIX TO TRY (non-framework) — NOT yet implemented
Make our dispatch pre-register the awaited request BEFORE producing the adapter request,
reusing the EXISTING `preRegisterAwaitedRequest` (the one `spawn_agent` uses). This closes the
window without touching the coordinator/loop machinery — it's our action + an existing helper.

Signature (from the spawn call site, L25020):
```
preRegisterAwaitedRequest(ctx, params, requestID, targetAgentID, targetAgentType,
                          requestsTopic, responsesTopic) error
```
For our dispatch: `requestID` = the id the dispatch generates; `targetAgentType` =
`"thunder-adapter"`; `targetAgentID` = "" (or the adapter id if known); `requestsTopic` =
`"system.adapter.thunder.requests"`; `responsesTopic` = the await-responses topic the dispatch
already computes (where the adapter replies).

Steps:
1. **Read `preRegisterAwaitedRequest` body (L57855)** and CONFIRM it sets
   `step_name`/`step_id`/`timeout_at`/`status='waiting'` from `params.ExecutionContext` /
   `params` (so `ClaimAwaitedRequest` + `handleCompleteResponse` resume off the pre-registered
   row). If it derives step from `params`, the dispatch call is enough; if it needs explicit
   step fields, pass/extend accordingly. (This read is the one step I had not finished.)
2. In `thunder_prepare_object_url_dispatch.go`, just before `ProduceWithValidation` (the send),
   call `preRegisterAwaitedRequest(...)` guarded by `if params.DB != nil` (as spawn does). Keep
   returning `await_response:true` — the coordinator's `processAwaitResponse` will then call
   `InsertAwaitedRequest`, hit "already exists", and no-op (handled at L1780).
3. This affects ALL presign dispatches (dataset/scripts/loop/final/resume) — they all become
   register-before-send. The working ones stay working (more robust); the loop stops racing.
4. Rebuild + redeploy the chassis (the dispatch compiles into the chassis image; bump from
   v1.0.1057). The `training-launcher` runs the chassis image.
5. Re-run (export `146a9a12` or `fef7be6b`). CONFIRM in the launcher pod: each
   `presign_checkpoints_iter_N_presign_one` should now log
   `ClaimAwaitedRequest: status_before=waiting … claimed:true`, the loop should run all 40
   iterations → `flatten_checkpoint_urls` → `MANIFEST_WRITTEN` → `ssh_exec_launch`.

If after this it STILL stalls, capture from the launcher pod at a stuck iteration:
`kubectl -n ai-persona-system logs <launcher-pod> | grep -Ei "ProcessResponse|ClaimAwaitedRequest|status_before|no matching|InsertAwaitedRequest"`
— that tells you whether the row now exists when the reply lands (race closed but another
issue) or the reply isn't being consumed at all (different problem).

## FALLBACK (structural) if the pre-register fix is fragile
Replace the 40-iteration loop with ONE batch adapter call: add a `prepare_object_urls`
(plural) handler that takes the key array (compute_checkpoint_keys already emits all 40) and
returns `[{key,url}]` (all local SigV4 presigns — no per-key network). One async round-trip
like `presign_dataset` → no loop, no `flatten`, no race class. Bigger change (adapter handler +
a dispatch + a migration replacing `presign_checkpoints`+`flatten` with one step + adjust
`assemble_manifest` inputs) but removes the failure class. Recommended if the race fix proves
flaky. Verify the adapter's current `prepare_object_url` handler + presign helpers
(`GetPresignedPutURL`/`PresignPutObject`) first and reuse them.

## CLEANUP NEEDED NOW
- Kill stuck launcher jobs (also stops the adapter re-presigning in a loop):
  ```bash
  kubectl -n ai-persona-system delete job agent-training-launcher-24eb2c59 agent-training-launcher-48c12fe3
  ```
- Confirm no live boxes (they were decommissioned at the 18h cap; verify):
  ```sql
  SELECT id, status, training_run_id FROM thunder_instances
   WHERE status NOT IN ('decommissioned','reaped','failed') ORDER BY created_at DESC;
  ```
  decommission any that remain (`tnr` / your usual path).
- Optional: mark the stalled runs failed to keep the table clean
  (`70156952-180b-41bf-ac54-e0288238a288`, plus the `a2a41ae2` and `fef7be6b` runs).

## Also pending (unchanged, lower priority)
- **model-trainer fall-through** after a failed `call_agent`: a failed `call_data_preparer`
  did not abort — it ran `call_provisioner`, which then failed on
  `preparation_result.training_run_id not found`. Separate model-trainer fix, non-blocking.
- **SAVE_STEPS**: the bundle in B2 is the `=50` version (re-uploads were byte-identical — same
  md5; editing `run.sh` without re-tarring doesn't change the bundle). For a fast checkpoint
  test re-pack with `SAVE_STEPS=10`:
  `tar -czf bundle.tar.gz 00_vm_setup.sh 02_train_llama_3_3_70b.py run.sh && b2 file upload personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz`
  (restore 50 after). At `=50` the first checkpoint lands ~1.5h in, not ~18 min.
- **Phase D launcher wiring** (D3 `dispatch_thunder_prepare_resume_url` + D4 migration `110`) —
  only after the upload path is green.
- **Enable `thunder-training-monitor` LAST** (the terminal/decommission branch has never run
  live).

## Key IDs / files
- defs (clients_db): model-trainer `94f5a069`, data-preparer `71ab9361`, training-launcher
  `1223bdc1`, monitor-orchestrator `c3b4c052`, monitor-worker `470c6b3f`.
- exports (`training_exports.runs`): `146a9a12` (1958 ✓), `fef7be6b` (1958 ✓), `a8484922`
  (0 ✗ — do not use).
- outputs: `109` / `109a` / `109b` .sql; **`thunder_prepare_object_url_dispatch.go`** (where the
  fix goes); the 3 Phase B actions (`compute_checkpoint_keys_action.go`,
  `assemble_upload_manifest_action.go`, `flatten_presign_results_action.go`); `run.sh`;
  `02_train_llama_3_3_70b.py`; `data_url_actions.go`; `adapter.go`;
  `RUNBOOK_phase_b_c_d_deploy.md`; `NOTES_phase5_training_launcher_running.md`;
  `PLAN_checkpoint_and_artefact_upload_b2.md`.
- diagnosis source (`production_agent-chassis-actions-current_context.txt`):
  `processAwaitResponse` L1677, `createAwaitedRequest` L1971, `ClaimAwaitedRequest` L396,
  `handleRequestTimeout` L2969, `preRegisterAwaitedRequest` L57855, `LoopCompleteAction`
  L39489; loop-expansion + `setLoopVariable` in `loop_expansion_handler.go`.

## Verify queries
- presign_one binding (expect `key_path:"ckpt_key"`, no `input_mapping`):
  ```sql
  SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints,config,sub_workflow,steps,presign_one,config}')
    FROM agent_definitions WHERE type='training-launcher'
     AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
  ```
- export rows (use a non-zero `actual_rows`):
  ```sql
  SELECT r.id, r.rows_exported, count(x.*) AS actual_rows
    FROM training_exports.runs r
    LEFT JOIN training_exports.rows x ON x.export_id = r.id
   GROUP BY r.id ORDER BY r.created_at DESC;
  ```
- the freshly-created run for a launch (orders by created_at so a `pending` row shows):
  ```sql
  SELECT id, status, created_at, started_at, thunder_instance_id
    FROM model_lifecycle.training_runs WHERE export_id='<EXPORT_ID>' ORDER BY created_at DESC;
  ```
