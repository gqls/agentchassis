#!/usr/bin/env bash
# 00_vm_setup.sh — one-off setup for Llama 3.3 70B QLoRA training
# Target: Thunder Compute H100/A100 80GB instance, Ollama template (Ubuntu 22.04 + CUDA)
# Idempotent — safe to re-run.

set -euo pipefail

VENV_DIR="${HOME}/unsloth_env"
WORKSPACE="/workspace"
PY_BIN="python3.12"

log() { echo "[$(date +%H:%M:%S)] $*"; }

# Detect CUDA version from nvidia-smi and map to torch wheel index.
# Override with: CUDA_TAG=cu121 ./00_vm_setup.sh
detect_cuda_tag() {
  local cuda_str
  cuda_str=$(nvidia-smi | grep -oE 'CUDA Version: [0-9]+\.[0-9]+' | awk '{print $3}')
  case "${cuda_str}" in
    12.1*) echo "cu121" ;;
    12.4*) echo "cu124" ;;
    12.6*) echo "cu126" ;;
    12.8*) echo "cu128" ;;
    13.0*) echo "cu130" ;;
    *)     echo "cu124" ;;  # safe modern default
  esac
}

# ---------------------------------------------------------------------------
# 1. GPU / driver sanity
# ---------------------------------------------------------------------------
log "Checking GPU..."
if ! command -v nvidia-smi >/dev/null; then
  echo "ERROR: nvidia-smi not found. Did you pick the Ollama template?" >&2
  exit 1
fi

VRAM_MIB=$(nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits | head -1 | tr -d ' ')
DRIVER=$(nvidia-smi --query-gpu=driver_version --format=csv,noheader | head -1)
GPU_NAME=$(nvidia-smi --query-gpu=name --format=csv,noheader | head -1)
log "GPU: ${GPU_NAME}, ${VRAM_MIB} MiB VRAM, driver ${DRIVER}"

if [ "${VRAM_MIB}" -lt 79000 ]; then
  echo "ERROR: ${VRAM_MIB} MiB < 79000 MiB. Llama 3.3 70B QLoRA needs an 80GB card." >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# 2. Workspace
# ---------------------------------------------------------------------------
sudo mkdir -p "${WORKSPACE}"
sudo chown "$(id -u):$(id -g)" "${WORKSPACE}"
log "Workspace: ${WORKSPACE}"

# ---------------------------------------------------------------------------
# 3. System packages + Python 3.12
# ---------------------------------------------------------------------------
log "Installing system packages..."
sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
  software-properties-common build-essential git jq curl wget tmux

if ! command -v ${PY_BIN} >/dev/null; then
  log "Adding deadsnakes PPA and installing Python 3.12..."
  sudo add-apt-repository -y ppa:deadsnakes/ppa
  sudo apt-get update -qq
  sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
    python3.12 python3.12-venv python3.12-dev
fi
log "Python: $(${PY_BIN} --version)"

# ---------------------------------------------------------------------------
# 4. Virtualenv
# ---------------------------------------------------------------------------
if [ ! -d "${VENV_DIR}" ]; then
  log "Creating venv at ${VENV_DIR}..."
  ${PY_BIN} -m venv "${VENV_DIR}"
fi
# shellcheck disable=SC1091
source "${VENV_DIR}/bin/activate"
pip install --upgrade pip wheel setuptools >/dev/null

# ---------------------------------------------------------------------------
# 5. Torch matched to driver's CUDA capability (autodetected from nvidia-smi)
# ---------------------------------------------------------------------------
CUDA_TAG="${CUDA_TAG:-$(detect_cuda_tag)}"
log "Installing torch (${CUDA_TAG})..."
pip install --quiet \
  torch torchvision torchaudio \
  --index-url "https://download.pytorch.org/whl/${CUDA_TAG}"

# ---------------------------------------------------------------------------
# 6. Unsloth + training stack (PyPI — git+ install does NOT pull unsloth_zoo)
# ---------------------------------------------------------------------------
# Two pins required for torch 2.6.0 (the latest available on the cu124 wheel
# index — there is no newer cu124 torch as of this writing):
#   - transformers<5    : 5.x is brand-new and untested with this stack
#   - torchao<0.17      : 0.17 calls torch.utils._pytree.register_constant,
#                         which was only added in torch 2.7. transformers
#                         (both 4.x and 5.x) imports torchao eagerly at
#                         package load, so the AttributeError surfaces
#                         the moment anything imports transformers.
# Long-term fix is to upgrade torch by switching to a newer wheel index
# (cu126 or cu128); cu124 is becoming a dead end.
log "Installing Unsloth + deps..."
pip install --quiet \
  unsloth \
  unsloth_zoo \
  "transformers<5" \
  "torchao<0.17"
pip install --quiet \
  datasets \
  trl \
  peft \
  accelerate \
  bitsandbytes \
  sentencepiece \
  protobuf \
  hf_transfer

# ---------------------------------------------------------------------------
# 7. Verify torch + CUDA + bf16
# ---------------------------------------------------------------------------
log "Verifying torch CUDA..."
python - <<'PY'
import torch, sys
print(f"  torch={torch.__version__}")
print(f"  cuda_available={torch.cuda.is_available()}")
if not torch.cuda.is_available():
    print("ERROR: CUDA not available to torch", file=sys.stderr)
    sys.exit(1)
print(f"  device={torch.cuda.get_device_name(0)}")
print(f"  vram_gib={torch.cuda.get_device_properties(0).total_memory / 1024**3:.1f}")
print(f"  bf16_supported={torch.cuda.is_bf16_supported()}")
if not torch.cuda.is_bf16_supported():
    print("  NOTE: bf16 unavailable, training will fall back to fp16.", file=sys.stderr)
PY

# ---------------------------------------------------------------------------
# 8. Verify Unsloth import (catches version-mismatch silently early)
# ---------------------------------------------------------------------------
log "Verifying Unsloth..."
python - <<'PY'
from unsloth import FastLanguageModel  # noqa: F401
import unsloth
print(f"  unsloth={unsloth.__version__}")
PY

# ---------------------------------------------------------------------------
# 9. Persist HF transfer accelerator for next shells (best-effort)
# ---------------------------------------------------------------------------
# On some base images ${HOME}/.bashrc is root-owned (the home dir is writable
# but the pre-seeded .bashrc is not). This step is only a convenience for future
# shells; the run does not depend on it, so it must NEVER abort setup (set -e).
# Guard on writability.
if [ -w "${HOME}/.bashrc" ] && ! grep -q HF_HUB_ENABLE_HF_TRANSFER "${HOME}/.bashrc" 2>/dev/null; then
  echo 'export HF_HUB_ENABLE_HF_TRANSFER=1' >> "${HOME}/.bashrc"
fi

log "Setup complete."
log "Activate with:  source ${VENV_DIR}/bin/activate"
log "Workspace:      ${WORKSPACE}"
log "Next step:      transfer the dataset, then run 02_train_llama_3_3_70b.py"
