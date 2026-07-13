# 006 — News Feed Pipeline

Implementation guide for the content feed pipeline. Covers: architecture, all deployed components, operational procedures, resolved issues, and expansion roadmap.

Test site: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`)

---

## Current Status — Working End-to-End

The pipeline is fully operational for gaswholesalers.com and ready for any site with a classification spec.

| Layer | Status | Evidence |
|-------|--------|----------|
| Source seeding | ✅ Deployed | `SeedContentSourcesAction` reads classification spec, creates `content_sources` |
| Ingestion (RSS) | ✅ Working | 5 articles per run from OilPrice |
| Ingestion (news_search) | ✅ Working | 2-5 articles per run via web search adapter |
| Ingestion (api_news/Grok) | ✅ Working | 4-5 articles per run via xAI Responses API with web + X search |
| Ingestion (scrape) | ✅ Working | 20 items per run (nav links — filtered by triage) |
| Triage (relevance + credibility) | ✅ Working | LLM scores relevance, credibility, source attribution per item |
| Render (JSON) | ✅ Working | `/data/latest-news.json` with 6 items committed to git |
| Git commit → S3 deploy | ✅ Working | GitHub Actions pushes to Backblaze S3 on each commit |
| Scheduling | ✅ Configured | `content-feed-trigger` agent, 6-hour `scheduled_tasks` entry |
| Page build integration | ✅ Deployed | `load_page_sections_from_spec` reads sections from `site_specs.site_plan` |
| Classification enrichment | ✅ Deployed | `evaluate_news_feed` writes recommendation to classification spec |
| Discovery checks | ✅ Deployed | 4 checks: missing_news_sources, missing_news_section, stale_news_section, all_sources_erroring |
| CSS snippet | ✅ Deployed | `Latest News Grid` in `css_snippets`, uses `applies_to && $1::jsonb` |

---

## Architecture

### Pipeline Flow

```
Scheduled trigger (every 6 hours)
  → content-feed-trigger
      → finds sites with news_feed.recommended = true AND deployed pages
      → for each site:
          → content-feed-orchestrator
              1. seed_sources       — create content_sources if none exist
              2. check_has_sources  — skip if no sources (not recommended)
              3. dispatch_sources   — spawn feed-ingester per due source (async)
              4. spawn_triage       — spawn feed-triage agent
              5. run_triage         — score ingested items from PRIOR runs
              6. render_news_json   — produce latest-news.json (top 6 items)
              7. check_has_news     — skip commit if nothing rendered
              8. commit_news        — git commit → GitHub Actions → S3
              9. complete

Feed ingesters (async, fire-and-forget per source):
  generic agent receives dispatch message
    → spawns feed-ingester K8s job
      → route_by_type:
          rss:         fetch_rss → write_items → update_timestamps
          api_news:    fetch_llm_news → write_items → update_timestamps
          news_search: fetch_news_search → normalize_search → write_items → update_timestamps
          scrape:      fetch_scrape → normalize_scrape → write_items → update_timestamps

