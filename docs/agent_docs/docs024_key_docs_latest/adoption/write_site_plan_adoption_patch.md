# Patch — write_site_plan_action.go: adoption preserve-directives + timed-lock transfer

Part of the adoption-faithfulness critical path (FOCUS_adoption_faithfulness_via_locks.md).
Three coordinated changes in `platform/orchestration/actions/write_site_plan_action.go`.

The adoption lock ORIGINATES here, on the first plan after adoption, detected by
`prevPlanID == uuid.Nil` (no prior plan) AND existing pages present (only true
after adoption — from-scratch sites have no pages until this action's own
sync_pages runs). Re-plans within the 90-day window carry the lock forward via
the (extended) transferDirectiveLocks.

---

## Change 1 — emit a preserve-directive per page

After `directives := flattenSiteScopeDirectives(params.CollectedData, logger)`
(currently ~line 86173 of the consolidated context), append one page-scoped
preserve-directive per planRow. These give transferDirectiveLocks a matching
row to land the lock on across every plan version.

```go
directives := flattenSiteScopeDirectives(params.CollectedData, logger)

// Emit a page-scoped "preserve" directive per page. On the first plan after
// adoption these get locked (Change 2) so the adopted pages stay faithful;
// on later plans they are the transfer target for the carried-forward lock.
for _, r := range planRows {
    name := r.Name // local copy — do not take &r.Name (loop-var aliasing)
    directives = append(directives, directiveRow{
        Scope:     "page",
        ScopeRef:  &name,
        Category:  "preserve",
        Subject:   "exists",
        Directive: "Preserve adopted page (faithful first pass).",
        Ordering:  0,
        Source:    "adoption-faithfulness",
    })
}
```

---

## Change 2 — lock the preserve-directives on the first plan after adoption

The directives are inserted in the existing loop (the `INSERT INTO
site_plan_directives ... VALUES (...)` block, ~line 86288). AFTER that loop,
and BEFORE the `transferDirectiveLocks` call (~line 86310), add a first-plan
lock step.

Detect "first plan after adoption": no prior plan (`prevPlanID == uuid.Nil`)
AND adoption left pages behind. Read existing_pages from collected_data for the
second condition (it was loaded by the load_existing_pages step).

```go
// First plan after adoption: lock the page preserve-directives so adopted
// pages stay faithful for 90 days. Detected by "no prior plan + pages already
// exist" (only true after adoption; from-scratch sites have no pages yet).
// Re-plans within the window rely on transferDirectiveLocks instead.
if prevPlanID == uuid.Nil {
    hadExistingPages := false
    if ev := datahelpers.ExtractNestedField(params.CollectedData, "existing_pages"); ev != nil {
        if ep, ok := ev.([]interface{}); ok && len(ep) > 0 {
            hadExistingPages = true
        }
    }
    if hadExistingPages {
        _, err = tx.ExecContext(ctx, `
            UPDATE site_plan_directives
               SET locked_at       = NOW(),
                   locked_by       = 'adoption',
                   lock_type       = 'timed',
                   lock_expires_at = NOW() + interval '90 days'
             WHERE plan_id  = $1
               AND scope    = 'page'
               AND category = 'preserve'
        `, planID)
        if err != nil {
            return nil, fmt.Errorf("lock adoption preserve-directives: %w", err)
        }
        logger.Info("WriteSitePlanAction: locked adoption preserve-directives (faithful first pass)",
            zap.String("plan_id", planID.String()))
    }
}
```

(`planID` is the new plan's id, already in scope where the directives are
inserted. `datahelpers` and `zap` are already imported in this file.)

---

## Change 3 — transferDirectiveLocks carries lock_type + lock_expires_at

Today it copies only `locked_at`, `locked_by`, `directive`. A timed adoption
lock would lose its timed-ness (and thus never expire correctly, or be treated
as permanent) on the first re-plan. Extend the SELECT and both UPDATE branches.

### 3a — SELECT (add the two columns)

```go
rows, err := tx.QueryContext(ctx, `
    SELECT scope, scope_ref, category, subject, ordering, directive,
           locked_at, locked_by, lock_type, lock_expires_at
    FROM site_plan_directives
    WHERE plan_id = $1 AND locked_at IS NOT NULL
`, prevPlanID)
```

### 3b — scan targets

```go
var (
    scope, category, subject, lockedBy, prevText string
    scopeRef                                     sql.NullString
    ordering                                     int
    lockedAt                                     sql.NullTime
    lockType                                     sql.NullString   // NEW
    lockExpiresAt                                sql.NullTime     // NEW
)
if err := rows.Scan(&scope, &scopeRef, &category, &subject, &ordering, &prevText,
    &lockedAt, &lockedBy, &lockType, &lockExpiresAt); err != nil {   // NEW two targets
    return transferred, fmt.Errorf("scan previous lock: %w", err)
}
```

### 3c — UPDATE (both the scope_ref-present and scope_ref-NULL branches)

scope_ref present branch:

```go
result, err = tx.ExecContext(ctx, `
    UPDATE site_plan_directives
       SET locked_at       = $1,
           locked_by       = $2,
           directive       = $3,
           lock_type       = $4,
           lock_expires_at = $5
     WHERE plan_id   = $6
       AND scope     = $7
       AND scope_ref = $8
       AND category  = $9
       AND subject   = $10
       AND ordering  = $11
`, lockedAt, lockedBy, prevText, lockType, lockExpiresAt,
   newPlanID, scope, scopeRef.String, category, subject, ordering)
```

scope_ref NULL branch:

```go
result, err = tx.ExecContext(ctx, `
    UPDATE site_plan_directives
       SET locked_at       = $1,
           locked_by       = $2,
           directive       = $3,
           lock_type       = $4,
           lock_expires_at = $5
     WHERE plan_id   = $6
       AND scope     = $7
       AND scope_ref IS NULL
       AND category  = $8
       AND subject   = $9
       AND ordering  = $10
`, lockedAt, lockedBy, prevText, lockType, lockExpiresAt,
   newPlanID, scope, category, subject, ordering)
```

`sql.NullString` / `sql.NullTime` pass through as NULL when not set, so
non-timed (permanent / human) locks keep `lock_expires_at = NULL` correctly.

---

## Why this is safe before the rest of the pipeline lands

- Change 1 emits preserve-directives unconditionally, but they are unlocked
  unless Change 2 fires. Unlocked directives are inert — nothing reads
  `category='preserve'` yet except the 054 load_existing_pages query.
- Change 2 only fires on first-plan-after-adoption. For from-scratch first
  plans (no existing pages) and all re-plans (prevPlanID set), it is a no-op.
- Change 3 is backward-compatible: existing locked directives have NULL
  lock_type today (until 053 backfills them); NULL passes through unchanged.

## Verify after deploy

```sql
-- After a fresh adoption + first plan, the adopted pages should have locked
-- preserve-directives with a ~90-day expiry.
SELECT d.scope_ref AS page, d.locked_by, d.lock_type, d.lock_expires_at
FROM site_plan_directives d
JOIN site_plans sp ON sp.id = d.plan_id AND sp.is_current = true
WHERE sp.site_id IN (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND d.scope = 'page' AND d.category = 'preserve'
ORDER BY d.scope_ref;
-- expect one row per adopted page, locked_by='adoption', lock_type='timed',
-- lock_expires_at ≈ now + 90 days.
```
