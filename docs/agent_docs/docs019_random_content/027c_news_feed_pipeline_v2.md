# 026 — News Feed Pipeline

Implementation guide for the content feed pipeline. First vertical: boxing/events.

Implements Phase 3 from `006_entity_feeds_livedata_implementation.md`. See `007_entity_feeds_livedata_architecture.md` for cross-vertical design rationale.

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
| Single `ingest_feed_source` action | Separate `fetch_rss`, `fetch_llm_news`, `normalize_to_feed_items`, `write_feed_items` actions | Keeps each action focused. RSS parsing, LLM HTTP calls, and DB writes are different concerns. Easier to test and debug individually. |
| RSS + scrape only | RSS + scrape + news_search + api_news (Grok/xAI) | Added LLM news providers (xAI/Grok has strong real-time news) and reuse of existing `web_search` action with `search_type: "news"` |
| `enabled` column | `is_active` column | Matches pattern used elsewhere in the codebase (`content_components.is_active`, etc.) |
| `url` column on content_sources | `config` JSONB (contains `feed_url`, `urls`, `queries`, etc.) | Different source types need different config shapes. A single `url` doesn't work for news_search (needs query arrays) or api_news (needs model + prompt) |
| `check_interval` column name | `fetch_interval` column name | Avoids confusion with health checks |
| All sources loop inside one agent | Orchestrator spawns separate feed-ingester per source | Cleaner logs, failure isolation, follows the "every agent is an orchestrator" principle |
| `feed-ingester` loops internally | `conditional_route` in workflow branches by source_type | Keeps workflow simple, each branch follows a straight line |

The table schema also adds `entity_type` (from 006 Phase 2 spec) so the same `content_sources` table serves both entity data sources and news feed sources, as designed.

---

## What Already Existed

| Asset | Status | Used by this pipeline |
|-------|--------|----------------------|
| `content_feed_items` table | Deployed | Yes — receives all ingested items |
| `web_search` action | Deployed | Yes — handles `news_search` source type via `search_type: "news"` |
| `firecrawl_scrape` action | Deployed | Yes — handles `scrape` source type |
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
{
    "feed_url": "https://www.boxingscene.com/feed",
    "max_items": 15
}
```

**news_search** (uses existing web search adapter):
```json
{
    "queries": ["boxing news today", "boxing fight results"],
    "num_results": 10,
    "provider": "firecrawl"
}
```

**api_news** (xAI/Grok, OpenAI, Perplexity — OpenAI-compatible chat completions):
```json
{
    "provider": "xai",
    "model": "grok-3",
    "prompt_template": "What are the latest boxing news stories from the last {{.hours}} hours? ...",
    "hours_lookback": 12,
    "max_items": 10
}
```

**scrape** (uses existing firecrawl adapter):
```json
{
    "url": "https://www.boxrec.com/en/news",
    "scrape_config": {"only_main_content": true},
    "max_items": 10
}
```

**api_data** (future — structured APIs like BoE rates, ticket prices):
```json
{
    "endpoint": "https://api.bankofengland.co.uk/...",
    "method": "GET",
    "headers": {},
    "response_path": "result.data",
    "data_type": "mortgage_rate"
}
```

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
- Provider resolution: `xai` → `api.x.ai`, `openai` → `api.openai.com`, `perplexity` → `api.perplexity.ai`
- API key from env: `XAI_API_KEY`, `OPENAI_API_KEY`, `PERPLEXITY_API_KEY`
- System prompt instructs JSON-only response
- Strips markdown code fences before parsing
- 60s HTTP timeout
- Generates stable `external_id` from title hash when no URL available

**write_feed_items details:**
- Dedup by `source_url` per site (existence check, not unique constraint)
- Handles nullable `source_id` (some items come from search, not a configured source)
- Parses `published_at` as RFC3339
- Items with no title AND no URL are skipped

**update_source_timestamps details:**
- Success: resets `error_count` to 0, `next_fetch_at = NOW() + fetch_interval`
- Failure: `error_count += 1`, backoff `next_fetch_at = NOW() + (fetch_interval * LEAST(error_count + 1, 4))`
- Caps backoff at 4x the normal interval

### feed_normalize_action.go — 1 action

| Action | Category | Purpose |
|--------|----------|---------|
| `normalize_to_feed_items` | feed | Transforms `web_search` or `firecrawl_scrape` output into the `{title, url, summary, external_id, published_at}` format expected by `write_feed_items` |

Two modes via `source_format` config:
- `"search"`: reads from `search_results.results` array, maps `snippet` → `summary`
- `"scrape"`: reads from `scrape_results`, extracts content (markdown preferred, then clean, then HTML-stripped), truncates to 500 chars for summary

### dispatch_feed_sources_action.go — 1 action

| Action | Category | Purpose |
|--------|----------|---------|
| `dispatch_feed_sources` | feed | The orchestrator's core: queries due sources, produces spawn+call messages to generic agent per source |

This is the "fire-and-forget" dispatch pattern used by the build dispatch loop. For each due source:

1. Queries `content_sources WHERE next_fetch_at <= NOW() AND is_active = true`
2. Checks `sites.settings.maintenance_profile.content_feed.enabled` (falls back to "enabled if sources exist")
3. For each source: builds a spawn+call inline workflow message targeting `feed-ingester`
4. Produces to `system.agent.generic.requests`
5. Optimistically updates `next_fetch_at` to prevent re-dispatch before completion

Each message creates an independent orchestration — separate logs, separate failure handling. The generic agent spawns the feed-ingester K8s job, calls it with `{site_id, source_id, source_type, source_config}`.

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

### feed-triage (stub)

| Field | Value |
|-------|-------|
| type | `feed-triage` |
| agent_category | `analyst` |
| status | `experimental` |
| input_contract | `{"required": ["site_id"], "optional": ["vertical"]}` |
| ai_service | `claude-sonnet-4-6`, `api_key_env_var: "ANTHROPIC_API_KEY"` |
| processing_mode | `task` |
| timeout | 300s |

Workflow: `load_items → check_has_items → score_relevance → apply_scores → complete`

Not yet runnable — the `apply_feed_scores` action needs to be built. The definition is here so the workflow structure is documented and ready.

---

## Registry Entries

**File:** `feed_registry_entries.go` — add to `GlobalActionRegistry` and `LocalActions` in `registry.go`.

```go
// FEED section in GlobalActionRegistry
"fetch_rss":                 { Handler: FetchRSSAction,              Category: "feed", IsLocal: true }
"fetch_llm_news":            { Handler: FetchLLMNewsAction,          Category: "feed", IsLocal: true }
"write_feed_items":          { Handler: WriteFeedItemsAction,        Category: "feed", IsLocal: true }
"load_due_sources":          { Handler: LoadDueSourcesAction,        Category: "feed", IsLocal: true }
"update_source_timestamps":  { Handler: UpdateSourceTimestampsAction, Category: "feed", IsLocal: true }
"normalize_to_feed_items":   { Handler: NormalizeToFeedItemsAction,  Category: "feed", IsLocal: true }
"dispatch_feed_sources":     { Handler: DispatchFeedSourcesAction,   Category: "feed", IsLocal: true }
```

All 7 also go in `LocalActions` map.

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
                          route_by_type → fetch → write_feed_items → update_source_timestamps → complete

Results land in content_feed_items with status = 'ingested'
```

Downstream (not yet built): `feed-triage` → `article-rewriter` → `feed-publisher` → `feed-lifecycle`

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

## Deployment Steps

### Prerequisites

- Agent chassis image with the new Go code compiled in
- `XAI_API_KEY` env var set on chassis deployment (only needed for api_news sources — RSS works without it)

### Step 1 — Create content_sources table

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db < 026_content_sources_table.sql
```

Verify:
```sql
\d content_sources
-- Should show: id, site_id, source_type, name, entity_type, config, fetch_interval, ...
-- Plus FK: content_feed_items_source_id_fkey
```

### Step 2 — Merge Go code into chassis

Files to add to `platform/orchestration/actions/`:
- `feed_actions.go`
- `feed_normalize_action.go`
- `dispatch_feed_sources_action.go`

Merge into `registry.go`:
- 7 entries in `GlobalActionRegistry` under new `FEED` section
- 7 entries in `LocalActions` map

### Step 3 — Rebuild and deploy chassis image

```bash
# Build
docker build -t docker.io/aqls/agent-chassis:v1.0.XXX .

