
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
