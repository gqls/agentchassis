# EXTRACTION U20 — legacy docs A (docs001_flow_general, docs001a_password_changing, docs002_hitl_parallel, docs003_firecrawl, docs004_website_capture_project)
Extracted 2026-07-13. Files in scope: 131. Concepts found: 79.

Era note: this is the platform's origin material (docs dated ~2025-09 → 2025-11). Status signals
are documentary — what these docs claim as of their writing. Where a live successor is evident
from the taxonomy seed (work-item pipeline, adoption pipeline, design-composition, render
pipeline, imagery pipeline), it is named in `relations` and the signal set accordingly.

## Coverage

| file | treatment |
|---|---|
| docs001_flow_general/100_content_page_build_handler_flow.md | full |
| docs001_flow_general/README.001.actions.md | full |
| docs001_flow_general/README.002.agent_orchestration1.philosophy.md | full |
| docs001_flow_general/README.003.calculate.md | full |
| docs001_flow_general/README.004.call_agent1.refactor_into_functions.md | full |
| docs001_flow_general/README.005.discovery.md | full |
| docs001_flow_general/README.006.executeLocalAction1.refactor_into_functions.md | full |
| docs001_flow_general/README.010.flow.md | full |
| docs001_flow_general/README.011.flow2.md | full |
| docs001_flow_general/README.012.flow3.md | full |
| docs001_flow_general/README.014.flow4.1.routingtooriginalsender.md | full |
| docs001_flow_general/README.015.flow4.sequence.md | full |
| docs001_flow_general/README.016.flow5.md | full |
| docs001_flow_general/README.017.flow6.md | full |
| docs001_flow_general/README.018.flow7.roleflow.md | full |
| docs001_flow_general/README.019.flow8.role_based_agent_pools.md | full |
| docs001_flow_general/README.020.flow9.topicflow.md | full |
| docs001_flow_general/README.021.flow10.initialrequestflow.md | full |
| docs001_flow_general/README.022.flow11.initialisationflow.md | full |
| docs001_flow_general/README.023.flow12.await_response.md | full |
| docs001_flow_general/README.024.flow14.input_data.md | full |
| docs001_flow_general/README.040.orchestration_actions.md | full |
| docs001_flow_general/README.041.role_flow.md | full |
| docs001_flow_general/README.042.spawn_actions.md | full |
| docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md | full |
| docs001_flow_general/README.044.spawn_actions3.spawn_rules.md | full |
| docs001_flow_general/README.045.spawn_actions4.refactor_into_functions.md | full |
| docs001_flow_general/README.046.workflow_actions1.refactor_into_functions.md | full |
| docs001_flow_general/README.050.overall_plan1.website_design.md | full |
| docs001_flow_general/README.060.groupagents1.md | full |
| docs001_flow_general/README.061.groupagents2.md | full |
| docs001_flow_general/README.062.groupagents3.databases.md | full |
| docs001_flow_general/README.070.a.centraliseddatanormalisation.md | full |
| docs001_flow_general/README.070.b.execution_context_flow.md | full |
| docs001_flow_general/README.080.a.packaging_data.md | full |
| docs001_flow_general/README.081.b.requestIDflow.md | full |
| docs001_flow_general/README.090.a.human_in_the_loop.md | full |
| docs001_flow_general/README.095.a.image_handling.git.057_image.md | full |
| docs001_flow_general/README.095b.gpu_image_static_dynamic_agent_strategy.md | full |
| docs001_flow_general/README.095c.image_handling_topics.md | full |
| docs001_flow_general/README.095d.mycurrentinputmessagebeforechanging.md | full |
| docs001_flow_general/README.095e.newmessage.md | family-delta (duplicates 095c/095d content + image test script) |
| docs001_flow_general/README.096a.robothands_image_test.md | header-scan (operational test quick-start; outline scanned) |
| docs001_flow_general/README.096b.robothandswebsite.md | full |
| docs001_flow_general/README.096d.robotics_startmessage.md | full |
| docs001_flow_general/README.097a.imagecreationandstorageflow.md | full |
| docs001_flow_general/README.097z.agent_definitions.md | family-delta (duplicate of 096b SQL) |
| docs001_flow_general/README.098.oldherocontentdefinition.d | full |
| docs001_flow_general/README.099a.image_storage_and_display_urls.md | full |
| docs001_flow_general/README.4.2.lifespanofresponsestopic.md | full |
| docs001_flow_general/call_agentold2.go.doc | header-scan (old Go code) |
| docs001_flow_general/spawn_actions_old | header-scan (old Go code) |
| docs001_flow_general/spawn_actionsold2.doc | header-scan (old Go code) |
| docs001a_password_changing/001_changing_passwords.md | full |
| docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md | full (data_helpers code body skimmed) |
| docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md | full |
| docs002_hitl_parallel/README.0100c.workflow_diagram.md | full |
| docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md | full |
| docs002_hitl_parallel/README.0100e.initial_message.md | full |
| docs002_hitl_parallel/README.0101.human_in_the_loop_flow.md | full |
| docs002_hitl_parallel/README.0102.hitl_agent_definitions | full |
| docs002_hitl_parallel/README.0102b.hitl_agent_defnitions_2 | family-delta (diff vs 0102: uses process_approval_decision) |
| docs002_hitl_parallel/README.0102c.hitl_agent_definitions_3 | family-delta (adds validate_input + ui_config editable_fields) |
| docs002_hitl_parallel/README.0103.hitl_start_message.md | full |
| docs002_hitl_parallel/README.0103b.hitl_alternative_start_message.md | header-scan (duplicate test guide) |
| docs002_hitl_parallel/README.0104.hitl_create_db.md | full |
| docs002_hitl_parallel/README.0105.hitl_message_format.md | full |
| docs002_hitl_parallel/README.0106.hitl_future.md | full |
| docs002_hitl_parallel/README.0106.hitl_multistep_approval.md | full |
| docs002_hitl_parallel/README.0107.hitl_expected_flow.md | full |
| docs002_hitl_parallel/README.0108.hitl_simple_workflow.md | full |
| docs002_hitl_parallel/README.0109.hitl_working_message.md | full |
| docs002_hitl_parallel/README.0110.parallel_execution_proposal.md | full |
| docs002_hitl_parallel/README.0111.hitl_timeouts.md | full |
| docs002_hitl_parallel/README.0112.hitl_timeout_testing.md | header-scan (example workflows, quick reference) |
| docs003_firecrawl/README.0120.11_agent_website_framework.md | full |
| docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md | full |
| docs003_firecrawl/README.0122.gemini_first_draft_implementation_11_agent_framework.md | header-scan (agent SQL examples) |
| docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md | full |
| docs003_firecrawl/README.0124.11_agent_summary.md | full |
| docs003_firecrawl/README.0125.claude_11_agent_summary.md | full |
| docs003_firecrawl/README.0126.firecrawl_agent_definition.md | family-latest (later SQL variants scanned) |
| docs003_firecrawl/README.0127.conditional_branching.md | full |
| docs003_firecrawl/README.0128.go_text_template.md | full |
| docs003_firecrawl/README.0129.testing_webscrape_message.md | full |
| docs003_firecrawl/README.0140.removing_constraint.md | full |
| docs004_website_capture_project/006semantic_themes/README.020.brand_theme_preparation.md | header-scan (SQL/CSS bodies outlined) |
| docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md | header-scan (agent SQL, first 180 lines) |
| docs004_website_capture_project/006semantic_themes/README.022.description.md | full |
| docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md | header-scan (agent SQL, first 220 lines) |
| docs004_website_capture_project/006semantic_themes/README.023a.description_for_conditional_routing_etc | full |
| docs004_website_capture_project/006semantic_themes/README.024.conditional_step_routing.md | header-scan (Go code) |
| docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion | full |
| docs004_website_capture_project/007different_types_of_site/027_css_js_schema.sql | header-scan (SQL DDL, dup of 020 content) |
| docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md | header-scan |
| docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql | header-scan (agent-definition SQL, prompts read) |
| docs004_website_capture_project/007different_types_of_site/030_about_page_and_privacy.sql | header-scan (group SQL) |
| docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md | header-scan |
| docs004_website_capture_project/998categorisation/031_add_categorisation_to_tables.sql | header-scan (DDL + category comments read) |
| docs004_website_capture_project/firecrawl/001claude_initial.md | header-scan (Python adapter code) |
| docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md | header-scan |
| docs004_website_capture_project/initial_messages/initial_messages.txt | header-scan (operational scripts) |
| docs004_website_capture_project/playwright/001_claude_thought_process.md | full |
| docs004_website_capture_project/playwright/capture_actions.go.doc | header-scan (Go code) |
| docs004_website_capture_project/playwright/data_helpers_reference.md | header-scan |
| docs004_website_capture_project/playwright/implementation_roadmap.md | header-scan |
| docs004_website_capture_project/playwright/playwright_actions.go.txt | header-scan (Go code) |
| docs004_website_capture_project/playwright/playwright_adapter.py | header-scan (Python code) |
| docs004_website_capture_project/playwright/playwright_adapter_requirements.md | full |
| docs004_website_capture_project/playwright/reissue_all_code.md | family-delta (reissue of same SQL/Go) |
| docs004_website_capture_project/playwright/test_playwright_adapter.py | header-scan (Python code) |
| docs004_website_capture_project/playwright/website_builder_integration_guide.md | header-scan |
| docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql | full |
| docs004_website_capture_project/playwright/website_capture_agent.sql | full |
| docs004_website_capture_project/webbuild_pipeline/001pipeline | full |
| docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md | full |
| docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md | full |
| docs004_website_capture_project/website_analysis/README.003.summary_for_development.md | full |
| docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md | full |
| docs004_website_capture_project/website_analysis/README.005.frontend_frameworks.md | full |
| docs004_website_capture_project/website_analysis/README.006.visual_to_code.md | full |
| docs004_website_capture_project/website_analysis/README.007.behavioural_models.md | full |
| docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md | full |
| docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md | full |
| docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md | full |
| docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md | header-scan (workflow SQL) |
| docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md | full |
| docs004_website_capture_project/website_analysis/README.015.various_ai_thoughts.md | full |
| docs004_website_capture_project/website_analysis/README.016.agent_definitions_002.md | header-scan (data-mapping section read) |
| docs004_website_capture_project/website_analysis/README.017.base_components.md | header-scan (component HTML/CSS bodies) |
| docs004_website_capture_project/website_analysis/README.018.brand_designer_agent.md | full |

