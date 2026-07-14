# EXTRACTION U01 — docs024_key_docs_latest numbered core docs (000–106, P1–P3, root .sql/.patch)
Extracted 2026-07-13. Files in scope: 68. Concepts found: 154 (rolled-up entries; many carry multiple sub-facts).

## Coverage
| file | treatment |
|---|---|
| 000_documentation_index(2).md | family-latest |
| 000_documentation_index.md | family-delta (only diff: (2) adds the 016b row; nothing dropped) |
| 001_development_guide(4).md | family-delta (nothing dropped vs (5)) |
| 001_development_guide(5).md | family-latest |
| 002_system_architecture.md | family-delta (dropped: `page-rerender` specialist row — reworded in later versions) |
| 002_system_architecture(2).md | family-delta (dropped: one page-build-handler description line) |
| 002_system_architecture(3).md | family-delta (dropped in (4): css_themes theme-library-growth loop; webdesign→brand-designer/layout-architect/style-generator split — superseded by composition/execution split) |
| 002_system_architecture(4).md | family-latest |
| 003_contracts_and_standards.md | family-delta (nothing dropped) |
| 003_contracts_and_standards(6).md | family-delta (nothing dropped) |
| 003_contracts_and_standards(7).md | family-delta (dropped in (8): literal-colour dark-section contract — superseded by section painting contract) |
| 003_contracts_and_standards_7_.md | family-delta (byte-identical to (7)) |
| 003_contracts_and_standards(8).md | family-latest |
| 004_improvement_loop.md | full |
| 005_tool_pipeline.md | family-delta (only diff: (1) adds the de-tool hazard) |
| 005_tool_pipeline(1).md | family-latest |
| 005_tool_pipeline_1_.md | family-delta (byte-identical to (1)) |
| 006_news_feed_pipeline_v2.md | full |
| 007_adoption_pipeline_v4.md | full |
| 007_adoption_pipeline_v4.patch | full (doc patch adding plan-domain/site_plan_directives updates) |
| 008_vet_med_pricing_pipeline.md | full |
| 009_model_infrastructure.md | full |
| 010_scheduler_and_tasks.md | full |
| 011_database_and_infrastructure.md | full |
| 012_admin_dashboard.md | full |
| 013_content_governance.md | full |
| 014_site_snapshots_and_revert.md | full |
| 015_batch_processing_architecture_v2.md | full |
| 016_additions_assumed_helper_and_cross_module.md | full |
| 016_debugging_guide_v2_58_consolidated.md | full (tail sections via heading-verified scan) |
| 016b_debugging_guide_8_consolidated.md | family-delta (only diff: consolidation note; no bug info dropped) |
| 016b_debugging_guide_merged(1).md | family-latest |
| 017_companies_house_enrichment.md | full (cascade section heading-verified) |
| 019_tool_library(2).md | full |
| 020_tool_lifecycle(2).md | full |
| 021_model_swap_and_rollback.sql | header-scan (migration 083: snapshot_agent/swap_agent_model/revert_agent/agent_snapshots view) |
| 021_site_spec_and_classifier.md | full |
| 022_ai_endpoint_health_and_flywheel_llm_call_log.sql | header-scan (migration 085: ai_endpoint_health + llm_call_log flywheel columns) |
| 022_dynamic_applications.md | full |
| 023_llm_quality_testing.md | full |
| 024_link_management_v2.md | full |
| 025_palette_layout_typography_migration(3).md | full |
| 026_component_regeneration_flow(1).md | family-delta (nothing dropped vs (2)) |
| 026_component_regeneration_flow(2).md | family-latest |
| 027_design_and_site_planner_v2.md | full |
| 028_platform_mission_and_pipeline_direction(2).md | full |
| 029_site_plan_and_reconciler(2).md | full |
| 030_phase1_plan_and_reconciler(5).md | full |
| 031_LOCKS_should_locks_expire.md | full |
| 031_locks.md | family-delta (dropped in (3): Pattern-B-as-live wording — (3) marks Pattern B dead) |
| 031_locks(3).md | family-latest |
| 031_locks_proposed_update.md | family-delta (intermediate revision of the same doc; no unique concepts) |
| 032_storage_architecture_and_credentials.md | full |
| 033_thunder_adapter_design(1).md | full |
| 034_github_action.md | full |
| 035_adapter_guide.md | full |
| 036_REFERENCE_styling_render_pipeline.md | full |
| 037_TOOL_DOCS_convention(1).md | full |
| 101_switch_to_haiku.sql | header-scan (077: bulk model downgrade to haiku with RESTORE section; records per-agent model state 2026-04-10) |
| 103_blog_nav_handoff-2026-04-12.md | full |
| 105_dispatch-pipeline-failures-report_v4.md | full |
| 106_claude_anthropic_skill.md | full |
| P1_build_expand_plan.md | full |
| P2_public_api_plan.md | full |
| P3_admin_api_plan.md | header-scan (full heading inventory: dual-auth gateway, endpoint inventory, known bugs/mock data, fix blocks A–F) |

## Concepts

### Pre-flight "does this already exist?" discipline (Step Zero)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) opens with it as "the most important step", with the asset-deploy-agent 3-hours-wasted example
- **what:** Before creating any agent/action, search agent_definitions, the action registry, Go code, gate functions, and workflows for existing equivalents; document findings; never create without demonstrating no existing coverage. Extends to documentation claims: verify at point of use and date what you verify (`[checked YYYY-MM-DD]`).
- **sources:** 001_development_guide(5).md#Pre-Flight, #API Verification Reference, #Reuse Before Creating
- **relations:** canonical field-path helpers; assumed-helper build failures (016 additions)
- **verify-later:** platform/orchestration/actions/registry.go; agent_definitions table

