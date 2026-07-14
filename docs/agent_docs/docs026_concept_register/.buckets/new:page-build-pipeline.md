
<!-- SOURCE: U03_idea_uk_section_data.md -->
### Rebuild vs rerender semantics and stale-render fossilisation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** CHECK 4 RESULTS: "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" — settling the 016-vs-026 documented tension "by direct evidence".
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler, status=triaged) re-runs plan_sections and re-renders everything. idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML (`var(--accent-color`), and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard from the migration sketch: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS #Migration-backfill; running_notes_scheme_to_components(55).md#So #Sh(migration route); HANDOFF_scheme_to_components_for_claude_code(1).md#Invariant (item 5)
- **relations:** dual chrome render paths; work-item crafting conventions; deployed-binary-predates-disk class.
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; 016/026 doc reconciliation.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### rerender-pages v6 workflow (refresh_site_components gate)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Tb): "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, which explains fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta #Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** dual chrome render paths; rebuild vs rerender semantics; work-item crafting conventions.
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### plan_sections field deferral semantics and needs_section_data escalation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (To): "Section BACK + both escalations self-closed = the deployed skip_field behaviour works"; W6.4: "two `needs_section_data` items in `needs_human_review` — `plan_sections` could not resolve `illustration_url` … and built each page WITHOUT the section."
- **what:** plan_sections resolves each schema field per its declared `source`; unresolvable required fields defer the WHOLE section, escalate a `needs_section_data` work item into needs_human_review, and the page builds without the section — a loud drop, not silent (guide refinement: fossil pages had been hiding the unresolved dependency). `on_missing: skip_field` is the established optional pattern: omit the field, let the template gate handle it. Edit A fixed the smell that a REQUIRED field with on_missing:skip_field fell to the default defer branch instead of honouring the declared intent. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** gobatch_01_plan_sections.md#Edit-A; RUNBOOK_scheme_to_components(50).md#W6.4 #W7-FINDINGS; running_notes_scheme_to_components(55).md#Tg #Tl #To; w6_05_section_data_read.sql
- **relations:** image fields optional-with-gate; section data source triad; deployed-binary-predates-disk.
- **verify-later:** plan_sections_action.go on_missing switch (required branch skip_field case present); needs_section_data item lifecycle.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Array item-fields prompt contract (019 migration + ItemFields)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu) 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched"; checkpoint (ss) documents the root cause and fragments verified at positions 2330/3402.
- **what:** Root cause of the differentiators empty cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — `title`/`body` against a template reading `name`/`description` renders empty; FAQ worked only because the natural guess happened to match. Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware (`[{ "k": "..." }]` for arrays). The migration is order-independent with the Go deploy ({{if .item_fields}} is simply false until populated), idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md; RUNBOOK_pcw_item_fields_fix.md
- **relations:** render-time item-key reconciler; component schema-template invariant; SQL change-management pattern.
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Render-time item-key reconciler (schema-sourced, non-fatal)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Checkpoint (uu): "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump for this specific change.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table (title/body → name/description etc.), never moving a synonym onto a key that is itself expected. Decision 1B hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change an optimisation, not a correctness requirement. Decision 2: unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data. Cross-file deploy constraint: rides the same image as plan_sections' extractArrayItemFields.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered; RUNBOOK_pcw_item_fields_fix.md#4-Logs
- **relations:** array item-fields contract; component schema-template invariant; needs_llm routing.
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys + wire-in; whether the carrying image shipped (log lines "reconcileGeneratedItemKeys" in writer pods).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### needs_llm routing via detectNeedsLLMContent
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (ss): "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because detectNeedsLLMContent returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative render_mode flip harmless to revert (differentiators back to 'template') and explains why a 'template' component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established #Correction-logged
- **relations:** section data source triad; array item-fields contract.
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### No component-level regeneration trigger (whole-page rebuild remedy)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn on hero, FAQ, narrative accepted as cost). Repeatedly parked on the hygiene/backlog lists; interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken; RUNBOOK_pcw_item_fields_fix.md#3
- **relations:** rebuild vs rerender semantics; content-governance (regeneration).
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary.

<!-- SOURCE: U05_content_quality_linking.md -->
### Re-render vs rebuild distinction (which path fixes what)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §2 "Render vs rebuild — what fixes what"; captured in 002/016 per NOTES(44).
- **what:** A load-bearing operational distinction: re-render (page-rerender / rerender-pages) re-applies templates to component data stored at last build; only a rebuild (work item → build-dispatch-loop → page-build-handler → writer) re-runs plan_sections source resolution and the resolver. Consequences: header/footer fixes need only re-render (data rebuilt fresh in Go); hero CTAs and hub URLs need rebuilds (stored data still carries phantoms); P4.2 proved page_rerender preserves sections but does NOT re-resolve schema-sourced CTAs.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22; running_notes_17(21).md#re-render-mechanics
- **relations:** no-LLM re-render path; interactive clobber (why rebuilds are dangerous); work-item routing.
- **verify-later:** page-rerender vs page-build-handler workflows; rerender_single_page_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-build-handler build path
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow chain confirmed repeatedly from live agent_definitions (HANDOFF_2026-06-09 Key references; NOTES(44) 2026-06-22 step-config dump).
- **what:** The per-page content build orchestrator: ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete. One linear flow, no item_type branch; deploys by spawning page-rerender + git commit, one commit per page. `spec.mode='recreate'` loads the adoption crawl to preserve original copy; `spec.suggestion` feeds writer rewrite_guidance.
- **sources:** HANDOFF_2026-06-09(2).md#key-references; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4; running_notes_17(21).md#page-build-handler-contract
- **relations:** page-content-writer; save_page_sections; silent-completion (complete_error exit); interactive clobber.
- **verify-later:** page-build-handler default_config.workflow.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer (task specialist, no persistence)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 9: writer def read — "no save_page_sections, no update_status, no deploy".
- **what:** The content-generation specialist: spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop (render/generate per section) → resolve_links → select_sections → compile_page. It only produces content (per-section outputs + compiled sections_metadata); persistence and deploy live in the page-build-handler wrapper — routing a discovery item straight at the writer can never deploy a page (a documented stale-handler bug in a dormant check). Its `complete` step's singular output_field was the Part-1 trigger.
- **sources:** running_notes_15(12).md#part-9; HANDOFF_2026-06-09(2).md#key-references; NOTES(44) writer key findings
- **relations:** result-contract; resolver wiring; recreate mode; prepare_link_context gap.
- **verify-later:** page-content-writer default_config; compile_page_sections_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### save_page_sections: DELETE+INSERT persistence with layered guards
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 Part 4 "DONE — patched save_page_sections (Layers 1+2) deployed on v1.0.1077".
- **what:** The single save path for page sections (three callers: page-build-handler, page-rerender, tool-recreation-handler): reads structured sections_metadata (primary) or an HTML-parse fallback (saveSectionsExtractFromHTML — extended with a single-fragment fallback after the `<div>`-not-`<section>` tool loss), snapshots page_component_history, then DELETE+INSERT of the produced set. Guards accreted through this unit: the content-regression guard (existing stripped text >200 and new < existing/4 → error — correct to refuse a wipe, threshold scales with page size); Layer 1 interactivity guard (existing page interactive, new set not → "interactivity regression blocked"); Layer 2 carry-forward of non-spec interactive sections (keep/replace/re-append by slot); source_item_id stamping into history via config-driven work_item_id_field.
- **sources:** NOTES(44) 2026-06-24 patch sessions; HANDOFF_2026-06-15(2).md#3; game_lost_its_tool/001_context; running_notes_17(21).md#index-save-read
- **relations:** interactive clobber; index stale-rebuild defect; save-failure visibility; content_data⊕resolved_data model.
- **verify-later:** save_page_sections_action.go (guards at ~L251-287, DELETE+INSERT ~L322-393, history ~L296-310); page_component_history.source_item_id population.

