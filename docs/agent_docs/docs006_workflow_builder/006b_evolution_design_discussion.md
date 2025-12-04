# Architecture Discussion Snapshot: Evolution, Variants, and Entity State

## Date: Session in progress

## Context

Discussing how to handle agent/group evolution, the relationship between base agents and variants, and where different types of learnings should live.

---

## The Big Picture

Building a **recipe management and improvement system** where:
- Handles diverse tasks (web design, finance, algorithms, calculators)
- Improves over time
- Doesn't randomly mutate
- Has controlled evolution with oversight

---

## The Fragility Problem

**Override model risk:**
```
base: landing-page-builder
  └── variant: landing-page-builder-saas (overrides hero section prompt)
        └── variant: landing-page-builder-saas-b2b (overrides CTA language)
```

If you change the base, all variants inherit the change - could fix everything or break everything.

**Snapshot model (preferred):**
```
landing-page-builder-v1 (frozen snapshot)
  └── landing-page-builder-saas-v1 (based on v1)

landing-page-builder-v2 (new base, breaking changes)
  └── landing-page-builder-saas-v2 (migrated to v2)
```

Variants explicitly reference a snapshot version. Base can evolve without breaking existing variants.

---

## Three Types of Evolution

| Type | Scope | Trigger | Oversight |
|------|-------|---------|-----------|
| **Bug fix** | Base agent | Developer finds issue | Low - just fix it |
| **Improvement** | Variant or base | Metrics suggest | Medium - test first |
| **Innovation** | New variant or base | New capability/task type | High - HITL review |

Not all evolution is equal. Random mutation isn't desirable. Controlled improvement is.

---

## What Triggers Evolution?

**Metrics-driven (automatic candidates):**
- Conversion rate dropped 20% after last deploy
- Time-to-complete increased
- Error rate spiked

**Observation-driven (agent suggests):**
- "This component library lacks a pricing table"
- "B2B sites consistently need different hero copy"

**Human-driven:**
- "We should add a review step for legal content"
- "Try the new Claude model for content writing"

The system **proposes** changes but requires HITL **approval** before applying.

---

## Where Learnings Live: A Four-Level Model

```
Level 1: Base Agents (agent_definitions)
  - Stable templates
  - Versioned and snapshotted
  - Changes are deliberate, tested, approved
  - Example: "content-writer v3"

Level 2: Task Variants (agent_variants)
  - Specializations of base agents
  - Reference a specific base version
  - Override prompts, parameters, sub-workflows
  - Example: "content-writer-legal v2 (based on content-writer v3)"

Level 3: Entity Learnings (entity_state_log)
  - What we know about THIS domain/project/client
  - Read at runtime, influences behaviour
  - Doesn't change the agent definition
  - Example: "ai-orchestration.com prefers technical tone"

Level 4: Proposed Improvements (improvement_proposals)
  - Suggestions from metrics or observation
  - Awaiting HITL review
  - Can be applied to Level 1, 2, or 3
  - Example: "Suggest: increase social proof emphasis for B2B SaaS"
```

---

## Dynamic vs Periodic Entity Updates

**Hybrid approach (recommended):**

- **Facts** (dynamic, safe to read) - "This domain is for B2B"
- **Strategy** (periodic, needs review) - "B2B sites should use authority positioning"

Agent reads entity_state for facts, but workflow/prompt changes require approval.

---

## Proposed Schema

```sql
-- Stable base definitions (versioned, snapshotted)
agent_definitions
  - id, type, version
  - default_config (workflow, prompts)
  - is_snapshot (boolean - frozen or living?)
  - domain (web, finance, etc.)

-- Task specializations (reference base version)
agent_variants
  - id, name
  - base_agent_type + base_agent_version (FK)
  - config_overrides (JSONB - patches to base)
  - metrics, usage_count
  - parent_variant_id (for variant lineage)

-- Per-entity learnings (append-only observations)
entity_state_log
  - entity_id, namespace, path
  - data, created_at

-- Proposed improvements (awaiting review)
improvement_proposals
  - id
  - target_type (base_agent | variant | entity)
  - target_id
  - proposed_changes (JSONB)
  - source (metrics | agent_observation | human)
  - status (pending | approved | rejected)
  - reviewed_by, reviewed_at
```

---

## Trade-offs Summary

| Approach | Stability | Flexibility | Complexity |
|----------|-----------|-------------|------------|
| Override from base | Medium | High | Medium |
| Override from snapshot | High | Medium | Medium |
| Full copy (groups today) | High | Low | Low |
| Dynamic entity adaptation | Low | Very high | High |
| Periodic batch updates | High | Medium | Medium |

---

## Key Decisions Made

1. **Every agent is an orchestrator** - no fundamental distinction
2. **Variants reference snapshot versions** - base can evolve safely
3. **Entity state is observation, not mutation** - agents read but don't auto-change
4. **HITL oversight for strategic changes** - system proposes, human approves
5. **Domain column on agents** - single table, filterable by domain

---

## Open Questions

1. How to handle cross-domain agents (utilities used everywhere)?
2. When does a successful variant get promoted to a new base?
3. How does client-specific configuration (client_X.agent_instances) fit in?
4. Retention/archival of old variants?

---

## Next: Organizational Framework Thought Experiment

Exploring how the framework might model an entire organization with CEO, departments, employees, each with their own agent instances and evolution patterns.

