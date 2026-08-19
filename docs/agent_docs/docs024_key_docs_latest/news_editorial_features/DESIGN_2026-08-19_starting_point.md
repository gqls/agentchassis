# DESIGN — news editorial features: the starting point (2026-08-19)

**What this is.** The buildable statement of the editorial-features workstream,
consolidating five prior designs (A–F in the PLAN) and the live measurements into
one document: what we are building first, on which site, under what lifecycle
policy, and what stays out of scope until the owner lifts the pooling hold.
Companion to `PLAN_2026-08-19_news_editorial_features.md` (the full collection);
this file is the part you act on.

---

## 1. The instrument, in one paragraph

An **editorial feature** is a page about one of the bigger current news stories —
not a headline list. It is built by noticing that several channels are covering
the same story (the feed already tags 9,622 items with topics; nothing yet groups
them), extracting the story's **premises** — the claims it stands on — and filling
out the background with **code-rendered charts whose every figure resolves to a
cited, registered fact**. It stays up, gets updated while the story lives, and is
retired deliberately when it stops mattering.

## 2. What we are building first — one hand-built example

Per design D's recommendation (its steps 1–2, the series substrate and the
time-series renderer, already shipped): **step 3, one hand-authored feature page,
built through the framework's own components, before any lane is designed.**

**Site: robot-hands.com.** Chosen over the news exemplar (gaswholesalers.com)
because the chart components fail closed without registered facts and
gaswholesalers has no evidence base; robot-hands has one, 2,835 relevant feed
items across 9 sources, items from today, and a live story cluster that is
genuinely chartable.

**Story: the industrial-robot demand story.** The feed holds it from several
channels at once — *"Industrial Robot Installations Hit Record Highs Amid Labor
Shortage Crisis"* (2026-08-19), *"Robot Orders Increase in Q2 as Automation
Demand Broadens Across Industries"*, the Korea Q2 revenue coverage, two
machine-tending/bin-picking market-forecast items. That is the multi-channel
cluster the ask names.

**The premise (not the topic).** Applying D's test — *if this were false, would
the story change?* — the load-bearing claim across the cluster is: **factory
robot demand has stepped up structurally, not cyclically** — annual installations
have held above half a million units for four consecutive years where 2020's
figure was 384,000, and the industry body forecasts further growth. Everything
else in the cluster (orders up, vendor revenues up, labour-shortage framing)
leans on that. It is checkable against one authoritative series: the IFR's annual
World Robotics installation figures.

**The substrate, verified 2026-08-19.** Five observations, each from its own
year's IFR press release, every quote fetched and checked verbatim this session:

| year | global installations | source (IFR press release, dateline) |
|---|---|---|
| 2020 | 384,000 | "Robot Sales Rise Again" — Frankfurt, Oct 28 2021 |
| 2021 | 517,385 | "All-Time High…" — Frankfurt, Oct 13 2022 |
| 2022 | 553,052 | "World Robotics 2023… Asia ahead" — Frankfurt, Sep 26 2023 |
| 2023 | 541,302 | "Record of 4 Million Robots…" — Frankfurt, Sep 24 2024 |
| 2024 | 542,000 | "Global Robot Demand… Doubles" — Frankfurt, Sep 25 2025 |

Disclosure that goes in the chart footnote: figures are **as first reported** in
each year's release; IFR revises in later editions (the 2024 release restates
2022 as 552,946 against the 553,052 first reported). Mixing revision states
inside one series is the trap the Thames build documented; stating one
consistent basis and disclosing it is the remedy.

Supporting magnitude facts for a second chart and the prose (same verification):
2024 operational stock 4,664,000 (+9%); 2025 forecast 575,000 (+6%); regional
split Asia 74% / Europe 16% / Americas 9%; China 295,000, Japan 44,500.

**Page composition** (all existing components, nothing new built):

| slot | component | carries |
|---|---|---|
| hero | `hero` | headline + standfirst |
| feature-analysis | `generic-text-block` | the editorial: what happened, the premise, the labour-shortage framing, every figure resolving to a registered fact |
| evidence-timeseries | `evidence-timeseries` | the five-year installation series, per-point citations |
| evidence-chart | `evidence-chart` | 2024 magnitudes: China / Japan / rest-of-world against the global total |
| feature-coverage | `generic-text-block` | **the multi-channel part made visible**: the cluster's own articles, linked out with source names — title + link, never full text (the standing rights posture) |
| call-to-action | `call-to-action` | into MatchMatrix / the tools — the site's actual product |

