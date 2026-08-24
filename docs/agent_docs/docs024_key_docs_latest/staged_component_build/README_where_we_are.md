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

**2026-08-15, afternoon.** The resolver ruling from this morning is now built. In plain
terms: when a workflow step needs a value nobody explicitly wired up, the platform used to
wander the whole run's data looking for anything with the right name and take the first thing
it bumped into — and "first" was literally random. Now it gathers every candidate, and if
they all agree it uses that value; if they disagree it still picks the same one every time
(the shallowest) but writes a warning naming every candidate, so after the next release we
can watch for a week and see exactly which pipelines were living on luck. The follow-up step
— refusing to pick at all when candidates disagree — waits for that week of evidence. There
is also a new opt-in "!" mark an author can put on a wiring line meaning "this exact source
or fail loudly, never guess"; the image pipeline's asset id is queued to be the first user of
it once the release is out (the switch is written but deliberately held until then). One of
the ruling's two named first users turned out to be unusable on its own earlier evidence — a
shared dispatch line where the optional mark is doing real work for hundreds of other item
types — so that half is recorded as corrected rather than done. Two long-standing flaky tests
were repaired along the way; both had been asserting the winner of what was actually a coin
flip. Everything is committed but changes nothing until the next release rolls.

**2026-08-16 (morning) — the reviewers said "revise", they were right, and the fix is built.**
Yesterday's review of the "never guess" change came back with one objection that actually
mattered: the warning we added to catch pipelines living on luck was only ever written to the
pod's log, and those logs last about a minute and a half. So the week-long watch we were
counting on could never have been read afterwards — we'd built a smoke detector with no
memory. The platform already has a table for exactly this kind of durable record, so this
morning every one of those warnings is also written there, permanently, one row per
occurrence, with the field name and every candidate that was found. That is written, tested
(including proving the tests fail if the write is removed), and committed; it goes live on
the next release. Two side findings: the first half of the change DID already go out in
yesterday's release, so it is running now (silently, which is the problem just fixed); and
the "!" strict-mark for the image pipeline is therefore also ready to be switched on — its
pre-flight checks were run this morning and passed, so that switch is the next thing to
apply. The reviewers' smaller points (a missing entry in our own decision log, two checks on
the switch's paperwork, a note the review board can read in the database rather than in a
file it cannot see) are all done, and the whole thing has been resubmitted for a second look.

**2026-08-16 (late morning) — the second look came back approved.** The review board passed
the fix on its second round, with three minor notes and nothing serious; the two that were
real questions ("could you have used an existing helper?" and "could two agents in one pod
mislabel each other's rows?") were both checked against the code and answered no, with the
reasons written down. Two things need a human hand: the image-pipeline strict switch (ready,
checks passed, but this session was not permitted to change live configuration — the exact
command is in the handoff), and the next release, which is what turns the new record-keeping
on. One confession: I sent the same review request twice by re-running the submit script to
re-read its output — it cost a duplicate review, and the lesson is written where the next
person will trip over it.

**2026-08-16 (afternoon) — the new release is out, the recorder is on, and it is already
telling us something.** You asked for the strict-switch migration; it's in the chat with its
checks, yours to run. More important: within four and a half hours of the release the new
record shows about 670 cases where the "guess between look-alike fields" search actually met
disagreeing candidates — mostly inside the long-running dispatch loop, where one field name
can have twenty to ninety look-alikes by the time it is read. So the plan's next step (stop
guessing entirely) is off the calendar until each of those cases has been given a proper
explicit wiring line — which is precisely what the week of watching was for, and it says
the "conflicts are rare" hunch was wrong. That triage is the next job on this lane.

---

**16 August 2026 — answering your build-pipeline question, and why I stopped before deleting the file**

Short answers first.

**"Will the full build-pipeline affect other pages too?" Yes — all 36 of them.** The full
pipeline re-plans the whole site: it re-runs the site planner and then rebuilds content page by
page. For one broken tool page that is far too blunt an instrument, and it costs LLM spend on 35
pages nobody asked about. There is a narrower mechanism that rebuilds *only* the pages you
explicitly flag, and that is the one to use. So: we do not need the big pipeline, and I would
recommend against it.

**Good news on the tool: the guide page already exists and works.** It was rebuilt on the 15th
and has real content. So "rebuild it with a guide" is already half done — what is missing is the
interactive converter itself, which is a page with a slot and nothing in the slot. Specifically
nine pieces of text the system was supposed to write and never did. The cleanest fix is to get
those nine written by the tool-building machinery, without re-running anything site-wide; if
that does not work, the fallback is the narrow one-page rebuild.

**On the stray image file — I did not delete it, and I want to be straight about why.** I traced
the whole mechanism and hit something worth telling you: *the system has no supported way to
delete an image file from a site.* The one deletion route that exists is deliberately restricted
to pages, and the general-purpose tool refuses deletion on purpose — that refusal is a decision
someone made deliberately, with a written rationale, not an oversight. The underlying machinery
*can* delete a file; it is the safety catch above it that says no.

That leaves three honest options: route around that safety catch just this once for a single
harmless file, build a proper narrow capability for deleting assets, or you delete the one file
by hand. My instinct is the first, recorded openly — but going around a deliberate safety
control is exactly the kind of thing I should not do quietly on your behalf while you are not
looking, so I have written up all three with the evidence and left the choice.

Everything is written down in a fresh handoff so a new chat can pick up both jobs cold. One
caution recorded there: another session is currently mid-build inside the same part of the code,
so whoever continues needs to stay out of those files until they are finished.

**2026-08-17 — the recorder has paid for itself, and it found somebody else's bug.** A day's
worth of the new record is in: about 1,570 cases where the "guess between look-alike fields"
search met disagreeing candidates. The big surprise is that **86% of them are one already-known
bug** (287, in the dispatch loop) rather than dozens of pipelines quietly living on luck — which
is much better news than it sounds, because it means the plan's next step is blocked by one fix,
not by a fleet-wide clean-up. We handed that lane our measurements rather than starting a
competing investigation; our numbers answered the one question their own diagnosis had left
open, so their fix is now better evidenced than it was.

We also found *why* the wrong value wins, and it is tidier than expected. When a loop runs, each
pass leaves a copy of every field behind under a numbered name, and the original un-numbered name
holds the *previous* pass's value. The search sorts its candidates alphabetically, and the plain
name always sorts before its numbered copies — so it hands back last pass's data, every time, by
construction. Making the search predictable (which we did last week) did not cause that; it
revealed it, turning "randomly wrong" into "reliably wrong", which is the only kind you can catch.

Nothing needs a decision from you today. Two things are simply waiting: the image-pipeline strict
switch you applied has still not had an occasion to prove itself (that pipeline has run three
times in nine days, and not once since), and the "stop guessing entirely" step waits on the other
lane's fix landing first.

---

**17 August 2026 — you asked me to look once more for a supported way to delete that file, and to build one if there wasn't. There wasn't. It's built.**

The fourth look turned up two things worth having even though they didn't change the answer.
First, there IS a "delete asset" button in the admin interface — and it only edits the
database record, touches no file, and reports "deleted: true" while the file goes on being
served. Worth knowing before anyone trusts it. Second, the one comparable removal anyone has
done here (a stray design file another lane removed last week at your direction) was done by
hand, outside the platform. And the clincher: our stray file has no database record at all,
so every mechanism that works from records is structurally unable to reach it. No supported
path exists. So, per your instruction, I built the narrow capability.

