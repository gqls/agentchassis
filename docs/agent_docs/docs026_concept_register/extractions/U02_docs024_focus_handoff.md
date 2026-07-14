# EXTRACTION U02 — docs024_key_docs_latest FOCUS / HANDOFF / ANALYSIS / ASSESSMENT / PLAN / ARCHITECTURAL / FUTURE files
Extracted 2026-07-13. Files in scope: 38. Concepts found: 105.

## Coverage
| file | treatment |
|---|---|
| ANALYSIS_chassis_response_consumer_group_race.md | full |
| ARCHITECTURAL_TENSIONS(3).md | family-latest |
| ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md | full |
| FOCUS-css_js_mechanisms.md | full |
| FOCUS_adoption_faithfulness_via_locks(2).md | family-latest |
| FOCUS_collected_data_analysis.md | full |
| FOCUS_content_quality.md | full |
| FOCUS_dispatch_diagnostic(4).md | family-latest |
| FOCUS_finetuning_flywheel_and_service(13).md | family-latest |
| FOCUS_interactive_content_generation(4).md | family-latest |
| FOCUS_internal_linking.md | full |
| FOCUS_language.md | full |
| FOCUS_llm_reliability_for_component_generation.md | full |
| FOCUS_naming_conventions_kebab_vs_snake.md | full |
| FOCUS_navigation.md | full |
| FOCUS_navigation_HANDOFF_navigation_fix.md | full |
| FOCUS_navigation_errors_to_be_fixed.md | full |
| FOCUS_page_build_handler_silent_completion.md | full |
| FOCUS_prompt_composition_pattern.md | full |
| FOCUS_site_spec_vs_site_plan.md | full |
| FUTURE_adoption_source_destination_separation.md | full |
| HANDOFF-pipeline-triage-april-2026.md | full |
| HANDOFF_2026-04-17_component_rendering_js_separation_quality.md | full |
| HANDOFF_2026-04-17_nav_empty_sections_footer(1).md | family-latest |
| HANDOFF_2026-04-17_triage_and_component_linking.md | full |
| HANDOFF_2026-04-18_design_and_styling_composable_theme_and_site_design_planner.md | full |
| HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md | full |
| HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md | full |
| HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md | family-latest |
| HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug.md | family-delta |
| HANDOFF_2026-04-20_composition_deployed_design_stuck.md | full |
| HANDOFF_2026-04-20_error_investigations.md | full |
| HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md | family-latest |
| HANDOFF_2026-04-23_dispatch_reliability_and_008_validated.md | family-delta |
| HANDOFF_2026-05-26_design_imagery_triggers_and_adoption_diagnosis.md | full |
| HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md | full |
| PLAN_imagery_loop_closure.md | full |

Family-delta notes: `HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug.md` (base) claimed a root cause for the `mode: rewrite` save-skip ("content writer produces output shape save_page_sections can't read") which version (2) retracts — investigation could not explain the skip and closed it as "don't pass unsupported mode values". `HANDOFF_2026-04-23_dispatch_reliability_and_008_validated.md` (base) lacks items 19 (discovery agents should skip dead/stub sites) and 20 (duplicate sites-row on re-adoption question), which appear only in (1). No concepts were dropped between versions.

---

## Concepts

### Finetuning flywheel four-lane programme (A export, B RAG, C training, D eval)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel A (data export) + B (RAG) done. Flywheel C (training) scripted, awaiting first run on GPU VM. Flywheel D (eval) paused." (doc last touched 2026-04-23)
- **what:** The internal AI training flywheel: production LLM calls become training data (A), verified knowledge feeds RAG (B), local models get fine-tuned on the exported data (C), and quality is compared against Claude (D), so local models replace API calls where quality holds and costs drop.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#1, #2, #15-changelog
- **relations:** llm_call_log capture; training-data export pipeline; Flywheel C QLoRA; Flywheel D replay-eval; three compounding improvement channels
- **verify-later:** `training_exports` schema; `flywheel_A_v3/`, `flywheel_C/` script dirs; `llm_call_log`; whether a first GPU training run ever happened after 2026-04-23

### llm_call_log LLM call capture
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "llm_call_log populating (flywheel columns wired through ai_actions.go)" [x]; but "Cleanup function exists (cleanup_old_llm_logs) but nothing schedules it yet" (2026-04-23)
- **what:** Every LLM call writes agent_type, step_name, model, prompt_rendered, response_text, tokens, latency, success, work_item_id, prompt_variant, vertical, rag_context_used to `llm_call_log` (migrations 081/085) via a fire-and-forget `LogLLMCall` goroutine. Retention 90/180 days; the join key for "what to train".
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.1, #4.2
- **relations:** training-data export; prompt evolution (prompt_variant A/B); vision cost tagging (PLAN_imagery_loop_closure 5.4)
- **verify-later:** `ai_actions.go` LogLLMCall; scheduling of `cleanup_old_llm_logs`; agent_type population on old rows

### knowledge_base RAG with load-bearing nomic task prefixes
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "Prefix patch deployed and verified live 2026-04-21 … log line \"prefix_applied\":true observed on rag_lookup step"; "Flywheel B is done" (chassis v1.0.979)
- **what:** `knowledge_base` table (migration 082, pgvector 768-dim, ivfflat+cosine) readable/writable by any agent via `rag_lookup`/`rag_index`, with trigram fallback when Ollama is down. Empirically established that `search_document:`/`search_query:` prefixes for nomic-embed-text are mandatory for correct ranking; metadata-filter-first-then-similarity is the retrieval rule.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.2, #2.4b
- **relations:** finetuning.uk RAG product reuses this infra; cpu-ollama adapter
- **verify-later:** `platform/orchestration/actions/rag_actions.go` applyNomicPrefix; `knowledge_base` metadata fields (vertical, component_type, source_quality)

### Training-data export as chassis agent + action, Postgres-backed
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "First real training dataset now in Postgres: export_id fef7be6b…, 1,958 rows, 21.2MB, reconciled manually. Spawning architecture fully validated" (2026-04-23); v3.2 strict-UPDATE code "awaiting chassis rebuild/deploy" at doc date
- **what:** Export of (prompt, response) pairs from llm_call_log as ChatML messages with metadata sidecar (source_log_id, agent_type, step_name, export_version), fence-stripped and JSON-validated, via `training_data_export` action + `training-data-exporter` worker + `training-data-export-orchestrator` wrapper. v3 writes named snapshot datasets into `training_exports.runs`/`training_exports.rows` (rejected: real-time streaming — batch snapshots preserve "the dataset we trained on"). Evolved v1 (template config, failed) → v2 (file output, wrong pod) → v3 (Postgres, dedicated pod).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e, #2.4f, #2.4g, #2.4i
- **relations:** orchestrator wrapper pattern; pgbouncer per-batch commits; negative-examples/DPO decision
- **verify-later:** `training_exports` schema in DB; `training_data_export_v3.go` and registry entry; whether v3.2 landed

### Flywheel C QLoRA training pipeline (Unsloth, Llama 3.3 70B)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "Pipeline shape now concrete, scripts written, awaiting first training run on a GPU VM" (2026-04-23)
- **what:** Five scripts (`flywheel_C/00-03` + README) pull a named export from Postgres, train a LoRA adapter on Llama 3.3 70B Instruct via Unsloth QLoRA on a single H100/A100, sanity-check JSON validity of outputs. 70B chosen because hardware was available; 8B flagged as likely 95% of quality at 10% cost — comparison run planned.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5
- **relations:** Flywheel C phase 2 automation; Flywheel D eval harness scores the result
- **verify-later:** existence of `flywheel_C/` scripts; any `model_training_runs` table or LoRA artefacts

### Flywheel C phase 2 automation (chassis drives, GPU VM serves)
- **category:** finetuning-flywheel
- **status-signal:** aspirational
- **status-evidence:** "design locked, not built … Gated on phase 1 (manual training run)" (2026-04-23)
- **what:** HTTP job server (~200 lines, POST /jobs → poll → fetch adapter, bearer auth) on the GPU VM; three new chassis components (model-trainer, model-evaluator specialists, training-flywheel-orchestrator wrapper chaining export→train→eval→conditional swap_agent_model); three new tables (model_training_runs, model_artefacts, model_evaluations). Rejected SSH-exec and Kafka-consumer-on-VM alternatives.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.5.1
- **relations:** model swap/revert functions; scheduled_tasks trigger option
- **verify-later:** whether any of the three agents/tables exist in agent_definitions / DB

### Flywheel D replay-eval methodology + dedicated ollama-eval pod
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "Flywheel D (eval) paused"; ollama-eval pod deployed with memory bump "requests: 24Gi / limits: 28Gi … fix persisted into kustomize base" (2026-04-23)
- **what:** Replay-don't-re-run evaluation: pull 20 diverse stored Claude prompts (DISTINCT ON orchestration_id), POST to candidate model via /api/chat, compare with 3 levels (structural jq checks → Claude-as-judge → manual review). Shared cpu-ollama contention made eval impractical (27 min/case), so a dedicated `ollama-eval` deployment (own PVC/service, invisible to production routing) was created. Memory rule: pod limit ≥ model file size + 8–12 GiB.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4d, #2.4d-comparison
- **relations:** Flywheel C evaluation preconditions; Ollama CPU ops envelope
- **verify-later:** `deployments/kustomize/services/ollama-eval/`; results.jsonl outcome; whether eval ever completed

### Negative examples reserved for DPO, excluded from SFT
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** "For our first training run: plain SFT, edge cases excluded" (decision 2026-04-22)
- **what:** Edge-case rows where Claude correctly produced prose instead of JSON are positive examples of the wrong shape for SFT and must be excluded from exports; they become the "rejected" side of DPO preference pairs later. Keeps them in llm_call_log, out of training_exports.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4e "Negative examples / edge cases"
- **relations:** training-data export filters (strict_json)
- **verify-later:** export action's strict_json / prose-exclusion behaviour

### Three compounding improvement channels (RAG, LoRA, prompt evolution)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** "From 009_model_infrastructure.md, decision 10" — RAG immediate, LoRA medium-term, prompt_variant A/B ongoing (2026-04)
- **what:** Three ways local model quality improves, each useful alone but compounding: RAG injects verified knowledge now; LoRA replicates a task cheaply later; `prompt_variant` column enables prompt A/B evolution. Good prompts + good RAG produce the best training data.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#3
- **relations:** flywheel programme; llm_call_log prompt_variant
- **verify-later:** any actual prompt_variant usage in llm_call_log

