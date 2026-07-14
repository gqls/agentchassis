# EXTRACTION U06 — docs024_key_docs_latest/finetuning (finetuning flywheel)
Extracted 2026-07-13. Files in scope: 165 (123 text files enumerated + 42 skipped binary/large/generated). Concepts found: 58.

Note on enumeration: the charter's `find … -size -1M` predicate matches only empty files
(GNU find rounds sizes UP to the unit, so `-1M` ≡ size 0). Byte-precise `-size -1048576c`
was used instead; the intended split (text files <1MB processed, everything else skipped)
is preserved. All paths below are relative to `docs/agent_docs/docs024_key_docs_latest/finetuning/`.

## Coverage

| file | treatment |
|---|---|
| BUSINESS_PLAN_finetuning_uk.md | full |
| training_data/003_backfill_jsonl_to_postgres.sh | header-scan |
| training_data/backfill_jsonl_to_postgres.sh | header-scan (duplicate of above) |
| training_data/page_content_writer_iter0.jsonl.gz | skipped-binary (2.0MB dataset) |
| working/docubundle/CONTEXT_PACK_thunder_checkpoint_race(1).md | family-latest |
| working/docubundle/README.md | full |
| working/docubundle/context_packages/thunder-checkpoint-race/production_thunder-checkpoint-race_context.txt | skipped-generated (633KB generated context capture) |
| working/docubundle/package_thunder_checkpoint_race.sh | header-scan |
| working/docubundle_old/package_thunder_checkpoint_race(0).sh | family-delta (header-scan; earlier packager revs) |
| working/docubundle_old/package_thunder_checkpoint_race(1).sh | family-delta |
| working/docubundle_old/package_thunder_checkpoint_race(2).sh | family-delta |
| working/docubundle_old/package_thunder_checkpoint_race(3).sh | family-delta |
| working/eval/001_test_comparison_with_claude.txt | full |
| working/eval/README.md | full |
| working/eval/iter0_eval/001_README.md | header-scan |
| working/eval/iter0_eval/002_analysing_results_claude_comparison | skipped-generated (no extension; session notes) |
| working/eval/iter0_eval/00_vm_setup.sh | header-scan (copy of scripts/00_vm_setup.sh) |
| working/eval/iter0_eval/04_eval_iter0.py | skipped-generated (py, stage-2 material) |
| working/eval/iter0_eval/05_level1.py | skipped-generated (py) |
| working/eval/iter0_eval/06_level2.py | skipped-generated (py) |
| working/eval/iter0_eval/07_build_report.py | skipped-generated (py) |
| working/eval/iter0_eval/held_out_cases_v1.jsonl | skipped-generated (582KB data) |
| working/eval/iter0_eval/iter0_eval_results_v1.jsonl | skipped-generated (236KB data) |
| working/eval/iter0_eval/iter0_evaluation_report.md | full |
| working/eval/iter0_eval/level1_metrics.json | skipped-generated |
| working/eval/iter0_eval/level2_judgments.jsonl | skipped-generated |
| working/eval/iter0_eval/lora_iter0_full/adapter_model.safetensors | skipped-binary (828MB model artifact) |
| working/eval/iter0_eval/lora_iter0_full/chat_template.jinja | skipped-generated |
| working/eval/v1/held_out_cases_v1.jsonl | skipped-generated (582KB data) |
| working/eval/v1/held_out_cases_v1.sql | full |
| working/eval/v1/kubectl_script | skipped-generated (no extension) |
| working/eval/v1/section_type_counter | skipped-generated (no extension) |
| working/flywheel_docs/002_README-flywheele_D_evaluation_pipeline.md | full |
| working/flywheel_docs/019_model_lifecycle_schema.sql | full |
| working/flywheel_docs/FOCUS_adapter_design(2).md | family-delta (headings diffed vs (3); no dropped concepts) |
| working/flywheel_docs/FOCUS_adapter_design(3).md | family-latest |
| working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(22).md | family-delta (same structure as (25)) |
| working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(23).md | family-delta |
| working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(24).md | family-delta |
| working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md | family-latest |
| working/flywheel_docs/FOCUS_finetuning_flywheel_and_service_v1.md | family-delta (yields superseded concierge-first offer structure) |
| working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md | full |
| working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md | full |
| working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted_PATCH_2026-05-06.md | full |
| working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md | full |
| working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md | full |
| working/flywheel_docs/PATCH_rag_actions_nomic_prefixes.md | full |
| working/flywheel_docs/README.md | full |
| working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md | family-delta (diffed vs base + 06_04) |
| working/flywheel_docs/STATUS_thunder_adapter_2026-05-12.md | family-delta |
| working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md | family-latest |
| working/flywheel_docs/flywheel_B_step0_checks.sql | header-scan |
| working/flywheel_docs/flywheel_B_step0_cluster_checks.sh | header-scan |
| working/flywheel_docs/flywheel_B_step1a_ollama_smoke.sh | header-scan |
| working/flywheel_docs/flywheel_B_step1b_pgvector_smoke.sql | header-scan |
| working/flywheel_docs/flywheel_B_step2_real_content.sh | header-scan |
| working/flywheel_docs/flywheel_B_step2_v2_portforward.sh | header-scan |
| working/flywheel_docs/flywheel_D_step0_discovery.sql | header-scan |
| working/flywheel_docs/flywheel_D_target_selection.sql | header-scan |
| working/flywheel_docs/ssh_probe.sh | header-scan (NOTE: contains a hardcoded Thunder API bearer token — flag for stage 2 hygiene) |
| working/flywheel_docs/terminology.md | full |
| working/iter0pretrigger/RUNBOOK_iter0_pretrigger.md | family-delta (earliest of the pretrigger runbook family) |
| working/phase1/001_rag_testing.sh | header-scan (pasted session transcript, not a real script) |
| working/phase1/00_vm_setup.sh | header-scan |
| working/phase1/01_pull_dataset_from_postgres.sh | header-scan |
| working/phase1/02_train_llama_3_3_70b.py | skipped-generated (py, stage-2 material) |
| working/phase1/02_train_llama_3_3_70b.py.orig | skipped-generated |
| working/phase1/03_inference_test.py | skipped-generated (py) |
| working/phase1/files/003_backfill_jsonl_to_postgres.sh | header-scan (copy) |
| working/phase1/files/export_fence_test.sql | header-scan |
| working/phase1/files/export_test.jsonl | skipped-generated (data) |
| working/phase1/files/export_test.sql | header-scan |
| working/phase1/files/flywheel_A_export_page_content_writer_iter0.sql | header-scan |
| working/phase1/files/training_fence_test.jsonl | skipped-generated (data) |
| working/phase1/files/training_iter0.jsonl | skipped-generated (21MB dataset) |
| working/phase1/files/training_test.jsonl | skipped-generated (data) |
| working/phase2/README.md | full |
| working/phase2/gpu_output_dump_01 | skipped-generated (no extension; log dump) |
| working/phase2/lora_iter0_full/README.md | header-scan (auto-generated HF model card) |
| working/phase2/lora_iter0_full/adapter_config.json | skipped-generated |
| working/phase2/lora_iter0_full/adapter_model.safetensors | skipped-binary (828MB model artifact) |
| working/phase2/lora_iter0_full/chat_template.jinja | skipped-generated |
| working/phase2/lora_iter0_full/manifest.json | skipped-generated |
| working/phase2/lora_iter0_full/special_tokens_map.json | skipped-generated |
| working/phase2/lora_iter0_full/tokenizer.json | skipped-binary (17MB) |
| working/phase2/lora_iter0_full/tokenizer_config.json | skipped-generated |
| working/phase5/102_training_launcher_real.sql | header-scan (agent-def migration; content covered via NOTES/HANDOFF) |
| working/phase5/103_call_data_preparer_optional_inputs.sql | header-scan |
| working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql | header-scan |
| working/phase5/105_launcher_workspace_sudo_mkdir.sql | header-scan |
| working/phase5/107_thunder_training_monitor_worker.sql | header-scan |
| working/phase5/109_launcher_upload_manifest_wiring.sql | family-delta |
| working/phase5/109_launcher_upload_manifest_wiring(1).sql | family-delta |
| working/phase5/109_launcher_upload_manifest_wiring(2).sql | family-delta |
| working/phase5/109_launcher_upload_manifest_wiring(3).sql | family-latest (header-scan) |
| working/phase5/109a_fix_write_manifest_workspace_perm.sql | header-scan |
| working/phase5/109b_fix_presign_one_loop_item_keypath.sql | header-scan |
| working/phase5/110_training_launcher_batch_presign.sql | family-delta |
| working/phase5/110_training_launcher_batch_presign(2).sql | family-latest (header-scan) |
| working/phase5/111_training_launcher_resume_wiring.sql | header-scan |
| working/phase5/CHASSIS_await_loop_extract.txt | header-scan (generated code extract for the race diagnosis) |
| working/phase5/CONTEXT_PACK_thunder_checkpoint_race.md | family-delta (docubundle (1) is later) |
| working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md | full |
| working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md | family-delta |
| working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(1).md | family-delta |
| working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md | family-latest |
| working/phase5/NEXT_CHAT_MANIFEST.md | header-scan (file inventory for a fresh session) |
| working/phase5/NOTES_phase5_training_launcher_running(36)a.md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(36)c.md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(40).md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(41).md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(42).md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(43).md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(44).md | family-delta |
| working/phase5/NOTES_phase5_training_launcher_running(45).md | family-latest (full, 1241 lines) |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2.md | family-delta |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2(2).md | family-delta |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2(3).md | family-delta |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2(4).md | family-delta |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2(5).md | family-delta |
| working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md | family-latest |
| working/phase5/README_files_used_in_this_investigation.md | header-scan (near-empty) |
| working/phase5/README_what_is_manifest.md | full |
| working/phase5/RUNBOOK_iter0_pretrigger(1).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(2).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(4).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(5).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(6).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(7).md | family-delta |
| working/phase5/RUNBOOK_iter0_pretrigger(8).md | family-latest (headings + content cross-verified in NOTES(45)) |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(5)).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(8).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(9).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(10).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(11).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(12).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(13).md | family-delta |
| working/phase5/RUNBOOK_phase_b_c_d_deploy(14).md | family-latest (headings + content cross-verified in NOTES/PLAN) |
| working/phase5/STATUS_thunder_adapter_2026-05-12_1_(1).md | family-delta (diffed: adds the 06-04 monitor update, same text as STATUS 06_04) |
| working/phase5/UPLOAD_bundle.sh | header-scan |
| working/phase5/bundle.tar.gz | skipped-binary |
| working/phase5/logs-training-launcher.json | skipped-generated (1.27MB log) |
| working/phase5/registry_patch.txt | header-scan |
| working/scripts/001_rag_testing.sh | header-scan (276KB pasted transcript; duplicate of phase1 copy) |
| working/scripts/00_vm_setup(old1).sh | family-delta |
| working/scripts/00_vm_setup.sh | header-scan (family-latest of setup script) |
| working/scripts/01_pull_dataset_from_postgres.sh | header-scan |
| working/scripts/02_train_llama_3_3_70b.py | skipped-generated (py, stage-2 material) |
| working/scripts/02_train_llama_3_3_70b.py.orig1 | skipped-generated |
| working/scripts/02_train_llama_3_3_70b.py.orig2 | skipped-generated |
| working/scripts/03_inference_test.py | skipped-generated (py) |
| working/scripts/04_eval_iter0.py | skipped-generated (py) |
| working/scripts/README_setup.md | full |
| working/scripts/bundle.tar.gz | skipped-binary |
| working/scripts/logs-thunder-adapter.json | skipped-generated (1.04MB log) |
| working/scripts/logs-training-launcher.log | skipped-generated (31MB log) |
| working/scripts/run.sh | header-scan |
| working/scripts/run.sh.orig1.smoke | skipped-generated (superseded script variant) |

