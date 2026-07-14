
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka topics model: generic entry point vs per-spawn job topics vs fixed adapter topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) topics section describes current mechanics incl. stable identity naming
- **what:** Three patterns: `system.agent.generic.requests` (shared entry door for callers with no spawn relationship — scheduler, manual triggers; explicitly "the current door", expected to evolve into a formalised entry API); `job.<corr8>-<orch8>-<type>-<step>.requests` per-spawn topics set by setupAgentTopics; fixed `system.agent.<type>.*` topics for long-lived adapters. agent_definitions.topics jsonb is declarative only — the Deployment manifest actually subscribes.
- **sources:** 001_development_guide(5).md#Topics; 002(4)#Infrastructure
- **relations:** wrapper-orchestrator; idle timeout; topic cleanup
- **verify-later:** setupAgentTopics/createTopics in spawn_actions.go; deployment manifests

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Agent message structure and HITL response shape
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) reference section
- **what:** Messages have Kafka headers (correlation/orchestration/request ids, message_type, action, responses_topic, sender) and a body of `headers`/`config`/`input_data`. Agents always reply to the caller's responses_topic. HITL responses go to the agent's responses topic with `in_response_to_request_id` from awaited_requests and `sender_agent_type: human`.
- **sources:** 001_development_guide(5).md#Agent Message Structure
- **relations:** adapter response envelope contract; awaited_requests
- **verify-later:** MessageProcessor, awaited_requests table

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Orchestration state and collected_data as the workflow data bag
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5)/016 §6.0 describe live mechanics
- **what:** Each orchestration is an orchestration_states row (workflow_plan, collected_data, current_step, status). Steps communicate solely via dotted paths into collected_data; agents themselves are DB rows (`agent_definitions.default_config.workflow`), not Go types — the Go codebase contains actions, not agents. "Every agent is an orchestrator" is literal.
- **sources:** 001_development_guide(5).md#Orchestration State; 016 §6.0 "What an agent actually is"
- **relations:** loop mechanisms; workflow result contract
- **verify-later:** orchestration_states schema; coordinator.go continueExecution

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5): "the #1 cause of pipeline stalls"; design with A/B/C classification
- **what:** Timeout goroutines die with pods, leaving AWAITING_RESPONSES orchestrations stuck. A periodic 60s DB sweep on chassis pods (FOR UPDATE SKIP LOCKED) classifies expired awaited requests: child completed → synthesize response; child failed → forward; none/running → retry up to 3 then fail parent. Handles topic-expired case by directly advancing parent state.
- **sources:** 001_development_guide(5).md#Stale Orchestration Sweeper
- **relations:** spawn-handler hang (timeout_at not enforced); claimed-item-timeout
- **verify-later:** sweeper implementation; stale-orchestration-reaper scheduled task

<!-- SOURCE: U01_docs024_numbered_core.md -->
### site-work-orchestrator ordering and the dispatch pattern
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) workflow and resolved decisions 16–21
- **what:** Unified orchestrator processes pending items → verifies previous → runs due discovery → triages → updates profile. Dispatch = spawn→call per item with raw identifiers; orchestrator never pre-spawns static chains, never derives handler data, never passes work-item awareness to handlers (handlers self-contained, CLI-callable). Status tracking stays in the loop.
- **sources:** 002(4)#The orchestrator, #Dispatch pattern, #Resolved Decisions
- **relations:** build-dispatch-loop; handler contract
- **verify-later:** site-work-orchestrator definition currency vs build-dispatch-loop

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Entity data model (state-based lifecycle, news triggers, client-side real-time)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** 002(4): tables site_entities/site_entity_relationships "exist", entity_sources/entity_sync_log "planned"; Phase 3 item
- **what:** Structured data that generates pages (events, performers, venues, ticket tiers) with state-based lifecycle (announced→on_sale→…→historical), setup mode (work items) + discovery mode (scheduled sync), significant state changes triggering news via entity_sources.news_triggers; real-time data (prices, availability) served client-side from a data API, never through the work queue.
- **sources:** 002(4)#Entity Data Agent Family; #Site Type Stress Tests (events/boxing)
- **relations:** news feed pipeline; site API router (007)
- **verify-later:** site_entities usage; any entity-data-agent definition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Maintenance profile per-site configuration
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) JSON shape; audit config consumed by improvement loop
- **what:** sites.settings.maintenance_profile controls per-domain cadence (content/links/seo/compliance/content_feed/entity), budgets (llm_calls_per_cycle, max_auto_fixes_per_cycle), build tier, and audit group enablement; audit_pass_count also lives here.
- **sources:** 002(4)#Per-Site Configuration; 002d#Site-Type-Specific Audit Configurations
- **relations:** audit pass cap; growth budget (separate: site_specs growth_config)
- **verify-later:** sites.settings shape in production

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Idle timeout for spawned agents + topic cleanup strategy
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4): mechanism, config column, sync.Once shutdown safety, tuning SQL
- **what:** agent_definitions.idle_timeout_seconds → env var → idle-monitor goroutine exits the pod after inactivity (0 = forever for deployments). Topics: EPHEMERAL_TOPICS per-spawn today; agents never clean up topics — a conservative 10-min CronJob deletes topics with no matching pod, Kafka 7-day retention as backstop; a future shared-topics design (pre-created per-type topics, header routing, static group membership) makes cleanup a no-op.
- **sources:** 002(4)#Idle Timeout, #Shared Topic Strategy, #Topic Cleanup Design
- **relations:** pod accumulation debugging (016 §1)
- **verify-later:** idle_timeout_seconds values; topic-cleanup CronJob

<!-- SOURCE: U01_docs024_numbered_core.md -->
### business-intel shared-pod pattern (multi-type agents on one static pod)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 017 architecture section; ai_service placement rule
- **what:** Multiple agent definitions share one static pod via message routing (config.agent_type → selectWorkflow/FindBestGroup); consequence: ai_service must live in STEP config, not agent_config top-level, because agent_config comes from the pod's own type. Workflows are minimal action→complete with logic in Go. Single-replica contention accepted for batch work.
- **sources:** 017#Deployment, #ai_service on Shared Pods
- **relations:** wrapper-orchestrator (contrast); med pipeline (same pod)
- **verify-later:** business-intel deployment manifest

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### CollectedData: single-channel orchestration working memory and its pathologies
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Analysis only. No code changes proposed yet" (2026-05-11); duplication called "structural", observed in every log
- **what:** CollectedData (orchestration_states.collected_data JSONB) is the single channel for step outputs, routing metadata, loop variables and parent-reply context — "the most overloaded data structure in the system". Documented pathologies: recursive `__raw_message__` nesting (write amplification ×15 optimistic-lock retries), dual storage at step_name AND output_field, InitialRequestData/__raw_message__ overlap, six conflated data categories in one flat namespace, loop iteration data stored 3-4×, CleanDataMap stripping legitimately-named response fields. Recommendations R1–R6 (strip system keys from __raw_message__, pick one storage key, namespacing, loop GC, delta writes) proposed, untriaged.
- **sources:** FOCUS_collected_data_analysis.md (whole)
- **relations:** flat-namespace collision risk; compensating mechanisms; consumer-group race (duplicate keys as evidence)
- **verify-later:** BuildCollectedData / storeActionResult in coordinator.go; whether R1 ever landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flat-namespace collision risk and the compensating-mechanism accretion
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** dev-guide-documented incident: "section-editor declared content_data as optional and the nested-source loop silently lifted site_record.content_data … and overwrote a hero section" (referenced 2026-05-11)
- **what:** Because caller inputs, step results and site context share one flat map, actions can silently pick up `site_record.site_id`/`content_data` instead of caller-supplied fields. The framework compensates with UnwrapDeep, FindByPath prefix fallbacks, extractReplyToMetadata 3-tier priority, output_mapping — accreting workarounds faster than it consolidates. New code should use collision-free names (target_site_id convention); existing code left alone.
- **sources:** FOCUS_collected_data_analysis.md#4.4, #5; ASSESSMENT_imagery_phase_0_1…md#Caveat-1
- **relations:** CollectedData analysis; ExtractActionInputs conventions
- **verify-later:** ExtractActionInputs nested-source loop behaviour for undeclared fields

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Response-topic consumer group race (per-pod groups fan out every response)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "Discovery, not yet remediated" (2026-05-10); ~85 consumer groups on system.agent.generic.responses, only 3 live; two pods ran ProcessResponse on the same message 215ms apart
- **what:** The requests topic uses a shared stable consumer group but each chassis pod joins the responses topic under its own per-pod UUID group, so every response is delivered to every pod; each independently advances orchestration state, and the loser of the version race can flip a step to FAILED (observed on call_logo_gen). Mostly silent (idempotent writes) but structurally wrong; the system relies on shared-pool semantics it doesn't have. Open questions: intended model, per-spawn job.* topic groups, CAS hardening in ProcessResponse, 82 stale groups cleanup.
- **sources:** ANALYSIS_chassis_response_consumer_group_race.md (whole)
- **relations:** dispatcher stall Bug 1; duplicate collected_data keys; Phase 2F migration testing blocked
- **verify-later:** AgentClient constructor wiring (consumerGroup argument); ProcessResponse CAS behaviour in coordinator.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Kafka empty partition assignment on simultaneous pod join
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "five agent-chassis pods were members of generic-requests-group but all showed #PARTITIONS: 0 … Fix applied: Delete one pod to force rebalance" (2026-04-20)
- **what:** After a deploy where all pods join within the same second, the group can go Stable with the partition unassigned — zero consumption while offsets pile up and work items sit triaged. Workaround: kill a pod. Watch item on every deploy: at least one member must show #PARTITIONS: 1.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#1
- **relations:** consumer-group race; dispatcher reliability
- **verify-later:** whether staggered restarts or a fix was adopted

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Observability gaps: owner_agent_type "generic" and orchestration_name
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "orchestration_states rows where the generic agent routed to a different workflow still show owner_agent_type = 'generic'" (2026-04-20, P3)
- **what:** When the generic chassis routes a message to another agent's workflow (FindBestGroup), the orchestration is filed under owner_agent_type='generic' and orchestration_name doesn't carry the scheduler's sched-<task> name — searches by agent type or task name find nothing, which caused the "trigger never runs" misdiagnosis. Fix shape: selectWorkflow sets owner_agent_type to the resolved type.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7, items 7-8
- **relations:** content-feed-trigger bug; execution_path not populated (flywheel note 2.4c)
- **verify-later:** selectWorkflow in processor.go

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Four overlapping chrome default stores and the update_site_defaults linkage
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Notes (Sh) F-section: intended chain documented in 003's Site Component Linkage Contract; SPEC W4(c): "keep function-lookup as the norm … Smallest honest change: the comment" — the linkage deliberately left unrepaired.
- **what:** Header/footer defaults coexist in four stores: `style_collections.header/footer_component_id` (the operative read, dead-NULL), `site_components` slots (copy target + pre-render cache; idea.uk's were pinned to inactive components), `sites.default_components` JSONB (UpdateSiteDefaultsAction's target — a tracking copy nothing reads on the render path), and `layouts.default_*_component_id` (FK, all NULL, nothing copies it onward). The intended chain — style_collections as source of truth, `update_site_defaults` copying into site_components — never runs in composition (003's documented failure mode #1 IS idea.uk's case). The fix chose function-lookup as the norm rather than reviving the chain; populating style_collections at install remains a possible per-site-variety feature.
- **sources:** running_notes_scheme_to_components(55).md#Sg #Sh; SPEC_scheme_to_components.md#W4; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3b)
- **relations:** chrome selection path; Q6's original layouts.default_* direction (superseded by this resolution); site_components repoint.
- **verify-later:** v3_site_actions.go UpdateSiteDefaultsAction; whether the misleading install comment was deleted; any later population of style_collections chrome ids.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Orchestrator-agent architecture conventions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Notes (Um) context reset restates them as standing rules: "every agent is an orchestrator owning a workflow of ≥1 steps that call ACTIONS. Children respond to the PARENT's responses_topic … Do NOT create sub-workflows in SQL — spawn sub-agents."
- **what:** The platform's structural conventions, carried as strict constraints through both threads: every agent is an orchestrator owning a workflow of steps that call Go actions; workflows stay simple with complexity in action code; no sub-workflows in SQL — spawn sub-agents with their own workflows (clean logs, separated responsibilities); child agents respond on the parent's responses topic; workflow variable names stay in sync with action expectations; identifiers are never renamed silently; `logger.Debug` is banned (invisible in the log pipeline — use Info); reuse/alter existing functions and architecture before creating new.
- **sources:** running_notes_scheme_to_components(55).md#Architecture-conventions #Um; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** house rules / standing preferences; platform mission.
- **verify-later:** agent-creation guidelines doc in repo; logging config that swallows Debug.