### Canonical field-path resolution helpers (datahelpers) vs 18+ duplicates
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): "18+ functions that resolve dot-paths… Do not add another one"; cleanup of 9+ duplicates listed under "Not yet built"
- **what:** `ExtractNestedField/String/Map`, `GetFieldFromPath(WithDefault)` in datahelpers are canonical (with `.response` auto-unwrap); six named duplicates in the actions package must not be copied. There is no `ExtractStringSlice` — compose `ExtractStringListHelper(ExtractNestedField(...))`.
- **sources:** 001_development_guide(5).md#Field Path Resolution; 016_additions_assumed_helper_and_cross_module.md
- **relations:** assumption checklist item on assumed helpers
- **verify-later:** platform/orchestration/datahelpers/*.go

### Actions are the unit of work — no wrapper+core split
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "The wrong pattern" section with WriteSiteSpec example
- **what:** All action logic lives inside the `XxxAction` function; composition happens via workflows, not Go-calling-Go; exporting a "core logic" function creates a duplicate API surface. Also: don't create subworkflows in SQL — spawn sub-agents.
- **sources:** 001_development_guide(5).md#Core Design Principles
- **relations:** every-agent-is-an-orchestrator
- **verify-later:** grep exported non-Action functions in actions package

### spawn→call pattern and target_role lookup
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5): "This is how every existing workflow does it"
- **what:** Agents are spawned (`spawn_agent`) then called (`call_agent`). `target_role` → findAgentByRole scans all collected_data keys (preferred); `agent_type` → findAgentByType only scans keys starting `spawn_` (a trap). Dynamic dispatch = fixed role + `agent_type_field` resolved from collected_data at runtime; no topic-construction bypass.
- **sources:** 001_development_guide(5).md#How call_agent finds the spawned agent, #Dynamic dispatch; 002(4)#Resolved Decisions 16
- **relations:** dispatch loop; wrapper-orchestrator pattern
- **verify-later:** spawn_actions.go, call_agent.go

### Wrapper-orchestrator pattern ("every pod-running agent needs a parent that spawned it")
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) full section; med-* wrappers and site-adoption-orchestrator live per 002(4)
- **what:** Agents get dedicated K8s Job pods only via spawn_agent from a parent. Anything reached via the generic entry point that does substantive work needs a tiny wrapper (spawn → call → complete) so real work runs in its own pod with clean logs; in-chassis work blocks shared pod slots. Canonical minimal wrapper: med-export-orchestrator. Map input fields individually (never `input_data: input_data`), mark caller-optional fields `?`.
- **sources:** 001_development_guide(5).md#Every pod-running agent needs a parent; 002(4)#Active agents note
- **relations:** topics model; site-adoption-orchestrator
- **verify-later:** agent_definitions rows for med-* orchestrators; spawnAgentKubernetesJobFromDefinition

### Kafka topics model: generic entry point vs per-spawn job topics vs fixed adapter topics
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) topics section describes current mechanics incl. stable identity naming
- **what:** Three patterns: `system.agent.generic.requests` (shared entry door for callers with no spawn relationship — scheduler, manual triggers; explicitly "the current door", expected to evolve into a formalised entry API); `job.<corr8>-<orch8>-<type>-<step>.requests` per-spawn topics set by setupAgentTopics; fixed `system.agent.<type>.*` topics for long-lived adapters. agent_definitions.topics jsonb is declarative only — the Deployment manifest actually subscribes.
- **sources:** 001_development_guide(5).md#Topics; 002(4)#Infrastructure
- **relations:** wrapper-orchestrator; idle timeout; topic cleanup
- **verify-later:** setupAgentTopics/createTopics in spawn_actions.go; deployment manifests

### Standardized input extraction (input_mapping / input_fields / ActionInputSpec) and the `?` optional suffix
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) three-layer table; resolver behaviour documented from code
- **what:** Caller maps data via input_mapping (dot-paths into collected_data), actions declare fields, Go extracts via ExtractActionInputs. `?` suffix on the destination key makes a mapping optional (silently skipped); unsuffixed fields hard-fail the call. In the dispatch loop, only site_id/domain/work_item_id may be non-optional; all spec.* mappings must use `?`.
- **sources:** 001_development_guide(5).md#Standardized Input Extraction, #Optional fields in dispatch loop
- **relations:** field name collisions; handler input path contract
- **verify-later:** ResolveInputMapping in coordinator.go

### Field-name collision via the nested-source loop
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) corrected wording: affects required AND optional fields; real section-editor content_data clobber
- **what:** ExtractActionInputs's late nested-source loop checks `current_page`, `rerender_pages`, `site_record`, `input_data` for any unresolved field name — so generic names (`content_data`, `sections`, `site_id`, `domain`, `status`…) can silently bind to the wrong source. Rules: new code avoids colliding names (prefix them); existing code left alone unless it bites; complex/array fields must never go in ActionInputSpec (read the config path directly).
- **sources:** 001_development_guide(5).md#Field name collisions; 016 §0 item 15 and §9 literal-key trap
- **relations:** section-editor clobber; resolve_internal_links review catch
- **verify-later:** datahelpers/action_inputs.go nestedSources

### Agent message structure and HITL response shape
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5) reference section
- **what:** Messages have Kafka headers (correlation/orchestration/request ids, message_type, action, responses_topic, sender) and a body of `headers`/`config`/`input_data`. Agents always reply to the caller's responses_topic. HITL responses go to the agent's responses topic with `in_response_to_request_id` from awaited_requests and `sender_agent_type: human`.
- **sources:** 001_development_guide(5).md#Agent Message Structure
- **relations:** adapter response envelope contract; awaited_requests
- **verify-later:** MessageProcessor, awaited_requests table

### Orchestration state and collected_data as the workflow data bag
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5)/016 §6.0 describe live mechanics
- **what:** Each orchestration is an orchestration_states row (workflow_plan, collected_data, current_step, status). Steps communicate solely via dotted paths into collected_data; agents themselves are DB rows (`agent_definitions.default_config.workflow`), not Go types — the Go codebase contains actions, not agents. "Every agent is an orchestrator" is literal.
- **sources:** 001_development_guide(5).md#Orchestration State; 016 §6.0 "What an agent actually is"
- **relations:** loop mechanisms; workflow result contract
- **verify-later:** orchestration_states schema; coordinator.go continueExecution

### Stale orchestration sweeper
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 001(5): "the #1 cause of pipeline stalls"; design with A/B/C classification
- **what:** Timeout goroutines die with pods, leaving AWAITING_RESPONSES orchestrations stuck. A periodic 60s DB sweep on chassis pods (FOR UPDATE SKIP LOCKED) classifies expired awaited requests: child completed → synthesize response; child failed → forward; none/running → retry up to 3 then fail parent. Handles topic-expired case by directly advancing parent state.
- **sources:** 001_development_guide(5).md#Stale Orchestration Sweeper
- **relations:** spawn-handler hang (timeout_at not enforced); claimed-item-timeout
- **verify-later:** sweeper implementation; stale-orchestration-reaper scheduled task

### Model aliases and the model selection strategy
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(5) rules; 009 per-step table; 101_switch_to_haiku.sql records live state 2026-04-10
- **what:** Agent definitions use short aliases (claude-sonnet-4-6, claude-haiku-4-5) resolved by model_aliases.go; sonnet is the default for LLM steps, haiku for routing, opus for chief-strategist/planner, ollama for fine-tuned classification. 101 SQL is a bulk cost lever switching all agents to haiku with a RESTORE section.
- **sources:** 001(5)#LLM Infrastructure; 009#Model Swap; 101_switch_to_haiku.sql
- **relations:** swap_agent_model; LLM tiering (029)
- **verify-later:** model_aliases.go; per-step ai_service in agent_definitions

### LLM call logging (llm_call_log) as ops visibility + training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 001(5): "Verified in production (March 2026) — 57+ rows"; 085 migration adds flywheel columns
- **what:** Every execute_llm_prompt call logged fire-and-forget (agent_type, step, model, rendered prompt, response, tokens, latency, `__sent_temperature`/`__sent_max_tokens` write-backs). Flywheel columns (work_item_id, prompt_variant, vertical, rag_context_used) link calls to outcomes for LoRA/RAG training exports. Known past bugs: schema/Go column drift (agent_id vs client_id), empty agent_type from buildActionParams.
- **sources:** 001(5)#LLM call logging, #Implementation Status; 022_ai_endpoint_health_and_flywheel_llm_call_log.sql
- **relations:** batch queue LogLLMCall paths; fine-tuning path
- **verify-later:** llm_call_log schema; llm_call_logger.go

### Ollama adapter (CPU embeddings + local classification)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009: "Ollama adapter on CPU cluster (2 replicas, mistral-small3.1 + nomic-embed-text)" checked done
- **what:** Permanent CPU adapter serving nomic-embed-text embeddings (~50-100ms) and quantized small models for classification (10-30s acceptable per-build). Same AIService interface as Anthropic incl. token-usage write-backs. Not for content generation or <2s latency.
- **sources:** 001(5)#Ollama adapter; 009#Implementation Status
- **relations:** RAG actions; endpoint health; fine-tuning path
- **verify-later:** ollama.go; ollama-adapter deployment

### RAG actions and knowledge_base shared store
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 001(5): migration 082 applied "Not yet populated"; rag actions Go "produced but not deployed" then 009 lists "registered, not workflow-tested"
- **what:** `rag_lookup` (embed→vector search→top-k, trigram fallback) and `rag_index` (chunk→embed→store, SHA256 dedup) over a shared `knowledge_base` table (vector(768)); deliberately actions not agents until a knowledge-indexer orchestration is needed. Tool docs also target a knowledge_base `tool_docs` row.
- **sources:** 001(5)#RAG actions, #Agent vs infrastructure; 037#Where the docs live
- **relations:** agent-vs-infrastructure test; per-tool docs convention
- **verify-later:** rag_actions.go registered?; knowledge_base row counts

### Fine-tuning path (log → export → LoRA → GGUF → Ollama → swap)
- **category:** finetuning-flywheel
- **status-signal:** partial
- **status-evidence:** 009 (2026-07-10 update): one training run `complete`, GPU provisioning via Thunder real and dynamic, but "No agent_definitions row currently points ai_service at llama3.3:70b — trained and tested, never used for production inference"
- **what:** Pipeline: accumulate 200+ examples in llm_call_log → export (Alpaca/ChatML) → LoRA fine-tune on GPU (unsloth) → GGUF → Ollama → swap agent definition → A/B against Claude. Candidates are short-output classifiers (site-classifier, vet-practice-verifier, etc.). The last mile (wiring the trained 70B into live inference) is explicitly outstanding.
- **sources:** 001(5)#Fine-tuning path; 009#Future incl. 2026-07-10 note; 023 (manual A/B comparisons)
- **relations:** model swap functions; Thunder adapter; drain mode
- **verify-later:** model_lifecycle.training_runs; agent_definitions ai_service providers

### Agent vs infrastructure boundary test
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) table (LLM logger no, Ollama provider no, rag actions no, knowledge-indexer future yes)
- **what:** Something becomes an agent only if it owns a domain, needs its own workflow, and benefits from independent spawn/debug. Otherwise it is an action or cross-cutting infrastructure.
- **sources:** 001_development_guide(5).md#Agent vs infrastructure
- **relations:** promotion pattern (002d)
- **verify-later:** —

### Specialist vs handler: the persistence boundary
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) post-mortem; hit twice (page-content-writer HTML trapped in site_work_items.result)
- **what:** A specialist returns data to its caller; a dispatch handler must persist its own outputs (page_components, site_components, assets) and update status. Specialists used as handlers need a wrapper (page-build-handler wraps page-content-writer with plan/validate/save/deploy). Test: callable from CLI with site_id+domain and everything lands in the right tables.
- **sources:** 001(5)#Lessons Learned; 002(4)#Page Build Handler Pipeline
- **relations:** dispatch loop; handler contract
- **verify-later:** page-build-handler definition; handler agents' save steps

### `pipeline` column as work-item routing namespace
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5): renamed from `domain` (clash with site domain); dispatch loop filters pipeline='build'
- **what:** site_work_items.pipeline ('build','design','maintenance'…) is the dispatch routing namespace, never the website domain. Everything in the initial build must be pipeline='build'; items emitted straight-to-triaged must set it at emission (triage rewrites it for detected items). Historical trap: dispatch once passed the namespace to handlers as the site domain.
- **sources:** 001(5)#Work item pipeline must be "build"; 016 Schema reminders
- **relations:** dispatch loop; schema-drift rule
- **verify-later:** find_dispatchable_site / LoadWorkItemsAction filters

### Terminal work items: every pipeline ends with assembly + deployment
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every pipeline must end with assembly; 004 step 9 (needs_rerender priority 99)
- **relations:** commit-is-deploy; rerender agents
- **verify-later:** WriteBuildItemsAction

### Extended thinking config and the no-temperature-to-Anthropic rule
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) dated 2026-05-27 temperature note
- **what:** `budget_tokens` in ai_service enables extended thinking (thinking blocks skipped in parsing, +30-90s). Since 2026-05-27 the Anthropic client sends no temperature at all (Opus 4.7+ 400s on non-default; thinking incompatible); Ollama still honours it. Steer Anthropic via budget_tokens and prompts.
- **sources:** 001(5)#Extended Thinking Configuration
- **relations:** LLM config shadowing (temperature dead paths)
- **verify-later:** anthropic.go client options

### Work-item blocking/unblocking and the `unresolved` two-strike mechanism
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 006 "Two-strike dedup ✅ Deployed"; 001(5) describes insertWorkItem mechanics
- **what:** Three block causes (missing handler → feasibility-recheck auto-promotes; spec blocked → manual; manual). insertWorkItem suppresses re-detection within 3h of a terminal duplicate and, after 2+ terminal attempts in 7 days, creates new items as status `unresolved` — visible, not dispatched. `wont_fix` + "superseded by active duplicate" is the dedup system working, not a bug.
- **sources:** 001(5)#Work Item Lifecycle; 006 issues table (12k duplicate cleanup); 016 §9 wont_fix entry
- **relations:** idx_swi_dedup; feasibility-recheck
- **verify-later:** load_work_item_actions.go insertWorkItem

### Loop mechanisms: dynamic workflow expansion
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) Appendix C full reference incl. production dispatch loop example
- **what:** Loops inject N×M steps into the workflow plan at runtime (`{loop}_iter_{N}_{substep}`), with setLoopVariable propagating iteration outputs to base names, per-iteration output suffixing, continue_on_error skipping, and LoopCompleteAction aggregation. Known hazards: the fast-response race fixed by the ErrLoopExpansionHandled sentinel; shared `loop_metadata` key; never nest loops — spawn a sub-agent instead.
- **sources:** 001(5)#Appendix C — Loop Mechanisms
- **relations:** dispatch loop pattern; O(K²) state-bloat failure (016)
- **verify-later:** loop_actions.go, loop_expansion_handler.go

### Domain submission tiers and mission/roadmap briefs
- **category:** onboarding-config
- **status-signal:** deployed
- **status-evidence:** 001(5) documents current domain-submitter workflow with persist steps
- **what:** domain-submitter is the entry point for new builds. Three tiers: domain only; domain+objective hint; domain+mission/roadmap (structured JSON for machine consumers + plain-text `mission_brief`/`roadmap_brief` that classifier/planner actually read). Persist steps skip gracefully via error_step when fields absent; briefs must be plain text parseable by small models.
- **sources:** 001(5)#Domain Submission; 007#Mission-Driven Sites
- **relations:** classifier weighting of inputs (028); vonc/Spark pattern
- **verify-later:** domain-submitter agent definition; site_specs mission aspects

### Authoring rules pack (schema-check, parameterised SQL, api_key_env_var, nil-guarded templates, code-fence stripping, error_step-in-config, text wrapped for write_site_spec)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** 001(5) "Summary of rules" 1–20, each backed by a dated bug
- **what:** The distilled 20-rule authoring discipline from the bug tally: `\d` live schema before SQL (dumps go stale; domain→pipeline, version_note→change_description); $1+params never {{.field}} in SQL; every LLM step needs api_key_env_var; {{if}} before {{range}} (query_database empty = null not []); run agent-def SQL before triggering (chassis silently runs empty workflow); strip markdown code fences before JSON parse; error_step only works inside step.Config; write_site_spec rejects scalars (wrap {"text": …}); to_jsonb('…'::text); verify fire-and-forget INSERTs actually land.
- **sources:** 001(5)#Appendix A + #Summary of rules
- **relations:** debugging assumption checklist; best-effort-needs-monitoring
- **verify-later:** —

### Design system three layers (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) opening section; live tables
- **what:** Layer 1 HTML components (self-contained blocks, inline style, CSS variables with fallbacks, never hardcode brand colours); Layer 2 CSS theme (one styles.css per site rendered from installed composition); Layer 3 style_collections bundling header/footer components + theme + palette/typography. Sites reference via sites.style_collection_id.
- **sources:** 002(4)#Design System Layers; 003(8) contracts
- **relations:** palette/layout/typography migration (025); composition (027)
- **verify-later:** content_components, css_themes, style_collections schemas

### Composition three stages: direction → composition → execution
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) and 027 describe the deployed reorder ("Applied"); install_theme removed from webdesign-agent
- **what:** (1) domain-research-classifier writes design_intent (structured palette/typography reference_values + style_direction scheme). (2) site-design-planner (deterministic, no LLM, `needs_composition` item) resolves layout (weighted scheme-aware tag match), typography (match-or-insert), palette (always new site-specific row) via signal cascades, then install_site_composition atomically writes css_themes+style_collections+sites pointer+resolved_composition spec (a decision record, not CSS; refuses overwrite). (3) webdesign-agent (needs_design, depends_on composition) renders and commits styles.css — sole writer.
- **sources:** 002(4)#Composition; 027 full; 025 (schema underneath)
- **relations:** scheme-aware matcher; renderer cascade; superseding-spec-doesn't-undo-install failure mode (028)
- **verify-later:** fork_theme_composition.go resolvers; install_site_composition; needs_composition/needs_design ordering

### Scheme-aware weighted layout matcher + needs_new_layout_candidate HITL signal
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4); 016 §9 v2_55 records its half-merged-deploy failure and fix shipped as one merged file
- **what:** Layout matching weights tags by rarity with category/description bonuses; the site's scheme (from design_intent.style_direction) is a near-hard constraint (light site never placed on dark layout while any non-dark fits). On fallback it queues `needs_new_layout_candidate` (status needs_human_review, skipped by dispatch) — the honest "library is missing a layout" signal. layouts.scheme nullable → incremental curation.
- **sources:** 002(4)#Composition; 027 §2; 016 §9 scheme-matcher entry
- **relations:** library growth; section-contrast open question (036)
- **verify-later:** resolveLayoutByTagsWeighted; layouts.scheme population

### Theme/layout library growth and the fork-with-review gate
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 003(8) forking rules deployed; 002(3)'s auto "search→reuse→generate→store" loop dropped from 002(4) (superseded by curated-library stance); LLM layout matching "a future step"
- **what:** Layouts are a curated shared grammar (no auto-generated bespoke layout per site). Growth: hand-added variants, or HITL route — ForkThemeFromSiteAction promotes a rendered design into css_themes+style_collections with needs_review=true and a needs_theme_review item; selectors must exclude needs_review rows; rejection only affects future sites. Lineage columns (origin/forked_from/source_site/forked_at) on themes and collections.
- **sources:** 002(4)#Library growth; 003(8)#CSS Theme Template Contract (lineage, review gate, forking rules)
- **relations:** 025 migration lineage model on palettes/layouts/typography_sets
- **verify-later:** fork_theme_from_site_action.go; needs_review filtering in selectors

### Superseded: single webdesign-agent brand+CSS role and the brand-designer/layout-architect/style-generator split
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** dropped between 002(3) and 002(4); 002(4): "The earlier 'one agent generates brand + CSS' shape is superseded by the composition/execution split"
- **what:** Earlier architecture had webdesign-agent doing brand analysis + CSS with a planned future split into brand-designer/layout-architect/style-generator, and an auto theme-library reuse loop. Replaced by site-design-planner (composition) + webdesign-agent (render); a finer split deferred until search-and-adapt clearly beats render-from-composition.
- **sources:** 002_system_architecture(3).md (family-delta); 002(4)#Design Agent Family
- **relations:** composition three stages
- **verify-later:** no brand-designer/layout-architect agent_definitions rows expected

### Unified build & maintenance via site_work_items (single queue, same code)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code"
- **what:** Every piece of work is a site_work_items row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified Build and Maintenance, #site_work_items; 003(8)#Cross-Domain Coordination
- **relations:** dispatch loop; work-item state machine (016); P1 expansion sources table
- **verify-later:** site_work_items schema incl. depends_on, item_key

### Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4) routing table dated with 2026-06-22 hazard confirmation
- **what:** Route by item_type never item_key; the distinction is whether copy is regenerated. needs_page/needs_content_page → full LLM rebuild via page-build-handler; page_rerender (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; needs_rerender → batch reassembly; link_resolution_rebuild is INTENDED links-only but runs the full writer (hazard). page_rerender on a NULL-content_data section escalates the page to needs_page (backfills content_data).
- **sources:** 002(4)#Work-item routing; 003(8)#Source of truth principle (two re-render paths)
- **relations:** interactive-page de-tool hazard; content_data source of truth
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

### Interactive-page de-tool hazard (content rebuild silently drops a tool/game)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14); "fix pending" in 002/005/016; 016b v2 update: two-layer save_page_sections fix WRITTEN, un-deployed
- **what:** A tool lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page or link_resolution_rebuild) regenerates from plan_sections and replaces it with generic-text-block; the prose-based content-regression guard doesn't catch markup/JS loss. Fix layers: interactivity-aware save guard + carry-forward of interactive sections in save_page_sections (written), source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender ruled out for CTA re-resolution — it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9 final entry; 016b Part 4
- **relations:** phantom-CTA resolution bug (separate); tool recreation mis-key (Part 3)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population

### site-work-orchestrator ordering and the dispatch pattern
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) workflow and resolved decisions 16–21
- **what:** Unified orchestrator processes pending items → verifies previous → runs due discovery → triages → updates profile. Dispatch = spawn→call per item with raw identifiers; orchestrator never pre-spawns static chains, never derives handler data, never passes work-item awareness to handlers (handlers self-contained, CLI-callable). Status tracking stays in the loop.
- **sources:** 002(4)#The orchestrator, #Dispatch pattern, #Resolved Decisions
- **relations:** build-dispatch-loop; handler contract
- **verify-later:** site-work-orchestrator definition currency vs build-dispatch-loop

### Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decisions 20–21; 010 seed schedules
- **what:** build-pipeline-trigger fires via kafka-scheduler, seeds queue, picks one dispatchable site (skipping sites with claimed items via NOT EXISTS), spawns build-dispatch-loop with await_response:false. Loop claims atomically, processes one item, completes — parallel sites, no batch accumulation, no OOM.
- **sources:** 002(4)#Dispatch Loop and Pipeline Trigger; 004#Entry Points
- **relations:** site-excluded-by-stuck-claim failure; scheduler concurrency groups
- **verify-later:** build-pipeline-trigger pre_query; find_dispatchable_site SQL

### Page-build-handler pipeline with plan_sections triage (Layer 0) and validate_content gate
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 002(4)/002d full pipeline; input-schema v2 in 003(8)
- **what:** ensure_site_record → load_page_record → plan_sections (resolve each section's input_schema sources; triage into ready/deferred(needs_human_review)/skipped; page deploys with whatever is ready) → content writer (only ready sections) → validate_content (algorithmic: placeholders, unrendered templates, cross-site contamination, broken links, hallucinated emails; blockers/errors → needs_human_review) → save_sections → deploy. Quality gates before generation AND after; content writers never fabricate non-llm-sourced data.
- **sources:** 002(4)#Page Build Handler Pipeline; 002d Layer 0; 003(8)#Component Input Schema v2, #Content Validation Contract
- **relations:** input schema v2; needs_section_data; growth budget
- **verify-later:** plan_sections_action.go; validate_page_content.go

### Component input schema v2 (sources, on_missing vocabulary)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full contract; 016b deferral-drop entry shows live enforcement including the default:defer trap
- **what:** content_components.input_schema declares per-field type/source/required/on_missing/fallback/min_items/llm_guidance. Source prefixes: llm, site_specs.*, site_assets.*, pages.*, config.*, renderer, static, query.*. on_missing: use_fallback/skip_field/skip_section/needs_human_review/block. Image fields must be required:false + skip_field + template-gated (imagery is async). Trap: required:true with on_missing skip_field/empty hits the switch default and defers the section.
- **sources:** 003(8)#Component Input Schema v2; 003(8) checklist 6b; 016b#Regenerated content section deferred
- **relations:** plan_sections; queryresolve; imagery async
- **verify-later:** planSection switch in plan_sections_action.go

### Commit IS deploy (git → GitHub Actions → Backblaze B2, Cloudflare DNS-only)
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decision 15; 034 shows the actual workflow
- **what:** Individual commits per work item; GH Actions fires per commit on a self-hosted runner, detects changed root-level domain directories, `b2 sync --delete --skip-newer` each to `b2://portfolio-sites/<domain>`, then purges Cloudflare cache per zone. No separate deploy step. The authoritative workflow lives in gqls/sites/.github/workflows — a stray copy under .git/workflows is a documented trap.
- **sources:** 002(4)#Git commit strategy; 034_github_action.md; 016 §0 item 24
- **relations:** git-adapter; "git committed is not proof of new content"
- **verify-later:** gqls/sites .github workflow; B2 bucket layout

### QA three layers, group auditor agents, and the promotion pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d architecture with live agents (design-audit-agent, visual-design-auditor, content-quality-auditor, site-review-agent)
- **what:** Layer 1 structural checks (algorithmic, every cycle), Layer 2 group LLM audits (shared context, ONE LLM call per group), Layer 3 strategic review. Group agents chosen over per-check agents (context reuse) and over a mini-action registry (every agent is an orchestrator); a check is promoted from action step to spawned agent by changing one workflow line, output_field unchanged. Site type enables audit groups via maintenance_profile.audit.
- **sources:** 002(4)/002d#Quality Assurance; #Promotion Pattern
- **relations:** improvement loop; audit-enforces-not-overrides
- **verify-later:** design-audit-agent workflow; maintenance_profile.audit config

### Audit enforces intent, doesn't override (chain of authority) + propose mode for spec-less sites
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d responsibility boundaries; resolved decision 20
- **what:** classifier decides intent → planner implements → composition installs → webdesign renders → audit checks build vs stated intent and emits work items; it never makes design decisions. Exception: where no classifier output exists, audit may PROPOSE a direction, flagged for HITL. Handlers never know their trigger (build/audit/manual all use the same webdesign-agent).
- **sources:** 002d#Responsibility Boundaries; 004#Human Direction Integration
- **relations:** dream spec; direction spec; 028 silent-override failure mode
- **verify-later:** auditor prompts reference design_intent/direction

### Dream spec / gap analysis / feasibility (one spec with status, not two documents)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked; gap analysis = blocked/planned subset; feasibility-recheck promotes blocked→planned when capability arrives; feasibility annotations prevent impossible work items. Older 002d dream_spec-in-content_data shape superseded by this. Phase 2 of 028 makes it mechanical.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis
- **relations:** fidelity dial; feasibility-recheck task
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns

### Entity data model (state-based lifecycle, news triggers, client-side real-time)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** 002(4): tables site_entities/site_entity_relationships "exist", entity_sources/entity_sync_log "planned"; Phase 3 item
- **what:** Structured data that generates pages (events, performers, venues, ticket tiers) with state-based lifecycle (announced→on_sale→…→historical), setup mode (work items) + discovery mode (scheduled sync), significant state changes triggering news via entity_sources.news_triggers; real-time data (prices, availability) served client-side from a data API, never through the work queue.
- **sources:** 002(4)#Entity Data Agent Family; #Site Type Stress Tests (events/boxing)
- **relations:** news feed pipeline; site API router (007)
- **verify-later:** site_entities usage; any entity-data-agent definition

### Section editor and buildRenderContextFromDB
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 002(4) section; edit paths live in 013
- **what:** Granular edits without the full pipeline: content_edit (field_updates merge or full replacement) and component_swap; every edit updates content_data first then re-renders (edits survive re-renders). buildRenderContextFromDB reconstructs RenderContext purely from DB (site data, style collection, theme, nav, page meta, content_data) — no collected_data needed.
- **sources:** 002(4)#Section Editor; 003(8)#Source of truth principle
- **relations:** content_data source of truth; admin dashboard edit paths
- **verify-later:** section-editor definition; buildRenderContextFromDB

### content_data is the source of truth (rendered_html is derived)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full section incl. the two re-render paths; "HTML patching was rejected as an edit mechanism"
- **what:** Every section stores content_data (structured) + rendered_html (derived). All edits go through content_data; the light path (rerender_page_sections) re-renders from stored content_data ⊕ fresh-resolved fields with no LLM, persisting the merged content_data so rows stay complete render sources; NULL content_data escalates to full rebuild.
- **sources:** 003(8)#Schema Enforcement/#Source of truth; 002(4) page-rerender row
- **relations:** work-item routing; section editor
- **verify-later:** rerender_page_sections persistence semantics

### Maintenance profile per-site configuration
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4) JSON shape; audit config consumed by improvement loop
- **what:** sites.settings.maintenance_profile controls per-domain cadence (content/links/seo/compliance/content_feed/entity), budgets (llm_calls_per_cycle, max_auto_fixes_per_cycle), build tier, and audit group enablement; audit_pass_count also lives here.
- **sources:** 002(4)#Per-Site Configuration; 002d#Site-Type-Specific Audit Configurations
- **relations:** audit pass cap; growth budget (separate: site_specs growth_config)
- **verify-later:** sites.settings shape in production

### Idle timeout for spawned agents + topic cleanup strategy
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 002(4): mechanism, config column, sync.Once shutdown safety, tuning SQL
- **what:** agent_definitions.idle_timeout_seconds → env var → idle-monitor goroutine exits the pod after inactivity (0 = forever for deployments). Topics: EPHEMERAL_TOPICS per-spawn today; agents never clean up topics — a conservative 10-min CronJob deletes topics with no matching pod, Kafka 7-day retention as backstop; a future shared-topics design (pre-created per-type topics, header routing, static group membership) makes cleanup a no-op.
- **sources:** 002(4)#Idle Timeout, #Shared Topic Strategy, #Topic Cleanup Design
- **relations:** pod accumulation debugging (016 §1)
- **verify-later:** idle_timeout_seconds values; topic-cleanup CronJob

### Spec-supersede rollback pattern (and full snapshot revert as its big brother)
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 002(4) rollback steps; 014 documents deployed snapshot system (migration 085)
- **what:** Targeted rollback = flip site_specs is_current to a previous aspect version + create rebuild work items; git history gives per-work-item revert of deployed HTML. Full point-in-time revert = site_snapshots (JSONB capture of site record/specs/pages+components/nav/site_components + git SHA), take_site_snapshot / revert_site_to_snapshot (always takes a pre_revert safety snapshot; does NOT git-revert or touch global content_components/research_results). Admin API + three workflow actions exist; recommended auto-triggers (post-deploy, pre-propagation, nightly) not yet wired.
- **sources:** 002(4)#Site Rollback Pattern; 014 full
- **relations:** component_versions; agent snapshots (different concern)
- **verify-later:** site_snapshots rows in prod; whether post-deploy snapshot step was added

### component_versions population and change_source provenance
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 014 (April 2026 change_source column); 001(5) records two years of silently-lost history pre-fix (version_note→change_description drift)
- **what:** Three best-effort writers: StoreGeneratedComponentAction create (v1) and regen (MAX+1, snapshot BEFORE update), UpdateComponentHTMLAction (tool improvements). change_source records originating work-item source. Unique (component_id, version_number). Lesson: best-effort operations need active monitoring — silent best-effort was "silent no-effort" for two years.
- **sources:** 014#Populating component_versions; 001(5) bug 18 second case; 026(2)
- **relations:** component regeneration flow; snapshots
- **verify-later:** component_versions row counts by changed_by

### Workflow result contract (flatten vs fields vs mapping; output_field/output_fields foot-gun)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) opening contract; fix deployed 2026-06-18 per 016 §9 (resolveResultSpec)
- **what:** A workflow's complete step declares its result via result_from (flatten — field contents become body), multiple_output_fields (nested per key), or result_mapping; deprecated aliases still resolve with a Warn. No key → fallback dump of collected_data, which can breach the ~900k cap. Parents read at `<call_output_field>.response.<key>`; wrong mode = silent null reads. Historically singular `output_field` was silently ignored → stub-with-success (Part 1 bug); the oversize path now returns an actionable error instead of a stub.
- **sources:** 003(8)#Workflow Result Contract; 016 §9 "Child workflow result silently replaced by a stub"
- **relations:** silent-completion family; result_spec.go
- **verify-later:** result_spec.go; remaining deprecated-alias agents

### Component naming contract (kebab function, data-component, uniqueness)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) with live DB constraint chk_function_kebab_case and partial unique index
- **what:** content_components.function is the canonical identifier (kebab, regex-constrained, one active row per function); data-component attribute on the root element must match exactly; page_components.slot_name mirrors function; GetComponentWithFallback (exact→normalized→generic-text-block) is a safety net not to be relied on. Naming patterns for page-specific heroes and header/footer/head variants.
- **sources:** 003(8)#Component Naming Contract
- **relations:** section_type vs function split (007); slot_name↔function mapping hazard
- **verify-later:** chk_function_kebab_case; idx_content_components_unique_active_function

### String-value naming convention (snake for identifiers, kebab for data)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): kebab CHECKs live; "snake_case columns have not yet had explicit constraints added — follow-up"
- **what:** Values used as identifiers in code (map keys, switch cases, dispatch routes) are snake_case (item_type, action names); pure-data values that end up in CSS/URLs/HTML are kebab-case (function, page_type, agent type); single words lowercase. Decision test: is the value ever a Go case/map key?
- **sources:** 003(8)#String-Value Naming Convention
- **relations:** page_type vocabulary; item_key canonicalization
- **verify-later:** snake-case CHECK constraints existence

### page_type vocabulary and "landing, not index"
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8)/016 §6.5: constraint chk_page_type_kebab_case since migration 051; canonical value table
- **what:** Canonical kebab page_types (landing/content/tool/guide/game/blog-post/blog-index/section-index/entity-page/entity-directory/news-index); the homepage's TYPE is landing while its NAME is index (name is storage convention, type is kind-of-page). CanonicalisePage normalises legacy snake inputs one-way. Guides nest at /guides/<slug>/index.html and appear in guide-lists only when typed guide AND active/deployed.
- **sources:** 003(8)#page_type vocabulary; 016 §6.5
- **relations:** CanonicalisePage; adoption slug-mangling
- **verify-later:** pages page_type distribution; constraint present

### JS content separation contract (js_content → /tools/assets/{function}.js)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full flow through separateInlineJS/collectJSAssets/multi-file git commit
- **what:** Component JS is split out of html_template into js_content and served as an asset file; html_template keeps only a `<script src>` reference. separateInlineJS extracts only attribute-less `<script>` tags (by design). js_snippets is a separate table for shared design effects, never component behaviour. Known failure class: pre-extraction rows render as empty shells (016b entry).
- **sources:** 003(8)#JS Content Separation Contract; 016b#Data-driven component shells render empty
- **relations:** tool doc header stripping; empty-shell taxonomy
- **verify-later:** separateInlineJS regex; script-balance validation hardening applied?

### Component creation & regeneration contract (created/regenerated; already_exists removed)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) + 026(2) dated 2026-04-20 describing deployed behaviour
- **what:** StoreGeneratedComponentAction: Layer-1 pre-store validation runs before either branch (rejection never touches the DB); create INSERTs + v1 snapshot; regenerate snapshots old state then UPDATEs in place (UUID preserved, FKs intact), marks dependent page_components pending, and raises one deduped needs_rerender per affected site (item_key component_regen_rerender:<uuid>). Downstream must not assume component_id is new nor create its own rerender items. Regen keying is by the LLM's EMITTED function — a mismatched name INSERTs a stray.
- **sources:** 003(8)#Component Creation & Regeneration Contract; 026(2) full; 016b#Manually invoking an agent (regeneration keying)
- **relations:** component_versions; markPagesForRebuild; system-stats key-mismatch incident
- **verify-later:** store_generated_component_action.go branches

### Site component linkage contract (slot_name↔function; fallback header hazard)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract + discovery check unlinked_site_components
- **what:** Every site_components row must have component_id → content_components; otherwise renderAndStoreSiteComponent falls to a generic function lookup (which cannot match because slot 'header' ≠ function 'header-<variant>') then to hardcoded RenderFallbackHeader (no logo, stacked nav, search icon, dark). Breakers: update_site_defaults not run, NULL collection header id, legacy data. Self-healing check + site-component-linker handler exist.
- **sources:** 003(8)#Site Component Linkage Contract; 004 discovery checks
- **relations:** four overlapping chrome stores (036); light-site-dark-chrome bug
- **verify-later:** update_site_defaults in workflows; unlinked check registration

### CSS colour inheritance model (--section-* with fallbacks)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8): "the single most important rule in the design system"
- **what:** Base CSS: body sets --color-text; h1-h6 use var(--section-heading, var(--color-primary)); p/li/blockquote use var(--section-text, inherit); strong/em/span set no color; links are the explicit exception. Painting sections override --section-* on their container and all children adapt. Setting color directly on elements bypasses the override — the light-on-light testimonial bug.
- **sources:** 003(8)#CSS Colour Inheritance Model; 036 §4
- **relations:** section painting contract; buildSectionDefaults
- **verify-later:** layouts' element rules follow the fallback pattern

### Section painting contract (four models, references-only) — supersedes literal dark-section overrides
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) checklist item 6 + Section Context Variable Contract; (7)→(8) diff shows the literal-colour version replaced; "fix_forced_text_colors enforces this mechanically"
- **what:** A template's appearance derives from what its own CSS paints; is_dark_section is catalogue metadata and must not key styling. A painting section picks exactly one model — pair band, palette band, image/layered (hero-ink), or ambient (no --section-* at all) — and re-exports --section-* AS REFERENCES to the tokens it paints with (color-mix for muted/surface/border). Literal colours in --section-* declarations forbidden. The older contract (dark sections set literal rgba/white values) is superseded.
- **sources:** 003_contracts_and_standards(8).md items 6/6b + #Section Context Variable Contract; 003(7) (family-delta, superseded form)
- **relations:** scheme-to-components work (016b light-site-dark-chrome); forced_text_colors check
- **verify-later:** fix_forced_text_colors action; component templates conformance

### CSS theme template contract (renderer vs template ownership; theme storage/lineage)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) responsibility split; render pipeline confirmed in 036
- **what:** Renderer owns palette injection, luminance-driven --section-* defaults (pickReadableOnBackground preserving palette character), and css_snippets appends; the theme template owns layout/typography/component styling using the fallback pattern and MUST NOT declare --section-* defaults or hardcode text hexes. css_template (Go template) vs css_content (frozen fork snapshot, reference only).
- **sources:** 003(8)#CSS Theme Template Contract; 036 §3–4
- **relations:** 025 palette/layout/typography split; buildSectionDefaults
- **verify-later:** render_css_from_spec_action.go; color_util.go

### Query parameterisation contract ($1 + params, never template interpolation)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): rule + named legacy offenders (tool-suggester, tool-improver) still to migrate
- **what:** All new query_database usage must use $1 placeholders with a params array of dot-paths (passed as query args); {{.field}} embedding is a SQL injection risk. QueryDatabaseAction gained params support after audit agents failed on $1-with-no-args.
- **sources:** 003(8)#Query Database Parameterisation Contract; 001(5) bug 1
- **relations:** authoring rules pack
- **verify-later:** tool-suggester/tool-improver migrated?

### Schema enforcement: flexible vs strict mode with approval snapshots
- **category:** contracts-and-standards
- **status-signal:** unknown
- **status-evidence:** 003(8) describes the design (schema_snapshot/content_snapshot at approval, sites.schema_mode) with no deployment claim or date
- **what:** Initial build runs flexible (best-effort substitution, warnings); at approval the structure locks: page_components.schema_snapshot + content_snapshot captured, sites.schema_mode → strict, mismatches become validation errors, template upgrades can't break approved pages.
- **sources:** 003(8)#Schema Enforcement (Flexible vs Strict Mode)
- **relations:** locks; content governance approval flows
- **verify-later:** schema_mode column usage; any strict-mode site

### Handler input-path contract (input_data.spec.*) + action-level defense
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract rule; 016 §9 "most common systematic failure" documents the violations
- **what:** The dispatch loop passes the work-item spec nested at input_data.spec; handlers MUST read spec fields there, never rely on top-level flattening (which exists only for legacy `?` promotions and silently nils). Go actions reading common fields implement a fallback chain (explicit config → input_data.spec.field → well-known spots). Known offenders were tool-improver/tool-auditor/rerender-pages. Manual spawn+call of work-item agents must satisfy BOTH the top-level input_contract AND the workflow's spec paths (provide fields in both shapes).
- **sources:** 003(8)#Handler agent contract/#Input data paths; 016 §9 path-mismatch; 016b#Manually invoking an agent
- **relations:** dispatch loop; input_contract validation
- **verify-later:** load_page_record_action.go fallback chain

### Legal rules schema and content_direction page-level instructions
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** 003(8) schemas defined; legal-content-agent "Planned" in 002(4)
- **what:** Per-site legal_rules (required disclaimers with triggers/placement, forbidden phrases, required pages seeded per industry) in sites.content_data; pages.content_direction jsonb (format/instruction) flows to the content writer for page-level rewrites via needs_rebuild.
- **sources:** 003(8)#Legal Rules Schema, #content_direction
- **relations:** content agent family; compliance discovery
- **verify-later:** any legal_rules populated; content_direction reads in writer prompt

### Improvement loop flow with pass cap, finding cap, and auto-reset ("sites breathe")
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 v3/v4 dated changes (2026-03-25/31); token math before/after (~88K→~30K)
- **what:** improvement-loop orchestrator: pass-limit gate (≥3 → complete_clean) → algorithmic discovery agents (quality/design/completeness) → LLM audits (TOP 5 findings each with current_value/acceptance_test/suggestion/max_fix_attempts) → triage → increment pass → insert needs_rerender p99 → dispatch. Auto-reset after 60 days / direction change / major rebuild / manual, pairing with lock expiry (the unimplemented half) to create the breathe rhythm.
- **sources:** 004 full; 031_LOCKS (the missing half)
- **relations:** section locking; direction spec; locks-should-expire question
- **verify-later:** improvement-loop workflow; audit_pass_count fields

### Discovery checks catalogue (quality/design/completeness) and ordering rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 tables with handlers per check; validate_component_standards sub-checks
- **what:** quality: broken_nav_links, placeholder_contact, generic_theme. design: hardcoded_section_colors, forced_text_colors, undeployed_assets, missing_css, validate_component_standards (9 sub-checks incl. unlinked components, slot mismatch, nav layout, unwanted search icon). completeness: empty_sections, empty_blog, orphan_pages. validate_component_standards runs BEFORE colour checks (structure before rendered-HTML checks). DiscoveryCheck interface: checks append WorkItemSpecs; the runner inserts with dedup — plugins must not insert their own rows.
- **sources:** 004#Discovery Agents, #Component Standards Validation; 016 §0 item 27 (interface shape)
- **relations:** two-strike dedup; handler routing
- **verify-later:** discovery_checks/ registry

### Section locking with lock types and expiry (design vs implementation gap)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_LOCKS (investigation): "The columns don't exist in schema… expiry mechanism specced never built" while pass-reset half IS implemented; 004/007 describe lock_type permanent/timed/review as if landed
- **what:** Design: components that pass verification lock; lock_type permanent/timed(default 90d)/review(HITL on expiry) with query filter expansion `(locked_at IS NULL OR lock_expires_at < NOW())`. Reality: only plain locked_at/locked_by exists; auto-lock-on-deploy fires on every dashboard edit, so lock proliferation monotonically shrinks the improvement loop's surface (three documented failure modes). Recommended: timed default for routine edits, permanent opt-in.
- **sources:** 031_LOCKS_should_locks_expire.md; 004#Section Locking; 007#Lock lifecycle
- **relations:** audit pass auto-reset; lock coherence debt
- **verify-later:** lock_type/lock_expires_at columns exist?; discovery query filters

### Lock semantics: hard gate for discovery, soft gate for execution, read-only rerender
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 013 Phase 1 ✅ with the four amended checks; behaviour tables
- **what:** Lock means "human controls this", not read-only: edit refreshes locked_at without unlock; unlock is a separate deliberate act. Discovery checks skip locked rows (hard gate); execution agents process explicit items regardless (soft gate); rerender reads everything. locked_by vocabulary: admin/admin-removed/checkpoint (human-only unlock) vs deploy (agents may clear). Three lock levels: component, site component, whole site (site lock stops all automation via LoadWorkItemsAction gate + pre_query filter).
- **sources:** 013#Three Levels of Lock, #How Agents Behave; 031(3)#rules
- **relations:** growth budget; suppression
- **verify-later:** lock_helpers.go; four discovery checks' filters

### Lock patterns A/B, Pattern B (pinned) is dead, and lock transfer across plan rebuilds
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031(3) verification 2026-05-19: "Pattern B is unenforced in the current code — treat it as dead"; lock transfer specced for Phase 1 site_plan_directives
- **what:** Pattern A locked_at/locked_by (+partial index) is the dominant per-row pattern (sites, page_components, site_components, site_plan_directives). Pattern B pinned boolean on site_specs was never wired (no reads/writes; every spec write is supersede-then-insert with no guard) — new tables must use A. Lock transfer: only the rewriting agent (write_site_plan) copies locks onto matching new rows by composite key; locked text beats LLM rewrite; unmatched locks drop with a log. Locks and snapshots are orthogonal (prevention vs restore); open question whether revert respects locks.
- **sources:** 031_locks(3).md; 030 Q1 directives schema; 013 (pinned column added Phase 4 — UI-level only)
- **relations:** plan-domain tables; spec pin/propagate UI
- **verify-later:** \d site_specs pinned; write_site_plan lock-transfer code

### Human direction channels and the pinned direction spec
- **category:** content-governance
- **status-signal:** partial
- **status-evidence:** 007 channels table; 004 v4 integration deployed for auditors; dashboard direction panel "pending" in 007 Phase 1 then ✅ in 013 Phase 4
- **what:** Three channels: work-item request (until completed), direction update (permanent, site_specs `direction` aspect, pinned — only humans change), reference suggestion (feeds next planning). Auditors must not flag must_have/should_have features for removal; strategist may add nice_to_have but can't remove must_have; direction change resets the audit pass counter.
- **sources:** 007#Human Direction; 004#Human Direction Integration; 013 Direction tab
- **relations:** audit pass reset; spec propagation
- **verify-later:** direction aspect reads in auditor prompts

### Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 005(1): "Pipeline Status: Fully Operational… all work without manual intervention" with per-site verified results
- **what:** check_missing_tools → tool-suggester (LLM judgement, 0-5 suggestions, library-vs-novel routing via check_is_library) → tool-deployer (fork) or tool-generator (novel) → create_cross_links (content_rewrite items per related page, item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (`input_data.spec.suggestion`) into the writer's nested loop prompt → rerender. The writer prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it (072 trap).
- **sources:** 005_tool_pipeline(1).md full; 020 agents detail
- **relations:** de-tool hazard; fork-on-deploy; tool doc header
- **verify-later:** migrations 070–073 applied; cross-link items in prod

### Fork-on-deploy tool ownership model
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019: "This is deliberate. A bad library change shouldn't break ten sites simultaneously."
- **what:** Library tools (component_level='tool', forked_from NULL) are blueprints never referenced by pages; deployment forks a copy per site (forked_from set) and the site owns it — library changes never cascade; pushing improvements to sites is per-site work items. Orphan-fork retry safety: two-stage existing-fork check (P105 fix); GetComponentByFunction excludes forks.
- **sources:** 019#Core Concept; 105 item 6 fix; 020 tool-deployer
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL)
- **verify-later:** deploy_tool_action.go two-stage check

### Tool doc header (sentinel comment; stripped at deploy)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no */). It never ships: StripToolDocHeader runs at the three outbound assembly points; DB rendered_html retains it for audit parity. Creation gate validates presence; improver preserves/updates it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by tool_health.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks
- **relations:** per-tool travelling docs (037); tool-auditor
- **verify-later:** platform/content/tool_doc_header.go; prompt migration applied

