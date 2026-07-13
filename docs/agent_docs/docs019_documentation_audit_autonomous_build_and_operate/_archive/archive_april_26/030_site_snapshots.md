# 026 — Site Snapshots and Point-in-Time Revert

## Date: 2026-03-26

---

## Purpose

Capture the complete state of a site at a point in time and revert to it later. A snapshot is a self-contained record of everything that defines a site: specs, pages, page components, navigation, and site-level components. Snapshots survive row deletions and schema changes because the data is stored as JSONB blobs, not foreign key references.

---

## What a Snapshot Captures

| Layer | Source table | Stored as |
|-------|-------------|-----------|
| Site record | `sites` | Key fields: domain, status, company_name, industry, schema_mode, default_components, content_data, timestamps |
| Spec aspects | `site_specs` (all rows where `is_current = true`) | Array of objects with aspect, data, source, pinned status |
| Pages | `pages` | All pages with name, url, title, type, status, nav settings, sections, rendered header/footer/head, page_spec, content_direction |
| Page components | `page_components` (per page) | Nested inside each page: component_id, position, slot_name, rendered_html, content_data, build_status |
| Navigation | `site_nav_groups` + `site_nav_items` | Groups with name/location/sort_order, items with label/url/page_id |
| Site components | `site_components` | Component_id, role, config, is_active |
| Git reference | Passed in by caller | `git_commit_sha` field links to the deployed file state |

Spec row IDs are also stored in a `spec_ids UUID[]` column for cross-referencing without parsing the JSONB.

---

## Database

### Table: `site_snapshots`

Created by migration `085_site_snapshots.sql`.

```
id                  UUID PRIMARY KEY
site_id             UUID NOT NULL → sites(id) ON DELETE CASCADE
trigger             TEXT NOT NULL        -- deploy, manual, pre_edit, scheduled, pre_revert
label               TEXT                 -- optional human label
git_commit_sha      TEXT                 -- optional link to git state
site_record         JSONB NOT NULL
spec_snapshot       JSONB NOT NULL
pages_snapshot      JSONB NOT NULL
nav_snapshot        JSONB NOT NULL
components_snapshot JSONB NOT NULL
spec_ids            UUID[]
created_at          TIMESTAMPTZ
created_by          TEXT NOT NULL
```

Indexes on `(site_id, created_at DESC)`, `git_commit_sha`, and `(trigger, created_at DESC)`.

### View: `v_site_snapshots`

Summary listing with domain name and array length counts:

```sql
SELECT * FROM v_site_snapshots;
```

Returns: id, site_id, domain, trigger, label, git_commit_sha, spec_count, page_count, created_at, created_by.

---

## SQL Functions

### `take_site_snapshot(site_id, trigger, git_sha, label, created_by)`

Captures the current state of a site into a single snapshot row. Returns the snapshot UUID.

```sql
SELECT take_site_snapshot(
    '2a8ebf9c-20a2-4c39-b191-840b012371da',
    'deploy',
    'abc123def456',
    'Post homepage redesign',
    'deploy-agent'
);
```

Parameters:

| Name | Type | Required | Description |
|------|------|----------|-------------|
| p_site_id | UUID | Yes | The site to snapshot |
| p_trigger | TEXT | Yes | Why the snapshot is being taken |
| p_git_sha | TEXT | No | Git commit SHA at time of snapshot |
| p_label | TEXT | No | Human-readable label |
| p_created_by | TEXT | No | Who initiated (default: 'system') |

Trigger values: `deploy`, `manual`, `pre_edit`, `scheduled`, `pre_revert`.

### `revert_site_to_snapshot(snapshot_id, reverted_by)`

Restores a site to the state captured in a snapshot. Returns a JSONB summary.

```sql
SELECT revert_site_to_snapshot(
    'a1b2c3d4-...',
    'admin'
);
```

What it does:

1. Takes a safety snapshot of the current state first (trigger = `pre_revert`)
2. Supersedes all current `site_specs` rows, inserts spec data from the snapshot
3. Deletes current `page_components` and `pages`, re-inserts from the snapshot (preserving original UUIDs)
4. Replaces `site_nav_groups` and `site_nav_items`
5. Replaces `site_components`
6. Updates `sites` row: status, schema_mode, default_components

What it does NOT do:

- Git revert — the caller must push the old files to GitHub separately if the deployed site needs to match
- Restore `content_components` templates — those are global (shared across sites), not per-site
- Restore `content_items` or `research_results` — these are input data, not site output

The safety snapshot means a revert is always reversible. If the revert produces a worse state, revert to the `pre_revert` snapshot.

Return value example:

```json
{
    "reverted": true,
    "snapshot_id": "a1b2c3d4-...",
    "safety_snapshot_id": "e5f6g7h8-...",
    "site_id": "2a8ebf9c-...",
    "specs_restored": 6,
    "pages_restored": 5,
    "components_restored": 18,
    "snapshot_trigger": "deploy",
    "snapshot_created_at": "2026-03-25T14:00:00Z"
}
```

