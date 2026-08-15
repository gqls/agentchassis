# Where we are — bugfix 213

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-08-10 — what this bug is, and why it is worth a fix rather than a tidy-up

The system files "work items" — little tickets saying something is wrong with a site.
Each ticket has a *type*. Before a ticket is allowed to close, the system re-checks
that the problem is actually gone, and it picks which check to run **by the ticket's
type**.

That works right up until two different parts of the system start filing tickets
under the same type name meaning different things. Then only one of them has written
the re-check, and everybody else's tickets get graded by a test that was never about
their problem. The test isn't broken — it answers its own question perfectly well. It
just answers it about the wrong thing, says "all clear", and the ticket closes with
the fault still there.

That is exactly what has been happening. Two producers file under
`hardcoded_section_colors`: a routine site sweep looking for hard-coded colours, and
the design audit, which files a specific complaint about a specific section and even
writes down its own pass condition. Only the sweep wrote the re-check. So every
design-audit ticket on that route has been graded against the sweep's question.

**The numbers are stark and they got worse while the bug sat open.** Eleven
design-audit tickets have closed clean, and not one of them has *ever* failed to
close — not once, in the whole life of the type. Meanwhile every ticket that did fail
to close belonged to the sweep. When one producer's tickets never fail, that isn't a
producer doing good work; it's a grader that cannot see their problem. When the bug
was written three days ago the count was seven. Four more closed clean since.

The worked example is a good one to hold onto: a ticket on gamesdesign.co.uk said a
section was rendering bright cyan behind white text, nearly unreadable. It closed
three minutes later. The component it complained about had last been touched
ten and a half hours *before* the ticket was even created — so nothing was written,
nothing was fixed, and the page still measures as unreadable today.

## What we've done about it

Two things, and the second is the one that matters.

The first is the obvious repair: the design audit now files under its own type name,
so the two producers stop colliding. That fixes today's instance. It does not stop
the next one.

The second is the real fix. A re-check can now declare **what it actually speaks
for** — and if a ticket turns up that it doesn't recognise as its own kind of
problem, it says so, and the ticket is refused rather than waved through. It lands in
the existing "this needs another look" machinery, loudly, instead of closing silently.

The important design decision, and the one I'd defend hardest: it works by looking at
**the ticket in front of it**, not by keeping a list of who's allowed to file what.
A list was the obvious answer and it had already been tried and rejected by another
thread, for a good reason — anything in this system can be reconfigured to file any
ticket type without a single line of code changing, so a list in the code would look
authoritative while being permanently out of date. Asking "is this my kind of
problem?" needs no list, is never out of date, and correctly handles a producer that
doesn't exist yet.

It is off by default. Nothing changes for any check that hasn't opted in.

## The bit I want to flag honestly

I did not write a test and declare victory. I wrote the test, then went back and
broke the fix on purpose to check the test noticed. It did — but only when I broke
*both* halves. Breaking just the first half left the test green, because the second
half independently covers that route. That's good news for the fix (two
independent protections) and a real gap in the test, and I've written it down as a
gap rather than letting it read as stronger than it is.

## Two things the shared tree did to us today

Worth recording because they'll happen to the next person.

First, I spent about an hour on the *wrong bug*. I picked one that every ownership
check said was free — and every one of those checks reads committed history, while
the session actually working on it had everything sitting uncommitted. I found out by
accident, an hour in, when I noticed a code comment dated today on a file with no
commit from today. I stood down immediately and handed over the measurements I'd
gathered, which turned out to be worth having: that bug is roughly four times bigger
than its own file records, and I proved one live site is serving broken images where
a perfectly good one had already been generated and paid for.

Second, my own commit named seven files and landed six. Another session's commit had
already swept one of them up — we'd both edited the same file, and whoever commits
first takes both sets of changes. Nothing was lost; the work is there. But if you
read my commit you'll see six changes described as seven.

## What's still owed

