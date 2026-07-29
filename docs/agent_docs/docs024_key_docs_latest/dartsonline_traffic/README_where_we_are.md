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
