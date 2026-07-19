# 002 RISK — portfolio-wide duplicate content from pooled feeds

**Raised:** 2026-07-19, during news-feed-pooling design.
**Status:** latent. Not reproducible today — the design it threatens is unbuilt.
**Becomes a bug** the day pooled feeds ship to more than a handful of same-pool
domains, at which point this file moves to `/bugs_open/`.
**Related:** `features_open/001_FEATURE_packaged_topic_features.md` (leading
mitigation), `docs024_key_docs_latest/news_feed_pooling/PLAN_2026-07-19_*.md`.

## The mechanism

Pooled feeds work by having many sites draw from one shared article pool. If
per-site selection ranks the same pool by broadly the same criteria — recency
plus relevance to a similar profile — then sites in the same pool converge on the
**same top-K articles**, and render near-identical "Latest News" blocks.

The convergence is worst exactly where the portfolio is densest. The `uk-money`
pool is ~231 domains. Within it sit families like `bestinsurancerate.co.uk` /
`bestinsurancerate.uk`, eleven variants of `insurance*`, ten of
`landlordinsurance*`, nine of `healthinsurance*` — 358 domains collapsing to 146
concepts across the list. Those are the sites most likely to rank identically,
because their profiles are nearly identical by construction.

## Why it is easy to miss

**It is invisible from inside any single site.** Every individual site looks
correct: a fresh, relevant, well-formed news block. The defect is a property of
the *set*, and nothing in our current verification looks at the set. Discovery
checks are per-site. The council reviews a diff. A visual check of one page passes.

This is the same shape as other defects this repo has been bitten by — a fix
that covers one branch of a two-branch router reads as done; a tag is not
evidence across services. Per-item verification cannot see a cross-item property.

## Why it matters more for this portfolio than in general

Duplicated boilerplate across a few sites is ordinary and mostly ignored. What is
not ordinary is a **large set of exact-match-keyword domains, on multiple TLDs of
the same phrase, thin on original content, carrying identical blocks**. That is
recognisably the shape of a low-quality domain network, and the risk is not
merely that the duplicated block ranks poorly — it is a **footprint risk** that
can attach to the portfolio as a whole.

Given that roughly a fifth of these domains currently have any traffic at all,
the asset being risked is future ranking potential across the whole estate, in
exchange for a content block of modest value. That is a bad trade to make
accidentally, which is precisely how it would be made.

## Mitigations, weakest to strongest

1. **Divergent ranking** — per-site profile embeddings, different recency decay,
   per-source diversity caps, rotation within the relevant band. Cheap, helps,
   nowhere near sufficient on its own: the pool's genuinely most-relevant story
   is the same story for everyone, and suppressing it to force variety means
   deliberately showing worse content.
2. **Our own summary, written once at enrichment** — so the prose is ours rather
   than the publisher's. Removes the *scraped*-duplicate problem; does nothing
   about the *internal* one, which is the one that matters here.
3. **Link-out only, title + short summary, never full text** — the existing
   rights posture. Keeps the duplicated surface small.
4. **Fewer sites carry a feed at all** — ~34% of the portfolio should have no
   feed regardless (no news supply). Within a duplicate family, plausibly only
   the canonical domain carries one.
5. **Per-site synthesis rather than shared selection** — `001`. The only
   mitigation that produces genuinely different text rather than differently
   ordered identical text.

## How to measure it (a risk with no test is an opinion)

Before rollout, and continuously after:

- Render the news block for every site in a pool, normalise, and compute
  **pairwise similarity across sites** (shingled Jaccard, or cosine over the
  block embedding — we already have pgvector and a local embedding model).
- Report the **distribution**, not the mean — the mean will look fine while the
  worst decile is identical. The number that matters is: how many *pairs* exceed
  the threshold, and are they concentrated in one family.
- Set the gate before seeing the data, so it cannot be rationalised afterwards.
- Also track **overlap of the article set** (how many of each site's top 6 are
  shared with its nearest sibling) separately from text similarity — they can
  diverge, and set overlap is the leading indicator.

Run this on the pool's *rendered output*, not on the selection query's intent.
Trust the rendered artefact, not the status.

## Do not

Ship pooled feeds fleet-wide and check afterwards. The failure mode is slow,
cumulative and attaches to the domains rather than to a deployment, so it is not
cleanly revertable — pulling the feeds later does not restore what the footprint
cost. Measure on a pilot within one pool first.
