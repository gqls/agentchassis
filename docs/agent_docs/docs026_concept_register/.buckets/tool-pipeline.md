
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 005(1): "Pipeline Status: Fully Operational… all work without manual intervention" with per-site verified results
- **what:** check_missing_tools → tool-suggester (LLM judgement, 0-5 suggestions, library-vs-novel routing via check_is_library) → tool-deployer (fork) or tool-generator (novel) → create_cross_links (content_rewrite items per related page, item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (`input_data.spec.suggestion`) into the writer's nested loop prompt → rerender. The writer prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it (072 trap).
- **sources:** 005_tool_pipeline(1).md full; 020 agents detail
- **relations:** de-tool hazard; fork-on-deploy; tool doc header
- **verify-later:** migrations 070–073 applied; cross-link items in prod

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tools pipeline (suggest / deploy-fork / generate / improve / audit)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "Five agents, two discovery checks, full lifecycle documented … Tier 3 (headless browser visual testing) is planned, not built"; but Path D flags interactive behaviour "reportedly currently don't work" (2026-05-14)
- **what:** tool-suggester (LLM over spec aspects + library, 0-5 suggestions with library_source routing), tool-deployer (library fork with forked_from + tool page + companion guide), tool-generator (novel LLM tool, same wiring), tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing. Missing vs the four-stage pattern: no parse stage (source tools not read), loose source-tool fidelity. Fork-retry idempotency fixed (P2: reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2
- **relations:** games gap (copies this shape); library model; quality model
- **verify-later:** actual tool interactivity failures (Path D); tool_health/tool-auditor definitions

<!-- SOURCE: U08_travelling_docs.md -->
### Tool creation never enqueues the final page deploy (planned-pages gap)
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Gap on record: tool creation ends at `complete` without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items).
- **what:** tool-generator creates component + page + nav but leaves the page `build_status='planned'`; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep. Recorded follow-up: a `create_rerender_item` tail on tool-generator. Interacts with acceptance timing (the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only sweep).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design; build/rerender pipeline.
- **verify-later:** whether tool-generator gained a rerender tail.

<!-- SOURCE: U08_travelling_docs.md -->
### Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, not on the deploy path
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** 016b v3 entry (separateInlineJS/collectJSAssets exist and are correct for the store path); Tier-2 first sweep proved the deploy path ships JS INLINE and never references the asset (js-not-extracted, Option B superseded the criteria).
- **what:** The store path's `separateInlineJS` extracts a bare inline `<script>` into `js_content`, replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by `collectJSAssets` — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; provocation-card additionally truncated mid-script — store validation checks unclosed `<style>` but not `<script>`). Meanwhile the generator/deploy route for new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that route. Hardening recorded: script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle (Option B); empty-shell/mode-b categories; vonc case evidence.
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships.

<!-- SOURCE: U10_imagery.md -->
### deploy_page files_field contract (co-located JS must ship)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)… fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools stalled (gas-unit-converter built with real JS but stuck build_status='pending'); the working-tools acceptance is deployed page + committed JS + resolving links, never "component generated".
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline, TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped) noted alongside.
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites.

<!-- SOURCE: U14_docs019_runbooks.md -->
### Commented-out tool route and the planned-tool-page seam
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B3 "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; §B5 "ON HOLD: coordination with the parallel tools chat … The §B5 interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision".
- **what:** The relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. dartsonline's headline tool-setup-builder differentiator) ship as prose via page-build-handler. Design fork recorded for the joint decision: (i) thin tool-build-handler driving generation into the synced page (page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no handler, a relay hop runs tool-suggester after site_plan and its pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3 (C3); docs019/RUNBOOK_builder_route(21).md#B4 (§B5 candidate); docs019/RUNBOOK_builder_route(21).md#B5
- **relations:** work-item relay; tool pipeline (active suggester/generator/deployer); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

