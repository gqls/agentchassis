# Critical Analysis: Multi-Track Sitemap Architecture

## The Core Problem

Current thinking treats a site as:
```
Site (global context) → List of Pages → Sections → Components
```

**This is insufficient** because:

1. **Real sites are journey-based**, not page-based
2. **Different audiences enter and navigate differently**
3. **Context changes along the journey**, not just globally
4. **Multiple narratives operate simultaneously**

## The Reality of Modern Sites

### Example 1: B2B SaaS (Consulting Framework)

**Audience Segments:**
- CTOs (technical evaluation)
- CFOs (ROI/business case)
- Procurement (vendor comparison)

**Current Broken Model:**
- Single "about" page tries to serve all three
- Homepage hero has generic message
- Voice is "middle ground" (pleases nobody)

**What Actually Happens:**
```
CTO Journey:
  Entry: Technical blog post
  → Documentation preview
  → API reference
  → GitHub integration demo
  → Technical trial signup
  
  Voice: Technical, precise, assumes knowledge
  CTAs: "Try the SDK", "View API docs"

CFO Journey:
  Entry: LinkedIn ad about cost savings
  → ROI calculator
  → Case study with metrics
  → Pricing comparison
  → Enterprise contact form
  
  Voice: Business-focused, ROI-driven
  CTAs: "Calculate savings", "Schedule demo"

Procurement Journey:
  Entry: G2 review link
  → Security whitepaper
  → Compliance certifications
  → Vendor comparison table
  → RFP submission form
  
  Voice: Formal, risk-mitigation
  CTAs: "Download security docs", "Submit RFP"
```

**Key Insight:** Same company, three completely different websites experienced by three different users.

### Example 2: Content Site (Leopardess Consulting)

**The User's Objective:** Create comprehensive corporate website

**Surface-Level Pages:**
- Company overview
- Leadership
- Services
- Blog
- Careers
- Contact

**But Actually Multiple Funnels:**

**Funnel A: Thought Leadership → Newsletter**
```
Blog post (SEO entry)
  → Related insights
  → Author profile (credibility)
  → Newsletter signup (low friction)
  
Voice: Educational, generous with knowledge
Goal: Build trust, long-term relationship
```

**Funnel B: Service Interest → Sales Contact**
```
Service page (direct traffic)
  → Case study with results
  → Process/methodology explanation
  → Calendar booking (high friction)
  
Voice: Professional, results-focused
Goal: Qualified lead generation
```

**Funnel C: Partnership/Hiring → Application**
```
Careers page
  → Culture/values
  → Team profiles
  → Application form
  
Voice: Authentic, culture-first
Goal: Talent acquisition
```

**The Problem:** If we treat this as "global context + pages", we lose the narrative arc of each funnel.

## Proposed Architecture: The Multi-Track Model

### 1. Sitemap as Directed Graph

Instead of flat page list, model as graph:

```json
{
  "site": {
    "domain": "leopardessconsulting.co.uk",
    "brand_dna": {
      "visual_identity": "modern-engineering-clean",
      "core_values": ["expertise", "transparency", "results"],
      "never_changes": {
        "logo": "...",
        "primary_color": "#0f172a",
        "typography": "Inter"
      }
    },
    
    "flows": [
      {
        "id": "thought_leadership_track",
        "name": "Thought Leadership → Newsletter",
        "audience": "industry_professionals",
        "entry_points": ["blog_post", "linkedin_share"],
        
        "narrative_arc": {
          "stage_1": {
            "objective": "establish_expertise",
            "voice": "educational, generous",
            "pacing": "deep, detailed",
            "tone": "collegial"
          },
          "stage_2": {
            "objective": "build_relationship",
            "voice": "warm, personal",
            "pacing": "moderate",
            "tone": "conversational"
          },
          "stage_3": {
            "objective": "low_friction_conversion",
            "voice": "inviting, non-pushy",
            "pacing": "quick",
            "tone": "friendly"
          }
        },
        
        "pages": [
          {
            "path": "insights/ai-agents-enterprise",
            "stage": "stage_1",
            "components": ["blog_header", "long_form_content", "author_bio"],
            "context_override": {
              "voice_formality": 0.7,
              "technical_depth": 0.9,
              "sales_pressure": 0.1
            }
          },
          {
            "path": "insights/related",
            "stage": "stage_2",
            "components": ["article_grid", "topic_tags"]
          },
          {
            "path": "newsletter",
            "stage": "stage_3",
            "components": ["value_proposition", "simple_form"],
            "context_override": {
              "voice_formality": 0.5,
              "urgency": 0.3
            }
          }
        ]
      },
      
      {
        "id": "service_conversion_track",
        "name": "Service Interest → Sales Call",
        "audience": "decision_makers",
        "entry_points": ["service_page", "google_search"],
        
        "narrative_arc": {
          "stage_1": "problem_identification",
          "stage_2": "solution_presentation",
          "stage_3": "proof_and_trust",
          "stage_4": "high_friction_conversion"
        },
        
        "pages": [
          {
            "path": "services/digital-transformation",
            "stage": "stage_1_and_2",
            "components": ["problem_hero", "solution_framework", "service_details"],
            "context_override": {
              "voice_formality": 0.8,
              "technical_depth": 0.6,
              "sales_pressure": 0.4
            }
          },
          {
            "path": "case-studies/fintech-transformation",
            "stage": "stage_3",
            "components": ["results_metrics", "client_quote", "process_breakdown"],
            "context_override": {
              "voice_formality": 0.9,
              "data_density": 0.8,
              "emotional_appeal": 0.3
            }
          },
          {
            "path": "contact",
            "stage": "stage_4",
            "components": ["calendar_embed", "qualifying_questions"],
            "context_override": {
              "voice_formality": 0.9,
              "urgency": 0.6
            }
          }
        ]
      }
    ]
  }
}
```

