# vetcomparison.uk — the bugs we hit along the way

Written 18 July 2026. A read-out account of what went wrong during the rebuild, in the order it
mattered. Some were ours, some were the platform's, and one of them came back a second time by
itself — which turned out to be the most important thing we learned.

## 1. The founding defect: invented prices published as fact

The site had been live since February publishing consultation and prescription fees for about
3,100 named real veterinary practices. None of those figures came from anywhere. The tell was
statistical: across 3,124 practices there were only 46 distinct consultation prices, and 692 of
them — better than one in five — were charging exactly £48. Real prices do not distribute like
that. Alongside them sat a notice claiming the data was our proprietary property, a quotation
attributed to the Competition and Markets Authority that the CMA never wrote, and an assertion
that a named practice in Glasgow charged £33 to write a prescription, with nothing behind it.
The same fabricated data had also been imported into the production database, where 997 invented
price rows sat against 235 genuinely verified practices.

## 2. An exporter pointed at a domain we do not own

The code that publishes price data to a website had `vetcomparison.co.uk` hardcoded as its
default target — a domain the business does not own. The same wrong value was seeded in two
places in the database as well. Nothing had fired, because the task was disabled, but it was a
loaded gun: enabling it would have committed our data to somebody else's address.

## 3. An action that was never registered

While fixing the above we found the medicine price exporter had never been registered in the
platform's action registry at all. The agent existed, the workflow existed, the scheduled task
existed — but the thing they all pointed at had no entry, so it could never have run. It had
been sitting there unable to work, silently, for months.

## 4. The directory was full of things that were not veterinary practices

Twenty entries marked "verified practice" turned out to be Yelp pages, StarOfService and
BestLocalRated listing pages, one American clinic and one college course page. A further 177 were
real practices whose recorded "website" was actually a wheree.com mirror of their site rather
than the practice's own. Seventeen more had a scraped page title where their name should be —
one read "26 Vets in Birmingham - Compare Prices & Reviews".

## 5. Two hundred and eighty duplicates from a missing normalisation

The directory listed the same practice repeatedly because different discovery routes had recorded
`croydonvets.co.uk`, `www.croydonvets.co.uk` and `croydonvets.co.uk/` as three different
businesses. One practice appeared five times. The root cause is a single missing step: the
website address is never normalised before a practice is stored, so trivial URL variants create
new records.

## 6. Prices with no provenance — and a check of mine that missed it

Our publication rule is that no price appears without a source. When the new exporter was first
tested it produced an empty per-practice price file, which looked like a bug and was not: every
one of our 803 historical price rows has an empty source URL. The provenance was never recorded
when the data was collected, and it cannot be reconstructed. Worse, I had earlier reported that
all 803 rows *had* sources — because I tested whether the field was null rather than whether it
was empty, and an empty string passed. The consequence is permanent: that data can only ever feed
anonymous area averages, never a figure against a named practice.

## 7. Adoption blocked by a database index that did not exist

Registering the site with the platform failed at the final step with a cryptic error. The insert
depends on a partial unique index whose definition must match the code's expectations exactly,
and that index was missing from the live database — it belonged to a change another workstream
had written the same day but not yet applied. The fix was not ours to make, so we waited rather
than hand-creating an index that might have mismatched from the other direction. It landed with
the next deploy and adoption then completed first time.

## 8. Scripts and clones that fought back

Several smaller ones worth knowing. A shell safety setting that should have aborted a script
after a failed check did not, and a push went out before its verification gate — no harm done,
but the lesson stands: verify and publish must be separate steps. The local copy of the website
repository turned out to be some seventeen hundred commits out of date and dirty with other
people's work, so every publish had to be done through an isolated checkout, and automated
processes pushing to the same branch repeatedly beat us to it mid-publish. And because several
Claude sessions work this codebase at once, our own files were repeatedly swept into other
threads' commits.

## 9. The one that matters: the platform re-created the fabrication by itself

This is the important one. Having cleaned the site, we registered it with the chassis so the
platform could manage it. Overnight the platform rebuilt the site autonomously — and its
tool-recreation agent, asked to recreate our practice search, wrote a synthetic data generator.
Its own code comment says so plainly: *"The original directory holds 2,100+ UK practices. For
this recreation we generate a large, realistic, deterministic dataset."* It assembled fake
practice names from a list of fragments — Abbey, Oakwood, Willow crossed with Veterinary Centre,
Vets, Animal Hospital — and invented postcodes with a seeded random number generator.

So four days after we removed fabricated veterinary data from this site and wrote a legal record
about it, the platform put fabricated veterinary data back on it, unprompted.

The same rebuild also introduced claims the site cannot support: it advertised "pricing
information" and "ownership data" we deliberately do not publish, offered "Price: Low to High"
sorting for prices we do not have, and described our real 2,109-practice directory as "a
representative sample for demonstration purposes".

Every work item behind that rebuild reported `complete`.

## 10. And the reason we nearly missed it

I checked the rebuilt site by looking for whether the right elements were present, saw the search
box missing from the page markup, and reported that the rebuild had "dropped" the directory. That
was wrong twice over: the search component was there and was well built, and the real problem —
that it was populated with invented practices — was invisible to the check I ran. It surfaced only
because the project's own coding standards say, in as many words, *trust the rendered artefact,
not the status*, and re-reading them prompted me to look at what the page actually said rather
than what it contained.

## What the pattern is

Nearly every one of these is the same shape: a system reporting success while the artefact it
produced was wrong. Prices that looked like prices. A registry entry that was missing rather than
broken. A source field that was present but empty. Work items marked complete over a page showing
invented data. None of them announced themselves; all of them were found by looking at the actual
output rather than the status of the process that made it.

That is why the site's publication rule is what it is — no figure appears unless we can show where
it came from and when we saw it. It is not only a legal position. It is the only check that fails
loudly when something upstream quietly starts making things up.
