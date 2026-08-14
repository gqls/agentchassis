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

**2026-08-06, an hour later — it is live, and it worked on the first pass.**

A new build went out (not mine — someone else's release that happened to carry my commit), so I
checked whether the change was actually in it rather than trusting the version number, and it is,
on both machines. Then I ran the sweep by hand to see what it would actually do.

**It closed twenty items, and every single one was a row it could not have reached yesterday.**
They were all raised between the 3rd and the 5th of August — the young end of the queue, which is
precisely the part that was being starved. The last run before the fix closed nothing at all. The
whole queue of judgeable work is now 168 items and the sweep looked at all 168, rather than the
first five hundred rows of a mostly-unjudgeable pile.

It also, for the first time, reported the real size of the problem it *cannot* solve: **611 parked
items are of kinds nothing knows how to re-check.** That number was always true and the old code
was structurally incapable of printing it — it could only count the ones that happened to fall
inside its batch, so the gap looked smallest exactly when it was worst. It is not a new problem and
I have not tried to fix it; closing it means teaching the sweep more kinds of item, one at a time.
But it is now visible every single run instead of invisible.

**And the one loose end from this morning turned out to be the same bug wearing a disguise.** There
was a row that looked like it was being individually skipped — it had no record of ever being
checked while its neighbours did — and the earlier note suspected a naming inconsistency. It
wasn't. That row was simply too new to be reached, and the neighbours it was compared against were
two weeks older. It got checked and closed at 10:03 this morning with the "inconsistency"
untouched. The lesson I have written down is that from a single row you genuinely cannot tell
"skipped" from "never looked at", and I nearly went hunting for a fault in the wrong place.

**Nothing in this lane is open now.** The only thing worth a glance is tomorrow morning's automatic
run, which should do the same thing unattended.

**2026-08-08 — it has been running itself for two days, and that was the thing left to prove.**

When I finished on the 6th I said the only thing worth a glance was whether the next morning's
automatic run would do the same job with nobody watching. It did, and so did this morning's. One
item closed on the 7th, three today. That is the right shape and I want to be clear about why,
because the numbers look like a collapse from the twenty I closed by hand on the 6th: those twenty
were a fortnight's worth of items that had piled up unreachable, so clearing them was a one-off
flush. One and three are the sweep simply keeping up. The number that would signal trouble is not
how many it closes — it is a flag saying it ran out of room, and that has stayed clear.

I also re-checked that the change is still in the running software, because five new builds have
gone out since and **none of them were mine** — my change has just been carried along each time.
It is still there on both machines. That check is worth repeating after any release, since nothing
in a version number tells you whose work it contains.

One thing that vindicates the paperwork more than I expected: the database only keeps about a day
of run history, so **the record of the run that proved this — the one that closed twenty items —
has already been deleted.** The figures now exist only because they were written down at the time.
If I had planned to "check the numbers later", there would be nothing to check.

**What is genuinely left, stated plainly, and I got this wrong once already.** On the 6th I wrote
"nothing open" and had to correct myself an hour later: I had closed the problem this lane was
reopened for, not the lane. Two things remain, both known and neither urgent. One is a
de-duplication change blocked behind a set of duplicate rows — and when I re-measured that blocker
I made a second mistake worth recording, guessing at a database filter from memory instead of
reading it, which produced a number nearly 50% too high that I almost wrote down as growth. Read
properly, it has grown from 48 to 53. The other is a dormant tripwire in a related check that will
matter to whoever next extends it, and is harmless until then.

The bigger honest number is 625: that many parked items are of kinds nothing knows how to re-check.
Not new, not mine, and now visible on every single run instead of invisible — which is what has to
be true before anyone can sensibly work on it.

There is a fresh cold-start handoff (`HANDOFF_2026-08-08_continue_here.md`) so a new session can
pick this up without reading the whole history.

---

**2026-08-08, afternoon.**

Picked the lane up cold. First thing was to check the sweep we fixed on Thursday was still
running on its own, because the chassis had been rebuilt twice since anyone looked. It is: the
daily job fired at 08:38, looked at the right 151 rows out of 773, closed 3, and reported the cap
as not binding — which is the number that matters, because if it ever says otherwise the old
starvation bug is back. Three closures on Friday, one on Thursday, both with nobody watching.

Then I picked up the "teach the sweep another item type" thread. The previous handoff named three
candidates. **Two of them were wrong, and the check that showed it took about ten seconds.**
`content_rewrite` looked like the biggest prize at 34 parked items — but 51 of them have already
been closed by a real fix pipeline that rewrites the prose and redeploys the page. It doesn't need
us. `needs_sprite_css` has a re-run path of its own that I haven't traced yet, so I left it.

The one that was genuinely stuck was `voice_tells` — the check that flags copy reading like a
machine wrote it. **Nothing has ever closed one of those. Not once.** Twenty-five of them, all
filed on 17 July, all on the leopardess site, still sitting there twenty-two days later. They were
never going to move, because the check deliberately files them for a human and there is no human
surface that drains them.

There was one thing that looked like it might stop me, and it's worth writing down because I
nearly took it at face value. The check's own code says the fix must be *"never an unreviewed
auto-rewrite"*, and another bug file quotes that line approvingly. Read quickly, that sounds like
"don't automate this". But it's about *fixing* the copy — rewriting it without a human. What I've
built doesn't rewrite anything. It re-reads the page and asks whether the complaint is still true;
if the words have since been changed and now read fine, it stops making the complaint. The human
still decides how to fix anything that genuinely reads badly. I checked both bug files that count
these items and neither of them puts the type under an open decision, so I wasn't stepping on
anyone.

The interesting part of the build was a trap I want to flag, because it's the kind that looks like
success. "The page has no problems" and "we couldn't read the page" produce **exactly the same
empty result**. If the page was deleted, or was never published, or every part of it is pinned by
a human, the scan comes back with nothing found — which reads identically to "somebody fixed the
prose". If I'd taken that at face value, the thing would have quietly closed live complaints on the
strength of having read nothing at all. So it now counts what it actually examined, and refuses to
close anything unless it genuinely read something.

I also checked whether this would do anything at all before building it — 13 of the 25 pages have
been edited since their complaint was filed, so there is real change to judge. If that had come
back zero I'd have said so and dropped it rather than shipping something that adds cost and closes
nothing.

Two honest caveats. The first is that the "page is pinned" and "page is gone" cases don't exist on
today's data, so that part is careful reasoning and unit tests, not something I've watched work.
The second is more interesting: the standard being measured can move. If a site relaxes its own
voice rules, complaints will close even though nobody touched the words. That's arguably right —
the site changed its mind about what good looks like — but it means a closed item isn't proof the
copy improved, and I've written that down in two places so nobody reads more into it later.

The code is committed and has gone to the review council; the verdict hadn't come back when I
wrote this. It won't do anything until the next chassis rebuild. I took the "before" measurement
first, which is the only way to prove it shipped afterwards — and my first attempt at that
measurement was wrong in a way worth admitting: I searched for too short a phrase and got six
matches on a build that doesn't contain my change at all. A short search term is somebody else's
words. Fixed, re-taken, and the real "before" reading is zero on both machines.

Still blocked, and getting slowly worse: the duplicate-rows cleanup. 55 clashing pairs today
against 53 this morning and 48 five days ago. It drifts up about two a day and still needs a
judgement call from you — which copy to keep, and whether throwing the others away loses anything
real.

---

**2026-08-09 — the retraction sweep closed its first item on its own, and one of my own
instructions nearly hid it**

The thing we built on Friday works. At 08:38 this morning the sweep ran on its schedule, with
nobody watching, and it retracted a finding on the leopardess site — a page called
`ai-readiness-quiz` that had been flagged in July as reading machine-written. The copy has since
been rewritten, the sweep re-read it against the site's own standard, found nothing wrong any
more, and closed the item. That is the first time anything has ever closed one of those; there
were 32 of them and they had been sitting there since July with no route out. All 32 got looked at
this morning. One was clean enough to close, the rest stay open, which is exactly right.

Here is the awkward part, and I want it on the record because it is the more useful half. The
handoff I left on Friday told this morning's session how to check whether it had worked: watch a
number called the "uncovered backlog" fall by about 32. I checked. It had not moved — 625 before,
625 after. If I had trusted my own instruction I would have written down "the change did nothing"
on the morning it worked perfectly. The number is a total across about forty different categories
of parked work, and while our 32 dropped out of it, other categories grew by exactly 32 in the
same few hours. The total was never capable of answering the question. The right check is to look
at the per-category breakdown and confirm ours has vanished from it entirely, which it has. I have
written that down as a trap so the next person adopting a category cannot inherit the bad recipe.

Second correction, and this one is about pressure I put on you. I have been telling you the
duplicate-rows problem is getting steadily worse — "about two more clashes a day, so it gets more
expensive the longer it waits". I re-measured it properly this morning and it had gone **down**,
from 55 clashes to 47. Looking at all four measurements together they bounce around rather than
climb. The measurement was right each time; the *trend* I drew through them was not, and it was
doing real work in how I described the urgency to you. The decision still needs you — which of
each duplicate pair to keep, and whether throwing the others away loses a real finding — but it is
not a clock running down.

Today's actual work: I taught the sweep a fifth category, the one that flags factual claims a page
makes that the site's evidence register does not support. Twenty-three of those across seven
sites, and, same as before, nothing had ever closed a single one. Before writing any code I
checked the thing that would have killed the idea — whether any of those pages had actually
changed since being flagged, because if none had, the sweep would just run and find the same
problems for ever. Sixteen of the twenty-three had changed. It has gone to the review council and
is committed; it will start working at the next deploy.

One thing I found and deliberately did **not** act on. There is a whole class of parked work the
sweep structurally cannot see — 467 items sitting in statuses it never looks at — and, more
awkwardly, the report we use to measure "how much work is left uncovered" is filtered the same
way, so it cannot see them either. It has been telling us 625 when the real number is closer to
1,100. That is a claim about how a shared piece of machinery is built rather than about one bug,
and our own rules say I should not assert something like that on my own reading. So I have put it
through the diagnosis loop and will act on what comes back. It may well come back saying the
scoping is deliberate and correct — that would be a good outcome, not a wasted run.

**2026-08-09 (later) — the review council pushed back, it was right, and there is one question
only you can answer**

Two things came back from the review board on this morning's work. I'll take the easy one first.

They caught me getting a fact wrong. I had described the thing I was building as covering a case
where *two* different checks file the same kind of work item, and I leaned on one of your rulings
from the 2nd — the one saying that situation doesn't need a full architecture review as long as
both checks are named in the register. It turns out there is only **one** check. The second file I
named isn't a check at all; it's a helper the first one calls. So the ruling I invoked never
applied to this change in the first place. The work is fine — actually simpler than I described
it, and the reviewer's sharper question ("are you sure both halves are re-checked the same way?")
has a good answer, because there is only one scan and both halves are inside it. But I asserted
something false, in the register, which is the file other reviewers treat as ground truth. I've
corrected it in all five places it had spread to, visibly rather than quietly.

Now the one for you. Three of the fifteen reviewers, working from completely different remits,
independently said the same thing and all three said it should go to a human rather than be
decided by them:

> This item type was made human-only **on purpose**, because it is about factual claims — whether
> a page is asserting something the site cannot back up. Letting a scheduled job close those
> unattended is a policy change, not a bug fix.

The sharpest version of it: **the evidence register proves provenance, not correctness.** My
machine can confirm that a number on a page now appears in the site's register of facts. It cannot
confirm the fact is *true*. So if someone adds a sloppy or wrong entry to that register, my sweep
will quietly retract a live claims-integrity finding and no human will ever look at it.

I think that concern is real and I have not tried to argue my way around it. It is genuinely
different from the voice one we shipped yesterday: that one closes findings about *tone*, and the
worst case is that some slightly stilted prose stops being flagged. This one closes findings about
*truth*, and the worst case is that an unsupported claim stays on a customer's website with the
warning switched off.

The reviewers asked for explicit sign-off from the owners of the two bug files that track this
work, rather than the notification I'd already sent them. So the question for you is simply:

**do you want a machine closing factual-claim review items at all, and if so, under what
condition?** Some options, cheapest first — I have not built any of them, I want your call before
I do. (a) Ship as is: it only ever closes when it can positively re-verify, and it refuses on
anything ambiguous. (b) Let it close only when the page's copy has actually changed since the item
was filed, so a register edit alone can never retract a finding — that is about four lines and it
kills the whole failure mode above. (c) Don't close at all for this type; downgrade the finding
and leave it for a person. (d) Leave it human-only and drop the change.

For what it's worth, (b) is what I'd suggest. It costs almost nothing and it turns the reviewers'
objection from a policy argument into a mechanical guarantee.

The code is committed either way — on this tree, committing *is* shipping, and I can't hold it
back. But it does nothing until the next deploy, so there is time.

**2026-08-09 (evening) — your call is in, and it's built**

You picked (b), so that's what's there now: the sweep will only close one of these findings if the
page's own text has actually been edited since the warning was raised. Someone adding an entry to
the approved-facts register can no longer make a warning disappear on its own. It's about a dozen
lines in the end, and it turns the reviewers' objection from something you have to trust us about
into something the code simply won't do.

Two things worth knowing about it.

It cost us something real, and I'd rather say so than let you find it later: the gate is stricter
than "did anyone fix this". If a page genuinely got corrected in a way that didn't leave a
timestamp on the component, the sweep will now decline to close it and leave it for a person. That
is the right way round — declining wastes a glance, closing wrongly hides a false claim on a
customer's site — but it does mean this category will drain more slowly than the 23 open items
suggest.

And I made a mistake building it that I want on the record, because it's the exact mistake the
change exists to prevent. I wrote the comment explaining the guard before I wrote the guard, and
the comment said the code would refuse when it didn't know when a warning was raised. It didn't —
it would have closed on any edit at all, however old, which is precisely backwards. A test I'd
written for that case failed on the first run and found it. So: a confidently-worded comment
describing behaviour that wasn't there, inside a change whose whole point is that comments aren't
controls. Fixed, and the test now pins all four ways that gate can go wrong.

I did **not** apply the same gate to the voice/tone category we shipped yesterday. It has the same
hole in principle — someone loosening a site's tone settings would retract warnings without the
copy changing — but it's already live and reviewed, and the reviewers' whole argument was that tone
is a lower-stakes surface than truth. Changing it is a separate decision, so I've written it down
rather than quietly doing it.

Your "date the pages when last checked" idea is captured as `features_open/031`. It's a genuinely
different question from the one I just solved, and a good one: my gate proves the *page* moved, but
nothing anywhere records whether we ever *looked*. Right now an empty review queue means either
"everything is fine" or "nobody checked" and we cannot tell which — which is the same ambiguity, one
level up, that we spent this week killing inside individual page scans. I've written up why it
matters, the open design questions, and what it is not (it's not a content-freshness feature — we
have two of those already).

The third review round is running now.

**2026-08-10 — the new build carries everything, and the honest score so far**

The fresh chassis went out this afternoon and I checked it rather than assuming: both machines are
running the new code, all four of the markers I look for are present, and the control marker that
proves the check itself works is there too. Nothing was lost in the roll.

Where we actually are. Both new categories are live and closing real work — this morning's run
looked at 243 findings and closed 37, eight of them in the factual-claims category you ruled on.
Every one of those eight was a page whose text had genuinely been edited since the warning was
raised, spread across ten days and several sites, so these are real fixes rather than one rebuild
clearing the board.

But I want to be straight about your gate: **it has held every time and it has never once stopped
anything.** Zero refusals so far. On the pages we've seen, every one that scanned clean had also
been edited, so the gate hasn't yet had to do its job. It's proven to be there and proven correct
where it applied — it is not yet proven to bite. I've written it up that way rather than letting it
read as though it saved us from something.

Two reviewers pushed on a real weakness in it, and they're right. The date we compare is for the
whole section of a page, not for the specific sentence that was flagged. So an unrelated tweak to
the same section technically satisfies the gate. I checked the obvious way that could have gone
wrong — everything closing at once because of a bulk rebuild — and that isn't what happened here.
The weakness stands though, and tightening it so we compare the actual flagged sentence is the best
next job. The finding already records what it objected to; nothing currently reads that back.

One thing I found today that's worth telling you, because it makes the case for your ruling much
stronger than I could. Another team wrote down, back on 31 July, exactly the concern the reviewers
raised — and with a live example. Their note says a registered fact makes the whole claims check
meaningless as evidence of truth, because **the register is also the list we hand the writer**. So a
false entry is self-ratifying: the platform tells the writer to state it, then vouches for it. Their
example is gamesdesign, which claims "10,000 Monte Carlo trials per query" about a tool whose code
contains no randomness whatsoever — and every check passes it, correctly. That's precisely the hole
your gate closes, described independently before we got here.

The review board has now said revise four times. Each time it was right, and I'd rather report that
plainly than round it up: once it caught a factual error of mine, once it insisted the decision was
yours, and twice it caught me describing my own work less carefully than I'd done it. The latest was
that I'd folded one change into another one's description, so a reviewer reading only my summary
correctly concluded the whole thing couldn't work. The code was fine; my paperwork wasn't. Round
four is running now with that filed properly.

Everything is written down for a fresh start: `HANDOFF_2026-08-10_continue_here.md`.

---

## 2026-08-10, later — round four never got a verdict, and finding out why turned into something much bigger

Round four came back neither approved nor revised. It came back dead. The run reached a state
called `complete_invalid`, which reads like the board rejecting my submission, and it isn't that
at all — my submission was fine and had been accepted. The actual message was buried one level
down: our Anthropic account has **hit its monthly spending limit**, and the API says access
returns on **1 September**. That's three weeks away.

So nothing that needs an AI model is working right now, anywhere on the fleet. Not the review
board, not the diagnosis tool we use to check our own theories, not the content writers. The last
successful call was 15:51 our time and everything since has been refused. This is not something a
thread can wait out or work around — **it's a billing limit only you can raise**, so it needs your
decision rather than my patience.

One piece of good news before the rest: **the work this lane actually shipped is unaffected and
still running.** The daily sweep that closes stale review items is plain Go and database queries
with no AI in it, so it ran this morning as usual and will keep running through the outage. The
outage stops us *reviewing* new work; it doesn't stop the automation we already built.

Then I went looking for where the month's budget had actually gone, and the answer was not what I
expected. **The review board is 88% of everything the whole fleet spends on AI.** Not the site
builders, not the content writers — the board we consult before making platform changes. Over the
first ten days of August it used 165 million words' worth of input out of 188 million for
everything combined.

The reason is worth explaining, because it's a genuine mistake in how we built it rather than the
board simply being expensive. Every review sends the submission to fifteen independent reviewers.
Each reviewer gets the whole thing: the database schema, the plan, all the evidence — about
270,000 characters. I compared three of those fifteen messages character by character, and
**98.6% of them is identical**. We are paying to send the same text fifteen times.

The fix for that is a standard feature called caching: send the shared part once, and the other
fourteen reviewers read it at a tenth of the price. We don't use it anywhere. But there's a second
problem underneath, and it's the one that makes this a real bug rather than a missed setting.
Caching only works on the *beginning* of a message. We put the bit that differs per reviewer at
the top and the enormous identical bit underneath — so the messages start differing at character
21, and caching would find nothing to reuse **even if we switched it on**. Both have to be fixed
together or neither pays.

By my arithmetic that's about a **76% cut** in what the board costs. I want to be honest that this
is arithmetic on measured numbers, not an observed bill — and that our logging can't currently
even tell us whether caching is working, so building that measurement in is part of the job, not a
follow-up to it.

I've written it up as `bugs_open/244`. Normally I'd put a claim this size through the diagnosis
tool before asserting it — that's our own rule — but that tool is one of the things the outage has
killed, so I did the verification by hand instead and said so plainly in the file.

One thing I got wrong along the way, worth recording: I first tried to measure how widespread the
outage was by searching the server logs, got zero results across four services, and nearly read
that as "it's only affecting me." Then I checked how far back those logs actually go: about two
minutes. The silence meant nothing. Another session had independently hit the same wall from a
completely separate service outside our cluster, which is what actually proves it's the account
and not something we broke — I've pointed at their findings rather than redoing them.

**What I need from you:** whether to raise the limit. If the answer is yes, everything resumes; if
it's no, the fleet does no AI work until 1 September and we should plan around that. Either way
the caching fix is worth doing, because it's the difference between the budget lasting ten days
and lasting most of a month.

---

## 2026-08-11 — correction: the three-week outage I told you to plan around lasted three hours

Yesterday I wrote that the fleet was out of AI credit until 1 September and asked you to decide
whether to raise the limit. **You raised it, and I should correct the record: the outage ran from
about 15:51 to 19:12 our time — roughly three and a quarter hours, not three weeks.**

The mistake was mine and worth naming precisely, because it's the sort I could repeat. The error
message stated a reset date of 1 September, and I reported that date as though it were a forecast
of when we'd be working again. It isn't. It's the vendor's worst case — what happens if nobody
touches the limit. The thing that actually decides the outcome is whether a human raises the cap,
and on this estate that took hours. I've written the check into the shared landmine file so the
next session confirms the *successes have resumed* rather than reasoning from the failures.

**What that does not change is the underlying problem, and I'd rather not let the quick fix bury
it.** The budget was exhausted on the *tenth* of the month. Nothing about raising the limit stops
that happening again around the same point next month — it just moves the wall. The review board
is still 88% of our AI spend, still sending fifteen reviewers a message that's 98.6% identical,
still with no caching and still ordered so caching couldn't work anyway. That fix is worth roughly
76% of the board's cost and it's the difference between a budget that lasts ten days and one that
lasts the month. It's filed as `bugs_open/244` and it stays open.

**Where the lane itself stands, all re-checked this morning against the new build:**

The fresh chassis (v1.0.1284) went out at 10:23 and I confirmed our code is actually in it rather
than assuming — all three of our markers present on both machines, with a control that correctly
finds nothing, so the check can tell the difference. The overnight sweep ran at 09:44 as usual.

**Your copy-changed gate has still never fired.** Eight items closed, nineteen still standing,
three it couldn't judge — and every one of the eight had genuinely edited copy behind it. So the
gate is holding the line without having had to block anything yet. I'll keep reporting it as
unproven rather than as working, because those are different claims.

The one thing outstanding is the review board round that died mid-flight yesterday. The board is
available again, so I can resubmit it — but a single round now costs about 1.6 million words of
input, which is exactly the spending I've just finished documenting. Given the cap was hit
yesterday, I'd rather you told me to fire it than assume.

---

## 2026-08-11, afternoon — you stopped me rebuilding something that already existed, and the review board is running again

Two things happened today worth writing down, and the first is a mistake you caught.

**I was about to build the caching fix, and it had already been built.** You asked me to check
first. It turns out another session shipped it on Monday evening — about two hours after I filed
the report. Both halves: the change to how we talk to the AI provider, and the reordering of the
review board's fifteen messages so the shared part comes first. It even went through the review
board itself on the way, and the board caught a real bug in it.

**Why I missed it is worth naming, because it's a habit rather than an accident.** I searched the
code for caching on Monday, found none, and then acted on that answer on Tuesday. This tree gets
something like fifteen hundred changes a week. A search only tells you what was true at the moment
you ran it, and I treated it as a standing fact. The specific thing I should have done — and
didn't — was look at the history of the one file I was about to edit. That would have shown both
commits sitting at the top of it.

**The good news is that it works, and now we can measure it rather than estimate.** Before the
fix, a review round cost about 806,000 words of input at full price. Now it costs about 128,000 at
full price, with the rest served from cache. That's roughly **58% cheaper per round**, and about
69% cheaper per unit of text. Nine out of ten reviewers are reading from cache rather than paying
full price.

**Two things I told you were wrong, and I'd rather correct them than let them stand.** I estimated
the saving at 76%; it's 58%. And I recommended we set the cache to hold for an hour, arguing the
default five minutes would expire during a long round. The data says the opposite — reviewers
arriving *after* five minutes hit the cache *more* often, not less, because each read keeps it
alive. The person who built it left the default alone and wrote a note saying they'd only change
it if measurement justified it. That measurement has now been done, and they were right.

**What's still genuinely unfinished:** only the review board uses this. Nothing else does. The
board was 88% of our spend so it was the right place to start, but the content writers and others
are still paying full price. That's the remaining piece.

**And the review board round is running.** You asked me to resubmit, so I have — unchanged, on the
new build that went out at lunchtime, after checking our own code is genuinely in it rather than
assuming. It started immediately instead of queueing. This is the fifth attempt at this review:
three came back asking for revisions and were right each time, and the fourth died mid-flight when
the account hit its limit. I'll report what this one says without rounding it up.

### The fifth attempt is ready to go, and I'd like a decision before I spend it

The review board came back again asking for changes — that's four in a row. But this one is
different from the others in a way that's worth a minute, because it says something good about the
board rather than about us.

Both of its main complaints were, in effect, **"you're telling me things I have no way to check."**
Not "you're wrong" — "I can't see it from here." One reviewer pointed out that the owner ruling we
keep citing lives in a markdown file it can't read, and that if we'd simply invented that ruling,
the whole justification collapses. It then named exactly where it *would* accept the evidence. Same
with our habit of citing "round 1 said this, round 3 said that" — it can't see our previous rounds
either, and asked where the real record is.

Both records existed. We'd just never handed them over. The ruling shows up in nine places in the
shared notes database; all four previous rounds are sitting in the board's own results table with
their verdicts. **So the fix wasn't to argue — it was to cite.** That's the fourth time this lane
has been caught describing the work less carefully than we actually did it.

**The code has not changed.** Every reviewer with a view on the design approved it rounds ago. What
changed is five pieces of bookkeeping: hand over those two records, properly list a test we'd
claimed as a safeguard but never actually filed, fix an off-by-one where we said "edit 9" in a
plan with eight, move a warning about blast radius out of a code comment and into the risks
section where a reviewer will see it, and re-do one measurement with a better method.

**Three things I found while doing it, two of which are corrections to us.**

The first: I nearly made the exact mistake I was in the middle of fixing. Writing up the test, I
copied a function's shape from elsewhere in our own plan instead of opening the actual code — and
got the arguments in the wrong order. Nothing in the plan was wrong; I'd have *introduced* a new
sloppiness inside the very edit whose job was to answer a reviewer about sloppiness. I caught it by
reading the real code. That's the whole lesson of this lane in one paragraph.

The second: we've been proving our code is live by checking "both copies of the service". There
aren't two — there are **41**, spread across twenty-odd different jobs that all run the same
program. The reviewer caught that, and it was a false claim of completeness on our part. The
better proof is cheaper as well as stronger: ask what *fingerprint* each running copy has. All 41
report the identical one, so checking any single copy now genuinely tells you about all of them. I
re-ran the check that way and our code is in there.

The third: while answering an unrelated complaint I stumbled over a **false statement in our own
handover notes** — twice carried forward. We'd written that a certain search index can't see a
certain kind of thing, and used that as the reason not to spend money on an investigation. It can
see them; there are 700 of them indexed, including the exact four we said were invisible. Nobody
has been harmed, because the false bit was an argument for *not* spending. But it's the kind of
error that survives precisely because it sounds like diligence: we'd checked it first-hand, just in
a place that could never have proved us wrong. It's written up in the shared mistakes log.

**Where that leaves us.** The revised submission is finished, checked against every validation the
tool applies, and committed. I have **not** fired it. A round is real money, and although the
caching fix that landed yesterday makes it roughly 58% cheaper than the figure we were worried
about, it's still your call rather than mine — the last time this lane fired one without asking was
the day after the budget blew. Say the word and it goes.

### It came back "revise" again — but this one is close, and there's a real decision in it

Fifth round, fifth revise. I know how that reads. But the shape has changed completely and I don't
want the headline to hide it: **fourteen of the sixteen reviewers approved.** Two objected. Last
round it was the reviewers' inability to check us that sank it; that part is fixed and they said so.

**The two we fixed stayed fixed.** The reviewer who blocked us last time — the one who couldn't
verify our claims — approved, and said the records we handed over are "the kind of thing I'd want a
future round to keep citing verbatim". The one who told us off for using a crude search instead of
the proper tool approved too, and called our substitute "the right substitute". A third went out of
its way to endorse the deploy check I rewrote this afternoon. So the work was the right work.

**What's left is two things, and they're different in kind.**

The first is embarrassing but easy, and it's ours. Two reviewers independently noticed that our
submission describes eight changes as though we're *about to make them*, while simultaneously
offering proof that they're *already running in production*. One of them put it plainly: if the code
is already shipped, this edit is pointless; if it isn't, your evidence is impossible. I checked, and
the code is indeed already live — it went out on the 9th. **Both of their conclusions are wrong, but
only because the true answer isn't written anywhere in the form we submit.** Our own rules say
reviews here happen *after* the change ships — that's deliberate, because on this shared setup you
genuinely cannot hold code back. We just never said so on the form. Five rounds, and no reviewer has
ever been told which of these edits already exist. That's one sentence, and it should have been
there since round one.

I also made a small new mess while cleaning up an old one. I'd caught myself getting a function's
arguments in the wrong order, went and read the real code, fixed my bit — and left the neighbouring
bit inconsistent with it. Same fault as the last four rounds, one level down. The reviewer spotted
it. It's minor, but I'd rather write it down than let it look like a clean round.

**The second is not a mistake, and it's the one I need you for.** The reviewer who blocked us is
raising, for the third time, that our safeguard checks the wrong thing. You signed off on a gate that
only closes a flagged claim if the page's copy has actually changed since we raised it. What it
actually verifies is that *something on that page moved* — not that *the specific claim we flagged*
was dealt with. So a typo fix, or a style tweak, could in principle satisfy it. The reviewer is
explicit that it isn't blocking and hasn't seen this go wrong — nothing in the live data shows it —
but says on a content-integrity check this sensitive, the gap deserves naming outright rather than
being tucked into a "we'll do it next" line. And it offered the fix: compare the actual flagged
wording, not the page's timestamp.

That is exactly the job we'd already listed as next. So the question is really about order:

- **Fix the paperwork and go again** — say plainly that the code is already live, tidy the
  inconsistency, and name the gap explicitly. Cheap, and it's the round most likely to pass.
- **Build the tighter check first, then submit.** More work, but it *answers* the objection instead
  of documenting it — and it's work we've already agreed is worth doing.
- **Stop submitting.** The code has been live and working for two days, every reviewer with a view on
  the design approved it rounds ago, and the one open objection is a policy judgement you have
  already ruled on once.

My honest recommendation is the second, then the first — build the tighter gate, and let one final
round carry both. The objection has now been raised three times by a seat that cannot veto, which
usually means it's right and being politely persistent.

### Built the tighter check you asked for — and the reviewer's own suggested fix turned out to be the wrong one

You picked "build it, then submit once". Done, committed, and round six is with the reviewers now.

**What it does in plain terms.** Before, we'd close a flagged claim if the page was clean *and*
something on that page had been edited since we raised it. Now we also require that **the actual
words we objected to have gone from the actual box they were in**. If we flagged "90,790 customers"
in the hero, and the hero still says "90,790 customers", the item stays open no matter how much else
on the page moved.

**The interesting part is that the reviewer's proposed fix would have made things worse, and I could
only tell by measuring.** It suggested comparing the surrounding *snippet* of text we'd recorded.
Sounds stricter. So before building it I tested both options against the cases where we already know
the answer — the items where the claim is definitely still on the page. The snippet test spotted the
claim **18 times out of 41**. The short-token test spotted it **40 times out of 41**.

That gap matters enormously because of which way the failure falls. A snippet is a long piece of
text, so any small change to spacing or markup breaks the match — and a broken match would read as
"the copy changed", which *grants* the closure. In other words the reviewer's version would have
waved through more than half of all claims. **It was right about the problem and wrong about the
cure, and the cure is the half you can actually test.** That's now written into the code comments so
nobody "improves" it back.

**I also nearly told you something false, and I'd rather flag it than bury it.** My first attempt
searched for the flagged text across the whole page, and found that 7 of 18 flagged texts were still
present — which would have meant half our existing closures were wrong. I was about to write that up
when I looked at the actual words instead of the count. Every one of those 7 was one to four
characters long: "5", "26", "50", "97". Of course they still appear somewhere on a page — they're
bare numbers, they turn up in dates, prices, style codes. The distinctive ones were genuinely gone.
Searching within the right box drops it from 7 to 2. **My measurement couldn't have come out any
other way, which is the definition of a useless check** — the second time in one session I've caught
myself doing that, so both are now written up where the next person will trip over them.

**One thing I've deliberately left for you.** The two checks are currently both required. But the new
one is genuinely stronger evidence: if the words have gone, the copy demonstrably changed — the
timestamp only *claims* it did. So the new check could replace the old one rather than sit alongside
it. I haven't made that call, because tightening your gate needs no permission and loosening it does.
The practical cost of keeping both is that a page fixed by an edit which didn't bump the timestamp
will now sit unresolved. Happy either way — it's a one-line change.

### The check you asked for worked — the reviewer who blocked us three times now approves outright

Round six is back. Headline first: **the reviewer that blocked this in rounds three, four and five
now approves with no objections at all.** That was the one raising a real concern about closing
factual-claim items on weak evidence, it's the one you ruled on, and the check we just built answers
it. So the thing we set out to do is done.

The verdict is still "revise", and I want to be straight about why, because it isn't the same
argument any more.

**Eleven of sixteen approved, up against five objections — where last round it was fourteen and two.**
More objections, and yet the substance improved. Looking at what the five actually say, none of them
is about the design or the code behaving wrongly. **They are about the writing.** And specifically:
almost every new objection lands on text I *added* in round six.

That's the lesson of this round, and it's a bit humbling. The paragraph I wrote to fix the previous
round's complaint attracted two new complaints of its own. **Answering a reviewer by adding an
assertion just gives the next reviewer something new to doubt.** One of them says so almost in as many
words — it objects that my rationale is heavy with dramatic capitals and self-congratulatory
measurement claims rather than plainly describing the change. Reading it back, that's fair. So round
seven needs to be *shorter and plainer*, not more thoroughly defended, which is the opposite of my
instinct.

**The thing that blocked it is already fixed, and it turned up something that affects everyone.** The
reviewer objected that I justified the whole submission by citing one of your rulings — the one saying
reviews here happen after the code ships, because on this shared setup nobody can hold code back — and
that it had no way to check that ruling exists. It was right. The ruling is in our main instructions
file, which the reviewers cannot read; and it had **no trace at all** in the database they *can* read.
So we've been arguing from a rulebook our reviewers can't open. That will have been quietly hurting
every submission that cites a rule. I've written it up as a fleet-wide trap and, in doing so, got the
ruling into the database they read — so it's now checkable with one query.

I also nearly fooled myself again, and it's the third time today, so I'll keep flagging it. My first
check for that ruling searched for the *date* and returned 130 hits, which looked like plenty of
evidence. Searching for the ruling's actual *words* returned zero. A search that can't come back
empty isn't a check.

**One reviewer found a genuine flaw in my code**, and it's a good catch: my new check searches the raw
page source, while the original audit reads the extracted prose. Those should be the same text — that
matching-predicates principle is the whole basis of this design. But I measured it before rushing to
fix it, and on today's data the two give **identical** results. So it's worth correcting for
correctness' sake, and I'm not going to present it as fixing anything.

**Where that leaves the decision.** The most valuable thing on the list isn't about the council at all:
two reviewers independently want a real safety check — not a comment — on a piece of shared plumbing
that all six of these drainers depend on, where a mistake fails silently for every one of them. They
also caught me arguing for exactly that principle and then not applying it. I'd build that regardless
of whether we submit again.

### Approved — and it went live on the build that just went out

Round seven passed. After six rounds of "revise", the review board approved it, with four advisory
notes and nothing serious. And the fresh build carries it, so both safeguards are now running on the
live system.

**The proof is better than anything we've had on this lane.** Until now, checking whether our code
was actually deployed meant grepping for a string we'd added — which tells you the string is there,
but not that the binary is newer than some other change. This time the loader work *removed* a
distinctive line of code. So I could check for something that must be **absent**, and it is. That's a
much stronger test, and it's the one our own bug notes have been asking for.

**The lesson from seven rounds is not the one I expected, and it's worth telling you plainly.** The
round that passed was the *shortest* submission of the whole series — I cut it by about a fifth,
stripping out all the argument-by-argument history I'd been accumulating. Rounds four, five and six
each answered a reviewer by *adding* an explanation, and each time a *different* reviewer objected to
the thing I'd just added. One of them eventually said so outright: the writing had become the
problem, not the change. **Defending it harder was making it worse.** I'll carry that into the next
one of these.

**The approving round also found a genuine flaw, and I've filed it rather than banking the win.** One
reviewer pointed out that both our safeguards check what's in the *database*, while the whole point of
this type of finding is what the *public website* says. A page can be corrected in the database and
not yet republished — and in that window we'd close the item and declare the claim removed while the
live site still shows it. I checked: the relevant columns exist and neither our scan nor either gate
reads them. Then I measured: **two of the nine items we've already closed are on pages where the
database is ahead of the last publish.** That doesn't prove those pages still carry the claims — it
proves our evidence can't show they don't, which on this subject is the same problem. It's written up
as bug 262 with the fix ranked.

**One decision is still yours**, unchanged from yesterday: the two safeguards are both required, but
the newer one is the stronger evidence, so it could replace the older rather than sit alongside it.
Keeping both means a page fixed without touching a timestamp stays open. One line either way.

I've written a fresh handover so this can be picked up cleanly in a new conversation — this one has
got long.

### Kept both, built the third, and lost half an hour to another session

Your call recorded: **both safeguards stay.** I've written it into the handover as closed, so nobody
re-opens it as an open question in a fresh chat.

**Bug 262 was unowned, so I took it on and it's fixed.** I checked three ways before starting —
the ownership script, the work queue, and the live sessions — and the three sessions that mentioned
the file had only ever seen its name in a directory listing. It's built, tested and committed, and it
goes live on the next build. There are now three conditions before we'll close one of these findings:
the copy changed, the specific words went, and **the page was actually published after the change**.
That last one is the new bit, and it's the one that stops us certifying a fix that's still sitting
unpublished in the database.

Two details I was careful about, because both could have made it wrong: a page marked "deployed" can
still be serving an older build, so I check the publish *time* separately from the *status*; and a
page published in the same instant as its edit **does** count as published, otherwise we'd strand
items whose clocks happen to agree.

**And a genuine annoyance worth telling you about.** Halfway through testing, another session ran a
plain `git stash` — which takes the *whole working tree*, not just their own files — and my
uncommitted test changes vanished. The confusing part is that it leaves no trace: the file reads as
unmodified, so it looks like your own edits never happened. It cost about half an hour, and it made me
report two test results that had never actually run. I've corrected that in the record, written the
trap up so the next person recognises it in seconds, and committed straight away before redoing the
work. The lesson is one we already have written down and I'd been ignoring for ten minutes at a time:
commit as soon as the work stands up on its own.

Nothing is left in flight. The handover file is current and a new chat can pick up from it.

---

**2026-08-12, later that night — the fix shipped, the bug is closed, and one honest caveat.**

The roll went out. The published gate is live on the fleet: 98 running copies of the service, all the
same build, and I asked the running program directly whether it contains the new code rather than
trusting the version number. It does. So `262` is now fixed *and* live, which is the bar for closing
a bug here, and I've moved it into the closed pile with the evidence written inside it.

**The caveat, and it's the useful part of tonight.** The gate is switched on, but it has not yet been
*asked anything*. This job runs once a day, and today's run happened at 8:44 in the morning — hours
before either of the two newest gates existed. So when the tally says "this gate has refused nothing",
that currently means "nobody has put a question to it", not "it looked and was happy". Those two read
identically in the numbers and they are completely different situations. The first real test is
tomorrow morning's run, against the 21 items still open. I've written that down in three places so
nobody quotes the zero as a clean bill of health in the meantime.

**Where I got something wrong.** I tried to work out how often this sweep runs by looking at the
timestamps it leaves on each item, and concluded it had only ever run twice, skipping a day. That was
wrong — the sweep *overwrites* that timestamp each time it looks at an item, so what I was reading was
"when was each item last touched", not "when did the job run". A missed day would be invisible. I
caught it a minute later because every still-open item had exactly the same timestamp, which can't
happen if those values are a history. The awkward bit is that the conclusion I actually needed — that
the new gates haven't been exercised — is still correct, and comes from the *same* column; it just
happens to ask a question that survives the overwriting. A right answer and a wrong answer one query
apart, with nothing in the output to tell them apart. Written up properly so the next person spots it
in seconds.

I also tidied four places in the shared reference notes that had gone out of date tonight — including
one that still asked you a question you'd already answered (whether to keep both of the earlier
gates; you said keep them). Those notes get read by the automated reviewers as if they were current
fact, so a stale question there isn't harmless.