### Tool quality three tiers (structural / LLM audit / headless-browser future)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 020: tiers 1–2 automated live, tier 3 "planned"
- **what:** tool_health tier 1 (Go, free): deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks — blockers create improve_tool. Tier 2: audit_tool queued (30-day/tool cooldown) → tool-auditor Sonnet code review across six categories, findings by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked. Tool removal is a human decision via dashboard.
- **sources:** 020#tool_health, #tool-auditor; 019#Tool Quality Standards
- **relations:** tool-improver; component-quality-auditor (sections)
- **verify-later:** check_tool_health.go cooldowns

### Never load html_template in listing queries (storage discipline)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + query audit section
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

### News feed pipeline (sources → async ingest → triage → JSON render → commit)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table all ✅ with gaswholesalers evidence; scheduled 6-hourly
- **what:** content-feed-trigger finds recommended sites → content-feed-orchestrator: seed_sources (from classification spec) → dispatch ingesters async per due source (rss/news_search/api_news/scrape) → feed-triage scores PRIOR runs' items → render latest-news.json (6) + news-archive.json (20) → git commit. Two-pass by design (ingest now, triage next run). Homepage snippet + /news.html listing page both client-fetch the JSON — news updates decoupled from page rerender.
- **sources:** 006 full
- **relations:** growth budget; content-gap-planner chain; source diversity
- **verify-later:** content_sources/content_feed_items; content-feed-refresh task