## Concepts

### Finetuning flywheel programme (lanes A/B/C/D)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 state block: "Flywheel A ✓ operational / B ✓ done / C phase 1 ✓ / C phase 2 → next priority / D ✓ iter_0 evaluated"; later phase-5 notes carry C-phase-2 to a mostly-green automated launch (2026-06-09) with the monitor still disabled.
- **what:** The internal self-improvement programme: the site-building pipeline emits training data as a byproduct (A: export), RAG injects verified knowledge immediately (B), local models are periodically fine-tuned on the captured data (C), and evaluated against Claude before any swap (D). The strategic goal is dropping API cost by swapping local models in for Claude calls where quality holds.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#1-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#current-state; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** every other concept in this unit; three improvement channels; model swap/revert
- **verify-later:** training_exports schema, model_lifecycle schema, llm_call_log columns, thunder-adapter deployment, scheduled_tasks rows for reaper/monitor

### llm_call_log training-data capture
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.1 "Every LLM call in the system writes to llm_call_log (migration 081, flywheel columns added in 085)"; §4.1 checkbox "[x] llm_call_log populating".
- **what:** Every LLM call logs agent_type/step_name (the "what to train" join key), model/provider, prompt_template/prompt_rendered/response_text (the raw training pair), token/latency/cost signals, success/error, work_item_id, prompt_variant and vertical. Write path is `LogLLMCall` in ai_actions.go — fire-and-forget goroutine, 5s timeout, never blocks the workflow. Retention 90d success / 180d error via `cleanup_old_llm_logs`, which exists but is NOT scheduled (open task). Historically-empty `agent_type` rows exist; recent writes are 100% populated.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.1,#2.4c,#4.2
- **relations:** training-data export agent; replay eval (pulls stored prompts); per-vertical slicing; prompt evolution (prompt_variant)
- **verify-later:** migrations 081/085; platform actions ai_actions.go LogLLMCall; cleanup_old_llm_logs scheduling; llm_call_log row counts

### Training-data export as chassis agent + action (flywheel A, v1→v3.2)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23: "training_data_export.go v3.2 … deployed as v3.2, verified working"; two 1,958-row exports landed in Postgres.
- **what:** Training export is a first-class chassis pipeline component, not ad-hoc SQL: a `training_data_export` action plus `training-data-exporter` worker and `training-data-export-orchestrator` wrapper. It queries llm_call_log, strips markdown fences via the shared `stripMarkdownFromResponse`, validates JSON per row, and writes batched inserts into `training_exports`. The v1→v3.2 evolution encodes several platform lessons (template-config not rendered for deterministic actions; file output on ephemeral pods; pgbouncer transaction limits; RowsAffected checks).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md
- **relations:** training_exports schema; orchestrator wrapper spawning pattern; pgbouncer per-batch commits
- **verify-later:** platform/orchestration/actions/training_data_export.go; agent_definitions rows training-data-exporter / training-data-export-orchestrator; training_exports.runs contents

### training_exports Postgres schema (named snapshot datasets)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-04-23 "Schema (applied, verified)"; export_ids 146a9a12 and fef7be6b each hold 1,958 rows.
- **what:** `training_exports.runs` (one row per export with filters, counts, completed_at) + `training_exports.rows` (ChatML messages + metadata JSONB per training record, unique on (export_id, source_log_id)). Datasets live in Postgres, not files or S3 — named, versioned snapshots referenced by export_id UUID, streamed out to JSONL at training time. Real-time streaming into the table was explicitly rejected to keep snapshot boundaries and avoid coupling observability to training. Known trap: `runs.rows_exported` can disagree with actual `rows` content (export a8484922 recorded 1957 but holds 0 rows) — always verify with a count before launching.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#done-verified; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06-2
- **relations:** training-data export agent; training-data-preparer (streams export to S3); dataset pull scripts
- **verify-later:** training_exports schema in clients_db; the a8484922 zero-row anomaly; recent_runs view

### ChatML export format with metadata sidecar
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22."
- **what:** Training rows are `{messages:[{role:user},{role:assistant}], metadata:{source_log_id, agent_type, step_name, orchestration_id, model, created_at, export_version}}`. Chosen for chat-tuned base-model parity, trainer-tool defaults (Unsloth/Axolotl), and `/api/chat` training-inference parity. Metadata gives row-level traceability back to llm_call_log; export_version future-proofs format evolution. Whole prompt_rendered goes in the user turn (no system/user split yet).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e; working/phase1/files/flywheel_A_export_page_content_writer_iter0.sql (header)
- **relations:** response cleaning; training_exports schema
- **verify-later:** export SQL/action output shape vs current llm_call_log columns

### Response cleaning and SFT negative-example exclusion
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4e iter_0 dataset audit table (97.4% clean JSON, 2.6% fenced, fences stripped on export); "For our first training run: plain SFT, edge cases excluded".
- **what:** Exports must strip markdown code fences (else the fine-tune learns to emit fences, exactly what prompts forbid) and exclude edge-case prose responses: plain SFT has no "don't do this" signal, so intelligent edge-case answers are positive examples of the wrong shape. Those rows stay in llm_call_log as future "rejected" halves of DPO preference pairs — DPO/RLHF is the named future home for negative examples.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4e
- **relations:** ChatML export format; `<no value>` contamination
- **verify-later:** stripMarkdownFromResponse; whether any DPO work exists later

### `<no value>` training-data contamination and the iter_1 filter floor
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-07 known issues: "Training rows from before that fix include the literal token `<no value>`. iter_1's export should filter created_at >= <fix_date>… Action: note the fix-deploy date"; still listed as needed in HANDOFF 2026-05-08.
- **what:** A prompt-builder rendering bug injected the literal token `<no value>` into production prompts; iter_0's training data (and its eval cases) inherit it. The fix-deploy date becomes the created_at filter floor for the iter_1 export. As of the last docs, the date had not been recorded — an open data-hygiene debt.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#known-issues; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#revised-iter_1-priorities
- **relations:** training_exports; held-out eval set (same artefact present)
- **verify-later:** git log for the prompt-builder fix; whether any iter_1 export exists with the filter

### Dataset profile and schema heterogeneity of page-content-writer iter_0
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4g dataset profile tables (n=1,957; p50 prompt 8,250 chars; three dominant JSON schemas: hero-with-CTAs 68%, minimal hero 18%, header/nav 9%).
- **what:** One (agent_type, step_name) training slice actually spans three component output schemas; the model must learn schema selection conditioned on the "Component: X" text in the prompt. A first-pass option to filter to the top-2 schemas (86% of rows) was noted but the full set was trained. Prompt/response size distribution anchors max_seq choice (some prompts approach 4,000 tokens → max_seq 4096).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4g; working/flywheel_docs/terminology.md#seq-4096
- **relations:** Llama 70B QLoRA config; inference-test success criteria (keys match trained schemas)
- **verify-later:** training_exports.rows key distribution query

