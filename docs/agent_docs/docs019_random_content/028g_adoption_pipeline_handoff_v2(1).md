# Adoption Pipeline — Handoff Document

Session date: 2026-04-03 (continuation of 2026-04-02, 2026-03-31)

---

## What Was Fixed This Session

### 1. Registry fix: select_representative_content

The Go action file existed but was never registered in `registry.go`. Added the entry. This unblocked the content direction pipeline:

```
crawl → format → analyze → select_content → derive_content_direction → apply_plan
```

Content direction specs now generate with ~17K chars of detailed writing style analysis, stored with `formatted = true` for the content writer to consume via `{{.site_specs.specs.content_direction.formatted}}`.

### 2. Component discovery pipeline unblocked (two bugs)

`CreateNeedsNewComponentItem` in `component_selector.go` had two bugs preventing `needs_new_component` work items from ever being created:

**Bug 1:** Column name `domain` doesn't exist on `site_work_items` — should be `pipeline`. Every INSERT failed with a schema error. The working `createDeferredItems` function used `pipeline` correctly.

**Bug 2:** Missing `::jsonb` cast on the spec parameter (`$3` → `$3::jsonb`) and raw `[]byte` instead of `string(specJSON)`.

Both fixed. The component-creator agent has now processed its first real items. Templates for tool-list, guide-list, game-list, provocation-card, lobby-grid, brief-explanation, platform-comparison all created and stored in `content_components` with correct function names matching section_type.

### 3. Claimed-item-timeout improved

The pre_query now auto-completes items at 15 minutes when there's evidence the work was done (page deployed, CSS generated after claim time). Items with no evidence still get reset at 40 minutes. Prevents wasteful rebuilds from the response-routing loss pattern.

### 4. Dispatch loop reaper threshold

Added a 30-minute idle threshold for `build-dispatch-loop` orchestrations specifically, while keeping 90 minutes for all other agent types. Prevents zombie dispatch loop pods from accumulating.

### 5. Health check scheduled task

`ai-endpoint-health-check` was targeting `build-dispatch-loop` instead of `endpoint-health-checker`. Fixed.

### 6. Stale model references

`claude-sonnet-4-5` → `claude-sonnet-4-6` in top-level `ai_service` configs.

---

## The Core Problem: Adoption Produces Generic Sites

gamedesign.uk is a hand-crafted dark-themed game developer tool site with:
- Custom `global.css` (dark theme, cyan accents, monospace elements, developer aesthetic)
- Working JavaScript tools (Drop Rate Simulator with Chart.js, Progression Architect)
- Playable games (Jelly Invaders canvas game, A* Pathfinding Playground, P2P Network Simulator)
- Distinctive page layouts (pillar cards, resource grids, stat boxes)
- Guide articles with technical depth (code blocks, callout boxes, working demos embedded)

The adoption pipeline correctly crawled and classified the site — the identity, design palette, content direction, and interactive_features are all captured. But the build pipeline then:

1. Ignored the crawled CSS and applied a generic style collection
2. Fed page content to the LLM content writer, which can't reproduce JavaScript applications
3. Used generic brochure components (hero-centered, features, call-to-action) for pages that had custom layouts
4. Lost all interactive tools — the rawHtml is in research_results but was never deployed

The result: a site that looks like every other brochure site the system builds, when the original was a distinctive interactive platform.

### Why the improvement loop makes it worse

After adoption, the improvement loop runs design audits and quality checks. These agents compare the site against the site specs. The site specs say `visual_tone: "Minimal, utilitarian, dark-themed developer tool aesthetic"` but the design spec lacks specificity — no hex values, no font names, no layout rules. The auditors see a generic dark theme and make it more generic. The content auditors see short listing pages and try to add "missing content" that doesn't match the original voice. Each cycle drifts further from the original.

---

## What Needs Building: Rich Adoption Analysis

The fundamental issue is that the adoption's LLM analysis (the `analyze_site` step) produces a high-level classification but doesn't capture what makes the site distinctive. The content direction spec captures writing style well (thanks to the select_content + derive_content_direction pipeline), but there's no equivalent depth for:

### Design fingerprint
The crawl has rawHtml with inline styles, `<link>` tags to CSS files, and the CSS file content itself. The adoption should extract:
- Actual hex colours (not "dark background")
- Actual font families and sizes
- Layout patterns (grid configurations, spacing values)
- Component-level styling (card borders, hover effects, shadow depths)
- Whether the site uses CSS variables (and what they are)

