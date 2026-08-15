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

---

**31 July 2026, evening still — one more thing before I stop, and it changes the next step**

I went to start that wiring work and did the thing this lane keeps learning to do: read the
actual code the plan pointed at before trusting the plan's own description of it. Good thing I
did. The piece that sends a test to the browser was built for tools, and it assumes something
that's true for a tool but not true for a component: that the thing being tested lives on
exactly one page. A tool does. A component doesn't — the one I checked, a teaser panel that
reveals more text when you click it, sits on five different pages across two different
websites, because that's the whole point of a shared component. So the existing wiring
genuinely cannot answer the question "which page do I test this on" for a component the way it
can for a tool. That's not a bug, it's just never been asked to do this before. It needs a
small, deliberate decision about how to teach it the new question, which I've written down as
two options rather than picking one myself.

There's also a step I'd half-forgotten was still outstanding: before any of that wiring
matters, the panel itself needs an actual written test — a real one, not the throwaway one I
built and deleted just to prove the plumbing works. That hasn't been written yet either.

Nothing is broken and nothing is urgent. I've written a fresh handoff document
(`HANDOFF_2026-07-31c_continue_here.md`) that lays out both remaining pieces in order, because
this chat is getting long and the next piece is genuinely new work rather than a continuation
of what's above, so it's cleaner to hand it to a fresh conversation than to keep pushing this
one further.

---

**2 August 2026 — the panel's real test is written, proven two different ways, and now lives in
the database**

Picked the page myself, as asked: of the five places this teaser panel appears, I went with the
one on the Leopardess consulting site, since it had been re-rendered most recently of the five.
Checked it was actually live and actually showing the panel before writing a single check
against it.

Read how the component actually behaves before writing its test, rather than guessing from the
plan document. Turns out the "click to reveal more text" part needs no JavaScript at all — it's
a native browser feature. What DOES need JavaScript is the bit where opening one card closes
whichever other one was open, and the bit where you can share a link that opens a specific card
directly. Both of those live in one shared script file used across the site, not in the panel
itself.

Wrote twelve checks covering what the panel actually promises: the panel and its cards are
there, a real click (not a shortcut that fakes a click) actually opens a card, opening a second
card actually closes the first, the text stays readable and isn't accidentally shrunk to
nothing, and nothing throws a JavaScript error. All twelve passed against the live page.

Then came the part that's supposed to prove the test isn't just agreeing with itself: deliberately
break each of the twelve things, one at a time, and confirm the right check goes red. The tool
this lane already had for that turned out to only work for a *different* component built
earlier — it was written generically-looking but was actually hardwired to that one tool's own
code. Rather than assume it would work here, I ran it and watched it fail to find almost
everything it was supposed to break. So I built a proper version of that tool for this panel
specifically. All twelve deliberate breaks were caught by exactly the check meant to catch them.

Two of those deliberate breaks — trying to click a card after making it un-clickable — took a
genuinely long time each, because the browser correctly refuses to let a "click" land on
something that can't be clicked, and it keeps politely retrying for about thirty seconds before
giving up. That's the test working as intended, just slow, and it's why this step ran long
enough to drop into the background rather than finishing in the usual couple of minutes.

Once both proofs were clean, I wrote the panel's test, its history, and today's build log into
the database properly — not the throwaway version from before, a real one that stays. Then, to
make sure writing it and reading it back actually agree with each other, I pulled it back out of
the database and ran the whole test again from that copy. Passed again, byte-for-byte the same
as what went in.

So: the piece of outstanding work from the note above — "the panel itself needs an actual
written test" — is now done. The other piece, the decision about how to point the browser-testing
machinery at one specific page out of a component's several, is still exactly where it was left:
written down as two options, not yet chosen.

---

**2 August 2026, later — a fresh build went out, checked it, nothing to worry about**

You mentioned a new build had gone out. Checked rather than took your word for it: yes,
both parts of the platform moved to a newer version, but for reasons that have nothing to
do with this project — other teams' fixes. Confirmed the database still has today's work
in it exactly as written, and the new build still understands what it needs to.

That was the last easy piece of this project. What's left — teaching the automatic testing
machinery to test one specific page out of several for a shared component — is a real change
to the platform's own code, in a part of it that goes through a review step before it ships.
That's a different kind of work to what's been done today, and this conversation has already
covered a lot of ground, so I've written a fresh handoff and I'd suggest picking it up in a
new conversation rather than continuing here.

---

**2 August 2026, later still — picked up the handoff, made the decision, and wrote the code**

Read the handoff and the docs it pointed at before doing anything. The open question was a
genuine either-or: teach the existing testing machinery a second way to find a page (by
asking it directly, for a component), or build it a small twin that only components use. I
went with the twin, mainly because the existing machinery is something every one of our
tools' automatic tests already leans on, and I'd rather add something new next to it than
add a branch inside it — if I'm wrong about something, the blast radius is nothing instead
of everything. Wrote that reasoning down properly before touching any code, as agreed.

Built it, and where the two share real work — building the request and sending it off — I
had the new one reuse the old one's code rather than copying it, so the two can't quietly
drift apart later. Checked my checking-out, not just the plan: read the whole of the
existing testing action before changing anything near it, then diffed my change against it
line by line afterwards to make sure the part that already works for every existing tool
still behaves exactly the same. It does. The project builds cleanly and its existing tests
still pass.

What's not done yet: this is a change to shared platform code, so before I ship it I want to
put it through the automated review step this project keeps offering, and then it needs to
go out in a proper build and actually get tried against the real page in the cluster — with
a deliberately wrong test thrown in alongside it, so I can be sure a "yes" really means yes.
That's the next sitting, not this one.

Submitted it for review straight after committing (small correction: I fumbled the trailer
on the commit itself — wrote a placeholder instead of the real tracking number before I'd
actually submitted. Harmless, but I've written down exactly what happened so it isn't
mysterious later, and I can't quietly rewrite the commit to fix it — that's a deliberate rule
here, not an oversight of mine).

You mentioned a fresh build had gone out. I checked rather than took it on trust, the same
habit as last time — and this time the pods genuinely hadn't moved: same two, same version,
158 minutes old. Said so plainly rather than assuming success, and worked out from the build
tag and an uncommitted file lying around that a build had probably happened somewhere but
hadn't reached this particular deployment yet. Asked you rather than guessing.

---

**2 August 2026, later still — you rolled it, I checked it, then proved it actually works**

You said you'd rolled a new image. Checked first: two brand-new pods this time, different
names, right version — a real roll, not the same ones as before. Before trusting that the
new code was actually inside them, I looked for it directly — searched the compiled program
on both machines for text that only exists if my change is in there, found it on both, and
also searched for something that definitely shouldn't be in there at all, to make sure the
search itself wasn't just rubber-stamping everything. Both came back exactly as expected.

Then I actually ran the real test: sent a request into the cluster asking the panel's page to
be checked for real, the same way the automatic system will one day do it on its own. All
fifteen checks that could run passed; the other nine were mobile-only variants correctly
sitting out because this browser session happened to be a desktop one — not a single one
silently gave up and pretended to succeed.

Alongside that, in the very same request, I deliberately pointed the machinery at the WRONG
page — a real page on the same website, just not the one with this panel on it — to make sure
it would correctly refuse rather than quietly testing the wrong thing. It did refuse, with
exactly the message I'd written for that situation, which is the part that actually proves
the safety check is doing its job rather than just happening to look safe.

So: the whole chain now works end to end, for real, in the live system — not just on my own
machine. The only thing left in this piece of work is reading the verdict from the review
step once it comes back, and doing whatever it asks if it isn't a clean pass.

---

**5 August 2026 — you've decided we take on the backlog; handoff written for a fresh start**

The review verdict came back a couple of days ago: approved, with a few advisory comments.
One of them was genuinely right — there was a slightly smaller way to build what I built,
using a hook that already existed and that I'd missed. The code that shipped works and is
staying, but I've written the lesson down where the next person building something similar
will see it.

With that closed, you made the call I'd left with you: we take on the backlog — writing
proper contracts and tests for the hundred-and-forty-odd tools and components that don't
have one yet. I've written a fresh handoff for that (`HANDOFF_2026-08-05_continue_here.md`).
The important choice in it: don't grind through all of them blindly — do about five first,
across both kinds, time them, and then come back to you with a cost-per-item so you can set
the pace deliberately rather than discovering it.

One small save along the way: the script that fires the live browser test for a component
only existed in a temporary folder that gets wiped between sessions — and it had already
been wiped. I've rebuilt it from this conversation's record and committed it properly this
time, so it can't be lost again.

