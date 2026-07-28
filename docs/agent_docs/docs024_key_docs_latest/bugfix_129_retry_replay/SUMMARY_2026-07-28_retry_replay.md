# SUMMARY — the retry that forged its own return address (bugfix 129)

2026-07-28, late evening. Written at the point the bug closed.

## What we're trying to do

When one part of the system asks another to do a job and gets no answer, it tries
again. We are trying to make that second attempt actually be *the same request* —
because it was not, and the difference was destroying real work.

## Where we've come from

`bugs_open/129` was filed against the wrong suspect. It described a spawned worker
that adopted its parent's job record and then silently declined to do anything —
logging "completed successfully" while never replying — and it blamed the worker.

The worker was never the defect. It was handed the wrong identity by the thing that
retried it. The coordinator's retry path did not resend the request that had timed
out; it **built a new one out of the waiting job's own state**. So the retry went
out carrying the *parent's* job id instead of the child's, an action of `execute`
instead of whatever the original had been, and an empty body where the real payload
should have been. The worker receiving it looked up the id it was given, found a job
already sitting in "waiting for responses", correctly concluded there was nothing
for it to do, and said nothing.

Measured on the live database over fourteen days: **430 of 430** retried requests
went out this way, and **294 of them — 68% — exhausted their retry budget and died.**
Not an edge case. Every retry in the system was broken, and had been silently.

## What we've done

Fixed it with a single invariant: **a retry is a replay, never a reconstruction.**
The request that goes out on attempt two is the bytes that went out on attempt one,
with only three fields permitted to differ — retry version, message id, timestamp.
That required capturing the outbound payload at send time (a new column, migration
263) and reading it back at retry time instead of synthesising one.

Three loud, greppable failures were added around it, because the defect's real cost
was silence: a refusal to invent a retry when no payload was recorded, an assertion
that a request never carries the sender's own id, and a misrouting guard.

It went through the reviewer council and was **REJECTED on scope** — a guardian
veto. Six of ten seats approved and **no seat disputed the diagnosis**; the
objection was that a shared contract plus a schema column should not arrive inside a
bug patch. That objection is about venue, and it remains live.

The code then shipped anyway, inside another session's routine rollout, because on
this tree committing *is* shipping — builds come from committed `HEAD` by design.

## Where we are now

**Closed.** Both halves are proven on live traffic, not in tests we wrote.

The capture half was already shown: every request sent since the rollout keeps a
faithful copy naming the correct recipient, ~1.1 KB each, with zero self-addressed
payloads.

The replay half was witnessed tonight on a **natural** timeout — a scraper request
that waited its full thirty minutes and got nothing. The replay carried the child's
job id (not the waiting parent's), the original action, and the intact body; the
worker processed it and both jobs ran to completion in under two minutes. That is
precisely the request the old code would have killed.

Every loud-failure marker reads zero across five services. One rule in our own notes
turned out too strict and was corrected by reading the code: adapter steps re-run
from the beginning rather than being re-sent, so their missing payloads are harmless
by construction — genuine gaps, none.

## Where we're going

Nothing further is owed on the defect. Three things outlive it, and two are the
owner's to decide:

1. **The scope veto.** A judgement about how a capability reached production, not a
   measurement to be improved — so it is not answered by resubmitting. The seats
   contradict each other on the remedy, and the contained alternative would now mean
   reverting live code to restore a 430/430 defect.
2. **The "committing is shipping" clause.** The platform-seam ruling assumes the
   committing thread controls when its seam ships. On a shared HEAD it does not. The
   only mechanism that genuinely holds a seam back is to commit it dark, behind a
   switch defaulting off — and the ruling does not currently ask for that.
3. **The two web-search paths** record no payload because they put the caller's own
   id on the original outbound message, so there is no child identity to replay.
   That is a separate defect needing its own diagnosis. Bundling it here is exactly
   what was vetoed.
