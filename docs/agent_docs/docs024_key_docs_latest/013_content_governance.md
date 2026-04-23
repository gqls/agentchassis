# 013 — Content Governance & Inline Editing

How admins and users get precise control over site content, from individual words to site-wide direction, while keeping automated agents useful but bounded.

---

## Status Summary

| Phase | Status | What it does |
|-------|--------|-------------|
| 1. Lock enforcement | ✅ Deployed | Discovery checks skip locked page_components |
| 2. Page structure API | ✅ Deployed | Endpoints to list pages, list/edit/lock/unlock/remove components |
| 3. Direct edit + deploy | ✅ Deployed | Dashboard page browser with inline editing, auto-lock, rerender |
| 4. Spec editor + propagation | ✅ Deployed | View/edit specs, pin/unpin, propagate changes to work items |
| 5. Section suppression | ✅ Deployed | Remove sections without recreation, restore with work item |
| 6. User portal | Future | Client auth, site-scoped permissions, same API |
| 7. Site-wide components | ✅ Deployed | Edit header/footer/CSS, lock nav and styles |
| 8. Site lock | ✅ Deployed | Freeze entire site — all automated activity stops |
| 9. Content briefs + regeneration | ✅ Deployed | View/edit content instructions, regenerate sections or pages via LLM |
| 10. Media browser | ✅ Deployed | Browse assets, view references, check deploy status |
| 11. Page growth budget | ✅ Deployed | Rate-limit page creation to prevent site bloat |

---

## Three Edit Paths

| Need | Tab | Button | What happens | Speed |
|------|-----|--------|-------------|-------|
| Fix a typo | HTML | Save & Deploy | Direct edit, auto-lock, rerender | Seconds |
| Change field values | Fields | Save & Deploy | Direct edit, auto-lock, rerender | Seconds |
| Change section direction | Brief | Regenerate | Brief saved, content_rewrite queued, LLM rewrites | Minutes |
| Change whole page | Page Purpose | Regenerate Page | All unlocked sections rewritten with new purpose | Minutes |
| Change site direction | Direction | Propagate | Work items created per page across site | Minutes-hours |

---

## Three Levels of Lock

### Component Lock (page_components)

Locks a single section on one page. Discovery agents skip it. Editing a locked component does not require unlocking — the lock stays on.

### Site Component Lock (site_components)

Locks the header, footer, or CSS across all pages. Same semantics, site-wide scope.

### Site Lock (sites)

Freezes the entire site. All automated agent activity stops. Admin can still manually create and process items.

- `LoadWorkItemsAction` returns empty if site locked
- `build-pipeline-trigger` pre_query skips locked sites

### Lock Semantics

The lock means "human controls this," not "read-only." Edit and unlock are separate operations:

- **Edit** = change content, lock stays on, `locked_at` refreshes
- **Unlock** = hand back to agents, a deliberate separate action

**Discovery agents** check locks before creating work items (hard gate).
**Execution agents** process explicit work items regardless (soft gate catches stale pre-lock items).
**Rerender** is read-only assembly — reads all components including locked ones, stitches together.

| `locked_by` | Meaning | Who can unlock |
|-------------|---------|----------------|
| `admin` | Human edited | Human only |
| `admin-removed` | Section removed | Human only |
| `checkpoint` | Checkpoint approved | Human only |
| `deploy` | Auto-locked on deploy | Agents can clear |

---

## Content Briefs & Regeneration (Phase 9)

Every page_component can have a `content_brief` jsonb field recording the instructions that generated its content:

```json
{
  "purpose": "Explain AI automation benefits for UK small businesses",
  "tone_direction": "Professional but approachable",
  "section_guidance": "Hero section with clear headline and value proposition"
}
```

### Component Regeneration

The "Brief" tab in the component edit panel shows these instructions. The admin can edit the purpose, tone, or guidance, then click "Regenerate." This saves the updated brief and creates a `content_rewrite` work item with the brief in the spec. The page-content-writer uses it as instructions when rewriting.

### Page Regeneration

The page purpose (from `pages.page_spec`) is shown in an amber bar at the top of the component list. Editable inline. "Regenerate Page" creates `content_rewrite` items for all unlocked sections — locked sections are skipped.

