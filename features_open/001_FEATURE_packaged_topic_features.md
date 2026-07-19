# 001 FEATURE — packaged topic features ("living dossiers")

**Raised:** 2026-07-19, by the owner, during news-feed-pooling design.
**Status:** idea captured. Not designed, not scheduled.
**Related:** `features_open/002_RISK_portfolio_duplicate_content.md` (this feature
is the leading candidate mitigation), `docs024_key_docs_latest/news_feed_pooling/`.

## The idea, in the owner's framing

> "I intend to select a pertinent topic, say per week, break down its parts —
> e.g. if it is gas prices then we'd look at previous gas, oil prices, the Hormuz
> war, inflation rates and collected opinions etc — and create a packaged feature
> about it, updated as we go until it gets irrelevant."

So: not an aggregated headline list. A **synthesised explainer** on one topic,
assembled from several dimensions at once, with a lifecycle — created, updated
while live, retired when it stops mattering.

## Why this is a different instrument from the news feed

The feed answers "what happened today". This answers "what is going on, and what
does it mean for you". They have different half-lives, different value, and —
critically — different duplication profiles.

A headline list is inherently near-identical across sites drawing on the same
pool. A synthesised feature does not have to be, because the synthesis is where
the angle lives.

## The structural point: the same pooling shape, one level up

**A package written once per pool and published to all its domains is *worse* for
duplicate content than a headline list, not better** — it is long-form
near-identical prose, which is the shape search engines penalise hardest. The
naive version of this feature makes 002 more dangerous, not less.

The resolution is the same split that makes pooled feeds work:

| layer | shared, once per package | per-site |
|---|---|---|
| Topic selection | ✅ | |
| Research: price history, timeline, related events, macro figures, source gathering, quote collection | ✅ **the expensive part** | |
| Fact/figure substrate with citations | ✅ | |
| **The angle** — what this means for *this* audience | | ✅ one focused generation |
| Headline, framing, examples, call to action | | ✅ |

One gas-prices research substrate; a haulage site gets "what Hormuz means for
diesel costs this quarter", a landlord-insurance site gets "energy costs and
tenant arrears risk", a manufacturer gets "input-cost hedging". Same facts, same
citations, genuinely different articles — because they *are* genuinely different
articles, not spun ones.

This also serves the owner's stated goal directly ("I want each domain separate
and focused to different target markets... so the news selection can hopefully be
targeted slightly differently"). The per-site angle is where that focus lives.

## Cost shape

Per-site angle generation *is* an LLM call per site per package — this is the one
place we deliberately spend per-site money, because it is the differentiating
output rather than the commodity input. At one package/week, the ~231-domain money
pool costs ~231 generations/week. Compare the naive per-site feed design, which
projected ~8,000 triage calls/**day** at 2,000 sites. Affordable, and spent on the
right layer.

## Open questions

- **Who picks the topic?** Editorially chosen by us, or detected from feed volume
  (a topic spike in the pool)? Detection is cheap given we already enrich and
  embed every article; a cluster forming in embedding space *is* the signal.
- **What is the update trigger and the retirement rule?** "Until it gets
  irrelevant" needs an operational test — decaying article volume in the cluster
  is the obvious candidate.
- **How does an updated package behave for SEO?** Updating one URL in place
  accrues authority; publishing a new URL each week fragments it. Probably
  in-place with a visible "updated" date and a changelog, but this needs deciding
  before the first one ships, because it is expensive to reverse.
- **Does the substrate get stored as a first-class entity?** A `topic_packages`
  table with its own lifecycle, versus a work item that produces pages. The
  former is reusable and lets a site join a package late; the latter is less to
  build.
- **Rights.** Collected opinions and quotes mean quoting third parties at length.
  The feed's existing posture is title + short summary + link-out, never
  full-text republication. A synthesised feature is safer (it is our own prose)
  but quote length and attribution still need a rule.

## Why it is worth building

It is the only mitigation for 002 that turns the duplicate-content constraint
into an *asset*: the expensive research is amortised across the pool exactly as
the feed fetching is, while the output is more differentiated than a headline
list, not less. And it is the only part of the news offering that is plausibly
worth paying for (see 003).
