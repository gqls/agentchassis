# PLAN — pooled news feeds for a thousand-domain portfolio

**Started:** 2026-07-19
**Status:** design. Nothing built. No code or schema change proposed for approval yet.

## The brief, as given

> "We will [be] rolling out thousands of domains and I'd like most of them to have
> a news feed but I don't want to pay for each individually. Please can you think
> about how we can get just a certain number of newsfeeds maybe 10 or a dozen or so
> that covers the vast majority of the verticals."

## Correction to the brief (2026-07-19)

The count is right; the coverage claim is not, and the difference changes what we build.

**A dozen pools is the correct order of magnitude — but they cover roughly two
thirds of the portfolio, not "the vast majority", because the remaining third
should not have a news feed at all.**

Measured against the 1,625-domain abbreviated list (see NOTES for method):
~34% of domains are either short brandables (`2v.uk`, `aakn.com`, `kzlu.com` —
126 domains of ≤6 characters) or ultra-long-tail product microsites
(`adjustablewalkingsticks.com`, `plasticducks.com`, `tiletrimmers.com`).
There is no news stream for a plastic duck. There is no price at which one exists.
Pointing a feed at these domains does not produce a cheap feed; it produces
generated filler on a page labelled "Latest News", which is worse than an absent
section — it is the slop signature, at portfolio scale, on domains whose only
asset is that they look legitimate.

So the taxonomy that governs the build is **news supply**, not industry. A vertical
earns a pool when a real, dateable, externally-sourced stream exists for it.

## Decision 1 — split the expensive half from the differentiating half

The reason per-site feeds cost per-site money is that every layer is site-scoped,
including the layers that have nothing site-specific in them. Fetching an RSS URL,
parsing it, de-duplicating it and extracting its topics produce a byte-identical
result no matter which of our domains asked. Only *ranking* is site-specific.

| Layer | Shared (~12×) | Per-site (N×) |
|---|---|---|
| Fetch RSS / search / scrape | ✅ | |
| LLM news search (Grok/Perplexity) | ✅ | |
| Canonicalise, dedup, enrich (topics, entities, embedding, own summary) | ✅ once per **article** | |
| Rank against this site's profile | | ✅ **zero LLM** |
| Render JSON + commit | | ✅ (cheap) |

**Why this works here specifically:** pgvector 0.8.0 is already installed on
`postgres-clients-0`, and the platform already generates embeddings via
self-hosted Ollama `nomic-embed-text` (768-dim, ivfflat cosine) for the RAG
knowledge base. Per-site ranking is therefore a SQL query against infrastructure
we already run, at no marginal API cost. This is reuse, not new machinery.

**Consequence:** LLM spend becomes O(articles), not O(sites × articles). The number
of news stories in the world does not increase when we add a domain, so cost is
flat in portfolio size. That is the whole point of the design; every other benefit
is secondary.

## Decision 2 — the pool set is derived from supply, not declared

A pool is justified when it can produce ~20 fresh relevant items/day. Below that it
merges upward; far above it with poor internal coherence, it splits. The list below
is the **straw man from the domain analysis**, to be re-cut once we know which
domains actually become sites.

| # | Pool | Domains | Notes |
|---|---|---|---|
| 1 | `uk-money` | ~231 | mortgages, loans, savings, credit, insurance, pensions, investing — **one** BoE/FCA/rates/Budget stream feeds all of it. Biggest single win in the portfolio. |
| 2 | `marketing-web-digital` | ~218 | SEO/algorithm/adtech news; abundant supply |
| 3 | `construction-trades` | ~79 | |
| 4 | `industrial-plant-logistics` | ~71 | |
| 5 | `travel-leisure` | ~69 | |
| 6 | `ai-tech` | ~66 | abundant supply |
| 7 | `vehicles-transport` | ~65 | |
| 8 | `health-medical` | ~63 | |
| 9 | `business-services` | ~54 | |
| 10 | `property-uk` | ~39 | may merge into `uk-money`; property news is abundant enough to stand alone |
| 11 | `energy-utilities` | ~39 | price cap / Ofgem; abundant |
| 12 | `vet-animal` | ~34 | existing vet workstreams give us a head start |
| 13 | `jobs-work` | ~20 | candidate to fold into `business-services` |