Results land in content_feed_items with status = 'ingested'
Next orchestrator run triages and renders them.
```

### Two-Pass Design

Dispatch fires ingesters that run **asynchronously**. Items they write arrive after the orchestrator finishes. Triage and render work on items from **prior** ingestion runs. This is intentional:

- Run 1: seeds sources, dispatches ingesters (items arrive async), triage finds nothing new, renders existing items
- Run 2: triage scores items from Run 1, render includes newly scored items, dispatch fires again
- Run N: steady state — each run triages prior items and dispatches for next time

### Data Ownership

| Data | Owner | Tables |
|------|-------|--------|
| Content sources config | seed_content_sources / manual | `content_sources` |
| Raw feed items | feed-ingester | `content_feed_items` |
| Relevance + credibility scores | feed-triage | `content_feed_items` (score, credibility, source_attribution) |
| News JSON file | render_news_section | `/data/latest-news.json` in git repo |
| News feed recommendation | evaluate_news_feed | `site_specs` (classification aspect) |
| Section lists | site planner / load_page_sections_from_spec | `site_specs.site_plan` (authoritative), `pages.sections` (synced) |

---

## Database Schema

### content_sources

One row per configured data source per site.

```
content_sources
├── id              UUID PK
├── site_id         UUID FK → sites(id) ON DELETE CASCADE
├── source_type     TEXT NOT NULL        — rss, news_search, api_news, scrape
├── name            TEXT NOT NULL        — human-readable, UNIQUE per site
├── entity_type     TEXT                 — NULL for news, set for entity sources
├── config          JSONB NOT NULL       — type-specific config (see below)
├── fetch_interval  INTERVAL             — default '6 hours'
├── last_fetched_at TIMESTAMPTZ
├── next_fetch_at   TIMESTAMPTZ          — NULL means "fetch now"
├── is_active       BOOLEAN              — default true
├── error_count     INT                  — for exponential backoff
├── last_error      TEXT
├── last_error_at   TIMESTAMPTZ
├── created_at      TIMESTAMPTZ
└── updated_at      TIMESTAMPTZ
```

Indexes: `idx_cs_site` (active by site), `idx_cs_due` (needing fetch), `idx_cs_site_name` (UNIQUE dedup).

### content_feed_items

```
content_feed_items
├── id                   UUID PK
├── site_id              UUID FK → sites
├── source_id            UUID FK → content_sources
├── external_id          TEXT                — GUID from RSS, URL, or hash
├── source_url           TEXT
├── source_title         TEXT
├── source_summary       TEXT
├── source_content       TEXT
├── source_published_at  TIMESTAMPTZ
├── relevance_score      FLOAT               — 0-100, set by triage
├── topics               JSONB               — tag array, set by triage
├── credibility          TEXT                — high/medium/low, set by triage
├── credibility_reason   TEXT                — why this credibility level
├── source_attribution   JSONB               — {original_source, found_via, source_tier}
├── entity_ids           UUID[]
├── duplicate_of         UUID
├── status               TEXT                — ingested/relevant/review/rejected/published/expired
├── work_item_id         UUID
├── created_at           TIMESTAMPTZ
├── updated_at           TIMESTAMPTZ
├── processed_at         TIMESTAMPTZ         — when triage scored it
└── ...
```

### Status lifecycle

```
ingested → relevant    (triage: score >= threshold, credibility ok)
ingested → review      (triage: score 20-49, held for manual check)
ingested → rejected    (triage: score < 20, flagged, or low credibility)
relevant → expired     (render_news_section: older than 30 days)
```

### Config shapes by source_type

**rss:**
```json
{"feed_url": "https://oilprice.com/rss/main", "max_items": 5}
```

**news_search:**
```json
{"query": "UK wholesale gas prices news", "num_results": 10}
```

**api_news (xAI with search):**
```json
{
    "provider": "xai",
    "model": "grok-4-1-fast",
    "prompt_template": "Search for the most important news from the last {{.hours}} hours about: wholesale gas prices, natural gas market...",
    "hours_lookback": 12,
    "max_items": 10,
    "search_tools": ["web_search", "x_search"]
}
```

**scrape:**
```json
{"url": "https://oilprice.com/Latest-Energy-News/World-News", "scrape_config": {"only_main_content": true}, "max_items": 10}
```

---

## Go Actions

### feed_actions.go

| Action | Sync/Async | Purpose |
|--------|------------|---------|
| `fetch_rss` | Sync | HTTP GET + XML parse (RSS/Atom), returns normalised items |
| `fetch_llm_news` | Sync | Calls xAI Responses API (web_search + x_search), OpenAI Responses API, or Perplexity chat completions |
| `write_feed_items` | Sync | Normalised items → `content_feed_items` with dedup by source_url |
| `load_due_sources` | Sync | Queries `content_sources WHERE next_fetch_at <= NOW()` |
| `update_source_timestamps` | Sync | Updates fetch timestamps, exponential backoff on errors |

**fetch_llm_news provider routing:**

| Provider | API | Endpoint | Model | Search |
|----------|-----|----------|-------|--------|
| xAI/Grok | Responses API | `api.x.ai/v1/responses` | `grok-4-1-fast` | web_search + x_search (real-time) |
| OpenAI | Responses API | `api.openai.com/v1/responses` | `gpt-4.1-mini` | web_search (Bing, real-time) |
| Perplexity | Chat completions | `api.perplexity.ai/chat/completions` | `sonar` | Built-in (always searches) |

All providers now do real-time web search. The old `grok-3-mini` via chat completions hallucinated URLs — replaced with the Responses API.

### Other action files

| File | Actions |
|------|---------|
| `feed_fetch_async_actions.go` | `fetch_scrape`, `fetch_news_search` (wrappers for existing adapters) |
| `feed_normalize_action.go` | `normalize_to_feed_items` (search results and scrape data → feed items) |
| `dispatch_feed_sources_action.go` | `dispatch_feed_sources` (queries due sources, produces Kafka messages per source) |
| `feed_triage_actions.go` | `apply_feed_scores`, `load_feed_items_for_triage` |
| `render_news_section_action.go` | `render_news_section` (produces latest-news.json from scored items) |
| `feed_news_recommendation_action.go` | `evaluate_news_feed` (deterministic vertical → news recommendation) |
| `seed_content_sources_action.go` | `seed_content_sources` (creates sources from classification spec) |
| `discovery_checks/check_news_feed.go` | 4 discovery checks |

14 registry entries total in `registry.go` under the FEED section.

---

## Agent Definitions

### content-feed-orchestrator

```
category:    orchestrator
agent_category: coordinator
timeout:     900s
workflow:    seed_sources → check_has_sources → dispatch_sources → spawn_triage
             → run_triage → render_news_json → check_has_news → commit_news → complete
