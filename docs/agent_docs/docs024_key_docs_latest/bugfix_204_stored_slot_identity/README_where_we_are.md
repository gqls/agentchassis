# Where we are — bug 204, the section names the planner keeps throwing away

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-08-21 — picking this up, and what I found

You asked me to look at bug 204. The first thing worth saying is that the headline
of that bug is already fixed, and has been since the 6th of August. So the question
was never "does this still bite" in the way the file's title suggests — it was "what
is left".

Here is the thing in plain terms.

When we build a site, each page is a list of **sections**. Most sections are named
after the kind of thing they are — `hero`, `faq`, `call-to-action`. But on sites we
took over from someone else, and then chopped their hand-written pages into pieces,
the sections are named by **position** instead: `prose-0`, `tool-1`, `prose-2`. The
name says "the first block of prose on this page", not "a component of type X". The
real identity of that block is a link stored in the database, pointing at whichever
component actually renders it.

Several parts of the platform look at a section name and try to work out which
component it means. They all did it the same way: check the catalogue of components
for one with that name. A positional name is not in the catalogue and never will be,
so the lookup fails. Back in August we fixed that in two places. The re-render path
and the page-build path both now look up the stored link first, and fall back to the
catalogue. Those work; I re-checked.

**What is left is a third place, and it behaves worse than the other two.** When the
site planner produces a new plan, a step called `validate_plan` checks every section
name the planner proposed and, if it cannot resolve one, **deletes it**. The other two
places merely postponed the section; this one removes it from the plan altogether. The
plan then gets written to the database, and the page's section list is overwritten with
the shortened version. The page keeps serving fine — the actual content rows are
untouched — but the record of what the page is *made of* is gone, so the next rebuild
has nothing to rebuild.

That is not theoretical. On the 20th of August, another session fired a replan at
loanandmortgagecalculator.co.uk to prove an unrelated fix, and **41 of that site's 45
live pages had their section lists emptied**. It was caught within the hour and put
back from a snapshot, so nothing broke in public. But the same run also queued 20 jobs
to "build" those now-empty pages, and had those been picked up they would have built
blank pages over live ones.

### The number that convinced me

Back in mid-August a different piece of work added a permanent record of every section
name this step throws away. It has been running since the 17th. In that time it has
recorded **140 discarded sections across 41 pages — and every single one of them is a
positional name.** Not one is a typo, a display name, or a stale component. The check
exists to catch mistakes of a kind that have not happened once; meanwhile the thing it
cannot see has accounted for one hundred per cent of what it threw away.

I like this number because it could easily have come out differently. If the check
were doing the job it claims, the list would be a mixture. It isn't a mixture.

### How much of the estate is exposed

Seven sites now carry section names that no component can resolve — 107 names in
total, up from 86 when the bug was filed. The bulk of it is
loanandmortgagecalculator.co.uk with 70 names across 41 pages. Six other sites carry
between two and eleven each. Any replan of any of them, today, would do what the 20th
of August did.

### What I have done so far

Checked that nobody else is working on it (two lanes have written *into* the bug file
recently, but both were reporting what they hit while doing something else, and both
said explicitly they were not taking it on). Confirmed the code is unchanged and the
setting that arms it is switched on for both planner agents. Filed the whole thing for
an independent second opinion through the diagnosis loop, rather than trusting my own
reading of it. And I have asked for a fix plan that covers the *whole* class rather
than the one site — because when I went looking, the same blind lookup turned out to
be used in **four** places, not one, and two of them write straight to live pages.

Next: the plan, then the review council, then the code.

## 2026-08-21 (later the same day) — the fix is written and committed

The change is in. It is not yet doing anything: Go only takes effect when a new
image is built and rolled out, so this sits dormant until the next release.

**What it does, in one sentence.** When the planner's checker meets a section name
it does not recognise, it now asks the page itself — *do you already have a section
called this?* — and if the answer is yes, it keeps it instead of deleting it.

That is the whole idea, and the reason it took care rather than five minutes is
that the obvious version of it was explicitly forbidden. There is a written warning
in our own notes, added last week, saying *do not fix this by widening the checker's
list of known components*, because three of the four places that use that list
belong to a path where someone deliberately decided **not** to widen it. So the
checker's list is untouched. Instead the question is asked per page, about that
page's own stored sections. The difference matters: widening the list would let the
planner place new things anywhere; asking the page only ever lets it keep what is
already there. There is a test that fails if that ever stops being true — if a
section stored on a *different* page of the same site were enough to rescue this
one, the test goes red.

