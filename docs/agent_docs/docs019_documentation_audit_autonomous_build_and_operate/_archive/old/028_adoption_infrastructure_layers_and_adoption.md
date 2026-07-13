# 026 — Infrastructure Layers, Site Adoption, and Backend Capabilities

How the platform serves sites, adopts existing sites, and scales to support backend functionality — from static brochure sites to full agent framework deployments.

---

## Infrastructure Separation

The platform operates in three distinct layers. Layer 1 never sees external traffic. Layer 2 serves the public. Layer 3 builds other platforms.

### Layer 1 — Core Platform (existing K8s cluster)

The factory. Builds websites, runs agent orchestration, collects data, generates artifacts.

- Agent chassis, Kafka, Postgres
- Content generation, design, improvement loops
- Data collection (vet crawls, medicine prices, research)
- Produces: HTML, CSS, JS, JSON data, config files, Terraform, Docker images
- Internal traffic only — no public-facing endpoints

### Layer 2 — Client Delivery Infrastructure (separate)

The showroom. Serves websites, handles client API endpoints, runs client backends.

- Static site serving via S3/Backblaze (exists now)
- Site API router for form handling, data queries (planned)
- Client-facing Postgres for site data (planned)
- OVH VMs or dedicated servers, not the core cluster
- Provisioned and managed by Layer 1

### Layer 3 — Framework Builder (future)

Building other factories. Deploys agent frameworks and AI pipelines for clients.

- Provisions infrastructure for client-specific deployments
- Deploys known frameworks (LlamaIndex, LangGraph, custom)
- Each client gets isolated infrastructure
- Managed, monitored, and maintained by the platform

### Data Flow Between Layers

```
Core cluster (Layer 1)              Client server (Layer 2)
┌─────────────────────┐            ┌──────────────────────┐
│ Agent orchestration  │            │ nginx                │
│ Content generation   │──git──────→│ S3 (static files)    │
│ Data collection      │──db sync──→│ Postgres (site data) │
│ Terraform/Ansible    │──deploy───→│ site-api-router      │
│ Monitoring agents    │←─metrics──│ Prometheus            │
└─────────────────────┘            └──────────────────────┘
```

The core cluster pushes. The client server serves. Monitoring flows back. No inbound traffic from the internet to the core cluster.

---

## Backend Capability Tiers

Not every site needs a server. Most don't. The system classifies sites by what backend they actually require.

### Tier 1 — Static + client-side JS

Brochure sites, portfolios, landing pages. Everything runs in the browser.

- Deployment: S3 + CDN
- Monthly cost: ~£0
- Examples: mortgagecalculator.co.uk (content pages), ai-agent-orchestration.com
- Status: fully supported now

### Tier 2 — Static + edge functions / API router

Contact forms, newsletter signup, quote requests, simple data submission.

- Deployment: S3 + site-api-router on client server
- Monthly cost: ~£5-15 (shared server allocation)
- Examples: any brochure site with a contact form
- Status: planned — site-api-router needed

### Tier 3 — Application with persistent state

User accounts, saved data, content management, booking systems, small e-commerce.

- Deployment: S3 + site-api-router + Postgres + auth service
- Monthly cost: ~£25-50
- Examples: small SaaS, booking platform, membership site
- Status: planned — requires auth integration and DB provisioning

### Tier 4 — Application with real-time features

Chat, live updates, collaborative editing, notifications. Needs persistent connections.

- Deployment: dedicated service with WebSocket support
- Monthly cost: ~£50-100
- Examples: community site, live chat, collaborative tools
- Status: future — requires dedicated process management

### Tier 5 — Full platform

Complex business logic, message queues, orchestration, background processing.

- Deployment: K8s namespace or dedicated cluster
- Monthly cost: ~£200+
- Examples: marketplace, agent platform deployment, multi-service application
- Status: future — requires per-client infrastructure provisioning

### Tier Classification in the Pipeline

The classifier's `hosting_trajectory` field determines the tier:

```json
{
  "hosting_trajectory": "static_only | static_plus_api | needs_persistent_state | needs_realtime | needs_server",
  "backend_features": ["contact_form", "user_auth", "search_index", "data_api"]
}
```

Work item handlers adapt based on tier. A contact form on a Tier 2 site generates an API route config. The same contact form concept on a Tier 1 site generates a mailto: link or third-party embed.

---

## The vetcomparison.uk Pattern

Worth examining because it demonstrates how much can be done without a server.

### Current architecture (no backend)

```
Data pipeline:  crawl → process → export JSON → git commit → S3
                (runs on core cluster, pushes artifacts)

Client-side:    browser loads vet-full-index.json (~2.5MB, 5000 vets)
                JS does search, filter, sort, calculate — all in-browser
                medicine data lazy-loaded by letter chunk
                no API calls, no server queries
```

### What this covers

- Full-text search over 5,000 vets by postcode or name
- Price filtering with range slider
- Sort by price ascending/descending
- Independent vs corporate filter
- Medicine cost comparison calculator
- Error reporting via mailto:

### Where this pattern has limits

| Need | JSON-on-S3 works? | Alternative |
|------|-------------------|-------------|
| Search over <50k items | Yes | — |
| Search over >50k items | No — too much data for browser | Pre-built search index (Pagefind) or API |
| Form submissions | No | Site-api-router POST endpoint |
| User accounts | No | Auth service + DB |
| Data changing hourly | No — rebuild too slow | API endpoint or WebSocket |
| Heavy computation | Depends | Server-side or WebAssembly |

### Data pipeline as work items

The vet data collection already runs as crawl jobs on the core cluster. The export step (DB → JSON → git → S3) maps to a scheduled work item:

```
needs_data_export | handler: data-export-agent
  spec: {
    site_id: "...",
    queries: [
      {name: "vet-full-index", sql: "SELECT ... FROM vet_practices", format: "json"},
      {name: "vets_by_postcode/{area}", sql: "SELECT ... WHERE area = $1", partition_by: "area"},
      {name: "medicine-index", sql: "SELECT id, name FROM medicines", format: "json"},
      {name: "medicines_by_letter/{letter}", sql: "SELECT ... WHERE letter = $1", partition_by: "letter"}
    ],
    output_path: "data/",
    commit_message: "Update vet and medicine data"
  }
```

The handler queries the DB, generates JSON files, commits to git. GitHub Actions pushes to S3. The site is updated with no server restart, no deploy, no downtime.

---

## Site API Router (Layer 2 Backend)

For sites that need POST endpoints or server-side data queries. Runs on the client-facing server, not the core cluster.

### Design: Option 4 — Shared function router

A single Go service that loads route configurations from the DB. Each "function" is a declarative action chain, similar to agent workflows but synchronous and HTTP-triggered.

```go
// site_api_routes table:
// site_id | method | path       | handler_config
// uuid    | POST   | /contact   | {"actions": ["validate_input", "store_submission", "send_email"]}
// uuid    | GET    | /api/vets  | {"actions": ["query_database"], "query": "SELECT ... FROM vets WHERE area = $1"}

func (s *SiteAPIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route := s.matchRoute(r.Host, r.Method, r.URL.Path)
    if route == nil {
        http.NotFound(w, r)
        return
    }
    result, err := s.executeActions(r.Context(), route.HandlerConfig, r.Body)
    // ...
}
```

### Why this approach

- No new infrastructure — reuses existing action library
- Configuration-driven — add routes via DB, no code deploy
- Scales to many sites on one service
- The core cluster generates and pushes route configs as part of site builds
- Easy to reason about: each route is a short action chain

### Where route configs live

In `site_specs` as an `api_routes` aspect:

```json
{
  "routes": [
    {
      "method": "POST",
      "path": "/api/contact",
      "actions": ["validate_input", "store_submission", "send_email"],
      "config": {
        "table": "form_submissions",
        "notify": "owner@example.com",
        "rate_limit": "5/minute"
      }
    }
  ]
}
```

