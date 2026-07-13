# 014 — Entity Data, Feed Pipeline & Live Data: Implementation Guide

Step-by-step build guide. Each step produces a testable deliverable. Follow in order.

See `014_entity_feed_livedata_architecture.md` for design rationale and cross-vertical application.

---

## What Already Exists

| Asset | Status | Notes |
|-------|--------|-------|
| `scrape_web` action | Deployed | max_pages, follow_links, extract_mode |
| `web_search` action | Deployed | num_results, query template |
| `content_feed_items` table | Deployed | Columns: source_id, external_id, source_url, source_title, source_content, relevance_score, status, entity_ids |
| `entity_state_log` table | Deployed | Generic entity state tracking |
| `append_entity_state` action | Deployed | |
| `read_latest_entity_state` action | Deployed | |
| `site_work_items.entity_id` column | Deployed | FK ready for entity-linked work items |
| `execute_llm_prompt` action | Deployed | JSON output parsing |
| `create_work_item` action | Deployed | |
| `page-build-handler` agent | Deployed | Wraps content writer with persistence |

**Missing:** `content_sources` table, `site_entities` table, entity-data-agent, feed pipeline agents, data proxy.

---

## Phase 2: Entity Data

### Step 2.1 — Create tables

**File:** `migration_entity_tables.sql`

```sql
-- Content sources: configured data feeds per site
CREATE TABLE IF NOT EXISTS content_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    source_type     TEXT NOT NULL DEFAULT 'scrape',  -- scrape, rss, api
    url             TEXT NOT NULL,
    config          JSONB DEFAULT '{}'::jsonb,
    -- config examples:
    --   scrape: {"follow_links": ["players/*"], "extract_mode": "text", "max_pages": 50}
    --   rss:    {"max_items": 20}
    --   api:    {"api_key_env": "SPORTRADAR_KEY", "endpoint": "/darts/rankings"}
    entity_type     TEXT,            -- which entity type this source provides
    enabled         BOOLEAN DEFAULT true,
    check_interval  INTERVAL DEFAULT '24 hours',
    last_checked_at TIMESTAMPTZ,
    error_count     INTEGER DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_cs_site ON content_sources(site_id);
CREATE INDEX idx_cs_enabled ON content_sources(site_id, enabled) WHERE enabled = true;

-- Site entities: structured data records that generate pages
CREATE TABLE IF NOT EXISTS site_entities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    entity_type     TEXT NOT NULL,        -- player, tournament, venue, practice, supplier
    slug            TEXT NOT NULL,        -- url-safe identifier: luke-humphries
    name            TEXT NOT NULL,        -- display name: Luke Humphries
    data            JSONB NOT NULL DEFAULT '{}',  -- all structured fields
    status          TEXT DEFAULT 'active',  -- active, archived, draft
    source_id       UUID REFERENCES content_sources(id),
    source_url      TEXT,                 -- where we got this data
    source_hash     TEXT,                 -- hash of source data for change detection
    page_id         UUID REFERENCES pages(id),  -- the generated page for this entity
    last_synced_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(site_id, entity_type, slug)
);

CREATE INDEX idx_se_site_type ON site_entities(site_id, entity_type);
CREATE INDEX idx_se_status ON site_entities(site_id, status) WHERE status = 'active';
CREATE INDEX idx_se_page ON site_entities(page_id) WHERE page_id IS NOT NULL;

-- Entity relationships: connections between entities
CREATE TABLE IF NOT EXISTS site_entity_relationships (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id             UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    source_entity_id    UUID NOT NULL REFERENCES site_entities(id) ON DELETE CASCADE,
    target_entity_id    UUID NOT NULL REFERENCES site_entities(id) ON DELETE CASCADE,
    relationship_type   TEXT NOT NULL,    -- won_tournament, plays_at_venue, defends_title
    data                JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ DEFAULT now(),
    UNIQUE(source_entity_id, target_entity_id, relationship_type)
);
```

**Test:** Run migration. `\d site_entities`, `\d content_sources`.