<!-- SOURCE: U04_idea_uk.md -->
### Coordinator result-extraction contract (resolveResultSpec) and the silent-stub class
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Coordinator result-extraction fix — field-validated 2026-06-19 (idea.uk index built + deployed)"; git archaeology settled the cause to commit 06a8c6ef (14 Jan), unchanged since.
- **what:** The class of bug and its structural fix. Bug: workflow `complete` steps declaring singular `output_field` were never honoured (coordinator read plural `output_fields` only since 14 Jan), falling back to a working-state dump; when a big multi-section page's dump cleared the 900k `MaxResultSizeBytes` cap, `extractMinimalResult` returned a `status:"completed"` **stub** — silent false completion (gamesdesign) or, where the claimed-item evidence gate refuses 0-component pages, honest claim-timeout failure (idea.uk's empty index). **Size was the trigger; the singular key necessary-but-not-sufficient** (bucket audit: 100 plural steps safe, ~59 dump-bucket agents fine because small, 4 singular — only the writer breaks). Fix: centralised result contract in `result_spec.go` — singular→FLATTEN (named field's contents become the response body), plural→FIELDS (unchanged), `output`→MAPPING (previously silently dumped), none→dump; completion metadata via setIfAbsent; **oversize is now a loud error** routed to notifyParentOfFailure with a per-field size breakdown; stub removed; deprecated keys alias to result_from/multiple_output_fields/result_mapping. Retest doctrine: requeue the one failed page, do NOT re-adopt.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Empty index diagnosis + fix-direction 1 + retest + git section); idea.uk/running_notes(63).md (ff–qq checkpoints); idea.uk/README_001_todo_list.md
- **relations:** claimed-item evidence gate; debugging trap "inferring writers from readers"; MaxResultSizeBytes guardrail (do not raise).
- **verify-later:** platform/orchestration/result_spec.go + coordinator.go; the mode=flatten log on a writer run.

<!-- SOURCE: U04_idea_uk.md -->
### Hosting split: static-serverless front + small always-on backend ("static front + small back end")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Architecture doc §3 and the live topology (page embedded in the binary on the VM; B2 for everything chassis-built).
- **what:** The hosting taxonomy idea.uk established for the platform: pure-static content sites are serverless on B2; anything running a minutes-long multi-LLM job with a payment webhook cannot be serverless or edge-shaped — it needs a small always-on service with a stable inbound address. The classifier's `build_approach: hybrid` / `hosting_trajectory: needs_server` fields are the framework's slot for this distinction (noted as not yet confirmed in the classifier output). This is the hinge Layer 5 eventually automates, and why the engine can never be a forked client-side tool component.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#3; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 2 hosting reality); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (reuse section)
- **relations:** Layer-5 wrapper; VM cutover; tool-library boundary.
- **verify-later:** where build_approach/hosting_trajectory actually live (strategist? architecture doc concept only?).

<!-- SOURCE: U05_content_quality_linking.md -->
### Result-contract resolution (resolveResultSpec: flatten/nest/mapping)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(44): "FIX DEPLOYED to production 2026-06-18 … Part 1 re-confirmed healthy in prod (06-21, stub_rows=0)".
- **what:** The chassis coordinator's extractWorkflowResult honoured only plural `output_fields` since commit 06a8c6e (2026-01-14); a child declaring singular `output_field` (e.g. page-content-writer) fell to a state-dump fallback, exceeded the 900k cap, and was replaced by a stub still reporting status:"completed" — silently dropping the compiled page. Fix: a centralised resolveResultSpec — singular output_field/result_from → FLATTEN, plural → nest, output/result_mapping → applied, deprecated names read with a Warn (deprecation census drives a later rename migration). Pure chassis change, verified against all live consumers (flatten corrective for writer→3 parents, site-planner, model-trainer).
- **sources:** NOTES_gamesdesign_silent_norebuild(44).md#root-cause + Plan Step 1; resolveResultSpec.go.orchestrator_patch.txt; RUNBOOK_gamesdesign_index_rebuild(29).md#context
- **relations:** oversize fail-loud hardening; silent-completion family; select_sections mapping follow-up; content-reviewer auto_eval repoint (latent follow-up).
- **verify-later:** platform/orchestration/result_spec.go; coordinator.go extractWorkflowResult; deprecation Warn census in logs.

<!-- SOURCE: U05_content_quality_linking.md -->
### Oversize-result fail-loud hardening (stub removed)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(44) follow-up 1 "DONE (implemented 2026-06-18 … deployed)"; README_some_notes.md describes the intended behaviour change.
- **what:** extractWorkflowResultWithSizeLimit no longer truncates strings or emits the status:"completed" stub (extractMinimalResult deleted); on oversize it returns oversizeResultError with a per-field size breakdown naming the largest field, and notifyParentOfSuccess converts to a loud CHILD_ORCHESTRATION_FAILED + agent_error_log entry. Chosen fail-loud over persist-and-reference and over recursive trimming ("truncating content delivered as success is a corrupt result"). Any agent previously stub-"succeeding" now fails loudly until it declares a result contract — the surfacing is the point.
- **sources:** NOTES(44) Follow-ups; README_some_notes.md; RUNBOOK_gamesdesign_index_rebuild(29).md#6
- **relations:** result-contract resolution; MaxResultSizeBytes guards the Kafka ceiling (do NOT raise it).
- **verify-later:** coordinator.go oversizeResultError; agent_error_log for delivery-cap entries.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### sites.status is an informational lifecycle label (validated vocabulary, no consumer filters)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "No on-disk code filters sites on status — it is an informational lifecycle label; build dispatch keys on site_work_items" (RUNBOOK_scheme(18) sites.status RESOLVED, from v3_site_actions.go:323–395).
- **what:** UpdateSiteStatusAction validates status ∈ {draft, building, review, published, deployed, archived, error} (and stamps last_deployed_at with status=deployed); 'active' and 'system' are legacy out-of-vocabulary values on old rows. Nothing filters sites by status at build time — an assumption (`WHERE s.status='active'`) borrowed from an old handoff silently wrecked a blast-radius count, hence the standing rule: never filter on status='active'.
- **sources:** RUNBOOK_scheme_to_components(18).md §sites.status RESOLVED; running_notes(22).md Sr, Ss
- **relations:** work-item dispatch (the real gate); needle-gate/verify-at-point-of-use discipline.
- **verify-later:** UpdateSiteStatusAction vocabulary; legacy-status rows.

<!-- SOURCE: U08_travelling_docs.md -->
### Aspiration: agent-creation and inter-agent message logging workstream
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES "Note on a separate, out-of-scope item" — "kept out of these docs to preserve separation of concerns. Can be specced separately." No later mention.
- **what:** A stated desire to closely log/track agent creation and inter-agent messages (headers + body) as a distinct workstream from travelling docs — different responsibility and data. Never designed or built within this unit's horizon.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#out-of-scope-note
- **relations:** travelling docs (deliberately separated); message envelope standards (035).
- **verify-later:** whether any message-logging workstream exists elsewhere in docs/.

<!-- SOURCE: U09_adoption.md -->
### pages.sections is the build-read field (load_page_sections_from_spec fallback chain)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Decisive code facts established… load_page_sections_from_spec reads site_specs.site_plan (absent for this site) → falls back to pages.sections. pages.sections is the build-read field; site_plan_sections is NOT on the build path" (HANDOFF_2026-06-09, from the chassis dump).
- **what:** The page build resolves its section list via `site_specs` aspect `site_plan` (syncing into pages.sections when present) → `pages.sections` fallback → empty. `site_plan_sections` is relational plan hygiene only. `plan_sections` computes `ready_count` from resolvable section names; empty list → early return ready_count=0. Consequence: manual repairs must write pages.sections (the load-bearing statement), and a re-sync re-derives pages.sections from the in-memory plan (`extractPagesFromPlan`), so manual fixes survive only until the next full re-plan.
- **sources:** running_notes_15(10)#part-1–2, CONTEXT_PACK_adoption_skinner_box.md, HANDOFF_2026-06-06
- **relations:** sectionless durability; upsertPage EXCLUDED.sections; write_site_plan writes both tables
- **verify-later:** load_page_sections_from_spec_action.go source order; plan_sections_action.go ready computation

<!-- SOURCE: U09_adoption.md -->
### Chassis config location bugs (max_tokens shadowing, temperature single-read, error_step dead field)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** "max_tokens workaround applied for site-adoption-agent; error_step fix deployed via Path A (struct field + reader fallback); temperature TODO outstanding" (2026-05-18).
- **what:** A family of author-intuition vs chassis-reader mismatches: (1) `ExecuteLLMPromptAction` resolves the whole ai_service object once — a top-level ai_service shadows step-level config even when missing the field, so site-adoption-agent's analyze_site fell to the hardcoded 2048-token Anthropic fallback and truncated 8 of 20 pages (workaround: top-level max_tokens=16000); (2) temperature is read only from the very top of default_config — 6 step-level settings are dead and `llm_call_log.temperature` is NULL for every call (observability gap unresolved); (3) `error_step` at step level was silently dropped because the Step struct had no field — 62 dead graceful-degradation paths across 18 agents; Path A fix (ErrorStep struct field + config fallback) deployed, turning e.g. a firecrawl CSS timeout from fatal adoption failure into fallback-to-analyze_site with a partial fingerprint. Proposed structural fix: per-field resolution chains, raise the 2048 floor, log actual sent options.
- **sources:** FOCUS_chassis_config_location_bugs.md, old2/FOCUS_step_level_llm_config_ignored(3).md
- **relations:** 023 llm-quality-testing per-step model swap breaks on shadowed agents; fetch_primary_css hard-fail trade-off changed by error_step fix
- **verify-later:** execute_llm_prompt_action.go lines 110–145/209/213-218; contracts.go Step.ErrorStep; llm_call_log temperature population

<!-- SOURCE: U09_adoption.md -->
### work-site-orchestrator (monolith) vs build-site-planner (thin planner)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README maps every monolith inline step to its new-path equivalent; "the thin-planner model is the one that matches your stated philosophy"; design/imagery gaps closed per the mapping; site-status advancement flagged "worth confirming".
- **what:** Two architectures, not two versions: the old monolith did plan/write/sync/nav/design/render/deploy inline in one agent; build-site-planner is `read_specs → plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan (→ emit_design → emit_imagery) → complete`, delegating everything else via work items to handler agents. The audit question is "is every monolith inline piece now emitted as a work item on the build path" — answered yes for design/composition/rerender/JS once emit_design landed; open: no clear terminal step advances `sites.build_status` to deployed/active (sites may sit `pending` indefinitely).
- **sources:** README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** doc 029 Phase 1 (planner stops emitting work items); emit_design/emit_imagery; pageflow-builder (another still-active plan-less caller)
- **verify-later:** what sets sites.build_status terminal on the build path; whether work-site-orchestrator is still invocable

<!-- SOURCE: U10_imagery.md -->
### Snapshot-shadowing agent-definition loader defect
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Snapshot-shadowing hypothesis confirmed via SQL… Two loaders patched… Builds deployed as v1.0.1006"; "Snapshot audit closed clean" (2026-05-12).
- **what:** `snapshot_agent()` inserts snapshots at version+1000, so any loader using `ORDER BY version DESC LIMIT 1` without filtering `is_snapshot`/`is_active` reads the snapshot instead of the active row — every snapshot silently shadows its agent until the loader is fixed. `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` were patched; other loaders were already correct. Structural residue: snapshot retention policy and a single AgentDefinitionRepository remain open.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-1, STATUS_imagery_2026-05-12.md#Loader-snapshot-defect
- **relations:** model-infrastructure snapshot/rollback (021_model_swap_and_rollback.sql); parked "deep discussion" trio with the consumer-group race.
- **verify-later:** grep `FROM agent_definitions` across Go for is_snapshot filters; snapshot row counts.

<!-- SOURCE: U12_docs024_archives.md -->
### Quality Assurance Agent Architecture — folded into system-architecture, not abandoned
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded (as a standalone numbered doc; the architecture itself is deployed/partial)
- **status-evidence:** The standalone `002d_quality_assurance_architecture.md` (older1) and its later revision `002de_quality_assurance_architecture_v2.md` (archive_april_26) both appear verbatim as a "# 002d — Quality Assurance Agent Architecture" section inside live `002_system_architecture(4).md` (starting at line 897), continuing the main doc's Resolved-Decisions numbering (18-25) and extending it with a new "Layer 0: Pre-Generation Data Triage (plan_sections)" section, a "Content Validation as a Third Mode" table, and two further resolved decisions (24 "Quality gates before generation, not just after", 25 "needs_human_review is a first-class status") absent from any archived 002d draft. Its "Responsibility Boundaries" table was also updated to match the later composition/design-planner split.
- **what:** A three-layer QA model: Layer 1 structural/algorithmic checks (free, no LLM), Layer 2 LLM-assisted design/content audit (grouped agents sharing context, one LLM call per group), Layer 3 LLM-required strategic review (dream-spec gap analysis); plus a later-added Layer 0 pre-generation data triage (`plan_sections`). Includes the "promotion pattern" (a check starts as a `query_database` action step and is promoted to a spawned sub-agent only once it needs multi-step workflows or external calls) and the rule that audit agents "enforce, not override" the classifier/planner's stated intent. This was never a genuinely dropped concept area — it was consolidated into the numbered `002_system_architecture` doc rather than kept standalone, and then actively extended.
- **sources:** old/older1/002d_quality_assurance_architecture.md (whole file); archive_april_26/002de_quality_assurance_architecture_v2.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"002d — Quality Assurance Agent Architecture" (line 897+)
- **relations:** design agent responsibility split (site-design-planner/webdesign-agent); improvement-loop (004); site-spec-and-classifier (021); triage drain loop
- **verify-later:** confirm `design-audit-agent`, `visual-design-auditor`, `content-quality-auditor`, `site-review-agent` agent_definitions still implement the three-layer split; confirm `plan_sections` and `needs_human_review` status are implemented as described.

<!-- SOURCE: U12_docs024_archives.md -->
### site_work_items work-routing column renamed domain → pipeline
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Live bug-log entry #18 in `001_development_guide(5).md`: "The `domain` column on site_work_items was renamed to `pipeline` in a migration."
- **what:** Two archived dev-guide drafts each devote a "Lessons Learned" section to `site_work_items.domain` being an internal work-routing namespace ("build"/"maintenance"/"marketing") that collides confusingly with the website's actual domain (e.g. "gaswholesalers.com") — citing real bugs this caused (a dispatch-loop filter mismatch, and a CSS-generation item never dispatching because it was written with `domain:"design"` instead of `domain:"build"`). Rather than keep relying on documentation warnings, the column was renamed to `pipeline` at the schema level, eliminating the ambiguity outright; the live doc drops the explanatory section entirely in favour of a terse bug-log line.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Work item domain is NOT the site domain"; old/001_development_guide.md#"Work item domain is NOT the site domain"; docs024_key_docs_latest/001_development_guide(5).md#18; old_design_and_styling/016_debugging_guide_v2.md#"Schema reminder"
- **relations:** dispatch-loop input_mapping; site_work_items table
- **verify-later:** confirm `site_work_items.pipeline` column exists in the current schema and no code still reads/writes `domain` for this purpose.

<!-- SOURCE: U12_docs024_archives.md -->
### Dispatch-loop input_mapping path mismatch (spec-nested vs flat)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Documented as "most common systematic failure" with three named affected agents (`tool-improver`, `tool-auditor`, `rerender-pages`); not confirmed whether the flatten-in-dispatch-loop fix or per-handler fix was adopted.
- **what:** `build-dispatch-loop` maps a work item's `spec` JSONB as nested (`input_data.spec.component_id`), but handlers read flat (`input_data.component_id`), producing path-resolution errors. Preferred fix: flatten in the dispatch loop's `input_mapping`, following the existing `page_name?`/`reviewed_brief?` pattern.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; ExtractActionInputs cross-link
- **verify-later:** current `build-dispatch-loop` `input_mapping` config.

<!-- SOURCE: U12_docs024_archives.md -->
### Site-work-orchestrator dispatch loop — asset self-resolving storage URI
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive documents specific Go additions (`PresignedURLToS3URI`, `resolveStorageURIFromAsset`) and a full finetuning.uk worked dispatch trace; none of this Go-function-level detail or the worked example appears in live `002_system_architecture(4).md`'s "Dispatch Loop... (from 004_site_work_orchestrator)" section, which keeps only the abstract principles.
- **what:** When the dispatch loop's discovery-written work items carry presigned HTTPS asset URLs but `deploy_image_asset` needs `s3://` URIs, the fix was to have `asset-deployer` resolve its own storage URI from `asset_id` via a DB lookup rather than have the orchestrator pre-resolve it — keeping handler self-containment.
- **sources:** archive_april_26/004_site_work_orchestrator.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"Dispatch Loop: Dynamic Work Item Routing"
- **relations:** dispatch-pattern spawn→call; asset-deployer agent; handler self-containment principle
- **verify-later:** `grep -n "PresignedURLToS3URI\|resolveStorageURIFromAsset" platform/` to confirm these functions still exist.

<!-- SOURCE: U12_docs024_archives.md -->
### Self-spawning flat dispatch-loop (pre-scheduler design, superseded)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive (dated 2026-02-24) states as a "Key Design Decision": "No sub_workflows — they've been problematic," "No loops in dispatch — one item per invocation, self-spawns for next item"; the eventual system uses a genuine `"action":"loop"` construct driven by a scheduled 30s/120s kafka-scheduler tick.
- **what:** An early design decision to avoid the framework's loop/sub_workflow mechanism entirely, having `build-dispatch-loop` process exactly one item then spawn a fresh copy of itself. Later abandoned in favour of the scheduler-driven periodic trigger combined with the fully-developed in-workflow loop mechanism.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Key Design Decisions"; docs024_key_docs_latest/010_scheduler_and_tasks.md#"build-pipeline-trigger"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C — Loop Mechanisms"
- **relations:** loop-mechanisms (dev-guide appendix), scheduler-and-tasks, build-dispatch-loop agent
- **verify-later:** confirm `build-dispatch-loop`'s current agent_definition workflow uses the loop action, not self-spawning.

<!-- SOURCE: U12_docs024_archives.md -->
### claim_work_item atomic claim action + load_work_items first_item patch
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Archive lists these as "created, not yet committed" (Feb 2026); live loop-mechanisms appendix shows `claim_work_item` as a fully standard, already-existing action used in production.
- **what:** `claim_work_item` performs an atomic `UPDATE ... WHERE status IN ('triaged','approved') RETURNING id` so concurrent dispatch loops can't double-process the same item. The companion `load_work_items` patch added a `first_item` convenience field since the framework's path resolver doesn't support array indexing.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Completed Artifacts"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C"
- **relations:** dispatch-loop pattern, loop mechanisms, scheduler-and-tasks
- **verify-later:** none needed — graduated cleanly from draft to shipped mechanism.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### collected_data growth causing OOM-kills and lost work
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** "Status: Diagnosed, not yet fixed" (FOCUS_platform_reliability doc header)
- **what:** component-quality-auditor orchestrations were observed holding 18MB `collected_data`, causing OOM-kills at the 512Mi pod memory limit. OOM mid-publish causes phantom-completed orchestrations and cascading parent-orchestration timeouts. Suspected bloat contributors: `__raw_message__` duplicating input_data, unbounded `processing_history`, large uncleared LLM responses. A separate consumer-group bug (per-pod `a.AgentID` with FirstOffset, causing backlog replay on restart) was flagged in the same investigation, deliberately held back.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-1, js_snippets_news_gaswholesalers/old/TODO_orchestration_memory_bloat.md
- **relations:** Reaper mechanisms and gap; consumer group bug
- **verify-later:** orchestration_states.collected_data for component-quality-auditor, agent.go a.AgentID consumer group

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Page content-creation build pipeline trace (page-build-handler workflow)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** each hop verified directly against chassis source, 2026-05-20
- **what:** Documents, hop by hop, how a `pages` row's bare `sections` list becomes populated, deployed HTML with linked `page_components`: load_page_record → plan_sections (triages by schema source) → content writer's `extractResponseContent` (flat string, dead end for structured fields) → RenderComponentAction → CompilePageSectionsAction → SavePageSectionsAction (structured-metadata path preferred, HTML-regex fallback, orphans a page_component when metadata isn't fully recovered).
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow, js_snippets_news_gaswholesalers/old/page_content_creation_flow.md
- **relations:** Stale site_plan gap; per-section briefs gap; isolated build test methodology; extractResponseContent flat-string hypothesis (superseded)
- **verify-later:** LoadPageRecordAction, PlanSectionsAction, RenderComponentAction, SavePageSectionsAction

<!-- SOURCE: U13_docs024_small_dirs.md -->
### extractResponseContent flat-string hypothesis for FAQ root cause (superseded)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** old/015_content_data_persisted.md hypothesizes the writer's string-only extraction is the cause; the isolated build test later proved the writer populates a `questions` array correctly when given a clean plan
- **what:** An intermediate working hypothesis that the content writer itself could never populate a structured field like FAQ's `questions` array. Superseded once the isolated `faq-test` build proved the writer works correctly standalone; the real cause was Defect 1 (duplicate content surfaces).
- **sources:** js_snippets_news_gaswholesalers/old/015_content_data_persisted.md, js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-test-that-settled-the-cause
- **relations:** FAQ duplicate content-surface bug; page content-creation build pipeline trace
- **verify-later:** grep/inspect `questions`; `faq-test`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Site-chrome rendering gap (missing nav/header/footer in relay build path)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Measured baseline table shows "nav: 0" on all four rendered pages of dartsonline.com; hypothesis: "the RELAY build path lacks the site-chrome rendering step... no chrome-render step was observed" — not yet confirmed against site_components rows
- **what:** A suspected structural gap discovered via direct measurement of a live site (dartsonline.com, zero `<nav>` elements on every page, single stylesheet link pointing at a CSS file whose `needs_design` item was still triaged/undelivered): the newer work-item-relay build path may never invoke the chrome-rendering step that the older `pageflow-builder` path has, while `build-site-planner`'s `populate_nav_tables` only writes nav data, not rendered chrome. Same failure class as the "Design/composition work-item emission gap" found independently via code-reading in plainjanedomain.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE,#THE THREE-WAY SPLIT
- **relations:** Design/composition work-item emission gap; Three-way split quality-gap diagnostic method
- **verify-later:** SELECT * FROM site_components WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'

<!-- SOURCE: U14_docs019_runbooks.md -->
### Generic orchestrate envelope as the universal manual trigger
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D/§6F full kcat scripts: "The envelope's action: orchestrate + config.agent_type is the generic entry point"; reused verbatim for code-indexer, diagnose, and §7E.
- **what:** One trigger shape for hand-running any agent: kcat-produce to `system.agent.generic.requests` with correlation/orchestration/message/request ids, `action=orchestrate`, `config.agent_type=<entry agent>`, and task-specific `input_data`. Known wrinkles recorded: `site_id` intermittently arrives empty (reproducibility bug, parked), and runtime selectors (`site_id` vs `runtime_site` domain) drive different evidence filters in load_runtime.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7E
- **relations:** diagnose-orchestrator wrapper; needs_diagnosis intake (future replacement); correlation-id discipline
- **verify-later:** drafts/084_TRIGGER_diagnose_v1.sh; 080c/082 trigger scripts; site_id-empty envelope bug

<!-- SOURCE: U14_docs019_runbooks.md -->
### Workflow result contract — resolveResultSpec and preferred-key unification
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild "page-content-writer's singular output_field now flattens into the response … resolveResultSpec emits a resolved-contract line, a deprecation Warn per old key" (deployed 2026-06-18); code_retrieval_route(21) "STATUS 2026-07-04: DEPLOYED (image v1.0.1092) and the rename migration APPLIED + VERIFIED — all four agents on preferred result_from".
- **what:** The contract by which a completing workflow's result is extracted (result_from / output_field(s) / flatten / mapping). Its failure was the root of the flagship silent no-op (singular output_field ignored → page collapsed to a stub). Evolution: flatten fix + contract logging + deprecation census (2026-06-18) → Option A unification: the resolution table lifted verbatim into datahelpers/result_contract.go (ResolveResultSpec + ApplyResultSpec), complete_workflow delegating to it, agents migrated to preferred key spellings — one source of truth for coordinator and action paths.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#what-this-exercises; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2); docs019/RUNBOOK(31)_diagnosis_loop.md#6G (the fixture cause)
- **relations:** gamesdesign fixture; oversize-result delivery; loop_scope_field lesson (dotted-config class); parent result key under step name
- **verify-later:** datahelpers/result_contract.go + 7-case test; NNN_rename_complete_keys_preferred.sql applied state

<!-- SOURCE: U14_docs019_runbooks.md -->
### Oversize-result delivery: loud failure, size guards, responses-are-summaries
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "MEASURED (2026-07-03): … cd_bytes=1,270,781 — THE COMPLETION RESPONSE CARRIED THE FULL collected_data alongside the 6KB result"; "FIX = NNN_fix_diagnose_complete_output_fields.sql … APPLIED 2026-07-03"; gamesdesign_index_rebuild fix #3 "an undeliverable (oversize) result now fails loudly (error_unrecoverable + agent_error_log) instead of a status:'completed' stub".
- **what:** The family of oversize-result failures and their doctrine. Mechanism found in code: `result_from` is a key CompleteWorkflowAction never reads, so its fallback shipped the ENTIRE collected_data (1.27MB > Kafka ~1MB cap → Message Size Too Large; child fails at complete). Fixes: output_fields selection per agent; a response SIZE GUARD (max_response_bytes, truncatedResponseStub naming where the real result lives); earlier, removal of the silent "Full result exceeded size limit" completed-stub in favour of loud error_unrecoverable. Doctrine: responses are summaries; heavy artifacts live in the DB, retrievable by correlation_id; raising the broker cap is inversion, last resort.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#6; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2)
- **relations:** workflow result contract; bundle size doctrine; diagnosis_artifacts egress (same principle)
- **verify-later:** workflow_actions.go size guard + truncatedResponseStub; MaxResultSizeBytes const

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parent stores child result under the step name
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) follow-up 1 "RESOLVED 2026-07-03: jsonb_object_keys shows the child response stored under the STEP NAME call_diagnoser … 'diagnose-agent_result' never existed. Migration WRITTEN: NNN_fix_orchestrator_complete_key.sql".
- **what:** Engine behaviour: when a call step has no output_field, the child's response is stored in the parent's collected_data under the STEP NAME — imagined synthetic keys like `<agent>_result` never exist. Two orchestrators carried imagined keys; fixed by pointing complete steps at the real step-name keys. A recurring class with the dotted-config lookups.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 1)
- **relations:** workflow result contract; loop_scope_field lesson
- **verify-later:** collected_data ? 'call_diagnoser' on a post-migration diagnose run

