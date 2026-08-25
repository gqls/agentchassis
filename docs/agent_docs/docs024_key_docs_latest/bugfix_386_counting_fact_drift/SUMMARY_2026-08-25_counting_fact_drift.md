# SUMMARY 2026-08-25 — counting-fact drift (bugs_open/386)

## What we're trying to do

Stop the platform convicting its own honest pages. Some site copy prints live counts that a nightly
job re-reads. When a count ticks, the register moves and the deployed page does not, and our honesty
checker reports the page for inventing a figure — at a severity that refuses the rebuild which would
fix it. We want the register and the checker to agree about numbers that were true when they were
published, and — following the owner's ruling — we want pages to stop printing brittle exact counts
in the first place.

## Where we've come from

The bug was found by the `bugs_open/364` lane's fleet claims census on 2026-08-24, filed separately
so the two mechanisms would not be conflated, and handed on as a residual when `bugs_open/380`
closed. It arrived with four candidate fixes and no owner. It sat unowned for a day.

## What we've done

Verified the mechanism first-hand at the code rather than taking the bug file's word for it, and
measured the scope: 295 facts across the estate, 29 that a nightly job moves, 13 matched strictly
enough to convict a page, on six sites. Showed the bug re-arms daily by comparing the register
against itself a day apart. Established that the fix's raw material already exists — we supersede
register rows rather than overwriting them, so 315 old versions survive back to mid-July, including
the exact value the convicted page prints.

Corrected three documents in the process: the bug file's premise that counting facts only ever
increase (ours fall when their table is reaped), and two documents describing a stale-page
discriminator built on a field that has no implementation in the code.

Then the owner ruled: express a counting fact as "at least N", or cancel the claim. That promoted
the cheapest candidate to the default and added an option the list had not contained — don't print
the number at all. The ruling's own caveat proved to be understated: the live example it cites has
grown from 4,068 to 7,281 since the figure was written down.

## Where we are now

Lane opened, standing five written, nothing changed on any live site. The plan is in three parts,
in this order: implement the ruling on the 13 exposed facts, but only after measuring what each
proposed "at least" would start accepting; then the durable fix, which teaches the register to
remember former values so a page printing last week's number is recognised rather than accused; then,
only if the evidence justifies the cost, re-render pages when a count they print has moved.

## Where we're going

Next is the measurement that gates everything else: for each of the 13 facts, diff the checker's
findings between today's register and a candidate "at least" register, and read every finding that
disappears. A finding that vanishes is something newly vouched for, and if a genuine invention is
among them the terms are too broad and get narrowed before anything is armed. Where a page does not
need its number, the recommendation will be to remove it — the owner's stronger option, and the only
one with no ongoing cost.
