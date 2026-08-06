# Where we are — bug 168, the asset path helper

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning

I picked up bug 168 off the open pile. It had been filed on 31 July by the lane that fixed
128, and nobody had touched it since. I checked that twice — once with the ownership script,
and once by reading the live transcripts of the 27 other Claude sessions currently working
this repo, because the script only knows about work that has already been committed and half
these sessions are mid-fix. None of them was near this one.

The bug is about a small function with a big job. When the platform generates an image — a
hero, a logo, an icon, a social card — it commits the file into the site's git repo and then
some page has to point at it. The function `DeployedWebPath` is what everything asks for that
path. Six different places call it.

The complaint in the bug file is that this function gets one case wrong: the social card. It
answers `og_card.png`, with an underscore, when the actual file on disk is `og-card.png`,
with a hyphen. That is true. But when I went to fix it I found the filed explanation of *why*
was wrong, and wrong in a way that mattered — one of the four suggested fixes would have made
things worse rather than better. That is the first thing worth saying.

**What was actually going on.** There are two different bits of code that publish these image
files, and they name files differently. The main one derives the name from the image's
purpose. The other one — the bit that makes the favicon and the social card — just writes
`favicon.png` and `og-card.png` as fixed names, because those are what browsers and Facebook
expect. `DeployedWebPath` only ever knew about the first one. And here is the thing: from
what a caller passes in, there is no way to tell which of the two published any given image.
So every single place that used this function had to separately remember "…except for the
favicon and the social card". Some did. One nearly didn't, and if it had shipped it would
have reported a broken social card and a broken favicon on every site we run.

The suggested fix I nearly took would have swapped underscores for hyphens everywhere. That
sounds harmless. It isn't: the main publisher *also* uses underscores in that situation, so
the two agree today, and the "fix" would have pulled them apart. I only caught it by going
and reading the publisher instead of trusting the comment on the function, which claims to
mirror it. A comment claiming two things match is not the same as two things matching.

**What I did instead.** I made it one function. Where before there were two copies of the
naming rule — one in the publisher, one in the lookup — kept in step by a comment saying they
matched, now both call the same code. They can't drift, because there's nothing to drift
from. And the special case for the favicon and the social card lives inside that one
function, so none of the six callers has to remember it any more. The one that had been
carefully remembering it can now stop.

**On being wrong, deliberately checked.** Before I asserted any of this I put it through the
diagnosis loop, which is the thing that reads the real code and the real database and tells
you whether your theory holds. It came back **REFUTED**. That is worth explaining rather than
hiding, because it was useful in both directions. It was right about the thing that matters
for how alarmed anyone should be: the code that writes the social card tag into the page
doesn't use this function at all, it writes the correct name directly — so nothing is broken
on any live site today, and this is a trap waiting to be sprung rather than a fire. That
matches what the bug file itself said, and it's why I've described the change as removing a
hazard rather than fixing an outage.

It was also wrong about something, and I've written that down rather than just accepting the
verdict. It claimed the function had only one caller. It has six. Its conclusion rested on
having missed five of them, including the one place where the problem genuinely bites. So I
took the useful half and recorded the rest as an error of the tool's, with the evidence.

**Where it stands.** The code is written, builds clean, and all the tests pass. I don't
trust "the tests pass" on its own, so I broke the code three separate ways on purpose to
check that each new test actually notices — it did, all three times, and then I put it back.
It's gone to the review council, which takes about half an hour to come back. It's committed,
because on this repo holding code back isn't actually available: everyone shares one branch,
and the next person who builds ships whatever is there.

The one thing I want to be plain about: this is **not live yet**. The change is Go code, and
Go code does nothing until someone builds a new image and rolls it out. Until that happens
the old behaviour is still what's running. So the bug ticket stays open — the house rule is
that a bug is only closed when the fix is actually running in production, not when it's
written — and I'd rather leave it honestly open than tidy it away into the closed pile.

## 2026-08-02, midday — what the review council said

The council came back **REVISE**, which means "not yet, fix these first". Twelve reviewers
looked at it; eight approved, four objected. That is a good outcome to get, and I want to
explain why rather than treat it as a setback.

