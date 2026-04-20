# Patch for 014_site_snapshots_and_revert.md

Context: the "Relationship to Existing Versioning" table describes
`component_versions` as "per-template HTML" that's "outside snapshot
scope". That's still accurate, but since April 2026 this table has
additional population paths worth noting.

## Change

After the versioning table (around line 290), before the final
paragraph, ADD a new subsection:

```markdown
### Populating component_versions

The `component_versions` table has three population paths, all
best-effort (a failed snapshot does not block the operation that
prompted it):

| Trigger | Writer | changed_by value | change_source |
|---------|--------|------------------|---------------|
| New component created | `StoreGeneratedComponentAction` | `component-creator:create` | work item's source, if any |
| Component regenerated | `StoreGeneratedComponentAction` | `component-creator:regen` | work item's source |
| Tool HTML improvement | `UpdateComponentHTMLAction` | `update_component_html` | caller-provided note |

The `change_source` column (added April 2026) captures the originating
work item's `source` field, so a version history entry can be traced
back to the audit, triage, or manual trigger that caused the change.
Null is allowed for legacy rows and direct programmatic edits that
don't originate from a work item.

Version numbers are monotonically increasing per component, computed
as `MAX(version_number) + 1` at write time. A unique index on
`(component_id, version_number)` rejects duplicates — concurrent
regenerations of the same component will produce one successful
snapshot and one logged-and-ignored conflict.
```

## Rationale

The existing table correctly notes component_versions is outside
snapshot scope. The new subsection documents the data that IS in the
table, so readers understand what component_versions DOES capture
and how to trace a change back to its trigger. This is useful context
for anyone doing forensics on "why did this component change?" or
building admin UI to browse version history.

The subsection is additive — no existing content changes, no
inaccuracies introduced. The table on line 286 stays as-is.
