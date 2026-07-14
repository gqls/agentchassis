# EXTRACTION U26 — docs/architecture, docs/humanintheloop, docs/basic_usage, docs/plans, docs/operations, docs/api
Extracted 2026-07-13. Files in scope: 45. Concepts found: 50.

Note on the source area: these are the OLDEST formal docs in the repo — mostly 2025-era
design conversations from the "AI Persona Platform / personae" phase of the platform
(namespace `ai-persona-system`, later `personae`). Much is superseded by the current
002-spine architecture (work items, site_plans, webdesign-agent) or was never built.
The durable survivors (chassis, stateless orchestration, DB-backed state, HITL
pause/resume, spawn/groups, three-database layout) are tagged with their real status.

## Coverage
| file | treatment |
|---|---|
| docs/api/reference.html | header-scan (generated Redoc bundle; titles + operation summaries extracted) |
| docs/architecture/001-agent-calls-agents-doc.md | full |
| docs/architecture/002-agent-chassis-docs.md | full |
| docs/architecture/003-flow-doc.md | full |
| docs/architecture/004-agent-chassis-architecture.md | full |
| docs/architecture/005-template-classification-and-evolution.md | full |
| docs/architecture/006-categorisation-template-search-evolution.md | full |
| docs/architecture/007-roadmap.md | full |
| docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md | full |
| docs/architecture/009-wordpress-discussion | full |
| docs/architecture/010-domain-value-maximisation.md | full |
| docs/architecture/011-example-domains | full |
| docs/architecture/012-investors.md | full |
| docs/architecture/014-Temporal-Airflow-adapters.md | full |
| docs/architecture/015-underserved-niche.md | full |
| docs/architecture/016-competitive-advantge.md | full |
| docs/architecture/017-audio-monitoring-discussion | full |
| docs/architecture/018-audio-monitoring-tech.md | full |
| docs/architecture/019-information-discovery-agent-spawning | full |
| docs/architecture/020-topic-amplifier-deep-digger.md | full |
| docs/architecture/021-current-framework-limitations | full |
| docs/architecture/022-possible-agent-structure | full |
| docs/architecture/023-spawning-agents.md | full |
| docs/architecture/024-agent-discovery-framework.md | full |
| docs/architecture/025-reusable-evolvable-agent-teams | full |
| docs/architecture/026-manual-vs-dynamic-templates | full |
| docs/architecture/027-create-website-creation-system | full |
| docs/architecture/030-content-creation-plan | full (stub: single ChatGPT link, no content) |
| docs/architecture/databases.md | full |
| docs/basic_usage/001basic_usage.txt | full |
| docs/basic_usage/002storage_of_results | full |
| docs/basic_usage/003_dynamic_prompt_improvement | full |
| docs/basic_usage/004_debugging | full |
| docs/humanintheloop/HITL_README.md | full |
| docs/humanintheloop/files.zip | skipped-binary (zip of the 9 sibling HITL files, dated 2025-11-03; contents identical to directory) |
| docs/humanintheloop/hitl_agent_definition.sql | full |
| docs/humanintheloop/hitl_agent_group_definition.sql | full |
| docs/humanintheloop/listen_for_approvals.sh | header-scan |
| docs/humanintheloop/monitor_workflow.sh | header-scan |
| docs/humanintheloop/quick_hitl_test.sh | header-scan |
| docs/humanintheloop/send_approval.sh | header-scan |
| docs/humanintheloop/start_hitl_workflow.sh | header-scan |
| docs/humanintheloop/verify_hitl_setup.sh | header-scan |
| docs/operations/README.md | full (empty file, 0 bytes) |
| docs/plans/stateless-first-agents-001 | full |

## Concepts

### Agent chassis — generic configurable agent executor
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md describes it as the running framework ("deploys as a scalable Kubernetes deployment, 3 replicas in production"); HITL agent definition (2025-11-03) still references image `docker.io/aqls/agent-chassis:v1.0.407`.
- **what:** A single reusable Go binary that becomes any agent type via configuration: it consumes Kafka messages, loads its workflow config from the database (agent_definitions / agent_instances), executes the workflow, handles fuel checks, errors, metrics, and health endpoints. New agent types are created by adding DB configuration, not code — "you're not creating new CODE, you're creating new CONFIGURATIONS".
- **sources:** docs/architecture/002-agent-chassis-docs.md; docs/architecture/023-spawning-agents.md#the-core-concept; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; agent spawning; distributed embedded orchestration
- **verify-later:** cmd/agent-chassis/main.go, platform/agentbase/, platform/messaging/processor.go, agent_definitions table

### Distributed embedded orchestration (no central orchestrator)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md: "every agent is both a worker and an orchestrator... eliminates single points of failure"; presented as a completed architecture report.
- **what:** Every agent pod embeds a full orchestrator (SagaCoordinator) instead of a central orchestration service. Any pod of an agent type can start a workflow, and any pod can pick up a response and continue it, because state is in the shared database. Key architectural decision distinguishing the platform from Temporal/Airflow-style central schedulers.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#distributed-orchestration-model; docs/architecture/003-flow-doc.md; docs/architecture/012-investors.md
- **relations:** stateless-first principle; database-backed workflow state; AI-native orchestration positioning
- **verify-later:** platform/orchestration/ (SagaCoordinator), orchestrations/orchestrator_state tables

### Database-backed workflow state (orchestrator_state → orchestrations)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md: "The orchestrator is stateless... Workflow state is in the database"; basic_usage/004 queries both the old `orchestrator_state` and a newer `orchestrations` table, showing the concept live through a schema evolution.
- **what:** All workflow execution state — status (RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED), current_step, workflow_plan, execution_path, collected_data, awaited_steps/awaited_requests, final_result — is persisted per correlation_id. Responses arriving at any pod are matched to awaited steps via causation_id, the state is updated, and the workflow continues when all responses are in.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/architecture/003-flow-doc.md; docs/basic_usage/004_debugging
- **relations:** stateless-first principle; fan-out and response correlation; workflow state machine
- **verify-later:** orchestrator_state and orchestrations tables in clients DB; column set differences between them

