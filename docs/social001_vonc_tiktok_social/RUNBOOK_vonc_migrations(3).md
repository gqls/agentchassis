# Runbook — vonc.com database migrations and operational fixes

## Before any migration: take a site snapshot

```sql
SELECT take_site_snapshot(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
    'pre-render-mode-migrations',
    NULL,
    'vonc render_mode fix session',
    'manual'
);
```

---

## Migration 001 — Mark broken-template components for regeneration

Targets components with `llm_field_count > 0` but `actual_template_slots = 0`.
These cannot receive LLM content and must be regenerated.

**Verify scope first (read-only):**
```sql
SELECT cc.id, cc.function, cc.quality_score,
       (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '{{.', '')))
           / LENGTH('{{.') AS actual_template_slots,
       (SELECT COUNT(*) FROM jsonb_each(cc.input_schema->'fields') f
        WHERE f.value->>'source' = 'llm') AS llm_field_count
FROM content_components cc
WHERE cc.is_active = true
  AND cc.component_level = 'section'
  AND cc.forked_from IS NULL
  AND cc.function IN ('gauntlet-cta', 'system-stats')
ORDER BY cc.function;
```

Confirm both rows show `actual_template_slots = 0` and `llm_field_count > 0`.

**No DB change needed for this migration** — it is handled by raising
`needs_component_regeneration` work items (see Plan P2). The component-creator
agent will regenerate and call StoreGeneratedComponentAction which now
(after code deployment) sets `render_mode = 'agent'` correctly.

---

## Migration 002 — Fix LLM routing condition in page-content-writer (REPLACES render_mode approach)

**Background:** The `check_render_mode` condition in the page-content-writer workflow reads
`current_section.component.render_mode == 'agent'`, but `plan_sections` never sets
`render_mode` on the section item — it sets `llm_field_specs` (populated from schema fields
with `source: "llm"`). The condition is therefore always false; every section routes to
`render_from_template` regardless of whether the component has LLM fields.

The fix is a one-line change to the agent_definition condition. No component table changes needed.
`render_mode` on content_components remains `'template'` for all components; `deriveRenderMode`
in `StoreGeneratedComponentAction` sets it correctly for newly created components, but the
workflow routing does not read it for section-level decisions.

**Backup first:**
```sql
SELECT snapshot_agent('page-content-writer', 'pre-llm-routing-fix');
```

**Verify current condition:**
```sql
SELECT default_config
    ->'workflow'->'steps'->'process_sections_loop'
    ->'config'->'sub_workflow'->'steps'->'check_render_mode'
    ->'config'->>'condition' AS current_condition
FROM agent_definitions
WHERE type = 'page-content-writer';
```
Expected: `current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true`

**Apply fix:**
```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,process_sections_loop,config,sub_workflow,steps,check_render_mode,config,condition}',
    '"current_section.llm_field_specs != null"'
),
updated_at = now()
WHERE type = 'page-content-writer';
```

**Verify:**
```sql
SELECT default_config
    ->'workflow'->'steps'->'process_sections_loop'
    ->'config'->'sub_workflow'->'steps'->'check_render_mode'
    ->'config'->>'condition' AS new_condition
FROM agent_definitions
WHERE type = 'page-content-writer';
```
Expected: `current_section.llm_field_specs != null`

**Rollback:**
```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,process_sections_loop,config,sub_workflow,steps,check_render_mode,config,condition}',
    '"current_section.component.render_mode == ''agent'' OR current_section.component.needs_llm == true"'
),
updated_at = now()
WHERE type = 'page-content-writer';
```

## Migration 003 — Unblock `pending` page_components for rerender

The `archetype-result-card` page_components on tool-gauntlet and
tool-archetype-taster-quiz have `build_status = 'pending'` after regeneration.
The rerender pipeline skips pending rows. This updates them to `deployed`
so the next rerender picks them up.

**Verify scope (read-only):**
```sql
SELECT pc.id, p.name AS page_name, pc.slot_name, pc.build_status,
       pc.updated_at
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'vonc.com'
  AND pc.build_status = 'pending';
```