### How the Content Writer Uses Briefs

The `content_rewrite` work item spec includes `content_brief` from the component. The page-content-writer constructs its LLM prompt from the brief alongside `site_specs.identity` (tone, audience) and the section type. This means admin edits to the brief directly control what the LLM generates.

---

## Media Browser (Phase 10)

`asset_admin_handlers.go` — four endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/assets` | List with deploy status, reference count, dimensions |
| GET | `/sites/:id/assets/:id/references` | Which components reference this asset |
| PATCH | `/sites/:id/assets/:id` | Update purpose, name, status |
| DELETE | `/sites/:id/assets/:id` | Soft-delete |

Dashboard "Media" tab shows assets grouped by deployment status. Each card shows purpose, type, thumbnail (for images), URL, size, origin type, and how many components reference it. Detail panel shows full preview, metadata, reference list with page and slot names.

---

## Page Growth Budget (Phase 11)

`page_growth_budget.go` — shared `CheckPageGrowthBudget()` function called by `apply_gap_plan` (new content pages) and `create_blog_posts` (blog posts).

### Configuration

Stored as a `growth_config` spec aspect (editable via Direction tab, pinnable):

```json
{
  "initial_target": 12,
  "weekly_content_pages_max": 3,
  "weekly_blog_posts_max": 2,
  "absolute_max": 60
}
```

### Behaviour

- Under `initial_target` → allows freely (first build creates its pages)
- Past initial target → checks 7-day rolling window by page type
- At `absolute_max` → blocks entirely
- Blog posts and content pages have separate weekly budgets
- Blocked items get status `blocked` (not failed) with descriptive message — retryable next week
- Blog posts that exceed budget are skipped with `continue` — remaining posts in the plan still get created up to the limit

### Enforcement Points

| Action | Where | What happens when over budget |
|--------|-------|-------------------------------|
| `apply_gap_plan` → `new_page` | Before page INSERT | Original work item marked `blocked` |
| `create_blog_posts` loop | Before each page INSERT | Post skipped, loop continues |

---

## Completed Phase Details

### Phase 1: Lock Enforcement ✅

`AND pc.locked_at IS NULL` added to four discovery checks: `findEmptySections`, `findPlaceholderContact`, `countHardcodedColorComponents`, `countForcedTextColorIssues`. `findEmptySections` also filters on `suppressed_sections`. Structural/bug checks exempt.

### Phase 2: Page Structure API ✅

Seven endpoints in `page_admin_handlers.go`: list pages, list components (with full HTML, content_brief, page_spec), edit component, lock, unlock, remove (soft-delete + suppress), restore section.

### Phase 3: Dashboard Page Browser ✅

Left panel: Site-Wide entry + page list. Right panel: component cards. Edit panel: Fields/HTML/Brief tabs. Page purpose bar with inline edit. Regenerate Page button. Suppressed sections with restore.

### Phase 4: Spec Editor + Propagation ✅

Four endpoints in `spec_admin_handlers.go`. Dashboard Direction tab with pin/unpin/edit/propagate. Propagation skips fully-locked pages.

### Phase 5: Section Suppression ✅

`suppressed_sections` jsonb on pages. Discovery checks filter on it. Remove → suppress → restore cycle fully wired.

### Phase 7: Site-Wide Component Editing ✅

Four endpoints for site_components (header/footer/head). Edit triggers full-site `needs_rerender`. Lock columns added to `site_components`.

### Phase 8: Site Lock ✅

Lock/unlock on sites table. `LoadWorkItemsAction` gate + pre_query filter. Dashboard toggle with confirmation.

---

## Work Item Retry Dedup Fix

The `idx_swi_dedup` unique index excludes `failed` items. When retrying back to `triaged`, conflicts can occur with newer duplicates. `HandleRetryWorkItem` now auto-completes conflicting items before the retry UPDATE.

---

## Schema Changes

```sql
-- Phase 2/5
ALTER TABLE pages ADD COLUMN IF NOT EXISTS suppressed_sections JSONB DEFAULT '[]'::jsonb;

-- Phase 4
ALTER TABLE site_specs ADD COLUMN IF NOT EXISTS pinned BOOLEAN DEFAULT false;