### Fine-tune candidate selection heuristic
- **category:** finetuning-flywheel
- **status-signal:** unknown
- **status-evidence:** Priority table (2026-04): knowledge-extractor, site-classifier, vet-practice-verifier, briefing-agent, domain-analyst, content-researcher good; page-content-writer/visual-design-auditor/chief-strategist "not good candidates"
- **what:** Agents with high-volume, structured-JSON, short outputs are swap candidates for local models; long creative output and judgement tasks stay on Claude. Drives which agent/step gets exported and trained first.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.6
- **relations:** Flywheel D target choice (page-content-writer iter_0 chosen for eval despite being a poor swap candidate, because it had the most logged data)
- **verify-later:** actual llm_call_log volumes per agent

### finetuning.uk self-service product strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Not started. Questions to answer before scoping"; 12 decisions dated 2026-04-21; shipping ladder "Aspirational dates, not promises"
- **what:** finetuning.uk as both a credible knowledge site and a revenue product: flagship = RAG platform with data curation as a first-class visible feature (parse/classify/dedupe/quality-score/PII-scan/inconsistency-flag pipeline reusing the framework), concierge-onboarded then self-serve; tiers from £199/mo platform to £15-30k bespoke; target user technical-adjacent SMEs; UI-first build as own operational cockpit; explicit not-to-ship list (multi-tenant fine-tuning SaaS, public API). Differentiation from positioning (UK residency, opinionated simplicity, self-improvement loop) not engineering.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#5, #7, #8, #8a, #10, #11
- **relations:** internal flywheel infra reuse table (#6); knowledge_base tenant_id plan
- **verify-later:** state of finetuning.uk site; any tenant_id on knowledge_base

### AI endpoint health routing
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Endpoint health routing deployed" [x] (2026-04); gpu-ollama "currently DOWN, not always-on"
- **what:** `ai_endpoint_health` (migration 085) tracks claude / cpu-ollama / gpu-ollama endpoints; healthy endpoint → work claims flow, unhealthy → items wait (back-to-triage). No separate batch scheduler for GPU: it's either healthy or not. Kafka-scheduler probes only endpoints listed here, so unlisted pods (ollama-eval) are invisible to production routing by design.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.3, #2.4d
- **relations:** Flywheel D dedicated eval pod; rate-limit transient classification
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state

### Per-agent model swap / revert
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Swap / revert functions deployed" [x] (migration 083)
- **what:** `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` safely snapshot an agent's `ai_service` block in agent_definitions.default_config before swapping model, per-agent per-step; full-table backup remains as the nuclear option. The mechanism the flywheel's deployment decision hangs on.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4; HANDOFF_2026-05-26 (snapshot_agent used before jsonb_set edit)
- **relations:** training-flywheel-orchestrator conditional swap
- **verify-later:** migration 083 functions in DB

### Ollama CPU operations envelope
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** "resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; memory rule from 2026-04-23 OOM incident
- **what:** Hard-won ops facts for CPU Ollama: RollingUpdate + RWO PVC deadlocks (use Recreate); OLLAMA_LOAD_TIMEOUT=10m / KEEP_ALIVE=30m; cold load ~45s for 14GB models; pod memory limit ≥ model size + 8–12GiB (cgroup, not host, is what Ollama constrains against); /api/chat not /api/generate; throughput ~150 tok/s prompt, ~2.5 tok/s generation for mistral-small3.1 Q4 on 8 CPU cores.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4c, #14 "Ollama specifics"
- **relations:** dedicated eval pod; endpoint health
- **verify-later:** kustomize base for ollama-adapter

### Orchestrator wrapper pattern for dedicated pod spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Spawning confirmed working (2026-04-23 trigger test) … processing_mode: 'orchestrator' at top level + agent_category = coordinator IS the combination that produces a dedicated spawned pod"
- **what:** To run work in its own spawned pod rather than one of the three shared chassis pods: an orchestrator wrapper (category=orchestrator, agent_category=coordinator, processing_mode=orchestrator at top level of default_config) with steps spawn_agent → call_agent(target_role, not agent_type) → complete_workflow, calling a worker (specialist, processing_mode=task). Input mapping maps fields individually with `?` suffix for optionals — never a whole `input_data` blob. File writes from non-spawned actions land on a random chassis pod and die with it.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f "Operational gotcha", #2.4h, #14 "Chassis action design patterns"
- **relations:** agent_definitions three-column semantics; training-data-exporter as reference implementation
- **verify-later:** training-data-export-orchestrator agent definition as the canonical example

### agent_definitions three-column semantics (category / agent_category / status)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Caught three agent_definitions column semantics confusions" (2026-04-23); reference row improvement-loop = category=orchestrator, agent_category=coordinator, status=experimental
- **what:** `category` is free-text functional role; `agent_category` is CHECK-constrained to strategist/executor/analyst/integrator/coordinator/specialist (NOT orchestrator); `status` is lifecycle. Naïve writes put lifecycle values in the wrong slot. Also: ON CONFLICT must target (type, version) with `WHERE deleted_at IS NULL`.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4h, #14
- **relations:** orchestrator wrapper pattern
- **verify-later:** CHECK constraint on agent_definitions.agent_category

### Chassis action input conventions (ExtractActionInputs / input_data, dual registration)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Rewritten to use the canonical pattern … matches every other action in the codebase" (2026-04-23); registry gap bit on 2026-04-20: composition actions "had NO entry in GlobalActionRegistry … rejected as 'requires a topic'"
- **what:** New actions use `datahelpers.RegisterActionInputSpec` in init() + `ExtractActionInputs` (5-strategy cascade) rather than raw ExtractNestedFieldString; parameters flow via `CollectedData["input_data"]` because `{{.input_data.X}}` templating does NOT render for deterministic-action step config; every new action needs BOTH the InputSpec registration AND a GlobalActionRegistry entry with IsLocal:true; results land in collected_data under output_field, never final_result. Config literal numbers must be read with `datahelpers.GetIntField(params.StepConfig.Config, …)` — `inputs.GetInt` reads collectedData, not config.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#14; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#2; HANDOFF_2026-04-17_triage_and_component_linking.md#1
- **relations:** CollectedData architecture; field-name collision risk
- **verify-later:** datahelpers package; registry.go

### pgbouncer per-batch transaction discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** v3.0 "driver: bad connection" failure → "v3.1 split into per-batch transactions, batch size 500 → 100 … Per-batch commits worked" (2026-04-23)
- **what:** Long-held transactions through pgbouncer (transaction pool mode) are fragile — bulk work must commit per small batch (<1s each), never wrap a streaming job in one transaction. Companion rule: check RowsAffected on single-row UPDATEs and fail loudly instead of Warn+continue (v3.1's final UPDATE silently didn't land).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4i, #14
- **relations:** training-data export v3 evolution
- **verify-later:** pgbouncer pool mode config; v3.2 UPDATE handling

### Kafka trigger payload discipline (flat single-line JSON here-strings)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "kcat heredoc was silently mis-routing messages to a 'No-op scheduled task' handler … Documented as permanent ops pattern in 016_debugging_guide_v2.md §9" (2026-04-23)
- **what:** Multi-line heredocs mangle kcat JSON payloads silently (routing falls through to no-op handlers with input_data null). Use `<<<'{…flat json…}'` here-strings or jq -nc. Related manual-trigger pattern: psql jsonb_build_object → pipe to kcat with standard headers, used to trigger handlers directly when dispatch is blocked.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f v2 smoke retest, #14; FOCUS_dispatch_diagnostic(4).md#Workarounds
- **relations:** dispatch workarounds; debugging guide §9
- **verify-later:** 016_debugging_guide §9 entry

### Work-item state machine (detected → triaged → claimed → complete/failed)
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15)
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step (registry.go:722) promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler / idx_swi_site_pending); handlers mark complete/failed (mark_work_item_complete / mark_work_item_failed steps). There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged.
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md
- **relations:** dispatch chain; auto-triage open question; two-strike rule; silent completion
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

### Dispatch chain: build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; scheduled_tasks row build-pipeline-trigger every 30s
- **what:** The scheduler fires build-pipeline-trigger, whose find_dispatchable_site step picks ONE site per tick (DISTINCT ON with no outer ORDER BY — effectively arbitrary among eligible sites) and spawns build-dispatch-loop scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers. Throughput cap ~5 items per site per 30s, one site at a time. build-pipeline-trigger doesn't write orchestration_states, making its decisions untraceable.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** NOT EXISTS blocker; Bug 3 site-targeting; fairness ORDER BY improvement
- **verify-later:** scheduled_tasks 'build-pipeline-trigger' row; build-pipeline-trigger / build-dispatch-loop agent definitions

### NOT EXISTS whole-site claim blocker
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "A single stuck claimed item on a site excludes the entire site from dispatch consideration until it clears … by design … but it makes stuck claims a system-stopping condition" (2026-05-15)
- **what:** find_dispatchable_site's NOT EXISTS clause excludes any site with ANY item in status='claimed' — an absolute blocker, not a deprioritiser. Prevents racing claims mid-execution but converts one dead handler into a site-wide stall. Proposed (cheap, high-leverage, not built): watchdog that resets claims older than ~15 min.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3
- **relations:** claim-timeout sweeper absence (Bug 2); dispatcher stall (Bug 1)
- **verify-later:** whether an auto-reset sweeper now exists

### pipeline column as soft routing label
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet."
- **what:** `site_work_items.pipeline` (renamed from `domain`; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only build-dispatch-loop exists — 'design' and 'maintenance' items sit dormant. It duplicates what handler_agent already implies, with nothing keeping them in sync (the unfulfilled_imagery_plan check emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4
- **relations:** unfulfilled_imagery_plan check; dispatch chain
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

### Silent completion pathology and the positive-evidence rule
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); mode 2 later "already confirmed fixed" per FOCUS_content_quality (2026-06-09); modes 1/3 ran at 66×/47× per week (2026-04-20)
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done". Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle.
- **sources:** FOCUS_page_build_handler_silent_completion.md (whole); HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** claim-timeout mechanism; validate_page_content gate; two-strike rule
- **verify-later:** reaper code paths setting status='complete'; whether modes 1 and 3 were fixed

### Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **category:** NEW:work-dispatch
- **status-signal:** unknown
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes"
- **what:** build-dispatch-loop orchestrations stall at process_item_iter_N_call_handler even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker; consumer group race; silent completion
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

### build-pipeline-trigger site targeting via pre_query
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23)
- **what:** The scheduled dispatcher fires with no site targeting so it lands on system.internal and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain; find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

### Two-strike rule for work items
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17
- **what:** insertWorkItem marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops. Cost: items born unresolved accumulate; re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. Centralised `workItemTerminalStatuses` const (work_items_common.go) keeps the dedup index and ON CONFLICT predicates from drifting.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike, deploy table; HANDOFF-pipeline-triage-april-2026.md#patterns
- **relations:** idx_swi_dedup migration 012; discovery noise on dead sites
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