---

### Step 2.2 — Entity config in site_specs

The classifier or manual submission stores entity types and sources. This tells the entity-data-agent what to scrape and how to structure it.

```sql
-- Example for dartsonline.com (run after site exists)
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
VALUES (
    '<dartsonline_site_id>', 'entity_config',
    '{
        "entity_types": {
            "player": {
                "display_name": "Player",
                "directory_page": "players",
                "page_url_pattern": "/players/{slug}.html",
                "sort_default": "data.ranking",
                "fields_for_listing": ["name", "nickname", "nationality", "ranking"],
                "page_sections": ["hero", "article-body", "call-to-action"]
            },
            "tournament": {
                "display_name": "Tournament",
                "directory_page": "tournaments",
                "page_url_pattern": "/tournaments/{slug}.html",
                "sort_default": "data.start_date",
                "fields_for_listing": ["name", "dates", "venue_name", "prize_money"],
                "page_sections": ["hero", "article-body", "call-to-action"]
            }
        }
    }'::jsonb,
    'manual', 'human'
);
```

Also insert the scrape sources:

```sql
INSERT INTO content_sources (site_id, name, source_type, url, config, entity_type, check_interval)
VALUES
    ('<dartsonline_site_id>', 'pdc-rankings', 'scrape',
     'https://www.pdc.tv/order-of-merit/pdc-order-of-merit',
     '{"extract_mode": "text", "max_pages": 1}'::jsonb,
     'player', '7 days'),
    ('<dartsonline_site_id>', 'pdc-profiles', 'scrape',
     'https://www.pdc.tv/players',
     '{"extract_mode": "text", "max_pages": 50, "follow_links": ["players/*"]}'::jsonb,
     'player', '7 days'),
    ('<dartsonline_site_id>', 'pdc-schedule', 'scrape',
     'https://www.pdc.tv/tournaments',
     '{"extract_mode": "text", "max_pages": 3}'::jsonb,
     'tournament', '30 days');
```

**Test:** `SELECT * FROM content_sources WHERE site_id = '...'` — 3 rows.

---

### Step 2.3 — Go actions

All go in `platform/orchestration/actions/`. Each needs a registry entry in `registry.go`.

#### 2.3a — `upsert_entities_action.go`

**Input (from collected_data):**
- `site_id` — UUID (path-resolved)
- `entity_type` — string (config literal)
- `entities` — array from LLM parse step (path-resolved)

**Logic:**
```go
for each entity in entities:
    hash = sha256(json.Marshal(entity.Data))
    INSERT INTO site_entities (site_id, entity_type, slug, name, data,
                               source_url, source_hash, last_synced_at)
    VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, now())
    ON CONFLICT (site_id, entity_type, slug) DO UPDATE
    SET data = EXCLUDED.data, source_hash = EXCLUDED.source_hash,
        last_synced_at = now(), updated_at = now()
    WHERE site_entities.source_hash != EXCLUDED.source_hash
```

**Output:** `{ inserted: N, updated: N, unchanged: N }`

**Registry:** `"upsert_entities": { Handler: UpsertEntitiesAction, Category: "site", IsLocal: true }`

**Test:** Insert 3 test entities manually via SQL. Call action. Check `site_entities`.

---

#### 2.3b — `create_entity_pages_action.go`

Creates page records for entities that don't have one yet.

**Input:**
- `site_id` — UUID
- `entity_type` — string
- `entity_config` — from site_specs (page_url_pattern, page_sections, etc)

**Logic:**
```go
rows = SELECT id, slug, name FROM site_entities
       WHERE site_id = $1 AND entity_type = $2
         AND page_id IS NULL AND status = 'active'

for each entity:
    url = strings.Replace(config.PageURLPattern, "{slug}", entity.Slug, 1)
    pageID = INSERT INTO pages (site_id, name, url, title, page_type,
                                sections, build_status, in_header, in_footer)
             VALUES (siteID, slug, url, name+" | "+domain, "entity-page",
                     config.PageSections, "planned", false, false)
             ON CONFLICT (site_id, name) DO UPDATE SET title = EXCLUDED.title
             RETURNING id

    UPDATE site_entities SET page_id = pageID WHERE id = entity.ID
```

