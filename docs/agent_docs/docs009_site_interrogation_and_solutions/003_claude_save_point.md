# Multipage Website Builder Evolution Discussion

## Current Date: 2025-12-13

## Context
Discussion on evolving the multipage-website-builder from batch generation to granular, intentional page-by-page creation with research and planning at every level.

## Vision Summary

### 1. Granular Content Creation
- Break pages into smallest meaningful units
- Each paragraph researched and planned
- Understanding what each element achieves
- Eventually: research successful sites, extract patterns (not copy)

### 2. Pattern Library Approach
- "Interrogate" successful sites to extract:
    - Design patterns and visual hierarchy
    - Content structure (how arguments build)
    - Psychological techniques
    - Conversion mechanics
- Store extracted patterns
- Create original content using successful strategies

### 3. Site Archetypes & Patterns
Different site types need different approaches:
- **Skyscraper Sales:** Long-form, multiple closing statements
- **Brochure:** Well-researched, clear funnels, authoritative
- **Content/Blog:** Reader retention, news synthesis
- **Portfolio:** Visual storytelling

### 4. **CRITICAL CHALLENGE: Consistency Within Complexity**

**The Problem:**
- Need site-wide consistency (brand, voice, design system)
- Need page-level nuance (different purposes, structures)
- "Constant context" mechanism required

**User's New Insight:**
> "The Site itself may have a narrative or flow - maybe several different types of funnel, so we could probably have a site map entity that also is changeable and has variations big or small."

**Implications:**
A site is not just a flat collection of pages with global context. It's a **multi-track system** where:
- Different user journeys exist simultaneously
- Different funnels operate in parallel
- Different narratives for different audience segments
- The sitemap itself should be flexible and variable

**Example Scenarios:**
1. **B2B SaaS Site:**
    - Track A: Technical documentation flow (engineers)
    - Track B: ROI/business case flow (executives)
    - Track C: Implementation flow (IT managers)
    - Each track has different voice, different CTAs, different narrative

2. **E-commerce Site:**
    - Track A: Browse → Category → Product → Cart (explorers)
    - Track B: Search → Product → Cart (buyers)
    - Track C: Blog → Related Product → Cart (content readers)

3. **Consulting Site:**
    - Track A: Problem awareness → Service pages → Contact
    - Track B: Blog/insights → Authority building → Newsletter
    - Track C: Case studies → Proof → RFP

**What This Means for Architecture:**

Instead of:
```
Site (global context) → Pages (flat list)
```

We need:
```
Site (base brand) → Flows/Tracks (narratives) → Pages (in journey) → Sections
```

### 5. Semantic Labeling (Future)
- Every element identifiable
- Natural language editing: "edit paragraph left of blue CTA"
- Structured markup with semantic IDs

### 6. Technical Patterns

**What's Working (landing-page-builder):**
- Spawns separate agents per stage
- Natural initialization delays
- Clear separation of concerns

**Current Issues:**
- Deep CollectedData nesting
- Complexity in field extraction
- Every spawn creates new overhead

**Opportunities:**
- Agent reuse for similar tasks
- Clearer data contracts
- Explicit context passing

## Architectural Decisions Discussed

### A. Page Generation Approach
Options considered:
1. One agent per page
2. One agent per page type
3. Stages per page (research → structure → content)
4. **Hybrid:** Shared context agent + per-page agents

### B. Context Management
How to maintain "constant context":
1. Shared context document (all agents receive same guide)
2. Parent orchestrator holds context (passes to children)
3. Context agent (dedicated query service)
4. **Database-backed context** (store, agents fetch)

### C. Data Structure Strategy
Managing deep nesting:
1. Accept nesting with robust extraction
2. Flatten at each level
3. Explicit contracts (output_field standardized)
4. **Hybrid:** Critical data flattened, raw responses nested

### D. Incremental Evolution
Recommended path:
1. **Phase 1:** Single page done deeply (with research/planning)
2. **Phase 2:** Extend to multiple pages with shared context
3. **Phase 3:** Add pattern library
4. **Phase 4:** Add semantic labeling

## From Extended Discussion Document

### Key Architectural Principles

**1. Everything is a Component (Recursive Containers)**
- No separate "is_container" flag needed
- Components detected by presence of `{{.Slot}}` placeholders
- Infinite nesting naturally supported

**2. Asset Deduplication via Bubble-Up**
- Children return assets to parents
- Merged at each level
- Final unique list injected once in `<head>`

**3. Unique Addressing for Editing**
- Every element gets `data-uuid` and `data-path`
- Enables spatial queries: "3rd paragraph on the left"
- Supports future natural language editing

**4. Simplified Component Schema**
```sql
content_components:
  - html_template (with {{.Slot}} and {{.Data}} placeholders)
  - defined_slots (array of slot names)
  - data_schema (what data this component needs)
  - wrapper_tag (NULL for "ghost" components to reduce nesting)
```

**5. Navigation via Global Context**
- Sitemap passed down to all components
- Header/Footer can dynamically link to whatever pages exist
- No hardcoding

**6. Atomic Content Generation**
- Paragraph-by-paragraph LLM calls
- Context flows from previous paragraphs
- Evaluator agent checks each paragraph for:
    - Hallucination (fact-checking)
    - Brand alignment
    - Logical flow

## Critical Open Questions

### 1. Multi-Track Site Architecture
**Question:** How do we model a site with multiple simultaneous funnels?

**Proposed Solution (needs discussion):**
```json
{
  "site": {
    "brand_identity": {...},
    "flows": [
      {
        "name": "technical_track",
        "audience": "engineers",
        "voice": "technical, precise",
        "pages": ["docs/index", "docs/api", "docs/sdk"],
        "narrative": "education → implementation → support"
      },
      {
        "name": "business_track",
        "audience": "executives",
        "voice": "strategic, ROI-focused",
        "pages": ["solutions", "roi-calculator", "enterprise"],
        "narrative": "problem → solution → proof → contact"
      }
    ]
  }
}
```

**Implications:**
- Components need "flow-aware" context, not just global
- Same page might exist in multiple flows with different voice
- Navigation needs to be flow-aware (different menus per track?)

### 2. Context Propagation Complexity
If we have:
- Global brand context (site-wide)
- Flow context (track-specific)
- Page context (this page's objective)
- Section context (this section's purpose)
- Paragraph context (this paragraph's goal)

How do we prevent context explosion? How does a content generator know which context layers to use?

### 3. Pattern Library Integration
When we "interrogate" successful sites:
- How do we classify which flow/track a pattern belongs to?
- How do we know when to use "long-form storytelling" vs "data-heavy comparison"?
- Who decides which patterns to use where?

### 4. Coherence vs Variation
Maintaining brand while varying:
- Voice/tone per audience
- Structure per funnel
- Pacing per journey stage

What are the invariants (must stay consistent) vs variables (should change)?

## Next Steps (When Ready)

1. Define sitemap entity structure (flows, tracks, journeys)
2. Model context layers and propagation rules
3. Design flow-aware content generation
4. Create pattern classification system
5. Build evaluator/coherence checker

## Related Files
- spawn_actions.go (two-delay fix for initialization)
- spawn_agent_two_delay_fix.md
- add_builder_questionnaires.sql
- BUILDER_QUESTIONNAIRES_EXPLAINED.md

## Status
- Spawn timing: FIXED (two delays)
- Questionnaires: SQL READY (not applied)
- Multi-page architecture: DESIGN PHASE (this discussion)