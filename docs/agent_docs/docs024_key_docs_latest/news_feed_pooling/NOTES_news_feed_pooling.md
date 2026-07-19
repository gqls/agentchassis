# NOTES — news feed pooling

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-07-19 — session 1, architecture survey and domain analysis

### What was asked

Owner: rolling out thousands of domains, wants most to have a news feed without
paying per domain; asked for ~10–12 feeds covering the majority of verticals.
Domain list supplied mid-session (1,625 rows, abbreviated from a larger set).

### Survey of the existing feed pipeline

Two read-only Explore agents mapped code and docs. Established:

- All feed work is registered workflow actions in
  `platform/orchestration/actions/` (`feed_actions.go`, `dispatch_feed_sources_action.go`,
  `feed_triage_actions.go`, `render_news_section_action.go`, `render_rss_feed_action.go`).
  15 actions, all `Category: "feed"`, registered at `registry.go:1251-1374`.
- `content_sources.site_id` is `NOT NULL REFERENCES sites(id)` — a source is
  definitionally site-owned (`sql_for_tables/027_content_sources_table.sql:8`).
- Dedup is site-scoped in Go: `WHERE site_id = $1 AND source_url = $2`
  (`feed_actions.go:898`).
- `DispatchFeedSourcesAction` spawns one K8s job per source per site
  (`dispatch_feed_sources_action.go:145`).
- No sharing/pooling layer of any kind exists. The only cross-site sharing is at
  the *recipe* level — the `verticalNewsMap` Go table and a
  `seed_boxing_sources(p_site_id)` plpgsql function that copies hardcoded
  BBC/ESPN feeds into per-site rows.

### MISSTEP 1 — I reported "1,176 sites want a news feed". Wrong by ~250×.

I ran `count(*) FROM site_specs WHERE aspect='classification' AND
data->'content_features'->'news_feed'->>'recommended'='true'` and read the result
as a site count. It is a **row** count, and `site_specs` is versioned:

```
 total_specs | distinct_site_ids
        1187 |                11
```

The live trigger filters `ss.is_current = true`. With that filter the answer is
**4**, and the four are `ai-agent-orchestration.com`, `gaswholesalers.com`,
`relojistas.com`, `robot-hands.com`.

**What caught it:** the number was absurd on its face next to `count(*) FROM sites`
= 12. Checking two figures against each other is what exposed it — neither alone
looked wrong. This is the "ground every figure against the live system" rule
earning its place: I had carried a subagent's DDL reading straight into a claim.

### MISSTEP 2 — I nearly reported the `LIMIT 5` throttle as unverified/absent.

I quoted `LIMIT 5` from `sql_for_agents/090_b_content_feed_trigger.sql:41` in my
first reply without checking it was live — and the Explore agent had explicitly
warned those SQL files contain repeated, drifting copies of the same DDL. When I
went to verify:

```
SELECT type, is_active, status, (task_workflow IS NOT NULL) t,
       (orchestrator_workflow IS NOT NULL) o, (orchestration_workflow IS NOT NULL) oo
  FROM agent_definitions WHERE type='content-feed-trigger';
-> content-feed-trigger|t|experimental|f|f|f
```

All three workflow columns NULL. For a few minutes I believed the trigger had no
live workflow at all. It does — it is in `default_config->'workflow'`. The
`LIMIT 5` **is live**, verbatim, along with the `is_current = true` filter that
exposed misstep 1.

**Lesson recorded in RUNBOOK:** for this table, absence from the workflow columns
is not absence of a workflow.

### Live state, verified 2026-07-19

| fact | value |
|---|---|
| `sites` | 12 |
| sites with a current spec wanting news | 4 |
| sites with any `content_sources` | 4 |
| `content_sources` rows | 24 (13 news_search, 6 rss, 4 api_news, 1 scrape) |
| `content_feed_items` rows | 6,564 across 4 sites |
| `content_feed_items.site_id` | **nullable** — pooling slot already exists |
| `idx_cfi_dedup` | `btree (source_url) WHERE status <> ALL(...)` — **non-unique** |
| pgvector | `vector 0.8.0` installed |
| embeddings | Ollama `nomic-embed-text`, 768-dim, ivfflat cosine (`rag_actions.go:282`) |

So the feed pipeline is a **4-site prototype**. There is no scaling wall being hit
today; the wall is entirely prospective. That is good — nothing to migrate, and the
pooled model can be the default before the fleet exists rather than a retrofit.

