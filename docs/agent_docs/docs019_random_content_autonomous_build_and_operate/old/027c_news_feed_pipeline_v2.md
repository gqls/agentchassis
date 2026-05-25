# 026 — News Feed Pipeline

Implementation guide for the content feed pipeline. First test site: gaswholesalers.com (energy vertical). Architecture is vertical-agnostic — boxing, finance, tech all use the same pipeline with different source configs.

Implements Phase 3 from `006_entity_feeds_livedata_implementation.md`. See `007_entity_feeds_livedata_architecture.md` for cross-vertical design rationale.

---

## Current Status

All four source types tested and confirmed working (March 2026):

| Source type | Action path | Test result | Items |
|-------------|-------------|-------------|-------|
| `rss` | `fetch_rss` (sync HTTP+XML) | Working, dedup confirmed | 5 written, 5 skipped on re-run |
| `scrape` | `fetch_scrape` → firecrawl adapter → `normalize_scrape` (async) | Working, dedup confirmed | 20 written, 20 skipped on re-run |
| `news_search` | `fetch_news_search` → web search adapter → `normalize_search` (async) | Working | 5 written |
| `api_news` | `fetch_llm_news` to xAI/Grok (sync HTTP) | Working | 4 written |

Test site: gaswholesalers.com (`5fe15466-4e2e-4ff2-981e-98c1b7074002`). Sources used: OilPrice RSS, OilPrice scrape, UK gas prices news search, Grok energy news.

Downstream pipeline (triage → rewrite → publish) not yet built.

---

## Relationship to Existing Docs

| Doc | Covers | Relationship |
|-----|--------|-------------|
| `002d_system_architecture_v4.md` §6 | Feed pipeline concept, agent family, lifecycle | Architecture source of truth — this doc is the build detail |
| `006_entity_feeds_livedata_implementation.md` Phase 3 | Step-by-step plan for feed agents | Original plan — this doc records what was actually built and where it differs |
| `007_entity_feeds_livedata_architecture.md` §Phase 3 | Feed pipeline agents, rewriter prompts, source lists | Design rationale — still accurate, referenced here |

---

## What Changed From the 006 Plan

The 006 plan specified a monolithic `ingest_feed_source` action that handled RSS parsing and web scraping in a single function, branching internally on `source_type`. What we built instead:

| 006 plan | What was built | Why |
|----------|---------------|-----|
| Single `ingest_feed_source` action | Separate `fetch_rss`, `fetch_llm_news`, `fetch_scrape`, `fetch_news_search`, `normalize_to_feed_items`, `write_feed_items` actions | Keeps each action focused. RSS parsing, LLM HTTP calls, and DB writes are different concerns. |
| RSS + scrape only | RSS + scrape + news_search + api_news (Grok/xAI) | Added LLM news providers (xAI/Grok has strong real-time news) and reuse of existing `web_search` action with `search_type: "news"` |
| `enabled` column | `is_active` column | Matches pattern used elsewhere in the codebase (`content_components.is_active`, etc.) |
| `url` column on content_sources | `config` JSONB (contains `feed_url`, `url`, `query`, etc.) | Different source types need different config shapes. A single `url` doesn't work for news_search (needs query) or api_news (needs model + prompt) |
| `check_interval` column name | `fetch_interval` column name | Avoids confusion with health checks |
| All sources loop inside one agent | Orchestrator spawns separate feed-ingester per source | Cleaner logs, failure isolation, follows the "every agent is an orchestrator" principle |
| `feed-ingester` loops internally | `conditional_route` in workflow branches by source_type | Keeps workflow simple, each branch follows a straight line |
| Direct use of `firecrawl_scrape`/`web_search` via workflow config paths | Wrapper actions `fetch_scrape`/`fetch_news_search` that read source_config in Go | Path resolver doesn't support array indexing. Wrappers follow dev guide rule #10: complexity in Go, not workflow config. |

The table schema also adds `entity_type` (from 006 Phase 2 spec) so the same `content_sources` table serves both entity data sources and news feed sources, as designed.

---

## What Already Existed

