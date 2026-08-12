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

# Optional env overrides (2026-07-31, finetuning_uk_service). Unset = behaviour
# identical to before: 02_train's own 70B default, checkpoint every 10 steps.
#   BASE_MODEL  — passed through as --base-model (e.g. an unsloth 1B/3B for the
#                 paid-demo runs). Set it on the invocation: BASE_MODEL=... run.sh
#                 (the launcher's ssh_exec command template must export it when
#                 the automated path adopts this — not wired yet).
#   SAVE_STEPS  — checkpoint cadence; 0 = no checkpoints (right for runs shorter
#                 than one cadence interval). The manifest is still passed when
#                 present: the FINAL adapter upload is what makes DONE durable.
#                 ⚠ The default MUST stay 10 — that is what the live B2 bundle
#                 hardcoded before this file was env-parameterised. It was
#                 written here as 50 (2026-07-31) alongside the claim "unset =
#                 behaviour identical to before"; the two contradicted each
#                 other and nothing failed, because no 70B run has started
#                 since 2026-06-13. Measured against the live bundle 2026-08-11.
#   CHAT_TEMPLATE / INSTRUCTION_PART / RESPONSE_PART (added 2026-08-11)
#                 — the prompt format. These MUST move together with BASE_MODEL:
#                 02_train hardcoded the llama-3.1 template and the two
#                 <|start_header_id|> masking literals, so BASE_MODEL alone
#                 rendered training text out of tokens a non-Llama model has
#                 never seen, and response-only masking then missed silently.
#                 For any *-Instruct repo that ships its own template, use
#                 CHAT_TEMPLATE=auto plus that model's real turn markers.
#                 SmolLM2-1.7B-Instruct (verified against its
#                 tokenizer_config.json, 2026-08-11) is ChatML:
#                   CHAT_TEMPLATE=auto
#                   INSTRUCTION_PART='<|im_start|>user\n'
#                   RESPONSE_PART='<|im_start|>assistant\n'
#                 Unset = the llama-3.1 defaults inside 02_train, unchanged.
BASE_MODEL="${BASE_MODEL:-}"
SAVE_STEPS="${SAVE_STEPS:-10}"
CHAT_TEMPLATE="${CHAT_TEMPLATE:-}"
INSTRUCTION_PART="${INSTRUCTION_PART:-}"
RESPONSE_PART="${RESPONSE_PART:-}"

# Assembled once and reused by both the smoke and the full pass, so the two can
# never disagree about the prompt format.
FORMAT_ARGS=()
if [[ -n "${CHAT_TEMPLATE}" ]];    then FORMAT_ARGS+=(--chat-template "${CHAT_TEMPLATE}"); fi
if [[ -n "${INSTRUCTION_PART}" ]]; then FORMAT_ARGS+=(--instruction-part "${INSTRUCTION_PART}"); fi
if [[ -n "${RESPONSE_PART}" ]];    then FORMAT_ARGS+=(--response-part "${RESPONSE_PART}"); fi

echo "RUN_SH_START ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "RUN_SH_MODEL base_model=${BASE_MODEL:-default-70b} save_steps=${SAVE_STEPS} chat_template=${CHAT_TEMPLATE:-llama-3.1-default}"

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
SMOKE_ARGS=(--data "${DATA}" --output "${SMOKE_OUT}" --limit 20 --epochs 1)
if [[ -n "${BASE_MODEL}" ]]; then
  SMOKE_ARGS+=(--base-model "${BASE_MODEL}")
fi
SMOKE_ARGS+=("${FORMAT_ARGS[@]+"${FORMAT_ARGS[@]}"}")
python "${TRAIN_PY}" "${SMOKE_ARGS[@]}"
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
# SAVE_STEPS is the cadence knob (~1.5h at ~110s/step at 70B; env-overridable,
# 0 = no checkpoints for short runs); the manifest carries enough checkpoint
# URLs to cover a full run at this cadence.
FULL_ARGS=(--data "${DATA}" --output "${FULL_OUT}")
if [[ -n "${BASE_MODEL}" ]]; then
  FULL_ARGS+=(--base-model "${BASE_MODEL}")
fi
FULL_ARGS+=("${FORMAT_ARGS[@]+"${FORMAT_ARGS[@]}"}")
if [[ -f "${MANIFEST}" ]]; then
  FULL_ARGS+=(--save-steps "${SAVE_STEPS}" --upload-manifest "${MANIFEST}")
  echo "RUN_SH_UPLOAD manifest=present save_steps=${SAVE_STEPS}"
else
  echo "RUN_SH_UPLOAD manifest=absent no_checkpoints_no_upload"
fi
python "${TRAIN_PY}" "${FULL_ARGS[@]}"
echo "RUN_SH_FULL_OK output=${FULL_OUT}"

echo "RUN_SH_DONE ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
