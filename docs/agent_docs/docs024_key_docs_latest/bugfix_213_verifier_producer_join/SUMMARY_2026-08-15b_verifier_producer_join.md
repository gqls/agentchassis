# SUMMARY 2026-08-15b — bugfix 213: the safeguard stopped being theoretical

*Written to be read aloud. Previous in the series: `SUMMARY_2026-08-15`, `…-08-14`, `…-08-11b`,
`…-08-11`. This one marks a real turn: this morning's summary closed the work as built-but-inert,
and this afternoon three of its four parts ran on live traffic for the first time — one of them
producing a finding about the platform that we did not previously have.*

---

## What we're trying to do

Make sure a ticket that says "done" means something, and make sure a ticket that says "failed"
can eventually stop saying it.

The platform files a work item when it finds a fault on a site, and hands it to an agent that is
supposed to repair it. Everything downstream trusts the ticket. A ticket that closes green
without anything being fixed teaches the platform the page is fine. But the opposite failure is
just as real: a ticket stuck at "failed" for a fault that has since been fixed by someone else
stays stuck for ever, because nothing in the system was ever able to say "that is not broken any
more".

## Where we've come from

The first half of this work stopped tickets being closed green by a repair agent that had
plainly changed nothing. That shipped, and it left a hole it created: those tickets now cycle to
"failed" and sit there, because the agent handling them provably cannot fix them and there is no
route back out.

The second half is the route back out. It watches the audit that raises these faults. If that
audit looks at the site and says nothing about that kind of fault, three times running, the
tickets close on that observation — not because anyone claimed a repair, but because the thing
that reported the fault has stopped reporting it.

That was built, reviewed, approved and shipped yesterday and this morning. And it was completely
untested in the real world, because the only scheduled job that drives those audits is switched
off for cost. This morning's summary said so plainly: built and inert, exactly like the half
before it.

## What we've done

The owner asked for it to be given a real run — either by switching the sweep on briefly, or by
triggering the audits by hand.

Switching the sweep on turned out to be the wrong lever twice over. The one that looked safe
because it only inspects and never repairs could never have worked at all: it drives a different
agent from the one that writes these audit findings, so it has never been a route to this code,
despite every handoff in this folder saying it was. It would also have selected no site at all,
because nothing on the estate is currently overdue by its seven-day rule. The other one does
work, but it does not only inspect — it also promotes what it finds and hands it to the machinery
that edits real pages, and that machinery is running every sixty seconds. It picks its site by
whichever has gone longest without attention, so nobody could have said in advance whose website
was about to be changed.

So we ran the audit itself, directly, at four sites chosen deliberately. That runs identical
code by an identical route, but only looks, and files what it finds as a suggestion that nothing
will act on. Four runs, then two more at one site.

**Three of the four parts of the safeguard have now worked on real traffic.** It refused to close
tickets at three sites where the fault was still being reported. It began counting at a fourth
where the audit went quiet. And when that same audit spoke again forty minutes later, it threw
the count away — which is the part that protects against closing something prematurely, and the
part that had never once run outside a test.

## Where we are now

The four stuck tickets are all still stuck, which is the correct outcome and was the correct
outcome at every step. Nothing was switched on or off; the scheduler is in exactly the state the
owner left it in on the 14th. Four new suggestions per audited site are sitting inert.

**The most valuable thing we learned was not about our safeguard at all.** Two audits of the same
page, forty minutes apart, gave opposite answers — and we ruled out every innocent explanation:
the page was byte-for-byte identical across both, nothing repaired it in between, and the fault
the second audit found was demonstrably present when the first one looked. The first audit simply
missed it.

That matters well beyond this ticket, because audit output is used across the estate as if it
described the page. It does not; it is one sample, and the unreliable direction is the one
everybody relies on — a fault *not* being mentioned. It is written up fleet-wide so the next
person to say "the audit no longer flags it, so it is fixed" has something to read first.

It also settles an argument in our favour that we had only assumed. The three-strikes rule was
chosen on evidence that never actually caught the audit being wrong. Now we have caught it. Had
we built this to close a ticket on a single clean look — which was seriously considered, and is
what the cheaper design would have done — it would have closed a ticket at five past two on a
site that was not clean.

Two things are honestly unfinished. The last part of the safeguard, the one that actually closes
a ticket, has still never run in production. It could have been forced this afternoon by running
the audit repeatedly until three clean looks fell in a row, and that was declined: with the audit
disagreeing with itself at roughly a coin toss, that is not a demonstration, it is running the
test until it gives the answer you wanted. And a correction is on the record — a check that was
offered as confirming the audit's silence turned out to confirm something narrower than the claim
it was supporting.

## Where we're going

Nothing here needs doing next, and that is a deliberate state rather than an abandoned one.

The last part will prove itself, honestly, the first time a genuinely repaired site gets three
clean looks on the ordinary schedule — which requires a sweep to be back on, and that is a cost
decision, not a technical one. Until then the code is live, exercised, and waiting, and we now
know from direct observation rather than assumption that its caution is warranted.

The finding about the audit disagreeing with itself is the piece most likely to matter to other
work, and it is filed where other threads will meet it before they need it rather than after.
