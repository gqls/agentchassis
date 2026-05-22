# PLAN — Lock-Model Coherence

Date: 2026-05-19
Status: planned; step 1 (policy function) started in this session.
Reference model: `031_locks.md` "Tech debt: lock-model coherence (target model)".

## Goal

Collapse the accreted lock model into one a reader learns once: **one storage
pattern (Pattern A), one lifecycle column (`lock_type`), one expiry predicate,
one policy function.** Retire Pattern B (`site_specs.pinned`) and the implicit
hard/soft string-switch in `check_component_lock.go`.

This is tech debt, not a feature. It ships separately from adoption-faithfulness
(which runs on the current model + 053).

## Target model (three columns, no `lock_class`)

| column | meaning |
|---|---|
| `locked_by` | identity only — who/what set it |
| `lock_type` | `permanent` \| `timed` \| `review` |
| `lock_expires_at` | release time; NULL for permanent |

- Hard/soft = `lock_type = 'permanent'` (relies on the invariant **permanent ⟺ human-set**).
- Improvable predicate (one form everywhere):
  ```sql
  locked_at IS NULL
  OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL AND lock_expires_at < NOW())
  ```
- Source → (lock_type, expiry) decided once at write time by a single policy function.

An earlier draft proposed a fourth `lock_class` column. Dropped as redundant:
`locked_by` carries identity, `lock_type` carries lifecycle, and the policy
function carries the source→expiry mapping. A class column would duplicate
what those already encode.

## Grounded inventory (what actually exists, 2026-05-19)

### Pattern A tables (confirmed)
`sites`, `page_components`, `site_components`, `site_plan_directives` — each has
`locked_at` + `locked_by`. `assets` gains them via the imagery work; 053 adds
`lock_type` + `lock_expires_at` to all four of the directive/component/asset set.

### The hard/soft switch (one site)
`check_component_lock.go`:
- `ComponentLockStatus` struct (with `IsHard bool`)
- switch on `locked_by`: `case "admin","admin-removed","checkpoint": IsHard = true; default: false`

This is the only classification logic. Replacing it is step 3.

### `locked_at` predicate call sites — three distinct semantics, not one
The grep count of 11 conflates three different uses. They must NOT be
blanket-expanded:

**(a) Improvable-filter — the actual sweep (6 sites), all `pc.locked_at IS NULL`:**
- ~line 45256, 46354, 46454, 46634, 48086, 49318 (discovery/audit data-loading
  queries that skip locked page_components)
These get the canonical predicate so timed locks become improvable on expiry.

**(b) Locked-row finders — `locked_at IS NOT NULL` (leave as-is or invert carefully):**
- ~line 37893 (finds locked rows for some report/check)
- ~line 86488 (`transferDirectiveLocks` reads locked directives — already being
  extended in the adoption work to skip expired timed locks)
These are asking "which rows ARE locked", not "which are improvable". Expanding
them blindly would be wrong.

**(c) Single-row lock checks:**
- ~line 83767 (`SELECT locked_at IS NOT NULL FROM sites WHERE id = $1`) — a
  boolean "is this site locked" check; needs the expiry-aware form if sites ever
  get timed locks, but sites locks are human/permanent today, so low priority.

### Pattern B — likely already dead
`site_specs` writes (lines ~22450, ~32891, ~33443) use **supersede-then-insert**
(`UPDATE ... SET is_current=false` then `INSERT ... is_current=true`), with **no
`pinned` guard**. There are **zero `pinned` references** in the chassis code.

Doc 031 describes a `pinned`-guarded `ON CONFLICT` upsert, but the live code
doesn't do that. So either `pinned` exists as an unenforced column, or it was
never added. **This must be verified before step 5** (standing rule: check
schema before SQL):

```sql
\d site_specs
-- Does a `pinned` column exist? Any trigger referencing it?
SELECT tgname FROM pg_trigger WHERE tgrelid = 'site_specs'::regclass;
```