### Discovery auto-triage and scheduled-audit open questions
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15)
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

### Discovery agents on dead/stub sites (noise at scale)
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "80+ needs_content_planning items for gamesdesign.co.uk, mostly [stale: triaged 48h+] … running on a sites row whose adoption had failed or been deleted" (2026-04-23, item 19, added in version (1) only)
- **what:** Discovery agents keep generating remediation items for sites that are deleted, stubs, or mid-adoption. Proposed precondition: skip site_ids with status deleted/archived, no current identity spec, or adoption in flight.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 19
- **relations:** two-strike pile-up; library-row cleanup pattern
- **verify-later:** discovery agent site-selection queries

### CollectedData: single-channel orchestration working memory and its pathologies
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Analysis only. No code changes proposed yet" (2026-05-11); duplication called "structural", observed in every log
- **what:** CollectedData (orchestration_states.collected_data JSONB) is the single channel for step outputs, routing metadata, loop variables and parent-reply context — "the most overloaded data structure in the system". Documented pathologies: recursive `__raw_message__` nesting (write amplification ×15 optimistic-lock retries), dual storage at step_name AND output_field, InitialRequestData/__raw_message__ overlap, six conflated data categories in one flat namespace, loop iteration data stored 3-4×, CleanDataMap stripping legitimately-named response fields. Recommendations R1–R6 (strip system keys from __raw_message__, pick one storage key, namespacing, loop GC, delta writes) proposed, untriaged.
- **sources:** FOCUS_collected_data_analysis.md (whole)
- **relations:** flat-namespace collision risk; compensating mechanisms; consumer-group race (duplicate keys as evidence)
- **verify-later:** BuildCollectedData / storeActionResult in coordinator.go; whether R1 ever landed

### Flat-namespace collision risk and the compensating-mechanism accretion
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** dev-guide-documented incident: "section-editor declared content_data as optional and the nested-source loop silently lifted site_record.content_data … and overwrote a hero section" (referenced 2026-05-11)
- **what:** Because caller inputs, step results and site context share one flat map, actions can silently pick up `site_record.site_id`/`content_data` instead of caller-supplied fields. The framework compensates with UnwrapDeep, FindByPath prefix fallbacks, extractReplyToMetadata 3-tier priority, output_mapping — accreting workarounds faster than it consolidates. New code should use collision-free names (target_site_id convention); existing code left alone.
- **sources:** FOCUS_collected_data_analysis.md#4.4, #5; ASSESSMENT_imagery_phase_0_1…md#Caveat-1
- **relations:** CollectedData analysis; ExtractActionInputs conventions
- **verify-later:** ExtractActionInputs nested-source loop behaviour for undeclared fields

### Response-topic consumer group race (per-pod groups fan out every response)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "Discovery, not yet remediated" (2026-05-10); ~85 consumer groups on system.agent.generic.responses, only 3 live; two pods ran ProcessResponse on the same message 215ms apart
- **what:** The requests topic uses a shared stable consumer group but each chassis pod joins the responses topic under its own per-pod UUID group, so every response is delivered to every pod; each independently advances orchestration state, and the loser of the version race can flip a step to FAILED (observed on call_logo_gen). Mostly silent (idempotent writes) but structurally wrong; the system relies on shared-pool semantics it doesn't have. Open questions: intended model, per-spawn job.* topic groups, CAS hardening in ProcessResponse, 82 stale groups cleanup.
- **sources:** ANALYSIS_chassis_response_consumer_group_race.md (whole)
- **relations:** dispatcher stall Bug 1; duplicate collected_data keys; Phase 2F migration testing blocked
- **verify-later:** AgentClient constructor wiring (consumerGroup argument); ProcessResponse CAS behaviour in coordinator.go

### Kafka empty partition assignment on simultaneous pod join
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "five agent-chassis pods were members of generic-requests-group but all showed #PARTITIONS: 0 … Fix applied: Delete one pod to force rebalance" (2026-04-20)
- **what:** After a deploy where all pods join within the same second, the group can go Stable with the partition unassigned — zero consumption while offsets pile up and work items sit triaged. Workaround: kill a pod. Watch item on every deploy: at least one member must show #PARTITIONS: 1.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#1
- **relations:** consumer-group race; dispatcher reliability
- **verify-later:** whether staggered restarts or a fix was adopted

### Observability gaps: owner_agent_type "generic" and orchestration_name
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "orchestration_states rows where the generic agent routed to a different workflow still show owner_agent_type = 'generic'" (2026-04-20, P3)
- **what:** When the generic chassis routes a message to another agent's workflow (FindBestGroup), the orchestration is filed under owner_agent_type='generic' and orchestration_name doesn't carry the scheduler's sched-<task> name — searches by agent type or task name find nothing, which caused the "trigger never runs" misdiagnosis. Fix shape: selectWorkflow sets owner_agent_type to the resolved type.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7, items 7-8
- **relations:** content-feed-trigger bug; execution_path not populated (flywheel note 2.4c)
- **verify-later:** selectWorkflow in processor.go

### content-feed-trigger workflow shape bug (array vs object count)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "Fix applied … output_format = 'object' ✓ items_field = 'news_sites.rows' ✓ … Pending verification on next fire" (2026-04-20)
- **what:** The scheduled news trigger was "broken for weeks" not because of routing (generic-agent routing works as designed) but because find_news_sites returned a bare array: check_has_sites read `.count` off an array (empty string → default branch), and the loop crashed on nil when no sites existed. Fixed by output_format object + items_field .rows. General lesson: condition fields need the object {rows,count} shape.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7
- **relations:** owner_agent_type observability gap (why it was misdiagnosed)
- **verify-later:** content-feed-trigger definition current shape

