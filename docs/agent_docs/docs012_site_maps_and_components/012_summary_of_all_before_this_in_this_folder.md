# Unified Component Integration Plan

## Overview

This plan defines when the system uses **database components** vs **LLM generation**, how **good examples** are collected, and how data flows through agents with clear contracts.

---

## Part 1: When to Use DB vs LLM

### The Decision Matrix

| Component Type | render_mode | DB Used | LLM Used | Notes |
|----------------|-------------|---------|----------|-------|
| **Headers/Footers** | `template` | ✅ html_template | ❌ | Never generated - always from library |
| **Structural (head, body-close)** | `template` | ✅ html_template | ❌ | Pure structure, no content |
| **Hero sections** | `template` | ✅ html_template | ✅ fills input_schema | DB provides structure, LLM fills content |
| **Service grids** | `template` | ✅ html_template | ✅ fills input_schema | DB provides structure, LLM fills content |
| **Testimonials** | `template` | ✅ html_template | ✅ or brief data | May use brief data if available |
| **FAQ** | `agent` | ✅ html_template | ✅ with research | Research-backed, LLM generates Q&A |
| **Long-form content** | `agent` | ❌ minimal template | ✅ with research | LLM generates paragraphs |
| **Pricing** | `template` | ✅ html_template | ❌ brief data | Pulled from questionnaire |
| **Contact** | `template` | ✅ html_template | ❌ brief data | Pulled from questionnaire |

### The Decision Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                  COMPONENT RENDER DECISION                          │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Has render_mode │
                    │ in DB record?   │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                             ▼
       render_mode =                 render_mode =
       'template'                    'agent'
              │                             │
              ▼                             ▼
    ┌─────────────────┐          ┌─────────────────┐
    │ Check if all    │          │ needs_research? │
    │ input_schema    │          └────────┬────────┘
    │ fields available│                   │
    │ from brief_data │         ┌─────────┴─────────┐
    └────────┬────────┘         ▼                   ▼
             │               Yes                   No
    ┌────────┴────────┐         │                   │
    ▼                 ▼         ▼                   ▼
 All fields       Missing   ┌──────────┐      ┌──────────┐
 available        fields    │ call     │      │ call LLM │
    │                │      │ research │      │ directly │
    ▼                ▼      │ agent    │      └────┬─────┘
┌──────────┐   ┌──────────┐ └────┬─────┘           │
│ RENDER   │   │ call LLM │      │                 │
│ TEMPLATE │   │ for      │      ▼                 ▼
│ directly │   │ missing  │ ┌──────────┐     ┌──────────┐
│ with     │   │ fields   │ │ call LLM │     │ LLM      │
│ brief    │   └────┬─────┘ │ with     │     │ generates│
│ data     │        │       │ research │     │ content  │
└──────────┘        ▼       └────┬─────┘     └────┬─────┘
                ┌──────────┐     │                 │
                │ RENDER   │     ▼                 ▼
                │ TEMPLATE │ ┌──────────┐     ┌──────────┐
                │ with LLM │ │ RENDER   │     │ RENDER   │
                │ content  │ │ TEMPLATE │     │ TEMPLATE │
                └──────────┘ │ with LLM │     │ with LLM │
                             │ content  │     │ content  │
                             └──────────┘     └──────────┘
```

### Concrete Examples

**Example 1: Header (pure template)**
```
Component: header-professional-dark
render_mode: template
input_schema: {logo_text, nav_items, cta_text, cta_url}
Data source: sites.content_data + navigation_structures
LLM: NOT USED
```

**Example 2: Hero section (template + LLM)**
```
Component: hero-centered
render_mode: template
input_schema: {headline, subheadline, cta_text, cta_url}
Data source: brief may have tagline, LLM fills gaps
LLM: Generates compelling headline/subheadline
```

**Example 3: FAQ section (agent + research)**
```
Component: faq-accordion
render_mode: agent
needs_research: true
input_schema: {items: [{question, answer}]}
Data source: research-agent finds common questions
LLM: Synthesizes research into Q&A pairs
```

---

## Part 2: The Data Flow

### Full Pipeline

```
┌─────────────────────────────────────────────────────────────────────┐
│ 1. INTAKE                                                           │
│    Input: {domain, objective, model?}                               │
│    Action: ensure_site_record                                       │
│    Output: site_record {site_id, domain}                            │
│    DB Write: sites table                                            │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 2. PLANNING (site-planner or chief-strategist)                      │
│    Input: input_data, site_record, component_library (from DB)      │
│    Action: execute_llm_prompt (with component catalog in prompt)    │
│    Output: site_plan {pages[], sitemap[], style_collection}         │
│    DB Write: none (yet)                                             │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 3. SYNC PAGES TO DB                                                 │
│    Input: site_id, site_plan                                        │
│    Action: sync_pages_to_db                                         │
│    Output: navigation structure, page_ids                           │
│    DB Write: pages table, navigation_structures (via trigger)       │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 4. SELECT STYLE COLLECTION                                          │
│    Input: site_plan.style_collection OR domain keywords             │
│    Action: select_style_collection                                  │
│    Output: style_collection {header, footer, colors, typography}    │
│    DB Read: style_collections table                                 │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 5. PAGE LOOP (for each page in site_plan.pages)                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┴─────────────────────┐
        ▼                                           │
