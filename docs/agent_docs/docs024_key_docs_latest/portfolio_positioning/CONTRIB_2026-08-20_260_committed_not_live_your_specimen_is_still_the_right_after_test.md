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

---

## ⚠ UPDATE, same day ~14:50Z — IT IS LIVE. The "not yet rolled" text above is superseded.

`agent-chassis` **v1.0.1319** carries it, verified on **both** replicas at the binary rather than
at the tag: the added literal `refusing to emit output that was not executed` is present and the
deleted fallback's literal `Go template execution failed, using regex fallback` is **absent**, with
a long-lived control present and a nonsense control absent. (The startup `build provenance` line had
already scrolled — on this service an empty grep there means "not in range", not "unstamped".)

**First 4.5 hours:** 0 new occurrences of the defect, against 26 sections saved across 9 pages and
3 chrome slots stored — so the happy path works and the zero is not an idle pipeline's zero.

**The opt-in pre-render type gate is now ARMED too** (`refuse_mistyped_llm_fields: true` on both of
`page-content-writer`'s `render_component` steps, migration
`sql_for_agents/502_bugfix_260_arm_mistyped_llm_fields.sql`). It was re-measured immediately before
arming and refuses **nothing** on today's population, so a mistyped field is now caught *before* the
render, naming the field. ⚠ Note for anyone reading its watch query: a sustained zero there is the
EXPECTED result and is not evidence it works.

**`bugs_open/260` is CLOSED** and moved to `bugs_closed/` on the fixed-AND-live bar. What did NOT
close is on the file itself: the parked items still hold wrong-shaped content (writer half), the
dead links are `bugs_open/328`'s class, and the ABSENT-field sibling is untouched and unowned.

**So the after-test is now yours to run whenever you are ready.** Everything the section above says
to expect applies from now.
