# Updated Multipage-Website-Builder Workflow

## Changes for Flow-Based Architecture

This shows how the multipage-website-builder workflow evolves to use the multi-track architecture.

## New Workflow Steps

```json
{
  "workflow": {
    "start_step": "create_brand_dna",
    "steps": {
      
      "create_brand_dna": {
        "action": "execute_sql",
        "description": "Store site-level brand identity and invariants",
        "config": {
          "query": "INSERT INTO site_brand_dna ...",
          "input_fields": [
            "input_data.domain",
            "reviewed_brief.company_name",
            "reviewed_brief.tone",
            "questionnaire.theme_preference"
          ]
        },
        "next_step": "create_primary_flow",
        "output_field": "brand_dna_id"
      },
      
      "create_primary_flow": {
        "action": "call_agent",
        "description": "Chief strategist creates the primary conversion flow",
        "config": {
          "agent_type": "chief-strategist",
          "target_role": "strategist",
          "input_fields": [
            "input_data.domain",
            "input_data.objective",
            "reviewed_brief",
            "brand_dna_id"
          ],
          "timeout_seconds": 120
        },
        "next_step": "store_flow",
        "output_field": "flow_definition"
      },
      
      "store_flow": {
        "action": "execute_sql",
        "description": "Store the flow definition in database",
        "config": {
          "query": "INSERT INTO site_flows ...",
          "input_fields": ["flow_definition"]
        },
        "next_step": "generate_pages_sequential",
        "output_field": "flow_id"
      },
      
      "generate_pages_sequential": {
        "action": "loop",
        "description": "Generate each page in flow sequence",
        "config": {
          "iterate_over": "flow_definition.pages",
          "loop_var": "current_page",
          "max_iterations": 20,
          "substeps": {
            
            "research_page": {
              "action": "call_agent",
              "description": "Research content for this specific page",
              "config": {
                "agent_type": "content-researcher",
                "input_fields": [
                  "current_page",
                  "flow_definition.narrative_arc",
                  "reviewed_brief"
                ],
                "timeout_seconds": 60
              },
              "output_field": "page_research"
            },
            
            "plan_page_structure": {
              "action": "call_agent",
              "description": "Plan component structure for page",
              "config": {
                "agent_type": "site-component-architect",
                "input_fields": [
                  "current_page.archetype",
                  "current_page.components",
                  "page_research"
                ],
                "timeout_seconds": 30
              },
              "output_field": "page_structure"
            },
            
            "generate_page_content": {
              "action": "call_agent",
              "description": "Generate content with full flow context",
              "config": {
                "agent_type": "content-creator",
                "input_fields": [
                  "page_structure",
                  "current_page",
                  "flow_definition.narrative_arc",
                  "brand_dna_id",
                  "page_research"
                ],
                "context_layers": {
                  "brand": "brand_dna_id",
                  "flow_stage": "current_page.stage",
                  "page_overrides": "current_page.context_overrides"
                },
                "timeout_seconds": 180
              },
              "output_field": "page_html"
            },
            
            "store_page": {
              "action": "execute_sql",
              "description": "Store completed page in flow_pages",
              "config": {
                "query": "INSERT INTO flow_pages ...",
                "input_fields": ["flow_id", "page_html", "current_page"]
              },
              "output_field": "page_id"
            }
          }
        },
        "next_step": "generate_shared_assets",
        "output_field": "all_pages"
      },
      
      "generate_shared_assets": {
        "action": "generate_css_and_nav",
        "description": "Generate theme CSS and flow-aware navigation",
        "config": {
          "theme_field": "questionnaire.theme_preference",
          "flow_field": "flow_definition",
          "pages_field": "all_pages"
        },
        "next_step": "assemble_site",
        "output_field": "shared_assets"
      },
      
      "assemble_site": {
        "action": "assemble_multi_page_site",
        "description": "Combine all pages with shared assets and navigation",
        "config": {
          "pages_field": "all_pages",
          "assets_field": "shared_assets",
          "flow_field": "flow_definition"
        },
        "next_step": "deploy",
        "output_field": "site_files"
      },
      
      "deploy": {
        "action": "call_agent",
        "description": "Deploy to Git with flow metadata",
        "config": {
          "agent_type": "deployer-agent",
          "target_role": "deployer",
          "input_fields": [
            "site_files",
            "input_data.domain",
            "flow_id"
          ],
          "timeout_seconds": 120
        },
        "next_step": "complete",
        "output_field": "deployment_result"
      },
      
      "complete": {
        "action": "complete_workflow",
        "description": "Site deployed with flow-based architecture"
      }
    }
  }
}
```

## Key Changes Explained

