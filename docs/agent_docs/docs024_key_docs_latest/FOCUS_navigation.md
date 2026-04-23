# Focus: Navigation — Build, Technical Workings, Discovery & Fix

This document pulls together everything related to site navigation from across the system. Use it as a single reference when debugging or improving navigation.

---

## Current State: What's Wrong

Sites frequently have bad, duplicated, or wrong navigation. The symptoms:

- Pages from previous builds appear in nav (stale pages not deactivated)
- Newly built pages missing from nav (build_status not checked)
- Nav links use anchor slugs (`#about`) instead of page URLs (`/about.html`)
- Orphan pages exist with no inbound nav links
- Tool pages added but nav entry creation silently fails (wrong column names)
- Fallback header renders with stacked nav (no flex CSS) and a search icon nobody wants
- Header/footer components not linked to style collection → generic fallback

---

## 1. How Navigation Gets Built

### Database Tables

| Table | Purpose | Key columns |
|---|---|---|
| `site_nav_groups` | Semantic categories (primary, utility, legal, footer) | `site_id`, `name`, `location`, `sort_order` |
| `site_nav_items` | Individual nav entries | `nav_group_id`, `label`, `url`, `page_id`, `sort_order` |
| `pages` | Page records with nav flags | `in_header`, `in_footer`, `build_status`, `status`, `nav_order` |

### Two Nav Systems (and they conflict)

**System 1: Nav tables** (`site_nav_groups` + `site_nav_items`) — the intended system. Populated by `populate_nav_tables` action. Used by `GetNavItems()` which is the shared query for header/footer rendering.

**System 2: Pages table flags** (`in_header`, `in_footer`) — the legacy fallback. `GetHeaderNavFromPages` and `GetFooterNavFromPages` read directly from the pages table. Used when nav tables are empty or as a fallback.

`GetNavItems()` tries nav tables first, falls back to pages table. This means if nav tables are partially populated, you get a mix.

### Build-Time Nav Flow

```
site-planner creates page records
  → SyncPagesToDBAction upserts pages (ON CONFLICT by site_id + name)
  → populate_nav_tables classifies pages into nav groups
  → render_site_components renders header with nav data
  → GetNavItems() reads nav tables (or falls back to pages)
  → Header template receives {{.nav_items}} or {{.nav_items_html}}
```

### The Stale Pages Problem (from 001_config_driven)

`SyncPagesToDBAction` uses `ON CONFLICT (site_id, name)` — it only overwrites matching page names. Pages from a prior build (e.g. `use-cases`, `insights`, `careers`) remain with `in_header=true`, `status='active'` and appear in nav even though the current plan doesn't include them.

**Fixes needed:**

1. `GetHeaderNavFromPages` and `GetFooterNavFromPages`: add `AND build_status = 'deployed'` to WHERE clause
2. `SyncPagesToDBAction`: before syncing, deactivate stale pages: `UPDATE pages SET in_header = false, in_footer = false WHERE site_id = $1 AND name != ALL($2)`

**Design decision:** New build flow deactivates stale pages. Maintenance/adopt-site flows should NOT deactivate — they preserve existing content. Use `deactivate_stale_pages: true` config flag.

### Header/Footer Rendering

The header component is rendered by `renderAndStoreSiteComponent`. It needs:

1. A `site_components` row for slot `header` with a valid `component_id`
2. That `component_id` points to a `content_components` template (e.g. `header-professional-dark`)
3. The template uses `{{.nav_items}}` or `{{.nav_items_html}}` — NOT hardcoded links
4. The style collection has `header_component_id` set

**Fallback chain when things are missing:**

```
site_components.component_id IS NOT NULL
  → YES: load template from content_components → render with nav data → store
  → NO: try generic lookup WHERE function = slot_name → usually fails
    → RenderFallbackHeader() — hardcoded Go function
      → generic HTML, no logo, stacked nav, search icon
```

The fallback header is the source of stacked nav links and unwanted search icons.

### Template Data Available to Header/Footer

| Variable | Source | Type |
|---|---|---|
| `{{.nav_items}}` | pages table via GetNavItems | slice of NavItem |
| `{{.nav_items_html}}` | pre-rendered `<li>` HTML | string |
| `{{.company_name}}` | sites table | string |
| `{{.logo_url}}` | sites table | string |
| `{{.primary_color}}` | style collection color_palette | string |
| `{{.theme_css}}` | css_themes.css_content | string |

