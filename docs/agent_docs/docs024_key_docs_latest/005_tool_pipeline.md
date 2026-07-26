# 005 — Tool Pipeline

Tool suggestion, generation, deployment, cross-linking, and content integration.

---

## Pipeline Status: Fully Operational

The complete tool pipeline works end-to-end autonomously:

```
check_missing_tools → evaluate_tools → tool-suggester (LLM + routing; related_pages travel
                                                       on the add_tool item spec)
  → tool-generator / tool-deployer (component + page + nav + guide
                                    + cross-link content_rewrite items, using the page's
                                      REAL url and gated on it going live)
  → build-dispatch-loop → page-build-handler → page-content-writer (with rewrite_guidance)
  → page-rerender → deployed with tool references in HTML
```

> **CORRECTED 2026-07-26 (bugs_open/029).** This flow used to show a `create_cross_links`
> step on tool-suggester, emitting the cross-link items at SUGGESTION time. That step is
> deleted (migration 211) because it constructed the tool page's URL as
> `/tools/{function}.html` — a URL the platform never produces consistently (three shapes
> exist), and one that cannot be looked up before the page row is created. **0 of 27 items
> emitted that way resolved to a real page.** The emit now happens inside the two build
> actions, which hold the `pages.url` they just wrote.

All deployed sites have tools. Tool links land in content pages via cross-linking. Novel tool generation, library forking, companion guides, nav entries, and content rewriting all work without manual intervention.

---

## Deployed Components

### SQL Migrations Applied

| Migration | What it does |
|-----------|-------------|
| 070 | tool-suggester routing: `check_is_library` conditional routes library tools to tool-deployer, novel tools to tool-generator |
| 071 | tool-suggester cross-linking: `related_pages` in LLM schema, instruction 6, `create_cross_links` step after `create_items_loop` — **the step is REMOVED by 211**; `related_pages` in the schema stays and is what the build path reads |
| 211 | bugs_open/029: delete `create_cross_links`; wire `related_pages` into `deploy_tool`/`save_tool` so the BUILD emits the cross-links with the real URL |
| 072 | Rewrite guidance threading: `rewrite_guidance?` in page-build-handler input_mapping, `rewrite_guidance` in page-content-writer's nested prompt |
| 073 | Tightened `created_from` constraint: removed `'tool-generator'`, migrated rows to `'generated'` |
| — | content-quality-auditor: fixed `check_empty_pages` GROUP BY (added `p.id`) |
| — | Password checkers deactivated: all `tool-password-entropy` components `is_active = false` |

### Go Actions Deployed

| Action | File | Purpose |
|--------|------|---------|
| `deploy_tool_to_site` | `deploy_tool_action.go` | Fork library tool, create page + content sections + companion guide |
| `create_tool_component` | `create_tool_component_action.go` | Novel tool from LLM HTML: component, page, page_component (rendered_html at position 2), nav entry, content work items, companion guide |
| `create_tool_cross_link_items` | `create_tool_cross_link_items.go` | **No longer the emitter (bugs_open/029).** Holds `emitToolCrossLinkItems`, called by the two build actions with the tool page's real `pages.url`. The action itself is kept registered but fail-safe: it resolves the tool to a real page and emits nothing when there is none, so a stale config naming it cannot fabricate. |
| `update_component_html` | `update_component_html_action.go` | Update tool HTML with optional version snapshot |

### Discovery Checks

| Check | File | Cooldowns |
|-------|------|-----------|
| `missing_tools` | `check_missing_tools.go` | 0 tools: 7 days, 1+ tools: 30 days |
| `tool_health` | `check_tool_health.go` | Tier 1 structural + Tier 2 audit queue (30-day per tool) |

---

## Key Implementation Details

### Cross-Linking Flow

tool-suggester's LLM returns `related_pages` per suggestion (1-3 page names), which travel on the `add_tool` work item spec. The **build** action for that tool (`deploy_tool_to_site` or `create_tool_component`) then calls `emitToolCrossLinkItems`, which:

1. Loads page map (name → ID) for the site
2. Iterates suggestions, extracts `related_pages` arrays
3. Skips pages starting with `tool-` (tool-to-tool linking isn't useful)
4. Creates `content_rewrite` work items with `page_name` in spec and `suggestion` containing natural-language guidance for the content writer
5. Dedup via `item_key = tool_crosslink:{function}:{page}:{site}`
6. Priority 110 (low urgency — cross-links are additive, not blocking)
7. **Takes the tool page's real `pages.url`** — never constructs one — and refuses to emit if
   handed anything that is not an absolute path
8. **Gates on the page going live**: emits immediately only when the tool page is
   `deployed`/`needs_rebuild`; otherwise `depends_on` = the open `needs_content_page` item for
   that page. No open item (or a terminally failed one) → nothing is emitted. A tool page that
   never deploys therefore leaves parked items rather than a live 404.

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
