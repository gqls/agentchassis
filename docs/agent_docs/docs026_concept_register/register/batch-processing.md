# Register — batch-processing

4 concepts, consolidated from 12 raw extractions (6 unique blocks, each present
twice in the source cluster file due to mechanical duplication in the input)
across units U01, U12, U17a, U19.

### BATCH-001 — Universal LLM batch-processing architecture (queue, three-gate control, callback contract)
- **status:** aspirational
- **status-evidence:** The core design doc (015_batch_processing_architecture_v2.md, v4 dated 2026-04-12) reads on one pass as a phased rollout plan with "Phase 0: deploy infrastructure, everything OFF" and no deployment claim found for the full submitter/retriever pipeline (unit U01); but comparing that same live v4 doc against an earlier archived version (dated 2026-04-06) shows the three-gate control model and SQL escalation functions (`escalate_batch_item`, `escalate_site_batch`) are concretely in place in the live doc/schema — i.e. the control-plane mechanics have been built out even if end-to-end batch execution against real traffic is unconfirmed (unit U12).
- **stage2-verified (2026-07-14):** partial → aspirational — 0 hits for llm_batch_queue, llm_batch_agent_config, QueueLLMBatchAction, queue_llm_batch, escalate_batch_item, escalate_site_batch, batch_callback.go anywhere under platform/, internal/, cmd/, pkg/, deployments/, k8s/, test/ — every hit confined to docs/agent_docs/... (design docs and the concept-register itself). N...
- **what:** `llm_batch_queue` is a provider-agnostic queue that stores the rendered prompt and resolved `callback_config` at queue time. Batch on/off is resolved through a three-gate check (global → `llm_batch_agent_config` per-agent-type opt-in with `batch_group` → provider) — evolved from an earlier two-gate design — plus a sync fallback that executes the whole path inline (`sync_executed` rows exist specifically to prove the restructured pipeline works before batch is switched on). A submitter would route to the Anthropic Batch API (50% discount, caching adjacency by batch_group), a GPU drain mode (worker pool, drain_until/stop-when-empty), or synchronous execution; a retriever polls, logs to `llm_call_log`, and executes callbacks with retries. Callbacks are deliberately context-free: they receive only a DB connection, the response text, and the resolved callback_config (no collected_data or orchestration state), so the post-LLM step of a workflow effectively disappears into the callback and multi-provider preference lists can auto-route without workflow edits. Urgent items can be escalated out of batch via dedicated SQL functions (`escalate_batch_item(id)` / `escalate_site_batch(site_id)`), making parallel sync calls and marking any late batch result superseded. Decision rule: scheduled-task-triggered work batches, user-facing/blocking work stays sync (an estimated 60–70% of spend is batchable).
- **sources:** 015 full (unit U01); old/older1/015_batch_processing_architecture.md#"The Table", #"3. Priority Override"; docs024_key_docs_latest/015_batch_processing_architecture_v2.md#"Per-Agent-Type Control Table" (unit U12)
- **relations:** callback contract; prompt caching; endpoint health; llm_call_log flywheel columns; QueueLLMBatchAction three-gate check; write_audit_findings as first callback; Work item processing_tier (a related but distinct GPU-batch routing lever)
- **verify-later:** do llm_batch_* tables exist; queue_llm_batch registered; batch_callback.go exists; llm_batch_agent_config table contents; escalation function definitions live

### BATCH-002 — Work item lifecycle (blocking, unblocking, unresolved)
- **status:** deployed
- **status-evidence:** 001(0) "The unresolved mechanism … Located in load_work_item_actions.go, in insertWorkItem"; 102_blog_handoff "Unresolved status mechanism … line ~893"
- **what:** Items get blocked three ways (missing handler agent → auto-unblocked; spec status blocked → manual; manual block). The unresolved mechanism suppresses a re-emitted item if the newest terminal item with the same `item_key` is <3h old, and creates a visible-but-undispatched `unresolved` item (attempt_count 0) if 2+ prior terminal items exist. Note: this concept is about the general site_work_items lifecycle rather than batch-LLM processing specifically; it was extracted under the batch-processing category but is more directly a companion to the dispatch-loop/scheduler material in register/scheduler-and-tasks.md.
- **sources:** WM/001_development_guide(0).md#work-item-lifecycle-blocking-unblocking-and-unresolved, ED/102_blog_handoff-2026-04-10.md#unresolved-status-mechanism
- **relations:** dispatch loop (scheduler-and-tasks register); feasibility-recheck; dedup index idx_swi_dedup; Dispatch loop & detected→triaged→claimed state machine (below)
- **verify-later:** load_work_item_actions.go insertWorkItem; site_work_items status enum

### BATCH-003 — Dispatch loop & detected→triaged→claimed state machine
- **status:** deployed
- **status-evidence:** FOCUS_dispatch_diagnostic(3) "detected is a valid intermediate state, not a bug"
- **what:** Discovery emits `detected` → design-audit-agent runs auditors then `triage_detected_items` → `triaged` → dispatch claims → `claimed` → handler → `complete`/`failed`. The dispatch chain is `scheduled_tasks`(30s) → `build-pipeline-trigger` (one site per tick) → `build-dispatch-loop`. A `NOT EXISTS`-on-claimed clause is an absolute per-site blocker. Note: like BATCH-002, this describes the general work-item dispatch pipeline rather than batch-LLM processing specifically — extracted under batch-processing but conceptually a companion to register/scheduler-and-tasks.md (which covers the same dispatch chain from the scheduling side).
- **sources:** WM/FOCUS_dispatch_diagnostic(3).md#tldr, #q3-the-dispatcher-architecture-one-site-per-tick-not-exists-blocked-researched-2026-05-15, #q4-what-is-the-pipeline-field-actually-for-surfaced-2026-05-15
- **relations:** work item lifecycle (above); loop mechanisms (scheduler-and-tasks register); claimed-item-timeout; triage_detected_items; Build pipeline trigger, Dispatch throughput bottleneck (Family J) (scheduler-and-tasks register)
- **verify-later:** build-pipeline-trigger find_dispatchable_site SQL; idx_swi_handler/idx_swi_site_pending; triage_detected_items registry.go:722

### BATCH-004 — Work item processing_tier (standard / batch_gpu)
- **status:** aspirational
- **status-evidence:** Bare ALTER in 018: "'standard' — process immediately... 'batch_gpu' — hold until GPU batch starts, then process via GPU Ollama"; no consumer shown.
- **stage2-verified (2026-07-14):** unknown → aspirational — processing_tier column defined only in docs/agent_docs/sql_for_tables/018_site_work_items.sql:447 (ALTER TABLE ... ADD COLUMN processing_tier TEXT NOT NULL DEFAULT 'standard'); grep for 'processing_tier' across platform/, internal/, cmd/, pkg/, deployments/, k8s/, test/ returns 0 hits — no reader or writer of the co...
- **what:** A routing column intended to hold selected work items until a GPU batch window opens so they run on GPU Ollama instead of Claude/CPU — a cost/throughput lever tying the work queue to GPU batch scheduling.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#processing_tier
- **relations:** Universal LLM batch-processing architecture (above); model-infrastructure (GPU Ollama); thunder GPU provisioning
- **verify-later:** any writer/reader of processing_tier
