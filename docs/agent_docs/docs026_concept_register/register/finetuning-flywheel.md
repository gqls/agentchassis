# Register — finetuning-flywheel

41 concepts, consolidated from 65 raw extractions across units U01, U02, U04,
U06, U12, U17a, U18, U19, U22, U24a, U24b. Heavy duplication as expected: U06
(the live docs024/finetuning tree) and U24b (its archived predecessor) cover
almost identical ground, and the top-level "flywheel programme" framing was
independently captured by five different units. Several entries here overlap
with raw blocks tagged model-infrastructure by the same units (Thunder adapter
internals, training-launcher mechanics, run.sh markers, setsid launch, batch
presign) — those were left in the model-infrastructure register per the
per-category assignment rule and are cross-referenced from here rather than
duplicated in full.

### FTW-001 — Finetuning flywheel four-lane programme (A export, B RAG, C training, D eval)
- **status:** partial
- **status-evidence:** most recent dated evidence (phase-5 notes, 2026-06-09) shows lane C mostly automated end-to-end with the monitor still disabled; lane D was last recorded "paused" (2026-04-23) and never resumed in these docs.
- **what:** The internal self-improvement programme: production LLM calls become training data (Lane A, export), verified knowledge feeds RAG (Lane B), local models get periodically fine-tuned on the exported data (Lane C), and quality is compared against Claude before any swap (Lane D). The strategic goal is dropping API cost by swapping local models in for Claude calls where quality holds. A and B were completed early; C progressed from manual scripts to a mostly-automated Kafka/saga pipeline; D produced one full evaluation cycle (iter_0) then paused.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#1-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#current-state; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md; FOCUS_finetuning_flywheel_and_service(13).md#1,#2,#15-changelog
- **relations:** every other concept in this register; three compounding improvement channels (FTW-002); model swap/revert (model-infrastructure)
- **verify-later:** training_exports schema; model_lifecycle schema; llm_call_log columns; thunder-adapter deployment; scheduled_tasks rows for reaper/monitor

### FTW-002 — Three compounding improvement channels (RAG, LoRA, prompt evolution)
- **status:** partial
- **status-evidence:** RAG deployed and verified live; LoRA iter_0 trained; the `prompt_variant` column exists in llm_call_log but no A/B usage was evidenced in any doc.
- **what:** A framing that recurs across the whole documentation history (from the earliest archived GPU-infrastructure drafts through the live flywheel docs): three independently-valuable channels compound — RAG injects verified knowledge immediately with no training required, LoRA fine-tunes replicate a task cheaply once enough data accrues, and deliberate prompt A/B testing (`prompt_variant`, an 80/20 traffic split promoted on audit-success-rate) evolves the prompts themselves. Good prompts plus good RAG produce the best training data, which produces the best fine-tuned model. The earliest version of this framing also proposed a `training-orchestrator` workflow (export → start_gpu_instance → train → evaluate → deploy_or_reject → stop_gpu_instance) and a scraped-data "AI slop" quality gate.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#3; FOCUS_finetuning_flywheel_and_service(25).md#3; old/older1/020d_gpu_and_model_infrastructure_v4.md#"Quality Improvement Flywheel"; old/older1/020c_gpu_and_model_infrastructure_v3.md#quality-improvement-flywheel
- **relations:** flywheel programme (FTW-001); knowledge_base RAG (FTW-036); llm_call_log prompt_variant column
- **verify-later:** any actual prompt_variant A/B analysis in code or docs; whether a training-orchestrator workflow exists

### FTW-003 — Fine-tuning path (log → export → LoRA → GGUF → Ollama → swap)
- **status:** partial
- **status-evidence:** one training run reached `complete` with real dynamic GPU provisioning via Thunder, but "No agent_definitions row currently points ai_service at llama3.3:70b — trained and tested, never used for production inference" (confirmed independently as late as a 2026-07-10 audit).
- **what:** The concrete technical pipeline underlying the flywheel: accumulate 200+ examples in llm_call_log → export as Alpaca/ChatML → LoRA fine-tune on GPU via Unsloth → GGUF export → load into Ollama → flip the agent definition to `provider: ollama` → A/B against Claude. Candidates are short-output classifiers (site-classifier, vet-practice-verifier, knowledge-extractor, etc.); long-form/judgement agents stay on Claude. The last mile — wiring a trained model into live production inference — is explicitly and repeatedly documented as outstanding, across both the earliest and the most recent sources.
- **sources:** 001_development_guide(5).md#Fine-tuning path; 009_model_infrastructure.md#Future incl. 2026-07-10 note; 023 (manual A/B comparisons); WM/001_development_guide(0).md#fine-tuning-path; docs021.../026_implementation_todo_vertical_architecture(2).md#phase-7
- **relations:** model swap/revert functions (model-infrastructure); Thunder adapter (FTW-040); drain mode; Fine-tuning candidate selection (FTW-023)
- **verify-later:** model_lifecycle.training_runs; agent_definitions ai_service providers

### FTW-004 — llm_call_log as ops visibility + training-data capture
- **status:** deployed
- **status-evidence:** "Verified in production (March 2026) — 57+ rows"; flywheel columns added by a later migration (085); FOCUS(25) confirms the write path is live and populating with 100% of recent rows carrying agent_type.
- **what:** Every `execute_llm_prompt` call is logged fire-and-forget (agent_type, step, model, rendered prompt, response, tokens, latency, success, `__sent_temperature`/`__sent_max_tokens` write-backs) via a `LogLLMCall` goroutine with a 5s timeout that never blocks the workflow. Flywheel columns (work_item_id, prompt_variant, vertical, rag_context_used) link calls to outcomes for LoRA/RAG training exports and per-vertical slicing. Retention is 90 days (success) / 180 days (error) via `cleanup_old_llm_logs`, which exists but is not scheduled — an open task flagged repeatedly across multiple dated sources. Known past bugs: schema/Go column drift (agent_id vs client_id) and historically-empty agent_type rows from buildActionParams.
- **sources:** 001_development_guide(5).md#LLM call logging; 022_ai_endpoint_health_and_flywheel_llm_call_log.sql; FOCUS_finetuning_flywheel_and_service(13).md#2.1,#4.2; FOCUS_finetuning_flywheel_and_service(25).md#2.1,#2.4c,#4.2; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql
- **relations:** training-data export (FTW-005); prompt evolution (FTW-002); ai_endpoint_health (model-infrastructure); Temperature/max_tokens logging gap (llm-call-observability)
- **verify-later:** llm_call_log schema; llm_call_logger.go; scheduling of cleanup_old_llm_logs

