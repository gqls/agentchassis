# Where we are — directory build handler (`bugs_open/206`)

Plain-prose running log, append-only, newest at the bottom. Started late (2026-08-24) — the
lane ran from 08-06 to 08-08 without one, which was a gap in the standing five, so the first
entry back-fills what happened in plain terms before carrying on.

---

## 2026-08-24 — back-fill: what this lane was about, and what it did in August

Two pages on vetcomparison.uk had never worked. One was meant to list the site's vet practices,
the other its advice guides, and both returned "not found" — while the site's own homepage
linked to the first of them. That was the original complaint.

The cause turned out to be two gaps stacked on each other. The platform keeps a list of "which
kind of page gets built by which builder", and for these kinds of page that list said, in a
comment, that no builder existed yet. Underneath that, the component meant to render a real
business listing pointed at a data query that had never been registered anywhere. So even a page
that got as far as being planned had nothing that could fill it in.

The lane built the missing piece: a query resolver that reads a site's own directory-export
configuration and pulls from the same business data the site already exports (so the page and
the export can never disagree), plus an action that fills in a page's layout **only** when it
has none at all — written that way deliberately so it can never overwrite a layout a person or
another process had already chosen. A new handler chains those together and then hands off to
the platform's ordinary page builder. No new content-writing logic was needed.

That went through three rounds of the council review gate before approval, and two further
defects showed up only when real builds ran — the kind you cannot find by reading code, only by
watching it fail. Both were fixed by follow-up migrations the same day. Both pages then built
and deployed through the ordinary queue, with no manual dispatch, which was the whole point.
The directory page listed sixty-one real practices with real postcodes; the guides page listed
exactly the three guides that actually exist and invented nothing.

Then the lane went quiet for sixteen days.

---

## 2026-08-24 — resumed: the fix was fine, but it was never fleet-wide

Picked the lane back up today. First job was to check the August work is still true, because a
fix that was live three weeks ago is not evidence about today. It is: both pages still serve
their real content, and — the part that actually matters — a fleet-wide re-render swept them
yesterday and the real listings survived it. So the pipeline genuinely reproduces the result
rather than a good page having been frozen in place. A second team working on vetcomparison
proved a third kind of page through the same handler this morning, first attempt.

Then the surprise. I went looking for the leftovers we knew about and found instead that the
August fix, while correct, was only ever installed on **one of the doors**.

The platform decides "which builder builds this page" in more than one place. In August we
taught one of them about the new handler. A second one — the routine that reconciles a site's
plan against what has actually been built — never asks that question at all: it had the answer
"use the generic builder" typed directly into it. So pages of exactly the kind this bug is about
keep arriving at a builder that cannot build them, failing in the same way, and parking for a
human who never comes.

Five pages are sitting in that state right now across three other sites. The one that makes the
point is a directory page on garden-tools.uk: it has been waiting fifteen days for a builder that
has existed and worked the entire time. It was never told the builder's name.

The fix is to stop having two copies of that decision. There is now one function that answers
"which builder builds this kind of page", both routines ask it, and it is covered by tests — the
decision had **no** test coverage at all before today, in either copy. I also made it so that a
page whose kind genuinely has no builder yet gets filed as a visible, deferred "we need this
capability" note, rather than being sent to a builder that will fail and then parked under an
error message blaming missing data when the real cause is a missing builder.

**One thing needs a decision that is not mine.** Another team looked at this same set of stuck
pages three weeks ago and deliberately decided to leave that routine alone, on the reasoning
that these stuck items are real findings and should stay visible. I think my version keeps them
just as visible and describes them more honestly — nothing is hidden, the same record is still
written at the same moment — but they made that call on purpose and with the council's backing,
so I have put it to the council rather than quietly overruling it, told that team in their own
file, and offered to take that half straight back out if they disagree. The other half of the
change — actually routing pages to the builder that exists — is untouched by their decision and
is where the real win is.

**Three things I got wrong today**, all caught by other people's checks rather than my own, and
all written up:

- I corrected another team about something in **my own lane's history**, from memory, and was
  wrong — my own notes already held the correction. They had already written my wrong version
  into their file before I caught it.
- My first survey of how widespread this is returned **zero**, and the zero was an artefact of
  how I asked. A neighbouring session mentioned, in passing, a lesson about testing your own
  filter against a case you know is real. I tried it; the filter was blind. The corrected survey
  found the five stuck pages and **changed what the fix should be** — the problem was the
  producer, not the list I had been staring at.
