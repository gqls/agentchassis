# Register — development-guide

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

87 concepts, consolidated from 113 raw extractions across units U01, U02, U03,
U04, U05, U06, U07, U08, U09, U10, U12, U13, U14, U15, U16, U17a, U17b, U18,
U19, U20, U21, U22, U23, U24c, U24f, U26. Note on the raw cluster input file:
the whole file arrived mechanically double-pasted (every block from unit U01
onward reappeared byte-for-byte a second time later in the file, plus three
new-category blocks were each pasted twice within that duplicated tail); those
exact duplicates were merged trivially and are not counted as "real" duplicates
below. Beyond that mechanical artifact, ~26 raw blocks were genuine
independent-unit duplicates (the same mechanism — e.g. the wrapper-orchestrator
pattern, or the manual kcat-trigger convention — extracted separately by 2-6
different units) and were merged into single entries; those merges are the
main source of the 113→87 reduction.

### DEV-001 — STEP ZERO — reuse-before-create discipline
- **status:** convention
- **status-evidence:** 001(5) opens with it as "the most important step" (asset-deploy-agent 3-hours-wasted example); restated verbatim in the archived docs019/WM 001(0) doc and re-exercised as a live practice throughout the finetuning-flywheel workstream (NOTES(45) 2026-06-04 reuse audit).
- **stage2-verified (2026-07-14):** deployed → convention — reuse-before-create discipline (STEP ZERO), a working practice not a single artifact
- **what:** Before creating any agent/action/function, search agent_definitions, the action registry, Go code, gate functions, and workflows for existing equivalents; document findings; never create without demonstrating no existing coverage. Extends to documentation claims (verify at point of use, date what you verify `[checked YYYY-MM-DD]`) and to functions generally (fetch-first applies to DB functions too, e.g. `\df` before drafting backup machinery). Repeatedly exercised in practice: reused ssh_get_status as a monitor probe, ListObjects for resume, datahelpers.GetIntField over a custom helper, snapshot_agent() instead of a new side-table migration.
- **sources:** 001_development_guide(5).md#Pre-Flight, #API Verification Reference, #Reuse Before Creating; WM/001_development_guide(0).md#pre-flight-does-this-already-exist-step-zero; NOTES_phase5_training_launcher_running(45).md#3(D3),#reuse-audit
- **relations:** canonical field-path helpers; assumed-helper build failures; guidelines audit (001/002/003 compliance)
- **verify-later:** platform/orchestration/actions/registry.go; agent_definitions table

### DEV-002 — Canonical field-path resolution helpers (datahelpers) vs duplicated resolvers
- **status:** partial
- **status-evidence:** 001(5): "18+ functions that resolve dot-paths… Do not add another one"; cleanup of 9+ duplicates listed under "Not yet built"; independently reconfirmed as "the single biggest code hygiene issue" with ~18 near-duplicate helpers enumerated.
- **what:** `ExtractNestedField/String/Map`, `GetFieldFromPath(WithDefault)` in datahelpers are canonical (with `.response` auto-unwrap); six named duplicates in the actions package must not be copied, and there is no `ExtractStringSlice` — compose `ExtractStringListHelper(ExtractNestedField(...))`. The ~18 near-duplicate resolvers (resolveFieldPath, ExtractNestedField, GetFieldFromPath, etc.) differ subtly in arg order, logging, and `.response` unwrapping; standing rule is reuse datahelpers before adding a new one. Related recurring bug: Go-template paths need leading dots (`{{.x.y}}`) and input_mappers are compulsory.
- **sources:** 001_development_guide(5).md#Field Path Resolution; 016_additions_assumed_helper_and_cross_module.md; docs020.../001_rag_agent_distribution_architecture.md#field-path; docs019_business/004_vet_practice_verifier.sql#go-template-fixes
- **relations:** STEP ZERO; assumption checklist item on assumed helpers; rag_actions.go helper cleanup
- **verify-later:** platform/orchestration/datahelpers/*.go; NullableString/TruncateString/NullableInt in datahelpers

### DEV-003 — Actions are the unit of work — no wrapper+core split
- **status:** deployed
- **status-evidence:** 001(5) "The wrong pattern" section with WriteSiteSpec example.
- **what:** All action logic lives inside the `XxxAction` function; composition happens via workflows, not Go-calling-Go; exporting a "core logic" function creates a duplicate API surface. Corollary: don't create subworkflows in SQL — spawn sub-agents instead.
- **sources:** 001_development_guide(5).md#Core Design Principles
- **relations:** every-agent-is-an-orchestrator; wrapper-orchestrator pattern
- **verify-later:** grep exported non-Action functions in actions package

### DEV-004 — spawn→call pattern and role-based targeting
- **status:** deployed
- **status-evidence:** 001(5): "This is how every existing workflow does it"; earlier documented (README.041/README.004) as CallAgentAction's core mechanism, "used in all robot-hands workflows."
- **what:** Agents are spawned (`spawn_agent`) then called (`call_agent`), which finds the previously spawned agent by searching CollectedData spawn results for a matching `target_role` (preferred — findAgentByRole scans all collected_data keys) or `agent_type` (findAgentByType, a trap — only scans keys starting `spawn_`). Role is the within-orchestration nickname distinguishing multiple agents of the same type; dynamic dispatch = fixed role + `agent_type_field` resolved from collected_data at runtime, never a topic-construction bypass.
- **sources:** 001_development_guide(5).md#How call_agent finds the spawned agent, #Dynamic dispatch; 002(4)#Resolved Decisions 16; docs001_flow_general/README.041.role_flow.md; docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** dispatch loop; wrapper-orchestrator pattern; spawn_agent DB-driven Job spawning
- **verify-later:** spawn_actions.go, call_agent.go findAgentByRole/findTargetAgent

### DEV-005 — Wrapper-orchestrator pattern ("every pod-running agent needs a parent that spawned it")
- **status:** deployed
- **status-evidence:** Confirmed working repeatedly and independently across at least six workstreams: 001(5) full section with med-* wrappers and site-adoption-orchestrator; "Spawning confirmed working (2026-04-23 trigger test)" (FOCUS finetuning); site-adoption-agent's 2026-04-22 wrapper; FOCUS(25) "Spawning architecture — fully confirmed working" with observed worker pods; vet-med pricing orchestrators; and generalised as "every agent is an orchestrator" (2026-06-17).
- **what:** Agents get dedicated K8s Job pods only via spawn_agent from a parent; anything reached via the generic entry point that does substantive work (LLM calls, crawls, heavy DB, minutes of runtime) needs a tiny wrapper — spawn_agent → call_agent(target_role, not agent_type) → complete_workflow — so real work runs in its own pod with clean logs, while in-chassis work would block one of the three shared chassis pod slots. The concrete mechanism: `processing_mode:"orchestrator"` at the TOP level of default_config plus `agent_category='coordinator'` (category is free-text 'orchestrator') is the combination that produces a dedicated spawned pod; the worker uses `processing_mode:"task"`/specialist. Canonical minimal forms: med-export-orchestrator, training-data-export-orchestrator, site-adoption-orchestrator. Input mapping must map fields individually with `?`-suffixed optionals — never a whole `input_data` blob (file writes from non-spawned actions land on a random chassis pod and die with it). Known sibling defect: all four med-* wrappers carry the broken `{"input_data": "input_data"}` double-wrap mapping (see the whole-blob input_data anti-pattern entry).
- **sources:** 001_development_guide(5).md#Every pod-running agent needs a parent; FOCUS_finetuning_flywheel_and_service(13).md#2.4f,#2.4h,#14; FOCUS_finetuning_flywheel_and_service(25).md#2.4f-2.4i; FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed; docs/agent_docs/sql_for_tables/034_vet_med_price_scrape_orchestrator.sql; NOTES_running_synthesis_v2(36).md 2026-06-17
- **relations:** agent_definitions three-column semantics; spawn→call pattern; whole-blob input_data anti-pattern; confirmed chassis workflow model
- **verify-later:** agent_definitions rows for med-*/site-adoption-orchestrator; spawnAgentKubernetesJobFromDefinition; spawn decision logic re: processing_mode placement

### DEV-006 — Standardized input extraction (input_mapping / input_fields / ActionInputSpec, `?` optional suffix)
- **status:** deployed
- **status-evidence:** 001(5) three-layer table with resolver behaviour documented from code; same doctrine present verbatim in the archived WM/001(0) predecessor and in docs017/008's "No more boilerplate — 40+ lines of extraction code per action eliminated."
- **what:** Three layers move data into an action: caller `input_mapping` (dot-paths into collected_data, with `?`-suffixed destination keys marking a mapping optional — silently skipped if the source path is missing, vs. unsuffixed fields which hard-fail the call), action `input_fields`/ActionInputSpec (Required/Optional/Defaults/Deprecated), and `ExtractActionInputs(spec)` which tries input_fields, falls back to deprecated `*_field` keys with warnings, and checks nested parents. In the dispatch loop specifically, only site_id/domain/work_item_id may be non-optional; all `spec.*` mappings must use `?`. A real production bug (the "site plan contamination bug"): ExtractActionInputs found `site_record.content_data` via nested lookup and overwrote the hero section with the site plan.
- **sources:** 001_development_guide(5).md#Standardized Input Extraction, #Optional fields in dispatch loop; WM/001_development_guide(0).md#standardized-input-extraction; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Decision-Standardized-Input-Extraction
- **relations:** field-name collision via nested-source loop; handler input path contract; input_mapping semantics (call_agent-only correction); chassis action input conventions (dual registration)
- **verify-later:** ResolveInputMapping in coordinator.go; datahelpers.ActionInputSpec/ExtractActionInputs/RegisterActionInputSpec

### DEV-007 — Field-name collision via the nested-source loop (affects required AND optional fields)
- **status:** deployed
- **status-evidence:** 001(5) corrected wording explicitly supersedes an earlier draft that claimed the collision affects only optional fields: "This loop iterates the full field list — both Required and Optional. It does not distinguish between them." Backed by a real section-editor content_data clobber.
- **what:** ExtractActionInputs's late nested-source loop checks `current_page`, `rerender_pages`, `site_record`, `input_data` for any unresolved field name — so generic names (`content_data`, `sections`, `site_id`, `domain`, `status`…) can silently bind to the wrong source, and this risk applies to required fields too (e.g. `site_id`), just usually masked because earlier resolution strategies (0-2) resolve them first. Rule: new code avoids colliding names (prefix them, e.g. `target_site_id`); existing code is left alone unless it actually misbehaves; complex/array fields must never go in ActionInputSpec (read the config path directly).
- **sources:** 001_development_guide(5).md#Field name collisions; 016 §0 item 15 and §9 literal-key trap; old/older1/001h_development_guide_new_agents_v8.md#"Field Name Collisions"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Note on the target_site_id input field name"
- **relations:** section-editor clobber; resolve_internal_links review catch; ExtractActionInputs / datahelpers resolution cascade (Strategy 0-4); whole-blob input_data anti-pattern
- **verify-later:** datahelpers/action_inputs.go nestedSources — confirm the loop still doesn't distinguish Required/Optional

### DEV-008 — RAG actions and knowledge_base shared store
- **status:** partial
- **status-evidence:** 001(5): migration 082 applied, "Not yet populated"; rag actions Go "produced but not deployed", then later "registered, not workflow-tested."
- **what:** `rag_lookup` (embed→vector search→top-k, trigram fallback) and `rag_index` (chunk→embed→store, SHA256 dedup) operate over a shared `knowledge_base` table (vector(768)); deliberately actions, not agents, until a knowledge-indexer orchestration is needed. Tool docs also target a knowledge_base `tool_docs` row.
- **sources:** 001(5)#RAG actions, #Agent vs infrastructure; 037#Where the docs live
- **relations:** agent-vs-infrastructure boundary test; per-tool docs convention; RAG knowledge_base (rag-knowledge-base register, same underlying table)
- **verify-later:** rag_actions.go registered?; knowledge_base row counts