input:       {site_id}
```

### content-feed-trigger

```
category:    orchestrator
agent_category: coordinator
timeout:     900s
workflow:    find_news_sites (SQL query) → check_has_sites → process_sites (loop, spawn+call orchestrator per site) → notify_scheduler → complete
input:       none (reads from DB)
scheduled:   every 6 hours via scheduled_tasks
```

### feed-ingester

```
category:    code-driven
agent_category: executor
timeout:     300s
workflow:    route_by_type → [fetch path] → write_items → update_timestamps → complete
input:       {site_id, source_id, source_type, source_config}
```

### feed-triage

```
category:    code-driven
agent_category: analyst
timeout:     300s
LLM:         claude-sonnet-4-6
workflow:    load_items → check_has_items → read_site_spec → score_relevance (LLM) → apply_scores → complete
input:       {site_id}
```

Triage prompt scores for **relevance** (0-100) and **credibility** (high/medium/low) with **source attribution** (original_source, found_via, source_tier). Source tiers: tier1_news (Reuters, BBC, FT), tier1_official (OPEC, regulators), tier2_trade (OilPrice, Rigzone), tier2_analysis (McKinsey), tier3_social (tweets), tier3_blog, unknown.

Items flagged for low credibility without corroboration, fabricated URLs, navigation links, or legal conflicts are rejected.

---

## Build Pipeline Integration

News flows through the standard build pipeline like any other section:

```
classifier → evaluate_news_feed → planner → webdesign → content writer → render → deploy
+ periodic: content-feed-orchestrator updates /data/latest-news.json every 6 hours
```

### Page sections from site_specs

The `page-build-handler` reads sections from `site_specs.site_plan` (authoritative) via `load_page_sections_from_spec`, falling back to `pages.sections`. This ensures `latest-news` appears in page builds when the planner includes it.

Flow: `load_existing_content → load_spec_sections → plan_sections → spawn_content_writer → ...`

### CSS

The `css_snippets` table has a `Latest News Grid` entry with `applies_to` containing `["latest-news"]`. The `render_css_from_spec` action queries `WHERE applies_to && $1::jsonb` to automatically include news CSS when the component is present.

### Content writer

Renders the `latest-news` template normally. At build time, items are empty — the template shows a placeholder. The first feed orchestrator run fills the JSON file with actual items.

---

## Operational Procedures

### Trigger manually

```bash
chmod +x trigger_feed_orchestrator.sh
./trigger_feed_orchestrator.sh
```

### Consumer offset reset (after chassis rebuild)

```bash
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=0
sleep 10
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group generic-requests-group \
  --topic system.agent.generic.requests \
  --reset-offsets --to-latest --execute