- I cited a database column that does not exist in a council submission, from misreading two
  command outputs as one. The council's reviewers caught it independently and sent the
  submission back. The code was right; the description of it was not.

And the one worth remembering: I re-discovered a problem another team had already measured,
ruled on, and closed **three weeks ago**. Our standard "check nobody else is on this" tools all
search by the bug's number or name — but this problem's identity is an error message, and no
number-based search will ever find it. I only got there because the council's own check returned
a number eight times larger than mine and I went to find out why.

**Where it stands**: the fix is committed and reviewed-pending; it is Go code, so it does
nothing until the next time the fleet's images are rebuilt. The bug file stays open — not out of
caution, but because the fault is genuinely still happening on the fleet today, and it stops
happening when the fix ships, not when it is written.

---

## 2026-08-24, later — the build landed, the code is in it, and nothing has changed yet

A fresh chassis build went out this afternoon (v1.0.1334) and I checked whether my change is
actually in it. It is — I asked the running program directly on both copies of the service, and
also asked it for a piece of text that should *not* be there, so the answer means something. One
detail worth recording because it is the sort of thing that usually gets fudged: the string I
looked for is one I only added in the *last* revision of the change, so finding it proves the
build contains the version that was approved, not merely something from this lane.

**And nothing on any site has changed, which is correct but not what either of us predicted.**

The other team and I had agreed a neat test: after the build, a directory page on their
test site should start working on its own, with nobody touching anything. It will not, and I
found two separate reasons, each of which would have been enough on its own.

The first is that the routine my change fixes does not run on a schedule. It runs as part of
building or publishing a site, so a site that nobody is currently building never reaches the
fixed code at all. Their site was last reconciled the day before the build.

The second is the one I should have seen, because I had already written it down this morning in
a different context. My change decides where a job is *sent when it is created*. The jobs for
these pages already exist — they are the ones sitting stuck — and while a stuck job is on file,
the system deliberately refuses to create a second one for the same page. So the fixed code will
look at that page, see a job already exists, and move on. The correct new decision never gets
made.

So the test we both liked could never have produced an answer either way. That is worth being
plain about: two of us reviewed it, we each told the other it was better than what we had, and
the agreement is precisely what stopped either of us checking. A second pair of eyes protects
you from being wrong about something you said; it does not protect you from something you both
assumed without noticing.

**What an honest proof needs** is to clear the stuck job so the page is eligible again, then give
the site a build so the routine actually runs, and then check the *new job* — that it was created
pointing at the right builder, by the system, with nobody steering it. The page coming back to
life follows from that, but the job is the thing that proves the fix.

I have not done it, because clearing that job means acting on another team's site, and they set
that site aside deliberately as a clean example of what the platform does unaided. I have asked
them and offered three options, including simply recording that the fix is proven by tests and by
the running binary but not yet on a live page — which is a real gap, and I would rather name it
than describe it as finished.

**Where that leaves the bug: still open, and it should be.** The code is right, reviewed and
live; the pages it exists to fix are still broken; and until one of them builds through the
ordinary route, nobody has watched the thing actually work.

---

## 2026-08-25 — the half we had to leave behind yesterday is done, and I found a hole in our own test

**The short version: the job that was blocked is finished and passed review first time, and
separately I found that the test we were going to use to prove all this could have been passed by
a page somebody had fixed by hand. That second thing is the more useful discovery.**

### The blocked job

Yesterday's note explained that there were two places in the code that decided "which builder
builds this kind of page", they disagreed, and that disagreement is what let a directory page sit
broken for fifteen days while the thing that could build it was running. We fixed one of them and
could not fix the other, because the file it lives in was in a state where committing it would have
broken everyone else's build. So we left it, and left a note telling whoever picked this up to
**check whether that was still true rather than assume it**.

It was not still true — other people had tidied that file overnight. So today the second copy is
deleted. There is now exactly one place in the whole system that answers "which builder builds this
kind of page", and both routes ask it. That also unblocked a page type we had deliberately held
back yesterday (the "index" pages that list other pages — `guides-index` and the like), because the
only reason for holding it back was the risk of the two copies disagreeing, and there are no longer
two copies.

