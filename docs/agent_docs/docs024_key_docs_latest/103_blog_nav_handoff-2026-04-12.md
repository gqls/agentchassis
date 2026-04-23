# Session Handoff — April 12 2026

## Completed This Session

### 1. rebuild_blog_listing_action.go — rewritten, deployed
**File**: `platform/orchestration/actions/rebuild_blog_listing_action.go`

Three bugs fixed:

**Bug 1: Slot name mismatch (Problem 1 from prior handoff)**
Old code always wrote to `slot_name = 'blog-listing'`. Some blog pages were planned with `article-grid`, `featured-article`, `content-listing`, etc. The component got created but page assembly ignored it because slot names didn't match.

Fix: New `findBlogListingSlot()` function with three strategies:
1. Checks existing `page_components` for the blog-index page against a priority list: `blog-listing`, `article-grid`, `content-listing`, `guide-list`, `featured-article`
2. Falls back to the page's `sections` JSONB array (the planner's plan)
3. Defaults to `blog-listing` only if nothing else found

Returns both the slot name AND the existing component UUID, so the upsert writes to the correct row.

**Bug 2: Wrong template loaded (Problem 2 from prior handoff)**
Old `loadBlogListingTemplate()` loaded the `blog-listing` content_component which was CSS-only — empty `input_schema: {}`, no `{{range}}` block. 10KB of CSS rendered, zero articles appeared.

Fix: Replaced with `loadContentListingTemplate()` which loads by `function = 'content-listing'` with an extra `LIKE '%range%'` guard to avoid CSS-only templates. Falls back to `article_grid` by name, then to a built-in default that has proper `{{range .articles}}` with article cards including `<a href>`.

Old `loadBlogListingTemplate()` removed — no other callers.

**Bug 3: Missing article links (Problem 3 from prior handoff)**
The `content-listing` template renders `<h3 class="article-card__title">{{.title}}</h3>` without `<a href>`. Articles displayed but weren't clickable.

Fix: Two layers:
1. SQL patch applied to `content_components` — replaces `{{.title}}` with `<a href="{{.url}}">{{.title}}</a>` in the `content-listing` template
2. New `ensureArticleLinks()` function as a post-render safety net — scans rendered HTML for titles not wrapped in `<a>` and patches them using the article data

**Additional improvements in the rewrite:**
- Writes `content_data` JSON alongside `rendered_html` (source-of-truth principle per contracts doc 003)
- New `estimateReadTime()` — computed from a subquery summing `page_components.rendered_html` length per blog post, excluding header/footer/head slots
- Blog post query now includes content_length subquery (new `contentLength` scan variable)
- Added `encoding/json` and `regexp` imports; removed unused `id` variable from blog post scan (was scanned but never used)
- `defaultBlogListingTemplate` now has proper article cards with `<a href>`, date, read_time, and excerpt

### 2. content-listing template SQL fix — applied
```sql
UPDATE content_components
SET html_template = REPLACE(
    html_template,
    '<h3 class="article-card__title">{{.title}}</h3>',
    '<h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>'
)
WHERE function = 'content-listing'
  AND html_template NOT LIKE '%href="{{.url}}"%';
```

### 3. nav_drift work items created for 2 sites
Created `triaged` nav_drift items for:
- robot-hands.com (00ff3af5-dad8-4770-9f70-3edc267a3c92)
- gaswholesalers.com (5fe15466-4e2e-4ff2-981e-98c1b7074002)

gaswholesalers.com has 4 pages with `in_header=true` or `in_footer=true` that aren't in `site_nav_items`:
- fuel-industry-insights
- fuel-supply-by-industry
- tool-fuel-cost-estimator
- tool-gas-unit-converter

robot-hands.com had a prior `complete` nav_drift (April 9) — the check query returned no drift pages for it now, so it may already be resolved. The item was created anyway since the prior `complete` wouldn't block it.

finetuning.uk and leopardessconsulting.co.uk were NOT inserted — they didn't match the `NOT EXISTS` guard (finetuning.uk has an `unresolved` item already) or aren't `active`.

### 4. Orphan page items cleaned up
```sql
UPDATE site_work_items
SET status = 'wont_fix', error = 'superseded — reclassified by updated orphan_pages check'
WHERE item_type = 'orphan_page' AND status = 'unresolved';
```

---

## Active Problem: nav-updater Never Spawns

### Symptoms
- All prior `nav_drift` items failed with "Claim timed out (attempts exhausted)"
- No nav-updater pod has ever appeared in `kubectl get pods`
- No "nav-updater" entries in agent-chassis logs
- New `triaged` items remain in `triaged` — not being picked up

### Investigation findings

**nav-updater agent_definition exists and is active** (image_tag v1.0.952):
```
ensure_site_record → refresh_nav_tables → render_site_components → get_pages
  → check_has_pages → spawn_rerenderer → spawn_deployer → rerender_loop → trigger_deploy → complete
```
Workflow is well-structured — uses `populate_nav_tables`, `render_site_components`, sub-spawns `page-rerender` and `deployer-agent`.

**Kafka topics exist** but no `.process` topic (which is expected — agents don't use `.process`):
- `system.agent.nav-updater.dlq`
- `system.agent.nav-updater.errors`
- `system.agent.nav-updater.requests`
- `system.agent.nav-updater.responses`

**build-dispatch-loop workflow is generic** — `spawn_handler` uses `agent_type_field: "current_item.handler_agent"`, so it dynamically resolves to "nav-updater" from the work item. No hardcoded handler list. Works for all other handler agents (page-build-handler, rerender-pages, etc. all spawn fine).

**Spawn mechanism**: `SpawnAgentAction` in `spawn_actions.go` creates K8s Jobs. Steps:
1. `extractSpawnConfiguration` — resolves agent_type from config or field path
2. `getAgentDefinition` — loads from DB
3. `createAgentInDBFromDefinition` — inserts agent record
4. `setupAgentTopics` — creates request/response topics
5. `spawnAgentKubernetesJobFromDefinition` — creates K8s Job
6. Pre-registers awaited request
7. Sends initialization message

### Where to investigate next

The items are stuck in `triaged`, meaning the dispatch loop hasn't even attempted them yet. Check:

1. **Are the dispatch loops even loading these items?**
   ```sql
   -- Check if there are claimed items blocking the sites
   SELECT swi.site_id, s.domain, swi.status, swi.item_type, swi.handler_agent
   FROM site_work_items swi
   JOIN sites s ON swi.site_id = s.id
   WHERE swi.site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92', '5fe15466-4e2e-4ff2-981e-98c1b7074002')
     AND swi.status IN ('claimed', 'triaged')
   ORDER BY swi.site_id, swi.status;
   ```
   The dispatch loop's `find_dispatchable_site` skips sites with claimed items. If another item is claimed for these sites, nav_drift waits.

2. **Check dispatch loop logs for these sites:**
   ```bash
   kubectl -n ai-persona-system logs -l app=agent-chassis --since=30m | grep -i "gaswholesalers\|robot-hands\|nav_drift\|nav-updater" | head -30
   ```

3. **If items DO get claimed but nav-updater never appears**, check spawn_actions logs:
   ```bash
   kubectl -n ai-persona-system logs -l app=agent-chassis --since=30m | grep -i "spawn.*nav\|failed.*spawn\|K8s job" | head -20
   ```

4. **The `input_spec` for nav-updater is NULL.** This might not matter (the dispatch loop maps input_data via `input_mapping` in `call_handler`), but worth checking if `ensure_site_record` can handle the input shape the dispatch loop sends. The input_mapping sends:
   ```json
   {
     "spec": {"check":"orphan_pages","orphan_type":"nav_drift","fix":"..."},
     "domain": "<from input_data.domain>",
     "site_id": "<from current_item.site_id>",
     "item_type": "nav_drift",
     "work_item_id": "<uuid>",
     "current_page": {"check":"orphan_pages","orphan_type":"nav_drift","fix":"..."}
   }
   ```
   `ensure_site_record` needs `site_id` or `domain` — both are present, so this should work.

5. **The dispatch loop gets `domain` from `input_data.domain`** — but the dispatch loop's own input comes from `build-pipeline-trigger` which seeds with `site_id` and `domain`. If the trigger doesn't pass `domain`, the `call_handler` input_mapping for `domain` resolves to nil. Check:
   ```sql
   SELECT type, jsonb_pretty(default_config->'workflow') as workflow
   FROM agent_definitions
   WHERE type = 'build-pipeline-trigger';
   ```

---

## Still Pending from Prior Handoff

### orphan_blog_posts — should self-resolve after deploy
The updated `rebuild_blog_listing_action.go` is deployed. Sites with `orphan_blog_posts` items routed to `rerender-pages` will now write to the correct slot with a working template. Affected sites: finetuning.uk (10 posts), gamedesign.uk (6), vonc.com (1), leopardessconsulting.co.uk (2), robot-hands.com (1).

Check after a dispatch cycle:
```sql
SELECT s.domain, swi.status, swi.updated_at
FROM site_work_items swi
JOIN sites s ON swi.site_id = s.id
WHERE swi.item_type = 'orphan_blog_posts'
ORDER BY swi.updated_at DESC;
```

### deactivated_component (3) → rerender-pages
Wrong handler. Head `site_component` points to deactivated `content_component`. Re-rendering uses the broken link. Needs relinking to an active component.

### Tool items (11 failed) — component_id mapping bug
`input_data.component_id` resolves to nil because spec has `component_id` at `input_data.spec.component_id`. Deferred to separate work.

### needs_internal_links (future handler)
New item type from updated `check_orphan_pages.go`. Routed to `internal-linker` which doesn't exist. Items sit in `detected`.

### generic_theme (7), missing_style_collection (3), needs_design_review (4)
All routed to webdesign-agent. Not investigated yet.

---

## Key SQL Queries

```sql
-- Blog listing status after deploy
SELECT s.domain, p.name as blog_page,
       pc.slot_name, LENGTH(pc.rendered_html) as html_len,
       pc.content_data IS NOT NULL as has_content_data,
       pc.updated_at
FROM pages p
JOIN sites s ON p.site_id = s.id
JOIN page_components pc ON pc.page_id = p.id
WHERE p.page_type = 'blog-index'
  AND pc.slot_name IN ('blog-listing', 'article-grid', 'content-listing', 'guide-list', 'featured-article')
ORDER BY s.domain;

-- Orphan blog posts status
SELECT s.domain, swi.status, swi.item_type, swi.updated_at
FROM site_work_items swi
JOIN sites s ON swi.site_id = s.id
WHERE swi.item_type IN ('orphan_blog_posts', 'nav_drift')
ORDER BY swi.updated_at DESC;

-- Overall unresolved/failed by type
SELECT item_type, handler_agent, status, COUNT(*)
FROM site_work_items
WHERE status IN ('unresolved', 'failed', 'triaged')
  AND pipeline = 'build'
GROUP BY 1, 2, 3
ORDER BY 4 DESC;
```

---

## Files Modified This Session
- `platform/orchestration/actions/rebuild_blog_listing_action.go` — full rewrite (deployed)
- `content_components` table — content-listing template patched (applied)
- `site_work_items` table — orphan_page unresolved items marked wont_fix, 2 nav_drift items created