---

**2026-08-12, still that night — I chased the last open review comment, and it paid off sideways.**

The review that approved this work left one comment nobody had acted on: *go and check the claim that
only one thing creates these findings — and check it against the page re-rendering code specifically,
because that's where somebody previously made a confident guess and got it wrong.* So I did, by
actually reading those two files rather than searching them, which is the mistake the note was warning
about.

**The claim was right.** The re-render code doesn't create these findings and doesn't even write to
the table in question — it hands off to a different step that does the writing. One producer, as
documented.

**But the thing sitting next to it was not.** The first of our three safety gates works by asking
"has any part of this page been edited since the complaint was filed?" — using a "last modified"
timestamp. I went through every piece of code that writes to that table and found two that update the
timestamp **without changing a single word of the page**: one repairs a status flag, the other stamps
a review across every component on a page at once. So "last modified changed" does not actually mean
"the wording changed", which is precisely what that gate assumes.

**It isn't currently causing harm, and I checked that two independent ways** rather than reasoning it
away. First, the *second* gate runs before the first one and asks a question those writes can't fool —
"are the exact words we complained about still on the page?" Second, neither of those two bits of code
has ever actually run: one appears zero times in five and a half thousand job records (against a
control that appears 28 times, so I know the search works), and the other leaves a fingerprint that is
absent from all 1,458 components.

