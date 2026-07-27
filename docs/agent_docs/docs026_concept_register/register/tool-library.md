# Register — tool-library

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

23 concepts, consolidated from 26 raw extractions across units U01, U02, U05, U08, U09, U12, U16, U17a, U18, U19, U20, U23, U24c, U25, plus one concept (TLIB-001) absorbing a duplicate entry originally tagged tool-lifecycle in unit U19.

### TLIB-001 — Fork-on-deploy tool ownership model
- **status:** deployed
- **status-evidence:** "This is deliberate. A bad library change shouldn't break ten sites simultaneously"; forked_from column + partial unique index on function scoped to canonical tools, plus a later constraint amendment fixing an add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL) — blueprints never referenced directly by pages. Deploying to a site forks a copy (forked_from = library id), referenced by page_components, and the site owns it from then on: library changes never cascade to forks; pushing improvements to already-deployed sites goes through per-site work items instead. Uniqueness of `function` applies only to ACTIVE LIBRARY ORIGINALS (WHERE component_level='tool' AND forked_from IS NULL AND is_active), so many site forks can share a function name; forks are only ever addressed by component_id, never by function. Orphan-fork retry safety: a two-stage existing-fork check makes fork-deploy retries idempotent (reuse orphaned forks rather than duplicate); GetComponentByFunction excludes forks so lookups always resolve to the canonical original.
- **sources:** 019_tool_library(2).md#Core Concept, #Bug history; 105 item 6 fix; 020 tool-deployer; docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL); Two divergent tool-creation paths (tool-lifecycle TL-003); Fork-divergence detection (tool-lifecycle TL-023)
- **verify-later:** deploy_tool_action.go two-stage check; deployer fork-copy code; fork counts per library tool

### TLIB-002 — Never load html_template in listing queries (storage discipline)
- **status:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + a documented query audit.
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

### TLIB-003 — Component selector/creator architecture: section_type vs function split, and the self-extending library narrative
- **status:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"; the fuller self-extending narrative (003d spark docs) shows component-creator demonstrably running in production but the scoring path's deployment state unverified.
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) is designed to pick the variant — but the live resolution path is still a direct function lookup, with the scorer "not exercised." No candidate → needs_new_component work item → component-creator LLM-generates against the full component contract prompt and stores with selection metadata (created_from='generated', quality NULL, usage 0). A quality feedback loop from the auditor creates a fitness landscape where good templates survive and spread, and second builds reuse everything — the aspiration is that new site types work without special-casing. Backward compatible: direct function lookup remains path 1 and is, per the more recent evidence, still the one actually exercised.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-solution,#Component-library-growth; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md
- **relations:** component regeneration (tool-lifecycle TL-026); component creation contract prompt (TLIB-011); component selector metadata (TLIB-016); component selector by functional requirement (TLIB-010, a related but never-built alternative)
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?; content_components columns section_type/suitable_site_types/usage_count/avg_quality_score; selector Go function; planner prompt

