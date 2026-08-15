# SUMMARY — bug 274, closed: completed workflows now deliver their results to their parents

*2026-08-15, the "bugfix 274" session (ae908c77). Written at closure — the lane opened and
closed within one day, so this is its only summary. The full technical record is
`bugs_closed/274_HANDOFF_2026-08-14_completed_workflows_cannot_deliver_their_result_to_the_parent_fleetwide.md`
(sections 10 and 11 are this lane's).*

## What we were trying to do

Fix the platform's biggest silent failure: for twelve days, every time a child workflow
finished its work and tried to hand the result back to the workflow that spawned it, the
hand-over failed — and the parent was told the child had *failed*, when in fact it had
succeeded. Roughly 16,900 occurrences across 60 agent types since the 3rd of August,
including the fleet's highest-volume repair paths. The knock-on damage was worse than lost
results: parents never retried (the fabricated error read as permanent), some marked whole
workflows failed, and some completed their work items with whatever unrelated data was lying
around — false completions that a sister investigation (bug 213) had been chasing from the
other end.

## Where we'd come from

The bug was filed the previous evening by the 213 lane as a handoff, with the root cause
already located from a live stack trace: the function that sends the "I succeeded, here is my
result" message builds its message headers by hand, and had *never* filled in two of them —
who is sending this, and which step it answers. Those two headers sat empty and harmless from
January until the 3rd of August, when an unrelated, correct fix (bug 158) routed this message
through the platform's outgoing-message validator for the first time. The validator requires
both headers, so from that day the success message was rejected every single time — and the
code's fallback for "couldn't deliver" is to send a failure notice instead, which travels by
an unvalidated path and always arrives. A deterministic defect, mislabelled "transient", told
everyone to try again forever.

## What we did

Verified everything first-hand at the current code before acting (three parallel research
passes: when it started and every caller; what the parent does with these messages; what
truthful values were available to fill the headers with). Then the fix, kept deliberately
narrow: fill both headers truthfully at the seam — the child's own resolved identity, and the
parent's spawning step name, both already recorded in the workflow's own state — and do the
same for the failure notice, which turned out to be missing its client id too. In the shared
kafka layer, the validator's rejection became a typed error which the delivery policy now
classifies as *undeliverable* (permanent for that message) rather than *transient*, so
operators are no longer told to retry the unretryable. The tests run the *real* validator
over the headers the code *really* produces — the previous tests used a stand-in that skipped
validation, which is exactly how this stayed invisible — plus a mutation control proving the
test can fail. The council gate approved it first round (nine seats; every objection advisory,
each dispositioned with the query it asked for; two consolidation ideas recorded as named
follow-ups rather than smuggled into the diff).

One shared-tree adventure worth retelling: mid-task, another session committed the contended
coordinator file and swept our half-finished edits into its commit — the known same-file
passenger hazard. Forward-only held: nothing was lost, the remainder went out under our own
message with the split stated in both places, and the joined result was verified to build and
pass from a clean checkout. The one misstep (writing "shipped in one commit" into the bug file
before the commit existed) went into WRONG_CALLS with the cheap check that prevents it.

## Where we are now

Closed — fixed, live, and proven with demand. The fleet rolled at 10:14Z today; the running
binary states a commit that contains both halves of the fix. In the two hours before the roll:
293 failures. After a nine-minute drain window (eight stragglers from old binaries, each
self-dated by carrying the *old* error label), the count is zero — while 23 child workflows
with parents completed in the same window, across exactly the agent types that used to fail
most. The "successfully notified parent" log line, unreachable since the 3rd of August, is
back in live pods. The bug file has moved to `bugs_closed/`, and every ledger this lane
touched (the debugging guide, the landmines file, the concept register's delivery-policy
entry) now tells the closed story.

## Where we're going

Nothing further is owed by this lane. Three small things are recorded for whoever wants them:
follow-up A, a shared reader for "which step am I answering" if a common home for it appears;
follow-up B, consolidating four identical hand-written identity blocks onto the new helper;
and two near-miss sites noted in the bug file that could reproduce this shape if their inputs
ever arrive empty. The parent-side question this bug does *not* fix — a parent completing a
work item with substituted data when no result arrives — remains bug 213's, now with its
trigger removed.