### Incremental build path

1. Start with a few hardcoded routes in the existing API gateway (contact form handler)
2. Extract into configurable route system when the pattern solidifies
3. Deploy as a standalone service on the client server
4. The core cluster manages the route configs as part of site specs

---

## Minimum Viable Layer 2

A single OVH VM serving client sites with API capabilities. Provisioned using the existing Terraform pattern.

### Components

```
OVH VM (dedicated server or VPS)
  ├── nginx (reverse proxy, SSL via certbot, rate limiting)
  │     → static sites: proxy to S3/Backblaze
  │     → site APIs: proxy to site-api-router
  ├── site-api-router (Go binary, systemd service)
  │     → loads route configs from Postgres
  │     → executes action chains
  ├── Postgres (site data: form submissions, entity data, route configs)
  ├── Prometheus + node-exporter (monitoring)
  ├── Grafana (dashboards)
  ├── fail2ban (security)
  └── cron jobs
        → data exports
        → search index rebuilds
        → cert renewal
```

### Provisioning

Extends the existing Terraform pattern (OVH base setup + nginx reverse proxy + certbot). Additional provisioners for:

- site-api-router binary deployment
- Postgres installation and schema migration
- Prometheus scrape config for the router service
- nginx upstream config for API routes

### Management

The core cluster manages the client server through:

- **Deployment**: build Go binary → rsync to server → systemctl restart
- **Config updates**: push route configs to Postgres → router hot-reloads
- **Monitoring**: Prometheus metrics scraped back to core Grafana
- **Health checks**: periodic HTTP probes, alert on failure
- **Backups**: pg_dump → S3 (same pattern as core DB backups)

Future: these management tasks become work items processed by infrastructure agents.

---

## Site Adoption

### The core principle

Adoption is just a richer set of initial specs. The pipeline doesn't change — the planner plans, the dispatch loop dispatches, the handlers handle. What changes is that specs contain `existing_content` and `adopt_from` data instead of being generated from scratch.

### Adoption stages

**Stage 1a — Static content only (mortgagecalculator.co.uk without tools)**

Crawl the site, extract text and structure, recreate pages with our component system.

```
seed: {adopt_from: "mortgagecalculator.co.uk", mode: "content_only"}
  → site-adoption-agent
      1. crawl_site (batch_webscrape — all pages)
      2. extract_site_identity (LLM: company name, tagline, industry from homepage)
      3. extract_design (LLM: parse CSS → palette, typography, spacing)
      4. classify_pages (LLM: map each page to page_types, identify sections)
      5. extract_page_content (per page: section text, headings, images)
      6. write_specs (identity, design, visual_direction, structure, adoption_source)
      7. write_work_items:
           needs_design (with extracted palette)
           needs_content_page × N (with mode: "recreate", existing_content per section)
           needs_rerender
```

Handlers don't know they're adopting. A content writer receiving `mode: "recreate"` with `existing_content` uses that content as source material instead of generating from scratch. A webdesign-agent receiving `adopt_from.color_palette` matches those colours.

**Stage 1b — Add interactive tools**

Extract JS/CSS for self-contained client-side tools. Classify each as self-contained (browser-only) vs API-dependent.

```
Additional work items:
  needs_tool_page × N | spec: {mode: "recreate", tool_js: "...", tool_css: "..."}
```

Self-contained tools (calculators, comparisons, filters) become tool components — `<section>` with embedded JS. The existing tool pipeline handles deployment.

**Stage 2 — Entity/directory sites (vetcomparison.uk)**

Crawl entity listings, extract structured data, populate the entity system.

```
Additional steps in adoption:
  8. extract_entities (LLM: identify entity types, extract structured data per entity)
  9. write_entity_items:
       needs_entity_setup (entity types, sources, relationships)
       needs_data_export (JSON generation for client-side search)
       needs_tool_page (search/filter UI)
```

The entity data flows through `site_entities` → export action → JSON → git → S3. The client-side JS loads the JSON and does search/filter in-browser.