**One of them found a real bug in my fix.** The reviewer that reads for code quality noticed
that my new function took the stored path for the favicon and the social card, threw away
everything except the filename, and then rebuilt the path by sticking it under the images
folder. That works for the two we have, because both live there. It would have been silently
wrong for the first one that didn't — and it gave the obvious example: lots of sites put
their favicon at the very top level, `/favicon.ico`, not in an images folder. My own tests
couldn't have caught it, because they only had the two existing cases to test with. I've
fixed it so the stored path is taken whole, and added a test that uses a top-level favicon
specifically. Then I broke the code again on purpose to check the new test actually notices.
It does.

**The other three objections were not about the code — they were about what I showed them.**
This is the part worth remembering. One reviewer said I hadn't checked whether the running
system actually had my change in it. I had — I just hadn't put it in the submission. Another
said there's a known-traps entry about this exact function that I hadn't accounted for. I had
account for it; I'd *rewritten* that entry as part of this work. And the reviewer whose job is
architecture said that although I'd correctly declared this was a big-picture change and
brought it to the right place, declaring it isn't the same as filing the document that records
it — so I've now written that document.

So three of the four objections were "you did the right thing and didn't show me", and those
were graded *more* serious than the one that was an actual defect. That is the right way
round, I think, and it's a lesson I've written into the notes: the council can only review
what you hand it. Evidence you have but don't cite is evidence you don't have.

**Two things I found by answering their questions**, which is the real value of being asked.
First, one reviewer asked whether a future caller could route the social card through the
deploying code and stamp on the real file. Before my change that would have written a file
nobody looks at — harmless litter. After my change it writes to the *real* path, so it would
overwrite the actual social card. Nothing does this today and I measured that, but the nature
of the failure has changed, and that deserves its own ticket rather than a paragraph in mine.
I've filed it as bug 179, along with a second escape hatch in the same code that lets a caller
put a file anywhere it likes with nothing able to predict where.

Second, and I nearly missed it: one of my changes has left a helper function with **no callers
at all**. It's harmless, but a function nobody calls looks exactly like a job someone finished,
and the next person to read it will assume it's load-bearing. I've said so out loud in the
resubmission rather than let someone find it later.

**Still not live.** I checked the running system directly, three different ways to be sure the
check itself was working, and confirmed the change is not in it — the current build predates my
commit. So the ticket stays open. It will go live the next time anyone builds and rolls out the
system, which happens several times a day here, and the verification commands are written down
ready for whoever does it.

## 2026-08-02, early afternoon — the council caught a real one, and I was wrong

The second review round came back **REVISE** again, and this time it was not about
presentation. Two reviewers, independently, said the same thing: my change made it possible
for one part of the system to overwrite a site's real social card and favicon, and the
protection against that was sitting in a *ticket* rather than in the code. I had told them
twice, with measurements attached, that this could not actually happen.

**They were right and I was wrong**, and the way I was wrong is worth writing down because it
looked exactly like being right.

I had measured two things. First, that the check which creates this kind of work item no
longer creates them for the favicon and social card — true, fixed back in July. Second, that
every part of the system that *reads* an image path only ever deals with heroes, icons and
illustrations — also true, and I'd been careful to confirm the query wasn't returning nothing
by accident. Both measurements were sound. **Neither of them could answer the question.** The
first is about work items created from now on. The second is about readers, and the risk was
in a writer.

The thing that answers it is the queue of work already sitting there. One query: **eleven
items** are queued right now asking to deploy the favicon or the social card, and two of them
are in a state that will get picked up and run. They've been there since mid-July, from
*before* the check was fixed. Under the old code they were harmless — they'd have written a
file nobody looks at. Under my change they would have replaced the real thing.

Fixing a check stops new bad items being created. It does nothing about the ones already in
the queue, and nothing here goes back and tidies those up. I'd leaned on "that was fixed in
July" without noticing that the fix was only ever true *of the code that was running in July*
— and I'd just changed the code.

What caught it, in the end, was going to prove the reviewers wrong. I thought they were being
cautious about something I'd already settled, and ran the query to show it. It came back with
eleven rows.

**So I've built the guard rather than filing it.** The deploying code now refuses to touch the
favicon or the social card, and says why — and it refuses *before* it downloads anything or
commits anything, which matters, because a check that runs after the file has already been
written isn't a check. I tested that by deliberately moving the guard to later in the function
to confirm the test notices. It does.

One piece of genuine luck I want to be honest about: none of this is live yet, so the risky
change and its guard will go out together in the same build. If my earlier work had already
been rolled out, this would have been a live incident rather than a review comment.

