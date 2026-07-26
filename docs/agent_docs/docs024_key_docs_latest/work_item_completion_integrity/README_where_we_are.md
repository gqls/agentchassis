# Where we are — work-item completion integrity

*Plain-prose log, append-only, newest at the bottom.*

---

**2026-07-18 — what this is about**

Some work items were marking themselves finished while carrying, in the same database
write, the error that proved they had done nothing. The improvement loop then believed
those defects were handled and stopped re-detecting them. So the platform was quietly
telling itself it had fixed things it hadn't touched. That's the bug.

It came in two halves. One action (`fix_forced_text_colors`) had been written but never
added to the registry, so every attempt to run it failed validation with a confusing
message about needing a "topic" — the workflow never ran at all. That's the small half.
The big half is that when those failures came back, the completion code looked at the
wrong field. There are two `status` fields one layer apart: one says "a reply arrived",
the other says "the work succeeded". It was reading the first and treating it as the
second.

**What I found that the bug report didn't**

The report said the cause was two hand-maintained lists drifting apart. That turned out
to be wrong — there was only ever one live list; the other was dead code that nothing had
called in a long time. What made it look like drift was a comment at the top of another
file telling developers to "register in TWO places". That comment had also been copied
into two guide documents. So the misinformation had outlived the code and was still
actively misleading people — including whoever filed this bug. All of it is now deleted.

The report also said two items were affected. There were 54, across six live sites, going
back to May.

**What I changed**

The completion code now refuses to mark something finished if the work itself reported
failure, and instead sends it back through the existing retry machinery. I picked the rule
by measuring rather than guessing: across the entire database history there is only one
failure word in use, and over the last 30 days the new check would have stopped 6 out of
1662 completions — all six genuinely broken. So it won't start rejecting healthy work.

I also added a build-time test so that an action which is written but never registered now
breaks the build, instead of failing silently in production months later. Another thread
hit the identical problem with a different action recently, which is why a one-off fix
wasn't enough.

**A decision I brought to you, and why**

The 54 bad rows needed correcting. You initially said re-queue them so they'd actually
run. I pushed back once, because that would have fired an action that has never once
executed successfully at five live sites — including a client rebuild and a site another
thread had just finished restoring. You settled it: mark them all failed and start fresh.
That's what I did, reversibly. Discovery can now re-file them cleanly if the defects are
real.

**Where I got it wrong**

Twice, and both are worth you knowing.

I told you the council resubmissions were being silently dropped and went hunting for a
transport bug. They weren't dropped — they were queued, about 16 minutes behind a backlog.
I resubmitted three times chasing hypotheses I hadn't tested, which turned one review into
four councils' worth of credits.

And I made a structural claim — that only one place in the platform could have this
bug — based on reading four functions and guessing about three more from their filenames.
The council's reviewers objected to exactly that, twice, and I brushed it off as
box-ticking. When you asked me to re-read CLAUDE.md I went and opened the three files. The
claim was right, but I hadn't earned it, and the reviewers were doing their job.

**What's left**

The Go changes don't do anything until a chassis image is built and rolled — that's a
fleet-wide action with other threads' work queued behind the same roll, so it's your call
when. Until it ships the bug stays in the open queue, because it's still reproducible in
production. After it ships, the check is to grep the running pod for the new function, not
to trust git or the image tag.

---

**2026-07-20 — it's live, and closed**

You shipped the image (v1.0.1139). I checked the running pod and both halves of the fix are
genuinely in it, so the case has moved to the closed queue.

Worth telling you how nearly I got that wrong. The obvious check was to grep the pod for the
action's name — it came back positive, and I almost signed off there. But that name was
already in the binary before the fix; the old image would have passed the identical check.
The thing I actually changed was a registry entry, so the honest test was to grep for a
phrase that only exists inside that entry — a bit of description text I'd written. That
came back positive too, and that one means something. Same for the guard: I grepped its
error message rather than the general area it lives in. I've written the pattern into the
debugging guide, because the misleading version of that check is the one CLAUDE.md tells
everyone to run.

Since the deploy, nothing is lying: the query that found 54 bad rows returns zero, eleven
items have completed through the new code without it wrongly blocking any of them, and
there are no validation failures anywhere.

One honest gap, which I've recorded rather than papered over: the part of the guard that
*blocks* a bad completion hasn't actually triggered in production yet, because nothing has
failed since the deploy — which is what you'd expect, given the other half of the fix
removed the thing that was causing those failures. Its logic is covered by tests, but I
can't claim I've watched it work on a live failure. If it ever does fire, it'll show up as
a work item whose error starts "completion blocked".