<!-- SOURCE: U15_docs019_running_notes.md -->
### SagaCoordinator output_field singular/plural extraction contract
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Fix = `resolveResultSpec` (result_spec.go) treats singular as FLATTEN, keeps plural unchanged. Pure chassis change" (v2(36)).
- **what:** The workflow-completion contract by which a step's declared `output_field` (singular) or `output_fields` (plural) determines how `extractWorkflowResult` shapes the final result; the singular case was silently mishandled before the fix, and `complete_workflow`'s `result_from` is the fixed, now-correct field the diagnose-agent's own complete step uses.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 gamesdesign entry.
- **relations:** Gamesdesign silent-no-op bug; diagnosis-loop chassis integration architecture.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Kafka completion-payload size guard (message-too-large bug + shared result contract)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "§7E attempt 1 (17933a83) FAILED: 'message too large'... the number: diagnosis 6KB, cd 1.27MB ⇒ completion ships FULL collected_data... both §7E blockers fixed in one change-set" (NOTES_running_synthesis_v4(39).md headers, 2026-07-03).
- **what:** A production-triggered bug where an agent's Kafka completion message failed with "message too large" because the completion producer ships the FULL accumulated `collected_data` (1.27MB), not just the declared result (6KB) — triaged to the child-completion producer, fixed alongside a second guard-vs-expansion blocker, and generalised into "Option A": a shared result contract plus a response-size guard applied platform-wide (not just to the diagnose agent).
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-03 §7E entries (headers + surrounding turn log).
- **relations:** SagaCoordinator output_field contract; diagnosis-loop chassis integration architecture.
- **verify-later:** The "Option A" shared result contract's actual deployed shape; whether the size guard applies to all agents or only the diagnose path.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Child-completion result key convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "follow-up #1 resolved: child result key = STEP NAME `call_diagnoser`" (NOTES_running_synthesis_v4(39).md header, 2026-07-03).
- **what:** Confirms (against real orchestrator rows, not guessed) that a spawned child agent's result is read by the parent under the CALLING STEP's own name/output_field (e.g. `call_diagnoser`), not a role-based or agent-type-based key — resolving several rounds of guessed migration SQL in the diagnosis-loop chassis port.
- **sources:** NOTES_running_synthesis_v4(39).md header 2026-07-03; NOTES_running_synthesis_v2(36).md "diagnose migration CORRECTED against real rows."
- **relations:** Diagnosis-loop chassis integration architecture; Workflow default_config location convention.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Workflow default_config location convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "the query result overturned my main assumption... task_workflow / orchestrator_workflow / orchestration_workflow are ALL NULL on every working orchestrator... The workflow lives in default_config" (v2(36), 2026-06-17).
- **what:** A load-bearing, empirically-corrected fact about the chassis schema: an agent's actual workflow (start_step/steps graph) lives in `agent_definitions.default_config`, never in the three separately-named `*_workflow` columns that exist on the table — discovered only by querying real working rows after an entire migration draft was written on the wrong assumption (inferred from the dev-guide's prose example, which showed a workflow object but never named its column).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnose migration CORRECTED against real rows."
- **relations:** Real-rows-beat-prose discipline; diagnosis-loop chassis integration architecture.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Work-item relay spine (batons, handler_agent, the 30s pump)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread §2 "Builder route §B0–§B4 CLOSED: spine decided = the work-item relay (§B3, pre-registered criteria)"; README_flows describes the live pump.
- **what:** Builds move as a relay: the baton is a site_work_items row naming a handler_agent; build-pipeline-trigger (every 30s, behind a pre-query gate) seeds build_queue and picks one dispatchable site; build-dispatch-loop claims items atomically and spawns a dynamic handler per item. One hop = one baton, one agent, one site_specs entry, one new baton. Around it a fully enabled immune system: evidence-based claimed-item timeout (its SQL documents the gamesdesign false-positive lesson), feasibility-recheck, both reapers, archiver, cleanup — with improvement-sweep currently disabled (the improvement loop not running) while content-feed-refresh runs six-hourly.
- **sources:** HANDOFF_builder_thread.md#1,#2; README_flows.md
- **relations:** hop-insertion pattern; needs_diagnosis intake (reuses the relay); scheduler-and-tasks category
- **verify-later:** scheduled_tasks rows; improvement-sweep flag state

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Autonomous Build-and-Operate — the trust-not-capability thesis
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) header "Status: synthesis spine, built over several turns. Sections 6–8 are deliberately stubs to detail next"; §1 "the technical pieces are proven … the blocker is LLM-response uncertainty"
- **what:** Umbrella vision: everything already built is apparatus for a single reliability problem — bound LLM uncertainty at each step enough to progressively remove the human. The whole plan targets building/operating a real site (vonc.com) autonomously by composing the companion FOCUS mechanisms into one toolkit across the full lifecycle.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#1, ED/MASTER_autonomous_build_and_operate(4).md#9
- **relations:** umbrella over all the salience/standards/mediator/context FOCUS concepts below; vonc
- **verify-later:** none (discussion doc)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Build-vs-operate asymmetry
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §4 "Build … isolatable … Competition is safe. Operate … live, stateful … Competition is risky"
- **what:** Build work (actions, workflows, components, agent defs) is branchable/sandboxable so competition is safe and the ratchet moves fast; operate work (provisioning, scaling, incident response) is live and stateful so it leans on known-good + canary + rollback + tighter HITL. The cascade's tier mix shifts by domain.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#4, ED/MASTER_autonomous_build_and_operate(4).md#7.6
- **relations:** lifecycle map (Tier A/B/C); reliability cascade
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Lifecycle map by verifiability + containment (Tier A/B/C)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.1–6.2 tables; "Ceiling is separate from current maturity"
- **what:** Every capability's autonomy ceiling is set by two independently-failing factors — verifiability (can we tell against ground truth it's correct) and containment (how bad/reversible if wrong). Tier A (Go actions, SQL, component-structural, observability, rollback) reaches autonomy; Tier C (security, sharding, replication, live remediation, meta-loop) stays gated regardless of agent capability. Ceiling ≠ maturity drives where to invest.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.1, ED/MASTER_autonomous_build_and_operate(4).md#6.2, ED/MASTER_autonomous_build_and_operate(4).md#6.4
- **relations:** build-vs-operate asymmetry; verification harness
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Autonomous control loop (route-produce-verify-gate-apply-feedback)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7 "the new machinery wraps each leaf task … the orchestrator's decompose-and-dispatch is reused unchanged"; §7.7
- **what:** The orchestrator's decompose-and-dispatch is reused unchanged; new machinery wraps each leaf: route (cascade), produce, verify (harness), gate (trust ledger level), apply→derived-state, feed back. Ops re-enters the same loop, triggered by derived state instead of a build goal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#7.6, ED/MASTER_autonomous_build_and_operate(4).md#7.7
- **relations:** cascade router; verification harness; trust ledger; mediator
- **verify-later:** existing orchestrator spawn/work-items machinery

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Mediator routing model (change → consultees)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS_mediator_routing_model.md#1 "the doc tree's metadata is the routing table … routing is matching a change descriptor against those tags"
- **what:** Routing reduces a change to a descriptor `{change_types, areas, touched_subsystems}` (paths→types via globs), queries the manifest for matching active standards, and acts on each by its own fields (run `check` validator / compose `reference` into prompt / consult concern agent / spawn area-owner). Runs a cheap tier always and an expensive tier on trigger; runs twice per change (pre from intent, post from diff).
- **sources:** ED/FOCUS_mediator_routing_model.md#2, ED/FOCUS_mediator_routing_model.md#5, ED/FOCUS_mediator_routing_model.md#6, ED/FOCUS_mediator_routing_model.md#7
- **relations:** atomic standard (fields are the routing table); self-dev Position C; concern curators
- **verify-later:** proposed change classifier; path→area glob map

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Wrapper-orchestrator pattern (pod lifecycle)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "Every pod-running agent needs a parent that spawned it … the rule we violated when we first wrote site-adoption-agent"; canonical minimal wrapper med-export-orchestrator
- **what:** Agents get a dedicated Kubernetes Job pod only when reached via `spawn_agent`→`call_agent`; substantive work reached via the generic entry point runs in-chassis with interleaved logs and blocks a shared pod slot. The fix is a tiny wrapper orchestrator (spawn→call→complete) so real work runs in its own pod.
- **sources:** WM/001_development_guide(0).md#every-pod-running-agent-needs-a-parent-that-spawned-it, WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/007_adoption_pipeline_v3.md#the-adoption-agent
- **relations:** generic entry point vs job topics; site-adoption-orchestrator; agent = row in agent_definitions
- **verify-later:** SpawnAgentAction; spawnAgentKubernetesJobFromDefinition; setupAgentTopics

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Kafka topic model (generic entry point vs per-spawn job topics)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "system.agent.generic.requests — the generic entry point … job.<stable-identity>.requests — per-spawn dedicated topics"
- **what:** Two distinct patterns: the shared generic entry point (consumed by long-lived chassis replicas that run the workflow in-process via `config.agent_type`) for anything starting a workflow from outside a spawn tree, versus per-spawn `job.<stable-identity>.requests` topics for agent-to-agent traffic inside a workflow; plus per-type fixed topics for long-lived adapters.
- **sources:** WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; scheduled tasks target_topic; adapters
- **verify-later:** KAFKA_TOPIC(S) env; createTopics; MessageProcessor.extractGroupInfo

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Loop mechanisms (workflow expansion, dispatch loop, ErrLoopExpansionHandled)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix C "Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps"
- **what:** A loop step resolves a collection, then `handleLoopExpansion` injects `{loop}_iter_{N}_{substep}` steps into the workflow plan plus a `_complete` aggregator; `setLoopVariable` sets the current item and propagates prior-substep outputs. The canonical use is the dispatch loop (claim→spawn→call→mark). `ErrLoopExpansionHandled` is a sentinel fixing a race where a fast child response would otherwise skip remaining iterations.
- **sources:** WM/001_development_guide(0).md#appendix-c-loop-mechanisms, WM/001_development_guide(0).md#the-dispatch-loop-pattern, WM/001_development_guide(0).md#the-race-condition-and-errloopexpansionhandled
- **relations:** dispatch loop state machine; dynamic dispatch
- **verify-later:** loop_actions.go; loop_expansion_handler.go; coordinator.go continueExecution

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Architectural tensions catalogue (infer-and-repair; multi-owner page identity)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** ARCH_TENSIONS(2) "An entry graduates from 'observed' to 'resolved' only when the resolution principle is actually enforced in code"
- **what:** A living catalogue naming genre-level design tensions that keep generating incidents. Tension #1: trusting LLM free-text structure as truth then repairing with starved heuristics vs deriving structure deterministically from the LLM's reliable signals. Tension #2: page identity re-derived in multiple stages that undo each other.
- **sources:** WM/ARCHITECTURAL_TENSIONS(2).md#tension-1-trusting-llm-free-text-structure-as-truth-infer-and-repair-vs-deriving-structure-deterministically, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; adoption faithfulness strip; site plan reconciler
- **verify-later:** ValidateRoles/nestedRoleFromURL; CanonicalisePage; normaliseRole vs normalisePageType

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Agent = row in agent_definitions (workflow model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.0 "An agent is not a Go type, class, or file. It is a row in agent_definitions whose default_config.workflow is a declarative graph of steps"
- **what:** Agents live in the database, not the Go source. A workflow is a step graph threaded by dotted-path reads from a shared data bag; "every agent is an orchestrator" is literal. `spawn_agent`+`call_agent` are a pair. Traps: the description can contradict the config; `agent_definitions` may be read from more than one DB.
- **sources:** WM/016_debugging_guide_v2_44.md#6.0, WM/016_debugging_guide_v2_44.md#6, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; kafka topic model; loop mechanisms
- **verify-later:** agent_definitions (templates_db vs clients_db); default_config.workflow; SpawnAgentAction

<!-- SOURCE: U18_sql_for_agents.md -->
### Wrapper-orchestrator pattern ("spawns a temporary pod to do X")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Named as the canonical pattern in 104 ("Pattern copied verbatim from med-export-orchestrator..."), 121 ("mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim — the canonical scheduler-triggered wrapper + task-worker"), 122 (dev guide "does this agent need a wrapper?" test).
- **what:** Convention: substantive in-chassis work (long LLM loops, crawls, collections) must not run in shared generic pods; instead a thin orchestrator (spawn_agent → call_agent → complete) creates a dedicated K8s Job pod for the worker, giving clean per-correlation logs, isolation, and idle-timeout cleanup. Spawn-before-call ordering is required for target_role lookups (109/111/112).
- **sources:** 096_vet_med_url_discover_orchestrator.sql; 104_site_adoption_orchestrator.sql; 121_intent_collector_agents.sql; 122_diagnose_agents.sql
- **relations:** idle_timeout_seconds; scheduler-and-tasks; K8s Job lifecycle
- **verify-later:** dev guide §wrapper test; spawn_actions.go

<!-- SOURCE: U18_sql_for_agents.md -->
### idle_timeout_seconds (Job pod auto-exit)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 075 ALTER TABLE + fleet-wide backfill (180s) with rationale ("timer resets on every message... multi-step workflow... stays alive as long as responses keep arriving").
- **what:** Column on agent_definitions controlling how long a spawned Job pod waits with no messages before exiting cleanly (0 = no timeout for Deployment agents). Paired with TTLSecondsAfterFinished for cleanup. The 075 list doubles as a census of the then-live spawnable fleet.
- **sources:** 075_various_timeout_column.sql
- **relations:** wrapper pattern; K8s cleanup; debugging (timeouts)
- **verify-later:** chassis idle-timer implementation

<!-- SOURCE: U19_sql_tables_components.md -->
### Site / area / page component hierarchy
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** site_components deployed and populated (012); site_areas/area_components created with default 'main' area backfill, but only the site level shows active use; get_page_component fallback function defined.
- **what:** Three-level slot resolution for page chrome: area_components (per site_area override) → site_components (site-wide header/footer/head with rendered_html + content_data for re-render, UNIQUE(site_id, slot_name)) → assembly. site_areas model major site sections with their own nav_style and theme_overrides; get_page_component(page, slot) walks area-then-site.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql; docs/agent_docs/sql_for_tables/014_site_areas.sql; docs/agent_docs/sql_for_tables/015_area_components.sql; docs/agent_docs/sql_for_tables/003_pages.sql#site_area_id
- **relations:** component-based headers; pages.site_area_id; locks (site_components lock columns).
- **verify-later:** area_components usage in production; get_page_component callers.

<!-- SOURCE: U19_sql_tables_components.md -->
### Pages / page_components split (structure vs content)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003 records the design correction: columns first added to pages then explicitly reverted — "Content (rendered_html, content_data) lives in page_components table. Pages table just needs workflow tracking fields"; live dump confirms.
- **what:** pages holds metadata, navigation and workflow (build_status planned→…→deployed/needs_rebuild, sections as planning reference, version) plus per-page rendered_header/rendered_footer/rendered_head for minimal reassembly; page_components holds the actual sections (position, slot_name, component_id, content_data, rendered_html, content_hash, review fields, deploy_commit, research_id). 004b describes the intended three layers: content (content_items) → layout (page_components) → structure (pages).
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** content_items layer; page build workflow; site snapshots capture both.
- **verify-later:** assembly path reading rendered_* columns; build_status writers.

<!-- SOURCE: U19_sql_tables_components.md -->
### awaited_requests global request/response registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Two schema generations (tables_sql 001 → sql_for_tables 001 matching the AwaitedRequest Go struct), plus later additions of 'processing' status, processing_started_at/processing_pod claim tracking, and cleanup function.
- **what:** DB-backed registry matching Kafka responses to waiting orchestrations, solving the race where a child creates a request while the parent receives the response. Keyed by request_id with orchestration/correlation context, target agent, responses/requests topics, retry_version, reply_to_request_id chaining, timeout_at, and status lifecycle waiting→processing→processed/expired/cancelled/error. Expired rows are marked then purged after 7 days by cleanup_expired_awaited_requests.
- **sources:** docs/agent_docs/sql_for_tables/001_awaited_requests.sql; docs/agent_docs/tables_sql/001_awaited_requests.sql
- **relations:** processed_messages idempotency; HITL runbook; orchestration_states.
- **verify-later:** state.go AwaitedRequest struct; cleanup scheduling.

<!-- SOURCE: U19_sql_tables_components.md -->
### processed_messages idempotency dedup
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Live \d output plus applied ALTERs adding retry_version and re-keying the PK to (correlation_id, request_id, agent_id, retry_version).
- **what:** Exactly-once message processing guard: each consumed message records correlation/request/agent identity; the composite PK including retry_version allows deliberate retries while blocking duplicate deliveries within a retry generation.
- **sources:** docs/agent_docs/sql_for_tables/007_processed_messages.sql
- **relations:** awaited_requests retry_version; Kafka consumer semantics.
- **verify-later:** consumer insert-or-skip logic.

<!-- SOURCE: U19_sql_tables_components.md -->
### Orchestration ↔ site linkage (orchestration_states.site_id)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Migration with three-path backfill from collected_data (input_data.site_id, site_record.site_id, top-level) and verification counts against gamedesign.uk.
- **what:** Direct nullable site_id column on orchestration_states (set at creation) replaces JSONB spelunking for "orchestrations for this site", with a partial index for active orchestrations per site. Nullable because not all orchestrations are site-scoped (health checks).
- **sources:** docs/agent_docs/sql_for_tables/036_orchestration_states.sql
- **relations:** debugging queries; improvement-sweep pre_query.
- **verify-later:** creation-time population in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Sites contact-identity denormalisation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Applied ALTERs + COALESCE backfills from content_data (company_name/business_name, tagline/slogan, email/contact_email, phone/contact_phone, logo_text fallback chain); one-off content_data patches for live sites.
- **what:** Frequently rendered identity/contact fields promoted from sites.content_data JSONB to first-class columns (company_name, tagline, email, phone, logo_url, logo_text, contact_address) feeding the render context for headers/footers/heads, with content_data retained as the brief-derived store of record.
- **sources:** docs/agent_docs/sql_for_tables/011_sites_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#issue-1a; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** component-based headers render context; site_specs identity aspect (overlapping data — coherence question).
- **verify-later:** which of sites columns vs site_specs.identity is authoritative for rendering today.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Universal orchestration principle ("every agent is an orchestrator")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "Current Implementation Status … ✅ Universal orchestration capability"; repeated as a working rule in docs002 ("Every agent is an orchestrator").
- **what:** No architectural distinction between orchestrator and worker agents. Every agent runs the same chassis, can spawn children, orchestrate workflows, and execute tasks simultaneously; complexity is fractal (agents compose into arbitrarily deep trees). This is the founding philosophy of the agent-chassis platform.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md
- **relations:** agent chassis; SagaCoordinator workflow engine; agent groups; superseded in practice for site building by the work-item pipeline (development-guide) but still the chassis foundation.
- **verify-later:** agent-chassis main; platform/orchestration/coordinator.go; whether dynamic spawn trees are still exercised vs. static handler deployments.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Stateless agents with database-backed orchestration state
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002: "✅ Stateless agent design with database-backed state"; orchestration_states schema with version optimistic locking shown.
- **what:** Agents are ephemeral execution containers (K8s pods/Jobs); all orchestration state lives in the `orchestration_states` table (orchestration_id, current_step, awaited_requests, status, processing_history, version). Pod crashes lose no work; the DB is the authoritative source of truth.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#orchestration-state; docs001_flow_general/README.012.flow3.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** SagaCoordinator; AwaitedRequests map; environment reset runbook (truncates this table).
- **verify-later:** clients_db orchestration_states table; UpdateStateWithVersion; whether table is still active or superseded by work_items.

<!-- SOURCE: U20_legacy_docs_a.md -->
### ExecutionContext unified message envelope and ID semantics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "✅ ExecutionContext as unified message structure"; detailed ID-trace docs (flow10, requestIDflow) resolving the semantics.
- **what:** Every Kafka message carries an ExecutionContext: correlation_id ties the whole end-to-end operation; orchestration_id identifies one workflow instance; request_id identifies a single request/response cycle (new per communication); parent_orchestration_id records who called you; plus tree depth/path, fuel budget, timeout, responses_topic. Sender constructs the child's context; receiver trusts headers.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md
- **relations:** perspective transformation; reply-to metadata; MessageType semantics.
- **verify-later:** platform/orchestration/types ExecutionContext; messaging/context.go NewMessageContext.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Topic-per-agent Kafka communication
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** README.020.flow9: "LEGACY TOPICS (Pre-created)… Why Legacy Topics Persist"; the doc itself designs the job-topic replacement.
- **what:** Original model: static topics `system.agent.{type}.requests/responses` per agent type. Kept only as bootstrap/well-known entry points (initial client contact) after message-stealing and routing conflicts pushed the design to dynamic job topics.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#topic-per-agent; docs001_flow_general/README.020.flow9.topicflow.md
- **relations:** superseded by job-specific dynamic topics (hybrid model); ultimately by the work-item pipeline for site building.
- **verify-later:** which system.agent.* topics still exist on the cluster.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Job-specific dynamic Kafka topics (hybrid bootstrap model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow9 "after discussion" decision section + spawn_actions2 walkthrough showing `job.{corrID}-{orchID}-{agentType}-{step}.requests/responses` created at spawn; robot-hands runs used them live.
- **what:** Each spawn creates private per-orchestration topics from a "stable identity" (correlation short + orch short + agent type + spawning step). Root agents listen on standard pre-created topics; spawned agents get their topics via REQUESTS_TOPIC/RESPONSES_TOPIC env vars. Parents talk to children on the child's job topic; children reply to the caller's responses topic carried in headers. Solves the chicken-and-egg bootstrap problem and message collision between parallel jobs.
- **sources:** docs001_flow_general/README.020.flow9.topicflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** stable identity; spawn_agent; environment reset runbook (deletes job.* topics).
- **verify-later:** kafka.CreateStableIdentity; topic creation in SpawnAgentAction; current topic list.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Two-phase agent lifecycle (spawn + initialize handshake)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow11: "initialize is not treated as a command to start a workflow… handled as a special protocol action"; multiple traced runs.
- **what:** Spawning creates a K8s Job then sends an `initialize` protocol message; the new pod configures itself (role, topics), sends an initialization response, and only then does the parent resume and send `process` work. Initialize bypasses the workflow engine entirely — its only purpose is setup/readiness confirmation. Isolates init failures from execution failures.
- **sources:** docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.010.flow.md
- **relations:** spawn_agent; await_response semantics; a fire-and-forget spawn variant caused ignored init responses (flow12).
- **verify-later:** processor.go initialize handling; SendInitializationResponse.

<!-- SOURCE: U20_legacy_docs_a.md -->
### SagaCoordinator DB-defined JSON workflow engine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Extensive traced executions (flow docs); workflows stored as JSON `{start_step, steps{action, config, next_step}}` in agent_definitions/agent_group_definitions and executed live.
- **what:** The coordinator loads a JSON workflow from the DB, executes steps via an action registry, stores each step's result in CollectedData under the step name, pauses on `await_response: true` by recording request IDs in an AwaitedRequests map (status AWAITING_RESPONSES), and resumes when matching `in_response_to_request_id` responses arrive (join when the map empties). complete_workflow packages results and replies to whoever is waiting — root vs child completion unified.
- **sources:** docs001_flow_general/README.010.flow.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.006.executeLocalAction1.refactor_into_functions.md; docs001_flow_general/README.046.workflow_actions1.refactor_into_functions.md
- **relations:** action registry (validate_input, transform_data, execute_llm_prompt, spawn_agent, call_agent, aggregate_data, conditional_branch, complete_workflow…); await_approval builds on the same pause mechanism.
- **verify-later:** platform/orchestration/coordinator.go; actions/registry.go; whether coordinator still runs under current handlers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Reply-to metadata (__work_request__) and respond-to-caller convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.081.b "Clean Reply-To Architecture… Store reply-to metadata when receiving a work request, use it when completing"; docs002/0100d states the convention as an operating rule.
- **what:** Each agent stores, at work-receipt time, the request_id it must answer and the parent's responses topic together (`__work_request__` in CollectedData) and uses them at complete_workflow. Rule: agents always respond to the *caller's* responses topic, never their own. Works at any hierarchy depth; replaced fragile multi-fallback lookups and fixed empty `in_response_to_request_id` bugs.
- **sources:** docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.014.flow4.1.routingtooriginalsender.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md#response-topic-routing
- **relations:** ExecutionContext; CompleteWorkflowAction; early routing failure modes.
- **verify-later:** BuildCollectedData storing __work_request__; workflow_actions.go completion path.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Perspective transformation (sender constructs context, receiver trusts headers)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow6: "The critical fix is in ProcessMessage… NewMessageContext(msg, headers, p.agentType). This ensures every agent sees the conversation from their own perspective"; flow10 codifies sender responsibility.
- **what:** On receipt, NewMessageContext transforms the message into the receiving agent's own perspective (its own OrchestrationID becomes primary; the caller's becomes ParentOrchestrationID). The *sender* is responsible for correctly constructing the child's context headers; the receiver only deserialises and trusts them — earlier receiver-side guessing caused validation failures and misrouting.
- **sources:** docs001_flow_general/README.017.flow6.md; docs001_flow_general/README.021.flow10.initialrequestflow.md
- **relations:** ExecutionContext; MessageType semantics.
- **verify-later:** messaging/context.go NewMessageContext signature and transformation logic.

<!-- SOURCE: U20_legacy_docs_a.md -->
### MessageType semantics (request = actively working, response = reporting back)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.070.b full conceptual write-up with log excerpts ("exec_ctx_message_type":"request" when parent resumes).
- **what:** MessageType describes what the agent is doing *now*, not what just happened: a parent that has received a child's response resumes its own workflow in "request" mode with InResponseTo cleared. Prevents routing/semantic confusion when continuing execution after responses.
- **sources:** docs001_flow_general/README.070.b.execution_context_flow.md
- **relations:** SagaCoordinator continueExecution; perspective transformation.
- **verify-later:** continueExecution fresh-context construction.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Fuel budget resource limiting
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** FuelBudget field in ExecutionContext (README.002) and `fuel_budget=1000` header in test messages; CreateResponseContext takes "fuel used — calculate properly in production" (0100) — no doc claims enforcement.
- **what:** A per-orchestration computational budget carried in the ExecutionContext, intended to bound resource consumption of agent trees ("if budget.Remaining() < estimated.Cost() return cheaperStrategy()"). Appears plumbed but never implemented as an enforced mechanism.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.061.groupagents2.md (header)
- **relations:** long-term resource optimisation objectives.
- **verify-later:** grep FuelBudget usage — is it ever decremented or checked?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Child-orchestration timeout monitor
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** README.040 "Key Features of This Implementation: Configurable Timeout… default 5 minutes… Prevents zombie orchestrations" (claims implemented); HITL timeout doc later shows the config→Step.Timeout mapping was broken.
- **what:** Parents launch a goroutine per awaited child; on timeout it checks whether the parent still awaits that child, sends a timeout error response so HandleResponse processes it normally, and optionally marks the child orchestration failed. Timeout goroutines are in-memory only — recovery on pod restart identified as a gap.
- **sources:** docs001_flow_general/README.040.orchestration_actions.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** HITL approval timeouts; DefaultRequestTimeout 180s.
- **verify-later:** handleRequestTimeout in coordinator.go; recoverPendingTimeouts existence.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Parallel / fan-out execution in the coordinator
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 lists "Fan-out (parallel) execution implementation" under Outstanding Work; 0110 is an explicit proposal ("Proposed Implementation Strategy… parallel_steps array") with no completion claim.
- **what:** Design for non-blocking workflows: a step's config carries a `parallel_steps` array; executeParallelSteps dispatches all children, records all request IDs in AwaitedRequests, pauses once; processResponse joins when the map empties. Included ExecutionMode enum (sequential/parallel/fan_out). Image workflows sketched parallel_image_generation/batch_image_generation actions on the same idea.
- **sources:** docs002_hitl_parallel/README.0110.parallel_execution_proposal.md; docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** AwaitedRequests join semantics (multi-response already worked); batch-processing category is the modern relative.
- **verify-later:** whether run_parallel/parallel_steps ever landed in coordinator.go.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Early long-term platform ambitions (self-organising networks, marketplace, multi-tenant, cross-cluster)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 "Long-Term Objectives (6-12 Months)": self-organising agent teams, agent marketplace, client-isolated multi-tenant namespaces, cross-cluster orchestration with geographic failover.
- **what:** The founding roadmap's horizon list. Multi-tenancy (client schemas) and cross-cluster work later materialised (multicluster docs); the agent marketplace and learned team compositions appear to have vanished.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#long-term-objectives
- **relations:** multicluster (live successor for cross-cluster); database-and-infrastructure client schemas; marketplace = abandoned idea worth registering.
- **verify-later:** none directly; council context only.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Environment variable validation framework (pre-spawn config validation)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** README.002 Week-3 objective ("EnvironmentBuilder… Validate all environment variables before agent spawn"); never mentioned again in any later doc.
- **what:** Planned framework to declare required/optional env vars per agent and validate before spawn to prevent runtime failures. Silently dropped.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md
- **relations:** spawn_agent env var plumbing.
- **verify-later:** grep EnvironmentBuilder.

<!-- SOURCE: U21_legacy_docs_b.md -->
### "Every agent is an orchestrator" — elimination of agent_group_definitions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs006/006 implementation doc: "GroupDiscovery is aliased to AgentDefinitionDiscovery... FindBestGroup now queries agent_definitions"; backward-compat table showing both group_type and agent_type message formats work.
- **what:** Architectural unification: a "group" is just an agent whose workflow spawns and calls other agents, so agent_group_definitions was eliminated in favour of agent_definitions carrying the orchestration workflow in default_config. spawn_group became a thin wrapper delegating to spawn_agent; discovery, message processor, and metadata all gained aliases for backward compatibility. This is the foundational premise of the current hierarchical agent tree.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-2; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Implementation; docs006_workflow_builder/004_agent_groups_or_not.md#Key-Decisions
- **relations:** spawn-before-call pattern; intake orchestrator; agent families.
- **verify-later:** platform/discovery/agent_discovery.go; absence/deprecation of agent_group_definitions table; spawn_group action code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### relationships table — first-class entity relationships
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Created in docs006/007 migration; docs012/006 (later era) says "Relationships (existing, empty)... This is PERFECT for semantic links between pages!" — table existed but unused, then earmarked for semantic page links.
- **what:** A generic first-class relationship entity (source/target entity id+type, relationship_type, direction, properties JSONB, status) modelled explicitly on website links ("relationships are like links — first-class objects with their own identity and state"), with relationship-scoped entity_state for learned communication preferences. Designed for org-framework roles, reused conceptually for pillar↔cluster semantic page relationships.
- **sources:** docs006_workflow_builder/006_conclude_role_entity_strategy.md#Relationships-as-First-Class-Objects; docs006_workflow_builder/007_new_tables_entity_state_log.sql#7; docs012_site_maps_and_components/006_start_concluding_links.md#Part-1
- **relations:** link-management (semantic links); org framework.
- **verify-later:** relationships table in clients_db and whether any rows exist; link_registry vs relationships usage.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent families architecture (nav/links/design/content/entity/tools/feed/maintenance)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** docs017/019b (v5) with per-family status columns ("populate_nav_tables — Deployed"; layout-architect — New; brand-designer — Future split) and a Data Ownership Summary table mapping every table to an owner agent; phased plan 1→4.
- **what:** The master blueprint of the specialist-agent era: eight agent families each owning a data domain — navigation (nav tables), links (algorithmic health), design (brand/layout/CSS split), content (marketing/legal/SEO/product writers + reviewer + researcher), entity data, tool builder tiers, news/content feed, and maintenance — with explicit "does NOT do" boundaries, a component-builder-v2 workflow sketch, site-type stress tests (brochure/e-commerce/finance/events/platform), and single-owner-per-table data governance. Much became real (nav actions, webdesign, feeds, maintenance→work items); some never did (layout-architect, nav-layout-agent, product-content-writer).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md; docs017_legacy_agent_rules_images_design_keydocs/002_full_new_agent_architecture.md; docs017_legacy_agent_rules_images_design_keydocs/018_agent_architecture_v3.md
- **relations:** nearly every other concept in this unit; data ownership prefigures council-agent domain ownership.
- **verify-later:** which family agents exist in agent_definitions today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** docs018/009_stale full design ("This is the #1 cause of pipeline stalls"; synthesize/retry/fail classification; "No schema changes needed"); no deployment confirmation in this unit.
- **what:** Replace lossy in-process timeout goroutines with a periodic DB sweep on every chassis pod: claim expired awaited_requests (FOR UPDATE SKIP LOCKED, 30s grace, LIMIT 20), classify — child COMPLETED means the response was lost, so synthesize a completion message from the child's final_result to the parent's topic; child FAILED forwards failure; no/running child retries up to retry_version 3 then fails the orchestration. Handles cascading stalls oldest-first and dead job topics by directly advancing parent state.
- **sources:** docs018_rerendering/009_stale_orchestration_sweeper_design.md
- **relations:** parent-timeout race; awaited_requests; debugging pipeline stalls; idle timeout/cleanup in system-architecture.
- **verify-later:** platform/orchestration/sweeper.go existence; sweeper startup in agentbase.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent message structure & spawn+call pattern (external triggering)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs018/009_agent and docs018/005 document the working format ("Agents don't exist as running processes until spawned... You cannot call an agent that hasn't been spawned in the current workflow"; responses_topic always the caller's).
- **what:** The canonical three-layer message (Kafka headers, mirrored JSON headers, config.workflow + input_data payload) for driving agents from CLI or external systems, with inline workflow support on the generic agent, mandatory spawn-before-call, and reply routing to the sender's responses_topic enabling parent-child orchestration. The operational lingua franca of the whole system.
- **sources:** docs018_rerendering/009_agent_initial_message_structure.md; docs018_rerendering/005_triggering_agent_from_kafka.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL protocol; generic agent as thin launcher; kafka reset/cleanup runbooks (docs007/003).
- **verify-later:** current message types in platform/orchestration/types.

<!-- SOURCE: U21_legacy_docs_b.md -->
### "Database is source of truth, Git is the deployment artifact"
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs012/010 Core Principles ("Content lives in the database; Git is the deployment artifact"); reversal of docs012/004's "GitHub as current source of truth, database for metadata/links"; everything after (rerender, section editor, work items) depends on it.
- **what:** The pivotal data-ownership decision of the era: page content, sections, nav, entities, and design specs live in Postgres; git repos hold only rendered deployment artifacts, rebuilt from DB at will. Enables rerendering, granular editing, locking, and maintenance — and makes external git edits an anomaly to detect (git_hook_adapter desync idea) rather than a normal input.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#Core-Principles; docs012_site_maps_and_components/004_more_on_links.md#Context; docs018_rerendering/010_section_editor_architecture.md#The-Source-of-Truth-Principle
- **relations:** page_components; rerender; site-snapshots-and-revert (later formalization); deployment-github.
- **verify-later:** n/a (doctrine, observable in every pipeline).

<!-- SOURCE: U22_recent_small_docs.md -->
### Layer-1 / Layer-2 hack-resistance model
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Stated as existing platform fact (Appendix B): "Layer 1 ... publishes outward and pulls inward but never serves inbound public traffic. Layer 2 is client delivery; today that is static assets on Backblaze S3 with nothing in the request path."
- **what:** The security posture the whole chatbot design defends: Layer 1 (core K8s cluster — agents, Kafka, Postgres, all credentials) never accepts inbound public traffic; it only publishes outward (site assets, data exports, context packs to S3) and pulls inward (recorded turns). Layer 2 is static-on-S3 with nothing in the request path — "nothing to compromise." The edge worker is the only new Layer-2 compute, and the whole chatbot design is arranged to preserve this (no API keys in the page, no central VM in front of static content). Sister-project appendix documents the "nginx box keeps getting hacked" experience that motivates it.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#1, #Appendix-B, docs025.../PLAN_isolated_chat_environment(4).md#Appendix
- **relations:** edge worker, isolated chat environment boundary contract, deploy-to-S3 path
- **verify-later:** ingress rules on ai-persona-system; B2 publish path

<!-- SOURCE: U23_docs_root_vonc.md -->
### Result-contract drop fix (child workflow result replaced by a stub)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016b Part 1: "DONE... Shipped 2026-06-18 (result_spec.go + coordinator.go); verified — gamesdesign index rebuilt+deployed 06-19."
- **what:** The chassis coordinator used to discard a child workflow's result (singular output_field, or oversize) and substitute a stub that still reported success — producing no-op saves under `complete` status (a root member of the silent-success family, and the resolution of the long-open "index returns thin content" question: it was the stub, not thin generation). Fixed in result_spec.go + coordinator.go. Carried here because the 016b copies in this unit are the guide's cumulative record; the docs024 consolidated guide is the canonical home.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 1)
- **relations:** complete_error family (sibling); trust-the-artifact doctrine
- **verify-later:** result_spec.go / coordinator.go in platform

---

## Dropped-concept notes from family deltas (audit)

- RUNNING_NOTES_vonc.md (base): migration renumbering — "DB migration: Migration 003" became
  "Migration 002 = agent_definition condition fix; render_mode sweep DROPPED" (captured as a
  concept above).
- RUNBOOK_vonc_migrations early versions: original "Migration 002 — Fix render_mode on
  components" heading and the "Recommendation: Option 1 for the first deployable version"
  line (both captured as abandoned concepts above); "Two fix options — choose one" framing.
- RUNBOOK_phase2 early versions: Gap-2 sub-options (a) creator-emits-companion-snippet vs
  (b) loader-snippets-as-library-fixtures (superseded by Tier E + loader-builder); the
  step-checklist framing (P2-1..P2-6, FX-1..FX-6, PD-1..PD-3) whose IDs the later docs still
  reference; all content otherwise carried forward into (29).
- PLAN_provocation-card(2)→(3): the pre-correction trim method (hand-UPDATE the live
  instance) — replaced by the bundle-verdict method; captured under "sanctioned edit paths".
- PLAN_dynamic_sections(2)→(4): `site_plan_directives` named as a site_plans child table in
  the pre-supersession decision text; not mentioned in the final doc (noted in the plan-storage
  concept).
- HANDOFF base/(1)→(2): same method correction (rejected mechanism → bundle §4.0).
- bundle_minilobby_trim base..(3)→(4): resolver evolution v1→v4 (registry-first resolver,
  module boundary, scope-by-symbol) — captured under the cmd/bundle concept.
- provocations.sample.json v1→v3: additive key evolution today/lobby → +arena → +archive —
  captured in the data-contract concept.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Gateway proxy pattern (auth-service → core-manager)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 008b: auth-service owns auth/user/subscription/projects; forwards all else to core-manager enriching `X-User-ID/Client-ID/Role/Tier/Email` headers; core-manager re-validates JWT with shared `JWT_SECRET_KEY` or falls back to `/auth/validate`.
- **what:** Single-front-door architecture: auth-service is the only HTTP ingress; `.Any()` wildcard routes proxy to core-manager which defines the actual method handlers. Core-manager independently validates already-issued tokens so it doesn't hard-depend on auth-service uptime.
- **sources:** archive_april_26/008b_admin_api_plan_v2.md#architecture-overview; 007b#gateway-pattern
- **relations:** Admin API; public API blocks 2/4
- **verify-later:** cmd/auth-service gateway/handlers.go; core-manager AuthMiddleware/TenantMiddleware

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### site-engine (API-only capture backend)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** service(24).go header "site-engine: the capture backend for VM-hosted backend sites (API only)"; builds/tests pass; running live on relojistas box.
- **what:** A stdlib-only Go binary that does the one thing a static page cannot: record a structured intent event server-side keyed by Host into a file store. Endpoints: `POST /intent` (capture then 303), `GET /api/hit` (1×1 beacon), `GET /stats` (key-gated), `GET /health`, plus later `GET /events` and `GET /access-digest`. No page rendering or content registry (the chassis owns both).
- **sources:** deploy_setup/working_dir/service(24).go#header, deploy_setup/working_dir/main.go#header, traffic_probe_runbook(12).md#1
- **relations:** replaces the abandoned standalone probe-go fork; page content owned by chassis intent-probe component
- **verify-later:** site-engine repo (`$OWNER/site-engine`); go.mod `module site-engine`

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Standalone "probe-go" service (abandoned first cut)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Session 1 "Forked idea.uk's Go service into probe-go … Caveat raised next session: this drifted into a separate project"; Session 2 reframed it as not-a-separate-project.
- **what:** The original framing forked idea.uk's multi-vhost Go service (page-by-Host-header, page.go + domains.json in Go) into a self-contained project. Rejected because it sat too far from the website-building chassis; page.go and domains.json were removed and the engine was trimmed to an API-only backend with content moved to chassis build outputs.
- **sources:** traffic_probe_running_notes(27).md#session-1, traffic_probe_running_notes(27).md#session-2, traffic_probe_running_notes(27).md#session-3
- **relations:** superseded by site-engine + Layer-4/thin-Layer-5 framing
- **verify-later:** n/a (removed page.go/domains.json)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Layer-4-build + thin-Layer-5-VM-deploy framing
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Session 2 conclusion "the probe is Layer 4 (build a targeted site) + a thin slice of Layer 5 (deploy a tiny backend to a VM instead of B2)"; decision to keep git→Actions and only swap the target.
- **what:** Rather than a side project, the probe reuses the existing build pipeline (Layer 4) and the git→self-hosted-Actions deploy seam (Layer 5), swapping only the destination from B2 to VM. The heavier chassis service-deployer adapter is the eventual move, not now.
- **sources:** traffic_probe_running_notes(27).md#session-2, traffic_probe_plan(11).md#where-we-are
- **relations:** underlies "commit is deploy" seam swap; defers P5 vmhost adapter
- **verify-later:** CONSOLIDATION_where_it_all_fits.md, PARALLEL_engine_deployment_and_layer5.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Phased plan P0–P5
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** plan(11) Phases: P0/P1/P2 done, P3 in progress ("Remaining for P3: land the chassis patch…"), "P4 … IN PROGRESS (this chat)", P5 not started.
- **what:** P0 structural decisions; P1 manual go-live (Path A); P2 wire deploy-on-update (two Actions); P3 make the probe a normal pipeline output (github_repo target selection + capture component + capability gate); P4 off-box collection + ranking; P5 registry + provisioning adapter.
- **sources:** traffic_probe_plan(11).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** contains most other concepts; earlier plan versions phrased P4/P5 differently
- **verify-later:** n/a

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### VM-Hosted Backend Sites class (proposed doc 024)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** plan(11) "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference; numbering operator's)".
- **what:** The genuinely-new infrastructure: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. The traffic probe is instance #1 of this class; future chat/board sections join it.
- **sources:** traffic_probe_plan(11).md#framework-integration, traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle
- **relations:** class parent of intent-probe; ties to D5 requires-backend gate
- **verify-later:** doc 024 existence; service_instances table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Pull architecture / no collector VM
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "No third 'collector' VM: the serving box buffers (JSONL); the CLUSTER pulls over key-gated HTTPS … Pull keeps every credential in the cluster — boxes never hold DB/cluster secrets".
- **what:** Collection is pull-only: the serving box buffers events in JSONL and exposes them via key-gated HTTPS (/events, /access-digest, /stats); the cluster's scheduled collector pulls. No third collector VM and no push, because push or a middle VM would put DB/cluster secrets on the box and add attack surface + a hop for no gain. B2 remains an optional cold backup.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#6
- **relations:** rationale for /events + collector design; boxes hold only the read-only stats_key
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk topology exception — static page + always-on backend, not pure edge
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md`/`(10).md` "Architecture (note: not edge-only)"; `running_notes(44).md`: "Topology note: idea.uk is NOT pure-static/edge like the other chat domains."
- **what:** Every other "simple paid chat" domain on the platform is designed as static-S3 + a synchronous edge worker (no always-on compute). idea.uk breaks this pattern because its "tool" is a minutes-long multi-LLM + web-search job, not a synchronous chat turn: it needs a small always-on backend running the engine as a background task, with the static/embedded page posting to it and Stripe's webhook pointed at it. Flagged repeatedly as a deliberate, understood exception to the platform's default serverless-edge model, not an oversight — "the PAGE is serverless..., the SERVICE is NOT and can't be."
- **sources:** `RUNBOOK_idea_uk(1).md` "Architecture" section; `running_notes(44).md` ("Topology note: idea.uk is NOT pure-static/edge")
- **relations:** idea.uk deployment topology; service-deployer pattern
- **verify-later:** contrast against the actual edge-worker chat domains for confirmation

<!-- SOURCE: U26_misc_dirs.md -->
### Agent chassis — generic configurable agent executor
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md describes it as the running framework ("deploys as a scalable Kubernetes deployment, 3 replicas in production"); HITL agent definition (2025-11-03) still references image `docker.io/aqls/agent-chassis:v1.0.407`.
- **what:** A single reusable Go binary that becomes any agent type via configuration: it consumes Kafka messages, loads its workflow config from the database (agent_definitions / agent_instances), executes the workflow, handles fuel checks, errors, metrics, and health endpoints. New agent types are created by adding DB configuration, not code — "you're not creating new CODE, you're creating new CONFIGURATIONS".
- **sources:** docs/architecture/002-agent-chassis-docs.md; docs/architecture/023-spawning-agents.md#the-core-concept; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; agent spawning; distributed embedded orchestration
- **verify-later:** cmd/agent-chassis/main.go, platform/agentbase/, platform/messaging/processor.go, agent_definitions table

<!-- SOURCE: U26_misc_dirs.md -->
### Distributed embedded orchestration (no central orchestrator)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md: "every agent is both a worker and an orchestrator... eliminates single points of failure"; presented as a completed architecture report.
- **what:** Every agent pod embeds a full orchestrator (SagaCoordinator) instead of a central orchestration service. Any pod of an agent type can start a workflow, and any pod can pick up a response and continue it, because state is in the shared database. Key architectural decision distinguishing the platform from Temporal/Airflow-style central schedulers.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#distributed-orchestration-model; docs/architecture/003-flow-doc.md; docs/architecture/012-investors.md
- **relations:** stateless-first principle; database-backed workflow state; AI-native orchestration positioning
- **verify-later:** platform/orchestration/ (SagaCoordinator), orchestrations/orchestrator_state tables

<!-- SOURCE: U26_misc_dirs.md -->
### Database-backed workflow state (orchestrator_state → orchestrations)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md: "The orchestrator is stateless... Workflow state is in the database"; basic_usage/004 queries both the old `orchestrator_state` and a newer `orchestrations` table, showing the concept live through a schema evolution.
- **what:** All workflow execution state — status (RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED), current_step, workflow_plan, execution_path, collected_data, awaited_steps/awaited_requests, final_result — is persisted per correlation_id. Responses arriving at any pod are matched to awaited steps via causation_id, the state is updated, and the workflow continues when all responses are in.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/architecture/003-flow-doc.md; docs/basic_usage/004_debugging
- **relations:** stateless-first principle; fan-out and response correlation; workflow state machine
- **verify-later:** orchestrator_state and orchestrations tables in clients DB; column set differences between them

<!-- SOURCE: U26_misc_dirs.md -->
### Local vs remote actions and the action registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md documents both patterns as implemented ("Executed within the orchestrator itself" vs "Executed by other agents via Kafka").
- **what:** Workflow steps are either local actions run synchronously in the orchestrator (validate_input, transform_data, spawn_agent, process_data...) registered in a Go actionRegistry, or remote actions dispatched to another agent's Kafka topic with state moved to AWAITING_RESPONSES. The registry grew over time (spawn_agent, execute_llm_prompt, await_approval added later).
- **sources:** docs/architecture/004-agent-chassis-architecture.md#local-vs-remote-actions; docs/architecture/025-reusable-evolvable-agent-teams#step-3; docs/basic_usage/003_dynamic_prompt_improvement
- **relations:** agent-centric call_agent; fan-out; HITL await_approval
- **verify-later:** platform/orchestration/actions/ directory; actionRegistry in coordinator.go

<!-- SOURCE: U26_misc_dirs.md -->
### Fan-out and awaited-response correlation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md walks a live fan-out (reasoning + image agents) with correlation_id/causation_id header matching; 001basic_usage.txt shows fan_out steps in the deployed website-builder workflow.
- **what:** A fan_out step sends parallel sub-tasks to multiple agent topics, records their request IDs in awaited_steps, and sets status AWAITING_RESPONSES. Each response carries correlation_id (workflow) and causation_id (the originating request_id); any receiving pod matches causation_id to an awaited step, stores the result under collected_data, and resumes when all are received.
- **sources:** docs/architecture/003-flow-doc.md; docs/architecture/004-agent-chassis-architecture.md#response-handling-flow; docs/basic_usage/001basic_usage.txt
- **relations:** database-backed workflow state; kafka topic conventions; message header contract
- **verify-later:** fan_out action implementation; awaited_steps vs awaited_requests handling

<!-- SOURCE: U26_misc_dirs.md -->
### Kafka topic conventions (process/responses → requests/responses)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Early docs use `system.agent.{type}.process` + `system.responses.{type}` (004); the stateless plan (v21) and HITL scripts (Nov 2025) use `system.agent.generic.requests`, per-type `.requests/.responses/.errors/.dlq` topics stored in agent_definitions.topics — the newer form names the older one's replacement.
- **what:** Naming scheme for per-agent-type Kafka topics plus system topics (`system.notifications.ui`, `system.commands.workflow.resume`, `system.errors.*`, DLQs). Topics are per agent TYPE, not per instance; all replicas share a consumer group so Kafka distributes work. The convention itself is durable; the specific `.process` form was superseded by `.requests/.responses`.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#kafka-topic-structure; docs/plans/stateless-first-agents-001#7-kafka-configuration; docs/humanintheloop/hitl_agent_definition.sql (topics JSONB)
- **relations:** stateless-first principle; HITL notification/resume topics
- **verify-later:** actual topic list in cluster; topics column of agent_definitions

<!-- SOURCE: U26_misc_dirs.md -->
### Message header contract (sender identity, in_response_to_*, status enum)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 (marked "v21", heavily iterated) defines the full header set; quick_hitl_test.sh (Nov 2025) sends live messages carrying orchestration_id, orchestration_name, step_name, message_type, from_agent_type, responses_topic headers — the contract in use.
- **what:** Rich request/response headers: sender AgentIdentity (agent_type, agent_id=pod name, version), correlation_id + human-readable correlation_name, orchestration_id/name, step_id/name, request_id, retry_version, parent orchestration linkage, message_id, fuel budget, timeout, routing topics. Responses echo in_response_to_request_id/step/orchestration and carry a status enum: awaiting | processing | complete | error_recoverable | error_unrecoverable, plus multipart flags, timing and fuel accounting.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/send_approval.sh
- **relations:** retry semantics; message deduplication; fuel budget
- **verify-later:** Go header structs in platform code; kafka message headers on live topics

<!-- SOURCE: U26_misc_dirs.md -->
### Stateless-first agent principle
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Core Principle: Agents are stateless executors. State lives in the database. Any replica can process any message for its agent type" — presented as the implementation spec (v21) that matches later operational docs.
- **what:** Agents hold no orchestration state in memory; pod crashes lose nothing; replicas scale horizontally with HPA (CPU + kafka consumer lag metrics); Kafka consumer groups distribute work; messages for one orchestration are ordered by using orchestration_id as the partition key. Formalises and extends the earlier distributed-orchestration model.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/plans/stateless-first-agents-001#8-kubernetes-deployment
- **relations:** distributed embedded orchestration; orchestration-as-identity; optimistic locking
- **verify-later:** deployment manifests (HPA config), consumer group setup, partition key usage

<!-- SOURCE: U26_misc_dirs.md -->
### Orchestration-as-identity model (AgentID = PodName)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Orchestration (in DB) + Step + Request = 'Agent Instance'... AgentID = PodName (changes on restart, but that's OK)". This resolved the earlier mandatory-AgentID debate (022).
- **what:** The persistent identity of "an agent doing a task" is the orchestration record, not the pod. Pod name serves as AgentID purely for debugging (processing_history records which pod handled each step). Supersedes the doc-022 proposal that workflows resolve and pin specific versioned agent instances (stable/canary selection strategies) — that instance-pinning design was not carried forward.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/architecture/022-possible-agent-structure#the-case-for-mandatory-agentid
- **relations:** supersedes mandatory agent-instance resolution (022); stateless-first principle
- **verify-later:** whether processing_history with pod_name exists in current orchestrations table

<!-- SOURCE: U26_misc_dirs.md -->
### Optimistic locking on orchestration state
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Fully specified in stateless-first-agents-001 (version column, update_orchestration_if_version() SQL function, retry loop with backoff) but no later doc in this unit confirms it shipped.
- **what:** Each orchestration row carries a version integer; replicas load state, apply a step, and save only if the version is unchanged (compare-and-swap), retrying on mismatch. Prevents two replicas from double-processing the same step. Paired with processing_history JSONB as the audit trail of which pod did what.
- **sources:** docs/plans/stateless-first-agents-001#3-database-backed-state-management; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** stateless-first principle; message deduplication
- **verify-later:** version column and update function in current schema; conflict-retry code

<!-- SOURCE: U26_misc_dirs.md -->
### Retry semantics: same request_id, incremented retry_version
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** stateless-first-agents-001 "Key Implementation Notes: Retry uses same request_id with incremented retry_version"; error_recoverable responses trigger up to 3 retries. No later confirmation in this unit.
- **what:** Failed remote calls are retried with the identical request_id and retry_version+1 so responses remain matchable and duplicates detectable. Recoverable errors retry (max 3), then fall through to unrecoverable which fails the orchestration and propagates an error to the parent. Progress statuses (awaiting/processing) are logged but never propagated upward; terminal states are processed exactly once.
- **sources:** docs/plans/stateless-first-agents-001#6-retry-logic; docs/plans/stateless-first-agents-001#key-implementation-notes
- **relations:** message header contract; message deduplication
- **verify-later:** retry handling in response processing code; awaited_requests (request_id, retry_version) PK

<!-- SOURCE: U26_misc_dirs.md -->
### Message deduplication (processed_messages, terminal-state-once)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Designed in detail (dedup key request_id:retry_version:status; processed_messages table with 24h cleanup) in stateless-first-agents-001; no operational evidence in this unit.
- **what:** Before processing, agents check a dedup key against a processed_messages table (or in-memory map); duplicate responses are dropped, and once any terminal state (complete/error_unrecoverable) is processed for a request, all further terminal responses for it are ignored. Ensures idempotency under Kafka redelivery and multi-replica consumption.
- **sources:** docs/plans/stateless-first-agents-001#7-deduplication-handler; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** retry semantics; optimistic locking
- **verify-later:** processed_messages table existence; dedup logic in message consumption path

<!-- SOURCE: U26_misc_dirs.md -->
### Fuel budget resource management
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md lists fuel management as a chassis feature ("Checks fuel budget from headers... Prevents execution if insufficient fuel"); fuel_budget=1000 header sent in live kcat commands (basic_usage 001/004); response headers carry FuelUsed/RemainingFuelBudget.
- **what:** Every workflow carries a fuel budget header; actions deduct fuel costs; sub-invocations pass a reduced budget down and report fuel used back up the chain. Serves as the cost/abuse control across multi-agent workflows. Current status in the 2026 platform unverified — no recent doc in this unit mentions it.
- **sources:** docs/architecture/002-agent-chassis-docs.md#key-features; docs/basic_usage/001basic_usage.txt; docs/plans/stateless-first-agents-001 (FuelUsed/RemainingFuelBudget headers)
- **relations:** message header contract; subscription/quota API
- **verify-later:** fuel handling in chassis code; whether current work-item system retains fuel

<!-- SOURCE: U26_misc_dirs.md -->
### Agent-centric architecture: steps call agents, not topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 022-possible-agent-structure proposes `call_agent` with agent_type replacing raw topics ("your current code already does 90% of this"); 027 and the HITL group definition (Nov 2025) use call_agent steps in production seeds.
- **what:** The primary abstraction is the agent (owning a 6–12 step workflow) rather than the workflow; steps invoke other agents (`action: call_agent, agent_type: X`) which have their own workflows, error boundaries and state, enabling recursive hierarchies (any agent can orchestrate, a copywriter can spawn a researcher). Topic resolution happens from agent type.
- **sources:** docs/architecture/022-possible-agent-structure#summary-agent-centric-architecture; docs/humanintheloop/hitl_agent_group_definition.sql; docs/architecture/023-spawning-agents.md#the-orchestrator-is-a-pod-too
- **relations:** agent chassis; agent spawning; supersedes inter-agent invocation protocol v1
- **verify-later:** call_agent action and agent-type→topic resolution code

<!-- SOURCE: U26_misc_dirs.md -->
### Inter-agent invocation protocol v1 (invoke_agent / agent_invocations)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** 001-agent-calls-agents-doc.md proposes InvokeAgentAction, ParallelInvokeAgentsAction, and an agent_invocations tracking table; later docs (022, stateless plan) replace this with call_agent + orchestration hierarchy headers, and the agent_invocations table never reappears.
- **what:** The first design for agent-calls-agent: a dedicated invocation request/response envelope, per-pair topics (`system.agent.requests.{from}.{to}`), an agent_invocations audit table, and parent_correlation_id columns. Its essential ideas (parent linkage, deadline, fuel passing) survived into the header contract; the specific mechanism did not.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#1.2; docs/architecture/001-agent-calls-agents-doc.md#phase-3
- **relations:** superseded by call_agent (022) and stateless header contract; project manager agent
- **verify-later:** confirm agent_invocations table absent from schema

<!-- SOURCE: U26_misc_dirs.md -->
### Project Manager / User Representative agent hierarchy
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Designed across 001 and 007 ("User Representative Agent... represents the users views against the project manager"); never appears in later seeds, groups, or the current 002-spine — silently vanishes after the website-builder group takes the orchestrator role.
- **what:** A top-level persona hierarchy: User → Project Manager agent (plans phases, delegates to specialist orchestrators, reviews deliverables) → Web Design Orchestrator → specialists, with a User-Persona agent negotiating on the user's behalf (stores preferences, approves/rejects deliverables). The review/approval intent resurfaced later as HITL steps and content governance instead.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#architecture-overview; docs/architecture/007-roadmap.md#2.1-user-representative-agent
- **relations:** website-builder group (took its place); HITL approval mechanism (absorbed the review role)
- **verify-later:** confirm no project-manager/user-representative agent_definitions exist

<!-- SOURCE: U26_misc_dirs.md -->
### HTML-first progressive enhancement delivery
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 008: "Starting with plain HTML/CSS/JS is actually a very smart architectural decision"; the html-developer seeds specify "vanilla" HTML with inline CSS; the current platform still renders plain HTML/CSS sites (render pipeline docs).
- **what:** Deliberate decision to generate plain HTML/CSS/JS websites rather than framework apps: easier for AI to generate and validate, no build step, universally hostable, fast; complexity added progressively (web components → PWA → framework only if needed). One of the few strategy decisions from this era that demonstrably survived into the present render pipeline.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#why-simple-html-css-js-is-the-right-start; docs/architecture/027-create-website-creation-system (html-developer config)
- **relations:** styling-render-pipeline (current successor context); WordPress export agent (rejected sequel)
- **verify-later:** current renderer output format

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow status state machine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Consistent across eras: RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED in 004; RUNNING / AWAITING_RESPONSE / COMPLETED / FAILED in HITL_README (Nov 2025); pending|processing|complete|failed variant in the stateless plan.
- **what:** The orchestration status vocabulary and its transitions: workflows run steps, park in an awaiting state while remote/human responses are outstanding, and terminate complete or failed. The HITL pause reuses the same awaiting state rather than introducing a special paused status. Minor naming drift across eras is itself a verification target.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/humanintheloop/HITL_README.md#workflow-states; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** database-backed workflow state; HITL approval mechanism
- **verify-later:** canonical status enum in current schema/code

<!-- SOURCE: U26_misc_dirs.md -->
### Human-readable orchestration and correlation names
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 mandates orchestration_name / correlation_name alongside UUIDs ("website-build-agrivisionary", "core-mgr-website-flow-0902-1030"); start_hitl_workflow.sh generates ORCHESTRATION_NAME="eborg-content-approval-$(date...)" in practice.
- **what:** Every orchestration and correlation carries a generated human-readable name in addition to its UUID, propagated through headers and stored in state, so debugging and monitoring read as narrative ("which pods processed core-mgr-website-flow") rather than UUID archaeology.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/start_hitl_workflow.sh
- **relations:** message header contract; kcat/db-inspector runbook
- **verify-later:** name-generation code; name columns in orchestrations table

<!-- SOURCE: U26_misc_dirs.md -->
### Agent teams: composite/family/service-agent patterns
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** 021 evaluates three options (PM pattern, peer-to-peer squads, service-oriented) and recommends service-oriented-then-squads; what actually shipped was the simpler agent-groups + call_agent model — the AgentFamily/SharedMemory and workflow-composition (sub-workflow) constructs never reappear.
- **what:** Design exploration for complex 50+-step workflows: composite agents (one external face, embedded sub-components), agent families with shared state and peer coordination, stateless reusable service-agents (date extractor, entity extractor) callable by any workflow, and workflows-invoking-workflows composition. Records the acknowledged framework limitation ("one agent = one workflow, flat orchestration, no concept of agent teams") that agent groups later addressed.
- **sources:** docs/architecture/021-current-framework-limitations; docs/architecture/022-possible-agent-structure
- **relations:** agent groups (the shipped resolution); agent-centric architecture
- **verify-later:** n/a — confirm no AgentFamily/sub-workflow constructs in code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka topics model: generic entry point vs per-spawn job topics vs fixed adapter topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) topics section describes current mechanics incl. stable identity naming
- **what:** Three patterns: `system.agent.generic.requests` (shared entry door for callers with no spawn relationship — scheduler, manual triggers; explicitly "the current door", expected to evolve into a formalised entry API); `job.<corr8>-<orch8>-<type>-<step>.requests` per-spawn topics set by setupAgentTopics; fixed `system.agent.<type>.*` topics for long-lived adapters. agent_definitions.topics jsonb is declarative only — the Deployment manifest actually subscribes.
- **sources:** 001_development_guide(5).md#Topics; 002(4)#Infrastructure
- **relations:** wrapper-orchestrator; idle timeout; topic cleanup
- **verify-later:** setupAgentTopics/createTopics in spawn_actions.go; deployment manifests

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Agent message structure and HITL response shape
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) reference section
- **what:** Messages have Kafka headers (correlation/orchestration/request ids, message_type, action, responses_topic, sender) and a body of `headers`/`config`/`input_data`. Agents always reply to the caller's responses_topic. HITL responses go to the agent's responses topic with `in_response_to_request_id` from awaited_requests and `sender_agent_type: human`.
- **sources:** 001_development_guide(5).md#Agent Message Structure
- **relations:** adapter response envelope contract; awaited_requests
- **verify-later:** MessageProcessor, awaited_requests table

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Orchestration state and collected_data as the workflow data bag
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5)/016 §6.0 describe live mechanics
- **what:** Each orchestration is an orchestration_states row (workflow_plan, collected_data, current_step, status). Steps communicate solely via dotted paths into collected_data; agents themselves are DB rows (`agent_definitions.default_config.workflow`), not Go types — the Go codebase contains actions, not agents. "Every agent is an orchestrator" is literal.
- **sources:** 001_development_guide(5).md#Orchestration State; 016 §6.0 "What an agent actually is"
- **relations:** loop mechanisms; workflow result contract
- **verify-later:** orchestration_states schema; coordinator.go continueExecution

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5): "the #1 cause of pipeline stalls"; design with A/B/C classification
- **what:** Timeout goroutines die with pods, leaving AWAITING_RESPONSES orchestrations stuck. A periodic 60s DB sweep on chassis pods (FOR UPDATE SKIP LOCKED) classifies expired awaited requests: child completed → synthesize response; child failed → forward; none/running → retry up to 3 then fail parent. Handles topic-expired case by directly advancing parent state.
- **sources:** 001_development_guide(5).md#Stale Orchestration Sweeper
- **relations:** spawn-handler hang (timeout_at not enforced); claimed-item-timeout
- **verify-later:** sweeper implementation; stale-orchestration-reaper scheduled task

<!-- SOURCE: U01_docs024_numbered_core.md -->
### site-work-orchestrator ordering and the dispatch pattern
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) workflow and resolved decisions 16–21
- **what:** Unified orchestrator processes pending items → verifies previous → runs due discovery → triages → updates profile. Dispatch = spawn→call per item with raw identifiers; orchestrator never pre-spawns static chains, never derives handler data, never passes work-item awareness to handlers (handlers self-contained, CLI-callable). Status tracking stays in the loop.
- **sources:** 002(4)#The orchestrator, #Dispatch pattern, #Resolved Decisions
- **relations:** build-dispatch-loop; handler contract
- **verify-later:** site-work-orchestrator definition currency vs build-dispatch-loop

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Entity data model (state-based lifecycle, news triggers, client-side real-time)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** 002(4): tables site_entities/site_entity_relationships "exist", entity_sources/entity_sync_log "planned"; Phase 3 item
- **what:** Structured data that generates pages (events, performers, venues, ticket tiers) with state-based lifecycle (announced→on_sale→…→historical), setup mode (work items) + discovery mode (scheduled sync), significant state changes triggering news via entity_sources.news_triggers; real-time data (prices, availability) served client-side from a data API, never through the work queue.
- **sources:** 002(4)#Entity Data Agent Family; #Site Type Stress Tests (events/boxing)
- **relations:** news feed pipeline; site API router (007)
- **verify-later:** site_entities usage; any entity-data-agent definition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Maintenance profile per-site configuration
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) JSON shape; audit config consumed by improvement loop
- **what:** sites.settings.maintenance_profile controls per-domain cadence (content/links/seo/compliance/content_feed/entity), budgets (llm_calls_per_cycle, max_auto_fixes_per_cycle), build tier, and audit group enablement; audit_pass_count also lives here.
- **sources:** 002(4)#Per-Site Configuration; 002d#Site-Type-Specific Audit Configurations
- **relations:** audit pass cap; growth budget (separate: site_specs growth_config)
- **verify-later:** sites.settings shape in production

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Idle timeout for spawned agents + topic cleanup strategy
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4): mechanism, config column, sync.Once shutdown safety, tuning SQL
- **what:** agent_definitions.idle_timeout_seconds → env var → idle-monitor goroutine exits the pod after inactivity (0 = forever for deployments). Topics: EPHEMERAL_TOPICS per-spawn today; agents never clean up topics — a conservative 10-min CronJob deletes topics with no matching pod, Kafka 7-day retention as backstop; a future shared-topics design (pre-created per-type topics, header routing, static group membership) makes cleanup a no-op.
- **sources:** 002(4)#Idle Timeout, #Shared Topic Strategy, #Topic Cleanup Design
- **relations:** pod accumulation debugging (016 §1)
- **verify-later:** idle_timeout_seconds values; topic-cleanup CronJob

<!-- SOURCE: U01_docs024_numbered_core.md -->
### business-intel shared-pod pattern (multi-type agents on one static pod)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 017 architecture section; ai_service placement rule
- **what:** Multiple agent definitions share one static pod via message routing (config.agent_type → selectWorkflow/FindBestGroup); consequence: ai_service must live in STEP config, not agent_config top-level, because agent_config comes from the pod's own type. Workflows are minimal action→complete with logic in Go. Single-replica contention accepted for batch work.
- **sources:** 017#Deployment, #ai_service on Shared Pods
- **relations:** wrapper-orchestrator (contrast); med pipeline (same pod)
- **verify-later:** business-intel deployment manifest

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### CollectedData: single-channel orchestration working memory and its pathologies
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Analysis only. No code changes proposed yet" (2026-05-11); duplication called "structural", observed in every log
- **what:** CollectedData (orchestration_states.collected_data JSONB) is the single channel for step outputs, routing metadata, loop variables and parent-reply context — "the most overloaded data structure in the system". Documented pathologies: recursive `__raw_message__` nesting (write amplification ×15 optimistic-lock retries), dual storage at step_name AND output_field, InitialRequestData/__raw_message__ overlap, six conflated data categories in one flat namespace, loop iteration data stored 3-4×, CleanDataMap stripping legitimately-named response fields. Recommendations R1–R6 (strip system keys from __raw_message__, pick one storage key, namespacing, loop GC, delta writes) proposed, untriaged.
- **sources:** FOCUS_collected_data_analysis.md (whole)
- **relations:** flat-namespace collision risk; compensating mechanisms; consumer-group race (duplicate keys as evidence)
- **verify-later:** BuildCollectedData / storeActionResult in coordinator.go; whether R1 ever landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Flat-namespace collision risk and the compensating-mechanism accretion
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** dev-guide-documented incident: "section-editor declared content_data as optional and the nested-source loop silently lifted site_record.content_data … and overwrote a hero section" (referenced 2026-05-11)
- **what:** Because caller inputs, step results and site context share one flat map, actions can silently pick up `site_record.site_id`/`content_data` instead of caller-supplied fields. The framework compensates with UnwrapDeep, FindByPath prefix fallbacks, extractReplyToMetadata 3-tier priority, output_mapping — accreting workarounds faster than it consolidates. New code should use collision-free names (target_site_id convention); existing code left alone.
- **sources:** FOCUS_collected_data_analysis.md#4.4, #5; ASSESSMENT_imagery_phase_0_1…md#Caveat-1
- **relations:** CollectedData analysis; ExtractActionInputs conventions
- **verify-later:** ExtractActionInputs nested-source loop behaviour for undeclared fields

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Response-topic consumer group race (per-pod groups fan out every response)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "Discovery, not yet remediated" (2026-05-10); ~85 consumer groups on system.agent.generic.responses, only 3 live; two pods ran ProcessResponse on the same message 215ms apart
- **what:** The requests topic uses a shared stable consumer group but each chassis pod joins the responses topic under its own per-pod UUID group, so every response is delivered to every pod; each independently advances orchestration state, and the loser of the version race can flip a step to FAILED (observed on call_logo_gen). Mostly silent (idempotent writes) but structurally wrong; the system relies on shared-pool semantics it doesn't have. Open questions: intended model, per-spawn job.* topic groups, CAS hardening in ProcessResponse, 82 stale groups cleanup.
- **sources:** ANALYSIS_chassis_response_consumer_group_race.md (whole)
- **relations:** dispatcher stall Bug 1; duplicate collected_data keys; Phase 2F migration testing blocked
- **verify-later:** AgentClient constructor wiring (consumerGroup argument); ProcessResponse CAS behaviour in coordinator.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Kafka empty partition assignment on simultaneous pod join
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "five agent-chassis pods were members of generic-requests-group but all showed #PARTITIONS: 0 … Fix applied: Delete one pod to force rebalance" (2026-04-20)
- **what:** After a deploy where all pods join within the same second, the group can go Stable with the partition unassigned — zero consumption while offsets pile up and work items sit triaged. Workaround: kill a pod. Watch item on every deploy: at least one member must show #PARTITIONS: 1.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#1
- **relations:** consumer-group race; dispatcher reliability
- **verify-later:** whether staggered restarts or a fix was adopted

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Observability gaps: owner_agent_type "generic" and orchestration_name
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** "orchestration_states rows where the generic agent routed to a different workflow still show owner_agent_type = 'generic'" (2026-04-20, P3)
- **what:** When the generic chassis routes a message to another agent's workflow (FindBestGroup), the orchestration is filed under owner_agent_type='generic' and orchestration_name doesn't carry the scheduler's sched-<task> name — searches by agent type or task name find nothing, which caused the "trigger never runs" misdiagnosis. Fix shape: selectWorkflow sets owner_agent_type to the resolved type.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7, items 7-8
- **relations:** content-feed-trigger bug; execution_path not populated (flywheel note 2.4c)
- **verify-later:** selectWorkflow in processor.go

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Four overlapping chrome default stores and the update_site_defaults linkage
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Notes (Sh) F-section: intended chain documented in 003's Site Component Linkage Contract; SPEC W4(c): "keep function-lookup as the norm … Smallest honest change: the comment" — the linkage deliberately left unrepaired.
- **what:** Header/footer defaults coexist in four stores: `style_collections.header/footer_component_id` (the operative read, dead-NULL), `site_components` slots (copy target + pre-render cache; idea.uk's were pinned to inactive components), `sites.default_components` JSONB (UpdateSiteDefaultsAction's target — a tracking copy nothing reads on the render path), and `layouts.default_*_component_id` (FK, all NULL, nothing copies it onward). The intended chain — style_collections as source of truth, `update_site_defaults` copying into site_components — never runs in composition (003's documented failure mode #1 IS idea.uk's case). The fix chose function-lookup as the norm rather than reviving the chain; populating style_collections at install remains a possible per-site-variety feature.
- **sources:** running_notes_scheme_to_components(55).md#Sg #Sh; SPEC_scheme_to_components.md#W4; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3b)
- **relations:** chrome selection path; Q6's original layouts.default_* direction (superseded by this resolution); site_components repoint.
- **verify-later:** v3_site_actions.go UpdateSiteDefaultsAction; whether the misleading install comment was deleted; any later population of style_collections chrome ids.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Orchestrator-agent architecture conventions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Notes (Um) context reset restates them as standing rules: "every agent is an orchestrator owning a workflow of ≥1 steps that call ACTIONS. Children respond to the PARENT's responses_topic … Do NOT create sub-workflows in SQL — spawn sub-agents."
- **what:** The platform's structural conventions, carried as strict constraints through both threads: every agent is an orchestrator owning a workflow of steps that call Go actions; workflows stay simple with complexity in action code; no sub-workflows in SQL — spawn sub-agents with their own workflows (clean logs, separated responsibilities); child agents respond on the parent's responses topic; workflow variable names stay in sync with action expectations; identifiers are never renamed silently; `logger.Debug` is banned (invisible in the log pipeline — use Info); reuse/alter existing functions and architecture before creating new.
- **sources:** running_notes_scheme_to_components(55).md#Architecture-conventions #Um; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** house rules / standing preferences; platform mission.
- **verify-later:** agent-creation guidelines doc in repo; logging config that swallows Debug.

<!-- SOURCE: U04_idea_uk.md -->
### Coordinator result-extraction contract (resolveResultSpec) and the silent-stub class
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Coordinator result-extraction fix — field-validated 2026-06-19 (idea.uk index built + deployed)"; git archaeology settled the cause to commit 06a8c6ef (14 Jan), unchanged since.
- **what:** The class of bug and its structural fix. Bug: workflow `complete` steps declaring singular `output_field` were never honoured (coordinator read plural `output_fields` only since 14 Jan), falling back to a working-state dump; when a big multi-section page's dump cleared the 900k `MaxResultSizeBytes` cap, `extractMinimalResult` returned a `status:"completed"` **stub** — silent false completion (gamesdesign) or, where the claimed-item evidence gate refuses 0-component pages, honest claim-timeout failure (idea.uk's empty index). **Size was the trigger; the singular key necessary-but-not-sufficient** (bucket audit: 100 plural steps safe, ~59 dump-bucket agents fine because small, 4 singular — only the writer breaks). Fix: centralised result contract in `result_spec.go` — singular→FLATTEN (named field's contents become the response body), plural→FIELDS (unchanged), `output`→MAPPING (previously silently dumped), none→dump; completion metadata via setIfAbsent; **oversize is now a loud error** routed to notifyParentOfFailure with a per-field size breakdown; stub removed; deprecated keys alias to result_from/multiple_output_fields/result_mapping. Retest doctrine: requeue the one failed page, do NOT re-adopt.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Empty index diagnosis + fix-direction 1 + retest + git section); idea.uk/running_notes(63).md (ff–qq checkpoints); idea.uk/README_001_todo_list.md
- **relations:** claimed-item evidence gate; debugging trap "inferring writers from readers"; MaxResultSizeBytes guardrail (do not raise).
- **verify-later:** platform/orchestration/result_spec.go + coordinator.go; the mode=flatten log on a writer run.

<!-- SOURCE: U04_idea_uk.md -->
### Hosting split: static-serverless front + small always-on backend ("static front + small back end")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Architecture doc §3 and the live topology (page embedded in the binary on the VM; B2 for everything chassis-built).
- **what:** The hosting taxonomy idea.uk established for the platform: pure-static content sites are serverless on B2; anything running a minutes-long multi-LLM job with a payment webhook cannot be serverless or edge-shaped — it needs a small always-on service with a stable inbound address. The classifier's `build_approach: hybrid` / `hosting_trajectory: needs_server` fields are the framework's slot for this distinction (noted as not yet confirmed in the classifier output). This is the hinge Layer 5 eventually automates, and why the engine can never be a forked client-side tool component.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#3; idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 2 hosting reality); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (reuse section)
- **relations:** Layer-5 wrapper; VM cutover; tool-library boundary.
- **verify-later:** where build_approach/hosting_trajectory actually live (strategist? architecture doc concept only?).

<!-- SOURCE: U05_content_quality_linking.md -->
### Result-contract resolution (resolveResultSpec: flatten/nest/mapping)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(44): "FIX DEPLOYED to production 2026-06-18 … Part 1 re-confirmed healthy in prod (06-21, stub_rows=0)".
- **what:** The chassis coordinator's extractWorkflowResult honoured only plural `output_fields` since commit 06a8c6e (2026-01-14); a child declaring singular `output_field` (e.g. page-content-writer) fell to a state-dump fallback, exceeded the 900k cap, and was replaced by a stub still reporting status:"completed" — silently dropping the compiled page. Fix: a centralised resolveResultSpec — singular output_field/result_from → FLATTEN, plural → nest, output/result_mapping → applied, deprecated names read with a Warn (deprecation census drives a later rename migration). Pure chassis change, verified against all live consumers (flatten corrective for writer→3 parents, site-planner, model-trainer).
- **sources:** NOTES_gamesdesign_silent_norebuild(44).md#root-cause + Plan Step 1; resolveResultSpec.go.orchestrator_patch.txt; RUNBOOK_gamesdesign_index_rebuild(29).md#context
- **relations:** oversize fail-loud hardening; silent-completion family; select_sections mapping follow-up; content-reviewer auto_eval repoint (latent follow-up).
- **verify-later:** platform/orchestration/result_spec.go; coordinator.go extractWorkflowResult; deprecation Warn census in logs.

<!-- SOURCE: U05_content_quality_linking.md -->
### Oversize-result fail-loud hardening (stub removed)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(44) follow-up 1 "DONE (implemented 2026-06-18 … deployed)"; README_some_notes.md describes the intended behaviour change.
- **what:** extractWorkflowResultWithSizeLimit no longer truncates strings or emits the status:"completed" stub (extractMinimalResult deleted); on oversize it returns oversizeResultError with a per-field size breakdown naming the largest field, and notifyParentOfSuccess converts to a loud CHILD_ORCHESTRATION_FAILED + agent_error_log entry. Chosen fail-loud over persist-and-reference and over recursive trimming ("truncating content delivered as success is a corrupt result"). Any agent previously stub-"succeeding" now fails loudly until it declares a result contract — the surfacing is the point.
- **sources:** NOTES(44) Follow-ups; README_some_notes.md; RUNBOOK_gamesdesign_index_rebuild(29).md#6
- **relations:** result-contract resolution; MaxResultSizeBytes guards the Kafka ceiling (do NOT raise it).
- **verify-later:** coordinator.go oversizeResultError; agent_error_log for delivery-cap entries.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### sites.status is an informational lifecycle label (validated vocabulary, no consumer filters)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "No on-disk code filters sites on status — it is an informational lifecycle label; build dispatch keys on site_work_items" (RUNBOOK_scheme(18) sites.status RESOLVED, from v3_site_actions.go:323–395).
- **what:** UpdateSiteStatusAction validates status ∈ {draft, building, review, published, deployed, archived, error} (and stamps last_deployed_at with status=deployed); 'active' and 'system' are legacy out-of-vocabulary values on old rows. Nothing filters sites by status at build time — an assumption (`WHERE s.status='active'`) borrowed from an old handoff silently wrecked a blast-radius count, hence the standing rule: never filter on status='active'.
- **sources:** RUNBOOK_scheme_to_components(18).md §sites.status RESOLVED; running_notes(22).md Sr, Ss
- **relations:** work-item dispatch (the real gate); needle-gate/verify-at-point-of-use discipline.
- **verify-later:** UpdateSiteStatusAction vocabulary; legacy-status rows.

<!-- SOURCE: U08_travelling_docs.md -->
### Aspiration: agent-creation and inter-agent message logging workstream
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES "Note on a separate, out-of-scope item" — "kept out of these docs to preserve separation of concerns. Can be specced separately." No later mention.
- **what:** A stated desire to closely log/track agent creation and inter-agent messages (headers + body) as a distinct workstream from travelling docs — different responsibility and data. Never designed or built within this unit's horizon.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#out-of-scope-note
- **relations:** travelling docs (deliberately separated); message envelope standards (035).
- **verify-later:** whether any message-logging workstream exists elsewhere in docs/.

<!-- SOURCE: U09_adoption.md -->
### pages.sections is the build-read field (load_page_sections_from_spec fallback chain)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Decisive code facts established… load_page_sections_from_spec reads site_specs.site_plan (absent for this site) → falls back to pages.sections. pages.sections is the build-read field; site_plan_sections is NOT on the build path" (HANDOFF_2026-06-09, from the chassis dump).
- **what:** The page build resolves its section list via `site_specs` aspect `site_plan` (syncing into pages.sections when present) → `pages.sections` fallback → empty. `site_plan_sections` is relational plan hygiene only. `plan_sections` computes `ready_count` from resolvable section names; empty list → early return ready_count=0. Consequence: manual repairs must write pages.sections (the load-bearing statement), and a re-sync re-derives pages.sections from the in-memory plan (`extractPagesFromPlan`), so manual fixes survive only until the next full re-plan.
- **sources:** running_notes_15(10)#part-1–2, CONTEXT_PACK_adoption_skinner_box.md, HANDOFF_2026-06-06
- **relations:** sectionless durability; upsertPage EXCLUDED.sections; write_site_plan writes both tables
- **verify-later:** load_page_sections_from_spec_action.go source order; plan_sections_action.go ready computation

<!-- SOURCE: U09_adoption.md -->
### Chassis config location bugs (max_tokens shadowing, temperature single-read, error_step dead field)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** "max_tokens workaround applied for site-adoption-agent; error_step fix deployed via Path A (struct field + reader fallback); temperature TODO outstanding" (2026-05-18).
- **what:** A family of author-intuition vs chassis-reader mismatches: (1) `ExecuteLLMPromptAction` resolves the whole ai_service object once — a top-level ai_service shadows step-level config even when missing the field, so site-adoption-agent's analyze_site fell to the hardcoded 2048-token Anthropic fallback and truncated 8 of 20 pages (workaround: top-level max_tokens=16000); (2) temperature is read only from the very top of default_config — 6 step-level settings are dead and `llm_call_log.temperature` is NULL for every call (observability gap unresolved); (3) `error_step` at step level was silently dropped because the Step struct had no field — 62 dead graceful-degradation paths across 18 agents; Path A fix (ErrorStep struct field + config fallback) deployed, turning e.g. a firecrawl CSS timeout from fatal adoption failure into fallback-to-analyze_site with a partial fingerprint. Proposed structural fix: per-field resolution chains, raise the 2048 floor, log actual sent options.
- **sources:** FOCUS_chassis_config_location_bugs.md, old2/FOCUS_step_level_llm_config_ignored(3).md
- **relations:** 023 llm-quality-testing per-step model swap breaks on shadowed agents; fetch_primary_css hard-fail trade-off changed by error_step fix
- **verify-later:** execute_llm_prompt_action.go lines 110–145/209/213-218; contracts.go Step.ErrorStep; llm_call_log temperature population

<!-- SOURCE: U09_adoption.md -->
### work-site-orchestrator (monolith) vs build-site-planner (thin planner)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README maps every monolith inline step to its new-path equivalent; "the thin-planner model is the one that matches your stated philosophy"; design/imagery gaps closed per the mapping; site-status advancement flagged "worth confirming".
- **what:** Two architectures, not two versions: the old monolith did plan/write/sync/nav/design/render/deploy inline in one agent; build-site-planner is `read_specs → plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan (→ emit_design → emit_imagery) → complete`, delegating everything else via work items to handler agents. The audit question is "is every monolith inline piece now emitted as a work item on the build path" — answered yes for design/composition/rerender/JS once emit_design landed; open: no clear terminal step advances `sites.build_status` to deployed/active (sites may sit `pending` indefinitely).
- **sources:** README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** doc 029 Phase 1 (planner stops emitting work items); emit_design/emit_imagery; pageflow-builder (another still-active plan-less caller)
- **verify-later:** what sets sites.build_status terminal on the build path; whether work-site-orchestrator is still invocable

<!-- SOURCE: U10_imagery.md -->
### Snapshot-shadowing agent-definition loader defect
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Snapshot-shadowing hypothesis confirmed via SQL… Two loaders patched… Builds deployed as v1.0.1006"; "Snapshot audit closed clean" (2026-05-12).
- **what:** `snapshot_agent()` inserts snapshots at version+1000, so any loader using `ORDER BY version DESC LIMIT 1` without filtering `is_snapshot`/`is_active` reads the snapshot instead of the active row — every snapshot silently shadows its agent until the loader is fixed. `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` were patched; other loaders were already correct. Structural residue: snapshot retention policy and a single AgentDefinitionRepository remain open.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-1, STATUS_imagery_2026-05-12.md#Loader-snapshot-defect
- **relations:** model-infrastructure snapshot/rollback (021_model_swap_and_rollback.sql); parked "deep discussion" trio with the consumer-group race.
- **verify-later:** grep `FROM agent_definitions` across Go for is_snapshot filters; snapshot row counts.

<!-- SOURCE: U12_docs024_archives.md -->
### Quality Assurance Agent Architecture — folded into system-architecture, not abandoned
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded (as a standalone numbered doc; the architecture itself is deployed/partial)
- **status-evidence:** The standalone `002d_quality_assurance_architecture.md` (older1) and its later revision `002de_quality_assurance_architecture_v2.md` (archive_april_26) both appear verbatim as a "# 002d — Quality Assurance Agent Architecture" section inside live `002_system_architecture(4).md` (starting at line 897), continuing the main doc's Resolved-Decisions numbering (18-25) and extending it with a new "Layer 0: Pre-Generation Data Triage (plan_sections)" section, a "Content Validation as a Third Mode" table, and two further resolved decisions (24 "Quality gates before generation, not just after", 25 "needs_human_review is a first-class status") absent from any archived 002d draft. Its "Responsibility Boundaries" table was also updated to match the later composition/design-planner split.
- **what:** A three-layer QA model: Layer 1 structural/algorithmic checks (free, no LLM), Layer 2 LLM-assisted design/content audit (grouped agents sharing context, one LLM call per group), Layer 3 LLM-required strategic review (dream-spec gap analysis); plus a later-added Layer 0 pre-generation data triage (`plan_sections`). Includes the "promotion pattern" (a check starts as a `query_database` action step and is promoted to a spawned sub-agent only once it needs multi-step workflows or external calls) and the rule that audit agents "enforce, not override" the classifier/planner's stated intent. This was never a genuinely dropped concept area — it was consolidated into the numbered `002_system_architecture` doc rather than kept standalone, and then actively extended.
- **sources:** old/older1/002d_quality_assurance_architecture.md (whole file); archive_april_26/002de_quality_assurance_architecture_v2.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"002d — Quality Assurance Agent Architecture" (line 897+)
- **relations:** design agent responsibility split (site-design-planner/webdesign-agent); improvement-loop (004); site-spec-and-classifier (021); triage drain loop
- **verify-later:** confirm `design-audit-agent`, `visual-design-auditor`, `content-quality-auditor`, `site-review-agent` agent_definitions still implement the three-layer split; confirm `plan_sections` and `needs_human_review` status are implemented as described.

<!-- SOURCE: U12_docs024_archives.md -->
### site_work_items work-routing column renamed domain → pipeline
*(merged from 2 independent findings)*
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Live bug-log entry #18 in `001_development_guide(5).md`: "The `domain` column on site_work_items was renamed to `pipeline` in a migration."
- **what:** Two archived dev-guide drafts each devote a "Lessons Learned" section to `site_work_items.domain` being an internal work-routing namespace ("build"/"maintenance"/"marketing") that collides confusingly with the website's actual domain (e.g. "gaswholesalers.com") — citing real bugs this caused (a dispatch-loop filter mismatch, and a CSS-generation item never dispatching because it was written with `domain:"design"` instead of `domain:"build"`). Rather than keep relying on documentation warnings, the column was renamed to `pipeline` at the schema level, eliminating the ambiguity outright; the live doc drops the explanatory section entirely in favour of a terse bug-log line.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Work item domain is NOT the site domain"; old/001_development_guide.md#"Work item domain is NOT the site domain"; docs024_key_docs_latest/001_development_guide(5).md#18; old_design_and_styling/016_debugging_guide_v2.md#"Schema reminder"
- **relations:** dispatch-loop input_mapping; site_work_items table
- **verify-later:** confirm `site_work_items.pipeline` column exists in the current schema and no code still reads/writes `domain` for this purpose.

<!-- SOURCE: U12_docs024_archives.md -->
### Dispatch-loop input_mapping path mismatch (spec-nested vs flat)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Documented as "most common systematic failure" with three named affected agents (`tool-improver`, `tool-auditor`, `rerender-pages`); not confirmed whether the flatten-in-dispatch-loop fix or per-handler fix was adopted.
- **what:** `build-dispatch-loop` maps a work item's `spec` JSONB as nested (`input_data.spec.component_id`), but handlers read flat (`input_data.component_id`), producing path-resolution errors. Preferred fix: flatten in the dispatch loop's `input_mapping`, following the existing `page_name?`/`reviewed_brief?` pattern.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; ExtractActionInputs cross-link
- **verify-later:** current `build-dispatch-loop` `input_mapping` config.

<!-- SOURCE: U12_docs024_archives.md -->
### Site-work-orchestrator dispatch loop — asset self-resolving storage URI
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive documents specific Go additions (`PresignedURLToS3URI`, `resolveStorageURIFromAsset`) and a full finetuning.uk worked dispatch trace; none of this Go-function-level detail or the worked example appears in live `002_system_architecture(4).md`'s "Dispatch Loop... (from 004_site_work_orchestrator)" section, which keeps only the abstract principles.
- **what:** When the dispatch loop's discovery-written work items carry presigned HTTPS asset URLs but `deploy_image_asset` needs `s3://` URIs, the fix was to have `asset-deployer` resolve its own storage URI from `asset_id` via a DB lookup rather than have the orchestrator pre-resolve it — keeping handler self-containment.
- **sources:** archive_april_26/004_site_work_orchestrator.md (whole file); docs024_key_docs_latest/002_system_architecture(4).md#"Dispatch Loop: Dynamic Work Item Routing"
- **relations:** dispatch-pattern spawn→call; asset-deployer agent; handler self-containment principle
- **verify-later:** `grep -n "PresignedURLToS3URI\|resolveStorageURIFromAsset" platform/` to confirm these functions still exist.

<!-- SOURCE: U12_docs024_archives.md -->
### Self-spawning flat dispatch-loop (pre-scheduler design, superseded)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Archive (dated 2026-02-24) states as a "Key Design Decision": "No sub_workflows — they've been problematic," "No loops in dispatch — one item per invocation, self-spawns for next item"; the eventual system uses a genuine `"action":"loop"` construct driven by a scheduled 30s/120s kafka-scheduler tick.
- **what:** An early design decision to avoid the framework's loop/sub_workflow mechanism entirely, having `build-dispatch-loop` process exactly one item then spawn a fresh copy of itself. Later abandoned in favour of the scheduler-driven periodic trigger combined with the fully-developed in-workflow loop mechanism.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Key Design Decisions"; docs024_key_docs_latest/010_scheduler_and_tasks.md#"build-pipeline-trigger"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C — Loop Mechanisms"
- **relations:** loop-mechanisms (dev-guide appendix), scheduler-and-tasks, build-dispatch-loop agent
- **verify-later:** confirm `build-dispatch-loop`'s current agent_definition workflow uses the loop action, not self-spawning.

<!-- SOURCE: U12_docs024_archives.md -->
### claim_work_item atomic claim action + load_work_items first_item patch
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Archive lists these as "created, not yet committed" (Feb 2026); live loop-mechanisms appendix shows `claim_work_item` as a fully standard, already-existing action used in production.
- **what:** `claim_work_item` performs an atomic `UPDATE ... WHERE status IN ('triaged','approved') RETURNING id` so concurrent dispatch loops can't double-process the same item. The companion `load_work_items` patch added a `first_item` convenience field since the framework's path resolver doesn't support array indexing.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Completed Artifacts"; docs024_key_docs_latest/001_development_guide(5).md#"Appendix C"
- **relations:** dispatch-loop pattern, loop mechanisms, scheduler-and-tasks
- **verify-later:** none needed — graduated cleanly from draft to shipped mechanism.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### collected_data growth causing OOM-kills and lost work
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** "Status: Diagnosed, not yet fixed" (FOCUS_platform_reliability doc header)
- **what:** component-quality-auditor orchestrations were observed holding 18MB `collected_data`, causing OOM-kills at the 512Mi pod memory limit. OOM mid-publish causes phantom-completed orchestrations and cascading parent-orchestration timeouts. Suspected bloat contributors: `__raw_message__` duplicating input_data, unbounded `processing_history`, large uncleared LLM responses. A separate consumer-group bug (per-pod `a.AgentID` with FirstOffset, causing backlog replay on restart) was flagged in the same investigation, deliberately held back.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-1, js_snippets_news_gaswholesalers/old/TODO_orchestration_memory_bloat.md
- **relations:** Reaper mechanisms and gap; consumer group bug
- **verify-later:** orchestration_states.collected_data for component-quality-auditor, agent.go a.AgentID consumer group

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Page content-creation build pipeline trace (page-build-handler workflow)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** each hop verified directly against chassis source, 2026-05-20
- **what:** Documents, hop by hop, how a `pages` row's bare `sections` list becomes populated, deployed HTML with linked `page_components`: load_page_record → plan_sections (triages by schema source) → content writer's `extractResponseContent` (flat string, dead end for structured fields) → RenderComponentAction → CompilePageSectionsAction → SavePageSectionsAction (structured-metadata path preferred, HTML-regex fallback, orphans a page_component when metadata isn't fully recovered).
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-page-content-creation-flow, js_snippets_news_gaswholesalers/old/page_content_creation_flow.md
- **relations:** Stale site_plan gap; per-section briefs gap; isolated build test methodology; extractResponseContent flat-string hypothesis (superseded)
- **verify-later:** LoadPageRecordAction, PlanSectionsAction, RenderComponentAction, SavePageSectionsAction

<!-- SOURCE: U13_docs024_small_dirs.md -->
### extractResponseContent flat-string hypothesis for FAQ root cause (superseded)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** old/015_content_data_persisted.md hypothesizes the writer's string-only extraction is the cause; the isolated build test later proved the writer populates a `questions` array correctly when given a clean plan
- **what:** An intermediate working hypothesis that the content writer itself could never populate a structured field like FAQ's `questions` array. Superseded once the isolated `faq-test` build proved the writer works correctly standalone; the real cause was Defect 1 (duplicate content surfaces).
- **sources:** js_snippets_news_gaswholesalers/old/015_content_data_persisted.md, js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#The-test-that-settled-the-cause
- **relations:** FAQ duplicate content-surface bug; page content-creation build pipeline trace
- **verify-later:** grep/inspect `questions`; `faq-test`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Site-chrome rendering gap (missing nav/header/footer in relay build path)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Measured baseline table shows "nav: 0" on all four rendered pages of dartsonline.com; hypothesis: "the RELAY build path lacks the site-chrome rendering step... no chrome-render step was observed" — not yet confirmed against site_components rows
- **what:** A suspected structural gap discovered via direct measurement of a live site (dartsonline.com, zero `<nav>` elements on every page, single stylesheet link pointing at a CSS file whose `needs_design` item was still triaged/undelivered): the newer work-item-relay build path may never invoke the chrome-rendering step that the older `pageflow-builder` path has, while `build-site-planner`'s `populate_nav_tables` only writes nav data, not rendered chrome. Same failure class as the "Design/composition work-item emission gap" found independently via code-reading in plainjanedomain.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE,#THE THREE-WAY SPLIT
- **relations:** Design/composition work-item emission gap; Three-way split quality-gap diagnostic method
- **verify-later:** SELECT * FROM site_components WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'

<!-- SOURCE: U14_docs019_runbooks.md -->
### Generic orchestrate envelope as the universal manual trigger
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** RUNBOOK(31) §6D/§6F full kcat scripts: "The envelope's action: orchestrate + config.agent_type is the generic entry point"; reused verbatim for code-indexer, diagnose, and §7E.
- **what:** One trigger shape for hand-running any agent: kcat-produce to `system.agent.generic.requests` with correlation/orchestration/message/request ids, `action=orchestrate`, `config.agent_type=<entry agent>`, and task-specific `input_data`. Known wrinkles recorded: `site_id` intermittently arrives empty (reproducibility bug, parked), and runtime selectors (`site_id` vs `runtime_site` domain) drive different evidence filters in load_runtime.
- **sources:** docs019/RUNBOOK(31)_diagnosis_loop.md#6D; docs019/RUNBOOK(31)_diagnosis_loop.md#6F; docs019/RUNBOOK_code_retrieval_route(21).md#7E
- **relations:** diagnose-orchestrator wrapper; needs_diagnosis intake (future replacement); correlation-id discipline
- **verify-later:** drafts/084_TRIGGER_diagnose_v1.sh; 080c/082 trigger scripts; site_id-empty envelope bug

<!-- SOURCE: U14_docs019_runbooks.md -->
### Workflow result contract — resolveResultSpec and preferred-key unification
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** gamesdesign_index_rebuild "page-content-writer's singular output_field now flattens into the response … resolveResultSpec emits a resolved-contract line, a deprecation Warn per old key" (deployed 2026-06-18); code_retrieval_route(21) "STATUS 2026-07-04: DEPLOYED (image v1.0.1092) and the rename migration APPLIED + VERIFIED — all four agents on preferred result_from".
- **what:** The contract by which a completing workflow's result is extracted (result_from / output_field(s) / flatten / mapping). Its failure was the root of the flagship silent no-op (singular output_field ignored → page collapsed to a stub). Evolution: flatten fix + contract logging + deprecation census (2026-06-18) → Option A unification: the resolution table lifted verbatim into datahelpers/result_contract.go (ResolveResultSpec + ApplyResultSpec), complete_workflow delegating to it, agents migrated to preferred key spellings — one source of truth for coordinator and action paths.
- **sources:** docs019/RUNBOOK_gamesdesign_index_rebuild.md#what-this-exercises; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2); docs019/RUNBOOK(31)_diagnosis_loop.md#6G (the fixture cause)
- **relations:** gamesdesign fixture; oversize-result delivery; loop_scope_field lesson (dotted-config class); parent result key under step name
- **verify-later:** datahelpers/result_contract.go + 7-case test; NNN_rename_complete_keys_preferred.sql applied state

<!-- SOURCE: U14_docs019_runbooks.md -->
### Oversize-result delivery: loud failure, size guards, responses-are-summaries
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) "MEASURED (2026-07-03): … cd_bytes=1,270,781 — THE COMPLETION RESPONSE CARRIED THE FULL collected_data alongside the 6KB result"; "FIX = NNN_fix_diagnose_complete_output_fields.sql … APPLIED 2026-07-03"; gamesdesign_index_rebuild fix #3 "an undeliverable (oversize) result now fails loudly (error_unrecoverable + agent_error_log) instead of a status:'completed' stub".
- **what:** The family of oversize-result failures and their doctrine. Mechanism found in code: `result_from` is a key CompleteWorkflowAction never reads, so its fallback shipped the ENTIRE collected_data (1.27MB > Kafka ~1MB cap → Message Size Too Large; child fails at complete). Fixes: output_fields selection per agent; a response SIZE GUARD (max_response_bytes, truncatedResponseStub naming where the real result lives); earlier, removal of the silent "Full result exceeded size limit" completed-stub in favour of loud error_unrecoverable. Doctrine: responses are summaries; heavy artifacts live in the DB, retrievable by correlation_id; raising the broker cap is inversion, last resort.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#7D (§7E attempt-1 blocker 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#6; docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 2)
- **relations:** workflow result contract; bundle size doctrine; diagnosis_artifacts egress (same principle)
- **verify-later:** workflow_actions.go size guard + truncatedResponseStub; MaxResultSizeBytes const

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parent stores child result under the step name
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) follow-up 1 "RESOLVED 2026-07-03: jsonb_object_keys shows the child response stored under the STEP NAME call_diagnoser … 'diagnose-agent_result' never existed. Migration WRITTEN: NNN_fix_orchestrator_complete_key.sql".
- **what:** Engine behaviour: when a call step has no output_field, the child's response is stored in the parent's collected_data under the STEP NAME — imagined synthetic keys like `<agent>_result` never exist. Two orchestrators carried imagined keys; fixed by pointing complete steps at the real step-name keys. A recurring class with the dotted-config lookups.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#follow-ups (item 1)
- **relations:** workflow result contract; loop_scope_field lesson
- **verify-later:** collected_data ? 'call_diagnoser' on a post-migration diagnose run

