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
3. ~~**Where do per-domain target-market profiles come from?**~~ **ANSWERED
   2026-07-19** by surveying the platform's existing cross-site machinery. See
   Decisions 7 and 8.

## Decision 7 — target-market profiles live in a `site_specs` aspect, forked from a pool default

Owner asked whether we could piggyback on existing cross-site machinery rather
than build new. We can, on three mechanisms, and the survey found a fourth that
reverses part of Decision 1's implementation (see Decision 8).

**(a) The home for a profile: `site_specs` aspect `audience`. Zero schema change.**

Verified live 2026-07-19:
- `site_specs` has **no constraint on `aspect`** — only `site_specs_pkey` and
  `site_specs_site_id_fkey`. The column is free-form text; Go checks only that it
  is non-empty (`site_spec_actions.go:178-180`). A new aspect needs no migration.
- An **`audience` aspect already exists** with 2 current rows
  (`ai-agent-orchestration.com`, `leopardessconsulting.co.uk`), written by
  `content-gap-planner`.
- The admin API **already branches on `aspect == "audience"`**
  (`internal/core-manager/admin/spec_admin_handlers.go:226`) — dead anticipatory
  code for an aspect nothing officially writes. Someone reached for this once.
- Aspects are versioned with `is_current` and a unique index on
  `(site_id, aspect) WHERE is_current` — history and rollback for free.