This becomes a `design_fingerprint` spec that the webdesign-agent reads instead of guessing from industry norms.

### Site archetype
The LLM classification says `page_type: "tool"` or `"content"` but doesn't capture what KIND of tool site this is. gamedesign.uk is a "developer utility platform" — it's not a brochure, not a SaaS landing page, not a content hub. The archetype determines:
- What components are appropriate (tool cards, not testimonial carousels)
- What the index page layout should be (resource grid, not hero → features → CTA)
- How navigation works (pillar-based, not linear)
- What the improvement loop should and shouldn't change

### Interactive element inventory
The adoption captures `interactive_features` with names and descriptions. But the build pipeline needs:
- The actual rawHtml for each tool/game
- External script dependencies (Chart.js, etc.)
- Whether the tool is self-contained (single HTML file) or multi-file (HTML + separate JS + CSS)
- Which page each tool belongs to and its position in the page

### Content inventory with page-level detail
Each adopted page needs a spec that says: "This page has 6 sections in this order, the first is a custom hero with this exact structure, the third is a resource grid with these items linking to these URLs." Currently the LLM outputs section names but the content writer reinvents the layout.

---

## The Rebuild Strategy for Adopted Sites

For pages where rawHtml exists and contains working interactive elements:

### Tier 1: Direct deployment
If the page's rawHtml is self-contained (no broken external references), deploy it as-is. Skip the content writer entirely. The page gets the site's header/footer injected but the body is the original. This is the only way to preserve JavaScript tools and games.

### Tier 2: Template extraction
If the rawHtml has a recognisable component structure (sections with data-component attributes), extract each section as a component template. Store in `content_components` with the correct function name and section_type. Future builds use the template with fresh content.

### Tier 3: LLM recreation with full context
If the rawHtml is too tightly coupled to deploy directly (broken references, CMS-generated markup), give the content writer the FULL original content (not a summary) and instruct it to preserve the structure, tone, and specific claims/examples. The content direction spec guides voice; the page's existing content guides substance.

### Tier 4: Described recreation
If rawHtml doesn't exist for a page (crawl missed it, or it's dynamically generated), the content writer works from the page's description in the adoption analysis plus the site's content direction.

The tier is determined per-page by `apply_adoption_plan` based on what data exists in the crawl.

---

## Implementation Plan

### Phase 1: Rich design capture

The `analyze_site` LLM prompt needs to extract CSS variables, colour values, and font families from the crawl. The `format_crawl_for_analysis` action already provides page summaries — extend it to include CSS extracts (the crawl has rawHtml which includes `<style>` blocks and `<link>` tags). Store as a `design_fingerprint` spec.

The webdesign-agent's CSS generation should read `design_fingerprint` when it exists and use those values instead of guessing. For adopted sites, the CSS should reproduce the original palette, not apply a style collection.

### Phase 2: Direct rawHtml deployment for interactive pages

When `apply_adoption_plan` creates work items for pages with `interactive_features`, set `mode: "direct_deploy"` instead of `mode: "recreate"`. The page-build-handler detects this mode and:
1. Loads rawHtml from `research_results`
2. Injects site header/footer
3. Deploys directly — no content writer, no LLM

This requires a new step in page-build-handler (or a conditional branch) that bypasses the content writer when the spec says direct_deploy.

### Phase 3: Site archetype spec

The adoption analysis should produce a `site_archetype` spec that classifies the site beyond page_type. This spec is read by the improvement loop agents to determine what's appropriate:
- "This is a developer utility platform — do NOT suggest testimonials or case studies"
- "This is a content hub with editorial depth — do NOT simplify to listicles"
- "The index page is a resource directory — do NOT replace with hero → features → CTA"

The improvement loop agents read this spec and constrain their suggestions accordingly.

### Phase 4: Adoption-aware improvement loop

The improvement loop currently treats all sites the same. For adopted sites:
- Design auditors should compare against the design_fingerprint, not generic standards
- Content auditors should compare against the content_direction writing rules, not generic quality
- The loop should NOT create work items that would remove or replace adopted interactive features
- Pages with `mode: "direct_deploy"` should be excluded from content rewrites

---

## Current Database State: gamedesign.uk

