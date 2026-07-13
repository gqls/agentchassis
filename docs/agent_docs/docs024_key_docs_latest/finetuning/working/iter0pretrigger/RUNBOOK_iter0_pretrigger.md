# Pre-trigger runbook — iter_0 automated training launch

Goal: get the first automated `model-trainer → … → training-launcher` run to
start training on a Thunder VM. Two code changes must ship before triggering
(neither is in the currently-deployed build); the migration is already applied.

Namespace: `ai-persona-system`. Kafka cluster: `personae-kafka-cluster` (`-n kafka`).

---

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

## 4. Adapter sanity: prepare_object_url round-trip

Before the full run, confirm the new action answers and the reply key is
`presigned_url`. Either tail the adapter during step 6's first presign, or do a
one-off: publish a `prepare_object_url` request (key
`finetuning/scripts/bundle.tar.gz`, method `GET`) to
`system.adapter.thunder.requests` with a `reply_to_topic` you can read, and
confirm the reply body has `presigned_url` (200-able). This also re-checks
`prepare_dataset_url` indirectly (shared `ObjectURL` path).

Also confirm the scripts bundle is actually in B2 at
`personae-model-training / finetuning/scripts/bundle.tar.gz` (the
`UPLOAD_bundle.sh` target). A missing bundle 404s at the VM curl.

## 5. Cost gate

The handoff flagged `thunder_config.daily_cap_usd=15` as too low for a real run
— the iter_0 manifest shows ~9.2h of 80GB-GPU time (`train_runtime_s≈33136`),
which is well above $15. Inspect the gate values and raise the cap as needed:

```
SELECT * FROM <thunder_config table>;   -- confirm column names + current values
```

Check the gate's comparison direction in the provision/cost-gate code before
changing a value, so the change actually permits the run rather than blocking it.

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