### Feed triage: relevance + credibility + source-attribution provenance
- **category:** news-feed-pipeline
- **status-signal:** partial
- **status-evidence:** 006 deployed, but Known Open Issues: "credibility always 0 … fields exist but aren't being populated"
- **what:** LLM triage scores relevance 0-100 and credibility high/medium/low with attribution chain {original_source, found_via, source_tier} across a 6-tier source taxonomy; rejects fabricated URLs, nav links, uncorroborated low-credibility claims. Status lifecycle ingested→relevant/review/rejected→expired(30d).
- **sources:** 006#feed-triage, #Issues; #Resolved Decisions 47
- **relations:** diversity scoring plan; Grok provider choice
- **verify-later:** credibility population bug fixed?

### Real-time-search news providers (Grok Responses API decision)
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 resolved decision 48; hallucinated-URL bug table entry
- **what:** api_news sources route to xAI grok-4-1-fast via Responses API with web_search+x_search, OpenAI Responses (gpt-4.1-mini) or Perplexity sonar — all real-time search; chat-completions grok-3-mini hallucinated 2023 URLs and was dropped.
- **sources:** 006#fetch_llm_news provider routing
- **relations:** feed triage credibility
- **verify-later:** provider keys in personae-default-secrets

### Render source-diversity interleaving
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** 006 status table ✅; decision recorded after single-source domination
- **what:** loadNewsItems uses ROW_NUMBER() OVER (PARTITION BY source_id) ordered by source_rank then recency so each source contributes at most ~2 of 6 display slots; with topic-focused sources this also yields topical diversity.
- **sources:** 006#Render action source diversity, #Content Diversity §6
- **relations:** topic-focused source splitting (planned)
- **verify-later:** render_news_section_action.go query

