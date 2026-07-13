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

1. **Adoption writes a per-page preserve directive** for each adopted page into
   `site_plan_directives`, locked `locked_by='adoption'`, `lock_type='timed'`,
   `lock_expires_at = NOW() + 90 days`.

2. **`load_existing_pages`** (the planner step added in 052) LEFT JOINs those
   directives and exposes an `adoption_locked` boolean per page — true while the
   lock is live, false once expired.

3. **`validate_site_plan` convergence layer** (drafted in
   `validate_site_plan_convergence.go`) force-preserves ONLY the
   `adoption_locked` pages:
   - During the 90-day window: every adopted page is locked → the whole adopted
     site is preserved faithfully. The planner LLM physically cannot drop,
     rename, or duplicate any of it (Pass A union, Pass B rename snap-back,
     Pass C section-collision dedup).
   - After the window: no page is `adoption_locked` → the convergence layer is a
     no-op → the planner and improvement loop develop the site normally,
     including deleting/editing/restructuring pages.
   - From-scratch builds: never `adoption_locked` → always a no-op.

This is the key elegance: **the timed lock is enforced at the data layer (the
query's `adoption_locked` flag), so the Go convergence logic doesn't special-
case adoption at all — it just preserves whatever is currently locked.** The
90-day boundary is a single `lock_expires_at > NOW()` comparison in one query.

## Dependency chain

Adoption-faithfulness depends on the timed-lock-expiry project:

```
053 schema (lock_type + lock_expires_at on 4 tables)   ← written, ready to apply
  └─> adoption writes per-page preserve directives (Go, adoption side)
       └─> load_existing_pages JOIN exposes adoption_locked (052 query update)
            └─> validate_site_plan convergence keys off adoption_locked (Go, drafted)
  └─> filter sweep: 11 `locked_at IS NULL` callsites get the expanded predicate (Go)
  └─> CheckComponentLock returns LockType + LockExpiresAt; honours expiry (Go)
  └─> expired_review_locks discovery check creates needs_lock_review (Go, the HITL hook)
```

## Implementation plan (Option A, ordered)

1. **[done] 053 schema migration** — `lock_type` + `lock_expires_at` on
   `page_components`, `site_components`, `site_plan_directives`, `assets`;
   CHECK constraint; backfill existing locked rows (conservative: existing
   timed locks get expiry from NOW(), so nothing releases on migration day);
   partial indexes for cheap expiry sweeps.

2. **Filter sweep (Go).** Replace the 11 `locked_at IS NULL` callsites with the
   expanded predicate. Mechanical; catalogue from grep. These are in the
   discovery/audit data-loading queries (`check_unlinked_page_components`, the
   empty-section finder, visual/content auditor queries, etc.).

3. **CheckComponentLock extension (Go).** Add `LockType` and `LockExpiresAt` to
   `ComponentLockStatus`; SELECT them; treat a timed lock past expiry as
   unlocked. Hard locks ignore expiry. Keep the existing hard/soft `locked_by`
   classification.

4. **Adoption directive writer (Go, adoption side).** When adoption applies the
   plan, write one `site_plan_directives` row per adopted page:
   `locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at=NOW()+interval
   '90 days'`, keyed so `load_existing_pages` can JOIN it (e.g. directive_key
   `preserve_page:<name>`). Confirm the `site_plan_directives` schema (columns,
   key shape) before writing — standing rule.

5. **load_existing_pages JOIN (SQL, 052 follow-up).** Update the query to expose
   `adoption_locked` (see the convergence file's integration note for the JOIN).

6. **Convergence layer integration (Go).** Wire
   `validate_site_plan_convergence.go` into `ValidateSitePlanAction` per its
   integration block. Verify write_site_plan handles page_type-only pages and
   ValidateRoles idempotency.

7. **expired_review_locks discovery check (Go).** For `review`-type locks past
   expiry, create `needs_lock_review` work items — the HITL hook doc 004 v4
   described. Lower priority; `adoption` uses `timed` not `review`, so this
   isn't on the adoption-faithfulness critical path.

8. **Dashboard view (downstream).** Show locks with classification and expiry.
   Not blocking.

Critical path for adoption-faithfulness specifically: 1 → 4 → 5 → 6. Steps 2, 3,
7, 8 complete the broader lock-expiry project but aren't strictly required for
the faithful-first-pass behaviour.

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