The code is committed and will ride the next build — it is **not** in the build that
went out today. After that: check it's actually live in the running pod, read the
council's verdict, and then the part that needs judgement — go back through those
eleven closed tickets and grade each one against what it actually promised. Some of
them may well have been fixed by accident. One is confirmed still broken; one is
confirmed fine. The other nine are unknown, and "eleven closed" is not the same claim
as "eleven wrongly closed" — I don't want that number quoted as if it were.

---

**2026-08-11, evening — the third ruling (D3) is built: a detector for the next time this happens.**

The bug we fixed was one verifier grading another producer's work items against the
wrong question, and closing them clean. The fix let a verifier say *"this isn't my
question"* — but a verifier only says that if somebody remembers to write the line.
The council warned about exactly that when it approved the fix, and the owner ruled:
build something that notices.

That is now built. Once a day, a small job asks the live database one question: is
there any kind of work item that has a checker attached, where the items are
plainly arriving from **two different sources**, while the checker has never said
which of them it speaks for? If it finds one, it files a work item saying so — one
per kind, deduped, with no handler attached, because the fix is a human writing four
lines of code, not a robot retrying something.

Two things about it are worth saying in plain words.

**It shows its working when it finds nothing.** A daily check that only ever says
"0 problems" is indistinguishable from a check that has quietly stopped looking —
and we have been bitten by that here before. So the report always names the case it
DID see and chose not to file: today it says, in effect, *"`hardcoded_section_colors`
still has two sources, and it is fine, because its checker now declares what it
grades."* There is also a switch that re-runs the same census with that suppression
turned off, and today it produces the original bug as a live finding. So the zero is
a zero that looked.

**Telling two sources apart is harder than it sounds, and the obvious ways are all
wrong.** Every intuitive marker — who created the row, which pipeline it came from,
which check name it names, or simply "the shape of the data is different" — fires on
kinds of work item that have only ever had ONE source. I measured all four against
the live database before choosing, and the one that survives is a fuzzy comparison
of the data's shape: two shapes that share more than half their fields are the same
source with a revision, two that share almost none are two sources. That sounds
loose and it is pinned by real numbers: in our whole fleet, same-source pairs share
at least two-thirds of their fields, and the one genuine two-source pair shares
**nothing at all**.

The council sent the first version back for revision, which was the right call and
cost about ten minutes. Two of its five objections were things I had actually built
but had not listed, so it could not see them. Two were real: it made me *check*
rather than assume that the "parked, nobody will touch this" status I was using
behaves the way I thought (it does, but only in combination with another field —
worth knowing), and it made me justify building this beside the existing per-site
check framework rather than inside it (the framework is site-shaped; this question
has no site).

One thing I found while measuring that matters more than the detector: the work
items that this bug was originally about are **already being re-found**, 14 of them
today, and 13 of those have already closed again — unchecked, because the new
category still has no checker of its own. That is the next piece of work (D1), and
it is no longer a theoretical gap: it is happening weekly, in the open.

---

**2026-08-12, afternoon.** The next piece of work (D1) had one instruction attached to it:
*measure before you re-route*. The idea on the table was that these new colour findings
might simply be handed to the fixer we already have. The honest position was that nobody
had checked — we had one example where it obviously wouldn't work and had generalised from
it, which is precisely the move this bug exists to punish.

So I checked, by the only method that can't drift: I took the fixer's own repair function
out of the codebase, pulled the actual page content out of the live database, and ran the
one on the other. **It changes nothing. Not on any of the 15 findings, and not anywhere on
those 15 sites** — 61 pieces of content tested, none of them altered. I ran three sanity
cases through the identical pipeline first, two that had to come out "changed" and one that
had to come out "unchanged", so we know the test can tell the difference.