### CSS assembly pipeline (composable theme → styles.css)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "fully built path" (2026-05-12); render_css_from_spec_action.go deterministic, verified live schema for css_snippets
- **what:** webdesign-agent: analyze_design (LLM) → render_css_from_spec (deterministic Go: theme composition from palettes/layouts/typography_sets FKs, css_snippets matched via applies_to JSONB overlap against the site's component functions, dark-section variants) → deploy_css git commit to assets/css/styles.css → B2 CDN sync. css-patch-agent is the bypass path for one-off fixes (patches the deployed file directly, not the snippet library).
- **sources:** FOCUS-css_js_mechanisms.md#1; HANDOFF_2026-04-18_design_and_styling…md
- **relations:** composable theme migration 025; site-design-planner
- **verify-later:** render_css_from_spec_action.go, render_css_composition_helpers.go

### JS three-path model (js_content deployed, js_snippets loader missing, inline legacy)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "declared in contracts, table populated, BUT NO LOADER IS WIRED UP" (verified 2026-05-12: 9 js_snippets rows, no reference in head templates or RenderHead)
- **what:** Path A (deployed): per-component JS in content_components.js_content, extracted at store time by separateInlineJS(), deployed as /tools/assets/{function}.js via collectJSAssets() multi-file git commits. Path B (aspirational): js_snippets shared utility table with applies_to scoping — a registry of intentions with no runtime loader; contracts' "loaded via head component" claim is aspirational. Path C (legacy anti-pattern): inline <script> baked into html_template, violating contract 003 — news components still there. Path D: html-assembler's inject_js flag has no visible reader. Interim tactic: insert the snippet row AND duplicate inline until the loader exists.
- **sources:** FOCUS-css_js_mechanisms.md#2, #3, #4; HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#2
- **relations:** component contract 003; JS separation deployment (2026-04-17)
- **verify-later:** RenderHead in component_library.go; js_snippets rows; whether a loader was ever built

### Component quality tracking (quality_score et al.)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "migration_component_quality.sql (applied) … compute_component_quality_action.go (pending Go deploy)" (2026-04-17); quality scoring described as working in the 04-18 handoff status table
- **what:** content_components gains template_variable_count, schema_field_count, template_closed, schema_template_synced, has_data_component, quality_score (100 minus deductions), quality_issues; scored inline on store and by a component-quality-auditor agent; planner prefers high scores, auditor targets low ones for regeneration. Backfill via system.internal work item. 43 pre-existing components had 0 template variables (content baked in) — regeneration targets, not mass-deletable.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#6, #Architecture-decisions
- **relations:** pre-store validation gates; component-creator prompt tiers
- **verify-later:** compute_component_quality registry entry; quality_score population in DB

### Pre-store component validation gates + planning deferrals + empty-section filter
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** deployed 2026-04-17 (three checks before INSERT; sectionHasVisibleContent; empty-schema deferral); root incident: max_tokens=4000 truncation left unclosed <section>, CSS rendered as page text on vonc.com
- **what:** Three layers preventing broken components/sections reaching pages: store-time rejection (template must contain <section>/<div>, balanced <style> tags, non-empty input_schema), plan-time deferral of content-type components with empty schemas, and render-time skipping of sections with <10 chars visible text. Component-creator max_tokens raised 4000→16000 and prompt made context-aware.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #6; HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#1, #6, #7
- **relations:** LLM reliability tracks; quality tracking
- **verify-later:** store_generated_component_action.go validation block; rerender_single_page_action.go sectionHasVisibleContent

### LLM reliability strategy for component generation (observability first, shrink the contract second)
- **category:** llm-quality-testing
- **status-signal:** partial
- **status-evidence:** "Track 1 — Make rejection observable. Done in this iteration" (2026-05-11); tracks 2-3 open
- **what:** LLMs are structurally good but unreliable at exact schema↔template list reconciliation (bookkeeping, not creativity). Strategy: (1) pre-store validator writes structured rejections to agent_error_log — done; (2) move bookkeeping out of the LLM: inject the root section wrapper at store time, declare Tier D sub-schemas centrally in queryresolve, optionally derive schema keys from the template parser; (3) prompt/model tweaks only after patterns are visible. Explicitly rejected: silent auto-correction at the validator; accumulating hand-written components without addressing the prompt.
- **sources:** FOCUS_llm_reliability_for_component_generation.md (whole)
- **relations:** validation gates; Tension #1 (same "don't trust LLM formal labels" family); tiered field classification
- **verify-later:** agent_error_log rejection rows; whether 2a/2b landed

### Tiered component field classification (Tier A voice / B tunable static / C site data)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "revise_component_creator_prompt.sql (applied)" (2026-04-17)
- **what:** Component schemas classify fields: Tier A voice content (source llm, required), Tier B tunable labels (source static, optional, with fallback), Tier C site data (source site_specs.*/site_assets.*). Prevents both "35 required fields" and "0 fields, everything hardcoded". Template/schema sync invariant: every {{.x}} has a schema entry and vice versa. Tier B static fallbacks later become the language problem ("Browse All Tools" on non-English sites) and the "soft static" override idea.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#5; FOCUS_language.md#static-fallbacks
- **relations:** LLM reliability; language surfaces
- **verify-later:** component-creator prompt in agent_definitions

### system.internal site convention
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Created for maintenance/library-level work items … id: eac60db8-…, domain: system.internal" (2026-04-17)
- **what:** A never-deployed sites row (brand_dna.is_system=true) that hosts library-level and maintenance work items not belonging to any customer site (e.g. component_quality_scan backfills). Side effect: its maintenance-pipeline items sit dormant (no maintenance-dispatch-loop) and it absorbs untargeted scheduler dispatches.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#7; FOCUS_dispatch_diagnostic(4).md#Q4; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** pipeline soft label; Bug 3 site targeting
- **verify-later:** sites row; items accumulated on it

### String-value naming convention (identifier-shaped snake, data-shaped kebab)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Status: applied (migration 051, page_canonical.go + page_role_validator.go updated, contracts doc v9, debug guide v2.10)" (2026-05-17)
- **what:** Decision rule for string-typed columns/enums: used as a Go identifier (switch case, registry key, dispatch route) → snake_case (site_work_items.item_type); pure data describing what a thing is → kebab-case (pages.page_type, content_components.function); single word → bare lowercase (statuses). Root incident: normalisePageType wrote snake while all readers expected kebab, silently hiding blog pages. Companion fix: homepage page_type 'index' → 'landing' (name vs type conflation). Snake-input fallback retained as a bounded migration-tail exception; tests document behaviour, not intent.
- **sources:** FOCUS_naming_conventions_kebab_vs_snake.md (whole)
- **relations:** Tension #2 canonicalisers; page_type vocabulary gap
- **verify-later:** migration 051; CHECK constraint on pages.page_type

### Architectural Tension #1 — infer-and-repair vs deterministic structure derivation
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Status 2026-05-25: Tension #1 has a deployed partial fix (Part A — ValidateRoles -index rule), pending a clean production test"
- **what:** The pipeline takes structural decisions (page role/type/URL) from LLM free-text labels then repairs with starved, vertical-hardcoded heuristics, producing silent structural corruption (section hubs flattened to content). Resolution principle: derive structure deterministically from the LLM's reliable signal — naming (`<section>-index` marks a hub, vertical-agnostically); schema-constrain generation to kill form errors (necessary but not sufficient); make fallback heuristics fail loud, never default to content. Explicit recommendation AGAINST a free parent-pointer tree (worst LLM reliability tier); a leaf's section, if needed, is a constrained choice over the enumerated hub set.
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-1; HANDOFF_2026-05-26 (page_type re-type as an instance)
- **relations:** Tension #2; page_type vocabulary gap; LLM reliability strategy (same principle, component scale)
- **verify-later:** ValidateRoles -index rule and de-hardcoded nestedRoleFromURL in page_role_validator.go

### Architectural Tension #2 — page identity derived in multiple places that undo each other
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Tension #2's residual confirmed cosmetic (see HANDOFF_2026-05-25)" but flavour-collapse residual "evidence-gated, not yet a code change" (2026-05-25)
- **what:** Adoption, planner-write and convergence each re-derive canonical page name/role/URL with no single owner, so a later stage can undo an earlier correct result (convergence preserved games-index; WriteSitePlanAction flattened it one step later). Principle: one canonical owner; canonicalisation idempotent on already-canonical input; downstream reads identity read-only. Part A made section indexes round-trip cleanly; the remaining residual is flavour collapse (validator emits generic section-index, losing blog-index/entity-directory flavour) — decide from a deployed run whether the component resolver needs the flavour before writing preservation code. Withdrawn: merging the two role-normalisers (intentionally layered).
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-2; HANDOFF_2026-05-26 (write vs sync canonicaliser divergence)
- **relations:** Tension #1; kebab/snake; canonicaliser divergence
- **verify-later:** CanonicalisePage/normaliseRole/normalisePageType in datahelpers/page_canonical.go; component resolver's page_type dependence

### site_specs vs site_plan two-layer architecture + aspect ownership contract
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "build-site-planner workflow writes both shapes during transition (old site_specs/site_plan aspect AND new plan tables)" (undated FOCUS, references docs 028-030)
- **what:** site_specs = strategic, brand-level, slow-changing, one owning agent per aspect (classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planner owns the four plan tables). site_plan tables = per-build, row-shaped, rebuilt per plan. Three ownership rules (don't read what you didn't spec; don't overwrite another's aspect; write outputs to the spec) with the classifier read-and-extend carve-out. Decision rules and anti-patterns for where new data lives (specs vs directives vs sibling structured tables).
- **sources:** FOCUS_site_spec_vs_site_plan.md (whole); ASSESSMENT_imagery_phase_0_1…md#What-Phase-1-changes
- **relations:** directive cascade; lock transfer; imagery placement
- **verify-later:** site_plans/site_plan_pages/site_plan_sections/site_plan_directives tables; legacy site_plan aspect readers (pageflow-builder)

### site_plan_directives cascade + brief renderer
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Reconciler is documented in doc 030 but the chassis-side implementation has been landing in stages"; brief renderer named as `datahelpers/page_brief.go` "per the work order"
- **what:** Cross-cutting guidance rows located by (scope site/page/section, scope_ref, category, subject) with HITL lock columns. Consumers never read rows directly: a Go brief renderer cascades site → page → section and applies cardinality semantics (single-valued subjects override at narrower scope; multi-valued accumulate), emitting short LLM-ready briefs. The pattern imagery/text/design guidance should all follow.
- **sources:** FOCUS_site_spec_vs_site_plan.md#directives; ASSESSMENT_imagery_phase_0_1…md#Amendments
- **relations:** lock transfer; site_plan_imagery sibling-table pattern
- **verify-later:** datahelpers/page_brief.go existence and consumers

### HITL lock transfer across plan rebuilds
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** described as run "inside write_site_plan" per doc 030; extended for imagery + lock_type/expiry per 2026-05 patches ("transferDirectiveLocks carries lock_type/expiry — written (patch doc)")
- **what:** On plan rebuild, locked directives from the previous current plan are matched to new rows by composite key (scope, scope_ref, category, subject, ordering); locked_at/locked_by and HITL-edited text copy over (HITL wins); unmatched locks drop with a log, previous plan kept as history. Any sibling table wanting HITL adopts the same shape.
- **sources:** FOCUS_site_spec_vs_site_plan.md#Lock-transfer; FOCUS_adoption_faithfulness_via_locks(2).md#dependency-chain
- **relations:** adoption-faithfulness timed locks; site_plan_imagery
- **verify-later:** transferDirectiveLocks in write_site_plan action code

### Build-time design/imagery trigger emission (Gap A)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "deployed step order in v1.0.1047 is read_specs → … → reconcile_site_plan → emit_design → emit_imagery → complete … So Gap A is closed on both fresh-build and adoption paths" (2026-05-26, verified on gamesdesign cascade)
- **what:** `emit_design_items` and `emit_imagery_items` (shared `imageryplan` package) wired as plan-time steps in build-site-planner, closing the long-standing gap where composition and imagery items were never emitted after the Phase-1 refactor moved the terminal step away from WriteBuildItemsAction. Nine needs_imagery items at documented priority bands (65 index-hero, 70 site-logo, 75/80 others, 98 clamped section-scope) observed live.
- **sources:** HANDOFF_2026-05-26_design_imagery_triggers_and_adoption_diagnosis.md#What-deployed
- **relations:** site-design-planner; imagery loop closure; site_plan_imagery
- **verify-later:** build-site-planner workflow steps in agent_definitions; imageryplan package

### site-design-planner composition resolver (composition before render)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "First successful composition run … completed cleanly" (2026-04-20 gamedesign.uk with all five IDs); "needs_composition ran via site-design-planner and install_site_composition populated sites.style_collection_id" (2026-05-26)
- **what:** A handler agent (needs_composition work item) that resolves layout (deterministic tag-overlap against layouts.industry_tags), typography (font-family/character match) and palette (fingerprint → mission → design_intent priority) BEFORE webdesign-agent renders, installing css_themes + style_collections rows transactionally and hard-failing when classification is missing. Fixes the fork+install conflation that produced first-render-with-wrong-layout (two commits, first knowingly wrong). Scope decisions: brave backfill, hard-fail loud logging, adoption and new builds unified, re-resolution deferred to HITL, fork-to-library gated behind two flags.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#3-6; HANDOFF_2026-04-20_composition_deployed…md; FOCUS_navigation_HANDOFF_navigation_fix.md#Architectural-Gap (origin: navigation/layout spec idea)
- **relations:** composable theme migration; navigation/layout specs; classification tags mismatch
- **verify-later:** site-design-planner agent definition; resolve_composition_*.go, install_site_composition.go

### Composable theme migration 025 (palette + layout + typography decomposition)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phases 1-3 (data model, layouts, seeding) are deployed and verified. Phases 4-5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified" (2026-04-18); renderer subsequently exercised in later cascades
- **what:** Themes decomposed into palettes, layouts (15 seeded CSS templates each passing a 7-point contract audit), typography_sets (6 seeded), FK-linked from css_themes/style_collections; renderer cutover to a single JOIN loader + FuncMap (palette/typo/token) with hard error on NULL FKs; fork action resolves the three pieces before insert.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#2
- **relations:** CSS assembly pipeline; site-design-planner
- **verify-later:** layouts/palettes/typography_sets row counts; render_css_composition_loader.go

### webdesign-agent post-merge loop bug and generate_css stuck mystery
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "This is a loop bug in my migration … Fix proposal (NOT YET APPLIED)"; "Even with the loop fixed, we STILL don't know why generate_css didn't execute" (2026-04-20); a later cascade (04-23) "proceeded through generate_css and deploy_css to check_should_fork" suggesting recovery
- **what:** The 010 migration left every path out of deploy_css looping back to generate_css (update_site.next_step and check_update_db.else_step should point at check_should_fork); separately, one run sat at generate_css (deterministic action) producing no log line, no heartbeat, evidence lost to pod rotation. Instrumentation runbook written for reproduction.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#A
- **relations:** silent completion; consumer-group race (candidate explanation)
- **verify-later:** current webdesign-agent next_step wiring; whether the loop fix SQL was applied

### Composition classification-tags mismatch (industry_tags empty)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "composition_layout.reason: 'fallback — no classification tags' … layout resolver fell back to brochure-formal for a site that clearly wants something dashboard/application-like" (2026-04-20); migration 008 (dynamic taxonomy, industry_tags array from classifier) validated 2026-04-23 likely addresses it
- **what:** The layout resolver read a nonexistent tags array while classification stored industry/sub_industry strings, so every site fell back to the generic layout and style_collections.industry_tags was written empty, breaking future library matching. Migration 008 made the classifier emit an industry_tags array against a dynamic taxonomy read from the layouts table (read_layout_taxonomy action), validated end-to-end with tool-portal-dark selected via library_match.
- **sources:** HANDOFF_2026-04-20_composition_deployed…md#B; HANDOFF_2026-04-23(1).md#deployed, #validated
- **relations:** site-design-planner; dynamic taxonomy classifier
- **verify-later:** readClassificationFromContext in resolve_composition_helpers.go; classifier output shape post-008

### Planner palette prose vs structured reference_values (Gap C)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "OPEN" (2026-05-26): planner emits colour decision as design_intent.colour_mood prose; composition palette cascade misses the design_intent slot and falls to layout-seed default
- **what:** Planned colours reach the render only via the webdesign-agent overlay, not the base composition. Fix options: planner emits a structured palette.reference_values block (primary/secondary/accent/background/surface/text/text_muted/border) or site-design-planner consumes colour_mood directly.
- **sources:** HANDOFF_2026-05-26…md#gaps, #Where-to-resume
- **relations:** palette cascade; site-design-planner
- **verify-later:** plan_site output schema; palette resolver slots

### Imagery loop closure plan (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Decisions table (new imagery-quality-auditor agent; max 2 regen attempts; asset locking mirrors page_components; per-section granularity deferred); later docs show Phase 2G/2H verified 2026-05-15 and asset lock columns landed via migration 053
- **what:** The sequenced plan closing the gap between imagery asked for (specs/plans) and imagery delivered: Phase 0 wire imagery_direction into prompts + populate origin_model; Phase 1 algorithmic discovery checks (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404) routed to the existing image-build-handler; Phase 2 assets locking + asset_key; Phase 3 adoption image mirror; Phase 4 text-only visual-auditor imagery awareness; Phase 5 vision-capable LLM path; Phase 6 imagery-quality-auditor. Explicit non-goals: per-section imagery_plan, icon resolver, infographic generator, provider router, img2img.
- **sources:** PLAN_imagery_loop_closure.md (whole); ASSESSMENT_imagery_phase_0_1…md
- **relations:** dispatch diagnostic (Phase 2G verification); adoption faithfulness locks; site_plan_imagery
- **verify-later:** which discovery checks exist under discovery_checks/; assets.locked_at/lock_type/asset_key columns

### imagery-quality-auditor agent
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 6 of the plan; not reported deployed anywhere in scope (plan pre-dates 2026-05)
- **what:** A vision-capable sibling of visual-design-auditor: loads assets + imagery_direction (lock-honouring, excluding human uploads), runs a vision LLM audit with imagery-specific categories (direction_mismatch, brand_mismatch, inconsistency, quality, inappropriate), writes findings with max_fix_attempts 2 routing back to image-build-handler; counts toward the existing 3-pass audit cap; gated rollout via design-audit-agent.
- **sources:** PLAN_imagery_loop_closure.md#Phase-6
- **relations:** vision-capable LLM path; asset locking; design-audit-agent
- **verify-later:** agent_definitions for imagery-quality-auditor

### Asset locking mirrors page_components (+ asset_key multi-image readiness)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** planned in Phase 2; FOCUS_adoption_faithfulness (2026-05-19) states 053 migration adds lock_type + lock_expires_at "on page_components, site_components, site_plan_directives, assets … written, ready to apply"
- **what:** assets gains locked_at/lock_type (same vocabulary and exclusion predicate as page_components) so audits/discovery skip locked assets; asset_key column (default = purpose, unique per site) opens multi-image purposes (adoption mirror writes adopted:<filename>) without breaking existing single-purpose upserts; old purpose-unique index dropped only after asset_key bedding-in.
- **sources:** PLAN_imagery_loop_closure.md#Phase-2; FOCUS_adoption_faithfulness_via_locks(2).md#implementation-plan
- **relations:** lock policy table; adoption image mirror; per-page hero resolver (assets.asset_key join)
- **verify-later:** assets table columns and indexes

### Imagery algorithmic discovery checks
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Phase 1 planned three checks; dispatch doc shows `unfulfilled_imagery_plan` check (Phase 2G.4) live and emitting 8 items on robot-hands (2026-05-14) — with the pipeline='design' emission bug
- **what:** No-LLM checks catching spec-to-delivery gaps: image prompt in plan but no asset; hardcoded fallback path in rendered_html with no asset (the silent-failure case); referenced image URL with no matching asset. Emit needs_imagery/needs_hero_image/needs_logo items to image-build-handler via the dispatch loop.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1; FOCUS_dispatch_diagnostic(4).md#why-stuck, #Q4
- **relations:** pipeline soft label bug; baked-fallback problem
- **verify-later:** discovery_checks/check_unfulfilled_image_prompt.go etc.; unfulfilled_imagery_plan pipeline value

### imagery_direction into image prompts + origin_model provenance (Phase 0/0.1)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "The Phase 0.1 deliverables stand" pending one verification query (assessment, undated ~2026-05); compatible with Phase 1 architecture which stabilises design_intent ownership
- **what:** Image generation reads site_specs design_intent.imagery_direction and prepends "Style direction: <direction>\n\nSubject: <prompt>" to the three-tier prompt; store_asset writes origin_model for provenance. The strategic-only read survives Phase 1 (per-page directives become the successor once site_plan_directives lands). Side benefit: pulls planner-invented hero prompts back toward the adopted look (partial mitigation of Bug 4).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0; ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md (whole)
- **relations:** planner ignores site_archetype imagery; image parameter-shaping
- **verify-later:** generate_image_actions.go composeImagePromptWithDirection; assets.origin_model population

### Adoption image mirror
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3 of the plan; "Adopted images are persisted but only as historical record for now" — no deployment claim in scope
- **what:** Stop discarding crawled imagery: mirror_adoption_images action downloads source images (capped count/size), uploads to S3, inserts assets rows (origin_type=adopted, asset_key=adopted:<filename>); wired into apply_adoption_plan plus a backfill discovery check and a one-step adoption-image-mirror agent. Future hook for img2img reference generation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3
- **relations:** asset_key; image parameter shaping (reference_image_uri)
- **verify-later:** mirror_adoption_images_action.go existence; assets rows with origin_type='adopted'

### Vision-capable LLM path
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 5 of the plan; no deployment claim in scope
- **what:** Extend aiservice with GenerateTextWithImages (Anthropic image content blocks, URL source), preferring extension of execute_llm_prompt with an image_urls_field config over a new action; presigned-URL freshness required; vision_call tagged in llm_call_log for cost tracking.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** imagery-quality-auditor (consumer); llm_call_log
- **verify-later:** platform/aiservice/anthropic.go vision support

### Image generation as parameter shaping (not prompt blending)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "The composer step is its own work" — recommendation recorded during Phase 2G step 5 design (undated, ~2026-05)
- **what:** Unlike text, images have a 77-token CLIP budget and no "don't" understanding — composition means deriving parameters (subject, negative_prompt from kind, style_preset/LoRA from imagery_direction, reference_image_uri from adopted images, aspect/cfg/steps per kind), not blending prose. A cheap compose_image_request step (Go rules or small LLM) producing a parameter envelope before image-generator is the candidate design; belongs with Phase 2H request-shape work.
- **sources:** FOCUS_prompt_composition_pattern.md#What-this-means-for-images
- **relations:** mega-prompt fragility (envelope pattern B); Phase 2H
- **verify-later:** image-generator request shape; any compose_image_request action

### site_plan_imagery sibling table
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** live JOIN in production code by 2026-06-02 ("site_plan_imagery.key = assets.asset_key"); write_site_plan step description updated to include imagery HITL-lock transfer (2026-05-26)
- **what:** Structured per-image plan rows (kind, key/asset_key, prompt, style hints, scope/scope_ref) mirroring site_plan_directives' scope+locking pattern — the successor to the legacy site_specs.site_plan.image_prompts dictionary; scoped page rows drive per-page heroes.
- **sources:** FOCUS_site_spec_vs_site_plan.md#where-imagery-lives; HANDOFF_2026-06-02…md#fix
- **relations:** per-page hero resolver; lock transfer; directive cascade
- **verify-later:** site_plan_imagery schema; emit_imagery_items writes

### Per-page hero resolver + rebuild-after-asset (baked-fallback fix)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "the fixes are in production" (2026-06-02) — plan_sections page-aware ensureAssets; flag_page_image_rebuild wired as image-build-handler terminal step (UPDATE 1 verified); registry entries "still required" at doc date
- **what:** Root cause triad: site-wide hero_url overwritten last-write-wins; async imagery completing after first render baked the on_missing fallback (/assets/images/hero.jpg) into rendered_html; terminal rerender reassembled stored HTML without re-planning. Fix: ensureAssets resolves this page's hero via site_plan_imagery JOIN assets (page scope) and site logo from site scope; flag_page_image_rebuild flags the page needs_rebuild and emits needs_page at priority 99 (dedup key page_rerender:<page>) so it re-resolves through plan_sections after its asset lands. Logo/header path deliberately out of scope (render_site_components, not plan_sections).
- **sources:** HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md#1
- **relations:** imagery checks; section-data reconciler (same handoff); two image-resolution paths open follow-up
- **verify-later:** registry.go entries for flag_page_image_rebuild/reconcile_section_data; hero component input_schema (field vs hardcoded template — open question)

### Planner ignores site_archetype imagery constraints (Bug 4)
- **category:** imagery
- **status-signal:** unknown
- **status-evidence:** "site_archetype.design.imagery … says 'minimal icons/diagrams, no decorative photography'. The planner's site_plan still produced lavish hero prompts" (2026-04-23)
- **what:** The planner invents hero image prompts contradicting the adopted archetype's imagery stance. Fix shape: planner prompt reads site_archetype.design.imagery and sets needs_images=false when it says none/minimal. Phase 0.1's style-direction prepend partially mitigates the symptom.
- **sources:** HANDOFF_2026-04-23(1).md Bug 4; ASSESSMENT_imagery_phase_0_1…md#Bug-4
- **relations:** imagery_direction prompt composition; adoption faithfulness
- **verify-later:** plan_site prompt for archetype imagery constraint

### Mega-prompt fragility and candidate replacement patterns
- **category:** NEW:prompt-composition
- **status-signal:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS, ~2026-05)
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas; six fragility concerns (untraceable failures, monotonic rule growth, coupled component vocabulary, one blend ratio, model coupling, token waste ~160MB/build-cycle at scale). Five candidate patterns: per-component templates; structured intermediate envelope (cheap-model stage 1 → focused stage 2, cacheable, lockable); tool-calling for schema; validation-instead-of-prompt-rules; hybrid baseline+overrides. Envelope (B) flagged strongest for both text and images.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (pattern D partially exists); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today