**Stage 3 — Sites with form handling / API needs**

Generate API route configurations for the site-api-router.

```
Additional work items:
  needs_api_route × N | spec: {method: "POST", path: "/api/contact", actions: [...]}
  needs_infrastructure_update (push route config to client server)
```

The handler generates the route config, writes it to `site_specs.api_routes`, and triggers a config push to the client server.

**Stage 4 — Sites with auth / persistent state**

Configure third-party or self-hosted auth and database.

```
Additional work items:
  needs_service_config | spec: {auth: "self_hosted", db: "postgres"}
  needs_db_schema | spec: {tables: [{name: "users", ...}, {name: "favourites", ...}]}
  needs_api_route × N (CRUD endpoints)
```

### Adoption agent classification

The adoption agent's first step classifies what it found:

```json
{
  "pages": [
    {"url": "/", "type": "content", "sections": ["hero", "search", "calculator", "guides"]},
    {"url": "/guides/independent-strategy.html", "type": "content", "sections": ["article"]}
  ],
  "interactive_features": [
    {"name": "vet-search", "type": "client_side_tool", "data_source": "vet-full-index.json", "self_contained": true},
    {"name": "medicine-calc", "type": "client_side_tool", "data_source": "medicine-index.json", "self_contained": true},
    {"name": "contact-form", "type": "form_submission", "self_contained": false, "needs": "api_endpoint"}
  ],
  "data_dependencies": [
    {"path": "/data/vet-full-index.json", "size": "~2.5MB", "type": "entity_index"},
    {"path": "/data/medicine-index.json", "size": "~50KB", "type": "product_index"},
    {"path": "/data/medicines_by_letter/{A-Z}.json", "type": "chunked_product_data"}
  ],
  "tier": 1,
  "hosting_trajectory": "static_only"
}
```

Features classified as `self_contained: false` generate work items for the appropriate tier. Features classified as `self_contained: true` generate tool component work items.

---

## Large Dataset Search (Pagefind Pattern)

For sites with >50k items where client-side JSON loading is impractical, pre-built search indexes replace the full data load.

Pagefind is a static search library — builds an index at deploy time, ships ~100KB of WASM + index chunks to the browser, does full-text search with no server.

Integration as a build step:

```
needs_search_index | handler: search-index-builder
  spec: {
    site_id: "...",
    source: "pages",  // or "entities" or "json_data"
    index_fields: ["title", "description", "content"],
    facet_fields: ["category", "location"]
  }
```

The handler:
1. Generates HTML or JSON source files from the data
2. Runs Pagefind CLI to build the index
3. Commits index files to git
4. Client-side JS uses the Pagefind API for search

This keeps the "no server" property for read-heavy sites up to ~500k items.

---

## Framework Builder (Layer 3)

For clients who need their own AI/agent infrastructure rather than just a website.

### Tiered approach

**90% of clients**: Static site + API router (Layer 2). Shared infrastructure. No dedicated deployment.

**8% of clients**: Dedicated VM with a deployed framework. Single service (LlamaIndex RAG, LangGraph workflow, custom chatbot). The platform provisions, deploys, and monitors.

**2% of clients**: Full orchestration deployment. K8s namespace or dedicated cluster with agent framework.

### Deploying known frameworks

A client requests "AI chatbot for customer service." The platform:

1. Classifies the need (RAG chatbot, workflow automation, data pipeline)
2. Selects framework (LlamaIndex for RAG, LangGraph for workflows)
3. Generates application config (prompts, data sources, tool definitions)
4. Provisions infrastructure (VM via Terraform)
5. Deploys framework with config (Docker Compose)
6. Configures monitoring, backups, SSL
7. Provides management dashboard

Work items:

```
needs_requirements_analysis    → classify what the client needs
needs_framework_selection      → pick framework
needs_infrastructure           → Terraform: provision VM, install Docker
needs_application_config       → generate framework config, prompts, tools
needs_deployment               → docker compose up, nginx config, SSL
needs_monitoring               → Prometheus endpoint, health checks
needs_data_ingestion           → load client documents into RAG pipeline
needs_testing                  → smoke test the deployment
```

