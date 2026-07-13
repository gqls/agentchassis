# PLAN — Checkpoint & final-adapter upload to B2 (presigned per-object PUTs)

Status: DESIGN AGREED, not built. Owner: Phase 5 (flywheel C). Date: 2026-06-04.

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
  "final":  {"key": "finetuning/adapters/<run_id>/adapter.tar.gz", "url": "<presigned PUT>"},
  "resume": {"key": ".../ckpt-7.tar.gz", "url": "<presigned GET>", "index": 7}   // present only on a resume launch
}
```

### Key by save-INDEX, not the Trainer's global_step
HF Trainer names checkpoints `checkpoint-{global_step}`, which is fragile to predict
exactly. Instead the manifest keys checkpoints by **save index** (1st save, 2nd
save, …). The on-save callback uploads its Nth save to `checkpoints[N].url`
regardless of the step the Trainer landed on. The launcher mints
`K = ceil(max_steps / save_steps) + buffer` URLs (iter_0: save_steps=50 →
~14 + buffer, K < 20 — trivial to pre-mint).

### Write-once
Each index key is distinct and written once. Turn on **B2 object versioning** as a
backstop so a replayed/leaked PUT cannot overwrite a checkpoint in place. Residual
(valid poison to a fresh key) is covered by the eval gate, not here.

## Component changes

### Launcher (Go action + workflow)
- A Go action that builds the upload manifest: loop `key → presigned PUT` reusing
  `dispatch_thunder_prepare_object_url` with `method=PUT`. **Extend** that action to
  PUT if it is GET-only, and ideally to accept a *list* of keys → list of URLs so it
  is one adapter round-trip instead of K. Complexity stays in Go; the workflow stays
  a thin presign → place → launch.
- `expires_at` = instance cap (`max_uptime_hours`) + margin (e.g. +2h), so a
  late-run checkpoint still has a live URL.
- SSH-place the manifest onto the VM alongside the dataset/scripts.

### `02_train_llama_3_3_70b.py`
- `save_strategy="no"` → `"steps"`, set `save_steps`, `save_total_limit=2` (local
  disk keeps only the last couple).
- New `TrainerCallback.on_save`: tar the checkpoint dir the Trainer just wrote and
  PUT the tarball to `manifest.checkpoints[N].url`.
  - **Gotcha:** presigned PUTs are signed for a specific `Content-Type`. Sign with a
    fixed `application/octet-stream` and have the VM send that exact header, or the
    PUT 403s.
  - A failed upload logs loudly but does **not** crash training (the next checkpoint
    re-establishes durability).
- Final save already exists (`model.save_pretrained(output_path)`, L159). After it,
  tar `output_path` and PUT to `manifest.final.url`; exit non-zero if that PUT fails.
- Resume: if `manifest.resume` is present, download + untar it and pass
  `resume_from_checkpoint=<path>` so optimizer/scheduler/RNG state is restored (not
  just weights).
- All upload/resume logic gated on the manifest being present, so the smoke pass
  (`--limit 20`, no manifest) uploads nothing.

### `run.sh`
- Pass `--upload-manifest` to the **full-train** invocation only (not smoke).
- Move `RUN_SH_DONE` so it is emitted only after `02_train` exits 0 — which now means
  **trained AND uploaded**. This single semantic change is what makes the monitor's
  `DONE_OK → decommission` safe; the probe's existing `RUN_SH_DONE` +
  `adapter_config.json` check then genuinely implies the adapter is in B2.

### Resume (cluster-side selection)
On a fresh box for the same `training_run_id`, the launcher (in-cluster, trusted):
lists the run's checkpoint prefix in B2, picks the highest index, presigns a GET for
it, and adds `resume` to the manifest. The VM only ever receives URLs.
- Needs a small **list-keys-under-prefix** capability on the adapter if it does not
  already have one (it holds the B2 creds; bounded addition).

## Net security position
The VM holds only single-object, write-only PUT URLs (plus one GET on a resume), no
B2 key, no DB access, nothing reachable from outside. A compromised box can overwrite
at most its own checkpoint objects within their expiry, bounded further by
versioning; it cannot read, list, write elsewhere, or mint new grants. B2 credentials
never leave the cluster. Integrity is guarded by the eval gate.

## Build order (each phase testable in isolation)
- **A. `02_train` callback + resume.** Hand-build a manifest with a few presigned
  PUT URLs (presign manually via the adapter or b2 CLI from the cluster) against a
  throwaway prefix; run a short training; confirm tarballs land in B2 with the
  correct Content-Type; confirm resume from a downloaded checkpoint works. No
  launcher/chassis changes needed to test this.
- **B. Launcher presign + manifest** (+ adapter list-keys for resume). Launch a run;
  confirm the manifest lands on the VM with K+1 URLs and checkpoints upload during a
  real run.
- **C. `run.sh` `DONE`-after-upload.** Full run → `RUN_SH_DONE` only after the
  adapter tarball is in B2. This is the flip that turns the monitor's decommission
  from unsafe to safe.
- **D. Resume end-to-end.** Kill a run mid-way; provision a fresh box for the same
  `run_id`; confirm the launcher finds the latest checkpoint, presigns a GET, and the
  VM resumes.
- **Then:** enable `thunder-training-monitor` (now safe — `DONE` means durable).

## Parameters / open decisions
- `save_steps` cadence vs upload volume: ~50 steps (~1.75h at 127 s/it) → ~14 ckpts
  × ~2GB ≈ 28GB over the run; `save_total_limit=2` locally. 100 = fewer/larger gaps,
  25 = more frequent/safer, more uploads. Tune to acceptable rework-on-crash.
- Checkpoint size ≈ 2GB: adapter weights (~414MB bf16) + AdamW optimizer state
  (~1.6GB). The 70B base is frozen 4-bit and NOT saved.
- URL expiry: cap + margin (≈50h for an 18–48h cap).
- `K` buffer: `ceil(max_steps/save_steps)` + ~25%.
- B2 bucket versioning: ON.

## Interim, for the CURRENT iter_0 run (before any of this is built)
When `fabfd7fa` hits `RUN_SH_DONE` (~hours away), scp `/workspace/adapter_out` off the
box (runbook §7 SSH recipe) BEFORE anything decommissions it, then it is safe to let
the monitor reconcile/decommission. Do NOT enable the schedule on this run until the
adapter is secured.

## Dependency / not part of this plan
The monitor itself must still be verified (the manual worker test left no chassis log
under `-l app=agent-chassis` — chase pod-label/consumer first). The upload plan and a
verified monitor are both prerequisites to enabling the schedule on a real run.
