# Cluster: finetuning-model-infra
Categories included: finetuning-flywheel, model-infrastructure, new:rag-knowledge-base, llm-quality-testing, new:llm-call-observability


<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM call logging (llm_call_log) as ops visibility + training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 001(5): "Verified in production (March 2026) — 57+ rows"; 085 migration adds flywheel columns
- **what:** Every execute_llm_prompt call logged fire-and-forget (agent_type, step, model, rendered prompt, response, tokens, latency, `__sent_temperature`/`__sent_max_tokens` write-backs). Flywheel columns (work_item_id, prompt_variant, vertical, rag_context_used) link calls to outcomes for LoRA/RAG training exports. Known past bugs: schema/Go column drift (agent_id vs client_id), empty agent_type from buildActionParams.
- **sources:** 001(5)#LLM call logging, #Implementation Status; 022_ai_endpoint_health_and_flywheel_llm_call_log.sql
- **relations:** batch queue LogLLMCall paths; fine-tuning path
- **verify-later:** llm_call_log schema; llm_call_logger.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fine-tuning path (log → export → LoRA → GGUF → Ollama → swap)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 009 (2026-07-10 update): one training run `complete`, GPU provisioning via Thunder real and dynamic, but "No agent_definitions row currently points ai_service at llama3.3:70b — trained and tested, never used for production inference"
- **what:** Pipeline: accumulate 200+ examples in llm_call_log → export (Alpaca/ChatML) → LoRA fine-tune on GPU (unsloth) → GGUF → Ollama → swap agent definition → A/B against Claude. Candidates are short-output classifiers (site-classifier, vet-practice-verifier, etc.). The last mile (wiring the trained 70B into live inference) is explicitly outstanding.
- **sources:** 001(5)#Fine-tuning path; 009#Future incl. 2026-07-10 note; 023 (manual A/B comparisons)
- **relations:** model swap functions; Thunder adapter; drain mode
- **verify-later:** model_lifecycle.training_runs; agent_definitions ai_service providers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM fallback extraction doubling as training data (med pricing)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 008: "response logged to llm_call_log… dual purpose: price extraction now, training data for future LoRA"
- **what:** Regex handles ~90% of pages; CPU Mistral (mistral-small3.1, temp 0.1) parses the remainder into a JSON variant array at 80-280s/call — acceptable because batch-tolerant — while accumulating markdown→JSON pairs for a future fine-tune. Future: product matching across retailers, price alerts from snapshot history, affiliate-feed switch.
- **sources:** 008#LLM Fallback, #Future Work
- **relations:** batch processing categorisation; llm_call_log flywheel
- **verify-later:** llm_call_log provider='ollama' step_name='scrape_prices'

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Thunder adapter (credential-boundary GPU provisioning with caps and reaper)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 033 confirmed decisions; 016 §9 entries show it running in prod (provision 400s, ssh findings, reaper window maths)
- **what:** All Thunder Compute interaction routes through a long-lived cluster adapter holding THUNDER_COMPUTE_API_KEY/B2 keys/ephemeral SSH keys; VMs are per-run ephemeral and credential-free (presigned URLs only; compromise blast radius = that run's files). Caps: $100/day rolling, 18h hard uptime, concurrency 2, 15-min reaper reconciling API↔thunder_instances. Formally retracts the on-VM HTTP job-server option. Operational lessons: lowercase gpu_type enums, template 'base', OpenAPI examples aren't valid values, tnr connect does server-side setup, login user ubuntu not root, live-instance-scoped partial unique index for recycled provider ids.
- **sources:** 033 full; 016 §9 Thunder entries
- **relations:** batch drain mode; training launcher/monitor
- **verify-later:** thunder-adapter deployment; reaper task

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Finetuning flywheel four-lane programme (A export, B RAG, C training, D eval)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (doc last touched 2026-04-23)
- **what:** The internal AI training flywheel: production LLM calls become training data (A), verified knowledge feeds RAG (B), local models get fine-tuned on the exported data (C), and quality is compared against Claude (D), so local models replace API calls where quality holds and costs drop.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#1, #2, #15-changelog
- **relations:** llm_call_log capture; training-data export pipeline; Flywheel C QLoRA; Flywheel D replay-eval; three compounding improvement channels
- **verify-later:** `training_exports` schema; `flywheel_A_v3/`, `flywheel_C/` script dirs; `llm_call_log`; whether a first GPU training run ever happened after 2026-04-23

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### llm_call_log LLM call capture
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "llm_call_log populating (flywheel columns wired through ai_actions.go)" [x]; but "Cleanup function exists (cleanup_old_llm_logs) but nothing schedules it yet" (2026-04-23)
- **what:** Every LLM call writes agent_type, step_name, model, prompt_rendered, response_text, tokens, latency, success, work_item_id, prompt_variant, vertical, rag_context_used to `llm_call_log` (migrations 081/085) via a fire-and-forget `LogLLMCall` goroutine. Retention 90/180 days; the join key for "what to train".
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.1, #4.2
- **relations:** training-data export; prompt evolution (prompt_variant A/B); vision cost tagging (PLAN_imagery_loop_closure 5.4)
- **verify-later:** `ai_actions.go` LogLLMCall; scheduling of `cleanup_old_llm_logs`; agent_type population on old rows

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### knowledge_base RAG with load-bearing nomic task prefixes
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "Prefix patch deployed and verified live 2026-04-21 … log line \"prefix_applied\":true observed on rag_lookup step"; "Flywheel B is done" (chassis v1.0.979)
- **what:** `knowledge_base` table (migration 082, pgvector 768-dim, ivfflat+cosine) readable/writable by any agent via `rag_lookup`/`rag_index`, with trigram fallback when Ollama is down. Empirically established that `search_document:`/`search_query:` prefixes for nomic-embed-text are mandatory for correct ranking; metadata-filter-first-then-similarity is the retrieval rule.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.2, #2.4b
- **relations:** finetuning.uk RAG product reuses this infra; cpu-ollama adapter
- **verify-later:** `platform/orchestration/actions/rag_actions.go` applyNomicPrefix; `knowledge_base` metadata fields (vertical, component_type, source_quality)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Training-data export as chassis agent + action, Postgres-backed
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "First real training dataset now in Postgres: export_id fef7be6b…, 1,958 rows, 21.2MB, reconciled manually. Spawning architecture fully validated" (2026-04-23); v3.2 strict-UPDATE code "awaiting chassis rebuild/deploy" at doc date
- **what:** Export of (prompt, response) pairs from llm_call_log as ChatML messages with metadata sidecar (source_log_id, agent_type, step_name, export_version), fence-stripped and JSON-validated, via `training_data_export` action + `training-data-exporter` worker + `training-data-export-orchestrator` wrapper. v3 writes named snapshot datasets into `training_exports.runs`/`training_exports.rows` (rejected: real-time streaming — batch snapshots preserve "the dataset we trained on"). Evolved v1 (template config, failed) → v2 (file output, wrong pod) → v3 (Postgres, dedicated pod).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** orchestrator wrapper pattern; pgbouncer per-batch commits; negative-examples/DPO decision
- **verify-later:** `training_exports` schema in DB; `training_data_export_v3.go` and registry entry; whether v3.2 landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel C QLoRA training pipeline (Unsloth, Llama 3.3 70B)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Pipeline shape now concrete, scripts written, awaiting first training run on a GPU VM" (2026-04-23)
- **what:** Five scripts (`flywheel_C/00-03` + README) pull a named export from Postgres, train a LoRA adapter on Llama 3.3 70B Instruct via Unsloth QLoRA on a single H100/A100, sanity-check JSON validity of outputs. 70B chosen because hardware was available; 8B flagged as likely 95% of quality at 10% cost — comparison run planned.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5
- **relations:** Flywheel C phase 2 automation; Flywheel D eval harness scores the result
- **verify-later:** existence of `flywheel_C/` scripts; any `model_training_runs` table or LoRA artefacts

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel C phase 2 automation (chassis drives, GPU VM serves)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "design locked, not built … Gated on phase 1 (manual training run)" (2026-04-23)
- **what:** HTTP job server (~200 lines, POST /jobs → poll → fetch adapter, bearer auth) on the GPU VM; three new chassis components (model-trainer, model-evaluator specialists, training-flywheel-orchestrator wrapper chaining export→train→eval→conditional swap_agent_model); three new tables (model_training_runs, model_artefacts, model_evaluations). Rejected SSH-exec and Kafka-consumer-on-VM alternatives.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5.1
- **relations:** model swap/revert functions; scheduled_tasks trigger option
- **verify-later:** whether any of the three agents/tables exist in agent_definitions / DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel D replay-eval methodology + dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel D (eval) paused"; ollama-eval pod deployed with memory bump "requests: 24Gi / limits: 28Gi … fix persisted into kustomize base" (2026-04-23)
- **what:** Replay-don't-re-run evaluation: pull 20 diverse stored Claude prompts (DISTINCT ON orchestration_id), POST to candidate model via /api/chat, compare with 3 levels (structural jq checks → Claude-as-judge → manual review). Shared cpu-ollama contention made eval impractical (27 min/case), so a dedicated `ollama-eval` deployment (own PVC/service, invisible to production routing) was created. Memory rule: pod limit ≥ model file size + 8–12 GiB.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4d, #2.4d-comparison
- **relations:** Flywheel C evaluation preconditions; Ollama CPU ops envelope
- **verify-later:** `deployments/kustomize/services/ollama-eval/`; results.jsonl outcome; whether eval ever completed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Negative examples reserved for DPO, excluded from SFT
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "For our first training run: plain SFT, edge cases excluded" (decision 2026-04-22)
- **what:** Edge-case rows where Claude correctly produced prose instead of JSON are positive examples of the wrong shape for SFT and must be excluded from exports; they become the "rejected" side of DPO preference pairs later. Keeps them in llm_call_log, out of training_exports.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e "Negative examples / edge cases"
- **relations:** training-data export filters (strict_json)
- **verify-later:** export action's strict_json / prose-exclusion behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Three compounding improvement channels (RAG, LoRA, prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "From 009_model_infrastructure.md, decision 10" — RAG immediate, LoRA medium-term, prompt_variant A/B ongoing (2026-04)
- **what:** Three ways local model quality improves, each useful alone but compounding: RAG injects verified knowledge now; LoRA replicates a task cheaply later; `prompt_variant` column enables prompt A/B evolution. Good prompts + good RAG produce the best training data.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#3
- **relations:** flywheel programme; llm_call_log prompt_variant
- **verify-later:** any actual prompt_variant usage in llm_call_log

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Fine-tune candidate selection heuristic
- **category:** finetuning-flywheel
- **status-signal:** unknown
- **status-evidence:** Priority table (2026-04): knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher good; page-content-writer/visual-design-auditor/chief-strategist "not good candidates"
- **what:** Agents with high-volume, structured-JSON, short outputs are swap candidates for local models; long creative output and judgement tasks stay on Claude. Drives which agent/step gets exported and trained first.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.6
- **relations:** Flywheel D target choice (page-content-writer iter_0 chosen for eval despite being a poor swap candidate, because it had the most logged data)
- **verify-later:** actual llm_call_log volumes per agent

<!-- SOURCE: U04_idea_uk.md -->
### Training checkpoint/adapter upload to B2 via pre-minted presigned single-object PUTs
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Status: Phase A (02_train) BUILT 2026-06-05, isolation test pending. Phases B–D not built." — with the isolation-test harness and its artefacts present in nginx/_iso/.
- **what:** Solves three coupled gaps of ephemeral Thunder training VMs (final adapter not durable; the monitor's DONE_OK path decommissions the disk; no checkpoints/resume on a ~24h run): the launcher pre-mints single-object, write-only presigned PUT URLs and hands them to the VM in a manifest — the hostile-VM threat model rejects standing scoped keys and callback endpoints (nothing on the box can mint or write beyond the fixed URL set). Explicit framing: this protects access, not artefact integrity — a malicious-but-valid adapter still needs the flywheel-D eval gate before promotion. Phase A validated box-free by isolation_test_phase_a.py (presign/PUT signature, tar round-trip, checkpoints/ exclusion, GET+extract byte-identical).
- **sources:** idea.uk/nginx/PLAN_checkpoint_and_artefact_upload_b2(1).md; idea.uk/nginx/isolation_test_phase_a.py (header); idea.uk/nginx/README_get_b2_details.md
- **relations:** Thunder adapter; B2 dead-drop (same one-way pattern); model-infrastructure eval gate.
- **verify-later:** 02_train upload hooks; thunder-training-monitor-worker decommission path.

<!-- SOURCE: U06_finetuning.md -->
### Finetuning flywheel programme (lanes A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 state block: "Flywheel A ✓ operational / B ✓ done / C phase 1 ✓ / C phase 2 → next priority / D ✓ iter_0 evaluated"; later phase-5 notes carry C-phase-2 to a mostly-green automated launch (2026-06-09) with the monitor still disabled.
- **what:** The internal self-improvement programme: the site-building pipeline emits training data as a byproduct (A: export), RAG injects verified knowledge immediately (B), local models are periodically fine-tuned on the captured data (C), and evaluated against Claude before any swap (D). The strategic goal is dropping API cost by swapping local models in for Claude calls where quality holds.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#1-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#current-state; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** every other concept in this unit; three improvement channels; model swap/revert
- **verify-later:** training_exports schema, model_lifecycle schema, llm_call_log columns, thunder-adapter deployment, scheduled_tasks rows for reaper/monitor

<!-- SOURCE: U06_finetuning.md -->
### llm_call_log training-data capture
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.1 "Every LLM call in the system writes to llm_call_log (migration 081, flywheel columns added in 085)"; §4.1 checkbox "[x] llm_call_log populating".
- **what:** Every LLM call logs agent_type/step_name (the "what to train" join key), model/provider, prompt_template/prompt_rendered/response_text (the raw training pair), token/latency/cost signals, success/error, work_item_id, prompt_variant and vertical. Write path is `LogLLMCall` in ai_actions.go — fire-and-forget goroutine, 5s timeout, never blocks the workflow. Retention 90d success / 180d error via `cleanup_old_llm_logs`, which exists but is NOT scheduled (open task). Historically-empty `agent_type` rows exist; recent writes are 100% populated.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.1,#2.4c,#4.2
- **relations:** training-data export agent; replay eval (pulls stored prompts); per-vertical slicing; prompt evolution (prompt_variant)
- **verify-later:** migrations 081/085; platform actions ai_actions.go LogLLMCall; cleanup_old_llm_logs scheduling; llm_call_log row counts

<!-- SOURCE: U06_finetuning.md -->
### Training-data export as chassis agent + action (flywheel A, v1→v3.2)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23: "training_data_export.go v3.2 … deployed as v3.2, verified working"; two 1,958-row exports landed in Postgres.
- **what:** Training export is a first-class chassis pipeline component, not ad-hoc SQL: a `training_data_export` action plus `training-data-exporter` worker and `training-data-export-orchestrator` wrapper. It queries llm_call_log, strips markdown fences via the shared `stripMarkdownFromResponse`, validates JSON per row, and writes batched inserts into `training_exports`. The v1→v3.2 evolution encodes several platform lessons (template-config not rendered for deterministic actions; file output on ephemeral pods; pgbouncer transaction limits; RowsAffected checks).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** training_exports schema; orchestrator wrapper spawning pattern; pgbouncer per-batch commits
- **verify-later:** platform/orchestration/actions/training_data_export.go; agent_definitions rows training-data-exporter / training-data-export-orchestrator; training_exports.runs contents

<!-- SOURCE: U06_finetuning.md -->
### training_exports Postgres schema (named snapshot datasets)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23 "Schema (applied, verified)"; export_ids 146a9a12 and fef7be6b each hold 1,958 rows.
- **what:** `training_exports.runs` (one row per export with filters, counts, completed_at) + `training_exports.rows` (ChatML messages + metadata JSONB per training record, unique on (export_id, source_log_id)). Datasets live in Postgres, not files or S3 — named, versioned snapshots referenced by export_id UUID, streamed out to JSONL at training time. Real-time streaming into the table was explicitly rejected to keep snapshot boundaries and avoid coupling observability to training. Known trap: `runs.rows_exported` can disagree with actual `rows` content (export a8484922 recorded 1957 but holds 0 rows) — always verify with a count before launching.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#done-verified; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06-2
- **relations:** training-data export agent; training-data-preparer (streams export to S3); dataset pull scripts
- **verify-later:** training_exports schema in clients_db; the a8484922 zero-row anomaly; recent_runs view

<!-- SOURCE: U06_finetuning.md -->
### ChatML export format with metadata sidecar
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22."
- **what:** Training rows are `{messages:[{role:user},{role:assistant}], metadata:{source_log_id, agent_type, step_name, orchestration_id, model, created_at, export_version}}`. Chosen for chat-tuned base-model parity, trainer-tool defaults (Unsloth/Axolotl), and `/api/chat` training-inference parity. Metadata gives row-level traceability back to llm_call_log; export_version future-proofs format evolution. Whole prompt_rendered goes in the user turn (no system/user split yet).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e; working/phase1/files/flywheel_A_export_page_content_writer_iter0.sql (header)
- **relations:** response cleaning; training_exports schema
- **verify-later:** export SQL/action output shape vs current llm_call_log columns

<!-- SOURCE: U06_finetuning.md -->
### Response cleaning and SFT negative-example exclusion
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e iter_0 dataset audit table (97.4% clean JSON, 2.6% fenced, fences stripped on export); "For our first training run: plain SFT, edge cases excluded".
- **what:** Exports must strip markdown code fences (else the fine-tune learns to emit fences, exactly what prompts forbid) and exclude edge-case prose responses: plain SFT has no "don't do this" signal, so intelligent edge-case answers are positive examples of the wrong shape. Those rows stay in llm_call_log as future "rejected" halves of DPO preference pairs — DPO/RLHF is the named future home for negative examples.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e
- **relations:** ChatML export format; `<no value>` contamination
- **verify-later:** stripMarkdownFromResponse; whether any DPO work exists later

<!-- SOURCE: U06_finetuning.md -->
### `<no value>` training-data contamination and the iter_1 filter floor
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-07 known issues: "Training rows from before that fix include the literal token `<no value>`. iter_1's export should filter created_at >= <fix_date>… Action: note the fix-deploy date"; still listed as needed in HANDOFF 2026-05-08.
- **what:** A prompt-builder rendering bug injected the literal token `<no value>` into production prompts; iter_0's training data (and its eval cases) inherit it. The fix-deploy date becomes the created_at filter floor for the iter_1 export. As of the last docs, the date had not been recorded — an open data-hygiene debt.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#known-issues; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#revised-iter_1-priorities
- **relations:** training_exports; held-out eval set (same artefact present)
- **verify-later:** git log for the prompt-builder fix; whether any iter_1 export exists with the filter

<!-- SOURCE: U06_finetuning.md -->
### Dataset profile and schema heterogeneity of page-content-writer iter_0
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4g dataset profile tables (n=1,957; p50 prompt 8,250 chars; three dominant JSON schemas: hero-with-CTAs 68%, minimal hero 18%, header/nav 9%).
- **what:** One (agent_type, step_name) training slice actually spans three component output schemas; the model must learn schema selection conditioned on the "Component: X" text in the prompt. A first-pass option to filter to the top-2 schemas (86% of rows) was noted but the full set was trained. Prompt/response size distribution anchors max_seq choice (some prompts approach 4,000 tokens → max_seq 4096).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/flywheel_docs/terminology.md#seq-4096
- **relations:** Llama 70B QLoRA config; inference-test success criteria (keys match trained schemas)
- **verify-later:** training_exports.rows key distribution query

<!-- SOURCE: U06_finetuning.md -->
### Flywheel C manual training pipeline (00–03 scripts, smoke-gates-full)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07: "executed flywheel C phase 1 end-to-end on a Thunder Compute A100 80GB… iter_0 adapter (791MB) … 5/5 valid JSON".
- **what:** Four scripts define the training path: `00_vm_setup.sh` (idempotent pinned env), `01_pull_dataset_from_postgres.sh`, `02_train_llama_3_3_70b.py` (Unsloth QLoRA, CLI-configurable, emits manifest), `03_inference_test.py` (JSON validity + schema keys sanity). Discipline: a 20-row/1-epoch smoke train and smoke inference always gate the full run — cheap insurance on unattended runs. The same bundle later becomes the automated launcher's payload.
- **sources:** working/flywheel_docs/README.md; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#flywheel-c; working/scripts/run.sh (header)
- **relations:** iter_0 baseline run; scripts bundle as deployment unit; run.sh markers
- **verify-later:** working/scripts/* against the bundle actually in B2 (finetuning/scripts/bundle.tar.gz)

<!-- SOURCE: U06_finetuning.md -->
### Base-model decision: Llama 3.3 70B Instruct QLoRA (with 8B ablation planned)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.5 "Base model: Llama 3.3 70B Instruct. Decision taken 2026-04-23"; HANDOFF 2026-04-23 decisions-locked list.
- **what:** `unsloth/Llama-3.3-70B-Instruct-bnb-4bit` via Unsloth QLoRA on a single A100/H100 80GB; defaults 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096. 70B was chosen because hardware was available and a strong baseline is useful — with an explicit acknowledgment that 8B likely delivers ~95% of quality at ~10% of inference cost for this narrow structured-JSON task; a same-dataset 8B comparison run is planned but never executed in these docs.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#base-model-decision
- **relations:** iter_0 baseline; cost anchors; epochs ablation
- **verify-later:** 02_train defaults; any 8B run in model_lifecycle.training_runs

<!-- SOURCE: U06_finetuning.md -->
### iter_0 baseline training run (real cost/time/loss anchors)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 run summary: 33,136s ≈ 9h12m, final_loss 0.2669 (trailing), peak VRAM 44.8GB, ~$20 total; "Anchor future estimates against $20/iter".
- **what:** The first real fine-tune: 1,958 rows → 1,934 effective, clean loss curve (ep1 1.49→0.27, ep3 →0.10), adapter 791MB fp32 safetensors. Epoch-3 loss gap suggests memorisation → a 2-epoch ablation is queued for iter_1. Cost anchor $20/iteration (training) + $1.50 (eval) ≈ $22/cycle. Later automated runs corrected the wall-time estimate: the full run is ~24h at ~119s/step without FA2, not the "30–90 min" the README claimed.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md; working/phase2/README.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** version pinning; snapshot economics; per-instance uptime bump; fp16 save decision
- **verify-later:** lora_iter0_full/manifest.json; model_lifecycle.training_runs rows 1cd65dd7/e6ab9fad

<!-- SOURCE: U06_finetuning.md -->
### GPU environment version pinning (cu124 stack)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Version pin discoveries (essential for any future run)" table with per-pin rationale.
- **what:** The working training environment is a narrow pin set: torch 2.6.0+cu124, transformers<5, torchao<0.17 (transformers imports torchao eagerly; incompatible torchao breaks import entirely), prebuilt flash-attn wheel (Thunder's Ollama template ships CUDA runtime, no nvcc), unsloth+unsloth_zoo both explicitly (git install misses the zoo), hf_transfer as separate package. cu124 is flagged a dead end — next rebuild should move to cu126/cu128. The Unsloth template (used for eval) differs: nvcc present, torch 2.10/cu128, xformers pre-installed, FA2 absent.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#version-pin-discoveries; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons; working/eval/iter0_eval/001_README.md
- **relations:** 00_vm_setup.sh as canonical environment; snapshot economics
- **verify-later:** working/scripts/00_vm_setup.sh pin lines

<!-- SOURCE: U06_finetuning.md -->
### Snapshot economics: setup script beats VM snapshots
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Created `unsloth-trainer-base-01` then deleted it… Break-even: ~18 training runs/month. Reality: 1-4."
- **what:** Thunder snapshots bill the full provisioned 100GB ($15/month) regardless of used bytes, saving only ~25min/$0.85 per cold start — uneconomic below ~18 runs/month. Decision: no snapshots; the version-pinned idempotent `00_vm_setup.sh` is the canonical, version-controlled "snapshot". Phase-2 automation therefore provisions fresh instances every run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#snapshot-decision
- **relations:** version pinning; phase-2 architecture
- **verify-later:** none (economic decision); revisit if run tempo >15/month

<!-- SOURCE: U06_finetuning.md -->
### GPU training performance model (smoke ≠ steady state; FA2; seq-length cost)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** terminology.md whole file; NOTES(45) 2026-06-04: "The smoke rate (116 s/step) predicted this — nobody extrapolated it".
- **what:** A small captured mental model of training performance: smoke-test speed is unrepresentative (one-time kernel autotune/CUDA-graph costs amortized over too few steps); steady-state emerges after 5–20 steps; FA2 vs xformers/SDPA is a 2–4× attention-speed lever; attention scales O(N²) so max_seq 4096 quadruples 2048's attention work. Operationally: extrapolate full-run wall time from smoke s/step (the 18h-cap overrun happened because nobody did).
- **sources:** working/flywheel_docs/terminology.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** iter_0 baseline; per-instance uptime bump; cap-sizing-from-smoke queued idea
- **verify-later:** whether cap-sizing-from-smoke was ever built

<!-- SOURCE: U06_finetuning.md -->
### fp16 adapter save decision
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** HANDOFF 2026-05-08 decisions: "Save adapters as fp16, not fp32, in iter_1. One-line script change." (iter_0 shipped fp32 at 791–828MB.)
- **what:** PEFT `save_pretrained()` defaults LoRA weights to fp32 even when training in bf16, doubling adapter size and transfer time (17min tnr scp for 791MB). The one-line fix (cast trainable params to fp16 pre-save) was agreed for iter_1 but iter_1 never ran in these docs.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#lessons(10); working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** adapter transport via S3; model_lifecycle.artefacts format field
- **verify-later:** current 02_train script save path

<!-- SOURCE: U06_finetuning.md -->
### Flywheel D replay-eval methodology
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Replay, don't re-run"; §2.4d full replay design and partial results.
- **what:** Evaluation replays stored production prompts from llm_call_log against the candidate model instead of re-invoking agents — no orchestration-state pollution, much faster, and directly comparable to the stored Claude output. Test sets use `DISTINCT ON (orchestration_id)` for diversity ("Diverse 20 > random 20"), exported as NDJSON. Fail fast on empty responses; monitor with a watch loop.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/flywheel_D_target_selection.sql (header)
- **relations:** three-level eval pipeline; held-out set; CPU-Ollama eval attempt
- **verify-later:** held_out_cases_v1.sql query against llm_call_log

<!-- SOURCE: U06_finetuning.md -->
### Three-level evaluation pipeline (L1 structural / L2 judge / L3 spot-check)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 002_README-flywheele_D_evaluation_pipeline.md (run instructions, ~$1 / ~5min total); iter0_evaluation_report.md generated 2026-05-08.
- **what:** Reusable eval stack: L1 structural metrics computed locally and side-by-side for both models (JSON validity, schema-key match, length ratios, forbidden phrases from the brief's avoid-list, fabrication-marker regexes); L2 Claude-as-judge scoring relevance/voice/integrity 1–5 plus winner, with anonymised randomised A/B and resume support; L3 auto-selected spot-check cases folded into a markdown report by build_report.py. The report deliberately reports confounds and makes no ship/no-ship call. Known limit: L1 fabrication regexes have poor recall — contextual fabrications need L2/L3.
- **sources:** working/flywheel_docs/002_README-flywheele_D_evaluation_pipeline.md; working/eval/iter0_eval/iter0_evaluation_report.md#methodology; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(6)
- **relations:** model_lifecycle.evaluations (l1_metrics/l2_metrics JSONB contract); judge-model choice; eval gate
- **verify-later:** working/eval/iter0_eval/05-07 scripts; any evaluations rows in model_lifecycle

<!-- SOURCE: U06_finetuning.md -->
### Held-out eval set v1 as the canonical cross-iteration comparison set
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decision: "held_out_cases_v1.jsonl is the canonical eval set across iterations — same 20 cases… so trends are meaningful."
- **what:** 50 cases pulled from llm_call_log post-training-export-cutoff (created_at > 2026-04-23 14:54:32Z), one per orchestration, defensively excluded from the training set by source_log_id; 20 used for iter_0, 30 reserved. The SQL is kept for reproducibility. Iterations evaluate against the same cases so deltas are trend, not noise. Fresh `_v2` sets are the mechanism for novelty checks.
- **sources:** working/eval/v1/held_out_cases_v1.sql; working/eval/iter0_eval/iter0_evaluation_report.md#sample-selection; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** replay eval; three-level pipeline
- **verify-later:** held_out_cases_v1.jsonl vs training export overlap

<!-- SOURCE: U06_finetuning.md -->
### Claude-as-judge with anonymised A/B and self-recognition bias handling
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** iter0 report L2: "5 cases had identical R/V/I scores… 5/5 went to Claude… consistent with residual self-recognition bias"; HANDOFF 2026-05-08 decision: "claude-opus-4-7 is the canonical judge model".
- **what:** Judge design: anonymise responses, randomise A/B positions, score dimensions before picking a winner, and use a *different* Claude model (Opus) than the training-label producer (Sonnet 4.6) to reduce — not eliminate — self-recognition. The bias was empirically observed: rubric-tied cases broke for Claude every time, so headline win-rates get an adjusted reading (16-4 → 12-4 with 4 judge-preference ties). Position bias is checked explicitly (A won 55%, no clear bias).
- **sources:** working/eval/iter0_eval/iter0_evaluation_report.md#level-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(5)
- **relations:** three-level pipeline; model_lifecycle.evaluations judge_model column (judge drift tracking index)
- **verify-later:** level2.py anonymisation logic

<!-- SOURCE: U06_finetuning.md -->
### iter_0 verdict: shippable for low-stakes; voice fidelity is the iter_1 lever
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decisions: "iter_0 is shippable for low-stakes use… Not for client-facing where Δ−0.20 on voice would be visible"; "Add improve voice fidelity".
- **what:** The evaluated position on the first adapter: iter_0 matches Claude on JSON validity (20/20 vs 19/20) and schema, comparable length, tiny dimension gaps (relevance −0.25, voice −0.20, integrity −0.10), 4 substantive wins. Verdict: usable for internal tooling and low-stakes sites; voice is the largest gap and the main iter_1 lever (more epochs? lora_r 32? stricter voice-compliant training rows). "Address verbosity" was explicitly dropped (data showed no gap). Fabrication is a both-models problem to solve with prompt-time guardrails or post-hoc verification, not adapter training.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md; working/eval/iter0_eval/iter0_evaluation_report.md#tldr; working/eval/001_test_comparison_with_claude.txt
- **relations:** eval gate; deployment_decision vocabulary; fine-tuning candidates
- **verify-later:** whether any deployment_decision row exists; whether iter_0 was ever served in production

<!-- SOURCE: U06_finetuning.md -->
### CPU-Ollama replay eval attempt and the dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.4d "mistral-small3.1 on a shared cpu-ollama adapter is not a practical substrate… 20 cases × 25-30 min = 10+ hours"; superseded by Thunder GPU eval (HANDOFF 2026-05-08 ran 20 inferences at ~22s/case for ~$0.50).
- **what:** The first flywheel-D attempt replayed prompts against mistral-small3.1 on the shared CPU Ollama adapter; production contention drove one case to 27 minutes (~4 s/token), so a dedicated `ollama-eval` pod (own PVC/service, invisible to production routing because kafka-scheduler only probes ai_endpoint_health entries) was spun up, with the pod-memory sizing rule learned (limit ≥ model file + 8–12GiB headroom). The whole CPU-eval path was then superseded by evaluating the trained adapter on Thunder GPU instances. Yields the durable prediction framework (swap-with-prompt-tweaks vs swap-after-finetuning vs different substrate).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** ai_endpoint_health; Ollama CPU adapter ops; replay eval
- **verify-later:** whether ollama-eval deployment still exists in kustomize/cluster

<!-- SOURCE: U06_finetuning.md -->
### Fine-tuning candidate prioritisation
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §2.6 priority table; flywheel_D_target_selection.sql header ("high volume, high success, short output, structured JSON, low reasoning complexity").
- **what:** Ranked list of agents worth fine-tuning locally: knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher — all high-volume structured-JSON emitters. Explicit non-candidates: page-content-writer long-form (though its iter_0 hero step WAS the first target), visual-design-auditor (judgement), chief-strategist (worth Claude cost). Selection criteria are encoded as a reusable SQL discovery query over llm_call_log volume/recency.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.6; working/flywheel_docs/flywheel_D_target_selection.sql
- **relations:** per-vertical training (vertical column in llm_call_log); model swap
- **verify-later:** current llm_call_log volumes per agent

<!-- SOURCE: U06_finetuning.md -->
### Three improvement channels compound (RAG / LoRA / prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(25) §3, sourced from 009_model_infrastructure decision 10; RAG deployed, LoRA iter_0 trained, prompt_variant column exists but no A/B usage evidenced.
- **what:** The framing that RAG (immediate, no training), LoRA fine-tunes (medium-term cost reduction), and prompt evolution via the `prompt_variant` A/B column are three independent levers that compound: good prompts + good RAG produce the best training data, which produces the best fine-tuned model, which needs good prompts and RAG to perform.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#3
- **relations:** knowledge_base RAG; llm_call_log capture; llm-quality-testing category
- **verify-later:** any prompt_variant A/B analysis in code or docs

<!-- SOURCE: U06_finetuning.md -->
### model_lifecycle schema (training_runs / artefacts / evaluations / deployable_adapters)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 019_model_lifecycle_schema.sql (full DDL with comments); NOTES(45) 2026-06-09(6) confirms `model_lifecycle.training_runs` live with CHECK status pending/running/complete/failed and live rows.
- **what:** The run-lifecycle namespace: `training_runs` (one row per QLoRA run, FK to training_exports.runs, JSONB hyperparameters for reproducibility, loss/VRAM/cost outcome metrics, thunder_instance_id breadcrumb), `artefacts` (adapter binaries decoupled from runs to allow requantisation, storage_uri + sha256 + format), `evaluations` (per artefact × eval_set × judge, JSONB l1/l2 metrics, free-text human deployment_decision), plus `deployable_adapters` view (latest shipped_% adapter per base model — the chassis's read point for "which adapter to load") and `latest_training_run_per_export`. Supersedes the earlier flat `model_training_runs` sketch in FOCUS §2.5.1/HANDOFF 05-08.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#tables-needed; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7
- **relations:** eval gate; thunder_instances (FK training_run_id); mark_training_run_running/terminal actions
- **verify-later:** schema in clients_db; whether deployable_adapters is read by any chassis code

<!-- SOURCE: U06_finetuning.md -->
### Eval gate before promotion (human deployment decision; integrity lives here)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 019 schema docstring "deployment_decision is set by human after review; nullable until then"; PLAN b2 "No upload scheme substitutes for evaluating the adapter"; HANDOFF 05-08 "Auto-deployment NOT included in v1".
- **what:** Adapters never auto-promote: a human reviews flywheel-D output and writes a free-text deployment_decision ('shipped_internal', 'rejected_voice_gap', …); anything `shipped_%` becomes deployable. Critically, the eval gate is also the *integrity* boundary for the hostile-VM upload design — a maliciously-crafted-but-valid adapter written through a legitimate presigned URL is caught by evaluation, not by credentials. The original phase-2 sketch had a conditional auto-swap (`swap_agent_model if score ≥ threshold`) which was walked back to human review.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql#evaluations; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#agents-needed
- **relations:** hostile-VM threat model; model swap/revert; deployable_adapters view
- **verify-later:** whether any evaluation row has a decision; whether swap_agent_model was ever wired to eval output

<!-- SOURCE: U06_finetuning.md -->
### Flywheel C phase-2 automation architecture (HTTP job server → SSH-exec → adapter dispatch)
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.5.1 "HTTP job server (Option B chosen)"; HANDOFF 2026-05-08 "Architecture decision: SSH-exec, not HTTP job server (initially)"; the built system (phase 5) uses thunder-adapter dispatch actions + presigned URLs + detached run.sh.
- **what:** The automation design went through three generations: (1) VM-side HTTP job server (POST /jobs, bearer auth, systemd, TLS) — designed, never built; (2) direct SSH-exec from chassis (simpler at low run frequency, no VM service to maintain) — the pivot decision; (3) the final built shape: chassis dispatch actions publish to thunder-adapter (provision/ssh_exec/presign), data moves only via presigned B2 URLs, training runs detached under run.sh with a separate monitor. Rejected throughout: Kafka consumer on the VM (connectivity + overkill). "Chassis drives, GPU VM serves" is the invariant across all three; each generation supersedes the previous.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5.1; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#architecture-decision; working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** model-trainer chain; training-launcher; presigned data plane
- **verify-later:** no HTTP job server should exist anywhere; adapter dispatch actions in registry.go

<!-- SOURCE: U06_finetuning.md -->
### model-trainer orchestration chain (spawn/call data-preparer → provisioner → launcher)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "model-trainer flow confirmed live: spawn_data_preparer → spawn_provisioner → spawn_launcher → call_data_preparer → call_provisioner → call_launcher → complete".
- **what:** The end-to-end automated training run is the `model-trainer` orchestrator (id 94f5a069): spawns three worker agents up front, then calls them in order. `training-data-preparer` (71ab9361) streams the export to S3 as JSONL and INSERTs the pending training_runs row; `gpu-provisioner` provisions the A100 through thunder-adapter; `training-launcher` (1223bdc1) presigns, writes the manifest, and SSH-launches. Known open bug: a failed call_agent step falls through to the next call instead of aborting the orchestration (produces confusing secondary errors).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#1,#update-2026-06-06; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#where-this-sits,#also-pending
- **relations:** training-launcher workflow; migrations 103/104; call_agent fall-through bug
- **verify-later:** agent_definitions 94f5a069/71ab9361/1223bdc1 in clients_db; the fall-through behaviour in coordinator code

<!-- SOURCE: U06_finetuning.md -->
### training-launcher real workflow (presign → manifest → detached SSH launch → mark_running)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… presign_dataset → presign_scripts → compute_keys → presign_checkpoints (ONE batch await) → presign_final → [check_resume] → assemble_manifest → write_manifest → ssh_exec_launch → mark_running → complete."
- **what:** The launcher replaced a stub (migration 102) with a workflow of dispatch actions cloned from the proven decommission pattern: presign dataset + scripts bundle (GET), compute K checkpoint keys, batch-presign checkpoint PUTs + final PUT, optionally resolve a resume checkpoint, assemble and SSH-place `/workspace/upload_manifest.json`, then launch training detached and flip training_runs pending→running (`mark_training_run_running`, hardcoded guarded transition). Constants live in step config; cross-step values resolve via config dot-paths; the ssh command is built from a `command_template` with `{token}` interpolation. Evolved through migrations 102→105→109/109a/109b→110→111.
- **sources:** working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#5,#update-2026-06-09-3; working/phase5/102_training_launcher_real.sql (header)
- **relations:** batch presign; upload manifest; setsid launch; migrations family
- **verify-later:** live training-launcher default_config (2d state check); registry entries for the dispatch actions

<!-- SOURCE: U06_finetuning.md -->
### setsid detached launch and the detached exit-0 false-success gap
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §4 (the single-line setsid command) and 2026-06-03 ~18:04: "exit_code 0 only because the command's last token is echo (the known detached-ssh_exec false-success)".
- **what:** The adapter's ssh_exec blocks until the remote command exits (5-min timeout), so the launch chain (curl bundle + dataset via presigned URLs, untar, nohup run.sh) runs under `setsid … & echo LAUNCH_PID=$!` — the SSH session returns immediately with the PID. The cost: exit_code 0 only proves the echo ran; VM-side failures inside the detached chain (e.g. the /workspace permission failure) are invisible to the launcher. Corollary lessons: the command_template must stand up its own workspace with sudo mkdir+chown (105/109a), and any best-effort VM step under `set -e` becomes fatal (the root-owned ~/.bashrc append killed a run at the last cosmetic setup step).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#4,#update-2026-06-03; working/phase5/105_launcher_workspace_sudo_mkdir.sql (header); working/phase5/109a_fix_write_manifest_workspace_perm.sql (header)
- **relations:** run.sh markers (the real success signal); training monitor (fills the observation gap)
- **verify-later:** current command_template in the live launcher def

<!-- SOURCE: U06_finetuning.md -->
### run.sh launch chain and RUN_SH_* marker protocol
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** run.sh header ("Emits grep-able RUN_SH_* markers for the future monitor"); NOTES(45) 2026-06-09(4) verified live: "RUN_SH_START → STEP setup → STEP smoke → SMOKE_OK → STEP full_train → RUN_SH_UPLOAD manifest=present".
- **what:** All heavy on-VM work lives in run.sh (setup → smoke train → full train), not in the chassis workflow, so the chain is editable by re-uploading the bundle with no DB migration. It emits a marker protocol to /workspace/train.log (`RUN_SH_START/STEP/SMOKE_OK/FULL_OK/DONE/FATAL`) that is the machine-readable contract for the training monitor's probe. After Phase C, `set -euo pipefail` plus the hard-gated final upload means `RUN_SH_DONE` ⇒ trained AND adapter durable in B2. A mid-train crash leaves no marker (GONE_UNKNOWN).
- **sources:** working/scripts/run.sh (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#run.sh; working/phase5/NOTES_phase5_training_launcher_running(45).md#8
- **relations:** monitor probe classification; scripts bundle as deployment unit; CheckpointUploader
- **verify-later:** run.sh in the live B2 bundle vs the repo copy

<!-- SOURCE: U06_finetuning.md -->
### Scripts bundle in B2 as the training deployment unit
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~19:1x: "re-uploading the object IS the whole deploy"; RUNBOOK(8) §4a flat-bundle verification steps.
- **what:** The on-VM scripts (run.sh, 00_vm_setup.sh, 02_train, 03_inference_test) ship as `finetuning/scripts/bundle.tar.gz` in the personae-model-training bucket; the launcher presigns a GET and the VM curls+untars it. The bundle must be flat (files at archive root). Re-uploading the object deploys new training code — no chassis or DB change — with the corollary that editing a script without re-tarring deploys nothing (byte-identical md5 trap). The agent def holds only the key.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x; working/phase5/UPLOAD_bundle.sh; working/phase5/RUNBOOK_iter0_pretrigger(8).md#4a; working/scripts/README_setup.md
- **relations:** run.sh; presigned data plane; SAVE_STEPS re-pack for fast tests
- **verify-later:** b2 ls of finetuning/scripts/; bundle contents vs working/scripts/

<!-- SOURCE: U06_finetuning.md -->
### Checkpoint & final-adapter durability via pre-minted presigned PUT manifest
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) status header: "Phases A, B, C BUILT… ckpt-0 confirmed in B2 on run 0ac806ab (the upload path is proven end-to-end)"; "Still unproven in prod is one run reaching RUN_SH_DONE with the final adapter.tar.gz durable".
- **what:** The Thunder VM disk is ephemeral and originally nothing moved training output off it (no checkpoints — save_strategy "no" — and the adapter saved only locally, so a reap = total loss and the monitor's DONE_OK→decommission would have destroyed the artefact). The fix: the launcher pre-mints single-object write-only presigned PUT URLs (K checkpoints + 1 final, keyed `finetuning/checkpoints/<run_id>/ckpt-<index>.tar.gz` and `finetuning/artefacts/<run_id>/adapter.tar.gz`) into `/workspace/upload_manifest.json`; the VM uploads through them. Checkpoints are keyed by save-INDEX not Trainer global_step (fragile to predict); write-once with B2 versioning as backstop; URL expiry must exceed max_uptime (expiry_minutes 3000). Checkpoint upload proven (ckpt-0 in B2); the final-adapter upload and a full RUN_SH_DONE run remain the empirical gate.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md; working/phase5/README_what_is_manifest.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-5
- **relations:** hostile-VM threat model; CheckpointUploader; resume; monitor enablement gate; batch presign
- **verify-later:** b2 contents under finetuning/checkpoints/ and artefacts/; migrations 109/110/111 state in the live def

<!-- SOURCE: U06_finetuning.md -->
### CheckpointUploader trainer callback (best-effort checkpoints, hard-gated final upload)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** PLAN(7) §02_train "BUILT 2026-06-05… Tier 1 (box-free) PASSED"; NOTES(45) 2026-06-05 update.
- **what:** `02_train` gained gated flags (`--save-steps`, `--save-total-limit`, `--upload-manifest`; defaults keep old behaviour byte-for-byte). A `CheckpointUploader(TrainerCallback).on_save` tars each checkpoint and PUTs it to its save-index URL synchronously (a background thread was rejected — races save_total_limit's dir deletion); checkpoint upload failure is best-effort (log and continue). The FINAL adapter upload is a hard gate: failure raises → non-zero exit → no RUN_SH_DONE → the monitor never treats the box as cleanly done (degrades to GONE_UNKNOWN→failed, never a false DONE_OK). Content-Type application/octet-stream confirmed accepted by the unbound presigned signature. Two manifests coexist: input upload_manifest.json vs the output run-metadata manifest.json.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#02_train; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-05
- **relations:** run.sh markers; durability manifest; save_steps cadence (50 ≈ one checkpoint/1.5–2h, ~2GB each: adapter + AdamW state)
- **verify-later:** 02_train in the live bundle; checkpoint sizes in B2

<!-- SOURCE: U06_finetuning.md -->
### Resume path (cluster-side checkpoint selection, presence-of-checkpoints as the signal)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) 2026-06-14 update: "batch + resume BUILT, APPLIED, and verified [def-state]… still unproven in prod"; migration 111 applied and 2d-verified.
- **what:** A relaunch for the same training_run_id becomes a continuation automatically: the launcher's `check_resume` step asks the adapter (`prepare_resume_url`) to list `finetuning/checkpoints/<run_id>/` in B2 (reusing the existing `storage.Client.ListObjects` — the presumed "list-keys gap" was wrong), pick the highest ckpt-N, and presign a GET; assemble_manifest emits a `resume` block only when found. 02_train downloads/extracts it and calls `trainer.train(resume_from_checkpoint=True)` (restores optimizer/scheduler/RNG/step). No separate resume mode — empty prefix = fresh start; the launcher owns save-index key allocation across resume launches. found=false is a valid answer; transient list failures return error_recoverable so the chassis retries.
- **sources:** working/phase5/README_what_is_manifest.md; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#resume; working/phase5/111_training_launcher_resume_wiring.sql (header)
- **relations:** durability manifest; monitor GONE_UNKNOWN (total-loss case that motivated this)
- **verify-later:** a real kill-and-resume test; dispatch_thunder_prepare_resume_url in registry

<!-- SOURCE: U06_finetuning.md -->
### Monitor enablement gate: DONE must mean durable
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7): "enable thunder-training-monitor (safe once DONE means durable — gated on the first run actually reaching RUN_SH_DONE)"; NOTES(45) §9 "Not enabled; enabling is RUNBOOK step 6".
- **what:** An explicit sequencing invariant: the monitor's DONE_OK path decommissions the box (destroying the disk), so the schedule stays disabled until the upload path proves that RUN_SH_DONE implies the adapter is in B2. Enabling early would have destroyed iter_0's adapter. The interim protocol for in-flight runs was manual: scp adapter_out off the box before anything decommissions it.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#build-order; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#update-2026-06-04
- **relations:** CheckpointUploader hard gate; run.sh markers; monitor
- **verify-later:** scheduled_tasks.enabled for thunder-training-monitor; whether a run has since reached RUN_SH_DONE

<!-- SOURCE: U06_finetuning.md -->
### knowledge_base RAG store and flywheel B verification
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b action items all [x]; "Flywheel B is done" (step 3 chassis integration COMPLETED on v1.0.979, 2026-04-21).
- **what:** `knowledge_base` (migration 082): pgvector(768) for nomic-embed-text, shared across agents via `rag_lookup`/`rag_index` actions, trigram fallback when Ollama is down, SHA256 dedup, metadata-first filtering doctrine (filter by vertical/component_type/content_type/source before ranking by similarity — else a vet example surfaces for gas-wholesale copy). Verified bottom-up in three single-focus steps: pgvector+ivfflat+cosine on synthetic vectors, real-content retrieval through cpu-ollama, then chassis integration via a deterministic 3-step `rag-test-agent`. Nomic v1 judged good enough; v2-moe named as a drop-in upgrade (same 768 dims). Open task: periodic REINDEX of the ivfflat index.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.2,#2.4b; working/flywheel_docs/flywheel_B_step*.{sql,sh} (headers)
- **relations:** nomic prefixes; RAG-platform product; three channels
- **verify-later:** rag_actions.go; knowledge_base row counts; whether REINDEX ever scheduled

<!-- SOURCE: U06_finetuning.md -->
### Nomic task prefixes are load-bearing (rag_actions prefix patch)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b "[x] Prefix patch deployed and verified live 2026-04-21… log line 'prefix_applied':true observed".
- **what:** Without `search_document:`/`search_query:` task prefixes, nomic embeddings ranked a Labrador chunk above the French Bulldog chunk on a BOAS-specific query; with prefixes the correct result won with 5× the margin. The patch adds a model-scoped `applyNomicPrefix` helper (only nomic-embed-*, double-prefix guard, prefix_applied logged) at both embed call sites; stored chunks and dedup hashes stay unprefixed; trigram fallback untouched.
- **sources:** working/flywheel_docs/PATCH_rag_actions_nomic_prefixes.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4b
- **relations:** knowledge_base RAG; Ollama CPU adapter
- **verify-later:** rag_actions.go contains applyNomicPrefix

<!-- SOURCE: U12_docs024_archives.md -->
### Quality improvement flywheel (RAG + LoRA + prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** Live `009_model_infrastructure.md` "Future": RAG actions "registered, not workflow-tested"; LoRA training pipeline and training-data export from `llm_call_log` both still open.
- **what:** Three independently-valuable, compounding improvement channels: RAG (inject retrieved good examples at call time), LoRA (retrain periodically on filtered successful outputs), and deliberate prompt A/B testing (80/20 traffic split, promote on audit-success-rate). A `training-orchestrator` workflow packages LoRA training as an adapter-driven workflow (export → start_gpu_instance → train → evaluate → deploy_or_reject → stop_gpu_instance → log). A scraped-data "AI slop" quality gate filters what may enter the training set.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Quality Improvement Flywheel", #"Scraped Data Quality Gate (AI Slop Prevention)"; docs024_key_docs_latest/009_model_infrastructure.md#"Future"
- **relations:** GPU/AI-endpoint scheduling; llm_call_log flywheel columns
- **verify-later:** whether any `training_runs` completed beyond the one noted in 009; whether RAG actions are workflow-exercised.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fine-tuning flywheel (call-log → LoRA → GGUF → Ollama)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** 001(0) "The training data pipeline: LLM call logging → export → LoRA fine-tune on GPU → GGUF export → load into Ollama → update agent definition to provider: ollama"; "Not yet built (future work)"
- **what:** A path to replace short-output classification/extraction agents with local fine-tuned models: accumulate 200+ successful `llm_call_log` examples, export Alpaca/ChatML, LoRA fine-tune on GPU (unsloth), export GGUF into Ollama, flip the agent definition to `provider: ollama`, then A/B test against Claude.
- **sources:** WM/001_development_guide(0).md#fine-tuning-path, WM/001_development_guide(0).md#implementation-status-llm-optimization, WM/033_thunder_adapter_design.md#tldr
- **relations:** LLM infrastructure; Thunder adapter; LLM tiering
- **verify-later:** training_data_export.sql; model_lifecycle.training_runs; unsloth

<!-- SOURCE: U18_sql_for_agents.md -->
### Finetuning flywheel Phase 5: training kickoff orchestration (model-trainer, training-data-preparer, gpu-provisioner, training-launcher)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 109/110 real; 111/112 explicitly STUBs "to unblock end-to-end testing... will be replaced by a real implementation in a future migration"; 116/117 monitor real running instances, implying provisioning later became real.
- **what:** model-trainer owns the KICKOFF phase: training-data-preparer exports a training_exports snapshot as JSONL to S3 and INSERTs the model_lifecycle.training_runs row (pending); gpu-provisioner calls Thunder Compute API for an A100, stores the SSH key as a k8s secret; training-launcher SCPs scripts/dataset and nohup-launches training, returning the pid. The workflow exits immediately — completion is deliberately handled by a separate scheduled monitor so no orchestration holds open for ~9 hours. Full hyperparameter set captured for reproducibility.
- **sources:** 109_model_trainer_orchestrator.sql; 110_training_data_preparer.sql; 111_gpu_provisioner_thunder.sql; 112_training_launcher.sql
- **relations:** training-data-exporter (106) upstream; thunder monitor/reaper downstream; thunder-adapter
- **verify-later:** real gpu-provisioner/training-launcher implementations vs stubs

<!-- SOURCE: U18_sql_for_agents.md -->
### Thunder instance lifecycle: reaper + training monitor (orchestrator/worker)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 114 (reaper, every 15 min, idempotent decommission); 116/117 with verified coordinator internals and the insert-DISABLED-until-actions-deploy discipline.
- **what:** Cost/safety controls for rented GPUs: thunder-reaper decommissions instances past max_uptime_hours (one per tick, pre_query LIMIT 1); thunder-training-monitor orchestrator finds every running training instance each tick and spawns a per-instance worker that probes via SSH, classifies (alive / unreachable-streak / done_ok / done_fail), reconciles training_runs and decommissions. 117 records WHY orchestrator-with-loop beats the reaper's scheduler-pre_query shape (must visit every instance, not just the top row) and why the loop must stay sequential (topic reuse safety).
- **sources:** 114_thunder_reaper.sql; 116_thunder_training_monitor_worker.sql; 117_thunder_training_monitor_orchestrator.sql
- **relations:** scheduler-and-tasks (pre_query dispatch patterns); thunder adapter; model-trainer
- **verify-later:** scheduled_tasks rows enabled; probe/classify actions

<!-- SOURCE: U18_sql_for_agents.md -->
### training-data-exporter (llm_call_log → ChatML JSONL)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 106 definition with concrete kubectl retrieval instructions and input payload example.
- **what:** Deterministic single-action agent exporting successful LLM calls from llm_call_log as NDJSON training data in ChatML + metadata format, filterable by agent_type/step/model, with fenced-output and strict-JSON options.
- **sources:** 106_training_data_exporter.sql; 040_optimise_which_llms.sql (llm_call_log)
- **relations:** training-data-preparer consumes exports; flywheel columns from 085
- **verify-later:** training_data_export action; training_exports schema

<!-- SOURCE: U19_sql_tables_components.md -->
### llm_call_log training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Migration 081 Part 3 + schema fixes (agent_id added, nullability relaxed to match Go's nullIfEmpty); export queries reference populated columns incl. work_item_id and vertical.
- **what:** Every LLM call logged with caller identity (agent_type/step/orchestration), model + model_resolved + provider, full prompt_template/prompt_rendered/response_text, token/latency usage and outcome — explicitly designed for training export. Export recipes produce JSONL per task (analyze_tool, recreate_tool, site classification, content writing) with quality filters joining site_work_items outcomes (only export calls whose work item completed), and per-vertical readiness counts.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql
- **relations:** training_exports (successor storage); site_chat_turns (deliberately separate); model upgrades.
- **verify-later:** logging middleware in aiservice; work_item_id/vertical columns present.

<!-- SOURCE: U19_sql_tables_components.md -->
### training_exports Postgres-backed datasets (flywheel A v3)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Schema with rationale "JSONL files landed on ephemeral chassis pods and vanished on restart"; dedup unique index on (export_id, metadata->>'source_log_id').
- **what:** Named, versioned training datasets in Postgres instead of ephemeral JSONL: runs (one per export — filter criteria matching llm_call_log columns, counts, skip reasons, format 'chatml', size, provenance) and rows (ChatML messages + metadata JSONB, ordered by row_index, CASCADE delete). Training-time extraction via \copy in export order. Schema named training_exports specifically to avoid confusion with the model-training pipeline (flywheel C).
- **sources:** docs/agent_docs/sql_for_tables/039_training_exports.sql
- **relations:** llm_call_log source; thunder training runs (flywheel C).
- **verify-later:** exporter action writing runs/rows.

<!-- SOURCE: U22_recent_small_docs.md -->
### llm_call_log (build-time training flywheel)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "The llm_call_log table is capturing every LLM call from every agent ... This started logging the moment you deployed the new chassis image" vs handoff listing the Go patches as ready-but-not-committed.
- **what:** A table capturing every `execute_llm_prompt` call (agent_type, step, model, rendered prompt, response, input/output tokens, latency, success) via a fire-and-forget goroutine logger. Feeds cost/latency analytics (`llm_call_stats` view) and accumulates toward the 200+-examples-per-agent fine-tuning threshold. Cleanup function exists but nothing calls it (table-bloat risk flagged, ~1GB/month).
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../005_PATCHES.md#patch-01-02, docs020.../001_rag_agent_distribution_architecture.md#item-2
- **relations:** anthropic.go usage capture patch, site_chat_turns (deliberately separate log), LoRA training data export
- **verify-later:** llm_call_log table + cleanup_old_llm_logs; LogLLMCall in ai_actions.go; anthropic.go __usage_input_tokens write-back

<!-- SOURCE: U22_recent_small_docs.md -->
### Fine-tuning pipeline (LoRA flywheel, deferred)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Phase 7: Fine-Tuning Pipeline (Deferred) ... becomes relevant after 200+ successful examples per agent type accumulate."
- **what:** The deferred end of the flywheel: once llm_call_log has 200+ examples/agent, export JSONL training data, QLoRA fine-tune a 7B via Unsloth on rented GPU, export GGUF, load into Ollama (`ollama create`), and switch the agent definition to `provider: ollama`. First candidates: site-classifier (high volume, short output), then domain-research-classifier, then the vertical knowledge extractor. Purpose: drive per-call inference cost to ~zero.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-7, docs023.../018_canine_biology.md#6, docs020.../010_simple_explanation.md
- **relations:** llm_call_log, Ollama provider, canine biology text LoRA, self-hosted LLM inference
- **verify-later:** training-data export queries; any GGUF/Ollama custom models; agent_definitions using provider:ollama

<!-- SOURCE: U22_recent_small_docs.md -->
### Text LoRA — veterinary knowledge extractor
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** Phase E todo "Text LoRA fine-tuning (week 6-7)" unchecked; full Unsloth/QLoRA recipe given as instructions.
- **what:** A concrete recipe to fine-tune a local 7-8B model (Unsloth QLoRA, r=16, 3 epochs) on accumulated knowledge-extraction examples, export Q4_K_M GGUF, load into Ollama, and swap `knowledge-extractor` to the local model to eliminate Claude API cost per extraction. Training data accrues naturally during the canine research phase (50 breeds + 30 conditions + 40 procedures ≈ 120, need 200+).
- **sources:** docs023.../018_canine_biology.md#6
- **relations:** fine-tuning pipeline, llm_call_log, Ollama provider
- **verify-later:** vet-extractor GGUF/Ollama model; knowledge-extractor agent provider

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Quality improvement flywheel (RAG + LoRA + prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 020c "Quality Improvement Flywheel": three channels (RAG more examples, LoRA retrain on best outputs, deliberate A/B prompt evolution with `prompt_config.active_variant`/`testing_allocation:0.2`); llm_call_log columns `work_item_id/prompt_variant/vertical/rag_context_used` "Not Yet Deployed"; LoRA training as `training-orchestrator` workflow (export→gpu→train→evaluate→deploy_or_reject).
- **what:** Compounding quality via three independent channels feeding site production: knowledge collection (scrape + successful Claude outputs + audit insights, quality-gated) → RAG injection, LoRA retraining (Unsloth QLoRA on ThunderCompute), and deliberate prompt A/B testing scored by audit success rate. llm_call_log gains flywheel columns; measurable metrics (first-pass success, rewrite count, lock rate) precede traffic metrics.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#quality-improvement-flywheel; #prompt-evolution
- **relations:** live FOCUS_finetuning_flywheel_and_service; RAG best practices; canine biology LoRA; per-vertical LoRA
- **verify-later:** llm_call_log flywheel columns; training-orchestrator workflow; training_runs table

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Internal AI training flywheel (A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §1 table "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (2026-04-23)
- **what:** The core internal loop: the site-building pipeline logs every LLM call as a byproduct; that data periodically fine-tunes local models that are swapped in for Claude calls where quality holds, dropping API cost. Four lanes — A (data export), B (RAG), C (LoRA training), D (Claude-vs-local eval). A and B were done; C scripted; D paused on infra contention.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#1, #2, #4
- **relations:** parents Flywheel A export, Flywheel B RAG, Flywheel C fine-tuning, Flywheel D eval; feeds Phase 5 launcher; three improvement channels
- **verify-later:** llm_call_log, knowledge_base, training_exports schema, ai_endpoint_health

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel A — training-data export pipeline
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4i "First real training dataset now in Postgres: export_id fef7be6b-…, 1,958 rows, 21.2MB" and "Spawning architecture fully validated" (2026-04-23)
- **what:** A chassis action (`training_data_export`) + `training-data-exporter` specialist agent (wrapped by `training-data-export-orchestrator`) that reads `llm_call_log`, strips markdown code fences via `stripMarkdownFromResponse`, validates JSON, and writes ChatML training rows. Evolved v1 (static file config, superseded) → v2 (reads params.CollectedData["input_data"], file output to /tmp) → v3/v3.1/v3.2 (writes to a dedicated `training_exports` Postgres schema, per-batch transactions to survive pgbouncer).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** produces datasets for Flywheel C; superseded intermediate: v1 file-output export; feeds training_exports schema concept
- **verify-later:** platform/orchestration/actions/ training_data_export_v3.go; training_exports.runs, training_exports.rows; agent_definitions training-data-exporter

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM call logging (llm_call_log) as ops visibility + training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 001(5): "Verified in production (March 2026) — 57+ rows"; 085 migration adds flywheel columns
- **what:** Every execute_llm_prompt call logged fire-and-forget (agent_type, step, model, rendered prompt, response, tokens, latency, `__sent_temperature`/`__sent_max_tokens` write-backs). Flywheel columns (work_item_id, prompt_variant, vertical, rag_context_used) link calls to outcomes for LoRA/RAG training exports. Known past bugs: schema/Go column drift (agent_id vs client_id), empty agent_type from buildActionParams.
- **sources:** 001(5)#LLM call logging, #Implementation Status; 022_ai_endpoint_health_and_flywheel_llm_call_log.sql
- **relations:** batch queue LogLLMCall paths; fine-tuning path
- **verify-later:** llm_call_log schema; llm_call_logger.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fine-tuning path (log → export → LoRA → GGUF → Ollama → swap)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 009 (2026-07-10 update): one training run `complete`, GPU provisioning via Thunder real and dynamic, but "No agent_definitions row currently points ai_service at llama3.3:70b — trained and tested, never used for production inference"
- **what:** Pipeline: accumulate 200+ examples in llm_call_log → export (Alpaca/ChatML) → LoRA fine-tune on GPU (unsloth) → GGUF → Ollama → swap agent definition → A/B against Claude. Candidates are short-output classifiers (site-classifier, vet-practice-verifier, etc.). The last mile (wiring the trained 70B into live inference) is explicitly outstanding.
- **sources:** 001(5)#Fine-tuning path; 009#Future incl. 2026-07-10 note; 023 (manual A/B comparisons)
- **relations:** model swap functions; Thunder adapter; drain mode
- **verify-later:** model_lifecycle.training_runs; agent_definitions ai_service providers

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM fallback extraction doubling as training data (med pricing)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 008: "response logged to llm_call_log… dual purpose: price extraction now, training data for future LoRA"
- **what:** Regex handles ~90% of pages; CPU Mistral (mistral-small3.1, temp 0.1) parses the remainder into a JSON variant array at 80-280s/call — acceptable because batch-tolerant — while accumulating markdown→JSON pairs for a future fine-tune. Future: product matching across retailers, price alerts from snapshot history, affiliate-feed switch.
- **sources:** 008#LLM Fallback, #Future Work
- **relations:** batch processing categorisation; llm_call_log flywheel
- **verify-later:** llm_call_log provider='ollama' step_name='scrape_prices'

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Thunder adapter (credential-boundary GPU provisioning with caps and reaper)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 033 confirmed decisions; 016 §9 entries show it running in prod (provision 400s, ssh findings, reaper window maths)
- **what:** All Thunder Compute interaction routes through a long-lived cluster adapter holding THUNDER_COMPUTE_API_KEY/B2 keys/ephemeral SSH keys; VMs are per-run ephemeral and credential-free (presigned URLs only; compromise blast radius = that run's files). Caps: $100/day rolling, 18h hard uptime, concurrency 2, 15-min reaper reconciling API↔thunder_instances. Formally retracts the on-VM HTTP job-server option. Operational lessons: lowercase gpu_type enums, template 'base', OpenAPI examples aren't valid values, tnr connect does server-side setup, login user ubuntu not root, live-instance-scoped partial unique index for recycled provider ids.
- **sources:** 033 full; 016 §9 Thunder entries
- **relations:** batch drain mode; training launcher/monitor
- **verify-later:** thunder-adapter deployment; reaper task

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Finetuning flywheel four-lane programme (A export, B RAG, C training, D eval)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (doc last touched 2026-04-23)
- **what:** The internal AI training flywheel: production LLM calls become training data (A), verified knowledge feeds RAG (B), local models get fine-tuned on the exported data (C), and quality is compared against Claude (D), so local models replace API calls where quality holds and costs drop.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#1, #2, #15-changelog
- **relations:** llm_call_log capture; training-data export pipeline; Flywheel C QLoRA; Flywheel D replay-eval; three compounding improvement channels
- **verify-later:** `training_exports` schema; `flywheel_A_v3/`, `flywheel_C/` script dirs; `llm_call_log`; whether a first GPU training run ever happened after 2026-04-23

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### llm_call_log LLM call capture
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "llm_call_log populating (flywheel columns wired through ai_actions.go)" [x]; but "Cleanup function exists (cleanup_old_llm_logs) but nothing schedules it yet" (2026-04-23)
- **what:** Every LLM call writes agent_type, step_name, model, prompt_rendered, response_text, tokens, latency, success, work_item_id, prompt_variant, vertical, rag_context_used to `llm_call_log` (migrations 081/085) via a fire-and-forget `LogLLMCall` goroutine. Retention 90/180 days; the join key for "what to train".
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.1, #4.2
- **relations:** training-data export; prompt evolution (prompt_variant A/B); vision cost tagging (PLAN_imagery_loop_closure 5.4)
- **verify-later:** `ai_actions.go` LogLLMCall; scheduling of `cleanup_old_llm_logs`; agent_type population on old rows

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### knowledge_base RAG with load-bearing nomic task prefixes
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "Prefix patch deployed and verified live 2026-04-21 … log line \"prefix_applied\":true observed on rag_lookup step"; "Flywheel B is done" (chassis v1.0.979)
- **what:** `knowledge_base` table (migration 082, pgvector 768-dim, ivfflat+cosine) readable/writable by any agent via `rag_lookup`/`rag_index`, with trigram fallback when Ollama is down. Empirically established that `search_document:`/`search_query:` prefixes for nomic-embed-text are mandatory for correct ranking; metadata-filter-first-then-similarity is the retrieval rule.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.2, #2.4b
- **relations:** finetuning.uk RAG product reuses this infra; cpu-ollama adapter
- **verify-later:** `platform/orchestration/actions/rag_actions.go` applyNomicPrefix; `knowledge_base` metadata fields (vertical, component_type, source_quality)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Training-data export as chassis agent + action, Postgres-backed
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "First real training dataset now in Postgres: export_id fef7be6b…, 1,958 rows, 21.2MB, reconciled manually. Spawning architecture fully validated" (2026-04-23); v3.2 strict-UPDATE code "awaiting chassis rebuild/deploy" at doc date
- **what:** Export of (prompt, response) pairs from llm_call_log as ChatML messages with metadata sidecar (source_log_id, agent_type, step_name, export_version), fence-stripped and JSON-validated, via `training_data_export` action + `training-data-exporter` worker + `training-data-export-orchestrator` wrapper. v3 writes named snapshot datasets into `training_exports.runs`/`training_exports.rows` (rejected: real-time streaming — batch snapshots preserve "the dataset we trained on"). Evolved v1 (template config, failed) → v2 (file output, wrong pod) → v3 (Postgres, dedicated pod).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** orchestrator wrapper pattern; pgbouncer per-batch commits; negative-examples/DPO decision
- **verify-later:** `training_exports` schema in DB; `training_data_export_v3.go` and registry entry; whether v3.2 landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel C QLoRA training pipeline (Unsloth, Llama 3.3 70B)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Pipeline shape now concrete, scripts written, awaiting first training run on a GPU VM" (2026-04-23)
- **what:** Five scripts (`flywheel_C/00-03` + README) pull a named export from Postgres, train a LoRA adapter on Llama 3.3 70B Instruct via Unsloth QLoRA on a single H100/A100, sanity-check JSON validity of outputs. 70B chosen because hardware was available; 8B flagged as likely 95% of quality at 10% cost — comparison run planned.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5
- **relations:** Flywheel C phase 2 automation; Flywheel D eval harness scores the result
- **verify-later:** existence of `flywheel_C/` scripts; any `model_training_runs` table or LoRA artefacts

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel C phase 2 automation (chassis drives, GPU VM serves)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "design locked, not built … Gated on phase 1 (manual training run)" (2026-04-23)
- **what:** HTTP job server (~200 lines, POST /jobs → poll → fetch adapter, bearer auth) on the GPU VM; three new chassis components (model-trainer, model-evaluator specialists, training-flywheel-orchestrator wrapper chaining export→train→eval→conditional swap_agent_model); three new tables (model_training_runs, model_artefacts, model_evaluations). Rejected SSH-exec and Kafka-consumer-on-VM alternatives.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5.1
- **relations:** model swap/revert functions; scheduled_tasks trigger option
- **verify-later:** whether any of the three agents/tables exist in agent_definitions / DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flywheel D replay-eval methodology + dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel D (eval) paused"; ollama-eval pod deployed with memory bump "requests: 24Gi / limits: 28Gi … fix persisted into kustomize base" (2026-04-23)
- **what:** Replay-don't-re-run evaluation: pull 20 diverse stored Claude prompts (DISTINCT ON orchestration_id), POST to candidate model via /api/chat, compare with 3 levels (structural jq checks → Claude-as-judge → manual review). Shared cpu-ollama contention made eval impractical (27 min/case), so a dedicated `ollama-eval` deployment (own PVC/service, invisible to production routing) was created. Memory rule: pod limit ≥ model file size + 8–12 GiB.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4d, #2.4d-comparison
- **relations:** Flywheel C evaluation preconditions; Ollama CPU ops envelope
- **verify-later:** `deployments/kustomize/services/ollama-eval/`; results.jsonl outcome; whether eval ever completed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Negative examples reserved for DPO, excluded from SFT
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "For our first training run: plain SFT, edge cases excluded" (decision 2026-04-22)
- **what:** Edge-case rows where Claude correctly produced prose instead of JSON are positive examples of the wrong shape for SFT and must be excluded from exports; they become the "rejected" side of DPO preference pairs later. Keeps them in llm_call_log, out of training_exports.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e "Negative examples / edge cases"
- **relations:** training-data export filters (strict_json)
- **verify-later:** export action's strict_json / prose-exclusion behaviour

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Three compounding improvement channels (RAG, LoRA, prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "From 009_model_infrastructure.md, decision 10" — RAG immediate, LoRA medium-term, prompt_variant A/B ongoing (2026-04)
- **what:** Three ways local model quality improves, each useful alone but compounding: RAG injects verified knowledge now; LoRA replicates a task cheaply later; `prompt_variant` column enables prompt A/B evolution. Good prompts + good RAG produce the best training data.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#3
- **relations:** flywheel programme; llm_call_log prompt_variant
- **verify-later:** any actual prompt_variant usage in llm_call_log

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Fine-tune candidate selection heuristic
- **category:** finetuning-flywheel
- **status-signal:** unknown
- **status-evidence:** Priority table (2026-04): knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher good; page-content-writer/visual-design-auditor/chief-strategist "not good candidates"
- **what:** Agents with high-volume, structured-JSON, short outputs are swap candidates for local models; long creative output and judgement tasks stay on Claude. Drives which agent/step gets exported and trained first.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.6
- **relations:** Flywheel D target choice (page-content-writer iter_0 chosen for eval despite being a poor swap candidate, because it had the most logged data)
- **verify-later:** actual llm_call_log volumes per agent

<!-- SOURCE: U04_idea_uk.md -->
### Training checkpoint/adapter upload to B2 via pre-minted presigned single-object PUTs
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Status: Phase A (02_train) BUILT 2026-06-05, isolation test pending. Phases B–D not built." — with the isolation-test harness and its artefacts present in nginx/_iso/.
- **what:** Solves three coupled gaps of ephemeral Thunder training VMs (final adapter not durable; the monitor's DONE_OK path decommissions the disk; no checkpoints/resume on a ~24h run): the launcher pre-mints single-object, write-only presigned PUT URLs and hands them to the VM in a manifest — the hostile-VM threat model rejects standing scoped keys and callback endpoints (nothing on the box can mint or write beyond the fixed URL set). Explicit framing: this protects access, not artefact integrity — a malicious-but-valid adapter still needs the flywheel-D eval gate before promotion. Phase A validated box-free by isolation_test_phase_a.py (presign/PUT signature, tar round-trip, checkpoints/ exclusion, GET+extract byte-identical).
- **sources:** idea.uk/nginx/PLAN_checkpoint_and_artefact_upload_b2(1).md; idea.uk/nginx/isolation_test_phase_a.py (header); idea.uk/nginx/README_get_b2_details.md
- **relations:** Thunder adapter; B2 dead-drop (same one-way pattern); model-infrastructure eval gate.
- **verify-later:** 02_train upload hooks; thunder-training-monitor-worker decommission path.

<!-- SOURCE: U06_finetuning.md -->
### Finetuning flywheel programme (lanes A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 state block: "Flywheel A ✓ operational / B ✓ done / C phase 1 ✓ / C phase 2 → next priority / D ✓ iter_0 evaluated"; later phase-5 notes carry C-phase-2 to a mostly-green automated launch (2026-06-09) with the monitor still disabled.
- **what:** The internal self-improvement programme: the site-building pipeline emits training data as a byproduct (A: export), RAG injects verified knowledge immediately (B), local models are periodically fine-tuned on the captured data (C), and evaluated against Claude before any swap (D). The strategic goal is dropping API cost by swapping local models in for Claude calls where quality holds.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#1-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#current-state; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** every other concept in this unit; three improvement channels; model swap/revert
- **verify-later:** training_exports schema, model_lifecycle schema, llm_call_log columns, thunder-adapter deployment, scheduled_tasks rows for reaper/monitor

<!-- SOURCE: U06_finetuning.md -->
### llm_call_log training-data capture
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.1 "Every LLM call in the system writes to llm_call_log (migration 081, flywheel columns added in 085)"; §4.1 checkbox "[x] llm_call_log populating".
- **what:** Every LLM call logs agent_type/step_name (the "what to train" join key), model/provider, prompt_template/prompt_rendered/response_text (the raw training pair), token/latency/cost signals, success/error, work_item_id, prompt_variant and vertical. Write path is `LogLLMCall` in ai_actions.go — fire-and-forget goroutine, 5s timeout, never blocks the workflow. Retention 90d success / 180d error via `cleanup_old_llm_logs`, which exists but is NOT scheduled (open task). Historically-empty `agent_type` rows exist; recent writes are 100% populated.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.1,#2.4c,#4.2
- **relations:** training-data export agent; replay eval (pulls stored prompts); per-vertical slicing; prompt evolution (prompt_variant)
- **verify-later:** migrations 081/085; platform actions ai_actions.go LogLLMCall; cleanup_old_llm_logs scheduling; llm_call_log row counts

<!-- SOURCE: U06_finetuning.md -->
### Training-data export as chassis agent + action (flywheel A, v1→v3.2)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23: "training_data_export.go v3.2 … deployed as v3.2, verified working"; two 1,958-row exports landed in Postgres.
- **what:** Training export is a first-class chassis pipeline component, not ad-hoc SQL: a `training_data_export` action plus `training-data-exporter` worker and `training-data-export-orchestrator` wrapper. It queries llm_call_log, strips markdown fences via the shared `stripMarkdownFromResponse`, validates JSON per row, and writes batched inserts into `training_exports`. The v1→v3.2 evolution encodes several platform lessons (template-config not rendered for deterministic actions; file output on ephemeral pods; pgbouncer transaction limits; RowsAffected checks).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** training_exports schema; orchestrator wrapper spawning pattern; pgbouncer per-batch commits
- **verify-later:** platform/orchestration/actions/training_data_export.go; agent_definitions rows training-data-exporter / training-data-export-orchestrator; training_exports.runs contents

<!-- SOURCE: U06_finetuning.md -->
### training_exports Postgres schema (named snapshot datasets)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23 "Schema (applied, verified)"; export_ids 146a9a12 and fef7be6b each hold 1,958 rows.
- **what:** `training_exports.runs` (one row per export with filters, counts, completed_at) + `training_exports.rows` (ChatML messages + metadata JSONB per training record, unique on (export_id, source_log_id)). Datasets live in Postgres, not files or S3 — named, versioned snapshots referenced by export_id UUID, streamed out to JSONL at training time. Real-time streaming into the table was explicitly rejected to keep snapshot boundaries and avoid coupling observability to training. Known trap: `runs.rows_exported` can disagree with actual `rows` content (export a8484922 recorded 1957 but holds 0 rows) — always verify with a count before launching.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#done-verified; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06-2
- **relations:** training-data export agent; training-data-preparer (streams export to S3); dataset pull scripts
- **verify-later:** training_exports schema in clients_db; the a8484922 zero-row anomaly; recent_runs view

<!-- SOURCE: U06_finetuning.md -->
### ChatML export format with metadata sidecar
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22."
- **what:** Training rows are `{messages:[{role:user},{role:assistant}], metadata:{source_log_id, agent_type, step_name, orchestration_id, model, created_at, export_version}}`. Chosen for chat-tuned base-model parity, trainer-tool defaults (Unsloth/Axolotl), and `/api/chat` training-inference parity. Metadata gives row-level traceability back to llm_call_log; export_version future-proofs format evolution. Whole prompt_rendered goes in the user turn (no system/user split yet).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e; working/phase1/files/flywheel_A_export_page_content_writer_iter0.sql (header)
- **relations:** response cleaning; training_exports schema
- **verify-later:** export SQL/action output shape vs current llm_call_log columns

<!-- SOURCE: U06_finetuning.md -->
### Response cleaning and SFT negative-example exclusion
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e iter_0 dataset audit table (97.4% clean JSON, 2.6% fenced, fences stripped on export); "For our first training run: plain SFT, edge cases excluded".
- **what:** Exports must strip markdown code fences (else the fine-tune learns to emit fences, exactly what prompts forbid) and exclude edge-case prose responses: plain SFT has no "don't do this" signal, so intelligent edge-case answers are positive examples of the wrong shape. Those rows stay in llm_call_log as future "rejected" halves of DPO preference pairs — DPO/RLHF is the named future home for negative examples.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e
- **relations:** ChatML export format; `<no value>` contamination
- **verify-later:** stripMarkdownFromResponse; whether any DPO work exists later

<!-- SOURCE: U06_finetuning.md -->
### `<no value>` training-data contamination and the iter_1 filter floor
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-07 known issues: "Training rows from before that fix include the literal token `<no value>`. iter_1's export should filter created_at >= <fix_date>… Action: note the fix-deploy date"; still listed as needed in HANDOFF 2026-05-08.
- **what:** A prompt-builder rendering bug injected the literal token `<no value>` into production prompts; iter_0's training data (and its eval cases) inherit it. The fix-deploy date becomes the created_at filter floor for the iter_1 export. As of the last docs, the date had not been recorded — an open data-hygiene debt.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#known-issues; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#revised-iter_1-priorities
- **relations:** training_exports; held-out eval set (same artefact present)
- **verify-later:** git log for the prompt-builder fix; whether any iter_1 export exists with the filter

<!-- SOURCE: U06_finetuning.md -->
### Dataset profile and schema heterogeneity of page-content-writer iter_0
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4g dataset profile tables (n=1,957; p50 prompt 8,250 chars; three dominant JSON schemas: hero-with-CTAs 68%, minimal hero 18%, header/nav 9%).
- **what:** One (agent_type, step_name) training slice actually spans three component output schemas; the model must learn schema selection conditioned on the "Component: X" text in the prompt. A first-pass option to filter to the top-2 schemas (86% of rows) was noted but the full set was trained. Prompt/response size distribution anchors max_seq choice (some prompts approach 4,000 tokens → max_seq 4096).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/flywheel_docs/terminology.md#seq-4096
- **relations:** Llama 70B QLoRA config; inference-test success criteria (keys match trained schemas)
- **verify-later:** training_exports.rows key distribution query

<!-- SOURCE: U06_finetuning.md -->
### Flywheel C manual training pipeline (00–03 scripts, smoke-gates-full)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07: "executed flywheel C phase 1 end-to-end on a Thunder Compute A100 80GB… iter_0 adapter (791MB) … 5/5 valid JSON".
- **what:** Four scripts define the training path: `00_vm_setup.sh` (idempotent pinned env), `01_pull_dataset_from_postgres.sh`, `02_train_llama_3_3_70b.py` (Unsloth QLoRA, CLI-configurable, emits manifest), `03_inference_test.py` (JSON validity + schema keys sanity). Discipline: a 20-row/1-epoch smoke train and smoke inference always gate the full run — cheap insurance on unattended runs. The same bundle later becomes the automated launcher's payload.
- **sources:** working/flywheel_docs/README.md; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#flywheel-c; working/scripts/run.sh (header)
- **relations:** iter_0 baseline run; scripts bundle as deployment unit; run.sh markers
- **verify-later:** working/scripts/* against the bundle actually in B2 (finetuning/scripts/bundle.tar.gz)

<!-- SOURCE: U06_finetuning.md -->
### Base-model decision: Llama 3.3 70B Instruct QLoRA (with 8B ablation planned)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.5 "Base model: Llama 3.3 70B Instruct. Decision taken 2026-04-23"; HANDOFF 2026-04-23 decisions-locked list.
- **what:** `unsloth/Llama-3.3-70B-Instruct-bnb-4bit` via Unsloth QLoRA on a single A100/H100 80GB; defaults 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096. 70B was chosen because hardware was available and a strong baseline is useful — with an explicit acknowledgment that 8B likely delivers ~95% of quality at ~10% of inference cost for this narrow structured-JSON task; a same-dataset 8B comparison run is planned but never executed in these docs.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#base-model-decision
- **relations:** iter_0 baseline; cost anchors; epochs ablation
- **verify-later:** 02_train defaults; any 8B run in model_lifecycle.training_runs

<!-- SOURCE: U06_finetuning.md -->
### iter_0 baseline training run (real cost/time/loss anchors)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 run summary: 33,136s ≈ 9h12m, final_loss 0.2669 (trailing), peak VRAM 44.8GB, ~$20 total; "Anchor future estimates against $20/iter".
- **what:** The first real fine-tune: 1,958 rows → 1,934 effective, clean loss curve (ep1 1.49→0.27, ep3 →0.10), adapter 791MB fp32 safetensors. Epoch-3 loss gap suggests memorisation → a 2-epoch ablation is queued for iter_1. Cost anchor $20/iteration (training) + $1.50 (eval) ≈ $22/cycle. Later automated runs corrected the wall-time estimate: the full run is ~24h at ~119s/step without FA2, not the "30–90 min" the README claimed.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md; working/phase2/README.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** version pinning; snapshot economics; per-instance uptime bump; fp16 save decision
- **verify-later:** lora_iter0_full/manifest.json; model_lifecycle.training_runs rows 1cd65dd7/e6ab9fad

<!-- SOURCE: U06_finetuning.md -->
### GPU environment version pinning (cu124 stack)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Version pin discoveries (essential for any future run)" table with per-pin rationale.
- **what:** The working training environment is a narrow pin set: torch 2.6.0+cu124, transformers<5, torchao<0.17 (transformers imports torchao eagerly; incompatible torchao breaks import entirely), prebuilt flash-attn wheel (Thunder's Ollama template ships CUDA runtime, no nvcc), unsloth+unsloth_zoo both explicitly (git install misses the zoo), hf_transfer as separate package. cu124 is flagged a dead end — next rebuild should move to cu126/cu128. The Unsloth template (used for eval) differs: nvcc present, torch 2.10/cu128, xformers pre-installed, FA2 absent.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#version-pin-discoveries; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons; working/eval/iter0_eval/001_README.md
- **relations:** 00_vm_setup.sh as canonical environment; snapshot economics
- **verify-later:** working/scripts/00_vm_setup.sh pin lines

<!-- SOURCE: U06_finetuning.md -->
### Snapshot economics: setup script beats VM snapshots
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Created `unsloth-trainer-base-01` then deleted it… Break-even: ~18 training runs/month. Reality: 1-4."
- **what:** Thunder snapshots bill the full provisioned 100GB ($15/month) regardless of used bytes, saving only ~25min/$0.85 per cold start — uneconomic below ~18 runs/month. Decision: no snapshots; the version-pinned idempotent `00_vm_setup.sh` is the canonical, version-controlled "snapshot". Phase-2 automation therefore provisions fresh instances every run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#snapshot-decision
- **relations:** version pinning; phase-2 architecture
- **verify-later:** none (economic decision); revisit if run tempo >15/month

<!-- SOURCE: U06_finetuning.md -->
### GPU training performance model (smoke ≠ steady state; FA2; seq-length cost)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** terminology.md whole file; NOTES(45) 2026-06-04: "The smoke rate (116 s/step) predicted this — nobody extrapolated it".
- **what:** A small captured mental model of training performance: smoke-test speed is unrepresentative (one-time kernel autotune/CUDA-graph costs amortized over too few steps); steady-state emerges after 5–20 steps; FA2 vs xformers/SDPA is a 2–4× attention-speed lever; attention scales O(N²) so max_seq 4096 quadruples 2048's attention work. Operationally: extrapolate full-run wall time from smoke s/step (the 18h-cap overrun happened because nobody did).
- **sources:** working/flywheel_docs/terminology.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** iter_0 baseline; per-instance uptime bump; cap-sizing-from-smoke queued idea
- **verify-later:** whether cap-sizing-from-smoke was ever built

<!-- SOURCE: U06_finetuning.md -->
### fp16 adapter save decision
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** HANDOFF 2026-05-08 decisions: "Save adapters as fp16, not fp32, in iter_1. One-line script change." (iter_0 shipped fp32 at 791–828MB.)
- **what:** PEFT `save_pretrained()` defaults LoRA weights to fp32 even when training in bf16, doubling adapter size and transfer time (17min tnr scp for 791MB). The one-line fix (cast trainable params to fp16 pre-save) was agreed for iter_1 but iter_1 never ran in these docs.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#lessons(10); working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** adapter transport via S3; model_lifecycle.artefacts format field
- **verify-later:** current 02_train script save path

<!-- SOURCE: U06_finetuning.md -->
### Flywheel D replay-eval methodology
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Replay, don't re-run"; §2.4d full replay design and partial results.
- **what:** Evaluation replays stored production prompts from llm_call_log against the candidate model instead of re-invoking agents — no orchestration-state pollution, much faster, and directly comparable to the stored Claude output. Test sets use `DISTINCT ON (orchestration_id)` for diversity ("Diverse 20 > random 20"), exported as NDJSON. Fail fast on empty responses; monitor with a watch loop.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/flywheel_D_target_selection.sql (header)
- **relations:** three-level eval pipeline; held-out set; CPU-Ollama eval attempt
- **verify-later:** held_out_cases_v1.sql query against llm_call_log

<!-- SOURCE: U06_finetuning.md -->
### Three-level evaluation pipeline (L1 structural / L2 judge / L3 spot-check)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 002_README-flywheele_D_evaluation_pipeline.md (run instructions, ~$1 / ~5min total); iter0_evaluation_report.md generated 2026-05-08.
- **what:** Reusable eval stack: L1 structural metrics computed locally and side-by-side for both models (JSON validity, schema-key match, length ratios, forbidden phrases from the brief's avoid-list, fabrication-marker regexes); L2 Claude-as-judge scoring relevance/voice/integrity 1–5 plus winner, with anonymised randomised A/B and resume support; L3 auto-selected spot-check cases folded into a markdown report by build_report.py. The report deliberately reports confounds and makes no ship/no-ship call. Known limit: L1 fabrication regexes have poor recall — contextual fabrications need L2/L3.
- **sources:** working/flywheel_docs/002_README-flywheele_D_evaluation_pipeline.md; working/eval/iter0_eval/iter0_evaluation_report.md#methodology; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(6)
- **relations:** model_lifecycle.evaluations (l1_metrics/l2_metrics JSONB contract); judge-model choice; eval gate
- **verify-later:** working/eval/iter0_eval/05-07 scripts; any evaluations rows in model_lifecycle

<!-- SOURCE: U06_finetuning.md -->
### Held-out eval set v1 as the canonical cross-iteration comparison set
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decision: "held_out_cases_v1.jsonl is the canonical eval set across iterations — same 20 cases… so trends are meaningful."
- **what:** 50 cases pulled from llm_call_log post-training-export-cutoff (created_at > 2026-04-23 14:54:32Z), one per orchestration, defensively excluded from the training set by source_log_id; 20 used for iter_0, 30 reserved. The SQL is kept for reproducibility. Iterations evaluate against the same cases so deltas are trend, not noise. Fresh `_v2` sets are the mechanism for novelty checks.
- **sources:** working/eval/v1/held_out_cases_v1.sql; working/eval/iter0_eval/iter0_evaluation_report.md#sample-selection; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** replay eval; three-level pipeline
- **verify-later:** held_out_cases_v1.jsonl vs training export overlap

<!-- SOURCE: U06_finetuning.md -->
### Claude-as-judge with anonymised A/B and self-recognition bias handling
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** iter0 report L2: "5 cases had identical R/V/I scores… 5/5 went to Claude… consistent with residual self-recognition bias"; HANDOFF 2026-05-08 decision: "claude-opus-4-7 is the canonical judge model".
- **what:** Judge design: anonymise responses, randomise A/B positions, score dimensions before picking a winner, and use a *different* Claude model (Opus) than the training-label producer (Sonnet 4.6) to reduce — not eliminate — self-recognition. The bias was empirically observed: rubric-tied cases broke for Claude every time, so headline win-rates get an adjusted reading (16-4 → 12-4 with 4 judge-preference ties). Position bias is checked explicitly (A won 55%, no clear bias).
- **sources:** working/eval/iter0_eval/iter0_evaluation_report.md#level-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(5)
- **relations:** three-level pipeline; model_lifecycle.evaluations judge_model column (judge drift tracking index)
- **verify-later:** level2.py anonymisation logic

<!-- SOURCE: U06_finetuning.md -->
### iter_0 verdict: shippable for low-stakes; voice fidelity is the iter_1 lever
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decisions: "iter_0 is shippable for low-stakes use… Not for client-facing where Δ−0.20 on voice would be visible"; "Add improve voice fidelity".
- **what:** The evaluated position on the first adapter: iter_0 matches Claude on JSON validity (20/20 vs 19/20) and schema, comparable length, tiny dimension gaps (relevance −0.25, voice −0.20, integrity −0.10), 4 substantive wins. Verdict: usable for internal tooling and low-stakes sites; voice is the largest gap and the main iter_1 lever (more epochs? lora_r 32? stricter voice-compliant training rows). "Address verbosity" was explicitly dropped (data showed no gap). Fabrication is a both-models problem to solve with prompt-time guardrails or post-hoc verification, not adapter training.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md; working/eval/iter0_eval/iter0_evaluation_report.md#tldr; working/eval/001_test_comparison_with_claude.txt
- **relations:** eval gate; deployment_decision vocabulary; fine-tuning candidates
- **verify-later:** whether any deployment_decision row exists; whether iter_0 was ever served in production

<!-- SOURCE: U06_finetuning.md -->
### CPU-Ollama replay eval attempt and the dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.4d "mistral-small3.1 on a shared cpu-ollama adapter is not a practical substrate… 20 cases × 25-30 min = 10+ hours"; superseded by Thunder GPU eval (HANDOFF 2026-05-08 ran 20 inferences at ~22s/case for ~$0.50).
- **what:** The first flywheel-D attempt replayed prompts against mistral-small3.1 on the shared CPU Ollama adapter; production contention drove one case to 27 minutes (~4 s/token), so a dedicated `ollama-eval` pod (own PVC/service, invisible to production routing because kafka-scheduler only probes ai_endpoint_health entries) was spun up, with the pod-memory sizing rule learned (limit ≥ model file + 8–12GiB headroom). The whole CPU-eval path was then superseded by evaluating the trained adapter on Thunder GPU instances. Yields the durable prediction framework (swap-with-prompt-tweaks vs swap-after-finetuning vs different substrate).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** ai_endpoint_health; Ollama CPU adapter ops; replay eval
- **verify-later:** whether ollama-eval deployment still exists in kustomize/cluster

<!-- SOURCE: U06_finetuning.md -->
### Fine-tuning candidate prioritisation
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §2.6 priority table; flywheel_D_target_selection.sql header ("high volume, high success, short output, structured JSON, low reasoning complexity").
- **what:** Ranked list of agents worth fine-tuning locally: knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher — all high-volume structured-JSON emitters. Explicit non-candidates: page-content-writer long-form (though its iter_0 hero step WAS the first target), visual-design-auditor (judgement), chief-strategist (worth Claude cost). Selection criteria are encoded as a reusable SQL discovery query over llm_call_log volume/recency.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.6; working/flywheel_docs/flywheel_D_target_selection.sql
- **relations:** per-vertical training (vertical column in llm_call_log); model swap
- **verify-later:** current llm_call_log volumes per agent

<!-- SOURCE: U06_finetuning.md -->
### Three improvement channels compound (RAG / LoRA / prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(25) §3, sourced from 009_model_infrastructure decision 10; RAG deployed, LoRA iter_0 trained, prompt_variant column exists but no A/B usage evidenced.
- **what:** The framing that RAG (immediate, no training), LoRA fine-tunes (medium-term cost reduction), and prompt evolution via the `prompt_variant` A/B column are three independent levers that compound: good prompts + good RAG produce the best training data, which produces the best fine-tuned model, which needs good prompts and RAG to perform.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#3
- **relations:** knowledge_base RAG; llm_call_log capture; llm-quality-testing category
- **verify-later:** any prompt_variant A/B analysis in code or docs

<!-- SOURCE: U06_finetuning.md -->
### model_lifecycle schema (training_runs / artefacts / evaluations / deployable_adapters)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 019_model_lifecycle_schema.sql (full DDL with comments); NOTES(45) 2026-06-09(6) confirms `model_lifecycle.training_runs` live with CHECK status pending/running/complete/failed and live rows.
- **what:** The run-lifecycle namespace: `training_runs` (one row per QLoRA run, FK to training_exports.runs, JSONB hyperparameters for reproducibility, loss/VRAM/cost outcome metrics, thunder_instance_id breadcrumb), `artefacts` (adapter binaries decoupled from runs to allow requantisation, storage_uri + sha256 + format), `evaluations` (per artefact × eval_set × judge, JSONB l1/l2 metrics, free-text human deployment_decision), plus `deployable_adapters` view (latest shipped_% adapter per base model — the chassis's read point for "which adapter to load") and `latest_training_run_per_export`. Supersedes the earlier flat `model_training_runs` sketch in FOCUS §2.5.1/HANDOFF 05-08.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#tables-needed; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7
- **relations:** eval gate; thunder_instances (FK training_run_id); mark_training_run_running/terminal actions
- **verify-later:** schema in clients_db; whether deployable_adapters is read by any chassis code

<!-- SOURCE: U06_finetuning.md -->
### Eval gate before promotion (human deployment decision; integrity lives here)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 019 schema docstring "deployment_decision is set by human after review; nullable until then"; PLAN b2 "No upload scheme substitutes for evaluating the adapter"; HANDOFF 05-08 "Auto-deployment NOT included in v1".
- **what:** Adapters never auto-promote: a human reviews flywheel-D output and writes a free-text deployment_decision ('shipped_internal', 'rejected_voice_gap', …); anything `shipped_%` becomes deployable. Critically, the eval gate is also the *integrity* boundary for the hostile-VM upload design — a maliciously-crafted-but-valid adapter written through a legitimate presigned URL is caught by evaluation, not by credentials. The original phase-2 sketch had a conditional auto-swap (`swap_agent_model if score ≥ threshold`) which was walked back to human review.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql#evaluations; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#agents-needed
- **relations:** hostile-VM threat model; model swap/revert; deployable_adapters view
- **verify-later:** whether any evaluation row has a decision; whether swap_agent_model was ever wired to eval output

<!-- SOURCE: U06_finetuning.md -->
### Flywheel C phase-2 automation architecture (HTTP job server → SSH-exec → adapter dispatch)
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.5.1 "HTTP job server (Option B chosen)"; HANDOFF 2026-05-08 "Architecture decision: SSH-exec, not HTTP job server (initially)"; the built system (phase 5) uses thunder-adapter dispatch actions + presigned URLs + detached run.sh.
- **what:** The automation design went through three generations: (1) VM-side HTTP job server (POST /jobs, bearer auth, systemd, TLS) — designed, never built; (2) direct SSH-exec from chassis (simpler at low run frequency, no VM service to maintain) — the pivot decision; (3) the final built shape: chassis dispatch actions publish to thunder-adapter (provision/ssh_exec/presign), data moves only via presigned B2 URLs, training runs detached under run.sh with a separate monitor. Rejected throughout: Kafka consumer on the VM (connectivity + overkill). "Chassis drives, GPU VM serves" is the invariant across all three; each generation supersedes the previous.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5.1; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#architecture-decision; working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** model-trainer chain; training-launcher; presigned data plane
- **verify-later:** no HTTP job server should exist anywhere; adapter dispatch actions in registry.go

<!-- SOURCE: U06_finetuning.md -->
### model-trainer orchestration chain (spawn/call data-preparer → provisioner → launcher)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "model-trainer flow confirmed live: spawn_data_preparer → spawn_provisioner → spawn_launcher → call_data_preparer → call_provisioner → call_launcher → complete".
- **what:** The end-to-end automated training run is the `model-trainer` orchestrator (id 94f5a069): spawns three worker agents up front, then calls them in order. `training-data-preparer` (71ab9361) streams the export to S3 as JSONL and INSERTs the pending training_runs row; `gpu-provisioner` provisions the A100 through thunder-adapter; `training-launcher` (1223bdc1) presigns, writes the manifest, and SSH-launches. Known open bug: a failed call_agent step falls through to the next call instead of aborting the orchestration (produces confusing secondary errors).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#1,#update-2026-06-06; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#where-this-sits,#also-pending
- **relations:** training-launcher workflow; migrations 103/104; call_agent fall-through bug
- **verify-later:** agent_definitions 94f5a069/71ab9361/1223bdc1 in clients_db; the fall-through behaviour in coordinator code

<!-- SOURCE: U06_finetuning.md -->
### training-launcher real workflow (presign → manifest → detached SSH launch → mark_running)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… presign_dataset → presign_scripts → compute_keys → presign_checkpoints (ONE batch await) → presign_final → [check_resume] → assemble_manifest → write_manifest → ssh_exec_launch → mark_running → complete."
- **what:** The launcher replaced a stub (migration 102) with a workflow of dispatch actions cloned from the proven decommission pattern: presign dataset + scripts bundle (GET), compute K checkpoint keys, batch-presign checkpoint PUTs + final PUT, optionally resolve a resume checkpoint, assemble and SSH-place `/workspace/upload_manifest.json`, then launch training detached and flip training_runs pending→running (`mark_training_run_running`, hardcoded guarded transition). Constants live in step config; cross-step values resolve via config dot-paths; the ssh command is built from a `command_template` with `{token}` interpolation. Evolved through migrations 102→105→109/109a/109b→110→111.
- **sources:** working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#5,#update-2026-06-09-3; working/phase5/102_training_launcher_real.sql (header)
- **relations:** batch presign; upload manifest; setsid launch; migrations family
- **verify-later:** live training-launcher default_config (2d state check); registry entries for the dispatch actions

<!-- SOURCE: U06_finetuning.md -->
### setsid detached launch and the detached exit-0 false-success gap
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §4 (the single-line setsid command) and 2026-06-03 ~18:04: "exit_code 0 only because the command's last token is echo (the known detached-ssh_exec false-success)".
- **what:** The adapter's ssh_exec blocks until the remote command exits (5-min timeout), so the launch chain (curl bundle + dataset via presigned URLs, untar, nohup run.sh) runs under `setsid … & echo LAUNCH_PID=$!` — the SSH session returns immediately with the PID. The cost: exit_code 0 only proves the echo ran; VM-side failures inside the detached chain (e.g. the /workspace permission failure) are invisible to the launcher. Corollary lessons: the command_template must stand up its own workspace with sudo mkdir+chown (105/109a), and any best-effort VM step under `set -e` becomes fatal (the root-owned ~/.bashrc append killed a run at the last cosmetic setup step).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#4,#update-2026-06-03; working/phase5/105_launcher_workspace_sudo_mkdir.sql (header); working/phase5/109a_fix_write_manifest_workspace_perm.sql (header)
- **relations:** run.sh markers (the real success signal); training monitor (fills the observation gap)
- **verify-later:** current command_template in the live launcher def

<!-- SOURCE: U06_finetuning.md -->
### run.sh launch chain and RUN_SH_* marker protocol
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** run.sh header ("Emits grep-able RUN_SH_* markers for the future monitor"); NOTES(45) 2026-06-09(4) verified live: "RUN_SH_START → STEP setup → STEP smoke → SMOKE_OK → STEP full_train → RUN_SH_UPLOAD manifest=present".
- **what:** All heavy on-VM work lives in run.sh (setup → smoke train → full train), not in the chassis workflow, so the chain is editable by re-uploading the bundle with no DB migration. It emits a marker protocol to /workspace/train.log (`RUN_SH_START/STEP/SMOKE_OK/FULL_OK/DONE/FATAL`) that is the machine-readable contract for the training monitor's probe. After Phase C, `set -euo pipefail` plus the hard-gated final upload means `RUN_SH_DONE` ⇒ trained AND adapter durable in B2. A mid-train crash leaves no marker (GONE_UNKNOWN).
- **sources:** working/scripts/run.sh (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#run.sh; working/phase5/NOTES_phase5_training_launcher_running(45).md#8
- **relations:** monitor probe classification; scripts bundle as deployment unit; CheckpointUploader
- **verify-later:** run.sh in the live B2 bundle vs the repo copy

<!-- SOURCE: U06_finetuning.md -->
### Scripts bundle in B2 as the training deployment unit
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~19:1x: "re-uploading the object IS the whole deploy"; RUNBOOK(8) §4a flat-bundle verification steps.
- **what:** The on-VM scripts (run.sh, 00_vm_setup.sh, 02_train, 03_inference_test) ship as `finetuning/scripts/bundle.tar.gz` in the personae-model-training bucket; the launcher presigns a GET and the VM curls+untars it. The bundle must be flat (files at archive root). Re-uploading the object deploys new training code — no chassis or DB change — with the corollary that editing a script without re-tarring deploys nothing (byte-identical md5 trap). The agent def holds only the key.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x; working/phase5/UPLOAD_bundle.sh; working/phase5/RUNBOOK_iter0_pretrigger(8).md#4a; working/scripts/README_setup.md
- **relations:** run.sh; presigned data plane; SAVE_STEPS re-pack for fast tests
- **verify-later:** b2 ls of finetuning/scripts/; bundle contents vs working/scripts/

<!-- SOURCE: U06_finetuning.md -->
### Checkpoint & final-adapter durability via pre-minted presigned PUT manifest
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) status header: "Phases A, B, C BUILT… ckpt-0 confirmed in B2 on run 0ac806ab (the upload path is proven end-to-end)"; "Still unproven in prod is one run reaching RUN_SH_DONE with the final adapter.tar.gz durable".
- **what:** The Thunder VM disk is ephemeral and originally nothing moved training output off it (no checkpoints — save_strategy "no" — and the adapter saved only locally, so a reap = total loss and the monitor's DONE_OK→decommission would have destroyed the artefact). The fix: the launcher pre-mints single-object write-only presigned PUT URLs (K checkpoints + 1 final, keyed `finetuning/checkpoints/<run_id>/ckpt-<index>.tar.gz` and `finetuning/artefacts/<run_id>/adapter.tar.gz`) into `/workspace/upload_manifest.json`; the VM uploads through them. Checkpoints are keyed by save-INDEX not Trainer global_step (fragile to predict); write-once with B2 versioning as backstop; URL expiry must exceed max_uptime (expiry_minutes 3000). Checkpoint upload proven (ckpt-0 in B2); the final-adapter upload and a full RUN_SH_DONE run remain the empirical gate.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md; working/phase5/README_what_is_manifest.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-5
- **relations:** hostile-VM threat model; CheckpointUploader; resume; monitor enablement gate; batch presign
- **verify-later:** b2 contents under finetuning/checkpoints/ and artefacts/; migrations 109/110/111 state in the live def

<!-- SOURCE: U06_finetuning.md -->
### CheckpointUploader trainer callback (best-effort checkpoints, hard-gated final upload)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** PLAN(7) §02_train "BUILT 2026-06-05… Tier 1 (box-free) PASSED"; NOTES(45) 2026-06-05 update.
- **what:** `02_train` gained gated flags (`--save-steps`, `--save-total-limit`, `--upload-manifest`; defaults keep old behaviour byte-for-byte). A `CheckpointUploader(TrainerCallback).on_save` tars each checkpoint and PUTs it to its save-index URL synchronously (a background thread was rejected — races save_total_limit's dir deletion); checkpoint upload failure is best-effort (log and continue). The FINAL adapter upload is a hard gate: failure raises → non-zero exit → no RUN_SH_DONE → the monitor never treats the box as cleanly done (degrades to GONE_UNKNOWN→failed, never a false DONE_OK). Content-Type application/octet-stream confirmed accepted by the unbound presigned signature. Two manifests coexist: input upload_manifest.json vs the output run-metadata manifest.json.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#02_train; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-05
- **relations:** run.sh markers; durability manifest; save_steps cadence (50 ≈ one checkpoint/1.5–2h, ~2GB each: adapter + AdamW state)
- **verify-later:** 02_train in the live bundle; checkpoint sizes in B2

<!-- SOURCE: U06_finetuning.md -->
### Resume path (cluster-side checkpoint selection, presence-of-checkpoints as the signal)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) 2026-06-14 update: "batch + resume BUILT, APPLIED, and verified [def-state]… still unproven in prod"; migration 111 applied and 2d-verified.
- **what:** A relaunch for the same training_run_id becomes a continuation automatically: the launcher's `check_resume` step asks the adapter (`prepare_resume_url`) to list `finetuning/checkpoints/<run_id>/` in B2 (reusing the existing `storage.Client.ListObjects` — the presumed "list-keys gap" was wrong), pick the highest ckpt-N, and presign a GET; assemble_manifest emits a `resume` block only when found. 02_train downloads/extracts it and calls `trainer.train(resume_from_checkpoint=True)` (restores optimizer/scheduler/RNG/step). No separate resume mode — empty prefix = fresh start; the launcher owns save-index key allocation across resume launches. found=false is a valid answer; transient list failures return error_recoverable so the chassis retries.
- **sources:** working/phase5/README_what_is_manifest.md; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#resume; working/phase5/111_training_launcher_resume_wiring.sql (header)
- **relations:** durability manifest; monitor GONE_UNKNOWN (total-loss case that motivated this)
- **verify-later:** a real kill-and-resume test; dispatch_thunder_prepare_resume_url in registry

<!-- SOURCE: U06_finetuning.md -->
### Monitor enablement gate: DONE must mean durable
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7): "enable thunder-training-monitor (safe once DONE means durable — gated on the first run actually reaching RUN_SH_DONE)"; NOTES(45) §9 "Not enabled; enabling is RUNBOOK step 6".
- **what:** An explicit sequencing invariant: the monitor's DONE_OK path decommissions the box (destroying the disk), so the schedule stays disabled until the upload path proves that RUN_SH_DONE implies the adapter is in B2. Enabling early would have destroyed iter_0's adapter. The interim protocol for in-flight runs was manual: scp adapter_out off the box before anything decommissions it.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#build-order; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#update-2026-06-04
- **relations:** CheckpointUploader hard gate; run.sh markers; monitor
- **verify-later:** scheduled_tasks.enabled for thunder-training-monitor; whether a run has since reached RUN_SH_DONE

<!-- SOURCE: U06_finetuning.md -->
### knowledge_base RAG store and flywheel B verification
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b action items all [x]; "Flywheel B is done" (step 3 chassis integration COMPLETED on v1.0.979, 2026-04-21).
- **what:** `knowledge_base` (migration 082): pgvector(768) for nomic-embed-text, shared across agents via `rag_lookup`/`rag_index` actions, trigram fallback when Ollama is down, SHA256 dedup, metadata-first filtering doctrine (filter by vertical/component_type/content_type/source before ranking by similarity — else a vet example surfaces for gas-wholesale copy). Verified bottom-up in three single-focus steps: pgvector+ivfflat+cosine on synthetic vectors, real-content retrieval through cpu-ollama, then chassis integration via a deterministic 3-step `rag-test-agent`. Nomic v1 judged good enough; v2-moe named as a drop-in upgrade (same 768 dims). Open task: periodic REINDEX of the ivfflat index.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.2,#2.4b; working/flywheel_docs/flywheel_B_step*.{sql,sh} (headers)
- **relations:** nomic prefixes; RAG-platform product; three channels
- **verify-later:** rag_actions.go; knowledge_base row counts; whether REINDEX ever scheduled

<!-- SOURCE: U06_finetuning.md -->
### Nomic task prefixes are load-bearing (rag_actions prefix patch)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b "[x] Prefix patch deployed and verified live 2026-04-21… log line 'prefix_applied':true observed".
- **what:** Without `search_document:`/`search_query:` task prefixes, nomic embeddings ranked a Labrador chunk above the French Bulldog chunk on a BOAS-specific query; with prefixes the correct result won with 5× the margin. The patch adds a model-scoped `applyNomicPrefix` helper (only nomic-embed-*, double-prefix guard, prefix_applied logged) at both embed call sites; stored chunks and dedup hashes stay unprefixed; trigram fallback untouched.
- **sources:** working/flywheel_docs/PATCH_rag_actions_nomic_prefixes.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4b
- **relations:** knowledge_base RAG; Ollama CPU adapter
- **verify-later:** rag_actions.go contains applyNomicPrefix

<!-- SOURCE: U12_docs024_archives.md -->
### Quality improvement flywheel (RAG + LoRA + prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** Live `009_model_infrastructure.md` "Future": RAG actions "registered, not workflow-tested"; LoRA training pipeline and training-data export from `llm_call_log` both still open.
- **what:** Three independently-valuable, compounding improvement channels: RAG (inject retrieved good examples at call time), LoRA (retrain periodically on filtered successful outputs), and deliberate prompt A/B testing (80/20 traffic split, promote on audit-success-rate). A `training-orchestrator` workflow packages LoRA training as an adapter-driven workflow (export → start_gpu_instance → train → evaluate → deploy_or_reject → stop_gpu_instance → log). A scraped-data "AI slop" quality gate filters what may enter the training set.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Quality Improvement Flywheel", #"Scraped Data Quality Gate (AI Slop Prevention)"; docs024_key_docs_latest/009_model_infrastructure.md#"Future"
- **relations:** GPU/AI-endpoint scheduling; llm_call_log flywheel columns
- **verify-later:** whether any `training_runs` completed beyond the one noted in 009; whether RAG actions are workflow-exercised.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fine-tuning flywheel (call-log → LoRA → GGUF → Ollama)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** 001(0) "The training data pipeline: LLM call logging → export → LoRA fine-tune on GPU → GGUF export → load into Ollama → update agent definition to provider: ollama"; "Not yet built (future work)"
- **what:** A path to replace short-output classification/extraction agents with local fine-tuned models: accumulate 200+ successful `llm_call_log` examples, export Alpaca/ChatML, LoRA fine-tune on GPU (unsloth), export GGUF into Ollama, flip the agent definition to `provider: ollama`, then A/B test against Claude.
- **sources:** WM/001_development_guide(0).md#fine-tuning-path, WM/001_development_guide(0).md#implementation-status-llm-optimization, WM/033_thunder_adapter_design.md#tldr
- **relations:** LLM infrastructure; Thunder adapter; LLM tiering
- **verify-later:** training_data_export.sql; model_lifecycle.training_runs; unsloth

<!-- SOURCE: U18_sql_for_agents.md -->
### Finetuning flywheel Phase 5: training kickoff orchestration (model-trainer, training-data-preparer, gpu-provisioner, training-launcher)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 109/110 real; 111/112 explicitly STUBs "to unblock end-to-end testing... will be replaced by a real implementation in a future migration"; 116/117 monitor real running instances, implying provisioning later became real.
- **what:** model-trainer owns the KICKOFF phase: training-data-preparer exports a training_exports snapshot as JSONL to S3 and INSERTs the model_lifecycle.training_runs row (pending); gpu-provisioner calls Thunder Compute API for an A100, stores the SSH key as a k8s secret; training-launcher SCPs scripts/dataset and nohup-launches training, returning the pid. The workflow exits immediately — completion is deliberately handled by a separate scheduled monitor so no orchestration holds open for ~9 hours. Full hyperparameter set captured for reproducibility.
- **sources:** 109_model_trainer_orchestrator.sql; 110_training_data_preparer.sql; 111_gpu_provisioner_thunder.sql; 112_training_launcher.sql
- **relations:** training-data-exporter (106) upstream; thunder monitor/reaper downstream; thunder-adapter
- **verify-later:** real gpu-provisioner/training-launcher implementations vs stubs

<!-- SOURCE: U18_sql_for_agents.md -->
### Thunder instance lifecycle: reaper + training monitor (orchestrator/worker)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 114 (reaper, every 15 min, idempotent decommission); 116/117 with verified coordinator internals and the insert-DISABLED-until-actions-deploy discipline.
- **what:** Cost/safety controls for rented GPUs: thunder-reaper decommissions instances past max_uptime_hours (one per tick, pre_query LIMIT 1); thunder-training-monitor orchestrator finds every running training instance each tick and spawns a per-instance worker that probes via SSH, classifies (alive / unreachable-streak / done_ok / done_fail), reconciles training_runs and decommissions. 117 records WHY orchestrator-with-loop beats the reaper's scheduler-pre_query shape (must visit every instance, not just the top row) and why the loop must stay sequential (topic reuse safety).
- **sources:** 114_thunder_reaper.sql; 116_thunder_training_monitor_worker.sql; 117_thunder_training_monitor_orchestrator.sql
- **relations:** scheduler-and-tasks (pre_query dispatch patterns); thunder adapter; model-trainer
- **verify-later:** scheduled_tasks rows enabled; probe/classify actions

<!-- SOURCE: U18_sql_for_agents.md -->
### training-data-exporter (llm_call_log → ChatML JSONL)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 106 definition with concrete kubectl retrieval instructions and input payload example.
- **what:** Deterministic single-action agent exporting successful LLM calls from llm_call_log as NDJSON training data in ChatML + metadata format, filterable by agent_type/step/model, with fenced-output and strict-JSON options.
- **sources:** 106_training_data_exporter.sql; 040_optimise_which_llms.sql (llm_call_log)
- **relations:** training-data-preparer consumes exports; flywheel columns from 085
- **verify-later:** training_data_export action; training_exports schema

<!-- SOURCE: U19_sql_tables_components.md -->
### llm_call_log training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Migration 081 Part 3 + schema fixes (agent_id added, nullability relaxed to match Go's nullIfEmpty); export queries reference populated columns incl. work_item_id and vertical.
- **what:** Every LLM call logged with caller identity (agent_type/step/orchestration), model + model_resolved + provider, full prompt_template/prompt_rendered/response_text, token/latency usage and outcome — explicitly designed for training export. Export recipes produce JSONL per task (analyze_tool, recreate_tool, site classification, content writing) with quality filters joining site_work_items outcomes (only export calls whose work item completed), and per-vertical readiness counts.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql
- **relations:** training_exports (successor storage); site_chat_turns (deliberately separate); model upgrades.
- **verify-later:** logging middleware in aiservice; work_item_id/vertical columns present.

<!-- SOURCE: U19_sql_tables_components.md -->
### training_exports Postgres-backed datasets (flywheel A v3)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Schema with rationale "JSONL files landed on ephemeral chassis pods and vanished on restart"; dedup unique index on (export_id, metadata->>'source_log_id').
- **what:** Named, versioned training datasets in Postgres instead of ephemeral JSONL: runs (one per export — filter criteria matching llm_call_log columns, counts, skip reasons, format 'chatml', size, provenance) and rows (ChatML messages + metadata JSONB, ordered by row_index, CASCADE delete). Training-time extraction via \copy in export order. Schema named training_exports specifically to avoid confusion with the model-training pipeline (flywheel C).
- **sources:** docs/agent_docs/sql_for_tables/039_training_exports.sql
- **relations:** llm_call_log source; thunder training runs (flywheel C).
- **verify-later:** exporter action writing runs/rows.

<!-- SOURCE: U22_recent_small_docs.md -->
### llm_call_log (build-time training flywheel)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "The llm_call_log table is capturing every LLM call from every agent ... This started logging the moment you deployed the new chassis image" vs handoff listing the Go patches as ready-but-not-committed.
- **what:** A table capturing every `execute_llm_prompt` call (agent_type, step, model, rendered prompt, response, input/output tokens, latency, success) via a fire-and-forget goroutine logger. Feeds cost/latency analytics (`llm_call_stats` view) and accumulates toward the 200+-examples-per-agent fine-tuning threshold. Cleanup function exists but nothing calls it (table-bloat risk flagged, ~1GB/month).
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../005_PATCHES.md#patch-01-02, docs020.../001_rag_agent_distribution_architecture.md#item-2
- **relations:** anthropic.go usage capture patch, site_chat_turns (deliberately separate log), LoRA training data export
- **verify-later:** llm_call_log table + cleanup_old_llm_logs; LogLLMCall in ai_actions.go; anthropic.go __usage_input_tokens write-back

<!-- SOURCE: U22_recent_small_docs.md -->
### Fine-tuning pipeline (LoRA flywheel, deferred)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Phase 7: Fine-Tuning Pipeline (Deferred) ... becomes relevant after 200+ successful examples per agent type accumulate."
- **what:** The deferred end of the flywheel: once llm_call_log has 200+ examples/agent, export JSONL training data, QLoRA fine-tune a 7B via Unsloth on rented GPU, export GGUF, load into Ollama (`ollama create`), and switch the agent definition to `provider: ollama`. First candidates: site-classifier (high volume, short output), then domain-research-classifier, then the vertical knowledge extractor. Purpose: drive per-call inference cost to ~zero.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-7, docs023.../018_canine_biology.md#6, docs020.../010_simple_explanation.md
- **relations:** llm_call_log, Ollama provider, canine biology text LoRA, self-hosted LLM inference
- **verify-later:** training-data export queries; any GGUF/Ollama custom models; agent_definitions using provider:ollama

<!-- SOURCE: U22_recent_small_docs.md -->
### Text LoRA — veterinary knowledge extractor
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** Phase E todo "Text LoRA fine-tuning (week 6-7)" unchecked; full Unsloth/QLoRA recipe given as instructions.
- **what:** A concrete recipe to fine-tune a local 7-8B model (Unsloth QLoRA, r=16, 3 epochs) on accumulated knowledge-extraction examples, export Q4_K_M GGUF, load into Ollama, and swap `knowledge-extractor` to the local model to eliminate Claude API cost per extraction. Training data accrues naturally during the canine research phase (50 breeds + 30 conditions + 40 procedures ≈ 120, need 200+).
- **sources:** docs023.../018_canine_biology.md#6
- **relations:** fine-tuning pipeline, llm_call_log, Ollama provider
- **verify-later:** vet-extractor GGUF/Ollama model; knowledge-extractor agent provider

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Quality improvement flywheel (RAG + LoRA + prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 020c "Quality Improvement Flywheel": three channels (RAG more examples, LoRA retrain on best outputs, deliberate A/B prompt evolution with `prompt_config.active_variant`/`testing_allocation:0.2`); llm_call_log columns `work_item_id/prompt_variant/vertical/rag_context_used` "Not Yet Deployed"; LoRA training as `training-orchestrator` workflow (export→gpu→train→evaluate→deploy_or_reject).
- **what:** Compounding quality via three independent channels feeding site production: knowledge collection (scrape + successful Claude outputs + audit insights, quality-gated) → RAG injection, LoRA retraining (Unsloth QLoRA on ThunderCompute), and deliberate prompt A/B testing scored by audit success rate. llm_call_log gains flywheel columns; measurable metrics (first-pass success, rewrite count, lock rate) precede traffic metrics.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#quality-improvement-flywheel; #prompt-evolution
- **relations:** live FOCUS_finetuning_flywheel_and_service; RAG best practices; canine biology LoRA; per-vertical LoRA
- **verify-later:** llm_call_log flywheel columns; training-orchestrator workflow; training_runs table

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Internal AI training flywheel (A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §1 table "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (2026-04-23)
- **what:** The core internal loop: the site-building pipeline logs every LLM call as a byproduct; that data periodically fine-tunes local models that are swapped in for Claude calls where quality holds, dropping API cost. Four lanes — A (data export), B (RAG), C (LoRA training), D (Claude-vs-local eval). A and B were done; C scripted; D paused on infra contention.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#1, #2, #4
- **relations:** parents Flywheel A export, Flywheel B RAG, Flywheel C fine-tuning, Flywheel D eval; feeds Phase 5 launcher; three improvement channels
- **verify-later:** llm_call_log, knowledge_base, training_exports schema, ai_endpoint_health

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel A — training-data export pipeline
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4i "First real training dataset now in Postgres: export_id fef7be6b-…, 1,958 rows, 21.2MB" and "Spawning architecture fully validated" (2026-04-23)
- **what:** A chassis action (`training_data_export`) + `training-data-exporter` specialist agent (wrapped by `training-data-export-orchestrator`) that reads `llm_call_log`, strips markdown code fences via `stripMarkdownFromResponse`, validates JSON, and writes ChatML training rows. Evolved v1 (static file config, superseded) → v2 (reads params.CollectedData["input_data"], file output to /tmp) → v3/v3.1/v3.2 (writes to a dedicated `training_exports` Postgres schema, per-batch transactions to survive pgbouncer).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** produces datasets for Flywheel C; superseded intermediate: v1 file-output export; feeds training_exports schema concept
- **verify-later:** platform/orchestration/actions/ training_data_export_v3.go; training_exports.runs, training_exports.rows; agent_definitions training-data-exporter

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model aliases and the model selection strategy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(5) rules; 009 per-step table; 101_switch_to_haiku.sql records live state 2026-04-10
- **what:** Agent definitions use short aliases (claude-sonnet-4-6, claude-haiku-4-5) resolved by model_aliases.go; sonnet is the default for LLM steps, haiku for routing, opus for chief-strategist/planner, ollama for fine-tuned classification. 101 SQL is a bulk cost lever switching all agents to haiku with a RESTORE section.
- **sources:** 001(5)#LLM Infrastructure; 009#Model Swap; 101_switch_to_haiku.sql
- **relations:** swap_agent_model; LLM tiering (029)
- **verify-later:** model_aliases.go; per-step ai_service in agent_definitions

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Ollama adapter (CPU embeddings + local classification)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009: "Ollama adapter on CPU cluster (2 replicas, mistral-small3.1 + nomic-embed-text)" checked done
- **what:** Permanent CPU adapter serving nomic-embed-text embeddings (~50-100ms) and quantized small models for classification (10-30s acceptable per-build). Same AIService interface as Anthropic incl. token-usage write-backs. Not for content generation or <2s latency.
- **sources:** 001(5)#Ollama adapter; 009#Implementation Status
- **relations:** RAG actions; endpoint health; fine-tuning path
- **verify-later:** ollama.go; ollama-adapter deployment

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Multi-endpoint model routing with ai_endpoint_health as the GPU scheduler
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 009: tables/functions applied+verified; the three Go patches (fast-fail, claim-gate, release-without-attempt) listed under "Next Deploy"; active pinging "starts after patches deployed"
- **what:** Endpoints (Claude API, CPU/GPU Ollama) tracked in ai_endpoint_health; healthy → items flow, unhealthy → items wait (no fallback chains — quality over speed; priority means importance only; items don't know about models). ClaimWorkItem checks handler's endpoint health before claiming; AIUnavailableError triggers reactive health update + release-to-triaged without attempt increment; Claude health dual-mode (reactive 402/401 + hourly 1-token ping).
- **sources:** 009#Decisions Made, #Health Check Architecture; 022 SQL
- **relations:** back-to-triage; endpoint-health-checker agent
- **verify-later:** were the three patches deployed since 2026-03-25

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model swap/snapshot/revert control plane (migration 083)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 021 SQL applied per 009; used operationally (016 §6.1 conventions)
- **what:** snapshot_agent()/swap_agent_model()/revert_agent() + agent_snapshots view make agent_definitions the model-routing control plane: per-step ai_service swaps with automatic snapshot, one-call revert. Post-migration snapshots live in agent_definitions_backup (snapshot_taken_at discriminator, restored_at audit trail); the legacy in-table is_snapshot rows caused a documented family of contamination/misroute bugs.
- **sources:** 021_model_swap_and_rollback.sql; 009#Model Swap Procedure; 016 §6.1/§9 snapshot bugs
- **relations:** backup naming discipline; LLM config shadowing (step-level swaps shadowed by top-level ai_service)
- **verify-later:** agent_definitions_backup schema

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM tiering (large/medium/small/none) + cluster-then-slot-fill scaling pattern
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029: "the chassis routes" described as design with llm_tier annotation to add; flip-to-local gated on Thunder health
- **what:** Every LLM call site declares a tier; chassis maps tier→endpoint via flippable config (large=Opus strategy/briefing; medium=Sonnet→local-70B for plan partials/audits; small=Haiku for slot-fills; none=deterministic Go). Product-listing scale: facts from feeds (Go), cluster ~10k products into ~20-50 groups algorithmically, one medium call per cluster for framing, small slot-fill per product.
- **sources:** 029#LLM tier per call site, #Affiliate/product listings
- **relations:** model aliases; batch queue routing
- **verify-later:** any llm_tier config keys in defs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### AI endpoint health routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Endpoint health routing deployed" [x] (2026-04); gpu-ollama "currently DOWN, not always-on"
- **what:** `ai_endpoint_health` (migration 085) tracks claude / cpu-ollama / gpu-ollama endpoints; healthy endpoint → work claims flow, unhealthy → items wait (back-to-triage). No separate batch scheduler for GPU: it's either healthy or not. Kafka-scheduler probes only endpoints listed here, so unlisted pods (ollama-eval) are invisible to production routing by design.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.3, #2.4d
- **relations:** Flywheel D dedicated eval pod; rate-limit transient classification
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Per-agent model swap / revert
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Swap / revert functions deployed" [x] (migration 083)
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` safely snapshot an agent's `ai_service` block in agent_definitions.default_config before swapping model, per-agent per-step; full-table backup remains as the nuclear option. The mechanism the flywheel's deployment decision hangs on.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4; HANDOFF_2026-05-26 (snapshot_agent used before jsonb_set edit)
- **relations:** training-flywheel-orchestrator conditional swap
- **verify-later:** migration 083 functions in DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Ollama CPU operations envelope
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; memory rule from 2026-04-23 OOM incident
- **what:** Hard-won ops facts for CPU Ollama: RollingUpdate + RWO PVC deadlocks (use Recreate); OLLAMA_LOAD_TIMEOUT=10m / KEEP_ALIVE=30m; cold load ~45s for 14GB models; pod memory limit ≥ model size + 8–12GiB (cgroup, not host, is what Ollama constrains against); /api/chat not /api/generate; throughput ~150 tok/s prompt, ~2.5 tok/s generation for mistral-small3.1 Q4 on 8 CPU cores.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4c, #14 "Ollama specifics"
- **relations:** dedicated eval pod; endpoint health
- **verify-later:** kustomize base for ollama-adapter

<!-- SOURCE: U06_finetuning.md -->
### thunder-training-monitor (periodic probe + reconcile + release)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** NOTES(45) 2026-06-04: "training-monitor VERIFIED live (both paths)… Terminal/decommission branch still never run live… Not enabled"; schedule inserted DISABLED (migration 108).
- **what:** A second periodic lifecycle agent beside the reaper: the `thunder-training-monitor` orchestrator runs `find_active_training_instances` then loops spawn_worker→call_worker per running training box (deliberately NOT the reaper's scheduler-pre_query shape, which merges only the first row per tick and would starve newer instances behind ALIVE boxes). Each `thunder-training-monitor-worker` probes via the adapter's `ssh_get_status` with a status_command that classifies run.sh markers into ALIVE | DONE_OK | DONE_FAIL | GONE_UNKNOWN (plus reachable:false as a valid answer), routes via classifier `next_step` override, reconciles `model_lifecycle.training_runs`, counts consecutive unreachable probes (≥3 → lost → decommission an unreachable-but-billing box), and on terminal verdicts releases the box through the shared idempotent decommission. Built as migrations 106/107/108 + 5 chassis actions. ALIVE path and orchestrator fan-out verified live; the terminal/decommission branch has never fired.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04,#update-2026-06-09-6; working/phase5/107_thunder_training_monitor_worker.sql (header)
- **relations:** thunder-reaper (responsibility split); run.sh markers; monitor enablement gate; reply-topic resolution bug (found here)
- **verify-later:** scheduled_tasks row enabled state; migrations 106-108; the 5 actions in registry.go

<!-- SOURCE: U06_finetuning.md -->
### Monitor/reaper responsibility split
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04: "Decision: build a separate thunder-training-monitor, NOT bolted into the time-reaper… the time-reaper is the last-line cost backstop and must stay dead-simple/dependency-free".
- **what:** Two periodic agents with deliberately distinct dependency profiles: the reaper (cost backstop) is pure DB + Thunder and must work even when the adapter is down; the monitor (completion-side) depends on adapter + SSH. They overlap only in calling the shared idempotent `decommission_instance`. The monitor exists because the launcher returns long before training ends (detached run) so completion can't be a workflow await.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md
- **relations:** thunder-reaper; thunder-training-monitor; orphan-sweep TODO (third member: boxes whose Thunder instance vanished)
- **verify-later:** both scheduled_tasks rows and their concurrency_group

<!-- SOURCE: U06_finetuning.md -->
### Thunder Compute adapter (provision/decommission lifecycle)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 2026-05-12(1) table: phases 2–3.5 "✅ Deployed and verified end-to-end"; FOCUS(25) §14 "Provision loop verified end-to-end (2026-05-22)".
- **what:** A Kafka adapter (`system.adapter.thunder.requests`) wrapping the Thunder Compute GPU API and owning its credentials. `provision_instance`: spend pre-check → ed25519 keypair → API create (public_key sent) → k8s Secret persist → WaitForRunning poll → INSERT thunder_instances with retry → compensating cleanup (fresh context) on partial failure. `decommission_instance`: lookup by provisioning_id or thunder identifier → atomic `decommissioning` transition as idempotency anchor → 404-tolerant API + Secret deletes → cost computed from running_since × snapshotted hourly rate. Error classification maps denial→unrecoverable, infra→recoverable. Includes hard-won API shape knowledge: base URL :8443/v1, lowercase gpu_type enums, camelCase string-numbers in responses vs snake_case ints in requests, recycled numeric ids requiring a partial unique index on live rows, real template names (`base`, not the OpenAPI example).
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md
- **relations:** gpu-provisioner; reaper; ssh_exec; presigned data plane; adapter design guide
- **verify-later:** internal/adapters/thunder/*; migrations 025/028/029; thunder_instances schema

<!-- SOURCE: U06_finetuning.md -->
### thunder-reaper scheduled task and per-instance uptime deadline
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) 3.5: "✅ Deployed and verified end-to-end (2026-05-14): synthetic row … picked up within 30s"; NOTES(45) 2026-06-04 live rescue of run fabfd7fa.
- **what:** A 15-min scheduled task whose pre_query finds `running` instances past `max_uptime_hours` and dispatches the idempotent decommission (one per tick, LIMIT 1). The deadline is OURS not Thunder's — computed as running_since + the per-row `max_uptime_hours` (default 18h; training provisions get 18h) — so a mid-train cap overrun can be rescued by bumping the single row's max_uptime_hours (done live: 18→48h when the 24h iter_0 train would have been reaped at hour 18, which with save_strategy=no meant total loss). Reaper reason strings are meaningful text for post-mortems.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#3.5; working/flywheel_docs/README.md#3.5-delivered; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** monitor/reaper split; spend gating; scheduler pre_query single-row semantics
- **verify-later:** migration 028; scheduled_tasks row; thunder_instances.max_uptime_hours column

<!-- SOURCE: U06_finetuning.md -->
### Thunder spend gating (DB-side provision check)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Spend gating lives in DB, not API"; NOTES(45) 2026-06-03 cost-gate check (cap 30, estimate 20, clears with $9 headroom).
- **what:** Before every create, the adapter consults the `thunder_provision_check` view: decommissioned 24h spend + running estimated spend + `estimated_new_run_cost_usd` must stay under `thunder_config.daily_cap_usd`. Operational learnings: keep the per-run estimate realistic (~$20 for a 9h+ A100 run — a $2 test estimate lets doomed runs through; a $25 default blocks legitimate tests); the 24h window is rolling so heavy test days trip the cap on legitimate past spend (raise the cap for the session, don't delete accurate rows).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-172x,#7
- **relations:** thunder adapter provision; reaper; cost anchors from iter_0
- **verify-later:** thunder_provision_check view definition; thunder_config values

<!-- SOURCE: U06_finetuning.md -->
### Orphan-sweep for stale live thunder_instances rows
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "TODO (deferred 2026-05-24) — orphan-sweep for stale live rows… Agreed design (not yet built)".
- **what:** Out-of-band deletions (manual `tnr delete`) leave DB rows `running` forever; because live rows hold the recycled Thunder id in a partial unique index, a stale row blocks the next provision with a duplicate-key error (bit on 2026-05-24). Agreed design: a `sweep_orphans` adapter action computes (live DB rows) minus (Thunder's live list) and dispatches the idempotent decommission per orphan, run as a 15–30min scheduled task sharing the reaper's concurrency group, with a safety guard never to sweep on a failed/partial Thunder list. Interim: manual row reconciliation after any manual delete.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics(TODO)
- **relations:** reaper (time-based) and monitor (completion-based) — this is the third, existence-based leg
- **verify-later:** whether sweep_orphans exists anywhere

<!-- SOURCE: U06_finetuning.md -->
### Adapter-managed SSH access to GPU boxes (ed25519 keys in k8s Secrets)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "RESOLVED & VERIFIED (Phase 4, 2026-05-24) — SSH connection mechanism + ssh_exec/ssh_get_status" with production verification detail.
- **what:** The adapter generates its own ed25519 keypair per provision, stores the private half in Secret `thunder-ssh-<db_row_id>` (deterministic name so orphan Secrets are reapable), sends the public half on create. `ssh_exec`/`ssh_get_status` dial `instance_ip:ssh_port` directly via x/crypto/ssh as user `ubuntu` (NOT root, despite Thunder's own ssh_command string), with a wait-for-sshd retry (~90s) because RUNNING precedes sshd. The port is the list-endpoint's `port` field, captured into thunder_instances.ssh_port. `reachable:false` is a valid answer, not an error — the probe primitive the monitor builds on. Manual-ops corollary: operators can extract the key from the Secret to watch train.log directly (StrictHostKeyChecking=no needed because Thunder recycles IPs).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/ssh_probe.sh (header); working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x
- **relations:** monitor probe; setsid launch; RBAC resourceNames trap (Secret permissions)
- **verify-later:** internal/adapters/thunder/ssh/*; the RBAC Role verbs

<!-- SOURCE: U06_finetuning.md -->
### Thunder Prototyping vs Production mode economics
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 lesson 8: "Prototyping (TGV virtualised) worked fine for 70B inference… Phase 2 should default to Prototyping for inference, Production for training (unverified that Prototyping handles long training runs well)."
- **what:** Thunder's Production mode ($1.79/hr A100 80GB) vs Prototyping ($0.78/hr, TGV-virtualised). Verified: Prototyping is fine for 70B inference. Unverified: whether virtualisation overhead degrades long QLoRA training runs enough to cancel the ~55% saving — flagged as an iter_1 experiment, never run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(8); working/phase2/README.md
- **relations:** cost anchors; gpu-provisioner defaults
- **verify-later:** provision defaults (mode) in gpu-provisioner/adapter config

<!-- SOURCE: U06_finetuning.md -->
### Ollama CPU adapter operational rules
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4c "RWO PVC rolling-restart deadlock on ollama-adapter — resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; §14 Ollama specifics list.
- **what:** Hard-won ops rules for CPU Ollama: Recreate (not RollingUpdate) deployment strategy because the PVC is RWO (classic new-pod-can't-mount deadlock); `OLLAMA_LOAD_TIMEOUT=10m` + `OLLAMA_KEEP_ALIVE=30m` (default 60s load timeout killed first inference after cold start — 14.4GB model loads in ~45s); pod memory limit ≥ model file size + 8–12GiB headroom (Ollama reads host /proc/meminfo but is constrained by the cgroup — misleading OOM messages); chassis calls `/api/chat` not `/api/generate`; measured CPU throughput ~150 tok/s prompt, ~2.5 tok/s generation on mistral-small3.1.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4c,#2.4d,#14
- **relations:** ai_endpoint_health; dedicated eval pod; CPU eval abandonment
- **verify-later:** ollama-adapter kustomize base

<!-- SOURCE: U06_finetuning.md -->
### ai_endpoint_health inference routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.3 endpoint table + §4.1 "[x] Endpoint health routing deployed"; gpu-ollama noted "currently DOWN, not always-on" and still DOWN in HANDOFF 2026-05-08.
- **what:** Three endpoints tracked in `ai_endpoint_health`: claude (default quality), cpu-ollama (embeddings + small models), gpu-ollama (70B/LoRAs — persistently down through this doc set). Healthy endpoint → work claims flow; unhealthy → items wait/back-to-triage, so GPU availability gates work without a separate batch scheduler. The kafka-scheduler only probes endpoints listed here — which is also why the dedicated eval pod stays invisible to production routing.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.3,#4.3
- **relations:** model swap; dedicated eval pod; deployment path options for the trained adapter
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

<!-- SOURCE: U06_finetuning.md -->
### Model swap / revert functions
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4 (migration 083) + §4.1 "[x] Swap / revert functions deployed"; snapshot_agent used live by migration 110's backup step.
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` — per-agent per-step snapshot-before-swap of the ai_service block in agent_definitions.default_config, with full-table backup as the nuclear option. This is the deployment mechanism a green flywheel-D verdict would use to move an agent from Claude to a local model, and snapshot/revert doubles as the sanctioned def-backup tool for hand-applied migrations. No doc records an actual production model swap having happened.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-3
- **relations:** eval gate; deployable_adapters view; hand-applied migrations
- **verify-later:** migration 083 functions; agent_definitions_backup contents

<!-- SOURCE: U09_adoption.md -->
### Thunder checkpoint & artefact upload to B2 (presigned per-object PUTs)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** "Phases A, B, C BUILT and audited 2026-06-05… Phase D adapter side BUILT; its launcher wiring… is the only code left"; "2026-06-09 update — race fix CONFIRMED in prod, but the loop is being RETIRED for batch… Not yet built."
- **what:** Training VMs are ephemeral and hostile: checkpoints and the final LoRA adapter upload via pre-minted single-object write-only presigned PUT URLs in a manifest (keyed by save-INDEX not global_step; write-once + B2 versioning; final-upload hard-gate so RUN_SH_DONE implies durable, making the monitor's decommission safe); resume via adapter `prepare_resume_url` reusing storage.Client.ListObjects (reuse-before-create — the "genuine adapter gap" claim was wrong). Integrity is the eval gate's job, not the URL scheme's. Two orchestration-layer findings with platform-wide relevance: the send-before-register await race (fixed by preRegisterAwaitedRequest) and the O(K²) cost of awaited loop substeps (each re-persists the full expanded workflow + growing collected_data) — which retired the K-iteration presign loop in favour of a batch `prepare_object_urls` call. (File lives in the adoption folder but belongs to the finetuning/thunder thread.)
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(5).md
- **relations:** finetuning-flywheel Phase 5/C; thunder-training-monitor gating; chassis await/loop mechanics
- **verify-later:** thunder_prepare_object_url_dispatch.go preRegisterAwaitedRequest; batch prepare_object_urls existence; migration 109/110 state

<!-- SOURCE: U12_docs024_archives.md -->
### Anthropic client temperature parameter removed unconditionally
*(merged from 2 independent findings)*
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live dev guide, dated inline "(2026-05-27)": "The Anthropic client no longer sends a temperature parameter on any call... Opus 4.7+ returns a 400 for any non-default temperature."
- **what:** Archived drafts (`old/older1/001h_development_guide_new_agents_v8.md`'s "Extended Thinking Configuration" section and `old/001_development_guide.md`'s "Extended thinking" section) state temperature is stripped only when `budget_tokens` (extended thinking) is set — implying ordinary non-thinking calls still send temperature. The live doc broadens this: because newer Claude Opus models reject any non-default temperature outright, the Anthropic client now omits temperature unconditionally on every call, thinking or not. Temperature remains honoured for other providers (e.g. Ollama) — only the Anthropic client special-cases it.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md; old/001_development_guide.md#"Extended thinking"; docs024_key_docs_latest/001_development_guide(5).md#"Temperature (2026-05-27)"
- **relations:** model-infrastructure (endpoints, provider clients); LLM call logging (`__sent_temperature`)
- **verify-later:** grep the Anthropic client source for unconditional temperature stripping.

<!-- SOURCE: U12_docs024_archives.md -->
### GPU/AI-endpoint scheduling mechanism selection
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "What's Deployed": `ai_endpoint_health` table + view "Applied, verified"; `endpoint-health-checker` agent + scheduled task "Applied".
- **what:** v1 posed four undecided GPU-scheduling options (priority-deprioritisation, boolean flag, health-check auto-discovery, back-to-triage only). v4 resolved this: a single `ai_endpoint_health` table (active vs reactive check modes) *is* the scheduler — dispatch skips claims against unhealthy endpoints; back-to-triage is the reactive safety net beneath it.
- **sources:** old/older1/020_gpu_and_model_infrastructure.md#"GPU Scheduling: Options Under Discussion"; old/older1/020d_gpu_and_model_infrastructure_v4.md#"Architecture: Three Layers"; docs024_key_docs_latest/009_model_infrastructure.md#"What's Deployed"
- **relations:** back-to-triage error handling (AIUnavailableError); model swap/revert functions
- **verify-later:** `ai_endpoint_health` table contents; `endpoint-health-checker` agent definition.

<!-- SOURCE: U12_docs024_archives.md -->
### agent_definitions backup naming convention (unversioned → _preNNN)
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live adds "Naming convention: agent_definitions_backup_YYYYMMDD_pre<NNN>... DO NOT use DROP TABLE IF EXISTS before CREATE TABLE."
- **what:** Archive's convention was a plain `agent_definitions_backup_YYYYMMDD` name with no migration tie and no never-drop rule. Live requires a `_pre<NNN>` suffix tying the backup to the migration it guards and forbids dropping/overwriting an existing backup.
- **sources:** old/009_model_infrastructure.md#"Migration Safety"; docs024_key_docs_latest/009_model_infrastructure.md#"Migration Safety"
- **relations:** model swap/rollback procedure
- **verify-later:** recent migration backup table names for `_preNNN` adoption.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Code-context retrieval infrastructure (analyser adapter + code_symbols)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION" (principles(59)); v4(39) DECISIONS: "Fix direction: migrate code-indexer's analysis step to analyse_repo_local".
- **what:** The chassis's in-cluster code-indexing pipeline: an `analyser-adapter` (Kafka worker, tarball-fetches a repo read-only, runs the shared `internal/analysis` Go-AST walker) feeds `index_code_symbols`, which embeds symbols (nomic-embed-text via the existing `AIService`/`ollama-adapter` seam, reusing the same `rag_index`/`rag_lookup` hybrid pattern as `knowledge_base`) into a sibling `code_symbols` pgvector table (HNSW index, identity-keyed on repo/path/symbol, commit-versioned, hard-deleted not soft-deleted since it's a rebuildable cache). Later found to be indexing a year-old stale tree (fix direction: swap to `analyse_repo_local`, the in-process fetch-and-analyse path already proven in the diagnose workflow).
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section (2026-06-11/12); NOTES_running_synthesis_v4(39).md 2026-07-02 "corpus check result: the index is the blocker" and DECISIONS.
- **relations:** Adapter response envelope contract; B4a embedding-quality finding; diagnosis loop.
- **verify-later:** `code_symbols` table population/freshness, `index_code_symbols` action's current data source.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thunder training-worker probe status taxonomy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** README_worker_statuses: "The worker's probe status_command encodes exactly this taxonomy, and it lines up with the plan."
- **what:** GPU training worker liveness as four probe outcomes: ALIVE (pgrep finds the training run → reset streak), DONE_OK (RUN_SH_DONE marker + adapter_config.json exists → mark_complete → decommission), DONE_FAIL (RUN_SH_FATAL → mark_failed → decommission), GONE_UNKNOWN (process gone, no marker — crash/OOM/reap → bump streak, mark_failed at 3 consecutive unreachable probes).
- **sources:** README_worker_statuses.md
- **relations:** model-infrastructure lifecycle/reaper concepts (docs009 units)
- **verify-later:** the status_command in the thunder worker config

<!-- SECTION-F -->

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM infrastructure (model aliases, call logging, Ollama, RAG)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 001(0) "Implementation Status: LLM Optimization … Deployed and verified: 081/082 migrations, logging … Not yet deployed: ollama.go, rag_actions.go"
- **what:** Cross-cutting LLM infra: short model aliases resolved via `model_aliases.go`; fire-and-forget `llm_call_log` doubling as fine-tune training data; an Ollama CPU adapter serving nomic embeddings and quantized 7B classification; and `rag_lookup`/`rag_index` actions over a shared `knowledge_base` pgvector table with trigram fallback.
- **sources:** WM/001_development_guide(0).md#llm-infrastructure, WM/001_development_guide(0).md#implementation-status-llm-optimization
- **relations:** fine-tuning flywheel; LLM tiering; doc-tree adoption (RAG); Thunder adapter
- **verify-later:** llm_call_log; knowledge_base; migrations 081/082; ai_actions.go createAIClient

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM tiering (large/medium/small/none → Opus/Sonnet/local-70B/Go)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029(1) "Every action that calls an LLM declares its tier … flip medium from Sonnet to local. No action code touched"
- **what:** A cross-cutting `llm_tier` annotation on each LLM call site that the chassis maps to an endpoint via flippable config: Opus for strategy, Sonnet→local-70B for plan partials/audits, Haiku→local for slot-fills, Go for reconciler/validation.
- **sources:** WM/029_site_plan_and_reconciler(1).md#llm-tier-per-call-site, WM/029_site_plan_and_reconciler(1).md#affiliate-product-listings-same-pattern-applied-at-scale
- **relations:** LLM infrastructure; Thunder/local models; reliability cascade
- **verify-later:** llm_tier config → endpoint map (proposed)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM step config shadowing bug (per-object resolution)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §6.6 "a top-level ai_service shadows step-level overrides … Tracked in FOCUS_step_level_llm_config_ignored.md"
- **what:** `ExecuteLLMPromptAction` resolves the `ai_service` object once, taking the first match wholesale even if it lacks `max_tokens`, so a top-level ai_service silently shadows step-level model/max_tokens overrides and `max_tokens` falls back to a hardcoded 2048. Temperature has only one read path and isn't logged.
- **sources:** WM/016_debugging_guide_v2_44.md#6.6, WM/016_debugging_guide_v2_44.md#7
- **relations:** LLM infrastructure; LLM tiering; llm_call_log
- **verify-later:** ExecuteLLMPromptAction; AnthropicClient.GenerateText 2048; FOCUS_step_level_llm_config_ignored.md

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Extended thinking configuration
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) "Extended Thinking Configuration … When budget_tokens is set … the client adds {thinking: {type: enabled, budget_tokens: N}}"
- **what:** Setting `budget_tokens` in an LLM step's `ai_service` config enables Anthropic extended thinking: temperature is removed (API requirement), response parsing skips thinking blocks, latency rises 30–90s.
- **sources:** WM/001_development_guide(0).md#extended-thinking-configuration
- **relations:** LLM infrastructure; model aliases
- **verify-later:** platform/aiservice/anthropic.go thinking block

<!-- SOURCE: U18_sql_for_agents.md -->
### ai_endpoint_health (GPU/model availability gating)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 085 creates table + seeds claude/cpu-ollama/gpu-ollama endpoints; "Checked by claim_work_item before claiming."
- **what:** Health registry for AI endpoints: healthy → work items flow, unhealthy → items wait. Active mode (scheduler pings, per-endpoint interval and ping path incl. 'claude_ping') and reactive mode (failure-driven). Integrates model availability into the dispatch loop's claim decision. Part B adds flywheel columns to llm_call_log (work-item link, prompt variants, verticals, RAG usage).
- **sources:** 085_ai_endpoint_health_checker.sql
- **relations:** build-dispatch-loop claim; Ollama/GPU infrastructure; finetuning flywheel
- **verify-later:** claim_work_item health check; scheduler ping task

<!-- SOURCE: U19_sql_tables_components.md -->
### Agent model-assignment upgrade sweeps
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 Parts 1–2: chief-strategist → opus-4-6; site-planner/domain-research-classifier/domain-strategist/site-classifier → sonnet-4-6; stale claude-3-5-sonnet-20241022 and claude-3-opus refs globally replaced.
- **what:** Model choices live inside agent_definitions.default_config and are upgraded by targeted text-replace UPDATEs, with an explicit tiering philosophy: high-leverage structural deciders get the best models. Also documents the historical model vocabulary embedded in configs.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#Part1-2
- **relations:** agent-definition registry; llm_call_log model_resolved.
- **verify-later:** current model distribution across agent_definitions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Static vs dynamic agent deployment + GPU cost strategy
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** README.095b is a design ("Integration Steps… Add 3 methods to Agent struct") with cost table ($1,440 static GPU vs $20+$50 CPU-router+dynamic); no implementation claim.
- **what:** Same agent code deployed two ways: static agents (pre-deployed Deployments listening on system.agent.* with pattern-subscribed response topics) and dynamic agents (spawned Jobs on job.* topics); IsStaticAgent() switches behaviour. GPU work handled by an always-on cheap CPU router that spawns short-lived GPU workers (TTL auto-terminate) only when needed — claimed 95% GPU cost reduction.
- **sources:** docs001_flow_general/README.095b.gpu_image_static_dynamic_agent_strategy.md
- **relations:** image-generator adapter (the CPU/GPU split case); model-infrastructure GPU/Ollama docs are the living area.
- **verify-later:** GPU_AGENT_STRATEGY env var; whether any router pattern exists in deployments.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Model-tiering by task ("the 3B problem")
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** docs016/003c "The 3B Problem" section: "A 3B model gets ~60-70% of this right. Errors at the leaf level propagate upward... Use the 3B model only for classification"; allocation table routing tasks across Opus/7B/BioMistral/NER/3B.
- **what:** A principled task-to-model allocation doctrine: frontier models only for structure-shaping decisions and top-level synthesis; domain-fine-tuned 7B for analysis; specialised tiny models (biomedical NER) beat general LLMs for structured extraction; embedded 3B only for binary classification; no LLM at all for retrieval. Pipeline design separates cheap structured extraction from semantic interpretation so the strong model gets one focused call. Generalizable beyond the canine project to any large-scale agent workload.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-3B-Problem; docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-Paper-Analysis-Pipeline; docs016_dogs_medicine_pathways/002_project_outline.md
- **relations:** model-infrastructure (Ollama/GPU hosting); finetuning-flywheel; embedded worker-pod models.
- **verify-later:** inference cluster configs; any vLLM/BioMistral deployment manifests.

<!-- SOURCE: U22_recent_small_docs.md -->
### Ollama provider + ollama-adapter
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** "The Ollama adapter is a pod running the Ollama inference server with nomic-embed-text loaded" in one doc; but the RAG deploy handoff still lists "Deploy Ollama adapter" as a not-yet-done next step — conflicting claims across sessions.
- **what:** An `ollama.go` provider implementing the AIService interface (GenerateText via /api/chat, GenerateEmbedding via /api/embeddings) plus an `ollama-adapter` kustomize deployment (third-party `ollama/ollama` image, PVC for model persistence, init container pulling nomic-embed-text, single replica, ClusterIP 11434). Provides local embeddings and a path to self-hosted local LLMs.
- **sources:** docs020.../008_README.md, docs020.../009_023_session_handoff_vertical_architecture(1).md, docs021.../026_implementation_todo_vertical_architecture(2).md#0.3
- **relations:** rag_index, rag_lookup, self-hosted LLM inference, LoRA fine-tuning (GGUF via ollama create)
- **verify-later:** aiservice/ollama.go; deployments/kustomize/services/ollama-adapter/*; createAIClient "ollama" case

<!-- SOURCE: U22_recent_small_docs.md -->
### Model alias upgrades (Sonnet/Opus 4.5–4.6)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 does idempotent `UPDATE agent_definitions SET default_config = replace(... claude-haiku-4-5 ... claude-sonnet-4-5 ...)`; handoff records chief-strategist→opus-4-6, planners/classifiers→sonnet-4-6, stale claude-3.x refs replaced.
- **what:** SQL migrations that upgrade per-agent model references in `agent_definitions.default_config` — planning/strategy agents to the strongest tier (chief-strategist→opus, site/domain planners+classifiers→sonnet), content generation kept on haiku for cost, and all stale `claude-3.x` aliases modernised. Model aliases resolve to API strings; both original and resolved names logged.
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../009_023_session_handoff_vertical_architecture(1).md#done
- **relations:** llm_call_log (logs resolved model), model_aliases.go
- **verify-later:** agent_definitions model values for chief-strategist/site-planner/site-classifier; model_aliases.go 4.6 entries

<!-- SOURCE: U22_recent_small_docs.md -->
### Self-hosted LLM inference (vLLM/GPU at scale)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Self-Hosted LLM Validation — Deploy vLLM or llama.cpp serving a 7B model"; cost tables for a 48-hour million-agent run.
- **what:** A plan to serve 7B models (Mistral/Llama 3/Qwen 2.5) on GPU via vLLM with continuous batching to escape per-token API costs at scale (1,000-2,000 req/min per A100). Bridges to the Ollama/local-model path and the LoRA fine-tuning targets. Estimated hybrid GPU+CPU cost $1,000-3,000 for a 48-hour million-agent burst.
- **sources:** docs021.../015_scaling_analysis.md#phase-2, docs021.../015_scaling_analysis.md#cost-estimates
- **relations:** Ollama provider, LoRA fine-tuning, worker pools
- **verify-later:** any vLLM/GPU inference deployment or stub_llm action

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### RAG best practices (filter-first-then-rank knowledge base)
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** 012_rag_best_practices (dated 2026-03-24, in old/older1); live successor docs/agent_docs/docs020_llm_training_rag/012b_rag_best_practices_v2.md exists.
- **what:** RAG guidance for the site pipeline: always filter `knowledge_base` by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking to avoid cross-vertical contamination; keep RAG at 20-30% of the context window (2-8 examples, quality over quantity); use nomic-embed-text with `search_query:`/`search_document:` task prefixes (recommend nomic-v2-moe upgrade); quality-gate scraped/Claude/human/audit sources; track `embedding_model` and never mix embedding spaces.
- **sources:** old/older1/012_rag_best_practices.md#core-principle, #embedding-model-choice, #avoiding-common-rag-failures
- **relations:** replacement = 012b_rag_best_practices_v2.md; quality flywheel; canine biology knowledge base
- **verify-later:** rag_index/rag_lookup actions filter+prefix; knowledge_base metadata columns

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### GPU/model infrastructure via endpoint health table
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** 020c_gpu_and_model_infrastructure_v3 (2026-03-24, old/older1); "Current Infrastructure State … Not Yet Deployed: ai_endpoint_health table, Back-to-triage error handling, Health check in claim_work_item"; live successor 009_model_infrastructure.md + 020d/020e exist.
- **what:** Three-layer model-availability architecture with `ai_endpoint_health` as the sole GPU scheduler: (L1) dispatch loop checks endpoint health before claiming and skips items whose handler's endpoint is down (item stays `triaged`); (L2) back-to-triage `AIUnavailableError` releases items without counting an attempt and marks the endpoint unhealthy on 401/402; (L3) GPU lifecycle is manual K8s Service creation the health-checker auto-discovers. Claude uses a dual-mode hourly 1-token haiku ping (~$0.002/mo) for auto-recovery. Items never know about models — agent definitions do.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#architecture-three-layers, #standing-decisions
- **relations:** replacement = 009_model_infrastructure.md; model routing via agent_definitions; back-to-triage
- **verify-later:** ai_endpoint_health table; claim_work_item_action.go health check; kafka-scheduler ping task

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Model quality assessment & per-agent model assignment
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 020c "Tested but Not Persistent": Llama 3.3 70B on H100 (classification 8/10, content 9/10, design 7/10), Mistral Small 3 CPU (5/6/3); recommended assignment table routes strategist/webdesign/planner→Claude, classifier/content-writer/triage→Llama70B GPU, briefing→Mistral CPU; cost projection ~$910-990 vs $15-30k all-Claude.
- **what:** Benchmarked model quality per task (Claude reference, Llama 70B near-parity on content/classification, Mistral weak on design) and a per-agent endpoint assignment mapping high-leverage structural work to Claude and bulk content/triage to GPU Llama, projecting ~95% cost reduction at 2000 domains. Model routing is controlled via agent_definitions `ai_service` (swap + snapshot).
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#model-quality-assessment, #cost-projection, #models-to-evaluate
- **relations:** endpoint health table; snapshot/swap/rollback (021); RAG/LoRA flywheel
- **verify-later:** agent_definitions ai_service per agent; model aliases claude-sonnet/opus

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel C — LoRA fine-tuning path (Unsloth QLoRA Llama 3.3 70B)
- **category:** model-infrastructure
- **status-signal:** deployed (first run closed out)
- **status-evidence:** NOTES(39) "Update — 2026-06-05: iter_0 CLOSED OUT … adapter_model.safetensors 828MB … training_run 1cd65dd7 reconciled to complete"; FOCUS(21) §2.5 pipeline scripted
- **what:** The training pipeline: pull dataset from Postgres → Unsloth QLoRA train Llama 3.3 70B Instruct (`unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, 3 epochs, batch 1, grad-accum 8, lr 2e-4, lora_r 16, max_seq 4096) → inference sanity test → LoRA adapter (~150MB). Base 70B chosen because hardware was already available, though 8B was flagged as likely 95% quality at 10% cost. Real run was ~24h (not the scripts' claimed 30-90 min).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150; eval/iter0_eval/lora_iter0_full/README.md (frontmatter)
- **relations:** consumes training_exports datasets; deployed via Phase 5 launcher; produces LoRA iter0 adapter; superseded automation design = Flywheel C Phase 2
- **verify-later:** flywheel_C/02_train_llama_3_3_70b.py, 01_pull_dataset_from_postgres.sh, run.sh

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel C Phase 2 — HTTP-job-server training automation
- **category:** model-infrastructure
- **status-signal:** abandoned
- **status-evidence:** FOCUS(21) §2.5.1 "design locked, not built" (2026-04-23), proposing `model-trainer/model-evaluator/training-flywheel-orchestrator` + `POST /jobs` VM server; superseded in practice by the Kafka/saga Phase 5 chain (NOTES(39) §1) where model-trainer is an orchestrator, not an HTTP-polling agent
- **what:** An abandoned design where a `model-trainer` specialist would POST a dataset to a ~200-line FastAPI-style HTTP job server running on the GPU VM (`POST /jobs`, `GET /jobs/{id}`, download adapter), polling to completion, with three new tables (`model_training_runs`, `model_artefacts`, `model_evaluations`). SSH-remote-exec and a VM Kafka consumer were both explicitly rejected. The actual Phase 5 build instead made the VM credential-free with the chassis driving via thunder-adapter presigned URLs.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5.1; #15 changelog 2026-04-23
- **relations:** superseded by Phase 5 training-launcher + model-trainer saga; the schema names live on as `model_lifecycle.training_runs`
- **verify-later:** model_lifecycle.training_runs; no `/jobs` HTTP server exists in repo

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Model swap / revert mechanism
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4/§4.1 "Swap / revert functions deployed" (migration 083)
- **what:** Per-agent per-step functions `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` that safely snapshot an agent's `ai_service` block in `agent_definitions.default_config` before swapping its LLM (e.g. Claude → a fine-tuned local model), with a full-table backup as the nuclear option. `snapshot_agent`/`revert_agent` are also the sanctioned backup path for agent-definition migrations (used by migration 110).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4; phase5/110_training_launcher_batch_presign(1).sql#0a
- **relations:** the deployment step of the flywheel (swap-if-eval-passes); used by migration snapshotting
- **verify-later:** migration 083; agent_definitions_backup table; snapshot_agent/revert_agent functions

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### AI endpoint health routing (claude / cpu-ollama / gpu-ollama)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.3 table lists gpu-ollama as "currently DOWN, not always-on"; §4.1 "Endpoint health routing deployed"
- **what:** `ai_endpoint_health` table (migration 085) tracks three inference endpoints — Claude (default high-quality), cpu-ollama (embeddings + mistral-small3.1/nomic), gpu-ollama (Llama 70B, future LoRAs). Healthy endpoint → work flows; unhealthy → items wait (back-to-triage). The kafka-scheduler only probes endpoints listed here, which is why the sibling `ollama-eval` pod stays invisible to production routing.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.3, #2.4d
- **relations:** substrate for Flywheel D eval and model swap; ThunderCompute H100 was the intended gpu-ollama
- **verify-later:** ai_endpoint_health table; ollama-adapter, ollama-gpu, ollama-eval services

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Phase 5 training-launcher + model-trainer orchestration chain
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Launcher green … LAUNCH_PID=216 … COMPLETED → notified parent success"
- **what:** The real `training-launcher` (migration 102, replacing a stub) driven by the `model-trainer` orchestrator, which spawns then calls `training-data-preparer → gpu-provisioner → training-launcher` over Kafka/saga. The launcher presigns dataset+scripts, computes checkpoint keys, presigns them, assembles an upload manifest, SSHes it onto the VM, and launches training detached. Two-level await distinction (child's intermediate adapter calls vs the child→parent final notification) is load-bearing.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#1, #5; RUNBOOK_iter0_pretrigger(3).md#6; flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** replaces Flywheel C Phase 2 HTTP-server design; children call thunder-adapter; superseded predecessor = 2026-05-24 Option A handoff
- **verify-later:** agent_definitions training-launcher (id 1223bdc1), model-trainer (94f5a069); migrations 102/109/110

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### setsid detached launch command
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §4 "ssh_exec blocks to command exit (§2), so the launch must return immediately … the SSH channel hits EOF right after echo"; confirmed `LAUNCH_PID=216` (2026-06-09)
- **what:** Because the adapter's `ssh_exec` runs `session.Run` and blocks up to a 5-min timeout for the remote command to exit, the launch command runs the fetch+train chain under `setsid bash -c '…' </dev/null >launch.log 2>&1 &` and echoes `LAUNCH_PID=$!`, so the SSH channel EOFs immediately. An early superseded version used `nohup`; a real bug found later: `write_manifest` (first VM-FS touch) needed `sudo mkdir`/`sudo chown /workspace` (fixed in 109a) because `/` isn't ssh-user-writable.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#4, #update-2026-06-05-deploy-step-2
- **relations:** part of training-launcher; 109a perm fix; run.sh markers
- **verify-later:** thunder_ssh_exec_dispatch.go; ssh_exec_actions.go sshCommandTimeout

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Batch presign (prepare_object_urls) — O(K²) loop retirement (migration 110)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 110 SQL header + NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Full launcher path completed in ~26s … state Version 30 … The O(K²) class is gone"
- **what:** After the race fix, the K=40 per-checkpoint loop still crawled to a halt by iter_9 (~9 min) because every awaited substep re-persisted the entire expanded ~80-substep workflow + growing collected_data/ProcessingHistory — O(K²). Decision: replace the loop + `flatten_checkpoint_urls` with one batch adapter call `prepare_object_urls` (keys[] → ordered presigned_urls[], reusing `DataURLAction.ObjectURL` per key, no new signing path). Migration 110 swaps `presign_checkpoints` to `dispatch_thunder_prepare_object_urls` and re-points `assemble_manifest.checkpoint_urls → ckpt_presign_batch.presigned_urls`, dropping flatten; workflow completed in ~26s.
- **sources:** phase5/110_training_launcher_batch_presign(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-09, #update-2026-06-09-2, #update-2026-06-09-3
- **relations:** the documented structural fallback that became the chosen path; retires the presign loop
- **verify-later:** data_url_actions.go handlePrepareObjectURLs; thunder_prepare_object_urls_dispatch.go; migration 110

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### run.sh RUN_SH markers + set -e durability hard-gate
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** PLAN(4) "run.sh — BUILT 2026-06-05 … set -euo pipefail plus 02_train's final-upload hard-gate … RUN_SH_DONE are only reached on exit 0, which now implies the adapter is in B2"
- **what:** The on-VM launch chain emits grep-able markers (`RUN_SH_START → RUN_SH_STEP setup → RUN_SH_SMOKE_OK → RUN_SH_STEP full_train → RUN_SH_FULL_OK → RUN_SH_DONE`). Because `set -euo pipefail` plus the final-upload raise means DONE only prints on exit 0, `RUN_SH_DONE` came to mean "trained AND uploaded" — the flip that makes the monitor's DONE_OK→decommission safe. `SAVE_STEPS` (cadence) lives in run.sh, default 50 (~1.5h/checkpoint); lowered to 10 for fast tests.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md#run.sh; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4; phase5/NOTES_phase5_training_launcher_running(39).md#8 (Healthy markers)
- **relations:** parsed by thunder-training-monitor probe; gates checkpoint upload path
- **verify-later:** run.sh (bundle at finetuning/scripts/bundle.tar.gz); 02_train --upload-manifest

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### iter0 pre-trigger + Phase B/C/D deploy runbooks
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_iter0_pretrigger(3) §6 trigger `model-trainer` with export 146a9a12; RUNBOOK bcd(7) "One-line summary of the gates" (2026-06)
- **what:** Two operational runbooks. The iter0 pretrigger runbook lists the gates to reach the first automated training launch (deploy adapter+chassis, upload the scripts bundle, adapter round-trip, cost gate, a gpu-provisioner smoke test of the D4 topic path, then trigger model-trainer). The Phase B/C/D deploy runbook stages the checkpoint-upload rollout: apply 109 → re-pack/re-upload bundle → Tier-2 short launch (B+C integration, SAVE_STEPS low) → resume (blocked on D3+migration) → enable the monitor last. Both hard-code b2 CLI (not aws) and the "verify positive evidence, complete≠succeeded" discipline.
- **sources:** phase5/RUNBOOK_iter0_pretrigger(3).md; phase5/RUNBOOK_phase_b_c_d_deploy(7).md; phase5/NOTES_phase5_training_launcher_running(39).md#7 (Pre-trigger gates)
- **relations:** operationalise the launcher + checkpoint upload + monitor; export a8484922 is the do-not-use trap
- **verify-later:** migrations 109/109a/109b; scheduled_tasks thunder-training-monitor enable step

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
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

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### RAG pipeline deployment (ollama-adapter, rag_lookup/rag_index, knowledge_base)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live repo confirms `platform/orchestration/actions/rag_actions.go` and other RAG-related action files exist and are registered, matching the deployment bundle's file-placement manifest.
- **what:** A deployment bundle for adding retrieval-augmented generation to the chassis: a new `ollama-adapter` k8s service (own kustomize base+overlay, `ollama/ollama` image, an idempotent init container that pulls an embedding model onto a PVC — `nomic-embed-text` ~300MB, sized for a 8Gi memory limit that also leaves room for future 7B models), two new SQL migrations (`llm_call_log` with a stats view, and `knowledge_base` for RAG storage), two new registered actions (`rag_index` — chunks and embeds content into `knowledge_base`; `rag_lookup` — embeds a query and returns top-k matches), plus a nullable-helpers Go package and patches to `ai_actions.go`/`registry.go`/`anthropic.go` (adds LLM call timing/token logging and an `ollama` case to `createAIClient`).
- **sources:** docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md
- **relations:** model-infrastructure (009 anchor); contextkit's embed.go (same Ollama-embeddings-endpoint pattern reused independently)
- **verify-later:** platform/orchestration/actions/rag_actions.go, deployments/kustomize/services/ollama-adapter/, llm_call_log / knowledge_base tables

<!-- SOURCE: U25_leopardess_social.md -->
### Per-workflow-step model routing (data-sovereignty mechanism)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** AUDIT P1/P2 (2026-07-10): "TRUE — ExecuteLLMPromptAction resolves ai_service in a three-tier lookup, tier 2 being workflow.steps[step].config.ai_service"; "ollama-adapter … ClusterIP only".
- **what:** Model/provider is selectable per workflow step with no new code (three-tier ai_service resolution), with live swap tooling (swap_agent_model, migration 083). A self-hosted step genuinely never leaves the cluster (stock ollama image, ClusterIP, in-cluster calls). This underpins the honest claim: "steps that touch your data can run on infrastructure you control; only steps that don't need to leave call a foundation model."
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6
- **relations:** text-provider wiring reality; data-sovereignty positioning; no-tenant-isolation fact
- **verify-later:** ai_actions.go ExecuteLLMPromptAction; 021_model_swap_and_rollback.sql; ollama-adapter service spec

<!-- SOURCE: U25_leopardess_social.md -->
### Text-provider wiring reality (two providers end-to-end)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** AUDIT P3: "createAIClient switch: anthropic and ollama only; openai is a stubbed error … 'Mistral' is … run through the same self-hosted Ollama pod."
- **what:** Only Anthropic and Ollama work end-to-end for text; nothing else is wired. Imagery is broader (Gemini + Stability, routed by kind not config). The news pipeline has a separate hand-rolled provider path (xAI /v1/responses with web_search+x_search, OpenAI, Perplexity) bypassing the generic AIService entirely. A real constraint on any "model choice" marketing claim (RUNBOOK landmine 12).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P3; docs/leopardessconsulting/RUNBOOK.md#landmine-12
- **relations:** per-step model routing; news-feed-pipeline (separate provider path)
- **verify-later:** createAIClient switch; feed_actions.go provider paths

<!-- SOURCE: U25_leopardess_social.md -->
### llama3.3:70b trained but never used for inference; dynamic GPU provisioning
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** AUDIT P4: "one complete run (2026-06-03→04) … No agent_definitions row points at llama3.3:70b — trained and tested, never used for production inference. TODO logged in 009_model_infrastructure.md."
- **what:** The larger self-hosted model exists as a completed training run; GPU provisioning is genuinely dynamic (thunder_instances, ThunderCompute, decommissioned per run) but experimental. Wiring it to production inference would strengthen the data-sovereignty positioning; deliberately logged in the model-infrastructure home doc rather than duplicated.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6; docs/leopardessconsulting/RUNBOOK.md#Reference
- **relations:** finetuning-flywheel; per-step model routing
- **verify-later:** model_lifecycle.training_runs; agent_definitions model references

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model aliases and the model selection strategy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(5) rules; 009 per-step table; 101_switch_to_haiku.sql records live state 2026-04-10
- **what:** Agent definitions use short aliases (claude-sonnet-4-6, claude-haiku-4-5) resolved by model_aliases.go; sonnet is the default for LLM steps, haiku for routing, opus for chief-strategist/planner, ollama for fine-tuned classification. 101 SQL is a bulk cost lever switching all agents to haiku with a RESTORE section.
- **sources:** 001(5)#LLM Infrastructure; 009#Model Swap; 101_switch_to_haiku.sql
- **relations:** swap_agent_model; LLM tiering (029)
- **verify-later:** model_aliases.go; per-step ai_service in agent_definitions

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Ollama adapter (CPU embeddings + local classification)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009: "Ollama adapter on CPU cluster (2 replicas, mistral-small3.1 + nomic-embed-text)" checked done
- **what:** Permanent CPU adapter serving nomic-embed-text embeddings (~50-100ms) and quantized small models for classification (10-30s acceptable per-build). Same AIService interface as Anthropic incl. token-usage write-backs. Not for content generation or <2s latency.
- **sources:** 001(5)#Ollama adapter; 009#Implementation Status
- **relations:** RAG actions; endpoint health; fine-tuning path
- **verify-later:** ollama.go; ollama-adapter deployment

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Multi-endpoint model routing with ai_endpoint_health as the GPU scheduler
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 009: tables/functions applied+verified; the three Go patches (fast-fail, claim-gate, release-without-attempt) listed under "Next Deploy"; active pinging "starts after patches deployed"
- **what:** Endpoints (Claude API, CPU/GPU Ollama) tracked in ai_endpoint_health; healthy → items flow, unhealthy → items wait (no fallback chains — quality over speed; priority means importance only; items don't know about models). ClaimWorkItem checks handler's endpoint health before claiming; AIUnavailableError triggers reactive health update + release-to-triaged without attempt increment; Claude health dual-mode (reactive 402/401 + hourly 1-token ping).
- **sources:** 009#Decisions Made, #Health Check Architecture; 022 SQL
- **relations:** back-to-triage; endpoint-health-checker agent
- **verify-later:** were the three patches deployed since 2026-03-25

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model swap/snapshot/revert control plane (migration 083)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 021 SQL applied per 009; used operationally (016 §6.1 conventions)
- **what:** snapshot_agent()/swap_agent_model()/revert_agent() + agent_snapshots view make agent_definitions the model-routing control plane: per-step ai_service swaps with automatic snapshot, one-call revert. Post-migration snapshots live in agent_definitions_backup (snapshot_taken_at discriminator, restored_at audit trail); the legacy in-table is_snapshot rows caused a documented family of contamination/misroute bugs.
- **sources:** 021_model_swap_and_rollback.sql; 009#Model Swap Procedure; 016 §6.1/§9 snapshot bugs
- **relations:** backup naming discipline; LLM config shadowing (step-level swaps shadowed by top-level ai_service)
- **verify-later:** agent_definitions_backup schema

<!-- SOURCE: U01_docs024_numbered_core.md -->
### LLM tiering (large/medium/small/none) + cluster-then-slot-fill scaling pattern
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029: "the chassis routes" described as design with llm_tier annotation to add; flip-to-local gated on Thunder health
- **what:** Every LLM call site declares a tier; chassis maps tier→endpoint via flippable config (large=Opus strategy/briefing; medium=Sonnet→local-70B for plan partials/audits; small=Haiku for slot-fills; none=deterministic Go). Product-listing scale: facts from feeds (Go), cluster ~10k products into ~20-50 groups algorithmically, one medium call per cluster for framing, small slot-fill per product.
- **sources:** 029#LLM tier per call site, #Affiliate/product listings
- **relations:** model aliases; batch queue routing
- **verify-later:** any llm_tier config keys in defs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### AI endpoint health routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Endpoint health routing deployed" [x] (2026-04); gpu-ollama "currently DOWN, not always-on"
- **what:** `ai_endpoint_health` (migration 085) tracks claude / cpu-ollama / gpu-ollama endpoints; healthy endpoint → work claims flow, unhealthy → items wait (back-to-triage). No separate batch scheduler for GPU: it's either healthy or not. Kafka-scheduler probes only endpoints listed here, so unlisted pods (ollama-eval) are invisible to production routing by design.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.3, #2.4d
- **relations:** Flywheel D dedicated eval pod; rate-limit transient classification
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Per-agent model swap / revert
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Swap / revert functions deployed" [x] (migration 083)
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` safely snapshot an agent's `ai_service` block in agent_definitions.default_config before swapping model, per-agent per-step; full-table backup remains as the nuclear option. The mechanism the flywheel's deployment decision hangs on.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4; HANDOFF_2026-05-26 (snapshot_agent used before jsonb_set edit)
- **relations:** training-flywheel-orchestrator conditional swap
- **verify-later:** migration 083 functions in DB

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Ollama CPU operations envelope
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; memory rule from 2026-04-23 OOM incident
- **what:** Hard-won ops facts for CPU Ollama: RollingUpdate + RWO PVC deadlocks (use Recreate); OLLAMA_LOAD_TIMEOUT=10m / KEEP_ALIVE=30m; cold load ~45s for 14GB models; pod memory limit ≥ model size + 8–12GiB (cgroup, not host, is what Ollama constrains against); /api/chat not /api/generate; throughput ~150 tok/s prompt, ~2.5 tok/s generation for mistral-small3.1 Q4 on 8 CPU cores.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4c, #14 "Ollama specifics"
- **relations:** dedicated eval pod; endpoint health
- **verify-later:** kustomize base for ollama-adapter

<!-- SOURCE: U06_finetuning.md -->
### thunder-training-monitor (periodic probe + reconcile + release)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** NOTES(45) 2026-06-04: "training-monitor VERIFIED live (both paths)… Terminal/decommission branch still never run live… Not enabled"; schedule inserted DISABLED (migration 108).
- **what:** A second periodic lifecycle agent beside the reaper: the `thunder-training-monitor` orchestrator runs `find_active_training_instances` then loops spawn_worker→call_worker per running training box (deliberately NOT the reaper's scheduler-pre_query shape, which merges only the first row per tick and would starve newer instances behind ALIVE boxes). Each `thunder-training-monitor-worker` probes via the adapter's `ssh_get_status` with a status_command that classifies run.sh markers into ALIVE | DONE_OK | DONE_FAIL | GONE_UNKNOWN (plus reachable:false as a valid answer), routes via classifier `next_step` override, reconciles `model_lifecycle.training_runs`, counts consecutive unreachable probes (≥3 → lost → decommission an unreachable-but-billing box), and on terminal verdicts releases the box through the shared idempotent decommission. Built as migrations 106/107/108 + 5 chassis actions. ALIVE path and orchestrator fan-out verified live; the terminal/decommission branch has never fired.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04,#update-2026-06-09-6; working/phase5/107_thunder_training_monitor_worker.sql (header)
- **relations:** thunder-reaper (responsibility split); run.sh markers; monitor enablement gate; reply-topic resolution bug (found here)
- **verify-later:** scheduled_tasks row enabled state; migrations 106-108; the 5 actions in registry.go

<!-- SOURCE: U06_finetuning.md -->
### Monitor/reaper responsibility split
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04: "Decision: build a separate thunder-training-monitor, NOT bolted into the time-reaper… the time-reaper is the last-line cost backstop and must stay dead-simple/dependency-free".
- **what:** Two periodic agents with deliberately distinct dependency profiles: the reaper (cost backstop) is pure DB + Thunder and must work even when the adapter is down; the monitor (completion-side) depends on adapter + SSH. They overlap only in calling the shared idempotent `decommission_instance`. The monitor exists because the launcher returns long before training ends (detached run) so completion can't be a workflow await.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md
- **relations:** thunder-reaper; thunder-training-monitor; orphan-sweep TODO (third member: boxes whose Thunder instance vanished)
- **verify-later:** both scheduled_tasks rows and their concurrency_group

<!-- SOURCE: U06_finetuning.md -->
### Thunder Compute adapter (provision/decommission lifecycle)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 2026-05-12(1) table: phases 2–3.5 "✅ Deployed and verified end-to-end"; FOCUS(25) §14 "Provision loop verified end-to-end (2026-05-22)".
- **what:** A Kafka adapter (`system.adapter.thunder.requests`) wrapping the Thunder Compute GPU API and owning its credentials. `provision_instance`: spend pre-check → ed25519 keypair → API create (public_key sent) → k8s Secret persist → WaitForRunning poll → INSERT thunder_instances with retry → compensating cleanup (fresh context) on partial failure. `decommission_instance`: lookup by provisioning_id or thunder identifier → atomic `decommissioning` transition as idempotency anchor → 404-tolerant API + Secret deletes → cost computed from running_since × snapshotted hourly rate. Error classification maps denial→unrecoverable, infra→recoverable. Includes hard-won API shape knowledge: base URL :8443/v1, lowercase gpu_type enums, camelCase string-numbers in responses vs snake_case ints in requests, recycled numeric ids requiring a partial unique index on live rows, real template names (`base`, not the OpenAPI example).
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md
- **relations:** gpu-provisioner; reaper; ssh_exec; presigned data plane; adapter design guide
- **verify-later:** internal/adapters/thunder/*; migrations 025/028/029; thunder_instances schema

<!-- SOURCE: U06_finetuning.md -->
### thunder-reaper scheduled task and per-instance uptime deadline
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) 3.5: "✅ Deployed and verified end-to-end (2026-05-14): synthetic row … picked up within 30s"; NOTES(45) 2026-06-04 live rescue of run fabfd7fa.
- **what:** A 15-min scheduled task whose pre_query finds `running` instances past `max_uptime_hours` and dispatches the idempotent decommission (one per tick, LIMIT 1). The deadline is OURS not Thunder's — computed as running_since + the per-row `max_uptime_hours` (default 18h; training provisions get 18h) — so a mid-train cap overrun can be rescued by bumping the single row's max_uptime_hours (done live: 18→48h when the 24h iter_0 train would have been reaped at hour 18, which with save_strategy=no meant total loss). Reaper reason strings are meaningful text for post-mortems.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#3.5; working/flywheel_docs/README.md#3.5-delivered; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** monitor/reaper split; spend gating; scheduler pre_query single-row semantics
- **verify-later:** migration 028; scheduled_tasks row; thunder_instances.max_uptime_hours column

<!-- SOURCE: U06_finetuning.md -->
### Thunder spend gating (DB-side provision check)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Spend gating lives in DB, not API"; NOTES(45) 2026-06-03 cost-gate check (cap 30, estimate 20, clears with $9 headroom).
- **what:** Before every create, the adapter consults the `thunder_provision_check` view: decommissioned 24h spend + running estimated spend + `estimated_new_run_cost_usd` must stay under `thunder_config.daily_cap_usd`. Operational learnings: keep the per-run estimate realistic (~$20 for a 9h+ A100 run — a $2 test estimate lets doomed runs through; a $25 default blocks legitimate tests); the 24h window is rolling so heavy test days trip the cap on legitimate past spend (raise the cap for the session, don't delete accurate rows).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-172x,#7
- **relations:** thunder adapter provision; reaper; cost anchors from iter_0
- **verify-later:** thunder_provision_check view definition; thunder_config values

<!-- SOURCE: U06_finetuning.md -->
### Orphan-sweep for stale live thunder_instances rows
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "TODO (deferred 2026-05-24) — orphan-sweep for stale live rows… Agreed design (not yet built)".
- **what:** Out-of-band deletions (manual `tnr delete`) leave DB rows `running` forever; because live rows hold the recycled Thunder id in a partial unique index, a stale row blocks the next provision with a duplicate-key error (bit on 2026-05-24). Agreed design: a `sweep_orphans` adapter action computes (live DB rows) minus (Thunder's live list) and dispatches the idempotent decommission per orphan, run as a 15–30min scheduled task sharing the reaper's concurrency group, with a safety guard never to sweep on a failed/partial Thunder list. Interim: manual row reconciliation after any manual delete.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics(TODO)
- **relations:** reaper (time-based) and monitor (completion-based) — this is the third, existence-based leg
- **verify-later:** whether sweep_orphans exists anywhere

<!-- SOURCE: U06_finetuning.md -->
### Adapter-managed SSH access to GPU boxes (ed25519 keys in k8s Secrets)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "RESOLVED & VERIFIED (Phase 4, 2026-05-24) — SSH connection mechanism + ssh_exec/ssh_get_status" with production verification detail.
- **what:** The adapter generates its own ed25519 keypair per provision, stores the private half in Secret `thunder-ssh-<db_row_id>` (deterministic name so orphan Secrets are reapable), sends the public half on create. `ssh_exec`/`ssh_get_status` dial `instance_ip:ssh_port` directly via x/crypto/ssh as user `ubuntu` (NOT root, despite Thunder's own ssh_command string), with a wait-for-sshd retry (~90s) because RUNNING precedes sshd. The port is the list-endpoint's `port` field, captured into thunder_instances.ssh_port. `reachable:false` is a valid answer, not an error — the probe primitive the monitor builds on. Manual-ops corollary: operators can extract the key from the Secret to watch train.log directly (StrictHostKeyChecking=no needed because Thunder recycles IPs).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/ssh_probe.sh (header); working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x
- **relations:** monitor probe; setsid launch; RBAC resourceNames trap (Secret permissions)
- **verify-later:** internal/adapters/thunder/ssh/*; the RBAC Role verbs

<!-- SOURCE: U06_finetuning.md -->
### Thunder Prototyping vs Production mode economics
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 lesson 8: "Prototyping (TGV virtualised) worked fine for 70B inference… Phase 2 should default to Prototyping for inference, Production for training (unverified that Prototyping handles long training runs well)."
- **what:** Thunder's Production mode ($1.79/hr A100 80GB) vs Prototyping ($0.78/hr, TGV-virtualised). Verified: Prototyping is fine for 70B inference. Unverified: whether virtualisation overhead degrades long QLoRA training runs enough to cancel the ~55% saving — flagged as an iter_1 experiment, never run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(8); working/phase2/README.md
- **relations:** cost anchors; gpu-provisioner defaults
- **verify-later:** provision defaults (mode) in gpu-provisioner/adapter config

<!-- SOURCE: U06_finetuning.md -->
### Ollama CPU adapter operational rules
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4c "RWO PVC rolling-restart deadlock on ollama-adapter — resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; §14 Ollama specifics list.
- **what:** Hard-won ops rules for CPU Ollama: Recreate (not RollingUpdate) deployment strategy because the PVC is RWO (classic new-pod-can't-mount deadlock); `OLLAMA_LOAD_TIMEOUT=10m` + `OLLAMA_KEEP_ALIVE=30m` (default 60s load timeout killed first inference after cold start — 14.4GB model loads in ~45s); pod memory limit ≥ model file size + 8–12GiB headroom (Ollama reads host /proc/meminfo but is constrained by the cgroup — misleading OOM messages); chassis calls `/api/chat` not `/api/generate`; measured CPU throughput ~150 tok/s prompt, ~2.5 tok/s generation on mistral-small3.1.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4c,#2.4d,#14
- **relations:** ai_endpoint_health; dedicated eval pod; CPU eval abandonment
- **verify-later:** ollama-adapter kustomize base

<!-- SOURCE: U06_finetuning.md -->
### ai_endpoint_health inference routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.3 endpoint table + §4.1 "[x] Endpoint health routing deployed"; gpu-ollama noted "currently DOWN, not always-on" and still DOWN in HANDOFF 2026-05-08.
- **what:** Three endpoints tracked in `ai_endpoint_health`: claude (default quality), cpu-ollama (embeddings + small models), gpu-ollama (70B/LoRAs — persistently down through this doc set). Healthy endpoint → work claims flow; unhealthy → items wait/back-to-triage, so GPU availability gates work without a separate batch scheduler. The kafka-scheduler only probes endpoints listed here — which is also why the dedicated eval pod stays invisible to production routing.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.3,#4.3
- **relations:** model swap; dedicated eval pod; deployment path options for the trained adapter
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

<!-- SOURCE: U06_finetuning.md -->
### Model swap / revert functions
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4 (migration 083) + §4.1 "[x] Swap / revert functions deployed"; snapshot_agent used live by migration 110's backup step.
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` — per-agent per-step snapshot-before-swap of the ai_service block in agent_definitions.default_config, with full-table backup as the nuclear option. This is the deployment mechanism a green flywheel-D verdict would use to move an agent from Claude to a local model, and snapshot/revert doubles as the sanctioned def-backup tool for hand-applied migrations. No doc records an actual production model swap having happened.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-3
- **relations:** eval gate; deployable_adapters view; hand-applied migrations
- **verify-later:** migration 083 functions; agent_definitions_backup contents

<!-- SOURCE: U09_adoption.md -->
### Thunder checkpoint & artefact upload to B2 (presigned per-object PUTs)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** "Phases A, B, C BUILT and audited 2026-06-05… Phase D adapter side BUILT; its launcher wiring… is the only code left"; "2026-06-09 update — race fix CONFIRMED in prod, but the loop is being RETIRED for batch… Not yet built."
- **what:** Training VMs are ephemeral and hostile: checkpoints and the final LoRA adapter upload via pre-minted single-object write-only presigned PUT URLs in a manifest (keyed by save-INDEX not global_step; write-once + B2 versioning; final-upload hard-gate so RUN_SH_DONE implies durable, making the monitor's decommission safe); resume via adapter `prepare_resume_url` reusing storage.Client.ListObjects (reuse-before-create — the "genuine adapter gap" claim was wrong). Integrity is the eval gate's job, not the URL scheme's. Two orchestration-layer findings with platform-wide relevance: the send-before-register await race (fixed by preRegisterAwaitedRequest) and the O(K²) cost of awaited loop substeps (each re-persists the full expanded workflow + growing collected_data) — which retired the K-iteration presign loop in favour of a batch `prepare_object_urls` call. (File lives in the adoption folder but belongs to the finetuning/thunder thread.)
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(5).md
- **relations:** finetuning-flywheel Phase 5/C; thunder-training-monitor gating; chassis await/loop mechanics
- **verify-later:** thunder_prepare_object_url_dispatch.go preRegisterAwaitedRequest; batch prepare_object_urls existence; migration 109/110 state

<!-- SOURCE: U12_docs024_archives.md -->
### Anthropic client temperature parameter removed unconditionally
*(merged from 2 independent findings)*
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live dev guide, dated inline "(2026-05-27)": "The Anthropic client no longer sends a temperature parameter on any call... Opus 4.7+ returns a 400 for any non-default temperature."
- **what:** Archived drafts (`old/older1/001h_development_guide_new_agents_v8.md`'s "Extended Thinking Configuration" section and `old/001_development_guide.md`'s "Extended thinking" section) state temperature is stripped only when `budget_tokens` (extended thinking) is set — implying ordinary non-thinking calls still send temperature. The live doc broadens this: because newer Claude Opus models reject any non-default temperature outright, the Anthropic client now omits temperature unconditionally on every call, thinking or not. Temperature remains honoured for other providers (e.g. Ollama) — only the Anthropic client special-cases it.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md; old/001_development_guide.md#"Extended thinking"; docs024_key_docs_latest/001_development_guide(5).md#"Temperature (2026-05-27)"
- **relations:** model-infrastructure (endpoints, provider clients); LLM call logging (`__sent_temperature`)
- **verify-later:** grep the Anthropic client source for unconditional temperature stripping.

<!-- SOURCE: U12_docs024_archives.md -->
### GPU/AI-endpoint scheduling mechanism selection
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "What's Deployed": `ai_endpoint_health` table + view "Applied, verified"; `endpoint-health-checker` agent + scheduled task "Applied".
- **what:** v1 posed four undecided GPU-scheduling options (priority-deprioritisation, boolean flag, health-check auto-discovery, back-to-triage only). v4 resolved this: a single `ai_endpoint_health` table (active vs reactive check modes) *is* the scheduler — dispatch skips claims against unhealthy endpoints; back-to-triage is the reactive safety net beneath it.
- **sources:** old/older1/020_gpu_and_model_infrastructure.md#"GPU Scheduling: Options Under Discussion"; old/older1/020d_gpu_and_model_infrastructure_v4.md#"Architecture: Three Layers"; docs024_key_docs_latest/009_model_infrastructure.md#"What's Deployed"
- **relations:** back-to-triage error handling (AIUnavailableError); model swap/revert functions
- **verify-later:** `ai_endpoint_health` table contents; `endpoint-health-checker` agent definition.

<!-- SOURCE: U12_docs024_archives.md -->
### agent_definitions backup naming convention (unversioned → _preNNN)
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live adds "Naming convention: agent_definitions_backup_YYYYMMDD_pre<NNN>... DO NOT use DROP TABLE IF EXISTS before CREATE TABLE."
- **what:** Archive's convention was a plain `agent_definitions_backup_YYYYMMDD` name with no migration tie and no never-drop rule. Live requires a `_pre<NNN>` suffix tying the backup to the migration it guards and forbids dropping/overwriting an existing backup.
- **sources:** old/009_model_infrastructure.md#"Migration Safety"; docs024_key_docs_latest/009_model_infrastructure.md#"Migration Safety"
- **relations:** model swap/rollback procedure
- **verify-later:** recent migration backup table names for `_preNNN` adoption.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Code-context retrieval infrastructure (analyser adapter + code_symbols)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION" (principles(59)); v4(39) DECISIONS: "Fix direction: migrate code-indexer's analysis step to analyse_repo_local".
- **what:** The chassis's in-cluster code-indexing pipeline: an `analyser-adapter` (Kafka worker, tarball-fetches a repo read-only, runs the shared `internal/analysis` Go-AST walker) feeds `index_code_symbols`, which embeds symbols (nomic-embed-text via the existing `AIService`/`ollama-adapter` seam, reusing the same `rag_index`/`rag_lookup` hybrid pattern as `knowledge_base`) into a sibling `code_symbols` pgvector table (HNSW index, identity-keyed on repo/path/symbol, commit-versioned, hard-deleted not soft-deleted since it's a rebuildable cache). Later found to be indexing a year-old stale tree (fix direction: swap to `analyse_repo_local`, the in-process fetch-and-analyse path already proven in the diagnose workflow).
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section (2026-06-11/12); NOTES_running_synthesis_v4(39).md 2026-07-02 "corpus check result: the index is the blocker" and DECISIONS.
- **relations:** Adapter response envelope contract; B4a embedding-quality finding; diagnosis loop.
- **verify-later:** `code_symbols` table population/freshness, `index_code_symbols` action's current data source.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Thunder training-worker probe status taxonomy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** README_worker_statuses: "The worker's probe status_command encodes exactly this taxonomy, and it lines up with the plan."
- **what:** GPU training worker liveness as four probe outcomes: ALIVE (pgrep finds the training run → reset streak), DONE_OK (RUN_SH_DONE marker + adapter_config.json exists → mark_complete → decommission), DONE_FAIL (RUN_SH_FATAL → mark_failed → decommission), GONE_UNKNOWN (process gone, no marker — crash/OOM/reap → bump streak, mark_failed at 3 consecutive unreachable probes).
- **sources:** README_worker_statuses.md
- **relations:** model-infrastructure lifecycle/reaper concepts (docs009 units)
- **verify-later:** the status_command in the thunder worker config

<!-- SECTION-F -->

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM infrastructure (model aliases, call logging, Ollama, RAG)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 001(0) "Implementation Status: LLM Optimization … Deployed and verified: 081/082 migrations, logging … Not yet deployed: ollama.go, rag_actions.go"
- **what:** Cross-cutting LLM infra: short model aliases resolved via `model_aliases.go`; fire-and-forget `llm_call_log` doubling as fine-tune training data; an Ollama CPU adapter serving nomic embeddings and quantized 7B classification; and `rag_lookup`/`rag_index` actions over a shared `knowledge_base` pgvector table with trigram fallback.
- **sources:** WM/001_development_guide(0).md#llm-infrastructure, WM/001_development_guide(0).md#implementation-status-llm-optimization
- **relations:** fine-tuning flywheel; LLM tiering; doc-tree adoption (RAG); Thunder adapter
- **verify-later:** llm_call_log; knowledge_base; migrations 081/082; ai_actions.go createAIClient

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM tiering (large/medium/small/none → Opus/Sonnet/local-70B/Go)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029(1) "Every action that calls an LLM declares its tier … flip medium from Sonnet to local. No action code touched"
- **what:** A cross-cutting `llm_tier` annotation on each LLM call site that the chassis maps to an endpoint via flippable config: Opus for strategy, Sonnet→local-70B for plan partials/audits, Haiku→local for slot-fills, Go for reconciler/validation.
- **sources:** WM/029_site_plan_and_reconciler(1).md#llm-tier-per-call-site, WM/029_site_plan_and_reconciler(1).md#affiliate-product-listings-same-pattern-applied-at-scale
- **relations:** LLM infrastructure; Thunder/local models; reliability cascade
- **verify-later:** llm_tier config → endpoint map (proposed)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### LLM step config shadowing bug (per-object resolution)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §6.6 "a top-level ai_service shadows step-level overrides … Tracked in FOCUS_step_level_llm_config_ignored.md"
- **what:** `ExecuteLLMPromptAction` resolves the `ai_service` object once, taking the first match wholesale even if it lacks `max_tokens`, so a top-level ai_service silently shadows step-level model/max_tokens overrides and `max_tokens` falls back to a hardcoded 2048. Temperature has only one read path and isn't logged.
- **sources:** WM/016_debugging_guide_v2_44.md#6.6, WM/016_debugging_guide_v2_44.md#7
- **relations:** LLM infrastructure; LLM tiering; llm_call_log
- **verify-later:** ExecuteLLMPromptAction; AnthropicClient.GenerateText 2048; FOCUS_step_level_llm_config_ignored.md

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Extended thinking configuration
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) "Extended Thinking Configuration … When budget_tokens is set … the client adds {thinking: {type: enabled, budget_tokens: N}}"
- **what:** Setting `budget_tokens` in an LLM step's `ai_service` config enables Anthropic extended thinking: temperature is removed (API requirement), response parsing skips thinking blocks, latency rises 30–90s.
- **sources:** WM/001_development_guide(0).md#extended-thinking-configuration
- **relations:** LLM infrastructure; model aliases
- **verify-later:** platform/aiservice/anthropic.go thinking block

<!-- SOURCE: U18_sql_for_agents.md -->
### ai_endpoint_health (GPU/model availability gating)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 085 creates table + seeds claude/cpu-ollama/gpu-ollama endpoints; "Checked by claim_work_item before claiming."
- **what:** Health registry for AI endpoints: healthy → work items flow, unhealthy → items wait. Active mode (scheduler pings, per-endpoint interval and ping path incl. 'claude_ping') and reactive mode (failure-driven). Integrates model availability into the dispatch loop's claim decision. Part B adds flywheel columns to llm_call_log (work-item link, prompt variants, verticals, RAG usage).
- **sources:** 085_ai_endpoint_health_checker.sql
- **relations:** build-dispatch-loop claim; Ollama/GPU infrastructure; finetuning flywheel
- **verify-later:** claim_work_item health check; scheduler ping task

<!-- SOURCE: U19_sql_tables_components.md -->
### Agent model-assignment upgrade sweeps
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 Parts 1–2: chief-strategist → opus-4-6; site-planner/domain-research-classifier/domain-strategist/site-classifier → sonnet-4-6; stale claude-3-5-sonnet-20241022 and claude-3-opus refs globally replaced.
- **what:** Model choices live inside agent_definitions.default_config and are upgraded by targeted text-replace UPDATEs, with an explicit tiering philosophy: high-leverage structural deciders get the best models. Also documents the historical model vocabulary embedded in configs.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#Part1-2
- **relations:** agent-definition registry; llm_call_log model_resolved.
- **verify-later:** current model distribution across agent_definitions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Static vs dynamic agent deployment + GPU cost strategy
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** README.095b is a design ("Integration Steps… Add 3 methods to Agent struct") with cost table ($1,440 static GPU vs $20+$50 CPU-router+dynamic); no implementation claim.
- **what:** Same agent code deployed two ways: static agents (pre-deployed Deployments listening on system.agent.* with pattern-subscribed response topics) and dynamic agents (spawned Jobs on job.* topics); IsStaticAgent() switches behaviour. GPU work handled by an always-on cheap CPU router that spawns short-lived GPU workers (TTL auto-terminate) only when needed — claimed 95% GPU cost reduction.
- **sources:** docs001_flow_general/README.095b.gpu_image_static_dynamic_agent_strategy.md
- **relations:** image-generator adapter (the CPU/GPU split case); model-infrastructure GPU/Ollama docs are the living area.
- **verify-later:** GPU_AGENT_STRATEGY env var; whether any router pattern exists in deployments.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Model-tiering by task ("the 3B problem")
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** docs016/003c "The 3B Problem" section: "A 3B model gets ~60-70% of this right. Errors at the leaf level propagate upward... Use the 3B model only for classification"; allocation table routing tasks across Opus/7B/BioMistral/NER/3B.
- **what:** A principled task-to-model allocation doctrine: frontier models only for structure-shaping decisions and top-level synthesis; domain-fine-tuned 7B for analysis; specialised tiny models (biomedical NER) beat general LLMs for structured extraction; embedded 3B only for binary classification; no LLM at all for retrieval. Pipeline design separates cheap structured extraction from semantic interpretation so the strong model gets one focused call. Generalizable beyond the canine project to any large-scale agent workload.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-3B-Problem; docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-Paper-Analysis-Pipeline; docs016_dogs_medicine_pathways/002_project_outline.md
- **relations:** model-infrastructure (Ollama/GPU hosting); finetuning-flywheel; embedded worker-pod models.
- **verify-later:** inference cluster configs; any vLLM/BioMistral deployment manifests.

<!-- SOURCE: U22_recent_small_docs.md -->
### Ollama provider + ollama-adapter
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** "The Ollama adapter is a pod running the Ollama inference server with nomic-embed-text loaded" in one doc; but the RAG deploy handoff still lists "Deploy Ollama adapter" as a not-yet-done next step — conflicting claims across sessions.
- **what:** An `ollama.go` provider implementing the AIService interface (GenerateText via /api/chat, GenerateEmbedding via /api/embeddings) plus an `ollama-adapter` kustomize deployment (third-party `ollama/ollama` image, PVC for model persistence, init container pulling nomic-embed-text, single replica, ClusterIP 11434). Provides local embeddings and a path to self-hosted local LLMs.
- **sources:** docs020.../008_README.md, docs020.../009_023_session_handoff_vertical_architecture(1).md, docs021.../026_implementation_todo_vertical_architecture(2).md#0.3
- **relations:** rag_index, rag_lookup, self-hosted LLM inference, LoRA fine-tuning (GGUF via ollama create)
- **verify-later:** aiservice/ollama.go; deployments/kustomize/services/ollama-adapter/*; createAIClient "ollama" case

<!-- SOURCE: U22_recent_small_docs.md -->
### Model alias upgrades (Sonnet/Opus 4.5–4.6)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 does idempotent `UPDATE agent_definitions SET default_config = replace(... claude-haiku-4-5 ... claude-sonnet-4-5 ...)`; handoff records chief-strategist→opus-4-6, planners/classifiers→sonnet-4-6, stale claude-3.x refs replaced.
- **what:** SQL migrations that upgrade per-agent model references in `agent_definitions.default_config` — planning/strategy agents to the strongest tier (chief-strategist→opus, site/domain planners+classifiers→sonnet), content generation kept on haiku for cost, and all stale `claude-3.x` aliases modernised. Model aliases resolve to API strings; both original and resolved names logged.
- **sources:** docs020.../003_llm_model_upgrades_and_logging.sql, docs020.../009_023_session_handoff_vertical_architecture(1).md#done
- **relations:** llm_call_log (logs resolved model), model_aliases.go
- **verify-later:** agent_definitions model values for chief-strategist/site-planner/site-classifier; model_aliases.go 4.6 entries

<!-- SOURCE: U22_recent_small_docs.md -->
### Self-hosted LLM inference (vLLM/GPU at scale)
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Self-Hosted LLM Validation — Deploy vLLM or llama.cpp serving a 7B model"; cost tables for a 48-hour million-agent run.
- **what:** A plan to serve 7B models (Mistral/Llama 3/Qwen 2.5) on GPU via vLLM with continuous batching to escape per-token API costs at scale (1,000-2,000 req/min per A100). Bridges to the Ollama/local-model path and the LoRA fine-tuning targets. Estimated hybrid GPU+CPU cost $1,000-3,000 for a 48-hour million-agent burst.
- **sources:** docs021.../015_scaling_analysis.md#phase-2, docs021.../015_scaling_analysis.md#cost-estimates
- **relations:** Ollama provider, LoRA fine-tuning, worker pools
- **verify-later:** any vLLM/GPU inference deployment or stub_llm action

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### RAG best practices (filter-first-then-rank knowledge base)
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** 012_rag_best_practices (dated 2026-03-24, in old/older1); live successor docs/agent_docs/docs020_llm_training_rag/012b_rag_best_practices_v2.md exists.
- **what:** RAG guidance for the site pipeline: always filter `knowledge_base` by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking to avoid cross-vertical contamination; keep RAG at 20-30% of the context window (2-8 examples, quality over quantity); use nomic-embed-text with `search_query:`/`search_document:` task prefixes (recommend nomic-v2-moe upgrade); quality-gate scraped/Claude/human/audit sources; track `embedding_model` and never mix embedding spaces.
- **sources:** old/older1/012_rag_best_practices.md#core-principle, #embedding-model-choice, #avoiding-common-rag-failures
- **relations:** replacement = 012b_rag_best_practices_v2.md; quality flywheel; canine biology knowledge base
- **verify-later:** rag_index/rag_lookup actions filter+prefix; knowledge_base metadata columns

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### GPU/model infrastructure via endpoint health table
- **category:** model-infrastructure
- **status-signal:** superseded
- **status-evidence:** 020c_gpu_and_model_infrastructure_v3 (2026-03-24, old/older1); "Current Infrastructure State … Not Yet Deployed: ai_endpoint_health table, Back-to-triage error handling, Health check in claim_work_item"; live successor 009_model_infrastructure.md + 020d/020e exist.
- **what:** Three-layer model-availability architecture with `ai_endpoint_health` as the sole GPU scheduler: (L1) dispatch loop checks endpoint health before claiming and skips items whose handler's endpoint is down (item stays `triaged`); (L2) back-to-triage `AIUnavailableError` releases items without counting an attempt and marks the endpoint unhealthy on 401/402; (L3) GPU lifecycle is manual K8s Service creation the health-checker auto-discovers. Claude uses a dual-mode hourly 1-token haiku ping (~$0.002/mo) for auto-recovery. Items never know about models — agent definitions do.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#architecture-three-layers, #standing-decisions
- **relations:** replacement = 009_model_infrastructure.md; model routing via agent_definitions; back-to-triage
- **verify-later:** ai_endpoint_health table; claim_work_item_action.go health check; kafka-scheduler ping task

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Model quality assessment & per-agent model assignment
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 020c "Tested but Not Persistent": Llama 3.3 70B on H100 (classification 8/10, content 9/10, design 7/10), Mistral Small 3 CPU (5/6/3); recommended assignment table routes strategist/webdesign/planner→Claude, classifier/content-writer/triage→Llama70B GPU, briefing→Mistral CPU; cost projection ~$910-990 vs $15-30k all-Claude.
- **what:** Benchmarked model quality per task (Claude reference, Llama 70B near-parity on content/classification, Mistral weak on design) and a per-agent endpoint assignment mapping high-leverage structural work to Claude and bulk content/triage to GPU Llama, projecting ~95% cost reduction at 2000 domains. Model routing is controlled via agent_definitions `ai_service` (swap + snapshot).
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#model-quality-assessment, #cost-projection, #models-to-evaluate
- **relations:** endpoint health table; snapshot/swap/rollback (021); RAG/LoRA flywheel
- **verify-later:** agent_definitions ai_service per agent; model aliases claude-sonnet/opus

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel C — LoRA fine-tuning path (Unsloth QLoRA Llama 3.3 70B)
- **category:** model-infrastructure
- **status-signal:** deployed (first run closed out)
- **status-evidence:** NOTES(39) "Update — 2026-06-05: iter_0 CLOSED OUT … adapter_model.safetensors 828MB … training_run 1cd65dd7 reconciled to complete"; FOCUS(21) §2.5 pipeline scripted
- **what:** The training pipeline: pull dataset from Postgres → Unsloth QLoRA train Llama 3.3 70B Instruct (`unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, 3 epochs, batch 1, grad-accum 8, lr 2e-4, lora_r 16, max_seq 4096) → inference sanity test → LoRA adapter (~150MB). Base 70B chosen because hardware was already available, though 8B was flagged as likely 95% quality at 10% cost. Real run was ~24h (not the scripts' claimed 30-90 min).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150; eval/iter0_eval/lora_iter0_full/README.md (frontmatter)
- **relations:** consumes training_exports datasets; deployed via Phase 5 launcher; produces LoRA iter0 adapter; superseded automation design = Flywheel C Phase 2
- **verify-later:** flywheel_C/02_train_llama_3_3_70b.py, 01_pull_dataset_from_postgres.sh, run.sh

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel C Phase 2 — HTTP-job-server training automation
- **category:** model-infrastructure
- **status-signal:** abandoned
- **status-evidence:** FOCUS(21) §2.5.1 "design locked, not built" (2026-04-23), proposing `model-trainer/model-evaluator/training-flywheel-orchestrator` + `POST /jobs` VM server; superseded in practice by the Kafka/saga Phase 5 chain (NOTES(39) §1) where model-trainer is an orchestrator, not an HTTP-polling agent
- **what:** An abandoned design where a `model-trainer` specialist would POST a dataset to a ~200-line FastAPI-style HTTP job server running on the GPU VM (`POST /jobs`, `GET /jobs/{id}`, download adapter), polling to completion, with three new tables (`model_training_runs`, `model_artefacts`, `model_evaluations`). SSH-remote-exec and a VM Kafka consumer were both explicitly rejected. The actual Phase 5 build instead made the VM credential-free with the chassis driving via thunder-adapter presigned URLs.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.5.1; #15 changelog 2026-04-23
- **relations:** superseded by Phase 5 training-launcher + model-trainer saga; the schema names live on as `model_lifecycle.training_runs`
- **verify-later:** model_lifecycle.training_runs; no `/jobs` HTTP server exists in repo

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Model swap / revert mechanism
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4/§4.1 "Swap / revert functions deployed" (migration 083)
- **what:** Per-agent per-step functions `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` that safely snapshot an agent's `ai_service` block in `agent_definitions.default_config` before swapping its LLM (e.g. Claude → a fine-tuned local model), with a full-table backup as the nuclear option. `snapshot_agent`/`revert_agent` are also the sanctioned backup path for agent-definition migrations (used by migration 110).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4; phase5/110_training_launcher_batch_presign(1).sql#0a
- **relations:** the deployment step of the flywheel (swap-if-eval-passes); used by migration snapshotting
- **verify-later:** migration 083; agent_definitions_backup table; snapshot_agent/revert_agent functions

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### AI endpoint health routing (claude / cpu-ollama / gpu-ollama)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.3 table lists gpu-ollama as "currently DOWN, not always-on"; §4.1 "Endpoint health routing deployed"
- **what:** `ai_endpoint_health` table (migration 085) tracks three inference endpoints — Claude (default high-quality), cpu-ollama (embeddings + mistral-small3.1/nomic), gpu-ollama (Llama 70B, future LoRAs). Healthy endpoint → work flows; unhealthy → items wait (back-to-triage). The kafka-scheduler only probes endpoints listed here, which is why the sibling `ollama-eval` pod stays invisible to production routing.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.3, #2.4d
- **relations:** substrate for Flywheel D eval and model swap; ThunderCompute H100 was the intended gpu-ollama
- **verify-later:** ai_endpoint_health table; ollama-adapter, ollama-gpu, ollama-eval services

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Phase 5 training-launcher + model-trainer orchestration chain
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Launcher green … LAUNCH_PID=216 … COMPLETED → notified parent success"
- **what:** The real `training-launcher` (migration 102, replacing a stub) driven by the `model-trainer` orchestrator, which spawns then calls `training-data-preparer → gpu-provisioner → training-launcher` over Kafka/saga. The launcher presigns dataset+scripts, computes checkpoint keys, presigns them, assembles an upload manifest, SSHes it onto the VM, and launches training detached. Two-level await distinction (child's intermediate adapter calls vs the child→parent final notification) is load-bearing.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#1, #5; RUNBOOK_iter0_pretrigger(3).md#6; flywheel_docs/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** replaces Flywheel C Phase 2 HTTP-server design; children call thunder-adapter; superseded predecessor = 2026-05-24 Option A handoff
- **verify-later:** agent_definitions training-launcher (id 1223bdc1), model-trainer (94f5a069); migrations 102/109/110

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### setsid detached launch command
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §4 "ssh_exec blocks to command exit (§2), so the launch must return immediately … the SSH channel hits EOF right after echo"; confirmed `LAUNCH_PID=216` (2026-06-09)
- **what:** Because the adapter's `ssh_exec` runs `session.Run` and blocks up to a 5-min timeout for the remote command to exit, the launch command runs the fetch+train chain under `setsid bash -c '…' </dev/null >launch.log 2>&1 &` and echoes `LAUNCH_PID=$!`, so the SSH channel EOFs immediately. An early superseded version used `nohup`; a real bug found later: `write_manifest` (first VM-FS touch) needed `sudo mkdir`/`sudo chown /workspace` (fixed in 109a) because `/` isn't ssh-user-writable.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#4, #update-2026-06-05-deploy-step-2
- **relations:** part of training-launcher; 109a perm fix; run.sh markers
- **verify-later:** thunder_ssh_exec_dispatch.go; ssh_exec_actions.go sshCommandTimeout

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Batch presign (prepare_object_urls) — O(K²) loop retirement (migration 110)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 110 SQL header + NOTES(39) "Update — 2026-06-09 (3): batch route CONFIRMED end-to-end … Full launcher path completed in ~26s … state Version 30 … The O(K²) class is gone"
- **what:** After the race fix, the K=40 per-checkpoint loop still crawled to a halt by iter_9 (~9 min) because every awaited substep re-persisted the entire expanded ~80-substep workflow + growing collected_data/ProcessingHistory — O(K²). Decision: replace the loop + `flatten_checkpoint_urls` with one batch adapter call `prepare_object_urls` (keys[] → ordered presigned_urls[], reusing `DataURLAction.ObjectURL` per key, no new signing path). Migration 110 swaps `presign_checkpoints` to `dispatch_thunder_prepare_object_urls` and re-points `assemble_manifest.checkpoint_urls → ckpt_presign_batch.presigned_urls`, dropping flatten; workflow completed in ~26s.
- **sources:** phase5/110_training_launcher_batch_presign(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-09, #update-2026-06-09-2, #update-2026-06-09-3
- **relations:** the documented structural fallback that became the chosen path; retires the presign loop
- **verify-later:** data_url_actions.go handlePrepareObjectURLs; thunder_prepare_object_urls_dispatch.go; migration 110

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### run.sh RUN_SH markers + set -e durability hard-gate
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** PLAN(4) "run.sh — BUILT 2026-06-05 … set -euo pipefail plus 02_train's final-upload hard-gate … RUN_SH_DONE are only reached on exit 0, which now implies the adapter is in B2"
- **what:** The on-VM launch chain emits grep-able markers (`RUN_SH_START → RUN_SH_STEP setup → RUN_SH_SMOKE_OK → RUN_SH_STEP full_train → RUN_SH_FULL_OK → RUN_SH_DONE`). Because `set -euo pipefail` plus the final-upload raise means DONE only prints on exit 0, `RUN_SH_DONE` came to mean "trained AND uploaded" — the flip that makes the monitor's DONE_OK→decommission safe. `SAVE_STEPS` (cadence) lives in run.sh, default 50 (~1.5h/checkpoint); lowered to 10 for fast tests.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md#run.sh; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4; phase5/NOTES_phase5_training_launcher_running(39).md#8 (Healthy markers)
- **relations:** parsed by thunder-training-monitor probe; gates checkpoint upload path
- **verify-later:** run.sh (bundle at finetuning/scripts/bundle.tar.gz); 02_train --upload-manifest

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### iter0 pre-trigger + Phase B/C/D deploy runbooks
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_iter0_pretrigger(3) §6 trigger `model-trainer` with export 146a9a12; RUNBOOK bcd(7) "One-line summary of the gates" (2026-06)
- **what:** Two operational runbooks. The iter0 pretrigger runbook lists the gates to reach the first automated training launch (deploy adapter+chassis, upload the scripts bundle, adapter round-trip, cost gate, a gpu-provisioner smoke test of the D4 topic path, then trigger model-trainer). The Phase B/C/D deploy runbook stages the checkpoint-upload rollout: apply 109 → re-pack/re-upload bundle → Tier-2 short launch (B+C integration, SAVE_STEPS low) → resume (blocked on D3+migration) → enable the monitor last. Both hard-code b2 CLI (not aws) and the "verify positive evidence, complete≠succeeded" discipline.
- **sources:** phase5/RUNBOOK_iter0_pretrigger(3).md; phase5/RUNBOOK_phase_b_c_d_deploy(7).md; phase5/NOTES_phase5_training_launcher_running(39).md#7 (Pre-trigger gates)
- **relations:** operationalise the launcher + checkpoint upload + monitor; export a8484922 is the do-not-use trap
- **verify-later:** migrations 109/109a/109b; scheduled_tasks thunder-training-monitor enable step

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
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

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### RAG pipeline deployment (ollama-adapter, rag_lookup/rag_index, knowledge_base)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Live repo confirms `platform/orchestration/actions/rag_actions.go` and other RAG-related action files exist and are registered, matching the deployment bundle's file-placement manifest.
- **what:** A deployment bundle for adding retrieval-augmented generation to the chassis: a new `ollama-adapter` k8s service (own kustomize base+overlay, `ollama/ollama` image, an idempotent init container that pulls an embedding model onto a PVC — `nomic-embed-text` ~300MB, sized for a 8Gi memory limit that also leaves room for future 7B models), two new SQL migrations (`llm_call_log` with a stats view, and `knowledge_base` for RAG storage), two new registered actions (`rag_index` — chunks and embeds content into `knowledge_base`; `rag_lookup` — embeds a query and returns top-k matches), plus a nullable-helpers Go package and patches to `ai_actions.go`/`registry.go`/`anthropic.go` (adds LLM call timing/token logging and an `ollama` case to `createAIClient`).
- **sources:** docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md
- **relations:** model-infrastructure (009 anchor); contextkit's embed.go (same Ollama-embeddings-endpoint pattern reused independently)
- **verify-later:** platform/orchestration/actions/rag_actions.go, deployments/kustomize/services/ollama-adapter/, llm_call_log / knowledge_base tables

<!-- SOURCE: U25_leopardess_social.md -->
### Per-workflow-step model routing (data-sovereignty mechanism)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** AUDIT P1/P2 (2026-07-10): "TRUE — ExecuteLLMPromptAction resolves ai_service in a three-tier lookup, tier 2 being workflow.steps[step].config.ai_service"; "ollama-adapter … ClusterIP only".
- **what:** Model/provider is selectable per workflow step with no new code (three-tier ai_service resolution), with live swap tooling (swap_agent_model, migration 083). A self-hosted step genuinely never leaves the cluster (stock ollama image, ClusterIP, in-cluster calls). This underpins the honest claim: "steps that touch your data can run on infrastructure you control; only steps that don't need to leave call a foundation model."
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6
- **relations:** text-provider wiring reality; data-sovereignty positioning; no-tenant-isolation fact
- **verify-later:** ai_actions.go ExecuteLLMPromptAction; 021_model_swap_and_rollback.sql; ollama-adapter service spec

<!-- SOURCE: U25_leopardess_social.md -->
### Text-provider wiring reality (two providers end-to-end)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** AUDIT P3: "createAIClient switch: anthropic and ollama only; openai is a stubbed error … 'Mistral' is … run through the same self-hosted Ollama pod."
- **what:** Only Anthropic and Ollama work end-to-end for text; nothing else is wired. Imagery is broader (Gemini + Stability, routed by kind not config). The news pipeline has a separate hand-rolled provider path (xAI /v1/responses with web_search+x_search, OpenAI, Perplexity) bypassing the generic AIService entirely. A real constraint on any "model choice" marketing claim (RUNBOOK landmine 12).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P3; docs/leopardessconsulting/RUNBOOK.md#landmine-12
- **relations:** per-step model routing; news-feed-pipeline (separate provider path)
- **verify-later:** createAIClient switch; feed_actions.go provider paths

<!-- SOURCE: U25_leopardess_social.md -->
### llama3.3:70b trained but never used for inference; dynamic GPU provisioning
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** AUDIT P4: "one complete run (2026-06-03→04) … No agent_definitions row points at llama3.3:70b — trained and tested, never used for production inference. TODO logged in 009_model_infrastructure.md."
- **what:** The larger self-hosted model exists as a completed training run; GPU provisioning is genuinely dynamic (thunder_instances, ThunderCompute, decommissioned per run) but experimental. Wiring it to production inference would strengthen the data-sovereignty positioning; deliberately logged in the model-infrastructure home doc rather than duplicated.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6; docs/leopardessconsulting/RUNBOOK.md#Reference
- **relations:** finetuning-flywheel; per-step model routing
- **verify-later:** model_lifecycle.training_runs; agent_definitions model references

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG knowledge_base (shared pgvector store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** deployed
- **status-evidence:** "knowledge_base table is sitting empty in Postgres, waiting for content ... ivfflat index"; migration 082 marked idempotent/deployed in the vertical-architecture handoff.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, ivfflat cosine index, and a trigram fallback index. Any agent reads via `rag_lookup`, any writes via `rag_index`. Later extended with source provenance columns (docs021).
- **sources:** docs020.../004_rag_knowledge_base.sql, docs020.../010_simple_explanation.md, docs020.../008_README.md
- **relations:** rag_lookup, rag_index, Ollama provider, vertical knowledge architecture
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_lookup action (vector search + trigram fallback)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** Registry patch written ("NEEDS PATCH — add 2 rag entries"); action code written but registry.go patch listed as not-yet-applied in the handoff.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend `search_query:` task prefix.
- **sources:** docs020.../010_simple_explanation.md#rag_lookup, docs020.../012b_rag_best_practices_v2.md, docs020.../005_PATCHES.md#patch-03
- **relations:** rag_index, knowledge_base, content-writer RAG injection
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_index action (chunk, embed, dedup, store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** New file `rag_actions.go` "ready to add"; registry patch pending; non-fatal embedding failure behaviour specified in the revised plan.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails the chunk is still stored (searchable via trigram). Intended to accept source_authority/vertical_slug/knowledge_type once schema extended.
- **sources:** docs020.../010_simple_explanation.md#rag_index, docs020.../012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup, knowledge-indexer agent, vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG best practices — filter-first, quality gating, token budget
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** Dated 2026-03-24 best-practices doc; "Implementation Priority" is a to-do list (add metadata columns, update actions), i.e. not yet applied.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failures and their fixes.
- **sources:** docs020.../012b_rag_best_practices_v2.md
- **relations:** rag_index, rag_lookup, knowledge sources (scraped/claude-output/human-curated/audit-insight)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

<!-- SOURCE: U22_recent_small_docs.md -->
### knowledge-indexer agent (deferred)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index, vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

<!-- SOURCE: U25_leopardess_social.md -->
### Concept-document RAG for content writers (v2+)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** 003c "Why not RAG for v1?" + "RAG integration (v2+)"; 003d drops RAG to "Offline reference (not ingested for v1) … agents in v2+ via RAG".
- **what:** Deferred design: when content surface outgrows page-level content_context, ingest the full concept document into the knowledge_base via the existing rag_index action and add a rag_lookup step to the content-writer workflow (query built from page slug + section function, results into the prompt). Deliberately not built for a 5-page v1 — the structured mission/roadmap fields carry enough context.
- **sources:** docs/social001_vonc_tiktok_social/003c_spark_strategic_planning_architecture(2).md#Why-not-RAG, #RAG-integration (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#What-goes-where
- **relations:** mission/roadmap aspects; documentation-system
- **verify-later:** rag_index/rag_lookup actions; knowledge_base table

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG knowledge_base (shared pgvector store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** deployed
- **status-evidence:** "knowledge_base table is sitting empty in Postgres, waiting for content ... ivfflat index"; migration 082 marked idempotent/deployed in the vertical-architecture handoff.
- **what:** A shared (not per-agent) `knowledge_base` table storing chunked content with a `vector(768)` embedding (nomic-embed-text), collection/industry/domain classification, SHA256 dedup, ivfflat cosine index, and a trigram fallback index. Any agent reads via `rag_lookup`, any writes via `rag_index`. Later extended with source provenance columns (docs021).
- **sources:** docs020.../004_rag_knowledge_base.sql, docs020.../010_simple_explanation.md, docs020.../008_README.md
- **relations:** rag_lookup, rag_index, Ollama provider, vertical knowledge architecture
- **verify-later:** knowledge_base table + idx_kb_embedding (ivfflat) + idx_kb_content_trgm; knowledge_base_stats view

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_lookup action (vector search + trigram fallback)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** Registry patch written ("NEEDS PATCH — add 2 rag entries"); action code written but registry.go patch listed as not-yet-applied in the handoff.
- **what:** An action that embeds the query via Ollama, runs pgvector cosine similarity within a collection, and returns both structured `rag_results` and a combined `rag_context` string for prompt injection; falls back to Postgres trigram text search when Ollama is down (reported in `search_method`). Best practice: filter by metadata (vertical/component/quality) before ranking, and prepend `search_query:` task prefix.
- **sources:** docs020.../010_simple_explanation.md#rag_lookup, docs020.../012b_rag_best_practices_v2.md, docs020.../005_PATCHES.md#patch-03
- **relations:** rag_index, knowledge_base, content-writer RAG injection
- **verify-later:** GlobalActionRegistry entry rag_lookup; RAGLookupAction min_authority/filter support

<!-- SOURCE: U22_recent_small_docs.md -->
### rag_index action (chunk, embed, dedup, store)
- **category:** NEW:rag-knowledge-base
- **status-signal:** partial
- **status-evidence:** New file `rag_actions.go` "ready to add"; registry patch pending; non-fatal embedding failure behaviour specified in the revised plan.
- **what:** An action that splits text into chunks (default ~1000 chars, 200 overlap, sentence-boundary), SHA256-hashes each for dedup, embeds via Ollama, and inserts into `knowledge_base` tagged by collection/metadata. If embedding fails the chunk is still stored (searchable via trigram). Intended to accept source_authority/vertical_slug/knowledge_type once schema extended.
- **sources:** docs020.../010_simple_explanation.md#rag_index, docs020.../012b_rag_best_practices_v2.md#implementation-priority
- **relations:** rag_lookup, knowledge-indexer agent, vertical research handler
- **verify-later:** GlobalActionRegistry entry rag_index; RAGIndexAction dedup on collection+content_hash

<!-- SOURCE: U22_recent_small_docs.md -->
### RAG best practices — filter-first, quality gating, token budget
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** Dated 2026-03-24 best-practices doc; "Implementation Priority" is a to-do list (add metadata columns, update actions), i.e. not yet applied.
- **what:** A methodology for the site-build RAG: always filter by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking; keep RAG context to 20-30% of the window and 3-5 examples; gate entries by source_quality (high/verified) for prompt injection; track embedding_model and never mix embedding spaces; prepend nomic task prefixes (search_document/search_query); prefer nomic-embed-text-v2-moe. Names five common RAG failures and their fixes.
- **sources:** docs020.../012b_rag_best_practices_v2.md
- **relations:** rag_index, rag_lookup, knowledge sources (scraped/claude-output/human-curated/audit-insight)
- **verify-later:** knowledge_base metadata columns (vertical/component_type/source_quality); task-prefix handling in rag actions

<!-- SOURCE: U22_recent_small_docs.md -->
### knowledge-indexer agent (deferred)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** "Future agent (owns the knowledge-building domain): knowledge-indexer agent ... For now, we implement the actions. The agent comes when we have a use case."
- **what:** A proposed but deliberately-unbuilt agent that would own the knowledge-building process (load indexing targets → web_scrape → rag_index → refresh), called by the maintenance orchestrator or build pipeline. Held back per the "reuse before creating — don't build an agent until the workflow demands one" principle; the rag_index/rag_lookup actions suffice for now.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#item-4
- **relations:** rag_index, vertical research handler (later realises this role)
- **verify-later:** agent_definitions for any knowledge-indexer/vertical-research-handler

<!-- SOURCE: U25_leopardess_social.md -->
### Concept-document RAG for content writers (v2+)
- **category:** NEW:rag-knowledge-base
- **status-signal:** aspirational
- **status-evidence:** 003c "Why not RAG for v1?" + "RAG integration (v2+)"; 003d drops RAG to "Offline reference (not ingested for v1) … agents in v2+ via RAG".
- **what:** Deferred design: when content surface outgrows page-level content_context, ingest the full concept document into the knowledge_base via the existing rag_index action and add a rag_lookup step to the content-writer workflow (query built from page slug + section function, results into the prompt). Deliberately not built for a 5-page v1 — the structured mission/roadmap fields carry enough context.
- **sources:** docs/social001_vonc_tiktok_social/003c_spark_strategic_planning_architecture(2).md#Why-not-RAG, #RAG-integration (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#What-goes-where
- **relations:** mission/roadmap aspects; documentation-system
- **verify-later:** rag_index/rag_lookup actions; knowledge_base table

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model quality assessment: local 70B comparable for some tasks
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** 009 test table dated 2026-03-24; 023 raw comparative transcripts
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 design; Mistral Small 3 CPU adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria captured in 023: JSON parse w/o fences, exact field names, specific headlines, action-verb CTAs, no invented claims. ThunderCompute quirks: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009#Model Quality Assessment, #ThunderCompute Notes; 023 full
- **relations:** fine-tuning path; LLM tiering
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### LLM reliability strategy for component generation (observability first, shrink the contract second)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open
- **what:** LLMs are structurally good but unreliable at exact schema↔template list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator; accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; Tension #1 (same "don't trust LLM formal labels" family); tiered field classification
- **verify-later:** agent_error_log rejection rows; whether 2a/2b landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Verification harness (build + ops)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.3 "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work"
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate; the ops side (canary, infra rollback, detection) is the thinnest, most-new part.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### LLM model governance: aliases, per-step model choice, llm_call_log
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** v2/027 regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); 040 upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (claude-sonnet-4-5 not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. 067 (filename: "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns in 085); ai_endpoint_health
- **verify-later:** alias resolver; whether extended thinking was ever enabled

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel D — Claude vs local-model quality eval (replay harness)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.4d "paused"; partial results table "Case 1 … 27 min … not a practical substrate for production-scale replay-eval"; §2.4d-comparison methodology "run after eval completes"
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting the dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Model quality assessment: local 70B comparable for some tasks
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** 009 test table dated 2026-03-24; 023 raw comparative transcripts
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 design; Mistral Small 3 CPU adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria captured in 023: JSON parse w/o fences, exact field names, specific headlines, action-verb CTAs, no invented claims. ThunderCompute quirks: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009#Model Quality Assessment, #ThunderCompute Notes; 023 full
- **relations:** fine-tuning path; LLM tiering
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### LLM reliability strategy for component generation (observability first, shrink the contract second)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open
- **what:** LLMs are structurally good but unreliable at exact schema↔template list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator; accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; Tension #1 (same "don't trust LLM formal labels" family); tiered field classification
- **verify-later:** agent_error_log rejection rows; whether 2a/2b landed

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Verification harness (build + ops)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** MASTER(4) §6.3 "Build side is easy … the ops side (canary, infra rollback, incident detection) is the thinnest part of the base and the real building work"
- **what:** Build-check / test-runner / validator / canary / rollback expressed as actions/adapters, checking output against ground truth. The build side reuses existing validate→regenerate; the ops side (canary, infra rollback, detection) is the thinnest, most-new part.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** toolchain validator (self-dev pipeline); lifecycle map
- **verify-later:** go build/vet/test runner; canary/rollback adapters (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### LLM model governance: aliases, per-step model choice, llm_call_log
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** v2/027 regex-replaces all dated model names with aliases ("only the alias resolver in code needs updating"); 040 upgrades planners to sonnet with rationale and creates llm_call_log.
- **what:** Conventions for model management across ~90 agent definitions: model aliases (claude-sonnet-4-5 not dated strings) resolved in code; deliberate per-step model tiering (haiku for cheap classification, sonnet for high-leverage planning, opus for plan_site and tool recreation); llm_call_log capturing calls for cost analysis and training data. 067 (filename: "not_yet_implemented") prepared extended-thinking budget_tokens for classifier/planner gated on a Go patch.
- **sources:** sql_for_agents_v2/027_replace_claude_model_names.sql; 040_optimise_which_llms.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** finetuning flywheel (llm_call_log flywheel columns in 085); ai_endpoint_health
- **verify-later:** alias resolver; whether extended thinking was ever enabled

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel D — Claude vs local-model quality eval (replay harness)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §2.4d "paused"; partial results table "Case 1 … 27 min … not a practical substrate for production-scale replay-eval"; §2.4d-comparison methodology "run after eval completes"
- **what:** A replay-not-rerun eval: pull 20 diverse stored production prompts (`DISTINCT ON (orchestration_id)`) from `llm_call_log`, POST each to a local Ollama model, and compare against the stored Claude response across three levels (structural jq checks, Claude-as-judge, manual review). Target agent was `page-content-writer/iter_0_generate_content`. Stalled on shared-adapter CPU contention, prompting the dedicated `ollama-eval` pod (24Gi/28Gi) and the GPU-substrate argument.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4d, #2.4d-comparison, #14 (Eval and replay methodology)
- **relations:** provides the ROI justification and the eval gate for promoting fine-tuned adapters; blocks enabling model swap
- **verify-later:** ollama-eval deployment; llm_call_log; results.jsonl runner

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Temperature/max_tokens logging gap in llm_call_log
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "the schema column exists but is never written, so llm_call_log.temperature remains NULL by construction" — verified against a 2026-05-26 chassis snapshot
- **what:** Although `llm_call_log` already has `temperature real` and `max_tokens integer` columns, the Go writer (`llm_call_logger.go`) never populates them, and the two call sites in `execute_llm_prompt_action.go` don't pass the values through — even though the actual values sent to the LLM API are already computed a few lines earlier. This makes it impossible to observe from the log alone whether a configured temperature ever reached the API call.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Verification table,#Step 1
- **relations:** Per-field LLM config resolution fallback chain; Possibility-A-vs-B diagnostic method
- **verify-later:** platform/orchestration/actions/llm_call_logger.go LLMCallLogParams struct; execute_llm_prompt_action.go call sites

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-field LLM config resolution fallback chain (temperature parity with max_tokens)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** "Per-field fix not applied" — temperature read is "Still single read: agentConfig['temperature'].(float64)" versus max_tokens' existing two-level fallback
- **what:** Proposes lifting temperature to the same multi-level fallback chain max_tokens already has (step config → agent top → step ai_service → top-level ai_service) via shared `readFloat`/`readNestedFloat` helpers, replacing the single inline float64 type assertion. Would activate 6 currently-dead step-level temperature settings configured in the DB but never actually taking effect.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 3,#Step 4
- **relations:** Temperature/max_tokens logging gap; Possibility-A-vs-B diagnostic method
- **verify-later:** readFloat/readNestedFloat helpers (proposed, not yet added)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Possibility-A-vs-B diagnostic method for silent LLM config failures
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "Smallest change that makes possibility A (logging gap) vs B (temperature never set → API default ~1.0) distinguishable from the log" with an exact before/after SQL audit query
- **what:** A diagnostic technique for a silently-broken config field: ship the cheapest possible observability fix first before attempting any structural resolution fix, then re-run a COUNT(*)/COUNT(temperature)/COUNT(max_tokens) audit query pre- and post-deploy. If temperature becomes non-null post-deploy, the bug was pure logging gap; if still null, the upstream read itself is silently failing — distinguishing the two determines whether every historical LLM call ran at the intended temperature.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 2,#Sequencing summary
- **relations:** Temperature/max_tokens logging gap; Per-field LLM config resolution fallback chain
- **verify-later:** SQL audit query in Step 2

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Default temperature hardening (chassis-level fallback ~0.4)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** Explicitly sequenced last and conditional: "Hold this until the read path is proven, so we don't stack a default on top of an unverified read."
- **what:** Once the observability and per-field resolution fixes are proven, proposes a chassis-level default temperature (~0.4) applied only when none is configured at any level, overridable by an explicit value — reasoning that Anthropic's API default (~1.0) is likely too high for the extraction/classification-style prompts most affected.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 5
- **relations:** Per-field LLM config resolution fallback chain
- **verify-later:** n/a (not yet implemented; gated on Steps 1-3)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Temperature/max_tokens logging gap in llm_call_log
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "the schema column exists but is never written, so llm_call_log.temperature remains NULL by construction" — verified against a 2026-05-26 chassis snapshot
- **what:** Although `llm_call_log` already has `temperature real` and `max_tokens integer` columns, the Go writer (`llm_call_logger.go`) never populates them, and the two call sites in `execute_llm_prompt_action.go` don't pass the values through — even though the actual values sent to the LLM API are already computed a few lines earlier. This makes it impossible to observe from the log alone whether a configured temperature ever reached the API call.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Verification table,#Step 1
- **relations:** Per-field LLM config resolution fallback chain; Possibility-A-vs-B diagnostic method
- **verify-later:** platform/orchestration/actions/llm_call_logger.go LLMCallLogParams struct; execute_llm_prompt_action.go call sites

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-field LLM config resolution fallback chain (temperature parity with max_tokens)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** "Per-field fix not applied" — temperature read is "Still single read: agentConfig['temperature'].(float64)" versus max_tokens' existing two-level fallback
- **what:** Proposes lifting temperature to the same multi-level fallback chain max_tokens already has (step config → agent top → step ai_service → top-level ai_service) via shared `readFloat`/`readNestedFloat` helpers, replacing the single inline float64 type assertion. Would activate 6 currently-dead step-level temperature settings configured in the DB but never actually taking effect.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 3,#Step 4
- **relations:** Temperature/max_tokens logging gap; Possibility-A-vs-B diagnostic method
- **verify-later:** readFloat/readNestedFloat helpers (proposed, not yet added)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Possibility-A-vs-B diagnostic method for silent LLM config failures
- **category:** NEW:llm-call-observability
- **status-signal:** partial
- **status-evidence:** "Smallest change that makes possibility A (logging gap) vs B (temperature never set → API default ~1.0) distinguishable from the log" with an exact before/after SQL audit query
- **what:** A diagnostic technique for a silently-broken config field: ship the cheapest possible observability fix first before attempting any structural resolution fix, then re-run a COUNT(*)/COUNT(temperature)/COUNT(max_tokens) audit query pre- and post-deploy. If temperature becomes non-null post-deploy, the bug was pure logging gap; if still null, the upstream read itself is silently failing — distinguishing the two determines whether every historical LLM call ran at the intended temperature.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 2,#Sequencing summary
- **relations:** Temperature/max_tokens logging gap; Per-field LLM config resolution fallback chain
- **verify-later:** SQL audit query in Step 2

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Default temperature hardening (chassis-level fallback ~0.4)
- **category:** NEW:llm-call-observability
- **status-signal:** aspirational
- **status-evidence:** Explicitly sequenced last and conditional: "Hold this until the read path is proven, so we don't stack a default on top of an unverified read."
- **what:** Once the observability and per-field resolution fixes are proven, proposes a chassis-level default temperature (~0.4) applied only when none is configured at any level, overridable by an explicit value — reasoning that Anthropic's API default (~1.0) is likely too high for the extraction/classification-style prompts most affected.
- **sources:** temperature/PLAN_temperature_observability_and_resolution.md#Step 5
- **relations:** Per-field LLM config resolution fallback chain
- **verify-later:** n/a (not yet implemented; gated on Steps 1-3)