### Nav Item Resolution

`GetNavItems()` returns items with:
- `label` — display text
- `url` — page URL (e.g. `/about.html`)
- `page_id` — FK to pages table (for active state highlighting)
- `sort_order` — display order

---

## 2. Nav Authority Tiers (from 002_system_architecture)

| Tier | When | Authority | What happens |
|---|---|---|---|
| Tier 1 — Strategist | New builds, major restructure | Planner plans full nav, nav agent validates | Full nav rebuild |
| Tier 2 — Nav Agent | Maintenance, minor additions | Autonomous incremental decisions | Add/remove individual items |
| Tier 3 — Drift Detection | Periodic | Comparison against original plan | Creates work items for divergence |

Currently only Tier 1 is implemented (via `populate_nav_tables` during build).

---

## 3. Discovery Checks Related to Navigation

### quality-discovery-agent

| Check | Detects | Handler | How |
|---|---|---|---|
| `broken_nav_links` | Nav using anchor links (`#slug`) instead of page URLs | `nav-link-fixer` | Compares nav item URLs against deployed page URLs |
| `placeholder_contact` | Generic contact details from templates | `page-content-writer` | Pattern matching on placeholder text |

### design-discovery-agent

| Check | Detects | Handler |
|---|---|---|
| `checkNavLayout` | Nav `<ul>` with no flex CSS → links stack vertically | `component-template-fixer` |
| `checkUnwantedElements` | Search icon in header when site has no search | `component-template-fixer` |
| `checkUnlinkedSiteComponents` | Header/footer with NULL component_id | `site-component-linker` |

### completeness-discovery-agent

| Check | Detects | Handler |
|---|---|---|
| `orphan_pages` | Deployed pages with no inbound links from nav, header/footer, or other pages | `rerender-pages` (blog) / `content-gap-planner` (content) |

### validate_component_standards

| Check | Detects | Handler |
|---|---|---|
| `checkMissingAssetRefs` | Logo URL set but header has no `<img>` tag | `site-component-linker` |

---

## 4. Fix Agents for Navigation

### nav-link-fixer
Fixes `broken_nav_links` items. Replaces anchor-style links with proper page URLs.

### site-component-linker
Fixes unlinked header/footer/head components:
- Loads style collection for the site
- Reads `header_component_id` / `footer_component_id` from collection
- Updates `site_components.component_id`
- Clears `rendered_html` to force re-render
- Creates `needs_rerender` work item

### component-template-fixer
Routes on `spec.fix_type`:
- `inject_nav_flex_css` — adds `display: flex` to nav `<ul>` in header
- `remove_element` — removes unwanted elements (e.g. search icon)
- `add_data_component` — adds missing `data-component` attribute
- `align_slot_name` — fixes slot_name vs data-component mismatch

### content-gap-planner
For orphan pages: adds a nav entry or internal links from other pages to make the page reachable.

### rerender-pages
For orphan blog posts: rebuilds the blog listing page which auto-generates links to all blog posts.

---

## 5. Tool Pipeline Nav Integration

When a tool is deployed (`create_tool_component` action), it creates:
- A page for the tool
- A `page_component` with the tool HTML
- A nav entry via `addToolToNav`

**Known bug (fixed):** `addToolToNav` used wrong column names for `site_nav_groups` and `site_nav_items`. The nav entry creation failed silently. Fix: corrected column names to match actual schema.

**Remaining issue:** Fuel Cost Estimator on gaswholesalers.com is missing its nav entry (old code, manually patched). Tool health audit will catch it on next 30-day cycle.

---

## 6. Side Effects That Create Nav Work Items

From the contracts (003):
```
Page added/removed → creates nav_update_needed work item
  domain='navigation', item_type='nav_update_needed', source='side_effect'
```

The `write_build_items` action creates `needs_rerender` as the terminal item (priority 99). The rerender step rebuilds header/footer which includes nav.

---

## 7. Snapshot and Revert

Site snapshots (014) capture navigation state:
- `site_nav_groups` → stored in `nav_snapshot` JSONB
- `site_nav_items` → stored in `nav_snapshot` JSONB

`revert_site_to_snapshot` replaces nav groups and items from the snapshot. This is the safest way to restore nav after a bad build.

---

## 8. Admin Dashboard Nav Controls (Future)

From the content governance plan (013):

