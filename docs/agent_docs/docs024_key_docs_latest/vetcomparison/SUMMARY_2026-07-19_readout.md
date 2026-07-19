# vetcomparison.uk — milestone read-out, 19 July 2026

Written to be read aloud. Figures re-checked against the live system on the day of writing.
For the story in order, see `README_where_we_are.md`; for evidence and commands, `NOTES_` and
`RUNBOOK_`.

## What we're trying to do

Make vetcomparison.uk the place UK pet owners go to compare what veterinary care costs, and the
place practices go to publish those costs. The timing is the whole opportunity: the Competition
and Markets Authority has concluded its investigation into the vet market and is about to require
every practice in the country to publish a standardised, comparable price list. Comparison becomes
possible for the first time, and we intend to be where it happens.

Two rules constrain everything. Anything built on the platform must be generic, so the next
comparison site reuses the machinery rather than having it written again. And we never publish a
figure we cannot show the source for — either the practice gave it to us, or we can link to where
they published it, with the date we saw it.

## Where we've come from

The site was already live, and it was publishing invented prices for roughly 3,100 named real
practices — figures that came from nowhere, under a false claim that the data was our property,
alongside a quotation attributed to the CMA that the CMA never wrote. It had been up since
February. That was the starting position five days ago.

We stripped every price, quarantined the fabricated data in the database, removed the guides, and
wrote a dated factual record of what had been published and what we did about it — the document a
solicitor would be briefed from if a practice or the regulator ever asks. What survived was the
genuinely good half: a real verified directory and a working collection pipeline.

## What we've done

We grounded the regulation from primary sources, and four business decisions followed: publish
per-practice prices only from a practice's own published list, attributed and dated, with an email
opt-out; stay independent rather than seek the RCVS's approved-partner badge, which keeps paid
placement possible later; require at least three practices behind any area average and always show
the count; and respond to the CMA's consultations, siding with independent practices.

Then we rebuilt it properly. The directory was cleaned to 2,109 genuinely verified practices —
junk entries removed, 280 duplicates collapsed, scraped page-titles repaired. Three guides were
rewritten against the final report with every figure sourced. The database was unified so services
and medicines share one structure, with the CMA's mandated 36 services seeded as the canonical list
everything maps onto. A generic exporter now publishes the directory, anonymous area statistics and
consented price lists for any vertical and domain, with the publication rules enforced in code
rather than by good intentions. The claim flow — a practice proves who it is, agrees to recorded
wording, and its own prices publish attributed to it — is built and proven end to end. The site was
registered with the platform, which rebuilt and redesigned it autonomously, and the exporter now
refreshes the data every two days with nobody touching it.

Then the platform put fabricated data back. Its tool-recreation agent rebuilt our practice search
as a component that invented practices in the browser — plausible names, invented postcodes — and
shipped it live, four days after we removed exactly that. We contained it, then fixed it at source:
the fabrication was still sitting in the database, so a future render would have republished it. It
is filed as a platform bug because any site with a data-backed tool is exposed to the same thing.

## Where we are now

The site is live, honest and working: a real searchable directory, three sourced guides, published
area statistics, and claim and opt-out routes. Nothing on it quotes a figure we cannot source — the
only prices anywhere are the CMA's own. The data pipeline runs on its own schedule.

We publish no per-practice prices at all, and that is correct rather than a gap. Our historical
price data was collected without recording where each figure came from, so under our own rule it
can only feed anonymous averages. Per-practice prices begin when we collect them again properly.
Nobody has claimed a listing yet — the flow works, but it has never been used in anger.

One risk is open: the fabrication fix is verified in the database and on the live page, but no
automatic render has run since, so the moment where those two meet has never been tested.

## Where we're going

Immediately, watch that first render and confirm the fix survives it. Before the end of the month,
submit the funding-consultation response, which argues the levy should scale with the size of the
business rather than charging a single-site independent the same as one branch of a national chain.
And when the CMA publishes the draft of the substantive Order — overdue, expected any day — respond
in detail, arguing for low-barrier approval of comparison services and an express right to reuse the
price lists practices will be compelled to publish.

Then the last piece of building: collecting prices from practices' own published lists with the
source recorded against every figure. From December, when the large groups must publish their
standard lists one click from their homepage, that becomes dramatically easier — and showing which
practices have published and which have not becomes both a service to pet owners and the strongest
argument we will ever have for a practice to claim its listing.
