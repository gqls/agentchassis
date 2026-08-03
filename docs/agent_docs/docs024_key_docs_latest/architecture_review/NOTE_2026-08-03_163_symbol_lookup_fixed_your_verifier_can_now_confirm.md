# To the architecture_review lane: RFC_005 §3.2's verifier can now confirm a symbol — and the half that stays yours

**From the `bugfix_163_symbol_lookup` lane, 2026-08-03.** Written because the 2026-07-29
owner ruling says a shared mechanism's consumers are *told*, not merely measured — and your
lane's newest continue-here (`HANDOFF_2026-07-28c`) predates RFC_005's ratification, so
nothing in your own chain would surface this.

## What changed under your mechanism

`bugs_open/163` (now closing): the `diagnose_code_lookup` symbol arm token-AND-ed the whole
query against the `symbol` column, so the `path:Symbol` form your `derive_checks` prompt
**defines** was unsatisfiable by construction. Every symbol check your verifier ever ran
returned 0 rows; 20 verdicts to date, 16 `NEEDS_HUMAN_REVIEW`, **0 `CONFIRMED`**.

Fixed in Go, commit `c3b02f035` (council corr `da1d3a40-8ec1-4ea1-9c65-8c585ec2d013`): path
half → `path` column, name half → token-AND on `symbol` (receiver forms preserved), a
`:LINE` suffix degrades to a path check, a path-qualified miss reports where the bare name
*does* resolve, and every 0-row answer names the predicate it executed. **Your prompt was
right; the executor now honours it. Nothing in your agent definition needs to change for
this.** Inert until the next chassis image rolls.

## What remains yours, deliberately not done in this pass

163's fix candidate 2 — *"make the verifier unable to invent a cause it has not checked"* —
is a change to your `verify` step's prompt (`sql_for_agents/276_landmine_verifier.sql`, live
`agent_definitions` row). I did not touch it: it is your seam, per the owner ruling. Two
facts that shrink the job:

1. The false "index predates the commit" cause was a **generalisation of the run-level
   staleness banner** onto per-check results — not a fabrication. The banner is correct and
   should stay.
2. The per-check `-- searched: <predicate>` line the fix adds gives the verify model, for
   the first time, a per-check fact that can override that banner. The remaining prompt
   change is therefore one sentence of instruction (a 0-row check may only be attributed to
   staleness if the symbol's file postdates the indexed `commit_sha` the banner names), not
   a redesign.

## One check worth adding to your next continue-here

After the next chassis roll, re-fire the verifier at any paren-form entry
(`./scripts/trigger-landmine-verifier.sh '<source>'` — *not* via `landmines-sync.py --apply`
first, which consumes the dispatch signal) and confirm a `CONFIRMED`/`STILL_VALID` verdict
that cites a resolved symbol. That closes DOC-069's falsified "CONFIRMED end to end" status
with an observation instead of a claim.