**Output:** `{ pages_created: N }`

**Registry:** `"create_entity_pages": { Handler: CreateEntityPagesAction, Category: "site", IsLocal: true }`

---

#### 2.3c — `create_entity_build_items_action.go`

Creates work items for entity pages that need content generated.

**Input:**
- `site_id` — UUID
- `entity_type` — string

**Logic:**
```go
rows = SELECT se.id, se.slug, se.name, se.data, se.page_id, p.build_status
       FROM site_entities se JOIN pages p ON p.id = se.page_id
       WHERE se.site_id = $1 AND se.entity_type = $2
         AND p.build_status IN ('planned', 'needs_rebuild')

for each entity:
    spec = {
        "page_name":    entity.Slug,
        "page_type":    "entity-page",
        "title":        entity.Name,
        "purpose":      "Profile page for " + entity.Name,
        "sections":     ["hero", "article-body", "call-to-action"],
        "entity_type":  entityType,
        "entity_data":  entity.Data   // <-- structured data for content writer
    }

    INSERT INTO site_work_items (
        site_id, source, domain, item_type, summary,
        page_id, entity_id, priority, handler_agent, status,
        created_by, spec, item_key
    ) VALUES (...) ON CONFLICT DO NOTHING
```

The `entity_data` field in the spec is the key — it gives the content writer real facts to use instead of hallucinating.

**Output:** `{ items_created: N }`

**Registry:** `"create_entity_build_items": { Handler: CreateEntityBuildItemsAction, Category: "site", IsLocal: true }`

---

#### 2.3d — `create_directory_page_action.go`

Creates the directory listing page (e.g. `/players.html`).

**Input:**
- `site_id` — UUID
- `entity_type` — string
- `entity_config` — from site_specs

**Logic:**
```go
pageName = config.DirectoryPage  // "players"
entityCount = SELECT COUNT(*) FROM site_entities WHERE site_id = $1 AND entity_type = $2

INSERT INTO pages (site_id, name, url, title, page_type, build_status,
                   in_header, in_footer, sections)
VALUES (siteID, pageName, "/"+pageName+".html",
        config.DisplayName+" Directory | "+domain,
        "entity-directory", "planned", true, true,
        '["hero","entity-directory","call-to-action"]')
ON CONFLICT (site_id, name) DO NOTHING

// Create work item for the directory page
INSERT INTO site_work_items (...) spec includes entity_list summary
```

**Output:** `{ directory_page_created: bool, entity_count: N }`

**Registry:** `"create_directory_page": { Handler: CreateDirectoryPageAction, Category: "site", IsLocal: true }`

---

### Step 2.4 — entity-data-agent definition

**File:** `entity_data_agent.sql`

Workflow steps:

```
ensure_site_record
  → load_entity_config
      action: read_site_spec
      config: { site_id: "site_record.site_id", aspect: "entity_config" }
      output_field: entity_config
  → load_sources
      action: query_database
      config: {
          query: "SELECT id, name, source_type, url, config, entity_type
                  FROM content_sources
                  WHERE site_id = $1 AND entity_type IS NOT NULL AND enabled = true",
          params: ["site_record.site_id"],
          output_format: "array"
      }
      output_field: entity_sources
  → check_has_sources (conditional — if none, complete early)
  → scrape_loop (loop over entity_sources.result)
      for each source:
        → scrape_source
            action: scrape_web
            config: {
                url_field: "current_source.url",
                extract_mode: "text",
                max_pages: "current_source.config.max_pages"
            }
        → parse_entities
            action: execute_llm_prompt
            config: { ai_service: { model: "claude-sonnet-4-6", ... },
                      prompt_template: "... entity extraction prompt ..." }
        → store_entities
            action: upsert_entities
            config: { site_id: "site_record.site_id",
                      entity_type: "current_source.entity_type",
                      entities: "parse_result.result.entities" }
  → entity_types_loop (loop over entity_config keys)
      for each entity_type:
        → create_entity_pages
        → create_directory_page
        → create_entity_build_items
  → complete
```