## Concepts

### Universal orchestration principle ("every agent is an orchestrator")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "Current Implementation Status … ✅ Universal orchestration capability"; repeated as a working rule in docs002 ("Every agent is an orchestrator").
- **what:** No architectural distinction between orchestrator and worker agents. Every agent runs the same chassis, can spawn children, orchestrate workflows, and execute tasks simultaneously; complexity is fractal (agents compose into arbitrarily deep trees). This is the founding philosophy of the agent-chassis platform.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md
- **relations:** agent chassis; SagaCoordinator workflow engine; agent groups; superseded in practice for site building by the work-item pipeline (development-guide) but still the chassis foundation.
- **verify-later:** agent-chassis main; platform/orchestration/coordinator.go; whether dynamic spawn trees are still exercised vs. static handler deployments.

### Stateless agents with database-backed orchestration state
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002: "✅ Stateless agent design with database-backed state"; orchestration_states schema with version optimistic locking shown.
- **what:** Agents are ephemeral execution containers (K8s pods/Jobs); all orchestration state lives in the `orchestration_states` table (orchestration_id, current_step, awaited_requests, status, processing_history, version). Pod crashes lose no work; the DB is the authoritative source of truth.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#orchestration-state; docs001_flow_general/README.012.flow3.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** SagaCoordinator; AwaitedRequests map; environment reset runbook (truncates this table).
- **verify-later:** clients_db orchestration_states table; UpdateStateWithVersion; whether table is still active or superseded by work_items.

### ExecutionContext unified message envelope and ID semantics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "✅ ExecutionContext as unified message structure"; detailed ID-trace docs (flow10, requestIDflow) resolving the semantics.
- **what:** Every Kafka message carries an ExecutionContext: correlation_id ties the whole end-to-end operation; orchestration_id identifies one workflow instance; request_id identifies a single request/response cycle (new per communication); parent_orchestration_id records who called you; plus tree depth/path, fuel budget, timeout, responses_topic. Sender constructs the child's context; receiver trusts headers.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md
- **relations:** perspective transformation; reply-to metadata; MessageType semantics.
- **verify-later:** platform/orchestration/types ExecutionContext; messaging/context.go NewMessageContext.

### Topic-per-agent Kafka communication
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** README.020.flow9: "LEGACY TOPICS (Pre-created)… Why Legacy Topics Persist"; the doc itself designs the job-topic replacement.
- **what:** Original model: static topics `system.agent.{type}.requests/responses` per agent type. Kept only as bootstrap/well-known entry points (initial client contact) after message-stealing and routing conflicts pushed the design to dynamic job topics.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#topic-per-agent; docs001_flow_general/README.020.flow9.topicflow.md
- **relations:** superseded by job-specific dynamic topics (hybrid model); ultimately by the work-item pipeline for site building.
- **verify-later:** which system.agent.* topics still exist on the cluster.

### Job-specific dynamic Kafka topics (hybrid bootstrap model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow9 "after discussion" decision section + spawn_actions2 walkthrough showing `job.{corrID}-{orchID}-{agentType}-{step}.requests/responses` created at spawn; robot-hands runs used them live.
- **what:** Each spawn creates private per-orchestration topics from a "stable identity" (correlation short + orch short + agent type + spawning step). Root agents listen on standard pre-created topics; spawned agents get their topics via REQUESTS_TOPIC/RESPONSES_TOPIC env vars. Parents talk to children on the child's job topic; children reply to the caller's responses topic carried in headers. Solves the chicken-and-egg bootstrap problem and message collision between parallel jobs.
- **sources:** docs001_flow_general/README.020.flow9.topicflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** stable identity; spawn_agent; environment reset runbook (deletes job.* topics).
- **verify-later:** kafka.CreateStableIdentity; topic creation in SpawnAgentAction; current topic list.

### Two-phase agent lifecycle (spawn + initialize handshake)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow11: "initialize is not treated as a command to start a workflow… handled as a special protocol action"; multiple traced runs.
- **what:** Spawning creates a K8s Job then sends an `initialize` protocol message; the new pod configures itself (role, topics), sends an initialization response, and only then does the parent resume and send `process` work. Initialize bypasses the workflow engine entirely — its only purpose is setup/readiness confirmation. Isolates init failures from execution failures.
- **sources:** docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.010.flow.md
- **relations:** spawn_agent; await_response semantics; a fire-and-forget spawn variant caused ignored init responses (flow12).
- **verify-later:** processor.go initialize handling; SendInitializationResponse.

### SagaCoordinator DB-defined JSON workflow engine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Extensive traced executions (flow docs); workflows stored as JSON `{start_step, steps{action, config, next_step}}` in agent_definitions/agent_group_definitions and executed live.
- **what:** The coordinator loads a JSON workflow from the DB, executes steps via an action registry, stores each step's result in CollectedData under the step name, pauses on `await_response: true` by recording request IDs in an AwaitedRequests map (status AWAITING_RESPONSES), and resumes when matching `in_response_to_request_id` responses arrive (join when the map empties). complete_workflow packages results and replies to whoever is waiting — root vs child completion unified.
- **sources:** docs001_flow_general/README.010.flow.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.006.executeLocalAction1.refactor_into_functions.md; docs001_flow_general/README.046.workflow_actions1.refactor_into_functions.md
- **relations:** action registry (validate_input, transform_data, execute_llm_prompt, spawn_agent, call_agent, aggregate_data, conditional_branch, complete_workflow…); await_approval builds on the same pause mechanism.
- **verify-later:** platform/orchestration/coordinator.go; actions/registry.go; whether coordinator still runs under current handlers.

### spawn_agent — database-definition-driven Kubernetes Job spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.042: "All deployment specs now come from database… No Code Changes for New Agents — Just insert into agent_definitions"; spawn walked through line-by-line in spawn_actions2.
- **what:** SpawnAgentAction reads the child's agent_definitions row (image repo/tag, resources, health config, env vars, default workflow), inserts an agent_instances row in the client schema, creates job topics, launches a K8s Job with topic env vars, sends the initialize message, and returns spawn results (agent_id, role, topics) into CollectedData for later call_agent lookup.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.045.spawn_actions4.refactor_into_functions.md
- **relations:** agent_definitions registry; job topics; role concept.
- **verify-later:** spawn_actions.go; client_{id}.agent_instances table.

### call_agent with role-based targeting
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.041: "role is the essential piece of information that links a specific task to a specific, previously spawned agent" with code walkthrough; used in all robot-hands workflows.
- **what:** CallAgentAction finds a previously spawned agent by searching CollectedData spawn results for a matching `target_role`, extracts its private requests_topic, and sends a `process` request there with await_response. Role acts as the within-orchestration nickname distinguishing multiple agents of the same type (adder vs multiplier calculators).
- **sources:** docs001_flow_general/README.041.role_flow.md; docs001_flow_general/README.018.flow7.roleflow.md; docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** spawn_agent; spawn step naming conventions; role-based agent pools proposal.
- **verify-later:** call_agent.go findAgentByRole/findTargetAgent.

### Role-based agent pools / atomic work-claim queue (proposal)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** flow8 is pure proposal ("Migration Path: Phase 1…Phase 4"), never referenced as built in later docs; the design (work_items table, `claimed_by IS NULL` atomic UPDATE claim, role queues, failover pickup) is recognisably the ancestor of today's work-item pipeline.
- **what:** Instead of spawning agents tied to IDs, agents register roles/capabilities and claim WorkItems atomically from role-specific queues (`system.roles.{role}.pending`); unclaimed work survives agent death; pools scale elastically. "The role becomes the contract, not the agent ID."
- **sources:** docs001_flow_general/README.019.flow8.role_based_agent_pools.md
- **relations:** successor: work-item lifecycle / page-build-handler pipeline (development-guide, docs 001 current); scheduler-and-tasks concurrency groups.
- **verify-later:** work_items table and claim semantics in the current codebase — compare with this 2025 sketch.

