# Where we are — the whole-page shrink floor's axis (bugs_open/293)

Plain prose, append-only, newest at the bottom.

## 2026-08-17, first session

**What the bug is, in plain terms.** When the pipeline rebuilds a page, something has to notice if
the rebuild has quietly deleted most of the writing. There is a guard that does exactly that: it
compares each section's text before and after, and refuses the whole save if a section loses more
than half of it. The problem is *what it counts as text*. It strips the HTML tags out and measures
what is left — but a page's styling instructions (CSS) and its interactive code (JavaScript) sit
*between* tags, not inside them, so all of that counts as "text" too.

The consequence is the one you would guess. If a rebuild replaces an article with a stylesheet, the
guard sees the character count go **up** and waves it through. That is not hypothetical: it is what
happened to a webdesign.co.uk article on 14 August, which served an empty page for about 23 hours.
The lane that fixed that incident fixed the *section editor* — one of the two doors this write can
come through — and filed this bug for the other, bigger door: whole-page rebuilds.

**Why they stopped rather than fixing both.** They could prove their fix was safe, because the
database keeps a copy of every section it overwrites, so they could replay 117 real edits through
the new rule and check it did not start refusing good work. Whole-page rebuilds do not leave that
trail — they delete every section and write fresh ones, so the "after" side appeared to be missing.
Changing a safety rule fleet-wide on evidence that does not cover the path is how a guard starts
blocking legitimate work, and then gets switched off. So they wrote down what evidence was needed
and left it.

**The evidence turned out to be there.** The "after" copy isn't missing — it is the row that is
*live right now*. And each live row records when it was created, which is independent proof that it
was written by the rebuild that had just deleted its predecessor. That gives 1,079 exactly-paired
rebuild writes — nine times the evidence the sibling fix had — with a check that could have gone
wrong and didn't: not one live row is older than the deletion it supposedly replaced. As a second
check I ran the same method over the *other* path and it reproduced the other lane's three known
findings to the character, plus one they had missed.

**What it says.** Two things, and they point the same way.

Switching the guard to count only the words a reader sees would have refused **none** of those 1,079
rebuilds. Across a wider, less reliable set of pairs it would have refused exactly **one** write in
eight days, and I read that one by hand: a genuine tightening of some prose on robot-hands.com,
which the operator could have let through with a config setting that already exists.

And the thing the current guard is missing is not small. Rather than wait for another incident, I
took all 1,079 real sections and *constructed* the failure — deleted every word, left the styling and
code exactly as they were, which is precisely the shape of the August incident — and asked the
guard. **It allows the total deletion of the prose on 724 of the 1,060 sections it looks at.** The
proposed measure allows none.

**Two things I did not expect to find.**

First, the guard's own cut-off is now wrong. It ignores sections under 500 characters, on the
grounds that short things shrink legitimately. But 500 was chosen when the count included all that
CSS, so on a real reading of the prose it excludes over half the sections on the page. Lowering it
roughly doubles the protected surface and — measured at every step down — does not refuse a single
additional write.

Second, there is a **third** copy of the same mistake, and it is the worst of them: an older,
page-wide version of this check that refuses a save if the whole page loses three-quarters of its
text. On the same measurement it would allow a whole-page prose wipe on 337 of 366 pages. So this is
not really a bug at one call site; it is one judgement that three places each decided how to measure
for themselves. Fixing only the one this bug names would leave the same trap for the fourth.

**And one thing I got wrong.** My first run reported a textbook hollowing on a leopardessconsulting
page and I nearly wrote it up as a finding. It wasn't real. Section names can repeat on one page,
and my query had paired one copy's "before" with a different copy's "after". Hand-checking the single
result is what caught it — the number looked completely plausible until I opened the page's history.
Worth the five minutes twice over, because the *same* wrong assumption is in the shipped guard: on
those pages it compares an arbitrary copy against an arbitrary copy, and which one wins depends on
the order the database happens to return rows in. A production defect found by tripping over it in
my own measuring instrument.

**Where this goes next.** The evidence and the measuring harness are committed — and the harness is
in the repo on purpose, because the instruction "re-run this calibration before changing it" was
sitting in the code with no way for anyone to do it. Next is the fix itself, which needs a decision
about how wide to go, and then the council.

## 2026-08-17, later the same day — the fix is written, and the council found something

**The fix, in plain terms.** All three guards now count only the words a reader would actually see.
The cut-off below which they don't bother judging a section moved from 500 characters to 200, because
500 was chosen back when the count included all the styling instructions — on an honest reading of the
prose it was excluding more than half the sections on a page. And the oldest of the three guards, the
page-wide one that was buried as an unnamed block of code in the middle of a much longer function, now
has a name, a test, and a switch an operator can turn off without rebuilding the software. It had
never had one.