### TLIB-004 — Component-creator agent (observed-pattern section components) — deployment specifics
- **status:** deployed
- **status-evidence:** Context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); a regeneration workflow path was noted missing at the time ("component-creator only handles needs_new_component").
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known historical gap: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path but not deactivated-row resurrection (unique-name collisions still need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1,#Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability; StoreGeneratedComponentAction regeneration branch (component-lifecycle CLC-002)
- **verify-later:** component-creator workflow; store_generated regen path today

### TLIB-005 — system-stats component key-contract break (regen renames fields, dependents empty)
- **status:** deployed
- **status-evidence:** "concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE — flag the platform bug to its owners."
- **stage2-verified (2026-07-14):** partial → deployed — Incident itself is historical/closed (status-evidence: 'concluded... closed-by-removal'). The fix it motivated (CLC-003 F1 field-contract guard) is deployed and wired: store_generated_component_action.go:308-340 blocks regeneration when isRegeneration strands old schema fields (the exact stat_1_number->stat1_value c...
- **what:** A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the shared system-stats component renaming its schema fields (stat_1_number → stat1_value etc.), then re-rendered every dependent from its EXISTING un-migrated content_data — all 5 live instances (across multiple sites) went text-empty in one 16ms batch. The regen mechanism exists but doesn't migrate dependents' content_data on a field rename (this is exactly the class the later F1 field-contract guard, component-lifecycle CLC-003, was built to reject). Side findings: usage_count is a stale counter (claimed 22, live 5); component_versions is now populated by component-creator; a concurrent-chat co-management protocol was applied (freshness probes before any shared-component write).
- **sources:** NOTES(44) 2026-06-24 system-stats sessions; RUNBOOK_gamesdesign_index_rebuild(29).md#part-5; HANDOFF_page_pipeline(11).md#2
- **relations:** sectionHasVisibleContent filter (correctly hid the empty shell); writer↔component-schema binding; F1 field-contract guard (component-lifecycle CLC-003, the fix this incident motivated)
- **verify-later:** content_components fdd92ad4 current schema; the 5 broken instances on other sites; component-creator regen migration behaviour

### TLIB-006 — Tier-D list components: queryresolve + items-array contract (vs numbered-flat fabrication)
- **status:** deployed
- **status-evidence:** "v1 (inline resolution in plan_sections + merge in page-content-writer) is what's deployed"; tool-list verified end-to-end; game-list Tier-D migration "delivered… validated… live on commit" (2026-06-04); guide-list resolving after guide re-type; migration_game_list_tier_d.sql applied.
- **what:** The list-component contract: a Tier-D component sources an `items` array field with `source: query.<name>` (e.g. `query.pages_where_type:tool`) resolved by the `queryresolve` Go package (a registry of concrete SQL queries, hard cap 24/default 12, `status IN ('active','deployed')`) at plan_sections time — a deliberate contract change from an earlier aspirational "resolve at render time" design (no render-time template engine exists). Templates use `{{range .items}}`; the query DSL is code-registered, never LLM-written. This replaces the legacy numbered-flat anti-pattern (game1_title…game6_* all source:llm, which fabricated and duplicated entries). game-list was migrated to tool-list-parity (identical field vocabulary so the writer/merge path treats all lists the same); richness was deliberately simplified because `pages` carries only url/title/meta_description. guide-list was Tier-D but starved by a page_type vocabulary gap. A dedicated `directory-builder` agent (a thin wrapper over the same resolver, adding re-triggerability) remains the deferred v2 — hybrid chosen over inline-forever or agent-first.
- **sources:** running_notes_14(26).md#part-13; PLAN_b4_b5_hubs_and_link_resolver(3).md#problem; FOCUS_directory_builder_and_list_components.md; migration_game_list_tier_d.sql; FOCUS_component_schema_patterns.md
- **relations:** guide page_type; B4/B5 hub links; component schema contracts; needs_section_data semantics (improvement-loop IMP-017)
- **verify-later:** game-list_pre_037 input_schema/items source; queryresolve resolvePagesWhereType; queryresolve/queryresolve.go; tool-list (migration 041) and game-list/guide-list schemas in content_components

### TLIB-007 — create_tool_component updates in place by function; unique index covers active library originals
- **status:** deployed
- **status-evidence:** Side finding (same function re-run → one component row, same id); index predicate read directly from the database.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK constraint covers {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns; Fork-on-deploy tool ownership model (TLIB-001)
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique

### TLIB-008 — Component selector by functional requirement (never-built proposal)
- **status:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)" in an early design document; no later confirmation found anywhere.
- **what:** Proposed capability-based search over content_components — finding a component by what it does rather than by name/category — paired with section recipes. Never implemented; superseded in practice by the simpler section_type-keyed selector (TLIB-003).
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching; Component selector architecture (TLIB-003)
- **verify-later:** any capability/tag-based component search implementation in component_library.go

### TLIB-009 — Tag-based deterministic tool-to-site matching (matchToolToSite) — superseded
- **status:** superseded
- **status-evidence:** Archive `findMissingTools` matched via site-type/industry affinity plus a "universal tools" carve-out; live doc's Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's semantic_tags against a site's type/industry. This produced the documented failure mode and was replaced entirely by tool-suggester, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; Mandatory minimum tool-suggestion count (tool-lifecycle TL-021)
- **verify-later:** confirm matchToolToSite function/code path has actually been removed

### TLIB-010 — Planned assets-table template/JS split for large tools — superseded plan
- **status:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate assets-table file — "This isn't built yet." Live doc: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the assets table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on content_components populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract; component-creator pipeline; Inline-JS extraction Path 1 (tool-pipeline TP-003)
- **verify-later:** SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL

### TLIB-011 — component-creator (LLM component template generation) + CSS variable naming contract
- **status:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with a STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like `--primary-color` "are WRONG and will produce broken output because they are undefined in every deployed stylesheet." Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor (TLIB-015); component selector (TLIB-003)
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

### TLIB-012 — JS tools documentation and provenance gap
- **status:** aspirational
- **status-evidence:** "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic RAG path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open questions: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; Tool doc header system (tool-lifecycle TL-007)
- **verify-later:** where JS tool sources live; any tool docs collection

### TLIB-013 — Known-good solution library (proposed)
- **status:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW)... proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table."
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the improvement-loop reliability cascade's tier-1 reuse would draw on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, #6.3, #7.5
- **relations:** reliability cascade (improvement-loop IMP-032); authored-vs-derived context
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

### TLIB-014 — component-quality-auditor (library health scoring)
- **status:** deployed
- **status-evidence:** 102 definition + a one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop; component-quality-auditor auto-regeneration threshold (tool-lifecycle TL-027, a boundary bug in this same mechanism)
- **verify-later:** compute_component_quality scoring criteria

### TLIB-015 — Component library (content_components) — the base schema
- **status:** deployed
- **status-evidence:** Live pg_dump backup shows the full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; an earlier dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; Fork-on-deploy tool ownership model (TLIB-001); component selector metadata (TLIB-016)
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions

### TLIB-016 — Component selector metadata and scoring (schema)
- **status:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library (TLIB-015); component quality tracking (tool-lifecycle TL-024); Component selector architecture (TLIB-003)
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage

### TLIB-017 — Seeded standalone tool library
- **status:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** Fork-on-deploy tool ownership model (TLIB-001); CSS variable contract (TLIB-011)
- **verify-later:** library tool rows and their deployment forks on live sites

### TLIB-018 — system.internal canonical library site
- **status:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components,#migration-042
- **relations:** site_work_items queue; component regeneration via component-creator
- **verify-later:** sites row for system.internal; work items with that site_id

### TLIB-019 — Semantic component library with vector embeddings (superseded vision)
- **status:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `ORDER BY (clip_embedding <=> [vector])` queries; design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** An early vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it. Successors: content_components / tool-library + tool registry matching; the embeddings idea resurfaces later in a diagnosis-loop contextkit.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library
- **verify-later:** pgvector extension usage; any table with embedding columns from this era

### TLIB-020 — Intelligent fallback component matching (P1/P2/P3) — historical
- **status:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block); the Generic Text Block fallback component was INSERTed; mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor, TLIB-003)
- **verify-later:** assemble_from_library in registry; fallback rows in content_components

