# SUMMARY — 2026-09-03 (second), bugfix 427 (event render)

Second summary of the day, warranted because the read-out genuinely changed: the first summary
said 427's fix "should populate on its own once a chassis rolls"; this one says it did, is on
the served artefact, and the underlying mechanism is a closed bug rather than an open one.

## What we're trying to do

boxingonline.com needed a fight calendar that lists the fixtures we can actually evidence. The
morning's summary covered how that got built. This summary covers what happened once the fix
actually reached production.

## Where we've come from

By this morning, the calendar page was showing an honest empty state instead of nothing at all,
and the reason it wasn't showing the one real fight was diagnosed as `bugs_open/454`: a struct
field computed and silently discarded in the light re-render path, affecting every page on the
estate that draws on non-authored data — 1,855 sections, not just this one. The fix was one line,
council-approved, and waiting on a chassis roll.

Waiting turned out to be two separate things. The first "fresh build" reported this afternoon
was checked and found to be nothing — same pods, same commit, started before the fix even
existed. That got recorded as a dated negative rather than acted on. When the real roll landed,
the re-render still failed at its last step: a different lane's guard, live in the very same
image, correctly refused to let a tool-type page save through the generic path. Their fix landed
within the hour, and the second roll carried both.

## What we've done

Re-dispatched the moment the second roll was confirmed real by arithmetic, not by pod count. It
completed cleanly through every step, including the one that had failed twice today. The fixture
resolved — `items` 0 to 1, rendered HTML 1,813 to 2,498 bytes — and the save reached the actual
deploy pipeline: a real git commit, a real GitHub Actions run, a real upload to the bucket that
serves the site. Traced past the job status to the "Sync to B2" step's own delete-then-upload
lines, which is the standard this lane set for itself and held to under time pressure.

`bugs_open/454` is now closed, not just fixed — moved to `bugs_closed/`, meeting the estate's own
bar of fixed AND live. Closing it produced the most interesting incident of the day: another
session, working an unrelated two-day-old regression that turned out to be this same mechanism,
was writing to the same file at the same moment it was being moved. Two well-formed, independent
git commits interleaved into a hole — for a few minutes, the shared history held zero copies of a
file two sessions actively cared about. It was recoverable, cheaply, because the session who
noticed diagnosed it precisely and declined to fix it unilaterally, correctly treating the close
as someone else's decision. Restored, verified at HEAD, and written up as a new class of trap
distinct from the one already on record — the existing landmine is about a mover dropping half of
their own commit; this one is about a third party's ordinary commit being made wrong by someone
else's concurrent move.

That same collision brought real value with it: independent corroboration of 454 from a second
regression, on a different component family and a different non-authored data source, including
an experiment whose before-state provably could not have faked the after-state. And, separately,
a third lane's canary — dispatched before any of this, resolved after — gave the first evidence
in the whole file verified all the way to a served page rather than a database row: fetched
independently, images present, nothing broken, checked against a baseline this session had taken
minutes earlier for an unrelated reason and so could not have been shaped to fit.

## Where we are now

Everything upstream of the artefact is proven. The fight-calendar page's underlying data is
correct and deployed; it will show the fixture once its own artefact catches up (git deploy is
real, DNS/preview lag is a separate, already-understood, pre-existing gap, not new work). The
mechanism bug that blocked it is closed. The council round covering this lane's three migrations
is approved.

What remains is not code. Answering the council's own advisories turned up that this lane's three
migrations — the fix for the section-order defect, and the fix for that fix — are transient: the
page's underlying plan still names the old composition, and the next routine sync for this site
will silently revert all three, re-arming the very detector they were written to disarm. That is
now the lane's top open item, and it is a decision about a table's relied-upon immutability, not
a migration to write at the end of a long session. And one detector, running overnight, is still
the actual closing signal for 427 itself.

## Where we're going

The next session inherits three things in order of size: watch for the overnight reclassification
that closes 427 formally; do not let a re-plan of this site quietly undo today's fixes without
someone having made the call about `site_plan_sections`; and treat the shared-tree collision this
afternoon as a pattern now written down, not a one-off to be surprised by again.
