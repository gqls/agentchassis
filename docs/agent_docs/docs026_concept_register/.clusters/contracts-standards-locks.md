# Cluster: contracts-standards-locks
Categories included: contracts-and-standards, locks, new:rag-retrieval


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Page-build-handler pipeline with plan_sections triage (Layer 0) and validate_content gate
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 002(4)/002d full pipeline; input-schema v2 in 003(8)
- **what:** ensure_site_record → load_page_record → plan_sections (resolve each section's input_schema sources; triage into ready/deferred(needs_human_review)/skipped; page deploys with whatever is ready) → content writer (only ready sections) → validate_content (algorithmic: placeholders, unrendered templates, cross-site contamination, broken links, hallucinated emails; blockers/errors → needs_human_review) → save_sections → deploy. Quality gates before generation AND after; content writers never fabricate non-llm-sourced data.
- **sources:** 002(4)#Page Build Handler Pipeline; 002d Layer 0; 003(8)#Component Input Schema v2, #Content Validation Contract
- **relations:** input schema v2; needs_section_data; growth budget
- **verify-later:** plan_sections_action.go; validate_page_content.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component input schema v2 (sources, on_missing vocabulary)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full contract; 016b deferral-drop entry shows live enforcement including the default:defer trap
- **what:** content_components.input_schema declares per-field type/source/required/on_missing/fallback/min_items/llm_guidance. Source prefixes: llm, site_specs.*, site_assets.*, pages.*, config.*, renderer, static, query.*. on_missing: use_fallback/skip_field/skip_section/needs_human_review/block. Image fields must be required:false + skip_field + template-gated (imagery is async). Trap: required:true with on_missing skip_field/empty hits the switch default and defers the section.
- **sources:** 003(8)#Component Input Schema v2; 003(8) checklist 6b; 016b#Regenerated content section deferred
- **relations:** plan_sections; queryresolve; imagery async
- **verify-later:** planSection switch in plan_sections_action.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### content_data is the source of truth (rendered_html is derived)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full section incl. the two re-render paths; "HTML patching was rejected as an edit mechanism"
- **what:** Every section stores content_data (structured) + rendered_html (derived). All edits go through content_data; the light path (rerender_page_sections) re-renders from stored content_data ⊕ fresh-resolved fields with no LLM, persisting the merged content_data so rows stay complete render sources; NULL content_data escalates to full rebuild.
- **sources:** 003(8)#Schema Enforcement/#Source of truth; 002(4) page-rerender row
- **relations:** work-item routing; section editor
- **verify-later:** rerender_page_sections persistence semantics

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Workflow result contract (flatten vs fields vs mapping; output_field/output_fields foot-gun)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) opening contract; fix deployed 2026-06-18 per 016 §9 (resolveResultSpec)
- **what:** A workflow's complete step declares its result via result_from (flatten — field contents become body), multiple_output_fields (nested per key), or result_mapping; deprecated aliases still resolve with a Warn. No key → fallback dump of collected_data, which can breach the ~900k cap. Parents read at `<call_output_field>.response.<key>`; wrong mode = silent null reads. Historically singular `output_field` was silently ignored → stub-with-success (Part 1 bug); the oversize path now returns an actionable error instead of a stub.
- **sources:** 003(8)#Workflow Result Contract; 016 §9 "Child workflow result silently replaced by a stub"
- **relations:** silent-completion family; result_spec.go
- **verify-later:** result_spec.go; remaining deprecated-alias agents

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component naming contract (kebab function, data-component, uniqueness)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) with live DB constraint chk_function_kebab_case and partial unique index
- **what:** content_components.function is the canonical identifier (kebab, regex-constrained, one active row per function); data-component attribute on the root element must match exactly; page_components.slot_name mirrors function; GetComponentWithFallback (exact→normalized→generic-text-block) is a safety net not to be relied on. Naming patterns for page-specific heroes and header/footer/head variants.
- **sources:** 003(8)#Component Naming Contract
- **relations:** section_type vs function split (007); slot_name↔function mapping hazard
- **verify-later:** chk_function_kebab_case; idx_content_components_unique_active_function

<!-- SOURCE: U01_docs024_numbered_core.md -->
### String-value naming convention (snake for identifiers, kebab for data)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): kebab CHECKs live; "snake_case columns have not yet had explicit constraints added — follow-up"
- **what:** Values used as identifiers in code (map keys, switch cases, dispatch routes) are snake_case (item_type, action names); pure-data values that end up in CSS/URLs/HTML are kebab-case (function, page_type, agent type); single words lowercase. Decision test: is the value ever a Go case/map key?
- **sources:** 003(8)#String-Value Naming Convention
- **relations:** page_type vocabulary; item_key canonicalization
- **verify-later:** snake-case CHECK constraints existence

<!-- SOURCE: U01_docs024_numbered_core.md -->
### page_type vocabulary and "landing, not index"
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8)/016 §6.5: constraint chk_page_type_kebab_case since migration 051; canonical value table
- **what:** Canonical kebab page_types (landing/content/tool/guide/game/blog-post/blog-index/section-index/entity-page/entity-directory/news-index); the homepage's TYPE is landing while its NAME is index (name is storage convention, type is kind-of-page). CanonicalisePage normalises legacy snake inputs one-way. Guides nest at /guides/<slug>/index.html and appear in guide-lists only when typed guide AND active/deployed.
- **sources:** 003(8)#page_type vocabulary; 016 §6.5
- **relations:** CanonicalisePage; adoption slug-mangling
- **verify-later:** pages page_type distribution; constraint present

<!-- SOURCE: U01_docs024_numbered_core.md -->
### JS content separation contract (js_content → /tools/assets/{function}.js)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full flow through separateInlineJS/collectJSAssets/multi-file git commit
- **what:** Component JS is split out of html_template into js_content and served as an asset file; html_template keeps only a `<script src>` reference. separateInlineJS extracts only attribute-less `<script>` tags (by design). js_snippets is a separate table for shared design effects, never component behaviour. Known failure class: pre-extraction rows render as empty shells (016b entry).
- **sources:** 003(8)#JS Content Separation Contract; 016b#Data-driven component shells render empty
- **relations:** tool doc header stripping; empty-shell taxonomy
- **verify-later:** separateInlineJS regex; script-balance validation hardening applied?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component creation & regeneration contract (created/regenerated; already_exists removed)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) + 026(2) dated 2026-04-20 describing deployed behaviour
- **what:** StoreGeneratedComponentAction: Layer-1 pre-store validation runs before either branch (rejection never touches the DB); create INSERTs + v1 snapshot; regenerate snapshots old state then UPDATEs in place (UUID preserved, FKs intact), marks dependent page_components pending, and raises one deduped needs_rerender per affected site (item_key component_regen_rerender:<uuid>). Downstream must not assume component_id is new nor create its own rerender items. Regen keying is by the LLM's EMITTED function — a mismatched name INSERTs a stray.
- **sources:** 003(8)#Component Creation & Regeneration Contract; 026(2) full; 016b#Manually invoking an agent (regeneration keying)
- **relations:** component_versions; markPagesForRebuild; system-stats key-mismatch incident
- **verify-later:** store_generated_component_action.go branches

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site component linkage contract (slot_name↔function; fallback header hazard)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract + discovery check unlinked_site_components
- **what:** Every site_components row must have component_id → content_components; otherwise renderAndStoreSiteComponent falls to a generic function lookup (which cannot match because slot 'header' ≠ function 'header-<variant>') then to hardcoded RenderFallbackHeader (no logo, stacked nav, search icon, dark). Breakers: update_site_defaults not run, NULL collection header id, legacy data. Self-healing check + site-component-linker handler exist.
- **sources:** 003(8)#Site Component Linkage Contract; 004 discovery checks
- **relations:** four overlapping chrome stores (036); light-site-dark-chrome bug
- **verify-later:** update_site_defaults in workflows; unlinked check registration

<!-- SOURCE: U01_docs024_numbered_core.md -->
### CSS colour inheritance model (--section-* with fallbacks)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8): "the single most important rule in the design system"
- **what:** Base CSS: body sets --color-text; h1-h6 use var(--section-heading, var(--color-primary)); p/li/blockquote use var(--section-text, inherit); strong/em/span set no color; links are the explicit exception. Painting sections override --section-* on their container and all children adapt. Setting color directly on elements bypasses the override — the light-on-light testimonial bug.
- **sources:** 003(8)#CSS Colour Inheritance Model; 036 §4
- **relations:** section painting contract; buildSectionDefaults
- **verify-later:** layouts' element rules follow the fallback pattern

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section painting contract (four models, references-only) — supersedes literal dark-section overrides
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) checklist item 6 + Section Context Variable Contract; (7)→(8) diff shows the literal-colour version replaced; "fix_forced_text_colors enforces this mechanically"
- **what:** A template's appearance derives from what its own CSS paints; is_dark_section is catalogue metadata and must not key styling. A painting section picks exactly one model — pair band, palette band, image/layered (hero-ink), or ambient (no --section-* at all) — and re-exports --section-* AS REFERENCES to the tokens it paints with (color-mix for muted/surface/border). Literal colours in --section-* declarations forbidden. The older contract (dark sections set literal rgba/white values) is superseded.
- **sources:** 003_contracts_and_standards(8).md items 6/6b + #Section Context Variable Contract; 003(7) (family-delta, superseded form)
- **relations:** scheme-to-components work (016b light-site-dark-chrome); forced_text_colors check
- **verify-later:** fix_forced_text_colors action; component templates conformance

<!-- SOURCE: U01_docs024_numbered_core.md -->
### CSS theme template contract (renderer vs template ownership; theme storage/lineage)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) responsibility split; render pipeline confirmed in 036
- **what:** Renderer owns palette injection, luminance-driven --section-* defaults (pickReadableOnBackground preserving palette character), and css_snippets appends; the theme template owns layout/typography/component styling using the fallback pattern and MUST NOT declare --section-* defaults or hardcode text hexes. css_template (Go template) vs css_content (frozen fork snapshot, reference only).
- **sources:** 003(8)#CSS Theme Template Contract; 036 §3–4
- **relations:** 025 palette/layout/typography split; buildSectionDefaults
- **verify-later:** render_css_from_spec_action.go; color_util.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Query parameterisation contract ($1 + params, never template interpolation)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): rule + named legacy offenders (tool-suggester, tool-improver) still to migrate
- **what:** All new query_database usage must use $1 placeholders with a params array of dot-paths (passed as query args); {{.field}} embedding is a SQL injection risk. QueryDatabaseAction gained params support after audit agents failed on $1-with-no-args.
- **sources:** 003(8)#Query Database Parameterisation Contract; 001(5) bug 1
- **relations:** authoring rules pack
- **verify-later:** tool-suggester/tool-improver migrated?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Schema enforcement: flexible vs strict mode with approval snapshots
- **category:** contracts-and-standards
- **status-signal:** unknown
- **status-evidence:** 003(8) describes the design (schema_snapshot/content_snapshot at approval, sites.schema_mode) with no deployment claim or date
- **what:** Initial build runs flexible (best-effort substitution, warnings); at approval the structure locks: page_components.schema_snapshot + content_snapshot captured, sites.schema_mode → strict, mismatches become validation errors, template upgrades can't break approved pages.
- **sources:** 003(8)#Schema Enforcement (Flexible vs Strict Mode)
- **relations:** locks; content governance approval flows
- **verify-later:** schema_mode column usage; any strict-mode site

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Handler input-path contract (input_data.spec.*) + action-level defense
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract rule; 016 §9 "most common systematic failure" documents the violations
- **what:** The dispatch loop passes the work-item spec nested at input_data.spec; handlers MUST read spec fields there, never rely on top-level flattening (which exists only for legacy `?` promotions and silently nils). Go actions reading common fields implement a fallback chain (explicit config → input_data.spec.field → well-known spots). Known offenders were tool-improver/tool-auditor/rerender-pages. Manual spawn+call of work-item agents must satisfy BOTH the top-level input_contract AND the workflow's spec paths (provide fields in both shapes).
- **sources:** 003(8)#Handler agent contract/#Input data paths; 016 §9 path-mismatch; 016b#Manually invoking an agent
- **relations:** dispatch loop; input_contract validation
- **verify-later:** load_page_record_action.go fallback chain

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Legal rules schema and content_direction page-level instructions
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** 003(8) schemas defined; legal-content-agent "Planned" in 002(4)
- **what:** Per-site legal_rules (required disclaimers with triggers/placement, forbidden phrases, required pages seeded per industry) in sites.content_data; pages.content_direction jsonb (format/instruction) flows to the content writer for page-level rewrites via needs_rebuild.
- **sources:** 003(8)#Legal Rules Schema, #content_direction
- **relations:** content agent family; compliance discovery
- **verify-later:** any legal_rules populated; content_direction reads in writer prompt

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tiered component field classification (Tier A voice / B tunable static / C site data)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "revise_component_creator_prompt.sql (applied)" (2026-04-17)
- **what:** Component schemas classify fields: Tier A voice content (source llm, required), Tier B tunable labels (source static, optional, with fallback), Tier C site data (source site_specs.*/site_assets.*). Prevents both "35 required fields" and "0 fields, everything hardcoded". Template/schema sync invariant: every {{.x}} has a schema entry and vice versa. Tier B static fallbacks later become the language problem ("Browse All Tools" on non-English sites) and the "soft static" override idea.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#5; FOCUS_language.md#static-fallbacks
- **relations:** LLM reliability; language surfaces
- **verify-later:** component-creator prompt in agent_definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### system.internal site convention
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Created for maintenance/library-level work items … id: eac60db8-…, domain: system.internal" (2026-04-17)
- **what:** A never-deployed sites row (brand_dna.is_system=true) that hosts library-level and maintenance work items not belonging to any customer site (e.g. component_quality_scan backfills). Side effect: its maintenance-pipeline items sit dormant (no maintenance-dispatch-loop) and it absorbs untargeted scheduler dispatches.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#7; FOCUS_dispatch_diagnostic(4).md#Q4; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** pipeline soft label; Bug 3 site targeting
- **verify-later:** sites row; items accumulated on it

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### String-value naming convention (identifier-shaped snake, data-shaped kebab)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Status: applied (migration 051, page_canonical.go + page_role_validator.go updated, contracts doc v9, debug guide v2.10)" (2026-05-17)
- **what:** Decision rule for string-typed columns/enums: used as a Go identifier (switch case, registry key, dispatch route) → snake_case (site_work_items.item_type); pure data describing what a thing is → kebab-case (pages.page_type, content_components.function); single word → bare lowercase (statuses). Root incident: normalisePageType wrote snake while all readers expected kebab, silently hiding blog pages. Companion fix: homepage page_type 'index' → 'landing' (name vs type conflation). Snake-input fallback retained as a bounded migration-tail exception; tests document behaviour, not intent.
- **sources:** FOCUS_naming_conventions_kebab_vs_snake.md (whole)
- **relations:** Tension #2 canonicalisers; page_type vocabulary gap
- **verify-later:** migration 051; CHECK constraint on pages.page_type

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Spec-is-primary-input contract for handler workflows
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Spec is primary input (contract rule) — all handler workflow configs must use input_data.spec.* paths" (architecture decision, 2026-04-17); root cause of gauntlet pages getting 0 components
- **what:** Dispatch only reliably populates input_data.spec.*; top-level flattened paths (input_data.page_name) depend on optional `?` input_mapping and silently resolve nil. Handler configs use spec paths; Go actions keep a defensive fallback chain.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#4, #Architecture-decisions
- **relations:** chassis input conventions; flat-namespace collisions
- **verify-later:** contracts doc handler-agent section