### Page growth budget (three-tier weekly limits)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 006/013 ✅; the news-index-blocked-by-content-budget bug drove the third tier
- **what:** CheckPageGrowthBudget shared by apply_gap_plan and create_blog_posts: free under initial_target (12), then rolling 7-day caps per type — content 3/wk, blog 2/wk, structural (news-index, privacy…) 3/wk — absolute_max 60. Over-budget items become `blocked` (retryable), blog posts skip-and-continue. Config in site_specs aspect growth_config (pinnable).
- **sources:** 013#Page Growth Budget; 006 growth budget rows
- **relations:** content-gap-planner; blog planner
- **verify-later:** page_growth_budget.go tiers

### Content diversity & original research pipeline (readership segments, timelines, scenario analysis, engagement)
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** 006 "What's Not Built Yet" lists every piece (topic splitting ready-to-implement; article-rewriter/feed-publisher/feed-lifecycle blocked/unbuilt)
- **what:** Planned evolution: topic-focused source splitting (SQL-only), coverage-gap pre-fetch step, multi-language regional discovery with triage translation, triage diversity scoring, research-agent multi-step investigations (fact/history/quotes/numbers) → writer targeted per readership segment (procurement/ops/trading/strategy) → eval agent quality gate → publish; continuous annotated timelines with pattern recognition; if/then scenario analysis (no predictions); client-side engagement measurement feeding content planning.
- **sources:** 006#Expansion Roadmap, #Content Diversity & Research Pipeline
- **relations:** research-agents; batch API integration
- **verify-later:** none built — check for article-rewriter definition

### Infrastructure three layers (core platform / client delivery / framework builder)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007: Layer 1 exists; Layer 2 "planned" beyond static serving; Layer 3 "future"
- **what:** Layer 1 (K8s factory) never serves external traffic — it produces artifacts and pushes them. Layer 2 = client delivery (S3 static now; site-api-router + client Postgres on OVH VMs planned, config-driven routes reusing the action library). Layer 3 = provisioning agent frameworks for clients. Backend capability tiers 1–5 (static→full platform) with static-first principle; vetcomparison JSON-on-S3 pattern handles up to ~50k items (Pagefind extends to ~500k).
- **sources:** 007#Infrastructure Separation, #Backend Capability Tiers, #Site API Router
- **relations:** dynamic applications tiers (022); P1 marketing/OpenClaw
- **verify-later:** any site-api-router code

### Adoption is a one-off capture, not a ceiling (specs separation)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 v4 principles + what-gets-stored table; patch updates to Phase-1 split
- **what:** Crawl data → research_results (adoption_crawl/adoption_page), never site_specs. Spec aspects: identity (with adopted_from provenance), design_reference (concrete extracted values, historical, never modified), design_intent (semantic brand-level brief, survives plan rebuilds post-030), site_archetype (character + inviolable constraints), content_direction (brand voice), structure. Webdesign reads intent not reference; evolution = update intent. The strategist then writes aspiration beyond the adopted baseline.
- **sources:** 007#Site Adoption, #What gets stored where (incl. patch revision); 004#Adopted Sites
- **relations:** 028 fidelity; 030 strategic-vs-plan-time split
- **verify-later:** site_specs aspects on an adopted site

### Adoption agent three-stage processing (Go fingerprint / LLM classify / Go content+plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 16-step workflow with runtime expectations and success signals
- **what:** site-adoption-orchestrator (wrapper) spawns site-adoption-agent: firecrawl crawl → Go extract_design_fingerprint (goquery over rawHTML: hexes, fonts, CSS vars, layout patterns, Google Fonts; external CSS fetched and merged; suggested_mapping to our var names) → LLM analyze_site on ~500-char page summaries + classify_archetype + derive_content_direction + generate_design_intent → Go apply_adoption_plan (buildCrawlPageIndex, page records, per-page markdown to research_results, design_reference spec, work items). Principle: LLM for reasoning, Go for extraction — never pay an LLM to transcribe.
- **sources:** 007#The adoption agent, #Three-stage processing, #Design Fingerprint Pipeline
- **relations:** wrapper pattern; classifier handoff
- **verify-later:** extract_design_fingerprint_action.go; adoption agent definition

### Source vs destination separation in adoption (target_url / destination_domain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 modes table + operational signals ("ApplyAdoptionPlanAction: source differs from destination")
- **what:** Adoption separates the crawled reference site from the site being built; ensure_site_record uses destination via domain_override_field, crawl hits target_url, provenance records source host — which also keys the crawl-content lookup (mismatch silently drops all adopted content). Legacy single-domain shape still accepted.
- **sources:** 007#Source vs destination, #Adoption modes
- **relations:** build-inspired-by mode (goes via classifier instead)
- **verify-later:** apply_adoption_plan source/destination handling

### Adoption → classifier handoff (needs_domain_research; no shortcuts)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 section + patch; grounded in 028's "classifier always runs in full"
- **what:** Adoption queues exactly one strategic item, needs_domain_research; it does NOT queue needs_composition/needs_design directly — the cascade produces them naturally. The classifier reads site_archetype/design_reference as ground truth, read-and-extends identity/content_direction/design_intent, always runs vertical research, emits classification, queues needs_strategy. Post-030: planner writes to plan-domain tables and the reconciler emits page items; the planner's job ends at "the new plan is durably current".
- **sources:** 007#Handoff to the classifier, #Post-adoption (patched version); 028#The classifier is the strategic brain
- **relations:** unified pipeline; reconciler
- **verify-later:** adoption work-item emissions in apply_adoption_plan

### Component selector / creator: section_type vs function split
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) picks the variant; no candidate → needs_new_component work item → component-creator generates against the full component contract prompt and stores with metadata; quality feedback loop from auditor scores creates a fitness landscape. Backward compatible: direct function lookup remains path 1.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path)
- **relations:** component regeneration; component creation contract prompt
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?

### Pattern extraction, code-as-reference, and RAG-fed generation
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** 007 Phase 3 items; "Runs as a side effect… patterns accumulate" future tense
- **what:** A pattern-extraction-agent mines research into reusable tool specs, layout/content patterns, and good/bad UX examples; complex tool builds include reference code in the prompt with explicit original-implementation instruction (never deployed directly — copyright stance); prompt+output pairs feed RAG so future generations retrieve both abstract specs and concrete prior successes.
- **sources:** 007#Research, Patterns, and the Component Library
- **relations:** knowledge_base; tool-recreation-handler
- **verify-later:** none built

### Vet med pricing pipeline (discovery → scrape+evidence → export)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** 008 dated 2026-04-08 with per-retailer coverage stats; scheduled tasks configured (disabled pending verification)
- **what:** business-intel pod spawns Job pods per stage: URL discovery (category scraper or Firecrawl /map, deny-list filtered, upserted to med_retailer_listings) → price collection (Firecrawl scrape+screenshot, section truncation, retailer regex cascade → CPU Mistral fallback when £ present but 0 variants; snapshots + evidence rows; materialized view refresh) → JSON export (index/full/by-letter/metadata files, git commit → live). Multi-site via input_data filters (species/categories/retailers). Evidence chain: markdown + content hash + screenshot re-uploaded to B2, indefinitely.
- **sources:** 008 full
- **relations:** med-* wrapper orchestrators; LLM fallback as training data
- **verify-later:** business_intel.med_* tables; scheduled tasks enabled?

### LLM fallback extraction doubling as training data (med pricing)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 008: "response logged to llm_call_log… dual purpose: price extraction now, training data for future LoRA"
- **what:** Regex handles ~90% of pages; CPU Mistral (mistral-small3.1, temp 0.1) parses the remainder into a JSON variant array at 80-280s/call — acceptable because batch-tolerant — while accumulating markdown→JSON pairs for a future fine-tune. Future: product matching across retailers, price alerts from snapshot history, affiliate-feed switch.
- **sources:** 008#LLM Fallback, #Future Work
- **relations:** batch processing categorisation; llm_call_log flywheel
- **verify-later:** llm_call_log provider='ollama' step_name='scrape_prices'

