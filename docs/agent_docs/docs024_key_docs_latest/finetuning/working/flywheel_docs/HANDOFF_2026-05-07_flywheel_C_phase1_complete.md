# HANDOFF — 2026-05-07 — Flywheel C phase 1 complete (iter_0 adapter trained)

## Context in one paragraph

Today's session executed flywheel C phase 1 end-to-end on a Thunder Compute A100 80GB instance. The 1,958-row training set produced from flywheel A's `training_exports.runs/rows` schema fine-tuned `Llama-3.3-70B-Instruct-bnb-4bit` via Unsloth QLoRA over 9h 12m at a real cost of ~$20. The iter_0 adapter (791MB safetensors) is exfiltrated to the laptop. Inference smoke on 5 held-back training rows returned 5/5 valid JSON with correct schemas. Instance deleted, snapshots not retained (the math didn't justify them for our run frequency). Phase 2 (chassis-driven training) is unchanged from the previous handoff's design but now has concrete cost/time numbers feeding it.

## Current state

```
Flywheel A — training data export pipeline    ✓ operational
Flywheel B — RAG infrastructure                 ✓ done (prior session)
Flywheel C phase 1 — manual training run        ✓ iter_0 adapter shipped
Flywheel C phase 2 — chassis automation         ⏸ designed, not built
Flywheel D — Claude vs adapter eval             ⏸ paused (next priority)
```

## iter_0 training run summary

```
adapter_path:    ./lora_iter0_full/ (laptop, ~791MB)
base_model:      unsloth/Llama-3.3-70B-Instruct-bnb-4bit
dataset:         training_exports export 146a9a12-c953-48eb-bf1f-c1856e5f13b7
                 1,958 rows → 1,934 effective (24 dropped: assistant turn truncated past max_seq)
hyperparameters: 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096
final_loss:      0.2669 (trailing avg; actual end-of-run ~0.10-0.13)
loss curve:      ep1 1.49→0.27 | ep2 0.34→0.18 | ep3 0.14→0.10
                 (clean curve, no spikes, no NaN, gradient norms stayed in 0.3-0.5)
peak_vram:       44.8GB of 80GB available (room to spare)
wall time:       33,136s ≈ 9h 12m
```

A note on the loss trajectory worth carrying into iter_1 planning: the gap between epoch 2 mean (~0.23) and epoch 3 mean (~0.12) suggests increasing memorisation rather than generalisation in epoch 3. Whether that hurts inference quality depends on flywheel D — but a `--epochs 2` ablation on iter_1 would be informative comparative data.

## Version pin discoveries (essential for any future run)

The Ollama-template-on-Thunder + `cu124` torch index is a working but narrow combination. The pins below are not cosmetic:

| Package | Pin | Why |
|---|---|---|
| torch | 2.6.0+cu124 | last torch on the cu124 wheel index; `cu124` was the cuda detected on Thunder's Ollama template |
| transformers | <5 (lands at 4.57.6) | 5.x is brand new; unsloth's tested band is 4.46-4.55ish |
| torchao | <0.17 | 0.17 calls `torch.utils._pytree.register_constant` which only exists in torch 2.7+. Without this pin, transformers (any version) fails at module load with an AttributeError because it imports torchao eagerly |
| flash-attn | 2.7.4+cu124torch2.6 | mjun0812 prebuilt wheel; the official flash-attn build needs nvcc, which the Ollama template does NOT ship (CUDA runtime, not toolkit). Source URL: `https://github.com/mjun0812/flash-attention-prebuild-wheels/releases/download/v0.7.16/flash_attn-2.7.4+cu124torch2.6-cp312-cp312-linux_x86_64.whl` |
| unsloth + unsloth_zoo | latest from PyPI | git+ install of unsloth does NOT pull unsloth_zoo as a transitive dep — must install both explicitly |
| huggingface_hub | latest, but NOT `huggingface_hub[hf_transfer]` | the `hf_transfer` extra was removed from hf-hub 1.x. Install `hf_transfer` as a separate package |

Long-term: `cu124` is becoming a dead end (no new torch wheels published). Next significant rebuild should switch to `cu126` or `cu128` index for torch 2.7+, which removes the torchao/transformers pin maintenance burden.

## Cost data (real, from this session)

| Stage | Wall time | Cost @ $1.79/hr Production |
|---|---|---|
| Setup + smoke (incl. flash-attn debugging) | ~1.5h | ~$2.70 |
| Full 3-epoch training run | 9.2h | $16.50 |
| Inference smoke + adapter exfil | ~30min | ~$0.90 |
| **Total iter_0** | **~11.2h** | **~$20** |

Estimates I gave during the session were consistently low — initial $2-3 quote was based on smoke-test step times that turned out not to be representative. Steady-state at 45s/step on A100 80GB with FA2, max_seq 4096, and LoRA r=16 on 70B is just genuinely expensive. **Anchor future estimates against $20/iter, not whatever Unsloth marketing materials suggest.**

Production vs Prototyping mode wasn't tested for training in this session — we picked Production for stability. Prototyping ($0.78/hr advertised) might cut costs in half but uses Thunder's TGV virtualised GPUs, and we don't know whether 70B QLoRA training survives the virtualisation overhead. Worth a one-off experiment on iter_1.

## Snapshot decision

Created `unsloth-trainer-base-01` then deleted it. Math:

- Snapshot: 100GB provisioned (Thunder snapshots whole disk, not used bytes) × $0.15/GB/month = $15/month standing cost
- Cold-start savings vs setup script: ~25 min, ~$0.85 at $1.79/hr
- Break-even: ~18 training runs/month
- Reality: 1-4 runs/month

The `00_vm_setup.sh` we ended up with after all the pinning discoveries is the real "snapshot" — version-pinned, idempotent, ~5 min slower than restoring from a Thunder snapshot, but free and version-controlled. That script (and `02_train_llama_3_3_70b.py`) need to be in the project repo if they're not already.

## Files / artefacts after this session

| File | Where | Note |
|---|---|---|
| `lora_iter0_full/adapter_model.safetensors` | laptop, 791MB | the actual trained adapter |
| `lora_iter0_full/manifest.json` | laptop | hyperparams + loss + runtime + peak_vram |
| `lora_iter0_full/adapter_config.json` | laptop | peft config — needed to reload |
| `00_vm_setup.sh` | laptop, conversation-canonical | autodetects CUDA, applies all version pins |
| `02_train_llama_3_3_70b.py` | laptop, conversation-canonical | uses SFTConfig (not TrainingArguments), train_on_responses_only, peak VRAM tracking, manifest emit |
| `HANDOFF_2026-04-23_..._PATCH_2026-05-06.md` | laptop | corrections to the kubectl-exec dataset-pull command |

The 791MB safetensors doesn't belong in git. Backblaze (`s3://...your.../adapters/iter0/`) per the wider project's existing pattern is right. The two scripts and the manifest do belong in git.

## Resumption — order of work for next session(s)

### 1. Flywheel D — evaluate iter_0 against Claude on novel briefs (next priority)

Until we know whether iter_0 is usable on *novel* briefs (not training rows), automating its production is automating something whose output may need to be discarded. The handoff from 2026-04-23 noted flywheel D was paused at "Ollama case 2 timed out (30-min wall-time ceiling)". Two unknowns to address:

- **Where do we serve the adapter for inference?** Options: (a) fresh Thunder instance with Unsloth's `FastLanguageModel` (simplest, costs GPU-time), (b) convert adapter to GGUF and serve via Ollama on the existing `gpu-ollama` endpoint that's currently DOWN, (c) vLLM with LoRA support. Option (a) is simplest but most expensive; we need to think about which makes sense for repeated eval runs.
- **Why did Ollama case 2 timeout?** Likely `num_predict` set too high or the prompt itself triggering long generation. Smaller `num_predict` or a different test set.

The eval should run novel briefs (real production sites that *aren't* in the training set) through both Claude Sonnet 4.6 and the iter_0 adapter, then score on:
- JSON validity
- Schema compliance (correct fields, no hallucinated ones)
- Brand/voice match against the brief
- Factual correctness against the brief (no invented case studies, metrics, contact info — RULE 11-14 from the training prompts)

Output of this work: a comparison table that tells us whether iter_0 ships, or what to filter for iter_1.

### 2. Flywheel C phase 2 — chassis-driven training pipeline

Phase 2 design from the previous handoff is unchanged in shape, but now has real numbers feeding the agent definitions and table schemas:

- `model_training_runs` table — needs columns for: `train_runtime_s`, `final_loss`, `peak_vram_gb`, `cost_usd`, the manifest fields, FK to `training_exports.runs`
- `model_artefacts` table — `storage_path` will be Backblaze, `sha256` for integrity
- `model_evaluations` table — feeds from flywheel D output
- `model-trainer` specialist — wraps Thunder Compute API calls (`POST /instances/{id}/up`, `POST /jobs` to the on-VM HTTP server, `POST /instances/{id}/down`)
- `model-evaluator` specialist — runs flywheel D harness against new adapter
- `training-flywheel-orchestrator` wrapper — chains export → train → evaluate → conditional-deploy

Two new design considerations from this session:
- **Snapshots aren't worth it at low run frequency.** Phase 2 should provision fresh instances and run `00_vm_setup.sh` rather than restore from snapshot. ~25 min is fine when the chassis owns the lifecycle.
- **Adapter saves as fp32 by default** — 791MB for an adapter that could be 400MB. Phase 2's `model-trainer` should cast LoRA weights to fp16 before saving. One-line fix in the training script:
  ```python
  for p in model.parameters():
      if p.requires_grad:
          p.data = p.data.to(torch.float16)
  ```

### 3. Back to main site work

The flywheel was a side-quest. The patch document accumulates lessons regardless.

## Lessons captured today (worth folding into FOCUS § 14)

1. **`COPY ... TO STDOUT` is not JSON-safe** — the TEXT format escape layer collides with `jsonb`'s own escapes. Use `psql -tAXc` with plain `SELECT` for JSONL extraction. (See PATCH_2026-05-06.)
2. **`kubectl exec -i` truncates stdout** when stdin isn't consumed — drop `-i` for non-interactive commands. Got 1716/1958 rows on the first pull because of this. (See PATCH_2026-05-06.)
3. **Thunder's Ollama template provides CUDA runtime, not toolkit.** No nvcc, no `CUDA_HOME` set. Source compiles of GPU-bound packages (flash-attn, anything that uses `torch.utils.cpp_extension.CUDAExtension`) won't work. Use prebuilt wheels.
4. **`unsloth @ git+...` does not pull `unsloth_zoo`.** Use PyPI install (`pip install unsloth unsloth_zoo`), not the git URL.
5. **`transformers` (4.x and 5.x both) imports torchao eagerly at package load.** If torchao is incompatible with installed torch, transformers can't even import. The fix isn't to pin transformers — it's to pin torchao.
6. **`huggingface_hub[hf_transfer]` extra no longer exists in hf-hub 1.x.** Install `hf_transfer` separately.
7. **`SFTTrainer` (trl 0.24) wants `SFTConfig`, not `TrainingArguments`.** SFT-specific options (`max_seq_length`, `dataset_text_field`, `packing`, `dataset_num_proc`) live in `SFTConfig`.
8. **`nohup python ... > log 2>&1 &` is block-buffered for stdout**, line-buffered for stderr (where tqdm writes). Loss prints go to stdout and stay in the buffer for hours. Use `python -u` or `PYTHONUNBUFFERED=1` for any long ML run with redirected stdout.
9. **Unsloth banner reads `FA2 = True/False`** to confirm flash-attn detection. Flag-True after install means the wheel was picked up; if False, no FA2 in this session — restart the process to refresh detection.
10. **PEFT's `save_pretrained()` defaults to fp32 for LoRA weights** even when training was in bf16. Cast to fp16 before save to halve adapter size.
11. **A100 80GB Production at $1.79/hr; full QLoRA run on 70B at max_seq 4096 with FA2 is ~9-10h, ~$20.** Anchor future estimates against this, not marketing-material projections.
12. **Thunder snapshots are 100GB regardless of used space** because they snapshot the whole provisioned disk. At $0.15/GB/month that's $15/month standing. Not worth it below ~18 training runs/month.

## Known issues / technical debt

- The `<no value>` rendering bug in the prompt builder was fixed yesterday/this morning. Training rows from before that fix include the literal token `<no value>`. iter_1's export should filter `created_at >= <fix_date>` to exclude pre-fix rows. **Action: someone (you?) note the fix-deploy date so the iter_1 export can use it as a cutoff.**
- Iter_0 has only been smoke-tested on training rows. Real evaluation against novel briefs is flywheel D's job.
- `gpu-ollama` endpoint is still DOWN. If flywheel D wants to reuse it, that's a separate fix. Otherwise we provision fresh GPU each eval run.
- The 30-minute wall-time ceiling on flywheel D's eval harness needs investigation — likely `num_predict` too high.

## Decisions locked today

- **Llama 3.3 70B Instruct + Unsloth QLoRA + max_seq 4096 + lora_r 16 + 3 epochs** is iter_0's training config. Iter_1 will likely vary epochs (2 vs 3) and may try lora_r 32.
- **Production mode on Thunder for training**, at least until a Prototyping mode test demonstrates it's usable.
- **No snapshots kept** — `00_vm_setup.sh` is the canonical environment recipe.
- **fp16 adapter saves for iter_1** to halve transfer/storage cost.
- **Flywheel D before flywheel C phase 2.** Don't automate a process whose product hasn't been validated.