### DEV-009 — Agent vs infrastructure boundary test
- **status:** convention
- **status-evidence:** 001(5) table (LLM logger no, Ollama provider no, rag actions no, knowledge-indexer future yes).
- **stage2-verified (2026-07-14):** deployed → convention — agent-vs-infrastructure design test/heuristic, not an artifact
- **what:** Something becomes an agent only if it owns a domain, needs its own workflow, and benefits from independent spawn/debug. Otherwise it is an action or cross-cutting infrastructure.
- **sources:** 001_development_guide(5).md#Agent vs infrastructure
- **relations:** RAG actions and knowledge_base; promotion pattern
- **verify-later:** —

### DEV-010 — Specialist vs handler: the persistence boundary
- **status:** deployed
- **status-evidence:** 001(5) post-mortem; hit twice (page-content-writer HTML trapped in site_work_items.result).
- **what:** A specialist returns data to its caller; a dispatch handler must persist its own outputs (page_components, site_components, assets) and update status. Specialists used as handlers need a wrapper (page-build-handler wraps page-content-writer with plan/validate/save/deploy). Test: callable from CLI with site_id+domain and everything lands in the right tables.
- **sources:** 001(5)#Lessons Learned; 002(4)#Page Build Handler Pipeline
- **relations:** dispatch loop; handler contract; specialist agent design doctrine
- **verify-later:** page-build-handler definition; handler agents' save steps

### DEV-011 — Extended thinking config and the no-temperature-to-Anthropic rule
- **status:** deployed
- **status-evidence:** 001(5) dated 2026-05-27 temperature note.
- **what:** `budget_tokens` in ai_service enables extended thinking (thinking blocks skipped in parsing, +30-90s). Since 2026-05-27 the Anthropic client sends no temperature at all (Opus 4.7+ 400s on non-default; thinking is incompatible with it); Ollama still honours temperature. Steer Anthropic behaviour via budget_tokens and prompts instead.
- **sources:** 001(5)#Extended Thinking Configuration
- **relations:** LLM config shadowing (temperature dead paths)
- **verify-later:** anthropic.go client options

### DEV-012 — Loop mechanisms: dynamic workflow expansion
- **status:** deployed
- **status-evidence:** 001(5) Appendix C is the full, complete reference (incl. production dispatch-loop example); this is the fully worked-out successor to earlier partial internal investigation notes (`LoopAction`/`handleLoopExpansion`/`setLoopVariable`) which are explicitly superseded once absorbed.
- **what:** Loops inject N×M steps into the workflow plan at runtime (`{loop}_iter_{N}_{substep}`) via a coordinator-side expansion handler, chained by NextStep; `setLoopVariable` propagates the iteration item into CollectedData under the loop var before each step and also back to base names, with per-iteration output suffixing, `continue_on_error` skipping, and a `loop_complete` step (LoopCompleteAction) aggregating results into output_field. Chosen over in-process iteration because steps can await async agent responses and survive crashes/restarts as ordinary persisted workflow steps. Known hazards: the fast-response race fixed by the ErrLoopExpansionHandled sentinel; a shared `loop_metadata` key; never nest loops — spawn a sub-agent instead.
- **sources:** 001(5)#Appendix C — Loop Mechanisms; docs010_multitrack_flows_persona_architecture/021_loop_action_discussion.md; docs010_multitrack_flows_persona_architecture/022_loop_actions_guide.md; archive_april_26/014_loop_mechanisms_guide.md
- **relations:** dispatch loop pattern (loop-action dispatch / migration 071); O(K²) state-bloat failure
- **verify-later:** loop_actions.go, loop_expansion_handler.go, loop_complete_action.go

### DEV-013 — Authoring rules pack (20-rule bundle)
- **status:** deployed
- **status-evidence:** 001(5) "Summary of rules" 1–20, each backed by a dated bug.
- **what:** The distilled 20-rule authoring discipline from the bug tally: `\d` live schema before SQL (dumps go stale; domain→pipeline, version_note→change_description renames); `$1`+params never `{{.field}}` in SQL; every LLM step needs api_key_env_var; `{{if}}` before `{{range}}` (query_database empty result is null, not []); run agent-def SQL before triggering (chassis silently runs an empty workflow otherwise); strip markdown code fences before JSON parse; error_step only works inside step.Config; write_site_spec rejects scalars (wrap as `{"text": …}`); use `to_jsonb('…'::text)`; verify fire-and-forget INSERTs actually land.
- **sources:** 001(5)#Appendix A + #Summary of rules
- **relations:** debugging assumption checklist; best-effort-needs-monitoring; error_step mechanics
- **verify-later:** —

### DEV-014 — agent_definitions three-column semantics (category / agent_category / status)
- **status:** deployed
- **status-evidence:** "Caught three agent_definitions column semantics confusions" (2026-04-23); reference row improvement-loop = category=orchestrator, agent_category=coordinator, status=experimental.
- **what:** `category` is a free-text functional role; `agent_category` is CHECK-constrained to strategist/executor/analyst/integrator/coordinator/specialist (NOT 'orchestrator' — a recurring naive-write trap); `status` is lifecycle. Also: ON CONFLICT must target (type, version) `WHERE deleted_at IS NULL`.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4h, #14
- **relations:** wrapper-orchestrator pattern
- **verify-later:** CHECK constraint on agent_definitions.agent_category

### DEV-015 — Chassis action input conventions (dual registration: ActionInputSpec + GlobalActionRegistry)
- **status:** deployed
- **status-evidence:** "Rewritten to use the canonical pattern … matches every other action in the codebase" (2026-04-23); registry gap bit on 2026-04-20 when composition actions "had NO entry in GlobalActionRegistry … rejected as 'requires a topic'."
- **what:** New actions use `datahelpers.RegisterActionInputSpec` in init() plus `ExtractActionInputs` (5-strategy cascade) rather than raw ExtractNestedFieldString; parameters flow via `CollectedData["input_data"]` because `{{.input_data.X}}` templating does NOT render for deterministic-action step config. Every new action needs BOTH the InputSpec registration AND a GlobalActionRegistry entry with `IsLocal:true`; results land in collected_data under `output_field`, never `final_result`. Config literal numbers must be read with `datahelpers.GetIntField(params.StepConfig.Config, …)` — `inputs.GetInt` reads collectedData, not config.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#14; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#2; HANDOFF_2026-04-17_triage_and_component_linking.md#1
- **relations:** standardized input extraction; field-name collision risk
- **verify-later:** datahelpers package; registry.go

### DEV-016 — pgbouncer per-batch transaction discipline
- **status:** deployed
- **status-evidence:** v3.0 "driver: bad connection" failure on a 500-row bulk insert → v3.1 split into per-batch transactions (batch 500→100), confirmed twice independently across two workstreams (2026-04-23).
- **what:** Long-held transactions through pgbouncer (transaction pool mode) are fragile — bulk work must commit per small batch (under a second each, e.g. batch size 100), never wrapping a streaming job in one transaction, with single-statement non-tx bookends. Companion rule from the same incidents: always check RowsAffected() on single-row UPDATEs and fail/error loudly rather than Warn+continue — an action can return perfect counts while its final UPDATE silently didn't land.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#2.4i,#14; FOCUS_finetuning_flywheel_and_service(25).md#2.4i,#14; HANDOFF_2026-04-23_flywheel_A_done_C_scripted.md#lessons
- **relations:** training-data export v3 evolution
- **verify-later:** pgbouncer pool mode config; training_data_export.go batch logic

### DEV-017 — Agent re-registration vs re-seed risk (DB row authoritative)
- **status:** deployed
- **status-evidence:** RUNBOOK 07-06 night: "deploys bump `agent_definitions.updated_at` without overwriting `default_config`… the user confirms component-creator is a dynamically spawned/registered agent, not a YAML-seeded one. So the DB row is authoritative."
- **what:** A durability model for DB-edited agent prompts: chassis deploys re-register agent definitions (bumping `updated_at`) but do not overwrite `default_config`, so SQL edits to prompts survive deploys for dynamically registered agents. The residual risk is an in-code prompt template driving an upsert; the check is a grep for a literal fragment of the OLD prompt in Go sources — a hit means mirroring the edit in code, no hit means nothing can revert the row. An earlier, heavier "seed check" over configs/deployments YAML was superseded by this confirmation.
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-C; RUNBOOK_scheme_to_components(49).md; running_notes_scheme_to_components(55).md#Uf #Uj
- **relations:** component-creator prompt re-aim; idempotent prompt migration pattern; prompt/workflow-jsonb migration convention
- **verify-later:** run the Step C grep; agent registration code path (upsert semantics on default_config)

### DEV-018 — Work-item manual-crafting discipline (real shapes, truthful provenance, never-guess)
- **status:** convention
- **status-evidence:** w4b_04: "crafted from the real rows … with truthful deviations noted: source 'manual' and created_by 'w4b_chrome_refresh' … lying in provenance columns costs later debugging" — pattern repeated across every W6/W7/W8/W9 insert; independently restated as "NEVER guess a needs_page spec" and as the schema captured 2026-07-01 for manual rerender/needs_page recipes.
- **stage2-verified (2026-07-14):** deployed → convention — manual work-item crafting discipline, a process not a code artifact
- **what:** A cluster of verify-at-point-of-use rules for hand-inserting `site_work_items`: copy the metadata of real rows produced by the owning code path (pipeline/severity/priority/handler_agent/status), deviate only truthfully in provenance columns (source=manual, created_by=<script name>), carry only spec fields the consuming workflow actually reads, and dedup check-first with a NOT EXISTS that mirrors `idx_swi_dedup` exactly (non-terminal statuses only, including 'unresolved'). URLs/paths come from `pages.url`, never invented (the phantom-CTA bug was an invented `/contact.html`). Manual page items require the FULL spec — page_id as a real UUID inline, domain, filename, page_name; placeholder strings get claimed and fail ("invalid UUID length"), and any fix must filter on the PLACEHOLDER string, not the intended value. Item_key families are stable conventions: `page_rerender:<page>`, `chrome_refresh_rerender:<site_id>`, `needs_imagery:section:<scope_ref>:<key>`, `component_regen_rerender:<uuid>`, `section_data_*`. Trigger flows through the real producer path where one exists, rather than a manual insert.
- **sources:** w4b_04_trigger_item.sql; w7b_01_imagery.sql; w8_01_post_deploy_rebuild.sql; NOTES(43).md §9l, §9ae, §9w–§9y; docs/RUNBOOK_vonc_migrations(14).md#reference-manual-rerender; docs/RUNNING_NOTES_vonc(36).md#work-item-fix-2026-06-24
- **relations:** work-item dedup mechanics; item_key canonicalization; F2 discriminators; link-management (phantom links)
- **verify-later:** idx_swi_dedup definition; site_work_items status vocabulary; \d site_work_items