### Workflow-as-configuration (JSON workflows in agent definitions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md gives the canonical `{"start_step": ..., "steps": {...}}` shape; HITL definition (Nov 2025) still uses exactly this workflow JSON structure with `next_step` chaining.
- **what:** Agent behaviour is a JSON workflow (start_step + named steps, each with an action, config, and next_step) stored in agent_definitions.default_config / task_workflow, overridable per agent_instances. Contrasted with Temporal/Airflow where workflows are compiled code — here business users can create workflows without deployment.
- **sources:** docs/architecture/002-agent-chassis-docs.md#how-workflows-work; docs/humanintheloop/hitl_agent_definition.sql; docs/architecture/012-investors.md#dynamic-workflow-creation
- **relations:** agent chassis; execute_llm_prompt action; local vs remote actions
- **verify-later:** agent_definitions.task_workflow / orchestrator_workflow columns; workflow validator code

### Local vs remote actions and the action registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md documents both patterns as implemented ("Executed within the orchestrator itself" vs "Executed by other agents via Kafka").
- **what:** Workflow steps are either local actions run synchronously in the orchestrator (validate_input, transform_data, spawn_agent, process_data...) registered in a Go actionRegistry, or remote actions dispatched to another agent's Kafka topic with state moved to AWAITING_RESPONSES. The registry grew over time (spawn_agent, execute_llm_prompt, await_approval added later).
- **sources:** docs/architecture/004-agent-chassis-architecture.md#local-vs-remote-actions; docs/architecture/025-reusable-evolvable-agent-teams#step-3; docs/basic_usage/003_dynamic_prompt_improvement
- **relations:** agent-centric call_agent; fan-out; HITL await_approval
- **verify-later:** platform/orchestration/actions/ directory; actionRegistry in coordinator.go

### Fan-out and awaited-response correlation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md walks a live fan-out (reasoning + image agents) with correlation_id/causation_id header matching; 001basic_usage.txt shows fan_out steps in the deployed website-builder workflow.
- **what:** A fan_out step sends parallel sub-tasks to multiple agent topics, records their request IDs in awaited_steps, and sets status AWAITING_RESPONSES. Each response carries correlation_id (workflow) and causation_id (the originating request_id); any receiving pod matches causation_id to an awaited step, stores the result under collected_data, and resumes when all are received.
- **sources:** docs/architecture/003-flow-doc.md; docs/architecture/004-agent-chassis-architecture.md#response-handling-flow; docs/basic_usage/001basic_usage.txt
- **relations:** database-backed workflow state; kafka topic conventions; message header contract
- **verify-later:** fan_out action implementation; awaited_steps vs awaited_requests handling

### Kafka topic conventions (process/responses → requests/responses)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Early docs use `system.agent.{type}.process` + `system.responses.{type}` (004); the stateless plan (v21) and HITL scripts (Nov 2025) use `system.agent.generic.requests`, per-type `.requests/.responses/.errors/.dlq` topics stored in agent_definitions.topics — the newer form names the older one's replacement.
- **what:** Naming scheme for per-agent-type Kafka topics plus system topics (`system.notifications.ui`, `system.commands.workflow.resume`, `system.errors.*`, DLQs). Topics are per agent TYPE, not per instance; all replicas share a consumer group so Kafka distributes work. The convention itself is durable; the specific `.process` form was superseded by `.requests/.responses`.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#kafka-topic-structure; docs/plans/stateless-first-agents-001#7-kafka-configuration; docs/humanintheloop/hitl_agent_definition.sql (topics JSONB)
- **relations:** stateless-first principle; HITL notification/resume topics
- **verify-later:** actual topic list in cluster; topics column of agent_definitions

### Message header contract (sender identity, in_response_to_*, status enum)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 (marked "v21", heavily iterated) defines the full header set; quick_hitl_test.sh (Nov 2025) sends live messages carrying orchestration_id, orchestration_name, step_name, message_type, from_agent_type, responses_topic headers — the contract in use.
- **what:** Rich request/response headers: sender AgentIdentity (agent_type, agent_id=pod name, version), correlation_id + human-readable correlation_name, orchestration_id/name, step_id/name, request_id, retry_version, parent orchestration linkage, message_id, fuel budget, timeout, routing topics. Responses echo in_response_to_request_id/step/orchestration and carry a status enum: awaiting | processing | complete | error_recoverable | error_unrecoverable, plus multipart flags, timing and fuel accounting.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/send_approval.sh
- **relations:** retry semantics; message deduplication; fuel budget
- **verify-later:** Go header structs in platform code; kafka message headers on live topics

### Stateless-first agent principle
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Core Principle: Agents are stateless executors. State lives in the database. Any replica can process any message for its agent type" — presented as the implementation spec (v21) that matches later operational docs.
- **what:** Agents hold no orchestration state in memory; pod crashes lose nothing; replicas scale horizontally with HPA (CPU + kafka consumer lag metrics); Kafka consumer groups distribute work; messages for one orchestration are ordered by using orchestration_id as the partition key. Formalises and extends the earlier distributed-orchestration model.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/plans/stateless-first-agents-001#8-kubernetes-deployment
- **relations:** distributed embedded orchestration; orchestration-as-identity; optimistic locking
- **verify-later:** deployment manifests (HPA config), consumer group setup, partition key usage

### Orchestration-as-identity model (AgentID = PodName)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Orchestration (in DB) + Step + Request = 'Agent Instance'... AgentID = PodName (changes on restart, but that's OK)". This resolved the earlier mandatory-AgentID debate (022).
- **what:** The persistent identity of "an agent doing a task" is the orchestration record, not the pod. Pod name serves as AgentID purely for debugging (processing_history records which pod handled each step). Supersedes the doc-022 proposal that workflows resolve and pin specific versioned agent instances (stable/canary selection strategies) — that instance-pinning design was not carried forward.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/architecture/022-possible-agent-structure#the-case-for-mandatory-agentid
- **relations:** supersedes mandatory agent-instance resolution (022); stateless-first principle
- **verify-later:** whether processing_history with pod_name exists in current orchestrations table

