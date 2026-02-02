# Selective Page Rebuild Guide

## Overview

The pageflow-builder uses two status columns to control which pages get rebuilt:

| Column | Purpose | Values |
|--------|---------|--------|
| `status` | Page lifecycle | `active`, `deleted`, `needs_attention` |
| `build_status` | Build state | `planned`, `needs_rebuild`, `deployed` |

The `get_pages_to_build` action filters by `build_status`, not `status`.

## How It Works

```sql
-- pageflow-builder's query (simplified)
SELECT * FROM pages 
WHERE site_id = $1 
  AND status = 'active'
  AND build_status IN ('planned', 'needs_rebuild')
```

Pages with `build_status = 'deployed'` are skipped.

## Triggering Selective Rebuilds

### Step 1: Mark pages for rebuild

```sql
-- Mark specific pages by name
UPDATE pages 
SET build_status = 'needs_rebuild'
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND name IN ('use-cases', 'insights', 'careers', 'privacy', 'terms');

-- Or mark pages missing stored sections
UPDATE pages 
SET build_status = 'needs_rebuild'
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND id NOT IN (
    SELECT DISTINCT page_id 
    FROM page_components 
    WHERE rendered_html IS NOT NULL
  );
```

### Step 2: Verify what will be rebuilt

```sql
SELECT name, build_status,
       (SELECT COUNT(*) FROM page_components pc 
        WHERE pc.page_id = p.id AND pc.rendered_html IS NOT NULL) as sections
FROM pages p
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
ORDER BY nav_order;
```

### Step 3: Trigger pageflow-builder

```bash
# Direct trigger (skips intake-orchestrator's planner)
./agentctl send pageflow-builder process \
  --data '{"site_id": "UUID", "domain": "example.com"}'
```

## What Happens During Rebuild

1. `get_pages_to_build` queries pages with `build_status IN ('planned', 'needs_rebuild')`
2. Each page goes through: content-writer → reviewer → save_sections → deploy
3. On successful deploy, `build_status` is set to `'deployed'`

## Common Scenarios

| Scenario | SQL |
|----------|-----|
| Rebuild all pages | `UPDATE pages SET build_status = 'needs_rebuild' WHERE site_id = ?` |
| Rebuild one page | `UPDATE pages SET build_status = 'needs_rebuild' WHERE site_id = ? AND name = 'about'` |
| Rebuild pages without sections | See "Mark pages missing stored sections" above |
| Reset failed builds | `UPDATE pages SET build_status = 'planned' WHERE build_status = 'error'` |

## Notes

- Rebuilding a page regenerates content via LLM - results may differ from original
- The `save_sections` step stores rendered HTML in `page_components.rendered_html`
- Rerender (maintenance) only works for pages with stored sections