### Design fingerprint extraction pattern (adoption parse stage for design)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** described as "the proven template" (interactive FOCUS, sequencing locked 2026-05-14); exercised live in 04-23 cascade (crawl → fingerprint → fetch CSS → enrich → analyze)
- **what:** Firecrawl rawHtml parsed Go-side (goquery) by extract_design_fingerprint for colours/fonts/CSS vars/layout/dark sections; external CSS fetched via firecrawl_scrape and merged (EnrichFingerprintWithCSSAction); an LLM step (generate_design_intent) produces the semantic brief; stored as design_reference (concrete) + design_intent (semantic) spec aspects. The template any other parse-stage extractor copies.
- **sources:** FOCUS_interactive_content_generation(4).md#Adoption; HANDOFF_2026-04-23(1).md#validated
- **relations:** interactive fingerprint (clone of this pattern); site-design-planner palette source
- **verify-later:** extract_design_fingerprint action; design_reference/design_intent aspects on adopted sites

### Interactive fingerprint parse stage (C1–C6)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** "C1 — extract_interactive_fingerprint (new action). Status: in progress, 2026-05-15"; C2–C6 planned
- **what:** New Go extractor over crawled rawHtml capturing canvas elements, inline/external scripts, event handlers, forms, library signals (rAF, canvas contexts, jQuery/Three/Phaser/React/Vue) and a per-page type_hint heuristic (calculator/game_or_animation/interactive_widget/static); then external-JS fetch loop (C3), enrich (C4), LLM interactive_intent brief with feasibility markers (C5), written to new interactive_reference/interactive_intent spec aspects (C6). Deliberately a new file, not an extension of the design extractor; AST parsing out of scope.
- **sources:** FOCUS_interactive_content_generation(4).md#Path-C
- **relations:** design fingerprint pattern; capability markers; Firecrawl executeJavascript escalation
- **verify-later:** extract_interactive_fingerprint_action.go existence; interactive_reference aspects in site_specs

