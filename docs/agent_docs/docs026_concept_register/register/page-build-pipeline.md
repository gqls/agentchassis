# Register — page-build-pipeline

24 concepts, consolidated from 29 raw extractions across units U03, U05, U23.

### PBP-001 — Rebuild vs rerender semantics and stale-render fossilisation
- **status:** deployed
- **status-evidence:** "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" (idea.uk CHECK 4 evidence); RUNBOOK_linking_phantom_fixes(7) confirms the same distinction generically ("Render vs rebuild — what fixes what").
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler) re-runs plan_sections and re-renders everything. Consequences: header/footer fixes need only re-render (data rebuilt fresh in Go); hero CTAs and hub URLs need rebuilds (stored data still carries phantoms). idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML, and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool — see the interactive-clobber entry).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS; running_notes_scheme_to_components(55).md#So/#Sh; RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22
- **relations:** rerender fossilisation (rebuild-cascade register); no-LLM re-render path (PBP-013); dual chrome render paths
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; rerender_single_page_action.go

### PBP-002 — rerender-pages v6 workflow (refresh_site_components gate)
- **status:** deployed
- **status-evidence:** "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, explaining fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta/#Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** chrome refresh gating (rebuild-cascade register); rebuild vs rerender semantics (PBP-001)
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step

### PBP-003 — plan_sections field-source resolution semantics (on_missing, required, defer) + needs_section_data escalation
- **status:** deployed
- **status-evidence:** running_notes_14(26) Part 14p "RESOLVED 2026-06-06" with code-confirmed semantics; independently re-derived and confirmed on the vonc site 2026-07-02/03 (index went 3→6 sections after populating a cta spec + relaxing a required field) and again via idea.uk's differentiators investigation (W6.4).
- **what:** plan_sections classifies each planned section by resolving its schema fields: `source=llm` is always available; `query.*`/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the `on_missing` switch, whose default case DEFERS the whole section ("default to defer for safety") — and empty `on_missing` defaults to `skip_field`, which is NOT a case in the required-field switch, so it silently falls through to defer. This is the trap that hid entire hub sections (a required `cta_url` with an unpopulated spec source deferred the whole section). Unresolvable required fields escalate a `needs_section_data` work item into needs_human_review; the page builds without the section — a loud drop, not silent. `on_missing: skip_field` is the established optional pattern (omit the field, let the template gate handle it). Authoring rule: never `required=true` + `on_missing=skip_field` — fix by populating the site data source or degrading the field. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** running_notes_14(26).md#part-14o-14p; gobatch_01_plan_sections.md#Edit-A; docs/RUNNING_NOTES_vonc(36).md#2026-07-02; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred
- **relations:** section data source triad (site-plan-and-reconciler register, PLAN-009); save_page_sections layered guards (PBP-011); carry-forward path (rebuild-cascade register)
- **verify-later:** plan_sections_action.go on_missing switch (required branch has no skip_field case); needs_section_data item lifecycle

### PBP-004 — Array item-fields prompt contract (019 migration + ItemFields)
- **status:** deployed
- **status-evidence:** Checkpoint 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched."
- **what:** Root cause of empty list-item cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — a guess against a template reading different keys renders empty (FAQ worked only because the natural guess happened to match). Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware. The migration is order-independent with the Go deploy, idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md
- **relations:** render-time item-key reconciler (PBP-005); component schema-template invariant
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population

### PBP-005 — Render-time item-key reconciler (schema-sourced, non-fatal)
- **status:** partial
- **status-evidence:** Checkpoint: "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table, never moving a synonym onto a key that is itself expected. A later decision hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change (PBP-004) an optimisation, not a correctness requirement. Unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered
- **relations:** array item-fields contract (PBP-004); content_data ⊕ resolved_data model (PBP-014); needs_llm routing (PBP-006)
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys wire-in; whether the carrying image shipped

### PBP-006 — needs_llm routing via detectNeedsLLMContent
- **status:** deployed
- **status-evidence:** "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because `detectNeedsLLMContent` returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative `render_mode` flip harmless to revert and explains why a nominally "template" component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established
- **relations:** render_mode derivation + LLM routing condition (PBP-023); render-time item-key reconciler (PBP-005)
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config