**A number worth repeating.** Since the record of deleted sections started on the
17th, it has logged 140 deletions across 41 pages, and **all 140 are the kind this
fix is about**. Not one was the typo the checker exists to catch. I have kept the
checker's real job intact — a name that is neither a component nor a stored section
is still deleted, and there is a test for that too.

**One thing I got wrong, and it is worth telling you because it is a shared-machine
problem rather than a coding one.** Several sessions edit this repository at the
same time. I wrote a new file, then wrote the line that calls it, and before I could
commit the new file, another session's unrelated commit picked up my calling line —
leaving a call to something that did not exist yet. For **33 seconds** the shared
codebase would not have compiled. Nobody built anything in that window, so no harm
was done, but that is luck rather than care, and it is the second time in two days
this exact trap has closed on luck. I have fixed the written warning about it: the
old advice said "commit the definition first", which sounds right and is useless,
because by the time you are committing, the calling line has already been sitting in
the shared tree for however long the rest of the work took. The rule I have written
instead is: **write the new thing, commit it, and only then write the line that uses
it.**

**What is still outstanding.**

1. The review council has the change (it takes about half an hour, and I committed
   before the verdict, which is the normal practice here — the commit gets credited
   automatically if approved). **Somebody must read that verdict.** If it comes back
   asking for revisions, the code is already on the shared branch, so it needs acting
   on rather than filing.
2. There is a second, separate protection I have deliberately *not* included: the
   step that saves a plan to the database will still happily write an empty section
   list over a real one. Its two neighbouring columns in the very same database
   statement were given exactly that protection on the 19th, and this one was left
   out. I have designed it and written it up, but it belongs in its own commit and
   its own review round — it protects against *any* future cause, not just this bug,
   and that deserves separate scrutiny.
3. After the next release, someone needs to run the canary and — importantly —
   **prove the detector still works** before trusting the count going to zero. A zero
   from a blind detector looks exactly like a fix.

I also asked for an independent second opinion through our diagnosis loop before
writing any of this. It ran out of iterations before reaching a formal verdict, but
it independently confirmed the core of the problem and found its own examples on
five pages I had not looked at. Its one unanswered question — whether a fourth piece
of code had the same blindness — I checked myself: it does not.

## 2026-08-21 (end of session) — the review came back approved, and it still caught something

The review council approved the change. It also raised four points, two of them
substantive, and one of those was right about something I had already looked at and
got wrong. That is worth telling you plainly, because it is the second time today
that a check I thought I had done turned out to be half-done.

Here it is. If the database read fails while the checker is deciding whether to keep
a section name, my code keeps the name rather than deleting it — deliberately, on the
grounds that a momentary database hiccup should never be able to empty a page. Fine.
But I had only made that visible in the *logs*, which on these services scroll away
within seconds. In the permanent record, a run that kept everything because the
database was unreachable filed **nothing at all**, and so looked exactly like a run
where everything was fine. That is the same disease this whole fix is for — something
going quietly wrong and leaving no trace — reproduced inside the cure. It now writes
its own permanent record, saying how many names it kept without being able to check
them.

The second point asked whether the *other* places in the codebase that look up a
page's stored sections already have the same class of flaw. I checked rather than
assuming. One of them turned out not to do that kind of lookup at all, so the concern
did not apply. One of them genuinely has no rule for what to do when a page has two
sections with the same name — it would silently pick whichever the database happened
to return last. I measured it: **no page on the estate is currently in that state**,
though eighteen pages do have repeated section names, so it is reachable rather than
impossible. I have written it down and deliberately not fixed it, because it is a
different question in someone else's area and folding it in here is exactly the kind
of scope creep the reviewers also (fairly) flagged.

**Where that leaves things.** The read-side fix and the write-side protection are both
committed. Neither does anything until the next release. The second review round is
still running. After the release, someone needs to run the canary — and, importantly,
prove the detector still fires before trusting the count going to zero, because a zero
from a blind detector looks exactly like success.

## 2026-08-21 (after the new build went out) — it's live, it's proven, and I found my own bad number

The new chassis build carries everything. I checked that properly rather than
assuming: I asked the running program which version of the code it was built from, and
then — because that answer could in principle be a coincidence — I also searched the
running binary for a made-up version string that should *not* be there. It wasn't. So
the check can tell the difference, which is the only reason the positive result means
anything.

