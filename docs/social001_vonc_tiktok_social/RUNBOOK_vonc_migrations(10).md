# Runbook — vonc.com component rendering fix

**Created:** 2026-06-22  
**Last updated:** 2026-06-24 ~16:50

---

## What this runbook is about

During the first full build of vonc.com (the Spark social platform), several
structural bugs were found in the component rendering pipeline that affected not
just vonc.com but the entire component library. This runbook records what was
broken, what was fixed, and the remaining steps to complete.

The core problem: every page section was being rendered from its HTML template
with no LLM content generation, even for sections that have rich content schemas.
The workflow's routing condition read a field (`render_mode == 'agent'`) that was
never populated by the upstream step — so every section silently fell through to
the template-only path. The fix is a one-line change to the `page-content-writer`
agent definition. A secondary problem was that several component templates were
stored as rendered output (with resolved `<no value>` strings) rather than as
proper Go templates with `{{.field}}` slots.

This runbook covers the database migrations, component regenerations, and
operational steps needed to complete the vonc.com build and leave the pipeline
in a correct state for all future sites.

---

## Quick-reference: vonc.com identifiers

**site_id:** `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`  
**Snapshot (pre-migrations):** `044a0b57-39b4-4221-86fa-bbbb2c4df17c`

| page | id | url |
|---|---|---|
| index | b4d24f8e-fccd-49df-9dad-aa56a0b20a68 | /index.html |
| about | a28abcd7-186b-4a33-9b89-5d7bfd727012 | /about.html |
| archetypes | 2d0fd96a-59ca-4941-9e32-331f0f15314d | /archetypes.html |
| contact | 56f049fb-3ffe-49ad-b5fa-f6a87edfcb26 | /contact.html |
| provocation | f204e18f-49a9-4dc0-8457-571a9deaeb65 | /blog/provocation.html |
| provocations-index | e4b3b195-919f-45ad-854e-201d3e846ea8 | /provocations/index.html |
| tool-archetype-taster-quiz | f1bc679f-5c48-46e8-9bb5-76cb8cf99ca5 | /tools/archetype-taster-quiz/index.html |
| tool-gauntlet | ecb637c1-845f-46bf-b174-9c92a43f9586 | /tools/gauntlet/index.html |

| component | id | issue |
|---|---|---|
| gauntlet-cta | 66bfd4a4-2163-4d34-be43-c42ee17e6af0 | 0 template slots, needs regeneration |
| system-stats | fdd92ad4-521a-4602-89cf-7ee1a66c10f1 | 0 template slots, needs regeneration |
| gauntlet-interface | 5da50747-7936-4b8f-a66d-c1ea98919c75 | template slots fixed (Mode A repair) |
| archetype-result-card | 2c7678fb-9940-428d-8b78-62e2510f6dbe | regenerated, quality 100 |

---

## Checklist — overall status

- [x] **Snapshot taken** (2026-06-23) — `044a0b57-39b4-4221-86fa-bbbb2c4df17c`
- [x] **Migration 001** — `write_site_spec` coercion fix — code delivered, needs deploy
- [x] **Migration 002** — LLM routing condition fixed in `page-content-writer` agent_definition
- [x] **Migration 003** — `pending` page_components unblocked (was no-op, already clear)
- [x] **Migration 004** — tool page rerenders queued and completed
- [x] **Index rebuild** — `needs_page` work item ran, hero content correct
- [x] **Deploy code** — deployed 2026-06-24 ~15:00
- [x] **Regenerate `gauntlet-cta`** — complete 2026-06-24 15:06 — quality 100, 20 slots
- [x] **Regenerate `system-stats`** — complete 2026-06-24 15:06 — quality 100, 22 slots
- [x] **Check tool page renders** — all 4 page_components `deployed` with rendered HTML ✓
- [ ] **Rebuild index** — `needs_page` queued 2026-06-24 ~15:20, awaiting completion
- [ ] **Verify index** — check deployed HTML has gauntlet-cta and system-stats content
- [x] **CSS variable names fixed** — `--primary-color`/`--secondary-color` → `--color-primary`/`--color-secondary` in hero template (2026-06-24 ~16:30)
- [x] **component-creator prompt patched** — CSS variable naming rule added (2026-06-24 ~16:50)
- [ ] **Verify index** after hero fix rerender completes
- [ ] **Remove stale `provocations.html`** from repo root (predates this submission)
- [ ] **Duplicate work items** — clean up duplicate `page_rerender` items in queue (see below)