<!-- SOURCE: U15_docs019_running_notes.md -->
### Reuse-checking retrieval architecture
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" (principles(59) §Reuse-checking).
- **what:** A framing (partially realised in the actual contextkit/code_symbols build) that reuse-checking should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error (a missed match) leaves no trace.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking (finding code that already solves the problem).
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Toolchain validator + repo read/search (net-new for code)
- **category:** tool-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) §4 "Validator changes kind: from contract checks … to a toolchain validator … the most important new piece"
- **what:** Low-regret net-new pieces for a self-coding pipeline: a toolchain validator giving ground-truth `go build/vet/test` + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#3, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#4, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool pipeline: tool-suggester → tool-generator/tool-deployer → cross-linking
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 062/062b/098 definitions and patches; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously."
- **what:** tool-suggester (evaluate_tools handler) uses LLM judgment over specs+pages to decide which interactive tools would genuinely help a site (not limited to library catalogue), creating add_tool items; tool-deployer forks a library tool to the site (component fork + tool page + page_component link, then normal render/deploy); tool-generator creates new tool HTML from brand context (and since 131 writes a travelling PLAN); 098 adds cross-linking — suggestions carry related_pages, and create_tool_cross_link_items generates content_rewrite items so page-build-handler weaves tool references into existing copy. missing_tools discovery check auto-seeds add_tool items.
- **sources:** 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql
- **relations:** tool-library; tool acceptance tiers; travelling docs
- **verify-later:** deploy_tool_to_site action; create_tool_cross_link_items

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 005(1): "Pipeline Status: Fully Operational… all work without manual intervention" with per-site verified results
- **what:** check_missing_tools → tool-suggester (LLM judgement, 0-5 suggestions, library-vs-novel routing via check_is_library) → tool-deployer (fork) or tool-generator (novel) → create_cross_links (content_rewrite items per related page, item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (`input_data.spec.suggestion`) into the writer's nested loop prompt → rerender. The writer prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it (072 trap).
- **sources:** 005_tool_pipeline(1).md full; 020 agents detail
- **relations:** de-tool hazard; fork-on-deploy; tool doc header
- **verify-later:** migrations 070–073 applied; cross-link items in prod

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tools pipeline (suggest / deploy-fork / generate / improve / audit)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "Five agents, two discovery checks, full lifecycle documented … Tier 3 (headless browser visual testing) is planned, not built"; but Path D flags interactive behaviour "reportedly currently don't work" (2026-05-14)
- **what:** tool-suggester (LLM over spec aspects + library, 0-5 suggestions with library_source routing), tool-deployer (library fork with forked_from + tool page + companion guide), tool-generator (novel LLM tool, same wiring), tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing. Missing vs the four-stage pattern: no parse stage (source tools not read), loose source-tool fidelity. Fork-retry idempotency fixed (P2: reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2
- **relations:** games gap (copies this shape); library model; quality model
- **verify-later:** actual tool interactivity failures (Path D); tool_health/tool-auditor definitions

<!-- SOURCE: U08_travelling_docs.md -->
### Tool creation never enqueues the final page deploy (planned-pages gap)
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Gap on record: tool creation ends at `complete` without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items).
- **what:** tool-generator creates component + page + nav but leaves the page `build_status='planned'`; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep. Recorded follow-up: a `create_rerender_item` tail on tool-generator. Interacts with acceptance timing (the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only sweep).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design; build/rerender pipeline.
- **verify-later:** whether tool-generator gained a rerender tail.

<!-- SOURCE: U08_travelling_docs.md -->
### Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, not on the deploy path
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** 016b v3 entry (separateInlineJS/collectJSAssets exist and are correct for the store path); Tier-2 first sweep proved the deploy path ships JS INLINE and never references the asset (js-not-extracted, Option B superseded the criteria).
- **what:** The store path's `separateInlineJS` extracts a bare inline `<script>` into `js_content`, replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by `collectJSAssets` — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; provocation-card additionally truncated mid-script — store validation checks unclosed `<style>` but not `<script>`). Meanwhile the generator/deploy route for new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that route. Hardening recorded: script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle (Option B); empty-shell/mode-b categories; vonc case evidence.
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships.

<!-- SOURCE: U10_imagery.md -->
### deploy_page files_field contract (co-located JS must ship)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)… fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools stalled (gas-unit-converter built with real JS but stuck build_status='pending'); the working-tools acceptance is deployed page + committed JS + resolving links, never "component generated".
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline, TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped) noted alongside.
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites.

<!-- SOURCE: U14_docs019_runbooks.md -->
### Commented-out tool route and the planned-tool-page seam
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B3 "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; §B5 "ON HOLD: coordination with the parallel tools chat … The §B5 interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision".
- **what:** The relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. dartsonline's headline tool-setup-builder differentiator) ship as prose via page-build-handler. Design fork recorded for the joint decision: (i) thin tool-build-handler driving generation into the synced page (page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no handler, a relay hop runs tool-suggester after site_plan and its pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3 (C3); docs019/RUNBOOK_builder_route(21).md#B4 (§B5 candidate); docs019/RUNBOOK_builder_route(21).md#B5
- **relations:** work-item relay; tool pipeline (active suggester/generator/deployer); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

<!-- SOURCE: U15_docs019_running_notes.md -->
### Reuse-checking retrieval architecture
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" (principles(59) §Reuse-checking).
- **what:** A framing (partially realised in the actual contextkit/code_symbols build) that reuse-checking should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error (a missed match) leaves no trace.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking (finding code that already solves the problem).
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Toolchain validator + repo read/search (net-new for code)
- **category:** tool-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) §4 "Validator changes kind: from contract checks … to a toolchain validator … the most important new piece"
- **what:** Low-regret net-new pieces for a self-coding pipeline: a toolchain validator giving ground-truth `go build/vet/test` + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#3, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#4, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool pipeline: tool-suggester → tool-generator/tool-deployer → cross-linking
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 062/062b/098 definitions and patches; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously."
- **what:** tool-suggester (evaluate_tools handler) uses LLM judgment over specs+pages to decide which interactive tools would genuinely help a site (not limited to library catalogue), creating add_tool items; tool-deployer forks a library tool to the site (component fork + tool page + page_component link, then normal render/deploy); tool-generator creates new tool HTML from brand context (and since 131 writes a travelling PLAN); 098 adds cross-linking — suggestions carry related_pages, and create_tool_cross_link_items generates content_rewrite items so page-build-handler weaves tool references into existing copy. missing_tools discovery check auto-seeds add_tool items.
- **sources:** 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql
- **relations:** tool-library; tool acceptance tiers; travelling docs
- **verify-later:** deploy_tool_to_site action; create_tool_cross_link_items
