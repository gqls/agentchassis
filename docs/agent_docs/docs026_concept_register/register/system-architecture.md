# Register — system-architecture

89 concepts, consolidated from 114 raw extractions across units U01, U02, U03, U04,
U05, U07, U08, U09, U10, U12, U13, U14, U15, U16, U17a, U18, U19, U20, U21, U22,
U23, U24a, U24c, U24e, U26.
(Note: the assigned cluster input file's entire content appeared mechanically
duplicated verbatim, twice, back-to-back — a bucketing artifact, not independent
evidence. The 114 counted here are the unique raw blocks after collapsing that
duplication; two of them — QA Agent Architecture and site_work_items domain→pipeline
— arrived already pre-merged from 2 findings each by the extraction stage.)

### SYS-001 — Kafka topic model evolution (topic-per-agent → job-specific dynamic → three-pattern current model)
- **status:** deployed
- **status-evidence:** 001_development_guide(5).md (docs024, most recent) describes the current three-pattern mechanics including stable-identity naming; two earlier eras are documented as explicitly superseded by it.
- **what:** The topic model evolved through three eras. Original: static topic-per-agent (`system.agent.{type}.process/responses`, later `.requests/.responses`) — kept only as bootstrap/well-known entry points after message-stealing and routing conflicts pushed the design onward. Middle: job-specific dynamic topics (`job.{corrID}-{orchID}-{agentType}-{step}.requests/responses`) created per spawn from a "stable identity", solving the chicken-and-egg bootstrap problem and cross-job message collision. Current (docs024, live): three coexisting patterns — `system.agent.generic.requests` as the shared entry door for callers with no spawn relationship (scheduler, manual triggers; expected to evolve into a formalised entry API), per-spawn `job.<corr8>-<orch8>-<type>-<step>.requests` topics set by setupAgentTopics, and fixed `system.agent.<type>.*` topics for long-lived adapters. `agent_definitions.topics` jsonb is declarative only — the Deployment manifest actually subscribes.
- **sources:** 001_development_guide(5).md#Topics; WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics; docs001_flow_general/README.020.flow9.topicflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs/architecture/004-agent-chassis-architecture.md#kafka-topic-structure
- **relations:** wrapper-orchestrator pattern; idle timeout/topic cleanup; message header contract; agent message contract
- **verify-later:** setupAgentTopics/createTopics in spawn_actions.go; deployment manifests; actual topic list on cluster; kafka.CreateStableIdentity

### SYS-002 — Agent message contract & "agent = row in agent_definitions" orchestrator convention
- **status:** deployed
- **status-evidence:** Restated as standing rules across every era, most recently idea.uk notes (2026-06) and docs024's 001(5)/016 §6.0.
- **what:** Agents live in the database, not Go source: an agent is a row in `agent_definitions` whose `default_config.workflow` is a declarative step graph threaded by dotted-path reads from a shared data bag — "every agent is an orchestrator" is literal, and the Go codebase contains actions, not agents. The canonical three-layer message (Kafka headers, mirrored JSON headers, `config.workflow`+`input_data`) drives agents from CLI or internally; `spawn_agent`+`call_agent` are a mandatory pair (an agent cannot be called before being spawned in the current workflow); every agent always replies to the *caller's* responses_topic, never its own; HITL responses use the same shape with `in_response_to_request_id` from awaited_requests and `sender_agent_type: human`. Standing house rules travel with this: no sub-workflows in SQL (spawn sub-agents instead), workflow variable names stay in sync with action expectations, identifiers are never renamed silently, `logger.Debug` is banned (invisible in the log pipeline — use Info), and existing functions/architecture are reused before new ones are created.
- **sources:** WM/016_debugging_guide_v2_44.md#6.0; running_notes_scheme_to_components(55).md#Architecture-conventions; docs018_rerendering/009_agent_initial_message_structure.md; 001_development_guide(5).md#Agent Message Structure
- **relations:** Kafka topic model evolution; wrapper-orchestrator pattern; reply-to metadata (__work_request__); universal orchestration principle
- **verify-later:** agent_definitions (templates_db vs clients_db); default_config.workflow; SpawnAgentAction; agent-creation guidelines doc; logging config that swallows Debug

### SYS-003 — Orchestration state and collected_data as the workflow data bag (current mechanics)
- **status:** deployed
- **status-evidence:** 001(5)/016 §6.0 describe live mechanics.
- **what:** Each orchestration is an `orchestration_states` row (workflow_plan, collected_data, current_step, status). Steps communicate solely via dotted paths into collected_data; agents themselves are DB rows, not Go types.
- **sources:** 001_development_guide(5).md#Orchestration State; 016 §6.0 "What an agent actually is"
- **relations:** loop mechanisms; workflow result contract; CollectedData pathologies
- **verify-later:** orchestration_states schema; coordinator.go continueExecution

### SYS-004 — Stale orchestration sweeper (DB sweep replacing lossy timeout goroutines)
- **status:** partial
- **status-evidence:** docs024's 001(5) documents it live ("the #1 cause of pipeline stalls") with a concrete A/B/C classification; an earlier docs018 design doc described the same mechanism without deployment confirmation — superseded by the later, deployed description.
- **stage2-verified (2026-07-14):** deployed → partial — no file named sweeper.go / no 'StaleOrchestrationSweeper' hit in .go; only tangential hits (thunder_decommission_dispatch.go, hitl_actions.go) — sweeper mechanism not confirmed present in code, only in docs prose
- **what:** Timeout goroutines die with the pod, leaving AWAITING_RESPONSES orchestrations stuck. A periodic ~60s DB sweep on every chassis pod (`FOR UPDATE SKIP LOCKED`, ~30s grace, LIMIT 20) classifies expired awaited_requests: child COMPLETED → synthesize a completion response from the child's final_result; child FAILED → forward the failure; none/running → retry up to retry_version 3, then fail the parent orchestration. Handles cascading stalls oldest-first and dead job topics by directly advancing parent state.
- **sources:** 001_development_guide(5).md#Stale Orchestration Sweeper; docs018_rerendering/009_stale_orchestration_sweeper_design.md
- **relations:** spawn-handler hang (timeout_at not enforced); claimed-item-timeout; awaited_requests registry
- **verify-later:** sweeper implementation; stale-orchestration-reaper scheduled task; platform/orchestration/sweeper.go existence

### SYS-005 — Work-item relay spine / dispatch-loop pattern (batons, handler_agent, the 30s pump)
- **status:** deployed
- **status-evidence:** HANDOFF_builder_thread §2 "Builder route CLOSED: spine decided = the work-item relay"; README_flows describes the live pump; an earlier docs024 description of "site-work-orchestrator ordering" covers the same dispatch shape from an earlier vantage point.
- **what:** Builds move as a relay: the baton is a `site_work_items` row naming a `handler_agent`; `build-pipeline-trigger` (every 30s, behind a pre-query gate) seeds `build_queue` and picks one dispatchable site; `build-dispatch-loop` claims items atomically and spawns a dynamic handler per item — dispatch is spawn→call per item with raw identifiers, never pre-spawning static chains or passing work-item awareness into handlers (handlers stay self-contained, CLI-callable). One hop = one baton, one agent, one site_specs entry, one new baton. Around it sits an immune system: evidence-based claimed-item timeout, feasibility-recheck, reapers, archiver, cleanup.
- **sources:** HANDOFF_builder_thread.md#1,#2; README_flows.md; 002(4)#The orchestrator, #Dispatch pattern, #Resolved Decisions
- **relations:** hop-insertion pattern; needs_diagnosis intake; work-site-orchestrator vs build-site-planner; claim_work_item action; loop mechanisms
- **verify-later:** scheduled_tasks rows; improvement-sweep flag state; site-work-orchestrator definition currency vs build-dispatch-loop

### SYS-006 — Entity data model (state-based lifecycle, news triggers, client-side real-time)
- **status:** aspirational
- **status-evidence:** 002(4): tables site_entities/site_entity_relationships "exist", entity_sources/entity_sync_log "planned"; Phase 3 item.
- **what:** Structured data that generates pages (events, performers, venues, ticket tiers) with a state-based lifecycle (announced→on_sale→…→historical), setup mode (work items) + discovery mode (scheduled sync), significant state changes triggering news via entity_sources.news_triggers; real-time data (prices, availability) served client-side from a data API, never through the work queue.
- **sources:** 002(4)#Entity Data Agent Family; #Site Type Stress Tests (events/boxing)
- **relations:** news feed pipeline; site API router
- **verify-later:** site_entities usage; any entity-data-agent definition

### SYS-007 — Maintenance profile per-site configuration
- **status:** deployed
- **status-evidence:** 002(4) JSON shape; audit config consumed by the improvement loop.
- **what:** `sites.settings.maintenance_profile` controls per-domain cadence (content/links/seo/compliance/content_feed/entity), budgets (llm_calls_per_cycle, max_auto_fixes_per_cycle), build tier, and audit group enablement; audit_pass_count also lives here.
- **sources:** 002(4)#Per-Site Configuration; 002d#Site-Type-Specific Audit Configurations
- **relations:** audit pass cap; growth budget (separate: site_specs growth_config)
- **verify-later:** sites.settings shape in production

### SYS-008 — Idle timeout for spawned agents + topic cleanup strategy
- **status:** deployed
- **status-evidence:** docs024's 002(4) describes the mechanism, config column, sync.Once shutdown safety, tuning SQL; the fleet-wide 075 migration backfilled 180s with rationale.
- **what:** `agent_definitions.idle_timeout_seconds` → env var → an idle-monitor goroutine exits the pod after inactivity (0 = forever, for Deployment-style agents); the timer resets on every message so a multi-step workflow stays alive as long as responses keep arriving. Topics: EPHEMERAL_TOPICS per-spawn today; agents never clean up their own topics — a conservative 10-min CronJob deletes topics with no matching pod, Kafka's 7-day retention is the backstop; a future shared-topics design (pre-created per-type topics, header routing, static group membership) would make cleanup a no-op.
- **sources:** 002(4)#Idle Timeout, #Shared Topic Strategy, #Topic Cleanup Design; 075_various_timeout_column.sql
- **relations:** pod accumulation debugging; wrapper-orchestrator pattern; Kafka topic model evolution
- **verify-later:** idle_timeout_seconds values; topic-cleanup CronJob; chassis idle-timer implementation

### SYS-009 — business-intel shared-pod pattern (multi-type agents on one static pod)
- **status:** deployed
- **status-evidence:** 017 architecture section; ai_service placement rule.
- **what:** Multiple agent definitions share one static pod via message routing (config.agent_type → selectWorkflow/FindBestGroup); consequence: `ai_service` must live in STEP config, not agent_config top-level, because agent_config comes from the pod's own type. Workflows are minimal action→complete with logic in Go. Single-replica contention is accepted for batch work.
- **sources:** 017#Deployment, #ai_service on Shared Pods
- **relations:** wrapper-orchestrator pattern (contrast); med pipeline (same pod)
- **verify-later:** business-intel deployment manifest

