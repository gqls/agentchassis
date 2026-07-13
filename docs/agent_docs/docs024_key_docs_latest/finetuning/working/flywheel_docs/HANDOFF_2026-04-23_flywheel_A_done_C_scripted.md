# HANDOFF — 2026-04-23 — Flywheel A done, Flywheel C scripted

## Context in one paragraph

Today's session took flywheel A (training data export) from a broken v2 (writing files to ephemeral chassis pods) through three iterations to a working v3.2 — Postgres-backed, transactional, with proper error handling. Two clean training datasets now live in `training_exports` (1,958 rows each). Flywheel C (GPU training) was scripted — five files ready to run Unsloth QLoRA on Llama 3.3 70B Instruct on a single H100/A100 80GB VM. Phase 2 (automating C as a chassis-driven flywheel with VM as HTTP job server) was designed but not built.

## Current state

```
Flywheel A — training data export pipeline    ✓ operational
Flywheel B — RAG infrastructure                 ✓ done (prior session)
Flywheel C — full training loop (GPU)           → scripts ready, not yet run on VM
Flywheel D — Claude vs Ollama eval              ⏸ paused (eval case 2 timed out)
```

## Primary dataset

```
export_id:     146a9a12-c953-48eb-bf1f-c1856e5f13b7
agent_type:    page-content-writer
step_name:     process_sections_loop_iter_0_generate_content
model_filter:  claude-sonnet-4-6
rows:          1,958 (row_index 0-1957)
rows_seen:     1,960 (2 skipped as invalid JSON — expected)
size_bytes:    21,450,116 (~20MB)
format:        chatml (messages + metadata JSONB)
completed_at:  2026-04-23 14:54:32 UTC
```

Backup dataset also available: `fef7be6b-887f-4bc9-b118-a5a9992c4179` (earlier run, slightly smaller).

## Flywheel A — what was built today

### Schema (applied, verified)

`training_exports.runs` (one row per export, filter + counts + timestamps)
`training_exports.rows` (one row per training record, FK CASCADE to runs)

Three indexes plus a unique constraint on `(export_id, metadata->>'source_log_id')` to prevent duplicate source rows within one export. `completed_at IS NULL` identifies incomplete runs.

### Action code

`platform/orchestration/actions/training_data_export.go` v3.2:
- Uses canonical `ExtractActionInputs` + `ActionInputSpec` pattern from doc 001
- Per-batch transactions (100 rows each) instead of one big tx
- Strict error handling on final UPDATE (RowsAffected check)
- Returns `export_id` in result map (no more file paths)

### Agent definitions

`training-data-exporter` — worker, category=`specialist`, agent_category=`specialist`, processing_mode=`task`
`training-data-export-orchestrator` — wrapper, category=`orchestrator`, agent_category=`coordinator`, processing_mode=`orchestrator`

Orchestrator's workflow: `spawn_agent → call_agent(target_role=exporter) → complete_workflow`. Input mapping maps individual fields explicitly (not `input_data` blob), with `?`-suffix for optionals.

### Files in `flywheel_A_v3/` (all in `/mnt/user-data/outputs/flywheel_A_v3/`)

| File | State |
|---|---|
| `001_training_exports_schema.sql` | applied to clients_db |
| `002_create_training_export_agents_v3.sql` | applied |
| `training_data_export_v3.go` | deployed as v3.2, verified working |
| `trigger_training_export_v3.sh` | tested, produces clean exports |
| `003_backfill_jsonl_to_postgres.sh` | has a COPY-step bug, not worth fixing — we have two valid exports already |

## Flywheel C — what's scripted but not yet run

### Files in `flywheel_C/` (all in `/mnt/user-data/outputs/flywheel_C/`)

| File | Purpose |
|---|---|
| `00_vm_setup.sh` | Python 3.12 venv + CUDA-matched torch + Unsloth + deps. Verifies 80GB VRAM, bf16 support. |
| `01_pull_dataset_from_postgres.sh` | `kubectl exec` + `COPY TO STDOUT` to stream training data to a local JSONL on the VM |
| `02_train_llama_3_3_70b.py` | Main training script. Defaults: 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096. Target model: `unsloth/Llama-3.3-70B-Instruct-bnb-4bit` |
| `03_inference_test.py` | Quick sanity check — generates N outputs from training prompts, reports JSON validity + keys |
| `README.md` | Prereqs, sequence, expected timings, tuning table |