### TLIB-021 — In-House Forge — content_components with data-function semantic contract (historical ancestor)
- **status:** deployed
- **status-evidence:** content_components seeded early (Generic Text Block, Document Head with base CSS…); the current page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" — the table survived into the live pipeline.
- **what:** The platform's own component library from its earliest era: rows with name, function (semantic purpose), html_template with `{{placeholders}}`, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (live successor); component library (TLIB-015); Intelligent fallback (TLIB-020)
- **verify-later:** content_components schema now vs then; data-function attributes in current templates

### TLIB-022 — Shared component library semantics + field-set guard + neutral-base/fork rule
- **status:** deployed
- **status-evidence:** Guard read directly in code (store_generated_component, referencing the fdd92ad4 system-stats incident); shared-component clobber check VERDICT (2026-07-04): "no contamination; base is neutral"; rule recorded as standing across two independent units.
- **what:** Components with `forked_from IS NULL` form a cross-site SHARED library keyed by `function` (e.g. brief-explanation shared by three separate sites). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident, TLIB-005: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes. Direct SQL UPDATEs bypass both the guard and component_versions snapshots, so hand edits must snapshot manually and check whether a change is really single-site first. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD (not built).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check,#2026-07-04-verdict,#2026-07-04-lobby-grid-verified; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3,#§8; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-09; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#6.3
- **relations:** Component regeneration in place (tool-lifecycle TL-026); F1 field-contract guard (component-lifecycle CLC-003, the code-level mechanism); content-governance (voice leakage); Component-creator invocation contract (tool-lifecycle TL-029)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage/distribution

### TLIB-023 — intent-probe capture component
- **status:** deployed
- **status-evidence:** "intent-probe INSERTED into the live library (INSERT 0 1)... second run's INSERT 0 0 is the ON CONFLICT idempotency."
- **what:** A NEW content-library section (built after a survey found nothing reusable among 83 existing sections) rendering an invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the `{{range}}`-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection,#2026-06-11-component-live; traffic_probe_plan(11).md#p3
- **relations:** carries a requires-backend tag; hand-instanced for two pilot sites
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql
