# Agent Orchestration Architecture v2

## Overview

This document captures the evolved agent architecture for intelligent multi-page website building and ongoing site maintenance. It extends the existing system (intake-orchestrator → classifier → briefing → pageflow-builder pipeline) with specialised agent families for navigation, links, design, content, and maintenance — each with clear responsibilities, data ownership, and interaction patterns.

The guiding principles are:

- **Agents are separate early** — even if responsibilities are small now, the structure supports growth without refactoring
- **New workflows rather than modifying existing ones** — copy and extend, don't break what works
- **Complexity in Go actions, simplicity in workflows** — workflows remain readable orchestration; business logic lives in action code
- **Algorithmic where possible, LLM where necessary** — not every agent needs model calls
- **Heartbeat-driven maintenance** — CronJob-triggered, per-site maintenance orchestrators with discovery/triage/fix cycle
- **Build and maintain are separate spawn chains** — build agents create from nothing, fix agents change specific things

---

## Design System Layers

The design system has three independent layers that vary separately. Understanding these layers is fundamental to how components, themes, and sites relate.

### Layer 1: HTML Components (structure & layout)

**Table:** `content_components`

Self-contained HTML blocks — a hero section, a testimonial grid, a FAQ accordion, a services card layout. Each has its own inline `<style>` for layout (grid, flexbox, spacing) and dark section overrides.

Multiple components exist for the same function (e.g. `hero-split`, `hero-fullwidth`, `hero-minimal`). What varies is structure — a "hero" could be full-bleed background image, split-panel with image left, or minimal centered text.

Components reference CSS variables with fallbacks: `var(--color-primary, #1a1a2e)`. They never hardcode brand colours. Dark sections (testimonials on dark backgrounds, CTA sections, footers) set `color: #fff` on their container element, and all children inherit light text automatically.

### Layer 2: CSS Theme (appearance)

**Table:** `css_themes`

A complete base stylesheet setting `:root` variables (colours, fonts, spacing), base resets (box-sizing, margin), typography scale (h1-h6 sizes), button styles, responsive breakpoints, and accessibility focus states.

What varies between themes: a "finance-dark" theme has navy/gold palette, tight spacing, system fonts. A "creative-bold" theme has vibrant colours, generous spacing, display fonts. A "minimal-light" theme has muted tones, lots of whitespace, serif headings.

Deployed as `/assets/css/styles.css` — one per site, committed to the site's git repo by the webdesign agent.

### Layer 3: Style Collection (the bridge)

**Table:** `style_collections`

A named grouping that ties together a header component + footer component + CSS theme + colour palette. Think of it as a "design kit."

| Field | References | Purpose |
|-------|-----------|---------|
| `header_component_id` | `content_components` | Which header HTML component to use |
| `header_home_component_id` | `content_components` | Alternate header for homepage (optional) |
| `footer_component_id` | `content_components` | Which footer HTML component to use |
| `css_theme_id` | `css_themes` | Which CSS theme to use |
| `color_palette` | JSONB | Primary, secondary, accent, background, text colours |
| `typography` | JSONB | Font family, base size, line height, heading font |

Examples: `professional-dark`, `modern-light`, `bold-creative`, `clean-minimal`.

### How They Connect

```
site (leopardessconsulting.co.uk)
  └── style_collection_id → style_collections (professional-dark)
        ├── header_component_id → content_components (header-professional-dark)
        ├── footer_component_id → content_components (footer-4-column)
        ├── css_theme_id → css_themes (base stylesheet → /assets/css/styles.css)
        └── color_palette → {primary: "#1a1a2e", accent: "#0f3460", ...}

  └── pages (from site plan)
        ├── index.html
        │     ├── hero component (content_components: hero-fullwidth)
        │     ├── differentiators component (content_components: differentiators-grid)
        │     ├── testimonials component (content_components: social-proof)
        │     └── cta component (content_components: call-to-action)
        ├── about.html
        │     └── ...different body components...
        └── services.html
              └── ...different body components...

  All pages share:
    - Same header (from style collection)
    - Same footer (from style collection)
    - Same /assets/css/styles.css (from CSS theme)
    - Same :root variables (from colour palette via head component)
```

### Mix and Match

Same components, different themes:
- Hero + services-grid + testimonials on `professional-dark` → navy bg, system fonts, tight spacing
- Hero + services-grid + testimonials on `creative-bold` → coral accent, display fonts, generous spacing

Same theme, different components:
- `professional-dark` + hero-fullwidth + testimonials-grid → big hero, card layout
- `professional-dark` + hero-minimal + testimonials-carousel → text hero, scrolling quotes

Same theme, different pages:
- Site A: 5 pages (home, about, services, case-studies, contact)
- Site B: 3 pages (home, services, contact) with different body components on each

### The Colour Inheritance Model

This is the single most important rule in the design system. Getting it wrong causes unreadable text (light text on light backgrounds).

```
/assets/css/styles.css (from css_themes)
  body { color: var(--color-text); }          ← sets default dark text for light sections
    ↓ inherits
    h1, h2, p, li, blockquote, strong         ← all inherit dark text (NO explicit color set)

Component inline <style>
  .social-proof-section { color: #fff; }      ← dark section overrides to light text
    ↓ inherits
    h2, blockquote, cite, p, strong           ← all inherit light text automatically
```

The rules for `styles.css`:

- `body` sets `color: var(--color-text)` — the ONLY place default text colour is set
- `h1-h6` use `color: inherit` (NOT `var(--color-primary)`)
- `p`, `li`, `blockquote`, `strong`, `em`, `cite` — do NOT set `color` at all
- `a` — `color: var(--color-accent)` is the one exception (links are explicit)
- `blockquote` — do NOT set `background-color` (components handle this contextually)

If the base CSS forces `color: var(--color-text)` on `p` or `blockquote`, dark sections break because children can't inherit `color: #fff` from their parent container.

### Theme Library Growth

The theme library grows over time:

1. Build a site → webdesign agent generates CSS → stored as new `css_themes` row
2. Tag it with industry, style category, colour characteristics
3. Next similar brief → search existing themes → reuse if match found, adjust `:root` values
4. Collect inspiration from external sites → extract their patterns → inform design spec → generate or select theme
5. The webdesign agent goes from "always generate" to "search → maybe reuse → maybe generate → always store"

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
| `webdesign-agent` | specialist | Generates design spec and CSS (see Design Agent Family) |
| `deployer-agent` | specialist | Git commit and Cloudflare deployment |
| `page-rerender` | specialist | Re-assembles single page from stored components |
| `rerender-pages` | orchestrator | Batch rerender across all site pages |

### Active Workflows

| Workflow | Entry Point | Function |
|----------|-------------|----------|
| `intake-orchestrator` | User submits domain + objective | Full pipeline: classify → brief → build |
| `pageflow-builder` | Spawned by intake | Plan → generate assets → build pages → deploy |
| `rerender-pages` | Manual or post-build | Re-assemble pages from stored components |

### Implemented Actions (not yet separate agents)

| Action | Runs In | Status | Function |
|--------|---------|--------|----------|
| `populate_nav_tables` | pageflow-builder | Deployed | Classifies pages into nav groups, populates `site_nav_items` |
| `GetNavItems()` | component_library.go | Deployed | Shared query function — reads nav tables with pages-table fallback |

These implement nav-agent responsibilities as actions within the existing pipeline. They will migrate to a standalone nav-agent when that's warranted.

### Existing Infrastructure

- **Auth Service**: JWT-based authentication, user management, project scoping, subscription tiers. API routes at `/api/v1/auth`, `/api/v1/user`, `/api/v1/projects`, WebSocket support
- **API Gateway**: Gin-based HTTP, proxies to core manager, template/instance management
- **Kafka**: Inter-agent messaging via request/response topics
- **PostgreSQL**: Sites, pages, content_components, page_components, style_collections, css_themes, site_nav_groups, site_nav_items, assets, link_registry, orchestration_states, site_entities, site_entity_relationships (exist), entity_sources, entity_sync_log, content_sources, content_feed_items, maintenance_findings, maintenance_tasks (planned)
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

Maintenance follows the same dynamic spawning pattern as the vet-batch-processor → vet-practice-verifier pipeline. A Kubernetes CronJob (every 8 hours) sends a Kafka message to agent-chassis, which spawns the maintenance-batch-scheduler. The scheduler queries which sites have maintenance due, claims a batch, and spawns a site-maintenance-orchestrator per site.

```
K8s CronJob (every 8 hours)
  → Kafka message to agent-chassis process topic
    → agent-chassis spawns maintenance-batch-scheduler (K8s job)
      → populates maintenance_tasks from site profiles
      → claims batch of N tasks (FOR UPDATE SKIP LOCKED)
      → groups by site
      → for each site:
          spawns site-maintenance-orchestrator (K8s job)
            → fixes pending items from previous cycles
            → verifies previous fixes
            → spawns discovery agents for due domains
            → triages new findings
            → updates last_run timestamps
```

Each site gets its own orchestrator instance. The orchestrator reads that site's `maintenance_profile` from `sites.settings` and spawns only the discovery agents that site needs. Different sites get different combinations — a finance site needs regulatory compliance, a brochure site needs content freshness and links only.

The batch scheduler's `batch_size` parameter controls concurrency — how many site orchestrators run simultaneously. This prevents flooding when thousands of sites are due.

A separate daily CronJob triggers the maintenance-catch-all agent, which handles stale findings, HITL reminders, cross-site pattern detection, and stuck task recovery.

Principles:
- Each site is independent — its own orchestrator, its own maintenance profile
- Discovery agents find problems, triage classifies them, fix agents resolve them
- Agents communicate through the `maintenance_findings` table, not by calling each other
- Failure recovery is simple — re-run next cycle, work items stay in queue
- On-demand invocation still available for urgent cases

See the **Maintenance Architecture Plan** document for full detail on discovery agents, fix agents, triage logic, and implementation phases.

---

## Agent Families

### 1. Navigation Agent Family

**Owner:** `nav-agent` (orchestrator)

Navigation is a first-class entity, not a side-effect of page syncing. The nav agent owns the complete navigable structure of a site, organised into semantic categories.

**Current implementation status:** The nav agent's core responsibilities are implemented as the `populate_nav_tables` action within the `pageflow-builder` workflow. This includes page classification into groups (primary, utility, legal), nav table population, and a shared `GetNavItems()` query function used by header/footer rendering. The full standalone nav-agent is planned but not yet needed.

#### Data Model

```sql
CREATE TABLE site_nav_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    group_key TEXT NOT NULL,
    group_label TEXT NOT NULL,
    group_type TEXT NOT NULL,
    parent_group_id UUID REFERENCES site_nav_groups(id),
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, group_key)
);

CREATE TABLE site_nav_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    group_id UUID NOT NULL REFERENCES site_nav_groups(id),
    parent_item_id UUID REFERENCES site_nav_items(id),
    label TEXT NOT NULL,
    url TEXT NOT NULL,
    page_id UUID REFERENCES pages(id),
    item_type TEXT NOT NULL DEFAULT 'page_link',
    position INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Group Types

| Type | Purpose | Examples |
|------|---------|---------|
| `primary` | Main site navigation | Home, About, Services, Contact |
| `subsection` | Child navigation within a primary section | Individual services, product lines |
| `content` | Navigation for content areas | Blog categories, archives, tags |
| `legal` | Regulatory/compliance pages | Privacy, Terms, Cookie Policy, Disclaimers |
| `utility` | Useful but not primary | FAQ, Careers, Press, Sitemap, Blog, Insights |
| `external` | Links to external properties | Documentation, Status Page, Social |
| `contextual` | Page-specific, relationship-driven | Related posts, sibling pages, entity links |

#### Utility Classification (implemented)

The `populate_nav_tables` action classifies pages into groups. Pages matching these names are routed to the `utility` group even if `in_header = true` on the pages table:

- FAQ, FAQs, Frequently Asked Questions
- Blog, Insights, News, Articles
- Careers, Jobs, Team
- Resources, Downloads, Guides
- Testimonials, Reviews
- Partners, Affiliates
- Sitemap

This keeps the primary nav clean (Home, About, Services, Case Studies, Contact) while utility pages appear in the footer.

#### Nav Data Flow (implemented)

```
sync_pages_to_db
  ↓ pages table populated
