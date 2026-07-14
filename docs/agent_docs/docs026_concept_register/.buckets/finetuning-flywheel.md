
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
