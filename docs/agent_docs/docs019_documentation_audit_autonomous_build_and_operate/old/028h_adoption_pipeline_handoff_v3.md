# Adoption Pipeline — Handoff Document v3

Session dates: 2026-04-02, 2026-04-03

---

## What Was Fixed Across These Sessions

### 1. Registry fix: select_representative_content
Action existed in Go but was never registered in `registry.go`. Added. This unblocked the content direction pipeline.

### 2. Component discovery pipeline unblocked (two bugs in CreateNeedsNewComponentItem)
- **Bug 1:** Column `domain` doesn't exist on `site_work_items` — should be `pipeline`
- **Bug 2:** Missing `::jsonb` cast on spec parameter
Both fixed and deployed. Component-creator agent has now processed real items. Templates generated: tool-list, guide-list, game-list (gamedesign.uk), provocation-card, lobby-grid, brief-explanation, platform-comparison (vonc.com).

### 3. Claimed-item-timeout improved
Two-phase: 15-min evidence-based auto-complete (checks page deployed, CSS generated), 40-min blind reset. Prevents wasteful rebuilds from response-routing loss.

### 4. Dispatch loop reaper threshold
30-minute idle threshold for `build-dispatch-loop` specifically, 90 minutes for everything else. Applied via SQL update to `stale-orchestration-reaper` pre_query.

### 5. Health check task target
`ai-endpoint-health-check` was targeting `build-dispatch-loop` instead of `endpoint-health-checker`. Fixed.

### 6. Stale model references
`claude-sonnet-4-5` → `claude-sonnet-4-6` in ai_service configs.

### 7. Site archetype classification step added to adoption workflow
New `classify_archetype` step added between `analyze_site` and `select_content`. SQL applied, verified:
```
analyze_site → classify_archetype → select_content → derive_content_direction → apply_plan
```
**Go change still needed:** Insert archetype extraction block in `apply_adoption_plan_action.go` at line 219. Patch file: `patch_apply_adoption_plan_archetype.go`. Without this, the LLM produces the archetype but `apply_plan` doesn't write it as a spec.

---

## Current Adoption Workflow (11 steps)

```
ensure_site_record → crawl_site → format_crawl → check_crawl_content
  → analyze_site (LLM: structure, identity, design, pages, sections)
  → classify_archetype (LLM: character, purpose, constraints, design patterns)
  → select_content (Go: pick representative pages)
  → derive_content_direction (LLM: writing style guide, ~17K chars)
  → apply_plan (Go: write specs, create pages, create work items)
  → complete
```

---

## The Core Problem: Adoption Produces Generic Sites

gamedesign.uk is a dark-themed developer utility platform with working JavaScript tools (Drop Rate Simulator, Progression Architect), playable games (Jelly Invaders, A* Pathfinding), and technical guides. The adoption pipeline correctly captures identity, design palette, interactive features, and content direction. But the build pipeline then:

1. Ignores crawled CSS and applies a generic style collection
2. Feeds content to the LLM content writer, which can't reproduce JavaScript applications
3. Uses generic brochure components (hero → features → CTA) for pages that had custom layouts
4. The improvement loop then audits against generic standards and makes it worse

**The site has been restored to its original state and is ready for re-crawling.**

---

## Site Archetype Classification

Designed and tested this session. The LLM produces a multi-dimensional classification:

- **label**: human-readable 2-4 word description ("Game Dev Utility Hub")
- **character**: feel, polish, budget impression, age impression, commercial intent, density
- **design**: palette mood, layout approach, typography feel, imagery, animation, responsive
- **content**: primary/secondary types, voice, media
- **purpose**: array of what the site exists to do
- **content_model**: array of content types present
- **interaction_patterns**: array of what users DO
- **revenue_model**: none/advertising/affiliate/e-commerce/etc
- **visual_character**: array of style tags
- **audience**: array of who it's for
- **structure**: index layout, listing style, navigation, page depth
- **constraints**: array of things the improvement loop should NEVER do

Tested against gamedesign.uk — output was accurate and actionable. The prompt is in the `classify_archetype` step config in the adoption agent definition.

The archetype is a snapshot, not an aspiration. It describes what the site IS. The strategist determines what it should become.

---

## What Needs Building Next