### Prompt resolution priority hierarchy
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.004 Part 3 "How the Flow Works Now" with three tested scenarios and log lines ("Using prompt from incoming message (Priority 1)").
- **what:** execute_llm_prompt resolves its prompt in priority order: (1) prompt passed in the incoming message/step config by the caller, (2) the agent's own prompt_template from agent_definitions, (3) workflow-step fallback. Lets parents override specialists while specialists keep good defaults.
- **sources:** docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** execute_llm_prompt action; agent_definitions default_config.
- **verify-later:** ai_actions.go ExecuteLLMPromptAction prompt lookup order.

### Reply-to metadata (__work_request__) and respond-to-caller convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.081.b "Clean Reply-To Architecture… Store reply-to metadata when receiving a work request, use it when completing"; docs002/0100d states the convention as an operating rule.
- **what:** Each agent stores, at work-receipt time, the request_id it must answer and the parent's responses topic together (`__work_request__` in CollectedData) and uses them at complete_workflow. Rule: agents always respond to the *caller's* responses topic, never their own. Works at any hierarchy depth; replaced fragile multi-fallback lookups and fixed empty `in_response_to_request_id` bugs.
- **sources:** docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.014.flow4.1.routingtooriginalsender.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md#response-topic-routing
- **relations:** ExecutionContext; CompleteWorkflowAction; early routing failure modes.
- **verify-later:** BuildCollectedData storing __work_request__; workflow_actions.go completion path.

### CollectedData normalisation and data_helpers safe-access layer
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full data_helpers.go source reproduced as "the new functionality" in docs002/0100 and used by all subsequent agents ("data_helpers.go functions ensure consistency", 0100c).
- **what:** One central layer (data_helpers.go) normalises every inbound message into a canonical CollectedData shape — `input_data` always at top level, system fields (`__execution_context__`, `__my_requests_topic__`, `__raw_message__`…) separated — and provides the only sanctioned accessors (GetInputData, GetStepData, GetMultipleStepData, GetFieldFromPath, TransformDataForAction, BuildRequestMessage/BuildResponseMessage/BuildInitializationRequest). Killed the `input_data.input_data` nesting chaos. Child input_data is always overwritten at top level — each agent's context is exactly what its parent sent (clean-slate encapsulation).
- **sources:** docs001_flow_general/README.070.a.centraliseddatanormalisation.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.024.flow14.input_data.md; docs001_flow_general/README.080.a.packaging_data.md
- **relations:** output_field/input_fields mapping contract; every action.
- **verify-later:** platform/orchestration/datahelpers package.

### Perspective transformation (sender constructs context, receiver trusts headers)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow6: "The critical fix is in ProcessMessage… NewMessageContext(msg, headers, p.agentType). This ensures every agent sees the conversation from their own perspective"; flow10 codifies sender responsibility.
- **what:** On receipt, NewMessageContext transforms the message into the receiving agent's own perspective (its own OrchestrationID becomes primary; the caller's becomes ParentOrchestrationID). The *sender* is responsible for correctly constructing the child's context headers; the receiver only deserialises and trusts them — earlier receiver-side guessing caused validation failures and misrouting.
- **sources:** docs001_flow_general/README.017.flow6.md; docs001_flow_general/README.021.flow10.initialrequestflow.md
- **relations:** ExecutionContext; MessageType semantics.
- **verify-later:** messaging/context.go NewMessageContext signature and transformation logic.

### MessageType semantics (request = actively working, response = reporting back)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.070.b full conceptual write-up with log excerpts ("exec_ctx_message_type":"request" when parent resumes).
- **what:** MessageType describes what the agent is doing *now*, not what just happened: a parent that has received a child's response resumes its own workflow in "request" mode with InResponseTo cleared. Prevents routing/semantic confusion when continuing execution after responses.
- **sources:** docs001_flow_general/README.070.b.execution_context_flow.md
- **relations:** SagaCoordinator continueExecution; perspective transformation.
- **verify-later:** continueExecution fresh-context construction.

### Fuel budget resource limiting
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** FuelBudget field in ExecutionContext (README.002) and `fuel_budget=1000` header in test messages; CreateResponseContext takes "fuel used — calculate properly in production" (0100) — no doc claims enforcement.
- **what:** A per-orchestration computational budget carried in the ExecutionContext, intended to bound resource consumption of agent trees ("if budget.Remaining() < estimated.Cost() return cheaperStrategy()"). Appears plumbed but never implemented as an enforced mechanism.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.061.groupagents2.md (header)
- **relations:** long-term resource optimisation objectives.
- **verify-later:** grep FuelBudget usage — is it ever decremented or checked?

### Agent groups — versioned, discoverable agent teams
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** README.060: "FindBestGroup… queries the database to find the best available version of that group, ordered by performance, usage, and version"; groups used in every website build; evolution/mutation service described but not evidenced live.
- **what:** agent_group_definitions rows are project recipes: a group_type, an agent_configs squad (role → agent_type), and an orchestration_workflow JSON, with integer versions as immutable snapshots (unique group_type+version). Requests name a capability (group_type) and the system picks the best version. An EvolutionService was designed to mutate groups into new versions with parent_id lineage and performance-based selection; the discovery/versioning part shipped, the evolutionary part appears aspirational.
- **sources:** docs001_flow_general/README.060.groupagents1.md; docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.062.groupagents3.databases.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** workflow selection priority; groups-as-project-recipes; spawn_group; site manifest pinning group_version.
- **verify-later:** agent_groups vs agent_group_definitions tables (both exist with different shapes — 062 shows the split); discovery/agent_discovery.go FindBestGroup; evolution.go.

### Workflow selection priority (inline override > group > agent default)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.060/061 define and implement selectWorkflow with the three-tier priority; HITL tests routinely use inline workflow overrides.
- **what:** processor.selectWorkflow resolves which workflow to run: (1) a full inline workflow in the message config (ephemeral/testing), (2) a group workflow found via group_type, (3) the agent's default workflow from agent_definitions. Keeps production versioned while allowing ad-hoc experiments.
- **sources:** docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.060.groupagents1.md; docs002_hitl_parallel/README.0106.hitl_multistep_approval.md (inline workflows in practice)
- **relations:** agent groups; SagaCoordinator.
- **verify-later:** processor.go selectWorkflow.

### agent_definitions registry (DB-driven agent config and versioning)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Dozens of INSERT/UPDATE statements across all four doc sets; constraint migration to UNIQUE(type, version) with previous_version_id (096b); category CHECK constraint managed in 0140.
- **what:** Every agent type is a row: type, display_name, category (constraint-checked: data-driven/code-driven/adapter/…), default_config (containing the workflow, ai_service model+provider, processing_mode, timeouts), capabilities, image_repository/tag (all agents share the agent-chassis image), resources, topics, health_config, env_vars, version + previous_version_id, task_workflow/orchestrator_workflow, delegation_preferences. Creating an agent is a database insert, not a code change.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.096b.robothandswebsite.md; docs003_firecrawl/README.0140.removing_constraint.md; docs001_flow_general/README.098.oldherocontentdefinition.d
- **relations:** spawn_agent; agent categorisation taxonomy (998); the docs024-era agent creation guide is the living successor doc.
- **verify-later:** agent_definitions schema and constraints today; how many of these early agent types still exist/are active.

### Child-orchestration timeout monitor
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** README.040 "Key Features of This Implementation: Configurable Timeout… default 5 minutes… Prevents zombie orchestrations" (claims implemented); HITL timeout doc later shows the config→Step.Timeout mapping was broken.
- **what:** Parents launch a goroutine per awaited child; on timeout it checks whether the parent still awaits that child, sends a timeout error response so HandleResponse processes it normally, and optionally marks the child orchestration failed. Timeout goroutines are in-memory only — recovery on pod restart identified as a gap.
- **sources:** docs001_flow_general/README.040.orchestration_actions.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** HITL approval timeouts; DefaultRequestTimeout 180s.
- **verify-later:** handleRequestTimeout in coordinator.go; recoverPendingTimeouts existence.

### Parallel / fan-out execution in the coordinator
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 lists "Fan-out (parallel) execution implementation" under Outstanding Work; 0110 is an explicit proposal ("Proposed Implementation Strategy… parallel_steps array") with no completion claim.
- **what:** Design for non-blocking workflows: a step's config carries a `parallel_steps` array; executeParallelSteps dispatches all children, records all request IDs in AwaitedRequests, pauses once; processResponse joins when the map empties. Included ExecutionMode enum (sequential/parallel/fan_out). Image workflows sketched parallel_image_generation/batch_image_generation actions on the same idea.
- **sources:** docs002_hitl_parallel/README.0110.parallel_execution_proposal.md; docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** AwaitedRequests join semantics (multi-response already worked); batch-processing category is the modern relative.
- **verify-later:** whether run_parallel/parallel_steps ever landed in coordinator.go.