### Multi-endpoint model routing with ai_endpoint_health as the GPU scheduler
- **category:** model-infrastructure
- **status-signal:** partial
- **status-evidence:** 009: tables/functions applied+verified; the three Go patches (fast-fail, claim-gate, release-without-attempt) listed under "Next Deploy"; active pinging "starts after patches deployed"
- **what:** Endpoints (Claude API, CPU/GPU Ollama) tracked in ai_endpoint_health; healthy → items flow, unhealthy → items wait (no fallback chains — quality over speed; priority means importance only; items don't know about models). ClaimWorkItem checks handler's endpoint health before claiming; AIUnavailableError triggers reactive health update + release-to-triaged without attempt increment; Claude health dual-mode (reactive 402/401 + hourly 1-token ping).
- **sources:** 009#Decisions Made, #Health Check Architecture; 022 SQL
- **relations:** back-to-triage; endpoint-health-checker agent
- **verify-later:** were the three patches deployed since 2026-03-25

### Model swap/snapshot/revert control plane (migration 083)
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** 021 SQL applied per 009; used operationally (016 §6.1 conventions)
- **what:** snapshot_agent()/swap_agent_model()/revert_agent() + agent_snapshots view make agent_definitions the model-routing control plane: per-step ai_service swaps with automatic snapshot, one-call revert. Post-migration snapshots live in agent_definitions_backup (snapshot_taken_at discriminator, restored_at audit trail); the legacy in-table is_snapshot rows caused a documented family of contamination/misroute bugs.
- **sources:** 021_model_swap_and_rollback.sql; 009#Model Swap Procedure; 016 §6.1/§9 snapshot bugs
- **relations:** backup naming discipline; LLM config shadowing (step-level swaps shadowed by top-level ai_service)
- **verify-later:** agent_definitions_backup schema

### Backup discipline: never drop or overwrite an existing backup
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009 rules section with naming convention `agent_definitions_backup_YYYYMMDD_pre<NNN>`
- **what:** Backups are named for the migration they guard; DROP TABLE IF EXISTS before CREATE destroys the recovery path exactly when needed (failed-and-retried migrations); name collision is the safety mechanism working — pick a new suffix. Nuclear full-table restore procedure retained.
- **sources:** 009#Operational Reference: Backup, Swap, and Revert
- **relations:** snapshot_agent; re-running-migrations hazard
- **verify-later:** —

### Model quality assessment: local 70B comparable for some tasks
- **category:** llm-quality-testing
- **status-signal:** deployed
- **status-evidence:** 009 test table dated 2026-03-24; 023 raw comparative transcripts
- **what:** Llama 3.3 70B (single H100, num_ctx 8192) scores 8-9/10 vs Claude for classification/content, 7/10 design; Mistral Small 3 CPU adequate only for low-stakes structured tasks (5/10 classification, 3/10 design). Evaluation criteria captured in 023: JSON parse w/o fences, exact field names, specific headlines, action-verb CTAs, no invented claims. ThunderCompute quirks: 2-GPU instances broken, num_ctx metadata bug, KEEP_ALIVE=-1.
- **sources:** 009#Model Quality Assessment, #ThunderCompute Notes; 023 full
- **relations:** fine-tuning path; LLM tiering
- **verify-later:** —

### Kafka scheduler (DB-driven heartbeat service)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 010 full deployment reference (migration 066, kustomize, terraform paths)
- **what:** Single-replica Go producer-only service ticking 30s over scheduled_tasks: interval elapsed + concurrency-group capacity + pre_query gating → publish standard orchestrate message (from kafka-scheduler identity, responses to system.scheduler.responses — currently unconsumed). Adding a schedule is an INSERT. Pre-queries provide dynamic input (first row merged into input_data) and gating (no rows = skip). timeout_seconds is the in-flight safety valve; double-fire tolerated via idempotent work-item dedup.
- **sources:** 010 full
- **relations:** build-pipeline-trigger; improvement-sweep; med tasks; batch submitter/retriever placement
- **verify-later:** scheduled_tasks rows; cmd/scheduler/main.go

### Three databases and the pgbouncer transaction-mode constraints
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 full operational reference
- **what:** clients_db + templates_db (in-cluster PG16 via pgbouncer:6432 transaction mode) and external MySQL auth DB (Clook cPanel, Remote-MySQL IP whitelist 134.213.168.%). Transaction mode forbids prepared statements/session state — conn strings need simple_protocol/cache_describe; pg_dump and LISTEN/NOTIFY must bypass pgbouncer. Go driver split: chassis still pgxpool, core-manager on database/sql (conversion cheat sheet). Auth-to-Postgres migration sketched, not urgent. All credentials in personae-platform-secrets (Terraform 047-base-configs).
- **sources:** 011 full
- **relations:** admin auth flow; backup cronjob
- **verify-later:** which binaries still pgxpool

### Client schema isolation (create_client_schema)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 runbook with current clients (demo_client, vetcomparison, test_client, system)
- **what:** Each client gets schema client_<id> with agent_instances/agent_spawn_history/projects (+optional website_projects/agent_memory/workflow_executions); spawn_agent resolves the schema from the client_id Kafka header (validator checks presence only, no DB lookup) and inserts exact columns — manual table creation must match spawn_actions.go's INSERT or spawning fails.
- **sources:** 011#Creating a New Client Schema
- **relations:** multicluster/multitenancy; scheduler client_id header requirement
- **verify-later:** create_client_schema function source

### Admin dashboard + nginx gateway architecture
- **category:** admin-dashboard-and-api
- **status-signal:** deployed
- **status-evidence:** 012 full; 013 phases 1-11 ✅ except user portal
- **what:** React SPA served by nginx that also gateways /api/v1/auth→auth-service and /api/v1→core-manager (rate limits, timeouts, immutable asset caching, security headers). Views: Sites (lock badges), Work Items (three review flows: placeholder/checkpoint/standard; bulk retry; cross-site tab), Pages (three-level browser, Fields/HTML/Brief tabs, page-purpose bar, suppressed-section restore), Direction (spec cards, pin/propagate), Media (assets + references). Access via port-forward today; WireGuard/bastion planned.
- **sources:** 012 full; 013 status table
- **relations:** content governance edit paths; admin API endpoints
- **verify-later:** frontends/admin-dashboard; nginx conf

### Three edit paths (direct edit / brief regenerate / direction propagate)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** 012/013 shared table; all phases marked deployed
- **what:** Typos/fields: direct edit + auto-lock + rerender (seconds, no LLM). Section direction: edit content_brief → Regenerate → content_rewrite item, writer prompts from the brief (briefs control regeneration). Page: purpose edit + Regenerate Page (all unlocked sections). Site: Direction tab spec edit + explicit Propagate creating per-page items (skips fully-locked pages). Section suppression (page-scoped remove/restore) completes the set.
- **sources:** 013#Three Edit Paths, #Content Briefs & Regeneration
- **relations:** lock semantics; page_components.content_brief
- **verify-later:** content_brief column; regenerate endpoints

### Universal LLM work queue (batch processing architecture)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 v4 (2026-04-12) is a design + phased rollout plan (Phase 0 "deploy infrastructure, everything OFF"); no deployment claim found
- **what:** llm_batch_queue as a provider-agnostic queue (rendered prompt + resolved callback_config stored at queue time) with three-gate resolution (global → agent_type opt-in → provider) and a sync fallback that executes the whole path inline (sync_executed rows prove the restructured pipeline before batch is enabled). Submitter routes to Anthropic Batch API (50% discount, caching adjacency by batch_group), GPU drain mode (worker pool, drain_until/stop-when-empty), or sync; retriever polls, logs to llm_call_log, executes callbacks with retries; urgent escalation makes parallel sync calls and marks late batch results superseded. Batch/sync decision rule: scheduled-task-triggered → batch; user-facing/blocking → sync (~60-70% of spend batchable).
- **sources:** 015 full
- **relations:** callback contract; prompt caching; endpoint health
- **verify-later:** do llm_batch_* tables exist; queue_llm_batch registered?

### Batch callback contract (resolved-at-queue-time, context-free callbacks)
- **category:** batch-processing
- **status-signal:** aspirational
- **status-evidence:** 015 design; eligibility test with three passing callbacks named
- **what:** Callbacks receive only DB + response + resolved callback_config (no collected_data/orchestration state); eligibility test: can it work from a DB connection, response text, and a handful of resolved IDs? Workflow restructure: the post-LLM step disappears into the callback; multi-provider preference lists auto-route without workflow edits.
- **sources:** 015#Callback Contract, #Workflow Restructure
- **relations:** write_audit_findings as first callback
- **verify-later:** batch_callback.go exists?

### Debugging assumption checklist (28-item process discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §0, each item tied to a dated real defect through 2026-06-13
- **what:** The distilled pre-change checks: per-action _field conventions; input_mapping required-by-default; error_preview before log grepping; partial DB rows; SQL-immediate-vs-Go-deployed; sibling functions as canonical pattern; token budgets vs structured output; set -u in trigger scripts; jq slurpfile nulls; manual triggers to isolate dispatch-vs-handler; parent/child orchestration rows; `?` placement; \d before SQL; refire-before-refactor; pod-rotation log loss; don't change evidence-proven values; deploy ≠ migration ran; interface widening breaks all importers; prompt_rendered proves input not output; updated_at is not authorship; re-resolve site_id after teardown (zero-row LEFT JOIN = wrong anchor); check design docs for deliberate deferral; output_fields plural; config authoritative only at its runtime read-path; prove the harness delivered input (dash vs bash); env vars vs shell locals + stale deployed copies; read the interface definition; agent_definitions UNIQUE(type,version).
- **sources:** 016 §0 items 1–28 + 016_additions
- **relations:** everything in §9; 016b durable invariants
- **verify-later:** —

### An agent is a DB row; trust default_config over prose; two possible definition sources
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §6.0 with build-dispatch-loop description-vs-config contradiction example
- **what:** Agents live in agent_definitions.default_config.workflow — grepping Go finds actions, not agents. Descriptions can contradict configs (trust the config). agent_definitions may be read from templates_db or clients_db depending on pod — confirm which copy the running pod loads before patching.
- **sources:** 016 §6.0
- **relations:** orchestration state; snapshot conventions
- **verify-later:** which DB each deployment reads definitions from

### LLM step config shadowing (ai_service resolution order; dead temperature paths)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §6.6: bug live as of 2026-05-18 (22 of ~60 agents shadowed); structural per-field fix "planned"
- **what:** ExecuteLLMPromptAction resolves ai_service top-level → step-level → StepConfig and stops at first match, so a top-level ai_service shadows every step override (incl. doc-023-style per-step model swaps); max_tokens falls to hardcoded 2048 (tell: output_tokens exactly 2048); step.config.max_tokens sibling is never read; temperature is read ONLY from default_config.temperature top-level (all other locations dead) and llm_call_log.temperature was universally NULL. Fix path: per-field fallback chain + raise floor to 8000 + log sent values.
- **sources:** 016 §6.6
- **relations:** model swap functions; __sent_* write-backs (001(5) suggests later capture landed)
- **verify-later:** whether per-field resolution shipped; llm_call_log temperature now populated

### Timeout chain ordering (claim > call_handler > handler workflow)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §7 with current values and both mis-ordering failure modes
- **what:** claim timeout (30m) must exceed dispatch call_handler (20m) must exceed handler workflow timeouts; otherwise duplicate handlers (claim reset mid-work) or orphaned completions (dispatch gave up early). Idle monitor 3600s fallback; K8s ActiveDeadline 24h ceiling.
- **sources:** 016 §7
- **relations:** claim-lease-too-short reproducible timeouts (v2_49 sub-case b)
- **verify-later:** current values across dispatch/handlers

### Work-item state machine, transition ownership, and site-exclusion by stuck claim
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 016 dedicated section with ownership table and the 2026-05-14 mis-diagnosis lesson
- **what:** detected→(design-audit-agent triage_detected_items)→triaged→(build-dispatch-loop claim)→claimed→(handler)→complete/failed; admin inserts at triaged. Most common "won't dispatch" cause: one stuck claimed item excludes the ENTIRE site via find_dispatchable_site's NOT EXISTS. Debugging trap #1: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** 016#Work item lifecycle; #Site excluded from dispatch
- **relations:** two-strike; claimed-item-timeout; operator reset SQL
- **verify-later:** triage_detected_items registration (registry.go:722)

### Reaper false-positive completions (claimed-item-timeout evidence checks too loose)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 with confirmed gaswholesalers instance (auto-complete 47 min before the real commit); fix described as small SQL, not marked applied
- **what:** The "verified done despite lost response" branch auto-completes on ANY page updated on the site (not the target page, and updated_at not deployed_at) — treat empty-result + 'Auto-completed' items as untrusted. Correct evidence: p.id = wi.page_id AND deployed_at > claimed_at; needs_rerender/needs_design shouldn't auto-complete this way. Sibling issue: orchestration engine doesn't enforce awaited_requests timeout_at (spawn-handler hangs until reapers paper over it).
- **sources:** 016 §9 claimed-item-timeout + spawn_handler-hang entries
- **relations:** silent-completion family; timeout chain
- **verify-later:** claimed-item-timeout pre_query current form

### needs_section_data: resolvable-by-query vs genuinely-human, and the section-data reconciler direction
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented; reconciler recorded in FUTURE_section_data_handler
- **what:** Some needs_section_data items are human-only (team, pricing); list/grid sections source from query.* and resolve mechanically once pages exist — but the unimplemented query name or pre-active timing defers them anyway. Read spec.missing[].source to classify. Direction: implement pages_under_section + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via closeResolvedDataRequest, and flags re-render.
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** input schema v2; deferral-drop bug
- **verify-later:** queryresolve vocabulary today

### jsonb && operator class bug (silent CSS-snippet failure vs hard JS failure)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9: css path "silently failing the entire time"; JS analog fixed in same change set (May 2026)
- **what:** Postgres has no jsonb&&jsonb overlap operator; `applies_to && $1::jsonb` errored forever, swallowed by a logger.Warn-return-"" handler, so css_snippets never reached any deployed styles.css. Fix: EXISTS + jsonb_array_elements_text. Wider lesson: silent-failure loaders + graceful consumers hide months-old breakage — prefer hard failure when the data is supposed to be there.
- **sources:** 016 §9 jsonb && entry
- **relations:** best-effort-needs-monitoring; audit grep pattern for other && uses
- **verify-later:** loadComponentCSSSnippets fixed in place

### Silent-completion family (trust the artefact, not the status)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016/016b recurring entries; 016b names it the first durable invariant
- **what:** Work items report complete while the work didn't happen, via: result-contract stub (fixed 06-18); content-regression guard error masked by error_step complete_error; pod dying mid-flight (complete with non-empty error); "git committed the file" re-committing stale stored components; zero-planned-sections completing as success. Verify against page_components timestamps + live HTML. Companion rules: completed_at is orchestration END not write instant (trace child orchestrations by page_id in collected_data — trap part 3); intermediate signals (work-item names, pod snapshots, mid-flight tables) lie (trap part 2).
- **sources:** 016 §9 several entries + traps 1–3; 016b invariants
- **relations:** workflow result contract; zero-planned-sections
- **verify-later:** —

### save_page_sections is the sole page_components writer; its section-regex fallback and content-regression guard
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016 §9 tool-pages-never-deploy (fix shipped 2026-05-28, end-to-end verification honest-open); guard masking fix flagged not confirmed applied
- **what:** save_page_sections DELETE+INSERTs page_components (history row written; source_item_id NULL on overwrite path — gap). Its HTML fallback extracted only `<section>` blocks, so `<div class="tool-page">` tool HTML was silently discarded (all tool/game pages n_rendered=0, rerender skips, no file ever committed); fixed by whole-fragment-as-one-section fallback (guarded against full documents). The content-regression guard (new text < existing/4) protects prose but returned errors that complete_error converted to success. Deferred sections' instances are dropped on save (carry-forward pending, cousin of the interactive clobber).
- **sources:** 016 §9 tool-pages entry + guard entry; 016b Part 5/regenerated-section entries
- **relations:** de-tool hazard fix layers; deployed→needs_rebuild flip
- **verify-later:** patched save_page_sections deployed to all three callers

### built_from_plan_version deploy-time stamp replaces the deployed→needs_rebuild flip (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 016 §9 dedicated entry (2026-05-28) "Fix shipped", completing the HANDOFF_2026-05-07 deferred design
- **what:** upsertPage's blunt flip (deployed→needs_rebuild on every sync) stood in for the unbuilt drift stamp; Option B stamps built_from_plan_version at the UpdatePageStatusAction deployed chokepoint and makes sync fill-if-null, retiring the flip so drift detection flows through the reconciler's decideEmit. Lesson (checklist 22): a "bug" may be a half-implemented design — complete it, don't patch around it.
- **sources:** 016 §9 flip entry; 029/030 design
- **relations:** reconciler; tool-page churn
- **verify-later:** any direct build_status='deployed' writes bypassing the action

### Adoption slug-mangling: two canonicalisation surfaces must agree
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 016 §9 chain of entries (2026-05-19→26): cause pinned to WriteSitePlanAction ValidateRoles strip + SyncPagesToDBAction canonicalising raw page_plan WITHOUT ValidateRoles; fix "CHOSEN" (option 2) not confirmed shipped
- **what:** ValidateRoles strips tool-/guide-/game- prefixes and -index; CanonicalisePage re-adds them only for tool/game/guide roles, so wrong page_types (hubs typed content, guides typed blog-post) permanently flatten names/URLs. sync_pages_to_db reads raw page_plan (not site_plan_pages), skips ValidateRoles, and its ON CONFLICT overwrites correct adoption-time rows — one logical page, two writers, divergent results (incl. tool-game-* double prefixes). Fix: run the identical ValidateRoles+CanonicalisePage pipeline in sync (works for all five callers incl. plan-less pageflow-builder); root fix upstream is correct page_type at adoption; endgame is 029's deterministic slug preservation.
- **sources:** 016 §9 three linked entries; 030 phase-0 result
- **relations:** CanonicalisePage; page_type vocabulary
- **verify-later:** SyncPagesToDBAction ValidateRoles call present?

### Thunder adapter (credential-boundary GPU provisioning with caps and reaper)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** 033 confirmed decisions; 016 §9 entries show it running in prod (provision 400s, ssh findings, reaper window maths)
- **what:** All Thunder Compute interaction routes through a long-lived cluster adapter holding THUNDER_COMPUTE_API_KEY/B2 keys/ephemeral SSH keys; VMs are per-run ephemeral and credential-free (presigned URLs only; compromise blast radius = that run's files). Caps: $100/day rolling, 18h hard uptime, concurrency 2, 15-min reaper reconciling API↔thunder_instances. Formally retracts the on-VM HTTP job-server option. Operational lessons: lowercase gpu_type enums, template 'base', OpenAPI examples aren't valid values, tnr connect does server-side setup, login user ubuntu not root, live-instance-scoped partial unique index for recycled provider ids.
- **sources:** 033 full; 016 §9 Thunder entries
- **relations:** batch drain mode; training launcher/monitor
- **verify-later:** thunder-adapter deployment; reaper task

### Adapter/response message envelope contract (normative)
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** 035 §1 "last verified against code 2026-06-11"; 003(8) now points here
- **what:** Any reply to a chassis request must be recognised as an awaited response or it silently falls to process-as-work (row stuck waiting, ~10-min retries, no error). Load-bearing field: in_response_to_request_id = incoming request_id (request_id fallback only; git adapter's reuse pattern favoured). Three header tiers (validator-enforced / coordinator-needed-but-unvalidated / observability). Body headers MUST be a typed struct with real bools (map[string]string string-bools fail unmarshal and drop the reply pre-claim — the multi-day thunder outage). Send via ProduceWithValidation. Request parsing: action from body, payload at body.data, accept reply-topic from three keys. Sibling race: local dispatches must preRegisterAwaitedRequest before send (confirmed fixed in prod 06-09).
- **sources:** 035 §1; 016 §9 bool-trap + race entries
- **relations:** awaited_requests; O(K²) batch presign
- **verify-later:** ValidateOutgoingMessage field list

### Presign loop collapse: batch adapter calls over awaited-loop iterations (O(K²) state bloat)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 marked DONE + CONFIRMED IN PROD 2026-06-09 (migration 110; 26s vs never-finishing)
- **what:** Every step transition re-persists the whole orchestration state (expanded workflow + collected_data + history), so a K-iteration awaited loop is O(K²) and geometrically slows; the structural fix is one batch adapter call (prepare_object_urls returning all URLs in one reply) — deleting both the race class and the bloat class. Related fix: configOrInput now coerces numeric config scalars (expiry_minutes 3000 was silently dropped by a .(string) assertion).
- **sources:** 016 §9 presign entries
- **relations:** loop mechanisms; envelope race
- **verify-later:** training-launcher def shape (2d state check)

### Hand-applied agent-def migrations have no ledger; re-running an earlier one reverts later ones
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-09): re-running 109 silently reverted 110+111; runbook corrected
- **what:** Agent-definition jsonb migrations are hand-applied with no runner/ledger — the live def SHAPE is the only source of truth. A migration is idempotent only vs its own prior application, never vs later migrations on the same object; recover from doubt by checking state, never by re-running. Per-migration state checks (runbook 2d) after every deploy.
- **sources:** 016 §9 re-running-idempotent-migration entry
- **relations:** backup discipline; deploy≠migration checklist item
- **verify-later:** —

### Image doesn't contain the binary (CrashLoop exec not-found ⇒ build/packaging fault)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 §9 (2026-06-14): thunder-adapter tag shipped the analyser binary (overwritten Dockerfile, shared digest); third deploy-regression in a row
- **what:** `exec ./X: no such file` means the image lacks X — inspect image contents (docker run ls /app; Image-ID vs Image digest tells tag collisions), restore the Dockerfile, push a FRESH tag (never re-push the poisoned one). Guard: pre-push ls /app or a CI binary-name assertion — "no guard between built and running" is the recurring gap.
- **sources:** 016 §9 CrashLoop entry
- **relations:** deploy≠migration; stale-artifact family (checklist 24/26)
- **verify-later:** CI assertion added?

### Companies House enrichment pipeline (bulk collect → local match → LLM review → detail fetch)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 status header with March 19 2026 results (5,780 companies, 634 confirmed matches, 23.2%)
- **what:** Five stages on the business-intel pod: bulk SIC collection into a local mirror (trigram-indexed); two-pass local matching (~10s, no API): postcode+name scoring cascade then GiST trigram name-only with three tiers (≥0.90+distinctive auto-accept / 0.50-0.90 pending_llm_review / reject) and a distinctive-word check against generic-word inflation; haiku LLM review in batches of 15 with industry context from step config (~$0.05/run); detail fetch (profile/officers/PSC, succession risk from officer DOBs); discovery of unmatched companies planned. Revised matching cascade (tiers 0–6: website company number, exact+geo, exact-unique, postcode+moderate, LLM full-context, corporate-group parent, HITL queue) is the forward plan.
- **sources:** 017 full
- **relations:** vertical profile registry; business-intel pod pattern
- **verify-later:** ch_vet_companies match_method distribution; cascade implemented?

### Vertical profile registry (generic-words/keywords/suffixes per industry)
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** 017 "Vertical Generalisation (Implemented)"; ch_vertical_profiles.go
- **what:** Matching heuristics (industry keywords, generic word lists, scoring bonuses, name-cleaning suffixes) live in a Go profile registry keyed by vertical_slug from step config; LLM industry context is free-text in the agent definition config — new verticals are config, not code. Principle: search generality → scoring specificity.
- **sources:** 017#Vertical Generalisation, #Name Cleaning
- **relations:** ch_local_match; generic LLM review
- **verify-later:** ch_vertical_profiles.go registry contents

### business-intel shared-pod pattern (multi-type agents on one static pod)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 017 architecture section; ai_service placement rule
- **what:** Multiple agent definitions share one static pod via message routing (config.agent_type → selectWorkflow/FindBestGroup); consequence: ai_service must live in STEP config, not agent_config top-level, because agent_config comes from the pod's own type. Workflows are minimal action→complete with logic in Go. Single-replica contention accepted for batch work.
- **sources:** 017#Deployment, #ai_service on Shared Pods
- **relations:** wrapper-orchestrator (contrast); med pipeline (same pod)
- **verify-later:** business-intel deployment manifest

### Site spec unification (site_specs aspects as the one authoritative spec)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 021 target-state doc with immediate/short/medium ordering; backfill + fallback both recommended; newer pipeline already uses aspects
- **what:** One versioned spec per site as independent aspect rows (classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance), is_current/superseded_at per-aspect versioning with source/source_agent/source_item_id provenance; write_site_spec deep-merges so every row is a complete self-contained record (pruning-safe). content_data is legacy; read_site_spec falls back to it. Classifier writes intent, planner implements, design agent executes, audits enforce. The 15 content-strategy questions map onto aspects; deep research is a future classifier enrichment.
- **sources:** 021 full; P1#Site Specification System
- **relations:** 028 ownership; 030 strategic-vs-plan-time; dream spec
- **verify-later:** backfill done for legacy sites; read_site_spec fallback

### Platform mission and the single unified pipeline
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 028 (living document, rev 2026-04-22) is the stated check-yourself-against anchor
- **what:** Mission: given domain names in any state, produce the best possible website end-to-end with minimal human input, "best" = most useful to probable visitors measured by real engagement AND best revenue via whatever model genuinely fits. One pipeline for blank/adopted/missioned/replication domains — differing only in input material and the fidelity dial. Revenue model shapes the site (default-to-brochure/consultancy is a named failure mode); classifier decides the commercial shape.
- **sources:** 028#The mission, #Commercial viability
- **relations:** fidelity dial; classifier as strategic brain
- **verify-later:** —

### Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 028 Phases: Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)
- **what:** Controls how much aspiration reaches first deployment vs is deferred to the improvement loop, and the loop's promotion rate. locked = adopted-exact, no promotions; high (default with adoption) = faithful launch, ~one substantive change per audit cycle; medium = modest extensions; low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Lives on the trigger input + a build_policy/adoption_meta aspect.
- **sources:** 028#Fidelity, #Phased implementation
- **relations:** spec status; improvement loop rate
- **verify-later:** any fidelity/build_policy aspect in prod

### Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes planner still writing design_intent pre-030
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. Classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** 030 planner redirect to directives; composition self-heal
- **verify-later:** build-site-planner no longer writes design_intent/content_direction post-030

### Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it (fresh specs + stale installed theme at sites.style_collection_id)
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers from long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.) — long-term the supersession itself should emit the recovery item. Test: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (027)
- **verify-later:** whether supersession-emits-item was ever built

### Plan as declarative artefact + reconciler (Kubernetes-style desired-vs-realised)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030: Phase 0 done (2026-05-04 re-adopt verified dedup); Phase 1 schema/decisions committed; 007 patch describes reconciler emitting needs_page as current behaviour
- **what:** The planner stops emitting work items; it writes desired state to plan-domain tables (site_plans one-current-per-site, site_plan_pages, site_plan_sections, site_plan_directives) and a deterministic Go reconciler diffs plan vs pages and emits idempotent needs_page items (with preference weights, cycle budget, dependency ordering). Fixes the two-writer duplicate-pages structural bug (adoption + planner not sharing identity space). Phase 2: discoverers/auditors read the plan for sharper fitness checks.
- **sources:** 029 full; 030 full; 007_adoption_pipeline_v4.patch
- **relations:** CanonicalisePage; built_from_plan_version; directives
- **verify-later:** site_plans tables live; reconciler action name

### Strategic vs plan-time guidance split (site_plan_directives)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030 Q1/Q2 decisions + 007 patch stating planner "no longer overwrites" adoption specs; lock-transfer designed
- **what:** site_specs.design_intent/content_direction stay strategic (classifier/adoption-owned); the planner's per-build guidance flattens into row-shaped site_plan_directives (scope site/page/section, category, subject, directive, source, Pattern-A locks) read by downstream agents via a brief renderer. One LLM call still produces structure+design+content together (coherence over three-call split); only the write targets change. HITL locks transfer across plan rebuilds by composite key inside write_site_plan.
- **sources:** 030#Q1/Q2, #Strategic vs plan-time naming; 031(3)#Lock transfer
- **relations:** B-029-4 design-intent clobber (motivating bug); lock transfer
- **verify-later:** site_plan_directives populated; brief renderer helper

### LLM tiering (large/medium/small/none) + cluster-then-slot-fill scaling pattern
- **category:** model-infrastructure
- **status-signal:** aspirational
- **status-evidence:** 029: "the chassis routes" described as design with llm_tier annotation to add; flip-to-local gated on Thunder health
- **what:** Every LLM call site declares a tier; chassis maps tier→endpoint via flippable config (large=Opus strategy/briefing; medium=Sonnet→local-70B for plan partials/audits; small=Haiku for slot-fills; none=deterministic Go). Product-listing scale: facts from feeds (Go), cluster ~10k products into ~20-50 groups algorithmically, one medium call per cluster for framing, small slot-fill per product.
- **sources:** 029#LLM tier per call site, #Affiliate/product listings
- **relations:** model aliases; batch queue routing
- **verify-later:** any llm_tier config keys in defs

### Per-page brief generation (lazy) and the no-empty-slots acceptance test
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** 029 B-029-2 promotes it to Phase-1 acceptance test; briefs "lazy" design section
- **what:** Component templates are named slots; a per-page brief enumerates slot content, generated lazily at build time. Without briefs, component-author defaults leak (empty img src, /services.html CTAs on sites without services). Acceptance: a Phase-1 build produces no empty slots and no leaked defaults — unbriefed slots either don't render or error before deploy.
- **sources:** 029#B-029-2, #Per-page brief generation
- **relations:** directives; B-029 bug list (dup nav items; theme vars never written)
- **verify-later:** brief generation exists?

### Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods)
- **what:** All storage-touching actions construct their own client via storage.NewS3Client with env-var names in ObjectStorageConfig (B2_APPLICATION_KEY_ID/KEY from personae-platform-secrets); injected params.StorageClient is unreliable (nil when IMAGE_BUCKET absent at startup). Spawn-time env forwarding (Path C) is gated by isStorageEnabledAgent/orchestrator/code-driven — keep the gate maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 full
- **relations:** spawn env propagation; thunder presigned URLs
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

### Styling render pipeline reference: two assembly paths and the scheme gap
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** 036 FINDING/THEORY-tagged reconstruction from code + live data
- **what:** Stylesheet and page-section renders are separate code paths meeting only in the browser via class names/custom properties. Key FINDINGS: resolved_composition doesn't record scheme (survives only on layouts.scheme); buildSectionDefaults emits --section-* only for dark bg/surface (light sites correctly get nothing); five surface classes are duplicated renderer+layouts (Phase 4.5 debt); hero/CTA components hardcode dark backgrounds + literal white text defeating the scheme; .{function}-section class contract broken by hero (.hero) and CTA (.cta-section); four overlapping chrome default stores (style_collections ids [live read], site_components slots [likely superseded], sites.default_components, layouts.default_* [all NULL]); RenderFallbackHeader is hardcoded dark; SectionStyles/component_selector are dead on the current path. Fix direction (Q4a): strictly variable-driven components + renderer-owned per-section --section-*.
- **sources:** 036 full; 016b light-site-dark-chrome entry
- **relations:** section painting contract; site component linkage; scheme-to-components runbook
- **verify-later:** F-thread confirmations (update_site_defaults on composition path)

### Per-tool travelling documentation convention (PLAN_/NOTES_ + taxonomy)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** 037 created 2026-06-29; layer 1 (manual markdown) now, DB tool_docs table "recommended target"
- **what:** Every tool/complex component carries PLAN_<function>.md (aim, source spec, behaviour contract, delivery mechanism Path1/Path2/build-time, dependencies, deliberate decisions) and NOTES_<function>.md (timestamped choices/bugs/dead-ends tagged with a shared problem-category taxonomy: css-variable-mismatch, empty-shell/mode-b-template, broken-template-slots, content-vs-runtime-mismatch, detool-on-rebuild, js-not-extracted, js-bundle-stale, schema-template-drift). Sits below 016/016b and site runbooks; end-state: docs generated automatically at creation and grown per change.
- **sources:** 037_TOOL_DOCS_convention(1).md
- **relations:** tool doc header (in-code anchor); 016b category tags
- **verify-later:** tool_docs table existence

### Blog listing rebuild and slot-detection strategy
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 103 handoff: rewritten + deployed with three bug fixes; 004 shows it in rerender-pages workflow
- **what:** rebuild_blog_listing runs in rerender-pages before get_pages: finds the actual listing slot via priority list → pages.sections → default; loads a template that genuinely has a {{range}} (guard against CSS-only components); ensures article links (SQL template patch + post-render safety net); writes content_data alongside rendered_html; computes read time from component lengths.
- **sources:** 103_blog_nav_handoff-2026-04-12.md; 004#Blog Listing Rebuild
- **relations:** empty_blog check → blog-content-planner
- **verify-later:** rebuild_blog_listing_action.go current form

### Open problem: nav-updater never spawns
- **category:** debugging
- **status-signal:** unknown
- **status-evidence:** 103 "Active Problem" — definition exists/active, topics exist, dispatch generic, yet no pod ever appeared; all nav_drift items claim-timeout
- **what:** nav_drift items route to nav-updater via the generic dynamic dispatch, but no nav-updater pod has ever started and items exhaust claim timeouts. Investigation was open at handoff (2026-04-12); distinct from the nav-link-fixer path.
- **sources:** 103#Active Problem
- **relations:** dispatch loop; missing-handler pattern (different: def exists)
- **verify-later:** whether resolved since; nav_drift item outcomes

### Dispatch failures triage report and the bug/recommendation/gap classification gap (P10)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** 105 v4 (2026-04-15): P1–P6 deployed/done, P7/P8 human-gated, P9 fix written, P10 deferred with design note
- **what:** Systematic queue triage produced ten priority fixes: component_id nil in load_tool; retry-safe fork deploys; rate-limit errors classified transient; load_page_record page_id fallback; plan-then-reconcile for section data requests (auto-close stale); design chain unblocked; audit gap findings rerouted to needs_content_page. P10 names an architectural gap: auditors emit opinions (recommendations) that the pipeline auto-fixes as if bugs — proposed three-way classification (bug/recommendation/gap) with specialist agents + per-site approval mode (~1 week, deferred).
- **sources:** 105 full
- **relations:** write_audit_findings; approval model (P1 plan)
- **verify-later:** P9/P10 status since April

### Anthropic product-knowledge skill (verify, don't recall)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** 106 is the skill file itself
- **what:** A skill instructing agents to consult official Anthropic docs (docs maps per product) rather than memory for any Claude API/Code/claude.ai facts — accuracy over guessing, source everything, distinguish the three products.
- **sources:** 106_claude_anthropic_skill.md
- **relations:** dated-claim verification convention
- **verify-later:** where the skill is installed/used

### build_queue domain queue with direction spectrum
- **category:** onboarding-config
- **status-signal:** partial
- **status-evidence:** P1 (marked "a bit out of date but still has merit"); P2 depends on it for POST /sites; seed_build_queue named in 032/other docs as real
- **what:** build_queue rows (domain, direction jsonb, status, batch, priority); direction spans null → objective hint → full brief (skip research+briefing) → adopt_from → fork_from (specs pre-populated). seed_build_queue takes N, ensures site records, writes initial specs, inserts the appropriate first work item; pacing by batch size. Initial chain: needs_domain_research → needs_briefing → needs_site_plan with spec outputs per handler.
- **sources:** P1#Domain Queue, #Initial Build
- **relations:** public API POST /sites; domain-submitter (newer entry path — reconcile)
- **verify-later:** build_queue table + seed action exist; relation to domain-submitter

### Approval model (auto / hitl / eval) with four override levels
- **category:** hitl
- **status-signal:** aspirational
- **status-evidence:** P1: "Future — column exists from day one" for eval; no deployment evidence
- **what:** site_work_items.approval_mode: auto (default), hitl (pending_review → human approves), eval (evaluation agent reviews handler output before completion). Overrides per-item, per-item-type, per-site, system default.
- **sources:** P1#Approval Model
- **relations:** content-reviewer as future eval agent; P10 recommendation specialists
- **verify-later:** approval_mode column exists?

### Side-effect rules engine (deterministic follow-on items)
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** P1 table of triggers; some live (side_effect source in 002/003 contracts), full Go rules engine unconfirmed
- **what:** After each handler completes, deterministic Go rules (not LLM) emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side effects; 003(8)#Cross-Domain Coordination
- **relations:** cross-domain coordination; snapshots triggers
- **verify-later:** rules engine implementation

### Marketing as work items + OpenClaw adapter
- **category:** NEW:marketing
- **status-signal:** aspirational
- **status-evidence:** P1 marketing section entirely future (agents/adapter unbuilt)
- **what:** SEM campaigns, landing pages, email sequences, social content, schema markup, ad copy all as work items with dedicated handler agents; an openclaw-adapter (adapter service, self-hosted) translates structured campaign specs to external platforms (Google/Meta/LinkedIn) and returns metrics; marketing-discovery-agent finds gaps (GBP, schema, page-2 rankings, competitor ads); SEM setup is HITL-gated.
- **sources:** P1#Marketing: SEM, Outbound, and Growth
- **relations:** work-item system extensibility
- **verify-later:** none built

### Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **category:** NEW:public-api
- **status-signal:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes
- **what:** site_ownership junction table (site/client/user/role) rather than columns on sites (shared sites; 15+ FKs untouched); all public queries scope through it. POST /sites writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2 full
- **relations:** admin API; needs_human_review status; build_queue
- **verify-later:** site_ownership table; which blocks landed

### Admin API current state: dual-auth gateway, inventory, and fix blocks
- **category:** admin-dashboard-and-api
- **status-signal:** partial
- **status-evidence:** P3 headings: known issues (bugs "code won't work as written", hardcoded/mock data, missing wiring) with blocks A–F and a target route map
- **what:** Two services one gateway: auth-service handles auth/user/subscription/projects directly and proxies admin site routes to core-manager; dual auth validation on both sides. The doc inventories every current route, catalogues bugs/mock data/unregistered handlers/design concerns, and sequences fixes: A fix bugs, B wire handlers, C replace hardcoded data, D performance, E new site-domain endpoints, F agent-definition admin improvements.
- **sources:** P3_admin_api_plan.md (header-scan)
- **relations:** admin dashboard; public API plan
- **verify-later:** which blocks completed (012 suggests site-domain endpoints largely live)

### Dynamic applications direction (three tiers; framework specs; thin generated backends)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** 022 tier 1 "now → near term", tiers 2–3 medium/longer term
- **what:** Tier 1 static+dynamic components (forms via external services, client search, client-side A/B); Tier 2 agent-powered per-site backends (workers/lightweight services fed by agents — business logic stays in agents, backend is a thin render layer); Tier 3 full application generation (admin panels, SaaS prototypes). Principles: framework specs stored for each target stack; one site one repo one deployment; generated-vs-human content marked, human edits precedent; incremental complexity (mailto → Formspree → Worker → CRM).
- **sources:** 022 full
- **relations:** infrastructure layers (007); CSS variable contract (shared)
- **verify-later:** none built beyond tier 1 basics

### Link management: link_registry as first-class links + gap to planned links family
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 024: schema + extract/sync + constraints + validation exist; links-orchestrator family "planned but not implemented"; delete-and-reinsert loses validation history (known)
- **what:** Every anchor in rendered HTML lives in link_registry (scope internal/page/external, type navigation/content/semantic, affiliate fields, validation state); extract_and_sync_links parses post-build (delete+reinsert per page); InjectLinkConstraints feeds valid pages into writer prompts to prevent invented links; validateInternalLinks warns (not blocks) on missing targets; nav structure is separate (site_nav_groups/items; populate_nav_tables classifies primary/legal/utility). Planned: link-crawler/validator/registry-sync/redirect-manager/affiliate-manager under an algorithmic links-orchestrator.
- **sources:** 024 full
- **relations:** orphan_pages/internal-linker; phantom-CTA bug; nav agent family
- **verify-later:** link_registry population; HTTP validation anywhere?

### Nav agent family and the three-tier authority model
- **category:** navigation
- **status-signal:** partial
- **status-evidence:** 002(4): owner "currently populate_nav_tables action within pageflow-builder"; tiers described as model
- **what:** Navigation as first-class entity (groups: primary/subsection/content/legal/utility/external; contextual groups planned). Tier 1 strategist authority (new builds), Tier 2 nav-agent autonomous maintenance, Tier 3 drift detection vs original plan. nav-updater/nav-link-fixer handle drift and broken template links today; nav dedup guard recommended after B-029-1 duplicate nav items.
- **sources:** 002(4)#Navigation Agent Family; 024; 029 B-029-1
- **relations:** nav-updater never spawns; populate_nav_tables
- **verify-later:** nav drift check + dedup guard status

### Palette/layout/typography composable-theme migration (025)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 025(3) is the migration plan; 027/036 confirm palettes/layouts/typography_sets live and read by renderer; legacy css_themes columns retained "until Phase 7"; 036 notes Phase 4.5 coupling still present
- **what:** Split css_themes into palettes (colours jsonb open slot map), layouts (css_template + structure_tokens + default header/footer FKs + scheme), typography_sets (fonts+scale), each with the origin/needs_review/fork lineage model; css_themes becomes a composition of three FKs. Motivation: the old library was one layout with 14 palette skins behind a silent standard-brochure fallback. Template data moves to map-based Palette/Typography/Structure (no Go change per new slot). Direct cutover, no shadow mode; selector unchanged in this phase.
- **sources:** 025(3) full; 036 §3
- **relations:** composition stages; layout scheme matcher
- **verify-later:** legacy columns still read anywhere; Phase 4.5/7 progress

### Renderer theme-resolution cascade and the emergency fallback
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 027 §4: theme_name literal cleared "as the cutover moment"; emergency fallback + logger.Error monitoring rule
- **what:** render_css_from_spec resolves theme by config.theme_id → config.theme_name → sites.style_collection_id join (production path); all-miss falls to standard-brochure WITH logger.Error — any emergency-fallback line is a pipeline bug. resolveThemeIDFromSiteContext never errors, warns with a distinguishing reason.
- **sources:** 027#Renderer Changes
- **relations:** install-before-render ordering; B-029-3 theme-vars-not-deployed bug
- **verify-later:** emergency fallback frequency in logs

### 016b durable invariants + wrong-turns log as a debugging methodology
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v1–v5 changelog; wrong-turns section explicitly kept "so the next pass doesn't re-walk them"
- **what:** Vol. 2 distils the paying-off heuristics (trust artefact; completed_at ≠ write instant; config-key-path no-ops; who writes page_components; 0 rows not decisive; negative inference needs mechanism checked in ALL cases; reuse before rebuild) and logs false leads per arc with the heuristic each violates. Also fixes doc process: the guide had forked across chats; v5 is the explicit merge point.
- **sources:** 016b#Orientation, #Durable invariants, #Wrong turns
- **relations:** 016 §0; travelling docs
- **verify-later:** —

### Zero-planned-sections silent no-op success (planning gap + complete_error anti-pattern)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b v5 + 2026-07-06 amendment: route confirmed from workflow dump; pages.sections UPDATE proven to change behaviour; guards/planner-invariant fixes listed as prevention, not applied
- **what:** A linked-everywhere page 404'd for two weeks while seven work items completed clean: planner emitted the page with no sections; page-build-handler's zero-ready branch is literally a complete_workflow step named complete_error ("an error path implemented as a successful completion" — diagnostic signature: result contains only site_record); rerender skips no-component pages quietly. Section sources in order: site_specs site_plan aspect → pages.sections; site_plan_sections table is NOT read by builds. Fixes: planner invariant (every page ≥1 section), fail-loud zero-planned guard, rerender warn, auditor rules (active+linked+planned; post-deploy URL HEAD), dynamic-list component vocabulary for archive pages.
- **sources:** 016b#Page build completes having built nothing + amendment
- **relations:** silent-noop-success/planning-gap tags; section-index vocabulary
- **verify-later:** complete_error branch fixed?; planner invariant added?

### Content↔template key-contract drift (system-stats class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 TRIAGED 2026-06-24, remedy un-applied; fleet-wide (usage_count 22)
- **what:** component-creator rewrote a template AFTER pages were built; stored content_data keys share ZERO keys with the new template placeholders → renders text-empty → visible-content filter correctly drops the section. Remedy: full content rebuild (not page_rerender, which reuses mis-keyed content_data); structural need: component schema changes must trigger dependent rebuilds, or fix writer↔input_schema binding. Diagnostic: diff the two key sets directly (a populated-but-unrendered section is a key-contract check, not a generation failure).
- **sources:** 016b Part 5 + wrong-turn #4
- **relations:** schema-template-drift tag; component regeneration rerender items
- **verify-later:** schema-change-triggers-rebuild mechanism

### Imagery: per-page assets vs site-wide last-write-wins resolution gap
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** 016 §9 deployed-hero-images entry: assets deploy fine; resolution gap diagnosed with fix direction (page-aware ensureAssets + re-render through plan_sections)
- **what:** site_plan_imagery plans per-page keys; store_asset writes content_data[<purpose>_url] keyed by purpose so every page hero overwrites the last (single site-wide hero_url); first render bakes the use_fallback static path; terminal rerender/CSS fixers patch stored HTML without re-resolving. Fix: resolve per-page from site_plan_imagery⋈assets, keep content_data as gap-fill; after an asset lands flag needs_rebuild → needs_page at p99. Logo is chrome (render_site_components) — separate path. imagery kind/scope model + chk_kind constraint implied by site_plan_imagery columns.
- **sources:** 016 §9 hero/logo fallback entry; 002(4) flag_page_image_rebuild → page_rerender
- **relations:** page_rerender image_landed reason; input schema image fields rule
- **verify-later:** ensureAssets page-aware now?; site_plan_imagery schema

### Log tables before pod stdout (agent_error_log, llm_call_log as forensic sources)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016 "hunting for logs" section; pod-rotation checklist item
- **what:** Persistent DB logs beat ephemeral pod stdout: agent_error_log (every reported error, filterable by context site/domain), llm_call_log (every call incl. failures with error_message). Pod logs vanish on rotation/rollout; zap JSON must be grepped by message string not field=value; logger.Debug is invisible in-cluster (house rule: logger.Info); verify deploys against the artifact (curl/DB), not log presence.
- **sources:** 016#hunting for logs; 016b#Verifying a deploy
- **relations:** silent-completion; assumption checklist 3/15
- **verify-later:** —

### SQL template-surgery method (needle-gate) and Postgres verification pitfalls
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry distilled from the scheme-to-components workstreams
- **what:** Safe in-DB template mutation: needle-gate read (LIKE booleans + occurrence counts, expectations counted mechanically not recalled), shell .bak of the column, guarded idempotent exact-string replace (or anchored regexp_replace), RETURNING checks, value-agnostic rollback. Pitfalls: regex quantifier bound ≤255; substring-with-parens returns first capture group; gradient-embedded hexes escape colon-anchored classification; % in needles breaks LIKE gates (use position()).
- **sources:** 016b#SQL verification pitfalls
- **relations:** marker-REPLACE anchoring entry (anchor attribute REPLACEs on the opening tag, not the bare attribute — the querySelector corruption bug)
- **verify-later:** —

### sites.status is informational (never scope by status='active')
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b v4 entry with the silently-dropped-site incident
- **what:** UpdateSiteStatusAction vocabulary draft/building/review/published/deployed/archived/error; 'active' is legacy hand-written; nothing filters on it — dispatch keys on site_work_items. Enumerate GROUP BY status before any blast-radius query. Reuse-gate corollary: check pg_proc/pg_trigger before adding helpers (shared set_updated_at exists).
- **sources:** 016b#sites.status
- **relations:** zero-rows-not-decisive
- **verify-later:** —

### Documentation consolidation system (numbered canonical docs + index)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** 000(2) "Consolidated from 57 source files"; consolidation notes at the head of 001/002/003 recording patch-incorporation and supersessions; 016 v2_58 full-diff consolidation note
- **what:** The docs024 set is the consolidated canonical documentation: one numbered doc per area with an index, consolidation notes stating which patches are already incorporated (and must not be re-applied), version families closed by full heading+content diffs, and continuation volumes when a doc hits size limits (016→016b). "Plans (review for currency)" section separates aspirationals.
- **sources:** 000_documentation_index(2).md; consolidation notes in 001(5)/002(4)/003(8)/016 v2_58
- **relations:** per-tool docs; travelling doc conventions
- **verify-later:** —

## Proposed NEW categories
- `NEW:work-item-system` — the work-item queue/lifecycle/dedup/dispatch mechanics are the platform's spine and cross-cut every pipeline; deserves its own council agent (routing table, pipeline column, two-strike, state machine, terminal items, side-effect rules, approval model).
- `NEW:marketing` — SEM/OpenClaw/marketing-discovery is a distinct planned domain not covered by business-strategy.
- `NEW:public-api` — user-facing API + site_ownership model is distinct from admin-dashboard-and-api.
