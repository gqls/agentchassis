# Cluster: hitl-onboarding-agentorg-scheduler
Categories included: hitl, onboarding-config, new:agent-spawning-and-groups, new:agent-memory-and-evolution, new:persona-architecture, new:flows-and-narrative, new:org-framework, new:agent-tree-navigation, new:agent-swarm-simulations, new:agent-definition-registry, scheduler-and-tasks, batch-processing


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Approval model (auto / hitl / eval) with four override levels
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** P1: "Future — column exists from day one" for eval; no deployment evidence
- **what:** site_work_items.approval_mode: auto (default), hitl (pending_review → human approves), eval (evaluation agent reviews handler output before completion). Overrides per-item, per-item-type, per-site, system default.
- **sources:** P1#Approval Model
- **relations:** content-reviewer as future eval agent; P10 recommendation specialists
- **verify-later:** approval_mode column exists?

<!-- SOURCE: U15_docs019_running_notes.md -->
### Governance/HITL confirm-not-initiate model
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** "Confirm-not-initiate. Decision reasoning is agent-led; the human confirms via a decision package" (principles(59) §Governance and HITL).
- **what:** A framing where every decision publishes its reasoning (not just its outcome) so drift detection is possible; a decision requiring gating routes through exactly ONE confirming component (never reimplemented per producer); a newer proposal for an already-pending target supersedes and expires the older one (freshness over queue order); and inheritance has two precedence directions — normal entries are child-wins, sealed constraints (legal floors, mission non-negotiables) are ancestor-wins.
- **sources:** NOTES_running_synthesis_principles(59) §Governance and HITL (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; diagnosis loop (embodies read-only + human-gated in practice); council hard_veto flag model.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Confirm-not-initiate HITL model (decision package)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §5 "Rules are not authored by humans from scratch and are not changed unilaterally by agents. The decision's reasoning is agent-led; the human confirms"
- **what:** Agents carry the analysis and produce a decision package (summary, tradeoffs/impact, genuine choices incl. reject and defer); the human confirms an informed choice, with rigor scaled to stakes. Acceptance writes a new version and deprecates the old (never deletes).
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#5, ED/FOCUS_best_practice_doc_tree(1).md#4.2, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** concern curators; coordinator; governance gate; deprecate-not-delete
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### User-representative advocate (intent + conflict triage)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §8 "A standing agent … representing the user sits inside the framing process as the check on the coordinator's framing power"
- **what:** A standing advocate for the user's intent that checks the coordinator's framing. Its signature job is triaging claimed conflicts before any reach the user: dissolve illusory or already-reconciled tensions, ask a clarifying question when unsure a conflict is real, and escalate only genuine tradeoffs.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#8, ED/FOCUS_standards_curation_and_governance(1).md#8.2, ED/FOCUS_standards_curation_and_governance(1).md#8.4
- **relations:** coordinator; decision authority; concern curators
- **verify-later:** intake-orchestrator; briefing-agent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Decision authority (co-equal voices, abstention, creator veto)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §9 "Default: co-equal voices in the frame … Abstention fallback: the advocate decides, bounded … Creator override / veto"
- **what:** When advocate and curator genuinely disagree, both cases go to the human as co-equal voices. If the user declines to choose, the advocate decides bounded by codified intent — but high-stakes abstention and blocker conflicts escalate to the creator, who holds veto or optional final choice. Three distinct human roles: user, confirmer, creator.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#9
- **relations:** user-representative advocate; coordinator; confirm-not-initiate
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Human direction channels + lock lifecycle + audit-pass cap
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 007 "Humans influence sites at three levels … Lock types: permanent / timed / review"; "The 3-pass audit cap prevents unbounded improvement cycles"
- **what:** Humans steer sites via three channels. HITL-requested content is protected by lock types (permanent, timed-90d, review-creates-HITL-item-on-expiry). A 3-pass audit cap bounds improvement cycles and resets on time/direction-change/major-rebuild/manual.
- **sources:** WM/007_adoption_pipeline_v3.md#human-direction, WM/007_adoption_pipeline_v3.md#hitl-requested-content-and-lock-lifecycle, WM/007_adoption_pipeline_v3.md#audit-pass-cap-reset
- **relations:** locks (031); improvement loop; direction spec
- **verify-later:** site_specs direction aspect; lock_type/lock_expires_at; last_audit_reset_at

<!-- SOURCE: U18_sql_for_agents.md -->
### intake-orchestrator (classify → brief → HITL confirm → spawn builder → rerender)
- **category:** hitl
- **status-signal:** superseded
- **status-evidence:** 068_domain_submitter_agent.sql creates a new entry point ("Entry point for new domain submissions... creates needs_domain_research work item"); intake files stop being patched after 030-era; 002 header shows HITL steps with `skip_if: input_data.hitl_mode == auto`.
- **what:** The v1/v2 entry pipeline: discovers available `%-builder` agents, spawns site-classifier and briefing-agent, runs two human-in-the-loop gates (confirm site type; review brief), spawns the recommended builder, then (added later) a rerender pass for nav consistency. Notable mechanisms: `dynamic_select` HITL fields fed from a live agent query; `skip_if` auto mode making HITL optional per run.
- **sources:** 002_intake_orchestrator.sql; sql_for_agents_v1/001_agent_definitions_etc.sql; sql_for_agents_v2/002_intake_orchestrator.sql
- **relations:** superseded by domain-submitter + build-dispatch-loop; HITL gate pattern survives in content-reviewer HITL mode
- **verify-later:** agent_definitions 'intake-orchestrator' status; whether any trigger path still uses it

<!-- SOURCE: U18_sql_for_agents.md -->
### content-reviewer (HITL + auto-eval dual mode with pre-validation)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 025 adds validate_content step "runs BEFORE review mode determination"; agent defined active in v2/025.
- **what:** Reviews generated page content in either human (request_human_input, editable) or auto-eval (LLM approve/flag) mode, selected at runtime. 025 adds algorithmic pre-validation: internal links must point to existing site pages; email addresses must match the site's contact_email — findings are handed to whichever review mode runs.
- **sources:** 025_content_reviewer_agent.sql; sql_for_agents_v2/025_content_reviewer_agent.sql
- **relations:** page-content-writer output consumer; HITL pattern from intake-orchestrator
- **verify-later:** validate_page_content action; current review-mode default

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item approval_mode (auto / hitl / eval)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Column added in Phase-0 Block A: "Controls whether items auto-dispatch or require human/eval approval. Values: 'auto' (default), 'hitl', 'eval'".
- **what:** Per-item gating between triage and dispatch: auto items flow straight through; hitl items wait for human approval; eval items wait for an evaluation agent. The schema-level hook for configurable human review gates in the build pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#4
- **relations:** work queue; input_requests; human change requests.
- **verify-later:** dispatch respects approval_mode; any 'eval' users.

<!-- SOURCE: U19_sql_tables_components.md -->
### Human change-request work items
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Applied INSERT batches for gaswholesalers.com and finetuning.uk: needs_design→webdesign-agent, content_edit→section-editor (field_updates spec), needs_logo/needs_hero_image→image-build-handler (image_prompts spec), needs_rerender→rerender-pages at priority 99.
- **what:** Human requests enter the same queue as agent-detected work: source='human', item_key 'human_<what>_<site>', pre-triaged, priority-ordered so content and imagery land before the final rerender. The spec JSONB is the full handler contract (edit_type/page_name/slot_name/field_updates for edits; purpose + image_prompts for imagery).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** work queue; section-editor; image-build-handler; dispatch loop (071).
- **verify-later:** admin UI path creating these; section-editor field_updates handling.

<!-- SOURCE: U19_sql_tables_components.md -->
### input_requests HITL persistence
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Full schema with pending view, expiry function and debugging queries; FK to orchestration_states with CASCADE.
- **what:** Human input requests persisted for querying and UI display: request_type (review/confirmation/questionnaire), title/message/data/ui_config, reply_to_topic for Kafka response routing, timeout/expires_at, status lifecycle pending→completed/expired/cancelled, and the response payload with responder identity. pending_input_requests view feeds the UI with seconds_remaining.
- **sources:** docs/agent_docs/sql_for_tables/006_input_requests.sql
- **relations:** awaited_requests (transport-level counterpart); approval_mode; admin dashboard.
- **verify-later:** UI consumption; expire_input_requests scheduling.

<!-- SOURCE: U19_sql_tables_components.md -->
### Manual HITL continuation runbook
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Working queries against awaited_requests for stuck escalate_to_human steps, plus the documented hitl_respond.sh Kafka invocation with real ids/topics.
- **what:** Operational procedure for un-sticking HITL flows: locate the awaited request (step_name LIKE %human%/%hitl%/%approval%), optionally reset an expired one, then publish the human response directly to the reply_to_topic via hitl_respond.sh — including Kafka topic existence checks and consuming system.notifications.ui.
- **sources:** docs/agent_docs/sql_for_hitl/001_hitl_requests.sql
- **relations:** awaited_requests registry; input_requests; debugging.
- **verify-later:** hitl_respond.sh script location.

<!-- SOURCE: U20_legacy_docs_a.md -->
### HITL approval-as-specialised-agent architecture (human-reviewer plan)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a is a full implementation plan (approval_tasks/approval_versions schema, human-reviewer agent yaml, pull-based REST API, coordinator pause/resume) with estimated effort; the simpler await_approval mechanism is what the later docs002 tests actually exercise.
- **what:** Design: approval handled by spawning a `human-reviewer` agent whose workflow is create_approval_request → wait (StatusPausedForHuman) → process decision → merge_data approved content back over the generating step's output. approval_tasks + approval_versions tables give a full audit trail; clients poll REST endpoints (list/get/approve/reject/upload-url); versioned image paths `/clients/{id}/jobs/{orch}/{generated|user-uploads|approved}/`. Explicit phase-1 scope cuts: no regeneration loops, no multi-reviewer, no auto-approval.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md
- **relations:** the built alternative: await_approval action; content-type approval capabilities; successor: docs011_api_hitl / humanintheloop (hitl category).
- **verify-later:** do approval_tasks/approval_versions tables exist, or only approval_requests.

<!-- SOURCE: U20_legacy_docs_a.md -->
### await_approval Kafka pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 0101: "a solid foundation for HITL already exists… `await_approval` action exists in registry.go"; 0109 records a working end-to-end message pair (resume via responses topic with in_response_to_request_id = approval token).
- **what:** The core HITL primitive: an await_approval workflow step generates an approval token (request_id), publishes a notification (with reply_to and data-to-approve) to `system.notifications.ui`, and returns await_response:true so the SagaCoordinator parks the orchestration. A human/UI resumes it by producing a response whose `in_response_to_request_id` matches the token — initially via `system.commands.workflow.resume`, in working practice via the paused agent's responses topic. Manual kcat testing procedure documented before any UI existed.
- **sources:** docs002_hitl_parallel/README.0101.human_in_the_loop_flow.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0109.hitl_working_message.md; docs002_hitl_parallel/README.0107.hitl_expected_flow.md
- **relations:** SagaCoordinator AwaitedRequests; process_approval_decision; approval_requests table; intake orchestrator HITL gates.
- **verify-later:** actions/hitl_actions.go; system.notifications.ui consumers today (admin dashboard?).

<!-- SOURCE: U20_legacy_docs_a.md -->
### approval_requests table
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** DDL in 0104 (request_id, orchestration_id, data, status, approved_by, comments); 0111 notes storeApprovalRequest "is a stub" and the table lacks timeout fields.
- **what:** Persistence for pending approvals so clients can poll and audits survive: token-keyed rows with status pending/…, approver identity, comments, timestamps. Created but the write path was initially stubbed.
- **sources:** docs002_hitl_parallel/README.0104.hitl_create_db.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** await_approval; HITL timeouts (adds timeout_seconds/timeout_at columns).
- **verify-later:** table existence + whether storeApprovalRequest/updateApprovalRequest are implemented.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Content-type-aware approval capabilities (text edit / image replace)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a capabilities matrix (text: can_edit_directly; image: can_replace with presigned upload_url methods) and buildResponseData branching; 0102c adds ui_config editable_fields.
- **what:** Approvals adapt to content type: text approvals allow inline editing (edited_content replaces llm_output, original preserved), image approvals allow replacement via pre-signed S3 upload or external URL; ui_config hints (title, editable_fields, actions) let a generic UI render each approval correctly.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md; docs002_hitl_parallel/README.0102c.hitl_agent_definitions_3; docs002_hitl_parallel/README.0105.hitl_message_format.md
- **relations:** human-reviewer plan; imagery storage paths.
- **verify-later:** any UI consuming ui_config; presigned upload endpoint.

<!-- SOURCE: U20_legacy_docs_a.md -->
### HITL approval timeouts (config mapping, defaults, restart recovery)
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 0111: "All approval requests currently timeout after 180 seconds… regardless of workflow config, because Step.Timeout is 0" — bug analysis with phased fix plan; 0112 quick-reference states approval default 24h, min 60s, max 7 days.
- **what:** Approval steps carry `timeout_seconds` (up to multi-day) but the value was sent to the UI and never mapped onto Step.Timeout, so the generic 180s DefaultRequestTimeout applied. Fix plan: map config→Step.Timeout at execution, store timeout_at in approval_requests, validate bounds (60s–7d, default 24h), and recover timeout goroutines from AwaitedRequests on pod restart (goroutines are memory-only).
- **sources:** docs002_hitl_parallel/README.0111.hitl_timeouts.md; docs002_hitl_parallel/README.0112.hitl_timeout_testing.md
- **relations:** child-orchestration timeout monitor; approval_requests table.
- **verify-later:** getTimeout(step) mapping today; recoverPendingTimeouts.

<!-- SOURCE: U20_legacy_docs_a.md -->
### process_approval_decision and rejection routing
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Registered action used in all HITL test workflows; config `stop_on_reject`, and Option-3 workflow routes on `process_approval.approved` via conditional_route to finalize/handle_rejection.
- **what:** Post-approval step that unpacks the human's decision (approved flag, comments, modified_data) into CollectedData and lets workflows branch on it — continue, stop, or run a rejection handler path.
- **sources:** docs002_hitl_parallel/README.0106.hitl_multistep_approval.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0102b.hitl_agent_defnitions_2
- **relations:** await_approval; conditional_route/evaluate_condition.
- **verify-later:** actions registry entries process_approval_decision, conditional_route.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HITL request/response protocol (awaited requests, three IDs)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs006/011 "Key Discovery: DO NOT use orchestration's original request_id; DO use HITL action's generated request_id from logs"; docs014/001 documents the same protocol with reply_to_topic and cleanup behaviour.
- **what:** The mechanism by which workflows pause for human input: request_human_input registers an entry in AwaitedRequests with a freshly generated request token; the human response (Kafka message with correlation_id, orchestration_id, in_response_to_request_id, in_response_to_step_name headers) is matched by the coordinator, which removes the awaited entry, stores the data, and resumes. Multiple sequential HITL pauses supported. Notable operational pain: request IDs had to be grepped from pod logs.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#HITL-Message-Requirements; docs014_research_agent/001_human_in_the_loop_response_flow.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL API endpoint; parent-timeout race; system.notifications.ui.
- **verify-later:** awaited_requests table; request_human_input action; coordinator response matching code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HITL API endpoint (/api/v1/hitl/respond)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs011/001 recommends "quick fix now, API endpoint later (2-3 hours)"; docs011/002 is the complete handler implementation guide "For Future Implementation"; the admin dashboard later exposes HITL response UI (per docs024 012 era).
- **what:** An HTTP gateway endpoint that accepts a JSON HITL response (correlation_id, orchestration_id, request_id, step_name, data) and constructs the correctly-headed Kafka response message, replacing fragile hand-built kcat commands (whose immediate bug was unsubstituted `${VAR}` template strings inside single-quoted heredocs).
- **sources:** docs011_api_hitl/001_hitl_api_analysis.md; docs011_api_hitl/002_implementation.md
- **relations:** HITL protocol; admin-dashboard-and-api; system.notifications.ui consumer gap.
- **verify-later:** internal/gateway/hitl_handler.go; route registration in production-api.

<!-- SOURCE: U21_legacy_docs_b.md -->
### system.notifications.ui topic and the missing HITL UI service
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** docs014/001: "It appears there is no service consuming system.notifications.ui to present HITL requests to humans" with input_requests/awaited_requests SQL to check pending items.
- **what:** HITL escalations publish a rich notification (request_type, reply_to_topic, ui_config with title/description/issues_field, timeout, editable flag) to a dedicated UI topic; a consumer service was required to display requests and collect responses. At the time nothing consumed it — the documented alternatives were building the UI service or raising auto-approval thresholds. Later matured into the admin dashboard's HITL surface.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#What-Needs-to-Consume
- **relations:** HITL API endpoint; content-reviewer escalate_to_human; admin-dashboard-and-api.
- **verify-later:** consumers of system.notifications.ui; input_requests table.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### HITL review flow (needs_human_review → retry/resolve/spec-edit)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 008b Block E3 "IMPLEMENTED": `/admin/work-items/:id/retry` converts to `content_rewrite`+reset+re-triage; `/resolve` marks complete with note in `error`; PATCH `/admin/sites/:id/specs/:aspect` supersedes spec. 007b describes the three resolution paths.
- **what:** When `validate_page_content` catches placeholder/contamination it emits a `needs_human_review` item and hides the section (HTML comment). An operator either edits the site spec then retries (regenerates with new data), retries directly (transient), or dismisses (section stays hidden). Retry rewrites to `content_rewrite`/`page-build-handler`, resets `attempt_count`, sets `triaged`.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#hitl-review-flow; 008b#block-e; PLAN_design-note-recommendation-specialists.md
- **relations:** approval_mode per-site/per-item; recommendation-specialist routing; content validation
- **verify-later:** validate_page_content action; site_work_items status machine

<!-- SOURCE: U26_misc_dirs.md -->
### HITL approval pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** HITL_README.md documents the working flow end-to-end with real topics and states; the workflow pauses in AWAITING_RESPONSE and resumes on a message to system.commands.workflow.resume — dated Nov 2025 with a working test kit.
- **what:** Workflows pause on an `await_approval` (earlier `pause_for_human_input`) step: an approval request (title, description, generated content, metadata) is published to `system.notifications.ui` carrying the correlation_id and a request_id that serves as the approval token; a human (or later, an API) publishes a response to `system.commands.workflow.resume` with in_response_to_request_id = token and an approved/comments body; the orchestration resumes, exposing approval status/comments/approver/timestamp to subsequent steps. Timeout configurable (default 300s).
- **sources:** docs/humanintheloop/HITL_README.md; docs/humanintheloop/hitl_agent_definition.sql (await_human_approval step); docs/humanintheloop/send_approval.sh
- **relations:** workflow state machine (AWAITING_RESPONSE); HITL API integration; conditional approval branching
- **verify-later:** await_approval action code; resume command consumer; docs011_api_hitl for the API successor

<!-- SOURCE: U26_misc_dirs.md -->
### HITL content-approval demo agent and group
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Complete SQL seeds (agent + group, versioned, ON CONFLICT upsert) dated 2025-11-03, referencing chassis image v1.0.407 — a working, loadable demo.
- **what:** `simple-content-writer-with-approval` agent: generate_draft (execute_llm_prompt, Claude 3.5 Sonnet) → await_human_approval → process_approval (merges content with approval metadata) → complete. Wrapped by the `content-approval-hitl` group whose orchestration spawns the writer, calls it with business input data, and aggregates results. The canonical minimal HITL example for the platform.
- **sources:** docs/humanintheloop/hitl_agent_definition.sql; docs/humanintheloop/hitl_agent_group_definition.sql
- **relations:** HITL approval mechanism; agent groups; execute_llm_prompt
- **verify-later:** whether these definitions are loaded in current DB

<!-- SOURCE: U26_misc_dirs.md -->
### HITL kcat test harness
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Six working shell scripts (listen, start, send approval, monitor, quick one-liner, verify setup) using kubectl-run kcat pods against the personae namespace; README walks the three-terminal procedure.
- **what:** Manual test kit for the HITL loop: listen_for_approvals.sh tails system.notifications.ui; start_hitl_workflow.sh / quick_hitl_test.sh publish an orchestrate request for the content-approval-hitl group with full headers; send_approval.sh publishes the resume message with the approval token; monitor_workflow.sh polls orchestrator_state (status, current step, awaited_requests, collected approval data); verify_hitl_setup.sh checks definitions, topics and chassis pods exist.
- **sources:** docs/humanintheloop/HITL_README.md#testing-the-hitl-flow; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/verify_hitl_setup.sh
- **relations:** HITL approval mechanism; kcat/db-inspector ops runbook
- **verify-later:** script topic names against current cluster topics

<!-- SOURCE: U26_misc_dirs.md -->
### Conditional approval branching
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#customization shows a `conditional_branch` step keyed on `{{.await_human_approval.approved}}` routing to finalize vs regenerate — offered as a customisation, not present in the shipped demo workflow.
- **what:** Approval outcomes drive workflow branching: approved → finalise; rejected → regenerate content and re-submit for approval, enabling iterative human-guided refinement loops rather than binary pass/fail.
- **sources:** docs/humanintheloop/HITL_README.md#conditional-approval
- **relations:** HITL approval mechanism; dynamic prompt improvement loop (flag-for-improvement variant)
- **verify-later:** conditional_branch action existence

<!-- SOURCE: U26_misc_dirs.md -->
### HITL API integration (approvals via REST/UI)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#api-integration-future: "Once the API is ready, the manual approval process can be replaced with: REST endpoint to fetch pending approvals, Web UI..., API call to approve/reject with the approval token."
- **what:** Planned replacement of manual Kafka approval messages with a REST surface and web UI over the identical underlying mechanism (same topics, same tokens). The taxonomy's hitl category (docs011_api_hitl) suggests this was subsequently pursued — this doc records the origin point.
- **sources:** docs/humanintheloop/HITL_README.md#api-integration-future
- **relations:** HITL approval mechanism; docs011_api_hitl (likely successor, other unit)
- **verify-later:** docs011_api_hitl extraction; approval endpoints in API code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Approval model (auto / hitl / eval) with four override levels
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** P1: "Future — column exists from day one" for eval; no deployment evidence
- **what:** site_work_items.approval_mode: auto (default), hitl (pending_review → human approves), eval (evaluation agent reviews handler output before completion). Overrides per-item, per-item-type, per-site, system default.
- **sources:** P1#Approval Model
- **relations:** content-reviewer as future eval agent; P10 recommendation specialists
- **verify-later:** approval_mode column exists?

<!-- SOURCE: U15_docs019_running_notes.md -->
### Governance/HITL confirm-not-initiate model
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** "Confirm-not-initiate. Decision reasoning is agent-led; the human confirms via a decision package" (principles(59) §Governance and HITL).
- **what:** A framing where every decision publishes its reasoning (not just its outcome) so drift detection is possible; a decision requiring gating routes through exactly ONE confirming component (never reimplemented per producer); a newer proposal for an already-pending target supersedes and expires the older one (freshness over queue order); and inheritance has two precedence directions — normal entries are child-wins, sealed constraints (legal floors, mission non-negotiables) are ancestor-wins.
- **sources:** NOTES_running_synthesis_principles(59) §Governance and HITL (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; diagnosis loop (embodies read-only + human-gated in practice); council hard_veto flag model.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Confirm-not-initiate HITL model (decision package)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §5 "Rules are not authored by humans from scratch and are not changed unilaterally by agents. The decision's reasoning is agent-led; the human confirms"
- **what:** Agents carry the analysis and produce a decision package (summary, tradeoffs/impact, genuine choices incl. reject and defer); the human confirms an informed choice, with rigor scaled to stakes. Acceptance writes a new version and deprecates the old (never deletes).
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#5, ED/FOCUS_best_practice_doc_tree(1).md#4.2, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** concern curators; coordinator; governance gate; deprecate-not-delete
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### User-representative advocate (intent + conflict triage)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §8 "A standing agent … representing the user sits inside the framing process as the check on the coordinator's framing power"
- **what:** A standing advocate for the user's intent that checks the coordinator's framing. Its signature job is triaging claimed conflicts before any reach the user: dissolve illusory or already-reconciled tensions, ask a clarifying question when unsure a conflict is real, and escalate only genuine tradeoffs.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#8, ED/FOCUS_standards_curation_and_governance(1).md#8.2, ED/FOCUS_standards_curation_and_governance(1).md#8.4
- **relations:** coordinator; decision authority; concern curators
- **verify-later:** intake-orchestrator; briefing-agent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Decision authority (co-equal voices, abstention, creator veto)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** FOCUS_standards_curation(1) §9 "Default: co-equal voices in the frame … Abstention fallback: the advocate decides, bounded … Creator override / veto"
- **what:** When advocate and curator genuinely disagree, both cases go to the human as co-equal voices. If the user declines to choose, the advocate decides bounded by codified intent — but high-stakes abstention and blocker conflicts escalate to the creator, who holds veto or optional final choice. Three distinct human roles: user, confirmer, creator.
- **sources:** ED/FOCUS_standards_curation_and_governance(1).md#9
- **relations:** user-representative advocate; coordinator; confirm-not-initiate
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Human direction channels + lock lifecycle + audit-pass cap
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 007 "Humans influence sites at three levels … Lock types: permanent / timed / review"; "The 3-pass audit cap prevents unbounded improvement cycles"
- **what:** Humans steer sites via three channels. HITL-requested content is protected by lock types (permanent, timed-90d, review-creates-HITL-item-on-expiry). A 3-pass audit cap bounds improvement cycles and resets on time/direction-change/major-rebuild/manual.
- **sources:** WM/007_adoption_pipeline_v3.md#human-direction, WM/007_adoption_pipeline_v3.md#hitl-requested-content-and-lock-lifecycle, WM/007_adoption_pipeline_v3.md#audit-pass-cap-reset
- **relations:** locks (031); improvement loop; direction spec
- **verify-later:** site_specs direction aspect; lock_type/lock_expires_at; last_audit_reset_at

<!-- SOURCE: U18_sql_for_agents.md -->
### intake-orchestrator (classify → brief → HITL confirm → spawn builder → rerender)
- **category:** hitl
- **status-signal:** superseded
- **status-evidence:** 068_domain_submitter_agent.sql creates a new entry point ("Entry point for new domain submissions... creates needs_domain_research work item"); intake files stop being patched after 030-era; 002 header shows HITL steps with `skip_if: input_data.hitl_mode == auto`.
- **what:** The v1/v2 entry pipeline: discovers available `%-builder` agents, spawns site-classifier and briefing-agent, runs two human-in-the-loop gates (confirm site type; review brief), spawns the recommended builder, then (added later) a rerender pass for nav consistency. Notable mechanisms: `dynamic_select` HITL fields fed from a live agent query; `skip_if` auto mode making HITL optional per run.
- **sources:** 002_intake_orchestrator.sql; sql_for_agents_v1/001_agent_definitions_etc.sql; sql_for_agents_v2/002_intake_orchestrator.sql
- **relations:** superseded by domain-submitter + build-dispatch-loop; HITL gate pattern survives in content-reviewer HITL mode
- **verify-later:** agent_definitions 'intake-orchestrator' status; whether any trigger path still uses it

<!-- SOURCE: U18_sql_for_agents.md -->
### content-reviewer (HITL + auto-eval dual mode with pre-validation)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 025 adds validate_content step "runs BEFORE review mode determination"; agent defined active in v2/025.
- **what:** Reviews generated page content in either human (request_human_input, editable) or auto-eval (LLM approve/flag) mode, selected at runtime. 025 adds algorithmic pre-validation: internal links must point to existing site pages; email addresses must match the site's contact_email — findings are handed to whichever review mode runs.
- **sources:** 025_content_reviewer_agent.sql; sql_for_agents_v2/025_content_reviewer_agent.sql
- **relations:** page-content-writer output consumer; HITL pattern from intake-orchestrator
- **verify-later:** validate_page_content action; current review-mode default

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item approval_mode (auto / hitl / eval)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Column added in Phase-0 Block A: "Controls whether items auto-dispatch or require human/eval approval. Values: 'auto' (default), 'hitl', 'eval'".
- **what:** Per-item gating between triage and dispatch: auto items flow straight through; hitl items wait for human approval; eval items wait for an evaluation agent. The schema-level hook for configurable human review gates in the build pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#4
- **relations:** work queue; input_requests; human change requests.
- **verify-later:** dispatch respects approval_mode; any 'eval' users.

<!-- SOURCE: U19_sql_tables_components.md -->
### Human change-request work items
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Applied INSERT batches for gaswholesalers.com and finetuning.uk: needs_design→webdesign-agent, content_edit→section-editor (field_updates spec), needs_logo/needs_hero_image→image-build-handler (image_prompts spec), needs_rerender→rerender-pages at priority 99.
- **what:** Human requests enter the same queue as agent-detected work: source='human', item_key 'human_<what>_<site>', pre-triaged, priority-ordered so content and imagery land before the final rerender. The spec JSONB is the full handler contract (edit_type/page_name/slot_name/field_updates for edits; purpose + image_prompts for imagery).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** work queue; section-editor; image-build-handler; dispatch loop (071).
- **verify-later:** admin UI path creating these; section-editor field_updates handling.

<!-- SOURCE: U19_sql_tables_components.md -->
### input_requests HITL persistence
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Full schema with pending view, expiry function and debugging queries; FK to orchestration_states with CASCADE.
- **what:** Human input requests persisted for querying and UI display: request_type (review/confirmation/questionnaire), title/message/data/ui_config, reply_to_topic for Kafka response routing, timeout/expires_at, status lifecycle pending→completed/expired/cancelled, and the response payload with responder identity. pending_input_requests view feeds the UI with seconds_remaining.
- **sources:** docs/agent_docs/sql_for_tables/006_input_requests.sql
- **relations:** awaited_requests (transport-level counterpart); approval_mode; admin dashboard.
- **verify-later:** UI consumption; expire_input_requests scheduling.

<!-- SOURCE: U19_sql_tables_components.md -->
### Manual HITL continuation runbook
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Working queries against awaited_requests for stuck escalate_to_human steps, plus the documented hitl_respond.sh Kafka invocation with real ids/topics.
- **what:** Operational procedure for un-sticking HITL flows: locate the awaited request (step_name LIKE %human%/%hitl%/%approval%), optionally reset an expired one, then publish the human response directly to the reply_to_topic via hitl_respond.sh — including Kafka topic existence checks and consuming system.notifications.ui.
- **sources:** docs/agent_docs/sql_for_hitl/001_hitl_requests.sql
- **relations:** awaited_requests registry; input_requests; debugging.
- **verify-later:** hitl_respond.sh script location.

<!-- SOURCE: U20_legacy_docs_a.md -->
### HITL approval-as-specialised-agent architecture (human-reviewer plan)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a is a full implementation plan (approval_tasks/approval_versions schema, human-reviewer agent yaml, pull-based REST API, coordinator pause/resume) with estimated effort; the simpler await_approval mechanism is what the later docs002 tests actually exercise.
- **what:** Design: approval handled by spawning a `human-reviewer` agent whose workflow is create_approval_request → wait (StatusPausedForHuman) → process decision → merge_data approved content back over the generating step's output. approval_tasks + approval_versions tables give a full audit trail; clients poll REST endpoints (list/get/approve/reject/upload-url); versioned image paths `/clients/{id}/jobs/{orch}/{generated|user-uploads|approved}/`. Explicit phase-1 scope cuts: no regeneration loops, no multi-reviewer, no auto-approval.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md
- **relations:** the built alternative: await_approval action; content-type approval capabilities; successor: docs011_api_hitl / humanintheloop (hitl category).
- **verify-later:** do approval_tasks/approval_versions tables exist, or only approval_requests.

<!-- SOURCE: U20_legacy_docs_a.md -->
### await_approval Kafka pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 0101: "a solid foundation for HITL already exists… `await_approval` action exists in registry.go"; 0109 records a working end-to-end message pair (resume via responses topic with in_response_to_request_id = approval token).
- **what:** The core HITL primitive: an await_approval workflow step generates an approval token (request_id), publishes a notification (with reply_to and data-to-approve) to `system.notifications.ui`, and returns await_response:true so the SagaCoordinator parks the orchestration. A human/UI resumes it by producing a response whose `in_response_to_request_id` matches the token — initially via `system.commands.workflow.resume`, in working practice via the paused agent's responses topic. Manual kcat testing procedure documented before any UI existed.
- **sources:** docs002_hitl_parallel/README.0101.human_in_the_loop_flow.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0109.hitl_working_message.md; docs002_hitl_parallel/README.0107.hitl_expected_flow.md
- **relations:** SagaCoordinator AwaitedRequests; process_approval_decision; approval_requests table; intake orchestrator HITL gates.
- **verify-later:** actions/hitl_actions.go; system.notifications.ui consumers today (admin dashboard?).

<!-- SOURCE: U20_legacy_docs_a.md -->
### approval_requests table
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** DDL in 0104 (request_id, orchestration_id, data, status, approved_by, comments); 0111 notes storeApprovalRequest "is a stub" and the table lacks timeout fields.
- **what:** Persistence for pending approvals so clients can poll and audits survive: token-keyed rows with status pending/…, approver identity, comments, timestamps. Created but the write path was initially stubbed.
- **sources:** docs002_hitl_parallel/README.0104.hitl_create_db.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** await_approval; HITL timeouts (adds timeout_seconds/timeout_at columns).
- **verify-later:** table existence + whether storeApprovalRequest/updateApprovalRequest are implemented.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Content-type-aware approval capabilities (text edit / image replace)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a capabilities matrix (text: can_edit_directly; image: can_replace with presigned upload_url methods) and buildResponseData branching; 0102c adds ui_config editable_fields.
- **what:** Approvals adapt to content type: text approvals allow inline editing (edited_content replaces llm_output, original preserved), image approvals allow replacement via pre-signed S3 upload or external URL; ui_config hints (title, editable_fields, actions) let a generic UI render each approval correctly.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md; docs002_hitl_parallel/README.0102c.hitl_agent_definitions_3; docs002_hitl_parallel/README.0105.hitl_message_format.md
- **relations:** human-reviewer plan; imagery storage paths.
- **verify-later:** any UI consuming ui_config; presigned upload endpoint.

<!-- SOURCE: U20_legacy_docs_a.md -->
### HITL approval timeouts (config mapping, defaults, restart recovery)
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 0111: "All approval requests currently timeout after 180 seconds… regardless of workflow config, because Step.Timeout is 0" — bug analysis with phased fix plan; 0112 quick-reference states approval default 24h, min 60s, max 7 days.
- **what:** Approval steps carry `timeout_seconds` (up to multi-day) but the value was sent to the UI and never mapped onto Step.Timeout, so the generic 180s DefaultRequestTimeout applied. Fix plan: map config→Step.Timeout at execution, store timeout_at in approval_requests, validate bounds (60s–7d, default 24h), and recover timeout goroutines from AwaitedRequests on pod restart (goroutines are memory-only).
- **sources:** docs002_hitl_parallel/README.0111.hitl_timeouts.md; docs002_hitl_parallel/README.0112.hitl_timeout_testing.md
- **relations:** child-orchestration timeout monitor; approval_requests table.
- **verify-later:** getTimeout(step) mapping today; recoverPendingTimeouts.

<!-- SOURCE: U20_legacy_docs_a.md -->
### process_approval_decision and rejection routing
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Registered action used in all HITL test workflows; config `stop_on_reject`, and Option-3 workflow routes on `process_approval.approved` via conditional_route to finalize/handle_rejection.
- **what:** Post-approval step that unpacks the human's decision (approved flag, comments, modified_data) into CollectedData and lets workflows branch on it — continue, stop, or run a rejection handler path.
- **sources:** docs002_hitl_parallel/README.0106.hitl_multistep_approval.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0102b.hitl_agent_defnitions_2
- **relations:** await_approval; conditional_route/evaluate_condition.
- **verify-later:** actions registry entries process_approval_decision, conditional_route.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HITL request/response protocol (awaited requests, three IDs)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs006/011 "Key Discovery: DO NOT use orchestration's original request_id; DO use HITL action's generated request_id from logs"; docs014/001 documents the same protocol with reply_to_topic and cleanup behaviour.
- **what:** The mechanism by which workflows pause for human input: request_human_input registers an entry in AwaitedRequests with a freshly generated request token; the human response (Kafka message with correlation_id, orchestration_id, in_response_to_request_id, in_response_to_step_name headers) is matched by the coordinator, which removes the awaited entry, stores the data, and resumes. Multiple sequential HITL pauses supported. Notable operational pain: request IDs had to be grepped from pod logs.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#HITL-Message-Requirements; docs014_research_agent/001_human_in_the_loop_response_flow.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL API endpoint; parent-timeout race; system.notifications.ui.
- **verify-later:** awaited_requests table; request_human_input action; coordinator response matching code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### HITL API endpoint (/api/v1/hitl/respond)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** docs011/001 recommends "quick fix now, API endpoint later (2-3 hours)"; docs011/002 is the complete handler implementation guide "For Future Implementation"; the admin dashboard later exposes HITL response UI (per docs024 012 era).
- **what:** An HTTP gateway endpoint that accepts a JSON HITL response (correlation_id, orchestration_id, request_id, step_name, data) and constructs the correctly-headed Kafka response message, replacing fragile hand-built kcat commands (whose immediate bug was unsubstituted `${VAR}` template strings inside single-quoted heredocs).
- **sources:** docs011_api_hitl/001_hitl_api_analysis.md; docs011_api_hitl/002_implementation.md
- **relations:** HITL protocol; admin-dashboard-and-api; system.notifications.ui consumer gap.
- **verify-later:** internal/gateway/hitl_handler.go; route registration in production-api.

<!-- SOURCE: U21_legacy_docs_b.md -->
### system.notifications.ui topic and the missing HITL UI service
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** docs014/001: "It appears there is no service consuming system.notifications.ui to present HITL requests to humans" with input_requests/awaited_requests SQL to check pending items.
- **what:** HITL escalations publish a rich notification (request_type, reply_to_topic, ui_config with title/description/issues_field, timeout, editable flag) to a dedicated UI topic; a consumer service was required to display requests and collect responses. At the time nothing consumed it — the documented alternatives were building the UI service or raising auto-approval thresholds. Later matured into the admin dashboard's HITL surface.
- **sources:** docs014_research_agent/001_human_in_the_loop_response_flow.md#What-Needs-to-Consume
- **relations:** HITL API endpoint; content-reviewer escalate_to_human; admin-dashboard-and-api.
- **verify-later:** consumers of system.notifications.ui; input_requests table.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### HITL review flow (needs_human_review → retry/resolve/spec-edit)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 008b Block E3 "IMPLEMENTED": `/admin/work-items/:id/retry` converts to `content_rewrite`+reset+re-triage; `/resolve` marks complete with note in `error`; PATCH `/admin/sites/:id/specs/:aspect` supersedes spec. 007b describes the three resolution paths.
- **what:** When `validate_page_content` catches placeholder/contamination it emits a `needs_human_review` item and hides the section (HTML comment). An operator either edits the site spec then retries (regenerates with new data), retries directly (transient), or dismisses (section stays hidden). Retry rewrites to `content_rewrite`/`page-build-handler`, resets `attempt_count`, sets `triaged`.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#hitl-review-flow; 008b#block-e; PLAN_design-note-recommendation-specialists.md
- **relations:** approval_mode per-site/per-item; recommendation-specialist routing; content validation
- **verify-later:** validate_page_content action; site_work_items status machine

<!-- SOURCE: U26_misc_dirs.md -->
### HITL approval pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** HITL_README.md documents the working flow end-to-end with real topics and states; the workflow pauses in AWAITING_RESPONSE and resumes on a message to system.commands.workflow.resume — dated Nov 2025 with a working test kit.
- **what:** Workflows pause on an `await_approval` (earlier `pause_for_human_input`) step: an approval request (title, description, generated content, metadata) is published to `system.notifications.ui` carrying the correlation_id and a request_id that serves as the approval token; a human (or later, an API) publishes a response to `system.commands.workflow.resume` with in_response_to_request_id = token and an approved/comments body; the orchestration resumes, exposing approval status/comments/approver/timestamp to subsequent steps. Timeout configurable (default 300s).
- **sources:** docs/humanintheloop/HITL_README.md; docs/humanintheloop/hitl_agent_definition.sql (await_human_approval step); docs/humanintheloop/send_approval.sh
- **relations:** workflow state machine (AWAITING_RESPONSE); HITL API integration; conditional approval branching
- **verify-later:** await_approval action code; resume command consumer; docs011_api_hitl for the API successor

<!-- SOURCE: U26_misc_dirs.md -->
### HITL content-approval demo agent and group
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Complete SQL seeds (agent + group, versioned, ON CONFLICT upsert) dated 2025-11-03, referencing chassis image v1.0.407 — a working, loadable demo.
- **what:** `simple-content-writer-with-approval` agent: generate_draft (execute_llm_prompt, Claude 3.5 Sonnet) → await_human_approval → process_approval (merges content with approval metadata) → complete. Wrapped by the `content-approval-hitl` group whose orchestration spawns the writer, calls it with business input data, and aggregates results. The canonical minimal HITL example for the platform.
- **sources:** docs/humanintheloop/hitl_agent_definition.sql; docs/humanintheloop/hitl_agent_group_definition.sql
- **relations:** HITL approval mechanism; agent groups; execute_llm_prompt
- **verify-later:** whether these definitions are loaded in current DB

<!-- SOURCE: U26_misc_dirs.md -->
### HITL kcat test harness
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Six working shell scripts (listen, start, send approval, monitor, quick one-liner, verify setup) using kubectl-run kcat pods against the personae namespace; README walks the three-terminal procedure.
- **what:** Manual test kit for the HITL loop: listen_for_approvals.sh tails system.notifications.ui; start_hitl_workflow.sh / quick_hitl_test.sh publish an orchestrate request for the content-approval-hitl group with full headers; send_approval.sh publishes the resume message with the approval token; monitor_workflow.sh polls orchestrator_state (status, current step, awaited_requests, collected approval data); verify_hitl_setup.sh checks definitions, topics and chassis pods exist.
- **sources:** docs/humanintheloop/HITL_README.md#testing-the-hitl-flow; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/verify_hitl_setup.sh
- **relations:** HITL approval mechanism; kcat/db-inspector ops runbook
- **verify-later:** script topic names against current cluster topics

<!-- SOURCE: U26_misc_dirs.md -->
### Conditional approval branching
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#customization shows a `conditional_branch` step keyed on `{{.await_human_approval.approved}}` routing to finalize vs regenerate — offered as a customisation, not present in the shipped demo workflow.
- **what:** Approval outcomes drive workflow branching: approved → finalise; rejected → regenerate content and re-submit for approval, enabling iterative human-guided refinement loops rather than binary pass/fail.
- **sources:** docs/humanintheloop/HITL_README.md#conditional-approval
- **relations:** HITL approval mechanism; dynamic prompt improvement loop (flag-for-improvement variant)
- **verify-later:** conditional_branch action existence

<!-- SOURCE: U26_misc_dirs.md -->
### HITL API integration (approvals via REST/UI)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#api-integration-future: "Once the API is ready, the manual approval process can be replaced with: REST endpoint to fetch pending approvals, Web UI..., API call to approve/reject with the approval token."
- **what:** Planned replacement of manual Kafka approval messages with a REST surface and web UI over the identical underlying mechanism (same topics, same tokens). The taxonomy's hitl category (docs011_api_hitl) suggests this was subsequently pursued — this doc records the origin point.
- **sources:** docs/humanintheloop/HITL_README.md#api-integration-future
- **relations:** HITL approval mechanism; docs011_api_hitl (likely successor, other unit)
- **verify-later:** docs011_api_hitl extraction; approval endpoints in API code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Domain submission tiers and mission/roadmap briefs
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(5) documents current domain-submitter workflow with persist steps
- **what:** domain-submitter is the entry point for new builds. Three tiers: domain only; domain+objective hint; domain+mission/roadmap (structured JSON for machine consumers + plain-text `mission_brief`/`roadmap_brief` that classifier/planner actually read). Persist steps skip gracefully via error_step when fields absent; briefs must be plain text parseable by small models.
- **sources:** 001(5)#Domain Submission; 007#Mission-Driven Sites
- **relations:** classifier weighting of inputs (028); vonc/Spark pattern
- **verify-later:** domain-submitter agent definition; site_specs mission aspects

<!-- SOURCE: U01_docs024_numbered_core.md -->
### build_queue domain queue with direction spectrum
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** P1 (marked "a bit out of date but still has merit"); P2 depends on it for POST /sites; seed_build_queue named in 032/other docs as real
- **what:** build_queue rows (domain, direction jsonb, status, batch, priority); direction spans null → objective hint → full brief (skip research+briefing) → adopt_from → fork_from (specs pre-populated). seed_build_queue takes N, ensures site records, writes initial specs, inserts the appropriate first work item; pacing by batch size. Initial chain: needs_domain_research → needs_briefing → needs_site_plan with spec outputs per handler.
- **sources:** P1#Domain Queue, #Initial Build
- **relations:** public API POST /sites; domain-submitter (newer entry path — reconcile)
- **verify-later:** build_queue table + seed action exist; relation to domain-submitter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Onboarding/config three-layer model
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** "The config has three layers with different derivability: mechanical (discovered + probed), conventions (inferred or doc-sourced — confirmed), intent (elicited)." (principles(59) §Onboarding and config).
- **what:** A framing that treats tenant/codebase onboarding as three separate problems with different confirmation mechanisms and different climb rates on the trust ratchet: mechanical facts (probed, confirmable by reality, climb fastest), conventions (inferred-then-confirmed even in docs-authoritative mode, since hallucinated conventions would manufacture drift), and intent (elicited progressively, never "done," captured just-in-time as work happens rather than as an upfront tax).
- **sources:** NOTES_running_synthesis_principles(59) §Onboarding and config (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; doc claim-verification convention (shares "inferred-then-confirmed" DNA).

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three-layer config: mechanical / conventions / intent (different derivability)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation "Status: plan"; nothing in these files claims implementation.
- **what:** Onboarding is three processes, not one: the mechanical layer is discoverable (inspect + probe; low stakes, confirmable by reality); conventions are inferred or doc-sourced (a strong draft, weak authority — code shows what it does, not what it should do); intent and standards are elicited (not derivable from source; the tenant is the source, and the part delivering the tool's distinctive value).
- **sources:** PLAN_onboarding_config_derivation.md#1; 001_onboarding_discussion.txt
- **relations:** the five onboarding agents; docs-authoritative decision
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Progressive onboarding — a ramp, never "done"
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §2 — plan.
- **what:** A tenant gets value from the mechanical layer alone (fresh code context, signatures, reuse search, schema) before any intent is captured; conventions and intent fill incrementally and the tool deepens as they arrive. Onboarding tracks the repo forever — active-with-pending is the steady state, and leaf-level intent is captured just-in-time during use rather than as a setup tax.
- **sources:** PLAN_onboarding_config_derivation.md#2; PLAN_onboarding_agent_specs(6).md#3.7,#4.3
- **relations:** intent-elicitation agent; config-maintenance agent
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Config as a maintained artifact (the wizard is the first pass; the lifecycle is the deliverable)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §3 — plan.
- **what:** The derived config drifts as the repo changes, so it gets the standards' own upkeep machinery: periodic re-derivation with divergence flagging, confirm-not-initiate on proposed changes, and per-entry provenance (discovered/inferred/supplied) determining trust and change authority. "Onboarding as a first-class deliverable" means this lifecycle, not a good setup script.
- **sources:** PLAN_onboarding_config_derivation.md#3; 001_onboarding_discussion.txt
- **relations:** config-maintenance agent; active-config provenance shape
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Inference quality scales with codebase quality — surface uncertainty
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §4 — named tension, design mitigation.
- **what:** On a messy repo, convention inference confidently drafts the repo's bad habits, and confirming that codifies the mess — so the more a tenant needs the tool, the less their repo can teach it. Mitigation: surface inconsistency as questions to resolve, never a silent majority pick; inconsistency found during onboarding is itself valuable output ("your conventions aren't actually conventions").
- **sources:** PLAN_onboarding_config_derivation.md#4; 001_onboarding_discussion.txt
- **relations:** conventions agent; docs-authoritative mode
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Docs-authoritative conventions for our own repo (the free drift audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §5 "Ours (decided): docs-authoritative" — the decision is recorded; the audit has not run.
- **what:** Source-of-truth for conventions is chosen per tenant by doc availability; for our repo, 001/003/the naming FOCUS are authoritative and code is read only to find disagreements. Each disagreement is recorded, not silently resolved — the set is a free audit of where the codebase drifted from its own documented standards, the drift detector's first run, on us. Our own onboarding is the template, not a special case.
- **sources:** PLAN_onboarding_config_derivation.md#5,#7; 001_onboarding_discussion.txt
- **relations:** conventions agent; drift audit three-bucket output
- **verify-later:** whether any drift audit ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### Conventions agent (extract-cite-confirm, then audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1 — spec only.
- **what:** Owns the conventions layer: extracts discrete convention atoms from the standards docs (each citing its exact doc span — extraction is inferred-then-confirmed, because auditing code against an invented convention manufactures fake drift, the one failure that would discredit the audit), gets the set human-confirmed BEFORE any audit, then checks code and records disagreements with location/convention/tier/confidence and a default disposition (code-drifted, doc-drifted, or legitimate exception — human confirms). Accepted exceptions are remembered so audits become incremental.
- **sources:** PLAN_onboarding_agent_specs(6).md#1
- **relations:** three checking tiers; docs-authoritative decision; check_*.go validators
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three checking tiers + three-bucket audit output (coverage honesty)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 — spec; the pattern recurs in maintenance §5.5.
- **what:** Conventions (and drift) are checked at three tiers: deterministic (static check settles it → violations), heuristic proxy (a measurable indicator flags candidates, not violations — "where to look, not what's wrong"; an optional LLM pass is still only a candidate flag, never a verdict), judgement-only (no proxy → reported as a coverage gap). The audit reports three numbers, never one — a clean tier-1 count beside many unchecked tier-3 conventions is a partial audit with known limits, and must say so. Companion role split: un-auditable conventions still serve as generation guidance (an atom can be audited, guiding, or both).
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#1.6; PLAN_onboarding_agent_specs(6).md#5.5
- **relations:** conventions agent; config-maintenance drift tiers; LLM-as-candidate principle
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Convention coverage IS capability reliability
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 closing — spec insight.
- **what:** When a bundle capability rests on a manual convention (log-correlation needs orchestration_id in every log line), the capability is only as reliable as the convention's coverage, not its existence. For any capability-bearing convention the audit reports how completely it is followed, and gaps surface as fixable (add the missing log statements) rather than hard limits — even on our own codebase, where the structure exists but coverage is unverified.
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#2.9
- **relations:** codebase-conditional capabilities; runtime evidence by orchestration_id
- **verify-later:** an orchestration_id logging coverage scan

<!-- SOURCE: U16_docs019_design_plans.md -->
### Stack-discovery agent (inspect → interpret → declared probe plan → probe → confirm)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2 — spec only.
- **what:** Owns the mechanical layer: read-only inspection emits facts; interpretation ("this Makefile target is probably the test command") emits proposals with confidence — the subtle point being that interpretation has uncertainty even at the mechanical layer; a declared probe plan (the security contract, kept even for our own use as audit) precedes sandboxed probes; probe results update confidence. A failing build is useful output, candidate-only interpreted, never fixed by this agent. The output document carries per-entry source/confidence/probe-result with uncertainties listed separately. Also records the structural facts bundle capabilities depend on (§2.9).
- **sources:** PLAN_onboarding_agent_specs(6).md#2
- **relations:** confirmation by reality; sandboxing envelope; codebase-conditional capabilities
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirmation by reality (the mechanical layer climbs the ratchet first)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.8; contract-set review §3.1 records the reconciliation.
- **what:** The mechanical layer can be confirmed by observation (the probed command actually works) — the strongest confirmation any config layer carries — so stack-discovery is the natural first capability to graduate past confirm_every. Reconciled with the gate: probe success is initially strong evidence inside the work-item gate (near-rubber-stamp, human still activates); only after trust-ledger graduation does probe success auto-activate. The gate is the starting position; graduation relaxes it — not a bypass.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.8; FOCUS_contract_set_review.md#3.1; PLAN_active_config_schema(3).md#5
- **relations:** trust ledger; confirm-not-initiate
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Sandboxed probing — the tenant-code security envelope
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.6: "gating for the service phase: no tenant code runs until sandboxing is solid."
- **what:** The first agent that may execute tenant code does so inside an ephemeral sandbox: repo mounted read-only, restricted network, time limit, no persistent state; the emitted probe plan is the contract the sandbox approves/restricts/denies per command. The Tier-C security concern made concrete; same gate applies to Phase-2 verification running tenant code.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.6; PLAN_context_assembly_tool_and_service(2).md#6
- **relations:** stack-discovery; service phase
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Intent-elicitation agent (progressive, value-returning interview)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §3 — spec only; reuse target (briefing_questionnaire column) verified real in FOCUS_schema_verification_findings §3.
- **what:** Captures the why-chain, per-node priority profiles and direction-of-travel via an interview that interleaves proposal-confirmation (where evidence exists — low friction, anchoring risk mitigated by citing the evidence so proposals are contestable) with free elicitation (blank page, unavoidable). Every exchange returns value (the captured piece changes the next bundle/mediation); the interview is not finite — leaf intent is captured just-in-time in the flow of work. A descendant of the briefing questionnaire / intake orchestrator, pointed at a codebase. Capture and use are separate roles (the user-rep advocate consumes what this captures). Open: detecting rubber-stamping.
- **sources:** PLAN_onboarding_agent_specs(6).md#3; FOCUS_schema_verification_findings.md#3
- **relations:** onboarding orchestrator; objectives table; user-rep advocate (salience doc, other unit)
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Onboarding orchestrator (dependency-graph flow; active-with-pending)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §4 — spec only.
- **what:** Coordinates the three layer agents: stack-discovery first (both others depend on its mechanical config), conventions and intent in parallel (independent of each other) — sequencing follows dependencies, not policy. Routes all proposals through confirm-not-initiate; surfaces a compact onboarding-state artifact (per-layer confirmed/partial/blocked, pending, drift-audit counts); a blocked layer doesn't stop the others; a tenant walking away pauses cleanly. Terminal state is active-with-pending, handing over to maintenance — never "fully done".
- **sources:** PLAN_onboarding_agent_specs(6).md#4; FOCUS_onboarding_system_view_check.md#1,#3.4
- **relations:** the three layer agents; config-maintenance handoff; work-items queue
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Config-maintenance agent (drift detection as the trust ratchet's signal source)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §5 — spec only.
- **what:** After baseline, detects drift across all three layers, event-driven (change-layer diffs) plus a periodic sweep, dispatching to the layer agents for re-checks rather than reimplementing them; targeted re-validation (implicated-only recheck) instead of full sweeps. Drift evidence uses the same three tiers; surfacing is prioritised to avoid alert fatigue (high-impact deterministic first, heuristic in paced batches, freshness nudges background). Its deeper role: sustained no-drift is graduation evidence and repeated drift is de-graduation evidence — without this agent the bidirectional ratchet has nothing to act on at the right timescale.
- **sources:** PLAN_onboarding_agent_specs(6).md#5; FOCUS_onboarding_system_view_check.md#2
- **relations:** trust ledger; change-layer integration; published-reasoning gap detection
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Active-config schema (four tables, computed-on-read effective values)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_active_config_schema(3) "Status: contract specification"; corrected to chassis conventions after schema verification.
- **what:** The load-bearing contract: tenant_configs (scope-holder row per tenant, created directly at init — not a gate violation), mechanical_config (one JSONB row, per-field embedded provenance), standards (flat concern atoms with scope constitution/domain/leaf, applies_to change types, rule/rationale/check/check_kind), objectives (nested why-chain nodes with priority_profile, direction_of_travel, standing_concerns). A common provenance shape (source/source_ref/confidence/status/last_verified_at/verified_by/freshness_until/version/previous_version_id/deleted_at) across all layers so consumers reason uniformly. Effective priority profile is computed at read time by walking root→node (store authored differences, compute effective on read); acyclicity must be enforced on write AND the walk bounded, since a human can confirm a cycle. The constitution is a view over standards WHERE scope=constitution, not a table. Two atom trees deliberately kept distinct: flat concern tree vs nested objective tree.
- **sources:** PLAN_active_config_schema(3).md; FOCUS_onboarding_system_view_check.md#3.1,#3.7; FOCUS_pre_build_edge_cases(1).md#1.1
- **relations:** all six contracts hang off it; bundle authored layer reads it
- **verify-later:** whether any of the four tables exist in clients_db

<!-- SOURCE: U16_docs019_design_plans.md -->
### Governed vocabularies and the hand-authored first constitution (prerequisites)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §6 — named prerequisites, "currently assumed, not called out"; a thin_slice_constitution.md flat file exists and rides in every bundle.
- **what:** The concern taxonomy (standards.concern) and priority dimensions are fixed vocabularies the conventions/intent agents classify INTO, so they must be authored before those agents run. The first constitution is hand-written from 001/003 + working preferences (the tool that would help write it doesn't exist yet); the thin-slice flat-file constitution is its interim form, later becoming standards rows with scope=constitution. Also: "us" is a real tenant row, not a sentinel, so single-tenant exercises the multi-tenant code path.
- **sources:** FOCUS_pre_build_edge_cases(1).md#6; PLAN_active_config_schema(3).md#1,#3.1; tasks/gameslink bundles (constitution section present)
- **relations:** active-config schema; thin-slice-first
- **verify-later:** thin_slice_constitution.md content vs standards rows

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Domain submission & mission-driven sites (three tiers)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(0) "The domain-submitter agent is the entry point for all new site builds … Three tiers of domain submission"
- **what:** `domain-submitter` is the entry point: it creates the site record, persists to `site_specs`, and emits the first `needs_domain_research` item. Three tiers: domain-only, domain+objective, and domain+`mission`/`roadmap`. Mission/roadmap aspects support any pre-planned site (e.g. vonc.com/Spark), bypassing the classifier's domain-discovery.
- **sources:** WM/001_development_guide(0).md#domain-submission-trigger-script-reference, WM/007_adoption_pipeline_v3.md#mission-driven-sites
- **relations:** classifier strategic brain; adoption modes; vonc
- **verify-later:** domain-submitter agent_definition; site_specs aspects mission/roadmap

<!-- SOURCE: U18_sql_for_agents.md -->
### build-briefing-agent (spec-reading briefing)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 050 definition, "Distinct from existing briefing-agent (v1) which... receives questionnaire directly as input. This version reads from site_specs."
- **what:** Handler for needs_briefing: answers the briefing questionnaire autonomously from site_specs identity + classification (no human), writes aspect "briefing", creates needs_site_plan. Marks the shift from HITL-driven briefing to spec-derived config.
- **sources:** 050_build_briefing_agent.sql
- **relations:** v1 briefing-agent (superseded for this path); build-site-planner downstream
- **verify-later:** briefing aspect shape

<!-- SOURCE: U20_legacy_docs_a.md -->
### Briefing agent (structured brief generation)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** Agent SQL (021) inserted into the 6-step pipeline; extended with site_type detection (023); later generalised to questionnaire execution (029). Onboarding/config-derivation docs are the live successor area.
- **what:** First pipeline stage: an LLM turns domain+objective into a comprehensive structured brief JSON — industry inference with confidence, audience demographics/psychographics, brand tone/personality/voice examples, value proposition/key messages/USPs, recommended sections, theme recommendation with semantic tags, content guidelines (avoid/emphasise), monetisation model and ad zones.
- **sources:** docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md
- **relations:** site classifier; questionnaire pattern; successor: onboarding-config (docs019 PLAN_onboarding) and site specs.
- **verify-later:** briefing-agent row; whether brief JSON shape survives in site_specs.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intake orchestrator with two HITL gates and per-group briefing questionnaires
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** 029.intake_and_groups.sql implements schema (briefing_questionnaire column), site-classifier, intake-orchestrator group, landing/content builder groups; Go actions written (request_human_input with skip conditions, fetch_group_questionnaire); registry additions still listed as "needed".
- **what:** A two-stage front door: classify project (site_type + recommended group) → HITL-1 confirm type → fetch the *target group's* briefing questionnaire (stored in agent_group_definitions, keeping the briefing agent generic) → execute questionnaire (LLM-inferred or human-answered) → HITL-2 review brief → spawn_group dynamically dispatches the chosen builder. HITL points have skip conditions (hitl_mode=auto) for automated runs.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#the-intake-orchestrator; docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md
- **relations:** await_approval mechanism (reused); successor: onboarding-config PLAN_onboarding / config derivation.
- **verify-later:** intake-orchestrator group row; request_human_input/fetch_group_questionnaire in registry; briefing_questionnaire column.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Briefing agent (pre-strategist brief enrichment)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs005 sketch ("A new agent type briefing-agent... Sits before chief-strategist"); docs006/011 shows it live in the intake workflow ("call_briefer → Briefer fills questionnaire (HITL or LLM)").
- **what:** An agent inserted before the strategist that takes raw user input (domain, rough objective), asks clarifying questions, and outputs a structured brief (audience, tone, USPs, competitors, key messages) with a human approval pause. Evolved into the briefing-agent + per-builder `briefing_questionnaire` with interactive (HITL) and auto (LLM-infer) modes.
- **sources:** docs005_briefing_agent_domain_authority/README.0130.briefing_agent.md; docs006_workflow_builder/011_working_landing_page_builder.md#Briefing-Agent; docs006_workflow_builder/003_current_state_of_agents.sql#3-BRIEFING-AGENT
- **relations:** builder questionnaires per site type; intake orchestrator; HITL pauses; successor: reviewed_brief in current build pipeline.
- **verify-later:** agent_definitions row 'briefing-agent'; briefing_questionnaire column on agent_definitions; current intake workflow.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Per-builder briefing questionnaires
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs006/002 full questionnaire JSON on landing-page-builder and content-site-builder definitions; docs007/001 contrasts landing (10 conversion fields) vs brochure (15+ corporate fields).
- **what:** Each builder agent definition carries a `briefing_questionnaire` JSONB (sections of typed questions — brand, value proposition, conversion, social proof for landing; company, services, leadership, case studies for brochure). `fetch_agent_questionnaire` retrieves the correct questionnaire for the chosen builder, and the briefing agent fills it via HITL or LLM inference.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Step-2; docs007_brochure_builder/001_brochure_builder_plan.md#Questionnaire-Differences; docs006_workflow_builder/003_current_state_of_agents.sql
- **relations:** briefing agent; site classifier; reviewed_brief.
- **verify-later:** briefing_questionnaire values in agent_definitions; fetch_agent_questionnaire action in Go.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Domain submission tiers and mission/roadmap briefs
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(5) documents current domain-submitter workflow with persist steps
- **what:** domain-submitter is the entry point for new builds. Three tiers: domain only; domain+objective hint; domain+mission/roadmap (structured JSON for machine consumers + plain-text `mission_brief`/`roadmap_brief` that classifier/planner actually read). Persist steps skip gracefully via error_step when fields absent; briefs must be plain text parseable by small models.
- **sources:** 001(5)#Domain Submission; 007#Mission-Driven Sites
- **relations:** classifier weighting of inputs (028); vonc/Spark pattern
- **verify-later:** domain-submitter agent definition; site_specs mission aspects

<!-- SOURCE: U01_docs024_numbered_core.md -->
### build_queue domain queue with direction spectrum
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** P1 (marked "a bit out of date but still has merit"); P2 depends on it for POST /sites; seed_build_queue named in 032/other docs as real
- **what:** build_queue rows (domain, direction jsonb, status, batch, priority); direction spans null → objective hint → full brief (skip research+briefing) → adopt_from → fork_from (specs pre-populated). seed_build_queue takes N, ensures site records, writes initial specs, inserts the appropriate first work item; pacing by batch size. Initial chain: needs_domain_research → needs_briefing → needs_site_plan with spec outputs per handler.
- **sources:** P1#Domain Queue, #Initial Build
- **relations:** public API POST /sites; domain-submitter (newer entry path — reconcile)
- **verify-later:** build_queue table + seed action exist; relation to domain-submitter

<!-- SOURCE: U15_docs019_running_notes.md -->
### Onboarding/config three-layer model
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** "The config has three layers with different derivability: mechanical (discovered + probed), conventions (inferred or doc-sourced — confirmed), intent (elicited)." (principles(59) §Onboarding and config).
- **what:** A framing that treats tenant/codebase onboarding as three separate problems with different confirmation mechanisms and different climb rates on the trust ratchet: mechanical facts (probed, confirmable by reality, climb fastest), conventions (inferred-then-confirmed even in docs-authoritative mode, since hallucinated conventions would manufacture drift), and intent (elicited progressively, never "done," captured just-in-time as work happens rather than as an upfront tax).
- **sources:** NOTES_running_synthesis_principles(59) §Onboarding and config (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; doc claim-verification convention (shares "inferred-then-confirmed" DNA).

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three-layer config: mechanical / conventions / intent (different derivability)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation "Status: plan"; nothing in these files claims implementation.
- **what:** Onboarding is three processes, not one: the mechanical layer is discoverable (inspect + probe; low stakes, confirmable by reality); conventions are inferred or doc-sourced (a strong draft, weak authority — code shows what it does, not what it should do); intent and standards are elicited (not derivable from source; the tenant is the source, and the part delivering the tool's distinctive value).
- **sources:** PLAN_onboarding_config_derivation.md#1; 001_onboarding_discussion.txt
- **relations:** the five onboarding agents; docs-authoritative decision
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Progressive onboarding — a ramp, never "done"
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §2 — plan.
- **what:** A tenant gets value from the mechanical layer alone (fresh code context, signatures, reuse search, schema) before any intent is captured; conventions and intent fill incrementally and the tool deepens as they arrive. Onboarding tracks the repo forever — active-with-pending is the steady state, and leaf-level intent is captured just-in-time during use rather than as a setup tax.
- **sources:** PLAN_onboarding_config_derivation.md#2; PLAN_onboarding_agent_specs(6).md#3.7,#4.3
- **relations:** intent-elicitation agent; config-maintenance agent
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Config as a maintained artifact (the wizard is the first pass; the lifecycle is the deliverable)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §3 — plan.
- **what:** The derived config drifts as the repo changes, so it gets the standards' own upkeep machinery: periodic re-derivation with divergence flagging, confirm-not-initiate on proposed changes, and per-entry provenance (discovered/inferred/supplied) determining trust and change authority. "Onboarding as a first-class deliverable" means this lifecycle, not a good setup script.
- **sources:** PLAN_onboarding_config_derivation.md#3; 001_onboarding_discussion.txt
- **relations:** config-maintenance agent; active-config provenance shape
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Inference quality scales with codebase quality — surface uncertainty
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §4 — named tension, design mitigation.
- **what:** On a messy repo, convention inference confidently drafts the repo's bad habits, and confirming that codifies the mess — so the more a tenant needs the tool, the less their repo can teach it. Mitigation: surface inconsistency as questions to resolve, never a silent majority pick; inconsistency found during onboarding is itself valuable output ("your conventions aren't actually conventions").
- **sources:** PLAN_onboarding_config_derivation.md#4; 001_onboarding_discussion.txt
- **relations:** conventions agent; docs-authoritative mode
- **verify-later:** n/a (plan)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Docs-authoritative conventions for our own repo (the free drift audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_config_derivation §5 "Ours (decided): docs-authoritative" — the decision is recorded; the audit has not run.
- **what:** Source-of-truth for conventions is chosen per tenant by doc availability; for our repo, 001/003/the naming FOCUS are authoritative and code is read only to find disagreements. Each disagreement is recorded, not silently resolved — the set is a free audit of where the codebase drifted from its own documented standards, the drift detector's first run, on us. Our own onboarding is the template, not a special case.
- **sources:** PLAN_onboarding_config_derivation.md#5,#7; 001_onboarding_discussion.txt
- **relations:** conventions agent; drift audit three-bucket output
- **verify-later:** whether any drift audit ran

<!-- SOURCE: U16_docs019_design_plans.md -->
### Conventions agent (extract-cite-confirm, then audit)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1 — spec only.
- **what:** Owns the conventions layer: extracts discrete convention atoms from the standards docs (each citing its exact doc span — extraction is inferred-then-confirmed, because auditing code against an invented convention manufactures fake drift, the one failure that would discredit the audit), gets the set human-confirmed BEFORE any audit, then checks code and records disagreements with location/convention/tier/confidence and a default disposition (code-drifted, doc-drifted, or legitimate exception — human confirms). Accepted exceptions are remembered so audits become incremental.
- **sources:** PLAN_onboarding_agent_specs(6).md#1
- **relations:** three checking tiers; docs-authoritative decision; check_*.go validators
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Three checking tiers + three-bucket audit output (coverage honesty)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 — spec; the pattern recurs in maintenance §5.5.
- **what:** Conventions (and drift) are checked at three tiers: deterministic (static check settles it → violations), heuristic proxy (a measurable indicator flags candidates, not violations — "where to look, not what's wrong"; an optional LLM pass is still only a candidate flag, never a verdict), judgement-only (no proxy → reported as a coverage gap). The audit reports three numbers, never one — a clean tier-1 count beside many unchecked tier-3 conventions is a partial audit with known limits, and must say so. Companion role split: un-auditable conventions still serve as generation guidance (an atom can be audited, guiding, or both).
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#1.6; PLAN_onboarding_agent_specs(6).md#5.5
- **relations:** conventions agent; config-maintenance drift tiers; LLM-as-candidate principle
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Convention coverage IS capability reliability
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §1.9 closing — spec insight.
- **what:** When a bundle capability rests on a manual convention (log-correlation needs orchestration_id in every log line), the capability is only as reliable as the convention's coverage, not its existence. For any capability-bearing convention the audit reports how completely it is followed, and gaps surface as fixable (add the missing log statements) rather than hard limits — even on our own codebase, where the structure exists but coverage is unverified.
- **sources:** PLAN_onboarding_agent_specs(6).md#1.9,#2.9
- **relations:** codebase-conditional capabilities; runtime evidence by orchestration_id
- **verify-later:** an orchestration_id logging coverage scan

<!-- SOURCE: U16_docs019_design_plans.md -->
### Stack-discovery agent (inspect → interpret → declared probe plan → probe → confirm)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2 — spec only.
- **what:** Owns the mechanical layer: read-only inspection emits facts; interpretation ("this Makefile target is probably the test command") emits proposals with confidence — the subtle point being that interpretation has uncertainty even at the mechanical layer; a declared probe plan (the security contract, kept even for our own use as audit) precedes sandboxed probes; probe results update confidence. A failing build is useful output, candidate-only interpreted, never fixed by this agent. The output document carries per-entry source/confidence/probe-result with uncertainties listed separately. Also records the structural facts bundle capabilities depend on (§2.9).
- **sources:** PLAN_onboarding_agent_specs(6).md#2
- **relations:** confirmation by reality; sandboxing envelope; codebase-conditional capabilities
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirmation by reality (the mechanical layer climbs the ratchet first)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.8; contract-set review §3.1 records the reconciliation.
- **what:** The mechanical layer can be confirmed by observation (the probed command actually works) — the strongest confirmation any config layer carries — so stack-discovery is the natural first capability to graduate past confirm_every. Reconciled with the gate: probe success is initially strong evidence inside the work-item gate (near-rubber-stamp, human still activates); only after trust-ledger graduation does probe success auto-activate. The gate is the starting position; graduation relaxes it — not a bypass.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.8; FOCUS_contract_set_review.md#3.1; PLAN_active_config_schema(3).md#5
- **relations:** trust ledger; confirm-not-initiate
- **verify-later:** n/a (design)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Sandboxed probing — the tenant-code security envelope
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §2.6: "gating for the service phase: no tenant code runs until sandboxing is solid."
- **what:** The first agent that may execute tenant code does so inside an ephemeral sandbox: repo mounted read-only, restricted network, time limit, no persistent state; the emitted probe plan is the contract the sandbox approves/restricts/denies per command. The Tier-C security concern made concrete; same gate applies to Phase-2 verification running tenant code.
- **sources:** PLAN_onboarding_agent_specs(6).md#2.6; PLAN_context_assembly_tool_and_service(2).md#6
- **relations:** stack-discovery; service phase
- **verify-later:** n/a (unbuilt)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Intent-elicitation agent (progressive, value-returning interview)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §3 — spec only; reuse target (briefing_questionnaire column) verified real in FOCUS_schema_verification_findings §3.
- **what:** Captures the why-chain, per-node priority profiles and direction-of-travel via an interview that interleaves proposal-confirmation (where evidence exists — low friction, anchoring risk mitigated by citing the evidence so proposals are contestable) with free elicitation (blank page, unavoidable). Every exchange returns value (the captured piece changes the next bundle/mediation); the interview is not finite — leaf intent is captured just-in-time in the flow of work. A descendant of the briefing questionnaire / intake orchestrator, pointed at a codebase. Capture and use are separate roles (the user-rep advocate consumes what this captures). Open: detecting rubber-stamping.
- **sources:** PLAN_onboarding_agent_specs(6).md#3; FOCUS_schema_verification_findings.md#3
- **relations:** onboarding orchestrator; objectives table; user-rep advocate (salience doc, other unit)
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Onboarding orchestrator (dependency-graph flow; active-with-pending)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §4 — spec only.
- **what:** Coordinates the three layer agents: stack-discovery first (both others depend on its mechanical config), conventions and intent in parallel (independent of each other) — sequencing follows dependencies, not policy. Routes all proposals through confirm-not-initiate; surfaces a compact onboarding-state artifact (per-layer confirmed/partial/blocked, pending, drift-audit counts); a blocked layer doesn't stop the others; a tenant walking away pauses cleanly. Terminal state is active-with-pending, handing over to maintenance — never "fully done".
- **sources:** PLAN_onboarding_agent_specs(6).md#4; FOCUS_onboarding_system_view_check.md#1,#3.4
- **relations:** the three layer agents; config-maintenance handoff; work-items queue
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Config-maintenance agent (drift detection as the trust ratchet's signal source)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_onboarding_agent_specs(6) §5 — spec only.
- **what:** After baseline, detects drift across all three layers, event-driven (change-layer diffs) plus a periodic sweep, dispatching to the layer agents for re-checks rather than reimplementing them; targeted re-validation (implicated-only recheck) instead of full sweeps. Drift evidence uses the same three tiers; surfacing is prioritised to avoid alert fatigue (high-impact deterministic first, heuristic in paced batches, freshness nudges background). Its deeper role: sustained no-drift is graduation evidence and repeated drift is de-graduation evidence — without this agent the bidirectional ratchet has nothing to act on at the right timescale.
- **sources:** PLAN_onboarding_agent_specs(6).md#5; FOCUS_onboarding_system_view_check.md#2
- **relations:** trust ledger; change-layer integration; published-reasoning gap detection
- **verify-later:** n/a (spec)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Active-config schema (four tables, computed-on-read effective values)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** PLAN_active_config_schema(3) "Status: contract specification"; corrected to chassis conventions after schema verification.
- **what:** The load-bearing contract: tenant_configs (scope-holder row per tenant, created directly at init — not a gate violation), mechanical_config (one JSONB row, per-field embedded provenance), standards (flat concern atoms with scope constitution/domain/leaf, applies_to change types, rule/rationale/check/check_kind), objectives (nested why-chain nodes with priority_profile, direction_of_travel, standing_concerns). A common provenance shape (source/source_ref/confidence/status/last_verified_at/verified_by/freshness_until/version/previous_version_id/deleted_at) across all layers so consumers reason uniformly. Effective priority profile is computed at read time by walking root→node (store authored differences, compute effective on read); acyclicity must be enforced on write AND the walk bounded, since a human can confirm a cycle. The constitution is a view over standards WHERE scope=constitution, not a table. Two atom trees deliberately kept distinct: flat concern tree vs nested objective tree.
- **sources:** PLAN_active_config_schema(3).md; FOCUS_onboarding_system_view_check.md#3.1,#3.7; FOCUS_pre_build_edge_cases(1).md#1.1
- **relations:** all six contracts hang off it; bundle authored layer reads it
- **verify-later:** whether any of the four tables exist in clients_db

<!-- SOURCE: U16_docs019_design_plans.md -->
### Governed vocabularies and the hand-authored first constitution (prerequisites)
- **category:** onboarding-config
- **status-signal:** aspirational
- **status-evidence:** FOCUS_pre_build_edge_cases §6 — named prerequisites, "currently assumed, not called out"; a thin_slice_constitution.md flat file exists and rides in every bundle.
- **what:** The concern taxonomy (standards.concern) and priority dimensions are fixed vocabularies the conventions/intent agents classify INTO, so they must be authored before those agents run. The first constitution is hand-written from 001/003 + working preferences (the tool that would help write it doesn't exist yet); the thin-slice flat-file constitution is its interim form, later becoming standards rows with scope=constitution. Also: "us" is a real tenant row, not a sentinel, so single-tenant exercises the multi-tenant code path.
- **sources:** FOCUS_pre_build_edge_cases(1).md#6; PLAN_active_config_schema(3).md#1,#3.1; tasks/gameslink bundles (constitution section present)
- **relations:** active-config schema; thin-slice-first
- **verify-later:** thin_slice_constitution.md content vs standards rows

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Domain submission & mission-driven sites (three tiers)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(0) "The domain-submitter agent is the entry point for all new site builds … Three tiers of domain submission"
- **what:** `domain-submitter` is the entry point: it creates the site record, persists to `site_specs`, and emits the first `needs_domain_research` item. Three tiers: domain-only, domain+objective, and domain+`mission`/`roadmap`. Mission/roadmap aspects support any pre-planned site (e.g. vonc.com/Spark), bypassing the classifier's domain-discovery.
- **sources:** WM/001_development_guide(0).md#domain-submission-trigger-script-reference, WM/007_adoption_pipeline_v3.md#mission-driven-sites
- **relations:** classifier strategic brain; adoption modes; vonc
- **verify-later:** domain-submitter agent_definition; site_specs aspects mission/roadmap

<!-- SOURCE: U18_sql_for_agents.md -->
### build-briefing-agent (spec-reading briefing)
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 050 definition, "Distinct from existing briefing-agent (v1) which... receives questionnaire directly as input. This version reads from site_specs."
- **what:** Handler for needs_briefing: answers the briefing questionnaire autonomously from site_specs identity + classification (no human), writes aspect "briefing", creates needs_site_plan. Marks the shift from HITL-driven briefing to spec-derived config.
- **sources:** 050_build_briefing_agent.sql
- **relations:** v1 briefing-agent (superseded for this path); build-site-planner downstream
- **verify-later:** briefing aspect shape

<!-- SOURCE: U20_legacy_docs_a.md -->
### Briefing agent (structured brief generation)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** Agent SQL (021) inserted into the 6-step pipeline; extended with site_type detection (023); later generalised to questionnaire execution (029). Onboarding/config-derivation docs are the live successor area.
- **what:** First pipeline stage: an LLM turns domain+objective into a comprehensive structured brief JSON — industry inference with confidence, audience demographics/psychographics, brand tone/personality/voice examples, value proposition/key messages/USPs, recommended sections, theme recommendation with semantic tags, content guidelines (avoid/emphasise), monetisation model and ad zones.
- **sources:** docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md
- **relations:** site classifier; questionnaire pattern; successor: onboarding-config (docs019 PLAN_onboarding) and site specs.
- **verify-later:** briefing-agent row; whether brief JSON shape survives in site_specs.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intake orchestrator with two HITL gates and per-group briefing questionnaires
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** 029.intake_and_groups.sql implements schema (briefing_questionnaire column), site-classifier, intake-orchestrator group, landing/content builder groups; Go actions written (request_human_input with skip conditions, fetch_group_questionnaire); registry additions still listed as "needed".
- **what:** A two-stage front door: classify project (site_type + recommended group) → HITL-1 confirm type → fetch the *target group's* briefing questionnaire (stored in agent_group_definitions, keeping the briefing agent generic) → execute questionnaire (LLM-inferred or human-answered) → HITL-2 review brief → spawn_group dynamically dispatches the chosen builder. HITL points have skip conditions (hitl_mode=auto) for automated runs.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#the-intake-orchestrator; docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md
- **relations:** await_approval mechanism (reused); successor: onboarding-config PLAN_onboarding / config derivation.
- **verify-later:** intake-orchestrator group row; request_human_input/fetch_group_questionnaire in registry; briefing_questionnaire column.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Briefing agent (pre-strategist brief enrichment)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs005 sketch ("A new agent type briefing-agent... Sits before chief-strategist"); docs006/011 shows it live in the intake workflow ("call_briefer → Briefer fills questionnaire (HITL or LLM)").
- **what:** An agent inserted before the strategist that takes raw user input (domain, rough objective), asks clarifying questions, and outputs a structured brief (audience, tone, USPs, competitors, key messages) with a human approval pause. Evolved into the briefing-agent + per-builder `briefing_questionnaire` with interactive (HITL) and auto (LLM-infer) modes.
- **sources:** docs005_briefing_agent_domain_authority/README.0130.briefing_agent.md; docs006_workflow_builder/011_working_landing_page_builder.md#Briefing-Agent; docs006_workflow_builder/003_current_state_of_agents.sql#3-BRIEFING-AGENT
- **relations:** builder questionnaires per site type; intake orchestrator; HITL pauses; successor: reviewed_brief in current build pipeline.
- **verify-later:** agent_definitions row 'briefing-agent'; briefing_questionnaire column on agent_definitions; current intake workflow.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Per-builder briefing questionnaires
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** docs006/002 full questionnaire JSON on landing-page-builder and content-site-builder definitions; docs007/001 contrasts landing (10 conversion fields) vs brochure (15+ corporate fields).
- **what:** Each builder agent definition carries a `briefing_questionnaire` JSONB (sections of typed questions — brand, value proposition, conversion, social proof for landing; company, services, leadership, case studies for brochure). `fetch_agent_questionnaire` retrieves the correct questionnaire for the chosen builder, and the briefing agent fills it via HITL or LLM inference.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Step-2; docs007_brochure_builder/001_brochure_builder_plan.md#Questionnaire-Differences; docs006_workflow_builder/003_current_state_of_agents.sql
- **relations:** briefing agent; site classifier; reviewed_brief.
- **verify-later:** briefing_questionnaire values in agent_definitions; fetch_agent_questionnaire action in Go.

<!-- SOURCE: U26_misc_dirs.md -->
### Agent spawning (agents as DB records claimed by generic pods)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001: "The spawn_group action creates new rows in the agent_instances table"; 004_debugging references `kubectl delete jobs -l spawned-by=orchestrator` — orchestrator-spawned jobs existed in the running cluster.
- **what:** A spawn_agent action creates (or reuses) an agent_instances row with type, workflow, capabilities and llm_config; a generic chassis pod (static env assignment, dynamic pool claim, or K8s Job spawned by the orchestrator) loads that config and becomes the agent. Includes existence-check-and-reuse logic and pod-type selection (CPU/GPU/memory) for specialised workloads.
- **sources:** docs/architecture/023-spawning-agents.md; docs/architecture/025-reusable-evolvable-agent-teams#step-1; docs/basic_usage/001basic_usage.txt#part-2
- **relations:** agent chassis; agent groups; agent discovery
- **verify-later:** spawn_actions.go; assigned_pod/status/last_heartbeat columns; Job-spawning code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent groups (reusable multi-agent teams)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 documents live spawn_group runs against the agent_groups table; hitl_agent_group_definition.sql (Nov 2025) inserts into a matured `agent_group_definitions` table with versioning and ON CONFLICT upsert.
- **what:** A named, versioned team definition: group_type, agent_configs (roles → agent types), an orchestration_workflow describing how they cooperate, capabilities/tags for search, and usage/performance metadata. `spawn_group` instantiates the team and starts the group workflow. Groups were intended to be saved from successful configurations, forked, and improved over generations.
- **sources:** docs/architecture/025-reusable-evolvable-agent-teams#phase-2; docs/humanintheloop/hitl_agent_group_definition.sql; docs/basic_usage/001basic_usage.txt
- **relations:** website-builder group; controlled group evolution; group discovery
- **verify-later:** agent_groups vs agent_group_definitions tables; SpawnGroupAction code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent and group discovery by capability and performance
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 024-agent-discovery-framework.md reads as pure proposal ("Yes! Let's design this") with an unconfirmed agent_metrics/heartbeat table, BUT a live repo check (not just docs) found `platform/discovery/` actually exists — `agent_discovery.go` plus `README.001.agentdefinitions.md` and `README.002.dbtables.md` — confirming a real, if unknown-scope, implementation landed. Upgraded from aspirational on that direct code evidence; the sophistication of the shipped version (whether performance-ranking/heartbeats made it in) is still stage-2 work.
- **what:** A registry service that finds the best existing agent (or group) for a task by required capabilities (JSONB containment), success rate, response time, availability (heartbeat) and fuel cost — spawning a new one only when nothing matches. The "self-organizing system" goal: agents discover each other, learn which perform best, optimise over time.
- **sources:** docs/architecture/024-agent-discovery-framework.md; docs/architecture/027-create-website-creation-system#phase-3; docs/architecture/025-reusable-evolvable-agent-teams#group-discovery-service
- **relations:** agent spawning; agent groups; template classification
- **verify-later:** platform/discovery/agent_discovery.go, platform/discovery/README.001.agentdefinitions.md, platform/discovery/README.002.dbtables.md — confirm scope of what actually shipped vs. this proposal

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow template library, lineage and marketplace
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 005/006 propose workflow_templates + template_versions tables, fork lineage graphs and a monetised marketplace ("marketplace of evolving workflow templates"); roadmap slots it at Phase 6 (weeks 10-12); no later doc in the repo era references these tables.
- **what:** Successful workflow executions are saved as reusable templates with lineage (parent_template_id, source_correlation_id), performance metrics, ratings and usage counts; users fork and improve templates ("collective intelligence — natural selection of best-performing templates"), with a WorkflowOptimizer suggesting parallelisation/reordering from execution history. A precursor idea whose spirit partially survives in agent group versioning.
- **sources:** docs/architecture/005-template-classification-and-evolution.md; docs/architecture/007-roadmap.md#phase-6; docs/architecture/016-competitive-advantge.md#workflow-marketplace-2.0
- **relations:** agent groups; template classification and search; controlled group evolution
- **verify-later:** confirm workflow_templates tables never created

<!-- SOURCE: U26_misc_dirs.md -->
### Multi-dimensional template classification and semantic search
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 006 is pure design (behavioral profiles, performance vectors, embedding fingerprints, vector DB + graph DB search service, recommendation engine); none of the infrastructure (vectorDB for templates, template_usage_metrics) appears in later docs.
- **what:** Replace flat tags with a rich classification for discovering workflows/templates: behavioral capabilities, execution style, resource usage, normalised performance vectors with trade-off scores, embedding-based semantic fingerprints, outcome-based deliverable descriptions and evolutionary metadata (lineage depth, fork count). A parallel multi-strategy search (semantic + behavioral + performance + lineage) with collaborative-filtering recommendations.
- **sources:** docs/architecture/006-categorisation-template-search-evolution.md
- **relations:** workflow template library; agent/group discovery
- **verify-later:** n/a (idea only) — check nothing similar exists under another name

<!-- SOURCE: U26_misc_dirs.md -->
### Controlled group evolution (observed mutation with rules)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 026 reads as pure future-tense recommendation ("Start with curated base templates that can dynamically evolve", MutationRules, sandbox testing, 10%-improvement gates) with nothing downstream in this unit confirming it. BUT a live repo check found `platform/evolution/` actually exists — `evolution.go`, `performance.go`, and a README stating in the present tense: "Evolution Service: Evaluates groups for potential improvements / Applies mutations (parallel agents, specialists, validators) / Tracks evolution history / Version management" and "Performance Analysis: Records and analyzes execution metrics / Identifies bottlenecks and failures / Generates improvement suggestions". Upgraded from abandoned on that direct code evidence — this idea did get built, just not confirmed anywhere in the docs sampled for this unit.
- **what:** Hybrid manual/dynamic strategy: hand-curated base agent and group templates act as "genetic seeds"; a metrics observer detects bottlenecks/missing capabilities after ≥5 uses and proposes constrained mutations (add parallel agent, add specialist, replace, adjust workflow, fork) which must beat a performance baseline in sandbox before becoming a new version with parent lineage. Human-in-the-loop approval tiers for major changes.
- **sources:** docs/architecture/026-manual-vs-dynamic-templates; docs/architecture/025-reusable-evolvable-agent-teams#usage-example
- **relations:** agent groups; dynamic prompt improvement loop; workflow template library
- **verify-later:** platform/evolution/evolution.go, platform/evolution/performance.go, platform/evolution/README.md — confirm how closely the shipped mutation rules match this design

<!-- SOURCE: U26_misc_dirs.md -->
### Dynamic prompt improvement loop (Prompt Improvement Agent)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** aspirational
- **status-evidence:** basic_usage/003 is a plan ("Phase 2: The Evolution Loop") — flag-for-improvement UI action, `system.prompt.improvement.request` topic, save_new_agent_definition action creating versioned definitions (html-developer-v2); the agent_definitions versioning columns (version, previous_version_id) DID ship (visible in hitl_agent_definition.sql), the loop itself is not evidenced.
- **what:** End-of-workflow human review offers "Approve" or "Flag for Improvement"; flagged runs dispatch the failing agent's prompt + failure context to a prompt-engineering specialist agent which generates an improved prompt, gets human approval, and saves a NEW versioned agent definition (never mutating the old). Includes bootstrap_prompt for generating a first prompt for brand-new agent types from a description.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement; docs/humanintheloop/hitl_agent_definition.sql (version/previous_version_id columns)
- **relations:** execute_llm_prompt; HITL approval mechanism; controlled group evolution; (spiritual ancestor of the current improvement-loop / finetuning flywheel)
- **verify-later:** prompt-improvement-agent definition; save_new_agent_definition action; version columns usage

<!-- SOURCE: U26_misc_dirs.md -->
### Agent spawning (agents as DB records claimed by generic pods)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001: "The spawn_group action creates new rows in the agent_instances table"; 004_debugging references `kubectl delete jobs -l spawned-by=orchestrator` — orchestrator-spawned jobs existed in the running cluster.
- **what:** A spawn_agent action creates (or reuses) an agent_instances row with type, workflow, capabilities and llm_config; a generic chassis pod (static env assignment, dynamic pool claim, or K8s Job spawned by the orchestrator) loads that config and becomes the agent. Includes existence-check-and-reuse logic and pod-type selection (CPU/GPU/memory) for specialised workloads.
- **sources:** docs/architecture/023-spawning-agents.md; docs/architecture/025-reusable-evolvable-agent-teams#step-1; docs/basic_usage/001basic_usage.txt#part-2
- **relations:** agent chassis; agent groups; agent discovery
- **verify-later:** spawn_actions.go; assigned_pod/status/last_heartbeat columns; Job-spawning code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent groups (reusable multi-agent teams)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 documents live spawn_group runs against the agent_groups table; hitl_agent_group_definition.sql (Nov 2025) inserts into a matured `agent_group_definitions` table with versioning and ON CONFLICT upsert.
- **what:** A named, versioned team definition: group_type, agent_configs (roles → agent types), an orchestration_workflow describing how they cooperate, capabilities/tags for search, and usage/performance metadata. `spawn_group` instantiates the team and starts the group workflow. Groups were intended to be saved from successful configurations, forked, and improved over generations.
- **sources:** docs/architecture/025-reusable-evolvable-agent-teams#phase-2; docs/humanintheloop/hitl_agent_group_definition.sql; docs/basic_usage/001basic_usage.txt
- **relations:** website-builder group; controlled group evolution; group discovery
- **verify-later:** agent_groups vs agent_group_definitions tables; SpawnGroupAction code

<!-- SOURCE: U26_misc_dirs.md -->
### Agent and group discovery by capability and performance
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 024-agent-discovery-framework.md reads as pure proposal ("Yes! Let's design this") with an unconfirmed agent_metrics/heartbeat table, BUT a live repo check (not just docs) found `platform/discovery/` actually exists — `agent_discovery.go` plus `README.001.agentdefinitions.md` and `README.002.dbtables.md` — confirming a real, if unknown-scope, implementation landed. Upgraded from aspirational on that direct code evidence; the sophistication of the shipped version (whether performance-ranking/heartbeats made it in) is still stage-2 work.
- **what:** A registry service that finds the best existing agent (or group) for a task by required capabilities (JSONB containment), success rate, response time, availability (heartbeat) and fuel cost — spawning a new one only when nothing matches. The "self-organizing system" goal: agents discover each other, learn which perform best, optimise over time.
- **sources:** docs/architecture/024-agent-discovery-framework.md; docs/architecture/027-create-website-creation-system#phase-3; docs/architecture/025-reusable-evolvable-agent-teams#group-discovery-service
- **relations:** agent spawning; agent groups; template classification
- **verify-later:** platform/discovery/agent_discovery.go, platform/discovery/README.001.agentdefinitions.md, platform/discovery/README.002.dbtables.md — confirm scope of what actually shipped vs. this proposal

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow template library, lineage and marketplace
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 005/006 propose workflow_templates + template_versions tables, fork lineage graphs and a monetised marketplace ("marketplace of evolving workflow templates"); roadmap slots it at Phase 6 (weeks 10-12); no later doc in the repo era references these tables.
- **what:** Successful workflow executions are saved as reusable templates with lineage (parent_template_id, source_correlation_id), performance metrics, ratings and usage counts; users fork and improve templates ("collective intelligence — natural selection of best-performing templates"), with a WorkflowOptimizer suggesting parallelisation/reordering from execution history. A precursor idea whose spirit partially survives in agent group versioning.
- **sources:** docs/architecture/005-template-classification-and-evolution.md; docs/architecture/007-roadmap.md#phase-6; docs/architecture/016-competitive-advantge.md#workflow-marketplace-2.0
- **relations:** agent groups; template classification and search; controlled group evolution
- **verify-later:** confirm workflow_templates tables never created

<!-- SOURCE: U26_misc_dirs.md -->
### Multi-dimensional template classification and semantic search
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 006 is pure design (behavioral profiles, performance vectors, embedding fingerprints, vector DB + graph DB search service, recommendation engine); none of the infrastructure (vectorDB for templates, template_usage_metrics) appears in later docs.
- **what:** Replace flat tags with a rich classification for discovering workflows/templates: behavioral capabilities, execution style, resource usage, normalised performance vectors with trade-off scores, embedding-based semantic fingerprints, outcome-based deliverable descriptions and evolutionary metadata (lineage depth, fork count). A parallel multi-strategy search (semantic + behavioral + performance + lineage) with collaborative-filtering recommendations.
- **sources:** docs/architecture/006-categorisation-template-search-evolution.md
- **relations:** workflow template library; agent/group discovery
- **verify-later:** n/a (idea only) — check nothing similar exists under another name

<!-- SOURCE: U26_misc_dirs.md -->
### Controlled group evolution (observed mutation with rules)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 026 reads as pure future-tense recommendation ("Start with curated base templates that can dynamically evolve", MutationRules, sandbox testing, 10%-improvement gates) with nothing downstream in this unit confirming it. BUT a live repo check found `platform/evolution/` actually exists — `evolution.go`, `performance.go`, and a README stating in the present tense: "Evolution Service: Evaluates groups for potential improvements / Applies mutations (parallel agents, specialists, validators) / Tracks evolution history / Version management" and "Performance Analysis: Records and analyzes execution metrics / Identifies bottlenecks and failures / Generates improvement suggestions". Upgraded from abandoned on that direct code evidence — this idea did get built, just not confirmed anywhere in the docs sampled for this unit.
- **what:** Hybrid manual/dynamic strategy: hand-curated base agent and group templates act as "genetic seeds"; a metrics observer detects bottlenecks/missing capabilities after ≥5 uses and proposes constrained mutations (add parallel agent, add specialist, replace, adjust workflow, fork) which must beat a performance baseline in sandbox before becoming a new version with parent lineage. Human-in-the-loop approval tiers for major changes.
- **sources:** docs/architecture/026-manual-vs-dynamic-templates; docs/architecture/025-reusable-evolvable-agent-teams#usage-example
- **relations:** agent groups; dynamic prompt improvement loop; workflow template library
- **verify-later:** platform/evolution/evolution.go, platform/evolution/performance.go, platform/evolution/README.md — confirm how closely the shipped mutation rules match this design

<!-- SOURCE: U26_misc_dirs.md -->
### Dynamic prompt improvement loop (Prompt Improvement Agent)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** aspirational
- **status-evidence:** basic_usage/003 is a plan ("Phase 2: The Evolution Loop") — flag-for-improvement UI action, `system.prompt.improvement.request` topic, save_new_agent_definition action creating versioned definitions (html-developer-v2); the agent_definitions versioning columns (version, previous_version_id) DID ship (visible in hitl_agent_definition.sql), the loop itself is not evidenced.
- **what:** End-of-workflow human review offers "Approve" or "Flag for Improvement"; flagged runs dispatch the failing agent's prompt + failure context to a prompt-engineering specialist agent which generates an improved prompt, gets human approval, and saves a NEW versioned agent definition (never mutating the old). Includes bootstrap_prompt for generating a first prompt for brand-new agent types from a description.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement; docs/humanintheloop/hitl_agent_definition.sql (version/previous_version_id columns)
- **relations:** execute_llm_prompt; HITL approval mechanism; controlled group evolution; (spiritual ancestor of the current improvement-loop / finetuning flywheel)
- **verify-later:** prompt-improvement-agent definition; save_new_agent_definition action; version columns usage

<!-- SOURCE: U21_legacy_docs_b.md -->
### entity_state_log — append-only cross-orchestration memory
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** Full schema + five Go actions (append_entity_state, read_latest_entity_state, read_entity_history, read_my_state, write_my_state) in docs006/002 and migration SQL in docs006/007; no later documents reference entity_state_log in the build pipeline.
- **what:** Persistent data that survives across orchestrations: an append-only log keyed by entity_id/namespace/path with accumulation patterns (additive, evolutionary, versioned, singleton), agent-namespaced storage ("read_my_state"/"write_my_state" use AGENT_TYPE as namespace), supersession pointers for future compaction, and LLM-based consolidation as a future enhancement. Intended for accumulating research, brand learnings, and build history per domain.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-5; docs006_workflow_builder/007_new_tables_entity_state_log.sql; docs006_workflow_builder/004_agent_groups_or_not.md#Where-Learnings-Live
- **relations:** four-level learnings model; relationships table; improvement_proposals; conceptual ancestor of per-site content_data accumulation.
- **verify-later:** entity_state_log table existence in clients_db; entity_state_actions.go in repo.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent variants + snapshot versioning
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "Snapshot model (preferred): variants explicitly reference a snapshot version"; is_snapshot column added in docs006/007 migration; agent_variants table proposed but never seen again.
- **what:** A controlled-evolution model for agent definitions: base agents are versioned and can be frozen as snapshots (is_snapshot flag); task variants (agent_variants) reference a specific base version with config_overrides, metrics, and lineage, so the base can evolve without breaking variants. Three evolution types (bug fix / improvement / innovation) with escalating oversight; promotion of successful variants to new bases left as an open question.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#The-Fragility-Problem; docs006_workflow_builder/006b_evolution_design_discussion.md; docs006_workflow_builder/007_new_tables_entity_state_log.sql
- **relations:** improvement_proposals; four-level learnings model; agent_definitions versioning today.
- **verify-later:** is_snapshot/usage_count columns on agent_definitions; whether agent_variants table was ever created.

<!-- SOURCE: U21_legacy_docs_b.md -->
### improvement_proposals — HITL-gated agent evolution queue
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "The system proposes changes but requires HITL approval before applying"; docs006/006 lists ReviewPerformanceAction → improvement_proposals and ApproveImprovementAction; not referenced by later architectures.
- **what:** A review queue where proposed changes to agent definitions, variants, or entity knowledge — sourced from metrics regressions, agent observations, or humans — wait as pending proposals until a human approves, rejects, or applies them. Included review_performance action recording execution metrics to entity_state_log and generating proposals.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#What-Triggers-Evolution; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Discovery-Actions-Changes
- **relations:** conceptual ancestor of the improvement-loop's suggest/flag resolution paths and HITL approvals.
- **verify-later:** improvement_proposals table; discovery_actions.go history.

<!-- SOURCE: U21_legacy_docs_b.md -->
### entity_state_log — append-only cross-orchestration memory
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** Full schema + five Go actions (append_entity_state, read_latest_entity_state, read_entity_history, read_my_state, write_my_state) in docs006/002 and migration SQL in docs006/007; no later documents reference entity_state_log in the build pipeline.
- **what:** Persistent data that survives across orchestrations: an append-only log keyed by entity_id/namespace/path with accumulation patterns (additive, evolutionary, versioned, singleton), agent-namespaced storage ("read_my_state"/"write_my_state" use AGENT_TYPE as namespace), supersession pointers for future compaction, and LLM-based consolidation as a future enhancement. Intended for accumulating research, brand learnings, and build history per domain.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-5; docs006_workflow_builder/007_new_tables_entity_state_log.sql; docs006_workflow_builder/004_agent_groups_or_not.md#Where-Learnings-Live
- **relations:** four-level learnings model; relationships table; improvement_proposals; conceptual ancestor of per-site content_data accumulation.
- **verify-later:** entity_state_log table existence in clients_db; entity_state_actions.go in repo.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent variants + snapshot versioning
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "Snapshot model (preferred): variants explicitly reference a snapshot version"; is_snapshot column added in docs006/007 migration; agent_variants table proposed but never seen again.
- **what:** A controlled-evolution model for agent definitions: base agents are versioned and can be frozen as snapshots (is_snapshot flag); task variants (agent_variants) reference a specific base version with config_overrides, metrics, and lineage, so the base can evolve without breaking variants. Three evolution types (bug fix / improvement / innovation) with escalating oversight; promotion of successful variants to new bases left as an open question.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#The-Fragility-Problem; docs006_workflow_builder/006b_evolution_design_discussion.md; docs006_workflow_builder/007_new_tables_entity_state_log.sql
- **relations:** improvement_proposals; four-level learnings model; agent_definitions versioning today.
- **verify-later:** is_snapshot/usage_count columns on agent_definitions; whether agent_variants table was ever created.

<!-- SOURCE: U21_legacy_docs_b.md -->
### improvement_proposals — HITL-gated agent evolution queue
- **category:** NEW:agent-memory-and-evolution
- **status-signal:** abandoned
- **status-evidence:** docs006/004: "The system proposes changes but requires HITL approval before applying"; docs006/006 lists ReviewPerformanceAction → improvement_proposals and ApproveImprovementAction; not referenced by later architectures.
- **what:** A review queue where proposed changes to agent definitions, variants, or entity knowledge — sourced from metrics regressions, agent observations, or humans — wait as pending proposals until a human approves, rejects, or applies them. Included review_performance action recording execution metrics to entity_state_log and generating proposals.
- **sources:** docs006_workflow_builder/004_agent_groups_or_not.md#What-Triggers-Evolution; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Discovery-Actions-Changes
- **relations:** conceptual ancestor of the improvement-loop's suggest/flag resolution paths and HITL approvals.
- **verify-later:** improvement_proposals table; discovery_actions.go history.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Copywriter persona roster
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/010 SQL seeds six personas (Elena Martinez B2B, James Chen technical, Marcus Williams conversion, Aisha Okonkwo thought-leadership, Raj Patel data, Sophie Dubois premium) with style agents; persona_assignments schema in docs010/009; no later builder references personas.
- **what:** A roster of copywriter personas — each a personality profile (biography, Big Five psychology, expertise weights, voice traits) with attached specialized style agents — assigned to flow stages or content types ("assign Marcus to all conversion pages") via personas / specialized_agents / persona_assignments tables and get_persona_for_page lookup (page → stage → default). Voice emerges from persona choice rather than parameter tuning; maps to real agency roles.
- **sources:** docs010_multitrack_flows_persona_architecture/008_example_personas.md; docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md
- **relations:** persona cognitive architecture; multi-track flows; page-content-writer.
- **verify-later:** personas/specialized_agents/persona_assignments tables in clients_db.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Persona cognitive architecture (swappable cognitive components)
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/015 "The architecture is ready for the full vision while starting simple today"; full schema + 8 cognitive actions + Dr Bimpton example SQL delivered; nothing downstream implements the Go actions.
- **what:** Personas as complete cognitive entities: immutable personality DNA plus swappable subsystems (perception, working/episodic/semantic memory, knowledge retrieval, reasoning engine, response generator, style applicator, learning system), each with pluggable implementations evolving Phase 1 all-LLM → vector-DB memory → fine-tuned persona models → multi-model per task → custom reasoning services, switchable via is_default without workflow changes. Running instances persist memory and emotional state; persona_knowledge holds facts/beliefs/opinions with confidence and future embeddings; task executions log full cognitive traces. Eight-step cognitive workflow per task (initialize→perceive→retrieve→reason→generate→style→learn→complete).
- **sources:** docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md; docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql; docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql
- **relations:** finetuning-flywheel (fine-tuned persona models); reasoning; entity_state_log (parallel memory design); copywriter roster.
- **verify-later:** personas/persona_cognitive_components/persona_instances/persona_knowledge/persona_task_executions tables; load_cognitive_system etc. in action registry (expected absent).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Copywriter persona roster
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/010 SQL seeds six personas (Elena Martinez B2B, James Chen technical, Marcus Williams conversion, Aisha Okonkwo thought-leadership, Raj Patel data, Sophie Dubois premium) with style agents; persona_assignments schema in docs010/009; no later builder references personas.
- **what:** A roster of copywriter personas — each a personality profile (biography, Big Five psychology, expertise weights, voice traits) with attached specialized style agents — assigned to flow stages or content types ("assign Marcus to all conversion pages") via personas / specialized_agents / persona_assignments tables and get_persona_for_page lookup (page → stage → default). Voice emerges from persona choice rather than parameter tuning; maps to real agency roles.
- **sources:** docs010_multitrack_flows_persona_architecture/008_example_personas.md; docs010_multitrack_flows_persona_architecture/009_persona_system_schema.sql; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md
- **relations:** persona cognitive architecture; multi-track flows; page-content-writer.
- **verify-later:** personas/specialized_agents/persona_assignments tables in clients_db.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Persona cognitive architecture (swappable cognitive components)
- **category:** NEW:persona-architecture
- **status-signal:** abandoned
- **status-evidence:** docs010/015 "The architecture is ready for the full vision while starting simple today"; full schema + 8 cognitive actions + Dr Bimpton example SQL delivered; nothing downstream implements the Go actions.
- **what:** Personas as complete cognitive entities: immutable personality DNA plus swappable subsystems (perception, working/episodic/semantic memory, knowledge retrieval, reasoning engine, response generator, style applicator, learning system), each with pluggable implementations evolving Phase 1 all-LLM → vector-DB memory → fine-tuned persona models → multi-model per task → custom reasoning services, switchable via is_default without workflow changes. Running instances persist memory and emotional state; persona_knowledge holds facts/beliefs/opinions with confidence and future embeddings; task executions log full cognitive traces. Eight-step cognitive workflow per task (initialize→perceive→retrieve→reason→generate→style→learn→complete).
- **sources:** docs010_multitrack_flows_persona_architecture/015_persona_README_architecture.md; docs010_multitrack_flows_persona_architecture/011_persona_cognitive_architecture.sql; docs010_multitrack_flows_persona_architecture/014_drBimpton_setup_example.sql
- **relations:** finetuning-flywheel (fine-tuned persona models); reasoning; entity_state_log (parallel memory design); copywriter roster.
- **verify-later:** personas/persona_cognitive_components/persona_instances/persona_knowledge/persona_task_executions tables; load_cognitive_system etc. in action registry (expected absent).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Multi-track flows (journeys, narrative arcs, layered context)
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** Full schema (site_flows, flow_pages, page_transitions, site_brand_dna) in docs010/002; docs010/005 "Configuration: Single-flow (production)... build for complexity, configure for simplicity"; docs012/007 MVP migration re-lists site_flows as still-to-create; no later doc shows flows populated.
- **what:** Model a site as choreographed audience journeys rather than a flat page list: each flow has an audience segment, entry points, a narrative arc of stages with per-stage voice parameters, and ordered pages with context_overrides; context inherits hierarchically SITE (immutable brand DNA) → FLOW (narrative) → PAGE (objective/overrides) → COMPONENT (paragraph tactics); navigation becomes flow-aware (different next-step CTAs per track); shared pages get per-flow variants; page_transitions support A/B weighting. "Stop thinking pages → start thinking journeys."
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql; docs010_multitrack_flows_persona_architecture/005_implementation_summary.md
- **relations:** brand DNA; voice parameters; persona assignment per stage; pattern library flow-stage tagging; conceptual ancestor of content strategy in site plans.
- **verify-later:** site_flows/flow_pages/page_transitions/site_brand_dna tables — created? populated?

<!-- SOURCE: U21_legacy_docs_b.md -->
### Brand DNA invariants with bounded variance
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** docs009/004 "brand_dna.invariants: core_message, forbidden_phrases, required_elements; variance_allowed: voice_formality [0.4,1.0]"; site_brand_dna table in docs010/002; later brand data lives in sites.content_data.brand_spec instead (docs017/019b).
- **what:** A site-level immutable identity layer — core message, values, visual system, forbidden phrases, required elements — plus explicit allowed ranges for voice variance, enforced by an evaluator check before content is accepted (vocabulary, contradiction, variance bounds, visual consistency). Solves coherence-vs-variation across multiple flows/voices.
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md#Q3; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql#BRAND-DNA
- **relations:** brand_spec in sites.content_data (descendant); content-reviewer coherence checks; design-composition brand decisions.
- **verify-later:** site_brand_dna table vs sites.content_data.brand_spec usage.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Voice parameters (numeric stage-tuned voice)
- **category:** NEW:flows-and-narrative
- **status-signal:** superseded
- **status-evidence:** docs010/019 Week 2 plan (get_voice_for_page SQL, formality 0.5 home / 0.7 elsewhere); docs010/007: "Instead of trying to tune voice parameters numerically (formality 0.7 → 0.8), we select the right copywriter persona."
- **what:** Continuous voice dials (formality, technical_depth, sales_pressure, urgency, data_density, emotional_appeal 0–1) attached to flow stages and page context_overrides, injected into content prompts so voice progresses through the journey (awareness casual → conversion formal). Explicitly superseded within its own directory by persona selection, which embodies the parameters naturally.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-2; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md#The-Key-Insight
- **relations:** copywriter persona roster (successor); multi-track flows.
- **verify-later:** get_voice_for_page function existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Multi-track flows (journeys, narrative arcs, layered context)
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** Full schema (site_flows, flow_pages, page_transitions, site_brand_dna) in docs010/002; docs010/005 "Configuration: Single-flow (production)... build for complexity, configure for simplicity"; docs012/007 MVP migration re-lists site_flows as still-to-create; no later doc shows flows populated.
- **what:** Model a site as choreographed audience journeys rather than a flat page list: each flow has an audience segment, entry points, a narrative arc of stages with per-stage voice parameters, and ordered pages with context_overrides; context inherits hierarchically SITE (immutable brand DNA) → FLOW (narrative) → PAGE (objective/overrides) → COMPONENT (paragraph tactics); navigation becomes flow-aware (different next-step CTAs per track); shared pages get per-flow variants; page_transitions support A/B weighting. "Stop thinking pages → start thinking journeys."
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql; docs010_multitrack_flows_persona_architecture/005_implementation_summary.md
- **relations:** brand DNA; voice parameters; persona assignment per stage; pattern library flow-stage tagging; conceptual ancestor of content strategy in site plans.
- **verify-later:** site_flows/flow_pages/page_transitions/site_brand_dna tables — created? populated?

<!-- SOURCE: U21_legacy_docs_b.md -->
### Brand DNA invariants with bounded variance
- **category:** NEW:flows-and-narrative
- **status-signal:** abandoned
- **status-evidence:** docs009/004 "brand_dna.invariants: core_message, forbidden_phrases, required_elements; variance_allowed: voice_formality [0.4,1.0]"; site_brand_dna table in docs010/002; later brand data lives in sites.content_data.brand_spec instead (docs017/019b).
- **what:** A site-level immutable identity layer — core message, values, visual system, forbidden phrases, required elements — plus explicit allowed ranges for voice variance, enforced by an evaluator check before content is accepted (vocabulary, contradiction, variance bounds, visual consistency). Solves coherence-vs-variation across multiple flows/voices.
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md#Q3; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql#BRAND-DNA
- **relations:** brand_spec in sites.content_data (descendant); content-reviewer coherence checks; design-composition brand decisions.
- **verify-later:** site_brand_dna table vs sites.content_data.brand_spec usage.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Voice parameters (numeric stage-tuned voice)
- **category:** NEW:flows-and-narrative
- **status-signal:** superseded
- **status-evidence:** docs010/019 Week 2 plan (get_voice_for_page SQL, formality 0.5 home / 0.7 elsewhere); docs010/007: "Instead of trying to tune voice parameters numerically (formality 0.7 → 0.8), we select the right copywriter persona."
- **what:** Continuous voice dials (formality, technical_depth, sales_pressure, urgency, data_density, emotional_appeal 0–1) attached to flow stages and page context_overrides, injected into content prompts so voice progresses through the journey (awareness casual → conversion formal). Explicitly superseded within its own directory by persona selection, which embodies the parameters naturally.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-2; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md#The-Key-Insight
- **relations:** copywriter persona roster (successor); multi-track flows.
- **verify-later:** get_voice_for_page function existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Organizational framework (roles, listeners, policy-as-filters)
- **category:** NEW:org-framework
- **status-signal:** abandoned
- **status-evidence:** Extended thought experiment across docs006/005, /006, /006c ("Acme Corp", "Sarah, Marketing Content Writer"); "Open Items for Later" never picked up; no later doc builds on roles/listeners.
- **what:** A design showing the framework is domain-agnostic by modelling a whole company: roles typed as identity/function/composite/position (only identity roles get schemas, like client_X); employees as clients with personal agent_instances; always-on shared listeners (like adapters) that spawn discrete orchestrations per task ("Sarah isn't running, she's ready"); authority as conditional filters (policy-owner agents like legal-review-agent injected by trigger conditions) rather than hierarchy; strategy flowing down as intake, decomposing at each level. Cross-cutting agents concluded to be "just agents many workflows call."
- **sources:** docs006_workflow_builder/005_acme_corp_org_chart.md; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Role-vs-Agent; docs006_workflow_builder/006c_org_framework_discussion.md
- **relations:** entity_state_log; relationships; policy filters prefigure legal-content-agent constraints; "strategy as intake" prefigures autonomous mission decomposition.
- **verify-later:** roles/role_assignments tables (expected absent).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Organizational framework (roles, listeners, policy-as-filters)
- **category:** NEW:org-framework
- **status-signal:** abandoned
- **status-evidence:** Extended thought experiment across docs006/005, /006, /006c ("Acme Corp", "Sarah, Marketing Content Writer"); "Open Items for Later" never picked up; no later doc builds on roles/listeners.
- **what:** A design showing the framework is domain-agnostic by modelling a whole company: roles typed as identity/function/composite/position (only identity roles get schemas, like client_X); employees as clients with personal agent_instances; always-on shared listeners (like adapters) that spawn discrete orchestrations per task ("Sarah isn't running, she's ready"); authority as conditional filters (policy-owner agents like legal-review-agent injected by trigger conditions) rather than hierarchy; strategy flowing down as intake, decomposing at each level. Cross-cutting agents concluded to be "just agents many workflows call."
- **sources:** docs006_workflow_builder/005_acme_corp_org_chart.md; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Role-vs-Agent; docs006_workflow_builder/006c_org_framework_discussion.md
- **relations:** entity_state_log; relationships; policy filters prefigure legal-content-agent constraints; "strategy as intake" prefigures autonomous mission decomposition.
- **verify-later:** roles/role_assignments tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer)
- **category:** NEW:agent-tree-navigation
- **status-signal:** aspirational
- **status-evidence:** Raw design-session transcript only ("The data model changes are small... The bigger piece of work is the API endpoints and the tree viewer UI"); no implementation claimed, buried inside a 273KB chat-transcript file the rest of the extraction treats as header-scan.
- **what:** A proposal for navigating the `orchestration_states` parent/child tree at massive scale (millions of rows, 8-10 levels deep) without recursive-CTE cost: add an `ltree`-typed `tree_path` column (materialised ancestry path, set cheaply at spawn time by prepending the parent's own path), enrich the existing `subtree_agents` jsonb with rolling status/type/failure counts so a UI can show summaries and only fetch detail on expand, add a `tags` jsonb column (GIN-indexed) for semantic queries ("find all bankrupt fast-food agents" rather than tree position), and a lightweight `agent_tree_index` table (~200 bytes/row, no heavy jsonb blobs) so a million-row tree fits comfortably in cache. Proposed REST API (`/trees/{correlation_id}`, `/agents/{id}/children`, `/agents/{id}/subtree`, `/trees/{id}/search?agent_type=...&status=...`) plus a WebSocket live tree viewer fed from existing Kafka response topics, giving filesystem-like drill-down ("root > uk-economy > fast-food-sector > dominos-agent-47").
- **sources:** docs021_multiclustering/021_2026-02-28-20-03-32-multi-cluster-dispatch-design.txt (sections "The fundamental query patterns" through "The user experience")
- **relations:** Multi-cluster scaling tiers, orchestration_states schema, Agent swarm simulation ideas (this viewer was requested specifically to make the swarm-simulation ideas practically navigable)
- **verify-later:** orchestration_states.tree_path / tags columns; any agent_tree_index table; core-manager tree API endpoints

<!-- SOURCE: U22_recent_small_docs.md -->
### Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer)
- **category:** NEW:agent-tree-navigation
- **status-signal:** aspirational
- **status-evidence:** Raw design-session transcript only ("The data model changes are small... The bigger piece of work is the API endpoints and the tree viewer UI"); no implementation claimed, buried inside a 273KB chat-transcript file the rest of the extraction treats as header-scan.
- **what:** A proposal for navigating the `orchestration_states` parent/child tree at massive scale (millions of rows, 8-10 levels deep) without recursive-CTE cost: add an `ltree`-typed `tree_path` column (materialised ancestry path, set cheaply at spawn time by prepending the parent's own path), enrich the existing `subtree_agents` jsonb with rolling status/type/failure counts so a UI can show summaries and only fetch detail on expand, add a `tags` jsonb column (GIN-indexed) for semantic queries ("find all bankrupt fast-food agents" rather than tree position), and a lightweight `agent_tree_index` table (~200 bytes/row, no heavy jsonb blobs) so a million-row tree fits comfortably in cache. Proposed REST API (`/trees/{correlation_id}`, `/agents/{id}/children`, `/agents/{id}/subtree`, `/trees/{id}/search?agent_type=...&status=...`) plus a WebSocket live tree viewer fed from existing Kafka response topics, giving filesystem-like drill-down ("root > uk-economy > fast-food-sector > dominos-agent-47").
- **sources:** docs021_multiclustering/021_2026-02-28-20-03-32-multi-cluster-dispatch-design.txt (sections "The fundamental query patterns" through "The user experience")
- **relations:** Multi-cluster scaling tiers, orchestration_states schema, Agent swarm simulation ideas (this viewer was requested specifically to make the swarm-simulation ideas practically navigable)
- **verify-later:** orchestration_states.tree_path / tags columns; any agent_tree_index table; core-manager tree API endpoints

<!-- SOURCE: U22_recent_small_docs.md -->
### Agent swarm simulation ideas (never built — hierarchical/fractal use-case brainstorm)
- **category:** NEW:agent-swarm-simulations
- **status-signal:** aspirational
- **status-evidence:** Pure ideation inside a raw chat transcript; closing exchange asks to "report all your ideas into a document... I'd like to use the document as a web page of use-cases... to try and get people interested in triggering their own project ideas" — recorded as marketing/pitch material, not a build plan.
- **what:** A large brainstormed catalogue of speculative applications for the platform's hierarchical spawn/call architecture at extreme scale (up to 1M agents), produced across two rounds. Flat-swarm ideas: an LLM-agent economy simulation (the author's top pick — emergent price equilibrium/monopolies, visually renderable), collaborative Wikipedia cross-fact-checking, a million-agent code-review swarm auditing a large codebase, emergent-language formation between agents with no shared vocabulary, distributed collaborative micro-fiction, adversarial red-team-vs-blue-team war-gaming. Hierarchy-specific ideas exploiting the platform's distinguishing trait (every parent/child stores its result independently, decomposition is semantically meaningful): recursive market-research report trees, organisational/corporate-directive-cascade simulation, fractal ecological modelling (e.g. the Amazon basin), legislation-impact analysis trees, a hierarchical self-debugging swarm for the platform's own Kafka/K8s/Postgres stack, scientific-literature mapping into a queryable tree, supply-chain stress testing, evolutionary/genetic idea-generation trees, plus a further batch: historical counterfactual simulation, language-family evolution trees, musical composition by fractal decomposition, personal health-trajectory modelling, adversarial peer-review simulation, M&A due-diligence trees, climate-migration modelling, argument/debate mapping, ecosystem-succession modelling, judicial case-reasoning trees, disaster-response command-structure coordination, and personal knowledge management. None were built or scoped into a plan.
- **sources:** docs021_multiclustering/021_2026-02-27-18-19-36-million-agent-scaling-plan.txt (the exchanges following "what other impressive things could we do with 1M vaguely intelligent agents?" and "...for the hierarchical/fractal model that I currently have")
- **relations:** Agent hierarchy tree navigation, Multi-cluster scaling tiers (10K/100K/1M), Worker pool architecture
- **verify-later:** n/a (pure ideation, no code or schema artefact)

<!-- SOURCE: U22_recent_small_docs.md -->
### Agent swarm simulation ideas (never built — hierarchical/fractal use-case brainstorm)
- **category:** NEW:agent-swarm-simulations
- **status-signal:** aspirational
- **status-evidence:** Pure ideation inside a raw chat transcript; closing exchange asks to "report all your ideas into a document... I'd like to use the document as a web page of use-cases... to try and get people interested in triggering their own project ideas" — recorded as marketing/pitch material, not a build plan.
- **what:** A large brainstormed catalogue of speculative applications for the platform's hierarchical spawn/call architecture at extreme scale (up to 1M agents), produced across two rounds. Flat-swarm ideas: an LLM-agent economy simulation (the author's top pick — emergent price equilibrium/monopolies, visually renderable), collaborative Wikipedia cross-fact-checking, a million-agent code-review swarm auditing a large codebase, emergent-language formation between agents with no shared vocabulary, distributed collaborative micro-fiction, adversarial red-team-vs-blue-team war-gaming. Hierarchy-specific ideas exploiting the platform's distinguishing trait (every parent/child stores its result independently, decomposition is semantically meaningful): recursive market-research report trees, organisational/corporate-directive-cascade simulation, fractal ecological modelling (e.g. the Amazon basin), legislation-impact analysis trees, a hierarchical self-debugging swarm for the platform's own Kafka/K8s/Postgres stack, scientific-literature mapping into a queryable tree, supply-chain stress testing, evolutionary/genetic idea-generation trees, plus a further batch: historical counterfactual simulation, language-family evolution trees, musical composition by fractal decomposition, personal health-trajectory modelling, adversarial peer-review simulation, M&A due-diligence trees, climate-migration modelling, argument/debate mapping, ecosystem-succession modelling, judicial case-reasoning trees, disaster-response command-structure coordination, and personal knowledge management. None were built or scoped into a plan.
- **sources:** docs021_multiclustering/021_2026-02-27-18-19-36-million-agent-scaling-plan.txt (the exchanges following "what other impressive things could we do with 1M vaguely intelligent agents?" and "...for the hierarchical/fractal model that I currently have")
- **relations:** Agent hierarchy tree navigation, Multi-cluster scaling tiers (10K/100K/1M), Worker pool architecture
- **verify-later:** n/a (pure ideation, no code or schema artefact)

<!-- SOURCE: U19_sql_tables_components.md -->
### Agent definition snapshot/revert via backup table
- **category:** NEW:agent-definition-registry
- **status-signal:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup.
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced.

<!-- SOURCE: U19_sql_tables_components.md -->
### Agent definition snapshot/revert via backup table
- **category:** NEW:agent-definition-registry
- **status-signal:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup.
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decisions 20–21; 010 seed schedules
- **what:** build-pipeline-trigger fires via kafka-scheduler, seeds queue, picks one dispatchable site (skipping sites with claimed items via NOT EXISTS), spawns build-dispatch-loop with await_response:false. Loop claims atomically, processes one item, completes — parallel sites, no batch accumulation, no OOM.
- **sources:** 002(4)#Dispatch Loop and Pipeline Trigger; 004#Entry Points
- **relations:** site-excluded-by-stuck-claim failure; scheduler concurrency groups
- **verify-later:** build-pipeline-trigger pre_query; find_dispatchable_site SQL

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka scheduler (DB-driven heartbeat service)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 010 full deployment reference (migration 066, kustomize, terraform paths)
- **what:** Single-replica Go producer-only service ticking 30s over scheduled_tasks: interval elapsed + concurrency-group capacity + pre_query gating → publish standard orchestrate message (from kafka-scheduler identity, responses to system.scheduler.responses — currently unconsumed). Adding a schedule is an INSERT. Pre-queries provide dynamic input (first row merged into input_data) and gating (no rows = skip). timeout_seconds is the in-flight safety valve; double-fire tolerated via idempotent work-item dedup.
- **sources:** 010 full
- **relations:** build-pipeline-trigger; improvement-sweep; med tasks; batch submitter/retriever placement
- **verify-later:** scheduled_tasks rows; cmd/scheduler/main.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### content-feed-trigger workflow shape bug (array vs object count)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "Fix applied … output_format = 'object' ✓ items_field = 'news_sites.rows' ✓ … Pending verification on next fire" (2026-04-20)
- **what:** The scheduled news trigger was "broken for weeks" not because of routing (generic-agent routing works as designed) but because find_news_sites returned a bare array: check_has_sites read `.count` off an array (empty string → default branch), and the loop crashed on nil when no sites existed. Fixed by output_format object + items_field .rows. General lesson: condition fields need the object {rows,count} shape.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7
- **relations:** owner_agent_type observability gap (why it was misdiagnosed)
- **verify-later:** content-feed-trigger definition current shape

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item claim/retry behaviour and the claim-timeout class
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** W6 FINAL VERIFY: "3.1 failure class: `Claim timed out — handler pod likely died` on all three retried items — dispatch infrastructure, not the template changes; retries recovered."
- **what:** Build items are claimed by the dispatch loop and retried on claim timeout; heavy page builds (19:18–22:45 for six pages) collide with claim durations, producing retried-then-complete items whose error text is retained — read the error class before calling retries healthy. Observed hygiene gaps: `site_work_items.updated_at` stays frozen at insert through claim/retry/completion (same family as the pre-trigger layouts.updated_at); a deploy can release claims mid-flight (claimed→triaged). All parked on the hygiene list, not actioned in-thread.
- **sources:** RUNBOOK_scheme_to_components(50).md#W6-FINAL-VERIFY; w6_03_final_verify.sql; running_notes_scheme_to_components(55).md#Te #Tf #Tp
- **relations:** work-item crafting conventions; debugging (pod health).
- **verify-later:** build dispatch loop claim timeout vs typical build durations; updated_at handling on site_work_items.

<!-- SOURCE: U05_content_quality_linking.md -->
### Dispatch throughput constraints (one-site-per-tick, NOT-EXISTS freeze)
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** running_notes_14(26) Part 9 confirms the mechanism; HANDOFF_2026-06-15(2) §5: "Rebuild pipeline takes MANY HOURS … NOT investigated".
- **what:** The build-dispatch-loop is one-site-per-tick (LIMIT 1, spawned per scheduler tick, ~5 items then exits) and excludes a site entirely while ANY of its items is claimed — so items serialise within a site and a dead handler freezes the whole site for the claim-timeout window. Catalogued as Family J with candidate levers (per-site bounded concurrency, per-item exclusion, shorter reaper window, trigger cadence) plus the standing speed-up TODO (batches take hours; single index rebuild ~610–770s). Parked, never closed in these docs.
- **sources:** running_notes_14(26).md#part-9; HANDOFF_2026-06-15(2).md#5; running_notes_17(21).md#missing-game
- **relations:** claimed-item-timeout reaper; operational rule "don't roll the chassis image while a batch drains".
- **verify-later:** build-dispatch-loop pre_query/LIMIT + NOT-EXISTS clause; scheduled_tasks build-pipeline-trigger cadence.

<!-- SOURCE: U11_traffic_probe.md -->
### Scheduler fires one message per tick — pre_query is a gate, not fan-out
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(c): "DESIGN CORRECTED by a real finding: the scheduler fires ONE message per tick — it does NOT fan out pre_query rows (ctx line 5236; thunder-monitor does in-agent loop fan-out, not scheduler fan-out)."
- **what:** A platform fact established from live source and used to correct the collector design: scheduled_tasks.pre_query does not produce per-row dispatch; the live improvement-sweep/thunder-monitor pattern is a count>0 GATE with the fired agent doing in-agent loop fan-out. The intent collector was rewritten from "collect one site from input" to a single self-querying loop-all action accordingly (complexity in Go, one-step workflow); the migration's per-row pre_query was superseded. Also the thunder-monitor convention: INSERT scheduled tasks DISABLED until the action is deployed.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-c, intent_events_migration(1).sql#scheduled-collector (gate form), deploy_setup/working_dir/intent_events_migration.sql (family-delta: superseded fan-out form)
- **relations:** intent collection topology, scheduler-and-tasks doc 010
- **verify-later:** kafka-scheduler dispatch code path (one fire per tick)

<!-- SOURCE: U12_docs024_archives.md -->
### CTE-only scheduled tasks pattern ("Always Return a Row" rule)
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive `011b_scheduler_and_tasks_guide.md` (a later revision than `011_kafka_scheduler_guide.md`, which is byte-identical to live) has a full section on this; none of it appears in live `010_scheduler_and_tasks.md`.
- **what:** Some scheduled tasks do their real work directly inside the pre_query's CTEs rather than triggering an agent — but the scheduler still requires the SELECT to return at least one row, or `last_triggered_at`/`last_completed_at` never advance, silently breaking firing cadence and concurrency-group accounting. This is a documented, previously-hit production bug pattern completely absent from the current live scheduler doc.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"Pre-Queries", #"The fire_message Column"; docs024_key_docs_latest/010_scheduler_and_tasks.md (confirmed absent)
- **relations:** concurrency-group starvation; last_completed_at ownership
- **verify-later:** `SELECT name, pre_query FROM scheduled_tasks WHERE fire_message = false` for current CTE-only tasks.

<!-- SOURCE: U12_docs024_archives.md -->
### Concurrency group starvation problem and prevention rules
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive documents a real incident ("the original maintenance group had both claimed-item-timeout and database-cleanup. When database-cleanup stalled, it blocked claim resets, which blocked the entire pipeline") and gives four prevention rules; entirely absent from live doc.
- **what:** Tasks sharing a `concurrency_group` can starve each other if one never updates `last_completed_at`, permanently occupying the group's `max_concurrent` slot. Prevention: set `timeout_seconds < interval_seconds`, never group unrelated tasks together, ensure every completion path updates `last_completed_at`.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"The Group Starvation Problem", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; last_completed_at ownership
- **verify-later:** query current `scheduled_tasks` group assignments against the archive's "Recommended Group Assignments" table.

<!-- SOURCE: U12_docs024_archives.md -->
### last_completed_at ownership contract and fire_message known-gap
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive explicitly documents: "The scheduler Go code does not currently read this column [fire_message]. It always sends a Kafka message"; none of these operational caveats appear in live doc.
- **what:** Agent-triggered scheduled tasks must include an explicit `notify_scheduler` step on every completion path to set `last_completed_at`; the scheduler itself never sets this column and never reads `fire_message`, flagged as a known low-priority gap.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"last_completed_at — Who Updates It?", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; concurrency group starvation
- **verify-later:** `grep -rn "fire_message" cmd/scheduler/` to check if the Go scheduler now reads this column.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Private inert pipeline statuses pattern
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "inertness matrix scores 0 against all six sweeps in both states" (PLAN_fixloop_pilot.md §F0.1d)
- **what:** A reusable pattern for giving a new pipeline namespace statuses that no existing sweep or claim path recognizes, so it is inert "by construction" rather than by luck of anchor-site choice. The diagnose pipeline uses `awaiting_diagnosis` (queued) → `diagnosing` (in-flight), claimed atomically via `UPDATE ... FOR UPDATE SKIP LOCKED ... RETURNING` rather than the shared `claim_work_item` (which only claims `triaged|approved`). Because opting out of shared sweeps also opts out of their cleanup, the private-status loop must reap its own dead runs.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d
- **relations:** diagnose-dispatch-loop; pipeline-blind dispatch surfaces (discovered platform gap)
- **verify-later:** site_work_items.status values in the diagnose pipeline; reap_stuck step logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pipeline-blind dispatch surfaces (discovered platform defect)
- **category:** scheduler-and-tasks
- **status-signal:** deployed (documented, not fixed — routed elsewhere)
- **status-evidence:** "Nothing in the relay filters work items by pipeline where it matters" (RUNBOOK(10)#Inherited gotchas); "Routed to the builder thread, not fixed here" (0NN_diagnose_dispatch_loop.sql#header)
- **what:** `build-dispatch-loop`'s `load_items` step and `build-pipeline-trigger`'s `find_dispatchable_site` query both lack any `item_pipeline`/pipeline filter, so any item of any pipeline on a claimable site gets dispatched to whatever handler_agent it names — this is the only reason the `maintenance` pipeline gets dispatched at all. `triage_detect_items` compounds this: it claims on `status='detected'` with no pipeline filter and rewrites `pipeline` to `'build'`, while its own comment falsely claims a filter exists. Fixing `build-dispatch-loop` naively would orphan the maintenance pipeline, so this was reported to the builder thread rather than fixed by the fix-loop team.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 5
- **relations:** private inert pipeline statuses pattern; two intake paths disagreement
- **verify-later:** build-dispatch-loop.load_items config; build-pipeline-trigger.find_dispatchable_site query; triage_detect_items query

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose-dispatch-loop (automatic dispatch)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "F0.1d — ✅ LANDED 2026-07-09, SHIPPED DISABLED" (PLAN_fixloop_pilot.md §F0.1d); "ships enabled=false on purpose"
- **what:** An `agent_definitions` orchestrator agent that claims one `awaiting_diagnosis` item on a 60s tick (via `diagnose-pipeline-trigger` scheduled task, `max_concurrent=1`), atomically moves it to `diagnosing`, spawns `diagnose-orchestrator`, and reaps its own runs older than 75 minutes as `failed`. Deliberately shipped with the scheduled task disabled until the chassis image is live and the benchmark's blinding is confirmed, since enabling it would let the loop claim and consume the benchmark item before blinding could be verified.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#CURRENT POSITION history
- **relations:** private inert pipeline statuses pattern; needs_diagnosis intake route
- **verify-later:** `scheduled_tasks.enabled` for name='diagnose-pipeline-trigger' (should still be false unless deliberately turned on)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper mechanisms and the work-item-claim reaper gap
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "Correction (2026-05-21). An earlier draft ... assumed the reapers were Go code ... They are not: the reapers are SQL pre_query entries in the scheduled_tasks table"
- **what:** Three/four reaper-like mechanisms recover stuck state at different layers: stuck-orchestration reaper (backed by scheduled_tasks SQL entries), `FailWorkItemAction`'s three retry paths, and `agent-job-cleanup` CronJob (k8s housekeeping only). The gap: no periodic sweep exists for work items stuck at `status='claimed'` when a pod dies uncleanly — `idx_swi_claimed` index exists for exactly this query but nothing uses it.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-2, js_snippets_news_gaswholesalers/old/reapers_and_stuck_state_recovery.md
- **relations:** collected_data/OOM bloat; two rerender trigger paths
- **verify-later:** scheduled_tasks table pre_query entries, idx_swi_claimed index

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper-location framing correction (superseded documentary claim)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** explicit dated correction: "An earlier draft of this section assumed the reapers were Go code in the coordinator. They are not"
- **what:** The original analysis framed all reaper logic as Go code in the chassis coordinator. A 2026-05-21 follow-up established the confirmed scheduled reapers are actually SQL `pre_query` entries in `scheduled_tasks`, with the Go on-access check being secondary.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md, js_snippets_news_gaswholesalers/old/TODO_remaining_work_2026-05-21.md
- **relations:** Reaper mechanisms and the work-item-claim reaper gap
- **verify-later:** grep/inspect `pre_query`; `scheduled_tasks`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Stale orchestration sweeper/reaper
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 001(0) "Timeout handling uses in-process goroutines. These die when pods restart … This is the #1 cause of pipeline stalls"
- **what:** In-process timeout goroutines die on pod restart, stranding orchestrations in AWAITING_RESPONSES. A periodic DB sweep (every 60s, `FOR UPDATE SKIP LOCKED`) classifies each expired awaited request: child completed (synthesize the lost response), child failed (forward), or no child/still-running (retry up to 3 then fail parent). The `stale-orchestration-reaper` scheduled task also fails 24h-stale orchestrations.
- **sources:** WM/001_development_guide(0).md#stale-orchestration-sweeper, WM/016_debugging_guide_v2_44.md#4, WM/007_adoption_pipeline_v3.md#known-issue-zombie-dispatch-loop-pods
- **relations:** timeout chain; work item lifecycle; awaited_requests
- **verify-later:** orchestration_states; awaited_requests; scheduled_tasks stale-orchestration-reaper

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### claimed-item-timeout & timeout chain
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §7 "Three timeouts interact and must be ordered correctly … claim_timeout > call_handler timeout > workflow timeout"
- **what:** A `claimed-item-timeout` scheduled task resets long-claimed items; three timeouts must stay ordered, else two handlers run one item. A two-phase reset (15-min evidence-based, 40-min blind) is used; the evidence check can produce false-positive completions.
- **sources:** WM/016_debugging_guide_v2_44.md#7, WM/016_debugging_guide_v2_44.md#9, WM/007_adoption_pipeline_v3.md#implementation-fixes-schema-notes-from-028j-handoff
- **relations:** dispatch loop; stale orchestration sweeper; work item lifecycle
- **verify-later:** claimed-item-timeout pre_query; scheduled_tasks; idle_timeout_seconds

<!-- SOURCE: U19_sql_tables_components.md -->
### Kafka scheduler and scheduled_tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Table DDL (066_kafka_scheduler) plus a long operational history of seeded tasks: build-pipeline-trigger, improvement-sweep, claimed-item-timeout, feasibility-recheck, content-feed-refresh, database-cleanup, vet-*, med-*, ch-enrichment, health checks, archiver.
- **what:** Interval-based scheduling in Postgres: each row names a target agent/topic, input_data, interval_seconds, timeout, and concurrency_group/max_concurrent (group-wide in-flight cap). The scheduler publishes Kafka trigger messages; last_triggered_at/last_completed_at implement a no-refire guard (with known operational pitfalls when nothing sets last_completed_at for fire-and-forget tasks — mitigated by shorter timeout windows).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql
- **relations:** pre_query SQL-worker pattern; every pipeline's periodic trigger.
- **verify-later:** kafka-scheduler service; fire_message column semantics.

<!-- SOURCE: U19_sql_tables_components.md -->
### pre_query SQL-worker pattern and self-healing tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Iterated in place: pre_query CTE UPDATEs doing the work, then `WHERE 1=0` / `HAVING COUNT(*) > 0` variants to control whether a Kafka message fires; vet-cleanup broadened to fail stuck AWAITING_RESPONSES orchestrations and reset orphaned collection tasks.
- **what:** scheduled_tasks.pre_query is a full worker channel, not just a gate: SQL that returns rows merges into input_data and fires the message; returning zero rows skips the tick. Maintenance tasks exploit this to run entire cleanup UPDATEs inside the pre_query (claimed-item reset, blocked-item promotion, orchestration failing, database cleanup) with row-suppression idioms deciding whether anything downstream is triggered (fire_message=false for pure-SQL tasks).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle and #vet-cleanup and #database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql
- **relations:** scheduler; claimed-item timeout; database cleanup.
- **verify-later:** scheduler's pre_query evaluation code; fire_message flag.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-training-monitor + worker (probe/classify/reconcile/decommission)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** NOTES(39) §9 "Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths) … Terminal/decommission branch still never run live … Not enabled"
- **what:** A periodic orchestrator (`thunder-training-monitor`, migration 108) that runs `find_active_training_instances → loop(spawn_worker → call_worker)` every 5 min (scheduled_tasks row, inserted DISABLED, gated pre_query). Each `thunder-training-monitor-worker` (migration 107) probes a box via the adapter's `ssh_get_status`, classifies run.sh markers (ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN) via `classify_training_probe`, reconciles `training_runs` via `mark_training_run_terminal`, and decommissions on terminal verdicts. Deliberately separate from the reaper (different dependencies); closes the running→complete/failed reconcile gap. Enabling it is gated on the upload path proving DONE⟹durable.
- **sources:** phase5/108_thunder_training_monitor_orchestrator.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150, #update-2026-06-04-1x; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** reuses ssh_get_status + dispatch_thunder_decommission; depends on unreachable counter; gated by RUNBOOK step 6
- **verify-later:** agent_definitions thunder-training-monitor (c3b4c052) / -worker (470c6b3f); 5 actions incl find_active_training_instances; scheduled_tasks

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder unreachable-probe counter
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 106 SQL header "counts CONSECUTIVE unreachable probes and only treats the instance as 'lost' … once the count crosses a threshold"
- **what:** Migration 106 adds `consecutive_unreachable_probes` + `last_probe_at` to `thunder_instances` so the monitor can distinguish a transient SSH blip from a truly-lost box. Each scheduler tick is a fresh sub-agent that can't hold count in memory, so the streak lives on the row: the `record_probe_streak` action bumps on unreachable (route to lost/decommission at threshold, default 3) and resets to 0 on any reachable probe.
- **sources:** phase5/106_thunder_unreachable_counter.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150 (Counter step)
- **relations:** part of thunder-training-monitor; keeps the classifier action pure
- **verify-later:** thunder_instances.consecutive_unreachable_probes; record_probe_streak_action.go

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P4 off-box collection (intent_events + CollectIntentEventsAction)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "Migration applied (CREATE TABLE + 3 indexes + INSERT 0 1 task)"; action "VERIFIED against live source + registered"; but agent deploy fields still to confirm and enable order pending.
- **what:** The cluster pulls intent over key-gated HTTPS with NO adapter and NO SSH. `intent_events` table (engine_event_id UNIQUE = structural idempotency, CHECK on kind/value len, host→site_id resolve, checkpoint = max(event_created_at) with no extra storage). `collect_intent_events` is a SINGLE Go action that self-queries all VM backend sites and loops (parameterised upserts), registered in GlobalActionRegistry (Category "data", IsLocal). Ingest contract: parameterised SQL only, per-line shape checks, burst dedupe, NFC normalisation + lowercasing here.
- **sources:** traffic_probe_plan(11).md#p4, traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-c
- **relations:** driven by intent-collection-orchestrator/intent-collector agents; extended with collectSiteStats + access-digest pull
- **verify-later:** intent_events_migration.sql; intent_collector_actions.go; registry.go DATA region

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded checkpoint-JSON / events-per-1k ranking (early P4)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** plan(1)/(4)/(5) P4 "checkpoint JSON, compute events-per-1k, rank domains"; plan(11) now "idempotent via unique engine_event_id; no extra checkpoint storage — since=max(event_created_at)".
- **what:** Early P4 phrasing planned an explicit checkpoint-JSON file to track collection progress and a direct events-per-1k rank. Dropped in favour of structural idempotency (unique engine_event_id) with the checkpoint derived as since=max(event_created_at) — no extra checkpoint storage. Ranking became a set of read-only SQL queries.
- **sources:** traffic_probe_plan(4).md#phases, traffic_probe_plan(1).md#phases, traffic_probe_plan(11).md#p4
- **relations:** replaced by intent_events unique-id design + intent_ranking_queries
- **verify-later:** intent_events.engine_event_id UNIQUE

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-collection-orchestrator + intent-collector agents
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** intent_collector_agents SQL headers "intent-collection-orchestrator + intent-collector (P4) … mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim"; running_notes 2026-06-13(g) INSERT bug fixed.
- **what:** A thin wrapper-orchestrator (spawn_collector → call_collector → complete, no substantive in-chassis work) that spawns the `intent-collector` task worker (one step: collect_intent_events, processing_mode "task"). Infra fields (image docker.io/aqls/agent-chassis v1.0.1063, resources, health_config, business-intel topics, delegation) copied verbatim from the med-export pair. Reached by the scheduler via target_topic=system.agent.generic.requests by agent_type. Idempotency uses `ON CONFLICT (type, version)`.
- **sources:** deploy_setup/working_dir/intent_collector_agents(3).sql#header, deploy_setup/working_dir/intent_collector_agents(1).sql, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** identical to live intent_collector_agents(2).sql; wrapper-orchestrator requirement; replaces a single in-pod collector
- **verify-later:** agent_definitions rows intent-collection-orchestrator / intent-collector

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Retention prune timer
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Added to setup.sh: RETENTION_DAYS param (default 90) + site-engine-prune.service/.timer (daily find-delete of old events-*.jsonl)".
- **what:** Because daily JSONL IS the rotation, logrotate on engine files would race the open handle; instead setup.sh installs a `site-engine-prune` systemd service+timer that daily find-deletes `events-*.jsonl` older than RETENTION_DAYS (default 90). nginx logs keep their existing size-based logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_runbook(12).md#3.5
- **relations:** part of the privacy posture
- **verify-later:** setup.sh site-engine-prune.timer; RETENTION_DAYS

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### claimed-item-timeout evidence-gated completion + reset (Lever A/C) — avoided building a duplicate watchdog
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 12: "'Auto-completed...' is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper... `migration_claimed_item_timeout_evidence_check.sql` ... is essentially Option A + Lever C, already authored... the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it." Part 12 addenda confirm the v2 migration (page_components-based evidence, not the untrustworthy `build_status='deployed'` flag) applied and verified live, plus the companion `pageHasComponents` deploy-guard (Option B) delivered.
- **what:** A `claimed-item-timeout` scheduled task's `pre_query` already implements both (a) evidence-gated auto-completion of stuck claims (only complete if the specific artefact shows positive evidence) and (b) a stale-claim reset-to-`triaged`/`failed` after a timeout with attempt counting. Mid-investigation, an agent nearly built a brand-new "reset stale claims" watchdog before discovering this — a documented reuse-over-build catch. The evidence signal itself evolved further: from trusting `pages.deployed_at`/`build_status='deployed'` (provably lying, per the homepage case) to checking `page_components.updated_at > claimed_at` directly.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 12–12 addenda
- **relations:** A4 homepage root cause (above); sectionless-page durability stack (below); dispatch throughput (Family J)
- **verify-later:** current `claimed-item-timeout` pre_query SQL in `scheduled_tasks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dispatch throughput bottleneck (Family J) — one-site-per-tick, NOT-EXISTS-blocked
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J: "the dispatcher is **one-site-per-tick** (selection `LIMIT 1`... processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'`... line 276)... Standard manual unstick for now... **To investigate in the separate thread.**"
- **what:** Multi-tool/multi-game adoption sites drain over hours, appearing stalled, because the build-dispatch mechanism processes one site per scheduler tick and blocks an entire site's queue while any single item on it is claimed (no bounded concurrency, no per-item exclusion). A dead handler leaving a stale claim freezes the whole site until a reaper resets it. Explicitly spun out as a separate, not-yet-investigated thread rather than fixed within this arc; running_notes_17(16) later notes it's still an open TODO ("SPEED UP the rebuild pipeline... Not yet investigated").
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J; running_notes_14(20) Part 9; running_notes_17(16) "Missing-game... + speed TODO"
- **relations:** claimed-item-timeout evidence-check reliability mini-project; A1 tool/game deploy gap
- **verify-later:** `build-pipeline-trigger` dispatcher current selection logic (`LIMIT 1`, NOT EXISTS clause, line ~276 at time of writing).

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decisions 20–21; 010 seed schedules
- **what:** build-pipeline-trigger fires via kafka-scheduler, seeds queue, picks one dispatchable site (skipping sites with claimed items via NOT EXISTS), spawns build-dispatch-loop with await_response:false. Loop claims atomically, processes one item, completes — parallel sites, no batch accumulation, no OOM.
- **sources:** 002(4)#Dispatch Loop and Pipeline Trigger; 004#Entry Points
- **relations:** site-excluded-by-stuck-claim failure; scheduler concurrency groups
- **verify-later:** build-pipeline-trigger pre_query; find_dispatchable_site SQL

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka scheduler (DB-driven heartbeat service)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 010 full deployment reference (migration 066, kustomize, terraform paths)
- **what:** Single-replica Go producer-only service ticking 30s over scheduled_tasks: interval elapsed + concurrency-group capacity + pre_query gating → publish standard orchestrate message (from kafka-scheduler identity, responses to system.scheduler.responses — currently unconsumed). Adding a schedule is an INSERT. Pre-queries provide dynamic input (first row merged into input_data) and gating (no rows = skip). timeout_seconds is the in-flight safety valve; double-fire tolerated via idempotent work-item dedup.
- **sources:** 010 full
- **relations:** build-pipeline-trigger; improvement-sweep; med tasks; batch submitter/retriever placement
- **verify-later:** scheduled_tasks rows; cmd/scheduler/main.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### content-feed-trigger workflow shape bug (array vs object count)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "Fix applied … output_format = 'object' ✓ items_field = 'news_sites.rows' ✓ … Pending verification on next fire" (2026-04-20)
- **what:** The scheduled news trigger was "broken for weeks" not because of routing (generic-agent routing works as designed) but because find_news_sites returned a bare array: check_has_sites read `.count` off an array (empty string → default branch), and the loop crashed on nil when no sites existed. Fixed by output_format object + items_field .rows. General lesson: condition fields need the object {rows,count} shape.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7
- **relations:** owner_agent_type observability gap (why it was misdiagnosed)
- **verify-later:** content-feed-trigger definition current shape

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item claim/retry behaviour and the claim-timeout class
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** W6 FINAL VERIFY: "3.1 failure class: `Claim timed out — handler pod likely died` on all three retried items — dispatch infrastructure, not the template changes; retries recovered."
- **what:** Build items are claimed by the dispatch loop and retried on claim timeout; heavy page builds (19:18–22:45 for six pages) collide with claim durations, producing retried-then-complete items whose error text is retained — read the error class before calling retries healthy. Observed hygiene gaps: `site_work_items.updated_at` stays frozen at insert through claim/retry/completion (same family as the pre-trigger layouts.updated_at); a deploy can release claims mid-flight (claimed→triaged). All parked on the hygiene list, not actioned in-thread.
- **sources:** RUNBOOK_scheme_to_components(50).md#W6-FINAL-VERIFY; w6_03_final_verify.sql; running_notes_scheme_to_components(55).md#Te #Tf #Tp
- **relations:** work-item crafting conventions; debugging (pod health).
- **verify-later:** build dispatch loop claim timeout vs typical build durations; updated_at handling on site_work_items.

<!-- SOURCE: U05_content_quality_linking.md -->
### Dispatch throughput constraints (one-site-per-tick, NOT-EXISTS freeze)
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** running_notes_14(26) Part 9 confirms the mechanism; HANDOFF_2026-06-15(2) §5: "Rebuild pipeline takes MANY HOURS … NOT investigated".
- **what:** The build-dispatch-loop is one-site-per-tick (LIMIT 1, spawned per scheduler tick, ~5 items then exits) and excludes a site entirely while ANY of its items is claimed — so items serialise within a site and a dead handler freezes the whole site for the claim-timeout window. Catalogued as Family J with candidate levers (per-site bounded concurrency, per-item exclusion, shorter reaper window, trigger cadence) plus the standing speed-up TODO (batches take hours; single index rebuild ~610–770s). Parked, never closed in these docs.
- **sources:** running_notes_14(26).md#part-9; HANDOFF_2026-06-15(2).md#5; running_notes_17(21).md#missing-game
- **relations:** claimed-item-timeout reaper; operational rule "don't roll the chassis image while a batch drains".
- **verify-later:** build-dispatch-loop pre_query/LIMIT + NOT-EXISTS clause; scheduled_tasks build-pipeline-trigger cadence.

<!-- SOURCE: U11_traffic_probe.md -->
### Scheduler fires one message per tick — pre_query is a gate, not fan-out
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(c): "DESIGN CORRECTED by a real finding: the scheduler fires ONE message per tick — it does NOT fan out pre_query rows (ctx line 5236; thunder-monitor does in-agent loop fan-out, not scheduler fan-out)."
- **what:** A platform fact established from live source and used to correct the collector design: scheduled_tasks.pre_query does not produce per-row dispatch; the live improvement-sweep/thunder-monitor pattern is a count>0 GATE with the fired agent doing in-agent loop fan-out. The intent collector was rewritten from "collect one site from input" to a single self-querying loop-all action accordingly (complexity in Go, one-step workflow); the migration's per-row pre_query was superseded. Also the thunder-monitor convention: INSERT scheduled tasks DISABLED until the action is deployed.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-c, intent_events_migration(1).sql#scheduled-collector (gate form), deploy_setup/working_dir/intent_events_migration.sql (family-delta: superseded fan-out form)
- **relations:** intent collection topology, scheduler-and-tasks doc 010
- **verify-later:** kafka-scheduler dispatch code path (one fire per tick)

<!-- SOURCE: U12_docs024_archives.md -->
### CTE-only scheduled tasks pattern ("Always Return a Row" rule)
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive `011b_scheduler_and_tasks_guide.md` (a later revision than `011_kafka_scheduler_guide.md`, which is byte-identical to live) has a full section on this; none of it appears in live `010_scheduler_and_tasks.md`.
- **what:** Some scheduled tasks do their real work directly inside the pre_query's CTEs rather than triggering an agent — but the scheduler still requires the SELECT to return at least one row, or `last_triggered_at`/`last_completed_at` never advance, silently breaking firing cadence and concurrency-group accounting. This is a documented, previously-hit production bug pattern completely absent from the current live scheduler doc.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"Pre-Queries", #"The fire_message Column"; docs024_key_docs_latest/010_scheduler_and_tasks.md (confirmed absent)
- **relations:** concurrency-group starvation; last_completed_at ownership
- **verify-later:** `SELECT name, pre_query FROM scheduled_tasks WHERE fire_message = false` for current CTE-only tasks.

<!-- SOURCE: U12_docs024_archives.md -->
### Concurrency group starvation problem and prevention rules
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive documents a real incident ("the original maintenance group had both claimed-item-timeout and database-cleanup. When database-cleanup stalled, it blocked claim resets, which blocked the entire pipeline") and gives four prevention rules; entirely absent from live doc.
- **what:** Tasks sharing a `concurrency_group` can starve each other if one never updates `last_completed_at`, permanently occupying the group's `max_concurrent` slot. Prevention: set `timeout_seconds < interval_seconds`, never group unrelated tasks together, ensure every completion path updates `last_completed_at`.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"The Group Starvation Problem", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; last_completed_at ownership
- **verify-later:** query current `scheduled_tasks` group assignments against the archive's "Recommended Group Assignments" table.

<!-- SOURCE: U12_docs024_archives.md -->
### last_completed_at ownership contract and fire_message known-gap
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive explicitly documents: "The scheduler Go code does not currently read this column [fire_message]. It always sends a Kafka message"; none of these operational caveats appear in live doc.
- **what:** Agent-triggered scheduled tasks must include an explicit `notify_scheduler` step on every completion path to set `last_completed_at`; the scheduler itself never sets this column and never reads `fire_message`, flagged as a known low-priority gap.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"last_completed_at — Who Updates It?", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; concurrency group starvation
- **verify-later:** `grep -rn "fire_message" cmd/scheduler/` to check if the Go scheduler now reads this column.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Private inert pipeline statuses pattern
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "inertness matrix scores 0 against all six sweeps in both states" (PLAN_fixloop_pilot.md §F0.1d)
- **what:** A reusable pattern for giving a new pipeline namespace statuses that no existing sweep or claim path recognizes, so it is inert "by construction" rather than by luck of anchor-site choice. The diagnose pipeline uses `awaiting_diagnosis` (queued) → `diagnosing` (in-flight), claimed atomically via `UPDATE ... FOR UPDATE SKIP LOCKED ... RETURNING` rather than the shared `claim_work_item` (which only claims `triaged|approved`). Because opting out of shared sweeps also opts out of their cleanup, the private-status loop must reap its own dead runs.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d
- **relations:** diagnose-dispatch-loop; pipeline-blind dispatch surfaces (discovered platform gap)
- **verify-later:** site_work_items.status values in the diagnose pipeline; reap_stuck step logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pipeline-blind dispatch surfaces (discovered platform defect)
- **category:** scheduler-and-tasks
- **status-signal:** deployed (documented, not fixed — routed elsewhere)
- **status-evidence:** "Nothing in the relay filters work items by pipeline where it matters" (RUNBOOK(10)#Inherited gotchas); "Routed to the builder thread, not fixed here" (0NN_diagnose_dispatch_loop.sql#header)
- **what:** `build-dispatch-loop`'s `load_items` step and `build-pipeline-trigger`'s `find_dispatchable_site` query both lack any `item_pipeline`/pipeline filter, so any item of any pipeline on a claimable site gets dispatched to whatever handler_agent it names — this is the only reason the `maintenance` pipeline gets dispatched at all. `triage_detect_items` compounds this: it claims on `status='detected'` with no pipeline filter and rewrites `pipeline` to `'build'`, while its own comment falsely claims a filter exists. Fixing `build-dispatch-loop` naively would orphan the maintenance pipeline, so this was reported to the builder thread rather than fixed by the fix-loop team.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 5
- **relations:** private inert pipeline statuses pattern; two intake paths disagreement
- **verify-later:** build-dispatch-loop.load_items config; build-pipeline-trigger.find_dispatchable_site query; triage_detect_items query

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose-dispatch-loop (automatic dispatch)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "F0.1d — ✅ LANDED 2026-07-09, SHIPPED DISABLED" (PLAN_fixloop_pilot.md §F0.1d); "ships enabled=false on purpose"
- **what:** An `agent_definitions` orchestrator agent that claims one `awaiting_diagnosis` item on a 60s tick (via `diagnose-pipeline-trigger` scheduled task, `max_concurrent=1`), atomically moves it to `diagnosing`, spawns `diagnose-orchestrator`, and reaps its own runs older than 75 minutes as `failed`. Deliberately shipped with the scheduled task disabled until the chassis image is live and the benchmark's blinding is confirmed, since enabling it would let the loop claim and consume the benchmark item before blinding could be verified.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#CURRENT POSITION history
- **relations:** private inert pipeline statuses pattern; needs_diagnosis intake route
- **verify-later:** `scheduled_tasks.enabled` for name='diagnose-pipeline-trigger' (should still be false unless deliberately turned on)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper mechanisms and the work-item-claim reaper gap
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "Correction (2026-05-21). An earlier draft ... assumed the reapers were Go code ... They are not: the reapers are SQL pre_query entries in the scheduled_tasks table"
- **what:** Three/four reaper-like mechanisms recover stuck state at different layers: stuck-orchestration reaper (backed by scheduled_tasks SQL entries), `FailWorkItemAction`'s three retry paths, and `agent-job-cleanup` CronJob (k8s housekeeping only). The gap: no periodic sweep exists for work items stuck at `status='claimed'` when a pod dies uncleanly — `idx_swi_claimed` index exists for exactly this query but nothing uses it.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-2, js_snippets_news_gaswholesalers/old/reapers_and_stuck_state_recovery.md
- **relations:** collected_data/OOM bloat; two rerender trigger paths
- **verify-later:** scheduled_tasks table pre_query entries, idx_swi_claimed index

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper-location framing correction (superseded documentary claim)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** explicit dated correction: "An earlier draft of this section assumed the reapers were Go code in the coordinator. They are not"
- **what:** The original analysis framed all reaper logic as Go code in the chassis coordinator. A 2026-05-21 follow-up established the confirmed scheduled reapers are actually SQL `pre_query` entries in `scheduled_tasks`, with the Go on-access check being secondary.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md, js_snippets_news_gaswholesalers/old/TODO_remaining_work_2026-05-21.md
- **relations:** Reaper mechanisms and the work-item-claim reaper gap
- **verify-later:** grep/inspect `pre_query`; `scheduled_tasks`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Stale orchestration sweeper/reaper
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 001(0) "Timeout handling uses in-process goroutines. These die when pods restart … This is the #1 cause of pipeline stalls"
- **what:** In-process timeout goroutines die on pod restart, stranding orchestrations in AWAITING_RESPONSES. A periodic DB sweep (every 60s, `FOR UPDATE SKIP LOCKED`) classifies each expired awaited request: child completed (synthesize the lost response), child failed (forward), or no child/still-running (retry up to 3 then fail parent). The `stale-orchestration-reaper` scheduled task also fails 24h-stale orchestrations.
- **sources:** WM/001_development_guide(0).md#stale-orchestration-sweeper, WM/016_debugging_guide_v2_44.md#4, WM/007_adoption_pipeline_v3.md#known-issue-zombie-dispatch-loop-pods
- **relations:** timeout chain; work item lifecycle; awaited_requests
- **verify-later:** orchestration_states; awaited_requests; scheduled_tasks stale-orchestration-reaper

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### claimed-item-timeout & timeout chain
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §7 "Three timeouts interact and must be ordered correctly … claim_timeout > call_handler timeout > workflow timeout"
- **what:** A `claimed-item-timeout` scheduled task resets long-claimed items; three timeouts must stay ordered, else two handlers run one item. A two-phase reset (15-min evidence-based, 40-min blind) is used; the evidence check can produce false-positive completions.
- **sources:** WM/016_debugging_guide_v2_44.md#7, WM/016_debugging_guide_v2_44.md#9, WM/007_adoption_pipeline_v3.md#implementation-fixes-schema-notes-from-028j-handoff
- **relations:** dispatch loop; stale orchestration sweeper; work item lifecycle
- **verify-later:** claimed-item-timeout pre_query; scheduled_tasks; idle_timeout_seconds

<!-- SOURCE: U19_sql_tables_components.md -->
### Kafka scheduler and scheduled_tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Table DDL (066_kafka_scheduler) plus a long operational history of seeded tasks: build-pipeline-trigger, improvement-sweep, claimed-item-timeout, feasibility-recheck, content-feed-refresh, database-cleanup, vet-*, med-*, ch-enrichment, health checks, archiver.
- **what:** Interval-based scheduling in Postgres: each row names a target agent/topic, input_data, interval_seconds, timeout, and concurrency_group/max_concurrent (group-wide in-flight cap). The scheduler publishes Kafka trigger messages; last_triggered_at/last_completed_at implement a no-refire guard (with known operational pitfalls when nothing sets last_completed_at for fire-and-forget tasks — mitigated by shorter timeout windows).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql
- **relations:** pre_query SQL-worker pattern; every pipeline's periodic trigger.
- **verify-later:** kafka-scheduler service; fire_message column semantics.

<!-- SOURCE: U19_sql_tables_components.md -->
### pre_query SQL-worker pattern and self-healing tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Iterated in place: pre_query CTE UPDATEs doing the work, then `WHERE 1=0` / `HAVING COUNT(*) > 0` variants to control whether a Kafka message fires; vet-cleanup broadened to fail stuck AWAITING_RESPONSES orchestrations and reset orphaned collection tasks.
- **what:** scheduled_tasks.pre_query is a full worker channel, not just a gate: SQL that returns rows merges into input_data and fires the message; returning zero rows skips the tick. Maintenance tasks exploit this to run entire cleanup UPDATEs inside the pre_query (claimed-item reset, blocked-item promotion, orchestration failing, database cleanup) with row-suppression idioms deciding whether anything downstream is triggered (fire_message=false for pure-SQL tasks).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle and #vet-cleanup and #database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql
- **relations:** scheduler; claimed-item timeout; database cleanup.
- **verify-later:** scheduler's pre_query evaluation code; fire_message flag.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-training-monitor + worker (probe/classify/reconcile/decommission)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** NOTES(39) §9 "Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths) … Terminal/decommission branch still never run live … Not enabled"
- **what:** A periodic orchestrator (`thunder-training-monitor`, migration 108) that runs `find_active_training_instances → loop(spawn_worker → call_worker)` every 5 min (scheduled_tasks row, inserted DISABLED, gated pre_query). Each `thunder-training-monitor-worker` (migration 107) probes a box via the adapter's `ssh_get_status`, classifies run.sh markers (ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN) via `classify_training_probe`, reconciles `training_runs` via `mark_training_run_terminal`, and decommissions on terminal verdicts. Deliberately separate from the reaper (different dependencies); closes the running→complete/failed reconcile gap. Enabling it is gated on the upload path proving DONE⟹durable.
- **sources:** phase5/108_thunder_training_monitor_orchestrator.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150, #update-2026-06-04-1x; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** reuses ssh_get_status + dispatch_thunder_decommission; depends on unreachable counter; gated by RUNBOOK step 6
- **verify-later:** agent_definitions thunder-training-monitor (c3b4c052) / -worker (470c6b3f); 5 actions incl find_active_training_instances; scheduled_tasks

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder unreachable-probe counter
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 106 SQL header "counts CONSECUTIVE unreachable probes and only treats the instance as 'lost' … once the count crosses a threshold"
- **what:** Migration 106 adds `consecutive_unreachable_probes` + `last_probe_at` to `thunder_instances` so the monitor can distinguish a transient SSH blip from a truly-lost box. Each scheduler tick is a fresh sub-agent that can't hold count in memory, so the streak lives on the row: the `record_probe_streak` action bumps on unreachable (route to lost/decommission at threshold, default 3) and resets to 0 on any reachable probe.
- **sources:** phase5/106_thunder_unreachable_counter.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150 (Counter step)
- **relations:** part of thunder-training-monitor; keeps the classifier action pure
- **verify-later:** thunder_instances.consecutive_unreachable_probes; record_probe_streak_action.go

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P4 off-box collection (intent_events + CollectIntentEventsAction)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "Migration applied (CREATE TABLE + 3 indexes + INSERT 0 1 task)"; action "VERIFIED against live source + registered"; but agent deploy fields still to confirm and enable order pending.
- **what:** The cluster pulls intent over key-gated HTTPS with NO adapter and NO SSH. `intent_events` table (engine_event_id UNIQUE = structural idempotency, CHECK on kind/value len, host→site_id resolve, checkpoint = max(event_created_at) with no extra storage). `collect_intent_events` is a SINGLE Go action that self-queries all VM backend sites and loops (parameterised upserts), registered in GlobalActionRegistry (Category "data", IsLocal). Ingest contract: parameterised SQL only, per-line shape checks, burst dedupe, NFC normalisation + lowercasing here.
- **sources:** traffic_probe_plan(11).md#p4, traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-c
- **relations:** driven by intent-collection-orchestrator/intent-collector agents; extended with collectSiteStats + access-digest pull
- **verify-later:** intent_events_migration.sql; intent_collector_actions.go; registry.go DATA region

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded checkpoint-JSON / events-per-1k ranking (early P4)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** plan(1)/(4)/(5) P4 "checkpoint JSON, compute events-per-1k, rank domains"; plan(11) now "idempotent via unique engine_event_id; no extra checkpoint storage — since=max(event_created_at)".
- **what:** Early P4 phrasing planned an explicit checkpoint-JSON file to track collection progress and a direct events-per-1k rank. Dropped in favour of structural idempotency (unique engine_event_id) with the checkpoint derived as since=max(event_created_at) — no extra checkpoint storage. Ranking became a set of read-only SQL queries.
- **sources:** traffic_probe_plan(4).md#phases, traffic_probe_plan(1).md#phases, traffic_probe_plan(11).md#p4
- **relations:** replaced by intent_events unique-id design + intent_ranking_queries
- **verify-later:** intent_events.engine_event_id UNIQUE

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-collection-orchestrator + intent-collector agents
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** intent_collector_agents SQL headers "intent-collection-orchestrator + intent-collector (P4) … mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim"; running_notes 2026-06-13(g) INSERT bug fixed.
- **what:** A thin wrapper-orchestrator (spawn_collector → call_collector → complete, no substantive in-chassis work) that spawns the `intent-collector` task worker (one step: collect_intent_events, processing_mode "task"). Infra fields (image docker.io/aqls/agent-chassis v1.0.1063, resources, health_config, business-intel topics, delegation) copied verbatim from the med-export pair. Reached by the scheduler via target_topic=system.agent.generic.requests by agent_type. Idempotency uses `ON CONFLICT (type, version)`.
- **sources:** deploy_setup/working_dir/intent_collector_agents(3).sql#header, deploy_setup/working_dir/intent_collector_agents(1).sql, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** identical to live intent_collector_agents(2).sql; wrapper-orchestrator requirement; replaces a single in-pod collector
- **verify-later:** agent_definitions rows intent-collection-orchestrator / intent-collector

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Retention prune timer
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Added to setup.sh: RETENTION_DAYS param (default 90) + site-engine-prune.service/.timer (daily find-delete of old events-*.jsonl)".
- **what:** Because daily JSONL IS the rotation, logrotate on engine files would race the open handle; instead setup.sh installs a `site-engine-prune` systemd service+timer that daily find-deletes `events-*.jsonl` older than RETENTION_DAYS (default 90). nginx logs keep their existing size-based logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_runbook(12).md#3.5
- **relations:** part of the privacy posture
- **verify-later:** setup.sh site-engine-prune.timer; RETENTION_DAYS

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### claimed-item-timeout evidence-gated completion + reset (Lever A/C) — avoided building a duplicate watchdog
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 12: "'Auto-completed...' is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper... `migration_claimed_item_timeout_evidence_check.sql` ... is essentially Option A + Lever C, already authored... the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it." Part 12 addenda confirm the v2 migration (page_components-based evidence, not the untrustworthy `build_status='deployed'` flag) applied and verified live, plus the companion `pageHasComponents` deploy-guard (Option B) delivered.
- **what:** A `claimed-item-timeout` scheduled task's `pre_query` already implements both (a) evidence-gated auto-completion of stuck claims (only complete if the specific artefact shows positive evidence) and (b) a stale-claim reset-to-`triaged`/`failed` after a timeout with attempt counting. Mid-investigation, an agent nearly built a brand-new "reset stale claims" watchdog before discovering this — a documented reuse-over-build catch. The evidence signal itself evolved further: from trusting `pages.deployed_at`/`build_status='deployed'` (provably lying, per the homepage case) to checking `page_components.updated_at > claimed_at` directly.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 12–12 addenda
- **relations:** A4 homepage root cause (above); sectionless-page durability stack (below); dispatch throughput (Family J)
- **verify-later:** current `claimed-item-timeout` pre_query SQL in `scheduled_tasks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dispatch throughput bottleneck (Family J) — one-site-per-tick, NOT-EXISTS-blocked
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J: "the dispatcher is **one-site-per-tick** (selection `LIMIT 1`... processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'`... line 276)... Standard manual unstick for now... **To investigate in the separate thread.**"
- **what:** Multi-tool/multi-game adoption sites drain over hours, appearing stalled, because the build-dispatch mechanism processes one site per scheduler tick and blocks an entire site's queue while any single item on it is claimed (no bounded concurrency, no per-item exclusion). A dead handler leaving a stale claim freezes the whole site until a reaper resets it. Explicitly spun out as a separate, not-yet-investigated thread rather than fixed within this arc; running_notes_17(16) later notes it's still an open TODO ("SPEED UP the rebuild pipeline... Not yet investigated").
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J; running_notes_14(20) Part 9; running_notes_17(16) "Missing-game... + speed TODO"
- **relations:** claimed-item-timeout evidence-check reliability mini-project; A1 tool/game deploy gap
- **verify-later:** `build-pipeline-trigger` dispatcher current selection logic (`LIMIT 1`, NOT EXISTS clause, line ~276 at time of writing).

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