### PBP-007 — No component-level regeneration trigger (whole-page rebuild remedy)
- **status:** deployed
- **status-evidence:** Checkpoint: "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn accepted as cost). Interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML rather than regenerating it).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken
- **relations:** rebuild vs rerender semantics (PBP-001); content-governance (regeneration)
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary

### PBP-008 — page-build-handler build path
- **status:** deployed
- **status-evidence:** Workflow chain confirmed repeatedly from live agent_definitions (HANDOFF_2026-06-09 key references; NOTES(44) 2026-06-22 step-config dump).
- **what:** The per-page content build orchestrator: ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete. One linear flow, no item_type branch; deploys by spawning page-rerender + git commit, one commit per page. `spec.mode='recreate'` loads the adoption crawl to preserve original copy; `spec.suggestion` feeds writer rewrite_guidance.
- **sources:** HANDOFF_2026-06-09(2).md#key-references; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4
- **relations:** page-content-writer (PBP-009); save_page_sections (PBP-011); complete_error family (PBP-020)
- **verify-later:** page-build-handler default_config.workflow

### PBP-009 — page-content-writer (task specialist, no persistence)
- **status:** deployed
- **status-evidence:** running_notes_15(12) Part 9: writer def read — "no save_page_sections, no update_status, no deploy."
- **what:** The content-generation specialist: spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop (render/generate per section) → resolve_links → select_sections → compile_page. It only produces content (per-section outputs + compiled sections_metadata); persistence and deploy live in the page-build-handler wrapper — routing a discovery item straight at the writer can never deploy a page (a documented stale-handler bug in a dormant check).
- **sources:** running_notes_15(12).md#part-9; HANDOFF_2026-06-09(2).md#key-references
- **relations:** page-build-handler build path (PBP-008); resolver wiring; recreate mode
- **verify-later:** page-content-writer default_config; compile_page_sections_action.go

### PBP-010 — Re-render vs rebuild distinction (which path fixes what)
- **status:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §2 "Render vs rebuild — what fixes what"; captured in 002/016 per NOTES(44); P4.2 proved page_rerender preserves sections but does NOT re-resolve schema-sourced CTAs.
- **what:** A load-bearing operational distinction, restated as an explicit design rule: re-render (page-rerender/rerender-pages) re-applies templates to component data stored at last build; only a rebuild (work item → build-dispatch-loop → page-build-handler → writer) re-runs plan_sections source resolution. This is the same underlying mechanism as PBP-001, documented separately as the actionable "which path fixes what" decision rule for maintenance work rather than as a specific investigation's finding.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22; running_notes_17(21).md#re-render-mechanics
- **relations:** rebuild vs rerender semantics (PBP-001); interactive clobber (PBP-012, why rebuilds are dangerous); work-item routing map (work-item-integrity register)
- **verify-later:** page-rerender vs page-build-handler workflows; rerender_single_page_action.go

### PBP-011 — save_page_sections: DELETE+INSERT persistence with layered guards
- **status:** deployed
- **status-evidence:** "DONE — patched save_page_sections (Layers 1+2) deployed on v1.0.1077."
- **what:** The single save path for page sections (three callers: page-build-handler, page-rerender, tool-recreation-handler): reads structured sections_metadata (primary) or an HTML-parse fallback (extended with a single-fragment fallback after a `<div>`-not-`<section>` tool loss), snapshots page_component_history, then DELETE+INSERT of the produced set. Guards accreted over the investigation: content-regression guard (existing stripped text >200 and new < existing/4 → error); Layer 1 interactivity guard (existing page interactive, new set not → blocked); Layer 2 carry-forward of non-spec interactive sections (keep/replace/re-append by slot); source_item_id stamping into history via config-driven work_item_id_field.
- **sources:** NOTES(44) 2026-06-24 patch sessions; HANDOFF_2026-06-15(2).md#3; HANDOFF_page_pipeline(11).md
- **relations:** interactive-page clobber (PBP-012); index stale-rebuild defect (PBP-015); save-failure visibility fix (PBP-016)
- **verify-later:** save_page_sections_action.go guards (~L251-287), DELETE+INSERT (~L322-393); page_component_history.source_item_id population

