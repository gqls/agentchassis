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

---

## 2026-08-26, after the new build went out — it's live, and it's proven live

The review board approved the change first time, with four pieces of advice and no objections
serious enough to hold anything up. All four were dealt with the same day — three by doing what
was asked, and one by showing with a measurement that the concern, though reasonable, didn't
apply here.

Then the new build was deployed, and I checked it the way we agreed checks should be done: not by
trusting the deploy report, but by asking the running program itself whether it contains the new
safeguard. It does — on both of the machines running it. I also ran two deliberate control checks
alongside (one thing that must be found, one that must not be), so the answer can't be an
artefact of a broken test.

One honest caveat, written down rather than glossed: since the deploy, the safeguard has reported
nothing — which is exactly what we expect, but in the short window since the roll no page
re-renders have actually happened yet, so "nothing reported" doesn't yet mean much on its own.
The proof it works remains the tests that deliberately break it on every build and watch it
complain.

So: the hole is closed and live, the tripwire against 200-odd similar holes is active, and the
paperwork — bug file, register, runbook — all says what is actually true. What's left belongs to
other lanes or to future decisions, and the handoff file lists each piece with its owner. This
lane's work is done unless someone reopens a piece of it.

> **Correction to the "225 places" figure above (same day):** the true count is **207**. My
> counting script also matched a second, harmless pattern that the safeguard doesn't apply to.
> Caught while building the tripwire, before anything shipped; the fuller story is in the
> notes file. The plan didn't change — either number is far too many to fix by hand.

---

**2026-08-31.** A new session picked the lane up from the handoff and did two things.

First it checked that everything we said on the 26th is still true five days on — it is. The
fix is still in both running pods (they've been redeployed twice since; checked again at the
binary, with the same controls). The count of risky scan-loops across the codebase is still
exactly 207 even though about twenty new code files landed last week — none of them added a new
one, which is the ratchet doing its job quietly. And the thing we couldn't say on the 26th, we
can now: back then the guard had never been exercised by real traffic, so its silence proved
nothing. Since Sunday there have been at least 87 real page-rerender runs through the exact code
we guarded, and the guard refused none of them — meaning the reads are genuinely completing. Its
silence is now evidence, not absence.

Second, it closed the one loose end we'd left in the code itself: a spot in the same loop where
a section whose content failed to decode would quietly carry on as an EMPTY section — and then
get saved back, wiping the stored content. We'd left it with a note saying "this needs a
decision: is an empty section acceptable?". The session answered it by measuring instead of
philosophising: the database column literally cannot hold broken JSON, and not one of the 2,751
live content entries has the only shape that could still fail. So making it refuse costs
nothing today, protects against the first future mistake, and it now uses the same refusal
mechanism the council already approved. The 55 pages that legitimately have no content at all
are unaffected — there's a test that pins that, so nobody can accidentally tighten it later.

Sent to the review council as its own round (submitted, verdict pending). One item left on this
lane's list for the future: the big fleet-wide "candidate 1" design round, which remains
unclaimed. The other candidate we'd listed was picked up and shipped by the 404 lane themselves.

---

**2026-09-02.** The new build went out and we checked it the proper way — asked both running
servers directly whether they carry the new protection. They do. So the loose end closed on the
31st is now actually protecting production, and with plenty of real traffic behind it: about
1,400 page-rebuild runs since the start of the month, 176 of them straight through the guarded
code, not one refused. Silence with traffic behind it is the good kind of silence.

That makes this lane finished. Everything we set out to do is built, reviewed, deployed and
verified. The one thing still attached to the bug file is a bigger design question that was
never ours — "when the system meets a value it doesn't understand, should it refuse rather than
quietly carry on?" — which needs its own round, and which the 404 team has half-asked from
their side too. Your call needed: either we split that question into its own file and close
this bug, or the bug stays open as the placeholder until someone takes it on. We'd suggest the
split — everything investigative in the file is done, and an open file with no one working it
just reads as neglect. Full picture for the next session:
HANDOFF_2026-09-02_continue_here.md in this folder.