### Priority 1: Deploy the Go archetype patch
Insert the archetype extraction block in `apply_adoption_plan_action.go` at line 219. Without this, `classify_archetype` runs but the output isn't persisted as a spec. Patch in `patch_apply_adoption_plan_archetype.go`.

### Priority 2: Re-crawl gamedesign.uk with rawHtml capture
The previous crawl did NOT capture rawHtml — only markdown. The `has_raw_html` column is false for all adoption_page records. Check whether firecrawl_crawl is configured to return rawHtml and fix if not. The original site is now restored and ready.

### Priority 3: Design fingerprint extraction
Extract actual CSS values from crawled rawHtml — hex colours, font families, CSS variables, spacing values. Store as `site_specs` aspect `design_fingerprint`. The webdesign-agent reads this instead of guessing from the vague design spec ("dark background", "bright accent"). This is the difference between reproducing the site's look and producing another generic dark theme.

### Priority 4: Direct rawHtml deployment for interactive pages
Pages with self-contained JavaScript tools/games should bypass the content writer entirely. The page-build-handler detects `mode: "direct_deploy"` in the work item spec and:
1. Loads rawHtml from research_results
2. Injects site header/footer
3. Deploys directly — no LLM, no template rendering

This is the only way to preserve Drop Rate Simulator, Jelly Invaders, A* Pathfinding, etc.

### Priority 5: Component recreation from functional specs
For components that can't be directly deployed (broken references, CMS dependencies), the adoption analysis should produce per-component functional specs. The component-creator builds from spec using our contracts and CSS system. The result works like the original but is maintainable by our agents.

The user has been working elsewhere on the framework's ability to recreate components if they don't already exist in the DB. Next session should start with that updated code.

### Priority 6: Adoption-aware improvement loop
- Design auditors compare against design_fingerprint, not generic standards
- Content auditors respect content_direction writing rules
- Pages with mode: "direct_deploy" excluded from content rewrites
- Site archetype constraints checked before any suggestion is applied

---

## Database State: gamedesign.uk

```
Site ID: 15a6cb16-5a86-4541-a8e4-d7106239b6a4
Domain:  gamedesign.uk
Status:  active
Image tag: v1.0.935

Specs:   identity, design, content_direction (17K, formatted=true), structure
         site_archetype NOT YET WRITTEN (Go patch not deployed)
Pages:   5 deployed (index, tools, guides, games, contact) — currently generic brochure
Research: 3 adoption_crawl results, adoption_page per page (NO rawHtml)

Component library has: tool-list, guide-list, game-list (auto-generated)
The original site has been restored and is live for re-crawling.
```

---

## Diagnosed But Not Applied: Coordinator Go Fixes

### Zombie dispatch loop pods — full root cause

Loop-expanded steps are lost from `workflow_plan` during concurrent state updates. When a later iteration fails, the timeout handler can't find the step, `failWorkflow` hits optimistic lock and can't save FAILED status. Pod stays alive.

**Identified fixes (not applied — too risky without dedicated testing):**
- `failWorkflow`: retry UpdateState with fresh version when fresh status is AWAITING_RESPONSES
- `handleRequestTimeout` line ~3039: check return value of `handleRecoverableError` (currently discarded)

**Applied mitigation:** 30-minute dispatch loop reaper threshold.

**Files reviewed:** coordinator.go, spawn_actions.go, loop_expansion_handler.go, loop_error_handler.go

---

## Key Schema Notes

- `site_work_items`: uses `pipeline` not `domain`, has `item_key` with partial unique index
- `agent_definitions`: uses `image_tag` not `container_image_tag`
- `sites`: no `deploy_status` column, uses `build_status`
- `scheduled_tasks`: uses `name` not `task_name`
- `orchestration_states`: uses `owner_agent_type`, workflow_plan key is lowercase `steps`

---

## Kubernetes Reference

```bash
kubectl -n ai-persona-system get pods
kubectl -n kafka get pods
# Kafka cluster: personae-kafka-cluster-combined-pool-prod-{0,1,2}
```

Deployment: git push → GitHub Actions → Backblaze S3.

---

## Guidelines Amendments

Detailed amendments documented in `028e_amendments.md` covering:
- 028e: workflow steps (7→11), content direction ownership, component discovery operational notes, phase plan updates, new principle (adopted state is the floor), dispatch loop/timeout notes
- 001g: column name `domain` → `pipeline`, new lesson on INSERT column drift and `::jsonb` cast pattern
