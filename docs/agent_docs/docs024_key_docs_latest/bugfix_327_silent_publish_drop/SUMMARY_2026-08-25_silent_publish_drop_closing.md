# Summary — the silent publish drop, at close

2026-08-25. Third and final entry in the series; the 08-23 and 08-24 files stand as written. This
is the one to read if you only read one — but the earlier two are where the wrong turns are, and
those are the part that cannot be rederived.

---

## What we were trying to do

Make it impossible for someone to launch a piece of work, be told it worked, and be wrong.

## Where we came from

One command starts a whole site build — the thing we sell. Sometimes you ran it, it printed every
reference number, it exited cleanly, and **nothing happened at all.** No build, nothing queued, no
error anywhere. One submission in three vanished that way.

The cause was a race. The command sent its message by starting a throwaway container and feeding
the message in through the container's input; if the container got going first it saw an empty
input, decided there was nothing to send, and exited successfully. The container was then deleted,
taking the evidence with it.

That much had been known and written down for a month. What made it a piece of work rather than a
patch was the census: **218 scripts published this way, 25 had the documented fix applied, and two
actually checked that it worked.**

## What we did

We stopped writing the remedy down and made it **callable**. One shared publisher that puts the
message in the container's start-up command, so the race cannot happen, and that **insists on
hearing the confirmation and fails loudly when it doesn't**.

It also answers a question nobody could answer before. When work doesn't appear there are two
possible causes, and the right response to each is the opposite of the other: the message never
left (send again at once), or it left and nothing consumed it (wait — sending again only makes a
duplicate). Those were indistinguishable, and our standing advice was correct for one and exactly
wrong for the other. They now produce different exit codes and different printed instructions, and
it checks the rejection records too, because a rejected message leaves the same silence as a lost
one.

Then we migrated every live script that used the old pattern — twenty-one callers in all, each
tested by pointing it at a broken address and confirming it stops and names what will not happen.
And we added a check that runs on every commit so no new ones can appear unnoticed.

## Where we are now

**Closed.** No live script can exit cleanly having sent nothing. What remains carrying the old
pattern is about fifty-five files nobody has touched in over a month, plus lane scripts that record
what somebody did once and should never be rewritten. If any is ever picked up, the commit-time
check catches it then.

Three things worth carrying forward, in order of how much they should change future behaviour.

**The strongest result wasn't ours.** Two lanes we never spoke to adopted the shared publisher
within a day of it existing — one committed, one mid-flight. The safe method had been documented
for a month and had two users. It became something you could *call*, and strangers reached for it
immediately. **When a documented mistake keeps recurring, writing the warning down more clearly is
the answer that has already been tried.** The scripts that got this wrong include ones whose own
headers warned about the exact trap they then fell into.

**We never reproduced the original failure.** Ten of ten old-style sends arrived on the day we
tested. That excludes the four-in-five loss rate measured last month but only bounds today's below
about a quarter. The new approach doesn't beat the race so much as sidestep it — a stronger
position, but a different claim, and we should not say we won a race we never observed. Separately,
the old method was caught sending one message *twice*, which on the real system means two builds.

**Our own writing kept breaking our own instruments — three times in four days.** The count of
remaining work didn't move as the work got done, because fixing a file adds a comment mentioning
the problem, and the count matched comments. A completion check flagged two files we'd just fixed.
A verification reported a claim false when it had matched a comment quoting the claim. Each time it
was caught within a minute by asking *why* the answer was surprising. **The fix lives in how the
check is written, not in knowing about it** — which is the same lesson as the first one, aimed at
ourselves.

## Where we're going

Nowhere, by design. Two optional improvements are written up and neither is blocking:

A **council review** is now possible (the scope was widened to cover our commit-time checks), but
would review the detector rather than the publisher, which is still outside scope. Genuinely
balanced; no recommendation.

A **stronger in-cluster version** would make a silent loss impossible rather than merely visible,
and would close a real gap — the record of whether a message *arrived* is kept for about two days,
while the record of whether it was *refused* is kept for a month, so after forty-eight hours one of
those questions becomes permanently unanswerable. Recommendation: wait. The trigger to revisit is
specific — a message observed lost *through* the new publisher. Nothing else.

Both live in `HANDOFF_2026-08-25_open_decisions.md`, deliberately outside the bug file, because a
question left inside a closed bug is how one gets forgotten.