The council reviewed it and **approved it first time — thirteen reviewers, no objections above
"low", no vetoes.** That is unusual and I am recording it because yesterday's equivalent took six
rounds.

### The hole in our own test, which matters more

We had a query written to prove the fix works. It looks for a page whose build job was created by
the automatic planner and points at the correct builder. If it finds one, the fix works.

**It does not follow.** The field that query reads can be changed by hand — and changing it by hand
is our own documented repair procedure for a stuck page. We have done it at least twice. So a page
that a person rescued in August looks exactly like a page the fix routed correctly.

I checked how bad that is: **every single row in the entire database that would have passed our
test is one of those hand repairs.** Three of them, all on the same site, all created in July,
all touched by a human weeks later. Had anyone run that query as written, we would have declared
the fix proven — using rows created by the very bug it fixes.

There is a clean way to tell them apart, and it now gates the query. The new code writes an extra
detail into the job record when it creates it; a hand repair only changes the one field and cannot
add that detail. Right now **no record in the database has that detail at all** — 508 of them, none
— so the first one that does was necessarily made by the fixed code. That is about as clean a test
as we are going to get.

### Something I nearly got wrong, and want on the record

Three of the reviewers asked a question I had not thought to ask: are there *other* places that
create these build jobs and skip the shared decision? I went and looked. **There are — six places
in total, not two, and three of them hardcode the generic builder.** One of them is the busiest job
creator in the fleet and was still creating jobs today.

The obvious conclusion was that the bug is still alive through those other doors, and I was one
sentence from writing that down. So I measured it instead: across every job those doors have ever
created for this kind of page — twenty-six of them — **not one** shows the failure this bug is
about. Their failures are a different, deliberate safety mechanism doing its job.

And there is a reason, which is why I believe the number rather than just reporting it. This bug is
about pages that have **no layout at all yet** — brand new pages, at the moment a site is planned.
The other doors act on pages that already exist and already have a layout. They skip the shared
decision because their pages do not need it.

**So the honest position is: no, the bug is not lurking behind those doors — and I would have said
the opposite if I had stopped at the search and not done the counting.**

### Where the bug stands

Still open, and still for the same reason as yesterday: **nobody has yet watched this work on a
real page.** The proof arrives free the next time anybody builds a new site with a directory or
index page on it — no site needs disturbing, nothing needs clearing. I checked today: no new site
has been built since the code went live, so there is nothing to read yet.

What changed today is that when that build does happen, we now have a test that can actually tell
us the answer — which yesterday we did not, and did not know we did not.

---

## 2026-08-25 (later) — I had the day's work reviewed by a fresh pair of eyes, and it found four real problems. Three were mine.

I asked a reviewer running on a different model to go over today's work and try to **break** it
rather than agree with it. That was worth doing, and I want the results here rather than only in the
technical notes, because one of them concerns your instructions and another is embarrassing in a way
that is useful.

### The one that needed fixing immediately: I told the next person to do something you had forbidden

This morning I wrote a handoff document, and one section of it listed three pages on
**garden-tools.uk** as an operator job, pointing at the recipe for clearing them.

**You had retracted permission for exactly that, five minutes earlier.** Another team recorded your
ruling — nothing is to be cleared on that site, because its whole value is that it was built with no
human help and four teams are measuring against it. They filed that note *into this lane's own
folder* at 11:05. I committed my handoff at 11:10 without having read it.

Worse: they filed the note **because yesterday's handoff had the same flaw** — one section saying
"do not clear these" and another eleven lines later saying "clear these". I then reproduced it in the
document written to replace it. That is now corrected: garden-tools.uk is marked do-not-touch with
your ruling attached, and the two pages on other sites that your ruling does *not* cover are
separated out, which is what the other team explicitly asked for.

### The embarrassing one, and the most useful

This morning's headline finding was that our test for "did the fix work?" could be passed by a page
a **person** had repaired rather than one the fix had routed. I found it, measured it, and published
a replacement test which I described as **airtight**.

It was not. The reviewer found two ways to pass it by hand — and one of them is a repair procedure
**I wrote myself**, in the same file, in this week's work.

And here is the part worth your attention: **the measurement that caught the original problem
contains the very check that fixes it.** When I proved those three records were hand repairs, I
proved it by looking at whether anything had touched the record since it was created. That is
exactly the check my replacement test needed. I used it to reach the conclusion and then left it out
of the remedy, four paragraphs later on the same page.