### DEV-019 — Standing session/working-contract rules (house rules)
- **status:** convention
- **status-evidence:** Repeated verbatim at the top of multiple running-notes journals ("Standing preferences (STRICT)") and restated independently in a different workstream as "Standing instruction from the user, in force."
- **stage2-verified (2026-07-14):** deployed → convention — file itself notes verify-later 'n/a (convention, not code)'
- **what:** The user's cross-thread working contract, treated as binding by every agent session regardless of workstream: Go not Python; British English; plain language, no hype/flattery, banned words "perfect/critical/excellent", no congratulation; confirm live schema/data before asserting or writing SQL (`\d` before SELECT/UPDATE); structural framework fixes over one-off patches; low risk appetite, reasonable step sizes, ≤1 question per reply; no summary documents unless asked; don't call fixes final; no `*-light`/`*-dark` component variants; keep runbook + journal current; never treat 0 rows as decisive; verify against deployed artifacts not pod logs; flag variable/signature changes; honest caveats including correcting one's own reads.
- **sources:** running_notes_scheme_to_components(55).md#Standing-preferences; HANDOFF_idea_uk_differentiators_section_data.md#House-rules; docs/RUNNING_NOTES_vonc_v2(28).md#standing-rules; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§1
- **relations:** running-notes journal discipline; guidelines audit (001/002/003 compliance); schema-before-SQL discipline
- **verify-later:** n/a (convention, not code) — check for a canonical repo home for these rules

### DEV-020 — Launch idioms: orchestrate vs work-item insert (and what each trigger does NOT do)
- **status:** deployed
- **status-evidence:** "Confirmed from the production trigger scripts" (2026-06-20); the 081c finding (no hand-rolled wrappers) cited.
- **what:** Two production ways work starts: (1) static agents are orchestrated by producing one Kafka message to `system.agent.generic.requests` (action=orchestrate, config.agent_type, full header set) via a one-off kcat pod; (2) dynamic handlers (page-build-handler etc.) cannot be orchestrated directly — INSERT a `site_work_items` row (status='triaged') and the running build-dispatch-loop claims and spawns them. Key caveat: content triggers (rerender-pages / page-rerender / page-rebuild) never re-resolve composition — palette changes must go through needs_composition/needs_design. Deploy topology is likewise two-path: Go changes ship in the chassis image (roll agents to the tag), site HTML ships via the sites monorepo → Actions → B2.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Launch idioms); idea.uk/running_notes_checkpoint_tt.md; idea.uk/HANDOFF(13).md
- **relations:** composition re-resolve; scheduler/dispatch loop; manual agent trigger via kcat orchestrate
- **verify-later:** 082_submit_domain_unified.sh; build-dispatch-loop definition

### DEV-021 — Array-item field contract for the page-content-writer (item_fields fix)
- **status:** partial
- **status-evidence:** "Prompt migration is already applied… But until a chassis image carrying the Go change is live, {{if .item_fields}} is always false… the applied prompt is inert on its own" (checkpoint tt, 2026-06-21).
- **what:** Root-cause class behind empty rendered sections (e.g. 7 blank differentiator cards): the writer's prompt listed an array field with its type but never its per-element shape, so the model guessed item keys (title/body) that rendered empty against templates reading name/description. Fix has three coupled parts: plan_sections populates `ItemFields` on each llm_field_spec (Go); the prompt migration renders the exact per-item field list in both What-To-Write and the JSON skeleton (idempotent by sentinel, safe under either deploy order); a render-time reconciler in v3_site_actions.go. Deploy order matters: chassis image first, then trigger.
- **sources:** idea.uk/019_pcw_prompt_item_fields.sql; idea.uk/running_notes_checkpoint_tt.md; idea.uk/README_assemble_bundle_idea_missing_sections.md
- **relations:** section-data reconciler; coordinator contract (sibling contract-mismatch class); the seam rule
- **verify-later:** plan_sections_action.go ItemFields; whether the chassis tag carrying it shipped

### DEV-022 — Sub-agent modelling conventions (agent_definitions row shape)
- **status:** deployed
- **status-evidence:** running_notes_17(21) "Agent-modeling facts (from research-agent row + 003)"; internal_link_resolver_agent.sql embodies them.
- **what:** How a called sub-agent is modelled: workflow lives inside `default_config.workflow` (agent_definitions has no processing_mode column — it lives inside default_config alongside timeout_seconds); agent_category specialist; input/output contracts required; templated topics (system.agent.{type}.process etc.); responds on the parent's responses topic; NOT-EXISTS-guarded idempotent seed SQL; image_repository/image_tag pinned to the batch image. Modelled on research-agent as the proven sibling.
- **sources:** internal_link_resolver_agent.sql; running_notes_17(21).md#agent-modeling-facts
- **relations:** workflow lives in default_config; agent_definitions registry; every-agent-is-an-orchestrator doctrine
- **verify-later:** research-agent/internal-link-resolver rows side by side

### DEV-023 — input_mapping semantics: call_agent-only; local-action config dot-paths; loop key_path
- **status:** deployed
- **status-evidence:** NOTES(45) §2 "[verified-source] Local action steps do not resolve input_mapping… input_mapping is dead config"; 109b header "CORRECTS a load-bearing assumption: input_mapping is NOT live for (local-action) loop substeps."
- **what:** A precise correction of the input_mapping doctrine: the coordinator honours `input_mapping` only for call_agent (building child input_data) and loop fan-out; on plain local action steps it is dead config. Local actions pull values via config keys whose values are dot-paths resolved from collected_data (`ExtractActionInputs` Strategy 0 / `resolveTemplateToken`); loop substeps read the iteration item via a config dot-path like `key_path:"ckpt_key"` (setLoopVariable puts the item in CollectedData) — using input_mapping there silently falls through to fallbacks (the dataset-key-presigned-40× bug). A proposed coordinator change to resolve input_mapping on local steps was deliberately withdrawn: fix the caller, don't teach the framework a new behaviour for one agent's misuse. Optional mapping fields still take a `?` suffix; missing required sources hard-fail.
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D1,D2),#8; working/phase5/109b_fix_presign_one_loop_item_keypath.sql; working/phase5/103_call_data_preparer_optional_inputs.sql
- **relations:** standardized input extraction; output_fields (plural) contract; loop_complete convention
- **verify-later:** coordinator ResolveInputMapping; input_mapping.go `?` semantics L101-128

### DEV-024 — Child-result shaping: output_fields (plural) contract
- **status:** deployed
- **status-evidence:** NOTES(45) 2026-06-03: "extractWorkflowResult reads completeStep.Config['output_fields'] — PLURAL only… singular output_field… is never read"; migration 104 confirmed live.
- **what:** An agent's final result shape is governed by its complete step's `output_fields` array; the singular `output_field` spelling is silently ignored, producing a step-name-keyed fallback dump that breaks consumers' documented paths (e.g. `provisioning_result.provisioning_id` buried under `dispatch_provision.response.…`). The resolver auto-unwraps one `.response` per path part but never crosses arbitrary step-name keys. Fix taken at the definition level (gpu-provisioner switched to plural + launcher mapping repointed) after the user vetoed a chassis-side change. Corollary rule: verify each call_* step's mapped source paths against the producer's REAL collected_data shape before firing anything expensive (e.g. booking a GPU).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#update-2026-06-03-173x—180x; working/phase5/104_provisioner_output_fields_and_launcher_mapping.sql
- **relations:** input_mapping semantics; output_field/input_fields group-memory contract (earlier-era analogue); data-path verification runbook step 2b
- **verify-later:** extractWorkflowResult in coordinator; whether other defs still use singular output_field (thunder-reaper was named)

### DEV-025 — Reply-topic derivation rules (own topic vs parent topic; two-level await)
- **status:** deployed
- **status-evidence:** NOTES(45) D4 "CONFIRMED live" (2026-06-02); STATUS 06_04 reply-topic orphan fix "VERIFIED 2026-06-04 18:21".
- **what:** Two awaits, two topics: a child's intermediate adapter calls are awaited by the child's OWN coordinator on `ExecutionContext.ResponsesTopic` (seeded from `__my_responses_topic__`); only the child→parent final notification uses `__parent_responses_topic__`. Dispatch actions that put the parent topic in an adapter envelope orphan the await (adapter replies where no one listens → infinite hang) — this bit twice. Verify against code, not an inherited handoff, which asserted the opposite convention. A shared `resolveAwaitResponsesTopic` helper is flagged as a future consolidation; a latent fallback caveat remains (the `system.agent.<type>.responses` fallback doesn't match some agents' actual configured responses topic).
- **sources:** working/phase5/NOTES_phase5_training_launcher_running(45).md#2,#3(D4),#6,#10; working/flywheel_docs/STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** send-before-register race; monitor build; adapter header tiers
- **verify-later:** determineResponsesTopic priority order in coordinator; whether the shared helper was built

### DEV-026 — F6: dedup status-list mismatch and itemsCreated overcount (open defect)
- **status:** aspirational
- **status-evidence:** "F6 flagged: the store's NOT EXISTS guard… omit 'unresolved' → Go guard STRICTER than index" — unfixed, parked as a flag.
- **what:** Two small aligned defects in the work-item dedup path: (1) the store's NOT EXISTS dedup status list and the `idx_swi_dedup` partial-unique index predicate disagree on `unresolved` (index treats it as terminal but the Go guard blocks on it — an unresolved squatter blocks createRerenderWorkItem where the index would not); (2) create_rerender_items increments `itemsCreated` without gating on RowsAffected, so ON CONFLICT DO NOTHING conflicts overcount the log. Both are one-line alignments, parked rather than fixed.
- **sources:** NOTES(43).md §9t, §9aa; RUNBOOK(49).md Part E F6; HANDOFF(7).md §Flags
- **relations:** work-item dedup mechanics; hygiene (40 stale unresolved items)
- **verify-later:** idx_swi_dedup definition vs createRerenderWorkItem NOT EXISTS list; create_rerender_items counter

### DEV-027 — Work-item dedup mechanics (idx_swi_dedup, suppression window, two-strike escalation)
- **status:** deployed
- **status-evidence:** Captured directly from pg_indexes: "idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (complete, verified, rejected, wont_fix, failed, unresolved)"; independently confirmed via insertWorkItem source read (line 46319): "a built-in two-strike rule — an item_key with ≥2 terminal attempts in 7 days is inserted as unresolved… anything <3h after a terminal item is suppressed."
- **what:** site_work_items dedup is two-layered: producers guard with a NOT EXISTS check over open statuses, and the DB independently enforces a partial unique index on (site_id, item_key) over non-terminal statuses. Terminal-status items free the key (why completed triggers can be re-inserted for retriggers). Platform-wide anti-churn machinery on insertion additionally applies a 3-hour suppression window after a terminal item and a two-strike escalation to `unresolved` for repeat-failing keys; `needs_human_review` is non-terminal (holds the dedup slot); terminal set = complete/failed/verified/rejected/wont_fix/unresolved; use DELETE+INSERT, not ON CONFLICT, against the partial index. Mirroring the producer's exact insert (columns, item_key scheme, dedup clause) is the established way to hand-create conforming items. See F6 for a known guard/index mismatch in this mechanism.
- **sources:** NOTES(43).md §9q, §9aa; RUNBOOK(49).md Part C Step 9b; w4b_03_read_rerender_config.sql; running_notes_15(10)#part-8; HANDOFF_2026-06-09#key-references; FOCUS_directory_builder_and_list_components.md#schema-quirks
- **relations:** F6 dedup status-list mismatch; work-item manual-crafting discipline; item_key canonicalization; sectionless durability S1
- **verify-later:** idx_swi_dedup definition; createRerenderWorkItem/insertWorkItem insert shape in chassis

