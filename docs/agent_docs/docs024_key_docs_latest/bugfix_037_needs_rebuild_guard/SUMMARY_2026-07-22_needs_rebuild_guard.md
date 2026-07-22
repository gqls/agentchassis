# SUMMARY — bug 037 closed + explicit redesign intent built (2026-07-22)

**What we're trying to do.** Two things, now both done: stop a re-plan silently discarding a
`needs_rebuild` page's built layout (bug 037), and give back a *clean, deliberate* way to redesign one
specific page (the follow-on feature 012).

**Where we've come from.** Yesterday we established, from the code, that `needs_rebuild` always means
"re-render as planned" (never "recompose"), fixed the silent-loss defect, and shipped it live on
v1.0.1146 — leaving two questions for the owner: close on that evidence, and whether to build an
explicit redesign signal.

**What we've done.** The owner ruled: close 037 on the current evidence, and build the redesign
signal. Both done. 037 moved to `/bugs_closed/`. The redesign signal is `recompose_pages` — a list of
page names on the re-plan request; named pages are released from the preserve guard so the LLM
redesigns them, everything else is protected. It needed no plumbing changes (the re-plan request
already carries its spec to the planner), so it is a small, self-contained, well-tested planner change.

**Where we are now.** Bug 037: **closed, fixed, live.** Feature 012: **built and committed, inert until
the next chassis image roll.** No open defects in this workstream. The redesign feature works from a
one-line SQL trigger today; a friendlier way to fire it is optional.

**Where we're going.** When the chassis next rolls, verify 012 live (re-plan a site with
`recompose_pages` set and watch one page redesign while its peers hold). One open design choice parked
for the owner: whether a recomposed page the LLM omits should be *dropped* (current behaviour) or
*kept*. Otherwise this workstream is complete.