So the lesson is not "I missed something". It is: **a correction is not exempt from the discipline
it is enforcing.** I spent the morning writing about checks that give the right answer for the wrong
reason, and then wrote one.

Both holes are closed now, and I have written down the permanent fix — which is a small code change
rather than a cleverer query — as a named next job rather than slipping it into an approved change.

### Two smaller ones

The reviewer ran experiments on my brand-new tests that I had not: it deliberately broke the code in
ways my tests should have caught, and **two of them slipped through**. One matters — my own change
today removed a safety check, leaving a single remaining safeguard that nothing was testing. Both are
now tested, and I confirmed the tests fail when the code is wrong.

And a "zero" I reported as proof was bounded by something I did not mention: three of the twenty-six
records I checked predate the very error-reporting that would have shown the problem, so for those
three, a silent failure and a success look identical. My conclusion still holds for the other
twenty-three and for a separate reason I gave, but I presented the count as the thing that settled
it, and it did not settle all of it.

### Where this leaves us

**No worse than this morning, and better instrumented.** The code change is unaffected — the
reviewer checked it line by line against the previous version and found it exact. What was wrong was
the *measuring*, twice over, and both are fixed. The bug is still open for the same reason as before:
nobody has watched it work on a real page, and no new site has been built since it went live.

I would rather report a review that found four things than one that found none.

---

## 2026-08-25 (evening) — the build we were all waiting for arrived, and for this bug it is a negative result

Another team (`bugs_open/381`) built a new site from scratch today — `homegarden.uk`. Four teams,
mine included, had been waiting on exactly that, because it is the only way to watch this fix work
without disturbing anything. They captured the evidence for me before the machinery consumed it,
which was generous and precisely what I asked for.

**It cannot prove our fix, and that is worth saying plainly rather than dressing up.**

The site planned 21 pages, and none of them is the kind of page where the bug and the fix behave
differently. Seventeen are "index" pages, the rest ordinary content — and all of those go to the
same builder whether the bug is present or not. So the records look perfect: created by the right
routine, correctly routed, untouched. And they would have looked exactly the same before we changed
anything. **The one page type that would settle it — a directory page — the site does not have.**

So the wait continues, but with a sharper condition than I had written down. I had told the next
person "the proof arrives free on the next new site". That was wrong: it arrives free on the next new
site **that contains a directory page**. There is now a one-line check to run against the plan before
anyone declares a build to be the proof. Better to find that out from a real build than to declare
victory on one of these and be corrected later.

### Two things I got wrong, one of which the other team caught

**I sent them a warning that was false, and they had already adopted it.** I told them their index
pages would silently fail to build, because the relevant fix is written but not yet running on the
live system. They took it seriously and put it in their build-acceptance checklist.

Then I checked. Those pages all have a layout, and the failure I warned about only happens to pages
that have *none* — one of their pages had already built and gone live through the very builder I was
warning about, before I retracted.

The reason it mattered more than a wrong guess: their own bug is about pages that come out thinner
than the plan promised, and this site is a textbook case of it — seventeen near-identical index pages
where the plan promised one month-by-month structure, all reporting success. My warning would have
told whoever investigated that to go and check the routing first, find nothing wrong, and have a
ready excuse to stop — on pages where the routing is correct. **I would have handed their real
finding an alibi.** Retracted the same hour, with the measurement.

**And they corrected me.** They noticed that something I had told the world was missing from these
records was in fact present on every one of theirs. They were right and my note was stale — the
mechanism went live yesterday, and my count had been taken during a window when nothing had run.
They had one site and the right answer; I had months of history and an out-of-date premise.

### Where this leaves us

The code is unaffected — nothing found today touches it. What today changed is what we know about
our own measuring: the condition for proving it is narrower than I wrote, one caveat I was carrying
turned out to be stale, and one turned out to be unnecessary (we now know which routine mints these
records on a new site, which I previously could only guess at).

I have logged both of my errors to the shared record, including the warning I sent to another team,
because that one had a cost outside this lane.

---

## 2026-08-25 (night) — it is live, and I have made the proof permanent instead of perishable

