# 014 — Entity Data, Feed Pipeline, and Live Data Implementation

How sites like dartsonline.com progress from static content pages to living resources with entity directories, news feeds, and live data widgets.

Applies to any vertical — darts is the first implementation but the architecture is domain-agnostic. Replace "player" with "vet practice", "product", "venue", "mortgage lender" etc.

---

## The Four Phases

| Phase | What | New agents | New Go code | LLM cost | External cost |
|-------|------|-----------|-------------|----------|--------------|
| 1 | Content pages | None | None | Moderate (content generation) | None |
| 2 | Entity data + directory pages | entity-data-agent | Moderate | Low (page generation from data) | None (scrape) |
| 3 | News/content feed | feed pipeline (5 agents) | Significant | Moderate (rewriting) | None (RSS/scrape) |
| 4 | Live data widgets | data-proxy API | Moderate | None | Paid APIs |

Each phase is independently deployable. A site can stay at Phase 1 indefinitely. Phases 2-4 add value incrementally.

---

## Phase 1: Content Pages (Available Now)

### What gets built

The classifier identifies the vertical and the planner creates content pages. For dartsonline.com:

| Page | Purpose | Content source |
|------|---------|---------------|
| index | Hero + value prop + featured content | LLM with research |
| about | What DartsOnline is, editorial voice | LLM |
| tournaments | Major tournament guide (PDC, BDO, WDF) | LLM with research-agent scraping PDC/BBC |
| how-to-watch | TV schedules, streaming options by country | LLM with research |
| betting-guide | How darts betting works, markets explained | LLM with research |
| beginners-guide | Rules, scoring, equipment, how to play | LLM |
| contact | Contact form | Template |
| blog | Index page (empty until Phase 3 or blog-content-planner runs) | - |

### How it works

Already implemented. The pipeline: `domain-submitter → classifier → planner → page-build-handler → deploy`. The research-agent scrapes relevant sources before the content writer generates each page.

### Monetisation

- Google AdSense display ads
- Betting affiliate links embedded in content (Sky Bet, Betfair, William Hill, Paddy Power)
- Typical sports betting affiliate: £25-50 per new depositing customer
- Betting guide and tournament preview pages have high commercial intent

### What's blocked at this phase

The classifier creates `blocked` work items for:
- `player-directory` → needs entity-data-agent
- `tournament-calendar` → needs entity-data-agent
- `news-feed` → needs feed pipeline
- `live-scores` → needs data API
- `odds-comparison` → needs betting API

These sit in `site_work_items` with `status: blocked` until the relevant agents are deployed.

---

## Phase 2: Entity Data + Directory Pages

### Concept

Entities are structured data records that generate pages. A player entity has name, ranking, nationality, titles. A tournament entity has dates, venue, format, prize money. Each entity gets a page. A directory page lists/filters entities.

### New tables

Already defined in 002c architecture doc:

```sql
-- site_entities: individual data records
-- Columns: id, site_id, entity_type, slug, name, data (JSONB), 
--          status, source, source_url, last_synced_at, created_at

-- site_entity_relationships: connections between entities
-- Columns: id, site_id, source_entity_id, target_entity_id, 
--          relationship_type, data (JSONB)
```

### Entity types for darts

```json
{
    "player": {
        "fields": {
            "name": "string",
            "nickname": "string (e.g. 'The Iceman')",
            "nationality": "string",
            "ranking": "integer (PDC Order of Merit position)",
            "tour_card": "boolean",
            "major_titles": ["list of {title, year}"],
            "career_nine_darters": "integer",
            "highest_average": "decimal",
            "photo_url": "string (if available)",
            "pdc_profile_url": "string"
        },
        "source": "PDC website scrape + Wikipedia",
        "update_frequency": "weekly",
        "estimated_count": "128 (tour card holders) + ~50 notable others"
    },
    "tournament": {
        "fields": {
            "name": "string",
            "organisation": "string (PDC, WDF, BDO)",
            "dates": "{start, end}",
            "venue_name": "string",
            "venue_city": "string",
            "format": "string (e.g. 'Sets, best of 13')",
            "prize_money": "string",
            "defending_champion": "string (player slug reference)",
            "status": "upcoming | in_progress | completed",
            "tv_coverage": "string"
        },
        "source": "PDC website scrape",
        "update_frequency": "monthly (schedule), daily (during events)",
        "estimated_count": "30-40 per year"
    },
    "venue": {
        "fields": {
            "name": "string",
            "city": "string",
            "country": "string",
            "capacity": "integer",
            "typical_events": ["list of tournament names"]
        },
        "source": "Manual + scrape",
        "update_frequency": "rarely",
        "estimated_count": "~20"
    }
}
```

### entity-data-agent

New agent. Workflow:

```
ensure_site_record
  → load_entity_config         (read entity types + sources from site_specs)
  → scrape_sources             (web scrape configured URLs)
  → parse_entities             (LLM: extract structured data from scraped content)
  → upsert_entities            (write to site_entities table)
  → create_entity_pages        (create page records for each entity)
  → create_directory_pages     (create directory/listing pages)
  → create_build_items         (work items for page-build-handler)
  → complete
```

### New Go actions needed

| Action | What it does |
|--------|-------------|
| `load_entity_config` | Reads entity type definitions from `site_specs.aspect = 'entity_config'` |
| `scrape_entity_sources` | Scrapes configured URLs, returns raw HTML/text per source |
| `parse_entities_from_scrape` | LLM call: given raw scraped content + entity schema, extract structured records |
| `upsert_entities` | Insert/update `site_entities` rows, detect changes |
| `create_entity_pages` | For each entity, create a page record if one doesn't exist |
| `create_directory_page` | Create the directory/listing page with filter config |

### Entity page generation

Entity pages are content pages with structured data injected into the render context. The content writer receives the entity data and generates a page around it:

```
page-build-handler receives:
  page_name: "luke-humphries"
  page_type: "entity-page"
  entity_type: "player"
  entity_data: { name: "Luke Humphries", nickname: "Cool Hand Luke", ranking: 1, ... }

Content writer generates: 
  hero with player name + nationality flag
  stats section from entity data
  career highlights from entity data + research
  upcoming events (cross-reference tournament entities)
  betting section (affiliate links)
```

### Directory page

The player directory is a content page with a special component:

```
Component: entity-directory
  Config: entity_type = "player", sort_by = "ranking", filterable = ["nationality", "ranking_tier"]
  Template: renders a filterable grid/table from site_entities
  JS: client-side filtering (no server needed for static sites)
```

This component queries `site_entities` at render time and produces static HTML with embedded JS for filtering. Deployed as a static page — no API needed.

### Entity sync (ongoing)

After initial setup, entity data needs periodic refresh. The entity-data-agent runs in discovery mode:

```
Scheduled task: entity-sync (every 24h or weekly depending on entity type)
  → scrape sources for changes
  → compare against stored entities
  → create work items for changed entities (re-generate their pages)
  → new entities get new pages
  → removed entities get archived
```

This uses the same work item pattern as the improvement loop. Changed entities create `needs_entity_update` items routed to `entity-data-agent`.

### Implementation steps

1. **Create tables** — `site_entities`, `site_entity_relationships` (migration SQL)
2. **`load_entity_config` action** — reads from site_specs
3. **`scrape_entity_sources` action** — wraps the existing `scrape_web` action for multiple URLs
4. **`parse_entities_from_scrape` action** — LLM call with entity schema + scraped content
5. **`upsert_entities` action** — bulk insert/update to site_entities
6. **`create_entity_pages` action** — creates page records from entity list
7. **`entity-directory` component** — HTML template with JS filtering
8. **`entity-page` component set** — player-profile, tournament-detail templates
9. **entity-data-agent definition** — SQL workflow composing the actions
10. **Entity sync scheduled task** — periodic re-scrape

### Darts-specific scrape sources

| Source | URL pattern | What we get |
|--------|------------|-------------|
| PDC rankings | `https://www.pdc.tv/order-of-merit` | Player names, rankings, prize money |
| PDC player profiles | `https://www.pdc.tv/players/[slug]` | Bio, stats, photo |
| PDC schedule | `https://www.pdc.tv/tournaments` | Tournament dates, venues, format |
| Wikipedia darts | `https://en.wikipedia.org/wiki/PDC_Order_of_Merit` | Career stats, titles |
| BBC Sport darts | `https://www.bbc.co.uk/sport/darts` | Results, news links |

Scraping is legal for factual data (rankings, results, dates). We synthesise and add value (analysis, cross-referencing, betting context) rather than reproducing content.

---

## Phase 3: News/Content Feed

### Concept

Automated content pipeline: ingest news from external sources, score for relevance, rewrite in the site's voice with entity cross-links, publish as blog posts. The blog index page automatically shows new posts.

### Feed pipeline agents

```
feed-ingester
  → fetch from configured RSS feeds and scrape sources
  → store raw items in content_feed_items table
  → deduplicate against existing items

feed-triage
  → LLM scores each item for relevance (0-100)
  → filters by threshold (e.g. >60 = publish, 30-60 = maybe, <30 = skip)
  → assigns urgency (breaking = immediate, standard = next cycle)
  → assigns angle (what makes this relevant to OUR audience)

article-rewriter
  → LLM rewrites in site voice (from content_direction spec)
  → adds entity cross-links (mentions "Luke Humphries" → links to player page)
  → adds betting context where relevant
  → adds internal links to related content
  → generates meta title, description, OG tags

feed-publisher
  → creates page record (page_type: blog-post)
  → creates page_components with rendered HTML
  → git commits and deploys
  → updates blog index (rerender)

feed-lifecycle
  → ages articles: featured (24h) → current (7d) → archive (30d) → prune
  → for darts: tournament preview articles stay featured until event ends
  → archived articles get noindex meta tag
```

