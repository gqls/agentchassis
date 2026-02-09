https://claude.ai/chat/fbdaef1b-bb4c-45dd-88e5-34349bfe27bf

# Agent Orchestration Architecture v2

## Overview

This document captures the evolved agent architecture for intelligent multi-page website building. It extends the existing system (intake-orchestrator → classifier → briefing → pageflow-builder pipeline) with specialised agent families for navigation, links, design, and content — each with clear responsibilities, data ownership, and interaction patterns.

The guiding principles are:

- **Agents are separate early** — even if responsibilities are small now, the structure supports growth without refactoring
- **New workflows rather than modifying existing ones** — copy and extend, don't break what works
- **Complexity in Go actions, simplicity in workflows** — workflows remain readable orchestration; business logic lives in action code
- **Algorithmic where possible, LLM where necessary** — not every agent needs model calls
- **Heartbeat-driven maintenance** — periodic sweeps rather than event-driven coupling between agents

---

## Current System (What Exists)

### Active Agents

| Agent | Type | Role |
|-------|------|------|
| `intake-orchestrator` | orchestrator | Entry point — classifies, briefs, spawns builder |
| `site-classifier` | specialist | Classifies site type from domain/objective |
| `briefing-agent` | specialist | Runs questionnaire, collects brief data |
| `site-planner` | specialist | Plans pages, selects components and style |
| `pageflow-builder` | orchestrator | Builds sites: plan → assets → content → deploy |
| `page-content-writer` | specialist | Writes content per page, section by section |
| `content-reviewer` | specialist | HITL or auto-eval content review |
| `research-agent` | specialist | Web research for content backing |
| `image-generator` | specialist | Generates logos and hero images |
| `webdesign-agent` | specialist | Generates design spec and CSS |
| `deployer-agent` | specialist | Git commit and Cloudflare deployment |
| `page-rerender` | specialist | Re-assembles single page from stored components |
| `rerender-pages` | orchestrator | Batch rerender across all site pages |

### Active Workflows

| Workflow | Entry Point | Function |
|----------|-------------|----------|
| `intake-orchestrator` | User submits domain + objective | Full pipeline: classify → brief → build |
| `pageflow-builder` | Spawned by intake | Plan → generate assets → build pages → deploy |
| `rerender-pages` | Manual or post-build | Re-assemble pages from stored components |

### Existing Infrastructure

- **Auth Service**: JWT-based authentication, user management, project scoping, subscription tiers. API routes at `/api/v1/auth`, `/api/v1/user`, `/api/v1/projects`, WebSocket support
- **API Gateway**: Gin-based HTTP, proxies to core manager, template/instance management
- **Kafka**: Inter-agent messaging via request/response topics
- **PostgreSQL**: Sites, pages, content_components, page_components, style_collections, assets, link_registry, orchestration_states
- **Kubernetes**: ai-persona-system namespace, Docker images, Terraform/Kustomize
- **Cloudflare**: Static site hosting and deployment

---

## Architecture Evolution

### Three-Tier Authority Model

Navigation and site structure decisions follow a hierarchy based on the scale of change:

**Tier 1 — Strategist Authority (new builds, major restructure, adopt)**
The strategist/planner plans the full site architecture and produces a recommended navigation structure. The nav agent validates and persists it. At this tier, the nav agent is largely deferential to the strategist's plan.

**Tier 2 — Nav Agent Authority (maintenance, minor additions)**
Adding or removing individual pages, minor adjustments. The nav agent makes autonomous incremental decisions: "this blog post doesn't go in nav" or "this new service page slots into the primary group." No strategist involvement.

**Tier 3 — Drift Detection and Reconciliation**
Periodically (or when triggered by accumulated changes), compare current nav structure against the original strategist plan. Detect whether drift has been intentional evolution or unintended degradation. The original plan is a reference point, not always the correct baseline — drift may represent valid evolution.

### Heartbeat Maintenance Model

Rather than event-driven notifications between agents, a periodic heartbeat (2-3 times daily) triggers maintenance sweeps:

- Each agent checks what has changed since its last run
- Agents read shared state from the database, not from messages
- Failure recovery is simple — re-run next cycle
- No coupling between agent workflows
- Reduces Kafka complexity
- On-demand invocation still available for urgent cases (e.g. new site launch validation)

The heartbeat scheduler is a simple cron-style trigger. Each agent determines its own work by querying for changes since `last_heartbeat_at`.

---

## Agent Families

### 1. Navigation Agent Family

**Owner:** `nav-agent` (orchestrator)

Navigation is a first-class entity, not a side-effect of page syncing. The nav agent owns the complete navigable structure of a site, organised into semantic categories.

#### Data Model

