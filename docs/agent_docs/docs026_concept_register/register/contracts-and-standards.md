# Register — contracts-and-standards

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

57 concepts, consolidated from 78 raw extractions across units U01_docs024_numbered_core, U02_docs024_focus_handoff, U03_idea_uk_section_data, U04_idea_uk, U09_adoption, U10_imagery, U12_docs024_archives, U13_docs024_small_dirs, U15_docs019_running_notes, U16_docs019_design_plans, U17a_docs019_archive_discussions_and_main, U18_sql_for_agents, U19_sql_tables_components, U21_legacy_docs_b, U23_docs_root_vonc, U24a_docs_archive_classic_and_docs024_misc, U24b_docs_archive_finetuning, U24c_docs_archive_traffic_probe, U24f_docs_archive_remaining_small, U25_leopardess_social.

Note on source material: the assigned cluster input file (`.clusters/contracts-standards-locks.md`) contained the entire raw block set duplicated byte-for-byte exactly twice (a mechanical bucketing artifact, verified via diff — not independent re-extraction). Counts and merges below are computed on the de-duplicated set of 78 real raw blocks.

---

### CTS-001 — Page-build-handler pipeline with plan_sections triage (Layer 0) and validate_content gate
- **status:** deployed
- **status-evidence:** 002(4)/002d document the full pipeline; input-schema v2 confirmed live in 003(8)
- **what:** The page build flow: ensure_site_record → load_page_record → plan_sections (resolves each section's input_schema sources; triages into ready/deferred(needs_human_review)/skipped; the page deploys with whatever is ready) → content writer (only ready sections) → validate_content (algorithmic: placeholders, unrendered templates, cross-site contamination, broken links, hallucinated emails; blockers/errors escalate to needs_human_review) → save_sections → deploy. Quality gates run both before generation and after; content writers never fabricate non-llm-sourced data.
- **sources:** 002(4)#Page Build Handler Pipeline; 002d Layer 0; 003(8)#Component Input Schema v2, #Content Validation Contract
- **relations:** input schema v2; needs_section_data; growth budget; validate_content
- **verify-later:** plan_sections_action.go; validate_page_content.go

### CTS-002 — Component input-schema source vocabulary (Tier A–D + renderer; proposed Tier E)
*(merged from 3 raw blocks: Component input schema v2 sources/on_missing vocabulary; Tiered component field classification Tier A/B/C; Component field-source tiers A–D+renderer/proposed Tier E)*
- **status:** partial
- **status-evidence:** Tiers A–D + renderer confirmed live in the component-creator prompt at deployed v1.0.1080 (2026-06-29 gap analysis); Tier E ("feed.{name}") is explicitly "proposed, pending decision" and not built as of the same date.
- **what:** content_components.input_schema declares per-field type/source/required/on_missing/fallback/min_items/llm_guidance. Fields are informally tiered by source: Tier A = llm (voice content, required), Tier B = static (tunable labels, optional, with fallback — later the source of the "Browse All Tools"-on-non-English-sites language problem and the deferred "soft static" override idea), Tier C = site_specs.*/site_assets.* (site data), Tier D = query.* (derived lists, resolved at plan_sections time, not render time), plus a "renderer" source (JS-filled single value with fallback). on_missing vocabulary: use_fallback/skip_field/skip_section/needs_human_review/block. Image fields must be required:false + skip_field + template-gated (imagery is async). Known trap: required:true with on_missing left as skip_field/empty hits the switch default and defers the whole section (see CTS-027). No tier exists yet for content fetched client-side from a JSON feed at runtime; a proposed Tier E would emit a stable-selector DOM shell + external loader for that case.
- **sources:** 003(8)#Component Input Schema v2; 003(8) checklist item 6b; HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#5; FOCUS_language.md#static-fallbacks; RUNNING_NOTES_vonc(36).md#2026-06-29 (GAP 2 CONFIRMED); PLAN_dynamic_sections_and_loaders(4).md#structural-gaps
- **relations:** plan_sections; queryresolve; imagery-async gate; plan_sections required-field deferral trap (CTS-027); generation-time guards for dynamic components (CTS-044); Tier D items-array shape (CTS-040)
- **verify-later:** planSection switch in plan_sections_action.go; component-creator prompt_template tier section; any feed.* source present in live input_schemas

### CTS-003 — content_data is the source of truth; HTML patching rejected as an edit mechanism
*(merged from 2 raw blocks: "content_data is the source of truth"; "Sanctioned content-edit paths")*
- **status:** deployed
- **status-evidence:** 003(8) full section including the two re-render paths; doc 003 quoted verbatim in a 2026-07-09 HANDOFF as still-governing ("this is why HTML patching was rejected as an edit mechanism")
- **what:** Every page_components row stores content_data (structured) + rendered_html (derived); all edits must go through content_data. The light re-render path (rerender_page_sections) regenerates rendered_html from stored content_data merged with freshly-resolved fields, with no LLM call, and persists the merged content_data so rows stay complete render sources; NULL content_data escalates to a full rebuild. Patching rendered_html directly was explicitly rejected as an edit mechanism (it's lost on the next re-render). Designated edit routes exist instead: fix_component_template_action fix types (e.g. remove_element) for template-level changes, with page-component content changes deferred to the section-editor workflow; when no supported path exists, the fallback is a full-text template UPDATE (never a multi-line REPLACE of nested markup), verified by length delta and propagated via a page_rerender item.
- **sources:** 003(8)#Schema Enforcement/#Source of truth; 002(4) page-rerender row; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/PLAN_provocation-card(3).md#method-corrected
- **relations:** work-item routing; section editor; fix_component_template_action; two re-render paths
- **verify-later:** rerender_page_sections persistence semantics; fix_component_template_action.go fix types; section_editor_actions.go component_swap

### CTS-004 — Workflow result contract (result_from / multiple_output_fields / result_mapping) and the dead-key stub bug class
*(merged from 2 raw blocks: "Workflow result contract"; "Result-contract dead-key class and Option A unification")*
- **status:** deployed
- **status-evidence:** 003(8) documents the contract as current; the underlying bug class was fixed and deployed v1.0.1092+ (HANDOFF_builder_thread: "Option A CLOSED")
- **what:** A workflow's complete step declares its result via result_from (flatten — a field's contents become the body), multiple_output_fields (nested per key), or result_mapping; deprecated aliases still resolve with a Warn. Historically CompleteWorkflowAction only honoured output_field/output_fields and otherwise dumped the entire collected_data — result_from was a key the action never read, so diagnose-agent completions always shipped everything, masked until a 515-file analysis blew the ~900k Kafka cap (Kafka Message Size Too Large at 1.27MB). A related instance: the orchestrator pointed output_fields at an imagined key ("diagnose-agent_result") when the engine actually stores a call step's response under the STEP NAME. Fix (Option A, shared datahelpers/result_contract.go): both readers now honour result_from/output_fields, a response-size guard replaced the silent oversize dump with an actionable error, and four diagnose/index agents were migrated to the preferred key. Parents read at `<call_output_field>.response.<key>`; using the wrong mode produces silent null reads, not errors.
- **sources:** 003(8)#Workflow Result Contract; 016 §9 "Child workflow result silently replaced by a stub"; NNN_fix_diagnose_complete_output_fields.sql; NNN_fix_orchestrator_complete_key.sql; HANDOFF_builder_thread.md#2
- **relations:** silent-completion family; call metadata vs response-data convention (CTS-038); bounded bundle egress (persist-and-reference)
- **verify-later:** result_spec.go; datahelpers/result_contract.go; extractFinalResult size guard; remaining deprecated-alias agents

### CTS-005 — Component naming contract (function = canonical kebab-case identifier)
*(merged from 3 raw blocks: "Component naming contract" (U01); "Kebab-case naming contract" (U19, function portion); "Component naming contract" (U21))*
- **status:** deployed
- **status-evidence:** Live DB constraint chk_function_kebab_case + partial unique index on active function (003(8), confirmed again in the 005_content_components.sql migration set)
- **what:** content_components.function is the canonical identifier for a component everywhere in the system: kebab-case, regex-constrained (`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`, empty allowed only for legacy rows), one active row per function (partial unique index). The template's data-component attribute on the root element must match function exactly; page_components.slot_name mirrors it; planners assign and rerenders match by it. GetComponentWithFallback (exact → normalized → generic-text-block) is a safety net, not something to rely on. NormalizeComponentFunction plus this 3-step fallback tolerate legacy data; adoption pipelines must translate external names rather than import them verbatim. This is the direct successor to the earlier data-function attribute convention (CTS-042, superseded).
- **sources:** 003(8)#Component Naming Contract; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md; docs/agent_docs/sql_for_tables/005_content_components.sql#component-naming-standardisation
- **relations:** section_type vs function split (007); slot_name↔function mapping hazard; data-function contract (ancestor, CTS-042); page_type vocabulary (companion kebab constraint, CTS-007)
- **verify-later:** chk_function_kebab_case constraint and unique index in DB; component_validation.go naming checks

### CTS-006 — String-value naming convention (snake for identifiers, kebab for data)
*(merged from 2 raw blocks, same concept documented twice)*
- **status:** deployed
- **status-evidence:** Most recent/specific evidence: "Status: applied (migration 051, page_canonical.go + page_role_validator.go updated, contracts doc v9, debug guide v2.10)" (2026-05-17) — supersedes the earlier "partial, follow-up needed" framing.
- **what:** Decision rule for string-typed columns/enums: values used as a Go identifier (switch case, registry key, dispatch route) are snake_case (e.g. site_work_items.item_type); pure data describing what a thing is is kebab-case (e.g. pages.page_type, content_components.function); single words are bare lowercase. Root incident that motivated the rule: normalisePageType wrote snake_case while all readers expected kebab-case, silently hiding blog pages; companion fix retyped the homepage's page_type from 'index' to 'landing' (separating storage name from page kind). A snake-input fallback is retained as a bounded migration-tail exception. Test suites document behaviour, not intent, per this same audit.
- **sources:** FOCUS_naming_conventions_kebab_vs_snake.md (whole); 003(8)#String-Value Naming Convention
- **relations:** page_type vocabulary (CTS-007); item_key canonicalization; thin-slice constitution (CTS-029, restates the same rule as a standing constitution line)
- **verify-later:** migration 051; CHECK constraint on pages.page_type; whether snake_case columns still lack explicit CHECK constraints

### CTS-007 — page_type vocabulary and "landing, not index"
- **status:** deployed
- **status-evidence:** Constraint chk_page_type_kebab_case live since migration 051; canonical value table documented in 003(8)/016 §6.5
- **what:** Canonical kebab-case page_types: landing/content/tool/guide/game/blog-post/blog-index/section-index/entity-page/entity-directory/news-index. The homepage's TYPE is landing while its NAME is index — name is a storage convention, type is the kind of page. CanonicalisePage normalises legacy snake_case inputs one-way. Guides nest at /guides/<slug>/index.html and appear in guide-lists only when typed guide AND active/deployed.
- **sources:** 003(8)#page_type vocabulary; 016 §6.5; docs/agent_docs/sql_for_tables/003_pages.sql#051_pages_page_type_kebab
- **relations:** CanonicalisePage; string-value naming convention (CTS-006); query-resolver list components (CTS-041)
- **verify-later:** pages page_type distribution; constraint present in live DB

### CTS-008 — JS content separation contract (js_content → /tools/assets/{function}.js)
*(merged from 3 raw blocks describing the same contract at different times/units)*
- **status:** deployed
- **status-evidence:** 003(8) full flow through separateInlineJS/collectJSAssets/multi-file git commit; independently re-confirmed 2026-06-11 ("js_content RESOLVED — the assets-split EXISTS")
- **what:** Component-specific JS is extracted out of html_template into content_components.js_content and served as an asset file at /tools/assets/{function}.js; html_template keeps only a `<script src>` reference. store_generated_component_action.go's separateInlineJS() extracts only attribute-less `<script>` tags (by design); RerenderSinglePageAction.collectJSAssets() assembles the resulting multi-file git commit. js_snippets is a separate table for shared design effects/utilities (e.g. formatNewsDate), never component-specific behaviour. Two known failure classes: (1) pre-extraction rows render as empty shells; (2) verified NOT used by the library-tool pipeline — tool-generator/improver mandate one inline script and the fork INSERT omits js_content, so a library tool adopting the split would fork with a dangling script reference and a 404'ing asset.
- **sources:** 003(8)#JS Content Separation Contract; 003(8)#JS Content Separation Contract (news-components restatement); js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#003-Contracts-and-Standards; NOTES_running_synthesis_principles(59) 2026-06-11
- **relations:** tool doc header stripping; empty-shell taxonomy; fork-divergence detection
- **verify-later:** separateInlineJS regex; whether script-balance validation hardening was applied; grep for remaining `<script>` in html_template

### CTS-009 — Component creation & regeneration contract (created/regenerated; already_exists removed)
- **status:** deployed
- **status-evidence:** 003(8) + 026(2), dated 2026-04-20, describing deployed behaviour
- **what:** StoreGeneratedComponentAction: a Layer-1 pre-store validation runs before either branch (rejection never touches the DB); create INSERTs plus a v1 snapshot; regenerate snapshots the old state then UPDATEs in place (UUID preserved, FKs intact), marks dependent page_components pending, and raises one deduped needs_rerender per affected site (item_key component_regen_rerender:<uuid>). Downstream callers must not assume component_id is new nor create their own rerender items. Regeneration keying is by the LLM's EMITTED function name — a mismatched name INSERTs a stray row rather than updating the intended one.
- **sources:** 003(8)#Component Creation & Regeneration Contract; 026(2) full; 016b#Manually invoking an agent (regeneration keying)
- **relations:** component_versions; markPagesForRebuild; system-stats key-mismatch incident
- **verify-later:** store_generated_component_action.go branches

### CTS-010 — Site component linkage contract (slot_name↔function; fallback header hazard)
- **status:** deployed
- **status-evidence:** 003(8) contract text plus the discovery check unlinked_site_components
- **what:** Every site_components row must have component_id → content_components; otherwise renderAndStoreSiteComponent falls to a generic function lookup (which cannot match, since slot 'header' ≠ function 'header-<variant>') and then to a hardcoded RenderFallbackHeader (no logo, stacked nav, search icon, dark). Breakers: update_site_defaults not run, NULL collection header id, legacy data. A self-healing check plus a site-component-linker handler exist to repair the gap.
- **sources:** 003(8)#Site Component Linkage Contract; 004 discovery checks
- **relations:** four overlapping chrome stores (036); light-site-dark-chrome bug
- **verify-later:** update_site_defaults in workflows; unlinked check registration

### CTS-011 — CSS colour inheritance model (var(--section-*, var(--color-*)) fallback chain)
*(merged from 3 raw blocks describing the same live mechanism)*
- **status:** deployed
- **status-evidence:** Called "the single most important rule in the design system" in 003(8); restated as the renderer contract in the tool-portal-light css_template header
- **what:** Base CSS resolves text/heading/muted/border colours through a two-level chain: h1–h6 use var(--section-heading, var(--color-primary)); p/li/blockquote use var(--section-text, inherit); strong/em/span set no colour; links are the explicit exception. A non-painting section inherits page ink automatically; a painting section overrides --section-* on its own container and every child element follows (see section painting contract, CTS-012). This is what makes "components declare no colours" viable. A component's inline `<style>` is an optional override, not its main CSS (explicit user correction on record) — setting colour directly on elements bypasses the chain (the light-on-light testimonial bug). An "ambient pass-through" fix (`--section-x: var(--color-x)`) exists because some internal consumers lack var() fallbacks and would otherwise fall to currentColor/transparent.
- **sources:** 003(8)#CSS Colour Inheritance Model; 036 §4; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; idea.uk/REPORT_scheme_does_not_reach_components.md#2
- **relations:** section painting contract (CTS-012); buildSectionDefaults; CSS section-colour model historical evolution (CTS-030, the multi-year hardening this is the current end-state of)
- **verify-later:** layouts' element rules follow the fallback pattern; the luminance-appender code location in the renderer

### CTS-012 — Section painting contract (four painting models, references-only) — supersedes literal dark-section overrides
*(merged from 2 raw blocks documenting the same rewrite)*
- **status:** deployed
- **status-evidence:** 003(8) checklist item 6 + Section Context Variable Contract; the (7)→(8) diff shows the literal-colour version replaced, with "fix_forced_text_colors enforces this mechanically". An earlier-dated project doc (RUNBOOK slice4b) shows the rewrite mid-flight ("repo copy still a pending user step") before landing as the canonical 003(8) text — the later, canonical doc is the deployed truth.
- **what:** A template's appearance derives from what its own CSS paints; is_dark_section is catalogue metadata and must not key styling (CTS-021). A painting section picks exactly one of four models — (a) pair band re-exporting the pair text colour, (b) palette band re-exporting the on-colour family, (c) image/layered background defining `--hero-ink` per branch, or (d) ambient (no background of its own and no --section-* at all) — and re-exports --section-* AS REFERENCES to the tokens it paints with (color-mix() for muted/surface/border). Literal colours in --section-* declarations are forbidden. This is the exact inverse of, and supersedes, the older contract where dark sections set literal rgba/white values gated on is_dark_section.
- **sources:** 003_contracts_and_standards(8).md items 6/6b + #Section Context Variable Contract; SPEC_scheme_to_components.md#The-contract; slice4b_003_contract.md
- **relations:** paired-variable standard (CTS-020); is_dark_section demotion (CTS-021); fix_forced_text_colors check; CSS section-colour model historical evolution (CTS-030)
- **verify-later:** fix_forced_text_colors action; component templates conformance

### CTS-013 — CSS theme template contract (renderer vs template ownership; theme storage/lineage)
- **status:** deployed
- **status-evidence:** 003(8) responsibility split; render pipeline confirmed in 036
- **what:** The renderer owns palette injection, luminance-driven --section-* defaults (pickReadableOnBackground, preserving palette character), and css_snippets appends. The theme template owns layout/typography/component styling using the fallback pattern and MUST NOT declare --section-* defaults or hardcode text hexes. css_template (Go template) is distinct from css_content (a frozen fork snapshot, reference only).
- **sources:** 003(8)#CSS Theme Template Contract; 036 §3–4
- **relations:** 025 palette/layout/typography split; buildSectionDefaults; CSS colour inheritance model (CTS-011)
- **verify-later:** render_css_from_spec_action.go; color_util.go

### CTS-014 — Query parameterisation contract ($1 + params, never template interpolation)
- **status:** partial
- **status-evidence:** 003(8) states the rule plus named legacy offenders (tool-suggester, tool-improver) still to migrate
- **what:** All new query_database usage must use $1 placeholders with a params array of dot-paths passed as query args; {{.field}} embedding directly in SQL is a SQL-injection risk. QueryDatabaseAction gained params support after audit agents were found using $1 placeholders with no args array.
- **sources:** 003(8)#Query Database Parameterisation Contract; 001(5) bug 1
- **relations:** authoring rules pack; thin-slice constitution (CTS-029, "parameterised queries only")
- **verify-later:** whether tool-suggester/tool-improver have been migrated

### CTS-015 — Schema enforcement: flexible vs strict mode with approval snapshots
- **status:** abandoned
- **status-evidence:** 003(8) describes the design (schema_snapshot/content_snapshot at approval, sites.schema_mode) with no deployment claim or date. Later, more specific evidence (2026-07-09/10, recorded in the locks material — see LOCK-006) shows the concrete implementation of this design — a BEFORE UPDATE trigger stamping schema_mode='strict' plus lock fields — was stillborn: no Go code ever read schema_mode, the companion snapshot columns were never created, it fired exactly once in the system's history, and it was deliberately dropped via migration 009 in 2026-07-10 (columns/functions retained but orphaned).
- **what:** The design: initial build runs "flexible" (best-effort substitution, warnings); at approval the structure would lock — page_components.schema_snapshot + content_snapshot captured, sites.schema_mode flipped to strict, mismatches become validation errors, and template upgrades can't break approved pages. In practice this subsystem was never wired end-to-end and was actively removed rather than completed.
- **sources:** 003(8)#Schema Enforcement (Flexible vs Strict Mode); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-09-the-dropped-trigger; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.2
- **relations:** auto_lock_on_deploy trigger / stillborn strict-mode subsystem (LOCK-006); content governance approval flows
- **verify-later:** schema_mode column usage (expected: unread by any Go code); 009_drop_auto_lock_on_deploy.sql

### CTS-016 — Handler dispatch input-path contract (input_data.spec.* primacy)
*(merged from 3 raw blocks: "Handler input-path contract" (U01); "Spec-is-primary-input contract" (U02); "Dispatch input contract for handler agents" (U10) — the same mechanism independently rediscovered/redocumented across April and July work)*
- **status:** deployed
- **status-evidence:** Architecture decision recorded 2026-04-17 as root cause of gauntlet pages getting 0 components; independently re-recorded 2026-07-12 as a "hard-won mechanism", showing the contract still needs re-teaching per project
- **what:** The dispatch loop (build-dispatch-loop) reliably populates only input_data.spec.* (plus input_data.site_id/domain/item_type); handlers MUST read spec fields there, never rely on top-level flattening, which exists only for legacy `?` input_mapping promotions and silently resolves nil otherwise. Go actions implement a defensive fallback chain (explicit config → input_data.spec.field → well-known spots); known past offenders were tool-improver/tool-auditor/rerender-pages. Divergence between dispatch-shape and direct-call shape is a live, recurring source of latent bugs — e.g. asset-deployer's check_mode deliberately tests both input_data.spec.mode and input_data.mode. Manual spawn+call of work-item agents must satisfy BOTH the top-level input_contract and the workflow's spec paths.
- **sources:** 003(8)#Handler agent contract/#Input data paths; 016 §9 path-mismatch ("most common systematic failure"); HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#4; HANDOFF_imagery_best_in_class.md#Mechanisms; SQL_2026-07-12_asset_deployer_explicit_paths.sql
- **relations:** dispatch loop; input_contract validation; ExtractActionInputs lesson; local-step input resolution / key_path (CTS-048, a distinct but adjacent input-resolution gap for local action steps)
- **verify-later:** load_page_record_action.go fallback chain; build-dispatch-loop spawn payload construction

### CTS-017 — Legal rules schema and content_direction page-level instructions
- **status:** aspirational
- **status-evidence:** 003(8) schemas are defined; legal-content-agent listed as "Planned" in 002(4)
- **what:** Per-site legal_rules (required disclaimers with triggers/placement, forbidden phrases, required pages seeded per industry) live in sites.content_data; pages.content_direction jsonb (format/instruction) is designed to flow to the content writer for page-level rewrites via needs_rebuild.
- **sources:** 003(8)#Legal Rules Schema, #content_direction
- **relations:** content agent family; compliance discovery
- **verify-later:** any legal_rules populated; content_direction reads in the writer prompt

### CTS-018 — system.internal site convention
- **status:** deployed
- **status-evidence:** "Created for maintenance/library-level work items … id: eac60db8-…, domain: system.internal" (2026-04-17)
- **what:** A never-deployed sites row (brand_dna.is_system=true) that hosts library-level and maintenance work items not belonging to any customer site (e.g. component_quality_scan backfills). Side effect: its maintenance-pipeline items sit dormant (no maintenance-dispatch-loop) and it absorbs untargeted scheduler dispatches.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#7; FOCUS_dispatch_diagnostic(4).md#Q4; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** pipeline soft label; Bug 3 site targeting
- **verify-later:** sites row; items accumulated on it

### CTS-019 — {function}-section class contract and data-component naming contract
- **status:** partial
- **status-evidence:** "The `{function}-section` contract is REAL + operative, honoured unevenly" — honoured by 5 surface sections + footer-with-disclaimer, not by hero (`.hero`) or CTA (`.cta-section`)
- **what:** Layouts and buildSectionDefaults key structural rules and surface treatment on `.{function}-section` class names, but CompilePageSectionsAction concatenates component HTML without wrapping, so the class is each component's own responsibility and adoption is inconsistent — the mechanism misses non-adopters and their inline CSS wins. Separately, every component root does carry `data-component="{function}"` (kebab-case, enforced by component_validation.go), giving an attribute-selector escape hatch the class mismatch cannot break.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se #Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Q4
- **relations:** CSS colour inheritance model (CTS-011); SectionStyles (dead consumer of the same names); section painting contract (CTS-012)
- **verify-later:** component_validation.go naming checks; class emission across content_components.html_template

### CTS-020 — Paired-variable ("on-colour") standard — the decision record
- **status:** deployed
- **status-evidence:** SPEC (2026-07-02): "The standard is paired variables … 18/18 layouts define --color-primary-text, 17/18 define --color-cta-text"; executed via W1–W6 and closed 2026-07-03
- **what:** Every paintable band colour has a matching text colour, curated per layout (and therefore per scheme), overridable per site through palette specialised slots (theme-wins merge), with per-instance control later available via site_plan_directives scope=section. Selected over four alternatives (stale-build, component-owned bands, renderer-owned via is_dark_section, full restructure) after the gating decision that "a light scheme must be able to render fully light, and may carry dark hero bands" — band darkness must be a choice, not a component constant. Existing variable names are reused (--color-header-bg/-text, --color-footer-bg/-text, --color-cta-bg/-text, --color-primary/-text, --color-hero-title/-subtitle); the direction is completion of existing architecture, not restructure.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sn #So; RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS
- **relations:** section painting contract (CTS-012); layout CTA pair curation; is_dark_section demotion (CTS-021)
- **verify-later:** layout css_templates pair definitions; palette specialised-slot merge in composition helpers

### CTS-021 — is_dark_section demoted to catalogue metadata
- **status:** deployed
- **status-evidence:** SPEC consequences: "is_dark_section is demoted to selection/imagery metadata (6 of 37 declarers contradict their own flag — never key styling on it)"; prompt text landed 2026-07-06
- **what:** is_dark_section is a component-level boolean authored by the component-creator LLM (extracted by store_generated_component from generated JSON), scheme-blind (the needs-new-component spec carries no scheme field), and unreliable (6/37 self-declarers contradict their own flag). The decision demotes it: nothing may style from it — styling derives from what the template's CSS actually paints (CTS-012). It survives only as selection/imagery metadata.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sh #Sn; slice4a_creator_prompt.sql
- **relations:** section painting contract (CTS-012); component-creator prompt re-aim (CTS-022)
- **verify-later:** store_generated_component_action.go extraction; component_selector.go non-use; creator prompt current text

### CTS-022 — component-creator prompt re-aim (painting rules, vocabulary, image-fields rule)
- **status:** deployed
- **status-evidence:** "4a evidence: gate t/t/t/t/f → UPDATE 1 t/t/t/f" (2026-07-06); slice4a_creator_prompt.sql RETURNING confirms the new blocks are in and the old DARK SECTIONS block is gone
- **what:** Four targeted needle-replaces inside agent_definitions.default_config->>'prompt_template' for component-creator: the dark-sections literal block becomes the four painting models (references only); the vocabulary gains the cta pair and extended tokens (surface-alt, hairline, code-bg, callout pair); Tier C gains the image-fields rule (described rather than shown in literal syntax, because the prompt itself is Go-template-rendered). Root cause addressed: the prompt was emitting components that consume chrome vars while self-declaring dark text.
- **sources:** slice4a_creator_prompt.sql; running_notes_scheme_to_components(55).md#Uf #Uh #Ui
- **relations:** section painting contract (CTS-012); image fields optional-with-gate contract (CTS-023); component creation contract (CTS-057)
- **verify-later:** agent_definitions component-creator prompt_template current text

### CTS-023 — Image fields optional-with-gate contract
- **status:** deployed
- **status-evidence:** User decision (2026-07-03): "imagery must not block section rendering"; gate applied 2026-07 ("UPDATE 1, gated t/t"); landed in both the creator prompt and 003 item 6b
- **what:** Any site_assets.*-sourced component field MUST be required:false with on_missing:skip_field, and its markup MUST be gated with a template conditional (Go templates treat "" and missing as false, covering the broken-image src="" case). Imagery arrives asynchronously and must never block or defer a section; the section renders imageless and the image is added later by the pipeline's own queued rebuild.
- **sources:** w7a_01_gate.sql; slice4b_003_contract.md#Edit-1 (6b); slice4a_creator_prompt.sql (R4)
- **relations:** plan_sections required-field deferral trap (CTS-027); section-scope imagery pipeline; component-creator prompt re-aim (CTS-022)
- **verify-later:** brief-explanation html_template gate; input_schema on_missing values across image-consuming components

### CTS-024 — Component schema/template/prompt three-way consistency invariant
*(merged from 2 raw blocks: "Component schema-template consistency invariant"; "Template syntax unification and three-way field alignment")*
- **status:** deployed
- **status-evidence:** Checkpoint note: "The governing invariant — a component's schema items must match its template tokens — is the right thing to hold; the reconciler enforces consistency toward the current schema"; earlier fix history in migration 005 already patched the same three-way mismatch class
- **what:** A component's input_schema (shape `{"fields": {...}}`) is the contract for its html_template tokens: array item field names in the schema must match what the template reads, and both must agree with what the LLM prompt asks for — divergence among these three breaks generation, rendering, and the reconciler coherently. A large family of early patches converted Handlebars-style seeds ({{#each}}/{{#if}} → Go {{range}}/{{if}}) and fixed recurring naming mismatches (headline vs title vs section_title; features[].name vs services[].title), including a root-cause fix for stored `<no value>` output apparently written back into a template column. Known unresolved violation: info-card-grid's html_template still literally contains `<no value>`, flagged as its own repair thread; services-grid (byte-identical schema to differentiators) was healed by the same fix.
- **sources:** running_notes_checkpoint_uu.md#Confirmed-during-the-hardening-review; running_notes_checkpoint_ss(1).md#Root-cause-in-code; docs/agent_docs/sql_for_tables/005_content_components.sql#templating-fixes and #fix-no-value
- **relations:** array item-fields contract; render-time reconciler; component library
- **verify-later:** info-card-grid html_template (still `<no value>`?); services-grid spot check; remaining Handlebars syntax in content_components

### CTS-025 — Numbered-flat-fields anti-pattern (25 components)
- **status:** partial
- **status-evidence:** "Twenty-five active components match the numbered-field signature" (May 2026 audit); game-list rewritten to Tier D 2026-06-04; the rest unmigrated
- **what:** Schemas declaring post1_title…post6_* with source:llm force the LLM to fabricate list items by structure (invented games, duplicate entries, links to nonexistent URLs) — no prompt rule can save a schema that demands invention. Groups: 8 navigation components (need a curated nav.* source, not query-resolvable), 7 card/grids (straight items-array rewrites), 5 tier/stat (may fit site_specs.<aspect>.items), 5 tool-internal field clusters (heterogeneous). Component-creator must refuse this shape for new "list of N things" components.
- **sources:** FOCUS_component_schema_patterns.md; migration_game_list_tier_d.sql header; CATALOGUE(9)#family-c
- **relations:** Tier D items-array shape (CTS-040, the fix pattern); curated-list source vocabulary decision (deferred)
- **verify-later:** re-run the `<prefix>1_/2_/3_` audit; component-creator prompt Tier-D block

### CTS-026 — Anti-fabrication content path (llm_field_specs, targeted prompt, merge_with)
- **status:** deployed
- **status-evidence:** "Step 2 status: PASS" (2026-05-12); Step 3 deployed with a verification surface; tool-list rendering real pages confirmed on later runs
- **what:** Closed the gap where page-content-writer fabricated list items despite queryresolve existing: plan_sections resolves query.* before the LLM call and carries a full per-section Component + resolved_data + llm_field_specs (built from llm_guidance) on section_plan.sections_ready; the writer's prompt was rewritten to ask only for source:llm fields (does not enumerate forbidden fields, avoiding the "pink elephant" effect); RenderComponentAction honours merge_with: current_section.resolved_data, overlaying resolved data as authoritative over LLM output.
- **sources:** FOCUS_directory_builder_and_list_components.md#implementation-history; old2/STEP2_changelog.md; old2/STEP3_changelog.md
- **relations:** Tier-D contract; page-content-writer workflow; numbered-flat-fields anti-pattern (CTS-025)
- **verify-later:** plan_sections_action.go llmFieldSpec; RenderComponentAction merge_with

### CTS-027 — plan_sections required-field deferral trap (default-defer on unresolved required fields)
*(merged from 2 raw blocks describing the same defect, found independently by the adoption workstream in June and the leopardess/social workstream in July)*
- **status:** deployed (as a known, worked-around defect — the engine behaviour itself is unchanged; individual components have been patched around it)
- **status-evidence:** June instance resolved: "guide-list empty on guides hub AND root index — RESOLVED (Part 14p) … set cta_url.required=false"; July instance: NOTES_brief-explanation 2026-07-02 traces the identical switch-default mechanism for an illustration field
- **what:** plan_sections' required-field switch handles use_fallback/skip_section/needs_human_review/block but has no skip_field case, and on_missing defaults to skip_field — so a required field whose source is unresolved (an unpopulated site_specs value, or a site_assets asset that doesn't exist) falls through to the switch's default, sets shouldDefer=true, and the whole section is silently dropped from save_page_sections (not errored, just absent) — producing hero-only hubs or missing sections. A query-sourced list field can never defer this way, since the resolver always returns a non-nil (possibly empty) slice, and an llm-sourced field can never defer either. The deliberate, repeated fix pattern is to fix the component schema (set the field required:false with an appropriate fallback/skip_field), not the engine — the defer-for-safety default on required fields is considered defensible. Documented instances: cta_url on guide-list/blog-listing (site_specs.identity.*_index_url unpopulated → href=""; game-list gained a /games/index.html fallback, tool-list still lacks one), and illustration_url on brief-explanation.
- **sources:** running_notes_14(25)#part-14p; HANDOFF_2026-06-06#resolved; NOTES_brief-explanation(5).md#2026-07-02, #2026-07-03; SPEC_provocations-archive-list(1).md#Design-decisions
- **relations:** needs_section_data; silent-fallback link family; component input-schema source vocabulary (CTS-002); image fields optional-with-gate contract (CTS-023); generation-time guards for dynamic components (CTS-044)
- **verify-later:** plan_sections_action.go required-field switch (~line 44340 of dump); content_components cta_url/illustration required flags across sibling list components