**Entity extraction prompt** (used in parse_entities step):

```
You are extracting structured {{.entity_type}} data from a web page.

## Expected Fields
{{.entity_schema}}

## Scraped Content
{{.scraped_content}}

Extract every {{.entity_type}} you can find. For each:
- Populate all fields you can determine from the content
- Set unknown fields to null
- Create a URL-safe slug from the name: "Luke Humphries" → "luke-humphries"

Return ONLY valid JSON:
{
  "entities": [
    {
      "slug": "url-safe-name",
      "name": "Display Name",
      "data": { ... fields from schema ... },
      "source_url": "url of page this came from"
    }
  ]
}
```

**Test:** Trigger entity-data-agent on dartsonline.com. Check `site_entities`, `pages`, `site_work_items`.

---

### Step 2.5 — Content writer entity support

The content writer already receives `current_page.spec` via input_mapping. For entity pages, this spec includes `entity_data`. Add to the content writer's LLM prompt:

```sql
-- Add entity data section to page-content-writer prompt
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
    to_jsonb(
        replace(
            default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
            '## Section Requirements',
            '{{if .current_page.spec.entity_data}}' || E'\n' ||
            '## Entity Data (use this — do not invent facts not present here)' || E'\n' ||
            'Type: {{.current_page.spec.entity_type}}' || E'\n' ||
            'Name: {{.current_page.spec.entity_data.name}}' || E'\n' ||
            'Data: {{.current_page.spec.entity_data}}' || E'\n' ||
            '{{end}}' || E'\n' || E'\n' ||
            '## Section Requirements'
        )
    )
),
updated_at = NOW()
WHERE type = 'page-content-writer' AND deleted_at IS NULL;
```

**Test:** Create a test entity page work item with entity_data in spec. Run page-build-handler. Check that the rendered page uses the entity facts.

---

### Step 2.6 — Entity sync scheduled task

Periodic rescrape to detect changes.

```sql
INSERT INTO scheduled_tasks (name, agent_type, enabled, interval_seconds, config)
VALUES ('entity-sync', 'entity-data-agent', true, 86400,  -- daily
        '{"find_sites_with_sources": true}'::jsonb);
```

The entity-data-agent's `upsert_entities` uses `source_hash` to detect changes. Unchanged entities skip page rebuilds. Changed entities get `needs_rebuild` on their page, picked up by the dispatch loop.

---

### Phase 2 checklist

| # | Deliverable | Type | Test |
|---|------------|------|------|
| 2.1 | Tables: `site_entities`, `content_sources`, `site_entity_relationships` | SQL | `\d site_entities` |
| 2.2 | Entity config spec + content sources for dartsonline.com | SQL | Query both tables |
| 2.3a | `upsert_entities_action.go` + registry | Go | Insert test entities |
| 2.3b | `create_entity_pages_action.go` + registry | Go | Check pages table |
| 2.3c | `create_entity_build_items_action.go` + registry | Go | Check work items |
| 2.3d | `create_directory_page_action.go` + registry | Go | Check directory page |
| 2.4 | `entity-data-agent` definition | SQL | Trigger on dartsonline.com |
| 2.5 | Content writer entity data prompt | SQL | Build an entity page |
| 2.6 | `entity-sync` scheduled task | SQL | Wait 24h or trigger manually |

**Phase 2 done when:** dartsonline.com has player profiles and a player directory page deployed.

---

## Phase 3: Feed Pipeline

### Step 3.1 — News sources

```sql
INSERT INTO content_sources (site_id, name, source_type, url, config, entity_type, check_interval)
VALUES
    ('<dartsonline_site_id>', 'bbc-darts', 'rss',
     'https://feeds.bbci.co.uk/sport/darts/rss.xml',
     '{"max_items": 20}'::jsonb,
     NULL, '6 hours'),
    ('<dartsonline_site_id>', 'pdc-news', 'scrape',
     'https://www.pdc.tv/news',
     '{"extract_mode": "text", "max_pages": 2, "follow_links": ["news/*"]}'::jsonb,
     NULL, '12 hours');
```