### SYS-010 — CollectedData: single-channel orchestration working memory and its pathologies
- **status:** deployed
- **status-evidence:** "Analysis only. No code changes proposed yet" (2026-05-11); duplication called "structural", observed in every log.
- **what:** CollectedData (`orchestration_states.collected_data` JSONB) is the single channel for step outputs, routing metadata, loop variables and parent-reply context — "the most overloaded data structure in the system." Documented pathologies: recursive `__raw_message__` nesting (write amplification ×15 optimistic-lock retries), dual storage at step_name AND output_field, InitialRequestData/__raw_message__ overlap, six conflated data categories in one flat namespace, loop iteration data stored 3-4×, CleanDataMap stripping legitimately-named response fields. Recommendations R1–R6 (strip system keys from __raw_message__, pick one storage key, namespacing, loop GC, delta writes) were proposed and left untriaged.
- **sources:** FOCUS_collected_data_analysis.md (whole)
- **relations:** flat-namespace collision risk; compensating mechanisms; consumer-group race; collected_data OOM growth
- **verify-later:** BuildCollectedData / storeActionResult in coordinator.go; whether R1 ever landed

### SYS-011 — Flat-namespace collision risk and the compensating-mechanism accretion
- **status:** deployed
- **status-evidence:** dev-guide-documented incident: "section-editor declared content_data as optional and the nested-source loop silently lifted site_record.content_data … and overwrote a hero section" (2026-05-11).
- **what:** Because caller inputs, step results and site context share one flat map, actions can silently pick up `site_record.site_id`/`content_data` instead of caller-supplied fields. The framework compensates with UnwrapDeep, FindByPath prefix fallbacks, extractReplyToMetadata 3-tier priority, output_mapping — accreting workarounds faster than it consolidates. New code should use collision-free names (`target_site_id` convention); existing code is left alone.
- **sources:** FOCUS_collected_data_analysis.md#4.4, #5; ASSESSMENT_imagery_phase_0_1…md#Caveat-1
- **relations:** CollectedData pathologies; ExtractActionInputs conventions
- **verify-later:** ExtractActionInputs nested-source loop behaviour for undeclared fields

### SYS-012 — Response-topic consumer group race (per-pod groups fan out every response)
- **status:** partial
- **status-evidence:** "Discovery, not yet remediated" (2026-05-10); ~85 consumer groups on system.agent.generic.responses, only 3 live; two pods ran ProcessResponse on the same message 215ms apart.
- **stage2-verified (2026-07-14):** unknown → partial — platform/agentbase/agent.go:341-343 generates consumerGroup+AgentID[0:8] passed into client.go:39/48 which appends '-responses', confirming the per-pod-unique responses-group design described; platform/orchestration/coordinator.go has ExecuteWithOptimisticLocking/CAS retry (maxOptimisticLockRetries=15) which mitigat...
- **what:** The requests topic uses a shared stable consumer group but each chassis pod joins the responses topic under its own per-pod UUID group, so every response is delivered to every pod; each independently advances orchestration state, and the loser of the version race can flip a step to FAILED (observed on call_logo_gen). Mostly silent (idempotent writes) but structurally wrong — the system relies on shared-pool semantics it doesn't have. Open questions: intended model, per-spawn job.* topic groups, CAS hardening in ProcessResponse, 82 stale groups cleanup.
- **sources:** ANALYSIS_chassis_response_consumer_group_race.md (whole)
- **relations:** dispatcher stall Bug 1; duplicate collected_data keys; Kafka empty partition assignment
- **verify-later:** AgentClient constructor wiring (consumerGroup argument); ProcessResponse CAS behaviour in coordinator.go

### SYS-013 — Kafka empty partition assignment on simultaneous pod join
- **status:** unknown
- **status-evidence:** "five agent-chassis pods were members of generic-requests-group but all showed #PARTITIONS: 0 … Fix applied: Delete one pod to force rebalance" (2026-04-20).
- **what:** After a deploy where all pods join within the same second, the consumer group can go Stable with the partition unassigned — zero consumption while offsets pile up and work items sit triaged. Workaround: kill a pod. Watch item on every deploy: at least one member must show #PARTITIONS: 1.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#1
- **relations:** consumer-group race; dispatcher reliability
- **verify-later:** whether staggered restarts or a permanent fix was adopted

### SYS-014 — Observability gaps: owner_agent_type "generic" and orchestration_name
- **status:** unknown
- **status-evidence:** "orchestration_states rows where the generic agent routed to a different workflow still show owner_agent_type = 'generic'" (2026-04-20, P3).
- **what:** When the generic chassis routes a message to another agent's workflow (FindBestGroup), the orchestration is filed under owner_agent_type='generic' and orchestration_name doesn't carry the scheduler's sched-<task> name — searches by agent type or task name find nothing, which caused a "trigger never runs" misdiagnosis. Fix shape: selectWorkflow sets owner_agent_type to the resolved type.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7, items 7-8
- **relations:** content-feed-trigger bug; execution_path not populated
- **verify-later:** selectWorkflow in processor.go

### SYS-015 — Four overlapping chrome default stores and the update_site_defaults linkage
- **status:** partial
- **status-evidence:** Intended chain documented in the Site Component Linkage Contract; the fix explicitly chose function-lookup as the norm rather than reviving the chain — "the linkage deliberately left unrepaired."
- **what:** Header/footer defaults coexist in four stores: `style_collections.header/footer_component_id` (the operative read, dead-NULL), `site_components` slots (copy target + pre-render cache — idea.uk's were pinned to inactive components), `sites.default_components` JSONB (UpdateSiteDefaultsAction's target — a tracking copy nothing reads on the render path), and `layouts.default_*_component_id` (FK, all NULL, nothing copies it onward). The intended chain — style_collections as source of truth, `update_site_defaults` copying into site_components — never runs in composition; the fix chose function-lookup as the norm instead of reviving the chain. Populating style_collections at install remains a possible per-site-variety feature.
- **sources:** running_notes_scheme_to_components(55).md#Sg #Sh; SPEC_scheme_to_components.md#W4; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3b)
- **relations:** chrome selection path; site_components repoint
- **verify-later:** v3_site_actions.go UpdateSiteDefaultsAction; whether the misleading install comment was deleted; any later population of style_collections chrome ids

### SYS-016 — Coordinator result-extraction contract (resolveResultSpec) and the silent-stub bug family
- **status:** deployed
- **status-evidence:** Fix shipped 2026-06-18 (result_spec.go + coordinator.go), field-validated 2026-06-19 (idea.uk index built+deployed), re-confirmed healthy in prod 2026-06-21 (stub_rows=0); "STATUS 2026-07-04: DEPLOYED (image v1.0.1092), rename migration APPLIED + VERIFIED."
- **what:** The chassis coordinator's `extractWorkflowResult` honoured only plural `output_fields` since commit 06a8c6e (2026-01-14). A workflow `complete` step declaring singular `output_field` was silently ignored, falling back to a working-state dump; when a big multi-section page's dump cleared the 900k `MaxResultSizeBytes` cap, `extractMinimalResult` returned a `status:"completed"` **stub** — silent false completion (gamesdesign) or, where a claimed-item evidence gate refuses 0-component pages, honest claim-timeout failure (idea.uk's empty index). Size was the trigger; the singular key was necessary-but-not-sufficient (a bucket audit found only 4 singular-key agents at risk, and only the writer actually broke). Fix: a centralised result contract in `result_spec.go` — singular→FLATTEN (named field's contents become the response body), plural→FIELDS (unchanged), `output`→MAPPING (previously silently dumped), none→dump; deprecated keys (result_from, multiple_output_fields, result_mapping) alias in with a deprecation Warn; the resolution table was later lifted verbatim into `datahelpers/result_contract.go` (ResolveResultSpec + ApplyResultSpec) as Option A, with agents migrated to preferred key spellings — one source of truth for coordinator and action paths.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Empty index diagnosis + fix-direction 1 + retest); NOTES_gamesdesign_silent_norebuild(44).md#root-cause; docs/016b_debugging_guide_merged(3).md#open-threads; docs019/RUNBOOK_gamesdesign_index_rebuild.md#what-this-exercises; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups; NOTES_running_synthesis_v2(36).md 2026-06-17
- **relations:** oversize-result fail-loud hardening; parent/child result key = step-name convention; claimed-item evidence gate; MaxResultSizeBytes guardrail (do not raise); workflow default_config location convention
- **verify-later:** platform/orchestration/result_spec.go + coordinator.go; datahelpers/result_contract.go + 7-case test; the mode=flatten log on a writer run; NNN_rename_complete_keys_preferred.sql applied state

### SYS-017 — Hosting split: static-serverless front + small always-on backend
- **status:** deployed
- **status-evidence:** Architecture doc §3 and the live topology (page embedded in the binary on the VM; B2 for everything chassis-built).
- **what:** The hosting taxonomy idea.uk established for the platform: pure-static content sites are serverless on B2; anything running a minutes-long multi-LLM job with a payment webhook cannot be serverless or edge-shaped — it needs a small always-on service with a stable inbound address. The classifier's `build_approach: hybrid` / `hosting_trajectory: needs_server` fields are the framework's slot for this distinction (not yet confirmed as populated). This is the hinge Layer 5 eventually automates, and why the engine can never be a forked client-side tool component.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#3; idea.uk/CONSOLIDATION_where_it_all_fits.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md
- **relations:** Layer-5 wrapper; VM cutover; idea.uk topology exception
- **verify-later:** where build_approach/hosting_trajectory actually live

