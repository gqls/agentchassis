
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fork-on-deploy tool ownership model
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019: "This is deliberate. A bad library change shouldn't break ten sites simultaneously."
- **what:** Library tools (component_level='tool', forked_from NULL) are blueprints never referenced by pages; deployment forks a copy per site (forked_from set) and the site owns it — library changes never cascade; pushing improvements to sites is per-site work items. Orphan-fork retry safety: two-stage existing-fork check (P105 fix); GetComponentByFunction excludes forks.
- **sources:** 019#Core Concept; 105 item 6 fix; 020 tool-deployer
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL)
- **verify-later:** deploy_tool_action.go two-stage check

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Never load html_template in listing queries (storage discipline)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + query audit section
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component selector / creator: section_type vs function split
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) picks the variant; no candidate → needs_new_component work item → component-creator generates against the full component contract prompt and stores with metadata; quality feedback loop from auditor scores creates a fitness landscape. Backward compatible: direct function lookup remains path 1.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path)
- **relations:** component regeneration; component creation contract prompt
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component-creator agent (observed-pattern section components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); regeneration workflow path noted missing ("component-creator only handles needs_new_component") 2026-04-17
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known gap at the time: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path (Track 2, 2026-04-20) but not deactivated-row resurrection (unique-name collisions need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability
- **verify-later:** component-creator workflow; store_generated regen path today

<!-- SOURCE: U05_content_quality_linking.md -->
### system-stats component key-contract break (regen renames fields, dependents empty)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE … flag the platform bug to its owners".
- **what:** A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the shared system-stats component renaming its schema fields (stat_1_number → stat1_value etc.), then re-rendered every dependent from its EXISTING un-migrated content_data — all 5 live instances (across multiple sites) went text-empty in one 16ms batch. The regen mechanism exists but doesn't migrate dependents' content_data on a field rename. Side findings: usage_count is a stale counter (claimed 22, live 5); component_versions is now populated by component-creator (future reverts possible; here only one version existed — no revert target); a concurrent-chat co-management protocol was applied (freshness probes before any shared-component write).
- **sources:** NOTES(44) 2026-06-24 system-stats sessions; RUNBOOK_gamesdesign_index_rebuild(29).md#part-5; HANDOFF_page_pipeline(11).md#2
- **relations:** sectionHasVisibleContent filter (correctly hid the empty shell); writer↔component-schema binding; component regeneration flow (026).
- **verify-later:** content_components fdd92ad4 current schema; the 5 broken instances on other sites; component-creator regen migration behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Tier-D list components (items array from real pages) vs numbered-flat fabrication
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 13 addendum: migration_game_list_tier_d.sql delivered/applied; Part 14g "pre-check shows game-list_pre_037, guide-list_pre_037, tool-list all query_sourced=t".
- **what:** The list-component contract: a Tier-D component sources `items` from query.pages_where_type:<type> (real realised pages), vs the legacy numbered-flat anti-pattern (game1_title…game6_* all source:llm) that fabricated and duplicated entries. game-list was migrated to tool-list-parity (identical field vocabulary so the writer/merge path treats all lists the same); richness deliberately simplified because pages carries only url/title/meta_description. guide-list was Tier-D but starved by the page_type vocabulary gap.
- **sources:** running_notes_14(26).md#part-13; PLAN_b4_b5_hubs_and_link_resolver(3).md#problem
- **relations:** guide page_type; B4/B5 hub links; component schema contracts (003).
- **verify-later:** game-list_pre_037 input_schema/items source; queryresolve resolvePagesWhereType.

<!-- SOURCE: U08_travelling_docs.md -->
### create_tool_component updates in place by function; unique index covers active library originals
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Side finding rev 33 (same function re-run → one component row, same id); index predicate read 2026-07-07.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns.
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique.

<!-- SOURCE: U09_adoption.md -->
### Tier-D list components: queryresolve + items-array contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** "v1 (inline resolution in plan_sections + merge in page-content-writer) is what's deployed" (FOCUS_directory_builder); tool-list verified end-to-end Step 3; game-list Tier-D migration "delivered… validated… live on commit" (2026-06-04); guide-list resolving after guide re-type.
- **what:** List/directory/grid components declare one `items` array field with `source: query.<name>` (e.g. `query.pages_where_type:tool`) resolved by the `queryresolve` Go package (registry of concrete SQL queries, hard cap 24/default 12, `status IN ('active','deployed')`) at plan_sections time — a deliberate contract change from doc 003's aspirational "at render time" (no render-time template engine exists). Templates use `{{range .items}}`; the query DSL is code-registered, never LLM-written. The dedicated `directory-builder` agent (doc 002 Phase 2 name) is the deferred v2 — a thin wrapper over the same resolver adding re-triggerability; hybrid chosen over inline-forever or agent-first.
- **sources:** FOCUS_directory_builder_and_list_components.md, migration_game_list_tier_d.sql, FOCUS_component_schema_patterns.md
- **relations:** numbered-flat anti-pattern (what it replaces); Step 2/3 anti-fabrication path; directory-builder v2 (aspirational)
- **verify-later:** queryresolve/queryresolve.go; tool-list (migration 041) and game-list/guide-list schemas in content_components

<!-- SOURCE: U12_docs024_archives.md -->
### Tag-based deterministic tool-to-site matching (matchToolToSite)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive `findMissingTools` matches via site-type/industry affinity plus a "universal tools" carve-out; live `020_tool_lifecycle(2).md` Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's `semantic_tags` against a site's type/industry. This produced the documented failure mode and was replaced entirely by `tool-suggester`, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; mandatory minimum tool-suggestion count (below)
- **verify-later:** confirm `matchToolToSite` function/code path has actually been removed.

<!-- SOURCE: U12_docs024_archives.md -->
### Planned assets-table template/JS split for large tools (superseded plan)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate `assets`-table file — "This isn't built yet." Live: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the `assets` table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on `content_components` populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract (003); component-creator pipeline
- **verify-later:** `SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Component selector by functional requirement
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)"; no later confirmation found.
- **what:** Proposed capability-based search over `content_components` — finding a component by what it does rather than by name/category — paired with section recipes.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching
- **verify-later:** any capability/tag-based component search implementation in `component_library.go`.

<!-- SOURCE: U16_docs019_design_plans.md -->
### JS tools documentation and provenance gap
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** FOCUS_js_tools_documentation: "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic rag path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; tool-doc header contract
- **verify-later:** where JS tool sources live; any tool docs collection

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Known-good solution library
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW) … proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table"
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the cascade's tier-1 reuse draws on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** reliability cascade; authored-vs-derived context (artifacts→docs arrow)
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### component-creator (LLM component template generation) + CSS variable naming contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with the STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like --primary-color "are WRONG and will produce broken output because they are undefined in every deployed stylesheet". Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor; component selector
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

<!-- SOURCE: U18_sql_for_agents.md -->
### component-quality-auditor (library health scoring)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 102 definition + one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop
- **verify-later:** compute_component_quality scoring criteria

<!-- SOURCE: U19_sql_tables_components.md -->
### Component library (content_components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Live pg_dump backup (005b) shows full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; 000 dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; tool fork model; component selector metadata.
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component selector metadata and scoring
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library; component quality tracking; site-plan sections resolution.
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Seeded standalone tool library
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** fork-on-deploy model; CSS variable contract.
- **verify-later:** library tool rows and their deployment forks on live sites.

<!-- SOURCE: U19_sql_tables_components.md -->
### system.internal canonical library site
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components and #migration-042
- **relations:** site_work_items queue; component regeneration via component-creator.
- **verify-later:** sites row for system.internal; work items with that site_id.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic component library with vector embeddings
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `SELECT … ORDER BY (clip_embedding <=> [vector])` queries (docs003); design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** Vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library + tool registry matching; embeddings idea resurfaces in contextkit (diagnosis-loop).
- **verify-later:** pgvector extension usage; any table with embedding columns from this era.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intelligent fallback component matching (P1/P2/P3)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block) and the Generic Text Block fallback component INSERTed (017); mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor).
- **verify-later:** assemble_from_library in registry; fallback rows in content_components.

<!-- SOURCE: U20_legacy_docs_a.md -->
### In-House Forge — content_components with data-function semantic contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** content_components seeded in 017 (Generic Text Block, Document Head with base CSS…); the *current* page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" (100_content_page_build_handler_flow.md) — the table survived into the live pipeline.
- **what:** The platform's own component library: rows with name, function (semantic purpose), html_template with {{placeholders}}, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (component contracts/slot specs — live successor); tool-library component library; intelligent fallback.
- **verify-later:** content_components schema now vs then; data-function attributes in current templates.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Shared component library semantics + field-set guard + neutral-base/fork rule
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Guard read in code (store_generated_component L315-335, referencing the fdd92ad4 incident); shared-component clobber check VERDICT 2026-07-04: "no contamination; brief-explanation base is neutral"; rule recorded as standing.
- **what:** Components with `forked_from IS NULL` are a cross-site SHARED library keyed by `function` (brief-explanation is shared by vonc + idea.uk + robot-hands). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule derived: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes; direct SQL UPDATEs bypass both the guard and component_versions snapshots. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check + #2026-07-04-verdict + #2026-07-04-lobby-grid-verified (store analysis); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** component regeneration in place; content-governance (voice leakage); section descriptor (fork-vs-base per site should live on the plan)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-probe capture component
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "intent-probe INSERTED into the live library (INSERT 0 1) … second run's INSERT 0 0 is the ON CONFLICT idempotency".
- **what:** A NEW content-library section (after STEP-ZERO found nothing reusable among 83 sections) rendering the invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the {{range}}-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_plan(11).md#p3, traffic_probe_running_notes(27).md#2026-06-11-component-live
- **relations:** carries requires-backend tag; hand-instanced for relojistas/wayfaringlondoner
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql

<!-- SOURCE: U25_leopardess_social.md -->
### Component selector + component creator (self-extending component library)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** component-creator demonstrably runs in production (NOTES_brief-explanation 083 regen 2026-07-01; archive-list creation 2026-07-06); the selector scoring/metadata design (suitable_site_types, usage_count, avg_quality_score, fitness landscape) is specified in 003d with a build sequence — deployment state of the scoring path unverified.
- **what:** New site types work without special-casing: the planner outputs section_types (structural need), the component selector queries content_components by section_type scoring site-type match + quality + usage, and a "no suitable component" result raises needs_new_component for the component-creator, which LLM-generates html_template + input_schema under the component contract and stores with selection metadata (created_from='generated', quality NULL, usage 0). Components then compete on audit scores — a fitness landscape where good templates survive and spread; second builds reuse everything.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-solution, #Component-library-growth; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md
- **relations:** component creation contract; component-creator invocation contract; shared component library semantics; tool-lifecycle
- **verify-later:** content_components columns section_type/suitable_site_types/usage_count/avg_quality_score; selector Go function; planner prompt

<!-- SOURCE: U25_leopardess_social.md -->
### Shared component library semantics (field-set guard, neutral base, fork rule)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation(5) 2026-07-09: "store_generated_component blocks a regeneration that DROPS or RENAMES a field on a shared component — the guard exists because of the fdd92ad4 system-stats incident"; brief-explanation verified shared across vonc + idea.uk + robot-hands with a neutral base.
- **what:** Components with forked_from IS NULL form a cross-site shared library keyed by function. Rules: regenerating a shared base is safe only for neutral, purely-additive changes (the field guard blocks drops/renames — renaming a field once silently emptied every dependent); site-specific voice must fork (forked_from = base_id); direct SQL UPDATEs bypass both the guard and component_versions snapshotting, so hand edits must snapshot manually and check D4-style "is it really single-site" first.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-09; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #5-D4; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#6.3
- **relations:** component selector/creator; three-per-row rule (shared grids untouchable); per-site style fork (same never-edit-shared principle)
- **verify-later:** store_generated_component field guard; forked_from distribution across content_components

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fork-on-deploy tool ownership model
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019: "This is deliberate. A bad library change shouldn't break ten sites simultaneously."
- **what:** Library tools (component_level='tool', forked_from NULL) are blueprints never referenced by pages; deployment forks a copy per site (forked_from set) and the site owns it — library changes never cascade; pushing improvements to sites is per-site work items. Orphan-fork retry safety: two-stage existing-fork check (P105 fix); GetComponentByFunction excludes forks.
- **sources:** 019#Core Concept; 105 item 6 fix; 020 tool-deployer
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL)
- **verify-later:** deploy_tool_action.go two-stage check

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Never load html_template in listing queries (storage discipline)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + query audit section
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component selector / creator: section_type vs function split
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) picks the variant; no candidate → needs_new_component work item → component-creator generates against the full component contract prompt and stores with metadata; quality feedback loop from auditor scores creates a fitness landscape. Backward compatible: direct function lookup remains path 1.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path)
- **relations:** component regeneration; component creation contract prompt
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component-creator agent (observed-pattern section components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); regeneration workflow path noted missing ("component-creator only handles needs_new_component") 2026-04-17
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known gap at the time: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path (Track 2, 2026-04-20) but not deactivated-row resurrection (unique-name collisions need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability
- **verify-later:** component-creator workflow; store_generated regen path today

<!-- SOURCE: U05_content_quality_linking.md -->
### system-stats component key-contract break (regen renames fields, dependents empty)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE … flag the platform bug to its owners".
- **what:** A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the shared system-stats component renaming its schema fields (stat_1_number → stat1_value etc.), then re-rendered every dependent from its EXISTING un-migrated content_data — all 5 live instances (across multiple sites) went text-empty in one 16ms batch. The regen mechanism exists but doesn't migrate dependents' content_data on a field rename. Side findings: usage_count is a stale counter (claimed 22, live 5); component_versions is now populated by component-creator (future reverts possible; here only one version existed — no revert target); a concurrent-chat co-management protocol was applied (freshness probes before any shared-component write).
- **sources:** NOTES(44) 2026-06-24 system-stats sessions; RUNBOOK_gamesdesign_index_rebuild(29).md#part-5; HANDOFF_page_pipeline(11).md#2
- **relations:** sectionHasVisibleContent filter (correctly hid the empty shell); writer↔component-schema binding; component regeneration flow (026).
- **verify-later:** content_components fdd92ad4 current schema; the 5 broken instances on other sites; component-creator regen migration behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Tier-D list components (items array from real pages) vs numbered-flat fabrication
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 13 addendum: migration_game_list_tier_d.sql delivered/applied; Part 14g "pre-check shows game-list_pre_037, guide-list_pre_037, tool-list all query_sourced=t".
- **what:** The list-component contract: a Tier-D component sources `items` from query.pages_where_type:<type> (real realised pages), vs the legacy numbered-flat anti-pattern (game1_title…game6_* all source:llm) that fabricated and duplicated entries. game-list was migrated to tool-list-parity (identical field vocabulary so the writer/merge path treats all lists the same); richness deliberately simplified because pages carries only url/title/meta_description. guide-list was Tier-D but starved by the page_type vocabulary gap.
- **sources:** running_notes_14(26).md#part-13; PLAN_b4_b5_hubs_and_link_resolver(3).md#problem
- **relations:** guide page_type; B4/B5 hub links; component schema contracts (003).
- **verify-later:** game-list_pre_037 input_schema/items source; queryresolve resolvePagesWhereType.

<!-- SOURCE: U08_travelling_docs.md -->
### create_tool_component updates in place by function; unique index covers active library originals
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Side finding rev 33 (same function re-run → one component row, same id); index predicate read 2026-07-07.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns.
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique.

<!-- SOURCE: U09_adoption.md -->
### Tier-D list components: queryresolve + items-array contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** "v1 (inline resolution in plan_sections + merge in page-content-writer) is what's deployed" (FOCUS_directory_builder); tool-list verified end-to-end Step 3; game-list Tier-D migration "delivered… validated… live on commit" (2026-06-04); guide-list resolving after guide re-type.
- **what:** List/directory/grid components declare one `items` array field with `source: query.<name>` (e.g. `query.pages_where_type:tool`) resolved by the `queryresolve` Go package (registry of concrete SQL queries, hard cap 24/default 12, `status IN ('active','deployed')`) at plan_sections time — a deliberate contract change from doc 003's aspirational "at render time" (no render-time template engine exists). Templates use `{{range .items}}`; the query DSL is code-registered, never LLM-written. The dedicated `directory-builder` agent (doc 002 Phase 2 name) is the deferred v2 — a thin wrapper over the same resolver adding re-triggerability; hybrid chosen over inline-forever or agent-first.
- **sources:** FOCUS_directory_builder_and_list_components.md, migration_game_list_tier_d.sql, FOCUS_component_schema_patterns.md
- **relations:** numbered-flat anti-pattern (what it replaces); Step 2/3 anti-fabrication path; directory-builder v2 (aspirational)
- **verify-later:** queryresolve/queryresolve.go; tool-list (migration 041) and game-list/guide-list schemas in content_components

<!-- SOURCE: U12_docs024_archives.md -->
### Tag-based deterministic tool-to-site matching (matchToolToSite)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive `findMissingTools` matches via site-type/industry affinity plus a "universal tools" carve-out; live `020_tool_lifecycle(2).md` Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's `semantic_tags` against a site's type/industry. This produced the documented failure mode and was replaced entirely by `tool-suggester`, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; mandatory minimum tool-suggestion count (below)
- **verify-later:** confirm `matchToolToSite` function/code path has actually been removed.

<!-- SOURCE: U12_docs024_archives.md -->
### Planned assets-table template/JS split for large tools (superseded plan)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate `assets`-table file — "This isn't built yet." Live: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the `assets` table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on `content_components` populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract (003); component-creator pipeline
- **verify-later:** `SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Component selector by functional requirement
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)"; no later confirmation found.
- **what:** Proposed capability-based search over `content_components` — finding a component by what it does rather than by name/category — paired with section recipes.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching
- **verify-later:** any capability/tag-based component search implementation in `component_library.go`.

