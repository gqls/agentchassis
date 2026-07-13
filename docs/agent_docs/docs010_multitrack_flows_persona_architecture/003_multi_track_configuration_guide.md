# Multi-Track Configuration Guide

## Overview

The multi-track architecture is built but configured for **single-flow by default**. This guide shows:
1. How to configure for 1 flow (production)
2. How to configure for 2 flows (debugging/testing)
3. How to expand when ready

## Single-Flow Configuration (Production)

### Example 1: B2B Consulting Site (Primary Audience: C-Suite)

```json
{
  "domain": "leopardessconsulting.co.uk",
  "objective": "Generate qualified consulting leads from enterprise decision-makers",
  
  "brand_dna": {
    "core_message": "Expert digital transformation consulting with proven results",
    "core_values": ["expertise", "transparency", "results"],
    "theme": "modern-engineering-clean",
    "voice_parameters": {
      "formality_range": [0.6, 0.9],
      "technical_depth_range": [0.4, 0.7],
      "sales_pressure_range": [0.2, 0.8]
    }
  },
  
  "flows": [
    {
      "flow_name": "primary_conversion",
      "is_primary": true,
      "audience_segment": "c_suite_executives",
      "audience_description": "CEOs and senior executives seeking strategic consulting for digital transformation",
      
      "narrative_arc": {
        "stage_1": {
          "name": "awareness",
          "objective": "establish_credibility",
          "voice_formality": 0.7,
          "technical_depth": 0.5,
          "sales_pressure": 0.2,
          "tone": "authoritative but approachable"
        },
        "stage_2": {
          "name": "consideration",
          "objective": "demonstrate_expertise",
          "voice_formality": 0.8,
          "technical_depth": 0.6,
          "sales_pressure": 0.4,
          "tone": "professional and data-driven"
        },
        "stage_3": {
          "name": "conversion",
          "objective": "drive_consultation_request",
          "voice_formality": 0.8,
          "technical_depth": 0.5,
          "sales_pressure": 0.7,
          "tone": "confident and action-oriented"
        }
      },
      
      "pages": [
        {
          "path": "index.html",
          "stage": "stage_1",
          "sequence": 1,
          "title": "Home - Strategic Digital Transformation",
          "archetype": "corporate_home",
          "components": ["hero_professional", "value_props", "client_logos", "cta_primary"]
        },
        {
          "path": "services/digital-transformation.html",
          "stage": "stage_2",
          "sequence": 2,
          "title": "Digital Transformation Services",
          "archetype": "service_detail",
          "components": ["service_hero", "methodology", "results_metrics", "cta_secondary"],
          "context_overrides": {
            "data_density": 0.8,
            "technical_depth": 0.7
          }
        },
        {
          "path": "case-studies/fintech-transform.html",
          "stage": "stage_2",
          "sequence": 3,
          "title": "Case Study: FinTech Transformation",
          "archetype": "case_study",
          "components": ["challenge", "solution", "results_table", "client_quote"],
          "context_overrides": {
            "data_density": 0.9,
            "emotional_appeal": 0.4
          }
        },
        {
          "path": "contact.html",
          "stage": "stage_3",
          "sequence": 4,
          "title": "Schedule Consultation",
          "archetype": "high_friction_conversion",
          "components": ["calendar_embed", "qualifying_questions", "trust_signals"],
          "context_overrides": {
            "urgency": 0.6,
            "sales_pressure": 0.8
          }
        }
      ],
      
      "entry_points": ["organic_search", "linkedin", "referral"],
      "success_metric": "consultation_booked"
    }
  ]
}
```

### Example 2: AI Framework Site (Primary Audience: Technical Decision-Makers)

