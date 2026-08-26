# SUMMARY 2026-08-26 — dispatch throughput

*(First summary of this lane, which opened 2026-08-18.)*

## What we're trying to do

Make the build pipeline serve many more domains at once. The estate is meant to grow from tens
of sites toward thousands, and this lane's job is to find what actually limits how fast work
across sites gets dispatched and processed, and to raise that rate without losing safety.

## Where we've come from

The lane began with a plan built on one belief: that the dispatch scheduler ran one turn at a
time — a single-flight slot — so adding a second scheduled row would roughly double throughput.
We shipped that second row with guard rails, and on the 25th we measured it properly. The belief
fell. The scheduler is fire-and-forget and always has been, since March: a scheduled row was
never a slot. The two rows fired about a second apart, locked in phase, picked the same site 94%
of the time, and by the following morning nearly 60% of all claim attempts were being wasted
bouncing off each other. Safety held throughout — zero double-handles across thousands of
handlers — but the gain was ten to fifteen percent, not double. Every document carrying the old
belief was corrected in place, the scheduler defect was filed (bug 398), and four options went
to the owner.

## What we've done

The owner ruled option B on the morning of the 26th: retire the second row and speed up the
original, from a fire every 90 seconds to one every 60. It was applied within the hour
(migration 637), passed council review, and the ruling is enforced mechanically — a daily check
now fails loudly if anyone re-enables a second row, or goes faster than ruled before the
spending governor exists. The same day: an overnight nine-hour account-credit outage was
measured and separated from dispatch behaviour; a false alarm in the daily safety check (a
crashed handler's clean-up time being counted as a live overlap) was diagnosed and the check
taught the difference; and a handover from another lane, chased to its mechanism, became a new
structural finding — bug 413.

## Where we are now

B is live and its first day looks the way we hoped. Fires are evenly spaced at a 60-second
median. Work spread across 22 different sites in two hours, where the old pair had piled onto
one. Wasted claim attempts fell from roughly six in ten to about one in ten. The fleet is
clearing around 270 claims an hour against a ceiling of about 300 at this cadence, and the
constraint is now capacity, not demand — over a thousand ready items are waiting across some
thirty sites. The full day's before/after verdict comes tomorrow morning.

Bug 413 is the important open finding: the site selector ranks sites by their oldest waiting
item, but the loader serves each picked site by priority. So one old, worst-priority item can
hold its site at the head of the queue indefinitely while never itself being processed — and
sites behind it can wait without bound. One site sat over ten hours with 73 ready items and no
service, while every fleet-level number read healthy, because the damage is an absence. The lane
now measures a per-site "worst wait" alongside the totals, which is the meter that sees it.

## Where we're going

Tomorrow, the full 24-hour reading of B, including that per-site floor. Then the next lever,
already measured to be the binding one: batch size, from five items per turn to eight. Then the
LLM spending governor, which gates any further speed-up — the outage made its case concrete,
because with no governor an empty balance stops everything at once for hours. And a decision is
needed on how to fix 413; the candidates are ranked in the bug file, from making the two
orderings agree to a simple age floor that bounds how long any site can wait.