**The fix is running.** The new build carries it, and I checked that properly this time: the servers
state which version of the code they were built from, and the change is an ancestor of it — with a
control, meaning I also confirmed that two changes made *after* the build correctly do **not** show
up. That control is the step I skipped yesterday, and skipping it was what produced three confidently
wrong readings in one command.

**Nothing has visibly changed, and that is correct.** The routine that files this work only runs when
a site is planned, and nothing schedules it. The already-stuck pages hold their own place in the
queue, which blocks them being re-filed. So the fix changes what happens on the *next* plan; it
cannot reach backwards. I mention it because "we rolled it and nothing happened" is exactly the shape
that gets misread as failure.

### The one piece of new work

Yesterday I found that our test for "did the fix work?" could be passed by a page a person had
repaired by hand. I patched it with a check that the record hadn't been touched since it was created.

That patch was correct but **perishable**: the moment the system legitimately picks the job up — a
few minutes later — the check stops working. A proof that dies when the system does its job is not
one you can count on being able to run.

So tonight both routines now write down *which builder they chose* alongside the record. A person
repairing a page changes the one field; they don't change that note. So the two agree when the
machine did it and disagree for ever when a human did — and the disagreement is itself the signal,
which is the right way round. It is written, reviewed by the council (verdict pending as I write),
and will take effect at the next build.

I deliberately did both routines rather than the one my test reads, because doing one would have
recreated — inside the measuring instrument — the exact defect this whole bug is about: two places
that should agree and don't.

### Where we are

The bug stays open for one reason, unchanged: **nobody has yet watched this work on the kind of page
where the bug and the fix behave differently.** Today's new site had none of those pages, which is
why it couldn't settle it. That's now written down as a precondition with a one-line check, so nobody
spends another build finding out the same way.

Everything else is done: the code is finished, reviewed and live; the follow-ups are named with
evidence and belong to other lanes; and the handoff for whoever picks this up next is written.

---

## 2026-08-26 — everything is built and running. I still don't think we should tick it off, and here's why.

The last piece went live overnight. Both parts of the fix, plus the permanent proof mechanism, are
now running and both have been through review. **There is no code left to write on this.**

Checking that took four steps rather than one, and it's worth a line because the easy check has
quietly stopped working: the servers announce which version they were built from once, at startup,
and on a busy service that announcement has scrolled out of reach within hours. It was gone from both
machines. So I read the version off the image itself, confirmed our changes are included, confirmed
that two changes made *after* the build are correctly *not* included — that's the control that makes
the answer mean something — and then checked the running program actually matches the image I'd read,
because a version label and what's running aren't automatically the same thing.

I've also corrected something I'd written into our own guide: I'd suggested you can settle this by
comparing timestamps. You can't. A build made an hour ago can be built from last week's code, so
timestamps can prove a change *isn't* there and never that it is.

### One count I'd been getting wrong for days

Every time I've told you "five stuck pages", then "six", I was running a query **filtered to the
sites I already knew about.** Unfiltered, there are **nine** — three on sites that had never appeared
in any of my counts.

It's an uncomfortable one because the query was correct and the numbers were real; what was wrong was
that its filter came from my own prior belief rather than from the data. A count restricted to the
things you already know about cannot find the thing you don't.

### Can we close it?

**The honest answer is: you can, and I'd rather we didn't.**

By the written rule — fixed and live — it qualifies. Nothing is outstanding in code.

But this bug exists *because* somebody once concluded the machinery worked without watching it work.
That's the actual sentence at the top of the file. And it was re-opened a second time for the same
reason: a fix from the 8th was live in one place and missing in another for fifteen days, and nobody
noticed because nobody looked at the result. Closing it now on "it's running and the tests pass" would
be that same move a third time, on the one bug that is *about* that move.

What's missing is small and specific: **nobody has yet watched the fixed code handle a single page of
the kind where it behaves differently.** That arrives free the next time anyone builds a new site with
a directory page on it — no site disturbed, nothing to clear. There's no way to force it cheaply: the
routine only runs as part of a full re-plan, and re-planning would destroy the untouched site four
teams are measuring against, which you've already ruled out.

So my recommendation: leave it open, with one clearly-stated thing outstanding. If you'd rather close
it, the honest way is to close it **saying plainly that we never observed it working**, and to move
the nine stuck pages somewhere they'll still be read — because caveats inside a closed file stop
being read, which is its own recurring problem here.