### DEV-028 — Deploy-ordering hard gate for coupled Go action + workflow-config changes
- **status:** deployed
- **status-evidence:** "LESSON (runbook gate tightened): 'deploy the Go action first' is insufficient — the migration is live instantly while the image rolls out. Hard gate: confirm… registered + live on ALL pods, THEN apply the migration."
- **what:** Workflow jsonb changes take effect immediately; Go actions only exist once the image is rolled out and the registry entry (IsLocal:true) is in the running build. Wiring a workflow step to a not-yet-live action makes the validator reject EVERY run of that agent (a real incident broke all component generation). The codified gate: deploy + confirm the action responds on all pods before applying the (idempotent) migration; `revert_agent('<type>')` is the immediate mitigation.
- **sources:** NOTES(43).md §9i, §9j; F1prompt_component_creator_preserve_field_names(1).sql PREREQUISITE header
- **relations:** prompt/workflow-jsonb migration convention; snapshot/revert_agent
- **verify-later:** workflow validator is_local check; revert_agent/snapshot_agent functions

### DEV-029 — Prompt/workflow-jsonb migration convention (snapshot-first, anchored, idempotent, drift-checked)
- **status:** deployed
- **status-evidence:** Implemented end-to-end and named explicitly as "Convention: snapshot-first, idempotent, drift-checked, live-row only"; independently restated as a "STANDING RULE (2026-07-06)" in a different workstream, with every subsequent migration in that workstream carrying the `snapshot_agent()` call.
- **what:** Agent behaviour lives in default_config jsonb; edits are live instantly, so they follow a strict convention: `SELECT snapshot_agent('<type>', '<migration>: pre-update')` first (the function already existed — a reuse win); anchor the edit on a unique existing string and abort if the anchor count ≠ 1 (drift check); an idempotency marker so re-runs no-op; filter to the live row (is_active, not snapshot, not deleted). The "072 nested-prompt trap": `prompt_template` may live at the top level of default_config OR nested in a step config — verify the path first or the migration is a silent no-op. Snapshots live in a separate store (agent_definitions_backup), so a defensive `is_snapshot` selector predicate is not load-bearing; companion `revert_agent('<type>')` exists for rollback.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c, §9d, §9k; RUNBOOK_travelling_docs(38).md#§0-REF,#task-1; RUNNING_NOTES_travelling_docs(39).md#rev18,#rev19,#rev22
- **relations:** deploy-ordering hard gate; correct-while-touching norm; agent re-registration vs re-seed risk
- **verify-later:** snapshot_agent/revert_agent function definitions; agent_definitions_backup table; component-creator prompt state

### DEV-030 — Correct-while-touching norm (bounded repair of adjacent inert bugs)
- **status:** convention
- **status-evidence:** Defined in a RUNBOOK mini-glossary as "Norm adopted in this chat, 2026-07-06"; exercised across migrations 125–146.
- **stage2-verified (2026-07-14):** deployed → convention — a norm/discipline, not a code artifact
- **what:** When a migration already modifies a workflow, it also fixes known-inert bugs in that SAME workflow (e.g. step-level `error_step` moved into config with original targets preserved, dead keys deleted), declared explicitly in the migration file — bounded repair, no separate campaign, and never copying the broken shape into new steps.
- **sources:** RUNBOOK_travelling_docs(38).md#mini-glossary; RUNNING_NOTES_travelling_docs(39).md#rev23,#rev26,#tier-4-continuous
- **relations:** error_step mechanics; snapshot-first migration convention
- **verify-later:** the declared correct-while-touching sections inside migrations 125–146

### DEV-031 — error_step mechanics (config-level placement, existing target, derive-from-next_step, loop corollary)
- **status:** deployed
- **status-evidence:** Live-validated ×5 in one run; mechanism documented in the 001 dev guide §16; independently rediscovered and restated as a debugging-guide "gotcha" in a different workstream ("error_step goes INSIDE a step's config — step-level error_step is silently ignored").
- **what:** The coordinator reads only `step.Config["error_step"]` — a step-LEVEL error_step (a sibling key, not nested in config) is parsed but silently ignored. Once placed correctly, the target must name an EXISTING step or the coordinator fails the whole workflow (a typo converts a recoverable failure into a fatal one). Pattern: derive `error_step` from the step's own `next_step` read from the same row (convergence by construction, nothing guessed); `jsonb_set` does not create parents — COALESCE-merge config instead. Loop corollary: inside loop substeps, `error_step`/`then_step`/`fallback_step` are iteration-prefixed at expansion and must name substeps of the same loop; `continue_on_error: true` is the iteration-scoped alternative. Dormant step-level instances have been found and left uncorrected in several workflows (e.g. page-build-handler) pending the next touch, per the correct-while-touching norm.
- **sources:** 016b_debugging_guide_7_3_(7).md#error_step-entry; RUNBOOK_travelling_docs(38).md#§8; RUNNING_NOTES_travelling_docs(39).md#rev9,#rev12,#rev13; fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas
- **relations:** docs-never-fail containment; correct-while-touching norm; loop_expansion_handler.go; max_tokens placement gotcha
- **verify-later:** `routeToErrorStepOrFail`/`continueExecution` in coordinator; loop_expansion_handler.go prefixing

### DEV-032 — The seam rule — every prompt consuming a spec field must render it
- **status:** deployed
- **status-evidence:** Migration 138 applied ("Mandatory Behaviour Requirements" section rendered from spec.interactive_features, marked as overriding the source).
- **what:** A requirement can survive the analysis step and still be ignored at generation if the generation prompt never renders it: `analyze_tool` rendered `spec.interactive_features`, `recreate_tool` didn't, and the model trusted the visible source HTML over requirements buried in a 20KB analysis JSON — faithfully recreating the bugs it was asked to fix. Rule: when adding a spec field, grep EVERY prompt_template that should render it; render requirements explicitly, marked as overriding the source.
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45-run2; HANDOFF_2026-07-10…md#§4
- **relations:** economy-simulator case; "passed checks ≠ working"; array-item field contract (page-content-writer, a concrete symptom of this same seam-rule class)
- **verify-later:** recreate_tool prompt's Mandatory Behaviour Requirements section (migration 138)

### DEV-033 — Manual agent trigger via kcat orchestrate envelope (never hand-roll spawn+call)
- **status:** deployed
- **status-evidence:** Independently proven and re-derived across at least four workstreams: the manual kcat trigger scripts 084–087 (travelling_docs, exercised repeatedly); "the documented system.intake pattern is STALE… the working mechanism is the kcat trigger script pattern" and "Do NOT hand-roll spawn_agent+call_agent inline workflows" (imagery); "the dual-placement trigger pattern worked again" (vonc, 2026-07-06).
- **what:** Manually triggering agents means posting an `action=orchestrate` envelope to `system.agent.generic.requests` with the house header set (correlation/orchestration/request/message ids) plus `config.agent_type=<target>` and input_data — known-good for improvement-loop, webdesign-agent, rerender-pages. Hand-crafted inline spawn+call parents fail because the spawned child runs its workflow on INIT and idles before the call arrives; work destined for spawned handlers must instead route through work items + dispatch. Encoded operational knowledge in the trigger-script family: target the spawn-wrapper orchestrator, not the agent directly (only a spawned pod gets certain secrets via the spawn gate); use an explicit REF, never HEAD; print effective subject/function as a go/no-go banner; default to DRY-RUN with an explicit SEND=1 to fire.
- **sources:** 084_TRIGGER_diagnose_v1(2).sh, 085/086/087 headers; RUNNING_NOTES_travelling_docs(39).md#rev10,#rev27; HANDOFF_imagery_best_in_class.md#Mechanisms; RUNNING_NOTES_imagery_best_in_class.md#Turn-18/#Turn-26; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6+#§8
- **relations:** launch idioms (orchestrate vs work-item insert); call_agent contract validation dual-placement; basic operations reference
- **verify-later:** whether the trigger scripts were promoted anywhere canonical (they live in drafts/); system.intake topic absence; scripts/initial_messages/210_vonc_trigger/ contents

### DEV-034 — call_agent contract validation vs input_data.spec.* convention (dual-placement requirement)
- **status:** deployed
- **status-evidence:** Mechanism confirmed directly in code (2026-07-01): ValidateInputContract checks only top-level keys; a patch (PATCH_validate_input_contract.go) was "WRITTEN, deploy PENDING" and still listed as a backlog item over a week later.
- **stage2-verified (2026-07-14):** partial → deployed — input_mapping.go:245 ValidateInputContract now checks BOTH top-level (line 265 comment '1) top-level') AND input_data.spec.* (line 269 '2) input_data.spec.* — the path handlers actually read (doc 003)'), with hint text at line 289 confirming dual-placement acceptance. git log shows commit ca2b89a79 'patch input cont...
- **what:** `call_agent` resolves input_mapping then validates the target's `input_contract.required` against TOP-LEVEL keys only, while handler workflows for work-item-driven agents read spec fields at `input_data.spec.*` (the work-item convention) — the validator and the workflow read different places. A required field like `section_type` can therefore be satisfied by neither pure-top-level (empty-context generic generation) nor pure-nested (contract violation); the working manual-trigger shape provides the field BOTH top-level and inside spec. The generic build-dispatch-loop mapping flattens no such field, so the designed work-item path would hit the same violation if exercised (predicted, unconfirmed at the time). The proposed framework fix: the validator should accept a required field top-level OR at `input_data.spec.X`, not enshrine the duplication or patch per-handler loop mappings.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~10:10 + #2026-07-01-~12:46 + #2026-07-01-~13:10; docs/PATCH_validate_input_contract.go.txt; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e; 016b_debugging_guide_7_3_(7).md#spawn-call-entry
- **relations:** manual agent trigger via kcat orchestrate; build-dispatch-loop genericity; component regeneration
- **verify-later:** input_mapping.go ValidateInputContract (patched?); a needs_component_regeneration item dispatched through the loop

### DEV-035 — ExtractActionInputs Strategy-0 explicit dot-paths lesson
- **status:** partial
- **status-evidence:** "FIXED workflow-only: SQL_2026-07-12_asset_deployer_explicit_paths.sql… Standing lesson: give ExtractActionInputs actions explicit dot-paths; never trust the search" — with a related dispatch-shape gap recorded but not fixed.
- **what:** ExtractActionInputs' aggressive recursive field search matched a stale `purpose` elsewhere in collected_data, so a sprite sheet deployed as a 900×900 hero-config JPG despite the child receiving `purpose='sprite_sheet'` — explicit Strategy-0 dot-path config values are resolved first and win, so giving actions explicit dot-paths defeats the mis-search risk. Latent siblings: items dispatched via build-dispatch-loop carry payload under `input_data.spec.*`, which the explicit paths miss; historical spawned deploys may have silently used hero dimensions.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-31; SQL_2026-07-12_asset_deployer_explicit_paths.sql; HANDOFF_imagery_best_in_class.md#I2.1
- **relations:** standardized input extraction; field-name collision via nested-source loop; deploy_image_asset defaults
- **verify-later:** asset-deployer deploy_asset step config paths; datahelpers extraction strategies

### DEV-036 — psql read-only PreToolUse gate
- **status:** deployed
- **status-evidence:** "added a PreToolUse permission hook (.claude/hooks/psql_readonly_gate.py)… tested against a 20-case matrix and proven live" (2026-07-08).
- **what:** Agent-session tooling: a hook auto-approves read-only SELECT/`\d` psql via the exact kubectl-exec form while mutations still prompt the human, reducing friction for the DB ground-truth checks every session performs. Session auth expires roughly daily.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-3; HANDOFF_imagery_best_in_class.md#Mechanisms; RUNBOOK_imagery_best_in_class.md#A1
- **relations:** context-bundle seeding; schema-before-SQL discipline
- **verify-later:** .claude/hooks/psql_readonly_gate.py; settings.local.json hook wiring

