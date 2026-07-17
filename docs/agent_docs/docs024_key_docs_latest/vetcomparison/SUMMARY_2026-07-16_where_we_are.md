# Where we are with vetcomparison.uk

Written 16 July 2026.

There are other docs in this folder that go deeper and I don't want to repeat them here, so:
the PLAN has the build phases and where each one got to, the LEGAL doc is the factual record of
what was published and what we did about it (that one is written for a solicitor to read, so
it's deliberately dry), the RUNBOOK has the operator steps, and there's a briefing on the CMA
consultation plus a draft response to it. This document is just the overall picture, so that
anyone picking this up cold, including me in a month when I've forgotten all of it, can see what
happened, where we got to and what happens next.

## What we found, and what we did about it

The site was live and it was publishing prices we had invented. Not estimated, invented: about
3,124 named real veterinary practices, each with a consultation fee and a prescription fee
attached to it, and none of those figures could be traced back to a source because there wasn't
one. It had been up since the 2nd of February. The giveaway was the shape of the data, 22% of
those practices were showing exactly £48 for a consultation and there were only 46 distinct
prices across the whole set, which is not how real prices behave. On top of that it carried a
notice claiming the data was our proprietary property and shouldn't be scraped, which was not
true either, and the guides had a quote in them attributed to the CMA that the CMA never said,
plus a claim that a named practice in Glasgow charged £33 to write a prescription, again with
nothing behind it.

So that was the position: we were making false statements about identifiable businesses'
pricing, on a site whose entire selling point is that it's accurate. That is the bit that
mattered, everything else could wait behind it.

We stripped it. Every price and the price calculator came off, the false proprietary notice came
off, and the guides came down. The database had the same problem, 997 invented price rows sat
against 235 real verified practices, and those are now quarantined rather than deleted so
there's still an audit trail of what was there. Then we wrote the LEGAL doc, which records what
was published, when, how we spotted it and what we did, with the commit hashes and the dates.
That's our file if a practice or the regulator ever asks us about it, and the reason we wrote it
straight away is that a contemporaneous record is worth a great deal more than one reconstructed
afterwards.

What survived all this is actually the good half. The directory is real, it comes from our own
sweep and verification pipeline, and the medicine price pipeline with its evidence chain is real
too. It was only ever the prototype data that was fabricated.

## The regulatory position, because everything hangs off it

The CMA published its final report on the vet market on the 24th of March. The Order that makes
it binding has not been made yet and is due by the 23rd of September, which is why some of the
dates below are approximate.

Once it lands, every practice has to publish a standard price list, 36 named services in five
categories, VAT included, no free text allowed, priced against six pet size categories, and no
more than one click from their homepage. Large groups have three months from the Order (so
around December 2026) and smaller practices six (around March 2027). There's also a £21 cap on
writing a prescription with £12.50 for each extra medicine on the same consultation, mandatory
disclosure of which group owns a practice, written estimates over £500, and by around September
2027 practices must submit their details to the RCVS "Find a Vet" platform, which the RCVS will
then share with approved third parties who might use it to run a comparison site.

The important part for us is that the CMA explicitly treats comparison services as legitimate,
including ones that collect prices by scraping practice websites. It says so in Part B at 3.320
to 3.321. It doesn't compel practices to let us scrape them and it doesn't grant us an express
right to republish, so there's still a genuine untested question about database right, but we
are not operating against the grain of the remedy, we're operating with it.

## What we decided

Four decisions were open and we've now taken all of them.

We will publish per-practice prices taken from a practice's own published price list, with a
link to where we got it and the date we saw it, a short disclaimer explaining that the CMA has
recognised comparison services of this kind, and an email opt-out for any practice that would
rather not appear. A solicitor should still look at the database right question but we're not
holding the build up for it, and the conditions we've put around this are what narrow the
exposure.

We are not going after the RCVS "approved third party" badge. Taking their data feed would ban
us from ever showing paid placement, and we'd rather keep that option, so we stay independent
and collect our own data. If we do run paid placement later it has to be labelled clearly,
because the CMA's not-misleading standard applies to everyone regardless of the badge.

Aggregates need at least 3 practices in an area before we'll publish a figure for it, and we
always show the n next to the number so people can judge it for themselves.

And we will respond to the CMA consultations, siding with the independents against the
corporates, and separately arguing our own corner.

## Where the build has got to

The live site is now an honest directory of 2,389 verified practices, plus three guides written
against the final report with every figure sourced and a review date on each, plus a "claim your
listing" call to action. There are no prices on it at all at the moment.

Underneath that, we've done the first three phases of the plan. The export code can no longer
publish to a domain unless someone names it explicitly, which matters because it used to default
to vetcomparison.co.uk and we don't own that. The price tables are now unified, so services and
medicines both live in one schema instead of two, and the CMA's 36 items are seeded in as the
canonical list that everything else maps onto. There's a new generic exporter that will build the
directory, the aggregates and the price files and commit them to the site, and it's written to
be config-driven so the next comparison site we do can use the same machinery rather than us
writing it again.

Cleaning the directory turned up more than expected. 20 of the "verified practices" were not
practices at all, they were Yelp and StarOfService and BestLocalRated listing pages, an American
clinic and a college course page, and another 177 had a wheree.com mirror page recorded as their
website instead of their actual site. Those 177 are real practices so they've gone back to
pending for proper re-verification rather than being thrown away.

## The thing I want to remember from today

When we smoke tested the new exporter the attributed prices output came back empty, and it took
a minute to work out that this was correct rather than broken. Every historical price row we
hold has an empty source URL. The provenance was never actually recorded, and it can't be
recovered from the observation records either, so under the rule we've set ourselves (no price
gets published unless we can show where it came from) none of those rows can ever be shown
against a named practice. They can feed the aggregates, because an area median doesn't name
anybody, and that's the 15 aggregate rows we can publish today.

