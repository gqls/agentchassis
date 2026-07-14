# EXTRACTION U24b — docs/_archive/agent_docs/docs024_key_docs_latest/finetuning/
Extracted 2026-07-13 (recovered from a sub-agent of U24 that completed before the parent hit the session limit). Part of U24 (docs/_archive/, ~372 files) — covers the archived copy of the finetuning working tree (docubundle context packs, phase5 SQL/notes, flywheel_docs FOCUS family, eval artefacts).

## Coverage
| file | treatment |
|---|---|
| docs/_archive/agent_docs/docs024_key_docs_latest/finetuning/BUSINESS_PLAN_finetuning_uk(1).md | family-delta (byte-identical to live BUSINESS_PLAN_finetuning_uk.md; no dropped concepts) |
| docs/_archive/.../finetuning/working/docubundle/GUIDE_deploy_from_context_packs.md | full |
| docs/_archive/.../finetuning/working/docubundle/context_packages/thunder-checkpoint-race/001_development_guide(3).md | family-delta (frozen snapshot; live is 001(4)/(5)) |
| docs/_archive/.../thunder-checkpoint-race/002_system_architecture.md | family-delta (frozen snapshot; live is 002(2)-(4)) |
| docs/_archive/.../thunder-checkpoint-race/003_contracts_and_standards.md | family-delta (frozen snapshot; live is 003(6)-(8)) |
| docs/_archive/.../thunder-checkpoint-race/016_debugging_guide_v2_35.md | family-delta (frozen snapshot; live is 016_v2_58_consolidated) |
| docs/_archive/.../thunder-checkpoint-race/CONTEXT_PACK_thunder_checkpoint_race.md | full |
| docs/_archive/.../thunder-checkpoint-race/FOCUS_adapter_design(3).md | family-delta (live counterpart flywheel_docs/FOCUS_adapter_design(3).md + 035_adapter_guide.md) |
| docs/_archive/.../thunder-checkpoint-race/FOCUS_finetuning_flywheel_and_service(25).md | family-delta (vs (21) read + live (25)) |
| docs/_archive/.../thunder-checkpoint-race/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md | full |
| docs/_archive/.../thunder-checkpoint-race/NOTES_phase5_training_launcher_running(35).md | family-delta (vs (39)) |
| docs/_archive/.../thunder-checkpoint-race/PLAN_checkpoint_and_artefact_upload_b2(4).md | full |
| docs/_archive/.../thunder-checkpoint-race/RUNBOOK_phase_b_c_d_deploy(4).md | family-delta (vs (7)) |
| docs/_archive/.../thunder-checkpoint-race/STATUS_thunder_adapter_2026-06_04.md | full |
| docs/_archive/.../finetuning/working/docubundle_old/CONTEXT_PACK_thunder_checkpoint_race(1).md | family-delta (byte-identical to pack CONTEXT_PACK) |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/README.md | full |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/adapter_config.json | skipped-generated |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/manifest.json | skipped-generated |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/special_tokens_map.json | skipped-generated |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/tokenizer.json | skipped-generated |
| docs/_archive/.../finetuning/working/eval/iter0_eval/lora_iter0_full/tokenizer_config.json | skipped-generated |
| docs/_archive/.../finetuning/working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md | family-latest |
| docs/_archive/.../finetuning/working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(1)..(20).md | family-delta (20 versions) |
| docs/_archive/.../finetuning/working/flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md | family-delta (live phase5 copy; superseded per pretrigger §0a) |
| docs/_archive/.../finetuning/working/phase5/106_thunder_unreachable_counter.sql | full |
| docs/_archive/.../finetuning/working/phase5/108_thunder_training_monitor_orchestrator.sql | full |
| docs/_archive/.../finetuning/working/phase5/108_thunder_training_monitor_orchestrator(1).sql | family-latest (byte-identical to 108 base) |
| docs/_archive/.../finetuning/working/phase5/109b_fix_presign_one_loop_item_keypath(1).sql | full |
| docs/_archive/.../finetuning/working/phase5/110_training_launcher_batch_presign(1).sql | full |
| docs/_archive/.../finetuning/working/phase5/NOTES_phase5_training_launcher_running.md (0)..(38) | family-delta (39 versions total) |
| docs/_archive/.../finetuning/working/phase5/NOTES_phase5_training_launcher_running(39).md | family-latest |
| docs/_archive/.../finetuning/working/phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md | family-latest of this archive (unique 06-08/06-09 race+batch deltas captured; live is (7)) |
| docs/_archive/.../finetuning/working/phase5/RUNBOOK_iter0_pretrigger(3).md | full |
| docs/_archive/.../finetuning/working/phase5/RUNBOOK_phase_b_c_d_deploy.md (0)..(6) | family-delta |
| docs/_archive/.../finetuning/working/phase5/RUNBOOK_phase_b_c_d_deploy(7).md | family-latest |

## Concepts