┌─────────────────────────────────────────────────┐ │
│ 5a. LOAD PAGE COMPONENTS                        │ │
│     Input: current_page.sections[]              │ │
│     Action: load_page_section_components        │ │
│     Output: section_components with templates   │ │
│     DB Read: content_components by function     │ │
└─────────────────────────────────────────────────┘ │
                              │                     │
                              ▼                     │
┌─────────────────────────────────────────────────┐ │
│ 5b. BUILD RENDER CONTEXT                        │ │
│     Input: brief, site_record, style, assets    │ │
│     Action: build_render_context                │ │
│     Output: render_context (flat object)        │ │
│     DB Read: sites.content_data, brand_assets   │ │
└─────────────────────────────────────────────────┘ │
                              │                     │
                              ▼                     │
┌─────────────────────────────────────────────────┐ │
│ 5c. SECTION LOOP (for each section)             │ │
│     ┌─────────────────────────────────────────┐ │ │
│     │ check_render_mode                       │ │ │
│     │   'template' → render_from_template     │ │ │
│     │   'agent' → check_needs_research        │ │ │
│     │              → call_researcher?         │ │ │
│     │              → generate_content         │ │ │
│     │              → render_section           │ │ │
│     └─────────────────────────────────────────┘ │ │
│     Output: rendered_section HTML               │ │
└─────────────────────────────────────────────────┘ │
                              │                     │
                              ▼                     │
┌─────────────────────────────────────────────────┐ │
│ 5d. COMPILE PAGE                                │ │
│     Input: rendered_sections[], header, footer  │ │
│     Action: compile_page_sections               │ │
│     Output: complete page HTML                  │ │
└─────────────────────────────────────────────────┘ │
                              │                     │
                              ▼                     │
┌─────────────────────────────────────────────────┐ │
│ 5e. EXTRACT LINKS                               │ │
│     Input: page_html                            │ │
│     Action: extract_and_sync_links              │ │
│     Output: link_count                          │ │
│     DB Write: link_registry                     │ │
└─────────────────────────────────────────────────┘ │
                              │                     │
        └─────────────────────┴─────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 6. ASSEMBLE SITE                                                    │
│    Input: all rendered pages                                        │
│    Action: assemble_multipage_site                                  │
│    Output: site_files {pages, sitemap.xml, robots.txt}              │
│    DB Read: navigation_structures for sitemap                       │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 7. DEPLOY                                                           │
│    Input: site_files, site_record                                   │
│    Action: call deployer-agent                                      │
│    Output: deployment_result {repo_url, commit_sha}                 │
│    DB Write: sites.last_deployed_at                                 │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Part 3: Collecting Good Examples

### The Pattern Extraction Pipeline

```
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 1: Site Discovery                                             │
├─────────────────────────────────────────────────────────────────────┤
│ Input: industry vertical, objective type                            │
│ Action: web_search for top sites in vertical                        │
│ Output: list of URLs to analyze                                     │
│ Example: "best SaaS landing pages 2024"                             │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 2: Site Capture                                               │
├─────────────────────────────────────────────────────────────────────┤
│ Agent: website-capture-firecrawl                                    │
│ Action: firecrawl_scrape with screenshot                            │
│ Output: {html, markdown, screenshot}                                │
│ DB Write: captured_sites table (future)                             │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 3: Structure Analysis (LLM)                                   │
├─────────────────────────────────────────────────────────────────────┤
│ Prompt: "Analyze this page. Identify:                               │
│   - Section types (hero, features, testimonials, etc.)              │
│   - Visual hierarchy and layout patterns                            │
│   - Content strategy (AIDA, PAS, etc.)                              │
│   - Conversion elements and their placement                         │
│   - Typography and color usage"                                     │
│ Output: structured analysis JSON                                    │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 4: Pattern Extraction (LLM)                                   │
├─────────────────────────────────────────────────────────────────────┤
│ Prompt: "Extract reusable patterns from this analysis.              │
│   For each pattern, identify:                                       │
│   - Pattern type (hero layout, social proof, CTA placement)         │
│   - Why it works (psychological principle)                          │
│   - When to use it (audience, stage in funnel)                      │
│   - Required elements (headline, image, button, etc.)"              │
│ Output: pattern definitions                                         │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│ PHASE 5: Component Creation                                         │
├─────────────────────────────────────────────────────────────────────┤
│ For each unique pattern:                                            │
│   1. Create html_template with placeholders                         │
│   2. Define input_schema (required fields)                          │
│   3. Tag with semantic_tags, category, function                     │
│   4. Note source_sites for attribution                              │
│   5. Set render_mode based on content needs                         │
│                                                                     │
│ DB Write: content_components table                                  │
│ DB Write: pattern_sources table (future - tracks origin)            │
└─────────────────────────────────────────────────────────────────────┘
```

