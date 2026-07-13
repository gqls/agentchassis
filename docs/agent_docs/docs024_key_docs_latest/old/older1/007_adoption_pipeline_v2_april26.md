# 007 — Adoption Pipeline

How the platform serves sites, adopts existing sites, handles human direction at any stage, and scales to support backend functionality.

---

## Infrastructure Separation

The platform operates in three distinct layers. Layer 1 never sees external traffic.

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

---

## Backend Capability Tiers

Not every site needs a server. Most don't.

### Tier 1 — Static + client-side JS

Brochure sites, portfolios, calculators, comparison tools. Everything runs in the browser. Data served as JSON from S3.

- Deployment: S3 + CDN
- Monthly cost: ~£0
- Status: fully supported now

### Tier 2 — Static + API router

Contact forms, newsletter signup, quote requests, error reports, simple data submissions.

- Deployment: S3 + site-api-router on client server
- Monthly cost: ~£5-15 (shared server allocation)
- Status: planned

### Tier 3 — Application with persistent state

User accounts, saved data, booking systems, small e-commerce.

- Deployment: S3 + site-api-router + Postgres + auth service
- Monthly cost: ~£25-50
- Status: planned

### Tier 4 — Application with real-time features

Chat, live updates, collaborative editing, notifications.

- Deployment: dedicated service with WebSocket support
- Monthly cost: ~£50-100
- Status: future

### Tier 5 — Full platform

Complex business logic, message queues, orchestration, background processing.

- Deployment: K8s namespace or dedicated cluster
- Monthly cost: ~£200+
- Status: future

### The vetcomparison.uk pattern (Tier 1 reference)

Demonstrates how much can be done without a server:

```
Data pipeline:  crawl → process → export JSON → git commit → S3
Client-side:    browser loads JSON index, does search/filter/sort/calculate
No API calls, no server queries, no backend
```

5,000 vets at ~500 bytes each ≈ 2.5MB JSON. Client-side search handles this. Medicine data lazy-loaded by letter chunk. The "backend" is a batch export that runs periodically on Layer 1 and pushes to S3.

This pattern works for up to ~50k items. Beyond that, pre-built search indexes (Pagefind) extend it to ~500k items without adding a server.

---

## Site API Router (Layer 2 Backend)

For sites needing POST endpoints or server-side queries. A shared Go service loading route configurations from the DB.

```go
// site_api_routes configuration in site_specs:
{
  "routes": [
    {
      "method": "POST",
      "path": "/api/contact",
      "actions": ["validate_input", "store_submission", "send_email"],
      "config": {"table": "form_submissions", "notify": "owner@example.com"}
    }
  ]
}
```

Reuses the existing action library. Configuration-driven — add routes via DB, no code deploy. The core cluster manages route configs as part of site specs.

### Minimum viable Layer 2

```
OVH VM (extending existing Terraform pattern)
  ├── nginx (reverse proxy, SSL via certbot, rate limiting)
  ├── site-api-router (Go binary, systemd service)
  ├── Postgres (form submissions, entity data, route configs)
  ├── Prometheus + Grafana (monitoring)
  ├── fail2ban (security)
  └── cron jobs (data exports, cert renewal)
```

---

## Human Direction

Humans influence sites at three levels. Each has different persistence and routing.

### Direction channels

| Channel | What it does | Persistence | Route |
|---------|-------------|-------------|-------|
| Work item request | "Do this specific thing" | Until completed | Dashboard → work item → dispatch loop |
| Direction update | "The site should be like this" | Permanent until changed | Dashboard → site_specs → strategist/planner/auditors |
| Reference suggestion | "Look at this site/idea" | Feeds next planning cycle | Dashboard → research context |

### Build-time direction

Already works. The trigger `input_data` carries instructions alongside domain/email/phone. The classifier sees it, incorporates it into specs. No changes needed.

### Post-build direction (any time)

The site is live. The human wants to add features, change tone, suggest references. This needs a dashboard panel that writes to `site_specs` as a `direction` aspect.

The direction spec is pinned — agents can read it but not supersede it. Only humans change direction.

```json
site_specs aspect: "direction" (pinned)
{
  "requested_features": [
    {"type": "tool", "name": "loan repayment calculator", "priority": "must_have"},
    {"type": "content", "name": "first-time buyer guides", "priority": "should_have"}
  ],
  "reference_sites": [
    {"url": "https://moneysavingexpert.com/mortgages", "what_to_take": "content depth"}
  ],
  "notes": "Keep it simple, no jargon, focus on first-time buyers"
}
```

The strategist/planner reads direction and creates appropriate work items. The improvement loop respects it — auditors won't remove `must_have` features.

### HITL-requested content and lock lifecycle