### Optimistic locking on orchestration state
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Fully specified in stateless-first-agents-001 (version column, update_orchestration_if_version() SQL function, retry loop with backoff) but no later doc in this unit confirms it shipped.
- **what:** Each orchestration row carries a version integer; replicas load state, apply a step, and save only if the version is unchanged (compare-and-swap), retrying on mismatch. Prevents two replicas from double-processing the same step. Paired with processing_history JSONB as the audit trail of which pod did what.
- **sources:** docs/plans/stateless-first-agents-001#3-database-backed-state-management; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** stateless-first principle; message deduplication
- **verify-later:** version column and update function in current schema; conflict-retry code

### Retry semantics: same request_id, incremented retry_version
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** stateless-first-agents-001 "Key Implementation Notes: Retry uses same request_id with incremented retry_version"; error_recoverable responses trigger up to 3 retries. No later confirmation in this unit.
- **what:** Failed remote calls are retried with the identical request_id and retry_version+1 so responses remain matchable and duplicates detectable. Recoverable errors retry (max 3), then fall through to unrecoverable which fails the orchestration and propagates an error to the parent. Progress statuses (awaiting/processing) are logged but never propagated upward; terminal states are processed exactly once.
- **sources:** docs/plans/stateless-first-agents-001#6-retry-logic; docs/plans/stateless-first-agents-001#key-implementation-notes
- **relations:** message header contract; message deduplication
- **verify-later:** retry handling in response processing code; awaited_requests (request_id, retry_version) PK

### Message deduplication (processed_messages, terminal-state-once)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Designed in detail (dedup key request_id:retry_version:status; processed_messages table with 24h cleanup) in stateless-first-agents-001; no operational evidence in this unit.
- **what:** Before processing, agents check a dedup key against a processed_messages table (or in-memory map); duplicate responses are dropped, and once any terminal state (complete/error_unrecoverable) is processed for a request, all further terminal responses for it are ignored. Ensures idempotency under Kafka redelivery and multi-replica consumption.
- **sources:** docs/plans/stateless-first-agents-001#7-deduplication-handler; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** retry semantics; optimistic locking
- **verify-later:** processed_messages table existence; dedup logic in message consumption path

### Fuel budget resource management
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md lists fuel management as a chassis feature ("Checks fuel budget from headers... Prevents execution if insufficient fuel"); fuel_budget=1000 header sent in live kcat commands (basic_usage 001/004); response headers carry FuelUsed/RemainingFuelBudget.
- **what:** Every workflow carries a fuel budget header; actions deduct fuel costs; sub-invocations pass a reduced budget down and report fuel used back up the chain. Serves as the cost/abuse control across multi-agent workflows. Current status in the 2026 platform unverified — no recent doc in this unit mentions it.
- **sources:** docs/architecture/002-agent-chassis-docs.md#key-features; docs/basic_usage/001basic_usage.txt; docs/plans/stateless-first-agents-001 (FuelUsed/RemainingFuelBudget headers)
- **relations:** message header contract; subscription/quota API
- **verify-later:** fuel handling in chassis code; whether current work-item system retains fuel

### Agent-centric architecture: steps call agents, not topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 022-possible-agent-structure proposes `call_agent` with agent_type replacing raw topics ("your current code already does 90% of this"); 027 and the HITL group definition (Nov 2025) use call_agent steps in production seeds.
- **what:** The primary abstraction is the agent (owning a 6–12 step workflow) rather than the workflow; steps invoke other agents (`action: call_agent, agent_type: X`) which have their own workflows, error boundaries and state, enabling recursive hierarchies (any agent can orchestrate, a copywriter can spawn a researcher). Topic resolution happens from agent type.
- **sources:** docs/architecture/022-possible-agent-structure#summary-agent-centric-architecture; docs/humanintheloop/hitl_agent_group_definition.sql; docs/architecture/023-spawning-agents.md#the-orchestrator-is-a-pod-too
- **relations:** agent chassis; agent spawning; supersedes inter-agent invocation protocol v1
- **verify-later:** call_agent action and agent-type→topic resolution code

### Inter-agent invocation protocol v1 (invoke_agent / agent_invocations)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** 001-agent-calls-agents-doc.md proposes InvokeAgentAction, ParallelInvokeAgentsAction, and an agent_invocations tracking table; later docs (022, stateless plan) replace this with call_agent + orchestration hierarchy headers, and the agent_invocations table never reappears.
- **what:** The first design for agent-calls-agent: a dedicated invocation request/response envelope, per-pair topics (`system.agent.requests.{from}.{to}`), an agent_invocations audit table, and parent_correlation_id columns. Its essential ideas (parent linkage, deadline, fuel passing) survived into the header contract; the specific mechanism did not.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#1.2; docs/architecture/001-agent-calls-agents-doc.md#phase-3
- **relations:** superseded by call_agent (022) and stateless header contract; project manager agent
- **verify-later:** confirm agent_invocations table absent from schema

### Project Manager / User Representative agent hierarchy
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Designed across 001 and 007 ("User Representative Agent... represents the users views against the project manager"); never appears in later seeds, groups, or the current 002-spine — silently vanishes after the website-builder group takes the orchestrator role.
- **what:** A top-level persona hierarchy: User → Project Manager agent (plans phases, delegates to specialist orchestrators, reviews deliverables) → Web Design Orchestrator → specialists, with a User-Persona agent negotiating on the user's behalf (stores preferences, approves/rejects deliverables). The review/approval intent resurfaced later as HITL steps and content governance instead.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#architecture-overview; docs/architecture/007-roadmap.md#2.1-user-representative-agent
- **relations:** website-builder group (took its place); HITL approval mechanism (absorbed the review role)
- **verify-later:** confirm no project-manager/user-representative agent_definitions exist