### Spawn/step naming conventions
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.044: "The naming conventions are now important because we're using them to find spawned agents" — spawn_ prefix required, unique step names with 3-letter suffixes.
- **what:** Workflow authoring rules: spawn steps must start `spawn_<descriptor>` (suffix hints the role), action steps use perform_/execute_/process_ prefixes and reference agents by role, and step names must be unique within a workflow.
- **sources:** docs001_flow_general/README.044.spawn_actions3.spawn_rules.md
- **relations:** call_agent role lookup; workflow authoring guide (development-guide successor docs).
- **verify-later:** whether current workflow JSON still relies on prefix conventions.

### Static vs dynamic agent deployment + GPU cost strategy
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** README.095b is a design ("Integration Steps… Add 3 methods to Agent struct") with cost table ($1,440 static GPU vs $20+$50 CPU-router+dynamic); no implementation claim.
- **what:** Same agent code deployed two ways: static agents (pre-deployed Deployments listening on system.agent.* with pattern-subscribed response topics) and dynamic agents (spawned Jobs on job.* topics); IsStaticAgent() switches behaviour. GPU work handled by an always-on cheap CPU router that spawns short-lived GPU workers (TTL auto-terminate) only when needed — claimed 95% GPU cost reduction.
- **sources:** docs001_flow_general/README.095b.gpu_image_static_dynamic_agent_strategy.md
- **relations:** image-generator adapter (the CPU/GPU split case); model-infrastructure GPU/Ollama docs are the living area.
- **verify-later:** GPU_AGENT_STRATEGY env var; whether any router pattern exists in deployments.

### generate_image action + image-generator adapter pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs002/0100: "Image creation is now working" (deployed then); architecture: agent → `system.adapter.image-generator.requests` → Stability AI → Backblaze/S3 → reply_to_topic. Taxonomy names site_plan_imagery as the current pipeline.
- **what:** Image generation as a first-class workflow action: GenerateImageAction resolves prompts (template-rendered from CollectedData), sends to a shared adapter topic consumed by 3 load-balanced Python adapter replicas (consumer group), which call Stability AI, upload PNG to S3/Backblaze under `images/{client_id}/{date}/{image_id}`, and respond to the requesting agent's topic. Circuit breaker for API failures. A notable bug/fix: GenerateImageAction originally bypassed the image-generator *agent* and posted straight to the adapter — corrected so the agent orchestrates (parent → agent → adapter → agent → parent).
- **sources:** docs001_flow_general/README.095.a.image_handling.git.057_image.md; docs001_flow_general/README.095c.image_handling_topics.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** successor: docs024 imagery / site_plan_imagery pipeline; adapter microservice pattern; GPU strategy.
- **verify-later:** internal/adapters/imagegenerator; whether Stability AI config survives; current imagery pipeline tables.

### Image storage and display URL strategy
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** README.099a recommends and implements dual URIs: "image_uri (s3:// for storage reference), image_url (https:// for web use)"; robot-hands pages embedded presigned URLs.
- **what:** Generated images return both an s3:// URI (storage reference) and a public HTTPS/CDN URL for embedding in HTML; options canvassed were public-bucket/CDN (chosen), presigned URLs (expiry problem for permanent sites), base64 embedding, and an image proxy service. Backblaze B2 public bucket setup documented.
- **sources:** docs001_flow_general/README.099a.image_storage_and_display_urls.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** storage-architecture (S3/B2 credentials); imagery pipeline.
- **verify-later:** ConvertS3URIToPublicURL or equivalent; current image URL scheme on live sites.

### Robot Hands website — first agent-built multi-page site
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning hero writer, image creator, about writer and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav; about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers and image handling.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; platform concepts evidenced: job topics, group workflows.
- **verify-later:** does robot-hands.com exist/what pipeline now owns it.

### aggregate_webpage HTML assembly action
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** Used in robot-hands-complete-website workflows (html_head/html_foot wrapper + section_order + response_fields, add_section_tags, page_name); replaced within docs004 by assemble_full_page/html-assembler and later by the current render pipeline.
- **what:** First-generation page renderer: wraps LLM-generated section content in a hard-coded HTML head (embedded CSS, nav) and footer, stitching named step outputs in a declared order into a complete page file. One action call per page.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** successor: assemble_full_page + html-assembler agent, then the current CSS/render pipeline (styling-render-pipeline docs 036).
- **verify-later:** does aggregate_webpage still exist in the action registry.

### Early long-term platform ambitions (self-organising networks, marketplace, multi-tenant, cross-cluster)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 "Long-Term Objectives (6-12 Months)": self-organising agent teams, agent marketplace, client-isolated multi-tenant namespaces, cross-cluster orchestration with geographic failover.
- **what:** The founding roadmap's horizon list. Multi-tenancy (client schemas) and cross-cluster work later materialised (multicluster docs); the agent marketplace and learned team compositions appear to have vanished.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#long-term-objectives
- **relations:** multicluster (live successor for cross-cluster); database-and-infrastructure client schemas; marketplace = abandoned idea worth registering.
- **verify-later:** none directly; council context only.

### Environment variable validation framework (pre-spawn config validation)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** README.002 Week-3 objective ("EnvironmentBuilder… Validate all environment variables before agent spawn"); never mentioned again in any later doc.
- **what:** Planned framework to declare required/optional env vars per agent and validate before spawn to prevent runtime failures. Silently dropped.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md
- **relations:** spawn_agent env var plumbing.
- **verify-later:** grep EnvironmentBuilder.

### Message-flow logging / observability plan
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** README.002 Week-2 objective: "MessageFlowLogger… Track every message through the system with database persistence"; docs002/0100 problem statement repeats the desire ("closely log and track the creation of agents, the messages…").
- **what:** Persist every send/receive event, agent creation, and topic routing decision to the DB for replay/debugging. Only zap logging plus orchestration_states processing_history is evidenced; a dedicated message-flow store never appears.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** debugging category (docs 016 successors); processed_messages table (exists — see reset runbook).
- **verify-later:** processed_messages table purpose; any message audit table.

### Database password rotation runbook (Postgres → platform secrets → PgBouncer)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Step-by-step live commands with make targets (deploy-065-pgbouncer, pgbouncer-restart/test) and the caution about preserving other secret keys.
- **what:** The password chain has three holders: PostgreSQL users, the `personae-platform-secrets` K8s secret (read by agents), and the `pgbouncer-userlist` secret. Safe rotation order: ALTER USER in PG (existing conns keep working) → update platform secret → rebuild+restart PgBouncer userlist → test → rollout-restart agent pods.
- **sources:** docs001a_password_changing/001_changing_passwords.md
- **relations:** pgbouncer; clients_db/templates_db users; credentials handling (database-and-infrastructure docs 011).
- **verify-later:** make targets still exist; secret key inventory.

### HITL approval-as-specialised-agent architecture (human-reviewer plan)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a is a full implementation plan (approval_tasks/approval_versions schema, human-reviewer agent yaml, pull-based REST API, coordinator pause/resume) with estimated effort; the simpler await_approval mechanism is what the later docs002 tests actually exercise.
- **what:** Design: approval handled by spawning a `human-reviewer` agent whose workflow is create_approval_request → wait (StatusPausedForHuman) → process decision → merge_data approved content back over the generating step's output. approval_tasks + approval_versions tables give a full audit trail; clients poll REST endpoints (list/get/approve/reject/upload-url); versioned image paths `/clients/{id}/jobs/{orch}/{generated|user-uploads|approved}/`. Explicit phase-1 scope cuts: no regeneration loops, no multi-reviewer, no auto-approval.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md
- **relations:** the built alternative: await_approval action; content-type approval capabilities; successor: docs011_api_hitl / humanintheloop (hitl category).
- **verify-later:** do approval_tasks/approval_versions tables exist, or only approval_requests.

### await_approval Kafka pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** 0101: "a solid foundation for HITL already exists… `await_approval` action exists in registry.go"; 0109 records a working end-to-end message pair (resume via responses topic with in_response_to_request_id = approval token).
- **what:** The core HITL primitive: an await_approval workflow step generates an approval token (request_id), publishes a notification (with reply_to and data-to-approve) to `system.notifications.ui`, and returns await_response:true so the SagaCoordinator parks the orchestration. A human/UI resumes it by producing a response whose `in_response_to_request_id` matches the token — initially via `system.commands.workflow.resume`, in working practice via the paused agent's responses topic. Manual kcat testing procedure documented before any UI existed.
- **sources:** docs002_hitl_parallel/README.0101.human_in_the_loop_flow.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0109.hitl_working_message.md; docs002_hitl_parallel/README.0107.hitl_expected_flow.md
- **relations:** SagaCoordinator AwaitedRequests; process_approval_decision; approval_requests table; intake orchestrator HITL gates.
- **verify-later:** actions/hitl_actions.go; system.notifications.ui consumers today (admin dashboard?).

### approval_requests table
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** DDL in 0104 (request_id, orchestration_id, data, status, approved_by, comments); 0111 notes storeApprovalRequest "is a stub" and the table lacks timeout fields.
- **what:** Persistence for pending approvals so clients can poll and audits survive: token-keyed rows with status pending/…, approver identity, comments, timestamps. Created but the write path was initially stubbed.
- **sources:** docs002_hitl_parallel/README.0104.hitl_create_db.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** await_approval; HITL timeouts (adds timeout_seconds/timeout_at columns).
- **verify-later:** table existence + whether storeApprovalRequest/updateApprovalRequest are implemented.