```json
{
  "domain": "ai-agent-orchestration.com",
  "objective": "Sell AI orchestration framework as a service to engineering teams",
  
  "brand_dna": {
    "core_message": "The Fractal Workforce: Stop Building Pipelines, Start Building Organizations",
    "core_values": ["architectural_excellence", "pragmatism", "transparency"],
    "theme": "modern-engineering-clean",
    "voice_parameters": {
      "formality_range": [0.5, 0.9],
      "technical_depth_range": [0.6, 0.95],
      "sales_pressure_range": [0.1, 0.6]
    }
  },
  
  "flows": [
    {
      "flow_name": "primary_technical_conversion",
      "is_primary": true,
      "audience_segment": "ctos_architects",
      "audience_description": "CTOs and senior architects evaluating agent orchestration solutions",
      
      "narrative_arc": {
        "stage_1": {
          "name": "technical_awareness",
          "objective": "demonstrate_architectural_superiority",
          "voice_formality": 0.6,
          "technical_depth": 0.8,
          "sales_pressure": 0.1,
          "tone": "technical but accessible"
        },
        "stage_2": {
          "name": "deep_evaluation",
          "objective": "prove_scalability_and_reliability",
          "voice_formality": 0.7,
          "technical_depth": 0.9,
          "sales_pressure": 0.3,
          "tone": "engineering-focused"
        },
        "stage_3": {
          "name": "trial_conversion",
          "objective": "start_technical_trial",
          "voice_formality": 0.7,
          "technical_depth": 0.7,
          "sales_pressure": 0.5,
          "tone": "pragmatic and confident"
        }
      },
      
      "pages": [
        {
          "path": "index.html",
          "stage": "stage_1",
          "sequence": 1,
          "archetype": "technical_landing",
          "components": ["hero_technical", "architecture_diagram", "key_differentiators", "cta_docs"]
        },
        {
          "path": "architecture.html",
          "stage": "stage_2",
          "sequence": 2,
          "archetype": "technical_deep_dive",
          "components": ["system_architecture", "code_examples", "scalability_metrics"],
          "context_overrides": {
            "technical_depth": 0.95,
            "code_samples": true
          }
        },
        {
          "path": "use-cases.html",
          "stage": "stage_2",
          "sequence": 3,
          "archetype": "application_showcase",
          "components": ["industry_examples", "implementation_patterns", "results_data"]
        },
        {
          "path": "start-trial.html",
          "stage": "stage_3",
          "sequence": 4,
          "archetype": "low_friction_technical_signup",
          "components": ["quick_start_guide", "api_key_generator", "documentation_links"],
          "context_overrides": {
            "sales_pressure": 0.4,
            "technical_support": 0.9
          }
        }
      ],
      
      "entry_points": ["technical_blog", "github", "hacker_news"],
      "success_metric": "trial_started"
    }
  ]
}
```

## Two-Flow Configuration (Debugging/Testing)

When you want to test multi-flow behavior, add a second flow:

```json
{
  "domain": "leopardessconsulting.co.uk",
  "objective": "Multiple audience targeting",
  
  "brand_dna": {
    "core_message": "Expert digital transformation consulting",
    "core_values": ["expertise", "transparency", "results"],
    "theme": "modern-engineering-clean",
    "voice_parameters": {
      "formality_range": [0.4, 0.9],
      "technical_depth_range": [0.3, 0.8]
    }
  },
  
  "flows": [
    {
      "flow_name": "executive_conversion",
      "is_primary": true,
      "audience_segment": "c_suite",
      
      "narrative_arc": {
        "stage_1": {"name": "awareness", "voice_formality": 0.7, "technical_depth": 0.4},
        "stage_2": {"name": "consideration", "voice_formality": 0.8, "technical_depth": 0.5},
        "stage_3": {"name": "conversion", "voice_formality": 0.8, "technical_depth": 0.4}
      },
      
      "pages": [
        {"path": "index.html", "stage": "stage_1", "sequence": 1},
        {"path": "services.html", "stage": "stage_2", "sequence": 2},
        {"path": "contact.html", "stage": "stage_3", "sequence": 3}
      ]
    },
    
    {
      "flow_name": "thought_leadership",
      "is_primary": false,
      "audience_segment": "industry_professionals",
      
      "narrative_arc": {
        "stage_1": {"name": "education", "voice_formality": 0.6, "technical_depth": 0.7},
        "stage_2": {"name": "relationship", "voice_formality": 0.6, "technical_depth": 0.6},
        "stage_3": {"name": "newsletter_signup", "voice_formality": 0.5, "technical_depth": 0.5}
      },
      
      "pages": [
        {"path": "insights/ai-agents.html", "stage": "stage_1", "sequence": 1},
        {"path": "insights/archive.html", "stage": "stage_2", "sequence": 2},
        {"path": "newsletter.html", "stage": "stage_3", "sequence": 3}
      ]
    }
  ]
}
```

### What to Debug with Two Flows

1. **Context Isolation:** Does each flow maintain its own voice?
2. **Shared Pages:** How does `/about` behave when accessed from different flows?
3. **Navigation:** Does flow-aware navigation work correctly?
4. **Brand Coherence:** Do both flows respect brand DNA invariants?

### After Debugging