News sources have `entity_type = NULL` — they're content items, not entity data.

---

### Step 3.2 — `ingest_feed_source_action.go`

Fetches from one source (RSS or scrape), stores raw items in `content_feed_items`.

**Input:**
- `source` — source record from content_sources
- `site_id` — UUID

**Logic:**
```go
if source.SourceType == "rss":
    // Fetch URL, parse RSS XML
    // Extract: title, link, description, pubDate per <item>
elif source.SourceType == "scrape":
    // Use existing scrapeWeb function internally
    // Extract article links + titles from listing page

for each item:
    INSERT INTO content_feed_items (
        site_id, source_id, external_id, source_url,
        source_title, source_summary, source_published_at, status
    ) VALUES (...) 
    ON CONFLICT (source_url) WHERE status NOT IN ('duplicate','expired','rejected')
    DO NOTHING

UPDATE content_sources SET last_checked_at = now(), error_count = 0
WHERE id = source.ID
```

**Output:** `{ items_ingested: N, items_skipped: N }`

**Registry:** `"ingest_feed_source": { Handler: IngestFeedSourceAction, Category: "content", IsLocal: true }`

---

### Step 3.3 — `feed-ingester` agent

```
ensure_site_record
  → load_feed_sources
      query_database: SELECT * FROM content_sources
          WHERE site_id = $1 AND entity_type IS NULL AND enabled = true
          AND (last_checked_at IS NULL OR now() - last_checked_at > check_interval)
  → check_has_sources (conditional)
  → ingest_loop (loop over sources)
      for each source:
        → ingest_feed_source (new action)
  → complete
```

**Test:** Insert BBC darts RSS source. Trigger feed-ingester. Check `content_feed_items` for new rows.

---

### Step 3.4 — `feed-triage` agent

Scores items for relevance via LLM.

```
ensure_site_record
  → read_site_spec (identity, content_direction)
  → load_untriaged
      query_database: SELECT id, source_title, source_summary, source_url, source_published_at
          FROM content_feed_items
          WHERE site_id = $1 AND status = 'ingested'
          ORDER BY source_published_at DESC LIMIT 20
  → check_has_items (conditional)
  → triage_llm
      execute_llm_prompt: score each item 0-100
  → update_scores_action
      for each scored item:
        UPDATE content_feed_items SET relevance_score = score,
            status = CASE WHEN score >= 60 THEN 'relevant' ELSE 'rejected' END
  → complete
```

**New action:** `update_feed_scores_action.go` — takes LLM output (array of {id, score, angle}), updates content_feed_items.

**Triage prompt:**
```
Score these news items 0-100 for relevance to {{.site_record.domain}}.
Target audience: {{.site_specs.specs.identity.target_audience}}

90-100: Major darts news, publish immediately
60-89:  Relevant, worth rewriting
30-59:  Marginal, skip unless slow news day
0-29:   Not relevant

{{range .items}}
ID: {{.id}}
Title: {{.source_title}}
Summary: {{.source_summary}}
{{end}}

Return JSON: { "scores": [{"id": "...", "score": N, "angle": "why relevant"}] }
```

**Test:** Trigger after ingester. Check `content_feed_items` for updated scores.

---

### Step 3.5 — `article-rewriter` agent

Rewrites relevant items in the site's voice, creates blog posts.

```
ensure_site_record
  → read_site_spec (identity, content_direction)
  → load_entities (for cross-linking)
      query_database: SELECT slug, name, entity_type FROM site_entities
          WHERE site_id = $1 AND status = 'active'
  → load_relevant_items
      query_database: SELECT * FROM content_feed_items
          WHERE site_id = $1 AND status = 'relevant'
          ORDER BY relevance_score DESC LIMIT 5
  → rewrite_loop (loop over items)
      for each item:
        → rewrite_llm (execute_llm_prompt — produces original article HTML)
        → create_blog_post (create_blog_posts action — reuse from blog planner)
        → update_feed_item_status
            UPDATE content_feed_items SET status = 'published', published_at = now()
  → create_blog_rerender (create_work_item for blog index)
  → complete
```