There was a second, smaller error the same hour. Checking whether a particular override
setting is ever used, I got nine hits — which appeared to contradict something I'd already
told the reviewers twice. All nine turned out to be **my own review submissions**: the system
stores the text of what you submit, and I'd written about that setting at length. The real
answer is zero. The more thoroughly I'd argued the point, the more evidence I appeared to
generate against myself.

Round three is with the reviewers now. Both mistakes are written up in the shared log of wrong
calls, with the cheap check that would have caught each.

## 2026-08-02, evening — approved, live, closed

Three things landed together.

**The reviewers approved it** on the third round, with two remaining comments that were
advisory rather than blocking. I acted on both anyway, because both were checkable in a couple
of minutes and one of them was a genuine miss on my part.

That one is worth recording. I had written that my new "refuse and explain" response followed
the platform's existing convention for a piece of code declining to act. A reviewer asked
whether that was actually true or whether I'd invented a shape. **I'd invented one.** Searching
the whole codebase for the key I'd used turned up exactly one occurrence: mine. The real
convention does exist — it's used by a sibling piece of code in the very same component — and
it's slightly different. I've switched to it and added a test that fails if anyone reintroduces
my version. Small thing, but I'd stated it as fact twice without ever looking.

The other comment found a hole in my own earlier fix: the code now refuses to mishandle a
badly-formed entry, but it refused *silently*, so anyone adding one by mistake would get no
warning at all. Since that list is written into the source rather than configured at runtime,
the right answer is to fail the build rather than log a warning — a warning arrives too late to
help. That's now a test.

**It went live**, on a build made by another session. That's how this repo works: builds come
from committed code, so my work rode out on someone else's release. I checked it properly —
both machines, and crucially by looking for a piece of text my change *deleted*, which should
now read zero. It does, on both. Checking only for something you added proves the search works,
not that your code shipped. Every favicon and social card across all twelve sites still loads
correctly afterwards.

**So the ticket is closed** and moved to the resolved pile. That's the house rule satisfied
honestly: fixed *and* running, not just written.

One last piece of honesty about the timing. The serious problem the reviewers caught in round
two — where my change could have overwritten a site's real social card — never had a window in
which it could actually happen, because the risky part and its guard were still unreleased
together. That was **luck, not planning**. If the first commit had gone out a few hours
earlier, that would have been a live incident on customer sites rather than a comment in a
review.

Three things are left, and none of them are mine to decide alone: whether eleven stale queued
jobs should be redirected rather than simply refused; a second bypass in the same code that I
measured as unused but did not close; and a bigger question I've written up separately about
the system guessing where a file went instead of recording it when it wrote it.

## 2026-08-02, later — the two choices in RFC_010, explained, and the owner's rulings

Appended at the owner's request. The explanation first, then what he decided.

### Why there was anything to decide

A work item is a claim — "this site's social card isn't deployed" — that was true at the moment
it was written. Nothing ever goes back and re-checks it. The danger isn't the stale claim
itself. It's that **what a handler does with that claim can change underneath it.** That is
exactly what nearly bit us: an item that meant "write a harmless junk file" in July meant
"overwrite the live social card" in August, and nothing re-read it in between.

### Decision 1 — how the queue learns a finding stopped being true

**Option 1: let a check say what it saw healthy.** A check gains a way to report "I positively
observed X to be fine", and the runner closes matching items in one place. This is copying what
one existing check already does, promoted from a one-off query into a proper contract. Cost:
one field on a shared structure, one runner change, then each check adopts at its own pace. It
makes retraction *possible* and keeps the closing logic in one place instead of reinvented
fifty times. What it does *not* do is stop a stale item being dispatched — a check only clears
what it actively confirmed healthy.

**Option 2: re-check the item at the moment it's dispatched.** Before a handler acts, re-run
the test that raised the item and drop it if it no longer holds. This is the only option where
a stale item can never be *acted on*, which is the actual harm. It costs more: the checks
aren't written as reusable standalone tests today, and it puts a live check on the dispatch
path.

**Option 3: version-stamp the detectors.** Honest and general, but it depends on a human
remembering to bump a version, and an un-bumped version is silently wrong.

**Option 4: expire anything old.** Simple, and I advised against it — it turns "stale finding"
into "silently dropped real defect", which is worse. Nearly 500 of our open items are over a
fortnight old and nobody knows how many are genuine.

