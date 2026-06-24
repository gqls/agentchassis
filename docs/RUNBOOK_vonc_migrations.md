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

## Migration 002 — Fix `render_mode` on components with LLM fields and good templates

Targets components that have both `actual_template_slots > 0` AND
`llm_field_count > 0`. These have correct templates but wrong render_mode.
After this migration `check_render_mode` will route them to LLM generation.

**Take snapshot first** (see above).

**Verify scope (read-only):**
```sql
SELECT cc.id, cc.function,
       (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '{{.', '')))
           / LENGTH('{{.') AS actual_template_slots,
       (SELECT COUNT(*) FROM jsonb_each(cc.input_schema->'fields') f
        WHERE f.value->>'source' = 'llm') AS llm_field_count
FROM content_components cc
WHERE cc.is_active = true
  AND cc.component_level = 'section'
  AND cc.forked_from IS NULL
  AND cc.render_mode != 'agent'
  AND (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '{{.', '')))
          / LENGTH('{{.') > 0
  AND (SELECT COUNT(*) FROM jsonb_each(cc.input_schema->'fields') f
       WHERE f.value->>'source' = 'llm') > 0;
```

Review the list. If it looks reasonable (should include hero, gauntlet-interface,
archetype-result-card, tool-archetype-taster-quiz and possibly others):

**Apply migration:**
```sql
BEGIN;

-- Backup: record current render_mode in a temp table for rollback reference
CREATE TEMP TABLE render_mode_backup AS
SELECT id, function, render_mode
FROM content_components
WHERE is_active = true
  AND component_level = 'section'
  AND forked_from IS NULL;

-- Apply: set render_mode = 'agent' where templates have slots AND schema has LLM fields
UPDATE content_components cc
SET render_mode = 'agent',
    updated_at  = now()
WHERE cc.is_active = true
  AND cc.component_level = 'section'
  AND cc.forked_from IS NULL
  AND cc.render_mode != 'agent'
  AND (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '{{.', '')))
          / LENGTH('{{.') > 0
  AND (SELECT COUNT(*) FROM jsonb_each(cc.input_schema->'fields') f
       WHERE f.value->>'source' = 'llm') > 0;

-- Verify: should show updated rows with render_mode = 'agent'
SELECT id, function, render_mode
FROM content_components
WHERE is_active = true
  AND component_level = 'section'
  AND forked_from IS NULL
  AND render_mode = 'agent'
ORDER BY function;

COMMIT;
```

**Rollback if needed:**
```sql
-- Use the temp table backup (only valid in same session before disconnect)
UPDATE content_components cc
SET render_mode = b.render_mode, updated_at = now()
FROM render_mode_backup b
WHERE cc.id = b.id;
```

---

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