<!-- SOURCE: U05_content_quality_linking.md -->
### Interactive-page clobber failure class (spec-planned rebuild drops the tool)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 "Part 4 — DONE … game-pathfinding A* tool re-created (interactive ~20KB, deployed 2026-06-26) and now protected from re-clobber".
- **what:** An interactive tool/game exists ONLY as bespoke `<canvas>`/JS markup in page_components.rendered_html — not in the page spec, not LLM-regeneratable. ANY full rebuild (needs_page/needs_content_page/content_rewrite/link_resolution_rebuild/admin regenerate) plans from the spec, omits the tool, and save's DELETE+INSERT drops it (a links-only maintenance task destroyed a working A* game). Text-based regression guards missed it because the loss is markup/JS, not prose. Fix landed at the save path (Layers 1+2 above), NOT routing (P4.2 falsified the page_rerender reroute) and NOT the planner (which traffics in section-name skeletons). Interactivity signal: rendered_html ILIKE canvas/game-container/tool-page (data-component alone is not a signal). Prior partial fix: findPreservedComponentIDs preserved only render_action components.
- **sources:** PLAN_pathfinding_missing_game.md; NOTES(44) 2026-06-22 clobber sessions; game_lost_its_tool/001_context; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4
- **relations:** save_page_sections guards; tool-recreation-handler; item_key mis-key (same page); sectionHasVisibleContent (second silent-drop path).
- **verify-later:** page_component_history for game-pathfinding; save_page_sections Layer 1/2 code; regression test: link rebuild on pathfinding blocks not clobbers.

<!-- SOURCE: U05_content_quality_linking.md -->
### No-LLM re-render path (rerender_page_sections, Part 2 / Option Y)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 2 — DEPLOYED 2026-06-21; image_landed verified; … finish P2.4–P2.7".
- **what:** A field-re-resolve + re-render capability that avoids the full LLM writer: an image landing or resolvable section data previously forced a full content rebuild (LLM spend + regression-guard exposure). New rerender_page_sections action re-renders ALL of a page's sections from stored content_data overlaid with FRESH resolved_data (reusing plan_sections' side-effect-free planSection/sourceResolver — route ii), renders via RenderTemplate, emits the exact sections_metadata shape save reads. Slotted into page-rerender as a pre-pass gated by spec.reason (image_landed / section_data_resolved); flag_page_image_rebuild + reconcile_section_data repointed to emit page_rerender-type items (closing their type/key mismatch). NULL content_data on any section → escalate the whole page to needs_page (self-healing one-time full rebuild that backfills content_data). Y-lean render context chosen after confirming templates use only content_data + CSS-var colours. Design alternatives recorded: Option X (no-LLM branch inside the writer) rejected; re-render-affected-section-only rejected in favour of re-render-all.
- **sources:** NOTES(44)#part-A sections (decision trail 2026-06-19→21); RUNBOOK_gamesdesign_index_rebuild(29).md#part-2; HANDOFF_page_pipeline(11).md#5
- **relations:** content_data⊕resolved_data model; re-render vs rebuild; P4.2 (does NOT re-resolve schema-sourced CTAs — that stayed with the writer path).
- **verify-later:** rerender_page_sections_action.go; page-rerender check_rerender_mode wiring; P2.4–P2.7 test outcomes.

<!-- SOURCE: U05_content_quality_linking.md -->
### content_data ⊕ resolved_data persistence model
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19: "UNKNOWN NOW RESOLVED … content_data IS complete enough to re-render from" (RenderComponentAction deliberate merge, per its comment).
- **what:** RenderComponentAction builds a section's content_data as LLM copy (content_from) overlaid with resolved_data (merge_with) — deliberately persisting resolved items/urls/labels alongside the copy, next to rendered_html. This is what makes no-LLM re-rendering possible (render again from stored content + fresh resolution). Corollary schema fact that cost a wrong turn: there is NO page_components.resolved_data column — resolved values live inside content_data.
- **sources:** NOTES(44) 2026-06-19; HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_17(21).md#schema-correction
- **relations:** no-LLM re-render; system-stats key-contract break (content_data keys vs template keys).
- **verify-later:** v3_site_actions.go RenderComponentAction (~L1372); page_components schema.