### Domain analysis method

1,625 rows written to scratch as `domain<TAB>views`. Stripped TLD (handling
`co.uk`/`org.uk`/`me.uk` as compound suffixes), removed hyphens, then an **ordered
regex classifier**, first rule wins, `misc-brandable` as the catch-all.

Results:

| theme | domains | % | views |
|---|---|---|---|
| misc-brandable | 339 | 20.9% | 1520 |
| retail-consumer | 220 | 13.5% | 1239 |
| marketing-web | 218 | 13.4% | 239 |
| money-personal | 97 | 6.0% | 248 |
| insurance | 81 | 5.0% | 34 |
| construction-trades | 79 | 4.9% | 268 |
| industrial-plant | 71 | 4.4% | 84 |
| travel-leisure | 69 | 4.2% | 2246 |
| ai-tech | 66 | 4.1% | 824 |
| vehicles | 65 | 4.0% | 60 |
| health-medical | 63 | 3.9% | 151 |
| business-services | 54 | 3.3% | 141 |
| investing-markets | 42 | 2.6% | 128 |
| property | 39 | 2.4% | 178 |
| energy-utilities | 39 | 2.4% | 17 |
| vet-animal | 34 | 2.1% | 51 |
| jobs-recruit | 20 | 1.2% | 10 |
| sport-events | 14 | 0.9% | 19 |
| pensions-tax | 11 | 0.7% | 1 |
| legal | 4 | 0.2% | 0 |

**Classifier caveat, unresolved:** substring matching produces false positives that
would matter if any single bucket were load-bearing — `carbondioxide` matched
`bond` into investing-markets, `childrensportraits` matched `sport`,
`book-air-taxi` matched `tax`. The aggregate shape is sound; individual counts are
±a few. Anchor the short patterns before using these numbers for anything
narrower than pool sizing.

### Traffic shape

- 352 of 1,625 domains (21.7%) have any views at all.
- Total 7,458 views across the whole list.
- `wayfaringlondoner.com` alone is 1,997 — **27% of all portfolio traffic**.
- Next: smartbusinesssupplies 748, traderboltai 655, zdec 409.

So this is overwhelmingly a portfolio of *unlaunched or dormant* domains. Any
per-domain cost is therefore pure speculative outlay against zero current traffic
— which sharpens the case for pooling considerably, and also argues for launching
feeds on the traffic-bearing tail first.

### Near-duplicate families

Collapsing plural/TLD/prefix variants: **146 concept families cover 358 domains**.
Largest: `insurance` ×11, `landlordinsurance` ×10, `healthinsurance` ×9,
`deliverydriverjob` ×5, `makeupdesign` ×5. Whether these are distinct sites or
redirects is an open question for the owner — it decides whether we need
intra-family feed divergence or no feed for the duplicates.

### Prior art checked

- No existing pooling work in `docs024_key_docs_latest/` (grep: `feed pool|pooled
  feed|shared feed|news pool` → nothing).
- Nothing relevant in the `needs_diagnosis` queue.
- Two open news bugs: `bugs_open/026` (news-listing hardcodes English, drops a
  required h1) and `bugs_open/027` (**news pages render no news without
  JavaScript**). 027 is material to this workstream — if the feed is the SEO
  instrument, a JS-only render defeats its purpose fleet-wide.
- `docs026_concept_register/register/news-feed-pipeline.md` records NEWS-004
  (per-source interleaving in the render) as **downgraded to aspirational on
  verification** — `loadNewsItems` has no `PARTITION BY`. Diversity constraints
  are to be built, not inherited.

---

## 2026-07-19 — session 1 continued: owner decisions, `features_open/` created

### Owner resolved all three open questions

1. **Duplicate families are separate sites**, each aimed at a different target
   market, so news selection can be angled per domain.
2. **Launch order follows traffic.**
3. **Paid tier: yes, but it needs to be more than news to be a product.**

Recorded as Decisions 4–6 in PLAN.

### Consequence I did not anticipate when proposing the design

Decision 1 (shared pool, per-site ranking) was framed as a *cost* argument. Under
Decision 4 the per-site ranking layer becomes **load-bearing product behaviour**:
it is the only thing distinguishing `bestinsurancerate.co.uk` from
`bestinsurancerate.uk`. That promotes a question I had not treated as urgent —
**where per-domain target-market profiles come from**. Deriving the profile from
the domain name is the obvious approach and is exactly wrong here: the names are
near-identical in precisely the families we most need to separate, so a
name-derived profile guarantees identical ranking. Logged as open question 3.