populate_nav_tables (action in pageflow-builder)
  ↓ classifies pages → site_nav_groups + site_nav_items (status='active')
  ↓ stores result in collectedData["nav_data"]
InjectHeader (component_library.go)
  ↓ GetNavItems(groups=["primary"], deployedOnly=true)
  ↓ reads nav tables directly, falls back to pages table if empty
InjectFooter (component_library.go)
  ↓ GetNavItems(groups=["primary","utility","legal"], deployedOnly=true)
multipage_actions / assemble_from_library
  ↓ extractNavItemsFromCollectedData() checks nav_data first, then db_sync
```

#### Nav Agent Responsibilities

**Always (regardless of flow):**
- Owns `site_nav_groups` and `site_nav_items` tables
- Single point of truth for "what is the current nav structure"
- Validates nav consistency before serving data

**New build flow:**
- Receives strategist's recommended nav structure
- Validates it (no broken references, sensible grouping)
- Writes to nav tables
- May flag issues ("10 primary items is high, suggest regrouping")

**Maintenance flow (heartbeat or on-demand):**
- Receives change event: "page X added/removed/renamed"
- Decides placement/removal based on rules and existing structure
- Updates nav tables

**Adopt flow:**
- Receives scraped nav data from existing site
- Parses into standard nav structure
- Maps to discovered page records
- Fills gaps, flags unmappable items

#### Consumer Queries

```sql
-- Footer nav for homepage
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

---

### 3. Design Agent Family

Design covers three distinct concerns mapped to the three design system layers: brand decisions (what it should look like), structural layout (how pages are organised), and CSS production (the actual stylesheet).

#### Agents

| Agent | Layer | Responsibility | LLM? | Status |
|-------|-------|---------------|------|--------|
| `webdesign-agent` | 2 (theme) | Analyses brand, generates design spec, produces CSS theme | Yes | Exists, prompt updated for colour inheritance |
| `brand-designer` | 2 (theme) | Colour scheme, typography, spacing, visual tone | Yes | Future split from webdesign-agent |
| `layout-architect` | 1 (components) | Page type skeletons, nav group placements, content zones | Yes | New |
| `style-generator` | 2 (theme) | CSS production from brand spec + layout | Yes (for now) | Future split from webdesign-agent |
| `nav-layout-agent` | 1+3 (components+bridge) | Maps nav groups to page positions per page type | Minimal | New |

#### Current: Webdesign Agent

The `webdesign-agent` handles both brand analysis and CSS generation in a single workflow:

1. **analyse_design** — LLM analyses site domain, industry, tagline → produces a design spec JSON (colour scheme, typography, spacing)
2. **generate_css** — LLM generates a complete `styles.css` from the design spec → deployed to git as `/assets/css/styles.css`
3. **update_site** — stores design spec on the site record

The CSS generation prompt enforces the colour inheritance model: `body` sets default text colour, all other elements inherit, no forced colours on `p`/`h1-h6`/`blockquote`/`li`/`strong`. This prevents dark sections from having unreadable text.

The generated CSS is effectively a new `css_themes` entry — it just isn't stored in that table yet. Connecting this is part of the theme library work.

#### Future: Split Design Agents

The webdesign-agent's two steps map cleanly to separate agents:

**Brand Designer** — runs once during initial build, produces:
- Colour palette (primary, secondary, accent, neutrals)
- Typography scale (heading fonts, body fonts, sizes, line heights)
- Spacing system (section padding, component gaps)
- Visual tone/mood description (guides other agents)
- Image style direction (guides image-generator)

Output stored in `sites.content_data` under `brand_spec`. Rarely changes.

**Style Generator** — takes brand spec + component list, produces CSS. Checks the theme library first (search by industry + style tags). If a matching theme exists, reuses it with adjusted `:root` values. If not, generates a new theme and stores it in `css_themes` for future reuse.

There's no rush on this split — the webdesign-agent works. The split happens when the theme library is large enough that "search and adapt" is more common than "generate from scratch."

#### Layout Architect

Decides page structure by page type. Produces page layout definitions:

```json
{
  "page_layouts": {
    "landing": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "minimal", "sticky": true },
        "footer": { "nav_groups": ["legal"], "style": "simple-row" }
      },
      "content_zones": ["hero", "body", "cta"],
      "max_body_components": 6
    },
    "standard": {
      "nav_placement": {
        "header": { "nav_groups": ["primary"], "style": "standard", "sticky": false },
        "footer": { "nav_groups": ["primary", "utility", "legal"], "style": "multi-column" }
      },
      "content_zones": ["hero", "body", "cta"],
      "max_body_components": 8
    }
  }
}
```

Stored in `sites.content_data` under `layout_definitions`. If not present, rendering falls back to sensible defaults: primary nav in header, primary + utility + legal in footer.

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

#### Legal Content Agent (new)

Template-based with jurisdiction awareness. Receives site info and produces legal pages from vetted templates. Also provides **legal constraints** to the content writer:

```json
{
  "legal_rules": {
    "industry": "finance",
    "required_disclaimers": [
      { "trigger": "any_financial_content", "text": "...", "placement": "section_footer" }
    ],
    "forbidden_phrases": ["guaranteed returns", "risk-free"],
    "required_pages": [
      { "name": "privacy", "template": "privacy-gdpr-uk", "nav_group": "legal" }
    ]
  }
}
```

#### SEO Content Agent (new)

Runs after page content is written. Meta titles, descriptions, structured data / JSON-LD, robots directives, canonical URLs, Open Graph. Algorithmic for validation, LLM for generation.

---

### 5. Entity Data Agent Family

**Owner:** `entity-data-agent` (orchestrator)

Manages structured data that generates pages. Products, events, people, venues, ticket tiers, tools — any real-world entity that a site presents.

Three of four tested site types need it (e-commerce, events/tickets, design platform). Only brochure sites work purely from creative briefs. First implementation target: boxing ticket/events site, then football tickets, then finance.

#### Core Data Model (exists)

