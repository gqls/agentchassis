| HITL-001 | Work item approval_mode (auto / hitl / eval) | partial | site_work_items.approval_mode column; auto/hitl live, eval defined but unused | hitl.md |
| HITL-002 | Confirm-not-initiate governance/HITL model (decision package) | aspirational | Agent-led reasoning + decision package, human confirms; new version deprecates old | hitl.md |
| HITL-003 | User-representative advocate (intent + conflict triage) | aspirational | Standing advocate triaging claimed conflicts before they reach the user | hitl.md |
| HITL-004 | Decision authority (co-equal voices, abstention, creator veto) | aspirational | Advocate/curator disagreement escalates to human; creator holds veto | hitl.md |
| HITL-005 | Human direction channels + lock lifecycle + audit-pass cap | partial | Three human-steering channels, content locks, 3-pass audit cap | hitl.md |
| HITL-006 | intake-orchestrator (classify → brief → HITL confirm → spawn builder → rerender) | superseded | v1/v2 entry pipeline with two HITL gates, superseded by domain-submitter | hitl.md |
| HITL-007 | content-reviewer (HITL + auto-eval dual mode with pre-validation) | deployed | Reviews page content via human or auto-eval mode with link/email pre-validation | hitl.md |
| HITL-008 | Human change-request work items | deployed | Human-submitted edits enter the same priority-ordered work queue as agent items | hitl.md |
| HITL-009 | input_requests HITL persistence | deployed | Persisted human input requests with pending view, expiry, and Kafka reply routing | hitl.md |
| HITL-010 | Manual HITL continuation runbook | deployed | Operational procedure for un-sticking HITL flows via awaited_requests + kcat | hitl.md |
| HITL-011 | HITL approval-as-specialised-agent architecture (human-reviewer plan) | aspirational | Never-built plan: dedicated human-reviewer agent, approval_tasks/versions tables | hitl.md |
| HITL-012 | await_approval / HITL pause-resume mechanism (AwaitedRequests, token matching) | deployed | Core pause/resume primitive: token-based approval request, matched Kafka resume | hitl.md |
| HITL-013 | approval_requests table | partial | Persistence for pending approvals; write path initially stubbed | hitl.md |
| HITL-014 | Content-type-aware approval capabilities (text edit / image replace) | aspirational | Approvals adapt UI/edit affordances by content type (text vs image) | hitl.md |
| HITL-015 | HITL approval timeouts (config mapping, defaults, restart recovery) | partial | timeout_seconds never mapped to Step.Timeout; fix plan incl. goroutine recovery | hitl.md |
| HITL-016 | process_approval_decision and rejection routing | deployed | Unpacks approval decision into CollectedData; branches continue/stop/reject | hitl.md |
| HITL-017 | HITL API for approvals (REST endpoint replacing manual Kafka messages) | partial | Planned /api/v1/hitl/respond endpoint; guide written, "For Future Implementation" | hitl.md |
| HITL-018 | system.notifications.ui topic and the missing HITL UI service | partial | Rich notification topic once had no consumer; later matured into admin dashboard | hitl.md |
| HITL-019 | HITL review flow (needs_human_review → retry/resolve/spec-edit) | deployed | Content-quality flags create needs_human_review items with 3 resolution paths | hitl.md |
| HITL-020 | HITL content-approval demo agent and group | deployed | Canonical minimal HITL example: simple-content-writer-with-approval agent/group | hitl.md |
| HITL-021 | HITL kcat test harness | deployed | Six shell scripts for manually testing the HITL approval loop end-to-end | hitl.md |
| HITL-022 | Conditional approval branching | aspirational | Approved/rejected outcomes drive finalize-vs-regenerate workflow branching | hitl.md |
| ONB-001 | Domain submission tiers and mission/roadmap briefs (domain-submitter entry point) | deployed | domain-submitter entry point; three tiers up to mission/roadmap briefs | onboarding-config.md |
| ONB-002 | build_queue domain queue with direction spectrum | partial | Queue table with direction spectrum from null to fork_from; seed_build_queue | onboarding-config.md |
| ONB-003 | Three-layer onboarding/config model (mechanical / conventions / intent) | aspirational | Onboarding as three problems with different derivability/confirmation needs | onboarding-config.md |
| ONB-004 | Progressive onboarding — a ramp, never "done" | aspirational | Value from mechanical layer first; onboarding never terminates | onboarding-config.md |
| ONB-005 | Config as a maintained artifact (wizard is first pass; lifecycle is deliverable) | aspirational | Derived config gets periodic re-derivation, drift flagging, provenance | onboarding-config.md |
| ONB-006 | Inference quality scales with codebase quality — surface uncertainty | aspirational | Messy repos yield confident-but-bad convention inference; surface as questions | onboarding-config.md |
| ONB-007 | Docs-authoritative conventions for our own repo (the free drift audit) | aspirational | Docs authoritative for our repo; code disagreements recorded as free drift audit | onboarding-config.md |
| ONB-008 | Conventions agent (extract-cite-confirm, then audit) | aspirational | Extracts convention atoms, human-confirms, then audits code for drift | onboarding-config.md |
| ONB-009 | Three checking tiers + three-bucket audit output (coverage honesty) | aspirational | Deterministic/heuristic/judgement-only tiers; audit reports three numbers | onboarding-config.md |
| ONB-010 | Convention coverage IS capability reliability | aspirational | A capability is only as reliable as its manual convention's audit coverage | onboarding-config.md |
| ONB-011 | Stack-discovery agent (inspect → interpret → probe → confirm) | aspirational | Owns mechanical layer: facts, interpreted proposals, declared probe plan | onboarding-config.md |
| ONB-012 | Confirmation by reality (mechanical layer climbs the ratchet first) | aspirational | Probed commands are strongest confirmation; first capability past confirm_every | onboarding-config.md |
| ONB-013 | Sandboxed probing — the tenant-code security envelope | aspirational | Ephemeral read-only sandbox gates the first agent allowed to run tenant code | onboarding-config.md |
| ONB-014 | Intent-elicitation agent (progressive, value-returning interview) | aspirational | Captures why-chain/priority profiles via proposal-confirmation + elicitation | onboarding-config.md |
| ONB-015 | Onboarding orchestrator (dependency-graph flow; active-with-pending) | aspirational | Coordinates three layer agents by dependency; terminal state active-with-pending | onboarding-config.md |
| ONB-016 | Config-maintenance agent (drift detection as trust ratchet's signal source) | aspirational | Post-baseline drift detection across all three layers, feeds trust ratchet | onboarding-config.md |
| ONB-017 | Active-config schema (four tables, computed-on-read effective values) | aspirational | tenant_configs/mechanical_config/standards/objectives contract specification | onboarding-config.md |
| ONB-018 | Governed vocabularies and the hand-authored first constitution (prerequisites) | aspirational | Fixed concern/priority vocabularies must exist before conventions/intent agents run | onboarding-config.md |
| ONB-019 | build-briefing-agent (spec-reading briefing) | deployed | Handler answers briefing questionnaire autonomously from site_specs, no human | onboarding-config.md |
| ONB-020 | Briefing agent (early industry-brief / clarifying-question stage, pre-questionnaire) | superseded | Early two-era briefing agent generating brief JSON, later superseded | onboarding-config.md |
| ONB-021 | Intake orchestrator with two HITL gates and per-group briefing questionnaires | partial | Classify→HITL confirm→group questionnaire→HITL review brief→spawn builder | onboarding-config.md |
| ONB-022 | Per-builder briefing questionnaires | superseded | briefing_questionnaire JSONB per builder agent; fetch_agent_questionnaire action | onboarding-config.md |
| ASG-001 | Agent spawning (agents as DB records claimed by generic pods) | deployed | spawn_agent creates agent_instances row; generic chassis pod loads config | agent-spawning-and-groups.md |
| ASG-002 | Agent groups (reusable multi-agent teams) | deployed | Named versioned team definitions; spawn_group instantiates and starts workflow | agent-spawning-and-groups.md |
| ASG-003 | Agent and group discovery by capability and performance | partial | Registry service matching capability/performance; platform/discovery/ confirmed real | agent-spawning-and-groups.md |
| ASG-004 | Workflow template library, lineage and marketplace | abandoned | Reusable workflow templates with lineage, ratings, monetised marketplace idea | agent-spawning-and-groups.md |
| ASG-005 | Multi-dimensional template classification and semantic search | abandoned | Rich behavioral/performance/embedding classification for template discovery | agent-spawning-and-groups.md |
| ASG-006 | Controlled group evolution (observed mutation with rules) | partial | Curated seeds + metrics-triggered mutations; platform/evolution/ confirmed real | agent-spawning-and-groups.md |
| ASG-007 | Dynamic prompt improvement loop (Prompt Improvement Agent) | aspirational | Flag-for-improvement dispatches failing prompt to specialist, saves new version | agent-spawning-and-groups.md |
| AME-001 | entity_state_log — append-only cross-orchestration memory | abandoned | Append-only cross-orchestration memory log with accumulation patterns | agent-memory-and-evolution.md |
| AME-002 | Agent variants + snapshot versioning | abandoned | Base agents versioned/frozen; task variants reference a snapshot version | agent-memory-and-evolution.md |
| AME-003 | improvement_proposals — HITL-gated agent evolution queue | abandoned | Proposed agent/variant changes wait for human approval before applying | agent-memory-and-evolution.md |
| PERS-001 | Copywriter persona roster | abandoned | Six seeded copywriter personas with style agents, assigned by flow stage | persona-architecture.md |
| PERS-002 | Persona cognitive architecture (swappable cognitive components) | abandoned | Personas as cognitive entities with swappable memory/reasoning subsystems | persona-architecture.md |
| FLW-001 | Multi-track flows (journeys, narrative arcs, layered context) | abandoned | Site modelled as audience journeys with hierarchical context inheritance | flows-and-narrative.md |
| FLW-002 | Brand DNA invariants with bounded variance | abandoned | Site-level immutable identity layer plus allowed voice-variance ranges | flows-and-narrative.md |
| FLW-003 | Voice parameters (numeric stage-tuned voice) | superseded | Numeric voice dials per flow stage; superseded by persona selection | flows-and-narrative.md |
| ORG-001 | Organizational framework (roles, listeners, policy-as-filters) | abandoned | Whole-company modelling thought experiment: roles, listeners, policy filters | org-framework.md |
| ATN-001 | Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer) | aspirational | ltree tree_path + subtree summaries + REST/WebSocket viewer for massive trees | agent-tree-navigation.md |
| ASS-001 | Agent swarm simulation ideas (never built — hierarchical/fractal use-case brainstorm) | aspirational | Large brainstormed catalogue of 1M-agent hierarchical/fractal use cases | agent-swarm-simulations.md |
| ADR-001 | Agent definition snapshot/revert via backup table | deployed | agent_definitions_backup table; snapshot_agent/revert_agent eliminate wrong-row bug | agent-definition-registry.md |
| SCH-001 | Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration | deployed | build-pipeline-trigger seeds queue, picks one dispatchable site per tick | scheduler-and-tasks.md |
| SCH-002 | Kafka scheduler (DB-driven heartbeat service + scheduled_tasks table) | deployed | Single-replica Go service ticks scheduled_tasks every 30s, publishes triggers | scheduler-and-tasks.md |
| SCH-003 | content-feed-trigger workflow shape bug (array vs object count) | deployed | News trigger broken for weeks by array-vs-object count field shape mismatch | scheduler-and-tasks.md |
| SCH-004 | Work-item claim/retry behaviour and the claim-timeout class | deployed | Heavy builds collide with claim durations producing retried-then-complete items | scheduler-and-tasks.md |
| SCH-005 | Dispatch throughput bottleneck (Family J): one-site-per-tick, NOT-EXISTS-blocked | unknown | Dispatcher processes one site per tick and blocks entire site on any claimed item | scheduler-and-tasks.md |
| SCH-006 | pre_query SQL-worker/gate pattern (one message per tick, not fan-out; self-healing tasks) | deployed | pre_query is gate + SQL worker; scheduler never fans out per row | scheduler-and-tasks.md |
| SCH-007 | CTE-only scheduled tasks pattern ("Always Return a Row" rule) | abandoned | Zero-row pre_query silently stalls last_triggered_at/last_completed_at | scheduler-and-tasks.md |
| SCH-008 | Concurrency group starvation problem and prevention rules | abandoned | Stalled task in a shared concurrency_group can starve the whole pipeline | scheduler-and-tasks.md |
| SCH-009 | last_completed_at ownership contract and fire_message known-gap | abandoned | Agent tasks must explicitly set last_completed_at; scheduler never reads fire_message | scheduler-and-tasks.md |
| SCH-010 | Private inert pipeline statuses pattern | deployed | New pipeline uses unrecognized statuses so it's inert to existing sweeps by construction | scheduler-and-tasks.md |
| SCH-011 | Pipeline-blind dispatch surfaces (discovered platform defect) | deployed | Dispatch queries lack pipeline filters; any item on a claimable site gets dispatched | scheduler-and-tasks.md |
| SCH-012 | diagnose-dispatch-loop (automatic dispatch) | partial | Orchestrator claims awaiting_diagnosis items; shipped with scheduled trigger disabled | scheduler-and-tasks.md |
| SCH-013 | Reaper mechanisms, the work-item-claim reaper gap, and the reaper-location correction | partial | Reapers are SQL pre_query entries, not Go code; no sweep for stuck claimed items | scheduler-and-tasks.md |
| SCH-014 | Stale orchestration sweeper/reaper | deployed | Periodic DB sweep classifies expired awaited requests after goroutine death on restart | scheduler-and-tasks.md |
| SCH-015 | claimed-item-timeout & timeout chain | partial | Three timeouts must stay ordered; two-phase evidence-based/blind reset | scheduler-and-tasks.md |
| SCH-016 | thunder-training-monitor + worker (probe/classify/reconcile/decommission) | partial | Periodic orchestrator probing training boxes; terminal/decommission branch unverified live | scheduler-and-tasks.md |
| SCH-017 | Thunder unreachable-probe counter | deployed | Consecutive-unreachable-probe counter distinguishes SSH blip from truly-lost box | scheduler-and-tasks.md |
| SCH-018 | P4 off-box collection (intent_events + CollectIntentEventsAction) | partial | Key-gated HTTPS intent pull with structural idempotency via engine_event_id | scheduler-and-tasks.md |
| SCH-019 | Superseded checkpoint-JSON / events-per-1k ranking (early P4) | superseded | Early explicit checkpoint-file design dropped for structural idempotency | scheduler-and-tasks.md |
| SCH-020 | intent-collection-orchestrator + intent-collector agents | partial | Thin wrapper-orchestrator spawning intent-collector, mirrors med-export pair | scheduler-and-tasks.md |
| SCH-021 | Retention prune timer | deployed | Daily systemd timer prunes events-*.jsonl older than RETENTION_DAYS (default 90) | scheduler-and-tasks.md |
| SCH-022 | claimed-item-timeout evidence-gated completion + reset (Lever A/C) | deployed | Evidence-gated auto-completion/reset already existed; avoided duplicate watchdog build | scheduler-and-tasks.md |
| BATCH-001 | Universal LLM batch-processing architecture (queue, three-gate control, callback contract) | partial | Provider-agnostic batch queue, three-gate control, context-free callbacks | batch-processing.md |
| BATCH-002 | Work item lifecycle (blocking, unblocking, unresolved) | deployed | Items blocked three ways; unresolved mechanism suppresses rapid re-emission | batch-processing.md |
| BATCH-003 | Dispatch loop & detected→triaged→claimed state machine | deployed | Full work-item state chain from detection through claim to completion | batch-processing.md |
| BATCH-004 | Work item processing_tier (standard / batch_gpu) | unknown | Routing column for holding items until a GPU batch window opens | batch-processing.md |
