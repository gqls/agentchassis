# Multi-Track Architecture: Implementation Summary

## What We've Built

A **flow-based website architecture** that supports sophisticated multi-audience journeys while defaulting to simple single-flow operation.

### Core Innovation

Instead of treating websites as flat collections of pages, we model them as **choreographed narratives** where:

- **Brand DNA** provides immutable identity
- **Flows** define audience-specific journeys
- **Stages** progress users from awareness to conversion
- **Context** adapts at every level (brand → flow → page → paragraph)

## The Architecture

### Database Schema (Ready to Deploy)

**File:** `multi_track_schema.sql`

**Tables:**
1. **site_brand_dna** - Immutable brand identity
2. **site_flows** - User journeys/narratives
3. **flow_pages** - Pages with stage-specific context
4. **page_transitions** - How users move through flows

**Key Features:**
- Supports 1-N flows (designed for single-flow use)
- Layered context inheritance with overrides
- Helper functions for common queries
- Migration utilities for existing sites

### Configuration (Documented)

**File:** `multi_track_configuration_guide.md`

**Configurations:**
- Single-flow (production)
- Two-flow (debugging)
- Expansion path

**Examples:**
- B2B consulting site (C-suite audience)
- AI framework site (technical audience)
- Multi-flow testing setup

### Workflow Integration (Specified)

**File:** `multipage_workflow_with_flows.md`

**Changes:**
1. Create brand DNA first
2. Chief strategist generates flow
3. Sequential page generation (not batches)
4. Layered context during content creation
5. Flow-aware navigation

## What This Solves

### Problem 1: Samey Content
**Before:** All pages sound the same (global context only)
**After:** Voice adapts to journey stage (awareness casual, conversion formal)

### Problem 2: Multiple Audiences
**Before:** One voice tries to serve everyone (pleases nobody)
**After:** Different flows for different audiences (each optimized)

### Problem 3: No Narrative Arc
**Before:** Random page collection
**After:** Choreographed journey with clear progression

### Problem 4: Inconsistent Branding
**Before:** Voice varies unpredictably
**After:** Variance within defined bounds (brand DNA enforces invariants)

## Production Strategy

### Phase 1: Single-Flow Sites (Now)

**Objective:** Build high-quality sites with narrative structure

**Configuration:**
- One primary flow per site
- 3-4 narrative stages
- 4-6 pages in sequence
- Stage-appropriate voice

**Example:**
```
Consulting site → Executive audience
  Stage 1 (awareness): Home page - professional but approachable
  Stage 2 (consideration): Services + Case study - data-driven
  Stage 3 (conversion): Contact - formal and action-oriented
```

### Phase 2: Multi-Flow Testing (Optional)

**Objective:** Debug multi-flow behavior

**Configuration:**
- Add secondary flow to existing site
- Test context isolation
- Validate brand coherence
- Verify navigation works

**Then:** Revert to single flow, keep architecture

### Phase 3: Production Multi-Flow (Future)

**Objective:** Serve multiple audiences per site

**When to use:**
- Distinct audience segments (execs vs engineers)
- Different conversion goals (newsletter vs demo)
- Significant voice shifts needed

**Criteria:**
- Analytics show different user paths
- Audience pain points fundamentally different
- Single flow underperforming for some users

## Implementation Steps

### Step 1: Deploy Database Schema

```bash
# Run the schema migration
psql $DATABASE_URL -f multi_track_schema.sql

# Verify tables created
psql $DATABASE_URL -c "\dt site_*"
psql $DATABASE_URL -c "\dt flow_*"
```

**Expected output:**
- site_brand_dna
- site_flows
- flow_pages
- page_transitions

### Step 2: Update Chief Strategist

**Current prompt:** Generates flat page list
**New prompt:** Generates flow with narrative arc

**Changes:**
1. Add audience segment analysis
2. Define narrative stages
3. Set voice parameters per stage
4. Sequence pages through stages
5. Return flow JSON (not just page list)

**File to update:**
```sql
UPDATE agent_definitions 
SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,generate_build_plan,config,prompt_template}',
  '"<new prompt from multipage_workflow_with_flows.md>"'::jsonb
)
WHERE type = 'chief-strategist';
```

### Step 3: Update Multipage Builder Workflow

**Current workflow:** Batch generation
**New workflow:** Sequential with flow context

**Changes:**
1. Add `create_brand_dna` step
2. Add `create_primary_flow` step
3. Replace batches with loop
4. Add research step per page
5. Pass layered context to content creator

**File to update:**
```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
  default_config,
  '{workflow}',
  '<new workflow from multipage_workflow_with_flows.md>'::jsonb
)
WHERE type = 'multipage-website-builder';
```

### Step 4: Update Content Creator

**Add context layer processing:**

```go
// In content creator action
type LayeredContext struct {
    Brand BrandDNA
    FlowStage FlowStageContext
    PageOverrides map[string]interface{}
    ParagraphRole ParagraphContext
}

func GenerateWithContext(ctx LayeredContext) string {
    // 1. Check brand invariants
    if violatesInvariants(ctx.Brand) {
        return error
    }
    
    // 2. Merge context layers
    effectiveContext := mergeContext(
        ctx.FlowStage,
        ctx.PageOverrides,
    )
    
    // 3. Generate with merged context
    content := callLLM(effectiveContext)
    
    // 4. Evaluate coherence
    if !checkCoherence(content, ctx) {
        retry()
    }
    
    return content
}
```

### Step 5: Test End-to-End