### SYS-018 — Oversize-result delivery: fail-loud hardening + size guards ("responses are summaries")
- **status:** deployed
- **status-evidence:** "DONE (implemented 2026-06-18 … deployed)"; code_retrieval_route(21) "MEASURED (2026-07-03): cd_bytes=1,270,781 — the completion response carried the full collected_data alongside the 6KB result"; FIX applied 2026-07-03.
- **what:** `extractWorkflowResultWithSizeLimit` no longer truncates strings or emits a `status:"completed"` stub (`extractMinimalResult` was deleted); on oversize it returns `oversizeResultError` with a per-field size breakdown naming the largest field, and `notifyParentOfSuccess` converts to a loud `CHILD_ORCHESTRATION_FAILED` + agent_error_log entry. Root mechanism: `result_from` was a key `CompleteWorkflowAction` never read, so its fallback shipped the ENTIRE collected_data (1.27MB > Kafka's ~1MB cap → "Message Size Too Large"). Fixes layered in: per-agent output_fields selection, a response SIZE GUARD (max_response_bytes, truncatedResponseStub naming where the real result lives), and removal of the earlier silent "exceeded size limit" completed-stub. Doctrine: responses are summaries; heavy artifacts live in the DB, retrievable by correlation_id; raising the broker cap is inversion, last resort.
- **sources:** NOTES_gamesdesign_silent_norebuild(44).md#Follow-ups; docs019/RUNBOOK_code_retrieval_route(21).md#7D,#follow-ups; docs019/RUNBOOK_gamesdesign_index_rebuild.md#6; NOTES_running_synthesis_v4(39).md 2026-07-03 §7E
- **relations:** Coordinator result-extraction contract; bundle size doctrine; MaxResultSizeBytes guardrail
- **verify-later:** coordinator.go oversizeResultError; workflow_actions.go size guard + truncatedResponseStub; agent_error_log for delivery-cap entries

### SYS-019 — sites.status is an informational lifecycle label (no build-time consumer filters)
- **status:** deployed
- **status-evidence:** "No on-disk code filters sites on status — it is an informational lifecycle label; build dispatch keys on site_work_items" (from v3_site_actions.go:323–395).
- **what:** UpdateSiteStatusAction validates status ∈ {draft, building, review, published, deployed, archived, error} (stamping last_deployed_at when status=deployed); 'active' and 'system' are legacy out-of-vocabulary values on old rows. Nothing filters sites by status at build time — an assumption (`WHERE s.status='active'`) borrowed from an old handoff silently wrecked a blast-radius count, hence the standing rule: never filter on status='active'.
- **sources:** RUNBOOK_scheme_to_components(18).md §sites.status RESOLVED; running_notes(22).md Sr, Ss
- **relations:** work-item dispatch (the real gate); sites.status vocabulary and the blast-radius filter trap (database-and-infrastructure angle, DBI-018)
- **verify-later:** UpdateSiteStatusAction vocabulary; legacy-status rows

### SYS-020 — Aspiration: agent-creation and inter-agent message logging workstream
- **status:** aspirational
- **status-evidence:** "Note on a separate, out-of-scope item… kept out of these docs to preserve separation of concerns. Can be specced separately." No later mention.
- **what:** A stated desire to closely log/track agent creation and inter-agent messages (headers + body) as a distinct workstream from travelling docs — different responsibility and data. Never designed or built within any extraction unit's horizon.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#out-of-scope-note
- **relations:** travelling docs (deliberately separated); message envelope standards
- **verify-later:** whether any message-logging workstream exists elsewhere in docs/

### SYS-021 — pages.sections is the build-read field (load_page_sections_from_spec fallback chain)
- **status:** deployed
- **status-evidence:** "Decisive code facts established… load_page_sections_from_spec reads site_specs.site_plan (absent for this site) → falls back to pages.sections. pages.sections is the build-read field; site_plan_sections is NOT on the build path."
- **what:** Page build resolves its section list via `site_specs` aspect `site_plan` (syncing into pages.sections when present) → `pages.sections` fallback → empty. `site_plan_sections` is relational plan hygiene only. `plan_sections` computes `ready_count` from resolvable section names; an empty list gives an early return of ready_count=0. Consequence: manual repairs must write pages.sections (the load-bearing statement), and a re-sync re-derives pages.sections from the in-memory plan, so manual fixes survive only until the next full re-plan.
- **sources:** running_notes_15(10)#part-1–2; CONTEXT_PACK_adoption_skinner_box.md; HANDOFF_2026-06-06
- **relations:** sectionless durability; upsertPage EXCLUDED.sections; write_site_plan writes both tables
- **verify-later:** load_page_sections_from_spec_action.go source order; plan_sections_action.go ready computation

### SYS-022 — Chassis config location bugs (max_tokens shadowing, temperature single-read, error_step dead field)
- **status:** partial
- **status-evidence:** "max_tokens workaround applied for site-adoption-agent; error_step fix deployed via Path A; temperature TODO outstanding" (2026-05-18).
- **what:** A family of author-intuition-vs-chassis-reader mismatches: (1) `ExecuteLLMPromptAction` resolves the whole ai_service object once — a top-level ai_service shadows step-level config even when missing a field, so site-adoption-agent's analyze_site fell to the hardcoded 2048-token Anthropic fallback and truncated 8 of 20 pages (workaround: top-level max_tokens=16000); (2) temperature is read only from the very top of default_config — 6 step-level settings are dead and `llm_call_log.temperature` is NULL for every call; (3) `error_step` at step level was silently dropped because the Step struct had no field — 62 dead graceful-degradation paths across 18 agents; a Path A fix (ErrorStep struct field + config fallback) shipped, turning e.g. a firecrawl CSS timeout from fatal adoption failure into fallback-to-analyze_site with a partial fingerprint.
- **sources:** FOCUS_chassis_config_location_bugs.md; old2/FOCUS_step_level_llm_config_ignored(3).md
- **relations:** llm-quality-testing per-step model swap breaks on shadowed agents; fetch_primary_css hard-fail trade-off
- **verify-later:** execute_llm_prompt_action.go lines 110–145/209/213-218; contracts.go Step.ErrorStep; llm_call_log temperature population

### SYS-023 — work-site-orchestrator (monolith) vs build-site-planner (thin planner)
- **status:** deployed
- **status-evidence:** README maps every monolith inline step to its new-path equivalent; "the thin-planner model is the one that matches your stated philosophy"; design/imagery gaps closed per the mapping.
- **what:** Two architectures, not two versions: the old monolith did plan/write/sync/nav/design/render/deploy inline in one agent; build-site-planner is `read_specs → plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan (→ emit_design → emit_imagery) → complete`, delegating everything else via work items to handler agents. The open question: no clear terminal step advances `sites.build_status` to deployed/active (sites may sit `pending` indefinitely).
- **sources:** README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** work-item relay spine; emit_design/emit_imagery; pageflow-builder (another still-active plan-less caller)
- **verify-later:** what sets sites.build_status terminal on the build path; whether work-site-orchestrator is still invocable

### SYS-024 — Snapshot-shadowing agent-definition loader defect
- **status:** deployed
- **status-evidence:** "Snapshot-shadowing hypothesis confirmed via SQL… Two loaders patched… Builds deployed as v1.0.1006"; "Snapshot audit closed clean" (2026-05-12).
- **what:** `snapshot_agent()` inserts snapshots at version+1000, so any loader using `ORDER BY version DESC LIMIT 1` without filtering `is_snapshot`/`is_active` reads the snapshot instead of the active row — every snapshot silently shadowed its agent until the loader was fixed. `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` were patched; other loaders were already correct. Structural residue: snapshot retention policy and a single AgentDefinitionRepository remain open.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-1; STATUS_imagery_2026-05-12.md#Loader-snapshot-defect
- **relations:** model-infrastructure snapshot/rollback; consumer-group race (parked together)
- **verify-later:** grep `FROM agent_definitions` across Go for is_snapshot filters; snapshot row counts

### SYS-025 — Quality Assurance Agent Architecture — folded into system-architecture, not abandoned
*(merged from 2 independent findings, upstream)*
- **status:** superseded (as a standalone numbered doc; the architecture itself is deployed/partial)
- **status-evidence:** The standalone `002d_quality_assurance_architecture.md` and its later revision both appear verbatim as a "002d — Quality Assurance Agent Architecture" section inside live `002_system_architecture(4).md` (from line 897), continuing the main doc's Resolved-Decisions numbering and extending it with a new "Layer 0: Pre-Generation Data Triage (plan_sections)" section and two further resolved decisions.
- **what:** A three-layer QA model: Layer 1 structural/algorithmic checks (free, no LLM), Layer 2 LLM-assisted design/content audit (grouped agents sharing context, one LLM call per group), Layer 3 LLM-required strategic review (dream-spec gap analysis); plus a later-added Layer 0 pre-generation data triage (`plan_sections`). Includes the "promotion pattern" (a check starts as a `query_database` action step and is promoted to a spawned sub-agent only once it needs multi-step workflows or external calls) and the rule that audit agents "enforce, not override" the classifier/planner's stated intent. Never genuinely dropped — consolidated into the numbered `002_system_architecture` doc and then actively extended.
- **sources:** old/older1/002d_quality_assurance_architecture.md; archive_april_26/002de_quality_assurance_architecture_v2.md; docs024_key_docs_latest/002_system_architecture(4).md#"002d"
- **relations:** design agent responsibility split; improvement-loop; site-spec-and-classifier; triage drain loop
- **verify-later:** confirm design-audit-agent/visual-design-auditor/content-quality-auditor/site-review-agent still implement the three-layer split; confirm plan_sections and needs_human_review status are implemented as described

### SYS-026 — site_work_items work-routing column renamed domain → pipeline
*(merged from 2 independent findings, upstream)*
- **status:** superseded
- **status-evidence:** Live bug-log entry #18 in `001_development_guide(5).md`: "The `domain` column on site_work_items was renamed to `pipeline` in a migration."
- **what:** `site_work_items.domain` was an internal work-routing namespace ("build"/"maintenance"/"marketing") that collided confusingly with the website's actual domain (e.g. "gaswholesalers.com"), causing real bugs (a dispatch-loop filter mismatch, a CSS-generation item never dispatching because it was written `domain:"design"` instead of `domain:"build"`). Rather than rely on documentation warnings, the column was renamed to `pipeline` at the schema level, eliminating the ambiguity outright.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md; docs024_key_docs_latest/001_development_guide(5).md#18; old_design_and_styling/016_debugging_guide_v2.md
- **relations:** dispatch-loop input_mapping; site_work_items table
- **verify-later:** confirm site_work_items.pipeline column exists and no code still reads/writes domain for this purpose

### SYS-027 — Dispatch-loop input_mapping path mismatch (spec-nested vs flat)
- **status:** unknown
- **status-evidence:** Documented as "most common systematic failure" with three named affected agents (tool-improver, tool-auditor, rerender-pages); not confirmed whether a fix was adopted.
- **what:** `build-dispatch-loop` maps a work item's `spec` JSONB as nested (`input_data.spec.component_id`), but handlers read flat (`input_data.component_id`), producing path-resolution errors. Preferred fix: flatten in the dispatch loop's `input_mapping`, following the existing `page_name?`/`reviewed_brief?` pattern.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; ExtractActionInputs cross-link
- **verify-later:** current build-dispatch-loop input_mapping config

### SYS-028 — Asset self-resolving storage URI (dispatch loop → asset-deployer)
- **status:** superseded (as a standalone archived doc — the mechanism itself may still be live)
- **status-evidence:** The archive documents specific Go additions and a worked dispatch trace; none of this function-level detail survives in the live `002_system_architecture(4).md`, which keeps only the abstract principle.
- **what:** When the dispatch loop's discovery-written work items carry presigned HTTPS asset URLs but `deploy_image_asset` needs `s3://` URIs, the fix was to have `asset-deployer` resolve its own storage URI from `asset_id` via a DB lookup rather than have the orchestrator pre-resolve it — keeping handler self-containment.
- **sources:** archive_april_26/004_site_work_orchestrator.md; docs024_key_docs_latest/002_system_architecture(4).md#"Dispatch Loop"
- **relations:** work-item relay spine; asset-deployer agent; handler self-containment principle
- **verify-later:** grep `PresignedURLToS3URI\|resolveStorageURIFromAsset` in platform/ to confirm these functions still exist

### SYS-029 — Self-spawning flat dispatch-loop (pre-scheduler design, superseded)
- **status:** superseded
- **status-evidence:** Archive (dated 2026-02-24): "No sub_workflows — they've been problematic," "No loops in dispatch — one item per invocation, self-spawns for next item"; the eventual system uses a genuine `action:"loop"` construct driven by a scheduled 30s/120s kafka-scheduler tick.
- **what:** An early design decision to avoid the framework's loop/sub_workflow mechanism entirely, having `build-dispatch-loop` process exactly one item then spawn a fresh copy of itself. Abandoned in favour of the scheduler-driven periodic trigger combined with the fully-developed in-workflow loop mechanism.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Key Design Decisions"; docs024_key_docs_latest/010_scheduler_and_tasks.md; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C"
- **relations:** loop mechanisms; scheduler-and-tasks; build-dispatch-loop agent
- **verify-later:** confirm build-dispatch-loop's current workflow uses the loop action, not self-spawning

### SYS-030 — claim_work_item atomic claim action + load_work_items first_item patch
- **status:** deployed
- **status-evidence:** Archive lists these as "created, not yet committed" (Feb 2026); the live loop-mechanisms appendix shows claim_work_item as a fully standard, already-existing production action.
- **what:** `claim_work_item` performs an atomic `UPDATE ... WHERE status IN ('triaged','approved') RETURNING id` so concurrent dispatch loops can't double-process the same item. The companion `load_work_items` patch added a `first_item` convenience field since the framework's path resolver doesn't support array indexing.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Completed Artifacts"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C"
- **relations:** work-item relay spine; loop mechanisms; scheduler-and-tasks
- **verify-later:** none needed — graduated cleanly from draft to shipped mechanism

### SYS-031 — collected_data growth causing OOM-kills and lost work
- **status:** partial
- **status-evidence:** "Status: Diagnosed, not yet fixed."
- **what:** component-quality-auditor orchestrations were observed holding 18MB `collected_data`, causing OOM-kills at the 512Mi pod memory limit. OOM mid-publish causes phantom-completed orchestrations and cascading parent-orchestration timeouts. Suspected bloat contributors: `__raw_message__` duplicating input_data, unbounded `processing_history`, large uncleared LLM responses. A separate consumer-group bug (per-pod `a.AgentID` with FirstOffset, causing backlog replay on restart) was flagged in the same investigation, deliberately held back.
- **sources:** FOCUS_platform_reliability_oom_and_reapers.md#Part-1; TODO_orchestration_memory_bloat.md
- **relations:** Reaper mechanisms and gap; CollectedData pathologies; consumer group bug
- **verify-later:** orchestration_states.collected_data for component-quality-auditor; agent.go a.AgentID consumer group

### SYS-032 — Page content-creation build pipeline trace (page-build-handler workflow)
- **status:** deployed
- **status-evidence:** Each hop verified directly against chassis source, 2026-05-20.
- **what:** Documents, hop by hop, how a `pages` row's bare `sections` list becomes populated, deployed HTML with linked `page_components`: load_page_record → plan_sections (triages by schema source) → content writer's `extractResponseContent` (flat string, dead end for structured fields) → RenderComponentAction → CompilePageSectionsAction → SavePageSectionsAction (structured-metadata path preferred, HTML-regex fallback, orphans a page_component when metadata isn't fully recovered).
- **sources:** FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow; old/page_content_creation_flow.md
- **relations:** Stale site_plan gap; extractResponseContent flat-string hypothesis; isolated build test methodology
- **verify-later:** LoadPageRecordAction, PlanSectionsAction, RenderComponentAction, SavePageSectionsAction

### SYS-033 — extractResponseContent flat-string hypothesis for FAQ root cause (superseded)
- **status:** superseded
- **status-evidence:** An intermediate hypothesis (old/015) was superseded once the isolated `faq-test` build proved the writer populates a `questions` array correctly given a clean plan.
- **what:** An intermediate working hypothesis that the content writer itself could never populate a structured field like FAQ's `questions` array. Superseded once the isolated build test proved the writer works correctly standalone; the real cause was duplicate content surfaces.
- **sources:** old/015_content_data_persisted.md; FOCUS_faq_empty_items_and_page_content.md#The-test-that-settled-the-cause
- **relations:** FAQ duplicate content-surface bug; page content-creation build pipeline trace
- **verify-later:** grep/inspect `questions`; `faq-test`

### SYS-034 — Site-chrome rendering gap (missing nav/header/footer in relay build path)
- **status:** partial
- **status-evidence:** Measured baseline shows "nav: 0" on all four rendered pages of dartsonline.com; hypothesis not yet confirmed against site_components rows.
- **what:** A suspected structural gap discovered by direct measurement of a live site (dartsonline.com, zero `<nav>` elements on every page, single stylesheet link pointing at a CSS file whose `needs_design` item was still triaged/undelivered): the newer work-item-relay build path may never invoke the chrome-rendering step that the older `pageflow-builder` path has, while `build-site-planner`'s `populate_nav_tables` only writes nav data, not rendered chrome.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE, #THE THREE-WAY SPLIT
- **relations:** Design/composition work-item emission gap; three-way split quality-gap diagnostic method
- **verify-later:** `SELECT * FROM site_components WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'`

### SYS-035 — Generic orchestrate envelope as the universal manual trigger
- **status:** deployed
- **status-evidence:** RUNBOOK(31) §6D/§6F full kcat scripts; reused verbatim for code-indexer and diagnose.
- **what:** One trigger shape for hand-running any agent: kcat-produce to `system.agent.generic.requests` with correlation/orchestration/message/request ids, `action=orchestrate`, `config.agent_type=<entry agent>`, and task-specific `input_data`. Known wrinkles: `site_id` intermittently arrives empty (reproducibility bug, parked), and runtime selectors (`site_id` vs `runtime_site` domain) drive different evidence filters in load_runtime.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D,#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7E
- **relations:** diagnose-orchestrator wrapper; needs_diagnosis intake; correlation-id discipline
- **verify-later:** drafts/084_TRIGGER_diagnose_v1.sh; 080c/082 trigger scripts; site_id-empty envelope bug

### SYS-036 — Parent/child result key = step-name convention
- **status:** deployed
- **status-evidence:** "RESOLVED 2026-07-03: jsonb_object_keys shows the child response stored under the STEP NAME call_diagnoser … 'diagnose-agent_result' never existed."
- **what:** Engine behaviour, confirmed against real orchestrator rows: when a call step has no output_field, the child's response is stored in the parent's collected_data under the STEP NAME — imagined synthetic keys like `<agent>_result` never exist. This is read by the parent under the CALLING STEP's own name/output_field, not a role- or agent-type-based key. Two orchestrators carried imagined keys and were fixed by pointing complete steps at the real step-name keys — a recurring class alongside dotted-config lookups.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 1); NOTES_running_synthesis_v4(39).md 2026-07-03; NOTES_running_synthesis_v2(36).md
- **relations:** Coordinator result-extraction contract; workflow default_config location convention
- **verify-later:** collected_data ? 'call_diagnoser' on a post-migration diagnose run

### SYS-037 — Workflow default_config location convention
- **status:** deployed
- **status-evidence:** "the query result overturned my main assumption... task_workflow / orchestrator_workflow / orchestration_workflow are ALL NULL on every working orchestrator... The workflow lives in default_config" (2026-06-17).
- **what:** A load-bearing, empirically-corrected fact about the chassis schema: an agent's actual workflow (start_step/steps graph) lives in `agent_definitions.default_config`, never in the three separately-named `*_workflow` columns that also exist on the table — discovered only by querying real working rows after an entire migration draft was written on the wrong assumption inferred from the dev-guide's prose.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17
- **relations:** Real-rows-beat-prose discipline; Coordinator result-extraction contract
- **verify-later:** n/a — confirmed against live rows

### SYS-038 — Autonomous Build-and-Operate — the trust-not-capability thesis
- **status:** aspirational
- **status-evidence:** MASTER(4) header "Status: synthesis spine, built over several turns. Sections 6–8 are deliberately stubs to detail next."
- **what:** Umbrella vision: everything already built is apparatus for a single reliability problem — bound LLM uncertainty at each step enough to progressively remove the human. The whole plan targets building/operating a real site (vonc.com) autonomously by composing companion FOCUS mechanisms into one toolkit across the full lifecycle.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#1, #9
- **relations:** umbrella over the salience/standards/mediator/context FOCUS concepts; vonc
- **verify-later:** none (discussion doc)

### SYS-039 — Build-vs-operate asymmetry
- **status:** aspirational
- **status-evidence:** MASTER(4) §4 "Build … isolatable … Competition is safe. Operate … live, stateful … Competition is risky."
- **what:** Build work (actions, workflows, components, agent defs) is branchable/sandboxable so competition is safe and the ratchet moves fast; operate work (provisioning, scaling, incident response) is live and stateful so it leans on known-good + canary + rollback + tighter HITL.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#4, #7.6
- **relations:** lifecycle map (Tier A/B/C); reliability cascade
- **verify-later:** none

### SYS-040 — Lifecycle map by verifiability + containment (Tier A/B/C)
- **status:** aspirational
- **status-evidence:** MASTER(4) §6.1–6.2 tables; "Ceiling is separate from current maturity."
- **what:** Every capability's autonomy ceiling is set by two independently-failing factors — verifiability (can we tell against ground truth it's correct) and containment (how bad/reversible if wrong). Tier A (Go actions, SQL, component-structural, observability, rollback) reaches autonomy; Tier C (security, sharding, replication, live remediation, meta-loop) stays gated regardless of agent capability.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.1, #6.2, #6.4
- **relations:** build-vs-operate asymmetry; verification harness
- **verify-later:** none

### SYS-041 — Autonomous control loop (route-produce-verify-gate-apply-feedback)
- **status:** aspirational
- **status-evidence:** MASTER(4) §7 "the new machinery wraps each leaf task … the orchestrator's decompose-and-dispatch is reused unchanged."
- **what:** The orchestrator's decompose-and-dispatch is reused unchanged; new machinery wraps each leaf: route (cascade), produce, verify (harness), gate (trust ledger level), apply→derived-state, feed back. Ops re-enters the same loop, triggered by derived state instead of a build goal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, #7.6, #7.7
- **relations:** cascade router; verification harness; trust ledger; mediator routing model
- **verify-later:** existing orchestrator spawn/work-items machinery

### SYS-042 — Mediator routing model (change → consultees)
- **status:** aspirational
- **status-evidence:** FOCUS_mediator_routing_model.md#1 "the doc tree's metadata is the routing table … routing is matching a change descriptor against those tags."
- **what:** Routing reduces a change to a descriptor `{change_types, areas, touched_subsystems}` (paths→types via globs), queries the manifest for matching active standards, and acts on each by its own fields (run `check` validator / compose `reference` into prompt / consult concern agent / spawn area-owner). Runs a cheap tier always and an expensive tier on trigger; runs twice per change (pre from intent, post from diff).
- **sources:** ED/FOCUS_mediator_routing_model.md#2, #5, #6, #7
- **relations:** atomic standard (fields are the routing table); autonomous control loop; concern curators
- **verify-later:** proposed change classifier; path→area glob map

### SYS-043 — Wrapper-orchestrator pattern (pod lifecycle for spawned work)
- **status:** deployed
- **status-evidence:** Named as the canonical pattern across multiple eras: "Every pod-running agent needs a parent that spawned it … the rule we violated when we first wrote site-adoption-agent" (001(0)); "Pattern copied verbatim from med-export-orchestrator" and "mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim" (SQL-agent era, 104/121).
- **what:** Agents get a dedicated Kubernetes Job pod only when reached via `spawn_agent`→`call_agent`; substantive work reached via the generic entry point instead runs in-chassis with interleaved logs and blocks a shared pod slot. The convention: a thin wrapper orchestrator (spawn→call→complete) creates a dedicated Job pod for real work, giving clean per-correlation logs, isolation, and idle-timeout cleanup. Spawn-before-call ordering is required for target_role lookups.
- **sources:** WM/001_development_guide(0).md#every-pod-running-agent-needs-a-parent-that-spawned-it; 096_vet_med_url_discover_orchestrator.sql; 104_site_adoption_orchestrator.sql; 121_intent_collector_agents.sql; 122_diagnose_agents.sql
- **relations:** idle timeout + topic cleanup; agent message contract; Kafka topic model evolution
- **verify-later:** dev guide §wrapper test; spawn_actions.go; SpawnAgentAction; spawnAgentKubernetesJobFromDefinition

### SYS-044 — Loop mechanisms (workflow expansion, dispatch loop, ErrLoopExpansionHandled)
- **status:** deployed
- **status-evidence:** 001(0) Appendix C "Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps."
- **what:** A loop step resolves a collection, then `handleLoopExpansion` injects `{loop}_iter_{N}_{substep}` steps into the workflow plan plus a `_complete` aggregator; `setLoopVariable` sets the current item and propagates prior-substep outputs. The canonical use is the dispatch loop (claim→spawn→call→mark). `ErrLoopExpansionHandled` is a sentinel fixing a race where a fast child response would otherwise skip remaining iterations.
- **sources:** WM/001_development_guide(0).md#appendix-c-loop-mechanisms, #the-dispatch-loop-pattern, #the-race-condition-and-errloopexpansionhandled
- **relations:** work-item relay spine; claim_work_item; dynamic dispatch
- **verify-later:** loop_actions.go; loop_expansion_handler.go; coordinator.go continueExecution

### SYS-045 — Architectural tensions catalogue (infer-and-repair; multi-owner page identity)
- **status:** partial
- **status-evidence:** ARCH_TENSIONS(2) "An entry graduates from 'observed' to 'resolved' only when the resolution principle is actually enforced in code."
- **what:** A living catalogue naming genre-level design tensions that keep generating incidents. Tension #1: trusting LLM free-text structure as truth then repairing with starved heuristics vs deriving structure deterministically from the LLM's reliable signals. Tension #2: page identity re-derived in multiple stages that undo each other.
- **sources:** WM/ARCHITECTURAL_TENSIONS(2).md#tension-1, #tension-2
- **relations:** CanonicalisePage; adoption faithfulness strip; site plan reconciler
- **verify-later:** ValidateRoles/nestedRoleFromURL; CanonicalisePage; normaliseRole vs normalisePageType

### SYS-046 — Site / area / page component hierarchy
- **status:** partial
- **status-evidence:** site_components deployed and populated; site_areas/area_components created with default 'main' area backfill, but only the site level shows active use.
- **what:** Three-level slot resolution for page chrome: area_components (per site_area override) → site_components (site-wide header/footer/head with rendered_html + content_data for re-render, UNIQUE(site_id, slot_name)) → assembly. site_areas model major site sections with their own nav_style and theme_overrides; get_page_component(page, slot) walks area-then-site.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql; 014_site_areas.sql; 015_area_components.sql; 003_pages.sql#site_area_id
- **relations:** component-based headers; pages.site_area_id; site_components lock columns
- **verify-later:** area_components usage in production; get_page_component callers

### SYS-047 — Pages / page_components split (structure vs content)
- **status:** deployed
- **status-evidence:** 003 records the design correction: columns first added to pages then explicitly reverted — "Content (rendered_html, content_data) lives in page_components table. Pages table just needs workflow tracking fields."
- **what:** pages holds metadata, navigation and workflow (build_status planned→…→deployed/needs_rebuild, sections as planning reference, version) plus per-page rendered_header/rendered_footer/rendered_head for minimal reassembly; page_components holds the actual sections (position, slot_name, component_id, content_data, rendered_html, content_hash, review fields, deploy_commit, research_id). The intended three layers: content (content_items) → layout (page_components) → structure (pages).
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql; 004b_content_items.md; 005c_bk_page_components.sql
- **relations:** content_items layer; page build workflow; site snapshots
- **verify-later:** assembly path reading rendered_* columns; build_status writers

### SYS-048 — awaited_requests global request/response registry
- **status:** deployed
- **status-evidence:** Two schema generations matching the AwaitedRequest Go struct, plus later additions of 'processing' status, processing_started_at/processing_pod claim tracking, and a cleanup function.
- **what:** DB-backed registry matching Kafka responses to waiting orchestrations, solving the race where a child creates a request while the parent receives the response. Keyed by request_id with orchestration/correlation context, target agent, responses/requests topics, retry_version, reply_to_request_id chaining, timeout_at, and status lifecycle waiting→processing→processed/expired/cancelled/error. Expired rows are marked then purged after 7 days by cleanup_expired_awaited_requests.
- **sources:** docs/agent_docs/sql_for_tables/001_awaited_requests.sql; tables_sql/001_awaited_requests.sql
- **relations:** processed_messages idempotency; HITL; stale orchestration sweeper
- **verify-later:** state.go AwaitedRequest struct; cleanup scheduling

### SYS-049 — Message deduplication (processed_messages, terminal-state-once)
- **status:** deployed
- **status-evidence:** Designed in detail (dedup key request_id:retry_version:status) in the stateless-first-agents plan; independently confirmed live via `\d` output plus applied ALTERs (retry_version added, PK re-keyed to correlation_id/request_id/agent_id/retry_version).
- **what:** Before processing, agents check a dedup key against a `processed_messages` table; duplicate responses are dropped, and once any terminal state (complete/error_unrecoverable) is processed for a request, all further terminal responses for it are ignored — ensuring idempotency under Kafka redelivery and multi-replica consumption. The composite primary key including retry_version allows deliberate retries while blocking duplicate deliveries within a retry generation.
- **sources:** docs/plans/stateless-first-agents-001#7-deduplication-handler, #9-database-schema; docs/agent_docs/sql_for_tables/007_processed_messages.sql
- **relations:** retry semantics; optimistic locking; awaited_requests
- **verify-later:** consumer insert-or-skip logic

### SYS-050 — Orchestration ↔ site linkage (orchestration_states.site_id)
- **status:** deployed
- **status-evidence:** Migration with three-path backfill from collected_data (input_data.site_id, site_record.site_id, top-level) and verification counts against gamedesign.uk.
- **what:** A direct nullable site_id column on orchestration_states (set at creation) replaces JSONB spelunking for "orchestrations for this site", with a partial index for active orchestrations per site. Nullable because not all orchestrations are site-scoped (health checks).
- **sources:** docs/agent_docs/sql_for_tables/036_orchestration_states.sql
- **relations:** debugging queries; improvement-sweep pre_query
- **verify-later:** creation-time population in Go

### SYS-051 — Sites contact-identity denormalisation
- **status:** deployed
- **status-evidence:** Applied ALTERs + COALESCE backfills from content_data; one-off content_data patches for live sites.
- **what:** Frequently rendered identity/contact fields promoted from sites.content_data JSONB to first-class columns (company_name, tagline, email, phone, logo_url, logo_text, contact_address) feeding the render context for headers/footers/heads, with content_data retained as the brief-derived store of record.
- **sources:** docs/agent_docs/sql_for_tables/011_sites_table.sql; 018_site_work_items.sql#issue-1a; sql_for_content/001_phone_number.sql
- **relations:** component-based headers render context; site_specs identity aspect
- **verify-later:** which of sites columns vs site_specs.identity is authoritative for rendering today

### SYS-052 — Universal orchestration principle ("every agent is an orchestrator") & the agent_group_definitions elimination
- **status:** deployed
- **status-evidence:** README.002 "Current Implementation Status … ✅ Universal orchestration capability"; docs006/006 implementation doc confirms the concrete mechanism: "GroupDiscovery is aliased to AgentDefinitionDiscovery... FindBestGroup now queries agent_definitions."
- **what:** No architectural distinction between orchestrator and worker agents — every agent runs the same chassis, can spawn children, orchestrate workflows, and execute tasks simultaneously; complexity is fractal (agents compose into arbitrarily deep trees). This founding philosophy was made structurally literal when `agent_group_definitions` was eliminated in favour of `agent_definitions` carrying the orchestration workflow in default_config: a "group" is just an agent whose workflow spawns and calls other agents. `spawn_group` became a thin wrapper delegating to `spawn_agent`; discovery, message processor, and metadata all gained backward-compat aliases.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-2; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Implementation; docs006_workflow_builder/004_agent_groups_or_not.md#Key-Decisions
- **relations:** agent message contract; SagaCoordinator engine; agent families; superseded in practice for site building by the work-item pipeline but still the chassis foundation
- **verify-later:** platform/discovery/agent_discovery.go; absence/deprecation of agent_group_definitions table; spawn_group action code; whether dynamic spawn trees are still exercised vs static handler deployments

### SYS-053 — Stateless-first agent principle + database-backed orchestration state
- **status:** deployed
- **status-evidence:** "Core Principle: Agents are stateless executors. State lives in the database. Any replica can process any message for its agent type" (stateless-first-agents-001, v21) — presented as the implementation spec matching later operational docs; workflow status vocabulary (RUNNING/AWAITING_RESPONSES/COMPLETED/FAILED) consistent from the earliest architecture doc through the HITL era.
- **what:** Agents hold no orchestration state in memory; pod crashes lose nothing; all workflow execution state — status, current_step, workflow_plan, execution_path, collected_data, awaited_steps/awaited_requests, final_result — is persisted per correlation_id in the orchestrations/orchestrator_state table(s). Replicas scale horizontally with HPA (CPU + Kafka consumer lag); messages for one orchestration are ordered by using orchestration_id as the partition key; responses arriving at any pod are matched to awaited steps and the workflow continues when all responses are in.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy, #8-kubernetes-deployment; docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/architecture/003-flow-doc.md; docs/basic_usage/004_debugging
- **relations:** SagaCoordinator engine; orchestration-as-identity model; optimistic locking; workflow status state machine
- **verify-later:** deployment manifests (HPA config), consumer group setup, partition key usage; orchestrator_state and orchestrations tables — column set differences

### SYS-054 — ExecutionContext unified message envelope and ID semantics
- **status:** deployed
- **status-evidence:** README.002 "✅ ExecutionContext as unified message structure"; detailed ID-trace docs resolving the semantics.
- **what:** Every Kafka message carries an ExecutionContext: correlation_id ties the whole end-to-end operation; orchestration_id identifies one workflow instance; request_id identifies a single request/response cycle (new per communication); parent_orchestration_id records who called you; plus tree depth/path, fuel budget, timeout, responses_topic. The sender constructs the child's context; the receiver trusts headers.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; README.021.flow10.initialrequestflow.md; README.081.b.requestIDflow.md; README.043.spawn_actions2_stepbystepthroughthecode.md
- **relations:** perspective transformation; reply-to metadata; MessageType semantics
- **verify-later:** platform/orchestration/types ExecutionContext; messaging/context.go NewMessageContext

### SYS-055 — Two-phase agent lifecycle (spawn + initialize handshake)
- **status:** deployed
- **status-evidence:** flow11: "initialize is not treated as a command to start a workflow… handled as a special protocol action"; multiple traced runs.
- **what:** Spawning creates a K8s Job then sends an `initialize` protocol message; the new pod configures itself (role, topics), sends an initialization response, and only then does the parent resume and send `process` work. Initialize bypasses the workflow engine entirely — its only purpose is setup/readiness confirmation, isolating init failures from execution failures.
- **sources:** docs001_flow_general/README.022.flow11.initialisationflow.md; README.021.flow10.initialrequestflow.md; README.010.flow.md
- **relations:** spawn_agent; await_response semantics; a fire-and-forget spawn variant caused ignored init responses
- **verify-later:** processor.go initialize handling; SendInitializationResponse

### SYS-056 — SagaCoordinator engine: embedded per-pod, distributed, no central orchestrator
- **status:** deployed
- **status-evidence:** "every agent is both a worker and an orchestrator... eliminates single points of failure" (presented as a completed architecture); extensive traced executions show workflows stored as JSON `{start_step, steps{action, config, next_step}}` in agent_definitions/agent_group_definitions and executed live.
- **what:** Every agent pod embeds a full orchestrator (SagaCoordinator) instead of relying on a central orchestration service. Any pod of an agent type can start a workflow, and any pod can pick up a response and continue it, because state is in the shared database. The coordinator loads a JSON workflow from the DB, executes steps via an action registry, stores each step's result in CollectedData under the step name, pauses on `await_response: true` by recording request IDs in an AwaitedRequests map (status AWAITING_RESPONSES), and resumes when matching `in_response_to_request_id` responses arrive. `complete_workflow` packages results and replies to whoever is waiting — root vs child completion unified. Key architectural decision distinguishing the platform from Temporal/Airflow-style central schedulers.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#distributed-orchestration-model; docs/architecture/003-flow-doc.md; docs001_flow_general/README.010.flow.md; README.023.flow12.await_response.md
- **relations:** stateless-first principle; agent message contract; local vs remote actions
- **verify-later:** platform/orchestration/coordinator.go; actions/registry.go; whether coordinator still runs under current handlers

### SYS-057 — Reply-to metadata (__work_request__) and respond-to-caller convention
- **status:** deployed
- **status-evidence:** README.081.b "Clean Reply-To Architecture… Store reply-to metadata when receiving a work request, use it when completing"; stated as an operating rule in docs002/0100d.
- **what:** Each agent stores, at work-receipt time, the request_id it must answer and the parent's responses topic together (`__work_request__` in CollectedData) and uses them at complete_workflow. Rule: agents always respond to the *caller's* responses topic, never their own. Works at any hierarchy depth; replaced fragile multi-fallback lookups and fixed empty `in_response_to_request_id` bugs.
- **sources:** docs001_flow_general/README.081.b.requestIDflow.md; README.014.flow4.1.routingtooriginalsender.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md
- **relations:** ExecutionContext; agent message contract; CompleteWorkflowAction
- **verify-later:** BuildCollectedData storing __work_request__; workflow_actions.go completion path

### SYS-058 — Perspective transformation (sender constructs context, receiver trusts headers)
- **status:** deployed
- **status-evidence:** flow6: "The critical fix is in ProcessMessage… NewMessageContext(msg, headers, p.agentType). This ensures every agent sees the conversation from their own perspective."
- **what:** On receipt, NewMessageContext transforms the message into the receiving agent's own perspective (its own OrchestrationID becomes primary; the caller's becomes ParentOrchestrationID). The *sender* is responsible for correctly constructing the child's context headers; the receiver only deserialises and trusts them — earlier receiver-side guessing caused validation failures and misrouting.
- **sources:** docs001_flow_general/README.017.flow6.md; README.021.flow10.initialrequestflow.md
- **relations:** ExecutionContext; MessageType semantics
- **verify-later:** messaging/context.go NewMessageContext signature and transformation logic