```sql
CREATE TABLE site_nav_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    group_key TEXT NOT NULL,           -- "primary", "legal", "blog", "products"
    group_label TEXT NOT NULL,         -- "Main Navigation", "Legal"
    group_type TEXT NOT NULL,          -- "primary", "subsection", "content", "legal", "utility", "external"
    parent_group_id UUID REFERENCES site_nav_groups(id),  -- for subsections
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, group_key)
);

CREATE TABLE site_nav_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    group_id UUID NOT NULL REFERENCES site_nav_groups(id),
    parent_item_id UUID REFERENCES site_nav_items(id),  -- nested items within a group
    label TEXT NOT NULL,
    url TEXT NOT NULL,
    page_id UUID REFERENCES pages(id),    -- NULL for external links
    item_type TEXT NOT NULL DEFAULT 'page_link',  -- "page_link", "external_link", "anchor", "section_header", "cta"
    position INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',  -- "active", "planned", "broken", "removed"
    metadata JSONB DEFAULT '{}',           -- icon, description, badge text, etc.
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_nav_groups_site ON site_nav_groups(site_id);
CREATE INDEX idx_nav_items_site ON site_nav_items(site_id);
CREATE INDEX idx_nav_items_group ON site_nav_items(group_id);
CREATE INDEX idx_nav_items_page ON site_nav_items(page_id);
```

#### Group Types

| Type | Purpose | Examples |
|------|---------|---------|
| `primary` | Main site navigation | Home, About, Services, Contact |
| `subsection` | Child navigation within a primary section | Individual services, product lines |
| `content` | Navigation for content areas | Blog categories, archives, tags |
| `legal` | Regulatory/compliance pages | Privacy, Terms, Cookie Policy, Disclaimers |
| `utility` | Useful but not primary | Careers, Press, Sitemap, Partner Program |
| `external` | Links to external properties | Documentation, Status Page, Social |
| `contextual` | Page-specific, relationship-driven | Related posts, sibling pages, entity links |

#### Nav Agent Responsibilities

**Always (regardless of flow):**
- Owns `site_nav_groups` and `site_nav_items` tables
- Single point of truth for "what is the current nav structure"
- Validates nav consistency before serving data (checks items against pages table for existence, build status, URL correctness)

**New build flow:**
- Receives strategist's recommended nav structure
- Validates it (no broken references, sensible grouping)
- Writes to nav tables
- May flag issues ("10 primary items is high, suggest regrouping")
- HITL-configurable for flagged issues

**Maintenance flow (heartbeat or on-demand):**
- Receives change event: "page X added/removed/renamed"
- Decides placement/removal based on rules and existing structure
- Updates nav tables
- Returns what changed for parent workflow to route

**Adopt flow:**
- Receives scraped nav data from existing site
- Parses into standard nav structure
- Maps to discovered page records
- Fills gaps, flags unmappable items
- Likely always HITL

#### Consumer Queries

Design agent requesting footer nav for homepage:
```sql
SELECT g.group_key, g.group_label, i.label, i.url, i.item_type, i.position
FROM site_nav_groups g
JOIN site_nav_items i ON i.group_id = g.id
LEFT JOIN pages p ON i.page_id = p.id
WHERE g.site_id = $1
  AND g.group_type IN ('primary', 'utility', 'legal')
  AND i.status = 'active'
  AND (i.page_id IS NULL OR p.build_status = 'deployed')
ORDER BY g.position, i.position;
```

Links agent checking all navigable internal pages:
```sql
SELECT DISTINCT i.url, i.page_id, p.name, p.build_status
FROM site_nav_items i
JOIN pages p ON i.page_id = p.id
WHERE i.site_id = $1
  AND i.item_type = 'page_link'
  AND i.status = 'active';
```

---

### 2. Links Agent Family

**Owner:** `links-orchestrator` (orchestrator, algorithmic — no LLM)

The links orchestrator coordinates link health across the site. All sub-agents are algorithmic, making no subjective decisions. Heartbeat-driven with on-demand capability.

#### Sub-Agents / Actions

| Agent/Action | Role | LLM? |
|-------------|------|------|
| `link-crawler` | Extract links from rendered HTML | No |
| `link-validator` | Check links against pages table and HTTP for externals | No |
| `link-registry-sync` | Update link_registry table with current state | No |
| `redirect-manager` | Create/manage redirects when URLs change | No |
| `affiliate-link-manager` | Manage affiliate URLs, cloaking, disclosure (phase 2) | No |

Each sub-agent can run independently or be orchestrated by the links-orchestrator during heartbeat sweeps.

#### Heartbeat Workflow

```
1. Query pages modified since last heartbeat
2. For each modified page → link-crawler (extract all links from HTML)
3. Batch all crawled links → link-validator (check internal + external)
4. Write results → link-registry-sync (update registry, mark broken/removed)
5. Check for URL changes → redirect-manager (create redirects, flag stale links)
6. Produce summary (broken count, new redirects, orphaned pages)
```

#### What Links Agent Does

- Extract all links from HTML (parsing, not judgment)
- Classify link type by URL pattern (internal/external/asset/anchor)
- Resolve internal links to page records
- HTTP HEAD checks for external links (periodic, rate-limited)
- Detect broken internal links (target page missing or not deployed)
- Detect orphaned pages (no inbound internal links)
- Generate redirects when page URLs change
- Maintain link counts per page (inbound/outbound)
- Track link text for accessibility (flag empty anchors)

#### What Links Agent Does NOT Do

