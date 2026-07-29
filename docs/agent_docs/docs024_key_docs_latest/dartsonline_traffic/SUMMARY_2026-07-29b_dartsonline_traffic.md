# SUMMARY — dartsonline.com, 2026-07-29b: the guides are out, the news page is live, and two of yesterday's claims were wrong

Written to be read aloud. Second summary in this lane; the first was written this
morning and is superseded on two specific points, both named below.

## What we're trying to do

Turn dartsonline.com into a leading site for darts players — steady fresh news, real
analysis, plenty of imagery, and a good number of interactive tools — so that when the
owner applies to an affiliate network the site is already worth accepting, and the day he
is accepted the product feed is a config switch rather than a project. Everything built
along the way lands in the framework generically, so the next domain — boxing, football,
whatever — inherits it rather than repeating it.

## Where we've come from

This morning's session found that the site had been describing a business that does not
exist: another firm's Portland address, phone and email, claimed stock relationships with
seven manufacturers, and a Shipping & Returns page promising couriers and 30-day returns
on a site with no checkout. Three homes of that fabrication were fixed — the research
identity, the About-page briefing, and the writing guide the content writer actually
reads. Nine buying guides that had existed as titles since 6 July were given section
layouts and two of them were built. The news pipeline was armed with nine sources.

That session closed by reporting all nine built pages clean of shop language, and warning
that the broken navigation links would persist until the site chrome was rebuilt.

**Both of those statements were wrong, and finding out how is the most useful thing in
this summary.**

The clean sweep searched every page for a fixed list of phrases — the ones removed from
the other pages that morning. `clearance`, `cut prices` and `sale range` were not on the
list, because they had not been seen yet, so `/sale.html` printed `clean` while serving
"We cut prices across our sale range" and "We move high-density tungsten barrels, shafts,
and flights into clearance regularly". The check was correct; the sentence written about
it was not. **A word search can only confirm the absence of what you thought of.**

The nav warning was wrong in a way that would have cost real work. The dead links were on
four pages, not all of them, and those four were exactly the ones not rebuilt since the
nav data was corrected — the header is regenerated per page at build time, so everything
rebuilt that morning came out clean without anybody touching chrome. The fix that had been
queued, a chrome rebuild, would have rebuilt the navigation table from stale rows and
produced a header still missing the Guides link, forcing every page to be built twice.

Both are logged in `WRONG_CALLS.md` with the ten-second check that would have caught each.

## What we've done

**Found the false premise's fourth, fifth and sixth homes.** Reading served page titles
rather than page bodies turned up `site_plan_pages` — the record the system rebuilds pages
from — still describing us as "Specialist Darts Equipment & Accessories" and the About
page as a "Specialist Darts Retailer". Six surfaces had been corrected and the one that
*regenerates* them had not, so the lie would have returned at the next reconcile. The
fifth was the imagery plan, whose prompts ask for "a delivery truck with a checkmark" and
"a curated product range". The sixth is the stylesheet: the site is built on an
`ecommerce-storefront` layout whose white card colour carries the comment "product cards
stay neutral regardless of palette — product images demand a clean backdrop."

**Released the guides.** Eight of nine are live: tungsten percentages, barrel weight,
board setup, brand comparison, flight shapes, shaft length, steel versus soft tip, and the
beginners guide. The ninth belongs to another workstream and was left alone. The board
setup page was read closely because it is the one readers will check hardest, and its
measurements are right — 1.73m to the bullseye, 2.37m throwing distance, and it correctly
says to measure from the board face rather than the wall.

**The news page exists and is live**, at `/news/index.html`, listing sixteen real darts
stories: Luke Littler's World Matchplay defence, Henry Coates' first ProTour title, the
updated Order of Merit. Its hero says, in its own words, "We don't hold stock or sell
anything."

**Repurposed the two shop pages rather than deleting them**, because on this platform
archiving a page does not take it off the web — an archived sale page would go on serving
"we cut prices" to every visitor and crawler indefinitely. `/sale.html` is now "How to
Spot a Genuine Darts Deal" and `/new-arrivals.html` is "New to Darts? Start Here". Both
keep their addresses.

**Made the site readable.** Measured in a real browser, the homepage had thirteen contrast
failures and a guide page five: the entire front-page card block was near-white text on
white boxes, at 1.12:1. The fix already existed in the platform, shipped two days ago for
another site, and was live in both server replicas — dartsonline was simply carrying a
stylesheet older than the fix. Regenerating it took under three minutes and took the site
from eighteen failures to one.

**Descriptions for every page that will exist.** Eighteen of twenty-one had none, which
meant publishing an empty one — worse than none.

**Your tool ratio was approved**, nine minutes after resubmission. Round one had found a
real mistake of mine and the second submission answered all three objections with
measurements rather than argument.

## Where we are now

The site is honest, readable, navigable and has eight buying guides plus a live news page.
The menu reads Guides, News, Start Here, Deals, plus About, Contact and Shipping &
Returns, and no page serves a broken link.

Three things are in flight rather than finished, and none should be reported as done. A
batch of about twenty-five page reassemblies is draining, which is what puts News into the
menu of the pages built before it existed. The Guides hub currently links to no guides —
its listing was written on 20 July, before any existed, and every rebuild since preserved
it, which is reassembly working exactly as designed and the reason a stale listing is
invisible in every status field. A regeneration is queued for it and for the homepage.
And four images are generating as a deliberately small probe: seventeen icons were
generated once before, marked active, and are all unusable because their prompts asked for
grey lines on a near-white ground and nobody looked. Every new one gets looked at before
it goes near a page.

Nothing has been built yet on tools, affiliate plumbing, or search-discovery files.

## Where we're going

In order: let the reassembly batch land and confirm by fetching pages rather than querying
the database; regenerate the two listings so the Guides hub and the homepage link to the
eight guides, which is the internal linking that matters most for search; look at the four
probe images and, if they are usable, generate the remaining eleven — eight of which are
one hero photograph per guide, which is most of the imagery the site needs; then the
discovery files (robots, sitemap, llms.txt) once the site stops changing under the probe;
then the setup-builder tool, which has been planned since July; and finally the affiliate
resolver and feed ingester, built dark and switched on the day a network says yes.

The thread running through today is worth stating plainly, because it will apply to the
next site as much as this one: **every claim that turned out wrong was one made about an
artefact by looking at something other than the artefact.** The sweep read stored HTML
instead of the page. The nav diagnosis read a database column instead of fetching the
site. The stale listing looks perfect in every status field it has. The fix is not more
care — it is fetching the thing you are about to make a claim about.