I also decided *not* to manufacture proof by re-running the colour fixer at one of the
three sites still holding old items. That would have exercised the fix nicely, but it would
also have edited a live site with an action the bug report itself called misconceived — and
your instruction was to mark them failed and start fresh, not to re-run them.

---

**2026-07-20 (later) — handing this over so you can pick it up in a fresh chat**

Everything is committed and nothing is left running, so there's no state trapped in this
conversation. The entry point for the next chat is
`HANDOFF_2026-07-20_start_here.md` in this folder — it has the 30-second state, what
shipped, what's next, and the traps we paid for, in reading order.

The short version of where this leaves you: the bug we started on is done and live, and the
thread now has **two things waiting**. Both arrived from the reasoning-dataset thread while
we were working, and both already carry council review that somebody has paid for — but
only one of them is still work for us, for the reason below.

The first one I nearly got wrong, and it's worth saying how. The note handed to us
describes a live defect: the one completion verifier we have treats "the component isn't
there any more" as "the problem was fixed" — when a missing component is equally what it
looks like when a rebuild has silently *deleted* it. So content loss could be filed as a
verified success, by the exact mechanism meant to stop us trusting "complete". Real
problem, and I was about to write "go and fix this" into the handover.

I checked first, and another thread had already fixed it that morning. What's left on it
isn't code at all — it just needs the next image build to go out, same as ours did
yesterday, and it belongs to the thread that owns that area. The general lesson, which I've
written down: a note handed to you describes how things were when it was written, not how
they are.

The second is the one you assigned to this thread on the 20th: adding a column that records
which auditor's judgement created a work item. Right now an auditor that flags twenty
non-issues looks identical in the data to one that flags twenty real defects, because
nothing connects a judgement to what happened to it. That one is three council rounds in
and close to approval.