### New tables

Already defined in 002c:

```sql
-- content_sources: configured feed sources per site
-- Columns: id, site_id, name, source_type (rss|scrape|api), 
--          url, config (JSONB), enabled, check_interval,
--          last_checked_at, created_at

-- content_feed_items: raw ingested items
-- Columns: id, site_id, source_id, external_id, title, 
--          content, url, author, published_at, 
--          relevance_score, status, created_at
```

### Free darts news sources

| Source | Type | URL | Frequency |
|--------|------|-----|-----------|
| BBC Sport Darts | RSS | `https://feeds.bbci.co.uk/sport/darts/rss.xml` | Several per day during events |
| Sky Sports Darts | Scrape | `https://www.skysports.com/darts/news` | Daily |
| PDC News | Scrape | `https://www.pdc.tv/news` | Several per week |
| Reddit r/darts | API | `https://www.reddit.com/r/Darts/.json` | Hourly (discussions, not news) |
| Darts news sites | Scrape | Various | Varies |

### Content rewriting rules

The article-rewriter follows the site's `content_direction` spec:
- Voice: enthusiastic but knowledgeable
- Emphasis: analysis, not just reporting
- Entity linking: every player/tournament mention links to their profile page
- Betting context: where relevant, note odds implications
- Never reproduce original article text — always synthesise and add value
- Attribution: "BBC Sport reports that..." with link to original

### Implementation steps

1. **Create tables** — `content_sources`, `content_feed_items` (migration SQL)
2. **`feed-ingester` agent** — fetches RSS, stores items, deduplicates
3. **`feed-triage` agent** — LLM relevance scoring
4. **`article-rewriter` agent** — LLM rewrite with entity cross-linking
5. **`feed-publisher` agent** — page creation + deploy (uses existing page-build-handler pattern)
6. **`feed-lifecycle` agent** — aging, archiving, pruning
7. **`feed-orchestrator` agent** — coordinates the pipeline per cycle
8. **Scheduled task** — triggers feed-orchestrator every N hours
9. **Configure sources** — insert rows into content_sources for darts RSS/scrape targets

### Rewriter prompt structure

```
You are rewriting a news item for DartsOnline.

## Site Voice
{{.content_direction.voice}}
{{.content_direction.emphasis}}

## Original Article
Title: {{.item.title}}
Source: {{.item.source_name}} ({{.item.url}})
Published: {{.item.published_at}}
Content: {{.item.content}}

## Known Entities (link to these when mentioned)
{{range .entities}}- {{.name}} → /players/{{.slug}}.html
{{end}}

## Rules
1. NEVER reproduce original text — synthesise, analyse, add context
2. Write 300-500 words of original content
3. Reference the source with attribution
4. Link player/tournament names to their profile pages
5. Add betting/odds context if relevant
6. Match the site voice — {{.content_direction.voice}}
7. Include a relevant call-to-action at the end

Return JSON:
{
    "title": "Your original headline",
    "meta_description": "SEO description, 150 chars",
    "content_html": "<p>Article content with <a href=\"/players/...\">entity links</a></p>",
    "tags": ["tournament-name", "player-name"],
    "related_entities": ["entity-slug-1", "entity-slug-2"]
}
```

---

## Phase 4: Live Data Widgets

### Concept

Client-side JavaScript widgets embedded in entity pages that fetch live data from a caching proxy API. The pages are still static HTML — the widgets hydrate at view time.

### Architecture

```
Static page (on CDN/S3)
  └── <div id="live-scores" data-tournament="premier-league-2026"></div>
  └── <script src="/assets/js/live-widget.js"></script>

Widget JS:
  → fetch('/api/scores/premier-league-2026')
  → render scores into the div

API proxy (lightweight service):
  → receives request
  → checks cache (Redis or in-memory, TTL 30s-5min)
  → if stale: fetch from upstream API (SportRadar, Betfair)
  → return cached/fresh data as JSON
```

### Data sources and costs

| Data | Provider | Cost | Update frequency |
|------|----------|------|-----------------|
| Live scores | SportRadar Darts API | ~$200/month | Real-time during events |
| Historical results | SportRadar | Included | After each event |
| Betting odds | Betfair Exchange API | Free (with Betfair account) | Every few minutes |
| Betting odds | The Odds API | Free tier: 500 requests/month | Hourly |
| Ticket availability | Ticketmaster API | Free tier available | Daily |
| TV schedules | Manual / scrape | Free | Weekly |

