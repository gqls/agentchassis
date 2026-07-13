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

Used by: `sites`, `page_components`, `site_components`, and (Phase 1) `site_plan_directives`.

This is the dominant pattern. Use it whenever the lockable thing is a row that itself represents a discrete unit of work — a component instance, a directive, a site record.

### Pattern B — `pinned boolean` (per-row flag)

One column on the row:

```
pinned     boolean default false
```

Intended for: `site_specs` only. A pinned spec row was meant to be treated as
authoritative — the next write of the same `(site_id, aspect)` skipping the
upsert when the existing row is pinned.

**Verification (2026-05-19): Pattern B is unenforced in the current code — treat
it as dead.** A full scan found no `pinned` reads or writes anywhere in the
chassis (the only `pinned` token is an unrelated image-tag comment), and every
`site_specs` write uses supersede-then-insert (`UPDATE … is_current=false` then
plain `INSERT`) with **no `pinned` guard and no `ON CONFLICT … WHERE pinned`**.
The guarded-upsert this section originally described was never implemented.
Whether the column physically exists in the DB is the only open point — confirm
with `\d site_specs`. Retiring Pattern B (PLAN_lock_coherence.md Step 5) is
therefore reader-free: almost certainly a doc correction plus an optional column
drop. **New tables must not use Pattern B.** Use Pattern A.

Note: "pinning a site" in the sense of *reverting a whole site to a known-good
state* is a different, live capability — the site-snapshot system, not this
column. See "Locks vs snapshots" below.

---

## Where locks live (current state)

| Table | Pattern | What gets locked | Set by |
|---|---|---|---|
| `sites` | A | Identity fields (logo_url, contact_address, company_name) when human-set | HITL UI (planned) |
| `page_components` | A | A specific section instance on a specific page (e.g. the "approved" hero copy) | Improvement-loop discovery checks; HITL UI |
| `site_components` | A | A site-wide component (e.g. the header used everywhere) | HITL UI |
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

## Locks vs snapshots — orthogonal durability mechanisms

Locks are one of three durability concepts in the platform, and they are easy to
conflate. They are not the same thing and operate at different scopes.

| concept | scope | mechanism | what it does |
|---|---|---|---|
| **Pattern A locks** | one row | `locked_at` / `locked_by` (+ `lock_type` / `lock_expires_at` when timed expiry lands) | prevent an agent overwriting a value on the next write |
| **Pattern B `pinned`** | a `site_specs` row | `pinned boolean` | *dead — see note below; never enforced in code* |
| **Site snapshots** | the whole site | `site_snapshots` JSONB rows + revert function | capture a known-good past state and restore the entire site to it |

**Locks and snapshots are orthogonal.** A lock stops the *next write* to a row.
A snapshot lets you *undo writes that already happened* by restoring a captured
state. One is forward-looking prevention; the other is backward-looking restore.
A site can have locked rows and also have snapshots; neither implies the other.

### The site-snapshot system (the "revert a whole site to known history" capability)

Implemented in `actions/site_snapshot_actions.go` + SQL functions:

- `take_site_snapshot(site_id, trigger, git_sha, …, label)` — captures
  `site_specs`, `pages`, `page_components`, `navigation`, and `site_components`
  into a single self-contained JSONB snapshot row in `site_snapshots`.
- `revert_site_to_snapshot(snapshot_id, reverted_by)` — takes a safety snapshot
  first (`trigger='pre_revert'`), then **wholesale-replaces** `site_specs`,
  `pages`, `page_components`, `navigation`, `site_components`, and updates the
  site record.
- `list_site_snapshots(site_id)` — lists available snapshots.
- Triggers: `deploy`, `manual`, `pre_edit`, `scheduled`.

This is the mechanism for "roll the whole site back to a known-good point." It is
**not** `pinned` and not a lock — historically the two have been mentally
grouped, but they're separate.