So my suggestion is the provenance column — the one you assigned anyway. It's genuinely
unstarted (I checked the database; the column isn't there), it's small and additive, and
it's already three council rounds in with the remaining objections written down.

The other thing left over is bigger and needs a decision from you rather than code: we have
exactly one of these completion verifiers, covering one item type out of about fifty. That
gap is why the bug we just fixed could happen at all. Worth deciding what shape the answer
takes before anyone writes anything.

---

**2026-07-20 evening — the second fix is live too, and where I'd pick it up**

You rolled another image (v1.0.1140) and I checked the running pod: the verifier work
from this afternoon is in it. I checked it the careful way this time — grepping for a
scrap of text that only exists inside the lines I changed, rather than a name the file
already contained. One of the checks came back zero and that turned out to be correct:
the bit it was looking for only ever runs during tests, so it never reaches the shipped
program at all. Worth knowing before it alarms someone.

Nothing has actually completed since the roll, so I can say the code is there but not
that I've watched it work. It shouldn't change any outcomes anyway — it hands the
checker more information about what it's checking, nothing more.

The bug we were fixing this afternoon stays open, and I want to be straight about why.
We now have the machinery for these completion checks and a guard that stops the gap
quietly reappearing — but there is still only **one** actual check, covering one kind
of work item out of eighty-six. I built the thing that makes them possible and the
alarm that notices when they're missing. I did not build many of them.

The guard did earn its place, though. After the council pushed back, I rewrote it to
read the source code rather than rely on a list I'd copied out of the database. It
immediately found seventeen kinds of work item that list couldn't have known about —
they exist in the code but have never once been created, so nothing in the database
could reveal them. Each would have sailed through unchecked the first time it ever
happened.

Also: the other thread closed the deleted-component bug off the same image roll, about
ten minutes before I got round to writing them a note saying it was ready to close.
Their checking was sound. That's the third time today I've nearly written something
that was already out of date — I've written the lesson down.

Everything's committed and nothing is running in the background. If you start a fresh
chat, `HANDOFF_2026-07-20_start_here.md` in this folder is the way in — I've rewritten
it properly rather than patching it again, because it had drifted into contradicting
itself. The next job is the provenance column you assigned, which is genuinely
untouched and close to approved.

I have deliberately **not** written a new summary for this. The rule you added this
morning says they're for real turning points, not for every session — and answering the
five headings today would mostly repeat this morning's. This belongs here and in the
notes instead.


---

**24 July, evening — a different session picking up the verifier thread.** The
completion-checking work sat quiet since the 20th, so when the owner pointed a
session at bug 021 tonight it checked nobody else was on it and then did the next
thing this folder's own handoff asked for: the first real verifier against the
widened contract, for the hardcoded-colour items.

The one lesson this folder keeps teaching — check what the *fixer* actually does,
not what the *detector* looks for — turned out to matter immediately. The
detector flags any hex colour background; the fixer only replaces the dark
six-digit ones inside style blocks, on purpose. So the new verifier asks the only
fair question: "would the fixer's own find-and-replace still change anything?"
If yes, the item can't complete. If no, it completes even when out-of-scope
colours remain, because leaving those is the fixer behaving as designed. On
today's data that distinction is live, not academic: thirty-two components across
eight sites still match the broad pattern, and twenty-one items have already been
marked complete against them.

Two housekeeping finds along the way. First, the build guard this folder created
was already doing its job: it was failing, because two new item types had been
added by other threads without saying how they'd be checked. They're classified
now. Second, the hand-refreshed list of known item types was four days stale and
eight types behind — refreshed, and a rule written down that the list only ever
grows, because old rows get cleaned out of the database and a type that vanishes
from the data hasn't vanished from the code.

The code is committed and goes live with the next image build. Still owed after
that: actually watching the verifier refuse a completion on a dirty site once,
so we're not taking its presence for its behaviour. The council is reviewing the
change in parallel; one open question left for this workstream — the detector
and fixer disagree about scope, so these items will keep being re-detected even
when handled correctly, and deciding which side moves is a design call, not a
bug fix.

---

**2026-07-26 — the open question above has been answered.**

The last entry ended with "the detector and fixer disagree about scope … deciding
which side moves is a design call, not a bug fix". The owner has made that call,
and the answer was neither of the two sides we had framed.

We had been arguing about whether to make the scanner look for less or the fixer
do more. The owner's answer was: the scanner is right to see everything, the fixer
is right to be careful, and the mistake was never the disagreement — it was that
the scanner threw all of it into one pile and labelled the pile "someone will fix
this". So it now sorts. What the fixer can genuinely repair goes to the fixer, with
a truthful count. Everything else goes onto a list of *things we can see and
cannot yet mend*, which is a different kind of item entirely and one the system
already had a name for. Nobody is dispatched at it; it reads as a request for a
new capability, which is what it is.

The nice part is that the list is not new. The platform has been quietly keeping
one since long before — for pages whose builders don't exist yet — and there was
already something that reads it and turns it into a roadmap. We just weren't
putting anything on it.

Two things fell out of this that were worse than the bug we set out to fix. While
writing a check that every scanner routes work at a fixer that actually exists, it
immediately found two that don't — one of them for a problem marked "high" and put
near the front of the queue. Those items were never going to be picked up by
anyone; they'd have been marked "blocked" the moment something tried. That check
now fails the build, so there won't be a third.

And one thing we got wrong and caught: this thread wrote down that the previous
thread's numbers needed correcting. They didn't. We had measured the same thing
with a blunter instrument, on purpose, and then read the blunter answer as a
contradiction. It's written up in the wrong-calls file, because the shape of it —
using a rough measurement to argue against a precise one — is the sort of thing
that will happen again.

Nothing here is live yet. The code goes out with the next image build, and the
tidy-up of the three misleading items in the backlog has to wait until after that
(doing it first would let the old scanner re-create them). The steps to check it
actually works, including the deliberately awkward one where we prove the scanner
still *does* file work on a site that has real fixable problems, are written out
at the bottom of the closed bug file.

---

**2026-07-26, later — it's live, and the tidy-up is done.**

A new build went out (v1.0.1171) and the change is running. I checked it the
careful way rather than the flattering way: instead of looking for the new wording
and being pleased to find it, I checked that the *old* wording has disappeared from
the running program. It has. The old sentence that used to go into the backlog —
the one that blamed a fixer for failing — cannot be produced any more.

With that confirmed, and only then, I ran the database tidy-up. Order mattered:
doing it the other way round would have let the old scanner immediately re-create
the three misleading entries. It changed exactly the three we predicted, and left
alone the sites that do have real fixable problems, which was the property worth
having.

Then the real test. On finetuning.uk — a site where the scanner sees eight problems
and the fixer can repair none of them — the system now files one entry that says so
plainly, marked as needing a new capability, assigned to nobody, and not queued for
anyone to attempt. It filed no work for the fixer at all. That is exactly right, and
it is the first time the backlog has told the truth about that site since April.

One thing is still outstanding and I want to be straight about it: I have not yet
watched the *other* half work on a live site — a site where there IS something the
fixer can mend, where it should file real work AND a capability note side by side. A
good result on a site with nothing to fix looks identical to a scanner that has
stopped working altogether, which is why that second test exists. It's pinned by
automated tests; it isn't yet pinned by the fleet. A run is in flight.

A small operational note, because it cost time twice today: dispatches sent shortly
after a new build goes out can vanish without a trace — no error, no record,
nothing. It happened to a review submission earlier and to one of these test runs.
The fix is a five-second check of when the system last restarted, *before* sending
anything, rather than waiting twenty minutes and then investigating.
