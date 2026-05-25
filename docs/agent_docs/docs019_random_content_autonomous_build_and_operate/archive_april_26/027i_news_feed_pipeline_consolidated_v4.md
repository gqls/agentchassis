# 027 — News Feed Pipeline (Consolidated, 2026-04-04)

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
