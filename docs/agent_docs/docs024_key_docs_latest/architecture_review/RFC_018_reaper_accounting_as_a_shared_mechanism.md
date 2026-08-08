# RFC 018 — reaper accounting as a shared mechanism

**Status:** IMPLEMENTED for its first consumer (owner go-ahead 2026-08-08); OPEN as
an invitation to the remaining reapers.
**Raised by:** the council's `architecture` seat on the bugs_open/205 fix
(2026-08-07, corr `2db88f8f…`): "the second reaper (possibly third, counting
thunder-reaper) to get its own hand-written retry_count/backoff/park SQL rather
than a shared, reusable reaper-accounting mechanism. Not a reason to block this
fix, but the class is recurring and each recurrence is currently solved locally."
**Owner decisions carried:** 2026-08-08 #3 ("each task type can declare its own
ceiling") and #4 ("go ahead and implement").

## The defect class

A reaper that rescues stale work without REMEMBERING its rescues converts any
deterministic failure into an infinite loop. This estate has now paid for that
lesson twice: `stale-work-item-reaper` (016b §9, 2026-07-25 — row-age keyed,
re-annotating its own work) and `stale-orchestration-reaper`'s collection-tasks
arm (bugs_open/205 — 1,575 wasted dispatches/day). Each time, the remedy was the
same three numbers (how stale before rescue, how many rescues before parking, how
long to back off) hand-written into that reaper's SQL.

## The mechanism (live since 2026-08-08, migration `sql_for_agents/335`)

1. **`reaper_policies`** — one row per (queue, item_type): `park_after`,
   `backoff_minutes`, `stale_after_minutes`, notes. A task type DECLARES its
   ceiling by inserting a row; an undeclared type gets the consuming function's
   documented defaults (5 / 20m / 20m). This answers the guardian seat's 205
   objection — a future task type no longer inherits a hidden literal; the choice
   has a place to live and a reviewer can read it.
2. **`business_intel.reap_stale_collection_tasks()`** — the accounting logic,
   once: find stale in_progress claims, increment `retry_count`, back off
   `scheduled_for` linearly, park as `'failed'` (message naming the policy and
   bug) at the ceiling. Returns `(reset_count, parked_count)` so the calling
   pre_query's note distinguishes rescue from parking.
3. The `stale-orchestration-reaper` pre_query's `reset_tasks` CTE is now
   `SELECT * FROM business_intel.reap_stale_collection_tasks()` — the reaper
   fires on its schedule; the logic lives in one reviewable place.

**Proven:** induced test 2026-08-08 (transaction, rolled back) — undeclared type
parked on the 5th reset at the defaults; a declared `park_after=2` policy honoured
with backoff stamped; zero residue. Behaviour for the existing population is
byte-identical to the 2026-08-07 inline CTE it replaces.

## The contract, and the deliberate stopping point

A queue opts in by having: a status column with claimable/claimed/terminal values,
a claim timestamp, `retry_count`, `scheduled_for` (or equivalent re-eligibility
gate its claimer honours), `error_message`, and a type column. The function is
NOT generalised over arbitrary tables yet — the remaining reapers differ in
column vocabulary and semantics (`site_work_items` has evidence-gated completion;
thunder reconciles against a vendor API), and a dynamic-SQL generalisation
written against ONE consumer would be speculation. **The rule this estate keeps
re-learning is that a mechanism with no live caller rots unexercised.** The
stopping point is explicit: when a SECOND queue adopts the contract, generalise
the body (format('%I') dynamic SQL or one function per queue reading the shared
policy table — decide then, with two real consumers to test against).

## Migration invitation (per-reaper, no urgency)

- `stale-work-item-reaper` / `claimed-item-timeout` (site_work_items): closest
  fit; already has `attempt_count` — adopt `reaper_policies` for its numbers
  first, executor second.
- `thunder-reaper`: policy-table fit only (its "park" is a vendor decommission).
- `stuck-task-reaper`: unexamined here; census before assuming shape.

## Review route

Config/SQL only — the council gate refuses config-only submissions client-side,
so per the RFC_006 precedent this RFC + the migration file + the register entry
(SCH-014 / the new mechanism entry) + the induced-test transcript in the
bugfix_205 NOTES are the reviewable artifacts. The architecture seat's own
objection is the provenance; the owner's 2026-08-08 go-ahead is the authority.
