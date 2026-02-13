Good topic. Both are mentioned in the architecture doc but at a fairly high level. Let me look at what's already defined and think through the practical aspects.Let me think through these properly. They're related but have different data shapes and lifecycles.

**News feeds** and **ticket/event data** share a common problem: external data sources that change regularly, need to be synced, and drive page content. But they differ in how the data flows through the system.

Let me walk through each.

---

**News feeds — the practical flow**

A site wants industry-relevant news to keep it fresh and give visitors a reason to return. The sources vary by site type — a finance site might pull from FT, Reuters, FCA announcements. A boxing site might pull from ESPN, BoxRec, promoter announcements. A vet site might pull from Vet Times, BVA, RCVS news.

The practical pipeline:

```
1. Source configuration (per site)
   What feeds, what APIs, what frequency, what keywords/filters

2. Ingestion (algorithmic, scheduled)
   Fetch RSS/API → store raw items in content_feed_items
   Each item: title, body, source, date, source_url

3. Deduplication (algorithmic or minimal LLM)
   Same story from multiple sources → keep best, link others
   Near-duplicate detection (headline similarity)

4. Triage (LLM)
   Is this relevant to THIS site?
   How urgent? (breaking vs background)
   What angle makes it relevant to this site's audience?
   Should it become a standalone article, or just a news brief?

5. Rewriting (LLM)
   Rewrite in site's voice
   Add context relevant to site's audience
   Link to relevant pages on the site (entity cross-links)
   Add required disclaimers if applicable (finance, legal)

6. Publication (algorithmic)
   Create page from rewritten article
   Assign to nav group (news/blog/insights)
   Deploy to git
   Update sitemap

7. Lifecycle (algorithmic, periodic)
   Featured → current → aging → archive → prune
   Lifecycle timing varies by site type:
     - News site: fast (24h featured, 7d current)
     - Brochure site with blog: slow (7d featured, 30d current)
     - Events site: tied to event calendar
```

The data model needs:

```sql
-- Source configuration per site
CREATE TABLE content_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_type TEXT NOT NULL,          -- 'rss', 'api', 'scrape'
    source_url TEXT NOT NULL,
    source_config JSONB DEFAULT '{}',   -- API keys, headers, selectors, filters
    category TEXT,                       -- 'industry_news', 'regulatory', 'competitor'
    poll_interval INTERVAL DEFAULT '1 hour',
    last_polled_at TIMESTAMPTZ,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Raw ingested items
CREATE TABLE content_feed_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_id UUID REFERENCES content_sources(id),
    external_id TEXT,                    -- dedup key from source
    title TEXT NOT NULL,
    body TEXT,
    source_url TEXT,
    published_at TIMESTAMPTZ,
    raw_data JSONB,                      -- full original item
    status TEXT DEFAULT 'ingested',      -- ingested, duplicate, triaged,
                                         -- approved, rewritten, published, archived
    triage_result JSONB,                 -- relevance score, suggested angle, urgency
    rewritten_content JSONB,             -- rewritten article content
    page_id UUID REFERENCES pages(id),  -- links to published page
    lifecycle_status TEXT,               -- featured, current, aging, archive, pruned
    lifecycle_changed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, source_id, external_id)
);
```

---

**Ticket/event data — the practical flow**

This is fundamentally different from news. Events and tickets are structured entities with specific fields, prices, dates, availability. They change frequently (prices go up, tickets sell out, new events announced) and the pages need to reflect current state.

Take a boxing events site. The data looks like:

```
Event: Fury vs Joshua
  Date: 2026-06-15
  Venue: Wembley Stadium, London
  Status: on_sale
  Performers: [Fury, Joshua]
  Ticket tiers:
    - Ringside: £500, available: 200
    - Premium: £250, available: 1400
    - Standard: £100, available: 8000
    - Nosebleed: £40, available: 15000
  Source: ticketmaster_api
  Last synced: 2026-02-13T10:00:00Z
  
  Related news: [fight announcement article, press conference article]
  Related events: [undercard bouts]
```

