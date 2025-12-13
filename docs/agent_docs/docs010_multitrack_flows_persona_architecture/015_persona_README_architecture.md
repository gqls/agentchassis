# Persona Cognitive Architecture

## Vision

Build AI personas that are **complete cognitive entities** - not just prompts, but systems with:
- **Memory** (episodic, semantic, working)
- **Reasoning** (planning, evaluation, learning)
- **Knowledge** (domain expertise, beliefs, biases)
- **Personality** (stable traits that shape all behavior)
- **Evolution** (can be upgraded from LLM-based → fine-tuned → multi-model → custom systems)

## What This Enables

**Today:**
- Dr. Michael Bimpton writes a climate blog post in his distinctive voice
- He remembers previous tasks within a project
- He applies his expertise and communication style

**Tomorrow:**
- Dr. Bimpton is backed by fine-tuned models trained on his writing
- His episodic memory uses vector search for better recall
- His knowledge base is a queryable graph

**Future:**
- Dr. Bimpton has multiple specialized models (writing vs conversation vs analysis)
- Custom reasoning engine with climate-specific logic
- Active learning system that identifies knowledge gaps
- Full cognitive architecture rivaling human cognitive capabilities

## Architecture Overview

```
Persona Agent (e.g., Dr. Bimpton)
│
├── Personality DNA (immutable)
│   ├── Biographical background
│   ├── Psychological profile (Big Five, biases, triggers)
│   ├── Expertise domains
│   ├── Communication style
│   └── Worldview
│
├── Cognitive Components (swappable)
│   ├── Perception (understand tasks)
│   ├── Working Memory (current context)
│   ├── Episodic Memory (past experiences)
│   ├── Semantic Memory (learned facts)
│   ├── Knowledge Retrieval (access knowledge base)
│   ├── Reasoning Engine (plan and decide)
│   ├── Response Generator (create content)
│   ├── Style Applicator (apply voice)
│   └── Learning System (update memory)
│
├── Running Instance (stateful)
│   ├── Working Memory (current task)
│   ├── Episodic Memory (accumulated experiences)
│   ├── Semantic Memory (learned patterns)
│   └── Emotional State
│
└── Knowledge Base (persona-specific)
    ├── Facts (with confidence levels)
    ├── Expertise (skills and limitations)
    ├── Beliefs (with strength and conflicts)
    └── Opinions (with supporting evidence)
```

## Database Schema

### Core Tables

**1. `personas`** - Core personality (immutable)
```sql
- id UUID
- name TEXT
- personality_dna JSONB  -- Complete personality profile
- capabilities TEXT[]     -- What they can do
```

**2. `persona_cognitive_components`** - Swappable subsystems
```sql
- persona_id UUID
- component_type TEXT    -- 'perception', 'memory', 'reasoning', etc.
- implementation JSONB   -- How this component works (LLM, vector DB, etc.)
- is_default BOOLEAN     -- Which implementation is active
```

**3. `persona_instances`** - Running state
```sql
- persona_id UUID
- orchestration_id UUID
- working_memory JSONB
- episodic_memory JSONB
- semantic_memory JSONB
- emotional_state JSONB
```

**4. `persona_knowledge`** - Knowledge base
```sql
- persona_id UUID
- knowledge_type TEXT  -- 'fact', 'expertise', 'belief', 'opinion'
- content JSONB
- embedding VECTOR     -- For semantic search (future)
```

**5. `persona_task_executions`** - Execution log for learning
```sql
- instance_id UUID
- task_type TEXT
- cognitive_trace JSONB  -- Which components were used, how
- learned_facts JSONB    -- What was learned
```

## Cognitive Workflow

When you call a persona agent:

```
1. Initialize
   └─> load_cognitive_system
       ├─ Load personality DNA
       ├─ Load cognitive components (which implementations)
       └─ Get or create instance (load memory)

2. Perceive
   └─> perceive_task
       ├─ Classify task type
       ├─ Assess complexity
       └─ Decompose if needed

3. Retrieve
   └─> retrieve_from_memory
       ├─ Search episodic memory (relevant experiences)
       ├─ Search semantic memory (relevant facts)
       └─ Query knowledge base

4. Reason
   └─> reason_and_plan
       ├─ Plan approach
       ├─ Consider personality traits
       └─ Identify challenges

5. Generate
   └─> generate_response
       └─ Create content using plan and knowledge

6. Style
   └─> apply_style
       └─ Apply persona's unique voice

7. Learn
   └─> update_memory
       ├─ Add to episodic memory
       ├─ Update semantic memory
       └─ Log execution

8. Complete
   └─> Return styled response
```

## Evolution Path

### Phase 1: All LLM-Based (Start Here)

Every cognitive component uses LLM prompts:

```json
{
    "component_type": "episodic_memory",
    "implementation": {
        "type": "llm_based",
        "model": "claude-sonnet-4",
        "prompt": "Review past experiences and find relevant ones..."
    }
}
```

**Pros:** Simple to implement, flexible, good quality
**Cons:** Slower, expensive, limited by context window

### Phase 2: Selective Specialization

Upgrade bottlenecks to specialized systems:

```json
{
    "component_type": "episodic_memory",
    "implementation": {
        "type": "vector_database",
        "provider": "pinecone",
        "index": "bimpton-episodic-memory",
        "embedding_model": "text-embedding-3-large"
    }
}
```

**Pros:** Faster, more scalable, better retrieval
**Cons:** More complex, need to manage multiple services

### Phase 3: Fine-Tuned Models

Train persona-specific models:

```json
{
    "component_type": "response_generator",
    "implementation": {
        "type": "fine_tuned",
        "model_id": "ft-dr-bimpton-v1",
        "base_model": "claude-sonnet-4"
    }
}
```

**Pros:** More consistent personality, faster, cheaper
**Cons:** Training cost, less flexible

### Phase 4: Multi-Model Architecture

Different models for different tasks:

```json
{
    "component_type": "response_generator",
    "implementation": {
        "type": "multi_model",
        "models": {
            "writing": "ft-bimpton-writing-v1",
            "conversation": "ft-bimpton-conversation-v1",
            "analysis": "ft-bimpton-analysis-v1"
        }
    }
}
```

**Pros:** Optimized for each task type
**Cons:** Complex orchestration, multiple models to manage

### Phase 5: Custom Cognitive Systems

Full custom implementation:

```json
{
    "component_type": "reasoning_engine",
    "implementation": {
        "type": "custom_service",
        "endpoint": "https://bimpton-reasoning.example.com",
        "capabilities": ["causal_reasoning", "climate_modeling", "ethical_evaluation"]
    }
}
```

**Pros:** Maximum control and optimization
**Cons:** Significant development effort

## Key Architectural Principles

### 1. **Separation of Concerns**
- **Personality DNA** (who they are) separate from **Implementation** (how they work)
- Can evolve implementation without changing personality
- Can A/B test implementations

### 2. **Swappable Components**
- Each cognitive subsystem has interface
- Multiple implementations per subsystem
- Switch via `is_default` flag
- No workflow changes needed

### 3. **Memory Persistence**
- Personas remember within orchestration scope
- Memory accumulates across tasks
- Learning improves performance over time

### 4. **Modularity**
- Add new personas easily
- Upgrade components independently
- Mix and match implementations

### 5. **Observability**
- Cognitive trace logs what happened
- Task executions tracked
- Can analyze and optimize

## Implementation Steps

### Step 1: Deploy Schema

```bash
psql $DATABASE_URL -f persona_cognitive_architecture.sql
```

Creates:
- `personas` table
- `persona_cognitive_components` table
- `persona_instances` table
- `persona_knowledge` table
- `persona_task_executions` table
- Helper functions

### Step 2: Create Persona