≈1,048 domains ≈ 64.5% of the list. **The other ~35% get no feed**, by design.

## Decision 3 — the no-feed third gets a different instrument, later

Retail/product microsites and seasonal domains (`christmaspresents.co.uk`,
`giftxmas.com` — 16 christmas/xmas domains) do not want news; they want *evergreen
and seasonal freshness*. That is a calendar and a content-refresh cadence, not a
feed. **Out of scope for this workstream.** Recording it here so the gap is a
decision rather than an oversight.

## Decision 4 — every domain is a separate site with its own target market (owner, 2026-07-19)

Answering the open question below: the near-duplicate families are **not**
redirects. Each domain is intended to be a distinct site aimed at a different
target market, so that news selection can be angled differently per domain.

**This raises the stakes on the per-site ranking layer rather than lowering them.**
It is now load-bearing product behaviour, not just a cost optimisation: the
per-site profile is what makes `bestinsurancerate.co.uk` and
`bestinsurancerate.uk` different sites rather than the same site twice. The
profile therefore needs a real target-market definition per domain, not a
keyword list derived from the domain name — two domains with near-identical names
must carry deliberately different audience profiles or they will rank the pool
identically.

It also makes `features_open/002_RISK_portfolio_duplicate_content.md` the
governing constraint of this workstream.

## Decision 5 — launch order follows traffic (owner, 2026-07-19)

Start with the domains that have readers. Only 21.7% of the list has any views;
`wayfaringlondoner.com` alone is 27% of portfolio traffic. Piloting on
traffic-bearing domains also gives the duplicate-content measurement (002) a real
signal instead of a synthetic one.

## Decision 6 — differentiation comes from synthesis, not selection (owner, 2026-07-19)

Owner: *"We will have to think how to avoid the serious duplicate content feed
problem more and maybe it comes when we start analysing feed topics and writing
our own brief rundowns."* Agreed and filed as
`features_open/001_FEATURE_packaged_topic_features.md`. The key structural note
recorded there: a package written once per pool and syndicated is **worse** than
a headline list; the package must be a shared research substrate with a per-site
angle generated on top.

## Open questions for the owner

1. ~~**The duplicate-TLD families.**~~ **RESOLVED 2026-07-19** — separate sites,
   separate target markets. See Decision 4.
2. ~~**Tiering.**~~ **RESOLVED 2026-07-19** — yes, but it needs to be more than
   news to be a product. Filed as
   `features_open/003_FEATURE_paid_tier_beyond_news.md`. Keep the per-site source
   path alive; the pooling work demotes it from default to opt-in, never deletes it.
3. **Where do per-domain target-market profiles come from?** Decision 4 makes this
   the critical unknown. Hand-authored per domain does not scale to thousands;
   derived from the domain name collapses exactly the families we need to
   separate. Candidate: derive once at site creation from the site's own
   classification spec + a stated audience, then embed. Needs deciding before the
   ranking layer is built.

## Risks carried into the build

1. **Near-duplicate content across the portfolio.** 500 sites rendering the same six
   headlines is invisible when testing one site and obvious to a search engine
   looking at the footprint. Mitigations: divergent per-site ranking, our own
   summary written once at enrichment, and the existing title+summary+link-out
   rights posture. Needs measuring, not assuming.
2. **Thin pools starve niche sites.** A broad pool may hold nothing for a
   boiler-servicing site on a given day. Needs an explicit relevance floor with a
   "render nothing" branch. An absent feed beats an irrelevant one.
3. **Git is the next bottleneck after LLM cost.** Rendering is O(sites) and cheap,
   but thousands of per-site commits per cycle is its own wall — and `bugs_open/014`
   already shows feed artefacts misrouting between repos.
4. **`bugs_open/027`: news pages render no news without JavaScript.** Directly
   material — if the feed is the SEO instrument, a JS-only render defeats the
   purpose fleet-wide. Fix before rollout, not after.
5. **Dedup index ↔ Go insert lockstep.** Adding `pool_id` to `content_feed_items`
   changes `idx_cfi_dedup`; this repo has already taken a fleet-wide 42P10 outage
   from exactly that class of drift.
