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

===
# Implementation: Every Agent Is An Orchestrator

## Overview

This implementation removes dependency on `agent_group_definitions` and unifies everything under `agent_definitions`. Every agent can orchestrate other agents through its workflow.

**Approach:** Minimal changes with backward compatibility. Existing code using `GroupDiscovery` and `group_type` will continue to work.

---

## Files to Update

### 1. Database Migration
**File:** Run `migration_entity_state_and_builder_agents.sql`

Creates:
- `entity_state_log` table (persistent cross-orchestration data)
- `relationships` table (first-class relationships between entities)

Updates `agent_definitions`:
- Adds `usage_count`, `is_snapshot`, `briefing_questionnaire` columns
- Creates `intake-orchestrator`, `landing-page-builder`, `content-site-builder`, `multipage-wrapper` agents

### 2. Discovery Service
**File:** Replace `platform/discovery/agent_discovery.go` with `agent_discovery_updated.go`

Key changes:
- New `AgentDefinitionDiscovery` struct queries `agent_definitions`
- `GroupDiscovery` is aliased to `AgentDefinitionDiscovery` (backward compat)
- `FindBestGroup` now queries `agent_definitions` instead of `agent_group_definitions`
- Existing code using `GroupDiscovery` will compile and work unchanged

### 3. Message Processor
**File:** Apply changes from `processor_patch.go` to `platform/agentbase/processor.go`

Minimal changes:
- Update `extractGroupInfo` to check `agent_type` first, fall back to `group_type`
- Update `isOrchestrationAction` to include `spawn_agent`
- Store metadata as both `agent_group` and `agent_definition` for compatibility

---

## Message Format

### Both formats work:

**New (preferred):**
```json
{
  "action": "orchestrate",
  "config": {
    "agent_type": "intake-orchestrator"
  },
  "input_data": {...}
}
```

**Legacy (still works):**
```json
{
  "action": "orchestrate",
  "config": {
    "group_type": "intake-orchestrator"
  },
  "input_data": {...}
}
```

---

## How It Works

```
Message arrives with agent_type: "intake-orchestrator"
    │
    ▼
extractGroupInfo() 
    ├── Checks config.agent_type first
    ├── Falls back to config.group_type
    │
    ▼
GroupDiscovery.FindBestGroup() [aliased to AgentDefinitionDiscovery]
    │
    ├── Queries: SELECT ... FROM agent_definitions WHERE type = $1
    ├── Returns AgentGroup struct (for backward compat)
    │
    ▼
Workflow from agent_definitions.default_config->'workflow'
    │
    ▼
Executes normally
```

---

## Backward Compatibility

| Old Code | Still Works? | Notes |
|----------|--------------|-------|
| `discovery.NewGroupDiscovery(db)` | ✓ | Aliased to AgentDefinitionDiscovery |
| `discovered.FindBestGroup(...)` | ✓ | Now queries agent_definitions |
| `config["group_type"]` in messages | ✓ | extractGroupInfo checks both |
| `CollectedData["agent_group"]` | ✓ | Still populated |
| `spawn_group` action | ✓ | Still recognized |

---

## New Capabilities

| Feature | How to Use |
|---------|------------|
| `agent_type` in messages | Preferred over `group_type` |
| `spawn_agent` action | Works like `spawn_group` |
| `CollectedData["agent_definition"]` | New metadata storage |
| `usage_count` tracking | Auto-incremented on each use |
| `is_snapshot` flag | Mark frozen versions |
| `entity_state_log` | Store persistent data across orchestrations |
| `relationships` | Model connections between entities |
| `improvement_proposals` | Queue improvements for HITL review |
| `discover_agents` action | Find agents by capabilities |
| `review_performance` action | Analyze and record execution metrics |

---

## Schema Summary

### agent_definitions (updated columns)
```sql
usage_count INTEGER DEFAULT 0        -- Discovery ranking
is_snapshot BOOLEAN DEFAULT false    -- Frozen version flag
briefing_questionnaire JSONB         -- Briefing workflows
```

### entity_state_log (new)
```sql
entity_id VARCHAR(255)       -- "example.com"
entity_type VARCHAR(100)     -- "domain"
namespace VARCHAR(100)       -- NULL=shared, or agent_type
path VARCHAR(255)            -- "brand.tone"
data JSONB                   -- The actual data
superseded_by BIGINT         -- For compaction
```

