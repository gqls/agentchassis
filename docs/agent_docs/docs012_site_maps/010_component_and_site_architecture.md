# https://claude.ai/chat/1ac4f28a-bee4-4bca-a9eb-aa3f0ca041a2

Component-Based Website Architecture

## Overview

This document describes the schema and agent architecture for building websites entirely from reusable components, with LLM assistance only where needed.

## Core Principles

1. **Components are the unit of reuse** - Headers, footers, sections, even entire page layouts are components
2. **Database is source of truth** - Content lives in the database; Git is the deployment artifact
3. **LLM = Agent** - Every LLM call gets its own agent with a workflow (research, draft, review)
4. **Deploy per page** - Build and deploy one page at a time to keep messages small
5. **Research is cited** - All LLM-generated content must cite sources, which are stored
6. **Assets are tracked** - Images (generated, uploaded, scraped) have full provenance

## Schema Overview

### Existing Tables (Extended)

| Table | Extensions | Purpose |
|-------|------------|---------|
| `sites` | content_data, brand_assets, default_components, deploy_config | Site-level config and content store |
| `pages` | build_status, deploy_commit | Track page build progress |
| `page_components` | build_status, reviewed_at, reviewed_by, research_id | Track section build + review status |
| `content_components` | render_mode, agent_type, data_sources | How to render each component |
| `link_registry` | affiliate_product_id, requires_disclosure | Affiliate link tracking |

### New Tables

| Table | Purpose |
|-------|---------|
| `research_results` | Store research findings with sources for any content |
| `assets` | Track all images/videos with full provenance |
| `products` | Product catalog for e-commerce/product sites |
| `product_assets` | Link products to their images |
| `affiliate_programs` | Affiliate network configurations |
| `affiliate_products` | Cached + custom affiliate product content |

## Data Flow

```
Brief Data (from questionnaire)
    ↓
sites.content_data = {
    company_name, tagline, services[], 
    target_audience, tone, contact_email...
}
    ↓
site-planner agent decides:
    - Which pages to create
    - Which components per page
    - Which style_collection to use
    ↓
pages + page_components records created
    ↓
For each page (sequential):
    ↓
    page-content-writer agent:
        - For template sections: render with data
        - For agent sections: research → write → cite sources
    ↓
    content-reviewer agent:
        - HITL or auto-eval
        - Edits saved to page_components
    ↓
    deploy_single_page action:
        - Assemble full HTML
        - Extract links → link_registry
        - Commit to git
        - Update page.deploy_commit
    ↓
Final: trigger Cloudflare deploy
```

## Component Rendering Modes

| render_mode | Behavior |
|-------------|----------|
| `template` | Direct template rendering with data substitution |
| `agent` | Spawn agent to generate content (uses agent_type field) |
| `composite` | Contains child components, assembled in order |

## Agent Architecture

```
intake-orchestrator
    └── multipage-website-builder (orchestrator)
            │
            │   Spawns upfront:
            ├── site-planner
            ├── image-generator
            ├── page-content-writer
            ├── content-reviewer
            └── site-deployer
            
            page-content-writer may spawn:
            └── research-agent (for sections needing research)
            └── section-writer (for LLM content)
```

### Agent Responsibilities

| Agent | Responsibility |
|-------|---------------|
| `site-planner` | Analyze brief, decide pages/components, select style |
| `image-generator` | Generate logo, hero images, OG images |
| `page-content-writer` | Render or generate each section of a page |
| `research-agent` | Web search, fetch sources, synthesize with citations |
| `section-writer` | Write content for a single section with research |
| `content-reviewer` | HITL or auto-eval review of content |
| `site-deployer` | Git commit and Cloudflare deployment |

## Content Storage

### sites.content_data
Stores all questionnaire/brief answers:
```json
{
    "company_name": "Leopardess Consulting",
    "tagline": "Sharp Insight. Quiet Strength.",
    "about_us": "...",
    "services": [
        {"name": "Digital Transformation", "description": "..."}
    ],
    "target_audience": "C-suite leaders...",
    "tone": "professional yet approachable",
    "contact_email": "hello@example.com"
}
```

### sites.brand_assets
References to generated/uploaded assets:
```json
{
    "logo": {
        "primary": {"asset_id": "uuid", "url": "..."},
        "dark_bg": {"asset_id": "uuid", "url": "..."}
    },
    "favicon": {"asset_id": "uuid", "url": "..."},
    "og_image": {"asset_id": "uuid", "url": "..."}
}
```

### sites.default_components
Default header/footer/head for all pages:
```json
{
    "head": "head-seo-standard",
    "header": "header-professional-dark",
    "footer": "footer-4-column"
}
```

### page_components.content_data
Section-specific content after rendering:
```json
{
    "headline": "Transform Your Business",
    "body": "We help companies...",
    "cta_text": "Get Started",
    "sources": ["research-uuid"]
}
```

## Research & Citations

All LLM-generated content must cite sources:

1. `research-agent` performs web search
2. Fetches and extracts relevant quotes from top results
3. Synthesizes findings with numbered citations
4. Stores in `research_results` table with full source list
5. `page_components.research_id` links content to its research
6. Optionally display citations on page (`sources_displayed` flag)

## Asset Provenance

Every asset tracks its origin:

| origin_type | Description |
|-------------|-------------|
| `generated` | AI-generated (stores prompt) |
| `uploaded` | User uploaded |
| `scraped` | Pulled from external URL |
| `stock` | Stock photo service |
| `affiliate` | From affiliate network |
| `derived` | Modified from another asset (links to origin) |

Alterations are tracked:
```json
[
    {"type": "resize", "params": {"width": 1200}, "at": "2024-01-15T..."},
    {"type": "background_remove", "at": "2024-01-15T..."}
]
```

## Affiliate Products

For sites with affiliate content:

1. `affiliate_programs` - Configure each network (Amazon, Awin, etc.)
2. `affiliate_products` - Cache product data with our custom overrides
3. `link_registry.requires_disclosure` - Flag links needing disclosure
4. Scheduled refresh via `affiliate-sync-agent` (future)

## Build Status Tracking

### Page Level (pages.build_status)
- `pending` - Not yet built
- `planning` - Components being selected
- `building` - Content being generated
- `reviewing` - In HITL or eval review
- `approved` - Ready to deploy
- `deployed` - Live on site
- `failed` - Build failed

### Section Level (page_components.build_status)
- `pending` - Not yet rendered
- `writing` - Agent generating content
- `reviewing` - In review
- `approved` - Content approved
- `deployed` - Committed to git

## Future Additions (Phase 2+)

### Page Editor Workflow
- Fetch existing page from DB
- Accept edit instructions
- LLM or direct edits
- Review changes (diff view)
- Save and redeploy

### Product Content Writer
- Generate product descriptions
- Create comparison tables
- Write reviews with affiliate disclosures

### Affiliate Sync Agent
- Pull latest from affiliate networks
- Update cached content
- Flag price changes
- Refresh stale images

## Migration Files

| File | Purpose |
|------|---------|
| `041_component_architecture_schema.sql` | All schema changes |
| `042_seed_default_components.sql` | Seed head/header/footer components |

## Related Agent Definitions

| File | Agent |
|------|-------|
| `017_multipage_website_builder.sql` | Updated workflow with spawn-first |
| `043_site_planner.sql` | New agent definition |
| `044_page_content_writer.sql` | New agent definition |
| `045_research_agent.sql` | New agent definition |
| `046_content_reviewer.sql` | New agent definition |