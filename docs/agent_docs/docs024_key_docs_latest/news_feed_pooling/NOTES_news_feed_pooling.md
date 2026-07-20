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

### MISSTEP 4 — I overwrote the morning's SUMMARY instead of writing a new one

Asked for a summary per CLAUDE.md, I rewrote `SUMMARY_2026-07-19_news_feed_pooling.md`
in place, reasoning from the rule that *"SUMMARY is current state only (no
chronology — that's the other two)"* that the file should therefore be replaced
when the state moves.

That reading was wrong, and the owner corrected it: **each summary is a new file;
the series is the record.** Current-state-only describes what goes *inside* one
file, not whether the file is disposable. The concept register already worked this
way (`SUMMARY_where_we_are_2026-07-16/17/17b/18.md`) and I had that pattern in
front of me during the earlier survey without drawing the inference.

**Repaired:** recovered the morning version via `git show 9647cadf2^:<path>` and
restored it to its original filename; the evening version moved to
`SUMMARY_2026-07-19b_news_feed_pooling.md`. Nothing lost — but only because it had
been committed. Had I written both within one commit window it would have been
gone, which is the real lesson.

**Recorded in CLAUDE.md** as an explicit directive with the reason, since this is
shared practice and not specific to this workstream: overwriting destroys the
record of how the understanding moved, and *"the new one is better"* is not the
test, because the replacement here was accurate and it was still a loss.

---

## 2026-07-20 — session 2: the audience question, and a caution I overstated

### MISSTEP 5 — I recorded "explicitly reverted, find out why" without opening the revert block

In the cross-site survey I wrote that `identity.audience_primary` /
`audience_secondary` / `sophistication` had been *"designed and then explicitly
reverted"* in `003_site_classifier.sql`, and flagged it as a caution to resolve
before rebuilding the same shape. I took that from a subagent's report and cited
the file without reading the revert block.

**Read now, the revert's stated reason is mechanical:**

> `003_site_classifier.sql:170-174` — *"Revert: Undo 004 + 005 changes to existing
> agents … Restores original state for agents modified by the **incorrect 004/005
> scripts**. Safe to run even if 004/005 were never applied (idempotent)."*

It rolls the whole `site-classifier` row back to the original single-step Haiku
task. The rich profile — audience fields included — was **collateral in a wholesale
rollback of two bad scripts**, not a design anyone judged and rejected. There is
even an informally typed *"reset back to what it was"* above the block (`:167`).

And `site-classifier` is **legacy**. Live reference counts:
```
site-classifier refs: 1
domain-research-classifier refs: 4
```
The live intake chain is `domain-research-classifier` (049). So 003 is a superseded
agent, and its audience fields are **usable prior art, not a warning**.

**Cost of the error:** it was about to make me design around a constraint that does
not exist. **What caught it:** doing the thing I had written down as a prerequisite
instead of skipping to the design. The caution I wrote was the reason I found it —
which is an argument for writing cautions down even when they turn out hollow.

### The two live `audience` rows changed the schema

Pulled both in full. Neither is an audience *description* — both are audience-derived
**editorial directives**, written by `content-gap-planner` (a remediation agent):

- `ai-agent-orchestration.com`, key `primary_buyer_hierarchy`: after describing the
  CTO buyer it continues *"Lead all page copy with the technical failure mode …
  CTA language is 'Technical Discovery Call' sitewide."*
- `leopardessconsulting.co.uk`, key `target_audience`: *"… Non-technical SMB buyers
  are explicitly out of scope. All pages should be written for the primary audience
  by default … Pages that currently address non-technical readers (e.g. about page)
  should be revised."*

Two consequences, both recorded as Decision 9:

1. **The aspect must separate identity from directive.** A ranking layer cannot
   consume "CTA language is 'Technical Discovery Call' sitewide" — that is an
   instruction to a content agent. The embedding used for feed selection must be
   computed from the *who* + *position* blocks only. Including `editorial` would
   make two sites with similar copy rules rank alike regardless of audience, which
   is precisely the failure Decision 4 exists to prevent.
2. **The differing key names are a symptom, not sloppiness.** Neither row was
   written to a schema because the aspect has never had one. That is the work.

Also worth keeping: leopardess's `out_of_scope` phrasing is genuinely useful — a
negative audience constraint converts directly into a ranking penalty. Adopted into
the schema on the strength of one real row using it unprompted.

