# NOTES — news editorial features

Running record, append-only, newest at the bottom. Technical log: evidence,
commands, what the system actually said, and every misstep.

---

## 2026-08-19 — session 1: collection pass

Owner's ask, verbatim in substance: *"editorial type feature articles relating to
current bigger news articles, filling the research with graphs and charts of
background information, and extracting the concepts out of the various news
articles from different channels to do this."* Stage 1 = collect the past
discussions into one thread. Nothing to be built this session.

### How the search was run

`grep -rli` over `docs/`, `bugs_open/`, `bugs_closed/`, `features_open/` for
`editorial`, `feature article`, `long-form`, `infographic`, `explainer`, `chart`,
`graph`. Then the concept register's own category list, which is the index that
answers "what exists and is callable" — four categories were directly on point
(`news-feed-pipeline`, `visualisation-and-charts`, `data-charts`,
`topic-intelligence`) and three more adjacent (`research-agents`,
`flows-and-narrative`, `content-quality`).

**The register's freeze caveat bit immediately and in the useful direction.**
`news-feed-pipeline.md` and `data-charts.md` both carry `covers-through:
2026-07-13`, but `visualisation-and-charts.md` carries `covers-through:
2026-07-28` and says in its own header that it exists *because* a 2026-07-27
handoff asserted in bold "there is no chart renderer" while two renderers were
live. That warning is the reason this pass checked the live `content_components`
table rather than trusting either the register or a Go grep.

### What is actually live, measured today

All figures read from `postgres-clients-0` on 2026-08-19, not carried from a doc.

`content_feed_items` population **[MEASURED 2026-08-19]**:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE topics IS NOT NULL AND topics <> '[]'::jsonb) AS with_topics,
       count(*) FILTER (WHERE entity_ids IS NOT NULL AND array_length(entity_ids,1)>0) AS with_entities,
       count(*) FILTER (WHERE published_page_id IS NOT NULL) AS published,
       count(*) FILTER (WHERE duplicate_of IS NOT NULL) AS dupes
  FROM content_feed_items;
--  total | with_topics | with_entities | published | dupes
--  10855 |        9622 |             0 |         0 |     0
```

Statuses: `rejected` 4977, `expired` 3106, `relevant` 2319, `review` 316,
`ingested` 137.

Sources: `api_news` 7 rows / 7 sites, `news_search` 31 / 7, `rss` 10 / 4,
`scrape` 1 / 1.

**The three zeros are the finding, and they are structural, not empty-by-chance.**
`grep -rn "duplicate_of" --include=*.go platform/ internal/ pkg/` returns
**nothing**; so does the same grep for `entity_ids`. Both columns are declared in
the schema and written by no code path in the repo. `published_page_id` is
likewise never set, because the agent that would set it does not exist (below).

Concept extraction, by contrast, **is already working**. `topics` is written at
`platform/orchestration/actions/feed_triage_actions.go:245`, parsed at :204 from
the `feed-triage` LLM's JSON. Live sample:

```
["darts", "Wayne Mardle", "Hawaii 501", "player profile"]
["darts", "PDC", "Barry Hearn", "tournament withdrawals", "player conduct"]
["industrial robots", "factory automation", "labor shortage", "robotics investment"]
```

That is per-item free text with no controlled vocabulary, no entity resolution
and no join key — 9,622 items' worth of concepts that nothing groups.

Absences, checked **with a positive control in the same query** so the empty
result is evidence rather than a broken filter:

```sql
SELECT type FROM agent_definitions WHERE type IN
 ('article-rewriter','feed-publisher','feed-lifecycle','news-analyst',
  'story-researcher','analysis-writer','visualization-renderer','data-chart-generator');
-- (0 rows)
SELECT type FROM agent_definitions WHERE type='feed-triage' LIMIT 1;
-- feed-triage          <- the control: the query shape works
SELECT to_regclass('public.event_timeline') IS NOT NULL,
       to_regclass('public.topic_packages') IS NOT NULL;