---

**5 August 2026, afternoon — the calibration batch is done: four contracts written and proven, three tested live, and a price per item for you**

You asked us to take on the backlog of tools and page-pieces that have no written
contract. As agreed, I didn't grind through all of them — I did a first batch of about
five, timed everything, and here's what came out.

**Four contracts now exist and are proven.** The fuel cost estimator and the
loan-versus-savings calculator (tools), and the hero banner and the closing
call-to-action block (the two most widely used page sections on the whole estate —
roughly two hundred pages each). For every one of them: I read the real live page first,
wrote the contract against what it actually does, proved every single check can catch a
deliberate break, and wrote it into the database properly. Three of the four were then
tested for real in the live system and passed everything — including, for the two page
sections, a deliberate wrong-page test that the system correctly refused.

**The first subject I picked turned out to be genuinely broken, which is the whole point
of reading before writing.** The gas unit converter's live page renders the full
structure of the tool with every piece of text missing — no heading, no labels, blank
buttons. A visitor sees an unlabelled form. The A/B test calculator on idea.uk is worse:
it shows raw template code to visitors. Both had already been spotted by the platform's
own checks weeks ago — the repair tickets are sitting in a human-review pile, one marked
"won't fix". So I didn't write contracts for broken pages; I'm flagging them to you
instead. There's also a smaller one: every page on the gas wholesalers site asks for a
logo image that isn't there, so the fuel estimator's live test stays parked until that
one missing file is put back — the contract itself is written and proven.

**One genuinely new lesson got learned and written down where everyone will find it.**
My loan-versus-savings contract used a testing feature that another team had added to the
codebase *that same morning* — my offline checks all passed because they run the newest
code, but the live system runs a build from two hours before the feature existed, so the
live test failed on vocabulary, not on the tool. The system then dutifully opened a
"fix this tool" ticket for a tool that wasn't broken — I cancelled it with the reason
written in, reworked the contract to not need the new feature, and the re-test passed
everything. That trap is now recorded in the shared landmines file so nobody else steps
on it.

**The price you asked for:** once the tooling is warm, a simple page section costs about
15 minutes end to end, and an interactive tool about 30–45 minutes. The backlog is
currently 39 tools (only 15 of which have working pages — the rest need page fixes
first) and 111 sections. Worth knowing: newly built tools have been getting their
contracts automatically since the 2nd, so this backlog is the old stock, not a growing
pile. **The question for you is pace**: at these prices the realistic options are
(a) steady — a handful per session as a background habit, (b) a focused push on the ~15
ready tools plus the top dozen most-used sections, which is roughly three or four
sessions, or (c) everything, which is on the order of ten to fifteen sessions and hits
diminishing returns on rarely-used pieces. My own recommendation is (b): it covers
everything a visitor is actually likely to meet.

---

**5 August 2026, later — you chose exhaustive: the whole backlog gets contracts**

Option (c). Understood: every section component and every tool with a working page gets
the full treatment, at the measured prices, over however many sessions it takes. Two
scope notes so nothing surprises you later: tools whose pages don't exist or don't serve
CANNOT get contracts yet — writing one would create exactly the misleading state the
naming-contract check exists to catch (a claimed subject that hard-errors when tested) —
so those get listed for page repair instead, and the header/footer/site-level pieces stay
out of scope as before (that extension is a separate decision you haven't been asked to
make). Starting now with the most-used sections; each one gets the same
read-prove-persist-live-test sequence as the first batch, no shortcuts.

---

**5 August 2026, evening — the exhaustive run is well underway: thirty-five contracts done and live-tested in one day**

You said "everything", so everything is what's happening. Since the morning's first
batch, the production line has run four more batches: thirty-three page-sections plus
the two tools now have written, break-proven contracts, every one of them tested for
real in the live cluster including a deliberate wrong-page control that was correctly
refused every time. That covers every section component used on three or more pages —
the heroes, the calls-to-action, the article bodies, the FAQ (whose contract clicks a
real question open, like a visitor would), the contact forms, the team grids, all of it.
The pace settled at about nine minutes per component once the line warmed up, roughly
half what I quoted you at lunchtime.

Three things the line itself dug up, which is the point of the whole exercise: the
article component has no protection against wide code samples, so technical blog posts
genuinely scroll sideways on a phone — that's the component's own styling, one line of
CSS someone should add. Half a dozen sites are still serving a missing hero image that
the platform's checker found five days ago — found, flagged, and then nobody was ever
sent to fix it; that gap between "detected" and "repaired" keeps showing up and deserves
a look of its own. And a page-porting quirk: fifty-eight database records claim a
component sits on pages that, when you actually fetch them, contain no such thing —
harmless today, misleading the moment anyone trusts the records over the pages.

What's left: the one-or-two-page components (a session's worth), the sixteen
interactive ones that need their JavaScript read properly (two to three sessions), and
the remaining ten tools (one to two sessions). A few pieces belong to other active
workstreams — the darts arena, the provocations feed — and I've deliberately left those
alone rather than write contracts over their owners' heads; they're listed for a
coordination note instead. The handoff for the next sitting has the full work-list.

---

**8 August 2026 — picked the line back up after three days: the simple components are now DONE, and the blockers have names**

Re-checked everything before trusting anything (three days is a long time on this
system): nobody else had touched the lane, the missing testing feature from Tuesday's
lesson has since shipped to the live system, and the gas wholesalers logo is *still*
missing — three days on.

Today's batch closed out every remaining simple page-section that has a healthy page to
test it on: nine more contracts, written, break-proven, and passed live with the
wrong-page control refused every time. That includes the intent-probe form on the watch
site, the portfolio showcase, and vonc's archetype explainer. **Every static section
component in the estate that CAN be proven now HAS a proven contract — forty-two
sections plus two tools in total.**

What's left is exactly three piles, and each has a clear unblock:
1. **Eight sections are written but stuck behind missing site images** — the same
   `hero.jpg`/logo 404s the platform's checker found on 31 July and nobody ever
   repaired. It's now at least seven sites, including one (vetcomparison) the original
   measurement missed. One image fix per site releases its subjects; this is the
   single most valuable small repair on the board.
2. **Sixteen interactive sections** (news listings, quizzes, calculators-in-sections)
   each need their JavaScript read properly before a contract can be honest — half an
   hour to forty-five minutes each, two to three sittings.
3. **The ten remaining tools with working pages**, plus the coordination notes for the
   darts-arena and provocations pieces that belong to other active workstreams.

---

**8 August 2026, late evening — the interactive pieces have started, and the first one we
opened up turns out to be lying to visitors**

Your fresh build had landed, so I checked it before trusting it — chassis and browser
robot both on the new version, both restarted about half an hour before I started, which
is what you want (there's a rule here that nothing dispatched within five minutes of a
restart actually runs). Nobody else had touched this workstream since this morning.

The simple page-pieces were finished this morning. Tonight I started the harder pile: the
ones with their own JavaScript, where a contract has to check what the thing *does*, not
just what it looks like. Five done end-to-end: the news archive page, the homepage news
teaser, the case-studies grid, the contact block, and the blog index. Every one written
against the real live page, every check deliberately broken first to prove it can catch a
fault, then written into the database and tested for real in the cluster — all five
passed, and all five correctly refused the deliberately-wrong page I pointed at them as a
control.

**The new rule I had to invent for this pile, and it matters.** For a static page-piece,
checking that the markup is there is enough. For an interactive one it isn't: you can
delete the component's entire script and every "is it there?" check still passes, so the
contract would happily certify a dead panel. So each of these five carries one check that
can only pass if the code actually ran — the news page has to show its item count, which
is only written after it fetches the feed; the two filter bars have to actually move when
I click a filter, and I made sure to click a filter that *isn't* the one already selected
when the page loads, because clicking the pre-selected one would pass with the script
deleted. That distinction is the entire difference between a real check and a decorative
one.

**Now the thing you'll want to know about.** The contact block — the form on
robot-hands' contact page, and on two other client sites — **does not send anything,
anywhere, and tells the visitor it has.** You fill it in, you press Send, it pauses for a
moment like it's talking to a server, and then it says in green: "Your message has been
sent. We'll be in touch shortly." It then clears everything you typed. There is no
address on the form for it to send to, and the script it runs has no code in it that
sends anything at all — I checked the actual file the visitor's browser downloads, not
just our copy. The pause is a timer. It exists purely to look convincing.