| Asset | Status | Used by this pipeline |
|-------|--------|----------------------|
| `content_feed_items` table | Deployed | Yes — receives all ingested items |
| `web_search` action | Deployed | Yes — `fetch_news_search` delegates to it with `search_type: "news"` |
| `firecrawl_scrape` / `WebscrapeAction` | Deployed | Yes — `fetch_scrape` delegates to it |
| `execute_llm_prompt` action | Deployed | Yes — used by feed-triage for relevance scoring |
| `query_database` action | Deployed | Yes — feed-triage loads pending items |
| `conditional_route` action | Deployed | Yes — feed-ingester routes by source_type |
| Spawn+call pattern via generic agent | Deployed | Yes — orchestrator dispatches ingesters this way |
| `sites.settings.maintenance_profile.content_feed` | Deployed | Yes — orchestrator checks `enabled` flag before dispatching |

---

## New Database Objects

### `content_sources` table

**File:** `026_content_sources_table.sql`

One row per configured data source per site. Scheduling is per-source — a fast-moving RSS feed can check every 2 hours while a scrape target checks every 6.

```
content_sources
├── id              UUID PK
├── site_id         UUID FK → sites(id) ON DELETE CASCADE
├── source_type     TEXT NOT NULL        -- rss, news_search, api_news, scrape, api_data
├── name            TEXT NOT NULL        -- human-readable, unique per site
├── entity_type     TEXT                 -- NULL for news, set for entity sources
├── config          JSONB NOT NULL       -- type-specific config (see below)
├── fetch_interval  INTERVAL             -- default '6 hours'
├── last_fetched_at TIMESTAMPTZ
├── next_fetch_at   TIMESTAMPTZ          -- NULL means "fetch now"
├── is_active       BOOLEAN              -- default true
├── error_count     INT                  -- for exponential backoff
├── last_error      TEXT
├── last_error_at   TIMESTAMPTZ
├── created_at      TIMESTAMPTZ
└── updated_at      TIMESTAMPTZ
```

Indexes: `idx_cs_site` (active sources by site), `idx_cs_due` (sources needing fetch), `idx_cs_entity` (entity-linked sources), `idx_cs_site_name` (UNIQUE for dedup).

FK added: `content_feed_items.source_id → content_sources.id` (was dangling).

### Config shapes by source_type

**rss:**
```json
{"feed_url": "https://oilprice.com/rss/main", "max_items": 5}
```

**news_search** (uses existing web search adapter):
```json
{"query": "UK wholesale gas prices news", "num_results": 5}
```

**api_news** (xAI/Grok, OpenAI, Perplexity — OpenAI-compatible chat completions):
```json
{
    "provider": "xai",
    "model": "grok-3",
    "prompt_template": "What are the latest ... Return as a JSON array with fields: title, summary, source_url (if known), source_name, published_at (ISO format).",
    "hours_lookback": 24,
    "max_items": 5
}
```

**scrape** (uses existing firecrawl adapter):
```json
{"url": "https://oilprice.com/Latest-Energy-News/World-News", "scrape_config": {"only_main_content": true}, "max_items": 10}
```

**api_data** (future — structured APIs like BoE rates, ticket prices):
```json
{"endpoint": "https://api.bankofengland.co.uk/...", "method": "GET", "response_path": "result.data", "data_type": "mortgage_rate"}
```

**Config format note:** Use flat strings (`"url"`, `"query"`) not arrays (`"urls"`, `"queries"`). One source = one URL/query = one set of error tracking and scheduling. Multiple queries for the same vertical become separate `content_sources` rows. The `fetch_scrape` and `fetch_news_search` Go actions handle both formats (flat or array, taking first element), but flat is recommended.

### Seed function

```sql
SELECT seed_boxing_sources('<site-uuid>');
```

Creates 7 sources for a boxing site: 2 news_search, 3 RSS (BoxingScene, ESPN, BBC Sport), 1 api_news (Grok), 2 scrape (BoxRec, Ring Magazine). Sets `next_fetch_at = now()` so they're picked up on the first orchestrator run.

---

## New Go Actions

All in `platform/orchestration/actions/`. Each follows the standard pattern: `ActionInputSpec` + `init()` registration + `XxxAction` function. No extra exported symbols.

