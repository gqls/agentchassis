# Where we are — news feeds across the domain portfolio

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-19 — first look

The question was: we're rolling out thousands of domains, most should have a news
feed, and we don't want to pay for each one separately. Could we get away with a
dozen feeds that cover most of the verticals?

The short answer is yes, a dozen is about right — but with one important
correction, which is that a dozen pools covers about two thirds of the domains,
not "the vast majority". The other third shouldn't have a news feed at all, and
that turns out to be the more useful finding.

Here's why. I went through the 1,625 domains on the list and sorted them by what
they're about. Roughly a third are either short brandable names with no meaning in
them — things like `2v.uk` and `aakn.com`, 126 of them are six characters or
shorter — or they're very narrow product sites: adjustable walking sticks, plastic
ducks, tile trimmers. There is no news about plastic ducks. There is no supplier we
could buy it from and no clever architecture that conjures it. If we point a news
feed at those domains, what we get is not a cheap news feed, it's invented filler
sitting under a heading that says "Latest News". Across hundreds of domains that is
actively harmful — these sites' main asset is that they look legitimate, and
nothing undermines that faster than a page of obvious generated padding.

So the useful way to divide the portfolio isn't by industry, it's by whether real
news actually exists for it. Where it does, we pool. Where it doesn't, we should
be doing something else entirely — seasonal and evergreen content refreshes rather
than a feed. The Christmas gift domains are the clearest example: there are sixteen
of them, and what they need is a calendar, not a news wire.

On the two thirds that do justify a feed, the good news is that they concentrate
much harder than I expected. The single biggest group is money — mortgages, loans,
savings, credit, insurance, pensions, investing — and that's about 231 domains all
of which can be fed by exactly one stream of UK financial news. Base rate moves,
FCA announcements, Budget, lender changes. One feed, 231 domains. That alone is a
seventh of the portfolio. After that it's marketing and web services (218),
construction trades, industrial and plant, travel, AI and tech, vehicles, health,
business services, property, energy, vets, and jobs. Thirteen pools, about a
thousand domains.

On the cost question itself, the reason it currently costs per-domain money is that
every part of the pipeline is built per-site, including the parts that have nothing
site-specific about them. Fetching an RSS feed and reading what's in it produces the
identical result no matter which of our domains asked for it. The only genuinely
per-site step is deciding which stories matter to that particular site — and that
step, it turns out, we can do for free, because the database already has the vector
search extension installed and we already run our own embedding model in-house for
other purposes. So the ranking that makes each site's feed feel bespoke costs us
nothing per site.

The effect is that our costs stop scaling with the number of domains and start
scaling with the number of news stories in the world, which doesn't go up when we
buy another domain. That's the whole design in one sentence.

Two things I got wrong today and corrected. I initially reported that 1,176 sites
were already set up wanting news feeds. That was wrong — I'd counted rows in a
versioned table rather than sites. The real number is four. Not four hundred, four:
gaswholesalers, relojistas, robot-hands and ai-agent-orchestration. The whole news
pipeline is a four-site prototype. That's genuinely good news, because it means
there's nothing to migrate and we can make the pooled design the default rather
than retrofitting it. The second thing was nearly reporting that a piece of the
scheduling machinery wasn't configured at all, when in fact its configuration lives
somewhere I wasn't looking. Both are written up properly in the notes.

One thing I need a decision on. There are a lot of near-duplicate domains — eleven
variations on "insurance", ten on "landlord insurance", nine on "health insurance",
`bestinsurancerate.co.uk` alongside `bestinsurancerate.uk` and so on. 358 domains
collapse into 146 concepts. Are those meant to be separate live sites, or are most
of them redirects pointing at one? It matters, because if they're separate sites
showing the same six headlines, that's a duplicate-content footprint across the
portfolio that a search engine will notice even though it's invisible when you look
at any one site.

Also worth flagging: 21.7% of these domains have any traffic at all, and one
domain — wayfaringlondoner.com — is 27% of the entire portfolio's views on its own.
So whatever we build, it's worth starting with the handful of domains that have
actual readers rather than spreading it thin across the dormant majority.

Nothing has been built. This is design only, and the plan, notes and runbook are
now in place alongside this file.

---

## 2026-07-19 — answers back, and a new directory

Owner came back on all three open questions.

The duplicate domains are meant to be **separate sites aimed at different target
markets**, so that the news on each can be angled differently. That's a bigger
decision than it sounds, because it moves the per-site ranking layer from being a
cost optimisation to being the thing that makes two similarly-named domains
genuinely different sites rather than the same site twice. It also means we now
need a real definition of who each domain is *for* — and that can't be derived
from the domain name, because the names are nearly identical in exactly the
families we most need to separate. I've flagged that as the critical unknown to
settle before anything gets built.

We're starting with the domains that have traffic. Agreed and recorded.

On the paid tier — yes, but the owner's point is that dedicated news sources
alone aren't a product, and I think that's right. A buyer can't tell a dedicated
feed from a pooled one. So that's filed as a direction with candidate axes
(bespoke research, calculators and tools, data pages that update themselves, lead
capture) rather than a spec, to be decided once the free tier actually works.

The most interesting thing to come out of this round is the owner's idea for
**packaged topic features**: pick a topic each week, break it into its parts —
for gas prices you'd pull in oil price history, the Hormuz situation, inflation,
collected opinion — and build a proper feature out of it that gets updated until
it stops being relevant. That's a much better instrument than a headline list,
and it's the owner's own answer to the duplicate content problem.

I'd add one thing to it, which I've written into the file. If we write one
package and publish it across all 231 money domains, that's *worse* for duplicate
content than headlines were, not better — long-form near-identical prose is the
shape that gets punished hardest. The fix is the same trick that makes the feed
pooling work in the first place, applied one level up: the research is shared and
done once (the expensive part), and each site generates its own angle on top of
it (the cheap part). Same facts and citations, genuinely different articles,
because they really are different articles. A haulage site gets what Hormuz means
for diesel; a landlord insurance site gets energy costs and tenant arrears. That
also happens to be exactly what the owner asked for when he said each domain
should be focused on a different target market.

New directory this session: **`features_open/`** at the repo root. We had nowhere
to put "things we've decided to build" — `bugs_open/` is explicitly for what's
broken in production right now, and putting design ideas there would wreck the
one question that directory exists to answer. So `features_open/` mirrors it, with
`FEATURE_` files for things to build and `RISK_` files for known hazards in things
we're about to build. Three entries so far: the packaged features idea, the paid
tier, and the duplicate content risk.

The duplicate content problem is worth reading properly (it's `002`). The short
version is that it's invisible from inside any single site — every site looks
fine on its own, and the defect only exists in the relationship between them. And
because so many of these domains are the same phrase on different endings, the
risk isn't just that a duplicated block ranks badly, it's that the whole estate
starts to look like a low-quality domain network. I've written down how to
actually measure it rather than just worry about it, and the rule that we test it
on one pool before rolling anything out fleet-wide — because pulling the feeds
back later doesn't undo the footprint.

027 is another thread's, so I've left it alone.