**Rewriter prompt:**
```
Rewrite this news item for {{.site_record.domain}}.

## Voice
{{.site_specs.specs.content_direction.voice}}

## Source Article
Title: {{.item.source_title}}
Source: {{.item.source_url}}
Content: {{.item.source_content}}

## Known Entities (link when mentioned)
{{range .entities}}- {{.name}} → /players/{{.slug}}.html
{{end}}

## Rules
1. Write 300-500 words of ORIGINAL content — never copy source text
2. Add analysis and context beyond what the source provides
3. Link entity names to their profile pages
4. Attribute the source: "as reported by BBC Sport"
5. Match the voice: {{.site_specs.specs.content_direction.voice}}

Return JSON:
{
  "title": "Original headline",
  "slug": "url-safe-slug",
  "meta_description": "150 char SEO description",
  "content_html": "<article>...</article>",
  "tags": ["tag1", "tag2"],
  "related_entities": ["entity-slug"]
}
```

**Test:** Run after triage. Check new blog-post pages created and deployed.

---

### Step 3.6 — `feed-publisher` agent

Deploys rendered blog posts to git. Uses existing actions.

```
ensure_site_record
  → load_unpublished_posts
      query_database: SELECT * FROM pages
          WHERE site_id = $1 AND page_type = 'blog-post'
          AND build_status = 'rendered'
  → deploy_loop
      for each page:
        → assemble_page (existing)
        → git_commit (existing)
        → update_page_status (existing)
  → rerender_blog_index
  → complete
```

No new Go code — all existing actions.

---

### Step 3.7 — `feed-orchestrator` agent

Ties the pipeline together. Runs on schedule.

```
ensure_site_record
  → spawn + call feed-ingester
  → spawn + call feed-triage
  → spawn + call article-rewriter
  → spawn + call feed-publisher
  → complete
```

Scheduled task:

```sql
INSERT INTO scheduled_tasks (name, agent_type, enabled, interval_seconds, config)
VALUES ('feed-cycle', 'feed-orchestrator', true, 21600,  -- 6 hours
        '{"find_sites_with_feeds": true}'::jsonb);
```

---

### Step 3.8 — `feed-lifecycle` agent

Ages and archives old posts. Runs weekly.

```
ensure_site_record
  → archive_old: UPDATE pages SET status = 'archived'
      WHERE page_type = 'blog-post' AND deployed_at < now() - interval '90 days'
  → noindex_archived: add meta noindex to archived page heads
  → prune_feed_items: UPDATE content_feed_items SET status = 'expired'
      WHERE published_at < now() - interval '180 days'
  → complete
```

**New action:** `archive_old_pages_action.go` — handles the archive + noindex logic.

---

### Phase 3 checklist

| # | Deliverable | Type | Test |
|---|------------|------|------|
| 3.1 | News sources in content_sources | SQL | Query table |
| 3.2 | `ingest_feed_source_action.go` + registry | Go | Ingest BBC RSS |
| 3.3 | `feed-ingester` agent | SQL | Trigger, check content_feed_items |
| 3.4 | `update_feed_scores_action.go` + registry | Go | — |
| 3.4b | `feed-triage` agent | SQL | Trigger, check scores updated |
| 3.5 | `article-rewriter` agent | SQL | Trigger, check blog posts created |
| 3.6 | `feed-publisher` agent | SQL | Trigger, check pages deployed |
| 3.7 | `feed-orchestrator` + scheduled task | SQL | Trigger full cycle |
| 3.8 | `archive_old_pages_action.go` + `feed-lifecycle` agent | Go + SQL | — |