Chart values are supplied as `content_data` copying the register exactly (no
display keys, no divergence), `rendered_html` produced by executing the live
component templates, rows **locked** so no model sits in the path — the Thames
pattern (`sql_for_agents/252`, `266`) verbatim. Claimscan runs before deploy.

**What the example is for.** Learning what the lane must automate. The bet,
stated up front so it can be falsified: the expensive step is **choosing the
story and its premises**, not generating the page.

## 3. Lifecycle policy — proposed for the owner's ratification

The owner's framing: *"possibly a long time, and for some pages every day or even
more often, for others weekly or monthly."* Proposal:

**Retention: pages stay up indefinitely, at one stable URL.** The July design's
SEO question answered the way its own analysis leans: updating one URL in place
accrues authority; a new URL per week fragments it. A feature carries a visible
"last updated" date. **Retirement is deliberate and is not deletion**: the page
leaves the nav and the feature index, gains `noindex` if judged stale enough to
harm, but the URL keeps serving — links out there keep working, and the platform
already has `pages.status='archived'` for exactly this. The retirement *trigger*
is the July design's open question answered operationally: when the story's
cluster stops accruing articles (measurable in the feed we already have), the
page stops earning updates; after a quiet quarter it retires.

**Update cadence is set per FACT, not per page.** This is the load-bearing
choice, and it is E's live mechanism applied to editorial. A feature page mixes
figures with different half-lives — a spot price moves daily, an annual series
moves once a year, a mechanism explainer barely moves at all. Pinning cadence to
the page forces everything to the fastest fact's rhythm and multiplies cost for
nothing. So:

| fact class | example on the worked page | refresh |
|---|---|---|
| fast series (prices, orders, live indicators) | *(none on this page — deliberately)* | daily or better, automated query where the source allows |
| annual/periodic series | IFR installations | on source publication (here: each September) |
| snapshot magnitudes | regional split, operational stock | with their source edition |
| framing prose | the analysis itself | only on a **narrative** change |

**Two update classes, priced differently** (from C, mechanised per E):
**substrate-only** — new observations land, charts re-render from the same
claims, prose untouched, near-zero cost, any frequency; **narrative** — the story
itself moved, prose regenerated/re-edited, costs money and a review, and needs
the blast-radius record (which derived pages cite this substrate). The
classification must be **compulsory**, per E's live precedent — an unclassified
change fails, or the stamp goes stale silently.

**Cadence tiers for future pages** (the owner's "some daily, some monthly"):
a **tracker** (fast facts, daily+ refresh, automated), a **feature** (weekly
while the cluster is active), an **explainer** (monthly or on-publication). The
tier is declared at page creation and is just the dominant fact class — no new
mechanism.

## 4. Pooling — what it is, and why it is on hold (context, not this workstream's decision)

Every site currently fetches and triages its own news. **Pooling** makes sites
that share a subject draw from one shared article pool — fetch once, enrich once,
rank per site — because the per-site design costs roughly 8,000 triage calls a
day at 2,000 sites. Seventeen pool sites exist (created 2026-07-20), deliberately
inert and invisible to every fleet loop, costing nothing.

**On hold by owner decision, 2026-07-20** (`features_open/005`), for three
reasons that are about order, not doubt: arming a pool means onboarding ~37
pilot domains onto a fleet of 11 (the biggest batch ever attempted); two known
rendering bugs gate it (`bugs_open/027` news renders nothing without JavaScript —
defeats the SEO purpose; `bugs_open/026` news listing hardcodes English — blocks
the non-English cohort); and the duplicate-content similarity gate
(`features_open/002`) does not exist yet, so every site built first adds to the
surface it must retro-check.

**What it gates here:** only the fleet-wide projection — shared substrate,
per-site angles, which also needs the per-site `audience` profiles that pooling's
onboarding writes. The single-site example in §2 needs none of it. When the hold
lifts, the substrate built for §2's page is the thing the angles project from —
which is why it must be registered facts, not prose.

## 5. Out of scope until the worked example has taught us

The grouping key (`duplicate_of`/`entity_ids` writers — remembering the render
path's cluster detector exists to *suppress*, so the new step must not touch that
call site); the premise-extraction agent; `topic_packages`/`event_timeline`
schema; any publishing agent; any auto-repair path (D's rule: this lane's output
enters human review, never auto-repair).
