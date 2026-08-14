# SUMMARY 2026-08-14b — CTA buttons fleet fix (bug 268): proven, repaired, closed

**What we're trying to do.** Stop the platform deleting call-to-action button
links when it rewrites a page's copy, and put back the links it had already
deleted. A button whose link is gone renders as nothing at all — no error,
no gap, just a page quietly missing its main conversion control.

**Where we've come from.** The bug was filed two days ago as "214 missing
buttons across 19 sites". Yesterday we established the mechanism: a
fast-path in the page planner, older than the safety net that protects
every other kind of field, skipped that net for exactly the fields that
hold button destinations. This morning the one-line fix was approved by the
review council first time and went out in the day's release. Checking
history also reshaped the problem: only ten of the missing buttons had ever
HAD links — the rest never got a destination assigned in the first place,
which is a different and older problem.

**What we've done since.** Three things, each verified on the running
system rather than by status flags. First, the proof: we ran a real copy
rewrite on a live darts-site page — the exact operation that used to delete
links — and every link survived; the build's own record shows the new
safety net is what carried them through. Second, the repair: all ten
genuinely-deleted links were restored from the history archive (each target
checked live first) and the seven affected pages re-rendered and verified
on the live sites. Third, the permanence check: we rewrote one of the
freshly repaired pages a second time, and the restored links survived that
too — the fix and the repair hold together. We also delivered the council's
two advisory follow-ups (a complete map of every field-type's path through
the planner, now in the concept register, and a new test for the
neighbouring path), and the bug file has moved to the closed pile.

**Where we are now.** The bug is closed: fixed, live, proven, repaired.
The fleet count of label-without-link buttons now stands at 194, and the
part of it this bug caused stands at zero. Two operational notes from the
run: work items for both rewrites reported "failed" although the work
succeeded — a result-delivery snag in the messaging layer, recorded in the
bug that owns that seam (217); and the dispatch queue serves oldest work
first fleet-wide, so our items would have waited hours behind a backlog —
we moved our own items forward and noted the synthetic timestamps.

**Where we're going.** Two decisions are yours. The ~194 remaining buttons
never had destinations: re-run destination-resolution site by site, accept
them as label-only, or open a dedicated lane — the handoff lays out the
options, and note the platform's own queue for these holds just 71 entries
across 6 sites, so most have never even been queued for a decision. And
webdesign.uk still carries the eight emergency locks from before the fix:
they can stay (belt and braces) or come off (the fix now protects those
rows); repair needed neither, since the row we thought was locked never
was.