### CTS-028 — Chrome templates must be variable-driven (pre-store hardcoded-link gate)
- **status:** aspirational
- **status-evidence:** "Architecture decided, implementation pending" (FOCUS_chrome_templates, 2026-05-06); zero of the active footer templates consume {{.nav_items_html}} today
- **what:** Header/footer components are LLM-generated with hardcoded `<li>` links, freezing nav at generation time and bypassing populate_nav_tables/classifyPagesForNav dedup entirely (a violation of doc 003's explicit rule). Proposed structural enforcement: a pre-store validation gate in store_generated_component_action.go rejecting chrome templates with hardcoded internal links outside {{range}}/{{if}} blocks, plus prompt teaching and a chrome-template-repair migration. Companion cleanup: buildServicesHTML is a parallel, dedup-less nav query ("Tools, Guides, Games, Games, Tools" verbatim) to be dropped once templates use quick_links_html.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-1; old2/HANDOFF_2026-05-07(1)#8
- **relations:** nav dedup guard B-029-1; render-context variables table; site component linkage contract (CTS-010)
- **verify-later:** store_generated_component gates; content_components chrome templates for {{.nav_items_html}} usage; buildServicesHTML existence

### CTS-029 — Thin-slice constitution (always-on rules; future standards rows)
*(merged from 2 raw blocks documenting the same flat-file document)*
- **status:** deployed
- **status-evidence:** "The always-on rules for any task on this codebase. Included in full in every bundle … Later it becomes the standards rows with scope = constitution" (thin_slice_constitution.md)
- **what:** A single always-on rules document, distinct from the task-specific 003 contracts (pulled in only when a task touches them): reuse before recreate; fix structural problems not symptoms; every agent is an orchestrator; reply on the caller's responses topic; workflows thin/complexity in Go; no SQL subworkflows — spawn sub-agents; check schema before SQL; parameterised queries only; snake_case for identifier-shaped values vs kebab-case for data-shaped values; text+CHECK not native enums; soft-delete via deleted_at; no logger.Debug; log orchestration/correlation ids; deployment path git→Actions→B2; plain pragmatic tone. Reinforced session rules: snapshot before any DB change, resolve site_id fresh via domain, don't conclude from partial signals. Explicitly designed to later graduate from a hand-pasted flat file into database rows with scope='constitution'.
- **sources:** docubundle/thin_slice_constitution.md; HANDOFF_2026-06-09#standing-rules; docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/thin_slice_constitution.md
- **relations:** 003 contracts; standards table (aspirational storage form); council-agent charters; string-value naming convention (CTS-006); chassis conventions verified (CTS-034)
- **verify-later:** whether a `standards` table with scope='constitution' was ever actually created (expected absent)

### CTS-030 — CSS section-colour model evolution: inheritance → hardcoded dark-section variables → renderer-computed defaults → token-referencing painting sections
*(merged from 2 raw blocks: a pre-merged 4-source archive finding, extended here with a 5th independent archive source describing the same middle era)*
- **status:** superseded
- **status-evidence:** Five independent archive/legacy lines converge on the same evolution. Earliest baseline (`old/older1/003_contracts_and_standards.md`, `old_design_and_styling/FOCUS_design_and_styling.md`): plain CSS inheritance, dark-background components just set literal color:#fff/inherit, no --section-* variables exist. Middle era (`old/older1/003_contracts_and_standards_v2..v7.md`, `debugging_old/003_contracts_and_standards_v8..v11.md`, `archive_april_26/003e_contracts_and_standards_v5.md`, and independently `docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md`): a --section-* custom-property contract keyed off a boolean is_dark_section column, with LITERAL hardcoded values enforced by ValidateDarkSectionContract() and, per 043, four enforcement layers (DB flag + audit, Go warnings, LLM prompt rules, periodic SQL audits). An intermediate renderer change (`PHASE_4_4_cleanup_summary.md`): buildSectionDefaults began computing these values automatically from palette luminance (WCAG-based). Live (`003_contracts_and_standards(8).md`): the "Section painting contract" — is_dark_section demoted to inert catalogue metadata, and any section that paints its own background must instead RE-EXPORT --section-* as references to theme tokens via one of four models, using color-mix(); literal colours are forbidden and mechanically enforced by fix_forced_text_colors.
- **what:** Documents the multi-year hardening of how section backgrounds get correctly-coloured text: from ad hoc inline colours, to a hardcoded-value contract gated on a boolean flag (which locked every dark section into literal white-on-dark), through a renderer-side automation step, to the current token-referencing "painting" model that treats is_dark_section as inert metadata and derives colours mechanically from the active palette.
- **sources:** old/older1/003_contracts_and_standards.md; old/older1/003_contracts_and_standards_v7.md#"Section Context Variable Contract (Dark Sections)"; docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md; old_design_and_styling/PHASE_4_4_cleanup_summary.md#"Phase 4.5"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Section Context Variable Contract (Painting Sections)"
- **relations:** styling-render-pipeline (036); CSS colour inheritance model (CTS-011, current end-state); section painting contract (CTS-012, current end-state); fix_forced_text_colors action
- **verify-later:** grep deployed component templates for literal #ffffff/rgba(255,255,255, inside --section-* declarations to confirm the old hardcoded pattern is fully gone; inspect fix_forced_text_colors and buildSectionDefaults Go source

### CTS-031 — Component Quality Contract (scoring formula, quality columns)
- **status:** abandoned
- **status-evidence:** Introduced in `003_contracts_and_standards_v6.md`; absent from v7 onward and absent from the live 003(8) doc (no "Component Quality Contract" heading), though quality_score/quality_issues fields still appear inline in the live doc's regeneration-contract JSON examples
- **what:** v6 fully specified a quality-tracking contract for content_components: eight quality columns (template_variable_count, schema_field_count, template_closed, schema_template_synced, has_data_component, quality_score 0–100, quality_checked_at, quality_issues), a scoring formula starting at 100 with fixed deductions per violation, three computation triggers (on-insert, periodic audit by component-quality-auditor, targeted rescan), an automatic needs_component_regeneration item below score 50, and planner preference for higher-scored components. The section vanished between v6 and v7 and was never restored, even though the live architecture doc still lists a component-quality-auditor agent and the live contracts doc still surfaces quality_score/quality_issues as return fields — suggesting the mechanism may partly persist in code/DB while its dedicated documentation disappeared.
- **sources:** old/older1/003_contracts_and_standards_v6.md#"Component Quality Contract"; docs024_key_docs_latest/003_contracts_and_standards(8).md (residual field mentions); docs024_key_docs_latest/002_system_architecture(4).md (component-quality-auditor agent row)
- **relations:** Component creation & regeneration contract (CTS-009); component-quality-auditor agent
- **verify-later:** whether compute_component_quality/ScoreAndPersistComponent and the content_components quality columns still exist and are actively populated

### CTS-032 — query.{name} field-source resolution timing (render-time → plan-time)
- **status:** superseded
- **status-evidence:** Archive v7 "Source prefixes" table: query.{name} resolved "At render time." Live table: resolved "At plan_sections time."
- **what:** In the Component Input Schema v2 Contract's source-prefix table, the query.{name} prefix (used for blog posts/categories lists) moved from being resolved at page-render time to being resolved earlier, during plan_sections, with the result projected into the field's declared shape — consistent with the broader shift toward front-loading data-availability checks (Layer 0 pre-generation triage) rather than discovering missing/stale query data only at render.
- **sources:** old/older1/003_contracts_and_standards_v7.md#"Source prefixes"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Source prefixes"
- **relations:** Component input-schema source vocabulary (CTS-002); Page-build-handler pipeline / Layer 0 (CTS-001)
- **verify-later:** plan_sections Go action for query-prefix handling, confirm it projects results to field shape at plan time

### CTS-033 — Adapter response envelope contract (single-sourced)
- **status:** deployed
- **status-evidence:** Migration(19), 2026-06-10/11: "Envelope contract resolved empirically … The contract now lives once in 035_adapter_guide.md §1 (FOCUS_adapter_design fully merged … and retired)"
- **what:** Resolved from code, not docs: the coordinator claims awaited requests on in_response_to_request_id first (request_id fallback); working adapters use typed body headers with real booleans + ProduceWithValidation; every adapter reads `action` from the BODY; a reply without the right Kafka headers silently falls through to process-as-work and times out (the documented thunder fault — found and fixed in the analyser adapter before deploy). Import-reuse verdict: reuse canonical types for the body, add a local Kafka-header builder. A 003-vs-FOCUS documentation contradiction was settled empirically and single-sourced into 035_adapter_guide.md.
- **sources:** PLAN_workflows_and_actions_migration(19).md
- **relations:** analyser adapter; doc-drift (docs contradicted; code decided); Adapter Response Envelope Contract — conditional traffic-probe application (CTS-054)
- **verify-later:** 035_adapter_guide.md §1; whether 003 §832-890 was replaced with the pointer

### CTS-034 — Chassis conventions verified: text+CHECK, previous_version_id, deleted_at
- **status:** deployed
- **status-evidence:** FOCUS_schema_verification_findings §1 — read directly off the live schemas; applied as corrections across the contract set
- **what:** The live chassis conventions every new table must follow: enumerated values are text with CHECK constraints (never native enums); versioning is version integer + previous_version_id uuid self-FK with unique (type,version); soft delete is deleted_at (never a status=archived); timestamptz defaults now(); jsonb for flexible payloads. This verification pass also corrected wrong reuse assumptions in the contracts docs ("the contracts are corrected to match reality, not the reverse") and confirmed real fields: approval_mode, pipeline, item_key, briefing_questionnaire, input/output_contract, agent_category CHECK set.
- **sources:** FOCUS_schema_verification_findings.md; PLAN_active_config_schema(3).md#2 note
- **relations:** schema-before-SQL discipline; thin-slice constitution (CTS-029)
- **verify-later:** n/a (verified from live schema dumps)

### CTS-035 — Priority profile (order not weights; sealed constraints)
- **status:** aspirational
- **status-evidence:** FOCUS_salience(4) §9: "Representation: an order, not numeric weights … A node stores only its differences from what it inherits … computed on demand"
- **what:** Requirement-relative priority among dimensions (security/speed/simplicity/generality/functionality/cost) lives on an objective-tree node as an order (with sealed/constraint flags), stored as differences-from-inherited and computed on read. Sealed constraints are ancestor-wins legal floors; a change triggers targeted re-validation of descendants holding conflicting overrides. Note: this concept belongs to an exploratory "ED" mediator/salience design track rather than the core agentchassis platform contracts; retained here per assigned-category instructions since no closer category exists.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#9, #9.7
- **relations:** why-chain; mediator; drift detection; Atomic standard (CTS-036, same exploratory track)
- **verify-later:** none

### CTS-036 — Atomic standard (generated-views doc tree)
- **status:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §2: "Optimal unit: the atomic standard, not the document … Documents are generated views over the atoms"
- **what:** The smallest addressable unit is one rule-atom with structured frontmatter (id, concern, scope, applies_to, kind, severity, status, version, supersedes, owner, check, related) and a body split into rule/rationale/examples. Constitution, per-concern handbooks, change-type bundles, and a machine manifest would all be generated views over one source, so nothing drifts between a doc copy and an agent copy. Same exploratory "ED" design track as CTS-035; note this mismatch in relations rather than a category of its own, per consolidation instructions.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#2, #4, #5
- **relations:** mediator routing model; doc-tree adoption; Priority profile (CTS-035); thin-slice constitution (CTS-029, the platform's actual current equivalent)
- **verify-later:** proposed `standards` table

### CTS-037 — Input/output contracts on agent definitions (input_contract / output_contract)
*(merged from 2 raw blocks spanning the concept's early aspirational form and its later enforced form)*
- **status:** deployed
- **status-evidence:** Contract UPDATE statements across many agent SQL files (011/022/024/025/029 etc.); the 2026-07 diagnosis work established the durable, actively-enforced rule that any input a workflow reads must be declared in the contract (137's "spec is UNDECLARED" fix) and that call-site input_mapping must satisfy the callee's contract — an escalation from the earlier, more aspirational docs012-era design.
- **what:** Every agent_definitions row carries input_contract (required/optional fields) and output_contract (produces). Originally a documentation/tracking convenience (docs012-era "expects/required/produces" design, partially realized, with the enforced end becoming ActionInputSpec at the action level), contracts are now also runtime validation hooks: an input the workflow reads must be declared, and call-site input_mapping must satisfy the input_contract (016b §9 spawn+call rule).
- **sources:** 011_site_deployer.sql; 129_wire_diagnosis_subject_threading.sql; 137_recreation_spec_and_note_subject.sql; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-4; docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md
- **relations:** remove-loops plan; workflow_contract_chain view; ActionInputSpec; Handler dispatch input-path contract (CTS-016)
- **verify-later:** chassis contract-validation code path; how strictly contracts fail fast in production; input_contract/output_contract columns' population rate

### CTS-038 — Call metadata vs response-data convention (output_field.response)
- **status:** deployed
- **status-evidence:** v2/001_general_rule.md states it as "The general rule going forward"; confirmed live in coordinator.go ("a step result is stored under BOTH the step name AND its output_field, adapter body under .response")
- **what:** Workflow data-shape convention: when a step calls another agent, call metadata (agent_id, request_id, topics) is stored directly at the step's output_field, while the called agent's response payload lands at output_field.response. Many prompt-template and field-path bugs trace to violating this shape.
- **sources:** sql_for_agents_v2/001_general_rule.md; 116_thunder_training_monitor_worker.sql; 003_site_classifier.sql
- **relations:** template field-path rules (134); input_mapping; Workflow result contract (CTS-004)
- **verify-later:** coordinator.go result-storage code (~L1636/L2408 per 116)

### CTS-039 — Component render modes (template | agent | composite | standalone)
*(merged from 2 raw blocks: schema-level design doc and live-state confirmation)*
- **status:** partial
- **status-evidence:** Columns and comments exist and a design/decision matrix was written early (docs012), but a later backup dump shows all 41 components at render_mode='template'; only 'standalone' (tools) is additionally observed in seeds — the 'agent'/'composite' modes appear designed but unexercised.
- **what:** render_mode declares how a component is produced: 'template' (direct substitution from brief/render_context data, LLM only fills missing schema fields), 'agent' (spawn agent_type with optional agent_workflow, data pulled via data_sources dot-paths, optionally preceded by a research-agent when needs_research=true), 'composite' (child_components list), and later 'standalone' for tools (html_template IS the final output). Design intent (docs012 decision matrix): headers/footers are always template, never generated; FAQ is agent+research; pricing/contact is template+brief data. Pure-structure components never touch an LLM.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART2; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql; docs012_site_maps_and_components/010_component_and_site_architecture.md
- **relations:** component library; standalone tool render; research agent
- **verify-later:** Go render path switch on render_mode; any rows with render_mode in ('agent','composite')

### CTS-040 — Tier D items-array component schema shape
- **status:** deployed
- **status-evidence:** Migration 041 hand-writes tool-list to Tier D; 042 queues guide-list regeneration through the pre-store validator; game-list rewrite mirrors tool-list ("field vocabulary IDENTICAL to tool-list")
- **what:** List components must declare a single items array with a sub-schema (title, url, meta_description, nav_label) plus top-level fields (eyebrow_label, section_heading, section_intro, cta_url, cta_label, card_link_label), replacing the legacy numbered-flat anti-pattern (guide_1_url…guide_6_url) that broke on sites with fewer items than the schema hardcoded. A pre-store validator enforces the structural contract on LLM-regenerated components; rejections land in agent_error_log.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-041, #migration-042, #game-list-rewrite
- **relations:** Query-resolver list components (CTS-041); Component naming contract (CTS-005); Numbered-flat-fields anti-pattern (CTS-025)
- **verify-later:** pre-store validator code; tool-list/guide-list/game-list current schemas

### CTS-041 — Query-resolver list components (pages_where_type) and canonical section URLs
- **status:** deployed
- **status-evidence:** gamesdesign migrations re-type guide pages to page_type='guide' so guide-list (items.source = query.pages_where_type:guide) resolves them, "mirrors the working game-list / page_type=game precedent"; URL migration to /guides/<slug>/index.html
- **what:** List components resolve their items dynamically from the pages table by page_type via a query resolver — no template change needed when pages are added. Depends on canonical page typing and the canonical nested URL shape /<section>/<slug>/index.html produced by CanonicalisePage, making tools/games/guides structural peers.
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql#migration_retype_guides_to_guide, #migration_guides_url_to_canonical
- **relations:** page_type vocabulary (CTS-007); Tier D shape (CTS-040); site-plan page roles
- **verify-later:** queryresolve Go code; link_registry sync after URL moves

### CTS-042 — data-function contract + intelligent component fallback (P1/P2/P3)
- **status:** superseded
- **status-evidence:** docs009/001: "A data-function attribute in the HTML acts as a 'shared contract' … P1 perfect match, P2 good match, P3 generic-text-block — the site always gets built"; superseded by the function/kebab-case + data-component contract (CTS-005), whose GetComponentWithFallback keeps the same 3-step fallback idea
- **what:** The original decoupling of structure from content: the architect assembles empty containers tagged by function (data-function="problem_statement") so the content pipeline can independently fill them; component lookup degrades gracefully (exact function → similar purpose → generic-text-block) so a build never fails for lack of a component.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-1; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#Lookup-Safety-Net
- **relations:** Component naming contract (CTS-005, successor); content_components.function; AssembleFromLibraryAction
- **verify-later:** GetComponentWithFallback in component_library.go; generic-text-block component row

### CTS-043 — Recursive component tree ("everything is a component")
- **status:** abandoned
- **status-evidence:** docs009/001: "We remove is_container … If the HTML template contains {{.Slot_main}}, it IS a container"; the shipped system instead uses a flat section list per page with header/footer injection
- **what:** A radically simplified component model where structure is defined entirely by template placeholders: components declare defined_slots and data_schema; the build plan is itself a component tree the architect walks recursively (RenderNode), handling any nesting depth; themes are just root components; "ghost" components (no wrapper tag) reduce div nesting. Content generation is decoupled by flattening the tree into a content_map of UUID→field requirements. The flat-sections production system never adopted the recursion, though slots re-surface in a later slot-based assembly proposal.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Simplification, #2-The-Recursion-Logic; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Key-Architectural-Principles
- **relations:** slot-based modular assembly (docs018/007); asset bubble-up; content injector pattern
- **verify-later:** defined_slots column on content_components (expected absent or unused)

### CTS-044 — Generation-time guards for dynamic components
- **status:** deployed
- **status-evidence:** 2026-07-06 result: "FIRST LIVE VALIDATION of baking the guards in at generation" — has_marker=t, has_inline_script=f on the created component; guards held through the real pipeline end to end
- **what:** Lessons baked into component GENERATION instead of post-hoc surgery, for components with client-side-fed dynamic content: emit `data-runtime-fill` in the template's section tag at generation (no string-REPLACE marker step); forbid inline `<script>` entirely (behaviour lives in an external loader); make header copy llm-sourced (no deferral risk); list entries pure markup (nothing for the resolver to fail on); hidden clone-template item plus a `[data-…-template]{display:none}` author rule; visible empty state. Declared "the pattern for all future dynamic sections".
- **sources:** docs/SPEC_provocations-archive-list.md#design-decisions; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-084-succeeded; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3, #§8
- **relations:** Component input-schema source vocabulary (CTS-002, proposed Tier E); marker anchoring lesson; hidden-vs-author-CSS lesson; plan_sections required-field deferral trap (CTS-027, the same "must not block render" principle applied to a different data source)
- **verify-later:** component-creator output for any newer dynamic component (marker present at generation?)

### CTS-045 — CSS variable naming convention (--color-*) + creator prompt STRICT RULE
- **status:** deployed
- **status-evidence:** Hero template fixed and verified (magenta CTA, dark bg) 2026-06-24/25; component-creator prompt patched with "USE ONLY THESE NAMES" + STRICT RULE, confirmed in DB 2026-06-24; library-wide audit complete
- **what:** System CSS custom properties follow --color-primary/-secondary/-accent/-background/... naming; LLM-generated components had emitted --primary-color-style names that don't exist in styles.css, silently firing fallback hexes (the "brochure-blue" index bug). Fix: a template REPLACE on the hero component plus a component-creator prompt section explicitly prohibiting the wrong names and separating palette tokens from layout tokens. Documented exception: --archetype-color is intentional per-card tinting with a --color-accent fallback.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:30, #2026-06-24-~16:50; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** post-025 CSS theme flow; component-creator prompt re-aim (CTS-022); legacy-variable "fossilised render" tell (016b chrome entry)
- **verify-later:** component-creator prompt_template section 7; grep new templates for --primary-color

### CTS-046 — API documentation convention (OpenAPI/Swagger + internal API.md)
- **status:** deployed
- **status-evidence:** Archived docs/_archive/api/API_DOCUMENTATION.md is byte-identical to the live docs/API_DOCUMENTATION.md; describes `make swagger`, `make validate-openapi`, per-service API.md files, `*_swagger.go` annotations
- **what:** Two-tier API documentation standard: external customer-facing OpenAPI 3.0 spec at internal/auth-service/api/openapi.yaml with swaggo annotations, and internal per-service API.md files documenting Kafka topics, DB schemas, env vars. Includes a CI workflow that lints the spec and fails on uncommitted swagger regen.
- **sources:** docs/_archive/api/API_DOCUMENTATION.md#external-api-documentation, #internal-api-documentation
- **relations:** superset context for the public/admin API plans; live counterpart docs/API_DOCUMENTATION.md + docs/api/reference.html
- **verify-later:** internal/auth-service/api/openapi.yaml; Makefile targets swagger/validate-openapi

### CTS-047 — Training-data export format (ChatML + metadata sidecar; SFT/DPO negative examples)
- **status:** deployed
- **status-evidence:** FOCUS(21) §2.4e: "Format: ChatML messages with metadata sidecar. Decided 2026-04-22"; iter_0 audit "1,970 total" rows
- **what:** Each training row is a ChatML messages array plus an ignored metadata sidecar (source_log_id, agent_type, step_name, orchestration_id, model, export_version). Code fences must be stripped; edge-case "prose instead of JSON" rows are excluded from SFT (they'd teach wrong-shape output) but kept in llm_call_log as future DPO "rejected" examples. Schema heterogeneity noted: one (agent_type, step_name) covers hero/minimal-hero/header schemas.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4g
- **relations:** part of Flywheel A; export_version enables downstream compatibility checks
- **verify-later:** stripMarkdownFromResponse in ai_actions.go; export_version field

### CTS-048 — Local-step input resolution: input_mapping dead, key_path for loop items
- **status:** deployed
- **status-evidence:** 109b SQL header: "input_mapping did NOT populate input_data.key for this loop substep … deriving the dataset key 40 times"; NOTES(39) "CORRECTS a load-bearing assumption"
- **what:** A load-bearing chassis contract discovered the hard way: the coordinator only resolves input_mapping for call_agent and loop fan-out, not for plain local action steps. Local actions (and local loop substeps) must read values via a config key holding a dot-path, resolved by ExtractActionInputs Strategy 0 / resolveTemplateToken / a key_path. Migration 109b fixed presign_one to read the loop item via key_path:"ckpt_key" instead of the dead input_mapping{key:ckpt_key} that had presigned the dataset key 40×.
- **sources:** phase5/109b_fix_presign_one_loop_item_keypath(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#2, #update-2026-06-06-2, #8
- **relations:** cause of the "presigns the dataset key" failure signature; Handler dispatch input-path contract (CTS-016, distinct mechanism — top-level spec paths vs local-action key_path)
- **verify-later:** loop_expansion_handler.go setLoopVariable; datahelpers ExtractActionInputs

### CTS-049 — Capability gate D5 — requires-backend semantic tag (supersedes intent-probe site-type gate)
*(merged from 2 raw blocks: the superseded original design and its replacement, both from the same decision thread)*
- **status:** partial
- **status-evidence:** plan(11) D5: "Planner gate (to apply): load_components gains AND NOT (… ? 'requires-backend')"; marked "Outstanding: apply the planner query change" — the gate design is settled but not yet wired into the planner query
- **what:** Gates backend-requiring components off static sites by CLASS, not site-type. A component carries semantic_tags:["requires-backend"]; the planner's load_components excludes such components unless opted in via roadmap section_types; the site side sets deploy_config with capabilities:["backend"]; a later audit check compares placed sections' requires-* tags to site capabilities. This replaced an earlier formulation that invented an "intent-probe" site type and gated via a suitable_site_types column, dropped after feedback that the distinguishing feature is the class (has a backend), not a bespoke site type — the invented site type was removed (suitable_site_types: []).
- **sources:** traffic_probe_plan(11).md#decision-5; traffic_probe_plan(4).md#decision-5-open; traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight
- **relations:** supersedes the intent-probe site-type gate (folded in above)
- **verify-later:** build-site-planner workflow JSON load_components query; component semantic_tags vs suitable_site_types columns

### CTS-050 — Class-level rename (probe → site-engine) and env-var churn
- **status:** superseded
- **status-evidence:** running_notes 2026-06-11 "RENAME MAP (every changed name)"; env var PROBE_DB_PATH → ENGINE_DB_PATH, then (2026-06-11) ENGINE_DB_PATH → ENGINE_DATA_DIR
- **what:** When the box became the home of the whole backend-site class (not just probes), "probe" defaults were neutralised to class-level names across engine + deploy artifacts: service/user site-engine, /opt/site-engine, /var/lib/site-engine, /etc/site-engine/site-engine.env, webroots /var/www/vm-sites/<d>, rate zone engine_rl, hook site-engine-deploy. The DB-path env var was renamed twice; store file probe_events.json → intent_events.json → dropped for JSONL. ProbeSearch/ProbeCategory/ProbeFreeText kind constants were kept (they name the feature, not the class).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-component-live, #2026-06-11-store-v2; traffic_probe_runbook(12).md#changelog
- **relations:** supersedes probe-go naming
- **verify-later:** grep for stale PROBE_/probe_events across artifacts

### CTS-051 — /stats endpoint + INTERNAL_API_KEY (stats internal key)
- **status:** deployed
- **status-evidence:** service(24).go "GET /stats key-gated per-host summary"; runbook(12) env table; verified over HTTPS 2026-06-12
- **what:** /stats returns a key-gated per-host summary (visits/events counters), gated by INTERNAL_API_KEY sent as header X-Internal-Key. Unset key → /stats returns 401. The same key doubles as the read-only capture-export key for /events and /access-digest; on the collector side it is stored in deploy_config.engine.stats_key. The env file (not a shell variable) is the source of truth.
- **sources:** deploy_setup/working_dir/service(24).go#header; traffic_probe_runbook(12).md#2
- **relations:** read by CollectIntentEventsAction
- **verify-later:** /etc/site-engine/site-engine.env INTERNAL_API_KEY; deploy_config.engine.stats_key

### CTS-052 — /events export endpoint (P4 collector interface)
- **status:** deployed
- **status-evidence:** running_notes 2026-06-12 "GET /events built + tested … Tests green ×6"; runbook(12) §6
- **what:** GET /events streams stored events as key-gated NDJSON oldest-first (original line bytes preserved); params since (RFC3339, strictly-after), host, limit (default 5000). Final line `{"_meta":{count,truncated,server_time}}` aids the collector checkpoint. Lock-free by design so a big export never blocks live captures; a torn mid-append tail line is skipped and arrives next pull.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-events-export; traffic_probe_runbook(12).md#6
- **relations:** consumed by CollectIntentEventsAction; the pull architecture
- **verify-later:** store.go StreamEvents; nginx /events location

### CTS-053 — Wrapper-orchestrator requirement finding (001:405-462)
- **status:** partial
- **status-evidence:** running_notes 2026-06-13(d): "STRUCTURAL finding … wrapper-orchestrator REQUIRED (001:405-462): the collector is reached via the SCHEDULER … AND does substantive in-chassis work → must NOT run in a shared agent-chassis pod"
- **what:** Because the collector is reached via the generic scheduler entry point AND does substantive work (HTTP to N boxes + multi-row upserts, unbounded as boxes grow), it must not run in a shared agent-chassis pod. Fix = a thin orchestrator that spawns a worker child into its own pod. Also corrected: the scheduler fires ONE message per tick and does NOT fan out pre_query rows, so the collector self-queries and loops, and the scheduled_tasks pre_query is a count>0 GATE.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-d, #2026-06-13-c, #standing-observations
- **relations:** shaped the two-agent topology; scheduled_tasks intent-collection target corrected to orchestrator
- **verify-later:** 001 guide §405-462; scheduled_tasks intent-collection target_topic

### CTS-054 — Adapter Response Envelope Contract (003) — conditional application, later dropped from plan
- **status:** superseded
- **status-evidence:** plan(4)/(5)/(6) P4/P5: "If collection runs as a chassis adapter, it MUST follow the Adapter Response Envelope Contract"; plan(11) reduces this to a one-line parenthetical once P4 was redesigned to need no adapter
- **what:** The guidelines-audit flagged that IF P4 collection or the P5 deployer were built as chassis adapters, replies must use the typed-struct envelope (see CTS-033) — getting it wrong means a silent drop until timeout (the documented thunder fault). Once P4 was redesigned as a key-gated HTTPS pull + one local action (no adapter needed), the prominent P4/P5 envelope warnings in the plan were demoted. This is a project-specific applicability decision about the general envelope contract, not a change to the contract itself.
- **sources:** traffic_probe_plan(6).md#phases; traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit; traffic_probe_plan(11).md#p4
- **relations:** applies only if a P5 vmhost adapter is built; Adapter response envelope contract (CTS-033, the general contract this decision references)
- **verify-later:** ProduceWithValidation usage if a vmhost adapter is ever built

### CTS-055 — Section resolvers override content_data on every render
- **status:** deployed
- **status-evidence:** HANDOFF §4.4 (2026-07-12): "hero background_image is auto-resolved to /assets/images/hero.jpg (plan_sections_action.go:1338) — you change the FILE"; a system-stats section was removed because its suffixes couldn't be overridden
- **what:** Field resolution beats stored instance data on re-render: the hero background image is re-resolved to a fixed path every render (content edits can't remove it — you must replace the file); source:"static" schema fields re-apply their schema fallback every render and cannot be overridden per instance; a forked section component does not survive rerender because save_page_sections re-links to the canonical component by function, whereas a header/footer fork sticks because it's wired via the style collection instead.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#landmine-14; docs/leopardessconsulting/HANDOFF.md#4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12
- **relations:** Static-source schema fields force fleet-generic labels (CTS-056); content_data source of truth (CTS-003)
- **verify-later:** plan_sections_action.go resolvers; save_page_sections re-link behaviour

### CTS-056 — Static-source schema fields force fleet-generic labels/suffixes
- **status:** partial
- **status-evidence:** RUNBOOK_minilobby §0 (2026-07-12): "FIXED … migration 090 — stat_1/2/3_label + cta_label were source='static' … forcing 'Clients Served'/'Satisfaction Rate' … on every render"; a system-stats suffix issue was worked around by removing the section rather than fixing the schema
- **what:** A recurring defect class: shared components whose label/suffix fields are source:'static' re-apply business-generic fallbacks on every render, so writers can only fill values, producing crossed pairs ("2,767%", "500+ Models / Clients Served") on every site. Fix pattern: flip source static→llm keeping the fallback as a safety net (migration 090 for content-block-about, across 13 pages/5 sites); a suffix-free stats component is a noted platform addition still needed.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/090_content_block_stat_labels_llm.sql; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12
- **relations:** Section resolvers override content_data (CTS-055); shared component library semantics
- **verify-later:** content_components 4e448d51 input_schema; system-stats schema fields

### CTS-057 — Component creation contract (the generator's embedded rulebook)
- **status:** deployed
- **status-evidence:** 003d carries the full contract prompt "included in the component-creator agent definition"; generated components (brief-explanation regen, archive-list) conform
- **what:** The complete contract an LLM must follow when generating a component: scoped `<style>` + `<section class="{function}-section" data-component="{function}">` + optional IIFE script; kebab-case function naming; {{.field}} template variables with a declared input_schema; all colours via CSS variables with fallbacks and dark-section --section-* custom properties; scoping and responsive rules; client-side-only JS, no CDN; quality bans (no placeholders, no unrendered variables, no fabricated content). Compiled from docs 003 + 018 + input-schema v2; storage location decision: static in the agent prompt for now, migrating to knowledge_base + rag_lookup if contracts start to churn.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Component-creation-contracts
- **relations:** component selector/creator; generation-time guards for dynamic components (CTS-044); component-creator prompt re-aim (CTS-022); JS content separation contract (CTS-008); RAG knowledge base (RAGR-001, the proposed future storage location)
- **verify-later:** component-creator agent_definitions prompt_template vs the 003d contract text