```
Site ID: 15a6cb16-5a86-4541-a8e4-d7106239b6a4
Domain:  gamedesign.uk
Status:  active

Specs:   identity (adoption), design (adoption), content_direction (adoption, 17K, formatted=true), structure (adoption)
Pages:   5 deployed (index, tools, guides, games, contact)
Research: 3 adoption_crawl results, adoption_page results per page (rawHtml stored)

Component library now has: tool-list, guide-list, game-list (generated by component-creator)
Also generated for vonc.com: provocation-card, lobby-grid, brief-explanation, platform-comparison
```

### Data available in research_results

The crawl data IS there. The rawHtml for each page is stored in `research_results` with `result_type = 'adoption_page'`. The `adoption_crawl` result has the full LLM analysis including interactive_features. The content_direction spec has the writing style guide. What's missing is the mechanism to deploy the rawHtml directly rather than rewriting it through the content writer.

### Interactive features identified

```json
[
  {"name": "Game Design Calculators", "page": "tools", "type": "calculator", "self_contained": true},
  {"name": "Interactive Game Prototypes", "page": "games", "type": "tool", "self_contained": true}
]
```

---

## Dispatch Loop / Response Routing Analysis

### The zombie pod problem (diagnosed, partially mitigated)

**Root cause chain:**
1. Loop expansion adds steps to `WorkflowPlan.Steps` in memory
2. Concurrent state updates (from fast-responding child agents) can overwrite the expanded steps via optimistic lock conflicts
3. When a later iteration's spawn fails, the timeout handler loads state from DB — expanded steps are gone
4. `routeToErrorStepOrFail` can't find the step, falls through to `failWorkflow`
5. `failWorkflow` tries `UpdateState` but hits another optimistic lock (stale version)
6. The save fails, error is logged but not retried for AWAITING_RESPONSES status
7. Orchestration stays stuck, pod stays alive

**Mitigations applied:**
- Dispatch loop reaper threshold reduced to 30 minutes (was 90)
- Claimed-item-timeout auto-completes items with evidence of completion at 15 minutes

**Go fixes identified but not applied (risk vs reward):**
- `failWorkflow`: retry UpdateState with fresh version when fresh status is AWAITING_RESPONSES
- `handleRequestTimeout`: check return value of `handleRecoverableError` (currently discarded)
- `routeToErrorStepOrFail`: parse loop-expanded step names via `parseLoopStepName` and resolve error_step from parent loop's sub_workflow

These are in `/mnt/user-data/outputs/fix_zombie_dispatch_loop_pods.go` but were not deployed due to risk in core coordinator code. Should be done in a dedicated session with thorough testing.

**Deeper investigation needed:**
- Why `handleLoopExpansion`'s state save loses to concurrent response handlers
- Whether `persistAwaitingStateWithRetry` (which has merge logic) should also be used for loop expansion saves

---

## Files in Project Knowledge to Update

| File | What needs updating |
|------|-------------------|
| `028f_adoption_pipeline_handoff.md` | Replace with this document |
| `001g_development_guide_new_agents_v7.md` | Add: component discovery pipeline pattern, `CreateNeedsNewComponentItem` column naming |
| `003e_contracts_and_standards_v5.md` | Add: adoption-specific component naming (function must match section_type) |

---

## Key Patterns to Remember

- `select_representative_content` is now registered and working
- `CreateNeedsNewComponentItem` uses `pipeline` not `domain` column, and `$3::jsonb` with `string(specJSON)`
- Component-creator generates templates with function names matching the requested section_type
- The `claimed-item-timeout` pre_query has two phases: 15-min evidence-based auto-complete, 40-min blind reset
- The `stale-orchestration-reaper` has a 30-min threshold for dispatch loops, 90-min for everything else
- `ai-endpoint-health-check` targets `endpoint-health-checker` (was incorrectly targeting `build-dispatch-loop`)
- Expanded loop steps can be lost from `workflow_plan` during concurrent state updates — this is a known issue
- `failWorkflow` doesn't retry on optimistic lock when state is AWAITING_RESPONSES — this causes zombie pods

---

## Kubernetes Reference

```bash
kubectl -n ai-persona-system get pods
kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50
kubectl -n kafka get pods
```

Deployment: git push → GitHub Actions → S3 → Cloudflare.

---

## Next Session Priority

1. **Rich adoption analysis** — design fingerprint, site archetype, interactive element inventory
2. **Direct rawHtml deployment** for pages with working interactive features
3. **Adoption-aware improvement loop** — don't regress adopted sites to brochure
4. **Re-adopt gamedesign.uk** with the enriched pipeline and verify the result