The reason is simple enough to say in one line, and it is the same reason a different bug
found a fortnight ago: **the fixer only knows two words.** It can replace a colour with
"the site's primary colour" or "the site's secondary colour", and that is the whole of its
vocabulary. Every one of these 15 findings asks for something else — a text colour for a
dark section, a heading colour, a muted variant. You cannot answer those with two words, on
any page, ever. So re-routing is off the table, and now for a stated reason rather than an
impression.

**While measuring that, I found something worse than the gap we were describing.** We have
been saying these items "close unchecked". One of them has now been caught closing
*wrongly*. On finetuning.uk the fixer ran, reported in its own record that it had changed
nothing at all, and the item was marked complete anyway — and the design audit came back
the next day and filed the identical finding again. Nothing on that page had changed in
between; I checked the timestamps. So this is not a theoretical hole any more: we have a
worked instance of a green tick over a repair that provably did not happen, caught only
because the audit happened to look again.

Two smaller things fell out of the same afternoon, both of which change what D1 costs.

The first: the plan of record was to build the checker by reading the "acceptance test"
each finding carries. I read all 15 of them. Ten ask for something you can only see in a
running browser; two ask for things no automated check could ever settle — one wants "no
visible seam", another wants text that is "visibly #f0eeff **or equivalent**". And they are
written fresh, in English, by the auditor each time: the same defect on the same component
of the same site produced two differently-worded tests on consecutive days. You cannot
write a checker against that. If we want it, the *auditor* has to start emitting a
machine-readable criterion — which is a bigger and different job from writing a checker,
and should be priced as one.

The second: the other lane working the neighbouring problem (bug 122, the 226 parked
contrast items) reached its own conclusion the same afternoon, and it is a good one — grade
these things when the audit next comes round, rather than at the moment the item closes.
That avoids the objection that has been blocking both of us, which is that nobody wants a
web-browser fetch sitting on the completion path. Their approach transfers to our case and
is actually cheaper for us than for them, because our re-detection loop demonstrably works
(that is how we caught the false green) and theirs has never once fired. The two lanes are
now formally independent, so we are not waiting on each other.

**What I have not done, and why.** I have not built anything. The choice between "refuse
the completion when the fixer says it changed nothing" (cheap, needs no browser, catches
exactly the failure above) and "grade it at the next audit" (better answer, needs the audit
to report which pages it looked at) is a design decision on a shared mechanism, and this
estate's rule is that those go through review before they go in. The measurements above are
what that review needs in front of it, and they are all written down now.

---

**2026-08-13.** Built the first of the two things I described yesterday, and stopped short
of the second on purpose — for a reason I did not see yesterday and want to put on record.

**What is built.** A completion gate that refuses to mark one of these colour items "done"
when the fixer's own report says it changed nothing. *A fix that changed nothing is not a
fix.* It is small, it needs no web browser and no page fetch, and it is switched on for
this one kind of item only — every other kind of work item in the fleet is untouched by
construction, which is the estate's standing rule for anything that alters a shared
mechanism. It is committed and will start working at the next fleet build.

One thing was nearly a wasted day. I began writing this as a "verifier", which is the
obvious home for it, and it cannot live there: the verifier is asked its question *before*
the fixer's report is saved, so it would have read the previous run's numbers and looked
like it was working. I found that by reading the order of two lines rather than by
assuming. It now sits next to an existing check that reads the same report at the right
moment — which was the correct home all along.

I also mutated the code four different ways to confirm the tests actually catch a break,
and the first attempt was worthless: I deleted a block in a way that stopped the program
compiling, saw "FAIL", and nearly recorded that as proof. A build error is not a test
catching anything. Redone properly, all four come out red as they should.

**What I could not do, and it needs you.** The cluster login token has expired — everything
that talks to the live system is refused. That blocked the council review: the script
printed a convincing submission reference and then failed to send anything, so no review
exists. I did not write a "submitted for review" note on the commit, because it wasn't. The
submission is written and ready; it needs one command once the token is back.