**Phase 3 done when:** BBC darts RSS items are automatically ingested, triaged, rewritten, and published as blog posts on dartsonline.com.

---

## Phase 4: Live Data Widgets

### Step 4.1 — Data proxy service

Separate Go service deployed in K8s (not an agent).

```
deployment: data-proxy
  image: custom Go service
  ports: 8081
  env: SPORTRADAR_KEY, BETFAIR_APP_KEY, REDIS_URL

endpoints:
  GET /api/scores/:tournament       → cached live scores
  GET /api/odds/:event              → cached betting odds
  GET /api/player/:slug/form        → recent results
  GET /api/schedule                 → upcoming events from site_entities
```

Cache: in-memory or Redis. TTL: scores 30s, odds 5min, player form 1h, schedule 24h.

The `/api/schedule` endpoint doesn't need a paid API — it reads from `site_entities` (tournament data from Phase 2 scraping).

### Step 4.2 — Widget JS library

Deploy as `/assets/js/widgets.js` to widget-enabled sites.

```javascript
document.querySelectorAll('[data-widget]').forEach(async el => {
    const type = el.dataset.widget;
    const slug = el.dataset.slug || '';
    const apiBase = el.dataset.apiBase || '/api';
    try {
        const res = await fetch(`${apiBase}/${type}/${slug}`);
        if (!res.ok) throw new Error(res.status);
        const data = await res.json();
        el.innerHTML = renderers[type](data);
    } catch {
        el.innerHTML = '<p class="widget-unavailable">Data temporarily unavailable</p>';
    }
});
```

### Step 4.3 — Widget components in content_components

```sql
INSERT INTO content_components (name, display_name, function, category, component_level, description)
VALUES
    ('live-scores-widget', 'Live Scores Widget', 'live-scores', 'widget', 'element',
     'Client-side live scores widget. Requires data-proxy service and SportRadar API key.'),
    ('odds-widget', 'Betting Odds Widget', 'odds', 'widget', 'element',
     'Client-side odds comparison widget. Requires data-proxy service and betting API key.'),
    ('schedule-widget', 'Schedule Widget', 'schedule', 'widget', 'element',
     'Upcoming events widget. Reads from site_entities — no paid API needed.');
```

### Step 4.4 — API provider integrations

Build only when traffic justifies:

| Provider | When | Cost |
|----------|------|------|
| Schedule (from site_entities) | Phase 2 complete | Free |
| Betfair Exchange | >10k monthly visits | Free (with account) |
| The Odds API | >20k monthly visits | Free tier: 500 req/month |
| SportRadar Darts | >50k monthly visits | ~$200/month |
| Ticketmaster | >20k monthly visits | Free tier available |

### Phase 4 checklist

| # | Deliverable | Type | Test |
|---|------------|------|------|
| 4.1 | Data proxy Go service + K8s deployment | Go | `/api/schedule` returns entity data |
| 4.2 | Widget JS library | Static JS | Widget renders on test page |
| 4.3 | Widget components in content_components | SQL | Components available to planner |
| 4.4 | Provider integrations (as needed) | Go | Live data renders in widgets |

---

## Build Order Summary

| Sprint | Steps | Outcome | New Go files |
|--------|-------|---------|-------------|
| **Now** | — | dartsonline.com Phase 1: content pages via existing pipeline | None |
| **1** | 2.1, 2.3a-d | Entity tables + 4 Go actions + registry | 4 files |
| **2** | 2.2, 2.4-2.6 | Entity agent + sources + content writer patch. Player pages on dartsonline.com | 0 files (SQL only) |
| **3** | 3.1-3.3 | Feed ingester + triage. BBC darts items scored | 2 files |
| **4** | 3.4-3.7 | Full feed pipeline. Automated blog posts | 1 file |
| **5** | 4.1-4.3 | Data proxy + schedule widget (free, uses entity data) | 1 service |
| **Later** | 4.4 | Paid API integrations when traffic justifies | Per provider |

Sprint 1 is the foundation. Sprint 2 delivers visible value. Everything after is incremental.

