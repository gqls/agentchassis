# PLAN — Checkpoint & final-adapter upload to B2 (presigned per-object PUTs)

Status: Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 isolation test
(box-free B2 round-trip) PASSED 2026-06-05; Tier-2 (in-loop callback + resume on a
GPU box) folds into the first real launch. Phase D **adapter side BUILT** (resume via
the existing `storage.Client.ListObjects` — see Resume below); its launcher wiring
(`dispatch_thunder_prepare_resume_url` + migration `110`) is the only code left.
Deploy + test sequence: **RUNBOOK_phase_b_c_d_deploy.md**. Owner: Phase 5 (flywheel C).
Design agreed 2026-06-04.

## Why this exists

Two coupled gaps, same root cause — the Thunder VM disk is ephemeral (it dies on
reap, crash, or decommission) and right now nothing moves training output off it:

1. **Final adapter is not durable.** `run.sh` writes the LoRA adapter only to
   `/workspace/adapter_out` and emits `RUN_SH_DONE` straight after the local save.
   Its header is explicit that "a later artefact-collector presigns a PUT for this
   dir's tarball" — that collector does not exist. So `RUN_SH_DONE` currently means
   "trained and saved locally", not "trained and durable".
2. **The monitor decommissions on completion.** `thunder-training-monitor-worker`'s
   `DONE_OK` path is `mark_complete → decommission`, and decommission tears down the
   VM disk. Enabling the monitor while (1) holds would destroy the trained adapter.
3. **No checkpoints.** `02_train` runs `save_strategy="no"`, so a crash/reap mid-run
   loses the entire run (the iter_0 run is ~24h; an interruption at hour 20 = total
   loss). There is no resume.

The same mechanism — "VM → B2 upload" — solves all three: checkpoints during the
run, the final adapter at the end, and (via resume) crash recovery.

## Chosen approach and why

**Presigned, single-object, write-only PUT URLs, minted by the adapter, pre-minted
at launch and handed to the VM in a manifest.** The adapter already holds the only
B2 credentials and already presigns URLs (`dispatch_thunder_prepare_object_url`,
returns `{presigned_url, key, expires_at, method}`); the launcher already pushes
files onto the VM by SSH before backgrounding `run.sh`. So this is extension, not
new machinery.

Threat model: **assume the Thunder VM is hostile.** Two alternatives were considered
and rejected as primary:
- *Standing scoped B2 key on the VM* — a bearer credential with prefix-wide write,
  living on the hostile box for the whole run; a leak means write/overwrite across
  the prefix.
- *Per-save callback to an in-cluster endpoint* — gives tighter per-URL expiry, but
  requires exposing an endpoint reachable from the Thunder network (added attack
  surface) and putting a per-run token on the box that can mint URLs across the
  prefix (worse leak).

Pre-minting keeps the box holding nothing but a fixed set of single-object,
write-only URLs (plus one GET URL on a resume). No B2 key, no DB access, no inbound
endpoint. The only cost is that the URLs are long-lived (valid for the run length),
which single-object + write-only + write-once keys bounds acceptably.

**Critical framing:** this protects *access*, not *artefact integrity*. The adapter
is produced on an untrusted box, so a maliciously-crafted-but-valid adapter could be
written through a legitimate URL. That is not a credential problem — it is what the
flywheel-D **eval gate** is for (evaluate before promotion). No upload scheme
substitutes for evaluating the adapter.

## Design

### The manifest is the channel
The launcher writes `/workspace/upload_manifest.json` onto the VM by SSH (exactly as
it places the dataset today) and `run.sh` passes its path to the full-train call.

```json
{ "run_id": "<training_run_id>",
  "checkpoints": [
    {"index": 1, "key": "finetuning/checkpoints/<run_id>/ckpt-1.tar.gz", "url": "<presigned PUT>"},
    ... up to K ],
  "final":  {"key": "finetuning/artefacts/<run_id>/adapter.tar.gz", "url": "<presigned PUT>"},
  "resume": {"key": ".../ckpt-7.tar.gz", "url": "<presigned GET>", "index": 7}   // present only on a resume launch
}
```

### Key by save-INDEX, not the Trainer's global_step
HF Trainer names checkpoints `checkpoint-{global_step}`, which is fragile to predict
exactly. Instead the manifest keys checkpoints by **save index** (1st save, 2nd
save, …). The on-save callback uploads its Nth save to `checkpoints[N].url`
regardless of the step the Trainer landed on. The launcher mints
`K = ceil(max_steps / save_steps) + buffer` URLs (iter_0 at save_steps=50 →
~15 + buffer — trivial to pre-mint).

