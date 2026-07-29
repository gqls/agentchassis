# SUMMARY — dartsonline.com, 2026-07-29: the first guides exist, and the site stopped lying

Written to be read aloud. First summary in this lane.

## What we're trying to do

Turn dartsonline.com into a leading site for darts players — steady fresh news, real
analysis, plenty of imagery, and a good number of interactive tools — so that when the
owner applies to an affiliate network the site is already worth accepting, and the day he
is accepted the product feed is a config switch rather than a project. Everything built
along the way has to land in the framework generically, so the next domain — boxing,
football, whatever — inherits it rather than repeating it.

## Where we've come from

dartsonline was the first site this platform ever built end-to-end by itself, back in
early July. It has been a standing embarrassment ever since: it deploys, and it is bare.
A quality programme was opened for it on 6 July, wrote down seven things that were wrong,
and stopped. News, feeds and tools were explicitly ruled "never in scope" for it.

Three weeks later the position was worse than that document knew. Only seven of
twenty-four planned pages were live. Three of the five links in the site's own header
pointed at pages that had never been built. The nine buying guides — tungsten
percentages, barrel weights, flight shapes, precisely what a darts buyer searches for —
existed as titles and nothing else, because nobody had ever recorded what sections a
guide page should contain; five attempts to build them had died with nothing to write.
Thirty-three images had been generated, paid for and deployed while the home page showed
emoji. And the site told visitors it stocked a full range of darts from seven named
manufacturers, from an address in Portland, Oregon, using a phone number and email
belonging to a different company entirely — all of it inherited from research that had
honestly flagged it as unconfirmed, and shipped as fact anyway.

## What we've done

**Stopped the site making false claims.** The fabrication had three homes, not one, and
that is the finding worth carrying to other sites. The first two were obvious: the
research identity, and the briefing the About page is written from. Both were rewritten
to say only what is true of a UK-based, online-only darts publication that holds no
stock, with the owner's real contact details, plus an explicit list of claims no future
page may make. The third was found only by building a page and reading it: the site's
**writing guide** still instructed the writer to produce shop copy — "product listings",
"Add to Bag", "show savings", "brand pages … not just list SKUs". So the very first guide
we built came back well-written, accurate, on-voice, and ended by inviting the reader to
"filter our ranges". Nothing to filter. That guide has been rewritten, and so has the
direction behind it: the voice is untouched, the shop assumptions are gone, and rules for
news and analysis exist for the first time.

Two more false pages surfaced the same way. The About page called us a "Specialist Darts
Retailer" in its title tag; the Shipping & Returns page promised couriers, 2–4 day UK
delivery, express options at checkout, a daily order cut-off and 30-day returns, on a
site with no checkout. Both are rewritten. About now says, in its own words, "We don't
sell darts, hold stock or ship products… an independent guide" — and turns that
independence into the reason to trust the advice.

**Unblocked the content engine.** The nine guides now have a section layout, taken from
the platform's own canonical article default rather than invented. Two are built and
live; barrel-weight is the first guide page this site has ever produced, and it is good —
21g to 26g for steel tip, the 22–24g sweet spot, what a heavier barrel does to your
release. The engine always worked; it had never been given anything to work with.

**Armed the news lane.** dartsonline now qualifies for the aggregation pipeline six other
sites already run. Seven candidate darts feeds were fetched and checked rather than
trusted: Darts World and the official PDC feed are live and posting daily, a Google News
darts search gives a wider net, and four failed — including a Sky Sports feed that
returns a perfectly healthy twenty items and is their general news feed with almost no
darts in it.

**Fixed the navigation** by finding out it wasn't a code problem. Three dead header links
turned out to be orphan rows from superseded plans, and the platform already prunes such
links correctly — dartsonline was simply serving a stale copy of its own header. A fleet
query confirmed no other site is affected.

**Built the owner's tool ratio into the framework.** The mechanism that decides when a
site needs another tool worked purely on a calendar. It now understands "one tool per six
published articles", off by default for every other site, with tests that pin the
off-case so absence can never start doing work. It is committed and with the review
council.

## Where we are now

The site is honest, its guides can be built, and its news feed is armed but not yet
flowing. Two guides are live, two more page rewrites are draining through the build
queue. Nothing about the site's imagery, its tools, or its affiliate plumbing has been
built yet — those were the later phases and they remain untouched.

Two planned fixes were deliberately **not** made, and both were cancelled by measurement
rather than by opinion. A fix to route stranded image items was dropped because that
queue's own bug file has already established that routing is not the bottleneck and
carries an explicit "decision pending, do not act". A fix for the 404 navigation was
dropped because the fleet query showed one affected site and an already-correct
mechanism. Both are recorded where the next person will look.

## Where we're going

In order: build the remaining seven guides now that the direction behind them is honest;
add the verified RSS feeds once the pipeline's first run has seeded its search sources
(inserting them early would stop those sources being created at all); create the news
page and put the news section on the home page; then imagery — the site has a full set of
generated assets it barely references, and the home page still shows emoji; then the
setup-builder tool, which has been planned since July and never made, as the first of the
tools the ratio now asks for; and finally the affiliate resolver and feed ingester, built
dark and switched on the day a network says yes.

The honesty work is the part that should not be rushed past. An affiliate network looks
at the site before it accepts you, and until this morning the site described a business
that does not exist.