### relationships (new)
```sql
source_entity_id, source_entity_type
target_entity_id, target_entity_type
relationship_type VARCHAR(100)   -- "reports_to", "collaborates_with"
properties JSONB                 -- Relationship config
status VARCHAR(50)               -- "active", "dormant", "ended"
```

### improvement_proposals (new)
```sql
target_type VARCHAR(50)      -- "agent_definition", "variant", "entity"
target_id VARCHAR(255)       -- Agent type or entity ID
proposed_changes JSONB       -- Analysis and suggestions
source VARCHAR(50)           -- "metrics", "agent_observation", "human"
status VARCHAR(50)           -- "pending", "approved", "rejected", "applied"
reviewed_by, reviewed_at
```

---

## Deployment Steps

1. **Run migration:**
   ```bash
   psql -d your_database -f migration_entity_state_and_builder_agents.sql
   ```

2. **Replace discovery file:**
   ```bash
   cp agent_discovery_updated.go platform/discovery/agent_discovery.go
   ```

3. **Apply processor patch:**
    - Open `processor_patch.go`
    - Apply the three changes to `platform/agentbase/processor.go`

4. **Build and deploy:**
   ```bash
   make build
   # Deploy as normal
   ```

5. **Test with new message format:**
   ```bash
   # Send to generic agent
   kafkacat -b kafka:9092 -t system.agent.generic.requests \
     -H "action=orchestrate" \
     -H "responses_topic=system.responses.generic" \
     -P <<EOF
   {"action":"orchestrate","config":{"agent_type":"intake-orchestrator"},"input_data":{"domain":"test.com","objective":"Test"}}
   EOF
   ```

---

## Verification

```sql
-- Check intake-orchestrator exists
SELECT type, display_name, version,
       default_config->'workflow'->'start_step' as start_step
FROM agent_definitions 
WHERE type = 'intake-orchestrator';

-- Check new columns exist
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'agent_definitions' 
AND column_name IN ('usage_count', 'is_snapshot', 'briefing_questionnaire');

-- Check new tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_name IN ('entity_state_log', 'relationships', 'improvement_proposals');
```

---

## Register New Actions

If you have an action registry, register the new actions:

```go
// In your action registry initialization
actionRegistry.Register("discover_agents", DiscoverAgentsAction)
actionRegistry.Register("review_performance", ReviewPerformanceAction)
actionRegistry.Register("approve_improvement", ApproveImprovementAction)
```

---

## Files Provided

| File | Purpose |
|------|---------|
| `agent_discovery_updated.go` | Replace `platform/discovery/agent_discovery.go` |
| `discovery_actions_updated.go` | Replace `platform/orchestration/actions/discovery_actions.go` |
| `processor_patch.go` | Changes to apply to `platform/agentbase/processor.go` |
| `migration_entity_state_and_builder_agents.sql` | Run in database |
| `architecture_snapshot_evolution_and_variants.md` | Evolution design discussion |
| `architecture_snapshot_organizational_framework.md` | Org framework discussion |

---

## Discovery Actions Changes

The `discovery_actions.go` file has been rewritten to work with the new architecture:

| Old Action | New Action | Changes |
|------------|------------|---------|
| `PlanAgentTeamAction` | `DiscoverAgentsAction` | Uses `AgentDefinitionDiscovery.FindByCapabilities()` |
| `ReviewPerformanceAction` | `ReviewPerformanceAction` | Records to `entity_state_log`, creates `improvement_proposals` |
| `ApproveAgentChangesAction` | `ApproveImprovementAction` | Updates proposal status, applies changes to `agent_definitions` |
| `ConditionalRouteActionOld` | Removed | Use `evaluate_condition` action instead |

### New Tables Used

- `entity_state_log` - Performance metrics stored per entity
- `improvement_proposals` - HITL review queue for suggested improvements
- `relationships` - (Available for future use)

### Removed Dependencies

- No more queries to `agent_groups` or `agent_group_definitions`
- No more `discovery.GroupDiscovery` direct calls (uses aliased version)