| Future item | Description |
|---|---|
| Structured nav editor | Edit `site_nav_items` as sortable list in dashboard |
| Site-component lock in discovery | `AND sc.locked_at IS NULL` in nav/style checks |

Currently no dashboard UI for direct nav editing. Manual SQL or triggering a rebuild are the options.

---

## 9. Diagnostic Queries

### What's in the nav tables for a site?

```sql
SELECT g.name as group_name, g.location, i.label, i.url, i.sort_order,
       p.name as page_name, p.build_status
FROM site_nav_groups g
JOIN site_nav_items i ON i.nav_group_id = g.id
LEFT JOIN pages p ON p.id = i.page_id
WHERE g.site_id = '<site_id>'
ORDER BY g.sort_order, i.sort_order;
```

### What pages think they're in nav?

```sql
SELECT name, url, in_header, in_footer, build_status, status, nav_order
FROM pages
WHERE site_id = '<site_id>'
  AND (in_header = true OR in_footer = true)
ORDER BY nav_order;
```

### Are header/footer components linked?

```sql
SELECT sc.slot_name, sc.component_id,
       cc.function as template_function,
       CASE WHEN sc.component_id IS NULL THEN 'FALLBACK' ELSE 'linked' END as status
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = '<site_id>'
  AND sc.slot_name IN ('header', 'footer', 'head');
```

### Does the header HTML have flex nav?

```sql
SELECT
    CASE WHEN rendered_html LIKE '%display: flex%' OR rendered_html LIKE '%display:flex%'
         THEN 'has flex' ELSE 'STACKED NAV' END as nav_layout,
    CASE WHEN rendered_html LIKE '%search-toggle%'
         THEN 'HAS SEARCH ICON' ELSE 'no search' END as search,
    CASE WHEN rendered_html LIKE '%<img%'
         THEN 'has image' ELSE 'NO LOGO IMAGE' END as logo,
    LENGTH(rendered_html) as html_length
FROM site_components
WHERE site_id = '<site_id>' AND slot_name = 'header';
```

### Orphan pages (no nav links)?

```sql
SELECT p.name, p.url, p.build_status
FROM pages p
WHERE p.site_id = '<site_id>'
  AND p.build_status = 'deployed'
  AND NOT EXISTS (
    SELECT 1 FROM site_nav_items ni
    JOIN site_nav_groups ng ON ng.id = ni.nav_group_id
    WHERE ng.site_id = p.site_id AND ni.page_id = p.id
  )
  AND p.in_header = false
  AND p.in_footer = false;
```

### Stale pages from previous builds?

```sql
SELECT name, build_status, status, in_header, in_footer, updated_at
FROM pages
WHERE site_id = '<site_id>'
  AND build_status != 'deployed'
  AND (in_header = true OR in_footer = true);
```

### Work items related to navigation?

```sql
SELECT item_type, status, handler_agent, LEFT(summary, 80), created_at
FROM site_work_items
WHERE site_id = '<site_id>'
  AND (item_type LIKE '%nav%' OR handler_agent LIKE '%nav%'
       OR handler_agent = 'site-component-linker'
       OR handler_agent = 'component-template-fixer')
ORDER BY created_at DESC;
```

---

## 10. Fix Sequence for a Site with Bad Nav

1. **Check component linkage first** — if header isn't linked to style collection, everything else is cosmetic
2. **Fix stale pages** — deactivate pages not in the current plan
3. **Rebuild nav tables** — trigger `populate_nav_tables` or a full rerender
4. **Check for orphans** — pages that exist but aren't reachable
5. **Check header HTML** — flex CSS, search icon, logo image
6. **Re-render** — trigger `needs_rerender` with `refresh_site_components: true`

```sql
-- Quick nav health check for a site
SELECT
    (SELECT COUNT(*) FROM site_nav_items ni JOIN site_nav_groups ng ON ng.id = ni.nav_group_id WHERE ng.site_id = '<site_id>') as nav_items,
    (SELECT COUNT(*) FROM pages WHERE site_id = '<site_id>' AND in_header = true) as pages_in_header,
    (SELECT COUNT(*) FROM pages WHERE site_id = '<site_id>' AND build_status = 'deployed') as deployed_pages,
    (SELECT CASE WHEN component_id IS NOT NULL THEN 'linked' ELSE 'UNLINKED' END FROM site_components WHERE site_id = '<site_id>' AND slot_name = 'header') as header_status;
```