**Apply migration:**
```sql
BEGIN;

UPDATE page_components pc
SET build_status = 'deployed',
    updated_at   = now()
FROM pages p
JOIN sites s ON s.id = p.site_id
WHERE pc.page_id = p.id
  AND s.domain = 'vonc.com'
  AND pc.build_status = 'pending';

-- Verify: 0 rows should remain pending for vonc.com
SELECT COUNT(*) AS still_pending
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'vonc.com'
  AND pc.build_status = 'pending';

COMMIT;
```

---

## Migration 004 — Trigger rerenders for affected pages

After migrations 002 and 003, trigger rerenders for pages that need
updated HTML. Use the correct spec shape (learned from session):

**Get page IDs first:**
```sql
SELECT id, name, url
FROM pages
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY name;
```

**Insert rerender work items (replace UUIDs from query above):**
```sql
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual', 'page_rerender',
    'medium', 'Rerender tool-gauntlet after render_mode fix',
    jsonb_build_object(
        'domain',    'vonc.com',
        'page_id',   '<tool-gauntlet page id>',
        'page_name', 'tool-gauntlet',
        'filename',  'tools/gauntlet/index.html'
    ),
    50, 'page-rerender', 'triaged', 'manual',
    'manual-rerender-tool-gauntlet-' || gen_random_uuid(), 'build'
),
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual', 'page_rerender',
    'medium', 'Rerender tool-archetype-taster-quiz after render_mode fix',
    jsonb_build_object(
        'domain',    'vonc.com',
        'page_id',   '<tool-archetype-taster-quiz page id>',
        'page_name', 'tool-archetype-taster-quiz',
        'filename',  'tools/archetype-taster-quiz/index.html'
    ),
    50, 'page-rerender', 'triaged', 'manual',
    'manual-rerender-tool-quiz-' || gen_random_uuid(), 'build'
);
```

Note: index page content cannot be fixed by rerender alone — it needs a
full `needs_page` rebuild because the hollow sections (provocation-card,
brief-explanation, lobby-grid) were rendered with render_context not LLM
content. See Plan P3.

---


## vonc.com page ID reference

| name | id | url |
|---|---|---|
| about | a28abcd7-186b-4a33-9b89-5d7bfd727012 | /about.html |
| archetypes | 2d0fd96a-59ca-4941-9e32-331f0f15314d | /archetypes.html |
| contact | 56f049fb-3ffe-49ad-b5fa-f6a87edfcb26 | /contact.html |
| index | b4d24f8e-fccd-49df-9dad-aa56a0b20a68 | /index.html |
| provocation | f204e18f-49a9-4dc0-8457-571a9deaeb65 | /blog/provocation.html |
| provocations-index | e4b3b195-919f-45ad-854e-201d3e846ea8 | /provocations/index.html |
| tool-archetype-taster-quiz | f1bc679f-5c48-46e8-9bb5-76cb8cf99ca5 | /tools/archetype-taster-quiz/index.html |
| tool-gauntlet | ecb637c1-845f-46bf-b174-9c92a43f9586 | /tools/gauntlet/index.html |

site_id: `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`
snapshot taken: `044a0b57-39b4-4221-86fa-bbbb2c4df17c` (pre-render-mode-migrations)

## Manual rerender spec shape (reference)

The `page-rerender` handler requires `page_id`, not just `page_name`.
Template from a working pipeline-generated item:

```json
{
    "domain":    "vonc.com",
    "page_id":   "<uuid from pages.id>",
    "page_name": "<pages.name>",
    "filename":  "<pages.url without leading slash>"
}
```

Example: `{"domain":"vonc.com","page_id":"e4b3b195-...","filename":"provocations/index.html","page_name":"provocations-index"}`

---

## Checking work item status

```sql
SELECT item_key, item_type, status, error,
       claimed_at, completed_at
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY created_at DESC
LIMIT 20;
```

---

## Code deliverables this session

All in `/mnt/user-data/outputs/`:

| File | Change |
|---|---|
| `site_spec_actions.go` | String coercion for spec_data |
| `store_generated_component_action.go` | `deriveRenderMode` helper; INSERT and UPDATE use it |
| `check_component_standards.go` | `checkBrokenTemplateSlots` sub-check |
| `fix_component_template_action.go` | `repair_template_slots` fix type; Mode A/B detection |
