# Register — model-infrastructure

> **covers-through: 2026-07-17** · MDL-038/039 added 2026-07-17 (post-freeze hand-patch).
> Everything else dates from the 2026-07-13 extraction freeze — absence
> here is not evidence of absence in the platform. See `bugs_open/106`.

39 concepts (37 from stage 1 + 2 added 2026-07-17, MDL-038/039, found by the
fix-loop's own first real-case run), consolidated from 53 raw extractions
across units U01, U02, U06,
U09, U12, U15, U16, U17a, U18, U19, U20, U21, U22, U24a, U24b, U24f, U25. Heavy
duplication as expected: U06 (live docs024/finetuning) and U24b (its archived
predecessor) describe near-identical Thunder/GPU-provisioning and
training-launcher mechanics, and endpoint-health/model-swap were each captured
independently by five-plus units. Several entries here overlap with raw blocks
tagged finetuning-flywheel by the same units (the Thunder adapter's
credential-boundary framing, the Phase-5 training-kickoff chain, run.sh
markers, setsid launch, the checkpoint-durability O(K²) fix) — those were left
in the finetuning-flywheel register per the per-category assignment rule and
are cross-referenced from here. Likewise two RAG-flavoured blocks (RAG best
practices, and the RAG deployment bundle) substantially overlap the dedicated
rag-knowledge-base register and are cross-referenced rather than re-described.

### MDL-001 — Model aliases and the model selection strategy
- **status:** deployed
- **status-evidence:** per-step model table (009); a bulk-switch-to-haiku SQL migration records live state 2026-04-10.
- **what:** Agent definitions use short aliases (`claude-sonnet-4-6`, `claude-haiku-4-5`) resolved by `model_aliases.go`; sonnet is the default for LLM steps, haiku for routing, opus for chief-strategist/planner, ollama for fine-tuned classification. A cross-cutting development-guide doc ties this together with llm_call_log, the Ollama adapter, and RAG actions as "LLM infrastructure" in one place. A bulk-cost-lever migration (101) can switch all agents to haiku with a RESTORE section for rollback.
- **sources:** 001_development_guide(5).md#LLM Infrastructure; 009#Model Swap; 101_switch_to_haiku.sql; WM/001_development_guide(0).md#llm-infrastructure
- **relations:** swap_agent_model (MDL-006); LLM tiering (MDL-007); Agent model-assignment upgrade sweeps (MDL-002); Ollama adapter (MDL-003)
- **verify-later:** model_aliases.go; per-step ai_service in agent_definitions

### MDL-002 — Agent model-assignment upgrade sweeps (migration 081)
- **status:** deployed
- **status-evidence:** migration 081 Parts 1-2: chief-strategist → opus-4-6; site-planner/domain-research-classifier/domain-strategist/site-classifier → sonnet-4-6; stale `claude-3-5-sonnet-20241022` and `claude-3-opus` references globally replaced.
- **what:** Model choices live inside `agent_definitions.default_config` and are upgraded by targeted text-replace UPDATEs, with an explicit tiering philosophy: high-leverage structural deciders get the best models, content generation stays on cheaper tiers for cost. Also documents the historical model-vocabulary embedded in configs so future migrations know what strings to expect.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#Part1-2; docs020.../003_llm_model_upgrades_and_logging.sql; docs020.../009_023_session_handoff_vertical_architecture(1).md#done
- **relations:** agent-definition registry; llm_call_log model_resolved column; Model aliases (MDL-001)
- **verify-later:** current model distribution across agent_definitions

### MDL-003 — Ollama adapter (CPU embeddings + local classification)
- **status:** deployed
- **status-evidence:** "Ollama adapter on CPU cluster (2 replicas, mistral-small3.1 + nomic-embed-text)" checked done in one doc; a RAG-deploy handoff elsewhere still lists "Deploy Ollama adapter" as a not-yet-done next step — conflicting claims across sessions/dates.
- **stage2-verified (2026-07-14):** partial → deployed — platform/aiservice/ollama.go exists; createAIClient in platform/orchestration/actions/ai_actions.go:821 has a live 'case "ollama"'; deployments/kustomize/services/ollama-adapter/base/{deployment,service,ollama-pvc,kustomization}.yaml + overlays/production/uk_001 all exist, deployment.yaml:8 replicas:1 (single replic...
- **what:** A permanent CPU adapter serving nomic-embed-text embeddings (~50-100ms) and quantized small models for classification (10-30s acceptable per-build), implemented as an `ollama.go` provider (AIService interface: GenerateText via /api/chat, GenerateEmbedding via /api/embeddings) plus an `ollama-adapter` kustomize deployment (third-party `ollama/ollama` image, PVC for model persistence, init container pulling nomic-embed-text, single replica, ClusterIP 11434). Same AIService interface as Anthropic, including token-usage write-backs. Not for content generation or sub-2s latency.
- **sources:** 001_development_guide(5).md#Ollama adapter; 009#Implementation Status; docs020.../008_README.md; docs020.../009_023_session_handoff_vertical_architecture(1).md; docs021.../026_implementation_todo_vertical_architecture(2).md#0.3
- **relations:** RAG actions (rag-knowledge-base); endpoint health (MDL-005); fine-tuning path (finetuning-flywheel); Self-hosted LLM inference (MDL-026); Ollama CPU adapter operational envelope (MDL-008)
- **verify-later:** ollama.go; ollama-adapter deployment; createAIClient "ollama" case

### MDL-004 — RAG pipeline deployment bundle (ollama-adapter + rag actions + migrations)
- **status:** deployed
- **status-evidence:** live repo confirms `platform/orchestration/actions/rag_actions.go` and other RAG-related action files exist and are registered, matching the deployment bundle's file-placement manifest.
- **what:** The original rollout bundle that added retrieval-augmented generation to the chassis in one go: the `ollama-adapter` k8s service (own kustomize base+overlay, an idempotent init container pulling `nomic-embed-text` ~300MB onto a PVC sized with an 8Gi memory limit leaving room for future 7B models), two new SQL migrations (`llm_call_log` with a stats view, and `knowledge_base` for RAG storage), two new registered actions (`rag_index`, `rag_lookup`), a nullable-helpers Go package, and patches to `ai_actions.go`/`registry.go`/`anthropic.go` (LLM call timing/token logging, an `ollama` case in `createAIClient`). The ongoing mechanism this bundle deployed is registered in full under rag-knowledge-base (RAGK-001/002/003); this entry is the historical deployment-event record.
- **sources:** docs/_archive/agent_docs/docs020_llm_training_rag/007_rag_deployment_README.md
- **relations:** model-infrastructure (MDL-001 anchor doc); Ollama adapter (MDL-003); RAG knowledge_base (rag-knowledge-base, RAGK-001)
- **verify-later:** platform/orchestration/actions/rag_actions.go; deployments/kustomize/services/ollama-adapter/; llm_call_log / knowledge_base tables

### MDL-005 — ai_endpoint_health: multi-endpoint model routing / GPU scheduler
- **status:** deployed
- **status-evidence:** table + view "Applied, verified"; endpoint-health-checker agent + scheduled task "Applied"; independently reconfirmed by a 2026-07-10 audit; gpu-ollama itself noted "currently DOWN, not always-on" across multiple dated sources.
- **what:** `ai_endpoint_health` (migration 085) tracks three endpoints — claude (default quality), cpu-ollama (embeddings + small models), gpu-ollama (70B/LoRAs) — as a single scheduler: healthy → work items flow; unhealthy → items wait (back-to-triage), with no fallback chains (quality over speed; priority means importance only; work items never know about models, only agent definitions do). `ClaimWorkItem` checks the handler's endpoint health before claiming; an `AIUnavailableError` triggers a reactive health update and releases the item to triaged without counting an attempt. Claude uses dual-mode health (reactive 401/402 + an hourly 1-token haiku ping, ~$0.002/mo, for auto-recovery). The kafka-scheduler only probes endpoints listed in this table — which is also why sibling pods like `ollama-eval` stay invisible to production routing by design.
- **sources:** 001_development_guide(5).md#LLM Infrastructure; 009#Decisions Made, #Health Check Architecture; 022_ai_endpoint_health_and_flywheel_llm_call_log.sql; FOCUS_finetuning_flywheel_and_service(13).md#2.3,#2.4d; FOCUS_finetuning_flywheel_and_service(25).md#2.3,#4.3; 085_ai_endpoint_health_checker.sql; FOCUS_finetuning_flywheel_and_service(21).md#2.3,#2.4d
- **relations:** back-to-triage error handling; Flywheel D dedicated eval pod (finetuning-flywheel); build-dispatch-loop claim; GPU/AI-endpoint scheduling design evolution (MDL-019, its superseded predecessor)
- **verify-later:** ai_endpoint_health rows; gpu-ollama current state; whether the three fast-fail/claim-gate/release patches were deployed; claim_work_item health check

### MDL-006 — Model swap/snapshot/revert control plane (migration 083)
- **status:** deployed
- **status-evidence:** migration 083 applied and verified per 009; used operationally per debugging-guide conventions; snapshot_agent used live by a later migration's backup step (110) and no doc records an actual production model swap having happened.
- **what:** `snapshot_agent()` / `swap_agent_model()` / `revert_agent()` plus an `agent_snapshots` view make `agent_definitions` the model-routing control plane: per-step `ai_service` swaps with automatic snapshot, one-call revert, full-table backup as the nuclear option. Post-migration snapshots live in `agent_definitions_backup` (snapshot_taken_at discriminator, restored_at audit trail) — the legacy in-table `is_snapshot` rows caused a documented family of contamination/misroute bugs. This is the deployment mechanism a green Flywheel-D verdict would use to move an agent from Claude to a local model, and doubles as the sanctioned backup tool for hand-applied migrations.
- **sources:** 021_model_swap_and_rollback.sql; 009#Model Swap Procedure; 016 §6.1/§9 snapshot bugs; FOCUS_finetuning_flywheel_and_service(13).md#2.4; FOCUS_finetuning_flywheel_and_service(25).md#2.4; FOCUS_finetuning_flywheel_and_service(21).md#2.4; NOTES_phase5_training_launcher_running(45).md#update-2026-06-09-3
- **relations:** backup naming discipline (MDL-020); LLM config shadowing (MDL-022, step-level swaps shadowed by top-level ai_service); Eval gate before promotion (finetuning-flywheel, FTW-025)
- **verify-later:** agent_definitions_backup schema; migration 083 functions in DB; agent_definitions_backup contents

### MDL-007 — LLM tiering (large/medium/small/none) + cluster-then-slot-fill scaling pattern
- **status:** aspirational
- **status-evidence:** described as design with an `llm_tier` annotation still to add; flip-to-local gated on Thunder health; "No action code touched" to flip medium from Sonnet to local.
- **what:** Every LLM call site declares a tier; the chassis maps tier→endpoint via flippable config (large=Opus for strategy/briefing; medium=Sonnet→local-70B for plan partials/audits; small=Haiku→local for slot-fills; none=deterministic Go). At product-listing scale: facts from feeds resolved in Go, ~10k products clustered algorithmically into ~20-50 groups, one medium call per cluster for framing, one small slot-fill call per product — the same pattern applied at scale for affiliate/product listings.
- **sources:** 029#LLM tier per call site, #Affiliate/product listings; WM/029_site_plan_and_reconciler(1).md#llm-tier-per-call-site, #affiliate-product-listings-same-pattern-applied-at-scale
- **relations:** model aliases (MDL-001); batch queue routing; Model-tiering by task / "3B problem" (MDL-025, a related but distinct doctrine)
- **verify-later:** any llm_tier config keys in agent_definitions

### MDL-008 — Ollama CPU adapter operational envelope
- **status:** deployed
- **status-evidence:** "resolved 2026-04-22. The strategy.type: Recreate pattern is now in the kustomize base"; a memory rule dated from a 2026-04-23 OOM incident.
- **what:** Hard-won ops facts for CPU Ollama: RollingUpdate + RWO PVC deadlocks (new pod can't mount while old pod holds it) — fixed by switching to a `Recreate` deployment strategy; `OLLAMA_LOAD_TIMEOUT=10m` / `OLLAMA_KEEP_ALIVE=30m` (the default 60s load timeout killed first inference after cold start — a 14.4GB model loads in ~45s); pod memory limit must be ≥ model file size + 8-12GiB headroom (Ollama reads host /proc/meminfo but is actually constrained by the cgroup, producing misleading OOM messages); the chassis calls `/api/chat` not `/api/generate`; measured throughput ~150 tok/s prompt, ~2.5 tok/s generation for mistral-small3.1 Q4 on 8 CPU cores.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4c, #14 "Ollama specifics"; FOCUS_finetuning_flywheel_and_service(25).md#2.4c, #2.4d, #14
- **relations:** dedicated eval pod (finetuning-flywheel, FTW-018); endpoint health (MDL-005)
- **verify-later:** kustomize base for ollama-adapter

### MDL-009 — thunder-training-monitor (periodic probe + reconcile + release)
- **status:** partial
- **status-evidence:** "training-monitor VERIFIED live (both paths)… Terminal/decommission branch still never run live… Not enabled"; the scheduled task was inserted DISABLED by migration.
- **what:** A second periodic lifecycle agent beside the reaper: the `thunder-training-monitor` orchestrator runs `find_active_training_instances` then loops spawn_worker→call_worker per running training box — deliberately not the reaper's scheduler-pre_query shape, which merges only the first row per tick and would starve newer instances behind ALIVE boxes. Each `thunder-training-monitor-worker` probes via the adapter's `ssh_get_status`, classifying run.sh markers into a four-way taxonomy: ALIVE (pgrep finds the run → reset streak), DONE_OK (RUN_SH_DONE marker + adapter_config.json exists → mark_complete → decommission), DONE_FAIL (RUN_SH_FATAL → mark_failed → decommission), GONE_UNKNOWN (process gone, no marker — crash/OOM/reap → bump streak, mark_failed at 3 consecutive unreachable probes), with `reachable:false` a valid answer rather than an error. Built as three migrations plus five chassis actions. The ALIVE path and orchestrator fan-out are verified live; the terminal/decommission branch has never actually fired.
- **sources:** STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04; NOTES_phase5_training_launcher_running(45).md#update-2026-06-04,#update-2026-06-09-6; 107_thunder_training_monitor_worker.sql (header); README_worker_statuses.md
- **relations:** thunder-reaper (MDL-012, responsibility split MDL-010); run.sh markers (finetuning-flywheel, FTW-030); monitor enablement gate (finetuning-flywheel, FTW-035); Thunder instance lifecycle (finetuning-flywheel, FTW-038)
- **verify-later:** scheduled_tasks row enabled state; the 5 actions in registry.go; the status_command in the worker config

### MDL-010 — Monitor/reaper responsibility split
- **status:** deployed
- **status-evidence:** "Decision: build a separate thunder-training-monitor, NOT bolted into the time-reaper… the time-reaper is the last-line cost backstop and must stay dead-simple/dependency-free" (2026-06-04).
- **what:** Two periodic agents with deliberately distinct dependency profiles: the reaper (cost backstop) is pure DB + Thunder API and must work even when the adapter's SSH path is down; the monitor (completion-side) depends on adapter + SSH. They overlap only in calling the shared idempotent `decommission_instance`. The monitor exists because the launcher returns long before training ends (a detached run), so completion can't be observed as a workflow await.
- **sources:** NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150; STATUS_thunder_adapter_2026-06_04.md
- **relations:** thunder-training-monitor (MDL-009); thunder-reaper (MDL-012); orphan-sweep (MDL-014, the third leg of the same lifecycle triad)
- **verify-later:** both scheduled_tasks rows and their concurrency_group

### MDL-011 — Thunder Compute adapter (provision/decommission lifecycle)
- **status:** deployed
- **status-evidence:** phases 2-3.5 "Deployed and verified end-to-end"; provision loop "verified end-to-end (2026-05-22)".
- **what:** A Kafka adapter (`system.adapter.thunder.requests`) wrapping the Thunder Compute GPU API and owning its credentials. `provision_instance`: spend pre-check → ed25519 keypair → API create (public_key sent) → k8s Secret persist → WaitForRunning poll → INSERT thunder_instances with retry → compensating cleanup on partial failure. `decommission_instance`: lookup by provisioning_id or thunder identifier → atomic `decommissioning` transition as idempotency anchor → 404-tolerant API + Secret deletes → cost computed from running_since × snapshotted hourly rate. Error classification maps denial→unrecoverable, infra→recoverable. Includes hard-won API-shape knowledge: base URL :8443/v1, lowercase gpu_type enums, camelCase string-numbers in responses vs snake_case ints in requests, recycled numeric ids requiring a partial unique index on live rows, and real template names (`base`, not the OpenAPI example). The finetuning-flywheel register carries a companion entry (FTW-040) framing the same adapter around its credential-boundary and cost-cap design.
- **sources:** STATUS_thunder_adapter_2026-05-12(1).md; FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; FOCUS_finetuning_flywheel_changelog_addition.md
- **relations:** gpu-provisioner; reaper (MDL-012); ssh_exec (MDL-015); presigned data plane (finetuning-flywheel, FTW-032); Thunder adapter credential boundary (finetuning-flywheel, FTW-040)
- **verify-later:** internal/adapters/thunder/*; migrations 025/028/029; thunder_instances schema

### MDL-012 — thunder-reaper scheduled task and per-instance uptime deadline
- **status:** deployed
- **status-evidence:** "Deployed and verified end-to-end (2026-05-14): synthetic row … picked up within 30s"; a live rescue of a specific run recorded 2026-06-04.
- **what:** A 15-min scheduled task whose pre_query finds `running` instances past `max_uptime_hours` and dispatches the idempotent decommission (one per tick, LIMIT 1). The deadline is set by the chassis, not Thunder — computed as running_since + the per-row `max_uptime_hours` (default 18h) — so a mid-train cap overrun can be rescued by bumping a single row's value (done live: 18→48h when a 24h iter_0 train would otherwise have been reaped at hour 18, and with save_strategy=no that would have meant total loss). Reaper reason strings are meaningful text for post-mortems.
- **sources:** STATUS_thunder_adapter_2026-05-12(1).md#3.5; README.md#3.5-delivered; NOTES_phase5_training_launcher_running(45).md#update-2026-06-04-1150
- **relations:** monitor/reaper split (MDL-010); spend gating (MDL-013); GPU training performance model (finetuning-flywheel, FTW-016, the smoke-rate extrapolation that would have prevented the overrun)
- **verify-later:** migration 028; scheduled_tasks row; thunder_instances.max_uptime_hours column

### MDL-013 — Thunder spend gating (DB-side provision check)
- **status:** deployed
- **status-evidence:** "Spend gating lives in DB, not API"; a live cost-gate check example (cap 30, estimate 20, clears with $9 headroom).
- **what:** Before every create, the adapter consults the `thunder_provision_check` view: decommissioned 24h spend + running estimated spend + `estimated_new_run_cost_usd` must stay under `thunder_config.daily_cap_usd`. Operational learnings: keep the per-run estimate realistic (~$20 for a 9h+ A100 run — a $2 test estimate lets doomed runs through; a $25 default blocks legitimate tests); the 24h window is rolling, so heavy test days trip the cap on legitimate past spend (the fix is to raise the cap for the session, not delete accurate rows).
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-172x,#7
- **relations:** Thunder Compute adapter (MDL-011); thunder-reaper (MDL-012); cost anchors from iter_0 (finetuning-flywheel, FTW-013)
- **verify-later:** thunder_provision_check view definition; thunder_config values

### MDL-014 — Orphan-sweep for stale live thunder_instances rows
- **status:** aspirational
- **status-evidence:** "TODO (deferred 2026-05-24) — orphan-sweep for stale live rows… Agreed design (not yet built)."
- **what:** Out-of-band deletions (a manual `tnr delete`) leave DB rows `running` forever; because live rows hold the recycled Thunder id in a partial unique index, a stale row blocks the next provision with a duplicate-key error (hit once in practice, 2026-05-24). Agreed design: a `sweep_orphans` adapter action computes (live DB rows) minus (Thunder's live list) and dispatches the idempotent decommission per orphan, run as a 15-30min scheduled task sharing the reaper's concurrency group, with a safety guard never to sweep on a failed/partial Thunder list. Interim: manual row reconciliation after any manual delete.
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics(TODO)
- **relations:** reaper (time-based, MDL-012) and monitor (completion-based, MDL-009) — this is the third, existence-based leg of the lifecycle triad
- **verify-later:** whether sweep_orphans exists anywhere

### MDL-015 — Adapter-managed SSH access to GPU boxes (ed25519 keys in k8s Secrets)
- **status:** deployed
- **status-evidence:** "RESOLVED & VERIFIED (Phase 4, 2026-05-24) — SSH connection mechanism + ssh_exec/ssh_get_status" with production verification detail.
- **what:** The adapter generates its own ed25519 keypair per provision, stores the private half in a Secret named `thunder-ssh-<db_row_id>` (deterministic, so orphan Secrets are reapable) and sends the public half on create. `ssh_exec`/`ssh_get_status` dial `instance_ip:ssh_port` directly via x/crypto/ssh as user `ubuntu` (not root, despite Thunder's own ssh_command string), with a wait-for-sshd retry (~90s) because RUNNING precedes sshd being up. The port comes from the list-endpoint's `port` field, captured into `thunder_instances.ssh_port`. `reachable:false` is a valid answer, not an error — the probe primitive the monitor builds on. Operators can extract the key from the Secret to watch train.log directly (StrictHostKeyChecking=no is needed because Thunder recycles IPs).
- **sources:** FOCUS_finetuning_flywheel_and_service(25).md#thunder-compute-api-specifics; ssh_probe.sh (header); NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-191x
- **relations:** monitor probe (MDL-009); setsid launch (finetuning-flywheel, FTW-029); RBAC resourceNames trap (Secret permissions)
- **verify-later:** internal/adapters/thunder/ssh/*; the RBAC Role verbs

### MDL-016 — Thunder Prototyping vs Production mode economics
- **status:** partial
- **status-evidence:** "Prototyping (TGV virtualised) worked fine for 70B inference… Phase 2 should default to Prototyping for inference, Production for training (unverified that Prototyping handles long training runs well)."
- **what:** Thunder's Production mode ($1.79/hr A100 80GB) vs Prototyping ($0.78/hr, TGV-virtualised). Verified: Prototyping is fine for 70B inference. Unverified: whether virtualisation overhead degrades long QLoRA training runs enough to cancel the ~55% saving — flagged as an iter_1 experiment that was never run.
- **sources:** HANDOFF_2026-05-08_flywheel_D_iter0_evaluated.md#lessons(8); working/phase2/README.md
- **relations:** cost anchors (finetuning-flywheel, FTW-013); gpu-provisioner defaults
- **verify-later:** provision defaults (mode) in gpu-provisioner/adapter config

### MDL-017 — Thunder checkpoint & artefact upload to B2 + O(K²) batch-presign retirement (migration 110)
- **status:** partial
- **status-evidence:** "Phases A, B, C BUILT and audited 2026-06-05… Phase D adapter side BUILT; its launcher wiring… is the only code left"; the race fix "CONFIRMED in prod" 2026-06-09.
- **what:** Training VMs are ephemeral and hostile: checkpoints and the final LoRA adapter upload via pre-minted single-object write-only presigned PUT URLs in a manifest (keyed by save-INDEX, not global_step; write-once + B2 versioning; a final-upload hard-gate so RUN_SH_DONE implies durable, making the monitor's decommission safe); resume works via the adapter's `prepare_resume_url` reusing `storage.Client.ListObjects` (reuse-before-create — an earlier "genuine adapter gap" claim was wrong). Integrity is the eval gate's job, not the URL scheme's. Two orchestration-layer findings surfaced here with platform-wide relevance: a send-before-register await race (fixed by `preRegisterAwaitedRequest`) and the O(K²) cost of awaited loop substeps (each re-persists the full expanded workflow + growing collected_data) — which retired the K-iteration presign loop in favour of one batch `prepare_object_urls` call, cutting the full launcher path to ~26s. The finetuning-flywheel register (FTW-032) covers the training-pipeline-facing view of this same durability mechanism in more depth.
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(5).md; 110_training_launcher_batch_presign(1).sql; NOTES_phase5_training_launcher_running(39).md#update-2026-06-09,#update-2026-06-09-2,#update-2026-06-09-3
- **relations:** finetuning-flywheel Phase 5/C (FTW-027, FTW-032); thunder-training-monitor gating (MDL-009); chassis await/loop mechanics
- **verify-later:** thunder_prepare_object_url_dispatch.go preRegisterAwaitedRequest; batch prepare_object_urls existence; migrations 109/110 state; data_url_actions.go handlePrepareObjectURLs

### MDL-018 — Anthropic client temperature parameter removed unconditionally
- **status:** superseded
- **status-evidence:** live dev guide, dated inline 2026-05-27: "The Anthropic client no longer sends a temperature parameter on any call... Opus 4.7+ returns a 400 for any non-default temperature."
- **what:** Archived drafts state temperature is stripped only when `budget_tokens` (extended thinking) is set — implying ordinary non-thinking calls still send temperature. The live doc broadens this: because newer Claude Opus models reject any non-default temperature outright, the Anthropic client now omits temperature unconditionally on every call, thinking or not. Temperature remains honoured for other providers (e.g. Ollama) — only the Anthropic client special-cases it. This is a distinct issue from the llm_call_log temperature-logging gap (a separate, still-open observability bug — see llm-call-observability).
- **sources:** old/older1/001h_development_guide_new_agents_v8.md; old/001_development_guide.md#"Extended thinking"; docs024_key_docs_latest/001_development_guide(5).md#"Temperature (2026-05-27)"
- **relations:** endpoints/provider clients (model-infrastructure); LLM call logging `__sent_temperature` (finetuning-flywheel, FTW-004); Temperature/max_tokens logging gap (llm-call-observability, LCO-001, a related but different bug)
- **verify-later:** grep the Anthropic client source for unconditional temperature stripping

### MDL-019 — GPU/AI-endpoint scheduling: design evolution to ai_endpoint_health (superseded drafts)
- **status:** superseded
- **status-evidence:** the live doc's "What's Deployed" section confirms `ai_endpoint_health` table + view "Applied, verified" and the `endpoint-health-checker` agent + scheduled task "Applied" — resolving what earlier archived drafts left as open options.
- **what:** Early design docs posed four undecided GPU-scheduling options (priority-deprioritisation, boolean flag, health-check auto-discovery, back-to-triage only) across a three-layer architecture: (L1) dispatch loop checks endpoint health before claiming and skips items whose handler's endpoint is down; (L2) back-to-triage `AIUnavailableError` releases items without counting an attempt and marks the endpoint unhealthy on 401/402; (L3) GPU lifecycle was manual K8s Service creation the health-checker would auto-discover. This was resolved into the single `ai_endpoint_health` table now documented as MDL-005 — this entry preserves the historical decision trail for anyone reading the archived drafts.
- **sources:** old/older1/020_gpu_and_model_infrastructure.md#"GPU Scheduling: Options Under Discussion"; old/older1/020d_gpu_and_model_infrastructure_v4.md#"Architecture: Three Layers"; old/older1/020c_gpu_and_model_infrastructure_v3.md#architecture-three-layers,#standing-decisions; docs024_key_docs_latest/009_model_infrastructure.md#"What's Deployed"
- **relations:** ai_endpoint_health (MDL-005, the resolved current state); back-to-triage error handling; model swap/revert (MDL-006)
- **verify-later:** ai_endpoint_health table contents; endpoint-health-checker agent definition; claim_work_item_action.go health check

### MDL-020 — agent_definitions backup naming convention (unversioned → _preNNN)
- **status:** superseded
- **status-evidence:** the live doc adds "Naming convention: agent_definitions_backup_YYYYMMDD_pre<NNN>... DO NOT use DROP TABLE IF EXISTS before CREATE TABLE."
- **what:** The archived convention was a plain `agent_definitions_backup_YYYYMMDD` name with no migration tie and no never-drop rule. The live convention requires a `_pre<NNN>` suffix tying the backup to the migration it guards and forbids dropping/overwriting an existing backup.
- **sources:** old/009_model_infrastructure.md#"Migration Safety"; docs024_key_docs_latest/009_model_infrastructure.md#"Migration Safety"
- **relations:** model swap/rollback procedure (MDL-006)
- **verify-later:** recent migration backup table names for `_preNNN` adoption

### MDL-021 — Code-context retrieval infrastructure (analyser adapter + code_symbols)
- **status:** deployed
- **status-evidence:** "MILESTONE 2026-06-12: analyser-adapter DEPLOYED TO PRODUCTION"; a later design doc records "Fix direction: migrate code-indexer's analysis step to analyse_repo_local."
- **what:** The chassis's in-cluster code-indexing pipeline: an `analyser-adapter` (Kafka worker, tarball-fetches a repo read-only, runs the shared `internal/analysis` Go-AST walker) feeds `index_code_symbols`, which embeds symbols (nomic-embed-text via the existing AIService/ollama-adapter seam, reusing the same rag_index/rag_lookup hybrid pattern as knowledge_base) into a sibling `code_symbols` pgvector table (HNSW index, identity-keyed on repo/path/symbol, commit-versioned, hard-deleted since it's a rebuildable cache). Later found to be indexing a year-old stale tree; the fix direction is to swap to `analyse_repo_local`, an in-process fetch-and-analyse path already proven in the diagnose workflow.
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section; NOTES_running_synthesis_v4(39).md 2026-07-02 corpus-check result and DECISIONS
- **relations:** Adapter response envelope contract; B4a embedding-quality finding; diagnosis loop; RAG knowledge_base (rag-knowledge-base, RAGK-001, same embedding/hybrid pattern reused)
- **verify-later:** code_symbols table population/freshness; index_code_symbols action's current data source

### MDL-022 — LLM step config shadowing bug (per-object resolution)
- **status:** partial
- **status-evidence:** "a top-level ai_service shadows step-level overrides … Tracked in FOCUS_step_level_llm_config_ignored.md."
- **what:** `ExecuteLLMPromptAction` resolves the `ai_service` object once, taking the first match wholesale even if it lacks `max_tokens`, so a top-level `ai_service` silently shadows step-level model/max_tokens overrides and `max_tokens` falls back to a hardcoded 2048. Temperature has only one read path and isn't logged (see llm-call-observability for the fuller temperature-specific writeup).
- **sources:** WM/016_debugging_guide_v2_44.md#6.6, #7
- **relations:** LLM infrastructure (MDL-001); LLM tiering (MDL-007); llm_call_log (finetuning-flywheel, FTW-004); Per-field LLM config resolution fallback chain (llm-call-observability, LCO-002, same shadowing family)
- **verify-later:** ExecuteLLMPromptAction; AnthropicClient.GenerateText 2048 fallback; FOCUS_step_level_llm_config_ignored.md

### MDL-023 — Extended thinking configuration
- **status:** deployed
- **status-evidence:** "When budget_tokens is set … the client adds {thinking: {type: enabled, budget_tokens: N}}."
- **what:** Setting `budget_tokens` in an LLM step's `ai_service` config enables Anthropic extended thinking: temperature is removed (an API requirement), response parsing skips thinking blocks, and latency rises 30-90s.
- **sources:** WM/001_development_guide(0).md#extended-thinking-configuration
- **relations:** LLM infrastructure (MDL-001); model aliases (MDL-001); LLM model governance (llm-quality-testing, LQT-004, the migration that prepared but gated this)
- **verify-later:** platform/aiservice/anthropic.go thinking block

### MDL-024 — Static vs dynamic agent deployment + GPU cost strategy
- **status:** aspirational
- **status-evidence:** the source doc is a design ("Integration Steps… Add 3 methods to Agent struct") with a cost table ($1,440 static GPU vs $20+$50 CPU-router+dynamic) and no implementation claim.
- **what:** Same agent code deployed two ways: static agents (pre-deployed Deployments listening on `system.agent.*` with pattern-subscribed response topics) and dynamic agents (spawned Jobs on `job.*` topics); `IsStaticAgent()` switches behaviour. GPU work is handled by an always-on cheap CPU router that spawns short-lived GPU workers (TTL auto-terminate) only when needed — claimed 95% GPU cost reduction versus a static GPU deployment.
- **sources:** docs001_flow_general/README.095b.gpu_image_static_dynamic_agent_strategy.md
- **relations:** image-generator adapter (the CPU/GPU split case in practice); model-infrastructure GPU/Ollama docs are the living area this feeds into
- **verify-later:** GPU_AGENT_STRATEGY env var; whether any router pattern exists in deployments

### MDL-025 — Model-tiering by task ("the 3B problem")
- **status:** aspirational
- **status-evidence:** "A 3B model gets ~60-70% of this right. Errors at the leaf level propagate upward... Use the 3B model only for classification"; an allocation table routes tasks across Opus/7B/BioMistral/NER/3B.
- **what:** A principled task-to-model allocation doctrine from a canine-medicine project: frontier models only for structure-shaping decisions and top-level synthesis; domain-fine-tuned 7B for analysis; specialised tiny models (biomedical NER) beat general LLMs for structured extraction; an embedded 3B model only for binary classification; no LLM at all for retrieval. Pipeline design separates cheap structured extraction from semantic interpretation so the strong model gets one focused call — a doctrine explicitly framed as generalizable beyond the canine project to any large-scale agent workload, and closely related to (but a distinct doctrine from) the chassis-wide LLM tiering pattern.
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md#The-3B-Problem, #The-Paper-Analysis-Pipeline; docs016_dogs_medicine_pathways/002_project_outline.md
- **relations:** model-infrastructure (Ollama/GPU hosting); finetuning-flywheel; LLM tiering (MDL-007); embedded worker-pod models
- **verify-later:** inference cluster configs; any vLLM/BioMistral deployment manifests

### MDL-026 — Self-hosted LLM inference (vLLM/GPU at scale)
- **status:** aspirational
- **status-evidence:** "Phase 2: Self-Hosted LLM Validation — Deploy vLLM or llama.cpp serving a 7B model"; cost tables for a 48-hour million-agent run.
- **what:** A plan to serve 7B models (Mistral/Llama 3/Qwen 2.5) on GPU via vLLM with continuous batching to escape per-token API costs at scale (1,000-2,000 req/min per A100). Bridges to the Ollama/local-model path and the LoRA fine-tuning targets. Estimated hybrid GPU+CPU cost $1,000-3,000 for a 48-hour million-agent burst.
- **sources:** docs021.../015_scaling_analysis.md#phase-2, #cost-estimates
- **relations:** Ollama provider (MDL-003); LoRA fine-tuning (finetuning-flywheel); worker pools
- **verify-later:** any vLLM/GPU inference deployment or stub_llm action

### MDL-027 — RAG best practices (filter-first-then-rank; superseded v1)
- **status:** superseded
- **status-evidence:** dated 2026-03-24, in the old/older1 archive; the live successor `docs020_llm_training_rag/012b_rag_best_practices_v2.md` exists and is the version registered in full under rag-knowledge-base (RAGK-004).
- **what:** The v1 RAG guidance for the site pipeline: always filter `knowledge_base` by structured metadata (vertical, component_type, source_quality) before embedding-similarity ranking to avoid cross-vertical contamination; keep RAG at 20-30% of the context window (2-8 examples, quality over quantity); use nomic-embed-text with `search_query:`/`search_document:` task prefixes (recommending a nomic-v2-moe upgrade); quality-gate scraped/Claude/human/audit sources; track `embedding_model` and never mix embedding spaces. Content-identical in spirit to its v2 successor, which is the canonical version to consult.
- **sources:** old/older1/012_rag_best_practices.md#core-principle, #embedding-model-choice, #avoiding-common-rag-failures
- **relations:** replacement = RAG best practices v2 (rag-knowledge-base, RAGK-004); quality flywheel (finetuning-flywheel, FTW-002); canine biology knowledge base
- **verify-later:** rag_index/rag_lookup actions filter+prefix; knowledge_base metadata columns

### MDL-028 — Model quality assessment & per-agent model assignment
- **status:** partial
- **status-evidence:** "Tested but Not Persistent": Llama 3.3 70B on H100 (classification 8/10, content 9/10, design 7/10), Mistral Small 3 CPU (5/6/3); a recommended assignment table routes strategist/webdesign/planner→Claude, classifier/content-writer/triage→Llama70B GPU, briefing→Mistral CPU; cost projection ~$910-990 vs $15-30k all-Claude.
- **what:** Benchmarked model quality per task (Claude as reference, Llama 70B near-parity on content/classification, Mistral weak on design) and a per-agent endpoint-assignment mapping high-leverage structural work to Claude and bulk content/triage to GPU Llama, projecting ~95% cost reduction at 2000-domain scale. Model routing is controlled via `agent_definitions.ai_service` (swap + snapshot). This is the same underlying benchmark data registered from the quality-testing angle under llm-quality-testing (LQT-001); this entry is the per-agent-assignment/cost-projection framing of the same evidence.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#model-quality-assessment, #cost-projection, #models-to-evaluate
- **relations:** endpoint health table (MDL-005); snapshot/swap/rollback (MDL-006); RAG/LoRA flywheel; Model quality assessment: local 70B comparable (llm-quality-testing, LQT-001)
- **verify-later:** agent_definitions ai_service per agent; model aliases claude-sonnet/opus

### MDL-029 — Flywheel C — LoRA fine-tuning path & iter0 adapter output (Unsloth QLoRA Llama 3.3 70B)
- **status:** deployed (first run closed out)
- **status-evidence:** "iter_0 CLOSED OUT … adapter_model.safetensors 828MB … training_run 1cd65dd7 reconciled to complete"; the eval tree's own model-card frontmatter confirms base model and training stack (PEFT 0.19.1, tags lora/sft/unsloth/peft).
- **what:** The training pipeline as seen from the model-infrastructure side: pull dataset from Postgres → Unsloth QLoRA train Llama 3.3 70B Instruct (`unsloth/Llama-3.3-70B-Instruct-bnb-4bit`, 3 epochs, batch 1, grad-accum 8, lr 2e-4, lora_r 16, max_seq 4096) → inference sanity test → LoRA adapter (~150-828MB depending on save precision). Base 70B chosen because hardware was already available, though 8B was flagged as likely 95% of quality at 10% of cost. The real run was ~24h (not the scripts' originally-claimed 30-90 min). The resulting adapter is held in `iter0_eval/lora_iter0_full/`, whose README is an unfilled auto-generated HuggingFace model-card template — the load-bearing content is just its YAML frontmatter. The finetuning-flywheel register (FTW-011, FTW-012, FTW-013) carries the fuller training-pipeline narrative including cost/time/loss anchors and version pinning.
- **sources:** NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150,#update-2026-06-05-iter_0-closed-out; FOCUS_finetuning_flywheel_and_service(21).md#2.5; working/eval/iter0_eval/lora_iter0_full/README.md#frontmatter
- **relations:** consumes training_exports datasets (finetuning-flywheel, FTW-006); deployed via Phase 5 launcher (MDL-031); produces the LoRA iter0 adapter; superseded automation design = Flywheel C Phase 2 (MDL-030)
- **verify-later:** flywheel_C/02_train_llama_3_3_70b.py, 01_pull_dataset_from_postgres.sh, run.sh; iter0_adapter_out/adapter_model.safetensors

### MDL-030 — Flywheel C Phase 2 — HTTP-job-server training automation
- **status:** abandoned
- **status-evidence:** "design locked, not built" (2026-04-23), proposing model-trainer/model-evaluator/training-flywheel-orchestrator + a `POST /jobs` VM server; superseded in practice by the Kafka/saga Phase 5 chain where model-trainer is an orchestrator, not an HTTP-polling agent.
- **what:** An abandoned design where a `model-trainer` specialist would POST a dataset to a ~200-line FastAPI-style HTTP job server running on the GPU VM (`POST /jobs`, `GET /jobs/{id}`, download adapter), polling to completion, with three new tables (`model_training_runs`, `model_artefacts`, `model_evaluations`). SSH-remote-exec and a VM Kafka consumer were both explicitly rejected at the time. The actual Phase 5 build instead made the VM credential-free, with the chassis driving via thunder-adapter presigned URLs. The finetuning-flywheel register (FTW-026) carries the fuller three-generation evolution story this design was the first chapter of.
- **sources:** FOCUS_finetuning_flywheel_and_service(21).md#2.5.1; #15 changelog 2026-04-23
- **relations:** superseded by Phase 5 training-launcher + model-trainer saga (MDL-031); the schema names live on as `model_lifecycle.training_runs` (finetuning-flywheel, FTW-024); Flywheel C phase-2 automation architecture (finetuning-flywheel, FTW-026)
- **verify-later:** model_lifecycle.training_runs; confirm no `/jobs` HTTP server exists in repo

### MDL-031 — Phase 5 training-launcher + model-trainer orchestration chain
- **status:** deployed
- **status-evidence:** "batch route CONFIRMED end-to-end … Launcher green … LAUNCH_PID=216 … COMPLETED → notified parent success" (2026-06-09).
- **what:** The real `training-launcher` (migration 102, replacing a stub) driven by the `model-trainer` orchestrator, which spawns then calls `training-data-preparer → gpu-provisioner → training-launcher` over Kafka/saga. The launcher presigns dataset+scripts, computes checkpoint keys, presigns them, assembles an upload manifest, SSHes it onto the VM, and launches training detached. A two-level await distinction (the child's intermediate adapter calls vs. the child→parent final notification) is load-bearing for correctness. The finetuning-flywheel register (FTW-027, FTW-028) covers this same chain with more granular step-by-step detail.
- **sources:** NOTES_phase5_training_launcher_running(39).md#1,#5; RUNBOOK_iter0_pretrigger(3).md#6; HANDOFF_2026-05-24_phase5_launcher_build.md
- **relations:** replaces Flywheel C Phase 2 HTTP-server design (MDL-030); children call thunder-adapter (MDL-011); model-trainer orchestration chain (finetuning-flywheel, FTW-027)
- **verify-later:** agent_definitions training-launcher (1223bdc1), model-trainer (94f5a069); migrations 102/109/110

### MDL-032 — setsid detached launch command
- **status:** deployed
- **status-evidence:** "ssh_exec blocks to command exit (§2), so the launch must return immediately … the SSH channel hits EOF right after echo"; confirmed `LAUNCH_PID=216` (2026-06-09).
- **what:** Because the adapter's `ssh_exec` runs `session.Run` and blocks up to a 5-min timeout for the remote command to exit, the launch command runs the fetch+train chain under `setsid bash -c '…' </dev/null >launch.log 2>&1 &` and echoes `LAUNCH_PID=$!`, so the SSH channel EOFs immediately. An early superseded version used `nohup`; a real bug found later required `write_manifest` (the first VM-filesystem touch) to use `sudo mkdir`/`sudo chown /workspace` because `/` isn't ssh-user-writable. The finetuning-flywheel register (FTW-029) covers this same mechanism plus its "detached exit-0 false-success gap" implications in more depth.
- **sources:** NOTES_phase5_training_launcher_running(39).md#4, #update-2026-06-05-deploy-step-2
- **relations:** part of training-launcher (MDL-031); 109a permission fix; run.sh markers (MDL-033); setsid detached launch + false-success gap (finetuning-flywheel, FTW-029)
- **verify-later:** thunder_ssh_exec_dispatch.go; ssh_exec_actions.go sshCommandTimeout

### MDL-033 — run.sh RUN_SH markers + set -e durability hard-gate
- **status:** deployed
- **status-evidence:** "run.sh — BUILT 2026-06-05 … set -euo pipefail plus 02_train's final-upload hard-gate … RUN_SH_DONE are only reached on exit 0, which now implies the adapter is in B2."
- **what:** The on-VM launch chain emits grep-able markers (`RUN_SH_START → RUN_SH_STEP setup → RUN_SH_SMOKE_OK → RUN_SH_STEP full_train → RUN_SH_FULL_OK → RUN_SH_DONE`). Because `set -euo pipefail` plus the final-upload raise means DONE only prints on exit 0, `RUN_SH_DONE` came to mean "trained AND uploaded" — the flip that makes the monitor's DONE_OK→decommission safe. `SAVE_STEPS` (checkpoint cadence) lives in run.sh, default 50 (~1.5h/checkpoint); lowered to 10 for fast tests. The finetuning-flywheel register (FTW-030) is the fuller writeup of this same marker protocol.
- **sources:** PLAN_checkpoint_and_artefact_upload_b2(4).md#run.sh; RUNBOOK_phase_b_c_d_deploy(7).md#step-4; NOTES_phase5_training_launcher_running(39).md#8 (Healthy markers)
- **relations:** parsed by thunder-training-monitor probe (MDL-009); gates checkpoint upload path (MDL-017); run.sh launch chain (finetuning-flywheel, FTW-030)
- **verify-later:** run.sh (bundle at finetuning/scripts/bundle.tar.gz); 02_train --upload-manifest

### MDL-034 — iter0 pre-trigger + Phase B/C/D deploy runbooks
- **status:** deployed
- **status-evidence:** the iter0 pretrigger runbook's §6 triggers `model-trainer` with a specific export id; the Phase B/C/D runbook's "One-line summary of the gates" is dated 2026-06.
- **what:** Two operational runbooks. The iter0 pretrigger runbook lists the gates to reach the first automated training launch (deploy adapter+chassis, upload the scripts bundle, adapter round-trip, cost gate, a gpu-provisioner smoke test of the D4 topic path, then trigger model-trainer). The Phase B/C/D deploy runbook stages the checkpoint-upload rollout: apply a migration → re-pack/re-upload the bundle → a short Tier-2 launch (B+C integration, SAVE_STEPS low) → resume (blocked on a later migration) → enable the monitor last. Both hard-code the `b2` CLI (not `aws`) and a "verify positive evidence, complete≠succeeded" discipline throughout.
- **sources:** RUNBOOK_iter0_pretrigger(3).md; RUNBOOK_phase_b_c_d_deploy(7).md; NOTES_phase5_training_launcher_running(39).md#7 (Pre-trigger gates)
- **relations:** operationalises the launcher (MDL-031) + checkpoint upload (MDL-017) + monitor (MDL-009); an export id is documented as a do-not-use trap
- **verify-later:** migrations 109/109a/109b; scheduled_tasks thunder-training-monitor enable step

### MDL-035 — Per-workflow-step model routing (data-sovereignty mechanism)
- **status:** deployed
- **status-evidence:** an audit dated 2026-07-10 confirms "TRUE — ExecuteLLMPromptAction resolves ai_service in a three-tier lookup, tier 2 being workflow.steps[step].config.ai_service"; "ollama-adapter … ClusterIP only."
- **what:** Model/provider is selectable per workflow step with no new code required (a three-tier `ai_service` resolution), with live swap tooling (`swap_agent_model`, migration 083). A self-hosted step genuinely never leaves the cluster (stock ollama image, ClusterIP-only, in-cluster calls). This underpins an honest marketing/positioning claim: "steps that touch your data can run on infrastructure you control; only steps that don't need to leave call a foundation model."
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6
- **relations:** text-provider wiring reality (MDL-036); data-sovereignty positioning; no-tenant-isolation fact; model swap/revert control plane (MDL-006)
- **verify-later:** ai_actions.go ExecuteLLMPromptAction; 021_model_swap_and_rollback.sql; ollama-adapter service spec

### MDL-036 — Text-provider wiring reality (two providers end-to-end)
- **status:** deployed
- **status-evidence:** an audit dated 2026-07-10: "createAIClient switch: anthropic and ollama only; openai is a stubbed error … 'Mistral' is … run through the same self-hosted Ollama pod."
- **what:** Only Anthropic and Ollama actually work end-to-end for text; nothing else is wired. Imagery is broader (Gemini + Stability, routed by kind not config). The news pipeline has a separate hand-rolled provider path (xAI /v1/responses with web_search+x_search, OpenAI, Perplexity) that bypasses the generic AIService entirely. A real constraint on any "model choice" marketing claim.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P3; docs/leopardessconsulting/RUNBOOK.md#landmine-12
- **relations:** per-step model routing (MDL-035); news-feed-pipeline (separate provider path)
- **verify-later:** createAIClient switch; feed_actions.go provider paths

### MDL-037 — llama3.3:70b trained but never used for inference; dynamic GPU provisioning
- **status:** partial
- **status-evidence:** an audit dated 2026-07-10: "one complete run (2026-06-03→04) … No agent_definitions row points at llama3.3:70b — trained and tested, never used for production inference. TODO logged in 009_model_infrastructure.md." Independently confirmed by the finetuning-flywheel register's own fine-tuning-path entry (FTW-003), which records the identical fact from the training-pipeline side.
- **what:** The larger self-hosted model exists as a completed training run; GPU provisioning is genuinely dynamic (thunder_instances, ThunderCompute, decommissioned per run) but experimental. Wiring it to production inference would strengthen the data-sovereignty positioning; deliberately logged in the model-infrastructure home doc rather than duplicated elsewhere.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6; docs/leopardessconsulting/RUNBOOK.md#Reference
- **relations:** Fine-tuning path (finetuning-flywheel, FTW-003, same fact); per-step model routing (MDL-035)
- **verify-later:** model_lifecycle.training_runs; agent_definitions model references

### MDL-038 — BUG A: GenerateText never decodes stop_reason (truncated success mislabelled complete)
- **status:** deployed (as a bug — confirmed present and unfixed in current source; found by the fix-loop's own first real-case CONFIRMED diagnosis, 2026-07-16)
- **status-evidence:** CONFIRMED on 3 citations by the diagnosis loop (correlation `e505f70f-b9e2-4654-9942-30fb13731ca9`, slug `needs_diagnosis:stop-reason-undecoded`): the response struct (only Content+Usage decoded), the text-block-return loop, and a state-tier citation of live `llm_call_log` rows (17 calls with `output_tokens == max_tokens`, all `success=true`) fetched by the loop's own data_request. Graded PASS against a pre-registered rubric (`fixloop_eg_dartsonline/RUBRIC_2026-07-16_two_config_bugs.md`, written before dispatch). Independently re-confirmed by direct code read: `platform/aiservice/anthropic.go:67` (`GenerateText`)'s response struct at lines 158-167 declares only `Content []struct{Type,Text}` and `Usage struct{InputTokens,OutputTokens}` — no `stop_reason` field anywhere.
- **what:** Anthropic's API returns `stop_reason` (e.g. `"end_turn"` vs `"max_tokens"`) on every response, but `GenerateText`'s anonymous response struct never declares that JSON field, so `json.Unmarshal` silently drops it. A response truncated mid-generation by hitting `max_tokens` returns HTTP 200 with well-formed (partial) content and is treated as a complete success at every layer above — no error, no warning, no failed work item. The loop's own `llm_call_log` state-tier evidence shows 17 calls where `output_tokens` exactly equals the configured `max_tokens` ceiling, all logged `success=true` — the signature of silent truncation happening for real, at scale, undetected.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 34 "FIRST REAL-CASE CONFIRMED"; fixloop_eg_dartsonline/RUBRIC_2026-07-16_two_config_bugs.md; platform/aiservice/anthropic.go:67,158-167
- **relations:** same "silently wrong, no error surfaced" shape as styling-render-pipeline's STY-049 (missingkey=zero) and its whole cross-cutting family — a config/response value quietly not doing what its presence implies, discovered the same week; MDL-039 (BUG B, same diagnosis-loop session, same failure flavour: silent misconfiguration believed correct)
- **verify-later:** whether a fix (decode + surface stop_reason, e.g. retry-on-truncation or a loud error) has been dispatched to fix-proposer — CONFIRMED but "fix dispatch awaits owner go" as of 2026-07-16/17; platform/aiservice/anthropic.go response struct for a stop_reason field; llm_call_log rows where output_tokens == max_tokens fleet-wide (scope of the blast radius beyond the 17 already found)

### MDL-039 — BUG B: root ai_service SHADOWS step-level ai_service (dead config believed live)
- **status:** deployed (as a bug — confirmed present in current source, non-trivial fleet blast radius; found the same session as MDL-038, terminal state PARTIAL not CONFIRMED — see status-evidence)
- **status-evidence:** Confirmed directly by code read: `platform/orchestration/actions/ai_actions.go`'s `ExecuteLLMPromptAction` checks `agentConfig["ai_service"]` (a top-level key in the agent's `default_config`, sibling to `"workflow"`) FIRST (lines ~150-157); only if that is nil does it fall back to the step's own `workflow.steps.<step>.config.ai_service` (lines ~160-178). Proven by direct experiment on `diagnose-agent` (2026-07-16): step-level `max_tokens: 8000` had never applied since 2026-07-10 (every verdict logged `max_tokens=2048`, the client default) until `max_tokens` was moved to the agent's ROOT `ai_service` block, after which the next call logged `max_tokens=32000`. The loop's own diagnosis run (`960b554d`) independently re-derived the identical 5 static code citations and graded the symptom rubric-perfect, but terminated **PARTIAL** (gated UNVERIFIABLE at iteration-cap) because the two-evidence-family guard requires both a static AND a state/runtime citation for CONFIRMED, and no state-tier evidence was fetched in that run — the mechanism is fully established by code + a direct experiment, just not by the loop's own formal verdict.
- **what:** An agent_definitions row's `default_config` can carry an `ai_service` block either at the TOP LEVEL (sibling to `workflow`) or nested inside an individual workflow step's own `config`. The documented runbook rule ("max_tokens lives inside a step's ai_service block; root is dead config") is **backwards** — it was only ever true for agents that happen to have no root block (e.g. page-content-writer, whose 2000→8000 fix worked and got wrongly generalised into a rule). The correct rule: **root wins; the step's entire ai_service block is dead the moment a root block exists.** Fleet blast radius: 17 agent_definitions have a root `ai_service` with no `max_tokens` set (silently defaulting to 2048), 10 of which (the whole content-creator-* family) separately declare a `max_tokens` elsewhere in a step config that is completely inert — a config believed live that has never once taken effect.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 34 "BUG B — root ai_service SHADOWS step-level"; platform/orchestration/actions/ai_actions.go:147-193 (ExecuteLLMPromptAction)
- **relations:** MDL-038 (BUG A, same session); this register's own stage-2 finding that `fix-proposer` (`fixloop_eg_dartsonline/0NN_fix_proposer.sql`, home of FIX-014/015/051/052/053 and the bug-historian pilot) has NO top-level `ai_service` key in its `default_config` — independently confirmed 2026-07-17 by re-reading the file, so the fix-proposer workflow (and its new bug-historian reviewer) is NOT among the 17 affected agents
- **verify-later:** fleet-wide query for agent_definitions rows with a root `ai_service` and no `max_tokens`, cross-referenced against ones that also declare `max_tokens` in a step config (the 10 content-creator-* family specifically); whether the 17-agent fleet fix has been applied, and whether it was done config-first or code-first (owner decision noted as still open in the source)