---

## Step 1 — Deploy code (prerequisite for all remaining steps)

Four files in `/mnt/user-data/outputs/` need deploying:

| file | what changed |
|---|---|
| `site_spec_actions.go` | `WriteSiteSpecAction` now accepts string `spec_data` — wraps plain text as `{"text": value}`, parses JSON strings to objects. Fixes `mission_brief`/`roadmap_brief` submission failures. |
| `store_generated_component_action.go` | Added `deriveRenderMode()` helper. Both INSERT and UPDATE paths now derive `render_mode` from schema (`"agent"` if any field has `source: "llm"`, else `"template"`) instead of hardcoding `"template"`. |
| `check_component_standards.go` | Added `checkBrokenTemplateSlots` sub-check — detects components with `<no value>` artifacts in `html_template` and raises `broken_template_slots` work items. |
| `fix_component_template_action.go` | Added `repair_template_slots` fix type with `repairNoValueSlots` helper (Mode A: repairable via string sub) and Mode B detection (no `</no>` tags → returns `needs_regeneration`, does not attempt repair). |

**Verify deploy by checking `store_generated_component_action.go` has `deriveRenderMode`:**
```bash
grep -n "deriveRenderMode" platform/orchestration/actions/store_generated_component_action.go
```
Expected: three hits (UPDATE call site, INSERT call site, function definition).

---

## Step 2 — Regenerate `gauntlet-cta` and `system-stats`

Both have rich LLM schemas (16 and 24 fields) but zero template slots — their
templates were stored as rendered output with all content erased. After code
deployment, `component-creator` will generate fresh templates and
`StoreGeneratedComponentAction` will set `render_mode = 'agent'` automatically.

**Verify they still need regeneration:**
```sql
SELECT cc.function, cc.template_variable_count,
       (LENGTH(cc.html_template) - LENGTH(REPLACE(cc.html_template, '{{.', '')))
           / LENGTH('{{.') AS actual_slots,
       (SELECT COUNT(*) FROM jsonb_each(cc.input_schema->'fields') f
        WHERE f.value->>'source' = 'llm') AS llm_fields
FROM content_components cc
WHERE cc.function IN ('gauntlet-cta', 'system-stats');
```
Expected: `actual_slots = 0`, `llm_fields > 0` for both.

**Raise regeneration work items:**
```sql
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_component_regeneration', 'high',
    'Regenerate gauntlet-cta — 0 template slots, 16 LLM fields',
    jsonb_build_object(
        'function',     'gauntlet-cta',
        'component_id', '66bfd4a4-2163-4d34-be43-c42ee17e6af0',
        'quality_score', 30,
        'quality_issues', '["0 template variables — stored as rendered output"]',
        'section_type', 'gauntlet-cta'
    ),
    5, 'component-creator', 'triaged', 'manual',
    'manual-regen-gauntlet-cta-' || gen_random_uuid(), 'build'
),
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_component_regeneration', 'high',
    'Regenerate system-stats — 0 template slots, 24 LLM fields',
    jsonb_build_object(
        'function',     'system-stats',
        'component_id', 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1',
        'quality_score', 30,
        'quality_issues', '["0 template variables — stored as rendered output"]',
        'section_type', 'system-stats'
    ),
    5, 'component-creator', 'triaged', 'manual',
    'manual-regen-system-stats-' || gen_random_uuid(), 'build'
);
```

**Monitor until complete:**
```sql
SELECT item_key, status, error, completed_at
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND item_type = 'needs_component_regeneration'
ORDER BY created_at DESC
LIMIT 5;
```

**Verify regeneration succeeded:**
```sql
SELECT function, quality_score, template_variable_count,
       (LENGTH(html_template) - LENGTH(REPLACE(html_template, '{{.', '')))
           / LENGTH('{{.') AS actual_slots,
       (LENGTH(html_template) - LENGTH(REPLACE(html_template, '<no value>', '')))
           / LENGTH('<no value>') AS no_value_count
FROM content_components
WHERE function IN ('gauntlet-cta', 'system-stats');
```
Expected: `actual_slots > 0`, `no_value_count = 0`, `quality_score >= 80`.

---

## Step 3 — Rebuild index page