**Why this is worth telling you.** You decided earlier to keep both gates, on the grounds that
removing one was the harder decision to undo. This gives you an actual reason rather than a cautious
one: **the first gate could not have safely stood on its own**, because its central assumption is
breakable. Your instinct was right and now there's a mechanism behind it.

One honest gap: the nine findings already closed were all closed *before* the second gate existed, so
they rest on the first one alone — and there's no way to go back and check them, because the timestamp
gets overwritten and keeps no history. Given that neither of those two bits of code has ever run, it's
almost certainly fine. "Almost certainly" is the accurate word and I'd rather use it than round up.

**A pattern I want to flag about my own work tonight.** Three separate times I wrote a database query
to check something, got a clean confident-looking answer, and the answer was meaningless — the query
was built on a column that is empty for every row, so it could only ever have returned the answer it
did. I caught two of them by deliberately testing whether the query was capable of returning anything
else. I did not catch the third in time and had to correct it. It's a good argument for a habit we
already have written down and that I'd been applying unevenly: before believing a zero, prove the
question could have come back non-zero.

---

**2026-08-13 — the test I was waiting for ran, and it showed I'd framed it wrongly.**

Yesterday I told you the next morning's automatic run would be the first real test of the two new
safety gates. It ran on time, and it looked at 21 outstanding findings. None of the three gates
refused anything — and the reason is more interesting than "they were happy".