### SYS-059 — MessageType semantics (request = actively working, response = reporting back)
- **status:** deployed
- **status-evidence:** README.070.b full conceptual write-up with log excerpts.
- **what:** MessageType describes what the agent is doing *now*, not what just happened: a parent that has received a child's response resumes its own workflow in "request" mode with InResponseTo cleared. Prevents routing/semantic confusion when continuing execution after responses.
- **sources:** docs001_flow_general/README.070.b.execution_context_flow.md
- **relations:** SagaCoordinator continueExecution; perspective transformation
- **verify-later:** continueExecution fresh-context construction

### SYS-060 — Fuel budget resource management
- **status:** partial
- **status-evidence:** Fully plumbed (FuelBudget field, fuel_budget=1000 header in live kcat commands, FuelUsed/RemainingFuelBudget response headers) but no doc claims it is actually enforced — "calculate properly in production" left as a TODO in the code path.
- **stage2-verified (2026-07-14):** unknown → partial — platform/governance/fuel.go CostTable/HasEnoughFuel/DeductFuel actually called in internal/agents/contentcreator/agent.go:234,244,590 (real rejection path on insufficient fuel) — genuine enforcement in one agent. coordinator.go:78 holds a fuelManager field but never calls HasEnoughFuel/DeductFuel (grep 0 hits) — not...
- **what:** A per-orchestration computational budget carried in the ExecutionContext, intended to bound resource consumption of agent trees (sub-invocations pass a reduced budget down and report fuel used back up the chain). Appears plumbed end-to-end but never confirmed as an enforced mechanism; current status in the 2026 platform is unverified.
- **sources:** docs/architecture/002-agent-chassis-docs.md#key-features; docs/plans/stateless-first-agents-001; docs001_flow_general/README.002.agent_orchestration1.philosophy.md; README.061.groupagents2.md
- **relations:** message header contract; long-term resource optimisation objectives
- **verify-later:** grep FuelBudget usage — is it ever decremented or checked?; whether the current work-item system retains fuel