### Agent spawning (agents as DB records claimed by generic pods)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001: "The spawn_group action creates new rows in the agent_instances table"; 004_debugging references `kubectl delete jobs -l spawned-by=orchestrator` — orchestrator-spawned jobs existed in the running cluster.
- **what:** A spawn_agent action creates (or reuses) an agent_instances row with type, workflow, capabilities and llm_config; a generic chassis pod (static env assignment, dynamic pool claim, or K8s Job spawned by the orchestrator) loads that config and becomes the agent. Includes existence-check-and-reuse logic and pod-type selection (CPU/GPU/memory) for specialised workloads.
- **sources:** docs/architecture/023-spawning-agents.md; docs/architecture/025-reusable-evolvable-agent-teams#step-1; docs/basic_usage/001basic_usage.txt#part-2
- **relations:** agent chassis; agent groups; agent discovery
- **verify-later:** spawn_actions.go; assigned_pod/status/last_heartbeat columns; Job-spawning code

### Agent groups (reusable multi-agent teams)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 documents live spawn_group runs against the agent_groups table; hitl_agent_group_definition.sql (Nov 2025) inserts into a matured `agent_group_definitions` table with versioning and ON CONFLICT upsert.
- **what:** A named, versioned team definition: group_type, agent_configs (roles → agent types), an orchestration_workflow describing how they cooperate, capabilities/tags for search, and usage/performance metadata. `spawn_group` instantiates the team and starts the group workflow. Groups were intended to be saved from successful configurations, forked, and improved over generations.
- **sources:** docs/architecture/025-reusable-evolvable-agent-teams#phase-2; docs/humanintheloop/hitl_agent_group_definition.sql; docs/basic_usage/001basic_usage.txt
- **relations:** website-builder group; controlled group evolution; group discovery
- **verify-later:** agent_groups vs agent_group_definitions tables; SpawnGroupAction code

### Agent and group discovery by capability and performance
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 024-agent-discovery-framework.md reads as pure proposal ("Yes! Let's design this") with an unconfirmed agent_metrics/heartbeat table, BUT a live repo check (not just docs) found `platform/discovery/` actually exists — `agent_discovery.go` plus `README.001.agentdefinitions.md` and `README.002.dbtables.md` — confirming a real, if unknown-scope, implementation landed. Upgraded from aspirational on that direct code evidence; the sophistication of the shipped version (whether performance-ranking/heartbeats made it in) is still stage-2 work.
- **what:** A registry service that finds the best existing agent (or group) for a task by required capabilities (JSONB containment), success rate, response time, availability (heartbeat) and fuel cost — spawning a new one only when nothing matches. The "self-organizing system" goal: agents discover each other, learn which perform best, optimise over time.
- **sources:** docs/architecture/024-agent-discovery-framework.md; docs/architecture/027-create-website-creation-system#phase-3; docs/architecture/025-reusable-evolvable-agent-teams#group-discovery-service
- **relations:** agent spawning; agent groups; template classification
- **verify-later:** platform/discovery/agent_discovery.go, platform/discovery/README.001.agentdefinitions.md, platform/discovery/README.002.dbtables.md — confirm scope of what actually shipped vs. this proposal

### Workflow template library, lineage and marketplace
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 005/006 propose workflow_templates + template_versions tables, fork lineage graphs and a monetised marketplace ("marketplace of evolving workflow templates"); roadmap slots it at Phase 6 (weeks 10-12); no later doc in the repo era references these tables.
- **what:** Successful workflow executions are saved as reusable templates with lineage (parent_template_id, source_correlation_id), performance metrics, ratings and usage counts; users fork and improve templates ("collective intelligence — natural selection of best-performing templates"), with a WorkflowOptimizer suggesting parallelisation/reordering from execution history. A precursor idea whose spirit partially survives in agent group versioning.
- **sources:** docs/architecture/005-template-classification-and-evolution.md; docs/architecture/007-roadmap.md#phase-6; docs/architecture/016-competitive-advantge.md#workflow-marketplace-2.0
- **relations:** agent groups; template classification and search; controlled group evolution
- **verify-later:** confirm workflow_templates tables never created

### Multi-dimensional template classification and semantic search
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** abandoned
- **status-evidence:** 006 is pure design (behavioral profiles, performance vectors, embedding fingerprints, vector DB + graph DB search service, recommendation engine); none of the infrastructure (vectorDB for templates, template_usage_metrics) appears in later docs.
- **what:** Replace flat tags with a rich classification for discovering workflows/templates: behavioral capabilities, execution style, resource usage, normalised performance vectors with trade-off scores, embedding-based semantic fingerprints, outcome-based deliverable descriptions and evolutionary metadata (lineage depth, fork count). A parallel multi-strategy search (semantic + behavioral + performance + lineage) with collaborative-filtering recommendations.
- **sources:** docs/architecture/006-categorisation-template-search-evolution.md
- **relations:** workflow template library; agent/group discovery
- **verify-later:** n/a (idea only) — check nothing similar exists under another name

### Controlled group evolution (observed mutation with rules)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** partial
- **status-evidence:** 026 reads as pure future-tense recommendation ("Start with curated base templates that can dynamically evolve", MutationRules, sandbox testing, 10%-improvement gates) with nothing downstream in this unit confirming it. BUT a live repo check found `platform/evolution/` actually exists — `evolution.go`, `performance.go`, and a README stating in the present tense: "Evolution Service: Evaluates groups for potential improvements / Applies mutations (parallel agents, specialists, validators) / Tracks evolution history / Version management" and "Performance Analysis: Records and analyzes execution metrics / Identifies bottlenecks and failures / Generates improvement suggestions". Upgraded from abandoned on that direct code evidence — this idea did get built, just not confirmed anywhere in the docs sampled for this unit.
- **what:** Hybrid manual/dynamic strategy: hand-curated base agent and group templates act as "genetic seeds"; a metrics observer detects bottlenecks/missing capabilities after ≥5 uses and proposes constrained mutations (add parallel agent, add specialist, replace, adjust workflow, fork) which must beat a performance baseline in sandbox before becoming a new version with parent lineage. Human-in-the-loop approval tiers for major changes.
- **sources:** docs/architecture/026-manual-vs-dynamic-templates; docs/architecture/025-reusable-evolvable-agent-teams#usage-example
- **relations:** agent groups; dynamic prompt improvement loop; workflow template library
- **verify-later:** platform/evolution/evolution.go, platform/evolution/performance.go, platform/evolution/README.md — confirm how closely the shipped mutation rules match this design

