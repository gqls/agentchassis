# 020 — Content Governance & Inline Editing Plan

How to give admins and users precise control over site content, from individual words to site-wide direction, while keeping automated agents useful but bounded.

---

## The Problem

Agents create content, audit it, and improve it continuously. But there's no concept of "this is final" or "a human approved this specific paragraph." The improvement sweep sees something it thinks could be better and creates a work item to change it. If an admin removes a section, the sweep sees a gap and creates a work item to add it back. There's no boundary between "suggestions welcome" and "leave this alone."

The system also lacks a way to make quick corrections. A typo in a heading currently requires either a manual git commit or waiting for an agent to rewrite the entire section. There's no in-between.

---

## What We Want

1. **See every page and its components** in the dashboard — a structured view, not raw HTML
2. **Edit anything** from a single word to an entire section, with changes deployed in seconds
3. **Lock content** so agents don't overwrite human-approved text
4. **Remove sections** without agents recreating them
5. **See and edit the target spec** that guides the site's direction
6. **Change direction** (tone, services, audience) and have agents gradually propagate changes
7. **Protect data fields** (phone numbers, addresses) from being overwritten
8. **Expose editing to end users** via a client portal (later, same API)

---

## What Already Exists

Infrastructure we can build on:

| Component | Status | Notes |
|-----------|--------|-------|
| `page_components.locked_at` / `locked_by` | Schema exists, indexed | Not checked by agents |
| `trigger_auto_lock_on_deploy` | Trigger exists | Auto-locks components when `deploy_commit` is set |
| `page_component_history` | Table exists | Tracks component versions |
| `page_components.content_data` | jsonb field | Structured content input to templates |
| `page_components.rendered_html` | text field | Template output, what gets deployed |
| `site_specs` with versioning | Table exists | `is_current`, `superseded_at`, aspect-based |
| `site_work_items` dispatch | Working | Agents claim and process items |
| Checkpoint approval flow | Working | Dashboard approves, creates follow-on items |
| Git commit + deploy actions | Working | Agents already commit HTML to repos |

---

## Two Edit Paths

Not all edits should go through agents. The system needs two distinct paths:

### Path 1: Direct Edits (seconds)

For typos, paragraph rewrites, heading changes, section removal. No agents involved.

```
Admin edits component in dashboard
  → API updates content_data (structured) or rendered_html (raw)
  → Component auto-locked (locked_by = 'admin')
  → Page reassembled (header + updated components + footer)
  → Git commit + deploy triggered
  → Live in seconds
```

This bypasses the work item queue entirely. The edit is applied immediately, the component is locked to prevent agents from overwriting it, and the page is redeployed.

### Path 2: Direction Changes (minutes to hours, via agents)

For tone shifts, design changes, audience pivots, new page structures. These affect multiple pages and need LLM-driven content generation.

```
Admin edits site spec in dashboard
  → Spec version updated in site_specs
  → System compares old spec vs new, identifies affected scope
  → Creates targeted work items (content_rewrite, style_fix, etc.)
  → Agents process items through normal dispatch
  → Site gradually evolves toward new direction
```

This uses the existing dispatch pipeline. The difference is the spec change is intentional and targeted, not a discovery-driven suggestion.

---

## Layer 1: Lock Enforcement

### Schema (already exists)

`page_components` already has `locked_at TIMESTAMP`, `locked_by TEXT`, with index `idx_page_components_locked`.

### What needs changing

Add `AND pc.locked_at IS NULL` (or equivalent Go-level check) to every query or action that creates work items for components or modifies components. This affects:

| Check / Action | File | Change |
|----------------|------|--------|
| `findEmptySections` | `check_empty_sections.go` | Add `AND pc.locked_at IS NULL` to WHERE clause |
| `write_audit_findings` | `write_audit_findings_action.go` | Skip findings where component is locked |
| `fix_component_template` | `fix_component_template_action.go` | Check lock before modifying, return `skipped` if locked |
| `page-build-handler` content writes | Various | Check lock before overwriting `rendered_html` |
| `content-quality-audit` | Audit checks | Exclude locked components from audit scope |

The pattern is the same everywhere: before creating a work item about a component or modifying a component, check `locked_at IS NOT NULL`. If locked, skip silently.

### Lock semantics

| `locked_by` value | Meaning | Who can unlock |
|-------------------|---------|----------------|
| `admin` | Human edited via dashboard | Human only |
| `admin-removed` | Section intentionally removed | Human only |
| `deploy` | Auto-locked by deploy trigger | Agents can clear on next improvement cycle |
| `checkpoint` | Locked by checkpoint approval | Human only |
| `NULL` (unlocked) | Agents can modify freely | N/A |

`deploy` locks are soft — they prevent concurrent modification during deploy but agents can clear them when claiming work. `admin` and `admin-removed` locks are hard — only a human can unlock via the dashboard.

---

## Layer 2: Page Structure API

New endpoints that expose pages and components for the dashboard:

### GET /admin/sites/:id/pages

Returns all pages for a site with component counts and lock status:

```json
{
  "pages": [
    {
      "id": "uuid",
      "name": "index",
      "title": "Home | Gas Wholesalers",
      "url": "/index.html",
      "page_type": "home",
      "status": "active",
      "build_status": "deployed",
      "component_count": 6,
      "locked_count": 2,
      "last_built_at": "2026-03-17T..."
    }
  ]
}
```

### GET /admin/sites/:id/pages/:page_name/components

Returns components in position order with content preview and lock status:

```json
{
  "page": { "id": "uuid", "name": "index", "title": "Home" },
  "components": [
    {
      "id": "uuid",
      "position": 1,
      "slot_name": "hero",
      "content_data": { "heading": "Wholesale Gas...", "subtitle": "..." },
      "html_preview": "<section class=\"hero\">...",
      "locked": true,
      "locked_by": "admin",
      "locked_at": "2026-03-17T...",
      "build_status": "deployed"
    }
  ]
}
```

The `html_preview` is the first 500 chars of `rendered_html` — enough for the dashboard to show what each section looks like without loading full page HTML.

### PATCH /admin/sites/:id/pages/:page_name/components/:component_id

Update a component's content and optionally lock it:

```json
{
  "content_data": { "heading": "Fixed Typo Here", "subtitle": "..." },
  "lock": true,
  "rebuild_page": true
}
```

The endpoint:
1. Saves current state to `page_component_history`
2. Updates `content_data` (and optionally `rendered_html` if raw edit)
3. If `lock: true`, sets `locked_at = NOW(), locked_by = 'admin'`
4. If `rebuild_page: true`, creates a `page_rerender` work item for immediate reassembly and deploy

### POST /admin/sites/:id/pages/:page_name/components/:component_id/unlock

Removes the lock, allowing agents to modify the component again.

### DELETE /admin/sites/:id/pages/:page_name/components/:component_id

Removes a section from a page:
1. Saves to `page_component_history` (preserves undo capability)
2. Sets `build_status = 'removed'` (soft delete, not hard delete)
3. Records suppression on the page (prevents recreation — see Layer 4)
4. Creates `page_rerender` work item to rebuild the page without this section

---

## Layer 3: Spec Visibility and Direction Control

### GET /admin/sites/:id/specs

Returns all current specs grouped by aspect. Already partially exists.

The dashboard enhancement adds a "Direction" tab showing all specs, with each aspect expandable and inline-editable.

### PATCH /admin/sites/:id/specs/:aspect

Already exists. Creates a new versioned spec entry.

### POST /admin/sites/:id/specs/:aspect/propagate

New endpoint. Takes a spec change and creates targeted work items:

```json
{
  "scope": "all_pages",
  "item_type": "content_rewrite",
  "summary_template": "Update {{page_name}} to reflect new {{aspect}} direction",
  "priority": 30
}
```

The endpoint:
1. Identifies affected pages based on aspect type (identity/tone → all content pages; design → all pages; page_plan → specific pages)
2. Creates work items for each affected page, excluding pages with all components locked
3. Returns the list of created items so the admin can review

### Spec Pinning

Add a `pinned` boolean to `site_specs`. When a spec is pinned, the classifier and improvement agents skip it. The admin can pin individual aspects — e.g. pin `identity` so the tone and audience aren't changed by sweeps, but leave `design_direction` unpinned for the webdesign agent to improve.

Agents check `WHERE pinned = false` (or `pinned IS NOT TRUE`) before creating a new version of a spec aspect.

---

## Layer 4: Preventing Recreation of Removed Content

When an admin removes a section, the empty_sections check would normally flag the gap and create a work item to fill it.

### Approach: page-level suppression list

Add a `suppressed_sections` jsonb column to `pages`:

```sql
ALTER TABLE pages ADD COLUMN suppressed_sections JSONB DEFAULT '[]'::jsonb;
```

When a section is removed via the dashboard, its `slot_name` is added:

```json
["call-to-action", "testimonials"]
```

The `findEmptySections` query adds:

```sql
AND pc.slot_name NOT IN (
    SELECT jsonb_array_elements_text(COALESCE(p.suppressed_sections, '[]'))
)
```

The content-gap-planner also checks `suppressed_sections` before suggesting new sections for a page.

### Restoring suppressed sections

The dashboard shows suppressed sections as greyed-out items at the bottom of the page structure view. Click to restore — removes from `suppressed_sections` and optionally creates a work item to populate the section.

---

## Layer 5: Dashboard UI

### Page Structure Browser

A new view accessible from the site card or a "Pages" tab:

```
┌─────────────────────────────────────────────────────────────┐
│  Pages — gaswholesalers.com                      [Edit Site] │
├─────────────────┬───────────────────────────────────────────┤
│                 │                                            │
│  index      (6) │  ┌─ hero ────────────────────── 🔒 ────┐  │
│  about      (4) │  │  Wholesale Gas Distribution...       │  │
│  services   (5) │  │  [Edit] [Unlock]                     │  │
│  contact    (3) │  └──────────────────────────────────────┘  │
│  pricing    (4) │  ┌─ features ──────────────────────────┐  │
│  blog       (2) │  │  3 feature cards: Competitive...     │  │
│                 │  │  [Edit] [Lock] [Remove]               │  │
│                 │  └──────────────────────────────────────┘  │
│                 │  ┌─ services-grid ─────────────────────┐  │
│                 │  │  (empty)                              │  │
│                 │  │  [Edit] [Suppress]                    │  │
│                 │  └──────────────────────────────────────┘  │
│                 │                                            │
│                 │  ── Suppressed ──                          │
│                 │  call-to-action [Restore]                  │
│                 │                                            │
└─────────────────┴───────────────────────────────────────────┘
```

Left side: page list with component counts. Right side: components in order, each showing slot name, content preview (from rendered_html or content_data), lock icon, and action buttons.

### Component Edit Panel

Clicking "Edit" on a component opens the detail panel:

- For structured content (`content_data` has fields): renders the `EditableReviewForm` component we already have (field-by-field editing with type detection)
- For raw HTML only: shows a textarea with `rendered_html`
- Lock toggle
- "Save & Deploy" button — applies change, locks component, triggers page reassembly and deploy
- "Save Draft" button — saves to DB without deploying (changes apply on next scheduled rebuild)

### Spec Editor

A "Direction" tab on the site view showing all current specs. Each aspect is expandable with inline editing. Pinned aspects show a lock icon. Editing a spec shows a diff summary and a "Propagate Changes" button that previews what work items will be created before committing.

### Data Fields

The existing "Edit Site" panel (company_name, email, phone, address, tagline) handles sites-table fields. For richer data (services list, team members, pricing plans), these live in `site_specs` and the spec editor covers them. The dashboard can surface the most-edited spec fields (identity, services) as a dedicated "Business Data" section alongside the site edit panel.

---

## How Agents Behave After This

| Agent Action | Locked Component | Unlocked Component |
|-------------|-----------------|-------------------|
| Improvement sweep finds issue | Skips, no work item created | Creates work item as normal |
| Content rewriter modifies | Skips, returns `skipped` | Modifies as normal |
| Template fixer updates CSS | Skips | Fixes as normal |
| Empty section check | Skips if locked or suppressed | Creates `empty_section` item |
| Content-gap-planner | Skips suppressed sections | Suggests additions as normal |
| Spec classifier | Skips pinned spec aspects | Updates as normal |

---

## Data Model Changes Summary

```sql
-- Phase 1: lock enforcement — query changes only, no schema needed

-- Phase 4: spec pinning
ALTER TABLE site_specs ADD COLUMN pinned BOOLEAN DEFAULT false;

-- Phase 5: section suppression
ALTER TABLE pages ADD COLUMN suppressed_sections JSONB DEFAULT '[]'::jsonb;
```

Everything else is API endpoints and dashboard UI. `locked_at`, `locked_by`, `page_component_history`, `content_data`, and `rendered_html` all already exist.

---

## Build Priority

| Phase | Scope | What it enables |
|-------|-------|----------------|
| 1. Lock enforcement | Small: query changes in 5 files | Locked components left alone by all agents |
| 2. Page structure API | Medium: 4 new endpoints | Dashboard can show and browse page components |
| 3. Direct edit + deploy | Medium: edit panel + reassembly | Fix typos and content in seconds, auto-lock |
| 4. Spec editor + propagation | Medium: spec UI + propagate endpoint | Change site direction, agents follow |
| 5. Section suppression | Small: 1 column + query filter | Remove sections without recreation |
| 6. User portal | Larger: auth + permissions | Clients edit their own sites |

Each phase is independently useful. Phase 1 can be deployed in an hour. Phases 2-3 together give the core editing experience. Phase 4 adds strategic control. Phase 5 is a refinement. Phase 6 is a separate project that reuses everything from phases 1-5.

---

## Key Design Decisions

**Direct edits bypass agents entirely.** A typo fix shouldn't wait for an LLM. The dashboard updates the component, reassembles the page, and deploys. This requires the reassembly and git commit actions to be callable from the admin API, not just from agent workflows.

**Locks are additive, not destructive.** Locking a component doesn't remove it from the page or hide it from view. It just means "don't change this." Agents can still read locked components for context (e.g. understanding the page's tone) — they just can't modify them or create work items about them.

**Spec changes are explicit, not automatic.** Editing a spec doesn't automatically rewrite every page. The admin clicks "Propagate" and sees what work items will be created before confirming. This prevents accidental cascading changes.

**Suppression is page-scoped.** Removing a CTA from the about page doesn't suppress CTAs everywhere — just on that page. Other pages can still get CTA suggestions.

**History is preserved.** Every direct edit saves to `page_component_history`. Removed sections are soft-deleted (`build_status = 'removed'`), not hard-deleted. Spec changes are versioned. Everything can be undone.

**Same API, two audiences.** The admin dashboard and the future user portal use the same endpoints. The difference is auth scope: admins see all sites, users see only their own. No separate API needed.