### Four-stage interactive-content pattern (parse / assess / generate / integrate)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Not a roadmap — a map of the territory" (family v4, updated through 2026-05-15); tools implement it "mostly", minus the parse stage
- **what:** The reference shape for handling any interactive content type encountered on adopted sites: parse the source, assess producibility (producible_now / producible_simpler / blocked per the 028 spec model with feasibility-recheck promotion), generate the artefact, integrate into the build pipeline. Agreed sequencing: Path C (parse stage) → Path D (tool reliability — tools "currently don't work") → Path A (games) → B (news publishing) / E (numbered-component cleanup).
- **sources:** FOCUS_interactive_content_generation(4).md#four-stage, #Sequencing
- **relations:** tools pipeline; games gap; news publishing gap; capability assessment
- **verify-later:** feasibility-recheck task existence; state of tool reliability work

### Tools pipeline (suggest / deploy-fork / generate / improve / audit)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "Five agents, two discovery checks, full lifecycle documented … Tier 3 (headless browser visual testing) is planned, not built"; but Path D flags interactive behaviour "reportedly currently don't work" (2026-05-14)
- **what:** tool-suggester (LLM over spec aspects + library, 0-5 suggestions with library_source routing), tool-deployer (library fork with forked_from + tool page + companion guide), tool-generator (novel LLM tool, same wiring), tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing. Missing vs the four-stage pattern: no parse stage (source tools not read), loose source-tool fidelity. Fork-retry idempotency fixed (P2: reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2
- **relations:** games gap (copies this shape); library model; quality model
- **verify-later:** actual tool interactivity failures (Path D); tool_health/tool-auditor definitions

### Games as a content type (largest pipeline gap)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Games — nothing yet … page_type='game' doesn't exist in the classifier vocabulary" (2026-05); the vocabulary absence later CAUSED the 05-26 duplication bug
- **what:** No game-suggester/generator/improver/auditor, no game template library, no game_health check, no spec aspect; game-list components force fabrication. Plan: copy the tools pattern wholesale; add `game` to the page_type vocabulary (Option 1 hardcode now, Option 4 page_types table later — canonicalise kebab/snake first). The missing `game` type is not cosmetic: the planner re-typed adopted game pages to `tool`, driving rename + duplication.
- **sources:** FOCUS_interactive_content_generation(4).md#Games, #classification-vocabulary; HANDOFF_2026-05-26…md#diagnosis
- **relations:** page_type vocabulary gap; four-stage pattern; library model
- **verify-later:** plan_site Canonical Page Types list today

### News publishing gap (curation → deployed posts)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "News — pipeline exists, publishing doesn't … The pipeline ends at curation" (2026-05)
- **what:** Ingestion/triage/diversity produce latest-news.json per site but nothing turns curated items into deployed blog posts; Path B connects news ingestion to page deployment via page-content-writer with a news-feed input, passing the site's deployed tool list for cross-linking.
- **sources:** FOCUS_interactive_content_generation(4).md#News, #Path-B
- **relations:** feed triage fixes; topic splitting
- **verify-later:** whether an article-publishing step now exists in news pipeline

### Feed triage scoring repair (config reads + wrapper unwrap)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "Triage is working. As of last check: 41 relevant, 23 rejected, 232 ingested (backlog clearing at 15 items per cycle)" (2026-04-17)
- **what:** 200+ items unscored since April 2nd due to three stacked bugs: LLM output truncation (max_items 50 → 15; max_tokens → 8192), config literal invisible to inputs.GetInt (use GetIntField on StepConfig.Config), and the execute_llm_prompt wrapper map ({type,result}) never unwrapped. Topic splitting of the single Grok source into topic-focused sources planned (SQL-only).
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#1, #4
- **relations:** chassis input conventions; content-feed-trigger workflow bug
- **verify-later:** feed_triage_actions.go; content_feed_items backlog state

### Two nav systems and the GetNavItems fallback
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Two Nav Systems (and they conflict)" — nav tables intended, pages flags legacy; partial population yields a mix (undated FOCUS, compiled ~2026-04/05)
- **what:** site_nav_groups/site_nav_items (populated by populate_nav_tables, read by GetNavItems) versus pages.in_header/in_footer legacy flags (GetHeaderNavFromPages fallback). GetNavItems tries tables first, falls back to pages — partial population mixes the two. Nav authority tiers designed (Tier 1 planner rebuild — only tier implemented; Tier 2 autonomous nav agent; Tier 3 drift detection). Nav state captured in snapshots and restorable via revert.
- **sources:** FOCUS_navigation.md#1, #2, #7
- **relations:** stale pages problem; nav discovery checks; site-design-planner navigation spec
- **verify-later:** GetNavItems fallback logic; whether Tier 2/3 exist

### Stale pages from previous builds polluting nav
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "SyncPagesToDBAction uses ON CONFLICT (site_id, name) — it only overwrites matching page names" with fixes listed as "needed" (FOCUS); still item 15 in the errors-to-fix list
- **what:** Pages from prior builds keep in_header=true/status=active and appear in nav though absent from the current plan. Fix design: build_status='deployed' filters on the pages-table nav readers; SyncPagesToDB deactivates stale pages gated by a deactivate_stale_pages flag (new builds deactivate; maintenance/adopt flows preserve).
- **sources:** FOCUS_navigation.md#stale-pages; FOCUS_navigation_errors_to_be_fixed.md#15
- **relations:** two nav systems; adoption faithfulness (preserve semantics)
- **verify-later:** SyncPagesToDBAction current behaviour

### Nav discovery checks and fix agents
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** check/handler tables in FOCUS_navigation (broken_nav_links→nav-link-fixer; checkNavLayout/checkUnwantedElements→component-template-fixer; checkUnlinkedSiteComponents→site-component-linker; orphan_pages→rerender-pages/content-gap-planner)
- **what:** The nav slice of the improvement loop: quality/design/completeness discovery agents detect anchor-slug links, stacked nav (missing flex), unwanted search icons, unlinked header/footer components, orphan pages, missing logo img; dedicated fixers repair templates, relink components (clearing rendered_html + needs_rerender), and make orphans reachable. component-template-fixer's idempotency was case-sensitive, injecting responsive CSS 4× (fix: lowercase compare).
- **sources:** FOCUS_navigation.md#3, #4; FOCUS_navigation_HANDOFF_navigation_fix.md#problems-10
- **relations:** fallback header; duplicate header/footer
- **verify-later:** discovery agent checks arrays; fixInjectResponsiveCSS case fix

### Duplicate header/footer pathology (site-level components in pages.sections)
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** Data fixes applied 2026-04-11 (12 pages.sections rows cleaned, 24 page_components deleted); but 10 dirty rows reappeared by 04-13/14 — "plan_sections filter NOT deployed" (2026-04-20 investigation 7)
- **what:** pages.sections listed site-level component names alongside content sections; rebuilds rendered header/footer as page_components, then InjectHeader/InjectFooter added a second copy. Code fixes designed but pending at doc date: filterSiteLevelSections in PlanSectionsAction (prevents recurrence), skip-if-present guards in InjectHeader/InjectFooter. A discovery check for duplicate headers inside <main> also missing.
- **sources:** FOCUS_navigation_HANDOFF_navigation_fix.md (whole); HANDOFF_2026-04-20_error_investigations.md#7; FOCUS_navigation_errors_to_be_fixed.md#1-2
- **relations:** nav fix agents; page-build-handler
- **verify-later:** plan_sections_action.go for filterSiteLevelSections; component_library.go inject guards

### Nav quality mechanisms of 2026-04-17 (tiers, child-page exclusion, label trust, quick links)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "What Was Deployed This Session" (2026-04-17): tiered priority, isChildPageURL, navLabelForPage, quick_links_html + footer template SQL
- **what:** populate_nav_tables gained a three-tier page priority (core / hubs+conversion / secondary, overflow to utility) replacing arbitrary nav_order truncation; child-page URL prefixes (/tools/, /blog/ …) excluded from all nav groups; nav labels trust page.NavLabel ≤30 chars and rendering no longer truncates to two words; footer Quick Links built from primary+utility groups via a new quick_links_html variable.
- **sources:** HANDOFF_2026-04-17_nav_empty_sections_footer(1).md#2-5
- **relations:** two nav systems; tool nav integration
- **verify-later:** populate_nav_tables_action.go navPriorityTier/isChildPageURL

### Hardcoded fallback nav/header defaults inventing structure
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** "Defect (lines 310–318 of multipage_actions.go): … injects a hardcoded fallback nav — Home/About/Services/Contact" (2026-06-09); RenderFallbackHeader stacked-nav/search-icon behaviour in FOCUS_navigation
- **what:** Two brochure-default fallbacks fabricate structure when resolution fails: RenderFallbackHeader (generic header, stacked nav, unwanted search icon) and AssembleMultipageSiteAction's hardcoded 4-item nav — the primary source of phantom /services.html links. Resolution direction: fallbacks must derive from real pages (buildNavigationFromPages) or fail loud, never invent URLs.
- **sources:** FOCUS_internal_linking.md#2; FOCUS_navigation.md#header-footer-rendering
- **relations:** phantom-link validation; Tension #1 (silent confident fallbacks)
- **verify-later:** multipage_actions.go lines ~310-318; RenderFallbackHeader callers

### Tool nav integration
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** "Known bug (fixed): addToolToNav used wrong column names … failed silently"; remaining: tools listed individually in primary nav, labels too long (errors-to-fix items 3-5, 18)
- **what:** create_tool_component adds a page, page_component and nav entry per tool; column-name bug fixed, but grouping strategy (single "Tools" entry vs individual items) and label shortening remain open design work — feeding the site-design-planner navigation.tools_strategy spec.
- **sources:** FOCUS_navigation.md#5; FOCUS_navigation_errors_to_be_fixed.md#3-5
- **relations:** site-design-planner navigation spec; tools pipeline
- **verify-later:** addToolToNav; nav grouping of tool entries on live sites

### Internal linking machinery and its defects
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** "current as of 2026-06-09. Grounded in multipage_actions.go, site_db_actions.go, queryresolve…"; defects: hardcoded fallback nav, unpopulated *_index_url specs, phantom /services.html
- **what:** The pages table (via upsertPage slug/url/nav_label) is the authority for link targets; nav built from real pages or DB nav structure; fixAnchorLinks bridges single-page anchors to multipage URLs; queryresolve fills list-hub cards; "Browse All X" buttons read *_index_url site_specs (inconsistent sources, often empty → href=""); ExtractAndSyncLinksAction maintains a per-page link_registry — the natural substrate for a phantom-link discovery check that does not yet exist. Hero CTA destinations are the linking half of the site-wide CTA defect; whether the CTA href is a resolvable field or hardcoded template is the gating open question.
- **sources:** FOCUS_internal_linking.md (whole)
- **relations:** hardcoded fallbacks; content quality catalogue; section-data reconciler
- **verify-later:** syncLinksToDB (records vs validates); link_registry schema; hero component input_schema

### Content quality defect catalogue (gamesdesign) and work order
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** "current as of 2026-06-09. Source of record: CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"
- **what:** Five live defect classes on built pages: hero CTA text↔destination mismatch site-wide (lead item, spans content+linking); guide copy tool-flavoured; brand suffix leaking into card titles; empty footer brand/contact; empty tool descriptions. Work order: settle CTA field-vs-template, reuse component-template-fixer's CTA handling; then footer/titles/descriptions batch; then guide re-flavouring. Routing reality check flagged: the three-way finding classification and specialist agents are PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (whole)
- **relations:** recommendation specialist architecture; internal linking; validate_page_content
- **verify-later:** whether identity-advisor/component-template-fixer CTA handling/sites.approval_mode exist

### validate_page_content gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review. (This is Mode 2 of the silent-completion work, already confirmed fixed.)" (2026-06-09)
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination) that any content fix must pass; failures now route consistently to human review. Known false-positive class: adopted content referencing the source domain trips the contamination heuristic (Bug 7 — needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged.
- **sources:** FOCUS_content_quality.md#machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue
- **relations:** silent completion mode 2; phantom-link check hook candidate
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist

### Recommendation specialist architecture (bug vs gap vs recommendation)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** "P10 … Deferred until HITL queue becomes a bottleneck" (April 2026); P9 (gap → needs_content_page) deployed
- **what:** LLM auditors mix factual bugs with opinions; the pipeline shouldn't auto-fix opinions. Proposed: finding_type classification (bug → auto-fix; gap → rebuild via needs_content_page — this part deployed as P9; recommendation → specialist agent decides apply/dismiss/escalate, e.g. identity-advisor for contact details), with per-site approval_mode (auto|review) gating.
- **sources:** HANDOFF-pipeline-triage-april-2026.md#P9, #P10; FOCUS_content_quality.md#machinery
- **relations:** content quality work order; two sources of truth for email
- **verify-later:** write_audit_findings_action.go Rule 4; finding_type field existence

### April 2026 pipeline triage fix set (P1–P5)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Triaged 57 build pipeline failures … wrote and deployed 7 code fixes (P1–P5, P9 across 12 files)" (April 2026)
- **what:** P1: component_id plumbed through create_work_item into site_work_items (unblocking tool-improver's load_tool). P2: idempotent tool fork deploy (reuse orphaned forks). P3: 429/rate-limit/billing errors classified transient in isAIUnavailable (items back to triaged without burning attempts; ~130 wasted attempts/day stopped). P4: load_page_record falls back page_name → page_id. P5: plan-then-reconcile for needs_section_data (auto-close stale requests when data arrives; create without duplicating when still missing) — "feedback loops need both directions".
- **sources:** HANDOFF-pipeline-triage-april-2026.md (whole)
- **relations:** section-data reconciler (P5's successor); two-strike rule
- **verify-later:** ai_errors.go isAIUnavailable patterns; plan_sections loadOpenSectionDataRequests/closeResolvedDataRequest

### Section-data deferral + reconciler loop
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "reconcile_section_data_action.go — new, not yet wired to a host"; pages_under_section implemented (2026-06-02)
- **what:** query.*-sourced section fields unresolvable at plan time defer as needs_section_data; the queryresolve package (pages_where_type, now pages_under_section joining site_areas) resolves them; a lightweight reconciler (not an LLM agent — the once-planned directory-builder was never built) rescans open items whose missing fields are all query-sourced and emits needs_page re-renders (dedup key page_rerender:<page>), leaving human-data items (team, pricing) in HITL. plan_sections closes items on re-render. Host (loop check or post-build finalize) still to pick.
- **sources:** HANDOFF_2026-06-02…md#2; FOCUS_internal_linking.md#4; HANDOFF-pipeline-triage-april-2026.md P5
- **relations:** P5 plan-then-reconcile; list hubs; self-contained components heuristic gap
- **verify-later:** reconcile_section_data host + registry entry; queryresolve switch cases

### Component linking enrichment saga (component_id NULL on rebuilt pages)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Resolved on homepage. 7/7 linked" (2026-04-20) after fixes deployed 04-17/18; slot_name normalized differentiators-section → differentiators proving the data-component branch fires
- **what:** page_components.component_id was wiped on every rebuild because sections_metadata from the content writer carries only rendered_html — extractSectionsFromMetadata defaulted ComponentName to "section" and the enrichment guard skipped every row. Fixes: run enrichSectionsWithPlannedNames before enrichSectionsWithComponentIDs; prefer the HTML data-component attribute over metadata names; strip -section/-container/-wrapper/-block suffixes; log at Info with candidates_tried. Long-term structural fix (deferred): compile_page_sections should emit component_name per metadata entry. Stale pre-fix rows self-heal on next natural rebuild. Companion facts: `mode` has exactly one legal value "recreate"; build_mode is a dead parameter.
- **sources:** HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md; HANDOFF_2026-04-19…md#1; HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md
- **relations:** data-component attribute contract (news-listing template fix); spec-is-primary-input contract
- **verify-later:** save_page_sections_action.go enrichment order; compile_page_sections metadata shape

### Spec-is-primary-input contract for handler workflows
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Spec is primary input (contract rule) — all handler workflow configs must use input_data.spec.* paths" (architecture decision, 2026-04-17); root cause of gauntlet pages getting 0 components
- **what:** Dispatch only reliably populates input_data.spec.*; top-level flattened paths (input_data.page_name) depend on optional `?` input_mapping and silently resolve nil. Handler configs use spec paths; Go actions keep a defensive fallback chain.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#4, #Architecture-decisions
- **relations:** chassis input conventions; flat-namespace collisions
- **verify-later:** contracts doc handler-agent section

### Adoption faithfulness via 90-day timed locks
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending" (2026-05-19); convergence layer marked [done]
- **what:** Adopted sites stay faithful to source for 90 days then develop normally — enforced as timed locks, not a permanent flag. Deliberately timed despite being user-initiated (a faithful starting point, not a frozen final value — documented so nobody "fixes" it to permanent). Because site_plan_directives are plan-scoped and adoption writes no plan, the lock originates at the FIRST write_site_plan (no-current-plan + pages-exist uniquely identifies adopted first plans): page-scoped preserve directives locked adoption/timed/90d; convergence (ValidateSitePlanAction) preserves whatever the 054 query flags adoption_locked; transferDirectiveLocks carries expiry across re-plans; after expiry everything is a no-op. Coexists with 30-day deploy locks at component scope (different questions, no contention).
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md (whole)
- **relations:** lock policy table; lock transfer; FOCUS_planner_ignores_adopted_state (the duplication this protects against)
- **verify-later:** 053/054 applied; write_site_plan first-plan lock branch; v3_site_actions.go convergence

### Lock policy table and the improvable-row predicate
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Approved policy table (with adoption added)" (2026-05-19); filter sweep of "11 locked_at IS NULL callsites" still pending
- **what:** Canonical lock semantics: human-set locks (admin/manual/checkpoint) permanent; auto-locks timed (deploy +30d on page_components; auditors +90d; adoption +90d on plan directives); audit_pending is not a lock. The improvable predicate — `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())` — must replace the 11 bare locked_at checks; CheckComponentLock to gain LockType/LockExpiresAt; expired review locks become needs_lock_review HITL items. Coherence rule: all four Pattern-A tables migrate in one migration, no partial state.
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md#policy, #predicate, #implementation-plan
- **relations:** adoption faithfulness; asset locking; Tension #3 candidate (lock-model coherence debt)
- **verify-later:** the 11 callsites; check_component_lock.go; expired_review_locks check existence

### Adoption source/destination separation and variant axis
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** FUTURE doc (2026-04-20) "Status: Future work … Option 1 sketched"; but 2026-04-23 handoff triggers "the standard adopt-separated kcat command SOURCE_URL=… DEST_DOMAIN=…" — Option 1 evidently landed
- **what:** Decouple the crawled target from the built destination (target_url + destination_domain inputs; ensure_site_record override) and gate spec-writing on an adoption_variant: reference (design only), structure (+archetype/pages), clone (+content_direction — old behaviour), analysis (aggregate competitor_landscape). Phase 2: sites.source_site_id provenance; Phase 3: adoption_references library. Risks: variant-gated data bleed, typo'd destination domains creating junk sites.
- **sources:** FUTURE_adoption_source_destination_separation.md (whole); HANDOFF_2026-04-23(1).md#priority-4
- **relations:** adoption faithfulness (fidelity axis); duplicate-sites-row question
- **verify-later:** apply_adoption_plan variant gating; extractDestinationDomain in ensure_site_record_action.go

### Adoption → classifier handoff (classifier as strategic brain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** apply_adoption_plan rewrite "NOT YET DEPLOYED" 2026-04-23, but 2026-05-26 verifies "Adoption does not bypass the planner — it routes through it via the strategy→briefing→site_plan chain, as 007_adoption_pipeline_v4.md intended"
- **what:** Adoption stops queueing needs_composition/needs_design directly; it writes its specs and emits a single needs_domain_research item so the classifier (with dynamic taxonomy) then strategist → briefing → planner run for adopted sites exactly as for fresh builds — doc 028's ownership model applied to adoption.
- **sources:** HANDOFF_2026-04-23(1).md#not-deployed; HANDOFF_2026-05-26…md#verified
- **relations:** dynamic taxonomy classifier; spec ownership contract; pipeline cascade
- **verify-later:** apply_adoption_plan_action.go current emissions

### page_type vocabulary gap forcing game→tool re-type (Gap B)
- **category:** site-plan-and-reconciler
- **status-signal:** unknown
- **status-evidence:** "root cause is confirmed from the planner's response_text … there is no `game` [in the Canonical Page Types list], so every adopted game is forced to `tool`" (2026-05-26); "OPEN structurally; may have been addressed by the other-chat fixes … Verify post-deploy"
- **what:** The plan_site prompt's closed page-type list lacks `game`; the LLM keeps names faithfully but re-types game pages as tool; canonicalisation's tool branch then renames, and a page_type change (not a name change) is what duplicates pages — 5 duplicate game-*/tool-game-* pairs on gamesdesign. Also exposed: WriteSitePlanAction and sync_pages_to_db canonicalise the same tool-typed page differently (tool-auto-battler vs tool-game-auto-battler) — code read required before fixing. Verification queries recorded (stem-grouped pages; response_text page_type; composition install).
- **sources:** HANDOFF_2026-05-26…md#diagnosis, #Where-to-resume
- **relations:** Tension #1/#2; games content type; adoption faithfulness locks
- **verify-later:** run the three handoff queries on a post-2026-05-26 adoption; page_canonical.go call sites

### Library-row cleanup pattern for failed cascades
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** executed 2026-04-23 with counts (4 css_themes, 7 palettes, 4 style_collections cleared for gamesdesign) and a NOT IN guard protecting seeded layouts
- **what:** Bad cascades leave one set of library rows (css_themes/palettes/style_collections/typography_sets) per resolve attempt; if left, the matcher can pick wrong-decision artefacts for future sites. Reverse-FK-order delete by source_domain is the recovery pattern. Related open item: site deletion should clean up unreferenced library rows (FKs are SET NULL, leaving orphans).
- **sources:** HANDOFF_2026-04-23(1).md#cleanup, item 18
- **relations:** site-design-planner re-resolution ambiguity; duplicate sites-row question (item 20)
- **verify-later:** any delete-site action's library handling

### Language handling: implicit mechanism plus minimal explicit prompt support
- **category:** NEW:language-i18n
- **status-signal:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS)
- **what:** Content language rides implicitly on the brief/specs/existing-content context; Step 3 made the page-content-writer prompt language-agnostic (## Language section, de-Anglicised rule examples, translate-the-intent note for English llm_guidance, any-language placeholder rule). Mapped remaining English-hardcoded surfaces: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, missing <html lang>. Deferred designs: sites.primary_language column (add when a consumer exists), explicit target-language parameter, "soft static" LLM override of Tier B labels, adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification (fallback problem); mega-prompt concerns
- **verify-later:** page-content-writer prompt ## Language section; head template lang attribute

### Two sources of truth for site contact email
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** "sites.email vs site_specs.identity.email can drift. loadSiteContactEmail uses COALESCE across both. Content writers may use either. Needs consolidation." (April 2026)
- **what:** Contact email lives in two places with no single owner; drift produces placeholder/incorrect contact details on pages (a recurring audit finding and false-positive source).
- **sources:** HANDOFF-pipeline-triage-april-2026.md#patterns-1
- **relations:** identity-advisor specialist; content quality catalogue (empty footer contact)
- **verify-later:** loadSiteContactEmail; identity aspect writers

### Discovery-checks list maintenance and the workflow-replace landmine
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "Closed — investigation found no overwriter" (2026-04-19); the jsonb `||` append pattern recommended; updateAgentWorkflow risk "Currently safe because nothing fires it"
- **what:** The suspected "checks keep falling off discovery agents" was manual SQL replacing the whole checks array (a stale in-code example being copy-pasted); the safe pattern is jsonb array append. Latent risk logged: updateAgentWorkflow does jsonb_set of the ENTIRE workflow subtree — when an automated improvement-proposal generator ships, partial proposals will silently erase workflows unless converted to deep-merge.
- **sources:** HANDOFF_2026-04-19_component_linking_news_template_discovery_checks.md#3, #4
- **relations:** improvement_proposals (empty table); ApproveImprovementAction
- **verify-later:** updateAgentWorkflow (context line ~61056); stale comment cleanup

### Debugging meta-lessons (evidence discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** codified into 016_debugging_guide entries: §0 item 19 added 2026-05-26; dispatch doc "Lesson learned" 2026-05-15; naming FOCUS "Tests document behaviour, not intent" (2026-05-17)
- **what:** Recurring investigation disciplines earned across these sessions: grep the whole codebase for the verb (triage/promote/claim) before concluding a writer doesn't exist; a LIKE on prompt_rendered proves what the model was told, never what it did — read response_text; check the guide before generating fresh hypotheses; design tests to falsify; tests assert what a function does, not what was intended; grep chassis logs by the `caller` field (msg gets truncated); logger.Debug is invisible in production; spawned pods are app=dynamic-agent with 600s idle timeout so capture logs before they evaporate; work the smallest useful step; trust suspicion of implausible numbers.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Lesson; HANDOFF_2026-05-26…md#wrong-turns; HANDOFF_2026-04-18_enrichment…md#greps, #false-starts; FOCUS_naming_conventions…md#flags; FOCUS_finetuning…(13).md#14
- **relations:** debugging guide 016 (the canonical home)
- **verify-later:** 016_debugging_guide §0 items

### sites.build_status vestigial column
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "defaulted to 'pending' at insert, never advanced by any code path … Decide whether to maintain or drop the column" (2026-05-26)
- **what:** Site-level build_status is dead; real state lives in last_built_at/last_deployed_at/last_reconciled_at and per-page/per-component build_status. A schema-hygiene decision waiting.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** mark_site_deployed (which flips sites.status, not build_status)
- **verify-later:** any writer of sites.build_status

### API key rotation flag (STABILITY_API_KEY, BANANA_API_KEY)
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "SECURITY — rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext exposure flagged in the imagery handoff). Ops-only action; not addressed" (2026-05-26)
- **what:** Two image-provider API keys were exposed in plaintext and flagged for rotation; still open at the last dated mention.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** imagery pipeline providers
- **verify-later:** whether keys were rotated (ops)

### Firecrawl capability escalation ladder (executeJavascript, waitFor, structured json)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "These are upgrades, not prerequisites" (interactive FOCUS)
- **what:** When plain rawHtml + external-fetch parsing misses dynamically-injected scripts or bundled logic, Firecrawl's executeJavascript actions (script inventory via querySelectorAll), waitFor, and schema-driven json extraction are the escalation path for the parse stage.
- **sources:** FOCUS_interactive_content_generation(4).md#Firecrawl-features
- **relations:** interactive fingerprint C1-C6
- **verify-later:** firecrawl adapter capabilities used today

### Generator architecture convergence (shared interactive-artefact-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Worth considering once two more generators exist; one isn't enough to abstract from"
- **what:** Every content-type generator (tools, games, news articles, dashboards) needs a brief contract, prompt template, persistence action, page-creation step, tiered quality checks and companion-content step; a shared base with per-type specialisation is anticipated once games exist. The library model (canonical templates, forked_from IS NULL, per-site forks) is the copyable storage shape.
- **sources:** FOCUS_interactive_content_generation(4).md#Generator-architecture, #Library-model, #Quality-model
- **relations:** tools pipeline; games gap
- **verify-later:** n/a (design idea)

### Component-creator agent (observed-pattern section components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); regeneration workflow path noted missing ("component-creator only handles needs_new_component") 2026-04-17
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known gap at the time: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path (Track 2, 2026-04-20) but not deactivated-row resurrection (unique-name collisions need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability
- **verify-later:** component-creator workflow; store_generated regen path today

### Improvement-sweep site starvation
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "Oldest updated_at site always wins; sites with frequent rebuilds dominate" — carried P3 across 04-17 → 04-20 handoffs, never picked up
- **what:** The improvement sweep's site selection starves some sites the same way find_dispatchable_site's arbitrary ordering does — scheduling fairness is an unowned concern across both loops.
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#known-issues; HANDOFF_2026-04-20…(2).md#5
- **relations:** dispatch chain fairness ORDER BY
- **verify-later:** improvement-sweep pre_query

### Duplicate sites-row on re-adoption (open investigation)
- **category:** adoption-pipeline
- **status-signal:** unknown
- **status-evidence:** item 20 (2026-04-23, version (1) only): "Couldn't confirm … worth checking on next adoption run"
- **what:** Suspicion that adopting a destination_domain that already has a sites row creates a second row, leaving orphan work items pointing at the stale row while a new cascade runs against the other. Decision needed: refuse when destination exists vs reuse as refresh; duplicate-creation is the worst option.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 20
- **relations:** source/destination separation; library-row cleanup
- **verify-later:** ensure_site_record behaviour on existing domain

---

## Proposed NEW categories

| slug | why |
|---|---|
| NEW:work-dispatch | The detected→triaged→claimed state machine, dispatch chain, claim blockers/timeouts, pipeline label, two-strike rule and silent-completion semantics form a coherent expert domain that spans (and is not owned by) improvement-loop or scheduler-and-tasks. 10 concepts landed here. |
| NEW:prompt-composition | Prompt architecture (mega-prompt fragility, envelope/tool-call/validation patterns, parameter-shaping for images) is design-of-prompts, distinct from llm-quality-testing's evaluation focus. |
| NEW:language-i18n | Language/i18n surfaces (implicit language mechanism, hardcoded-English map, lang attribute, soft statics) have no home in the seed taxonomy. |