### Flywheel C manual training pipeline (00–03 scripts, smoke-gates-full)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07: "executed flywheel C phase 1 end-to-end on a Thunder Compute A100 80GB… iter_0 adapter (791MB) … 5/5 valid JSON".
- **what:** Four scripts define the training path: `00_vm_setup.sh` (idempotent pinned env), `01_pull_dataset_from_postgres.sh`, `02_train_llama_3_3_70b.py` (Unsloth QLoRA, CLI-configurable, emits manifest), `03_inference_test.py` (JSON validity + schema keys sanity). Discipline: a 20-row/1-epoch smoke train and smoke inference always gate the full run — cheap insurance on unattended runs. The same bundle later becomes the automated launcher's payload.
- **sources:** working/flywheel_docs/README.md; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#flywheel-c; working/scripts/run.sh (header)
- **relations:** iter_0 baseline run; scripts bundle as deployment unit; run.sh markers
- **verify-later:** working/scripts/* against the bundle actually in B2 (finetuning/scripts/bundle.tar.gz)

### Base-model decision: Llama 3.3 70B Instruct QLoRA (with 8B ablation planned)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.5 "Base model: Llama 3.3 70B Instruct. Decision taken 2026-04-23"; HANDOFF 2026-04-23 decisions-locked list.
- **what:** `unsloth/Llama-3.3-70B-Instruct-bnb-4bit` via Unsloth QLoRA on a single A100/H100 80GB; defaults 3 epochs, batch 1, grad_accum 8, lr 2e-4, lora_r 16, max_seq 4096. 70B was chosen because hardware was available and a strong baseline is useful — with an explicit acknowledgment that 8B likely delivers ~95% of quality at ~10% of inference cost for this narrow structured-JSON task; a same-dataset 8B comparison run is planned but never executed in these docs.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#base-model-decision
- **relations:** iter_0 baseline; cost anchors; epochs ablation
- **verify-later:** 02_train defaults; any 8B run in model_lifecycle.training_runs

### iter_0 baseline training run (real cost/time/loss anchors)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 run summary: 33,136s ≈ 9h12m, final_loss 0.2669 (trailing), peak VRAM 44.8GB, ~$20 total; "Anchor future estimates against $20/iter".
- **what:** The first real fine-tune: 1,958 rows → 1,934 effective, clean loss curve (ep1 1.49→0.27, ep3 →0.10), adapter 791MB fp32 safetensors. Epoch-3 loss gap suggests memorisation → a 2-epoch ablation is queued for iter_1. Cost anchor $20/iteration (training) + $1.50 (eval) ≈ $22/cycle. Later automated runs corrected the wall-time estimate: the full run is ~24h at ~119s/step without FA2, not the "30–90 min" the README claimed.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md; working/phase2/README.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** version pinning; snapshot economics; per-instance uptime bump; fp16 save decision
- **verify-later:** lora_iter0_full/manifest.json; model_lifecycle.training_runs rows 1cd65dd7/e6ab9fad

### GPU environment version pinning (cu124 stack)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Version pin discoveries (essential for any future run)" table with per-pin rationale.
- **what:** The working training environment is a narrow pin set: torch 2.6.0+cu124, transformers<5, torchao<0.17 (transformers imports torchao eagerly; incompatible torchao breaks import entirely), prebuilt flash-attn wheel (Thunder's Ollama template ships CUDA runtime, no nvcc), unsloth+unsloth_zoo both explicitly (git install misses the zoo), hf_transfer as separate package. cu124 is flagged a dead end — next rebuild should move to cu126/cu128. The Unsloth template (used for eval) differs: nvcc present, torch 2.10/cu128, xformers pre-installed, FA2 absent.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#version-pin-discoveries; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons; working/eval/iter0_eval/001_README.md
- **relations:** 00_vm_setup.sh as canonical environment; snapshot economics
- **verify-later:** working/scripts/00_vm_setup.sh pin lines

### Snapshot economics: setup script beats VM snapshots
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-07 "Created `unsloth-trainer-base-01` then deleted it… Break-even: ~18 training runs/month. Reality: 1-4."
- **what:** Thunder snapshots bill the full provisioned 100GB ($15/month) regardless of used bytes, saving only ~25min/$0.85 per cold start — uneconomic below ~18 runs/month. Decision: no snapshots; the version-pinned idempotent `00_vm_setup.sh` is the canonical, version-controlled "snapshot". Phase-2 automation therefore provisions fresh instances every run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#snapshot-decision
- **relations:** version pinning; phase-2 architecture
- **verify-later:** none (economic decision); revisit if run tempo >15/month

### GPU training performance model (smoke ≠ steady state; FA2; seq-length cost)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** terminology.md whole file; NOTES(45) 2026-06-04: "The smoke rate (116 s/step) predicted this — nobody extrapolated it".
- **what:** A small captured mental model of training performance: smoke-test speed is unrepresentative (one-time kernel autotune/CUDA-graph costs amortized over too few steps); steady-state emerges after 5–20 steps; FA2 vs xformers/SDPA is a 2–4× attention-speed lever; attention scales O(N²) so max_seq 4096 quadruples 2048's attention work. Operationally: extrapolate full-run wall time from smoke s/step (the 18h-cap overrun happened because nobody did).
- **sources:** working/flywheel_docs/terminology.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** iter_0 baseline; per-instance uptime bump; cap-sizing-from-smoke queued idea
- **verify-later:** whether cap-sizing-from-smoke was ever built

### fp16 adapter save decision
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** HANDOFF 2026-05-08 decisions: "Save adapters as fp16, not fp32, in iter_1. One-line script change." (iter_0 shipped fp32 at 791–828MB.)
- **what:** PEFT `save_pretrained()` defaults LoRA weights to fp32 even when training in bf16, doubling adapter size and transfer time (17min tnr scp for 791MB). The one-line fix (cast trainable params to fp16 pre-save) was agreed for iter_1 but iter_1 never ran in these docs.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-07_flywheel_C_phase1_complete.md#lessons(10); working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** adapter transport via S3; model_lifecycle.artefacts format field
- **verify-later:** current 02_train script save path

### Flywheel D replay-eval methodology
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Replay, don't re-run"; §2.4d full replay design and partial results.
- **what:** Evaluation replays stored production prompts from llm_call_log against the candidate model instead of re-invoking agents — no orchestration-state pollution, much faster, and directly comparable to the stored Claude output. Test sets use `DISTINCT ON (orchestration_id)` for diversity ("Diverse 20 > random 20"), exported as NDJSON. Fail fast on empty responses; monitor with a watch loop.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/flywheel_D_target_selection.sql (header)
- **relations:** three-level eval pipeline; held-out set; CPU-Ollama eval attempt
- **verify-later:** held_out_cases_v1.sql query against llm_call_log

### Three-level evaluation pipeline (L1 structural / L2 judge / L3 spot-check)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 002_README-flywheele_D_evaluation_pipeline.md (run instructions, ~$1 / ~5min total); iter0_evaluation_report.md generated 2026-05-08.
- **what:** Reusable eval stack: L1 structural metrics computed locally and side-by-side for both models (JSON validity, schema-key match, length ratios, forbidden phrases from the brief's avoid-list, fabrication-marker regexes); L2 Claude-as-judge scoring relevance/voice/integrity 1–5 plus winner, with anonymised randomised A/B and resume support; L3 auto-selected spot-check cases folded into a markdown report by build_report.py. The report deliberately reports confounds and makes no ship/no-ship call. Known limit: L1 fabrication regexes have poor recall — contextual fabrications need L2/L3.
- **sources:** working/flywheel_docs/002_README-flywheele_D_evaluation_pipeline.md; working/eval/iter0_eval/iter0_evaluation_report.md#methodology; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(6)
- **relations:** model_lifecycle.evaluations (l1_metrics/l2_metrics JSONB contract); judge-model choice; eval gate
- **verify-later:** working/eval/iter0_eval/05-07 scripts; any evaluations rows in model_lifecycle

### Held-out eval set v1 as the canonical cross-iteration comparison set
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decision: "held_out_cases_v1.jsonl is the canonical eval set across iterations — same 20 cases… so trends are meaningful."
- **what:** 50 cases pulled from llm_call_log post-training-export-cutoff (created_at > 2026-04-23 14:54:32Z), one per orchestration, defensively excluded from the training set by source_log_id; 20 used for iter_0, 30 reserved. The SQL is kept for reproducibility. Iterations evaluate against the same cases so deltas are trend, not noise. Fresh `_v2` sets are the mechanism for novelty checks.
- **sources:** working/eval/v1/held_out_cases_v1.sql; working/eval/iter0_eval/iter0_evaluation_report.md#sample-selection; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** replay eval; three-level pipeline
- **verify-later:** held_out_cases_v1.jsonl vs training export overlap

### Claude-as-judge with anonymised A/B and self-recognition bias handling
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** iter0 report L2: "5 cases had identical R/V/I scores… 5/5 went to Claude… consistent with residual self-recognition bias"; HANDOFF 2026-05-08 decision: "claude-opus-4-7 is the canonical judge model".
- **what:** Judge design: anonymise responses, randomise A/B positions, score dimensions before picking a winner, and use a *different* Claude model (Opus) than the training-label producer (Sonnet 4.6) to reduce — not eliminate — self-recognition. The bias was empirically observed: rubric-tied cases broke for Claude every time, so headline win-rates get an adjusted reading (16-4 → 12-4 with 4 judge-preference ties). Position bias is checked explicitly (A won 55%, no clear bias).
- **sources:** working/eval/iter0_eval/iter0_evaluation_report.md#level-2; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(5)
- **relations:** three-level pipeline; model_lifecycle.evaluations judge_model column (judge drift tracking index)
- **verify-later:** level2.py anonymisation logic

### iter_0 verdict: shippable for low-stakes; voice fidelity is the iter_1 lever
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** HANDOFF 2026-05-08 decisions: "iter_0 is shippable for low-stakes use… Not for client-facing where Δ−0.20 on voice would be visible"; "Add improve voice fidelity".
- **what:** The evaluated position on the first adapter: iter_0 matches Claude on JSON validity (20/20 vs 19/20) and schema, comparable length, tiny dimension gaps (relevance −0.25, voice −0.20, integrity −0.10), 4 substantive wins. Verdict: usable for internal tooling and low-stakes sites; voice is the largest gap and the main iter_1 lever (more epochs? lora_r 32? stricter voice-compliant training rows). "Address verbosity" was explicitly dropped (data showed no gap). Fabrication is a both-models problem to solve with prompt-time guardrails or post-hoc verification, not adapter training.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md; working/eval/iter0_eval/iter0_evaluation_report.md#tldr; working/eval/001_test_comparison_with_claude.txt
- **relations:** eval gate; deployment_decision vocabulary; fine-tuning candidates
- **verify-later:** whether any deployment_decision row exists; whether iter_0 was ever served in production

### CPU-Ollama replay eval attempt and the dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.4d "mistral-small3.1 on a shared cpu-ollama adapter is not a practical substrate… 20 cases × 25-30 min = 10+ hours"; superseded by Thunder GPU eval (HANDOFF 2026-05-08 ran 20 inferences at ~22s/case for ~$0.50).
- **what:** The first flywheel-D attempt replayed prompts against mistral-small3.1 on the shared CPU Ollama adapter; production contention drove one case to 27 minutes (~4 s/token), so a dedicated `ollama-eval` pod (own PVC/service, invisible to production routing because kafka-scheduler only probes ai_endpoint_health entries) was spun up, with the pod-memory sizing rule learned (limit ≥ model file + 8–12GiB headroom). The whole CPU-eval path was then superseded by evaluating the trained adapter on Thunder GPU instances. Yields the durable prediction framework (swap-with-prompt-tweaks vs swap-after-finetuning vs different substrate).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4d,#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md
- **relations:** ai_endpoint_health; Ollama CPU adapter ops; replay eval
- **verify-later:** whether ollama-eval deployment still exists in kustomize/cluster

### Fine-tuning candidate prioritisation
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §2.6 priority table; flywheel_D_target_selection.sql header ("high volume, high success, short output, structured JSON, low reasoning complexity").
- **what:** Ranked list of agents worth fine-tuning locally: knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher — all high-volume structured-JSON emitters. Explicit non-candidates: page-content-writer long-form (though its iter_0 hero step WAS the first target), visual-design-auditor (judgement), chief-strategist (worth Claude cost). Selection criteria are encoded as a reusable SQL discovery query over llm_call_log volume/recency.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.6; working/flywheel_docs/flywheel_D_target_selection.sql
- **relations:** per-vertical training (vertical column in llm_call_log); model swap
- **verify-later:** current llm_call_log volumes per agent

### Three improvement channels compound (RAG / LoRA / prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** FOCUS(25) §3, sourced from 009_model_infrastructure decision 10; RAG deployed, LoRA iter_0 trained, prompt_variant column exists but no A/B usage evidenced.
- **what:** The framing that RAG (immediate, no training), LoRA fine-tunes (medium-term cost reduction), and prompt evolution via the `prompt_variant` A/B column are three independent levers that compound: good prompts + good RAG produce the best training data, which produces the best fine-tuned model, which needs good prompts and RAG to perform.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#3
- **relations:** knowledge_base RAG; llm_call_log capture; llm-quality-testing category
- **verify-later:** any prompt_variant A/B analysis in code or docs

### model_lifecycle schema (training_runs / artefacts / evaluations / deployable_adapters)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 019_model_lifecycle_schema.sql (full DDL with comments); NOTES(45) 2026-06-09(6) confirms `model_lifecycle.training_runs` live with CHECK status pending/running/complete/failed and live rows.
- **what:** The run-lifecycle namespace: `training_runs` (one row per QLoRA run, FK to training_exports.runs, JSONB hyperparameters for reproducibility, loss/VRAM/cost outcome metrics, thunder_instance_id breadcrumb), `artefacts` (adapter binaries decoupled from runs to allow requantisation, storage_uri + sha256 + format), `evaluations` (per artefact × eval_set × judge, JSONB l1/l2 metrics, free-text human deployment_decision), plus `deployable_adapters` view (latest shipped_% adapter per base model — the chassis's read point for "which adapter to load") and `latest_training_run_per_export`. Supersedes the earlier flat `model_training_runs` sketch in FOCUS §2.5.1/HANDOFF 05-08.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#tables-needed; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7
- **relations:** eval gate; thunder_instances (FK training_run_id); mark_training_run_running/terminal actions
- **verify-later:** schema in clients_db; whether deployable_adapters is read by any chassis code

### Eval gate before promotion (human deployment decision; integrity lives here)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 019 schema docstring "deployment_decision is set by human after review; nullable until then"; PLAN b2 "No upload scheme substitutes for evaluating the adapter"; HANDOFF 05-08 "Auto-deployment NOT included in v1".
- **what:** Adapters never auto-promote: a human reviews flywheel-D output and writes a free-text deployment_decision ('shipped_internal', 'rejected_voice_gap', …); anything `shipped_%` becomes deployable. Critically, the eval gate is also the *integrity* boundary for the hostile-VM upload design — a maliciously-crafted-but-valid adapter written through a legitimate presigned URL is caught by evaluation, not by credentials. The original phase-2 sketch had a conditional auto-swap (`swap_agent_model if score ≥ threshold`) which was walked back to human review.
- **sources:** working/flywheel_docs/019_model_lifecycle_schema.sql#evaluations; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#agents-needed
- **relations:** hostile-VM threat model; model swap/revert; deployable_adapters view
- **verify-later:** whether any evaluation row has a decision; whether swap_agent_model was ever wired to eval output

### Flywheel C phase-2 automation architecture (HTTP job server → SSH-exec → adapter dispatch)
- **category:** finetuning-flywheel
- **status-signal:** superseded
- **status-evidence:** FOCUS(25) §2.5.1 "HTTP job server (Option B chosen)"; HANDOFF 2026-05-08 "Architecture decision: SSH-exec, not HTTP job server (initially)"; the built system (phase 5) uses thunder-adapter dispatch actions + presigned URLs + detached run.sh.
- **what:** The automation design went through three generations: (1) VM-side HTTP job server (POST /jobs, bearer auth, systemd, TLS) — designed, never built; (2) direct SSH-exec from chassis (simpler at low run frequency, no VM service to maintain) — the pivot decision; (3) the final built shape: chassis dispatch actions publish to thunder-adapter (provision/ssh_exec/presign), data moves only via presigned B2 URLs, training runs detached under run.sh with a separate monitor. Rejected throughout: Kafka consumer on the VM (connectivity + overkill). "Chassis drives, GPU VM serves" is the invariant across all three; each generation supersedes the previous.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.5.1; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#architecture-decision; working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** model-trainer chain; training-launcher; presigned data plane
- **verify-later:** no HTTP job server should exist anywhere; adapter dispatch actions in registry.go

### model-trainer orchestration chain (spawn/call data-preparer → provisioner → launcher)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "model-trainer flow confirmed live: spawn_data_preparer → spawn_provisioner → spawn_launcher → call_data_preparer → call_provisioner → call_launcher → complete".
- **what:** The end-to-end automated training run is the `model-trainer` orchestrator (id 94f5a069): spawns three worker agents up front, then calls them in order. `training-data-preparer` (71ab9361) streams the export to S3 as JSONL and INSERTs the pending training_runs row; `gpu-provisioner` provisions the A100 through thunder-adapter; `training-launcher` (1223bdc1) presigns, writes the manifest, and SSH-launches. Known open bug: a failed call_agent step falls through to the next call instead of aborting the orchestration (produces confusing secondary errors).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#1,#update-2026-06-06; working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md#where-this-sits,#also-pending
- **relations:** training-launcher workflow; migrations 103/104; call_agent fall-through bug
- **verify-later:** agent_definitions 94f5a069/71ab9361/1223bdc1 in clients_db; the fall-through behaviour in coordinator code

### training-launcher real workflow (presign → manifest → detached SSH launch → mark_running)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… presign_dataset → presign_scripts → compute_keys → presign_checkpoints (ONE batch await) → presign_final → [check_resume] → assemble_manifest → write_manifest → ssh_exec_launch → mark_running → complete."
- **what:** The launcher replaced a stub (migration 102) with a workflow of dispatch actions cloned from the proven decommission pattern: presign dataset + scripts bundle (GET), compute K checkpoint keys, batch-presign checkpoint PUTs + final PUT, optionally resolve a resume checkpoint, assemble and SSH-place `/workspace/upload_manifest.json`, then launch training detached and flip training_runs pending→running (`mark_training_run_running`, hardcoded guarded transition). Constants live in step config; cross-step values resolve via config dot-paths; the ssh command is built from a `command_template` with `{token}` interpolation. Evolved through migrations 102→105→109/109a/109b→110→111.
- **sources:** working/phase5/HANDOFF_2026-05-24_phase5_launcher_build.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#5,#update-2026-06-09-3; working/phase5/102_training_launcher_real.sql (header)
- **relations:** batch presign; upload manifest; setsid launch; migrations family
- **verify-later:** live training-launcher default_config (2d state check); registry entries for the dispatch actions

### setsid detached launch and the detached exit-0 false-success gap
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §4 (the single-line setsid command) and 2026-06-03 ~18:04: "exit_code 0 only because the command's last token is echo (the known detached-ssh_exec false-success)".
- **what:** The adapter's ssh_exec blocks until the remote command exits (5-min timeout), so the launch chain (curl bundle + dataset via presigned URLs, untar, nohup run.sh) runs under `setsid … & echo LAUNCH_PID=$!` — the SSH session returns immediately with the PID. The cost: exit_code 0 only proves the echo ran; VM-side failures inside the detached chain (e.g. the /workspace permission failure) are invisible to the launcher. Corollary lessons: the command_template must stand up its own workspace with sudo mkdir+chown (105/109a), and any best-effort VM step under `set -e` becomes fatal (the root-owned ~/.bashrc append killed a run at the last cosmetic setup step).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#4,#update-2026-06-03; working/phase5/105_launcher_workspace_sudo_mkdir.sql (header); working/phase5/109a_fix_write_manifest_workspace_perm.sql (header)
- **relations:** run.sh markers (the real success signal); training monitor (fills the observation gap)
- **verify-later:** current command_template in the live launcher def

### run.sh launch chain and RUN_SH_* marker protocol
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** run.sh header ("Emits grep-able RUN_SH_* markers for the future monitor"); NOTES(45) 2026-06-09(4) verified live: "RUN_SH_START → STEP setup → STEP smoke → SMOKE_OK → STEP full_train → RUN_SH_UPLOAD manifest=present".
- **what:** All heavy on-VM work lives in run.sh (setup → smoke train → full train), not in the chassis workflow, so the chain is editable by re-uploading the bundle with no DB migration. It emits a marker protocol to /workspace/train.log (`RUN_SH_START/STEP/SMOKE_OK/FULL_OK/DONE/FATAL`) that is the machine-readable contract for the training monitor's probe. After Phase C, `set -euo pipefail` plus the hard-gated final upload means `RUN_SH_DONE` ⇒ trained AND adapter durable in B2. A mid-train crash leaves no marker (GONE_UNKNOWN).
- **sources:** working/scripts/run.sh (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#run.sh; working/phase5/NOTES_phase5_training_launcher_running(45).md#8
- **relations:** monitor probe classification; scripts bundle as deployment unit; CheckpointUploader
- **verify-later:** run.sh in the live B2 bundle vs the repo copy

### Scripts bundle in B2 as the training deployment unit
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~19:1x: "re-uploading the object IS the whole deploy"; RUNBOOK(8) §4a flat-bundle verification steps.
- **what:** The on-VM scripts (run.sh, 00_vm_setup.sh, 02_train, 03_inference_test) ship as `finetuning/scripts/bundle.tar.gz` in the personae-model-training bucket; the launcher presigns a GET and the VM curls+untars it. The bundle must be flat (files at archive root). Re-uploading the object deploys new training code — no chassis or DB change — with the corollary that editing a script without re-tarring deploys nothing (byte-identical md5 trap). The agent def holds only the key.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x; working/phase5/UPLOAD_bundle.sh; working/phase5/RUNBOOK_iter0_pretrigger(8).md#4a; working/scripts/README_setup.md
- **relations:** run.sh; presigned data plane; SAVE_STEPS re-pack for fast tests
- **verify-later:** b2 ls of finetuning/scripts/; bundle contents vs working/scripts/

### Checkpoint & final-adapter durability via pre-minted presigned PUT manifest
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) status header: "Phases A, B, C BUILT… ckpt-0 confirmed in B2 on run 0ac806ab (the upload path is proven end-to-end)"; "Still unproven in prod is one run reaching RUN_SH_DONE with the final adapter.tar.gz durable".
- **what:** The Thunder VM disk is ephemeral and originally nothing moved training output off it (no checkpoints — save_strategy "no" — and the adapter saved only locally, so a reap = total loss and the monitor's DONE_OK→decommission would have destroyed the artefact). The fix: the launcher pre-mints single-object write-only presigned PUT URLs (K checkpoints + 1 final, keyed `finetuning/checkpoints/<run_id>/ckpt-<index>.tar.gz` and `finetuning/artefacts/<run_id>/adapter.tar.gz`) into `/workspace/upload_manifest.json`; the VM uploads through them. Checkpoints are keyed by save-INDEX not Trainer global_step (fragile to predict); write-once with B2 versioning as backstop; URL expiry must exceed max_uptime (expiry_minutes 3000). Checkpoint upload proven (ckpt-0 in B2); the final-adapter upload and a full RUN_SH_DONE run remain the empirical gate.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md; working/phase5/README_what_is_manifest.md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-5
- **relations:** hostile-VM threat model; CheckpointUploader; resume; monitor enablement gate; batch presign
- **verify-later:** b2 contents under finetuning/checkpoints/ and artefacts/; migrations 109/110/111 state in the live def

### Hostile-VM threat model for the training data plane
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14 "the VM holds no B2 credentials, just a time-limited URL".
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives: standing scoped B2 key on the box (prefix-wide bearer leak risk) and per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** presigned data plane; eval gate; storage credential architecture decision
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

### CheckpointUploader trainer callback (best-effort checkpoints, hard-gated final upload)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** PLAN(7) §02_train "BUILT 2026-06-05… Tier 1 (box-free) PASSED"; NOTES(45) 2026-06-05 update.
- **what:** `02_train` gained gated flags (`--save-steps`, `--save-total-limit`, `--upload-manifest`; defaults keep old behaviour byte-for-byte). A `CheckpointUploader(TrainerCallback).on_save` tars each checkpoint and PUTs it to its save-index URL synchronously (a background thread was rejected — races save_total_limit's dir deletion); checkpoint upload failure is best-effort (log and continue). The FINAL adapter upload is a hard gate: failure raises → non-zero exit → no RUN_SH_DONE → the monitor never treats the box as cleanly done (degrades to GONE_UNKNOWN→failed, never a false DONE_OK). Content-Type application/octet-stream confirmed accepted by the unbound presigned signature. Two manifests coexist: input upload_manifest.json vs the output run-metadata manifest.json.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#02_train; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-05
- **relations:** run.sh markers; durability manifest; save_steps cadence (50 ≈ one checkpoint/1.5–2h, ~2GB each: adapter + AdamW state)
- **verify-later:** 02_train in the live bundle; checkpoint sizes in B2

### Resume path (cluster-side checkpoint selection, presence-of-checkpoints as the signal)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7) 2026-06-14 update: "batch + resume BUILT, APPLIED, and verified [def-state]… still unproven in prod"; migration 111 applied and 2d-verified.
- **what:** A relaunch for the same training_run_id becomes a continuation automatically: the launcher's `check_resume` step asks the adapter (`prepare_resume_url`) to list `finetuning/checkpoints/<run_id>/` in B2 (reusing the existing `storage.Client.ListObjects` — the presumed "list-keys gap" was wrong), pick the highest ckpt-N, and presign a GET; assemble_manifest emits a `resume` block only when found. 02_train downloads/extracts it and calls `trainer.train(resume_from_checkpoint=True)` (restores optimizer/scheduler/RNG/step). No separate resume mode — empty prefix = fresh start; the launcher owns save-index key allocation across resume launches. found=false is a valid answer; transient list failures return error_recoverable so the chassis retries.
- **sources:** working/phase5/README_what_is_manifest.md; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#resume; working/phase5/111_training_launcher_resume_wiring.sql (header)
- **relations:** durability manifest; monitor GONE_UNKNOWN (total-loss case that motivated this)
- **verify-later:** a real kill-and-resume test; dispatch_thunder_prepare_resume_url in registry

### thunder-training-monitor (periodic probe + reconcile + release)
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** NOTES(45) 2026-06-04: "training-monitor VERIFIED live (both paths)… Terminal/decommission branch still never run live… Not enabled"; schedule inserted DISABLED (migration 108).
- **what:** A second periodic lifecycle agent beside the reaper: the `thunder-training-monitor` orchestrator runs `find_active_training_instances` then loops spawn_worker→call_worker per running training box (deliberately NOT the reaper's scheduler-pre_query shape, which merges only the first row per tick and would starve newer instances behind ALIVE boxes). Each `thunder-training-monitor-worker` probes via the adapter's `ssh_get_status` with a status_command that classifies run.sh markers into ALIVE | DONE_OK | DONE_FAIL | GONE_UNKNOWN (plus reachable:false as a valid answer), routes via classifier `next_step` override, reconciles `model_lifecycle.training_runs`, counts consecutive unreachable probes (≥3 → lost → decommission an unreachable-but-billing box), and on terminal verdicts releases the box through the shared idempotent decommission. Built as migrations 106/107/108 + 5 chassis actions. ALIVE path and orchestrator fan-out verified live; the terminal/decommission branch has never fired.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04,#update-2026-06-09-6; working/phase5/107_thunder_training_monitor_worker.sql (header)
- **relations:** thunder-reaper (responsibility split); run.sh markers; monitor enablement gate; reply-topic resolution bug (found here)
- **verify-later:** scheduled_tasks row enabled state; migrations 106-108; the 5 actions in registry.go

### Monitor/reaper responsibility split
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04: "Decision: build a separate thunder-training-monitor, NOT bolted into the time-reaper… the time-reaper is the last-line cost backstop and must stay dead-simple/dependency-free".
- **what:** Two periodic agents with deliberately distinct dependency profiles: the reaper (cost backstop) is pure DB + Thunder and must work even when the adapter is down; the monitor (completion-side) depends on adapter + SSH. They overlap only in calling the shared idempotent `decommission_instance`. The monitor exists because the launcher returns long before training ends (detached run) so completion can't be a workflow await.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md
- **relations:** thunder-reaper; thunder-training-monitor; orphan-sweep TODO (third member: boxes whose Thunder instance vanished)
- **verify-later:** both scheduled_tasks rows and their concurrency_group

### Monitor enablement gate: DONE must mean durable
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** PLAN(7): "enable thunder-training-monitor (safe once DONE means durable — gated on the first run actually reaching RUN_SH_DONE)"; NOTES(45) §9 "Not enabled; enabling is RUNBOOK step 6".
- **what:** An explicit sequencing invariant: the monitor's DONE_OK path decommissions the box (destroying the disk), so the schedule stays disabled until the upload path proves that RUN_SH_DONE implies the adapter is in B2. Enabling early would have destroyed iter_0's adapter. The interim protocol for in-flight runs was manual: scp adapter_out off the box before anything decommissions it.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#build-order; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#update-2026-06-04
- **relations:** CheckpointUploader hard gate; run.sh markers; monitor
- **verify-later:** scheduled_tasks.enabled for thunder-training-monitor; whether a run has since reached RUN_SH_DONE

### Thunder Compute adapter (provision/decommission lifecycle)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 2026-05-12(1) table: phases 2–3.5 "✅ Deployed and verified end-to-end"; FOCUS(25) §14 "Provision loop verified end-to-end (2026-05-22)".
- **what:** A Kafka adapter (`system.adapter.thunder.requests`) wrapping the Thunder Compute GPU API and owning its credentials. `provision_instance`: spend pre-check → ed25519 keypair → API create (public_key sent) → k8s Secret persist → WaitForRunning poll → INSERT thunder_instances with retry → compensating cleanup (fresh context) on partial failure. `decommission_instance`: lookup by provisioning_id or thunder identifier → atomic `decommissioning` transition as idempotency anchor → 404-tolerant API + Secret deletes → cost computed from running_since × snapshotted hourly rate. Error classification maps denial→unrecoverable, infra→recoverable. Includes hard-won API shape knowledge: base URL :8443/v1, lowercase gpu_type enums, camelCase string-numbers in responses vs snake_case ints in requests, recycled numeric ids requiring a partial unique index on live rows, real template names (`base`, not the OpenAPI example).
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md
- **relations:** gpu-provisioner; reaper; ssh_exec; presigned data plane; adapter design guide
- **verify-later:** internal/adapters/thunder/*; migrations 025/028/029; thunder_instances schema

### thunder-reaper scheduled task and per-instance uptime deadline
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) 3.5: "✅ Deployed and verified end-to-end (2026-05-14): synthetic row … picked up within 30s"; NOTES(45) 2026-06-04 live rescue of run fabfd7fa.
- **what:** A 15-min scheduled task whose pre_query finds `running` instances past `max_uptime_hours` and dispatches the idempotent decommission (one per tick, LIMIT 1). The deadline is OURS not Thunder's — computed as running_since + the per-row `max_uptime_hours` (default 18h; training provisions get 18h) — so a mid-train cap overrun can be rescued by bumping the single row's max_uptime_hours (done live: 18→48h when the 24h iter_0 train would have been reaped at hour 18, which with save_strategy=no meant total loss). Reaper reason strings are meaningful text for post-mortems.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#3.5; working/flywheel_docs/README.md#3.5-delivered; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** monitor/reaper split; spend gating; scheduler pre_query single-row semantics
- **verify-later:** migration 028; scheduled_tasks row; thunder_instances.max_uptime_hours column

### Thunder spend gating (DB-side provision check)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "Spend gating lives in DB, not API"; NOTES(45) 2026-06-03 cost-gate check (cap 30, estimate 20, clears with $9 headroom).
- **what:** Before every create, the adapter consults the `thunder_provision_check` view: decommissioned 24h spend + running estimated spend + `estimated_new_run_cost_usd` must stay under `thunder_config.daily_cap_usd`. Operational learnings: keep the per-run estimate realistic (~$20 for a 9h+ A100 run — a $2 test estimate lets doomed runs through; a $25 default blocks legitimate tests); the 24h window is rolling so heavy test days trip the cap on legitimate past spend (raise the cap for the session, don't delete accurate rows).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-172x,#7
- **relations:** thunder adapter provision; reaper; cost anchors from iter_0
- **verify-later:** thunder_provision_check view definition; thunder_config values

### Orphan-sweep for stale live thunder_instances rows
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "TODO (deferred 2026-05-24) — orphan-sweep for stale live rows… Agreed design (not yet built)".
- **what:** Out-of-band deletions (manual `tnr delete`) leave DB rows `running` forever; because live rows hold the recycled Thunder id in a partial unique index, a stale row blocks the next provision with a duplicate-key error (bit on 2026-05-24). Agreed design: a `sweep_orphans` adapter action computes (live DB rows) minus (Thunder's live list) and dispatches the idempotent decommission per orphan, run as a 15–30min scheduled task sharing the reaper's concurrency group, with a safety guard never to sweep on a failed/partial Thunder list. Interim: manual row reconciliation after any manual delete.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics(TODO)
- **relations:** reaper (time-based) and monitor (completion-based) — this is the third, existence-based leg
- **verify-later:** whether sweep_orphans exists anywhere

### Adapter-managed SSH access to GPU boxes (ed25519 keys in k8s Secrets)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §14 "RESOLVED & VERIFIED (Phase 4, 2026-05-24) — SSH connection mechanism + ssh_exec/ssh_get_status" with production verification detail.
- **what:** The adapter generates its own ed25519 keypair per provision, stores the private half in Secret `thunder-ssh-<db_row_id>` (deterministic name so orphan Secrets are reapable), sends the public half on create. `ssh_exec`/`ssh_get_status` dial `instance_ip:ssh_port` directly via x/crypto/ssh as user `ubuntu` (NOT root, despite Thunder's own ssh_command string), with a wait-for-sshd retry (~90s) because RUNNING precedes sshd. The port is the list-endpoint's `port` field, captured into thunder_instances.ssh_port. `reachable:false` is a valid answer, not an error — the probe primitive the monitor builds on. Manual-ops corollary: operators can extract the key from the Secret to watch train.log directly (StrictHostKeyChecking=no needed because Thunder recycles IPs).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; working/flywheel_docs/ssh_probe.sh (header); working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x
- **relations:** monitor probe; setsid launch; RBAC resourceNames trap (Secret permissions)
- **verify-later:** internal/adapters/thunder/ssh/*; the RBAC Role verbs

### Presigned-URL data plane (adapter mints URLs; bytes never transit Kafka)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23".
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (note: `finetuning/` is a folder prefix, not a bucket; the preparer agent-def's `s3_bucket=finetuning` is stale/logical and cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (they now delegate to ObjectURL — one signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&` as &; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh
- **relations:** hostile-VM threat model; storage credential decision; batch presign; docubundle B2 notes
- **verify-later:** data_url_actions.go; TRAINING_BUCKET env; personae-storage-secrets wiring

### Storage credential architecture decision (no storage-adapter service)
- **category:** storage-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "Decision (2026-05-22): hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (max.message.bytes ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs. The acknowledged mess: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix: one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared storage.Client interface forced rebuilding every binary importing platform/storage.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture
- **relations:** presigned data plane; adapters category; debugging guide item 18
- **verify-later:** whether NewDefaultClient exists; secret sourcing per service

### Thunder Prototyping vs Production mode economics
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-05-08 lesson 8: "Prototyping (TGV virtualised) worked fine for 70B inference… Phase 2 should default to Prototyping for inference, Production for training (unverified that Prototyping handles long training runs well)."
- **what:** Thunder's Production mode ($1.79/hr A100 80GB) vs Prototyping ($0.78/hr, TGV-virtualised). Verified: Prototyping is fine for 70B inference. Unverified: whether virtualisation overhead degrades long QLoRA training runs enough to cancel the ~55% saving — flagged as an iter_1 experiment, never run.
- **sources:** working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(8); working/phase2/README.md
- **relations:** cost anchors; gpu-provisioner defaults
- **verify-later:** provision defaults (mode) in gpu-provisioner/adapter config

### Adapter design guide (adapter vs agent vs inline; canonical structure)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) opening: "Canonical guide for building long-running cluster services… Examples drawn from the working adapters in the repository."
- **what:** The decision rule (one external API + multiple internal callers → adapter; long per-orchestration work → spawned agent; short single-agent call → inline; shared infra like DB/Kafka → nothing) plus the canonical shape: struct fields, ordered NewAdapter with manual cleanup on every failure path, sequential fetch-handle-loop (no goroutine-per-message by default), handleMessage parse/dispatch/respond, health endpoints, sync.Once shutdown, topic conventions (`system.adapter.<name>.requests` for new work), config YAML field-name traps, credentials from env only.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md (whole)
- **relations:** thunder adapter (the guide's newest example); response header tiers; deployment essentials
- **verify-later:** consistency of existing adapters with the guide

### Adapter response-header tier taxonomy and the validator-coverage gap
- **category:** adapters
- **status-signal:** partial
- **status-evidence:** FOCUS_adapter_design(3) "TODO — tighten validator coverage… Tier-2 fields are necessary for the orchestration to advance but not validated… Tracking issue: not yet filed."
- **what:** Response headers split into Tier 1 (five fields the platform Validator enforces; `is_error=true` bypasses), Tier 2 (what the chassis needs to route the reply to the awaiting orchestration — `in_response_to_request_id`, message_type, status vocabulary complete/error_recoverable/error_unrecoverable, is_complete/is_error, etc.; missing these means a silent AWAITING_RESPONSES hang the validator won't catch), Tier 3 (observability). Known live consequence of the gap: the matcher fix of 2026-05-22 (typed response-header struct so booleans serialise as real bools — a map[string]string sent string bools and the chassis dropped the reply). Proposal to extend the validator exists but is unfiled.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#sending-responses,#todo; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics (matcher fix)
- **relations:** reply-topic derivation; send-before-register race (same stuck-await symptom family)
- **verify-later:** platform/validation/Validator current coverage

### Adapter deployment essentials (manifest, cluster resources, RBAC, Makefile)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** FOCUS_adapter_design(3) "Deployment essentials — Real lessons from deploying thunder-adapter Phase 2. Every item below is something the deployment failed without."
- **what:** The complete pre-flight for shipping an adapter: serviceAccountName + imagePullSecrets + `command:` (not `args:` — Dockerfiles use CMD, so args replaces the binary path), required Secrets/SA/Docker-Hub grants, explicit KafkaTopic CRDs (Strimzi auto-create is off; missing reply topics fail only at first response), Recreate strategy, single replica, RBAC trap (resourceNames supports no globs — scope by verbs instead for dynamic names like thunder-ssh-<uuid>), four Makefile insertion points and the newName/newTag overlay split, pre/post-deploy checklists.
- **sources:** working/flywheel_docs/FOCUS_adapter_design(3).md#deployment-essentials; working/flywheel_docs/FOCUS_finetuning_flywheel_changelog_addition.md (phase-2 deploy saga)
- **relations:** adapter design guide; wrong-binary image incident; debugging guide §10
- **verify-later:** thunder-adapter kustomize base vs the checklist

### Wrong-binary adapter image incident and the built-vs-running guard
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-14(8): "thunder-adapter:v1.0.1063 actually contains the analyser-adapter binary… Pattern (third deploy-regression in a row)… No guard between built and running. Logged in debug guide v2_47."
- **what:** An overwritten Dockerfile shipped the analyser-adapter binary under the thunder-adapter tag; the pod CrashLoopBackOff'd for ~31h and every provision parked runs at `pending`. Named as the third consecutive "the deploy didn't ship what I thought" regression (109 re-run revert; chassis/adapter tag confusion; Dockerfile overwrite). Prescribed guards: per-build `docker run --rm --entrypoint ls <image> -la /app` before push, never re-push a poisoned tag, and structurally a CI step failing the build if the expected binary is absent — the deploy-side sibling of the migration 2d state check.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-14-8; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** hand-applied migrations lesson; deployment essentials
- **verify-later:** whether the CI image-content guard was ever added

### input_mapping semantics: call_agent-only; config dot-paths for local steps; key_path for loop items
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §2 "[verified-source] Local action steps do not resolve input_mapping… input_mapping is dead config"; 109b header "CORRECTS a load-bearing assumption: input_mapping is NOT live for (local-action) loop substeps."
- **what:** The coordinator honours `input_mapping` only for call_agent (building child input_data) and loop fan-out; on plain local action steps it is dead config. Local actions pull values via config keys whose values are dot-paths resolved from collected_data (`ExtractActionInputs` Strategy 0 / `resolveTemplateToken`); loop substeps read the iteration item via a config dot-path like `key_path:"ckpt_key"` (setLoopVariable puts the item in CollectedData) — using input_mapping there silently falls through to fallbacks (the dataset-key-presigned-40× bug). A proposed coordinator change to resolve input_mapping on local steps was deliberately withdrawn (D1: fix the caller, don't teach the framework a new behaviour for one agent's misuse). Optional mapping fields take a `?` suffix; missing required sources hard-fail (migration 103's fix).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D1,D2),#8; working/phase5/109b_fix_presign_one_loop_item_keypath.sql; working/phase5/103_call_data_preparer_optional_inputs.sql
- **relations:** output_fields contract; launcher workflow; loop_complete convention
- **verify-later:** coordinator ResolveInputMapping; input_mapping.go `?` semantics L101-128

### Child-result shaping: output_fields (plural) contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x: "extractWorkflowResult reads completeStep.Config['output_fields'] — PLURAL only… singular output_field… is never read → falls to the fallback branch that dumps every non-internal collected key"; migration 104 confirmed live.
- **what:** An agent's final result shape is governed by its complete step's `output_fields` array; the singular `output_field` spelling is silently ignored, producing a step-name-keyed fallback dump that breaks consumers' documented paths (`provisioning_result.provisioning_id` buried under `dispatch_provision.response.…`). The resolver auto-unwraps one `.response` per path part but never crosses arbitrary step-name keys. Fix taken at the def level (gpu-provisioner switched to plural + launcher mapping repointed, migration 104) after the user vetoed a chassis change; recorded as debugging-guide gotcha #23. Corollary rule: verify each call_* step's mapped source paths against the producer's REAL collected_data shape before firing anything that books a GPU.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-173x—180x; working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header)
- **relations:** input_mapping semantics; data-path verification runbook step 2b
- **verify-later:** extractWorkflowResult in coordinator; whether other defs still use singular output_field (thunder-reaper was named)

### Reply-topic derivation rules (own topic vs parent topic; two-level await)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) D4 + 2026-06-02 16:12 "D4 CONFIRMED live"; STATUS 06_04 reply-topic orphan fix "VERIFIED 2026-06-04 18:21".
- **what:** Two awaits, two topics: a child's intermediate adapter calls are awaited by the child's OWN coordinator on `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`); only the child→parent final notification uses `__parent_responses_topic__`. Dispatch actions that put the parent topic in an adapter envelope orphan the await (adapter replies where no one listens → infinite hang) — this bit twice (launcher dispatches pre-D4; `dispatch_thunder_ssh_get_status` cloned from ssh_exec). The inherited handoff asserted the opposite convention — corrected against source ("verify against code, not the handoff"). A shared `resolveAwaitResponsesTopic` helper is flagged as the future consolidation; a latent fallback caveat remains (the `system.agent.<type>.responses` fallback doesn't match the launcher's actual `system.responses.training-launcher` topic).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D4),#6,#10; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** send-before-register race; monitor build; adapter header tiers
- **verify-later:** determineResponsesTopic priority order in coordinator; whether the shared helper was built

### Send-before-register await race and preRegisterAwaitedRequest
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09: "Race fix works [verified-log]. Every presign_checkpoints_iter_N_presign_one logged ClaimAwaitedRequest: status_before=waiting … claimed:true"; recorded as the fourth cause of stuck-`waiting` awaits in debugging guide v2_36 §9.
- **what:** Local dispatch actions produced the adapter request and returned await_response:true BEFORE the coordinator inserted the awaited_requests row; a fast (~1s) reply beat the insert, ClaimAwaitedRequest found no `waiting` row, the reply was dropped, and the timeout handler re-dispatched forever with fresh request_ids (RetryVersion pinned at 0). spawn_agent/call_agent don't race because they call `preRegisterAwaitedRequest` (register-before-send, ON CONFLICT DO NOTHING). Fix: the dispatch pre-registers with the same request_id it uses everywhere — one row, one timeout owner; caveats: the helper hardcodes a 120s timeout that wins over step config, and the per-request timeout goroutine is skipped (background expiry sweep is the net). Moving stall point ⇒ race is the diagnostic heuristic.
- **sources:** working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-08,#update-2026-06-09; working/docubundle/CONTEXT_PACK_thunder_checkpoint_race(1).md
- **relations:** O(K²) loop cost (found immediately after); reply-topic rules; awaited_requests machinery
- **verify-later:** preRegisterAwaitedRequest call in thunder_prepare_object_url_dispatch.go and the batch/resume dispatches

### O(K²) loop state-bloat and the batch-presign replacement
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(3): "Full launcher path completed in ~26s… ONE batch await… returned all 40 ckpt PUT URLs… Contrast the retired loop: Version 86 / still at iter_9 nine minutes in. The O(K²) class is gone."
- **what:** Every awaited loop substep re-persists the full orchestration state — the expanded ~80-substep workflow with verbose descriptions, growing collected_data, and ProcessingHistory — so a 40-iteration awaited loop costs O(K²) (iter_0-4 ~2-3s, iter_8 ~100s, then Kafka i/o timeouts) while a GPU bills throughout. Structural cure, not tuning: replace the per-item awaited loop with one batch adapter call (`prepare_object_urls`: keys[]→ordered urls[], reusing the single ObjectURL primitive per key), one await, one persist, no flatten step (migration 110). General platform lesson: awaited loops over cheap local operations are an anti-pattern; batch at the adapter.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09—(3); working/phase5/110_training_launcher_batch_presign(2).sql (header); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-09-update
- **relations:** send-before-register race; loop_complete convention (every production loop ends on an explicit loop_complete substep — checked against all 11 production loops); durability manifest
- **verify-later:** orchestration state persistence cost in coordinator; whether other awaited loops exist at risk

### Hand-applied agent-def migrations: no ledger, re-run reverts, 2d state check
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(7): "re-running 109 silently REVERTED both [110 and 111]… A migration is idempotent only against its OWN prior application, never against LATER migrations that mutate the same object… There is NO migration runner."
- **what:** The flywheel-C def migrations (102–111) are hand-applied jsonb mutations to agent_definitions with no schema_migrations ledger — the def's live shape is the only "did it run" truth. Consequences codified: never re-run an earlier migration "to make sure" (it reverts later ones); run a per-migration state-check query (RUNBOOK 2d) after every deploy and before any launch; back up defs with the sanctioned `snapshot_agent()`/`revert_agent()` (hand-rolled CREATE TABLE backups collide with the existing agent_definitions_backup — discover DB helpers with `\df` first). Optional future hardening: a migration runner or applied_migrations log.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-7,(3); working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#2026-06-14-update
- **relations:** wrong-binary incident (same "shipped what I thought?" family); model swap/revert functions (snapshot_agent reuse)
- **verify-later:** RUNBOOK 2d query vs live launcher def

### agent_definitions source-of-truth: clients_db, not templates_db (for the rich schema)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x CORRECTION: "templates_db.agent_definitions has the OLD schema… holds only the 8 original website-builder agents… PIN (corrected): for the flywheel-C agent_definitions, always read AND patch clients_db."
- **what:** agent_definitions exists physically in BOTH clients_db and templates_db; the architecture doc's "source of truth is templates_db" refers only to the legacy website-builder catalog (old schema, no version column). The chassis loader (filters is_active/is_snapshot, ORDER BY version) can only run against clients_db's rich schema — so all flywheel-C and modern defs live there. This whipsawed twice in one day (103 first applied to the wrong DB, then the "always templates_db" pin issued and then reversed) — a live example of doc-claims diverging from code, and of why the clients_db copy of one def can silently diverge from the live one.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03; working/phase5/103_call_data_preparer_optional_inputs.sql (header carries the superseded templates_db guidance); working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header carries the correction)
- **relations:** hand-applied migrations; documentation-system (stale doc line in 002_system_architecture.md)
- **verify-later:** chassis definition-loader query; 002_system_architecture.md wording

### Orchestrator wrapper spawning pattern (dedicated pods for workers)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i "Spawning architecture — fully confirmed working"; worker pod agent-training-data-exporter-… observed.
- **what:** To run work in a dedicated spawned Job pod rather than the shared chassis pool: a wrapper agent with `processing_mode:"orchestrator"` at the TOP level of default_config, category='orchestrator' (free text), agent_category='coordinator' (CHECK-constrained — 'orchestrator' is not allowed), running `spawn_agent → call_agent(target_role=…) → complete_workflow`; the worker uses processing_mode:"task"/specialist. Includes the three-confused-columns trap on agent_definitions (category vs agent_category vs status; reference row improvement-loop) and ON CONFLICT (type,version). The monitor later reuses the same pattern in a loop (per-instance spawn+call, sequential).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i,#chassis-action-design-patterns; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data-exporter v3; monitor orchestrator; ExtractActionInputs canonical pattern
- **verify-later:** spawn decision logic in chassis (processing_mode placement)

### pgbouncer long-transaction fragility → per-batch commits
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i v3.0 failure "'bulk insert 500 rows: driver: bad connection'" and the v3.1 restructure; §14 pattern entry.
- **what:** Long-held transactions through pgbouncer (transaction pool mode) trip connection-level failures; bulk work defaults to per-batch commits (batch 100, each under a second) with single-statement non-tx bookends. Companion rule from the same incident: always check RowsAffected() on single-row UPDATEs and error rather than warn — an action can return perfect counts while its final UPDATE silently didn't land.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4i,#14; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data export v3.1/v3.2
- **verify-later:** training_data_export.go batch logic

### CLI/ops data-transfer pitfalls (kcat heredoc, COPY-vs-psql, kubectl exec/cp)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** PATCH_2026-05-06 (both bugs validated + corrected command); FOCUS(25) §2.4f v2 smoke retest (kcat heredoc mis-routing); HANDOFF 2026-04-23 lesson 7.
- **what:** A cluster of verified transfer traps: (1) Kafka trigger JSON must be flat single-line via here-string — multi-line kcat heredocs mangle payloads silently and route to a No-op handler; (2) `COPY … TO STDOUT` is not JSON-safe for jsonb (double escape layers) — use `psql -tAXc` with plain SELECT for JSONL; (3) `kubectl exec -i` without consumed stdin sporadically truncates stdout (1716/1958 rows, "next reader: unexpected EOF"); (4) `kubectl cp` truncates large files silently — use `exec cat > local`; (5) `tnr scp` of directories nests `{dest}/{source_basename}/` both ways.
- **sources:** working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted_PATCH_2026-05-06.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#14; working/flywheel_docs/HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(4)
- **relations:** dataset pull path; 016 debugging guide §9
- **verify-later:** 01_pull_dataset_from_postgres.sh uses the corrected form

### configOrInput numeric config coercion (expiry_minutes silently dropped)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-09(5): "expiry_minutes override — FIXED… configOrInput read config via Config[name].(string), so the JSON-number 3000 failed the assertion → fell through → adapter default". Debug guide v2_43.
- **what:** The shared configOrInput helper type-asserted config values to string, so JSON-number config (expiry_minutes:3000, timeout_seconds) silently fell through to defaults — presigned PUTs came back at 24h instead of 50h. Fixed with a `coerceConfigScalar` (string/float64/json.Number/int/bool). Class lesson: shared config readers must coerce scalars, and a numeric setting "applied" in a def is only proven by observing the effect (X-Amz-Expires on the URL).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-4,(5)
- **relations:** presigned data plane expiry caveat; launcher dispatch family
- **verify-later:** coerceConfigScalar in thunder_ssh_exec_dispatch.go

### Scheduler-fired chassis-resident observability gotcha (owner_agent_type='generic')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** STATUS 05-12(1) architectural follow-ups: "filtering orchestration_states by owner_agent_type MISSES top-level chassis-resident workflows, which are owned by 'generic'".
- **what:** Scheduler-fired agents that run in the generic chassis (thunder-reaper, build-pipeline-trigger, etc.) have orchestration_states.owner_agent_type='generic'; the real agent type lives at `collected_data->'config'->>'agent_type'` and orchestration_name follows `sched-<task>-<ts>`. Filter on those instead. Related cosmetic anomaly, unresolved: a stale non-DB agent_config stub (old reaper-style no-op) persists in message envelopes across redeploys while the full WorkflowPlan executes — source of the cached representation never found.
- **sources:** working/flywheel_docs/STATUS_thunder_adapter_2026-05-12(1).md#6; working/phase5/NOTES_phase5_training_launcher_running(45).md#stub-source-narrowed
- **relations:** monitor testing; debugging guide
- **verify-later:** where the stale agent_config envelope field loads from

### knowledge_base RAG store and flywheel B verification
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b action items all [x]; "Flywheel B is done" (step 3 chassis integration COMPLETED on v1.0.979, 2026-04-21).
- **what:** `knowledge_base` (migration 082): pgvector(768) for nomic-embed-text, shared across agents via `rag_lookup`/`rag_index` actions, trigram fallback when Ollama is down, SHA256 dedup, metadata-first filtering doctrine (filter by vertical/component_type/content_type/source before ranking by similarity — else a vet example surfaces for gas-wholesale copy). Verified bottom-up in three single-focus steps: pgvector+ivfflat+cosine on synthetic vectors, real-content retrieval through cpu-ollama, then chassis integration via a deterministic 3-step `rag-test-agent`. Nomic v1 judged good enough; v2-moe named as a drop-in upgrade (same 768 dims). Open task: periodic REINDEX of the ivfflat index.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.2,#2.4b; working/flywheel_docs/flywheel_B_step*.{sql,sh} (headers)
- **relations:** nomic prefixes; RAG-platform product; three channels
- **verify-later:** rag_actions.go; knowledge_base row counts; whether REINDEX ever scheduled

### Nomic task prefixes are load-bearing (rag_actions prefix patch)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4b "[x] Prefix patch deployed and verified live 2026-04-21… log line 'prefix_applied':true observed".
- **what:** Without `search_document:`/`search_query:` task prefixes, nomic embeddings ranked a Labrador chunk above the French Bulldog chunk on a BOAS-specific query; with prefixes the correct result won with 5× the margin. The patch adds a model-scoped `applyNomicPrefix` helper (only nomic-embed-*, double-prefix guard, prefix_applied logged) at both embed call sites; stored chunks and dedup hashes stay unprefixed; trigram fallback untouched.
- **sources:** working/flywheel_docs/PATCH_rag_actions_nomic_prefixes.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4b
- **relations:** knowledge_base RAG; Ollama CPU adapter
- **verify-later:** rag_actions.go contains applyNomicPrefix

### Ollama CPU adapter operational rules
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4c "RWO PVC rolling-restart deadlock on ollama-adapter — resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; §14 Ollama specifics list.
- **what:** Hard-won ops rules for CPU Ollama: Recreate (not RollingUpdate) deployment strategy because the PVC is RWO (classic new-pod-can't-mount deadlock); `OLLAMA_LOAD_TIMEOUT=10m` + `OLLAMA_KEEP_ALIVE=30m` (default 60s load timeout killed first inference after cold start — 14.4GB model loads in ~45s); pod memory limit ≥ model file size + 8–12GiB headroom (Ollama reads host /proc/meminfo but is constrained by the cgroup — misleading OOM messages); chassis calls `/api/chat` not `/api/generate`; measured CPU throughput ~150 tok/s prompt, ~2.5 tok/s generation on mistral-small3.1.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4c,#2.4d,#14
- **relations:** ai_endpoint_health; dedicated eval pod; CPU eval abandonment
- **verify-later:** ollama-adapter kustomize base

### ai_endpoint_health inference routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.3 endpoint table + §4.1 "[x] Endpoint health routing deployed"; gpu-ollama noted "currently DOWN, not always-on" and still DOWN in HANDOFF 2026-05-08.
- **what:** Three endpoints tracked in `ai_endpoint_health`: claude (default quality), cpu-ollama (embeddings + small models), gpu-ollama (70B/LoRAs — persistently down through this doc set). Healthy endpoint → work claims flow; unhealthy → items wait/back-to-triage, so GPU availability gates work without a separate batch scheduler. The kafka-scheduler only probes endpoints listed here — which is also why the dedicated eval pod stays invisible to production routing.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.3,#4.3
- **relations:** model swap; dedicated eval pod; deployment path options for the trained adapter
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

### Model swap / revert functions
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4 (migration 083) + §4.1 "[x] Swap / revert functions deployed"; snapshot_agent used live by migration 110's backup step.
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` — per-agent per-step snapshot-before-swap of the ai_service block in agent_definitions.default_config, with full-table backup as the nuclear option. This is the deployment mechanism a green flywheel-D verdict would use to move an agent from Claude to a local model, and snapshot/revert doubles as the sanctioned def-backup tool for hand-applied migrations. No doc records an actual production model swap having happened.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4; working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-3
- **relations:** eval gate; deployable_adapters view; hand-applied migrations
- **verify-later:** migration 083 functions; agent_definitions_backup contents

### finetuning.uk product strategy (RAG platform flagship, data curation as the product)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN "Last touched: 2026-04-21… Numbers are planning estimates"; FOCUS(25) §7 decisions 11-12; no later doc records any milestone reached.
- **what:** The external-product thesis: finetuning.uk becomes a RAG platform for technical-adjacent SMEs (10-50 people, knowledge-intensive, UK/EU) whose differentiator is automatic data curation (parse/classify/dedup/quality-score/PII-scan/inconsistency-flag with a visible curation report) — "competitors treat bad data as the customer's problem; we treat it as the product." RAG chosen ahead of text/image LoRA tiers (users arrive with docs, not training pairs; the infra is built). Self-service *fine-tuning* SaaS is explicitly deferred/not-shipped. Reuse map: same Ollama/Unsloth/export/eval plumbing; entirely new: multi-tenancy, billing, UI, support, legal. Week-1 technical items named: tenant_id on knowledge_base enforced in rag_lookup/rag_index; auth stack choice.
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#5-13
- **relations:** knowledge_base RAG; internal flywheel (shared plumbing); UI-first decision
- **verify-later:** tenant_id on knowledge_base; any finetuning.uk app code; site_specs for finetuning.uk

### finetuning.uk business plan (pricing, unit economics, milestones)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN changelog: "2026-04-21 — Initial draft… refine as data arrives"; no subsequent updates in this tree.
- **what:** Solo-operator plan: tiers Trial/£199/£499/£1,499/Enterprise plus concierge fees (£750 audit → £15-30k bespoke); gross margin 57-78% per Growth customer; break-even ~5 Growth customers; 12-month target £9-12k/month and ~£100k year-1 revenue; content-led cold acquisition only (the framework as content engine is the claimed structural moat); interim gigs capped at 50% of time; assumption list with explicit 60-day tests; milestone/decision gates at months 1/3/6/12. Superseded staging inside the family: v1's "concierge first, UI later" three-tier structure was replaced by the UI-first "build our own cockpit" revision (2026-04-21 fourth pass).
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service_v1.md#8,#10 (superseded shape); working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#8,#10
- **relations:** product strategy; vonc/business-strategy docs elsewhere
- **verify-later:** nothing technical; check for any later business docs superseding this

### Docubundle context packager (thunder-checkpoint-race package)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** docubundle/README.md usage note + the generated 633KB production context file dated in-tree; packager header "Patterned on package_page_build_debug.sh".
- **what:** A self-contained packager script (`package_thunder_checkpoint_race.sh`) that bundles the async-await + loop machinery of the chassis, the checkpoint-upload path, the working docs, and optionally a read-only live capture (schemas, decisive queries, workflows, runtime state) into one context file to seed a fresh AI-assistant thread on a specific blocker. Paired with hand-written CONTEXT_PACK / NEXT_CHAT_MANIFEST docs that state the blocker, the verified root cause, the applied fix, and next actions. An instance of the wider bundle/context-package pattern (cf. docs019 contextkit) applied to the finetuning workstream; the targeted CHASSIS_await_loop_extract ("use the targeted extract, not the 72k-line file") shows deliberate context-size curation.
- **sources:** working/docubundle/README.md; working/docubundle/package_thunder_checkpoint_race.sh (header); working/phase5/NEXT_CHAT_MANIFEST.md; working/phase5/CHASSIS_await_loop_extract.txt (header)
- **relations:** diagnosis-loop bundles/contextkit; send-before-register race (its subject)
- **verify-later:** relation to z_bundles/context_packages tooling at repo root

### Epistemic tagging and handoff-correction discipline
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** NOTES(45) header: "Epistemic tags used below: [verified-source]… [verified-db]… [deployed?]… [assumed]… [gap]"; §10 "Handoff-correction log (institutional memory)… Pattern: verify against code, not the handoff."
- **what:** The phase-5 notes operate a working epistemology: every claim carries a tag distinguishing read-from-source, confirmed-by-production-query, assumed, or known-gap; and a dedicated correction log records where inherited handoffs contradicted deployed reality (reply-topic direction, prepare_object_url existence, the "list-keys gap" that already existed as ListObjects). Multiple bugs in this unit trace to trusting a doc over code (templates_db pin, backup-vs-live def divergence, runbook "safe to re-run"). This is a documentation-system convention worth institutionalising: docs are claims; code and DB state are evidence.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#header,#10; working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md (correction notes throughout)
- **relations:** docs026 programme itself (stage-2 verification mirrors this); hand-applied migrations lesson
- **verify-later:** n/a (convention)

### Reuse-first build discipline (grep before adding; delegate, don't parallel)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04 reuse audit ("ssh_get_status ALREADY EXISTS… no adapter change"); D3 "Reuse over parallel code in the adapter presigner"; guideline audits run against 001/002/003 for every artifact batch.
- **what:** A repeatedly-exercised discipline in this workstream: before building, audit what exists (ssh_get_status reused as the monitor probe; ListObjects reused for resume; DatasetURL/ArtefactURL refactored to delegate to ObjectURL rather than a third signer; datahelpers GetIntField over a custom helper; preRegisterAwaitedRequest reused for the race fix). Each new artifact batch is audited against the dev guide/architecture/contracts docs before deploy, with violations fixed or explicitly accepted (the one accepted tradeoff: launcher reading through the provisioner's step name).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#3(D3),#update-2026-06-04(reuse-audit),#guideline-audit,#update-2026-06-05(guideline-audit)
- **relations:** adapter design guide; input_mapping/output_fields contracts
- **verify-later:** n/a (practice); the accepted step-name coupling in call_launcher mapping

### Kafka topic-creation race self-heal (transient "Topic not yet on broker")
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-06: "Transient `Topic not yet on broker` for the launcher .responses topic self-healed on attempt 2 (topic-creation race) — normal."
- **what:** Per-spawn child topics (`job.<id>.requests`, per-agent responses topics) are created on demand; a first-publish race against broker propagation produces a transient failure that retries resolve. Recorded so it isn't chased as a real fault. Contrast: a *permanently* missing topic (Strimzi auto-create off) fails every attempt — the distinguishing signature is self-heal on retry.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-06; working/flywheel_docs/FOCUS_adapter_design(3).md#required-cluster-resources
- **relations:** adapter deployment essentials (KafkaTopic CRDs)
- **verify-later:** topic auto-creation settings for spawned-agent topics vs adapter topics

## Proposed NEW categories

None. Everything in this unit fits existing seed slugs — predominantly `finetuning-flywheel`, with `model-infrastructure` (Thunder/Ollama/endpoint/swap), `adapters`, `storage-architecture` (presign/credential boundary), `development-guide`/`debugging` (chassis contracts and failure signatures), `business-strategy` (finetuning.uk), `diagnosis-loop` (docubundle), and `documentation-system` (epistemic tagging).

## Cross-cutting flags for stage 2

- Hardcoded Thunder API bearer token committed in `working/flywheel_docs/ssh_probe.sh` — credential hygiene check.
- Persistent open items to verify: monitor schedule enabled?; first RUN_SH_DONE + final adapter in B2?; orphan-sweep built?; model-trainer call_agent fall-through fixed?; validator Tier-2 coverage extended?; iter_1 ever trained (fp16, 2-epoch, `<no value>` filter)?; any production model swap executed?