**Option 5: keep repairing by hand.** What I did for the eleven. Doesn't scale to 909, and each
manual repair is a fresh chance to cancel the wrong rows.

The honest summary I gave: option 1 is cheap and makes the problem *tractable*; option 2 is
expensive and makes it *safe*. They compose.

> **OWNER RULING: option 1 now, option 2 later.**

### Decision 2 — what `unresolved` actually means

Smaller, and a live bug rather than a design question. `unresolved` currently sits in a
contradiction: it is not terminal, not deduplicated, not retractable, but is excluded from most
dispatch. So items pile up invisibly — that is why nine of our eleven were duplicate copies of
just two findings. Either it is terminal (and belongs in the dedup and completed indexes) or it
is open (and retraction and dedup must both be able to reach it). It is currently neither,
which is not defensible either way.

> **OWNER RULING: `unresolved` is OPEN** — so retraction and deduplication must both be able to
> reach it.

### The one condition attached to all of it

**Retraction must only ever fire on a positive observation of health, never on "the check found
nothing".** A check that errored, or that was silently blinded by a bug, returns an empty
result that looks identical to a clean site. Getting that backwards would quietly close real
defects across the whole fleet — the opposite failure, and a much more expensive one.

---

## 2026-08-03, later — the retraction mechanism has its first real user

Picking up where the last session left off. The thing it built — the ability for a check to say
"this problem I reported has gone away, close it" — was finished, approved and switched on, and
had never once fired. That was expected: it only works for checks that opt in, and the only check
that had opted in is one nothing runs. So the job was to find a check that should use it.

That turned out to be the whole job, and not for the reason I expected.

**The obvious candidate was a trap.** The natural choice was the check that looks for images that
were generated but never deployed. It has the biggest backlog by far — 95 open items, some going
back to April — and it already works out which images *are* fine and then throws that away, which
is exactly what we need. I was ready to use it.

The problem is that two different parts of the system file that same kind of item, deliberately,
so they share a slot in the queue. One is the check above. The other is the render audit, which
loads a real page in a browser and reports "this image is broken on screen". And here is the bit
that matters: **an image can only be broken on screen if the page refers to it** — which is the
exact condition the first check treats as proof that everything is fine. So if I had let that
check close items, it would have closed every genuine "this image is broken" finding the render
audit ever raised, silently, across every site.

Nothing about that is visible from the code I would have been editing. I only found it by
counting who else files that kind of item before touching anything. I have written it up as a
trap for whoever adopts this next, because we are actively encouraging different parts of the
system to share queue slots — so the situation will come up again.

**So I used a different check**: the one that finds page sections that render empty. It has one
producer, so nobody else's findings are at risk, and it already contains a tested function
answering "does this section actually have content in it?" — written for a different purpose. So
the closing logic reuses the answer we already had rather than inventing a second one that can
drift from the first.

**What it will actually do, measured on the real queue before I wrote any code:** of 47 open
"empty section" items, **17 across 6 sites will close** — most of them sections that were fixed
weeks or months ago and that nothing has been able to tidy up since. The other 30 stay open, and
they are the interesting part:

- 19 are still genuinely empty. Good — it is not just closing everything.
- 10 name a section where **the component has vanished entirely**. That reads like a fix, but it
  reads identically to a page rebuild having silently deleted the thing — which is one of our
  most repeated failures. So it refuses to guess, and leaves them open.
- 1 is a page where three components share one slot, and one of them is still empty. Also left
  alone.

Those last eleven are the point. A lazier version — "close anything I did not find a problem with
this time" — would have closed all eleven on no evidence at all. That is the failure the owner's
condition was written to prevent, and now there is a live number attached to it.

**One thing I checked late and should have checked first.** There is a rule that if the same
problem gets filed three times, the third one is marked "unresolved" and stops being worked. It
counts how many times that item has been closed. Closing an item *counts*, which means this new
tidying-up could, in principle, use up one of those chances and cause a genuinely recurring
problem to be shelved. I measured it: none of the 17 are recent enough to count, so nothing is
used up today. I have flagged it to the reviewers as a live question rather than waving it away,
because the last session was caught out doing exactly that — twice telling the review board
something was unreachable when it was not.