### Widget types

| Widget | Data source | Where it appears |
|--------|------------|-----------------|
| Live scores | SportRadar | Tournament page, homepage (during events) |
| Current odds | Betfair/Odds API | Tournament page, betting guide |
| Player form | SportRadar historical | Player profile pages |
| Upcoming events | Entity data + Ticketmaster | Homepage, tournament calendar |
| Ticket availability | Ticketmaster | Tournament pages |

### Implementation steps

1. **API proxy service** — lightweight Go or Node service, deployed in K8s
2. **Cache layer** — Redis or in-memory with TTL per data type
3. **Widget JS library** — generic widget loader, data-attribute configured
4. **SportRadar integration** — API client, response mapping
5. **Betfair integration** — API client, odds normalisation
6. **Widget components** — HTML templates that include widget divs + JS
7. **Tool-deployer integration** — widgets are "tools" in the component system

### When to invest

Live data costs money. The decision framework:

| Monthly traffic | Revenue | Invest in |
|----------------|---------|-----------|
| < 5,000 visits | < £200 | Phase 1 only (content + ads) |
| 5,000-20,000 | £200-800 | Phase 2 (entities from free scrape) |
| 20,000-50,000 | £800-2,000 | Phase 3 (news feed from free RSS) |
| > 50,000 | > £2,000 | Phase 4 (paid APIs justify their cost) |

SportRadar at ~$200/month is justified when ad revenue exceeds ~$500/month. Before that, scrape-based data (rankings, results after the fact) provides most of the value.

---

## Cross-Vertical Application

The same four phases apply to any vertical:

| Vertical | Phase 2 entities | Phase 3 feeds | Phase 4 live data |
|----------|-----------------|---------------|-------------------|
| Darts | Players, tournaments, venues | BBC/Sky/PDC news | SportRadar scores, Betfair odds |
| Veterinary | Practices, breeds, procedures | Pet health news, RCVS updates | Appointment availability |
| Gas wholesale | Suppliers, regions, contract types | Energy market news, Ofgem updates | Wholesale gas prices |
| Mortgages | Lenders, products, rates | Financial news, BoE decisions | Live mortgage rates |
| Boxing | Fighters, events, venues | Boxing news, results | Ticket availability, odds |

The agent architecture is the same. The entity schemas, scrape sources, and feed configurations differ per vertical. The classifier's spec identifies which entity types and feed sources are appropriate.

---

## Agent Dependency Chain

```
Phase 1 (exists):
  domain-submitter → classifier → planner → page-build-handler → deploy

Phase 2 (new: entity-data-agent):
  entity-data-agent
    ├── scrape_entity_sources (new action)
    ├── parse_entities_from_scrape (new action, uses LLM)
    ├── upsert_entities (new action)
    ├── create_entity_pages (new action)
    └── page-build-handler (exists)

Phase 3 (new: feed pipeline):
  feed-orchestrator
    ├── feed-ingester (new agent)
    ├── feed-triage (new agent, uses LLM)
    ├── article-rewriter (new agent, uses LLM)
    ├── feed-publisher (new agent, uses page-build-handler)
    └── feed-lifecycle (new agent)

Phase 4 (new: data proxy):
  data-proxy-service (new K8s deployment, not an agent)
    ├── SportRadar client
    ├── Betfair client
    ├── cache layer
    └── widget JS library
  tool-deployer (exists) deploys widget components to pages
```

### Feasibility check

When the classifier produces the spec, it checks `agent_definitions` for:
- `entity-data-agent` → if missing, entity pages are `blocked`
- `feed-ingester` → if missing, news feed is `blocked`
- `data-proxy` → not an agent, check via config flag → if missing, live widgets are `blocked`

The `feasibility-recheck` scheduled task promotes blocked items when agents appear.

---

## What to Build First

For dartsonline.com specifically:

1. **Now:** Run Phase 1. Classifier + planner + content pages. Get the site live with tournaments guide, betting guide, how-to-watch. Start affiliate programme applications.

2. **Next sprint:** Build `entity-data-agent` with PDC scraper. Create player profiles and tournament pages. This is the biggest traffic driver — player name searches are high volume.

3. **When traffic reaches 5k/month:** Add feed pipeline with BBC Sport darts RSS. Automated news brings recurring visitors.

4. **When traffic reaches 20k/month:** Evaluate SportRadar and Betfair APIs. Live scores during events are the "bookmark" feature that turns visitors into daily users.

The entity-data-agent is the highest-impact Phase 2 work because it applies to every vertical, not just darts. Building it for darts players means it works for vet practices, mortgage lenders, gas suppliers with minimal changes.