### Dynamic prompt improvement loop (Prompt Improvement Agent)
- **category:** NEW:agent-spawning-and-groups
- **status-signal:** aspirational
- **status-evidence:** basic_usage/003 is a plan ("Phase 2: The Evolution Loop") — flag-for-improvement UI action, `system.prompt.improvement.request` topic, save_new_agent_definition action creating versioned definitions (html-developer-v2); the agent_definitions versioning columns (version, previous_version_id) DID ship (visible in hitl_agent_definition.sql), the loop itself is not evidenced.
- **what:** End-of-workflow human review offers "Approve" or "Flag for Improvement"; flagged runs dispatch the failing agent's prompt + failure context to a prompt-engineering specialist agent which generates an improved prompt, gets human approval, and saves a NEW versioned agent definition (never mutating the old). Includes bootstrap_prompt for generating a first prompt for brand-new agent types from a description.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement; docs/humanintheloop/hitl_agent_definition.sql (version/previous_version_id columns)
- **relations:** execute_llm_prompt; HITL approval mechanism; controlled group evolution; (spiritual ancestor of the current improvement-loop / finetuning flywheel)
- **verify-later:** prompt-improvement-agent definition; save_new_agent_definition action; version columns usage

### execute_llm_prompt generic action with DB prompt templates
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Planned in basic_usage/003 ("the reusable 'chef' that cooks the 'recipes'"); in live use by Nov 2025 — hitl_agent_definition.sql's workflow uses `"action": "execute_llm_prompt"` with a Go-template prompt_template.
- **what:** A single generic action that reads the agent's prompt_template and ai_service config (provider, model, api_key_env_var) from its definition, renders the template with Go text/template placeholders ({{.input_data.field}}) filled from collected workflow data, calls the configured LLM, and returns the text. Makes every LLM agent a pure data configuration.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement#step-1.2; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; dynamic prompt improvement loop
- **verify-later:** platform/orchestration/actions/ai_actions.go; prompt template rendering

### Website-builder agent group (six-specialist pipeline)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Ran in production per basic_usage/001 ("Step-by-Step Guide to Your First Website Build", migrations 005/007/009 referenced); the current platform builds sites via the site_plans domain / webdesign-agent pipeline (002 spine, docs 029/030), which replaced this group.
- **what:** The original end-to-end website creation flow: an orchestrator agent calls domain-analyst (business categorisation via web-search) → site-architect (page structure, pausing for human approval) → fan-out of content-researcher + visual-designer (image search/generation, logo) → html-developer (per-page vanilla HTML/CSS fan-out) → site-publisher (s3_upload, preview URL). Seeded as agent_definitions + an agent_groups row; triggered by one spawn_group Kafka message.
- **sources:** docs/architecture/027-create-website-creation-system; docs/basic_usage/001basic_usage.txt; docs/basic_usage/003_dynamic_prompt_improvement#step-1.1
- **relations:** superseded by site_plans + webdesign-agent + design-composition pipeline; HITL pause in site-architect; result storage split
- **verify-later:** migrations 005/007/009 in platform/database/migrations/; whether group still seeded

### HTML-first progressive enhancement delivery
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 008: "Starting with plain HTML/CSS/JS is actually a very smart architectural decision"; the html-developer seeds specify "vanilla" HTML with inline CSS; the current platform still renders plain HTML/CSS sites (render pipeline docs).
- **what:** Deliberate decision to generate plain HTML/CSS/JS websites rather than framework apps: easier for AI to generate and validate, no build step, universally hostable, fast; complexity added progressively (web components → PWA → framework only if needed). One of the few strategy decisions from this era that demonstrably survived into the present render pipeline.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#why-simple-html-css-js-is-the-right-start; docs/architecture/027-create-website-creation-system (html-developer config)
- **relations:** styling-render-pipeline (current successor context); WordPress export agent (rejected sequel)
- **verify-later:** current renderer output format

### WordPress export agent and content-subscription plugin
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 008 designs it enthusiastically ("This is GENIUS!"); 009 immediately deconstructs it ("You're not unique in 'AI builds WordPress sites'. That market is saturated... Only add WordPress if they're begging for it") — never mentioned again anywhere.
- **what:** An agent converting generated HTML sites into installable WordPress themes + WXR content exports, paired with a WP plugin subscribing to the platform for auto-published fresh content (recurring revenue). Explicitly killed by the competitive analysis in 009, which redirected differentiation toward "sites that update themselves" / continuous content ecosystems.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#wordpress-export-agent-design; docs/architecture/009-wordpress-discussion#the-hard-truth
- **relations:** HTML-first delivery; living-content differentiation (which the platform did pursue via news feed / content pipelines)
- **verify-later:** n/a (never built)

### Domain value maximisation pipeline (domain flipping)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 010/011/015 lay out the strategy against the user's real domain portfolio (collateralfinancing.com, holidaytime.com, websitedesign.com...) with 48-hour development timelines; no later doc pursues domain flipping — the platform pivoted to operating its own sites.
- **what:** Use the agent platform to develop parked domains into sites with content, traffic and revenue to multiply sale value (naked $500 → revenue-bearing $10k+): domain classification (brandable/exact-match/local/product), tiered portfolio treatment, 48h batch development, monetisation setup (leads/affiliate/ads), and "self-selling" footers that market the build service from every developed domain.
- **sources:** docs/architecture/010-domain-value-maximisation.md; docs/architecture/011-example-domains; docs/architecture/015-underserved-niche.md#your-domain-portfolio-is-your-marketplace
- **relations:** deep-research domain insight agent; underserved-niche strategy; site-case-studies (the surviving practice of operating exemplar sites)
- **verify-later:** n/a