**Where it stands.** The code is committed and submitted for review. It is **not live yet**, and
I am holding off deliberately: pushing a new build restarts the services, and that would kill the
review currently running on this very change. Once the verdict is in, it goes out and I will
measure the first real retraction rather than claiming it works.

## 2026-08-03, evening — it fired, and it left the right things alone

The build went out and the closing mechanism did its first real work. A routine sweep of
leopardessconsulting.co.uk **closed four findings raised in April** — over three months old, and
until today nothing in the platform could close them at all.

The number I care about more: **the same sweep left six of that site's ten open.** Three sections
are still genuinely empty. Two are flagged for a human. One names a section whose component has
disappeared altogether — which looks like a fix, but looks identical to a rebuild having quietly
deleted it, so it refuses to guess and leaves it alone. If it had closed all ten it would have
looked better and been the exact failure you warned against.

Fourteen more will close on the other five sites as they get swept.

Two things I decided as you asked, and both are written down where the next person will find
them. The two-strike question is **accepted as-is and tracked** rather than fixed — it affects
nothing today and changing it would touch the insert path of every work item we have. And I did
not roll the build myself; it went out on someone else's, which is how this estate works.

One thing worth knowing for next time: the review board **cannot see the documentation half of
our work** — it refuses docs, so four seats objected that we hadn't written up a hazard we had
already written up an hour earlier. Cheap fix, and it is in the handoff.

## 2026-08-03, later — a second check can now close its own findings, and the trap we wrote down was not the trap we hit

The closing mechanism has a second user: the check that flags components missing the fields their
own schema says they must have. Same shape as the first, and it drains a queue that had **no way
out at all** — those items have no automatic fixer, so nobody could ever close one except by hand
on the database. Fifty-nine of them are sitting there, the oldest from mid-July.

Six will close on the next sweep of the two sites concerned. That is a modest number and it is the
honest one. The fifty that stay open are the point: they are still genuinely broken, and a
mechanism that closed all fifty-nine would have looked far more impressive and been useless.

**The thing worth your attention.** After the last round we wrote down the trap that nearly bit
us — a check that bails out early when it finds nothing, which switches the new closing behaviour
off on exactly the sites that need it. I followed our own note, looked at the top of this check,
and found no such bail-out. It looked safe.

It was not. The same bail-out is there, just **buried in the middle** rather than sitting at the
top, triggered by a cap that stops the check after 25 complaints so it does not spam a badly-built
site. So it would have switched itself off precisely on the worst sites — and every test would
have passed, because no test has 25 complaints in it. Our own note would not have caught it,
because it told you to look in one place. It now tells you to look at every exit. That correction
is probably worth more than the six items.

I also went looking for guards that only *appear* to be doing something, by deliberately breaking
each one and checking a test complains. Eight of them; seven complained straight away. The eighth
did not — and the reason is subtle enough to be worth a sentence: it was being covered by the
guard standing behind it, so removing it changed nothing *today*. It would have started mattering
the moment anyone rearranged the query. Now it has a test of its own.

Nothing is live yet. It is committed and it will go out on the next build, the same way the last
one did.

## 2026-08-04 — it shipped, it works, and I had already built something we had

The build went out overnight and I verified the new code is genuinely in both running pods. That
part is clean. One thing nearly tripped me: my first check said the change was missing, and the
check was wrong, not the code — the text I was searching for contains a long dash, and the dash
got scrambled on its way into the pod. Worth knowing because a wrong answer there looks exactly
like a failed deploy.

**Then the bad news, and it is worth your attention because it is the kind of mistake that is
expensive to notice.**

I built this on the argument that these particular flagged items had *no way of ever being
closed* — no automatic fixer, so a human would have to go into the database by hand. That was the
whole reason to do the work. It is not true. We already have a thing that closes them, it has
been running since the 27th of July, it uses the same test I wrote, and it reaches **all** of
them, not some. In two respects it is more careful than what I wrote.

What caught it was re-measuring after the deploy. The six items I expected to close had already
closed — **two hours before my code was even running** — by that other mechanism. If my code had
gone out a day earlier it would have closed them first, I would have reported six successes, and
nobody would ever have looked.

I want to be straight about how this got past everything. I did check. The handoff I was working
from *requires* a specific check before doing this kind of work, I ran it thoroughly and two
different ways, and it passed — because that check looks at what **creates** these items and the
problem was in what **closes** them. The review board approved it too; one of the reviewers
specifically praised the claim as unusually well evidenced. It was well evidenced. It just proved
a slightly narrower thing than what I then wrote down. That gap — between what I proved and what
I claimed — is the actual error, and it is a small one that did a lot of work.