<!-- SOURCE: U15_docs019_running_notes.md -->
### SagaCoordinator output_field singular/plural extraction contract
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Fix = `resolveResultSpec` (result_spec.go) treats singular as FLATTEN, keeps plural unchanged. Pure chassis change" (v2(36)).
- **what:** The workflow-completion contract by which a step's declared `output_field` (singular) or `output_fields` (plural) determines how `extractWorkflowResult` shapes the final result; the singular case was silently mishandled before the fix, and `complete_workflow`'s `result_from` is the fixed, now-correct field the diagnose-agent's own complete step uses.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 gamesdesign entry.
- **relations:** Gamesdesign silent-no-op bug; diagnosis-loop chassis integration architecture.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Kafka completion-payload size guard (message-too-large bug + shared result contract)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "§7E attempt 1 (17933a83) FAILED: 'message too large'... the number: diagnosis 6KB, cd 1.27MB ⇒ completion ships FULL collected_data... both §7E blockers fixed in one change-set" (NOTES_running_synthesis_v4(39).md headers, 2026-07-03).
- **what:** A production-triggered bug where an agent's Kafka completion message failed with "message too large" because the completion producer ships the FULL accumulated `collected_data` (1.27MB), not just the declared result (6KB) — triaged to the child-completion producer, fixed alongside a second guard-vs-expansion blocker, and generalised into "Option A": a shared result contract plus a response-size guard applied platform-wide (not just to the diagnose agent).
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-03 §7E entries (headers + surrounding turn log).
- **relations:** SagaCoordinator output_field contract; diagnosis-loop chassis integration architecture.
- **verify-later:** The "Option A" shared result contract's actual deployed shape; whether the size guard applies to all agents or only the diagnose path.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Child-completion result key convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "follow-up #1 resolved: child result key = STEP NAME `call_diagnoser`" (NOTES_running_synthesis_v4(39).md header, 2026-07-03).
- **what:** Confirms (against real orchestrator rows, not guessed) that a spawned child agent's result is read by the parent under the CALLING STEP's own name/output_field (e.g. `call_diagnoser`), not a role-based or agent-type-based key — resolving several rounds of guessed migration SQL in the diagnosis-loop chassis port.
- **sources:** NOTES_running_synthesis_v4(39).md header 2026-07-03; NOTES_running_synthesis_v2(36).md "diagnose migration CORRECTED against real rows."
- **relations:** Diagnosis-loop chassis integration architecture; Workflow default_config location convention.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Workflow default_config location convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "the query result overturned my main assumption... task_workflow / orchestrator_workflow / orchestration_workflow are ALL NULL on every working orchestrator... The workflow lives in default_config" (v2(36), 2026-06-17).
- **what:** A load-bearing, empirically-corrected fact about the chassis schema: an agent's actual workflow (start_step/steps graph) lives in `agent_definitions.default_config`, never in the three separately-named `*_workflow` columns that exist on the table — discovered only by querying real working rows after an entire migration draft was written on the wrong assumption (inferred from the dev-guide's prose example, which showed a workflow object but never named its column).
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "diagnose migration CORRECTED against real rows."
- **relations:** Real-rows-beat-prose discipline; diagnosis-loop chassis integration architecture.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Work-item relay spine (batons, handler_agent, the 30s pump)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread §2 "Builder route §B0–§B4 CLOSED: spine decided = the work-item relay (§B3, pre-registered criteria)"; README_flows describes the live pump.
- **what:** Builds move as a relay: the baton is a site_work_items row naming a handler_agent; build-pipeline-trigger (every 30s, behind a pre-query gate) seeds build_queue and picks one dispatchable site; build-dispatch-loop claims items atomically and spawns a dynamic handler per item. One hop = one baton, one agent, one site_specs entry, one new baton. Around it a fully enabled immune system: evidence-based claimed-item timeout (its SQL documents the gamesdesign false-positive lesson), feasibility-recheck, both reapers, archiver, cleanup — with improvement-sweep currently disabled (the improvement loop not running) while content-feed-refresh runs six-hourly.
- **sources:** HANDOFF_builder_thread.md#1,#2; README_flows.md
- **relations:** hop-insertion pattern; needs_diagnosis intake (reuses the relay); scheduler-and-tasks category
- **verify-later:** scheduled_tasks rows; improvement-sweep flag state

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Autonomous Build-and-Operate — the trust-not-capability thesis
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) header "Status: synthesis spine, built over several turns. Sections 6–8 are deliberately stubs to detail next"; §1 "the technical pieces are proven … the blocker is LLM-response uncertainty"
- **what:** Umbrella vision: everything already built is apparatus for a single reliability problem — bound LLM uncertainty at each step enough to progressively remove the human. The whole plan targets building/operating a real site (vonc.com) autonomously by composing the companion FOCUS mechanisms into one toolkit across the full lifecycle.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#1, ED/MASTER_autonomous_build_and_operate(4).md#9
- **relations:** umbrella over all the salience/standards/mediator/context FOCUS concepts below; vonc
- **verify-later:** none (discussion doc)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Build-vs-operate asymmetry
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §4 "Build … isolatable … Competition is safe. Operate … live, stateful … Competition is risky"
- **what:** Build work (actions, workflows, components, agent defs) is branchable/sandboxable so competition is safe and the ratchet moves fast; operate work (provisioning, scaling, incident response) is live and stateful so it leans on known-good + canary + rollback + tighter HITL. The cascade's tier mix shifts by domain.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#4, ED/MASTER_autonomous_build_and_operate(4).md#7.6
- **relations:** lifecycle map (Tier A/B/C); reliability cascade
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Lifecycle map by verifiability + containment (Tier A/B/C)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.1–6.2 tables; "Ceiling is separate from current maturity"
- **what:** Every capability's autonomy ceiling is set by two independently-failing factors — verifiability (can we tell against ground truth it's correct) and containment (how bad/reversible if wrong). Tier A (Go actions, SQL, component-structural, observability, rollback) reaches autonomy; Tier C (security, sharding, replication, live remediation, meta-loop) stays gated regardless of agent capability. Ceiling ≠ maturity drives where to invest.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.1, ED/MASTER_autonomous_build_and_operate(4).md#6.2, ED/MASTER_autonomous_build_and_operate(4).md#6.4
- **relations:** build-vs-operate asymmetry; verification harness
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Autonomous control loop (route-produce-verify-gate-apply-feedback)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7 "the new machinery wraps each leaf task … the orchestrator's decompose-and-dispatch is reused unchanged"; §7.7
- **what:** The orchestrator's decompose-and-dispatch is reused unchanged; new machinery wraps each leaf: route (cascade), produce, verify (harness), gate (trust ledger level), apply→derived-state, feed back. Ops re-enters the same loop, triggered by derived state instead of a build goal.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#7.6, ED/MASTER_autonomous_build_and_operate(4).md#7.7
- **relations:** cascade router; verification harness; trust ledger; mediator
- **verify-later:** existing orchestrator spawn/work-items machinery

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Mediator routing model (change → consultees)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS_mediator_routing_model.md#1 "the doc tree's metadata is the routing table … routing is matching a change descriptor against those tags"
- **what:** Routing reduces a change to a descriptor `{change_types, areas, touched_subsystems}` (paths→types via globs), queries the manifest for matching active standards, and acts on each by its own fields (run `check` validator / compose `reference` into prompt / consult concern agent / spawn area-owner). Runs a cheap tier always and an expensive tier on trigger; runs twice per change (pre from intent, post from diff).
- **sources:** ED/FOCUS_mediator_routing_model.md#2, ED/FOCUS_mediator_routing_model.md#5, ED/FOCUS_mediator_routing_model.md#6, ED/FOCUS_mediator_routing_model.md#7
- **relations:** atomic standard (fields are the routing table); self-dev Position C; concern curators
- **verify-later:** proposed change classifier; path→area glob map

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Wrapper-orchestrator pattern (pod lifecycle)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "Every pod-running agent needs a parent that spawned it … the rule we violated when we first wrote site-adoption-agent"; canonical minimal wrapper med-export-orchestrator
- **what:** Agents get a dedicated Kubernetes Job pod only when reached via `spawn_agent`→`call_agent`; substantive work reached via the generic entry point runs in-chassis with interleaved logs and blocks a shared pod slot. The fix is a tiny wrapper orchestrator (spawn→call→complete) so real work runs in its own pod.
- **sources:** WM/001_development_guide(0).md#every-pod-running-agent-needs-a-parent-that-spawned-it, WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/007_adoption_pipeline_v3.md#the-adoption-agent
- **relations:** generic entry point vs job topics; site-adoption-orchestrator; agent = row in agent_definitions
- **verify-later:** SpawnAgentAction; spawnAgentKubernetesJobFromDefinition; setupAgentTopics

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Kafka topic model (generic entry point vs per-spawn job topics)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) "system.agent.generic.requests — the generic entry point … job.<stable-identity>.requests — per-spawn dedicated topics"
- **what:** Two distinct patterns: the shared generic entry point (consumed by long-lived chassis replicas that run the workflow in-process via `config.agent_type`) for anything starting a workflow from outside a spawn tree, versus per-spawn `job.<stable-identity>.requests` topics for agent-to-agent traffic inside a workflow; plus per-type fixed topics for long-lived adapters.
- **sources:** WM/001_development_guide(0).md#topics-the-generic-entry-point-vs-per-spawn-dedicated-topics, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; scheduled tasks target_topic; adapters
- **verify-later:** KAFKA_TOPIC(S) env; createTopics; MessageProcessor.extractGroupInfo

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Loop mechanisms (workflow expansion, dispatch loop, ErrLoopExpansionHandled)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix C "Loops are not Go for-loops — they are dynamic workflow expansion. At runtime, the loop step injects N × M steps"
- **what:** A loop step resolves a collection, then `handleLoopExpansion` injects `{loop}_iter_{N}_{substep}` steps into the workflow plan plus a `_complete` aggregator; `setLoopVariable` sets the current item and propagates prior-substep outputs. The canonical use is the dispatch loop (claim→spawn→call→mark). `ErrLoopExpansionHandled` is a sentinel fixing a race where a fast child response would otherwise skip remaining iterations.
- **sources:** WM/001_development_guide(0).md#appendix-c-loop-mechanisms, WM/001_development_guide(0).md#the-dispatch-loop-pattern, WM/001_development_guide(0).md#the-race-condition-and-errloopexpansionhandled
- **relations:** dispatch loop state machine; dynamic dispatch
- **verify-later:** loop_actions.go; loop_expansion_handler.go; coordinator.go continueExecution

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Architectural tensions catalogue (infer-and-repair; multi-owner page identity)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** ARCH_TENSIONS(2) "An entry graduates from 'observed' to 'resolved' only when the resolution principle is actually enforced in code"
- **what:** A living catalogue naming genre-level design tensions that keep generating incidents. Tension #1: trusting LLM free-text structure as truth then repairing with starved heuristics vs deriving structure deterministically from the LLM's reliable signals. Tension #2: page identity re-derived in multiple stages that undo each other.
- **sources:** WM/ARCHITECTURAL_TENSIONS(2).md#tension-1-trusting-llm-free-text-structure-as-truth-infer-and-repair-vs-deriving-structure-deterministically, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; adoption faithfulness strip; site plan reconciler
- **verify-later:** ValidateRoles/nestedRoleFromURL; CanonicalisePage; normaliseRole vs normalisePageType

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Agent = row in agent_definitions (workflow model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016 v2_44 §6.0 "An agent is not a Go type, class, or file. It is a row in agent_definitions whose default_config.workflow is a declarative graph of steps"
- **what:** Agents live in the database, not the Go source. A workflow is a step graph threaded by dotted-path reads from a shared data bag; "every agent is an orchestrator" is literal. `spawn_agent`+`call_agent` are a pair. Traps: the description can contradict the config; `agent_definitions` may be read from more than one DB.
- **sources:** WM/016_debugging_guide_v2_44.md#6.0, WM/016_debugging_guide_v2_44.md#6, WM/001_development_guide(0).md#agent-message-structure
- **relations:** wrapper-orchestrator; kafka topic model; loop mechanisms
- **verify-later:** agent_definitions (templates_db vs clients_db); default_config.workflow; SpawnAgentAction

<!-- SOURCE: U18_sql_for_agents.md -->
### Wrapper-orchestrator pattern ("spawns a temporary pod to do X")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Named as the canonical pattern in 104 ("Pattern copied verbatim from med-export-orchestrator..."), 121 ("mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim — the canonical scheduler-triggered wrapper + task-worker"), 122 (dev guide "does this agent need a wrapper?" test).
- **what:** Convention: substantive in-chassis work (long LLM loops, crawls, collections) must not run in shared generic pods; instead a thin orchestrator (spawn_agent → call_agent → complete) creates a dedicated K8s Job pod for the worker, giving clean per-correlation logs, isolation, and idle-timeout cleanup. Spawn-before-call ordering is required for target_role lookups (109/111/112).
- **sources:** 096_vet_med_url_discover_orchestrator.sql; 104_site_adoption_orchestrator.sql; 121_intent_collector_agents.sql; 122_diagnose_agents.sql
- **relations:** idle_timeout_seconds; scheduler-and-tasks; K8s Job lifecycle
- **verify-later:** dev guide §wrapper test; spawn_actions.go

<!-- SOURCE: U18_sql_for_agents.md -->
### idle_timeout_seconds (Job pod auto-exit)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 075 ALTER TABLE + fleet-wide backfill (180s) with rationale ("timer resets on every message... multi-step workflow... stays alive as long as responses keep arriving").
- **what:** Column on agent_definitions controlling how long a spawned Job pod waits with no messages before exiting cleanly (0 = no timeout for Deployment agents). Paired with TTLSecondsAfterFinished for cleanup. The 075 list doubles as a census of the then-live spawnable fleet.
- **sources:** 075_various_timeout_column.sql
- **relations:** wrapper pattern; K8s cleanup; debugging (timeouts)
- **verify-later:** chassis idle-timer implementation

<!-- SOURCE: U19_sql_tables_components.md -->
### Site / area / page component hierarchy
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** site_components deployed and populated (012); site_areas/area_components created with default 'main' area backfill, but only the site level shows active use; get_page_component fallback function defined.
- **what:** Three-level slot resolution for page chrome: area_components (per site_area override) → site_components (site-wide header/footer/head with rendered_html + content_data for re-render, UNIQUE(site_id, slot_name)) → assembly. site_areas model major site sections with their own nav_style and theme_overrides; get_page_component(page, slot) walks area-then-site.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql; docs/agent_docs/sql_for_tables/014_site_areas.sql; docs/agent_docs/sql_for_tables/015_area_components.sql; docs/agent_docs/sql_for_tables/003_pages.sql#site_area_id
- **relations:** component-based headers; pages.site_area_id; locks (site_components lock columns).
- **verify-later:** area_components usage in production; get_page_component callers.

<!-- SOURCE: U19_sql_tables_components.md -->
### Pages / page_components split (structure vs content)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003 records the design correction: columns first added to pages then explicitly reverted — "Content (rendered_html, content_data) lives in page_components table. Pages table just needs workflow tracking fields"; live dump confirms.
- **what:** pages holds metadata, navigation and workflow (build_status planned→…→deployed/needs_rebuild, sections as planning reference, version) plus per-page rendered_header/rendered_footer/rendered_head for minimal reassembly; page_components holds the actual sections (position, slot_name, component_id, content_data, rendered_html, content_hash, review fields, deploy_commit, research_id). 004b describes the intended three layers: content (content_items) → layout (page_components) → structure (pages).
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** content_items layer; page build workflow; site snapshots capture both.
- **verify-later:** assembly path reading rendered_* columns; build_status writers.

<!-- SOURCE: U19_sql_tables_components.md -->
### awaited_requests global request/response registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Two schema generations (tables_sql 001 → sql_for_tables 001 matching the AwaitedRequest Go struct), plus later additions of 'processing' status, processing_started_at/processing_pod claim tracking, and cleanup function.
- **what:** DB-backed registry matching Kafka responses to waiting orchestrations, solving the race where a child creates a request while the parent receives the response. Keyed by request_id with orchestration/correlation context, target agent, responses/requests topics, retry_version, reply_to_request_id chaining, timeout_at, and status lifecycle waiting→processing→processed/expired/cancelled/error. Expired rows are marked then purged after 7 days by cleanup_expired_awaited_requests.
- **sources:** docs/agent_docs/sql_for_tables/001_awaited_requests.sql; docs/agent_docs/tables_sql/001_awaited_requests.sql
- **relations:** processed_messages idempotency; HITL runbook; orchestration_states.
- **verify-later:** state.go AwaitedRequest struct; cleanup scheduling.

<!-- SOURCE: U19_sql_tables_components.md -->
### processed_messages idempotency dedup
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Live \d output plus applied ALTERs adding retry_version and re-keying the PK to (correlation_id, request_id, agent_id, retry_version).
- **what:** Exactly-once message processing guard: each consumed message records correlation/request/agent identity; the composite PK including retry_version allows deliberate retries while blocking duplicate deliveries within a retry generation.
- **sources:** docs/agent_docs/sql_for_tables/007_processed_messages.sql
- **relations:** awaited_requests retry_version; Kafka consumer semantics.
- **verify-later:** consumer insert-or-skip logic.

<!-- SOURCE: U19_sql_tables_components.md -->
### Orchestration ↔ site linkage (orchestration_states.site_id)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Migration with three-path backfill from collected_data (input_data.site_id, site_record.site_id, top-level) and verification counts against gamedesign.uk.
- **what:** Direct nullable site_id column on orchestration_states (set at creation) replaces JSONB spelunking for "orchestrations for this site", with a partial index for active orchestrations per site. Nullable because not all orchestrations are site-scoped (health checks).
- **sources:** docs/agent_docs/sql_for_tables/036_orchestration_states.sql
- **relations:** debugging queries; improvement-sweep pre_query.
- **verify-later:** creation-time population in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Sites contact-identity denormalisation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Applied ALTERs + COALESCE backfills from content_data (company_name/business_name, tagline/slogan, email/contact_email, phone/contact_phone, logo_text fallback chain); one-off content_data patches for live sites.
- **what:** Frequently rendered identity/contact fields promoted from sites.content_data JSONB to first-class columns (company_name, tagline, email, phone, logo_url, logo_text, contact_address) feeding the render context for headers/footers/heads, with content_data retained as the brief-derived store of record.
- **sources:** docs/agent_docs/sql_for_tables/011_sites_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#issue-1a; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** component-based headers render context; site_specs identity aspect (overlapping data — coherence question).
- **verify-later:** which of sites columns vs site_specs.identity is authoritative for rendering today.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Universal orchestration principle ("every agent is an orchestrator")
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "Current Implementation Status … ✅ Universal orchestration capability"; repeated as a working rule in docs002 ("Every agent is an orchestrator").
- **what:** No architectural distinction between orchestrator and worker agents. Every agent runs the same chassis, can spawn children, orchestrate workflows, and execute tasks simultaneously; complexity is fractal (agents compose into arbitrarily deep trees). This is the founding philosophy of the agent-chassis platform.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md
- **relations:** agent chassis; SagaCoordinator workflow engine; agent groups; superseded in practice for site building by the work-item pipeline (development-guide) but still the chassis foundation.
- **verify-later:** agent-chassis main; platform/orchestration/coordinator.go; whether dynamic spawn trees are still exercised vs. static handler deployments.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Stateless agents with database-backed orchestration state
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002: "✅ Stateless agent design with database-backed state"; orchestration_states schema with version optimistic locking shown.
- **what:** Agents are ephemeral execution containers (K8s pods/Jobs); all orchestration state lives in the `orchestration_states` table (orchestration_id, current_step, awaited_requests, status, processing_history, version). Pod crashes lose no work; the DB is the authoritative source of truth.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#orchestration-state; docs001_flow_general/README.012.flow3.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md
- **relations:** SagaCoordinator; AwaitedRequests map; environment reset runbook (truncates this table).
- **verify-later:** clients_db orchestration_states table; UpdateStateWithVersion; whether table is still active or superseded by work_items.

<!-- SOURCE: U20_legacy_docs_a.md -->
### ExecutionContext unified message envelope and ID semantics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.002 "✅ ExecutionContext as unified message structure"; detailed ID-trace docs (flow10, requestIDflow) resolving the semantics.
- **what:** Every Kafka message carries an ExecutionContext: correlation_id ties the whole end-to-end operation; orchestration_id identifies one workflow instance; request_id identifies a single request/response cycle (new per communication); parent_orchestration_id records who called you; plus tree depth/path, fuel budget, timeout, responses_topic. Sender constructs the child's context; receiver trusts headers.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md
- **relations:** perspective transformation; reply-to metadata; MessageType semantics.
- **verify-later:** platform/orchestration/types ExecutionContext; messaging/context.go NewMessageContext.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Topic-per-agent Kafka communication
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** README.020.flow9: "LEGACY TOPICS (Pre-created)… Why Legacy Topics Persist"; the doc itself designs the job-topic replacement.
- **what:** Original model: static topics `system.agent.{type}.requests/responses` per agent type. Kept only as bootstrap/well-known entry points (initial client contact) after message-stealing and routing conflicts pushed the design to dynamic job topics.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#topic-per-agent; docs001_flow_general/README.020.flow9.topicflow.md
- **relations:** superseded by job-specific dynamic topics (hybrid model); ultimately by the work-item pipeline for site building.
- **verify-later:** which system.agent.* topics still exist on the cluster.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Job-specific dynamic Kafka topics (hybrid bootstrap model)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow9 "after discussion" decision section + spawn_actions2 walkthrough showing `job.{corrID}-{orchID}-{agentType}-{step}.requests/responses` created at spawn; robot-hands runs used them live.
- **what:** Each spawn creates private per-orchestration topics from a "stable identity" (correlation short + orch short + agent type + spawning step). Root agents listen on standard pre-created topics; spawned agents get their topics via REQUESTS_TOPIC/RESPONSES_TOPIC env vars. Parents talk to children on the child's job topic; children reply to the caller's responses topic carried in headers. Solves the chicken-and-egg bootstrap problem and message collision between parallel jobs.
- **sources:** docs001_flow_general/README.020.flow9.topicflow.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** stable identity; spawn_agent; environment reset runbook (deletes job.* topics).
- **verify-later:** kafka.CreateStableIdentity; topic creation in SpawnAgentAction; current topic list.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Two-phase agent lifecycle (spawn + initialize handshake)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow11: "initialize is not treated as a command to start a workflow… handled as a special protocol action"; multiple traced runs.
- **what:** Spawning creates a K8s Job then sends an `initialize` protocol message; the new pod configures itself (role, topics), sends an initialization response, and only then does the parent resume and send `process` work. Initialize bypasses the workflow engine entirely — its only purpose is setup/readiness confirmation. Isolates init failures from execution failures.
- **sources:** docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.021.flow10.initialrequestflow.md; docs001_flow_general/README.010.flow.md
- **relations:** spawn_agent; await_response semantics; a fire-and-forget spawn variant caused ignored init responses (flow12).
- **verify-later:** processor.go initialize handling; SendInitializationResponse.

<!-- SOURCE: U20_legacy_docs_a.md -->
### SagaCoordinator DB-defined JSON workflow engine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Extensive traced executions (flow docs); workflows stored as JSON `{start_step, steps{action, config, next_step}}` in agent_definitions/agent_group_definitions and executed live.
- **what:** The coordinator loads a JSON workflow from the DB, executes steps via an action registry, stores each step's result in CollectedData under the step name, pauses on `await_response: true` by recording request IDs in an AwaitedRequests map (status AWAITING_RESPONSES), and resumes when matching `in_response_to_request_id` responses arrive (join when the map empties). complete_workflow packages results and replies to whoever is waiting — root vs child completion unified.
- **sources:** docs001_flow_general/README.010.flow.md; docs001_flow_general/README.023.flow12.await_response.md; docs001_flow_general/README.006.executeLocalAction1.refactor_into_functions.md; docs001_flow_general/README.046.workflow_actions1.refactor_into_functions.md
- **relations:** action registry (validate_input, transform_data, execute_llm_prompt, spawn_agent, call_agent, aggregate_data, conditional_branch, complete_workflow…); await_approval builds on the same pause mechanism.
- **verify-later:** platform/orchestration/coordinator.go; actions/registry.go; whether coordinator still runs under current handlers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Reply-to metadata (__work_request__) and respond-to-caller convention
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.081.b "Clean Reply-To Architecture… Store reply-to metadata when receiving a work request, use it when completing"; docs002/0100d states the convention as an operating rule.
- **what:** Each agent stores, at work-receipt time, the request_id it must answer and the parent's responses topic together (`__work_request__` in CollectedData) and uses them at complete_workflow. Rule: agents always respond to the *caller's* responses topic, never their own. Works at any hierarchy depth; replaced fragile multi-fallback lookups and fixed empty `in_response_to_request_id` bugs.
- **sources:** docs001_flow_general/README.081.b.requestIDflow.md; docs001_flow_general/README.014.flow4.1.routingtooriginalsender.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md#response-topic-routing
- **relations:** ExecutionContext; CompleteWorkflowAction; early routing failure modes.
- **verify-later:** BuildCollectedData storing __work_request__; workflow_actions.go completion path.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Perspective transformation (sender constructs context, receiver trusts headers)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** flow6: "The critical fix is in ProcessMessage… NewMessageContext(msg, headers, p.agentType). This ensures every agent sees the conversation from their own perspective"; flow10 codifies sender responsibility.
- **what:** On receipt, NewMessageContext transforms the message into the receiving agent's own perspective (its own OrchestrationID becomes primary; the caller's becomes ParentOrchestrationID). The *sender* is responsible for correctly constructing the child's context headers; the receiver only deserialises and trusts them — earlier receiver-side guessing caused validation failures and misrouting.
- **sources:** docs001_flow_general/README.017.flow6.md; docs001_flow_general/README.021.flow10.initialrequestflow.md
- **relations:** ExecutionContext; MessageType semantics.
- **verify-later:** messaging/context.go NewMessageContext signature and transformation logic.

