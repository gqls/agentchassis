
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