### Where to file the new material — options considered

Owner asked me to search and suggest rather than assume. Checked:

- **`bugs_open/`** — its README scopes it to "what is biting production right
  now"; `bugs_closed/README.md` reinforces that the bar is about current
  reproducibility in prod. A latent defect in an unbuilt design fails that bar,
  and filing it there degrades the one question the directory answers. **Rejected.**
- **`docs026_concept_register/`** — a *derived* extraction swept from `docs/`
  ("Nothing outside this directory is modified by this work"), status-tagged from
  documentary signals. Hand-authoring forward-looking entries is off-pattern.
  **Rejected.**
- **The workstream PLAN** — already carries the risk, but is workstream-scoped;
  concurrent threads will not read it. Insufficient alone. **Kept, plus a
  root-level home.**
- **New `features_open/`** — mirrors `bugs_open/`/`bugs_closed/`, root-level and
  discoverable, own independent numbering sequence. **Chosen.** Suggested to the
  owner with the reasoning rather than presented as done.

Created with a README stating the bar explicitly, and the `FEATURE_` vs `RISK_`
distinction — a latent defect in an unbuilt design is a RISK here, and *becomes*
a bug (and moves) the day the design ships and the defect is reproducible.

### Substantive addition to the owner's packaged-features idea

The owner proposed weekly topic packages as the duplicate-content mitigation. Noted
in `001` that the **naive version makes the problem worse, not better**: one
package syndicated to 231 domains is long-form near-identical prose, the most
heavily penalised duplication shape — strictly worse than a duplicated headline
list. The fix is the same shared-substrate/per-site-projection split that makes
pooled feeds work, applied one level up: research once (expensive, shared), angle
per site (cheap, differentiating). This also satisfies Decision 4 directly.

Cost check on that: one package/week × ~231 money-pool domains = ~231 generations
/week. Against the naive per-site feed design's projected ~8,000 triage calls/day
at 2,000 sites, this is affordable and spent on the differentiating layer rather
than the commodity one.

### Not touched

`bugs_open/027` (news renders nothing without JavaScript) — **another thread is on
it**, per the owner. Left alone; still noted in PLAN as a rollout blocker because
it gates fleet rollout regardless of who fixes it.

---

## 2026-07-19 — session 1 part 3: cross-site machinery survey

Owner: *"look hard in the docs to see how we handle cross-site/multi-site
decisions and we can hopefully piggy back on that work."* Two Explore agents over
`docs/`, `platform/`, `internal/`, `pkg/`; findings then verified against the live
DB where they were load-bearing.

### MISSTEP 3 — my own Decision 1 implementation sketch ran against the platform's idiom

In my first reply I proposed adding `pool_id` to `content_sources` and
`content_feed_items`, making `content_sources.site_id` nullable, and treating
`content_feed_items.site_id`'s existing nullability as "the pooling slot that
already exists". I described that nullability as a structural hint that pooling
was anticipated.

**It is not the platform's idiom, and I proposed it before looking.** The
established pattern for work with no owning customer site is a **synthetic site
row**. `system.internal` exists precisely because `site_work_items.site_id` is
`NOT NULL` and the platform chose a synthetic site *"rather than inventing a
null-site mechanism"* (DBI-010). TLIB-018 makes the intent explicit: a synthetic
site record owns library-level work *"so the ordinary site_work_items/dispatch
machinery can operate on the shared component library exactly as it does on a
customer site."*

Verified live:
```
SELECT id, domain, status FROM sites WHERE domain='system.internal';
 eac60db8-b032-432b-b36d-76f37632045d | system.internal | system
```
Matches `triageSystemSiteID` at `diagnose_triage_action.go:41`.

**Making each pool a synthetic site** satisfies `content_sources.site_id NOT NULL`
with no schema change, leaves every ingest action and the dedup index untouched,
and removes the `idx_cfi_dedup` lockstep hazard entirely — the class that already
caused a fleet-wide 42P10 in this repo. Recorded as Decision 8 with a visible
CORRECTION block in PLAN.

**What caught it:** the survey the owner asked for. Nothing about my original
sketch was self-evidently wrong; it was locally coherent and I had a plausible
story for it ("the nullable column is the anticipated slot"). It was wrong because
I had not looked at how the platform already solves this shape of problem. Same
failure mode CLAUDE.md's diagnosis-loop correction describes — not missing
information, just not looking.