### 1. Brand DNA First

**Why:** Establish immutable brand identity before any content generation.

**What it stores:**
- Visual theme
- Core message
- Core values
- Voice parameter boundaries
- Forbidden/required phrases

**Usage:** All downstream agents reference this for consistency checks.

### 2. Flow Creation

**Chief Strategist new responsibility:**
- Analyze target audience
- Define narrative arc (3-4 stages)
- Set voice parameters per stage
- Recommend page archetypes
- Create page sequence

**Output example:**
```json
{
  "flow_name": "primary_conversion",
  "audience_segment": "c_suite_executives",
  "narrative_arc": {
    "stage_1": {
      "name": "awareness",
      "voice_formality": 0.7,
      "technical_depth": 0.5,
      "sales_pressure": 0.2
    },
    "stage_2": {...},
    "stage_3": {...}
  },
  "pages": [
    {
      "path": "index.html",
      "stage": "stage_1",
      "sequence": 1,
      "archetype": "corporate_home",
      "components": ["hero", "value_props", "cta"]
    },
    ...
  ]
}
```

### 3. Sequential Page Generation (Not Batches)

**Old way:**
```
generate_batch_1 (pages 1-4)
generate_batch_2 (pages 5-8)
...
```

**New way:**
```
FOR EACH page IN flow.pages:
  1. Research this specific page topic
  2. Plan component structure
  3. Generate content with full context
  4. Store in database
```

**Benefits:**
- Each page gets dedicated research
- Context flows from previous pages
- Quality over speed
- Easier to debug individual pages

### 4. Layered Context During Generation

**Content creator receives:**

```json
{
  "brand_context": {
    "core_message": "Expert consulting",
    "core_values": ["expertise", "results"],
    "forbidden_phrases": ["revolutionary"]
  },
  
  "flow_context": {
    "audience": "c_suite_executives",
    "current_stage": {
      "name": "consideration",
      "objective": "build_trust",
      "voice_formality": 0.8,
      "technical_depth": 0.6,
      "sales_pressure": 0.4
    }
  },
  
  "page_context": {
    "path": "services/consulting.html",
    "archetype": "service_detail",
    "stage_overrides": {
      "voice_formality": 0.85,
      "data_density": 0.8
    }
  },
  
  "paragraph_context": {
    "position": "section_2_para_3",
    "purpose": "transition_to_results",
    "previous_content": "...what we said before..."
  }
}
```

**How it uses this:**
1. Check brand invariants (never violate core message)
2. Apply flow stage baseline (formality 0.8)
3. Apply page overrides (bump to 0.85)
4. Write paragraph with full context
5. Check coherence with previous paragraphs

### 5. Flow-Aware Navigation

**Navigation component gets:**
```json
{
  "flow": {
    "name": "primary_conversion",
    "pages_in_order": [
      {"path": "/", "title": "Home", "stage": "awareness"},
      {"path": "/services", "title": "Services", "stage": "consideration"},
      {"path": "/case-studies", "title": "Case Studies", "stage": "consideration"},
      {"path": "/contact", "title": "Contact", "stage": "conversion"}
    ]
  }
}
```

**Header renders:**
```html
<nav>
  <a href="/">Home</a>
  <a href="/services">Services</a>
  <a href="/case-studies">Case Studies</a>
  
  <!-- Stage-appropriate CTA -->
  {{if current_stage == "conversion"}}
    <a href="/contact" class="cta-strong">Schedule Consultation</a>
  {{else}}
    <a href="/contact" class="cta-subtle">Get in Touch</a>
  {{end}}
</nav>
```

## Updated Chief Strategist Prompt

```
You are a strategic website architect creating a conversion flow.

Domain: {{.input_data.domain}}
Objective: {{.input_data.objective}}
Target Audience: {{.reviewed_brief.target_audience}}

Your task:
1. Define the PRIMARY AUDIENCE SEGMENT (be specific: not "everyone", but "C-suite executives" or "technical architects")

2. Create a NARRATIVE ARC with 3-4 stages that guides this audience from awareness to conversion:
   - Stage 1: Usually "awareness" or "problem recognition"
   - Stage 2: Usually "consideration" or "solution evaluation"  
   - Stage 3: Usually "conversion" or "decision"
   - Optional Stage 4: "retention" or "advocacy"

3. For EACH STAGE, define:
   - objective (what we want to achieve)
   - voice_formality (0.0 = very casual, 1.0 = very formal)
   - technical_depth (0.0 = no jargon, 1.0 = expert-level)
   - sales_pressure (0.0 = no selling, 1.0 = strong CTA)
   - pacing (engaging/informative/urgent)

4. Design 4-6 PAGES that progress through these stages:
   - Assign each page to a stage
   - Choose appropriate archetypes: corporate_home, service_detail, case_study, testimonial_showcase, contact_form
   - Recommend components for each page
   - Note any context overrides (e.g., case study might need higher data_density)

Return ONLY valid JSON following this structure:
{
  "flow_name": "primary_conversion",
  "audience_segment": "...",
  "audience_description": "...",
  "narrative_arc": {
    "stage_1": {...},
    "stage_2": {...},
    "stage_3": {...}
  },
  "pages": [
    {
      "path": "index.html",
      "title": "...",
      "stage": "stage_1",
      "sequence": 1,
      "archetype": "...",
      "components": [...],
      "context_overrides": {}
    }
  ],
  "entry_points": ["organic_search", "referral"],
  "success_metric": "consultation_booked"
}
```

