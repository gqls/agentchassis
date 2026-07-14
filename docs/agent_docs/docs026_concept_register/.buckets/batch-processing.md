
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Universal LLM work queue (batch processing architecture)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 v4 (2026-04-12) is a design + phased rollout plan (Phase 0 "deploy infrastructure, everything OFF"); no deployment claim found
- **what:** llm_batch_queue as a provider-agnostic queue (rendered prompt + resolved callback_config stored at queue time) with three-gate resolution (global → agent_type opt-in → provider) and a sync fallback that executes the whole path inline (sync_executed rows prove the restructured pipeline before batch is enabled). Submitter routes to Anthropic Batch API (50% discount, caching adjacency by batch_group), GPU drain mode (worker pool, drain_until/stop-when-empty), or sync; retriever polls, logs to llm_call_log, executes callbacks with retries; urgent escalation makes parallel sync calls and marks late batch results superseded. Batch/sync decision rule: scheduled-task-triggered → batch; user-facing/blocking → sync (~60-70% of spend batchable).
- **sources:** 015 full
- **relations:** callback contract; prompt caching; endpoint health
- **verify-later:** do llm_batch_* tables exist; queue_llm_batch registered?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Batch callback contract (resolved-at-queue-time, context-free callbacks)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 design; eligibility test with three passing callbacks named
- **what:** Callbacks receive only DB + response + resolved callback_config (no collected_data/orchestration state); eligibility test: can it work from a DB connection, response text, and a handful of resolved IDs? Workflow restructure: the post-LLM step disappears into the callback; multi-provider preference lists auto-route without workflow edits.
- **sources:** 015#Callback Contract, #Workflow Restructure
- **relations:** write_audit_findings as first callback
- **verify-later:** batch_callback.go exists?

<!-- SOURCE: U12_docs024_archives.md -->
### Batch-processing control model evolution (two-gate → three-gate, manual → function escalation)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** Live `015_batch_processing_architecture_v2.md` (dated v4: 2026-04-12) shows the three-gate model and SQL escalation functions in place; archive only had earlier forms (dated 2026-04-06).
- **what:** Two mechanisms evolved: (1) batch on/off control moved from a two-gate check to a three-gate resolution (global → `llm_batch_agent_config` per-agent-type opt-in with `batch_group` → provider); (2) urgent-item escalation moved from raw UPDATE statements to dedicated SQL functions `escalate_batch_item(id)`/`escalate_site_batch(site_id)`. A new `sync_executed` status was added for auditable batch-off proving runs.
- **sources:** old/older1/015_batch_processing_architecture.md#"The Table", #"3. Priority Override"; docs024_key_docs_latest/015_batch_processing_architecture_v2.md#"Per-Agent-Type Control Table"
- **relations:** llm_call_log flywheel columns; QueueLLMBatchAction three-gate check
- **verify-later:** `llm_batch_agent_config` table contents; escalation function definitions.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Work item lifecycle (blocking, unblocking, unresolved)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** 001(0) "The unresolved mechanism … Located in load_work_item_actions.go, in insertWorkItem"; 102_blog_handoff "Unresolved status mechanism … line ~893"
- **what:** Items get blocked three ways (missing handler agent → auto-unblocked; spec status blocked → manual; manual block). The unresolved mechanism suppresses a re-emitted item if the newest terminal item with the same `item_key` is <3h old, and creates a visible-but-undispatched `unresolved` item (attempt_count 0) if 2+ prior terminal items exist.
- **sources:** WM/001_development_guide(0).md#work-item-lifecycle-blocking-unblocking-and-unresolved, ED/102_blog_handoff-2026-04-10.md#unresolved-status-mechanism
- **relations:** dispatch loop; feasibility-recheck; dedup index idx_swi_dedup
- **verify-later:** load_work_item_actions.go insertWorkItem; site_work_items status enum

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Dispatch loop & detected→triaged→claimed state machine
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** FOCUS_dispatch_diagnostic(3) "detected is a valid intermediate state, not a bug"
- **what:** Discovery emits `detected` → design-audit-agent runs auditors then `triage_detected_items` → `triaged` → dispatch claims → `claimed` → handler → `complete`/`failed`. The dispatch chain is `scheduled_tasks`(30s) → `build-pipeline-trigger` (one site per tick) → `build-dispatch-loop`. A `NOT EXISTS`-on-claimed clause is an absolute per-site blocker.
- **sources:** WM/FOCUS_dispatch_diagnostic(3).md#tldr, WM/FOCUS_dispatch_diagnostic(3).md#q3-the-dispatcher-architecture-one-site-per-tick-not-exists-blocked-researched-2026-05-15, WM/FOCUS_dispatch_diagnostic(3).md#q4-what-is-the-pipeline-field-actually-for-surfaced-2026-05-15
- **relations:** work item lifecycle; loop mechanisms; claimed-item-timeout; triage_detected_items
- **verify-later:** build-pipeline-trigger find_dispatchable_site SQL; idx_swi_handler/idx_swi_site_pending; triage_detected_items registry.go:722

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item processing_tier (standard / batch_gpu)
- **category:** batch-processing
- **status-signal:** unknown
- **status-evidence:** Bare ALTER in 018: "'standard' — process immediately... 'batch_gpu' — hold until GPU batch starts, then process via GPU Ollama"; no consumer shown.
- **what:** A routing column intended to hold selected work items until a GPU batch window opens so they run on GPU Ollama instead of Claude/CPU — a cost/throughput lever tying the work queue to GPU batch scheduling.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#processing_tier
- **relations:** batch-processing; model-infrastructure (GPU Ollama); thunder GPU provisioning.
- **verify-later:** any writer/reader of processing_tier.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Universal LLM work queue (batch processing architecture)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 v4 (2026-04-12) is a design + phased rollout plan (Phase 0 "deploy infrastructure, everything OFF"); no deployment claim found
- **what:** llm_batch_queue as a provider-agnostic queue (rendered prompt + resolved callback_config stored at queue time) with three-gate resolution (global → agent_type opt-in → provider) and a sync fallback that executes the whole path inline (sync_executed rows prove the restructured pipeline before batch is enabled). Submitter routes to Anthropic Batch API (50% discount, caching adjacency by batch_group), GPU drain mode (worker pool, drain_until/stop-when-empty), or sync; retriever polls, logs to llm_call_log, executes callbacks with retries; urgent escalation makes parallel sync calls and marks late batch results superseded. Batch/sync decision rule: scheduled-task-triggered → batch; user-facing/blocking → sync (~60-70% of spend batchable).
- **sources:** 015 full
- **relations:** callback contract; prompt caching; endpoint health
- **verify-later:** do llm_batch_* tables exist; queue_llm_batch registered?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Batch callback contract (resolved-at-queue-time, context-free callbacks)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 design; eligibility test with three passing callbacks named
- **what:** Callbacks receive only DB + response + resolved callback_config (no collected_data/orchestration state); eligibility test: can it work from a DB connection, response text, and a handful of resolved IDs? Workflow restructure: the post-LLM step disappears into the callback; multi-provider preference lists auto-route without workflow edits.
- **sources:** 015#Callback Contract, #Workflow Restructure
- **relations:** write_audit_findings as first callback
- **verify-later:** batch_callback.go exists?

