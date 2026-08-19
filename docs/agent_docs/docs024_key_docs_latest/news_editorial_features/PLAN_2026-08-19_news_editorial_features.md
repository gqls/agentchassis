# PLAN — news editorial features (2026-08-19)

**What the owner asked for:** editorial feature articles built around the bigger
current news stories, with the research filled out by graphs and charts of
background information, assembled by extracting the concepts out of news articles
arriving from several different channels.

**Stage 1 (this document): collect every past discussion of this into one place.**
No build this session. Decisions and their reasons are recorded here as they land.

**Diagnosis run:** `090` filed on the mechanism before any design work, per the
cross-cutting-claim rule. Run correlation
`3802fb10-c34f-4eff-9914-b2959c723bd5`, **verdict CONFIRMED** (three iterations,
`stopped_by: confirmed`). It also found a file this pass had missed — §3a below,
and the full read-out in `NOTES_news_editorial_features.md`.

---

## 1. The headline finding

**This has been designed three separate times, by three different threads, and
never built.** The three designs do not contradict each other — they are
successive refinements, and each contributes something the others lack:

| # | when | where | the contribution that survives |
|---|---|---|---|
| A | ~2026-04 | `docs019/_archive/archive_april_26/027g_news_expansion_architecture.md` §"Tier 3" | the **pipeline shape** — cluster → research → write → visualise → publish, as named agents, with an `event_timeline` table for continuity |
| B | ~2026-04 | same archive, `036b_news_content_diversity_plan_v2.md` §"Original Research & Writing Pipeline" | the **quality machinery** — readership segments, a research→writer→eval loop, continuous annotated timelines, scenario analysis, and the explicit no-prediction boundary |
| C | 2026-07-19 | `features_open/001_FEATURE_packaged_topic_features.md` | the **economics and the safety property** — shared research substrate, per-site angle, and why the naive version is actively harmful |

The current live doc, `docs024_key_docs_latest/006_news_feed_pipeline_v2.md`,
carries A forward as **Tier 3** of its three-tier roadmap: *"Research analysis —
multi-source analysis with timelines, graphs — `/insights/` — Planned."* Tiers 1
and 2 are marked working. So the thing being asked for is the platform's own
stated next tier, and it has sat at "Planned" since April.

Concept register entries covering the same ground:
- `news-feed-pipeline.md` **NEWS-005** (content diversity & original research
  pipeline) — status **aspirational**
- **NEWS-006** (the publishing gap: "the pipeline ends at curation") —
  **aspirational**
- **NEWS-009(b)** (news to infographic) — **aspirational**
- **NEWS-010** (`/insights/` as the Tier-2 target) — **superseded**; the
  archive-first listing page took the Tier-2 slot, and the rewritten-article idea
  was pushed down into Tier 3, where it still sits
- `topic-intelligence.md` **TPI-001/002/003** — all **abandoned**; a much larger
  ambition (audio monitoring, auto-spawned topic agents, cross-domain
  intelligence) whose engineering notes on dedup, temporal tracking and the
  LLM-vs-code division of labour are still worth reading
- `research-agents.md` **RES-006/007** — watchlists and a deep-research
  classifier, aspirational and abandoned respectively

## 2. What already exists, and it is more than the docs suggest

Verified live 2026-08-19 (queries in `NOTES`, figures re-measured, not carried):

**Concept extraction is already running.** `content_feed_items.topics` is
populated on **9,622 of 10,855 items**, written by `feed-triage` at
`platform/orchestration/actions/feed_triage_actions.go:245`. The triage prompt
also produces credibility, a source-tier classification and a provenance chain
(`original_source` / `found_via` / `source_tier`). That is the "extract the
concepts" half of the ask, live and at volume.

**Chart rendering exists and is proven.** `evidence-chart`,
`evidence-timeseries` and `mechanism-flow` are active section components.
`evidence-timeseries` has served a real five-observation series on a live page.
These are **data**, not Go code — a source grep will not find them, which is
precisely the error `visualisation-and-charts.md` was written to stop repeating.

**A research substrate exists.** `research_results` holds 177 rows across 8
sites; `research-agent`, `evidence-researcher` and `content-researcher` are
active agent definitions.

**Seventeen news pools exist and are dormant** (`sites.status='pool'`), created
2026-07-20 and deliberately parked.

## 3. What is missing — the three zeros

The gap is not the concepts and not the charts. It is **everything between one
item and a story**:

```sql
count(*) FILTER (WHERE entity_ids        IS NOT NULL) -- 0
count(*) FILTER (WHERE duplicate_of      IS NOT NULL) -- 0
count(*) FILTER (WHERE published_page_id IS NOT NULL) -- 0
```

- `duplicate_of` and `entity_ids` are declared in `content_feed_items` and
  **written by no code path in the repo** (grep over `platform/`, `internal/`,
  `pkg/` returns nothing for either; control grep `relevance_score` returns 7
  hits over the same paths, so the search shape works). So the same story
  arriving from four channels is four unrelated rows.

  > **CORRECTED 2026-08-19, by the diagnosis run** (`3802fb10`, verdict
  > CONFIRMED): this bullet first read *"the 'different channels' part of the ask
  > has no mechanism at all"*. There is no **grouping** mechanism, but there is a
  > **cluster-detection** one, pointed the other way — see §3a. The run found the
  > file this pass never opened.
- `published_page_id` is never set because nothing publishes: `article-rewriter`,
  `feed-publisher`, `news-analyst`, `story-researcher`, `analysis-writer`,
  `visualization-renderer` and `data-chart-generator` are **all absent** from
  `agent_definitions` (checked with a positive control).
- `event_timeline` and `topic_packages` **do not exist as tables**.

`topics` is free text with no controlled vocabulary, so it is a starting signal
for grouping, not a grouping key.

## 3a. The near-miss: a cluster detector that exists to discard clusters

`platform/orchestration/actions/queryresolve/news_items.go` already notices when
several headlines are about the same thing — and **exists to throw that signal
away**. `newsTopicalTokens` (:185) builds a topical vocabulary from the site's own
source queries, and additionally derives one by document frequency over item
titles once the pool reaches 12 items (threshold `len(items)/4`, floor 5).
`capNewsItemsPerTool` (:222) then **drops** any item whose title shares a tool key
with `maxPer` already-kept items, so that no single story dominates the feed.

Its own comment is the sharpest statement of this workstream's problem anywhere
in the repo, written for the opposite purpose:

> Frequency derivation needs a pool large enough that "appears a lot" means "is
> the subject matter", not "is one well-covered story". Below 12 items a genuine
> tool cluster (four Firefox headlines) would cross any workable threshold.

**"Four Firefox headlines" is exactly the input an editorial feature wants.** The
display layer's requirement (suppress the cluster) and the editorial layer's
requirement (find the cluster) read the same signal in opposite directions.

Two consequences for the design:

1. **Do not bolt a grouping step onto this path.** Its contract is suppression,
   and another feature depends on that. The detection *heuristic* is worth
   copying; the call site is not.
2. **The thresholds have already been thought about.** `>= 12 items`, `len/4`
   floor 5 — tuned against real feeds. A grouping step starts from a tuned
   heuristic rather than a blank page.

And one naming trap: `registry.go:1386` describes `WriteFeedItemsAction` as
*"Normalise and write feed items to content_feed_items **with dedup**"*, which
reads like cross-source de-duplication and is not. The dedup skips items whose
**`source_url` already exists for this site** (`feed_actions.go:777-778,896-905`,
via the `idx_cfi_dedup` partial index). Exact URL, per site. Reuters and the BBC
on one story are two rows by design. **The one thing named "dedup" here solves a
different problem**, and the registry description will mislead a reader who does
not open the file.

Finally, what the render path can select on: `QueryNewsItems` filters on
`site_id` and `status IN ('relevant','ingested')` and projects title, summary,
url, published date, source name and **`topics`**. `duplicate_of` and
`entity_ids` are not selected at all — so adding a grouping key needs a change at
the read side too, not only a writer.

## 4. The one design decision already taken, and it is load-bearing

From `features_open/001`, owner-endorsed 2026-07-19:

> **A package written once per pool and published to all its domains is *worse*
> for duplicate content than a headline list, not better** — it is long-form
> near-identical prose, the shape search engines penalise hardest.

The resolution is a two-layer split, and it is the same shape the component
library already uses (`forked_from IS NULL` then per-site fork):

| layer | shared, once per topic | per-site |
|---|---|---|
| topic selection | yes | |
| research: history, timeline, related events, macro figures, quotes | yes — **the expensive part** | |
| fact/figure substrate with citations | yes | |
| **the angle** — what this means for *this* audience | | yes, one generation |
| headline, framing, examples, CTA | | yes |

Cost: ~1 package/week x ~231 money-pool domains = ~231 generations/week, against
the naive per-site feed design's projected ~8,000 triage calls/**day** at 2,000
sites. Affordable, and spent on the differentiating layer.