Its validation is real, which is exactly why nobody has caught it: mistype your email and
it tells you properly, so the form looks wired up. I checked every other form on the whole
estate and this is the only one like it — the others all have a real destination.

I've written it up as a bug (`bugs_open/228`) with the three fix options ranked, and I've
put a warning in the shared traps file so the next person to touch it doesn't repeat my
first assumption. One deliberate choice worth flagging: the contract I wrote for that
component checks that an *invalid* submission gets a proper error back, and pointedly does
**not** check the "message sent" message — because if it did, our own quality system would
start certifying that claim as correct, and we'd have baked the lie in. We've been bitten
by exactly that before.

Two smaller finds: five case-study images are missing on finetuning.uk (the same
"we detected it and never fixed it" pattern as the hero images), and one more page whose
records say a component is on it when the page plainly hasn't got one.

Forty-seven page-pieces and two tools now have proven contracts. Left: about a dozen more
interactive ones, ten tools, and the pieces still blocked behind those missing images.

---

**9 August 2026, early hours — two more done, one turned out not to be interactive at all,
and I got a number wrong and had to correct myself**

Two more contracts written and proven live: the games index on gamesdesign, and the
AI-readiness quiz on leopardess. The quiz one is the most thorough contract this lane has
written — it presses Start like a visitor, checks the first question actually appears,
answers it, and checks the Next button unlocks. All of that had to be broken deliberately
first to prove the checks can catch a fault, and all of it passed live in the cluster.

**The games index taught me something about our own backlog.** It was on my list as an
"interactive" piece needing three-quarters of an hour. It isn't interactive at all: its
JavaScript looks for a filter bar and a "load more" button that don't exist anywhere in
the component, and no page even loads the file. Nothing is broken for a visitor — it
renders correctly as a plain list and the missing script would add nothing — but I'd have
budgeted 45 minutes for a 9-minute job. So I checked all thirty-eight pages that carry a
JavaScript-driven piece, and this is the only one like it: two pages, both on gamesdesign.
Everything else loads its script properly. Worth knowing rather than worth fixing.

**And a correction I owe you.** Last night I told you the fake contact form was on three
live pages. It's two. The third — finetuning's case-studies page — has a record saying the
form is on it, and the actual page doesn't have it. The frustrating part is that I'd been
careful everywhere else in that write-up: I went and fetched the real page and the real
script rather than trusting our database, and then for the one number that tells you how
urgent it is, I asked the database and wrote down its answer. I've fixed the bug report,
noted what caught it, and written the lesson into the shared file we keep for exactly this.

The defect itself is unchanged and still live on robot-hands' contact page and on
leopardess's quiz page. **That's the one thing I need a decision from you on**: either we
give the form a real destination (we already have two working patterns to copy from), or
we take the form off and leave the contact details, which are correct. I'd do the first,
but where the enquiries should actually land is your call, not mine.

Forty-nine page-pieces and two tools now have proven contracts.

---

**9 August 2026, late morning — the contact forms now actually work; and I duplicated
someone else's work getting there**

You asked me to enable the contact forms end to end. They are enabled, and I need to tell
you two things: what works, and a mistake of mine that you should know about.

**What works.** Both pages with the fake form — robot-hands' contact page and leopardess's
quiz page — now genuinely deliver. Fill the form, press send, and your email app opens with
the message already written, addressed to that site's own enquiry inbox. I tested this by
driving the real live pages in a browser like a visitor would, and checking the actual
address the browser was sent to: the right inbox, the visitor's name, their message, their
reply address. The screen no longer says "your message has been sent" — it says it is
opening your email app, which is the true statement — and it no longer wipes what you typed.

I also fixed the *other* contact form, the one on thirteen pages across the estate. That one
was never lying, but I measured what browsers actually do with the way it was set up and the
answer was: they may quietly drop the message. Twelve of the thirteen are now live and
correct. The thirteenth is idea.uk, which is hosted differently from the rest and hasn't
picked up the change yet — that's a deploy path to chase, not a rewrite.

**The honest limit.** This sends via the visitor's own email app. That is how this platform
has always done contact forms, because the sites are static files with no server behind
them. If you want the message to arrive without the visitor doing anything — a form that
posts straight to us — that is buildable: we already run a public API the sites talk to, and
there is a mail-sending component built and reviewed but never yet used, with contact forms
written into its own notes as the next intended user. It needs an email password we don't
currently hold anywhere and a deploy of that API, which is your call, not mine. The new code
already handles that case, so switching later is a setting, not a rewrite.

**Now the mistake.** Another session was already fixing this bug. It had a plan, had been
through the review council twice, had committed a deeper fix to the framework itself, and
had deliberately decided to wait before touching the live component. I checked whether
anyone owned this bug when I *filed* it last night — nobody did — and I did not check again
before I *fixed* it this morning. Twelve hours was enough for a whole workstream to spring
up. We independently designed the same change, character for character in one place, which
is reassuring about the design and wasteful about the effort.

I found out only because I was reading the framework code to work out why something wasn't
happening, and the comment I was reading quoted my own bug number. Someone else had written
it, minutes earlier.

I have written all of this into the bug file for them, including one useful discovery: their
change needed a full fleet release before the live fix could happen, and it turns out there
was a way round that — so the forms are working now rather than waiting. Their framework fix
is still the better long-term one and still wants that release. And where we each wrote a
different version of the same script, I have not declared mine the winner; I have left them
the test harness so they can run theirs through the same five checks and choose.

I've also written the lesson down where we keep these: an ownership check goes stale in
hours on this system, and the moment that matters is when you're about to *write*, not when
you pick something up.

---

**9 August 2026, afternoon — the release went out, and I checked it by deleting my own workaround**

Your fresh build carried the other session's framework fix. I checked it was really in the
running code rather than trusting the version label — it is, on both machines, and it
demonstrably wasn't there three hours earlier, so that's a real change and not a hopeful
reading.

Then I did the part that actually proves something. This morning, to get the contact forms
working without waiting for a release, I'd hand-written a small setting onto each of the two
pages. With the proper fix now live, that hand-written setting should be unnecessary — so I
deleted it from both pages and rebuilt them. Both still produce the correct email address.
That's the check worth having: if the framework fix weren't really doing the work, those
pages would have come back broken, which is exactly what they did this morning before it
shipped. There is now no hand-written special case left anywhere for someone to trip over
later.

I also closed the last thing I owed on this. The automated contract for that contact
component had been written cautiously: it checked that the form rejects bad input, but
deliberately did *not* check that the form works — because at the time, saying "this form
works" would have been our own quality system certifying a lie. Now that it genuinely works,
I've added the missing check: the form must have somewhere to send to. I proved it catches
both of the broken states this bug actually went through, including the one I briefly caused
myself this morning. So if anyone ever removes the destination again, the automated check
fails instead of quietly passing.

---

**9 August 2026, evening — five more interactive contracts done, planned by one model and
built by another; and the checking machinery caught a real layout bug on the way**

You asked for a plan first this time, from the other model, with implementation after. That
worked well and one part of its discipline is worth keeping: before budgeting time on an
"interactive" piece, prove it actually is one. Of the nine candidates, it disqualified four
for four different reasons — one whose script binds controls that don't exist, two whose
data feeds have never once served (so their refresh has never worked in production), and one
whose refresh redraws exactly what the server already drew.

Five contracts landed end-to-end tonight: the vendor-trust checklist (whose tick-the-box
scoring is now proven by really ticking a box and watching the score move — and whose
checkbox size check is the very measurement that exposed July's measuring bug, now standing
guard permanently), the gripper cycle-time estimator (press Calculate, a real number must
appear), the archetype quiz (answer, Next unlocks, question two arrives), the report-request
form, and the model directory. The last two are deliberately checked as static — one because
the only way to exercise its script is to file a fake sales lead into the idea.uk funnel,
which we will not do to ourselves on every test run; the other because its refresh is
indistinguishable from the server's own render. Both contracts say so in writing, so nobody
"improves" them into doing harm.

The catch of the night: the ROI estimator's page genuinely scrolls sideways on a phone — a
heading inside the tool has a hard-coded width. The new contract's own trial found it. We
have not bent the check to look away; that contract sits ready and unpublished until the
one-line style fix is made, and the defect is on your list.

Also posted a note to the team that owns the two tracker pages: their data files have never
been published, so the pages' self-refresh has silently never worked. Likely a one-command
fix on their side.

Fifty-six pieces now carry proven contracts. Interactive backlog: four left, each waiting on
something specific rather than on effort.

---

