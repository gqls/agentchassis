# 020 — Content Governance & Inline Editing

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

---

## Two Edit Paths

### Path 1: Direct Edits (seconds)

For typos, paragraph rewrites, heading changes, section removal. No agents involved.

```
Admin edits component in dashboard
  → API updates content_data or rendered_html
  → Component auto-locked (locked_by = 'admin')
  → History saved to page_component_history
  → page_rerender work item created
  → Rerender agent rebuilds page → git commit → deploy
  → Live in seconds
```

### Path 2: Direction Changes (minutes to hours, via agents)

For tone shifts, design changes, audience pivots, new page structures.

```
Admin edits site spec in dashboard
  → Spec version updated in site_specs
  → Admin clicks "Propagate"
  → Creates targeted work items (content_rewrite, etc.)
  → Agents process through normal dispatch
  → Site gradually evolves toward new direction
```

---

## Three Levels of Lock

### Component Lock (page_components)

Locks a single section on one page. Discovery agents skip it, but the rerender agent still reads it when assembling the full page. Editing a locked component does not require unlocking — the lock stays on.

- `locked_at` / `locked_by` columns (existed before this work)
- `idx_page_components_locked` index (existed)
- Four discovery checks filter on `AND pc.locked_at IS NULL`

### Site Component Lock (site_components)

Locks the header, footer, or CSS across all pages. Same semantics as component locks but site-wide scope.

- `locked_at` / `locked_by` columns (added in Phase 7)
- `idx_site_components_locked` index (added in Phase 7)

### Site Lock (sites)

Freezes the entire site. All automated agent activity stops — no discovery checks run, no work items are dispatched. Admin can still manually create and process items.

- `locked_at` / `locked_by` columns on `sites` table (added in Phase 7)
- `LoadWorkItemsAction` returns empty if `sites.locked_at IS NOT NULL`
- `build-pipeline-trigger` pre_query skips locked sites with `WHERE s.locked_at IS NULL`

---

## How Locks Interact with Editing and Execution

The lock means "human controls this," not "read-only." There is no need to unlock a component to edit it. The two operations are separate:

- **Edit** = change the content. The lock stays on. `locked_at` refreshes, `locked_by` stays `admin`.
- **Unlock** = hand control back to agents. A deliberate separate action meaning "I'm done, agents can improve this again."

**Discovery agents** check the lock before *creating* work items. The `AND pc.locked_at IS NULL` clause means: don't suggest changes for locked components. No work items are created.

**Execution agents** don't refuse locked components when processing an explicit work item. The work item is the authorisation. But agents that write directly to a component do a soft check to catch stale pre-lock items: skip if `locked_by IN ('admin', 'admin-removed', 'checkpoint')`, proceed if `deploy` or NULL.

**No race condition on edit:**

1. Component is locked (`locked_by = 'admin'`)
2. Audit sweep runs → sees lock → creates no work items
3. Admin edits content → API updates `content_data`, refreshes `locked_at`
4. API creates `page_rerender` work item
5. Rerender agent reads all components (locked one included), assembles page, deploys
6. Component stays locked — no window for agents

The rerender agent doesn't modify components — it reads `rendered_html` from all components and stitches them together with header and footer. It's a read-only assembly operation.

### Lock values

| `locked_by` | Meaning | Who can unlock |
|-------------|---------|----------------|
| `admin` | Human edited via dashboard | Human only |
| `admin-removed` | Section intentionally removed | Human only |
| `checkpoint` | Locked by checkpoint approval | Human only |
| `deploy` | Auto-locked by deploy trigger | Agents can clear |
| `NULL` | Agents can modify freely | N/A |

---

## Completed Phases

### Phase 1: Lock Enforcement ✅

Added `AND pc.locked_at IS NULL` to four discovery check queries:

| Check | File |
|-------|------|
| `findEmptySections` | `check_empty_sections.go` |
| `findPlaceholderContact` | `check_placeholder_contact.go` |
| `countHardcodedColorComponents` | `check_hardcoded_section_colors.go` |
| `countForcedTextColorIssues` | `check_forced_text_colors.go` |

`findEmptySections` also filters on `suppressed_sections` (Phase 5).

Structural/bug checks (contamination, unrendered templates, slot mismatches) do NOT filter on locks — those are bugs regardless.