It is deliberately the most cautious thing I have shipped here. It can only touch files
under the assets folder — pages, feeds and site chrome are structurally out of reach
whatever a caller types. It refuses to delete anything the platform still believes is a
real asset, anything the favicon machinery owns, and anything a live page actually links to —
that last check is the same one that twice saved us from overwriting working logos during
last week's clean-up, now built in rather than remembered. And by default it deletes
nothing: it audits, and tells you what it WOULD do and what it refused. Making it actually
delete requires an explicit, deliberate setting. I broke each of those five safeguards on
purpose in a scratch copy and watched a test fail for each one — all five alarms work.

Governance done properly: this does NOT reopen the door the reviewers deliberately shut on
general-purpose file deletion — it is the narrow, guarded shape their own review said such
things should take, and I cited that decision rather than working around it. Submitted for
council review, entered in the register, all committed.

One honest limitation: like all our code changes it does nothing until you next rebuild and
roll the fleet. Everything for afterwards is staged and written down — the check that the
new code is really running, the one settings file to apply, and a ready-made script that
does the dry run first, then the real deletion only when told, then proves the right file
died and the right logo survived. Nothing will fire on its own.

**2026-08-17 (evening) — the other lane's fix landed, and our recorder is what marked its
homework.** The dispatch-loop bug we handed measurements to yesterday was fixed and shipped
today. Our record shows it worked: the specific failure went from about 800 occurrences to
**zero**, across eleven runs of real traffic — and we checked the traffic was still flowing, so
the zero means "fixed", not "quiet". Two details only our record could show: the number of
look-alike candidates the search had to choose between collapsed from a peak of 190 to 22, and
the overall rate fell by nearly three-quarters. Their fix also used our new "this exact source or
fail loudly" marker, in the safe order we warned was essential — put it on before the underlying
fix and it would have broken every dispatch in the fleet.

What is left is now a short, named list rather than a fog: about six places where the search
still guesses, most of them only a few times an hour. Working through that list is one session's
work, and it is the last thing standing between us and the final step (stopping the guessing
altogether). Nothing needs a decision from you.

**2026-08-18 — the strict switch is proven, and one of my numbers was wrong.** The image
pipeline finally had work to do overnight — twenty-six runs — and the switch you applied on
Sunday is proven: every one of them passed the right value through, and none tripped the failure
we were watching for. That closes the last open item on that thread.

Two honest corrections. First, I told you the other lane's fix was a seventy-three per cent
improvement; that was measured over eighty minutes and eleven runs. With a full night's data it
is fifty-three per cent — still large, still real, but I published a number my sample could not
support, and it had been sitting in their bug file for a day. It is corrected in both places and
logged. Second, while proving the image switch I noticed that fourteen of those twenty-six
deployments failed afterwards, on an error about fetching a repository branch — nothing to do
with our change, but a fifty-four per cent failure rate on asset deployments that nobody seems to
be looking at. I have written it where it will be found rather than quietly leaving it in a
success report.

There is a new milestone summary in the lane folder
(`SUMMARY_2026-08-18_the_instrument_caught_its_first_real_bug.md`) — the first one for this
thread of work, written to be read aloud if you want to describe it to someone.

**2026-08-18, later — I picked up the handoff and the next job turned out to be mostly wrong,
which is good news caught at the cheapest possible moment.**

Some background first, in plain terms, because the decisions below only make sense with it. When
one part of the system hands work to another, it has to find the values to pass along — a page,
an item number, a domain. The tidy way is that the instruction says exactly where to look. The
untidy way, which this platform has always had as a fallback, is a search: rummage through
everything collected so far for anything with the right name and take what you find. That search
is the thing we have spent the last fortnight fencing in, because when it finds two different
answers it used to pick one at random, and twice that put the wrong page's content on a page.

Since Sunday we have had a recorder running that writes down every time the search finds two
answers that disagree. It has written about seven and a half thousand of those notes in two days.
The handoff I picked up said we now knew why, and specified the fix: the system was searching even
for values it had already been told exactly where to find, and then throwing the search's answer
away. Stop searching for those, and the noise mostly goes.

**The reasoning was sound and the evidence pointer underneath it was wrong.** The handoff cited a
particular line of code as proof that skipping the search was safe. I opened it before building
anything. It is a different function from the one described — the real one is three hundred lines
further down, and it does more than claimed: it searches for three extra values *whether anybody
asked for them or not*, and one of those, the current page, is **seventy-two per cent of all the
notes we have collected**. So the specified fix cannot touch the majority of the problem, because
that search never consults the list the fix would prune.

Worse, it flips the reassuring half of the conclusion. The handoff argued the recorder was
over-counting: if the answer gets thrown away, a bad search result harms nothing. That is true for
the smaller class. For the big class it is the opposite — nothing else supplies the current page,
so the guessed answer is not thrown away, it is used. **Those notes are not noise. They are the
system guessing which page it is working on, thousands of times a day.**

Nothing has been built and nothing has been shipped. The correction is written up where the wrong
claim lives, and logged in the fleet-wide wrong-calls file. What I need from you is below, in the
chat — four choices, and one of them (whether those disagreements are actually different pages, or
just the same page in two shapes) decides whether this is a tidying job or a live bug.

One more thing, and it corrects my own note from this morning. I told you fourteen of twenty-six
asset deployments failed and called it a fifty-four per cent failure rate nobody was watching.
That was a window sampled inside an outage. Read across the whole fleet it is a two-hour-forty-
three-minute event yesterday afternoon, a repository lookup failing, about eight hundred and
fifteen failed steps across ten different parts of the system and nine sites — and **nothing since
four o'clock yesterday**. So it is not a standing rate and not an asset-deployment problem. But it
did leave **a hundred pieces of work marked failed, eighty-one of them page re-renders, and they
are still sitting there twenty-one hours later** with nothing having picked them up. That is the
part worth a decision, not the outage itself. I made the same mistake this morning that I spent
the morning correcting — a number taken from too narrow a window and reported as a rate.

**2026-08-18, later still — I ran the check, and the answer is "not broken today, and held up by
something nobody wrote down".**

The question was whether those thousands of disagreements are the system confusing two different
pages, or just the same page written down two different ways. It turned out I could not answer it
from the recorder at all: when it notes a disagreement it writes down *where* it looked, but never
*what it found*. So I went to the stored run data and resolved every one of those locations by
hand, across a hundred and thirty-nine runs of each kind.

There are four separate things happening, not one, and lumping them together is what made the
earlier accounts confusing.

**Ninety-one per cent of it is the system searching for something nobody asked for, finding the
wrong answer, and then not using it.** The dispatch loop — the part that hands out jobs — never
asks for a page at all. A safety net deep in the code tops up a handful of values on every single
lookup, whether the caller wanted them or not, and it is that net doing the searching. It finds
leftovers from earlier rounds of the same loop, so its answer really is wrong; it just goes into a
slot that the job-handout code never reads. Pure waste, no harm.

**The other nine per cent is read, and there the answer is currently right.** For the page writer
it is genuinely the same page in two shapes — full record versus just its name — a hundred and
thirty-nine out of a hundred and thirty-nine. For the page builder it is sharper than that:
**thirteen runs out of a hundred and thirty-nine had a genuinely different page among the
candidates** — `disclaimer` sitting next to `contact-index`, `index` next to
`fuel-pricing-framework`. In every one of them the system picked the right page.

