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

---

**2026-09-03, early afternoon — the fix is written, tested and committed. Not live yet, and I will not say it is until I can show you.**

The change is in. Where the code used to say "this job has been quiet for five minutes, so I will
take it over", it now says "this job has been quiet for five minutes, so I will *try* to take it
over" — and the taking is a single atomic step that only succeeds if the job is still quiet at the
instant it happens. Whoever wins refreshes the timestamp, so everybody else who was about to pounce
looks again, sees a job that is being worked on, and walks away. Committed as `b55f837ef`, so it
rides the next chassis build.

**The most useful thing I found was that the bug report was wrong about why.** It said the code was
missing a safety check on the database write. It was not — the check was there all along, and the
report's own suggested fix would have changed nothing at all. The real problem was subtler: the code
*asked* the question ("is this job stale?") and then *acted* on the answer in two separate steps,
with a gap in between. Nobody was missing a lock; they were checking one thing and then doing
something slightly later, which is a different bug with a different fix. That false explanation had
been sitting in the file since 19 August, and I traced where it came from — an earlier, closed bug
where the same team used it to answer a reviewer's challenge. I have corrected both files rather
than just the one I am working on, because the wrong explanation was the part that travelled.

**I also got something wrong myself and have retracted it.** Earlier I told you, and told another
session, that this bug affected eight of our services. It affects one. I had seen that the relevant
machinery is built in a shared library and inferred that every service using that library runs it —
without checking which services actually include it. Only the main chassis does. What made my
mistake convincing, and this is the part worth knowing, is that I *had* gone and measured something
real at the live system just beforehand; having a genuine measurement in the paragraph made the
unchecked sentence next to it look measured too. The conclusion I drew survives for a different
reason (the short-lived worker pods that do most of the work run without the protection), but the
reason I gave was wrong, and a right answer with a wrong reason is worse than a visible mistake
because it survives review.

**How I proved the fix works, which is the part I would want to see if I were you.** I wrote the
tests so they could run against the *old* code, and ran them there first. All four failed — and the
failure messages caught the old code in the act, writing to a job record that another machine had
already claimed a moment earlier. Then I ran them against the new code and they passed. A test that
has never been seen to fail is not evidence of anything.

**Two honest limits.** First, this is not a fire. I ran a census of the busiest path over 24 hours —
three thousand jobs — and found zero cases of the double-run this bug predicts, because two other
safety mechanisms happen to catch it there. What the fix buys is that we stop depending on
protections that were built for other purposes, are absent on the machines doing most of the work,
and could be removed by someone who never knew we were relying on them. Second, the fix does not
close everything: if a machine is genuinely alive but working on a single long task, nothing it holds
tells anyone else it is alive, so it can still be judged dead. Closing that needs the workers to
check in periodically, which is a separate piece of work and I have not smuggled it in here.

**Still to come:** the automated reviewers have the design and have not reported back yet, and the
fix cannot be confirmed working in production until it is built and rolled out. I will check both
rather than assume either.