<!-- SOURCE: U05_content_quality_linking.md -->
### Index stale-rebuild defect (writer output ≠ save input path)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "index rebuild VERIFIED on a real build … sections_metadata array, sm_count=5 … Part 1 result contract working".
- **what:** The unit's opening mystery: index rebuilds completed and git-committed while all five page_components stayed frozen at 06-06. The investigation is a model falsification chain — load, concurrent deploy, claim-lease duration, caller timeout, component locks, and the content-regression guard (writer measured at 33k chars >> the 5760 threshold) were each raised and eliminated — landing on the writer's compiled result being replaced by the size-limit stub before save (the Part-1 result-contract bug). Resolved by the flatten fix; verified end-to-end 06-19 and 06-24 (deployed hero "Your Probability Maths Is Wrong", real hub CTAs).
- **sources:** HANDOFF_2026-06-15_index_stale_rebuild(2).md; NOTES_gamesdesign_silent_norebuild(44).md; running_notes_17(21).md#index-deep-dive
- **relations:** result-contract resolution; silent-completion; save-failure visibility; "git committed ≠ new content" heuristic.
- **verify-later:** orchestration_states 472eed7d/4e0b339a; page_components index timestamps.

<!-- SOURCE: U05_content_quality_linking.md -->
### Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity
- **category:** NEW:page-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** page_build_handler_save_failure_visible.sql delivered 2026-06-15 with "unmet prerequisite (which error_step the engine reads)"; no later doc records applying it.
- **what:** Routes save_sections' error to a new mark_save_failed step (fail_work_item → needs_human_review) instead of complete_error, so a blocked/failed save surfaces instead of laundering into `complete`. Blocked on a real engine unknown: the save_sections step carries error_step in TWO places (step-level and config-level) and it is unconfirmed which the workflow engine honours for routing — "DO NOT GUESS". Companion (also unbuilt): gate deploy_page on sections_saved>0 so a no-write save can't re-commit stale components.
- **sources:** page_build_handler_save_failure_visible.sql; HANDOFF_2026-06-15(2).md#3-bugs; running_notes_17(21).md#FIX-written
- **relations:** silent-completion family; complete_error semantics (Fix B, deferred).
- **verify-later:** whether the SQL was ever applied; chassis engine error_step resolution (step.ErrorStep vs config["error_step"]).

<!-- SOURCE: U05_content_quality_linking.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-09(2): "Durability code WRITTEN this session (NOT yet deployed)"; running_notes_16(1) carries it as deploy-pending; later docs never record S1 enablement.
- **what:** A planned page reaching build with empty pages.sections silently completed as success ("Content writer skipped — page has no sections defined"). Three-layer durability: 2b — load_page_sections_from_spec gains a final fallback synthesising the layout from a same-role sibling's modal section list (layout skeleton only, WARN-logged, writes pages.sections); S1 — new discovery check check_sectionless_pages flags current-plan pages with empty sections that a sibling can fix and re-triggers to page-build-handler; S2 — check_has_ready_sections ELSE repointed from complete_error to mark_no_sections (needs_human_review). Decisive build fact: pages.sections is the build-read field; site_plan_sections is NOT on the build path (plan hygiene only). Also documented: checkEmptyPageSections is dormant, half-superseded code (wrapper never enabled, wrong handler) — a dedicated check was chosen over reviving it.
- **sources:** running_notes_15(12).md (whole arc); package_module/running_notes_16_adoption_sections.md (same content); HANDOFF_2026-06-09(2).md
- **relations:** Fix A prerequisite; silent-completion; skinner-box case; complete_error semantics.
- **verify-later:** load_page_sections_from_spec_action.go 2b fallback deployed?; completeness-discovery-agent checks array contains "sectionless_pages"?; page-build-handler mark_no_sections step.

<!-- SOURCE: U05_content_quality_linking.md -->
### plan_sections field-source resolution semantics (on_missing, required, defer)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14p "RESOLVED 2026-06-06" with code-confirmed semantics.
- **what:** The engine semantics governing when a section renders, defers, or drops a field: query.* fields return a non-nil empty slice (never defer, never consult on_missing); on_missing defaults to skip_field; the REQUIRED-field switch has NO skip_field case, so a required field defaulting to skip_field falls to defer — the trap that hid the guides hub (guide-list cta_url required=true + unpopulated spec source deferred the whole section). Fix chosen at the component (required=false) not the engine (the defer-for-safety default is defensible). Related deferral machinery: needs_section_data items for query-resolvable gaps, with reconcile_section_data as the designed loop-closer — registered but STILL UNHOSTED (nothing calls it; query-resolvable items sit at needs_human_review).
- **sources:** running_notes_14(26).md#part-14o-14p; HANDOFF_2026-06-09(2).md#june-02-actions; running_notes_16_content_quality_and_internal_linking(1).md#carried-forward
- **relations:** B4/B5 (query fields + template gates); no-LLM re-render (reuses planSection); component schema contracts.
- **verify-later:** plan_sections_action.go on_missing switches; reconcile_section_data host (still none?); guide-list/blog-listing required flags.

<!-- SOURCE: U05_content_quality_linking.md -->
### sectionHasVisibleContent assembler filter
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "approx_visible_len = 0 … the filter correctly drops it"; "the filter is right".
- **what:** rerender_single_page's getPageSections strips style/script/tags/entities and DROPS any section with ≤10 visible chars (WARN-only). Verified correct for text-empty shells (system-stats), but recognised as a SECOND silent-drop path for interactive content independent of save_page_sections — a low-prose game could be stripped at assembly even after the carry-forward preserves it in the DB. Open question noted: should it share the Part-4 interactivity signal rather than a pure text heuristic (the same text-heuristic blind spot as the regression guard).
- **sources:** NOTES(44) 2026-06-24 system-stats/assembler sessions
- **relations:** interactive clobber; system-stats break; text-heuristic blind spot family.
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent; game-auto-battler visible_len.

<!-- SOURCE: U05_content_quality_linking.md -->
### A1 — adopted tools/games never deployed a file (parser + status-churn chain)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 10: "A1 VERIFIED CLOSED … all five games committed … tools deploy".
- **what:** No tool/game page produced a deployable file because saveSectionsExtractFromHTML extracted only `<section>` blocks while recreate_tool emits `<div class="tool-page">` → zero page_components → assemblePage returned "" → rerender skipped → no git commit, all while work items read complete. Fixed by the single-fragment fallback (whole fragment as one section, guarded against full documents), coupled with the deployed→needs_rebuild flip removal and deploy-time plan-version stamping. Established the durable read: getPageSections reads page_components, not pages.sections — "has sections" and "has rendered components" are different facts.
- **sources:** running_notes_14(26).md#part-7-10
- **relations:** save_page_sections; interactive clobber (later same-family loss); tool-recreation-handler.
- **verify-later:** saveSectionsExtractFromHTML fallback; deployed repo /tools//games/ trees.