```sql
CREATE TABLE IF NOT EXISTS site_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    entity_type TEXT NOT NULL,
    entity_key TEXT NOT NULL,
    entity_data JSONB NOT NULL DEFAULT '{}',
    source TEXT DEFAULT 'manual',
    source_url TEXT,
    page_id UUID REFERENCES pages(id),
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
    relationship_type TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Source Management (new)

Entity sources are configured per site, per entity type. Different entity types may poll at different frequencies — event ticket availability changes hourly, venue data changes weekly.

```sql
CREATE TABLE entity_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    entity_type TEXT NOT NULL,            -- 'event', 'performer', 'venue', 'ticket_tier'
    source_type TEXT NOT NULL,            -- 'api', 'scrape', 'manual', 'feed'
    source_url TEXT,
    source_config JSONB DEFAULT '{}',     -- API endpoint, credentials, mapping rules,
                                          -- pagination, rate limits
    field_mapping JSONB,                  -- maps source fields to entity_data fields
    poll_interval INTERVAL DEFAULT '1 hour',
    last_polled_at TIMESTAMPTZ,
    news_triggers JSONB,                  -- entity state changes that should trigger
                                          -- news articles (see Feed connection below)
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sync log for auditing and change detection
CREATE TABLE entity_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_id UUID REFERENCES entity_sources(id),
    entity_id UUID REFERENCES site_entities(id),
    sync_type TEXT NOT NULL,              -- 'created', 'updated', 'unchanged', 'removed'
    changes JSONB,                         -- what fields changed, old vs new values
    synced_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### Entity Types for Events/Tickets

For a boxing ticket site, the entity types and their relationships:

```
event (entity_type: 'event')
  entity_data: {
    name: "Fury vs Joshua",
    date: "2026-06-15",
    venue_entity_key: "wembley-stadium",
    status: "on_sale",            -- announced, on_sale, selling_fast, sold_out,
                                  -- event_day, past, cancelled
    card_position: "main_event",  -- main_event, co_main, undercard
    weight_class: "heavyweight",
    scheduled_rounds: 12,
    result: null                  -- filled after event: { winner, method, round }
  }

performer (entity_type: 'performer')
  entity_data: {
    name: "Tyson Fury",
    nickname: "The Gypsy King",
    record: { wins: 34, losses: 1, draws: 1 },
    weight_class: "heavyweight",
    nationality: "British",
    ranking: 1,
    image_url: "..."
  }

venue (entity_type: 'venue')
  entity_data: {
    name: "Wembley Stadium",
    city: "London",
    capacity: 90000,
    address: "...",
    coordinates: { lat: 51.556, lng: -0.279 }
  }

ticket_tier (entity_type: 'ticket_tier')
  entity_data: {
    tier_name: "Ringside",
    event_entity_key: "fury-vs-joshua-2026-06-15",
    price_gbp: 500,
    original_price_gbp: 500,
    available: 200,
    total: 250,
    purchase_url: "https://ticketmaster.co.uk/...",
    source: "ticketmaster"
  }

Relationships:
  event → performer (relationship_type: 'features', both directions)
  event → venue (relationship_type: 'held_at')
  event → ticket_tier (relationship_type: 'has_tickets')
  event → event (relationship_type: 'on_same_card')
  performer → performer (relationship_type: 'opponent_in', metadata: {event_id})
```

Football tickets follow a similar shape but with different entity types (match, team, competition, season) and different relationship patterns.

#### Sync Pipeline

```
entity-data-agent (orchestrator, CronJob-triggered):
  1. load_due_sources        → query entity_sources where poll is due
  2. for each source:
     └── fetch_and_sync      → call source API/feed
                               → for each item in response:
                                  - map fields via field_mapping
                                  - upsert into site_entities
                                  - write entity_sync_log entry
                                  - if sync_type = 'created' or 'updated':
                                    flag entity for page processing
  3. process_changes         → for entities flagged:
     ├── new entity           → decide: does it get a page?
     │                          create page record, generate content,
     │                          deploy (template-based, minimal LLM)
     ├── changed entity       → re-render affected page sections
     │                          from updated entity_data
     │                          (no full rewrite — template re-render)
     └── significant change   → check news_triggers config
                                if matches → write content_feed_item
                                for the feed pipeline to pick up
  4. update_last_polled      → update entity_sources.last_polled_at
  5. complete
```

#### Entity Lifecycle (state-based, not time-based)

Unlike news articles which decay over time, entity lifecycle is driven by status:

| Entity Status | Page Behaviour | Nav Behaviour |
|--------------|----------------|---------------|
| `announced` | Teaser page, no ticket links | May appear in "upcoming" section |
| `on_sale` | Full page with ticket tiers, buy links | Featured in nav, homepage widget |
| `selling_fast` | Full page with urgency messaging | Prominent placement |
| `sold_out` | Page stays, buy links replaced with waitlist/alternatives | Stays in nav |
| `event_day` | Page updated with event-day info (times, streaming) | Prominent |
| `past` | Results, stats, review links added. Buy links removed | Moves to "past events" |
| `cancelled` | Status notice, refund info | Removed from active nav |
| `historical` | Permanent SEO asset (people search for past fights) | Archive section |

Status transitions can be detected automatically from API data (tickets remaining = 0 → `sold_out`, event date passed → `past`) or manual (results entry after a fight).

#### What Entity Agent Does NOT Do