What this means in practice is that per-practice prices start from fresh scrapes, and any
scraper or verifier we run from here has to record the source URL for each individual price as
it goes. If it doesn't, the output is unpublishable and we won't find out until we try to export
it. I've put that in the plan as an acceptance criterion so it doesn't get missed.

It's also worth being honest that our real price coverage was always much thinner than the old
site implied, we have prices of any kind for roughly 330 practices out of 2,700 odd, and now
that we've applied the provenance rule the per-practice figure is zero until we go and collect
them again.

## What happens next

The nearest thing with a deadline is the funding consultation, which closes on the 30th of July.
The draft response is written and sitting in this folder waiting on me to check the levy figures
against the actual Notice and submit it through the portal. The bigger one, the consultation on
the substantive Order, hadn't been published as of today but must appear shortly for the CMA to
hit September, and when it does we turn the briefing's arguments into a proper clause by clause
response.

On the build, the next phase is the claim flow, which is the part that actually matters
commercially: a practice claims its listing, we verify they are who they say they are, and their
own prices go up attributed to them and dated. The exporter already knows how to publish claimed
prices, so it starts working the moment the first practice claims. The Go changes from the last
three phases are written and tested but not deployed yet, they ride the next chassis image build,
and the exporter task stays disabled until that happens.

After that it's re-verifying those 176 wheree practices and collecting prices properly with the
source URLs recorded, and then adopting the site onto the chassis. Come December, when the large
groups have to publish their standard lists, scraping gets much easier because the format is
fixed and the list is one click from the homepage, and at that point we can also show which
practices have published and which haven't, which is useful to pet owners and is also the best
argument we'll ever have for a practice to come and claim their listing.

The strategy hasn't changed through any of this. The CMA is about to make every practice publish
comparable prices, and we intend to be the place where comparing them is actually useful. The
directory and the aggregates carry us now, claimed listings are the product we sell to practices
as their compliance shop window, attributed scrapes fill in the coverage, and we never publish a
number we can't show the source for. That last rule is the whole reason this site is worth
building, given where it started.