**9 August 2026 — your ruling recorded: the problems get FIXED, through the framework, and prevented at birth**

Two sessions worked this lane in parallel yesterday and today — the other one took the
interactive components (twelve more contracts, and the first real catch: contact forms
whose "message sent" was a timer, not a delivery; that's now fixed and proven on all
fifteen served contact pages). I've folded your instruction into the lane's papers as
decision D11: every defect on the standing list is now routed to the framework mechanism
that fixes it — the checkers that already detect the missing images get a repair path
instead of a flag nobody acts on, the placement records that lie get their own checker,
the broken tool pages go back through the content pipeline once its dispatch bug is
fixed, and the build pipeline gains the gates that stop these classes being born at all.
The updated handoff ends with a suggested order, cheapest unblock first: one line of CSS
releases an already-proven contract today; one image-repair programme releases nine more
plus three entries on the defect list. A fresh chat starts at
`HANDOFF_2026-08-09_continue_here.md` and has everything.

> **Correction, 9 August (later):** the fresh-chat starting point is now
> `HANDOFF_2026-08-09b_continue_here.md` — one consolidated file (state from both
> sessions, the full production-line rules, and the D11 work programme) so a new
> conversation no longer has to chain through three handoffs. Content unchanged,
> just gathered.

---

**10 August 2026, morning — the tool backlog is measured and handed to a fresh session**

Checked your new build first (right version on all machines, rolled last night), then
checked whether anything we're waiting on had moved overnight: the ROI style fix, the
tracker data files, the contact-form decision. None had — no surprise on a Sunday night.

Rather than start the tool contracts at the tail of a very long session, I've done the
measuring half properly and cut a clean handoff. The seventeen tools without contracts
split neatly: five are ready to fence immediately (both robot-hands calculators, the
games-design ranking tool, the darts setup builder, and the LLM cost calculator); nine
belong to loancalculator.co.uk, where the team already keeps a set of verified correct
answers for every tool — so their contracts should be built FROM those answers rather than
invented, and I've said so in writing before anyone does it the wrong way round; and three
are blocked — two behind the gas wholesalers' missing logo (six days now, and I've
confirmed it genuinely prevents testing, unlike the harmless missing favicon elsewhere),
one being the broken unlabelled converter that's on your list already.

The next session starts from one file, with the checklist at the bottom.

## 2026-08-10, later — the tools batch turned out to be a third the size it looked, and the first one is done

I picked up that handoff and started with its own checklist, which is where the day turned.

The list of seventeen ready tools was qualified on the wrong thing. This morning I checked
that each tool's page loads properly and has no broken images — sensible, and not the thing
that matters. What actually matters is whether the testing machinery can FIND the page.
It looks the page up by name, and the name it looks for is the component's name. **Nine of
the seventeen have pages filed under a different name, so the test would never find them —
it would stop with "no page" before looking at anything.** Had I just worked down the list,
I'd have written nine careful contracts for tools that can't be tested, and the lane's own
health check would have gone from a clean sheet to nine broken entries.

They split into two piles with different owners. One is ours and is a five-minute fix we've
done before: the games-design ranking tool's page is missing a prefix that fifteen other
pages on that same site have. The other eight all belong to loancalculator.co.uk, and
they're not a typo — the pages are genuinely named differently from the components ("car
finance calculator" vs "car finance pcp hp"). That's a decision for that team about which
name is the real one, and it's a much more concrete reason to talk to them than the one I
gave this morning. **Until somebody decides, none of their nine tools can be tested at all.**

I also found that the "five clean single tools" included one that isn't single: the LLM cost
calculator has four copies on other sites, and they all share one contract, so a contract
written against the original could fail on a copy. So the genuinely clean, unblocked set is
three, not five.

Then I took one of the three all the way through: **the darts setup builder is done and
verified in the live system.** Fifteen checks pass, none fail. I picked that one first on
purpose — a failing test can trigger an automatic rewriter, and the other two live on a site
another team owns, so I'm not firing at theirs without asking.

Two things worth telling you, because both are the system catching itself.

**The first is a mistake I made and the process caught.** I wrote a check meaning "this tool
has three questions". It passed. Then the stage where I deliberately break the page to prove
each check can fail told me that check DIDN'T notice when I removed a question. It turns out
the check type is called "count" but doesn't count — it only asks "is there at least one?",
and the number I gave it was quietly thrown away. So it would have passed with three
questions, or two, or none-but-one, for ever, while reading in our records as a firm promise
about three. I rewrote it, and I've written the trap into the shared landmines file, because
the general version — **you can write a check that asserts less than it appears to, and
nothing will tell you** — will bite someone else. Worth saying plainly: the test passed
green, and only the deliberate-breakage stage found it. That stage keeps paying for itself.

**The second is a real gap and I'd like your steer.** Every one of these tool tests has two
halves: check the page mechanically, then take a screenshot and actually LOOK at it. The
looking half has failed on **every single run in the records** — 26 out of 26 — with the
same error, that it can't reach the storage it needs to fetch the screenshots. It fails
quietly at the end, after the mechanical results are already in, which is why nobody noticed:
the runs look successful because the part we read succeeded. So all our green results are
real, but they're only ever half the test, and the missing half is precisely the one that
would catch things a selector can't see — a tool that's present, correct and invisible.

I've deliberately NOT guessed why the storage isn't reachable. That's shared infrastructure
and the rule here is to have it properly diagnosed rather than assert a cause; it isn't in
the diagnosis queue and nobody's filed it. **Do you want me to file it and put it through the
diagnosis loop?** It affects every tool acceptance run we do, so I think it's worth it.

## 2026-08-10, evening — the missing screenshot-viewer is explained, and it's two gaps wearing one error

You asked me to investigate, with the pointer that storage gets loaded into spawned
containers if the container type is on the list in the spawn code. That's exactly right,
and it's half the story.

The tool tests run in two different places. The overnight sweep spins up a fresh little
container per test — and the spawn code only hands storage access to container types on
its list, which this one isn't on (twenty of the twenty-six failures). But the manual runs
we fire by hand — including both of mine — don't spawn a container at all; they run inside
the standing service, which deliberately has no storage configured: you ruled on 8 August
that bucket access shouldn't be spread across that shared deployment, and the note in the
config file records it. So every route to a tool test arrives at the same error for one of
two different reasons. Meanwhile the screenshots themselves are fine — they're taken and
stored correctly every run; it's only the reading-them-back half that's missing.

I've written it all up as bug 243, with the evidence for each link. The clean fix for the
part that matters most — the unattended overnight sweep — is one line: put this container
type on the spawn list, which is precisely the mechanism your 8 August ruling intended for
granting storage to a specific type. That still leaves the manual runs without the
screenshot half, and there I need your steer rather than a workaround: accept it, make
manual runs spawn like the sweep does, or revisit the ruling. I'd also like to make the
loss VISIBLE either way — twenty-six consecutive failures read as successes because the
failing step's name looks like a deliberate "nothing to look at" branch, and that's the
part that let this go unnoticed.

Say the word on the one-liner and I'll put it through the review gate.

## 2026-08-10, night — the one-liner is in and under review, and the credentials question is written up

Done as asked. The one-line fix is committed and submitted to the review council — the
tool-test container type is now on the list that gets storage access at spawn time, so
once the next release rolls out, the overnight sweep's tests will have their seeing half
for the first time. The proof will have to come from a sweep-spawned run, not a hand-fired
one, and I've written down exactly what that proof looks like.

One wrinkle handled on the way: the concept register had recorded your 8 August ruling in
its strictest form — effectively "nobody may give these containers storage credentials at
all" — which would have made this fix look like a violation. Your instruction today makes
the intended line clear: no broad spreading of credentials, but the per-type list is the
proper mechanism. I corrected the register entry visibly and said why in the review
submission, so the reviewers judge the real question rather than the stale one.

And the B2 credentials point is filed as bug 245 rather than done immediately, for a good
reason: the spawner currently copies those credentials from its OWN environment into the
containers it creates. Strip them from the chassis first and every storage-using
container starts up fine and then fails at its first real operation — quietly. The
write-up puts the safe order on record: change the spawner to hand containers a reference
to the secret instead of a copied value (we already do exactly this for the GitHub
token), roll that, prove it on a real spawn, and only then take the credentials off the
chassis. Say the word when you want that executed.

---

**10 August 2026, evening — the logo turned out to be a four-month fault in our own repair machinery, and I could not honestly fix it tonight**

You asked for four things. Here they are in plain terms, and one of them did not go the way
either of us expected.