The code is harmless: the two mechanisms cannot fight, and nothing is at risk sitting where it is.
But it is duplicate machinery, and duplicate machinery that drifts apart is precisely the thing
this project keeps getting bitten by.

**So there is a decision for you, and I have deliberately not made it.** There is one narrow case
the older mechanism does not cover and mine does — it is real but nothing is in it today. The
options are: teach the older, better mechanism that one extra case and delete mine (what I would
do); keep both and accept the duplication; or just delete mine and leave the gap for the
follow-up we already have open. It is written up with the costs in the handoff.

I have logged the mistake in the fleet-wide log of wrong calls, with the ten-second query that
would have caught it before I wrote anything, and added the missing step to the standing checklist
so the next person doing this is told to look at both ends, not one.

---

**2026-08-06, later the same day (new session).** I picked this up cold from the handoff, which
said the lane was finished except for one inherited problem: the daily queue sweep was doing 94%
wasted work. It is now fixed, twice over — once immediately, once properly.

First, what the handoff had not quite worked out. It knew the sweep was re-reading the same rows
every day and wasting most of its effort. What it had not asked was whether any of the rows it
never got to were rows it could actually have *done something about*. They were. **Sixty-four of
them** — thirty-eight percent of all the work this sweep exists to do. And not "it gets to them
eventually": never. The queue is read oldest-first, and the front of the queue is packed with items
the sweep has no way to judge and which therefore never leave. They sit there permanently, and
anything behind them is unreachable.

The detail I find most telling: the items being starved were the **newest** ones. The sweep reads
oldest-first on the reasoning that old findings are the most likely to be out of date. But a
finding raised last week is exactly the one a recent rebuild has probably already fixed — so the
rows most likely to be closable were the ones guaranteed never to be looked at. The design was
working against its own logic.

I fixed it in two steps deliberately. The first is a config change, live within the hour: raise the
number of rows the sweep will look at in one pass, from 500 to 1500. That alone un-starved all
sixty-four immediately, needs no software release, and can be undone with one line. But it only
buys time — the queue grew by seven rows while I was working on it, and it has grown two hundred in
the last three days — so a limit sized to today's queue would be back in the same state within a
week or two.

The second is the real fix, and it is now written, reviewed and committed: the sweep should only
ever load the kinds of item it is actually able to judge. Then its limit applies to useful work, and
a pile-up of some unrelated kind of item can no longer crowd out the work it is supposed to be
doing. It is wired so that the list of "kinds I can judge" is taken directly from the list of
"kinds I know how to judge" — they are the same list, so they cannot drift apart as more are added.
That part needs a software release before it does anything; the config change is protecting the
queue in the meantime.

**One thing I have to flag, because it is uncomfortable and it is the more useful half of this.**
The handoff recommended a specific way of fixing this — set up several scheduled jobs, one per kind
of item. That would not have worked. The setting it relies on is read from a different place than
the schedule can reach, which is a trap the very same page of the handoff had documented, two
paragraphs earlier, after paying to learn it. The diagnosis was right and the prescription walked
straight back into the hole. I have logged it in the fleet-wide log of wrong calls, because the
lesson generalises: **a trap you have just written down is not thereby disarmed for your next
paragraph**, and the recommendation half of a handoff gets far less scrutiny than the evidence half
despite being the part the next person actually does.

The review board approved it first time, with four advisory notes and none serious. Four of them
were things I could go and check rather than argue about, so I did. Two turned up facts worth
having: the scheduled job **is** genuinely switched on (worth confirming — a fix shipped onto a
dead schedule does nothing), and **no existing caller** can be affected by the one behaviour change
I was least sure about. One note I could not fully answer and have written down as unresolved
rather than dressed up: the new warning I added, for when the sweep runs out of room again, is only
useful if somebody reads it — and logs here rotate within minutes. I have put the query in the
handoff, but that still depends on a person running it. It is better than what was there before,
which was a number sitting in a data blob that nobody looked at for a fortnight. It is not a
solved problem.

Nothing here needs a decision from you. The next session's jobs are: confirm the release picked the
change up, and check the first scheduled run afterwards reads the whole judgeable queue instead of
the first five hundred rows.