### The audience question — verified live, better news than expected

| check | result |
|---|---|
| `site_specs` constraints | only `site_specs_pkey` + `site_specs_site_id_fkey` — **no aspect constraint**. Free-form; new aspect = no migration. |
| distinct current aspects | **38** |
| `audience` aspect | **already exists**, 2 current rows (`ai-agent-orchestration.com`, `leopardessconsulting.co.uk`), `source_agent = content-gap-planner` |
| `identity.target_audience` | populated for **all 11** sites with rich prose |

The two existing `audience` rows have **inconsistent key shapes** —
`primary_buyer_hierarchy` in one, `target_audience` in the other. Written ad hoc
by an agent that was not designed to own this aspect. So the aspect exists but has
no settled schema; that is the work, not creating it.

Also found: `internal/core-manager/admin/spec_admin_handlers.go:226` already
branches on `aspect == "audience"` — dead anticipatory code for an aspect nothing
officially writes.

**Aspect sprawl is real and this decision can worsen it.** 38 aspects, many
singletons, with already-overlapping families: `voice` / `voice_and_tone` /
`voice_and_audience`; `content_direction` / `content_standards` / `content_rules`.
Adding a 39th ad-hoc shape is the wrong move; settling `audience` is the right one.

### Cross-site machinery: what is actually live

Verified BUILT (subagent read the source; statuses cross-checked against the
concept register):

| mechanism | anchor | use to us |
|---|---|---|
| shared component library, `forked_from IS NULL` | `component_library.go:176`, `deploy_tool_action.go:10-13`, TLIB-022 | **the share/fork rule for audience profiles and package substrates** |
| field-set guard on shared-base regen | `store_generated_component_action.go:331` | additions ok, renames/drops rejected — mirror for profile changes |
| blast radius counted + recorded before shared mutation | `fix_component_template_action.go:411-433` | substrate updates fan out to N angles; record it |
| `system.internal` synthetic site | `diagnose_triage_action.go:41`, DBI-010/TLIB-018 | **pools are synthetic sites** (Decision 8) |
| `js_snippets.applies_to` declarative targeting | `render_js_snippets_for_site_action.go:150-159` | site→pool binding without a join table |
| cross-site duplicate-palette check | `check_duplicate_palette.go:69-83` | **the template for the duplicate-content check in `features_open/002`** |
| fleet scan over deployed sites | `maintenance_actions.go:694-698` | the only whole-fleet loop |
| `count(DISTINCT site_id)` fleet-pattern signal | `diagnose_triage_action.go:331,359` | how a finding becomes fleet-wide |

Verified NOT usable:
- **`networks`** — a dead FK. One hardcoded `default` row auto-created at
  `site_db_actions.go:983`; `networks.settings` never read by any Go code; zero
  references in core-manager or frontends. DBI-007 status `superseded`. It is a
  clean empty hook, nothing more.
- **`vertical_registry`** — 0 hits in any `.go` or `.sql`. VKA-001..004 all
  `aspirational`. (Do not confuse with `business_intel.business_verticals`, which
  is real but scopes data-collection agents, not sites.)
- **site groups / cohorts / portfolios** — PEV-001/002/003 all `abandoned`.
  PEV-001 is notable for taking an explicit *anti*-fleet-wide-decision stance.

### The genuine gap

Nothing records *"this site is positioned differently from our **other** site"*.
Every differentiation field is against external competitors
(`strategy.competitive_position`, `identity.unique_selling_points`,
`identity.competitors_found`). Intra-portfolio positioning has no home anywhere in
the platform. That is the one new thing Decision 4 actually requires.

### Two cautions carried into PLAN

- `identity.audience_primary`/`audience_secondary`/`sophistication` was designed
  and **explicitly reverted** in `003_site_classifier.sql`. Reason unknown —
  **find out before rebuilding the same shape.** Not treated as a blocker, but it
  is exactly the kind of prior art that turns into a rediscovered problem.
- The audience question was **dropped** from the live briefing questionnaire
  (`026_pageflow_builder.sql:868` has no audience section; the backups had a
  required one). We have been losing audience capture over time.

### Also relevant, not chased

`docs024_key_docs_latest/021_site_spec_and_classifier.md:17` records a live bug:
two parallel paths to `target_audience` (the `site_specs` path and the
`render_context` path), causing the content-quality-auditor to report "no target
audience defined" when the data is in `content_data.response`. Any profile work
must pick one path deliberately rather than adding a third.