```bash
psql $DATABASE_URL -f dr_bimpton_setup_example.sql
```

Sets up Dr. Bimpton with:
- Complete personality profile
- 8 LLM-based cognitive components
- Sample knowledge base entries
- Agent definition

### Step 3: Implement Custom Actions

In your agent-chassis Go codebase, implement:

1. `load_cognitive_system` - Initialize persona
2. `perceive_task` - Classify task
3. `retrieve_from_memory` - Access memory/knowledge
4. `reason_and_plan` - Plan approach
5. `generate_response` - Create content
6. `apply_style` - Apply voice
7. `update_memory` - Learn from task

See `persona_cognitive_actions.sql` for detailed specs.

### Step 4: Test

```json
{
    "action": "call_agent",
    "config": {
        "agent_type": "persona-dr-bimpton",
        "input_fields": {
            "task_type": "write_content",
            "task_input": {
                "topic": "Ocean acidification",
                "length": "800_words",
                "audience": "general_public"
            }
        }
    }
}
```

### Step 5: Evolve

When ready, upgrade specific components:

```sql
-- Add vector database for episodic memory
INSERT INTO persona_cognitive_components (
    persona_id,
    component_type,
    component_name,
    implementation,
    is_default
) VALUES (
    'dr-bimpton-id',
    'episodic_memory',
    'vector_db_v1',
    '{"type": "vector_database", "provider": "pinecone", ...}',
    false  -- test first
);

-- A/B test, then make default when satisfied
UPDATE persona_cognitive_components
SET is_default = true
WHERE component_name = 'vector_db_v1';
```

## Example Personas

### Dr. Michael Bimpton (Climate Scientist)
- **Expertise:** Climate science, literature, philosophy
- **Voice:** Academic but accessible, sailing metaphors, Socratic questions
- **Use cases:** Science communication, expert analysis, teaching
- **Personality:** High openness, moderate conscientiousness, introverted

### Elena Martinez (B2B Copywriter)
- **Expertise:** Marketing, value propositions, executive communication
- **Voice:** Professional, benefit-focused, second-person
- **Use cases:** Website copy, service pages, marketing content
- **Personality:** High conscientiousness, moderate extraversion

### Marcus Williams (Conversion Specialist)
- **Expertise:** Psychology, A/B testing, landing pages
- **Voice:** Direct, urgent, action-oriented
- **Use cases:** CTAs, landing pages, conversion optimization
- **Personality:** Results-driven, analytical

## Integration with Existing System

Personas integrate seamlessly with your agent system:

```json
// Content creator can call persona
{
    "workflow": {
        "steps": {
            "generate_content": {
                "action": "call_agent",
                "config": {
                    "agent_type": "persona-dr-bimpton",
                    "input_fields": ["content_requirements"]
                }
            }
        }
    }
}
```

Personas can also call other agents:

```json
// Dr. Bimpton calls research agent
{
    "workflow": {
        "steps": {
            "research": {
                "action": "call_agent",
                "config": {
                    "agent_type": "research-agent",
                    "input_fields": ["research_query"]
                }
            }
        }
    }
}
```

## Benefits

✅ **Future-proof:** Start simple, evolve to sophisticated
✅ **Modular:** Upgrade components independently
✅ **Memory:** Personas remember and learn
✅ **Consistent:** Stable personality across tasks
✅ **Scalable:** Add personas easily
✅ **Observable:** Full cognitive trace
✅ **Realistic:** Captures human-like complexity

## Files

1. **`persona_cognitive_architecture.sql`** - Database schema
2. **`persona_cognitive_actions.sql`** - Action specifications
3. **`dr_bimpton_setup_example.sql`** - Complete example setup

## Next Steps

1. Deploy schema to database
2. Review Dr. Bimpton example
3. Implement custom actions in Go
4. Test with simple tasks
5. Monitor and optimize
6. Upgrade components as needed
7. Create additional personas

The architecture is ready for the full vision while starting simple today.