**Why I did not build the second thing.** The plan was to grade these items when the design
audit next comes round — if it does not re-report the fault, the fault is gone. The
neighbouring lane reached the same conclusion for their problem and they are right for
theirs. But their audit is a **measurement**: a browser computes a contrast ratio, and
silence means the number came out fine. Ours is a **language model reading the page and
writing prose**. Its silence is not the same thing. A model that does not mention a
problem on Tuesday has not established the problem was fixed on Monday night — it may
simply not have said so this time.

I have direct evidence the wording varies: the same fault on the same site, filed one day
apart, came back described in different words both times. What I do not yet know is whether
the *set of faults it finds* varies, and that is the question that decides whether this
approach is safe. It is a cheap thing to measure and I could not run it today, because it
needs the database. **So the second piece is specified and not started, deliberately** —
building retraction on top of an unstable detector would close real faults on nothing more
than the model's mood, which is a worse failure than the one we are fixing.

---

**2026-08-14.** The gate went out with the new build and I can prove it is in there — I asked
both copies of the service directly, and asked in a way that could have said no: I looked for
three things in the running program, one of mine that had to be present, one long-standing one
that had to be present, and one made-up one that had to be absent. All three answered
correctly, on both copies.

**But it has never once run, and I want to be straight about why, because it changes the story
I told you earlier this week.** The thing that dispatches these colour items — the improvement
sweep — was switched off on Tuesday, by another lane, after it turned out to be costing three
times what was expected. Nothing has dispatched one of these items since. So nothing completes,
so no false "done" ticks are being minted, and my gate has had nothing to catch.

Which means: **the leak I found stopped on Tuesday because a switch was turned off, not because
of anything I built.** The audit is still finding these faults — a sixteenth site turned up
yesterday — they just sit in a queue. What the gate actually buys us is that turning that sweep
back on is now safe. That is a real thing to have, but it is not "fixed the leak", and I would
rather say so than let a good-looking green tick stand in for it.

There is one item sitting in the queue right now, on mortgagecalculator.co.uk, with nothing
that will ever pick it up while the sweep is off. **That single item is the cheapest proof
available** — send it to the fixer deliberately, and we get to watch the gate refuse it for
real, on a real site, in about a minute. That needs your say-so, because it is a live action
rather than a test. And it would settle a second question at the same time: this bug cannot
close on its own terms, because the fix removed the very traffic that would have demonstrated
the fix, and one deliberate dispatch answers both.

Two smaller things worth knowing. The council approved the gate on the second attempt, and the
first attempt's rejection was fair — it caught that I had proved my new code harmless to the
things it does not touch while proving nothing about the existing code I had moved to make room
for it. That is now covered. And the other lane working the neighbouring problem shipped their
version of the "grade it at the next audit" mechanism while I was working, which is good news
and also means my remaining piece has to be built on top of theirs rather than beside it —
their own reviewers left a note saying a third copy of that pattern should be shared code, and
mine would be the third.

---

## 2026-08-15 — the second half is built, and both bugs are closed

You said half two could proceed and that 213 and 216 could close. All three are done.

**What half two actually is, in plain terms.** The gate we built last time could only say no.
When the design audit spots a dark section and hands it to a fixer that cannot fix it, the gate
now refuses to let that ticket be stamped "done" — which was the whole problem — but it left the
ticket sitting there marked "failed" with no way of ever being closed honestly, even if someone
later fixed the page by hand. There are four tickets in exactly that state right now. Half two
is the way out: **when the audit stops reporting the fault, the ticket closes itself, and the
reason it gives says what was observed rather than just "done".**

The interesting part was deciding *when* to believe silence, and the answer came from the data
rather than from judgement.

**It has to be about the whole site, not the page.** The audit records which page a fault is on
as free text, and it is genuinely free — the live values include "index", "all" and "all pages",
and those last two were written on the same day. There is no way to turn that into a page. So
the question we ask is about the site: did this audit look at this site and report nothing at
all?