### Internal AI training flywheel (A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §1 table "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (2026-04-23)
- **what:** The core internal loop: the site-building pipeline logs every LLM call as a byproduct; that data periodically fine-tunes local models that are swapped in for Claude calls where quality holds, dropping API cost. Four lanes — A (data export), B (RAG), C (LoRA training), D (Claude-vs-local eval). A and B were done; C scripted; D paused on infra contention.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#1, #2, #4
- **relations:** parents Flywheel A export, Flywheel B RAG, Flywheel C fine-tuning, Flywheel D eval; feeds Phase 5 launcher; three improvement channels
- **verify-later:** llm_call_log, knowledge_base, training_exports schema, ai_endpoint_health

### Flywheel A — training-data export pipeline
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4i "First real training dataset now in Postgres: export_id fef7be6b-…, 1,958 rows, 21.2MB" and "Spawning architecture fully validated" (2026-04-23)
- **what:** A chassis action (`training_data_export`) + `training-data-exporter` specialist agent (wrapped by `training-data-export-orchestrator`) that reads `llm_call_log`, strips markdown code fences via `stripMarkdownFromResponse`, validates JSON, and writes ChatML training rows. Evolved v1 (static file config, superseded) → v2 (reads params.CollectedData["input_data"], file output to /tmp) → v3/v3.1/v3.2 (writes to a dedicated `training_exports` Postgres schema, per-batch transactions to survive pgbouncer).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** produces datasets for Flywheel C; superseded intermediate: v1 file-output export; feeds training_exports schema concept
- **verify-later:** platform/orchestration/actions/ training_data_export_v3.go; training_exports.runs, training_exports.rows; agent_definitions training-data-exporter

### training_exports Postgres schema (versioned dataset snapshots)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4g schema SQL + §2.4i "1,958 rows landed" with export_id `fef7be6b-…`; RUNBOOK bcd(7) §4b uses `training_exports.runs`/`rows` live (2026-06)
- **what:** Two-table schema (`runs` metadata + `rows` ChatML JSONB) chosen over S3 because 21MB–2GB fits Postgres TOAST and avoids a second storage system. A unique index on `(export_id, metadata->>'source_log_id')` blocks duplicate source rows; real-time streaming into it was considered and rejected in favour of named batch snapshots for A/B-comparable training sets. A load-bearing gotcha surfaced later: `runs.rows_exported` can disagree with the real `rows` count (export `a8484922` had rows_exported=1957 but 0 actual rows).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4g; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-2; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4
- **relations:** output of Flywheel A; consumed by training-data-preparer / model-trainer
- **verify-later:** training_exports.runs, training_exports.rows; flywheel_A_v3/001_training_exports_schema.sql

### Flywheel B — RAG knowledge base with nomic task prefixes
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4b "Step 3 (chassis integration) passed on 2026-04-21 … Flywheel B is done"; "Prefix patch deployed and verified live 2026-04-21"
- **what:** `knowledge_base` pgvector(768) table read/written by `rag_lookup`/`rag_index` actions on the cpu-ollama nomic-embed-text endpoint, with trigram fallback. Empirically established that nomic `search_document:`/`search_query:` task prefixes are load-bearing for ranking (French Bulldog BOAS test), now patched into production. Best practice: filter by metadata (vertical, component_type, source) first, then rank by similarity.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.2, #2.4b; #14 (Ollama specifics)
- **relations:** short-term lever paired with LoRA (long-term); the flagship of the finetuning.uk RAG product
- **verify-later:** platform/orchestration/actions/rag_actions.go (applyNomicPrefix); knowledge_base table; PATCH_rag_actions_nomic_prefixes