**I asked Fable to plan it before writing anything, and that was worth doing.** It found a fourth
place the same mistake lives — one I had missed — and that find changed the design. I had intended to
simply lower the 500 to 200 where it was defined. But something else uses that number for a completely
different purpose: deciding whether two sections are similar enough to be paired up. Lowering it in
place would have quietly changed that behaviour with no evidence behind the change. So the number
became a setting passed in by each guard, and the old one stayed exactly where it was.

It also talked me out of what I thought was the tidier fix. I wanted to make the shared decision
measure the text itself, so no caller could get it wrong. That would have broken the measuring
harness, which works precisely because it can feed the real decision *either* measure and compare. So
the check is now a test that fails the build if any caller measures the wrong way — a test can ask
"was this measured properly?"; a type cannot.

**Then the council reviewed it and sent it back, and the reviewer was right.** The blocking question
was one line long: *does the page-wide guard's filter actually match any rows?* It selects sections
marked "deployed" — and there is a known trap recorded in our own notes where the equivalent column on
a neighbouring table never holds that value at all. If the same were true here, the guard I had just
carefully extracted, named, tested and documented would never run on any page, and the test suite
would have cheerfully confirmed it worked, because tests hand the code whatever rows you tell them to.

It was one query. "Deployed" is by far the commonest value — 1,575 sections across 617 pages, 85% of
all pages — so the trap doesn't apply here and nothing needed changing.

**But the lesson is the part I want to record, because I got it wrong in a way that felt careful.** I
had kept that filter *deliberately* unchanged, and said so in the code, on the grounds that quietly
changing which sections a guard looks at is exactly the sort of unreviewed behaviour change that
causes incidents. That reasoning is sound. What I never did was check whether the filter matched
anything in the first place. "I didn't change it" protects you from *introducing* a fault and does
nothing whatsoever about *inheriting* one — and moving a line of code into a function with a proper
name, a paragraph of explanation and a test suite makes it look far more trustworthy than it did as an
anonymous fragment, without adding a single piece of evidence. Every figure I had published about that
guard rested on a number I had not counted.

**Three other things the reviewers asked for, all now done.** The page-wide guard, when it can't read
the page at all, lets the save through rather than blocking it — inherited behaviour I chose not to
change. A reviewer pointed out that it did so *silently*, which means a future content loss would be
diagnosed as "the guard should have caught that" when in fact it never ran. It now leaves a record,
under a label that says plainly nothing was blocked. Another asked whether our idea of "visible text"
matches the one the page assembler uses when it decides a section is empty and can be dropped — a fair
worry, because if they disagree, a save could pass this guard and have its content thrown away
downstream anyway. They *do* differ in how they're built; I measured 6,585 real sections and none of
them could be dropped that way, and there's now a test that keeps it so. The third was that I had
oversold my own test as "closing" a gap when it only narrowed one; I've corrected the wording and made
the test do the broader thing it was claiming.

**Where this stands.** Everything is committed. It's Go code, so none of it does anything until the
next chassis image is built and rolled — and a thing worth knowing came out of checking that: the
sibling fix from this morning, which everyone believes went live, has *not*. The running image was
built yesterday. So both halves of this correction will start working at the same moment, on the next
roll. Round two is with the council now.

## 2026-08-17, evening — it rolled, it's live, and it has already done its job ten times

The new chassis image went out (v1.0.1307) and the fix is in it. I checked that by asking the running
program what it was built from rather than trusting the version number, because a version tag is a
claim and the binary is a fact — and both pods gave the same answer, with a control to prove the check
itself works.

**Both halves arrived together, as expected.** This morning's sibling fix — the one everyone thought
had already shipped — went live on this same roll alongside ours. That's the tidier outcome, since
they share a threshold.

**And it's been exercised, which is the part I care about.** Eleven page rebuilds ran in the first
three-quarters of an hour and none was blocked. On its own that number means nothing: "no refusals"
looks identical whether the guard is working perfectly or not looking at anything at all. So I checked
the thing behind it — **ten of those eleven sections were big enough for the guard to judge**, and it
judged them and let them through. That's exactly what the measurements predicted. And on those same
ten sections, the old measure would have allowed someone to delete every word of the prose and not
noticed.

**Two things I want to be straight about.**

First, no refusal has actually fired yet. Everything says it will work when one is needed — nine
deliberate sabotage tests of the code, and the guard demonstrably running on real pages — but it hasn't
yet had to say no to anything, and I'd rather tell you that than imply otherwise. Based on the history,
expect roughly one a week.