**The gas wholesalers logo.** It was never missing. The picture has been sitting on the site
the whole time under a nonsense filename — literally `input-data.asset-key.jpg`, which is a
piece of our own configuration that failed to fill in and got used as a filename instead.
The page asks for `logo.png`, so the browser gets nothing. I downloaded the mystery file and
looked at it: it is your real Gas Wholesalers wordmark, perfectly good.

The reason nobody caught this is the part worth your attention. Our system **has** been
detecting the problem — four times since April — and it **has** been running the repair, and
the repair **reports success every time**. It commits a file, says "done", and the page keeps
failing, so the detector raises it again next sweep. It is a loop that cannot ever finish. I
put it through the diagnosis loop and it came back confirmed on the first pass, agreeing
with the cause I had found.

There are two faults, and I proved the second one live by triggering a repair and watching
it happen: the repair deployed the logo as `logo.jpg` and wrote "Deploy **hero** image" in
its own commit message. It does not know a logo from a hero photograph, because the piece of
plumbing that should tell it never gets filled in. That affects 118 image records across ten
sites — though I want to be careful here: I checked the other sites and four of the five
comparable ones are serving their logos perfectly well, so that number is the size of the
mess in the records, **not** the size of the damage to your sites. Gas wholesalers is the one
actually broken.

I have **not** fixed it, and I want to be straight about why. The fix is a change to shared
machinery that every site's images go through. Making that change quietly at half past seven
in the evening, on a system where several other sessions are working at the same time, is
exactly the kind of thing our review process exists to prevent. It is written up in full with
the fix ranked by which option closes the door properly. One small confession: my test
attempt left a stray unused `logo.jpg` file on the site. It is harmless and referenced by
nothing, but it is mine and I have flagged it for removal rather than pretending it away.

**The broken converter tool.** Unparked, as you asked, and recorded as your decision rather
than something I did quietly. But unparking it does not fix it, and I would rather say so
than let it look done. The page is an empty shell — it has no content plan at all, which is
why the repair handler has twice looked at it and correctly done nothing. Nine pieces of
writing are missing, and the item tracking that has **no repair mechanism anywhere in the
system** — it only ever closes when something else happens to write the words. The thing that
would genuinely fix it is putting the page back through the full build pipeline, which
rebuilds the whole site from scratch. That is your call, not mine to take on a live site.

**The tracker feeds.** Nobody is handling them. That team's work stopped on 26 July and has
not moved since — no one is mid-fix, so it is free to pick up. I also had to correct our own
note from yesterday: we guessed the fix was "probably one command". It is not. The feature
was switched on for one afternoon in July, published the wrong file three times under the
right-sounding labels, was switched back off the same hour, and was never switched on again.
The underlying code is complete and the data is all there — 32 companies, 4 protocols, ready
to publish. It needs a configuration change, no new software, and I have written the exact
steps and the check that would catch a repeat of the July mistake into their folder.

**The email question.** Short version: `platform/mailer` is real, working, tested code that
nothing uses. To make contact forms actually deliver you need three things, and only one is
software. You need somewhere to receive the form — no such thing exists in the built system
today. You need a mail account for it to send through, and there is genuinely no email
credential anywhere in the cluster right now. And you need one real fix to the mailer itself:
it promises to give up quickly if the mail server hangs, but that promise only holds on a
connection type our hosting cannot use — on the one we can, it would sit there indefinitely.
I also checked a claim in our own notes that a public endpoint "already accepts" requests
from these sites. It does not — I tested it, and it refuses them exactly as it refuses a
made-up address. The design is sound, the word "already" was wrong, and it is one database
row per site to change that, not a code change.

## 2026-08-10, late night — 245 is done at the code level, and both fixes now wait on the same kind of proof

You asked me to go ahead with the credentials work, and the code half is in and under
review. The spawner no longer copies storage credentials out of its own environment into
the containers it creates — it now hands each container a reference to the secret, the
same way we already handle the GitHub token, and a container missing its key now fails
visibly at start instead of quietly at first use. I checked all four keys exist in the
secret before writing a line.

What I have deliberately NOT done yet is take the credentials off the chassis itself.
That's the half you actually asked about, and it must wait: the new spawner code has to
reach production first, and then I want to see a spawned container actually complete a
real storage operation — not just start up — before the old supply line is cut. A
colleague's note on the bug file made that bar sharper with live evidence, and they're
right. Once that proof lands, the removal is a small config edit with a checklist already
written.

The screenshot fix from earlier is in the same position: the new build you deployed was
made after my change went in, but this particular change leaves no fingerprint I can
check in the binary, so the only honest proof is watching tonight's sweep produce a test
that keeps its seeing half. The query to run is written down; first thing next session.

Everything a fresh chat needs is in one file: HANDOFF_2026-08-10b_continue_here.md.

---

## 2026-08-10, evening — two more tools locked down, and a check we trusted turned out to be checking less than we thought

Two chats got handed the same instruction sheet today. The other one had already been going a
couple of hours, had finished the darts setup builder and had moved on to a separate problem the
owner asked about. Rather than quietly do the same work twice, I wrote down who had what — in the
instruction sheet itself, since that is the one page both chats read — and took the two tools it
had deliberately left alone: the grip-force calculator and MatchMatrix, both on robot-hands.

**Why it had left them alone, and why it was right to.** When one of these contract checks fails,
the system does not just report it — it can dispatch an automated rewriter at the page to make it
pass. Both robot-hands pages are marked as belonging to another workstream. Pointing a robot at
someone else's page because your own check went red is not a good way to behave.

So before doing anything I checked whether "belongs to someone else" is actually a stop sign. It
is not: eight tools already under contract sit on pages marked that way. What matters is that
there is a proper switch for this — a flag on the contract that says *if this fails, tell a human,
do not send the rewriter*. I confirmed that flag exists in the running system (and again after a
fresh build landed mid-session), and both new contracts carry it.

**The part worth telling you about.** These contracts work by driving the tool and checking the
numbers that come out. I had assumed — and this lane had written the same thing down in writing
this morning — that this doubles as proof the tool actually *showed* the answer to the visitor,
because a hidden element reads as blank. It does not. I proved it does not, twice: once by
breaking the "show the results" step and watching every number still check out, and once by
hiding the panel purely in the styling, changing no code at all. Same result. The browser hands
back the text of a hidden panel quite happily.

Nothing shipped wrong because of it — both new contracts already had a separate "and the visitor
can see it" check, and so does the darts one, for a different and genuine reason. But the *reason*
we had written down was wrong, and that reason was about to be copied into the next dozen
contracts. It is now recorded in the fleet-wide traps file so the next person gets it for free.

**And a smaller thing that bothered me more.** The tool we use to prove a contract can actually
fail prints a line at the end saying "checks watched red: 13 of 13". It turns out that line was
counting what the *author had claimed* each test would break, not what actually broke. So a
mis-written test file would confidently certify itself as fully covered. It only misleads when a
run fails, which is why nobody had caught it. Fixed — it now counts what was really observed.

**Where the two tools ended up:** both pass their full contract on the live page, both had every
single check individually proven capable of going red (13 deliberate breakages each, all caught),
both contracts are stored and read back byte-for-byte identical, and both passed the real
end-to-end run on the cluster — twice, since a fresh build rolled halfway through and I did not
want to quote evidence from the old one. That takes the lane to 59 subjects proven.

**One mistake of mine, recorded properly.** I overwrote the other chat's file because I assumed a
file with an obvious name was mine to create. Git had a copy, so it cost nothing but a few
minutes — which is exactly the argument for committing early and often on a tree this busy. The
near-miss was worse than the mistake: the next step would have quietly replaced their finished
work in the live database with a fresh copy of itself, destroying the evidence trail they had just
written down. Both are written up.

## 2026-08-11 — the acceptance tests got their eyes back, and the first thing they saw was real

**The fresh build proved both of yesterday's fixes this morning, properly.** Rather than wait
up to a week for the overnight scheduler to happen to pick a tool (it deliberately skips
anything tested in the last seven days, and we tested everything on Sunday), I queued one
acceptance run by hand for the darts setup-builder and watched it go through the machinery.
A dedicated worker pod was spun up for it, exactly as the fix intended; that pod finally had
the storage settings it has always been missing; and its credentials arrived as *references
to the vault* rather than copied-out strings — which was the other fix. The run finished
clean, all fifteen checks green.

