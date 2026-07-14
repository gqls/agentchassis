# Cluster: development-guide
Categories included: development-guide, new:workflow-authoring, new:prompt-composition, new:language-i18n


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Pre-flight "does this already exist?" discipline (Step Zero)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) opens with it as "the most important step", with the asset-deploy-agent 3-hours-wasted example
- **what:** Before creating any agent/action, search agent_definitions, the action registry, Go code, gate functions, and workflows for existing equivalents; document findings; never create without demonstrating no existing coverage. Extends to documentation claims: verify at point of use and date what you verify (`[checked YYYY-MM-DD]`).
- **sources:** 001_development_guide(5).md#Pre-Flight, #API Verification Reference, #Reuse Before Creating
- **relations:** canonical field-path helpers; assumed-helper build failures (016 additions)
- **verify-later:** platform/orchestration/actions/registry.go; agent_definitions table

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Canonical field-path resolution helpers (datahelpers) vs 18+ duplicates
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): "18+ functions that resolve dot-paths… Do not add another one"; cleanup of 9+ duplicates listed under "Not yet built"
- **what:** `ExtractNestedField/String/Map`, `GetFieldFromPath(WithDefault)` in datahelpers are canonical (with `.response` auto-unwrap); six named duplicates in the actions package must not be copied. There is no `ExtractStringSlice` — compose `ExtractStringListHelper(ExtractNestedField(...))`.
- **sources:** 001_development_guide(5).md#Field Path Resolution; 016_additions_assumed_helper_and_cross_module.md
- **relations:** assumption checklist item on assumed helpers
- **verify-later:** platform/orchestration/datahelpers/*.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Actions are the unit of work — no wrapper+core split
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "The wrong pattern" section with WriteSiteSpec example
- **what:** All action logic lives inside the `XxxAction` function; composition happens via workflows, not Go-calling-Go; exporting a "core logic" function creates a duplicate API surface. Also: don't create subworkflows in SQL — spawn sub-agents.
- **sources:** 001_development_guide(5).md#Core Design Principles
- **relations:** every-agent-is-an-orchestrator
- **verify-later:** grep exported non-Action functions in actions package

<!-- SOURCE: U01_docs024_numbered_core.md -->
### spawn→call pattern and target_role lookup
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5): "This is how every existing workflow does it"
- **what:** Agents are spawned (`spawn_agent`) then called (`call_agent`). `target_role` → findAgentByRole scans all collected_data keys (preferred); `agent_type` → findAgentByType only scans keys starting `spawn_` (a trap). Dynamic dispatch = fixed role + `agent_type_field` resolved from collected_data at runtime; no topic-construction bypass.
- **sources:** 001_development_guide(5).md#How call_agent finds the spawned agent, #Dynamic dispatch; 002(4)#Resolved Decisions 16
- **relations:** dispatch loop; wrapper-orchestrator pattern
- **verify-later:** spawn_actions.go, call_agent.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Wrapper-orchestrator pattern ("every pod-running agent needs a parent that spawned it")
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) full section; med-* wrappers and site-adoption-orchestrator live per 002(4)
- **what:** Agents get dedicated K8s Job pods only via spawn_agent from a parent. Anything reached via the generic entry point that does substantive work needs a tiny wrapper (spawn → call → complete) so real work runs in its own pod with clean logs; in-chassis work blocks shared pod slots. Canonical minimal wrapper: med-export-orchestrator. Map input fields individually (never `input_data: input_data`), mark caller-optional fields `?`.
- **sources:** 001_development_guide(5).md#Every pod-running agent needs a parent; 002(4)#Active agents note
- **relations:** topics model; site-adoption-orchestrator
- **verify-later:** agent_definitions rows for med-* orchestrators; spawnAgentKubernetesJobFromDefinition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Standardized input extraction (input_mapping / input_fields / ActionInputSpec) and the `?` optional suffix
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) three-layer table; resolver behaviour documented from code
- **what:** Caller maps data via input_mapping (dot-paths into collected_data), actions declare fields, Go extracts via ExtractActionInputs. `?` suffix on the destination key makes a mapping optional (silently skipped); unsuffixed fields hard-fail the call. In the dispatch loop, only site_id/domain/work_item_id may be non-optional; all spec.* mappings must use `?`.
- **sources:** 001_development_guide(5).md#Standardized Input Extraction, #Optional fields in dispatch loop
- **relations:** field name collisions; handler input path contract
- **verify-later:** ResolveInputMapping in coordinator.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Field-name collision via the nested-source loop
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) corrected wording: affects required AND optional fields; real section-editor content_data clobber
- **what:** ExtractActionInputs's late nested-source loop checks `current_page`, `rerender_pages`, `site_record`, `input_data` for any unresolved field name — so generic names (`content_data`, `sections`, `site_id`, `domain`, `status`…) can silently bind to the wrong source. Rules: new code avoids colliding names (prefix them); existing code left alone unless it bites; complex/array fields must never go in ActionInputSpec (read the config path directly).
- **sources:** 001_development_guide(5).md#Field name collisions; 016 §0 item 15 and §9 literal-key trap
- **relations:** section-editor clobber; resolve_internal_links review catch
- **verify-later:** datahelpers/action_inputs.go nestedSources

<!-- SOURCE: U01_docs024_numbered_core.md -->
### RAG actions and knowledge_base shared store
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): migration 082 applied "Not yet populated"; rag actions Go "produced but not deployed" then 009 lists "registered, not workflow-tested"
- **what:** `rag_lookup` (embed→vector search→top-k, trigram fallback) and `rag_index` (chunk→embed→store, SHA256 dedup) over a shared `knowledge_base` table (vector(768)); deliberately actions not agents until a knowledge-indexer orchestration is needed. Tool docs also target a knowledge_base `tool_docs` row.
- **sources:** 001(5)#RAG actions, #Agent vs infrastructure; 037#Where the docs live
- **relations:** agent-vs-infrastructure test; per-tool docs convention
- **verify-later:** rag_actions.go registered?; knowledge_base row counts

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Agent vs infrastructure boundary test
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) table (LLM logger no, Ollama provider no, rag actions no, knowledge-indexer future yes)
- **what:** Something becomes an agent only if it owns a domain, needs its own workflow, and benefits from independent spawn/debug. Otherwise it is an action or cross-cutting infrastructure.
- **sources:** 001_development_guide(5).md#Agent vs infrastructure
- **relations:** promotion pattern (002d)
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Specialist vs handler: the persistence boundary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) post-mortem; hit twice (page-content-writer HTML trapped in site_work_items.result)
- **what:** A specialist returns data to its caller; a dispatch handler must persist its own outputs (page_components, site_components, assets) and update status. Specialists used as handlers need a wrapper (page-build-handler wraps page-content-writer with plan/validate/save/deploy). Test: callable from CLI with site_id+domain and everything lands in the right tables.
- **sources:** 001(5)#Lessons Learned; 002(4)#Page Build Handler Pipeline
- **relations:** dispatch loop; handler contract
- **verify-later:** page-build-handler definition; handler agents' save steps

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Extended thinking config and the no-temperature-to-Anthropic rule
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) dated 2026-05-27 temperature note
- **what:** `budget_tokens` in ai_service enables extended thinking (thinking blocks skipped in parsing, +30-90s). Since 2026-05-27 the Anthropic client sends no temperature at all (Opus 4.7+ 400s on non-default; thinking incompatible); Ollama still honours it. Steer Anthropic via budget_tokens and prompts.
- **sources:** 001(5)#Extended Thinking Configuration
- **relations:** LLM config shadowing (temperature dead paths)
- **verify-later:** anthropic.go client options

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Loop mechanisms: dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) Appendix C full reference incl. production dispatch loop example
- **what:** Loops inject N×M steps into the workflow plan at runtime (`{loop}_iter_{N}_{substep}`), with setLoopVariable propagating iteration outputs to base names, per-iteration output suffixing, continue_on_error skipping, and LoopCompleteAction aggregation. Known hazards: the fast-response race fixed by the ErrLoopExpansionHandled sentinel; shared `loop_metadata` key; never nest loops — spawn a sub-agent instead.
- **sources:** 001(5)#Appendix C — Loop Mechanisms
- **relations:** dispatch loop pattern; O(K²) state-bloat failure (016)
- **verify-later:** loop_actions.go, loop_expansion_handler.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Authoring rules pack (schema-check, parameterised SQL, api_key_env_var, nil-guarded templates, code-fence stripping, error_step-in-config, text wrapped for write_site_spec)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "Summary of rules" 1–20, each backed by a dated bug
- **what:** The distilled 20-rule authoring discipline from the bug tally: `\d` live schema before SQL (dumps go stale; domain→pipeline, version_note→change_description); $1+params never {{.field}} in SQL; every LLM step needs api_key_env_var; {{if}} before {{range}} (query_database empty = null not []); run agent-def SQL before triggering (chassis silently runs empty workflow); strip markdown code fences before JSON parse; error_step only works inside step.Config; write_site_spec rejects scalars (wrap {"text": …}); to_jsonb('…'::text); verify fire-and-forget INSERTs actually land.
- **sources:** 001(5)#Appendix A + #Summary of rules
- **relations:** debugging assumption checklist; best-effort-needs-monitoring
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Orchestrator wrapper pattern for dedicated pod spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Spawning confirmed working (2026-04-23 trigger test) … processing_mode: 'orchestrator' at top level + agent_category = coordinator IS the combination that produces a dedicated spawned pod"
- **what:** To run work in its own spawned pod rather than one of the three shared chassis pods: an orchestrator wrapper (category=orchestrator, agent_category=coordinator, processing_mode=orchestrator at top level of default_config) with steps spawn_agent → call_agent(target_role, not agent_type) → complete_workflow, calling a worker (specialist, processing_mode=task). Input mapping maps fields individually with `?` suffix for optionals — never a whole `input_data` blob. File writes from non-spawned actions land on a random chassis pod and die with it.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f "Operational gotcha", #2.4h, #14 "Chassis action design patterns"
- **relations:** agent_definitions three-column semantics; training-data-exporter as reference implementation
- **verify-later:** training-data-export-orchestrator agent definition as the canonical example

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### agent_definitions three-column semantics (category / agent_category / status)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Caught three agent_definitions column semantics confusions" (2026-04-23); reference row improvement-loop = category=orchestrator, agent_category=coordinator, status=experimental
- **what:** `category` is free-text functional role; `agent_category` is CHECK-constrained to strategist/executor/analyst/integrator/coordinator/specialist (NOT orchestrator); `status` is lifecycle. Naïve writes put lifecycle values in the wrong slot. Also: ON CONFLICT must target (type, version) with `WHERE deleted_at IS NULL`.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4h, #14
- **relations:** orchestrator wrapper pattern
- **verify-later:** CHECK constraint on agent_definitions.agent_category

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Chassis action input conventions (ExtractActionInputs / input_data, dual registration)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Rewritten to use the canonical pattern … matches every other action in the codebase" (2026-04-23); registry gap bit on 2026-04-20: composition actions "had NO entry in GlobalActionRegistry … rejected as 'requires a topic'"
- **what:** New actions use `datahelpers.RegisterActionInputSpec` in init() + `ExtractActionInputs` (5-strategy cascade) rather than raw ExtractNestedFieldString; parameters flow via `CollectedData["input_data"]` because `{{.input_data.X}}` templating does NOT render for deterministic-action step config; every new action needs BOTH the InputSpec registration AND a GlobalActionRegistry entry with IsLocal:true; results land in collected_data under output_field, never final_result. Config literal numbers must be read with `datahelpers.GetIntField(params.StepConfig.Config, …)` — `inputs.GetInt` reads collectedData, not config.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#14; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#2; HANDOFF_2026-04-17_triage_and_component_linking.md#1
- **relations:** CollectedData architecture; field-name collision risk
- **verify-later:** datahelpers package; registry.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pgbouncer per-batch transaction discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** v3.0 "driver: bad connection" failure → "v3.1 split into per-batch transactions, batch size 500 → 100 … Per-batch commits worked" (2026-04-23)
- **what:** Long-held transactions through pgbouncer (transaction pool mode) are fragile — bulk work must commit per small batch (<1s each), never wrap a streaming job in one transaction. Companion rule: check RowsAffected on single-row UPDATEs and fail loudly instead of Warn+continue (v3.1's final UPDATE silently didn't land).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4i, #14
- **relations:** training-data export v3 evolution
- **verify-later:** pgbouncer pool mode config; v3.2 UPDATE handling

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Agent re-registration vs re-seed risk (DB row authoritative)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** RUNBOOK 07-06 night: "deploys bump `agent_definitions.updated_at` without overwriting `default_config` (the old prompt survived today's deploy)… the user confirms component-creator is a dynamically spawned/registered agent, not a YAML-seeded one. So the DB row is authoritative."
- **what:** A durability model for DB-edited agent prompts: chassis deploys re-register agent definitions (bumping `updated_at`) but do not overwrite `default_config`, so SQL edits to prompts survive deploys for dynamically registered agents. The residual risk is an in-code prompt template driving an upsert; the check is one grep for a literal fragment of the OLD prompt in Go sources — a hit means mirroring the edit in code, no hit means nothing can revert the row. (Earlier drafts had a heavier "seed check" over configs/deployments YAML; superseded by the user's confirmation.)
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-C; RUNBOOK_scheme_to_components(49).md (the earlier five-minute variant, family-delta); running_notes_scheme_to_components(55).md#Uf #Uj
- **relations:** component-creator prompt re-aim; 019 idempotent prompt migration pattern.
- **verify-later:** run the Step C grep; agent registration code path (upsert semantics on default_config).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item crafting conventions (real shapes, truthful provenance, dedup)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** w4b_04 comments: "crafted from the real rows … with truthful deviations noted: source 'manual' and created_by 'w4b_chrome_refresh' … lying in provenance columns costs later debugging"; every W6/W7/W8/W9 insert repeats the pattern.
- **what:** The discipline for hand-inserting `site_work_items`: copy the metadata of real rows produced by the owning code path (pipeline/severity/priority/handler_agent/status), deviate only truthfully in provenance columns (source=manual, created_by=<script name>), carry only spec fields the consuming workflow actually reads, and dedup check-first with a NOT EXISTS that mirrors `idx_swi_dedup` exactly (non-terminal statuses only — including 'unresolved', a status the index taught the thread it had missed). Item_key families are stable conventions: `page_rerender:<page>`, `chrome_refresh_rerender:<site_id>`, `needs_imagery:section:<scope_ref>:<key>`, `component_regen_rerender:<uuid>`, `section_data_*`. The check-first pattern is borrowed from CreateNeedsNewComponentItem.
- **sources:** w4b_04_trigger_item.sql; w7b_01_imagery.sql; w8_01_post_deploy_rebuild.sql; running_notes_scheme_to_components(55).md#Tb #Tc
- **relations:** rerender-pages v6; work-item claim/retry; scheduler-and-tasks.
- **verify-later:** idx_swi_dedup definition; site_work_items status vocabulary.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### House rules and standing preferences (the working contract)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated verbatim at the top of the journal ("Standing preferences (STRICT)") and in both HANDOFFs so fresh chats inherit them.
- **what:** The user's cross-thread working contract, treated as binding by every agent session: Go not Python; British English; plain language, no hype/flattery, banned words "perfect/critical/excellent", no congratulation; confirm live schema/data before asserting or writing SQL; schema-first (`\d` before SELECT/UPDATE); structural framework fixes over one-off patches; low risk appetite, reasonable step sizes, ≤1 question per reply; no summary documents unless asked; don't call fixes final; no `*-light`/`*-dark` component variants; keep runbook + journal current; honest caveats including correcting one's own reads ("corrections owned").
- **sources:** running_notes_scheme_to_components(55).md#Standing-preferences; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** running-notes journal discipline; orchestrator conventions.
- **verify-later:** n/a (convention, not code) — check for a canonical repo home for these rules.

<!-- SOURCE: U04_idea_uk.md -->
### Launch idioms: orchestrate vs work-item insert (and what each trigger does NOT do)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Confirmed from the production trigger scripts" (2026-06-20 iii); the 081c finding (no hand-rolled wrappers) cited.
- **what:** Two production ways work starts: (1) static agents are orchestrated by producing one Kafka message to system.agent.generic.requests (action=orchestrate, config.agent_type, full header set) via a one-off kcat pod; (2) dynamic handlers (page-build-handler etc.) cannot be orchestrated directly — INSERT a `site_work_items` row (status='triaged') and the running build-dispatch-loop claims and spawns them. Key caveat learned on idea.uk: the content triggers (rerender-pages / page-rerender / page-rebuild) never re-resolve composition — palette changes must go through needs_composition/needs_design. Deploy topology is likewise two-path: Go changes ship in the chassis image (roll agents to the tag), site HTML ships via the sites monorepo → Actions → B2.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Launch idioms); idea.uk/running_notes_checkpoint_tt.md (deploy topology + trigger mechanism); idea.uk/HANDOFF(13).md
- **relations:** composition re-resolve; scheduler/dispatch loop; debugging guide kcat sections.
- **verify-later:** 082_submit_domain_unified.sh; build-dispatch-loop definition.

<!-- SOURCE: U04_idea_uk.md -->
### Array-item field contract for the page-content-writer (item_fields fix)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "Prompt migration is already applied… But until a chassis image carrying the Go change is live, {{if .item_fields}} is always false… the applied prompt is inert on its own" (checkpoint tt, 2026-06-21).
- **what:** Root-cause class behind empty rendered sections (the 7 blank differentiator cards): the writer's prompt listed an array field with its type but never its per-element shape, so the model guessed item keys (title/body) that render empty against templates reading name/description. Fix has three coupled parts: plan_sections populates `ItemFields` on each llm_field_spec (Go); the prompt migration renders the exact per-item field list in both What-To-Write and the JSON skeleton (019_pcw_prompt_item_fields.sql — idempotent by sentinel, no broken intermediate state either deploy order); a render-time reconciler in v3_site_actions.go. Deploy order matters: chassis image first, then trigger.
- **sources:** idea.uk/019_pcw_prompt_item_fields.sql; idea.uk/running_notes_checkpoint_tt.md; idea.uk/README_assemble_bundle_idea_missing_sections.md (the bundled problem statement)
- **relations:** section-data reconciler; coordinator contract (sibling contract-mismatch class); diagnosis-loop bundles (the assemble-bundle invocation).
- **verify-later:** plan_sections_action.go ItemFields; whether the chassis tag carrying it shipped.

<!-- SOURCE: U05_content_quality_linking.md -->
### Sub-agent modelling conventions (agent_definitions row shape)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** running_notes_17(21) "Agent-modeling facts (from research-agent row + 003)"; internal_link_resolver_agent.sql embodies them.
- **what:** How a called sub-agent is modelled: workflow inside default_config.workflow (agent_definitions has NO processing_mode column — it lives inside default_config with timeout_seconds); agent_category specialist; input/output contracts required; templated topics (system.agent.{type}.process etc.); responds on the parent's responses topic; NOT-EXISTS-guarded idempotent seed SQL; image_repository/image_tag pinned to the batch image. Modelled on research-agent as the proven sibling.
- **sources:** internal_link_resolver_agent.sql; running_notes_17(21).md#agent-modeling-facts
- **relations:** internal-link-resolver; every-agent-is-an-orchestrator doctrine; 003 contracts.
- **verify-later:** research-agent/internal-link-resolver rows side by side.

<!-- SOURCE: U06_finetuning.md -->
### input_mapping semantics: call_agent-only; config dot-paths for local steps; key_path for loop items
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §2 "[verified-source] Local action steps do not resolve input_mapping… input_mapping is dead config"; 109b header "CORRECTS a load-bearing assumption: input_mapping is NOT live for (local-action) loop substeps."
- **what:** The coordinator honours `input_mapping` only for call_agent (building child input_data) and loop fan-out; on plain local action steps it is dead config. Local actions pull values via config keys whose values are dot-paths resolved from collected_data (`ExtractActionInputs` Strategy 0 / `resolveTemplateToken`); loop substeps read the iteration item via a config dot-path like `key_path:"ckpt_key"` (setLoopVariable puts the item in CollectedData) — using input_mapping there silently falls through to fallbacks (the dataset-key-presigned-40× bug). A proposed coordinator change to resolve input_mapping on local steps was deliberately withdrawn (D1: fix the caller, don't teach the framework a new behaviour for one agent's misuse). Optional mapping fields take a `?` suffix; missing required sources hard-fail (migration 103's fix).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D1,D2),#8; working/phase5/109b_fix_presign_one_loop_item_keypath.sql; working/phase5/103_call_data_preparer_optional_inputs.sql
- **relations:** output_fields contract; launcher workflow; loop_complete convention
- **verify-later:** coordinator ResolveInputMapping; input_mapping.go `?` semantics L101-128

<!-- SOURCE: U06_finetuning.md -->
### Child-result shaping: output_fields (plural) contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x: "extractWorkflowResult reads completeStep.Config['output_fields'] — PLURAL only… singular output_field… is never read → falls to the fallback branch that dumps every non-internal collected key"; migration 104 confirmed live.
- **what:** An agent's final result shape is governed by its complete step's `output_fields` array; the singular `output_field` spelling is silently ignored, producing a step-name-keyed fallback dump that breaks consumers' documented paths (`provisioning_result.provisioning_id` buried under `dispatch_provision.response.…`). The resolver auto-unwraps one `.response` per path part but never crosses arbitrary step-name keys. Fix taken at the def level (gpu-provisioner switched to plural + launcher mapping repointed, migration 104) after the user vetoed a chassis change; recorded as debugging-guide gotcha #23. Corollary rule: verify each call_* step's mapped source paths against the producer's REAL collected_data shape before firing anything that books a GPU.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-173x—180x; working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header)
- **relations:** input_mapping semantics; data-path verification runbook step 2b
- **verify-later:** extractWorkflowResult in coordinator; whether other defs still use singular output_field (thunder-reaper was named)

<!-- SOURCE: U06_finetuning.md -->
### Reply-topic derivation rules (own topic vs parent topic; two-level await)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) D4 + 2026-06-02 16:12 "D4 CONFIRMED live"; STATUS 06_04 reply-topic orphan fix "VERIFIED 2026-06-04 18:21".
- **what:** Two awaits, two topics: a child's intermediate adapter calls are awaited by the child's OWN coordinator on `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`); only the child→parent final notification uses `__parent_responses_topic__`. Dispatch actions that put the parent topic in an adapter envelope orphan the await (adapter replies where no one listens → infinite hang) — this bit twice (launcher dispatches pre-D4; `dispatch_thunder_ssh_get_status` cloned from ssh_exec). The inherited handoff asserted the opposite convention — corrected against source ("verify against code, not the handoff"). A shared `resolveAwaitResponsesTopic` helper is flagged as the future consolidation; a latent fallback caveat remains (the `system.agent.<type>.responses` fallback doesn't match the launcher's actual `system.responses.training-launcher` topic).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D4),#6,#10; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** send-before-register race; monitor build; adapter header tiers
- **verify-later:** determineResponsesTopic priority order in coordinator; whether the shared helper was built

<!-- SOURCE: U06_finetuning.md -->
### Orchestrator wrapper spawning pattern (dedicated pods for workers)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i "Spawning architecture — fully confirmed working"; worker pod agent-training-data-exporter-… observed.
- **what:** To run work in a dedicated spawned Job pod rather than the shared chassis pool: a wrapper agent with `processing_mode:"orchestrator"` at the TOP level of default_config, category='orchestrator' (free text), agent_category='coordinator' (CHECK-constrained — 'orchestrator' is not allowed), running `spawn_agent → call_agent(target_role=…) → complete_workflow`; the worker uses processing_mode:"task"/specialist. Includes the three-confused-columns trap on agent_definitions (category vs agent_category vs status; reference row improvement-loop) and ON CONFLICT (type,version). The monitor later reuses the same pattern in a loop (per-instance spawn+call, sequential).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i,#chassis-action-design-patterns; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data-exporter v3; monitor orchestrator; ExtractActionInputs canonical pattern
- **verify-later:** spawn decision logic in chassis (processing_mode placement)

<!-- SOURCE: U06_finetuning.md -->
### pgbouncer long-transaction fragility → per-batch commits
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i v3.0 failure "'bulk insert 500 rows: driver: bad connection'" and the v3.1 restructure; §14 pattern entry.
- **what:** Long-held transactions through pgbouncer (transaction pool mode) trip connection-level failures; bulk work defaults to per-batch commits (batch 100, each under a second) with single-statement non-tx bookends. Companion rule from the same incident: always check RowsAffected() on single-row UPDATEs and error rather than warn — an action can return perfect counts while its final UPDATE silently didn't land.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4i,#14; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data export v3.1/v3.2
- **verify-later:** training_data_export.go batch logic

<!-- SOURCE: U06_finetuning.md -->
### Reuse-first build discipline (grep before adding; delegate, don't parallel)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04 reuse audit ("ssh_get_status ALREADY EXISTS… no adapter change"); D3 "Reuse over parallel code in the adapter presigner"; guideline audits run against 001/002/003 for every artifact batch.
- **what:** A repeatedly-exercised discipline in this workstream: before building, audit what exists (ssh_get_status reused as the monitor probe; ListObjects reused for resume; DatasetURL/ArtefactURL refactored to delegate to ObjectURL rather than a third signer; datahelpers GetIntField over a custom helper; preRegisterAwaitedRequest reused for the race fix). Each new artifact batch is audited against the dev guide/architecture/contracts docs before deploy, with violations fixed or explicitly accepted (the one accepted tradeoff: launcher reading through the provisioner's step name).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#3(D3),#update-2026-06-04(reuse-audit),#guideline-audit,#update-2026-06-05(guideline-audit)
- **relations:** adapter design guide; input_mapping/output_fields contracts
- **verify-later:** n/a (practice); the accepted step-name coupling in call_launcher mapping

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F6 — dedup status-list mismatch and itemsCreated overcount
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "F6 flagged: the store's NOT EXISTS guard… omit 'unresolved' → Go guard STRICTER than index" (NOTES §9aa); unfixed Part E flag.
- **what:** Two small aligned defects: (1) the store's NOT EXISTS dedup status list and the `idx_swi_dedup` partial-unique predicate disagree on `unresolved` (index-terminal but guard-blocking — an unresolved squatter blocks createRerenderWorkItem where the index would not); (2) create_rerender_items increments `itemsCreated` without gating on RowsAffected, so ON CONFLICT DO NOTHING conflicts overcount the log. One-line alignments, parked.
- **sources:** NOTES(43).md §9t, §9aa; RUNBOOK(49).md Part E F6; HANDOFF(7).md §Flags
- **relations:** work-item dedup semantics; F3 (same action); hygiene: 40 stale unresolved items.
- **verify-later:** idx_swi_dedup definition vs createRerenderWorkItem NOT EXISTS list; create_rerender_items counter.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Work-item dedup semantics (item_key + idx_swi_dedup partial unique index)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (complete, verified, rejected, wont_fix, failed, unresolved)" — captured from pg_indexes (NOTES §9aa).
- **what:** site_work_items dedup is two-layered: producers guard with NOT EXISTS over open statuses and the DB enforces a partial unique index on (site_id, item_key) over non-terminal statuses. Terminal-status items free the key (why completed triggers can be re-inserted for retriggers); mirroring the producer's exact insert (columns, item_key scheme, dedup clause) is the established way to hand-create conforming items. See F6 for the known guard/index mismatch.
- **sources:** NOTES(43).md §9q, §9aa; RUNBOOK(49).md Part C Step 9b; w4b_03_read_rerender_config.sql
- **relations:** F6; work-item spec-cloning discipline; F3.
- **verify-later:** idx_swi_dedup definition; createRerenderWorkItem insert shape.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Deploy-ordering hard gate for coupled Go action + workflow-config changes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "LESSON (runbook gate tightened): 'deploy the Go action first' is insufficient — the migration is live instantly while the image rolls out. Hard gate: confirm… registered + live on ALL pods, THEN apply the migration" (NOTES §9i–§9j).
- **what:** Workflow jsonb changes take effect immediately; Go actions only exist once the image is rolled out and the registry entry (IsLocal:true) is in the running build. Wiring a workflow step to a not-yet-live action makes the validator reject EVERY run of that agent (WORKFLOW_INVALID broke all component generation during F2 3a). The codified gate: deploy + confirm the action responds on all pods before applying the (idempotent) migration; `revert_agent('<type>')` is the immediate mitigation.
- **sources:** NOTES(43).md §9i, §9j; F1prompt_component_creator_preserve_field_names(1).sql PREREQUISITE header
- **relations:** F1-prompt (where it bit); prompt-migration convention; snapshot/revert_agent.
- **verify-later:** workflow validator is_local check; revert_agent/snapshot_agent functions.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Prompt/workflow-jsonb migration convention (snapshot-first, anchored, idempotent, drift-checked; the 072 trap)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** The F1prompt migrations implement it end-to-end ("Convention: snapshot-first, idempotent, drift-checked, live-row only" — F1prompt(1).sql); prompt located top-level after `prompt_is_top_level=f` proof (NOTES §9d).
- **what:** Agent behaviour lives in default_config jsonb; edits are live instantly and follow a strict convention: snapshot_agent first; anchor the edit on a unique existing string and abort if the anchor count ≠ 1 (drift check); idempotency marker so re-runs no-op; filter to the live row (is_active, not snapshot, not deleted). The "072 nested-prompt trap": prompt_template may live at the top level of default_config OR nested in a step config — verify the path first or the migration is a silent no-op. Anti-drift prompt anchors have precedent (tool-doc-header rule on tool-improver): prompt rule = the anchor, store guard = the gate.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c, §9d, §9k (dead-block cleanup — an idempotency-check subtlety)
- **relations:** F1-prompt; F3c config edit; D2b-2 prompt edit; deploy-ordering gate.
- **verify-later:** snapshot_agent/revert_agent; component-creator prompt state (no dead {{if .existing_field_names}} block).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Never-guess disciplines: clone real work-item specs, look up real URLs (phantom-CTA lesson)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "NEVER guess a needs_page spec" (HANDOFF(6→7)); "MY CORRECTION: I had baked a guessed path… replaced with a pages.url lookup query (never guess paths; pages.url is the source)" (NOTES §9ae); spec-shape reads staged before every insert (w4b_03).
- **what:** A cluster of verify-at-point-of-use rules that recur across both threads: hand-created work items are mirrored column-for-column from a real captured item (SELECT-based inserts, real spec shapes, conforming item_keys), never composed from memory; URLs/paths come from pages.url, never invented (the phantom-CTA bug was an invented /contact.html; the recovery re-verified the same value before trusting it); schema before SQL (column names checked against \d, e.g. occurred_at not created_at); trigger flows through the real producer path rather than manual inserts where one exists (needs_design via build-site-planner / the proven 076 trigger).
- **sources:** NOTES(43).md §9l, §9ae, §9w–§9y, §9bd, §9bl–§9bm; w4b_03_read_rerender_config.sql; HANDOFF(7).md §Immediate next action
- **relations:** work-item dedup; section readiness (spec presence ≠ validity); F2 discriminators; link-management (phantom links).
- **verify-later:** n/a (convention; instances cited).

<!-- SOURCE: U08_travelling_docs.md -->
### Snapshot-before-update standing rule + the platform's snapshot_agent()
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Snapshot before updating (STANDING RULE 2026-07-06)"; write-location confirmed 2026-07-07 (stored OUTSIDE agent_definitions); every subsequent migration carries the call.
- **what:** `SELECT snapshot_agent('<type>', '<migration>: pre-update')` is prepended to every agent-updating migration. The function already existed (a reuse win — the drafted side-table migration was superseded un-applied; lesson: fetch-first applies to FUNCTIONS too, `\df` before drafting backup machinery). Snapshots live in a separate store (later identified as agent_definitions_backup in the FYI), so the defensive `is_snapshot` selector predicate is not load-bearing. Companion `revert_agent('<type>')` exists per 016b.
- **sources:** RUNBOOK_travelling_docs(38).md#§0-REF,#task-1; RUNNING_NOTES_travelling_docs(39).md#rev18,#rev19,#rev22; FYI_from_fixloop_2026-07-10…md
- **relations:** migrations system; correct-while-touching.
- **verify-later:** `snapshot_agent`/`revert_agent` function definitions; agent_definitions_backup table.

<!-- SOURCE: U08_travelling_docs.md -->
### Correct-while-touching norm (bounded repair of adjacent inert bugs)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Defined in the RUNBOOK mini-glossary, "Norm adopted in this chat, 2026-07-06"; exercised in migrations 125–146 (e.g. all ten of recreation's step-level error_steps corrected while adding its note tail; 146's cooldown fix).
- **what:** When a migration already modifies a workflow, it also fixes known-inert bugs in that SAME workflow (e.g. step-level `error_step` moved into config with original targets, dead keys deleted), declared explicitly in the file — bounded repair, no separate campaign, never copying the broken shape into new steps.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev23,#rev26,#tier-4-continuous
- **relations:** error_step-in-config; snapshot rule.
- **verify-later:** the declared correct-while-touching sections inside migrations 125–146.

<!-- SOURCE: U08_travelling_docs.md -->
### error_step mechanics — config-level placement, existing target, derive-from-next_step, loop corollary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b(7) §9 entry (live-validated ×5 in run-3); mechanism documented in 001 §16; dormant step-level instances corrected across tool-generator/fix agents by migrations.
- **what:** The coordinator reads only `step.Config["error_step"]` — step-LEVEL error_step is parsed but silently ignored. Once placed correctly, the target must name an EXISTING step or the coordinator fails the whole workflow (a typo converts a recoverable failure into a fatal one). Pattern: derive `error_step` from the step's own `next_step` read from the same row (convergence by construction, nothing guessed); `jsonb_set` does not create parents — COALESCE-merge config. Loop corollary: inside loop substeps, `error_step`/`then_step`/`fallback_step` are iteration-prefixed at expansion and must name substeps of the same loop; `continue_on_error: true` is the iteration-scoped alternative.
- **sources:** 016b_debugging_guide_7_3_(7).md#error_step-entry; RUNBOOK_travelling_docs(38).md#§8; RUNNING_NOTES_travelling_docs(39).md#rev9,#rev12,#rev13
- **relations:** docs-never-fail containment; correct-while-touching; loop_expansion_handler.go.
- **verify-later:** `routeToErrorStepOrFail`/`continueExecution` in coordinator; loop_expansion_handler.go prefixing.

<!-- SOURCE: U08_travelling_docs.md -->
### The seam rule — every prompt consuming a spec field must render it
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration 138 applied ("Mandatory Behaviour Requirements" section rendered from spec.interactive_features, marked as overriding the source); PLAN(6) rollout-outcomes bullet.
- **what:** A requirement can survive the analysis step and still be ignored at generation if the generation prompt never renders it: `analyze_tool` rendered `spec.interactive_features`, `recreate_tool` didn't, and Opus trusted the visible source HTML over requirements buried in a 20KB analysis JSON — faithfully recreating the bugs it was asked to fix. Rule: when adding a spec field, grep EVERY prompt_template that should render it; render requirements explicitly, marked as overriding the source.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45-run2; HANDOFF_2026-07-10…md#§4
- **relations:** economy-simulator case; "passed checks ≠ working"; doc_notes `seam` category.
- **verify-later:** recreate_tool prompt's Mandatory Behaviour Requirements section (migration 138).

<!-- SOURCE: U08_travelling_docs.md -->
### Manual kcat trigger scripts (084–087) — the canonical manual-orchestrate envelope
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** All four scripts exist, exercised repeatedly; 086/087 add DRY-RUN-by-default + single-line enforcement.
- **what:** A family of shell triggers producing `action=orchestrate` messages to `system.agent.generic.requests` with the house envelope (correlation/orchestration/request/message ids; `config.agent_type=<target>`). Encoded operational knowledge: target the spawn-wrapper orchestrator, not the agent directly (the SPAWNED pod gets GITHUB_READ_TOKEN via the spawn gate; an in-place run on a shared pod fails pre-fetch); REF explicit, never HEAD (user decision 2026-07-02); banners print effective subject/function as the go/no-go tell; DRY-RUN default with SEND=1; declared real side effects on the live trial site.
- **sources:** 084_TRIGGER_diagnose_v1(2).sh, 085/086/087 headers; RUNNING_NOTES_travelling_docs(39).md#rev10,#rev27; 086_input_data_recreate_economy_simulator.json
- **relations:** kcat line-delimited trap; env-prefix trap; spawn+call input shape.
- **verify-later:** whether these scripts were promoted anywhere canonical (they live in drafts/ and this docs dir).

<!-- SOURCE: U08_travelling_docs.md -->
### Manual spawn+call input shape — satisfy the contract top-level AND the workflow's input_data.spec.*
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b entry with the 081/082/083 trigger trilogy as evidence (contract violation vs empty-context generation vs correct).
- **what:** `call_agent` validates input_mapping output against the target's input_contract at TOP level, while workflows for work-item-driven agents read `input_data.spec.*` — so manual invocation must provide required fields both top-level and inside a spec object (or better, drive the agent via its designed work-item trigger). Flagged as a latent design smell: contract and workflow should agree. Companion: `store_generated_component` keys regeneration on the LLM's EMITTED function — a mismatched name INSERTs a stray duplicate.
- **sources:** 016b_debugging_guide_7_3_(7).md#spawn-call-entry
- **relations:** diagnosis subject threading (same contract rule); trigger scripts.
- **verify-later:** call_agent contract validation code; component-creator contract vs workflow paths.

<!-- SOURCE: U09_adoption.md -->
### Wrapper-orchestrator pattern (site-adoption-orchestrator)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "The wrapper-orchestrator pattern also landed here: site-adoption-agent … now runs under a site-adoption-orchestrator wrapper (spawn_adopter → call_adopter → complete)" (FOCUS_adoption_fidelity_and_variants, 2026-04-22).
- **what:** Every pod-running agent needs a parent that spawned it; coordination work (conditionals, HITL, spawn/call) stays in-chassis, substantive work (LLM, crawls, heavy DB) runs in a spawned pod. site-adoption-agent was the outlier running in-chassis and got a thin `site-adoption-orchestrator` wrapper modelled on med-export-orchestrator. Known sibling defect: all four med-* wrappers carry a broken `{"input_data": "input_data"}` double-wrap mapping.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed, old2/HANDOFF_2026-04-22#wrapper-orchestrator-pattern
- **relations:** adoption pipeline; baseline rule recorded in 001_development_guide
- **verify-later:** agent_definitions `site-adoption-orchestrator`, med-* wrapper input mappings

<!-- SOURCE: U09_adoption.md -->
### insertWorkItem two-strike rule and dedup slot
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Key enabling discovery (insertWorkItem, line 46319): a built-in two-strike rule — an item_key with ≥2 terminal attempts in 7 days is inserted as unresolved… anything <3h after a terminal item is suppressed; live dups blocked by ON CONFLICT (site_id,item_key) WHERE status NOT IN terminal."
- **what:** Platform-wide anti-churn machinery on work-item insertion: partial-unique dedup index on (site_id, item_key) over non-terminal statuses, a 3-hour suppression window after a terminal item, and a two-strike escalation to `unresolved`. Discovery checks and retriggers lean on this instead of implementing their own loop protection. Related facts: `needs_human_review` is non-terminal (holds the dedup slot); terminal set = complete/failed/verified/rejected/wont_fix/unresolved; use DELETE+INSERT not ON CONFLICT against the partial index.
- **sources:** running_notes_15(10)#part-8, HANDOFF_2026-06-09#key-references, FOCUS_directory_builder_and_list_components.md#schema-quirks
- **relations:** sectionless durability S1; work-item lifecycle
- **verify-later:** insertWorkItem in chassis; idx_swi_dedup definition

<!-- SOURCE: U10_imagery.md -->
### ExtractActionInputs Strategy-0 explicit dot-paths lesson
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "FIXED workflow-only: SQL_2026-07-12_asset_deployer_explicit_paths.sql… Standing lesson: give ExtractActionInputs actions explicit dot-paths; never trust the search" — with the dispatch-shape (`input_data.spec.*`) gap recorded, not fixed.
- **what:** ExtractActionInputs' aggressive recursive field search matched a stale `purpose` elsewhere in collected_data, so the sprite sheet deployed as a 900×900 hero-config JPG despite the child receiving purpose='sprite_sheet' — explicit Strategy-0 dot-path config values are resolved first and win. Latent siblings: items dispatched via build-dispatch-loop carry payload under input_data.spec.* which the explicit paths miss; historical spawned deploys may have silently used hero dimensions (May icons' file dims worth checking).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-31, SQL_2026-07-12_asset_deployer_explicit_paths.sql, HANDOFF_imagery_best_in_class.md#I2.1
- **relations:** dispatch input contract; deploy_image_asset defaults ("purpose":"hero").
- **verify-later:** asset-deployer deploy_asset step config paths; datahelpers extraction strategies.

<!-- SOURCE: U10_imagery.md -->
### Manual agent-trigger pattern (kcat orchestrate; never hand-roll spawn+call)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "the documented system.intake pattern is STALE (topic doesn't exist) — the working mechanism is the kcat trigger script pattern" (Turn 18); "Do NOT hand-roll spawn_agent+call_agent inline workflows" (Turn 26 lesson).
- **what:** Manually triggering agents means an `action=orchestrate` envelope to `system.agent.generic.requests` with config.agent_type + input_data — known-good for improvement-loop, webdesign-agent, rerender-pages. Hand-crafted inline spawn+call parents fail because the spawned child runs its workflow on INIT and idles before the call arrives; work destined for spawned handlers must route through work items + dispatch instead.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-18/#Turn-26
- **relations:** dispatch input contract; brand-head activation was the proving case.
- **verify-later:** 033_rerender_pages_trigger.sh precedent; system.intake topic absence.

<!-- SOURCE: U10_imagery.md -->
### psql read-only PreToolUse gate
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "added a PreToolUse permission hook (.claude/hooks/psql_readonly_gate.py)… tested against a 20-case matrix and proven live" (Turn 3, 2026-07-08).
- **what:** Agent-session tooling: a hook auto-approves read-only SELECT/`\d` psql via the exact kubectl-exec form while mutations still prompt the human — reducing friction for the DB ground-truth checks every session performs. Session auth expires ~daily (runbook A1 re-login ritual).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-3, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNBOOK_imagery_best_in_class.md#A1
- **relations:** context-bundle seeding; operator runbook rituals.
- **verify-later:** .claude/hooks/psql_readonly_gate.py; settings.local.json hook wiring.

<!-- SOURCE: U12_docs024_archives.md -->
### ExtractActionInputs nested-source collision affects required fields too (corrected scope)
*(merged from 2 independent findings)*
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Live `001_development_guide(5).md`'s opening consolidation note states it "supersedes the prior copy (which still had the older 'Field Name Collisions' wording — the nested-source collision affects **required and optional** fields, corrected here)"; live text: "This loop iterates the full field list — both Required and Optional. It does not distinguish between them."
- **what:** Two independent archived drafts (`old/older1/001h_development_guide_new_agents_v8.md` and `old/001_development_guide.md`) both claimed the nested-source lookup collision in `ExtractActionInputs` (an unmapped field silently matching `site_record.<field>`/`input_data.<field>`) applies only to optional fields. The live doc corrects this: the nested-source loop iterates the full field list regardless of Required/Optional status; required fields (e.g. `site_id`) carry the same latent risk, it's just usually masked because earlier resolution strategies (0-2) resolve them first. The live doc adds a "latent risk (required field)" example and recommends collision-free names (`target_site_id`) for new required fields, while leaving existing code alone unless it actually misbehaves.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Field Name Collisions"; old/001_development_guide.md#"Field Name Collisions"; docs024_key_docs_latest/001_development_guide(5).md#"Field name collisions"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Note on the target_site_id input field name"
- **relations:** ExtractActionInputs / datahelpers resolution cascade (Strategy 0-4); whole-blob input_data anti-pattern; target_site_id naming convention
- **verify-later:** confirm `platform/orchestration/datahelpers/action_inputs.go`'s nested-source loop still doesn't distinguish Required/Optional.

<!-- SOURCE: U12_docs024_archives.md -->
### Whole-blob input_data passthrough mapping (anti-pattern)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive presents `{"input_data": "input_data"}` as a working pattern used by three orchestrators; live: "It does not do what it looks like it does... map each expected field by name."
- **what:** A wrapper-orchestrator shorthand documented as valid in the archive. Live guide identifies it as broken (double-nests the caller's data) and replaces it with explicit per-field mapping using `?`-suffixed optional keys.
- **sources:** old/001_development_guide.md#"Standardized Input Extraction"; docs024_key_docs_latest/001_development_guide(5).md#"Map fields individually, not the whole input_data blob"
- **relations:** ExtractActionInputs nested-source collision; input_mapping `?` suffix convention
- **verify-later:** grep current agent_definitions for `"input_data": "input_data"` mapping still in use.

<!-- SOURCE: U12_docs024_archives.md -->
### Loop array-iteration internals (early investigation notes)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive's "Open Investigation" records partial early findings about `LoopAction`/`handleLoopExpansion`/`setLoopVariable`; the fully worked-out version is documented completely in `archive_april_26/014_loop_mechanisms_guide.md`, itself fully absorbed into the live dev guide as Appendix C.
- **what:** Records still-open Feb-2026 questions about how loop expansion and substep naming work internally, later fully resolved and documented.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Open Investigation"; archive_april_26/014_loop_mechanisms_guide.md
- **relations:** claim_work_item / self-spawning dispatch loop (above)
- **verify-later:** none — superseded by a complete, later document.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Roadmap-phases enforcement gap (routed to builder thread)
- **category:** development-guide
- **status-signal:** deployed (as a documented finding; fix owned elsewhere)
- **status-evidence:** "RECLASSIFIED: this is the builder thread's MAIN queue item now (item 6...)" (RUNBOOK(9)#2026-07-07 later still)
- **what:** Guidelines 001 (~lines 1503-1560) already define a Tier-3 roadmap-with-phases mechanism, but `082_submit_domain_unified.sh` has no `--roadmap` entry point and `build-site-planner`'s prompt has no else-branch for an absent roadmap — so absent a roadmap, phase constraints vanish rather than degrade gracefully. Confirmed in code as an absent decision point, and routed to the builder thread as a relay-wide fix rather than fixed by this workstream.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#ROOT CONTEXT FOUND IN THE DOCS, fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#2026-07-07 later still
- **relations:** abandoned pilot candidates
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner prompt roadmap_brief block

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: error_step must be inside step config
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "error_step goes INSIDE a step's config — step-level error_step is silently ignored (001 §16)" (RUNBOOK(10)#Inherited gotchas)
- **what:** A platform-wide workflow-authoring gotcha: a step's `error_step` must be nested inside that step's `config` object; a step-level sibling `error_step` key is silently ignored. Live dormant instances found in page-build-handler's several steps — flagged to be corrected whenever that workflow is next touched.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#COLLISION SURFACE
- **relations:** max_tokens placement gotcha (same family)
- **verify-later:** grep/inspect `error_step`; `config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: max_tokens must live inside ai_service
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "max_tokens at a step-config's root is DEAD CONFIG for execute_llm_prompt" (RUNBOOK(10)#BENCHMARK RUN 2)
- **what:** `execute_llm_prompt` reads `max_tokens` only from the agent's top-level config or from inside the step's `ai_service` block; a root-level step-config `max_tokens` is silently ignored and the Anthropic client defaults to 2048 output tokens. This capped the diagnose-agent verdict step at 2048 tokens through all five benchmark runs undetected, and truncated the fix-proposer's plan mid-JSON twice.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 2, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** error_step placement gotcha; fix-proposer's truncation failures
- **verify-later:** grep/inspect `execute_llm_prompt`; `max_tokens`; `ai_service`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: verify deployed contents against the pod, never tag/git
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "verify deployed contents against the POD binary, never the tag, never git" (NOTES(10)#Turn 23)
- **what:** A same-tag deploy trap: bumping source without bumping `IMAGE_TAG` means `rollout restart` reuses the cached image, so a reported "deploy" can silently ship a stale binary. The only reliable verification is grepping the running pod's binary for control strings. Caught the v1.0.1107→v1.0.1108 "first deploy" being a no-op.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Live gotchas
- **relations:** round-counting scope bug
- **verify-later:** grep/inspect `IMAGE_TAG`; `rollout restart`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: rebalance window after chassis restart
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "after make release redeploy-agents, wait for the chassis deployment to settle before firing a diagnosis" (NOTES(10)#Turn 9)
- **what:** Firing an orchestration within roughly 300 seconds of a chassis pod restart risks the spawn's init response falling into a Kafka consumer-rebalance window and dying silently — cost 8 hours of debugging the first time. Standing workaround: wait ~300s after any deploy before firing a run.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 9, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** same-tag deploy trap
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: BST/UTC timestamp mismatch
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "orchestration_states.last_activity is timestamp WITHOUT time zone... dev host runs BST while DB is UTC" (RUNBOOK(10)#Inherited gotchas)
- **what:** `last_activity` is stored without time zone while `created_at` is timestamptz, so `NOW() - last_activity` arithmetic is silently wrong by the local UTC offset; combined with the dev host running BST against a UTC database, a run can appear to have finished before it started.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 6
- **relations:** rebalance window gotcha
- **verify-later:** grep/inspect `last_activity`; `created_at`; `NOW() - last_activity`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Observability signature-fields pattern (proving which code path ran via collected_data)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** design_actions_observability_patch.md code diff applied; FOCUS_visual_pipeline doc shows the fields live in the result map
- **what:** A reusable debugging convention: when patching a code path whose old/new behaviour is otherwise indistinguishable, write new marker fields into the result map (flowing into `orchestration_states.collected_data`) that are absent from the old code — their presence proves the new code executed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Observability, js_snippets_news_gaswholesalers/old/design_actions_observability_patch.md
- **relations:** CSS component-list fallback bug
- **verify-later:** grep/inspect `orchestration_states.collected_data`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### needs_section_data review items appearing on successful builds (open question)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** "Worth a separate look at why a successful structured build still raises a section-data review item" — listed open, never investigated
- **what:** Even the clean `faq-test` isolated build spawned a child `needs_section_data` work item with `status=needs_human_review` and no `handler_agent`, unexplained.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** Post-build validation of structured components; isolated build test methodology
- **verify-later:** needs_section_data work-item creation path, wont_fix auto-resolution pattern

<!-- SOURCE: U14_docs019_runbooks.md -->
### Cross-module engine port procedure
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4c "surfaced FOUR build errors in the first real make build-core-manager … None of these four were logic bugs … doing Steps 1–4 in order pays it once."
- **what:** The validated sequence for moving a package between Go modules (contextkit internal/diagnose → chassis pkg/diagnose): (1) copy the WHOLE package as a unit and diff file lists; (2) rewrite the moved-package import path everywhere; (3) build+test the package alone before the binary; (4) grep every shared-package call the new code makes against the REAL helper surface (datahelpers.ExtractStringSlice didn't exist; compose ExtractStringListHelper(ExtractNestedField(...))). The chassis copies keep the agentchassis import (step.go) while the prototype keeps contextkit's.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4c; docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 build note)
- **relations:** ReadSymbolBody dual placement; module-copy drift
- **verify-later:** pkg/diagnose file list vs contextkit/internal/diagnose

<!-- SOURCE: U15_docs019_running_notes.md -->
### Curated best-in-class standing expectation
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Best-in-class/curated-list idea homed: standing expectation (guides+tools+news+non-affiliate curated top-N) + 'not-original-can-still-be-best' clause → 001_development_guide" (NOTES_running_synthesis_v4(39).md, 2026-07-07).
- **what:** A proposed platform-wide addition to the development guide requiring every commerce-shaped domain to carry a baseline of guides, tools, a news feed, and a curated non-affiliate top-N list — enforced the same way as the roadmap gap (relay-wide strategist/planner prompts, not per-message or the constitution) — with the explicit doctrine that "useful-but-unoriginal still counts as best-in-class."
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-07 "pilot candidate 2" entry; NOTES_running_fixloop(9).md "Builder queue item 7" references.
- **relations:** Roadmap-phase enforcement gap; diagnosis→fix loop workstream founding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Building-discipline edge cases (pre-registered engineering checklist)
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Self-referential structures need a cycle guard... A multi-step apply must be all-or-nothing... Bulk operations need bulk confirmation" (principles(59) §Building discipline).
- **what:** A checklist of edge cases the design docs insist be caught before building anything with self-modifying or autonomous behaviour: cycle guards on any parent-link/version-chain walk; transactional multi-write-plus-event apply (outbox pattern) so a crash can't leave a dangling row with no event; reading a consistent point-in-time snapshot when assembling from multiple tables; "one live thing per target, all the way down" (dedup at every layer, not just the queue); bulk-confirmation for large batches; filtering transient/infra failures out of any trust-affecting evidence signal; and "tell not-yet apart from broken" (missing-because-unonboarded degrades gracefully, missing-because-malformed fails loudly).
- **sources:** NOTES_running_synthesis_principles(59) §Building discipline (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; untested-code / behaviour-testing discipline.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Untested-code / behaviour-testing discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "PRINCIPLE 2026-06-13: untested code is a liability, surfaced by the dedup -move bug... COMPILE/gofmt/vet prove syntax, NOT behaviour." (principles(59)).
- **what:** A hard-won, explicitly codified lesson (triggered by the dedup tool's silent destructive-flag bug) that compiling/gofmt/vet only prove syntax, never behaviour; any destructive CLI operation must be report-only by default; and Go's `flag.Parse()` stopping at the first positional argument is a specific, recurring footgun requiring manual value-flag-aware argument separation in every CLI that takes a positional followed by flags — audited across `dedup`, `thin_versions`, `resolve_targets`, `embed`, `assembler`, `fuse`, `eval_targets`, `dbcontext`.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 "PRINCIPLE" entry; NOTES_running_synthesis_v2(36).md 2026-06-14 "runbook code audit" (a second, independent instance of the same class of bug).
- **relations:** Docs archiving toolchain (dedup); building-discipline edge cases.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Cross-module copy-drift lesson
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "This is the 5th instance of the same root pattern (drafted/validated artefact vs what's on disk): import path, missing file, stale sibling, assumed helper API, now stale CLI." (NOTES_running_synthesis_v2(36).md, 2026-06-17).
- **what:** A hard-won lesson from porting the diagnose engine into the chassis across a real module boundary, surfacing five DISTINCT failure classes in sequence — a wrong import path on a copied file, an entire file silently omitted, a stale (pre-refactor) copy of a sibling file, an assumed helper-package API that didn't actually exist (`datahelpers.ExtractStringSlice`), and a stale CLI binary predating a library change — all of which passed silently in the source module and surfaced only on first build/run in the target. Durable prevention recorded: copy the WHOLE package directory as one unit and diff the file list, rather than cherry-picking files across versions; grep every shared-package call against the real package before authoring, not after.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 (four consecutive build-gap entries); mirrored in `016_additions_assumed_helper_and_cross_module.md` per the notes.
- **relations:** Diagnosis-loop chassis integration architecture; untested-code / behaviour-testing discipline.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Schema-before-SQL discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Schema: bundle CODE names tables; only \d gives columns AND persistence (hit 4×: page_id, no status column, file_count, the 3-NULL workflow columns)." (v2(36) STATE DIGEST standing lessons).
- **what:** A recurring, explicitly named discipline that code reliably names which DB tables are involved, but only a live `\d` (schema dump) gives real column names and reveals whether a field is persisted at all vs. computed at runtime — hit repeatedly across this project (a wrong `page_id` column, an assumed-but-nonexistent `status` column on `site_plan_sections`, a wrong `fileCount` vs `file_count` JSON key, and the workflow-column misassumption) and eventually generalised into "real rows/examples beat prose/inference" as its own standing lesson.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST "Standing lessons"; NOTES_running_synthesis_principles(59) multiple 2026-06-14 gamesdesign-diagnosis entries.
- **relations:** Workflow default_config location convention; DB discipline / snapshot_agent convention; gamesdesign silent-no-op bug.

<!-- SOURCE: U15_docs019_running_notes.md -->
### "Every agent is an orchestrator" spawn pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "ORCHESTRATOR vs long-running service (user's Q): follow 'every agent is an orchestrator' — a thin diagnose-orchestrator spawns a diagnose-agent worker pod... exactly as site-adoption-orchestrator spawns site-adoption-agent." (v2(36), 2026-06-17).
- **what:** A standing platform convention that any substantive in-chassis work (multiple iterations, LLM calls, minutes of runtime) must be wrapped: a thin coordinator/orchestrator agent spawns a dedicated worker agent as a Job pod that runs the actual work and replies to the caller's own responses topic, rather than building a bespoke long-running service that would duplicate the chassis's spawn/await/topic machinery.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "chassis integration... STEP ZERO search."
- **relations:** Diagnosis-loop chassis integration architecture; workflow default_config location convention.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Real-rows-beat-prose-or-assumption discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Real rows/examples beat prose/inference: the agent_definitions workflow lives in default_config — the dev guide prose didn't say which column; the row did." (v2(36) STATE DIGEST standing lessons).
- **what:** A generalised standing lesson (distilled from several specific incidents across this file set) that when a dev-guide or design doc's prose is ambiguous or silent about an implementation detail, the correct source of truth is a real, live example row/file, not inference from the prose — repeatedly the deciding move that caught a wrong migration or wrong action draft before it was applied.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST; NOTES_running_synthesis_principles(59) "GROUNDED... tools/provenance/docs design corrected" entries (same pattern, different subsystem).
- **relations:** Workflow default_config location convention; schema-before-SQL discipline; child-completion result key convention.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Workflow lives in default_config, not the workflow columns
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NNN_seed_diagnose_agents(2) header: "confirmed: those agents put their workflow there; the task_workflow / orchestrator_workflow columns are NULL"; the move/fix/restore migration sequence applied.
- **what:** The loader reads `agent_definitions.default_config` ({workflow:{start_step,steps}, processing_mode, timeout_seconds}); the three workflow columns are dead for working agents. A workflow seeded into orchestration_workflow silently never loads — this bit the diagnose pair (seeded 2026-06-20 into the wrong column, then the orchestrator's workflow was lost entirely during the move and had to be re-seeded). Key correction learned by reading a real row rather than docs.
- **sources:** NNN_move_diagnose_workflow_to_default_config(1).sql; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17)
- **relations:** schema-before-SQL discipline; spawn-consumed columns lesson (sibling class)
- **verify-later:** workflow loader code path; whether the three columns are still consulted anywhere

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirmed chassis workflow model (group agents, promotion pattern, wrapper orchestrator)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration(19) "Confirmed model (from the guideline docs …)" plus 2026-06-09 changelog confirming against two real agent_definitions rows and the action code.
- **what:** The model the migration work verified: workflows are declarative JSON steps in default_config; a generic config-driven action library (query_database, spawn_agent, call_agent, loop, conditional, rag_lookup, work-item lifecycle) is reused before writing Go; LLM-assisted checks group into one agent per shared context load (explicitly rejecting a registry of mini-action agents), promoted to spawned sub-agents only when one needs independence (a one-line workflow change); the wrapper orchestrator (spawn→call→complete) is the canonical small form; spawning is spawnAgentKubernetesJobFromDefinition with per-spawn job topics. Reuse discipline is encoded as queries (search agent_definitions; default_config::text ILIKE '%<action>%').
- **sources:** PLAN_workflows_and_actions_migration(19).md (confirmed model + changelog); DESIGN_diagnosis_loop_chassis_integration(6).md#0
- **relations:** diagnose pair; index-orchestrator; onboarding agents (all reuse it)
- **verify-later:** 001/002 guideline docs (canonical home, other units)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Development Guide (agent-build daily reference)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** 001(0) consolidation note "This is the canonical 001_development_guide. It supersedes the prior copy"; archive copy with live successor in docs024_key_docs_latest
- **what:** The consolidated practical reference for building/debugging/maintaining agents: core design principles (agents own their domain, callers pass raw data, workflows simple with complexity in Go, actions are the unit of work, spawn-before-call, reply-to-caller's-topic), a new-agent checklist, migration guide, and 20+ lessons-learned bug entries.
- **sources:** WM/001_development_guide(0).md#core-design-principles, WM/001_development_guide(0).md#checklist-for-new-specialist-agent, WM/001_development_guide(0).md#summary-of-rules-for-the-dev-guide
- **relations:** superseded by docs024 live 001; STEP ZERO; wrapper-orchestrator; loop mechanisms
- **verify-later:** platform/orchestration/actions/*

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### STEP ZERO — reuse-before-create discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "Pre-Flight: Does This Already Exist? … Real example: We built asset-deploy-agent … The existing asset-deployer already did the same thing. Three hours wasted"
- **what:** The mandatory pre-flight before creating any agent/action/function: search `agent_definitions`, the action registry, Go funcs, gate functions and workflows for every noun in the proposed name, document what was found, and prefer patching an existing thing. Includes the canonical field-path resolution rule (use `datahelpers.ExtractNestedField*`, don't add another).
- **sources:** WM/001_development_guide(0).md#pre-flight-does-this-already-exist-step-zero, WM/001_development_guide(0).md#field-path-resolution-use-the-canonical-functions, WM/001_development_guide(0).md#reuse-before-creating
- **relations:** Development Guide; STEP-ZERO-for-standards (curation); reliability cascade (reuse tier)
- **verify-later:** registry.go; datahelpers package; isStorageEnabledAgent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Standardized input extraction (ActionInputSpec, ? optional, field collisions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "The three layers … input_mapping / input_fields / ActionInputSpec"; "Field name collisions … runs a nested-source loop late in its resolution chain"
- **what:** Three layers move data into an action: caller `input_mapping` (with `?`-optional destination keys), action `input_fields`, and `ExtractActionInputs(spec)` with a documented resolution chain. The nested-source loop iterates required AND optional fields, so names like `site_id`/`content_data`/`domain` can silently resolve from the wrong nested source — prefer collision-free names.
- **sources:** WM/001_development_guide(0).md#standardized-input-extraction, WM/030_phase1_plan_and_reconciler(4).md#note-on-the-target_site_id-input-field-name, WM/016_debugging_guide_v2_44.md#0
- **relations:** dispatch loop input_mapping; ? suffix; target_site_id convention
- **verify-later:** datahelpers/action_inputs.go; ResolveInputMapping; coordinator.go resolveInputMapping

<!-- SOURCE: U17b_docs019_gofiles.md -->
### thin-slice constitution (always-on rules doc)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Included in full in every bundle... Later it becomes the `standards` rows with `scope = constitution`; the content is the same." (thin_slice_constitution.md)
- **what:** The flat-file version of the chassis's always-on rules, pasted into every assembler/bundle output: reuse-before-recreate, fix-structural-not-symptoms, every-agent-is-an-orchestrator, no subworkflows-in-SQL (spawn sub-agents instead), the snake_case/kebab-case naming split, storage conventions (text+CHECK enums, version+previous_version_id, deleted_at soft-delete), logging rules (no `logger.Debug`, log the orchestration_id/correlation_id), deployment path (GitHub → Actions → Backblaze S3), and plain/pragmatic generated-text tone. Task-specific 003 contracts are listed but pulled in only when a task touches them.
- **sources:** contextkit/thin_slice_constitution.md
- **relations:** assembler (always includes it), docselect.go (adds the task-specific 003 sections this doc defers), contracts-and-standards (003)
- **verify-later:** whether the constitution has since actually migrated to `standards` rows with `scope = constitution` as the doc anticipates, or is still the flat file

<!-- SOURCE: U18_sql_for_agents.md -->
### v1 monolithic LLM-chain site builders (website-builder, domain-analyst, site-architect, content-creator, html-developer, multipage-wrapper, html-assembler, site-deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026_pageflow_builder.sql renames multipage-website-builder v3 → pageflow-builder ("Component-based website builder... uses DB components for structure, LLM only for content"); v2/027_old_agent_definitions.sql captures the whole v1 fleet as "old"; root files never patch these agents again.
- **what:** The first architecture (2025-11/12): a website-builder orchestrator spawns a chain of one-LLM-call specialists — domain-analyst (audience/tone JSON), site-architect (page structure + colours), content-creator (copy JSON), html-developer (whole-page HTML), multipage-wrapper (file map), site-deployer (git commit). Everything is free-form LLM output; no component library, no DB page records.
- **sources:** sql_for_agents_v1/004_website_builder.sql; sql_for_agents_v1/005_domain_analyst.sql; sql_for_agents_v1/008_html_developer.sql; sql_for_agents_v2/027_old_agent_definitions.sql
- **relations:** superseded by pageflow-builder (component-based); site-deployer contract survives in 011_site_deployer.sql
- **verify-later:** agent_definitions rows for these types (is_active, deleted_at); whether any workflow still references them

<!-- SOURCE: U18_sql_for_agents.md -->
### Batched multi-page generation (multipage-website-builder) and chunked HTML generation
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026 renames its v3 to pageflow-builder; the batch-of-4-pages prompts (generate_batch_1..5) appear only in v1/015 and v1/017 snapshots.
- **what:** Anti-token-limit strategies from the v1 era: build 20-page sites by generating pages in five batches of four ("Return as JSON map of filename to HTML"), with shared CSS generated once and injected at assembly; html-developer-chunked generated structure/styles/sections in separate calls. Both are ideas the component architecture made unnecessary.
- **sources:** sql_for_agents_v1/015_example_20_page_workflow.sql; sql_for_agents_v1/017_multipage_website_builder.sql; sql_for_agents_v1/014_html_developer_chunked.sql
- **relations:** replaced by pageflow-builder per-page loop and later the one-item-per-run dispatch loop
- **verify-later:** none needed — historical

<!-- SOURCE: U18_sql_for_agents.md -->
### Remove-loops plan: input_mapping, contract validation, sequential_fan_out, page-builder worker
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** input_mapping conversion executed in 030/030b/023; contracts added across many files ("These contracts define what each agent expects... fail fast"); but build_pages_loop still present in 026 and sequential_fan_out/page-builder never appear in any later agent file — that half is effectively abandoned in favour of the dispatch loop.
- **what:** A four-phase plan to replace loop/substep injection: (1) explicit `input_mapping` instead of `input_fields` path-hunting plus runtime input/output contract validation with hard fails and `__raw_message__` deprecation; (2) a `sequential_fan_out` action spawning one child orchestration per page; (3) a page-builder worker agent; (4) rewire pageflow-builder. Phases 1 landed; phases 2–4 were superseded by the site_work_items dispatch-loop architecture, which achieves the same "one visible orchestration per unit of work" goal differently. 001_validator_sql.sql is a jsonb_path_query audit extracting every field path referenced in workflows.
- **sources:** 001_remove_loops_in_workflow.md; 001b_implementation_plan.md; 030_input_mapping_changes.sql; 030b_remaining_agents_needing_input_mapping; 001_validator_sql.sql
- **relations:** input contracts appear in nearly every agent file (002, 011, 022, 024, 025, 029...); dispatch loop (051) is the spiritual successor of sequential_fan_out
- **verify-later:** chassis code: contract validation enforcement; whether `sequential_fan_out` action exists in the registry; `__raw_message__` fallback removal status

<!-- SOURCE: U19_sql_tables_components.md -->
### site_work_items unified work queue and lifecycle
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full DDL (023_site_work_items) with dedup index, plus dozens of live operational patches (resets, handler re-routing, attempt bumps) against real sites.
- **what:** Every piece of platform work is a row: source (planner/discovery/content_feed/manual/improvement/side_effect/human/validation), pipeline (originally `domain`, later renamed), item_type, severity, spec JSONB, target refs (page/component/entity/url), triage enrichment (impact, resolution_path, suggested_action, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete/pending_verify/verified/failed/rejected/wont_fix plus 'blocked' (handler missing), dependencies (depends_on UUID[], parent/related/batch), attempts, and deterministic item_key with a partial unique index for dedup among non-terminal items. A same-structure archive table receives terminal items.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#approval_mode; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle
- **relations:** dispatch loop; claimed-item timeout; content_feed_items; archiver; approval_mode; processing_tier.
- **verify-later:** current status distribution; pipeline column rename (`domain` dropped in 018).

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item archival
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** archive_completed_work_items(age, batch) function + archiver agent definition + daily scheduled task, with schema-sync ALTERs and FK handling (parent self-ref cleared, content_feed_items references deleted).
- **what:** Terminal work items (complete/failed/wont_fix) older than a configurable age move to site_work_items_archive in batches, keeping the live queue small. Function handles column drift between live and archive tables explicitly.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-archiver
- **relations:** work queue; scheduler.
- **verify-later:** archiver task enabled; archive row counts.

<!-- SOURCE: U19_sql_tables_components.md -->
### build_queue site seeding
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Phase-0 Block A table with direction semantics enumerated.
- **what:** Domain-level intake queue for new sites: a row per domain with direction JSONB (null | {objective} | {adopt_from} | {fork_from} | {brief_complete...}), status and priority. seed_build_queue reads it, creates site records and initial work items according to direction — the entry point into the work-item pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#2
- **relations:** work queue; adoption pipeline (adopt_from); onboarding-config (brief_complete).
- **verify-later:** seed_build_queue action; build-pipeline-trigger seeding behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Loop-action dispatch (migration 071)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Applied UPDATE of build-dispatch-loop default_config: "Step-chaining... processes only one work item per trigger. The loop action is proven in maintenance-triage and pageflow-builder."
- **what:** The dispatch loop loads all dispatchable items upfront (dependency-filtered, priority-ordered, max 50) and iterates with the `loop` action running a sub_workflow per item: claim → check_claim → spawn_handler (dynamic agent type from current_item.handler_agent) → call_handler → mark_complete/mark_failed, with continue_on_error. Introduces item_variable scoping (current_item.*) and optional `?`-suffixed input_mapping fields silently skipped for handlers that don't need them (section-editor compatibility).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql#migration-071
- **relations:** work queue; spawn-orchestrator pattern; claimed-item timeout.
- **verify-later:** build-dispatch-loop live config; loop action implementation.

<!-- SOURCE: U19_sql_tables_components.md -->
### Spawn-orchestrator thin-wrapper pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated across med pricing (scrape, discover, map, export orchestrators): spawn_agent (role) → call_agent (target_role, input_mapping passthrough, timeout) → complete_workflow; scheduled tasks target the orchestrator, not the worker.
- **what:** The standard shape for burst workloads: a permanently-resident category pod receives the trigger, a thin orchestrator workflow spawns a temporary worker pod of the right agent_type, forwards input_data, awaits the result, and completes — worker terminates (idle_timeout 0). Non-secret worker config rides env_vars on the agent definition; secrets come via spawn_actions secretKeyRef.
- **sources:** docs/agent_docs/sql_for_tables/034_vet_med_price_scrape_orchestrator.sql; docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql; docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** scheduler; agent definitions; vet med pipelines.
- **verify-later:** spawn_agent/call_agent actions; ON CONFLICT (type, version) upsert convention.

<!-- SOURCE: U20_legacy_docs_a.md -->
### spawn_agent — database-definition-driven Kubernetes Job spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.042: "All deployment specs now come from database… No Code Changes for New Agents — Just insert into agent_definitions"; spawn walked through line-by-line in spawn_actions2.
- **what:** SpawnAgentAction reads the child's agent_definitions row (image repo/tag, resources, health config, env vars, default workflow), inserts an agent_instances row in the client schema, creates job topics, launches a K8s Job with topic env vars, sends the initialize message, and returns spawn results (agent_id, role, topics) into CollectedData for later call_agent lookup.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.045.spawn_actions4.refactor_into_functions.md
- **relations:** agent_definitions registry; job topics; role concept.
- **verify-later:** spawn_actions.go; client_{id}.agent_instances table.

<!-- SOURCE: U20_legacy_docs_a.md -->
### call_agent with role-based targeting
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.041: "role is the essential piece of information that links a specific task to a specific, previously spawned agent" with code walkthrough; used in all robot-hands workflows.
- **what:** CallAgentAction finds a previously spawned agent by searching CollectedData spawn results for a matching `target_role`, extracts its private requests_topic, and sends a `process` request there with await_response. Role acts as the within-orchestration nickname distinguishing multiple agents of the same type (adder vs multiplier calculators).
- **sources:** docs001_flow_general/README.041.role_flow.md; docs001_flow_general/README.018.flow7.roleflow.md; docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** spawn_agent; spawn step naming conventions; role-based agent pools proposal.
- **verify-later:** call_agent.go findAgentByRole/findTargetAgent.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Role-based agent pools / atomic work-claim queue (proposal)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** flow8 is pure proposal ("Migration Path: Phase 1…Phase 4"), never referenced as built in later docs; the design (work_items table, `claimed_by IS NULL` atomic UPDATE claim, role queues, failover pickup) is recognisably the ancestor of today's work-item pipeline.
- **what:** Instead of spawning agents tied to IDs, agents register roles/capabilities and claim WorkItems atomically from role-specific queues (`system.roles.{role}.pending`); unclaimed work survives agent death; pools scale elastically. "The role becomes the contract, not the agent ID."
- **sources:** docs001_flow_general/README.019.flow8.role_based_agent_pools.md
- **relations:** successor: work-item lifecycle / page-build-handler pipeline (development-guide, docs 001 current); scheduler-and-tasks concurrency groups.
- **verify-later:** work_items table and claim semantics in the current codebase — compare with this 2025 sketch.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Prompt resolution priority hierarchy
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.004 Part 3 "How the Flow Works Now" with three tested scenarios and log lines ("Using prompt from incoming message (Priority 1)").
- **what:** execute_llm_prompt resolves its prompt in priority order: (1) prompt passed in the incoming message/step config by the caller, (2) the agent's own prompt_template from agent_definitions, (3) workflow-step fallback. Lets parents override specialists while specialists keep good defaults.
- **sources:** docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** execute_llm_prompt action; agent_definitions default_config.
- **verify-later:** ai_actions.go ExecuteLLMPromptAction prompt lookup order.

<!-- SOURCE: U20_legacy_docs_a.md -->
### CollectedData normalisation and data_helpers safe-access layer
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full data_helpers.go source reproduced as "the new functionality" in docs002/0100 and used by all subsequent agents ("data_helpers.go functions ensure consistency", 0100c).
- **what:** One central layer (data_helpers.go) normalises every inbound message into a canonical CollectedData shape — `input_data` always at top level, system fields (`__execution_context__`, `__my_requests_topic__`, `__raw_message__`…) separated — and provides the only sanctioned accessors (GetInputData, GetStepData, GetMultipleStepData, GetFieldFromPath, TransformDataForAction, BuildRequestMessage/BuildResponseMessage/BuildInitializationRequest). Killed the `input_data.input_data` nesting chaos. Child input_data is always overwritten at top level — each agent's context is exactly what its parent sent (clean-slate encapsulation).
- **sources:** docs001_flow_general/README.070.a.centraliseddatanormalisation.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.024.flow14.input_data.md; docs001_flow_general/README.080.a.packaging_data.md
- **relations:** output_field/input_fields mapping contract; every action.
- **verify-later:** platform/orchestration/datahelpers package.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Agent groups — versioned, discoverable agent teams
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** README.060: "FindBestGroup… queries the database to find the best available version of that group, ordered by performance, usage, and version"; groups used in every website build; evolution/mutation service described but not evidenced live.
- **what:** agent_group_definitions rows are project recipes: a group_type, an agent_configs squad (role → agent_type), and an orchestration_workflow JSON, with integer versions as immutable snapshots (unique group_type+version). Requests name a capability (group_type) and the system picks the best version. An EvolutionService was designed to mutate groups into new versions with parent_id lineage and performance-based selection; the discovery/versioning part shipped, the evolutionary part appears aspirational.
- **sources:** docs001_flow_general/README.060.groupagents1.md; docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.062.groupagents3.databases.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** workflow selection priority; groups-as-project-recipes; spawn_group; site manifest pinning group_version.
- **verify-later:** agent_groups vs agent_group_definitions tables (both exist with different shapes — 062 shows the split); discovery/agent_discovery.go FindBestGroup; evolution.go.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Workflow selection priority (inline override > group > agent default)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.060/061 define and implement selectWorkflow with the three-tier priority; HITL tests routinely use inline workflow overrides.
- **what:** processor.selectWorkflow resolves which workflow to run: (1) a full inline workflow in the message config (ephemeral/testing), (2) a group workflow found via group_type, (3) the agent's default workflow from agent_definitions. Keeps production versioned while allowing ad-hoc experiments.
- **sources:** docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.060.groupagents1.md; docs002_hitl_parallel/README.0106.hitl_multistep_approval.md (inline workflows in practice)
- **relations:** agent groups; SagaCoordinator.
- **verify-later:** processor.go selectWorkflow.

<!-- SOURCE: U20_legacy_docs_a.md -->
### agent_definitions registry (DB-driven agent config and versioning)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Dozens of INSERT/UPDATE statements across all four doc sets; constraint migration to UNIQUE(type, version) with previous_version_id (096b); category CHECK constraint managed in 0140.
- **what:** Every agent type is a row: type, display_name, category (constraint-checked: data-driven/code-driven/adapter/…), default_config (containing the workflow, ai_service model+provider, processing_mode, timeouts), capabilities, image_repository/tag (all agents share the agent-chassis image), resources, topics, health_config, env_vars, version + previous_version_id, task_workflow/orchestrator_workflow, delegation_preferences. Creating an agent is a database insert, not a code change.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.096b.robothandswebsite.md; docs003_firecrawl/README.0140.removing_constraint.md; docs001_flow_general/README.098.oldherocontentdefinition.d
- **relations:** spawn_agent; agent categorisation taxonomy (998); the docs024-era agent creation guide is the living successor doc.
- **verify-later:** agent_definitions schema and constraints today; how many of these early agent types still exist/are active.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Spawn/step naming conventions
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.044: "The naming conventions are now important because we're using them to find spawned agents" — spawn_ prefix required, unique step names with 3-letter suffixes.
- **what:** Workflow authoring rules: spawn steps must start `spawn_<descriptor>` (suffix hints the role), action steps use perform_/execute_/process_ prefixes and reference agents by role, and step names must be unique within a workflow.
- **sources:** docs001_flow_general/README.044.spawn_actions3.spawn_rules.md
- **relations:** call_agent role lookup; workflow authoring guide (development-guide successor docs).
- **verify-later:** whether current workflow JSON still relies on prefix conventions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### evaluate_condition — template-based conditional branching
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 0127/0128 document the working mechanism ("The orchestrator uses this to pick the next step from the next_step map") including Go text/template functions (and/or/not/eq/gt…) and a live website-analyzer group UPDATE.
- **what:** Workflow steps gain branching: evaluate_condition renders a Go text/template expression against CollectedData and returns true/false; `next_step` becomes a map {"true": …, "false": …}. Enables data-driven workflow paths (e.g. extract_structured? crawl_pages? previous step success?).
- **sources:** docs003_firecrawl/README.0127.conditional_branching.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** conditional_branch/conditional_route actions; route_by_field/conditional_call_agent (later, richer routing).
- **verify-later:** evaluate_condition in registry; coordinator support for map-typed next_step.

<!-- SOURCE: U20_legacy_docs_a.md -->
### MVP site builder pipeline (strategist → architect → content-creator → deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Full group SQL + per-agent Kafka payloads documented and run (boxing-tickets.com); renamed/extended into landing-page-builder and 6-step pipeline within docs004; today's site building is the work-item pipeline.
- **what:** The first end-to-end production pipeline: chief-strategist (LLM → build_plan JSON of functional sections), site-component-architect (assemble_from_library → empty semantically-tagged HTML template + content_requirements "shopping list"), content-creator (fills slots), deployer-agent (commit_to_git). Group workflow spawns all four then calls them in sequence, threading outputs through output_field/input_fields.
- **sources:** docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** grew into: + brand-designer, + briefing-agent, + html-assembler, + specialist architects; successor: current page-build-handler/work-item pipeline (see 100_content_page_build_handler_flow.md).
- **verify-later:** agent_group_definitions mvp-site-builder / landing-page-builder rows.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Dynamic agent routing (route_by_field / conditional_call_agent)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Both Go actions written with registry additions listed; conditional_call_agent chosen because it "wraps CallAgentAction internally — no coordinator changes needed".
- **what:** Data-driven agent selection inside workflows: route_by_field maps a dot-path field value to a next step via a routes table with default; conditional_call_agent reads e.g. brief_data…site_type, maps value→agent type (landing→landing-page-architect …) and calls that agent in one step, returning routing metadata.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023a.description_for_conditional_routing_etc; docs004_website_capture_project/006semantic_themes/README.024.conditional_step_routing.md
- **relations:** evaluate_condition (simpler predecessor); spawn_group dynamic group_type (group-level equivalent).
- **verify-later:** registry entries conditional_call_agent, route_by_field.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Groups as project recipes + immutable versioning + agent pinning
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 025 "decision" section: "A group isn't 'agents that work together' — it's a project recipe"; versioning model (immutable version rows, sites pinned to group_version, per-agent image_tag pinning) is design; UNIQUE(group_type, version) constraint added in 0100b — partially realised.
- **what:** Each buildable *kind* of output (landing page, content site, 11ty blog, ecommerce) is a self-contained group: its own agent squad, workflow, questionnaire, and outputs. Divergence in output structure/build/deployment means a new group, not conditional routing. Group versions are immutable snapshots; a site records the group_version that built it and rebuilds with it unless upgraded; groups may pin specific agent versions where stability matters. Duplication across similar groups is accepted for clarity.
- **sources:** docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion; docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md (constraint)
- **relations:** agent groups; site manifest; tool-lifecycle versioning is the analogous live discipline.
- **verify-later:** group version rows per group_type; any site→group_version reference.

<!-- SOURCE: U20_legacy_docs_a.md -->
### spawn_group action with DB group lookup and dynamic group_type
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 028 discovers an existing SpawnGroupAction (config-provided agents) and revises the new version (spawn_group_from_db.go) to align — DB lookup of agent_group_definitions, dynamic group_type_field from collected_data, questionnaire fetch.
- **what:** Spawning an entire agent group as a unit: original action spawned each configured agent and returned subtree info; enhanced version resolves the group definition (agents + workflow + questionnaire) from the database, with the group_type optionally taken dynamically from prior step output — enabling the intake orchestrator's dispatch.
- **sources:** docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** intake orchestrator; agent groups.
- **verify-later:** spawn_group vs spawn_group_from_db in codebase.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Agent/group categorisation taxonomy (category, status, domain_tags)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** Migration SQL 031_add_categorisation with CHECK constraints (category: builder/analyzer/collector/transformer/evaluator/researcher/workflow/monitor; status: active/experimental/deprecated/demo/template) and GIN-indexed domain_tags; no doc confirms it was applied.
- **what:** Organisational metadata over agent_definitions and agent_group_definitions: what the agent *does* (domain-agnostic category), its lifecycle status, and flexible domain tags — an early attempt at the registry hygiene the concept register itself now pursues.
- **sources:** docs004_website_capture_project/998categorisation/031_add_categorisation_to_tables.sql
- **relations:** agent_definitions registry; documentation-system indexing.
- **verify-later:** do the category/status/domain_tags columns exist?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Aggregation patterns (aggregate_data, aggregator agent, input_from_collected_data)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** aggregate_data's failures traced (count:0 on verbose child responses, flow2); redesign to a spawned aggregator agent fed via input_from_collected_data path mapping (flow11); aggregate_webpage became the shipped variant for pages.
- **what:** Combining multi-step results: the local aggregate_data action broke against verbose child state objects; the redesign either normalises responses (data helpers) or delegates aggregation to a spawned aggregator agent whose call config maps CollectedData paths into its input. Response data keyed as response_{requestID} in CollectedData.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.010.flow.md
- **relations:** data_helpers NormalizeResponseData (the actual fix); aggregate_webpage.
- **verify-later:** aggregate_data current implementation.

<!-- SOURCE: U20_legacy_docs_a.md -->
### output_field / input_fields group-memory data mapping contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016 "Packet Flow" walkthrough resolving the exact semantics ("Take the entire result from the Strategist and store it under build_plan_data… path is simply build_plan_data.build_plan_json") producing the "Golden Copy" workflows.
- **what:** The inter-agent data plumbing convention: a call_agent step's `output_field` names the key under which the child's entire result lands in group memory; the next step's `input_fields` selects which keys are passed on; consumers address values by `<output_field>.<producer's own output key>` paths. Most orchestration bugs of the era were mis-mappings of this contract.
- **sources:** docs004_website_capture_project/website_analysis/README.016.agent_definitions_002.md; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md
- **relations:** CollectedData normalisation; template rendering paths; note the execute_llm_prompt flattening quirk (input_fields:["input_data"] flattens, so templates use {{.domain}} not {{.input_data.domain}} — 029 fix).
- **verify-later:** call_agent output_field handling in coordinator.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Deliberate discovery + human-approved agent evolution
- **category:** development-guide
- **status-signal:** abandoned
- **status-evidence:** README.005 principles ("Deliberate discovery — only at planning and review stages; Human approval — all agent changes require approval; Performance-based evolution") never reappear as a mechanism in later eras.
- **what:** Early governance rules for agent self-modification: the system only creates/modifies agents when starting a new task type, after poor performance review, and always with human approval — no heartbeats or automatic decisions. Paired with per-group performance recording and version incrementing.
- **sources:** docs001_flow_general/README.005.discovery.md
- **relations:** agent groups evolution service; HITL; tool-lifecycle health checks are the modern relative.
- **verify-later:** none.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Website build overall plan v0 (first multi-agent website roadmap)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** README.050 is a 6-phase/12-step plan (domain-analyst → content-creator → html-developer, then site-architect/visual-designer/site-publisher, data contracts, spawn_group team) written against the calculator-era platform; every element was rebuilt differently in docs002–004.
- **what:** The first articulation of "build a website with agents": minimal 3-agent workflow, explicit JSON data contracts between agents, progressive enhancement, mock-LLM-first testing, upload_to_s3 deployment. Registers as the origin point of the entire site-building programme.
- **sources:** docs001_flow_general/README.050.overall_plan1.website_design.md; docs001_flow_general/README.001.actions.md (action inventory of that moment: many mocks — deploy_to_hosting, http_request, cache_lookup all fake)
- **relations:** superseded by MVP site builder, then the work-item pipeline.
- **verify-later:** none.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Data-path resolution problem (agent vs local action nesting)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** docs006/001 ("Local Action: CollectedData[\"wrap_multipage\"] ... input_data.site_files.wrap_multipage ← Extra layer!"); docs009/002 ("collected_data.spawn_x.call_x.spawn_y...result"); resolved later by input_mapping + ActionInputSpec (docs017/008).
- **what:** The recurring class of runtime failures where workflow config referenced CollectedData paths that didn't match where actions actually stored results — agent calls store flat, local actions add a step-name layer, and each spawn/call deepens nesting. Drove multiple generations of mitigation: workflow builder path computation, explicit output_field conventions, data-flow verification matrices, and finally standardized input extraction.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#The-Problem; docs009_site_interrogation_and_solutions/002_claude_discussion#C; docs015_data_flow_verification/001_data_flow_verification.md
- **relations:** ActionInputSpec/ExtractActionInputs; workflow builder; data-flow verification practice.
- **verify-later:** datahelpers.ResolveInputMapping and FindByPath in platform code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Loop action via dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs010/022 integration guide with concrete file placements; docs010/023 walkthrough ("transforms the loop into individual workflow steps... Async-compatible... Recoverable... resume from exact step"); every later builder uses build_pages_loop.
- **what:** The `loop` action doesn't execute iterations itself — a coordinator-side expansion handler injects one workflow step per iteration×substep (generate_pages_loop_iter_0_research …), chained by NextStep, with loop_metadata in CollectedData, setLoopVariable placing the current item under the loop_var before each step, and a loop_complete step aggregating results into output_field. Design chosen (over in-process execution) because steps can await async agent responses and survive crashes/restarts as ordinary persisted workflow steps.
- **sources:** docs010_multitrack_flows_persona_architecture/021_loop_action_discussion.md; docs010_multitrack_flows_persona_architecture/022_loop_actions_guide.md; docs010_multitrack_flows_persona_architecture/023_loop_explanation.md
- **relations:** sequential page generation; work-item loops; orchestration state persistence.
- **verify-later:** loop_action.go, loop_expansion_handler.go, loop_complete_action.go in platform.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Data-flow verification matrix practice
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs015/001 is a complete per-step verification table ("Config | Value | Verified ✓") including output structures and registration checklist; docs017/044 repeats the practice for the site-work-orchestrator.
- **what:** A documentation/QA practice: before deploying a workflow, trace every step's config paths against the action implementations — where each output lands in collected_data, its structure, and each input's exact path — plus response-header compliance and action-registration checklists. The manual ancestor of automated contract validation.
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** input contracts; workflow validator; ActionInputSpec.
- **verify-later:** n/a (practice, not code).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Specialist agent design doctrine (agents own their domain)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Five versions of the checklist culminating in docs017/008; audited against real code in docs017/044 ("Audit Notes vs 008_checklist..."), including accepted divergences.
- **what:** The core agent-design rulebook: agents are self-contained and independently callable, with dedicated load_* actions gathering their own data; callers pass raw domain identifiers, never derived values ("if changing the child requires updating the caller, you've leaked responsibility"); reuse/patch existing actions before creating new ones; workflows stay declarative (templates/config = intent OK; loops/branching = Go); orchestrator vs agent boundary (what/order vs how); standalone + integrated dual modes; spawn before call; agents reply to the caller's topic; no container config in definitions. v2's interim "use input_fields not explicit paths" rule was replaced by the ActionInputSpec regime in v3.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md; docs017_legacy_agent_rules_images_design_keydocs/007_checklist_for_new_specialist_agent_v4.md#Orchestrator-Boundaries; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Audit-Notes
- **relations:** ActionInputSpec; webdesign-agent (first exemplar, docs017/003); current development-guide doc 001.
- **verify-later:** how closely current agents follow load_* pattern.

<!-- SOURCE: U21_legacy_docs_b.md -->
### ActionInputSpec / ExtractActionInputs standardized extraction
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs017/008 full spec ("No more boilerplate — 40+ lines of extraction code per action eliminated"; deprecation warnings for *_field patterns); real bug documented: "the site plan contamination bug — ExtractActionInputs found site_record.content_data via nested lookup... overwriting the hero section with the site plan."
- **what:** Every action declares an ActionInputSpec (Required/Optional/Defaults/Deprecated) and calls one extraction function that tries input_fields, falls back to deprecated *_field keys with warnings, checks nested parents (current_page/site_record/input_data/rerender_pages), validates and defaults. Includes the hazard doctrine — never name optional fields after common nested keys (content_data, domain, status...), prefix when in doubt — and the `?` suffix for optional input_mapping fields (skip silently if source path missing) supporting multi-mode agents. Literal config values must be read directly from config, not through path resolution.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Decision-Standardized-Input-Extraction; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Avoid-Field-Names; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Fixed-during-audit
- **relations:** data-path problem (root cause); input contracts; workflow validation.
- **verify-later:** datahelpers.ActionInputSpec/ExtractActionInputs/RegisterActionInputSpec in platform.

<!-- SOURCE: U22_recent_small_docs.md -->
### Field-path-resolution duplication tech debt
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "the codebase has at least 18 functions that resolve dot-separated field paths ... This is the single biggest code hygiene issue"; datahelpers canonical vs 9+ scattered duplicates enumerated.
- **what:** A recognised code-hygiene problem: ~18 near-duplicate dot-path resolution helpers (resolveFieldPath, ExtractNestedField, GetFieldFromPath, etc.) spread across datahelpers and the actions package, differing subtly in arg order, logging, and `.response` unwrapping. Canonical is `datahelpers.ExtractNestedField`; the standing rule is reuse datahelpers before adding new resolvers. Related recurring bug: Go-template paths need leading dots (`{{.x.y}}`) and input_mappers are compulsory.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#field-path, docs019_business/004_vet_practice_verifier.sql#go-template-fixes
- **relations:** datahelpers, rag_actions.go helper cleanup
- **verify-later:** datahelpers vs actions-package path resolvers; NullableString/TruncateString/NullableInt in datahelpers

<!-- SOURCE: U23_docs_root_vonc.md -->
### call_agent contract validation vs input_data.spec convention (dual placement; validator patch)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** Mechanism confirmed in code 2026-07-01 (ValidateInputContract checks only top level); PATCH_validate_input_contract.go "WRITTEN, deploy PENDING" (carried-forward state; still backlog item 3 on 2026-07-09).
- **what:** call_agent resolves input_mapping then validates the target's input_contract.required against TOP-LEVEL keys, while handler workflows read spec fields at `input_data.spec.*` (the work-item convention). The two read different places, so component-creator (required: section_type) can be satisfied neither by pure-top-level (empty-context generic generation — the 081 stray) nor pure-nested (contract violation — 082); the working manual shape (083) provides section_type BOTH top-level AND inside spec. build-dispatch-loop's generic mapping flattens no section_type, so the designed work-item path would hit the same violation (predicted, unconfirmed). Framework fix: the validator accepts a required field top-level OR at input_data.spec.X — not per-handler loop mappings, not enshrining the duplication.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~10:10 + #2026-07-01-~12:46 + #2026-07-01-~13:10; docs/PATCH_validate_input_contract.go.txt; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e; docs/083_regenerate_brief-explanation_vonc.sh (header)
- **relations:** manual agent trigger pattern; build-dispatch-loop genericity (002 §414); component regeneration
- **verify-later:** input_mapping.go ValidateInputContract (patched?); a needs_component_regeneration item dispatched through the loop

<!-- SOURCE: U23_docs_root_vonc.md -->
### Manual agent trigger via the generic entry point (spawn+call pattern)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Proven repeatedly: 083 (2026-07-01), 084 ("the dual-placement trigger pattern worked again", 2026-07-06), trigger-asset-renderer and rerender scripts.
- **what:** One-off manual agent runs post a spawn_agent+call_agent message to `system.agent.generic.requests` (kcat with correlation/request/client headers), with input_mapping delivering the payload. Hard-won sub-rules: dual placement of contract-required fields; a QUOTE-FREE description (name attribute values in prose) to survive the kcat/JSON escaping pipeline; JSON embedded literally (no jq dependency); watch via orchestration_states by correlation_id. The numbered trigger-script series (080–085) in scripts/initial_messages/210_vonc_trigger/ is the operational library, including make_085 which sed-copies a proven trigger for a new page (reuse-first).
- **sources:** docs/084_create_provocations-archive-list_vonc.sh (header); docs/make_085_rerender_provocations.sh; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6 + #§8; docs/summary.txt (kcat basics)
- **relations:** call_agent contract validation; work-item conventions
- **verify-later:** scripts/initial_messages/210_vonc_trigger/ contents

<!-- SOURCE: U23_docs_root_vonc.md -->
### Work-item conventions and manual spec shapes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01 (spec jsonb, item_key dedup via idx_swi_dedup, handler_agent, status flow, pipeline default 'build'); manual rerender/needs_page recipes proven repeatedly.
- **what:** site_work_items is the unit of work: `spec` jsonb (not spec_data at this layer), `item_key` (dedup), `handler_agent`, status detected→triaged→claimed→complete (dispatch picks up triaged/approved), `pipeline`. Manual page items require the FULL spec — page_id (real UUID inline), domain, filename, page_name; placeholder strings get claimed and fail ("invalid UUID length: 18"), and fixing them must filter on the PLACEHOLDER string, not the intended value (the wrong-WHERE no-op lesson). Duplicate insertions are cleaned by grouping on spec->>'page_name' and deleting the older of each pair. Fresh gen_random_uuid item_keys make re-fires safe.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-manual-rerender + #duplicate-work-item-cleanup; docs/RUNNING_NOTES_vonc(36).md#work-item-fix-2026-06-24 + #2026-07-01-~13:40; docs/RUNBOOK_vonc_session(1).md#correct-spec-shape
- **relations:** item_key canonicalization; build-dispatch-loop; complete_error family
- **verify-later:** \d site_work_items; idx_swi_dedup definition

<!-- SOURCE: U23_docs_root_vonc.md -->
### item_key canonicalization (workItemKey builder)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 016b Part 3: "CODE PREPARED; NOT APPLIED" — workItemKey(itemType, target) builder in work_items_common.go; apply gated behind Part-2 verification.
- **what:** item_key prefixes drifted from item_type across creators: the adoption creator keyed BOTH needs_content_page and needs_tool_recreation as `needs_page:<name>`, so a tool and a content page of the same name collide on the dedup index and one is silently dropped. Fix: a shared workItemKey builder; the tool item moves to its own prefix while the content item deliberately keeps `needs_page:` co-dedup with planner builds (Option B, decision recorded).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 3)
- **relations:** work-item conventions; dedup index
- **verify-later:** work_items_common.go workItemKey applied?; adoption creator key prefixes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Standing engineering rules (the session working constitution)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Restated as "Standing rules (brief)" in the v2 carried-forward state and as "Standing instruction from the user, in force" in the HANDOFF.
- **what:** The recurring project rules this corpus operates under: schema-before-SQL (\d first); reuse/alter before create (STEP ZERO); structural over quick fixes; workflows THIN with logic in Go actions; no sub-workflows in SQL — spawn sub-agents; every agent is an orchestrator; agents respond to the CALLER's responses topic; no logger.Debug (invisible in cluster logs); British English; flag variable/signature changes; never treat 0 rows as decisive; verify against deployed artifacts not pod logs; no summary docs unless asked; work in reasonable step sizes.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#standing-rules; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§1; docs/HANDOFF_vonc_write_site_spec_spec_data.md#standing-rules
- **relations:** debugging doctrine; development-guide (001) anchors
- **verify-later:** n/a (convention; verify against 001/002/003 docs)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Basic operations reference (kcat spawn, scale, monitoring)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** summary.txt is a concatenation of basic_usage docs actively describing the current operating procedure (spawn_group website-builder message shape, headers, monitoring queries).
- **what:** The operator's basic-usage layer: scale the deployment set up/down (agent-chassis, auth-service, content-creator-agent, core-manager, image-generator-adapter, reasoning-agent, web-search-adapter); post spawn_group/orchestrate messages via kcat from a test pod to the cross-namespace Kafka bootstrap with required headers (correlation_id, request_id, client_id, agent_instance_id, fuel_budget); monitor via orchestrator_state/orchestration_states by correlation_id. The fuel_budget header and the fixed header set are part of the platform's message contract.
- **sources:** docs/summary.txt; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6
- **relations:** manual agent trigger pattern; system-architecture (topics)
- **verify-later:** docs/basic_usage originals; current deployment list

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Guidelines audit (001/002/003 compliance)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Read the dev guide, architecture, and contracts. Existing code: no violations"; 2026-06-13(d) action "audited against 001/002/003 — code is COMPLIANT".
- **what:** Recurring audits confirming the engine and collector honour the house rules: engine is standalone package main; the no-JS HTML form satisfies JS Content Separation; parameterised SQL only; no logger.Debug; kebab-case/snake_case names; private same-file helpers are allowed; stats_key never logged.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_running_notes(27).md#2026-06-13-d
- **relations:** produced the wrapper-orchestrator finding and the envelope-contract flag
- **verify-later:** 001 dev guide; 002 architecture; 003 contracts

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### workflow field-path audit query (jsonb_path_query over agent_definitions)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** A single ad-hoc diagnostic query with no surrounding narrative or dated claim of use.
- **what:** A small but reusable diagnostic technique: a recursive `jsonb_path_query('$.**.<key>')` sweep over every `agent_definitions.default_config->'workflow'->'steps'` row to extract every field-path value referenced by a fixed set of workflow keys (`agent_type_field`, `default_from`, `content_field`, `iterate_over`, and any `*_from`/`*_field` wildcard key) across the whole workflow corpus at once — a way to audit real field-path usage in stored workflow JSON without opening each agent definition individually.
- **sources:** docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v2/001_validator_sql.sql
- **relations:** development-guide (001 anchor, "grep before using" / field-path resolution canonical-functions guidance)
- **verify-later:** none — a query technique, not a stored artifact

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow-as-configuration (JSON workflows in agent definitions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md gives the canonical `{"start_step": ..., "steps": {...}}` shape; HITL definition (Nov 2025) still uses exactly this workflow JSON structure with `next_step` chaining.
- **what:** Agent behaviour is a JSON workflow (start_step + named steps, each with an action, config, and next_step) stored in agent_definitions.default_config / task_workflow, overridable per agent_instances. Contrasted with Temporal/Airflow where workflows are compiled code — here business users can create workflows without deployment.
- **sources:** docs/architecture/002-agent-chassis-docs.md#how-workflows-work; docs/humanintheloop/hitl_agent_definition.sql; docs/architecture/012-investors.md#dynamic-workflow-creation
- **relations:** agent chassis; execute_llm_prompt action; local vs remote actions
- **verify-later:** agent_definitions.task_workflow / orchestrator_workflow columns; workflow validator code

<!-- SOURCE: U26_misc_dirs.md -->
### execute_llm_prompt generic action with DB prompt templates
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Planned in basic_usage/003 ("the reusable 'chef' that cooks the 'recipes'"); in live use by Nov 2025 — hitl_agent_definition.sql's workflow uses `"action": "execute_llm_prompt"` with a Go-template prompt_template.
- **what:** A single generic action that reads the agent's prompt_template and ai_service config (provider, model, api_key_env_var) from its definition, renders the template with Go text/template placeholders ({{.input_data.field}}) filled from collected workflow data, calls the configured LLM, and returns the text. Makes every LLM agent a pure data configuration.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement#step-1.2; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; dynamic prompt improvement loop
- **verify-later:** platform/orchestration/actions/ai_actions.go; prompt template rendering

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Pre-flight "does this already exist?" discipline (Step Zero)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) opens with it as "the most important step", with the asset-deploy-agent 3-hours-wasted example
- **what:** Before creating any agent/action, search agent_definitions, the action registry, Go code, gate functions, and workflows for existing equivalents; document findings; never create without demonstrating no existing coverage. Extends to documentation claims: verify at point of use and date what you verify (`[checked YYYY-MM-DD]`).
- **sources:** 001_development_guide(5).md#Pre-Flight, #API Verification Reference, #Reuse Before Creating
- **relations:** canonical field-path helpers; assumed-helper build failures (016 additions)
- **verify-later:** platform/orchestration/actions/registry.go; agent_definitions table

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Canonical field-path resolution helpers (datahelpers) vs 18+ duplicates
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): "18+ functions that resolve dot-paths… Do not add another one"; cleanup of 9+ duplicates listed under "Not yet built"
- **what:** `ExtractNestedField/String/Map`, `GetFieldFromPath(WithDefault)` in datahelpers are canonical (with `.response` auto-unwrap); six named duplicates in the actions package must not be copied. There is no `ExtractStringSlice` — compose `ExtractStringListHelper(ExtractNestedField(...))`.
- **sources:** 001_development_guide(5).md#Field Path Resolution; 016_additions_assumed_helper_and_cross_module.md
- **relations:** assumption checklist item on assumed helpers
- **verify-later:** platform/orchestration/datahelpers/*.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Actions are the unit of work — no wrapper+core split
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "The wrong pattern" section with WriteSiteSpec example
- **what:** All action logic lives inside the `XxxAction` function; composition happens via workflows, not Go-calling-Go; exporting a "core logic" function creates a duplicate API surface. Also: don't create subworkflows in SQL — spawn sub-agents.
- **sources:** 001_development_guide(5).md#Core Design Principles
- **relations:** every-agent-is-an-orchestrator
- **verify-later:** grep exported non-Action functions in actions package

<!-- SOURCE: U01_docs024_numbered_core.md -->
### spawn→call pattern and target_role lookup
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5): "This is how every existing workflow does it"
- **what:** Agents are spawned (`spawn_agent`) then called (`call_agent`). `target_role` → findAgentByRole scans all collected_data keys (preferred); `agent_type` → findAgentByType only scans keys starting `spawn_` (a trap). Dynamic dispatch = fixed role + `agent_type_field` resolved from collected_data at runtime; no topic-construction bypass.
- **sources:** 001_development_guide(5).md#How call_agent finds the spawned agent, #Dynamic dispatch; 002(4)#Resolved Decisions 16
- **relations:** dispatch loop; wrapper-orchestrator pattern
- **verify-later:** spawn_actions.go, call_agent.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Wrapper-orchestrator pattern ("every pod-running agent needs a parent that spawned it")
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) full section; med-* wrappers and site-adoption-orchestrator live per 002(4)
- **what:** Agents get dedicated K8s Job pods only via spawn_agent from a parent. Anything reached via the generic entry point that does substantive work needs a tiny wrapper (spawn → call → complete) so real work runs in its own pod with clean logs; in-chassis work blocks shared pod slots. Canonical minimal wrapper: med-export-orchestrator. Map input fields individually (never `input_data: input_data`), mark caller-optional fields `?`.
- **sources:** 001_development_guide(5).md#Every pod-running agent needs a parent; 002(4)#Active agents note
- **relations:** topics model; site-adoption-orchestrator
- **verify-later:** agent_definitions rows for med-* orchestrators; spawnAgentKubernetesJobFromDefinition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Standardized input extraction (input_mapping / input_fields / ActionInputSpec) and the `?` optional suffix
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) three-layer table; resolver behaviour documented from code
- **what:** Caller maps data via input_mapping (dot-paths into collected_data), actions declare fields, Go extracts via ExtractActionInputs. `?` suffix on the destination key makes a mapping optional (silently skipped); unsuffixed fields hard-fail the call. In the dispatch loop, only site_id/domain/work_item_id may be non-optional; all spec.* mappings must use `?`.
- **sources:** 001_development_guide(5).md#Standardized Input Extraction, #Optional fields in dispatch loop
- **relations:** field name collisions; handler input path contract
- **verify-later:** ResolveInputMapping in coordinator.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Field-name collision via the nested-source loop
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) corrected wording: affects required AND optional fields; real section-editor content_data clobber
- **what:** ExtractActionInputs's late nested-source loop checks `current_page`, `rerender_pages`, `site_record`, `input_data` for any unresolved field name — so generic names (`content_data`, `sections`, `site_id`, `domain`, `status`…) can silently bind to the wrong source. Rules: new code avoids colliding names (prefix them); existing code left alone unless it bites; complex/array fields must never go in ActionInputSpec (read the config path directly).
- **sources:** 001_development_guide(5).md#Field name collisions; 016 §0 item 15 and §9 literal-key trap
- **relations:** section-editor clobber; resolve_internal_links review catch
- **verify-later:** datahelpers/action_inputs.go nestedSources

<!-- SOURCE: U01_docs024_numbered_core.md -->
### RAG actions and knowledge_base shared store
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): migration 082 applied "Not yet populated"; rag actions Go "produced but not deployed" then 009 lists "registered, not workflow-tested"
- **what:** `rag_lookup` (embed→vector search→top-k, trigram fallback) and `rag_index` (chunk→embed→store, SHA256 dedup) over a shared `knowledge_base` table (vector(768)); deliberately actions not agents until a knowledge-indexer orchestration is needed. Tool docs also target a knowledge_base `tool_docs` row.
- **sources:** 001(5)#RAG actions, #Agent vs infrastructure; 037#Where the docs live
- **relations:** agent-vs-infrastructure test; per-tool docs convention
- **verify-later:** rag_actions.go registered?; knowledge_base row counts

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Agent vs infrastructure boundary test
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) table (LLM logger no, Ollama provider no, rag actions no, knowledge-indexer future yes)
- **what:** Something becomes an agent only if it owns a domain, needs its own workflow, and benefits from independent spawn/debug. Otherwise it is an action or cross-cutting infrastructure.
- **sources:** 001_development_guide(5).md#Agent vs infrastructure
- **relations:** promotion pattern (002d)
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Specialist vs handler: the persistence boundary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) post-mortem; hit twice (page-content-writer HTML trapped in site_work_items.result)
- **what:** A specialist returns data to its caller; a dispatch handler must persist its own outputs (page_components, site_components, assets) and update status. Specialists used as handlers need a wrapper (page-build-handler wraps page-content-writer with plan/validate/save/deploy). Test: callable from CLI with site_id+domain and everything lands in the right tables.
- **sources:** 001(5)#Lessons Learned; 002(4)#Page Build Handler Pipeline
- **relations:** dispatch loop; handler contract
- **verify-later:** page-build-handler definition; handler agents' save steps

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Extended thinking config and the no-temperature-to-Anthropic rule
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) dated 2026-05-27 temperature note
- **what:** `budget_tokens` in ai_service enables extended thinking (thinking blocks skipped in parsing, +30-90s). Since 2026-05-27 the Anthropic client sends no temperature at all (Opus 4.7+ 400s on non-default; thinking incompatible); Ollama still honours it. Steer Anthropic via budget_tokens and prompts.
- **sources:** 001(5)#Extended Thinking Configuration
- **relations:** LLM config shadowing (temperature dead paths)
- **verify-later:** anthropic.go client options

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Loop mechanisms: dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) Appendix C full reference incl. production dispatch loop example
- **what:** Loops inject N×M steps into the workflow plan at runtime (`{loop}_iter_{N}_{substep}`), with setLoopVariable propagating iteration outputs to base names, per-iteration output suffixing, continue_on_error skipping, and LoopCompleteAction aggregation. Known hazards: the fast-response race fixed by the ErrLoopExpansionHandled sentinel; shared `loop_metadata` key; never nest loops — spawn a sub-agent instead.
- **sources:** 001(5)#Appendix C — Loop Mechanisms
- **relations:** dispatch loop pattern; O(K²) state-bloat failure (016)
- **verify-later:** loop_actions.go, loop_expansion_handler.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Authoring rules pack (schema-check, parameterised SQL, api_key_env_var, nil-guarded templates, code-fence stripping, error_step-in-config, text wrapped for write_site_spec)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "Summary of rules" 1–20, each backed by a dated bug
- **what:** The distilled 20-rule authoring discipline from the bug tally: `\d` live schema before SQL (dumps go stale; domain→pipeline, version_note→change_description); $1+params never {{.field}} in SQL; every LLM step needs api_key_env_var; {{if}} before {{range}} (query_database empty = null not []); run agent-def SQL before triggering (chassis silently runs empty workflow); strip markdown code fences before JSON parse; error_step only works inside step.Config; write_site_spec rejects scalars (wrap {"text": …}); to_jsonb('…'::text); verify fire-and-forget INSERTs actually land.
- **sources:** 001(5)#Appendix A + #Summary of rules
- **relations:** debugging assumption checklist; best-effort-needs-monitoring
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Orchestrator wrapper pattern for dedicated pod spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Spawning confirmed working (2026-04-23 trigger test) … processing_mode: 'orchestrator' at top level + agent_category = coordinator IS the combination that produces a dedicated spawned pod"
- **what:** To run work in its own spawned pod rather than one of the three shared chassis pods: an orchestrator wrapper (category=orchestrator, agent_category=coordinator, processing_mode=orchestrator at top level of default_config) with steps spawn_agent → call_agent(target_role, not agent_type) → complete_workflow, calling a worker (specialist, processing_mode=task). Input mapping maps fields individually with `?` suffix for optionals — never a whole `input_data` blob. File writes from non-spawned actions land on a random chassis pod and die with it.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4f "Operational gotcha", #2.4h, #14 "Chassis action design patterns"
- **relations:** agent_definitions three-column semantics; training-data-exporter as reference implementation
- **verify-later:** training-data-export-orchestrator agent definition as the canonical example

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### agent_definitions three-column semantics (category / agent_category / status)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Caught three agent_definitions column semantics confusions" (2026-04-23); reference row improvement-loop = category=orchestrator, agent_category=coordinator, status=experimental
- **what:** `category` is free-text functional role; `agent_category` is CHECK-constrained to strategist/executor/analyst/integrator/coordinator/specialist (NOT orchestrator); `status` is lifecycle. Naïve writes put lifecycle values in the wrong slot. Also: ON CONFLICT must target (type, version) with `WHERE deleted_at IS NULL`.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4h, #14
- **relations:** orchestrator wrapper pattern
- **verify-later:** CHECK constraint on agent_definitions.agent_category

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Chassis action input conventions (ExtractActionInputs / input_data, dual registration)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Rewritten to use the canonical pattern … matches every other action in the codebase" (2026-04-23); registry gap bit on 2026-04-20: composition actions "had NO entry in GlobalActionRegistry … rejected as 'requires a topic'"
- **what:** New actions use `datahelpers.RegisterActionInputSpec` in init() + `ExtractActionInputs` (5-strategy cascade) rather than raw ExtractNestedFieldString; parameters flow via `CollectedData["input_data"]` because `{{.input_data.X}}` templating does NOT render for deterministic-action step config; every new action needs BOTH the InputSpec registration AND a GlobalActionRegistry entry with IsLocal:true; results land in collected_data under output_field, never final_result. Config literal numbers must be read with `datahelpers.GetIntField(params.StepConfig.Config, …)` — `inputs.GetInt` reads collectedData, not config.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#14; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#2; HANDOFF_2026-04-17_triage_and_component_linking.md#1
- **relations:** CollectedData architecture; field-name collision risk
- **verify-later:** datahelpers package; registry.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pgbouncer per-batch transaction discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** v3.0 "driver: bad connection" failure → "v3.1 split into per-batch transactions, batch size 500 → 100 … Per-batch commits worked" (2026-04-23)
- **what:** Long-held transactions through pgbouncer (transaction pool mode) are fragile — bulk work must commit per small batch (<1s each), never wrap a streaming job in one transaction. Companion rule: check RowsAffected on single-row UPDATEs and fail loudly instead of Warn+continue (v3.1's final UPDATE silently didn't land).
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4i, #14
- **relations:** training-data export v3 evolution
- **verify-later:** pgbouncer pool mode config; v3.2 UPDATE handling

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Agent re-registration vs re-seed risk (DB row authoritative)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** RUNBOOK 07-06 night: "deploys bump `agent_definitions.updated_at` without overwriting `default_config` (the old prompt survived today's deploy)… the user confirms component-creator is a dynamically spawned/registered agent, not a YAML-seeded one. So the DB row is authoritative."
- **what:** A durability model for DB-edited agent prompts: chassis deploys re-register agent definitions (bumping `updated_at`) but do not overwrite `default_config`, so SQL edits to prompts survive deploys for dynamically registered agents. The residual risk is an in-code prompt template driving an upsert; the check is one grep for a literal fragment of the OLD prompt in Go sources — a hit means mirroring the edit in code, no hit means nothing can revert the row. (Earlier drafts had a heavier "seed check" over configs/deployments YAML; superseded by the user's confirmation.)
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-C; RUNBOOK_scheme_to_components(49).md (the earlier five-minute variant, family-delta); running_notes_scheme_to_components(55).md#Uf #Uj
- **relations:** component-creator prompt re-aim; 019 idempotent prompt migration pattern.
- **verify-later:** run the Step C grep; agent registration code path (upsert semantics on default_config).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item crafting conventions (real shapes, truthful provenance, dedup)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** w4b_04 comments: "crafted from the real rows … with truthful deviations noted: source 'manual' and created_by 'w4b_chrome_refresh' … lying in provenance columns costs later debugging"; every W6/W7/W8/W9 insert repeats the pattern.
- **what:** The discipline for hand-inserting `site_work_items`: copy the metadata of real rows produced by the owning code path (pipeline/severity/priority/handler_agent/status), deviate only truthfully in provenance columns (source=manual, created_by=<script name>), carry only spec fields the consuming workflow actually reads, and dedup check-first with a NOT EXISTS that mirrors `idx_swi_dedup` exactly (non-terminal statuses only — including 'unresolved', a status the index taught the thread it had missed). Item_key families are stable conventions: `page_rerender:<page>`, `chrome_refresh_rerender:<site_id>`, `needs_imagery:section:<scope_ref>:<key>`, `component_regen_rerender:<uuid>`, `section_data_*`. The check-first pattern is borrowed from CreateNeedsNewComponentItem.
- **sources:** w4b_04_trigger_item.sql; w7b_01_imagery.sql; w8_01_post_deploy_rebuild.sql; running_notes_scheme_to_components(55).md#Tb #Tc
- **relations:** rerender-pages v6; work-item claim/retry; scheduler-and-tasks.
- **verify-later:** idx_swi_dedup definition; site_work_items status vocabulary.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### House rules and standing preferences (the working contract)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated verbatim at the top of the journal ("Standing preferences (STRICT)") and in both HANDOFFs so fresh chats inherit them.
- **what:** The user's cross-thread working contract, treated as binding by every agent session: Go not Python; British English; plain language, no hype/flattery, banned words "perfect/critical/excellent", no congratulation; confirm live schema/data before asserting or writing SQL; schema-first (`\d` before SELECT/UPDATE); structural framework fixes over one-off patches; low risk appetite, reasonable step sizes, ≤1 question per reply; no summary documents unless asked; don't call fixes final; no `*-light`/`*-dark` component variants; keep runbook + journal current; honest caveats including correcting one's own reads ("corrections owned").
- **sources:** running_notes_scheme_to_components(55).md#Standing-preferences; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; HANDOFF_scheme_to_components_for_claude_code(1).md#Constraints
- **relations:** running-notes journal discipline; orchestrator conventions.
- **verify-later:** n/a (convention, not code) — check for a canonical repo home for these rules.

<!-- SOURCE: U04_idea_uk.md -->
### Launch idioms: orchestrate vs work-item insert (and what each trigger does NOT do)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Confirmed from the production trigger scripts" (2026-06-20 iii); the 081c finding (no hand-rolled wrappers) cited.
- **what:** Two production ways work starts: (1) static agents are orchestrated by producing one Kafka message to system.agent.generic.requests (action=orchestrate, config.agent_type, full header set) via a one-off kcat pod; (2) dynamic handlers (page-build-handler etc.) cannot be orchestrated directly — INSERT a `site_work_items` row (status='triaged') and the running build-dispatch-loop claims and spawns them. Key caveat learned on idea.uk: the content triggers (rerender-pages / page-rerender / page-rebuild) never re-resolve composition — palette changes must go through needs_composition/needs_design. Deploy topology is likewise two-path: Go changes ship in the chassis image (roll agents to the tag), site HTML ships via the sites monorepo → Actions → B2.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Launch idioms); idea.uk/running_notes_checkpoint_tt.md (deploy topology + trigger mechanism); idea.uk/HANDOFF(13).md
- **relations:** composition re-resolve; scheduler/dispatch loop; debugging guide kcat sections.
- **verify-later:** 082_submit_domain_unified.sh; build-dispatch-loop definition.

<!-- SOURCE: U04_idea_uk.md -->
### Array-item field contract for the page-content-writer (item_fields fix)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "Prompt migration is already applied… But until a chassis image carrying the Go change is live, {{if .item_fields}} is always false… the applied prompt is inert on its own" (checkpoint tt, 2026-06-21).
- **what:** Root-cause class behind empty rendered sections (the 7 blank differentiator cards): the writer's prompt listed an array field with its type but never its per-element shape, so the model guessed item keys (title/body) that render empty against templates reading name/description. Fix has three coupled parts: plan_sections populates `ItemFields` on each llm_field_spec (Go); the prompt migration renders the exact per-item field list in both What-To-Write and the JSON skeleton (019_pcw_prompt_item_fields.sql — idempotent by sentinel, no broken intermediate state either deploy order); a render-time reconciler in v3_site_actions.go. Deploy order matters: chassis image first, then trigger.
- **sources:** idea.uk/019_pcw_prompt_item_fields.sql; idea.uk/running_notes_checkpoint_tt.md; idea.uk/README_assemble_bundle_idea_missing_sections.md (the bundled problem statement)
- **relations:** section-data reconciler; coordinator contract (sibling contract-mismatch class); diagnosis-loop bundles (the assemble-bundle invocation).
- **verify-later:** plan_sections_action.go ItemFields; whether the chassis tag carrying it shipped.

<!-- SOURCE: U05_content_quality_linking.md -->
### Sub-agent modelling conventions (agent_definitions row shape)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** running_notes_17(21) "Agent-modeling facts (from research-agent row + 003)"; internal_link_resolver_agent.sql embodies them.
- **what:** How a called sub-agent is modelled: workflow inside default_config.workflow (agent_definitions has NO processing_mode column — it lives inside default_config with timeout_seconds); agent_category specialist; input/output contracts required; templated topics (system.agent.{type}.process etc.); responds on the parent's responses topic; NOT-EXISTS-guarded idempotent seed SQL; image_repository/image_tag pinned to the batch image. Modelled on research-agent as the proven sibling.
- **sources:** internal_link_resolver_agent.sql; running_notes_17(21).md#agent-modeling-facts
- **relations:** internal-link-resolver; every-agent-is-an-orchestrator doctrine; 003 contracts.
- **verify-later:** research-agent/internal-link-resolver rows side by side.

<!-- SOURCE: U06_finetuning.md -->
### input_mapping semantics: call_agent-only; config dot-paths for local steps; key_path for loop items
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) §2 "[verified-source] Local action steps do not resolve input_mapping… input_mapping is dead config"; 109b header "CORRECTS a load-bearing assumption: input_mapping is NOT live for (local-action) loop substeps."
- **what:** The coordinator honours `input_mapping` only for call_agent (building child input_data) and loop fan-out; on plain local action steps it is dead config. Local actions pull values via config keys whose values are dot-paths resolved from collected_data (`ExtractActionInputs` Strategy 0 / `resolveTemplateToken`); loop substeps read the iteration item via a config dot-path like `key_path:"ckpt_key"` (setLoopVariable puts the item in CollectedData) — using input_mapping there silently falls through to fallbacks (the dataset-key-presigned-40× bug). A proposed coordinator change to resolve input_mapping on local steps was deliberately withdrawn (D1: fix the caller, don't teach the framework a new behaviour for one agent's misuse). Optional mapping fields take a `?` suffix; missing required sources hard-fail (migration 103's fix).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D1,D2),#8; working/phase5/109b_fix_presign_one_loop_item_keypath.sql; working/phase5/103_call_data_preparer_optional_inputs.sql
- **relations:** output_fields contract; launcher workflow; loop_complete convention
- **verify-later:** coordinator ResolveInputMapping; input_mapping.go `?` semantics L101-128

<!-- SOURCE: U06_finetuning.md -->
### Child-result shaping: output_fields (plural) contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-03 ~17:3x: "extractWorkflowResult reads completeStep.Config['output_fields'] — PLURAL only… singular output_field… is never read → falls to the fallback branch that dumps every non-internal collected key"; migration 104 confirmed live.
- **what:** An agent's final result shape is governed by its complete step's `output_fields` array; the singular `output_field` spelling is silently ignored, producing a step-name-keyed fallback dump that breaks consumers' documented paths (`provisioning_result.provisioning_id` buried under `dispatch_provision.response.…`). The resolver auto-unwraps one `.response` per path part but never crosses arbitrary step-name keys. Fix taken at the def level (gpu-provisioner switched to plural + launcher mapping repointed, migration 104) after the user vetoed a chassis change; recorded as debugging-guide gotcha #23. Corollary rule: verify each call_* step's mapped source paths against the producer's REAL collected_data shape before firing anything that books a GPU.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-173x—180x; working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql (header)
- **relations:** input_mapping semantics; data-path verification runbook step 2b
- **verify-later:** extractWorkflowResult in coordinator; whether other defs still use singular output_field (thunder-reaper was named)

<!-- SOURCE: U06_finetuning.md -->
### Reply-topic derivation rules (own topic vs parent topic; two-level await)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) D4 + 2026-06-02 16:12 "D4 CONFIRMED live"; STATUS 06_04 reply-topic orphan fix "VERIFIED 2026-06-04 18:21".
- **what:** Two awaits, two topics: a child's intermediate adapter calls are awaited by the child's OWN coordinator on `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`); only the child→parent final notification uses `__parent_responses_topic__`. Dispatch actions that put the parent topic in an adapter envelope orphan the await (adapter replies where no one listens → infinite hang) — this bit twice (launcher dispatches pre-D4; `dispatch_thunder_ssh_get_status` cloned from ssh_exec). The inherited handoff asserted the opposite convention — corrected against source ("verify against code, not the handoff"). A shared `resolveAwaitResponsesTopic` helper is flagged as the future consolidation; a latent fallback caveat remains (the `system.agent.<type>.responses` fallback doesn't match the launcher's actual `system.responses.training-launcher` topic).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D4),#6,#10; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** send-before-register race; monitor build; adapter header tiers
- **verify-later:** determineResponsesTopic priority order in coordinator; whether the shared helper was built

<!-- SOURCE: U06_finetuning.md -->
### Orchestrator wrapper spawning pattern (dedicated pods for workers)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i "Spawning architecture — fully confirmed working"; worker pod agent-training-data-exporter-… observed.
- **what:** To run work in a dedicated spawned Job pod rather than the shared chassis pool: a wrapper agent with `processing_mode:"orchestrator"` at the TOP level of default_config, category='orchestrator' (free text), agent_category='coordinator' (CHECK-constrained — 'orchestrator' is not allowed), running `spawn_agent → call_agent(target_role=…) → complete_workflow`; the worker uses processing_mode:"task"/specialist. Includes the three-confused-columns trap on agent_definitions (category vs agent_category vs status; reference row improvement-loop) and ON CONFLICT (type,version). The monitor later reuses the same pattern in a loop (per-instance spawn+call, sequential).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i,#chassis-action-design-patterns; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data-exporter v3; monitor orchestrator; ExtractActionInputs canonical pattern
- **verify-later:** spawn decision logic in chassis (processing_mode placement)

<!-- SOURCE: U06_finetuning.md -->
### pgbouncer long-transaction fragility → per-batch commits
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) §2.4i v3.0 failure "'bulk insert 500 rows: driver: bad connection'" and the v3.1 restructure; §14 pattern entry.
- **what:** Long-held transactions through pgbouncer (transaction pool mode) trip connection-level failures; bulk work defaults to per-batch commits (batch 100, each under a second) with single-statement non-tx bookends. Companion rule from the same incident: always check RowsAffected() on single-row UPDATEs and error rather than warn — an action can return perfect counts while its final UPDATE silently didn't land.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#2.4i,#14; working/flywheel_docs/HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data export v3.1/v3.2
- **verify-later:** training_data_export.go batch logic

<!-- SOURCE: U06_finetuning.md -->
### Reuse-first build discipline (grep before adding; delegate, don't parallel)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NOTES(45) 2026-06-04 reuse audit ("ssh_get_status ALREADY EXISTS… no adapter change"); D3 "Reuse over parallel code in the adapter presigner"; guideline audits run against 001/002/003 for every artifact batch.
- **what:** A repeatedly-exercised discipline in this workstream: before building, audit what exists (ssh_get_status reused as the monitor probe; ListObjects reused for resume; DatasetURL/ArtefactURL refactored to delegate to ObjectURL rather than a third signer; datahelpers GetIntField over a custom helper; preRegisterAwaitedRequest reused for the race fix). Each new artifact batch is audited against the dev guide/architecture/contracts docs before deploy, with violations fixed or explicitly accepted (the one accepted tradeoff: launcher reading through the provisioner's step name).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#3(D3),#update-2026-06-04(reuse-audit),#guideline-audit,#update-2026-06-05(guideline-audit)
- **relations:** adapter design guide; input_mapping/output_fields contracts
- **verify-later:** n/a (practice); the accepted step-name coupling in call_launcher mapping

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F6 — dedup status-list mismatch and itemsCreated overcount
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "F6 flagged: the store's NOT EXISTS guard… omit 'unresolved' → Go guard STRICTER than index" (NOTES §9aa); unfixed Part E flag.
- **what:** Two small aligned defects: (1) the store's NOT EXISTS dedup status list and the `idx_swi_dedup` partial-unique predicate disagree on `unresolved` (index-terminal but guard-blocking — an unresolved squatter blocks createRerenderWorkItem where the index would not); (2) create_rerender_items increments `itemsCreated` without gating on RowsAffected, so ON CONFLICT DO NOTHING conflicts overcount the log. One-line alignments, parked.
- **sources:** NOTES(43).md §9t, §9aa; RUNBOOK(49).md Part E F6; HANDOFF(7).md §Flags
- **relations:** work-item dedup semantics; F3 (same action); hygiene: 40 stale unresolved items.
- **verify-later:** idx_swi_dedup definition vs createRerenderWorkItem NOT EXISTS list; create_rerender_items counter.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Work-item dedup semantics (item_key + idx_swi_dedup partial unique index)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (complete, verified, rejected, wont_fix, failed, unresolved)" — captured from pg_indexes (NOTES §9aa).
- **what:** site_work_items dedup is two-layered: producers guard with NOT EXISTS over open statuses and the DB enforces a partial unique index on (site_id, item_key) over non-terminal statuses. Terminal-status items free the key (why completed triggers can be re-inserted for retriggers); mirroring the producer's exact insert (columns, item_key scheme, dedup clause) is the established way to hand-create conforming items. See F6 for the known guard/index mismatch.
- **sources:** NOTES(43).md §9q, §9aa; RUNBOOK(49).md Part C Step 9b; w4b_03_read_rerender_config.sql
- **relations:** F6; work-item spec-cloning discipline; F3.
- **verify-later:** idx_swi_dedup definition; createRerenderWorkItem insert shape.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Deploy-ordering hard gate for coupled Go action + workflow-config changes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "LESSON (runbook gate tightened): 'deploy the Go action first' is insufficient — the migration is live instantly while the image rolls out. Hard gate: confirm… registered + live on ALL pods, THEN apply the migration" (NOTES §9i–§9j).
- **what:** Workflow jsonb changes take effect immediately; Go actions only exist once the image is rolled out and the registry entry (IsLocal:true) is in the running build. Wiring a workflow step to a not-yet-live action makes the validator reject EVERY run of that agent (WORKFLOW_INVALID broke all component generation during F2 3a). The codified gate: deploy + confirm the action responds on all pods before applying the (idempotent) migration; `revert_agent('<type>')` is the immediate mitigation.
- **sources:** NOTES(43).md §9i, §9j; F1prompt_component_creator_preserve_field_names(1).sql PREREQUISITE header
- **relations:** F1-prompt (where it bit); prompt-migration convention; snapshot/revert_agent.
- **verify-later:** workflow validator is_local check; revert_agent/snapshot_agent functions.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Prompt/workflow-jsonb migration convention (snapshot-first, anchored, idempotent, drift-checked; the 072 trap)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** The F1prompt migrations implement it end-to-end ("Convention: snapshot-first, idempotent, drift-checked, live-row only" — F1prompt(1).sql); prompt located top-level after `prompt_is_top_level=f` proof (NOTES §9d).
- **what:** Agent behaviour lives in default_config jsonb; edits are live instantly and follow a strict convention: snapshot_agent first; anchor the edit on a unique existing string and abort if the anchor count ≠ 1 (drift check); idempotency marker so re-runs no-op; filter to the live row (is_active, not snapshot, not deleted). The "072 nested-prompt trap": prompt_template may live at the top level of default_config OR nested in a step config — verify the path first or the migration is a silent no-op. Anti-drift prompt anchors have precedent (tool-doc-header rule on tool-improver): prompt rule = the anchor, store guard = the gate.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c, §9d, §9k (dead-block cleanup — an idempotency-check subtlety)
- **relations:** F1-prompt; F3c config edit; D2b-2 prompt edit; deploy-ordering gate.
- **verify-later:** snapshot_agent/revert_agent; component-creator prompt state (no dead {{if .existing_field_names}} block).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Never-guess disciplines: clone real work-item specs, look up real URLs (phantom-CTA lesson)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "NEVER guess a needs_page spec" (HANDOFF(6→7)); "MY CORRECTION: I had baked a guessed path… replaced with a pages.url lookup query (never guess paths; pages.url is the source)" (NOTES §9ae); spec-shape reads staged before every insert (w4b_03).
- **what:** A cluster of verify-at-point-of-use rules that recur across both threads: hand-created work items are mirrored column-for-column from a real captured item (SELECT-based inserts, real spec shapes, conforming item_keys), never composed from memory; URLs/paths come from pages.url, never invented (the phantom-CTA bug was an invented /contact.html; the recovery re-verified the same value before trusting it); schema before SQL (column names checked against \d, e.g. occurred_at not created_at); trigger flows through the real producer path rather than manual inserts where one exists (needs_design via build-site-planner / the proven 076 trigger).
- **sources:** NOTES(43).md §9l, §9ae, §9w–§9y, §9bd, §9bl–§9bm; w4b_03_read_rerender_config.sql; HANDOFF(7).md §Immediate next action
- **relations:** work-item dedup; section readiness (spec presence ≠ validity); F2 discriminators; link-management (phantom links).
- **verify-later:** n/a (convention; instances cited).

<!-- SOURCE: U08_travelling_docs.md -->
### Snapshot-before-update standing rule + the platform's snapshot_agent()
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Snapshot before updating (STANDING RULE 2026-07-06)"; write-location confirmed 2026-07-07 (stored OUTSIDE agent_definitions); every subsequent migration carries the call.
- **what:** `SELECT snapshot_agent('<type>', '<migration>: pre-update')` is prepended to every agent-updating migration. The function already existed (a reuse win — the drafted side-table migration was superseded un-applied; lesson: fetch-first applies to FUNCTIONS too, `\df` before drafting backup machinery). Snapshots live in a separate store (later identified as agent_definitions_backup in the FYI), so the defensive `is_snapshot` selector predicate is not load-bearing. Companion `revert_agent('<type>')` exists per 016b.
- **sources:** RUNBOOK_travelling_docs(38).md#§0-REF,#task-1; RUNNING_NOTES_travelling_docs(39).md#rev18,#rev19,#rev22; FYI_from_fixloop_2026-07-10…md
- **relations:** migrations system; correct-while-touching.
- **verify-later:** `snapshot_agent`/`revert_agent` function definitions; agent_definitions_backup table.

<!-- SOURCE: U08_travelling_docs.md -->
### Correct-while-touching norm (bounded repair of adjacent inert bugs)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Defined in the RUNBOOK mini-glossary, "Norm adopted in this chat, 2026-07-06"; exercised in migrations 125–146 (e.g. all ten of recreation's step-level error_steps corrected while adding its note tail; 146's cooldown fix).
- **what:** When a migration already modifies a workflow, it also fixes known-inert bugs in that SAME workflow (e.g. step-level `error_step` moved into config with original targets, dead keys deleted), declared explicitly in the file — bounded repair, no separate campaign, never copying the broken shape into new steps.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev23,#rev26,#tier-4-continuous
- **relations:** error_step-in-config; snapshot rule.
- **verify-later:** the declared correct-while-touching sections inside migrations 125–146.

<!-- SOURCE: U08_travelling_docs.md -->
### error_step mechanics — config-level placement, existing target, derive-from-next_step, loop corollary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b(7) §9 entry (live-validated ×5 in run-3); mechanism documented in 001 §16; dormant step-level instances corrected across tool-generator/fix agents by migrations.
- **what:** The coordinator reads only `step.Config["error_step"]` — step-LEVEL error_step is parsed but silently ignored. Once placed correctly, the target must name an EXISTING step or the coordinator fails the whole workflow (a typo converts a recoverable failure into a fatal one). Pattern: derive `error_step` from the step's own `next_step` read from the same row (convergence by construction, nothing guessed); `jsonb_set` does not create parents — COALESCE-merge config. Loop corollary: inside loop substeps, `error_step`/`then_step`/`fallback_step` are iteration-prefixed at expansion and must name substeps of the same loop; `continue_on_error: true` is the iteration-scoped alternative.
- **sources:** 016b_debugging_guide_7_3_(7).md#error_step-entry; RUNBOOK_travelling_docs(38).md#§8; RUNNING_NOTES_travelling_docs(39).md#rev9,#rev12,#rev13
- **relations:** docs-never-fail containment; correct-while-touching; loop_expansion_handler.go.
- **verify-later:** `routeToErrorStepOrFail`/`continueExecution` in coordinator; loop_expansion_handler.go prefixing.

<!-- SOURCE: U08_travelling_docs.md -->
### The seam rule — every prompt consuming a spec field must render it
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration 138 applied ("Mandatory Behaviour Requirements" section rendered from spec.interactive_features, marked as overriding the source); PLAN(6) rollout-outcomes bullet.
- **what:** A requirement can survive the analysis step and still be ignored at generation if the generation prompt never renders it: `analyze_tool` rendered `spec.interactive_features`, `recreate_tool` didn't, and Opus trusted the visible source HTML over requirements buried in a 20KB analysis JSON — faithfully recreating the bugs it was asked to fix. Rule: when adding a spec field, grep EVERY prompt_template that should render it; render requirements explicitly, marked as overriding the source.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45-run2; HANDOFF_2026-07-10…md#§4
- **relations:** economy-simulator case; "passed checks ≠ working"; doc_notes `seam` category.
- **verify-later:** recreate_tool prompt's Mandatory Behaviour Requirements section (migration 138).

<!-- SOURCE: U08_travelling_docs.md -->
### Manual kcat trigger scripts (084–087) — the canonical manual-orchestrate envelope
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** All four scripts exist, exercised repeatedly; 086/087 add DRY-RUN-by-default + single-line enforcement.
- **what:** A family of shell triggers producing `action=orchestrate` messages to `system.agent.generic.requests` with the house envelope (correlation/orchestration/request/message ids; `config.agent_type=<target>`). Encoded operational knowledge: target the spawn-wrapper orchestrator, not the agent directly (the SPAWNED pod gets GITHUB_READ_TOKEN via the spawn gate; an in-place run on a shared pod fails pre-fetch); REF explicit, never HEAD (user decision 2026-07-02); banners print effective subject/function as the go/no-go tell; DRY-RUN default with SEND=1; declared real side effects on the live trial site.
- **sources:** 084_TRIGGER_diagnose_v1(2).sh, 085/086/087 headers; RUNNING_NOTES_travelling_docs(39).md#rev10,#rev27; 086_input_data_recreate_economy_simulator.json
- **relations:** kcat line-delimited trap; env-prefix trap; spawn+call input shape.
- **verify-later:** whether these scripts were promoted anywhere canonical (they live in drafts/ and this docs dir).

<!-- SOURCE: U08_travelling_docs.md -->
### Manual spawn+call input shape — satisfy the contract top-level AND the workflow's input_data.spec.*
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016b entry with the 081/082/083 trigger trilogy as evidence (contract violation vs empty-context generation vs correct).
- **what:** `call_agent` validates input_mapping output against the target's input_contract at TOP level, while workflows for work-item-driven agents read `input_data.spec.*` — so manual invocation must provide required fields both top-level and inside a spec object (or better, drive the agent via its designed work-item trigger). Flagged as a latent design smell: contract and workflow should agree. Companion: `store_generated_component` keys regeneration on the LLM's EMITTED function — a mismatched name INSERTs a stray duplicate.
- **sources:** 016b_debugging_guide_7_3_(7).md#spawn-call-entry
- **relations:** diagnosis subject threading (same contract rule); trigger scripts.
- **verify-later:** call_agent contract validation code; component-creator contract vs workflow paths.

<!-- SOURCE: U09_adoption.md -->
### Wrapper-orchestrator pattern (site-adoption-orchestrator)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "The wrapper-orchestrator pattern also landed here: site-adoption-agent … now runs under a site-adoption-orchestrator wrapper (spawn_adopter → call_adopter → complete)" (FOCUS_adoption_fidelity_and_variants, 2026-04-22).
- **what:** Every pod-running agent needs a parent that spawned it; coordination work (conditionals, HITL, spawn/call) stays in-chassis, substantive work (LLM, crawls, heavy DB) runs in a spawned pod. site-adoption-agent was the outlier running in-chassis and got a thin `site-adoption-orchestrator` wrapper modelled on med-export-orchestrator. Known sibling defect: all four med-* wrappers carry a broken `{"input_data": "input_data"}` double-wrap mapping.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed, old2/HANDOFF_2026-04-22#wrapper-orchestrator-pattern
- **relations:** adoption pipeline; baseline rule recorded in 001_development_guide
- **verify-later:** agent_definitions `site-adoption-orchestrator`, med-* wrapper input mappings

<!-- SOURCE: U09_adoption.md -->
### insertWorkItem two-strike rule and dedup slot
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Key enabling discovery (insertWorkItem, line 46319): a built-in two-strike rule — an item_key with ≥2 terminal attempts in 7 days is inserted as unresolved… anything <3h after a terminal item is suppressed; live dups blocked by ON CONFLICT (site_id,item_key) WHERE status NOT IN terminal."
- **what:** Platform-wide anti-churn machinery on work-item insertion: partial-unique dedup index on (site_id, item_key) over non-terminal statuses, a 3-hour suppression window after a terminal item, and a two-strike escalation to `unresolved`. Discovery checks and retriggers lean on this instead of implementing their own loop protection. Related facts: `needs_human_review` is non-terminal (holds the dedup slot); terminal set = complete/failed/verified/rejected/wont_fix/unresolved; use DELETE+INSERT not ON CONFLICT against the partial index.
- **sources:** running_notes_15(10)#part-8, HANDOFF_2026-06-09#key-references, FOCUS_directory_builder_and_list_components.md#schema-quirks
- **relations:** sectionless durability S1; work-item lifecycle
- **verify-later:** insertWorkItem in chassis; idx_swi_dedup definition

<!-- SOURCE: U10_imagery.md -->
### ExtractActionInputs Strategy-0 explicit dot-paths lesson
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "FIXED workflow-only: SQL_2026-07-12_asset_deployer_explicit_paths.sql… Standing lesson: give ExtractActionInputs actions explicit dot-paths; never trust the search" — with the dispatch-shape (`input_data.spec.*`) gap recorded, not fixed.
- **what:** ExtractActionInputs' aggressive recursive field search matched a stale `purpose` elsewhere in collected_data, so the sprite sheet deployed as a 900×900 hero-config JPG despite the child receiving purpose='sprite_sheet' — explicit Strategy-0 dot-path config values are resolved first and win. Latent siblings: items dispatched via build-dispatch-loop carry payload under input_data.spec.* which the explicit paths miss; historical spawned deploys may have silently used hero dimensions (May icons' file dims worth checking).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-31, SQL_2026-07-12_asset_deployer_explicit_paths.sql, HANDOFF_imagery_best_in_class.md#I2.1
- **relations:** dispatch input contract; deploy_image_asset defaults ("purpose":"hero").
- **verify-later:** asset-deployer deploy_asset step config paths; datahelpers extraction strategies.

<!-- SOURCE: U10_imagery.md -->
### Manual agent-trigger pattern (kcat orchestrate; never hand-roll spawn+call)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "the documented system.intake pattern is STALE (topic doesn't exist) — the working mechanism is the kcat trigger script pattern" (Turn 18); "Do NOT hand-roll spawn_agent+call_agent inline workflows" (Turn 26 lesson).
- **what:** Manually triggering agents means an `action=orchestrate` envelope to `system.agent.generic.requests` with config.agent_type + input_data — known-good for improvement-loop, webdesign-agent, rerender-pages. Hand-crafted inline spawn+call parents fail because the spawned child runs its workflow on INIT and idles before the call arrives; work destined for spawned handlers must route through work items + dispatch instead.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-18/#Turn-26
- **relations:** dispatch input contract; brand-head activation was the proving case.
- **verify-later:** 033_rerender_pages_trigger.sh precedent; system.intake topic absence.

<!-- SOURCE: U10_imagery.md -->
### psql read-only PreToolUse gate
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "added a PreToolUse permission hook (.claude/hooks/psql_readonly_gate.py)… tested against a 20-case matrix and proven live" (Turn 3, 2026-07-08).
- **what:** Agent-session tooling: a hook auto-approves read-only SELECT/`\d` psql via the exact kubectl-exec form while mutations still prompt the human — reducing friction for the DB ground-truth checks every session performs. Session auth expires ~daily (runbook A1 re-login ritual).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-3, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNBOOK_imagery_best_in_class.md#A1
- **relations:** context-bundle seeding; operator runbook rituals.
- **verify-later:** .claude/hooks/psql_readonly_gate.py; settings.local.json hook wiring.

<!-- SOURCE: U12_docs024_archives.md -->
### ExtractActionInputs nested-source collision affects required fields too (corrected scope)
*(merged from 2 independent findings)*
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Live `001_development_guide(5).md`'s opening consolidation note states it "supersedes the prior copy (which still had the older 'Field Name Collisions' wording — the nested-source collision affects **required and optional** fields, corrected here)"; live text: "This loop iterates the full field list — both Required and Optional. It does not distinguish between them."
- **what:** Two independent archived drafts (`old/older1/001h_development_guide_new_agents_v8.md` and `old/001_development_guide.md`) both claimed the nested-source lookup collision in `ExtractActionInputs` (an unmapped field silently matching `site_record.<field>`/`input_data.<field>`) applies only to optional fields. The live doc corrects this: the nested-source loop iterates the full field list regardless of Required/Optional status; required fields (e.g. `site_id`) carry the same latent risk, it's just usually masked because earlier resolution strategies (0-2) resolve them first. The live doc adds a "latent risk (required field)" example and recommends collision-free names (`target_site_id`) for new required fields, while leaving existing code alone unless it actually misbehaves.
- **sources:** old/older1/001h_development_guide_new_agents_v8.md#"Field Name Collisions"; old/001_development_guide.md#"Field Name Collisions"; docs024_key_docs_latest/001_development_guide(5).md#"Field name collisions"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Note on the target_site_id input field name"
- **relations:** ExtractActionInputs / datahelpers resolution cascade (Strategy 0-4); whole-blob input_data anti-pattern; target_site_id naming convention
- **verify-later:** confirm `platform/orchestration/datahelpers/action_inputs.go`'s nested-source loop still doesn't distinguish Required/Optional.

<!-- SOURCE: U12_docs024_archives.md -->
### Whole-blob input_data passthrough mapping (anti-pattern)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive presents `{"input_data": "input_data"}` as a working pattern used by three orchestrators; live: "It does not do what it looks like it does... map each expected field by name."
- **what:** A wrapper-orchestrator shorthand documented as valid in the archive. Live guide identifies it as broken (double-nests the caller's data) and replaces it with explicit per-field mapping using `?`-suffixed optional keys.
- **sources:** old/001_development_guide.md#"Standardized Input Extraction"; docs024_key_docs_latest/001_development_guide(5).md#"Map fields individually, not the whole input_data blob"
- **relations:** ExtractActionInputs nested-source collision; input_mapping `?` suffix convention
- **verify-later:** grep current agent_definitions for `"input_data": "input_data"` mapping still in use.

<!-- SOURCE: U12_docs024_archives.md -->
### Loop array-iteration internals (early investigation notes)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Archive's "Open Investigation" records partial early findings about `LoopAction`/`handleLoopExpansion`/`setLoopVariable`; the fully worked-out version is documented completely in `archive_april_26/014_loop_mechanisms_guide.md`, itself fully absorbed into the live dev guide as Appendix C.
- **what:** Records still-open Feb-2026 questions about how loop expansion and substep naming work internally, later fully resolved and documented.
- **sources:** archive_april_26/006b_useful_notes_handoff_summary.md#"Open Investigation"; archive_april_26/014_loop_mechanisms_guide.md
- **relations:** claim_work_item / self-spawning dispatch loop (above)
- **verify-later:** none — superseded by a complete, later document.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Roadmap-phases enforcement gap (routed to builder thread)
- **category:** development-guide
- **status-signal:** deployed (as a documented finding; fix owned elsewhere)
- **status-evidence:** "RECLASSIFIED: this is the builder thread's MAIN queue item now (item 6...)" (RUNBOOK(9)#2026-07-07 later still)
- **what:** Guidelines 001 (~lines 1503-1560) already define a Tier-3 roadmap-with-phases mechanism, but `082_submit_domain_unified.sh` has no `--roadmap` entry point and `build-site-planner`'s prompt has no else-branch for an absent roadmap — so absent a roadmap, phase constraints vanish rather than degrade gracefully. Confirmed in code as an absent decision point, and routed to the builder thread as a relay-wide fix rather than fixed by this workstream.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#ROOT CONTEXT FOUND IN THE DOCS, fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#2026-07-07 later still
- **relations:** abandoned pilot candidates
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner prompt roadmap_brief block

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: error_step must be inside step config
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "error_step goes INSIDE a step's config — step-level error_step is silently ignored (001 §16)" (RUNBOOK(10)#Inherited gotchas)
- **what:** A platform-wide workflow-authoring gotcha: a step's `error_step` must be nested inside that step's `config` object; a step-level sibling `error_step` key is silently ignored. Live dormant instances found in page-build-handler's several steps — flagged to be corrected whenever that workflow is next touched.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#COLLISION SURFACE
- **relations:** max_tokens placement gotcha (same family)
- **verify-later:** grep/inspect `error_step`; `config`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: max_tokens must live inside ai_service
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "max_tokens at a step-config's root is DEAD CONFIG for execute_llm_prompt" (RUNBOOK(10)#BENCHMARK RUN 2)
- **what:** `execute_llm_prompt` reads `max_tokens` only from the agent's top-level config or from inside the step's `ai_service` block; a root-level step-config `max_tokens` is silently ignored and the Anthropic client defaults to 2048 output tokens. This capped the diagnose-agent verdict step at 2048 tokens through all five benchmark runs undetected, and truncated the fix-proposer's plan mid-JSON twice.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 2, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** error_step placement gotcha; fix-proposer's truncation failures
- **verify-later:** grep/inspect `execute_llm_prompt`; `max_tokens`; `ai_service`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: verify deployed contents against the pod, never tag/git
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "verify deployed contents against the POD binary, never the tag, never git" (NOTES(10)#Turn 23)
- **what:** A same-tag deploy trap: bumping source without bumping `IMAGE_TAG` means `rollout restart` reuses the cached image, so a reported "deploy" can silently ship a stale binary. The only reliable verification is grepping the running pod's binary for control strings. Caught the v1.0.1107→v1.0.1108 "first deploy" being a no-op.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Live gotchas
- **relations:** round-counting scope bug
- **verify-later:** grep/inspect `IMAGE_TAG`; `rollout restart`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: rebalance window after chassis restart
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "after make release redeploy-agents, wait for the chassis deployment to settle before firing a diagnosis" (NOTES(10)#Turn 9)
- **what:** Firing an orchestration within roughly 300 seconds of a chassis pod restart risks the spawn's init response falling into a Kafka consumer-rebalance window and dying silently — cost 8 hours of debugging the first time. Standing workaround: wait ~300s after any deploy before firing a run.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 9, fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** same-tag deploy trap
- **verify-later:** n/a — process/design record, no separate code artifact beyond what is already cited in sources

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Development-guide gotcha: BST/UTC timestamp mismatch
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "orchestration_states.last_activity is timestamp WITHOUT time zone... dev host runs BST while DB is UTC" (RUNBOOK(10)#Inherited gotchas)
- **what:** `last_activity` is stored without time zone while `created_at` is timestamptz, so `NOW() - last_activity` arithmetic is silently wrong by the local UTC offset; combined with the dev host running BST against a UTC database, a run can appear to have finished before it started.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 6
- **relations:** rebalance window gotcha
- **verify-later:** grep/inspect `last_activity`; `created_at`; `NOW() - last_activity`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Observability signature-fields pattern (proving which code path ran via collected_data)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** design_actions_observability_patch.md code diff applied; FOCUS_visual_pipeline doc shows the fields live in the result map
- **what:** A reusable debugging convention: when patching a code path whose old/new behaviour is otherwise indistinguishable, write new marker fields into the result map (flowing into `orchestration_states.collected_data`) that are absent from the old code — their presence proves the new code executed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Observability, js_snippets_news_gaswholesalers/old/design_actions_observability_patch.md
- **relations:** CSS component-list fallback bug
- **verify-later:** grep/inspect `orchestration_states.collected_data`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### needs_section_data review items appearing on successful builds (open question)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** "Worth a separate look at why a successful structured build still raises a section-data review item" — listed open, never investigated
- **what:** Even the clean `faq-test` isolated build spawned a child `needs_section_data` work item with `status=needs_human_review` and no `handler_agent`, unexplained.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** Post-build validation of structured components; isolated build test methodology
- **verify-later:** needs_section_data work-item creation path, wont_fix auto-resolution pattern

<!-- SOURCE: U14_docs019_runbooks.md -->
### Cross-module engine port procedure
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** design_diagnosis_loop(7) §4c "surfaced FOUR build errors in the first real make build-core-manager … None of these four were logic bugs … doing Steps 1–4 in order pays it once."
- **what:** The validated sequence for moving a package between Go modules (contextkit internal/diagnose → chassis pkg/diagnose): (1) copy the WHOLE package as a unit and diff file lists; (2) rewrite the moved-package import path everywhere; (3) build+test the package alone before the binary; (4) grep every shared-package call the new code makes against the REAL helper surface (datahelpers.ExtractStringSlice didn't exist; compose ExtractStringListHelper(ExtractNestedField(...))). The chassis copies keep the agentchassis import (step.go) while the prototype keeps contextkit's.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4c; docs019/RUNBOOK(31)_diagnosis_loop.md#current-position (#1 build note)
- **relations:** ReadSymbolBody dual placement; module-copy drift
- **verify-later:** pkg/diagnose file list vs contextkit/internal/diagnose

<!-- SOURCE: U15_docs019_running_notes.md -->
### Curated best-in-class standing expectation
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Best-in-class/curated-list idea homed: standing expectation (guides+tools+news+non-affiliate curated top-N) + 'not-original-can-still-be-best' clause → 001_development_guide" (NOTES_running_synthesis_v4(39).md, 2026-07-07).
- **what:** A proposed platform-wide addition to the development guide requiring every commerce-shaped domain to carry a baseline of guides, tools, a news feed, and a curated non-affiliate top-N list — enforced the same way as the roadmap gap (relay-wide strategist/planner prompts, not per-message or the constitution) — with the explicit doctrine that "useful-but-unoriginal still counts as best-in-class."
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-07 "pilot candidate 2" entry; NOTES_running_fixloop(9).md "Builder queue item 7" references.
- **relations:** Roadmap-phase enforcement gap; diagnosis→fix loop workstream founding.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Building-discipline edge cases (pre-registered engineering checklist)
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "Self-referential structures need a cycle guard... A multi-step apply must be all-or-nothing... Bulk operations need bulk confirmation" (principles(59) §Building discipline).
- **what:** A checklist of edge cases the design docs insist be caught before building anything with self-modifying or autonomous behaviour: cycle guards on any parent-link/version-chain walk; transactional multi-write-plus-event apply (outbox pattern) so a crash can't leave a dangling row with no event; reading a consistent point-in-time snapshot when assembling from multiple tables; "one live thing per target, all the way down" (dedup at every layer, not just the queue); bulk-confirmation for large batches; filtering transient/infra failures out of any trust-affecting evidence signal; and "tell not-yet apart from broken" (missing-because-unonboarded degrades gracefully, missing-because-malformed fails loudly).
- **sources:** NOTES_running_synthesis_principles(59) §Building discipline (shared preamble).
- **relations:** Trust ratchet & capability ceiling model; untested-code / behaviour-testing discipline.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Untested-code / behaviour-testing discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "PRINCIPLE 2026-06-13: untested code is a liability, surfaced by the dedup -move bug... COMPILE/gofmt/vet prove syntax, NOT behaviour." (principles(59)).
- **what:** A hard-won, explicitly codified lesson (triggered by the dedup tool's silent destructive-flag bug) that compiling/gofmt/vet only prove syntax, never behaviour; any destructive CLI operation must be report-only by default; and Go's `flag.Parse()` stopping at the first positional argument is a specific, recurring footgun requiring manual value-flag-aware argument separation in every CLI that takes a positional followed by flags — audited across `dedup`, `thin_versions`, `resolve_targets`, `embed`, `assembler`, `fuse`, `eval_targets`, `dbcontext`.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 "PRINCIPLE" entry; NOTES_running_synthesis_v2(36).md 2026-06-14 "runbook code audit" (a second, independent instance of the same class of bug).
- **relations:** Docs archiving toolchain (dedup); building-discipline edge cases.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Cross-module copy-drift lesson
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "This is the 5th instance of the same root pattern (drafted/validated artefact vs what's on disk): import path, missing file, stale sibling, assumed helper API, now stale CLI." (NOTES_running_synthesis_v2(36).md, 2026-06-17).
- **what:** A hard-won lesson from porting the diagnose engine into the chassis across a real module boundary, surfacing five DISTINCT failure classes in sequence — a wrong import path on a copied file, an entire file silently omitted, a stale (pre-refactor) copy of a sibling file, an assumed helper-package API that didn't actually exist (`datahelpers.ExtractStringSlice`), and a stale CLI binary predating a library change — all of which passed silently in the source module and surfaced only on first build/run in the target. Durable prevention recorded: copy the WHOLE package directory as one unit and diff the file list, rather than cherry-picking files across versions; grep every shared-package call against the real package before authoring, not after.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 (four consecutive build-gap entries); mirrored in `016_additions_assumed_helper_and_cross_module.md` per the notes.
- **relations:** Diagnosis-loop chassis integration architecture; untested-code / behaviour-testing discipline.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Schema-before-SQL discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Schema: bundle CODE names tables; only \d gives columns AND persistence (hit 4×: page_id, no status column, file_count, the 3-NULL workflow columns)." (v2(36) STATE DIGEST standing lessons).
- **what:** A recurring, explicitly named discipline that code reliably names which DB tables are involved, but only a live `\d` (schema dump) gives real column names and reveals whether a field is persisted at all vs. computed at runtime — hit repeatedly across this project (a wrong `page_id` column, an assumed-but-nonexistent `status` column on `site_plan_sections`, a wrong `fileCount` vs `file_count` JSON key, and the workflow-column misassumption) and eventually generalised into "real rows/examples beat prose/inference" as its own standing lesson.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST "Standing lessons"; NOTES_running_synthesis_principles(59) multiple 2026-06-14 gamesdesign-diagnosis entries.
- **relations:** Workflow default_config location convention; DB discipline / snapshot_agent convention; gamesdesign silent-no-op bug.

<!-- SOURCE: U15_docs019_running_notes.md -->
### "Every agent is an orchestrator" spawn pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "ORCHESTRATOR vs long-running service (user's Q): follow 'every agent is an orchestrator' — a thin diagnose-orchestrator spawns a diagnose-agent worker pod... exactly as site-adoption-orchestrator spawns site-adoption-agent." (v2(36), 2026-06-17).
- **what:** A standing platform convention that any substantive in-chassis work (multiple iterations, LLM calls, minutes of runtime) must be wrapped: a thin coordinator/orchestrator agent spawns a dedicated worker agent as a Job pod that runs the actual work and replies to the caller's own responses topic, rather than building a bespoke long-running service that would duplicate the chassis's spawn/await/topic machinery.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 "chassis integration... STEP ZERO search."
- **relations:** Diagnosis-loop chassis integration architecture; workflow default_config location convention.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Real-rows-beat-prose-or-assumption discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Real rows/examples beat prose/inference: the agent_definitions workflow lives in default_config — the dev guide prose didn't say which column; the row did." (v2(36) STATE DIGEST standing lessons).
- **what:** A generalised standing lesson (distilled from several specific incidents across this file set) that when a dev-guide or design doc's prose is ambiguous or silent about an implementation detail, the correct source of truth is a real, live example row/file, not inference from the prose — repeatedly the deciding move that caught a wrong migration or wrong action draft before it was applied.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST; NOTES_running_synthesis_principles(59) "GROUNDED... tools/provenance/docs design corrected" entries (same pattern, different subsystem).
- **relations:** Workflow default_config location convention; schema-before-SQL discipline; child-completion result key convention.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Workflow lives in default_config, not the workflow columns
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** NNN_seed_diagnose_agents(2) header: "confirmed: those agents put their workflow there; the task_workflow / orchestrator_workflow columns are NULL"; the move/fix/restore migration sequence applied.
- **what:** The loader reads `agent_definitions.default_config` ({workflow:{start_step,steps}, processing_mode, timeout_seconds}); the three workflow columns are dead for working agents. A workflow seeded into orchestration_workflow silently never loads — this bit the diagnose pair (seeded 2026-06-20 into the wrong column, then the orchestrator's workflow was lost entirely during the move and had to be re-seeded). Key correction learned by reading a real row rather than docs.
- **sources:** NNN_move_diagnose_workflow_to_default_config(1).sql; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md (2026-06-14/17)
- **relations:** schema-before-SQL discipline; spawn-consumed columns lesson (sibling class)
- **verify-later:** workflow loader code path; whether the three columns are still consulted anywhere

<!-- SOURCE: U16_docs019_design_plans.md -->
### Confirmed chassis workflow model (group agents, promotion pattern, wrapper orchestrator)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Migration(19) "Confirmed model (from the guideline docs …)" plus 2026-06-09 changelog confirming against two real agent_definitions rows and the action code.
- **what:** The model the migration work verified: workflows are declarative JSON steps in default_config; a generic config-driven action library (query_database, spawn_agent, call_agent, loop, conditional, rag_lookup, work-item lifecycle) is reused before writing Go; LLM-assisted checks group into one agent per shared context load (explicitly rejecting a registry of mini-action agents), promoted to spawned sub-agents only when one needs independence (a one-line workflow change); the wrapper orchestrator (spawn→call→complete) is the canonical small form; spawning is spawnAgentKubernetesJobFromDefinition with per-spawn job topics. Reuse discipline is encoded as queries (search agent_definitions; default_config::text ILIKE '%<action>%').
- **sources:** PLAN_workflows_and_actions_migration(19).md (confirmed model + changelog); DESIGN_diagnosis_loop_chassis_integration(6).md#0
- **relations:** diagnose pair; index-orchestrator; onboarding agents (all reuse it)
- **verify-later:** 001/002 guideline docs (canonical home, other units)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Development Guide (agent-build daily reference)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** 001(0) consolidation note "This is the canonical 001_development_guide. It supersedes the prior copy"; archive copy with live successor in docs024_key_docs_latest
- **what:** The consolidated practical reference for building/debugging/maintaining agents: core design principles (agents own their domain, callers pass raw data, workflows simple with complexity in Go, actions are the unit of work, spawn-before-call, reply-to-caller's-topic), a new-agent checklist, migration guide, and 20+ lessons-learned bug entries.
- **sources:** WM/001_development_guide(0).md#core-design-principles, WM/001_development_guide(0).md#checklist-for-new-specialist-agent, WM/001_development_guide(0).md#summary-of-rules-for-the-dev-guide
- **relations:** superseded by docs024 live 001; STEP ZERO; wrapper-orchestrator; loop mechanisms
- **verify-later:** platform/orchestration/actions/*

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### STEP ZERO — reuse-before-create discipline
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "Pre-Flight: Does This Already Exist? … Real example: We built asset-deploy-agent … The existing asset-deployer already did the same thing. Three hours wasted"
- **what:** The mandatory pre-flight before creating any agent/action/function: search `agent_definitions`, the action registry, Go funcs, gate functions and workflows for every noun in the proposed name, document what was found, and prefer patching an existing thing. Includes the canonical field-path resolution rule (use `datahelpers.ExtractNestedField*`, don't add another).
- **sources:** WM/001_development_guide(0).md#pre-flight-does-this-already-exist-step-zero, WM/001_development_guide(0).md#field-path-resolution-use-the-canonical-functions, WM/001_development_guide(0).md#reuse-before-creating
- **relations:** Development Guide; STEP-ZERO-for-standards (curation); reliability cascade (reuse tier)
- **verify-later:** registry.go; datahelpers package; isStorageEnabledAgent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Standardized input extraction (ActionInputSpec, ? optional, field collisions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(0) "The three layers … input_mapping / input_fields / ActionInputSpec"; "Field name collisions … runs a nested-source loop late in its resolution chain"
- **what:** Three layers move data into an action: caller `input_mapping` (with `?`-optional destination keys), action `input_fields`, and `ExtractActionInputs(spec)` with a documented resolution chain. The nested-source loop iterates required AND optional fields, so names like `site_id`/`content_data`/`domain` can silently resolve from the wrong nested source — prefer collision-free names.
- **sources:** WM/001_development_guide(0).md#standardized-input-extraction, WM/030_phase1_plan_and_reconciler(4).md#note-on-the-target_site_id-input-field-name, WM/016_debugging_guide_v2_44.md#0
- **relations:** dispatch loop input_mapping; ? suffix; target_site_id convention
- **verify-later:** datahelpers/action_inputs.go; ResolveInputMapping; coordinator.go resolveInputMapping

<!-- SOURCE: U17b_docs019_gofiles.md -->
### thin-slice constitution (always-on rules doc)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "Included in full in every bundle... Later it becomes the `standards` rows with `scope = constitution`; the content is the same." (thin_slice_constitution.md)
- **what:** The flat-file version of the chassis's always-on rules, pasted into every assembler/bundle output: reuse-before-recreate, fix-structural-not-symptoms, every-agent-is-an-orchestrator, no subworkflows-in-SQL (spawn sub-agents instead), the snake_case/kebab-case naming split, storage conventions (text+CHECK enums, version+previous_version_id, deleted_at soft-delete), logging rules (no `logger.Debug`, log the orchestration_id/correlation_id), deployment path (GitHub → Actions → Backblaze S3), and plain/pragmatic generated-text tone. Task-specific 003 contracts are listed but pulled in only when a task touches them.
- **sources:** contextkit/thin_slice_constitution.md
- **relations:** assembler (always includes it), docselect.go (adds the task-specific 003 sections this doc defers), contracts-and-standards (003)
- **verify-later:** whether the constitution has since actually migrated to `standards` rows with `scope = constitution` as the doc anticipates, or is still the flat file

<!-- SOURCE: U18_sql_for_agents.md -->
### v1 monolithic LLM-chain site builders (website-builder, domain-analyst, site-architect, content-creator, html-developer, multipage-wrapper, html-assembler, site-deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026_pageflow_builder.sql renames multipage-website-builder v3 → pageflow-builder ("Component-based website builder... uses DB components for structure, LLM only for content"); v2/027_old_agent_definitions.sql captures the whole v1 fleet as "old"; root files never patch these agents again.
- **what:** The first architecture (2025-11/12): a website-builder orchestrator spawns a chain of one-LLM-call specialists — domain-analyst (audience/tone JSON), site-architect (page structure + colours), content-creator (copy JSON), html-developer (whole-page HTML), multipage-wrapper (file map), site-deployer (git commit). Everything is free-form LLM output; no component library, no DB page records.
- **sources:** sql_for_agents_v1/004_website_builder.sql; sql_for_agents_v1/005_domain_analyst.sql; sql_for_agents_v1/008_html_developer.sql; sql_for_agents_v2/027_old_agent_definitions.sql
- **relations:** superseded by pageflow-builder (component-based); site-deployer contract survives in 011_site_deployer.sql
- **verify-later:** agent_definitions rows for these types (is_active, deleted_at); whether any workflow still references them

<!-- SOURCE: U18_sql_for_agents.md -->
### Batched multi-page generation (multipage-website-builder) and chunked HTML generation
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** v2/026 renames its v3 to pageflow-builder; the batch-of-4-pages prompts (generate_batch_1..5) appear only in v1/015 and v1/017 snapshots.
- **what:** Anti-token-limit strategies from the v1 era: build 20-page sites by generating pages in five batches of four ("Return as JSON map of filename to HTML"), with shared CSS generated once and injected at assembly; html-developer-chunked generated structure/styles/sections in separate calls. Both are ideas the component architecture made unnecessary.
- **sources:** sql_for_agents_v1/015_example_20_page_workflow.sql; sql_for_agents_v1/017_multipage_website_builder.sql; sql_for_agents_v1/014_html_developer_chunked.sql
- **relations:** replaced by pageflow-builder per-page loop and later the one-item-per-run dispatch loop
- **verify-later:** none needed — historical

<!-- SOURCE: U18_sql_for_agents.md -->
### Remove-loops plan: input_mapping, contract validation, sequential_fan_out, page-builder worker
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** input_mapping conversion executed in 030/030b/023; contracts added across many files ("These contracts define what each agent expects... fail fast"); but build_pages_loop still present in 026 and sequential_fan_out/page-builder never appear in any later agent file — that half is effectively abandoned in favour of the dispatch loop.
- **what:** A four-phase plan to replace loop/substep injection: (1) explicit `input_mapping` instead of `input_fields` path-hunting plus runtime input/output contract validation with hard fails and `__raw_message__` deprecation; (2) a `sequential_fan_out` action spawning one child orchestration per page; (3) a page-builder worker agent; (4) rewire pageflow-builder. Phases 1 landed; phases 2–4 were superseded by the site_work_items dispatch-loop architecture, which achieves the same "one visible orchestration per unit of work" goal differently. 001_validator_sql.sql is a jsonb_path_query audit extracting every field path referenced in workflows.
- **sources:** 001_remove_loops_in_workflow.md; 001b_implementation_plan.md; 030_input_mapping_changes.sql; 030b_remaining_agents_needing_input_mapping; 001_validator_sql.sql
- **relations:** input contracts appear in nearly every agent file (002, 011, 022, 024, 025, 029...); dispatch loop (051) is the spiritual successor of sequential_fan_out
- **verify-later:** chassis code: contract validation enforcement; whether `sequential_fan_out` action exists in the registry; `__raw_message__` fallback removal status

<!-- SOURCE: U19_sql_tables_components.md -->
### site_work_items unified work queue and lifecycle
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full DDL (023_site_work_items) with dedup index, plus dozens of live operational patches (resets, handler re-routing, attempt bumps) against real sites.
- **what:** Every piece of platform work is a row: source (planner/discovery/content_feed/manual/improvement/side_effect/human/validation), pipeline (originally `domain`, later renamed), item_type, severity, spec JSONB, target refs (page/component/entity/url), triage enrichment (impact, resolution_path, suggested_action, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete/pending_verify/verified/failed/rejected/wont_fix plus 'blocked' (handler missing), dependencies (depends_on UUID[], parent/related/batch), attempts, and deterministic item_key with a partial unique index for dedup among non-terminal items. A same-structure archive table receives terminal items.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#approval_mode; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle
- **relations:** dispatch loop; claimed-item timeout; content_feed_items; archiver; approval_mode; processing_tier.
- **verify-later:** current status distribution; pipeline column rename (`domain` dropped in 018).

<!-- SOURCE: U19_sql_tables_components.md -->
### Work item archival
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** archive_completed_work_items(age, batch) function + archiver agent definition + daily scheduled task, with schema-sync ALTERs and FK handling (parent self-ref cleared, content_feed_items references deleted).
- **what:** Terminal work items (complete/failed/wont_fix) older than a configurable age move to site_work_items_archive in batches, keeping the live queue small. Function handles column drift between live and archive tables explicitly.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-archiver
- **relations:** work queue; scheduler.
- **verify-later:** archiver task enabled; archive row counts.

<!-- SOURCE: U19_sql_tables_components.md -->
### build_queue site seeding
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Phase-0 Block A table with direction semantics enumerated.
- **what:** Domain-level intake queue for new sites: a row per domain with direction JSONB (null | {objective} | {adopt_from} | {fork_from} | {brief_complete...}), status and priority. seed_build_queue reads it, creates site records and initial work items according to direction — the entry point into the work-item pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#2
- **relations:** work queue; adoption pipeline (adopt_from); onboarding-config (brief_complete).
- **verify-later:** seed_build_queue action; build-pipeline-trigger seeding behaviour.

<!-- SOURCE: U19_sql_tables_components.md -->
### Loop-action dispatch (migration 071)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Applied UPDATE of build-dispatch-loop default_config: "Step-chaining... processes only one work item per trigger. The loop action is proven in maintenance-triage and pageflow-builder."
- **what:** The dispatch loop loads all dispatchable items upfront (dependency-filtered, priority-ordered, max 50) and iterates with the `loop` action running a sub_workflow per item: claim → check_claim → spawn_handler (dynamic agent type from current_item.handler_agent) → call_handler → mark_complete/mark_failed, with continue_on_error. Introduces item_variable scoping (current_item.*) and optional `?`-suffixed input_mapping fields silently skipped for handlers that don't need them (section-editor compatibility).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql#migration-071
- **relations:** work queue; spawn-orchestrator pattern; claimed-item timeout.
- **verify-later:** build-dispatch-loop live config; loop action implementation.

<!-- SOURCE: U19_sql_tables_components.md -->
### Spawn-orchestrator thin-wrapper pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated across med pricing (scrape, discover, map, export orchestrators): spawn_agent (role) → call_agent (target_role, input_mapping passthrough, timeout) → complete_workflow; scheduled tasks target the orchestrator, not the worker.
- **what:** The standard shape for burst workloads: a permanently-resident category pod receives the trigger, a thin orchestrator workflow spawns a temporary worker pod of the right agent_type, forwards input_data, awaits the result, and completes — worker terminates (idle_timeout 0). Non-secret worker config rides env_vars on the agent definition; secrets come via spawn_actions secretKeyRef.
- **sources:** docs/agent_docs/sql_for_tables/034_vet_med_price_scrape_orchestrator.sql; docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql; docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** scheduler; agent definitions; vet med pipelines.
- **verify-later:** spawn_agent/call_agent actions; ON CONFLICT (type, version) upsert convention.

<!-- SOURCE: U20_legacy_docs_a.md -->
### spawn_agent — database-definition-driven Kubernetes Job spawning
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.042: "All deployment specs now come from database… No Code Changes for New Agents — Just insert into agent_definitions"; spawn walked through line-by-line in spawn_actions2.
- **what:** SpawnAgentAction reads the child's agent_definitions row (image repo/tag, resources, health config, env vars, default workflow), inserts an agent_instances row in the client schema, creates job topics, launches a K8s Job with topic env vars, sends the initialize message, and returns spawn results (agent_id, role, topics) into CollectedData for later call_agent lookup.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.045.spawn_actions4.refactor_into_functions.md
- **relations:** agent_definitions registry; job topics; role concept.
- **verify-later:** spawn_actions.go; client_{id}.agent_instances table.

<!-- SOURCE: U20_legacy_docs_a.md -->
### call_agent with role-based targeting
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.041: "role is the essential piece of information that links a specific task to a specific, previously spawned agent" with code walkthrough; used in all robot-hands workflows.
- **what:** CallAgentAction finds a previously spawned agent by searching CollectedData spawn results for a matching `target_role`, extracts its private requests_topic, and sends a `process` request there with await_response. Role acts as the within-orchestration nickname distinguishing multiple agents of the same type (adder vs multiplier calculators).
- **sources:** docs001_flow_general/README.041.role_flow.md; docs001_flow_general/README.018.flow7.roleflow.md; docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** spawn_agent; spawn step naming conventions; role-based agent pools proposal.
- **verify-later:** call_agent.go findAgentByRole/findTargetAgent.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Role-based agent pools / atomic work-claim queue (proposal)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** flow8 is pure proposal ("Migration Path: Phase 1…Phase 4"), never referenced as built in later docs; the design (work_items table, `claimed_by IS NULL` atomic UPDATE claim, role queues, failover pickup) is recognisably the ancestor of today's work-item pipeline.
- **what:** Instead of spawning agents tied to IDs, agents register roles/capabilities and claim WorkItems atomically from role-specific queues (`system.roles.{role}.pending`); unclaimed work survives agent death; pools scale elastically. "The role becomes the contract, not the agent ID."
- **sources:** docs001_flow_general/README.019.flow8.role_based_agent_pools.md
- **relations:** successor: work-item lifecycle / page-build-handler pipeline (development-guide, docs 001 current); scheduler-and-tasks concurrency groups.
- **verify-later:** work_items table and claim semantics in the current codebase — compare with this 2025 sketch.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Prompt resolution priority hierarchy
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.004 Part 3 "How the Flow Works Now" with three tested scenarios and log lines ("Using prompt from incoming message (Priority 1)").
- **what:** execute_llm_prompt resolves its prompt in priority order: (1) prompt passed in the incoming message/step config by the caller, (2) the agent's own prompt_template from agent_definitions, (3) workflow-step fallback. Lets parents override specialists while specialists keep good defaults.
- **sources:** docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** execute_llm_prompt action; agent_definitions default_config.
- **verify-later:** ai_actions.go ExecuteLLMPromptAction prompt lookup order.

<!-- SOURCE: U20_legacy_docs_a.md -->
### CollectedData normalisation and data_helpers safe-access layer
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full data_helpers.go source reproduced as "the new functionality" in docs002/0100 and used by all subsequent agents ("data_helpers.go functions ensure consistency", 0100c).
- **what:** One central layer (data_helpers.go) normalises every inbound message into a canonical CollectedData shape — `input_data` always at top level, system fields (`__execution_context__`, `__my_requests_topic__`, `__raw_message__`…) separated — and provides the only sanctioned accessors (GetInputData, GetStepData, GetMultipleStepData, GetFieldFromPath, TransformDataForAction, BuildRequestMessage/BuildResponseMessage/BuildInitializationRequest). Killed the `input_data.input_data` nesting chaos. Child input_data is always overwritten at top level — each agent's context is exactly what its parent sent (clean-slate encapsulation).
- **sources:** docs001_flow_general/README.070.a.centraliseddatanormalisation.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.024.flow14.input_data.md; docs001_flow_general/README.080.a.packaging_data.md
- **relations:** output_field/input_fields mapping contract; every action.
- **verify-later:** platform/orchestration/datahelpers package.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Agent groups — versioned, discoverable agent teams
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** README.060: "FindBestGroup… queries the database to find the best available version of that group, ordered by performance, usage, and version"; groups used in every website build; evolution/mutation service described but not evidenced live.
- **what:** agent_group_definitions rows are project recipes: a group_type, an agent_configs squad (role → agent_type), and an orchestration_workflow JSON, with integer versions as immutable snapshots (unique group_type+version). Requests name a capability (group_type) and the system picks the best version. An EvolutionService was designed to mutate groups into new versions with parent_id lineage and performance-based selection; the discovery/versioning part shipped, the evolutionary part appears aspirational.
- **sources:** docs001_flow_general/README.060.groupagents1.md; docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.062.groupagents3.databases.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion#versioning-model
- **relations:** workflow selection priority; groups-as-project-recipes; spawn_group; site manifest pinning group_version.
- **verify-later:** agent_groups vs agent_group_definitions tables (both exist with different shapes — 062 shows the split); discovery/agent_discovery.go FindBestGroup; evolution.go.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Workflow selection priority (inline override > group > agent default)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.060/061 define and implement selectWorkflow with the three-tier priority; HITL tests routinely use inline workflow overrides.
- **what:** processor.selectWorkflow resolves which workflow to run: (1) a full inline workflow in the message config (ephemeral/testing), (2) a group workflow found via group_type, (3) the agent's default workflow from agent_definitions. Keeps production versioned while allowing ad-hoc experiments.
- **sources:** docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.060.groupagents1.md; docs002_hitl_parallel/README.0106.hitl_multistep_approval.md (inline workflows in practice)
- **relations:** agent groups; SagaCoordinator.
- **verify-later:** processor.go selectWorkflow.

<!-- SOURCE: U20_legacy_docs_a.md -->
### agent_definitions registry (DB-driven agent config and versioning)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Dozens of INSERT/UPDATE statements across all four doc sets; constraint migration to UNIQUE(type, version) with previous_version_id (096b); category CHECK constraint managed in 0140.
- **what:** Every agent type is a row: type, display_name, category (constraint-checked: data-driven/code-driven/adapter/…), default_config (containing the workflow, ai_service model+provider, processing_mode, timeouts), capabilities, image_repository/tag (all agents share the agent-chassis image), resources, topics, health_config, env_vars, version + previous_version_id, task_workflow/orchestrator_workflow, delegation_preferences. Creating an agent is a database insert, not a code change.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.096b.robothandswebsite.md; docs003_firecrawl/README.0140.removing_constraint.md; docs001_flow_general/README.098.oldherocontentdefinition.d
- **relations:** spawn_agent; agent categorisation taxonomy (998); the docs024-era agent creation guide is the living successor doc.
- **verify-later:** agent_definitions schema and constraints today; how many of these early agent types still exist/are active.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Spawn/step naming conventions
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** README.044: "The naming conventions are now important because we're using them to find spawned agents" — spawn_ prefix required, unique step names with 3-letter suffixes.
- **what:** Workflow authoring rules: spawn steps must start `spawn_<descriptor>` (suffix hints the role), action steps use perform_/execute_/process_ prefixes and reference agents by role, and step names must be unique within a workflow.
- **sources:** docs001_flow_general/README.044.spawn_actions3.spawn_rules.md
- **relations:** call_agent role lookup; workflow authoring guide (development-guide successor docs).
- **verify-later:** whether current workflow JSON still relies on prefix conventions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### evaluate_condition — template-based conditional branching
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 0127/0128 document the working mechanism ("The orchestrator uses this to pick the next step from the next_step map") including Go text/template functions (and/or/not/eq/gt…) and a live website-analyzer group UPDATE.
- **what:** Workflow steps gain branching: evaluate_condition renders a Go text/template expression against CollectedData and returns true/false; `next_step` becomes a map {"true": …, "false": …}. Enables data-driven workflow paths (e.g. extract_structured? crawl_pages? previous step success?).
- **sources:** docs003_firecrawl/README.0127.conditional_branching.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** conditional_branch/conditional_route actions; route_by_field/conditional_call_agent (later, richer routing).
- **verify-later:** evaluate_condition in registry; coordinator support for map-typed next_step.

<!-- SOURCE: U20_legacy_docs_a.md -->
### MVP site builder pipeline (strategist → architect → content-creator → deployer)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** Full group SQL + per-agent Kafka payloads documented and run (boxing-tickets.com); renamed/extended into landing-page-builder and 6-step pipeline within docs004; today's site building is the work-item pipeline.
- **what:** The first end-to-end production pipeline: chief-strategist (LLM → build_plan JSON of functional sections), site-component-architect (assemble_from_library → empty semantically-tagged HTML template + content_requirements "shopping list"), content-creator (fills slots), deployer-agent (commit_to_git). Group workflow spawns all four then calls them in sequence, threading outputs through output_field/input_fields.
- **sources:** docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** grew into: + brand-designer, + briefing-agent, + html-assembler, + specialist architects; successor: current page-build-handler/work-item pipeline (see 100_content_page_build_handler_flow.md).
- **verify-later:** agent_group_definitions mvp-site-builder / landing-page-builder rows.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Dynamic agent routing (route_by_field / conditional_call_agent)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Both Go actions written with registry additions listed; conditional_call_agent chosen because it "wraps CallAgentAction internally — no coordinator changes needed".
- **what:** Data-driven agent selection inside workflows: route_by_field maps a dot-path field value to a next step via a routes table with default; conditional_call_agent reads e.g. brief_data…site_type, maps value→agent type (landing→landing-page-architect …) and calls that agent in one step, returning routing metadata.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023a.description_for_conditional_routing_etc; docs004_website_capture_project/006semantic_themes/README.024.conditional_step_routing.md
- **relations:** evaluate_condition (simpler predecessor); spawn_group dynamic group_type (group-level equivalent).
- **verify-later:** registry entries conditional_call_agent, route_by_field.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Groups as project recipes + immutable versioning + agent pinning
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 025 "decision" section: "A group isn't 'agents that work together' — it's a project recipe"; versioning model (immutable version rows, sites pinned to group_version, per-agent image_tag pinning) is design; UNIQUE(group_type, version) constraint added in 0100b — partially realised.
- **what:** Each buildable *kind* of output (landing page, content site, 11ty blog, ecommerce) is a self-contained group: its own agent squad, workflow, questionnaire, and outputs. Divergence in output structure/build/deployment means a new group, not conditional routing. Group versions are immutable snapshots; a site records the group_version that built it and rebuilds with it unless upgraded; groups may pin specific agent versions where stability matters. Duplication across similar groups is accepted for clarity.
- **sources:** docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion; docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md (constraint)
- **relations:** agent groups; site manifest; tool-lifecycle versioning is the analogous live discipline.
- **verify-later:** group version rows per group_type; any site→group_version reference.

<!-- SOURCE: U20_legacy_docs_a.md -->
### spawn_group action with DB group lookup and dynamic group_type
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 028 discovers an existing SpawnGroupAction (config-provided agents) and revises the new version (spawn_group_from_db.go) to align — DB lookup of agent_group_definitions, dynamic group_type_field from collected_data, questionnaire fetch.
- **what:** Spawning an entire agent group as a unit: original action spawned each configured agent and returned subtree info; enhanced version resolves the group definition (agents + workflow + questionnaire) from the database, with the group_type optionally taken dynamically from prior step output — enabling the intake orchestrator's dispatch.
- **sources:** docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** intake orchestrator; agent groups.
- **verify-later:** spawn_group vs spawn_group_from_db in codebase.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Agent/group categorisation taxonomy (category, status, domain_tags)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** Migration SQL 031_add_categorisation with CHECK constraints (category: builder/analyzer/collector/transformer/evaluator/researcher/workflow/monitor; status: active/experimental/deprecated/demo/template) and GIN-indexed domain_tags; no doc confirms it was applied.
- **what:** Organisational metadata over agent_definitions and agent_group_definitions: what the agent *does* (domain-agnostic category), its lifecycle status, and flexible domain tags — an early attempt at the registry hygiene the concept register itself now pursues.
- **sources:** docs004_website_capture_project/998categorisation/031_add_categorisation_to_tables.sql
- **relations:** agent_definitions registry; documentation-system indexing.
- **verify-later:** do the category/status/domain_tags columns exist?

<!-- SOURCE: U20_legacy_docs_a.md -->
### Aggregation patterns (aggregate_data, aggregator agent, input_from_collected_data)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** aggregate_data's failures traced (count:0 on verbose child responses, flow2); redesign to a spawned aggregator agent fed via input_from_collected_data path mapping (flow11); aggregate_webpage became the shipped variant for pages.
- **what:** Combining multi-step results: the local aggregate_data action broke against verbose child state objects; the redesign either normalises responses (data helpers) or delegates aggregation to a spawned aggregator agent whose call config maps CollectedData paths into its input. Response data keyed as response_{requestID} in CollectedData.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.010.flow.md
- **relations:** data_helpers NormalizeResponseData (the actual fix); aggregate_webpage.
- **verify-later:** aggregate_data current implementation.

<!-- SOURCE: U20_legacy_docs_a.md -->
### output_field / input_fields group-memory data mapping contract
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 016 "Packet Flow" walkthrough resolving the exact semantics ("Take the entire result from the Strategist and store it under build_plan_data… path is simply build_plan_data.build_plan_json") producing the "Golden Copy" workflows.
- **what:** The inter-agent data plumbing convention: a call_agent step's `output_field` names the key under which the child's entire result lands in group memory; the next step's `input_fields` selects which keys are passed on; consumers address values by `<output_field>.<producer's own output key>` paths. Most orchestration bugs of the era were mis-mappings of this contract.
- **sources:** docs004_website_capture_project/website_analysis/README.016.agent_definitions_002.md; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md
- **relations:** CollectedData normalisation; template rendering paths; note the execute_llm_prompt flattening quirk (input_fields:["input_data"] flattens, so templates use {{.domain}} not {{.input_data.domain}} — 029 fix).
- **verify-later:** call_agent output_field handling in coordinator.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Deliberate discovery + human-approved agent evolution
- **category:** development-guide
- **status-signal:** abandoned
- **status-evidence:** README.005 principles ("Deliberate discovery — only at planning and review stages; Human approval — all agent changes require approval; Performance-based evolution") never reappear as a mechanism in later eras.
- **what:** Early governance rules for agent self-modification: the system only creates/modifies agents when starting a new task type, after poor performance review, and always with human approval — no heartbeats or automatic decisions. Paired with per-group performance recording and version incrementing.
- **sources:** docs001_flow_general/README.005.discovery.md
- **relations:** agent groups evolution service; HITL; tool-lifecycle health checks are the modern relative.
- **verify-later:** none.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Website build overall plan v0 (first multi-agent website roadmap)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** README.050 is a 6-phase/12-step plan (domain-analyst → content-creator → html-developer, then site-architect/visual-designer/site-publisher, data contracts, spawn_group team) written against the calculator-era platform; every element was rebuilt differently in docs002–004.
- **what:** The first articulation of "build a website with agents": minimal 3-agent workflow, explicit JSON data contracts between agents, progressive enhancement, mock-LLM-first testing, upload_to_s3 deployment. Registers as the origin point of the entire site-building programme.
- **sources:** docs001_flow_general/README.050.overall_plan1.website_design.md; docs001_flow_general/README.001.actions.md (action inventory of that moment: many mocks — deploy_to_hosting, http_request, cache_lookup all fake)
- **relations:** superseded by MVP site builder, then the work-item pipeline.
- **verify-later:** none.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Data-path resolution problem (agent vs local action nesting)
- **category:** development-guide
- **status-signal:** superseded
- **status-evidence:** docs006/001 ("Local Action: CollectedData[\"wrap_multipage\"] ... input_data.site_files.wrap_multipage ← Extra layer!"); docs009/002 ("collected_data.spawn_x.call_x.spawn_y...result"); resolved later by input_mapping + ActionInputSpec (docs017/008).
- **what:** The recurring class of runtime failures where workflow config referenced CollectedData paths that didn't match where actions actually stored results — agent calls store flat, local actions add a step-name layer, and each spawn/call deepens nesting. Drove multiple generations of mitigation: workflow builder path computation, explicit output_field conventions, data-flow verification matrices, and finally standardized input extraction.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#The-Problem; docs009_site_interrogation_and_solutions/002_claude_discussion#C; docs015_data_flow_verification/001_data_flow_verification.md
- **relations:** ActionInputSpec/ExtractActionInputs; workflow builder; data-flow verification practice.
- **verify-later:** datahelpers.ResolveInputMapping and FindByPath in platform code.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Loop action via dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs010/022 integration guide with concrete file placements; docs010/023 walkthrough ("transforms the loop into individual workflow steps... Async-compatible... Recoverable... resume from exact step"); every later builder uses build_pages_loop.
- **what:** The `loop` action doesn't execute iterations itself — a coordinator-side expansion handler injects one workflow step per iteration×substep (generate_pages_loop_iter_0_research …), chained by NextStep, with loop_metadata in CollectedData, setLoopVariable placing the current item under the loop_var before each step, and a loop_complete step aggregating results into output_field. Design chosen (over in-process execution) because steps can await async agent responses and survive crashes/restarts as ordinary persisted workflow steps.
- **sources:** docs010_multitrack_flows_persona_architecture/021_loop_action_discussion.md; docs010_multitrack_flows_persona_architecture/022_loop_actions_guide.md; docs010_multitrack_flows_persona_architecture/023_loop_explanation.md
- **relations:** sequential page generation; work-item loops; orchestration state persistence.
- **verify-later:** loop_action.go, loop_expansion_handler.go, loop_complete_action.go in platform.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Data-flow verification matrix practice
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs015/001 is a complete per-step verification table ("Config | Value | Verified ✓") including output structures and registration checklist; docs017/044 repeats the practice for the site-work-orchestrator.
- **what:** A documentation/QA practice: before deploying a workflow, trace every step's config paths against the action implementations — where each output lands in collected_data, its structure, and each input's exact path — plus response-header compliance and action-registration checklists. The manual ancestor of automated contract validation.
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** input contracts; workflow validator; ActionInputSpec.
- **verify-later:** n/a (practice, not code).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Specialist agent design doctrine (agents own their domain)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Five versions of the checklist culminating in docs017/008; audited against real code in docs017/044 ("Audit Notes vs 008_checklist..."), including accepted divergences.
- **what:** The core agent-design rulebook: agents are self-contained and independently callable, with dedicated load_* actions gathering their own data; callers pass raw domain identifiers, never derived values ("if changing the child requires updating the caller, you've leaked responsibility"); reuse/patch existing actions before creating new ones; workflows stay declarative (templates/config = intent OK; loops/branching = Go); orchestrator vs agent boundary (what/order vs how); standalone + integrated dual modes; spawn before call; agents reply to the caller's topic; no container config in definitions. v2's interim "use input_fields not explicit paths" rule was replaced by the ActionInputSpec regime in v3.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md; docs017_legacy_agent_rules_images_design_keydocs/007_checklist_for_new_specialist_agent_v4.md#Orchestrator-Boundaries; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Audit-Notes
- **relations:** ActionInputSpec; webdesign-agent (first exemplar, docs017/003); current development-guide doc 001.
- **verify-later:** how closely current agents follow load_* pattern.

<!-- SOURCE: U21_legacy_docs_b.md -->
### ActionInputSpec / ExtractActionInputs standardized extraction
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** docs017/008 full spec ("No more boilerplate — 40+ lines of extraction code per action eliminated"; deprecation warnings for *_field patterns); real bug documented: "the site plan contamination bug — ExtractActionInputs found site_record.content_data via nested lookup... overwriting the hero section with the site plan."
- **what:** Every action declares an ActionInputSpec (Required/Optional/Defaults/Deprecated) and calls one extraction function that tries input_fields, falls back to deprecated *_field keys with warnings, checks nested parents (current_page/site_record/input_data/rerender_pages), validates and defaults. Includes the hazard doctrine — never name optional fields after common nested keys (content_data, domain, status...), prefix when in doubt — and the `?` suffix for optional input_mapping fields (skip silently if source path missing) supporting multi-mode agents. Literal config values must be read directly from config, not through path resolution.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Decision-Standardized-Input-Extraction; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Avoid-Field-Names; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Fixed-during-audit
- **relations:** data-path problem (root cause); input contracts; workflow validation.
- **verify-later:** datahelpers.ActionInputSpec/ExtractActionInputs/RegisterActionInputSpec in platform.

<!-- SOURCE: U22_recent_small_docs.md -->
### Field-path-resolution duplication tech debt
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "the codebase has at least 18 functions that resolve dot-separated field paths ... This is the single biggest code hygiene issue"; datahelpers canonical vs 9+ scattered duplicates enumerated.
- **what:** A recognised code-hygiene problem: ~18 near-duplicate dot-path resolution helpers (resolveFieldPath, ExtractNestedField, GetFieldFromPath, etc.) spread across datahelpers and the actions package, differing subtly in arg order, logging, and `.response` unwrapping. Canonical is `datahelpers.ExtractNestedField`; the standing rule is reuse datahelpers before adding new resolvers. Related recurring bug: Go-template paths need leading dots (`{{.x.y}}`) and input_mappers are compulsory.
- **sources:** docs020.../001_rag_agent_distribution_architecture.md#field-path, docs019_business/004_vet_practice_verifier.sql#go-template-fixes
- **relations:** datahelpers, rag_actions.go helper cleanup
- **verify-later:** datahelpers vs actions-package path resolvers; NullableString/TruncateString/NullableInt in datahelpers

<!-- SOURCE: U23_docs_root_vonc.md -->
### call_agent contract validation vs input_data.spec convention (dual placement; validator patch)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** Mechanism confirmed in code 2026-07-01 (ValidateInputContract checks only top level); PATCH_validate_input_contract.go "WRITTEN, deploy PENDING" (carried-forward state; still backlog item 3 on 2026-07-09).
- **what:** call_agent resolves input_mapping then validates the target's input_contract.required against TOP-LEVEL keys, while handler workflows read spec fields at `input_data.spec.*` (the work-item convention). The two read different places, so component-creator (required: section_type) can be satisfied neither by pure-top-level (empty-context generic generation — the 081 stray) nor pure-nested (contract violation — 082); the working manual shape (083) provides section_type BOTH top-level AND inside spec. build-dispatch-loop's generic mapping flattens no section_type, so the designed work-item path would hit the same violation (predicted, unconfirmed). Framework fix: the validator accepts a required field top-level OR at input_data.spec.X — not per-handler loop mappings, not enshrining the duplication.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~10:10 + #2026-07-01-~12:46 + #2026-07-01-~13:10; docs/PATCH_validate_input_contract.go.txt; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e; docs/083_regenerate_brief-explanation_vonc.sh (header)
- **relations:** manual agent trigger pattern; build-dispatch-loop genericity (002 §414); component regeneration
- **verify-later:** input_mapping.go ValidateInputContract (patched?); a needs_component_regeneration item dispatched through the loop

<!-- SOURCE: U23_docs_root_vonc.md -->
### Manual agent trigger via the generic entry point (spawn+call pattern)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Proven repeatedly: 083 (2026-07-01), 084 ("the dual-placement trigger pattern worked again", 2026-07-06), trigger-asset-renderer and rerender scripts.
- **what:** One-off manual agent runs post a spawn_agent+call_agent message to `system.agent.generic.requests` (kcat with correlation/request/client headers), with input_mapping delivering the payload. Hard-won sub-rules: dual placement of contract-required fields; a QUOTE-FREE description (name attribute values in prose) to survive the kcat/JSON escaping pipeline; JSON embedded literally (no jq dependency); watch via orchestration_states by correlation_id. The numbered trigger-script series (080–085) in scripts/initial_messages/210_vonc_trigger/ is the operational library, including make_085 which sed-copies a proven trigger for a new page (reuse-first).
- **sources:** docs/084_create_provocations-archive-list_vonc.sh (header); docs/make_085_rerender_provocations.sh; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6 + #§8; docs/summary.txt (kcat basics)
- **relations:** call_agent contract validation; work-item conventions
- **verify-later:** scripts/initial_messages/210_vonc_trigger/ contents

<!-- SOURCE: U23_docs_root_vonc.md -->
### Work-item conventions and manual spec shapes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01 (spec jsonb, item_key dedup via idx_swi_dedup, handler_agent, status flow, pipeline default 'build'); manual rerender/needs_page recipes proven repeatedly.
- **what:** site_work_items is the unit of work: `spec` jsonb (not spec_data at this layer), `item_key` (dedup), `handler_agent`, status detected→triaged→claimed→complete (dispatch picks up triaged/approved), `pipeline`. Manual page items require the FULL spec — page_id (real UUID inline), domain, filename, page_name; placeholder strings get claimed and fail ("invalid UUID length: 18"), and fixing them must filter on the PLACEHOLDER string, not the intended value (the wrong-WHERE no-op lesson). Duplicate insertions are cleaned by grouping on spec->>'page_name' and deleting the older of each pair. Fresh gen_random_uuid item_keys make re-fires safe.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-manual-rerender + #duplicate-work-item-cleanup; docs/RUNNING_NOTES_vonc(36).md#work-item-fix-2026-06-24 + #2026-07-01-~13:40; docs/RUNBOOK_vonc_session(1).md#correct-spec-shape
- **relations:** item_key canonicalization; build-dispatch-loop; complete_error family
- **verify-later:** \d site_work_items; idx_swi_dedup definition

<!-- SOURCE: U23_docs_root_vonc.md -->
### item_key canonicalization (workItemKey builder)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 016b Part 3: "CODE PREPARED; NOT APPLIED" — workItemKey(itemType, target) builder in work_items_common.go; apply gated behind Part-2 verification.
- **what:** item_key prefixes drifted from item_type across creators: the adoption creator keyed BOTH needs_content_page and needs_tool_recreation as `needs_page:<name>`, so a tool and a content page of the same name collide on the dedup index and one is silently dropped. Fix: a shared workItemKey builder; the tool item moves to its own prefix while the content item deliberately keeps `needs_page:` co-dedup with planner builds (Option B, decision recorded).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 3)
- **relations:** work-item conventions; dedup index
- **verify-later:** work_items_common.go workItemKey applied?; adoption creator key prefixes

<!-- SOURCE: U23_docs_root_vonc.md -->
### Standing engineering rules (the session working constitution)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Restated as "Standing rules (brief)" in the v2 carried-forward state and as "Standing instruction from the user, in force" in the HANDOFF.
- **what:** The recurring project rules this corpus operates under: schema-before-SQL (\d first); reuse/alter before create (STEP ZERO); structural over quick fixes; workflows THIN with logic in Go actions; no sub-workflows in SQL — spawn sub-agents; every agent is an orchestrator; agents respond to the CALLER's responses topic; no logger.Debug (invisible in cluster logs); British English; flag variable/signature changes; never treat 0 rows as decisive; verify against deployed artifacts not pod logs; no summary docs unless asked; work in reasonable step sizes.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#standing-rules; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§1; docs/HANDOFF_vonc_write_site_spec_spec_data.md#standing-rules
- **relations:** debugging doctrine; development-guide (001) anchors
- **verify-later:** n/a (convention; verify against 001/002/003 docs)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Basic operations reference (kcat spawn, scale, monitoring)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** summary.txt is a concatenation of basic_usage docs actively describing the current operating procedure (spawn_group website-builder message shape, headers, monitoring queries).
- **what:** The operator's basic-usage layer: scale the deployment set up/down (agent-chassis, auth-service, content-creator-agent, core-manager, image-generator-adapter, reasoning-agent, web-search-adapter); post spawn_group/orchestrate messages via kcat from a test pod to the cross-namespace Kafka bootstrap with required headers (correlation_id, request_id, client_id, agent_instance_id, fuel_budget); monitor via orchestrator_state/orchestration_states by correlation_id. The fuel_budget header and the fixed header set are part of the platform's message contract.
- **sources:** docs/summary.txt; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6
- **relations:** manual agent trigger pattern; system-architecture (topics)
- **verify-later:** docs/basic_usage originals; current deployment list

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Guidelines audit (001/002/003 compliance)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Read the dev guide, architecture, and contracts. Existing code: no violations"; 2026-06-13(d) action "audited against 001/002/003 — code is COMPLIANT".
- **what:** Recurring audits confirming the engine and collector honour the house rules: engine is standalone package main; the no-JS HTML form satisfies JS Content Separation; parameterised SQL only; no logger.Debug; kebab-case/snake_case names; private same-file helpers are allowed; stats_key never logged.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_running_notes(27).md#2026-06-13-d
- **relations:** produced the wrapper-orchestrator finding and the envelope-contract flag
- **verify-later:** 001 dev guide; 002 architecture; 003 contracts

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### workflow field-path audit query (jsonb_path_query over agent_definitions)
- **category:** development-guide
- **status-signal:** unknown
- **status-evidence:** A single ad-hoc diagnostic query with no surrounding narrative or dated claim of use.
- **what:** A small but reusable diagnostic technique: a recursive `jsonb_path_query('$.**.<key>')` sweep over every `agent_definitions.default_config->'workflow'->'steps'` row to extract every field-path value referenced by a fixed set of workflow keys (`agent_type_field`, `default_from`, `content_field`, `iterate_over`, and any `*_from`/`*_field` wildcard key) across the whole workflow corpus at once — a way to audit real field-path usage in stored workflow JSON without opening each agent definition individually.
- **sources:** docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v2/001_validator_sql.sql
- **relations:** development-guide (001 anchor, "grep before using" / field-path resolution canonical-functions guidance)
- **verify-later:** none — a query technique, not a stored artifact

<!-- SOURCE: U26_misc_dirs.md -->
### Workflow-as-configuration (JSON workflows in agent definitions)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 002-agent-chassis-docs.md gives the canonical `{"start_step": ..., "steps": {...}}` shape; HITL definition (Nov 2025) still uses exactly this workflow JSON structure with `next_step` chaining.
- **what:** Agent behaviour is a JSON workflow (start_step + named steps, each with an action, config, and next_step) stored in agent_definitions.default_config / task_workflow, overridable per agent_instances. Contrasted with Temporal/Airflow where workflows are compiled code — here business users can create workflows without deployment.
- **sources:** docs/architecture/002-agent-chassis-docs.md#how-workflows-work; docs/humanintheloop/hitl_agent_definition.sql; docs/architecture/012-investors.md#dynamic-workflow-creation
- **relations:** agent chassis; execute_llm_prompt action; local vs remote actions
- **verify-later:** agent_definitions.task_workflow / orchestrator_workflow columns; workflow validator code

<!-- SOURCE: U26_misc_dirs.md -->
### execute_llm_prompt generic action with DB prompt templates
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Planned in basic_usage/003 ("the reusable 'chef' that cooks the 'recipes'"); in live use by Nov 2025 — hitl_agent_definition.sql's workflow uses `"action": "execute_llm_prompt"` with a Go-template prompt_template.
- **what:** A single generic action that reads the agent's prompt_template and ai_service config (provider, model, api_key_env_var) from its definition, renders the template with Go text/template placeholders ({{.input_data.field}}) filled from collected workflow data, calls the configured LLM, and returns the text. Makes every LLM agent a pure data configuration.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement#step-1.2; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; dynamic prompt improvement loop
- **verify-later:** platform/orchestration/actions/ai_actions.go; prompt template rendering

<!-- SOURCE: U21_legacy_docs_b.md -->
### Workflow Builder & Validator (YAML DSL)
- **category:** NEW:workflow-authoring
- **status-signal:** abandoned
- **status-evidence:** docs006/001 full design with roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool; workflows continued to be hand-written SQL.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. CLI (`workflow-builder build/validate/test/list/show/docs`), planned HTTP API, web UI, and git-based CI/CD workflow deployment.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem; workflow validator tool (docs017/002_standardising); superseded in spirit by input_mapping/ActionInputSpec conventions.
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Workflow Builder & Validator (YAML DSL)
- **category:** NEW:workflow-authoring
- **status-signal:** abandoned
- **status-evidence:** docs006/001 full design with roadmap claiming "[x] Phase 1: Core parser & validator, [x] Phase 2: Path resolution, [x] Phase 3: JSON generation"; no later doc references the tool; workflows continued to be hand-written SQL.
- **what:** A validation-first system for authoring orchestration workflows in human-readable YAML instead of raw JSON: parses a DSL, validates agent types exist in agent_definitions, detects circular dependencies and invalid input references, auto-computes CollectedData paths (agent call vs local action nesting), generates the orchestration_workflow JSON, test cases, and docs, then inserts into the DB. CLI (`workflow-builder build/validate/test/list/show/docs`), planned HTTP API, web UI, and git-based CI/CD workflow deployment.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#Architecture; docs006_workflow_builder/001_workflow_builder.md#Path-Resolution; docs006_workflow_builder/001_workflow_builder.md#Roadmap
- **relations:** data-path resolution problem; workflow validator tool (docs017/002_standardising); superseded in spirit by input_mapping/ActionInputSpec conventions.
- **verify-later:** platform/workflowbuilder/ directory existence in repo history; any workflow YAML files.

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Mega-prompt fragility and candidate replacement patterns
- **category:** NEW:prompt-composition
- **status-signal:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS, ~2026-05)
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas; six fragility concerns (untraceable failures, monotonic rule growth, coupled component vocabulary, one blend ratio, model coupling, token waste ~160MB/build-cycle at scale). Five candidate patterns: per-component templates; structured intermediate envelope (cheap-model stage 1 → focused stage 2, cacheable, lockable); tool-calling for schema; validation-instead-of-prompt-rules; hybrid baseline+overrides. Envelope (B) flagged strongest for both text and images.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (pattern D partially exists); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Mega-prompt fragility and candidate replacement patterns
- **category:** NEW:prompt-composition
- **status-signal:** aspirational
- **status-evidence:** "Treat the existing 6KB text prompt as technical debt, not a model … Not blocking imagery or anything else; just don't propagate the pattern" (undated FOCUS, ~2026-05)
- **what:** page-content-writer's single ~6KB prompt blends 11+ inputs, 16 growing STRICT RULES, and six worked output schemas; six fragility concerns (untraceable failures, monotonic rule growth, coupled component vocabulary, one blend ratio, model coupling, token waste ~160MB/build-cycle at scale). Five candidate patterns: per-component templates; structured intermediate envelope (cheap-model stage 1 → focused stage 2, cacheable, lockable); tool-calling for schema; validation-instead-of-prompt-rules; hybrid baseline+overrides. Envelope (B) flagged strongest for both text and images.
- **sources:** FOCUS_prompt_composition_pattern.md (whole)
- **relations:** image parameter shaping; validate_page_content (pattern D partially exists); LLM reliability strategy
- **verify-later:** page-content-writer default_config prompt size/shape today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Language handling: implicit mechanism plus minimal explicit prompt support
- **category:** NEW:language-i18n
- **status-signal:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS)
- **what:** Content language rides implicitly on the brief/specs/existing-content context; Step 3 made the page-content-writer prompt language-agnostic (## Language section, de-Anglicised rule examples, translate-the-intent note for English llm_guidance, any-language placeholder rule). Mapped remaining English-hardcoded surfaces: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, missing <html lang>. Deferred designs: sites.primary_language column (add when a consumer exists), explicit target-language parameter, "soft static" LLM override of Tier B labels, adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification (fallback problem); mega-prompt concerns
- **verify-later:** page-content-writer prompt ## Language section; head template lang attribute

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Language handling: implicit mechanism plus minimal explicit prompt support
- **category:** NEW:language-i18n
- **status-signal:** partial
- **status-evidence:** "After Step 3 the page-content-writer prompt has only one explicit language signal — a ## Language section"; "There is no language field on sites, pages, content_components, or site_specs today" (undated FOCUS)
- **what:** Content language rides implicitly on the brief/specs/existing-content context; Step 3 made the page-content-writer prompt language-agnostic (## Language section, de-Anglicised rule examples, translate-the-intent note for English llm_guidance, any-language placeholder rule). Mapped remaining English-hardcoded surfaces: Tier B static fallbacks, admin briefs, strategist internal text, other agents' prompts, missing <html lang>. Deferred designs: sites.primary_language column (add when a consumer exists), explicit target-language parameter, "soft static" LLM override of Tier B labels, adoption-time language detection.
- **sources:** FOCUS_language.md (whole)
- **relations:** tiered field classification (fallback problem); mega-prompt concerns
- **verify-later:** page-content-writer prompt ## Language section; head template lang attribute
