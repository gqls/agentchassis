# Plan — vonc.com remaining work

## Priority order

### P0 — Code deployments (unblock everything)

1. Deploy `site_spec_actions.go` — fixes `write_site_spec` string rejection.
   Required before any resubmission.

2. Deploy `store_generated_component_action.go` — fixes `render_mode`
   derivation. Required before any new component creation or regeneration
   produces correct routing.

3. Deploy `check_component_standards.go` and `fix_component_template_action.go`
   — discovery checker now detects broken templates; fixer handles both repair
   and regeneration routing.

### P1 — Database migrations (fix existing data)

Run migrations 001–004 in order (see runbook). Each migration is preceded
by a `take_site_snapshot` call and can be rolled back.

Migration 001: Mark broken components for regeneration.
Migration 002: Fix `render_mode` on components with LLM fields and good templates.
Migration 003: Unblock `pending` page_components for rerender.
Migration 004: Trigger rerenders for affected pages.

### P2 — Regenerate broken components

`gauntlet-cta` and `system-stats` have LLM fields but zero template slots.
After P0 code is deployed, raise `needs_component_regeneration` work items:

```sql
INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary, spec,
    priority, handler_agent, status, created_by, item_key, pipeline
) VALUES
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_component_regeneration', 'high',
    'Regenerate gauntlet-cta — 0 template slots, 16 LLM fields',
    '{"function":"gauntlet-cta","component_id":"<id>","quality_score":30,
      "quality_issues":["0 template variables"],"section_type":"gauntlet-cta"}',
    5, 'component-creator', 'triaged', 'manual',
    'manual-regen-gauntlet-cta-' || gen_random_uuid(), 'build'
),
(
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_component_regeneration', 'high',
    'Regenerate system-stats — 0 template slots, 24 LLM fields',
    '{"function":"system-stats","component_id":"<id>","quality_score":30,
      "quality_issues":["0 template variables"],"section_type":"system-stats"}',
    5, 'component-creator', 'triaged', 'manual',
    'manual-regen-system-stats-' || gen_random_uuid(), 'build'
);
```
Get the component IDs first:
```sql
SELECT id, function FROM content_components
WHERE function IN ('gauntlet-cta', 'system-stats') AND is_active = true;
```

### P3 — Rebuild index page content

After P1 migrations run and render_mode is correct, the index page sections
that used `render_from_template` (provocation-card, brief-explanation,
lobby-grid) need a full content rebuild via `needs_page` work item:

```sql
-- Get page_id first
SELECT id FROM pages WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND name = 'index';

INSERT INTO site_work_items (
    site_id, source, item_type, severity, summary, spec,
    priority, handler_agent, status, created_by, item_key, pipeline
) VALUES (
    '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'manual',
    'needs_page', 'high', 'Rebuild index page with correct render_mode routing',
    jsonb_build_object(
        'domain', 'vonc.com',
        'page_id', '<id>',
        'page_name', 'index',
        'filename', 'index.html'
    ),
    5, 'page-build-handler', 'triaged', 'manual',
    'manual-rebuild-index-' || gen_random_uuid(), 'build'
);
```

### P4 — Structural fixes for the codebase (not vonc-specific)

These should be filed as separate work items / PRs:

1. `StoreGeneratedComponentAction` `template_variable_count` is not updated
   on the UPDATE (regeneration) path — only on INSERT. After regeneration
   it stays at the old value. Add `template_variable_count = <count>` to the
   UPDATE SET clause.

2. The `fix_component_template_action.go` file header comment should list
   `repair_template_slots` as a supported fix type.

3. Wider library audit: run `checkBrokenTemplateSlots` across all sites
   to find other components with Mode A or Mode B artifacts.

### P5 — Ongoing (not blocking build)

- Remove stale `provocations.html` from repo root.
- `unresolved_cta` items (index hero, archetypes) — will self-resolve when
  section-index hubs become active.
- `needs_section_data` items — require manual content input.