**The part worth saying out loud: the screenshot-judging half of acceptance ran for the very
first time anywhere, and on its first look it found a genuine problem.** On the darts
setup-builder page, several of the answer buttons — and the main "Get my recommendation"
button — have text so low-contrast it is close to invisible, on desktop and mobile alike.
Every one of our selector-based checks passes on that page, because the text is *there*; you
just can't read it. This is precisely the class of defect we built the vision half for, and
it justified itself in one run.

**Two loose ends from that, one of which needs a decision.** First: the vision verdict is
currently written into a database field nobody reads — the run stayed green and raised no
follow-up work despite finding that defect. Making a vision finding visible (bug 243's
candidate 3) now has a worked example arguing for it. Second: the darts page itself needs its
contrast fixing — that belongs to the fixloop/darts lane, and it is written where they will
find it.

**Housekeeping completed while I was in there:** the chassis no longer holds the storage keys
it could never use — the credential block came out of the deployment config once the new
spawn path was proven end-to-end (the keys now live only in the vault and in the pods that
genuinely need them). One straggler of the same shape was found and noted rather than fixed:
the Firecrawl key still gets copied into worker pods as a string. Same disease, different
key, deliberately left for its own decision rather than smuggled into this one.

---

**2026-08-11, later the same morning (a second chat, running in parallel).** Two of us were
given the same handoff again and both drove the same proof — I got there about an hour
behind and only found out when the other session's commit appeared in the log mid-way
through my own measuring. Nothing was lost, and their write-up of the proof stands, so I
deleted my duplicate version of it rather than leave two accounts of one event. But it is
the second day running that this has happened, and it is worth saying plainly: the only
thing that tells one chat what another is doing is a line written into a file, and a line
written at 09:41 cannot mention a commit made at 10:48.

What I did add is a measurement, and it changes two of the decisions in front of you.

**First, the contrast defect the machine's restored eyesight found is worse than it looked,
and it is not the site's fault.** The other session sensibly guessed this might be the
palette bug we already know about — the one where a site's colours get churned by a generic
theme. It isn't. The site's colours are exactly what they are meant to be. The fault is in
the *tool component* itself: when you pick an option, the component paints the selected
button with the site's "primary" colour and then writes the label on it in the site's
"surface" colour — on the assumption that surface is always a pale background shade that
will stand out against primary. On dartsonline both of those colours are near-identical
dark navy. The measured contrast is **1.06 to 1**. The accessibility floor for readable text
is 4.5 to 1. The correct pairing on that same page would have been 14.65 to 1. So the text
is not "hard to read", it is invisible, and it has been for as long as the tool has existed.

I then checked how far the pattern goes, because a component is shared. The same idiom is in
**9 components across 7 tool types, live on 8 sites**. Six of the eight are perfectly
legible — which is the real point: *the pattern is not wrong, it is unguarded*. It happens
to work wherever a site's surface colour is pale, and fails silently wherever it isn't. The
second casualty is **mortgagecalculator.co.uk**, at 2.95 to 1, on two of its calculators.
That is below the floor too. So this is no longer a darts ticket to hand to one lane — it is
either a fix to the shared component, or a check at build time that a site's palette
actually satisfies what its components assume. **That is a scope decision, and it is yours,
because it reaches into two lanes' sites.**

**Second, and I think more important: the eyes we just restored are wired to nothing.**
Yesterday's fix means the acceptance agent can finally look at screenshots, and on its very
first look it found a genuine defect that all fifteen automated checks had passed. Then it
wrote that finding into a note category that **no code in the entire platform reads** — I
grepped for it and got zero hits — and, in the same second, stamped the page **PASSED**. No
alarm, no work item, nothing for anyone to pick up. The finding only exists today because a
human happened to be watching the run.

So the decision the other session put to you as "should we build the thing that makes vision
findings visible?" is sharper than it was. It isn't a nice-to-have on top of a working
system. Until it is built, the honest description of what we fixed yesterday is: **we gave
the acceptance tests their eyes back, and they are reporting into a void.** My
recommendation is to build it next, ahead of finishing the remaining batch-8 tools, because
every acceptance run between now and then produces findings nobody will ever see.

## 2026-08-11 (later) — your decisions actioned: two done today, one specified for next time

You made four calls this morning and three are already real. The Firecrawl key no longer
travels as a copied-out string into worker pods — it joined the same vault-reference list
as the other API keys, which as a bonus fixes a second, quieter bug: the *other* pod
spawner had never been handing the key out at all. The change is committed and submitted
for review; it takes effect at the next build roll.

The manual test trigger is reshaped as you chose. Firing an acceptance test by hand now
goes through exactly the same machinery as the overnight scheduler — a proper worker pod
with working storage and eyes — instead of the shortcut path that silently skipped the
screenshot check. I tested the rewritten trigger for real: it spun up a fresh worker,
ran all fifteen checks green, and the screenshot pass ran too. It also now refuses
politely up front if someone points it at a tool that can't be tested, instead of
appearing to succeed while testing nothing.

The third call — making the screenshot judge's findings visible instead of letting them
vanish into an unread database field — is the biggest piece of work, so it's written up
as a precise brief for a fresh session rather than squeezed in at the end of this one.

