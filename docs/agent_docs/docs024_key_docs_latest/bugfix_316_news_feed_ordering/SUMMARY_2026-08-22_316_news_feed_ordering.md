# SUMMARY — 2026-08-22 — the news feed that served the alphabet

## What we're trying to do

Nine of our sites refresh their news feeds automatically. A job wakes every six hours, asks the database
which sites are due, and refreshes the first five it gets back. We are trying to make sure that "the
first five" means *the five that have waited longest*, and not *the five whose names come earliest in
the alphabet* — which is what it actually meant.

## Where we've come from

The problem was reported three days ago, on 19 August, by the lane that had been investigating silent
row limits elsewhere in the platform. It noticed that the query picking sites ended with an instruction
to sort them by domain name and take five. Sorting by name is stable, so whenever more than five sites
were due at once — which was every single run — the same names won. Four sites at the back of the
alphabet were always late; five at the front were never late. The split fell exactly on the cut-off.

That report was filed, correctly, as two separate problems. One was fairness: the alphabet was deciding
who suffered. The other was capacity: even with perfect fairness, the sites collectively ask for more
refreshes than the job supplies. It named the second as a decision for the owner rather than a defect,
and it was right to.

Nobody had picked it up since. The lane that filed it closed the same day and explicitly listed this as
one of three tickets standing on their own.

## What we've done

**We confirmed the problem was still there, and found it had got considerably worse.** When filed, the
worst-affected site was seven minutes behind. By this morning `webdesign.co.uk` — last alphabetically —
had not been refreshed for thirty-one hours on a six-hour schedule, and had been passed over at four
consecutive runs while sitting in the queue, eligible, every time.

**We found two errors in the original report's own recommendations**, both by reading code rather than by
reasoning about it.

The first: its suggested fix, taken literally, would have created a worse problem. The query has a second
job nobody had noticed — it is also how a brand-new site gets its news sources set up. Such a site has no
schedule yet, and the suggested ordering would have put it permanently at the front of the queue, because
nothing about refreshing can give it a timestamp it does not have. If its setup ever quietly failed it
would have jammed the queue for every other site, indefinitely and silently. We sent those sites to the
back instead, and wrote down why: the failure we avoid is unbounded and invisible, the one we accept is
bounded and obvious.

The second: its suggestion to raise the limit from five to ten would have changed nothing. There is a
*second* limit of five in the step that processes the sites. Raising only the first would have handed
over ten sites while five were still processed — and, worse, the instrument we use to detect this problem
would have stopped reporting it. We would have measured success where nothing had improved.

**We fixed the ordering.** One query change, reviewed by the platform's council (approved), applied at
10:54Z today. It is live now. The site that had been starved is currently first in the queue.

**We built a check for the whole class of problem, not just this instance.** The distinction it encodes
is the part worth keeping: when a queue's work *gets finished and leaves the list*, processing it
alphabetically is only a delay. When the work *comes back round on a clock*, processing it alphabetically
means the back of the list is never served at all. We already had a check watching this exact query, and
it could not see the problem — it counts how many rows come back, and that number is identical whether
the order is fair or not. It had been truthfully reporting "at the limit" every six hours for days.

We proved the new check works before the fix could hide the evidence: run against the live system
beforehand it found exactly one problem across 194 agents; run afterwards, none.

## Where we are now

The fairness half is **fixed, reviewed, applied and live**. The query now serves whoever has waited
longest.

The relief itself is **not yet observed**, and we are not claiming it. Everything verified so far
concerns the query; the next actual refresh run is at 14:37Z. Until a run has happened and the pattern of
lateness has visibly moved off the alphabet, this stays open. That is the estate's own bar and it is the
right one — several things in this report would have looked fine at the query level and still not worked.

The check is built and committed but **not yet running**: it is Go code, so it waits for the next fleet
release, and its image has never been built.

The capacity question is untouched and is the owner's: nine sites ask for **42 refreshes a day**, the job
supplies **20**. Fair ordering shares out a shortfall; it cannot remove one.

## Where we're going

Three things, in order.

First, after 14:37Z, check that lateness has *rotated* rather than disappeared — it cannot disappear
while demand is twice supply, and if it appears to, the measurement is wrong.

Second, the owner decides the capacity question: spend more (run more often, or raise **both** limits
together), or accept that some of the configured schedules were more ambitious than intended — one site
asks to be refreshed every three hours.

Third, there is a structural issue we deliberately did not fix: setting up a new site and refreshing an
existing one compete in the same limited queue. Our change makes the trade-off safe rather than
resolving it, and it is written down as a follow-up rather than quietly worked around.
