# SUMMARY — work-item completion integrity

## What we're trying to do

Make sure that when the platform says a piece of work is finished, it actually happened.
The improvement loop is only as good as its record of what it has already fixed — if that
record can lie, the loop stops looking at real defects and its progress becomes fictional.

## Where we've come from

A bug filed from the robot-hands thread (`bugs_open/017`) noticed two work items that had
"completed" twice without changing anything, each storing the error proving nothing ran.
The filed diagnosis blamed two hand-maintained lists drifting apart, and put the scale at
two items.

Both of those turned out to be wrong. There was only one live list — the other was dead
code, kept alive in people's heads by a stale comment instructing developers to register
things in two places, a comment that had propagated into two guide documents and had
misled the bug's own author. And the real scale was 54 items across six live sites,
stretching back to May.

## What we've done

Closed both halves of the bug and the class behind it. An action that had been written but
never registered is now registered. More importantly, the completion path no longer
confuses "a reply arrived" with "the work succeeded" — two `status` fields one layer apart
that had been read interchangeably. A failed saga is now routed back into the existing
retry machinery rather than stamped complete.

The rule for what counts as failure was chosen by measurement, not intuition: only one
failure word exists anywhere in the database's history, and over 30 days the new check
would have blocked 6 of 1662 completions, all six genuine. An unfamiliar verdict still
completes — refusing to guess — but is now recorded to a queryable table rather than a
pod log that vanishes on the next rollout.

To stop it recurring, the build now fails if an action is written without being registered,
with an explicit list of the two that are dormant on purpose. Another thread had hit the
identical problem with a different action days earlier, which is why registering one action
was never going to be enough.

The 54 bad rows were corrected to `failed`, reversibly, on the owner's decision — visible
as failures, free to be re-detected cleanly, but not force-dispatched at live sites.

The change went through the advisory council twice, ending at eight approvals to two
objections, and both rounds' objections produced real improvements.

## Where we are now

Committed in three commits on `085_debug_and_feature_loops` and **inert**. Go changes do
nothing until a chassis image is built and rolled, so the defect remains reproducible in
production and the bug correctly stays in the open queue.

Two process failures are recorded alongside the fix, because they cost more than the bug
did: a queued council run was misread as a dropped one and resubmitted three times on
untested hypotheses, and a structural claim about the platform was asserted from filenames
before the files were opened — the council caught the latter twice and was initially waved
off.

## Where we're going

One step: ship a chassis image and verify against the running pod by grepping its binary
for the new guard, never against git or the image tag. That roll is a fleet-wide decision
with other threads' work queued behind it, so it is the owner's to sequence. Once it is
live and verified, this case moves to `/bugs_closed/`.