### feed_actions.go — 5 actions

| Action | Category | Sync/Async | Purpose |
|--------|----------|------------|---------|
| `fetch_rss` | feed | Sync (HTTP GET + XML parse) | Fetches RSS/Atom feed, returns normalised items array |
| `fetch_llm_news` | feed | Sync (HTTP POST to xAI/OpenAI-compatible API) | Asks LLM for recent news, parses JSON response |
| `write_feed_items` | feed | Sync (DB writes) | Normalised items → `content_feed_items` with dedup |
| `load_due_sources` | feed | Sync (DB read) | Queries `content_sources` for sources past `next_fetch_at` |
| `update_source_timestamps` | feed | Sync (DB write) | Updates `last_fetched_at`, `next_fetch_at`, error tracking with exponential backoff |

**fetch_rss details:**
- Tries RSS XML first, falls back to Atom XML parsing
- Handles common date formats (RFC1123Z, RFC3339, RFC1123)
- 30s HTTP timeout, 5MB body limit
- Returns `{items: [...], source_id, feed_url, item_count, fetched_at}`
- On HTTP error: returns empty items with error field (non-fatal)

**fetch_llm_news details:**
- Calls OpenAI-compatible chat completions endpoint
- Provider resolution: `xai`/`grok` → `api.x.ai`, `openai` → `api.openai.com`, `perplexity` → `api.perplexity.ai`
- API key from env: checks `XAI_API_KEY` first, falls back to `GROK_API_KEY`
- System prompt instructs JSON-only response
- Strips markdown code fences before parsing
- 60s HTTP timeout
- Generates stable `external_id` from title hash when no URL available

**write_feed_items details:**
- `items` is optional in the InputSpec — returns `{written: 0, skipped: 0}` gracefully when no items (scrape of a page with no extractable links)
- Dedup by `source_url` per site (existence check)
- Handles nullable `source_id`
- Items with no title AND no URL are skipped

**update_source_timestamps details:**
- Success: resets `error_count` to 0, `next_fetch_at = NOW() + fetch_interval`
- Failure: `error_count += 1`, backoff `next_fetch_at = NOW() + (fetch_interval * LEAST(error_count + 1, 4))`
- Caps backoff at 4x the normal interval

### feed_fetch_async_actions.go — 2 actions

| Action | Category | Sync/Async | Purpose |
|--------|----------|------------|---------|
| `fetch_scrape` | feed | Async (delegates to WebscrapeAction) | Reads URL from source_config in Go, delegates to existing firecrawl adapter |
| `fetch_news_search` | feed | Async (delegates to WebSearchAction) | Reads query from source_config in Go, delegates to existing web search adapter with search_type=news |

These follow the same wrapper pattern as `FirecrawlScrapeAction` (which sets config then calls `WebscrapeAction`). They read `source_config` from collected_data in Go rather than relying on workflow config path threading.

Both handle flexible config shapes:
- `fetch_scrape`: accepts `"url": "..."` (string) or `"urls": [...]` (array, takes first), also tries `"feed_url"`
- `fetch_news_search`: accepts `"query": "..."` (string) or `"queries": [...]` (array, takes first), also tries `"topic"`

### feed_normalize_action.go — 1 action

| Action | Category | Purpose |
|--------|----------|---------|
| `normalize_to_feed_items` | feed | Transforms adapter response into `{title, url, summary, external_id, published_at}` format for `write_feed_items` |

Two modes via `source_format` config:
- `"search"`: reads from `search_results.results` array, maps `snippet` → `summary`
- `"scrape"`: navigates firecrawl `response.data` wrapper, extracts `links` array from listing pages, filters navigation/social links via `isNavigationLink()`. Falls back to single-item mode if no usable links found.

**Scrape link filtering note:** The `isNavigationLink` filter removes anchors, javascript:, mailto:, common non-article paths (/privacy, /contact, /login, etc.), and social media domains. It does not yet filter category-level navigation links (e.g. `/Energy/Natural-Gas/`). This was observed in testing — the OilPrice scrape picked up 20 links including category sections. A minimum path depth check or content-vs-nav heuristic would improve this.

### dispatch_feed_sources_action.go — 1 action

