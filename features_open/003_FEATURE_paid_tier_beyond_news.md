# 003 FEATURE — paid tier: per-site sources, and more than news

**Raised:** 2026-07-19, by the owner, during news-feed-pooling design.
**Status:** direction agreed, scope open. Not designed.
**Related:** `features_open/001_FEATURE_packaged_topic_features.md`,
`docs026_concept_register/register/business-strategy.md` (BIZ-014),
`docs026_concept_register/register/payments.md`.

## The idea

Pooled feeds are the free/default tier: shared sources, shared enrichment,
per-site ranking, near-zero marginal cost. Above that, a **paid tier** where a
domain gets dedicated treatment.

Owner's framing:

> "The per site sources as a paid tier would work but we'll need more than just
> the news."

That caveat is the whole design problem. Dedicated news sources alone are not a
product — a buyer cannot easily tell a dedicated feed from a pooled one, and the
thing they'd be paying for (fresher, narrower sourcing) is the *least*
differentiated part of what we can do.

## Candidate axes (unranked, not yet chosen)

**Content depth**
- Bespoke topic packages (`001`) on the domain's own chosen subjects, not just
  the pool's.
- Original research: primary-source data, freedom-of-information style pulls,
  price/rate series we assemble ourselves rather than reference.
- Higher publication cadence, and structural pages (guides, glossaries) rather
  than only feed items.

**Interactive**
- Calculators and tools specific to the vertical — this platform already builds
  these, and they are the highest-intent surface on a money or insurance site.
- Data pages that update on their own (rate tables, price trackers, comparison
  grids) — closer to a live product than an article.

**Distribution**
- Outbound RSS, newsletter, syndication to the domain's own channels.

**Commercial**
- Lead capture and routing, affiliate/commission integration, conversion
  reporting — the parts a buyer measures directly.

**Operational**
- Faster refresh, human review before publish, uptime/quality guarantees.

## Open questions

- **What is the actual unit of sale** — the domain, the vertical, or the
  outcome (leads)? BIZ-014 already resolved that the domain is the unit of
  *separability-for-sale* while the cluster is the unit of blast-radius
  isolation; the pricing unit is a further question.
- **Who is the buyer?** A client renting one of our domains, a client bringing
  their own domain, or ourselves operating it and monetising directly? The
  feature set differs sharply between those, and BIZ-014's answer was
  "operator-primary at scale, vendor-optional per domain" — so this tier may be
  mostly an *internal* quality tier rather than a sold one.
- **Does the tier flag already exist?** BIZ-014 asks for a build-tier/cost-profile
  flag (`saas_cheap` vs `portfolio`) driving cheaper model and batching choices.
  If that lands first, this tier is largely a matter of what the flag gates.
  Check before designing a parallel mechanism.
- **Entitlement plumbing.** `payments.md` records a local `client_entitlements`
  cache table, needed because "the maintenance heartbeat must join across
  thousands of sites per tick" and cannot call auth synchronously. Any paid tier
  gating must read that cache, not call out.

## Why it is filed rather than designed

The pooled-feed work does not depend on this, and designing a commercial tier
before the free tier exists risks specifying against a product we have not yet
seen working. The one thing to preserve meanwhile: **keep the per-site source
path alive**. The current pipeline is entirely per-site, so the tier's mechanism
already exists — the pooling work must not delete it, only demote it from
default to opt-in.