### Write-once
Each index key is distinct and written once. Turn on **B2 object versioning** as a
backstop so a replayed/leaked PUT cannot overwrite a checkpoint in place. Residual
(valid poison to a fresh key) is covered by the eval gate, not here.

## Component changes

### Launcher (Go action + workflow)
- `dispatch_thunder_prepare_object_url` ALREADY exists and ALREADY supports an explicit
  key + `method=PUT` (adapter `ObjectURL`/`handlePrepareObjectURL`), and the launcher
  ALREADY calls it to presign the dataset (GET) and scripts bundle (GET). So no adapter
  change is needed for the upload path — Phase B just signs more keys through the proven
  path. Confirmed against `data_url_actions.go` 2026-06-05 (the earlier "may be missing"
  note was from an incomplete copy).
- A Go action computes the checkpoint key list: `K = ceil(max_steps/save_steps) + buffer`
  from run_id + epochs + n_examples + batch×grad_accum + save_steps (all known at launch),
  producing keys `finetuning/checkpoints/<run_id>/ckpt-<index>.tar.gz`.
- Presign K checkpoint PUTs + 1 final PUT. Two options, reuse-first:
  - *Loop (recommended first):* loop the existing single-key `prepare_object_url` over
    the key list within the launcher def. Presigning is local HMAC (no B2 call), so K≈15
    awaited round-trips at launch cost ~seconds, one-time — negligible. Zero adapter
    change. The final PUT can reuse `prepare_artefact_url` (purpose-built, key from run_id)
    or `prepare_object_url` with an explicit key.
  - *Batch (later optimization):* add a plural `prepare_object_urls` (keys[] → urls[])
    mirroring `ObjectURL`, so the workflow is one presign step instead of a loop. Only
    worth it if the loop's per-iteration log expansion becomes annoying.
- **Expiry caveat:** the PUT default is `defaultArtefactURLExpiryMin` = 24h, which is
  SHORTER than a run that can hit the 48h cap. The launcher MUST pass `expiry_minutes`
  explicitly = `max_uptime_hours×60 + margin` (e.g. +120) so a late-run checkpoint still
  has a live URL. The dispatch already relays `expiry_minutes`.
- A Go action assembles `upload_manifest.json` from the presign results, then SSH-place it
  onto the VM alongside the dataset/scripts (reuse the existing placement mechanism).
- Complexity stays in Go (key computation, manifest assembly); the workflow stays a thin
  presign → place → launch.

### `02_train_llama_3_3_70b.py` — BUILT 2026-06-05
As-built (all gated; with no new flags the script is byte-for-byte the old behaviour,
so the smoke pass and the way iter_0 ran are unchanged):
- New flags: `--save-steps N` (default 0), `--save-total-limit` (default 2),
  `--upload-manifest <path>` (default "").
- `--save-steps N > 0` flips `save_strategy` from `"no"` to `"steps"` and sets
  `save_steps`/`save_total_limit`. With the default 0 it stays `"no"` (the one behaviour
  change, gated). Checkpoints land in `<output>/checkpoints/checkpoint-<step>`.
