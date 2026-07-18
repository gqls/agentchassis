# vetcomparison.uk — what we set out to do, what we've done, where we are, where we're going

Written 18 July 2026. A read-out summary. Companion: SUMMARY_2026-07-18_bugs_journey.md covers
what went wrong along the way, including the one that came back.

## What we set out to achieve

To rebuild vetcomparison.uk into the best place in the UK for pet owners to compare what veterinary
care costs, and secondarily to compare practices themselves — timed to a genuine regulatory
opening. The Competition and Markets Authority concluded its investigation into the vet market in
March, and is about to require every practice to publish a standardised, comparable price list.
That creates a moment where comparison becomes possible for the first time, and we want practices
to treat our site as the place it happens.

Two constraints shaped every decision. Everything built on the platform had to be generic, so the
next comparison site reuses the same machinery rather than having it written again. And we would
not publish data we had no right to publish, or figures we could not stand behind.

## What we've done

We began by discovering that the existing site was publishing invented prices against roughly
3,100 named real practices, and had been since February. We stripped every price from the site,
quarantined the fabricated data in the database, removed guides that contained an invented CMA
quotation, and wrote a dated factual record of what had been published and what we did about it —
the file a solicitor would be briefed from if a practice or the regulator ever asks.

We then grounded the regulatory position from primary sources rather than memory, and you took
four decisions that define the business: we publish per-practice prices only from a practice's own
published list, attributed and dated, with an email opt-out; we stay independent rather than
seeking the RCVS's approved-partner badge, which keeps paid placement available later; area
averages need at least three practices behind them and always show the count; and we will respond
to the CMA's consultations, siding with independent practices.

From there we built the thing properly. The directory was cleaned — junk entries removed, 280
duplicates collapsed, names repaired — down to 2,109 genuinely verified practices, each linking to
its own website. Three guides were rewritten against the CMA's final report with every figure
sourced and a review date shown. The database schema was unified so services and medicines share
one structure, and the CMA's mandated 36 services were seeded as the canonical list every price
maps onto. A generic exporter was written that publishes the directory, anonymous area statistics
and consented price lists for any vertical and any domain, with the publication rules enforced in
the code rather than by good intentions. The claim flow — a practice proves who it is, agrees to
recorded wording, and its own prices publish attributed to it — was built and proven end to end.
And the site was registered with the platform, which then rebuilt and re-designed it autonomously.

## Where we are now

The site is live and working: a real directory of 2,109 practices, searchable, three sourced
guides, published area statistics, and a claim-your-listing route with an opt-out beside it. The
exporter refreshes the data every two days without anyone touching it — the pipeline from database
to public page now runs on its own.

Today's work was a correction rather than an advance. The platform's overnight rebuild had
replaced our homepage with a search component that generated fake veterinary practices — invented
names and postcodes, produced in the browser — and had added claims about pricing and ownership
data that we do not publish. That is the same failure the whole project exists to remedy,
reintroduced by our own tooling four days after we removed it. It is now off the site: the verified
homepage is restored and live, and I have confirmed against the published page that no generated
data and no unsupported claim remains.

The honest position on coverage is that we publish no per-practice prices at all yet. Our
historical price data was collected without recording where each figure came from, so under our own
rule it can only feed anonymous area averages. Per-practice prices begin when we collect them
again, properly.

## Where we're going

Immediately: make sure the platform cannot rebuild fabricated data onto the site again. Restoring
the page fixes today; the durable fix is at the specification level, so the next automated render
produces the right page rather than reverting to the wrong one. That is the first job, and there
are a handful of build items waiting for your review in the admin queue.

Before the end of the month: the response to the CMA's funding consultation, which closes on the
30th and is drafted — it argues the levy should scale with the size of the business rather than
charging a single-site independent the same as one branch of a national chain. And when the CMA
publishes the draft of the substantive Order, which is due imminently, a detailed response arguing
for low-barrier approval of comparison services and an express right to reuse the price lists
practices will be compelled to publish.

Then the last piece of building: collecting prices from practices' own published lists with the
source recorded against every figure, which is what fills the per-practice comparison the site is
ultimately for. From December, when the large groups must publish their standard lists one click
from their homepage, that collection becomes dramatically easier — and showing which practices have
published and which have not becomes both a service to pet owners and our strongest argument for a
practice to come and claim its listing.

The strategy in a sentence: the CMA is about to make every vet publish comparable prices, and we
intend to be where that comparison is actually useful — directory and statistics today, claimed
listings as the product we offer practices, collected prices for coverage, and never a number we
cannot show the source for.
