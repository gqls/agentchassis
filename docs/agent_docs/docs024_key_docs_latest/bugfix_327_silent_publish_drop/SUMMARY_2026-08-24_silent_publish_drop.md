# Summary — the silent publish drop, and what it turned out to be about

2026-08-24. Written to be read aloud. Second in the series; the 08-23 entry stands as written,
and two of its headline claims have since inverted — which is why this one exists.

---

## What we're trying to do

Make it impossible for someone to launch a piece of work, be told it worked, and be wrong.

## Where we've come from

There is one command that starts a whole site build — the thing we sell. Sometimes you ran it,
it printed all the right reference numbers, it exited cleanly, and **nothing happened at all.**
No build, nothing queued, no error anywhere. One submission in three vanished that way.

The cause was a race: the command sent its message by starting a throwaway container and feeding
the message in through the container's input, and if the container got going first it saw an
empty input, decided there was nothing to send, and exited successfully.

That much was already known and written down. What made it a project rather than a patch was the
census: **218 scripts sent messages this way, 25 had the documented fix applied, and two actually
checked that it worked.** The remedy had been published for a month and was being copied without
being made to function.

## What we've done

We stopped writing the remedy down and made it **callable** — one shared publisher that puts the
message in the container's start-up command so the race cannot occur, and that **insists on
hearing the confirmation and fails loudly when it doesn't**.

It also answers a question nobody could answer before. When work doesn't appear there are two
possible reasons, and the right response to each is the opposite of the other: the message never
left (send it again at once) or it left and nothing picked it up (wait — sending again just
makes a duplicate). Those were indistinguishable. They now produce different exit codes and
different instructions, and it checks the rejection records too, because a rejected message
leaves exactly the same silence as a lost one.

Since yesterday we have migrated the tools that matter most: the build trigger, the council
submission trigger, the tool that verifies our own hazard documentation, and the five `fire-*`
operator scripts. Ten callers now use it.

We also added a check that runs on every commit and warns anyone introducing the old pattern —
measured on 300 real commits before switching it on, where it fired five times and was right
five times.

## Where we are now

**The review question has reversed.** Yesterday the fix could not be reviewed at all: the code
lives in a directory the council doesn't cover, it declined, and we recorded that rather than
overriding it. You then widened the council's scope to cover our checking machinery — a file
holding 22 checks that run on every commit in every session, about half our detection surface and
the half that runs most often. The same submission that was refused on the 23rd is accepted on
the 24th. Running it is now simply a question of whether the credits are worth it.

**The "how much is left" answer has also reversed, and the old number was mine and misleading.**
I had been saying about 178 scripts remained. Two separate faults inflated that. Some of those
files cannot run at all — pasted notes with a `.sh` extension. And, more awkwardly, **a search
for the dangerous pattern also matches the warnings people wrote about it** — including the
explanatory comment I add to every file I fix. So the count didn't move when the work started
working. Corrected, and the honest breakdown is that of 155 remaining, 86 are one-off lane
scripts that should never be rewritten, 102 are dormant, six are duplicate copies, and eighteen
aren't publishers at all. **The genuine remaining work is eleven files.**

**And the strongest result isn't ours.** Two lanes we have never spoken to have picked up the
shared publisher on their own — one committed, one in progress. The safe method had been
documented for a month and had two users; it became something you could call, and strangers
started using it within a day.

Two things from yesterday still stand and should not be quietly dropped: we never did reproduce
the original failure under test, so the fix sidesteps the race rather than being shown to beat it;
and the old method was separately caught sending one message twice, which on the real system
means two builds.

## Where we're going

The bug stays open, and now says exactly what would close it: the eleven live scripts migrated,
the commit-time check left in place, and the dormant one-off scripts declared out of scope in
writing. That last part matters — without it this stays open for ever against a number dominated
by files nobody runs, while nothing actually improves.

Beyond that sits one better end state we have costed and not built: sending these messages from
inside the cluster, where the broker itself confirms receipt. That would make a silent loss
impossible rather than merely visible, and it would leave a permanent record — which matters more
than it sounds, because we discovered the table that records whether a message arrived keeps only
about two days, while the one recording rejections keeps a month. After forty-eight hours you can
still learn whether a message was refused but never whether it arrived.

If there is one thing to carry out of this piece of work, it is the last paragraph of the
handoff: **the missing thing was never the knowledge.** The scripts that got this wrong include
ones whose own headers warned about the exact trap they then fell into. When a class of mistake
keeps recurring despite being documented, writing the warning down more clearly is the answer
that has already been tried.