### PBP-012 — Interactive/deferred-section clobber on plan-driven rebuild + carry-forward fix
- **status:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 4 — DONE … game-pathfinding A* tool re-created (interactive ~20KB, deployed 2026-06-26) and now protected from re-clobber." The same failure class recurred independently on vonc (016b: "Part 4... fix WRITTEN (un-deployed)" then demonstrated live 2026-07-02: 6 planned → 3 saved sections).
- **what:** An interactive tool/game exists ONLY as bespoke `<canvas>`/JS markup in `page_components.rendered_html` — not in the page spec, not LLM-regeneratable. A `needs_page` rebuild is PLAN-driven, not pending-driven: the writer renders ALL ready planned sections and `save_page_sections` DELETE+INSERTs the page's components, so ANY full rebuild (needs_page/content_rewrite/link_resolution_rebuild/admin regenerate) plans from the spec, omits the tool or a deferred section, and the DELETE+INSERT drops it (a links-only maintenance task once destroyed a working A* game). Text-based regression guards missed it because the loss is markup/JS, not prose. Fix landed at the save path (PBP-011's Layer 1/2), NOT routing and NOT the planner. Interactivity signal: `rendered_html` ILIKE canvas/game-container/tool-page (data-component alone is not a signal). Prior partial fix `findPreservedComponentIDs` preserved only render_action components.
- **sources:** PLAN_pathfinding_missing_game.md; game_lost_its_tool/001_context; docs/016b_debugging_guide_merged(3).md#open-threads Part 4; docs/RUNNING_NOTES_vonc(36).md#2026-07-01/#2026-07-02
- **relations:** save_page_sections guards (PBP-011); item_key mis-key on the same page (work-item-integrity register); sectionHasVisibleContent second silent-drop path (PBP-019)
- **verify-later:** page_component_history for game-pathfinding; whether the guard/carry-forward is deployed on all three callers; page_component_history.source_item_id

### PBP-013 — No-LLM re-render path (rerender_page_sections, Part 2 / Option Y)
- **status:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 2 — DEPLOYED 2026-06-21; image_landed verified; … finish P2.4–P2.7."
- **what:** A field-re-resolve + re-render capability that avoids the full LLM writer: an image landing or resolvable section data previously forced a full content rebuild (LLM spend + regression-guard exposure). `rerender_page_sections` re-renders ALL of a page's sections from stored content_data overlaid with FRESH resolved_data (reusing plan_sections' side-effect-free planSection/sourceResolver), renders via RenderTemplate, emits the exact sections_metadata shape save reads. Slotted into page-rerender as a pre-pass gated by `spec.reason` (image_landed / section_data_resolved); NULL content_data on any section escalates the whole page to needs_page (self-healing one-time full rebuild that backfills content_data). Design alternatives rejected: a no-LLM branch inside the writer; re-render-affected-section-only (chosen re-render-all instead).
- **sources:** NOTES(44)#part-A sections; RUNBOOK_gamesdesign_index_rebuild(29).md#part-2; HANDOFF_page_pipeline(11).md#5
- **relations:** content_data ⊕ resolved_data model (PBP-014); rebuild vs rerender (PBP-001); two re-render paths (PBP-025); assemble-only vs section re-render (rebuild-cascade register)
- **verify-later:** rerender_page_sections_action.go; page-rerender check_rerender_mode wiring; P2.4-P2.7 test outcomes

### PBP-014 — content_data ⊕ resolved_data persistence model
- **status:** deployed
- **status-evidence:** NOTES(44) 2026-06-19: "UNKNOWN NOW RESOLVED … content_data IS complete enough to re-render from" (RenderComponentAction deliberate merge, per its comment).
- **what:** `RenderComponentAction` builds a section's `content_data` as LLM copy (content_from) overlaid with resolved_data (merge_with) — deliberately persisting resolved items/urls/labels alongside the copy, next to rendered_html. This is what makes no-LLM re-rendering possible (render again from stored content + fresh resolution). Corollary schema fact that cost a wrong turn: there is NO `page_components.resolved_data` column — resolved values live inside `content_data`.
- **sources:** NOTES(44) 2026-06-19; HANDOFF_2026-06-15(2).md#schema-corrections
- **relations:** no-LLM re-render path (PBP-013); render-time item-key reconciler (PBP-005)
- **verify-later:** v3_site_actions.go RenderComponentAction (~L1372); page_components schema