- Decide *where* to place links in content (that's the content writer)
- Navigate site structure decisions (that's the nav agent)
- SEO link strategy analysis (that's the SEO agent)
- Suggest new links or related content (needs LLM, out of scope)

#### Data Model Extensions

The existing `link_registry` table likely needs extending:

```sql
-- Check/extend existing link_registry
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    link_context TEXT DEFAULT 'body';  -- "body", "nav_header", "nav_footer", "cta", "sidebar", "image"
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    link_type TEXT DEFAULT 'internal';  -- "internal", "external", "anchor", "asset"
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    last_checked_at TIMESTAMPTZ;
ALTER TABLE link_registry ADD COLUMN IF NOT EXISTS
    redirect_target TEXT;

-- New table for redirects
CREATE TABLE IF NOT EXISTS site_redirects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    from_url TEXT NOT NULL,
    to_url TEXT NOT NULL,
    redirect_type INTEGER DEFAULT 301,
    reason TEXT,                        -- "page_renamed", "page_removed", "url_restructure"
    created_by TEXT,                    -- agent type that created it
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, from_url)
);
```

#### Phase 2: Affiliate Link Management

```sql
CREATE TABLE IF NOT EXISTS affiliate_programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    program_name TEXT NOT NULL,
    program_config JSONB DEFAULT '{}',
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS affiliate_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    program_id UUID REFERENCES affiliate_programs(id),
    original_url TEXT NOT NULL,
    affiliate_url TEXT NOT NULL,
    cloak_path TEXT,                    -- "/go/product-name"
    product_name TEXT,
    last_validated_at TIMESTAMPTZ,
    status TEXT DEFAULT 'active'
);
```

---

### 3. Design Agent Family

Design is split into separate agents early because the responsibilities are large and distinct. Each agent has sensible defaults — if upstream data isn't available, it falls back gracefully.

#### Agents

| Agent | Responsibility | LLM? | Exists? |
|-------|---------------|------|---------|
| `brand-designer` | Colour scheme, typography, spacing, visual tone | Yes | Exists (partial, in webdesign-agent) |
| `layout-architect` | Page type skeletons, nav group placements, content zones | Yes | New |
| `style-generator` | CSS production from brand spec + layout | Yes (for now) | Exists (partial, in webdesign-agent) |
| `nav-layout-agent` | Maps nav groups to page positions per page type | Minimal | New |

#### Brand Designer

Runs once during initial build. Produces:

- Colour palette (primary, secondary, accent, neutrals)
- Typography scale (heading fonts, body fonts, sizes, line heights)
- Spacing system (section padding, component gaps)
- Visual tone/mood description (guides other agents)
- Image style direction (guides image-generator)

Output stored in `sites.content_data` under `brand_spec`. Rarely changes. In maintenance, only runs if client requests a rebrand.

#### Layout Architect

Decides page structure by page type. Produces **page layout definitions** — the mapping of nav groups to positions, content zones, and structural constraints.

```json
{
  "page_layouts": {
    "landing": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "minimal", "sticky": true },
        "footer": { "nav_groups": ["legal"], "style": "simple-row" },
        "sidebar": null
      },
      "content_zones": ["hero", "body", "cta"],
      "max_body_components": 6
    },
    "content": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "standard", "sticky": false },
        "footer": { "nav_groups": ["primary", "utility", "legal"], "style": "multi-column" },
        "sidebar": { "nav_groups": ["content"], "style": "vertical-list" }
      },
      "content_zones": ["hero", "body", "sidebar", "related"],
      "max_body_components": 10
    },
    "standard": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "standard", "sticky": false },
        "footer": { "nav_groups": ["primary", "utility", "legal"], "style": "multi-column" },
        "sidebar": null
      },
      "content_zones": ["hero", "body", "cta"],
      "max_body_components": 8
    },
    "tool": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "minimal", "sticky": true },
        "footer": { "nav_groups": ["legal"], "style": "simple-row" },
        "sidebar": null
      },
      "content_zones": ["tool", "supporting_content"],
      "max_body_components": 4
    }
  }
}
```

Stored in `sites.content_data` under `layout_definitions` or as a separate artifact. If not present, rendering falls back to sensible defaults: primary nav in header, primary + legal in footer, no sidebar.

#### Style Generator

Takes brand spec + layout architecture, produces CSS. Currently LLM-driven (webdesign-agent). Could become more algorithmic over time as patterns are codified. Produces the site stylesheet deployed to `assets/css/global.css`.

#### Relationship to Existing webdesign-agent

The current `webdesign-agent` remains as-is in the existing `pageflow-builder` workflow. The new agents (`brand-designer`, `layout-architect`, `style-generator`) are used in new workflows. Over time, the `webdesign-agent` may be deprecated in favour of the split agents, but there's no rush — it works.

---

### 4. Content Agent Family

Content agents are split by **what** they write because different content types require genuinely different approaches, constraints, and voice.

#### Agents

| Agent | Responsibility | LLM? | Exists? |
|-------|---------------|------|---------|
| `page-content-writer` | Marketing/editorial page content | Yes | Exists |
| `legal-content-agent` | Privacy, terms, disclaimers, compliance text | Template + minimal LLM | New |
| `seo-content-agent` | Meta titles, descriptions, structured data, robots | LLM for generation, algorithmic for validation | New |
| `product-content-writer` | Product reviews, descriptions from structured data | Yes | New (phase 2) |
| `research-agent` | Web research for content backing | Yes | Exists |
| `content-reviewer` | HITL or auto-eval review | Yes | Exists |

#### Page Content Writer (existing)

The workhorse. Handles marketing content for page sections. Already supports research sub-agents and component-aware writing. Needs one extension: accepting structured entity data as input alongside brief data (for entity-backed pages in future site types).

#### Legal Content Agent (new)

Template-based with jurisdiction awareness. Receives site info (company name, domain, contact details, applicable jurisdictions) and produces legal pages from vetted templates. Minimal LLM involvement — primarily template population and jurisdiction selection.

Also provides **legal constraints** to the content writer as input rules:

- Industry-specific required disclaimers (e.g., "not financial advice" for finance sites)
- Forbidden phrases per industry
- Required disclosure text for pages containing affiliate links
- Regulatory notices by jurisdiction

```json
{
  "legal_rules": {
    "industry": "finance",
    "required_disclaimers": [
      {
        "trigger": "any_financial_content",
        "text": "This information is for educational purposes only and does not constitute financial advice.",
        "placement": "section_footer"
      }
    ],
    "forbidden_phrases": [
      "guaranteed returns", "risk-free", "you should invest"
    ],
    "required_pages": [
      { "name": "privacy", "template": "privacy-gdpr-uk", "nav_group": "legal" },
      { "name": "terms", "template": "terms-standard", "nav_group": "legal" },
      { "name": "disclaimer", "template": "financial-disclaimer", "nav_group": "legal" }
    ]
  }
}
```

#### SEO Content Agent (new)

Generates and optimises meta content. Runs after page content is written (needs the content to summarise). Responsibilities:

- Meta titles (with length validation, keyword inclusion)
- Meta descriptions (summarise page content, respect character limits)
- Structured data / JSON-LD (per page type — Article, Product, LocalBusiness, etc.)
- Robots directives
- Canonical URLs
- Open Graph / social metadata

Algorithmic for validation (length checks, schema validation), LLM for generation (writing compelling meta descriptions).

---

### 5. Entity Data Agent (new concept)

**Owner:** `entity-data-agent` (orchestrator)

Manages structured data that generates pages. Products, events, people, venues, tools — any real-world entity that a site presents.

#### Why This Matters

Three of four tested site types need it:
- **E-commerce**: Products (name, price, specs, images, merchant URL)
- **Events/tickets**: Events, fighters, venues (with relationships)
- **Design platform**: Users, projects, sites (with state)

Only brochure sites work purely from creative briefs. Everything else needs structured external data flowing into the content pipeline.

#### Data Model

```sql
CREATE TABLE IF NOT EXISTS site_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    entity_type TEXT NOT NULL,          -- "product", "event", "person", "venue", "tool"
    entity_key TEXT NOT NULL,           -- unique within site+type, e.g. "kitchen-aid-mixer"
    entity_data JSONB NOT NULL DEFAULT '{}',
    source TEXT DEFAULT 'manual',       -- "manual", "api", "scraped"
    source_url TEXT,
    page_id UUID REFERENCES pages(id), -- the page generated from this entity
    status TEXT DEFAULT 'active',
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, entity_type, entity_key)
);

CREATE TABLE IF NOT EXISTS site_entity_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    from_entity_id UUID NOT NULL REFERENCES site_entities(id),
    to_entity_id UUID NOT NULL REFERENCES site_entities(id),
    relationship_type TEXT NOT NULL,    -- "fights_in", "held_at", "related_to", "category_of"
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_entities_site_type ON site_entities(site_id, entity_type);
CREATE INDEX idx_entities_page ON site_entities(page_id);
CREATE INDEX idx_entity_rels_from ON site_entity_relationships(from_entity_id);
CREATE INDEX idx_entity_rels_to ON site_entity_relationships(to_entity_id);
```

#### How It Integrates

The entity data agent sits between data sourcing and content creation:

1. **Intake**: Entity data is sourced (API, scraping, manual entry)
2. **Storage**: Entities stored with relationships in `site_entities`
3. **Page generation**: Each entity optionally generates a page (entity → page mapping)
4. **Content writing**: Content writer receives entity data alongside brief context
5. **Nav integration**: Entity relationships feed contextual nav groups
6. **Links**: Internal links between entity pages are relationship-driven

This agent is expandable — new entity types can be added without architectural changes. The `entity_data` JSONB field accommodates any schema per entity type.

---

### 6. Tool Builder Agent (phase 2)

For interactive components — calculators, configurators, simple tools. Two approaches:

**LLM-generated tools**: For simpler tools (mortgage calculators, unit converters, simple simulations). The LLM produces self-contained HTML/CSS/JS. The uploaded mind map and layout engine examples demonstrate this is feasible — modern LLMs can produce working interactive tools.

**Pre-built component library**: For complex, tested tools (Monte Carlo simulations, rich editors, collaborative tools). Engineers build and test these; the agent selects and configures them.

The tool builder agent decides which approach to use based on tool complexity and whether a pre-built version exists in the library.

#### Component Library Tiers

| Tier | Description | Creation | Examples |
|------|-------------|----------|---------|
| Static | HTML templates with CSS | Existing component library | Hero, services-grid, testimonials |
| Dynamic | Self-contained JS applications in a page | LLM-generated or pre-built | Calculators, data visualisations, interactive tools |
| Application | Full web apps with API integration | Engineer-built only | Mood boards, layout editors, real-time collaboration |

---

### 7. News and Content Feed Agent Family

**Owner:** `content-feed-orchestrator` (orchestrator)

Most sites benefit from fresh content — news articles, blog posts, industry updates, commentary. The content feed system handles sourcing, deduplication, rewriting, publication, and lifecycle management of recurring content.

#### The Pipeline

News content flows through a distinct pipeline that's different from static page creation:

```
Sources (RSS, API, scrape, LLM-generated)
    ↓
Ingestion (raw content into staging)
    ↓
Deduplication (same story from multiple sources)
    ↓
Triage (is this relevant? how urgent? what angle?)
    ↓
Rewriting (LLM rewrites to original content in site voice)
    ↓
Entity linking (connect to site entities — fighters, products, events)
    ↓
Publication (create page, assign to content nav group, deploy)
    ↓
Lifecycle (featured → current → archive → pruned)
```

#### Sub-Agents

| Agent | Role | LLM? |
|-------|------|------|
| `feed-ingester` | Fetch from configured sources, store raw items | No |
| `feed-deduplicator` | Detect duplicate/near-duplicate stories | Minimal (similarity scoring, could be algorithmic) |
| `feed-triage` | Assess relevance, urgency, angle for site | Yes (needs editorial judgment) |
| `article-rewriter` | Rewrite raw content into original articles in site voice | Yes |
| `feed-publisher` | Create page from rewritten article, deploy | No (uses existing deploy pipeline) |
| `feed-lifecycle` | Age, archive, prune old content | No (algorithmic, time-based) |

Each can run independently or be orchestrated by `content-feed-orchestrator`. The orchestrator runs on heartbeat schedule — check sources, ingest new items, process the rewriting queue, manage lifecycle.

#### Source Types

| Source | Mechanism | Example |
|--------|-----------|---------|
| RSS/Atom feeds | HTTP fetch, parse XML | ESPN boxing RSS, industry blog feeds |
| API feeds | Authenticated API calls, parse JSON | Sports data providers, news APIs (NewsAPI, etc.) |
| Web scraping | HTTP fetch, HTML parse, extract content | Competitor news sections, press release pages |
| LLM-generated | Prompt-driven original content | "Write a weekly roundup of mortgage rate trends" |
| Manual/HITL | Human submits article draft or topic | Editor writes or pastes content for processing |

#### Data Model

```sql
-- Where we get content from
CREATE TABLE IF NOT EXISTS content_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_name TEXT NOT NULL,
    source_type TEXT NOT NULL,              -- "rss", "api", "scrape", "llm", "manual"
    source_config JSONB NOT NULL DEFAULT '{}',  -- URL, auth, selectors, prompts, etc.
    check_interval_minutes INTEGER DEFAULT 60,
    content_categories TEXT[] DEFAULT '{}', -- which content categories this feeds
    is_active BOOLEAN DEFAULT true,
    last_checked_at TIMESTAMPTZ,
    last_item_at TIMESTAMPTZ,              -- timestamp of most recent item found
    error_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Raw ingested items (staging, before rewriting)
CREATE TABLE IF NOT EXISTS content_feed_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_id UUID NOT NULL REFERENCES content_sources(id),
    external_id TEXT,                      -- source's own ID for dedup
    raw_title TEXT NOT NULL,
    raw_content TEXT,
    raw_summary TEXT,
    source_url TEXT,                        -- original article URL
    source_author TEXT,
    published_at TIMESTAMPTZ,              -- when the source published it
    ingested_at TIMESTAMPTZ DEFAULT NOW(),
    content_hash TEXT,                      -- for deduplication
    dedup_cluster_id UUID,                 -- groups duplicate stories together
    triage_status TEXT DEFAULT 'pending',   -- "pending", "approved", "rejected", "rewriting", "published"
    triage_score FLOAT,                    -- relevance/urgency score from triage
    triage_angle TEXT,                     -- suggested angle for rewriting
    triage_categories TEXT[],              -- suggested content categories
    metadata JSONB DEFAULT '{}',           -- images, tags, entities mentioned
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_feed_items_site ON content_feed_items(site_id);
CREATE INDEX idx_feed_items_status ON content_feed_items(triage_status);
CREATE INDEX idx_feed_items_hash ON content_feed_items(content_hash);
CREATE INDEX idx_feed_items_source ON content_feed_items(source_id);
CREATE INDEX idx_feed_items_published ON content_feed_items(published_at DESC);
```

Rewritten articles become entries in `site_entities` with `entity_type = 'article'`:

```json
{
  "entity_type": "article",
  "entity_key": "boxing-joshua-fury-rematch-confirmed-2026-02",
  "entity_data": {
    "title": "Joshua vs Fury Rematch Confirmed for June",
    "slug": "joshua-fury-rematch-confirmed",
    "content_html": "<p>The long-awaited rematch...</p>",
    "summary": "Anthony Joshua and Tyson Fury will meet again...",
    "author": "boxing-tickets.com editorial",
    "category": "fight-news",
    "tags": ["joshua", "fury", "heavyweight"],
    "featured_image": "/assets/images/joshua-fury-2.jpg",
    "published_at": "2026-02-08T14:00:00Z",
    "source_attribution": "Based on reporting from ESPN and BoxingScene",
    "source_feed_item_id": "uuid-of-raw-item",
    "related_entities": ["fighter:anthony-joshua", "fighter:tyson-fury", "event:fury-joshua-2"],
    "lifecycle_status": "current",
    "expires_at": null
  },
  "source": "feed_rewrite",
  "page_id": "uuid-of-generated-page"
}
```

This means the existing entity-data-agent pipeline handles articles the same way it handles products or events. The content writer receives entity data, the nav agent places articles in content nav groups, and the links agent tracks cross-references to related entities.

#### Deduplication Strategy

The same boxing news story might appear on ESPN, BoxingScene, and Sky Sports RSS within minutes of each other. The deduplicator needs to:

1. **Hash-based exact match**: `content_hash` on normalised title + summary catches identical syndicated content
2. **Similarity scoring**: For stories about the same event but written differently — compare title similarity, mentioned entities, publication time proximity
3. **Cluster assignment**: Group duplicate stories under a `dedup_cluster_id`. The triage agent sees the cluster and picks the best source (most detail, most authoritative) as the basis for rewriting

This is mostly algorithmic (text similarity, entity extraction, time-window grouping). Minimal LLM involvement — possibly for edge cases where algorithmic similarity is ambiguous.

#### Rewriting Constraints

The article-rewriter is the LLM-heavy sub-agent. It must:

- Produce genuinely original content (not just synonym substitution)
- Preserve factual accuracy from the source material
- Write in the site's editorial voice and tone (from brand spec)
- Add context from entity data (link to fighter profiles, event pages)
- Include proper attribution ("Based on reporting from..." or "Sources indicate...")
- Respect the legal-content-agent's rules (disclaimers, forbidden phrases)
- Generate meta content (title, description) or flag for seo-content-agent
- Suggest a category and tags for nav/content organisation

The rewriter receives: raw feed item, brand spec (for voice), entity data (for context), legal rules (for constraints), existing articles (to avoid duplicating recent coverage).

#### Lifecycle Management

The feed-lifecycle agent runs on heartbeat and manages content aging:

| Age | Status | Behaviour |
|-----|--------|-----------|
| 0-24 hours | `featured` | Appears in "Latest News" prominently, may be on homepage |
| 1-7 days | `current` | Appears in category listings, searchable |
| 7-30 days | `aging` | Drops from prominent positions, still in archives |
| 30-90 days | `archive` | In archive pages only, may be de-indexed (noindex) |
| 90+ days | `prune_candidate` | Could be removed or retained based on traffic/SEO value |

These thresholds are configurable per site. A news-heavy site like `boxing-tickets.com` might have aggressive aging. A blog-style site might keep content current for months.

The lifecycle agent also:
- Updates homepage/category page content to reflect current articles
- Manages "related articles" sections (using entity relationships)
- Triggers rerender of affected pages when article status changes
- Flags articles with ongoing high traffic for retention regardless of age

#### Feed Generation (outbound)

Sites with articles should also produce their own feeds — RSS/Atom for subscribers, JSON feed for API consumers, sitemap entries for search engines.

The feed-publisher generates these as static files deployed alongside the site:
- `/feed.xml` — RSS 2.0 feed of recent articles
- `/feed.json` — JSON Feed format
- `/sitemap-articles.xml` — article-specific sitemap with lastmod dates

These are regenerated on each heartbeat when new articles are published.

#### How It Fits the Architecture

The content-feed-orchestrator participates in two modes:

**Heartbeat mode (2-3x daily):**
1. feed-ingester checks all active sources for new items
2. feed-deduplicator clusters new items
3. feed-triage assesses relevance and priority
4. article-rewriter processes approved items from queue
5. feed-publisher creates pages and deploys
6. feed-lifecycle ages existing content
7. Regenerate RSS/JSON/sitemap feeds

**On-demand mode:**
- Manual article submission via HITL → goes straight to rewriting queue
- Breaking news trigger → ingester runs immediately for specific source
- Bulk import → ingester processes a batch of items

The content-feed-orchestrator is called by the heartbeat-scheduler agent alongside the links-orchestrator, SEO agent, and other maintenance agents. Each runs independently, reading shared state from the database.

---

## New Workflow: Component-Based Builder v2

A new workflow that uses the split agent families. Copies the structure of `pageflow-builder` but uses the new agents. The existing `pageflow-builder` remains untouched.

### Flow Overview

```
intake-orchestrator
  ├── site-classifier
  ├── briefing-agent
  └── component-builder-v2          (NEW workflow)
        ├── site-planner             (existing, may extend output)
        ├── brand-designer           (NEW, or reuse existing)
        ├── layout-architect         (NEW)
        ├── nav-agent                (NEW)
        │     └── populates nav tables from planner output
        ├── legal-content-agent      (NEW)
        │     └── produces legal pages + constraint rules
        ├── image-generator          (existing)
        ├── style-generator          (from webdesign-agent lineage)
        ├── page-content-writer      (existing)
        │     └── receives: brief, entity data, legal constraints, link context
        ├── content-reviewer         (existing)
        ├── deployer-agent           (existing)
        ├── seo-content-agent        (NEW, post-content)
        └── links-orchestrator       (NEW, post-deploy validation)

heartbeat-scheduler (periodic, 2-3x daily)
  ├── content-feed-orchestrator      (NEW)
  │     ├── feed-ingester            → check sources, stage raw items
  │     ├── feed-deduplicator        → cluster duplicates
  │     ├── feed-triage              → assess relevance (LLM)
  │     ├── article-rewriter         → rewrite to original content (LLM)
  │     ├── feed-publisher           → create pages, deploy
  │     └── feed-lifecycle           → age, archive, prune
  ├── links-orchestrator             (link health sweep)
  ├── seo-content-agent              (meta content audit)
  ├── nav-agent                      (consistency check, drift detection)
  └── (future: marketing, analytics, etc.)
```

### Key Differences from pageflow-builder

1. **Nav is separate**: Nav agent creates nav structure from planner output before page building starts. Rendering queries nav tables directly instead of deriving from page booleans.

2. **Legal pages handled separately**: Legal content agent produces legal pages from templates before the main content loop. These are predictable, don't need the full content writer pipeline.

3. **Design is split**: Brand spec produced first, then layout definitions, then CSS. Each stored as separate artifacts on the site record.

4. **SEO runs post-content**: After all pages are written, SEO content agent generates/optimises meta content in a sweep.

5. **Links validation post-deploy**: After all pages are deployed, links orchestrator validates the full link graph.

6. **Layout definitions drive rendering**: The `render_site_components` and `assemble_page` actions consume layout definitions to determine which nav groups appear where, rather than hardcoded header/footer queries.

---

## Site Type Stress Tests

### Brochure Site (existing, works)
All agents have clear roles. Nav is simple (primary + legal). No entity data needed. Legal is standard templates. Links are straightforward internal links.
**News/Feed**: Optional but beneficial — industry news blog section sourced from RSS feeds. Keeps site fresh for SEO. Content-feed-orchestrator with industry-relevant RSS sources, moderate rewriting.

### E-commerce / Product Review Site
**Nav**: primary + content (categories) + legal (including affiliate disclosure). Works with categorised groups.
**Content**: Product content writer needed — writes reviews from structured product data, not creative briefs.
**Entity data**: Products are entities with structured fields (name, price, specs, merchant URL, images).
**Links**: Commercial outbound links to merchants. Links agent manages tracking parameters. Affiliate link manager (phase 2).
**News/Feed**: Product news, deal alerts, new product announcements. Sources: manufacturer RSS, deal aggregator APIs. High rewriting requirement to differentiate from competitors covering the same products. Feed lifecycle fast — deal content ages in days, not weeks.
**Gap**: Product data sourcing — entity-data-agent with API integration to product feeds.

### Finance / Tools Site
**Nav**: primary + tools (calculator pages) + content (guides) + legal (extensive). Works with categorised groups.
**Content**: Guides are standard editorial. Calculator pages need tool-builder agent for interactive components. Legal constraints pervasive — every financial content page needs disclaimers.
**Entity data**: Limited (mortgage products, lenders could be entities).
**Tools**: Mortgage calculators, affordability tools — LLM-generated dynamic components or pre-built library.
**Legal**: Per-page disclaimers, forbidden phrases, regulatory notices. Legal content agent provides rules to content writer.
**News/Feed**: Rate changes, market updates, regulatory news, economic indicators. Sources: financial news APIs, central bank RSS, regulatory body feeds. Legal constraints apply heavily to rewritten news — article-rewriter must receive legal rules as input. Disclaimer injection mandatory on every financial news article.
**Gap**: Interactive component generation. Server-side hosting for dynamic features (future step, not a design limitation).

### Events / Tickets Site
**Nav**: primary + content (news categories) + contextual (per-event, per-fighter). Contextual nav populated from entity relationships.
**Content**: Editorial (news, previews) + entity-backed (event pages, fighter profiles). Content writer needs structured entity data as input.
**Entity data**: Events, fighters, venues with relationships (fights_in, held_at). Core requirement.
**Links**: Dense internal linking driven by entity relationships. Commercial outbound links to ticketing platforms.
**News/Feed**: Core to the site. Fight announcements, weigh-in results, post-fight analysis, rankings changes, injury reports. Sources: sports news RSS (ESPN, BoxingScene), sports data APIs, social media. Entity linking is heavy — every article references fighters and events. The feed-triage agent needs strong relevance filtering (boxing news only, not general sports). Lifecycle tied to event calendar — pre-fight content peaks before events, results content peaks immediately after.
**Temporal**: Events have dates, become historical. News becomes stale. Heartbeat handles lifecycle transitions.

### Interactive Platform Site (e.g. website-design.com)
**Marketing pages**: Standard brochure pipeline — homepage, features, pricing, about.
**Application pages**: Pre-built application components (mood board, layout editor, mind map). These are engineer-built Tier 3 components.
**API layer**: Agents exposed as external HTTP endpoints via existing auth-service gateway. Agent-as-API pattern.
**Multi-tenant**: User-scoped site building. Each user's site build is a separate orchestration context.
**HITL**: User IS the human in the loop. HITL interfaces become product UIs, not internal review tools.
**News/Feed**: Blog/tutorial content — web design trends, platform updates, how-to guides. Partially LLM-generated (source_type="llm"), partially sourced from web design RSS feeds. Lower volume, longer lifecycle — educational content stays relevant for months.

---

## Data Ownership Summary

| Data | Owner Agent | Tables |
|------|-------------|--------|
| Site record, brand_assets, content_data | site-planner / brand-designer | `sites` |
| Page records, sections | site-planner | `pages` |
| Navigation structure | nav-agent | `site_nav_groups`, `site_nav_items` |
| Link registry | links-orchestrator | `link_registry` |
| Redirects | redirect-manager (sub of links) | `site_redirects` |
| Page component HTML | page-content-writer / page-rerender | `page_components` |
| Site-level components | render_site_components action | `site_components` |
| Style collection | style-generator / webdesign-agent | `style_collections` |
| Content components | (library, manually maintained) | `content_components` |
| Entity data | entity-data-agent | `site_entities`, `site_entity_relationships` |
| Layout definitions | layout-architect | `sites.content_data.layout_definitions` |
| Brand spec | brand-designer | `sites.content_data.brand_spec` |
| Legal rules | legal-content-agent | `sites.content_data.legal_rules` |
| SEO metadata | seo-content-agent | `pages` (meta fields) or `page_seo` table |
| Affiliate programs | affiliate-link-manager | `affiliate_programs`, `affiliate_links` |
| Content sources | content-feed-orchestrator | `content_sources` |
| Raw feed items | feed-ingester | `content_feed_items` |
| Published articles | article-rewriter / feed-publisher | `site_entities` (type=article) + `pages` |

---

## Implementation Phasing

### Phase 1 — Foundation (build now)

| Agent | Priority | Reason |
|-------|----------|--------|
| `nav-agent` | High | Fixes current nav issues, enables categorised groups |
| `links-orchestrator` | High | Link health across all site types, heartbeat-ready |
| `heartbeat-scheduler` | High | Required for maintenance model, drives links/feed/SEO sweeps |
| `brand-designer` | Medium | Separate from webdesign-agent, cleaner output |
| `layout-architect` | Medium | Enables page-type-specific nav placement |
| `style-generator` | Medium | Separated CSS generation |
| `legal-content-agent` | Medium | Template-based legal pages + constraint rules |
| `seo-content-agent` | Medium | Post-content meta generation |

Schema work:
- `site_nav_groups` and `site_nav_items` tables
- `site_redirects` table
- `link_registry` extensions
- Layout definitions in `sites.content_data`

New workflow:
- `component-builder-v2` — copies pageflow-builder structure, uses new agents

### Phase 2 — Content and Data

| Agent | Priority | Reason |
|-------|----------|--------|
| `content-feed-orchestrator` | High | News/article pipeline — most sites need fresh content |
| `feed-ingester` | High | Source checking, raw content staging |
| `article-rewriter` | High | LLM rewrite to original content |
| `feed-publisher` | High | Article page creation and deployment |
| `feed-lifecycle` | Medium | Content aging, archiving, pruning |
| `entity-data-agent` | High | Required for e-commerce, events, any data-driven site |
| `product-content-writer` | Medium | Structured product reviews from entity data |
| `tool-builder-agent` | Medium | LLM-generated interactive components |
| `affiliate-link-manager` | Low | Commercial link management |

Schema work:
- `content_sources` table
- `content_feed_items` table
- `site_entities` and `site_entity_relationships` tables
- `affiliate_programs` and `affiliate_links` tables
- Component library tier 2 (dynamic components)

### Phase 3 — Platform

| Capability | Priority | Reason |
|-----------|----------|--------|
| Agent-as-API layer | High | Expose agents via HTTP for platform products |
| Multi-tenant scoping | High | User-scoped site building |
| Dynamic component library | Medium | Pre-built interactive tools (calculators, editors) |
| Real-time collaboration | Low | WebSocket-based collaborative editing |
| Agent marketplace | Low | User-composed agent workflows |

---

## Migration Notes

- The existing `pageflow-builder` workflow is NOT modified. It continues to work as-is.
- New workflow `component-builder-v2` is registered as a new builder type in `agent_definitions`
- The `site-classifier` may need updating to recommend `component-builder-v2` for new projects
- Existing sites built with `pageflow-builder` are unaffected
- The `in_header` / `in_footer` booleans on the `pages` table remain for backward compatibility with existing workflows; new workflows use nav tables instead
- The `webdesign-agent` remains available for existing workflows; new workflows use the split design agents

---

## Resolved Decisions

1. **Nav agent during build sequence**: Nav agent runs after planner, before content loop. Content writers need the nav structure finalised to know what pages exist for internal linking. Rendering needs nav tables populated before it starts.

2. **Heartbeat trigger mechanism**: A dedicated `heartbeat-scheduler` agent. More consistent with the architecture, observable via the same logging/monitoring as other agents, and can itself be orchestrated or configured via the standard agent definitions.

3. **Layout definitions storage**: JSONB on `sites.content_data` under `layout_definitions` key. Simpler, consumed as a whole. Can migrate to a separate table later if cross-site querying becomes needed.

4. **Legal rules scope**: Per-site, stored on the site record (`sites.content_data.legal_rules`). Duplicative but provides the granularity needed — each site may have specific legal requirements even within the same industry. Templates can seed common rules per industry, but the live rules belong to the site.

5. **Entity data sourcing**: API-first. Ticket feeds, product APIs, and similar structured sources will arrive soon. Manual/HITL entry for initial seeding (e.g., hand-entering products). Both paths write to the same `site_entities` table — the `source` field distinguishes origin.
