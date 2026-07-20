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

~~**Open:** whether to migrate the two existing rows now.~~ **RESOLVED — owner
chose migrate, thoroughly. DONE 2026-07-20.** Both rows migrated to `audience.v1`
in one transaction (supersede-then-insert; the partial unique index on
`(site_id, aspect) WHERE is_current` forces that order). Verified: originals
preserved as superseded versions (leopardess retains an older March version too —
the history chain is intact), all four v1 keys present on both current rows,
`position` correctly null (the source prose contained none — not invented),
and distinctive clauses from each original spot-checked present
(`Technical Discovery Call`, `bridges both registers`, `production credibility`,
`register consistency`). SQL preserved in RUNBOOK.

## Decision 10 — pools v2: the big groups split (owner, 2026-07-20)

Owner: with ~231 domains in one finance group we can afford more specific
categories. Re-cut with the money and marketing groups split; now **17 pools,
1,037 domains (63.8%)** — the coverage barely moves, the pools get sharper:

| pool | domains | views | was |
|---|---|---|---|
| design-creative | 126 | 207 | marketing-web |
| insurance | 83 | 34 | uk-money |
| mortgages-lending | 80 | 113 | uk-money |
| construction-trades | 78 | 268 | — |
| travel-leisure | 74 | 2,246 | — |
| industrial-plant | 68 | 84 | — |
| savings-investing | 68 | 265 | uk-money |
| health-medical | 65 | 151 | — |
| vehicles-transport | 64 | 65 | — |
| business-services | 62 | 164 | — |
| web-tech | 54 | 12 | marketing-web |
| marketing-digital | 51 | 19 | marketing-web |
| ai-agents | 47 | 812 | ai-tech |
| energy-utilities | 41 | 20 | — |
| property | 37 | 177 | — |
| vet-animal | 26 | 50 | — |
| jobs-work | 13 | 2 | — |

uk-money (231) → insurance + mortgages-lending + savings-investing (+ property
already separate); marketing-web (218) → design-creative + web-tech +
marketing-digital. Each split still clears the supply test: insurance trade
press, BoE/lender news, and markets news are three genuinely different streams.

Two candidate merges if any pool runs thin in practice: `jobs-work` (13) into
`business-services`, `vet-animal` (26) already has a head start from the vet
workstreams so it stays despite size. **These numbers are regex-crude sizing, not
assignments** — final pool membership is set per-site at onboarding, where the
classifier caveat (RUNBOOK) doesn't bite.

## Decision 11 — pilot cohort: traffic-bearing, pool-eligible (owner, 2026-07-19/20)

Pilot = pool-eligible domains with ≥10 views. **31 domains, ~4,300 views** after
hand-correcting two classifier misfits (`whatvacancy.com` → jobs-work, matched
"vacan"→travel; `greenpowerjuicers.com` → no-feed, a juicer shop that matched
"power"). Top of the cohort:

| views | domain | pool |
|---|---|---|
| 1,997 | wayfaringlondoner.com | travel-leisure |
| 655 | traderboltai.com | ai-agents |
| 226 | kitchensep.com | construction-trades |
| 146 | monitorizare.com | ai-agents |
| 131 | thecentralbanker.com | savings-investing |
| 125 | hcare.co.uk | health-medical |
| 113 | thoroughcleaners.com | business-services |
| 108 | lodgeswithhottubs.club | travel-leisure |
| 105 | toletonline.com | property |
| 88 | hoeinvestereninvastgoed.com | savings-investing |

The cohort touches **14 of the 17 pools**, so the pilot exercises breadth as well
as traffic. Two flags:

1. **The no-feed class holds serious traffic that needs an owner pass.** 12
   excluded domains have ≥25 views, including the portfolio's #2
   (`smartbusinesssupplies.com`, 748), `zdec.com` (409), `komunikatif.com` (253),
   `makeitaquote.com` (226), `buysportskit.com` (215). Most are brandables or
   tools (correctly excluded), but `buysportskit`/`sportswearinc` could join a
   sport-adjacent pool and `smartbusinesssupplies` could take business news.
   Worth ten minutes of owner eyes rather than a rule.
