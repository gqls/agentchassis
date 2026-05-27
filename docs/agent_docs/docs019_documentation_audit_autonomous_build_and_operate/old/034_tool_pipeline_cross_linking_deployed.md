# Handoff: Tool Pipeline — Cross-Linking Deployed + End-to-End Verified

## Context

We're building a system that intelligently plans and builds multipage websites given domain names. Sites get interactive tools (calculators, converters, quizzes etc.) suggested and deployed automatically by an agent pipeline. This session took the tool pipeline from "cross-linking planned but not deployed" to "cross-linking fully working end-to-end with tool references landing in deployed page HTML."

Prior handoff: `033_tool_pipeline_end_to_end_handoff.md`

Key docs:
- `012b_tool_lifecycle_guide_v3.md` — full pipeline, all agents, work items, quality tiers
- `010b_tool_library_guide_v3.md` — library inventory, deployment, templates, both fork and generation paths
- `001g_development_guide_new_agents_v8.md` — agent creation guidelines, updated bug history

---

## What's Deployed and Working — Full Pipeline

### Complete end-to-end flow (all verified)

```
check_missing_tools → evaluate_tools → tool-suggester (LLM evaluation + routing + cross-linking)
  → tool-generator / tool-deployer (create component + page + nav + guide)
  → create_cross_links (content_rewrite items for related pages)
  → build-dispatch-loop → page-build-handler → page-content-writer (with rewrite_guidance)
  → page-rerender → deployed with tool references in HTML
```

### SQL changes applied this session

1. **071_tool_cross_linking.sql** — tool-suggester workflow updates:
   - Added `related_pages` field to LLM output JSON schema
   - Added instruction 6: LLM picks 1-3 existing page names per tool whose topic connects naturally
   - Added `create_cross_links` step after `create_items_loop` → calls `create_tool_cross_link_items` action → then `complete`
   - `create_items_loop.next_step` changed from `"complete"` to `"create_cross_links"`
   - Added `cross_links_created` to complete step output_fields
   - Used `jsonb_set` + DO block with escaping detection (not `regexp_replace` — the original 071 approach was risky with escaped JSON)

2. **072_cross_link_rewrite_support.sql** — threads rewrite guidance through to the content writer:
   - **page-build-handler**: added `"rewrite_guidance?": "input_data.spec.suggestion"` to `call_content_writer` input_mapping
   - **page-content-writer**: added `rewrite_guidance` to `generate_content` input_fields (inside `process_sections_loop` sub_workflow)
   - **page-content-writer prompt**: added `{{if .rewrite_guidance}}## Rewrite Guidance` block before `## Section Requirements`
   - Note: the prompt is nested inside `process_sections_loop.config.sub_workflow.steps.generate_content.config.prompt_template` — top-level step iteration won't find it

3. **073_tighten_created_from.sql** — constraint cleanup:
   - Migrated 1 row from `created_from = 'tool-generator'` to `'generated'`
   - Tightened constraint: `CHECK (created_from IS NULL OR created_from IN ('manual', 'generated', 'adopted', 'tool', 'forked'))`
   - `'tool-generator'` is no longer a valid value

4. **content-quality-auditor fix** — `check_empty_pages` query had ungrouped column error:
   - Added `p.id` to `GROUP BY` clause: `GROUP BY p.id, p.name`
   - The `HAVING NOT EXISTS` subquery referenced `p.id` which wasn't in the GROUP BY

### Go changes deployed this session

1. **`create_tool_cross_link_items.go`** — two fixes:
   - Added `"page_name": pageName` to the spec map (alongside existing `"page"` field). The `build-dispatch-loop` maps `spec.page_name` → `input_data.page_name`; without this, page-build-handler couldn't find the page.
   - Added tool page filter: `strings.HasPrefix(pageName, "tool-")` → skip. Cross-linking one tool page to another tool page isn't useful.

### Verified results

**gaswholesalers.com** — 18 cross-link items created, tool links confirmed in deployed HTML:
- `fuel-supply-by-industry` ✓ (tool link in rendered HTML)
- `fleet-fuel-services` ✓
- `natural-gas-distribution` ✓
- `wholesale-fuel-distribution` ✓
- Pages completed before 072 fix: no tool links (expected — rewrite_guidance wasn't reaching the LLM)
- `tool-fuel-cost-estimator`: no link (tool page referencing another tool — now filtered by Go fix)

**ai-agent-orchestration.com** — 9 cross-link items created after tool page filter deployed, none targeting tool pages ✓

### Tool counts across deployed sites

| Domain | Tool Count |
|--------|-----------|
| ai-agent-orchestration.com | 1 |
| finetuning.uk | 2 |
| leopardessconsulting.co.uk | 2 |

---

## Bugs Found and Fixed This Session

1. **`page_name` missing from cross-link spec** — `build-dispatch-loop` maps `current_item.spec.page_name` but the spec only had `"page"`. Page-build-handler got empty page_name and couldn't load the page record. Fix: added `"page_name": pageName` to spec.

2. **072 prompt injection missed its target** — the DO block searched top-level workflow steps via `jsonb_each(steps)`, but the content writer's LLM prompt is nested inside `process_sections_loop → sub_workflow → generate_content`. Fix: targeted the correct jsonb path directly.

3. **content-quality-auditor SQL error** — `check_empty_pages` query: `HAVING NOT EXISTS (... WHERE lpc.page_id = p.id ...)` referenced `p.id` which wasn't in `GROUP BY p.name`. Postgres rejects this. Fix: `GROUP BY p.id, p.name`.

4. **Cross-link priority too low** — items created at priority 110, sitting behind everything else. Bumped test items to 30 for verification. Default of 110 is acceptable long-term (cross-links are low urgency).

---

## Database Schema Reminders

```
agent_definitions: type (not name), default_config (jsonb)
site_nav_groups: id, site_id, group_key, group_label, group_type, position
  UNIQUE: (site_id, group_key)
site_nav_items: id, site_id, group_id, label, url, page_id, item_type, position, status
  NO unique constraint (use NOT EXISTS for dedup)
content_components.created_from: CHECK allows manual, generated, adopted, tool, forked
  ('tool-generator' has been removed — all rows migrated to 'generated')
```

**Nested prompt location** — the page-content-writer's per-section LLM prompt is at:
```
default_config -> workflow -> steps -> process_sections_loop -> config -> sub_workflow -> steps -> generate_content -> config -> prompt_template
```
This is NOT visible from top-level `jsonb_each(default_config->'workflow'->'steps')` iteration.

---

## Remaining Cleanup Items

1. **Fuel Cost Estimator on gaswholesalers** — created with old Go code, manually patched. Missing: companion guide, content work items, proper nav entry. Tool health audit will improve it, or it can be deleted and recreated on the next 30-day evaluation.

2. **Orphaned tool pages** — gaswholesalers has `tool-fuel-cost-estimator` page with only 1 page_component (no hero/CTA sections). Cross-linking now avoids targeting tool pages directly.

3. **gaswholesalers cross-link items pre-072** — 3 items completed before the prompt fix, so those pages (why-gas-wholesalers, rack-pricing-programs) don't have tool links. They'll get picked up on the next evaluation cycle or can be manually reset.

---

## Future: News → Tool Linking

The news pipeline currently produces `latest-news.json` (rendered JSON feed committed to git via GitHub Actions → S3). It doesn't write full blog posts from news items yet.

When news article writing is added, the natural integration point is passing the site's deployed tool list to the article writer's prompt — similar to how `link_context` constrains internal links today. The content writer already receives `rewrite_guidance` so the pattern is established.

---

## Files Produced This Session

| File | Status |
|------|--------|
| `071_tool_cross_linking_v2.sql` | ✅ Applied to DB |
| `072_cross_link_rewrite_support.sql` (rewritten for nested prompt) | ✅ Applied to DB |
| `073_tighten_created_from.sql` | ✅ Applied to DB |
| `content-quality-auditor` check_empty_pages fix | ✅ Applied to DB |
| `create_tool_cross_link_items.go` (page_name + tool page filter) | ✅ Deployed |

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
- When searching for prompts in agent definitions, check sub_workflow nesting — not just top-level steps
- `build-dispatch-loop` input_mapping: `"page_name?": "current_item.spec.page_name"` — any spec that needs page routing must include `page_name`