### Open question — does revert respect locks? (resolve before relying on either)

The revert is documented as a **wholesale replace** of `page_components` (and the
other captured tables). The lock-awareness lives in the SQL functions, which
aren't in the Go source — so two things are currently **unverified**:

1. Does `take_site_snapshot` capture the lock columns (`locked_at`, `locked_by`,
   and future `lock_type` / `lock_expires_at`) alongside content? If it captures
   `to_jsonb(pc.*)` it does; if it selects a content-only column list, lock state
   is lost on capture.
2. Does `revert_site_to_snapshot` preserve a currently human-locked component, or
   does the wholesale replace clobber a human edit made after the snapshot was
   taken?

If revert ignores locks, a restore could silently wipe a human-locked edit — a
real correctness gap in the durability model. Resolve by reading the functions:

```sql
\sf take_site_snapshot
\sf revert_site_to_snapshot
```

The coherent end-state should make the answer explicit: a revert should either
(a) preserve rows whose current lock post-dates the snapshot, or (b) be
documented as a deliberate "revert overrides locks" operation that itself
requires a human action (it already takes a `pre_revert` safety snapshot, so the
clobber is undoable). Either is defensible; leaving it unspecified is the
problem. Tracked in `PLAN_lock_coherence.md`.

---

## Lock lifecycle (timed and review locks)

Per doc 007 (adoption pipeline) the platform anticipates richer lock types beyond "permanent until manually unlocked":

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

And the read condition for "is this row locked?" expands from `locked_at IS NULL` to `(locked_at IS NULL OR lock_expires_at < NOW())`. Discovery checks can then auto-create `needs_lock_review` work items for review-type locks at expiry.

This is anticipated, not required. **Update (2026-05-19):** the `lock_type` /
`lock_expires_at` columns are being added across all four Pattern A tables
(migration 053), and the first real consumer of timed expiry is adoption
faithfulness (`locked_by='adoption'`, `lock_type='timed'`, 90-day) — see
`FOCUS_adoption_faithfulness_via_locks.md`. So locks are no longer all
effectively `permanent`: human-set locks remain permanent, while `deploy`
auto-locks, auditor approvals, and adoption locks are `timed`. The canonical
"is this row improvable?" predicate is now:

```sql
(locked_at IS NULL
 OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))
```

See the tech-debt section at the end of this doc for the target coherent model.

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

---

## Tech debt: lock-model coherence (target model)

**Status (2026-05-19): the lock model has accreted and is harder to understand
and manage than it should be. This section records the debt and the target
shape so the cleanup, when taken up, has a single reference.**

### What's incoherent today

The rules a reader must assemble live across four documents and several code
sites, and the model carries four overlapping concepts that aren't expressed
uniformly:

1. **Two storage patterns.** Pattern A (`locked_at` + `locked_by`) on four
   tables, Pattern B (`pinned boolean`) on `site_specs`. B exists only for
   migration-history reasons and is already "don't use for new tables" — but
   it's still live, so every reader of `site_specs` needs different logic.

2. **Hard vs soft is implicit in a string.** Whether a lock is human-permanent
   or agent-soft is derived by matching `locked_by` against a hardcoded set
   (`admin`/`admin-removed`/`checkpoint`) in `check_component_lock.go`. The
   classification isn't a column; it's a switch statement. Add a new human
   source and you must remember to edit that switch or it silently becomes
   "soft."

3. **Timed expiry is half-built.** `lock_type` + `lock_expires_at` are designed
   (004 v4, this doc) and now being added (053), but the hard/soft string-match
   and the new `lock_type` enum encode overlapping intent: `permanent` largely
   means "hard", `timed`/`review` largely mean "soft". Two mechanisms for one
   distinction.

4. **Source, type, scope, and expiry are entangled.** `locked_by` currently
   carries identity (who), an implicit hard/soft class (what kind), and — by
   the policy table — an implied default expiry. One string, three jobs.