<!-- SOURCE: U16_docs019_design_plans.md -->
### JS tools documentation and provenance gap
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** FOCUS_js_tools_documentation: "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic rag path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; tool-doc header contract
- **verify-later:** where JS tool sources live; any tool docs collection

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Known-good solution library
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW) … proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table"
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the cascade's tier-1 reuse draws on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** reliability cascade; authored-vs-derived context (artifacts→docs arrow)
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### component-creator (LLM component template generation) + CSS variable naming contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with the STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like --primary-color "are WRONG and will produce broken output because they are undefined in every deployed stylesheet". Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor; component selector
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

<!-- SOURCE: U18_sql_for_agents.md -->
### component-quality-auditor (library health scoring)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 102 definition + one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop
- **verify-later:** compute_component_quality scoring criteria

<!-- SOURCE: U19_sql_tables_components.md -->
### Component library (content_components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Live pg_dump backup (005b) shows full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; 000 dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; tool fork model; component selector metadata.
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component selector metadata and scoring
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library; component quality tracking; site-plan sections resolution.
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Seeded standalone tool library
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** fork-on-deploy model; CSS variable contract.
- **verify-later:** library tool rows and their deployment forks on live sites.

<!-- SOURCE: U19_sql_tables_components.md -->
### system.internal canonical library site
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components and #migration-042
- **relations:** site_work_items queue; component regeneration via component-creator.
- **verify-later:** sites row for system.internal; work items with that site_id.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic component library with vector embeddings
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `SELECT … ORDER BY (clip_embedding <=> [vector])` queries (docs003); design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** Vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library + tool registry matching; embeddings idea resurfaces in contextkit (diagnosis-loop).
- **verify-later:** pgvector extension usage; any table with embedding columns from this era.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intelligent fallback component matching (P1/P2/P3)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block) and the Generic Text Block fallback component INSERTed (017); mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor).
- **verify-later:** assemble_from_library in registry; fallback rows in content_components.