-- f | f
```

So every agent named in the 2026-04 Tier-3 design, and the `event_timeline` table
that design called "the core enabler", are all absent.

What **is** live and reusable:

| thing | state today |
|---|---|
| `evidence-chart`, `evidence-timeseries`, `mechanism-flow` | active section components |
| `Latest News Feed`, `news-listing` | active section components |
| `research_results` | 177 rows, 8 sites, latest 2026-08-11 |
| `site_specs` aspect `evidence_base` | 17 current rows, 17 sites |
| 17 news pools | `sites.status='pool'`, all dormant |

The `evidence_base` figure is worth flagging against the docs: `features_open/023`
measured **8 sites** on 2026-07-25 and reasoned from that scarcity. It is 17
today. Anyone re-reading 023 must re-measure before repeating its "8 of the
fleet" framing.

### The diagnosis loop run

Fired `090` on the one genuinely non-obvious mechanism question, since the claim
this workstream will build on is structural and cross-cutting — exactly the case
CLAUDE.md says to file before asserting.

- intake correlation `6ec5ae74-eb3d-4db9-83e7-ddc28fd7e95c`
- **run correlation `3802fb10-c34f-4eff-9914-b2959c723bd5`** (the key artifacts
  are written under — the intake id is not)
- symptom: items enriched one at a time and never grouped across sources;
  `duplicate_of`/`entity_ids` written by no code path while `topics` is
  populated; asked which step was intended to write them, whether any
  cross-source grouping key exists, and what the render path can select on.

Verdict recorded below when it lands.

### MISSTEP 1 — nearly trusted the register's status line

`data-charts.md` CHRT-001 says the chart capability is **aspirational** ("L7
Charts — not started"). Read alone that would have gone into the plan as "charts
must be built". It is stale by construction: the entry covers through
2026-07-13 and `evidence-chart` shipped after it. `visualisation-and-charts.md`
— written 2026-07-28 from first-hand reads — says the opposite, and the live
`content_components` table agrees with the newer file. Two register entries on
the same subject disagreeing is not a contradiction to resolve by reading harder;
it is a dated snapshot next to a fresher one, and the live table settles it.

The landmine text for exactly this ("a concept-register STATUS line is a snapshot
that outlives its truth") fired at session start against a dirty file in this
tree. It was right.

---

## 2026-08-19 — the diagnosis verdict: CONFIRMED, and it found a file this pass missed

Run correlation `3802fb10-c34f-4eff-9914-b2959c723bd5`, completed 10:12:53Z,
three iterations, `stopped_by: confirmed`.

**Outcome: CONFIRMED**, with all three symptom clauses marked `explained` and
cited:

| clause | how it was grounded |
|---|---|
| items enriched one at a time, never grouped across sources | `WriteFeedItemsAction` inserts each item individually with no grouping key; `QueryNewsItems` selects them individually |
| `duplicate_of` / `entity_ids` written by no code path | every write in scope — the `WriteFeedItemsAction` INSERT, `ApplyFeedScoresAction`'s UPDATE, `RenderNewsSectionAction`'s expire UPDATE — omits both from its column list, and the live count (10855 / 0 / 0) confirms it holds in production, not only in the code as written |
| `topics` populated by triage | `ApplyFeedScoresAction` sets `topics = $3`; `llm_call_log` shows `feed-triage/score_relevance` succeeding at 2026-08-19T08:42:43Z, i.e. the mechanism is actually running |

**Note the shape of that second row.** The loop did not stop at "no code writes
it" — it paired the static absence with a live count, because code-as-written and
production can disagree. That is the pairing this workstream's own claims should
copy.

### The find: `queryresolve/news_items.go`, which this pass never opened

The loop's scope walk surfaced `platform/orchestration/actions/queryresolve/news_items.go`
(421 lines) — `QueryNewsItems`, `resolveLatestNews`, `resolveNewsArchive`,
`newsTopicalTokens`, `titleToolKeys`, `capNewsItemsPerTool`, `projectNewsItems`.
I read it first-hand rather than trusting the report.

**There IS a mechanism that notices several headlines are about the same thing —
and it exists to throw that signal away.** `newsTopicalTokens` (:185) builds a
topical vocabulary from the site's own source queries, and *additionally* derives
one by document frequency over the item titles when the pool is at least 12
items, at a threshold of `len(items)/4` (floor 5). `capNewsItemsPerTool` (:222)
then walks the ranked items and **drops** any whose title shares a tool key with
`maxPer` already-kept items.

Its own comment is the sharpest statement of this workstream's central problem
that exists anywhere in the repo, and it was written for the opposite purpose:

> Frequency derivation needs a pool large enough that "appears a lot" means "is
> the subject matter", not "is one well-covered story". Below 12 items a genuine
> tool cluster (four Firefox headlines) would cross any workable threshold, so
> small pools rely on query topics alone.

**"Four Firefox headlines" is precisely the input an editorial feature wants and
this code deletes.** The display layer's requirement (no single story dominates
the feed) and the editorial layer's requirement (find the story several channels
are covering at once) read the same signal in opposite directions. Any grouping
step built here must not be bolted onto this path, because this path's contract
is suppression.

### Second find: what "with dedup" actually means

`registry.go:1386` describes `WriteFeedItemsAction` as *"Normalise and write feed
items to content_feed_items with dedup"*, which reads like cross-source
de-duplication and is not. `feed_actions.go:777-778, 896-905` — dedup **skips
items whose `source_url` already exists for this site**, via the `idx_cfi_dedup`
partial index on `source_url`. So it is exact-URL, per-site. The same story
reported by Reuters and the BBC is two rows by design; a syndicated copy at a
different URL is also two rows. **Cross-source grouping is not merely unbuilt —
the one thing named "dedup" in this pipeline solves a different problem**, and a
reader skimming the registry description would reasonably conclude otherwise.

### Third: what the render path can select on

`QueryNewsItems` filters `WHERE cfi.site_id = $1 AND cfi.status IN ('relevant',
'ingested')` and projects title, summary, url, `source_published_at`, source
name and `topics`. So `topics` **is** available to the render path, and
`duplicate_of` / `entity_ids` are not selected at all. If a grouping key is
added, the render path needs a column change, not just a writer.

### A harness defect worth knowing about, since it will bite the next run

Iteration 1's `NeededEvidence` records that the requested `SEED_SCOPE` **did not
arrive** — the bundle carried an unrelated symbol-search fallback
(`intake_repo.go`, `plan_sections_action.go`, `contentcreator/agent.go`,
`evolution.go`) instead of the feed files actually named, and the attached
runtime evidence was `agent_error_log` rows about the page-build pipeline, which
bears on nothing here. The loop **noticed and said so** rather than reasoning
from the wrong files, then re-scoped and got the right ones by iteration 2. That
is `bugs_open/174`, and the loop names it itself.

The lesson for reading any future verdict from this lane: **check
`evidence_trail[].Verdict.NeededEvidence` before trusting an early iteration** —
a bundle that did not deliver the requested scope still produces a
confident-looking iteration.

### What this changes in the PLAN

`PLAN_2026-08-19_news_editorial_features.md` §3 said the "different channels"
half of the ask "has no mechanism at all". **Corrected**: there is no *grouping*
mechanism, but there is a *cluster-detection* mechanism pointed the other way.
That is better news than "nothing exists" — the detection heuristic has already
been thought about, tuned and shipped — and worse news than it looks, because
reusing it means inverting a contract another feature depends on. §3 amended.

---

## 2026-08-19 — second sweep: two more threads, and the newest one is the closest

The first sweep's grep terms (`editorial`, `feature article`, `infographic`,
`chart`) missed two documents, both found on a second pass using `deep dive`,
`topic package`, `living dossier`, `insights section`. **Recording the miss
because the terms that found them are the terms a future sweep should start
with** — the closest prior art called itself neither "editorial" nor "feature".

### Thread D — `docs024_key_docs_latest/oufe/DESIGN_2026-07-28_premise_branching_and_deepthink.md`

The most recent, the most specific, and it states in its own opening that the
owner raised it as generic for *"this type of site and similar to the news
editorial requirement"*. So it is not adjacent prior art — it is this ask,
recorded three weeks ago against a different site.

Three things in it that nothing else has:

**1. The premise/topic distinction, which is the sharpest design primitive found
anywhere in this collection.**

> *Topic:* "the stakeholders in the Thames Water restructuring." Produces an
> encyclopaedia page. Nobody needs it; it is a worse Wikipedia.
> *Premise:* "the outcome turns on whether the Class A group's valuation of the
> relevant alternative is the one the court accepts." Produces a page with a
> reason to exist, and it tells you **what tool and what graph belong on it** —
> because a premise is a claim that can be tested.

The test is mechanical: **if this turned out to be false, would the main article
change?** If not it is background and belongs in a sentence, not a page. That
also answers "which one or two points do we branch on" by ranking rather than by
taste.

And the extractor must be able to **decline** a premise with a reason as a
first-class output — the worked example is "competitors" for a regulated regional
monopoly, where a generic extractor will happily invent a competitive landscape.

**2. It identifies 001 and the branch-page request as ONE feature with two
projections**, which is the observation that stops us building the substrate
twice:

| | `features_open/001` (across sites) | the branch-page request (across pages) |
|---|---|---|
| built once, expensive | research substrate | research substrate |
| generated many times, cheap | one angle per domain, from `audience` | one branch page per premise |
| what varies | who is reading | which question is being answered |

The owner's ask today wants **both** projections at once — many channels' articles
into one feature (001's axis) and that feature filled out with background graphs
(the branch axis). So neither prior design covers it alone, and the shared
substrate is the thing both agree must not fork.

**3. A hard rule for automation, from `bugs_open/126`:** an artefact generated by
this lane on an evidence-gated site **enters human review; it never enters
auto-repair.** The cited defect is a failing Tier-4 tool acceptance auto-raising
an `improve_tool` item that carried the failing criteria as the spec — on the
waterfall tool the only way to satisfy it was to delete a legally load-bearing
consent gate, and it was cancelled by hand. That is survivable while tools are
built one at a time with a human watching; a lane generating per premise, per
page, per site removes the human from exactly that position.

Also a warning worth carrying: **branching one good page into six thin ones is a
reliable way to dilute a site.** The SEO question in 001's open list bites this
projection hardest.

### The part of that design that is now DONE — verified, not assumed

The note's §3 said the real blocker was the substrate, not the rendering, and
forced an ordering: **series-shaped facts first, renderer second.** Its suggested
first slice was four steps.

Checked live today:

```
platform/orchestration/datahelpers/claims_series.go — exists, 15 Observation refs
content_components: evidence-timeseries — active, section level
site_specs aspect=evidence_base, is_current, data::text LIKE '%observations%' — 1
  (control: 17 current evidence_base specs in total)