# Push
docker push docker.io/aqls/agent-chassis:v1.0.XXX

# Update deployment
kubectl -n ai-persona-system set image deployment/agent-chassis \
  agent-chassis=docker.io/aqls/agent-chassis:v1.0.XXX
```

### Step 4 — Create agent definitions

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db < 026b_agent_definitions_feed_pipeline.sql
```

Verify:
```sql
SELECT type, agent_category, status FROM agent_definitions
WHERE type IN ('feed-ingester', 'content-feed-orchestrator', 'feed-triage')
  AND deleted_at IS NULL;
-- Should show 3 rows
```

### Step 5 — Seed sources for a boxing site

```sql
-- Find your boxing site
SELECT id, domain FROM sites WHERE domain ILIKE '%box%' OR domain ILIKE '%fight%';

-- Seed sources
SELECT seed_boxing_sources('<site-uuid>');

-- Verify
SELECT name, source_type, fetch_interval, next_fetch_at
FROM content_sources WHERE site_id = '<site-uuid>';
-- Should show 7 rows, all with next_fetch_at = now()
```

### Step 6 — Enable content feed in site settings

```sql
UPDATE sites
SET settings = jsonb_set(
    COALESCE(settings, '{}'::jsonb),
    '{maintenance_profile,content_feed}',
    '{"enabled": true, "every": "6h", "max_articles_per_cycle": 5}'::jsonb,
    true
)
WHERE id = '<site-uuid>';
```

---

## Testing the RSS Path

The RSS path is the simplest to test — fully sync, no API keys, no Kafka async.

### Quick test: fetch_rss action directly

This tests the Go action in isolation. From a running chassis pod:

```sql
-- Insert a single test source
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval, next_fetch_at)
VALUES ('<site-uuid>', 'rss', 'Test BBC Boxing', '{"feed_url": "https://feeds.bbci.co.uk/sport/boxing/rss.xml", "max_items": 5}'::jsonb, '1 hour', now());
```

Then trigger the full pipeline via CLI message to the generic agent:

```bash
SITE_ID="<site-uuid>"
SOURCE_ID="<source-uuid-from-above>"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n ai-persona-system exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.requests \
  --property "parse.headers=true" \
  --property "headers.delimiter=|" << EOF
correlation_id:${CORRELATION_ID}|orchestration_id:${ORCHESTRATION_ID}|message_type:request|action:orchestrate|client_id:demo_client	{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_type":"request","action":"orchestrate","client_id":"demo_client","message_id":"${MESSAGE_ID}","request_id":"${REQUEST_ID}","timestamp":"${TIMESTAMP}","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"}},"config":{"workflow":{"start_step":"spawn_ingester","steps":{"spawn_ingester":{"action":"spawn_agent","config":{"agent_type":"feed-ingester","role":"rss-test"},"next_step":"call_ingester"},"call_ingester":{"action":"call_agent","config":{"agent_type":"feed-ingester","target_role":"rss-test","input_mapping":{"site_id":"input_data.site_id","source_id":"input_data.source_id","source_type":"input_data.source_type","source_config":"input_data.source_config"}},"next_step":"complete"},"complete":{"action":"complete_workflow"}}},"processing_mode":"orchestrator","timeout_seconds":300},"input_data":{"site_id":"${SITE_ID}","source_id":"${SOURCE_ID}","source_type":"rss","source_name":"Test BBC Boxing","source_config":{"feed_url":"https://feeds.bbci.co.uk/sport/boxing/rss.xml","max_items":5}}}
EOF
```

### What to check

```sql
-- Items should appear within ~30 seconds
SELECT source_title, source_url, status, created_at
FROM content_feed_items
WHERE site_id = '<site-uuid>'
ORDER BY created_at DESC
LIMIT 10;

-- Source should be updated
SELECT name, last_fetched_at, next_fetch_at, error_count
FROM content_sources
WHERE id = '<source-uuid>';
```

### Debugging