When humans request specific content or features, the resulting components need protection from auditors. But they also shouldn't be frozen forever.

**Lock types:**

| Lock type | Behaviour | Use case |
|-----------|-----------|----------|
| `permanent` | Never expires, manual unlock only | Brand elements, legal disclaimers, human-crafted content |
| `timed` | Expires after N days (e.g. 90) | HITL-requested content that should eventually re-enter improvement cycle |
| `review` | Creates HITL review item on expiry | Content needing human approval before agents touch it again |

Implementation: `lock_type` and `lock_expires_at` columns on `page_components` and `site_components`. Discovery check queries expand from `AND locked_at IS NULL` to `AND (locked_at IS NULL OR lock_expires_at < NOW())`.

When a `review` lock expires, a discovery check creates a `needs_lock_review` work item. Human decides: re-lock, release, or update. If no human response within a configurable period, the system can auto-release (per-site config).

### Audit pass cap reset

The 3-pass audit cap prevents unbounded improvement cycles. Resets:

| Trigger | Mechanism |
|---------|-----------|
| Time-based (default) | Improvement sweep checks `last_audit_reset_at`, resets after N days (e.g. 60) |
| Direction change | Human updates direction spec → auto-reset |
| Major rebuild | N pages rebuilt in one cycle → auto-reset |
| Manual | Human clicks "re-audit" in dashboard |

Combined with timed lock expiry, this creates a natural rhythm:

```
Build → audit × 3 → cap reached → site quiet
  ... 60 days pass ...
  → pass counter resets, expired locks release
  → improvement loop runs fresh
  → finds new issues (content aged, design dated, new opportunities)
  → audit × 3 → quiet again
```

---

## Site Adoption

### Core principle

Adoption is a one-off data capture, not a permanent state. After adoption, the site is like any other — evolving, improving, with the strategist setting aspirational goals.

### What adoption is NOT

Adoption does not write a spec that describes the status quo as the target. If the adopted site has 5 pages, the spec shouldn't say "this site should have exactly 5 pages." The adoption captures what exists. The strategist then writes what the site SHOULD BECOME — which goes beyond what was adopted.

### Separation of concerns

```
crawl data              → stored in research_results (full markdown + rawHTML per page)
                           available for content writer recreate mode
                           available for revert/comparison later
                           NOT in site_specs

identity spec           → what the site is (company name, industry, tone)
                           includes adopted_from and adopted_at for provenance
                           evolves with improvements

design_reference spec   → what the original site looked like (concrete values)
                           hex colours, font families, CSS variables, layout patterns
                           extracted from crawled HTML + external CSS by Go action
                           historical record — not modified after adoption

design_intent spec      → what this site should look like (semantic + reference values)
                           character descriptions ("dark IDE aesthetic, functional not atmospheric")
                           reference values as starting points, not exact targets
                           auto-generated from design_reference by LLM at adoption time
                           evolves via strategist, human, or improvement loop proposals

content_direction spec  → writing style (aspirational, written by strategist)
structure spec          → page layout (evolves as site grows)
site_archetype spec     → what kind of site this is (character, constraints, purpose)
```

### The adoption agent

A dedicated `site-adoption-agent` triggered with a domain and URL. Runs its own workflow, not through the classifier pipeline. The adoption flow has a different starting point from a new build — it starts with a crawl, not a brief.

```
./trigger-adopt-site.sh gamedesign.uk

site-adoption-agent workflow:
  1. ensure_site_record
  2. firecrawl_crawl (all pages, async via webscrape adapter)
  3. format_crawl_for_analysis (produce lightweight summaries for LLM)
  4. check_crawl_content (conditional: proceed or fail)
  5. extract_design_fingerprint (Go: parse <style> blocks, inline styles, <link> tags from rawHTML)
  6. check_has_external_css (conditional: fetch CSS or skip)
  7. fetch_primary_css (firecrawl_scrape via webscrape adapter — only if external CSS found)
  8. enrich_fingerprint_with_css (Go: merge fetched CSS into fingerprint)
  9. analyze_site (LLM: classify site structure from summaries — output_format: json)
  10. classify_archetype (LLM: site character, design patterns, purpose, constraints)
  11. select_representative_content (Go: pick 2-3 prose-heavy pages)
  12. derive_content_direction (LLM: writing style guide from actual content)
  13. apply_adoption_plan (Go: write specs, pages, work items, extract content)
  14. generate_design_intent (LLM: semantic design brief from fingerprint + identity)
  15. write_design_intent (write_site_spec: persist design_intent to site_specs)
  16. complete
```

After completion, the dispatch loop picks up work items (`needs_design`, `needs_content_page × N`, `needs_rerender`) and builds the site through the normal pipeline.