```

So **steps 1 and 2 of that recommendation shipped**, the day after the note was
written (VIZ-002/VIZ-003 date them 2026-07-29), and one site has registered a real
series. The recommendation now sits at **step 3: one hand-authored branch page on
the strongest premise**, then step 4, design the lane against a real example.

That is the single most useful thing this second sweep produced — a three-week-old
design whose first two blocking prerequisites are already cleared, which nothing
in the first sweep would have told us.

### Thread E — `docs024_key_docs_latest/experience_register/COORDINATION_2026-07-28_packaged_topic_features.md`

An advisory coordination note written the same day, explicitly disclaiming
ownership of 001. Two things it contributes:

**001's hardest stated problem is already solved and live somewhere else.** 001
names its own worst risk twice — when the substrate updates, every derived angle
is potentially stale, and that fan-out must be counted at update time rather than
discovered later. `write_experience_pattern_action.go` implements exactly that
shape: every column is classified **contract / selection / cosmetic / system**; a
change to a *contract* field demotes an approved entry to `draft` and logs which
fields changed; a *cosmetic* change does not; comparison is on canonical JSON so
key order cannot cause a spurious demotion. And the classification is
**compulsory** — a column classified into none of the four **fails the build**
(`TestExperiencePatternColumns_EveryColumnIsClassified`).

001's "substrate-only vs narrative" is that classification applied to facts
instead of clauses. **Copy the mechanism rather than re-deriving it**, and copy
the compulsory part especially: without it the list silently goes stale and a
changed fact keeps a stamp saying the angle was checked against it.

**And a seam that must not be crossed:** the experience register holds
*behaviour*; the substrate holds *editorial content*. Putting prose in
`experience_patterns` would break the property the design rests on — a base entry
is site-agnostic and carries no site-specific values (`bugs_closed/045`). An angle
is site-specific by definition. *"A dossier page's behaviour comes from the
register; its content does not."*

Note the coordination file closes with its own `[UNVERIFIED]` marker: its author
had not read the pooling docs or `features_open/002`. That marker is doing real
work — it is why this lane read 001 and 002 directly rather than through it.

### Thread F — NEWS-019, the earliest ancestor

`register/news-feed-pipeline.md` NEWS-019 (status **superseded**) records a
mid-era design in `docs017_legacy_agent_rules_images_design_keydocs/` with a
**`deduplicator` sub-agent doing near-duplicate headline detection**, feeding
triage, then an `article-rewriter` producing original articles with entity
cross-links, then time-based lifecycle decay (featured → current → aging →
archive → prune). So cross-source grouping was in the design from the very
beginning and was dropped somewhere between that design and the deployed
descendant — worth knowing before treating it as a novel addition.

---

## 2026-08-19 — session 2: the hand-built worked example (design D's step 3)

Owner approved: summary doc, hand-built example on a site of my choice, propose
the lifecycle policy, explain pooling. `DESIGN_2026-08-19_starting_point.md`
carries all four; this entry is the build log.

### Site and story selection

**robot-hands.com** over gaswholesalers (the news exemplar): gaswholesalers has
**no evidence base** and the chart components fail closed — it would have
rendered a feature with no charts, which is the whole point missed.
robot-hands: evidence base (extended 2026-08-19 by another lane to 8 catalogue
facts), 2,835 relevant items, 9 sources, items from today.

**Story cluster, from the live feed**: the robot-demand story is on at least
four channels this week — investingnews (installations/labour-shortage),
BusinessWire's A3 orders release **carried near-verbatim by theaiinsider and
roboticsandautomationnews** (the same story on three channels — the multi-channel
signal made concrete), seekingalpha (earnings framing), and **scdigest carrying
the divergent read** ("US Robot Orders Weak in Q2 … while Revenues Much
Stronger"). The divergence is not noise — it is design B's "multiple
perspectives" rule with a live instance, and the regional chart is what
reconciles it (US weak + global plateau are compatible because the Americas are
9% of deployments).

**The premise** (D's test applied): *factory robot demand stepped up
structurally, not cyclically*. Falsifiable against one series: IFR annual
installations. If installations had reverted to 2020 levels, every article in
the cluster would need rewriting — so it is load-bearing.

### The substrate: 9 facts, every quote fetched and substring-verified in-session

`sql_for_agents/491_robot_hands_ifr_facts.sql`. One **series** fact
(`rh-ifr-installations-series`, 5 observations 2020–2024, each with its own
IFR press-release citation) + 8 metrics (world/China/Japan 2024, 2025 forecast,
2024 stock, three regional shares).

The verification discipline, per fact: WebSearch to find the release → WebFetch
asking for **verbatim sentences** → the quote stored is the fetched sentence.
Figures: 2020: 384,000 · 2021: 517,385 · 2022: 553,052 · 2023: 541,302 ·
2024: 542,000. Basis: **as first reported in each year's release** — the 2024
edition restates 2022 as 552,946 vs the 553,052 first reported, so mixing
editions mixes revision states (mig 265's Thames lesson); one basis stated in
the chart footnote instead. Four supporting metrics cite The Robot Report's
coverage of WR2025 (verbatim-quoted) where the IFR release page did not carry
the figure; publisher named honestly in each citation.

### The page: `/insights/robot-demand-step-change.html`

`sql_for_agents/492_robot_hands_demand_feature_page.sql`. Six sections, all
existing components, **Thames pattern verbatim**: content_data copies the
register exactly, `rendered_html` produced by a local harness replicating
`executeGoTemplate` (text/template, missingkey=zero, the call_agent.go:1168
funcmap, `<no value>` stripped — harness at scratchpad `build/render.go`),
rows locked `permanent`, `rebuild_policy='owned'`, slot names matching
`pages.sections` entry-for-entry (the 095 lesson).

hero → feature-analysis (`generic-text-block`; the editorial, every figure
within its fact's context-term window) → `evidence-timeseries-ifr` (the 5-year
series, scale = the 2025 forecast fact) → `evidence-chart-2024` (two figures:
regional shares scaled to Asia's own share; China/Japan/world scaled to the
world total) → feature-coverage (`generic-text-block`; the cluster's articles
linked out, title + link only, including the two-channels-one-story pair and
the divergent A3 read) → call-to-action (house copy verbatim — its figures are
already registered facts).

### Applied + verified

- 491: applied clean; verify block confirmed 5 observations each carrying its
  own citation. 492: applied clean; verify block confirmed 6 locked components,
  none under 500 bytes of HTML, sections↔slots exact.
- Rendered values checked before seeding: bars `--v:384000..542000;--m:575000`,
  ticks 2020–2024, five ifr.org citations in the sources block; evchart rows
  Asia 74/74, Europe 16/74, Americas 9/74; World 542000/542000, China 295000,
  Japan 44500.
- **claimscan: 0 findings across 6 components** (10 fleet-wide banned patterns
  included). Read correctly per the oufe runbook: the deterministic scan does
  not see finance vocabulary — here the risk surface is numbers, which it does
  see, and every number resolves.
- Deploy: assemble-only `page-rerender` dispatched (corr
  `1227617e-8ba6-43db-9e0a-badc6033cf49`) — **by hand-rolled envelope, not
  `TRIGGER_rerender_page.sh`**, because the script defaults an empty reason to
  `section_data_resolved`, and on a page whose chart is `render_mode='agent'`
  the safe path for fully-authored locked rows is the assemble branch, which
  never touches a renderer at all.

### MISSTEP 2 — the first components.tsv export dropped two claimscan columns

First export used `slot || tab || base64` — claimscan wants
`page, slot, base64[, page_type]` (`cmd/claimscan/main.go:117`) and would have
skipped every line as malformed. Caught by reading the parser before running,
not after a suspicious clean pass — worth recording because a malformed TSV
produces **"0 findings"** with only a stderr warning per line, which reads
exactly like a clean scan if you don't check the component count in the
summary line ("across 6 component(s)" is the tell).

---

## 2026-08-20 — owner rulings, and the start of rollout + design uplift

### Rulings recorded (owner, 2026-08-20)

1. **Lifecycle policy RATIFIED in full** — indefinite retention at one stable URL,
   deliberate de-listing rather than deletion, **cadence per fact**, and the
   tracker/feature/explainer tiers. `DESIGN_2026-08-19_starting_point.md` §3 is
   now the RULE, marked as ratified in place.
2. **Hero default = image + semi-transparent overlay**, ahead of gradient-only.
   Verified before acting: the live `hero` template **already** emits exactly
   that when given an image —
   `background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('<hero_url>')`
   — and falls back to the `--color-primary` gradient only when `hero_url` and
   `background_image` are both absent. So this is a **content/data default plus a
   generation step, not a new component**. The live editorial page was rendering
   the fallback branch, confirmed by curl:
   `class="hero" ... style="--hero-ink: var(--color-primary-text); background: linear-gradient(135deg, var(--color-primary) ...)"`
   — no `background-image`, which is the gap.
3. **Rollout** across other sites, **top nav** under Insights/Blog, and
   **news → editorial "further analysis" links**.
4. **Design uplift is its own workstream**, and the composition plan is to be
   written by **Fable** specifically.

### The two discussions the owner remembered — found, and they are ONE document

`docs024_key_docs_latest/inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md`
(design-only, **nothing built**):

- **§1–8 = interleaving copy with images/graphs.** Owner ask 2026-08-13. The
  problem it exists to solve: `article-body` holds the whole article in ONE
  llm-owned `content` field, so an in-body `<figure>` **works today and is
  destroyed by the next prose rewrite** — there is a LANDMINE for it, with a
  measured ~90ms loss window on `dartsonline.com/blog/flight-shapes.html`.
  Recommended remedy is **plan-as-truth**: one locked `site_plan_imagery` row per
  figure (`scope='section'`, `scope_ref='<page>:<ordinal>'`), consumed by the
  **already-live IMG-056 kind-alias resolver** (`ensureAssets`,
  `plan_sections_action.go:475-518`), plus a Phase-3 `style_hints.placement`
  splice injecting the figure into `rendered_html` only — writer marker optional
  and never load-bearing. Blast radius named: `article-body` = 93 instances /
  18 sites. `style_hints.placement` rows fleet-wide: **0** [MEASURED 2026-08-14].
- **§9 = components comprising other components** — the owner's steer of
  2026-08-15, quoted in full there.

### Composition: designed three times, exercised zero times [MEASURED 2026-08-20]

| mechanism | state |
|---|---|
| `page_components.parent_instance_id` | column + FK + index exist; **0 of 1580 rows**; **zero Go references** |
| `content_components.render_mode='composite'` + `child_components` | columns exist; `deriveRenderMode` (`store_generated_component_action.go:1481-1506`) can only emit `agent` or `template`, so `composite` is **unreachable by construction** (CTS-039, status partial) |
| `component_level` (site/page/section/element/…) | used as a flat classification filter only, never as a containment tree |

Assembly today is flat concatenation (`assemble_from_library.go:256-296`); the
single template executor (`executeGoTemplate`, `call_agent.go:1170-1220`) has no
`{{template}}`/partial support and a six-function funcmap. The deepest hierarchy
in the estate is `{{range}}` over nested JSON — `evidence-chart`'s charts→points,
which this lane is already using.

**So adopting composition is build-and-prove, not wiring** — and that inverts the
phasing and probably the architecture-scope call in the inline plan. Which is
exactly what §9 says the revision must re-take.

### MISSTEP 3 — Fable is blocked, and I did NOT substitute

`features_open/035_FEATURE_component_hierarchy.md` is a slot reserved on
2026-08-15 and still unwritten because **three prior Fable agents died on model
limits**. The handoff's instruction is explicit: *"The owner specifically wanted
Fable for the design work — do not silently substitute a model; ask."* The owner
confirmed it again on 2026-08-20 ("We want to be using Fable for the plan").

Dispatched a Fable agent with the full brief; it failed on
**"You've reached your Fable 5 limit"** — an account limit, not a transient
error, so an immediate retry would fail identically and was not attempted.
**035 remains unwritten, deliberately.** It is the fourth failure of the same
kind, which makes the pattern itself the finding: *this plan is blocked on
capacity, not on knowledge* — the brief is ready and everything it must read is
catalogued above.

### 2026-08-20 outcomes — all verified at the artefact, not the status

| thing | evidence |
|---|---|
| **Hero image live with overlay** | asset `content_hero_robot_demand_step_change` generated by image-build-handler and `active`; served page carries `background-image: linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6)), url('/assets/images/content-hero-robot-demand-step-change.jpg')`; the JPEG itself returns **200 image/jpeg**. Page grew 42,890 → 43,140 bytes. |
| **Insights in the TOP NAV** | live header now reads Tools · About · News · **Insights** · MatchMatrix · Selection Guide · Learning Center. `nav_drift` items both `complete`. |
| **Insights hub live** | `/insights/index.html` 24,555 bytes, correct title, links the feature. |
| **News → editorial link substrate** | `published_page_id` stamped on **8** feed rows, 1 page — first non-zero fleet-wide. |

**Hero work used the framework, not a hand-upload:** the request was filed under
the platform's own convention key
(`needs_imagery:page:<page>:content_hero_<page_underscored>` — `imageryplan.ItemKey`
/ `ContentHeroKey`), in the same item shape `check_content_image_missing` emits,
routed to `image-build-handler`. It went `detected → claimed → complete` in about
two minutes.

**Why the item was filed in that exact shape rather than left to a sweep:**
`improvement-sweep` is **DISABLED** (`enabled=f`, last triggered 2026-08-17), so
an item waiting on it would have sat for ever — the "detection works, dispatch
does not" class. `detected-item-promoter` (900s) and `build-pipeline-trigger`
(60s) ARE live and ticking, and they are what actually moved it. **Check which
sweep is live before filing anything at `detected`.**

### MISSTEP 4 — I read a 200 as "the page is live" when the body said "Not found"

Deploying the Insights hub, `curl -o /dev/null -w "%{http_code}"` reported **200**
while the body was **9 bytes reading `Not found`** — CDN propagation mid-flight.
A following `curl -L` returned the real 24,555-byte page.

**The check that catches it: assert on BYTES or on a known string, never on the
status code alone.** A status code is exactly the kind of proxy this repo's own
rule warns about ("trust the rendered artefact, not the status"), and I used it
anyway because it is the convenient flag. The honest one-liner is
`curl -sL <url> | wc -c` plus a `grep` for a string only the right page contains.

### Still open after this session

- **`analysis_url` code half** — `published_page_id` now has data and still no
  reader. The Go change to `render_news_section_action.go` (+ the two news
  components' JS) is designed, not built; it is a platform change and goes
  through the council gate.
- **Rollout to a second site** — the plan pauses here deliberately for the owner
  to look before batching.
- **035 / Fable** — blocked on capacity.

## 2026-08-20 — rollout site 2: dartsonline.com, LIVE

`/insights/darts-calendar-density.html` — 39,867 bytes, hero image present,
series ticks 2022-2026, claimscan 0/6, render audit **2 approximate failures and
no firm ones**.

**The premise, and why it is a premise rather than a topic:** the cluster (Hearn's
warning, Littler's Euro Tour withdrawals, Van der Voort's criticism, The Sun on
circuit life) reads as four discipline stories. The load-bearing claim is that it
is a **schedule-density** story — falsifiable against exactly one series, the
number of Players Championship events per season. Flat count → the argument dies
and the discipline framing stands. Not flat: **30/30/30** through 2022-24, then
**34** in 2025 and 2026, with the European Tour moving 14 → 15.

**Kept the dissenting item deliberately again.** On robot-hands it was the
weak-US-orders headline; here it is The Sun's "disheartening" piece — the same
story from the inside, and the reason to look at the calendar at all. Two for two
on the rule that the contradicting channel is the most valuable one in the cluster.

**Firsts on this page:** the hero shipped WITH its image (the owner's ruling
applied from the start rather than retrofitted), and 495's verify block asserts
it mechanically — the page cannot be seeded on the gradient fallback branch.

**Source honesty:** the calendar counts come from each season's own summary page
on Wikipedia, not the PDC's own calendar. Verified verbatim in-session, publisher
named in every citation, and the chart footnote says so on the page. That is
weaker sourcing than robot-hands' IFR press releases and is written as weaker.

`published_page_id` now stands at **15 rows across 2 pages** fleet-wide.

**Not done on dartsonline:** the `/insights/` hub + top-nav entry. The site's
header already carries 6 items, so it fits, but one feature does not need a hub —
the robot-hands hub was built because the owner asked for nav placement and that
site had the first feature. Second feature on dartsonline is the trigger.

### MISSTEP 5 — every editorial article was appending itself to every page's footer

Caught by the **`dartsonline_traffic` lane**, which messaged this session
unprompted with five measured warnings. The one that landed:

> "if you set a flag on each article, every article you publish appends itself to
> the footer of every page. With no flag it's omitted from nav entirely and is
> reachable only from its listing."

**Verified before acting** rather than taken on trust:

```
pages: robot-demand-step-change  in_header=f  in_footer=t
       darts-calendar-density    in_header=f  in_footer=t
