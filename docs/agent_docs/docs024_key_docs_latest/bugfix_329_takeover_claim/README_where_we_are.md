# Where we are — bug 329, two machines can start the same job

Plain-prose log for the owner. Append only, newest at the bottom.

---

**2026-09-03, late morning — I picked this one up because nobody had it, and it is still real.**

You asked me to take an unowned bug. This is the one I took, and here is what it is in plain terms
before any of the detail.

When one of our machines picks up a piece of work, it writes down that it is working on it. If that
note has not been touched for five minutes, another machine is allowed to assume the first one has
died, and take the job over. The problem is the word "assume". It does not ask permission and
nothing hands the job over — it just starts. So if the first machine is actually alive and simply
busy, or merely slow to update its note, **two machines run the same job at the same time**, and
whatever that job does to the outside world, it does twice.

The fix we want is not a longer wait. Making the timer bigger only makes the collision rarer; it
does not make it impossible, and this estate's rule is to rank fixes by what makes the bad
situation *unrepresentable* rather than unlikely. The good news is that we already own the right
tool — there is an existing mechanism elsewhere in the same file that does a proper hand-over, where
exactly one machine can win and the loser is told it lost. Neither of the two places with this bug
uses it.

**Two things I found at the live system that the original bug note did not know.**

The first makes it worse. The bug was written on 19 August and never says how many machines we are
actually running. We are running **two**, right now. So this is not theoretical.

The second makes it more interesting, and I want to flag it because it is the kind of thing that
makes a fix look like it worked when it did not. The main chassis has, since that note was written,
gained a separate safety mechanism for an unrelated reason: incoming messages are now queued and a
worker has to claim a job before running it, and that claim is held in the shared database so only
one worker anywhere can hold it. That claim mostly hides this bug on that one service. It does not
remove it, for three reasons I verified rather than guessed: every *other* service we run has no
such claim at all and there are eight of those running two or three machines each; the claim's lease
runs out after three minutes while the "assume it is dead" timer is five, so a job can change hands
before it even looks stuck; and when a claim does change hands, the old holder finishes the piece of
work it is already inside — it only stops before starting the *next* one. That last window is
precisely the situation this bug describes.

So the honest summary is: the bug is real, it is partly masked on one service by a mechanism built
for something else, and the masking is exactly the sort of thing that would let me "prove" a fix
with a test that was never testing my fix. I have written that trap into the notes so I do not fall
into it, and so the next person does not either.

**What I am doing next.** I have asked for a design, and I will put it through the automated
reviewers before it goes anywhere near production. You have told me a fresh chassis build is going
out within the hour. I will only commit this in time for that build if it is genuinely finished and
tested — a half-verified change to the part of the system that decides who runs what is not
something to hurry into a release. If it is not ready, it waits for the next one, and I will say so
rather than implying it made it.
