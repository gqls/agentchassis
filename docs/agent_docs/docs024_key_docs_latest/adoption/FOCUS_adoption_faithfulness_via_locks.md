# FOCUS — Adoption Faithfulness via Timed Locks

Date: 2026-05-19
Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending.

## The requirement

The first pass of an adoption should be faithful to the source site — the only
changes being components needed to fit our code and making tools/games work
(rewriting if necessary). After that initial window, the site should develop
according to its own spec and plan like any other site, including eventually
editing, restructuring, and deleting pages.

So the protection must be **temporary and self-releasing**, not a permanent
flag. That is exactly a *timed lock* — the mechanism doc 004 v4 designed and
docs 031_locks / 031_LOCKS_should_locks_expire approved.

## Why locks, not a permanent adopted-flag

A permanent "this site was adopted" flag would either protect adopted pages
forever (the site never develops) or require manual clearing (operational
burden, easy to forget). A timed lock:

- Protects the adopted pages during the initial window
- Releases automatically — no human action, no flag to clear
- Ages out on the same rhythm as deploy-locks and auditor approvals, so the
  whole site "breathes" together (doc 004 v4's intended rhythm)
- Coexists with HITL: if the user hand-edits an adopted page during the window,
  that edit gets a `permanent` (admin) lock that outlives the adoption window —
  exactly right

## The two options considered

**Option A — implement the full timed-lock-expiry project now.**
Docs 031 sequenced this *after* the imagery loop and flagged it as
"anticipatory, shouldn't displace observed-problem work." But adoption-
faithfulness is now an *observed* driver (this conversation), which changes the
cost-benefit. Doc 031 estimated it as a focused 1-2 day piece. It also insisted
the schema change be unified across all four Pattern A tables in one migration
("Coherence beats incremental delivery here") to avoid the situation where one
table has `lock_expires_at` and others don't.

**Option B — interim: a single expiry timestamp on adoption directives only.**
Smaller, but introduces exactly the partial state docs 031 warned against (one
table with expiry semantics, others without), and consumers would have to know
which is which. A deliberate temporary divergence.

**Decision (2026-05-19): Option A, 90-day adoption window.** The full project,
done coherently across all four tables, because adoption-faithfulness is a real
driver and the partial-state risk of B isn't worth the small saving.

## Approved policy table (with `adoption` added)

| locked_by | lock_type | expiry |
|---|---|---|
| admin / admin-removed / checkpoint | permanent | — |
| manual | permanent | — |
| deploy (auto-lock-on-deploy) | timed | +30 days |
| visual-design-auditor / imagery-quality-auditor | timed | +90 days |
| **adoption (faithful first pass)** | **timed** | **+90 days** |
| audit-pending | not a lock (cleared on completion) | — |

Human-set locks remain permanent. Only auto-locks, auditor approvals, and now
adoption get timed expiry.

## The "is this row improvable?" predicate

Expands from `locked_at IS NULL` to:

```sql
(locked_at IS NULL
 OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))
```

Permanent/human locks (`lock_expires_at` NULL) are never improvable by agents.

## How this drives adoption-faithfulness end to end

REVISED 2026-05-19 after schema check: `site_plan_directives` is **plan-scoped**
(keyed by plan_id), and adoption writes pages + specs but **not** plans or
directives. So the lock cannot originate at adoption time — there's no plan to
attach a directive to. It originates instead at the planner's first
`write_site_plan`, detected deterministically.

1. **First plan after adoption — the origination.**
   - `load_existing_pages` returns each page with `adoption_locked`. On the
     first plan there is **no current `site_plan`** yet, so the query marks all
     existing pages `adoption_locked = true`. (This branch only ever fires for
     adopted sites: from-scratch sites have no pages until the planner's own
     `sync_pages` runs, so "no current plan + pages exist" uniquely means
     "adopted pages, first plan".)
   - `validate_site_plan` convergence preserves all of them.
   - `write_site_plan` emits a page-scoped `preserve` directive per page and,
     because `prevPlanID == nil` and existing pages were present, **locks** them
     `locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at = NOW()+90d`.

2. **Re-plans within the 90-day window.**
   - `load_existing_pages` now finds live adoption preserve-directives on the
     current plan → `adoption_locked = true` for those pages.
   - convergence preserves them; `write_site_plan` re-emits the preserve-
     directives and `transferDirectiveLocks` carries the locks (now including
     `lock_type` + `lock_expires_at`) from the previous plan onto the new one.

3. **After the window.**
   - `lock_expires_at < NOW()` → `load_existing_pages` computes
     `adoption_locked = false` → convergence is a no-op → the planner and
     improvement loop develop the site normally, including
     deleting/editing/restructuring pages.

4. **From-scratch builds.** No current plan AND no pages on first plan →
   `existing_pages` empty → convergence no-op. Subsequent plans: no adoption
   preserve-directives → `adoption_locked = false`. Never constrained.

The Go convergence layer doesn't special-case adoption at all — it preserves
whatever the query flags `adoption_locked`. The faithful↔normal boundary is a
single `lock_expires_at > NOW()` comparison plus the "no current plan yet"
first-plan branch, both in one query (054).

## Dependency chain

```
053 schema (lock_type + lock_expires_at on 4 tables)        ← written, ready to apply
  └─> 054 load_existing_pages exposes adoption_locked         ← written (SQL)
       (first-plan branch + live-adoption-directive branch)
  └─> write_site_plan: emit + first-plan-lock preserve-directives,
       transferDirectiveLocks carries lock_type/expiry          ← written (patch doc)
  └─> validate_site_plan convergence keys off adoption_locked   ← done (v3_site_actions.go)
  └─> filter sweep: 11 `locked_at IS NULL` callsites             (Go, broader project)
  └─> CheckComponentLock returns LockType + LockExpiresAt        (Go, broader project)
  └─> expired_review_locks discovery check                       (Go, broader project; not on
                                                                  adoption critical path)
```

Note: the adoption lock originates in `write_site_plan` (first plan after
adoption), NOT in adoption itself — because `site_plan_directives` is
plan-scoped and adoption creates no plan. There is no adoption-side Go change.

## Implementation plan (Option A, ordered)

1. **[done] 053 schema migration** — `lock_type` + `lock_expires_at` on
   `page_components`, `site_components`, `site_plan_directives`, `assets`;
   CHECK constraint; backfill existing locked rows (conservative: existing
   timed locks get expiry from NOW(), so nothing releases on migration day);
   partial indexes for cheap expiry sweeps.

2. **[done] Convergence layer (Go).** `v3_site_actions.go` —
   `ValidateSitePlanAction` reconciles against `adoption_locked` pages
   (union / rename snap-back / section-collision dedup) and truncates
   preserving locked pages. Inert until 054 + write_site_plan land.

3. **[done] 054 load_existing_pages query (SQL).** Exposes `adoption_locked`
   via the first-plan branch (no current plan) and the live-adoption-directive
   branch.

4. **[done, patch doc] write_site_plan changes.** Emit a page-scoped `preserve`
   directive per page; lock them adoption/timed/90d on the first plan after
   adoption (`prevPlanID == nil` AND existing pages present); extend
   `transferDirectiveLocks` to carry `lock_type` + `lock_expires_at`. See
   `write_site_plan_adoption_patch.md`.

5. **Filter sweep (Go).** Replace the 11 `locked_at IS NULL` callsites with the
   expanded predicate `(locked_at IS NULL OR (lock_type='timed' AND
   lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))`. Mechanical;
   discovery/audit data-loading queries. Completes the broader lock-expiry
   project; not strictly required for adoption-faithfulness (those callsites are
   page_components, not the preserve-directives).

6. **CheckComponentLock extension (Go).** Add `LockType` and `LockExpiresAt` to
   `ComponentLockStatus`; SELECT them; treat a timed lock past expiry as
   unlocked. Hard locks ignore expiry. Keep the hard/soft `locked_by`
   classification. Broader project.

7. **write_site_plan page_type / ValidateRoles verification.** Confirm
   write_site_plan reads `page_type` from the unioned preserve pages and that
   ValidateRoles is idempotent on already-canonical pages (no
   `guides-index` → `guides-index-index`). One test run.

8. **expired_review_locks discovery check (Go).** For `review`-type locks past
   expiry, create `needs_lock_review` work items — the HITL hook doc 004 v4
   described. Lower priority; `adoption` uses `timed` not `review`, so this
   isn't on the adoption-faithfulness critical path.

9. **Dashboard view (downstream).** Show locks with classification and expiry.
   Not blocking.

Critical path for adoption-faithfulness specifically: **1 → 3 → 4 → 2** (schema,
054 query, write_site_plan lock origination, convergence — the last already
done). Steps 5, 6, 8, 9 complete the broader lock-expiry project but aren't
required for the faithful-first-pass behaviour. Step 7 is a one-time verification.

## Boundaries / caveats

- **Window length.** 90 days agreed. Long enough for the user to see the
  faithful copy and decide their own direction before agents restructure it.
- **Restructure during the window.** While locked, the planner can't remove
  adopted pages. If a user wants to restructure inside the window, they edit via
  HITL (which sets a permanent admin lock on what they touch) or manually clear
  the adoption lock. Deliberate: the default is "stay faithful for 90 days."
- **Backfill basis.** Existing timed locks get expiry from NOW() (grace window),
  not from `locked_at`, so the migration doesn't trigger a mass lock release on
  day one. New locks measure from `locked_at` via the writers.
- **Coherence.** All four tables get the columns in one migration (053), per
  docs 031's insistence — no partial state.

## References

- `004_improvement_loop.md` v4 — original timed-lock design, audit-pass rhythm
- `031_locks.md` — canonical lock semantics, Pattern A, hard vs soft
- `031_LOCKS_should_locks_expire.md` — investigation, approved policy, sequencing
- `031_locks_proposed_update.md` — lock lifecycle table, approved policy
- `FOCUS_planner_ignores_adopted_state.md` — the duplication problem this protects against
- `052_planner_reads_realised_state.sql` — planner reads realised state (max_pages 80, max_tokens 16000)
- `053_lock_expiry.sql` — the schema migration (this project's foundation)
- `validate_site_plan_convergence.go` — deterministic convergence, keyed off adoption_locked
- Code: `check_component_lock.go` (CheckComponentLock, hard/soft), 11 `locked_at IS NULL` callsites
- This conversation, 2026-05-19

---

# Implementation — write_site_plan_action.go patch

> Folded in from `write_site_plan_adoption_patch.md` (the newer revision that
> skips already-expired timed locks in the transfer SELECT). This is the code
> for step 4 of the implementation plan above — three coordinated changes in
> `platform/orchestration/actions/write_site_plan_action.go`. The lock
> ORIGINATES here, on the first plan after adoption (`prevPlanID == uuid.Nil`
> AND existing pages present); re-plans carry it via `transferDirectiveLocks`.

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

### 3a — SELECT (add the two columns; skip already-expired timed locks)

```go
rows, err := tx.QueryContext(ctx, `
    SELECT scope, scope_ref, category, subject, ordering, directive,
           locked_at, locked_by, lock_type, lock_expires_at
    FROM site_plan_directives
    WHERE plan_id = $1
      AND locked_at IS NOT NULL
      -- Don't propagate already-expired timed locks: they should release,
      -- not chain forward as dead rows (doc 031 "expired locks release").
      AND (lock_type IS DISTINCT FROM 'timed'
           OR lock_expires_at IS NULL
           OR lock_expires_at > NOW())
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
