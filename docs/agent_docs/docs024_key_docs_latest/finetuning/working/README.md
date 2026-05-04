# Flywheel C — first training run

Fine-tune Llama 3.3 70B Instruct on the page-content-writer iter_0 dataset
(1,958 examples) using QLoRA on a single H100 80GB or A100 80GB VM.

## Files

| File | Purpose |
|---|---|
| `00_vm_setup.sh` | One-time: install Python, PyTorch, Unsloth, verify GPU |
| `01_pull_dataset_from_postgres.sh` | Stream training data from `training_exports.rows` into a local JSONL |
| `02_train_llama_3_3_70b.py` | The QLoRA training script |
| `03_inference_test.py` | Quick sanity check — generate outputs and inspect them |

## Prerequisites

- GPU VM with H100 80GB or A100 80GB (single card is enough for QLoRA on 70B)
- NVIDIA driver installed (`nvidia-smi` works)
- `kubectl` on the VM with access to `ai-persona-system` cluster, **OR**
  direct psql connectivity to `clients_db` (edit the pull script accordingly)
- ~150GB free disk (base model download ~40GB + workspace)
- ~64GB system RAM (Unsloth offloads activations to system RAM)

## Run

```bash
# 0. Setup — first time only, takes 5-10 minutes
chmod +x 00_vm_setup.sh
./00_vm_setup.sh

# Activate the env in new shells
source ~/unsloth_env/bin/activate

# 1. Pull dataset. Pick an export_id from training_exports.runs.
chmod +x 01_pull_dataset_from_postgres.sh
./01_pull_dataset_from_postgres.sh \
    146a9a12-c953-48eb-bf1f-c1856e5f13b7 \
    /workspace/training_iter0.jsonl

# 2. Smoke train — 20 rows, 1 epoch, ~5 minutes — verifies the whole pipeline works
python 02_train_llama_3_3_70b.py \
    --data /workspace/training_iter0.jsonl \
    --output /workspace/lora_smoke \
    --limit 20 \
    --epochs 1

# 3. Smoke inference test
python 03_inference_test.py \
    --adapter /workspace/lora_smoke \
    --data /workspace/training_iter0.jsonl \
    --n 3

# 4. Full training run — 1958 rows, 3 epochs, ~30-90 minutes on H100
python 02_train_llama_3_3_70b.py \
    --data /workspace/training_iter0.jsonl \
    --output /workspace/lora_iter0_full

# 5. Inference test on full LoRA
python 03_inference_test.py \
    --adapter /workspace/lora_iter0_full \
    --data /workspace/training_iter0.jsonl \
    --n 5 \
    --skip 1900   # pull samples from end of file — approximation of held-out
```

## Expected timings on H100 80GB

| Phase | Time |
|---|---|
| First `00_vm_setup.sh` | 5-10 min (depends on network for torch wheels) |
| Base model download (first train only) | 10-15 min (40GB from HF) |
| Smoke train (20 rows, 1 epoch) | 5-10 min |
| Full train (1958 rows, 3 epochs) | 30-90 min |
| Inference test per prompt | 5-15 sec |

## What success looks like

**Training:**
- `train_loss` drops from ~2.5 at start to ~0.5-1.0 by end (exact values
  depend on data difficulty)
- VRAM peak should be around 50-70 GB (well under the 80GB ceiling)
- No CUDA OOM errors

**Inference test:**
- Generated outputs parse as JSON (json.loads succeeds)
- Keys match what we trained on: `headline`, `subheadline`, `primary_cta`,
  `primary_cta_url`, `secondary_cta`, `secondary_cta_url` for the common hero
  schema, or `headline`, `subheadline` for the minimal hero
- Content is topical to the prompt brief (not generic placeholder)
- No preamble like "Here's the JSON..."
- No trailing ```markdown``` fences

## What to tune if it goes wrong

| Symptom | Try |
|---|---|
| CUDA OOM during training | Lower `--max-seq-length` to 2048 |
| Training loss not dropping | Increase `--lr` to 3e-4 or `--lora-r` to 32 |
| Outputs ignore the brief (generic content) | More epochs (`--epochs 5`) |
| Outputs don't parse as JSON | More epochs + check your expected outputs in training data ARE JSON |
| First batch takes forever | Usually model download + CUDA kernel warmup; should be fast thereafter |

## After this — the rest of flywheel C

1. Evaluate on held-out prompts (reuse flywheel D test_cases.jsonl)
2. Claude-as-judge comparison: Claude vs trained Llama 70B
3. Decide deployment path:
   - Merge + export GGUF for Ollama serving
   - Serve raw LoRA + base via vLLM on GPU VM
   - Upload to Together AI / Fireworks for managed inference
4. Wire into chassis as an alternative `ai_service` provider
5. Measure cost-per-page on production workloads

For now: get a trained LoRA, inspect its outputs, decide if it's good enough
to productionise.