---

## Admin API

All endpoints require admin authentication. Routes are under `/api/v1/admin/sites/:site_id/snapshots`.

### Take a snapshot

```
POST /api/v1/admin/sites/:site_id/snapshots
```

Body (all fields optional):

```json
{
    "trigger": "manual",
    "label": "Before redesign",
    "git_commit_sha": "abc123"
}
```

Returns `201` with `{ "id": "...", "site_id": "...", "trigger": "manual" }`.

### List snapshots

```
GET /api/v1/admin/sites/:site_id/snapshots
```

Returns up to 50 snapshots, newest first. Each entry includes spec_count, page_count, trigger, label, and timestamps but not the full JSONB blobs.

### Get snapshot detail

```
GET /api/v1/admin/sites/:site_id/snapshots/:snapshot_id
```

Returns the full snapshot including all JSONB blobs (site_record, spec_snapshot, pages_snapshot, nav_snapshot, components_snapshot).

### Revert to snapshot

```
POST /api/v1/admin/sites/:site_id/snapshots/:snapshot_id/revert
```

No body required. Returns the revert summary from `revert_site_to_snapshot`.

---

## Workflow Actions

Three actions are registered for use in agent workflows.

### `take_site_snapshot`

```json
{
    "action": "take_site_snapshot",
    "config": {
        "site_id_field": "site_record.site_id",
        "trigger": "deploy",
        "git_sha_field": "page_deployed.commit_sha",
        "label": "Post-deploy snapshot"
    },
    "output_field": "snapshot_result"
}
```

Config fields:

| Field | Default | Description |
|-------|---------|-------------|
| site_id_field | site_record.site_id | Dot-path to site UUID in collected_data |
| trigger | manual | Snapshot trigger type |
| git_sha_field | — | Dot-path to git SHA in collected_data |
| git_sha | — | Literal git SHA (field takes precedence) |
| label | — | Optional label |

### `revert_site_snapshot`

```json
{
    "action": "revert_site_snapshot",
    "config": {
        "snapshot_id": "a1b2c3d4-..."
    },
    "output_field": "revert_result"
}
```

Config fields: `snapshot_id_field` (dot-path) or `snapshot_id` (literal UUID).

### `list_site_snapshots`

```json
{
    "action": "list_site_snapshots",
    "config": {
        "site_id_field": "site_record.site_id",
        "limit": 10
    },
    "output_field": "available_snapshots"
}
```

---

## When Snapshots Happen

### Currently implemented

- Manual via SQL: `SELECT take_site_snapshot(...)` from psql
- Manual via admin API: `POST /admin/sites/:site_id/snapshots`
- Manual via workflow action: `take_site_snapshot` in any agent workflow
- Automatic before revert: `revert_site_to_snapshot` always takes a `pre_revert` snapshot first

### Recommended integration points

Add the `take_site_snapshot` action as a workflow step in these places:

1. **After deploy** — in the `build-dispatch-loop` or `pageflow-builder` workflow, after the final `git_commit` step succeeds. Pass the commit SHA through. This gives you a snapshot tied to every deployed version.

2. **Before spec propagation** — when the admin propagates a spec change via the dashboard, take a `pre_edit` snapshot before creating the work items. This lets you revert if the propagation produces bad results.

3. **Scheduled** — a nightly scheduled task that snapshots all deployed sites. Add a `scheduled_tasks` entry with `target_agent_type: snapshot-agent` or use a simple task that calls `take_site_snapshot` for each deployed site.

---

## File Locations

| File | Path |
|------|------|
| Migration | `platform/database/migrations/085_site_snapshots.sql` |
| Workflow actions | `platform/orchestration/actions/site_snapshot_actions.go` |
| Admin API handlers | `internal/core-manager/admin/snapshot_admin_handlers.go` |
| Registry entries | Added to `platform/orchestration/actions/registry.go` |
| Route registration | Added to `internal/core-manager/api/server.go` |

---

## Relationship to Existing Versioning

Site snapshots are the top-level coordination layer. The existing per-table versioning continues to work independently:

| Mechanism | Scope | Continues working? |
|-----------|-------|-------------------|
| `site_specs` is_current/superseded_at | Per-aspect | Yes — snapshots reference spec IDs but don't interfere |
| `page_component_history` | Per-component content | Yes — history rows are preserved, snapshots don't touch them |
| `component_versions` | Per-template HTML | Yes — global templates are outside snapshot scope |
| `agent_definitions` is_snapshot | Per-agent config | Yes — separate concern, agent-level not site-level |
| Git commits | Deployed files | Yes — snapshot stores the SHA for cross-reference |

The snapshot adds the missing "everything together at this moment" capability. Individual aspect rollbacks via `site_specs` versioning remain the right tool for targeted changes. Full site revert via snapshots is for when you need to undo a coordinated set of changes across specs, pages, and navigation together.