<!-- SOURCE: U20_legacy_docs_a.md -->
### MessageType semantics (request = actively working, response = reporting back)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** README.070.b full conceptual write-up with log excerpts ("exec_ctx_message_type":"request" when parent resumes).
- **what:** MessageType describes what the agent is doing *now*, not what just happened: a parent that has received a child's response resumes its own workflow in "request" mode with InResponseTo cleared. Prevents routing/semantic confusion when continuing execution after responses.
- **sources:** docs001_flow_general/README.070.b.execution_context_flow.md
- **relations:** SagaCoordinator continueExecution; perspective transformation.
- **verify-later:** continueExecution fresh-context construction.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Fuel budget resource limiting
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** FuelBudget field in ExecutionContext (README.002) and `fuel_budget=1000` header in test messages; CreateResponseContext takes "fuel used — calculate properly in production" (0100) — no doc claims enforcement.
- **what:** A per-orchestration computational budget carried in the ExecutionContext, intended to bound resource consumption of agent trees ("if budget.Remaining() < estimated.Cost() return cheaperStrategy()"). Appears plumbed but never implemented as an enforced mechanism.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.061.groupagents2.md (header)
- **relations:** long-term resource optimisation objectives.
- **verify-later:** grep FuelBudget usage — is it ever decremented or checked?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Child-orchestration timeout monitor
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** README.040 "Key Features of This Implementation: Configurable Timeout… default 5 minutes… Prevents zombie orchestrations" (claims implemented); HITL timeout doc later shows the config→Step.Timeout mapping was broken.
- **what:** Parents launch a goroutine per awaited child; on timeout it checks whether the parent still awaits that child, sends a timeout error response so HandleResponse processes it normally, and optionally marks the child orchestration failed. Timeout goroutines are in-memory only — recovery on pod restart identified as a gap.
- **sources:** docs001_flow_general/README.040.orchestration_actions.md; docs002_hitl_parallel/README.0111.hitl_timeouts.md
- **relations:** HITL approval timeouts; DefaultRequestTimeout 180s.
- **verify-later:** handleRequestTimeout in coordinator.go; recoverPendingTimeouts existence.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Parallel / fan-out execution in the coordinator
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 lists "Fan-out (parallel) execution implementation" under Outstanding Work; 0110 is an explicit proposal ("Proposed Implementation Strategy… parallel_steps array") with no completion claim.
- **what:** Design for non-blocking workflows: a step's config carries a `parallel_steps` array; executeParallelSteps dispatches all children, records all request IDs in AwaitedRequests, pauses once; processResponse joins when the map empties. Included ExecutionMode enum (sequential/parallel/fan_out). Image workflows sketched parallel_image_generation/batch_image_generation actions on the same idea.
- **sources:** docs002_hitl_parallel/README.0110.parallel_execution_proposal.md; docs001_flow_general/README.002.agent_orchestration1.philosophy.md; docs001_flow_general/README.095c.image_handling_topics.md
- **relations:** AwaitedRequests join semantics (multi-response already worked); batch-processing category is the modern relative.
- **verify-later:** whether run_parallel/parallel_steps ever landed in coordinator.go.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Early long-term platform ambitions (self-organising networks, marketplace, multi-tenant, cross-cluster)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** README.002 "Long-Term Objectives (6-12 Months)": self-organising agent teams, agent marketplace, client-isolated multi-tenant namespaces, cross-cluster orchestration with geographic failover.
- **what:** The founding roadmap's horizon list. Multi-tenancy (client schemas) and cross-cluster work later materialised (multicluster docs); the agent marketplace and learned team compositions appear to have vanished.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md#long-term-objectives
- **relations:** multicluster (live successor for cross-cluster); database-and-infrastructure client schemas; marketplace = abandoned idea worth registering.
- **verify-later:** none directly; council context only.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Environment variable validation framework (pre-spawn config validation)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** README.002 Week-3 objective ("EnvironmentBuilder… Validate all environment variables before agent spawn"); never mentioned again in any later doc.
- **what:** Planned framework to declare required/optional env vars per agent and validate before spawn to prevent runtime failures. Silently dropped.
- **sources:** docs001_flow_general/README.002.agent_orchestration1.philosophy.md
- **relations:** spawn_agent env var plumbing.
- **verify-later:** grep EnvironmentBuilder.

