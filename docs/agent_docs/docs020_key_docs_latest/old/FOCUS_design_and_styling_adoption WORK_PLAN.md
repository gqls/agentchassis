# Design, Adoption & Site-Level Planning: Unified Work Plan

Date: 2026-04-11
References: 001 (Development Guide), 003 (Contracts), 007 (Adoption Pipeline), 014 (Snapshots), 021 (Site Spec & Classifier), FOCUS_design_and_styling, FOCUS_navigation_HANDOFF

---

## Decisions Made & Rationale

### 1. `design_reference` is a new spec aspect for adopted sites

The current adoption pipeline writes a vague `design` spec from the LLM's guess. This is replaced by `design_reference` — concrete values extracted from the crawled HTML by a Go action. Contains hex colours, font families, CSS variables, layout patterns, spacing values, in the original site's terms. Historical record, like `adopted_from` in the identity spec.

The raw crawl data stays in `research_results` (per 007). `design_reference` is a structured extraction from that raw data — a convenience layer so downstream agents don't parse raw HTML.

### 2. `design_intent` already exists — adoption just needs to write one

Per 021, `design_intent` is a defined spec aspect: "style direction, colour mood, typography, imagery, layout." The classifier writes it for new builds. The adoption pipeline doesn't write one at all — that's the core gap. The fix is adding a translation step that derives `design_intent` from `design_reference` during adoption, using the same format the classifier produces.

### 3. design_intent is semantic with reference values as guidance

It describes character ("dark IDE aesthetic, functional not atmospheric") with reference values as starting points ("the reference site used #121212, #00bcd4 — start here"). The webdesign-agent has creative freedom to interpret the intent. This gives the improvement loop room to propose changes that stay within the character.

Contrast with prescriptive: if design_intent specified exact hex codes, the improvement loop would have nothing to evolve. The webdesign-agent would be a template filler, not a designer.

### 4. Section recipes: semantic purpose + reference implementation

When adopting a site, sections are described as recipes. Each recipe has:
- **Purpose**: what it achieves and why it works ("three elevated cards presenting equal-weight choices")
- **Structure**: what elements repeat, how they're arranged
- **Reference implementation**: how the original did it — "use as a guide, not a specification" ("grid-template-columns: repeat(3, 1fr), box-shadow: 0 10px 30px")
- **Component match**: closest existing component and what gaps remain

### 5. Every build is conceptually an adoption

New builds start from a reference too — from the design library (accumulated from past builds) or generated fresh. This diversifies output because each site starts from different reference material rather than always "standard-brochure."

### 6. Site design decomposition: structure × identity × effects

- **Structure**: what sections exist, how content flows (brochure, tool-platform, editorial)
- **Identity**: palette, typography, spacing with character descriptions
- **Effects**: elevation, corners, animation, density

"Brochure" is a structural constraint, not a visual identity. The webdesign-agent combines all three into CSS.

### 7. Option B: dedicated site-design-planner agent

A new agent after classifier, before page planning. Writes `navigation` and `layout` spec aspects. Separate from classifier because the classifier is already doing a lot (identity, industry, audience, content direction). Per 003 orchestrator boundaries: "If changing an agent's internal approach requires updating its parent orchestrator, the boundary is wrong."

### 8. Style collections become less dominant

The site-design-planner makes header/footer/layout decisions individually. Style collections may persist as starting-point suggestions, snapshotted under IDs so each selection is recorded.

### 9. Design agent write-back — reconciling with 021

Doc 021 says "the design agent reads design_intent and produces CSS. It does not write to site_specs." Our proposal has it writing back what it produced so the next run has a baseline. Resolution: the webdesign-agent writes resolved values to `css_themes` metadata or `sites.content_data` — not to site_specs. This preserves the convention that the design agent is a consumer of specs, not a writer. The audit loop compares deployed CSS against `design_intent` (from site_specs). If we need the audit loop to know exactly what was deployed, we extend `css_themes` to store resolved values alongside the CSS text — a build artifact, not a spec.

---

## Alignment with Existing System

### What we're using as-is

| Component | Role | Reference |
|---|---|---|
| `site_specs` with `is_current`/`superseded_at` versioning | All new spec aspects follow this pattern | 021 |
| `write_site_spec` action | Writing `design_reference`, `design_intent`, `navigation`, `layout` | 021 |
| `read_site_spec` action | Loading specs for webdesign-agent, planner, audit loop | 021 |
| `research_results` for raw crawl storage | Raw HTML stays here, `design_reference` is extracted from it | 007 |
| `site_context` schema | Webdesign-agent input format — we enrich it, not replace it | 003 |
| `buildCrawlPageIndex` and `crawlPageContent` struct | Already stores rawHTML per page — fingerprint reads from this | 007 |
| Agent definition SQL conventions | New agents follow the standard | 003 |
| `take_site_snapshot` | Snapshot before any design changes | 014 |
| Component naming contract (kebab-case) | Any new components follow this | 003 |

### What we're modifying

