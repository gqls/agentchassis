# 002 — System Architecture

Complete reference for the agent orchestration system. Covers the design system, all agent families, the unified build/maintenance pipeline, and implementation phases.

---

## Design System Layers

The design system has three independent layers that vary separately.

### Layer 1: HTML Components (structure and layout)

**Table:** `content_components`

Self-contained HTML blocks — a hero section, a testimonial grid, a FAQ accordion, a services card layout. Each has its own inline `<style>` for layout (grid, flexbox, spacing) and dark section overrides.

Multiple components exist for the same function (e.g. `hero-split`, `hero-fullwidth`, `hero-minimal`). What varies is structure. Components reference CSS variables with fallbacks: `var(--color-primary, #1a1a2e)`. They never hardcode brand colours. Dark sections set `color: #fff` on their container element, and all children inherit light text automatically.

### Layer 2: CSS Theme (appearance)

**Table:** `css_themes`

A complete base stylesheet setting `:root` variables (colours, fonts, spacing), base resets, typography scale, button styles, responsive breakpoints, and accessibility focus states.

Deployed as `/assets/css/styles.css` — one per site, committed to git by the webdesign agent.

The colour inheritance model (see 003_contracts_and_standards for full rules): `body` sets `color: var(--color-text)` as the base default. Text elements reference `--section-*` variables with light-theme fallbacks (e.g. `h1-h6 { color: var(--section-heading, var(--color-primary)); }`). Dark-section components override `--section-*` on their container and all children adapt automatically.

### Layer 3: Style Collection (the bridge)

**Table:** `style_collections`

A named grouping that ties together header component + footer component + CSS theme + colour palette. Think of it as a "design kit."

| Field | References | Purpose |
|---|---|---|
| `header_component_id` | `content_components` | Which header HTML to use |
| `header_home_component_id` | `content_components` | Alternate header for homepage (optional) |
| `footer_component_id` | `content_components` | Which footer HTML to use |
| `css_theme_id` | `css_themes` | Which CSS theme to use |
| `color_palette` | JSONB | Primary, secondary, accent, background, text |
| `typography` | JSONB | Font family, base size, line height, heading font |

### How they connect

```
site (leopardessconsulting.co.uk)
  └── style_collection_id → style_collections (professional-dark)
        ├── header_component_id → content_components (header-professional-dark)
        ├── footer_component_id → content_components (footer-4-column)
        ├── css_theme_id → css_themes (→ /assets/css/styles.css)
        └── color_palette → {primary: "#1a1a2e", accent: "#0f3460", ...}
  └── pages (from site plan)
        ├── index.html → hero, differentiators, testimonials, cta components
        ├── about.html → different body components
        └── services.html → different body components
  All pages share same header, footer, CSS, and :root variables.
```

### Theme library growth

1. Build site → webdesign agent generates CSS → stored as new `css_themes` row
2. Tag with industry, style category, colour characteristics
3. Next similar brief → search existing themes → reuse if match, adjust `:root` values
4. Webdesign agent goes from "always generate" to "search → maybe reuse → maybe generate → always store"

---

## Current System

### Active agents

| Agent | Type | Role |
|---|---|---|
| `intake-orchestrator` | orchestrator | Entry point — classifies, briefs, spawns builder |
| `site-classifier` | specialist | Classifies site type from domain/objective |
| `briefing-agent` | specialist | Runs questionnaire, collects brief data |
| `site-planner` | specialist | Plans pages, selects components and style |
| `pageflow-builder` | orchestrator | Builds sites: plan → assets → content → deploy |
| `page-content-writer` | specialist | Writes content per page, section by section |
| `content-reviewer` | specialist | HITL or auto-eval content review |
| `research-agent` | specialist | Web research for content backing |
| `image-generator` | specialist | Generates logos and hero images |
| `asset-deployer` | specialist | Downloads S3 image, optimizes, commits to git |
| `webdesign-agent` | specialist | Generates design spec and CSS |
| `deployer-agent` | specialist | Git commit and Cloudflare deployment |
| `page-rerender` | specialist | Re-assembles single page from stored components |
| `rerender-pages` | orchestrator | Batch rerender across all site pages |
| `section-editor` | specialist | Granular edits to individual page sections |

