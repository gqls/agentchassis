# 029 — Locks: HITL durability across the platform

How the platform protects human-edited or human-approved data from being silently overwritten by agents — across components, specs, plans, and sites.

This doc is the canonical reference for lock semantics. Other docs (007 adoption, 030 plan reconciler, 003 contracts) reference locks in passing; the rules and patterns live here.

---

## Why locks exist

Agents iterate. The improvement loop runs audits, fixes, rerenders. The plan-builder rebuilds plans when direction changes. Adoption re-runs when a source site is re-crawled. Each of these can rewrite data that a human cared enough to touch.

Without locks, every human edit is fragile — the next agent run reverts it. With locks, a human can mark a value (a component template, a spec field, a directive line, a whole site) as "this version, keep it" and the iteration machinery respects that.

The cost is mechanism: every writer that could touch lockable data has to know about locks. Keeping the patterns small and the rules consistent matters more than the locks themselves.

---

## The two patterns in use

The platform uses two complementary patterns. Choose the one that matches the scope of what's being locked.

### Pattern A — `locked_at` + `locked_by` (per-row lock)

Two columns on the row:

```
locked_at  timestamptz       -- when the lock was applied; NULL = unlocked
locked_by  text              -- who/what applied it (user id, agent name, "manual")
```

Plus a partial index for cheap lookups of locked rows:

```sql
CREATE INDEX idx_<table>_locked
  ON <table> (...)
  WHERE locked_at IS NOT NULL;
```

Used by: `sites`, `page_components`, `site_components`, `assets` (Phase 2A,
2026-05-08), and (Phase 1) `site_plan_directives`.

This is the dominant pattern. Use it whenever the lockable thing is a row that itself represents a discrete unit of work — a component instance, a directive, a site record.

**Hard vs soft locks.** The `locked_by` value distinguishes locks that
only humans can clear (hard) from locks that agents can clear when doing
legitimate work (soft). Encoded by convention:

| `locked_by` value | Class | Who can clear |
|---|---|---|
| `admin` | hard | human only |
| `admin-removed` | hard | human only |
| `checkpoint` | hard | human only |
| `deploy` | soft | agents can clear |
| anything else | soft | agents can clear |