**It has to be about the site rather than the individual fault, and this is the one I nearly got
wrong.** The obvious design tracks each fault separately. But the name of the audit was changed
from "design-audit" to "visual-design-audit" between the 12th and the 13th, and that name is
baked into every ticket's identity — so one site has two tickets for the same fault under two
different names. A per-fault design would have read that single rename as fifteen faults being
fixed at once. I only found it because I listed the actual rows before designing anything.

**And it takes three silences, not one.** We measured earlier that the audit re-reported a known
unfixed fault seven times out of seven. That sounds like it never misses, but seven out of seven
only proves the miss rate is below about 35% — so closing on one silence would close a live fault
about a third of the time it ran. Three silences brings that to about 4%. Three is simply the
first number that gets under 5%. It is arithmetic, not caution.

**The thing that nearly bit us, which had nothing to do with the design.** To count silences you
have to write a counter somewhere on the ticket. There is a housekeeping job that sweeps up
tickets which have sat untouched for 48 hours — and the database automatically stamps "last
touched" on any write at all, so there is no way to write the counter without resetting that
clock. Writing a counter every fifteen minutes would have made those tickets permanently
un-sweepable. Nothing would have looked wrong; the damage would have been a cleanup that simply
never happens. We avoid it by never writing to tickets that are actively queued. I found it by
going to look for who reads that column before writing the code, not by any test — no test can
see a scheduled job's configuration.

**One thing I got wrong and want on the record.** I wrote a test to prove that a garbled reply
from the audit is ignored rather than treated as silence. The test passed. It also passes if you
delete the protection entirely — it was being satisfied by the test framework refusing the call
rather than by our code refusing it. The only thing that caught it was deliberately breaking the
code to check the test noticed, which is now routine here. Five other tests in the same file were
fine; this one looked identical to them. It is fixed, and both the mistake and the general shape
of it are written down for other people.

**The review panel approved it first time**, in eleven minutes, with five advisory comments. Three
of them were checkable and I checked all three rather than filing them. The best one asked
whether the safeguard that stops one audit closing another audit's tickets might accidentally
seal shut the very four tickets this work exists to free — a fair question I had not asked, and
the answer is no, they all match. Another asked whether a truncated reply could be mistaken for
silence; the audit has made 4,088 calls and never once come close to being cut off. The third
found a genuine gap: I had reorganised a piece of code that all six audit types run through, and
only tested my own. That is now covered.

**On the two bugs.** 216 was fixed, live and proven back on the 8th and was only being held open
by a rule that had since been replaced. 213's own closing condition had become impossible to
satisfy — the fix removed the very traffic that would have demonstrated it — so the closure note
says plainly that one branch of it has never run in production, and explains why the query for it
will read zero for ever, so nobody later reads that zero as "it never worked".

**Two things are honestly still open and I would rather say so.** First, none of this can run
yet: the sweep that drives these audits is still switched off for cost, so half two is built and
inert, exactly as the gate was. Second, we still do not know why some of these tickets carry a
result that belongs to a different piece of work altogether. The investigation into that died on
the 14th when the account hit its usage limit — and I found today that the limit was lifted about
ninety minutes later, so it had been re-runnable for a day and nobody noticed. I have restarted
it. I had also written in our notes that the fleet was out of action until 1 September, which was
wrong and is now corrected; the limit message says that, but it is what would happen if nobody
intervened, not a forecast.

---

**2026-08-15, later that day.** The owner asked me to switch the improvement loop on briefly and
then off again, or else to trigger the audits by hand, so that the safeguard we built would
actually get a run rather than sitting there untested. I did the second, and I want to explain
why, because the first turned out to be a worse idea than it looked.