### Content-type-aware approval capabilities (text edit / image replace)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** README.090.a capabilities matrix (text: can_edit_directly; image: can_replace with presigned upload_url methods) and buildResponseData branching; 0102c adds ui_config editable_fields.
- **what:** Approvals adapt to content type: text approvals allow inline editing (edited_content replaces llm_output, original preserved), image approvals allow replacement via pre-signed S3 upload or external URL; ui_config hints (title, editable_fields, actions) let a generic UI render each approval correctly.
- **sources:** docs001_flow_general/README.090.a.human_in_the_loop.md; docs002_hitl_parallel/README.0102c.hitl_agent_definitions_3; docs002_hitl_parallel/README.0105.hitl_message_format.md
- **relations:** human-reviewer plan; imagery storage paths.
- **verify-later:** any UI consuming ui_config; presigned upload endpoint.

### HITL approval timeouts (config mapping, defaults, restart recovery)
- **category:** hitl
- **status-signal:** partial
- **status-evidence:** 0111: "All approval requests currently timeout after 180 seconds… regardless of workflow config, because Step.Timeout is 0" — bug analysis with phased fix plan; 0112 quick-reference states approval default 24h, min 60s, max 7 days.
- **what:** Approval steps carry `timeout_seconds` (up to multi-day) but the value was sent to the UI and never mapped onto Step.Timeout, so the generic 180s DefaultRequestTimeout applied. Fix plan: map config→Step.Timeout at execution, store timeout_at in approval_requests, validate bounds (60s–7d, default 24h), and recover timeout goroutines from AwaitedRequests on pod restart (goroutines are memory-only).
- **sources:** docs002_hitl_parallel/README.0111.hitl_timeouts.md; docs002_hitl_parallel/README.0112.hitl_timeout_testing.md
- **relations:** child-orchestration timeout monitor; approval_requests table.
- **verify-later:** getTimeout(step) mapping today; recoverPendingTimeouts.

### process_approval_decision and rejection routing
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Registered action used in all HITL test workflows; config `stop_on_reject`, and Option-3 workflow routes on `process_approval.approved` via conditional_route to finalize/handle_rejection.
- **what:** Post-approval step that unpacks the human's decision (approved flag, comments, modified_data) into CollectedData and lets workflows branch on it — continue, stop, or run a rejection handler path.
- **sources:** docs002_hitl_parallel/README.0106.hitl_multistep_approval.md; docs002_hitl_parallel/README.0105.hitl_message_format.md; docs002_hitl_parallel/README.0102b.hitl_agent_defnitions_2
- **relations:** await_approval; conditional_route/evaluate_condition.
- **verify-later:** actions registry entries process_approval_decision, conditional_route.

### EBORG — evidence-based organisational planning (venture concept)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Appears only as the demo business for HITL content approval (0103, full pitch text in the trigger script); never seen in later doc eras.
- **what:** A business idea used as the HITL demo client: map every role/responsibility/objective in an organisation and pair each with a framework of AI agents that gather research, assess options, and provide evidence-based reasoning — "human-centered, continuously learning organisation". Also spawned the simple-content-writer-with-approval agent.
- **sources:** docs002_hitl_parallel/README.0103.hitl_start_message.md; docs002_hitl_parallel/README.0102.hitl_agent_definitions
- **relations:** HITL content approval group (content-approval-hitl); thematically echoes the later council-of-experts idea in docs026 stage 3.
- **verify-later:** none (idea registry only).

### 11-agent website analysis framework (four agent groups)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Whole docs003 set is planning ("Here is the detailed, point-by-point analysis…"); docs004 explicitly reframes it ("The old numbers are meaningless now… rename them") into the Learn/Execute playbook model.
- **what:** The original web-capture master plan: Strategy & Content group (Strategist A10, Content Infuser A11), Library & Storage (Librarian A7, S3+Postgres/pgvector), Design Ingestion (Prospector A0, Site Profiler A1, Capture Bot A2/Playwright, Layout & Labeling A3 XY-Cut+LLaVA, Component Generator A4 VLM screenshot-to-code, Style Extractor A5 getComputedStyle — later eliminated in favour of Firecrawl branding data, Behavior Extractor A6 CodeLlama), Generation (Publisher A8 "Dribbble-like" showcase site, Architect A9 template builder querying by CLIP embedding). All implemented as agent_definitions rows + new action adapters, not new binaries.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md; docs003_firecrawl/README.0124.11_agent_summary.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** successor chain: playbook model (docs004) → MVP site builder → current adoption-pipeline (docs 007) and site-spec-and-classifier (docs 021). Publisher A8's public design-library site was abandoned.
- **verify-later:** which of the 11 agent types ever got agent_definitions rows.