The shared helper `CheckComponentLock` in
`platform/orchestration/actions/check_component_lock.go` returns this
classification via the `IsHard` field on `ComponentLockStatus`. Discovery
agents skip both hard and soft locks (they don't write, just observe).
Execution agents skip hard locks entirely; they may clear and proceed
through soft locks when the work item is explicit and recent.

**For `assets` (Phase 2A)**, the `locked_by` vocabulary extends:

| `locked_by` value | Class | Source |
|---|---|---|
| `manual` | hard | human upload via dashboard |
| `admin` | hard | human edit via dashboard |
| `visual-design-auditor` | soft | Phase 4 auditor approved |
| `imagery-quality-auditor` | soft | Phase 6 auditor (future) |
| `audit-pending` | soft | transient, set at audit start, cleared on completion |

These are convention not constraint — no CHECK is added so future callers
can introduce new identifiers without a schema change.

### Pattern B — `pinned boolean` (per-row flag)

One column on the row:

```
pinned     boolean default false
```

Used by: `site_specs` only. A pinned spec row is treated as authoritative; the next write of the same `(site_id, aspect)` skips the upsert when the existing row is pinned.

Pattern B is simpler but carries less metadata (no who/when). It exists because `site_specs` predates the `locked_at`/`locked_by` convention and migration risk wasn't worth the small metadata gain. **New tables should not use Pattern B.** If you're adding lockability to a new table, use Pattern A.

---

## Where locks live (current state)

| Table | Pattern | What gets locked | Set by |
|---|---|---|---|
| `sites` | A | Identity fields (logo_url, contact_address, company_name) when human-set | HITL UI (planned) |
| `page_components` | A | A specific section instance on a specific page (e.g. the "approved" hero copy) | Improvement-loop discovery checks; HITL UI |
| `site_components` | A | A site-wide component (e.g. the header used everywhere) | HITL UI |
| `assets` | A | An image (logo, hero, hero variant) marked authoritative — protects against the regeneration loop overwriting human-uploaded or auditor-approved imagery | HITL upload UI; visual-design-auditor (Phase 4); imagery-quality-auditor (Phase 6) |
| `site_specs` | B | A particular aspect (identity, design_intent, voice) marked authoritative | HITL UI; some agents on completion |
| `site_plan_directives` (Phase 1) | A | A specific directive at site / page / section scope | HITL UI; lock transfer in `write_site_plan` |

---

## The rules every writer must follow

Whatever the pattern, every action that updates a lockable row must know two things:

1. **Read the lock state before deciding to write.** A locked row's value is owned by whoever locked it; the agent's new value is at most a candidate.
2. **Preserve the lock state when superseding.** Updates by the lock owner (or with the owner's permission) clear the lock; updates by automated machinery either skip the write or pass through without touching `locked_at` / `locked_by`.

Patterns differ in the mechanics of (1) and (2):

### Pattern A writers

```sql
-- option 1: skip the write entirely if locked
UPDATE <table>
   SET <fields...>
 WHERE id = $1
   AND locked_at IS NULL;

-- option 2: read first, then decide in code
SELECT locked_at, locked_by FROM <table> WHERE id = $1;
-- if locked, log and skip; if not, write
```

Most code uses option 1 — single-statement, atomic, no race window. Use option 2 only when you need to log specifically that a lock blocked the write.

### Pattern B writers

```sql
INSERT INTO site_specs (...)
VALUES (...)
ON CONFLICT (site_id, aspect)
WHERE is_current = true AND pinned = false
DO UPDATE SET ...;
```

The partial conflict target makes pinned rows invisible to the upsert. New writes become orphans (current-but-not-the-one-the-conflict-saw) — practical concern depends on the consumer. The simpler approach is to read first:

```sql
SELECT pinned FROM site_specs
 WHERE site_id = $1 AND aspect = $2 AND is_current = true;
-- if pinned, skip; otherwise upsert
```

---

## Lock transfer across rebuilds

Some lockable rows are short-lived — they're written fresh by an agent that rebuilds the parent record. The most relevant case is `site_plan_directives`: every `write_site_plan` run produces a new plan row with new directive rows; the previous plan's directives become history. A lock on a previous-plan directive is meaningless once the plan is superseded — unless something transfers it.

The pattern: **the rewriting agent is the only one that knows about lock transfer**. No reader downstream needs to merge across plan versions or follow chains of locked-history rows. The rewriting agent looks up the previous version's locked rows and copies the lock onto the matching new rows by a stable composite key.

For `site_plan_directives` (Phase 1), the match key is `(scope, scope_ref, category, subject, ordering)` under the same site. The action `write_site_plan` runs the transfer in the same transaction as the new plan write. When the new text from the LLM differs from the locked text, the locked text wins (HITL approval > LLM rewrite).

When the new plan has no row matching a previous locked row's key, the lock is dropped and a log line records it. The history of the previous plan persists (`is_current = false`, `superseded_at` set), so audit retrieval is possible.

This pattern doesn't apply to `page_components` or `site_components` because those rows survive across rebuilds — the same row gets updated rather than replaced. Transfer is needed only when row identity is per-version.

---

## Lock lifecycle (timed and review locks)

Per doc 007 (adoption pipeline) v3/v4 and doc 004 (improvement loop) v4
the platform anticipates richer lock types beyond "permanent until
manually unlocked":

| Lock type | Behaviour | Use case |
|---|---|---|
| `permanent` | Never expires; manual unlock only | Brand elements, legal disclaimers, human-crafted content |
| `timed` | Expires after N days (e.g. 90) | HITL-requested content that should eventually re-enter improvement cycle |
| `review` | Creates HITL review item on expiry | Content needing human approval before agents touch it again |

These are not all implemented yet. When they are, the column shape on Pattern A tables grows by two fields:

```
lock_type         text            -- 'permanent' | 'timed' | 'review'
lock_expires_at   timestamptz     -- when the lock should release (null for permanent)
```

And the read condition for "is this row locked?" expands from `locked_at IS NULL` to `(locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at < NOW()))`. Discovery checks can then auto-create `needs_lock_review` work items for review-type locks at expiry.

### State today (2026-05-08)

**No time-based expiry exists in code or schema.** Confirmed against all
four Pattern A tables (`page_components`, `site_components`,
`site_plan_directives`, `assets`) and all production query call sites.
Today's filter is `locked_at IS NULL` everywhere; no `< NOW()` comparison
on lock columns anywhere in code.

What DOES exist today is the hard-vs-soft distinction encoded in
`locked_by` (see "Pattern A" section above). Hard locks are functionally
permanent. Soft locks are released by agents when they have legitimate
work that needs to clear them.

### Anticipated rhythm (paired with audit-pass-reset)

Doc 004 v4 designed timed expiry to pair with the audit-pass-counter
auto-reset that IS implemented. The intended rhythm:

```
Build → audit × 3 → cap reached → site quiet
  ... 60 days ...
  → pass counter resets, expired locks release
  → improvement loop runs fresh
  → finds new issues (content aged, design dated, new opportunities)
  → audit × 3 → quiet again
```

The pass-counter half is wired up in the improvement-loop workflow.
The lock-expiry half is not.

### Approved policy when timed expiry is implemented (2026-05-08)

When the lock-expiry project is taken up:

| Source | Default lock_type | Default expiry |
|---|---|---|
| `'admin'` (human edit via dashboard) | `permanent` | — |
| `'admin-removed'` | `permanent` | — |
| `'checkpoint'` (explicit human approval) | `permanent` | — |
| `'deploy'` (auto-lock-on-deploy) | `timed` | NOW() + 30 days |
| `'manual'` (asset upload) | `permanent` | — |
| Auditor approvals (visual-design, imagery-quality) | `timed` | NOW() + 90 days |
| `'audit-pending'` | not a lock — clear immediately on completion | — |

**Human-set locks remain permanent.** The conservative choice:
deliberate human edits should never silently expire. Only auto-locks
(`'deploy'`) and auditor approvals get timed expiry, which is enough to
restore the rhythm 004 v4 envisaged for the machine-generated parts of
the site.

The "lock proliferation" concern doesn't disappear under this policy —
human edits still accumulate permanent locks — but that's a deliberate
user-controlled choice rather than an inadvertent side effect. A
dashboard "show locks older than X days" view becomes important under
this policy because users need a way to find and review their old locks
when they want to revisit hand-crafted content. That UX is downstream
of the schema work.

### Sequencing

The lock-expiry project is a focused 1-2 day piece, scoped to:

1. Add `lock_type` and `lock_expires_at` columns to all four Pattern A
   tables in a single migration.
2. Update the auto-lock writers (mostly `'deploy'`-source writers) to
   set defaults per the policy table above.
3. Update lock-aware queries to use the expanded filter.
4. Extend `CheckComponentLock` to return classification and expiry.
5. Add `expired_review_locks` discovery check for the HITL hook.

Sequenced **after** the imagery loop work completes — see
LOCKS_should_locks_expire.md for the rationale and full implementation
sketch.

---

## What NOT to do

- **Don't introduce a third lock pattern.** The two current ones are enough. If new lockable data appears, fit it into Pattern A unless there's a strong reason not to.
- **Don't use Pattern B (`pinned boolean`) for new tables.** It exists because of migration history; new code should use `locked_at` + `locked_by`.
- **Don't merge locks across multiple tables in a downstream reader.** When a consumer asks "is this thing locked", the answer should come from one row in one table. The lock-transfer pattern above is how you make per-version locks behave like durable ones — without forcing every reader to chase history.
- **Don't put lock-aware logic in the workflow JSON.** Lock decisions are deterministic and should live in the Go action that writes the lockable row. Workflows pass data; they don't reason about lock state.
- **Don't override locks "just this once" from an automated path.** If an agent has a legitimate reason to update a locked row, it should fail loudly and surface that as a HITL review work item, not silently overwrite.

---

## Anti-patterns observed in the past

- **Override tables.** An earlier draft of doc 030 proposed a separate `site_plan_directive_overrides` table where HITL-approved values lived alongside per-version directive rows. Every reader (planner, brief renderer, page-build-handler, content writer) would have had to merge the two. Rejected — too many readers needing too much knowledge of the override mechanism. Replaced with Pattern A on `site_plan_directives` plus lock transfer in `write_site_plan` only.
- **Per-call lock checks in callers.** Some early code put `if locked_at == nil { ... } else { skip }` checks in workflow steps via conditional branches. The lock check should live inside the action that does the write, not as a workflow condition.
- **Locking JSONB blobs at field granularity.** Trying to lock individual JSON path inside a JSONB column (e.g. "lock `palette.primary` but not `palette.secondary`") is awkward in SQL and fragile in code. The right answer is to break the blob into rows so each lockable thing has its own row. This was one of the drivers for the per-directive row design in `site_plan_directives` (doc 030).

---

## Quick reference

**Adding a new lockable table** — use Pattern A:

```sql
ALTER TABLE <table>
  ADD COLUMN locked_at timestamptz,
  ADD COLUMN locked_by text;

CREATE INDEX idx_<table>_locked
  ON <table> (...)
  WHERE locked_at IS NOT NULL;
```

**Writing to a Pattern A row from an agent** — guard with `locked_at IS NULL` in the WHERE clause, or read first and skip in code.

**Reading lock state** — `SELECT locked_at, locked_by FROM <table> WHERE …`.

**Locks across plan rebuilds** — only the rewriting agent (`write_site_plan` for plans) handles transfer. Downstream readers see uniform locked/unlocked rows.

**Plan partials** — locks are on `site_plan_directives` (Pattern A). Older references in doc 029 to "the lock is on the plan partial (`pinned` already exists on `site_specs`)" are pre-Phase-1 and superseded — site_specs `pinned` is for *strategic* specs, not plan-time directives.
