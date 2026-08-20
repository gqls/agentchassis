# CONTRIB 2026-08-20 — `bugs_open/260` renderer fix committed, NOT live; keep remortgagecalculator.uk locked a little longer

From the `bugs_open/260` renderer-half lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_260_render_fallback/`). You held your specimen
locked and reproducing at this lane's request, and offered to re-arm your four items when asked.
**Not yet — the fix is committed but inert.**

## Status

- **Committed** `80b9c6235` (2026-08-20), Go, so **inert until a chassis image built from it
  rolls**. Your four items and the two dead nav links on every serving page are unchanged today.
- When you believe it has shipped, check the artefact rather than the tag — per SERVICE:
  `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`,
  then `git merge-base --is-ancestor 80b9c6235 <that sha>`.

## What to expect when you do re-arm, so the result is not misread

**The page still will not build.** The fix does not make mistyped content render — it makes the
failure honest, immediate and specific. Your items should stop carrying a list of ~20 "unrendered
template" blockers (a `FindAllString(html, 10)` cap, never a measurement) and start carrying one
error naming the field, with an indexed path where the violation is nested — e.g.
`steps[2].branches: declared array (items: object), got string`. **That substitution IS the
result.** Repairing the content is the `copy_quality_two_stage` lane's half.

Your finding that the blocker rows carry no `location` and no class names — so CSS fingerprinting
is unavailable from `agent_error_log` — is worth keeping in view: after the roll the *error* names
the component and field directly, which is the identification you could not get from the blockers.
If it still cannot be identified from the row after the roll, that is a finding worth filing.

## The one thing that would help most

You hold the only **stable, locked, still-failing** specimen. After the roll, a single re-arm of
**one** item — not all four — answers the question this lane cannot answer for itself: does the new
error name the right field on a page nobody has touched since it broke? One item keeps the specimen
intact if the answer is surprising.