### Base model decision — Llama 3.3 70B Instruct

Rationale captured in FOCUS §2.5. 70B because hardware is available. Follow-up comparison against 8B is planned — 8B likely delivers 95% of quality at 10% of inference cost for this narrow structured-JSON task.

### Expected run parameters

```
Base model (4-bit):    ~38GB VRAM to load
LoRA + optimizer:      ~50-70GB peak VRAM
Training time:         30-90 min on H100 for 1,958 rows × 3 epochs
Adapter output size:   ~150MB
```

### Run sequence

```bash
# On GPU VM (one-off)
chmod +x 00_vm_setup.sh && ./00_vm_setup.sh
source ~/unsloth_env/bin/activate

# Pull dataset (or scp from laptop if you prefer)
chmod +x 01_pull_dataset_from_postgres.sh
./01_pull_dataset_from_postgres.sh \
    146a9a12-c953-48eb-bf1f-c1856e5f13b7 \
    /workspace/training_iter0.jsonl

# Smoke first (20 rows, 1 epoch, ~5-10 min)
python 02_train_llama_3_3_70b.py \
    --data /workspace/training_iter0.jsonl \
    --output /workspace/lora_smoke \
    --limit 20 --epochs 1

python 03_inference_test.py \
    --adapter /workspace/lora_smoke \
    --data /workspace/training_iter0.jsonl \
    --n 3

# If smoke looks sensible — full run (~30-90 min)
python 02_train_llama_3_3_70b.py \
    --data /workspace/training_iter0.jsonl \
    --output /workspace/lora_iter0_full

python 03_inference_test.py \
    --adapter /workspace/lora_iter0_full \
    --data /workspace/training_iter0.jsonl \
    --n 5 --skip 1900
```

### What success looks like

- `train_loss` drops from ~2.5 → ~0.5-1.0
- VRAM peak 50-70GB, no CUDA OOM
- Inference outputs parse as JSON (no code fences, no preamble)
- Keys match trained schemas: `headline`/`subheadline`/`primary_cta`/`primary_cta_url`/`secondary_cta`/`secondary_cta_url` (68% of training data), or minimal `headline`/`subheadline` (18%)
- Content is topical to the brief, not generic placeholder

## Alternative to `kubectl exec` in step 1

User preference is to not have the GPU VM connect to the cluster. Workaround for phase 1: run the psql COPY on laptop, scp the result to the VM:

```bash
# On laptop
kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- \
    psql -U clients_user -d clients_db -tA -c "
COPY (SELECT jsonb_build_object('messages', messages, 'metadata', metadata)
      FROM training_exports.rows WHERE export_id = '146a9a12-c953-48eb-bf1f-c1856e5f13b7'::uuid
      ORDER BY row_index) TO STDOUT
" > training_iter0.jsonl

scp training_iter0.jsonl user@gpu-vm:/workspace/training_iter0.jsonl
```

Skips step 1 entirely, VM stays credential-free.

## Flywheel C Phase 2 — designed, not built

Full design in FOCUS §2.5.1. Summary:

**Architecture:** chassis drives, GPU VM serves. Chassis agents call an HTTP job server running on the VM.

**VM side (~200 lines of Python):**
- `POST /jobs` — accepts dataset + hyperparameters, spawns training subprocess, returns job_id
- `GET /jobs/{id}` — returns status, progress, loss
- `GET /jobs/{id}/adapter` — downloads resulting LoRA
- Bearer-token auth, systemd unit, TLS via Caddy or cloud LB

**Chassis side — three new components:**
1. `model-trainer` specialist — fetches export rows, POSTs to VM, polls, downloads adapter, records in `model_training_runs`
2. `model-evaluator` specialist — runs flywheel D test harness against the new adapter, scores vs Claude
3. `training-flywheel-orchestrator` wrapper — chains export → train → evaluate → conditional deploy