I read back what the system actually said about each of those 21 decisions. Eighteen of them came back
"this page still carries claims the register doesn't support" — meaning the complaint is simply still
true, so the check stops right there. Two were "the page is gone or has no content to read", one was
"this site no longer has a register to measure against". **Not one of the 21 ever got as far as the
gates.**

That's because the gates sit at the *end* of the process. They only come into play once a page reads
clean, and no page read clean. So the run happening is not the same as the gates being tested — which
is exactly what I assumed yesterday and shouldn't have. The honest position is that these gates are
built, switched on, verified present in the running software, and **still never actually consulted**.
They may stay that way for a while, because what they need is a page somebody has genuinely fixed.

I've also confirmed the timing properly rather than trusting it: the run happened while the software
containing all three gates was live, and the newer build deployed this afternoon still contains the
gate. So the result is a real result, not an artefact of the wrong version being loaded.

Nothing is broken here. The daily check is working correctly and refusing to close 18 findings whose
claims are still on the page — that is the system doing its job. The caveat is only about what we can
*claim* to have proven.

---

### 2026-08-14 — we can now see where the daily check stops, and I got two things wrong on the way

Yesterday's entry ended on an honest but unsatisfying note: the three safety gates are built, live,
and have never actually been consulted, and I had no way to *show* that other than reading the
wording of each decision by hand. That is a bad position to be in, because the only thing standing
between "the gate approved this" and "the gate was never asked" was me reading prose carefully.