-- Phase 7
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_by TEXT;
CREATE INDEX IF NOT EXISTS idx_site_components_locked ON site_components (locked_at) WHERE locked_at IS NOT NULL;

-- Phase 8
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_by TEXT;

-- Phase 9
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS content_brief JSONB;
```

---

## How Agents Behave

### Discovery (hard gate)

| Target | Locked | Unlocked | Site locked |
|--------|--------|----------|-------------|
| page_components content/style | Skips | Creates item | No checks run |
| Suppressed section | Skips | Skips | No checks run |
| site_components nav/style | Skips | Creates item | No checks run |
| Structural bugs | Checks regardless | Checks | No checks run |
| Pinned specs | Skips | Updates | No checks run |

### Execution (soft gate)

| Agent | Locked (admin) | Locked (deploy) | Unlocked |
|-------|---------------|-----------------|----------|
| Rerender (reads) | Reads normally | Reads normally | Reads normally |
| Content rewriter (admin) | Writes, lock stays | Writes | Writes |
| Template fixer (stale) | Skips | Proceeds | Proceeds |

### Dispatch

| Site locked? | Behaviour |
|-------------|-----------|
| No | Normal dispatch |
| Yes | LoadWorkItemsAction returns empty, pre_query skips |

### Growth Budget

| Budget state | New content page | Blog post |
|-------------|-----------------|-----------|
| Under initial target | Allowed | Allowed |
| Weekly content limit hit | Blocked | Allowed (separate budget) |
| Weekly blog limit hit | Allowed | Skipped |
| Absolute max | Blocked | Skipped |

---

## Go Files

| File | Location | Purpose |
|------|----------|---------|
| `site_admin_handlers.go` | `internal/core-manager/admin/` | Sites, work items, site lock, retry, approve |
| `page_admin_handlers.go` | `internal/core-manager/admin/` | Pages, components, site-components, regenerate, page spec |
| `spec_admin_handlers.go` | `internal/core-manager/admin/` | Spec list, pin/unpin, propagate |
| `asset_admin_handlers.go` | `internal/core-manager/admin/` | Asset list, references, update, delete |
| `lock_helpers.go` | `platform/orchestration/actions/` | Shared lock check functions |
| `page_growth_budget.go` | `platform/orchestration/actions/` | Growth budget check |
| `check_empty_sections.go` | `platform/orchestration/actions/discovery_checks/` | Lock + suppression filtering |
| `check_placeholder_contact.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `check_hardcoded_section_colors.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `check_forced_text_colors.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `load_work_items_action.go` | `platform/orchestration/actions/` | Site lock gate |
| `server.go` | `internal/core-manager/api/` | Route registration |

---

## Key Design Decisions

**Direct edits bypass agents.** Typo fix → seconds to live. No LLM needed.

**Locks are additive.** "Don't change unsolicited" not "read-only." Edit doesn't require unlock.

**Site-wide edits are reassembly.** Header/footer/CSS changes rebuild all pages but don't modify any section content.

**Site lock stops everything.** Discovery, dispatch, sweeps — all stop. Admin retains manual control.

**Briefs control regeneration.** The content_brief on each component is the instruction set for the LLM. Edit the brief, regenerate, get content that matches the new direction.

**Spec propagation is explicit.** Admin clicks Propagate, sees what will be created, confirms.

**Suppression is page-scoped.** Removing CTA from one page doesn't affect others.

**Growth budgets prevent bloat.** Initial pages created freely, then rate-limited by weekly rolling window with separate content and blog limits.

**Dedup-safe retries.** Auto-resolves conflicting duplicates.

**Same API, two audiences.** Admin and future user portal share endpoints. Auth scope differs.

---

## Future Work

| Item | Description |
|------|-------------|
| User portal (Phase 6) | Client auth, site-scoped permissions, simplified browser |
| Structured nav editor | Edit `site_nav_items` as sortable list |
| CSS variable editor | Parse `head` CSS, present variables as form fields |
| Image upload | Upload/replace assets via dashboard |
| Pin enforcement in agents | Classifier checks `WHERE pinned = false` |
| Site-component lock in discovery | `AND sc.locked_at IS NULL` in nav/style checks |
| Content writer brief reading | page-content-writer reads `spec.content_brief` for prompt construction |