But it picks the right one for a reason nobody chose deliberately: when two candidates are equally
close, it keeps whichever it happened to add to the list first. That is a real rule in the code, it
is stable, and it is currently correct — and nothing declares it, no test protects it, and the next
person to reorder that code would have no idea they had broken anything. **So: not a bug, but we
are one innocent-looking edit away from one, in the one place where the value is actually used.**

I also chased a scare and it came to nothing, which I am recording because it changes the plan if
it ever changes. One step in that lookup picks an item out of an unordered collection — the exact
coin-toss we spent a fortnight removing, still present one level down. I measured whether it can
fire in the runs we have: zero out of a hundred and thirty-nine. So it is not a defect and I have
not filed one. It is a place where the randomness could come back.

**The consequence for the plan is the awkward part.** The next planned step ("when the system finds
two answers that disagree, refuse to answer") is harmless for the ninety-one per cent nobody reads,
and harmful for the nine per cent that is read — because in those runs the candidates disagree
*every single time*, so refusing means the page writer and page builder get no page at all, in one
hundred per cent of their runs. **As it stands, that step would only bite the parts that currently
work.** It needs reordering, not cancelling, and I have set out how in the chat.

**2026-08-18, evening — your five decisions are done or answered, and the one still open is
re-presented in chat with better numbers than this morning's.**

The bug you asked to be taken out of the discussion paper and filed properly is filed — two of
them, in fact. The first records that the system currently picks the right page by an accident
of code ordering nobody wrote down, with the fixes ranked. The second records yesterday's
outage properly: the retry machinery we already have did exactly what it was built to do, three
times per item — all three tries inside the same three-hour outage, because nothing tells a
retry to wait. Eighty-eight of the hundred dead items died that way. Your ruling — a blip
should send work back to the queue, not kill it — is written at the top of that file, along
with the one trap for whoever implements it: a repository that has genuinely been deleted
produces exactly the same error as the blip, so "retry forever on this error" would grind on a
dead repo for ever. The hundred stranded items themselves are marked inactive, as you asked,
each with a note saying why.

The efficiency fix you approved is built, tested and submitted for review. I proved it the
paranoid way: switched it off again and watched the right tests fail, then switched it back on
and watched everything pass. It should remove roughly a quarter of the recorder's noise once
it ships in the next build.

The question I owed you an answer on — who is actually living on the safety net — came back
much cleaner than I feared. The three page-related values it injects have almost no secret
consumers: one HTML-generating agent quietly relies on it, and the failure would be cosmetic
(a missing "Page:" label in an internal prompt), fixable with a one-line settings change
before we touch anything. The three business values — domain, objective, model — are the
opposite: **fifty-five places across thirty-one different agents use them without declaring
them**. Those must never be gated; the split between the two groups is the whole answer.

One honest downgrade of my own morning claim: I said the page builder "reads" the guessed page
value. Having now enumerated every reader I can find, I cannot point at the line that reads
it — the population where a wrong guess would genuinely land somewhere is smaller and more
specific than I said, and the chat message names it precisely. The remaining decision — when
and in what form to switch on "refuse rather than guess" — is re-presented there with a
recommended order.

**2026-08-18, late — you chose the sequenced path, and it is now written down as the plan of
record.** In order: the efficiency fix ships and we re-read the recorder; the two small
safety fixes from the new bug file (write down the tie-break rule the code already follows,
and remove the last remaining lucky-dip); then switch off the safety net for the three page
values only — never the business values — with the one affected agent given its one-line
settings fix first; then repair the handful of places that genuinely store the same page in
two shapes, at the source; and only then turn on "refuse rather than guess", which by that
point should have almost nothing left to refuse — which is exactly the state we want it
armed in, as a permanent guarantee rather than a cleanup. Each step waits on evidence from
the recorder, not on a calendar. The full path is in the handoff as a table a fresh session
can execute.

**2026-08-19 — three of the five steps on your path moved today, and one thing I found along
the way is worth your attention.**

First, the efficiency fix you approved yesterday shipped in last night's build, and sixteen hours
of data says it did exactly what I claimed: the class of noise it targets has gone to zero, and
the class it deliberately cannot touch is still there, waiting for the next step. That step is
done.

Second, the two small safety fixes — writing down the tie-break rule the code already follows,
and removing the last remaining lucky-dip — are built, tested, and approved by the reviewers
without a single objection. They change nothing visible today; they exist so that nobody can
break the current behaviour by accident. They ship with the next build.

Third, the bigger one: switching off the safety net for the three page values. The settings fix
for the one affected agent is already applied and live. While doing it I found that agent has
**never run** — not once, ever — and had no settings at all for what it reads. That is exactly
the kind of thing that breaks a year from now when somebody wakes it up, long after everyone has
forgotten why; giving it an explicit list was cheap and closes that door for good. The code
change itself is built and sitting with the reviewers now. One more honest note on it: when I
applied the change, **every existing test still passed** — nothing had ever checked that the
safety net did what it did. That is the same pattern as the tie-break: behaviour nobody wrote
down, doing real work. The new tests check both that the net is off for those three values and,
more importantly, that it stays firmly on for the three business values that fifty-five places
depend on.

Nothing has shipped beyond the first fix; the other two ride the next build. What remains after
that is the handful of places that genuinely store the same page in two shapes — the last step
before "refuse rather than guess" gets switched on.

**2026-08-19, later — the safety-net change is approved, after a review round that earned its
keep.**

The reviewers sent it back once, and they were right to. The whole safety case rested on my
census of who consumes the injected values — and my census had only looked at the top layer of
each agent's workflow, missing everything nested inside loops. Embarrassingly, one of the steps
it missed was the very step I had traced by hand two hours earlier; it should have collided
with my own count and I did not notice the absence. I redid the census properly — every level,
every agent — and the conclusion survived: still nobody consumes the three page values without
declaring them, and the do-not-touch list for the three business values actually grew by one.
The lesson is logged in the wrong-calls file with a note that this exact trap was already
written down three times in our landmines file and I had not looked. On their smaller points:
the "never ran" claim about the dormant agent had been read from a table that only keeps five
weeks of history — re-checked against records going back to March and a counter that has read
zero since December, so the fact stands on proper ground now. And a reviewer's fair complaint
that a missing value would vanish silently became a small addition: the system now says, on a
log line it already writes, exactly which values it declined to inject and why.

Second round: approved, no blocking objections. So all three code changes on your path are now
built, tested, and reviewed — the first is already live and proven, the other two wait only on
the next build you roll. After that roll, the recorder should go quiet on the big noisy class,
what remains is the one name-collision repair, and then the final switch.

**2026-08-19, late afternoon — the name-collision repair is built and in review. It turned out to
be smaller and in a different place than I had said this morning.**

The build you rolled at lunchtime carried all three of the earlier changes, and the recorder did
what we predicted: the big noisy class went to zero, the retry-echo class went to zero, and
exactly one thing was left — the same row, over and over, twenty-three times in four hours. It
is the collision I described this morning: when the writer asks for "the current page", the
system finds the page itself in one place and, one level down, a plain text label with the same
name — the page's name string that the layout step writes for the navigation. Two different
things, one name, in one tree, so the recorder flags it every run even though it always picks
the right one.

This morning's plan was to rename the label at its definition. Before building it I measured
who actually reads that label, and the measurement changed the plan. The label's name is also
what every component template reads to know which page it is on — and while only one template
in three hundred actually uses it today, renaming it there means editing live templates in step
with a build, and it touches a place that has no problem. The re-render path also builds the
same label by hand and depends on the same name. So the repair is narrower: the label keeps its
name everywhere templates see it, and only the copy that gets filed into the shared data tree —
where it collides with the page record — is filed under a longer name, "current page name". The
reader side accepts both spellings for a while so anything mid-flight when the build rolls is
unaffected. Templates see exactly what they saw before.

I proved it three ways: tests that fail if the old name is ever filed again (including sneaking
in through a side door), tests that fail if the new name is not read back, and a test that runs
the real lookup over a tree shaped like the live run and checks the recorder stays silent — with
a control that runs the old shape and checks the recorder still fires, so a quiet recorder cannot
be mistaken for a fixed one. It is committed and with the reviewers; it goes live on your next
build. After that, the recorder should be completely quiet on this class, and the last step — the
switch from "pick one and log it" to "refuse to guess" — is the next thing to build.

One note for honesty: this morning's handoff described this fix as a rename of the definition.
That was the right target and the wrong tool; the notes file records the correction and what
caught it (measuring before building).

**2026-08-19, evening — the name-collision repair is approved, after one round that was about a
filing cabinet, not the fix.**

The reviewers sent it back once. Their blocking point was that an old bug about this same page
label — the one where it always came through empty — was still sitting in the "open" folder,
and my write-up had not mentioned it. Fair question: could the two be one problem seen from two
ends? I went and read that file. It had been fixed and proven live on the 27th of July; its own
last section says so in detail. Nobody had moved it to the "closed" folder or updated the line at
the top, so for three weeks its location said one thing and its contents said another. I moved
it, marked the top with the date, and wrote a short closing section explaining why the two
things are different: that bug was about the label's value being empty; this one is about the
label's name clashing with the page record's name. Each has its own proof and neither can
disguise the other. The old bug's own test is still in place and is one of the checks this
repair was proved against.

The other real point was about a note in our landmines file warning anyone touching the
identity code in this area. Read in full, that note is actually the reason my edit exists — it
lists the very key I renamed as one of the three that work — so I engaged with it, re-measured
on every stored run, and corrected the note in place with today's date.

Second round: approved. One reviewer asked for a guard against the rename table quietly growing
— sensible, and it is now a test that fails if anyone adds a second entry without doing it on
purpose. So step four is done and reviewed; it goes live on your next build. Then the last step:
flip the recorder from "log it" to "refuse to guess".

**2026-08-19, evening — the new build is verified, and I discovered a colleague had been busy.**

The build you rolled this afternoon carries the first three steps of the path, and I proved all
three on it: the noisy class is at zero against live demand, the class we expect to survive is
surviving, and — new — the "declined to inject" log message is now witnessed in live production
output, closing the one warning the earlier check had left open.

Then a surprise, of the good kind. While I was verifying, another session working this same
lane had already: built the third fix option on the tie-break bug (the one I had set aside as
"not bundled"), taken it through diagnosis and review properly, closed that bug outright — and
**built step four as well**, with a better design than my sketch: my version would have renamed
a field everywhere; theirs renames it only at the boundary where the collision happens, and
their measuring found a second producer my design had missed. It is approved and waits, with
everything else, on your next build. Where our two accounts disagreed I have reconciled them in
the shared handoff, and one embarrassment is recorded honestly: I nearly closed that bug a
second time with a wrong summary, caught only because the file had already moved.

With the path fully blocked on your next roll, I used the time on the outage bug from
yesterday: the twelve items that died on their first attempt are now explained — five different
agents carry a "mark it failed" step that ignores the three-attempt allowance entirely, so for
them ONE failure is terminal even in fair weather. Combined with the other two defects already
on file, the remedy converges on a single change: one properly-guarded way to write a work
item's final status. The design skeleton is written into the bug file; building it is the next
coherent piece of work, ready for whoever takes it — this session, a fresh one, or another
lane.

**What waits on you: one more build roll.** After it: verify step four the same way, then the
final switch.

---

**2026-08-19, night.** Still waiting on your roll — I checked properly rather than assuming, and
the live build genuinely does not contain step four's code, so there was nothing to verify yet.
Two useful things came out of the waiting.

The first is small but worth having. The test for step four is "after the roll, this warning
stops appearing". A warning that stops is only meaningful if it was happening beforehand, so I
recorded that it was: fifteen occurrences in the three and a half hours before I looked, the
most recent one a minute before I ran the query. Now the silence afterwards will mean something.

The second is bigger, and it changes the last step of the plan. That final step flips the
system from "when the search finds two different answers, pick one" to "when it finds two
different answers, refuse". The code itself says what has to be true before we are allowed to
do that: either the warning has gone quiet everywhere, or every place still producing it has
been given an explicit answer to use instead. I had been assuming we were nearly there. **We are
not.** Nineteen distinct places have produced this warning since the instrument went in;
everything we have built so far accounts for five of them. Fourteen are unaddressed.

What made this easy to get wrong is that most of the fourteen look dead — their last warning was
a day or two ago. But I checked whether the code that produces them had run at all since, and it
had, heavily: one of them ran nearly four hundred times in a day without tripping. So they are
not fixed. They are waiting for the right data to come past. If we flip the switch believing
they are gone, we will find out otherwise in production.

Chasing the largest of those fourteen to the bottom turned up a real bug, which I have filed.
When the system builds a small interactive tool for a site, it is supposed to link to that tool
from whichever pages are genuinely related. When nobody has said which pages those are, the
correct answer is "link from none". Instead the search goes looking for anything anywhere called
"related pages", finds the list belonging to a *different, unrelated* tool, and uses that. On
webdesign.co.uk the result is that nine different tools — an SVG optimiser, a JSON cleaner, a CSS
generator — have all been queued to be linked from the same two statistics pages. The good news
is that none of those links reached the live site; they sat in the queue and failed. Four other
sites do this correctly, which is how I know the substitution is the fault and not the design.

I record one wrong turn. My first explanation was that an old fix had been undone, and I had
started writing that up before I opened the file and found the fix exactly where it should be —
I had looked for it in the wrong place and read "not here" as "not anywhere". That is in the
wrong-calls log.

**What waits on you is unchanged: one more build roll.** What changes is what comes after it.
The final switch now needs a piece of design work first — deciding what each of those fourteen
places should get instead of a guess — rather than being the quick flip we had penciled in.

---

**2026-08-20, morning.** Your build rolled overnight and **step four is done — live, and proven
properly.** The service's own startup line that says which code it is running had already
scrolled away by the time I looked (it lasts hours, not days), so I asked the running binary
directly whether it contains the new code, on both machines, with two control checks either side
so a broken probe couldn't pass for a good answer. It does. And the warning step four was built
to eliminate has gone silent: zero occurrences overnight, where the rate beforehand predicted
about nine. I'll be straight about the limit — only three of the relevant jobs ran in that
window, so I can say the old every-time behaviour is gone, but not that some rare version of it
is. If that distinction matters we can look again in a day.

**That leaves one step, and it is no longer the quick flip we had penciled in.** Two things came
out of the checking.

The first is a correction to our own plan, and I'd rather record it than quietly work around it.
The plan said we could safely delete a bit of backwards-compatibility code because old records
expire after a day. They don't — there are records in that table from mid-July. More to the
point, that compatibility code is also used when re-rendering a page, and page data never expires
at all; twenty live pages across twelve sites still hold data in the old shape. **The conclusion
turned out to be right anyway**, for two better reasons I've now written down and can prove with a
query. But the reason we had was wrong, and if nobody had checked, we'd have deleted the code on
the strength of it.

The second is a real cost, and it's the useful kind — specific and fixable. When the system
finishes a piece of work it records which code change deployed it. Nobody ever configured where
that value comes from, so it is being found by the same "search everywhere and hope" mechanism
this whole workstream exists to remove. It appears to be landing on the right answer today, by
luck rather than design. **The final step would switch that mechanism off, and this value would
quietly stop being recorded** — no error, just an empty field, on something another workstream's
page-publishing fix depends on. So I have written to that workstream, told them what we found and
what we plan, and asked them for the correct answer rather than picking one myself, since the
whole point is to stop guessing. It needs a one-line configuration change, not a code change.

That is the shape of the remaining work: about fourteen places like it, each needing to be told
explicitly where its value comes from before we turn the guessing off. Most will be quicker than
this one. It is a list, not a mystery — which is a much better position than we were in last
night, when it was fourteen unknowns.

**Nothing waits on you right now.** The next chat picks up the list.

> **CORRECTED 2026-08-20, same morning:** I said "about fourteen places" above. The final census
> puts it at **thirteen** — one more (the page-build-handler case) turned out to have been killed
> by an earlier step than I had credited. Caught by re-running the count before writing the
> handoff, which is the only reason the two documents don't now disagree. The precise list, with
> the evidence beside each row, is in
> `docs/agent_docs/docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-20_continue_here.md` §3.

**2026-08-20, morning — the repair is live and proven, and the silence it created let us hear
two quieter problems.**

Your overnight build carries the rename. I verified it the strict way — asked the running
binary what it was built from, with a control that must be absent — and then read the recorder:
the collision that fired twenty-three times in four hours yesterday has fired **zero** times
since the roll. Better than the zero: the three page-writer runs since the roll all file the
label under its new name and none file the old one, so the clash isn't just quiet, it's
impossible.

But the plan's last step — flipping the recorder from "log it" to "refuse to guess" — needs the
recorder to be quiet across the board, and with the loudest class gone two smaller ones are now
audible. One is brand new as of last night and fires on every build the dispatcher completes:
a bookkeeping step wants to note "which commit shipped this work", the reply from the git
service only started carrying that information yesterday (a different team's improvement), and
the dispatcher's loop keeps every previous item's reply lying around — so the search finds five
copies with five different answers. It picks the right one, deterministically, thanks to the
tie-break we shipped last week — but "right by sort order" is exactly the situation we've been
draining. I filed it properly (bug 334, with the diagnosis loop dispatched) with three ranked
fixes; the cheapest is a one-line config change. The other is an older, slower drip from the
tool generator asking for very generic names like "reason" — untraced yet, next session's job.

So: step four done and proven; step five waits on those two; nothing waits on you.

**2026-08-20, correction to my entry above (same morning):** I wrote it without reading what the
overnight session had already put in this file and the notes — that session verified the repair
first, traced the tool-generator drip (it had already produced bug 330), counted the full
remaining population (about nineteen name-clashes, fourteen with no plan, in four families), and
measured the two things I had labelled as guesses: the dispatcher noise started because of a big
repair batch three minutes earlier, not the git service's release, and the "picks the right one"
claim was checked against real runs, not just reasoned. My additions that stand: the bookkeeping
clash is now a proper bug file (334) with ranked fixes — the cheapest is a one-line config
change — and the diagnosis loop is running on it. Lesson recorded in the wrong-calls file: read
the bottom of the shared notes before acting on them, every time you wake, not just at the start.

**2026-08-20, late morning — the "how big is the guessing problem really" question now has a
measured answer, and it is smaller and better-shaped than feared.**

The night notes said about fourteen name-clashes stood between us and the final "refuse to
guess" switch, and that nobody knew how many quiet cases sat underneath them. I measured it.
Across the whole fleet there are 451 places where a step names an exact path for one of its
inputs; on the busiest handlers I then checked, run by run, how often that path comes up empty
and the system's guesswork quietly fills the gap. First pass said ten places; an hour later I
had to correct my own number to FOUR — my probe had been counting text in areas the search
deliberately ignores, the same "your measurement answers what you encoded" trap in our own
lessons file, caught by looking at one real run instead of trusting the count.

Of the four: two are the page builder reading a page's identity from a record that half the
time was never loaded — and there the guesswork always lands on the right answer (I checked
forty runs; the loaded record and the request agree every single time either exists). Because
the copies agree, the final switch never even sees them — they don't block anything, they just
need a written "safe, and here's why" note, which the measurement now provides. One is a single
odd run on the re-render path, to be looked at once. And exactly ONE is the wrong-value case —
the cross-links bug from last night, where the fix IS "answer nothing". So the road to the
switch is: the named clashes (each now filed or grouped), one note, and one trace — not
fourteen unknowns.

---

**2026-08-20, evening.** Short answer to "can we close this lane": **not yet, but the end is now in
sight and it is smaller than it looked this morning.**

The build that rolled at 16:09 contains nothing of ours — the last of our four changes went out in
the previous one and the fifth isn't written. I checked rather than assuming, and I'm noting the
non-event so nobody invents a job for it.

The useful work today was turning the remaining step from a count into a shape. Yesterday I told
you there were about thirteen places still relying on the guessing, and that was true but not very
actionable. Today I checked each one against what the receiving code actually *requires*, and they
fall into three groups of very different size:

- **Two are genuinely hard.** The receiving code treats the value as mandatory, so switching the
  guessing off turns today's silent wrong answer into an outright failure. That may well be the
  right outcome — better to stop than to build a tool under another tool's name — but it is a
  decision to take deliberately, not something to find out about afterwards.
- **One is the real gate.** It's the "which code change deployed this" value I wrote to the other
  workstream about yesterday. It's optional to the receiving code, and that is exactly why it's
  dangerous: nothing fails, the value just quietly stops being recorded. It needs a one-line
  configuration change, and I still want their answer for what the correct source is rather than
  choosing one myself.
- **About ten are paperwork.** For these, "no value" is the *correct* answer — that is the whole
  point of the change. So satisfying the rule means writing down, once per case, that nothing is
  the right answer. A paragraph each, not a piece of engineering.

So: one external answer, two decisions, and ten paragraphs — then the switch, then the lane closes.

Two other things worth telling you. Another session was working this same lane this morning and did
a substantial piece of measurement I'd flagged as needed — sizing how widely this guessing actually
rescues missing values across the estate. They found four real cases out of 451 candidates, and
notably they caught and corrected their own first answer of ten within the hour. That materially
shrinks the risk of the final switch, and I've credited it in our notes.

And a mistake of mine, which cost nothing but could have. I wrote a new "start here" document this
morning and marked the old one obsolete — while that other session was still updating the old one.
For about nine hours we had two documents each claiming to be the single entry point, and the one I
called obsolete had the newer content. I've merged them into one and logged the lesson: declaring
another document dead is a claim about it being dead, and that's the one thing a single session
can't see on its own.

**Nothing waits on you.** The next session has one external answer to chase and a list to work
down.

Picked this up cold this morning (2026-08-21), about seventeen hours after the entry above, on a
tree where roughly 270 more commits from other work had landed overnight. Worth recording plainly
what changed and what didn't.

The two "genuinely hard" cases from yesterday turned out not to be hard at all — they were old.
Every single row of both classes happened before the very first fix in this whole piece of work
shipped, and that fix's own code comment explains why: back then, a value that was already correctly
supplied still got run past the guessing machinery out of habit, the guess was logged, and then
thrown away in favour of the correct value nobody ever saw the guess. Once I lined the timestamps up
against the fix's rollout, both "hard" cases turned out to be noise from before the door was shut,
not defects still open. So the list is smaller than yesterday's count said, not because anything got
fixed today but because two entries on it were never really there.

One of the "about ten paperwork" cases is now actually done. It was the news-section note being
mistaken for a tool page's reason for rerendering — harmless today only by luck, because the only
values that would have mattered happen not to be the ones on that page right now. I wrote down what
the page actually needs to be told, and that went through review and applied cleanly. I haven't been
able to prove it worked yet, for the most boring reason possible: the specific kind of work that
would trip the old bug hasn't happened again since I made the change. That's not a failure, it's just
waiting on the right kind of traffic to show up, and I've said so plainly rather than claiming a win
I can't back up yet.

A second case turned out NOT to be paperwork after all, and this is the one worth your attention.
Checking each remaining case against what actually happens if the guess goes missing, I found one
where the guessed value is currently right, and losing it would quietly make things worse rather
than nothing at all — the code falls back to using the page's web address in a field that's supposed
to hold the page's category, which would throw off how a page-building tool picks its components.
So instead of writing "nothing is correct here," which is what I did for the first case, this one
needed the opposite treatment: telling the system explicitly where the right value already lives, so
it stops guessing and just uses it. That's written, matches how the neighbouring setting on the same
page is already written, and is sitting in review now.

The one external answer is still the one thing genuinely waiting on someone else, and it's grown
since yesterday — the class is now roughly three times the size it was when I first flagged it, still
apparently landing on the right value by luck rather than by design. I sent a plain status update to
that lane today rather than pushing further; picking that value ourselves is exactly the kind of
guess this whole piece of work exists to stop, so it stays theirs to call.

**Nothing waits on you either.** What's left: one more of the "paperwork" cases needing a decision,
about seven more that look straightforward, the one external answer, and then the switch.

---

**2026-08-21, midday.** Good progress, and one thing is genuinely stuck for a reason worth knowing.

Another session worked this lane last evening and built the key enabler: a way for a step to say
"take this value from exactly here, or accept that it's missing — but never go hunting". That is the
tool the whole remaining job needs, and **I confirmed this morning that it is live in the running
system.** Confirming it was harder than it should have been, and the story is worth two sentences
because it changed a rule. Our normal way to prove a change is running is to ask the program
whether it contains a distinctive phrase from that change. This change has no such phrase — the
only quotable lines in it are comments, which are thrown away when the program is built. So the
check came back "not present" on a system that definitely had it, with all the safety checks
behaving perfectly. I only caught it because I checked whether the phrase I was searching for was
real code before believing the answer. I've written the lesson down along with a method that does
work, and a note to future authors: if your change will ever need proving, put one provable marker
in it deliberately.

They also solved a puzzle I'd left open. I'd flagged one case where the evidence contradicted
itself and said it needed investigating. The answer turned out to be a date: every one of those
records predates a fix we shipped days ago, and before that fix the system recorded a complaint
even when it wasn't acting on it. So those records were never a problem — and the measurement I'd
taken the day before had already predicted that, if I'd joined the two facts up. Two of the three
remaining hard cases evaporated on that finding.

**I've built the next fix and it is with the reviewers now.** It concerns a step that plans which
components go on a page and needs to know what kind of page it is. Nobody ever told it where to
look, so it searches — and among the things it finds are the page-types of *every other page on the
site*. Measuring it properly turned up something worse than the records showed: on **eighteen of
thirty-one** runs the page's own details aren't in scope at all, so the only candidates are other
pages — and when those other pages happen to agree with each other, the substitution leaves no
trace at all. So this was never just tidying for the final switch; it's a live wrong-answer path,
on most runs, and the fix improves things today.

**The stuck thing.** Yesterday's fix to a related case was applied but deliberately left unverified,
on the reasonable expectation that a few hours of normal traffic would prove it. I checked
seventeen hours later: the component in question has not run once. Not because it's broken —
because there is no work left for it. All forty-four queued tool jobs are finished. So that
verification has no natural path to completion, and waiting cannot fix it. I could force it by
commissioning a real tool build on a real site, but that's a decision with real output and not one
I'll make unasked. I've written it up as "applied and explained, but not demonstrated", and flagged
that nobody should let "applied, and nothing has gone wrong since" quietly become "verified" — it
isn't the same claim, and here the difference is the whole of the evidence.

**Where that leaves the lane:** one external answer still outstanding, one fix in review, one
verification blocked on there being no work to observe, and a set of written-decision items. Then
the switch. **Nothing waits on you** unless you'd like me to commission that tool build.

---

**2026-08-21, afternoon.** The fix I described this morning is now **finished and proven working in
the live system** — and proving it turned out to be the interesting part.

The change itself went through review, came back needing revisions, and the reviewer was right in a
way worth telling you. My evidence that the fix would work had two halves: I'd read the relevant
code, and separately I'd proved the running system contained the update that added it. Those sound
like one argument. They aren't — I read the code as it exists *today*, while the running system was
built from an *earlier* point, so the part I read could have arrived after the system was built. It
happened to be fine, but I'd claimed it verified on a check I hadn't actually done. I've written
that up, because it's the sort of gap that reads as thorough right up until it isn't.

The reviewer also found a third place with the same problem that I'd said had only two — it was
hidden inside a nested structure that the obvious way of counting can't see. That one didn't change
the fix, and for a good reason: the other two places don't have the piece of data this fix points
at, so pointing them at it would have been inventing something. They're written down as separate
jobs needing their own measurement.

Then the fix was approved, applied, and I set about proving it actually does something. **Two
reviewers had independently pointed out that everything so far was an argument from reading code,
and nothing had ever actually run through the new path.** They were right, so I built a test that
watches the system make the decision and reports which way it went.

**It took three attempts, and the first two failed silently** — which is the dangerous kind. The
watcher reported nothing for an hour while the work was happily running, because I was watching the
wrong machines: this particular job doesn't run where I assumed, it gets a brand-new temporary
machine for every single run. One database column would have told me that, and I hadn't looked.

The third attempt caught it, and the result is clean:

> the system asked for `section_facts`, `pipeline` and `site_type` — and **not** `page_type`

which is exactly right. `page_type` is now taken from the page's own record instead of being
guessed. And there's a control sitting in the same line: `site_type` is in the same category, also
unconfigured, and is *still* being guessed — so the change did what it was aimed at and nothing
else. That's the difference between "the warnings stopped" and "I watched it work".

This is also the **first live use of the new mechanism** the other session built yesterday, so this
one line is the proof that mechanism functions in production — which unblocks the other fixes
queued behind it.

**Still waiting on real work to appear:** the tool-related verification from yesterday, which needs
someone to queue a tool build. Nothing waits on you unless you want that forced.

---

**2026-08-21, late afternoon.** **The last blocker is fixed and proven, so the final step is now
unblocked.** Two things worth telling you, one good and one awkward.

The awkward one first: **another session built the exact same fix I did, and applied theirs about an
hour before mine finished review.** That's the second time in two days two of us have independently
built the same migration. Theirs ran, mine is retired. Theirs is genuinely better in one way — it
finds the piece of configuration it needs to edit rather than assuming where it lives — so I've
retired mine and moved the three pieces of analysis mine had that theirs didn't into their file,
rather than leaving them to rot in a dead document. I also found their migration had been applied
but never written into the ledger that stops it being run again, which would have jammed the next
person's batch, so I recorded it and told them.

The good one: **the fix is proven working, and the proof came out unusually clean.** I predicted, in
writing and before looking, that any job started after the change would behave differently from any
job started before it. Then I checked every relevant run and joined each one to its own start time:

> two jobs started **before** the change — both still doing the old thing
> two jobs started **after** it — both doing the new thing

Four out of four, no exceptions. The failures are what make it convincing: a test where everything
passes proves nothing, and these didn't.

**And it nearly went wrong twice, in ways worth recording.** My first check looked for a step by its
name and found nothing at all — which reads exactly like "this never happened". The step is named
differently at runtime than in the configuration: the system adds the loop iteration to the front of
it. That's the third time in two days one of my checks has been quietly asking the wrong question
and returning a confident empty answer, so I've written the general rule down: **test your check
against a line you have actually seen, not against a name you believe in.**

The second near-miss: one of the jobs started *before* the change was still doing the old thing
**eight and a half minutes after** it went live — because a job in flight keeps the settings it
started with. If I'd judged by the clock rather than by which job was which, I'd have reported the
fix as broken. It happened to be recoverable because the logs carry the job's identity; the other
half of our instrumentation doesn't, which is now written down as a reason to prefer this kind of
check.

**Where that leaves us: the final switch is the only work left.** Everything it was waiting on is
now either fixed, proven, or written down as a deliberate decision.

---

**2026-08-21, evening. The final step is built and approved.** After a month, the change the whole
workstream was for is written, tested, reviewed and committed: when the system searches for a piece
of information and finds two different answers, it now says *nothing* instead of picking one.

It took three rounds of review, and I want to record what those rounds were actually worth, because
"approved on the third attempt" could read as three tries at persuasion. It wasn't.

**Round one found a real hole in my reasoning.** I had proved that every case the instrument had
ever *seen* was handled — nineteen of them, each fixed or given an explicit answer. A reviewer asked
how many cases there *could* be. I had written the phrase "every step that falls through to this
search" myself, in my own justification, and never turned it into a number. It's 137 steps across 71
agents. The argument survived — the change only affects cases where the two answers actually
disagree — but I couldn't have said that before, because I didn't know the denominator. **Proving
every observed case is handled is not the same claim as bounding the population**, and the two read
almost identically in prose.

**Round two's blocking objection was wrong, and proving that was the useful part.** The reviewer said
one of our old bug records was still open, which would have made this change repeat a documented
failure. It isn't open — it was closed two days ago, *by this same workstream, answering the same
reviewer*. Their index is reading an older snapshot. I proved it from the repository rather than
asserting it, and I've written a note for whoever submits next: if you cite a recently-closed bug,
bring the proof, because saying "it's closed" won't clear it.

**Round three approved it, with three advisories I've recorded as owned rather than ticked off.** The
sharpest is worth repeating to you plainly: this change swaps a *silently wrong* value for a
*silently absent* one — and downstream, an absent value still renders as a blank with no error, at
fourteen of fifteen places. **That is not fixed by this, and I have not claimed it is.** It's a
separate filed bug. What this change does add is that every refusal is *recorded* the moment it
happens, naming the field and the caller — which is precisely what the blanking problem lacks, and
why it has been so hard to find.

**It doesn't take effect until the next build rolls** — it's program code, not configuration. And I
have committed this workstream to actually watching it afterwards rather than declaring victory: a
48-hour check with the terms written down in advance, including the trap that a job already running
keeps its old behaviour for several minutes, so the clock is the wrong way to judge it.

A second session has joined this lane and is taking the small follow-on cleanup. Nothing waits on you.

---

**2026-08-22, morning. It's live.** The change this workstream existed for went out in this
morning's build and is confirmed running: when the system searches for a value and finds two
different answers, it now returns nothing instead of picking one. The companion cleanup shipped with
it. **All five steps are built, shipped and verified.**

Confirming it needed three different checks, because the two changes are different shapes. One adds
a new marker to the program, so I could ask the running system whether it contains it — and did,
with a control alongside so a broken check couldn't pass for a good answer. That mattered: two of
the machines answered "not present" while the *control also failed*, which means the check itself
wasn't working there, not that the code was missing. The other change only *deletes* code, so there
is nothing to ask for; for that one I read the source at the exact commit the running build was made
from, which showed the deleted branch genuinely gone.

**A caution I want on the record, because it would be easy to claim more than we have.** I have set a
48-hour watch running. **It cannot prove the change works, and I won't present it as though it can.**
The reason is that we fixed all the live cases *before* switching the guessing off — so the warning
we'd be watching for had already gone quiet yesterday, for a different reason. Silence over the next
two days is what we'd see either way. What the watch is genuinely for is two things: catching a
machine still running the old code, and catching a *new* case we've never seen before — and for that
second one the system now records every refusal the moment it happens, naming what was asked for and
who asked. That is the part that makes the residual risk observable rather than invisible.

What actually carries the claim that the change works is the thirteen tests that fail if you undo
the one line, plus the confirmation it's in the running program. The watch adds coverage, not proof.

**Can we close this lane?** Nearly, and I'd rather be precise than tidy. Everything the workstream
set out to do is done. Three things remain, and none of them is more of this work:

- the 48-hour watch, which closes Sunday morning;
- one small verification that has no path until someone queues a tool build;
- and three follow-ons that belong to other lanes or to a later decision, all named and none started.

Nothing waits on you.

---

**2026-08-22, late morning — the last piece of building is done, and one of the open bugs is now
properly closed.**

Two things happened since the note above, and neither of them changes the plan; they finish it.

**First, the commit-tracking bug is closed.** For background: when a build finishes a piece of work,
it records which code change the work produced. That value used to be found by searching the whole
job for anything that looked like one, and in a job that loops — a build that handles ten pages one
after another — it would sometimes pick up the wrong loop's answer and attach it to the wrong piece
of work. We fixed that in two halves yesterday: ten of the workers now state their answer in one
agreed place, and the loop now reads that place and nowhere else. What was missing was the closing
check, so I did it before moving the file: over the last eighteen hours the system has run
ninety-three of these loops and has not recorded a single one of the old confusions, while a hundred
and fifty-four completed pieces of work still carry their commit. That second number is the one that
matters, and it is why I am willing to call it closed — if the fix had gone too far and simply
dropped the value, both numbers would be zero, and they are not.

**Second, I built the standing guard we said we owed.** The fix above was protected by a check that
ran once, at the moment we applied it, proving every worker that could produce a commit was
announcing it properly. But new workers get added over time, and one that arrives tomorrow without
that announcement would silently stop recording — no error, nothing in a log, nothing to notice,
because "nothing here" is a legitimate answer in the new design. So that one-off check now runs
every morning at 6:45 and writes down what it found, clean or not. I proved it in the live system
rather than trusting the deployment: it ran, it examined all 194 workers, it found the same eight
that can produce commits and confirmed all eight announce properly, and it wrote its row. I also
proved it can *fail* — fed a made-up worker that doesn't announce, it says so and exits red. A check
that has never been seen to fail is not yet a check.

**A caution on the 48-hour watch, unchanged and worth repeating.** Its first reading this morning is
clean, over real traffic — a hundred and eighty-nine jobs across twenty-six different worker types
since the watch opened, and none of them hit the situation we're watching for. That is the expected
result and it still is not proof, for the reason in the note above: we fixed the live cases before
switching the guessing off, so silence is what we would see either way. It closes Sunday morning.

**One thing I noticed and deliberately did not act on.** Tool builds have been refused more often
today than yesterday — the system's own quality gate turning away generated code that would clash
with the rest of a page. It looks like a jump, but some of those refusals are our own test build
being turned away, and the records only go back two days, so there is no honest baseline to compare
against. I have written down the numbers and said plainly that they are not yet a rate. Someone
should look again in a few days when our own interference has aged out. Nothing is broken by it —
the gate is doing its job; the question is only whether the code being generated has got worse.

**Where that leaves us: nothing waits on you.** The watch closes Sunday, one small verification
still needs somebody to queue a tool build, and the remaining follow-ons belong to other lanes.

## Sunday 24 August, later — the fix quietly went live, and the reviewers caught me being sloppy about my own paperwork

Three things happened this afternoon, and the middle one is the interesting one.

**First: the fix is live.** The change we made to stop new tools losing their cross-links had been
sitting finished but not shipped — it only takes effect when the system's software is rebuilt and
restarted, and that hadn't happened yet when the last note was written. It has now. The restart
went out at about half past nine this morning.

Worth saying how I checked, because the obvious way would have failed. Each service announces
which version of the code it is running when it starts up, but that announcement scrolls away
within hours on a busy service, and it had already gone. Rather than guess, I asked the running
program directly whether it contains the new behaviour — and, importantly, asked it two control
questions at the same time: one thing that must be there, and one thing that cannot possibly be
there. Both answered correctly, so the answer to the real question can be trusted. A check that
would say "yes" no matter what is not a check.

**Second, and this is the one to hold on to: live is not the same as proven, and I nearly wrote it
up as if it were.** Since the restart there have been no failures of the kind we fixed. Yesterday
there were five. It is very tempting to call that a result. It isn't one — and the reason is worth
a sentence. Three new tools have been built since the restart, and all three had no related pages
listed at all, so they never got as far as the part we changed. The new code has not actually been
put to the test yet; it has simply not had the opportunity to fail. So the honest position is
"live, and waiting for its first real case", and I have written it that way everywhere rather than
banking the zero.

I mention this because it is the same shape as the mistake I made yesterday with the page
addresses — a number that looks like good news, believed too quickly. The difference is that this
time the question "could this result have come out any other way?" got asked before it was written
down rather than after.

**Third: the review panel sent the fix back a second time, and it was right to — but the fault was
in my description, not the code.** When you submit a change for review you attach a sketch of what
you changed. The panel spotted that my sketch still showed the old, faulty version of one line —
the very line the previous round had told me to fix. I *had* fixed it in the actual code; I just
never updated the description of it. From the reviewer's side those two situations look identical,
so objecting was the right call, and they even said explicitly which check would tell the two
apart. That cost a round, and I have written it into the shared log of mistakes with the fix, which
is nearly free: copy the sketch out of the actual file rather than from memory of what you meant
to do.

**The genuinely useful thing the review found** was a second point I had no good answer to. We had
proved the new decision-making function gives the right answer for every combination of inputs —
but nothing at all proved that the real code hands it the right inputs in the first place. That is
precisely how this bug survived nineteen days in the first place: the pieces were all tested, the
join between them was not. So rather than argue the point I wrote the missing test, and then
deliberately broke the code to check the test actually notices — it does, and the two older tests
carry on passing while it fails, which is exactly the reviewer's point made concrete.

That has gone back for a third round of review. Nothing here needs a decision from you.

---

**2026-08-24, evening — the "worth a look" item from this morning turned out to be our own fix
working, and underneath it was something we had never looked at.**

This morning's note flagged something odd: a particular "nothing to link here" message had been
recorded twelve times ever, and four of those were today. A third of all occurrences on one day
looks like something starting to go wrong.

It isn't. It is a change we made ourselves on Friday evening, doing exactly what we designed it to
do.

The background: when a new tool is built for a site, we try to add a mention of it on a couple of
related pages, so people can find it. Which pages those should be is meant to come from the request
that asked for the tool. Until Friday, if the request didn't say, the system quietly went looking
elsewhere and grabbed *another* tool's list — which is why nine tools on webdesign.co.uk all ended
up pointing at the same two pages. Friday's change stopped it guessing. Now, if the request doesn't
name any pages, we add no mentions and record why.

So the count going up is the cover coming off, not a new fault. We swapped wrong links for no
links, which was the whole point.

**What I did not expect was the next question: why does the request so often not say?**

The answer is clean enough to be worth stating plainly. There are two ways a tool gets requested.
One is the automatic route, where the system looks at a site and proposes tools — and that route
fills in the related pages **every single time**: eleven for eleven. The other is by hand, which is
how we and the owner have been ordering tools all week — and that one has **never once** included
them. Fifty-eight requests, none of them.

It isn't a case of the automatic route being unreliable. The two routes write genuinely different
requests, and the hand-written one grew from a template that simply never had that field in it.
Nobody was told, because nothing complains: the tool builds, the page goes live, and the only trace
of the missing mentions is an informational line that says "no related pages were named" — which
reads like a decision somebody made, rather than a question nobody was asked.

The practical effect is that for the last three days, every tool we have built has got no
cross-mentions at all. Thirteen out of thirteen.

**One consequence for the bug we closed this morning.** I had written that the last remaining check
on it was "a wait" — we just needed a real case to come through and prove the new code works. That
was wrong, and I have corrected it. Because every request we are making by hand stops earlier than
the code we fixed, the case we are waiting for cannot arrive on its own. It needs either the
automatic route to run on a new site, or one hand-written request that includes the pages. So it is
a small task, not a wait.

I have put this into the diagnosis loop for an independent read rather than just asserting it,
since it is the kind of claim other people will inherit. Nothing here needs a decision from you
today — but if you want tools to carry cross-mentions when you order them by hand, that is a real
choice and I would rather you made it than I assumed it.

**Later the same evening — the independent check came back empty, and I am not going to dress that
up.**

I said above that I had put the finding into our diagnosis loop for an independent read rather than
just asserting it. It ran, and it produced no answer at all — four rounds of gathering evidence and
then nothing, no conclusion either way. That is a known way for that tool to fail: it finishes
cleanly, so from outside it looks exactly like a run that worked.

It did not refute anything. It simply did not speak. So the finding stands on my own checking, and I
have written that plainly in all three places it appears rather than leaving "sent for review" doing
the work of "reviewed".

The checking itself is solid and it is one query: eleven requests out of eleven from the automatic
route carry the missing information, and none of the sixty-six from every other route do. I also read
the relevant code directly, and I took Friday's change date from our own written record rather than
from a database timestamp — which is what caught my one wrong turn today.

One more thing worth logging while I remember. There is a known failure mode in that tool where it
goes quiet because the file it is reading is too big, and we have a written check for it: measure the
file first. That check passed here — the file is well under the limit — and it went quiet anyway, for
a different reason. I have added that to the warning so the next person does not spend a round
narrowing something that was never the problem.