<!-- SOURCE: U21_legacy_docs_b.md -->
### "Every agent is an orchestrator" — elimination of agent_group_definitions
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs006/006 implementation doc: "GroupDiscovery is aliased to AgentDefinitionDiscovery... FindBestGroup now queries agent_definitions"; backward-compat table showing both group_type and agent_type message formats work.
- **what:** Architectural unification: a "group" is just an agent whose workflow spawns and calls other agents, so agent_group_definitions was eliminated in favour of agent_definitions carrying the orchestration workflow in default_config. spawn_group became a thin wrapper delegating to spawn_agent; discovery, message processor, and metadata all gained aliases for backward compatibility. This is the foundational premise of the current hierarchical agent tree.
- **sources:** docs006_workflow_builder/002_removing_agent_group_definitions.md#Part-2; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Implementation; docs006_workflow_builder/004_agent_groups_or_not.md#Key-Decisions
- **relations:** spawn-before-call pattern; intake orchestrator; agent families.
- **verify-later:** platform/discovery/agent_discovery.go; absence/deprecation of agent_group_definitions table; spawn_group action code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### relationships table — first-class entity relationships
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** Created in docs006/007 migration; docs012/006 (later era) says "Relationships (existing, empty)... This is PERFECT for semantic links between pages!" — table existed but unused, then earmarked for semantic page links.
- **what:** A generic first-class relationship entity (source/target entity id+type, relationship_type, direction, properties JSONB, status) modelled explicitly on website links ("relationships are like links — first-class objects with their own identity and state"), with relationship-scoped entity_state for learned communication preferences. Designed for org-framework roles, reused conceptually for pillar↔cluster semantic page relationships.
- **sources:** docs006_workflow_builder/006_conclude_role_entity_strategy.md#Relationships-as-First-Class-Objects; docs006_workflow_builder/007_new_tables_entity_state_log.sql#7; docs012_site_maps_and_components/006_start_concluding_links.md#Part-1
- **relations:** link-management (semantic links); org framework.
- **verify-later:** relationships table in clients_db and whether any rows exist; link_registry vs relationships usage.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent families architecture (nav/links/design/content/entity/tools/feed/maintenance)
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** docs017/019b (v5) with per-family status columns ("populate_nav_tables — Deployed"; layout-architect — New; brand-designer — Future split) and a Data Ownership Summary table mapping every table to an owner agent; phased plan 1→4.
- **what:** The master blueprint of the specialist-agent era: eight agent families each owning a data domain — navigation (nav tables), links (algorithmic health), design (brand/layout/CSS split), content (marketing/legal/SEO/product writers + reviewer + researcher), entity data, tool builder tiers, news/content feed, and maintenance — with explicit "does NOT do" boundaries, a component-builder-v2 workflow sketch, site-type stress tests (brochure/e-commerce/finance/events/platform), and single-owner-per-table data governance. Much became real (nav actions, webdesign, feeds, maintenance→work items); some never did (layout-architect, nav-layout-agent, product-content-writer).
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md; docs017_legacy_agent_rules_images_design_keydocs/002_full_new_agent_architecture.md; docs017_legacy_agent_rules_images_design_keydocs/018_agent_architecture_v3.md
- **relations:** nearly every other concept in this unit; data ownership prefigures council-agent domain ownership.
- **verify-later:** which family agents exist in agent_definitions today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** docs018/009_stale full design ("This is the #1 cause of pipeline stalls"; synthesize/retry/fail classification; "No schema changes needed"); no deployment confirmation in this unit.
- **what:** Replace lossy in-process timeout goroutines with a periodic DB sweep on every chassis pod: claim expired awaited_requests (FOR UPDATE SKIP LOCKED, 30s grace, LIMIT 20), classify — child COMPLETED means the response was lost, so synthesize a completion message from the child's final_result to the parent's topic; child FAILED forwards failure; no/running child retries up to retry_version 3 then fails the orchestration. Handles cascading stalls oldest-first and dead job topics by directly advancing parent state.
- **sources:** docs018_rerendering/009_stale_orchestration_sweeper_design.md
- **relations:** parent-timeout race; awaited_requests; debugging pipeline stalls; idle timeout/cleanup in system-architecture.
- **verify-later:** platform/orchestration/sweeper.go existence; sweeper startup in agentbase.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent message structure & spawn+call pattern (external triggering)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs018/009_agent and docs018/005 document the working format ("Agents don't exist as running processes until spawned... You cannot call an agent that hasn't been spawned in the current workflow"; responses_topic always the caller's).
- **what:** The canonical three-layer message (Kafka headers, mirrored JSON headers, config.workflow + input_data payload) for driving agents from CLI or external systems, with inline workflow support on the generic agent, mandatory spawn-before-call, and reply routing to the sender's responses_topic enabling parent-child orchestration. The operational lingua franca of the whole system.
- **sources:** docs018_rerendering/009_agent_initial_message_structure.md; docs018_rerendering/005_triggering_agent_from_kafka.md; docs007_brochure_builder/003_original_message_copy
- **relations:** HITL protocol; generic agent as thin launcher; kafka reset/cleanup runbooks (docs007/003).
- **verify-later:** current message types in platform/orchestration/types.

<!-- SOURCE: U21_legacy_docs_b.md -->
### "Database is source of truth, Git is the deployment artifact"
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** docs012/010 Core Principles ("Content lives in the database; Git is the deployment artifact"); reversal of docs012/004's "GitHub as current source of truth, database for metadata/links"; everything after (rerender, section editor, work items) depends on it.
- **what:** The pivotal data-ownership decision of the era: page content, sections, nav, entities, and design specs live in Postgres; git repos hold only rendered deployment artifacts, rebuilt from DB at will. Enables rerendering, granular editing, locking, and maintenance — and makes external git edits an anomaly to detect (git_hook_adapter desync idea) rather than a normal input.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#Core-Principles; docs012_site_maps_and_components/004_more_on_links.md#Context; docs018_rerendering/010_section_editor_architecture.md#The-Source-of-Truth-Principle
- **relations:** page_components; rerender; site-snapshots-and-revert (later formalization); deployment-github.
- **verify-later:** n/a (doctrine, observable in every pipeline).

<!-- SOURCE: U22_recent_small_docs.md -->
### Layer-1 / Layer-2 hack-resistance model
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Stated as existing platform fact (Appendix B): "Layer 1 ... publishes outward and pulls inward but never serves inbound public traffic. Layer 2 is client delivery; today that is static assets on Backblaze S3 with nothing in the request path."
- **what:** The security posture the whole chatbot design defends: Layer 1 (core K8s cluster — agents, Kafka, Postgres, all credentials) never accepts inbound public traffic; it only publishes outward (site assets, data exports, context packs to S3) and pulls inward (recorded turns). Layer 2 is static-on-S3 with nothing in the request path — "nothing to compromise." The edge worker is the only new Layer-2 compute, and the whole chatbot design is arranged to preserve this (no API keys in the page, no central VM in front of static content). Sister-project appendix documents the "nginx box keeps getting hacked" experience that motivates it.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#1, #Appendix-B, docs025.../PLAN_isolated_chat_environment(4).md#Appendix
- **relations:** edge worker, isolated chat environment boundary contract, deploy-to-S3 path
- **verify-later:** ingress rules on ai-persona-system; B2 publish path

<!-- SOURCE: U23_docs_root_vonc.md -->
### Result-contract drop fix (child workflow result replaced by a stub)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016b Part 1: "DONE... Shipped 2026-06-18 (result_spec.go + coordinator.go); verified — gamesdesign index rebuilt+deployed 06-19."
- **what:** The chassis coordinator used to discard a child workflow's result (singular output_field, or oversize) and substitute a stub that still reported success — producing no-op saves under `complete` status (a root member of the silent-success family, and the resolution of the long-open "index returns thin content" question: it was the stub, not thin generation). Fixed in result_spec.go + coordinator.go. Carried here because the 016b copies in this unit are the guide's cumulative record; the docs024 consolidated guide is the canonical home.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 1)
- **relations:** complete_error family (sibling); trust-the-artifact doctrine
- **verify-later:** result_spec.go / coordinator.go in platform

