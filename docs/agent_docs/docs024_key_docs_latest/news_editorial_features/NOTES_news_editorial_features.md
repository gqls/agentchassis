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