So I've made the system say it outright. Every decision now records **which step of the process
decided it** — a short, fixed label like "the page still trips the check" or "the words the finding
quoted are still there". There are eighteen such steps, and the ones that sit at the end, where the
three gates live, are all named so they stand out. The question "did anything get as far as a gate?"
is now something you look up, instead of something you argue about from wording.

This changes nothing about how anything is decided. No gate got stricter or looser, nothing new can
close a finding, and nothing in the system reads the new label to make a decision — it is purely a
record of what happened. I was careful about that: a measuring instrument that can change the thing
it measures is no longer evidence about it.

It went through the review council and was **approved first time**, which for this lane is worth
noting — the last change here took seven rounds. Ten reviewers looked at it.

**Two things I got wrong, both caught, both worth writing down.**

The first was caught by my own test before anyone saw it. I had labelled the gate steps so they could
be found by a common prefix, then written the test to check exactly that — and it failed, because the
label for "all three gates passed, finding closed" deliberately reads as a closure rather than as a
gate. So a search by prefix alone would have found only the cases where a gate *refused* and missed
every case where one *approved*. That is precisely the confusion this whole change exists to remove,
and I had rebuilt it inside the fix. The test now checks the real question rather than a proxy for it.

The second was caught by the review council, and it is the more embarrassing of the two. I claimed the
new label lets us ask whether a gate has **ever** been reached. It does not: each finding's record is
overwritten every time the daily check runs, so the label only ever tells you about the most recent
run. A finding that reached a gate last week and is stuck earlier today shows only today's answer.
What makes it worse is that **this lane wrote the warning about that exact column two days earlier** —
and the reviewer quoted our own note back at us. I have corrected the wording everywhere it appeared
and logged it. The genuinely historical record does exist elsewhere (each run keeps its own full
set of decisions), but I checked rather than assumed, and only **one** run is currently retained — so
that is a surface that will become useful, not one that is useful now.

