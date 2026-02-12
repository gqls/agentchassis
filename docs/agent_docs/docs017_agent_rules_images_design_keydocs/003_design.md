# Webdesign Agent Implementation

## Overview

This implementation adds CSS generation capabilities to the site building system through a new specialist agent (`webdesign-agent`) that can be called standalone or integrated into builder workflows.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    BUILDER WORKFLOWS                             │
│       (pageflow-builder, landing-page-builder, etc.)            │
│                                                                  │
│  Simple call: { "site_id": "uuid" }                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SPECIALIST AGENTS                             │
├─────────────────────────────────────────────────────────────────┤
│  webdesign-agent                                                 │
│  ┌─────────────────┐                                            │
│  │ load_site_for_  │ ← Loads own context from DB                │
│  │ design          │   (site, pages, components, colors)        │
│  └────────┬────────┘                                            │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ analyze_design  │ ← LLM generates design spec                │
│  │ (LLM prompt)    │   (colors, typography, spacing)            │
│  └────────┬────────┘                                            │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ generate_css    │ ← LLM creates production CSS               │
│  │ (LLM prompt)    │                                            │
│  └────────┬────────┘                                            │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ deploy_css      │ ← git_commit with file_path config         │
│  │ (git_commit)    │   → /assets/css/styles.css                 │
│  └─────────────────┘                                            │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  site-scraper (companion agent)                                  │
│  ┌─────────────────┐                                            │
│  │ firecrawl_scrape│ ← Uses existing webscrape adapter          │
│  └────────┬────────┘                                            │
│           ▼                                                      │
│  ┌─────────────────┐                                            │
│  │ analyze_design  │ ← LLM extracts colors, fonts, components   │
│  │ (LLM prompt)    │                                            │
│  └────────┬────────┘                                            │
│           ▼                                                      │
│  │ Returns site_context for webdesign-agent                     │
└─────────────────────────────────────────────────────────────────┘
```

## Design Principles

1. **Specialist agents own their data gathering** - Each agent loads what it needs via dedicated actions, keeping builder workflows simple.

2. **Reuse existing code** - Small patches to existing actions (`git_commit`, `WebscrapeAction`) rather than creating duplicates.

3. **Standardized interfaces** - `site_context` schema allows different data sources (DB, scrape, manual) to feed the same design logic.

4. **Standalone capability** - Agents can be called independently for maintenance, testing, or ad-hoc operations.

## Files

### Patches to Existing Code

| File | Target | Change |
|------|--------|--------|
| `patch_01_git_commit_file_path.go` | `git_deployer_actions.go` | Adds `file_path` config for non-HTML files (CSS, JS, etc.) |
| `patch_02_webscrape_url_field.go` | `webscrape_actions.go` | Adds `url_field` config for flexible URL extraction |

### New Action

| File | Purpose |
|------|---------|
| `02_load_site_for_design_action.go` | Loads comprehensive site context from database including pages, components, style collections |

### Agent Definitions

| File | Agent Type | Purpose |
|------|------------|---------|
| `01_webdesign_agent.sql` | `webdesign-agent` | Generates and deploys CSS stylesheets |
| `04_site_scraper_agent.sql` | `site-scraper` | Scrapes live websites to extract design context |

### Integration & Fixes

| File | Purpose |
|------|---------|
| `05_integrate_webdesign_pageflow.sql` | Adds spawn + call steps to pageflow-builder |
| `06_add_stylesheet_link_to_head.sql` | Adds `<link rel="stylesheet">` to head components |
| `fix_hero_css.sql` | Immediate inline CSS fix for hero components |

## Deployment Order

### 1. Go Code Changes

```bash
# Apply patches to existing files
# patch_01 → git_deployer_actions.go (add file_path support)
# patch_02 → webscrape_actions.go (add url_field support)

# Add new action
# 02_load_site_for_design_action.go → platform/orchestration/actions/

# Register action in action_registry.go:
# "load_site_for_design": LoadSiteForDesignAction,

