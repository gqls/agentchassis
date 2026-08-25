# SUMMARY 2026-08-25b — counting-fact drift (bugs_open/386): the fix is live and armed

## What we're trying to do

Stop the platform accusing its own honest pages of making things up.

Some of our sites print live counts — "feed items collected: 11,513". Those counts come from a
register of verified facts, and a job re-reads them every night. When the real number ticks up, the
register moves and the already-published page does not. Our honesty checker then compares the page
against the register, finds a figure the register no longer holds, and reports the page for stating
an unsupported number.

Nothing false was ever published. The page was correct on the day it was written, and it even prints
the date it was checked. But the report is filed at a severity that refuses to rebuild the page — so
the page cannot be corrected through the normal route — and in the review queue it looks exactly like
a page that invented a figure. That last part is the real cost: it spends the credibility of a
warning we need people to trust.

## Where we've come from

The bug was found on 24 August by a different piece of work, filed separately so the two problems
would not be confused, and left unowned. It arrived with four suggested fixes and no decision.

The day it was picked up, the owner ruled on it: a live counter should be written as "at least N", or
not printed at all. That was the better question — a counter's honest form *is* a lower bound.
"We have collected 11,646 items" is false a minute later; "at least 11,000" stays true for months and
needs nobody to republish anything.

## What we've done

First, established the size of it rather than guessing. Of 295 facts across the estate, only 29 are
the kind a nightly job moves, and once you exclude the ones that never actually change, the real
exposure is six figures on two sites. It is not a fleet-wide problem.

Second, found that the owner's ruling was already running in production on five other facts,
hand-written by somebody before the ruling existed — "more than 10 live production sites", with the
true count of 26 kept behind the scenes and an explicit instruction never to state it. So there was a
working template to copy rather than a design to invent.

Third — and this changed the plan — discovered that the ruling could not reach the case that started
the bug. The five figures on the broken page carry no writing instruction at all. They are not
sentences; they are a chart that draws the register directly. There is nothing to rephrase. That is
exactly the case the owner's ruling had explicitly left the other fix standing for.

So we built that other fix: the register now remembers the values it used to hold. A page printing
last Saturday's number is recognised as having been right on Saturday, rather than accused of
invention. It is off by default and switched on one fact at a time, because the thing being switched
on is "accept more numbers", which should never be a default.

It went through the review council and was approved first time. Then, with the owner's go-ahead, it
was switched on for the five figures on the affected site.

## Where we are now

**It works, and we checked it on the real thing rather than in a test.** Against the live site, the
checker now reports zero problems where it reported five this morning. The five pages are no longer
accused.

**And we checked that it still catches liars.** The worry with any fix like this is that you have
simply switched the alarm off. So we took a number the register has *never* held — one digit away
from a real one, in the same sentence — and put it on the page. It was still caught. The fix spares
the figures we genuinely published and nothing next to them.

Two things are deliberately not finished. The register can now *remember* a value, and we have proved
it *reads* those memories correctly — but we have not yet watched it *record* a new one, which will
only happen the next time one of those counters actually moves. And the underlying problem still
exists for every counter we have not switched this on for, which is why the bug stays open rather
than being marked fixed.

The one remaining exposed figure sits on a different site, and it is a genuinely open question rather
than a repeat of this decision: that counter can go *down* as well as up, which makes "at least N"
unsafe for it in a way it is not for the others.

## Where we're going

Watch for the next nightly run that moves one of those five counters, and confirm the register
records the outgoing value by itself. That is the last piece of the mechanism we have not seen work
in production.

Then decide the remaining figure on the other site — floor, or memory — on its own evidence.

And a note worth carrying: today's work was corrected four times by a colleague working the
neighbouring bug, and corrected them twice in return. Every one of those corrections landed on
something real, including one that changed which of the owner's options actually applied. The record
is more accurate for it than either of us would have managed alone.
