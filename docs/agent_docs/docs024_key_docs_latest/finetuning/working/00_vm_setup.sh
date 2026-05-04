#!/bin/bash
# ============================================================================
# 00_vm_setup.sh — one-time environment setup on the GPU VM
# ============================================================================
# Creates a venv, installs Unsloth + friends, verifies CUDA / torch / GPU.
# Assumes Ubuntu-like Linux with an NVIDIA driver already installed (run
# `nvidia-smi` first and confirm you see your H100/A100 80GB before running
# this).
#
# Python 3.12 is the sweet spot — Unsloth explicitly supports 3.10-3.13.
# ============================================================================

set -euo pipefail

echo "=== GPU check ==="
nvidia-smi || { echo "nvidia-smi failed — driver not installed?"; exit 1; }
echo ""

# ── Python venv ─────────────────────────────────────────────────────────────

if ! command -v python3.12 &>/dev/null; then
    echo "Installing Python 3.12..."
    sudo apt update
    sudo apt install -y python3.12 python3.12-venv python3.12-dev git
fi

python3.12 -m venv ~/unsloth_env
source ~/unsloth_env/bin/activate

echo ""
echo "=== Python ==="
python --version
pip install --upgrade pip

# ── Install torch matched to CUDA ───────────────────────────────────────────
# Detect CUDA from nvidia-smi
CUDA_STR=$(nvidia-smi | grep -oE 'CUDA Version: [0-9]+\.[0-9]+' | awk '{print $3}')
echo ""
echo "=== CUDA detected: $CUDA_STR ==="
# Normalise to short form for torch wheel index
case "$CUDA_STR" in
    12.1*) TORCH_CUDA="cu121" ;;
    12.4*) TORCH_CUDA="cu124" ;;
    12.6*) TORCH_CUDA="cu126" ;;
    12.8*) TORCH_CUDA="cu128" ;;
    13.0*) TORCH_CUDA="cu130" ;;
    *)
        echo "Unknown CUDA version ($CUDA_STR) — defaulting to cu124. Adjust if this fails."
        TORCH_CUDA="cu124"
        ;;
esac
echo "Using torch index: $TORCH_CUDA"

pip install torch torchvision torchaudio \
    --index-url "https://download.pytorch.org/whl/$TORCH_CUDA"

# ── Install Unsloth + dependencies ─────────────────────────────────────────
# 2026-era: plain `pip install unsloth` works for modern CUDA on Linux.
pip install unsloth
pip install transformers datasets trl peft accelerate bitsandbytes

# ── Verify ──────────────────────────────────────────────────────────────────
echo ""
echo "=== Verifying install ==="
python <<'PY'
import torch
from unsloth import FastLanguageModel

print(f"torch:    {torch.__version__}")
print(f"cuda:     {torch.version.cuda}")
print(f"gpu:      {torch.cuda.get_device_name(0)}")
print(f"vram:     {torch.cuda.get_device_properties(0).total_memory / 1e9:.1f} GB")

# Check bfloat16 support (required for Llama 3.3 70B)
ok = torch.cuda.is_bf16_supported()
print(f"bf16:     {'yes' if ok else 'NO — H100/A100 should support this'}")

# Check we've got enough VRAM
free, total = torch.cuda.mem_get_info(0)
print(f"vram free: {free / 1e9:.1f} GB of {total / 1e9:.1f} GB")

if total < 75e9:
    raise SystemExit("Need at least ~80GB VRAM for 70B QLoRA. Abort.")

print("\nOK — ready to train.")
PY

echo ""
echo "To activate this env later:"
echo "  source ~/unsloth_env/bin/activate"
echo ""
echo "Next: run 01_pull_dataset_from_postgres.sh to get your training data."