### Pattern Storage Schema

```sql
-- Future table for tracking pattern sources
CREATE TABLE pattern_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    component_id UUID REFERENCES content_components(id),
    source_url TEXT NOT NULL,
    captured_at TIMESTAMPTZ DEFAULT now(),
    
    -- What we learned
    analysis_summary TEXT,
    psychological_principle TEXT,  -- "social proof", "scarcity", etc.
    effectiveness_notes TEXT,
    
    -- Where it applies
    industry_vertical TEXT,
    funnel_stage TEXT,  -- "awareness", "consideration", "decision"
    audience_segment TEXT,
    
    -- How it performed (if we can track)
    conversion_data JSONB  -- {avg_time_on_page, bounce_rate, etc.}
);

-- Enhanced content_components fields
ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    origin_type TEXT DEFAULT 'manual';  -- 'manual', 'extracted', 'generated'

ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    industry_tags TEXT[];  -- ['saas', 'finance', 'healthcare']

ALTER TABLE content_components ADD COLUMN IF NOT EXISTS
    funnel_stages TEXT[];  -- ['awareness', 'consideration']
```

### Example: Extracting a Hero Pattern

**Input: Screenshot + HTML of successful SaaS landing page**

**LLM Analysis Output:**
```json
{
  "section_type": "hero",
  "pattern_name": "hero-split-demo",
  "why_it_works": "Combines value prop (left) with proof (right). Demo video reduces friction by showing product without signup.",
  "visual_hierarchy": "60/40 split, headline largest, subhead medium, CTA prominent orange",
  "psychological_principles": ["social proof via demo", "clarity of value prop"],
  "when_to_use": {
    "audience": "technical decision-makers",
    "funnel_stage": "awareness",
    "product_type": "complex SaaS needing demo"
  },
  "required_elements": {
    "headline": "5-8 words, benefit-focused",
    "subheadline": "15-25 words, elaborates on benefit",
    "primary_cta": "action-oriented, contrasting color",
    "demo_element": "video embed or interactive preview"
  }
}
```

**Generated Component:**
```sql
INSERT INTO content_components (
    name, function, category, description,
    html_template, input_schema, render_mode,
    industry_tags, funnel_stages, origin_type
) VALUES (
    'hero-split-demo',
    'hero',
    'hero',
    'Split hero with value prop left and demo video right. For SaaS products needing visual proof.',
    '<section class="hero-split">
        <div class="hero-content">
            <h1>{{headline}}</h1>
            <p class="subheadline">{{subheadline}}</p>
            <a href="{{cta_url}}" class="btn-primary">{{cta_text}}</a>
        </div>
        <div class="hero-demo">
            {{#if demo_video_url}}
            <iframe src="{{demo_video_url}}" ...></iframe>
            {{else}}
            <img src="{{demo_image_url}}" alt="{{demo_alt}}">
            {{/if}}
        </div>
    </section>',
    '{
        "type": "object",
        "required": ["headline", "subheadline", "cta_text", "cta_url"],
        "properties": {
            "headline": {"type": "string", "maxLength": 60},
            "subheadline": {"type": "string", "maxLength": 200},
            "cta_text": {"type": "string"},
            "cta_url": {"type": "string"},
            "demo_video_url": {"type": "string"},
            "demo_image_url": {"type": "string"},
            "demo_alt": {"type": "string"}
        }
    }',
    'template',
    ARRAY['saas', 'tech', 'b2b'],
    ARRAY['awareness'],
    'extracted'
);
```

---

## Part 4: Agent Input/Output Contracts

### Contract Format

Each agent defines:
- **expects**: What fields it needs in input
- **required**: Which of those are mandatory
- **produces**: What fields it outputs

### Key Agent Contracts

#### site-planner
```json
{
  "expects": {
    "input_data": "object with domain, objective, model",
    "site_record": "object with site_id, domain",
    "component_library": "array of available components"
  },
  "required": ["input_data", "site_record"],
  "produces": {
    "site_plan": {
      "pages": "array of {name, title, sections[], purpose}",
      "sitemap": "array of {label, page, url, in_header, in_footer}",
      "style_collection": "string - name of style collection to use",
      "image_prompts": "object with prompts for logo, hero images"
    }
  }
}
```