Second, I chose **not** to force one. I could trigger a refusal deliberately, but on this path a save
that *isn't* refused destroys the page's writing — the exact damage we're fixing — and I don't think
that's a sensible thing to do to a live site to prove a point I've already proven nine other ways.
There is a safe version: temporarily tighten the threshold in the database so the next ordinary rebuild
gets refused, then put it back. Nothing is destroyed, because a refusal writes nothing. But it does
block someone else's page build for a few minutes and it edits a shared setting other work depends on,
so it's your call rather than mine. **Say the word and I'll do it.**

**One more thing rides the next roll.** The council asked for a refinement after the image was built —
one of the three guards should refuse rather than shrug when it can't take a measurement at all. That
change is written, approved and committed, but it isn't in the current image. It's a small
improvement, not a gap: the guard behaves exactly as it always did in that situation, and the two
guards that run immediately after it would refuse anyway.

The bug is closed. What's left over is written down rather than forgotten: about a tenth of sections
are still too small for the text guards to judge, and the platform's older coverage test still can't
see this write path at all.

## 2026-08-18 — we made it refuse something, on purpose, and watched it hold

You asked for the safe version of the test, so here it is and it worked.

**First, the new image carries everything.** The latest chassis build has all three of our code changes
in it, including the refinement the council asked for after the last roll. I checked by asking the
running program what it was built from, not by trusting the version label.

**The test, and why it couldn't have gone wrong.** I sent the pipeline a request to save one page — and
the content I sent was *that page's own content, copied exactly*, with the threshold turned up so high
that even an identical copy counts as a loss. That makes both possible outcomes harmless: if the guard
works it refuses and writes nothing, and if the guard were broken it would write the page's own content
straight back, which changes nothing. I picked a small contact page on an internal site, and I took a
copy of everything first.

It refused, in 12 seconds:

> *PAGE CONTENT REGRESSION REFUSED for page "contact" — the incoming sections carry **581 chars of
> visible text** against 581 deployed across 3 sections (100% kept, floor 150%), with stylesheet and
> script content excluded from both sides. Nothing was written.*

**The number in there is the whole point of this project.** That page is 7,343 characters of HTML and
only **581** of them are words a reader sees. The old measure counted all 7,343. That gap — nearly
thirteen to one — is the room in which an article can be replaced by a stylesheet and the guard
congratulates it on growing.

And the page is genuinely untouched. I checked not just that the text is identical but that the
database rows are the *same rows* — so the delete step never even ran.

**I also checked the opposite, because a guard that refuses everything would pass the test above.** Six
real page rebuilds happened on this image while I was working; four were large enough for the guard to
judge, and it let all four through. The only refusal on the entire image was the one I caused.

**And the test found a bug in my own work.** To make the test safe I had to set that threshold above
100%, and it let me — because unlike the two guards either side of it, the one I extracted didn't limit
how high the setting could go. Harmless as I used it, but a typo in a config file could have set it to
150% permanently and quietly refused every page save on that step, which on this path can also stall the
rest of a build. It's now capped like its siblings, with a test that fails if anyone removes the cap.

Worth saying plainly: that gap survived a council review, ten deliberate sabotage tests and a green test
suite — because every one of them exercised the *default* value. **A setting nobody has ever set is
untested by definition, and I only found out by using it.** Written up in the wrong-calls log.

I also fixed something misleading in our own runbook: it told the next person to read the build version
out of the service's log, which on this service returns *another team's text*, because the pipeline logs
whole messages and those quote the same phrase. It now reads the version out of the program itself.

That's the thread. The bug is closed, the fix is live and has been seen doing its job in both
directions, and the cap rides the next roll.

## 2026-08-18, closing — this time the new build hadn't actually reached the chassis

The one thing left over was the cap I added yesterday evening. It isn't running yet, and the reason is
simple: **no new chassis image has been built since the one already running.** The two long-lived
chassis pods haven't restarted, the version is the same, and the program still reports the same commit
it did before. Decisively, the code that build was cut from is *older* than my cap — by about an hour.

What made it look like a deployment is worth knowing, because it will happen again: the fleet is full of
short-lived pods that spawn for one job and exit, so a listing always shows things that started minutes
ago. Five had started at 16:40. All of them on the *old* version. **Fresh pods are not fresh code** —
and the giveaway is that the two permanent chassis pods didn't move at all.

To be clear about the record: the two previous "fresh build" notices in this thread *did* carry new
code, and that's how the fix and its council refinements got live. This one didn't.

Nothing is stuck. The bug is closed, the fix is live and has been watched working in both directions,
and the cap is committed with a test that fails if it's removed. It'll come into effect on whichever
build gets cut next — no action needed from me, since a release goes out fleet-wide and you run that.