### Deploying cut-down agent framework

For clients who need multi-agent orchestration:

```
Client's infrastructure:
  ├── agent-chassis (single pod or Docker container)
  ├── Postgres (their data)
  ├── Redis Streams or small Kafka (messaging)
  ├── nginx (API gateway)
  └── Their agent definitions + workflows
```

The platform generates the Docker Compose or K8s manifests, provisions the server, deploys, and monitors. Updates are pushed from the core cluster.

### Self-managing infrastructure

The servers themselves become entities in the system:

```
site_entities:
  entity_type: "server"
  state: "provisioning" → "active" → "draining" → "terminated"
  data: {provider: "ovh", ip: "...", specs: "...", services: [...]}
```

Management as work items:

```
needs_server_provision    → Terraform action → creates VM
needs_software_install    → Ansible/shell action via SSH
needs_service_deploy      → Docker compose or K8s manifest
needs_health_check        → periodic HTTP probes (discovery check pattern)
needs_backup              → scheduled pg_dump → S3
needs_scaling             → Terraform action to add/remove resources
needs_security_update     → apt upgrade, restart services
```

The existing entity lifecycle (state-based transitions) and discovery check patterns (periodic scans that create work items) apply directly.

---

## Phase Plan

### Phase 1 — Adoption pipeline (content only)

- Site-adoption-agent: crawl → extract → classify → write specs → write work items
- Content writer `mode: "recreate"` support
- Test on mortgagecalculator.co.uk content pages (no tools)
- Design extraction (CSS → palette/typography)

### Phase 2 — Tool and entity adoption

- Tool extraction: identify self-contained JS, create tool components
- Entity extraction: structured data from directory/listing pages
- Data export action: DB → JSON → git → S3
- Test on mortgagecalculator.co.uk (with calculators) and vetcomparison.uk

### Phase 3 — Client-facing backend (Layer 2)

- Site-api-router Go service
- Route configuration in site_specs
- OVH VM provisioning with existing Terraform pattern
- Form handling (contact, submissions, error reports)
- Test with vetcomparison.uk error reporting

### Phase 4 — Persistent state and auth

- Auth service integration (self-hosted or third-party)
- DB provisioning for client data
- CRUD endpoint generation
- User accounts, saved data

### Phase 5 — Framework deployments (Layer 3)

- Infrastructure provisioning agents (Terraform actions)
- Framework deployment handlers (LlamaIndex, LangGraph)
- Server entity management (health checks, backups, updates)
- Management dashboard per client

### Phase 6 — Self-managing infrastructure

- Infrastructure agents manage Layer 2 and Layer 3 servers
- Auto-scaling based on traffic metrics
- Security patching as scheduled work items
- Disaster recovery: detect failure → provision replacement → restore from backup

---

## Principles

1. **Layer 1 never serves external traffic.** It produces artifacts and pushes them. No inbound connections from the internet.

2. **No vendor lock-in for core capabilities.** Use third-party services (Cloudflare, S3, OVH) but always with the exit path designed. The site-api-router is our own code. Infrastructure provisioning supports multiple providers via Terraform.

3. **Static-first.** Default to the lowest tier that serves the need. The JSON-on-S3 pattern (vetcomparison.uk) handles more cases than people expect. Only add backend when the site genuinely requires it.

4. **Same pipeline, richer specs.** Adoption, new builds, maintenance, and framework deployment all flow through the work item pipeline. What changes is what writes the items and what handler agents exist. The orchestration layer is the same.

5. **Servers are entities.** Infrastructure follows the same patterns as site data — state-based lifecycle, discovery checks for health, work items for management tasks, specs for configuration.

6. **The framework is a framework builder.** The same system that generates websites can generate infrastructure configs, framework deployments, and management automation. The LLM agents generate Terraform the same way they generate HTML.