**Stated dependency, and its direction matters:** the angle needs an audience to
be angled at. That is the `site_specs` aspect **`audience`** settled as Decision 7
of the pooling workstream — pool-level default, forked per site. 001 is explicit
that this is *a prerequisite, not a parallel task*: "building packages before
profiles exist produces 231 variations on one article."

## 5. What the earlier designs add that 001 does not

**From A (027g):** the concrete pipeline, and the `event_timeline` table as the
continuity mechanism — each analysis piece both reads from and writes to it, so
the tenth article on a subject has context the first could not. Also the
publishing plumbing: `news-post` page type at `/insights/{slug}.html`, a
`rebuild_news_listing` action mirroring `rebuild_blog_listing`, and one line in
`link_constraints.go`'s URL builder.

**From B (036b):** readership segments as the axis of differentiation
(procurement / operations / trading / strategy for one vertical — the same event
framed four ways), the research-writer-**eval** loop with an explicit quality
bar, and the **scenario-analysis boundary**: structured if/then with cited
reasoning is allowed; confident directional prediction is not, and the eval agent
enforces it.

**Where A is now wrong.** A specifies visualisations as **static SVG embedded in
`page_components.rendered_html`**. Since then the platform has learnt three
things that make that the wrong default:
1. `<svg>` is in `nonAssertionElements` — **text inside an SVG is invisible to
   the claims gate** (`claims.go:137`; VIZ-009). A chart drawn in SVG text leaves
   the verification net entirely.
2. **There is no arithmetic in the render funcmap, and a missing function is a
   parse error** (VIZ-007) — so a template cannot compute SVG coordinates; the
   component renders *nothing*, it does not degrade.
3. The live components solve both by drawing bars and connectors in **CSS** from
   custom properties, with real HTML text for every label and figure.

So the visualisation layer should be `evidence-chart` / `evidence-timeseries` /
`mechanism-flow`, not a new SVG emitter. That also inherits their honesty
property: **a chart point cannot carry its own number** — every plotted value
resolves through a registered fact id, and a series observation carries its own
citation rather than inheriting the parent's (VIZ-001, VIZ-003).

## 6. Open questions, carried forward unanswered

From `features_open/001`, still open:
- **Who picks the topic** — editorially, or detected from a volume spike in the
  pool? 001 notes a cluster forming in embedding space *is* the signal.
- **Update trigger and retirement rule.** "Until it gets irrelevant" needs an
  operational test; decaying article volume in the cluster is the candidate.
- **SEO shape of an update** — update one URL in place (accrues authority) versus
  a new URL each week (fragments it). Expensive to reverse, so decide before the
  first one ships.
- **Is the substrate a first-class entity** (`topic_packages` with its own
  lifecycle, so a site can join a package late) or just a work item that produces
  pages?
- **Rights.** Collected opinion and quotes mean quoting third parties at length.
  The feed's existing posture is title + short summary + link-out, never
  full-text republication.

New, from this pass:
- **The substrate update fan-out.** 001 names it as the row most likely to be
  skipped and most likely to hurt: when the substrate changes, every derived
  angle is potentially stale. It distinguishes **substrate-only** updates (new
  figures, angles re-render from the same claims) from **narrative** updates (the
  story changed, angles must be regenerated) — only the second costs per-site
  money and only the second needs a blast-radius record.
- **`entity_ids` is an empty column with no writer.** Before designing a grouping
  key, establish what it was for — reusing a designed-but-unwired column is
  cheaper than inventing a parallel one, and this platform's idiom is reuse.
- **`evidence_base` coverage moved.** `features_open/023` reasoned from **8
  sites** on 2026-07-25; it is **17** today. Any argument in 023 that turns on
  scarcity needs re-measuring before it is repeated.

## 7. Blockers that are not ours but gate us

- **Pooling is parked by owner decision** (2026-07-20, "nothing further happens on
  this workstream until the owner says go"). The 17 pools are dormant; pilot
  onboarding and the first pool arming are packaged in
  `features_open/005_FEATURE_pilot_onboarding_and_first_pool.md`.
- The `audience` aspect (section 4) rides on that onboarding for most sites.
- **The fleet-wide version needs pooling; a single-site version does not.** A
  first editorial feature could be built on one site that already has a live feed
  and an `evidence_base`, with no pool and no per-site angle layer at all. That is
  the cheapest path to finding out whether the output is any good, and it is the
  build order both A and B independently argued for ("don't build Tier 3 until
  Tier 2 has proven the article quality is good enough").