kubectl -n ai-persona-system scale deployment/agent-chassis --replicas=3
```

### Monitoring queries

```sql
-- Pipeline status for a site
SELECT status, COUNT(*) FROM content_feed_items
WHERE site_id = '<site-id>' GROUP BY status ORDER BY status;

-- Source health
SELECT name, source_type, last_fetched_at, next_fetch_at, error_count, last_error
FROM content_sources WHERE site_id = '<site-id>' AND is_active = true;

-- Latest orchestrator run
SELECT collected_data->'dispatch_result' as dispatch,
       collected_data->'news_render_result'->>'item_count' as rendered,
       collected_data->'seed_result'->>'has_sources' as has_sources,
       current_step, status, error
FROM orchestration_states
WHERE owner_agent_type = 'content-feed-orchestrator'
ORDER BY created_at DESC LIMIT 1;

-- Credibility breakdown
SELECT credibility, source_attribution->>'source_tier' as tier, COUNT(*)
FROM content_feed_items
WHERE site_id = '<site-id>' AND credibility IS NOT NULL
GROUP BY 1, 2 ORDER BY 3 DESC;
```

### Debugging

```bash
# Feed-ingester pod logs
kubectl -n ai-persona-system logs -l agent-type=feed-ingester --tail=100

# Content-feed-orchestrator pod logs
kubectl -n ai-persona-system logs -l agent-type=content-feed-orchestrator --tail=200

# Generic agent (receives dispatch messages, spawns ingesters)
kubectl -n ai-persona-system logs deploy/agent-chassis --tail=200 | grep -i "feed\|dispatch"

# Consumer group lag
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```

---

## Deployed SQL Files (in order)

| File | Status | What |
|------|--------|------|
| `026_content_sources_table.sql` | ✅ Run | Table, indexes, FK |
| `026b_agent_definitions_feed_pipeline.sql` | ✅ Run | feed-ingester, content-feed-orchestrator, feed-triage definitions |
| `026c_latest_news_component.sql` | ✅ Run | `latest-news` template in content_components |
| `026_stage2_deploy.sql` | ✅ Run | CSS snippet (applies_to), evaluate_news_feed enrichment |
| `026h_page_build_handler_amend.sql` | ✅ Run | load_spec_sections step in page-build-handler |
| `027i_feed_orchestrator_update.sql` | ✅ Run | Orchestrator: seed → triage → render → commit |
| `027j_content_feed_scheduling.sql` | ✅ Run | content-feed-trigger agent + scheduled_tasks entry |
| `027l_feed_credibility.sql` | ✅ Run | credibility/source_attribution columns + triage prompt update |

### Revert commands

```sql
-- Revert orchestrator workflow
UPDATE agent_definitions
SET default_config = (SELECT default_config FROM agent_def_content_feed_orch_backup_20260402 LIMIT 1),
    updated_at = NOW()
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;

-- Revert page-build-handler
UPDATE agent_definitions
SET default_config = (SELECT default_config FROM agent_def_page_build_handler_backup_20260402 LIMIT 1),
    updated_at = NOW()