### SYS-061 — Child-orchestration timeout monitor
- **status:** partial
- **status-evidence:** README.040 claims a configurable timeout (default 5 minutes) preventing zombie orchestrations; the later HITL timeout doc shows the config→Step.Timeout mapping was actually broken.
- **what:** Parents launch a goroutine per awaited child; on timeout it checks whether the parent still awaits that child, sends a timeout error response so HandleResponse processes it normally, and optionally marks the child orchestration failed. Timeout goroutines are in-memory only — recovery on pod restart was identified as a gap (later addressed by the stale orchestration sweeper).
- **sources:** docs001_flow_general/README.040.orchestration_actions.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** stale orchestration sweeper; HITL approval timeouts; DefaultRequestTimeout 180s
- **verify-later:** handleRequestTimeout in coordinator.go; recoverPendingTimeouts existence

### SYS-062 — Fan-out and awaited-response correlation
- **status:** deployed
- **status-evidence:** 003-flow-doc.md walks a live fan-out (reasoning + image agents) with correlation_id/causation_id header matching; basic_usage shows fan_out steps in the deployed website-builder workflow.
- **what:** A fan_out step sends parallel sub-tasks to multiple agent topics, records their request IDs in awaited_steps, and sets status AWAITING_RESPONSES. Each response carries correlation_id (workflow) and causation_id (the originating request_id); any receiving pod matches causation_id to an awaited step, stores the result under collected_data, and resumes when all are received. A later, unimplemented proposal (`parallel_steps` array, ExecutionMode enum) sketched extending this to non-blocking parallel dispatch; no doc claims it landed.
- **sources:** docs/architecture/003-flow-doc.md; docs/architecture/004-agent-chassis-architecture.md#response-handling-flow; docs/basic_usage/001basic_usage.txt; docs002_hitl_parallel/README.0110.parallel_execution_proposal.md
- **relations:** database-backed workflow state; Kafka topic model evolution; message header contract
- **verify-later:** fan_out action implementation; awaited_steps vs awaited_requests handling; whether run_parallel/parallel_steps ever landed in coordinator.go