#### page-content-writer
```json
{
  "expects": {
    "current_page": "object with name, title, sections[]",
    "site_record": "object with site_id, domain",
    "reviewed_brief": "object with company_name, services, about_us, etc",
    "style_collection": "object with colors, typography, component refs",
    "brand_assets": "object with logo, images (optional)"
  },
  "required": ["current_page", "site_record", "reviewed_brief"],
  "produces": {
    "page_content": {
      "page_name": "string",
      "sections": "array of {component_id, rendered_html, content_data}",
      "research_ids": "array of UUIDs referencing research_results"
    }
  }
}
```

#### research-agent
```json
{
  "expects": {
    "current_section": "object with topic or research_query",
    "reviewed_brief": "object with industry, company context",
    "site_record": "object with site_id for storing results"
  },
  "required": ["current_section"],
  "produces": {
    "id": "uuid - research_results record ID",
    "query": "string - the search query used",
    "summary": "string - synthesized findings with citations",
    "sources": "array of {url, title, domain, quotes[], accessed_at}",
    "source_count": "number"
  }
}
```

---

## Part 5: Flexible vs Strict Mode

### Mode Transitions

```
INITIAL BUILD (flexible mode)
  │
  │ 1. Domain → site created (schema_mode = 'flexible')
  │ 2. LLM plans pages, selects components
  │ 3. Components rendered with best-effort substitution
  │ 4. Missing fields: warn, continue
  │ 5. Human reviews, makes edits
  │
  ▼
APPROVAL (transition point)
  │
  │ - page_components.schema_snapshot = current input_schema
  │ - page_components.content_snapshot = actual content used
  │ - site.schema_mode = 'strict' (optional, can be per-page)
  │
  ▼
EDITING MODE (strict mode)
  │
  │ 1. Edits must match schema_snapshot
  │ 2. Missing required fields: error, block save
  │ 3. Component template updates don't break existing pages
  │ 4. Can rollback to content_snapshot
```

---

## Part 6: Implementation Priorities

### Phase 1: Core Component Flow (Current)
1. ✅ content_components table with html_template, input_schema
2. ✅ style_collections linking to header/footer components
3. ✅ page-content-writer with render_mode checks
4. 🔄 Fix template substitution (placeholders not rendering)
5. 🔄 research-agent integration

### Phase 2: Pattern Collection
1. Add pattern_sources table
2. Create site-analyzer agent
3. Workflow: discover → capture → analyze → extract → store
4. Tag components with industry/funnel metadata

### Phase 3: Smart Selection
1. Planner receives component_library with metadata
2. LLM selects components based on industry + objective
3. Fallback: use semantic_tags matching
4. Track which components perform well

### Phase 4: Multi-Track Sites
1. Add site_flows table
2. Pages belong to flows with context_overrides
3. Content varies by flow stage
4. Navigation is flow-aware

---

## Summary

| Question | Answer |
|----------|--------|
| **When DB only?** | Headers, footers, structural components, contact forms (brief data) |
| **When LLM fills DB template?** | Hero, features, services, pricing structure |
| **When LLM + research?** | FAQ, blog content, industry-specific copy |
| **How collect examples?** | Capture → Analyze → Extract patterns → Create components |
| **How ensure consistency?** | render_context merges site/page/section context; brand invariants checked |
| **How track contracts?** | input_contract/output_contract fields in agent_definitions |


----------

I'll search the project knowledge to understand the current codebase, then synthesize with your uploaded documents to create a unified plan.Now let me view the project files to understand the current workflow and schema structures:Now I have a comprehensive picture. Let me synthesize everything into a unified integration plan.Here's the synthesized plan covering your component/LLM strategy and example collection.

**Key decisions from the analysis:**

**When to use what:**
- **DB only (no LLM)**: Headers, footers, structural elements, contact sections with brief data
- **DB template + LLM content**: Hero, features, services - the template provides structure, LLM fills the `input_schema` fields
- **LLM with research**: FAQ, long-form content, industry-specific copy where the component's `render_mode = 'agent'` and `needs_research = true`

**The render_mode field** on `content_components` is the key differentiator:
- `template`: Render directly with data from brief or render_context
- `agent`: Spawn LLM (optionally with research-agent first)

**Collecting good examples** follows a pipeline:
1. Web search for top sites in vertical
2. Capture with firecrawl (HTML + screenshot)
3. LLM analyzes structure and identifies why things work
4. Extract reusable patterns with psychological principles noted
5. Create templated components tagged with industry/funnel metadata

**The data flow** you already have mostly works - the plan shows how each step uses DB vs LLM, and where the key contracts need to be clear (input_contract/output_contract on agent_definitions).