### Adapter microservice pattern (Kafka/HTTP adapters + secure external-API proxies)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 0123 codifies the pattern; image-generator, firecrawl/webscrape, playwright, git adapters all follow it; "We will use this exact same pattern for all our Python-based actions."
- **what:** Go agents never embed heavy dependencies: a workflow action produces a Kafka message to `system.adapter.<name>.requests` (or an internal HTTP call); a containerised worker service (Python or Go) in its own Deployment consumes via a shared consumer group, does the work (Playwright, Firecrawl, Stability, git), and replies to the reply_to topic. External GPU/API providers get a dedicated Go proxy adapter that holds the secret key and translates request formats — swap providers by changing one adapter, no workflow changes.
- **sources:** docs003_firecrawl/README.0123.actions_needed_firstdraftpython.md; docs004_website_capture_project/playwright/implementation_roadmap.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md
- **relations:** adapters category anchor (docs 033/035 successor); image adapter; firecrawl adapter; thunder adapter (taxonomy) is a descendant of the "ThunderCompute LLaVA proxy" idea here.
- **verify-later:** internal/adapters/* inventory; which adapter topics exist.

### Firecrawl scraping adapter and actions
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Agent definitions website-capture-firecrawl and webscrape-simple with image tags v1.0.407→v1.0.424 (iteration = real use); v1→v2 migration doc fixing "Unrecognized keys" errors and adding S3 ownership of screenshots/images.
- **what:** Firecrawl API adapter (Kafka consumer on system.adapter.firecrawl.requests) exposing scrape/crawl/extract actions to workflows (firecrawl_scrape, firecrawl_crawl, firecrawl_extract, plus a registered scrape_web action with upload_results to S3). v2 migration: formats array incl. screenshot+links, downloading Google-Cloud-hosted screenshots/images into own S3 (webscrape/client/date/id/ layout) for data ownership since Firecrawl assets expire in 30 days. Chosen over the half-built Playwright adapter to reduce MVP load.
- **sources:** docs003_firecrawl/README.0126.firecrawl_agent_definition.md; docs004_website_capture_project/firecrawl/001claude_initial.md; docs004_website_capture_project/firecrawl/002firecrawl_visual_flow.md; docs003_firecrawl/README.0129.testing_webscrape_message.md
- **relations:** adoption-pipeline crawling (live successor); playwright adapter (the road not taken); storage-architecture.
- **verify-later:** web-scrape-adapter deployment (referenced in initial_messages.txt scale-down list — so it was deployed); FIRECRAWL_API_KEY secret.

### evaluate_condition — template-based conditional branching
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 0127/0128 document the working mechanism ("The orchestrator uses this to pick the next step from the next_step map") including Go text/template functions (and/or/not/eq/gt…) and a live website-analyzer group UPDATE.
- **what:** Workflow steps gain branching: evaluate_condition renders a Go text/template expression against CollectedData and returns true/false; `next_step` becomes a map {"true": …, "false": …}. Enables data-driven workflow paths (e.g. extract_structured? crawl_pages? previous step success?).
- **sources:** docs003_firecrawl/README.0127.conditional_branching.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** conditional_branch/conditional_route actions; route_by_field/conditional_call_agent (later, richer routing).
- **verify-later:** evaluate_condition in registry; coordinator support for map-typed next_step.

### website-analyzer conditional scraping group
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Tested with kcat messages (boxing-tickets.com, both basic and structured+crawl variants) and a live UPDATE of its orchestration_workflow.
- **what:** An agent group that takes target_url + flags (extract_structured, crawl_pages, crawl_limit/depth) and conditionally routes between basic scrape, structured extraction, and multi-page crawl using evaluate_condition — the first "smart" capture entry point.
- **sources:** docs003_firecrawl/README.0129.testing_webscrape_message.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** firecrawl adapter; successor: adoption pipeline crawl/classify flow.
- **verify-later:** agent_group_definitions row group_type='website-analyzer'.

### Semantic component library with vector embeddings
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `SELECT … ORDER BY (clip_embedding <=> [vector])` queries (docs003); design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** Vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library + tool registry matching; embeddings idea resurfaces in contextkit (diagnosis-loop).
- **verify-later:** pgvector extension usage; any table with embedding columns from this era.

### Playbook > Strategic Pattern > Component hierarchy (Librarian as system brain)
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Extensive design (Playbooks/Strategic_Patterns/Pattern_Component_Slots/Components schema, success_score feedback) across website_analysis 001–003; no implementation era follows — the MVP path (chief-strategist + in-house components) shipped instead, and the schema never reappears.
- **what:** "Strategy-to-website engine": the library stores *business solutions*, not just components — Playbooks (objective+vertical strategies with success scores, e.g. affiliate product-review), containing Strategic Patterns (comparison-table, best-of listicle), containing Components. Learn loop classifies scraped winners into this hierarchy; Execute loop queries "best playbook for objective X in vertical Y" and assembles it; A/B results feed success_score back. The Librarian is the sole read/write gatekeeper; "the link is the database schema".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** behavioural models library (the surviving cousin); site-spec-and-classifier archetype system is the spiritual live successor; affiliate content-type placement knowledge (reviews/comparisons/listicles) embedded here.
- **verify-later:** confirm no Playbooks/Strategic_Patterns tables exist.

### Behavioural models library and functional component labelling
- **category:** NEW:conversion-playbooks
- **status-signal:** partial
- **status-evidence:** "PAS" shipped as a real input (`"model": "PAS"` in mvp-site-builder trigger messages; chief-strategist prompt takes {{.model}}); the wider library (AIDA, Fogg B=MAP, Cialdini, Hook) and deep inference labelling remained design.
- **what:** Components are labelled by *behavioural function*, not visual pattern: not "hero" but "attention_capture"/"problem_statement"/"social_proof", drawn from marketing science (AIDA, PAS, Fogg Behaviour Model, Cialdini's persuasion principles, the Hook model). Build plans map a chosen behavioural model to a sequence of functional sections; the architect assembles "a psychological argument, not just a visual page". Self-critiques recorded: inference black-box risk (LLM can't reliably tell "agitation" from "interest"), theory-vs-reality gap, new-generic monoculture trap.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md; docs004_website_capture_project/website_analysis/README.007.behavioural_models.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** MVF; data-function contract; content strategy in current content pipeline (content-quality docs) is the descendant.
- **verify-later:** whether current build plans still carry a behavioural model field.

### Minimal Viable Funnel (pragmatic-first Day-1 build)
- **category:** NEW:conversion-playbooks
- **status-signal:** superseded
- **status-evidence:** Fully built as mvp-site-builder (boxing-tickets.com runs); superseded within docs004 itself by the briefing→specialist-architect pipeline and later by the current work-item site build.
- **what:** Anti-boil-the-ocean strategy: start with one behavioural model (PAS) and three generic in-house components (problem/agitate/solution blocks) so a strategically coherent landing page can be built with zero scraped data — solving the cold-start problem. Scraping demoted to an "iteration engine" suggesting upgrades.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md#minimal-viable-funnel; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** MVP site builder pipeline; intelligent fallback; in-house forge.
- **verify-later:** —

### Intelligent fallback component matching (P1/P2/P3)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block) and the Generic Text Block fallback component INSERTed (017); mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor).
- **verify-later:** assemble_from_library in registry; fallback rows in content_components.

### Strategic fallback stubs for non-replicable components
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Design-only ("Store 'Stubs' with 'Fallbacks'… two-pronged output") in website_analysis 001/003; no stub tables or developer-task topics appear later.
- **what:** When ingestion finds a component it can't replicate (e.g. a mortgage calculator), record a Stub with its *strategic goal* (lead-gen-quote) and a linked simple fallback component (CTA form). The live site ships the working fallback; simultaneously a developer task goes to a HITL queue ("developer.tasks.required") to build the real thing as v2. The site is always complete and strategically sound.
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#strategic-fallback; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** dynamic-applications (the current interactive app generation finally addresses "non-replicable dynamic apps"); HITL queue.
- **verify-later:** none — idea registry.

### Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs (content_quality_and_internal_linking), research-agents; persona idea → persona architecture across the platform.
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline.

### MVP site builder pipeline (strategist → architect → content-creator → deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Full group SQL + per-agent Kafka payloads documented and run (boxing-tickets.com); renamed/extended into landing-page-builder and 6-step pipeline within docs004; today's site building is the work-item pipeline.
- **what:** The first end-to-end production pipeline: chief-strategist (LLM → build_plan JSON of functional sections), site-component-architect (assemble_from_library → empty semantically-tagged HTML template + content_requirements "shopping list"), content-creator (fills slots), deployer-agent (commit_to_git). Group workflow spawns all four then calls them in sequence, threading outputs through output_field/input_fields.
- **sources:** docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** grew into: + brand-designer, + briefing-agent, + html-assembler, + specialist architects; successor: current page-build-handler/work-item pipeline (see 100_content_page_build_handler_flow.md).
- **verify-later:** agent_group_definitions mvp-site-builder / landing-page-builder rows.

### Git deployment: commit_to_git + GitHub Action sync to B2
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** webbuild_pipeline/001pipeline: "Deployer commits sites/boxing-tickets.com/index.html to GitHub → GitHub Action automatically syncs that folder to B2 → Site is live."
- **what:** Deployment path: a git-adapter microservice (Kafka topic system.adapter.git.requests) commits generated site files to a repo (per-domain repos in the original design; a sites/<domain>/ folder in practice); a GitHub Action syncs to Backblaze B2 which serves the live site. GitCommitAction is the workflow-side action.
- **sources:** docs004_website_capture_project/webbuild_pipeline/001pipeline; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** deployment-github (docs 034 live successor: git-adapter deploy surface); storage-architecture B2.
- **verify-later:** git-adapter service; the GitHub Action workflow file; sites/ repo layout.

### Brand designer agent (theme selection)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Agent SQL + mvp-site-builder workflow insertion (spawn/call_brand_designer feeding brand_theme to the architect); superseded by content-creator's theme recommendation + semantic tag matching in 006semantic_themes, and later by the design-composition system.
- **what:** An LLM agent that analyses domain + objective and picks a CSS theme from the named library (boxing, bakery, tech, professional-dark, default) with reasoning — the first brand/design decision point in the pipeline.
- **sources:** docs004_website_capture_project/website_analysis/README.018.brand_designer_agent.md
- **relations:** semantic CSS theme system; successor: site-design-planner / palette resolution (design-composition docs 025-027).
- **verify-later:** brand-designer agent_definitions row.

### Semantic CSS theme and snippet system (theme_tags, css_themes, css_snippets, js_snippets)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Full DDL + seed data in two iterations (020 text[]; 027 jsonb) with helper matching functions; themes are complete `:root` CSS-variable palettes. The design-composition palette/typography system is the taxonomy-named successor.
- **what:** A semantic tagging vocabulary (mood/style/industry/audience/functional/colour tags with related_tags pairing) applied to: css_themes (full CSS-variable palettes: calm-minimal, bold-conversion, warm-friendly, dark-modern, premium-elegant…), css_snippets (hover/animation/effect/pattern/utility fragments), and js_snippets (nav, scroll animations, accordion, clipboard, form interactions with trigger metadata). Content-creator recommends theme + theme_tags; assembler matches snippets by tags. All theming via CSS variables — the ancestor of the platform's CSS-variable contract.
- **sources:** docs004_website_capture_project/006semantic_themes/README.020.brand_theme_preparation.md; docs004_website_capture_project/007different_types_of_site/027_css_js_schema.sql; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md
- **relations:** successors: contracts-and-standards CSS variables; design-composition palette resolution; styling-render-pipeline.
- **verify-later:** css_themes/css_snippets/js_snippets/theme_tags tables today.

### Content/structure separation: JSON content + html-assembler (assemble_full_page)
- **category:** styling-render-pipeline
- **status-signal:** superseded
- **status-evidence:** 021/022: content-creator refactored to "structured JSON, not full HTML"; html-assembler agent with assemble_full_page action (template render → theme query → snippet queries → document assembly); the current render pipeline is the taxonomy successor.
- **what:** Separation of concerns that defines the modern pipeline: architect emits an empty {{placeholder}} template + content_requirements; content-creator emits pure content JSON (meta, theme recommendation, per-component sections); html-assembler merges template+content via Go templates then injects the CSS theme, tag-matched CSS snippets, and JS snippets into a complete document. Deployer receives finished HTML.
- **sources:** docs004_website_capture_project/006semantic_themes/README.022.description.md; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: styling-render-pipeline (docs 036) + component render contracts (docs 003); content_components input_schema.
- **verify-later:** assemble_full_page in registry; html-assembler agent row.

### Briefing agent (structured brief generation)
- **category:** onboarding-config
- **status-signal:** superseded
- **status-evidence:** Agent SQL (021) inserted into the 6-step pipeline; extended with site_type detection (023); later generalised to questionnaire execution (029). Onboarding/config-derivation docs are the live successor area.
- **what:** First pipeline stage: an LLM turns domain+objective into a comprehensive structured brief JSON — industry inference with confidence, audience demographics/psychographics, brand tone/personality/voice examples, value proposition/key messages/USPs, recommended sections, theme recommendation with semantic tags, content guidelines (avoid/emphasise), monetisation model and ad zones.
- **sources:** docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md
- **relations:** site classifier; questionnaire pattern; successor: onboarding-config (docs019 PLAN_onboarding) and site specs.
- **verify-later:** briefing-agent row; whether brief JSON shape survives in site_specs.

### Site classifier and site_type taxonomy
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** site-classifier agent SQL with the landing/content/portfolio/brochure taxonomy and recommended_group mapping, plus a template-flattening prompt fix (evidence it was actually run); taxonomy names the live classifier architecture (docs 021).
- **what:** A lightweight LLM agent classifying a project into site types — landing (conversion single-CTA), content (publishing/ads/SEO), portfolio (showcase), brochure/directory (multi-page business / listings) — with confidence, reasoning, detected signals, and a recommended builder group. The direct ancestor of the platform's archetype/classification system.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** successor: site-spec-and-classifier (classification architecture, archetype); briefing agent; intake orchestrator.
- **verify-later:** site-classifier agent row; current archetype enum vs this 4-type taxonomy.

### Specialist architects per site type
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-site components (article grid, sidebar, ad zones, category nav), portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated (025) and the group-per-project-type model won conceptually.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing.
- **verify-later:** the three architect rows; content-site component rows.

### Dynamic agent routing (route_by_field / conditional_call_agent)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Both Go actions written with registry additions listed; conditional_call_agent chosen because it "wraps CallAgentAction internally — no coordinator changes needed".
- **what:** Data-driven agent selection inside workflows: route_by_field maps a dot-path field value to a next step via a routes table with default; conditional_call_agent reads e.g. brief_data…site_type, maps value→agent type (landing→landing-page-architect …) and calls that agent in one step, returning routing metadata.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023a.description_for_conditional_routing_etc; docs004_website_capture_project/006semantic_themes/README.024.conditional_step_routing.md
- **relations:** evaluate_condition (simpler predecessor); spawn_group dynamic group_type (group-level equivalent).
- **verify-later:** registry entries conditional_call_agent, route_by_field.

### Groups as project recipes + immutable versioning + agent pinning
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 025 "decision" section: "A group isn't 'agents that work together' — it's a project recipe"; versioning model (immutable version rows, sites pinned to group_version, per-agent image_tag pinning) is design; UNIQUE(group_type, version) constraint added in 0100b — partially realised.
- **what:** Each buildable *kind* of output (landing page, content site, 11ty blog, ecommerce) is a self-contained group: its own agent squad, workflow, questionnaire, and outputs. Divergence in output structure/build/deployment means a new group, not conditional routing. Group versions are immutable snapshots; a site records the group_version that built it and rebuilds with it unless upgraded; groups may pin specific agent versions where stability matters. Duplication across similar groups is accepted for clarity.
- **sources:** docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion; docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md (constraint)
- **relations:** agent groups; site manifest; tool-lifecycle versioning is the analogous live discipline.
- **verify-later:** group version rows per group_type; any site→group_version reference.

### Intake orchestrator with two HITL gates and per-group briefing questionnaires
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** 029.intake_and_groups.sql implements schema (briefing_questionnaire column), site-classifier, intake-orchestrator group, landing/content builder groups; Go actions written (request_human_input with skip conditions, fetch_group_questionnaire); registry additions still listed as "needed".
- **what:** A two-stage front door: classify project (site_type + recommended group) → HITL-1 confirm type → fetch the *target group's* briefing questionnaire (stored in agent_group_definitions, keeping the briefing agent generic) → execute questionnaire (LLM-inferred or human-answered) → HITL-2 review brief → spawn_group dynamically dispatches the chosen builder. HITL points have skip conditions (hitl_mode=auto) for automated runs.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#the-intake-orchestrator; docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md
- **relations:** await_approval mechanism (reused); successor: onboarding-config PLAN_onboarding / config derivation.
- **verify-later:** intake-orchestrator group row; request_human_input/fetch_group_questionnaire in registry; briefing_questionnaire column.

### spawn_group action with DB group lookup and dynamic group_type
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 028 discovers an existing SpawnGroupAction (config-provided agents) and revises the new version (spawn_group_from_db.go) to align — DB lookup of agent_group_definitions, dynamic group_type_field from collected_data, questionnaire fetch.
- **what:** Spawning an entire agent group as a unit: original action spawned each configured agent and returned subtree info; enhanced version resolves the group definition (agents + workflow + questionnaire) from the database, with the group_type optionally taken dynamically from prior step output — enabling the intake orchestrator's dispatch.
- **sources:** docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** intake orchestrator; agent groups.
- **verify-later:** spawn_group vs spawn_group_from_db in codebase.

### Agent/group categorisation taxonomy (category, status, domain_tags)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** Migration SQL 031_add_categorisation with CHECK constraints (category: builder/analyzer/collector/transformer/evaluator/researcher/workflow/monitor; status: active/experimental/deprecated/demo/template) and GIN-indexed domain_tags; no doc confirms it was applied.
- **what:** Organisational metadata over agent_definitions and agent_group_definitions: what the agent *does* (domain-agnostic category), its lifecycle status, and flexible domain tags — an early attempt at the registry hygiene the concept register itself now pursues.
- **sources:** docs004_website_capture_project/998categorisation/031_add_categorisation_to_tables.sql
- **relations:** agent_definitions registry; documentation-system indexing.
- **verify-later:** do the category/status/domain_tags columns exist?

### Playwright capture adapter + website-capture agent
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Complete deliverables (adapter py, capture_actions.go, agent SQL: desktop/mobile viewports, hover/focus states, scroll intervals with parallax/sticky detection, asset extraction, S3 upload) — but docs004/website_analysis 002 records "Agent 5 eliminated… use Firecrawl branding data instead" and firecrawl/001 adapts the MVP away from Playwright.
- **what:** Deep browser-based capture: Playwright adapter on system.adapter.playwright.requests capturing full-page desktop + mobile screenshots, DOM, computed styles, interaction states (hover/focus for up to 50 selectors), scroll-position screenshots (0/25/50/75/100%) with parallax/sticky detection, asset extraction, and organised S3 upload with manifest. Deferred in favour of the managed Firecrawl service for MVP; the deeper capture ideas (interaction/scroll states) never resurfaced.
- **sources:** docs004_website_capture_project/playwright/website_capture_agent.sql; docs004_website_capture_project/playwright/playwright_adapter.py; docs004_website_capture_project/playwright/implementation_roadmap.md; docs004_website_capture_project/firecrawl/001claude_initial.md
- **relations:** firecrawl adapter (chosen replacement); adoption-pipeline crawling successor; behaviour capture (rrweb) idea from docs003 also abandoned.
- **verify-later:** playwright-adapter deployment existence.

### Website-builder orchestrator (capture → vision → code → synthesis → content → library)
- **category:** adoption-pipeline
- **status-signal:** abandoned
- **status-evidence:** Orchestrator SQL references agent types (website-vision, website-code-analyzer, website-synthesis, content-strategist) and actions (analyze_input_type, parallel_section_generation, store_component) that are never defined or mentioned again; the MVP builder took a different shape.
- **what:** A master workflow to rebuild a site from a captured one: capture data → visual analysis (layout/palette from screenshots) → code cleaning/analysis → synthesis correlating visual+code into a template → content planning → parallel section generation → aggregate → store components with embeddings in the library. The maximal "clone-and-improve" vision.
- **sources:** docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql; docs004_website_capture_project/playwright/website_builder_integration_guide.md; docs003_firecrawl/README.0125.claude_11_agent_summary.md
- **relations:** successor in spirit: adoption-pipeline content recreation (docs 007); vision analysis resurfaces in the current image-analysis tooling.
- **verify-later:** confirm none of the four sub-agent types exist in agent_definitions.

### Site manifest + external-edit desynchronisation detection
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** Design tables in 004 (git_hook_adapter, Manifest Sync Agent, status 'desynchronized', HITL review) and manifest.json "winning genes" tracking in 008/014; no implementation evidence.
- **what:** Every generated site carries a manifest.json recording what built it (group_type, group_version, brief, build plan, component genes). A git webhook adapter watches all site repos; a human commit flags the manifest desynchronised, halting agent edit workflows and queueing HITL review — protecting human work from being overwritten and agents from stale state.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** content-governance locks (the live mechanism protecting human edits); deployment-github git-adapter.
- **verify-later:** any manifest.json in site repos; git webhook receivers.

### Adopting existing external sites ("Adopt" workflow)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Designed in 004 ("Adopt workflow… status: 'adopted_partial'… match_confidence"); the taxonomy's adoption-pipeline (docs 007: site crawling, classification, content recreation) is the named live successor.
- **what:** Run the Learn loop against an existing site the platform didn't build: scrape, deconstruct layout, match found blocks to the in-house component library with confidence scores, generate a manifest marking it adopted_partial — making external sites partially manageable by agent edit workflows.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md
- **relations:** successor: adoption-pipeline (docs 007).
- **verify-later:** compare with current adoption pipeline design.

### WordPress handoff (XML export, plugin shortcodes, SQL brand injection)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Detailed plan in 004 (WordPress Formatter agent, wordpress-export.xml, [wpforms] shortcodes for stubs, wp_options SQL injection for branding); frontend-framework survey in 005; never mentioned again anywhere.
- **what:** Client-handoff strategy: transpile a generated site into a single WordPress import file so a client's developer gets a standard maintainable WP site in minutes; complex components become plugin shortcodes; brand colours/fonts injected into theme settings via one SQL file. Part of a broader survey of exit routes (traditional CMS vs SaaS builders vs headless/Jamstack).
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.005.frontend_frameworks.md
- **relations:** business-strategy (client/exit strategy); deployment-github (the retained path).
- **verify-later:** none — abandoned idea registry.

### Pragmatic Evolution model (explore/exploit portfolio cohorts)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** Full strategy synthesis in 008/014 (cohorts: top-10% untouched, middle-40% careful P1-P2 A/B tests, bottom-50% high-velocity churn; site-specific optimisation "no monoculture"); no subsequent doc era operates a portfolio this way — the platform pivoted to per-site quality loops.
- **what:** An evolutionary algorithm over a large site portfolio: select worst performers, radically mutate them with mixed component "genes", evaluate fitness after 3 months. Critique recorded and resolved into an explore/exploit design: attribution black hole and SEO destabilisation confine chaos to a "loser" cohort where attribution is deliberately ignored; winners graduate. Winning changes are applied only to individual sites where they actually won, and content evolves on a separate continuous track from layout to protect SEO.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** improvement-loop (the live per-site discovery→audit→fix cycle is the surviving descendant in spirit); traffic-analytics (fitness signal dependency).
- **verify-later:** none — strategy registry; check if any cohort/experiment tables exist.

### Hypothesis priority list (learn loop as idea generator, not fact finder)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** 008/014: "All scraped data is treated as messy, high-correlation ideas, not truth… Librarian generates a Hypothesis Priority List (P1–P5)"; scorecard interrogation of sites against all behavioural models; no implementation follows.
- **what:** Epistemics for the scraping programme: accept that ingestion finds correlation ("cargo cults"), rank target sites by external success metrics (Ahrefs/Semrush APIs via an seo_api_adapter), interrogate each against every behavioural model to produce confidence scorecards, and emit a prioritised backlog of testable hypotheses for the Evolve loop to convert into causation.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** Prospector/seo_api_adapter (never built); llm-quality-testing shares the evaluation mindset.
- **verify-later:** none.

### Multi-page site support (wrap_multipage, multipage-site-builder)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 030 SQL creates multipage-site-builder (index/about/contact + privacy); 031 shows the wrap_multipage step after html_assembler with CollectedData trace; today's pages/site_plans domain is the successor.
- **what:** Extending the single-page pipeline to small multi-page sites: after assembly, a wrap_multipage action derives index/about/contact (and privacy) pages, and the deployer commits all files. The first step from "landing page generator" toward the current multi-page site model.
- **sources:** docs004_website_capture_project/007different_types_of_site/030_about_page_and_privacy.sql; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: site_plans / pages domain (site-plan-and-reconciler docs 029/030); robot-hands 3-page build (earlier sibling).
- **verify-later:** wrap_multipage in registry.

### In-House Forge — content_components with data-function semantic contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** content_components seeded in 017 (Generic Text Block, Document Head with base CSS…); the *current* page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" (100_content_page_build_handler_flow.md) — the table survived into the live pipeline.
- **what:** The platform's own component library: rows with name, function (semantic purpose), html_template with {{placeholders}}, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (component contracts/slot specs — live successor); tool-library component library; intelligent fallback.
- **verify-later:** content_components schema now vs then; data-function attributes in current templates.

### Page-content-writer + admin content brief regeneration flow
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 100_content_page_build_handler_flow.md documents the live flow (page-build-handler → page-content-writer → load_site_specs → prepare_link_context → load_page_components → process_sections_loop) then states "What's missing: The prompt has no awareness of page_components.content_brief" and specifies the new content_rewrite flow.
- **what:** The bridge document into the modern era: the current work-item pipeline's content generation path, and the gap it fixes — admin edits a brief in the dashboard (page_components.content_brief), clicks Regenerate creating a content_rewrite work item, and the writer's generate_content prompt gains an "## Admin Content Brief" block for briefed sections while unbriefed sections behave as before.
- **sources:** docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** content-governance (briefs, regeneration — anchor doc 013); development-guide work-item lifecycle; content_components definitions.
- **verify-later:** page_components.content_brief column; content_rewrite work item type; load_page_section_components step.

### Aggregation patterns (aggregate_data, aggregator agent, input_from_collected_data)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** aggregate_data's failures traced (count:0 on verbose child responses, flow2); redesign to a spawned aggregator agent fed via input_from_collected_data path mapping (flow11); aggregate_webpage became the shipped variant for pages.
- **what:** Combining multi-step results: the local aggregate_data action broke against verbose child state objects; the redesign either normalises responses (data helpers) or delegates aggregation to a spawned aggregator agent whose call config maps CollectedData paths into its input. Response data keyed as response_{requestID} in CollectedData.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.010.flow.md
- **relations:** data_helpers NormalizeResponseData (the actual fix); aggregate_webpage.
- **verify-later:** aggregate_data current implementation.

### output_field / input_fields group-memory data mapping contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016 "Packet Flow" walkthrough resolving the exact semantics ("Take the entire result from the Strategist and store it under build_plan_data… path is simply build_plan_data.build_plan_json") producing the "Golden Copy" workflows.
- **what:** The inter-agent data plumbing convention: a call_agent step's `output_field` names the key under which the child's entire result lands in group memory; the next step's `input_fields` selects which keys are passed on; consumers address values by `<output_field>.<producer's own output key>` paths. Most orchestration bugs of the era were mis-mappings of this contract.
- **sources:** docs004_website_capture_project/website_analysis/README.016.agent_definitions_002.md; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md
- **relations:** CollectedData normalisation; template rendering paths; note the execute_llm_prompt flattening quirk (input_fields:["input_data"] flattens, so templates use {{.domain}} not {{.input_data.domain}} — 029 fix).
- **verify-later:** call_agent output_field handling in coordinator.

### Orchestration environment reset runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** The identical script repeated across ≥5 docs: scale agent-chassis to 0, TRUNCATE processed_messages/orchestration_states/pending_requests, delete spawned jobs, delete all job.* topics, delete bootstrap topics, reset all consumer-group offsets to earliest, scale back up.
- **what:** The standard clean-slate procedure for the early platform's test cycles — also documents the persistence surface of the era: processed_messages (dedupe), orchestration_states, pending_requests tables; job.* + system.agent.* topics; spawned-by=orchestrator job labels.
- **sources:** docs001_flow_general/README.095d.mycurrentinputmessagebeforechanging.md; docs001_flow_general/README.096d.robotics_startmessage.md; docs004_website_capture_project/initial_messages/initial_messages.txt
- **relations:** debugging (docs 016 successors); stateless-agents concept (what gets truncated).
- **verify-later:** pending_requests/processed_messages tables still present?

### Early message-routing failure modes (case-study catalogue)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Each bug has a trace + fix doc: nested-vs-flat input_data mismatch (flow2), verbose child responses breaking aggregation (flow2), silent root completion (flow2/3), duplicate second response to own topic = "poisoned pill" crash-loop (flow5), responses_topic dropped in header parsing → __initial_responses_topic__ empty (4.2), missing in_response_to_request_id (081.b), fire-and-forget spawn ignoring init responses (flow12).
- **what:** The canon of failure modes that shaped the architecture: every major convention (data normalisation, reply-to storage, perspective transformation, single completion path, await semantics) exists as the fix to one of these traced production bugs. Valuable as diagnostic priors for any council debugging agent.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.016.flow5.md; docs001_flow_general/README.4.2.lifespanofresponsestopic.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.012.flow3.md
- **relations:** all system-architecture concepts above; debugging heuristics (docs 016b successor).
- **verify-later:** none — historical lessons.

### Deliberate discovery + human-approved agent evolution
- **category:** development-guide
- **status-signal:** abandoned
- **status-evidence:** README.005 principles ("Deliberate discovery — only at planning and review stages; Human approval — all agent changes require approval; Performance-based evolution") never reappear as a mechanism in later eras.
- **what:** Early governance rules for agent self-modification: the system only creates/modifies agents when starting a new task type, after poor performance review, and always with human approval — no heartbeats or automatic decisions. Paired with per-group performance recording and version incrementing.
- **sources:** docs001_flow_general/README.005.discovery.md
- **relations:** agent groups evolution service; HITL; tool-lifecycle health checks are the modern relative.
- **verify-later:** none.

### Website build overall plan v0 (first multi-agent website roadmap)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** README.050 is a 6-phase/12-step plan (domain-analyst → content-creator → html-developer, then site-architect/visual-designer/site-publisher, data contracts, spawn_group team) written against the calculator-era platform; every element was rebuilt differently in docs002–004.
- **what:** The first articulation of "build a website with agents": minimal 3-agent workflow, explicit JSON data contracts between agents, progressive enhancement, mock-LLM-first testing, upload_to_s3 deployment. Registers as the origin point of the entire site-building programme.
- **sources:** docs001_flow_general/README.050.overall_plan1.website_design.md; docs001_flow_general/README.001.actions.md (action inventory of that moment: many mocks — deploy_to_hosting, http_request, cache_lookup all fake)
- **relations:** superseded by MVP site builder, then the work-item pipeline.
- **verify-later:** none.