2. **The pilot is multilingual on day one.** `monitorizare` (Romanian),
   `hoeinvestereninvastgoed` (Dutch), `apetlon` (German), `ideal-property-mallorca`
   (Spanish market). `bugs_open/026` (news-listing hardcodes English) is therefore
   a **pilot** blocker, not just a fleet blocker.

## Decision 12 — duplicate-content council seat (owner, 2026-07-20: "we might need")

Filed as `features_open/004_FEATURE_duplicate_content_council_seat.md` rather than
built — the seat becomes reviewable the moment pooled selection exists, which is
also the moment it has something to veto. Mechanism is standard: seat
`fix-proposer`, then run the 099 roster mirror (CLAUDE.md). The seat's footprint
and review posture are specified in the feature file.

## Decision 13 — other-language domains are exceptions with per-language feeds (owner, 2026-07-20)

Non-English domains get feeds **in their own language**, following the relojistas
precedent (Spanish Grok prompt + Spanish RSS sources, live today). A pool's
language is part of its identity: a Dutch-language savings-investing domain
(`hoeinvestereninvastgoed.com`) is served by Dutch sources, not by translating
the UK pool. Whether that means a language-variant of the pool or a per-site
source set is a sizing question — one Dutch finance domain does not justify a
pool; five might. `bugs_open/026` (news-listing hardcodes English) remains a
pilot blocker regardless — rendering must honour the site's language before any
non-English feed ships.

## Decision 14 — real-traffic domains get dedicated feeds (owner, 2026-07-20)

Traffic-bearing domains get **their own specific news sources on top of pool
membership** — "to give them the best chance in life". This is the
`features_open/003` per-site mechanism, applied internally to our best domains
rather than sold: the existing per-site `content_sources` path (which Decision 8
deliberately left untouched) is the implementation, so nothing new is needed —
dedicated sources are rows, exactly as relojistas' five hand-verified Spanish RSS
feeds are today. The pool remains the floor; dedicated sources are the ceiling.
Which sources each traffic domain gets is informed by the domain-history research
(in flight — deep-research run `wf_2f8a91fd-b77` on the opaque traffic domains).

## Decision 15 — audience profiles: all 11 live sites DONE; pilot domains wait for onboarding (2026-07-20)

**Every live site now carries a current `audience.v1` row** (authored 2026-07-20,
verified: 11 current rows, all four keys, originals preserved beneath — 16 rows
total including history).

Authoring rules applied, and binding for future profile authorship:
- `who` restructures the site's real `identity.target_audience` — nothing invented.
- `sophistication` from evidence (result: 4 technical, 2 professional, 3 casual,
  2 editorial — the spread itself shows the field carries signal).