<!-- SOURCE: U12_docs024_archives.md -->
### Batch-processing control model evolution (two-gate → three-gate, manual → function escalation)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** Live `015_batch_processing_architecture_v2.md` (dated v4: 2026-04-12) shows the three-gate model and SQL escalation functions in place; archive only had earlier forms (dated 2026-04-06).
- **what:** Two mechanisms evolved: (1) batch on/off control moved from a two-gate check to a three-gate resolution (global → `llm_batch_agent_config` per-agent-type opt-in with `batch_group` → provider); (2) urgent-item escalation moved from raw UPDATE statements to dedicated SQL functions `escalate_batch_item(id)`/`escalate_site_batch(site_id)`. A new `sync_executed` status was added for auditable batch-off proving runs.
- **sources:** old/older1/015_batch_processing_architecture.md#"The Table", #"3. Priority Override"; docs024_key_docs_latest/015_batch_processing_architecture_v2.md#"Per-Agent-Type Control Table"
- **relations:** llm_call_log flywheel columns; QueueLLMBatchAction three-gate check
- **verify-later:** `llm_batch_agent_config` table contents; escalation function definitions.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Work item lifecycle (blocking, unblocking, unresolved)
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** 001(0) "The unresolved mechanism … Located in load_work_item_actions.go, in insertWorkItem"; 102_blog_handoff "Unresolved status mechanism … line ~893"
- **what:** Items get blocked three ways (missing handler agent → auto-unblocked; spec status blocked → manual; manual block). The unresolved mechanism suppresses a re-emitted item if the newest terminal item with the same `item_key` is <3h old, and creates a visible-but-undispatched `unresolved` item (attempt_count 0) if 2+ prior terminal items exist.
- **sources:** WM/001_development_guide(0).md#work-item-lifecycle-blocking-unblocking-and-unresolved, ED/102_blog_handoff-2026-04-10.md#unresolved-status-mechanism
- **relations:** dispatch loop; feasibility-recheck; dedup index idx_swi_dedup
- **verify-later:** load_work_item_actions.go insertWorkItem; site_work_items status enum

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Dispatch loop & detected→triaged→claimed state machine
- **category:** batch-processing
- **status-signal:** deployed
- **status-evidence:** FOCUS_dispatch_diagnostic(3) "detected is a valid intermediate state, not a bug"
- **what:** Discovery emits `detected` → design-audit-agent runs auditors then `triage_detected_items` → `triaged` → dispatch claims → `claimed` → handler → `complete`/`failed`. The dispatch chain is `scheduled_tasks`(30s) → `build-pipeline-trigger` (one site per tick) → `build-dispatch-loop`. A `NOT EXISTS`-on-claimed clause is an absolute per-site blocker.
- **sources:** WM/FOCUS_dispatch_diagnostic(3).md#tldr, WM/FOCUS_dispatch_diagnostic(3).md#q3-the-dispatcher-architecture-one-site-per-tick-not-exists-blocked-researched-2026-05-15, WM/FOCUS_dispatch_diagnostic(3).md#q4-what-is-the-pipeline-field-actually-for-surfaced-2026-05-15
- **relations:** work item lifecycle; loop mechanisms; claimed-item-timeout; triage_detected_items
- **verify-later:** build-pipeline-trigger find_dispatchable_site SQL; idx_swi_handler/idx_swi_site_pending; triage_detected_items registry.go:722

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item processing_tier (standard / batch_gpu)
- **category:** batch-processing
- **status-signal:** unknown
- **status-evidence:** Bare ALTER in 018: "'standard' — process immediately... 'batch_gpu' — hold until GPU batch starts, then process via GPU Ollama"; no consumer shown.
- **what:** A routing column intended to hold selected work items until a GPU batch window opens so they run on GPU Ollama instead of Claude/CPU — a cost/throughput lever tying the work queue to GPU batch scheduling.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#processing_tier
- **relations:** batch-processing; model-infrastructure (GPU Ollama); thunder GPU provisioning.
- **verify-later:** any writer/reader of processing_tier.