There were supposedly two things that could set these audits going. One of them, the design
rotation, looked like the safe choice because it only looks for problems and never tries to fix
them. It turns out it could never have worked at all: it drives a different agent from the one
that actually writes the audit findings, so it has never been a route to this code in the first
place. Every handoff in this folder has said otherwise, and that was simply wrong. On top of
that, it picks a site only if that site has not been looked at for a week, and nothing on the
estate is currently a week overdue — so switching it on would have produced a run that selected
no site, did nothing, and looked for all the world like a clean test.

The other one, the improvement sweep, does work, but it does not only inspect. It also promotes
what it finds and hands it to the machinery that makes real changes to a real site — and that
machinery is running right now, checking every sixty seconds. The sweep chooses its site by
whichever has gone longest without attention, so I would not have been able to say in advance
whose website was about to be edited. That is not a thing to switch on to satisfy a curiosity.

So instead I ran the audit itself, directly, against one site I picked deliberately —
gamesdesign.co.uk. That runs exactly the same code by exactly the same route, but only inspects,
and files what it finds as a suggestion rather than an instruction. Before firing it I checked in
the code that findings are recorded as suggestions, and checked that the machinery which makes
changes only ever picks up instructions, so there was no path from my test to anybody's live
page. I also confirmed the safeguard was genuinely present in the software currently running,
rather than assuming it, because the servers had been replaced again since yesterday.

**It worked, and it did the right thing.** The audit looked at the site, found the dark section
still there, and reported it — and because it had something to report, the safeguard correctly
did not close the old ticket. That is the outcome we wanted: the whole danger with this kind of
rule is that it quietly closes a ticket for a fault that is still present, and here it had every
opportunity and declined. Better still, it did not touch the old ticket at all — not even to
write a note on it — which matters more than it sounds, because writing to a ticket for
bookkeeping reasons is the exact thing that can make it invisible to the housekeeping process
later. We had reasoned that it would not; now we have watched it not.

**What I cannot claim.** This proves the half that refuses to act. It does not prove the half
that eventually closes a ticket when a fault really has been fixed, because none of the four
outstanding faults have been fixed, so the audit has nothing to be silent about. That half needs
three consecutive clean looks at a site that has genuinely been repaired, and we cannot
manufacture that honestly.

One small surprise worth writing down: the run filed a fresh ticket for the same fault alongside
the old one, because the old one had been marked failed and the system treats that as finished
for the purpose of avoiding duplicates. So the site now carries two tickets for one dark section.
That is harmless and it cannot grow beyond two, but the count in the report reads two rather than
one, and I would rather you saw that here than wondered about it later. The cost of all this was
one audit, and four new suggestions on that one site, none of which anything will act on while
the sweep is off.

---

**2026-08-15, later still — and I was wrong about the interesting part.** I had recommended
against running the other three sites, on the grounds that they would all come back the same and
tell us nothing new. The owner said run them anyway. He was right and I was wrong, and it is
worth being precise about how, because the reasoning I used was the problem rather than the luck.

Two of the three did come back exactly as I predicted. The third, the mortgage calculator, did
not: **the audit looked at the site and found no dark section at all.** That matters because it
is the other half of the safeguard — the half that starts counting towards eventually closing a
ticket, which until this afternoon had never once run outside a test. The ticket is now marked
"one clean look so far". It needs three before it closes, so it is still open, and correctly so.

I then checked whether the audit was right to go quiet, rather than taking its word for it. I
fetched the actual page and its stylesheet and read the colours directly. **The fault described
in the ticket is not on that page, and on the evidence it may never have been.** The ticket says
the call-to-action falls back to a gold background with dark text on it, which would indeed be
unreadable — but that fallback only applies if the site has not set its own colour, and it has:
the panel is near-black text on a light cream background, which is about as readable as it gets.
The audit had been reasoning about what *would* happen in a case that never arises. So the
safeguard is not closing a repaired fault; it is on its way to closing a ticket that arguably
should not have been raised. That is still exactly what we want it doing, and I would rather it
reached that outcome by observation than by anyone going in and deleting the row by hand.