### Three-stage processing: Go extracts design, LLM classifies, Go extracts content

The adoption agent splits work between Go and LLM based on what each is good at.

**Stage 1 — Go design extraction (no LLM):**

The `extract_design_fingerprint` step parses rawHTML from every crawled page using goquery. It extracts hex colours (classified as background/text/accent), font families from `<style>` blocks and inline styles, CSS variable declarations, layout patterns (grid/flex, max-width, spacing), and dark section detection. Google Fonts URLs are parsed from `<link>` tags.

If external CSS files are found (e.g. `<link rel="stylesheet" href="css/global.css">`), the workflow fetches them via the webscrape adapter (`firecrawl_scrape`) and merges the parsed data into the fingerprint. This is where the most valuable design tokens live — CSS custom properties like `--bg-color`, `--primary-color`, `--font-family`.

The fingerprint output includes a `suggested_mapping` that translates the original site's variable names to our CSS variable conventions.

**Stage 2 — LLM classification (lightweight):**

The `format_crawl_for_analysis` step produces a summary of each crawled page — ~500 characters, enough to see headings and the first paragraph. This goes to the LLM, which classifies the site: identity (company name, industry, tone), page types, section types per page, interactive features. The LLM reasons about structure from summaries.

For a 50-page site: 50 × 500 = ~25k characters. Well within context limits regardless of site size.

A second LLM step (`classify_archetype`) produces a site archetype classification: character, visual density, polish level, commercial intent, content model, and constraints the improvement loop should respect.

A third LLM step (`derive_content_direction`) analyses 2-3 representative pages to extract a detailed writing style guide.

A fourth LLM step (`generate_design_intent`) reads the fingerprint and identity to produce a rich semantic design brief — character descriptions explaining WHY the design values work, with the extracted values as reference starting points.

**Stage 3 — Go content extraction and plan application (no LLM):**

The `apply_adoption_plan` Go action reads the full crawl response directly (not through the LLM). It builds an index of page markdown keyed by URL using `buildCrawlPageIndex`. When creating page records, it matches each page's URL to the crawl index and stores the full markdown in `research_results`.