**One other thing worth reporting, because I nearly filed it as a bug and it isn't one.** Yesterday I
noticed the audit examining a page marked "archived", and assumed that was a defect — why audit a page
nobody serves? I measured it: three of the thirty findings sit on non-active pages. Then, before
writing it up, I actually fetched the pages. **One of them is archived and still serving 31KB to the
public.** So auditing archived pages is right, and the code's decision not to filter them out is
correct for a better reason than the one written beside it. The real distinction is not whether a page
is archived but whether it is actually *being served*, and nothing currently checks that. Two findings
are consequently parked for ever, asking a human to fix copy on pages that return "not found". I've
recorded that against the existing bug that owns this area rather than opening a competing one.

Where that leaves us: the change is committed and approved but **not yet live** — it ships on the next
release. The first daily check that carries it should show roughly eighteen findings stopping at "the
page still trips the check" and **zero** reaching a gate. That will look like nothing happened. It is
the instrument working, and for the first time it will be checkable rather than arguable.

---

### 2026-08-14 (later) — the new measurement worked, and it caught me out again within the hour

The daily check ran at 08:45 this morning, and for the first time it recorded **which step decided
each finding** rather than leaving us to read the wording. I had written down what I expected to see
*before* it ran, which is the only way that kind of claim is worth anything. It came out exactly as
predicted: eighteen findings stopped at "the page still carries these claims", **none** reached any of
the three gates, and none showed the "this part isn't instrumented yet" marker.