If `pinned` is absent or unenforced, Pattern B retirement is a doc correction
(+ optional column drop), not a reader migration. If something does enforce it
(a trigger, or a reader outside the chassis context dump), scope expands.

## Steps (ordered, with risk)

### Step 1 — policy function (LOW risk, additive) — STARTED
Add `platform/orchestration/locks/policy.go` (new package or existing helpers
file): `LockPolicyFor(lockedBy string) (lockType string, expiresAt *time.Time)`
encoding the approved policy table. Pure function, no DB, no callers forced to
change yet. Becomes the single source of truth that writers migrate to.

Provided this session as `lock_policy.go`.

### Step 2 — migrate lock-writers to the policy function (MEDIUM)
Find every site that sets `locked_at`/`locked_by` and have it also set
`lock_type`/`lock_expires_at` from `LockPolicyFor`. Writers today:
- `auto_lock_on_deploy` trigger (DB-side — may need a companion or a Go path)
- HITL UI lock-set paths (`admin`, `checkpoint`)
- auditor approval writes (`visual-design-auditor`, `imagery-quality-auditor`)
- adoption preserve-directive lock (already in the write_site_plan patch)
Enumerate precisely before editing; each writer is small but there are several.

### Step 3 — replace the hard/soft switch (LOW, post-053)
In `check_component_lock.go`: add `LockType` + `LockExpiresAt` to
`ComponentLockStatus`; SELECT them; set `IsHard = (lockType == "permanent")`;
delete the `locked_by` case set. Safe once 053's backfill has established
`permanent ⟺ human` (so the new derivation matches the old switch exactly).

### Step 4 — filter sweep (LOW, mechanical)
Replace the 6 `pc.locked_at IS NULL` improvable-filters (inventory (a) above)
with the canonical predicate. Leave (b) and (c) alone. One commit, easy to
review against the 6 known line areas.

### Step 5 — retire Pattern B (RISK GATED ON VERIFICATION)
Verify `site_specs.pinned` enforcement first (query above). Then:
- If dead: drop `pinned`, remove the Pattern B description from `031_locks.md`,
  done.
- If enforced somewhere: migrate those readers to a Pattern A check on
  `site_specs` (add `locked_at`/`locked_by`/`lock_type`), then drop `pinned`.
Either way the end state is one storage pattern.

### Step 6 — final doc pass
Update `031_locks.md`: remove Pattern B, fold hard/soft into `lock_type`,
collapse the two-pattern "Where locks live" table to one. The tech-debt section
becomes the implemented model.

## Sequencing and dependencies

- Steps 1, 3, 4 are independent and low-risk; can land in any order after 053.
- Step 2 is the broadest (touches several writers); do after 1 so they have a
  function to call.
- Step 5 is gated on the schema verification and is the only one that might
  touch readers; do last.
- None of this blocks adoption-faithfulness — that ships on 053 + the current
  shape and benefits from steps 3/4 later without depending on them.

## Verification per step

- Step 1: unit test `LockPolicyFor` for each source in the policy table.
- Step 2: after deploy, `SELECT locked_by, lock_type, lock_expires_at FROM
  page_components WHERE locked_at IS NOT NULL` shows types populated, not NULL.
- Step 3: a permanent lock returns `IsHard=true`; an expired timed lock is
  treated as unlocked. Compare against the old switch on a sample.
- Step 4: a component whose only lock is a `timed` one past expiry now appears
  in discovery results; a `permanent`-locked one still doesn't.
- Step 5: `\d site_specs` shows no `pinned`; spec writes still respect
  human-set authority via Pattern A.

## References
- `031_locks.md` — target model + tech-debt section
- `031_LOCKS_should_locks_expire.md` — approved policy table, conservative stance
- `FOCUS_adoption_faithfulness_via_locks.md` — first timed-lock consumer; policy table
- `053_lock_expiry.sql` — adds lock_type/lock_expires_at + backfill establishing permanent⟺human
- `check_component_lock.go` — the switch to retire (step 3)
- This conversation, 2026-05-19