This is entity data. It fits the `site_entities` table, but the sync pattern is different from news. News items are fire-and-forget (ingest, process, publish). Entity data is living — it updates, and those updates need to flow through to already-published pages.

The practical pipeline:

```
1. Source configuration (per site, per entity type)
   Which APIs for events? (Ticketmaster, SeatGeek, promoter feeds)
   Which APIs for performers? (BoxRec, Wikipedia, social media)
   Which APIs for venues? (Google Places, venue sites)
   Polling frequency: events hourly (prices change), venues weekly

2. Ingestion (algorithmic, scheduled)
   Fetch API → upsert into site_entities
   Each entity: type, key, data, source, last_synced_at
   Change detection: has anything changed since last sync?

3. Change processing
   If new entity → decide: does it get a page?
   If changed entity → what changed?
     - Price change → update entity_data, flag page for re-render
     - Availability change → update entity_data, flag page for re-render
     - Status change (on_sale → sold_out) → update entity_data,
       flag page for re-render, may change nav/featured status
     - Date change → update entity_data, flag page for re-render,
       may affect lifecycle scheduling

4. Page generation (for new entities that warrant pages)
   Create page record
   Map entity fields to component templates
   Generate any editorial content (LLM for descriptions, context)
   Deploy

5. Page update (for changed entities)
   Re-render affected sections from updated entity_data
   No full content rewrite — just template re-render with new data
   Redeploy page

6. Lifecycle (tied to entity status, not age)
   announced → on_sale → selling_fast → sold_out → event_day → past → historical
   Unlike news, this isn't time-based decay — it's state-based
   "Past" events might still have value (results, stats, review)
   "Historical" events are SEO assets (people search for past fights)

7. Relationship management
   Event → venue, event → performers, event → ticket_tiers
   Performer page lists all their events
   Venue page lists all events at that venue
   These cross-links are entity-driven, not editorial
```

The entity data model already exists (`site_entities`, `site_entity_relationships`). What's missing is the source management and sync orchestration.

```sql
-- Entity source configuration per site
CREATE TABLE entity_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    entity_type TEXT NOT NULL,          -- 'event', 'performer', 'venue', 'ticket_tier'
    source_type TEXT NOT NULL,          -- 'api', 'scrape', 'manual', 'feed'
    source_config JSONB DEFAULT '{}',   -- API endpoint, credentials, mapping rules
    poll_interval INTERVAL DEFAULT '1 hour',
    last_polled_at TIMESTAMPTZ,
    field_mapping JSONB,                -- maps source fields to entity_data fields
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sync log for auditing and change detection
CREATE TABLE entity_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_id UUID REFERENCES entity_sources(id),
    entity_id UUID REFERENCES site_entities(id),
    sync_type TEXT NOT NULL,            -- 'created', 'updated', 'unchanged', 'removed'
    changes JSONB,                       -- what fields changed
    synced_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

**Where they overlap — and where they don't**

Both need:
- Per-site source configuration (what to fetch, from where, how often)
- Scheduled polling/fetching
- Change detection
- Pages that reflect current data

But they differ in:

| Aspect | News feeds | Entity data (tickets/events) |
|--------|-----------|------------------------------|
| Data shape | Unstructured text | Structured fields |
| Processing | Dedup → triage → rewrite (LLM-heavy) | Upsert → change detect → re-render (mostly algorithmic) |
| Page creation | Every accepted item → new page | Only significant entities → pages |
| Page updates | Rarely (news doesn't change) | Frequently (prices, availability change) |
| Lifecycle | Time-based decay (featured → aging → archive) | State-based (on_sale → sold_out → past) |
| Cross-links | Article links to entity pages | Entity pages link to each other |
| LLM usage | High (triage + rewriting) | Low (maybe descriptions, mostly templates) |

The source configuration tables could potentially be unified — `content_sources` and `entity_sources` have very similar shapes. But keeping them separate makes responsibilities clearer: the feed pipeline owns `content_sources`, the entity pipeline owns `entity_sources`.

---

**How they connect to each other**

This is where it gets interesting. On an events site, news and entities aren't independent:

```
Event announced (entity: created)
  → triggers news article: "Fury vs Joshua announced for June"
  → entity page created with ticket pre-registration
  
