● That's an important constraint, and it's worth separating two things that are easy to run together, because they have different answers.

Lawful collection and lawful republication are different tests. Your scrape service solving the first doesn't settle the second. Getting a price off a public page legally doesn't by itself give you the
right to publish it as your dataset.

What's actually protected. A price is a fact, and facts aren't copyrightable — "Ardene House charges £66 for a consultation" is not someone's creative expression. Two things are live risks, and they're
the ones to design around:

- UK database right (Copyright and Rights in Databases Regulations 1997) — this is the real exposure. It protects substantial investment in compiling a database, and extracting/reutilising a substantial
  part infringes even though the individual facts are free. Systematically republishing an online pharmacy's whole price list is squarely in this territory.
- Verbatim text — service descriptions, blurbs, anything expressive. Copying those is straightforward copyright infringement. Prices are not; prose is.

Site terms of use are a third thing again — contract, not copyright — but they bite on the scraping side.

The architecture that dissolves most of this. Don't publish a rival dataset; publish facts you have permission for, and point at everything else:

Store the scrape as evidence, internal, which is what the existing evidence store already does. Publish per-practice figures where you have a licence — and the licence comes from the vet claiming their
listing. Publish aggregates and derived statistics for everything else (median consult fee in BS7, how a practice sits against its area) — aggregates are both far safer and, for a pet owner, more useful
than a scraped table. And deep-link to the vet's own price page rather than reproducing it, with a timestamp and source attribution.

The part worth dwelling on: this is your business model, not a workaround. The CMA gives vets a legal reason to publish prices somewhere; you give them the place. A vet who claims their listing to comply
is, in the same act, granting you the right to publish their prices. Consent converts the copyright problem into a licensing problem, and it's the same mechanism as "we want them to use us." The scrape
then does what it should: seed the directory, detect who isn't compliant, and give a claimed listing something to correct rather than start from blank.

I'd want a solicitor to sign off the database-right position before launch — I can structure it, but I shouldn't be the last word on it.

Two things to close the loop:

Which data did you mean by "can't publish directly" — the medicine prices scraped from the online pharmacies (Pet Drugs Online, Animed, Hyperdrug), the vet service prices scraped from practice sites, or
both? They carry different risk: the pharmacy catalogues are compiled databases and the riskier of the two; a vet's own consultation fee, which the CMA is compelling them to publish anyway, is much weaker
ground for them to object on.

And I still need the read-only count I was cut off on — of the 2,767 verified practices, how many have real prices? That decides whether the launch product leads with comparison or with the
directory-plus-claim flow.