### Live state re-checked

| check | result |
|---|---|
| `domain-research-classifier` | active, `experimental`, updated 2026-07-20 |
| `site-classifier` | active, `experimental`, but 1 live ref — legacy |
| `content-gap-planner` | active, `active` — the only writer of `audience` today |

Note both classifiers show `updated_at` of today, likely from a fleet re-seed by
another session. Not chased; flagged in case a later thread finds classifier
behaviour changed under it.

---

## 2026-07-20 — session 2 continued: four owner directives executed

Owner: pilot traffic-bearing sites first; maybe a duplicate-content council seat;
split the big pools; migrate the two audience rows thoroughly.

### Audience migration — DONE, live

Pre-checks before writing: no active agent config references `specs.audience`
(regex over `agent_definitions.default_config`); the only Go reader is the dead
allowlist branch (`spec_admin_handlers.go:226`); `content-gap-planner`'s
`update_spec` route actually targets `identity.target_audience` — its two
`audience`-aspect rows were the LLM improvising an aspect name. So nothing
consumes the old shape; a future clobber would supersede, not destroy.

Executed as one transaction (supersede→insert, forced by
`idx_site_specs_current`). Verified:
- version chains intact — leopardess shows THREE rows (a March pair + today's),
  so the aspect's history predates us and survived;
- both current rows carry the four v1 keys, `sophistication='technical'`,
  leopardess `out_of_scope='Non-technical SMB buyers.'`;
- `position` null in both — the source prose contained none and none was
  invented;
- distinctive-phrase spot-check: `Technical Discovery Call`, `bridges both
  registers`, `production credibility`, `register consistency` all present.

Clause-by-clause mapping was done at write time (every sentence of both originals
assigned to who/editorial; nothing dropped, nothing added). SQL + gotchas now in
RUNBOOK.

### Pools v2 — 17 pools, coverage unchanged (63.8%)

Re-cut with uk-money split (insurance 83 / mortgages-lending 80 /
savings-investing 68) and marketing-web split (design-creative 126 / web-tech 54
/ marketing-digital 51). Full table in PLAN Decision 10.

**Classifier misfits caught in the pilot list** (regex crudeness biting exactly
where predicted): `whatvacancy.com` → travel via "vacan"; `greenpowerjuicers.com`
→ energy via "power". Hand-corrected in Decision 11; the caveat in RUNBOOK stands.

### Pilot cohort — 31 domains, ~4,300 views, 14/17 pools

Two findings worth the owner's attention (in PLAN Decision 11):
- 12 no-feed domains hold ≥25 views each, including the portfolio's #2
  (`smartbusinesssupplies.com` 748). Some are honestly excluded (zdec,
  komunikatif, makeitaquote — brandables/tools); `buysportskit` (215) and
  `smartbusinesssupplies` are arguable. Needs eyes, not a rule.
- The pilot is multilingual on day one (Romanian, Dutch, German, Spanish-market
  domains in the top 25), which **promotes `bugs_open/026` from fleet blocker to
  pilot blocker**.

### Council seat — specified, not seated

`features_open/004`. Deliberately not built: the seat has nothing to veto until
pooled selection exists, and it wants the 002 similarity check running first so
verdicts can cite a baseline instead of intuition. Footprint + four-form review
posture specified; mechanism is the standard 099 roster mirror.

---

## 2026-07-20 — session 2 continued: profiles authored for all 11 live sites; research in flight

Owner (same message): archive.org deep-research on the opaque traffic domains;
other-language domains are exceptions with per-language feeds (relojistas
precedent); real-traffic domains get dedicated feeds. Then mid-turn: "please go
ahead with the audience profiles."

### Deep research launched, running in background

Run `wf_2f8a91fd-b77` (task `w0i5fis9g`) over the 9 opaque/traffic no-feed
domains (zdec, komunikatif, makeitaquote, nanangmrk, bigotime, outfax, ijih,
smartbusinesssupplies, buysportskit) — prior content, active period, which legacy
URLs still attract the traffic, rebuild recommendation. Results to land in this
directory when complete.

### Scope call on "the audience profiles"

The 31 pilot domains have **no sites rows** — a profile cannot exist for them
(`site_specs.site_id NOT NULL`), and authoring from the domain name alone is the
Decision 4 trap. So "go ahead" was executed as: **the 11 live sites, now**, with
pilot-domain profiles deferred to onboarding where the classifier chain provides
a real derivation source. Recorded as Decision 15.

### Authoring — 9 new rows + 2 v2s, one transaction, verified

Derivation: `who` restructured verbatim from each site's `identity.target_audience`
(pulled fresh); `sophistication` judged from evidence; `out_of_scope` only where
the source stated a contrast (gaswholesalers); `editorial` null everywhere —
authoring copy directives is `content_direction`/`voice` ground, not this pass;
`position` **only for the AI trio** (finetuning / ai-agent-orchestration /
leopardess), the sole live sibling cluster, each position written against the
others' actual audience rows. relojistas' `who` kept in Spanish deliberately
(Decision 13 — the audience is Spanish-speaking; an English rewrite would be
translation loss in the embeddable field).

v2 mechanics for the two migrated rows: supersede → insert with
`data || '{"position": ...}'` — top-level jsonb merge, `who`/`editorial` carried
unchanged.

Verified: 11 current rows, all `audience.v1`, sophistication spread
4 technical / 2 professional / 3 casual / 2 editorial, position exactly on the
trio, 16 total rows (5 superseded versions preserved: 2 gap-planner originals,
2 migrated v1s, 1 pre-existing March leopardess).

### MISSTEP 6 (caught in seconds, worth one line) — wrong spot-check phrases

First v2 verification query tested leopardess's row for "Technical Discovery
Call" and "build-vs-buy" — **ai-agent-orchestration's phrases**. It returned
false and briefly looked like the `||` merge had dropped who/editorial. Re-ran
with leopardess's own distinctive clauses ("production credibility", "register
consistency") → both true, merge intact. The verification failure mode from the
guide — a check that greps the wrong thing — cuts both ways: it can pass wrongly
(generic CSS property) or fail wrongly (someone else's phrase). Match the probe
to the artefact.

### Live totals after this session

`site_specs` aspect `audience`: 11 current / 16 total. Every live site is now
rankable the moment pooled selection exists.

---

## 2026-07-20 — session 2 continued: domain-history research landed (partially), synthesised by hand

### Workflow outcome, honestly stated

`wf_2f8a91fd-b77`: 103 agents, 83 completed, 20 failed — **all 20 failures were
"session limit / resets 1:20pm"**, concentrated on the buysportskit and
smartbusinesssupplies *verification* votes and the final synthesis step. So the
workflow returned 14 verified claims + 6 unverified, unmerged. I synthesised the
report myself (`RESEARCH_2026-07-20_dormant_domain_history.md`), keeping
per-claim vote counts and marking every domain [verified]/[unverified]/[direct].
Two verifier notes also flagged that the safety classifier was unavailable for
two subagents — their outputs were re-read before use.

Gaps the workflow left (`zdec`, `ijih` — no claims at all) were filled by direct
CDX pulls. **WebFetch cannot reach web.archive.org** (hard refusal) — `curl` can;
recorded here because it will bite the next thread that tries.

### zdec decode — the GBK trap

The 2017 zdec.com snapshot 500s through `utf-8` decoding (`0xc9`). It is **GBK**
(Chinese). Decoded: an industrial control-systems company site whose footer is
stuffed with injected Macau casino spam + a literal "出售外链" (backlinks for
sale) contact. High views (409) ≈ poisoned link-farm residue, not demand.

### MISSTEP 7 — my bigotime "watch-adjacent" speculation was wrong

The relojistas audience-row note (written this morning) speculated bigotime.com
"may be watch-adjacent" from the name. Evidence: templated dropship-storefront
shell route (`/index/selectLogistic?coll_id=`), zero CDX captures, now an
Afternic for-sale redirect. Not watches. The DB row's note was hedged ("may be…
under deep-research — revisit"), so it self-corrects on read; resolution recorded
here and in the research file rather than churning the row. The lesson is the
name-derivation trap AGAIN, in miniature — I inferred content from a domain name
in the same session in which Decision 4 said not to.

### The two structural findings

1. **Ownership reconciliation needed on 4 of 9** (bigotime + buysportskit →
   Afternic for-sale; nanangmrk + ijih → live content operated by someone).
   Blocking for those domains only. Also: nanangmrk 403s non-browser agents —
   crawler-based "hosting nothing" measurements undercount Cloudflare-fronted
   sites, a fleet-wide measurement caveat.
2. **Views ≠ value**: the largest opaque number (zdec 409) is the least valuable
   traffic in the set.

### Outcomes folded into the plan

smartbusinesssupplies → business-services pool (reverses no-feed);
komunikatif → Indonesian-language exception with a real news legacy (strongest
news-shaped inheritance in the set); makeitaquote → tool-build (quote-image
generator; name-collision demand from a bot in ~1.12M Discord servers; sibling
of memecreator.co.uk — a real future `position` case); outfax → fax-tools theme;
buysportskit → sport-retail if ownership confirms; zdec/ijih/bigotime stay
no-feed. Impersonation lines drawn for the two domains shadowing real businesses.

---

## 2026-07-20 — session 2 close: ownership answered, Decisions 16–19

Owner resolved everything in one message: **all nine domains are his**, the
Afternic redirects are registrar parking (so the research's "conflicts with the
premise that we own it" claim was a wrong inference from true evidence — the
redirect is real, the conclusion "not ours" was not); **nanangmrk is owner-run**
(site AND the 551k-sub YouTube channel) and is to be **adopted into the
framework** — the relojistas class of work, flagged as its own future workstream,
check `adoption-pipeline.md` register state first.

New instrument recorded (Decision 17): **retailer-directory utilities** for the
commerce-legacy domains (buysportskit, smartbusinesssupplies, outfax) —
categorised product/service listings linking to real retailers, utility-first,
affiliate wrapping deferred. Neither a feed nor a brochure; prior art to check:
register `affiliate-commerce.md`, strategist `revenue_models`.

makeitaquote (Decision 18): build the tool differentiated — own branding, webby
feature set the bot doesn't have, explicit non-affiliation line. The domain name
describing the function is fine; passing-off risk lives in imitating identity.

Pilot now ~36 (Decision 19): + smartbusinesssupplies, makeitaquote, buysportskit,
nanangmrk (adoption track), outfax. komunikatif recommended, not yet directed.
zdec deliberately unstrategised — measure-first proposal recorded in PLAN (2–3
weeks of holding-page analytics to characterise the 409 views before any
investment; spam-era backlink profile may mean search distrust).

---

## 2026-07-20 — session 2 final: komunikatif YES; 17 pool sites CREATED

Owner: yes to komunikatif (Decision 20, pilot ~37); mid-turn: the zdec
measurement mechanism already exists — **the relojistas traffic-probe setup**
(VM box + nginx logs + CF real-ip). Folded into the PLAN proposal, carrying that
workstream's landmine: CF edge IPs make visitor counts impossible until real-ip
is configured — a measurement prerequisite, not a nice-to-have.

### Pool creation, gated on verification (Decision 21)

Checks run BEFORE the insert, in order:
1. `\d sites` — `status` is varchar(50), default 'active', **no CHECK
   constraints** (`pg_constraint contype='c'` → empty). New status value safe.
2. Grep for every `sites.status` predicate: fleet loops use `status='deployed'`
   only (`maintenance_actions.go:694,697`); all other status predicates are
   over pages/nav_items/companies.
3. Grep `status='system'` / `system.internal` / `eac60db8`: the synthetic site
   is referenced **only by hardcoded UUID** — no single-row-by-status assumption
   exists, so a second synthetic status value breaks nothing.
4. Feed-trigger arming condition re-read: needs current classification with
   `news_feed.recommended=true` + deployed page. Pool sites get neither.

Then: 17 × (site + pool-default audience row) in one idempotent transaction
(34 inserts, all `INSERT 0 1`). Verified after: status distribution
deployed 11 / pool 17 / system 1; 17/17 pools carry current `audience.v1`
rows; both safety invariants zero (no pool in the deployed predicate, no pool
with a classification spec).

Design choices worth remembering:
- **`status='pool'` over `'system'`** — not because 'system' breaks (it
  doesn't, verified), but because a distinct value is self-describing in any
  `GROUP BY status` and can never collide with a future single-system-site
  convention.
- **`settings.pool.slug`** so pools are machine-identifiable without parsing
  domains.
- Pool-default `who` texts are composites of the Decision 10 domain analysis —
  written to be forked, and each row's notes carry the TLIB-022 rule verbatim
  (site voice must FORK, never edit the shared base).
- **Ingestion left structurally off.** Sources per pool = real curation + real
  credits; that is its own deliberate step on ONE pool first, after pilot
  onboarding.