After gauntlet-cta and system-stats regenerate, the index page needs a full
rebuild so those sections receive LLM-generated content. A rerender alone won't
do it — the page_components need fresh content_data written by page-content-writer.

```sql
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES (
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_page', 'high',
    'Rebuild index after gauntlet-cta and system-stats regeneration',
    jsonb_build_object(
        'domain',    'vonc.com',
        'page_id',   'b4d24f8e-fccd-49df-9dad-aa56a0b20a68',
        'page_name', 'index',
        'filename',  'index.html'
    ),
    5, 'page-build-handler', 'triaged', 'manual',
    'manual-rebuild-index-postgeneregen-' || gen_random_uuid(), 'build'
);
```

---

## Step 4 — Check tool page renders

Confirm tool-gauntlet and tool-archetype-taster-quiz rendered correctly after the
archetype-result-card regeneration and the routing fix.

```sql
SELECT p.name, pc.slot_name, pc.build_status,
       LEFT(pc.rendered_html, 150) AS html_preview
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'vonc.com'
  AND p.name IN ('tool-gauntlet', 'tool-archetype-taster-quiz')
ORDER BY p.name, pc.position;
```

All rows should show `build_status = 'deployed'` and `html_preview` starting
with component CSS rather than empty content.

If any slot still shows empty rendered_html, trigger individual rerenders
using the correct spec shape (see below).

---

## Step 5 — Remove stale `provocations.html` from repo root

A `provocations.html` file from a previous run exists at the repo root.
The current site uses `/provocations/index.html` (section-index URL).
Delete the stale file directly from the git repo to avoid confusion.

---


## Step 6 — Fix CSS variable injection

**Resolved (2026-06-24 ~16:30).** Root cause was CSS variable name mismatch in
the hero component template, not a CSS injection failure.

`styles.css` was correct — deployed 2 days ago by webdesign-agent with the right
palette values (`--color-primary: #7c3cff`, `--color-accent: #fc5c7d` etc.).
The `/* Theme-specific styles injected here: */` empty block is expected — the
post-025 renderer reads composition via FKs at render time; `css_content` is
intentionally not stored.

**Actual problem:** Hero template used `--primary-color` and `--secondary-color`
(LLM-generated names) instead of `--color-primary` and `--color-secondary`
(the system convention). These variables don't exist in styles.css so the fallback
hex values fired instead.

**Fix applied:** `REPLACE` in `content_components` for hero function.
**Rerender queued:** index page rerender after fix.
**Library-wide scan:** only `archetype-grid` has a potentially non-standard variable
(`--archetype-color`) — confirmed intentional: it is a per-card tinting variable with `--color-accent` as fallback. No fix needed. Library CSS variable audit complete.

**Diagnose — check resolved_composition:**
```sql
SELECT data->>'css_theme_id' AS css_theme_id,
       data->>'palette_id'   AS palette_id,
       data->>'layout_name'  AS layout_name
FROM site_specs
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND aspect = 'resolved_composition'
  AND is_current = true;
```

**Diagnose — check the CSS theme content:**
```sql
SELECT name, LEFT(css_content, 600) AS css_preview
FROM css_themes
WHERE id = '<css_theme_id from above>';
```
Expected: `:root { --color-primary: #7c3cff; --color-background: #0a0a0f; ... }`

**Diagnose — check styles.css in the deployed repo:**
Look at `/vonc.com/assets/css/styles.css` in the git repo.
If it exists and has the correct variables, the problem is in the HTML template injection.
If it is missing or empty, the asset deployer didn't write it.

**Design spec palette (from `design_intent` site spec):**
```
--color-primary:    #7c3cff
--color-secondary:  #2d1b69
--color-accent:     #fc5c7d
--color-background: #0a0a0f
--color-surface:    #13121f
--color-text:       #f0eeff
--color-text-muted: #8b85b0
--color-border:     #2a2640
```

## Duplicate work item cleanup (2026-06-24 ~15:20)

A full set of 8 `page_rerender` items was inserted twice — once from the runbook
steps and once from a separate SQL document. Clean up before they all run:

**See duplicates:**
```sql
SELECT spec->>'page_name' AS page, COUNT(*) AS count,
       array_agg(id::text ORDER BY created_at) AS ids
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND item_type = 'page_rerender'
  AND status = 'triaged'
GROUP BY spec->>'page_name'
HAVING COUNT(*) > 1;
```