curl https://robot-hands.com/about.html | grep robot-demand-step-change  → HIT
```

So the article was in the footer of a page it has nothing to do with, and the
clutter would have grown **linearly with every feature shipped** — invisible on
the feature page itself, which is where I was looking.

**Why I got it wrong:** I reasoned correctly that `/insights/` articles cannot
reach the *primary* nav and set `in_footer=true` so they would not be "omitted
from nav entirely". That is the right reading of the classifier and the wrong
conclusion about what to want — the hub page IS the listing, so an article needs
no nav membership at all. **Fixed: both articles now `in_header=f, in_footer=f`**,
with a `nav_drift` filed to rebuild.

**What I got right and will keep:** the section-index hub (their point 1 is the
same mechanism I found independently), letting `nav_drift`/`populate_nav_tables`
derive rows rather than inserting `site_nav_items` by hand (their point 2), and
**assemble-only rerenders with no `spec.reason`** (their point 4 — sending
`reason=section_data_resolved` can escalate a page to the content writer, and on
`article-body` that destroys in-body figures; they lost 4 guide figures that way).

**Two things of theirs I have not yet acted on:**
1. **Chrome propagation.** They measured a nav change reaching **2 of 23 served
   pages**, with `pages.rendered_footer` reading "absent" even for a page that
   was serving it — so that column cannot grade it, only served bytes can. Here
   the Insights link DID reach `/about.html` [MEASURED], so propagation is better
   on robot-hands than their case — but "better on one sample" is not "fine", and
   `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh` is the proven fix
   if it is patchy.
2. **`sitemap.xml` on dartsonline is a static generated file** and does not
   update itself, so `/insights/darts-calendar-density.html` is absent from it
   until `scripts/site-discovery-files.py dartsonline.com --write` is re-run and
   pushed. Theirs to run (they offered); flagged in the reply.

**Also from them, and it matters for 494:** dartsonline's `evidence_base` aspect
existed with an empty `facts[]` *deliberately*, and they would rather the shape
were agreed than each lane invented one. I had already populated it with 5 facts
(migration 494) — so the reply states the shape I used and invites correction
rather than presenting it as settled. Their related point checks out in our
favour: `validate_page_content.go` gates the unregistered-number scan on
`evidence_base` existing, and **`blog-post` is on the exempt list**
(`claims.go:752`), which is the page_type both features use.

**Footer fix: applied at source, propagation PENDING.** `pages.in_footer=false` on
both articles [MEASURED]. The `nav_drift` that republishes chrome sat at
`triaged` through ~8 minutes of polling (`attempt_count=0`, priority 100,
handler `nav-updater`) while an identical item three hours earlier completed in
about two minutes — so this is **dispatch latency, not a defect**: the site
carries a large non-claimable backlog (45 `undeployed_asset` unresolved, 33
`contrast_failure` deferred, 29 `head_essentials_missing` detected) and its turn
in the build rotation had not come round.

**So the served footer still lists the article at time of writing.** Two honest
readings to avoid: the DB is not the page (the fix is real but not yet visible),
and a `triaged` item is not a failed one (nothing needs re-filing). If it has not
drained by the next session, the manual path is
`docs/leopardessconsulting/scripts/reconcile_footer_nav.sh <site_id> <domain> <marker> [rounds]`
— the peer lane's proven fix, which took them from 2/23 to 23/23 served pages.
**Grade it on served bytes, not on `pages.rendered_footer`**, which they measured
reading "absent" for a page that was in fact serving the change.

### Footer propagation: mechanism PROVEN, 30 pages outstanding, and the reconcile tool cannot grade a REMOVAL

State after the `nav_drift` completed (17:50Z), measured rather than assumed:

| layer | state |
|---|---|
| `pages.in_footer` | **false** on both articles ✓ |
| `site_components.footer` (stored chrome) | article link **absent**, hub link **present** ✓ |
| served pages | **stale on ~30 of 31** ✗ |

This is the peer lane's point 3 reproduced exactly: *"a nav change updates stored
chrome everywhere and republishes almost nothing"*. Spot-checked `/`,
`/matchmatrix.html`, `/selection-guide.html`, `/how-it-works.html`,
`/news/index.html` — all five still serve the article link.

**Mechanism proven on one page.** An assemble-only `page-rerender` (no
`spec.reason`) on `about.html` took it from `article=1, hub=3` to
**`article=0, hub=3`** — the stale link gone, the new nav intact, 81,702 bytes.
So the fix path is right; it simply has to be applied per page.

**A real limitation of `reconcile_footer_nav.sh`, worth recording for whoever
extends it:** its marker is **presence-based** — it re-renders any page that does
NOT contain the marker string, and loops until all pages serve it. That grades an
*addition* perfectly and **cannot grade a REMOVAL at all.** My change adds nothing
new (every page already carries `/insights/index.html` from the earlier rebuild);
the only difference is the article link's absence. Passing any present string
makes the script conclude every page is current and do nothing; passing an absent
string makes it re-render everything every round without ever converging.

**What it would need:** an inverted mode — `--absent <marker>`, converging when
the string is gone. That is a small change to a proven script and would cover
every de-listing, which is now a first-class operation in this lane's lifecycle
policy (*"retirement is deliberate de-listing"*). **De-listing a retired feature
will hit this exact wall**, so it is worth building before the first retirement,
not after.

**Not doing 30 hand-rolled dispatches.** The script's own header warns the
`kubectl run -i --rm … kcat -P` stdin form (which every dispatch in this lane has
used, one at a time) **loses ~4 of 5 messages at exit 0** in a loop, which is why
it uses the container-COMMAND form with a `PUBLISH_OK` receipt. Reimplementing
that badly is how a "reconciled" site quietly stays stale. Offered to the peer
lane, which has run the tested loop to 23/23.

**Honest severity: cosmetic.** The stale link points at a real, live, good page —
nothing is broken, and the structural defect (footer clutter growing with every
feature shipped) is fixed at source. This is residue, not a live fault.

### ⚠ A consequence of `rebuild_policy='owned'` I did not note when I chose it

Surfaced by the `dartsonline_traffic` lane after building `--absent` mode:
**`reconcile_footer_nav.sh` cannot touch an owned page** — `save_sections` refuses
them outright, so firing at one produces a FAILED orchestration. On robot-hands
the script skipped **6 owned pages**, and 2 of the 6 are mine:

```
tool-grip-force-friction-calculator | tool          | owned
tool-matchmatrix                    | tool          | owned
tool-gripper-cycle-time-estimator   | tool          | owned
tool-gripper-payload-calculator     | tool          | owned
robot-demand-step-change            | blog-post     | owned   <- editorial feature
insights-index                      | section-index | owned   <- the Insights hub
```

**The trade-off, stated properly this time.** I chose `owned` + locked components
so no generic rebuild path could overwrite authored copy — and that is still
right, it is what protects the editorial text and the hand-rendered charts. What I
did not note is the other half: **an owned page is also excluded from chrome
propagation.** It keeps its authored body *and* its chrome frozen at whatever was
current when it was last assembled.

**Measured now, and it is fine — by luck, not by mechanism** [MEASURED 2026-08-20]:

| page | stale article link | current hub link |
|---|---|---|
| `/insights/robot-demand-step-change.html` | 0 | 3 |
| `/insights/index.html` | 0 | 6 |
| `dartsonline.com/insights/darts-calendar-density.html` | 0 | n/a |

All three were built or re-assembled *after* the nav change, so they picked the
current chrome up on the way past. The peer lane made the same observation about
the dartsonline page and called it correctly: **that is luck about ordering, not a
property to rely on.** The next chrome change that does not coincide with a
deploy of these pages will leave them behind silently.

**The remedy exists and is named for exactly this case** —
`docs/leopardessconsulting/scripts/refresh_owned_page_chrome.sh` (verified present,
5,031 bytes), cited in `LANDMINES.md:765` alongside `reconcile_footer_nav.sh` as
the owned-page counterpart. Neither the peer lane nor this one has run it, so it
is **unvouched** — do not describe it as proven.

**Why this must be handled before the first de-listing, not after.** The ratified
lifecycle makes retirement *"deliberate de-listing"*. A retired editorial feature
is: owned (so the ordinary reconcile skips it), and the *subject* of the chrome
change (its own link must disappear from every other page's nav). So the first
retirement needs BOTH the peer's new `--absent` mode for the other 30 pages AND
the owned-page path for the retired page itself. Neither half is exercised today.
**Added to the lane's open list rather than left as a discovery for whoever
retires the first feature.**

### CORRECTION — my "~30 of 31" was a spot-check extrapolation; measured it is 19

I wrote that ~30 of 31 pages carried the stale link, from **five** spot-checked
URLs that all returned 1. The peer's dry run measured the real distribution: **26
reconcilable, 6 owned skipped, 1 unfetchable, 19 actually carrying it.**

Five for five is a perfectly good reason to believe "most", and no reason at all
to write "~30 of 31" — that number came from `31 deployed pages` minus a guess,
not from counting. The population had three classes I had not looked for (owned,
unfetchable, already-current) and my sample could not have revealed any of them.
**Say "5 of 5 sampled" or count the population; do not dress a sample as a
census.**

### MISSTEP 6 — my own premise test leaked into a public meta description, and I wrote it by hand

Found by the `meta_description_never_backfilled` lane (`bugs_open/320`, `339`).
The live `pages.meta_description` on `darts-calendar-density`, 291 characters:

> "…Set against the calendar itself — 30 Players Championship events a season
> through 2024, 34 since 2025 — **these are one story about schedule density, not
> four about discipline**."

That closing clause is **design D's premise test**, straight out of this lane's own
design doc — an instruction about how to frame the piece — printed as the sentence
a search engine shows under the title. `robot-demand-step-change` was milder but
the same shape: 242 chars opening "An editorial feature reading this week's…",
which describes the artefact to a colleague rather than the subject to a reader.

**Both were hand-authored by me, in migrations 492 and 495.** Not a pipeline
defect.

**The part worth keeping, and it is about their investigation rather than my
error.** That lane eliminated the site planner (absent from `site_plan_pages` in
every plan), the tool path, `apply_gap_plan_action` (none of its five
`ON CONFLICT` clauses writes the column) and the rerender (the page already held
the text before the earliest orchestration carrying it). **That chain was
exhaustive over code paths, and the answer was not on one** — a session typed the
string into a seed.

**So the finding for `339` is a category, not a producer:** a seed migration
writing `pages.meta_description` directly **bypasses every producer-side
control**. Their proposed remedy for the tool half — *"don't pass the brief as a
candidate at all, compose the public sentence separately"* — is right there and
**cannot apply here, because there is no composer in the path to fix.** For a
hand-authored row the guard is the only control, and `PublicMetaDescription` is
measured not to fire in the 200–320 band (0 of 693 live descriptions exceed 320;
the marker regex matches none of the 11–12 in that band). Fixing this class needs
either a guard that catches it or a check at seed time.

**Why I made it.** I composed the description in the same sitting as the premise,
from the same sentence, and never re-read it as a reader. The tell was available
and I did not look for it: **a description that argues with an alternative reading
is written for someone deciding how to write the piece.** A reader has no
alternative reading to be corrected out of.

**Fixed** (`sql_for_agents/497`, with a backup table and a verify block that
refuses a description over 160 chars or carrying a brief-shaped construction).
Both rewritten reader-facing and under 160 — 153 and 157, against 291 and 242,
since search snippets truncate around there anyway. Both redeployed.

**A third page in the same shape is NOT mine** —
`leopardessconsulting.co.uk/hierarchical-multi-agent-orchestration-explained`
belongs to another lane; I have not touched it and have said so rather than let
it look handled.

**The standing check for this lane:** every seeded page's `meta_description` is
public copy and gets read back as a stranger would read it, before the seed is
applied. It is the one field in these migrations that is neither prose the
framework wrote nor a figure resolving to a fact.

### The two originals, preserved VERBATIM — they are a detector corpus, not just an error

The `320` lane made a point that changes what these strings are for. They tried to
measure my proposed structural tell (*"a description that argues with an
alternative reading was written for someone deciding how to write the piece"*)
against the 704 live descriptions, and got: the lexical proxy
`\y(not|rather than|as opposed to|instead of)\y` fires **34 times, zero in the
200–320 band**, all 34 on legitimate copy where "not" does ordinary work
("…built around your business and your data, **not** l…", "A loan from an
FCA-unauthorised lender is **not** legally enforceable…").

**And then they said the test was not fair, which is the part worth keeping:**

> "your two were fixed before I measured, so the population no longer contains a
> single true positive. I was fitting a signal against a corpus its own
> remediation had emptied."

That is the **repro-destroyed-by-the-render** shape, and it is already in this
estate's memory under that name — I did not spot it and they did. "Zero in the
band" is not evidence the tell fails in the band; it is evidence there is nothing
left in the band to fire on. **My remediation destroyed the only true positives
for the detector my own suggestion needs.**

So the two strings below are **the only known corpus for sub-class (2)**
(hand-authored brief-shaped descriptions). They exist in exactly two places: the
DB table `pages_backup_20260821_meta_desc`, and here. **The
version-controlled copy is the durable one.** Do not "tidy" either away.

> **CORRECTED 2026-08-21 — the specific risk I asserted here does not exist, and
> I had not measured it.** The paragraph above originally read *"a table named
> `*_backup_*` is precisely the shape a cleanup routine targets —
> `database-cleanup` runs hourly and its config is opaque"*. The `320` lane
> checked, and I then verified independently: `database-cleanup`'s **`pre_query`**
> issues `DELETE FROM` against five **named** tables (`agent_error_log`,
> `orchestration_state_audit`, `orchestration_states`, `palettes`,
> `typography_sets`) — no `DROP`, no wildcard, nothing matching `pages_backup_*`.
> No production Go path drops a table at all: the sole `DROP TABLE` in
> `platform/`, `internal/`, `cmd/` is a **test fixture**
> (`refresh_evidence_base_test.go:22`, the string
> `"SELECT count(*) FROM sites; DROP TABLE sites"`) asserting that
> multi-statement SQL is *refused*, which cuts the other way. And
> `pages_backup_20260717_r6` has sat there untouched since 17 July.
>
> **How I got it wrong:** I read `scheduled_tasks.input_data`, found `{}`, and
> called the config opaque. The work is in `pre_query`, a column I did not open —
> so "opaque" described my query, not the task, and I built a risk claim on top of
> it. Reading one column of a row is not reading the row.
>
> **The reasoning for the committed copy still stands, and is the better one
> anyway:** a table's safety is a fact about *today's* cleanup config, and the
> file's safety is a fact about git. Keep both; just don't cite a threat that
> isn't there.

**`darts-calendar-density`** (291 chars, was live 2026-08-20 → 2026-08-21):

> Barry Hearn warned top players about skipping tournaments and Euro Tour withdrawals left organisers with a headache. Set against the calendar itself — 30 Players Championship events a season through 2024, 34 since 2025 — these are one story about schedule density, not four about discipline.

**`robot-demand-step-change`** (242 chars, same window):

> An editorial feature reading this week's robot-demand coverage across several channels, charted against the IFR's own five-year installation series - a step change that has held at altitude, and what that plateau means for end-of-arm tooling.

**What a detector-builder should notice about the pair**, since two examples is
all there is: they fail in *different* ways, and only one of them contains "not".
The first argues with an alternative reading; the second never mentions the
subject at all — it describes **the artefact** ("An editorial feature reading…
charted against… ") to someone deciding whether the piece was well made. So the
common property is not the contrastive construction I proposed. It is closer to
**wrong audience**: both address a person making editorial decisions rather than
a person deciding whether to click. That is semantic, a regex cannot see it, and
one of my two examples would have escaped my own suggested rule.

Recorded as a correction to my own tell rather than a defence of it.

## 2026-08-21 — third feature LIVE, and the first one built after the fixes

`robot-hands.com/insights/electric-vs-pneumatic-economics.html` — 86,602 bytes,
hero image present, 10 chart rows, 6 ink-token references, claimscan 0/5.
The Insights hub now lists **both** robot-hands features and is itself clean
(0 contrast failures).

**This is the first feature built after 496 and after the meta-description guard,
so it is the test of whether those fixes hold for new work rather than just
repairing old.** They do:

| | robot-hands feature 1 (before fixes) | feature 3 (built after) |
|---|---|---|
| contrast failures | 10 | **4** |
| of which OURS | 6 | **0** |

All four remaining are the pre-existing shared-component family — the 1.00:1
white-on-white `.cta-btn cta-btn-primary` and three over-an-image approximations —
byte-identical to what the untouched control pages carry. **A page built today
inherits the repointed components and introduces no contrast defect of its own.**
Stylesheet control run first, 25,559 B.

### The guard I wrote yesterday caught me today

`498`'s verify block refused the first apply: *meta_description is 162 chars, over
the 160 limit*, and rolled the whole transaction back. That guard exists because
of misstep 6 — a hand-typed commissioning note serving as a public description —
and it fired on the very next feature I wrote.

**What that does and does not prove.** It caught the **length** half, which is
what a machine can check. The **tone** half — the actual defect in misstep 6 — is
still caught only by a human reading the sentence as a stranger. So the guard is
worth having and is not the control; I re-read both new descriptions as a reader
before applying, and that step is still the one doing the real work.

### Sourcing was a step better this time, deliberately

Every figure is a **verbatim quote from the primary** — the ENERGY STAR / US DOE
sheet — extracted with `pdftotext` in-session. The search results were dominated
by compressor-vendor pages restating the same numbers, and the dartsonline feature
had already had to disclose weaker (Wikipedia) sourcing on the page. Where a range
was given (7–8 hp), the registered fact takes the **low end**, because the
conservative end favours the side the piece argues against.

### A fourth feature NOT built, and why

The cobot cluster (six items: "Collaborative Robot Usage Shifts to Factory
Strategy", Techman dual-arm, Movotrak seventh axis, Hirebotics explosion-proof,
portable-cobot market forecast) was the next strongest and **I stopped**. IFR's own
release yields exactly one verifiable figure — *"Cobots reached a market share of
10.5% of industrial robots installed worldwide in 2023"* — and the 2024 numbers
(11.9%, 64,542 units) exist only in a search summary and on a page returning 403.

One cited figure is not a background section, and padding it with second-hand
restatements is the exact failure this lane exists to prevent. **Recorded as a
ready candidate rather than a gap:** the cluster is real, the premise is good (the
coverage reads as a takeover; ~1 in 10 installations says otherwise), and it needs
one thing — a primary source for the 2024 share.

## 2026-08-21 — cobot feature PARKED with its unblocker named, and the de-listing path prepared

### The fourth feature is parked, not abandoned

Went back for the 2024 cobot figures rather than leaving the candidate vague.
**They are not verifiable from a primary.** IFR's own cobot page
(`how-robots-work-alongside-humans`, position paper updated 2024-12-04) contains
exactly two usable numbers, both 2023:

> "cobots reached a market share of 10.5% of industrial robots installed
> worldwide in 2023" · "Cobots accounted for 10.5% of the total 541,302
> industrial robots installed in 2023"

The 11.9% / 64,542 figures for 2024 appear only in an unattributed search summary;
a targeted search for them returned *"the specific percentages and unit numbers
you mentioned don't appear in these search results"*, and the one page carrying
them 403s. **Two figures from one year is not a background section.**

**Parked with its unblocker stated:** the cluster is real (six items) and the
premise is good — the coverage reads as a takeover, ~1 in 10 installations says
otherwise. It needs **one** thing: a primary source for the 2024 share. The IFR
World Robotics report itself is the obvious candidate and is paid; the annual
cobot press release is the free one and has not yet been published for 2025 data.
**Recording the negative result so nobody re-walks the search.**

### `refresh_owned_page_chrome.sh` — safety property now VOUCHED, fix property not

Ran it deliberately on a page where the expected outcome was *no change*, which is
what makes it a safe first exercise: `electric-vs-pneumatic-economics`, owned,
5 permanently-locked components, chrome already current.

Result: flipped to `generic`, published, render COMPLETED, and — in its own
words — *"restoring ownership before verifying (protection first, cosmetics
second)"*. Afterwards: `rebuild_policy='owned'`, **5 locked rows intact**, served
page **byte-identical at 86,602** with hero and 10 chart rows.

**Be precise about which half that vouches for.** It proves the script *does no
harm* and restores protection. It does **not** prove it propagates stale chrome,
because the test page had none — the run was a no-op by construction, and a no-op
cannot demonstrate a fix. That is the same shape as the blind-stylesheet pass:
a measurement that could not have come out otherwise. Stated here so the next
session does not read "tested" as "proven".

Full procedure written up as `RUNBOOK` §10, including the marker trap (a census
over-reports on the hub and on the retired page itself, which are the two pages
every de-listing touches by definition) and the `rebuild_policy='generic'` window
with the two things that bound it.
