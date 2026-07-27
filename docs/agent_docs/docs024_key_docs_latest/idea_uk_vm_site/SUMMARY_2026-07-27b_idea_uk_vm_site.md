# idea.uk — where we are, 27 July 2026 (second read-out of the day)

*A new file, not an edit of `SUMMARY_2026-07-27`. That one closed on the first sale. This one
exists because the last engineering item on the list is done, and the question in front of us has
changed shape.*

## What we're trying to do

Run idea.uk as a complete, working business: a site that teaches people how to take an idea from
"I wonder" to something real, with free tools along the way and one paid product at the end — a
£29 report that assesses the idea you send us, checks it against what already exists, and tells
you honestly whether it is worth your time. And to make it good enough that the rest of the
portfolio can be built to the same standard rather than each site being invented from scratch.

## Where we've come from

For most of this month the work was correctness. The site and the tool were two halves that could
not see each other, so we put them on one address. The chrome was broken, so we fixed it. The
free taster didn't work in the page, so we fixed that. Then the paid report itself turned out to
be selling something slightly different from what it delivered, so we extended the engine to
match the promise.

In the last two days that turned into something sharper. Preparing a routine deployment, we read
the code properly and found two faults that were quietly costing money: the £29 report could be
taken without paying, and a visitor could choose the identity they were rate-limited under.
Both were live, neither had been exploited, and nothing in the system would ever have reported
either. We fixed both and then attacked our own site to prove it.

Then someone paid. The whole chain — request, review, approval, a real card payment, a signed
receipt from Stripe, delivery, the slot freeing itself — ran in production for the first time.

## What we've done since this morning

**Closed the last engineering item.** When the report engine was working and the service
restarted — which happens every single time we deploy — the work was lost, because it only ever
existed in memory, and the order sat marked "running" forever holding one of five slots. Five of
those and the site quietly stops accepting work. That is fixed, deployed and proven by
deliberately breaking a spare order and watching it recover.

The fix we had written down for it would have cost us money. It said "reset anything still marked
running". Under the other way the site can be configured, the customer has already paid by that
point, and resetting would have sent them a second payment link. We found that by reading rather
than by testing, and the fix now asks one question about each stranded order — has this person
been sent a bill? — before deciding what to do with it.

**Measured what a report costs.** About sixty-four cents of model spend, across two of the five
AI calls in one report; the other three weren't recorded, so the real total is somewhat higher. We
can't say exactly what, and we're not going to guess. The useful part is the shape: it is small
enough against £29 that margin is not the thing to worry about. That measurement gap is now closed
— the engine records its own cost on every call — so the next real order will tell us exactly,
without anyone going to look.

**And we caught ourselves nearly reporting a false result.** The same run was supposed to confirm
three small writing faults were fixed. Every check came back clean. It wasn't a clean result: two
of those three faults couldn't have shown up in that run at all, because of what we happened to
submit and because the engine decided none of the further ideas it found were good enough to
include, so the section containing one of them was never written. Looking for a fault in a
paragraph that doesn't exist proves nothing. The fixes are in the running software and covered by
tests, and they are still not proven in a real report.

## Where we are now

**The site is complete, the product works, and the engineering queue is empty.** Nine guides in
journey order, four tools, a paid report that has been bought and delivered, security holes
closed and proven closed, orders that expire on their own, interrupted work that recovers itself,
and costs that now record themselves.

**What it has never had is a customer who isn't us.** One sale, to the owner. Everything we have
proven is that the machine works. None of it is evidence that anyone wants the product. That is
now the binding constraint, and it is not an engineering one — which is a real change, because
until today there was always something broken worth fixing first.

Two small things sit between us and finding out. Two DNS records are outstanding that make our
email provably ours — which mattered little while the only recipient was the owner, and matters a
great deal the first time we email a stranger a £29 report. And the page that sells the report
puts the request form nearly three-quarters of the way down, which is a long way to ask someone to
read before they can act.

## Where we're going

Stop building and find out whether anyone wants it. Clear the email records, shorten the path to
the form, then put the site in front of people who aren't us — and let real traffic decide what to
fix next, because with no visitors we cannot tell a good change from a bad one.

Two questions are open and worth deciding deliberately rather than drifting into. Whether to
improve the report we have or add a second, different paid tool so the two can be compared against
each other honestly. And when to start treating idea.uk as the template the other sites are built
to, which is what it was always meant to become and is now close to being.

The caution from this morning still stands, and today added to it. A product can be complete,
verified and demonstrably working and still never have done the thing it exists to do. It can also
pass every check you thought to run and still not be proven — if the checks were ones it could
never have failed.
