# EXTRACTION U19 — sql_for_tables / sql_for_components / sql_for_tools / sql_for_hitl / sql_for_content / tables_sql
Extracted 2026-07-13. Files in scope: 70. Concepts found: 63.

## Coverage
| file | treatment |
|---|---|
| docs/agent_docs/sql_for_components/001_style_collections.sql | full |
| docs/agent_docs/sql_for_components/002_styles_documentation.md | full |
| docs/agent_docs/sql_for_components/003_styles_implementation.md | full |
| docs/agent_docs/sql_for_components/004_component_architecture_schema.sql | full |
| docs/agent_docs/sql_for_components/005_seed_default_components.sql | header-scan |
| docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql | header-scan |
| docs/agent_docs/sql_for_components/007_add_components.sql | header-scan |
| docs/agent_docs/sql_for_content/001_phone_number.sql | full |
| docs/agent_docs/sql_for_hitl/001_hitl_requests.sql | full |
| docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql | full |
| docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql | family-delta |
| docs/agent_docs/sql_for_tables/001_awaited_requests.sql | full |
| docs/agent_docs/sql_for_tables/003_pages.sql | full |
| docs/agent_docs/sql_for_tables/004_content_items.sql | full |
| docs/agent_docs/sql_for_tables/004b_content_items.md | full |
| docs/agent_docs/sql_for_tables/005_content_components.sql | header-scan |
| docs/agent_docs/sql_for_tables/005b_bk_content_components.sql | family-delta |
| docs/agent_docs/sql_for_tables/005c_bk_page_components.sql | family-delta |
| docs/agent_docs/sql_for_tables/006_input_requests.sql | full |
| docs/agent_docs/sql_for_tables/007_processed_messages.sql | full |
| docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql | full |
| docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql | full |
| docs/agent_docs/sql_for_tables/009_research_results.sql | full |
| docs/agent_docs/sql_for_tables/010_orchestration_state_audit.sql | full |
| docs/agent_docs/sql_for_tables/011_sites_table.sql | full |
| docs/agent_docs/sql_for_tables/012_site_components.sql | full |
| docs/agent_docs/sql_for_tables/014_site_areas.sql | full |
| docs/agent_docs/sql_for_tables/015_area_components.sql | full |
| docs/agent_docs/sql_for_tables/016_nav_tables.sql | full |
| docs/agent_docs/sql_for_tables/017_site_nav_groups.sql | full |
| docs/agent_docs/sql_for_tables/018_site_work_items.sql | full |
| docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql | full |
| docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql | full |
| docs/agent_docs/sql_for_tables/021_auth_db.sql | full |
| docs/agent_docs/sql_for_tables/022_agent_error_log.sql | full |
| docs/agent_docs/sql_for_tables/023_companies_house_data.sql | full |
| docs/agent_docs/sql_for_tables/024_database_cleanup.sql | full |
| docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql | full |
| docs/agent_docs/sql_for_tables/026_http_request_log.sql | full |
| docs/agent_docs/sql_for_tables/027_content_sources_table.sql | full |
| docs/agent_docs/sql_for_tables/028_vet_med_prices.sql | full |
| docs/agent_docs/sql_for_tables/029_vet_med_retailers.sql | full |
| docs/agent_docs/sql_for_tables/029b_vet_med_test_seed.sql | full |
| docs/agent_docs/sql_for_tables/030_vet_med_pricing_schema_migrations.sql | family-delta |
| docs/agent_docs/sql_for_tables/031_site_snapshots.sql | full |
| docs/agent_docs/sql_for_tables/032_business_intel_med_scrape_evidence | full |
| docs/agent_docs/sql_for_tables/033-business_intel.med_retailer_listings.sql | full |
| docs/agent_docs/sql_for_tables/034_vet_med_price_scrape_orchestrator.sql | full |
| docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql | full |
| docs/agent_docs/sql_for_tables/036_orchestration_states.sql | full |
| docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql | full |
| docs/agent_docs/sql_for_tables/038_style_collections.sql | full |
| docs/agent_docs/sql_for_tables/039_training_exports.sql | full |
| docs/agent_docs/sql_for_tables/040_site_plans_schema.sql | full |
| docs/agent_docs/sql_for_tables/041_assets.sql | full |
| docs/agent_docs/sql_for_tables/042_thunder.sql | full |
| docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql | full |
| docs/agent_docs/sql_for_tables/044_css_snippets.sql | full |
| docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql | full |
| docs/agent_docs/sql_for_tables/046_site_chat_turns.sql | full |
| docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql | full |
| docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql | full |
| docs/agent_docs/sql_for_tables/049_page_components_build_status_check.sql | full |
| docs/agent_docs/sql_for_tables/bk_site_specs.sql | family-delta |
| docs/agent_docs/sql_for_tools/001_initial_toolset.sql | header-scan |
| docs/agent_docs/sql_for_tools/002_tool_migration.sql | full |
| docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql | header-scan |
| docs/agent_docs/sql_for_tools/004_finance_tools_batch2.sql | header-scan |
| docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql | header-scan |
| docs/agent_docs/tables_sql/001_awaited_requests.sql | family-delta |

Notes: the three pg_dump backups (005b, 005c, bk_site_specs) had DDL, constraints and COMMENTs read in full; COPY data bodies skipped. 000 is a psql expanded-record dump of the pre-070-refactor component library (41 rows) — used for delta against the current library shape. 030 is a byte-duplicate of 028. tables_sql/001 is an earlier variant of sql_for_tables/001.

## Concepts

### Component library (content_components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Live pg_dump backup (005b) shows full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; 000 dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; tool fork model; component selector metadata.
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions.

### Component render modes (template | agent | composite | standalone)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Columns and comments exist (004 PART 2), but the 000 backup shows all 41 components render_mode='template'; only 'standalone' (tools) is additionally observed in seeds.
- **what:** render_mode declares how a component is produced: 'template' (direct substitution), 'agent' (spawn agent_type with optional agent_workflow, data pulled via data_sources dot-paths), 'composite' (child_components list), and later 'standalone' for tools (html_template IS the final output). The agent/composite modes appear designed but unexercised.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART2; docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component library; standalone tool render.
- **verify-later:** Go render path switch on render_mode; any rows with render_mode in ('agent','composite').

### Kebab-case naming contract (component function + pages.page_type)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** chk_function_kebab_case CHECK in the live dump; migration 051 adds chk_page_type_kebab_case with pre/post distribution audit; data-component attributes reconciled to function.
- **what:** Identifier-shaped values are kebab-case, enforced by CHECK constraints: content_components.function (regex `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`, empty allowed for legacy), pages.page_type (same regex, snake rows migrated: blog_post→blog-post etc.). data-component attributes in templates must equal function; a partial unique index enforces one active component per function. Also separates page NAME 'index' from page TYPE 'landing'.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-naming-standardisation; docs/agent_docs/sql_for_tables/003_pages.sql#051_pages_page_type_kebab; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql
- **relations:** contracts doc 003/042 naming contract; query-resolver list components (rely on page_type values).
- **verify-later:** pg_constraint rows chk_function_kebab_case, chk_page_type_kebab_case; idx_cc_tool_function_unique.

### Component selector metadata and scoring
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library; component quality tracking; site-plan sections resolution.
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage.

### Component quality tracking (0–100 score)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL" (005 ~9848).
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata; improvement loop auditors.
- **verify-later:** compute_component_quality action in registry; populated quality_score values.

### Component versioning (component_versions)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled".
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number) so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations; unclear whether any writer maintains it.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance).
- **verify-later:** row count in component_versions; writers in Go.

### Tier D items-array component schema shape
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration 041 hand-writes tool-list to Tier D; 042 queues guide-list regeneration through the pre-store validator; game-list rewrite mirrors tool-list, "field vocabulary IDENTICAL to tool-list".
- **what:** List components must declare a single `items` array with a sub-schema (title, url, meta_description, nav_label) plus top-level fields (eyebrow_label, section_heading, section_intro, cta_url, cta_label, card_link_label), replacing the legacy numbered-flat anti-pattern (guide_1_url…guide_6_url) that broke sites with fewer items. A pre-store validator enforces the structural contract on LLM-regenerated components; rejections land in agent_error_log.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-041 and #migration-042 and #game-list-rewrite
- **relations:** query-resolver list components; component naming contract; agent_error_log.
- **verify-later:** pre-store validator code; tool-list/guide-list/game-list current schemas.