### PBP-015 — Index stale-rebuild defect (writer output ≠ save input path)
- **status:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "index rebuild VERIFIED on a real build … sections_metadata array, sm_count=5 … Part 1 result contract working."
- **what:** The opening mystery of a multi-week investigation: index rebuilds completed and git-committed while all five page_components stayed frozen at an old date. Model-falsification chain (concurrent deploy, claim-lease duration, caller timeout, component locks, content-regression guard) each raised and eliminated — landing on the writer's compiled result being silently replaced by a size-limit stub before save (the "result-contract" bug). Resolved by a flatten fix; verified end-to-end. Established the durable diagnostic read: `getPageSections` reads `page_components`, not `pages.sections` — "has sections" and "has rendered components" are different facts.
- **sources:** HANDOFF_2026-06-15_index_stale_rebuild(2).md; NOTES_gamesdesign_silent_norebuild(44).md
- **relations:** silent-completion family (work-item-integrity register); save-failure visibility fix (PBP-016); "git committed ≠ new content" heuristic
- **verify-later:** orchestration_states for the affected runs; page_components index timestamps

### PBP-016 — Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity
- **status:** aspirational
- **status-evidence:** page_build_handler_save_failure_visible.sql delivered 2026-06-15 with "unmet prerequisite (which error_step the engine reads)"; no later doc records applying it.
- **what:** Routes save_sections' error to a new `mark_save_failed` step (fail_work_item → needs_human_review) instead of `complete_error`, so a blocked/failed save surfaces instead of laundering into `complete`. Blocked on a real engine unknown: the save_sections step carries `error_step` in TWO places (step-level and config-level) and it is unconfirmed which the workflow engine honours for routing — "DO NOT GUESS." Companion (also unbuilt): gate `deploy_page` on `sections_saved>0` so a no-write save can't re-commit stale components.
- **sources:** page_build_handler_save_failure_visible.sql; HANDOFF_2026-06-15(2).md#3-bugs
- **relations:** complete_error silent-success family (PBP-020); silent-completion failure family (work-item-integrity register)
- **verify-later:** whether the SQL was ever applied; chassis engine error_step resolution (step.ErrorStep vs config["error_step"])

### PBP-017 — Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag)
- **status:** partial
- **status-evidence:** HANDOFF_2026-06-09(2): "Durability code WRITTEN this session (NOT yet deployed)"; running_notes_16(1) carries it as deploy-pending; later docs never record S1 enablement.
- **what:** A planned page reaching build with empty `pages.sections` silently completed as success. Three-layer durability: 2b — `load_page_sections_from_spec` gains a final fallback synthesising the layout from a same-role sibling's section list (WARN-logged, writes pages.sections); S1 — new discovery check `check_sectionless_pages` flags current-plan pages with empty sections that a sibling can fix and re-triggers to page-build-handler; S2 — `check_has_ready_sections` ELSE repointed from `complete_error` to `mark_no_sections` (needs_human_review). Decisive build fact: `pages.sections` is the build-read field; `site_plan_sections` is NOT on the build path (plan hygiene only). Also documented: `checkEmptyPageSections` is dormant, half-superseded code (wrapper never enabled, wrong handler) — a dedicated check was chosen over reviving it.
- **sources:** running_notes_15(12).md (whole arc); HANDOFF_2026-06-09(2).md
- **relations:** complete_error family (PBP-020); three section sources (site-plan-and-reconciler register, PLAN-038)
- **verify-later:** load_page_sections_from_spec_action.go 2b fallback deployed?; completeness-discovery-agent "sectionless_pages" check; page-build-handler mark_no_sections step

### PBP-018 — render_mode derivation + LLM routing condition (migration 002)
- **status:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has `source='llm'`. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — `llm_field_specs` is. A prior plan to sweep `render_mode` across 65 existing library components was dropped once this was established (existing rows are fine as-is; only the agent_definition condition needed fixing).
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded/#migration-002-outcome; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002; docs/PLAN_vonc_next_steps(1).md#p1
- **relations:** needs_llm routing via detectNeedsLLMContent (PBP-006); plan_sections readiness triage (PBP-003)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