Shared helper `lock_helpers.go` provides `CheckComponentLock()` and `CheckPageHasHardLocks()` for execution-agent soft checks.

### Phase 2: Page Structure API ✅

`page_admin_handlers.go` — 7 endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/pages` | Page list with component/locked/empty counts |
| GET | `/sites/:id/pages/:name/components` | Components with content_data, full HTML, lock status, suppressed_sections |
| PATCH | `/sites/:id/pages/:name/components/:id` | Edit content, auto-lock, save history, trigger rerender |
| POST | `.../components/:id/lock` | Lock a component |
| POST | `.../components/:id/unlock` | Unlock a component |
| DELETE | `.../components/:id` | Soft-remove, suppress, trigger rerender |
| POST | `/sites/:id/pages/:name/restore-section` | Restore suppressed section |

Components return full `rendered_html` (not truncated) — the dashboard handles display truncation client-side via `stripHtmlTags().slice(0, 200)`.

### Phase 3: Dashboard Page Browser ✅

Page browser accessible via "Pages" button on site cards:

- Left panel: page list with section count, locked count, empty count
- "Site-Wide" entry at top for header/footer/CSS (Phase 7)
- Right panel: component cards showing slot name, text preview (HTML stripped), lock/empty badges
- Edit panel: two modes — "Fields" (structured via `EditableReviewForm`) and "HTML" (textarea with full content)
- "Save & Deploy" auto-locks and queues rerender
- Lock/Unlock/Remove buttons on each component
- Suppressed sections shown at bottom with dashed border and "Restore" button

### Phase 4: Spec Editor + Propagation ✅

`spec_admin_handlers.go` — 4 endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/specs` | All current specs with pinned status |
| POST | `/sites/:id/specs/:aspect/pin` | Pin — agents won't override |
| POST | `/sites/:id/specs/:aspect/unpin` | Unpin — agents can update |
| POST | `/sites/:id/specs/:aspect/propagate` | Create work items for affected pages |

Dashboard "Direction" tab shows all specs as cards with compact data preview (top 6 fields), pin/unpin/edit/propagate actions. Propagation skips fully-locked pages. Spec pinning adds `pinned BOOLEAN DEFAULT false` to `site_specs`.

### Phase 5: Section Suppression ✅

- `suppressed_sections` jsonb column on `pages`
- `findEmptySections` filters with `AND NOT (COALESCE(p.suppressed_sections, '[]'::jsonb) ? COALESCE(pc.slot_name, ''))`
- `HandleRemoveComponent` writes to suppressed_sections
- `HandleRestoreSection` removes from it, restores the component (`build_status = 'pending'`, clears lock), creates populate work item
- Dashboard shows suppressed sections with strikethrough text and green "Restore" button

### Phase 7: Site-Wide Component Editing ✅

`page_admin_handlers.go` — 4 additional endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/site-components` | List header/footer/head with full HTML, lock status, size |
| PATCH | `/sites/:id/site-components/:slot` | Edit HTML, auto-lock, create `needs_rerender` for all pages |
| POST | `.../site-components/:slot/lock` | Lock |
| POST | `.../site-components/:slot/unlock` | Unlock |

The PATCH endpoint creates a `needs_rerender` work item with `refresh_site_components: true` and all deployed page names. The rerender-pages agent reassembles every page with the updated shared component.

Dashboard: "Site-Wide" entry at top of page list shows header, footer, and CSS/Styles cards. Each shows size in kb, lock status, text preview. Edit opens HTML textarea. Save triggers full-site rebuild. Note at bottom explains the scope.

**Key point: site-wide edits are safe.** Editing the header/footer/CSS triggers a *reassembly*, not a *rewrite*. The rerender agent reads existing `page_components.rendered_html` and the updated `site_components.rendered_html`, stitches them together, and deploys. No page section content is modified.

### Phase 8: Site Lock ✅

`site_admin_handlers.go` — 2 endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/sites/:id/lock` | Lock site — all automated activity stops |
| POST | `/sites/:id/unlock` | Unlock site — agents resume |

Enforcement at two levels:

1. **`LoadWorkItemsAction`** — checks `sites.locked_at IS NOT NULL` before loading work items. Returns empty list with `skipped_reason: "site_locked"` if locked.
2. **`build-pipeline-trigger` pre_query** — `WHERE s.locked_at IS NULL AND EXISTS (...)` skips locked sites entirely.