```bash
# Check feed-ingester pod logs
kubectl -n ai-persona-system logs -l app=agent-feed-ingester --tail=100

# Check generic agent logs for dispatch
kubectl -n ai-persona-system logs deployment/agent-chassis --tail=200 | grep -i "feed\|rss\|dispatch"

# Check orchestration state
SELECT orchestration_id, status, current_step, 
       collected_data->'write_result' as write_result,
       collected_data->'fetched_items'->'item_count' as items_fetched
FROM orchestration_states
WHERE correlation_id = '<correlation-id>'
ORDER BY created_at DESC;
```

---

## What's Not Built Yet

In pipeline order, with the 006/007 step references:

| Component | 006 ref | Type | Blocked on |
|-----------|---------|------|------------|
| `apply_feed_scores` action | Step 3.4 | Go action | Nothing — next to build |
| `feed-triage` agent (runnable) | Step 3.4 | Agent def already exists | `apply_feed_scores` action |
| `article-rewriter` agent | Step 3.5 | SQL + workflow | feed-triage working |
| `feed-publisher` agent | Step 3.6 | SQL + workflow (uses existing actions) | article-rewriter working |
| `feed-orchestrator` scheduled task | Step 3.7 | K8s CronJob manifest | Full pipeline working |
| `feed-lifecycle` agent | Step 3.8 | Go action + SQL | Not blocking anything |
| Multi-query support for news_search | — | Patch to feed-ingester workflow | Currently picks only `queries.0` |
| Grok/xAI real-time news grounding | — | May need `search: true` parameter | Grok API feature availability |

### Dependency chain for full pipeline

```
apply_feed_scores action
  → feed-triage agent runnable
    → article-rewriter agent
      → feed-publisher agent
        → K8s CronJob for content-feed-orchestrator
          → Full automated news cycle
```

Each step produces a testable deliverable. Don't build the next until the previous works.

---

## Environment Variables

| Variable | Required for | Default |
|----------|-------------|---------|
| `XAI_API_KEY` | `api_news` sources with `provider: "xai"` | None (api_news sources fail without it) |
| `OPENAI_API_KEY` | `api_news` sources with `provider: "openai"` | None |
| `PERPLEXITY_API_KEY` | `api_news` sources with `provider: "perplexity"` | None |
| `ANTHROPIC_API_KEY` | `feed-triage` LLM scoring | Already set on chassis |

RSS and scrape sources need no additional env vars — they use existing Firecrawl adapter infrastructure.

---

## Resolved Decisions

26. **One action per concern, not one action per source type.** The 006 plan put RSS parsing and web scraping in the same `ingest_feed_source` action. We split into `fetch_rss`, `fetch_llm_news`, `normalize_to_feed_items`, `write_feed_items`. The workflow's `conditional_route` handles source-type branching — actions handle their single responsibility.

27. **LLM news as a source type.** The original architecture (002d, 006, 007) didn't include LLM-based news fetching. Grok/xAI has strong real-time news capability, so `api_news` was added as a source type. Uses the OpenAI-compatible chat completions API format, making it provider-agnostic.

28. **Fire-and-forget dispatch, not sequential loop.** The orchestrator doesn't loop over sources internally. It produces independent messages per source, each creating its own orchestration. A slow RSS feed doesn't block a fast one. Follows the same pattern as the build dispatch loop.

29. **Config JSONB, not a `url` column.** Different source types need structurally different configuration. RSS needs `feed_url` + `max_items`. News search needs `queries` array + `num_results`. LLM news needs `provider` + `model` + `prompt_template`. A single `url` column (as in 006) doesn't accommodate this.

30. **Optimistic timestamp update.** `dispatch_feed_sources` updates `next_fetch_at` immediately after dispatch, before the ingester runs. This prevents the next orchestrator run from re-dispatching the same source. The ingester's `update_source_timestamps` then sets the real next time (or applies backoff on error).

31. **Scrape config uses `url` (string) not `urls` (array).** The field path resolver doesn't support array index notation (`.0`). Multiple URLs become separate `content_sources` rows, which is better for error tracking and scheduling anyway — one source, one URL, one set of timestamps.

32. **Scrape normalization extracts links from listing pages.** Scraping a news index page returns one HTML document, not individual articles. The `normalize_to_feed_items` action navigates into the firecrawl response wrapper (`response.data`), extracts the `links` array, filters out navigation/social links, and treats each article link as a feed item. Falls back to treating the whole page as a single item if no usable links are found.