What I cannot tell you is whether that colour was always set, or whether it was set by the five
page re-renders another thread of yours ran on that site at about a quarter to two, twelve
minutes before my audit. I have no copy of the old stylesheet, so both stories are open and I am
not going to pick one.

**Why my recommendation was wrong, in one line:** I told you the result in advance and used that
as the reason not to look, and the thing I was confident about turned out to rest on a sentence I
had copied from our own handoff without re-checking it — a sentence that had stopped being true
that same afternoon, when your re-renders ran. Being sure what an experiment will show is a
reason to run the cheap version of it, not a reason to skip it. I have written that up in the
fleet-wide ledger of wrong calls, because the tally there is what eventually justifies automating
a check.

**One caution about what we can now do.** Because the ticket needs three clean looks, I could
simply run the audit twice more in the next few minutes and watch it close, which would prove the
whole mechanism end to end. I would rather flag something first: the rule counts *runs*, and
nothing stops those runs happening seconds apart. The three-strikes figure was chosen from
evidence gathered over days, where each look was at a page that had had time to change. Three
looks at an unchanged page in one minute are really one look repeated, so closing the ticket that
way would satisfy the rule without earning it. On the normal schedule this never comes up. It
only comes up because I have just built a way to drive it by hand.

---

**2026-08-15, the last run of the day — and it turned the whole thing round.** You said run the
two, so I ran the first of them. It came back the other way: **the audit looked at the same site
again and this time it did find a dark section.** The count of clean looks went straight back to
zero, and the ticket is open exactly as it was.

I want to be careful about what that means, because the obvious reading is wrong. This was not
the site changing under us. I had saved a copy of the page after the first audit; I fetched it
again after the second and the two are **identical, byte for byte** — same page, same size,
nothing touched it in between. And the thing the second audit complained about was already there
during the first one. So the first audit did not see a fixed site. **The first audit simply
missed it.**

That is the single most useful thing we have learned today, and it is worth saying plainly:
**the audit is not consistent with itself.** Ask it the same question about the same page forty
minutes apart and you can get opposite answers. We had always suspected this — the three-strikes
rule exists precisely because one clean look cannot be trusted — but until this afternoon it was
an assumption backed by a run of seven where nothing ever went wrong. Now we have watched it go
wrong once, directly, with everything else held still.

**And that is the safeguard being vindicated, not embarrassed.** If we had built it to close a
ticket on a single clean look, it would have closed that ticket at five past two, and forty
minutes later the same audit would have raised the same kind of fault on the same site. Instead
it counted one, then correctly threw the count away. The part that threw it away had never once
run outside a test before today. So across this afternoon we have now seen three of the four
parts of this safeguard work on real traffic: refusing to close a ticket while the fault is
reported, starting to count when it goes quiet, and wiping the count when it speaks again.

**A correction I owe you from an hour ago.** I told you I had confirmed the audit was right to
go quiet by reading the actual page. What I actually confirmed was that *the specific fault
written on that ticket* is not present — that part is still true. But going quiet is a claim
about the *whole site*, not that one ticket, and the fault the second audit found was in a
different part of the page altogether. I checked a smaller thing than the safeguard was
claiming and presented it as though it covered the whole claim. It is a slightly awkward one to
own because the check itself was sound and well evidenced — it was pointed at the wrong
question, which is much harder to notice than a sloppy check. It is written up in the wrong-calls
ledger.

**I stopped there rather than running the second of the two you approved.** With the count back
at zero, getting to three now means firing the audit over and over until three clean looks happen
to fall in a row — and on the evidence of today that is roughly a coin toss each time. That would
not be proving the last part works; it would be running the test until it gives me the answer I
had already told you I wanted. So the last part of the safeguard stays unproven, and I would
rather it stayed honestly unproven than be closed that way. If you want it demonstrated, the
clean way is to let it happen on the real schedule once a sweep is back on.
