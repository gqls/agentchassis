# Where we are — staged component build

*The owner's running plain-prose log. Append only, newest at the bottom. No jargon,
no tables of field names. The owner maintains this too — never rewrite or reorder it;
add a dated correction below instead.*

---

**2026-07-30 — the lane picks this up properly.**

You said the provenance and ladder work is now this lane's project, so it is. Up to
now it was a proposal sitting in a folder waiting for someone else to start; now it has
the same five working documents every other workstream here keeps, and it has an owner.

The first thing I did was the thing the proposal itself said to do first, which was to
stop assuming and check one specific thing: whether the existing machinery that lets a
*tool* carry its own specification and history in the database would fit a *component*
without needing to be changed. I had marked that as unverified rather than guessing at
it, precisely so it would be the first job.

The answer is no, and it is the best kind of no. Nothing about the design is wrong —
there is a single line in the database that lists what kinds of thing are allowed to
have one of these documents, and "component" is not on the list. Adding it is a
one-line change, it cannot break anything that already works, and it has been done four
times before for other kinds of thing, so there is a well-written example to copy.
Perhaps twenty minutes of work whenever it goes through review.

There was a genuinely nice surprise underneath that. Two tables are involved — one for
the specification, one for the running history — and only the history one has a column
for which website it refers to. That turns out to be exactly the split we need, and it
was already there. A component's *design* is shared across sites: the same carousel
serves eleven different websites. But whether it actually works is a question about one
page on one site. So the specification being site-less is correct, and the verdicts
belong in the history. We didn't have to design that; the tables already assumed it,
which is quiet evidence they were built with something like this in mind.

I also caught a trap before it fired rather than after. The two tables don't currently
have identical lists, because one of them also allows a category another team added
last week. The obvious way to make the change — copy one list over the other so they
match — would have silently invalidated fifty-seven rows of somebody else's work. It's
written down beside the change now.

The other thing worth telling you is what came out of reviewing the other team's
report, because it turned into a real constraint on this project rather than just a
comment on their document. When one of our checks doesn't recognise the kind of test
it's been asked to run, it doesn't fail — it quietly *skips*. And a set of tests where
everything skipped counts as a pass, and then stops re-checking for a week. That is
tolerable for a single checklist. For a ladder it is corrosive, because the whole point
of a ladder is that passing one rung is what earns you the next. So a rung that couldn't
actually run its own test now has to report "don't know", never "fine". That's written
into the plan as a requirement, not a nice-to-have.

And it isn't theoretical. The newest and most useful test we have — the one that can
tell the difference between something being *on* the page and something being big enough
to actually see and click — was written this afternoon and hasn't been deployed yet. So
at this exact moment, the single best test for this project is also the one that would
silently do nothing. I've logged that where people will trip over it.

One mistake of my own worth recording, because it nearly became a false report. To check
whether that new test was deployed, I searched the running program for its name and got
nothing — which looked like a clean answer. But I also searched for a test I *knew* was
there, and got nothing for that too. It turns out short names get compiled away and
genuinely aren't findable, so my "it's missing" would have been wrong for the wrong
reason. Searching for a longer, distinctive phrase instead gave the real answer. The
lesson is dull and important: when a search comes back empty, check that the same search
can find something you're certain is there.

Lastly, on how this fits with the two other items you pointed at. The other team
proposed a division that I think is better than what I'd written: the site maturity
ladder is the *vocabulary* of levels, this project is the *mechanism* of gates, and the
render-and-check feature is the *instrument* they both need. That makes them three
things that fit together rather than one big thing, which means this lane can get on
with its part without waiting on the others. I have deliberately *not* taken ownership
of the site maturity ladder — that's still yours to place.

---

**31 July 2026, afternoon (fresh thread, picked up from the handoff)**

The job was the one the handoff called the concrete next build: our own tool on
fundamentallyai.com, the review-council simulator, had a written specification but nothing a
machine could check. So when the test system was pointed at it, the run started, found nothing
to test, and stopped — and that came back looking like a clean result rather than like nothing
having happened. That is the failure this lane exists to remove, and it was the last remaining
instance of it that we could actually fix.

It is fixed. The tool now has eighteen checks attached to it, and they are the sort that can
genuinely be wrong: the panel actually builds in the visitor's browser, moving the strictness
slider actually changes the answer, each of the four headline numbers is really a number rather
than the dash it is served as, the buttons that select eight, twenty-six and two reviewers each
select that many, and — the one I would keep if I could only keep one — **when you deselect
every reviewer, the tool says "n/a" rather than inventing a hundred per cent.** That last one
is the honesty we promised in writing when we built it, and it is now a machine-checked
property instead of a sentence in a document.

**Two things are worth telling you about how it went, because neither was the plan.**

First, I built the test-the-test tool before writing any of the checks, and it paid for itself
on its first run. The eighteen checks passed against the live site on desktop and mobile,
thirty-six for thirty-six, first attempt. That is the exact shape this lane has learned to
distrust, and it was right to: when I then deliberately broke the tool one way at a time, one
of my own checks — the one about the strictness slider — **carried on passing.** It turned out
to be checking something weaker than its own name claimed. I renamed it to what it actually
proves and wrote down the limitation. A check that quietly means less than its name is the same
mistake as the one that caught us yesterday, and this is the first time in this lane that it was
caught before the thing was published rather than afterwards by a reader.

Second, the same tool told me four of my checks had never been *seen* to fail — they only broke
when I broke the whole page, which proves they depend on the page working, not that each one is
watching its own number. So I added four more deliberate breakages, one per check. Final state:
seventeen deliberate breakages, seventeen caught, all eighteen checks watched to fail, and the
harness refuses to report a pass if any check has no breakage of its own. There is no way to get
a green with a hole in it.

**One thing I deliberately left out**, so it does not look like an oversight later. There is a
newer check that measures whether an element is actually big enough to see and click, and it is
now live on the cluster — but it is broken in a way another team has already diagnosed and
filed: any element whose size is a whole number of pixels reads as zero, so it accuses correct
elements of being invisible. Our roster's checkboxes are exactly that shape. Using it today
would have produced a confident failure about a tool that is fine. It is noted in the tool's
specification with "add these when that bug closes", and I left the bug to the team that holds
the reproducer rather than starting a second thread on it.

**Where that leaves the naming problem.** Of thirty tools, the one broken-and-misleading case
is now down to a single orphan — a tool component with no page under any name. That one is not a
rename, it is a decision about whether the thing should exist at all, and it needs a human.
There are also ten tools with a page and no written specification at all; that is honest rather
than misleading — nothing claims they were tested — so it is a backlog, not a defect. And worth
noting: the newest tools are arriving already correct, three in a row now, because the code path
that creates them enforces the naming itself. The problem is confined to older and ported ones.

**A recurring annoyance you should know about**, because it has now happened three sessions
running: the numbers move while I am working. The tool count went from twenty-nine to thirty
mid-session — another team created a tool ten minutes before I ran the check. I reconciled it
rather than reporting the new figure as if it were mine, and the only reason that was possible
is that the check prints its own denominator next to its breakdown.
