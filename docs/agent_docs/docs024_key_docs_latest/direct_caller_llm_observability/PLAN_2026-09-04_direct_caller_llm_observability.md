# PLAN 2026-09-04 — direct-caller LLM observability (bugs_open/480)

**Status: NOT STARTED.** This file exists so that the first session has a place to record decisions
rather than a blank page, per the standing-five rule. Nothing below is decided.

## The problem, stated once

`llm_call_log` is written from `platform/orchestration/actions` and nowhere else
[MEASURED 2026-09-04]. Six model call sites live outside that package. They are therefore absent
from every truncation query, every token-pressure check, and every "which steps are running out of
room" answer the estate can give — and an absence in that table is indistinguishable from a step
that never ran.

## Open questions, in the order they need answering

1. **Can the existing logger be reached, or must a seam be built?**
   `platform/orchestration/actions/llm_call_logger.go` already does the job. `internal/agents/*` and
   `internal/tools-api` would have to import it. The known hazard is the layering that 257 hit:
   `platform/orchestration/actions` imports `platform/aiservice`, so anything pushed DOWN into
   `aiservice` is a genuine Go import cycle. Read `platform/aiservice/max_tokens.go`'s comment on
   this before designing — it costs five minutes and it is the same wall.

2. **Does tools-api even have a database handle on the path that would log?**
   It is a separate service. If it does not, this is not "call the logger" but "choose a transport",
   which is a much larger change and may argue for a different answer for that service than for
   `internal/agents/*`.

3. **What must a row be able to DISTINGUISH?** This is the design constraint, not a detail. A logged
   number that a hardcoded default could equally have produced tells you nothing (257 §2c: a Go
   literal of 2000 beside a configured 2000, and no query could separate them). At minimum a row
   should record where the budget came from, not only what it was — the ladder already computes that
   (`actions.ResolveStepBudget` returns the level name).

4. **Does the budget guard extend, or do we state that it will not?**
   `llm_budget_call_sites_test.go` binds "no hardcoded budget" at the package, and a Go test is
   package-scoped by nature. Options: one test per package; a `scripts/pattern-check.py` check
   (in council scope since 2026-08-23); or a stated decision that these six sites are exempt and why.

## What is deliberately out of scope

Budget RESOLUTION. Fixed in 257 round 3 — the ladder is live in code (`d88afbf84`) and the
configuration migrated (`769`, applied 2026-09-04). This lane is about whether a call that has
already chosen its budget is visible.

## First actions

Everything in `bugs_open/480` §4. The first of them is re-running the census rather than quoting it:
`git log --since=2026-09-04 --diff-filter=A -- internal/ platform/` tells you whether §2's table is
still complete, and a census that goes stale goes stale by ADDITION, silently.