### SYS-063 — Early long-term platform ambitions (self-organising networks, marketplace, multi-tenant, cross-cluster)
- **status:** aspirational
- **status-evidence:** README.002 "Long-Term Objectives (6-12 Months)": self-organising agent teams, agent marketplace, client-isolated multi-tenant namespaces, cross-cluster orchestration with geographic failover.
- **what:** The founding roadmap's horizon list. Multi-tenancy (client schemas) and cross-cluster work later materialised (multicluster docs); the agent marketplace and learned team compositions appear to have vanished.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#long-term-objectives
- **relations:** multicluster (live successor for cross-cluster); schema-per-client multi-tenancy; agent teams (abandoned)
- **verify-later:** none directly; council context only

### SYS-064 — Environment variable validation framework (abandoned)
- **status:** abandoned
- **status-evidence:** README.002 Week-3 objective ("EnvironmentBuilder… Validate all environment variables before agent spawn"); never mentioned again in any later doc.
- **what:** Planned framework to declare required/optional env vars per agent and validate before spawn, to prevent runtime failures. Silently dropped.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md
- **relations:** spawn_agent env var plumbing
- **verify-later:** grep EnvironmentBuilder

### SYS-065 — relationships table — first-class entity relationships
- **status:** partial
- **status-evidence:** Created in a docs006-era migration; a later doc (docs012) notes "Relationships (existing, empty)... This is PERFECT for semantic links between pages!" — table existed but unused, then earmarked for semantic page links.
- **what:** A generic first-class relationship entity (source/target entity id+type, relationship_type, direction, properties JSONB, status) modelled explicitly on website links ("relationships are like links — first-class objects with their own identity and state"), with relationship-scoped entity_state for learned communication preferences. Designed for org-framework roles, later reused conceptually for pillar↔cluster semantic page relationships.
- **sources:** docs006_workflow_builder/006_conclude_role_entity_strategy.md#Relationships-as-First-Class-Objects; 007_new_tables_entity_state_log.sql#7; docs012_site_maps_and_components/006_start_concluding_links.md#Part-1
- **relations:** link-management (semantic links); org framework
- **verify-later:** relationships table in clients_db and whether any rows exist; link_registry vs relationships usage