**Schema additions:**
- `model_training_runs` — id, export_id FK, adapter_path, final_loss, hyperparameters, timings
- `model_artefacts` — id, training_run_id FK, storage_path, sha256
- `model_evaluations` — id, training_run_id FK, scores, deployment_decision

**Preconditions:**
- Phase 1 complete (manual run proves scripts work)
- Flywheel D eval harness working or known-paused
- VM has stable public endpoint with TLS

Rejected alternatives in design discussion: SSH + remote exec (synchronous, fragile), Kafka consumer on VM (overkill for single consumer).

## Lessons captured today (now in FOCUS §14)

1. **Long-held transactions through pgbouncer are fragile** — default to per-batch commits for bulk work, not one big tx
2. **Check RowsAffected() on single-row UPDATEs** — can silently return 0 rows with no error
3. **Three confused columns on `agent_definitions`:** `category` (free-text: orchestrator/specialist), `agent_category` (CHECK-constrained to 6 values, no `orchestrator` allowed), `status` (lifecycle: active/experimental/deprecated/demo/template). Reference: `improvement-loop` has category=orchestrator, agent_category=coordinator, status=experimental
4. **Use `ExtractActionInputs` + `ActionInputSpec`** — canonical input extraction pattern. Register via `init()` with `RegisterActionInputSpec`. Don't duplicate with direct `ExtractNestedFieldString` calls
5. **Orchestrator wrapper input_mapping — map fields individually, not as input_data blob.** Use `target_role` not `agent_type` for `call_agent` lookup. Suffix `?` on optional fields
6. **Kafka trigger JSON must be flat single-line** — use `jq -nc` + `<<<"$PAYLOAD"` here-string, not multi-line heredoc
7. **`kubectl cp` truncates large files silently** — use `kubectl exec $POD -- cat FILE > LOCAL` instead

## Resumption checklist

When you come back, three obvious entry points:

**A. Run flywheel C phase 1 on the VM** (most likely next step)
- Transfer the scripts to the VM
- Run through the sequence above
- Report back: smoke training loss, smoke inference JSON quality, then same for full run

**B. Build flywheel C phase 2** (substantial session, probably after A)
- VM-side HTTP service (FastAPI + systemd + TLS)
- Three new chassis agents with definitions
- Three new tables
- Requires phase 1 proven working as reference point

**C. Something else entirely**
- Flywheel D eval timeout investigation (Ollama case 2 hit 30min ceiling)
- Back to main site work
- Whatever's pressing at the time

## Key references for resumption

| Document | Purpose |
|---|---|
| `/mnt/user-data/outputs/FOCUS_finetuning_flywheel_and_service.md` | Full working memory — read §2.4-2.5.1 and §15 changelog first |
| `/mnt/user-data/outputs/flywheel_A_v3/` | All flywheel A v3.2 artefacts (applied, working) |
| `/mnt/user-data/outputs/flywheel_C/` | Training scripts ready to run |
| `001_development_guide.md` in project | Canonical patterns for chassis work |
| Previous transcript (per compaction notes) | Full detail of debugging journey |

## Known issues / technical debt

- `003_backfill_jsonl_to_postgres.sh` has a COPY-step bug. Not fixed; two valid exports already cover the data.
- Flywheel D eval paused at Ollama case 2 (30-min wall-time ceiling hit). Will need smaller `num_predict` or different strategy to restart.
- No current baseline for "how good is the trained model" — phase 1 produces that baseline, phase 2 automates evaluation.

## Decisions locked today

- **Postgres storage over file outputs for training data** — in FOCUS §2.4g changelog
- **Manual trigger for flywheel A exports** (no scheduling yet) — simplest, easy to upgrade later
- **Llama 3.3 70B Instruct as first training target** — hardware available; 8B comparison follow-up planned
- **Unsloth for training** — fits single-GPU 80GB, 2x faster + 70% less VRAM than vanilla HF
- **Phase 1 before phase 2** — prove manual before automating
- **HTTP job server on VM for phase 2** — Option B, not SSH, not Kafka
