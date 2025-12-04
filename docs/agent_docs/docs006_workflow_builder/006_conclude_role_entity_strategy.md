# Architecture Discussion Snapshot: Organizational Framework

## Date: Session in progress (continued from evolution snapshot)

## Context

Exploring how the framework could model an entire organization, with employees, roles, relationships, and continuous operation. This builds on the earlier evolution/variants discussion.

---

## The Parallel Structure

The framework is domain-agnostic. Same patterns apply to web building and organizations:

| Web Building Domain | Organization Domain |
|---------------------|---------------------|
| Domain (ai-orchestration.com) | Employee/Role (Sarah, Marketing Writer) |
| Objective (sell AI framework) | Strategy (increase brand awareness) |
| Intake orchestrator | Strategy-to-task breakdown |
| Builder workflow | Role workflow (daily/weekly work) |
| Site deployer | Output delivery (publish, send, ship) |
| Legal/brand review (optional) | Legal/HR/approval (conditional filters) |
| Entity state (domain learnings) | Employee state (role learnings) |
| Client schema | Employee schema |

---

## Cross-Cutting Agents

**Key insight:** Cross-cutting agents are just agents that many other agents call. Nothing special about their definition.

A legal-review-agent:
- Is a normal agent_definition
- Has a workflow (receive content → analyze → approve/reject/flag)
- Accepts certain input_fields
- Returns a response

The "interface" is implicit in what it reads from input_data and returns in response.

**How callers know to invoke them:**

| Method | How it works |
|--------|--------------|
| Hardcoded | Workflow has explicit `call_agent: legal-review-agent` step |
| Policy-injected | Legal's policy says "if competitor mentioned, inject legal review step" |
| Capability discovery | Query for agent with capability "legal_review" |

Cross-cutting is a usage pattern, not a definition pattern.

---

## Relationships as First-Class Objects

Relationships are like website links - first-class entities with their own identity and state.

| Website Links | Employee Relationships |
|---------------|----------------------|
| Source page → Target page | Sarah → Marketing Manager |
| Type: internal, external, affiliate | Type: reports_to, collaborates_with, mentors |
| Properties: nofollow, sponsored | Properties: communication_style, trust_level |
| Metrics: clicks, conversions | Metrics: response_time, collaboration_success |
| Can break (404) | Can break (conflict, departure) |

**Schema:**

```sql
CREATE TABLE relationships (
    id UUID PRIMARY KEY,
    
    -- Endpoints
    source_entity_id VARCHAR(255),
    source_entity_type VARCHAR(100),  -- "role", "agent", "external"
    target_entity_id VARCHAR(255),
    target_entity_type VARCHAR(100),
    
    -- Relationship properties
    relationship_type VARCHAR(100),   -- "reports_to", "collaborates_with"
    direction VARCHAR(20),            -- "one_way", "bidirectional"
    
    -- Relationship-specific config
    properties JSONB,
    
    -- State
    status VARCHAR(50),               -- "active", "strained", "dormant"
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

Relationships have their own entity_state:

```
entity_id: "relationship.sarah.marketing-manager"
entity_type: "relationship"
namespace: "communication"
path: "preferred_channel" → "slack"
path: "feedback_style" → "bullet_points"
```

The relationship learns and evolves independently of either party.

---

## Role vs Agent

**Roles and agents are separate concepts:**

| Concept | What it is | Framework implementation |
|---------|------------|-------------------------|
| Agent | Execution capability | agent_definition, agent_instance |
| Role | Organizational position | Separate entity, owns context |

**Role types:**

| Type | Examples | Characteristics |
|------|----------|-----------------|
| Identity | Sarah, Bob, AI-Agent-7 | Persistent, has schema, has relationships |
| Function | Writer, Reviewer, Approver | Capability template, stateless |
| Composite | Marketing Writer, Senior Dev | Bundles other roles/agents |
| Position | Incident Commander | Slot that can be filled by different identities |

**How they relate:**

- A function role maps to an agent_definition
- A composite role maps to an orchestrator agent_definition
- An identity role owns a schema (like client_X)
- A position is a role slot with its own state, fillable by identities

**Example:**

```
Position: Marketing Content Writer
  ├── state: position's learnings (what works for this role)
  ├── relationships: who this role works with
  └── filled_by: sarah (identity)
        ├── schema: employee_sarah
        ├── state: sarah's personal learnings
        └── relationships: sarah's personal connections
```

When Sarah works as Marketing Writer:
- Reads position state (what's expected of this role)
- Reads her state (her personal approach)
- Uses the intersection

---

## Continuous Intake: Listeners

**The question:** If an employee's work is ongoing, how is that modeled?

**Answer:** Listeners are infrastructure (like adapters), orchestrations are discrete.

### Listener Pattern

Same as existing adapters:

```
Git Adapter                      Employee Listener
  - Pod: always running            - Pod: always running (shared)
  - Listens on: git.requests       - Listens on: employee.*.requests
  - Receives: git commands         - Receives: task assignments
  - Does: git operations           - Does: spawns orchestrations
  - Returns: results               - Returns: acknowledgments
```

### Shared vs Dedicated Listeners

**Option A: One pod per employee** (expensive, simple)
**Option B: Shared listener with routing** (efficient, preferred)
**Option C: Extend generic agent** (reuse existing infrastructure)

Recommendation: Shared listener (like generic agent) that routes based on topic.

### Topic Structure

```
system.agent.generic.requests      ← current entry point

