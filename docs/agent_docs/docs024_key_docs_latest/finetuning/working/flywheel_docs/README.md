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


--------------------------------------------

phase 3 scope:
Phase 2 is verified — the response is showing in the adapter logs:
"in_response_to_request_id":"cc2fd4e9-21a8-4c43-a592-f71fb207330f"   ← echoes our REQUEST_ID
"sender_agent_type":"thunder-adapter"                                  ← validator tier-1
"status":"error_unrecoverable","is_error":"true"                       ← correct vocabulary
"action":"provision_instance","error":"not_implemented"                 ← right handler reached
Headers and body shape exactly as designed. Now updating the FOCUS doc:Ran 2 commands, viewed a file, edited a fileRan 2 commands, viewed a file, edited a fileSummary of the FOCUS doc update
Replaced the old 56-line "Deployment essentials" section with a 200-line one that's structured around six sub-sections:
Sub-sectionCapturesThe full manifest patternAll three things missing from v1: serviceAccountName, imagePullSecrets, command: (vs args alone). With inline comments explaining the failure mode each one prevents.Required cluster resources before first deployFive-row table of dependencies — Secrets, SA, Docker Hub grant, and Kafka topics — with how to verify each. Plus the KafkaTopic CRD pattern.Service permissions (when applicable)Scoped Role/RoleBinding example for adapters managing k8s resources (relevant for thunder-adapter Phase 3+).Makefile integrationFour insertion points, plus the newName/newTag overlay gotcha (sed only updates newTag).Pre-deploy verification checklistEight items to tick off before make release-backend.Post-deploy verificationFour-step bring-up check.
Cross-references section 10 of the debugging guide so anyone hitting a deployment problem has a clear path to the symptom-to-cause table.
Phase 2 wrap
Verified end-to-end:

Schema migration 025 applied (thunder_config singleton seeded, thunder_instances empty, provision-check view returns can_provision=t)
Image built and pushed: docker.io/aqls/thunder-adapter:v1.0.1010
Pod running: 1/1 Ready
Kafka topic provisioned: system.thunder.smoke.responses
Round-trip verified: request → adapter consumed → not_implemented error response → producer sent → headers and body shape as designed

Three docs updated this session:

016_debugging_guide_v2.md — new section 10 on adapter deployment failure modes
FOCUS_adapter_design.md — Deployment essentials rewritten with the real lessons learned
(Earlier in the session) Tier 1/2/3 response header taxonomy and validator TODO

# Phase 3 scope
Per 013_thunder_adapter_design.md / your file 033_thunder_adapter_design_1_.md, Phase 3 is the provisioning lifecycle:

Thunder Compute API client — Go HTTP client wrapping the Thunder Compute REST API (create instance, list instances, terminate instance, get instance status). Lives at internal/adapters/thunder/thunder_api.go.
provision_instance action handler — pre-check via thunder_provision_check view, generate SSH keypair, store in k8s Secret, call Thunder API to create instance, wait for "running" state, INSERT into thunder_instances, return instance_ip, ssh_user, ssh_key_secret_name, thunder_instance_id, provisioned_at.
decommission_instance action handler — call Thunder API to terminate, delete the k8s SSH Secret, UPDATE thunder_instances.status='decommissioned'.
thunder-reaper scheduled task — runs every 15 min, finds instances older than hard_uptime_hours, dispatches decommission for each.

Want to kick off with Phase 3.1 (the Thunder API client) since it's the foundation everything else depends on? Or do you want to settle the smoke-test KafkaTopic CRD properly first (e.g. apply it permanently rather than as a 5-minute retention)? Either way is reasonable — let me know how you want to start.

Phase 3.3
What 3.3 will do
The provision_instance action handler — wires everything together. Approximate shape:
1. Read incoming request fields:    training_run_id, ?gpu, ?mode, ?vcpus, ?disk_size_gb
2. Pre-check:                        SELECT * FROM thunder_provision_check;
   If can_provision=false → return error with denial_reason
3. Generate keypair:                 kp := ssh.GenerateKeypair("training-run-" + training_run_id)
4. Call Thunder create:              resp := thunderAPI.CreateInstance(ctx, {gpu, mode, vcpus, ..., PublicKey: kp.PublicAuthorizedKey})
5. Store keypair in Secret:          secretMgr.CreateKeypairSecret(ctx, resp.UUID, kp)
6. Poll until RUNNING:               inst := thunderAPI.WaitForRunning(ctx, resp.Identifier, 5*time.Second)
   ctx wrapped with 5-minute timeout from config
7. INSERT thunder_instances row:     thunder_uuid=resp.UUID, identifier=resp.Identifier,
   instance_ip=inst.IP, ssh_user='ubuntu', ssh_key_secret=name,
   training_run_id=...,  status='running', provisioned_at=NOW()
8. Return shaped response:           {instance_ip, ssh_user, ssh_key_secret_name, thunder_instance_id,
   thunder_uuid, provisioned_at}
   Plus error handling for:

Pre-check denial (don't even try to provision)
API auth error (token misconfigured)
API 5xx / rate-limit (transient → return error_recoverable so chassis retries)
WaitForRunning timeout (instance stuck in PENDING → decommission, return error)
Terminal status (ERROR/TERMINATED → return error)
DB INSERT failure after successful provision (compensating action: decommission the orphan)

The compensating-action piece is the most interesting — if we successfully create the Thunder instance but the DB INSERT fails (e.g. transient pg outage), we'd be billed for an instance we can't track. Need a small recovery path that calls DeleteInstance + Secret cleanup before returning the error. I'll structure it as deferred cleanup that only fires on the error path.

# 3.4
Design recap
Inputs (from the calling agent):

provisioning_id (DB row UUID — preferred, the natural handle), OR
thunder_identifier (numeric Thunder API ID — fallback for cases where caller only has that)

Flow:
1. Look up thunder_instances row by id (or by thunder_instance_id if only identifier given)
2. UPDATE status='decommissioning' to mark intent (idempotency anchor)
3. Call Thunder API DeleteInstance — idempotent on 404
4. Delete k8s Secret — idempotent on 404
5. Compute cost: (NOW() - running_since) * hourly_rate_usd
6. UPDATE status='decommissioned', decommissioned_at=NOW(), cost_usd=<computed>
7. Return success
   Idempotency:

Row already decommissioned: return success with the existing row's fields
Row in decommissioning (mid-flight crash recovery): finish the cleanup, then mark decommissioned
Row not found: return error (caller has bad state)
Thunder 404 / Secret 404: treat as success (already gone)

Three files (mirroring 3.3 structure):

internal/adapters/thunder/store/instances.go — DB access (lookup, status update, finalize)
internal/adapters/thunder/decommission_action.go — the action itself
Patch to adapter.go — dispatch decommission_instance


