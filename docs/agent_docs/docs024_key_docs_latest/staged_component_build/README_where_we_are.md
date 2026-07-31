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

---

**31 July 2026, later the same afternoon**

Two things happened after I wrote the note above, and both are more interesting than the work
they interrupted.

**The first: the tests I had just attached to the tool did not work when the cluster ran them,
even though they were right.** I had checked them thoroughly on my own machine — thirty-six
checks, three times over, all passing in about eleven seconds. When the platform ran the same
checks itself, it gave up after two minutes and reported that it could not start a browser.
That error is misleading: there was nothing wrong with the browser. There is a two-minute
ceiling on the whole exercise, and the cluster is roughly ten times slower at this than my
machine is, so thirty-six checks simply do not fit. The message you get in that situation
points at the browser rather than at the size of your test list, which is exactly the sort of
thing that sends the next person looking in the wrong place.

The fix was better than a workaround. Most of those checks were being run twice — once on a
desktop screen and once on a phone screen — to test things that cannot differ between the two,
like whether the arithmetic is right or whether a button selects twenty-six reviewers. Only
four of them genuinely need both: does the page load, did the interactive part actually start,
does anything spill off the side of a narrow screen, and are there errors in the browser
console. So the phone now runs those four and the desktop runs all eighteen. **Twenty-two
checks instead of thirty-six, no assertion lost — only duplicated ones — and the run now
finishes in eighteen seconds against a two-minute limit.** It came back twenty-two passed,
nothing failed. So the tool is now genuinely, demonstrably tested by the platform itself,
which is what this was all for.

The honest lesson, which I have written into the standing files: **checking something on my
own machine proves it is correct, not that it works where it has to work.** I had proof of the
first and treated it as proof of the second.

**The second: the "orphan" is not an orphan, and my own check is why we believed it was.**
The last outstanding case was a tool component that the handoff described as having no page
anywhere, with a note saying someone needed to decide whether it should exist at all. I went
to gather evidence for that decision and found that it is live. It is on vonc.com, at
`/tools/arena/index.html`, it was redeployed a few minutes before I looked, and it serves
perfectly — I confirmed the component's own markup is present in the page the public gets.

**We were one step away from deliberating whether to delete a working tool.** The reason is
that my check asked the wrong question. It looked for a page whose *name* matched one of two
patterns it guessed at, found neither, and then reported "no page at all" — which is a
different and much stronger claim than "no page under the two names I tried". The page is
named `tool-arena` while the component is called `tool-arena-interface`, and the site also
files its pages as `arena/index.html` rather than `arena.html`, so both guesses missed. There
was a second trap layered on top: searching the published page for the component's own name
finds nothing, because this component does not print its name into its markup. Only asking the
database which page a component is attached to gives the true answer.

I have fixed the check to ask that question first, and it now correctly reports this as a
naming mismatch on a live page, with the remedy. **I have not made the change itself** — it is
another site's live page, nobody asked me to touch it, and the sensible order is to measure
what else depends on that name first, the way the earlier rename on our own site was done. But
the question has changed shape: it is no longer "should this exist" but "which of the two names
should move".

Where that leaves things: the misleading-test problem is down to that one page-naming case, and
it is understood rather than mysterious. Nothing else is outstanding on it.

---

**31 July 2026, evening — the check passes, and the rename found something**

You asked for the safer of the two arena options and for the owning team to be told. Both done,
and the check that started all of this now **passes for the first time**: no tool on the fleet
claims a test it cannot actually run. Thirty tools, none misleading. Ten still have no written
test at all, which is honest rather than misleading — nothing claims they were tested — so
that is a backlog, not a fault.

**The rename needed two changes, not one, and the second one is the interesting part.** I had
called the page-rename the safer option, and it was, but I found on closer inspection that the
page's name is also used to join it to the site's plan — and a different automated check
relies on that join to notice pages that are in the plan with no content. Renaming only the
page would have quietly dropped the arena page out of *that* check's view. Nothing would have
warned anyone: the page would still serve, the nav would still read correctly, and it would
have looked healthier only because nothing was examining it any more. **Trading one defect for
a blind spot would have been a worse deal than leaving it alone.** So both records moved
together, and I re-ran the other check afterwards to confirm the page is still in its sights.