### Underserved-niche and vertical showcase strategy
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 015/016 propose niches (compliance docs, local-business packages, academic assistants, affiliate content) and per-industry showcase domains with a workflow marketplace ("DIY $500 / DFY $200mo / White Label $2000mo"); positioning discussion only, no implementation trail.
- **what:** Rather than competing with Temporal/LangChain/Zapier broadly, own narrow verticals where multi-agent coordination wins: each showcase domain demos an industry solution (legal docs, restaurant launch, real-estate listings) and funnels to purchasable workflows. Includes the pricing-tier and "Business-in-a-Box" (site + content pipeline + email + social) framings, and the investor-demo positioning of the framework as the star with swappable use cases.
- **sources:** docs/architecture/015-underserved-niche.md; docs/architecture/016-competitive-advantge.md#who-actually-pays-for-ai-sites; docs/architecture/012-investors.md#the-portfolio-approach
- **relations:** domain value maximisation; EBORG organizational OS
- **verify-later:** n/a

### AI-native orchestration positioning (vs Temporal/Airflow)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 012/014 are interview/investor argumentation ("You could build this on Temporal, but it would be like using Kubernetes to run a single container"); the accompanying Temporal/Airflow adapter agents were never built.
- **what:** The articulated build-vs-buy rationale for the platform: AI-specific needs (dynamic JSON workflows without deployment, token/fuel tracking, prompt management, AI-failure handling, multi-tenant agent isolation, workflows spawning from AI decisions) justify a purpose-built orchestrator. Includes proposed Temporal-adapter and Airflow-adapter agents to bridge into enterprise workflow estates as a migration path ("we don't replace your existing systems, we enhance them").
- **sources:** docs/architecture/012-investors.md#better-answer; docs/architecture/014-Temporal-Airflow-adapters.md
- **relations:** adapters (current adapter guide is a different, real lineage); distributed embedded orchestration
- **verify-later:** confirm no temporal/airflow adapter code exists

### Deep-research domain insight agent
- **category:** research-agents
- **status-signal:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent; docs/architecture/016-competitive-advantge.md#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage
- **verify-later:** n/a

### Audio-monitoring topic discovery with auto-spawned topic agents
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 017/018 fully design the pipeline (Bloomberg/podcast transcription via Whisper → topic extraction → novelty check → spawn agent) with a phased plan starting "Week 1: financial podcasts"; nothing downstream ever references it.
- **what:** A self-expanding intelligence network: audio streams/podcasts are transcribed, novel topic clusters detected (novel-phrase and frequency-spike detection against a 30-day corpus), and a specialised monitoring agent is automatically spawned per new topic (sources, sentiment, players, trajectory, content generation, subscriber alerts) — "Bloomberg mentions topic at 9:00 AM, your system publishes analysis by 9:30". Included a Domain Intelligence Orchestrator (DIO) deciding which intelligence strategy fits each domain.
- **sources:** docs/architecture/017-audio-monitoring-discussion; docs/architecture/018-audio-monitoring-tech.md#realistic-implementation-path
- **relations:** topic amplifier engine; agent spawning; cross-domain intelligence network
- **verify-later:** n/a

### Topic amplifier / deep digger engine
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 019/020 catalogue the hard problems and Python component designs (MinHash LSH dedup, spaCy extraction, verification engine, source discovery, PG+Elasticsearch+Redis storage) with a 6-week plan; no implementation trace exists.
- **what:** The engineering backbone for topic intelligence: data collection (news/social/RSS/scraping), temporal tracking with velocity/anomaly detection, structured extraction (dates, money, entities), claim verification against trusted sources, source discovery (link following, social-graph expansion, citation mining), scalable near-duplicate detection, and a hybrid division of labour — LLMs for context understanding/relevance/noise-filtering (rated "very strong"), traditional code for collection, temporal/quantitative analysis, dedup and storage. Honest bootstrap/noise/evolution problem analysis included.
- **sources:** docs/architecture/020-topic-amplifier-deep-digger.md; docs/architecture/019-information-discovery-agent-spawning#the-honest-assessment; docs/architecture/019-information-discovery-agent-spawning#llms-in-the-loop
- **relations:** audio-monitoring topic discovery; deep-research domain insight agent
- **verify-later:** n/a

### Cross-domain intelligence network and subscription tiers
- **category:** NEW:topic-intelligence
- **status-signal:** abandoned
- **status-evidence:** 016's "Hidden Superpowers" section (living knowledge graphs, insight arbitrage between domains, $10/$99/$999/$9,999 subscription tiers, "Organizational OS") is pure vision with no follow-through in later documentation.
- **what:** Developed domains share intelligence: patterns detected on one site alert sibling sites to opportunities ("vehicle-hire.com notices courier demand spike → couriervans.com gets alert"); accumulated contextual memory, relationship mapping and time-series pattern recognition become sellable subscriptions (industry intelligence, trend prediction, competitive clusters) and ultimately an org-wide agent deployment ("every employee gets a personal agent dashboard").
- **sources:** docs/architecture/016-competitive-advantge.md#the-hidden-superpowers-of-your-system; docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** EBORG; audio-monitoring topic discovery; business-strategy subscription models
- **verify-later:** n/a

