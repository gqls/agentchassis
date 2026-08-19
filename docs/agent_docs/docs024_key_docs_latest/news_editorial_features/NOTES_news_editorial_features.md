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
