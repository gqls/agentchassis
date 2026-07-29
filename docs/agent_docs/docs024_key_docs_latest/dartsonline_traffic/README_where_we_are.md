# Where we are — dartsonline.com

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-29 — opening the lane

You asked for dartsonline to become a leading site for darts players: fresh news,
analysis, lots of imagery, tool-heavy, and ready to take an affiliate feed when you
apply — with everything built so it works for the next domain too, whether that's boxing
or football.

Before planning anything I had the site and the platform surveyed properly. The short
version of what we found is that dartsonline is in worse shape than it looks from a
distance, and in better shape than it looks once you know where the machinery already is.

**What's actually wrong with the site right now.** Three of the five links in the site's
own header go to pages that do not exist — Shop, Brands and Guides all 404. Of the 24
pages that were planned, only 7 are live. The nine buying guides — tungsten percentages,
barrel weights, flight shapes, the exact things a darts buyer searches for — exist as
titles and nothing else; nobody ever decided what sections they should contain, so they
could never be built. There is nothing to buy anywhere on the site. And 33 images have
been generated, paid for and deployed, while the home page shows emoji instead of them.

**The part that mattered most.** The About page tells visitors we stock a full range of
darts and carry the major manufacturers. We don't — there is no stock and no brand
relationship. Worse, the research specs behind the site carried a US company's address in
Portland, that company's phone number and email, and a link to an Australian darts firm's
Facebook page. None of it is ours. The original research had actually flagged all of this
as unconfirmed back in July; the site got built from it anyway. Since an affiliate network
looks at your site before accepting you, this was both a straightforward honesty problem
and the thing most likely to get an application rejected, so it was the first thing fixed.

**What I've done so far.** The identity and briefing specs are rewritten to say only true
things: a UK-based, online-only specialist darts site that publishes spec-first buying
guides and holds no stock, with your real contact details. I've added an explicit list of
things no future page is allowed to claim — no stocking, no brand partnerships, no
address, no shipping promises — so the next rebuild can't quietly reintroduce them. The
nav is repointed at the real hub pages instead of the three dead links, and thirteen
navigation labels that had entire page titles stuffed into them are trimmed. The old
records are all backed up, and every change is a new version rather than an overwrite, so
nothing is lost.

**Two things I planned to build and then didn't, which I think is the more useful news.**

The first was a fix to make generated images actually reach the site. It looked like a
two-line bug. But the queue those items sit in has a bug file of its own, and its owner
has already worked out that the real problem isn't routing at all — there are 325 items
sitting in the queue humans *can* see, the oldest from March, unread. He's put a decision
to you and written "do not act until it's recorded". So I've left it alone and will drain
dartsonline's own items directly instead, which doesn't touch anything shared.

The second was a fix for the 404 navigation, which looked like a platform-wide defect
worth fixing properly. Before writing it I checked how many other sites had the same
problem. The answer was none — dartsonline is the only one, and the platform already
fixes this correctly; our site is just serving an old copy of its own header. So that fix
wrote itself out of existence, which is the cheapest kind of fix there is.

**On your tool-heavy request** — roughly one tool per six articles — there's already a
mechanism that decides when a site needs another tool. It currently works on a timer
(check every 7 days if you have no tools, every 30 if you have some) rather than on how
much content the site has. Teaching it your ratio is a small, genuinely generic change
that would apply to every site in the estate, which is exactly the shape you asked for.

**What's next**, in order: get the guides buildable and built (that's the traffic), turn
on the news feed (the machinery is live on six other sites already and dartsonline is one
config row away from joining), then imagery, then the affiliate plumbing built dark so
that on the day you're accepted it's a config switch rather than a project.

---

## 2026-07-29, later — the first guides are built, and page one caught a lie the specs didn't

Since the last entry the site has actually started moving, and one thing happened
that's worth telling properly because it changes how I'd approach the rest.

**The nine buying guides can now be built, and two of them are live.** The blocker was
mundane once found: nobody had ever recorded what sections a guide page should contain,
so the writer opened each one, found no structure, and gave up — five times, over three
weeks. I gave all nine the platform's own standard article layout (hero, article body,
call to action — I took the list from the code that creates blog posts rather than
inventing one) and released just two for building. **barrel-weight is the first guide
page this site has ever produced**, and beginners followed it.

The writing is genuinely good. It talks about 21g to 26g for steel tip, the 22–24g sweet
spot, what a heavier barrel does to your throw — specific, in the site's voice, no filler.
That's the thing worth knowing: the content engine works, it was just never given
anything to work with.

**Then I read the page, and the last paragraph said "Filter our ranges by weight and
tungsten percentage."** We have no ranges. Nothing to filter.

This is the part I'd want you to take away. This morning I rewrote the two places that
held the fake identity — the research record and the About page's source. I thought that
was the honesty problem solved. It wasn't. There was a **third** place: the site's writing
guide, which tells the writer how every page should sound. It says to write product
listings, to end pages with "Add to Bag", to show savings on prices, to make brand pages
that don't just list SKUs. Every one of those instructions assumes a shop. The writer
wasn't making things up — it was following orders none of us had thought to read.

So the same false premise lived in three places, and the one I'd missed was the one with
the most reach, because it shapes every page the site will ever write. I've now rewritten
it: the voice stays exactly as it was (that part was always good), but the shop
instructions are gone, replaced with rules for a publication, plus explicit rules for news
and analysis, plus a plain list of things no page may ever claim. Both guides are being
rebuilt against it.

**The lesson, and it's cheap to state and expensive to learn:** I only found this because
I built one page and read it. No amount of looking at the configuration would have caught
it, because each piece is perfectly sensible on its own — it's only wrong when measured
against a business that doesn't exist any more. Releasing two guides instead of all nine
is what kept it to two pages instead of nine.

**Two more false pages found the same way.** The About page still claimed we stock and
carry; its title tag literally called us a "Specialist Darts Retailer". Worse, the
Shipping & Returns page promises dispatch, tracking, couriers, 2–4 day UK delivery,
express options at checkout, a daily order cut-off and 30-day returns. There is no
checkout. If someone had relied on that page there'd be a real problem, and it's exactly
the page an affiliate network would check. Both are being rewritten now — Shipping &
Returns keeps its address on the web but becomes an honest explanation that you buy from
the retailer we link to and their terms apply, which is a genuinely useful page for the
site we're actually building.

**The news feed is armed.** The site now qualifies for the pipeline that six other sites
already run. I checked seven candidate darts feeds by actually fetching them: Darts World
and the official PDC feed are good and posting daily, a Google News darts search gives a
wider net, and four were dead or wrong — including one Sky Sports feed that returned a
perfectly healthy-looking 20 items and turns out to be their general news feed with
almost no darts in it. There's a deliberate wait before I add those feeds: the seeding
code skips everything if any source already exists, so adding mine too early would stop
the automatic search sources being created at all.

**On your tool-heavy request**, that's now built and committed. The platform already had
a mechanism deciding when a site needs another tool, but it worked purely on a calendar —
every 30 days, whether the site had published two guides or thirty. It now understands
your ratio: one tool per six published articles. It's off by default for every other site
(I checked: 13 of the 14 have no setting at all, so nothing changes for them), and it's
gone to the review council. Nothing about dartsonline's own tools is built yet — that's
next, starting with the setup-builder that was planned long ago and never made.

---

## Session 2, same afternoon — two things I told you yesterday were wrong

I want to lead with the corrections rather than bury them, because one of them was in the
handover note I wrote for whoever picked this up next.

**I said all nine built pages were clean of shop language. They weren't.** The sale page
was, and still is as I write this, telling visitors "We cut prices across our sale range"
and "We move high-density tungsten barrels, shafts, and flights into clearance
regularly". We don't cut prices. We don't move anything. There is no sale.

The way I got it wrong is worth a sentence, because it isn't carelessness exactly and it
could happen again. My check searched every page for a list of specific phrases — the
ones I'd just spent the morning removing from other pages. "Clearance" and "cut prices"
weren't on that list, because I hadn't seen them yet. So the check came back clean and I
reported it as "the pages are clean", which is a much bigger claim than the one I'd
actually tested. A word search can only ever tell you that the words you thought of
aren't there. I found the real state of the page by opening it in the way a visitor
would, an hour later, for an unrelated reason.

**I also said the broken navigation links would stay on the site until we rebuilt the
site header.** That was wrong too, and the fix I had lined up would have made more work
rather than less. The dead links are on four pages, not all of them, and they're on
exactly the four we haven't rebuilt since fixing the navigation data. The header is
regenerated every time a page is built, so every page rebuilt yesterday came out right
without anyone touching it. What was actually stale was the navigation table itself — it
still listed three deleted pages and had never heard of the Guides section. So I rebuilt
that first, on its own, checked it, and only then started rebuilding pages. Had I done it
the other way round, seven pages would have been built with a menu missing the Guides
link and would all have needed doing again.

**And there was a fourth place the old lie was living.** I found it by looking at the
page titles rather than the page text. The site's *plan* — the record the system rebuilds
pages from when anything gets regenerated — still described us as "Specialist Darts
Equipment & Accessories" and the About page as a "Specialist Darts Retailer". I'd fixed
four things that read from that plan and left the plan itself alone, so the next time
anything regenerated, the lie would have quietly come back. That's fixed now.

**What I've done with the two shop pages.** The obvious move was to delete them. I
didn't, and the reason matters: on this platform, archiving a page doesn't take it off
the web. An archived sale page would go on serving "we cut prices" to every visitor and
every search engine indefinitely. So I've repurposed both, keeping their web addresses:

- The sale page becomes **"How to Spot a Genuine Darts Deal"** — what a real discount
  looks like, why a cheap set can still be the wrong set, what to check before buying.
  It's honest without needing a product feed, it's something people actually search for,
  and it's the natural home for affiliate links the day we have them.
- The new-arrivals page becomes **"New to Darts? Start Here"** — the decisions a beginner
  faces, in the order they face them, each pointing at the guide that answers it.

The second one has already rebuilt and reads well. It's a bit thin — it can't list links
to the individual guides because that page shape has nowhere to put a link list — so
that's half-finished rather than done, and I've written it down as such.

**The menu now says what the site is.** Guides, Start Here, Deals, plus About, Contact
and Shipping & Returns. Shop and Brands are out of it: they're retail sections with
nothing to put in them until an affiliate feed exists.

**Descriptions.** Eighteen of twenty-one pages had no description for search results at
all — and the way the system works, that means it was publishing an empty one, which is
worse than none. I've written them for every page that will exist. Two of the three that
already had one were mine from yesterday, and both were too long for Google to show in
full; I'd written the rule about length and then not measured my own.

**Your tool ratio passed review.** The council approved it, nine minutes after
submission. Round one had come back asking for changes, and one of the three points was
a real mistake of mine: I'd claimed something about the code structure without checking
it. This time I checked it and attached the numbers. Three reviewers then made the same
new suggestion — that the setting should be written down where the next person will look
for it, not only where it's used — which I've done.

**Building now:** the sale page, the contact page, and four of the seven remaining buying
guides. Plus, for the first time, a proper **news page** — the feed has fourteen relevant
darts stories waiting and nowhere to show them.

---

## Later the same afternoon — the site can now be read

**The homepage was unreadable and nobody had noticed.** I measured it in a real
browser rather than reading the stylesheet, and the whole block of cards on the front
page — the four "what this site does" panels — was near-white text on white boxes, on a
site whose background is nearly black. Thirteen separate failures on that page, five
more on a guide page. It has presumably looked like that since the site was built.

The cause turned out to be something worth explaining, because it's a trap we'll hit
again. The site is built on a **shop layout**, and that layout has white as its
built-in colour for cards — with a comment in the code saying, in effect, "product
cards stay white because product photos need a clean backdrop". Our dark colour scheme
overrode the page background but never overrode the card colour, because the card
colour isn't one of the eight colours a site's palette defines. So the dark scheme and
the shop layout's white cards coexisted, each internally consistent, producing white
text on white.

**Somebody had already fixed this two days ago** — for a different site, and properly,
in the code. But a code fix can't reach into a stylesheet that was written once and has
been sitting on the CDN ever since. So dartsonline was carrying a stale file. I
regenerated it, which took under three minutes, and the page went from thirteen
failures to one. I've written that lesson up in the debugging guide, because it isn't
about colours: **when we fix something that generates files, only future files get
fixed, and nobody ever counts the ones already out there.**

The one remaining failure I've deliberately left alone, and I want to be clear why.
The layout uses a single colour setting for two contradictory jobs — as the background
of buttons (with pale text on top) and as the text colour of small labels (on the dark
page). I worked out the numbers for every candidate, including our own brand red: there
is no colour that works for both. Changing it would fix one small label and break the
text on every button. So I left it and wrote the arithmetic into the bug file for
whoever fixes the generator.

**The news page now exists.** It's at /news/index.html, it's wired into the menu in
second position, and it will list the darts stories the feed has been collecting. That
was the biggest single gap — we'd armed the feed yesterday and had nowhere to put what
it found.

**The menu is right, and I checked it the proper way this time.** All nine live pages
now serve a header with no broken links. The menu reads Home, Guides, Start Here,
Deals, Get Started, and News joins it once that page goes live. I confirmed this by
fetching each page and looking, not by querying the database — that being the exact
mistake from this morning.

**Eight of the nine buying guides are built and live**, including the three that
matter most for search: tungsten percentages, dartboard setup measurements, and the
brand comparison. I read the board-setup one closely because it's the page readers will
check hardest, and the measurements are right (1.73m to the bullseye, 2.37m throwing
distance, and it correctly tells you to measure from the board face rather than the
wall). The ninth, grip styles, belongs to another workstream and I've left it alone.

**On imagery**, I found the reason none of the generated icons are usable, and it isn't
what anyone assumed. The instructions we give the image generator literally ask for
"a darker grey line on a flat solid light grey background" — on a site with a
near-black background. Seventeen icons were generated exactly as instructed and every
one is unusable. Two of them also draw things that aren't true: a delivery van, and a
"curated product range". I've rewritten all the instructions, written a proper style
guide for the site's imagery, and marked the old icons as superseded so the system will
generate replacements. Nothing has been generated yet — no credits spent — and I'll
look at each new image before anything goes on the site.