### Active workflows

| Workflow | Entry Point | Function |
|---|---|---|
| `intake-orchestrator` | User submits domain + objective | Full pipeline: classify → brief → build |
| `pageflow-builder` | Spawned by intake | Plan → generate assets → build pages → deploy |
| `rerender-pages` | Manual or post-build | Re-assemble pages from stored components |

### Implemented actions (not yet separate agents)

| Action | Runs In | Function |
|---|---|---|
| `populate_nav_tables` | pageflow-builder | Classifies pages into nav groups |
| `GetNavItems()` | component_library.go | Shared query — reads nav tables with pages-table fallback |

### Infrastructure

- **Auth Service**: JWT-based authentication, user management, project scoping
- **API Gateway**: Gin-based HTTP, proxies to core manager
- **Kafka**: Inter-agent messaging via request/response topics
- **PostgreSQL**: Sites, pages, content_components, page_components, style_collections, css_themes, site_nav_groups, site_nav_items, assets, link_registry, orchestration_states, site_entities, site_entity_relationships
- **Kubernetes**: ai-persona-system namespace, Docker images, Terraform/Kustomize
- **Deployment**: Git commit → GitHub Actions → Backblaze S3. Cloudflare is DNS only.

note:
broker pods are personae-kafka-cluster-combined-pool-prod-0
kubectl -n kafka get pods
NAME                                                      READY   STATUS    RESTARTS       AGE
personae-kafka-cluster-combined-pool-prod-0               1/1     Running   0              23h
personae-kafka-cluster-combined-pool-prod-1               1/1     Running   0              23h
personae-kafka-cluster-combined-pool-prod-2               1/1     Running   0              23h
personae-kafka-cluster-entity-operator-5dfd87f6f4-7kpv4   2/2     Running   12 (23h ago)   23h

---

## Unified Build and Maintenance

Build and maintenance are the same process. A new site is a set of work items that need doing. An existing site generates work items through discovery. Both go through the same queue, processed by the same orchestrator, handled by the same agents.

### Coexistence with existing pipeline

The existing `pageflow-builder`, `site-classifier`, `site-planner`, and `intake-orchestrator` continue to work unchanged. The new system uses separate agent definitions:

| Existing (keeps working) | New (parallel system) |
|---|---|
| `intake-orchestrator` | `intake-orchestrator-v2` |
| `site-classifier` | `site-classifier-v2` |
| `site-planner` | `site-planner-v2` |
| `pageflow-builder` | `site-work-orchestrator` |
| `maintenance-triage` + `page-rebuild` | `maintenance-batch-scheduler` + `site-work-orchestrator` |

### site_work_items table

Every piece of work — building a new page, fixing stale content, adding a tool, publishing a news article — is a work item.

Key fields:
- `source`: 'planner', 'discovery', 'content_feed', 'manual', 'improvement', 'side_effect'
- `domain`: 'build', 'content', 'links', 'seo', 'compliance', 'structural', 'design', 'navigation', 'entity', 'tools'
- `item_type`: 'needs_content_page', 'needs_tool_page', 'stale_date_reference', 'broken_link', etc.
- `severity`: 'info', 'low', 'medium', 'high', 'urgent'
- `spec`: JSONB — work-specific structured data (page_spec, tool_spec, finding detail)
- `handler_agent`: which agent processes this
- `status`: 'detected', 'triaged', 'approved', 'claimed', 'in_progress', 'complete', 'pending_verify', 'verified', 'failed', 'rejected', 'wont_fix'
- `priority`: computed from severity + impact. Lower = higher priority.
- `depends_on`: UUID array — items that must complete first
- `item_key`: deterministic key for deduplication, UNIQUE per site
- `result`: JSONB — includes commit_sha for git tracking

### The site lifecycle

```
Phase 1: Initial Build
  1. Classify + brief (human involved)
  2. Planner writes work items (source: 'planner')
  3. Site work orchestrator processes items → handler agents build pages
  4. Individual git commits per page → each goes live via GitHub Action → S3
  5. Basic site is live

Phase 2+: Incremental Improvement
  6. Planner or improvement agent writes more items
  7. Orchestrator processes → site grows
  8. Each batch leaves site in working state

Ongoing: Maintenance
  9. Discovery agents scan → write items (source: 'discovery')
  10. Triage enriches items
  11. Fix agents process → site stays fresh
  12. Feed orchestrator writes article items (source: 'content_feed')
  13. Article writers publish → site gains content
  14. Goto 9
```

