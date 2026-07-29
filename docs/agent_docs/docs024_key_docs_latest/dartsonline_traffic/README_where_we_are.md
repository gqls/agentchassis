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
