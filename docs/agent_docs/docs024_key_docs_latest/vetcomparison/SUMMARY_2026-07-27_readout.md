# vetcomparison.uk — milestone read-out, 27 July 2026

Written to be read aloud. Previous read-out: `SUMMARY_2026-07-26_readout.md` (written the evening
before, after closing bug 061). This one exists because the question that read-out named as the
next thing to measure has been answered, and the answer changes the plan. For the story in order
see `README_where_we_are.md`; for evidence and commands, `NOTES_` and `RUNBOOK_`.

## What we're trying to do

Make vetcomparison.uk the place UK pet owners go to compare what veterinary care costs, and the
place practices go to publish those costs. The CMA has finished its investigation into the vet
market and published a draft Order that will require every practice to publish a standardised,
comparable price list. Comparison becomes possible for the first time and we intend to be where it
happens.

Two rules constrain everything. Platform changes must be generic, so the next comparison site
reuses the machinery. And we never publish a figure we cannot show the source for.

## Where we've come from

The site went live in February publishing invented prices for around 3,100 named real practices.
In mid-July we stripped every price, quarantined the fabricated data and wrote a dated factual
record. What survived was the good half: a verified directory and a collection pipeline.

By last night we had closed the last of the fabrication bugs and diagnosed why the site still feels
thin — not missing pages, but **under-publishing**. Almost everything we hold is barred by our own
provenance rule, and collection has been switched off since March. The plan was: restart
collection, starting with a small pilot to check that a live run records where each fact came
from.

## What we've done

**We answered that question, and it did not need the pilot.** The verifier cannot record
provenance. The code that saves results looks for a "source URL" **inside the AI's answer** — a
field the AI is never asked to produce — while the component that actually fetched the page, and
therefore genuinely knows the address, is not connected to the component that saves the record.
All 2,970 records we hold confirm it: the field is not empty, it is **absent**. This is now filed
as bug 100.

**We measured the size of the prize without touching anyone's records.** Running the real verifier
over practices would have written AI-extracted, unsourced facts over their current entries just to
learn a percentage. Instead we used a read-only probe built from the exact same pattern-matching
code the live system uses, over 100 practices chosen at random.

**About one practice in five publishes its company number on its home page** (22 of 100); reading
terms and privacy pages as well takes it to **three in ten** (30 of 100). Sixteen of the twenty
distinct companies found match a real veterinary company at Companies House.

**The ownership finding is the valuable one.** Those 30 hits are only 20 companies, because eight
of the hundred practices are owned by the same company — VetPartners — and three more by CVS. That
is exactly the picture our own records currently get wrong: we have 870 practices flagged
"independent" that also carry a group name. A company number turns that from a label into evidence
anyone can check.

We also found (bug 101) that the scrape step is configured to look like a six-page crawl with a
sensible fallback and is in fact a single page fetch — four of its settings are read by no code at
all.

## Where we are now

**The crawl you approved is blocked, and blocked for a good reason.** Restarting collection today
would spend a fleet-wide crawl to produce data we still could not publish. The fix is a genuine
code change — review, build, deploy — and it should carry both bugs together, since they touch the
same step.

Importantly, the obvious one-line fix is the wrong one. Adding "and tell us where you found it" to
the AI's instructions would make the column fill up while making the problem worse, because the
provenance would then be the AI's own claim about its own evidence. That is precisely what this
site was remediated for in July.

**One thing is deliberately left open.** Our probe read raw web pages; the live system fetches
through a third-party service whose setting for keeping or discarding page footers is, because of
how the code is written, never actually sent — so it falls back to its own preference and we do
not know what that is. Company numbers live in footers. We tried to settle it against 2,450 pages
already fetched that way and the evidence is genuinely mixed. Rather than guess, it is recorded as
unsettled at the top of bug 101, because it decides whether the planned fix would even work. One
real run answers it.

## Where we're going

1. **Settle the footer question** — one verification run, read what comes back. It is the cheapest
   step and it gates the rest.
2. **One code change covering both bugs**: record the URL actually fetched, and make the scrape
   read the pages its config already claims to read. Through the review gate, then built and
   deployed.
3. **Then restart collection**, staged rather than all at once, and re-measure the hit rate on real
   runs rather than on a probe.
4. **Then ownership** — rebuild it as something derived from evidence rather than asserted, which
   is what the company numbers make possible.

The three cheap page-building items from the standing review queue (directory index, guides index,
compliance calculator) are unaffected by any of this and can proceed in parallel whenever you want
them.
