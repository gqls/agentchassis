# Patch for 001_development_guide.md

Context: section 18 "Schema column renames — always check the live schema"
currently documents the `domain` → `pipeline` rename on `site_work_items`
as a case study. Add a second case study showing the `version_note` →
`change_description` rename on `component_versions` followed the same
pattern — and fixed in April 2026 alongside the regeneration work.

## Change

Find this block (around line 1467):

```markdown
**Root cause:** The `domain` column on site_work_items was renamed to `pipeline` in a migration. The project's `some_schemas` dump still showed `domain`. Go code written against the stale dump used the wrong column name. The INSERT silently failed, caught by error handling, and logged as a warning in pod logs (not in agent_error_log).

**Fix:** Always run `\d table_name` against the live database before writing SQL or Go code that references columns. Never trust cached schema dumps in the repository.

**Rule:** The live database is the source of truth for column names. Schema dumps go stale. `\d site_work_items` takes 2 seconds and prevents hours of debugging invisible failures.
```

Replace with:

```markdown
**Root cause:** The `domain` column on site_work_items was renamed to `pipeline` in a migration. The project's `some_schemas` dump still showed `domain`. Go code written against the stale dump used the wrong column name. The INSERT silently failed, caught by error handling, and logged as a warning in pod logs (not in agent_error_log).

**Fix:** Always run `\d table_name` against the live database before writing SQL or Go code that references columns. Never trust cached schema dumps in the repository.

**Rule:** The live database is the source of truth for column names. Schema dumps go stale. `\d site_work_items` takes 2 seconds and prevents hours of debugging invisible failures.

### Second case: version_note → change_description on component_versions (April 2026)

`UpdateComponentHTMLAction` wrote to `component_versions.version_note`.
The live schema had `change_description`. The INSERT failed on every
call. But because the snapshot was marked best-effort (the failure was
logged as a Warn and ignored), the bug was invisible in normal flow —
no observable failure, just silent data loss. Every component update
that should have left a history row left nothing.

This was discovered by code audit during the April 2026 component
regeneration work, when we added `StoreGeneratedComponentAction`
regeneration semantics that rely on snapshots. Two years' worth of
`update_component_html` calls produced no `component_versions` rows
before the fix.

**Lesson:** `best-effort` operations need active monitoring. A
persistent snapshot failure should eventually surface — either in
metrics (e.g. "snapshots attempted vs committed" ratio) or via a
smoke test that asserts at least ONE row got written. Silent
best-effort is really "silent no-effort" when the code path is
always broken.
```

## Rationale

The first case study (`domain` rename) teaches the "check live schema"
rule via an immediate failure. The second case (`version_note` rename)
teaches the less-obvious lesson: best-effort operations can fail
100% of the time without any observable symptom. Together they cover
both the "schema renames exist, look for them" and "best-effort
operations need monitoring" angles.

The April 2026 dating locates the fix in time so anyone following a
git blame trail can find the concrete change.