### EBORG — Evidence-Based Organisational Planning
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** The Nov 2025 HITL demo is branded "For EBORG" with a full pitch paragraph in start_hitl_workflow.sh ("mapping every role, responsibility, and objective, then pairing each with a framework of AI agents"); no other doc in this unit elaborates whether it became a product.
- **what:** A product concept: organisations map roles/responsibilities/objectives and pair each with AI agents that gather research, assess options and provide evidence-based reasoning — "a human-centered, continuously learning organisation" and the concrete descendant of 016's Organizational OS idea. Used as the demo business in the HITL content-approval workflow.
- **sources:** docs/humanintheloop/start_hitl_workflow.sh; docs/humanintheloop/hitl_agent_definition.sql (header comment); docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** cross-domain intelligence network; HITL content-approval demo
- **verify-later:** any EBORG references in business/vonc docs (other units)

### HITL approval pause/resume mechanism
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** HITL_README.md documents the working flow end-to-end with real topics and states; the workflow pauses in AWAITING_RESPONSE and resumes on a message to system.commands.workflow.resume — dated Nov 2025 with a working test kit.
- **what:** Workflows pause on an `await_approval` (earlier `pause_for_human_input`) step: an approval request (title, description, generated content, metadata) is published to `system.notifications.ui` carrying the correlation_id and a request_id that serves as the approval token; a human (or later, an API) publishes a response to `system.commands.workflow.resume` with in_response_to_request_id = token and an approved/comments body; the orchestration resumes, exposing approval status/comments/approver/timestamp to subsequent steps. Timeout configurable (default 300s).
- **sources:** docs/humanintheloop/HITL_README.md; docs/humanintheloop/hitl_agent_definition.sql (await_human_approval step); docs/humanintheloop/send_approval.sh
- **relations:** workflow state machine (AWAITING_RESPONSE); HITL API integration; conditional approval branching
- **verify-later:** await_approval action code; resume command consumer; docs011_api_hitl for the API successor

### HITL content-approval demo agent and group
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Complete SQL seeds (agent + group, versioned, ON CONFLICT upsert) dated 2025-11-03, referencing chassis image v1.0.407 — a working, loadable demo.
- **what:** `simple-content-writer-with-approval` agent: generate_draft (execute_llm_prompt, Claude 3.5 Sonnet) → await_human_approval → process_approval (merges content with approval metadata) → complete. Wrapped by the `content-approval-hitl` group whose orchestration spawns the writer, calls it with business input data, and aggregates results. The canonical minimal HITL example for the platform.
- **sources:** docs/humanintheloop/hitl_agent_definition.sql; docs/humanintheloop/hitl_agent_group_definition.sql
- **relations:** HITL approval mechanism; agent groups; execute_llm_prompt
- **verify-later:** whether these definitions are loaded in current DB

### HITL kcat test harness
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Six working shell scripts (listen, start, send approval, monitor, quick one-liner, verify setup) using kubectl-run kcat pods against the personae namespace; README walks the three-terminal procedure.
- **what:** Manual test kit for the HITL loop: listen_for_approvals.sh tails system.notifications.ui; start_hitl_workflow.sh / quick_hitl_test.sh publish an orchestrate request for the content-approval-hitl group with full headers; send_approval.sh publishes the resume message with the approval token; monitor_workflow.sh polls orchestrator_state (status, current step, awaited_requests, collected approval data); verify_hitl_setup.sh checks definitions, topics and chassis pods exist.
- **sources:** docs/humanintheloop/HITL_README.md#testing-the-hitl-flow; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/verify_hitl_setup.sh
- **relations:** HITL approval mechanism; kcat/db-inspector ops runbook
- **verify-later:** script topic names against current cluster topics

### Conditional approval branching
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#customization shows a `conditional_branch` step keyed on `{{.await_human_approval.approved}}` routing to finalize vs regenerate — offered as a customisation, not present in the shipped demo workflow.
- **what:** Approval outcomes drive workflow branching: approved → finalise; rejected → regenerate content and re-submit for approval, enabling iterative human-guided refinement loops rather than binary pass/fail.
- **sources:** docs/humanintheloop/HITL_README.md#conditional-approval
- **relations:** HITL approval mechanism; dynamic prompt improvement loop (flag-for-improvement variant)
- **verify-later:** conditional_branch action existence

### HITL API integration (approvals via REST/UI)
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** HITL_README.md#api-integration-future: "Once the API is ready, the manual approval process can be replaced with: REST endpoint to fetch pending approvals, Web UI..., API call to approve/reject with the approval token."
- **what:** Planned replacement of manual Kafka approval messages with a REST surface and web UI over the identical underlying mechanism (same topics, same tokens). The taxonomy's hitl category (docs011_api_hitl) suggests this was subsequently pursued — this doc records the origin point.
- **sources:** docs/humanintheloop/HITL_README.md#api-integration-future
- **relations:** HITL approval mechanism; docs011_api_hitl (likely successor, other unit)
- **verify-later:** docs011_api_hitl extraction; approval endpoints in API code

### Three-database architecture (MySQL auth + PG clients + PG templates)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md is a factual "Database Architecture Summary" (not a proposal); basic_usage/001 contains live credentials/connection commands for both MySQL (rs17.uk-noc.com) and postgres-clients; current taxonomy (011) still lists MySQL + client schemas.
- **what:** Authentication isolated in MySQL (users, JWT refresh tokens, profiles, projects, subscriptions/tiers, permissions, activity logs; BINARY(16) UUIDs); agent/AI runtime in PostgreSQL clients DB (global agent_definitions, orchestrator_state, clients_info + per-client schemas); shared persona templates in a second PostgreSQL DB. Core Manager owns clients/templates access with read-only auth access.
- **sources:** docs/architecture/databases.md; docs/basic_usage/001basic_usage.txt
- **relations:** schema-per-client multi-tenancy; AI Persona Platform API (auth endpoints)
- **verify-later:** current DB inventory; whether templates DB still exists separately

