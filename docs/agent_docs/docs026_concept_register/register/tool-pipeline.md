# Register — tool-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

7 concepts, consolidated from 9 raw extractions across units U01, U02, U08, U10, U14, U15, U17a, U18.

### TP-001 — Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite → improve → audit)
- **status:** deployed
- **status-evidence:** "Pipeline Status: Fully Operational... all work without manual intervention" with per-site verified results; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously." Three independent documentation passes (docs024 narrative, focus-handoff review, SQL migrations) all describe the same steady-state pipeline.
- **what:** check_missing_tools / missing_tools discovery check auto-seeds add_tool items → tool-suggester (LLM judgment over specs+pages, 0-5 suggestions — zero is a valid answer — deciding library-vs-novel routing via check_is_library, carrying related_pages for cross-linking) → tool-deployer (library fork: component fork + tool page + page_component link + companion guide, then normal render/deploy) or tool-generator (novel LLM tool from brand context; writes a travelling PLAN since migration 131) → create_cross_links / create_tool_cross_link_items generates content_rewrite items per related page (item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (input_data.spec.suggestion) into the writer's nested loop prompt (the prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it, a documented trap) → rerender. Beyond the happy path: tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing complete the lifecycle loop. Known gaps recorded against this pipeline: no parse stage (source tools not read during adoption-style fidelity), and a specific historical report that interactive behaviour "reportedly currently don't work" (Path D, 2026-05-14) alongside fork-retry idempotency that was subsequently fixed (reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** 005_tool_pipeline(1).md full; 020_tool_lifecycle agents detail; FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2; 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql; 100_portfolio (autonomy claim)
- **relations:** de-tool hazard (tool-lifecycle TL-001); Fork-on-deploy (tool-library TLIB-001); Tool doc header (tool-lifecycle TL-007); games gap (copies this shape, games-lifecycle GML-001)
- **verify-later:** migrations 070–073 applied; cross-link items in prod; actual tool interactivity failures (Path D) current status; deploy_tool_to_site action; create_tool_cross_link_items

### TP-002 — Tool creation never enqueues the final page deploy (planned-pages gap)
- **status:** partial
- **status-evidence:** "Gap on record: tool creation ends at complete without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items in the interim).
- **what:** tool-generator creates component + page + nav but leaves the page build_status='planned'; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep touches them. Recorded follow-up: a create_rerender_item tail on tool-generator. This interacts directly with acceptance timing — it is the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only tool_acceptance_due sweep (tool-lifecycle TL-031).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design (tool-lifecycle TL-014, TL-031); build/rerender pipeline
- **verify-later:** whether tool-generator gained a rerender tail

### TP-003 — Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, never on the live deploy path
- **status:** partial
- **status-evidence:** separateInlineJS/collectJSAssets exist and are correct for the store path; the Tier-2 acceptance checker's first live sweep proved the deploy path ships JS INLINE and never references the asset — the "js-not-extracted" finding that led to the Option B delivered-reality supersede (tool-lifecycle TL-016).
- **what:** The store path's separateInlineJS extracts a bare inline `<script>` into js_content, nominally replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by collectJSAssets — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; a provocation-card component additionally truncated mid-script, since store validation checks unclosed `<style>` but not `<script>`, see tool-lifecycle TL-028). Meanwhile the generator/deploy route for brand-new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that live route. `deploy_page`'s files_field contract (TP-005) is the separate, load-bearing reason co-located JS must ship at all regardless of this extraction question. Hardening recorded: a script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle / Option B (tool-lifecycle TL-016); Mode-B rendered-artifact templates (tool-lifecycle TL-030); Planned assets-table split (tool-library TLIB-010, the earlier superseded design this partially realizes)
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships on the generator/deploy route

### TP-004 — Commented-out tool route and the planned-tool-page seam
- **status:** partial
- **status-evidence:** "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; "ON HOLD: coordination with the parallel tools chat... The interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision."
- **what:** The build-relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. a headline differentiator meant to become a tool) ship as prose via page-build-handler instead. Design fork recorded for the joint decision: (i) a thin tool-build-handler driving generation into the synced page (risks a page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no dedicated handler at all, a relay hop runs tool-suggester after site_plan and the existing tool pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3,#B4,#B5
- **relations:** work-item relay; tool pipeline (TP-001); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

### TP-005 — deploy_page files_field contract (co-located JS must ship)
- **status:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)... fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools historically stalled (one gas-unit-converter tool was built with real JS but stuck at build_status='pending'); the working-tools acceptance bar is deployed page + committed JS + resolving links, never merely "component generated."
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline; TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped, noted alongside); Inline-JS extraction (TP-003)
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites

### TP-006 — Reuse-checking retrieval architecture
- **status:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" — partially realised in the actual contextkit/code_symbols build.
- **what:** A framing that reuse-checking (for code or components) should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error — a missed match — leaves no trace. This is a general capability-reuse philosophy more than a tool-pipeline-specific mechanism; kept here as tagged, closest fit among the assigned categories.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure; Known-good solution library (tool-library TLIB-013)
- **verify-later:** n/a

### TP-007 — Toolchain validator + repo read/search (net-new for a self-coding pipeline)
- **status:** aspirational
- **status-evidence:** "Validator changes kind: from contract checks... to a toolchain validator... the most important new piece."
- **what:** Low-regret net-new pieces identified for a hypothetical self-coding pipeline: a toolchain validator giving ground-truth go build/vet/test + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer from the existing site-building pipeline. This is about a proposed self-development/coding pipeline rather than the site-tool pipeline; kept here as tagged, closest fit among the assigned categories.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2,#3,#4
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review