### FTW-005 — Training-data export as chassis agent + action (Flywheel A pipeline)
- **status:** deployed
- **status-evidence:** "First real training dataset now in Postgres: export_id fef7be6b…, 1,958 rows, 21.2MB, reconciled manually. Spawning architecture fully validated" (2026-04-23); v3.2 confirmed "deployed as v3.2, verified working".
- **what:** Export of (prompt, response) pairs from llm_call_log as ChatML messages with a metadata sidecar (source_log_id, agent_type, step_name, export_version), fence-stripped via `stripMarkdownFromResponse` and JSON-validated, via a `training_data_export` action plus a `training-data-exporter` worker and a `training-data-export-orchestrator` wrapper. Evolved v1 (template config, failed) → v2 (file output to /tmp, wrong ephemeral pod) → v3/v3.1/v3.2 (writes batched inserts to a dedicated `training_exports` Postgres schema, per-batch transactions to survive pgbouncer). Rejected: real-time streaming into the table — batch snapshots preserve "the dataset we trained on".
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e,#2.4f,#2.4g,#2.4i; FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i; HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md; 106_training_data_exporter.sql
- **relations:** training_exports schema (FTW-006); orchestrator wrapper pattern; pgbouncer per-batch commits; negative-examples/DPO decision (FTW-008)
- **verify-later:** training_exports schema in DB; training_data_export_v3.go and registry entry; agent_definitions rows training-data-exporter/training-data-export-orchestrator

### FTW-006 — training_exports Postgres schema (named snapshot datasets)
- **status:** deployed
- **status-evidence:** HANDOFF 2026-04-23 "Schema (applied, verified)"; two real export_ids (146a9a12, fef7be6b) each hold 1,958 rows.
- **what:** `training_exports.runs` (one row per export: filters, counts, completed_at) plus `training_exports.rows` (ChatML messages + metadata JSONB per training record, unique on (export_id, source_log_id), ordered by row_index, CASCADE delete). Datasets live in Postgres, not files or S3 — named, versioned snapshots referenced by export_id UUID, streamed out to JSONL at training time via `\copy`. Schema deliberately named `training_exports` (not "training_runs") to avoid confusion with the model-training pipeline itself. Known trap: `runs.rows_exported` can disagree with actual row content (one export recorded 1957 but held 0 rows) — always verify with a count before launching training.
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#2.4g; docs/agent_docs/sql_for_tables/039_training_exports.sql; HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#done-verified
- **relations:** training-data export (FTW-005); training-data-preparer (streams export to S3, FTW-027); model_lifecycle schema (FTW-024)
- **verify-later:** training_exports schema in clients_db; the a8484922 zero-row anomaly; recent_runs view

### FTW-007 — ChatML export format with metadata sidecar
- **status:** deployed
- **status-evidence:** "Format: ChatML messages with metadata sidecar. Decided 2026-04-22."
- **what:** Training rows are `{messages:[{role:user},{role:assistant}], metadata:{source_log_id, agent_type, step_name, orchestration_id, model, created_at, export_version}}`. Chosen for chat-tuned base-model parity, trainer-tool defaults (Unsloth/Axolotl), and `/api/chat` training-inference parity. Metadata gives row-level traceability back to llm_call_log; export_version future-proofs format evolution. The whole prompt_rendered currently goes into a single user turn — no system/user split yet.
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#2.4e; flywheel_A_export_page_content_writer_iter0.sql (header)
- **relations:** response cleaning (FTW-008); training_exports schema (FTW-006)
- **verify-later:** export SQL/action output shape vs current llm_call_log columns

### FTW-008 — Response cleaning and SFT negative-example exclusion (DPO reserve)
- **status:** deployed
- **status-evidence:** iter_0 dataset audit table (97.4% clean JSON, 2.6% fenced, fences stripped on export); "For our first training run: plain SFT, edge cases excluded" (decision 2026-04-22).
- **what:** Exports strip markdown code fences (else the fine-tune learns to emit fences, exactly what prompts forbid) and exclude edge-case prose responses — cases where Claude correctly produced prose instead of the expected JSON shape. Plain SFT has no "don't do this" signal, so these are positive examples of the wrong output shape; they stay in llm_call_log rather than training_exports, reserved as the future "rejected" half of DPO preference pairs.
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#2.4e; FOCUS_finetuning_flywheel_and_service(13).md#2.4e "Negative examples / edge cases"
- **relations:** ChatML export format (FTW-007); `<no value>` contamination (FTW-009); training-data export filters (strict_json)
- **verify-later:** stripMarkdownFromResponse; export action's strict_json/prose-exclusion behaviour; whether any DPO work exists later

### FTW-009 — `<no value>` training-data contamination and the iter_1 filter floor
- **status:** partial
- **status-evidence:** HANDOFF 2026-05-07 known issues: "Training rows from before that fix include the literal token `<no value>`. iter_1's export should filter created_at >= <fix_date>"; still listed as needed as of HANDOFF 2026-05-08.
- **what:** A prompt-builder rendering bug injected the literal token `<no value>` into production prompts; iter_0's training data (and its eval cases) inherit it. The fix-deploy date becomes the created_at filter floor for the iter_1 export, but as of the last docs the date had not been recorded — an open data-hygiene debt that would silently contaminate any next iteration.
- **sources:** HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#known-issues; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#revised-iter_1-priorities
- **relations:** training_exports (FTW-006); held-out eval set (FTW-020, same artefact affected)
- **verify-later:** git log for the prompt-builder fix; whether any iter_1 export exists with the filter applied

### FTW-010 — Dataset profile and schema heterogeneity of page-content-writer iter_0
- **status:** deployed
- **status-evidence:** dataset profile tables (n=1,957; p50 prompt 8,250 chars; three dominant JSON schemas: hero-with-CTAs 68%, minimal hero 18%, header/nav 9%).
- **what:** One (agent_type, step_name) training slice actually spans three component output schemas, so the model must learn schema selection conditioned on the "Component: X" text in the prompt. A first-pass option to filter to the top-2 schemas (86% of rows) was noted but the full set was trained anyway. Prompt/response size distribution anchors the max_seq choice (some prompts approach 4,000 tokens → max_seq 4096).
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#2.4g; terminology.md#seq-4096
- **relations:** Base-model decision (FTW-012); inference-test success criteria (keys match trained schemas)
- **verify-later:** training_exports.rows key distribution query

