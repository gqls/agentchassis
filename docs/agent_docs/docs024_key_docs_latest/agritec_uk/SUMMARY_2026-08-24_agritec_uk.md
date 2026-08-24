# agritec.uk — where the rebuild stands, 24 August 2026

Written to be read aloud. First summary for this workstream: the earlier days were groundwork and
would have read the same as each other, which is the test for whether a milestone has happened.

## What we're trying to do

Take agritec.uk, which is a hand-built site of calculators and technical articles, and rebuild it
inside the framework so that everything on it is generated, checked and maintainable rather than
typed in by hand once and left. The point is not that the old site looks bad. The point is that a
hand-built site opts out of every control we have — nothing checks its numbers, nothing notices
when a page stops being linked, nothing tells anyone when a figure goes out of date. And all three
of those had happened to it.

The owner asked for the same subjects covered afresh, to the same level of detail and greater,
with more imagery and diagrams, every figure sourced and linked so a reader can check it, and the
site hosted back on its own domain in place of what's there now.

## Where we've come from

The first thing we found was that the copy of the site in our own repository was badly out of
date — six calculators where the live site has thirteen, and missing an entire seven-part
engineering series. Planning from it would have silently dropped more than half the site.

The second was that the owner's recollection about articles not being linked was right, and worse
than remembered. One guide was reachable from no index page at all. Three calculators were on the
home page but missing from the tools index. The sitemap was a third inventory that agreed with
neither of the others: seven of its entries were dead links, and the entire deep-dive series was
absent from it, so that work was invisible to search engines.

The third was the market ticker on the homepage — wheat, oil, carbon and fertiliser prices. It
was fabricated. The small program written to feed it generates the numbers with a random number
generator and labels its own output "Simulated Exchange".

Then, sourcing the figures properly, we found the one that actually costs money. **The SFI
calculator pays a subsidy that no longer exists.** It leads with a promise of £20 per hectare on
your first fifty hectares, describes the first £1,000 as effectively guaranteed, and sits next to
a link to the official guidance that makes the whole thing look checked. The government abolished
that payment for SFI26. When we went further and pulled the real scheme rules, it turned out only
two of the calculator's nine revenue lines were correct: four of its actions no longer exist at
all, and of the two survivors that had moved, herbal leys is £224 a hectare rather than £382 — an
overstatement of £158 a hectare, close to eight thousand pounds on a fifty-hectare block.

## What we've done

The site now exists inside the framework. It has an evidence register holding **105 sourced
facts** — the whole SFI26 action set, UK non-domestic electricity prices from the Office for
National Statistics, crop light requirements from Virginia Cooperative Extension — every one with
a link a reader can click and a date we captured it.

Nine further facts were registered and then removed on review, which is worth saying plainly
because it is the part that makes the rest trustworthy. Four were domestic energy price-cap
figures: true, properly cited, and about households rather than commercial growers, so wrong for
every reader this site has. One asserted a single electricity price that was actually one cell of
a two-column table. One was nine years old and said "now".

Eleven pages are built, exactly the ones we specified and nothing invented. Six of them are the
technical explainers, and they came out at **1,400 to 1,803 words against the old site's 315 to
453** — roughly four times the depth. That was not luck. We measured what the framework produces
at each page type first, found that one setting decides whether an article lands near 1,600 words
or near 500, and chose accordingly. The obvious choice would have produced the old depth while
looking correct.

Seventeen images have been generated in the site's own visual language — cutaway diagrams of
growing chambers, plan views of a farm holding divided into parcels — rather than the stock
photography this subject usually attracts.

## Where we are now

Two things went wrong that are worth knowing about, because both were self-inflicted and both were
caught.

The guides index rendered with no links on it at all. The cause was our own depth decision: the
component that lists articles looks for one page type, and we had chosen the other one for the
word count. Both halves of that decision were right on their own and they fought each other.
Fixed, after a first attempt that the pipeline quietly undid because we had edited a copy of the
setting rather than the setting itself.

Then two explainers refused to build, blocked by our own anti-fabrication rules. One rule, written
to catch invented commodity prices, was matching the phrase "Carbon Brief, May 2025" — a citation.
The other was blocking a properly sourced past-tense sentence about the abolished payment, which
is exactly what the site should say. Both are narrowed and both were re-tested in both directions
this time. Notably, each failure showed up as a page refusing to build rather than a sentence
quietly disappearing, which is the failure mode you want.

All six explainers are now built and deployed. Every one is reachable. The guides index is
re-rendering to pick up the last two.

## Where we're going

Next is the calculators themselves — the SFI one first, since its specification is written and its
figures are registered, including the caps the old tool models none of: a £100,000 annual ceiling,
a three-hectare minimum, and a limit on how much of a holding certain actions can cover.

After that, the machine-vision and edge-computing half of the existing site, then the additions
the owner asked for: news, editorial features, and a supplier directory — the last of which is
genuinely new work rather than a switch to flip.

One decision is waiting. The site has ended up with two colour schemes: the existing agritec brand,
which all seventeen images use, and a darker scheme the framework designed independently, which
the site's stylesheet will use. One of them has to give way, and which is a question about the
brand rather than about the software.

And one older question still stands: the old calculator is still live, still promising farmers
money the scheme no longer pays.
