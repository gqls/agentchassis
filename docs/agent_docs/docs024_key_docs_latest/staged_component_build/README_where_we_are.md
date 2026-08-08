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