### Flywheel C — LoRA fine-tuning path (Unsloth QLoRA Llama 3.3 70B)
- **category:** model-infrastructure
- **status-signal:** deployed (first run closed out)
- **status-evidence:** NOTES(39) "Update — 2026-06-05: iter_0 CLOSED OUT … adapter_model.safetensors 828MB … training_run 1cd65dd7 reconciled to complete"; FOCUS(21) §2.5 pipeline scripted
- **what:** The training pipeline: pull dataset from Postgres → Unsloth QLoRA train Llama 3.3 70B Instruct (`unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, 3 epochs, batch 1, grad-accum 8, lr 2e-4, lora_r 16, max_seq 4096) → inference sanity test → LoRA adapter (~150MB). Base 70B chosen because hardware was already available, though 8B was flagged as likely 95% quality at 10% cost. Real run was ~24h (not the scripts' claimed 30-90 min).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150; eval/iter0_eval/lora_iter0_full/README.md (frontmatter)
- **relations:** consumes training_exports datasets; deployed via Phase 5 launcher; produces LoRA iter0 adapter; superseded automation design = Flywheel C Phase 2
- **verify-later:** flywheel_C/02_train_llama_3_3_70b.py, 01_pull_dataset_from_postgres.sh, run.sh

### Flywheel D — Claude vs local-model quality eval (replay harness)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.4d "paused"; partial results table "Case 1 … 27 min … not a practical substrate for production-scale replay-eval"; §2.4d-comparison methodology "run after eval completes"
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting the dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner

### Flywheel C Phase 2 — HTTP-job-server training automation
- **category:** model-infrastructure
- **status-signal:** abandoned
- **status-evidence:** FOCUS(21) §2.5.1 "design locked, not built" (2026-04-23), proposing `model-trainer/model-evaluator/training-flywheel-orchestrator` + `POST /jobs` VM server; superseded in practice by the Kafka/saga Phase 5 chain (NOTES(39) §1) where model-trainer is an orchestrator, not an HTTP-polling agent
- **what:** An abandoned design where a `model-trainer` specialist would POST a dataset to a ~200-line FastAPI-style HTTP job server running on the GPU VM (`POST /jobs`, `GET /jobs/{id}`, download adapter), polling to completion, with three new tables (`model_training_runs`, `model_artefacts`, `model_evaluations`). SSH-remote-exec and a VM Kafka consumer were both explicitly rejected. The actual Phase 5 build instead made the VM credential-free with the chassis driving via thunder-adapter presigned URLs.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5.1; #15 changelog 2026-04-23
- **relations:** superseded by Phase 5 training-launcher + model-trainer saga; the schema names live on as `model_lifecycle.training_runs`
- **verify-later:** model_lifecycle.training_runs; no `/jobs` HTTP server exists in repo

### finetuning.uk self-service RAG product (business strategy)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN(1) §1 "A RAG platform for SMEs"; FOCUS(21) §5 "This is a product decision, not a technical one" and §10 shipping ladder with "Aspirational dates, not promises"
- **what:** The plan to turn finetuning.uk into a paid product. Direction pivoted several times: from "self-service fine-tuning SaaS" → RAG-over-your-docs platform with automatic data curation as the named differentiator; from "concierge first, UI later" → UI-first ("build the cockpit we use ourselves"); target user refined up from non-technical owners to technical-adjacent SME ops leads. Pricing £199–1,499/mo + setup fees; £9-12k/mo solo-operator target; £100k year-1 projection.
- **sources:** BUSINESS_PLAN_finetuning_uk(1).md#1, #5, #7; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#5, #7, #8, #10
- **relations:** reuses internal-flywheel infra (Ollama, Unsloth, rag_index); replacement direction for the abandoned "self-serve fine-tuning" pitch
- **verify-later:** finetuning.uk site_specs; tenant_id on knowledge_base (not yet built)

### Training-data export format (ChatML + metadata sidecar; SFT/DPO negative examples)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22"; iter_0 audit "1,970 total" rows
- **what:** Convention that each training row is a ChatML `messages` array plus an ignored metadata sidecar (source_log_id, agent_type, step_name, orchestration_id, model, export_version). Code fences must be stripped; edge-case "prose instead of JSON" rows are excluded from SFT (they'd teach wrong-shape output) but kept in `llm_call_log` as future DPO "rejected" examples. Schema heterogeneity noted: one (agent_type, step_name) covers hero/minimal-hero/header schemas.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4g (dataset profile)
- **relations:** part of Flywheel A; export_version enables downstream compatibility checks
- **verify-later:** stripMarkdownFromResponse in ai_actions.go; export_version field

### Model swap / revert mechanism
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4/§4.1 "Swap / revert functions deployed" (migration 083)
- **what:** Per-agent per-step functions `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` that safely snapshot an agent's `ai_service` block in `agent_definitions.default_config` before swapping its LLM (e.g. Claude → a fine-tuned local model), with a full-table backup as the nuclear option. `snapshot_agent`/`revert_agent` are also the sanctioned backup path for agent-definition migrations (used by migration 110).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4; phase5/110_training_launcher_batch_presign(1).sql#0a
- **relations:** the deployment step of the flywheel (swap-if-eval-passes); used by migration snapshotting
- **verify-later:** migration 083; agent_definitions_backup table; snapshot_agent/revert_agent functions

### AI endpoint health routing (claude / cpu-ollama / gpu-ollama)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.3 table lists gpu-ollama as "currently DOWN, not always-on"; §4.1 "Endpoint health routing deployed"
- **what:** `ai_endpoint_health` table (migration 085) tracks three inference endpoints — Claude (default high-quality), cpu-ollama (embeddings + mistral-small3.1/nomic), gpu-ollama (Llama 70B, future LoRAs). Healthy endpoint → work flows; unhealthy → items wait (back-to-triage). The kafka-scheduler only probes endpoints listed here, which is why the sibling `ollama-eval` pod stays invisible to production routing.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.3, #2.4d
- **relations:** substrate for Flywheel D eval and model swap; ThunderCompute H100 was the intended gpu-ollama
- **verify-later:** ai_endpoint_health table; ollama-adapter, ollama-gpu, ollama-eval services

### Adapter design pattern (canonical guide)
- **category:** adapters
- **status-signal:** superseded
- **status-evidence:** context-pack FOCUS_adapter_design(3).md is a frozen copy; live docs advance to flywheel_docs/FOCUS_adapter_design(3).md and docs024_key_docs_latest/035_adapter_guide.md (38KB, Jun 2026)
- **what:** The canonical guide for building single-replica Kafka-consuming "adapter" services that wrap one external API and hold its credentials (git, web-scrape, image-generator, ollama, thunder). Covers the Adapter struct, NewAdapter cleanup convention, sequential Run loop, handleMessage dispatch, the three-tier response-header contract (Tier-1 validator-required, Tier-2 chassis-routing e.g. `in_response_to_request_id`, Tier-3 observability), health/shutdown, topic naming conventions A vs B, and deployment essentials (serviceAccountName, imagePullSecrets, `command:` not `args:`, Strimzi topic pre-creation).
- **sources:** docubundle/.../FOCUS_adapter_design(3).md#TL;DR, #Responsibilities, #Sending-responses, #Deployment-essentials
- **relations:** superseded by live 035_adapter_guide.md; instantiated by thunder-adapter; response-header contract underpins the reply-topic bugs
- **verify-later:** internal/adapters/*/adapter.go; platform/validation/Validator; 035_adapter_guide.md

### thunder-adapter — Thunder Compute GPU provisioning adapter
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter_2026-06_04 §1 phases 3.0–3.6; FOCUS(21) §14 "Provision loop verified end-to-end (2026-05-22)"
- **what:** The adapter that wraps the Thunder Compute API to provision/decommission on-demand GPU VMs, holding the Thunder token and B2 keys. Actions: `provision_instance` (spend-check → ed25519 keypair → create → WaitForRunning → INSERT `thunder_instances` with compensating cleanup), `decommission_instance` (idempotent, computes cost from running_since), plus SSH (`ssh_exec`, `ssh_get_status`) and presign actions. Two matcher bugs that blocked it for days: response headers must be a typed struct (not map[string]string) so is_complete/is_error serialise as JSON bools; and `thunder_instance_id` uniqueness must be a partial index on live rows because Thunder recycles numeric ids.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1, #3; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics)
- **relations:** called by gpu-provisioner, training-launcher, thunder-reaper, thunder-training-monitor; credential boundary for presigned URLs
- **verify-later:** internal/adapters/thunder/api/types.go, provision_action.go, decommission_action.go; thunder_instances, thunder_config, thunder_provision_check

### Thunder Compute API specifics (field/casing/template traps)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §14 "Request/response field shape is asymmetric (verified 2026-05-20 via tnr status --json)"
- **what:** Hard-won Thunder API facts: base URL `https://api.thundercompute.com:8443/v1`; CREATE uses snake_case ints (gpu_type, cpu_cores, num_gpus) but STATUS/LIST returns camelCase with numbers as JSON strings and UPPERCASE enums; real templates are `base`/`ollama`/`comfy-ui`/`forge-neo`/`unsloth` (the OpenAPI `ubuntu-22.04` example is rejected); the login user is `ubuntu` not `root`; SSH needs wait-for-sshd (RUNNING ≠ sshd ready); the SSH port from list is unreliable, use `tnr connect --json`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Thunder Compute API specifics; Phase 4 SSH item)
- **relations:** underpins thunder-adapter; `IdentifierInt()`/`IsReadyStatus` handling
- **verify-later:** internal/adapters/thunder/api/types.go (CreateInstanceRequest vs Instance)

### Presigned-URL storage boundary (prepare_object_url / dataset / artefact)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §5a "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049 … signing against personae-model-training" (2026-06-02)
- **what:** The adapter mints presigned B2 URLs but never moves data — only the few-hundred-byte URL travels over Kafka, the actual bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h; DatasetURL/ArtefactURL delegate to it). Bucket is `personae-model-training` (the preparer's `s3_bucket=finetuning` agent_def value is stale/logical); B2 keys live in `personae-storage-secrets`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** enables the checkpoint/artefact upload; contrasted with rejected storage-adapter; verification gotcha: presigned GET fails HEAD (curl -I) with 403
- **verify-later:** internal/adapters/thunder/data_url_actions.go (ObjectURL/handlePrepareObjectURL); storage.Client.GetPresignedPutURL

### Storage-credential architecture decision (no storage-adapter; presign not blob-pipe)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §14 "Storage credential architecture — decision (2026-05-22) … do NOT build a storage-adapter service"
- **what:** A decision that routing multi-MB datasets / hundreds-of-MB artefacts through a storage-adapter over Kafka is a real failure (max.message.bytes ~1MB), so a storage-adapter is only safe for minting URLs, dangerous for moving data. Interim: hardcode the adapter's B2 env. Eventual (deferred, not built): one shared `storage.NewDefaultClient` reading `personae-storage-secrets` everywhere to kill the four-conventions credential mess; untrusted agents get presigned URLs, never a blob pipe.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Storage credential architecture)
- **relations:** justifies the presigned-URL boundary; GetPresignedPutURL added to storage.Client forced rebuild of every binary importing platform/storage
- **verify-later:** platform/storage/interface.go; personae-storage-secrets

### Phase 5 training-launcher + model-trainer orchestration chain
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Launcher green … LAUNCH_PID=216 … COMPLETED → notified parent success"
- **what:** The real `training-launcher` (migration 102, replacing a stub) driven by the `model-trainer` orchestrator, which spawns then calls `training-data-preparer → gpu-provisioner → training-launcher` over Kafka/saga. The launcher presigns dataset+scripts, computes checkpoint keys, presigns them, assembles an upload manifest, SSHes it onto the VM, and launches training detached. Two-level await distinction (child's intermediate adapter calls vs the child→parent final notification) is load-bearing.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#1, #5; RUNBOOK_iter0_pretrigger(3).md#6; flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** replaces Flywheel C Phase 2 HTTP-server design; children call thunder-adapter; superseded predecessor = 2026-05-24 Option A handoff
- **verify-later:** agent_definitions training-launcher (id 1223bdc1), model-trainer (94f5a069); migrations 102/109/110

### setsid detached launch command
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §4 "ssh_exec blocks to command exit (§2), so the launch must return immediately … the SSH channel hits EOF right after echo"; confirmed `LAUNCH_PID=216` (2026-06-09)
- **what:** Because the adapter's `ssh_exec` runs `session.Run` and blocks up to a 5-min timeout for the remote command to exit, the launch command runs the fetch+train chain under `setsid bash -c '…' </dev/null >launch.log 2>&1 &` and echoes `LAUNCH_PID=$!`, so the SSH channel EOFs immediately. An early superseded version used `nohup`; a real bug found later: `write_manifest` (first VM-FS touch) needed `sudo mkdir`/`sudo chown /workspace` (fixed in 109a) because `/` isn't ssh-user-writable.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#4, #update-2026-06-05-deploy-step-2
- **relations:** part of training-launcher; 109a perm fix; run.sh markers
- **verify-later:** thunder_ssh_exec_dispatch.go; ssh_exec_actions.go sshCommandTimeout

### Launcher reply-topic own-vs-parent derivation (Decision D4)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-02 16:12: D4 CONFIRMED live … the adapter's reply went to system.agent.generic.responses — the agent's own ExecutionContext.ResponsesTopic"
- **what:** An intermediate adapter reply must be routed to the agent's own `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`), NOT `__parent_responses_topic__` (which is only for the child→parent final notification). The inherited handoff had this backwards; provision/decommission always used own-topic and worked. The same class of bug bit `dispatch_thunder_ssh_get_status` (cloned from ssh_exec) and was fixed to prefer `execCtx.ResponsesTopic`.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#3 (D4), #6, #10; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** corrects the superseded 2026-05-24 handoff claims; a latent same bug remains in ssh_exec dispatch if fired top-level
- **verify-later:** thunder_prepare_object_url_dispatch.go, thunder_ssh_exec_dispatch.go, thunder_ssh_get_status_dispatch.go; coordinator determineResponsesTopic

### clients_db vs templates_db agent_definitions source-of-truth
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:3x: CORRECTION — chassis reads clients_db, NOT templates_db … templates_db.agent_definitions has the OLD schema (NO version column) … only the 8 original website-builder agents"
- **what:** A multi-session saga establishing that the flywheel-C/rich-schema `agent_definitions` (model-trainer, gpu-provisioner, training-launcher) live and are loaded from `clients_db`, not `templates_db`. Migration 103 first mis-applied to clients_db, then "corrected" to templates_db, then re-corrected: the chassis loader query filters `is_snapshot`/`version` columns that exist only in clients_db's rich schema. The 002 architecture doc's "source of truth is templates_db" refers to the old website-builder catalog.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:31, #update-2026-06-03-17:3x
- **relations:** every Phase 5 migration targets clients_db; contradicts the frozen 002_system_architecture pack copy
- **verify-later:** clients_db.agent_definitions vs templates_db.agent_definitions; migration 103

### Local-step input resolution: input_mapping dead, key_path for loop items (109b)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 109b SQL header "input_mapping did NOT populate input_data.key for this loop substep … deriving the dataset key 40 times"; NOTES(39) "CORRECTS a load-bearing assumption"
- **what:** A load-bearing chassis contract discovered the hard way: the coordinator only resolves `input_mapping` for `call_agent` and loop fan-out, not for plain local action steps. Local actions (and local loop substeps) must read values via a config key holding a dot-path, resolved by `ExtractActionInputs` Strategy 0 / `resolveTemplateToken` / a `key_path`. Migration 109b fixed `presign_one` to read the loop item via `key_path:"ckpt_key"` (from CollectedData where setLoopVariable puts it) instead of the dead `input_mapping{key:ckpt_key}` that had presigned the dataset key 40×.
- **sources:** phase5/109b_fix_presign_one_loop_item_keypath(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#2, #update-2026-06-06-2, #8
- **relations:** cause of the "presigns the dataset key" failure signature; distinct from the await race
- **verify-later:** loop_expansion_handler.go setLoopVariable; content_search.go; datahelpers ExtractActionInputs

### gpu-provisioner output shape flattening (output_fields plural)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:5x: 104 written" — "extractWorkflowResult … reads output_fields — PLURAL only. The gpu-provisioner complete uses output_field (SINGULAR) … falls to the fallback branch"
- **what:** `call_launcher` failed on `provisioning_result.provisioning_id not found` because gpu-provisioner's `complete` step used singular `output_field` (which `extractWorkflowResult` never reads), so its result came out step-name-keyed as `{dispatch_provision, input_data}`. Migration 104 fixed the provisioner's `complete` to plural `output_fields:["dispatch_provision"]` and re-pointed the launcher mapping to `provisioning_result.dispatch_provision.provisioning_id`; a proper chassis fix (honour singular output_field) was vetoed in favour of making the non-compliant agent conform.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:47, #17:4x, #17:5x
- **relations:** launcher input-mapping contract; same singular bug latent in thunder-reaper
- **verify-later:** extractWorkflowResult; agent_definitions gpu-provisioner (0bf9fa8a); migration 104

### 2026-05-24 launcher build handoff (superseded Option A)
- **category:** documentation-system
- **status-signal:** superseded
- **status-evidence:** RUNBOOK_iter0_pretrigger(3) §0a "treat the 2026-05-24 handoff as superseded (it's the source of the prepare_object_url-added and __parent_responses_topic__ claims that turned out wrong)"
- **what:** The handoff documenting the first real training-launcher build ("Option A": a 5-step orchestrator workflow of dispatch-action clones of DispatchThunderDecommissionAction, the run.sh on-VM launch chain, migration 102). It carries two claims later proven wrong — that `prepare_object_url` had been added to the deployed adapter (it hadn't; v1.0.1048 was Phase-4 only) and that replies should route to `__parent_responses_topic__`. The uploaded 102 was also a pre-revision draft (nohup, input_mapping, singular output_field) that must not be re-run.
- **sources:** flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md; phase5/RUNBOOK_iter0_pretrigger(3).md#0a
- **relations:** superseded by the D4 own-topic decision and NOTES(39) §10 handoff-correction log
- **verify-later:** migration 102_training_launcher_real.sql (revised vs draft)

### Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** PLAN(4) "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left"
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Threat model assumes the VM is hostile (holds no B2 key, only single-object URLs); a standing scoped key and per-save callback endpoint were both rejected. Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** makes the monitor's DONE_OK→decommission safe; still-LEFT Phase D3 dispatch_thunder_prepare_resume_url + migration for check_resume; corrects the "list-keys is the ONE adapter gap" claim
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

### Loop-await send-before-register race + preRegisterAwaitedRequest fix
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09: pre-register fix CONFIRMED in prod … Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true … The send-before-register race is closed"
- **what:** The central thunder-checkpoint-race bug: the local dispatch `dispatch_thunder_prepare_object_url` produced the adapter request and returned `await_response:true` BEFORE the coordinator inserted the `awaited_requests` row, so a fast (~1s) reply beat the insert, `ClaimAwaitedRequest` (WHERE status='waiting') found nothing, the reply was dropped, and the timeout re-dispatched every ~3 min forever. `spawn_agent`/`call_agent` avoid it via `preRegisterAwaitedRequest` (register-before-send, `ON CONFLICT (request_id) DO NOTHING`). Fix: call the same helper in the dispatch before `ProduceWithValidation` (guarded `if params.DB != nil`); note the helper's hardcoded 120s timeout_at then pins every presign await.
- **sources:** docubundle/.../HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md; docubundle/.../CONTEXT_PACK_thunder_checkpoint_race.md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-3, #2026-06-08, #2026-06-08-2
- **relations:** fourth cause of the `awaited_requests`-stuck-waiting symptom (016 §9); the batch presign superseded the loop that exposed it
- **verify-later:** thunder_prepare_object_url_dispatch.go; spawn_actions.go preRegisterAwaitedRequest; awaited_requests table

### Batch presign (prepare_object_urls) — O(K²) loop retirement (migration 110)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 110 SQL header + NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Full launcher path completed in ~26s … state Version 30 … The O(K²) class is gone"
- **what:** After the race fix, the K=40 per-checkpoint loop still crawled to a halt by iter_9 (~9 min) because every awaited substep re-persisted the entire expanded ~80-substep workflow + growing collected_data/ProcessingHistory — O(K²). Decision: replace the loop + `flatten_checkpoint_urls` with one batch adapter call `prepare_object_urls` (keys[] → ordered presigned_urls[], reusing `DataURLAction.ObjectURL` per key, no new signing path). Migration 110 swaps `presign_checkpoints` to `dispatch_thunder_prepare_object_urls` and re-points `assemble_manifest.checkpoint_urls → ckpt_presign_batch.presigned_urls`, dropping flatten; workflow completed in ~26s.
- **sources:** phase5/110_training_launcher_batch_presign(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-09, #update-2026-06-09-2, #update-2026-06-09-3
- **relations:** the documented structural fallback that became the chosen path; retires the presign loop
- **verify-later:** data_url_actions.go handlePrepareObjectURLs; thunder_prepare_object_urls_dispatch.go; migration 110

### run.sh RUN_SH markers + set -e durability hard-gate
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** PLAN(4) "run.sh — BUILT 2026-06-05 … set -euo pipefail plus 02_train's final-upload hard-gate … RUN_SH_DONE are only reached on exit 0, which now implies the adapter is in B2"
- **what:** The on-VM launch chain emits grep-able markers (`RUN_SH_START → RUN_SH_STEP setup → RUN_SH_SMOKE_OK → RUN_SH_STEP full_train → RUN_SH_FULL_OK → RUN_SH_DONE`). Because `set -euo pipefail` plus the final-upload raise means DONE only prints on exit 0, `RUN_SH_DONE` came to mean "trained AND uploaded" — the flip that makes the monitor's DONE_OK→decommission safe. `SAVE_STEPS` (cadence) lives in run.sh, default 50 (~1.5h/checkpoint); lowered to 10 for fast tests.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md#run.sh; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4; phase5/NOTES_phase5_training_launcher_running(39).md#8 (Healthy markers)
- **relations:** parsed by thunder-training-monitor probe; gates checkpoint upload path
- **verify-later:** run.sh (bundle at finetuning/scripts/bundle.tar.gz); 02_train --upload-manifest

### thunder-training-monitor + worker (probe/classify/reconcile/decommission)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** NOTES(39) §9 "Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths) … Terminal/decommission branch still never run live … Not enabled"
- **what:** A periodic orchestrator (`thunder-training-monitor`, migration 108) that runs `find_active_training_instances → loop(spawn_worker → call_worker)` every 5 min (scheduled_tasks row, inserted DISABLED, gated pre_query). Each `thunder-training-monitor-worker` (migration 107) probes a box via the adapter's `ssh_get_status`, classifies run.sh markers (ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN) via `classify_training_probe`, reconciles `training_runs` via `mark_training_run_terminal`, and decommissions on terminal verdicts. Deliberately separate from the reaper (different dependencies); closes the running→complete/failed reconcile gap. Enabling it is gated on the upload path proving DONE⟹durable.
- **sources:** phase5/108_thunder_training_monitor_orchestrator.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150, #update-2026-06-04-1x; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** reuses ssh_get_status + dispatch_thunder_decommission; depends on unreachable counter; gated by RUNBOOK step 6
- **verify-later:** agent_definitions thunder-training-monitor (c3b4c052) / -worker (470c6b3f); 5 actions incl find_active_training_instances; scheduled_tasks

### Thunder unreachable-probe counter
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 106 SQL header "counts CONSECUTIVE unreachable probes and only treats the instance as 'lost' … once the count crosses a threshold"
- **what:** Migration 106 adds `consecutive_unreachable_probes` + `last_probe_at` to `thunder_instances` so the monitor can distinguish a transient SSH blip from a truly-lost box. Each scheduler tick is a fresh sub-agent that can't hold count in memory, so the streak lives on the row: the `record_probe_streak` action bumps on unreachable (route to lost/decommission at threshold, default 3) and resets to 0 on any reachable probe.
- **sources:** phase5/106_thunder_unreachable_counter.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150 (Counter step)
- **relations:** part of thunder-training-monitor; keeps the classifier action pure
- **verify-later:** thunder_instances.consecutive_unreachable_probes; record_probe_streak_action.go

### thunder-reaper + cost gate (spend backstop)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** STATUS_thunder_adapter §1 "3.5 thunder-reaper … Deployed and verified end-to-end (2026-05-14)"; NOTES(39) "Reaper SAVE (done, verified)" bumped max_uptime_hours 18→48
- **what:** Two DB-driven cost controls. The `thunder-reaper` scheduled task (every 15 min) decommissions `running` instances older than their per-instance `max_uptime_hours` (default 18; the cap is ours, not Thunder's — computed from `running_since`, extendable per-row without a Thunder call). The provisioning cost gate is the `thunder_provision_check` view (checks 24h spend + estimated_new_run_cost vs `thunder_config.daily_cap_usd`, defaults cap $100 / per-run $25), called before every create; a ~9h/~$18 run needs the estimate/cap made realistic.
- **sources:** docubundle/.../STATUS_thunder_adapter_2026-06_04.md#1; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Spend gating); phase5/RUNBOOK_iter0_pretrigger(3).md#5
- **relations:** shared decommission_instance action with the monitor; distinct from the completion-monitor
- **verify-later:** thunder_provision_check view; thunder_config (daily_cap_usd, estimated_new_run_cost_usd, max_uptime_hours); migration 028

### iter0 pre-trigger + Phase B/C/D deploy runbooks
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_iter0_pretrigger(3) §6 trigger `model-trainer` with export 146a9a12; RUNBOOK bcd(7) "One-line summary of the gates" (2026-06)
- **what:** Two operational runbooks. The iter0 pretrigger runbook lists the gates to reach the first automated training launch (deploy adapter+chassis, upload the scripts bundle, adapter round-trip, cost gate, a gpu-provisioner smoke test of the D4 topic path, then trigger model-trainer). The Phase B/C/D deploy runbook stages the checkpoint-upload rollout: apply 109 → re-pack/re-upload bundle → Tier-2 short launch (B+C integration, SAVE_STEPS low) → resume (blocked on D3+migration) → enable the monitor last. Both hard-code b2 CLI (not aws) and the "verify positive evidence, complete≠succeeded" discipline.
- **sources:** phase5/RUNBOOK_iter0_pretrigger(3).md; phase5/RUNBOOK_phase_b_c_d_deploy(7).md; phase5/NOTES_phase5_training_launcher_running(39).md#7 (Pre-trigger gates)
- **relations:** operationalise the launcher + checkpoint upload + monitor; export a8484922 is the do-not-use trap
- **verify-later:** migrations 109/109a/109b; scheduled_tasks thunder-training-monitor enable step

### Context-pack deploy workflow (docubundle)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** GUIDE_deploy_from_context_packs.md "One general loop, then the deploy mechanics" (per-project quick reference incl. thunder checkpoint race)
- **what:** A meta-workflow for taking a frozen "context pack" into a fresh chat: attach the pack's listed docs+code, pull fresh live context (schema/rows/pods), verify the one decisive fact the pack names before acting (packs restate stale earlier context), do the work under standing rules, deploy via the right mechanism (A chassis image / B DB migration / C work-items / D orchestration trigger / E static sites / F idea.uk binary), and verify positive evidence. The docubundle bundles frozen copies of 001/002/003/016 plus pack-specific docs.
- **sources:** docubundle/GUIDE_deploy_from_context_packs.md; docubundle/.../CONTEXT_PACK_thunder_checkpoint_race.md
- **relations:** frames the thunder-checkpoint-race pack; the frozen 001/002/003/016 copies are reference snapshots
- **verify-later:** docubundle/context_packages/ structure

### LoRA iter0 evaluation adapter (first flywheel output)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** eval README frontmatter `base_model: unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, tags `lora, sft, unsloth, peft`, `PEFT 0.19.1`; NOTES(39) iter_0 adapter_model.safetensors 828MB
- **what:** The actual first-iteration LoRA adapter produced by Flywheel C on page-content-writer data — a PEFT/LoRA adapter over Llama-3.3-70B-Instruct-bnb-4bit trained with Unsloth/TRL SFT, held in `iter0_eval/lora_iter0_full/`. The README.md is an unfilled auto-generated HuggingFace model-card template (all "[More Information Needed]"); the load-bearing content is the YAML frontmatter confirming the base model and training stack. Sits alongside skipped generated tokenizer/config artefacts.
- **sources:** working/eval/iter0_eval/lora_iter0_full/README.md#frontmatter; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-iter_0-closed-out
- **relations:** output of Flywheel C; input to the (paused) Flywheel D eval gate; live counterpart eval tree has iter0_evaluation_report.md
- **verify-later:** iter0_adapter_out/adapter_model.safetensors; flywheel_D eval harness

## Scope-handling notes
The three named version families were read at their highest archived N (FOCUS(21), NOTES(39), RUNBOOK bcd(7)); earlier members were delta-scanned and add nothing beyond what each latest doc's changelog/update-log already records. `BUSINESS_PLAN_finetuning_uk(1).md` is byte-identical to the live copy, and `docubundle_old/CONTEXT_PACK_thunder_checkpoint_race(1).md`, `108(1).sql`, and both live-copy diffs confirmed no unique concepts in those duplicates. The four frozen context-pack copies (001/002/003/016_v2_35) are older snapshots of docs owned by other extractors' scopes; their only finetuning-relevant deltas (the thunder-adapter typed-struct response envelope and the checkpoint-race §9 debugging entry) are already captured under the thunder-adapter and loop-await-race concepts above. This finding set substantially overlaps U06 (docs024_key_docs_latest/finetuning/, the live tree) — consolidation should de-duplicate against U06 rather than re-litigate; where this archive unit's evidence adds a NEW dated fact not in U06, it is retained above.