### SYS-066 — Agent families architecture (nav/links/design/content/entity/tools/feed/maintenance)
- **status:** partial
- **status-evidence:** A per-family status table ("populate_nav_tables — Deployed"; layout-architect — New; brand-designer — Future split) plus a Data Ownership Summary table mapping every table to an owner agent; phased plan 1→4.
- **what:** The master blueprint of the specialist-agent era: eight agent families each owning a data domain — navigation, links, design, content, entity data, tool builder tiers, news/content feed, and maintenance — with explicit "does NOT do" boundaries, a component-builder-v2 workflow sketch, site-type stress tests, and single-owner-per-table data governance. Much became real (nav actions, webdesign, feeds, maintenance→work items); some never did (layout-architect, nav-layout-agent, product-content-writer).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md; 002_full_new_agent_architecture.md; 018_agent_architecture_v3.md
- **relations:** nearly every other concept in this cluster; data ownership prefigures council-agent domain ownership
- **verify-later:** which family agents exist in agent_definitions today

### SYS-067 — "Database is source of truth, Git is the deployment artifact"
- **status:** deployed
- **status-evidence:** docs012/010 Core Principles ("Content lives in the database; Git is the deployment artifact"); an explicit reversal of an earlier docs012/004 stance ("GitHub as current source of truth"); everything after (rerender, section editor, work items) depends on it.
- **what:** The pivotal data-ownership decision of the era: page content, sections, nav, entities, and design specs live in Postgres; git repos hold only rendered deployment artifacts, rebuilt from DB at will. Enables rerendering, granular editing, locking, and maintenance — and makes external git edits an anomaly to detect rather than a normal input.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#Core-Principles; 004_more_on_links.md#Context; docs018_rerendering/010_section_editor_architecture.md
- **relations:** page_components; rerender; site-snapshots-and-revert
- **verify-later:** n/a (doctrine, observable in every pipeline)

### SYS-068 — Layer-1 / Layer-2 hack-resistance model
- **status:** deployed
- **status-evidence:** Stated as existing platform fact: "Layer 1 ... publishes outward and pulls inward but never serves inbound public traffic. Layer 2 is client delivery; today that is static assets on Backblaze S3 with nothing in the request path."
- **what:** The security posture the whole chatbot design defends: Layer 1 (core K8s cluster — agents, Kafka, Postgres, all credentials) never accepts inbound public traffic; it only publishes outward (site assets, data exports, context packs to S3) and pulls inward (recorded turns). Layer 2 is static-on-S3 with nothing in the request path. The edge worker is the only new Layer-2 compute; a sister-project appendix documents the "nginx box keeps getting hacked" experience motivating it.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#1, #Appendix-B; PLAN_isolated_chat_environment(4).md#Appendix
- **relations:** edge worker; isolated chat environment boundary contract; deploy-to-S3 path
- **verify-later:** ingress rules on ai-persona-system; B2 publish path

### SYS-069 — Gateway proxy pattern (auth-service → core-manager)
- **status:** deployed
- **status-evidence:** auth-service owns auth/user/subscription/projects; forwards all else to core-manager enriching `X-User-ID/Client-ID/Role/Tier/Email` headers; core-manager re-validates JWT with a shared JWT_SECRET_KEY or falls back to /auth/validate.
- **what:** Single-front-door architecture: auth-service is the only HTTP ingress; `.Any()` wildcard routes proxy to core-manager, which defines the actual method handlers. Core-manager independently validates already-issued tokens so it doesn't hard-depend on auth-service uptime.
- **sources:** archive_april_26/008b_admin_api_plan_v2.md#architecture-overview; 007b#gateway-pattern
- **relations:** Admin API; public API blocks
- **verify-later:** cmd/auth-service gateway/handlers.go; core-manager AuthMiddleware/TenantMiddleware

### SYS-070 — site-engine (API-only capture backend)
- **status:** deployed
- **status-evidence:** service header "site-engine: the capture backend for VM-hosted backend sites (API only)"; builds/tests pass; running live on the relojistas box.
- **what:** A stdlib-only Go binary that does the one thing a static page cannot: record a structured intent event server-side keyed by Host into a file store. Endpoints: `POST /intent` (capture then 303), `GET /api/hit` (1×1 beacon), `GET /stats` (key-gated), `GET /health`, plus later `GET /events` and `GET /access-digest`. No page rendering or content registry (the chassis owns both).
- **sources:** deploy_setup/working_dir/service(24).go#header; main.go#header; traffic_probe_runbook(12).md#1
- **relations:** replaces the abandoned standalone probe-go fork; page content owned by chassis intent-probe component
- **verify-later:** site-engine repo; go.mod module site-engine

### SYS-071 — Standalone "probe-go" service (abandoned first cut)
- **status:** abandoned
- **status-evidence:** Session 1 "Forked idea.uk's Go service into probe-go … Caveat raised next session: this drifted into a separate project"; Session 2 reframed it as not-a-separate-project.
- **what:** The original framing forked idea.uk's multi-vhost Go service (page-by-Host-header, page.go + domains.json in Go) into a self-contained project. Rejected because it sat too far from the website-building chassis; page.go and domains.json were removed and the engine was trimmed to an API-only backend with content moved to chassis build outputs.
- **sources:** traffic_probe_running_notes(27).md#session-1, #session-2, #session-3
- **relations:** superseded by site-engine + Layer-4/thin-Layer-5 framing
- **verify-later:** n/a (removed page.go/domains.json)

### SYS-072 — Layer-4-build + thin-Layer-5-VM-deploy framing
- **status:** deployed
- **status-evidence:** Session 2 conclusion "the probe is Layer 4 (build a targeted site) + a thin slice of Layer 5 (deploy a tiny backend to a VM instead of B2)"; decision to keep git→Actions and only swap the target.
- **what:** Rather than a side project, the probe reuses the existing build pipeline (Layer 4) and the git→self-hosted-Actions deploy seam (Layer 5), swapping only the destination from B2 to VM. The heavier chassis service-deployer adapter is the eventual move, not now.
- **sources:** traffic_probe_running_notes(27).md#session-2; traffic_probe_plan(11).md#where-we-are
- **relations:** underlies "commit is deploy" seam swap; defers P5 vmhost adapter
- **verify-later:** CONSOLIDATION_where_it_all_fits.md; PARALLEL_engine_deployment_and_layer5.md

### SYS-073 — Phased plan P0–P5 (traffic probe / VM-hosted backend sites)
- **status:** partial
- **status-evidence:** P0/P1/P2 done, P3 in progress ("Remaining for P3: land the chassis patch…"), P4 in progress at unit close, P5 not started.
- **what:** P0 structural decisions; P1 manual go-live (Path A); P2 wire deploy-on-update (two Actions); P3 make the probe a normal pipeline output (github_repo target selection + capture component + capability gate); P4 off-box collection + ranking; P5 registry + provisioning adapter.
- **sources:** traffic_probe_plan(11).md#phases; traffic_probe_running_notes(27).md#open-threads
- **relations:** contains most other traffic-probe concepts
- **verify-later:** n/a

### SYS-074 — VM-Hosted Backend Sites class (proposed doc 024)
- **status:** aspirational
- **status-evidence:** plan(11) "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)."
- **what:** The genuinely-new infrastructure: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. The traffic probe is instance #1 of this class.
- **sources:** traffic_probe_plan(11).md#framework-integration; traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle
- **relations:** class parent of intent-probe; D5 requires-backend gate
- **verify-later:** doc 024 existence; service_instances table

### SYS-075 — Pull architecture / no collector VM
- **status:** deployed
- **status-evidence:** "No third 'collector' VM: the serving box buffers (JSONL); the CLUSTER pulls over key-gated HTTPS … Pull keeps every credential in the cluster — boxes never hold DB/cluster secrets."
- **what:** Collection is pull-only: the serving box buffers events in JSONL and exposes them via key-gated HTTPS (/events, /access-digest, /stats); the cluster's scheduled collector pulls. No third collector VM and no push, because push or a middle VM would put DB/cluster secrets on the box and add attack surface for no gain. B2 remains an optional cold backup.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2; traffic_probe_runbook(12).md#6
- **relations:** rationale for /events + collector design; boxes hold only the read-only stats_key
- **verify-later:** n/a

### SYS-076 — idea.uk topology exception — static page + always-on backend, not pure edge
- **status:** deployed
- **status-evidence:** "Architecture (note: not edge-only)"; "Topology note: idea.uk is NOT pure-static/edge like the other chat domains."
- **what:** Every other "simple paid chat" domain on the platform is designed as static-S3 + a synchronous edge worker (no always-on compute). idea.uk breaks this pattern because its "tool" is a minutes-long multi-LLM + web-search job, not a synchronous chat turn: it needs a small always-on backend running the engine as a background task, with the static/embedded page posting to it and Stripe's webhook pointed at it. Flagged repeatedly as a deliberate, understood exception, not an oversight.
- **sources:** RUNBOOK_idea_uk(1).md "Architecture"; running_notes(44).md
- **relations:** hosting split (static front + small backend); service-deployer pattern
- **verify-later:** contrast against the actual edge-worker chat domains for confirmation

### SYS-077 — Agent chassis — generic configurable agent executor
- **status:** deployed
- **status-evidence:** Described as the running framework ("deploys as a scalable Kubernetes deployment, 3 replicas in production"); the HITL agent definition (2025-11-03) still references image `docker.io/aqls/agent-chassis:v1.0.407`.
- **what:** A single reusable Go binary that becomes any agent type via configuration: it consumes Kafka messages, loads its workflow config from the database (agent_definitions / agent_instances), executes the workflow, handles fuel checks, errors, metrics, and health endpoints. New agent types are created by adding DB configuration, not code.
- **sources:** docs/architecture/002-agent-chassis-docs.md; docs/architecture/023-spawning-agents.md#the-core-concept; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; agent spawning; SagaCoordinator engine
- **verify-later:** cmd/agent-chassis/main.go; platform/agentbase/; platform/messaging/processor.go; agent_definitions table