The three site problems are laid out in the chat for your decisions: the loan calculator
site (where the better fix turns out to be teaching the tester to use the page's address
rather than renaming eight finished pages), the invisible-text defect (two sites failing,
a shared component's assumption at fault), and the gas wholesalers' missing logo.
Whichever way you call them, I'll carry the word to each lane's own paperwork.

---

**2026-08-11, afternoon — the same parallel chat, after you decided all five.** Everything
you decided this morning is now either done or waiting only on the next release.

**The invisible text is fixed and live.** Nine tool designs across eight sites shared the
bad colour assumption; I changed all nine in one migration to a pairing each site itself
guarantees — the colour its own body text uses — and re-published the affected pages. Every
page now serves the fix except one: the gas wholesalers estimator, whose page is marked as
owned by the tool pipeline, and the system correctly refused a generic overwrite. That page
was never illegible (its colours happened to work), so nothing is broken there; the fixed
design simply ships the next time the tool pipeline rebuilds it. Two things came out of the
verification worth knowing. First, the platform already has a proper "text that goes on the
brand colour" token — the tool designs just weren't using it. Second, on
mortgagecalculator.co.uk that token is itself wrong: white text on their gold, below the
accessibility floor, by the site's own declared palette. That is a one-token fix with
site-wide effect, and it belongs to that site's lane — I have written it up in their
directory with the numbers and how to undo my change if they prefer.

**The eyes now have somewhere to report to.** The new piece reads the vision critique and,
when it says it found something, files one work item addressed to a human — deduplicated,
so a defect re-found every night updates one ticket rather than minting thirty. Two design
choices you should know about: it can never touch the pass/fail verdict (the checks decide
that; the eye only adds), and if the critique is garbled the system files anyway rather
than staying quiet — after twenty-six silent losses, the failure mode we will accept is a
human occasionally glancing at a non-issue, never silence again. The code is committed and
before the council; it takes effect at the next release, plus one held config switch that
must be applied after it (the file says exactly when and how).

**The loan calculator tools are unblocked without renaming anything.** The missing config
key is in: an acceptance run can now be told which page it is about instead of guessing
from the name. Nothing changes for any existing run until someone actually passes a page
address, so it is safe by construction. The renames are now a tidiness question for that
site's lane, not a blocker — and one of their tools lives on the homepage, which no rename
could ever have fixed.

The other chat did the manual-trigger rework and the Firecrawl key this morning; both are
recorded in their own right. Two council verdicts are outstanding and should be read by
whoever is next in: theirs for the key, mine for the vision piece.

---

**2026-08-11, early afternoon — after the fresh build went out.** The new release carries
the vision-findings piece, so I switched it on and put it through its first real run.
Everything passed, and the run itself was quietly satisfying: the machine looked at the
darts setup-builder page — the one whose invisible text started all this yesterday — wrote
a short critique for a human, ended with the new machine-readable "nothing to report" line,
and correctly filed nothing. Better still, its prose confirms the colour fix with fresh
eyes: no contrast problems on the page that failed the measurement a day ago. So the whole
chain now works: the eyes see, the verdict stays the checks' business, a clean look files
nothing, and a defect (when one next occurs) files exactly one ticket for a person.

The chassis credential question is also fully closed: the new build's standing pods carry
zero storage keys — the last box on that bug is ticked.

One piece of honest process news: the reviewer council asked for changes to the
vision-findings design before approving it. Two of their three points made the work better
and I have done them — if the system ever fails to file a finding, it now leaves a durable
written trace rather than just a log line; and the paperwork gap they spotted is fixed.
Their biggest point was fair in spirit: "you are filing tickets into a queue — does anyone
actually look at that queue?" I checked: the dashboard does show these items with a proper
review flow, and the old bug where the queue falsely displayed as empty is fixed. What
nobody has yet built is a routine that brings the queue to a human on a schedule — that is
a known, separate open item, and it affects every kind of ticket, not just these. My view,
stated to the council: better to fix that once, for the one shared queue, than to invent a
private notification channel per feature. Their second verdict is due shortly; whoever
picks up next should read it.

## 2026-08-11 (end of afternoon) — all four decisions are done, and most of them were done by the other chat

Short version: everything you decided this morning is now live, and I did about a third of
it. The other session working this same lane got to the two big pieces first — the
screenshot-findings work and the invisible-text fix — and had them built, reviewed, shipped
and proven while I was still reading the code for the first one. That is a good outcome and
worth being plain about: no work was duplicated, because I checked before starting.

**What is now true.** The screenshot judge's findings are no longer written into a void —
there is code that reads them and raises a visible item, and it has been proven not to raise
one when the page is clean. The invisible-text defect is fixed at the shared component
rather than on one site: nine templates changed, eight pages re-rendered and measured
legible, and the ninth deliberately left because that site's page is owned by a different
process and was legible anyway. The loan-calculator problem is solved the way we hoped — the
tester can now be told which page to look at, so nothing on their site has to be renamed,
including the calculator that lives on their homepage and never could have been. And the
Firecrawl key is a vault reference now, approved on review and confirmed in a live pod.

**My own share of it**, for the record: the manual test trigger (now going through the
proper machinery, proven three times today), the Firecrawl conversion, the piece the other
session explicitly left to me — teaching that trigger to name the page it means, which is
what actually makes the loan-calculator route usable by hand — and the four notes to the
other lanes carrying your decisions.

**One honest limitation.** My last test run proves the new page-naming route does not break
anything; it does not prove the route was actually *used*, because for the tool I tested,
both routes point at the same page. The test that settles it is the first loan-calculator
tool someone writes a contract for. I have written that down rather than let a green run
imply more than it shows.

**And two mistakes of mine, both caught the same session.** I read a database setting as
missing because I looked it up under the wrong name — it had been added an hour earlier by
the other chat — and I was minutes from rebuilding work that was already finished, precisely
because having your decision in hand made it feel unnecessary to check whether anyone else
was on it. Both are written into the fleet-wide mistakes log, along with the one-line check
that catches each.

**2026-08-11, later — a fresh session, picked up from the handoff.** Good news and one
correction. The vision-findings work (item 1 in the last handoff) went through a second
round of automatic review and got sent back again, for a genuinely useful reason: the
explanation I'd written for "who checks these findings later" said flatly that nothing
does it automatically. That turned out to be wrong — there IS a daily automatic check, it
really does run every day — it just doesn't happen to cover this new kind of finding yet.
So the honest story is narrower than what I'd written: not "nothing happens", but "the
thing that happens doesn't reach this yet". I've corrected that rather than argue with the
reviewer, and left the question of whether to teach the daily check about vision findings
as a separate decision for later, since it's not obvious how a computer would re-check a
"this looks ugly" judgement the way it can re-check "is this field still empty".

The other three points the review raised were all versions of the same small thing: when
this code fails to save a finding, it should tell someone in the ONE place the platform
already has for that ("something went wrong here, durably"), not invent its own note for
it — especially since the note it was inventing turned out to share its filing cabinet
with an unrelated, routine note, so a human could never have told the two apart. Fixed
that properly rather than patching around it — small change, tested, resubmitted for
review on the same review thread as before. Waiting on that verdict now.

**2026-08-11/12, a separate session, picked up the same handoff independently.** Found
the same round-2 rejection above, but by the time I'd worked out what it meant, another
session (the one who wrote the entries just above) was already answering it. Rather than
race them to a fix, I set up a simple watch — check every 15 minutes, read-only, no edits
— and let it run. It paid off quickly: within the first check, that other session's fix
went to review a third time and was approved and shipped. So the vision-findings feature
is now properly finished, reviewed and live; nothing left open on it.

After that the watch mostly reported quiet — many checks in a row with nothing new,
across a long overnight stretch — until you told me a fresh build had gone out and asked
me to pick back up. I checked: the new build does include everything from this lane, so
nothing was lost or rolled back. But in that whole overnight stretch, the two remaining
jobs on this lane's list — a calculator page and a ranking page that still need their
"does this actually work" checks written — hadn't moved at all, even though several other
sessions had clearly glanced at them along the way. So they're genuinely free to pick up,
not secretly claimed by someone who just hasn't committed yet.

Given how much of this session ended up being about the watching rather than the doing,
I've written a fresh handoff for whoever (or whichever fresh version of me) picks this up
next, with the two remaining jobs spelled out plainly enough to start straight away rather
than having to re-read this whole story first.

**2026-08-12, one of those two jobs got finished.** The ranking-page tool on the games
site had never once been checkable, for a small and slightly silly reason: the page was
called "bayesian-ranking" but the checking system was looking for a page called
"tool-bayesian-ranking" — one word short of matching, so every attempt to test it had been
quietly finding nothing to test rather than failing loudly. Renamed the page (a one-line
change that doesn't affect what visitors see — checked the page byte-for-byte before and
after, identical), then wrote a proper contract for what the tool should do: two products
being compared, a slider for how much you trust small sample sizes, and a score that
should flip which product "wins" once the underpowered one gets enough data behind it.
Proved the contract by deliberately breaking the page seven different ways and checking
that each break got caught by the right test — including one break I invented on purpose
where a "winner" badge should have gone away after a recalculation but I made the code
forget to clear it, which is exactly the kind of subtle bug this whole exercise exists to
catch. Then, rather than stop at "it works on my machine", I fired the real test off at
the actual live cluster and watched it come back a minute later: fifteen checks run,
fifteen passed, nothing failed. That page is now genuinely, provably working, not just
believed to be.

While doing this I noticed another session had, in the meantime, spotted my work in
progress and sensibly picked up the OTHER remaining job (the cost-calculator tool)
instead of getting in my way — so both jobs are now moving, by two different sessions,
without anyone duplicating anyone else's work. I've left that one alone entirely.

---

**12 August 2026 — the logo repair is fixed in code (not a handler, as it turned out), and two questions answered with a measurement rather than an opinion**

**The repair machinery.** You asked whether we needed a handler. No — and that is the
useful part of the answer. The thing that does the work already exists, already runs,
already commits the file and already tells the truth about what it did. What was broken was
what it got TOLD. Two separate pieces of information were going missing on the way in, and
the machine had no way to notice either was absent, so it filled the gaps with guesses:
it guessed the filename from a piece of its own configuration, and it guessed that every
image was a large landscape photograph. Adding a handler would have given us a fourth thing
to get the inputs wrong.

Both gaps are now closed, and closed in a way that cannot come back: the filename guess is
deleted rather than switched off, because there is no way for a caller to use it safely;
and the machine now asks the database what an image actually IS when nobody has told it.
There is also a new general safeguard: the system can now tell the difference between
"someone chose this setting" and "nobody said anything so we used the default". That sounds
small. It is the reason logos were being published as photographs for four months, and it
almost certainly affects other things we have not looked at yet — I have written down that
nobody has checked which.

I proved each fix by deliberately breaking it again and watching the alarm go off. Three
breakages, three alarms, all restored. **The change is committed but does nothing until the
fleet is rebuilt and rolled — which is your command to run, not mine.** After that roll the
gas wholesalers logo should appear, and I have written down the exact check, which is to
look at the file extension rather than at any success message. One caveat worth having: the
roll stops NEW files being misnamed; the 150-odd already sitting under wrong names do not
repair themselves, and draining those is a separate job nobody has designed yet.

One deliberate choice you may want to overrule. There was a faster route — a configuration
change that would have fixed half of it immediately with no rebuild at all. I did not take
it, because it meant widening a piece of shared plumbing that every job in the system passes
through, in order to fix one job. Narrower and slower beat wider and faster. If you would
rather have it working today than working safely, say so and I will make that change instead.

**The tools API.** You asked me to open it up to the other sites "if we need that". We do
not, and here is the measurement rather than my opinion: across the whole fleet exactly two
things call that service, both on vonc.com — and vonc.com is already the one site allowed
in. The permission list matches reality exactly. I also checked the thing that would have
mattered for contact forms: there is no address on that service for a contact form to send
to. I tried three plausible ones from the site that IS allowed in, and all three came back
"no such thing". So opening the door would not have moved contact forms one step forward;
it would only have handed two sites access to an expensive service they never call. I have
left it alone.

**The email sender — what I need from you.** Three things stand between us and working
contact forms. Only one of them is code, and the other two are yours.

*First, and this is the blocker: an email account for the system to send through.* There is
no email credential anywhere in our cluster today — I checked all three of our secret
stores. Our own notes from an earlier attempt say the ordinary route does not work here:
our hosting provider blocks most outgoing mail, and shared mail relays filter this kind of
message as spam. The recommendation on record is a dedicated Amazon SES sending account,
with the sending domain properly signed so mail is trusted. That is an account to open and
a couple of DNS records to add — a person's job, not a program's. There is one trap already
written down from last time: the SES username is the access-key string, not the account
name, and using the wrong one fails with an unhelpful error.

*Second, a decision: where the form actually posts to.* Two options. Put it on the existing
small service that already runs on its own machine — it already handles multiple sites and
would need one database row per domain, no new infrastructure. Or build a new service
inside the main cluster — cleaner in the long run, but it needs a build target, a
deployment, a secret and a public address, none of which exist. I would take the first;
it is days rather than weeks, and it is reversible.

*Third, the code, which is ours and is small.* The sender itself is written, tested and
used by absolutely nothing. It has one real defect I verified myself: it promises to give
up quickly if a mail server stops responding, but that promise is only wired into one type
of connection — and it is the type our hosting cannot use. On the type we must use, it
would wait indefinitely, which on a live web form means the visitor's browser hangs. That
is a small fix, but it must happen before this is put anywhere near a real form.

The order matters: the credential first, because until that exists the rest cannot be
tested against anything real.

**2026-08-14.** Picked up one of the other open items you asked about, and it turned into a
longer story than expected. The mailer thread from the note above turned out to be a dead
end for now — the actual contact-form bug is already fixed and working, I'd just found a
stale leftover comment saying otherwise. So instead I picked up an older bug: for months,
one client's site (Gas Wholesalers) has been showing a broken image where their logo should
be. The real file was there, just filed under the wrong name, because of a small mix-up in
how the system decides what to call a freshly generated image. A fix for that had already
been written and was waiting for review.

That review turned into four rounds back and forth with the automatic reviewers, and each
round was worth having: the second round found that the SAME small mix-up existed in a
second place nobody had checked; the third round asked me to double-check two very specific
"is this actually the code that runs" questions, both of which came back fine on inspection;
the fourth round raised a genuinely bigger question — not "is this fix wrong" but "should
this whole category of shortcut be allowed at all, anywhere in the system" — which is a
question for you, not something a few more rounds of automatic review can settle. I've
written it up rather than continuing to argue it in circles.

While all that was happening, a new version of the software went out on its own, and it
included the actual code fix. So rather than just trust that it worked, I found the exact
broken image, manually told the system to try deploying it again, and watched it happen.
It worked — the logo now loads correctly, for the first time in months, and I checked it
myself in a browser rather than just reading a success message. One more site with the
same problem still needs the same check; I've written down exactly how to do it. And the
roughly 150 other images sitting under the wrong name fleet-wide won't fix themselves —
that's a proper clean-up job someone still needs to plan out, not a quick fix at the end of
a session.

---

**14 August 2026, late — coming back after two days: the logo fix landed and was proven, and re-counting the leftovers found two ways the count lies**

I picked this back up after two days away and the first job was to find out what had happened
without me. A great deal had: other sessions took the repair fix through four rounds of review,
answered every objection, shipped it, and proved it working on both of the sites that were
broken. **The gas wholesalers logo is live. The mortgage calculator's missing header image is
live.** I checked both myself rather than taking the note's word for it, and I also checked that
the fix is genuinely inside the software that is currently running — not merely written down —
by asking the running program directly and including a deliberately fake thing to look for, so
that a "yes" means something. Real yes, fake absent.

One thing I should own: when I made the change I deliberately avoided a wider adjustment to some
shared plumbing, on the grounds that it was riskier than the problem justified. The reviewers
disagreed and marked it a high-priority objection, and a later session made that change. They
were right and I was over-cautious. Worth recording, because it is the review process doing
exactly what it is for.

The remaining work is a clean-up: a batch of image records left pointing at the wrong filenames
by the old bug. Someone ran a careful pilot on the first eleven of them yesterday evening — all
eleven clean. That took the count from 140 down to 98, and my job today was to re-count the rest
properly before anyone runs the next, much bigger batch. **Two things came out of that, and both
mean the list is not what it looks like.**

**First, eleven of those 98 records must never be touched.** They refer to images that have since
been *replaced* by newer ones. Re-publishing them would push an old picture back over a current
one. That is worse than leaving them alone, and — this is the awkward part — the safety check the
pilot used would not catch it, because a genuinely outdated image is *supposed* to be missing, so
the record looks like honest outstanding work. The real list is 87, not 98.

**Second, the two items that look most obviously "still to do" are the two the pilot deliberately
decided NOT to do.** They are logos on two sites that are already displaying correctly. The
person running the pilot spotted that publishing them would overwrite a working logo, and skipped
them on purpose. But nothing in the records says "we decided to leave this" — so the next person
along sees two easy items with a proven fix and does the damage the last person avoided. I have
written that in three places now, because it is the kind of thing that only bites once you are
confident.

I have not run any of the clean-up. The next batch is 57 items across live sites, and one site
alone accounts for 28 of them, so a single-site trial run is both cheap and informative. That is
a real change to live sites and I would rather you knew it was happening.

Two older things I promised to keep an eye on have **not** moved, and I re-checked rather than
assumed. The two industry tracker pages still have no data files — that team's work stopped on
26 July and nobody has picked it up; the fix is a settings change with no new software, and the
instructions are sitting in their folder. And the broken gas converter tool is still parked
awaiting your decision, for the same reason as before: the page has no content plan at all, so
the automatic repair correctly does nothing.

Last, a tidy-up I owe: the stray unused image file I left on the gas wholesalers site during
testing on the 10th is still there. It is harmless and nothing points at it, but it is mine and
it should come off.

## 2026-08-15 — you gave the go-ahead, and the clean-up turned out to need no publishing at all

You answered the open decisions this morning: run the whole clean-up (after checking nobody
else was working the same pages), do the record-keeping fix by hand this once, and — on the
question I couldn't answer — you confirmed there IS a site lock system, and none of these
twelve sites needs locking while we're in heavy development.

Before touching anything I checked for other in-flight work: another thread is re-running the
call-to-action links across the fleet today, but that work rewrites pages, not image files, so
the two cannot collide. Nobody else was touching the image backlog. The fix itself is still in
the running software — I checked the actual binary again, with controls.

Then the surprise, and it is a good one. Before publishing anything I checked every single one
of the 84 outstanding items at the live sites — does the correctly-named image already serve?
**All 84 do.** Every page had already been repaired by the ordinary day-to-day rebuilds since
the fix went in; only our own records never caught up. So the "clean-up across eleven live
sites" turned out to be: publish nothing, correct 85 database records, and cancel three stale
to-do entries that would have tempted a future session into re-publishing over working images.
I kept a before-copy of every record I changed, in this folder, in case we ever want to look
back.

I also double-checked the checks themselves: I asked each site for an image that definitely
should NOT exist and confirmed they all said no. So "all 84 said yes" means the images are
really there, not that the sites say yes to everything.

The placeholder-image saga is therefore DONE: the bug is fixed in the running software, both
original broken pages have been right for days, and the backlog is at zero. I have moved the
bug file to the closed pile.

The two policy questions you asked me to settle are settled and written down. For the search
mechanism (the one that guesses where a field's value lives): my ruling, recorded in the RFC,
is that the guess is only allowed when it is not really a guess — if the search finds exactly
one candidate it may use it, and if it finds several different candidates it must refuse and
say so loudly, rather than pick one. We measure first, then tighten. The resolver questions
from the other thread you had indeed already answered — I've marked that file as ruled so
nobody re-litigates it.

Still with you / still parked: the gas converter tool now has your ruling (build a proper
repair handler fleet-wide) — I've filed that as its own numbered piece of work. The tracker
feeds: the lane to wake is `model_directory_pipeline` — the instructions are in
`model_directory_pipeline/FINDING_2026-08-10_the_tracker_publisher_was_reverted_and_never_re_extended.md`.
The stray image on gas wholesalers stays put for now, per your word.