### DEV-037 — Whole-blob input_data passthrough mapping (anti-pattern)
- **status:** superseded
- **status-evidence:** An archived doc presents `{"input_data": "input_data"}` as a working pattern used by three orchestrators; the live guide identifies it as broken: "It does not do what it looks like it does... map each expected field by name."
- **what:** A wrapper-orchestrator shorthand once documented as valid. It actually double-nests the caller's data; the live convention replaces it with explicit per-field mapping using `?`-suffixed optional keys. Known live instances: all four med-* wrappers still carry this broken mapping.
- **sources:** old/001_development_guide.md#"Standardized Input Extraction"; docs024_key_docs_latest/001_development_guide(5).md#"Map fields individually, not the whole input_data blob"
- **relations:** ExtractActionInputs nested-source collision; input_mapping `?` suffix convention; wrapper-orchestrator pattern
- **verify-later:** grep current agent_definitions for `"input_data": "input_data"` mapping still in use

### DEV-038 — Roadmap-phases enforcement gap (routed to builder thread)
- **status:** deployed (as a documented finding; fix owned elsewhere)
- **status-evidence:** "RECLASSIFIED: this is the builder thread's MAIN queue item now."
- **what:** The dev guide already defines a Tier-3 roadmap-with-phases mechanism, but `082_submit_domain_unified.sh` has no `--roadmap` entry point and `build-site-planner`'s prompt has no else-branch for an absent roadmap — so absent a roadmap, phase constraints vanish rather than degrade gracefully. Confirmed in code as an absent decision point, and routed to the builder thread as a relay-wide fix rather than fixed by the workstream that found it.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(9).md#ROOT CONTEXT FOUND IN THE DOCS; fixloop_eg_dartsonline/NOTES_running_fixloop(9).md#2026-07-07 later still
- **relations:** curated best-in-class standing expectation; abandoned pilot candidates
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner prompt roadmap_brief block

### DEV-039 — Development-guide gotcha: max_tokens must live inside ai_service
- **status:** deployed
- **status-evidence:** "max_tokens at a step-config's root is DEAD CONFIG for execute_llm_prompt."
- **what:** `execute_llm_prompt` reads `max_tokens` only from the agent's top-level config or from inside the step's `ai_service` block; a root-level step-config `max_tokens` is silently ignored and the Anthropic client defaults to 2048 output tokens. This capped a diagnose-agent verdict step at 2048 tokens through five benchmark runs undetected, and truncated a fix-proposer's plan mid-JSON twice.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#BENCHMARK RUN 2; fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** error_step placement gotcha (same family); fix-proposer's truncation failures
- **verify-later:** grep/inspect `execute_llm_prompt`; `max_tokens`; `ai_service`

### DEV-040 — Development-guide gotcha: verify deployed contents against the pod, never tag/git
- **status:** convention
- **status-evidence:** "verify deployed contents against the POD binary, never the tag, never git."
- **stage2-verified (2026-07-14):** deployed → convention — deploy-verification discipline, not a built artifact
- **what:** A same-tag deploy trap: bumping source without bumping `IMAGE_TAG` means `rollout restart` reuses the cached image, so a reported "deploy" can silently ship a stale binary. The only reliable verification is grepping the running pod's binary for control strings. Caught a first-deploy that was actually a no-op.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 23; fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#Live gotchas
- **relations:** rebalance-window gotcha
- **verify-later:** grep/inspect `IMAGE_TAG`; `rollout restart`

### DEV-041 — Development-guide gotcha: rebalance window after chassis restart
- **status:** convention
- **status-evidence:** "after make release redeploy-agents, wait for the chassis deployment to settle before firing a diagnosis" — cost 8 hours of debugging the first time it bit.
- **stage2-verified (2026-07-14):** deployed → convention — operational workaround/timing rule, not a code artifact; file itself says 'n/a — process/design record'
- **what:** Firing an orchestration within roughly 300 seconds of a chassis pod restart risks the spawn's init response falling into a Kafka consumer-rebalance window and dying silently. Standing workaround: wait ~300s after any deploy before firing a run.
- **sources:** fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 9; fixloop_eg_dartsonline/HANDOFF_turn21_2026-07-10.md#Gotchas
- **relations:** same-tag deploy trap
- **verify-later:** n/a — process/design record

### DEV-042 — Development-guide gotcha: BST/UTC timestamp mismatch
- **status:** deployed
- **status-evidence:** "orchestration_states.last_activity is timestamp WITHOUT time zone... dev host runs BST while DB is UTC."
- **what:** `last_activity` is stored without time zone while `created_at` is timestamptz, so `NOW() - last_activity` arithmetic is silently wrong by the local UTC offset; combined with a dev host running BST against a UTC database, a run can appear to have finished before it started.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas; fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 6
- **relations:** rebalance-window gotcha
- **verify-later:** grep/inspect `last_activity`; `created_at`; `NOW() - last_activity`

### DEV-043 — Observability signature-fields pattern
- **status:** deployed
- **status-evidence:** A code diff implementing the pattern was applied; a separate FOCUS doc shows the fields live in the result map.
- **what:** A reusable debugging convention: when patching a code path whose old/new behaviour is otherwise indistinguishable, write new marker fields into the result map (flowing into `orchestration_states.collected_data`) that are absent from the old code — their presence proves the new code executed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_visual_pipeline_css_and_component_lists.md#Observability; js_snippets_news_gaswholesalers/old/design_actions_observability_patch.md
- **relations:** CSS component-list fallback bug
- **verify-later:** grep/inspect `orchestration_states.collected_data`

### DEV-044 — needs_section_data review items appearing on successful builds (open question)
- **status:** unknown
- **status-evidence:** "Worth a separate look at why a successful structured build still raises a section-data review item" — listed open, never investigated.
- **what:** Even a clean, isolated test build spawned a child `needs_section_data` work item with `status=needs_human_review` and no `handler_agent`, unexplained.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md; js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** post-build validation of structured components; isolated build test methodology
- **verify-later:** needs_section_data work-item creation path; wont_fix auto-resolution pattern

