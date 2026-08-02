# 029 FEATURE — 30 of 33 fleet-wide criteria fences predate the `has_visible_area` fix and don't use it

**Raised:** 2026-08-02 by the `staged_component_build` lane, while re-checking the PLAN's
own phasing after P1+P2 closed. Not a bug — `has_visible_area` works correctly since
`bugs_closed/157` (live from `v1.0.1216`); this is a coverage gap in what already-published
fences assert, now that a check type they couldn't safely use is fixed.

**Status:** FILED, unowned, not designed. Named in `staged_component_build`'s own handoffs
since 2026-07-31 as owed but explicitly not this lane's to duplicate — most of the affected
fences belong to other lanes' subjects. This entry exists so the gap is visible instead of
living only in a handoff's "behind this one" list.

## The gap, measured

Queried every current fence-carrying PLAN (`doc_plans WHERE is_current AND body ~
'```criteria'`), 2026-08-02:

| | count |
|---|---|
| fences fleet-wide | 33 |
| fences already using `has_visible_area` | 3 |
| fences with NO `has_visible_area` check | **30** |

The 3 that do (`tool-ai-vendor-trust-checklist`, `tool-review-council-simulator`,
`teaser-reveal-panel`) were all authored 2026-07-30 or later — i.e. after the fix landed.
Every fence authored before that date (the other 30, `tool-drop-rate-tuner` through
`tool-cma-obligation-checker` by `updated_at`) was written while the check type was known
broken (`bugs_open/157`, before it closed) or simply predates its existence, and nobody has
gone back to add it now that it's safe to use.

## Why this matters, and why it's still not urgent

`has_visible_area` catches a defect class TL-034 was built for and no other check type
covers: an element present in the DOM and passing every other assertion, but rendered at
zero visible size (the exact `bugs_closed/157` symptom — a 24×24 checkbox measuring 0×0 was
the bug's own reproducer, not a hypothetical). A fence without it can pass 100% while a
control the user actually needs to see or click is invisible. That said: nothing currently
demonstrates this defect class is LIVE and undetected in any of the 30 — this entry reports
absence of coverage, not a confirmed miss. Costed as a backlog item, not escalated as a bug.

## Ownership — deliberately not resolved here

The 30 fences span many lanes' own subjects (`gauntlet-round-record`, `vonc-spark-game`,
`tool-arena-interface`, `tool-cma-obligation-checker`, ...), most with no connection to
`staged_component_build`. Editing another lane's fence unilaterally is not this lane's call,
and `who-owns.py` doesn't cover `features_open/` or fence provenance — there is no single
owner to route this to. Filing it, rather than fixing it, is the deliberate choice here.

## Fix candidates, ordered by what closes the door vs. what merely nags

1. **A discovery check**, same shape as `features_open/028`'s candidate 2: a query (or a
   discovery-handler agent, per the existing 22-registered-handlers pattern —
   `bugs_open/149` is the standing caution against adding a 23rd nobody runs) that reports
   `doc_plans` fences with a `visible-area`-shaped assertion opportunity (any check whose
   `selector` targets an interactive or informational element) and no `has_visible_area`
   check present. Cheapest, and the only candidate that also tells you WHICH fences, not
   just the count.
2. **Backfill by owning lane, on their own schedule.** Once (1) exists, route its findings
   to each fence's actual owner (via commit history / workstream-doc grep, the same method
   `who-owns.py` uses for bugs) rather than one thread touching 30 fences it doesn't
   otherwise work in.
3. **A gate that requires `has_visible_area` on new fences going forward** — cheapest to
   assert, does nothing for the 30 that already exist, and risks the same "gate proliferates
   into dead config" failure mode `bugs_open/149` already measured once (D8's own reasoning
   for cutting this lane's ladder to three funded gates applies here too: don't build a
   fourth check kind on top of three that are already proven, without evidence it's needed).

**(1) first.** It is the only step that costs less than the problem and produces a number
nobody currently has.

## How to verify a fix

Re-run the query in "The gap, measured" above — the denominator (33, or whatever it has
grown to) and the `has_visible_area` count should converge, or candidate (1)'s discovery
check should report zero outstanding for any fence it judges eligible.

## Cross-links

`bugs_closed/157` (the fix this backlog follows from); TL-034 (`has_visible_area`, concept
register); `features_open/028` (the precedent this entry's shape and candidate-ordering
follows, same lane, same week); `bugs_open/149` (the "gates multiply into dead config"
caution behind candidate 3's ranking); `staged_component_build/PLAN_2026-07-30_staged_component_build.md`
(where this was named "owed" but deferred, 2026-07-31 onward).
