# vetcomparison.uk — what we set out to do, what we've done, where we are, where we're going

Written 18 July 2026. A read-out summary of four days' work (14–18 July).

## What we set out to achieve

The goal was to rebuild vetcomparison.uk into the best place for UK pet owners to compare
veterinary prices — and, secondarily, services — riding a genuine regulatory moment: the
Competition and Markets Authority has just concluded that vets must publish standardised,
comparable prices, and we want practices to treat our site as the place that comparison
happens. Two constraints shaped everything: every platform change had to be generic to the
chassis so the next comparison site reuses it, and we would not republish other people's data
unlawfully — no price without a source or the practice's consent.

## What we found, and what we did about it

The existing site turned out to be publishing invented prices against roughly 3,100 named real
practices, under a false claim that the data was proprietary, with a fabricated regulator quote
in its guides. It had been live since February. We treated that as the emergency it was:
stripped every price and the calculator off the site, quarantined 997 fabricated database rows,
removed the guides, and wrote a dated factual record of the discovery and remediation — our
defence file if a practice or the CMA ever asks. What survived was the genuinely good half:
a real, verified directory of practices and a working price-collection pipeline with an
evidence chain.

We then grounded the regulatory facts from primary sources: final report 24 March 2026, the
binding Order due by 23 September, a mandated 36-service price list in a fixed format
(no free text — deliberately comparable), a £21 cap on written prescriptions, compulsory
ownership disclosure, and an RCVS data platform from 2027 that will feed approved comparison
sites. The CMA explicitly recognises independent comparison services, including scraping-based
ones. Four business decisions got taken on the back of that: we publish per-practice prices
only from a practice's own published list, attributed and dated, with an email opt-out; we stay
independent rather than seeking the RCVS badge, keeping paid placement open as a future line;
area statistics need at least three practices behind them, with the count always shown; and we
will respond to the CMA's consultations, siding with independent practices and arguing for an
express right to reuse the mandated lists.

## What we built

In four days the whole plan except the final scraping phase went from design to live:

- The site was rebuilt honestly: a deduplicated directory of 2,109 verified practices (280
  duplicates and a long tail of junk entries — Yelp pages, mirror sites, a US clinic —
  removed), three guides rewritten against the final report with every figure sourced and a
  review date on each, and claim-your-listing and opt-out routes.
- The data platform was unified on the chassis, generically: one price schema for services and
  medicines, the CMA's 36 services seeded as the canonical taxonomy every price maps onto, and
  a config-driven exporter that publishes the directory, k-anonymous area statistics, and
  consented or attributed price lists for any vertical and domain — fail-closed at every point
  where the old system had a loaded default.
- The claim flow — the commercial core — was built and proven end to end: a practice claims its
  listing, we verify they are who they say, record the exact consent wording, and their own
  figures publish attributed to them; opting out hides prices but keeps the practice findable;
  a claim reverses an opt-out. Every step leaves an audit trail.
- The site was adopted onto the chassis, which then built it autonomously overnight — strategy,
  design, imagery, new pages — and the exporter made its first fully autonomous publish: five
  data files, committed and deployed by the platform with no human in the loop, every
  publication rule holding.

One finding mattered more than expected: our historical price data carries no per-price source
URLs, so under our own rule it can never be shown against a named practice — it feeds anonymous
area statistics only. Per-practice prices start from fresh collection that records provenance
as it goes. Real coverage was always thinner than the old site pretended, and we now say so.

## Where we are now

The site is live, honest, and platform-managed: working directory, sourced guides, published
area statistics, and an exporter refreshing the data every two days. The pipeline from database
to public page runs without us. Nothing on the site quotes a figure we cannot source — the only
prices anywhere are the CMA's own.

The autonomous build also cost us something: the rebuilt homepage dropped the directory search
box and the claim/opt-out section. The data and machinery behind both are untouched — it is a
page-markup restoration, first job for the next session, and a handful of build items are
waiting for human review in the admin queue.

## Where we're going

Near term: restore the homepage function; submit the funding-consultation response before
30 July (drafted — it argues the levy should scale by practice size rather than hit a
single-site independent as hard as a corporate branch); and respond to the consultation on the
substantive Order the moment the CMA publishes the draft, arguing for low-barrier third-party
approval and an express right to reuse the mandated price lists.

Then the last build phase: provenance-first price collection — re-verifying the practices with
bad website records and scraping practices' own price pages with the source recorded per price,
which is what fills the attributed-prices file the exporter already publishes. From December,
when large groups must publish the standard list one click from their homepage, collection
becomes dramatically easier — and showing which practices have and haven't published becomes
both a public service and our strongest argument for a practice to claim its listing.

The strategy in one sentence: the CMA is making every vet publish comparable prices; we intend
to be where that comparison is useful — the directory and statistics carry the site today,
claimed listings are the product we offer practices as their compliance shop window, attributed
collection fills the coverage, and we never publish a number we cannot show the source for.