### DEV-045 — Cross-module port / copy-drift discipline
- **status:** convention
- **status-evidence:** Both the validated port procedure ("surfaced FOUR build errors in the first real make build-core-manager … None of these four were logic bugs") and an independent later incident ("This is the 5th instance of the same root pattern… import path, missing file, stale sibling, assumed helper API, now stale CLI") converge on the same prevention.
- **stage2-verified (2026-07-14):** deployed → convention — port/copy-drift procedure, not a code artifact itself
- **what:** The validated sequence for moving a package between Go modules (e.g. contextkit's internal/diagnose → chassis's pkg/diagnose): (1) copy the WHOLE package as a unit and diff file lists — cherry-picking individual files across versions silently drops or staleifies siblings; (2) rewrite the moved-package import path everywhere; (3) build+test the package alone before the binary; (4) grep every shared-package call the new code makes against the REAL helper surface rather than assuming an API exists (`datahelpers.ExtractStringSlice` didn't exist; compose `ExtractStringListHelper(ExtractNestedField(...))`). A real incident surfaced five distinct failure classes this way in sequence (wrong import path, an entire file silently omitted, a stale pre-refactor sibling, an assumed-but-nonexistent helper API, a stale CLI binary predating a library change) — all passed silently in the source module and surfaced only on first build/run in the target.
- **sources:** docs019/RUNBOOK_design_diagnosis_loop(7).md#4c; docs019/RUNBOOK(31)_diagnosis_loop.md#current-position; NOTES_running_synthesis_v2(36).md 2026-06-17; 016_additions_assumed_helper_and_cross_module.md
- **relations:** ReadSymbolBody dual placement; canonical field-path resolution helpers; untested-code/behaviour-testing discipline
- **verify-later:** pkg/diagnose file list vs contextkit/internal/diagnose

### DEV-046 — Curated best-in-class standing expectation
- **status:** aspirational
- **status-evidence:** "Best-in-class/curated-list idea homed: standing expectation (guides+tools+news+non-affiliate curated top-N) + 'not-original-can-still-be-best' clause → 001_development_guide" (2026-07-07).
- **what:** A proposed platform-wide addition to the development guide requiring every commerce-shaped domain to carry a baseline of guides, tools, a news feed, and a curated non-affiliate top-N list — enforced the same way as the roadmap-phases gap (relay-wide strategist/planner prompts, not per-message or the constitution) — with the explicit doctrine that "useful-but-unoriginal still counts as best-in-class."
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-07 "pilot candidate 2" entry; NOTES_running_fixloop(9).md "Builder queue item 7"
- **relations:** roadmap-phases enforcement gap; diagnosis→fix loop workstream founding
- **verify-later:** —

### DEV-047 — Building-discipline edge cases (pre-registered engineering checklist)
- **status:** aspirational
- **status-evidence:** "Self-referential structures need a cycle guard... A multi-step apply must be all-or-nothing... Bulk operations need bulk confirmation."
- **what:** A checklist of edge cases the design docs insist be caught before building anything with self-modifying or autonomous behaviour: cycle guards on any parent-link/version-chain walk; transactional multi-write-plus-event apply (outbox pattern) so a crash can't leave a dangling row with no event; reading a consistent point-in-time snapshot when assembling from multiple tables; "one live thing per target, all the way down" (dedup at every layer, not just the queue); bulk-confirmation for large batches; filtering transient/infra failures out of any trust-affecting evidence signal; and "tell not-yet apart from broken" (missing-because-unonboarded degrades gracefully, missing-because-malformed fails loudly).
- **sources:** NOTES_running_synthesis_principles(59) §Building discipline
- **relations:** trust ratchet & capability ceiling model; untested-code / behaviour-testing discipline
- **verify-later:** —

### DEV-048 — Untested-code / behaviour-testing discipline
- **status:** deployed
- **status-evidence:** "PRINCIPLE 2026-06-13: untested code is a liability, surfaced by the dedup-move bug... COMPILE/gofmt/vet prove syntax, NOT behaviour."
- **what:** A hard-won, explicitly codified lesson (triggered by a dedup tool's silent destructive-flag bug) that compiling/gofmt/vet only prove syntax, never behaviour; any destructive CLI operation must be report-only by default; and Go's `flag.Parse()` stopping at the first positional argument is a specific, recurring footgun requiring manual value-flag-aware argument separation in every CLI that takes a positional followed by flags — audited across `dedup`, `thin_versions`, `resolve_targets`, `embed`, `assembler`, `fuse`, `eval_targets`, `dbcontext`.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-13 "PRINCIPLE" entry; NOTES_running_synthesis_v2(36).md 2026-06-14 "runbook code audit"
- **relations:** docs archiving toolchain (dedup); building-discipline edge cases; cross-module copy-drift discipline
- **verify-later:** —

### DEV-049 — Schema-before-SQL discipline
- **status:** convention
- **status-evidence:** "Schema: bundle CODE names tables; only \d gives columns AND persistence (hit 4×: page_id, no status column, file_count, the 3-NULL workflow columns)."
- **stage2-verified (2026-07-14):** deployed → convention — schema-before-SQL discipline; verify-later explicitly '—'
- **what:** A recurring, explicitly named discipline that code reliably names which DB tables are involved, but only a live `\d` (schema dump) gives real column names and reveals whether a field is persisted at all vs. computed at runtime — hit repeatedly (a wrong `page_id` column, an assumed-but-nonexistent `status` column on `site_plan_sections`, a wrong `fileCount` vs `file_count` JSON key, and a workflow-column misassumption) and eventually generalised into "real rows/examples beat prose/inference" as its own standing lesson.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST "Standing lessons"; NOTES_running_synthesis_principles(59) multiple 2026-06-14 gamesdesign-diagnosis entries
- **relations:** workflow lives in default_config; real-rows-beat-prose discipline; house rules
- **verify-later:** —

### DEV-050 — Real-rows-beat-prose-or-assumption discipline
- **status:** deployed
- **status-evidence:** "Real rows/examples beat prose/inference: the agent_definitions workflow lives in default_config — the dev guide prose didn't say which column; the row did."
- **what:** A generalised standing lesson, distilled from several specific incidents, that when a dev-guide or design doc's prose is ambiguous or silent about an implementation detail, the correct source of truth is a real, live example row/file, not inference from the prose — repeatedly the deciding move that caught a wrong migration or wrong action draft before it was applied.
- **sources:** NOTES_running_synthesis_v2(36).md STATE DIGEST; NOTES_running_synthesis_principles(59) "GROUNDED... tools/provenance/docs design corrected" entries
- **relations:** workflow lives in default_config; schema-before-SQL discipline; child-completion result key convention
- **verify-later:** —

### DEV-051 — Workflow lives in default_config, not the workflow columns
- **status:** deployed
- **status-evidence:** "confirmed: those agents put their workflow there; the task_workflow / orchestrator_workflow columns are NULL"; the move/fix/restore migration sequence applied after being bitten by it.
- **what:** The loader reads `agent_definitions.default_config` (`{workflow:{start_step,steps}, processing_mode, timeout_seconds}`); the three workflow columns (task_workflow/orchestrator_workflow/the third) are dead for working agents. A workflow seeded into `orchestration_workflow` silently never loads — this bit a real agent pair (seeded into the wrong column, then the orchestrator's workflow was lost entirely during a move and had to be re-seeded). Correction learned by reading a real row rather than docs.
- **sources:** NNN_move_diagnose_workflow_to_default_config(1).sql; NNN_restore_diagnose_orchestrator_workflow(1).sql; PLAN_workflows_and_actions_migration(19).md
- **relations:** schema-before-SQL discipline; sub-agent modelling conventions; confirmed chassis workflow model
- **verify-later:** workflow loader code path; whether the three columns are still consulted anywhere

### DEV-052 — Confirmed chassis workflow model (group agents, promotion pattern, wrapper orchestrator)
- **status:** deployed
- **status-evidence:** "Confirmed model (from the guideline docs …)" plus a changelog confirming against two real agent_definitions rows and the action code.
- **what:** The model verified via migration work: workflows are declarative JSON steps in default_config; a generic config-driven action library (query_database, spawn_agent, call_agent, loop, conditional, rag_lookup, work-item lifecycle) is reused before writing Go; LLM-assisted checks group into one agent per shared context load (explicitly rejecting a registry of mini-action agents), promoted to spawned sub-agents only when one needs independence (a one-line workflow change); the wrapper orchestrator (spawn→call→complete) is the canonical small form; spawning is `spawnAgentKubernetesJobFromDefinition` with per-spawn job topics. Reuse discipline is encoded as queries (search agent_definitions; `default_config::text ILIKE '%<action>%'`).
- **sources:** PLAN_workflows_and_actions_migration(19).md; DESIGN_diagnosis_loop_chassis_integration(6).md#0
- **relations:** wrapper-orchestrator pattern; workflow lives in default_config; STEP ZERO
- **verify-later:** 001/002 guideline docs (canonical home)

### DEV-053 — Development Guide (agent-build daily reference) [doc artifact]
- **status:** superseded
- **status-evidence:** The live 001 doc's consolidation note states "This is the canonical 001_development_guide. It supersedes the prior copy."
- **what:** The consolidated practical reference for building/debugging/maintaining agents: core design principles (agents own their domain, callers pass raw data, workflows simple with complexity in Go, actions are the unit of work, spawn-before-call, reply-to-caller's-topic), a new-agent checklist, a migration guide, and 20+ lessons-learned bug entries. The archived copy referenced here has a live successor in docs024_key_docs_latest.
- **sources:** WM/001_development_guide(0).md#core-design-principles, #checklist-for-new-specialist-agent, #summary-of-rules-for-the-dev-guide
- **relations:** STEP ZERO; wrapper-orchestrator pattern; loop mechanisms; thin-slice constitution
- **verify-later:** platform/orchestration/actions/*

### DEV-054 — thin-slice constitution (always-on rules doc)
- **status:** deployed
- **status-evidence:** "Included in full in every bundle... Later it becomes the `standards` rows with `scope = constitution`; the content is the same."
- **what:** The flat-file version of the chassis's always-on rules, pasted into every assembler/bundle output: reuse-before-recreate, fix-structural-not-symptoms, every-agent-is-an-orchestrator, no subworkflows-in-SQL (spawn sub-agents instead), the snake_case/kebab-case naming split, storage conventions (text+CHECK enums, version+previous_version_id, deleted_at soft-delete), logging rules (no `logger.Debug`, log the orchestration_id/correlation_id), deployment path (GitHub → Actions → Backblaze S3), and plain/pragmatic generated-text tone. Task-specific 003 contracts are listed but pulled in only when a task touches them.
- **sources:** contextkit/thin_slice_constitution.md
- **relations:** assembler; docselect.go; contracts-and-standards (003); development guide
- **verify-later:** whether the constitution has since migrated to `standards` rows with `scope = constitution`, or is still the flat file

### DEV-055 — v1 monolithic LLM-chain site builders
- **status:** superseded
- **status-evidence:** A later migration renames the v1 successor "multipage-website-builder v3 → pageflow-builder" and explicitly captures the whole v1 fleet as "old"; root files never patch these agents again.
- **what:** The first architecture (2025-11/12): a website-builder orchestrator spawns a chain of one-LLM-call specialists — domain-analyst (audience/tone JSON), site-architect (page structure + colours), content-creator (copy JSON), html-developer (whole-page HTML), multipage-wrapper (file map), site-deployer (git commit). Everything is free-form LLM output; no component library, no DB page records.
- **sources:** sql_for_agents_v1/004_website_builder.sql; sql_for_agents_v1/005_domain_analyst.sql; sql_for_agents_v1/008_html_developer.sql; sql_for_agents_v2/027_old_agent_definitions.sql
- **relations:** superseded by pageflow-builder (component-based); site-deployer contract survives in 011_site_deployer.sql
- **verify-later:** agent_definitions rows for these types; whether any workflow still references them

### DEV-056 — Batched multi-page generation and chunked HTML generation
- **status:** superseded
- **status-evidence:** v2 renames its v3 to pageflow-builder; the batch-of-4-pages prompts appear only in v1 snapshots.
- **what:** Anti-token-limit strategies from the v1 era: build 20-page sites by generating pages in five batches of four ("Return as JSON map of filename to HTML"), with shared CSS generated once and injected at assembly; html-developer-chunked generated structure/styles/sections in separate calls. Both ideas were made unnecessary by the component architecture.
- **sources:** sql_for_agents_v1/015_example_20_page_workflow.sql; sql_for_agents_v1/017_multipage_website_builder.sql; sql_for_agents_v1/014_html_developer_chunked.sql
- **relations:** replaced by pageflow-builder per-page loop and later the one-item-per-run dispatch loop
- **verify-later:** none needed — historical

### DEV-057 — Remove-loops plan: input_mapping, contract validation, sequential_fan_out, page-builder worker
- **status:** partial
- **status-evidence:** input_mapping conversion executed; contracts added across many files; but `build_pages_loop` remained and `sequential_fan_out`/page-builder worker never appear in any later agent file.
- **what:** A four-phase plan to replace loop/substep injection: (1) explicit `input_mapping` instead of `input_fields` path-hunting plus runtime input/output contract validation with hard fails and `__raw_message__` deprecation; (2) a `sequential_fan_out` action spawning one child orchestration per page; (3) a page-builder worker agent; (4) rewire pageflow-builder. Phase 1 landed; phases 2–4 were superseded by the site_work_items dispatch-loop architecture, which achieves the same "one visible orchestration per unit of work" goal differently.
- **sources:** 001_remove_loops_in_workflow.md; 001b_implementation_plan.md; 030_input_mapping_changes.sql; 030b_remaining_agents_needing_input_mapping; 001_validator_sql.sql
- **relations:** standardized input extraction; loop-action dispatch (spiritual successor); site_work_items unified work queue
- **verify-later:** chassis code: contract validation enforcement; whether `sequential_fan_out` exists in the registry; `__raw_message__` fallback removal status

### DEV-058 — site_work_items unified work queue and lifecycle
- **status:** deployed
- **status-evidence:** Full DDL (023_site_work_items) with dedup index, plus dozens of live operational patches (resets, handler re-routing, attempt bumps) against real sites.
- **what:** Every piece of platform work is a row: source (planner/discovery/content_feed/manual/improvement/side_effect/human/validation), pipeline (originally `domain`, later renamed), item_type, severity, spec JSONB, target refs (page/component/entity/url), triage enrichment (impact, resolution_path, suggested_action, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete/pending_verify/verified/failed/rejected/wont_fix plus 'blocked' (handler missing), dependencies (depends_on UUID[], parent/related/batch), attempts, and a deterministic item_key with a partial unique index for dedup among non-terminal items. A same-structure archive table receives terminal items.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#approval_mode; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle
- **relations:** loop-action dispatch; claimed-item timeout; work-item dedup mechanics; work item archival; build_queue site seeding
- **verify-later:** current status distribution; pipeline column rename (`domain` dropped)

### DEV-059 — Work item archival
- **status:** deployed
- **status-evidence:** `archive_completed_work_items(age, batch)` function plus archiver agent definition plus daily scheduled task, with schema-sync ALTERs and FK handling.
- **what:** Terminal work items (complete/failed/wont_fix) older than a configurable age move to `site_work_items_archive` in batches, keeping the live queue small. The function handles column drift between live and archive tables explicitly (parent self-ref cleared, content_feed_items references deleted).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-archiver
- **relations:** site_work_items unified work queue; scheduler
- **verify-later:** archiver task enabled; archive row counts

### DEV-060 — build_queue site seeding
- **status:** deployed
- **status-evidence:** Phase-0 Block A table with direction semantics enumerated.
- **what:** Domain-level intake queue for new sites: a row per domain with direction JSONB (null | {objective} | {adopt_from} | {fork_from} | {brief_complete...}), status and priority. `seed_build_queue` reads it, creates site records and initial work items according to direction — the entry point into the work-item pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#2
- **relations:** site_work_items unified work queue; adoption pipeline (adopt_from); onboarding-config (brief_complete)
- **verify-later:** seed_build_queue action; build-pipeline-trigger seeding behaviour

### DEV-061 — Loop-action dispatch (build-dispatch-loop, migration 071)
- **status:** deployed
- **status-evidence:** Applied UPDATE of build-dispatch-loop default_config: "Step-chaining... processes only one work item per trigger. The loop action is proven in maintenance-triage and pageflow-builder."
- **what:** The dispatch loop loads all dispatchable items upfront (dependency-filtered, priority-ordered, max 50) and iterates with the `loop` action running a sub_workflow per item: claim → check_claim → spawn_handler (dynamic agent type from `current_item.handler_agent`) → call_handler → mark_complete/mark_failed, with `continue_on_error`. Introduces `item_variable` scoping (`current_item.*`) and optional `?`-suffixed input_mapping fields silently skipped for handlers that don't need them (section-editor compatibility).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql#migration-071
- **relations:** site_work_items unified work queue; spawn-orchestrator pattern; claimed-item timeout; remove-loops plan
- **verify-later:** build-dispatch-loop live config; loop action implementation

### DEV-062 — spawn_agent — database-definition-driven Kubernetes Job spawning
- **status:** deployed
- **status-evidence:** "All deployment specs now come from database… No Code Changes for New Agents — Just insert into agent_definitions"; spawn walked through line-by-line in a dedicated doc.
- **what:** `SpawnAgentAction` reads the child's agent_definitions row (image repo/tag, resources, health config, env vars, default workflow), inserts an agent_instances row in the client schema, creates job topics, launches a K8s Job with topic env vars, sends the initialize message, and returns spawn results (agent_id, role, topics) into CollectedData for later call_agent lookup.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.043.spawn_actions2_stepbystepthroughthecode.md; docs001_flow_general/README.045.spawn_actions4.refactor_into_functions.md
- **relations:** agent_definitions registry; job topics; spawn→call pattern
- **verify-later:** spawn_actions.go; client_{id}.agent_instances table

### DEV-063 — Role-based agent pools / atomic work-claim queue (proposal, superseded)
- **status:** superseded
- **status-evidence:** A pure proposal ("Migration Path: Phase 1…Phase 4") never referenced as built in later docs; the design is recognisably the ancestor of today's work-item pipeline.
- **what:** Instead of spawning agents tied to IDs, agents would register roles/capabilities and claim WorkItems atomically from role-specific queues (`system.roles.{role}.pending`); unclaimed work survives agent death; pools scale elastically. "The role becomes the contract, not the agent ID."
- **sources:** docs001_flow_general/README.019.flow8.role_based_agent_pools.md
- **relations:** successor: site_work_items unified work queue / page-build-handler pipeline; scheduler-and-tasks concurrency groups
- **verify-later:** work_items table and claim semantics in the current codebase, compared with this 2025 sketch

### DEV-064 — Prompt resolution priority hierarchy
- **status:** deployed
- **status-evidence:** Documented with three tested scenarios and log lines ("Using prompt from incoming message (Priority 1)").
- **what:** `execute_llm_prompt` resolves its prompt in priority order: (1) prompt passed in the incoming message/step config by the caller, (2) the agent's own `prompt_template` from agent_definitions, (3) workflow-step fallback. Lets parents override specialists while specialists keep good defaults.
- **sources:** docs001_flow_general/README.004.call_agent1.refactor_into_functions.md
- **relations:** execute_llm_prompt generic action; agent_definitions default_config
- **verify-later:** ai_actions.go ExecuteLLMPromptAction prompt lookup order

### DEV-065 — CollectedData normalisation and data_helpers safe-access layer
- **status:** deployed
- **status-evidence:** Full data_helpers.go source reproduced as "the new functionality" and used by all subsequent agents ("data_helpers.go functions ensure consistency").
- **what:** One central layer (data_helpers.go) normalises every inbound message into a canonical CollectedData shape — `input_data` always at top level, system fields (`__execution_context__`, `__my_responses_topic__`, `__raw_message__`…) separated — and provides the only sanctioned accessors (GetInputData, GetStepData, GetMultipleStepData, GetFieldFromPath, TransformDataForAction, BuildRequestMessage/BuildResponseMessage/BuildInitializationRequest). Killed the `input_data.input_data` nesting chaos. Child input_data is always overwritten at top level — each agent's context is exactly what its parent sent (clean-slate encapsulation).
- **sources:** docs001_flow_general/README.070.a.centraliseddatanormalisation.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md; docs001_flow_general/README.024.flow14.input_data.md; docs001_flow_general/README.080.a.packaging_data.md
- **relations:** output_field/input_fields mapping contract; canonical field-path resolution helpers (successor lineage)
- **verify-later:** platform/orchestration/datahelpers package

### DEV-066 — Agent groups — versioned project-recipe teams (discovery, immutable versioning, pinning)
- **status:** partial
- **status-evidence:** "FindBestGroup… queries the database to find the best available version of that group, ordered by performance, usage, and version"; groups used in every website build of that era; an EvolutionService was designed to mutate groups but is not evidenced live; a UNIQUE(group_type, version) constraint was added, a partial realisation of the versioning model.
- **what:** `agent_group_definitions` rows are project recipes, not "agents that work together": a `group_type`, an `agent_configs` squad (role → agent_type), and an `orchestration_workflow` JSON, with integer versions as immutable snapshots (unique group_type+version). Requests name a capability (group_type) and the system picks the best version via FindBestGroup. Each buildable *kind* of output (landing page, content site, 11ty blog, ecommerce) is its own self-contained group with its own squad/workflow/questionnaire; divergence in output structure/build/deployment means a new group, not conditional routing; a site records the group_version that built it and rebuilds with it unless upgraded; groups may pin specific agent versions where stability matters; duplication across similar groups is accepted for clarity. An EvolutionService designed to mutate groups into new versions with parent_id lineage and performance-based selection appears aspirational — the discovery/versioning half shipped.
- **sources:** docs001_flow_general/README.060.groupagents1.md; docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.062.groupagents3.databases.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion; docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md
- **relations:** workflow selection priority; spawn_group action; site manifest pinning group_version; agent_definitions registry
- **verify-later:** agent_groups vs agent_group_definitions tables (both exist with different shapes); discovery/agent_discovery.go FindBestGroup; evolution.go; group version rows per group_type

### DEV-067 — Workflow selection priority (inline override > group > agent default)
- **status:** deployed
- **status-evidence:** `selectWorkflow` is defined and implemented with the three-tier priority; HITL tests routinely use inline workflow overrides.
- **what:** `processor.selectWorkflow` resolves which workflow to run: (1) a full inline workflow in the message config (ephemeral/testing), (2) a group workflow found via group_type, (3) the agent's default workflow from agent_definitions. Keeps production versioned while allowing ad-hoc experiments.
- **sources:** docs001_flow_general/README.061.groupagents2.md; docs001_flow_general/README.060.groupagents1.md; docs002_hitl_parallel/README.0106.hitl_multistep_approval.md
- **relations:** agent groups; SagaCoordinator
- **verify-later:** processor.go selectWorkflow

### DEV-068 — agent_definitions registry (DB-driven agent config and versioning)
- **status:** deployed
- **status-evidence:** Dozens of INSERT/UPDATE statements across all eras; constraint migration to UNIQUE(type, version) with previous_version_id; category CHECK constraint managed separately.
- **what:** Every agent type is a row: type, display_name, category (constraint-checked: data-driven/code-driven/adapter/…), default_config (containing the workflow, ai_service model+provider, processing_mode, timeouts), capabilities, image_repository/tag (all agents share the agent-chassis image), resources, topics, health_config, env_vars, version + previous_version_id, task_workflow/orchestrator_workflow, delegation_preferences. Creating an agent is a database insert, not a code change.
- **sources:** docs001_flow_general/README.042.spawn_actions.md; docs001_flow_general/README.096b.robothandswebsite.md; docs003_firecrawl/README.0140.removing_constraint.md; docs001_flow_general/README.098.oldherocontentdefinition.d
- **relations:** spawn_agent; agent/group categorisation taxonomy; agent groups
- **verify-later:** agent_definitions schema and constraints today; how many early agent types still exist/are active

### DEV-069 — Spawn/step naming conventions
- **status:** deployed
- **status-evidence:** "The naming conventions are now important because we're using them to find spawned agents" — spawn_ prefix required, unique step names with 3-letter suffixes.
- **what:** Workflow authoring rules: spawn steps must start `spawn_<descriptor>` (suffix hints the role), action steps use perform_/execute_/process_ prefixes and reference agents by role, and step names must be unique within a workflow.
- **sources:** docs001_flow_general/README.044.spawn_actions3.spawn_rules.md
- **relations:** call_agent role lookup; spawn→call pattern
- **verify-later:** whether current workflow JSON still relies on prefix conventions

### DEV-070 — evaluate_condition — template-based conditional branching
- **status:** deployed
- **status-evidence:** Documents the working mechanism ("The orchestrator uses this to pick the next step from the next_step map") including Go text/template functions and a live website-analyzer group UPDATE.
- **what:** Workflow steps gain branching: `evaluate_condition` renders a Go text/template expression against CollectedData and returns true/false; `next_step` becomes a map `{"true": …, "false": …}`. Enables data-driven workflow paths.
- **sources:** docs003_firecrawl/README.0127.conditional_branching.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** conditional_branch/conditional_route actions; dynamic agent routing (later, richer routing)
- **verify-later:** evaluate_condition in registry; coordinator support for map-typed next_step

### DEV-071 — MVP site builder pipeline (strategist → architect → content-creator → deployer)
- **status:** superseded
- **status-evidence:** Full group SQL + per-agent Kafka payloads documented and run against a real site; renamed/extended into a later pipeline; today's site building is the work-item pipeline.
- **what:** The first end-to-end production pipeline: chief-strategist (LLM → build_plan JSON of functional sections), site-component-architect (assemble_from_library → empty semantically-tagged HTML template + content_requirements "shopping list"), content-creator (fills slots), deployer-agent (commit_to_git). Group workflow spawns all four then calls them in sequence, threading outputs through output_field/input_fields.
- **sources:** docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** grew into +brand-designer/+briefing-agent/+html-assembler/+specialist architects; successor: page-build-handler/work-item pipeline
- **verify-later:** agent_group_definitions mvp-site-builder / landing-page-builder rows

### DEV-072 — Dynamic agent routing (route_by_field / conditional_call_agent)
- **status:** deployed
- **status-evidence:** Both Go actions written with registry additions listed; conditional_call_agent chosen because it "wraps CallAgentAction internally — no coordinator changes needed."
- **what:** Data-driven agent selection inside workflows: `route_by_field` maps a dot-path field value to a next step via a routes table with default; `conditional_call_agent` reads e.g. `brief_data…site_type`, maps value→agent type and calls that agent in one step, returning routing metadata.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023a.description_for_conditional_routing_etc; docs004_website_capture_project/006semantic_themes/README.024.conditional_step_routing.md
- **relations:** evaluate_condition (simpler predecessor); spawn_group dynamic group_type (group-level equivalent)
- **verify-later:** registry entries conditional_call_agent, route_by_field

### DEV-073 — spawn_group action with DB group lookup and dynamic group_type
- **status:** aspirational
- **status-evidence:** Discovers an existing SpawnGroupAction (config-provided agents) and revises the new version to align — DB lookup of agent_group_definitions, dynamic group_type_field from collected_data, questionnaire fetch.
- **stage2-verified (2026-07-14):** partial → aspirational — registry.go:89 only maps "spawn_group": SpawnGroupAction (spawn_group.go:31, the config-provided-agents version). The DB-lookup version (doc-commented 'SpawnGroupFromDBAction', actually implemented as func SpawnGroupActionNewerOld at spawn_group.go:163) is never registered — the 'Add to GlobalActionRegistry' instruc...
- **what:** Spawning an entire agent group as a unit: the original action spawned each configured agent and returned subtree info; the enhanced version resolves the group definition (agents + workflow + questionnaire) from the database, with group_type optionally taken dynamically from a prior step's output — enabling the intake orchestrator's dispatch.
- **sources:** docs004_website_capture_project/007different_types_of_site/028.agent_group_selection_and_workflow.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** intake orchestrator; agent groups
- **verify-later:** spawn_group vs spawn_group_from_db in codebase

### DEV-074 — Agent/group categorisation taxonomy (category, status, domain_tags)
- **status:** unknown
- **status-evidence:** Migration SQL with CHECK constraints (category: builder/analyzer/collector/transformer/evaluator/researcher/workflow/monitor; status: active/experimental/deprecated/demo/template) and GIN-indexed domain_tags; no doc confirms it was applied.
- **what:** Organisational metadata over agent_definitions and agent_group_definitions: what the agent *does* (domain-agnostic category), its lifecycle status, and flexible domain tags — an early attempt at the registry hygiene the concept register itself now pursues.
- **sources:** docs004_website_capture_project/998categorisation/031_add_categorisation_to_tables.sql
- **relations:** agent_definitions registry; documentation-system indexing
- **verify-later:** do the category/status/domain_tags columns exist?

### DEV-075 — Aggregation patterns (aggregate_data, aggregator agent, input_from_collected_data)
- **status:** partial
- **status-evidence:** aggregate_data's failures traced (count:0 on verbose child responses); redesign to a spawned aggregator agent fed via input_from_collected_data path mapping; aggregate_webpage became the shipped variant for pages.
- **what:** Combining multi-step results: the local aggregate_data action broke against verbose child state objects; the redesign either normalises responses (data helpers) or delegates aggregation to a spawned aggregator agent whose call config maps CollectedData paths into its input. Response data keyed as `response_{requestID}` in CollectedData.
- **sources:** docs001_flow_general/README.011.flow2.md; docs001_flow_general/README.022.flow11.initialisationflow.md; docs001_flow_general/README.010.flow.md
- **relations:** data_helpers NormalizeResponseData (the actual fix); aggregate_webpage
- **verify-later:** aggregate_data current implementation

### DEV-076 — output_field / input_fields group-memory data mapping contract
- **status:** deployed
- **status-evidence:** A "Packet Flow" walkthrough resolves the exact semantics producing the "Golden Copy" workflows of that era.
- **what:** The inter-agent data plumbing convention of the group-agents era: a call_agent step's `output_field` names the key under which the child's entire result lands in group memory; the next step's `input_fields` selects which keys are passed on; consumers address values by `<output_field>.<producer's own output key>` paths. Most orchestration bugs of that era were mis-mappings of this contract. Distinct from, but an ancestor of, the later complete-step `output_fields` (plural) contract.
- **sources:** docs004_website_capture_project/website_analysis/README.016.agent_definitions_002.md; docs004_website_capture_project/website_analysis/README.012.first_agent_definitions_etc.md
- **relations:** CollectedData normalisation; child-result shaping output_fields (plural) contract (later mechanism); execute_llm_prompt flattening quirk
- **verify-later:** call_agent output_field handling in coordinator

### DEV-077 — Deliberate discovery + human-approved agent evolution (abandoned)
- **status:** abandoned
- **status-evidence:** Principles ("Deliberate discovery — only at planning and review stages; Human approval — all agent changes require approval; Performance-based evolution") never reappear as a mechanism in later eras.
- **what:** Early governance rules for agent self-modification: the system only creates/modifies agents when starting a new task type or after poor performance review, always with human approval — no heartbeats or automatic decisions. Paired with per-group performance recording and version incrementing.
- **sources:** docs001_flow_general/README.005.discovery.md
- **relations:** agent groups evolution service; HITL; tool-lifecycle health checks (modern relative)
- **verify-later:** none

### DEV-078 — Website build overall plan v0 (first multi-agent website roadmap)
- **status:** superseded
- **status-evidence:** A 6-phase/12-step plan written against the calculator-era platform; every element was rebuilt differently in later eras.
- **what:** The first articulation of "build a website with agents": minimal 3-agent workflow, explicit JSON data contracts between agents, progressive enhancement, mock-LLM-first testing, upload_to_s3 deployment. Registers as the origin point of the entire site-building programme.
- **sources:** docs001_flow_general/README.050.overall_plan1.website_design.md; docs001_flow_general/README.001.actions.md
- **relations:** superseded by MVP site builder pipeline, then the work-item pipeline
- **verify-later:** none

### DEV-079 — Data-path resolution problem (agent vs local action nesting)
- **status:** superseded
- **status-evidence:** Documented failures ("Local Action: CollectedData[\"wrap_multipage\"] ... Extra layer!"; "collected_data.spawn_x.call_x.spawn_y...result") resolved later by input_mapping + ActionInputSpec.
- **what:** The recurring class of runtime failures where workflow config referenced CollectedData paths that didn't match where actions actually stored results — agent calls store flat, local actions add a step-name layer, and each spawn/call deepens nesting. Drove multiple generations of mitigation: workflow builder path computation, explicit output_field conventions, data-flow verification matrices, and finally standardized input extraction.
- **sources:** docs006_workflow_builder/001_workflow_builder.md#The-Problem; docs009_site_interrogation_and_solutions/002_claude_discussion#C; docs015_data_flow_verification/001_data_flow_verification.md
- **relations:** ActionInputSpec/ExtractActionInputs; workflow builder (abandoned, see workflow-authoring register); data-flow verification practice
- **verify-later:** datahelpers.ResolveInputMapping and FindByPath in platform code

### DEV-080 — Data-flow verification matrix practice
- **status:** deployed
- **status-evidence:** A complete per-step verification table ("Config | Value | Verified ✓") including output structures and a registration checklist; the practice is repeated for a later site-work-orchestrator.
- **what:** A documentation/QA practice: before deploying a workflow, trace every step's config paths against the action implementations — where each output lands in collected_data, its structure, and each input's exact path — plus response-header compliance and action-registration checklists. The manual ancestor of automated contract validation.
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** input contracts; workflow validator; ActionInputSpec
- **verify-later:** n/a (practice, not code)

### DEV-081 — Specialist agent design doctrine (agents own their domain)
- **status:** deployed
- **status-evidence:** Five versions of the checklist culminating in a v5 doc; audited against real code in a maintenance doc, including accepted divergences.
- **what:** The core agent-design rulebook: agents are self-contained and independently callable, with dedicated load_* actions gathering their own data; callers pass raw domain identifiers, never derived values ("if changing the child requires updating the caller, you've leaked responsibility"); reuse/patch existing actions before creating new ones; workflows stay declarative (templates/config = intent OK; loops/branching = Go); orchestrator vs agent boundary (what/order vs how); standalone + integrated dual modes; spawn before call; agents reply to the caller's topic; no container config in definitions. An interim v2 rule ("use input_fields not explicit paths") was replaced by the ActionInputSpec regime in v3.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md; docs017_legacy_agent_rules_images_design_keydocs/007_checklist_for_new_specialist_agent_v4.md#Orchestrator-Boundaries; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md#Audit-Notes
- **relations:** ActionInputSpec; specialist vs handler persistence boundary; development guide (current successor)
- **verify-later:** how closely current agents follow the load_* pattern

### DEV-082 — item_key canonicalization (workItemKey builder)
- **status:** partial
- **status-evidence:** "CODE PREPARED; NOT APPLIED" — a `workItemKey(itemType, target)` builder exists in work_items_common.go; apply gated behind further verification.
- **what:** item_key prefixes drifted from item_type across creators: an adoption creator keyed BOTH `needs_content_page` and `needs_tool_recreation` as `needs_page:<name>`, so a tool and a content page of the same name collide on the dedup index and one is silently dropped. Fix: a shared workItemKey builder; the tool item moves to its own prefix while the content item deliberately keeps `needs_page:` co-dedup with planner builds (a recorded decision).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 3)
- **relations:** work-item dedup mechanics; work-item manual-crafting discipline
- **verify-later:** work_items_common.go workItemKey applied?; adoption creator key prefixes

### DEV-083 — Basic operations reference (kcat spawn, scale, monitoring)
- **status:** convention
- **status-evidence:** A concatenation of basic_usage docs actively describing the current operating procedure (spawn_group message shape, headers, monitoring queries).
- **stage2-verified (2026-07-14):** deployed → convention — ops reference/procedure description, not a single artifact
- **what:** The operator's basic-usage layer: scale the deployment set up/down (agent-chassis, auth-service, content-creator-agent, core-manager, image-generator-adapter, reasoning-agent, web-search-adapter); post spawn_group/orchestrate messages via kcat from a test pod to the cross-namespace Kafka bootstrap with required headers (correlation_id, request_id, client_id, agent_instance_id, fuel_budget); monitor via orchestrator_state/orchestration_states by correlation_id. The fuel_budget header and the fixed header set are part of the platform's message contract.
- **sources:** docs/summary.txt; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6
- **relations:** manual agent trigger via kcat orchestrate; system-architecture (topics)
- **verify-later:** docs/basic_usage originals; current deployment list

### DEV-084 — Guidelines audit (001/002/003 compliance)
- **status:** convention
- **status-evidence:** "Read the dev guide, architecture, and contracts. Existing code: no violations"; a later audit action states "audited against 001/002/003 — code is COMPLIANT."
- **stage2-verified (2026-07-14):** deployed → convention — recurring compliance-audit practice, not a code artifact
- **what:** Recurring audits confirming a given engine/collector honours the house rules: standalone package main where required; JS content separation where applicable; parameterised SQL only; no logger.Debug; kebab-case/snake_case names; private same-file helpers allowed; sensitive keys never logged.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, #2026-06-13-d
- **relations:** STEP ZERO; house rules; wrapper-orchestrator finding this audit produced
- **verify-later:** 001 dev guide; 002 architecture; 003 contracts

### DEV-085 — workflow field-path audit query (jsonb_path_query over agent_definitions)
- **status:** unknown
- **status-evidence:** A single ad-hoc diagnostic query with no surrounding narrative or dated claim of use.
- **what:** A small but reusable diagnostic technique: a recursive `jsonb_path_query('$.**.<key>')` sweep over every `agent_definitions.default_config->'workflow'->'steps'` row to extract every field-path value referenced by a fixed set of workflow keys (`agent_type_field`, `default_from`, `content_field`, `iterate_over`, and any `*_from`/`*_field` wildcard key) across the whole workflow corpus at once — auditing real field-path usage in stored workflow JSON without opening each agent definition individually.
- **sources:** docs/_archive/agent_docs/sql_for_agents/sql_for_agents_v2/001_validator_sql.sql
- **relations:** canonical field-path resolution helpers (grep-before-using guidance)
- **verify-later:** none — a query technique, not a stored artifact

### DEV-086 — Workflow-as-configuration (JSON workflows in agent definitions)
- **status:** deployed
- **status-evidence:** A canonical `{"start_step": ..., "steps": {...}}` shape is documented; a HITL definition from the same era still uses exactly this workflow JSON structure with `next_step` chaining.
- **what:** Agent behaviour is a JSON workflow (start_step + named steps, each with an action, config, and next_step) stored in `agent_definitions.default_config`/task_workflow, overridable per agent_instances. Contrasted with Temporal/Airflow where workflows are compiled code — here business users can create workflows without deployment.
- **sources:** docs/architecture/002-agent-chassis-docs.md#how-workflows-work; docs/humanintheloop/hitl_agent_definition.sql; docs/architecture/012-investors.md#dynamic-workflow-creation
- **relations:** agent_definitions registry; execute_llm_prompt action; workflow lives in default_config
- **verify-later:** agent_definitions.task_workflow / orchestrator_workflow columns; workflow validator code

### DEV-087 — execute_llm_prompt generic action with DB prompt templates
- **status:** deployed
- **status-evidence:** Planned as "the reusable 'chef' that cooks the 'recipes'"; in live use by the HITL era — a workflow already uses `"action": "execute_llm_prompt"` with a Go-template prompt_template.
- **what:** A single generic action that reads the agent's prompt_template and ai_service config (provider, model, api_key_env_var) from its definition, renders the template with Go text/template placeholders (`{{.input_data.field}}`) filled from collected workflow data, calls the configured LLM, and returns the text. Makes every LLM agent a pure data configuration.
- **sources:** docs/basic_usage/003_dynamic_prompt_improvement#step-1.2; docs/humanintheloop/hitl_agent_definition.sql
- **relations:** workflow-as-configuration; prompt resolution priority hierarchy; dynamic prompt improvement loop
- **verify-later:** platform/orchestration/actions/ai_actions.go; prompt template rendering