Dashboard: site cards show purple "🔒 Locked" badge with purple border. "Lock Site" / "🔒 Unlock Site" toggle button (lock requires confirmation). Locked sites have reduced opacity. `HandleListSites` returns `locked`, `locked_by`, `locked_at` per site.

---

## Work Item Retry Dedup Fix

`HandleRetryWorkItem` handles a dedup index conflict. The `idx_swi_dedup` unique index excludes `failed` items. When retrying a `failed` item back to `triaged`, it re-enters the index scope and may conflict with a newer item (same `item_key`) created by a discovery sweep while the original was failed.

The fix: before the retry UPDATE, a preliminary UPDATE completes any other active items with the same `site_id + item_key`, marking them "Superseded by admin retry." This clears the dedup slot so the original can safely move back to `triaged`.

---

## Schema Changes Applied

```sql
-- Phase 1: query changes only (locked_at/locked_by already existed on page_components)

-- Phase 2/5: section suppression
ALTER TABLE pages ADD COLUMN IF NOT EXISTS suppressed_sections JSONB DEFAULT '[]'::jsonb;

-- Phase 4: spec pinning
ALTER TABLE site_specs ADD COLUMN IF NOT EXISTS pinned BOOLEAN DEFAULT false;

-- Phase 7: site_components lock columns
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE site_components ADD COLUMN IF NOT EXISTS locked_by TEXT;
CREATE INDEX IF NOT EXISTS idx_site_components_locked
    ON site_components (locked_at) WHERE locked_at IS NOT NULL;

-- Phase 8: site-level lock
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sites ADD COLUMN IF NOT EXISTS locked_by TEXT;
```

---

## API Endpoints Summary

### Sites

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites` | List sites with work item counts and lock status |
| GET | `/sites/:id` | Site detail with specs |
| PATCH | `/sites/:id` | Update site fields |
| POST | `/sites/:id/lock` | Lock site |
| POST | `/sites/:id/unlock` | Unlock site |

### Page Structure (Phases 2-5)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/pages` | Page list with counts |
| GET | `/sites/:id/pages/:name/components` | Component list with content, locks, suppressed |
| PATCH | `/sites/:id/pages/:name/components/:id` | Edit, auto-lock, rerender |
| POST | `.../components/:id/lock` | Lock |
| POST | `.../components/:id/unlock` | Unlock |
| DELETE | `.../components/:id` | Remove + suppress |
| POST | `/sites/:id/pages/:name/restore-section` | Restore suppressed |

### Site-Wide Components (Phase 7)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/site-components` | List header/footer/head |
| PATCH | `/sites/:id/site-components/:slot` | Edit, auto-lock, full-site rerender |
| POST | `.../site-components/:slot/lock` | Lock |
| POST | `.../site-components/:slot/unlock` | Unlock |

### Spec Direction (Phase 4)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/specs` | All current specs |
| PATCH | `/sites/:id/specs/:aspect` | Update spec (versioned) |
| POST | `/sites/:id/specs/:aspect/pin` | Pin |
| POST | `/sites/:id/specs/:aspect/unpin` | Unpin |
| POST | `/sites/:id/specs/:aspect/propagate` | Create propagation items |