### Template syntax unification and three-way field alignment
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Long sequence of applied fixes in 005: Handlebars {{#each}}/{{#if}} → Go {{range}}/{{if}}, missing-dot placeholders ({{logo_text}} → {{.logo_text}}), and the "<no value>" root-cause fix aligning LLM prompt output / template fields / input_schema.
- **what:** Templates are Go text/template; a large family of patches converted early Handlebars-style seeds and fixed the recurring three-way mismatch where the LLM prompt, the template field names, and the input_schema disagreed (headline vs title vs section_title; features[].name vs services[].title). Render-context vocabulary standardised (nav_items_html, services_html, footer_nav_html, cta_text, logo_text, company_name).
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#templating-fixes and #fix-no-value; docs/agent_docs/sql_for_tables/012_site_components.sql#replace_template_var
- **relations:** component library; component-based headers; Tier D shape (later formalisation).
- **verify-later:** remaining Handlebars syntax in content_components; render context builder in Go.

### Query-resolver list components (pages_where_type) and canonical section URLs
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** gamesdesign migrations re-type guide pages to page_type='guide' so guide-list (items.source = query.pages_where_type:guide) resolves them, "mirrors the working game-list / page_type=game precedent"; URL migration to /guides/<slug>/index.html.
- **what:** List components resolve their items dynamically from the pages table by page_type via a query resolver — no template change needed when pages are added. Depends on canonical page typing and the canonical nested URL shape /<section>/<slug>/index.html produced by CanonicalisePage, making tools/games/guides structural peers.
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql#migration_retype_guides_to_guide and #migration_guides_url_to_canonical
- **relations:** kebab naming contract; Tier D shape; site-plan page roles.
- **verify-later:** queryresolve Go code; link_registry sync after URL moves.

### Tool library fork-on-deploy model
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** forked_from column, partial unique index on function scoped to canonical tools, and the later constraint amendment "Forks (forked_from IS NOT NULL) are excluded from the uniqueness check" fixing the add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL); deploying to a site copies the row as a fork (forked_from = library id) referenced by page_components. Library changes never cascade to forks; fleet updates go through per-site work items. Uniqueness of `function` applies only to active canonical tools so many site forks can share a function; forks are only ever addressed by component_id.
- **sources:** docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** component library; seeded tool library; improvement-loop fleet updates.
- **verify-later:** deployer fork-copy code; fork counts per library tool.

### Seeded standalone tool library
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** fork-on-deploy model; CSS variable contract.
- **verify-later:** library tool rows and their deployment forks on live sites.

### system.internal canonical library site
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components and #migration-042
- **relations:** site_work_items queue; component regeneration via component-creator.
- **verify-later:** sites row for system.internal; work items with that site_id.

### Style collections
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Two generations of the migration in 001 (initial + 030_style_collections), sites.style_collection_id FK, seed collections professional-dark / minimal-light / bold-gradient with linked header/footer components.
- **what:** A style collection bundles the components and tokens defining a site's visual identity: header/header-home/footer component ids, css_theme_id, color_palette and typography JSONB, category and industry_tags. Sites link to one collection and may override via sites.style_overrides without forking the collection. Original motivation: replace inconsistent LLM-generated headers with tested templates.
- **sources:** docs/agent_docs/sql_for_components/001_style_collections.sql; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_components/002_styles_documentation.md
- **relations:** component-based headers; palette/layout/typography decomposition; design lineage columns.
- **verify-later:** style_collections rows; assignment logic in EnsureSiteRecordAction / classification.

### Component-based headers replacing LLM-generated chrome
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002/003 md docs lay out the plan (store tested header templates, render with site data, inject replacing LLM header); 012 executes population and SQL-side rendering of site_components for header/footer/head.
- **what:** The founding decision that page chrome (header/footer/head) is never LLM-generated per page: tested templates render with a site-derived context (logo from domain, nav from pages/nav tables, colours from collection+overrides) and are injected at assembly. Benefits table: consistency, instant DB-side updates, A/B-able collections.
- **sources:** docs/agent_docs/sql_for_components/002_styles_documentation.md; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_tables/012_site_components.sql
- **relations:** style collections; site/area/page component hierarchy; template syntax unification.
- **verify-later:** RenderHeaderForSite / render_site_components action.

### Palette / layout / typography decomposition (migration 025 phase 2)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Three new tables, empty after this migration... new columns are read only once Phase 4 ships. Phase 3 seeds... Phase 7 drops the legacy columns" (038 header).
- **what:** Splits css_themes.css_template's conflated concerns into three independently versioned tables: palettes (free-shape colours JSONB consumed via {{palette "key" "fallback"}}), layouts (Go CSS template + structure_tokens + default header/footer component ids), typography_sets (fonts + scale via {{typo}}). css_themes becomes a composition row via nullable FKs; renderer migrates in later phases. Also created 10 library layout components (header-with-categories, header-docs, directory-listing, product-grid, etc.).
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components
- **relations:** style collections; design lineage; site_plan_sections resolved palette/layout/typography ids.
- **verify-later:** palettes/layouts/typography_sets row counts; renderer read path (phase 4); legacy column drops (phase 7).

### Design-asset fork lineage (origin / needs_review / source_site)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 038 Part 2(b): lineage columns "required by the Phase 5 fork_theme_from_site action. A prior session reported them as already added but the current schema shows them absent... nothing needs review (fork action hasn't shipped yet)".
- **what:** Uniform provenance on palettes, layouts, typography_sets, css_themes and style_collections: origin ('seed' default), needs_review, forked_from_<entity>_id, source_site_id, source_domain, forked_at. Enables adopting a live site's design into the library as a reviewed fork.
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql#PART2
- **relations:** adoption-pipeline (design adoption); tool fork model (same pattern for tools).
- **verify-later:** fork_theme_from_site action existence; any rows with origin != 'seed'.

### CSS responsibility barrier and CSS variable contract
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "CSS Responsibility Barrier Implementation — Global CSS handles all appearance... Components should NOT re-declare colors" plus the component CSS-variables migration (var(--variable-name, fallback)) applied across all seeded components; hardcoded-colour discovery audit (063b).
- **what:** Global styles.css (from webdesign-agent) owns colours/fonts; component CSS owns only layout/spacing, consuming CSS custom properties with fallbacks (var(--color-primary, #...)). Components must not re-declare colours global CSS styles, with an explicit exception protocol for dark/inverted sections. Audit queries exist to find violators.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#global-vs-local-css and #component-css-variables and #063b_hardcoded_colors_discovery
- **relations:** section-contrast model; style collections; webdesign-agent.
- **verify-later:** styles.css generation; remaining hardcoded colours in component templates.

### Section-contrast model (is_dark_section + --section-* variables)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** Live COMMENT: "is_dark_section ... MUST set --section-text, --section-text-muted, --section-heading, --section-surface, --section-border on container"; 014 section-context variable migration in 005.
- **what:** Components with dark backgrounds are flagged is_dark_section=true and must define the --section-* variable set on their container so text/heading/surface colours invert correctly regardless of the global palette. Migration audited false positives (components using #1a1a2e as text colour, not background) and back-filled the variables per naming contract.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#is_dark_section-comment; docs/agent_docs/sql_for_tables/005_content_components.sql#014-section-context
- **relations:** CSS responsibility barrier; component naming contract.
- **verify-later:** is_dark_section rows vs presence of --section-* in their templates.

### css_snippets / js_snippets with missing JS loader
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 044: js_snippets row news-date-formatter inserted but "THIS ROW IS NOT CURRENTLY LOADED ANYWHERE — the head component template has no snippet-loading mechanism... A small half-day piece of work to mirror loadComponentCSSSnippets" (TODO).
- **what:** Per-component CSS lives in css_snippets (canonical; picked up when webdesign-agent runs) and is loaded via loadComponentCSSSnippets. A parallel js_snippets table exists but no loader; shared JS (e.g. formatNewsDate) is therefore duplicated inline in component IIFEs and page_components.rendered_html as a documented temporary violation of contract 003.
- **sources:** docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** inline-JS extraction contract; news feed rendering; contracts doc 003.
- **verify-later:** js_snippets loader in RenderHead; duplication of formatNewsDate inline.

### Inline JS extraction contract (js_content / separateInlineJS)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "add js_content column for interactive components — Add the column for future use" (005 ~9779); 044 notes the news component's inline <script> "violates contract 003. Properly extracting it via separateInlineJS() would make js_content the source of truth, with /tools/assets/latest-news.js as the served file."
- **what:** Interactive components should store scripts in content_components.js_content and serve them as external files under /tools/assets/, not as inline <script> in html_template. Column added; extraction not consistently done.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#js_content; docs/agent_docs/sql_for_tables/044_css_snippets.sql#why-temporary
- **relations:** css/js snippets; standalone tools (which embed script by design).
- **verify-later:** separateInlineJS usage; js_content population.

### Site / area / page component hierarchy
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** site_components deployed and populated (012); site_areas/area_components created with default 'main' area backfill, but only the site level shows active use; get_page_component fallback function defined.
- **what:** Three-level slot resolution for page chrome: area_components (per site_area override) → site_components (site-wide header/footer/head with rendered_html + content_data for re-render, UNIQUE(site_id, slot_name)) → assembly. site_areas model major site sections with their own nav_style and theme_overrides; get_page_component(page, slot) walks area-then-site.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql; docs/agent_docs/sql_for_tables/014_site_areas.sql; docs/agent_docs/sql_for_tables/015_area_components.sql; docs/agent_docs/sql_for_tables/003_pages.sql#site_area_id
- **relations:** component-based headers; pages.site_area_id; locks (site_components lock columns).
- **verify-later:** area_components usage in production; get_page_component callers.

### Pages / page_components split (structure vs content)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 003 records the design correction: columns first added to pages then explicitly reverted — "Content (rendered_html, content_data) lives in page_components table. Pages table just needs workflow tracking fields"; live dump confirms.
- **what:** pages holds metadata, navigation and workflow (build_status planned→…→deployed/needs_rebuild, sections as planning reference, version) plus per-page rendered_header/rendered_footer/rendered_head for minimal reassembly; page_components holds the actual sections (position, slot_name, component_id, content_data, rendered_html, content_hash, review fields, deploy_commit, research_id). 004b describes the intended three layers: content (content_items) → layout (page_components) → structure (pages).
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** content_items layer; page build workflow; site snapshots capture both.
- **verify-later:** assembly path reading rendered_* columns; build_status writers.

### page_components.build_status CHECK constraint
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "APPLIED 2026-07-11" header; pre-flight survey documented (deployed 597, pending 20; writers enumerated); constraint proved via pg_get_constraintdef.
- **what:** build_status was free text, which let apply_section_edit invent 'approved' and silently remove a live section from every discovery check's audit surface (all filter build_status='deployed'). CHECK now restricts to deployed/pending/approved/removed/needs_rebuild — turning invented values into loud write failures. 'removed' and 'needs_rebuild' retained without writers so future writers need no migration; residual legitimate-'approved'-stuck case covered by the page_component_status_drift check.
- **sources:** docs/agent_docs/sql_for_tables/049_page_components_build_status_check.sql
- **relations:** improvement-loop discovery checks; PLAN_generalise_fixes_to_fleet §4; evidence-based claimed-item timeout (deployed-flag trust).
- **verify-later:** page_components_build_status_check constraint; check_page_component_status_drift.

### Schema-mode strict/flexible subsystem (abandoned)
- **category:** content-governance
- **status-signal:** abandoned
- **status-evidence:** 009 drop migration (2026-07-09): "only partially applied and then abandoned... snapshot columns were never created in production... no Go code reads schema_mode... auto_lock_on_deploy fired exactly once in the system's history"; trigger and function dropped, single strict row normalised.
- **what:** A designed governance regime where approved sections lock to 'strict' schema mode (schema_snapshot + content_snapshot captured; lock_section_to_strict / unlock_section_for_redesign functions; auto-lock on first deploy per sites.strict_mode_trigger). It became an active liability when the apply_section_edit build_status fix would have made every edited section the only locked row on a site for a feature nothing consumed. Orphan functions and columns deliberately left in place; function body preserved as backup.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql; docs/agent_docs/sql_for_tables/009_drop_auto_lock_on_deploy.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#item6
- **relations:** superseded in spirit by Pattern A locks + page_component_history; component_versions.
- **verify-later:** absence of trigger; orphan columns schema_mode/strict_mode_trigger still present.

### Pattern A lock convention (locked_at / locked_by, hard vs soft)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 041 Phase 2A codifies: "a row is locked if locked_at IS NOT NULL. No time comparison... timed expiry is documented design intent (004 v4, 007 v4) but not implemented"; canonical classifier named (check_component_lock.go CheckComponentLock → IsHard).
- **what:** Uniform HITL/agent lock across four tables (page_components, site_components, assets, site_plan_directives — plus site_plan_imagery): locked_at timestamp + locked_by identity. Hard locks ('admin', 'admin-removed', 'checkpoint', 'manual' upload) only humans clear; soft locks ('deploy', auditor names, 'audit-pending') agents may clear when a work item references the row. Discovery skips both; execution skips hard. locked_by vocabulary is convention, not CHECK, to allow new identifiers without migration. A future lock-expiry project would add lock_type/lock_expires_at across all Pattern A tables in one migration.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2A; docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7a; docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#directives
- **relations:** 031_locks.md canonical doc; site-level lock; imagery/directive lock transfer.
- **verify-later:** CheckComponentLock consumers; lock-expiry project status.

### Site-level lock (sites.locked_at)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "Phase 7: Site-level lock — prevents all automated agent activity" (012 tail); scheduled-task pre_query patched to exclude locked sites (020 site-lock section).
- **what:** locked_at/locked_by on sites acts as a master switch: when set, no automated agent activity (discovery, dispatch, improvement) touches the site. Scheduler pre_queries filter locked sites out of candidate selection.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#site-lock
- **relations:** Pattern A locks; scheduler pre_query gating.
- **verify-later:** all dispatch/discovery entry points honour sites.locked_at.

### Asset key multi-image identity (Phase 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Staged migrations with live psql output pasted (11 rows backfilled; backup table assets_backup_20260508_pre_phase2d; ON_ERROR_STOP guard; old (site_id,purpose) unique index dropped).
- **what:** Replaces one-image-per-purpose with per-row asset_key: unique on (site_id, asset_key) WHERE active, enabling multiple images per logical purpose (e.g. adoption-mirror imports as 'adopted:<filename>'). Four-phase rollout: 2B add+backfill (asset_key=purpose), 2C StoreAssetAction writes asset_key and switches ON CONFLICT, 2D drops old purpose uniqueness after straggler sanity check.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2B and #Phase2D
- **relations:** assets provenance; site_plan_imagery key → namespaced asset_key.
- **verify-later:** idx_assets_site_asset_key_unique; StoreAssetAction ON CONFLICT target.

### Assets table with full provenance
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 004 PART 6 creates assets with origin tracking; later phases (locks, asset_key) applied to a live table with 11 rows.
- **what:** All binary assets (image/video/document/logo/favicon) with storage location (provider/path/url), file metadata, and provenance: origin_type (generated/uploaded/scraped/stock/affiliate/derived), origin_url/prompt/model, origin_asset_id for derivations, alterations history JSONB, attribution/license. Purpose field ('hero', 'og_image'...) drives placement.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART6; docs/agent_docs/sql_for_tables/041_assets.sql
- **relations:** asset_key identity; image-build-handler work items; storage-architecture (providers).
- **verify-later:** StoreAssetAction; storage_provider values in use.

### site_plan_imagery structured imagery plan
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "PURE DDL with NO BEHAVIOUR CHANGE. The table is empty until step 2 (write_site_plan extension) and step 3 (planner prompt extension)" with the 5-step Phase 2G sequencing listed (043 header).
- **what:** Sibling of site_plan_directives holding structured imagery requirements at site/page/section scope: key (asset_key stem, namespaced by the discovery check), kind enum via chk_kind (logo, hero, illustration, icon, infographic — product deliberately excluded, it comes from the affiliate_products resolver), required prompt, style_hints/constraints JSONB that cascade ADDITIVELY with directives' imagery_direction, ordering, source enum, and HITL locking with lock-transfer across plan rebuilds. chk_scope_ref_consistency enforces NULL / page_name / 'page:ordering' shapes; unique on (plan, scope, COALESCE(scope_ref,''), key).
- **sources:** docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql
- **relations:** site_plans domain; PLAN_imagery_phase_2g.md; check_unfulfilled_imagery_plan (step 4); image-build-handler (step 5).
- **verify-later:** table population; steps 2–5 delivery status.

### site_plans declarative plan domain
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Migration 031 (both drafts) with detailed rationale referencing doc 030; later tables (site_plan_imagery, work-item flows) depend on and reference it.
- **what:** The plan is a separate versioned artefact from site_specs: site_plans (version anchor, one is_current per site), site_plan_pages (row per planned page: canonical name/role/slug/url, parent_section for section-index detection, nav flags), site_plan_sections (structural per-section rows carrying resolved component_version/palette/layout/typography ids for HTML data-* provenance), site_plan_directives. Row-per-thing chosen over JSONB blobs for 1000+ page scale and surgical HITL edits; versioning mirrors site_specs (is_current + superseded_at).
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql
- **relations:** site_specs (strategic vs operational boundary); reconciler; naming note that plan_sections/save_page_sections actions "share a noun and nothing else".
- **verify-later:** write_site_plan action; plan row counts per site.

### Directive cascade and HITL lock transfer
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 040 second draft: scope_ref encoding, cardinality lookup in brief renderer, "write_site_plan... transfers the lock onto the equivalent new directive row" matched by (scope, scope_ref, category, subject, ordering).
- **what:** Design/content/voice/structural guidance stored row-per-directive at site/page/section scope; a Go brief renderer walks the cascade (site → page → section) and emits prompt-ready text — consumers never read directives directly. Cardinality (override vs accumulate) is renderer knowledge, not schema. Human-locked directives survive plan rebuilds via stable-composite-key lock transfer performed only by write_site_plan.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_directives
- **relations:** Pattern A locks; site_plan_imagery (same pattern); doc 030 "Directive cascade and brief assembly".
- **verify-later:** brief renderer helper; lock-transfer code in write_site_plan.

### Plan drift detection and reconciler scheduling
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** pages.built_from_plan_version + sites.last_reconciled_at columns with reconciler semantics documented; later migrations reset built_from_plan_version=NULL to force rebuilds.
- **what:** Each built page records the plan version that produced it; the reconciler diffs site_plan_pages against pages, flags pages whose plan version lags current (NULL = never built under a plan), and emits needs_page/rebuild work items. sites.last_reconciled_at lets the scheduled tick skip recently reconciled sites; deliberately no FK so hard-deleted plans read as drift.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#4 and #5; docs/agent_docs/sql_for_tables/003_pages.sql#rebuild-flips
- **relations:** site_plans domain; site_work_items; scheduler.
- **verify-later:** reconcile_site_plan action; scheduled reconciler task.

### site_plan_partials with lazy page briefs (early plan shape)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** First draft of migration 031 defines site_plan_partials ('design_direction', 'content_strategy' eager; 'page_brief:<name>' lazy via build_page_brief); the second draft in the same file replaces it with site_plan_sections + site_plan_directives.
- **what:** The initial plan-domain design stored design direction, content strategy and per-page briefs as versioned JSONB partials, with lazy page briefs written on demand by page-build-handler. Superseded by the row-per-section/row-per-directive shape for scale and surgical edits.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_partials
- **relations:** superseded by site_plan_sections + site_plan_directives.
- **verify-later:** whether site_plan_partials exists in production or only the directive shape shipped.

### site_specs aspect-versioned specification store
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Table created in Phase-0 migration (019); live pg_dump backup (bk_site_specs.sql) shows the production shape including pinned; extensive backfills for real sites.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current per site+aspect). write_site_spec deep-merges partials so each row is self-contained. `pinned` (Phase 4) prevents agents overriding human-set specs.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a-team-data
- **relations:** site_plans (operational counterpart); site snapshots capture current specs; identity enrichment (departments/team).
- **verify-later:** write_site_spec action; pinned enforcement in writers.

### Growth budget spec (growth_config)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Seed migration inserts default growth_config for existing active sites: '{"initial_target": 12, "weekly_content_pages_max": 3, "weekly_blog_posts_max": 2, "absolute_max": 60}'.
- **what:** Per-site growth limits stored as a site_specs aspect: initial page target, weekly content/blog caps, absolute page maximum. Admin-overridable via the dashboard Direction tab; consumed by growth/budget calculations in planning.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#growth_config
- **relations:** site_specs store; page_growth_budget.go (referenced from 042 thunder step-zero).
- **verify-later:** budget enforcement in planner/discovery.

### content_items reusable content layer
- **category:** content-governance
- **status-signal:** unknown
- **status-evidence:** Full DDL + helper get_component_content + v_content_usage view exist and page_components.content_item_id survives into the live dump, but no later file shows content_items being written.
- **what:** Separates "what to say" from "how to show it": typed reusable content rows (headline, tagline, service_description, testimonial, bio, cta, faq...) with semantic content_key, plain_text search, library sharing (site_id NULL + is_library + industry_vertical + library_tags), assets-style origin tracking, and status workflow. page_components reference a content_item with content_data acting as shallow-merge override (get_component_content). Would let one tagline appear in hero, footer and meta without duplication and let library content seed new sites.
- **sources:** docs/agent_docs/sql_for_tables/004_content_items.sql; docs/agent_docs/sql_for_tables/004b_content_items.md; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql#content_item_id
- **relations:** pages/page_components split; assets origin pattern.
- **verify-later:** content_items row count; any writer using content_item_id.

### page_component_history full-snapshot content history
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** Created in Phase-0 Block A (019) explicitly as the replacement for the dropped content_snapshot/schema_snapshot columns.
- **what:** Before any content_data write to a page_component, the current value is copied here as a complete snapshot (not a diff) with source ('content-writer', 'section-editor', 'rollback'...) and triggering work item id — the rollback/audit substrate for section edits.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#3
- **relations:** replaces schema-mode snapshots; section-editor; site snapshots (page-level vs site-level revert).
- **verify-later:** writers copying into history before UPDATE.

### Section governance columns: content_brief and suppressed_sections
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "content_brief: records the instructions that generated each component's content. Enables admins to see, edit, and regenerate" (008 tail); "suppressed_sections... prevents discovery checks from recreating sections that were intentionally removed. The DELETE component endpoint writes to this column" (003).
- **what:** Two small governance mechanisms: page_components.content_brief JSONB preserves the generation instructions per section for admin-editable regeneration; pages.suppressed_sections lists intentionally removed section functions so discovery does not resurrect them (component-removal flow Phases 2/5).
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#content-brief; docs/agent_docs/sql_for_tables/003_pages.sql#suppressed_sections
- **relations:** improvement-loop discovery checks; inline editing / regeneration.
- **verify-later:** DELETE component endpoint; discovery filters on suppressed_sections.

### Placeholder-content suppression sweep
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval.
- **verify-later:** validation agent producing these; recurrence of placeholder text.

### site_work_items unified work queue and lifecycle
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Full DDL (023_site_work_items) with dedup index, plus dozens of live operational patches (resets, handler re-routing, attempt bumps) against real sites.
- **what:** Every piece of platform work is a row: source (planner/discovery/content_feed/manual/improvement/side_effect/human/validation), pipeline (originally `domain`, later renamed), item_type, severity, spec JSONB, target refs (page/component/entity/url), triage enrichment (impact, resolution_path, suggested_action, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete/pending_verify/verified/failed/rejected/wont_fix plus 'blocked' (handler missing), dependencies (depends_on UUID[], parent/related/batch), attempts, and deterministic item_key with a partial unique index for dedup among non-terminal items. A same-structure archive table receives terminal items.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#approval_mode; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle
- **relations:** dispatch loop; claimed-item timeout; content_feed_items; archiver; approval_mode; processing_tier.
- **verify-later:** current status distribution; pipeline column rename (`domain` dropped in 018).

### Work item archival
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** archive_completed_work_items(age, batch) function + archiver agent definition + daily scheduled task, with schema-sync ALTERs and FK handling (parent self-ref cleared, content_feed_items references deleted).
- **what:** Terminal work items (complete/failed/wont_fix) older than a configurable age move to site_work_items_archive in batches, keeping the live queue small. Function handles column drift between live and archive tables explicitly.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-archiver
- **relations:** work queue; scheduler.
- **verify-later:** archiver task enabled; archive row counts.

### build_queue site seeding
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Phase-0 Block A table with direction semantics enumerated.
- **what:** Domain-level intake queue for new sites: a row per domain with direction JSONB (null | {objective} | {adopt_from} | {fork_from} | {brief_complete...}), status and priority. seed_build_queue reads it, creates site records and initial work items according to direction — the entry point into the work-item pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#2
- **relations:** work queue; adoption pipeline (adopt_from); onboarding-config (brief_complete).
- **verify-later:** seed_build_queue action; build-pipeline-trigger seeding behaviour.

### Work item approval_mode (auto / hitl / eval)
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Column added in Phase-0 Block A: "Controls whether items auto-dispatch or require human/eval approval. Values: 'auto' (default), 'hitl', 'eval'".
- **what:** Per-item gating between triage and dispatch: auto items flow straight through; hitl items wait for human approval; eval items wait for an evaluation agent. The schema-level hook for configurable human review gates in the build pipeline.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#4
- **relations:** work queue; input_requests; human change requests.
- **verify-later:** dispatch respects approval_mode; any 'eval' users.

### Work item processing_tier (standard / batch_gpu)
- **category:** batch-processing
- **status-signal:** unknown
- **status-evidence:** Bare ALTER in 018: "'standard' — process immediately... 'batch_gpu' — hold until GPU batch starts, then process via GPU Ollama"; no consumer shown.
- **what:** A routing column intended to hold selected work items until a GPU batch window opens so they run on GPU Ollama instead of Claude/CPU — a cost/throughput lever tying the work queue to GPU batch scheduling.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#processing_tier
- **relations:** batch-processing; model-infrastructure (GPU Ollama); thunder GPU provisioning.
- **verify-later:** any writer/reader of processing_tier.

### Human change-request work items
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Applied INSERT batches for gaswholesalers.com and finetuning.uk: needs_design→webdesign-agent, content_edit→section-editor (field_updates spec), needs_logo/needs_hero_image→image-build-handler (image_prompts spec), needs_rerender→rerender-pages at priority 99.
- **what:** Human requests enter the same queue as agent-detected work: source='human', item_key 'human_<what>_<site>', pre-triaged, priority-ordered so content and imagery land before the final rerender. The spec JSONB is the full handler contract (edit_type/page_name/slot_name/field_updates for edits; purpose + image_prompts for imagery).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** work queue; section-editor; image-build-handler; dispatch loop (071).
- **verify-later:** admin UI path creating these; section-editor field_updates handling.

### input_requests HITL persistence
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Full schema with pending view, expiry function and debugging queries; FK to orchestration_states with CASCADE.
- **what:** Human input requests persisted for querying and UI display: request_type (review/confirmation/questionnaire), title/message/data/ui_config, reply_to_topic for Kafka response routing, timeout/expires_at, status lifecycle pending→completed/expired/cancelled, and the response payload with responder identity. pending_input_requests view feeds the UI with seconds_remaining.
- **sources:** docs/agent_docs/sql_for_tables/006_input_requests.sql
- **relations:** awaited_requests (transport-level counterpart); approval_mode; admin dashboard.
- **verify-later:** UI consumption; expire_input_requests scheduling.

### Manual HITL continuation runbook
- **category:** hitl
- **status-signal:** deployed
- **status-evidence:** Working queries against awaited_requests for stuck escalate_to_human steps, plus the documented hitl_respond.sh Kafka invocation with real ids/topics.
- **what:** Operational procedure for un-sticking HITL flows: locate the awaited request (step_name LIKE %human%/%hitl%/%approval%), optionally reset an expired one, then publish the human response directly to the reply_to_topic via hitl_respond.sh — including Kafka topic existence checks and consuming system.notifications.ui.
- **sources:** docs/agent_docs/sql_for_hitl/001_hitl_requests.sql
- **relations:** awaited_requests registry; input_requests; debugging.
- **verify-later:** hitl_respond.sh script location.

### awaited_requests global request/response registry
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Two schema generations (tables_sql 001 → sql_for_tables 001 matching the AwaitedRequest Go struct), plus later additions of 'processing' status, processing_started_at/processing_pod claim tracking, and cleanup function.
- **what:** DB-backed registry matching Kafka responses to waiting orchestrations, solving the race where a child creates a request while the parent receives the response. Keyed by request_id with orchestration/correlation context, target agent, responses/requests topics, retry_version, reply_to_request_id chaining, timeout_at, and status lifecycle waiting→processing→processed/expired/cancelled/error. Expired rows are marked then purged after 7 days by cleanup_expired_awaited_requests.
- **sources:** docs/agent_docs/sql_for_tables/001_awaited_requests.sql; docs/agent_docs/tables_sql/001_awaited_requests.sql
- **relations:** processed_messages idempotency; HITL runbook; orchestration_states.
- **verify-later:** state.go AwaitedRequest struct; cleanup scheduling.

### processed_messages idempotency dedup
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Live \d output plus applied ALTERs adding retry_version and re-keying the PK to (correlation_id, request_id, agent_id, retry_version).
- **what:** Exactly-once message processing guard: each consumed message records correlation/request/agent identity; the composite PK including retry_version allows deliberate retries while blocking duplicate deliveries within a retry generation.
- **sources:** docs/agent_docs/sql_for_tables/007_processed_messages.sql
- **relations:** awaited_requests retry_version; Kafka consumer semantics.
- **verify-later:** consumer insert-or-skip logic.

### Orchestration ↔ site linkage (orchestration_states.site_id)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Migration with three-path backfill from collected_data (input_data.site_id, site_record.site_id, top-level) and verification counts against gamedesign.uk.
- **what:** Direct nullable site_id column on orchestration_states (set at creation) replaces JSONB spelunking for "orchestrations for this site", with a partial index for active orchestrations per site. Nullable because not all orchestrations are site-scoped (health checks).
- **sources:** docs/agent_docs/sql_for_tables/036_orchestration_states.sql
- **relations:** debugging queries; improvement-sweep pre_query.
- **verify-later:** creation-time population in Go.

### orchestration_state_audit investigation trigger
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Create-trigger + analysis queries (time_since_prev via LAG, pg_backend_pid, application_name) and explicit "Remove trigger when done investigating" teardown.
- **what:** A temporary, attachable audit table + AFTER UPDATE trigger capturing every version/status/current_step transition on orchestration_states — used to diagnose state races and stuck orchestrations, then removed. Distinct from permanent logs; also cleaned up by database-cleanup (keeps last 100k rows).
- **sources:** docs/agent_docs/sql_for_tables/010_orchestration_state_audit.sql; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup
- **relations:** debugging guide; database cleanup.
- **verify-later:** whether trigger currently attached.

### agent_error_log persistent error record
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "This replaces digging through kubectl logs to find error details"; captured from routeToErrorStep and notifyParentOfFailure; referenced later as the sink for Tier-D validator rejections.
- **what:** Queryable record of every agent error: what failed (site/domain/work_item), where (orchestration, agent_type/id, pod, step, action), the error (message, error_code, severity), a JSONB context snapshot, and resolution tracking (resolved/resolved_by). Indexed for dashboard recency, per-site, unresolved, and per-agent-type frequency views.
- **sources:** docs/agent_docs/sql_for_tables/022_agent_error_log.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-042
- **relations:** database cleanup retention; fix loops consuming structured errors.
- **verify-later:** writers in chassis error paths.

### http_request_log outbound HTTP observability
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Migration "Follows the same pattern as llm_call_log"; stats view with calls_last_5min for rate-limit monitoring; cleanup 90d success / 180d errors.
- **what:** Centralised log of every outbound HTTP call from Go actions: caller identity (agent/step/orchestration/action_name), method/url/domain/path, response status/bytes/latency/success, metadata JSONB. Purposes: operational visibility and per-domain rate-limit tracking (e.g. Companies House).
- **sources:** docs/agent_docs/sql_for_tables/026_http_request_log.sql
- **relations:** llm_call_log (pattern sibling); companies-house rate limiting.
- **verify-later:** HTTP client wrapper writing rows.

### Loop-action dispatch (migration 071)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Applied UPDATE of build-dispatch-loop default_config: "Step-chaining... processes only one work item per trigger. The loop action is proven in maintenance-triage and pageflow-builder."
- **what:** The dispatch loop loads all dispatchable items upfront (dependency-filtered, priority-ordered, max 50) and iterates with the `loop` action running a sub_workflow per item: claim → check_claim → spawn_handler (dynamic agent type from current_item.handler_agent) → call_handler → mark_complete/mark_failed, with continue_on_error. Introduces item_variable scoping (current_item.*) and optional `?`-suffixed input_mapping fields silently skipped for handlers that don't need them (section-editor compatibility).
- **sources:** docs/agent_docs/sql_for_hitl/002_adding_some_requests.sql#migration-071
- **relations:** work queue; spawn-orchestrator pattern; claimed-item timeout.
- **verify-later:** build-dispatch-loop live config; loop action implementation.

### Spawn-orchestrator thin-wrapper pattern
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Repeated across med pricing (scrape, discover, map, export orchestrators): spawn_agent (role) → call_agent (target_role, input_mapping passthrough, timeout) → complete_workflow; scheduled tasks target the orchestrator, not the worker.
- **what:** The standard shape for burst workloads: a permanently-resident category pod receives the trigger, a thin orchestrator workflow spawns a temporary worker pod of the right agent_type, forwards input_data, awaits the result, and completes — worker terminates (idle_timeout 0). Non-secret worker config rides env_vars on the agent definition; secrets come via spawn_actions secretKeyRef.
- **sources:** docs/agent_docs/sql_for_tables/034_vet_med_price_scrape_orchestrator.sql; docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql; docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** scheduler; agent definitions; vet med pipelines.
- **verify-later:** spawn_agent/call_agent actions; ON CONFLICT (type, version) upsert convention.

### Kafka scheduler and scheduled_tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Table DDL (066_kafka_scheduler) plus a long operational history of seeded tasks: build-pipeline-trigger, improvement-sweep, claimed-item-timeout, feasibility-recheck, content-feed-refresh, database-cleanup, vet-*, med-*, ch-enrichment, health checks, archiver.
- **what:** Interval-based scheduling in Postgres: each row names a target agent/topic, input_data, interval_seconds, timeout, and concurrency_group/max_concurrent (group-wide in-flight cap). The scheduler publishes Kafka trigger messages; last_triggered_at/last_completed_at implement a no-refire guard (with known operational pitfalls when nothing sets last_completed_at for fire-and-forget tasks — mitigated by shorter timeout windows).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql
- **relations:** pre_query SQL-worker pattern; every pipeline's periodic trigger.
- **verify-later:** kafka-scheduler service; fire_message column semantics.

### pre_query SQL-worker pattern and self-healing tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Iterated in place: pre_query CTE UPDATEs doing the work, then `WHERE 1=0` / `HAVING COUNT(*) > 0` variants to control whether a Kafka message fires; vet-cleanup broadened to fail stuck AWAITING_RESPONSES orchestrations and reset orphaned collection tasks.
- **what:** scheduled_tasks.pre_query is a full worker channel, not just a gate: SQL that returns rows merges into input_data and fires the message; returning zero rows skips the tick. Maintenance tasks exploit this to run entire cleanup UPDATEs inside the pre_query (claimed-item reset, blocked-item promotion, orchestration failing, database cleanup) with row-suppression idioms deciding whether anything downstream is triggered (fire_message=false for pure-SQL tasks).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle and #vet-cleanup and #database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql
- **relations:** scheduler; claimed-item timeout; database cleanup.
- **verify-later:** scheduler's pre_query evaluation code; fire_message flag.

### Claimed-item timeout with evidence-based auto-completion
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** v1 then v2 ("SUPERSEDES... Apply THIS one") migrations with two confirmed production false positives dated 2026-05-12 and 2026-06-04 (gamesdesign homepage auto-completed with ZERO page_components — root cause of the missing root index.html).
- **what:** The stuck-claim recovery task distinguishes "work actually finished but the response was lost" from "handler died": items claimed >15 min are auto-completed only on artifact-specific evidence — needs_content_page requires page_components rows for that page updated after the claim (ground truth, not the untrustworthy build_status='deployed' flag), page_rerender requires page.deployed_at after claim, needs_design keeps a caveated site-level check; needs_rerender is deliberately excluded (site-level, retry is cheap). Everything else resets at >40 min with attempt accounting and fail-on-exhaustion.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#migration_claimed_item_timeout_evidence_v2; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#claimed-item-timeout
- **relations:** work queue lifecycle; build_status CHECK (flag trust); UpdatePageStatusAction 0-component guard ("Option B").
- **verify-later:** live pre_query text of claimed-item-timeout; debugging guide section 9.

### Improvement-sweep and build-pipeline-trigger scheduling
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Seeded tasks with evolving pre_queries: queue-size gate (skip when >20 open build items), round-robin site selection by least-recently-checked, skip sites with claimed items or locks.
- **what:** The improvement loop's cadence lives in scheduled_tasks: build-pipeline-trigger (2 min) finds sites with triaged/approved items and fires the dispatch loop; improvement-sweep (10 min) picks the next site for discovery checks, gated so discovery never floods an already-backed-up queue and locked sites are skipped. Both share the 'dispatch' concurrency group.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#seed-data and #improvement-sweep-fixes
- **relations:** scheduler; work queue; site-level lock.
- **verify-later:** current pre_query for improvement-sweep; discovery agent set.

### Database cleanup and log retention policy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** database-cleanup task (agent_error_log 14/30 days, audit last 100k, orchestrations 7 days/24h stuck, orchestration_requests FK made CASCADE) plus per-table cleanup functions (llm 90/180, http 90/180, awaited 7 days) and the always-return-a-row HAVING fix.
- **what:** A uniform retention discipline: every high-churn operational table has an explicit cleanup function or scheduled CTE with distinct retention for successes vs errors, and the cleanup task itself is written to always mark itself executed. Includes the FK CASCADE fix required so orchestration deletion cascades to requests.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql; docs/agent_docs/sql_for_tables/026_http_request_log.sql#cleanup
- **relations:** scheduler pre_query pattern; agent_error_log; llm/http logs.
- **verify-later:** database-cleanup enabled and returning rows.

### Migration discipline: pre-change snapshots, renumbering, footguns
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Recurring practice across files: CREATE TABLE ... AS SELECT snapshots inside the txn (pages_bak_retype_guides, assets_backup_20260508...), NNN placeholders with "confirm the next free migration number" notes (048, 046, 047), deliberate plain CREATE TABLE to error on shape mismatch ("the migration-110 trap, §6.1"), and "code shipped but migration unapplied has bitten this project repeatedly" (042 ssh_port).
- **what:** The project's migration conventions as embodied in the files: snapshot rows before destructive UPDATEs with pasted rollback SQL; verify blocks (DO $$ ... RAISE EXCEPTION) inside transactions; idempotence via IF NOT EXISTS except where silent no-op would hide a shape conflict; migration numbers confirmed against the live runner before applying; migrations applied separately from code deploys.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#BEFORE-APPLYING; docs/agent_docs/sql_for_tables/003_pages.sql#snapshots; docs/agent_docs/sql_for_tables/042_thunder.sql#ssh_port
- **relations:** debugging guide §6.1 / item 17; agent snapshot/revert.
- **verify-later:** migration runner and numbering source of truth.

### Early schema inventory and since-dropped tables
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** 006 is a psql \dt+ snapshot listing tables absent from later docs: flow_pages, site_flows, navigation_structures, pending_requests, improvement_proposals, approval_requests, agent_groups/agent_group_definitions/agent_group_members, agent_metrics, theme_tags, system_events, event_statistics matview; 016 explicitly replaces "the navigation_structures cache table".
- **what:** A point-in-time inventory of clients_db that preserves abandoned concepts: site flows/flow pages (a flow-based site model), a navigation cache table, standalone improvement_proposals and approval_requests tables (roles later absorbed by site_work_items and input_requests), and an agent-groups mechanism. Valuable as the "what silently vanished" record.
- **sources:** docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql; docs/agent_docs/sql_for_tables/016_nav_tables.sql
- **relations:** superseded by site_work_items, input_requests, site_nav_* tables.
- **verify-later:** whether these tables still exist in production (dead weight) or were dropped.

### Navigation tables (site_nav_groups / site_nav_items)
- **category:** navigation
- **status-signal:** deployed
- **status-evidence:** DDL plus a real-site query result (primary/legal groups for a live site) and the applied global template fix converting anchor links to page URLs.
- **what:** First-class navigation model replacing scattered pages-table queries and the navigation_structures cache: groups per site (group_key primary/legal/utility/content, group_type, hierarchy via parent_group_id) containing typed items (page_link/external_link/anchor/section_header, FK to pages with SET NULL, position, status, metadata). Sites without rows fall back to Go logic querying pages directly. Render context supplies both .slug and .url per item; templates must link {{.url}} (061 fix purged href="#{{.slug}}" from all header/footer/nav templates).
- **sources:** docs/agent_docs/sql_for_tables/016_nav_tables.sql; docs/agent_docs/sql_for_tables/017_site_nav_groups.sql
- **relations:** site snapshots capture nav; component-based headers consume nav_items.
- **verify-later:** nav writer agent; fallback path in Go.

### llm_call_log training-data flywheel
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Migration 081 Part 3 + schema fixes (agent_id added, nullability relaxed to match Go's nullIfEmpty); export queries reference populated columns incl. work_item_id and vertical.
- **what:** Every LLM call logged with caller identity (agent_type/step/orchestration), model + model_resolved + provider, full prompt_template/prompt_rendered/response_text, token/latency usage and outcome — explicitly designed for training export. Export recipes produce JSONL per task (analyze_tool, recreate_tool, site classification, content writing) with quality filters joining site_work_items outcomes (only export calls whose work item completed), and per-vertical readiness counts.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql
- **relations:** training_exports (successor storage); site_chat_turns (deliberately separate); model upgrades.
- **verify-later:** logging middleware in aiservice; work_item_id/vertical columns present.

### training_exports Postgres-backed datasets (flywheel A v3)
- **category:** finetuning-flywheel
- **status-signal:** deployed
- **status-evidence:** Schema with rationale "JSONL files landed on ephemeral chassis pods and vanished on restart"; dedup unique index on (export_id, metadata->>'source_log_id').
- **what:** Named, versioned training datasets in Postgres instead of ephemeral JSONL: runs (one per export — filter criteria matching llm_call_log columns, counts, skip reasons, format 'chatml', size, provenance) and rows (ChatML messages + metadata JSONB, ordered by row_index, CASCADE delete). Training-time extraction via \copy in export order. Schema named training_exports specifically to avoid confusion with the model-training pipeline (flywheel C).
- **sources:** docs/agent_docs/sql_for_tables/039_training_exports.sql
- **relations:** llm_call_log source; thunder training runs (flywheel C).
- **verify-later:** exporter action writing runs/rows.

### Agent model-assignment upgrade sweeps
- **category:** model-infrastructure
- **status-signal:** deployed
- **status-evidence:** Migration 081 Parts 1–2: chief-strategist → opus-4-6; site-planner/domain-research-classifier/domain-strategist/site-classifier → sonnet-4-6; stale claude-3-5-sonnet-20241022 and claude-3-opus refs globally replaced.
- **what:** Model choices live inside agent_definitions.default_config and are upgraded by targeted text-replace UPDATEs, with an explicit tiering philosophy: high-leverage structural deciders get the best models. Also documents the historical model vocabulary embedded in configs.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#Part1-2
- **relations:** agent-definition registry; llm_call_log model_resolved.
- **verify-later:** current model distribution across agent_definitions.

### Agent definition snapshot/revert via backup table
- **category:** NEW:agent-definition-registry
- **status-signal:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup.
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced.

### knowledge_base RAG store
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** Migration 082 (idempotent) with pgvector + pg_trgm; 048 later confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** Industry/marketing content chunks for retrieval: collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via ollama-adapter) with IVFFlat cosine index, trigram GIN fallback for keyword retrieval when embeddings are unavailable, source tracking, quality_score and usage_count lifecycle, stats view.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** code_symbols (sibling shape); ollama-adapter embedder; content grounding.
- **verify-later:** collections in use; retrieval actions.

### code_symbols per-repo code index (context tool)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "CONFIRMED (2026-06-09, clients_db): \dt found no existing *code*/*symbol* table, and \dx shows vector 0.8.0... Both gates pass — HNSW stands."
- **what:** The context tool's code index: one row per symbol keyed (repo, path, symbol) with kind CHECK (func/method/struct/interface/alias/type/var/const), signature/doc/line range (bodies read from the repo at commit_sha, not stored), content text that is both embedded (HNSW cosine, chosen over IVFFlat for incremental churn) and trigram-matched, content_hash to skip re-embedding unchanged symbols. Deliberate departures flagged: no version/soft-delete — a rebuildable cache versioned by commit_sha, pruned by hard delete. Ships the full usage contract in comments: indexing upsert, prune, semantic/lexical retrieval, and hybrid RRF fusion in SQL (constant 60) replacing in-Go fuse.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql
- **relations:** knowledge_base shape reuse; diagnosis-loop code retrieval; contextkit.
- **verify-later:** indexing workflow; code_symbols row counts per repo.

### Thunder adapter schema and provisioning gates
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Full schema with recorded user decisions ($100/day cap, 2 concurrent, 18h uptime, $1.80/hr A100, $25 estimated run); production fix dated 2026-05-22 for identifier recycling; ssh_port verification dated 2026-05-24.
- **what:** GPU VM lifecycle for training: thunder_instances (one row per VM ever provisioned — inserted BEFORE the API call so the reaper always has a record; status machine provisioning→running→decommissioning→decommissioned with reaped/lost/failed terminals; cost snapshot; reaper bookkeeping; FK to model_lifecycle.training_runs), thunder_config singleton (CHECK-enforced single row; caps and pause switch), and computed views thunder_spend_24h (rolling cost incl. running estimates, no drifting counter) and thunder_provision_check (can_provision + denial_reason evaluated at every provision request). Identifier recycling fixed by replacing global uniqueness with a partial unique index over live states only; ssh_port captured at provision so ssh_exec dials directly.
- **sources:** docs/agent_docs/sql_for_tables/042_thunder.sql
- **relations:** thunder unreachable streak; training_runs (flywheel C); 013_thunder_adapter_design.md.
- **verify-later:** adapter reading thunder_provision_check; reaper behaviour.

### Thunder consecutive-unreachable probe streak
- **category:** adapters
- **status-signal:** deployed
- **status-evidence:** Migration 106 with in-transaction verification; rationale documented (single down-probe could be a transient SSH blip; each scheduler tick is a fresh sub-agent that can't hold state in memory).
- **what:** thunder-training-monitor durability: consecutive_unreachable_probes counter (+ last_probe_at) on thunder_instances, bumped/reset by the record_probe_streak action; only after the streak crosses a threshold is the instance treated as 'lost' (fail run + decommission). State lives on the row because monitor ticks are stateless.
- **sources:** docs/agent_docs/sql_for_tables/047_thunder_unreachable_counter.sql
- **relations:** thunder adapter; scheduler tick statelessness.
- **verify-later:** record_probe_streak action; threshold value in monitor config.

### Vet med pricing schema (products / retailers / listings / snapshots)
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Schema migration applied (028, duplicated in 030), retailer seeds with live corrections ("URL structure changed", "Domain is animed.co.uk not animeddirect.co.uk — updated from plan"), test seeds, manual listing matches.
- **what:** Four tables + matview in business_intel: med_products (canonical catalog: generic/brand/manufacturer/species[]/category/form/strength, prescription flag), med_retailers (4 UK pharmacies with group ownership — IVC Evidensia, CVS, Covetrus, Independent — category_urls for discovery, delivery costs, scrape_config hints), med_retailer_listings (retailer URL per product with match_confidence/match_method manual|llm|exact_name, NULL product until matched, denormalised last_price), med_price_snapshots (per size_variant price history incl. typical_vet_price), and med_price_current materialized view (latest price per listing/variant within 14 days) for export.
- **sources:** docs/agent_docs/sql_for_tables/028_vet_med_prices.sql; docs/agent_docs/sql_for_tables/029_vet_med_retailers.sql; docs/agent_docs/sql_for_tables/029b_vet_med_test_seed.sql; docs/agent_docs/sql_for_tables/033-business_intel.med_retailer_listings.sql
- **relations:** scrape evidence; spawn orchestrators; JSON export.
- **verify-later:** matview refresh cadence; listing match coverage.

### Med scrape evidence store
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** Table comment: "Raw scraped page content as evidence of prices. One row per page fetch. Retention: keep at least 90 days."
- **what:** Every price is traceable to the page it came from: one row per fetch with the Firecrawl markdown content, SHA256 content_hash for unchanged-page detection, variants_found vs prices_stored accounting, and response metadata — the audit trail for price provenance.
- **sources:** docs/agent_docs/sql_for_tables/032_business_intel_med_scrape_evidence
- **relations:** med pricing schema; vet-med-pricing evidence requirement (doc 008).
- **verify-later:** evidence rows per scrape run; retention enforcement.

### Med URL discovery via Firecrawl /map
- **category:** vet-med-pricing
- **status-signal:** partial
- **status-evidence:** med-url-mapper seeded with status 'experimental' ("Particularly useful for VioVet where category-page scraping misses products"); registry.go entry supplied as a comment, i.e. Go side pending at write time.
- **what:** A second, broader product-URL discovery path using Firecrawl's /map endpoint site-wide, alongside category-page crawling; wrapped in the standard spawn orchestrator (med-url-map-orchestrator).
- **sources:** docs/agent_docs/sql_for_tables/035_vet_med_url_mapper_and_orchestrator.sql
- **relations:** spawn-orchestrator pattern; med pricing discovery.
- **verify-later:** med_map_urls in registry.go; experimental → active status.

### Configurable med price JSON export to site repos
- **category:** vet-med-pricing
- **status-signal:** deployed
- **status-evidence:** med-json-exporter agent seeded 'active' with full config (domain, repo, data_path, outputs index/full/by_letter/metadata); scheduled task med-export-json seeded every 48h but enabled=false initially.
- **what:** One generic export action serves many consumer sites via config: query med_price_current, apply filters (species/category/retailers), build JSON artefacts, and commit them into the target site's git repo (e.g. vetcomparison.co.uk /data). The price data pipeline's publishing edge — sites consume static JSON, not the DB.
- **sources:** docs/agent_docs/sql_for_tables/037_vet_med_export_orchestrator_prices_json.sql
- **relations:** deployment-github (commit path); client-side JSON rendering pattern.
- **verify-later:** exports landing in site repos; task enablement.

### Companies House enrichment with succession-risk signals
- **category:** companies-house-enrichment
- **status-signal:** deployed
- **status-evidence:** Schema + scheduled task (ch-enrichment every 20 min, seeded disabled "until Go actions are built") and a later applied accounts-fetch migration (accounts_fetched tracking, financial columns), indicating progression to live collection.
- **what:** Post-verification enrichment of business_intel.businesses: company identity/status/SIC, financials from filed accounts (accounts_type micro/small/medium/full, assets/net worth/turnover/PL, employees), officers and PSC JSONB, and derived owner-age/succession signals (owner_dob from CH month/year, estimated age, tenure, is_sole_director, is_corporate_owned → succession_risk high/medium/low/acquired). Deliberately polite rate limiting (~7% of CH's 600 req/5min). Match metadata records confidence/method/search query; accounts fetch is tracked separately on ch_vet_companies with an LLM-review exclusion filter.
- **sources:** docs/agent_docs/sql_for_tables/023_companies_house_data.sql
- **relations:** business-intel collection pipeline; http_request_log rate monitoring; vet vertical.
- **verify-later:** ch-enricher agent; enrichment coverage counts.

### Business-intel sweep/verify collection pipeline (vet-intel)
- **category:** NEW:business-intel-collection
- **status-signal:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks and #vet-cleanup and #vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query
- **relations:** companies-house enrichment; batch-processing; scheduler self-healing.
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition.

### research_results with source attribution
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type) e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** content grounding; finetuning flywheel (training triples); content_items origin_research_id.
- **verify-later:** research-agent writers; result_type vocabulary.

### Affiliate and products domain
- **category:** NEW:affiliate-and-products
- **status-signal:** partial
- **status-evidence:** Full schema in 004 (products, product_assets, affiliate_programs, affiliate_products, link_registry.affiliate_product_id + requires_disclosure); 043 (2026) still references "the affiliate_products resolver" as the source of product imagery, so the domain is alive but no seeds/operations appear in this unit.
- **what:** Commerce layer: first-party products (pricing incl. price_display "From £99", SEO fields, per-site slug uniqueness) with asset junctions; affiliate networks (tracking param templates, commission terms, API refs) and affiliate_products with cached network data + custom editorial overlay (pros/cons/verdict/rating, content_status cached→enhanced→reviewed) and availability checking. Link registry marks affiliate links and FTC/ASA disclosure requirements.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART7-11; docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql#kind-comment
- **relations:** link-management (registry); product-card/product-grid library components; imagery (product images excluded from planner).
- **verify-later:** affiliate_products resolver code; any populated programs.

### site_chat_turns per-domain chatbot logging
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration drafted with "NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the next free migration number... before applying" — written against a snapshot, application unconfirmed in this unit.
- **what:** End-user chatbot turns from the site chatbot edge worker: one row per prompt/answer (PII), populated by a Layer-1 puller draining the edge sink with idempotent ingest via edge-supplied uuid PK; bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to llm_call_log, GDPR-conscious salted client_ip_hash instead of raw IPs, edge vs ingest timestamps, per-site cascade delete. Explicitly distinct from llm_call_log (build-time flywheel vs end-user data with its own retention/access profile).
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql
- **relations:** llm_call_log; rag-retrieval (context packs / grounding chunks); edge workers.
- **verify-later:** table existence in production; edge worker + Layer-1 puller implementations.

### Sites contact-identity denormalisation
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Applied ALTERs + COALESCE backfills from content_data (company_name/business_name, tagline/slogan, email/contact_email, phone/contact_phone, logo_text fallback chain); one-off content_data patches for live sites.
- **what:** Frequently rendered identity/contact fields promoted from sites.content_data JSONB to first-class columns (company_name, tagline, email, phone, logo_url, logo_text, contact_address) feeding the render context for headers/footers/heads, with content_data retained as the brief-derived store of record.
- **sources:** docs/agent_docs/sql_for_tables/011_sites_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#issue-1a; docs/agent_docs/sql_for_content/001_phone_number.sql
- **relations:** component-based headers render context; site_specs identity aspect (overlapping data — coherence question).
- **verify-later:** which of sites columns vs site_specs.identity is authoritative for rendering today.

### News feed pipeline: content_sources and feed-item lifecycle
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** content_sources DDL (twice-iterated) + seed_boxing_sources function; content_feed_items in 018; applied handler-routing fixes and content-feed-refresh task (6h); live Grok config update to grok-4-1-fast with search_tools for gaswholesalers.
- **what:** Per-site content sources with typed configs — news_search (web search adapter), rss, api_news (LLM news via xAI/Grok incl. prompt_template, hours_lookback, search_tools), scrape (Firecrawl), api_data (structured APIs like BoE rates) — scheduled by fetch_interval/next_fetch_at with error tracking. Fetched items flow through content_feed_items' separate lifecycle (ingested→filtered→relevant→queued→published/rejected/expired/duplicate) with per-site relevance scoring, entity cross-referencing and dedup, becoming a site_work_items row only at publish time. Routing contract: missing_news_sources / stale_news_section / all_sources_erroring → content-feed-orchestrator; missing_news_section → content-gap-planner.
- **sources:** docs/agent_docs/sql_for_tables/027_content_sources_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#028_news_feed_handler_routing_fixes
- **relations:** work queue; latest-news client rendering; scheduler.
- **verify-later:** content-feed-orchestrator workflow; feed item volumes.

### Client-side latest-news JSON rendering
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** Applied rendered_html update installing the IIFE that fetches /data/latest-news.json (headline, subheadline, items[title,url,summary,source,date], insights_url) on gaswholesalers' index; 044 adds formatNewsDate and the redesigned news CSS.
- **what:** News sections render client-side from a static JSON artefact deployed alongside the site (/data/latest-news.json), so news refresh is a data commit, not a page rebuild. Component ships noscript fallback, date humanisation (formatNewsDate expanding "2d ago"), and canonical CSS in css_snippets picked up on the next webdesign run.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#news-feed-js; docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** med JSON export (same static-JSON publishing pattern); css/js snippets.
- **verify-later:** JSON writer for /data/latest-news.json; per-site adoption.

### Site snapshots: point-in-time capture and revert
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** 085 migration with take_site_snapshot / revert_site_to_snapshot plpgsql functions, iterated twice in-file with column-name fixes — indicating it was actually run and debugged against the live schema.
- **what:** Full site state captured into one self-contained JSONB row (survives row deletions): site record key fields, all current site_specs, all pages with their page_components (content_data + rendered_html), nav groups/items, site_components; git_commit_sha links DB state to deployed files. Revert takes a safety pre_revert snapshot first, then supersedes specs, delete-and-reinserts pages/components/nav/site_components and restores site fields — explicitly NOT a git revert and does not touch global content_components templates. Triggers: deploy, manual, pre_edit, scheduled.
- **sources:** docs/agent_docs/sql_for_tables/031_site_snapshots.sql
- **relations:** page_component_history (finer grain); agent snapshot/revert (same philosophy for agents); deployment-github (file-side counterpart).
- **verify-later:** snapshot triggers actually firing on deploy; v_site_snapshots contents.

### AI persona team and departments marketing model
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Applied site_specs updates (075a/076) injecting team and departments JSONB for ai-agent-orchestration.com, finetuning.uk, leopardessconsulting.co.uk, with audience-tuned copy per site.
- **what:** The platform presents itself through named AI managing-agent personas — Archivist (Research), Sentinel (Quality), Quartermaster (Operations) — alongside the human principal, plus an 8-department / 70+ agent structure with per-department agent counts and capability summaries. Stored as identity-spec data consumed by the content writer for team/departments sections; departments-grid component renders it as the leadership-team alternative.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#076
- **relations:** site_specs identity; departments-grid component; pitch/business docs.
- **verify-later:** rendered team/departments sections on the three sites.

### Auth database provisioning
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Raw CREATE DATABASE auth_db / CREATE USER auth_user with a subsequent password ALTER (credentials visible in file).
- **what:** A separate auth_db with its own user for the authentication service, provisioned by hand. The file preserves a real credential — a hygiene finding for stage 2 (secret in docs).
- **sources:** docs/agent_docs/sql_for_tables/021_auth_db.sql
- **relations:** database-and-infrastructure credentials; admin dashboard auth.
- **verify-later:** whether that password is still live (rotate); auth service consumer.