### 2. Layered Context Model

**Problem:** Content generators receive too much or too little context.

**Solution:** Hierarchical context inheritance with overrides.

```
┌─────────────────────────────────────┐
│ SITE LAYER (Immutable Brand DNA)   │
│ - Visual identity                   │
│ - Core values                       │
│ - Company facts                     │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│ FLOW LAYER (Journey Narrative)     │
│ - Target audience                   │
│ - Narrative arc stages              │
│ - Overall conversion goal           │
│ - Voice/tone baseline               │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│ PAGE LAYER (Specific Objective)    │
│ - Stage in narrative                │
│ - Specific conversion goal          │
│ - Context overrides (formality++)   │
│ - Component selection               │
└─────────────────┬───────────────────┘
                  │
                  ▼
┌─────────────────────────────────────┐
│ COMPONENT LAYER (Tactical)         │
│ - Paragraph-level objectives        │
│ - Specific data to display          │
│ - Micro-interactions                │
└─────────────────────────────────────┘
```

**How Content Generator Uses This:**

When writing a paragraph for `/case-studies/fintech-transformation`:

1. **Inherits from Site:** Brand values, company name, visual system
2. **Inherits from Flow:** This is "Service Conversion Track", stage 3 (proof)
3. **Inherits from Page:** High formality (0.9), data-heavy (0.8), low emotional (0.3)
4. **Component Context:** This specific paragraph is "results_metrics" → use numbers, cite specific improvements

**Result:** Content is coherent with brand but appropriately tuned for the specific user journey stage.

### 3. Navigation That Understands Flows

**Problem:** Traditional nav is flat. Doesn't understand user journey.

**Solution:** Flow-aware navigation.

```html
<!-- Global Nav (always present) -->
<nav class="global">
  <a href="/">Home</a>
  <a href="/services">Services</a>
  <a href="/about">About</a>
</nav>

<!-- Flow-Specific Nav (contextual) -->
<nav class="flow-contextual" data-flow="thought_leadership_track">
  <!-- Previous in journey -->
  <a href="/insights" class="back">← More Insights</a>
  
  <!-- Next logical step in THIS flow -->
  <a href="/newsletter" class="cta-subtle">Get weekly insights</a>
</nav>

<!-- vs Different Flow -->
<nav class="flow-contextual" data-flow="service_conversion_track">
  <a href="/services" class="back">← All Services</a>
  <a href="/contact" class="cta-strong">Schedule consultation</a>
</nav>
```

**Key:** Same footer component, different CTAs based on which flow user is in.

### 4. Implementation Changes

**Database Schema:**

```sql
CREATE TABLE site_flows (
    id UUID PRIMARY KEY,
    site_id UUID,
    flow_name TEXT,
    audience_segment TEXT,
    narrative_arc JSONB, -- stages with voice/tone parameters
    entry_points TEXT[],
    success_metric TEXT
);

CREATE TABLE flow_pages (
    id UUID PRIMARY KEY,
    flow_id UUID,
    page_path TEXT,
    stage_in_narrative TEXT,
    context_overrides JSONB, -- voice_formality, urgency, technical_depth, etc
    sequence_order INT
);

CREATE TABLE page_transitions (
    from_page_id UUID,
    to_page_id UUID,
    transition_type TEXT, -- 'next_in_flow', 'alternate_path', 'exit_flow'
    weight DECIMAL -- for A/B testing which transition works better
);
```

**Chief Strategist Changes:**

Instead of:
```
"Generate a list of pages for this site"
```

Now:
```
"Given domain {{.domain}} and objective {{.objective}}, 
identify the distinct audience segments and create a 
conversion flow for each. For each flow, define:
- Entry points (how they discover the site)
- Narrative arc (stages they progress through)
- Voice/tone evolution (how messaging changes)
- Exit goals (what conversion looks like)"
```

**Content Creator Changes:**

Instead of:
```
context = {global_brand + page_objective}
```

Now:
```
context = {
  brand_dna: site.brand,
  flow: flow.narrative_arc,
  stage: current_stage,
  overrides: page.context_overrides,
  previous_page: flow.pages[n-1] // what did we just say?
}
```

## Critical Questions to Resolve

### Q1: How Many Flows is Too Many?

**Trade-off:** More flows = better targeting, but more complexity.

**Proposal:**
- **MVP:** 2-3 flows max per site
- **Production:** 5-7 flows (one per major audience segment)
- **Enterprise:** Unlimited (different flows per product line, region, etc.)

**Decision Rule:** Create new flow when:
1. Audience has fundamentally different pain point
2. Conversion goal is different
3. Voice/tone needs to shift significantly

### Q2: Shared Pages Across Flows

What if a page exists in multiple flows but with different context?

**Example:** "About Us" page

- From thought leadership flow: Emphasize expertise, research
- From service conversion flow: Emphasize results, client roster
- From hiring flow: Emphasize culture, team

**Option A: One page, flow-aware components**
```html
<section class="about-intro">
  {{if .flow == "thought_leadership"}}
    <p>Founded by researchers with 20+ years...</p>
  {{else if .flow == "service_conversion"}}
    <p>Trusted by Fortune 500 companies...</p>
  {{else if .flow == "hiring"}}
    <p>A team that believes in...</p>
  {{end}}
</section>
```

**Option B: Separate pages per flow**
```
/about (default)
/about?flow=thought_leadership (variant)
/about?flow=hiring (variant)
```

**Recommendation:** Start with Option B (separate pages). Easier to reason about, clearer analytics.

### Q3: How Do We Maintain Coherence?

With multiple flows, how do we prevent:
- Contradictory messaging across flows
- Brand voice fragmentation
- Visual inconsistency

**Solution: The Brand Invariants**

```json
{
  "brand_dna": {
    "invariants": {
      "core_message": "AI orchestration made pragmatic",
      "visual_system": "modern-engineering-clean",
      "forbidden_phrases": ["cutting-edge", "revolutionary"],
      "required_elements": ["data-driven proof", "transparent pricing"]
    },
    "variance_allowed": {
      "voice_formality": [0.4, 1.0], // can range from casual to formal
      "technical_depth": [0.3, 0.9],
      "sales_pressure": [0.1, 0.8]
    }
  }
}
```

**Evaluator Agent:** Before any content is accepted, check:
1. ✓ Uses allowed vocabulary
2. ✓ Doesn't contradict core message
3. ✓ Stays within variance bounds for voice parameters
4. ✓ Maintains visual system

### Q4: Pattern Library Integration

When we "interrogate" successful sites, we now need to ask:
- "What flow/journey stage is this pattern from?"
- "What audience segment does this serve?"

**Enhanced Pattern Storage:**

```json
{
  "pattern_id": "finance_proof_section",
  "domain": "finance",
  "flow_stage": "trust_building",
  "audience": "institutional_investors",
  "components": ["data_table", "regulatory_certifications", "auditor_statement"],
  "voice_parameters": {
    "formality": 0.95,
    "technical_depth": 0.7,
    "emotional_appeal": 0.2
  },
  "conversion_data": {
    "avg_time_on_page": "4:23",
    "conversion_lift": "+34%",
    "bounce_rate": "12%"
  }
}
```

When building a "proof" stage in a finance flow, we can query:
```sql
SELECT * FROM design_patterns 
WHERE flow_stage = 'trust_building' 
  AND domain = 'finance'
ORDER BY conversion_lift DESC
LIMIT 5
```

## Recommended Implementation Path

### Phase 1: Single Flow (MVP)
- Build one site with one flow start-to-finish
- Validate layered context works
- Test flow-aware navigation
- Prove content quality with stage-appropriate voice

### Phase 2: Multi-Flow
- Add second flow to same site
- Test shared pages with flow variants
- Validate flows don't contradict
- Measure which flow converts better

### Phase 3: Pattern Library
- Interrogate 10 successful sites
- Extract patterns tagged by flow stage
- Store in pattern DB
- Use patterns to inform component selection

### Phase 4: Cross-Site Optimization
- Track which flows work across domains
- Identify universal patterns (work everywhere)
- Identify domain-specific patterns
- Build recommendation engine: "For finance sites in trust-building stage, use X"

## Summary

**Current Model:** Flat sitemap with global context
**Problem:** Doesn't handle multiple audiences, journeys, or narratives

**New Model:** Multi-track flows with layered context
**Benefits:**
- Each audience gets optimized journey
- Brand stays coherent via invariants
- Content quality improves via stage-appropriate voice
- Analytics can track flow performance
- Pattern library becomes domain + stage aware

**Key Change:**
Stop thinking "pages" → Start thinking "journeys"

The sitemap becomes a **choreographed set of narratives** rather than a list of URLs.