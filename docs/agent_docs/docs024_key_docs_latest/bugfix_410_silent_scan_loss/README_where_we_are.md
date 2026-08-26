# Where we are — bug 410, the silent scan loss

Plain-prose running log, append-only, newest at the bottom. Owner's document.

---

## 2026-08-26, late morning — what this bug is, and why I picked it up

**The bug in one sentence.** When the system re-renders a page, it first reads that page's
stored pieces out of the database. If reading one of those pieces goes wrong, the code writes a
line to the log, skips that piece, and carries on as if nothing happened. The page is then
rebuilt **without** it. Nothing fails, nothing is flagged, the job is marked complete, and the
page gets a fresh "last built" stamp. So a page can quietly lose a chunk of itself and end up
looking *more* freshly built than the page next to it that is fine.

**Why it was filed as a pattern rather than a bug.** Three different people, on three different
bits of work, hit the same shape in one week. Each time the system chose the quiet option when it
hit something it did not understand, and each time it reported success. The person who filed it
made the point that this is worse than three bugs, and I think they are right: **the safe default
and the silent-failure default are the same default.** Doing the quiet thing is nearly always
correct here, so nobody would ever think to look — which is exactly why every kind of drift ends
up landing there.

**Why nobody was working on it.** It was filed deliberately unowned. The lane that found the
third instance had it in their hands and explicitly refused to fix it inside their own piece of
work, on the grounds that it sits on the busiest pipeline we have and does not belong bolted onto
a feature. That was the right call and it left the bug sitting there. I checked with the
ownership tool, checked the running sessions, and read both lanes' notes — no fix in progress,
just people writing the case up. So I have taken it.

## What I found in the first hour, and one thing I got wrong

**The problem is much bigger than the one place.** I wrote a script to go through the whole
codebase and count how many places read database rows and quietly skip the ones they cannot read.
There are **225** of them. That immediately rules out the obvious plan: you cannot go and fix 225
places by hand, and trying would be a bigger risk than the bug. So the shape of the fix is: fix
the one that is biting, and put a tripwire in place so no **new** ones can be added without
somebody noticing.

**The fix is not something we have to invent.** This is the good news. We already do exactly this
in three other places in the same codebase, one of them a few files away — including one where a
reviewer *made* us do it, on this same shape, and refused to let the original quiet version
through. So instead of arguing for a new safety mechanism, I can point at one we already agreed
to and simply say: this place never got it. That is a much easier case to make and a much easier
one to review.

**The question a reviewer was always going to ask, answered up front.** Somebody warned me that
when they touched this same pipeline this morning, the reviewers pushed back hard: it is the
busiest thing we run, so what happens on day one? Fair question, and the honest version of it is
uncomfortable — my change turns "quietly does the wrong thing" into "stops and complains". If
pages are currently limping along broken, my fix converts them into visible failures.

So I measured it. **The answer is that nothing changes on day one, and not by luck.** Every piece
of data being read is either guaranteed by the database to exist, or already has a fallback for
when it does not. There is no page today that could trip the new check. It can only ever fire if
somebody later changes the code in a way that breaks the read — which is precisely the change
that caused seven tests to fail this morning and told nobody why. That is a far better answer
than "we'll roll it out carefully and watch", and it is the one I will lead with.

## The embarrassing part, which is also the useful part

I got three measurements wrong in that same hour, and a fourth was caught by somebody else.

The fourth is the one worth telling you about. Another lane had published a figure that had
already been flagged as second-hand. I did the diligent thing — refused to take their number,
re-ran it myself, got **exactly** the same answer, and wrote down that I had now confirmed it
first-hand.

Then they messaged to say their own figure had been wrong all along. The table we had both
queried quietly moves old rows out to an archive once they are finished with, so we had both
counted a slice and called it the whole. The real number was off by a factor of about seventy.

**I verified their number by making their mistake.** And because we matched to the digit, it felt
like confirmation — it was actually just two people using the same wrong method. The lesson I
have written down is that re-running someone's query only checks their arithmetic; it does not
check whether either of you is looking at the right thing in the first place. Two people agreeing
exactly is evidence that they used the same method, and nothing more.

I mention it because it is the same failure as the bug itself. A thing that returns less than it
was given, and reports success. I managed to do it to myself four times in an hour while writing
about it.

## Where we are now

The bug is confirmed still real, the code has not moved under me, and the two lanes whose work
touches the same file have both replied. One has handed me the function and told me to go first;
the other has confirmed we are not building overlapping checks and handed over a useful precedent
unprompted. I have a green test baseline, the design question settled, and the reviewer's hardest
question already answered with a number.

Next is the fix itself, then the review board, then getting it committed for the next build.