Steps 3, 7, and 11 are the same code.

### The orchestrator

**site-work-orchestrator** replaces both the build orchestrator and the site-maintenance-orchestrator:

```
Workflow:
  1. load_site_context        → site record, maintenance profile, work items summary
  2. process_pending_items    → items WHERE status IN ('triaged', 'approved')
                                AND dependencies satisfied, ORDER BY priority
                              → for each: route to handler_agent → do work → git commit → mark complete
  3. verify_previous_work     → items with status = 'pending_verify' → re-check → mark verified/failed
  4. run_discovery            → for each domain in maintenance profile that is due:
                                spawn discovery agent → collect findings
                                (skipped on first build — nothing to discover on empty site)
  5. triage_new_items         → items with status = 'detected' → cross-reference impact →
                                classify resolution_path → assign handler → score priority → update to 'triaged'
  6. update_last_run          → update maintenance profile timestamps
  7. complete
```

Order is deliberate: process existing work first, verify previous work, then discover new work, then triage.

### Dispatch pattern: spawn→call per work item

The orchestrator processes work items in a loop. Each iteration spawns the handler agent dynamically and calls it with raw identifiers. The handler loads its own context.

```
fix_items_loop:
  for each work item:
    spawn_handler   → role: "fix_handler", agent_type_field: "current_fix_item.handler_agent"
    call_handler    → target_role: "fix_handler", input_mapping: { site_id, domain, spec fields? }
    mark_complete   → complete_work_item
```

The agent type comes from the work item's `handler_agent` field. The spawn resolves it dynamically via `agent_type_field`. The call finds the just-spawned agent via `target_role: "fix_handler"`. Standard spawn→call, same as everywhere else in the system.

**What the orchestrator passes:**

```json
{
    "site_id": "site_record.site_id",
    "domain": "site_record.domain",
    "asset_id?": "current_fix_item.spec.asset_id",
    "purpose?": "current_fix_item.spec.purpose",
    "check?": "current_fix_item.spec.check",
    "page_id?": "current_fix_item.page_id"
}
```

Optional fields (`?` suffix) are silently skipped if absent in the work item spec. This means the same dispatch loop handles `missing_css` items (webdesign-agent needs `site_id`, `domain`), `undeployed_asset` items (asset-deployer needs `asset_id`, `purpose`), and future handler types — without the orchestrator knowing what each handler needs.