<!-- SOURCE: U20_legacy_docs_a.md -->
### In-House Forge — content_components with data-function semantic contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** content_components seeded in 017 (Generic Text Block, Document Head with base CSS…); the *current* page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" (100_content_page_build_handler_flow.md) — the table survived into the live pipeline.
- **what:** The platform's own component library: rows with name, function (semantic purpose), html_template with {{placeholders}}, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (component contracts/slot specs — live successor); tool-library component library; intelligent fallback.
- **verify-later:** content_components schema now vs then; data-function attributes in current templates.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Shared component library semantics + field-set guard + neutral-base/fork rule
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Guard read in code (store_generated_component L315-335, referencing the fdd92ad4 incident); shared-component clobber check VERDICT 2026-07-04: "no contamination; brief-explanation base is neutral"; rule recorded as standing.
- **what:** Components with `forked_from IS NULL` are a cross-site SHARED library keyed by `function` (brief-explanation is shared by vonc + idea.uk + robot-hands). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule derived: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes; direct SQL UPDATEs bypass both the guard and component_versions snapshots. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check + #2026-07-04-verdict + #2026-07-04-lobby-grid-verified (store analysis); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** component regeneration in place; content-governance (voice leakage); section descriptor (fork-vs-base per site should live on the plan)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-probe capture component
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "intent-probe INSERTED into the live library (INSERT 0 1) … second run's INSERT 0 0 is the ON CONFLICT idempotency".
- **what:** A NEW content-library section (after STEP-ZERO found nothing reusable among 83 sections) rendering the invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the {{range}}-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_plan(11).md#p3, traffic_probe_running_notes(27).md#2026-06-11-component-live
- **relations:** carries requires-backend tag; hand-instanced for relojistas/wayfaringlondoner
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql

<!-- SOURCE: U25_leopardess_social.md -->
### Component selector + component creator (self-extending component library)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** component-creator demonstrably runs in production (NOTES_brief-explanation 083 regen 2026-07-01; archive-list creation 2026-07-06); the selector scoring/metadata design (suitable_site_types, usage_count, avg_quality_score, fitness landscape) is specified in 003d with a build sequence — deployment state of the scoring path unverified.
- **what:** New site types work without special-casing: the planner outputs section_types (structural need), the component selector queries content_components by section_type scoring site-type match + quality + usage, and a "no suitable component" result raises needs_new_component for the component-creator, which LLM-generates html_template + input_schema under the component contract and stores with selection metadata (created_from='generated', quality NULL, usage 0). Components then compete on audit scores — a fitness landscape where good templates survive and spread; second builds reuse everything.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-solution, #Component-library-growth; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md
- **relations:** component creation contract; component-creator invocation contract; shared component library semantics; tool-lifecycle
- **verify-later:** content_components columns section_type/suitable_site_types/usage_count/avg_quality_score; selector Go function; planner prompt

<!-- SOURCE: U25_leopardess_social.md -->
### Shared component library semantics (field-set guard, neutral base, fork rule)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation(5) 2026-07-09: "store_generated_component blocks a regeneration that DROPS or RENAMES a field on a shared component — the guard exists because of the fdd92ad4 system-stats incident"; brief-explanation verified shared across vonc + idea.uk + robot-hands with a neutral base.
- **what:** Components with forked_from IS NULL form a cross-site shared library keyed by function. Rules: regenerating a shared base is safe only for neutral, purely-additive changes (the field guard blocks drops/renames — renaming a field once silently emptied every dependent); site-specific voice must fork (forked_from = base_id); direct SQL UPDATEs bypass both the guard and component_versions snapshotting, so hand edits must snapshot manually and check D4-style "is it really single-site" first.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-09; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #5-D4; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#6.3
- **relations:** component selector/creator; three-per-row rule (shared grids untouchable); per-site style fork (same never-edit-shared principle)
- **verify-later:** store_generated_component field guard; forked_from distribution across content_components
