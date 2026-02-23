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
- **PostgreSQL**: Sites, pages, content_components, page_components, style_collections, css_themes, site_nav_groups, site_nav_items, assets, link_registry, orchestration_states
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

### 5. Entity Data Agent

**Owner:** `entity-data-agent` (orchestrator)

Manages structured data that generates pages. Products, events, people, venues, tools — any real-world entity that a site presents.

Three of four tested site types need it (e-commerce, events/tickets, design platform). Only brochure sites work purely from creative briefs.

#### Data Model

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

#### The Pipeline

```
Sources (RSS, API, scrape, LLM-generated)
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

#### Lifecycle

| Age | Status | Behaviour |
|-----|--------|-----------|
| 0-24 hours | `featured` | Prominent placement, may be on homepage |
| 1-7 days | `current` | Category listings, searchable |
| 7-30 days | `aging` | Drops from prominent positions |
| 30-90 days | `archive` | Archive pages only, may be noindexed |
| 90+ days | `prune_candidate` | Remove or retain based on traffic |

Rewritten articles become `site_entities` with `entity_type = 'article'`, so the entity pipeline handles them like products or events.

---

## New Workflow: Component-Based Builder v2

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

heartbeat-scheduler (periodic, 2-3x daily)
  ├── content-feed-orchestrator      (NEW)
  ├── links-orchestrator             (link health sweep)
  ├── seo-content-agent              (meta content audit)
  ├── nav-agent                      (consistency check, drift detection)
  └── (future: marketing, analytics, etc.)
```

### Key Differences from pageflow-builder

1. **Nav is separate**: Nav agent runs before content loop. Rendering queries nav tables directly.
2. **Legal pages handled separately**: Template-based, before main content loop.
3. **Design is split**: Brand spec → layout definitions → CSS, each stored separately.
4. **SEO runs post-content**: Sweep after all pages written.
5. **Links validation post-deploy**: Full link graph validation after deployment.
6. **Layout definitions drive rendering**: Nav group placement per page type, not hardcoded.

---

## Site Type Stress Tests

### Brochure Site (existing, works)
Nav: primary + utility + legal. No entity data. Standard templates. Straightforward internal links.
Optional news/feed for SEO freshness.

### E-commerce / Product Review Site
Nav: primary + content (categories) + legal (affiliate disclosure). Products as entities. Commercial outbound links. Fast feed lifecycle (deals age in days).

### Finance / Tools Site
Nav: primary + tools + content (guides) + legal (extensive). Pervasive disclaimers. LLM-generated calculators or pre-built library. Legal constraints on all rewritten news.

### Events / Tickets Site
Nav: primary + content (news) + contextual (per-event, per-fighter). Dense entity relationships. Entity-backed pages. Feed lifecycle tied to event calendar.

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
| Layout definitions | layout-architect | `sites.content_data.layout_definitions` |
| Brand spec | brand-designer | `sites.content_data.brand_spec` |
| Legal rules | legal-content-agent | `sites.content_data.legal_rules` |
| SEO metadata | seo-content-agent | `pages` (meta fields) |
| Content sources | content-feed-orchestrator | `content_sources` |
| Raw feed items | feed-ingester | `content_feed_items` |
| Published articles | article-rewriter / feed-publisher | `site_entities` (type=article) + `pages` |

---

## Implementation Phasing

### Phase 1 — Foundation (in progress)

| Agent/Action | Priority | Status | Notes |
|-------|----------|--------|-------|
| `populate_nav_tables` action | High | **Implemented** | Runs in pageflow-builder, classifies pages into groups |
| `GetNavItems()` shared query | High | **Implemented** | Used by header/footer rendering, fallback to pages table |
| Nav tables schema | High | **Deployed** | `site_nav_groups`, `site_nav_items` with status, utility classification |
| Webdesign CSS inheritance fix | High | **Ready** | Prompt updated to enforce colour inheritance model |
| `links-orchestrator` | High | Not started | Link health across all site types, heartbeat-ready |
| `heartbeat-scheduler` | High | Not started | Required for maintenance model |
| `brand-designer` | Medium | Not started | Separate from webdesign-agent |
| `layout-architect` | Medium | Not started | Page-type-specific nav placement |
| `style-generator` + theme library | Medium | Not started | CSS generation with search-first reuse |
| `legal-content-agent` | Medium | Not started | Template-based legal pages + constraint rules |
| `seo-content-agent` | Medium | Not started | Post-content meta generation |

### Phase 2 — Content and Data

| Agent | Priority |
|-------|----------|
| `content-feed-orchestrator` | High |
| `entity-data-agent` | High |
| `product-content-writer` | Medium |
| `tool-builder-agent` | Medium |
| `affiliate-link-manager` | Low |

### Phase 3 — Platform

| Capability | Priority |
|-----------|----------|
| Agent-as-API layer | High |
| Multi-tenant scoping | High |
| Dynamic component library | Medium |
| Real-time collaboration | Low |
| Agent marketplace | Low |

---

## Migration Notes

- The existing `pageflow-builder` workflow is NOT modified (except adding the `populate_nav` step). It continues to work as-is.
- Nav table population is currently an action within pageflow-builder, not a separate agent. It will migrate to a standalone agent when the maintenance/heartbeat model is built.
- The `in_header` / `in_footer` booleans on the `pages` table remain for backward compatibility; nav tables are authoritative when populated, pages table is the fallback.
- The `webdesign-agent` remains available for all workflows; the split into brand-designer + style-generator happens when the theme library justifies it.

---

## Resolved Decisions

1. **Nav agent during build sequence**: Runs after planner, before content loop. Currently implemented as `populate_nav` action step.
2. **Heartbeat trigger mechanism**: Dedicated `heartbeat-scheduler` agent, observable via standard logging/monitoring.
3. **Layout definitions storage**: JSONB on `sites.content_data` under `layout_definitions` key.
4. **Legal rules scope**: Per-site on `sites.content_data.legal_rules`. Templates seed common rules, live rules belong to the site.
5. **Entity data sourcing**: API-first. Manual/HITL for initial seeding. Both write to `site_entities`, `source` field distinguishes origin.
6. **CSS colour inheritance**: Base stylesheet sets `color: var(--color-text)` on `body` only. All other elements use `color: inherit` or omit colour. Components set `color: #fff` on dark containers. Webdesign-agent prompt enforces this.
7. **Theme reuse vs generation**: Currently generates per site, not stored in `css_themes`. Plan: store generated themes, search before generating, build reusable library. Split into brand-designer + style-generator when library is large enough.