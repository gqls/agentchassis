# Patch: design_actions.go — fix the status filter

## What to change

In `loadPagesWithComponents`, replace:

```go
		WHERE p.site_id = $1
		  AND p.status IN ('deployed', 'published', 'draft', 'planned')
```

with:

```go
		WHERE p.site_id = $1
		  AND p.status = 'active'
```

(Or simply remove the status filter entirely — there are no other statuses in
use right now, and excluding archived/deleted pages can be a separate concern.
`status = 'active'` matches the pattern `LoadSitePagesAction` already uses.)

## Why

Diagnostic confirmed all 29 pages on gaswholesalers (and likely every other
site in the database) have `status = 'active'`. The names I assumed
(`'deployed'`, `'published'`, `'draft'`, `'planned'`) don't exist in the data.
My filter excluded everything, `loadPagesWithComponents` returned an empty
slice, `allComponents` stayed empty, fallback fired.

The original (broken) `loadPagesWithComponents` had the same status filter,
so users never noticed — the bigger bug (reading from `pages.sections`)
masked it. My new code triggers the loud Warn ("NO COMPONENTS FOUND") which
made the empty result visible.

## Verification before the fix lands

Quick check on at least one other site, to confirm the same status is used
across the platform:

```sql
SELECT DISTINCT status, COUNT(*) AS pages
FROM pages
GROUP BY status
ORDER BY pages DESC;
```

If `'active'` is the only or dominant value, the fix is safe to apply.
If there are also other statuses worth including (e.g. some sites use
'published' or 'live'), the WHERE clause should accommodate them.

## Lesson for the debugging guide

Add to Section 9 of the debugging guide:

> ### Assumed-status-values trap
>
> When writing or modifying SQL that filters by a status column, ALWAYS
> query `SELECT DISTINCT status FROM <table>` first to see what values
> actually exist. The pages.status column uses `'active'` exclusively;
> the values `'deployed'`, `'published'`, `'draft'`, `'planned'` that
> appeared in older queries don't exist in the data. Pattern-matching
> status names from other systems or from how things "should" work
> leads to silent zero-result queries.
>
> The status filter on `loadPagesWithComponents` excluded every page on
> every site for the duration of this filter being in the function. The
> upstream bug (reading from `pages.sections`) masked it for months —
> when that was fixed, this filter then became the next layer of the
> onion.