---

## Dropped-concept notes from family deltas (audit)

- RUNNING_NOTES_vonc.md (base): migration renumbering — "DB migration: Migration 003" became
  "Migration 002 = agent_definition condition fix; render_mode sweep DROPPED" (captured as a
  concept above).
- RUNBOOK_vonc_migrations early versions: original "Migration 002 — Fix render_mode on
  components" heading and the "Recommendation: Option 1 for the first deployable version"
  line (both captured as abandoned concepts above); "Two fix options — choose one" framing.
- RUNBOOK_phase2 early versions: Gap-2 sub-options (a) creator-emits-companion-snippet vs
  (b) loader-snippets-as-library-fixtures (superseded by Tier E + loader-builder); the
  step-checklist framing (P2-1..P2-6, FX-1..FX-6, PD-1..PD-3) whose IDs the later docs still
  reference; all content otherwise carried forward into (29).
- PLAN_provocation-card(2)→(3): the pre-correction trim method (hand-UPDATE the live
  instance) — replaced by the bundle-verdict method; captured under "sanctioned edit paths".
- PLAN_dynamic_sections(2)→(4): `site_plan_directives` named as a site_plans child table in
  the pre-supersession decision text; not mentioned in the final doc (noted in the plan-storage
  concept).
- HANDOFF base/(1)→(2): same method correction (rejected mechanism → bundle §4.0).
- bundle_minilobby_trim base..(3)→(4): resolver evolution v1→v4 (registry-first resolver,
  module boundary, scope-by-symbol) — captured under the cmd/bundle concept.
