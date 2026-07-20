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

**Live and closed.** Shipped in v1.0.1139 on 2026-07-20 and verified against the running
pod — not against git, and not against the image tag. Since the deploy the defining query
returns zero, eleven items have completed through the new path without a single false
block, and there are no validation failures anywhere. The case has moved to
`/bugs_closed/`.

The verification itself produced the sharpest lesson of the work. The obvious check — grep
the running binary for the action's name — passed, and would have passed identically
against the old image, because that string predated the fix. Proof required a symbol that
could not exist unless the change shipped: a phrase from the registry entry's own
description, and the guard's own error message. That pattern is now in the debugging guide,
because the misleading version of the check is the one the project's own instructions tell
every thread to run.

Two process failures are recorded alongside the fix, because they cost more than the bug
did: a queued council run was misread as a dropped one and resubmitted three times on
untested hypotheses, and a structural claim about the platform was asserted from filenames
before the files were opened — the council caught the latter twice and was initially waved
off.

## Where we're going

**Cold-start entry point: `HANDOFF_2026-07-20_start_here.md`.**

The thread does not stop here. Two council-reviewed assignments arrived from the
reasoning-dataset thread while this work was in flight, and neither is started. The first
is a live defect in the one completion verifier that exists: it reads a missing component
as a successful fix, when absence is equally the signature of a rebuild having deleted it —
so content loss can be recorded as a verified success by the mechanism built to stop
`complete` being taken on trust. The second, assigned to this thread by the owner on
2026-07-20, records which auditor's judgement created a work item, so that ~15,000 LLM
judgements a month stop being unattributable — today an auditor that flags twenty
non-issues is indistinguishable in the data from one that flags twenty real defects.

Nothing is outstanding on the 017 case itself. One thing is watched rather than finished: the guard's
*blocking* path has not yet fired in production, because nothing has failed since the
deploy — which is the expected consequence of the other half of the fix removing what was
causing those failures. Its logic rests on tests rather than on an observed live block, and
that is stated plainly in the case file rather than implied away. If it fires it will be
visible as a work item whose error begins "completion blocked", or as an
`UNKNOWN_HANDLER_VERDICT` row in the agent error log.