### PBP-019 — sectionHasVisibleContent assembler filter
- **status:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "approx_visible_len = 0 … the filter correctly drops it"; "the filter is right."
- **what:** `rerender_single_page`'s `getPageSections` strips style/script/tags/entities and DROPS any section with ≤10 visible chars (WARN-only). Verified correct for text-empty shells, but recognised as a SECOND silent-drop path for interactive content independent of `save_page_sections` — a low-prose game could be stripped at assembly even after the carry-forward preserves it in the DB. Open question: should it share the interactivity signal from PBP-012 rather than a pure text heuristic (the same text-heuristic blind spot as the content-regression guard).
- **sources:** NOTES(44) 2026-06-24 system-stats/assembler sessions
- **relations:** interactive/deferred-section clobber (PBP-012); text-heuristic blind spot family
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent; game-auto-battler visible_len

### PBP-020 — complete_error silent-success family (page build completes having built nothing)
- **status:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog, not built. Root instance also documented independently: `saveSectionsExtractFromHTML` extracted only `<section>` blocks while a recreate action emitted `<div class="tool-page">` → zero page_components → empty assembled page → no git commit — all while work items read complete (A1, now fixed by a single-fragment fallback).
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a `complete_workflow` with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33-65s completes) once hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (vs. healthy `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: a plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `pages.last_built_at` is never written by build or rerender (dead column).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; running_notes_14(26).md#part-7-10 (A1 parser instance)
- **relations:** planner ≥1-section invariant (site-plan-and-reconciler register, PLAN-040); silent-completion failure family (work-item-integrity register); load_page_record lookup semantics (PBP-021)
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase; deployed repo /tools//games/ trees

### PBP-021 — load_page_record lookup semantics (name-first, page_id fallback)
- **status:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by `page_name` against `pages.name` first, falling back to `page_id` only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch (PBP-020). Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name inconsistently.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family (PBP-020); three section sources (site-plan-and-reconciler register, PLAN-038)
- **verify-later:** load_page_record_action.go nonPageNames list

### PBP-022 — Two re-render paths + assemble-only rerender distinction
- **status:** deployed
- **status-evidence:** Doc-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — `rerender_page_sections` behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, escalating to a full rebuild when content_data is NULL (PBP-013); (3) ASSEMBLE-ONLY — `rerender_single_page` (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Sections with NULL content_data make the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4
- **relations:** no-LLM re-render path (PBP-013); rebuild vs rerender semantics (PBP-001); assemble-only vs section re-render (rebuild-cascade register)
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing

### PBP-023 — UpdatePageStatusAction zero-component deploy guard (Option B)
- **status:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2 "Option B delivered"; a reaper comment later cites the hardening as in place.
- **what:** A page must never be marked `deployed` with zero real components: the deployed branch is guarded by `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html); on zero components it refuses `deployed`, sets `needs_rebuild` + clears the plan stamp, fail-open on check errors. Keeps `build_status` honest as evidence for the reaper (the homepage had been 'deployed' with 0 components and no file). The same `pageHasComponents` gate is discussed independently in the work-item-integrity register as the "positive-evidence deploy guard" (same code, different investigation).
- **sources:** running_notes_14(26).md#part-11-12
- **relations:** positive-evidence deploy guard (work-item-integrity register); built_from_plan_version stamp (site-plan-and-reconciler register, PLAN-004); index stale-rebuild defect (PBP-015)
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch

### PBP-024 — Deploy-observability bookkeeping gap
- **status:** partial
- **status-evidence:** NOTES(44) 2026-06-21: "Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at NULL though deployed_at is set — deploy step isn't writing those back."
- **what:** The deploy path sets `deployed_at` but never writes `page_components.deploy_commit` or `pages.last_built_at`, and `content_hash` is empty on investigated pages — so change detection falls back to `updated_at` + rendered_html length. Folded into a later deploy-observability fix; small but it repeatedly complicated verification during the investigation.
- **sources:** NOTES(44) 2026-06-21 update; running_notes_17(21).md (content_hash note)
- **relations:** "git committed ≠ new content" debugging heuristic; save_page_sections (PBP-011); complete_error family (PBP-020, last_built_at dead column)
- **verify-later:** deploy_page/git_commit write-backs