- provocations.sample.json v1→v3: additive key evolution today/lobby → +arena → +archive —
  captured in the data-contract concept.

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Gateway proxy pattern (auth-service → core-manager)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 008b: auth-service owns auth/user/subscription/projects; forwards all else to core-manager enriching `X-User-ID/Client-ID/Role/Tier/Email` headers; core-manager re-validates JWT with shared `JWT_SECRET_KEY` or falls back to `/auth/validate`.
- **what:** Single-front-door architecture: auth-service is the only HTTP ingress; `.Any()` wildcard routes proxy to core-manager which defines the actual method handlers. Core-manager independently validates already-issued tokens so it doesn't hard-depend on auth-service uptime.
- **sources:** archive_april_26/008b_admin_api_plan_v2.md#architecture-overview; 007b#gateway-pattern
- **relations:** Admin API; public API blocks 2/4
- **verify-later:** cmd/auth-service gateway/handlers.go; core-manager AuthMiddleware/TenantMiddleware

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### site-engine (API-only capture backend)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** service(24).go header "site-engine: the capture backend for VM-hosted backend sites (API only)"; builds/tests pass; running live on relojistas box.
- **what:** A stdlib-only Go binary that does the one thing a static page cannot: record a structured intent event server-side keyed by Host into a file store. Endpoints: `POST /intent` (capture then 303), `GET /api/hit` (1×1 beacon), `GET /stats` (key-gated), `GET /health`, plus later `GET /events` and `GET /access-digest`. No page rendering or content registry (the chassis owns both).
- **sources:** deploy_setup/working_dir/service(24).go#header, deploy_setup/working_dir/main.go#header, traffic_probe_runbook(12).md#1
- **relations:** replaces the abandoned standalone probe-go fork; page content owned by chassis intent-probe component
- **verify-later:** site-engine repo (`$OWNER/site-engine`); go.mod `module site-engine`

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Standalone "probe-go" service (abandoned first cut)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Session 1 "Forked idea.uk's Go service into probe-go … Caveat raised next session: this drifted into a separate project"; Session 2 reframed it as not-a-separate-project.
- **what:** The original framing forked idea.uk's multi-vhost Go service (page-by-Host-header, page.go + domains.json in Go) into a self-contained project. Rejected because it sat too far from the website-building chassis; page.go and domains.json were removed and the engine was trimmed to an API-only backend with content moved to chassis build outputs.
- **sources:** traffic_probe_running_notes(27).md#session-1, traffic_probe_running_notes(27).md#session-2, traffic_probe_running_notes(27).md#session-3
- **relations:** superseded by site-engine + Layer-4/thin-Layer-5 framing
- **verify-later:** n/a (removed page.go/domains.json)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Layer-4-build + thin-Layer-5-VM-deploy framing
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Session 2 conclusion "the probe is Layer 4 (build a targeted site) + a thin slice of Layer 5 (deploy a tiny backend to a VM instead of B2)"; decision to keep git→Actions and only swap the target.
- **what:** Rather than a side project, the probe reuses the existing build pipeline (Layer 4) and the git→self-hosted-Actions deploy seam (Layer 5), swapping only the destination from B2 to VM. The heavier chassis service-deployer adapter is the eventual move, not now.
- **sources:** traffic_probe_running_notes(27).md#session-2, traffic_probe_plan(11).md#where-we-are
- **relations:** underlies "commit is deploy" seam swap; defers P5 vmhost adapter
- **verify-later:** CONSOLIDATION_where_it_all_fits.md, PARALLEL_engine_deployment_and_layer5.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Phased plan P0–P5
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** plan(11) Phases: P0/P1/P2 done, P3 in progress ("Remaining for P3: land the chassis patch…"), "P4 … IN PROGRESS (this chat)", P5 not started.
- **what:** P0 structural decisions; P1 manual go-live (Path A); P2 wire deploy-on-update (two Actions); P3 make the probe a normal pipeline output (github_repo target selection + capture component + capability gate); P4 off-box collection + ranking; P5 registry + provisioning adapter.
- **sources:** traffic_probe_plan(11).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** contains most other concepts; earlier plan versions phrased P4/P5 differently
- **verify-later:** n/a

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### VM-Hosted Backend Sites class (proposed doc 024)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** plan(11) "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference; numbering operator's)".
- **what:** The genuinely-new infrastructure: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. The traffic probe is instance #1 of this class; future chat/board sections join it.
- **sources:** traffic_probe_plan(11).md#framework-integration, traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle
- **relations:** class parent of intent-probe; ties to D5 requires-backend gate
- **verify-later:** doc 024 existence; service_instances table

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Pull architecture / no collector VM
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "No third 'collector' VM: the serving box buffers (JSONL); the CLUSTER pulls over key-gated HTTPS … Pull keeps every credential in the cluster — boxes never hold DB/cluster secrets".
- **what:** Collection is pull-only: the serving box buffers events in JSONL and exposes them via key-gated HTTPS (/events, /access-digest, /stats); the cluster's scheduled collector pulls. No third collector VM and no push, because push or a middle VM would put DB/cluster secrets on the box and add attack surface + a hop for no gain. B2 remains an optional cold backup.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#6
- **relations:** rationale for /events + collector design; boxes hold only the read-only stats_key
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk topology exception — static page + always-on backend, not pure edge
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md`/`(10).md` "Architecture (note: not edge-only)"; `running_notes(44).md`: "Topology note: idea.uk is NOT pure-static/edge like the other chat domains."
- **what:** Every other "simple paid chat" domain on the platform is designed as static-S3 + a synchronous edge worker (no always-on compute). idea.uk breaks this pattern because its "tool" is a minutes-long multi-LLM + web-search job, not a synchronous chat turn: it needs a small always-on backend running the engine as a background task, with the static/embedded page posting to it and Stripe's webhook pointed at it. Flagged repeatedly as a deliberate, understood exception to the platform's default serverless-edge model, not an oversight — "the PAGE is serverless..., the SERVICE is NOT and can't be."
- **sources:** `RUNBOOK_idea_uk(1).md` "Architecture" section; `running_notes(44).md` ("Topology note: idea.uk is NOT pure-static/edge")
- **relations:** idea.uk deployment topology; service-deployer pattern
- **verify-later:** contrast against the actual edge-worker chat domains for confirmation

<!-- SOURCE: U26_misc_dirs.md -->
### Agent chassis — generic configurable agent executor
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md describes it as the running framework ("deploys as a scalable Kubernetes deployment, 3 replicas in production"); HITL agent definition (2025-11-03) still references image `docker.io/aqls/agent-chassis:v1.0.407`.
- **what:** A single reusable Go binary that becomes any agent type via configuration: it consumes Kafka messages, loads its workflow config from the database (agent_definitions / agent_instances), executes the workflow, handles fuel checks, errors, metrics, and health endpoints. New agent types are created by adding DB configuration, not code — "you're not creating new CODE, you're creating new CONFIGURATIONS".
- **sources:** docs/architecture/002-agent-chassis-docs.md; docs/architecture/023-spawning-agents.md#the-core-concept; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; agent spawning; distributed embedded orchestration
- **verify-later:** cmd/agent-chassis/main.go, platform/agentbase/, platform/messaging/processor.go, agent_definitions table

<!-- SOURCE: U26_misc_dirs.md -->
### Distributed embedded orchestration (no central orchestrator)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md: "every agent is both a worker and an orchestrator... eliminates single points of failure"; presented as a completed architecture report.
- **what:** Every agent pod embeds a full orchestrator (SagaCoordinator) instead of a central orchestration service. Any pod of an agent type can start a workflow, and any pod can pick up a response and continue it, because state is in the shared database. Key architectural decision distinguishing the platform from Temporal/Airflow-style central schedulers.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#distributed-orchestration-model; docs/architecture/003-flow-doc.md; docs/architecture/012-investors.md
- **relations:** stateless-first principle; database-backed workflow state; AI-native orchestration positioning
- **verify-later:** platform/orchestration/ (SagaCoordinator), orchestrations/orchestrator_state tables

<!-- SOURCE: U26_misc_dirs.md -->
### Database-backed workflow state (orchestrator_state → orchestrations)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md: "The orchestrator is stateless... Workflow state is in the database"; basic_usage/004 queries both the old `orchestrator_state` and a newer `orchestrations` table, showing the concept live through a schema evolution.
- **what:** All workflow execution state — status (RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED), current_step, workflow_plan, execution_path, collected_data, awaited_steps/awaited_requests, final_result — is persisted per correlation_id. Responses arriving at any pod are matched to awaited steps via causation_id, the state is updated, and the workflow continues when all responses are in.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/architecture/003-flow-doc.md; docs/basic_usage/004_debugging
- **relations:** stateless-first principle; fan-out and response correlation; workflow state machine
- **verify-later:** orchestrator_state and orchestrations tables in clients DB; column set differences between them

<!-- SOURCE: U26_misc_dirs.md -->
### Local vs remote actions and the action registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 004-agent-chassis-architecture.md documents both patterns as implemented ("Executed within the orchestrator itself" vs "Executed by other agents via Kafka").
- **what:** Workflow steps are either local actions run synchronously in the orchestrator (validate_input, transform_data, spawn_agent, process_data...) registered in a Go actionRegistry, or remote actions dispatched to another agent's Kafka topic with state moved to AWAITING_RESPONSES. The registry grew over time (spawn_agent, execute_llm_prompt, await_approval added later).
- **sources:** docs/architecture/004-agent-chassis-architecture.md#local-vs-remote-actions; docs/architecture/025-reusable-evolvable-agent-teams#step-3; docs/basic_usage/003_dynamic_prompt_improvement
- **relations:** agent-centric call_agent; fan-out; HITL await_approval
- **verify-later:** platform/orchestration/actions/ directory; actionRegistry in coordinator.go

<!-- SOURCE: U26_misc_dirs.md -->
### Fan-out and awaited-response correlation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003-flow-doc.md walks a live fan-out (reasoning + image agents) with correlation_id/causation_id header matching; 001basic_usage.txt shows fan_out steps in the deployed website-builder workflow.
- **what:** A fan_out step sends parallel sub-tasks to multiple agent topics, records their request IDs in awaited_steps, and sets status AWAITING_RESPONSES. Each response carries correlation_id (workflow) and causation_id (the originating request_id); any receiving pod matches causation_id to an awaited step, stores the result under collected_data, and resumes when all are received.
- **sources:** docs/architecture/003-flow-doc.md; docs/architecture/004-agent-chassis-architecture.md#response-handling-flow; docs/basic_usage/001basic_usage.txt
- **relations:** database-backed workflow state; kafka topic conventions; message header contract
- **verify-later:** fan_out action implementation; awaited_steps vs awaited_requests handling

<!-- SOURCE: U26_misc_dirs.md -->
### Kafka topic conventions (process/responses → requests/responses)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** Early docs use `system.agent.{type}.process` + `system.responses.{type}` (004); the stateless plan (v21) and HITL scripts (Nov 2025) use `system.agent.generic.requests`, per-type `.requests/.responses/.errors/.dlq` topics stored in agent_definitions.topics — the newer form names the older one's replacement.
- **what:** Naming scheme for per-agent-type Kafka topics plus system topics (`system.notifications.ui`, `system.commands.workflow.resume`, `system.errors.*`, DLQs). Topics are per agent TYPE, not per instance; all replicas share a consumer group so Kafka distributes work. The convention itself is durable; the specific `.process` form was superseded by `.requests/.responses`.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#kafka-topic-structure; docs/plans/stateless-first-agents-001#7-kafka-configuration; docs/humanintheloop/hitl_agent_definition.sql (topics JSONB)
- **relations:** stateless-first principle; HITL notification/resume topics
- **verify-later:** actual topic list in cluster; topics column of agent_definitions

<!-- SOURCE: U26_misc_dirs.md -->
### Message header contract (sender identity, in_response_to_*, status enum)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 (marked "v21", heavily iterated) defines the full header set; quick_hitl_test.sh (Nov 2025) sends live messages carrying orchestration_id, orchestration_name, step_name, message_type, from_agent_type, responses_topic headers — the contract in use.
- **what:** Rich request/response headers: sender AgentIdentity (agent_type, agent_id=pod name, version), correlation_id + human-readable correlation_name, orchestration_id/name, step_id/name, request_id, retry_version, parent orchestration linkage, message_id, fuel budget, timeout, routing topics. Responses echo in_response_to_request_id/step/orchestration and carry a status enum: awaiting | processing | complete | error_recoverable | error_unrecoverable, plus multipart flags, timing and fuel accounting.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/quick_hitl_test.sh; docs/humanintheloop/send_approval.sh
- **relations:** retry semantics; message deduplication; fuel budget
- **verify-later:** Go header structs in platform code; kafka message headers on live topics

<!-- SOURCE: U26_misc_dirs.md -->
### Stateless-first agent principle
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Core Principle: Agents are stateless executors. State lives in the database. Any replica can process any message for its agent type" — presented as the implementation spec (v21) that matches later operational docs.
- **what:** Agents hold no orchestration state in memory; pod crashes lose nothing; replicas scale horizontally with HPA (CPU + kafka consumer lag metrics); Kafka consumer groups distribute work; messages for one orchestration are ordered by using orchestration_id as the partition key. Formalises and extends the earlier distributed-orchestration model.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/plans/stateless-first-agents-001#8-kubernetes-deployment
- **relations:** distributed embedded orchestration; orchestration-as-identity; optimistic locking
- **verify-later:** deployment manifests (HPA config), consumer group setup, partition key usage

<!-- SOURCE: U26_misc_dirs.md -->
### Orchestration-as-identity model (AgentID = PodName)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001: "Orchestration (in DB) + Step + Request = 'Agent Instance'... AgentID = PodName (changes on restart, but that's OK)". This resolved the earlier mandatory-AgentID debate (022).
- **what:** The persistent identity of "an agent doing a task" is the orchestration record, not the pod. Pod name serves as AgentID purely for debugging (processing_history records which pod handled each step). Supersedes the doc-022 proposal that workflows resolve and pin specific versioned agent instances (stable/canary selection strategies) — that instance-pinning design was not carried forward.
- **sources:** docs/plans/stateless-first-agents-001#architecture-philosophy; docs/architecture/022-possible-agent-structure#the-case-for-mandatory-agentid
- **relations:** supersedes mandatory agent-instance resolution (022); stateless-first principle
- **verify-later:** whether processing_history with pod_name exists in current orchestrations table

<!-- SOURCE: U26_misc_dirs.md -->
### Optimistic locking on orchestration state
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Fully specified in stateless-first-agents-001 (version column, update_orchestration_if_version() SQL function, retry loop with backoff) but no later doc in this unit confirms it shipped.
- **what:** Each orchestration row carries a version integer; replicas load state, apply a step, and save only if the version is unchanged (compare-and-swap), retrying on mismatch. Prevents two replicas from double-processing the same step. Paired with processing_history JSONB as the audit trail of which pod did what.
- **sources:** docs/plans/stateless-first-agents-001#3-database-backed-state-management; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** stateless-first principle; message deduplication
- **verify-later:** version column and update function in current schema; conflict-retry code

<!-- SOURCE: U26_misc_dirs.md -->
### Retry semantics: same request_id, incremented retry_version
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** stateless-first-agents-001 "Key Implementation Notes: Retry uses same request_id with incremented retry_version"; error_recoverable responses trigger up to 3 retries. No later confirmation in this unit.
- **what:** Failed remote calls are retried with the identical request_id and retry_version+1 so responses remain matchable and duplicates detectable. Recoverable errors retry (max 3), then fall through to unrecoverable which fails the orchestration and propagates an error to the parent. Progress statuses (awaiting/processing) are logged but never propagated upward; terminal states are processed exactly once.
- **sources:** docs/plans/stateless-first-agents-001#6-retry-logic; docs/plans/stateless-first-agents-001#key-implementation-notes
- **relations:** message header contract; message deduplication
- **verify-later:** retry handling in response processing code; awaited_requests (request_id, retry_version) PK

<!-- SOURCE: U26_misc_dirs.md -->
### Message deduplication (processed_messages, terminal-state-once)
- **category:** system-architecture
- **status-signal:** unknown
- **status-evidence:** Designed in detail (dedup key request_id:retry_version:status; processed_messages table with 24h cleanup) in stateless-first-agents-001; no operational evidence in this unit.
- **what:** Before processing, agents check a dedup key against a processed_messages table (or in-memory map); duplicate responses are dropped, and once any terminal state (complete/error_unrecoverable) is processed for a request, all further terminal responses for it are ignored. Ensures idempotency under Kafka redelivery and multi-replica consumption.
- **sources:** docs/plans/stateless-first-agents-001#7-deduplication-handler; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** retry semantics; optimistic locking
- **verify-later:** processed_messages table existence; dedup logic in message consumption path

<!-- SOURCE: U26_misc_dirs.md -->
### Fuel budget resource management
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md lists fuel management as a chassis feature ("Checks fuel budget from headers... Prevents execution if insufficient fuel"); fuel_budget=1000 header sent in live kcat commands (basic_usage 001/004); response headers carry FuelUsed/RemainingFuelBudget.
- **what:** Every workflow carries a fuel budget header; actions deduct fuel costs; sub-invocations pass a reduced budget down and report fuel used back up the chain. Serves as the cost/abuse control across multi-agent workflows. Current status in the 2026 platform unverified — no recent doc in this unit mentions it.
- **sources:** docs/architecture/002-agent-chassis-docs.md#key-features; docs/basic_usage/001basic_usage.txt; docs/plans/stateless-first-agents-001 (FuelUsed/RemainingFuelBudget headers)
- **relations:** message header contract; subscription/quota API
- **verify-later:** fuel handling in chassis code; whether current work-item system retains fuel

<!-- SOURCE: U26_misc_dirs.md -->
### Agent-centric architecture: steps call agents, not topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 022-possible-agent-structure proposes `call_agent` with agent_type replacing raw topics ("your current code already does 90% of this"); 027 and the HITL group definition (Nov 2025) use call_agent steps in production seeds.
- **what:** The primary abstraction is the agent (owning a 6–12 step workflow) rather than the workflow; steps invoke other agents (`action: call_agent, agent_type: X`) which have their own workflows, error boundaries and state, enabling recursive hierarchies (any agent can orchestrate, a copywriter can spawn a researcher). Topic resolution happens from agent type.
- **sources:** docs/architecture/022-possible-agent-structure#summary-agent-centric-architecture; docs/humanintheloop/hitl_agent_group_definition.sql; docs/architecture/023-spawning-agents.md#the-orchestrator-is-a-pod-too
- **relations:** agent chassis; agent spawning; supersedes inter-agent invocation protocol v1
- **verify-later:** call_agent action and agent-type→topic resolution code

<!-- SOURCE: U26_misc_dirs.md -->
### Inter-agent invocation protocol v1 (invoke_agent / agent_invocations)
- **category:** system-architecture
- **status-signal:** superseded
- **status-evidence:** 001-agent-calls-agents-doc.md proposes InvokeAgentAction, ParallelInvokeAgentsAction, and an agent_invocations tracking table; later docs (022, stateless plan) replace this with call_agent + orchestration hierarchy headers, and the agent_invocations table never reappears.
- **what:** The first design for agent-calls-agent: a dedicated invocation request/response envelope, per-pair topics (`system.agent.requests.{from}.{to}`), an agent_invocations audit table, and parent_correlation_id columns. Its essential ideas (parent linkage, deadline, fuel passing) survived into the header contract; the specific mechanism did not.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#1.2; docs/architecture/001-agent-calls-agents-doc.md#phase-3
- **relations:** superseded by call_agent (022) and stateless header contract; project manager agent
- **verify-later:** confirm agent_invocations table absent from schema

<!-- SOURCE: U26_misc_dirs.md -->
### Project Manager / User Representative agent hierarchy
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Designed across 001 and 007 ("User Representative Agent... represents the users views against the project manager"); never appears in later seeds, groups, or the current 002-spine — silently vanishes after the website-builder group takes the orchestrator role.
- **what:** A top-level persona hierarchy: User → Project Manager agent (plans phases, delegates to specialist orchestrators, reviews deliverables) → Web Design Orchestrator → specialists, with a User-Persona agent negotiating on the user's behalf (stores preferences, approves/rejects deliverables). The review/approval intent resurfaced later as HITL steps and content governance instead.
- **sources:** docs/architecture/001-agent-calls-agents-doc.md#architecture-overview; docs/architecture/007-roadmap.md#2.1-user-representative-agent
- **relations:** website-builder group (took its place); HITL approval mechanism (absorbed the review role)
- **verify-later:** confirm no project-manager/user-representative agent_definitions exist

<!-- SOURCE: U26_misc_dirs.md -->
### HTML-first progressive enhancement delivery
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 008: "Starting with plain HTML/CSS/JS is actually a very smart architectural decision"; the html-developer seeds specify "vanilla" HTML with inline CSS; the current platform still renders plain HTML/CSS sites (render pipeline docs).
- **what:** Deliberate decision to generate plain HTML/CSS/JS websites rather than framework apps: easier for AI to generate and validate, no build step, universally hostable, fast; complexity added progressively (web components → PWA → framework only if needed). One of the few strategy decisions from this era that demonstrably survived into the present render pipeline.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#why-simple-html-css-js-is-the-right-start; docs/architecture/027-create-website-creation-system (html-developer config)
- **relations:** styling-render-pipeline (current successor context); WordPress export agent (rejected sequel)
- **verify-later:** current renderer output format

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow status state machine
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Consistent across eras: RUNNING / AWAITING_RESPONSES / COMPLETED / FAILED in 004; RUNNING / AWAITING_RESPONSE / COMPLETED / FAILED in HITL_README (Nov 2025); pending|processing|complete|failed variant in the stateless plan.
- **what:** The orchestration status vocabulary and its transitions: workflows run steps, park in an awaiting state while remote/human responses are outstanding, and terminate complete or failed. The HITL pause reuses the same awaiting state rather than introducing a special paused status. Minor naming drift across eras is itself a verification target.
- **sources:** docs/architecture/004-agent-chassis-architecture.md#database-architecture; docs/humanintheloop/HITL_README.md#workflow-states; docs/plans/stateless-first-agents-001#9-database-schema
- **relations:** database-backed workflow state; HITL approval mechanism
- **verify-later:** canonical status enum in current schema/code

<!-- SOURCE: U26_misc_dirs.md -->
### Human-readable orchestration and correlation names
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** stateless-first-agents-001 mandates orchestration_name / correlation_name alongside UUIDs ("website-build-agrivisionary", "core-mgr-website-flow-0902-1030"); start_hitl_workflow.sh generates ORCHESTRATION_NAME="eborg-content-approval-$(date...)" in practice.
- **what:** Every orchestration and correlation carries a generated human-readable name in addition to its UUID, propagated through headers and stored in state, so debugging and monitoring read as narrative ("which pods processed core-mgr-website-flow") rather than UUID archaeology.
- **sources:** docs/plans/stateless-first-agents-001#1-stateless-agent-architecture; docs/humanintheloop/start_hitl_workflow.sh
- **relations:** message header contract; kcat/db-inspector runbook
- **verify-later:** name-generation code; name columns in orchestrations table

<!-- SOURCE: U26_misc_dirs.md -->
### Agent teams: composite/family/service-agent patterns
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** 021 evaluates three options (PM pattern, peer-to-peer squads, service-oriented) and recommends service-oriented-then-squads; what actually shipped was the simpler agent-groups + call_agent model — the AgentFamily/SharedMemory and workflow-composition (sub-workflow) constructs never reappear.
- **what:** Design exploration for complex 50+-step workflows: composite agents (one external face, embedded sub-components), agent families with shared state and peer coordination, stateless reusable service-agents (date extractor, entity extractor) callable by any workflow, and workflows-invoking-workflows composition. Records the acknowledged framework limitation ("one agent = one workflow, flat orchestration, no concept of agent teams") that agent groups later addressed.
- **sources:** docs/architecture/021-current-framework-limitations; docs/architecture/022-possible-agent-structure
- **relations:** agent groups (the shipped resolution); agent-centric architecture
- **verify-later:** n/a — confirm no AgentFamily/sub-workflow constructs in code