- Write editorial content about entities (that's the content writer or article-rewriter)
- Decide news angles (that's the feed-triage agent)
- Manage ticket purchasing or payment (external, via purchase_url links)
- SEO optimisation of entity pages (that's the seo-content-agent)

---

### 6. Tool Builder Agent (phase 2)

For interactive components — calculators, configurators, simple tools.

| Tier | Description | Creation | Examples |
|------|-------------|----------|---------|
| Static | HTML templates with CSS | Existing component library | Hero, services-grid, testimonials |
| Dynamic | Self-contained JS applications | LLM-generated or pre-built | Calculators, data visualisations |
| Application | Full web apps with API integration | Engineer-built only | Mood boards, layout editors |

---

### 7. News and Content Feed Agent Family

**Owner:** `content-feed-orchestrator` (orchestrator)

Keeps sites fresh with industry-relevant content. Sources vary by site type — a boxing site pulls from ESPN, BoxRec, promoter announcements. A finance site from FT, Reuters, FCA. A vet site from Vet Times, BVA, RCVS.

#### Data Model

```sql
-- Source configuration per site
CREATE TABLE content_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    source_type TEXT NOT NULL,            -- 'rss', 'api', 'scrape', 'entity_event'
    source_url TEXT,
    source_config JSONB DEFAULT '{}',     -- API keys, headers, selectors, keyword filters
    category TEXT,                         -- 'industry_news', 'regulatory', 'competitor',
                                          -- 'event_triggered'
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
    external_id TEXT,                      -- dedup key from source
    title TEXT NOT NULL,
    body TEXT,
    source_url TEXT,
    published_at TIMESTAMPTZ,
    raw_data JSONB,                        -- full original item
    status TEXT DEFAULT 'ingested',        -- ingested, duplicate, triaged,
                                           -- approved, rewritten, published, archived
    triage_result JSONB,                   -- relevance score, suggested angle, urgency
    rewritten_content JSONB,               -- rewritten article content
    page_id UUID REFERENCES pages(id),    -- links to published page
    entity_ids UUID[],                     -- related entity references for cross-linking
    lifecycle_status TEXT,                  -- featured, current, aging, archive, pruned
    lifecycle_changed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(site_id, source_id, external_id)
);
```

#### The Pipeline

```
Sources (RSS, API, scrape, entity state changes)
    → Ingestion → Deduplication → Triage → Rewriting → Entity linking → Publication → Lifecycle
```

#### Sub-Agents

| Agent | Role | LLM? |
|-------|------|------|
| `feed-ingester` | Fetch from configured sources, store raw items | No |
| `feed-deduplicator` | Detect duplicate/near-duplicate stories | Minimal |
| `feed-triage` | Assess relevance, urgency, angle for site | Yes |
| `article-rewriter` | Rewrite raw content into original articles in site voice | Yes |
| `feed-publisher` | Create page from rewritten article, deploy | No |
| `feed-lifecycle` | Age, archive, prune old content | No |

#### Lifecycle (time-based decay)

| Age | Status | Behaviour |
|-----|--------|-----------|
| 0-24 hours | `featured` | Prominent placement, may be on homepage |
| 1-7 days | `current` | Category listings, searchable |
| 7-30 days | `aging` | Drops from prominent positions |
| 30-90 days | `archive` | Archive pages only, may be noindexed |
| 90+ days | `prune_candidate` | Remove or retain based on traffic |

Lifecycle timing varies by site type. A news-heavy site: fast (24h featured, 7d current). A brochure site with a blog: slow (7d featured, 30d current). An events site: tied to event calendar — a fight preview becomes stale when the fight is over.

Rewritten articles become `site_entities` with `entity_type = 'article'`, so the entity pipeline handles them like products or events.

#### Connection to Entity Data

Entity state changes can trigger news articles. When the entity-data-agent detects a significant change during sync, it writes a `content_feed_item` with `source_type = 'entity_event'`:

```
entity-data-agent detects: event.status changed 'announced' → 'on_sale'
  → writes content_feed_item:
      source_type: 'entity_event'
      title: "Tickets on sale: Fury vs Joshua"
      body: { event_entity_id, change_type: 'status_change',
              old_status: 'announced', new_status: 'on_sale',
              entity_data: {...} }
  → feed-triage decides: yes, this is newsworthy for this site
  → article-rewriter produces article with entity cross-links
  → feed-publisher creates page, linking to event entity page
```

Not every entity change triggers this. The `entity_sources.news_triggers` config controls which changes are significant:

```json
{
  "news_triggers": {
    "status_changes": ["announced", "on_sale", "sold_out", "past"],
    "price_changes_above_percent": 20,
    "availability_below_threshold": 100
  }
}
```

A routine price change (£99 → £101) produces no article. A fight announcement, tickets going on sale, a sell-out, or results after the event — those are newsworthy.

#### Connection to Maintenance

Feed lifecycle is a maintenance concern. The `feed-lifecycle` agent's findings can feed into `maintenance_findings`:

- "Article X is 45 days old with low traffic → archive candidate" (finding_type: `stale_article`)
- "Article Y references an event that's now past → needs update or redirect" (finding_type: `article_entity_stale`)
- "Article Z links to an event page where tickets are sold out → update CTA" (finding_type: `article_cta_stale`)

The content-discovery-agent (from the maintenance family) can include checks for:
- Entity pages where `entity.last_synced_at` is stale relative to poll_interval
- Entity pages where entity status changed but page hasn't been re-rendered
- News articles referencing entities whose status has since changed

---

### 8. Maintenance Agent Family

**Overview:** Maintenance is fundamentally different from build. During build, work flows in one direction. During maintenance, problems arrive from everywhere — content ages, links break, competitors evolve, regulations update. The maintenance family uses three layers: discovery agents find problems, triage classifies them, fix agents resolve them. All coordinated through a shared `maintenance_findings` table.

**Full detail:** See the Maintenance Architecture Plan document. Summary below.

#### Orchestration Agents

| Agent | Role | Spawned By |
|-------|------|-----------|
| `maintenance-batch-scheduler` | Populates work queue, claims batch, spawns per-site orchestrators | agent-chassis (via K8s CronJob) |
| `site-maintenance-orchestrator` | Per-site: fix → verify → discover → triage cycle | maintenance-batch-scheduler |
| `maintenance-catch-all` | Daily cleanup: stale findings, HITL reminders, cross-site patterns | agent-chassis (via daily CronJob) |

#### Discovery Agents (spawned by site-maintenance-orchestrator)

| Agent | Domain | Method | Phase |
|-------|--------|--------|-------|
| `content-discovery-agent` | content | Date patterns, entity drift, thin content, stale stats | Phase 0 |
| `links-discovery-agent` | links | Internal/external link checks, orphans, redirect chains | Phase 0 |
| `seo-discovery-agent` | seo | Sitemap sync, schema validation, meta freshness | Phase 0 |
| `compliance-discovery-agent` | compliance | Disclaimers, legal templates, regulatory changes, tool compliance | Phase 1 |
| `structural-discovery-agent` | structural | Nav complexity, redundant content, competitor gaps | Phase 3 |

#### Fix Agents (spawned by site-maintenance-orchestrator)

Fix agents are purpose-built for narrow, targeted changes. They are separate from build-phase agents because build agents do "create from nothing" (broad scope, full brief) while fix agents do "change this specific thing because of this specific reason" (narrow scope, finding-driven).

| Agent | Handles | LLM? | Phase |
|-------|---------|------|-------|
| `section-rewriter` | Stale content, entity drift, thin sections, broken external link rewording | Yes | Phase 2 |
| `redirect-manager` | Redirect chains, broken internal links, link removal | No | Phase 1 |
| `sitemap-regenerator` | Sitemap out of sync | No | Phase 1 |
| `nav-updater` | Orphaned nav items, nav restructuring | No | Phase 1 |
| `legal-updater` | Outdated legal templates, missing disclaimers | Yes | Phase 2 |
| `schema-fixer` | Invalid structured data, stale meta descriptions | Minimal | Phase 2 |
| `image-optimiser` | Format conversion, resizing, missing alt text | Minimal | Phase 2 |
| `css-patcher` | CSS variable drift, missing responsive rules | No | Phase 3 |

#### Triage

Triage runs as a step within the site-maintenance-orchestrator, not as a separate agent. It reads `detected` findings, cross-references other domains' tables for impact assessment (read-only), classifies resolution path (`auto_fix`, `suggest`, `flag`, `monitor`, `ignore`), routes to the correct fix agent, and scores priority.

#### Cross-Domain Coordination

Domains are independent. Coordination happens through:
- **Impact reads during triage** — triage reads link_registry, site_nav_items, pages to assess blast radius before classifying
- **Side-effect findings during fix** — when a fix agent changes something that affects another domain, it writes a new finding with `parent_finding_id` pointing to the original

No agent calls another agent for coordination. They communicate through the `maintenance_findings` table.

#### Per-Site Configuration

Stored in `sites.settings` as `maintenance_profile`. Controls which domains and sub-checks run, at what cadence, with budget limits:

```json
{
  "maintenance_profile": {
    "content": { "enabled": true, "every": "7d", "last_run_at": "...", "agents": {...} },
    "links":   { "enabled": true, "every": "8h", "last_run_at": "...", "agents": {...} },
    "seo":     { "enabled": true, "every": "7d", "agents": {...} },
    "compliance": { "enabled": true, "every": "30d", "regulatory_bodies": ["FCA"] },
    "structural": { "enabled": false },
    "budget": { "llm_calls_per_cycle": 20, "max_auto_fixes_per_cycle": 5 }
  }
}
```

```
intake-orchestrator
  ├── site-classifier
  ├── briefing-agent
  └── component-builder-v2          (NEW workflow)
        ├── site-planner             (existing)
        ├── brand-designer           (NEW, or reuse webdesign-agent)
        ├── layout-architect         (NEW)
        ├── nav-agent                (NEW, currently populate_nav action)
        ├── legal-content-agent      (NEW)
        ├── image-generator          (existing)
        ├── style-generator          (from webdesign-agent)
        ├── page-content-writer      (existing)
        ├── content-reviewer         (existing)
        ├── deployer-agent           (existing)
        ├── seo-content-agent        (NEW, post-content)
        └── links-orchestrator       (NEW, post-deploy)

heartbeat-scheduler (K8s CronJob, every 8 hours)
  └── agent-chassis
        └── maintenance-batch-scheduler
              └── site-maintenance-orchestrator (per site, K8s job)
                    ├── content-discovery-agent   (spawned)
                    ├── links-discovery-agent     (spawned)
                    ├── seo-discovery-agent       (spawned)
                    ├── compliance-discovery-agent (spawned)
                    ├── structural-discovery-agent (spawned)
                    ├── section-rewriter          (spawned for fixes)
                    ├── redirect-manager          (spawned for fixes)
                    ├── sitemap-regenerator       (spawned for fixes)
                    └── (other fix agents as needed)

daily-maintenance-cron (K8s CronJob, daily)
  └── agent-chassis
        └── maintenance-catch-all

content-feed-cron (separate schedule)
  └── content-feed-orchestrator      (NEW)
```

### Key Differences from pageflow-builder

1. **Nav is separate**: Nav agent runs before content loop. Rendering queries nav tables directly.
2. **Legal pages handled separately**: Template-based, before main content loop.
3. **Design is split**: Brand spec → layout definitions → CSS, each stored separately.
4. **SEO runs post-content**: Sweep after all pages written.
5. **Links validation post-deploy**: Full link graph validation after deployment.
6. **Layout definitions drive rendering**: Nav group placement per page type, not hardcoded.
7. **Maintenance is a separate spawn chain**: Not part of the build workflow. CronJob → agent-chassis → per-site maintenance orchestrators with their own discovery and fix agents.

---

## Site Type Stress Tests

### Brochure Site (existing, works)
Nav: primary + utility + legal. No entity data. Standard templates. Straightforward internal links.
Optional news/feed for SEO freshness.

### E-commerce / Product Review Site
Nav: primary + content (categories) + legal (affiliate disclosure). Products as entities. Commercial outbound links. Fast feed lifecycle (deals age in days).

### Finance / Tools Site
Nav: primary + tools + content (guides) + legal (extensive). Pervasive disclaimers. LLM-generated calculators or pre-built library. Legal constraints on all rewritten news. Entity types: financial instruments, rates, regulatory notices. Feed sources: FT, Reuters, FCA announcements.

### Events / Tickets Site (first target — boxing, then football)
Nav: primary + content (news) + contextual (per-event, per-performer). Dense entity relationships — events link to performers, venues, ticket tiers. Entity-backed pages with template rendering from structured data. Entity lifecycle is state-based (announced → on_sale → sold_out → past → historical), not time-based. Ticket tier data updates frequently (prices, availability) and must flow through to pages quickly. News feed tied to event calendar — entity state changes (fight announced, tickets on sale, results) trigger news articles via `entity_sources.news_triggers`. Entity types: event, performer, venue, ticket_tier. API sources: Ticketmaster, SeatGeek, BoxRec (boxing), football league APIs.

### Interactive Platform Site
Marketing pages via standard pipeline. Engineer-built Tier 3 components. Agent-as-API pattern. User IS the HITL. Blog/tutorial content with long lifecycle.

---

## Data Ownership Summary

| Data | Owner Agent | Tables |
|------|-------------|--------|
| Site record, brand_assets, content_data | site-planner / brand-designer | `sites` |
| Page records, sections | site-planner | `pages` |
| Navigation structure | nav-agent (currently populate_nav action) | `site_nav_groups`, `site_nav_items` |
| Link registry | links-orchestrator | `link_registry` |
| Redirects | redirect-manager | `site_redirects` |
| Page component HTML | page-content-writer / page-rerender | `page_components` |
| Site-level components | render_site_components action | `site_components` |
| Style collection | style-generator / webdesign-agent | `style_collections` |
| CSS themes | webdesign-agent (future: style-generator) | `css_themes` |
| Content components | (library, manually maintained) | `content_components` |
| Entity data | entity-data-agent | `site_entities`, `site_entity_relationships` |
| Entity source config | entity-data-agent | `entity_sources` |
| Entity sync history | entity-data-agent | `entity_sync_log` |
| Layout definitions | layout-architect | `sites.content_data.layout_definitions` |
| Brand spec | brand-designer | `sites.content_data.brand_spec` |
| Legal rules | legal-content-agent | `sites.content_data.legal_rules` |
| SEO metadata | seo-content-agent | `pages` (meta fields) |
| Content sources | content-feed-orchestrator | `content_sources` |
| Raw feed items | feed-ingester | `content_feed_items` |
| Published articles | article-rewriter / feed-publisher | `site_entities` (type=article) + `pages` |
| Maintenance work queue | maintenance-batch-scheduler | `maintenance_tasks` |
| Maintenance findings | discovery agents (write), triage (enrich), fix agents (update) | `maintenance_findings` |
| Maintenance profile | site owner / system defaults | `sites.settings.maintenance_profile` |

---

## Implementation Phasing

### Phase 1 — Foundation (in progress)

| Agent/Action | Priority | Status | Notes |
|-------|----------|--------|-------|
| `populate_nav_tables` action | High | **Implemented** | Runs in pageflow-builder, classifies pages into groups |
| `GetNavItems()` shared query | High | **Implemented** | Used by header/footer rendering, fallback to pages table |
| Nav tables schema | High | **Deployed** | `site_nav_groups`, `site_nav_items` with status, utility classification |
| Webdesign CSS inheritance fix | High | **Ready** | Prompt updated to enforce colour inheritance model |
| `links-orchestrator` | High | Not started | Link health across all site types |
| `brand-designer` | Medium | Not started | Separate from webdesign-agent |
| `layout-architect` | Medium | Not started | Page-type-specific nav placement |
| `style-generator` + theme library | Medium | Not started | CSS generation with search-first reuse |
| `legal-content-agent` | Medium | Not started | Template-based legal pages + constraint rules |
| `seo-content-agent` | Medium | Not started | Post-content meta generation |

### Phase 1b — Maintenance Foundation

| Item | Priority | Notes |
|------|----------|-------|
| `maintenance_findings` table | High | Coordination point for all maintenance |
| `maintenance_tasks` table | High | Work queue (same pattern as `collection_tasks`) |
| `maintenance-batch-scheduler` agent | High | CronJob trigger → claim batch → spawn per-site orchestrators |
| `site-maintenance-orchestrator` agent | High | Per-site find → triage → fix → verify cycle |
| `content-discovery-agent` (date refs only) | High | Algorithmic — regex scan for time-sensitive patterns |
| `links-discovery-agent` (internal only) | High | Algorithmic — verify internal link targets exist |
| `seo-discovery-agent` (sitemap sync only) | High | Algorithmic — compare sitemap against pages table |
| K8s CronJob manifests | High | 8-hourly maintenance trigger, daily catch-all trigger |
| HITL notifications for findings | High | Using existing HITL message pattern |

### Phase 2a — Entity Data Foundation (first priority after maintenance foundation)

Target site: boxing tickets. Entity data drives pages and ticket information.

| Item | Priority | Notes |
|------|----------|-------|
| `entity_sources` table | High | Source configuration per site, per entity type |
| `entity_sync_log` table | High | Change tracking and audit trail |
| `entity-data-agent` orchestrator | High | Scheduled sync, change detection, page flagging |
| Entity → page generation (template-based) | High | Event pages, performer pages, venue pages from entity data |
| Entity → page re-render on change | High | Price changes, availability changes trigger section re-render |
| Entity relationship management | High | Event ↔ performer, event ↔ venue, event ↔ ticket_tier |
| Entity lifecycle state machine | Medium | announced → on_sale → sold_out → past → historical |
| Boxing entity types | High | event, performer, venue, ticket_tier with field schemas |
| Football entity types | High | match, team, competition, venue, ticket_tier |
| Manual/HITL entity creation | Medium | For initial data entry before API integrations |

Phase 2a is useful immediately — even without API integrations, sites can have manually-entered entity data that generates pages.

### Phase 2b — Feed Pipeline Foundation

| Item | Priority | Notes |
|------|----------|-------|
| `content_sources` table | High | RSS/API/scrape source config per site |
| `content_feed_items` table | High | Raw items + processing state |
| `feed-ingester` agent | High | RSS polling, simple API fetch |
| `feed-triage` agent | High | LLM relevance assessment, angle suggestion |
| `article-rewriter` agent | High | LLM rewriting in site voice with entity cross-links |
| `feed-publisher` agent | High | Page creation + deploy from rewritten articles |
| `feed-lifecycle` agent | Medium | Aging, archiving, pruning |
| `feed-deduplicator` agent | Medium | Near-duplicate detection across sources |

### Phase 2c — API Integrations and Entity Sync

First: boxing ticket APIs. Then football. Then finance data feeds.

| Item | Priority | Notes |
|------|----------|-------|
| Ticketmaster API connector | High | Event discovery, ticket pricing, availability |
| Boxing news feed sources | High | ESPN, BoxRec, promoter RSS/APIs |
| Entity state → news trigger mechanism | High | Significant entity changes write to content_feed_items |
| SeatGeek / StubHub API connector | Medium | Alternative/secondary ticket sources |
| Football ticket API connectors | Medium | Similar pattern to boxing, different entity types |
| Price change tracking and alerting | Medium | Track price history, detect significant changes |
| Finance data feed sources (FT, Reuters, FCA) | Later | Different entity types (instruments, rates, regulatory notices) |

### Phase 2d — Cross-Linking and Intelligence

| Item | Priority | Notes |
|------|----------|-------|
| Entity ↔ article cross-linking | High | Articles reference entities, entity pages list related articles |
| Entity relationship-driven navigation | High | Contextual nav: fighter page lists all their events |
| Entity-driven components | Medium | Upcoming events widget, price comparison table, ticket availability grid |
| Entity-driven SEO | Medium | Structured data (Event schema, Offer schema) from entity fields |
| Maintenance integration | Medium | content-discovery checks for stale entity pages, stale article CTAs |

### Phase 2e — Maintenance Triage and Simple Fixes (parallel with entity/feed work)

| Item | Priority | Notes |
|------|----------|-------|
| Maintenance triage with impact cross-referencing | High | Enriches findings with blast radius data |
| `redirect-manager` fix agent | High | Shorten chains, update broken internal links |
| `sitemap-regenerator` fix agent | High | Rebuild sitemap from pages table |
| `nav-updater` fix agent | High | Orphaned nav items, nav restructuring |
| `maintenance-catch-all` agent | High | Daily cron: stale findings, HITL reminders, cross-site patterns |
| `links-discovery-agent` + external checks, orphan detection | Medium | Extends Phase 1b foundation |
| `compliance-discovery-agent` (disclaimers) | Medium | Algorithmic disclaimer presence checking |

### Phase 3 — LLM-Assisted Maintenance, Fixes, and Advanced Content

| Item | Priority |
|------|----------|
| `section-rewriter` fix agent | High |
| `legal-updater` fix agent | High |
| `product-content-writer` (entity-driven product/review pages) | High |
| `schema-fixer` fix agent | Medium |
| `image-optimiser` fix agent | Medium |
| `content-discovery-agent` + entity drift, stale statistics | Medium |
| `seo-discovery-agent` + meta freshness (LLM) | Medium |
| `compliance-discovery-agent` + legal template versioning | Medium |
| `tool-builder-agent` (calculators, configurators) | Medium |
| `affiliate-link-manager` | Low |

### Phase 4 — Strategic, Competitive, and Platform

| Capability | Priority |
|-----------|----------|
| `structural-discovery-agent` (competitor analysis, nav complexity) | High |
| Adopt/research pipeline | High |
| Agent-as-API layer | High |
| Multi-tenant scoping | High |
| `css-patcher` fix agent | Medium |
| Analytics integration (Google Analytics API) | Medium |
| Dynamic component library | Medium |
| Cross-site pattern learning | Low |
| Real-time collaboration | Low |
| Agent marketplace | Low |

---

## Migration Notes

- The existing `pageflow-builder` workflow is NOT modified (except adding the `populate_nav` step). It continues to work as-is.
- Nav table population is currently an action within pageflow-builder, not a separate agent. It will migrate to a standalone agent when the maintenance/heartbeat model is built.
- The `in_header` / `in_footer` booleans on the `pages` table remain for backward compatibility; nav tables are authoritative when populated, pages table is the fallback.
- The `webdesign-agent` remains available for all workflows; the split into brand-designer + style-generator happens when the theme library justifies it.
- Maintenance is a completely separate spawn chain from build. No build workflows are modified to support maintenance. The maintenance-batch-scheduler, site-maintenance-orchestrator, discovery agents, and fix agents are all new agent definitions.
- The `maintenance_tasks` table follows the same claiming pattern as `business_intel.collection_tasks` (FOR UPDATE SKIP LOCKED). The `maintenance_findings` table is new.
- The links-orchestrator's build-phase responsibilities (link extraction, registry sync) remain separate from the links-discovery-agent's maintenance responsibilities (finding broken links, orphan pages). They share the `link_registry` table but have different read/write patterns.

---

## Resolved Decisions

1. **Nav agent during build sequence**: Runs after planner, before content loop. Currently implemented as `populate_nav` action step.
2. **Heartbeat trigger mechanism**: K8s CronJob publishes Kafka message to agent-chassis, which spawns maintenance-batch-scheduler. Not a persistent agent — follows the vet-batch-processor spawn pattern.
3. **Layout definitions storage**: JSONB on `sites.content_data` under `layout_definitions` key.
4. **Legal rules scope**: Per-site on `sites.content_data.legal_rules`. Templates seed common rules, live rules belong to the site.
5. **Entity data sourcing**: API-first. Manual/HITL for initial seeding. Both write to `site_entities`, `source` field distinguishes origin.
6. **CSS colour inheritance**: Base stylesheet sets `color: var(--color-text)` on `body` only. All other elements use `color: inherit` or omit colour. Components set `color: #fff` on dark containers. Webdesign-agent prompt enforces this.
7. **Theme reuse vs generation**: Currently generates per site, not stored in `css_themes`. Plan: store generated themes, search before generating, build reusable library. Split into brand-designer + style-generator when library is large enough.
8. **Maintenance profile location**: `sites.settings` — simple, already exists, queryable via JSONB operators.
9. **Fix agents vs build agents**: Fix agents are separate, purpose-built for narrow targeted changes. Build agents do "create from nothing" (broad scope). Fix agents do "change this specific thing" (narrow scope, finding-driven). They share underlying actions but have separate workflows.
10. **Maintenance coordination model**: Agents communicate through the `maintenance_findings` table, not by calling each other. Triage reads other domains' tables for impact assessment (read-only). Fix agents write cross-domain side-effect findings.
11. **Discovery agents as spawned K8s jobs**: Separately spawned by the site-maintenance-orchestrator. Cleaner logs, failure isolation, independent scaling.
12. **Site independence in maintenance**: Each site gets its own orchestrator instance. No shared state between site maintenance runs except cross-site pattern detection in the catch-all agent.
13. **Entity lifecycle is state-based, not time-based**: Events go through announced → on_sale → sold_out → past → historical based on real-world status, not age. News articles decay by age. Both patterns coexist on the same site.
14. **Entity state changes trigger news via feed pipeline**: Significant entity changes (configured per source via `news_triggers`) write to `content_feed_items`. The feed-triage agent decides whether to produce an article. Not every change is newsworthy.
15. **First entity/ticket target**: Boxing tickets site, then football tickets. Same entity pattern (event, performer/team, venue, ticket_tier) with different field schemas per sport. Finance follows later with different entity types.
16. **Entity sources and content sources are separate tables**: Similar shape but different ownership. `entity_sources` owned by entity-data-agent, `content_sources` owned by content-feed-orchestrator. Keeps responsibilities clear.