### Schema-per-client multi-tenancy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md documents `create_client_schema()` and per-client tables as implemented; every operational query in basic_usage targets `client_demo_client.*` schemas.
- **what:** Each client gets an isolated PostgreSQL schema (client_{id}) containing agent_instances, agent_memory (pgvector embeddings), projects, workflow_executions and usage_analytics, created by a SQL function; global resources (agent_definitions, templates, orchestrator_state) are shared. Strong tenant isolation on shared infrastructure.
- **sources:** docs/architecture/databases.md#2-postgresql-database-1; docs/basic_usage/001basic_usage.txt
- **relations:** three-database architecture; agent spawning (instances live per-schema)
- **verify-later:** create_client_schema function; pgvector agent_memory usage in current code

### Workflow monitoring REST endpoints
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 004-agent-chassis-architecture.md lists GET /monitor/workflows, /monitor/workflow/{id}, /monitor/stuck?hours=n, /monitor/metrics as built ("Each agent exposes monitoring endpoints") but no later doc in this unit uses them — operational debugging instead goes through psql/db-inspector.
- **what:** Per-agent HTTP monitoring API over orchestration state: list active workflows per client, inspect a workflow's execution path/state, find stuck workflows not progressing for N hours, and aggregate metrics. Complemented by per-step execution_path timing records and execution_metadata counters in the state row.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#monitoring-and-observability
- **relations:** database-backed workflow state; kcat/db-inspector runbook (the surviving practice); current debugging docs (016/016b spine)
- **verify-later:** /monitor routes in chassis HTTP server code

### kcat + db-inspector operational runbook
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** basic_usage/001 and 004 are working command logs (with real outputs pasted, e.g. correlation IDs returned, "0 rows" failure cases) for triggering and tracing workflows in the live cluster.
- **what:** The early ops playbook: scale deployments up/down; inject workflow-start messages via kcat from an in-cluster pod with full header sets; fetch the latest correlation_id from orchestrator_state; watch progress with the db-inspector tool (-watch); trace specific agents by finding spawned instance IDs then grepping shared chassis pod logs (agents don't get dedicated pods); check consumer-group lag, response topics, ServiceAccount job-creation rights, and events for spawned jobs.
- **sources:** docs/basic_usage/001basic_usage.txt; docs/basic_usage/004_debugging
- **relations:** agent spawning; website-builder group; current debugging spine (016)
- **verify-later:** tools/db-inspector, tools/kafka-producer existence; whether runbook matches current namespace/topics

### Result storage split (DB paper-trail + S3 artefacts)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** basic_usage/002 states it as fact: final_result column in orchestrator_state, a website_projects table per client schema with preview/live URLs, and site-publisher's s3_upload of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated final_result JSON + website_projects metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product".
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system (site-publisher s3_upload)
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs

### AI Persona Platform public API
- **category:** admin-dashboard-and-api
- **status-signal:** superseded
- **status-evidence:** docs/api/reference.html is a generated Redoc bundle titled "AI Persona Platform API" covering the persona-era surface; the current API surface is the admin dashboard/nginx gateway (spine 012) and the persona-instance concepts do not appear in current docs.
- **what:** The v1 REST surface of the persona era: JWT auth (register/login/refresh/validate/logout), user profile/password/delete, projects CRUD, subscription with usage stats and quota checks, persona template listing, persona instance list/create, health check, and a WebSocket connection endpoint. Documents the original productisation of the platform as "AI personas" for end users.
- **sources:** docs/api/reference.html (tags: Authentication, Users, Projects, Subscriptions, Templates, Instances, System, WebSocket; paths /api/v1/auth/*, /api/v1/projects, /api/v1/subscription/*, /api/v1/personas/instances)
- **relations:** three-database architecture (auth DB backs these endpoints); superseded by current admin-dashboard-and-api (012)
- **verify-later:** which endpoints survive in core-manager/api-gateway code

### Workflow status state machine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Consistent across eras: RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED in 004; RUNNING / AWAITING_RESPONSE / COMPLETED / FAILED in HITL_README (Nov 2025); pending|processing|complete|failed variant in the stateless plan.
- **what:** The orchestration status vocabulary and its transitions: workflows run steps, park in an awaiting state while remote/human responses are outstanding, and terminate complete or failed. The HITL pause reuses the same awaiting state rather than introducing a special paused status. Minor naming drift across eras is itself a verification target.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/humanintheloop/HITL_README.md#workflow-states; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** database-backed workflow state; HITL approval mechanism
- **verify-later:** canonical status enum in current schema/code

### Human-readable orchestration and correlation names
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 mandates orchestration_name / correlation_name alongside UUIDs ("website-build-agrivisionary", "core-mgr-website-flow-0902-1030"); start_hitl_workflow.sh generates ORCHESTRATION_NAME="eborg-content-approval-$(date...)" in practice.
- **what:** Every orchestration and correlation carries a generated human-readable name in addition to its UUID, propagated through headers and stored in state, so debugging and monitoring read as narrative ("which pods processed core-mgr-website-flow") rather than UUID archaeology.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/start_hitl_workflow.sh
- **relations:** message header contract; kcat/db-inspector runbook
- **verify-later:** name-generation code; name columns in orchestrations table

### Agent teams: composite/family/service-agent patterns
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** 021 evaluates three options (PM pattern, peer-to-peer squads, service-oriented) and recommends service-oriented-then-squads; what actually shipped was the simpler agent-groups + call_agent model — the AgentFamily/SharedMemory and workflow-composition (sub-workflow) constructs never reappear.
- **what:** Design exploration for complex 50+-step workflows: composite agents (one external face, embedded sub-components), agent families with shared state and peer coordination, stateless reusable service-agents (date extractor, entity extractor) callable by any workflow, and workflows-invoking-workflows composition. Records the acknowledged framework limitation ("one agent = one workflow, flat orchestration, no concept of agent teams") that agent groups later addressed.
- **sources:** docs/architecture/021-current-framework-limitations; docs/architecture/022-possible-agent-structure
- **relations:** agent groups (the shipped resolution); agent-centric architecture
- **verify-later:** n/a — confirm no AgentFamily/sub-workflow constructs in code