| Component | Change | Why |
|---|---|---|
| `apply_adoption_plan` | Write `design_reference` instead of vague `design` spec. Include fingerprint in `needs_design` work item spec. | Core gap — webdesign-agent currently gets `spec: "{}"` |
| `load_site_for_design` | Also load `design_reference` and `design_intent` from site_specs | Webdesign-agent needs this data |
| `analyze_design` LLM prompt | Conditional section for adoption context with reference values | The prompt currently designs from industry name alone |
| `populate_nav_tables` | Read from `navigation` spec with fallback to current mechanical behaviour | Phase 3 — driven by design decisions not hardcoded defaults |
| `select_style_collection` | Becomes less dominant — site-design-planner can override | Phase 3 |
| Adoption workflow (`site-adoption-agent`) | Insert `extract_fingerprint` step, add `design_intent` generation step | Phases 1-2 |

### What we're creating

| Component | Type | Phase |
|---|---|---|
| `extract_design_fingerprint` | Go action (local, no LLM) | Phase 1 |
| `design_reference` | spec aspect | Phase 1 |
| `navigation` | spec aspect | Phase 3 |
| `layout` | spec aspect | Phase 3 |
| `site-design-planner` | New agent (specialist category) | Phase 3 |
| Design intent generation for adoption | LLM workflow step | Phase 2 |

### Step Zero verification (per 001)

Before creating `extract_design_fingerprint`:
- `site-scraper` agent exists — does single-page scrape + LLM analysis. Different scope (single page, LLM-based, no CSS parsing). Not a replacement.
- `format_crawl_for_analysis` — formats crawl for LLM consumption (markdown summaries). Different purpose. No CSS extraction.
- No existing action parses `<style>` blocks or `<link>` tags from rawHTML for design data.
- `RenderCSSFromSpecAction` — consumes design spec, doesn't extract one. Downstream consumer.
- Decision: New action needed. No existing action covers CSS extraction from crawled HTML.

---

## Work Plan

### Phase 0: Nav Fixes (from handoff — prevents active bugs)

Data fixes already applied. These are code fixes.

| # | Task | File | Size |
|---|------|------|------|
| 0a | `plan_sections` filter — strip header/footer names from sections before content writer | `PlanSectionsAction` | Small |
| 0b | `InjectHeader`/`InjectFooter` skip-if-present guard — `strings.Contains(html, "site-header")` | `component_library.go` | Small |
| 0c | Component-template-fixer idempotency — `strings.ToLower` for "responsive fix" check | `fixInjectResponsiveCSS` | Trivial |
| 0d | Trigger rerender for ai-agent-orchestration.com and finetuning.uk | SQL | Trivial |
| 0e | Fix hover colour #0f3460 invisible on #1a1a2e in header-professional-dark | `content_components` template update | Small |
| 0f | Fix footer `background: ;` empty — primary colour not reaching footer template | Footer template / render context | Small |
| 0g | Fix `logo_url` flow — add as direct field on RenderContext per 003 note | `component_library.go` | Small |
| 0h | Fix `buildServicesHTML` returning wrong pages for footer | Nav/footer query | Small |

### Phase 1: Design Fingerprint Extraction

| # | Task | File | Size |
|---|------|------|------|
| 1a | `extract_design_fingerprint` Go action — parse `<style>` blocks, inline styles, `<link>` tags from rawHTML pages (same page access pattern as `format_crawl_for_analysis`). Extract: hex colours with role classification (bg/text/accent), font families + Google Fonts URLs, CSS variable declarations (design-relevant only), max-width/display/spacing/gap patterns, dark section detection. Uses goquery (already in codebase). | New action file | Medium |
| 1b | Register in action registry + local_actions + ActionInputSpec | `registry.go`, `local_actions.go` | Small |
| 1c | Insert `extract_fingerprint` step into adoption workflow — after `check_crawl_content`, before `analyze_site`. Output field: `design_fingerprint` | SQL update to `agent_definitions` | Small |
| 1d | Update `apply_adoption_plan` — write `design_reference` spec from fingerprint (replaces current vague `design` spec). Uses existing `specAspects` pattern and `write_site_spec` SQL pattern. | `apply_adoption_plan_action.go` | Small |
| 1e | Include fingerprint in `needs_design` work item spec — currently `spec: "{}"`, change to include `adopt_from.design` when fingerprint exists | Same file | Small |

### Phase 2: Webdesign-Agent Awareness