Tickets go on sale (entity: status changed)
  → triggers news article: "Tickets now available for Fury vs Joshua"
  → entity page updated with prices and buy links
  
Fight happens (entity: status → past)
  → triggers news article: "Fury defeats Joshua by KO in round 7"
  → entity page updated with results
  → article links to event entity page
  → event entity page links to article

Price change (entity: data updated)
  → no news article (routine)
  → entity page re-rendered silently
```

So entity state changes can trigger news items. Not always — a price change from £99 to £101 isn't news. But a status change (announced, on_sale, sold_out, results) often is.

This could work through a simple mechanism: when the entity-data-agent processes a sync and detects a significant state change, it writes a `content_feed_item` with `source_type = 'entity_event'` and a structured payload. The feed pipeline's triage agent then decides whether this warrants an article.

```
entity-data-agent detects: event.status changed from 'announced' to 'on_sale'
  → writes content_feed_item:
      source_type: 'entity_event'
      title: "Tickets on sale: Fury vs Joshua"
      body: { event_entity_id, change_type: 'status_change',
              old_status: 'announced', new_status: 'on_sale',
              entity_data: {...} }
  → feed-triage decides: yes, this is newsworthy
  → article-rewriter produces article with entity cross-links
  → feed-publisher creates page
```

Not every entity change triggers this. The entity source config could include rules:

```json
{
  "news_triggers": {
    "status_changes": ["announced", "on_sale", "sold_out", "past"],
    "price_changes_above_percent": 20,
    "availability_threshold": 100
  }
}
```

---

**How they connect to maintenance**

Both feed into the maintenance system naturally:

News lifecycle management is essentially a maintenance concern — the `feed-lifecycle` agent is a discovery agent that finds stale content. It could write to `maintenance_findings`:
- "Article X is 45 days old with low traffic → archive candidate"
- "Article Y references an event that's now past → needs update or redirect"

Entity staleness is also a maintenance concern:
- "Event entity hasn't been synced in 48 hours — source may be down"
- "Event page shows prices from 3 days ago — sync lag"
- "Event is past but page still shows 'buy tickets' CTA"

The content-discovery-agent we designed could include checks for:
- Entity pages where `entity.last_synced_at` is stale relative to the source's `poll_interval`
- Entity pages where the entity status has changed but the page hasn't been re-rendered
- News articles referencing entities whose status has since changed

---

**Where this fits in the phases**

Looking at the current phasing, news feeds and entity data are both listed in Phase 2. I think they should be broken out more carefully:

**Phase 2a — Entity data foundation**
- `site_entities` table already exists
- `entity_sources` table (source config)
- `entity_sync_log` table (change tracking)
- `entity-data-agent` with manual/HITL entity creation
- Entity → page generation (template-based, minimal LLM)
- Entity → page re-render when data changes

This is useful immediately — even without API integrations, sites can have manually-entered entity data that generates pages. A vet site's team members, a law firm's practice areas, a restaurant's menu items.

**Phase 2b — Feed pipeline foundation**
- `content_sources` table
- `content_feed_items` table
- `feed-ingester` (RSS polling, simple)
- `feed-triage` (LLM relevance assessment)
- `article-rewriter` (LLM rewriting in site voice)
- `feed-publisher` (page creation + deploy)
- `feed-lifecycle` (aging + archiving)

**Phase 2c — API integrations**
- Ticketmaster / SeatGeek / event API connectors
- Automated entity syncing from APIs
- Change detection and page re-rendering
- Entity state → news trigger mechanism

**Phase 2d — Cross-linking and intelligence**
- Entity ↔ article cross-linking
- Entity relationship-driven navigation (contextual nav)
- Entity-driven components (upcoming events widget, price comparison table)

Phase 2a could be relatively soon — it builds on existing tables and doesn't need external API work. Phase 2b needs the feed pipeline agents but they're self-contained. Phase 2c needs API connector work which is more engineering-heavy. Phase 2d ties everything together.

Does this breakdown match how you're thinking about it? And are there specific source types (particular APIs, particular feed formats) that are near-term priorities?