Once satisfied, **revert to single flow** for production:
- Keep the primary flow
- Delete or deactivate the secondary flow
- Maintain the multi-flow architecture (dormant but ready)

## How Agents Use This

### Chief Strategist

**Input:**
```json
{
  "domain": "example.com",
  "objective": "...",
  "target_audience": "c_suite_executives",
  "enable_multi_flow": false
}
```

**Prompt (simplified for single-flow):**
```
You are creating a website for {{.domain}} with objective {{.objective}}.

Target audience: {{.target_audience}}

Create a SINGLE primary conversion flow with:
1. Identify the narrative arc (3-4 stages from awareness to conversion)
2. For each stage, define voice parameters:
   - voice_formality (0-1 scale)
   - technical_depth (0-1 scale)  
   - sales_pressure (0-1 scale)
3. Design 4-6 pages that progress through the stages
4. Specify page archetypes and components

Return JSON following the schema.
```

**Output:** Single flow configuration

### Site Architect

**How it uses flows:**

```go
// 1. Get primary flow for this site
flow := getFlowForPage(pageID)

// 2. Get context for current page (merges flow + page overrides)
pageContext := getPageContext(pageID)

// 3. Build components with layered context
context := {
    brand: getBrandDNA(orchestrationID),
    flow: flow,
    page: pageContext,
}

// 4. Pass context down to components during rendering
html := renderComponentTree(components, context)
```

### Content Creator

**How it generates paragraphs:**

```go
// For each paragraph in page
for _, paragraph := range page.Paragraphs {
    
    // Build full context stack
    context := {
        // Layer 1: Brand (immutable)
        brand: {
            coreMessage: "Expert consulting",
            coreValues: ["expertise", "results"],
        },
        
        // Layer 2: Flow stage
        flowStage: {
            name: "consideration",
            objective: "build_trust",
            voiceFormality: 0.8,
            technicalDepth: 0.6,
        },
        
        // Layer 3: Page overrides
        pageContext: {
            voiceFormality: 0.85, // slightly higher than flow
            dataDensity: 0.7,
        },
        
        // Layer 4: Paragraph role
        paragraphRole: {
            position: "section_2_para_3",
            purpose: "transition_to_proof",
            precedingContent: "...summary of previous para...",
        },
    }
    
    // Generate with full context
    content := generateParagraph(context)
    
    // Evaluate before accepting
    if !evaluateCoherence(content, context) {
        retry()
    }
}
```

## Expansion Path

### When to Add Second Flow

Add a second flow when:
1. You have a clearly distinct audience segment
2. The voice/tone needs to shift significantly
3. The conversion goal is fundamentally different
4. Analytics show different user paths

### How to Add

1. **Define the new flow:**
```sql
SELECT create_default_flow(
    'orchestration-id',
    'example.com',
    'technical_evaluators',
    'technical_deep_dive_flow'
);
```

2. **Update `is_primary` flags:**
```sql
UPDATE site_flows 
SET is_primary = false 
WHERE flow_name = 'technical_deep_dive_flow';
-- Keep one primary flow
```

3. **Add flow-specific pages:**
```sql
INSERT INTO flow_pages (flow_id, page_path, stage_in_narrative, sequence_order, ...)
VALUES (...);
```

4. **Update strategist prompt** to generate multiple flows

## Testing Checklist

### Single-Flow Testing
- [ ] Flow created successfully
- [ ] Pages assigned to correct stages
- [ ] Context overrides working
- [ ] Voice parameters applied correctly
- [ ] Brand DNA respected
- [ ] Navigation links correct

### Two-Flow Testing
- [ ] Both flows independent
- [ ] Context isolation working
- [ ] Shared components adapt to flow
- [ ] Navigation flow-aware
- [ ] Brand coherence across flows
- [ ] No context bleeding between flows

### Before Production
- [ ] Revert to single primary flow
- [ ] Deactivate secondary flow
- [ ] Clean up test data
- [ ] Document flow configuration
- [ ] Ready for future expansion

## Summary

**Current State:**
- Full multi-flow architecture built
- Configured for single-flow by default
- Ready to scale when needed

**Production Usage:**
- Create one primary flow per site
- Define 3-4 narrative stages
- Assign pages to stages
- Apply context overrides as needed

**Future Expansion:**
- Architecture supports N flows
- Add flows as audience segments emerge
- Test with 2 flows, then scale
- Maintain brand coherence via DNA

The structure is there, but we keep it simple.