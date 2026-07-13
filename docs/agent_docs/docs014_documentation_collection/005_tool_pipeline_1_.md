# 005 — Tool Pipeline

Tool suggestion, generation, deployment, cross-linking, and content integration.

---

## Pipeline Status: Fully Operational

The complete tool pipeline works end-to-end autonomously:

```
check_missing_tools → evaluate_tools → tool-suggester (LLM + routing + cross-linking)
  → tool-generator / tool-deployer (component + page + nav + guide)
  → create_cross_links (content_rewrite items for related pages)
  → build-dispatch-loop → page-build-handler → page-content-writer (with rewrite_guidance)
  → page-rerender → deployed with tool references in HTML
```

All deployed sites have tools. Tool links land in content pages via cross-linking. Novel tool generation, library forking, companion guides, nav entries, and content rewriting all work without manual intervention.

> **Known hazard (CONFIRMED 2026-06-22, fix pending) — a content rebuild de-tools a tool page.** A deployed tool is NOT protected from a later *content* rebuild. A `link_resolution_rebuild` (intended as links-only) — or any `needs_page`/`needs_content_page` — for a tool/game page is handled by `page-build-handler` → page-content-writer, which regenerates the page from `plan_sections`. The plan has no knowledge of the interactive tool (it lives as a section's `rendered_html`), so the tool is silently replaced with generated content and the page falls back to `generic-text-block`. Confirmed on gamesdesign `game-pathfinding` (its 18 KB A* canvas overwritten 2026-06-14 20:07; root-caused 06-22; blast radius one page). The content-regression guard does not catch it (markup/JS loss, not prose). Fix direction: route link maintenance through a preserve-sections re-render path (reuse the `page_rerender` machinery); stamp `source_item_id` into `page_component_history`; add an interactivity-aware save guard; then re-run the tool-recreation for the affected page. See 016b debugging guide, 020 Tool Lifecycle, 026 Component Regeneration Flow, 002 (Work-item routing → interactive-page hazard).

---

## Deployed Components

### SQL Migrations Applied

| Migration | What it does |
|-----------|-------------|
| 070 | tool-suggester routing: `check_is_library` conditional routes library tools to tool-deployer, novel tools to tool-generator |
| 071 | tool-suggester cross-linking: `related_pages` in LLM schema, instruction 6, `create_cross_links` step after `create_items_loop` |
| 072 | Rewrite guidance threading: `rewrite_guidance?` in page-build-handler input_mapping, `rewrite_guidance` in page-content-writer's nested prompt |
| 073 | Tightened `created_from` constraint: removed `'tool-generator'`, migrated rows to `'generated'` |
| — | content-quality-auditor: fixed `check_empty_pages` GROUP BY (added `p.id`) |
| — | Password checkers deactivated: all `tool-password-entropy` components `is_active = false` |

### Go Actions Deployed

| Action | File | Purpose |
|--------|------|---------|
| `deploy_tool_to_site` | `deploy_tool_action.go` | Fork library tool, create page + content sections + companion guide |
| `create_tool_component` | `create_tool_component_action.go` | Novel tool from LLM HTML: component, page, page_component (rendered_html at position 2), nav entry, content work items, companion guide |
| `create_tool_cross_link_items` | `create_tool_cross_link_items.go` | Create `content_rewrite` items for related pages. Includes `page_name` in spec (for dispatch loop mapping). Filters out `tool-` prefixed pages. |
| `update_component_html` | `update_component_html_action.go` | Update tool HTML with optional version snapshot |

### Discovery Checks

| Check | File | Cooldowns |
|-------|------|-----------|
| `missing_tools` | `check_missing_tools.go` | 0 tools: 7 days, 1+ tools: 30 days |
| `tool_health` | `check_tool_health.go` | Tier 1 structural + Tier 2 audit queue (30-day per tool) |

---

## Key Implementation Details

### Cross-Linking Flow

tool-suggester's LLM now returns `related_pages` per suggestion (1-3 page names). After `create_items_loop`, the `create_cross_links` step calls `create_tool_cross_link_items`, which:

1. Loads page map (name → ID) for the site
2. Iterates suggestions, extracts `related_pages` arrays
3. Skips pages starting with `tool-` (tool-to-tool linking isn't useful)
4. Creates `content_rewrite` work items with `page_name` in spec and `suggestion` containing natural-language guidance for the content writer
5. Dedup via `item_key = tool_crosslink:{function}:{page}:{site}`
6. Priority 110 (low urgency — cross-links are additive, not blocking)

### Rewrite Guidance Threading

For cross-link items to actually produce tool references in page content:

1. **page-build-handler** maps `"rewrite_guidance?": "input_data.spec.suggestion"` to the content writer
2. **page-content-writer** receives `rewrite_guidance` in the per-section LLM `input_fields`
3. The prompt includes `{{if .rewrite_guidance}}## Rewrite Guidance` block before `## Section Requirements`

The content writer's prompt is nested deep:
```
default_config → workflow → steps → process_sections_loop → config → sub_workflow → steps → generate_content → config → prompt_template
```
Top-level `jsonb_each(steps)` iteration will NOT find it.

### Dispatch Loop Input Mapping

The `build-dispatch-loop` maps work item spec fields to handler input:
```json
"page_name?": "current_item.spec.page_name"
```
Any work item spec that needs page routing **must include `page_name`** — not just `page` or `page_id`.

---

## Bug History

| Bug | Cause | Fix |
|-----|-------|-----|
| Password checker on every site | `matchToolToSite` classified security tags as universal | Removed tag matching, use LLM evaluation |
| tool-suggester SQL errors | Queried `p.slug`, `p.purpose`, `p.sort_order` (don't exist) | Fixed to `p.url`, `p.meta_description`, `p.nav_order` |
| Novel tools failed at tool-deployer | All suggestions routed to tool-deployer, null `tool_component_id` | Added `check_is_library` routing in loop sub_workflow |
| `created_from` constraint | Code used `'tool-generator'`, constraint only allowed `'generated'` | Temporarily widened, then tightened after migration |
| `addToolToNav` silent failure | Wrong column names for `site_nav_groups` and `site_nav_items` | Fixed to match actual schema |
| Cross-link items: no page_name | Spec had `"page"` but dispatch loop maps `spec.page_name` | Added `"page_name": pageName` to spec |
| 072 prompt injection missed | DO block searched top-level steps, prompt is in loop sub_workflow | Targeted correct nested jsonb path |
| content-quality-auditor SQL | `HAVING NOT EXISTS` referenced `p.id` not in GROUP BY | Added `p.id` to GROUP BY |
| Cross-link priority too low | Priority 110 queued behind everything | Acceptable long-term; bumped to 30 for testing |
| Tool page de-tooled by a content rebuild | `link_resolution_rebuild`/`needs_page` routes to page-build-handler → full writer regenerates from `plan_sections`, which doesn't know the tool → tool dropped (`game-pathfinding`, 06-22) | CAUSE CONFIRMED 2026-06-22, fix pending — route link maintenance through a preserve-sections re-render path; see 016b / 020 / 026 |

---

## Verified Results

### gaswholesalers.com
- Gas Unit Converter (novel, deployed, live)
- Fuel Cost Estimator (old code, manually patched)
- 18 cross-link items created, tool links confirmed in deployed HTML for fuel-supply-by-industry, fleet-fuel-services, natural-gas-distribution, wholesale-fuel-distribution

### ai-agent-orchestration.com
- AI Agent ROI Estimator (novel, deployed)
- 9 cross-link items, no tool pages targeted (filter working)

### finetuning.uk
- AI Readiness Quiz (novel) + AI Time Savings Estimator (library fork)

### leopardessconsulting.co.uk
- 2 tools deployed

---

## Remaining Cleanup

1. **Fuel Cost Estimator on gaswholesalers** — old code, manually patched. Missing companion guide and proper nav entry. Tool health audit will catch it, or delete and let tool-suggester recreate on next 30-day cycle.

2. **Pre-072 cross-link pages** — why-gas-wholesalers and rack-pricing-programs completed before the prompt fix, so no tool links. Will be addressed on next evaluation cycle.

---

## Future: News → Tool Linking

The news pipeline produces `latest-news.json` but doesn't write blog posts yet. When article writing is added, pass the site's deployed tool list to the writer's prompt — the `rewrite_guidance` pattern is already established.

---

## Architecture Reminders

- Every agent is an orchestrator
- Workflows in SQL, complexity in Go
- Don't create sub-workflows — spawn sub-agents
- Agents respond to parent's response topic
- `logger.Info` not `logger.Debug`
- Check schemas before SQL
- `agent_definitions.type` not `name`
- Kubernetes: `-n ai-persona-system` for pods, `-n kafka` for Kafka
- Check sub_workflow nesting when searching agent definition prompts
- Work item specs needing page routing must include `page_name`
