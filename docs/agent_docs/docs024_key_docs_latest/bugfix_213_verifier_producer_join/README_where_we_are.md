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