### FTW-011 — Flywheel C training pipeline (Unsloth QLoRA scripts 00-03, smoke-gates-full)
- **status:** deployed
- **status-evidence:** "Pipeline shape now concrete, scripts written, awaiting first training run" (2026-04-23); later HANDOFF 2026-05-07 confirms it was "executed end-to-end on a Thunder Compute A100 80GB … iter_0 adapter (791MB) … 5/5 valid JSON" — resolving the earlier aspirational status.
- **what:** Four scripts define the manual training path: `00_vm_setup.sh` (idempotent pinned environment), `01_pull_dataset_from_postgres.sh`, `02_train_llama_3_3_70b.py` (Unsloth QLoRA, CLI-configurable, emits a manifest), `03_inference_test.py` (JSON validity + schema-key sanity check). Discipline: a 20-row/1-epoch smoke train and smoke inference always gate the full run — cheap insurance on unattended multi-hour runs. The same bundle later becomes the automated launcher's payload (FTW-031).
- **sources:** working/flywheel_docs/README.md; HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#flywheel-c; HANDOFF_2026-05-07_flywheel_C_phase1_complete.md; FOCUS_finetuning_flywheel_and_service(13).md#2.5; working/scripts/run.sh (header)
- **relations:** iter_0 baseline run (FTW-013); scripts bundle as deployment unit (FTW-031); run.sh markers (FTW-030); Base-model decision (FTW-012)
- **verify-later:** working/scripts/* against the bundle actually in B2 (finetuning/scripts/bundle.tar.gz)

### FTW-012 — Base-model decision: Llama 3.3 70B Instruct QLoRA (8B ablation planned)
- **status:** deployed
- **status-evidence:** "Base model: Llama 3.3 70B Instruct. Decision taken 2026-04-23"; used for the actual iter_0 run.
- **what:** `unsloth/Llama-3.3-70B-Instruct-bnb-4bit` via Unsloth QLoRA on a single A100/H100 80GB; defaults 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096. 70B was chosen because hardware was already available and a strong baseline is useful — with an explicit acknowledgment that 8B likely delivers ~95% of quality at ~10% of inference cost for this narrow structured-JSON task. A same-dataset 8B comparison run was planned but never executed in any of these docs.
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#2.5; HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#base-model-decision
- **relations:** iter_0 baseline run (FTW-013); cost anchors; epochs ablation
- **verify-later:** 02_train defaults; any 8B run in model_lifecycle.training_runs

### FTW-013 — iter_0 baseline training run (real cost/time/loss anchors)
- **status:** deployed
- **status-evidence:** run summary: 33,136s ≈ 9h12m, final_loss 0.2669 (trailing), peak VRAM 44.8GB, ~$20 total; the resulting adapter's own model-card frontmatter (`base_model: unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, tags lora/sft/unsloth/peft, PEFT 0.19.1) confirms the training stack.
- **what:** The first real fine-tune: 1,958 rows → 1,934 effective, clean loss curve (ep1 1.49→0.27, ep3 →0.10), adapter 791-828MB fp32 safetensors, held in `iter0_eval/lora_iter0_full/`. Epoch-3 loss gap suggests memorisation → a 2-epoch ablation was queued for iter_1 (never run in these docs). Cost anchor $20/iteration (training) + $1.50 (eval) ≈ $22/cycle. Later automated runs corrected the wall-time estimate: the full run is ~24h at ~119s/step without FA2, not the "30-90 min" the README originally claimed.
- **sources:** HANDOFF_2026-05-07_flywheel_C_phase1_complete.md; working/phase2/README.md; NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150; working/eval/iter0_eval/lora_iter0_full/README.md#frontmatter; NOTES(39)#update-2026-06-05-iter_0-closed-out
- **relations:** GPU version pinning (FTW-014); snapshot economics (FTW-015); GPU training performance model (FTW-016); fp16 save decision (FTW-017); iter_0 verdict (FTW-022)
- **verify-later:** lora_iter0_full/manifest.json; model_lifecycle.training_runs rows 1cd65dd7/e6ab9fad; iter0_adapter_out/adapter_model.safetensors

### FTW-014 — GPU environment version pinning (cu124 stack)
- **status:** deployed
- **status-evidence:** "Version pin discoveries (essential for any future run)" table with per-pin rationale, HANDOFF 2026-05-07.
- **what:** The working training environment is a narrow pin set: torch 2.6.0+cu124, transformers<5, torchao<0.17 (transformers imports torchao eagerly; an incompatible torchao breaks import entirely), a prebuilt flash-attn wheel (Thunder's Ollama template ships CUDA runtime but no nvcc), unsloth+unsloth_zoo both explicitly installed (a git-only install misses the zoo), hf_transfer as a separate package. cu124 is flagged a dead end — next rebuild should move to cu126/cu128. The separate Unsloth template used for eval differs materially: nvcc present, torch 2.10/cu128, xformers pre-installed, FA2 absent.
- **sources:** HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#version-pin-discoveries; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons; working/eval/iter0_eval/001_README.md
- **relations:** 00_vm_setup.sh as canonical environment (FTW-011); snapshot economics (FTW-015)
- **verify-later:** working/scripts/00_vm_setup.sh pin lines

### FTW-015 — Snapshot economics: setup script beats VM snapshots
- **status:** convention
- **status-evidence:** "Created `unsloth-trainer-base-01` then deleted it… Break-even: ~18 training runs/month. Reality: 1-4."
- **stage2-verified (2026-07-14):** deployed → convention — Economic decision (no snapshots); verify-later explicitly says 'none (economic decision)' — not a code claim
- **what:** Thunder snapshots bill the full provisioned 100GB (~$15/month) regardless of used bytes, saving only ~25min/$0.85 per cold start — uneconomic below ~18 runs/month. Decision: no snapshots; the version-pinned idempotent `00_vm_setup.sh` is the canonical, version-controlled "snapshot" instead. Phase-2 automation therefore provisions a fresh instance every run.
- **sources:** HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#snapshot-decision
- **relations:** version pinning (FTW-014); phase-2 automation architecture (FTW-026)
- **verify-later:** none (economic decision); revisit if run tempo exceeds ~15/month

### FTW-016 — GPU training performance model (smoke ≠ steady state; FA2; seq-length cost)
- **status:** convention
- **status-evidence:** "The smoke rate (116 s/step) predicted this — nobody extrapolated it" (2026-06-04).
- **stage2-verified (2026-07-14):** deployed → convention — Mental/performance model narrative, not a built artifact; verify-later asks 'whether cap-sizing-from-smoke was ever built' (unresolved) — reclassified as an idea/observation, not deployed code
- **what:** A captured mental model of training performance: smoke-test speed is unrepresentative (one-time kernel autotune/CUDA-graph costs amortised over too few steps); steady-state emerges after 5-20 steps; FA2 vs xformers/SDPA is a 2-4x attention-speed lever; attention scales O(N²), so max_seq 4096 quadruples 2048's attention work relative to linear cost elsewhere. Operationally: extrapolate full-run wall time from smoke s/step — the 18h uptime-cap overrun (FTW documented under thunder-reaper in model-infrastructure) happened because nobody did this the first time.
- **sources:** working/flywheel_docs/terminology.md; NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** iter_0 baseline run (FTW-013); thunder-reaper per-instance uptime deadline (model-infrastructure)
- **verify-later:** whether cap-sizing-from-smoke was ever built

### FTW-017 — fp16 adapter save decision
- **status:** aspirational
- **status-evidence:** "Save adapters as fp16, not fp32, in iter_1. One-line script change." (iter_0 shipped fp32 at 791-828MB.)
- **what:** PEFT's `save_pretrained()` defaults LoRA weights to fp32 even when training in bf16, doubling adapter size and transfer time (17min `tnr scp` for 791MB). The one-line fix (cast trainable params to fp16 pre-save) was agreed for iter_1, but iter_1 never ran in these docs.
- **sources:** HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#lessons(10); HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** adapter transport via S3/B2 (FTW-032); model_lifecycle.artefacts format field (FTW-024)
- **verify-later:** current 02_train script save path

### FTW-018 — Flywheel D replay-eval methodology + CPU-eval-pod evolution
- **status:** partial (paused)
- **status-evidence:** "Flywheel D (eval) paused"; the CPU attempt is explicitly superseded — "mistral-small3.1 on a shared cpu-ollama adapter is not a practical substrate … 20 cases × 25-30 min = 10+ hours" versus the eventual Thunder-GPU run at ~22s/case (~$0.50).
- **what:** Evaluation replays stored production prompts from llm_call_log against a candidate model instead of re-invoking agents — no orchestration-state pollution, directly comparable to the stored Claude output. Test sets use `DISTINCT ON (orchestration_id)` for diversity ("diverse 20 > random 20"), exported as NDJSON, with fail-fast on empty responses. The first attempt replayed against mistral-small3.1 on the shared CPU Ollama adapter; production contention drove one case to 27 minutes, so a dedicated `ollama-eval` deployment (own PVC/service, 24Gi/28Gi memory, invisible to production routing since kafka-scheduler only probes ai_endpoint_health entries) was created — then the whole CPU-eval path was itself superseded by evaluating the trained adapter directly on Thunder GPU instances.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4d,#2.4d-comparison; FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md; flywheel_D_target_selection.sql (header)
- **relations:** three-level eval pipeline (FTW-019); held-out eval set (FTW-020); Ollama CPU adapter ops (model-infrastructure); ai_endpoint_health (model-infrastructure)
- **verify-later:** deployments/kustomize/services/ollama-eval/; results.jsonl outcome; whether eval ever resumed after pausing

### FTW-019 — Three-level evaluation pipeline (L1 structural / L2 judge / L3 spot-check)
- **status:** deployed
- **status-evidence:** run instructions costed at ~$1/~5min total; iter0_evaluation_report.md generated 2026-05-08.
- **what:** A reusable eval stack: L1 structural metrics computed locally, side-by-side for both models (JSON validity, schema-key match, length ratios, forbidden phrases from the brief's avoid-list, fabrication-marker regexes); L2 Claude-as-judge scoring relevance/voice/integrity 1-5 plus a winner call, with anonymised randomised A/B and resume support; L3 auto-selected spot-check cases folded into a markdown report by `build_report.py`. The report deliberately surfaces confounds and makes no ship/no-ship call itself. Known limit: L1 fabrication regexes have poor recall — contextual fabrications need L2/L3 to catch.
- **sources:** working/flywheel_docs/002_README-flywheele_D_evaluation_pipeline.md; working/eval/iter0_eval/iter0_evaluation_report.md#methodology; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(6)
- **relations:** model_lifecycle.evaluations JSONB contract (FTW-024); Claude-as-judge (FTW-021); eval gate (FTW-025)
- **verify-later:** working/eval/iter0_eval/05-07 scripts; any evaluations rows in model_lifecycle

### FTW-020 — Held-out eval set v1 as the canonical cross-iteration comparison set
- **status:** deployed
- **status-evidence:** "held_out_cases_v1.jsonl is the canonical eval set across iterations — same 20 cases … so trends are meaningful."
- **what:** 50 cases pulled from llm_call_log after the training-export cutoff (created_at > 2026-04-23 14:54:32Z), one per orchestration, defensively excluded from the training set by source_log_id; 20 used for iter_0, 30 reserved for later. The selection SQL is kept for reproducibility. Iterations evaluate against the same cases so deltas read as trend, not noise; fresh `_v2` sets are the intended mechanism for novelty checks.
- **sources:** working/eval/v1/held_out_cases_v1.sql; iter0_evaluation_report.md#sample-selection; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** replay eval (FTW-018); three-level pipeline (FTW-019); `<no value>` contamination (FTW-009, same source artefact)
- **verify-later:** held_out_cases_v1.jsonl vs training export overlap

### FTW-021 — Claude-as-judge with anonymised A/B and self-recognition bias handling
- **status:** deployed
- **status-evidence:** "5 cases had identical R/V/I scores… 5/5 went to Claude… consistent with residual self-recognition bias"; "claude-opus-4-7 is the canonical judge model."
- **what:** Judge design: anonymise responses, randomise A/B positions, score dimensions before picking a winner, and use a different Claude model (Opus) than the training-label producer (Sonnet 4.6) to reduce — not eliminate — self-recognition. The bias was empirically observed: rubric-tied cases broke for Claude every time, so headline win-rates get an adjusted reading (16-4 raw → 12-4 with 4 judge-preference ties). Position bias was checked explicitly (A won 55%, no clear bias found).
- **sources:** working/eval/iter0_eval/iter0_evaluation_report.md#level-2; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(5)
- **relations:** three-level pipeline (FTW-019); model_lifecycle.evaluations judge_model column (judge-drift tracking)
- **verify-later:** level2.py anonymisation logic

### FTW-022 — iter_0 verdict: shippable for low-stakes; voice fidelity is the iter_1 lever
- **status:** convention
- **status-evidence:** "iter_0 is shippable for low-stakes use… Not for client-facing where Δ-0.20 on voice would be visible"; "Add improve voice fidelity" to iter_1 priorities.
- **stage2-verified (2026-07-14):** deployed → convention — This is an evaluative verdict/judgment call, not a built artifact — no code/infra claim to verify
- **what:** The evaluated position on the first adapter: iter_0 matches Claude on JSON validity (20/20 vs 19/20) and schema, comparable length, with only tiny dimension gaps (relevance -0.25, voice -0.20, integrity -0.10) and 4 substantive wins. Verdict: usable for internal tooling and low-stakes sites; voice is the largest gap and the main iter_1 lever (more epochs? lora_r 32? stricter voice-compliant training rows). "Address verbosity" was explicitly dropped from iter_1 priorities (data showed no gap). Fabrication was judged a both-models problem to solve with prompt-time guardrails or post-hoc verification, not adapter training.
- **sources:** HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md; iter0_evaluation_report.md#tldr; working/eval/001_test_comparison_with_claude.txt
- **relations:** eval gate (FTW-025); Fine-tuning candidate selection (FTW-023)
- **verify-later:** whether any deployment_decision row exists; whether iter_0 was ever served in production

### FTW-023 — Fine-tuning candidate selection/prioritisation
- **status:** aspirational
- **status-evidence:** priority table (2026-04) ranks candidates by structured criteria; a reusable SQL discovery query captures the same logic over llm_call_log volume/recency.
- **what:** Agents with high-volume, structured-JSON, short outputs are swap candidates for local models (knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher); long-form creative output and judgement tasks stay on Claude (page-content-writer's long-form work, visual-design-auditor, chief-strategist are explicit non-candidates). Drives which agent/step gets exported and trained first — though the actual first target, page-content-writer's hero step, was chosen for Flywheel D eval precisely because it had the most logged data, despite being flagged a poor swap candidate on paper.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.6; FOCUS_finetuning_flywheel_and_service(25).md#2.6; flywheel_D_target_selection.sql
- **relations:** Fine-tuning path (FTW-003); per-vertical training (vertical column in llm_call_log)
- **verify-later:** actual llm_call_log volumes per agent

### FTW-024 — model_lifecycle schema (training_runs / artefacts / evaluations / deployable_adapters)
- **status:** deployed
- **status-evidence:** full DDL with comments; confirmed live with CHECK status pending/running/complete/failed and real rows as of 2026-06-09.
- **what:** The run-lifecycle namespace: `training_runs` (one row per QLoRA run, FK to training_exports.runs, JSONB hyperparameters for reproducibility, loss/VRAM/cost outcome metrics, thunder_instance_id breadcrumb), `artefacts` (adapter binaries decoupled from runs to allow requantisation, storage_uri + sha256 + format), `evaluations` (per artefact × eval_set × judge, JSONB l1/l2 metrics, free-text human deployment_decision), plus a `deployable_adapters` view (latest shipped_% adapter per base model — the chassis's intended read point for "which adapter to load") and a `latest_training_run_per_export` view. Supersedes an earlier flat `model_training_runs` sketch from the Phase-2 design.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#tables-needed; NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7
- **relations:** eval gate (FTW-025); thunder_instances FK; model-trainer chain (FTW-027)
- **verify-later:** schema in clients_db; whether deployable_adapters is read by any chassis code

### FTW-025 — Eval gate before promotion (human deployment decision; integrity lives here)
- **status:** partial
- **status-evidence:** "deployment_decision is set by human after review; nullable until then"; "No upload scheme substitutes for evaluating the adapter"; "Auto-deployment NOT included in v1."
- **what:** Adapters never auto-promote: a human reviews Flywheel D output and writes a free-text deployment_decision ('shipped_internal', 'rejected_voice_gap', …); anything `shipped_%` becomes deployable. Critically, the eval gate also serves as the *integrity* boundary for the hostile-VM upload design (FTW-032) — a maliciously-crafted-but-valid adapter written through a legitimate presigned URL is caught by evaluation, not by credentials. The original Phase-2 sketch had a conditional auto-swap (`swap_agent_model if score ≥ threshold`) that was walked back to mandatory human review.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql#evaluations; PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#agents-needed
- **relations:** hostile-VM threat model (FTW-032); model swap/revert (model-infrastructure); deployable_adapters view (FTW-024)
- **verify-later:** whether any evaluation row has a decision; whether swap_agent_model was ever wired to eval output

### FTW-026 — Flywheel C phase-2 automation architecture (evolution)
- **status:** superseded (by its final generation, which is deployed)
- **status-evidence:** three successive design docs each supersede the last; the built system (Phase 5) uses thunder-adapter dispatch actions + presigned URLs + detached run.sh, confirmed live by 2026-06-09 notes.
- **what:** The automation design went through three generations, with "chassis drives, GPU VM serves" as the invariant across all three: (1) a VM-side HTTP job server (POST /jobs, bearer auth, systemd, TLS) — designed with three proposed new agents (model-trainer, model-evaluator, training-flywheel-orchestrator) and three new tables, never built; (2) direct SSH-exec from the chassis (simpler at low run frequency, no VM service to maintain) — the pivot decision; (3) the final built shape: chassis dispatch actions publish to thunder-adapter (provision/ssh_exec/presign), data moves only via presigned B2 URLs, training runs detached under run.sh with a separate monitor. A Kafka-consumer-on-the-VM alternative was rejected throughout (connectivity + overkill).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5.1; FOCUS_finetuning_flywheel_and_service(25).md#2.5.1; HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#architecture-decision; HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** model-trainer chain (FTW-027); training-launcher (FTW-028); presigned data plane (FTW-032)
- **verify-later:** no HTTP job server should exist anywhere; adapter dispatch actions in registry.go

### FTW-027 — model-trainer orchestration chain / Phase 5 training kickoff
- **status:** deployed
- **status-evidence:** "model-trainer flow confirmed live: spawn_data_preparer → spawn_provisioner → spawn_launcher → call_data_preparer → call_provisioner → call_launcher → complete" (2026-06-06).
- **what:** The end-to-end automated training run is the `model-trainer` orchestrator: it spawns three worker agents up front, then calls them in order. `training-data-preparer` streams the export to S3 as JSONL and INSERTs the pending training_runs row; `gpu-provisioner` provisions the A100 through thunder-adapter, storing the SSH key as a k8s Secret; `training-launcher` presigns, writes the manifest, and SSH-launches training (returning a pid immediately — completion is handled separately by the training monitor, never held open in an orchestration for ~9 hours). Full hyperparameter set captured for reproducibility. Known open bug: a failed call_agent step falls through to the next call instead of aborting the orchestration, producing confusing secondary errors.
- **sources:** NOTES_phase5_training_launcher_running(45).md#1,#update-2026-06-06; HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#where-this-sits; 109_model_trainer_orchestrator.sql; 110_training_data_preparer.sql; 111_gpu_provisioner_thunder.sql; 112_training_launcher.sql
- **relations:** training-launcher workflow (FTW-028); migrations 103/104/109-112; call_agent fall-through bug; Thunder instance lifecycle (FTW-038)
- **verify-later:** agent_definitions 94f5a069/71ab9361/1223bdc1 in clients_db; the fall-through behaviour in coordinator code; real vs stub gpu-provisioner/training-launcher implementations

### FTW-028 — training-launcher real workflow (presign → manifest → detached SSH launch → mark_running)
- **status:** deployed
- **status-evidence:** "Full launcher path completed in ~26s… presign_dataset → presign_scripts → compute_keys → presign_checkpoints (ONE batch await) → presign_final → [check_resume] → assemble_manifest → write_manifest → ssh_exec_launch → mark_running → complete" (2026-06-09).
- **what:** The launcher replaced an earlier stub with a workflow of dispatch actions cloned from the proven decommission pattern: presign dataset + scripts bundle (GET), compute K checkpoint keys, batch-presign checkpoint PUTs + final PUT, optionally resolve a resume checkpoint, assemble and SSH-place `/workspace/upload_manifest.json`, then launch training detached and flip training_runs pending→running (a hardcoded guarded transition). Constants live in step config; cross-step values resolve via config dot-paths; the SSH command is built from a `command_template` with `{token}` interpolation. Evolved through a chain of migrations from the initial stub through the batch-presign fix.
- **sources:** HANDOFF_2026-05-24_phase5_launcher_build.md; NOTES_phase5_training_launcher_running(45).md#5,#update-2026-06-09-3; 102_training_launcher_real.sql (header)
- **relations:** batch presign / checkpoint durability (FTW-032); setsid launch (FTW-029); model-trainer chain (FTW-027)
- **verify-later:** live training-launcher default_config; registry entries for the dispatch actions

### FTW-029 — setsid detached launch and the detached exit-0 false-success gap
- **status:** deployed
- **status-evidence:** "exit_code 0 only because the command's last token is echo (the known detached-ssh_exec false-success)" (2026-06-03).
- **what:** The adapter's ssh_exec blocks until the remote command exits (5-min timeout), so the launch chain (curl bundle + dataset via presigned URLs, untar, run under nohup) runs under `setsid bash -c '…' </dev/null >launch.log 2>&1 & echo LAUNCH_PID=$!` — the SSH session returns immediately with the PID. The cost: exit_code 0 only proves the echo ran; VM-side failures inside the detached chain (e.g. a `/workspace` permission failure hit in practice) are invisible to the launcher. Corollary lessons: the command_template must stand up its own workspace with sudo mkdir+chown, and any best-effort VM step under `set -e` becomes fatal (a root-owned `~/.bashrc` append once killed a run at the last cosmetic setup step).
- **sources:** NOTES_phase5_training_launcher_running(45).md#4,#update-2026-06-03; 105_launcher_workspace_sudo_mkdir.sql (header); 109a_fix_write_manifest_workspace_perm.sql (header)
- **relations:** run.sh markers as the real success signal (FTW-030); training monitor (fills the observation gap, model-infrastructure)
- **verify-later:** current command_template in the live launcher definition

### FTW-030 — run.sh launch chain and RUN_SH_* marker protocol
- **status:** deployed
- **status-evidence:** run.sh header ("Emits grep-able RUN_SH_* markers for the future monitor"); verified live 2026-06-09: "RUN_SH_START → STEP setup → STEP smoke → SMOKE_OK → STEP full_train → RUN_SH_UPLOAD manifest=present."
- **what:** All heavy on-VM work lives in run.sh (setup → smoke train → full train), not the chassis workflow, so the chain is editable by re-uploading the bundle with no DB migration. It emits a marker protocol to `/workspace/train.log` (`RUN_SH_START/STEP/SMOKE_OK/FULL_OK/DONE/FATAL`) that is the machine-readable contract for the training monitor's probe (model-infrastructure). After the checkpoint-durability work landed, `set -euo pipefail` plus a hard-gated final upload means `RUN_SH_DONE` implies "trained AND adapter durable in B2." A mid-train crash leaves no marker at all (probed as GONE_UNKNOWN).
- **sources:** working/scripts/run.sh (header); PLAN_checkpoint_and_artefact_upload_b2(7).md#run.sh; NOTES_phase5_training_launcher_running(45).md#8
- **relations:** monitor probe classification (model-infrastructure); scripts bundle deployment unit (FTW-031); CheckpointUploader (FTW-033)
- **verify-later:** run.sh in the live B2 bundle vs the repo copy

### FTW-031 — Scripts bundle in B2 as the training deployment unit
- **status:** deployed
- **status-evidence:** "re-uploading the object IS the whole deploy" (2026-06-03); flat-bundle verification steps confirmed in a later runbook.
- **what:** The on-VM scripts (run.sh, 00_vm_setup.sh, 02_train, 03_inference_test) ship as `finetuning/scripts/bundle.tar.gz` in the personae-model-training bucket; the launcher presigns a GET and the VM curls+untars it. The bundle must be flat (files at archive root). Re-uploading the object deploys new training code — no chassis or DB change required — with the corollary that editing a script without re-tarring deploys nothing (a byte-identical md5 trap). The agent definition holds only the object key.
- **sources:** NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x; working/phase5/UPLOAD_bundle.sh; RUNBOOK_iter0_pretrigger(8).md#4a; working/scripts/README_setup.md
- **relations:** run.sh (FTW-030); presigned data plane (FTW-032); Flywheel C training pipeline (FTW-011)
- **verify-later:** b2 ls of finetuning/scripts/; bundle contents vs working/scripts/

### FTW-032 — Checkpoint & final-adapter durability via pre-minted presigned PUT manifest
- **status:** partial
- **status-evidence:** "Phases A, B, C BUILT… ckpt-0 confirmed in B2 on a real run (the upload path is proven end-to-end)"; "Still unproven in prod is one run reaching RUN_SH_DONE with the final adapter.tar.gz durable" as of the latest dated status.
- **what:** The Thunder VM disk is ephemeral, and originally nothing moved training output off it (no checkpoints, adapter saved only locally — so a reap meant total loss, and the monitor's DONE_OK→decommission path would have destroyed the artefact). The fix: the launcher pre-mints single-object, write-only presigned PUT URLs (K checkpoints + 1 final, keyed by save-INDEX not Trainer global_step) into `/workspace/upload_manifest.json`; the VM uploads through them, with B2 versioning as a write-once backstop and URL expiry set beyond max_uptime. The hostile-VM threat model rejects standing scoped keys and callback endpoints — nothing on the box can mint or write beyond the fixed URL set; this protects access, not artefact integrity (that's the eval gate's job, FTW-025). A second, related finding from the same body of work: the K-iteration per-checkpoint presign loop crawled to a halt by iteration 9 (~9 min) because every awaited substep re-persisted the entire expanded workflow plus growing collected_data — an O(K²) cost. This was retired in favour of one batch adapter call, `prepare_object_urls` (keys[] → ordered presigned_urls[]), cutting the full launcher path to ~26s; a related send-before-register await race was fixed by `preRegisterAwaitedRequest`.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md; idea.uk/nginx/PLAN_checkpoint_and_artefact_upload_b2(1).md; idea.uk/nginx/isolation_test_phase_a.py (header); working/phase5/README_what_is_manifest.md; NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-5,#update-2026-06-09,#update-2026-06-09-2,#update-2026-06-09-3; 110_training_launcher_batch_presign(1).sql
- **relations:** CheckpointUploader (FTW-033); resume path (FTW-034); monitor enablement gate (FTW-035); eval gate (FTW-025); training-launcher workflow (FTW-028)
- **verify-later:** b2 contents under finetuning/checkpoints/ and artefacts/; migrations 109/110/111 state; data_url_actions.go handlePrepareObjectURLs

### FTW-033 — CheckpointUploader trainer callback (best-effort checkpoints, hard-gated final upload)
- **status:** deployed
- **status-evidence:** "Tier 1 (box-free) PASSED"; confirmed built 2026-06-05.
- **what:** `02_train` gained gated flags (`--save-steps`, `--save-total-limit`, `--upload-manifest`; defaults keep old behaviour byte-for-byte). A `CheckpointUploader(TrainerCallback).on_save` tars each checkpoint and PUTs it to its save-index URL synchronously (a background thread was rejected — it would race save_total_limit's directory deletion); checkpoint upload failure is best-effort (log and continue). The FINAL adapter upload is a hard gate: failure raises → non-zero exit → no RUN_SH_DONE → the monitor never treats the box as cleanly done (degrades to GONE_UNKNOWN→failed, never a false DONE_OK).
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(7).md#02_train; NOTES_phase5_training_launcher_running(45).md#update-2026-06-05
- **relations:** run.sh markers (FTW-030); checkpoint durability manifest (FTW-032); save_steps cadence
- **verify-later:** 02_train in the live bundle; checkpoint sizes in B2

### FTW-034 — Resume path (cluster-side checkpoint selection, presence-of-checkpoints as the signal)
- **status:** partial
- **status-evidence:** "batch + resume BUILT, APPLIED, and verified [def-state]… still unproven in prod" as of the most recent dated update.
- **what:** A relaunch for the same training_run_id becomes a continuation automatically: the launcher's `check_resume` step asks the adapter (`prepare_resume_url`) to list `finetuning/checkpoints/<run_id>/` in B2 (reusing the existing storage-client ListObjects call), pick the highest ckpt-N, and presign a GET; `assemble_manifest` emits a `resume` block only when one is found. `02_train` downloads/extracts it and calls `trainer.train(resume_from_checkpoint=True)`. There is no separate resume mode — an empty prefix means fresh start; found=false is a valid answer, and transient list failures return error_recoverable so the chassis retries.
- **sources:** working/phase5/README_what_is_manifest.md; PLAN_checkpoint_and_artefact_upload_b2(7).md#resume; 111_training_launcher_resume_wiring.sql (header)
- **relations:** checkpoint durability manifest (FTW-032); monitor GONE_UNKNOWN (the total-loss case that motivated this, model-infrastructure)
- **verify-later:** a real kill-and-resume test; dispatch_thunder_prepare_resume_url in registry

### FTW-035 — Monitor enablement gate: DONE must mean durable
- **status:** partial
- **status-evidence:** "enable thunder-training-monitor (safe once DONE means durable — gated on the first run actually reaching RUN_SH_DONE)"; "Not enabled; enabling is RUNBOOK step 6" as of the latest dated notes.
- **what:** An explicit sequencing invariant: the training monitor's DONE_OK path decommissions the box (destroying the disk), so the schedule stays disabled until the upload path proves that RUN_SH_DONE genuinely implies the adapter is durable in B2. Enabling early would have destroyed iter_0's adapter. The interim protocol for in-flight runs was manual: scp the adapter off the box before anything decommissions it.
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(7).md#build-order; FOCUS_finetuning_flywheel_and_service(25).md#update-2026-06-04
- **relations:** CheckpointUploader hard gate (FTW-033); run.sh markers (FTW-030); thunder-training-monitor (model-infrastructure)
- **verify-later:** scheduled_tasks.enabled for thunder-training-monitor; whether a run has since reached RUN_SH_DONE

### FTW-036 — knowledge_base RAG store + Flywheel B verification
- **status:** deployed
- **status-evidence:** "Prefix patch deployed and verified live 2026-04-21 … log line \"prefix_applied\":true observed on rag_lookup step"; "Flywheel B is done" (chassis v1.0.979).
- **what:** From the flywheel's perspective, Lane B (RAG) is the `knowledge_base` table (migration 082, pgvector 768-dim, ivfflat+cosine) readable/writable by any agent via `rag_lookup`/`rag_index`, with trigram fallback when Ollama is down and a metadata-filter-first-then-similarity retrieval rule. It was verified bottom-up in three single-focus steps (pgvector+ivfflat+cosine on synthetic vectors, real-content retrieval through cpu-ollama, then chassis integration via a deterministic 3-step `rag-test-agent`), and is empirically established as requiring `search_document:`/`search_query:` nomic task prefixes for correct ranking (see FTW-037 for the specific bug this fixed). The full mechanism (table shape, actions, best practices) is registered in more depth under the rag-knowledge-base category (RAGK-001/RAGK-002/RAGK-003); this entry is the flywheel-lane record of its deployment and verification. Open task noted: periodic REINDEX of the ivfflat index was never scheduled.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.2,#2.4b; FOCUS_finetuning_flywheel_and_service(25).md#2.2,#2.4b; working/flywheel_docs/flywheel_B_step*.{sql,sh} (headers)
- **relations:** RAG knowledge_base (rag-knowledge-base category, RAGK-001); nomic prefix patch (FTW-037); cpu-ollama adapter (model-infrastructure); three compounding channels (FTW-002)
- **verify-later:** rag_actions.go; knowledge_base row counts; whether REINDEX was ever scheduled

### FTW-037 — Nomic task prefixes are load-bearing (rag_actions prefix patch)
- **status:** deployed
- **status-evidence:** "Prefix patch deployed and verified live 2026-04-21 … log line 'prefix_applied':true observed."
- **what:** Without `search_document:`/`search_query:` task prefixes, nomic embeddings ranked a Labrador chunk above a French Bulldog chunk on a BOAS-specific query; with prefixes the correct result won with 5x the margin. The patch adds a model-scoped `applyNomicPrefix` helper (only nomic-embed-*, double-prefix guard, prefix_applied logged) at both embed call sites; stored chunks and dedup hashes stay unprefixed, and the trigram fallback is untouched.
- **sources:** working/flywheel_docs/PATCH_rag_actions_nomic_prefixes.md; FOCUS_finetuning_flywheel_and_service(25).md#2.4b
- **relations:** knowledge_base RAG store (FTW-036); Ollama CPU adapter (model-infrastructure)
- **verify-later:** rag_actions.go contains applyNomicPrefix

### FTW-038 — Thunder instance lifecycle: reaper + training monitor (orchestrator/worker)
- **status:** deployed
- **status-evidence:** migrations create the reaper (every 15 min, idempotent decommission) and the monitor orchestrator/worker pair, with verified coordinator internals and an insert-DISABLED-until-actions-deploy discipline.
- **what:** Cost/safety controls for rented GPUs, as captured at the SQL-migration level: `thunder-reaper` decommissions instances past `max_uptime_hours` (one per tick, pre_query LIMIT 1); `thunder-training-monitor` is a separate orchestrator that finds every running training instance each tick and spawns a per-instance worker to probe via SSH, classify (alive / unreachable-streak / done_ok / done_fail), reconcile `training_runs`, and decommission on terminal verdicts. The orchestrator-with-loop shape deliberately differs from the reaper's single-row pre_query (it must visit every instance, not just the top one), and the loop must stay sequential for topic-reuse safety. The model-infrastructure register carries the deeper mechanical breakdown of the monitor and reaper as separate entries — this entry is the SQL/migration-level summary of the pair as a unit.
- **sources:** 114_thunder_reaper.sql; 116_thunder_training_monitor_worker.sql; 117_thunder_training_monitor_orchestrator.sql
- **relations:** scheduler-and-tasks pre_query dispatch patterns; thunder adapter (FTW-040); model-trainer chain (FTW-027); thunder-training-monitor, Monitor/reaper responsibility split, thunder-reaper (all model-infrastructure, more detailed)
- **verify-later:** scheduled_tasks rows enabled; probe/classify actions

### FTW-039 — LLM fallback extraction doubling as training data (medicine pricing)
- **status:** deployed
- **status-evidence:** "response logged to llm_call_log… dual purpose: price extraction now, training data for future LoRA."
- **what:** In a medicine-pricing scrape pipeline, regex handles ~90% of pages; a CPU Mistral model (mistral-small3.1, temp 0.1) parses the remainder into a JSON variant array at 80-280s/call — acceptable because the workload is batch-tolerant — while every call incidentally accumulates markdown→JSON pairs toward a future fine-tune. Named future extensions: product matching across retailers, price alerts from snapshot history, an affiliate-feed switch.
- **sources:** 008_*.md#LLM Fallback, #Future Work
- **relations:** batch processing categorisation; llm_call_log flywheel (FTW-004)
- **verify-later:** llm_call_log provider='ollama' step_name='scrape_prices'

### FTW-040 — Thunder adapter (credential-boundary GPU provisioning with caps and reaper)
- **status:** deployed
- **status-evidence:** decisions confirmed and cited as running in production (provision ~400s, SSH findings, reaper window arithmetic all observed live).
- **what:** All Thunder Compute interaction routes through a long-lived cluster adapter holding THUNDER_COMPUTE_API_KEY/B2 keys/ephemeral SSH keys; VMs are per-run ephemeral and credential-free (presigned URLs only — a compromise's blast radius is limited to that run's files). Operational caps: $100/day rolling spend, 18h hard uptime default, concurrency 2, a 15-min reaper reconciling the Thunder API against `thunder_instances`. Formally retracts the on-VM HTTP job-server option (superseded by the adapter-dispatch design, FTW-026). Hard-won operational lessons: lowercase gpu_type enums, the real template name is `base` (not the OpenAPI example), `tnr connect` does server-side setup, the login user is `ubuntu` not `root`, and a live-instance-scoped partial unique index is needed for recycled provider ids. The model-infrastructure register carries a more mechanically detailed entry for the same adapter's provision/decommission internals; this entry is the credential-boundary/caps framing.
- **sources:** 033_thunder_adapter_design.md (full); 016 debugging guide §9 Thunder entries
- **relations:** batch drain mode; training launcher/monitor (FTW-027, FTW-038); Thunder Compute adapter (model-infrastructure, provision/decommission internals)
- **verify-later:** thunder-adapter deployment; reaper task

### FTW-041 — Text LoRA — veterinary knowledge extractor
- **status:** aspirational
- **status-evidence:** Phase E todo "Text LoRA fine-tuning (week 6-7)" unchecked; a full Unsloth/QLoRA recipe is given as instructions but no run is recorded.
- **what:** A concrete recipe to fine-tune a local 7-8B model (Unsloth QLoRA, r=16, 3 epochs) on accumulated knowledge-extraction examples, export Q4_K_M GGUF, load into Ollama, and swap the `knowledge-extractor` agent to the local model to eliminate Claude API cost per extraction. Training data was meant to accrue naturally during a canine-biology research phase (50 breeds + 30 conditions + 40 procedures ≈ 120 examples, needing 200+ before triggering the fine-tune).
- **sources:** docs023.../018_canine_biology.md#6
- **relations:** Fine-tuning pipeline / candidate criteria (FTW-023); llm_call_log (FTW-004); Ollama provider (model-infrastructure)
- **verify-later:** vet-extractor GGUF/Ollama model; knowledge-extractor agent provider