| # | Task | File | Size |
|---|------|------|------|
| 2a | Update `load_site_for_design` — add `design_reference` and `design_intent` to the spec query (currently loads identity, classification, strategy only). Merge into `site_context` following the existing schema from 003. | `load_site_for_design_action.go` | Small |
| 2b | Update `analyze_design` LLM prompt — conditional adoption section: "If design reference is available, honour these reference values. Here are the original palette, fonts, spacing. Generate a design that fits our CSS variable conventions while preserving the character." Keep existing prompt for non-adopted sites. | `agent_definitions` SQL for webdesign-agent | Medium |
| 2c | Baseline awareness — check if CSS already exists for this site (from `css_themes` or work item spec). If yes, prompt shifts from "generate from scratch" to "here's the current stylesheet, here's the design intent, what changes would you make?" | `load_site_for_design` + prompt | Medium |
| 2d | CSS variable name translation — Go function mapping common original variable names to our conventions: `--bg-color` → `--color-background`, `--primary-color` → `--color-primary`. Mechanical, not LLM. Part of fingerprint output or standalone helper. | Part of 1a or helper function | Small |
| 2e | Generate rich `design_intent` for adopted sites — LLM step in adoption workflow after fingerprint extraction. Reads `design_reference` + identity, produces character descriptions + reference values + guidance. Same format the classifier produces for new builds. | New workflow step in adoption pipeline | Medium |

### Phase 3: Site-Design-Planner Agent

New agent following 003 conventions: agent_category = 'specialist', input_contract, output_contract, docker_image/tag.

| # | Task | File | Size |
|---|------|------|------|
| 3a | Define `navigation` and `layout` spec aspect schemas — document expected fields, validation rules | Documentation | Small |
| 3b | Create `site-design-planner` agent definition — LLM prompt reads identity, design_intent, classification, site_archetype. Outputs `navigation` + `layout` specs. For adopted sites, also reads `design_reference` to match original nav/layout. Uses `write_site_spec` action for output. | `agent_definitions` SQL | Medium |
| 3c | Wire into build pipeline — after classifier, before page planning. New-build and adoption flows. | Workflow updates | Medium |
| 3d | Update `populate_nav_tables` — read `navigation` spec: `primary_items`, `tools_strategy`, `max_visible_items`, `cta`. Fallback to current mechanical behaviour when no spec exists. Addresses nav problems N3, N4, N5, N8 from handoff. | `PopulateNavTablesAction` | Medium |
| 3e | Update header/footer template selection — `layout.header_style`, `layout.footer_style` drive template choice. Fallback to style collection when no layout spec exists. | `render_site_components` | Medium |
| 3f | Update `InjectHeader`/`InjectFooter` — read `layout.hero_nav_merged` and `page_overrides` for skip logic | `component_library.go` | Small |
| 3g | Adoption nav extraction — during adoption, extract nav structure from crawl (items, grouping, CTA, footer complexity). Feed into site-design-planner context. | Adoption workflow | Medium |

### Phase 4: Requirement-Driven Components (longer term)

| # | Task | Size |
|---|------|------|
| 4a | Section recipe generation during adoption — LLM decomposes sections into recipes (purpose + structure + reference implementation + component match) | Medium |
| 4b | Component selector by functional requirement — search `content_components` by capability, not by name | Medium-Large |
| 4c | `needs_new_component` work items when no match — recipe becomes the brief | Small |
| 4d | Visual identity library — accumulated palettes/typography/effects from past builds, searchable by purpose/audience | Medium |
| 4e | Effects library — elevation, corners, animation, density as composable modifiers | Medium |

### Nav Bug Fixes (from handoff — addressed across phases)

| # | Problem | When |
|---|---------|------|
| N3 | Tools listed individually in primary nav | Phase 3d (`navigation.tools_strategy`) |
| N4 | Tool labels too long | Phase 3d (label optimisation in nav spec) |
| N5 | Truncated to meaningless "AI Agent" | Phase 0 (fix `rerenderSimplifyNavLabel`) or Phase 3d |
| N8 | max_header_items: 8 too generous | Phase 3d (`navigation.max_visible_items`) |
| N9 | Hover colour invisible | Phase 0e |
| N10 | Responsive CSS injected 4x | Phase 0c |
| N11 | Footer background empty | Phase 0f |
| N12 | Logo missing | Phase 0g |
| N13 | Placeholder email | Data fix (now) |
| N14 | Footer services junk | Phase 0h |

---

## Implementation Order

```
Phase 0 (now — active bug fixes)
  0a-0h: nav fixes from handoff

Phase 1 (next — foundation)
  1a-1e: fingerprint extraction, workflow wiring, spec writing

Phase 2 (then — quality win)
  2a-2e: webdesign-agent uses design data, adoption gets design_intent

Phase 3 (after — structural decisions)
  3a-3g: site-design-planner agent, nav/layout specs, consumer updates

Phase 4 (later — component diversity)
  4a-4e: recipes, requirement-driven selection, identity library
```

---

## What Stays the Same

- Page planner plans pages and sections
- Content writer fills sections
- Component library is the building blocks (grows organically)
- `site_components` table and injection mechanism
- Git deployment pipeline
- Rerender pipeline
- Audit loop structure (now checks against design_intent instead of nothing)
- Strategist writes aspirational direction
- `research_results` stores raw crawl data
- `site_specs` versioning with `is_current`/`superseded_at`
- Snapshot system captures full site state before changes
- Content validation contract
- Dark section CSS variable contract