**What the orchestrator does NOT do:**
- Pre-spawn handler agents in a static chain. Agent types aren't known until work items are loaded.
- Build derived data for handlers (e.g. resolving s3:// URIs from asset IDs). That's the handler's job.
- Pass work item IDs or work item awareness to handlers. Handlers don't know about the work item system.

**Handlers are self-contained.** The webdesign-agent receives `site_id` and `domain`, does `check_site_context → load_site_for_design → generate_css → deploy`. The asset-deployer receives `asset_id`, `purpose`, `domain`, resolves the storage URI itself from the assets/sites tables. Both can be called directly from CLI with the same inputs — no orchestrator required.

Status tracking (marking items in_progress/complete) is the orchestrator's concern, handled in the loop before and after the handler call.

### Batch scheduler

```
K8s CronJob (every 8 hours)
  └── agent-chassis → maintenance-batch-scheduler
        1. populate_work_queue  → query sites with due maintenance OR pending build items
        2. claim_batch          → FOR UPDATE SKIP LOCKED, batch_size sites
        3. process_batch        → spawn site-work-orchestrator per site
        4. complete
```

For initial builds, the intake-orchestrator triggers the site-work-orchestrator directly.

### The planner as discovery agent

The planner writes work items to DB instead of returning JSON:

```
For a simple brochure site:
  needs_content_page  | index    | priority 10 | handler: page-content-writer
  needs_content_page  | about    | priority 10 | handler: page-content-writer
  needs_content_page  | services | priority 10 | handler: page-content-writer
  needs_logo          | site     | priority 5  | handler: image-generator
  needs_design        | site     | priority 8  | handler: webdesign-agent

For a mixed site (wykefarm):
  -- Tier 1: core content pages (priority 10-12)
  -- Tier 1: site infrastructure (priority 5-8)
  -- Tier 2: tools (priority 30, built after core site is live)
  -- Tier 3: entity directory (priority 50+, depends_on entity setup)
```

Priority numbers control ordering. The orchestrator processes lowest-priority-number-first, respecting dependencies.

### The classifier (v2 output)

```json
{
  "build_approach": "page-assembled | application | hybrid",
  "primary_archetype": "descriptive label",
  "page_types": ["content", "tool", "commerce", "entity-directory", "news"],
  "entity_types": [],
  "hosting_trajectory": "static_only | static_now_api_later | needs_server",
  "detected_industry": "industry name",
  "recommended_builder": "pageflow-builder"
}
```

HITL reviews and can adjust build approach, page types, entity types, industry detection.

### Git commit strategy

Individual commits per work item. Each commit is atomic and reversible. Commit message includes what changed, why (source + item_type), and the work item ID.

**Commit IS deploy.** GitHub Actions fires on each commit, writes changed files to S3 (Backblaze B2). No separate deploy step, no batching, no gap between commit and live.

### Post-processing (the assembly step)

After any handler produces HTML for a page:

```
handler returns page HTML
  → assemble_page (inject head, header, footer from stored site chrome)
  → git_commit (individual commit with item reference)
  → save_page_sections (store rendered sections in page_components)
  → update work item status → 'complete'
  → update page build_status → 'deployed'
```

Same existing actions for all handlers. For site-wide items (design, nav), different post-processing path.

---

## Agent Families

### 1. Navigation Agent Family

Owner: `nav-agent` (currently `populate_nav_tables` action within pageflow-builder).

Navigation is a first-class entity. Tables: `site_nav_groups` (semantic categories), `site_nav_items` (individual nav entries).

**Group types:** primary, subsection, content, legal, utility, external. Contextual groups (per-entity, per-category) planned for entity-driven sites.

**Three-tier authority model:**
- **Tier 1 — Strategist Authority:** New builds and major restructure. Planner plans full nav, nav agent validates and persists.
- **Tier 2 — Nav Agent Authority:** Maintenance, minor additions. Autonomous incremental decisions.
- **Tier 3 — Drift Detection:** Periodic comparison of current nav against original plan.

### 2. Links Agent Family

Owner: `links-orchestrator` (algorithmic, no LLM).

Sub-agents: link-crawler, link-validator, link-registry-sync, redirect-manager, affiliate-link-manager (phase 2).

**Does:** Extract links from HTML, classify types, resolve internals, HTTP HEAD checks, detect broken/orphaned, generate redirects, track link counts. **Does not:** Decide where to place links (content writer), navigate structure (nav agent), SEO strategy (SEO agent).

### 3. Design Agent Family

Maps to the three design system layers.

| Agent | Layer | Responsibility | LLM? | Status |
|---|---|---|---|---|
| `webdesign-agent` | 2 (theme) | Analyses brand, generates design spec, produces CSS | Yes | Exists |
| `brand-designer` | 2 (theme) | Colour, typography, spacing, visual tone | Yes | Future split |
| `layout-architect` | 1 (components) | Page type skeletons, nav group placements | Yes | Planned |
| `style-generator` | 2 (theme) | CSS production from brand spec + layout | Yes | Future split |

Current: webdesign-agent handles both brand analysis and CSS generation. The split happens when the theme library is large enough that "search and adapt" beats "generate from scratch."

### 4. Content Agent Family

Split by what they write because different content types require different approaches.

| Agent | Responsibility | LLM? | Status |
|---|---|---|---|
| `page-content-writer` | Marketing/editorial page content | Yes | Exists |
| `legal-content-agent` | Privacy, terms, disclaimers | Template + minimal LLM | Planned |
| `seo-content-agent` | Meta titles, descriptions, structured data | LLM + algorithmic | Planned |
| `product-content-writer` | Product reviews from structured data | Yes | Phase 2 |
| `research-agent` | Web research for content backing | Yes | Exists |
| `content-reviewer` | HITL or auto-eval review | Yes | Exists |

### 5. Entity Data Agent Family

Owner: `entity-data-agent`. Manages structured data that generates pages — products, events, people, venues, ticket tiers.

**Tables:** `site_entities`, `site_entity_relationships` (exist), `entity_sources`, `entity_sync_log` (planned).

**Two modes:**
- **Setup mode (work items):** Configure source → fetch initial data → render pages → build directory. Dependency chain.
- **Discovery mode (scheduled):** Check sources for changes → write work items for drifted/new/removed entities.

**Entity lifecycle is state-based, not time-based:** Events go through announced → on_sale → selling_fast → sold_out → event_day → past → historical based on real-world status. Different from news which decays by age.

**Entity types for boxing/events:** event, performer, venue, ticket_tier with relationships (features, held_at, has_tickets, on_same_card).

**Real-time data** (prices, availability, live scores) is NOT handled through the work queue. Entity pages include client-side JS that fetches from a data API at view time.

### 6. News and Content Feed Agent Family

Owner: `content-feed-orchestrator`.

**Tables:** `content_sources` (config per site), `content_feed_items` (raw ingested items).

**Pipeline:** Sources → Ingestion → Deduplication → Triage → Rewriting → Entity linking → Publication → Lifecycle.

| Agent | Role | LLM? |
|---|---|---|
| `feed-ingester` | Fetch from configured sources | No |
| `feed-deduplicator` | Near-duplicate detection | Minimal |
| `feed-triage` | Relevance, urgency, angle | Yes |
| `article-rewriter` | Rewrite in site voice with entity cross-links | Yes |
| `feed-publisher` | Create page, deploy | No |
| `feed-lifecycle` | Age, archive, prune | No |

**Lifecycle timing varies:** News site: 24h featured, 7d current. Blog: 7d featured, 30d current. Events: tied to event calendar.

**Connection to entities:** Entity state changes can trigger news articles via `entity_sources.news_triggers`. Not every change — only configured significant changes (status transitions, large price changes, low availability).

**News display is a design concern, not a feed concern.** The feed pipeline handles content supply. Visual presentation (card layouts, article templates) is part of the design system.

### 7. Tool Builder Agent (phase 2)

For interactive components — calculators, configurators, simple tools.

| Tier | Description | Creation |
|---|---|---|
| Static | HTML templates with CSS | Existing component library |
| Dynamic | Self-contained JS applications | LLM-generated or pre-built |
| Application | Full web apps with API | Engineer-built only |

### 8. Maintenance Agent Family

**Discovery agents** find problems. **Triage** classifies them. **Fix agents** resolve them. All coordinated through `site_work_items` table.

**Discovery agents:**

| Agent | Domain | Checks | Schedule |
|---|---|---|---|
| `content-discovery-agent` | content | date refs, entity drift, thin content, stale stats | weekly |
| `links-discovery-agent` | links | internal/external links, orphans, redirects | 8 hours |
| `seo-discovery-agent` | seo | sitemap sync, schema, meta freshness | weekly |
| `compliance-discovery-agent` | compliance | disclaimers, legal templates | monthly |
| `structural-discovery-agent` | structural | nav complexity, redundant content, competitors | monthly |

Discovery agents write items. They do not fix anything. They do not call other agents.

**Fix agents (handler agents):**

Called via the dispatch loop (spawn→call per work item). Each receives raw identifiers, loads its own context. No work-item-specific code in handlers.

Existing agents adapted as handlers:

| Agent | Handles | Notes |
|---|---|---|
| `page-content-writer` | `needs_content_page` | Receives page spec from item.spec |
| `image-generator` | `needs_logo`, `needs_hero_image` | Receives prompt from item.spec |
| `asset-deployer` | `undeployed_asset` | Downloads S3, optimizes, commits to git |
| `webdesign-agent` | `needs_design`, `missing_css` | Generates/updates CSS theme |

New fix agents:

| Agent | Handles | Phase |
|---|---|---|
| `section-rewriter` | stale_date_reference, entity_data_drift | Phase 2 |
| `redirect-manager` | redirect_chain, broken_internal_link | Phase 1 |
| `sitemap-regenerator` | sitemap_out_of_sync | Phase 1 |
| `nav-updater` | nav_item_orphaned | Phase 1 |
| `tool-builder` | needs_tool_page | Phase 2 |
| `article-writer` | publish_article | Feed Phase 1 |

**Triage** runs as a step within the site-work-orchestrator. For build items from the planner, triage is pre-done — items arrive as 'triaged'.

**Cross-domain coordination:** Agents communicate through the work items table, not by calling each other. When a fix creates a side-effect, it writes a new work item with `source: 'side_effect'` and `parent_item_id`.

---

## Section Editor

Performs granular edits to individual page sections without re-running the full content generation pipeline.

### Source-of-truth principle

Every section has two representations: `content_data` (structured JSON) and `rendered_html` (final HTML). **content_data is always the source of truth.** Every edit updates content_data first, then re-renders the template. Edits survive all future re-renders.

### Edit types

**content_edit:** Two modes — `field_updates` (merge into existing content_data) or full `content_data` replacement.

**component_swap:** Replace the component template while keeping existing content_data.

### buildRenderContextFromDB

Key function that constructs `RenderContext` entirely from database state (no collected_data needed):
1. `loadSiteDataFull()` → company name, domain, email, phone, logo
2. `GetStyleCollectionForSite()` → colour palette
3. `getThemeByID()` → CSS theme
4. `GetNavItems()` → header and footer navigation
5. Page metadata → title, description
6. Section content_data → merged as RenderContext.ContentData

---

## JavaScript Management

### JS Snippets Library (`js_snippets` table)

Pre-built, reusable JS for standard interactivity (mobile nav, smooth scroll, form validation). Selected during planning, injected via head component. Site-level, not per-page.

### Custom JS for Tools

Tool pages generate custom JS as part of their build. Self-contained, lives inline in page HTML or as page-specific script file. Managed through `page_components` content_data.

### JS and Components

Component templates declare JS dependencies. Assembly step ensures required JS is included when a component is used on a page.

---

## Per-Site Configuration

Stored in `sites.settings.maintenance_profile`:

```json
{
  "maintenance_profile": {
    "build": { "current_tier": 2, "auto_advance_tiers": true },
    "content": { "enabled": true, "every": "7d", "auto_fix_enabled": false },
    "links": { "enabled": true, "every": "8h" },
    "seo": { "enabled": true, "every": "7d" },
    "compliance": { "enabled": true, "every": "30d" },
    "content_feed": { "enabled": false, "every": "24h", "max_articles_per_cycle": 3 },
    "entity": { "enabled": false, "types": [], "sources": [] },
    "budget": { "llm_calls_per_cycle": 20, "max_auto_fixes_per_cycle": 5 }
  }
}
```

---

## Site Type Stress Tests

**Brochure Site** — primary + utility + legal nav. No entity data. Standard templates. Optional news for SEO freshness.

**E-commerce / Product Review** — categories in nav. Products as entities. Commercial outbound links. Fast feed lifecycle.

**Finance / Tools** — tools + content + extensive legal nav. Pervasive disclaimers. LLM-generated calculators. Legal constraints on rewritten news. Feed sources: FT, Reuters, FCA.

**Events / Tickets (first target: boxing, then football)** — primary + content (news) + contextual (per-event, per-performer) nav. Dense entity relationships. State-based lifecycle. Entity changes trigger news via `news_triggers`. API sources: Ticketmaster, SeatGeek, BoxRec.

**Interactive Platform** — Marketing pages via standard pipeline. Engineer-built Tier 3 components. Agent-as-API pattern. User IS the HITL.

---

## Data Ownership

| Data | Owner Agent | Tables |
|---|---|---|
| Site record, brand_assets, content_data | site-planner / brand-designer | `sites` |
| Page records, sections | site-planner | `pages` |
| Navigation structure | nav-agent | `site_nav_groups`, `site_nav_items` |
| Link registry | links-orchestrator | `link_registry` |
| Redirects | redirect-manager | `site_redirects` |
| Page component HTML | page-content-writer / page-rerender | `page_components` |
| Site-level components | render_site_components action | `site_components` |
| Style collection | webdesign-agent | `style_collections` |
| Content components | library (manually maintained) | `content_components` |
| Entity data | entity-data-agent | `site_entities`, `site_entity_relationships` |
| Layout definitions | layout-architect | `sites.content_data.layout_definitions` |
| Brand spec | brand-designer | `sites.content_data.brand_spec` |
| Legal rules | legal-content-agent | `sites.content_data.legal_rules` |
| SEO metadata | seo-content-agent | `pages` meta fields |
| Content sources | content-feed-orchestrator | `content_sources` |
| Raw feed items | feed-ingester | `content_feed_items` |

---

## Design Consistency (how fix agents don't break sites)

1. **Site chrome is stored in DB** via `page_components` site-level slots. No agent regenerates chrome unless explicitly told to.
2. **CSS theme is a single file** (`/assets/css/styles.css`). Components use CSS variables. Changing content doesn't change styling.
3. **Section reassembly** loads ALL sections for the page, replaces only the changed section, re-assembles with current chrome.
4. **Design changes are explicit work items.** No other agent modifies CSS.
5. **The render pipeline** is shared infrastructure. Build and maintenance use the same functions.

---

## Implementation Phases

### Phase 0 — Foundation

1. `site_work_items` table
2. `content_feed_items` table
3. Planner action: write work items to DB
4. `site-work-orchestrator` agent
5. Intake orchestrator: updated flow
6. Existing `page-content-writer` adapted for work item input

### Phase 1 — Maintenance Discovery + Simple Fixes

7. `maintenance-batch-scheduler`
8. Content, links, SEO discovery agents
9. Triage step in orchestrator
10. redirect-manager, sitemap-regenerator, nav-updater fix agents
11. maintenance-catch-all (daily cron)
12. K8s CronJob manifests

### Phase 2 — Tools + LLM Fixes

13. tool-builder, directory-builder agents
14. section-rewriter, legal-updater, schema-fixer, image-optimiser
15. LLM-based discovery checks
16. compliance-discovery-agent

### Phase 3 — Entities + Feeds

17. entity-data-agent (setup + sync)
18. entity-page-builder
19. content-feed-orchestrator and sub-agents
20. Entity source integrations (boxing APIs → football → finance)

### Phase 4 — Application Sites + Advanced

21. app-builder agent
22. structural-discovery-agent
23. css-patcher, analytics integration
24. Agent-as-API layer, multi-tenant scoping

---

## Resolved Decisions

1. Nav agent during build: runs after planner, before content loop. Currently `populate_nav` action.
2. Heartbeat trigger: K8s CronJob → agent-chassis → spawns maintenance-batch-scheduler.
3. Layout definitions: JSONB on `sites.content_data` under `layout_definitions`.
4. Legal rules: per-site on `sites.content_data.legal_rules`. Templates seed common rules.
5. Entity data sourcing: API-first. Manual/HITL for initial seeding. Both write to `site_entities`.
6. CSS colour inheritance: base stylesheet sets `color: var(--color-text)` on `body`. Text elements use `--section-*` variables with light-theme fallbacks. Dark-section components override `--section-*` on their container.
7. Theme reuse vs generation: currently generates per site. Plan: store, search before generating.
8. Maintenance profile: `sites.settings`.
9. Fix vs build agents: separate. Build = create from nothing (broad). Fix = change specific thing (narrow, finding-driven). Share underlying actions.
10. Discovery agents as spawned K8s jobs: cleaner logs, failure isolation.
11. Site independence: each site gets its own orchestrator instance.
12. Entity lifecycle: state-based, not time-based.
13. Entity state changes trigger news via feed pipeline: `news_triggers` config controls significance.
14. Entity sources and content sources: separate tables, separate ownership.
15. Commit IS deploy: GitHub Actions fires on each commit → S3 → live. No separate deploy step needed in new system.
16. Dispatch loop uses standard spawn→call: `spawn_agent` supports `agent_type_field` for dynamic type resolution. Combined with a fixed `role` and `target_role` lookup, this gives dynamic dispatch with no special Go code. Do not bypass spawn with direct topic construction — agents must be spawned to get proper orchestration tracking, topic setup, and DB registration.
17. Handlers don't know about work items: The orchestrator maps spec fields to handler input_data via `input_mapping`. Handlers receive raw identifiers (`site_id`, `domain`, `asset_id`, etc.) and load their own context. The work item system is the orchestrator's concern, not the handler's.