**Substrate to start from, not a blank page:** `identity.target_audience` is
populated for all 11 sites with genuinely useful prose (e.g. relojistas:
*"Aficionados, coleccionistas y curiosos de la relojería en España, México,
Chile…"*). Seed the new aspect from it.

**(b) The share/fork rule: copy the component library's, exactly.**

`forked_from IS NULL` is the platform's **only live, incident-hardened cross-site
decision mechanism**. TLIB-022's standing rule transfers verbatim:

> regenerate a shared base only for neutral, purely-additive improvements;
> site-specific voice must FORK.

So: a **pool-level default audience profile** (the shared base) that every domain
in the pool inherits, and a **per-site fork** the moment a domain needs its own
angle. Near-identical families (`bestinsurancerate.co.uk` / `.uk`) are precisely
the case that must fork — a shared profile guarantees identical ranking, which is
the failure mode Decision 4 exists to prevent.

Heed the incidents too: TLIB-005 records a shared-row field rename that *"silently
emptied every dependent"* across five instances on multiple sites. Any change to a
pool-level profile has fleet blast radius. `store_generated_component_action.go:331`
implements the mitigation — a field-set guard that permits additions and rejects
renames/drops. Mirror it.

**(c) Binding sites to pools: `js_snippets.applies_to`.**

A declarative "one row → N sites" selector already exists and is tested:
`render_js_snippets_for_site_action.go:150-159` matches a library row's JSONB
`applies_to` set against the site's own set. The same shape binds a domain to a
pool without a join table.

**What genuinely cannot be piggybacked — and is the real gap:**

Nothing in the platform records *"this site is positioned differently from our
**other** site"*. Every differentiation field is against **external** competitors:
`strategy.competitive_position`, `identity.unique_selling_points`,
`identity.competitors_found`. Intra-portfolio positioning has no home. That is the
one genuinely new thing Decision 4 requires, and the `audience` aspect is where it
should go.

**Two cautions found in the survey:**
- ~~`identity.audience_primary` / `audience_secondary` / `sophistication` was
  designed and then **explicitly reverted** in `003_site_classifier.sql`.
  **Find out why before rebuilding the same shape.**~~
  > **CORRECTED 2026-07-20 — the caution was overstated, and I had not read the
  > revert block before recording it.** The revert is real but its stated reason is
  > mechanical, not a judgement on the design: *"Undo 004 + 005 changes to existing
  > agents … Restores original state for agents modified by the **incorrect 004/005
  > scripts**"* (`003_site_classifier.sql:170-174`). The rich profile was collateral
  > — it lived in the same `site-classifier` row that got rolled back wholesale to
  > the original single-step Haiku task. **The audience fields were never rejected
  > on their merits.**
  > Further, `site-classifier` is **legacy**: the live intake chain uses
  > `domain-research-classifier` (referenced 4× across live agent configs vs 1× for
  > `site-classifier`). So this is prior art from a superseded agent, not a warning.
  > **Treat `audience_primary` / `audience_secondary` / `sophistication` as a
  > usable precedent** — someone had already reached the same shape.
  > *What caught it:* reading the revert block, which I had cited without opening.
- The audience question was **dropped** from the current briefing questionnaire —
  the backup had a required *"Who is your target audience?"* textarea; the live
  `026_pageflow_builder.sql:868` has no audience section. We have been losing this
  capture over time; re-adding it is part of the fix, not a separate task.
- Aspect sprawl is real and this decision worsens it if taken carelessly: 38
  distinct current aspects, many singletons, with overlapping families already
  present (`voice` / `voice_and_tone` / `voice_and_audience`;
  `content_direction` / `content_standards` / `content_rules`). Settle the
  `audience` schema rather than adding a 39th ad-hoc shape.

## Decision 8 — a pool is a synthetic site, not a new column

> **CORRECTION 2026-07-19 to Decision 1's implementation sketch.** The original
> plan proposed adding `pool_id` to `content_sources` and `content_feed_items`,
> making `content_sources.site_id` nullable, and treating
> `content_feed_items.site_id`'s existing nullability as "the pooling slot that
> already exists". **That runs against the platform's established idiom**, and I
> proposed it before surveying how the platform actually handles ownerless work.
> What caught it: the cross-site machinery survey the owner asked for.

The platform's idiom for "work that belongs to the platform rather than to a
customer site" is a **synthetic site row**, not a nullable owner. `system.internal`
(verified live: `eac60db8-b032-432b-b36d-76f37632045d`, `status='system'`) exists
precisely because `site_work_items.site_id` is `NOT NULL`, and rather than invent
a null-site mechanism the platform created a site record to own shared work.
TLIB-018 states the intent:

> A synthetic site record that owns library-level work … **so the ordinary
> `site_work_items`/dispatch machinery can operate on the shared component library
> exactly as it does on a customer site.**

**Applied to feeds: each pool is a synthetic site.** Consequences:

- `content_sources.site_id NOT NULL` is satisfied — **no schema change**.
- `FetchRSSAction`, `WriteFeedItemsAction`, `LoadDueSourcesAction`,
  `UpdateSourceTimestampsAction`, `DispatchFeedSourcesAction` and the site-scoped
  dedup all work **unchanged**. The pool ingests exactly as a site does today.
- No change to `idx_cfi_dedup`, which removes the lockstep hazard listed in Risk 5
  — the class that has already caused a fleet-wide 42P10 outage in this repo.
- Real sites then *select* from the pool rather than ingesting; only the selection
  and render paths are new.

This is substantially less code than the column-based design, and it is a
mechanism the platform has already hardened in production rather than one we
invent. It also keeps per-site sources working untouched, which Decision 6 /
`features_open/003` requires.

**Open sub-question:** whether pool sites are excluded from fleet loops that
iterate `sites WHERE status='deployed'` (`maintenance_actions.go:694-698`).
`system.internal` uses `status='system'`, which already sits outside that
predicate — likely free, but must be verified, not assumed. CTS-018 notes
`system.internal` has a live side effect of absorbing untargeted scheduler
dispatches; check whether pool sites would too.

## Decision 9 — the `audience` aspect separates who they are from what to do about it

Settled 2026-07-20 after reading the two live `audience` rows. **The evidence
changed the shape**, so the reasoning matters more than the schema.

**What the live rows actually contain.** Both current `audience` rows were written
by `content-gap-planner` — a *remediation* agent — and both mix audience identity
with editorial instruction in one prose blob:

> `ai-agent-orchestration.com` (key: `primary_buyer_hierarchy`): *"Primary buyer is
> the CTO or VP Engineering … **Lead all page copy with the technical failure mode**
> … **CTA language is 'Technical Discovery Call' sitewide.**"*

> `leopardessconsulting.co.uk` (key: `target_audience`): *"Primary: CTOs,
> engineering leads … Non-technical SMB buyers are explicitly out of scope. **All
> pages should be written for the primary audience by default** … **Pages that
> currently address non-technical readers (e.g. about page) should be revised.**"*

That conflation is fatal for our purpose: **a ranking layer cannot consume "CTA
language is 'Technical Discovery Call' sitewide."** It is an instruction to a
content agent, not a description of a reader. The differing key names
(`primary_buyer_hierarchy` vs `target_audience`) are a symptom of the same thing —
neither row was written to a schema because the aspect has never had one.

**So the aspect splits in three:**

| block | purpose | consumed by |
|---|---|---|
| `who` | who the reader is — prose + `sophistication` enum. The **embeddable** part. | feed ranking, angle generation |
| `position` | how this domain differs from **our own** sibling domains | feed ranking (divergence), angle generation |
| `editorial` | what that implies for copy, CTAs, register | content agents — **never** the ranking layer |

`who` reuses the prior art the caution above wrongly warned us off:
`audience_primary`, `audience_secondary`, `sophistication` (enum:
`technical|professional|casual|luxury|institutional|editorial`). Add
`out_of_scope`, because the leopardess row shows the field is genuinely used in
practice (*"Non-technical SMB buyers are explicitly out of scope"*) and a negative
constraint is directly usable as a ranking penalty.

`position` is the genuinely new part — the intra-portfolio gap identified in
Decision 7. It names sibling domains and states how this one differs. **This is
the field that makes `bestinsurancerate.co.uk` and `bestinsurancerate.uk`
different sites**, and nothing else in the platform carries it.

**Seed, don't start blank:** `identity.target_audience` is populated for all 11
sites with usable prose. It seeds `who.audience_primary` directly.

**Embedding:** the vector used for ranking is computed from `who` + `position`
only. Excluding `editorial` is not tidiness — including it would make two sites
with similar copy rules rank alike regardless of audience, which is the exact
failure we are designing against.

**Open:** whether to migrate the two existing rows now (splitting their prose into
the three blocks) or leave them and write the schema forward. Leaning migrate —
two rows is cheap, and leaving two non-conforming rows in a newly-schema'd aspect
is how the next thread learns the wrong shape.

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