- `CheckpointUploader(TrainerCallback).on_save`: tars the just-written checkpoint dir
  and PUTs it to `manifest.checkpoints[save_index].url`, keyed by **save index**
  (0,1,2,…), not global_step. Best-effort — a failed checkpoint upload logs and
  training continues (the next checkpoint re-establishes durability).
  - **Content-Type:** the adapter's `GetPresignedPutURL` passes only key+expiry — no
    Content-Type — so the presigned PUT is *not* signature-bound to a content type, and
    the callback sends `application/octet-stream` (constant `_OCTET`). Verify this in the
    isolation test (a 403 `SignatureDoesNotMatch` would mean the presigner *does* bind a
    content type and we'd match it). This is the single riskiest mechanic to confirm.
  - **Synchronous:** `on_save` runs inline in the training loop, so each checkpoint
    stalls training for the tar+upload (~1–2 min at ~2GB). A background thread was
    rejected — it races `save_total_limit`'s deletion of the dir being uploaded. The
    `w:gz` gzip is the tuning knob: switch `_tar_dir` to mode `w` (no compression) to
    roughly halve the stall if cadence increases.
- Final adapter (the hard gate): after `model.save_pretrained` + the run-metadata
  `manifest.json`, tars `output_path` (excluding `checkpoints/`) and PUTs to
  `manifest.final.url`. Unlike checkpoints this **raises** on failure → non-zero exit
  → run.sh emits no DONE marker → the monitor will not treat the box as cleanly done.
- Resume: if `manifest.resume` is present, GETs + extracts that tarball into
  `<output>/checkpoints/` and calls `trainer.train(resume_from_checkpoint=True)`, so the
  Trainer restores optimizer/scheduler/RNG/step (not just weights). `save_index` resets
  per process — the **launcher** owns key allocation across resume launches (Phase B),
  the script does not.
- NOTE the two distinct manifests: the *input* `--upload-manifest` (presigned URLs, not
  written by the script) and the *output* run-metadata `manifest.json` (written as
  before, L162-187 — this is the 554-byte file seen in `iter0_adapter_out`). Different
  files, no collision.
- Packaging: the edited script (and any helper module it grows) must be in
  `bundle_tar.gz` to reach the VM — the "did it ship" check, analogous to a `go build`.

### `run.sh` — BUILT 2026-06-05
- Pass `--save-steps 50 --upload-manifest` to the **full-train** invocation only (not
  smoke), and only when the manifest file is present (a manifest-less manual run behaves
  exactly as before).
- **No marker was moved.** `set -euo pipefail` plus `02_train`'s final-upload hard-gate
  (raises → non-zero exit) means `RUN_SH_FULL_OK`/`RUN_SH_DONE` are only reached on exit 0,
  which now implies the adapter is in B2. That is what makes the monitor's
  `DONE_OK → decommission` safe; the probe's `RUN_SH_DONE` + `adapter_config.json` check
  then genuinely implies durability. `SAVE_STEPS` is the cadence knob in `run.sh`.

### Resume (cluster-side selection) — adapter side BUILT 2026-06-05
On a fresh box for the same `training_run_id`, the launcher (in-cluster, trusted) asks
the adapter to list the run's checkpoint prefix in B2, pick the highest `ckpt-<N>`,
presign a GET for it; the launcher then adds `resume` to the manifest. The VM only ever
receives URLs.
- **No new storage method was needed.** The earlier note that this was "the ONE genuine
  adapter gap" was wrong: `storage.Client` ALREADY declares
  `ListObjects(ctx, prefix) ([]ObjectInfo, error)` (interface.go) and `*S3Client` ALREADY
  implements it (aws-sdk-go-v2 `ListObjectsV2Paginator`, against the same `c.bucket` the
  existing dataset/artefact presigns use). The blast-radius caution (016 item 18) is about
  *adding* a method to that broad interface — listing is already on it, so there is nothing
  to widen. The narrow `objectLister` interface built earlier was removed in favour of
  calling `ListObjects` directly (reuse-before-create).
- As-built: adapter action `prepare_resume_url` (`DataURLAction.ResumeURL` +
  `handlePrepareResumeURL`; dispatch switch case added in `adapter.go`) →
  `ListObjects(prefix)` → `latestCheckpointKey` (parses `ckpt-<N>.tar.gz`, takes max) →
  `GetPresignedURL`. Reply `{found, presigned_url, key, index, expires_at}`. `found=false`
  (empty prefix — not an error; `ListObjects` returns no objects) ⇒ the launcher proceeds
  fresh, no resume block. A transient **list** failure (B2 network) is returned
  `error_recoverable`, so the chassis retries `check_resume` instead of failing the launch;
  presign (local SigV4 signing — no network) and bad-input errors stay `error_unrecoverable`.
- LEFT (Phase D launcher wiring): a chassis dispatch `dispatch_thunder_prepare_resume_url`
  + a `check_resume` step (migration `110`) before `assemble_manifest`, whose
  `resume_url`/`resume_key`/`resume_index` config paths point at `check_resume.*`.
  `assemble_upload_manifest` already emits `resume` only when `resume_url` is non-empty,
  so `found=false` needs no special handling.

## Net security position
The VM holds only single-object, write-only PUT URLs (plus one GET on a resume), no
B2 key, no DB access, nothing reachable from outside. A compromised box can overwrite
at most its own checkpoint objects within their expiry, bounded further by
versioning; it cannot read, list, write elsewhere, or mint new grants. B2 credentials
never leave the cluster. Integrity is guarded by the eval gate.

## Build order (each phase testable in isolation)
- **A. `02_train` callback + resume — BUILT 2026-06-05.** Isolation test (two tiers,
  no launcher/chassis changes needed):
  - *Tier 1 (box-free, no GPU) — PASSED 2026-06-05* (`isolation_test_phase_a.py`,
    against personae-model-training @ us-east-005): presigned-PUT signature accepts
    `application/octet-stream`, the `checkpoints/` exclusion holds, and a checkpoint
    GET+extract round-trips byte-identical. Presigned directly here (not via the
    chassis) only because it is a box-free test; the adapter's `prepare_object_url`
    (`ObjectURL`/`handlePrepareObjectURL`) DOES exist and is what Phase B reuses.
  - *Tier 2 (one GPU box):* short real run `--limit 32 --epochs 1 --save-steps 5
    --upload-manifest …` to confirm the callback fires inside the Trainer loop and
    `resume_from_checkpoint=True` actually resumes. Do this on the next box provisioned,
    not a dedicated spend.
- **B. Launcher presign + manifest — BUILT 2026-06-05.** Three pure actions
  (`compute_checkpoint_keys`, `flatten_presign_results`, `assemble_upload_manifest`), a
  `key_path` source added to `dispatch_thunder_prepare_object_url`, and migration `109`
  wiring the chain `compute_keys → presign_checkpoints (loop) → flatten_checkpoint_urls →
  presign_final → assemble_manifest → write_manifest → ssh_exec_launch`. The loop has an
  explicit `loop_complete` terminal (`presign_done`) and the async presign substep chains
  into it — matches every production loop. Test: launch a run; confirm the manifest lands
  on the VM with K+1 URLs and a checkpoint uploads during a real run.
- **C. `run.sh` `DONE`-after-upload — BUILT 2026-06-05.** No marker moved: `set -e` +
  `02_train`'s final-upload hard-gate means `RUN_SH_DONE` only prints on exit 0 = trained
  AND uploaded. Full-train gets `--save-steps 50 --upload-manifest …` only when the
  manifest is present. This is the flip that turns the monitor's decommission safe.
- **D. Resume end-to-end — adapter side BUILT 2026-06-05; launcher wiring LEFT.** Adapter
  `prepare_resume_url` built (see Resume above; reuses `ListObjects`). Remaining:
  `dispatch_thunder_prepare_resume_url` + migration `110` (`check_resume` step). Test: kill
  a run mid-way; provision a fresh box for the same `run_id`; confirm the launcher finds
  the latest checkpoint, presigns a GET, and the VM resumes. **Gated on B+C being proven
  first** — resume reads the checkpoints the upload path writes.
- **Then:** enable `thunder-training-monitor` (now safe — `DONE` means durable).

## Parameters / open decisions
- `save_steps` cadence (recommendation **50–65**, ≈ one checkpoint every ~1.5–2h). This
  is the number of optimizer steps *between* checkpoints, not the checkpoint count. Over
  iter_0's ~726 steps at ~110 s/it: 50 → ~15 checkpoints (~1.5h apart); ~65 → ~11
  checkpoints (~2h apart). Fewer, larger gaps = fewer synchronous save-stalls (the
  preferred trade here — ~2h of rework on a crash is acceptable). Lives in the
  launcher/run.sh (Phase B/C), **not** hardcoded in `02_train` (it is the `--save-steps`
  arg, default 0).
- Checkpoint size ≈ 2GB: adapter weights (~414MB bf16) + AdamW optimizer state
  (~1.6GB). The 70B base is frozen 4-bit and NOT saved.
- URL expiry: cap + margin (≈50h for an 18–48h cap).
- `K` buffer: `ceil(max_steps/save_steps)` + ~25%.
- B2 bucket versioning: ON.

## Interim, for iter_0 — DONE 2026-06-05
`fabfd7fa` reached its final save; `/workspace/adapter_out` was scp'd to
`~/projects/agentchassis/iter0_adapter_out` (adapter_model.safetensors 828MB +
tokenizer + run-metadata manifest.json), training_run `1cd65dd7` reconciled to
`complete`, and the box decommissioned + confirmed gone. The schedule was never
enabled on this run.

## Dependency / not part of this plan
The monitor was verified end-to-end 2026-06-05 (worker-direct ALIVE path and the
orchestrator→spawn→call path; the earlier "no chassis log" was a reply-topic orphan in
`thunder_ssh_get_status_dispatch.go`, since fixed to use `execCtx.ResponsesTopic`). Its
terminal/decommission branch (`mark_complete/mark_failed → decommission`) has still
**never run live** — it fires on the next box that finishes. Enabling the schedule
remains gated on this plan reaching Phase C (so `RUN_SH_DONE` means durable) plus a
Phase D resume check.