WHERE type = 'page-build-handler' AND deleted_at IS NULL;

-- Revert triage prompt
UPDATE agent_definitions
SET default_config = (SELECT default_config FROM agent_def_feed_triage_backup_20260403 LIMIT 1),
    updated_at = NOW()
WHERE type = 'feed-triage' AND deleted_at IS NULL;
```

---

## Issues Resolved During Deployment

| Issue | Root cause | Fix |
|-------|-----------|-----|
| Orchestrator not picked up after chassis rebuild | Stale Kafka consumer group offsets | Offset reset procedure (documented above) |
| Dispatch: 4 errors, 0 dispatched | Missing `sender_agent_type` and `step_name` in Kafka headers | Added to `dispatch_feed_sources_action.go` headers map |
| Ingesters spawn but call_ingester fails | `input_mapping` values were literals, not dot-paths | Changed to `input_data.site_id` etc. (paths into collected_data) |
| Triage spawn fails: "no spawned feed-triage agent found" | Missing `spawn_agent` step before `call_agent` | Added `spawn_triage` step to orchestrator workflow |
| Triage finds 0 items | Timing: items ingested async, triage runs same cycle | By design — two-pass: ingest now, triage next run |
| Grok returns hallucinated 2023 URLs | `grok-3-mini` via chat completions has no web access | Switched to `grok-4-1-fast` via Responses API with web_search + x_search tools |
| Triage scores not applied | `execute_llm_prompt` returns `type: "text"` (string), not parsed array | Added string→JSON parse fallback in `ApplyFeedScoresAction` |
| Render: 0 items | Existing items older than 72h `max_age_hours` | Items age out naturally; fresh ingestion brings new ones |
| `content_feed_items.updated_at` not found | Column doesn't exist on that table | Fixed expire query in `render_news_section_action.go` |
| `sites.deleted_at` not found | Column doesn't exist on sites table | Removed from trigger query |
| `processing_type` column error | Should be `category` in agent_definitions | Fixed INSERT |
| `agent_type` column error in scheduled_tasks | Should be `target_agent_type` | Fixed INSERT |
| `extractStringSlice` signature mismatch | Existing function takes `(map, key)` not `(value)` | Changed calls to `extractStringSlice(newsFeed, "source_types")` |
| Scrape ingests nav links | `/Energy/`, `/Geopolitics/` are category pages not articles | Triage rejects them; `isNavigationLink` filter could be tightened |

---

## Environment Variables

| Variable | Required for | Where set |
|----------|-------------|-----------|
| `XAI_API_KEY` / `GROK_API_KEY` | api_news sources (xAI provider) | `personae-default-secrets` or agent def `env_vars` |
| `OPENAI_API_KEY` | api_news sources (OpenAI provider) | Same treatment needed |
| `PERPLEXITY_API_KEY` | api_news sources (Perplexity provider) | Same treatment needed |
| `ANTHROPIC_API_KEY` | feed-triage LLM scoring | `personae-default-secrets` (already in spawn_actions.go) |

---

## Expansion Roadmap

### Three Tiers

| Tier | What | Where | Status |
|------|------|-------|--------|
| 1. Homepage snippets | Title + summary + link via JSON | Homepage section | ✅ Working |
| 2. Insights section | Curated, rewritten articles with own pages | `/insights/` with nav | Future |
| 3. Research analysis | Multi-source analysis with timelines, graphs | `/insights/` (rich pages) | Future |

### Content Vision (from 027b)

For each article:
- What is it about, what points is it making (facts vs opinion vs noise)
- Who is mentioned and in what capacity
- Prioritise points against site criteria
- Supporting evidence from other sources
- Top 3 actionable insights (analysis, reasoning, tables, infographics)
- Timeline and labelling
- Related videos (YouTube, verified quality)
- Original graphics, graphs, timelines

### What's Not Built Yet

| Component | Blocked on |
|-----------|-----------|
| `article-rewriter` agent — rewrites articles in site voice | Triage credibility working first |
| `feed-publisher` agent — publishes rewritten articles as blog posts | Article-rewriter |
| `feed-lifecycle` agent — expires old items | Not blocking anything |
| Scrape link quality filtering (category-level path check) | Low priority, triage handles |
| Content-quality-auditor news check (post-display safety net) | latest-news component on pages |
| Batch API integration for LLM calls (50% cost saving) | Separate design task |

---

## Resolved Decisions

1–42: See previous pipeline docs (preserved in repo).

43. **JSON for homepage news**, not page rerender. Decouples news updates from page structure.
44. **Three-tier expansion**: homepage snippets → insights section → research analysis. Each independently valuable.
45. **Event timeline table** for research continuity (future).
46. **Source seeding from classification spec.** `SeedContentSourcesAction` creates `content_sources` automatically from `content_features.news_feed.vertical_keywords` and `source_types`. No manual SQL needed for new sites.
47. **Triage scores credibility + provenance, not just relevance.** Source attribution chain (original_source, found_via, source_tier) stored per item. Low-credibility unverified claims are rejected.
48. **Grok via Responses API with web_search + x_search**, not chat completions. Real-time search prevents hallucinated URLs. X/Twitter access provides broadest news coverage.
49. **Scheduled trigger via content-feed-trigger agent.** Runs every 6 hours, finds sites with `news_feed.recommended = true`, dispatches orchestrator per site.
50. **Page sections read from site_specs.site_plan** (authoritative), synced to pages.sections. Ensures planner decisions propagate to page builds.
51. **Two-pass ingestion/triage design.** Ingesters run async (fire-and-forget). Triage and render work on items from prior runs. Avoids timing dependencies.
52. **Dispatch uses Kafka header validation.** `sender_agent_type` and `step_name` required by `ValidateOutgoingMessage`. Dispatch action includes both.
53. **Call_agent input_mapping uses paths, not literals.** Values in `input_mapping` are resolved as dot-paths against `collected_data`. Literal values from dispatch go into `message.input_data`, then `input_mapping` references `input_data.field_name`.
-e 

---

## Content Diversity & Research Pipeline (from 036b)

## Problem

The homepage news feed draws from 4 sources (RSS, news_search, api_news/Grok, scrape) but in practice one source dominates because RSS is the most frequent and reliable. All sources search the same broad topic space independently, so they find the same stories. The result is 6 items from OilPrice rather than a mix of perspectives.

## Current Source Setup (gaswholesalers.com)

| Source | Type | Frequency | Typical yield |
|--------|------|-----------|---------------|
| OilPrice Energy News | rss | 6h | 5 items/run |
| Gas wholesale energy news search | news_search | 6h | 2-5 items/run |
| Grok energy news | api_news (xAI Responses API, web_search + x_search) | 6h | 4-5 items/run |
| OilPrice Latest News Scrape | scrape | 6h | ~20 raw, filtered by triage |

## Root Cause

Each source independently fetches "gas/energy news" without knowing what the others found. The render query picks the highest-scored most-recent items, which cluster around the same breaking stories from the most prolific source.

## Approach

### 1. Topic-Focused Source Splitting

Replace the single broad Grok source with multiple specialised sources, each with a different prompt template targeting a different angle of the vertical:

- **Geopolitical/supply chain** — OPEC decisions, pipeline disruptions, sanctions, shipping routes
- **UK/EU regulatory** — Ofgem, energy policy, carbon pricing, storage obligations
- **Market data/pricing** — spot prices, futures, spreads, seasonal patterns
- **Emerging/transition** — hydrogen, biogas, renewable gas, LNG infrastructure

Each is a `content_sources` row with a different `prompt_template` in its config. No code changes needed — just SQL inserts.

### 2. Coverage Gap Analysis

A pre-fetch step that reads existing feed items and identifies what's already covered before dispatching ingesters. The output is a list of saturated topics and under-reported angles.

This becomes a new step in the content-feed-orchestrator workflow, between `seed_sources` and `dispatch_sources`. The dispatch step passes coverage gaps as context to the api_news prompt templates.

### 3. Multi-Language/Region Discovery

Grok's X search can find content in any language. Regional sources surface stories that English-language feeds miss:

- German energy policy (drives EU gas pricing)
- Japanese LNG buyer behaviour (spot market signals)
- Arabic Gulf production decisions
- Norwegian pipeline/storage data

Each regional search is a separate `content_sources` entry. Triage needs a translation/summary sub-task — the LLM scores relevance from a translated summary, and the rendered item links to the original source with a note like "Originally reported by Handelsblatt (Germany)."

### 4. Triage Diversity Scoring

The triage prompt gets additional context: titles of items already in the display set. Items covering a different angle score higher. This is a prompt change to the feed-triage agent's `score_relevance` step.

### 5. Source Chain Provenance

When Grok finds a tweet that links to a Reuters article, the `source_attribution` chain records both the discovery path (X/@energy_analyst) and the original source (Reuters). This supports:

- Credibility assessment (tweet found it, Reuters wrote it)
- Deduplication (same story found via different paths)
- Display diversity (show the Reuters attribution, not "found on Twitter")

Already partially implemented in the triage scoring — `source_attribution.original_source`, `found_via`, `source_tier` fields exist in `content_feed_items`.

### 6. Render Interleaving

The render query uses `ROW_NUMBER() OVER (PARTITION BY source_id)` to interleave items from different sources. With 4+ sources and 6 display slots, each source contributes at most 2 items. Already implemented in the current `render_news_section_action.go`.

With topic-focused sources, this naturally produces topical diversity too — one geopolitical item, one regulatory, one pricing, etc.

---

## Original Research & Writing Pipeline

The sections above improve diversity of *sourced* news. This section describes producing *original* content — researched, written, and published articles that serve specific readership segments better than generic trade press.

### Readership Model

"Gas industry professionals" is too broad. The actual readership segments for a site like gaswholesalers.com might be:

- **Procurement managers** — price movements, contract terms, supply reliability
- **Operations/logistics** — infrastructure status, pipeline capacity, storage levels, delivery schedules
- **Finance/trading** — futures spreads, hedging strategies, regulatory cost impacts
- **C-suite/strategy** — market structure changes, M&A, policy shifts, energy transition

Each segment wants the same underlying events framed differently. Hormuz closure → procurement sees supply risk, trading sees price opportunity, operations sees delivery disruption, strategy sees long-term diversification argument.

The classifier already produces `target_audience` and `industry`. This needs extending to describe readership segments with enough detail to drive content targeting — what each segment cares about, what decisions they make, what data they need. If the current classification doesn't have this depth, the classifier prompt can be enriched to produce it, or a separate `readership_profile` spec aspect can be created.

For a given vertical, the number of meaningful angle combinations is roughly:

- ~10-20 recurring topic areas per vertical
- ~3-4 readership segments
- ~50-80 specific angle combinations per site

Across all sites and verticals, this reaches hundreds. The structure is the same everywhere — the content is parameterised by classification and readership.

### Research Agent Pipeline

For each news angle, a multi-step investigation rather than a single "find news about X" prompt:

```
Topic monitor detects signal (price moved, tweet from OPEC, regulatory filing)
  → Research agent: gather context from multiple sources
      → What happened? (fact gathering — multiple sources)
      → What's the history? (has this happened before, what was the impact)
      → Who said what? (official statements, analyst quotes, X discourse)
      → What are the numbers? (price data, volume data, contract data)
  → Writer agent: produce article draft targeted at readership segment
      → Frame for the audience (procurement angle vs trading angle)
      → Include actionable insight ("if you're hedging Q3, consider...")
      → Source attribution throughout
      → Reasoning explained — why we think X follows from Y
  → Eval agent: score against quality criteria
      → Factual accuracy (claims match cited sources?)
      → Audience fit (would a procurement manager find this useful?)
      → Originality (does this add value beyond summarising existing articles?)
      → Tone/compliance (no financial advice, no speculation presented as fact)
      → Structure quality (news hook → data → context → implications)
  → Publish or reject