## Database Queries During Workflow

### Store Brand DNA
```sql
INSERT INTO site_brand_dna (
    orchestration_id,
    domain,
    theme_name,
    core_message,
    core_values,
    voice_parameters
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id;
```

### Store Flow
```sql
INSERT INTO site_flows (
    orchestration_id,
    domain,
    flow_name,
    is_primary,
    audience_segment,
    audience_description,
    narrative_arc,
    entry_points,
    success_metric
) VALUES (
    $1, $2, $3, true, $4, $5, $6, $7, $8
)
RETURNING id;
```

### Store Each Page
```sql
INSERT INTO flow_pages (
    flow_id,
    page_path,
    page_title,
    stage_in_narrative,
    sequence_order,
    page_archetype,
    components,
    context_overrides
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id;
```

### Get Full Context for Page
```sql
SELECT 
    sbd.*,
    sf.narrative_arc,
    sf.narrative_arc->>fp.stage_in_narrative as stage_context,
    fp.context_overrides,
    fp.page_archetype
FROM flow_pages fp
JOIN site_flows sf ON fp.flow_id = sf.id
JOIN site_brand_dna sbd ON sf.orchestration_id = sbd.orchestration_id
WHERE fp.id = $1;
```

## Testing the New Workflow

### 1. Single Flow Test

**Input:**
```json
{
  "domain": "test-consulting.com",
  "objective": "Generate consulting leads",
  "hitl_mode": "auto",
  "model": "AIDA"
}
```

**Expected Behavior:**
1. ✓ Brand DNA created
2. ✓ Single primary flow created with 3 stages
3. ✓ 4-5 pages generated sequentially
4. ✓ Each page has stage-appropriate voice
5. ✓ Navigation reflects flow structure
6. ✓ Deployed to Git

### 2. Two Flow Debug Test

**Input:**
```json
{
  "domain": "test-consulting.com",
  "objective": "Multiple audiences",
  "hitl_mode": "auto",
  "enable_multi_flow": true,
  "audiences": ["executives", "technical_teams"]
}
```

**Expected Behavior:**
1. ✓ Two flows created (executive primary, technical secondary)
2. ✓ Each flow has distinct voice parameters
3. ✓ Shared pages adapt context to flow
4. ✓ Navigation is flow-aware
5. ✓ Brand coherence maintained

### 3. Revert to Single Flow

**After debugging:**
```sql
-- Mark secondary flow as inactive
UPDATE site_flows 
SET is_active = false 
WHERE orchestration_id = $1 
  AND is_primary = false;

-- Keep only primary flow
DELETE FROM flow_pages 
WHERE flow_id IN (
    SELECT id FROM site_flows 
    WHERE is_active = false
);
```

## Migration Path

### For Existing Multipage Sites

```sql
-- 1. Create brand DNA from existing brief data
INSERT INTO site_brand_dna (orchestration_id, domain, ...)
SELECT orchestration_id, domain, ...
FROM orchestration_state
WHERE ...;

-- 2. Create default flow
SELECT create_default_flow(
    orchestration_id,
    domain,
    'primary_audience'
);

-- 3. Migrate existing pages into flow_pages
INSERT INTO flow_pages (flow_id, page_path, ...)
SELECT 
    (SELECT id FROM site_flows WHERE orchestration_id = ...),
    path,
    ...
FROM existing_pages;
```

## Summary

**Old multipage-website-builder:**
- Batch generation (4 pages at a time)
- Global context only
- No narrative structure
- All pages same voice

**New flow-based builder:**
- Sequential generation (one page at a time)
- Layered context (brand → flow → page → paragraph)
- Narrative arc with stages
- Voice adapts to stage

**Result:**
- Higher quality content
- Better user journeys
- Scalable to multiple flows
- Maintains brand coherence