### Target model (when the cleanup is taken up)

Three orthogonal columns on Pattern A everywhere — no more, no fewer. (An
earlier draft proposed a fourth `lock_class` column; it's redundant. Given the
policy invariant **permanent ⟺ human**, "is this a hard/human lock?" is just
`lock_type = 'permanent'`. `locked_by` already carries identity, and the
source→expiry mapping only matters at write time. A separate class column would
duplicate what `locked_by` + the policy table already encode.)

| column | meaning | values |
|---|---|---|
| `locked_by` | identity only — who/what set it | user id, agent name, `adoption`, `deploy`, auditor names, … |
| `lock_type` | lifecycle | `permanent` \| `timed` \| `review` |
| `lock_expires_at` | when a timed/review lock releases | timestamptz, NULL for permanent |

Rules become uniform and table-driven, not switch-statement-derived:

- **Hard/soft** stops being a `locked_by` string-match; it's `lock_type = 'permanent'`.
- **"Is this row improvable?"** — one predicate everywhere:
  ```sql
  locked_at IS NULL
  OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW())
  ```
- **`review`** locks don't auto-release; on expiry a discovery check raises a
  `needs_lock_review` HITL item.
- **Source → (lock_type, expiry)** is decided once, at lock-set time, by a single
  policy function (the table in `FOCUS_adoption_faithfulness_via_locks.md`).
  Writers consult it; readers never re-derive.

**Invariant to preserve:** `permanent ⟺ human-set`. The policy gives every
human source (`admin`, `admin-removed`, `checkpoint`, `manual`) `permanent`, and
gives only those sources `permanent`. As long as that holds, `lock_type =
'permanent'` is an exact stand-in for the old hard/soft switch. If a future
requirement ever needs a human *timed* lock or an agent *permanent* lock, that's
the moment to revisit — likely via the `review` type rather than a new column.

The single source of truth for "what type/expiry does a given `locked_by` get"
is one policy function, applied at lock-set time. This replaces the implicit
mapping currently split between the `check_component_lock.go` switch and the
prose policy table.

### Migration path (sketch — see PLAN_lock_coherence.md for the grounded version)

1. Land 053 (`lock_type` + `lock_expires_at` on the four Pattern A tables) —
   done as part of adoption work. Its backfill establishes `permanent ⟺ human`.
2. Centralise the policy as a single Go function `LockPolicyFor(lockedBy)
   → (lockType, expiry)`; migrate lock-writers to call it.
3. Replace the `check_component_lock.go` `locked_by` switch with
   `IsHard = (lockType == "permanent")`; also return `LockType` + `LockExpiresAt`.
   Delete the hardcoded source set.
4. Filter sweep: the 6 `pc.locked_at IS NULL` improvable-filters → the canonical
   predicate above. (The `IS NOT NULL` locked-row finders and the lock-transfer
   query are different semantics — handled separately, not blanket-expanded.)
5. Retire Pattern B: verify `site_specs.pinned` is genuinely unenforced (the
   current writes use supersede-then-insert with no `pinned` guard), then drop
   the column and fold `site_specs` onto Pattern A. Scope depends on the
   verification — likely a dead-column drop, not a reader migration.
6. Centralise any remaining ad-hoc lock-writes onto the policy function from (2)
   so source→(type, expiry) lives in exactly one place.

End state: one storage pattern (A), one lifecycle column (`lock_type`), one
expiry predicate, one policy function. Pattern B gone. The `check_component_lock`
string-switch gone. A reader learns the model once and it applies everywhere.

### Why not now

This is debt, not a blocker. Adoption-faithfulness and the timed-expiry project
ship on the current model (053 + the conservative policy). The coherence
cleanup is a separate, larger change touching every lock reader and retiring
Pattern B — worth doing deliberately, not folded into feature work. Recorded
here so it isn't rediscovered from scratch.