### SYS-078 — Local vs remote actions and the action registry
- **status:** deployed
- **status-evidence:** Both patterns documented as implemented ("Executed within the orchestrator itself" vs "Executed by other agents via Kafka").
- **what:** Workflow steps are either local actions run synchronously in the orchestrator (validate_input, transform_data, spawn_agent, process_data...) registered in a Go actionRegistry, or remote actions dispatched to another agent's Kafka topic with state moved to AWAITING_RESPONSES. The registry grew over time (spawn_agent, execute_llm_prompt, await_approval added later).
- **sources:** docs/architecture/004-agent-chassis-architecture.md#local-vs-remote-actions; 025-reusable-evolvable-agent-teams#step-3; docs/basic_usage/003_dynamic_prompt_improvement
- **relations:** agent-centric call_agent architecture; fan-out; HITL await_approval
- **verify-later:** platform/orchestration/actions/ directory; actionRegistry in coordinator.go

### SYS-079 — Message header contract (sender identity, in_response_to_*, status enum)
- **status:** deployed
- **status-evidence:** The stateless-first-agents plan (v21) defines the full header set; quick_hitl_test.sh (Nov 2025) sends live messages carrying orchestration_id, orchestration_name, step_name, message_type, from_agent_type, responses_topic headers.
- **what:** Rich request/response headers: sender AgentIdentity (agent_type, agent_id=pod name, version), correlation_id + human-readable correlation_name, orchestration_id/name, step_id/name, request_id, retry_version, parent orchestration linkage, message_id, fuel budget, timeout, routing topics. Responses echo in_response_to_request_id/step/orchestration and carry a status enum: awaiting | processing | complete | error_recoverable | error_unrecoverable, plus multipart flags, timing and fuel accounting.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/quick_hitl_test.sh; send_approval.sh
- **relations:** retry semantics; message deduplication; fuel budget
- **verify-later:** Go header structs in platform code; kafka message headers on live topics

### SYS-080 — Orchestration-as-identity model (AgentID = PodName)
- **status:** deployed
- **status-evidence:** "Orchestration (in DB) + Step + Request = 'Agent Instance'... AgentID = PodName (changes on restart, but that's OK)" — resolved an earlier mandatory-AgentID debate.
- **what:** The persistent identity of "an agent doing a task" is the orchestration record, not the pod. Pod name serves as AgentID purely for debugging (processing_history records which pod handled each step). Supersedes an earlier proposal that workflows resolve and pin specific versioned agent instances (stable/canary selection strategies) — that instance-pinning design was not carried forward.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/architecture/022-possible-agent-structure#the-case-for-mandatory-agentid
- **relations:** supersedes mandatory agent-instance resolution; stateless-first principle
- **verify-later:** whether processing_history with pod_name exists in current orchestrations table

### SYS-081 — Optimistic locking on orchestration state
- **status:** deployed
- **status-evidence:** Fully specified (version column, update_orchestration_if_version() SQL function, retry loop with backoff) but no later doc in this cluster confirms it shipped.
- **stage2-verified (2026-07-14):** unknown → deployed — state.go:789 'WHERE orchestration_id = $21 AND version = $22' (CAS update), state.go:826 optimistic-lock-failure error, state.go:834-872 UpdateStateWithRetry with maxRetries loop, coordinator.go:43-50 maxOptimisticLockRetries=15/ErrVersionMismatch. Resolves unknown -> deployed.
- **what:** Each orchestration row carries a version integer; replicas load state, apply a step, and save only if the version is unchanged (compare-and-swap), retrying on mismatch. Prevents two replicas from double-processing the same step. Paired with processing_history JSONB as the audit trail of which pod did what.
- **sources:** docs/plans/stateless-first-agents-001#3-database-backed-state-management, #9-database-schema
- **relations:** stateless-first principle; message deduplication
- **verify-later:** version column and update function in current schema; conflict-retry code

### SYS-082 — Retry semantics: same request_id, incremented retry_version
- **status:** deployed
- **status-evidence:** "Key Implementation Notes: Retry uses same request_id with incremented retry_version"; error_recoverable responses trigger up to 3 retries. No later confirmation found in this cluster.
- **stage2-verified (2026-07-14):** unknown → deployed — platform/orchestration/helpers.go:204 'if awaited.RetryVersion < 3' and :237 RetryVersion++ implement exactly the documented max-3-retries-with-incremented-retry_version scheme; state.go:167-221 uses retry_version in processed_messages idempotency key (ON CONFLICT (correlation_id, request_id, agent_id, retry_version...
- **what:** Failed remote calls are retried with the identical request_id and retry_version+1 so responses remain matchable and duplicates detectable. Recoverable errors retry (max 3), then fall through to unrecoverable which fails the orchestration and propagates an error to the parent. Progress statuses (awaiting/processing) are logged but never propagated upward; terminal states are processed exactly once.
- **sources:** docs/plans/stateless-first-agents-001#6-retry-logic, #key-implementation-notes
- **relations:** message header contract; message deduplication
- **verify-later:** retry handling in response processing code; awaited_requests (request_id, retry_version) PK

### SYS-083 — Agent-centric architecture: steps call agents, not topics
- **status:** deployed
- **status-evidence:** call_agent (with agent_type replacing raw topics) proposed as "your current code already does 90% of this"; later production seeds (HITL group definition) use call_agent steps.
- **what:** The primary abstraction is the agent (owning a 6–12 step workflow) rather than the workflow; steps invoke other agents (`action: call_agent, agent_type: X`) which have their own workflows, error boundaries and state, enabling recursive hierarchies (any agent can orchestrate; a copywriter can spawn a researcher). Topic resolution happens from agent type.
- **sources:** docs/architecture/022-possible-agent-structure#summary-agent-centric-architecture; docs/humanintheloop/hitl_agent_group_definition.sql; 023-spawning-agents.md
- **relations:** agent chassis; agent spawning; supersedes inter-agent invocation protocol v1
- **verify-later:** call_agent action and agent-type→topic resolution code

### SYS-084 — Inter-agent invocation protocol v1 (invoke_agent / agent_invocations) — superseded
- **status:** superseded
- **status-evidence:** Proposed InvokeAgentAction, ParallelInvokeAgentsAction, and an agent_invocations tracking table; later docs replace this with call_agent + orchestration hierarchy headers, and the agent_invocations table never reappears.
- **what:** The first design for agent-calls-agent: a dedicated invocation request/response envelope, per-pair topics (`system.agent.requests.{from}.{to}`), an agent_invocations audit table, and parent_correlation_id columns. Its essential ideas (parent linkage, deadline, fuel passing) survived into the header contract; the specific mechanism did not.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#1.2, #phase-3
- **relations:** superseded by agent-centric call_agent architecture and stateless header contract; project manager agent
- **verify-later:** confirm agent_invocations table absent from schema

### SYS-085 — Project Manager / User Representative agent hierarchy (abandoned)
- **status:** abandoned
- **status-evidence:** Designed across two early docs ("User Representative Agent... represents the users views against the project manager"); never appears in later seeds, groups, or the current spine.
- **what:** A top-level persona hierarchy: User → Project Manager agent (plans phases, delegates to specialist orchestrators, reviews deliverables) → Web Design Orchestrator → specialists, with a User-Persona agent negotiating on the user's behalf. The review/approval intent resurfaced later as HITL steps and content governance instead.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#architecture-overview; 007-roadmap.md#2.1-user-representative-agent
- **relations:** website-builder group (took its place); HITL approval mechanism
- **verify-later:** confirm no project-manager/user-representative agent_definitions exist

### SYS-086 — HTML-first progressive enhancement delivery
- **status:** deployed
- **status-evidence:** "Starting with plain HTML/CSS/JS is actually a very smart architectural decision"; html-developer seeds specify vanilla HTML with inline CSS; the current platform still renders plain HTML/CSS sites.
- **what:** Deliberate decision to generate plain HTML/CSS/JS websites rather than framework apps: easier for AI to generate and validate, no build step, universally hostable, fast; complexity added progressively (web components → PWA → framework only if needed). One of the few strategy decisions from this era that demonstrably survived into the present render pipeline.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md; 027-create-website-creation-system (html-developer config)
- **relations:** styling-render-pipeline; WordPress export agent (rejected sequel)
- **verify-later:** current renderer output format

### SYS-087 — Workflow status state machine
- **status:** deployed
- **status-evidence:** Consistent across eras: RUNNING/AWAITING_RESPONSES/COMPLETED/FAILED in the original architecture doc; RUNNING/AWAITING_RESPONSE/COMPLETED/FAILED in the HITL README; pending|processing|complete|failed variant in the stateless plan.
- **what:** The orchestration status vocabulary and its transitions: workflows run steps, park in an awaiting state while remote/human responses are outstanding, and terminate complete or failed. The HITL pause reuses the same awaiting state rather than introducing a special paused status. Minor naming drift across eras is itself a verification target.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/humanintheloop/HITL_README.md#workflow-states; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** stateless-first principle; HITL approval mechanism
- **verify-later:** canonical status enum in current schema/code

### SYS-088 — Human-readable orchestration and correlation names
- **status:** deployed
- **status-evidence:** stateless-first-agents-001 mandates orchestration_name/correlation_name alongside UUIDs; start_hitl_workflow.sh generates human-readable names in practice.
- **what:** Every orchestration and correlation carries a generated human-readable name in addition to its UUID, propagated through headers and stored in state, so debugging and monitoring read as narrative rather than UUID archaeology.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/start_hitl_workflow.sh
- **relations:** message header contract; kcat/db-inspector runbook
- **verify-later:** name-generation code; name columns in orchestrations table

### SYS-089 — Agent teams: composite/family/service-agent patterns (abandoned)
- **status:** abandoned
- **status-evidence:** An evaluation of three options (PM pattern, peer-to-peer squads, service-oriented) recommended service-oriented-then-squads; what actually shipped was the simpler agent-groups + call_agent model — the AgentFamily/SharedMemory and workflow-composition constructs never reappear.
- **what:** Design exploration for complex 50+-step workflows: composite agents (one external face, embedded sub-components), agent families with shared state and peer coordination, stateless reusable service-agents (date extractor, entity extractor) callable by any workflow, and workflows-invoking-workflows composition. Records the acknowledged framework limitation ("one agent = one workflow, flat orchestration, no concept of agent teams") that agent groups later addressed.
- **sources:** docs/architecture/021-current-framework-limitations; docs/architecture/022-possible-agent-structure
- **relations:** agent groups (the shipped resolution); agent-centric architecture; early long-term platform ambitions
- **verify-later:** n/a — confirm no AgentFamily/sub-workflow constructs in code