<!-- SOURCE: U03_idea_uk_section_data.md -->
### {function}-section class contract and data-component naming contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Notes (Sf): "The `{function}-section` contract is REAL + operative, honoured unevenly" — honoured by the 5 surface sections + footer-with-disclaimer; not by hero (`.hero`) or CTA (`.cta-section`).
- **what:** Layouts and `buildSectionDefaults` key structural rules and surface treatment on `.{function}-section` class names, but the compiler (`CompilePageSectionsAction`) concatenates component HTML without wrapping, so the class is each component's own responsibility and adoption is inconsistent — the mechanism misses non-adopters and their inline CSS wins. Separately, every component root does carry `data-component="{function}"` (kebab-case, enforced by component_validation.go), giving an attribute-selector escape hatch the class mismatch cannot break.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se #Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Q4
- **relations:** Colour Inheritance Model; SectionStyles (dead consumer of the same names); section painting contract.
- **verify-later:** component_validation.go naming checks; class emission across `content_components.html_template`.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Colour Inheritance Model (var(--section-*, var(--color-*)) chains)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** HANDOFF §Established: "Element rules follow the Colour Inheritance Model: `var(--section-*, var(--color-*))`."
- **what:** The base element rules in layouts/base CSS resolve text/heading/muted/border colours through a two-level chain: a section-scoped custom property if declared, else the palette-level colour. This is what makes "components declare no colours" viable: a non-painting section inherits page ink automatically; a painting section re-exports its context onto the `--section-*` layer and every child element follows. The W3e "ambient pass-through" fix (`--section-x: var(--color-x)`) exists because some internal consumers lack var() fallbacks — deletion would fall to currentColor/transparent.
- **sources:** HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#The-contract; running_notes_scheme_to_components(55).md#Sx (fix rationale)
- **relations:** section painting contract; buildSectionDefaults.
- **verify-later:** base head CSS / layout css_templates element rules.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Paired-variable ("on-colour") standard — the decision record
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC (2026-07-02) "The standard is **paired variables** … Checks 4b/4c show it is the existing library standard incompletely adopted: 18/18 layouts define `--color-primary-text`, 17/18 define `--color-cta-text`"; executed via W1–W6 and closed 07-03.
- **what:** Every paintable band colour has a matching text colour, curated per layout (and therefore per scheme), overridable per site through palette specialised slots (theme-wins merge), with per-instance control later available via `site_plan_directives` scope=section. Selected over four alternatives (Alt 0 stale-build, Alt A component-owned bands, Alt B renderer-owned via is_dark_section, Alt full-025) after the user's gating answer: "a light scheme must be able to render fully light, and may carry dark hero bands" — band darkness must be a choice, not a component constant. Existing names are reused (`--color-header-bg/-text`, `--color-footer-bg/-text`, `--color-cta-bg/-text`, `--color-primary/-text`, `--color-hero-title/-subtitle`); the direction is completion of existing architecture, not restructure.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sn #So; HANDOFF_scheme_to_components_for_claude_code(1).md#The-Decision; RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS
- **relations:** section painting contract (its component-facing rules); layout CTA pair curation; is_dark_section demotion; supersedes SectionStyles and defers Phase 4.5.
- **verify-later:** layout css_templates pair definitions; palette specialised-slot merge in composition helpers.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section painting contract (003 item 6 rewrite: four painting models)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** slice4b delivers the 003 rewrite as a patched doc; RUNBOOK 07-06-night: "Slice 4b: DELIVERED as a patched doc … STEP A — copy `outputs/003_contracts_and_standards_7_.md` over the repo's 003 doc" — repo copy still a pending user step at last dated evidence.
- **what:** Replaces 003's "Dark Section Contract" (if is_dark_section=true, template MUST set `--section-*`) with: a template's appearance derives from what its own CSS paints; a painting section chooses exactly one model and re-exports `--section-*` AS REFERENCES ONLY — (a) pair band re-exporting the pair text, (b) palette band re-exporting the on-colour family (`--color-primary-text, var(--color-background)`), (c) image/layered background defining `--hero-ink` per branch, (d) ambient: no background of its own and NO `--section-*` at all. Literal colours in `--section-*` declarations are forbidden; muted/border/surface derive via `color-mix`. The old contract is the exact inverse — the concept records a full contract reversal.
- **sources:** SPEC_scheme_to_components.md#The-contract; slice4b_003_contract.md; running_notes_scheme_to_components(55).md#Sh (old 003 item 6) #Ui
- **relations:** paired-variable standard; component-creator prompt re-aim; fix_forced_text_colours re-aim (mechanical enforcer); image fields rule 6b.
- **verify-later:** repo `docs/.../003_contracts_and_standards*.md` item 6/6b current text; whether outputs copy landed.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### is_dark_section demoted to catalogue metadata
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC consequences: "`is_dark_section` is demoted to selection/imagery metadata (6 of 37 declarers contradict their own flag — never key styling on it)"; slice4a prompt text landed 07-06 ("catalogue metadata ONLY").
- **what:** `is_dark_section` is a component-level boolean authored by the component-creator LLM (`store_generated_component` extracts it from generated JSON), scheme-blind (the needs-new-component spec carries no scheme field, so the library skewed dark independent of sites), and unreliable (6/37 self-declarers contradict their own flag). The decision demotes it: nothing may style from it — styling derives from what the template's CSS paints. It survives only as selection/imagery metadata. The earlier Q5/E design question ("where does per-section contrast intent live — site_plan_sections?") dissolves under the paired-variable model.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sh #Sn; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (flag hygiene); slice4a_creator_prompt.sql
- **relations:** section painting contract; fix_forced_text_colours re-aim (isDarkSection param kept-ignored).
- **verify-later:** store_generated_component_action.go extraction; component_selector.go non-use; creator prompt current text.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### component-creator prompt re-aim (painting rules, vocabulary, image-fields rule)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Notes (Ui): "4a evidence: gate t/t/t/t/f → UPDATE 1 t/t/t/f" (2026-07-06); slice4a_creator_prompt.sql RETURNING confirms SECTION PAINTING + IMAGE FIELDS RULE in, old DARK SECTIONS block gone.
- **what:** Four targeted needle replaces inside `agent_definitions.default_config->>'prompt_template'` for component-creator: the dark-sections literal block becomes the four painting models (references only; is_dark_section reported honestly but "nothing may style from it"); the consumer chain replaces the dark line; item 7's vocabulary gains the cta pair and extended tokens (surface-alt, hairline, code-bg, callout pair); Tier C gains the image-fields rule (site_assets.* fields required:false + skip_field + gated markup — described rather than shown because the prompt is itself Go-template-rendered and literal if-syntax would execute). Root cause it addresses: the generated half-migrated footer and brief-explanation proved the prompt was emitting components that consume chrome vars while self-declaring dark text — drift continues until the contract lives in the prompt.
- **sources:** slice4a_creator_prompt.sql; running_notes_scheme_to_components(55).md#Uf #Uh #Ui; RUNBOOK_scheme_to_components(50).md#CHECK-2-RESULTS (corollary)
- **relations:** section painting contract; agent re-registration vs re-seed risk; image fields optional-with-gate.
- **verify-later:** agent_definitions component-creator prompt_template current text; the Step C grep ("DARK SECTIONS (if the section has a dark background)" in Go sources).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Image fields optional-with-gate contract
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** User decision (Th, 2026-07-03): "imagery must not block section rendering"; W7a gate applied (Tm: "UPDATE 1, gated t/t"); the rule landed in the creator prompt (4a) and 003 6b; `illustration_url` was already required:false + skip_field in schema.
- **what:** Any `site_assets.*`-sourced component field MUST be `required: false` with `on_missing: skip_field`, and its markup MUST be gated with a template conditional (`{{if .illustration_url}}` around brief-explanation's image wrapper is the model — Go templates treat "" and missing as false, covering the src="" broken-image case). Imagery arrives asynchronously and must never block or defer a section; the section renders imageless and the image is added by the pipeline's own queued rebuild. Codified as 003 item 6b and the creator prompt's IMAGE FIELDS RULE.
- **sources:** w7a_01_gate.sql; slice4b_003_contract.md#Edit-1 (6b); slice4a_creator_prompt.sql (R4); running_notes_scheme_to_components(55).md#Th #Tl #Tm
- **relations:** plan_sections field deferral semantics; section-scope imagery pipeline; component-creator prompt re-aim.
- **verify-later:** brief-explanation html_template gate; input_schema on_missing values across image-consuming components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Component schema-template consistency invariant
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "The governing invariant — a component's schema `items` must match its template tokens — is the right thing to hold; the reconciler enforces consistency toward the current schema."
- **what:** A component's `input_schema` (shape `{"fields": {...}}`, unmarshalled in component_library.go) is the contract for its `html_template` tokens: array item field names in the schema must match what the template reads. The reconciler, the prompt, and generation all derive from input_schema, so divergence breaks all three coherently. Known violation: info-card-grid's stored html_template literally contains `<no value>` (rendered-against-nil output apparently written back into the template column) — flagged as its own repair thread and never fixed inside this unit. services-grid shares differentiators' schema byte-identically and was healed by the same fix.
- **sources:** running_notes_checkpoint_uu.md#Confirmed-during-the-hardening-review; running_notes_checkpoint_ss(1).md#Root-cause-in-code; RUNBOOK_pcw_item_fields_fix.md#Follow-on
- **relations:** array item-fields contract; render-time reconciler.
- **verify-later:** info-card-grid html_template (still `<no value>`?); services-grid first-use spot check.

<!-- SOURCE: U04_idea_uk.md -->
### CSS colour inheritance / --section-* luminance model (inline style = override, not main CSS)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Documented in 003 ("CSS Colour Inheritance Model") and restated as renderer contract in the tool-portal-light css_template header; the dark components *bypassing* it is the observed live behaviour.
- **what:** The platform's component colour contract as evidenced here: styles.css sets body/heading/element rules through `var(--section-*, var(--color-*))`; the renderer appends `--section-*` defaults after rendering based on palette/background luminance; a dark callout section overrides `--section-*` on its own container (sanctioned); layouts MUST NOT declare `--section-*` defaults; renderer-managed surface classes must be surface-coloured; a component's inline `<style>` is an **optional override, not its main CSS** (user correction, checkpoint mmm). The scheme gap exists precisely because dark components violate this — hardcoding backgrounds and `--section-*` inline. Two parallel styling systems (layout class vocabulary vs component class vocabulary) is the structural tension to resolve (Q4).
- **sources:** idea.uk/running_notes_2(6).md (lll/mmm); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/migration_layouts_scheme_and_light_tool_portal.sql (renderer contract comment); idea.uk/001_component_flow.md
- **relations:** scheme-as-override thesis; styling-render-pipeline (036); 003 contracts doc.
- **verify-later:** the luminance-appender code location (report investigation G names finding it).

<!-- SOURCE: U09_adoption.md -->
### Numbered-flat-fields anti-pattern (25 components)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** "Twenty-five active components match the numbered-field signature" (May 2026 audit); game-list rewritten Tier-D 2026-06-04; the rest (nav, card/grid, tier/stat, tool-internal clusters) unmigrated.
- **what:** Schemas declaring `post1_title…post6_*` with `source: llm` force the LLM to fabricate list items by structure (invented games, duplicate Jelly Invaders, links to nonexistent URLs) — no prompt rule can save a schema that demands invention. Groups: 8 navigation components (need a curated `nav.*` source, not query-resolvable), 7 card/grids (straight items-array rewrites), 5 tier/stat (may fit `site_specs.<aspect>.items`), 5 tool-internal field clusters (heterogeneous, case-by-case). Component-creator must refuse the shape for new "list of N things" components.
- **sources:** FOCUS_component_schema_patterns.md, migration_game_list_tier_d.sql header, CATALOGUE(9)#family-c
- **relations:** Tier-D contract; curated-list source vocabulary decision (deferred)
- **verify-later:** re-run the `<prefix>1_/2_/3_` audit; component-creator prompt Tier-D block

<!-- SOURCE: U09_adoption.md -->
### Anti-fabrication content path (Step 2/3: llm_field_specs, targeted prompt, merge_with)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Step 2 status: PASS" (2026-05-12 observations); Step 3 deployed with verification surface; tool-list rendering real pages confirmed on later runs.
- **what:** Closed the gap where `page-content-writer` fabricated items despite queryresolve existing: plan_sections resolves query.* before the LLM call and carries a full per-section `Component` + `resolved_data` + `llm_field_specs` (built from `llm_guidance`) on `section_plan.sections_ready`; the writer's prompt was rewritten to ask only for the `source: llm` fields (anti-"pink-elephant": it does not enumerate forbidden fields); `RenderComponentAction` honours `merge_with: current_section.resolved_data`, overlaying resolved data as authoritative over LLM output. Shared `loadSectionComponents` helper extracted.
- **sources:** FOCUS_directory_builder_and_list_components.md#implementation-history, old2/STEP2_changelog.md, old2/STEP3_changelog.md, old2/step3b_prompt_template.txt
- **relations:** Tier-D contract; page-content-writer workflow (spawn_research → load_site_specs → prepare_link_context → build_render_context → loop → compile_page)
- **verify-later:** plan_sections_action.go llmFieldSpec; page-content-writer def process_sections_loop iterate_over; RenderComponentAction merge_with

<!-- SOURCE: U09_adoption.md -->
### plan_sections on_missing semantics and the cta_url required-field deferral
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "guide-list empty on guides hub AND root index — RESOLVED (Part 14p)… set cta_url.required=false on guide-list_pre_037 and blog-listing_pre_037 (APPLIED)."
- **what:** plan_sections' required-field switch handles use_fallback/skip_section/needs_human_review/block but has no `skip_field` case, and `on_missing` defaults to skip_field — so a required field with an unpopulated site_specs source fell to default-defer and held the entire section (hero-only hubs). A query-sourced list field can never defer (resolver returns a non-nil empty slice). Deliberate decision: fix the component (required=false), not the engine — the defer-for-safety default on required fields is defensible. Related follow-ups: cta_url fallbacks (`site_specs.identity.*_index_url` unpopulated → `href=""`; game-list gained a `/games/index.html` fallback, tool-list still lacks one), and inconsistent cta source vocabulary across sibling list components.
- **sources:** running_notes_14(25)#part-14p, HANDOFF_2026-06-06#resolved, NEXT_CHAT_INPUTS_2026-06-06 §4
- **relations:** needs_section_data; silent-fallback link family; Tier-D contract
- **verify-later:** plan_sections_action.go required-field switch (~line 44340 of dump); content_components cta_url required flags

<!-- SOURCE: U09_adoption.md -->
### Chrome templates must be variable-driven (pre-store hardcoded-link gate)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** "Architecture decided, implementation pending" (FOCUS_chrome_templates, 2026-05-06); zero of the active footer templates consume `{{.nav_items_html}}` today.
- **what:** Header/footer components are LLM-generated with hardcoded `<li>` links, freezing nav at generation time and bypassing populate_nav_tables/classifyPagesForNav dedup entirely (doc 003's explicit rule violated). Proposed structural enforcement: a pre-store validation gate in store_generated_component_action.go rejecting chrome templates with hardcoded internal links outside {{range}}/{{if}} blocks, plus prompt teaching and a chrome-template-repair migration. Companion: `buildServicesHTML` is a parallel, dedup-less nav query ("Tools, Guides, Games, Games, Tools" verbatim) — drop it and its `services_html` context field once templates use quick_links_html. Principle: "as algorithmic as possible and enforced", not prompt-led.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-1, old2/HANDOFF_2026-05-07(1)#8
- **relations:** nav dedup guard B-029-1; render-context variables table (nav_items_html, quick_links_html, categories, footerLinks…)
- **verify-later:** store_generated_component gates; content_components chrome templates for {{.nav_items_html}} usage; buildServicesHTML existence

<!-- SOURCE: U09_adoption.md -->
### Thin-slice constitution (always-on rules; future standards rows)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "The always-on rules for any task on this codebase. Included in full in every bundle… Later it becomes the standards rows with scope = constitution" (thin_slice_constitution.md).
- **what:** The flat-file constitution: reuse before recreate; fix structural problems not symptoms; every agent is an orchestrator; reply on the caller's responses topic; workflows thin/complexity in Go; no SQL subworkflows — spawn sub-agents; check schema before SQL; parameterised queries only; snake_case for identifier-shaped values vs kebab-case for data-shaped values; text+CHECK not native enums; soft-delete via deleted_at; no logger.Debug; log orchestration/correlation ids and inter-agent messages; deployment path git→Actions→B2; plain pragmatic tone. Task-specific 003 contracts are pulled in only when touched. Reinforced session rules: snapshot before any DB change (restorable bak tables, in-txn, NAMEDATALEN 63), resolve site_id fresh via domain, don't conclude from partial signals.
- **sources:** docubundle/thin_slice_constitution.md, HANDOFF_2026-06-09#standing-rules, running_notes_14(25)#note-db-change-snapshot-standard
- **relations:** 003 contracts; standards table (aspirational storage form); council-agent charters
- **verify-later:** standards table existence with scope=constitution (expected absent)

<!-- SOURCE: U10_imagery.md -->
### Dispatch input contract for handler agents
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Recorded as a hard-won mechanism 2026-07-12: handlers invoked by build-dispatch-loop receive `input_data.spec`, `input_data.site_id`, `input_data.domain`, `input_data.item_type`.
- **what:** The canonical payload shape a dispatched handler sees; step conditions and input paths must be written against it (e.g. asset-deployer's check_mode tests both `input_data.spec.mode` and `input_data.mode` to cover dispatch and direct-call shapes). Divergence between dispatch-shape (`input_data.spec.*`) and direct-call shape (`input_data.*`) is a live source of latent extraction bugs.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, SQL_2026-07-11_asset_deployer_brand_head_mode.sql, SQL_2026-07-12_asset_deployer_explicit_paths.sql (NOTE block)
- **relations:** ExtractActionInputs lesson; work-item dispatch semantics.
- **verify-later:** build-dispatch-loop spawn payload construction.

<!-- SOURCE: U12_docs024_archives.md -->
### CSS section-colour model: inheritance → hardcoded dark-section variables → renderer-computed defaults → token-referencing painting sections
*(merged from 4 independent findings)*
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Four independent archive lines converge on the same evolution. Earliest baseline (`old/older1/003_contracts_and_standards.md`, `old_design_and_styling/FOCUS_design_and_styling.md`): plain CSS inheritance, dark-background components just set literal `color:#fff`/`inherit`, no `--section-*` variables exist. Middle era (`old/older1/003_contracts_and_standards_v2..v7.md`, `debugging_old/003_contracts_and_standards_v8..v11.md`, `archive_april_26/003e_contracts_and_standards_v5.md`): a `--section-*` custom-property contract keyed off a boolean `is_dark_section` column, with LITERAL hardcoded values (`--section-heading:#ffffff`, `rgba(255,255,255,0.9)`) enforced by `ValidateDarkSectionContract()`. An intermediate renderer change (`old_design_and_styling/PHASE_4_4_cleanup_summary.md`): the renderer's `buildSectionDefaults` began computing these values automatically from palette luminance (WCAG-based), removing the manual per-component declaration burden. Live (`003_contracts_and_standards(8).md`): the "Section painting contract" — `is_dark_section` is demoted to inert catalogue metadata ("MUST NOT key styling"), and any section that paints its own background must instead RE-EXPORT `--section-*` as references to theme tokens via one of four models (pair band, palette band, image/ink-derived, ambient/no-background) using `color-mix()`, so colours flip automatically with the site's scheme; literal colours are forbidden and mechanically enforced by `fix_forced_text_colors`.
- **what:** Documents the multi-year hardening of how section backgrounds get correctly-coloured text: from ad hoc inline colours, to a hardcoded-value contract gated on a boolean flag (which locked every dark section into literal white-on-dark), through a renderer-side automation step, to the current token-referencing "painting" model that treats `is_dark_section` as inert metadata and derives colours mechanically from the active palette.
- **sources:** old/older1/003_contracts_and_standards.md; old/older1/003_contracts_and_standards_v7.md#"Section Context Variable Contract (Dark Sections)"; debugging_old/003_contracts_and_standards_v11.md#"Section Context Variable Contract (Dark Sections)"; archive_april_26/003e_contracts_and_standards_v5.md#"Section Context Variable Contract"; old_design_and_styling/FOCUS_design_and_styling.md#"CSS Colour Inheritance Model"; old_design_and_styling/PHASE_4_4_cleanup_summary.md#"Phase 4.5"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Section Context Variable Contract (Painting Sections)"
- **relations:** styling-render-pipeline (036); design-composition (site-design-planner palette resolution feeds these tokens); fix_forced_text_colors action
- **verify-later:** grep deployed component templates for literal `#ffffff`/`rgba(255,255,255,` inside `--section-*` declarations to confirm the old hardcoded pattern is gone; inspect `fix_forced_text_colors` and `buildSectionDefaults` Go source.

<!-- SOURCE: U12_docs024_archives.md -->
### Component Quality Contract (scoring formula, quality columns)
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** Introduced in `old/older1/003_contracts_and_standards_v6.md` ("## Component Quality Contract"); absent from v7 onward and absent from live `003_contracts_and_standards(8).md` (no "Component Quality Contract" heading), though `quality_score`/`quality_issues` fields still appear inline in the live doc's "Component Creation & Regeneration Contract" JSON examples.
- **what:** v6 fully specified a quality-tracking contract for `content_components`: eight quality columns (`template_variable_count`, `schema_field_count`, `template_closed`, `schema_template_synced`, `has_data_component`, `quality_score` 0-100, `quality_checked_at`, `quality_issues`), a scoring formula starting at 100 with fixed deductions per violation, three computation triggers (on-insert, periodic audit by `component-quality-auditor`, targeted rescan), an automatic `needs_component_regeneration` work item below score 50, and planner preference for higher-scored components. This entire standalone contract section vanished between v6 and v7 and was never restored, even though the live system-architecture doc still lists a `component-quality-auditor` agent and the live contracts doc still surfaces `quality_score`/`quality_issues` as return-payload fields — suggesting the mechanism may partly persist in code/DB while its dedicated documentation disappeared.
- **sources:** old/older1/003_contracts_and_standards_v6.md#"Component Quality Contract"; docs024_key_docs_latest/003_contracts_and_standards(8).md (residual field mentions); docs024_key_docs_latest/002_system_architecture(4).md (component-quality-auditor agent row)
- **relations:** StoreGeneratedComponentAction / component regeneration contract; component-quality-auditor agent
- **verify-later:** check whether `compute_component_quality`/`ScoreAndPersistComponent` and the `content_components` quality columns still exist and are actively populated.

<!-- SOURCE: U12_docs024_archives.md -->
### `query.{name}` field-source resolution timing (render-time → plan-time)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Archive v7 "Source prefixes" table: `query.{name}` resolved "At render time." Live table: resolved "At `plan_sections` time."
- **what:** In the Component Input Schema v2 Contract's source-prefix table, the `query.{name}` prefix (used for blog posts/categories lists) moved from being resolved at page-render time to being resolved earlier, during `plan_sections`, with the result projected into the field's declared shape. Consistent with the broader shift toward front-loading data-availability checks (the "Layer 0" pre-generation triage) rather than discovering missing/stale query data only at render.
- **sources:** old/older1/003_contracts_and_standards_v7.md#"Source prefixes"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Source prefixes"
- **relations:** Component Input Schema v2 Contract; plan_sections / Layer 0 pre-generation data triage; page_rerender item_type
- **verify-later:** check plan_sections Go action for query-prefix handling, confirm it projects results to field shape at plan time.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component JS/CSS contract (contract 003): no inline script, split js_content vs js_snippets
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Migration B implements all three for the news components" — inline IIFE extracted to js_content, html_template gets `<script src>`, shared formatNewsDate moved to js_snippets
- **what:** Doc 003's contract: no inline `<script>` blocks in `html_template`; component-specific JS lives in `content_components.js_content` served at `/tools/assets/{function}.js`; shared utilities used by multiple components live in `js_snippets`. Also covers component CSS scoping (no bare element rules, dark-section CSS vars with fallbacks) and idempotent migrations.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#003-Contracts-and-Standards, js_snippets_news_gaswholesalers/old/guidelines_compliance_check(1).md
- **relations:** js_snippets site-level rendering pipeline; files_field deploy bug
- **verify-later:** grep/inspect `<script>`; `html_template`; `content_components.js_content`

<!-- SOURCE: U15_docs019_running_notes.md -->
### JS content separation contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** "js_content RESOLVED — the assets-split EXISTS (003 'JS Content Separation Contract'...)" (principles(59), 2026-06-11).
- **what:** For interactive components, `store_generated_component_action.go`'s `separateInlineJS()` extracts inline `<script>` bodies into a separate `content_components.js_content` column, replacing them in `html_template` with a `<script src="/tools/assets/{function}.js">` reference; `RerenderSinglePageAction.collectJSAssets()` assembles the resulting multi-file git commit. Verified NOT used by the library-tool pipeline (tool-generator/improver mandate one inline script; the fork INSERT omits `js_content`), creating a landmine where a library tool adopting the split would fork with a dangling script reference and a 404'ing asset.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 "js_content RESOLVED" and "VERIFIED... 019's 'isn't built yet'" entries.
- **relations:** Tool-doc header system; fork-divergence detection.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Result-contract dead-key class and Option A unification
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** NNN_fix_diagnose_complete_output_fields (run 17933a83, Kafka Message Size Too Large at 1.27MB); HANDOFF_builder_thread "Option A CLOSED: shared result contract (datahelpers/result_contract.go), response size guard, four agents on preferred result_from; deployed v1.0.1092+".
- **what:** CompleteWorkflowAction honoured output_field/output_fields and otherwise shipped the ENTIRE collected_data; `result_from` was a key the action never read, so diagnose completions always shipped everything (masked until the 515-file analysis blew the Kafka cap). Same class: the orchestrator pointed output_fields at an imagined key ("diagnose-agent_result") when the engine stores a call step's response under the STEP NAME. Fixes: point at real keys; then Option A — a shared result contract with both readers honouring result_from/output_fields plus a response size guard; NNN_rename_complete_keys_preferred moves the four diagnose/index agents to the preferred key once that image is deployed. Standing rule made mechanical: keep workflow variable names in sync with what actions read.
- **sources:** NNN_fix_diagnose_complete_output_fields.sql; NNN_fix_orchestrator_complete_key.sql; NNN_rename_complete_keys_preferred.sql; HANDOFF_builder_thread.md#2
- **relations:** loop-back plumbing fault class; bounded bundle egress (persist-and-reference)
- **verify-later:** datahelpers/result_contract.go; extractFinalResult size guard; census query for deprecated keys

<!-- SOURCE: U16_docs019_design_plans.md -->
### Adapter response envelope contract (single-sourced)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration(19) 2026-06-10/11: "Envelope contract resolved empirically … The contract now lives once in 035_adapter_guide.md §1 (FOCUS_adapter_design fully merged … and retired)".
- **what:** Resolved from code, not docs: the coordinator claims awaited requests on in_response_to_request_id first (request_id fallback); working adapters use typed body headers with real booleans + ProduceWithValidation; every adapter reads `action` from the BODY; a reply without the right Kafka headers silently falls through to process-as-work and times out (the documented thunder fault — found and fixed in the analyser adapter before deploy). Import-reuse verdict: reuse canonical types for the body, add a local Kafka-header builder (canonical ResponseHeaders lacks request_id/message_id/ToKafkaHeaders). A 003-vs-FOCUS documentation contradiction was settled empirically and single-sourced.
- **sources:** PLAN_workflows_and_actions_migration(19).md (guideline audit + dispatcher fix + envelope resolution entries)
- **relations:** analyser adapter; doc-drift (docs contradicted; code decided); 035_adapter_guide (canonical home, another unit)
- **verify-later:** 035_adapter_guide.md §1; whether 003 §832-890 was replaced with the pointer (was PENDING)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Chassis conventions verified: text+CHECK, previous_version_id, deleted_at
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS_schema_verification_findings §1 — read off the live schemas; applied as corrections across the contract set.
- **what:** The live chassis conventions every new table must follow: enumerated values are text with CHECK constraints (never native enums); versioning is version integer + previous_version_id uuid self-FK with unique (type,version); soft delete is deleted_at (never a status=archived); timestamptz defaults now(); jsonb for flexible payloads. The verification pass also corrected wrong reuse assumptions in the contracts ("the contracts are corrected to match reality, not the reverse") and confirmed real fields: approval_mode, pipeline, item_key, briefing_questionnaire, input/output_contract, agent_category CHECK set.
- **sources:** FOCUS_schema_verification_findings.md; PLAN_active_config_schema(3).md#2 note
- **relations:** schema-before-SQL discipline; all six contracts
- **verify-later:** n/a (verified from live schema dumps)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Priority profile (order not weights; sealed constraints)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §9 "Representation: an order, not numeric weights … A node stores only its differences from what it inherits … computed on demand"
- **what:** Requirement-relative priority among dimensions (security/speed/simplicity/generality/functionality/cost) lives on the objective-tree node as an *order* (with sealed/constraint flags), stored as differences-from-inherited and computed on read. Sealed constraints are ancestor-wins legal floors; a change triggers targeted re-validation of descendants holding conflicting overrides. The open crux (§9.7) is choosing the entry node.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#9, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.7
- **relations:** why-chain; mediator; drift detection; direction-of-travel
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Atomic standard (generated-views doc tree)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §2 "Optimal unit: the atomic standard, not the document … Documents are generated views over the atoms"
- **what:** The smallest addressable unit is one rule-atom with structured frontmatter (id, concern, scope, applies_to, kind, severity, status, version, supersedes, owner, check, related) and a body split into rule/rationale/examples. Constitution, per-concern handbooks, change-type bundles, and the machine manifest are all *generated views* over one source, so nothing drifts between a doc copy and an agent copy.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#2, ED/FOCUS_best_practice_doc_tree(1).md#4, ED/FOCUS_best_practice_doc_tree(1).md#5
- **relations:** mediator routing model (the atom fields are the routing table); doc-tree adoption; concern curators
- **verify-later:** proposed `standards` table

<!-- SOURCE: U18_sql_for_agents.md -->
### Input/output contracts on agent definitions
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Contract UPDATE statements across 011/022/024/025/029 etc.; 129 extends contracts as live behaviour ("input_mapping must satisfy the input_contract — 016b §9 spawn+call rule").
- **what:** Every agent row carries `input_contract` (required/optional fields) and `output_contract` (produces). Contracts are both documentation and runtime validation hooks; the 2026-07 diagnosis work established the durable rule that an input the workflow reads must be declared in the contract (137's "spec is UNDECLARED" fix) and that call-site input_mapping must satisfy the callee's contract.
- **sources:** 011_site_deployer.sql; 022_site_planner.sql; 129_wire_diagnosis_subject_threading.sql; 137_recreation_spec_and_note_subject.sql; sql_for_agents_v1/009_all
- **relations:** remove-loops plan; workflow_contract_chain view (v1/010)
- **verify-later:** chassis contract-validation code path; how strictly contracts fail fast in production

<!-- SOURCE: U18_sql_for_agents.md -->
### Call metadata vs response-data convention (output_field.response)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** v2/001_general_rule.md states it as "The general rule going forward"; 116 confirms it verified in coordinator.go ("a step result is stored under BOTH the step name AND its output_field, adapter body under .response").
- **what:** Workflow data-shape convention: when a step calls another agent, call metadata (agent_id, request_id, topics) is stored directly at the step's output_field while the called agent's response payload lands at `output_field.response`. Many prompt-template and field-path bugs in this directory trace to violating this shape.
- **sources:** sql_for_agents_v2/001_general_rule.md; 116_thunder_training_monitor_worker.sql; 003_site_classifier.sql (classification.response.result paths)
- **relations:** template field-path rules (134); input_mapping
- **verify-later:** coordinator.go result-storage code (~L1636/L2408 per 116)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component render modes (template | agent | composite | standalone)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Columns and comments exist (004 PART 2), but the 000 backup shows all 41 components render_mode='template'; only 'standalone' (tools) is additionally observed in seeds.
- **what:** render_mode declares how a component is produced: 'template' (direct substitution), 'agent' (spawn agent_type with optional agent_workflow, data pulled via data_sources dot-paths), 'composite' (child_components list), and later 'standalone' for tools (html_template IS the final output). The agent/composite modes appear designed but unexercised.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART2; docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component library; standalone tool render.
- **verify-later:** Go render path switch on render_mode; any rows with render_mode in ('agent','composite').

<!-- SOURCE: U19_sql_tables_components.md -->
### Kebab-case naming contract (component function + pages.page_type)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** chk_function_kebab_case CHECK in the live dump; migration 051 adds chk_page_type_kebab_case with pre/post distribution audit; data-component attributes reconciled to function.
- **what:** Identifier-shaped values are kebab-case, enforced by CHECK constraints: content_components.function (regex `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`, empty allowed for legacy), pages.page_type (same regex, snake rows migrated: blog_post→blog-post etc.). data-component attributes in templates must equal function; a partial unique index enforces one active component per function. Also separates page NAME 'index' from page TYPE 'landing'.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-naming-standardisation; docs/agent_docs/sql_for_tables/003_pages.sql#051_pages_page_type_kebab; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql
- **relations:** contracts doc 003/042 naming contract; query-resolver list components (rely on page_type values).
- **verify-later:** pg_constraint rows chk_function_kebab_case, chk_page_type_kebab_case; idx_cc_tool_function_unique.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tier D items-array component schema shape
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration 041 hand-writes tool-list to Tier D; 042 queues guide-list regeneration through the pre-store validator; game-list rewrite mirrors tool-list, "field vocabulary IDENTICAL to tool-list".
- **what:** List components must declare a single `items` array with a sub-schema (title, url, meta_description, nav_label) plus top-level fields (eyebrow_label, section_heading, section_intro, cta_url, cta_label, card_link_label), replacing the legacy numbered-flat anti-pattern (guide_1_url…guide_6_url) that broke sites with fewer items. A pre-store validator enforces the structural contract on LLM-regenerated components; rejections land in agent_error_log.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-041 and #migration-042 and #game-list-rewrite
- **relations:** query-resolver list components; component naming contract; agent_error_log.
- **verify-later:** pre-store validator code; tool-list/guide-list/game-list current schemas.

<!-- SOURCE: U19_sql_tables_components.md -->
### Template syntax unification and three-way field alignment
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Long sequence of applied fixes in 005: Handlebars {{#each}}/{{#if}} → Go {{range}}/{{if}}, missing-dot placeholders ({{logo_text}} → {{.logo_text}}), and the "<no value>" root-cause fix aligning LLM prompt output / template fields / input_schema.
- **what:** Templates are Go text/template; a large family of patches converted early Handlebars-style seeds and fixed the recurring three-way mismatch where the LLM prompt, the template field names, and the input_schema disagreed (headline vs title vs section_title; features[].name vs services[].title). Render-context vocabulary standardised (nav_items_html, services_html, footer_nav_html, cta_text, logo_text, company_name).
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#templating-fixes and #fix-no-value; docs/agent_docs/sql_for_tables/012_site_components.sql#replace_template_var
- **relations:** component library; component-based headers; Tier D shape (later formalisation).
- **verify-later:** remaining Handlebars syntax in content_components; render context builder in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Query-resolver list components (pages_where_type) and canonical section URLs
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** gamesdesign migrations re-type guide pages to page_type='guide' so guide-list (items.source = query.pages_where_type:guide) resolves them, "mirrors the working game-list / page_type=game precedent"; URL migration to /guides/<slug>/index.html.
- **what:** List components resolve their items dynamically from the pages table by page_type via a query resolver — no template change needed when pages are added. Depends on canonical page typing and the canonical nested URL shape /<section>/<slug>/index.html produced by CanonicalisePage, making tools/games/guides structural peers.
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql#migration_retype_guides_to_guide and #migration_guides_url_to_canonical
- **relations:** kebab naming contract; Tier D shape; site-plan page roles.
- **verify-later:** queryresolve Go code; link_registry sync after URL moves.

<!-- SOURCE: U21_legacy_docs_b.md -->
### data-function contract + intelligent component fallback (P1/P2/P3)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** docs009/001: "A data-function attribute in the HTML acts as a 'shared contract'... P1 perfect match, P2 good match, P3 generic-text-block — the site always gets built"; superseded by the function/kebab-case + data-component contract (docs017/042) whose GetComponentWithFallback keeps the 3-step fallback.
- **what:** The original decoupling of structure from content: the architect assembles empty containers tagged by function (data-function="problem_statement") so the content pipeline can independently fill them; component lookup degrades gracefully (exact function → similar purpose → generic-text-block) so a build never fails for lack of a component.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-1; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#Lookup-Safety-Net
- **relations:** component naming contract (successor); content_components.function; AssembleFromLibraryAction.
- **verify-later:** GetComponentWithFallback in component_library.go; generic-text-block component row.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Recursive component tree ("everything is a component")
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** docs009/001: "We remove is_container... If the HTML template contains {{.Slot_main}}, it IS a container"; RenderNode recursive algorithm, ghost components (wrapper_tag NULL), slot merging; the shipped system instead uses a flat section list per page with header/footer injection.
- **what:** A radically simplified component model where structure is defined entirely by template placeholders: components declare defined_slots and data_schema; the build plan is itself a component tree the architect walks recursively (RenderNode), handling any nesting depth; themes are just root components; "ghost" components (no wrapper tag) reduce div nesting. Content generation is decoupled by flattening the tree into a content_map of UUID→field requirements. The flat-sections production system never adopted the recursion, though slots re-surface in docs018's slot-based assembly proposal.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Simplification; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#2-The-Recursion-Logic; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Key-Architectural-Principles
- **relations:** slot-based modular assembly (docs018/007); asset bubble-up; content injector pattern.
- **verify-later:** defined_slots column on content_components (expected absent or unused).

<!-- SOURCE: U21_legacy_docs_b.md -->
### render_mode decision matrix (DB template vs agent vs research)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs012/012 decision matrix ("Headers/Footers: template, never generated; FAQ: agent + research; Pricing/Contact: template + brief data") and "The render_mode field on content_components is the key differentiator".
- **what:** Each component declares how its content is produced: render_mode='template' renders directly from brief/render_context data (LLM only fills missing schema fields); render_mode='agent' spawns LLM generation, optionally preceded by the research-agent when needs_research=true. Pure-structure components never touch an LLM; research-backed components always cite. Governs the per-section branch inside the page build loop.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-1; docs012_site_maps_and_components/010_component_and_site_architecture.md
- **relations:** research agent; page-content-writer section loop; "LLM = Agent" principle (every LLM call gets its own agent with research/draft/review).
- **verify-later:** render_mode/needs_research/agent_type/data_sources columns on content_components.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent input/output contracts (expects/required/produces)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** docs012/012 Part 4 contract JSON for site-planner/page-content-writer/research-agent; "input_contract/output_contract fields in agent_definitions" cited as the tracking mechanism; docs017/002_standardising exports them for validation.
- **what:** Formal per-agent declarations of expected input fields (with required subset) and produced output shapes, stored on agent_definitions, intended to make cross-agent data flow checkable (workflow validator) and self-documenting. Partially realized; the enforced end of it became ActionInputSpec at the action level.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-4; docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md
- **relations:** ActionInputSpec; workflow builder/validator; contracts-and-standards doc 003 (current descendant).
- **verify-later:** input_contract/output_contract columns populated?

<!-- SOURCE: U21_legacy_docs_b.md -->
### Dark-section context variable contract (--section-*)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/043 lists 10 dark components migrated with is_dark_section=true and --section-* vars; four enforcement layers specified with deployment order ("Run 014_section_context_migration.sql...").
- **what:** Any dark-background component must set --section-text/-text-muted/-heading/-surface/-border custom properties on its container; global CSS reads them with light-theme fallbacks. Enforced in depth: DB flag (is_dark_section) + audit queries, Go warnings in RenderComponentAction/SavePageSectionsAction, LLM prompt rules in webdesign-agent and page-content-writer, and periodic SQL audits. Direct ancestor of the current section-contrast model.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md
- **relations:** colour inheritance model; component naming contract (companion doc); maintenance audits.
- **verify-later:** is_dark_section column; validate_dark_section.go; current section-contrast implementation.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Component naming contract (function = canonical kebab-case ID)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/042: "content_components.function is the canonical identifier... DB constraint chk_function_kebab_case... partial unique index on active function"; migration table of renames (social_proof → social-proof, 5 heroes disambiguated).
- **what:** One rule ending a class of chain-breaking bugs: `function` (kebab-case, regex-constrained, unique among active components) identifies a component everywhere — the template's data-component attribute must equal it, page_components.slot_name stores it, planners assign by it, rerenders match by it. NormalizeComponentFunction + 3-step fallback tolerate legacy data; adoption pipelines must translate external names, never import them.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md
- **relations:** data-function contract (ancestor); page_components; adoption pipeline mapping; SavePageSections/page-rerender.
- **verify-later:** chk_function_kebab_case constraint and unique index in DB; component_validation.go.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component field-source tiers (A/B/C/D + renderer) and proposed Tier E runtime-feed
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Tiers A–D + renderer confirmed from the component-creator prompt at v1.0.1080 (deployed); Tier E is "proposed, pending decision" (2026-06-29) and gap 2 of the autonomy plan (not built).
- **what:** component-creator's schema contract sources each field from Tier A (voice/llm), B (tunable labels/static+fallback), C (site data — site_specs/site_assets), D (derived lists, query.* resolved at plan time), plus a "renderer" source (JS-filled single value with fallback). There is NO tier for content fetched client-side from a JSON feed at runtime — so regenerating a daily-feed component as-is would wrongly bake a build-time provocation into the template. Proposed Tier E ("feed.{name}"): emit a stable-selector DOM shell + declared DOM contract + (originally) an inline loader following a canonical pattern; the archive build refined this to marker-at-generation + external loader.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 (GAP 2 CONFIRMED); docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNBOOK_phase2_provocation_js(29).md#gap-2
- **relations:** section descriptor; generation-time guards; loader-builder agent. Dropped earlier framing: Gap-2 sub-options "(a) component-creator emits a companion js_snippet" vs "(b) loader snippets are library fixtures" (early runbook versions) — superseded by the Tier-E + loader-builder design.
- **verify-later:** component-creator prompt_template tier section; any feed.* source in input_schemas

<!-- SOURCE: U23_docs_root_vonc.md -->
### Generation-time guards for dynamic components
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 084 result 2026-07-06: "FIRST LIVE VALIDATION of baking the guards in at generation" — has_marker=t, has_inline_script=f on the created component; guards held through the real pipeline end to end.
- **what:** Lessons from the whole thread baked into component GENERATION instead of post-hoc surgery: emit `data-runtime-fill` in the template's section tag at generation (no string-REPLACE marker step); forbid inline `<script>` entirely in dynamic components (extraction-bug class becomes impossible; behaviour lives in the external loader); make header copy llm-sourced (no deferral risk); list entries pure markup (nothing for the resolver to fail on); hidden clone-template item plus a `[data-…-template]{display:none}` author rule (hidden alone loses to author CSS); visible empty state. Declared "the pattern for all future dynamic sections".
- **sources:** docs/SPEC_provocations-archive-list.md#design-decisions; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-084-succeeded; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** Tier E proposal; marker anchoring lesson; hidden-vs-author-CSS lesson; store-path validation
- **verify-later:** component-creator output for any newer dynamic component (marker present at generation?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### CSS variable naming convention (--color-*) + creator prompt STRICT RULE
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Hero template fixed and verified (magenta CTA, dark bg) 2026-06-24/25; component-creator prompt patched with "USE ONLY THESE NAMES" + STRICT RULE, UPDATE confirmed in DB 2026-06-24 ~16:50; library-wide audit complete.
- **what:** System CSS custom properties follow `--color-primary/-secondary/-accent/-background/...` naming; LLM-generated components had emitted `--primary-color`-style names that don't exist in styles.css, so fallback hexes fired (the "brochure-blue" index). Fix: template REPLACE on hero + a component-creator prompt section explicitly prohibiting the wrong names and separating Palette from Layout tokens. Documented exception: `--archetype-color` is intentional per-card tinting with `--color-accent` fallback. All new components inherit the correct names.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:30 + #2026-06-24-~16:50; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** post-025 CSS theme flow; legacy-variable "fossilised render" tell (016b chrome entry)
- **verify-later:** component-creator prompt_template section 7; grep new templates for --primary-color

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sanctioned content-edit paths (content_data is truth; HTML patching rejected)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Doc 003 quoted verbatim in the HANDOFF: "this is why HTML patching was rejected as an edit mechanism"; fix_component_template's remove_element fix type and its header deferral to the section-editor confirmed from file headers; section_editor_actions.go confirmed to exist as Go code.
- **what:** `content_data` is the source of truth for section content; patching `page_components.rendered_html` is a bridge at best (lost on the next re-render) and was explicitly rejected as an edit mechanism. Template changes have designated routes: `fix_component_template_action` fix types (including `remove_element` — "removes HTML elements matching a pattern"), with page-component content changes deferred to the section-editor workflow. The mini-lobby trim's method question — which action edits a template, which re-render propagates it, what a NULL-content_data section does — was deliberately settled by bundle verdict rather than guessed. Fallback when no supported path exists: full-text template UPDATE (never multi-line REPLACE of nested markup), verified by length delta, propagated by a page_rerender item.
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/PLAN_provocation-card(3).md#method-corrected; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-primed; docs/bundle_minilobby_trim(4).sh (header)
- **relations:** two re-render paths; per-tool docs (method correction recorded); section-editor
- **verify-later:** fix_component_template_action.go fix types; section_editor_actions.go component_swap

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### API documentation convention (OpenAPI/Swagger + internal API.md)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Archived docs/_archive/api/API_DOCUMENTATION.md is byte-identical to live docs/API_DOCUMENTATION.md; describes `make swagger`, `make validate-openapi`, per-service `API.md` files, `*_swagger.go` annotations.
- **what:** Two-tier API documentation standard: external customer-facing OpenAPI 3.0 spec at `internal/auth-service/api/openapi.yaml` with swaggo annotations, and internal per-service `API.md` files documenting Kafka topics, DB schemas, env vars. Includes a CI workflow that lints the spec and fails on uncommitted swagger regen.
- **sources:** docs/_archive/api/API_DOCUMENTATION.md#external-api-documentation, #internal-api-documentation
- **relations:** superset context for the public/admin API plans (007b/008b); live counterpart docs/API_DOCUMENTATION.md + docs/api/reference.html
- **verify-later:** internal/auth-service/api/openapi.yaml; Makefile targets swagger/validate-openapi

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Training-data export format (ChatML + metadata sidecar; SFT/DPO negative examples)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22"; iter_0 audit "1,970 total" rows
- **what:** Convention that each training row is a ChatML `messages` array plus an ignored metadata sidecar (source_log_id, agent_type, step_name, orchestration_id, model, export_version). Code fences must be stripped; edge-case "prose instead of JSON" rows are excluded from SFT (they'd teach wrong-shape output) but kept in `llm_call_log` as future DPO "rejected" examples. Schema heterogeneity noted: one (agent_type, step_name) covers hero/minimal-hero/header schemas.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4g (dataset profile)
- **relations:** part of Flywheel A; export_version enables downstream compatibility checks
- **verify-later:** stripMarkdownFromResponse in ai_actions.go; export_version field

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Local-step input resolution: input_mapping dead, key_path for loop items (109b)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 109b SQL header "input_mapping did NOT populate input_data.key for this loop substep … deriving the dataset key 40 times"; NOTES(39) "CORRECTS a load-bearing assumption"
- **what:** A load-bearing chassis contract discovered the hard way: the coordinator only resolves `input_mapping` for `call_agent` and loop fan-out, not for plain local action steps. Local actions (and local loop substeps) must read values via a config key holding a dot-path, resolved by `ExtractActionInputs` Strategy 0 / `resolveTemplateToken` / a `key_path`. Migration 109b fixed `presign_one` to read the loop item via `key_path:"ckpt_key"` (from CollectedData where setLoopVariable puts it) instead of the dead `input_mapping{key:ckpt_key}` that had presigned the dataset key 40×.
- **sources:** phase5/109b_fix_presign_one_loop_item_keypath(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#2, #update-2026-06-06-2, #8
- **relations:** cause of the "presigns the dataset key" failure signature; distinct from the await race
- **verify-later:** loop_expansion_handler.go setLoopVariable; content_search.go; datahelpers ExtractActionInputs

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Capability gate D5 — requires-backend semantic tag
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** plan(11) D5 "Planner gate (to apply): load_components gains AND NOT (… ? 'requires-backend')"; running_notes SQL revised; marked "Outstanding: apply the planner query change".
- **what:** Gates backend-requiring components off static sites by CLASS, not site-type. Component carries `semantic_tags:["requires-backend"]`; planner load_components excludes them unless opted in via roadmap section_types; site side sets `deploy_config || {"target":"vm","capabilities":["backend"]}`; a later audit check compares placed sections' requires-* tags to site capabilities.
- **sources:** traffic_probe_plan(11).md#decision-5, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection
- **relations:** supersedes the intent-probe site-type gate
- **verify-later:** build-site-planner workflow JSON load_components query

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded D5 — suitable_site_types / "intent-probe" site type gate
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4) "Decision 5 (OPEN) … component carries suitable_site_types:['intent-probe'] … planner's load_components gains AND suitable_site_types = '[]'::jsonb"; running_notes "the invented site type is GONE (suitable_site_types: [])".
- **what:** The earlier D5 formulation invented an `intent-probe` site type and gated via `suitable_site_types`. Dropped after operator feedback that "intent-probe is the wrong label" — the distinguishing feature is the class (has a backend), so the tag-based `requires-backend` gate replaced it.
- **sources:** traffic_probe_plan(4).md#decision-5-open, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight
- **relations:** replaced by requires-backend semantic tag gate (D5)
- **verify-later:** component semantic_tags vs suitable_site_types columns

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Class-level rename (probe → site-engine) and env-var churn
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** running_notes 2026-06-11 "RENAME MAP (every changed name)"; env var `PROBE_DB_PATH → ENGINE_DB_PATH`; then 2026-06-11 store-v2 "env var ENGINE_DB_PATH → ENGINE_DATA_DIR".
- **what:** When the box became the home of the whole backend-site class (not just probes), "probe" defaults were neutralised to class-level names across engine + deploy artifacts: service/user `site-engine`, `/opt/site-engine`, `/var/lib/site-engine`, `/etc/site-engine/site-engine.env`, webroots `/var/www/vm-sites/<d>`, rate zone `engine_rl`, hook `site-engine-deploy`. The DB-path env var was renamed twice: `PROBE_DB_PATH` → `ENGINE_DB_PATH` → `ENGINE_DATA_DIR`; store file `probe_events.json` → `intent_events.json` → (dropped for JSONL). ProbeSearch/ProbeCategory/ProbeFreeText kind constants kept (they name the feature, not the class).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-component-live, traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#changelog
- **relations:** supersedes probe-go naming
- **verify-later:** grep for stale PROBE_/probe_events across artifacts

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /stats endpoint + INTERNAL_API_KEY (stats internal key)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** service(24).go "GET /stats key-gated per-host summary"; runbook(12) env table "INTERNAL_API_KEY gates /stats (sent as X-Internal-Key) … unset → /stats 401"; verified over HTTPS 2026-06-12.
- **what:** `/stats` returns a key-gated per-host summary (visits/events counters), gated by `INTERNAL_API_KEY` sent as header `X-Internal-Key`. Unset key → /stats returns 401. The same key doubles as the read-only capture-export key for /events and /access-digest; on the collector side it is stored in `deploy_config.engine.stats_key` (low sensitivity, one accessor, movable to a secrets table later). The env file (not a shell variable) is the source of truth.
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#2, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** read by CollectIntentEventsAction
- **verify-later:** /etc/site-engine/site-engine.env INTERNAL_API_KEY; deploy_config.engine.stats_key

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /events export endpoint (P4 collector interface)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "GET /events built + tested … Tests green ×6"; runbook(12) §6 "Export endpoint (P4 collector interface)".
- **what:** `GET /events` streams stored events as key-gated NDJSON oldest-first (original line bytes preserved); params `since` (RFC3339, strictly-after), `host`, `limit` (default 5000). Final line `{"_meta":{count,truncated,server_time}}` aids the collector checkpoint (store max created_at → duplicate-free pulls). Lock-free by design so a big export never blocks live captures; a torn mid-append tail line is skipped and arrives next pull.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-events-export, traffic_probe_runbook(12).md#6
- **relations:** consumed by CollectIntentEventsAction; the pull architecture
- **verify-later:** store.go StreamEvents; nginx /events location

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wrapper-orchestrator requirement finding (001:405-462)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "STRUCTURAL finding … wrapper-orchestrator REQUIRED (001:405-462): the collector is reached via the SCHEDULER … AND does substantive in-chassis work → must NOT run in a shared agent-chassis pod".
- **what:** Because the collector is reached via the generic scheduler entry point AND does substantive work (HTTP to N boxes + multi-row upserts, unbounded as boxes grow), it must not run in a shared agent-chassis pod. Fix = a thin orchestrator that spawns a worker child into its own pod. Also corrected: the scheduler fires ONE message per tick and does NOT fan out pre_query rows, so the collector self-queries and loops, and the scheduled_tasks pre_query is a count>0 GATE.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-d, traffic_probe_running_notes(27).md#2026-06-13-c, traffic_probe_running_notes(27).md#standing-observations
- **relations:** shaped the two-agent topology; scheduled_tasks intent-collection target corrected to orchestrator
- **verify-later:** 001 guide §405-462; scheduled_tasks intent-collection target_topic

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Adapter Response Envelope Contract (003) — conditional, later dropped from plan
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4)/(5)/(6) P4/P5 "If collection runs as a chassis adapter, it MUST follow the Adapter Response Envelope Contract (typed-struct bool headers, reuse request_id, message_id, ProduceWithValidation)"; plan(11) reduces this to a one-line parenthetical.
- **what:** The guidelines-audit flagged that IF P4 collection or the P5 deployer were built as chassis adapters, replies must use a typed-struct envelope — getting it wrong = silent drop until timeout (the documented multi-day thunder fault). Once P4 was redesigned to need no adapter (key-gated HTTPS pull + one local action), the prominent P4/P5 envelope warnings were demoted.
- **sources:** traffic_probe_plan(6).md#phases, traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_plan(11).md#p4
- **relations:** applies only if P5 vmhost adapter is built
- **verify-later:** 003 contracts doc; ProduceWithValidation

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### flat-file constitution (precursor to the `standards` table, scope=constitution)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** "This is the flat-file version for the thin slice. Later it becomes the `standards` rows with `scope = constitution`; the content is the same."
- **what:** A single always-on rules document meant to be included in full in every LLM context bundle, distinct from the task-specific 003 contracts which are pulled in only when a task touches them. Covers reuse-before-recreate, structural-over-symptomatic fixes, one-orchestrator-per-agent, no-subworkflows-in-SQL, the snake_case/kebab-case string-naming test, chassis storage conventions (text+CHECK not native enums, version+previous_version_id, deleted_at not status=archived), logging discipline (no `logger.Debug`, always log the orchestration_id), and deployment/namespace facts. Explicitly designed to later graduate from a hand-pasted flat file into database rows.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/thin_slice_constitution.md
- **relations:** contextkit toolchain (above); trust ledger contract (below, references a `standards` lifecycle)
- **verify-later:** whether a `standards` table with `scope='constitution'` was ever actually created

<!-- SOURCE: U25_leopardess_social.md -->
### Section resolvers override content_data on every render
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** HANDOFF §4.4 (2026-07-12): "hero background_image is auto-resolved to /assets/images/hero.jpg (plan_sections_action.go:1338) — you change the FILE"; system-stats section removed because suffixes can't be overridden.
- **what:** Field resolution beats stored instance data: the hero background image is re-resolved to a fixed path every render (content edits can't remove it — replace the file); source:"static" schema fields re-apply their schema fallback every render and cannot be overridden per instance; a forked *section* component does not survive rerender because save_page_sections re-links to the canonical component by function, while a header/footer fork sticks because it is wired via the style collection.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#landmine-14; docs/leopardessconsulting/HANDOFF.md#4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12
- **relations:** static-source schema field misuse; per-site style fork
- **verify-later:** plan_sections_action.go resolvers; save_page_sections re-link behaviour

<!-- SOURCE: U25_leopardess_social.md -->
### Static-source schema fields force fleet-generic labels/suffixes
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** RUNBOOK_minilobby §0 (2026-07-12): "FIXED … migration 090 — stat_1/2/3_label + cta_label were source='static' … forcing 'Clients Served'/'Satisfaction Rate' … on every render"; leopardess system-stats suffixes (%/ms/+/x) still schema-forced (section removed instead).
- **what:** A recurring defect class: shared components whose label/suffix fields are source:'static' re-apply business-generic fallbacks on every render, so writers can only fill values, producing crossed pairs ("2,767%", "500+ Models / Clients Served") on every site. Fix pattern: flip source static→llm keeping fallback as safety net (migration 090 for content-block-about, 13 pages/5 sites); a suffix-free stats component is a noted platform addition. 
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/090_content_block_stat_labels_llm.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12; docs/leopardessconsulting/HANDOFF.md#8
- **relations:** section resolvers override content_data; shared component library semantics
- **verify-later:** content_components 4e448d51 input_schema; system-stats schema fields

<!-- SOURCE: U25_leopardess_social.md -->
### Component creation contract (the generator's embedded rulebook)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003d carries the full contract prompt "included in the component-creator agent definition"; generated components (brief-explanation regen, archive-list) conform.
- **what:** The complete contract an LLM must follow when generating a component: scoped `<style>` + `<section class="{function}-section" data-component="{function}">` + optional IIFE script; kebab-case function naming; {{.field}} template variables with a declared input_schema (type/source/required/llm_guidance, sources llm|site_specs.*|site_assets.*|renderer|static); all colours via CSS variables with fallbacks and dark-section --section-* custom properties; scoping and responsive rules; client-side-only JS, no CDN; quality bans (no placeholders, no unrendered variables, no fabricated content). Compiled from docs 003 + 018 + input-schema v2; storage location decision: static in the agent prompt now, knowledge_base + rag_lookup if contracts churn.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Component-creation-contracts
- **relations:** component selector/creator; generation-time guards for dynamic components; contracts-and-standards docs 003/018
- **verify-later:** component-creator agent_definitions prompt_template vs the 003d contract text

<!-- SOURCE: U25_leopardess_social.md -->
### plan_sections deferral semantics (required + unresolvable = silently dropped section)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation 2026-07-02: "required field + source unresolved + on_missing empty → hits default: → shouldDefer=true → section deferred → save_page_sections keeps only ready sections."
- **what:** Schema authoring contract: a required field whose data source cannot resolve (e.g. site_assets.illustration with no asset) defers the whole section out of the page save — the instance is dropped, not errored. required=true + on_missing=skip_field is contradictory (still defers). Fix patterns: make decorative fields optional with skip_field/fallback, and populate shared spec gaps once (site_specs.cta.primary_url served gauntlet-cta + system-stats + brief-explanation). llm-sourced fields can never defer — which is why generation-time guards make header copy llm.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-02, #2026-07-03; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Design-decisions
- **relations:** generation-time guards; illustration asset resolution; component creation contract
- **verify-later:** planSection deferral switch in plan_sections_action.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Page-build-handler pipeline with plan_sections triage (Layer 0) and validate_content gate
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 002(4)/002d full pipeline; input-schema v2 in 003(8)
- **what:** ensure_site_record → load_page_record → plan_sections (resolve each section's input_schema sources; triage into ready/deferred(needs_human_review)/skipped; page deploys with whatever is ready) → content writer (only ready sections) → validate_content (algorithmic: placeholders, unrendered templates, cross-site contamination, broken links, hallucinated emails; blockers/errors → needs_human_review) → save_sections → deploy. Quality gates before generation AND after; content writers never fabricate non-llm-sourced data.
- **sources:** 002(4)#Page Build Handler Pipeline; 002d Layer 0; 003(8)#Component Input Schema v2, #Content Validation Contract
- **relations:** input schema v2; needs_section_data; growth budget
- **verify-later:** plan_sections_action.go; validate_page_content.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component input schema v2 (sources, on_missing vocabulary)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full contract; 016b deferral-drop entry shows live enforcement including the default:defer trap
- **what:** content_components.input_schema declares per-field type/source/required/on_missing/fallback/min_items/llm_guidance. Source prefixes: llm, site_specs.*, site_assets.*, pages.*, config.*, renderer, static, query.*. on_missing: use_fallback/skip_field/skip_section/needs_human_review/block. Image fields must be required:false + skip_field + template-gated (imagery is async). Trap: required:true with on_missing skip_field/empty hits the switch default and defers the section.
- **sources:** 003(8)#Component Input Schema v2; 003(8) checklist 6b; 016b#Regenerated content section deferred
- **relations:** plan_sections; queryresolve; imagery async
- **verify-later:** planSection switch in plan_sections_action.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### content_data is the source of truth (rendered_html is derived)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full section incl. the two re-render paths; "HTML patching was rejected as an edit mechanism"
- **what:** Every section stores content_data (structured) + rendered_html (derived). All edits go through content_data; the light path (rerender_page_sections) re-renders from stored content_data ⊕ fresh-resolved fields with no LLM, persisting the merged content_data so rows stay complete render sources; NULL content_data escalates to full rebuild.
- **sources:** 003(8)#Schema Enforcement/#Source of truth; 002(4) page-rerender row
- **relations:** work-item routing; section editor
- **verify-later:** rerender_page_sections persistence semantics

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Workflow result contract (flatten vs fields vs mapping; output_field/output_fields foot-gun)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) opening contract; fix deployed 2026-06-18 per 016 §9 (resolveResultSpec)
- **what:** A workflow's complete step declares its result via result_from (flatten — field contents become body), multiple_output_fields (nested per key), or result_mapping; deprecated aliases still resolve with a Warn. No key → fallback dump of collected_data, which can breach the ~900k cap. Parents read at `<call_output_field>.response.<key>`; wrong mode = silent null reads. Historically singular `output_field` was silently ignored → stub-with-success (Part 1 bug); the oversize path now returns an actionable error instead of a stub.
- **sources:** 003(8)#Workflow Result Contract; 016 §9 "Child workflow result silently replaced by a stub"
- **relations:** silent-completion family; result_spec.go
- **verify-later:** result_spec.go; remaining deprecated-alias agents

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component naming contract (kebab function, data-component, uniqueness)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) with live DB constraint chk_function_kebab_case and partial unique index
- **what:** content_components.function is the canonical identifier (kebab, regex-constrained, one active row per function); data-component attribute on the root element must match exactly; page_components.slot_name mirrors function; GetComponentWithFallback (exact→normalized→generic-text-block) is a safety net not to be relied on. Naming patterns for page-specific heroes and header/footer/head variants.
- **sources:** 003(8)#Component Naming Contract
- **relations:** section_type vs function split (007); slot_name↔function mapping hazard
- **verify-later:** chk_function_kebab_case; idx_content_components_unique_active_function

<!-- SOURCE: U01_docs024_numbered_core.md -->
### String-value naming convention (snake for identifiers, kebab for data)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): kebab CHECKs live; "snake_case columns have not yet had explicit constraints added — follow-up"
- **what:** Values used as identifiers in code (map keys, switch cases, dispatch routes) are snake_case (item_type, action names); pure-data values that end up in CSS/URLs/HTML are kebab-case (function, page_type, agent type); single words lowercase. Decision test: is the value ever a Go case/map key?
- **sources:** 003(8)#String-Value Naming Convention
- **relations:** page_type vocabulary; item_key canonicalization
- **verify-later:** snake-case CHECK constraints existence

<!-- SOURCE: U01_docs024_numbered_core.md -->
### page_type vocabulary and "landing, not index"
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8)/016 §6.5: constraint chk_page_type_kebab_case since migration 051; canonical value table
- **what:** Canonical kebab page_types (landing/content/tool/guide/game/blog-post/blog-index/section-index/entity-page/entity-directory/news-index); the homepage's TYPE is landing while its NAME is index (name is storage convention, type is kind-of-page). CanonicalisePage normalises legacy snake inputs one-way. Guides nest at /guides/<slug>/index.html and appear in guide-lists only when typed guide AND active/deployed.
- **sources:** 003(8)#page_type vocabulary; 016 §6.5
- **relations:** CanonicalisePage; adoption slug-mangling
- **verify-later:** pages page_type distribution; constraint present

<!-- SOURCE: U01_docs024_numbered_core.md -->
### JS content separation contract (js_content → /tools/assets/{function}.js)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) full flow through separateInlineJS/collectJSAssets/multi-file git commit
- **what:** Component JS is split out of html_template into js_content and served as an asset file; html_template keeps only a `<script src>` reference. separateInlineJS extracts only attribute-less `<script>` tags (by design). js_snippets is a separate table for shared design effects, never component behaviour. Known failure class: pre-extraction rows render as empty shells (016b entry).
- **sources:** 003(8)#JS Content Separation Contract; 016b#Data-driven component shells render empty
- **relations:** tool doc header stripping; empty-shell taxonomy
- **verify-later:** separateInlineJS regex; script-balance validation hardening applied?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component creation & regeneration contract (created/regenerated; already_exists removed)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) + 026(2) dated 2026-04-20 describing deployed behaviour
- **what:** StoreGeneratedComponentAction: Layer-1 pre-store validation runs before either branch (rejection never touches the DB); create INSERTs + v1 snapshot; regenerate snapshots old state then UPDATEs in place (UUID preserved, FKs intact), marks dependent page_components pending, and raises one deduped needs_rerender per affected site (item_key component_regen_rerender:<uuid>). Downstream must not assume component_id is new nor create its own rerender items. Regen keying is by the LLM's EMITTED function — a mismatched name INSERTs a stray.
- **sources:** 003(8)#Component Creation & Regeneration Contract; 026(2) full; 016b#Manually invoking an agent (regeneration keying)
- **relations:** component_versions; markPagesForRebuild; system-stats key-mismatch incident
- **verify-later:** store_generated_component_action.go branches

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site component linkage contract (slot_name↔function; fallback header hazard)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract + discovery check unlinked_site_components
- **what:** Every site_components row must have component_id → content_components; otherwise renderAndStoreSiteComponent falls to a generic function lookup (which cannot match because slot 'header' ≠ function 'header-<variant>') then to hardcoded RenderFallbackHeader (no logo, stacked nav, search icon, dark). Breakers: update_site_defaults not run, NULL collection header id, legacy data. Self-healing check + site-component-linker handler exist.
- **sources:** 003(8)#Site Component Linkage Contract; 004 discovery checks
- **relations:** four overlapping chrome stores (036); light-site-dark-chrome bug
- **verify-later:** update_site_defaults in workflows; unlinked check registration

<!-- SOURCE: U01_docs024_numbered_core.md -->
### CSS colour inheritance model (--section-* with fallbacks)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8): "the single most important rule in the design system"
- **what:** Base CSS: body sets --color-text; h1-h6 use var(--section-heading, var(--color-primary)); p/li/blockquote use var(--section-text, inherit); strong/em/span set no color; links are the explicit exception. Painting sections override --section-* on their container and all children adapt. Setting color directly on elements bypasses the override — the light-on-light testimonial bug.
- **sources:** 003(8)#CSS Colour Inheritance Model; 036 §4
- **relations:** section painting contract; buildSectionDefaults
- **verify-later:** layouts' element rules follow the fallback pattern

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section painting contract (four models, references-only) — supersedes literal dark-section overrides
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) checklist item 6 + Section Context Variable Contract; (7)→(8) diff shows the literal-colour version replaced; "fix_forced_text_colors enforces this mechanically"
- **what:** A template's appearance derives from what its own CSS paints; is_dark_section is catalogue metadata and must not key styling. A painting section picks exactly one model — pair band, palette band, image/layered (hero-ink), or ambient (no --section-* at all) — and re-exports --section-* AS REFERENCES to the tokens it paints with (color-mix for muted/surface/border). Literal colours in --section-* declarations forbidden. The older contract (dark sections set literal rgba/white values) is superseded.
- **sources:** 003_contracts_and_standards(8).md items 6/6b + #Section Context Variable Contract; 003(7) (family-delta, superseded form)
- **relations:** scheme-to-components work (016b light-site-dark-chrome); forced_text_colors check
- **verify-later:** fix_forced_text_colors action; component templates conformance

<!-- SOURCE: U01_docs024_numbered_core.md -->
### CSS theme template contract (renderer vs template ownership; theme storage/lineage)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) responsibility split; render pipeline confirmed in 036
- **what:** Renderer owns palette injection, luminance-driven --section-* defaults (pickReadableOnBackground preserving palette character), and css_snippets appends; the theme template owns layout/typography/component styling using the fallback pattern and MUST NOT declare --section-* defaults or hardcode text hexes. css_template (Go template) vs css_content (frozen fork snapshot, reference only).
- **sources:** 003(8)#CSS Theme Template Contract; 036 §3–4
- **relations:** 025 palette/layout/typography split; buildSectionDefaults
- **verify-later:** render_css_from_spec_action.go; color_util.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Query parameterisation contract ($1 + params, never template interpolation)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** 003(8): rule + named legacy offenders (tool-suggester, tool-improver) still to migrate
- **what:** All new query_database usage must use $1 placeholders with a params array of dot-paths (passed as query args); {{.field}} embedding is a SQL injection risk. QueryDatabaseAction gained params support after audit agents failed on $1-with-no-args.
- **sources:** 003(8)#Query Database Parameterisation Contract; 001(5) bug 1
- **relations:** authoring rules pack
- **verify-later:** tool-suggester/tool-improver migrated?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Schema enforcement: flexible vs strict mode with approval snapshots
- **category:** contracts-and-standards
- **status-signal:** unknown
- **status-evidence:** 003(8) describes the design (schema_snapshot/content_snapshot at approval, sites.schema_mode) with no deployment claim or date
- **what:** Initial build runs flexible (best-effort substitution, warnings); at approval the structure locks: page_components.schema_snapshot + content_snapshot captured, sites.schema_mode → strict, mismatches become validation errors, template upgrades can't break approved pages.
- **sources:** 003(8)#Schema Enforcement (Flexible vs Strict Mode)
- **relations:** locks; content governance approval flows
- **verify-later:** schema_mode column usage; any strict-mode site

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Handler input-path contract (input_data.spec.*) + action-level defense
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003(8) contract rule; 016 §9 "most common systematic failure" documents the violations
- **what:** The dispatch loop passes the work-item spec nested at input_data.spec; handlers MUST read spec fields there, never rely on top-level flattening (which exists only for legacy `?` promotions and silently nils). Go actions reading common fields implement a fallback chain (explicit config → input_data.spec.field → well-known spots). Known offenders were tool-improver/tool-auditor/rerender-pages. Manual spawn+call of work-item agents must satisfy BOTH the top-level input_contract AND the workflow's spec paths (provide fields in both shapes).
- **sources:** 003(8)#Handler agent contract/#Input data paths; 016 §9 path-mismatch; 016b#Manually invoking an agent
- **relations:** dispatch loop; input_contract validation
- **verify-later:** load_page_record_action.go fallback chain

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Legal rules schema and content_direction page-level instructions
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** 003(8) schemas defined; legal-content-agent "Planned" in 002(4)
- **what:** Per-site legal_rules (required disclaimers with triggers/placement, forbidden phrases, required pages seeded per industry) in sites.content_data; pages.content_direction jsonb (format/instruction) flows to the content writer for page-level rewrites via needs_rebuild.
- **sources:** 003(8)#Legal Rules Schema, #content_direction
- **relations:** content agent family; compliance discovery
- **verify-later:** any legal_rules populated; content_direction reads in writer prompt

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tiered component field classification (Tier A voice / B tunable static / C site data)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "revise_component_creator_prompt.sql (applied)" (2026-04-17)
- **what:** Component schemas classify fields: Tier A voice content (source llm, required), Tier B tunable labels (source static, optional, with fallback), Tier C site data (source site_specs.*/site_assets.*). Prevents both "35 required fields" and "0 fields, everything hardcoded". Template/schema sync invariant: every {{.x}} has a schema entry and vice versa. Tier B static fallbacks later become the language problem ("Browse All Tools" on non-English sites) and the "soft static" override idea.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#5; FOCUS_language.md#static-fallbacks
- **relations:** LLM reliability; language surfaces
- **verify-later:** component-creator prompt in agent_definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### system.internal site convention
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Created for maintenance/library-level work items … id: eac60db8-…, domain: system.internal" (2026-04-17)
- **what:** A never-deployed sites row (brand_dna.is_system=true) that hosts library-level and maintenance work items not belonging to any customer site (e.g. component_quality_scan backfills). Side effect: its maintenance-pipeline items sit dormant (no maintenance-dispatch-loop) and it absorbs untargeted scheduler dispatches.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#7; FOCUS_dispatch_diagnostic(4).md#Q4; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** pipeline soft label; Bug 3 site targeting
- **verify-later:** sites row; items accumulated on it

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### String-value naming convention (identifier-shaped snake, data-shaped kebab)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Status: applied (migration 051, page_canonical.go + page_role_validator.go updated, contracts doc v9, debug guide v2.10)" (2026-05-17)
- **what:** Decision rule for string-typed columns/enums: used as a Go identifier (switch case, registry key, dispatch route) → snake_case (site_work_items.item_type); pure data describing what a thing is → kebab-case (pages.page_type, content_components.function); single word → bare lowercase (statuses). Root incident: normalisePageType wrote snake while all readers expected kebab, silently hiding blog pages. Companion fix: homepage page_type 'index' → 'landing' (name vs type conflation). Snake-input fallback retained as a bounded migration-tail exception; tests document behaviour, not intent.
- **sources:** FOCUS_naming_conventions_kebab_vs_snake.md (whole)
- **relations:** Tension #2 canonicalisers; page_type vocabulary gap
- **verify-later:** migration 051; CHECK constraint on pages.page_type

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Spec-is-primary-input contract for handler workflows
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Spec is primary input (contract rule) — all handler workflow configs must use input_data.spec.* paths" (architecture decision, 2026-04-17); root cause of gauntlet pages getting 0 components
- **what:** Dispatch only reliably populates input_data.spec.*; top-level flattened paths (input_data.page_name) depend on optional `?` input_mapping and silently resolve nil. Handler configs use spec paths; Go actions keep a defensive fallback chain.
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#4, #Architecture-decisions
- **relations:** chassis input conventions; flat-namespace collisions
- **verify-later:** contracts doc handler-agent section

<!-- SOURCE: U03_idea_uk_section_data.md -->
### {function}-section class contract and data-component naming contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Notes (Sf): "The `{function}-section` contract is REAL + operative, honoured unevenly" — honoured by the 5 surface sections + footer-with-disclaimer; not by hero (`.hero`) or CTA (`.cta-section`).
- **what:** Layouts and `buildSectionDefaults` key structural rules and surface treatment on `.{function}-section` class names, but the compiler (`CompilePageSectionsAction`) concatenates component HTML without wrapping, so the class is each component's own responsibility and adoption is inconsistent — the mechanism misses non-adopters and their inline CSS wins. Separately, every component root does carry `data-component="{function}"` (kebab-case, enforced by component_validation.go), giving an attribute-selector escape hatch the class mismatch cannot break.
- **sources:** running_notes_scheme_to_components(55).md#Sc #Se #Sf; HANDOFF_scheme_to_components_for_claude_code(1).md#Established; PLAN_scheme_to_components(1).md#Q4
- **relations:** Colour Inheritance Model; SectionStyles (dead consumer of the same names); section painting contract.
- **verify-later:** component_validation.go naming checks; class emission across `content_components.html_template`.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Colour Inheritance Model (var(--section-*, var(--color-*)) chains)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** HANDOFF §Established: "Element rules follow the Colour Inheritance Model: `var(--section-*, var(--color-*))`."
- **what:** The base element rules in layouts/base CSS resolve text/heading/muted/border colours through a two-level chain: a section-scoped custom property if declared, else the palette-level colour. This is what makes "components declare no colours" viable: a non-painting section inherits page ink automatically; a painting section re-exports its context onto the `--section-*` layer and every child element follows. The W3e "ambient pass-through" fix (`--section-x: var(--color-x)`) exists because some internal consumers lack var() fallbacks — deletion would fall to currentColor/transparent.
- **sources:** HANDOFF_scheme_to_components_for_claude_code(1).md#Established; SPEC_scheme_to_components.md#The-contract; running_notes_scheme_to_components(55).md#Sx (fix rationale)
- **relations:** section painting contract; buildSectionDefaults.
- **verify-later:** base head CSS / layout css_templates element rules.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Paired-variable ("on-colour") standard — the decision record
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC (2026-07-02) "The standard is **paired variables** … Checks 4b/4c show it is the existing library standard incompletely adopted: 18/18 layouts define `--color-primary-text`, 17/18 define `--color-cta-text`"; executed via W1–W6 and closed 07-03.
- **what:** Every paintable band colour has a matching text colour, curated per layout (and therefore per scheme), overridable per site through palette specialised slots (theme-wins merge), with per-instance control later available via `site_plan_directives` scope=section. Selected over four alternatives (Alt 0 stale-build, Alt A component-owned bands, Alt B renderer-owned via is_dark_section, Alt full-025) after the user's gating answer: "a light scheme must be able to render fully light, and may carry dark hero bands" — band darkness must be a choice, not a component constant. Existing names are reused (`--color-header-bg/-text`, `--color-footer-bg/-text`, `--color-cta-bg/-text`, `--color-primary/-text`, `--color-hero-title/-subtitle`); the direction is completion of existing architecture, not restructure.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sn #So; HANDOFF_scheme_to_components_for_claude_code(1).md#The-Decision; RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS
- **relations:** section painting contract (its component-facing rules); layout CTA pair curation; is_dark_section demotion; supersedes SectionStyles and defers Phase 4.5.
- **verify-later:** layout css_templates pair definitions; palette specialised-slot merge in composition helpers.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section painting contract (003 item 6 rewrite: four painting models)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** slice4b delivers the 003 rewrite as a patched doc; RUNBOOK 07-06-night: "Slice 4b: DELIVERED as a patched doc … STEP A — copy `outputs/003_contracts_and_standards_7_.md` over the repo's 003 doc" — repo copy still a pending user step at last dated evidence.
- **what:** Replaces 003's "Dark Section Contract" (if is_dark_section=true, template MUST set `--section-*`) with: a template's appearance derives from what its own CSS paints; a painting section chooses exactly one model and re-exports `--section-*` AS REFERENCES ONLY — (a) pair band re-exporting the pair text, (b) palette band re-exporting the on-colour family (`--color-primary-text, var(--color-background)`), (c) image/layered background defining `--hero-ink` per branch, (d) ambient: no background of its own and NO `--section-*` at all. Literal colours in `--section-*` declarations are forbidden; muted/border/surface derive via `color-mix`. The old contract is the exact inverse — the concept records a full contract reversal.
- **sources:** SPEC_scheme_to_components.md#The-contract; slice4b_003_contract.md; running_notes_scheme_to_components(55).md#Sh (old 003 item 6) #Ui
- **relations:** paired-variable standard; component-creator prompt re-aim; fix_forced_text_colours re-aim (mechanical enforcer); image fields rule 6b.
- **verify-later:** repo `docs/.../003_contracts_and_standards*.md` item 6/6b current text; whether outputs copy landed.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### is_dark_section demoted to catalogue metadata
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** SPEC consequences: "`is_dark_section` is demoted to selection/imagery metadata (6 of 37 declarers contradict their own flag — never key styling on it)"; slice4a prompt text landed 07-06 ("catalogue metadata ONLY").
- **what:** `is_dark_section` is a component-level boolean authored by the component-creator LLM (`store_generated_component` extracts it from generated JSON), scheme-blind (the needs-new-component spec carries no scheme field, so the library skewed dark independent of sites), and unreliable (6/37 self-declarers contradict their own flag). The decision demotes it: nothing may style from it — styling derives from what the template's CSS paints. It survives only as selection/imagery metadata. The earlier Q5/E design question ("where does per-section contrast intent live — site_plan_sections?") dissolves under the paired-variable model.
- **sources:** SPEC_scheme_to_components.md#Decision-record; running_notes_scheme_to_components(55).md#Sh #Sn; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (flag hygiene); slice4a_creator_prompt.sql
- **relations:** section painting contract; fix_forced_text_colours re-aim (isDarkSection param kept-ignored).
- **verify-later:** store_generated_component_action.go extraction; component_selector.go non-use; creator prompt current text.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### component-creator prompt re-aim (painting rules, vocabulary, image-fields rule)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Notes (Ui): "4a evidence: gate t/t/t/t/f → UPDATE 1 t/t/t/f" (2026-07-06); slice4a_creator_prompt.sql RETURNING confirms SECTION PAINTING + IMAGE FIELDS RULE in, old DARK SECTIONS block gone.
- **what:** Four targeted needle replaces inside `agent_definitions.default_config->>'prompt_template'` for component-creator: the dark-sections literal block becomes the four painting models (references only; is_dark_section reported honestly but "nothing may style from it"); the consumer chain replaces the dark line; item 7's vocabulary gains the cta pair and extended tokens (surface-alt, hairline, code-bg, callout pair); Tier C gains the image-fields rule (site_assets.* fields required:false + skip_field + gated markup — described rather than shown because the prompt is itself Go-template-rendered and literal if-syntax would execute). Root cause it addresses: the generated half-migrated footer and brief-explanation proved the prompt was emitting components that consume chrome vars while self-declaring dark text — drift continues until the contract lives in the prompt.
- **sources:** slice4a_creator_prompt.sql; running_notes_scheme_to_components(55).md#Uf #Uh #Ui; RUNBOOK_scheme_to_components(50).md#CHECK-2-RESULTS (corollary)
- **relations:** section painting contract; agent re-registration vs re-seed risk; image fields optional-with-gate.
- **verify-later:** agent_definitions component-creator prompt_template current text; the Step C grep ("DARK SECTIONS (if the section has a dark background)" in Go sources).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Image fields optional-with-gate contract
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** User decision (Th, 2026-07-03): "imagery must not block section rendering"; W7a gate applied (Tm: "UPDATE 1, gated t/t"); the rule landed in the creator prompt (4a) and 003 6b; `illustration_url` was already required:false + skip_field in schema.
- **what:** Any `site_assets.*`-sourced component field MUST be `required: false` with `on_missing: skip_field`, and its markup MUST be gated with a template conditional (`{{if .illustration_url}}` around brief-explanation's image wrapper is the model — Go templates treat "" and missing as false, covering the src="" broken-image case). Imagery arrives asynchronously and must never block or defer a section; the section renders imageless and the image is added by the pipeline's own queued rebuild. Codified as 003 item 6b and the creator prompt's IMAGE FIELDS RULE.
- **sources:** w7a_01_gate.sql; slice4b_003_contract.md#Edit-1 (6b); slice4a_creator_prompt.sql (R4); running_notes_scheme_to_components(55).md#Th #Tl #Tm
- **relations:** plan_sections field deferral semantics; section-scope imagery pipeline; component-creator prompt re-aim.
- **verify-later:** brief-explanation html_template gate; input_schema on_missing values across image-consuming components.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Component schema-template consistency invariant
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "The governing invariant — a component's schema `items` must match its template tokens — is the right thing to hold; the reconciler enforces consistency toward the current schema."
- **what:** A component's `input_schema` (shape `{"fields": {...}}`, unmarshalled in component_library.go) is the contract for its `html_template` tokens: array item field names in the schema must match what the template reads. The reconciler, the prompt, and generation all derive from input_schema, so divergence breaks all three coherently. Known violation: info-card-grid's stored html_template literally contains `<no value>` (rendered-against-nil output apparently written back into the template column) — flagged as its own repair thread and never fixed inside this unit. services-grid shares differentiators' schema byte-identically and was healed by the same fix.
- **sources:** running_notes_checkpoint_uu.md#Confirmed-during-the-hardening-review; running_notes_checkpoint_ss(1).md#Root-cause-in-code; RUNBOOK_pcw_item_fields_fix.md#Follow-on
- **relations:** array item-fields contract; render-time reconciler.
- **verify-later:** info-card-grid html_template (still `<no value>`?); services-grid first-use spot check.

<!-- SOURCE: U04_idea_uk.md -->
### CSS colour inheritance / --section-* luminance model (inline style = override, not main CSS)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Documented in 003 ("CSS Colour Inheritance Model") and restated as renderer contract in the tool-portal-light css_template header; the dark components *bypassing* it is the observed live behaviour.
- **what:** The platform's component colour contract as evidenced here: styles.css sets body/heading/element rules through `var(--section-*, var(--color-*))`; the renderer appends `--section-*` defaults after rendering based on palette/background luminance; a dark callout section overrides `--section-*` on its own container (sanctioned); layouts MUST NOT declare `--section-*` defaults; renderer-managed surface classes must be surface-coloured; a component's inline `<style>` is an **optional override, not its main CSS** (user correction, checkpoint mmm). The scheme gap exists precisely because dark components violate this — hardcoding backgrounds and `--section-*` inline. Two parallel styling systems (layout class vocabulary vs component class vocabulary) is the structural tension to resolve (Q4).
- **sources:** idea.uk/running_notes_2(6).md (lll/mmm); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/migration_layouts_scheme_and_light_tool_portal.sql (renderer contract comment); idea.uk/001_component_flow.md
- **relations:** scheme-as-override thesis; styling-render-pipeline (036); 003 contracts doc.
- **verify-later:** the luminance-appender code location (report investigation G names finding it).

<!-- SOURCE: U09_adoption.md -->
### Numbered-flat-fields anti-pattern (25 components)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** "Twenty-five active components match the numbered-field signature" (May 2026 audit); game-list rewritten Tier-D 2026-06-04; the rest (nav, card/grid, tier/stat, tool-internal clusters) unmigrated.
- **what:** Schemas declaring `post1_title…post6_*` with `source: llm` force the LLM to fabricate list items by structure (invented games, duplicate Jelly Invaders, links to nonexistent URLs) — no prompt rule can save a schema that demands invention. Groups: 8 navigation components (need a curated `nav.*` source, not query-resolvable), 7 card/grids (straight items-array rewrites), 5 tier/stat (may fit `site_specs.<aspect>.items`), 5 tool-internal field clusters (heterogeneous, case-by-case). Component-creator must refuse the shape for new "list of N things" components.
- **sources:** FOCUS_component_schema_patterns.md, migration_game_list_tier_d.sql header, CATALOGUE(9)#family-c
- **relations:** Tier-D contract; curated-list source vocabulary decision (deferred)
- **verify-later:** re-run the `<prefix>1_/2_/3_` audit; component-creator prompt Tier-D block

<!-- SOURCE: U09_adoption.md -->
### Anti-fabrication content path (Step 2/3: llm_field_specs, targeted prompt, merge_with)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Step 2 status: PASS" (2026-05-12 observations); Step 3 deployed with verification surface; tool-list rendering real pages confirmed on later runs.
- **what:** Closed the gap where `page-content-writer` fabricated items despite queryresolve existing: plan_sections resolves query.* before the LLM call and carries a full per-section `Component` + `resolved_data` + `llm_field_specs` (built from `llm_guidance`) on `section_plan.sections_ready`; the writer's prompt was rewritten to ask only for the `source: llm` fields (anti-"pink-elephant": it does not enumerate forbidden fields); `RenderComponentAction` honours `merge_with: current_section.resolved_data`, overlaying resolved data as authoritative over LLM output. Shared `loadSectionComponents` helper extracted.
- **sources:** FOCUS_directory_builder_and_list_components.md#implementation-history, old2/STEP2_changelog.md, old2/STEP3_changelog.md, old2/step3b_prompt_template.txt
- **relations:** Tier-D contract; page-content-writer workflow (spawn_research → load_site_specs → prepare_link_context → build_render_context → loop → compile_page)
- **verify-later:** plan_sections_action.go llmFieldSpec; page-content-writer def process_sections_loop iterate_over; RenderComponentAction merge_with

<!-- SOURCE: U09_adoption.md -->
### plan_sections on_missing semantics and the cta_url required-field deferral
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "guide-list empty on guides hub AND root index — RESOLVED (Part 14p)… set cta_url.required=false on guide-list_pre_037 and blog-listing_pre_037 (APPLIED)."
- **what:** plan_sections' required-field switch handles use_fallback/skip_section/needs_human_review/block but has no `skip_field` case, and `on_missing` defaults to skip_field — so a required field with an unpopulated site_specs source fell to default-defer and held the entire section (hero-only hubs). A query-sourced list field can never defer (resolver returns a non-nil empty slice). Deliberate decision: fix the component (required=false), not the engine — the defer-for-safety default on required fields is defensible. Related follow-ups: cta_url fallbacks (`site_specs.identity.*_index_url` unpopulated → `href=""`; game-list gained a `/games/index.html` fallback, tool-list still lacks one), and inconsistent cta source vocabulary across sibling list components.
- **sources:** running_notes_14(25)#part-14p, HANDOFF_2026-06-06#resolved, NEXT_CHAT_INPUTS_2026-06-06 §4
- **relations:** needs_section_data; silent-fallback link family; Tier-D contract
- **verify-later:** plan_sections_action.go required-field switch (~line 44340 of dump); content_components cta_url required flags

<!-- SOURCE: U09_adoption.md -->
### Chrome templates must be variable-driven (pre-store hardcoded-link gate)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** "Architecture decided, implementation pending" (FOCUS_chrome_templates, 2026-05-06); zero of the active footer templates consume `{{.nav_items_html}}` today.
- **what:** Header/footer components are LLM-generated with hardcoded `<li>` links, freezing nav at generation time and bypassing populate_nav_tables/classifyPagesForNav dedup entirely (doc 003's explicit rule violated). Proposed structural enforcement: a pre-store validation gate in store_generated_component_action.go rejecting chrome templates with hardcoded internal links outside {{range}}/{{if}} blocks, plus prompt teaching and a chrome-template-repair migration. Companion: `buildServicesHTML` is a parallel, dedup-less nav query ("Tools, Guides, Games, Games, Tools" verbatim) — drop it and its `services_html` context field once templates use quick_links_html. Principle: "as algorithmic as possible and enforced", not prompt-led.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-1, old2/HANDOFF_2026-05-07(1)#8
- **relations:** nav dedup guard B-029-1; render-context variables table (nav_items_html, quick_links_html, categories, footerLinks…)
- **verify-later:** store_generated_component gates; content_components chrome templates for {{.nav_items_html}} usage; buildServicesHTML existence

<!-- SOURCE: U09_adoption.md -->
### Thin-slice constitution (always-on rules; future standards rows)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "The always-on rules for any task on this codebase. Included in full in every bundle… Later it becomes the standards rows with scope = constitution" (thin_slice_constitution.md).
- **what:** The flat-file constitution: reuse before recreate; fix structural problems not symptoms; every agent is an orchestrator; reply on the caller's responses topic; workflows thin/complexity in Go; no SQL subworkflows — spawn sub-agents; check schema before SQL; parameterised queries only; snake_case for identifier-shaped values vs kebab-case for data-shaped values; text+CHECK not native enums; soft-delete via deleted_at; no logger.Debug; log orchestration/correlation ids and inter-agent messages; deployment path git→Actions→B2; plain pragmatic tone. Task-specific 003 contracts are pulled in only when touched. Reinforced session rules: snapshot before any DB change (restorable bak tables, in-txn, NAMEDATALEN 63), resolve site_id fresh via domain, don't conclude from partial signals.
- **sources:** docubundle/thin_slice_constitution.md, HANDOFF_2026-06-09#standing-rules, running_notes_14(25)#note-db-change-snapshot-standard
- **relations:** 003 contracts; standards table (aspirational storage form); council-agent charters
- **verify-later:** standards table existence with scope=constitution (expected absent)

<!-- SOURCE: U10_imagery.md -->
### Dispatch input contract for handler agents
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Recorded as a hard-won mechanism 2026-07-12: handlers invoked by build-dispatch-loop receive `input_data.spec`, `input_data.site_id`, `input_data.domain`, `input_data.item_type`.
- **what:** The canonical payload shape a dispatched handler sees; step conditions and input paths must be written against it (e.g. asset-deployer's check_mode tests both `input_data.spec.mode` and `input_data.mode` to cover dispatch and direct-call shapes). Divergence between dispatch-shape (`input_data.spec.*`) and direct-call shape (`input_data.*`) is a live source of latent extraction bugs.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, SQL_2026-07-11_asset_deployer_brand_head_mode.sql, SQL_2026-07-12_asset_deployer_explicit_paths.sql (NOTE block)
- **relations:** ExtractActionInputs lesson; work-item dispatch semantics.
- **verify-later:** build-dispatch-loop spawn payload construction.

<!-- SOURCE: U12_docs024_archives.md -->
### CSS section-colour model: inheritance → hardcoded dark-section variables → renderer-computed defaults → token-referencing painting sections
*(merged from 4 independent findings)*
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Four independent archive lines converge on the same evolution. Earliest baseline (`old/older1/003_contracts_and_standards.md`, `old_design_and_styling/FOCUS_design_and_styling.md`): plain CSS inheritance, dark-background components just set literal `color:#fff`/`inherit`, no `--section-*` variables exist. Middle era (`old/older1/003_contracts_and_standards_v2..v7.md`, `debugging_old/003_contracts_and_standards_v8..v11.md`, `archive_april_26/003e_contracts_and_standards_v5.md`): a `--section-*` custom-property contract keyed off a boolean `is_dark_section` column, with LITERAL hardcoded values (`--section-heading:#ffffff`, `rgba(255,255,255,0.9)`) enforced by `ValidateDarkSectionContract()`. An intermediate renderer change (`old_design_and_styling/PHASE_4_4_cleanup_summary.md`): the renderer's `buildSectionDefaults` began computing these values automatically from palette luminance (WCAG-based), removing the manual per-component declaration burden. Live (`003_contracts_and_standards(8).md`): the "Section painting contract" — `is_dark_section` is demoted to inert catalogue metadata ("MUST NOT key styling"), and any section that paints its own background must instead RE-EXPORT `--section-*` as references to theme tokens via one of four models (pair band, palette band, image/ink-derived, ambient/no-background) using `color-mix()`, so colours flip automatically with the site's scheme; literal colours are forbidden and mechanically enforced by `fix_forced_text_colors`.
- **what:** Documents the multi-year hardening of how section backgrounds get correctly-coloured text: from ad hoc inline colours, to a hardcoded-value contract gated on a boolean flag (which locked every dark section into literal white-on-dark), through a renderer-side automation step, to the current token-referencing "painting" model that treats `is_dark_section` as inert metadata and derives colours mechanically from the active palette.
- **sources:** old/older1/003_contracts_and_standards.md; old/older1/003_contracts_and_standards_v7.md#"Section Context Variable Contract (Dark Sections)"; debugging_old/003_contracts_and_standards_v11.md#"Section Context Variable Contract (Dark Sections)"; archive_april_26/003e_contracts_and_standards_v5.md#"Section Context Variable Contract"; old_design_and_styling/FOCUS_design_and_styling.md#"CSS Colour Inheritance Model"; old_design_and_styling/PHASE_4_4_cleanup_summary.md#"Phase 4.5"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Section Context Variable Contract (Painting Sections)"
- **relations:** styling-render-pipeline (036); design-composition (site-design-planner palette resolution feeds these tokens); fix_forced_text_colors action
- **verify-later:** grep deployed component templates for literal `#ffffff`/`rgba(255,255,255,` inside `--section-*` declarations to confirm the old hardcoded pattern is gone; inspect `fix_forced_text_colors` and `buildSectionDefaults` Go source.

<!-- SOURCE: U12_docs024_archives.md -->
### Component Quality Contract (scoring formula, quality columns)
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** Introduced in `old/older1/003_contracts_and_standards_v6.md` ("## Component Quality Contract"); absent from v7 onward and absent from live `003_contracts_and_standards(8).md` (no "Component Quality Contract" heading), though `quality_score`/`quality_issues` fields still appear inline in the live doc's "Component Creation & Regeneration Contract" JSON examples.
- **what:** v6 fully specified a quality-tracking contract for `content_components`: eight quality columns (`template_variable_count`, `schema_field_count`, `template_closed`, `schema_template_synced`, `has_data_component`, `quality_score` 0-100, `quality_checked_at`, `quality_issues`), a scoring formula starting at 100 with fixed deductions per violation, three computation triggers (on-insert, periodic audit by `component-quality-auditor`, targeted rescan), an automatic `needs_component_regeneration` work item below score 50, and planner preference for higher-scored components. This entire standalone contract section vanished between v6 and v7 and was never restored, even though the live system-architecture doc still lists a `component-quality-auditor` agent and the live contracts doc still surfaces `quality_score`/`quality_issues` as return-payload fields — suggesting the mechanism may partly persist in code/DB while its dedicated documentation disappeared.
- **sources:** old/older1/003_contracts_and_standards_v6.md#"Component Quality Contract"; docs024_key_docs_latest/003_contracts_and_standards(8).md (residual field mentions); docs024_key_docs_latest/002_system_architecture(4).md (component-quality-auditor agent row)
- **relations:** StoreGeneratedComponentAction / component regeneration contract; component-quality-auditor agent
- **verify-later:** check whether `compute_component_quality`/`ScoreAndPersistComponent` and the `content_components` quality columns still exist and are actively populated.

<!-- SOURCE: U12_docs024_archives.md -->
### `query.{name}` field-source resolution timing (render-time → plan-time)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** Archive v7 "Source prefixes" table: `query.{name}` resolved "At render time." Live table: resolved "At `plan_sections` time."
- **what:** In the Component Input Schema v2 Contract's source-prefix table, the `query.{name}` prefix (used for blog posts/categories lists) moved from being resolved at page-render time to being resolved earlier, during `plan_sections`, with the result projected into the field's declared shape. Consistent with the broader shift toward front-loading data-availability checks (the "Layer 0" pre-generation triage) rather than discovering missing/stale query data only at render.
- **sources:** old/older1/003_contracts_and_standards_v7.md#"Source prefixes"; docs024_key_docs_latest/003_contracts_and_standards(8).md#"Source prefixes"
- **relations:** Component Input Schema v2 Contract; plan_sections / Layer 0 pre-generation data triage; page_rerender item_type
- **verify-later:** check plan_sections Go action for query-prefix handling, confirm it projects results to field shape at plan time.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component JS/CSS contract (contract 003): no inline script, split js_content vs js_snippets
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** "Migration B implements all three for the news components" — inline IIFE extracted to js_content, html_template gets `<script src>`, shared formatNewsDate moved to js_snippets
- **what:** Doc 003's contract: no inline `<script>` blocks in `html_template`; component-specific JS lives in `content_components.js_content` served at `/tools/assets/{function}.js`; shared utilities used by multiple components live in `js_snippets`. Also covers component CSS scoping (no bare element rules, dark-section CSS vars with fallbacks) and idempotent migrations.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#003-Contracts-and-Standards, js_snippets_news_gaswholesalers/old/guidelines_compliance_check(1).md
- **relations:** js_snippets site-level rendering pipeline; files_field deploy bug
- **verify-later:** grep/inspect `<script>`; `html_template`; `content_components.js_content`

<!-- SOURCE: U15_docs019_running_notes.md -->
### JS content separation contract
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** "js_content RESOLVED — the assets-split EXISTS (003 'JS Content Separation Contract'...)" (principles(59), 2026-06-11).
- **what:** For interactive components, `store_generated_component_action.go`'s `separateInlineJS()` extracts inline `<script>` bodies into a separate `content_components.js_content` column, replacing them in `html_template` with a `<script src="/tools/assets/{function}.js">` reference; `RerenderSinglePageAction.collectJSAssets()` assembles the resulting multi-file git commit. Verified NOT used by the library-tool pipeline (tool-generator/improver mandate one inline script; the fork INSERT omits `js_content`), creating a landmine where a library tool adopting the split would fork with a dangling script reference and a 404'ing asset.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 "js_content RESOLVED" and "VERIFIED... 019's 'isn't built yet'" entries.
- **relations:** Tool-doc header system; fork-divergence detection.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Result-contract dead-key class and Option A unification
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** NNN_fix_diagnose_complete_output_fields (run 17933a83, Kafka Message Size Too Large at 1.27MB); HANDOFF_builder_thread "Option A CLOSED: shared result contract (datahelpers/result_contract.go), response size guard, four agents on preferred result_from; deployed v1.0.1092+".
- **what:** CompleteWorkflowAction honoured output_field/output_fields and otherwise shipped the ENTIRE collected_data; `result_from` was a key the action never read, so diagnose completions always shipped everything (masked until the 515-file analysis blew the Kafka cap). Same class: the orchestrator pointed output_fields at an imagined key ("diagnose-agent_result") when the engine stores a call step's response under the STEP NAME. Fixes: point at real keys; then Option A — a shared result contract with both readers honouring result_from/output_fields plus a response size guard; NNN_rename_complete_keys_preferred moves the four diagnose/index agents to the preferred key once that image is deployed. Standing rule made mechanical: keep workflow variable names in sync with what actions read.
- **sources:** NNN_fix_diagnose_complete_output_fields.sql; NNN_fix_orchestrator_complete_key.sql; NNN_rename_complete_keys_preferred.sql; HANDOFF_builder_thread.md#2
- **relations:** loop-back plumbing fault class; bounded bundle egress (persist-and-reference)
- **verify-later:** datahelpers/result_contract.go; extractFinalResult size guard; census query for deprecated keys

<!-- SOURCE: U16_docs019_design_plans.md -->
### Adapter response envelope contract (single-sourced)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration(19) 2026-06-10/11: "Envelope contract resolved empirically … The contract now lives once in 035_adapter_guide.md §1 (FOCUS_adapter_design fully merged … and retired)".
- **what:** Resolved from code, not docs: the coordinator claims awaited requests on in_response_to_request_id first (request_id fallback); working adapters use typed body headers with real booleans + ProduceWithValidation; every adapter reads `action` from the BODY; a reply without the right Kafka headers silently falls through to process-as-work and times out (the documented thunder fault — found and fixed in the analyser adapter before deploy). Import-reuse verdict: reuse canonical types for the body, add a local Kafka-header builder (canonical ResponseHeaders lacks request_id/message_id/ToKafkaHeaders). A 003-vs-FOCUS documentation contradiction was settled empirically and single-sourced.
- **sources:** PLAN_workflows_and_actions_migration(19).md (guideline audit + dispatcher fix + envelope resolution entries)
- **relations:** analyser adapter; doc-drift (docs contradicted; code decided); 035_adapter_guide (canonical home, another unit)
- **verify-later:** 035_adapter_guide.md §1; whether 003 §832-890 was replaced with the pointer (was PENDING)

<!-- SOURCE: U16_docs019_design_plans.md -->
### Chassis conventions verified: text+CHECK, previous_version_id, deleted_at
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS_schema_verification_findings §1 — read off the live schemas; applied as corrections across the contract set.
- **what:** The live chassis conventions every new table must follow: enumerated values are text with CHECK constraints (never native enums); versioning is version integer + previous_version_id uuid self-FK with unique (type,version); soft delete is deleted_at (never a status=archived); timestamptz defaults now(); jsonb for flexible payloads. The verification pass also corrected wrong reuse assumptions in the contracts ("the contracts are corrected to match reality, not the reverse") and confirmed real fields: approval_mode, pipeline, item_key, briefing_questionnaire, input/output_contract, agent_category CHECK set.
- **sources:** FOCUS_schema_verification_findings.md; PLAN_active_config_schema(3).md#2 note
- **relations:** schema-before-SQL discipline; all six contracts
- **verify-later:** n/a (verified from live schema dumps)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Priority profile (order not weights; sealed constraints)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_salience(4) §9 "Representation: an order, not numeric weights … A node stores only its differences from what it inherits … computed on demand"
- **what:** Requirement-relative priority among dimensions (security/speed/simplicity/generality/functionality/cost) lives on the objective-tree node as an *order* (with sealed/constraint flags), stored as differences-from-inherited and computed on read. Sealed constraints are ancestor-wins legal floors; a change triggers targeted re-validation of descendants holding conflicting overrides. The open crux (§9.7) is choosing the entry node.
- **sources:** ED/FOCUS_salience_and_multi_author_mediation(4).md#9, ED/FOCUS_salience_and_multi_author_mediation(4).md#9.7
- **relations:** why-chain; mediator; drift detection; direction-of-travel
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Atomic standard (generated-views doc tree)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** FOCUS_best_practice_doc_tree(1) §2 "Optimal unit: the atomic standard, not the document … Documents are generated views over the atoms"
- **what:** The smallest addressable unit is one rule-atom with structured frontmatter (id, concern, scope, applies_to, kind, severity, status, version, supersedes, owner, check, related) and a body split into rule/rationale/examples. Constitution, per-concern handbooks, change-type bundles, and the machine manifest are all *generated views* over one source, so nothing drifts between a doc copy and an agent copy.
- **sources:** ED/FOCUS_best_practice_doc_tree(1).md#2, ED/FOCUS_best_practice_doc_tree(1).md#4, ED/FOCUS_best_practice_doc_tree(1).md#5
- **relations:** mediator routing model (the atom fields are the routing table); doc-tree adoption; concern curators
- **verify-later:** proposed `standards` table

<!-- SOURCE: U18_sql_for_agents.md -->
### Input/output contracts on agent definitions
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Contract UPDATE statements across 011/022/024/025/029 etc.; 129 extends contracts as live behaviour ("input_mapping must satisfy the input_contract — 016b §9 spawn+call rule").
- **what:** Every agent row carries `input_contract` (required/optional fields) and `output_contract` (produces). Contracts are both documentation and runtime validation hooks; the 2026-07 diagnosis work established the durable rule that an input the workflow reads must be declared in the contract (137's "spec is UNDECLARED" fix) and that call-site input_mapping must satisfy the callee's contract.
- **sources:** 011_site_deployer.sql; 022_site_planner.sql; 129_wire_diagnosis_subject_threading.sql; 137_recreation_spec_and_note_subject.sql; sql_for_agents_v1/009_all
- **relations:** remove-loops plan; workflow_contract_chain view (v1/010)
- **verify-later:** chassis contract-validation code path; how strictly contracts fail fast in production

<!-- SOURCE: U18_sql_for_agents.md -->
### Call metadata vs response-data convention (output_field.response)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** v2/001_general_rule.md states it as "The general rule going forward"; 116 confirms it verified in coordinator.go ("a step result is stored under BOTH the step name AND its output_field, adapter body under .response").
- **what:** Workflow data-shape convention: when a step calls another agent, call metadata (agent_id, request_id, topics) is stored directly at the step's output_field while the called agent's response payload lands at `output_field.response`. Many prompt-template and field-path bugs in this directory trace to violating this shape.
- **sources:** sql_for_agents_v2/001_general_rule.md; 116_thunder_training_monitor_worker.sql; 003_site_classifier.sql (classification.response.result paths)
- **relations:** template field-path rules (134); input_mapping
- **verify-later:** coordinator.go result-storage code (~L1636/L2408 per 116)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component render modes (template | agent | composite | standalone)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Columns and comments exist (004 PART 2), but the 000 backup shows all 41 components render_mode='template'; only 'standalone' (tools) is additionally observed in seeds.
- **what:** render_mode declares how a component is produced: 'template' (direct substitution), 'agent' (spawn agent_type with optional agent_workflow, data pulled via data_sources dot-paths), 'composite' (child_components list), and later 'standalone' for tools (html_template IS the final output). The agent/composite modes appear designed but unexercised.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART2; docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component library; standalone tool render.
- **verify-later:** Go render path switch on render_mode; any rows with render_mode in ('agent','composite').

<!-- SOURCE: U19_sql_tables_components.md -->
### Kebab-case naming contract (component function + pages.page_type)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** chk_function_kebab_case CHECK in the live dump; migration 051 adds chk_page_type_kebab_case with pre/post distribution audit; data-component attributes reconciled to function.
- **what:** Identifier-shaped values are kebab-case, enforced by CHECK constraints: content_components.function (regex `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`, empty allowed for legacy), pages.page_type (same regex, snake rows migrated: blog_post→blog-post etc.). data-component attributes in templates must equal function; a partial unique index enforces one active component per function. Also separates page NAME 'index' from page TYPE 'landing'.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-naming-standardisation; docs/agent_docs/sql_for_tables/003_pages.sql#051_pages_page_type_kebab; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql
- **relations:** contracts doc 003/042 naming contract; query-resolver list components (rely on page_type values).
- **verify-later:** pg_constraint rows chk_function_kebab_case, chk_page_type_kebab_case; idx_cc_tool_function_unique.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tier D items-array component schema shape
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Migration 041 hand-writes tool-list to Tier D; 042 queues guide-list regeneration through the pre-store validator; game-list rewrite mirrors tool-list, "field vocabulary IDENTICAL to tool-list".
- **what:** List components must declare a single `items` array with a sub-schema (title, url, meta_description, nav_label) plus top-level fields (eyebrow_label, section_heading, section_intro, cta_url, cta_label, card_link_label), replacing the legacy numbered-flat anti-pattern (guide_1_url…guide_6_url) that broke sites with fewer items. A pre-store validator enforces the structural contract on LLM-regenerated components; rejections land in agent_error_log.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-041 and #migration-042 and #game-list-rewrite
- **relations:** query-resolver list components; component naming contract; agent_error_log.
- **verify-later:** pre-store validator code; tool-list/guide-list/game-list current schemas.

<!-- SOURCE: U19_sql_tables_components.md -->
### Template syntax unification and three-way field alignment
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Long sequence of applied fixes in 005: Handlebars {{#each}}/{{#if}} → Go {{range}}/{{if}}, missing-dot placeholders ({{logo_text}} → {{.logo_text}}), and the "<no value>" root-cause fix aligning LLM prompt output / template fields / input_schema.
- **what:** Templates are Go text/template; a large family of patches converted early Handlebars-style seeds and fixed the recurring three-way mismatch where the LLM prompt, the template field names, and the input_schema disagreed (headline vs title vs section_title; features[].name vs services[].title). Render-context vocabulary standardised (nav_items_html, services_html, footer_nav_html, cta_text, logo_text, company_name).
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#templating-fixes and #fix-no-value; docs/agent_docs/sql_for_tables/012_site_components.sql#replace_template_var
- **relations:** component library; component-based headers; Tier D shape (later formalisation).
- **verify-later:** remaining Handlebars syntax in content_components; render context builder in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Query-resolver list components (pages_where_type) and canonical section URLs
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** gamesdesign migrations re-type guide pages to page_type='guide' so guide-list (items.source = query.pages_where_type:guide) resolves them, "mirrors the working game-list / page_type=game precedent"; URL migration to /guides/<slug>/index.html.
- **what:** List components resolve their items dynamically from the pages table by page_type via a query resolver — no template change needed when pages are added. Depends on canonical page typing and the canonical nested URL shape /<section>/<slug>/index.html produced by CanonicalisePage, making tools/games/guides structural peers.
- **sources:** docs/agent_docs/sql_for_tables/003_pages.sql#migration_retype_guides_to_guide and #migration_guides_url_to_canonical
- **relations:** kebab naming contract; Tier D shape; site-plan page roles.
- **verify-later:** queryresolve Go code; link_registry sync after URL moves.

<!-- SOURCE: U21_legacy_docs_b.md -->
### data-function contract + intelligent component fallback (P1/P2/P3)
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** docs009/001: "A data-function attribute in the HTML acts as a 'shared contract'... P1 perfect match, P2 good match, P3 generic-text-block — the site always gets built"; superseded by the function/kebab-case + data-component contract (docs017/042) whose GetComponentWithFallback keeps the 3-step fallback.
- **what:** The original decoupling of structure from content: the architect assembles empty containers tagged by function (data-function="problem_statement") so the content pipeline can independently fill them; component lookup degrades gracefully (exact function → similar purpose → generic-text-block) so a build never fails for lack of a component.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-1; docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md#Lookup-Safety-Net
- **relations:** component naming contract (successor); content_components.function; AssembleFromLibraryAction.
- **verify-later:** GetComponentWithFallback in component_library.go; generic-text-block component row.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Recursive component tree ("everything is a component")
- **category:** contracts-and-standards
- **status-signal:** abandoned
- **status-evidence:** docs009/001: "We remove is_container... If the HTML template contains {{.Slot_main}}, it IS a container"; RenderNode recursive algorithm, ghost components (wrapper_tag NULL), slot merging; the shipped system instead uses a flat section list per page with header/footer injection.
- **what:** A radically simplified component model where structure is defined entirely by template placeholders: components declare defined_slots and data_schema; the build plan is itself a component tree the architect walks recursively (RenderNode), handling any nesting depth; themes are just root components; "ghost" components (no wrapper tag) reduce div nesting. Content generation is decoupled by flattening the tree into a content_map of UUID→field requirements. The flat-sections production system never adopted the recursion, though slots re-surface in docs018's slot-based assembly proposal.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-The-Simplification; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#2-The-Recursion-Logic; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Key-Architectural-Principles
- **relations:** slot-based modular assembly (docs018/007); asset bubble-up; content injector pattern.
- **verify-later:** defined_slots column on content_components (expected absent or unused).

<!-- SOURCE: U21_legacy_docs_b.md -->
### render_mode decision matrix (DB template vs agent vs research)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs012/012 decision matrix ("Headers/Footers: template, never generated; FAQ: agent + research; Pricing/Contact: template + brief data") and "The render_mode field on content_components is the key differentiator".
- **what:** Each component declares how its content is produced: render_mode='template' renders directly from brief/render_context data (LLM only fills missing schema fields); render_mode='agent' spawns LLM generation, optionally preceded by the research-agent when needs_research=true. Pure-structure components never touch an LLM; research-backed components always cite. Governs the per-section branch inside the page build loop.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-1; docs012_site_maps_and_components/010_component_and_site_architecture.md
- **relations:** research agent; page-content-writer section loop; "LLM = Agent" principle (every LLM call gets its own agent with research/draft/review).
- **verify-later:** render_mode/needs_research/agent_type/data_sources columns on content_components.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Agent input/output contracts (expects/required/produces)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** docs012/012 Part 4 contract JSON for site-planner/page-content-writer/research-agent; "input_contract/output_contract fields in agent_definitions" cited as the tracking mechanism; docs017/002_standardising exports them for validation.
- **what:** Formal per-agent declarations of expected input fields (with required subset) and produced output shapes, stored on agent_definitions, intended to make cross-agent data flow checkable (workflow validator) and self-documenting. Partially realized; the enforced end of it became ActionInputSpec at the action level.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-4; docs017_legacy_agent_rules_images_design_keydocs/002_standardising_deployment_implementation_plan.md
- **relations:** ActionInputSpec; workflow builder/validator; contracts-and-standards doc 003 (current descendant).
- **verify-later:** input_contract/output_contract columns populated?

<!-- SOURCE: U21_legacy_docs_b.md -->
### Dark-section context variable contract (--section-*)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/043 lists 10 dark components migrated with is_dark_section=true and --section-* vars; four enforcement layers specified with deployment order ("Run 014_section_context_migration.sql...").
- **what:** Any dark-background component must set --section-text/-text-muted/-heading/-surface/-border custom properties on its container; global CSS reads them with light-theme fallbacks. Enforced in depth: DB flag (is_dark_section) + audit queries, Go warnings in RenderComponentAction/SavePageSectionsAction, LLM prompt rules in webdesign-agent and page-content-writer, and periodic SQL audits. Direct ancestor of the current section-contrast model.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/043_section_naming_contract.md
- **relations:** colour inheritance model; component naming contract (companion doc); maintenance audits.
- **verify-later:** is_dark_section column; validate_dark_section.go; current section-contrast implementation.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Component naming contract (function = canonical kebab-case ID)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** docs017/042: "content_components.function is the canonical identifier... DB constraint chk_function_kebab_case... partial unique index on active function"; migration table of renames (social_proof → social-proof, 5 heroes disambiguated).
- **what:** One rule ending a class of chain-breaking bugs: `function` (kebab-case, regex-constrained, unique among active components) identifies a component everywhere — the template's data-component attribute must equal it, page_components.slot_name stores it, planners assign by it, rerenders match by it. NormalizeComponentFunction + 3-step fallback tolerate legacy data; adoption pipelines must translate external names, never import them.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/042_component_naming_contract.md
- **relations:** data-function contract (ancestor); page_components; adoption pipeline mapping; SavePageSections/page-rerender.
- **verify-later:** chk_function_kebab_case constraint and unique index in DB; component_validation.go.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component field-source tiers (A/B/C/D + renderer) and proposed Tier E runtime-feed
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Tiers A–D + renderer confirmed from the component-creator prompt at v1.0.1080 (deployed); Tier E is "proposed, pending decision" (2026-06-29) and gap 2 of the autonomy plan (not built).
- **what:** component-creator's schema contract sources each field from Tier A (voice/llm), B (tunable labels/static+fallback), C (site data — site_specs/site_assets), D (derived lists, query.* resolved at plan time), plus a "renderer" source (JS-filled single value with fallback). There is NO tier for content fetched client-side from a JSON feed at runtime — so regenerating a daily-feed component as-is would wrongly bake a build-time provocation into the template. Proposed Tier E ("feed.{name}"): emit a stable-selector DOM shell + declared DOM contract + (originally) an inline loader following a canonical pattern; the archive build refined this to marker-at-generation + external loader.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 (GAP 2 CONFIRMED); docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNBOOK_phase2_provocation_js(29).md#gap-2
- **relations:** section descriptor; generation-time guards; loader-builder agent. Dropped earlier framing: Gap-2 sub-options "(a) component-creator emits a companion js_snippet" vs "(b) loader snippets are library fixtures" (early runbook versions) — superseded by the Tier-E + loader-builder design.
- **verify-later:** component-creator prompt_template tier section; any feed.* source in input_schemas

<!-- SOURCE: U23_docs_root_vonc.md -->
### Generation-time guards for dynamic components
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 084 result 2026-07-06: "FIRST LIVE VALIDATION of baking the guards in at generation" — has_marker=t, has_inline_script=f on the created component; guards held through the real pipeline end to end.
- **what:** Lessons from the whole thread baked into component GENERATION instead of post-hoc surgery: emit `data-runtime-fill` in the template's section tag at generation (no string-REPLACE marker step); forbid inline `<script>` entirely in dynamic components (extraction-bug class becomes impossible; behaviour lives in the external loader); make header copy llm-sourced (no deferral risk); list entries pure markup (nothing for the resolver to fail on); hidden clone-template item plus a `[data-…-template]{display:none}` author rule (hidden alone loses to author CSS); visible empty state. Declared "the pattern for all future dynamic sections".
- **sources:** docs/SPEC_provocations-archive-list.md#design-decisions; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-084-succeeded; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** Tier E proposal; marker anchoring lesson; hidden-vs-author-CSS lesson; store-path validation
- **verify-later:** component-creator output for any newer dynamic component (marker present at generation?)

<!-- SOURCE: U23_docs_root_vonc.md -->
### CSS variable naming convention (--color-*) + creator prompt STRICT RULE
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Hero template fixed and verified (magenta CTA, dark bg) 2026-06-24/25; component-creator prompt patched with "USE ONLY THESE NAMES" + STRICT RULE, UPDATE confirmed in DB 2026-06-24 ~16:50; library-wide audit complete.
- **what:** System CSS custom properties follow `--color-primary/-secondary/-accent/-background/...` naming; LLM-generated components had emitted `--primary-color`-style names that don't exist in styles.css, so fallback hexes fired (the "brochure-blue" index). Fix: template REPLACE on hero + a component-creator prompt section explicitly prohibiting the wrong names and separating Palette from Layout tokens. Documented exception: `--archetype-color` is intentional per-card tinting with `--color-accent` fallback. All new components inherit the correct names.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:30 + #2026-06-24-~16:50; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** post-025 CSS theme flow; legacy-variable "fossilised render" tell (016b chrome entry)
- **verify-later:** component-creator prompt_template section 7; grep new templates for --primary-color

<!-- SOURCE: U23_docs_root_vonc.md -->
### Sanctioned content-edit paths (content_data is truth; HTML patching rejected)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Doc 003 quoted verbatim in the HANDOFF: "this is why HTML patching was rejected as an edit mechanism"; fix_component_template's remove_element fix type and its header deferral to the section-editor confirmed from file headers; section_editor_actions.go confirmed to exist as Go code.
- **what:** `content_data` is the source of truth for section content; patching `page_components.rendered_html` is a bridge at best (lost on the next re-render) and was explicitly rejected as an edit mechanism. Template changes have designated routes: `fix_component_template_action` fix types (including `remove_element` — "removes HTML elements matching a pattern"), with page-component content changes deferred to the section-editor workflow. The mini-lobby trim's method question — which action edits a template, which re-render propagates it, what a NULL-content_data section does — was deliberately settled by bundle verdict rather than guessed. Fallback when no supported path exists: full-text template UPDATE (never multi-line REPLACE of nested markup), verified by length delta, propagated by a page_rerender item.
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/PLAN_provocation-card(3).md#method-corrected; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-primed; docs/bundle_minilobby_trim(4).sh (header)
- **relations:** two re-render paths; per-tool docs (method correction recorded); section-editor
- **verify-later:** fix_component_template_action.go fix types; section_editor_actions.go component_swap

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### API documentation convention (OpenAPI/Swagger + internal API.md)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Archived docs/_archive/api/API_DOCUMENTATION.md is byte-identical to live docs/API_DOCUMENTATION.md; describes `make swagger`, `make validate-openapi`, per-service `API.md` files, `*_swagger.go` annotations.
- **what:** Two-tier API documentation standard: external customer-facing OpenAPI 3.0 spec at `internal/auth-service/api/openapi.yaml` with swaggo annotations, and internal per-service `API.md` files documenting Kafka topics, DB schemas, env vars. Includes a CI workflow that lints the spec and fails on uncommitted swagger regen.
- **sources:** docs/_archive/api/API_DOCUMENTATION.md#external-api-documentation, #internal-api-documentation
- **relations:** superset context for the public/admin API plans (007b/008b); live counterpart docs/API_DOCUMENTATION.md + docs/api/reference.html
- **verify-later:** internal/auth-service/api/openapi.yaml; Makefile targets swagger/validate-openapi

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Training-data export format (ChatML + metadata sidecar; SFT/DPO negative examples)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4e "Format: ChatML messages with metadata sidecar. Decided 2026-04-22"; iter_0 audit "1,970 total" rows
- **what:** Convention that each training row is a ChatML `messages` array plus an ignored metadata sidecar (source_log_id, agent_type, step_name, orchestration_id, model, export_version). Code fences must be stripped; edge-case "prose instead of JSON" rows are excluded from SFT (they'd teach wrong-shape output) but kept in `llm_call_log` as future DPO "rejected" examples. Schema heterogeneity noted: one (agent_type, step_name) covers hero/minimal-hero/header schemas.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4e, #2.4g (dataset profile)
- **relations:** part of Flywheel A; export_version enables downstream compatibility checks
- **verify-later:** stripMarkdownFromResponse in ai_actions.go; export_version field

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Local-step input resolution: input_mapping dead, key_path for loop items (109b)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 109b SQL header "input_mapping did NOT populate input_data.key for this loop substep … deriving the dataset key 40 times"; NOTES(39) "CORRECTS a load-bearing assumption"
- **what:** A load-bearing chassis contract discovered the hard way: the coordinator only resolves `input_mapping` for `call_agent` and loop fan-out, not for plain local action steps. Local actions (and local loop substeps) must read values via a config key holding a dot-path, resolved by `ExtractActionInputs` Strategy 0 / `resolveTemplateToken` / a `key_path`. Migration 109b fixed `presign_one` to read the loop item via `key_path:"ckpt_key"` (from CollectedData where setLoopVariable puts it) instead of the dead `input_mapping{key:ckpt_key}` that had presigned the dataset key 40×.
- **sources:** phase5/109b_fix_presign_one_loop_item_keypath(1).sql; phase5/NOTES_phase5_training_launcher_running(39).md#2, #update-2026-06-06-2, #8
- **relations:** cause of the "presigns the dataset key" failure signature; distinct from the await race
- **verify-later:** loop_expansion_handler.go setLoopVariable; content_search.go; datahelpers ExtractActionInputs

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Capability gate D5 — requires-backend semantic tag
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** plan(11) D5 "Planner gate (to apply): load_components gains AND NOT (… ? 'requires-backend')"; running_notes SQL revised; marked "Outstanding: apply the planner query change".
- **what:** Gates backend-requiring components off static sites by CLASS, not site-type. Component carries `semantic_tags:["requires-backend"]`; planner load_components excludes them unless opted in via roadmap section_types; site side sets `deploy_config || {"target":"vm","capabilities":["backend"]}`; a later audit check compares placed sections' requires-* tags to site capabilities.
- **sources:** traffic_probe_plan(11).md#decision-5, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection
- **relations:** supersedes the intent-probe site-type gate
- **verify-later:** build-site-planner workflow JSON load_components query

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded D5 — suitable_site_types / "intent-probe" site type gate
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4) "Decision 5 (OPEN) … component carries suitable_site_types:['intent-probe'] … planner's load_components gains AND suitable_site_types = '[]'::jsonb"; running_notes "the invented site type is GONE (suitable_site_types: [])".
- **what:** The earlier D5 formulation invented an `intent-probe` site type and gated via `suitable_site_types`. Dropped after operator feedback that "intent-probe is the wrong label" — the distinguishing feature is the class (has a backend), so the tag-based `requires-backend` gate replaced it.
- **sources:** traffic_probe_plan(4).md#decision-5-open, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight
- **relations:** replaced by requires-backend semantic tag gate (D5)
- **verify-later:** component semantic_tags vs suitable_site_types columns

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Class-level rename (probe → site-engine) and env-var churn
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** running_notes 2026-06-11 "RENAME MAP (every changed name)"; env var `PROBE_DB_PATH → ENGINE_DB_PATH`; then 2026-06-11 store-v2 "env var ENGINE_DB_PATH → ENGINE_DATA_DIR".
- **what:** When the box became the home of the whole backend-site class (not just probes), "probe" defaults were neutralised to class-level names across engine + deploy artifacts: service/user `site-engine`, `/opt/site-engine`, `/var/lib/site-engine`, `/etc/site-engine/site-engine.env`, webroots `/var/www/vm-sites/<d>`, rate zone `engine_rl`, hook `site-engine-deploy`. The DB-path env var was renamed twice: `PROBE_DB_PATH` → `ENGINE_DB_PATH` → `ENGINE_DATA_DIR`; store file `probe_events.json` → `intent_events.json` → (dropped for JSONL). ProbeSearch/ProbeCategory/ProbeFreeText kind constants kept (they name the feature, not the class).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-component-live, traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#changelog
- **relations:** supersedes probe-go naming
- **verify-later:** grep for stale PROBE_/probe_events across artifacts

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /stats endpoint + INTERNAL_API_KEY (stats internal key)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** service(24).go "GET /stats key-gated per-host summary"; runbook(12) env table "INTERNAL_API_KEY gates /stats (sent as X-Internal-Key) … unset → /stats 401"; verified over HTTPS 2026-06-12.
- **what:** `/stats` returns a key-gated per-host summary (visits/events counters), gated by `INTERNAL_API_KEY` sent as header `X-Internal-Key`. Unset key → /stats returns 401. The same key doubles as the read-only capture-export key for /events and /access-digest; on the collector side it is stored in `deploy_config.engine.stats_key` (low sensitivity, one accessor, movable to a secrets table later). The env file (not a shell variable) is the source of truth.
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#2, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** read by CollectIntentEventsAction
- **verify-later:** /etc/site-engine/site-engine.env INTERNAL_API_KEY; deploy_config.engine.stats_key

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### /events export endpoint (P4 collector interface)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "GET /events built + tested … Tests green ×6"; runbook(12) §6 "Export endpoint (P4 collector interface)".
- **what:** `GET /events` streams stored events as key-gated NDJSON oldest-first (original line bytes preserved); params `since` (RFC3339, strictly-after), `host`, `limit` (default 5000). Final line `{"_meta":{count,truncated,server_time}}` aids the collector checkpoint (store max created_at → duplicate-free pulls). Lock-free by design so a big export never blocks live captures; a torn mid-append tail line is skipped and arrives next pull.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-events-export, traffic_probe_runbook(12).md#6
- **relations:** consumed by CollectIntentEventsAction; the pull architecture
- **verify-later:** store.go StreamEvents; nginx /events location

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wrapper-orchestrator requirement finding (001:405-462)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "STRUCTURAL finding … wrapper-orchestrator REQUIRED (001:405-462): the collector is reached via the SCHEDULER … AND does substantive in-chassis work → must NOT run in a shared agent-chassis pod".
- **what:** Because the collector is reached via the generic scheduler entry point AND does substantive work (HTTP to N boxes + multi-row upserts, unbounded as boxes grow), it must not run in a shared agent-chassis pod. Fix = a thin orchestrator that spawns a worker child into its own pod. Also corrected: the scheduler fires ONE message per tick and does NOT fan out pre_query rows, so the collector self-queries and loops, and the scheduled_tasks pre_query is a count>0 GATE.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-d, traffic_probe_running_notes(27).md#2026-06-13-c, traffic_probe_running_notes(27).md#standing-observations
- **relations:** shaped the two-agent topology; scheduled_tasks intent-collection target corrected to orchestrator
- **verify-later:** 001 guide §405-462; scheduled_tasks intent-collection target_topic

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Adapter Response Envelope Contract (003) — conditional, later dropped from plan
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4)/(5)/(6) P4/P5 "If collection runs as a chassis adapter, it MUST follow the Adapter Response Envelope Contract (typed-struct bool headers, reuse request_id, message_id, ProduceWithValidation)"; plan(11) reduces this to a one-line parenthetical.
- **what:** The guidelines-audit flagged that IF P4 collection or the P5 deployer were built as chassis adapters, replies must use a typed-struct envelope — getting it wrong = silent drop until timeout (the documented multi-day thunder fault). Once P4 was redesigned to need no adapter (key-gated HTTPS pull + one local action), the prominent P4/P5 envelope warnings were demoted.
- **sources:** traffic_probe_plan(6).md#phases, traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_plan(11).md#p4
- **relations:** applies only if P5 vmhost adapter is built
- **verify-later:** 003 contracts doc; ProduceWithValidation

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### flat-file constitution (precursor to the `standards` table, scope=constitution)
- **category:** contracts-and-standards
- **status-signal:** aspirational
- **status-evidence:** "This is the flat-file version for the thin slice. Later it becomes the `standards` rows with `scope = constitution`; the content is the same."
- **what:** A single always-on rules document meant to be included in full in every LLM context bundle, distinct from the task-specific 003 contracts which are pulled in only when a task touches them. Covers reuse-before-recreate, structural-over-symptomatic fixes, one-orchestrator-per-agent, no-subworkflows-in-SQL, the snake_case/kebab-case string-naming test, chassis storage conventions (text+CHECK not native enums, version+previous_version_id, deleted_at not status=archived), logging discipline (no `logger.Debug`, always log the orchestration_id), and deployment/namespace facts. Explicitly designed to later graduate from a hand-pasted flat file into database rows.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/thin_slice_constitution.md
- **relations:** contextkit toolchain (above); trust ledger contract (below, references a `standards` lifecycle)
- **verify-later:** whether a `standards` table with `scope='constitution'` was ever actually created

<!-- SOURCE: U25_leopardess_social.md -->
### Section resolvers override content_data on every render
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** HANDOFF §4.4 (2026-07-12): "hero background_image is auto-resolved to /assets/images/hero.jpg (plan_sections_action.go:1338) — you change the FILE"; system-stats section removed because suffixes can't be overridden.
- **what:** Field resolution beats stored instance data: the hero background image is re-resolved to a fixed path every render (content edits can't remove it — replace the file); source:"static" schema fields re-apply their schema fallback every render and cannot be overridden per instance; a forked *section* component does not survive rerender because save_page_sections re-links to the canonical component by function, while a header/footer fork sticks because it is wired via the style collection.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#landmine-14; docs/leopardessconsulting/HANDOFF.md#4; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12
- **relations:** static-source schema field misuse; per-site style fork
- **verify-later:** plan_sections_action.go resolvers; save_page_sections re-link behaviour

<!-- SOURCE: U25_leopardess_social.md -->
### Static-source schema fields force fleet-generic labels/suffixes
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** RUNBOOK_minilobby §0 (2026-07-12): "FIXED … migration 090 — stat_1/2/3_label + cta_label were source='static' … forcing 'Clients Served'/'Satisfaction Rate' … on every render"; leopardess system-stats suffixes (%/ms/+/x) still schema-forced (section removed instead).
- **what:** A recurring defect class: shared components whose label/suffix fields are source:'static' re-apply business-generic fallbacks on every render, so writers can only fill values, producing crossed pairs ("2,767%", "500+ Models / Clients Served") on every site. Fix pattern: flip source static→llm keeping fallback as safety net (migration 090 for content-block-about, 13 pages/5 sites); a suffix-free stats component is a noted platform addition. 
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/090_content_block_stat_labels_llm.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12; docs/leopardessconsulting/HANDOFF.md#8
- **relations:** section resolvers override content_data; shared component library semantics
- **verify-later:** content_components 4e448d51 input_schema; system-stats schema fields

<!-- SOURCE: U25_leopardess_social.md -->
### Component creation contract (the generator's embedded rulebook)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 003d carries the full contract prompt "included in the component-creator agent definition"; generated components (brief-explanation regen, archive-list) conform.
- **what:** The complete contract an LLM must follow when generating a component: scoped `<style>` + `<section class="{function}-section" data-component="{function}">` + optional IIFE script; kebab-case function naming; {{.field}} template variables with a declared input_schema (type/source/required/llm_guidance, sources llm|site_specs.*|site_assets.*|renderer|static); all colours via CSS variables with fallbacks and dark-section --section-* custom properties; scoping and responsive rules; client-side-only JS, no CDN; quality bans (no placeholders, no unrendered variables, no fabricated content). Compiled from docs 003 + 018 + input-schema v2; storage location decision: static in the agent prompt now, knowledge_base + rag_lookup if contracts churn.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Component-creation-contracts
- **relations:** component selector/creator; generation-time guards for dynamic components; contracts-and-standards docs 003/018
- **verify-later:** component-creator agent_definitions prompt_template vs the 003d contract text

<!-- SOURCE: U25_leopardess_social.md -->
### plan_sections deferral semantics (required + unresolvable = silently dropped section)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation 2026-07-02: "required field + source unresolved + on_missing empty → hits default: → shouldDefer=true → section deferred → save_page_sections keeps only ready sections."
- **what:** Schema authoring contract: a required field whose data source cannot resolve (e.g. site_assets.illustration with no asset) defers the whole section out of the page save — the instance is dropped, not errored. required=true + on_missing=skip_field is contradictory (still defers). Fix patterns: make decorative fields optional with skip_field/fallback, and populate shared spec gaps once (site_specs.cta.primary_url served gauntlet-cta + system-stats + brief-explanation). llm-sourced fields can never defer — which is why generation-time guards make header copy llm.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-02, #2026-07-03; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Design-decisions
- **relations:** generation-time guards; illustration asset resolution; component creation contract
- **verify-later:** planSection deferral switch in plan_sections_action.go

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section locking with lock types and expiry (design vs implementation gap)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_LOCKS (investigation): "The columns don't exist in schema… expiry mechanism specced never built" while pass-reset half IS implemented; 004/007 describe lock_type permanent/timed/review as if landed
- **what:** Design: components that pass verification lock; lock_type permanent/timed(default 90d)/review(HITL on expiry) with query filter expansion `(locked_at IS NULL OR lock_expires_at < NOW())`. Reality: only plain locked_at/locked_by exists; auto-lock-on-deploy fires on every dashboard edit, so lock proliferation monotonically shrinks the improvement loop's surface (three documented failure modes). Recommended: timed default for routine edits, permanent opt-in.
- **sources:** 031_LOCKS_should_locks_expire.md; 004#Section Locking; 007#Lock lifecycle
- **relations:** audit pass auto-reset; lock coherence debt
- **verify-later:** lock_type/lock_expires_at columns exist?; discovery query filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock semantics: hard gate for discovery, soft gate for execution, read-only rerender
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 013 Phase 1 ✅ with the four amended checks; behaviour tables
- **what:** Lock means "human controls this", not read-only: edit refreshes locked_at without unlock; unlock is a separate deliberate act. Discovery checks skip locked rows (hard gate); execution agents process explicit items regardless (soft gate); rerender reads everything. locked_by vocabulary: admin/admin-removed/checkpoint (human-only unlock) vs deploy (agents may clear). Three lock levels: component, site component, whole site (site lock stops all automation via LoadWorkItemsAction gate + pre_query filter).
- **sources:** 013#Three Levels of Lock, #How Agents Behave; 031(3)#rules
- **relations:** growth budget; suppression
- **verify-later:** lock_helpers.go; four discovery checks' filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock patterns A/B, Pattern B (pinned) is dead, and lock transfer across plan rebuilds
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031(3) verification 2026-05-19: "Pattern B is unenforced in the current code — treat it as dead"; lock transfer specced for Phase 1 site_plan_directives
- **what:** Pattern A locked_at/locked_by (+partial index) is the dominant per-row pattern (sites, page_components, site_components, site_plan_directives). Pattern B pinned boolean on site_specs was never wired (no reads/writes; every spec write is supersede-then-insert with no guard) — new tables must use A. Lock transfer: only the rewriting agent (write_site_plan) copies locks onto matching new rows by composite key; locked text beats LLM rewrite; unmatched locks drop with a log. Locks and snapshots are orthogonal (prevention vs restore); open question whether revert respects locks.
- **sources:** 031_locks(3).md; 030 Q1 directives schema; 013 (pinned column added Phase 4 — UI-level only)
- **relations:** plan-domain tables; spec pin/propagate UI
- **verify-later:** \d site_specs pinned; write_site_plan lock-transfer code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption faithfulness via 90-day timed locks
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending" (2026-05-19); convergence layer marked [done]
- **what:** Adopted sites stay faithful to source for 90 days then develop normally — enforced as timed locks, not a permanent flag. Deliberately timed despite being user-initiated (a faithful starting point, not a frozen final value — documented so nobody "fixes" it to permanent). Because site_plan_directives are plan-scoped and adoption writes no plan, the lock originates at the FIRST write_site_plan (no-current-plan + pages-exist uniquely identifies adopted first plans): page-scoped preserve directives locked adoption/timed/90d; convergence (ValidateSitePlanAction) preserves whatever the 054 query flags adoption_locked; transferDirectiveLocks carries expiry across re-plans; after expiry everything is a no-op. Coexists with 30-day deploy locks at component scope (different questions, no contention).
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md (whole)
- **relations:** lock policy table; lock transfer; FOCUS_planner_ignores_adopted_state (the duplication this protects against)
- **verify-later:** 053/054 applied; write_site_plan first-plan lock branch; v3_site_actions.go convergence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Lock policy table and the improvable-row predicate
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Approved policy table (with adoption added)" (2026-05-19); filter sweep of "11 locked_at IS NULL callsites" still pending
- **what:** Canonical lock semantics: human-set locks (admin/manual/checkpoint) permanent; auto-locks timed (deploy +30d on page_components; auditors +90d; adoption +90d on plan directives); audit_pending is not a lock. The improvable predicate — `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())` — must replace the 11 bare locked_at checks; CheckComponentLock to gain LockType/LockExpiresAt; expired review locks become needs_lock_review HITL items. Coherence rule: all four Pattern-A tables migrate in one migration, no partial state.
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md#policy, #predicate, #implementation-plan
- **relations:** adoption faithfulness; asset locking; Tension #3 candidate (lock-model coherence debt)
- **verify-later:** the 11 callsites; check_component_lock.go; expired_review_locks check existence

<!-- SOURCE: U05_content_quality_linking.md -->
### page_components locking subsystem + non-functional adoption re-plan window
- **category:** locks
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) schema corrections (lock columns + trigger exist, all NULL on index); running_notes_14(26) 14i: "90-day RE-PLAN window is non-functional".
- **what:** page_components (and assets, site_components, site_plan_directives) carry locked_at/locked_by/lock_type(permanent|timed|review)/lock_expires_at plus a trigger_auto_lock_on_deploy — but observed unlocked in practice on the investigated pages, and 013 doctrine says execution agents process explicit items regardless of locks. The adoption-faithfulness design's 90-day timed re-plan lock is non-functional: transferDirectiveLocks copies only locked_at/locked_by (no type/expiry) and nothing creates the adoption/timed lock — only the first-plan convergence branch works. Open question recorded: does save_page_sections honor locked_at (locking a tool section as a zero-code clobber mitigation — probably not).
- **sources:** HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_14(26).md#part-14h-14i; NOTES(44) open sub-questions
- **relations:** adoption faithfulness convergence; interactive clobber mitigations; locks doc 031/053/054.
- **verify-later:** auto_lock_on_deploy function; transferDirectiveLocks; write_site_plan lock creation.

<!-- SOURCE: U09_adoption.md -->
### Adoption faithfulness via timed locks (90-day window)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Verified landed state (2026-06-05)… 053 schema — applied… 054 partially live (first-plan branch only)… write_site_plan Changes 1-3 — not deployed… Consequence: the 90-day re-plan window is non-functional. The only working faithful↔normal boundary today is the first-plan branch."
- **what:** The faithful first pass after adoption is protected by a timed lock (`locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at=NOW()+90d`) on page-scoped preserve-directives, self-releasing so the site later develops normally. Approved policy table adds `adoption` alongside deploy (+30d) and auditor (+90d) timed locks; human locks stay permanent. As landed, only the first-plan branch works; re-plans within 90 days rely on the LLM "preserve existing pages" prompt, not locks.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, PLAN_lock_coherence.md, HANDOFF_2026-05-25#larger-work
- **relations:** first-plan branch detection; preserve-directive lock origination (pending); lock-model coherence plan; 031_locks docs
- **verify-later:** 053_lock_expiry.sql applied state (lock_type/lock_expires_at on assets, page_components, site_components, site_plan_directives); live `load_existing_pages` query; `transferDirectiveLocks` in write_site_plan_action.go (still copies locked_at/locked_by only?)

<!-- SOURCE: U09_adoption.md -->
### Adoption-side lock origination (superseded design)
- **category:** locks
- **status-signal:** superseded
- **status-evidence:** "REVISED 2026-05-19 after schema check: `site_plan_directives` is plan-scoped… adoption writes pages + specs but not plans or directives. So the lock cannot originate at adoption time" (FOCUS_adoption_faithfulness_via_locks(5)); the old2 base version still describes "Adoption writes a per-page preserve directive… locked locked_by='adoption'".
- **what:** The original design had adoption itself writing locked preserve-directives into site_plan_directives. Superseded because directives are keyed by plan_id and adoption creates no plan; the lock now originates at the planner's first `write_site_plan` (detected by `prevPlanID == uuid.Nil` AND existing pages present). There is no adoption-side Go change.
- **sources:** old2/FOCUS_adoption_faithfulness_via_locks.md, FOCUS_adoption_faithfulness_via_locks(5).md#how-this-drives
- **relations:** replaced by write_site_plan first-plan lock origination
- **verify-later:** confirm no adoption-side directive writes exist

<!-- SOURCE: U09_adoption.md -->
### write_site_plan preserve-directives + lock transfer patch (Changes 1–3)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "write_site_plan Changes 1-3 — not deployed. `transferDirectiveLocks` (verified) still copies locked_at/locked_by only… nothing emits page preserve directives or creates an adoption/timed/+90d lock" (2026-06-05).
- **what:** Three coordinated changes written as a patch doc but never deployed: (1) emit a page-scoped `preserve` directive per plan row; (2) on the first plan after adoption, lock those directives adoption/timed/90d; (3) extend `transferDirectiveLocks` to carry `lock_type` + `lock_expires_at` and skip already-expired timed locks (so expired locks release rather than chain forward). Needed only to protect re-plans within the window; re-prioritised low after the first-plan branch proved sufficient for the faithful first pass.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#implementation, old2/write_site_plan_adoption_patch(1).md
- **relations:** adoption faithfulness via timed locks; lock coherence plan step 2
- **verify-later:** `write_site_plan_action.go` transferDirectiveLocks SELECT/UPDATE column lists

<!-- SOURCE: U09_adoption.md -->
### Lock-model coherence plan (one pattern, one lifecycle column, one predicate, one policy function)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Status: PLAN ONLY — NOTHING IN THIS PLAN HAS BEEN APPLIED… held deliberately" (PLAN_lock_coherence, 2026-05-19).
- **what:** Collapse the accreted lock model: Pattern A everywhere (`locked_by` identity, `lock_type` permanent|timed|review, `lock_expires_at`), one improvable predicate, a single `LockPolicyFor(lockedBy)` policy function; retire Pattern B (`site_specs.pinned`, functionally dead in chassis code but exposed via core-manager pin/unpin HTTP endpoints) and the hard/soft `locked_by` string-switch in `check_component_lock.go` (`IsHard = lock_type=='permanent'`). Also resolves the snapshot×lock interaction (does revert_site_to_snapshot clobber human locks?). A fourth `lock_class` column was considered and dropped as redundant.
- **sources:** PLAN_lock_coherence.md, old2/PLAN_lock_coherence(2).md
- **relations:** 031_locks target model; adoption faithfulness runs on the current model without waiting for this
- **verify-later:** `check_component_lock.go` switch; `site_specs.pinned` column existence; `server.go` HandlePinSpec/HandleUnpinSpec; `\sf take_site_snapshot`/`revert_site_to_snapshot`; the 6 improvable-filter callsites vs locked-row finders (three distinct predicate semantics)

<!-- SOURCE: U10_imagery.md -->
### Asset locking (2A) and hard-vs-soft lock semantics
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "2A — assets.locked_at + locked_by ✅ delivered 2026-05-09"; docstring "v3 (final, applied 2026-05-08)".
- **what:** `assets` gains `locked_at timestamptz` + `locked_by text` + partial index, mirroring `page_components` exactly. Canonical lock model (settled after three docstring iterations): detection via `locked_at IS NULL`; classification hard (admin/admin-removed/checkpoint) vs soft (deploy/manual/auditor names) via `locked_by`; NO time-based expiry exists in production. Human uploads/locked assets are excluded from auditor queries and regeneration.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2A, PLAN_imagery_loop_closure.md#2A, old/README.md
- **relations:** logo permanence (D5) is the first real consumer; timed lock-expiry project (deferred).
- **verify-later:** `check_component_lock.go`; assets table DDL; the store-asset lock guard `WHERE assets.locked_at IS NULL`.

<!-- SOURCE: U10_imagery.md -->
### Timed lock-expiry project (deferred)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Approved policy (2026-05-08): implement timed expiry as a focused future project… Sequenced after the imagery loop work completes."
- **what:** One migration adding `lock_type` + `lock_expires_at` to all four Pattern A tables (page_components, site_components, site_plan_directives, assets); auto-lock writers default from a policy table ('admin' permanent, 'deploy' timed/30, auditor approvals timed/90); ~8–10 callsite filter expansions; CheckComponentLock extended; new `expired_review_locks` discovery check. Restores the rhythm doc 004 v4 designed, of which only the audit-pass-counter-reset half shipped.
- **sources:** old/README.md, STATUS_imagery_2026-05-08.md#Lock-expiry-investigation, PLAN_imagery_loop_closure.md#Decisions
- **relations:** references LOCKS_should_locks_expire.md (outside this unit); asset locking 2A.
- **verify-later:** whether lock_type/lock_expires_at columns exist on any Pattern A table.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Locks — HITL durability across the platform
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_locks(2) "This doc is the canonical reference for lock semantics"; "Tech debt: lock-model coherence (target model) … Status (2026-05-19): the lock model has accreted"
- **what:** Two per-row lock patterns protect human-edited data: Pattern A (`locked_at`+`locked_by`, dominant) and legacy Pattern B (`pinned` boolean on site_specs, don't use for new tables). Every writer must read lock state before writing and preserve it when superseding. A coherence cleanup to three orthogonal columns under the invariant permanent⟺human is recorded as deferred tech debt.
- **sources:** WM/031_locks(2).md#the-two-patterns-in-use, WM/031_locks(2).md#lock-transfer-across-rebuilds, WM/031_locks(2).md#tech-debt-lock-model-coherence-target-model, WM/030_phase1_plan_and_reconciler(4).md#lock-transfer-across-plan-rebuilds
- **relations:** human direction/lock lifecycle (007); adoption faithfulness via locks; site plan directives
- **verify-later:** migration 053; check_component_lock.go; FOCUS_adoption_faithfulness_via_locks.md; PLAN_lock_coherence.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Section/component locking with timed expiry
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 115 header: "the project doc 004 v4 designed and docs 031_locks... approved (2026-05-08). Implemented now (Option A)... This migration is SCHEMA + BACKFILL only. The Go follow-on... lands as separate code changes."
- **what:** Locking is the improvement loop's termination and protection mechanism: verified/human-edited rows get locked_at set; auditors exclude locked rows (086). 115 adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables (page_components, site_components, site_plan_directives, +1) in one transaction for coherence. Policy: admin/manual/checkpoint = permanent; deploy = timed +30d; visual-design-auditor / imagery-quality-auditor / adoption (new, faithful-first-pass) = timed +90d. Unlock predicate: `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`. Go-side sweep of 11 callsites still pending at write time.
- **sources:** 115_locks.sql; 086_visual_design_auditor.sql
- **relations:** adoption faithfulness (FOCUS_adoption_faithfulness_via_locks.md); expired_review_locks discovery check (planned)
- **verify-later:** the 11 `locked_at IS NULL` callsites; CheckComponentLock extension; whether expiry sweep landed

<!-- SOURCE: U19_sql_tables_components.md -->
### Pattern A lock convention (locked_at / locked_by, hard vs soft)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 041 Phase 2A codifies: "a row is locked if locked_at IS NOT NULL. No time comparison... timed expiry is documented design intent (004 v4, 007 v4) but not implemented"; canonical classifier named (check_component_lock.go CheckComponentLock → IsHard).
- **what:** Uniform HITL/agent lock across four tables (page_components, site_components, assets, site_plan_directives — plus site_plan_imagery): locked_at timestamp + locked_by identity. Hard locks ('admin', 'admin-removed', 'checkpoint', 'manual' upload) only humans clear; soft locks ('deploy', auditor names, 'audit-pending') agents may clear when a work item references the row. Discovery skips both; execution skips hard. locked_by vocabulary is convention, not CHECK, to allow new identifiers without migration. A future lock-expiry project would add lock_type/lock_expires_at across all Pattern A tables in one migration.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2A; docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7a; docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#directives
- **relations:** 031_locks.md canonical doc; site-level lock; imagery/directive lock transfer.
- **verify-later:** CheckComponentLock consumers; lock-expiry project status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Site-level lock (sites.locked_at)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "Phase 7: Site-level lock — prevents all automated agent activity" (012 tail); scheduled-task pre_query patched to exclude locked sites (020 site-lock section).
- **what:** locked_at/locked_by on sites acts as a master switch: when set, no automated agent activity (discovery, dispatch, improvement) touches the site. Scheduler pre_queries filter locked sites out of candidate selection.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#site-lock
- **relations:** Pattern A locks; scheduler pre_query gating.
- **verify-later:** all dispatch/discovery entry points honour sites.locked_at.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Auto-lock on deploy (page_components lock trigger)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01: "trigger_auto_lock_on_deploy auto-locks on deploy... lock_type permanent|timed|review"; lock check run pre-rebuild (all 4 index rows unlocked).
- **what:** page_components carries locked_at/lock_type/locked_by with a trigger that auto-locks components on deploy (fires on UPDATE). Operational consequence observed: deployed components MAY be locked, so rebuilds/re-renders must check lock state (a lock could block re-render of a target or protect neighbours); on the vonc index all rows were NULL-locked so the behaviour never actually bit in this corpus.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:25 + #2026-07-01-~13:55; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** locks category (031); save_page_sections (does it honour locks? open question in 016b Part 4)
- **verify-later:** trigger_auto_lock_on_deploy definition; save_page_sections lock handling

<!-- SOURCE: U25_leopardess_social.md -->
### auto_lock_on_deploy trigger and the stillborn strict-mode subsystem
- **category:** locks
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_minilobby (2026-07-09 record): "the strict-mode subsystem it belonged to was stillborn — no Go code reads schema_mode … snapshot columns never created … fired exactly once in the system's history"; dropped via migration 009 with saved reversal.
- **what:** A BEFORE UPDATE trigger stamping schema_mode='strict' + lock fields when a row reached deployed on first_deploy sites. Never functional as designed: save_page_sections INSERTs rows already deployed (trigger never fires), its companion snapshot columns were never created, and nothing reads the lock. It nearly sabotaged the section-editor fix (every edit would have locked its row) and was dropped 2026-07-10 with the function body backed up. schema_mode/strict_mode_trigger columns and the orphaned lock/unlock functions deliberately retained.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-09-the-dropped-trigger; docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.2
- **relations:** locks category (031 lock semantics); build_status defect (the near-collision)
- **verify-later:** trigger absence on page_components; 009_drop_auto_lock_on_deploy.sql; leftover lock functions

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section locking with lock types and expiry (design vs implementation gap)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_LOCKS (investigation): "The columns don't exist in schema… expiry mechanism specced never built" while pass-reset half IS implemented; 004/007 describe lock_type permanent/timed/review as if landed
- **what:** Design: components that pass verification lock; lock_type permanent/timed(default 90d)/review(HITL on expiry) with query filter expansion `(locked_at IS NULL OR lock_expires_at < NOW())`. Reality: only plain locked_at/locked_by exists; auto-lock-on-deploy fires on every dashboard edit, so lock proliferation monotonically shrinks the improvement loop's surface (three documented failure modes). Recommended: timed default for routine edits, permanent opt-in.
- **sources:** 031_LOCKS_should_locks_expire.md; 004#Section Locking; 007#Lock lifecycle
- **relations:** audit pass auto-reset; lock coherence debt
- **verify-later:** lock_type/lock_expires_at columns exist?; discovery query filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock semantics: hard gate for discovery, soft gate for execution, read-only rerender
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 013 Phase 1 ✅ with the four amended checks; behaviour tables
- **what:** Lock means "human controls this", not read-only: edit refreshes locked_at without unlock; unlock is a separate deliberate act. Discovery checks skip locked rows (hard gate); execution agents process explicit items regardless (soft gate); rerender reads everything. locked_by vocabulary: admin/admin-removed/checkpoint (human-only unlock) vs deploy (agents may clear). Three lock levels: component, site component, whole site (site lock stops all automation via LoadWorkItemsAction gate + pre_query filter).
- **sources:** 013#Three Levels of Lock, #How Agents Behave; 031(3)#rules
- **relations:** growth budget; suppression
- **verify-later:** lock_helpers.go; four discovery checks' filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock patterns A/B, Pattern B (pinned) is dead, and lock transfer across plan rebuilds
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031(3) verification 2026-05-19: "Pattern B is unenforced in the current code — treat it as dead"; lock transfer specced for Phase 1 site_plan_directives
- **what:** Pattern A locked_at/locked_by (+partial index) is the dominant per-row pattern (sites, page_components, site_components, site_plan_directives). Pattern B pinned boolean on site_specs was never wired (no reads/writes; every spec write is supersede-then-insert with no guard) — new tables must use A. Lock transfer: only the rewriting agent (write_site_plan) copies locks onto matching new rows by composite key; locked text beats LLM rewrite; unmatched locks drop with a log. Locks and snapshots are orthogonal (prevention vs restore); open question whether revert respects locks.
- **sources:** 031_locks(3).md; 030 Q1 directives schema; 013 (pinned column added Phase 4 — UI-level only)
- **relations:** plan-domain tables; spec pin/propagate UI
- **verify-later:** \d site_specs pinned; write_site_plan lock-transfer code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption faithfulness via 90-day timed locks
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending" (2026-05-19); convergence layer marked [done]
- **what:** Adopted sites stay faithful to source for 90 days then develop normally — enforced as timed locks, not a permanent flag. Deliberately timed despite being user-initiated (a faithful starting point, not a frozen final value — documented so nobody "fixes" it to permanent). Because site_plan_directives are plan-scoped and adoption writes no plan, the lock originates at the FIRST write_site_plan (no-current-plan + pages-exist uniquely identifies adopted first plans): page-scoped preserve directives locked adoption/timed/90d; convergence (ValidateSitePlanAction) preserves whatever the 054 query flags adoption_locked; transferDirectiveLocks carries expiry across re-plans; after expiry everything is a no-op. Coexists with 30-day deploy locks at component scope (different questions, no contention).
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md (whole)
- **relations:** lock policy table; lock transfer; FOCUS_planner_ignores_adopted_state (the duplication this protects against)
- **verify-later:** 053/054 applied; write_site_plan first-plan lock branch; v3_site_actions.go convergence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Lock policy table and the improvable-row predicate
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Approved policy table (with adoption added)" (2026-05-19); filter sweep of "11 locked_at IS NULL callsites" still pending
- **what:** Canonical lock semantics: human-set locks (admin/manual/checkpoint) permanent; auto-locks timed (deploy +30d on page_components; auditors +90d; adoption +90d on plan directives); audit_pending is not a lock. The improvable predicate — `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())` — must replace the 11 bare locked_at checks; CheckComponentLock to gain LockType/LockExpiresAt; expired review locks become needs_lock_review HITL items. Coherence rule: all four Pattern-A tables migrate in one migration, no partial state.
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md#policy, #predicate, #implementation-plan
- **relations:** adoption faithfulness; asset locking; Tension #3 candidate (lock-model coherence debt)
- **verify-later:** the 11 callsites; check_component_lock.go; expired_review_locks check existence

<!-- SOURCE: U05_content_quality_linking.md -->
### page_components locking subsystem + non-functional adoption re-plan window
- **category:** locks
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) schema corrections (lock columns + trigger exist, all NULL on index); running_notes_14(26) 14i: "90-day RE-PLAN window is non-functional".
- **what:** page_components (and assets, site_components, site_plan_directives) carry locked_at/locked_by/lock_type(permanent|timed|review)/lock_expires_at plus a trigger_auto_lock_on_deploy — but observed unlocked in practice on the investigated pages, and 013 doctrine says execution agents process explicit items regardless of locks. The adoption-faithfulness design's 90-day timed re-plan lock is non-functional: transferDirectiveLocks copies only locked_at/locked_by (no type/expiry) and nothing creates the adoption/timed lock — only the first-plan convergence branch works. Open question recorded: does save_page_sections honor locked_at (locking a tool section as a zero-code clobber mitigation — probably not).
- **sources:** HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_14(26).md#part-14h-14i; NOTES(44) open sub-questions
- **relations:** adoption faithfulness convergence; interactive clobber mitigations; locks doc 031/053/054.
- **verify-later:** auto_lock_on_deploy function; transferDirectiveLocks; write_site_plan lock creation.

<!-- SOURCE: U09_adoption.md -->
### Adoption faithfulness via timed locks (90-day window)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Verified landed state (2026-06-05)… 053 schema — applied… 054 partially live (first-plan branch only)… write_site_plan Changes 1-3 — not deployed… Consequence: the 90-day re-plan window is non-functional. The only working faithful↔normal boundary today is the first-plan branch."
- **what:** The faithful first pass after adoption is protected by a timed lock (`locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at=NOW()+90d`) on page-scoped preserve-directives, self-releasing so the site later develops normally. Approved policy table adds `adoption` alongside deploy (+30d) and auditor (+90d) timed locks; human locks stay permanent. As landed, only the first-plan branch works; re-plans within 90 days rely on the LLM "preserve existing pages" prompt, not locks.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, PLAN_lock_coherence.md, HANDOFF_2026-05-25#larger-work
- **relations:** first-plan branch detection; preserve-directive lock origination (pending); lock-model coherence plan; 031_locks docs
- **verify-later:** 053_lock_expiry.sql applied state (lock_type/lock_expires_at on assets, page_components, site_components, site_plan_directives); live `load_existing_pages` query; `transferDirectiveLocks` in write_site_plan_action.go (still copies locked_at/locked_by only?)

<!-- SOURCE: U09_adoption.md -->
### Adoption-side lock origination (superseded design)
- **category:** locks
- **status-signal:** superseded
- **status-evidence:** "REVISED 2026-05-19 after schema check: `site_plan_directives` is plan-scoped… adoption writes pages + specs but not plans or directives. So the lock cannot originate at adoption time" (FOCUS_adoption_faithfulness_via_locks(5)); the old2 base version still describes "Adoption writes a per-page preserve directive… locked locked_by='adoption'".
- **what:** The original design had adoption itself writing locked preserve-directives into site_plan_directives. Superseded because directives are keyed by plan_id and adoption creates no plan; the lock now originates at the planner's first `write_site_plan` (detected by `prevPlanID == uuid.Nil` AND existing pages present). There is no adoption-side Go change.
- **sources:** old2/FOCUS_adoption_faithfulness_via_locks.md, FOCUS_adoption_faithfulness_via_locks(5).md#how-this-drives
- **relations:** replaced by write_site_plan first-plan lock origination
- **verify-later:** confirm no adoption-side directive writes exist

<!-- SOURCE: U09_adoption.md -->
### write_site_plan preserve-directives + lock transfer patch (Changes 1–3)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "write_site_plan Changes 1-3 — not deployed. `transferDirectiveLocks` (verified) still copies locked_at/locked_by only… nothing emits page preserve directives or creates an adoption/timed/+90d lock" (2026-06-05).
- **what:** Three coordinated changes written as a patch doc but never deployed: (1) emit a page-scoped `preserve` directive per plan row; (2) on the first plan after adoption, lock those directives adoption/timed/90d; (3) extend `transferDirectiveLocks` to carry `lock_type` + `lock_expires_at` and skip already-expired timed locks (so expired locks release rather than chain forward). Needed only to protect re-plans within the window; re-prioritised low after the first-plan branch proved sufficient for the faithful first pass.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#implementation, old2/write_site_plan_adoption_patch(1).md
- **relations:** adoption faithfulness via timed locks; lock coherence plan step 2
- **verify-later:** `write_site_plan_action.go` transferDirectiveLocks SELECT/UPDATE column lists

<!-- SOURCE: U09_adoption.md -->
### Lock-model coherence plan (one pattern, one lifecycle column, one predicate, one policy function)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Status: PLAN ONLY — NOTHING IN THIS PLAN HAS BEEN APPLIED… held deliberately" (PLAN_lock_coherence, 2026-05-19).
- **what:** Collapse the accreted lock model: Pattern A everywhere (`locked_by` identity, `lock_type` permanent|timed|review, `lock_expires_at`), one improvable predicate, a single `LockPolicyFor(lockedBy)` policy function; retire Pattern B (`site_specs.pinned`, functionally dead in chassis code but exposed via core-manager pin/unpin HTTP endpoints) and the hard/soft `locked_by` string-switch in `check_component_lock.go` (`IsHard = lock_type=='permanent'`). Also resolves the snapshot×lock interaction (does revert_site_to_snapshot clobber human locks?). A fourth `lock_class` column was considered and dropped as redundant.
- **sources:** PLAN_lock_coherence.md, old2/PLAN_lock_coherence(2).md
- **relations:** 031_locks target model; adoption faithfulness runs on the current model without waiting for this
- **verify-later:** `check_component_lock.go` switch; `site_specs.pinned` column existence; `server.go` HandlePinSpec/HandleUnpinSpec; `\sf take_site_snapshot`/`revert_site_to_snapshot`; the 6 improvable-filter callsites vs locked-row finders (three distinct predicate semantics)

<!-- SOURCE: U10_imagery.md -->
### Asset locking (2A) and hard-vs-soft lock semantics
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "2A — assets.locked_at + locked_by ✅ delivered 2026-05-09"; docstring "v3 (final, applied 2026-05-08)".
- **what:** `assets` gains `locked_at timestamptz` + `locked_by text` + partial index, mirroring `page_components` exactly. Canonical lock model (settled after three docstring iterations): detection via `locked_at IS NULL`; classification hard (admin/admin-removed/checkpoint) vs soft (deploy/manual/auditor names) via `locked_by`; NO time-based expiry exists in production. Human uploads/locked assets are excluded from auditor queries and regeneration.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2A, PLAN_imagery_loop_closure.md#2A, old/README.md
- **relations:** logo permanence (D5) is the first real consumer; timed lock-expiry project (deferred).
- **verify-later:** `check_component_lock.go`; assets table DDL; the store-asset lock guard `WHERE assets.locked_at IS NULL`.

<!-- SOURCE: U10_imagery.md -->
### Timed lock-expiry project (deferred)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Approved policy (2026-05-08): implement timed expiry as a focused future project… Sequenced after the imagery loop work completes."
- **what:** One migration adding `lock_type` + `lock_expires_at` to all four Pattern A tables (page_components, site_components, site_plan_directives, assets); auto-lock writers default from a policy table ('admin' permanent, 'deploy' timed/30, auditor approvals timed/90); ~8–10 callsite filter expansions; CheckComponentLock extended; new `expired_review_locks` discovery check. Restores the rhythm doc 004 v4 designed, of which only the audit-pass-counter-reset half shipped.
- **sources:** old/README.md, STATUS_imagery_2026-05-08.md#Lock-expiry-investigation, PLAN_imagery_loop_closure.md#Decisions
- **relations:** references LOCKS_should_locks_expire.md (outside this unit); asset locking 2A.
- **verify-later:** whether lock_type/lock_expires_at columns exist on any Pattern A table.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Locks — HITL durability across the platform
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_locks(2) "This doc is the canonical reference for lock semantics"; "Tech debt: lock-model coherence (target model) … Status (2026-05-19): the lock model has accreted"
- **what:** Two per-row lock patterns protect human-edited data: Pattern A (`locked_at`+`locked_by`, dominant) and legacy Pattern B (`pinned` boolean on site_specs, don't use for new tables). Every writer must read lock state before writing and preserve it when superseding. A coherence cleanup to three orthogonal columns under the invariant permanent⟺human is recorded as deferred tech debt.
- **sources:** WM/031_locks(2).md#the-two-patterns-in-use, WM/031_locks(2).md#lock-transfer-across-rebuilds, WM/031_locks(2).md#tech-debt-lock-model-coherence-target-model, WM/030_phase1_plan_and_reconciler(4).md#lock-transfer-across-plan-rebuilds
- **relations:** human direction/lock lifecycle (007); adoption faithfulness via locks; site plan directives
- **verify-later:** migration 053; check_component_lock.go; FOCUS_adoption_faithfulness_via_locks.md; PLAN_lock_coherence.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Section/component locking with timed expiry
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 115 header: "the project doc 004 v4 designed and docs 031_locks... approved (2026-05-08). Implemented now (Option A)... This migration is SCHEMA + BACKFILL only. The Go follow-on... lands as separate code changes."
- **what:** Locking is the improvement loop's termination and protection mechanism: verified/human-edited rows get locked_at set; auditors exclude locked rows (086). 115 adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables (page_components, site_components, site_plan_directives, +1) in one transaction for coherence. Policy: admin/manual/checkpoint = permanent; deploy = timed +30d; visual-design-auditor / imagery-quality-auditor / adoption (new, faithful-first-pass) = timed +90d. Unlock predicate: `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`. Go-side sweep of 11 callsites still pending at write time.
- **sources:** 115_locks.sql; 086_visual_design_auditor.sql
- **relations:** adoption faithfulness (FOCUS_adoption_faithfulness_via_locks.md); expired_review_locks discovery check (planned)
- **verify-later:** the 11 `locked_at IS NULL` callsites; CheckComponentLock extension; whether expiry sweep landed

<!-- SOURCE: U19_sql_tables_components.md -->
### Pattern A lock convention (locked_at / locked_by, hard vs soft)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 041 Phase 2A codifies: "a row is locked if locked_at IS NOT NULL. No time comparison... timed expiry is documented design intent (004 v4, 007 v4) but not implemented"; canonical classifier named (check_component_lock.go CheckComponentLock → IsHard).
- **what:** Uniform HITL/agent lock across four tables (page_components, site_components, assets, site_plan_directives — plus site_plan_imagery): locked_at timestamp + locked_by identity. Hard locks ('admin', 'admin-removed', 'checkpoint', 'manual' upload) only humans clear; soft locks ('deploy', auditor names, 'audit-pending') agents may clear when a work item references the row. Discovery skips both; execution skips hard. locked_by vocabulary is convention, not CHECK, to allow new identifiers without migration. A future lock-expiry project would add lock_type/lock_expires_at across all Pattern A tables in one migration.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2A; docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7a; docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#directives
- **relations:** 031_locks.md canonical doc; site-level lock; imagery/directive lock transfer.
- **verify-later:** CheckComponentLock consumers; lock-expiry project status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Site-level lock (sites.locked_at)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "Phase 7: Site-level lock — prevents all automated agent activity" (012 tail); scheduled-task pre_query patched to exclude locked sites (020 site-lock section).
- **what:** locked_at/locked_by on sites acts as a master switch: when set, no automated agent activity (discovery, dispatch, improvement) touches the site. Scheduler pre_queries filter locked sites out of candidate selection.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#site-lock
- **relations:** Pattern A locks; scheduler pre_query gating.
- **verify-later:** all dispatch/discovery entry points honour sites.locked_at.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Auto-lock on deploy (page_components lock trigger)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01: "trigger_auto_lock_on_deploy auto-locks on deploy... lock_type permanent|timed|review"; lock check run pre-rebuild (all 4 index rows unlocked).
- **what:** page_components carries locked_at/lock_type/locked_by with a trigger that auto-locks components on deploy (fires on UPDATE). Operational consequence observed: deployed components MAY be locked, so rebuilds/re-renders must check lock state (a lock could block re-render of a target or protect neighbours); on the vonc index all rows were NULL-locked so the behaviour never actually bit in this corpus.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:25 + #2026-07-01-~13:55; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** locks category (031); save_page_sections (does it honour locks? open question in 016b Part 4)
- **verify-later:** trigger_auto_lock_on_deploy definition; save_page_sections lock handling

<!-- SOURCE: U25_leopardess_social.md -->
### auto_lock_on_deploy trigger and the stillborn strict-mode subsystem
- **category:** locks
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_minilobby (2026-07-09 record): "the strict-mode subsystem it belonged to was stillborn — no Go code reads schema_mode … snapshot columns never created … fired exactly once in the system's history"; dropped via migration 009 with saved reversal.
- **what:** A BEFORE UPDATE trigger stamping schema_mode='strict' + lock fields when a row reached deployed on first_deploy sites. Never functional as designed: save_page_sections INSERTs rows already deployed (trigger never fires), its companion snapshot columns were never created, and nothing reads the lock. It nearly sabotaged the section-editor fix (every edit would have locked its row) and was dropped 2026-07-10 with the function body backed up. schema_mode/strict_mode_trigger columns and the orphaned lock/unlock functions deliberately retained.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-09-the-dropped-trigger; docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.2
- **relations:** locks category (031 lock semantics); build_status defect (the near-collision)
- **verify-later:** trigger absence on page_components; 009_drop_auto_lock_on_deploy.sql; leftover lock functions

<!-- SOURCE: U18_sql_for_agents.md -->
### RAG knowledge base (shared pgvector store) and rag_index/rag_lookup
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** 041 creates knowledge_base (vector(768), nomic-embed-text, content_hash dedup); 105 rag-test-agent verifies chassis registration; 141 finally lands first tool_docs rows after the chunk-loop saga.
- **what:** Shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info and component usage patterns, queryable by any content-creating agent. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Embedding-model column tracks provenance; changing dimensions requires column ALTER + reindex.
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** travelling docs (doc_plans indexed into tool_docs); rag_actions.go; code-indexer (separate code_symbols store)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers

<!-- SOURCE: U19_sql_tables_components.md -->
### knowledge_base RAG store
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** Migration 082 (idempotent) with pgvector + pg_trgm; 048 later confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** Industry/marketing content chunks for retrieval: collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via ollama-adapter) with IVFFlat cosine index, trigram GIN fallback for keyword retrieval when embeddings are unavailable, source tracking, quality_score and usage_count lifecycle, stats view.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** code_symbols (sibling shape); ollama-adapter embedder; content grounding.
- **verify-later:** collections in use; retrieval actions.

<!-- SOURCE: U18_sql_for_agents.md -->
### RAG knowledge base (shared pgvector store) and rag_index/rag_lookup
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** 041 creates knowledge_base (vector(768), nomic-embed-text, content_hash dedup); 105 rag-test-agent verifies chassis registration; 141 finally lands first tool_docs rows after the chunk-loop saga.
- **what:** Shared (not per-agent) embedded knowledge store for scraped exemplar sites, research, curated industry info and component usage patterns, queryable by any content-creating agent. Collections partition use-cases (industry_sites, research, components, tool_docs, flywheel_b_chassis_test). Embedding-model column tracks provenance; changing dimensions requires column ALTER + reindex.
- **sources:** 041_rag_knowledge_base.sql; 105_rag_test_agent.sql; 141_reenable_index_plan_after_chunk_fix.sql
- **relations:** travelling docs (doc_plans indexed into tool_docs); rag_actions.go; code-indexer (separate code_symbols store)
- **verify-later:** knowledge_base row counts per collection; rag_lookup consumers

<!-- SOURCE: U19_sql_tables_components.md -->
### knowledge_base RAG store
- **category:** NEW:rag-retrieval
- **status-signal:** deployed
- **status-evidence:** Migration 082 (idempotent) with pgvector + pg_trgm; 048 later confirms live extension versions on clients_db (vector 0.8.0) and describes knowledge_base as the "proven SHAPE".
- **what:** Industry/marketing content chunks for retrieval: collection + industry + domain classification, content with content_hash dedup per collection, vector(768) embeddings (nomic-embed-text via ollama-adapter) with IVFFlat cosine index, trigram GIN fallback for keyword retrieval when embeddings are unavailable, source tracking, quality_score and usage_count lifecycle, stats view.
- **sources:** docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#082; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#WHY-A-SIBLING
- **relations:** code_symbols (sibling shape); ollama-adapter embedder; content grounding.
- **verify-later:** collections in use; retrieval actions.
