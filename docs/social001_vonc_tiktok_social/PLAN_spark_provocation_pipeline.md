# Plan — Spark daily provocation content pipeline (vonc.com)

**Created:** 2026-06-25
**Goal:** Fill the empty-shell components (provocation-card, lobby-grid,
brief-explanation) with daily-regenerated content via the runtime JS pipeline
(Option 2), modelled on the existing news feed pipeline.

---

## Why Option 2 (not Option 1)

The Spark v1 roadmap (from the original design conversation) specifies these
components display **daily-regenerated content from a scraping pipeline**, shown
via client-side JS reading a static JSON file. This is the intended product
behaviour — provocations change daily, are shareable, and act as SEO content.
Option 1 (bake static LLM content at build time) would freeze a single set of
provocations permanently, defeating the daily-content product hypothesis.

Critically, **the delivery mechanism already exists** — this is not net-new
infrastructure. The news feed pipeline is the proven template.

---

## Architecture — mirror the news feed pipeline

### Existing news pipeline (the template)
```
content-feed-trigger (scheduled, 6h)
  → content-feed-orchestrator
      1. seed_sources       — create content_sources if none
      2. load_due_sources
      3. dispatch_sources   — spawn feed-ingester per source
      4. spawn_triage       — spawn feed-triage
  → feed-ingester (scrape/search → content_feed_items, status='ingested')
  → feed-triage (score relevance + credibility)
  → render step → /data/latest-news.json
```

### Proposed Spark provocation pipeline (parallel structure)
```
provocation-refresh (scheduled, daily)
  → provocation-orchestrator
      1. seed_provocation_sources  — scraping targets for trending topics
      2. load_due_sources
      3. dispatch_sources          — spawn ingester per source
      4. spawn_generation          — spawn provocation-generator
  → feed-ingester (REUSE — scrape trending topics → content_feed_items)
  → provocation-generator (LLM: turn raw topics into provocations + AI takes)
  → render step → /data/provocations.json
```

---

## What to reuse vs build

### Reuse as-is
- `content_sources` table — same schema, different rows (Spark's scraping targets)
- `content_feed_items` table — raw scraped topics land here, same as news
- `feed-ingester` agent + `fetch_scrape`/`fetch_news_search` actions — the
  scraping layer is identical
- `dispatch_feed_sources` action — spawns ingesters per due source
- `render_js_snippets_for_site` + `site-asset-renderer` — the JS delivery
- `scheduled_tasks` mechanism — same as content-feed-refresh

### Build new
1. **`provocation-generator` agent** — the one genuinely new piece. Takes raw
   scraped topics from `content_feed_items` and runs an LLM to produce
   provocations (contested claims) + AI takes + metadata. Writes structured
   provocation data. Modelled on `feed-triage` but generative not just scoring.
2. **A render action** producing `/data/provocations.json` — modelled on whatever
   produces `latest-news.json`. Shapes the provocation data for the JS to consume.
3. **`js_snippets` rows** for `provocation-card`, `lobby-grid`, `brief-explanation`
   — fetch `/data/provocations.json`, populate the shells. The shells already
   have the right DOM structure (`.pc-headline`, `.pc-card`, `.pc-stat-value`
   etc.) — the JS just fills them.
4. **`provocation-orchestrator` agent** — thin orchestrator mirroring
   content-feed-orchestrator. Mostly a clone with the generation step swapped in.
5. **A scheduled task** `provocation-refresh` — daily trigger.

---

## Build sequence (incremental, testable at each step)

### Phase 1 — Confirm nothing exists yet
Run the 5 diagnostic queries (in runbook Step 8) to verify no content_sources,
content_feed_items, scheduled_tasks, agents, or js_snippets for Spark already
exist. If any partial work exists, extend it rather than duplicate.

### Phase 2 — JS snippets against static test data first
Before building the data pipeline, prove the JS layer works:
1. Hand-write a `/data/provocations.json` with 2-3 sample provocations and commit it.
2. Write the `js_snippets` rows for provocation-card (and lobby-grid,
   brief-explanation) that fetch and render it.
3. Add `applies_to` matching these component functions.
4. Trigger `site-asset-renderer` for vonc.com to render snippets.js.
5. Verify the shells fill correctly in the browser.

This de-risks the hardest-to-test part (DOM population) using static data,
before any pipeline exists. If the JS fills the shells from static JSON, the
data layer just needs to produce that JSON shape.

### Phase 3 — Data pipeline
1. Define the provocation JSON shape (from what the Phase 2 JS expects).
2. Seed `content_sources` for vonc.com (trending-topic scraping targets).
3. Build `provocation-generator` agent (LLM topic→provocation).
4. Build the render action → `/data/provocations.json`.
5. Build `provocation-orchestrator` (clone content-feed-orchestrator, swap
   generation step).
6. Test end-to-end manually (trigger orchestrator once, verify JSON produced).

### Phase 4 — Schedule + automate
1. Add `provocation-refresh` scheduled task (daily).
2. Verify the daily cycle runs and regenerates the JSON.
3. Confirm the static site picks up new provocations after each run.

---

## Open questions to resolve before Phase 3

1. **What sources?** News RSS, trending topics (Reddit/HN/X), or a curated
   topic list? The roadmap says "scraping pipeline" — need to define targets.
2. **JSON shape** — driven by Phase 2 JS, but needs to cover: provocation text,
   AI take, shareable URL slug, stats (positions filed, etc.), and the lobby
   "today's other provocations" miniature.
3. **How many provocations per day?** The index shows one main + a lobby of
   others. The provocations page shows all of today's.
4. **Does the provocations-index page** (currently zero components) also read
   from this JSON, or does it need its own build?

---

## Standing constraints (from project rules)

- Every agent is an orchestrator; workflows thin, complexity in Go actions.
- Don't create sub-workflows in SQL — spawn sub-agents with their own workflows.
- Reuse before building: feed-ingester, content_sources, content_feed_items,
  render_js_snippets_for_site all already exist.
- Agents respond to parent's response topic.
- `logger.Info` not `logger.Debug`.
- Check schemas before SQL.
- Kubernetes: `-n ai-persona-system` for pods, `-n kafka` for Kafka.
- Work item specs needing page routing must include `page_name`.
- Known hazard: a `needs_page` rebuild on a tool/interactive page regenerates
  from `plan_sections` and silently drops interactive content. The provocation
  shells are populated by JS not page_components, so a rebuild of index won't
  lose them — but keep this hazard in mind for the tool pages.