| Action | Category | Purpose |
|--------|----------|---------|
| `dispatch_feed_sources` | feed | The orchestrator's core: queries due sources, produces spawn+call messages to generic agent per source |

Fire-and-forget dispatch pattern. For each due source: builds a spawn+call inline workflow message targeting `feed-ingester`, produces to `system.agent.generic.requests`. Optimistically updates `next_fetch_at` to prevent re-dispatch.

### feed_triage_actions.go — 2 actions

| Action | Category | Purpose |
|--------|----------|---------|
| `load_feed_items_for_triage` | feed | Loads unscored items (`status = 'ingested'`) joined with source metadata. Returns `{items, count, site_id}` |
| `apply_feed_scores` | feed | Reads LLM-produced scores, updates `content_feed_items` with `relevance_score`, `status`, `topics`, `processed_at` |

**apply_feed_scores details:**
- Input: scores array from LLM (id, score, reason, topics, flagged)
- Status thresholds (configurable per site, default 50):
    - `score >= threshold` → status = `relevant` (displays on site)
    - `score 20..threshold` → status = `review` (held)
    - `score < 20` → status = `rejected` (never displays)
    - `flagged: true` → always `rejected` (values/legal conflict)
- Only updates items with `status = 'ingested'` (won't re-score already triaged items)
- Threshold override from `sites.settings.maintenance_profile.content_feed.relevance_threshold`

**Two-layer conformance checking:**
1. **Pre-display gate (feed-triage):** LLM scores items against full site spec (identity, classification, content_direction, legal_rules). Only items with `status = 'relevant'` appear on pages.
2. **Post-display audit (improvement loop):** content-quality-auditor checks displayed news items are appropriate during regular audit cycle. Creates findings → reject bad items → rerender.

---

## New Agent Definitions

**File:** `026b_agent_definitions_feed_pipeline.sql`

### feed-ingester

| Field | Value |
|-------|-------|
| type | `feed-ingester` |
| agent_category | `executor` |
| status | `experimental` |
| input_contract | `{"required": ["site_id", "source_id", "source_type", "source_config"]}` |
| output_contract | `{"produces": ["write_result", "timestamp_result"]}` |
| processing_mode | `task` |
| timeout | 300s |

Workflow: `route_by_type → [fetch_rss | fetch_llm_news | search_news→normalize_search | scrape_source→normalize_scrape] → write_items → update_timestamps → complete`

The `conditional_route` on `source_type` branches to the appropriate fetch path. All paths converge at `write_items`. No LLM calls — this agent is purely mechanical.

### content-feed-orchestrator

| Field | Value |
|-------|-------|
| type | `content-feed-orchestrator` |
| agent_category | `coordinator` |
| status | `experimental` |
| input_contract | `{"required": ["site_id"]}` |
| output_contract | `{"produces": ["dispatch_result"]}` |
| processing_mode | `orchestrator` |
| timeout | 600s |

Workflow: `dispatch_sources → complete`. All complexity is in the Go action.

### feed-triage

| Field | Value |
|-------|-------|
| type | `feed-triage` |
| agent_category | `analyst` |
| status | `experimental` |
| input_contract | `{"required": ["site_id"], "optional": ["relevance_threshold"]}` |
| output_contract | `{"produces": ["triage_result"]}` |
| ai_service | `claude-sonnet-4-6`, `api_key_env_var: "ANTHROPIC_API_KEY"` |
| processing_mode | `task` |
| timeout | 300s |

Workflow: `load_items → check_has_items → read_site_spec → score_relevance → apply_scores → complete`

Loads all site spec aspects before scoring so the LLM has identity, classification, content_direction, and legal_rules context. The `check_has_items` step uses `evaluate_condition` with `condition_field: "pending_items.count"` — routes to `complete` if count is 0, otherwise proceeds to read spec and score.

The LLM prompt instructs the model to score 0-100 and flag items that conflict with site values or legal constraints. `apply_feed_scores` then updates rows: relevant (shows), review (held), rejected (never shows).

---

## Registry Entries

**File:** `feed_registry_entries.go` — add to `GlobalActionRegistry` and `LocalActions` in `registry.go`.

11 actions total:

```go
// FEED section in GlobalActionRegistry
"fetch_rss":                   { Handler: FetchRSSAction,                Category: "feed", IsLocal: true }
"fetch_llm_news":              { Handler: FetchLLMNewsAction,            Category: "feed", IsLocal: true }
"write_feed_items":            { Handler: WriteFeedItemsAction,          Category: "feed", IsLocal: true }
"load_due_sources":            { Handler: LoadDueSourcesAction,          Category: "feed", IsLocal: true }
"update_source_timestamps":    { Handler: UpdateSourceTimestampsAction,  Category: "feed", IsLocal: true }
"normalize_to_feed_items":     { Handler: NormalizeToFeedItemsAction,    Category: "feed", IsLocal: true }
"dispatch_feed_sources":       { Handler: DispatchFeedSourcesAction,     Category: "feed", IsLocal: true }
"fetch_scrape":                { Handler: FetchScrapeAction,             Category: "feed", IsLocal: true }
"fetch_news_search":           { Handler: FetchNewsSearchAction,         Category: "feed", IsLocal: true }
"apply_feed_scores":           { Handler: ApplyFeedScoresAction,         Category: "feed", IsLocal: true }
"load_feed_items_for_triage":  { Handler: LoadFeedItemsForTriageAction,  Category: "feed", IsLocal: true }
```

All 11 also go in `LocalActions` map.

---

## Pipeline Flow

```
K8s CronJob (or manual trigger)
  → content-feed-orchestrator (per site)
      → dispatch_feed_sources action
          → queries content_sources WHERE next_fetch_at <= NOW()
          → for each due source:
              produces message → system.agent.generic.requests
                → generic agent runs inline workflow:
                    1. spawn_agent(feed-ingester)
                    2. call_agent(feed-ingester) with {site_id, source_id, source_type, source_config}
                    3. complete
                      → feed-ingester runs its own workflow:
                          route_by_type
                            → rss:         fetch_rss → write_items → update_timestamps → complete
                            → api_news:    fetch_llm_news → write_items → update_timestamps → complete
                            → news_search: fetch_news_search → normalize_search → write_items → update_timestamps → complete
                            → scrape:      fetch_scrape → normalize_scrape → write_items → update_timestamps → complete

Results land in content_feed_items with status = 'ingested'

Then (after each ingestion cycle or on schedule):
  feed-triage (per site)
    → load unscored items
    → load site spec (identity, classification, content_direction, legal_rules)
    → LLM scores each item against spec
    → apply_feed_scores: relevant / review / rejected
    → Only status = 'relevant' items appear on pages
```

Downstream (not yet built): `latest-news` component → scheduled rerender → improvement loop audit

---

## Data Flow Through content_feed_items

```
Status lifecycle:
  ingested → relevant    (feed-triage scores above threshold)
  ingested → rejected    (feed-triage scores below threshold)
  ingested → duplicate   (dedup check finds matching source_url)
  relevant → published   (article-rewriter + feed-publisher complete)
  published → expired    (feed-lifecycle ages out old items)
```

Key columns and who sets them:

| Column | Set by | When |
|--------|--------|------|
| `source_title`, `source_summary`, `source_url` | write_feed_items | Ingestion |
| `source_id` | write_feed_items | Ingestion (links to content_sources row) |
| `external_id` | write_feed_items | Ingestion (GUID from RSS, URL, or hash) |
| `source_published_at` | write_feed_items | Ingestion (from RSS pubDate or LLM) |
| `relevance_score` | feed-triage | Triage (LLM score 0-100) |
| `topics` | feed-triage | Triage (JSONB array of topic tags) |
| `status` | Various | See lifecycle above |
| `work_item_id` | article-rewriter | Links to site_work_items when article is being written |
| `published_page_id` | feed-publisher | Links to the deployed blog post page |
| `published_at` | feed-publisher | When the blog post went live |
| `expires_at` | feed-lifecycle | When to archive |

---

## Deployment

### Files

**Go files → `platform/orchestration/actions/`:**

| File | Actions |
|------|---------|
| `feed_actions.go` | `fetch_rss`, `fetch_llm_news`, `write_feed_items`, `load_due_sources`, `update_source_timestamps` |
| `feed_normalize_action.go` | `normalize_to_feed_items` |
| `feed_fetch_async_actions.go` | `fetch_scrape`, `fetch_news_search` |
| `dispatch_feed_sources_action.go` | `dispatch_feed_sources` |
| `feed_triage_actions.go` | `apply_feed_scores`, `load_feed_items_for_triage` |

**Registry → merge into `registry.go`:** 11 entries in `GlobalActionRegistry` under `FEED` section + 11 entries in `LocalActions` map.

**SQL:**

| File | What |
|------|------|
| `026_content_sources_table.sql` | Table, indexes, FK, seed function |
| `026b_agent_definitions_feed_pipeline.sql` | 3 agent definitions (feed-ingester, content-feed-orchestrator, feed-triage) |

### Sequence

1. Run `026_content_sources_table.sql` on clients DB
2. Merge Go files + registry entries into chassis repo
3. Rebuild and deploy chassis image
4. Run `026b_agent_definitions_feed_pipeline.sql` on clients DB
5. Insert content sources for target site
6. Enable content feed in site settings (optional — falls back to "enabled if sources exist")

### Environment variables for spawned jobs

Spawned K8s jobs get env vars from a hardcoded list in `spawn_actions.go` + the agent definition's `env_vars` JSONB column. The chassis deployment's env vars are NOT automatically inherited by jobs.

`ANTHROPIC_API_KEY` is in the hardcoded list (from `personae-default-secrets`). `GROK_API_KEY` is not.

**Quick fix (no rebuild):** Set the key via agent definition `env_vars`:

```sql
UPDATE agent_definitions 
SET env_vars = '[{"name": "GROK_API_KEY", "value": "<key-value>"}]'::jsonb,
    updated_at = NOW()
WHERE type = 'feed-ingester' AND deleted_at IS NULL;
```

**Proper fix (next rebuild):** Add `GROK_API_KEY` to `personae-default-secrets` and reference it in `spawn_actions.go` alongside `ANTHROPIC_API_KEY`:

```go
{
    Name: "GROK_API_KEY",
    ValueFrom: &corev1.EnvVarSource{
        SecretKeyRef: &corev1.SecretKeySelector{
            LocalObjectReference: corev1.LocalObjectReference{
                Name: "personae-default-secrets",
            },
            Key: "GROK_API_KEY",
        },
    },
},
```

---

## Testing

### Trigger pattern

All tests use the kcat spawn+call pattern through the generic agent:

```bash
SITE_ID="<site-uuid>"
SOURCE_ID="<source-uuid>"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-feed-test-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H orchestration_name=feed-test \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=cli \
    -H from_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses <<JSON
{"headers":{...},"config":{"workflow":{"start_step":"spawn_ingester","steps":{"spawn_ingester":{"action":"spawn_agent","config":{"agent_type":"feed-ingester","role":"<type>-test"},"next_step":"call_ingester"},"call_ingester":{"action":"call_agent","config":{"target_role":"<type>-test","input_mapping":{"site_id":"input_data.site_id","source_id":"input_data.source_id","source_type":"input_data.source_type","source_config":"input_data.source_config"}},"next_step":"complete"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":300},"input_data":{"site_id":"$SITE_ID","source_id":"$SOURCE_ID","source_type":"<type>","source_name":"<name>","source_config":{...}}}
JSON
```

### Checking results

```sql
-- Items written
SELECT source_title, status, created_at,
       (SELECT name FROM content_sources WHERE id = cfi.source_id) as source_name
FROM content_feed_items cfi
WHERE site_id = '<site-uuid>'
ORDER BY created_at DESC LIMIT 15;

-- Orchestration state (pass correlation_id from trigger)
SELECT status, current_step,
       collected_data->'write_result'->>'written' as written,
       collected_data->'write_result'->>'skipped' as skipped,
       collected_data->'route_by_type' as route
FROM orchestration_states 
WHERE correlation_id = '<correlation-id>'
  AND collected_data->'write_result' IS NOT NULL;

-- All recent ingestion results
SELECT 
  collected_data->'input_data'->>'source_type' as type,
  status,
  collected_data->'write_result'->>'written' as written,
  collected_data->'write_result'->>'skipped' as skipped,
  created_at
FROM orchestration_states 
WHERE collected_data->'input_data'->>'site_id' = '<site-uuid>'
  AND collected_data->'write_result' IS NOT NULL
ORDER BY created_at DESC LIMIT 10;

-- Source scheduling state
SELECT name, source_type, last_fetched_at, next_fetch_at, error_count, last_error
FROM content_sources WHERE site_id = '<site-uuid>';
```

### Debugging

```bash
# Feed-ingester pod logs (spawned as K8s jobs)
kubectl -n ai-persona-system logs -l agent-type=feed-ingester --tail=100

# Generic agent logs (receives trigger, spawns ingester)
kubectl -n ai-persona-system logs deployment/agent-chassis --tail=200 | grep -i "feed\|rss\|dispatch"

# Check if consumer group is stuck (after chassis restart/rebalance)
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```

### Consumer group offset reset

If chassis pods restart and messages aren't being consumed, the consumer group offsets may be stale:

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

---

## What's Not Built Yet

In pipeline order, with the 006/007 step references:

| Component | 006 ref | Type | Blocked on |
|-----------|---------|------|------------|
| `latest-news` component template + input_schema | — | Component + SQL | Nothing — next to build |
| `configure_feed_sources` action | — | Go action | Classifier prompt addition |
| News section render support (`query.*` source resolution) | — | Go code in plan_sections | latest-news component |
| Classifier prompt: `news_feed` in classification spec | — | Prompt change | Nothing |
| Planner prompt: include `latest-news` when spec says to | — | Prompt change | Classifier addition |
| `missing_news_sources` discovery check | — | Go check in completeness-discovery | Classifier prompt |
| `stale_news_section` discovery check | — | Go check | latest-news component on pages |
| `missing_news_section` discovery check | — | Go check | latest-news component existing |
| Scheduled rerender of news sections | — | CronJob or maintenance task | News section on pages |
| `article-rewriter` agent | Step 3.5 | SQL + workflow | feed-triage working |
| `feed-publisher` agent | Step 3.6 | SQL + workflow (uses existing actions) | article-rewriter working |
| `feed-orchestrator` scheduled task | Step 3.7 | K8s CronJob manifest | Full pipeline working |
| `feed-lifecycle` agent | Step 3.8 | Go action + SQL | Not blocking anything |
| Scrape link quality filtering | — | Improve `isNavigationLink` | Category-level links currently pass through |
| `GROK_API_KEY` in `spawn_actions.go` | — | Go code change | Currently using agent_definition env_vars workaround |
| Triage evolves to rewriter | — | Richer prompt + write rewritten summary | Triage working first |

### Dependency chain for full pipeline

```
latest-news component template
  → news section render support (query.* source resolution)
    → classifier prompt (news_feed in classification spec)
      → planner prompt (places latest-news on homepage)
        → discovery checks (missing_news_sources, stale_news_section)
          → scheduled rerender of news sections
            → Full automated news cycle

Feed triage (already built, parallel to above):
  → test with existing ingested items
    → triage evolves to filter + rewriter
```

Each step produces a testable deliverable. Don't build the next until the previous works.

---

## Environment Variables

| Variable | Required for | Where set | Note |
|----------|-------------|-----------|------|
| `GROK_API_KEY` or `XAI_API_KEY` | `api_news` sources with `provider: "xai"` | `personae-default-secrets` or agent def `env_vars` | Go checks both names. Must be on spawned jobs, not just chassis deployment |
| `OPENAI_API_KEY` | `api_news` sources with `provider: "openai"` | Would need same treatment | Not yet tested |
| `PERPLEXITY_API_KEY` | `api_news` sources with `provider: "perplexity"` | Would need same treatment | Not yet tested |
| `ANTHROPIC_API_KEY` | `feed-triage` LLM scoring | `personae-default-secrets` (already in spawn_actions.go) | Already works on spawned jobs |

---

## Resolved Decisions

26. **One action per concern, not one action per source type.** The 006 plan put RSS parsing and web scraping in the same `ingest_feed_source` action. We split into separate fetch, normalize, and write actions. The workflow's `conditional_route` handles source-type branching — actions handle their single responsibility.

27. **LLM news as a source type.** The original architecture (002d, 006, 007) didn't include LLM-based news fetching. Grok/xAI has strong real-time news capability, so `api_news` was added as a source type. Uses the OpenAI-compatible chat completions API format, making it provider-agnostic.

28. **Fire-and-forget dispatch, not sequential loop.** The orchestrator doesn't loop over sources internally. It produces independent messages per source, each creating its own orchestration. A slow RSS feed doesn't block a fast one. Follows the same pattern as the build dispatch loop.

29. **Config JSONB, not a `url` column.** Different source types need structurally different configuration. RSS needs `feed_url` + `max_items`. News search needs `query` + `num_results`. LLM news needs `provider` + `model` + `prompt_template`. A single `url` column (as in 006) doesn't accommodate this.

30. **Optimistic timestamp update.** `dispatch_feed_sources` updates `next_fetch_at` immediately after dispatch, before the ingester runs. This prevents the next orchestrator run from re-dispatching the same source. The ingester's `update_source_timestamps` then sets the real next time (or applies backoff on error).

31. **Scrape config uses `url` (string) not `urls` (array).** Multiple URLs become separate `content_sources` rows, which is better for error tracking and scheduling — one source, one URL, one set of timestamps. The `fetch_scrape` action handles both formats but flat is recommended.

32. **Scrape normalization extracts links from listing pages.** Scraping a news index page returns one HTML document, not individual articles. The `normalize_to_feed_items` action navigates into the firecrawl response wrapper (`response.data`), extracts the `links` array, filters out navigation/social links, and treats each article link as a feed item. Falls back to treating the whole page as a single item if no usable links are found.

33. **Wrapper actions for async source types.** `fetch_scrape` and `fetch_news_search` wrap the existing `WebscrapeAction` and `WebSearchAction` respectively. They read `source_config` in Go and set up the config the underlying action expects. This follows the dev guide rule #10 (complexity in Go, not workflow config) and avoids path resolution issues with array indexing. Same pattern as `FirecrawlScrapeAction` which already wraps `WebscrapeAction`.

34. **Spawned job env vars are not inherited from the chassis deployment.** `spawn_actions.go` has a hardcoded list of env vars for K8s jobs. New API keys (like `GROK_API_KEY`) must be added to either the agent definition's `env_vars` column (quick, key in DB) or to `spawn_actions.go` + `personae-default-secrets` (proper, key in K8s secret). This caught us during Grok testing — the key was set on the chassis but the spawned ingester job didn't have it.

35. **`items` is optional in `WriteFeedItemsInputSpec`.** A scrape of a page with no extractable links produces `normalize_scrape: {items: null, item_count: 0}`. If `items` were required, the whole workflow would fail. Making it optional lets the action return `{written: 0, skipped: 0}` gracefully.

36. **Two-layer conformance checking.** Pre-display: feed-triage scores items against the full site spec (identity, classification, content_direction, legal_rules) before they can appear on pages. Post-display: the content-quality-auditor in the improvement loop periodically checks that displayed news items are appropriate. The pre-display gate catches most issues; the post-display audit catches drift and edge cases.

37. **Triage reads all site spec aspects, not just vertical keywords.** The original stub used `input_data.vertical` (a single string). The runnable version loads all spec aspects via `read_site_spec`. This means the LLM can check against tone, legal constraints, forbidden phrases, and audience targeting — not just topic relevance. This is what makes the values/legal flag meaningful.

38. **Triage threshold is configurable per site.** Default 50, stored in `sites.settings.maintenance_profile.content_feed.relevance_threshold`. Energy news sites with many relevant items can set 70 (strict). General business sites might set 40 (permissive). The `apply_feed_scores` action reads this from collected_data if available.

39. **Triage evolves to rewriter (future).** Currently triage is a filter — items pass or fail. Future evolution: for items that pass, triage also rewrites the summary to frame the news for the site's specific audience. The rewritten summary would be stored in a new `display_summary` column while preserving the original `source_summary`. No architecture change needed — just a richer prompt and one more column.


