# 028i — Adoption Pipeline & Component Infrastructure Handoff (Consolidated)

Replaces: 028h_adoption_pipeline_handoff_v3, 028i_v4, 034_handoff_vonc_component_infrastructure

Session dates: 2026-04-02 through 2026-04-07

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

After `apply_plan`, the dispatch loop routes:
- Static pages → `page-build-handler` (item_type: `needs_content_page`)
- Interactive pages → `tool-recreation-handler` (item_type: `needs_tool_recreation`)
- Design → `webdesign-agent` (item_type: `needs_design`)
- Rerender → `rerender-pages` (item_type: `needs_rerender`)

---

## Fixes Applied

| Fix | Detail |
|-----|--------|
| Registry: `select_representative_content` | Action existed but wasn't registered |
| `CreateNeedsNewComponentItem` | Column `domain` → `pipeline`, added `::jsonb` cast |
| Claimed-item-timeout | Two-phase: 15-min evidence-based, 40-min blind reset |
| Dispatch loop reaper | 30-min threshold for build-dispatch-loop, 90-min others |
| Health check target | Was targeting build-dispatch-loop instead of endpoint-health-checker |
| Model references | `claude-sonnet-4-5` → `claude-sonnet-4-6` |
| `error_step` placement | Must be inside `step.Config`, not at step level |
| `chk_created_from_valid` | Added `'tool'` and `'forked'`; later tightened to remove `'tool-generator'` |
| `store_generated_component` | Added `stripCodeBlocks()` — fixes corrupted html_template from markdown-wrapped LLM output |
| Component → page rebuild | `markPagesForRebuild()` in store action + `check_unresolved_sections` discovery check |

---

## Mission-Driven Sites (Tier 3)

The domain-submitter accepts structured mission data. Three tiers:

- **Tier 1** — Domain only. Classifier infers everything.
- **Tier 2** — Domain + objective. Classifier uses objective as hint.
- **Tier 3** — Domain + mission + roadmap + briefs. Pre-defined product vision drives everything.

Briefs stored as `{"text": "..."}` in site_specs. Classifier reads `mission_brief.text`, planner reads `roadmap_brief.text`. Roadmap overrides the component list when present.

---

## Site Archetype Classification

LLM produces a multi-dimensional classification: label, character (feel, polish, density), design (palette, layout, typography), content (types, voice), purpose, interaction_patterns, revenue_model, visual_character, audience, structure, and constraints (things the improvement loop must never do).

The archetype is a snapshot of what the site IS, not what it should become.

**Go patch needed:** Archetype extraction block in `apply_adoption_plan_action.go`. Without it, the LLM produces the archetype but `apply_plan` doesn't persist it as a spec.

---

## Tool Recreation Handler

For interactive pages (tools, games, calculators) that can't be reproduced by the content writer. Two-stage LLM pipeline:

1. **analyze_tool** (Sonnet) — reads rawHtml source + context, produces JSON functional spec
2. **recreate_tool** (Opus, 64K tokens) — produces self-contained HTML/CSS/JS from the spec

Completeness check: `check_tool_completeness` Go action verifies completion marker, balanced tags, minimum length.

---

## Component Lifecycle (creation → page rebuild)

When `plan_sections` can't find a component for a section_type:
1. Path 3 fires: `CreateNeedsNewComponentItem()` creates a work item
2. `component-creator` generates template via LLM
3. `store_generated_component` saves it and calls `markPagesForRebuild()` — marks all deployed pages referencing that section_type as `needs_rebuild`
4. `check_unresolved_sections` discovery check (completeness-discovery-agent) catches edge cases on next sweep
5. Normal build pipeline picks up the page, `plan_sections` now finds the component

This loop also fires for the "already_exists" path — if multiple sites request the same section_type, the first creates it and the others trigger rebuilds.

---

## The Core Problem: Adoption Produces Generic Sites

The adoption pipeline correctly captures identity, design, and content direction. But the build pipeline:
1. Ignores crawled CSS and applies generic styles
2. Can't reproduce JavaScript applications via content writer
3. Uses generic brochure components for custom layouts
4. Improvement loop audits against generic standards

**Mitigations deployed:** Interactive pages route to tool-recreation-handler. Component selector + creator handles custom section types. Design fingerprint extraction (priority 3, not yet built) would address CSS reproduction.

---

## What Needs Building Next

| Priority | Item | Status |
|----------|------|--------|
| 1 | Deploy Go archetype patch | Ready, not deployed |
| 2 | Re-crawl gamedesign.uk with rawHtml | Original site restored, ready |
| 3 | Design fingerprint extraction (CSS from crawled HTML) | Not started |
| 4 | Test tool recreation against real crawled tools | Handler deployed, untested |
| 5 | Adoption-aware improvement loop (respect constraints) | Not started |

---

## Database State: Key Sites

**vonc.com** — 12 site_specs aspects, classified as `interactive-platform`. 11 Spark components generated. Pages processing.

**gamedesign.uk** — Original site restored and live. Old adoption data needs cleanup before re-crawl. No rawHtml in current crawl data.

---

## Key Schema Notes

- `site_work_items`: `pipeline` not `domain`
- `sites`: `build_status` not `deploy_status`
- `scheduled_tasks`: `name` not `task_name`
- `site_specs`: `data` not `spec_data`
- `agent_definitions` column mapping: `image_repository` (not container_image), `resources` (not resource_config), `topics` (not topic_config), `delegation_preferences` (not delegation_config), `agent_category` (not role), `domain_tags` (not tags), `idle_timeout_seconds` (not timeout_seconds)

---

## Known Issue: Zombie Dispatch Loop Pods

Loop-expanded steps lost from `workflow_plan` during concurrent state updates. Timeout handler can't find the step, `failWorkflow` hits optimistic lock. **Mitigation:** 30-minute reaper threshold. **Root cause fixes identified but not applied** — too risky without dedicated testing.