employee.sarah.requests            ← sarah's inbox
employee.bob.requests              ← bob's inbox
role.incident-commander.requests   ← position inbox
team.marketing.requests            ← team inbox
department.engineering.requests    ← department inbox
```

Listener subscribes to patterns:
```
employee.*.requests
role.*.requests
team.*.requests
```

### What Happens When Work Arrives

```
Message arrives on: employee.sarah.requests
  │
  ▼
Shared Role Listener
  │
  ├── Extract: role_context = "sarah"
  ├── Load: sarah's entity_state, relationships, instances
  │
  ▼
Spawn Orchestration
  │
  ├── workflow: from agent_definition (e.g., marketing-writer)
  ├── collected_data: includes sarah's context
  │
  ▼
Workflow executes
  │
  ├── Reads sarah's state when needed
  ├── Uses sarah's agent instances (her personalized variants)
  ├── Writes learnings to sarah's entity_state
  │
  ▼
Completes
```

### The Listener's Job (Minimal)

1. **Subscribe** to relevant topics
2. **Authenticate** incoming messages
3. **Load context** (schema, state, relationships)
4. **Spawn orchestration** with that context
5. **Route responses** back to caller

It doesn't:
- Hold state (that's in the database)
- Make decisions (that's the workflow)
- Know what work looks like (that's the agent definition)

---

## Authority and Policy as Filters

Traditional: Hierarchy of approvals
Framework: Conditional filters in workflow

```
Sarah writes
  → [if budget > $1000] → finance-approval-agent
  → [if competitor mentioned] → legal-review-agent
  → [if external publication] → brand-guardian-agent
  → Published
```

Policy example:

```json
{
  "policy_owner": "legal-agent",
  "rule": "competitor_mention_review",
  "trigger": {
    "content_type": ["blog_post", "social", "press_release"],
    "condition": "mentions_competitor == true"
  },
  "action": "require_approval",
  "approver": "legal-review-agent",
  "can_disable": false
}
```

Policies live with their owner agents. Some are mandatory (legal), some optional (brand tone).

---

## Strategy as Root Intake

Company strategy flows down and becomes tasks:

```
Company Strategy
  "Increase market share by 15% through brand awareness"
      │
      ▼
Marketing Strategy (owned by CMO agent)
  "Create thought leadership content targeting enterprise CTOs"
      │
      ▼
Marketing Tasks (broken down by Marketing Manager agent)
  ├── "Write 4 blog posts on AI automation"
  ├── "Launch LinkedIn campaign"
  └── "Redesign case studies page"
      │
      ▼
Sarah's Intake
  "Write blog post: Why AI Orchestration Matters for Enterprise"
```

Strategy IS the intake. It flows down, gets more specific at each level.

---

## Complete Concept Mapping

| Concept | What It Is | Lifecycle |
|---------|------------|-----------|
| Role (identity) | Sarah | Persists indefinitely, owns schema |
| Role (function) | Blog Writer | Template, instantiated per use |
| Role (composite) | Marketing Writer | Bundle, references components |
| Role (position) | Incident Commander | Slot, can change fillers |
| Agent definition | blog-writer | Capability template |
| Agent instance | Sarah's blog-writer | Personalized for Sarah's context |
| Entity state | Sarah's learnings | Append-only log in her schema |
| Relationship | Sarah ↔ Manager | First-class, has own state |
| Orchestration | "Write blog post X" | Discrete, starts and ends |
| Listener | Role inbox handler | Infrastructure, always on, shared |
| Policy | Legal review rules | Owned by responsible agent |

---

## Updated Schema Summary

```sql
-- Base agent capabilities
agent_definitions
  - type, version, workflow, capabilities
  - is_snapshot (frozen or living)
  - domain (web, finance, org, etc.)

-- Specialized variants
agent_variants
  - base_agent_type + version
  - config_overrides
  - metrics, lineage

-- Roles (organizational positions)
roles
  - id, name, role_type (identity | function | composite | position)
  - parent_role_id (for composition)
  - schema_name (for identity roles)
  - default_workflow

-- Role assignments (who fills what)
role_assignments
  - role_id (position)
  - filler_role_id (identity)
  - valid_from, valid_until

-- Per-role instance configuration
role_[name].agent_instances
  - Personalized agents for this role

-- Per-entity learnings
entity_state_log
  - entity_id, entity_type, namespace, path, data

-- Relationships between entities
relationships
  - source, target, type, direction
  - properties, status

-- Improvement proposals
improvement_proposals
  - target, proposed_changes, source, status
```

---

## Key Architectural Decisions

1. **Roles are separate from agents** - organizational concept vs execution capability
2. **Listeners are shared infrastructure** - like adapters, not per-employee pods
3. **Orchestrations are discrete** - each task is bounded, listeners spawn them
4. **Relationships are first-class** - own identity, own state, own evolution
5. **Cross-cutting agents are just agents** - nothing special about definition
6. **Policy lives with owners** - legal owns legal policies, HR owns HR policies
7. **Authority is filtering, not hierarchy** - conditions trigger reviews, not org chart
8. **Strategy flows down as intake** - each level breaks down to more specific tasks
9. **Identity roles own schemas** - like client_X pattern, get their own namespace
10. **Listeners can have default workflows** - overrideable, like generic agent

---

## Open Items for Later

1. Restart/reboot handling when listeners go down
2. How role assignments get managed (HR workflow?)
3. Detailed policy injection mechanism
4. Team/department as composite roles or separate concept
5. Cross-role collaboration protocols

---

## Next Steps

Ready to implement:
1. Entity state log (from earlier discussion)
2. Relationships table
3. Role listener infrastructure
4. Policy-owner pattern in agents

Future consideration:
1. Roles table and assignments
2. Position vs identity distinction
3. Strategy breakdown workflows