It also writes `design_reference` spec from the fingerprint data (preferring concrete fingerprint values over the LLM's vague design descriptions) and creates work items with enriched specs — the `needs_design` work item includes the fingerprint's suggested mapping, CSS variables, typography, and dark section data so the webdesign-agent has concrete values to work from.

Why Go and not LLM for content extraction:
- Firecrawl returns clean markdown. Headings are `#`, paragraphs are text. The content is already structured — nothing to interpret.
- No token limits. A 200-page site is just a loop. Each page can be arbitrarily long.
- No cost. No API calls per page.
- Deterministic. Same input always produces same output.
- The LLM would be doing expensive copy-paste. Its value is in classification and reasoning, not transcription.

```
Crawl result (full markdown + rawHTML per page)
  │
  ├──→ extract_design_fingerprint (Go) → colours, fonts, CSS vars, layout
  │       │
  │       └──→ [if external CSS] fetch via webscrape → enrich_fingerprint
  │                                                        ↓
  │                                              design_fingerprint (enriched)
  │
  ├──→ format_crawl_for_analysis → 500 char summaries → LLM (classification)
  │                                                        ↓
  │                                              pages, sections, identity, design
  │
  ├──→ classify_archetype (LLM) → site character, constraints
  │
  ├──→ derive_content_direction (LLM) → writing style guide
  │
  ├──→ apply_adoption_plan (Go) → buildCrawlPageIndex → full content per URL
  │                                                        ↓
  │                                              research_results per page
  │                                              page records + work items
  │                                              design_reference spec (from fingerprint)
  │
  └──→ generate_design_intent (LLM) → semantic design brief
                                         ↓
                                write_design_intent → design_intent spec
```

### What gets stored where

| Data | Location | Purpose |
|------|----------|---------|
| Full crawl analysis (LLM output) | `research_results` (result_type: `adoption_crawl`) | Reference, revert, comparison |
| Per-page markdown content | `research_results` (result_type: `adoption_page`) | Content writer source material |
| Identity (company, industry, tone) | `site_specs` aspect: `identity` | Ongoing site definition |
| Design reference (concrete CSS values) | `site_specs` aspect: `design_reference` | Historical record of original design. Extracted by Go from crawled HTML + external CSS. Contains hex colours, font families, CSS variables, layout patterns, suggested mapping to our variable names. |
| Design intent (semantic brief) | `site_specs` aspect: `design_intent` | Forward-looking design direction. LLM-generated from fingerprint + identity. Character descriptions with reference values as guidance. Read by webdesign-agent. Evolves via strategist or human. |
| Site archetype (character, constraints) | `site_specs` aspect: `site_archetype` | What kind of site this is. Constraints the improvement loop must respect. |
| Content direction (writing style) | `site_specs` aspect: `content_direction` | Detailed writing style guide. LLM-derived from actual content samples. |
| Page structure | `site_specs` aspect: `structure` | Planner reference |
| `adopted_from` URL | Inside `identity` spec | Provenance only |

No raw crawl data in site_specs. The specs contain clean, forward-looking data. Research_results holds the raw material.

### Adoption modes

| Mode | Trigger | What happens |
|------|---------|-------------|
| Take over domain | `./trigger-adopt-site.sh mortgagecalculator.co.uk` | Adoption agent: crawl → classify → recreate → improve |
| Build inspired by | `{"domain": "loancalculator.co.uk", "reference_sites": ["https://mortgagecalculator.co.uk"]}` | Classifier uses reference in research, strategist takes what's useful, content is fresh |
| Normal build | `{"domain": "newsite.co.uk"}` | Classifier researches the vertical, finds competitors organically |

The adoption agent handles "take over domain." The other modes go through the existing classifier/planner pipeline with richer input_data.

### Post-adoption: the strategist writes aspirational direction

After the adoption agent completes, the site has pages and content matching the original. The improvement loop's first run triggers the strategist, which reads the identity spec and current page structure. The strategist writes `content_direction` that goes BEYOND what was adopted: "This site has mortgage calculators. It should also have comparison tools, educational content about interest rates, and a blog covering market changes."

The improvement loop then has room to push toward the aspiration. The adopted state is the starting line, not the finish.

---

## Component Selector and Creator

### The separation of concerns

The planner and component selection have been conflated — the planner picks specific component names from a list in its prompt. This conflation means new components require prompt updates and there's no scoring or learning.

The separation:

- **Planner** decides WHAT section types a page needs: "this page should have a hero, a tool-grid, and a call-to-action." Structural decision based on page type, site type, and content requirements.
- **Component selector** decides WHICH template to use for each section type: "for a tool-grid section on a developer-tools site, use categorised-tool-grid." Matching + scoring decision based on metadata.
- **Component creator** handles the "no suitable component" path: generates a template following contracts, stores it for reuse.

New components are automatically discoverable without prompt changes — the selector queries the DB.

### What `function` does today and why it's doing two jobs

`content_components.function` is the unique identifier that threads through the entire pipeline. When the planner assigns `"hero"` to a page section, that string flows through:

1. **plan_sections** loads `content_components WHERE function = 'hero'` to get the input_schema
2. **Content writer** loads the `html_template` by function to fill template variables
3. **Page assembly** stores `page_components.slot_name = 'hero'` — how the section is identified on the page
4. **Rerendering** joins page_components back to content_components via function to re-apply templates
5. **CSS generation** creates `.hero-section` class names from the function
6. **data-component attribute** on the root HTML element matches function exactly — used by auditors and maintenance agents

This means `function` currently conflates two roles:

- **"What role does this section play on the page?"** — it's a hero, it's social proof, it's a call to action. The abstract purpose.
- **"Which specific HTML template do we use?"** — there's one template called `hero`, one called `social-proof`. The concrete implementation.

These are the same thing today because there's roughly one template per purpose. But the moment you want two different hero layouts — `hero-split` (image left, text right) and `hero-fullwidth` (big background, text centred) — `function` has to pick one. The planner makes a structural decision AND an implementation decision in the same LLM call, from a prompt that lists all available function names.

`section_type` separates these:

- **section_type** = "what role does this section play?" → `hero`
- **function** = "which specific template?" → `hero-split`

Multiple components can share the same `section_type` but each has its own unique `function`. The planner says "this page needs a hero section." The selector picks the best template. The winning function flows through the rest of the pipeline exactly as before. Nothing downstream changes.

### What makes site types different (currently vs after)

**Currently:** there is no structured relationship between site_type and components. The classifier says "brochure," the planner LLM sees the full component list, and the LLM uses general knowledge to pick appropriate components. A new site_type that the LLM hasn't seen triggers guesswork. No learning from what worked.

**After:** `suitable_site_types` on each component makes the relationship explicit and queryable. When the selector queries for a `hero` section_type for an `interactive-platform` site, components declaring `"interactive-platform"` in their `suitable_site_types` score higher than those only declaring `"brochure"`. Over time, `usage_count` and `avg_quality_score` add empirical evidence — a component used 30 times with score 0.9 gets picked over one used twice with score 0.6, even if both claim to suit the same site type.

### Backward compatibility

The selector integrates into `plan_sections` as a fallback path:

```
for each section name from the planner:
  1. Try direct function lookup (existing path — handles all current sites)
     → found? Use it. Nothing changed.
  2. Try section_type lookup via selector (new path)
     → found? Use the winning component's function. Pipeline continues as normal.
  3. Neither? Create needs_new_component work item. Section is deferred.
```

Existing sites that send function names hit path 1 and are unaffected. New sites that send section_types hit path 2. Sites requesting section_types that don't exist yet hit path 3, which triggers the component-creator.

### Selection metadata on content_components

Each component carries selection metadata (new columns via schema migration):

| Column | Type | Purpose |
|---|---|---|
| `section_type` | text | The abstract section type this component implements (e.g. "hero", "tool-grid", "provocation-card") |
| `suitable_site_types` | jsonb | Which site types it fits: `["brochure", "saas-platform"]` |
| `suitable_page_types` | jsonb | Which page types: `["landing", "product-listing"]` |
| `content_shape` | text | What kind of content it expects: `"structured_list"`, `"prose"`, `"structured_card"` |
| `visual_density` | text | How much it packs in: `"low"`, `"medium"`, `"high"` |
| `usage_count` | integer | Incremented when assigned to a page |
| `avg_quality_score` | float | From auditor feedback, NULL = unproven |
| `created_from` | text | Provenance: `"manual"`, `"generated"`, `"adopted"` |

### How selection scoring works

When the planner says "this page needs a tool-grid section," the selector queries:

```sql
SELECT * FROM content_components
WHERE section_type = 'tool-grid'
  AND component_level = 'section'
ORDER BY
  CASE WHEN suitable_site_types @> $site_type THEN 0.4 ELSE 0.1 END
  + COALESCE(avg_quality_score, 0.3) * 0.3
  + CASE WHEN array_length(suitable_site_types, 1) < 3 THEN 0.2 ELSE 0.05 END
  + LEAST(usage_count::float / 50.0, 1.0) * 0.1
DESC
LIMIT 3
```

Top candidate is used. If no candidates score above threshold (or none exist for that section_type), the system creates a `needs_new_component` work item.

### The "no suitable component" path

```
Planner: "page needs a provocation-card section"
Selector: no components with section_type = 'provocation-card'
  → Creates work item:
      item_type: needs_new_component
      spec: {
        section_type: "provocation-card",
        site_type: "interactive-platform",
        page_context: "landing page for game-like social platform",
        description: "Challenge card with provocation text, text input, countdown timer",
        design_direction: "Dark theme, vibrant accents, game energy"
      }
      handler: component-creator
  → Page build pauses (depends_on this work item) OR uses a fallback generic component
```

### Component creator handler

The `component-creator` agent processes `needs_new_component` work items:

1. Reads the spec (section_type, description, design_direction, reference content)
2. Loads component creation contracts into the prompt (see below)
3. LLM generates `html_template` + `input_schema`
4. Stores in `content_components` with full selection metadata
5. Marks work item complete
6. Dependent page builds continue

### Component creation contracts

The component-creator's LLM prompt must include the full component contract so every generated template follows the system's rules. Compiled from 003 (contracts and standards), 018 (dynamic application guidelines), and the input schema v2 contract.

```
You are creating a reusable HTML component template for the agent system.

SECTION TYPE: {{.section_type}}
SUITABLE SITE TYPES: {{.suitable_site_types}}
DESCRIPTION: {{.description}}
DESIGN DIRECTION: {{.design_direction}}
REFERENCE CONTENT: {{.reference_content}}

== COMPONENT CONTRACT — YOU MUST FOLLOW ALL OF THESE ==

1. STRUCTURE:
   <style> scoped CSS </style>
   <section class="{function}-section" data-component="{function}">
     HTML using {{.variable}} template placeholders
   </section>
   <script> if interactive, self-contained JS in IIFE </script>

2. NAMING:
   - function value in kebab-case: lowercase, digits, hyphens only
   - Root element has data-component="{function}" matching exactly
   - Class on root element: {function}-section

3. TEMPLATE VARIABLES:
   - Use {{.field_name}} for all content that varies per instance
   - Generate an input_schema declaring each field:
     {"fields": {"field_name": {"type": "text|array|image|url|boolean",
      "source": "llm|site_specs.{path}|site_assets.{type}|renderer|static",
      "required": true|false, "llm_guidance": "hint for content writer"}}}

4. CSS RULES:
   - ALL colours via CSS variables with fallbacks
   - Light sections: color: var(--color-text); headings: var(--color-heading)
   - Dark sections: color: var(--section-text, rgba(255,255,255,0.9));
     headings: var(--section-heading, #ffffff)
   - NEVER hardcode hex colours on text elements
   - Scope ALL CSS to .{function}-section — no global element rules (h1 {}, p {})
   - Include @media (max-width: 768px) responsive rules
   - Mobile-first: touch targets >= 44px

5. DARK SECTIONS (if is_dark_section = true):
   Set these CSS custom properties on the root container:
   --section-text: rgba(255,255,255,0.9);
   --section-text-muted: rgba(255,255,255,0.7);
   --section-heading: #ffffff;
   --section-surface: rgba(255,255,255,0.05);
   --section-border: rgba(255,255,255,0.2);

6. CSS VARIABLES AVAILABLE:
   Colours: --color-primary, --color-primary-hover, --color-primary-text,
     --color-secondary, --color-accent, --color-text, --color-text-muted,
     --color-heading, --color-background, --color-surface, --color-card-bg,
     --color-border, --color-header-bg, --color-header-text,
     --color-footer-bg, --color-footer-text, --color-white
   Layout: --container-max-width (1200px), --spacing-section (5rem 2rem),
     --border-radius, --shadow

7. INTERACTIVE ELEMENTS (if section has JS):
   - Client-side only, no external API calls
   - Wrap in IIFE: (function() { ... })();
   - No global variable pollution
   - Progressive enhancement — works without JS where possible
   - No external CDN imports unless explicitly listed

8. QUALITY:
   - No placeholder text (Lorem ipsum, TODO, [INSERT], NEEDS HUMAN REVIEW)
   - No unrendered template variables in output
   - Semantic HTML (section, article, nav, header — not div soup)
   - Accessible: labels on inputs, ARIA where needed, focus states
   - No fabricated content (stats, testimonials, quotes)

== END CONTRACT ==

Generate the html_template and input_schema for this component.
```

Contract storage: in the component-creator agent definition's `default_config.prompt_template` for now. If contracts evolve frequently, extract into the knowledge_base for `rag_lookup` at runtime.

### The variant concept

A section_type like "hero" might have multiple component variants:

```
hero-centered      → section_type: "hero", visual_variant: "centered"
hero-split-image   → section_type: "hero", visual_variant: "split-image"
hero-gradient      → section_type: "hero", visual_variant: "gradient"
```

The selector picks the variant based on context — does the site have a hero image? Is the design minimal or bold? Which variant scored best on similar sites? The planner doesn't need to know about variants.

### Component discovery during adoption

When the adoption agent classifies a crawled site, it identifies section patterns. For sections matching existing section_types, proceed normally. For novel sections:

1. The page build creates a `needs_new_component` work item with the observed pattern description and reference markdown
2. The component-creator generates the template
3. The new component enters the library with metadata
4. Future sites with similar pages find and reuse it

### Quality feedback loop

```
New component created (score: null, usage: 0)
  → First site uses it → content writer fills → deployed
  → Auditor scores rendered output → 0.7
  → avg_quality_score updated
  → Second site uses same component → better fill → auditor: 0.85
  → usage_count = 2, avg_quality_score = 0.775
  → Manual improvement to template → score jumps to 0.9
  → Meanwhile, another variant for same section_type has score 0.95, usage 30
  → The better one wins for most sites
  → But the generated one might still win for specific site types
```

Low-scoring components (below 0.5 after 3+ uses) get flagged for review. High-scoring components get promoted in selection. Natural fitness landscape.

### Category growth

| Category | Examples | Grows through |
|---|---|---|
| `hero` | hero, hero-split, hero-minimal | Existing library |
| `features` | differentiators-3-column, feature-grid | Existing library |
| `interactive-platform` | provocation-card, lobby-grid, gauntlet-interface | Novel site builds |
| `community` | reaction-panel, chain-display, duel-interface | Future builds |
| `tool-calculator` | ab-test-calculator, mortgage-calculator | Tool library |
| `custom` | Catch-all for one-offs | Any site with novel needs |

When the classifier outputs a site_type, the selector loads components filtered by matching `suitable_site_types`. First build generates them. Second build reuses them. The library grows organically.

---

## Research, Patterns, and the Component Library

### Research data retention

Currently, classifier research is transient — lives in `collected_data`, feeds into classification, then gone. This wastes valuable crawl data.

Change: the classifier persists its research findings in `site_specs` as a `research` aspect. This is a summary of what was found (industry patterns, competitor analysis, opportunities), not the raw crawl dump.

### Pattern extraction (Phase 3)

A separate `pattern-extraction-agent` analyses research data and extracts reusable patterns for the component library:

- **Tool specs:** "A mortgage calculator that accepts loan amount, rate, term. Shows monthly payment, amortisation breakdown, multiple scenarios." Detailed enough for the tool-builder to create a functionally equivalent tool.
- **Layout patterns:** "Financial calculator sites typically have: hero with key stat, calculator tool, comparison table, educational guides, FAQ, trust signals."
- **Content patterns:** "Guide structure: intro, key takeaways box, detailed sections with examples, related links, disclaimer."
- **UX observations:** Good and bad examples with descriptions. "Live-updating results as inputs change (good). Results only after clicking calculate (bad)."

Runs as a side effect after classification, or on a schedule, or manually triggered. Patterns accumulate over time and become available to future planners/builders.

### Code as reference input

For complex tools (games, interactive visualisations, advanced calculators), descriptions alone aren't enough for the LLM to understand the intended behaviour. The reference code is included in the tool-builder's prompt as context:

```
"Here is reference code for a similar tool. Study the mechanics,
interaction patterns, and scoring system. Create an ORIGINAL 
implementation achieving similar functionality with [these differences].
Do not copy the code."

Reference code:
[actual JS from research]
```

The LLM sees the code, understands how it works, produces something new. This is not copying — it's the same way a developer studies an existing implementation before building their own.

**Copyright considerations:** We never deploy extracted code directly. The research stores code snippets as reference material. The tool-builder's prompt explicitly instructs original implementation. The reference material is internal — it's never served to users.

### Integration with RAG and training pipelines

Stored LLM prompts and successful outputs (tools, components, content) feed the existing RAG pipeline. When the tool-builder generates a mortgage calculator, the prompt + output pair is stored. Future tool-builder calls retrieve relevant examples via RAG. This improves output quality over time without fine-tuning.

The pattern library (tool specs, good/bad examples) enriches RAG retrieval — when building a calculator, the system retrieves both the abstract spec ("should have amortisation breakdown") and concrete examples ("here's a successful calculator we built previously").

---

## Mission-Driven Sites

Sites can carry a `mission` and `roadmap` in their `input_data`, bypassing the classifier's domain-discovery step. The classifier reads the mission to produce aligned specs. The planner reads the roadmap to know what to build in the current phase. See 003d (Spark Strategic Planning Architecture) for the full pattern including the `mission` and `roadmap` site_spec aspects and the input_data structure.

This mechanism supports any pre-planned site, not just Spark/vonc.com. Any site with a strategic brief can use the same pattern.

---

## Phase Plan

### Phase 1 — Adoption pipeline (current)

- Site-adoption-agent with dedicated workflow
- Firecrawl v2 crawl (scrapeOptions fix)
- Three-stage processing: Go design extraction, LLM classification, Go content extraction
- Full crawl stored in research_results, clean specs in site_specs
- **Design fingerprint extraction** — Go action parses rawHTML for colours, fonts, CSS variables, layout patterns. External CSS files fetched via webscrape adapter and merged. Produces `design_reference` spec with concrete values.
- **Design intent generation** — LLM produces semantic design brief from fingerprint + identity. Produces `design_intent` spec with character descriptions and reference values as guidance.
- **Webdesign-agent three-way priority** — reads design_intent (creative freedom) → design_reference (reproduce faithfully) → generate from industry (new builds). Design is locked to reference until design_intent is written.
- Site archetype classification — character, constraints, visual density, purpose
- Content direction extraction — detailed writing style guide from actual content
- Content writer `existing_content` / `mode: "recreate"` support (pending)
- Dashboard direction panel for post-build HITL (pending)

### Design evolution lifecycle

```
Adoption completes → design_reference (locked values) + design_intent (semantic brief)
  → Webdesign-agent reads design_intent → generates CSS with creative freedom
  → Audit loop checks deployed CSS against design_intent
  → Audit proposes changes (work items) — does NOT modify design_intent directly
  → Strategist or human updates design_intent → palette evolves
  → 3-pass audit cap prevents unbounded cycles
  → design_reference stays as historical record throughout
```

### Phase 2 — Tool and entity adoption

- Tool extraction: detailed specs from reference code
- Entity extraction: structured data from directory/listing pages
- Data export action: DB → JSON → git → S3
- Lock types (permanent, timed, review) on components
- Component discovery: `needs_new_component` work items for novel section types

### Phase 3 — Component selector, patterns, and research

- Schema migration: section_type and selection metadata on content_components
- Component selector with scoring (Go function used by plan_sections)
- Component-creator handler agent with contract prompt
- Component quality feedback from auditors
- Planner outputs section_types, not component names
- Pattern extraction agent
- Tool spec library (reusable across sites)
- Good/bad example library for LLM prompts
- RAG integration for pattern retrieval
- Audit pass auto-reset (time-based)

### Phase 4 — Client-facing backend (Layer 2)

- Site-api-router Go service
- OVH VM provisioning with Terraform
- Form handling, data queries
- Auth and persistent state

### Phase 5 — Framework deployments (Layer 3)

- Infrastructure provisioning agents
- Framework deployment handlers (LlamaIndex, LangGraph)
- Server entity management
- Self-managing infrastructure

---

## Principles

1. **Layer 1 never serves external traffic.** It produces artifacts and pushes them.

2. **No vendor lock-in.** Use third-party services but always with the exit path designed. The site-api-router is our own code. Infrastructure provisioning supports multiple providers.

3. **Static-first.** Default to the lowest tier that serves the need. The JSON-on-S3 pattern handles more cases than expected.

4. **Same pipeline, richer specs.** Adoption, new builds, maintenance, and framework deployment all flow through the work item pipeline. The orchestration layer is the same.

5. **Adoption is a starting point, not a ceiling.** The adopted state is the baseline. The strategist writes aspirational direction. The improvement loop pushes toward it. No spec describes the status quo as the target.

6. **LLM for reasoning, Go for extraction.** The LLM classifies, reasons, and generates. Go code handles content extraction, data transformation, CSS parsing, and anything deterministic. Don't send content through an LLM just to get it back unchanged. Don't ask an LLM to read hex values when a regex can do it.

7. **Design reference is history, design intent is direction.** The `design_reference` spec records what the original site looked like — concrete values extracted from real CSS. The `design_intent` spec describes what the site should look like — semantic character descriptions with reference values as guidance. The webdesign-agent reads intent, not reference. The audit loop checks against intent. Evolution happens by updating intent, not by modifying reference.

8. **Research is reference, not template.** Crawled code and content inform generation but are never deployed directly. The system creates original implementations inspired by research.

9. **Human direction persists.** HITL requests go into pinned specs. Agents respect them. Lock types control how long protection lasts. The system self-maintains but humans set the direction.

10. **The component library grows through use.** Adoption discovers new section types. Novel site builds trigger the component-creator. Quality scores accumulate from auditor feedback. Components that work well spread; poor ones get deprecated.

11. **Servers are entities.** Infrastructure follows the same patterns as site data — state-based lifecycle, discovery checks for health, work items for management tasks.

12. **The framework is a framework builder.** The same system that generates websites can generate infrastructure configs, framework deployments, and management automation.

---

## Implementation Fixes & Schema Notes (from 028j handoff)

### Fixes Applied

| Fix | Detail |
|-----|--------|
| Registry: `select_representative_content` | Action existed but wasn't registered |
| `CreateNeedsNewComponentItem` | Column `domain` → `pipeline`, added `::jsonb` cast |
| Claimed-item-timeout | Two-phase: 15-min evidence-based, 40-min blind reset |
| `error_step` placement | Must be inside `step.Config`, not at step level |
| `store_generated_component` | Added `stripCodeBlocks()` for markdown-wrapped LLM output |
| Component → page rebuild | `markPagesForRebuild()` + `check_unresolved_sections` discovery check |

### Schema Column Reminders

- `site_work_items`: `pipeline` not `domain`
- `sites`: `build_status` not `deploy_status`
- `scheduled_tasks`: `name` not `task_name`
- `site_specs`: `data` not `spec_data`
- `agent_definitions`: `image_repository` (not container_image), `resources` (not resource_config), `topics` (not topic_config), `agent_category` (not role), `domain_tags` (not tags)

### Known Issue: Zombie Dispatch Loop Pods

Loop-expanded steps lost from `workflow_plan` during concurrent state updates. Mitigation: 30-minute reaper threshold.

### Design Fingerprint Pipeline (added 2026-04-12)

| Component | File | Type |
|-----------|------|------|
| `extract_design_fingerprint` action | `extract_design_fingerprint_action.go` | Go action (local, no LLM) |
| `enrich_fingerprint_with_css` action | `enrich_fingerprint_with_css_action.go` | Go action (local, no LLM) |
| Adoption workflow steps | `agent_definitions` (type: `site-adoption-agent`) | SQL workflow config |
| Webdesign-agent prompt (three-way priority) | `agent_definitions` (type: `webdesign-agent`) | SQL workflow config |
| `design_reference` spec aspect | Written by `apply_adoption_plan` | site_specs |
| `design_intent` spec aspect | Written by `generate_design_intent` → `write_design_intent` | site_specs |

Key decisions documented in `design_adoption_work_plan.md`:
- design_intent is semantic not prescriptive (creative freedom for webdesign-agent)
- Audit loop proposes but doesn't apply design changes directly
- CSS variable name translation is mechanical (Go), character descriptions are LLM
- External CSS fetched via webscrape adapter, not direct HTTP from Go action