### Work Items

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/work-items` | Create |
| GET | `/work-items` | List (filterable) |
| GET | `/work-items/:id` | Detail |
| PATCH | `/work-items/:id` | Update |
| POST | `/work-items/:id/retry` | Reset to triaged (with dedup handling) |
| POST | `/work-items/:id/resolve` | Mark complete |
| POST | `/work-items/:id/approve` | Approve checkpoint |

---

## How Agents Behave (Complete Picture)

### Discovery agents (hard gate — no items created if locked)

| Check target | Locked | Unlocked | Site locked |
|-------------|--------|----------|-------------|
| `page_components` content/style | Skips | Creates work item | No checks run |
| `page_components` on suppressed section | Skips | Skips | No checks run |
| `site_components` nav/style | Skips | Creates work item | No checks run |
| `site_components` structural bugs | Checks regardless | Checks regardless | No checks run |
| Pinned spec aspects | Skips | Updates | No checks run |

### Execution agents (soft gate — catches stale items)

| Agent | Locked (admin) | Locked (deploy) | Unlocked |
|-------|---------------|-----------------|----------|
| Page rerender (reads components) | Reads normally | Reads normally | Reads normally |
| Content rewriter (admin-created) | Writes, lock stays | Writes, lock stays | Writes |
| Template fixer (stale pre-lock item) | Skips | Proceeds | Proceeds |
| Page-build-handler (dispatched) | Soft check: skip if admin | Proceeds | Proceeds |

### Dispatch loop

| Condition | Behaviour |
|-----------|-----------|
| Site unlocked | Normal dispatch — picks up triaged items |
| Site locked | `LoadWorkItemsAction` returns empty, pre_query skips site entirely |

---

## Dashboard Views Summary

**Sites Overview** — cards with domain, company name, work item counts, lock badge. Buttons: Work Items, Pages, Direction, Lock/Unlock Site.

**Work Items** — filterable list with status counts in dropdown, split-pane detail, review forms (placeholder/checkpoint/standard), retry/resolve/approve actions, bulk "Retry All Failed" button, error previews on failed items, domain badges in all-items view.

**Pages** — left panel page list + "Site-Wide" entry, right panel component cards with preview/lock/edit/remove, edit panel with Fields and HTML modes, suppressed sections with restore.

**Direction** — spec cards with aspect name, source, date, pinned badge, data preview. Edit opens full form. Pin/unpin/propagate per aspect.

---

## Go Files Reference

| File | Location | Purpose |
|------|----------|---------|
| `site_admin_handlers.go` | `internal/core-manager/admin/` | Site CRUD, lock/unlock, work item CRUD, retry (with dedup fix), approve |
| `page_admin_handlers.go` | `internal/core-manager/admin/` | Page structure, component CRUD, site-component CRUD, section restore |
| `spec_admin_handlers.go` | `internal/core-manager/admin/` | Spec list, pin/unpin, propagate |
| `lock_helpers.go` | `platform/orchestration/actions/` | Shared `CheckComponentLock`, `CheckPageHasHardLocks` |
| `check_empty_sections.go` | `platform/orchestration/actions/discovery_checks/` | Lock + suppression filtering |
| `check_placeholder_contact.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `check_hardcoded_section_colors.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `check_forced_text_colors.go` | `platform/orchestration/actions/discovery_checks/` | Lock filtering |
| `load_work_items_action.go` | `platform/orchestration/actions/` | Site lock gate |
| `server.go` | `internal/core-manager/api/` | Route registration |

---

## Key Design Decisions

**Direct edits bypass agents entirely.** A typo fix shouldn't wait for an LLM. The API updates, queues a rerender, and it deploys.

**Locks are additive.** "Don't change this unsolicited" — not "read-only." Editing doesn't require unlocking. Unlocking is a separate deliberate action.

**Site-wide edits trigger reassembly, not rewrite.** Header/footer/CSS changes cause all pages to be reassembled with the updated shared components. No page section content is modified.

**Site lock stops all automation.** Discovery, dispatch, improvement sweeps — everything stops. Admin can still create and process items manually for targeted fixes.

**Spec changes are explicit.** Editing a spec doesn't auto-rewrite pages. The admin clicks "Propagate" and sees what items will be created.

**Suppression is page-scoped.** Removing a CTA from about doesn't suppress CTAs everywhere.

**History preserved.** Component edits saved to `page_component_history`. Sections soft-deleted. Specs versioned. Everything undoable.

**Dedup-safe retries.** Retrying failed items automatically resolves conflicting duplicates created while the original was in failed status.

**Same API, two audiences.** Admin dashboard and future user portal share endpoints. Auth scope is the only difference.

---

## Future Work

| Item | Description |
|------|-------------|
| User portal (Phase 6) | Client auth, site-scoped permissions, simplified page browser |
| Structured nav editor | Edit `site_nav_items` as sortable list instead of raw header HTML |
| CSS variable editor | Parse `head` component CSS, present `--color-*` etc as form fields |
| Image/media manager | Browse `assets` table, upload/replace, see component references |
| Pin enforcement in agents | Classifier and improvement agents check `WHERE pinned = false` before overriding specs |
| Site-component lock in discovery | Add `AND sc.locked_at IS NULL` to `checkNavLayout`, `checkUnwantedElements`, `checkMissingAssetRefs` |