Better than that, two of the numbers independently reproduce yesterday's hand-reading. Yesterday I had
to read twenty-one decisions one by one to work out that two were "page missing" and one was "this site
has no register any more". Today those come out as counts, automatically, and they match. That is the
useful kind of agreement — two different methods, same answer.

So the honest position has finally changed shape. It is no longer "we believe the gates have never been
consulted"; it is "no finding reached a gate, and that is a query anyone can re-run". The reason is
unchanged and is not a fault: the findings are still genuinely true, so the check correctly stops
before it ever gets as far as the safety gates.

**And it caught me out a third time, which I'd rather report than bury.** I had written a warning note
saying that a blank value in the new field would never happen. It happens: the nine findings that were
already closed carry no value at all, because a closed item is never re-examined and so its record is
frozen at the day it closed. Anyone querying for "parts not yet instrumented" the obvious way would get
those nine back and read them as a gap, which they aren't — they're just old. Corrected, as a dated
note underneath the original rather than by rewriting it.

There is a small bonus in that. Those blanks now *prove* something we could previously only infer from
dates: every one of the nine closures happened before this measurement existed, so none of them can
ever be explained by it. That was listed as an unanswerable question two days ago and it is now simply
visible.

Nothing is outstanding from my side on this piece. The remaining work on this lane is the shared-helper
tidy-up and two older loose ends, all named in the handoff.