**Delete the older of each duplicate pair:**
```sql
DELETE FROM site_work_items
WHERE id IN (
    SELECT (array_agg(id ORDER BY created_at))[1]
    FROM site_work_items
    WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
      AND item_type = 'page_rerender'
      AND status = 'triaged'
    GROUP BY spec->>'page_name'
    HAVING COUNT(*) > 1
);
```

**Verify clean:**
```sql
SELECT spec->>'page_name' AS page, status, created_at
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND item_type = 'page_rerender'
  AND status = 'triaged'
ORDER BY spec->>'page_name';
```

---

## Reference — manual rerender spec shape

The `page-rerender` handler requires `page_id` (UUID), not just `page_name`.
Always look up the UUID from the pages table and embed it inline — never use
placeholder strings.

```sql
-- Look up page_id first
SELECT id, name, url FROM pages
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY name;

-- Then insert — use the real UUID directly
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key, pipeline
) VALUES (
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual', 'page_rerender',
    'medium', 'Manual rerender of <page>',
    jsonb_build_object(
        'domain',    'vonc.com',
        'page_id',   '<id from pages table — paste directly>',
        'page_name', '<name>',
        'filename',  '<url without leading slash>'
    ),
    50, 'page-rerender', 'triaged', 'manual',
    'manual-rerender-<page>-' || gen_random_uuid(), 'build'
);
```

**DO NOT** use placeholder strings like `<tool-gauntlet id>`. They will be
claimed and fail before you can fix them. If this happens by mistake: check
`SELECT spec->>'page_id' FROM site_work_items WHERE ...` — if the placeholder
row is still present, DELETE it filtering on the placeholder string. If
`DELETE 0`, the retry mechanism purged it; just reinsert with the real UUID.

---

## Reference — checking work item status

```sql
SELECT item_type, spec->>'page_name' AS page, status, error,
       claimed_at, completed_at
FROM site_work_items
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
ORDER BY created_at DESC
LIMIT 20;
```

---

## Reference — taking a new snapshot before significant changes

```sql
SELECT take_site_snapshot(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
    'pre-<description>',
    NULL,
    '<human readable label>',
    'manual'
);
```

---

## Background — completed migrations (for reference)

### Migration 001 (2026-06-23) — `write_site_spec` string coercion
Code fix in `site_spec_actions.go`. Fixes "spec_data must be a JSON object, got string"
errors when `mission_brief`/`roadmap_brief` are submitted as plain text strings.
Status: code delivered, deploy pending.

### Migration 002 (2026-06-24 ~11:30) — LLM routing condition
**DONE.** Changed `check_render_mode` condition in `page-content-writer`
agent_definition from:
```
current_section.component.render_mode == 'agent' OR current_section.component.needs_llm == true
```
to:
```
current_section.llm_field_specs != null
```
The old condition read a field never populated by `plan_sections`. The new condition
reads `llm_field_specs`, which `plan_sections` populates from the component schema
for any field with `source: "llm"`. Took effect immediately for all sites.

### Migration 003 (2026-06-24) — Unblock pending page_components
Was a no-op — the pending rows were already cleared before the migration ran.

### Migration 004 (2026-06-24) — Tool page rerenders
**DONE.** tool-gauntlet and tool-archetype-taster-quiz rerender work items
inserted with correct UUIDs. (Earlier attempt failed due to placeholder strings
in page_id — see placeholder UUID warning above.)

### Index rebuild (2026-06-24 ~11:46)
**DONE.** `needs_page` work item for index claimed and completed. Hero section
now has correct Spark-branded LLM content. provocation-card, lobby-grid,
brief-explanation confirmed as intentionally static (empty schemas, JS-populated
at runtime by daily provocation pipeline).

### gauntlet-interface template repair (2026-06-23)
**DONE.** Mode A repair — `<no value>FIELD</no>` → `{{.FIELD}}` via SQL regexp.
Component id: `5da50747-7936-4b8f-a66d-c1ea98919c75`.

### archetype-result-card regeneration (2026-06-23)
**DONE.** Mode B failure (bare `<no value>`, field names lost). Regenerated via
`needs_component_regeneration` work item. Quality now 100, 28 template variables.
Component id: `2c7678fb-9940-428d-8b78-62e2510f6dbe`.