# Build and deploy
make build
# Deploy to kubernetes
```

### 2. SQL Patches (after Go deploy)

```bash
# Create agents
psql -f 01_webdesign_agent.sql
psql -f 04_site_scraper_agent.sql

# Integrate with pageflow-builder
psql -f 05_integrate_webdesign_pageflow.sql

# Update head components
psql -f 06_add_stylesheet_link_to_head.sql

# Immediate hero fix (optional, run first for quick win)
psql -f fix_hero_css.sql
```

## Usage

### Standalone webdesign-agent

```json
{
  "action": "process",
  "agent_type": "webdesign-agent",
  "data": {
    "site_id": "4851f6fc-71cf-4160-a270-e03d6d3e0732"
  }
}
```

Or with domain:

```json
{
  "action": "process",
  "agent_type": "webdesign-agent",
  "data": {
    "domain": "example.com"
  }
}
```

### Standalone site-scraper

```json
{
  "action": "process",
  "agent_type": "site-scraper",
  "data": {
    "url": "https://example.com"
  }
}
```

### Scrape → Design Pipeline

```json
{
  "action": "process",
  "agent_type": "site-scraper",
  "data": {
    "url": "https://competitor.com"
  }
}
// Returns site_context

{
  "action": "process",
  "agent_type": "webdesign-agent",
  "data": {
    "site_id": "your-site-uuid",
    "site_context": { /* from scraper */ }
  }
}
```

### Automatic (via pageflow-builder)

When building a site, the webdesign-agent runs automatically after pages are built:

```
... → build_pages_loop → apply_site_design → trigger_site_deploy → ...
```

## site_context Schema

Standardized format for design data from any source:

```json
{
  "site_id": "uuid (if from DB)",
  "domain": "example.com",
  "company_name": "Example Corp",
  "tagline": "Making examples simple",
  "industry": "technology",
  
  "color_palette": {
    "primary": "#1a1a2e",
    "secondary": "#16213e",
    "accent": "#0f3460",
    "background": "#ffffff",
    "text": "#333333"
  },
  
  "typography": {
    "font_family": "-apple-system, sans-serif",
    "heading_font": "inherit",
    "base_size": "16px",
    "line_height": "1.6"
  },
  
  "pages": [
    {"title": "Home", "slug": "/", "component_functions": ["hero", "services-grid"]},
    {"title": "About", "slug": "/about", "component_functions": ["about-hero", "team"]}
  ],
  
  "all_component_functions": ["hero", "services-grid", "about-hero", "team", "cta"],
  
  "source": "database|scrape|manual",
  "source_url": "https://... (if scraped)"
}
```

## Generated CSS Structure

The LLM generates CSS with:

1. **CSS Variables** (`:root`) - Colors, typography, spacing from design spec
2. **Minimal Reset** - box-sizing, margin/padding normalization
3. **Typography** - body, h1-h6, p, a, lists
4. **Layout** - `.container`, `.section`, grid utilities
5. **Buttons** - `.btn`, `.btn-primary`, `.btn-secondary`, `.btn-large`
6. **Components** - Styles for each component function used on the site
7. **Responsive** - Mobile-first with 768px and 1024px breakpoints
8. **Accessibility** - Focus states, reduced motion support
9. **Print** - Print-friendly styles

Output location: `/assets/css/styles.css`

## Future Specialists

This pattern can be extended for other specialist agents:

| Agent | Action | Output |
|-------|--------|--------|
| `link-manager` | `load_site_for_links` | Updated navigation across all pages |
| `seo-agent` | `load_site_for_seo` | Meta tags, structured data, sitemap |
| `image-optimizer` | `load_site_for_images` | Optimized images, WebP conversion |
| `accessibility-agent` | `load_site_for_a11y` | ARIA labels, contrast fixes |

Each agent:
- Loads its own context via dedicated action
- Performs focused analysis/generation
- Deploys changes to git
- Can be called standalone or from builders