---

### 2026-08-14 (afternoon) — you said "yes, clean a page", so I did, and the three gates finally ran

Short version: a page that was making a claim it couldn't back up no longer makes it, and as a side
effect the three safety checks we built and had never once seen run, ran, and passed.

**The page and the claim.** Leopardess Consulting's case-studies page said, of our own platform,
"75,061 orchestration state records". The register of facts that site is allowed to draw on says that
number is currently 2,578, and it's a "no more than the live count" sort of fact. So the page was
overstating it by about twenty-nine times. This wasn't a judgement call on my part — two separate
automatic checks had already flagged it independently: the claims audit, and the freshness check that
watches for a registered fact drifting away from what the copy says.

**What I actually did.** I deleted one sentence. Not rewrote, not rephrased, not substituted a fresh
number — deleted "75,061 orchestration state records." and left the rest alone. That's the narrow
thing your 6th-of-August ruling allows, and it's also what the check's own instructions ask a human to
do with an unsupported number: either register it or take it out. The sentences either side stand on
their own, so nothing needed patching up. There's a nice accident in it: the very next line of that
paragraph reads "we would rather say so than let the number do work it has not earned", which is
more or less the reason the number went.

**One thing I got wrong, and caught before it did damage.** I had planned a single edit, to the stored
content, on the assumption that re-rendering the page would rebuild the published HTML from it. It
doesn't. That particular re-render is a straight assembly job — it glues together HTML that was
already rendered and saved earlier. So my edit would have been quietly ignored, the page would have
republished the claim word for word, and the next audit would have reported the same finding. That
would have looked like the audit being broken rather than like me having done nothing. I found it by
reading the code before firing it rather than by anything going wrong, and I've written the trap up
where the next person will hit it.

**Then the interesting part.** Before running anything I wrote down what I expected, so it could come
out wrong. It came out as predicted: the finding went through all three gates and closed. The record
it left states what each gate actually checked — the exact words the finding cited are gone from the
component they were quoted from; the copy was edited after the finding was raised; and the page was
published *after* that edit, so the public is seeing the corrected version rather than a fix sitting
in the database. That last one is the check we added because a page can be fixed in the database and
never republished.

**What I want to be careful about.** This proves the gates work when everything is in order. It does
**not** prove any of them would stop a bad closure, because none of them refused anything — they all
passed. So the line you've had from me for a week still stands unchanged: I can't yet say any gate has
prevented anything. What's changed is smaller than it sounds and I'd rather undersell it: the gates
have gone from never having been reached to having been reached once and passed.

There was a refusal within arm's reach and I missed it by about three minutes. If the check had run in
the gap between my editing the page and the page being republished, it would have refused with "the
correction is sitting unpublished" — which is precisely the case we built that gate for. The
re-render closed that gap by publishing immediately. It'll happen on its own the first time the daily
check lands in the middle of someone else's edit, and it costs nothing to wait for.

**One thing you should know because I caused it.** The daily check is fleet-wide, not per-page, so
running it by hand also closed four other findings that other people had genuinely fixed earlier
today — two tool pages on webdesign.co.uk, a directory page on ai-agent-orchestration.com, and a
link on webdesign.uk's front page. All of those would have closed tomorrow morning anyway; I just
brought them forward. Nothing closed that shouldn't have, and each one is individually reversible. I
deliberately didn't touch the daily schedule itself, so it still runs at its usual time.

Incidentally the whole job happened to need no AI at all, which is just as well, because the fleet's
AI capability went down this afternoon on a monthly spending cap and won't be back until the 1st of
September unless you raise it. Someone else has written that up properly.

---

**Later the same evening.** Nothing was broken and nothing needed deciding, so I went back and paid
off the last outstanding criticism from the review panel that approved this work a couple of days
ago. It's a small thing but it had been sitting on the list for two days and it's the only item left
that doesn't need the AI fleet, which is still down on the spending cap.

The criticism was this. Back on the 9th I moved a filter — the rule that says "don't audit a bit of
copy a human has pinned in place" — from the database query into the surrounding code. I wrote at
the time that this changed nothing about what the audit reports. Two of the reviewers, independently,
said the same thing back to me: *you have asserted that, you haven't shown it.* They were right, and
they were right to care, because this audit runs across every site we have, so if I had got it wrong
the effect would have been to quietly change which pages get flagged — and nobody would have seen a
failure, just a different set of results.

So I've shown it. The test runs the audit twice over the same page: once given only the copy the old
database query would have handed it, and once given everything including the pinned copy, which the
new code is supposed to step over. The two runs have to produce identical findings, character for
character. They do.

**But a test that passes proves nothing by itself** — that is the trap this whole lane keeps
relearning. So I deliberately broke the code, three different ways, to check the test would actually
notice. Two of the three it caught immediately. The third one is the interesting one: I put the old
database filter back, and the test carried on passing. That is not a hole in the test — it is the
answer to the reviewers' question, arriving from the other direction. The two versions really are
interchangeable in terms of what gets reported. The only thing that changes is whether a page whose
copy is *entirely* pinned shows up at all, and we want it to show up, because that is how the system
tells "this page is clean" apart from "I wasn't allowed to read this page". Those two look identical
if you only count findings, and confusing them is how you get a machine confidently telling you a
page is fine when it never looked at it.

I did all the breaking in a scratch copy rather than in the shared working folder, because several
of us are working in the same place at once and a half-broken file sitting there for even a minute
can end up in somebody else's commit.

**One near miss worth recording.** My first version of one test searched the source code for the
filter's text to check whether it was still there. That would have been wrong in an embarrassing and
invisible way: the exact phrase it was searching for *is* in the file — in a comment explaining that
the filter was removed. The test would have found the comment and cheerfully reported the opposite
of the truth. There's a standing note in our traps file about this exact class of mistake, which is
what stopped me; I'd only ever thought of it as something that catches other people's code checks,
not the tests I write myself.

**Left undone, deliberately, and written down so it isn't mistaken for finished.** Only half of that
old change actually moved. The equivalent filter for headers and footers is still in the database
query, so pinned copy there is still being dropped without being counted — the very thing the first
half was fixing. It doesn't bite us today. The test will now fail loudly if anyone changes it without
also adding the counting, which is the right way round.