- `out_of_scope` only where the source states one (gaswholesalers' "over
  low-touch commodity brokering" contrast; leopardess' explicit exclusion).
- `editorial` stays null until a genuine directive exists — `content_direction`
  and `voice` own that ground today.
- **`position` only where grounded against LIVE siblings.** The AI trio
  (finetuning.uk / ai-agent-orchestration.com / leopardessconsulting.co.uk) is
  the only live cluster, so those three have positions — each written against the
  others' actual audience rows: SME-outcomes vs enterprise-architecture vs
  delivery-credibility. The two migrated rows took a v2 (supersede→insert, `||`
  merge preserving who/editorial).
- Null positions carry the **reason** in `notes` (which sibling isn't live yet),
  so onboarding threads know when to fill them. vetcomparison's note flags the
  20+ vet* domains as the estate's clearest future position-need.

**The 31 pilot domains get profiles at onboarding, not before.** They have no
sites rows, and authoring from nothing but a domain name is the name-derived trap
Decision 4 exists to prevent. The onboarding chain (domain-research-classifier →
`identity.target_audience`) provides the derivation source; re-adding the dropped
audience question to the briefing questionnaire (Decision 7 caution) is part of
that path.

## Decision 16 — ownership resolved: all nine are ours; nanangmrk is owner-run and gets ADOPTED (owner, 2026-07-20)

The reconciliation question from the research is answered: **the owner owns all
nine**, including the two showing Afternic for-sale redirects (registrar parking,
not third-party ownership). **nanangmrk.com is run by the owner personally** —
the live Indonesian tutorial site and its 551k-subscriber YouTube channel are
his. Directive: **adopt it into the framework and start managing it**.

Adoption is a distinct track from a domain build — the site has real content, a
real audience and a real external traffic engine (YouTube), so this is the
relojistas class of work (take over a living property without breaking it), not
the greenfield class. The platform has an adoption pipeline concept
(`docs026_concept_register/register/adoption-pipeline.md`,
`adopting-and-scraping.md`) — check its real state before starting; this becomes
its own workstream, not a task inside this one. Its content instrument is
evergreen tutorials + YouTube companionship, not a news feed.

## Decision 17 — retail-legacy domains become retailer-directory utilities (owner, 2026-07-20)

For the commerce-legacy traffic domains — **buysportskit.com,
smartbusinesssupplies.com, and outfax.com on the same pattern** — the owner's
directive: *"set up categorised product listings that link through to real
retailers, we can later do affiliate feeds but for now just be as useful to the
users as we can."*

This is a **new content instrument** — neither a news feed nor a brochure site: a
categorised directory of real products/services at real retailers, honest about
what it is, that inherits the arriving purchase intent by actually serving it.
Sequencing is utility-first: no affiliate wrapping until the listings are
genuinely useful (which also keeps the early sites clean of the thin-affiliate
signature while they re-establish standing).

- buysportskit: teamwear/kit categories → real UK kit retailers. The predecessor
  (BSK Pro) still trades at `.shop` — **no Errea product naming lifted from their
  old URLs, no club-shop branding**; serve the intent, not their catalogue.
- smartbusinesssupplies: office/business-supplies categories → real UK suppliers.
  The shadowed company is dead but may exist legally — our own identity throughout.
- outfax: "send a fax online" → current fax service providers, comparison-style.
  The fax-affiliate space is mature, so the later-affiliate step is natural here.

Prior art to check before building: the concept register has an
`affiliate-commerce.md` entry and the strategist's `revenue_models` already
includes `lead_generation`/`affiliate` — reuse the machinery view first.

## Decision 18 — makeitaquote: build the tool, deliberately different (owner, 2026-07-20)

Build the quote-image tool but *"slightly different (better) so we're not copying
or landing into passing off territory"*. Concretely that means: our own name and
branding on the tool itself (the domain name describes the function, which is
fine — passing-off risk lives in imitating the bot's look, name-styling, or
claiming to BE the bot); a web-first feature set the Discord bot doesn't offer
(paste text, upload avatar, style/palette choices, direct download/share) rather
than a webified clone of its reply-to-a-message flow; and an explicit "not
affiliated with the Discord bot" line. Sibling positioning vs memecreator.co.uk /
memegenerator.uk goes in the `position` field at onboarding — this trio is the
second genuine position cluster after the AI trio.

## Decision 19 — pilot list grows to ~36 (owner, 2026-07-20)

Added on the strength of their legacies: **smartbusinesssupplies.com (748),
makeitaquote.com (226), buysportskit.com (215), nanangmrk.com (95, via the
adoption track), outfax.com (64)**. That puts the portfolio's #2, #6 and #7
traffic domains in the pilot. komunikatif.com (253, Indonesian news legacy)
remains **recommended but not yet directed** — it is the strongest news-shaped
inheritance in the set and the natural first Indonesian language exception.

**zdec.com is deliberately NOT strategised yet.** Owner: "we can completely
restrategise zdec.com then, ideas welcome." Proposal recorded below as a
proposal, not a decision:

> **Measure before strategising.** The domain's 409 views are of unknown quality
> (hacked-era casino-spam backlinks; possible search-engine distrust). Cheapest
> honest first step: serve a minimal holding page with analytics for 2–3 weeks
> and characterise the traffic — human vs bot, referrers, geography, landing
> paths. Then branch: (a) if there is real Chinese-industrial residue, the
> industrial-controls history suggests an industrial-plant-pool site (possibly
> bilingual); (b) if the traffic is spam residue, treat zdec as a clean-slate
> 4-letter brandable and accept that its search standing may need a disavow file
> and patience — build for direct/type-in value, not SEO, until standing is
> proven. Committing content before measuring risks investing in a penalised
> asset.

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