<!-- SOURCE: U05_content_quality_linking.md -->
### UpdatePageStatusAction zero-component deploy guard (Option B)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2 "Option B delivered" + Part 11/12 arc; reaper comment later cites the hardening as in place.
- **what:** A page must never be marked deployed with zero real components: the deployed branch is guarded by pageHasComponents (EXISTS on page_components with non-null component_id + non-empty rendered_html); on zero components it refuses `deployed`, sets needs_rebuild + clears the plan stamp, fail-open on check errors. Keeps build_status honest as evidence for the reaper (the homepage had been 'deployed' with 0 components and no file).
- **sources:** running_notes_14(26).md#part-11-12
- **relations:** evidence-gated reaper; auto-complete false positive; built_from_plan_version stamp.
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch.

<!-- SOURCE: U05_content_quality_linking.md -->
### Deploy-observability bookkeeping gap
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** NOTES(44) 2026-06-21: "Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at NULL though deployed_at is set — deploy step isn't writing those back."
- **what:** The deploy path sets deployed_at but never writes page_components.deploy_commit or pages.last_built_at, and content_hash is empty on investigated pages — so change detection falls back to updated_at + rendered_html length. Folded into a later deploy-observability fix; small but it repeatedly complicated verification.
- **sources:** NOTES(44) 2026-06-21 update; running_notes_17(21).md (content_hash note)
- **relations:** debugging heuristics (git committed ≠ new content); save_page_sections.
- **verify-later:** deploy_page/git_commit write-backs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### render_mode derivation + LLM routing condition (migration 002)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24 ~15:00; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has source='llm'. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — llm_field_specs is.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded + #migration-002-outcome + #2026-07-02-~19:20; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002
- **relations:** render_mode sweep (dropped); plan_sections deferral (render_mode red herring)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component-table render_mode sweep (65 components) — dropped migration
- **category:** NEW:page-build-pipeline
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_vonc.md base vs (1) diff: "Migration 002 (render_mode sweep across 65 components) is DROPPED"; PLAN_vonc_next_steps(1): "The 65-component render_mode update is DROPPED; existing components are fine as-is."
- **what:** The first plan for fixing LLM routing was a DB sweep updating `render_mode` on 65 existing library components. Dropped once it was established that workflow routing reads `llm_field_specs` (set by plan_sections from the schema), not the stored render_mode — so only the agent_definition condition needed fixing and existing component rows were fine as-is. Captures the earliest documented shape of the fix; useful provenance for why component rows still carry historical render_mode values.
- **sources:** docs/RUNNING_NOTES_vonc.md#4 (pre-edit base); docs/PLAN_vonc_next_steps(1).md#p1; docs/RUNBOOK_vonc_migrations(1).md (earlier "Fix render_mode on components" migration heading, dropped from later versions)
- **relations:** render_mode derivation + routing condition (the replacement)
- **verify-later:** none (historical)

<!-- SOURCE: U23_docs_root_vonc.md -->
### plan_sections readiness triage and deferral semantics
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Confirmed in code 2026-07-02 (planSection read); fix validated end-to-end 2026-07-03 (index went 3 → 6 sections after populating cta spec + relaxing illustration field); 016b §9 entry.
- **what:** plan_sections classifies each planned section by resolving its schema fields: source=llm always available; query.*/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the on_missing switch, whose `default:` case DEFERS the section ("default to defer for safety") — and empty on_missing defaults to skip_field, which is not a case in the required switch, so it defers. save_page_sections then persists only the ready set, dropping deferred sections' page instances. Authoring rule: never `required=true` + `on_missing=skip_field`; fix by populating the site data source or degrading the field.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:20 + #2026-07-02-~19:35; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f
- **relations:** site_specs cta aspect; resolver asset kinds gap; plan-driven rebuild + clobber
- **verify-later:** plan_sections_action.go planSection on_missing switch; save_page_sections_action.go

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan-driven rebuild + interactive/deferred-section clobber (carry-forward fix)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** 016b: "Part 4... fix WRITTEN (un-deployed)" — Layer 1 interactivity guard + Layer 2 carry-forward in patched save_page_sections_action.go; the 2026-07-02 vonc rebuild demonstrated the drop live (6 planned → 3 saved, brief-explanation instance gone).
- **what:** A needs_page rebuild is PLAN-driven, not pending-driven: load sections from the plan → triage → the writer renders ALL ready planned sections → save_page_sections DELETE+INSERTs the page's components. Sections present in page_components but absent from the plan (interactive tools stored only as rendered_html) or deferred by triage get silently dropped. Fix (written, not deployed): interactivity-aware guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections; three callers to bump (page-build-handler, page-rerender, tool-recreation-handler); plus source_item_id stamping for traceability.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + 2026-06-24 update); docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:40 + #2026-07-02-~19:00; docs/PLAN_spark_provocation_pipeline.md#standing-constraints
- **relations:** plan_sections deferral; page_components single-writer (save_page_sections); interactive tool pages stored as rendered_html
- **verify-later:** save_page_sections_action.go (is the guard/carry-forward deployed?); page_component_history.source_item_id

