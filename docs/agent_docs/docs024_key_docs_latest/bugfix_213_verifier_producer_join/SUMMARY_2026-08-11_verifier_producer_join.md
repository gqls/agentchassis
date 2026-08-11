# SUMMARY 2026-08-11 — the verifier that answered the wrong question

First summary for this lane (`bugs_open/213`). Written to be read aloud.

---

## What we're trying to do

Make sure that when the system says a fault has been fixed, it actually has been.

The platform finds problems with our sites and files them as tickets. Something then
tries to fix each one. Before a ticket is allowed to close, the system re-checks that
the problem has really gone — that check is the only thing standing between "a robot
said it did the work" and "the work happened". We are trying to make that check
trustworthy, because everything downstream believes it.

## Where we've come from

The re-check is chosen by the ticket's *type*. Each type has one test registered
against it, and the system runs that test.

That arrangement quietly assumes something it never verifies: that everybody filing a
ticket of a given type means the same thing by it. Three days before we picked this up,
another thread noticed that assumption had broken. Two different parts of the system
were filing tickets under one type name — a routine sweep looking for hard-coded
colours, and the design audit, which files a specific complaint about a specific
section and even writes down its own pass condition. Only the sweep had written the
re-check.

So every design-audit ticket was being graded by a test about somebody else's problem.
And here is what makes it nasty rather than merely wrong: the test isn't broken. It
answers its own question perfectly well. It just answers it about the wrong thing, says
"all clear", and the ticket closes with the fault still sitting there. Nothing errors.
Nothing looks unusual in any log.

The numbers make the case better than the explanation does. Eleven design-audit tickets
have closed clean, and **not one of them has ever failed to close** — not once, in the
entire life of that ticket type. Every ticket that *did* fail to close belonged to the
sweep. When one party's tickets never fail, that is not a party doing good work; it is
a grader that cannot see their problem. When the bug was written the count was seven.
Four more closed clean while it sat waiting.

The clearest single example: a ticket said a section on one of our sites was rendering
bright cyan behind white text, close to unreadable. It closed three minutes later. The
thing it complained about had last been touched ten and a half hours *before the ticket
even existed* — so nothing was written, nothing was fixed, and the page still measures
as unreadable today.

## What we've done

Two things. The first is the obvious repair: the design audit now files under its own
ticket type, so the two parties stop colliding. That fixes today's instance and nothing
more.

The second is the real fix, and it is the one worth explaining. A re-check can now
declare **what it actually speaks for**. If a ticket turns up that it doesn't recognise
as its kind of problem, it says so, and the ticket is refused rather than waved
through — it goes back into the existing "this needs another look" queue, loudly,
instead of closing silently.

The design decision I would defend hardest is *how* it recognises its own tickets. It
looks at the ticket in front of it. It does **not** keep a list of who is allowed to
file what. A list was the obvious answer, and another thread had already tried and
rejected it for a good reason: anything in this system can be reconfigured to file any
ticket type without a single line of code changing, so a list in the code would look
authoritative while being permanently out of date. Asking "is this my kind of problem?"
needs no list, can never go stale, and correctly handles a filer that doesn't exist
yet. It is switched off by default, so nothing changes for any check that hasn't opted
in.

It went through the review council and was approved first time. I read the objections
rather than just taking the verdict, and two of them were worth acting on — one asked
for a test I genuinely hadn't written, and one asked for a check I genuinely hadn't
run. Both are done; the check came back clean.

I also did not simply write a test and declare victory. I wrote it, then deliberately
broke the fix to see whether the test noticed. It did — but only when I broke *both*
halves. Breaking just one left it green, because the other half independently covers
that case. That is good news for the fix and a real gap in the test, and I have written
it down as a gap rather than letting it read as stronger than it is.

## Where we are now

The fix is live on both servers and I have proved it is actually running, not merely
built. That proof had to be redone: the project's own guidance was rewritten this
morning and banned the verification method I had used, on the grounds that it had
produced three confidently wrong readings elsewhere in a single day. It also turned out
that the release I was reading contains three different revisions under one label, so
the label proves nothing about what is in it. I re-checked properly rather than
defending the original claim. The answer held — and improved, because the stricter
method gave me a control I had previously talked myself out of.

Three things are honestly still open.

**The fix is deployed but has not yet been exercised.** It hasn't had occasion to
refuse anything. Deployed and working are different claims, and I am not making the
second one. A sister change made exactly that mistake last week and had to correct
itself the next day.

**I have moved one problem rather than removed it.** The design audit's tickets now
have their own type — but that type has no re-check of its own yet. So those tickets
have gone from being graded by the *wrong* test to not being graded *at all*. Both
close clean. Two of the review council's seats spotted this independently and they were
right to. It is the main thing outstanding.

**The eleven wrongly-closed tickets have not been repaired.** One is confirmed still
broken, one is confirmed fine, nine are unknown. I want to be careful with that number:
"eleven closed" is an upper bound on the damage, not a count of it, and I would rather
it wasn't quoted as though it were eleven broken pages.

## Where we're going

Three decisions were taken today and they set the whole of the remaining work.

**Build the missing re-check.** Every one of those eleven tickets already carries a
written, mechanical pass condition — "this section's background must be dark", "this
text must meet the contrast standard". Nothing in the system currently reads that
field. Making it read it is what turns these tickets from unverifiable into verifiable,
and it is the proper closure of the gap above. It needs its own review round, because
it means asking the closing step to look at a real rendered page, which we have
deliberately kept it from doing until now.

**Leave the eleven alone and let the system find them again.** The site-inspection
rotation is live and running weekly, so anything still broken will be re-filed by
itself. This depends on the previous item, though — until the new ticket type has a
re-check, the rotation will find each fault and then lose it again. If the re-check
slips, this decision needs revisiting.

**Build a detector for the next collision.** The protection we shipped is opt-in, which
means the next time two parties converge on one ticket type, the same bug returns
unless somebody remembers to write the protection — and remembering is precisely the
discipline that already failed here. The review council's architecture seat made this
point and it is the sharpest thing in the whole review. The query that finds such a
collision already exists; the work is turning it into something that runs on a schedule
rather than something a person has to think to run.

One last thing worth recording, because it cost an hour and it will cost somebody else
one. I began on a different bug entirely, having checked four ways that nobody was
working on it. All four checks read committed history, and the person working on it had
everything saved but uncommitted. I found out by accident and stood down. The hour
wasn't wasted: the measurements I had gathered showed that bug is roughly four times
larger than its own write-up records, and I proved one of our live sites is serving
broken images while the correct, already-paid-for image sits on the server one filename
away. I handed that to its owner rather than competing with them.