```

The research depth is the key differentiator from current feed ingestion. Current api_news does one LLM call. The research agent does a multi-step workflow with verification.

### Continuous Timelines

Each topic thread maintains a running timeline — a persistent, evolving dataset that builds historical credibility over time. For a gas wholesale site, a thread might be:

- **NBP spot price** — daily data point, annotated with the news events that moved it
- **UK storage levels** — weekly data from Grid data, annotated with demand/supply context
- **LNG cargo arrivals** — shipping data, annotated with geopolitical context
- **Regulatory calendar** — Ofgem consultations, policy deadlines, annotated with market reactions

Each timeline produces:

1. **An interactive graph** — price/volume over time with labelled spikes and troughs. Each label links to the article that covered the event. The graph itself becomes valuable reference material that readers return to.

2. **Annotated history** — every data point has context: "Price rose 8% on 14 March — OPEC announced production cut of 500k bpd (see our coverage)." Over months, this builds a searchable, cited record that no single article provides.

3. **Pattern recognition** — as the dataset grows, the system can identify patterns: "The last 3 times UK storage dropped below 40%, NBP spot rose 12-18% within 2 weeks." These become the basis for scenario analysis.

### Scenario Analysis

Rather than predicting prices (which we shouldn't do), we can present structured "if/then" analysis:

- "If OPEC maintains current production targets, and UK storage stays below seasonal average, here are three scenarios for Q3 procurement costs" — with ranges, not point predictions
- "The EU carbon border adjustment takes effect in October. Here's what it means for gas-to-power economics under current vs proposed pricing"
- Decision framing: "If you're a procurement manager deciding between a fixed Q3 contract and spot exposure, here are the factors to weigh" — with current data, historical parallels, and cited analyst views

This is opinion-adjacent but grounded in data and cited reasoning. The eval agent needs to enforce the boundary: structured analysis with explained reasoning and cited sources is fine; confident directional predictions are not.

### Quality Framework

Every piece of original content must meet:

- **All claims cited** — no unsourced assertions. Every factual claim links to its source.
- **Reasoning explained** — "we think X because of Y and Z" not just "X is likely"
- **Multiple perspectives** — if analysts disagree, show both views
- **Audience-appropriate** — written for the readership segment, not for SEO robots
- **Structured consistently** — news hook → supporting data → expert context → implications for reader

Quality templates can be derived from studying how OilPrice, Argus Media, ICIS, and similar trade publications structure their articles. Not copying content — studying the *pattern*: lead structure, data placement, quote integration, conclusion framing. The eval agent scores against this pattern.

As reasoning models improve, the research depth and analytical quality improve automatically. The pipeline structure stays the same — better models produce better research and writing within the same workflow.

### Engagement Measurement

To know whether original content is working, we need to measure reader behaviour:

- **Read depth** — do readers scroll through the full article or bounce at the headline?
- **Return visits** — do timeline/graph pages get repeat visits? (indicates reference value)
- **Click-through from homepage** — which article framings get clicked? (indicates headline/segment fit)
- **Time on page** — proxy for engagement quality
- **Cross-article navigation** — do readers follow "related" links? (indicates topic interest mapping)

This data feeds back into the system:

- Topics with high engagement get more frequent coverage
- Readership segments with low engagement may need reframing or different angle selection
- Article structures that perform well become reinforced in the writer prompts
- Timelines that get return visits get more frequent data updates

The engagement pipeline is a separate instrumentation project (client-side analytics → data collection → feedback to content planning). It doesn't block the research/writing pipeline but is essential for calibrating quality over time.