<!-- SOURCE: U23_docs_root_vonc.md -->
### complete_error silent-success family (page build completes having built nothing)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog item 1 in the HANDOFF, not built.
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a complete_workflow with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33–65s completes) hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (healthy: `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `last_built_at` is never written by build or rerender (dead column — write it or drop it).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed + #2026-07-08; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§9
- **relations:** three section sources; planner ≥1-section invariant; trust-the-artifact doctrine
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase

<!-- SOURCE: U23_docs_root_vonc.md -->
### load_page_record lookup semantics (name-first, page_id fallback)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by page_name against `pages.name` first, falling back to page_id only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch. Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family; three section sources
- **verify-later:** load_page_record_action.go nonPageNames list

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two re-render paths + assemble-only rerender distinction
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Doc-003-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header; light-path escalation rule quoted from 003.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — rerender_page_sections behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, and escalates the whole page to a full rebuild when content_data is NULL; (3) ASSEMBLE-ONLY — rerender_single_page (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Mode-B sections likely have NULL content_data, making the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/016b_debugging_guide_merged(3).md#open-threads (Part 2)
- **relations:** sanctioned edit paths; assemble-time visible-content filter; two chrome assembly paths
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Rebuild vs rerender semantics and stale-render fossilisation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** CHECK 4 RESULTS: "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" — settling the 016-vs-026 documented tension "by direct evidence".
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler, status=triaged) re-runs plan_sections and re-renders everything. idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML (`var(--accent-color`), and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard from the migration sketch: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS #Migration-backfill; running_notes_scheme_to_components(55).md#So #Sh(migration route); HANDOFF_scheme_to_components_for_claude_code(1).md#Invariant (item 5)
- **relations:** dual chrome render paths; work-item crafting conventions; deployed-binary-predates-disk class.
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; 016/026 doc reconciliation.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### rerender-pages v6 workflow (refresh_site_components gate)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Tb): "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, which explains fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta #Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** dual chrome render paths; rebuild vs rerender semantics; work-item crafting conventions.
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### plan_sections field deferral semantics and needs_section_data escalation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (To): "Section BACK + both escalations self-closed = the deployed skip_field behaviour works"; W6.4: "two `needs_section_data` items in `needs_human_review` — `plan_sections` could not resolve `illustration_url` … and built each page WITHOUT the section."
- **what:** plan_sections resolves each schema field per its declared `source`; unresolvable required fields defer the WHOLE section, escalate a `needs_section_data` work item into needs_human_review, and the page builds without the section — a loud drop, not silent (guide refinement: fossil pages had been hiding the unresolved dependency). `on_missing: skip_field` is the established optional pattern: omit the field, let the template gate handle it. Edit A fixed the smell that a REQUIRED field with on_missing:skip_field fell to the default defer branch instead of honouring the declared intent. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** gobatch_01_plan_sections.md#Edit-A; RUNBOOK_scheme_to_components(50).md#W6.4 #W7-FINDINGS; running_notes_scheme_to_components(55).md#Tg #Tl #To; w6_05_section_data_read.sql
- **relations:** image fields optional-with-gate; section data source triad; deployed-binary-predates-disk.
- **verify-later:** plan_sections_action.go on_missing switch (required branch skip_field case present); needs_section_data item lifecycle.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Array item-fields prompt contract (019 migration + ItemFields)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu) 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched"; checkpoint (ss) documents the root cause and fragments verified at positions 2330/3402.
- **what:** Root cause of the differentiators empty cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — `title`/`body` against a template reading `name`/`description` renders empty; FAQ worked only because the natural guess happened to match. Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware (`[{ "k": "..." }]` for arrays). The migration is order-independent with the Go deploy ({{if .item_fields}} is simply false until populated), idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md; RUNBOOK_pcw_item_fields_fix.md
- **relations:** render-time item-key reconciler; component schema-template invariant; SQL change-management pattern.
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Render-time item-key reconciler (schema-sourced, non-fatal)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Checkpoint (uu): "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump for this specific change.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table (title/body → name/description etc.), never moving a synonym onto a key that is itself expected. Decision 1B hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change an optimisation, not a correctness requirement. Decision 2: unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data. Cross-file deploy constraint: rides the same image as plan_sections' extractArrayItemFields.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered; RUNBOOK_pcw_item_fields_fix.md#4-Logs
- **relations:** array item-fields contract; component schema-template invariant; needs_llm routing.
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys + wire-in; whether the carrying image shipped (log lines "reconcileGeneratedItemKeys" in writer pods).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### needs_llm routing via detectNeedsLLMContent
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (ss): "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because detectNeedsLLMContent returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative render_mode flip harmless to revert (differentiators back to 'template') and explains why a 'template' component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established #Correction-logged
- **relations:** section data source triad; array item-fields contract.
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### No component-level regeneration trigger (whole-page rebuild remedy)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn on hero, FAQ, narrative accepted as cost). Repeatedly parked on the hygiene/backlog lists; interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken; RUNBOOK_pcw_item_fields_fix.md#3
- **relations:** rebuild vs rerender semantics; content-governance (regeneration).
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary.

<!-- SOURCE: U05_content_quality_linking.md -->
### Re-render vs rebuild distinction (which path fixes what)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §2 "Render vs rebuild — what fixes what"; captured in 002/016 per NOTES(44).
- **what:** A load-bearing operational distinction: re-render (page-rerender / rerender-pages) re-applies templates to component data stored at last build; only a rebuild (work item → build-dispatch-loop → page-build-handler → writer) re-runs plan_sections source resolution and the resolver. Consequences: header/footer fixes need only re-render (data rebuilt fresh in Go); hero CTAs and hub URLs need rebuilds (stored data still carries phantoms); P4.2 proved page_rerender preserves sections but does NOT re-resolve schema-sourced CTAs.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22; running_notes_17(21).md#re-render-mechanics
- **relations:** no-LLM re-render path; interactive clobber (why rebuilds are dangerous); work-item routing.
- **verify-later:** page-rerender vs page-build-handler workflows; rerender_single_page_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-build-handler build path
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow chain confirmed repeatedly from live agent_definitions (HANDOFF_2026-06-09 Key references; NOTES(44) 2026-06-22 step-config dump).
- **what:** The per-page content build orchestrator: ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete. One linear flow, no item_type branch; deploys by spawning page-rerender + git commit, one commit per page. `spec.mode='recreate'` loads the adoption crawl to preserve original copy; `spec.suggestion` feeds writer rewrite_guidance.
- **sources:** HANDOFF_2026-06-09(2).md#key-references; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4; running_notes_17(21).md#page-build-handler-contract
- **relations:** page-content-writer; save_page_sections; silent-completion (complete_error exit); interactive clobber.
- **verify-later:** page-build-handler default_config.workflow.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer (task specialist, no persistence)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 9: writer def read — "no save_page_sections, no update_status, no deploy".
- **what:** The content-generation specialist: spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop (render/generate per section) → resolve_links → select_sections → compile_page. It only produces content (per-section outputs + compiled sections_metadata); persistence and deploy live in the page-build-handler wrapper — routing a discovery item straight at the writer can never deploy a page (a documented stale-handler bug in a dormant check). Its `complete` step's singular output_field was the Part-1 trigger.
- **sources:** running_notes_15(12).md#part-9; HANDOFF_2026-06-09(2).md#key-references; NOTES(44) writer key findings
- **relations:** result-contract; resolver wiring; recreate mode; prepare_link_context gap.
- **verify-later:** page-content-writer default_config; compile_page_sections_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### save_page_sections: DELETE+INSERT persistence with layered guards
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 Part 4 "DONE — patched save_page_sections (Layers 1+2) deployed on v1.0.1077".
- **what:** The single save path for page sections (three callers: page-build-handler, page-rerender, tool-recreation-handler): reads structured sections_metadata (primary) or an HTML-parse fallback (saveSectionsExtractFromHTML — extended with a single-fragment fallback after the `<div>`-not-`<section>` tool loss), snapshots page_component_history, then DELETE+INSERT of the produced set. Guards accreted through this unit: the content-regression guard (existing stripped text >200 and new < existing/4 → error — correct to refuse a wipe, threshold scales with page size); Layer 1 interactivity guard (existing page interactive, new set not → "interactivity regression blocked"); Layer 2 carry-forward of non-spec interactive sections (keep/replace/re-append by slot); source_item_id stamping into history via config-driven work_item_id_field.
- **sources:** NOTES(44) 2026-06-24 patch sessions; HANDOFF_2026-06-15(2).md#3; game_lost_its_tool/001_context; running_notes_17(21).md#index-save-read
- **relations:** interactive clobber; index stale-rebuild defect; save-failure visibility; content_data⊕resolved_data model.
- **verify-later:** save_page_sections_action.go (guards at ~L251-287, DELETE+INSERT ~L322-393, history ~L296-310); page_component_history.source_item_id population.

<!-- SOURCE: U05_content_quality_linking.md -->
### Interactive-page clobber failure class (spec-planned rebuild drops the tool)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 "Part 4 — DONE … game-pathfinding A* tool re-created (interactive ~20KB, deployed 2026-06-26) and now protected from re-clobber".
- **what:** An interactive tool/game exists ONLY as bespoke `<canvas>`/JS markup in page_components.rendered_html — not in the page spec, not LLM-regeneratable. ANY full rebuild (needs_page/needs_content_page/content_rewrite/link_resolution_rebuild/admin regenerate) plans from the spec, omits the tool, and save's DELETE+INSERT drops it (a links-only maintenance task destroyed a working A* game). Text-based regression guards missed it because the loss is markup/JS, not prose. Fix landed at the save path (Layers 1+2 above), NOT routing (P4.2 falsified the page_rerender reroute) and NOT the planner (which traffics in section-name skeletons). Interactivity signal: rendered_html ILIKE canvas/game-container/tool-page (data-component alone is not a signal). Prior partial fix: findPreservedComponentIDs preserved only render_action components.
- **sources:** PLAN_pathfinding_missing_game.md; NOTES(44) 2026-06-22 clobber sessions; game_lost_its_tool/001_context; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4
- **relations:** save_page_sections guards; tool-recreation-handler; item_key mis-key (same page); sectionHasVisibleContent (second silent-drop path).
- **verify-later:** page_component_history for game-pathfinding; save_page_sections Layer 1/2 code; regression test: link rebuild on pathfinding blocks not clobbers.

<!-- SOURCE: U05_content_quality_linking.md -->
### No-LLM re-render path (rerender_page_sections, Part 2 / Option Y)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 2 — DEPLOYED 2026-06-21; image_landed verified; … finish P2.4–P2.7".
- **what:** A field-re-resolve + re-render capability that avoids the full LLM writer: an image landing or resolvable section data previously forced a full content rebuild (LLM spend + regression-guard exposure). New rerender_page_sections action re-renders ALL of a page's sections from stored content_data overlaid with FRESH resolved_data (reusing plan_sections' side-effect-free planSection/sourceResolver — route ii), renders via RenderTemplate, emits the exact sections_metadata shape save reads. Slotted into page-rerender as a pre-pass gated by spec.reason (image_landed / section_data_resolved); flag_page_image_rebuild + reconcile_section_data repointed to emit page_rerender-type items (closing their type/key mismatch). NULL content_data on any section → escalate the whole page to needs_page (self-healing one-time full rebuild that backfills content_data). Y-lean render context chosen after confirming templates use only content_data + CSS-var colours. Design alternatives recorded: Option X (no-LLM branch inside the writer) rejected; re-render-affected-section-only rejected in favour of re-render-all.
- **sources:** NOTES(44)#part-A sections (decision trail 2026-06-19→21); RUNBOOK_gamesdesign_index_rebuild(29).md#part-2; HANDOFF_page_pipeline(11).md#5
- **relations:** content_data⊕resolved_data model; re-render vs rebuild; P4.2 (does NOT re-resolve schema-sourced CTAs — that stayed with the writer path).
- **verify-later:** rerender_page_sections_action.go; page-rerender check_rerender_mode wiring; P2.4–P2.7 test outcomes.

<!-- SOURCE: U05_content_quality_linking.md -->
### content_data ⊕ resolved_data persistence model
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19: "UNKNOWN NOW RESOLVED … content_data IS complete enough to re-render from" (RenderComponentAction deliberate merge, per its comment).
- **what:** RenderComponentAction builds a section's content_data as LLM copy (content_from) overlaid with resolved_data (merge_with) — deliberately persisting resolved items/urls/labels alongside the copy, next to rendered_html. This is what makes no-LLM re-rendering possible (render again from stored content + fresh resolution). Corollary schema fact that cost a wrong turn: there is NO page_components.resolved_data column — resolved values live inside content_data.
- **sources:** NOTES(44) 2026-06-19; HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_17(21).md#schema-correction
- **relations:** no-LLM re-render; system-stats key-contract break (content_data keys vs template keys).
- **verify-later:** v3_site_actions.go RenderComponentAction (~L1372); page_components schema.

<!-- SOURCE: U05_content_quality_linking.md -->
### Index stale-rebuild defect (writer output ≠ save input path)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "index rebuild VERIFIED on a real build … sections_metadata array, sm_count=5 … Part 1 result contract working".
- **what:** The unit's opening mystery: index rebuilds completed and git-committed while all five page_components stayed frozen at 06-06. The investigation is a model falsification chain — load, concurrent deploy, claim-lease duration, caller timeout, component locks, and the content-regression guard (writer measured at 33k chars >> the 5760 threshold) were each raised and eliminated — landing on the writer's compiled result being replaced by the size-limit stub before save (the Part-1 result-contract bug). Resolved by the flatten fix; verified end-to-end 06-19 and 06-24 (deployed hero "Your Probability Maths Is Wrong", real hub CTAs).
- **sources:** HANDOFF_2026-06-15_index_stale_rebuild(2).md; NOTES_gamesdesign_silent_norebuild(44).md; running_notes_17(21).md#index-deep-dive
- **relations:** result-contract resolution; silent-completion; save-failure visibility; "git committed ≠ new content" heuristic.
- **verify-later:** orchestration_states 472eed7d/4e0b339a; page_components index timestamps.

<!-- SOURCE: U05_content_quality_linking.md -->
### Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity
- **category:** NEW:page-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** page_build_handler_save_failure_visible.sql delivered 2026-06-15 with "unmet prerequisite (which error_step the engine reads)"; no later doc records applying it.
- **what:** Routes save_sections' error to a new mark_save_failed step (fail_work_item → needs_human_review) instead of complete_error, so a blocked/failed save surfaces instead of laundering into `complete`. Blocked on a real engine unknown: the save_sections step carries error_step in TWO places (step-level and config-level) and it is unconfirmed which the workflow engine honours for routing — "DO NOT GUESS". Companion (also unbuilt): gate deploy_page on sections_saved>0 so a no-write save can't re-commit stale components.
- **sources:** page_build_handler_save_failure_visible.sql; HANDOFF_2026-06-15(2).md#3-bugs; running_notes_17(21).md#FIX-written
- **relations:** silent-completion family; complete_error semantics (Fix B, deferred).
- **verify-later:** whether the SQL was ever applied; chassis engine error_step resolution (step.ErrorStep vs config["error_step"]).

<!-- SOURCE: U05_content_quality_linking.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-09(2): "Durability code WRITTEN this session (NOT yet deployed)"; running_notes_16(1) carries it as deploy-pending; later docs never record S1 enablement.
- **what:** A planned page reaching build with empty pages.sections silently completed as success ("Content writer skipped — page has no sections defined"). Three-layer durability: 2b — load_page_sections_from_spec gains a final fallback synthesising the layout from a same-role sibling's modal section list (layout skeleton only, WARN-logged, writes pages.sections); S1 — new discovery check check_sectionless_pages flags current-plan pages with empty sections that a sibling can fix and re-triggers to page-build-handler; S2 — check_has_ready_sections ELSE repointed from complete_error to mark_no_sections (needs_human_review). Decisive build fact: pages.sections is the build-read field; site_plan_sections is NOT on the build path (plan hygiene only). Also documented: checkEmptyPageSections is dormant, half-superseded code (wrapper never enabled, wrong handler) — a dedicated check was chosen over reviving it.
- **sources:** running_notes_15(12).md (whole arc); package_module/running_notes_16_adoption_sections.md (same content); HANDOFF_2026-06-09(2).md
- **relations:** Fix A prerequisite; silent-completion; skinner-box case; complete_error semantics.
- **verify-later:** load_page_sections_from_spec_action.go 2b fallback deployed?; completeness-discovery-agent checks array contains "sectionless_pages"?; page-build-handler mark_no_sections step.

<!-- SOURCE: U05_content_quality_linking.md -->
### plan_sections field-source resolution semantics (on_missing, required, defer)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14p "RESOLVED 2026-06-06" with code-confirmed semantics.
- **what:** The engine semantics governing when a section renders, defers, or drops a field: query.* fields return a non-nil empty slice (never defer, never consult on_missing); on_missing defaults to skip_field; the REQUIRED-field switch has NO skip_field case, so a required field defaulting to skip_field falls to defer — the trap that hid the guides hub (guide-list cta_url required=true + unpopulated spec source deferred the whole section). Fix chosen at the component (required=false) not the engine (the defer-for-safety default is defensible). Related deferral machinery: needs_section_data items for query-resolvable gaps, with reconcile_section_data as the designed loop-closer — registered but STILL UNHOSTED (nothing calls it; query-resolvable items sit at needs_human_review).
- **sources:** running_notes_14(26).md#part-14o-14p; HANDOFF_2026-06-09(2).md#june-02-actions; running_notes_16_content_quality_and_internal_linking(1).md#carried-forward
- **relations:** B4/B5 (query fields + template gates); no-LLM re-render (reuses planSection); component schema contracts.
- **verify-later:** plan_sections_action.go on_missing switches; reconcile_section_data host (still none?); guide-list/blog-listing required flags.

<!-- SOURCE: U05_content_quality_linking.md -->
### sectionHasVisibleContent assembler filter
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "approx_visible_len = 0 … the filter correctly drops it"; "the filter is right".
- **what:** rerender_single_page's getPageSections strips style/script/tags/entities and DROPS any section with ≤10 visible chars (WARN-only). Verified correct for text-empty shells (system-stats), but recognised as a SECOND silent-drop path for interactive content independent of save_page_sections — a low-prose game could be stripped at assembly even after the carry-forward preserves it in the DB. Open question noted: should it share the Part-4 interactivity signal rather than a pure text heuristic (the same text-heuristic blind spot as the regression guard).
- **sources:** NOTES(44) 2026-06-24 system-stats/assembler sessions
- **relations:** interactive clobber; system-stats break; text-heuristic blind spot family.
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent; game-auto-battler visible_len.

<!-- SOURCE: U05_content_quality_linking.md -->
### A1 — adopted tools/games never deployed a file (parser + status-churn chain)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 10: "A1 VERIFIED CLOSED … all five games committed … tools deploy".
- **what:** No tool/game page produced a deployable file because saveSectionsExtractFromHTML extracted only `<section>` blocks while recreate_tool emits `<div class="tool-page">` → zero page_components → assemblePage returned "" → rerender skipped → no git commit, all while work items read complete. Fixed by the single-fragment fallback (whole fragment as one section, guarded against full documents), coupled with the deployed→needs_rebuild flip removal and deploy-time plan-version stamping. Established the durable read: getPageSections reads page_components, not pages.sections — "has sections" and "has rendered components" are different facts.
- **sources:** running_notes_14(26).md#part-7-10
- **relations:** save_page_sections; interactive clobber (later same-family loss); tool-recreation-handler.
- **verify-later:** saveSectionsExtractFromHTML fallback; deployed repo /tools//games/ trees.

<!-- SOURCE: U05_content_quality_linking.md -->
### UpdatePageStatusAction zero-component deploy guard (Option B)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2 "Option B delivered" + Part 11/12 arc; reaper comment later cites the hardening as in place.
- **what:** A page must never be marked deployed with zero real components: the deployed branch is guarded by pageHasComponents (EXISTS on page_components with non-null component_id + non-empty rendered_html); on zero components it refuses `deployed`, sets needs_rebuild + clears the plan stamp, fail-open on check errors. Keeps build_status honest as evidence for the reaper (the homepage had been 'deployed' with 0 components and no file).
- **sources:** running_notes_14(26).md#part-11-12
- **relations:** evidence-gated reaper; auto-complete false positive; built_from_plan_version stamp.
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch.

<!-- SOURCE: U05_content_quality_linking.md -->
### Deploy-observability bookkeeping gap
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** NOTES(44) 2026-06-21: "Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at NULL though deployed_at is set — deploy step isn't writing those back."
- **what:** The deploy path sets deployed_at but never writes page_components.deploy_commit or pages.last_built_at, and content_hash is empty on investigated pages — so change detection falls back to updated_at + rendered_html length. Folded into a later deploy-observability fix; small but it repeatedly complicated verification.
- **sources:** NOTES(44) 2026-06-21 update; running_notes_17(21).md (content_hash note)
- **relations:** debugging heuristics (git committed ≠ new content); save_page_sections.
- **verify-later:** deploy_page/git_commit write-backs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### render_mode derivation + LLM routing condition (migration 002)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24 ~15:00; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has source='llm'. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — llm_field_specs is.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded + #migration-002-outcome + #2026-07-02-~19:20; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002
- **relations:** render_mode sweep (dropped); plan_sections deferral (render_mode red herring)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component-table render_mode sweep (65 components) — dropped migration
- **category:** NEW:page-build-pipeline
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_vonc.md base vs (1) diff: "Migration 002 (render_mode sweep across 65 components) is DROPPED"; PLAN_vonc_next_steps(1): "The 65-component render_mode update is DROPPED; existing components are fine as-is."
- **what:** The first plan for fixing LLM routing was a DB sweep updating `render_mode` on 65 existing library components. Dropped once it was established that workflow routing reads `llm_field_specs` (set by plan_sections from the schema), not the stored render_mode — so only the agent_definition condition needed fixing and existing component rows were fine as-is. Captures the earliest documented shape of the fix; useful provenance for why component rows still carry historical render_mode values.
- **sources:** docs/RUNNING_NOTES_vonc.md#4 (pre-edit base); docs/PLAN_vonc_next_steps(1).md#p1; docs/RUNBOOK_vonc_migrations(1).md (earlier "Fix render_mode on components" migration heading, dropped from later versions)
- **relations:** render_mode derivation + routing condition (the replacement)
- **verify-later:** none (historical)

<!-- SOURCE: U23_docs_root_vonc.md -->
### plan_sections readiness triage and deferral semantics
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Confirmed in code 2026-07-02 (planSection read); fix validated end-to-end 2026-07-03 (index went 3 → 6 sections after populating cta spec + relaxing illustration field); 016b §9 entry.
- **what:** plan_sections classifies each planned section by resolving its schema fields: source=llm always available; query.*/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the on_missing switch, whose `default:` case DEFERS the section ("default to defer for safety") — and empty on_missing defaults to skip_field, which is not a case in the required switch, so it defers. save_page_sections then persists only the ready set, dropping deferred sections' page instances. Authoring rule: never `required=true` + `on_missing=skip_field`; fix by populating the site data source or degrading the field.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:20 + #2026-07-02-~19:35; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f
- **relations:** site_specs cta aspect; resolver asset kinds gap; plan-driven rebuild + clobber
- **verify-later:** plan_sections_action.go planSection on_missing switch; save_page_sections_action.go

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan-driven rebuild + interactive/deferred-section clobber (carry-forward fix)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** 016b: "Part 4... fix WRITTEN (un-deployed)" — Layer 1 interactivity guard + Layer 2 carry-forward in patched save_page_sections_action.go; the 2026-07-02 vonc rebuild demonstrated the drop live (6 planned → 3 saved, brief-explanation instance gone).
- **what:** A needs_page rebuild is PLAN-driven, not pending-driven: load sections from the plan → triage → the writer renders ALL ready planned sections → save_page_sections DELETE+INSERTs the page's components. Sections present in page_components but absent from the plan (interactive tools stored only as rendered_html) or deferred by triage get silently dropped. Fix (written, not deployed): interactivity-aware guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections; three callers to bump (page-build-handler, page-rerender, tool-recreation-handler); plus source_item_id stamping for traceability.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + 2026-06-24 update); docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:40 + #2026-07-02-~19:00; docs/PLAN_spark_provocation_pipeline.md#standing-constraints
- **relations:** plan_sections deferral; page_components single-writer (save_page_sections); interactive tool pages stored as rendered_html
- **verify-later:** save_page_sections_action.go (is the guard/carry-forward deployed?); page_component_history.source_item_id

<!-- SOURCE: U23_docs_root_vonc.md -->
### complete_error silent-success family (page build completes having built nothing)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog item 1 in the HANDOFF, not built.
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a complete_workflow with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33–65s completes) hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (healthy: `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `last_built_at` is never written by build or rerender (dead column — write it or drop it).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed + #2026-07-08; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§9
- **relations:** three section sources; planner ≥1-section invariant; trust-the-artifact doctrine
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase

<!-- SOURCE: U23_docs_root_vonc.md -->
### load_page_record lookup semantics (name-first, page_id fallback)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by page_name against `pages.name` first, falling back to page_id only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch. Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family; three section sources
- **verify-later:** load_page_record_action.go nonPageNames list

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two re-render paths + assemble-only rerender distinction
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Doc-003-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header; light-path escalation rule quoted from 003.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — rerender_page_sections behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, and escalates the whole page to a full rebuild when content_data is NULL; (3) ASSEMBLE-ONLY — rerender_single_page (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Mode-B sections likely have NULL content_data, making the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/016b_debugging_guide_merged(3).md#open-threads (Part 2)
- **relations:** sanctioned edit paths; assemble-time visible-content filter; two chrome assembly paths
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing
