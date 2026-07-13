# Site Work Orchestrator — Data Flow Trace

Step-by-step data flow verification for phase 0 (initial build).

## 1. spawn_planner → spawn_content_writer → spawn_reviewer
Standard spawns. Creates Kafka topics for communication.
- `planner_agent`, `content_writer_agent`, `reviewer_agent` in collectedData

## 2. ensure_site_record
- Input: `input_data` (from message body: domain, objective, reviewed_brief)
- Output: `site_record` → `{site_id, domain, name, status, ...}`

## 3. call_site_planner
- Calls existing `site-planner` agent (unchanged)
- Input mapping: `input_data, site_record, input_data.reviewed_brief`
- Output: `site_plan` → `{response: {plan_data: {pages: [...], needs_logo, needs_images, ...}}}`
- ✅ Same planner, same output format as pageflow-builder

## 4. store_reviewed_brief
- Merges `input_data.reviewed_brief` into `sites.content_data`

## 5. store_site_plan
- Merges `site_plan` into `sites.content_data`

## 6. sync_pages_to_db
- Creates page records in `pages` table from site_plan
- Output: `db_sync` → `{pages: [{id, name, url, ...}], ...}`
- ✅ Pages now exist in DB with UUIDs

## 7. write_build_items  ← NEW
- Queries `pages` table for pages with build_status='planned'
- Each page becomes a work item in `site_work_items`
- `spec` = full page record (id, name, url, title, sections, etc.)
- Also creates tracking items for logo, hero, design (for audit)
- Output: `build_items_written` → `{items_inserted: N, batch_id: uuid}`

## 8. populate_nav
- Builds nav tables from page records (same as pageflow-builder)

## 9-16. Asset generation (logo, hero) + style collection + defaults + render
- Identical to pageflow-builder steps
- Assets handled via explicit workflow, not via queue

## 17. load_work_items  ← NEW
- Queries: `site_work_items WHERE site_id=$1 AND status IN ('triaged','approved') AND domain='build' AND handler_agent='page-content-writer'`
- Config uses `item_domain` (not `domain`) to avoid collision with `site_record.domain`
- Config uses `handler_agent` and `max_items` read directly from config (not via ExtractActionInputs)
- Respects dependency ordering
- Output: `work_items` → `{items: [{id, spec: {name, id, url, title, sections, ...}, handler_agent, ...}], has_items: true}`
- ✅ `spec` contains full page record because write_build_items stored it from pages table

## 18. check_has_items
- Condition: `work_items.has_items == true`

## 19. build_items_loop — per item:

### 19a. write_page_content
- Calls `page-content-writer`
- Input mapping: `current_page: current_item.spec`
    - `current_item.spec.name` → page name ✅
    - `current_item.spec.id` → page DB UUID ✅
    - `current_item.spec.title` → page title ✅
    - `current_item.spec.sections` → sections array ✅
    - `current_item.spec.url` → page URL ✅
- Also passes: `db_sync, site_plan, site_record, reviewed_brief, style_collection`
- ✅ page-content-writer sees identical data to what pageflow-builder provides

### 19b. review_page_content
- Calls `content-reviewer`
- Input mapping: `current_page: current_item.spec`
- ✅ Same data shape

### 19c. check_review_approved
- If rejected → fail_item → loop_complete

### 19d. assemble_page
- `content_field: "page_content.response.page_html"`
- ✅ Same as pageflow-builder

### 19e. deploy_page (git_commit)
- `page_field: "current_item.spec"` → uses spec.name and spec.url for filename
- `domain_field: "site_record.domain"`
- `content_field: "assembled_page.html"`
- Output: `page_deployed` → `{commit_sha: "abc123", ...}`
- ✅ GitHub Action fires → S3 → live

### 19f. save_sections
- `page_name_field: "current_item.spec.name"`
- ✅ Same field path

### 19g. update_page_status
- `page_id_field: "current_item.spec.id"` → DB page UUID
- `commit_from: "page_deployed.commit_sha"`
- ✅ Page marked as deployed

### 19h. complete_work_item  ← NEW
- `work_item_id: "current_item.id"` → work item UUID
- `commit_sha: "page_deployed.commit_sha"`
- Marks item as 'complete' with result containing commit SHA

## 20. apply_site_design
- Calls `webdesign-agent` (same as pageflow-builder)

## 21. update_site_status
- Marks site as deployed

## 22. complete

---

## Key difference from pageflow-builder

The only structural difference is steps 7, 17, and 19h:

| Step | pageflow-builder | site-work-orchestrator |
|------|-----------------|----------------------|
| Source of pages | `get_pages_to_build` (queries pages table) | `load_work_items` (queries site_work_items, spec has page data) |
| Loop variable | `current_page` (page record) | `current_item` (work item, `.spec` = page record) |
| Post-deploy | update_page_status only | update_page_status + complete_work_item |
| Tracking | pages.build_status | pages.build_status + site_work_items.status |

The content writer, reviewer, assembler, git commit, and save sections are all identical.

---

## Audit Notes (vs 008_checklist_for_new_specialist_agents_v5)

### Fixed during audit

1. **`domain` field collision** — `LoadWorkItemsInputSpec` originally had `domain` as Optional.
   `site_record.domain` would be found via nested lookup, giving "example.com" instead of "build".
   Fix: Renamed config key to `item_domain`, moved to direct config read (not ExtractActionInputs).

2. **Literal config values treated as paths** — `handler_agent: "page-content-writer"` and
   `error_message: "Content review not approved"` would be treated as paths by ExtractActionInputs,
   failing to resolve. Fix: Read these directly from `params.StepConfig.Config`.

3. **`error_message` collision risk** — Removed from FailWorkItemInputSpec Optional list.
   Read directly from config instead.

### Accepted divergences from checklist

- **page-content-writer receives assembled context** — The checklist says callers should pass raw
  identifiers. Our orchestrator sends db_sync, site_plan, style_collection etc. This is the
  existing pageflow-builder pattern. Changing it requires rewriting page-content-writer with a
  load action — flagged for future refactoring.

- **docker_image/docker_tag in SQL** — Checklist says "should NOT include". All existing agent
  definitions include them. Table schema requires them. spawn_actions.go overrides at runtime.

### Config value patterns in this orchestrator

| Config key | Value | Type | Read via |
|-----------|-------|------|----------|
| `site_id` | `"site_record.site_id"` | path | ExtractActionInputs |
| `site_plan` | `"site_plan"` | path | ExtractActionInputs |
| `work_item_id` | `"current_item.id"` | path | ExtractActionInputs |
| `commit_sha` | `"page_deployed.commit_sha"` | path | ExtractActionInputs |
| `item_domain` | `"build"` | literal | direct config read |
| `handler_agent` | `"page-content-writer"` | literal | direct config read |
| `max_items` | `20` | literal | direct config read |
| `error_message` | `"Content review not approved"` | literal | direct config read |