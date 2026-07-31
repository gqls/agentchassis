# PLAN — bugs_open/150: the improvement loop reports "site is clean" after promoting findings

**Started** 2026-07-31 by session "bugfix 27" (`3bec7dd7`). **Bug:** `bugs_open/150`.
**Ownership when picked up:** `who-owns.py 150` → no owning workstream; the only live
session with the number in its transcript (`f3d9debb`) was writing the 016b §10 index row,
not fixing it. Re-checked the live `.jsonl` transcripts for the CODE symbols
(`triage_detected_items`, `improvement-loop`, `complete_clean`) as well as the number,
because every ownership check is lagging.

---

## The defect, in one paragraph

`triage_detected_items` is a step in **three** live agents — `improvement-loop`
(`triage_findings`), `design-audit-agent` (`triage`), `site-review-agent` (`triage`) — and
its promotion is unconditional over the site (`WHERE site_id = $1 AND status = 'detected'`,
no type filter). The improvement loop calls both children **before** running its own copy,
so the first copy to run takes every row and the parent's copy honestly returns
`promoted: 0`. The parent then branches on `triage_result.has_items == true` — its own
copy's output — takes `else_step`, and terminates at `complete_clean`, whose success
message is *"No issues found — site is clean"*, skipping `insert_rerender_item` →
`spawn_dispatch` → `call_dispatch`.

## Decisions, and why

**D1 — Add a site-scoped signal; do NOT redefine `has_items`.** `has_items` is a
fleet-wide convention across actions ("my own result set was non-empty"). Three other live
conditions read it, each about its own loader, correctly:
`build-dispatch-loop.check_has_items`, `site-work-orchestrator.check_has_items` and
`.check_has_fix_items`. Redefining it here would repair one branch by making a shared word
mean two things in two places. So the action gains `site_dispatchable` /
`site_dispatchable_count` beside it.

**D2 — Do not take the bug file's own first-ranked candidate.** It ranks "one triage, one
owner" (remove the step from the two children) first. Rejected as the first move: it
requires auditing every other parent of those two agents, and it leaves the identical
defect available to the next agent that gains a triage step. A site-scoped signal makes the
ordering **irrelevant** instead of making one ordering **mandatory** — the bad state stops
being representable rather than stopping being reachable by today's callers. Both children
keep their triage steps.

**D3 — The count predicate is deliberately narrower than the dispatcher's.**
`LoadWorkItemsAction` also filters `attempt_count`, `approval_mode` and `depends_on`. Those
answer *"will the loader return this row on its next tick"*; the branch asks *"is there
unfinished promoted work here"*. Anything they exclude — attempt-exhausted, awaiting
approval, dependency unlanded — is still work, so the narrower predicate can only err
toward NOT clean. For a branch whose false side ends the run, that is the safe direction.

**D4 — A failed count reports `site_dispatchable = true`.** "We could not answer" must
never render as "no work". A needless closing rerender costs a render; a false clean costs
the findings. The reason is carried in the result (`site_dispatchable_error`) and the count
is the `-1` sentinel, so a fail-safe is never mistaken for a measurement.

**D5 — Config half is written and NOT applied.** `sql_for_agents/281_…sql`. On a chassis
that predates the Go half the new field resolves to nil → the condition is false → **every**
run takes the clean branch, including the ones that promote. That is strictly worse than
the bug. Ordering is pinned by a unit test as well as by the migration's banner.

**D6 — This session does not roll the image** (owner's choice, asked before starting). So
the fix is committed and inert, and **150 stays in `bugs_open/`**: the bar is fixed AND
live. What is owed is written into the bug file rather than assumed.

## Correction to the originating brief

The bug file's §Confidence says: *"The escape hatch would be a step that CREATES detected
items between the last child triage and the parent's — `site-review-agent.write_strategic_findings`
is the only candidate, and in this run it created no work items."* The control run of
2026-07-31 shows that hatch **opening and not helping**: site-review promoted **3** items
of its own, i.e. it created findings and then triaged them itself — so the parent still saw
0. The hatch is closed by the child's own triage step, which makes the defect more robust
than the file allowed for, not less.

## Scope kept out, deliberately

`check_audit_pass_limit` routes straight to `complete_clean` when
`get_audit_pass_count(site) >= 3`, so a capped site is told "site is clean" when what
happened is "we skipped auditing" — and its `detected` pile is never promoted at all. Same
false claim, second route. **`[MEASURED 2026-07-31]` 0 of 25 sites are at the limit**, so it
is latent. It needs an honest terminal step rather than this condition; recorded in the bug
file and in WDS-015 as an open question, not fixed here.