**Test case 1: Simple consulting site**
```json
{
  "domain": "test-consulting.com",
  "objective": "Generate executive consulting leads",
  "target_audience": "c_suite_executives"
}
```

**Expected:**
1. Brand DNA created with theme and core message
2. Flow created: primary_conversion (3 stages)
3. Pages generated: home → services → case-study → contact
4. Voice progresses: 0.7 → 0.8 → 0.85 formality
5. Navigation links all pages in flow order
6. Deployed to Git

**Test case 2: Technical product site**
```json
{
  "domain": "test-framework.com",
  "objective": "Convert technical evaluators to trial users",
  "target_audience": "ctos_and_architects"
}
```

**Expected:**
1. Flow with high technical depth (0.8-0.95)
2. Pages: landing → architecture → use-cases → trial
3. Voice: technical but accessible
4. Code examples in appropriate pages
5. Low sales pressure (0.1-0.4)

## Quality Assurance

### Voice Consistency Checks

```sql
-- Check voice parameter ranges respect brand DNA
SELECT 
    fp.page_path,
    sf.narrative_arc->>fp.stage_in_narrative as stage_context,
    fp.context_overrides,
    sbd.voice_parameters
FROM flow_pages fp
JOIN site_flows sf ON fp.flow_id = sf.id
JOIN site_brand_dna sbd ON sf.orchestration_id = sbd.orchestration_id
WHERE fp.flow_id = '<flow-id>';
```

**Validate:**
- Formality within allowed range
- Technical depth appropriate
- Sales pressure reasonable
- No forbidden phrases in content

### Flow Coherence Checks

```sql
-- Check pages flow in logical sequence
SELECT 
    sequence_order,
    page_path,
    stage_in_narrative,
    context_overrides->>'voice_formality' as formality
FROM flow_pages
WHERE flow_id = '<flow-id>'
ORDER BY sequence_order;
```

**Validate:**
- Sequence makes sense (no jumps)
- Stages progress logically
- Voice evolves smoothly
- CTAs appropriate to stage

### Brand Coherence Check

```sql
-- Verify all flows respect brand DNA
SELECT 
    sf.flow_name,
    sbd.core_message,
    sbd.voice_parameters
FROM site_flows sf
JOIN site_brand_dna sbd ON sf.orchestration_id = sbd.orchestration_id
WHERE sf.orchestration_id = '<orchestration-id>';
```

**Validate:**
- All flows share core message
- Voice variance within bounds
- Visual theme consistent

## Rollout Plan

### Week 1: Infrastructure
- [ ] Deploy database schema
- [ ] Test schema with manual inserts
- [ ] Verify helper functions work
- [ ] Document any issues

### Week 2: Agent Updates
- [ ] Update chief-strategist prompt
- [ ] Update multipage-builder workflow
- [ ] Update content-creator context handling
- [ ] Test individual agents

### Week 3: Integration Testing
- [ ] Run end-to-end test (consulting site)
- [ ] Run end-to-end test (technical site)
- [ ] Validate voice progression
- [ ] Check brand coherence

### Week 4: Multi-Flow Debug
- [ ] Create two-flow test site
- [ ] Verify context isolation
- [ ] Test flow-aware navigation
- [ ] Validate no context bleeding

### Week 5: Production
- [ ] Revert to single-flow config
- [ ] Deploy to production
- [ ] Monitor first real sites
- [ ] Document learnings

## Success Metrics

### Qualitative
- Content sounds stage-appropriate
- Voice progresses logically
- Brand feels coherent
- Narratives guide users

### Quantitative
- Pages generate without errors
- Context overrides apply correctly
- Database queries performant
- Navigation links correct

### Readiness
- Schema deployed
- Agents updated
- Tests passing
- Documentation complete

## Future Enhancements

### Pattern Library Integration

When interrogating successful sites:
```json
{
  "pattern_id": "finance_trust_builder",
  "flow_stage": "consideration",
  "audience": "institutional_investors",
  "voice_parameters": {
    "formality": 0.95,
    "technical_depth": 0.7,
    "data_density": 0.9
  },
  "components": ["regulatory_certs", "audit_statement", "data_table"],
  "conversion_lift": "+34%"
}
```

**Usage:** Query patterns by stage and audience, apply to similar flows

### Analytics Integration

Track which flows convert better:
```sql
CREATE TABLE flow_performance (
    flow_id UUID,
    metric_name TEXT,
    metric_value DECIMAL,
    recorded_at TIMESTAMP
);
```

**Optimize:** Adjust stage parameters based on what works

### A/B Testing

Test different narrative arcs:
```sql
-- Flow A: 3 stages (fast conversion)
-- Flow B: 4 stages (relationship building)
-- Measure which converts better
```

## Files Reference

| File | Purpose |
|------|---------|
| `multi_track_schema.sql` | Database tables and functions |
| `multi_track_configuration_guide.md` | Config examples and patterns |
| `multipage_workflow_with_flows.md` | Updated workflow specification |
| `multi_track_sitemap_architecture.md` | Deep architectural analysis |
| `multipage_builder_evolution_discussion.md` | Original discussion and vision |

## Summary

**What we have:**
- Full multi-flow architecture (database + patterns)
- Configuration for 1-N flows
- Updated workflow specification
- Quality assurance approach

**What we'll use:**
- Single primary flow per site (production)
- Layered context (brand → flow → page)
- Narrative arc with stages
- Voice progression

**What we can expand to:**
- Multiple flows per site
- Different audiences
- A/B testing
- Pattern library integration

**The key insight:**
Stop thinking "pages" → Start thinking "journeys"

The architecture is built for complexity but configured for simplicity.