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
#   3. full train            — defaults (3 epochs) to /workspace/adapter_out
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
#   (a later artefact-collector presigns a PUT for this dir's tarball)
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
python "${TRAIN_PY}" \
  --data "${DATA}" \
  --output "${FULL_OUT}"
echo "RUN_SH_FULL_OK output=${FULL_OUT}"

echo "RUN_SH_DONE ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