**The protection is proven against the real database.** This is the bit I most wanted.
Our unit tests can check that we send the database the right instruction, but they
cannot check that the database *does* the right thing with it, because they don't run a
real database. So I ran the real instruction against the live database inside a
transaction and then undid it. All four cases behaved: an empty list did not overwrite a
real one; a real list still overwrote a real one (so we haven't accidentally frozen
anything); the deliberate "yes, empty this page" route still works; and emptying an
already-empty page is quiet. Nothing was left behind.

**Then I found a mistake of my own, and it's an interesting one.** I ran a query asking
"of the section names that can't be resolved, how many does my fix actually reach?" —
and on three sites the answer looked bad, as though my fix had a hole. So I listed the
names to find the hole. They weren't the kind of name this bug is about at all: they
were ordinary names written with underscores instead of hyphens, which the system
already handles fine.

The fix didn't have a hole. **My counting method did.** The query everyone has been
using — including the one written into the original bug report back on the 5th of
August, which I copied and re-published six times — compares names *exactly*, while the
real code first tidies them up. So it has been over-counting from the start. The honest
figure is 88 names across 6 sites, not 107 across 7. One site comes off the list
entirely.

None of the conclusions change, and I want to be clear about why, because it would be
easy to over-correct. The number that mattered was never this one — it was the count of
sections the system was *observed* deleting, 140 out of 140 of them the kind this bug is
about. That's a record of what actually happened, not an estimate of what might. The
census was only ever "how many pages are exposed", and the notes already said in several
places that it isn't the measure of success.

What I have now measured properly is coverage: **83 of the 88 are protected by the fix.
The other 5 are names that refer to nothing at all** — no component anywhere matches
them, and the pages don't have sections by those names either. Deleting those is the
checker doing its actual job. So it protects everything that should be protected.

**What I have not done, deliberately.** The fix has not yet run for real. No site
planner has run anywhere since the incident on the 20th, so the current clean readings
prove nothing — which is exactly the trap I warned about in the last note, met about my
own work. To make it run I'd have to trigger a replan on one of the sites with these
positional names, and all four of them are being actively worked today, one of them by a
session that committed to it this afternoon. Triggering a replan there is precisely what
caused the original incident. I wasn't prepared to repeat that in order to test my own
fix. I've written down what I'd need — a quiet site, a snapshot first, and the job queue
cancelled before any repair — and I've left a note for the team working the most
affected site telling them what to check if they replan it anyway.

**One more thing I found while looking for a safe way to test it.** There is a second,
much older planner in the system that has the same setting switched on but is *not*
covered by my fix, because its workflow never tells it which site it's working on. I
left it alone and wrote down why: it has no way to save anything, so its mistakes go
nowhere, and it has run twice in its entire life — both times on the 9th of August. I
could have added a guess at where to find the site identity, but I'd have been shipping
code I couldn't test, which is a habit that has bitten this system repeatedly.

## 2026-08-21 (close) — moved to bugs_closed, on your direction

You've retired the standing instruction to leave fixed bugs in the open folder, so I've
applied the normal rule: fixed **and** live means closed.

Two files moved, not one. **204** is this lane's. **189** is a different bug from earlier
in the month that was fully fixed, live and behaviourally verified on 6 August — it had
been sitting in the open folder for a fortnight for no reason other than that
instruction. Those two were the only files in the whole open folder held there by it; I
checked rather than assumed, and both say so in their own text.

I've put the closed status at the **top** of each file rather than 360 lines down, because
anything that shortens a document — a listing, a search result — shows the opening and
stops. And 204's status block says plainly what is *not* proven, so nobody reads "closed"
as "fully exercised".

One small thing worth telling you because it nearly caught me: my own duplicate-check said
189 looked wrong, because two completely unrelated bugs share the number 189. Checking by
name rather than number cleared it. That ambiguity is documented and I still walked into
it — the estate's own advice is to resolve these by name, never by number.

The old paths (`bugs_open/204`) now appear in about 170 places across other teams' notes.
I've repointed the ones this lane owns and deliberately left the rest: they refer to *case*
204, not to a file location, both folders are searched by convention, and rewriting 35
files that other sessions are actively editing to change a directory name in prose is
exactly the kind of broad edit that swept my own half-finished work this morning.