The visitor-facing page is provably untouched — identical to the byte, same checksum.

**And the moment the page became testable, it failed — for a real reason nobody could have
seen before.** The arena has had a written test since 14 July that had never once run. It
tests that you can type into a box called `#take-input`. **That box does not exist on the
page** — the page has no input fields at all, only the site's menu button. So the test could
never have passed, and until today it never got far enough to say so.

This is the best argument yet for why the naming work came first. It was never bookkeeping:
the naming fault was *hiding* a disagreement between what a tool promises and what it actually
is. Fixing the name did not create a problem, it stopped concealing one.

**I did not decide which side is wrong, and I did not fire the test.** Either the test is out
of date (the arena became a static display) or the tool is missing its input. That is a design
question about someone else's tool, and firing the test would have filed a job for an automatic
code-fixer aimed at a page marked as hand-owned — which is precisely the wrong way to settle a
design question. I wrote it up for the team that owns it, with both readings, the exact command,
and the measurements.

**On telling them:** I put the write-up in their own folder *and* added a dated line into the
document a fresh session of theirs reads first. Their document did not mention the arena at
all, so without that line the write-up would have sat unread — this fleet has already learned
that leaving a file in someone's directory is not the same as telling them.

Nothing is now waiting on me for this piece of work.

---

**31 July 2026, later that evening — the last open item on this piece is closed, and the pod
told us so itself**

One thing was outstanding: we had changed both halves of a rule about what kinds of thing can
carry a written contract, and we could only prove one of them. The database half we could see
directly. The program half we could only *argue*: the running program was built after the
change went in, so it ought to contain it. That is the shape of reasoning this fleet has been
bitten by repeatedly, and there was a specific reason to distrust it here — the exact same
mistake has now been made three times on this one rule, and the most recent instance sat
undetected in production for two days.

**It is now proven, and better than proven — the running program printed its own list of
accepted types back to us.** The trick was to ask it two questions in one go: first the real
question, and then a deliberately nonsense one. The nonsense question has to fail, and when it
fails the program says what it *would* have accepted. So we got a straight answer rather than
an inference: seven types, `component` among them.

**The second question is the important one, and it is a habit worth naming.** Without it, a
clean result is ambiguous — it could mean "the check passed" or it could mean "the check never
ran", and those look identical from the outside. This has been the recurring fault in this
lane's work all week: something reports health it never actually measured. So the probe now has
three possible outcomes rather than two, and the third one is called VOID: *nothing was
tested*. Naming it is what stops it being mistaken for a pass.

Two useful things fell out of it. It confirmed, incidentally, that a different team's fix is
genuinely live — their list entry was printed in the same read-out, and they had only been able
to infer it the same weak way. And it did the whole thing **without adding anything to the
system**: the test travelled inside the message rather than being installed as a new agent, so
there was nothing to switch off afterwards. The one temporary record it did need was deleted;
I checked, and the database is back exactly where it started.

**One thing I want to be precise about, because it would be easy to overclaim.** What is proven
is that the *capability* works. Nothing yet actually *uses* it — there is still no real written
contract for any component, only the throwaway one I wrote to test with and then removed. Those
are two different statements and the register keeps them apart deliberately; "we can do this"
is not "we do this".

**And I found an error in my own handover notes, which had been copied forward twice.** They
told the next person what a failure would look like, and named the wrong message — there are two
near-identical checks in the code with different wordings, and I had quoted the one belonging to
the other route. Anyone following the instructions literally would have searched for text that
could never appear, on success or failure, and been left unsure whether the test had worked. It
is corrected in place, and logged in the fleet's list of wrong calls, because the useful part is
the tally: this is the second time this week a claim was written in the same voice as a measured
fact when nobody had checked it.

Nothing is waiting on me for the component-contract work. The next piece of real work in this
lane is wiring a component's test through to the browser the way tool tests already go, which is
plumbing rather than invention.
