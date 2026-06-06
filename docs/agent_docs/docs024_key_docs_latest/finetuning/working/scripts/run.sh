#!/usr/bin/env bash
# FILE: flywheel_C/run.sh
#
# On-VM launch chain for an automated training run. The training-launcher
# agent fetches this (inside the scripts bundle) plus the dataset via presigned
# B2 URLs, then backgrounds THIS script with nohup. All the heavy, long-running
# work lives here — not in the chassis workflow — so the workflow stays a thin
# "fetch + run" and this chain is editable by re-uploading the bundle (no DB
# migration).
#
# Sequence:
#   1. 00_vm_setup.sh        — venv + CUDA torch + Unsloth (idempotent, 5-10 min)
#   2. smoke train           — --limit 20 --epochs 1 to a throwaway dir; gates (3)
#   3. full train            — defaults (3 epochs) to /workspace/adapter_out;
#                              checkpoints + uploads to B2 when a manifest is present
#
# The smoke pass is chained with && so a format/OOM/dep failure stops BEFORE the
# 30-90 min full run is started — cheap insurance on an unattended first run.
#
# Expected inputs (already placed by the launcher before this runs):
#   /workspace/training_iter0.jsonl   — the dataset (curl'd from presigned URL)
#   /workspace/00_vm_setup.sh         — from the bundle
#   /workspace/02_train_llama_3_3_70b.py — from the bundle
#
# Output:
#   /workspace/adapter_out/           — LoRA adapter + manifest.json (full run)
#   When /workspace/upload_manifest.json is present (placed by the launcher's
#   write_manifest step), the full run ALSO checkpoints periodically and uploads
#   each checkpoint + the final adapter to B2 via 02_train's presigned PUT URLs.
#   The final upload is a hard gate inside 02_train (non-zero exit on failure),
#   so with `set -e` a RUN_SH_DONE here implies the adapter is in B2 — which is
#   what makes the monitor's DONE -> decommission safe.
#
# Logging: this script's stdout/stderr is redirected by the launcher to
# /workspace/train.log. Steps are also echoed with markers the monitor can grep.

set -euo pipefail

WORKSPACE="/workspace"
DATA="${WORKSPACE}/training_iter0.jsonl"
SMOKE_OUT="${WORKSPACE}/smoke_out"
FULL_OUT="${WORKSPACE}/adapter_out"
VENV="${HOME}/unsloth_env"
TRAIN_PY="${WORKSPACE}/02_train_llama_3_3_70b.py"
SETUP_SH="${WORKSPACE}/00_vm_setup.sh"
MANIFEST="${WORKSPACE}/upload_manifest.json"

echo "RUN_SH_START ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cd "${WORKSPACE}"

# Fail fast with clear markers if the launcher's pre-placed files are missing.
for f in "${DATA}" "${TRAIN_PY}" "${SETUP_SH}"; do
  if [[ ! -f "${f}" ]]; then
    echo "RUN_SH_FATAL missing_required_file=${f}"
    exit 2
  fi
done

# ── 1. Setup (idempotent) ──────────────────────────────────────────────────
echo "RUN_SH_STEP step=setup"
chmod +x "${SETUP_SH}"
"${SETUP_SH}"

# Activate the venv the setup script created.
if [[ ! -f "${VENV}/bin/activate" ]]; then
  echo "RUN_SH_FATAL venv_missing=${VENV}/bin/activate"
  exit 3
fi
# shellcheck disable=SC1091
source "${VENV}/bin/activate"

# ── 2. Smoke pass (gates the full run) ─────────────────────────────────────
echo "RUN_SH_STEP step=smoke"
python "${TRAIN_PY}" \
  --data "${DATA}" \
  --output "${SMOKE_OUT}" \
  --limit 20 \
  --epochs 1
echo "RUN_SH_SMOKE_OK"

# ── 3. Full train (defaults: 3 epochs) ─────────────────────────────────────
echo "RUN_SH_STEP step=full_train"
# When the launcher placed an upload manifest (Phase B write_manifest step),
# checkpoint every SAVE_STEPS optimiser steps and stream each checkpoint + the
# final adapter to B2 via 02_train's presigned PUT URLs. 02_train's final upload
# is a hard gate (non-zero exit on failure), and with `set -e` that means
# RUN_SH_DONE below is only reached once the adapter is in B2 — so the monitor
# can safely decommission on DONE. Without a manifest (e.g. a manual run) the
# args are unchanged: no checkpoints, no upload — identical to before.
# SAVE_STEPS is the cadence knob (~1.5h at ~110s/step); the manifest carries
# enough checkpoint URLs to cover a full run at this cadence.
SAVE_STEPS=50
FULL_ARGS=(--data "${DATA}" --output "${FULL_OUT}")
if [[ -f "${MANIFEST}" ]]; then
  FULL_ARGS+=(--save-steps "${SAVE_STEPS}" --upload-manifest "${MANIFEST}")
  echo "RUN_SH_UPLOAD manifest=present save_steps=${SAVE_STEPS}"
else
  echo "RUN_SH_UPLOAD manifest=absent no_checkpoints_no_upload"
fi
python "${TRAIN_PY}" "${FULL_ARGS[@]}"
echo "RUN_SH_FULL_OK output=${FULL_OUT}"

echo "RUN_SH_DONE ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
