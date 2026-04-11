# Design, Adoption & Site-Level Planning: Work Plan

Date: 2026-04-11 (updated)
References: 001, 003, 007, 014, 021, FOCUS_design_and_styling, FOCUS_navigation_HANDOFF

---

## Decisions Made

1. **`design_reference`** — new spec aspect for adopted sites. Concrete values extracted from crawled HTML by Go action. Replaces the vague `design` spec.

2. **`design_intent`** — already a defined spec aspect (per 021). The adoption pipeline now auto-generates it from `design_reference` via LLM (Phase 2e). Semantic with reference values as guidance, not prescriptive.

3. **Three-way priority in webdesign-agent prompt**: design_intent (creative freedom) → design_reference (reproduce faithfully) → generate from industry (new builds). This determines when the palette can change — locked until design_intent is written.

4. **Audit loop proposes but doesn't apply** design changes directly. It creates work items. Design_intent updates flow through the strategist or human, not the audit loop.

5. **Section recipes** (Phase 4): semantic purpose + reference implementation as guidance.

6. **Every build conceptually an adoption** (Phase 4): new builds start from reference material too.

7. **Site design decomposition** (Phase 3): structure × identity × effects.

8. **Option B: dedicated site-design-planner agent** (Phase 3).

9. **Design agent write-back** goes to css_themes metadata, not site_specs (per 021 convention).

---

## Status

### Phase 0: Nav Fixes
| # | Task | Status |
|---|------|--------|
| 0a | `plan_sections` filter — strip header/footer from sections | **Go patch ready** |
| 0b | `InjectHeader`/`InjectFooter` skip-if-present guard | **Go patch ready** |
| 0c | Component-template-fixer idempotency (case-sensitive check) | **Go patch ready** |
| 0d | Trigger rerender for affected sites | Ready (run after Go deploy) |
| 0e | Fix hover colour invisible on dark bg | Needs template inspection |
| 0f | Footer background empty | **Resolved** — was the duplicate footer |
| 0g | `LogoURL` as direct RenderContext field | **Go patches ready** (5 files) |
| 0h | Footer services wrong pages | Data fix (tools in_footer), proper fix Phase 3 |

### Phase 1: Design Fingerprint Extraction
| # | Task | Status |
|---|------|--------|
| 1a | `extract_design_fingerprint` Go action | **Code ready** — needs deploying |
| 1b | Register in action registry | **Instruction ready** — add to registry.go |
| 1c | Insert `extract_fingerprint` step into adoption workflow | ✅ **Applied to DB** |
| 1d | Write `design_reference` spec in apply_adoption_plan | ✅ **Already in live code** |
| 1e | Enrich `needs_design` work item spec with fingerprint | **Go patch ready** |

### Phase 2: Webdesign-Agent Awareness
| # | Task | Status |
|---|------|--------|
| 2a | Load design_reference/design_intent in load_site_for_design | ✅ **Not needed** — read_site_spec already loads all aspects |
| 2b | Update analyze_design LLM prompt (three-way priority) | ✅ **Applied to DB** |
| 2c | Baseline awareness — check if CSS already exists before generating | **Not started** — useful for evolution, not blocking adoption |
| 2d | CSS variable name translation | ✅ **Covered by** fingerprint's `suggested_mapping` in 1a |
| 2e | Generate design_intent from design_reference in adoption workflow | ✅ **Applied to DB** |

### Phase 3: Site-Design-Planner Agent (not started)
| # | Task | Status |
|---|------|--------|
| 3a | Define `navigation` and `layout` spec schemas | Not started |
| 3b | Create `site-design-planner` agent definition | Not started |
| 3c | Wire into build pipeline | Not started |
| 3d | Update `populate_nav_tables` to read navigation spec | Not started |
| 3e | Update header/footer template selection from layout spec | Not started |
| 3f | InjectHeader/Footer reads layout.hero_nav_merged | Not started |
| 3g | Adoption nav extraction from crawl | Not started |

### Phase 4: Requirement-Driven Components (not started)
| # | Task | Status |
|---|------|--------|
| 4a | Section recipe generation during adoption | Not started |
| 4b | Component selector by functional requirement | Not started |
| 4c | `needs_new_component` work items when no match | Not started |
| 4d | Visual identity library | Not started |
| 4e | Effects library | Not started |

---

## What Needs Deploying Now

### Go changes (single deployment):
1. **New file**: `extract_design_fingerprint_action.go` (Phase 1a)
2. **Registry**: Add entry in `registry.go` (Phase 1b)
3. **Patch**: `apply_adoption_plan_action.go` — enrich needs_design work item spec (Phase 1e)
4. **Patch**: `plan_sections_action.go` — filterSiteLevelSections (Phase 0a)
5. **Patch**: `component_library.go` — InjectHeader/Footer skip guard (Phase 0b)
6. **Patch**: `component_library.go` — fixInjectResponsiveCSS case fix (Phase 0c)
7. **Patches**: LogoURL across 5 files (Phase 0g)

### SQL already applied:
- ✅ Phase 1c: Adoption workflow has `extract_fingerprint` step
- ✅ Phase 2b: Webdesign-agent prompt has three-way design priority
- ✅ Phase 2e: Adoption workflow has `generate_design_intent` + `write_design_intent` steps

### SQL to run after Go deploy:
- Phase 0d: Trigger rerenders for ai-agent-orchestration.com and finetuning.uk

---

## Remaining Phase 2 Work

**2c: Baseline awareness** — the webdesign-agent currently generates CSS from scratch every time. For evolution (not first build), it should see the current deployed CSS and make targeted changes rather than regenerating. This requires:
- Loading current CSS from `css_themes` in `load_site_for_design`
- Adding a conditional prompt section: "Here is the current stylesheet. What changes would you make to better match the design intent?"
- This is not blocking for adoption (first build generates from reference). It helps when the improvement loop triggers the webdesign-agent for refinement.

---

## Adoption Workflow (current state)

```
ensure_site_record
  → crawl_site (firecrawl)
  → format_crawl (summaries for LLM)
  → check_crawl_content (conditional)
  → extract_fingerprint (Go — NEW, extracts CSS/fonts/layout from rawHTML)
  → analyze_site (LLM — identity, design, pages, sections)
  → classify_archetype (LLM — site character, constraints)
  → select_content (Go — pick representative pages)
  → derive_content_direction (LLM — writing style guide)
  → apply_plan (Go — write specs, pages, work items)
  → generate_design_intent (LLM — NEW, semantic brief from fingerprint)
  → write_design_intent (write_site_spec — NEW, persists design_intent)
  → complete
```

Then dispatch loop picks up work items:
- `needs_design` → webdesign-agent (now reads design_intent/design_reference)
- `needs_content_page` × N → page-content-writer
- `needs_rerender` → rerender-pages

---

## What Stays the Same

- Page planner, content writer, component library
- site_components injection mechanism
- Git deployment, rerender, audit loop structure
- Strategist, research_results, site_specs versioning
- Snapshot system, content validation, dark section contract
