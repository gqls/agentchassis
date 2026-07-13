-- ============================================================================
-- PERSONA COGNITIVE ARCHITECTURE
-- Full cognitive system with swappable components (LLM-based → specialized)
-- ============================================================================

-- ============================================================================
-- 1. PERSONAS (Core Identity - Immutable)
-- ============================================================================
CREATE TABLE IF NOT EXISTS personas (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    persona_type TEXT, -- 'expert', 'creative', 'analyst', 'conversational'

-- Core personality (the "soul" of the persona)
    personality_dna JSONB NOT NULL,
    /*
    {
        "biographical": {...},
        "psychological": {
            "big_five": {...},
            "core_values": [...],
            "cognitive_biases": [...],
            "emotional_triggers": [...]
        },
        "expertise_domains": ["climate_science", "literature", "philosophy"],
        "communication_style": {...},
        "worldview": {...}
    }
    */

    -- What this persona can do
    capabilities TEXT[], -- ['write', 'analyze', 'debate', 'teach', 'review']

    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

-- ============================================================================
-- 2. COGNITIVE COMPONENTS (Swappable Subsystems)
-- ============================================================================
CREATE TABLE IF NOT EXISTS persona_cognitive_components (
                                                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id UUID REFERENCES personas(id),

    -- Which cognitive subsystem
    component_type TEXT NOT NULL,
    /* Types:
       - 'perception': Understand and classify tasks
       - 'working_memory': Current task context
       - 'episodic_memory': Remember past experiences
       - 'semantic_memory': Long-term knowledge storage
       - 'procedural_memory': How-to knowledge
       - 'knowledge_retrieval': Access knowledge base
       - 'reasoning_engine': Planning and decision-making
       - 'response_generator': Generate outputs
       - 'style_applicator': Apply personality to outputs
       - 'learning_system': Update knowledge from experience
    */

    component_name TEXT NOT NULL, -- e.g., "llm_based_v1", "vector_db_v1", "knowledge_graph_v1"

-- Implementation configuration
    implementation JSONB NOT NULL,
    /*
    For LLM-based (Phase 1):
    {
        "type": "llm_based",
        "provider": "anthropic",
        "model": "claude-sonnet-4",
        "prompt_template": "...",
        "temperature": 0.7
    }

    For specialized (Phase 2+):
    {
        "type": "vector_database",
        "provider": "pinecone",
        "index_name": "bimpton-episodic-memory",
        "embedding_model": "text-embedding-3-large"
    }

    Or:
    {
        "type": "knowledge_graph",
        "provider": "neo4j",
        "graph_id": "bimpton-knowledge-v1",
        "query_engine": "cypher"
    }

    Or:
    {
        "type": "custom_service",
        "endpoint": "https://reasoning-engine.example.com",
        "api_key_env_var": "REASONING_ENGINE_KEY"
    }
    */

    -- Status
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    version INTEGER DEFAULT 1,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    UNIQUE(persona_id, component_type, component_name)
    );

CREATE INDEX idx_cognitive_components_persona ON persona_cognitive_components(persona_id);
CREATE INDEX idx_cognitive_components_type ON persona_cognitive_components(component_type);

-- ============================================================================
-- 3. PERSONA INSTANCES (Running State)
-- ============================================================================
CREATE TABLE IF NOT EXISTS persona_instances (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id UUID REFERENCES personas(id),
    orchestration_id UUID NOT NULL,

    -- Current cognitive state
    working_memory JSONB DEFAULT '{}',
    /*
    {
        "current_task": {...},
        "active_context": {...},
        "temporary_facts": [...]
    }
    */

    episodic_memory JSONB DEFAULT '{"experiences": []}',
    /*
    {
        "experiences": [
            {
                "timestamp": "2025-01-15T10:30:00Z",
                "task_type": "write_blog_post",
                "topic": "ocean acidification",
                "outcome": "success",
                "learned": "user prefers accessible explanations"
            }
        ]
    }
    */

    semantic_memory JSONB DEFAULT '{"facts": {}}',
    /*
    {
        "facts": {
            "user_preferences": {...},
            "project_context": {...},
            "accumulated_knowledge": [...]
        }
    }
    */

    emotional_state JSONB DEFAULT '{}',
    /*
    {
        "engagement_level": 0.8,
        "recent_triggers": [],
        "current_mood": "curious"
    }
    */

    -- Lifecycle
    status TEXT DEFAULT 'active', -- 'active', 'hibernating', 'terminated'
    spawned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    last_active TIMESTAMP WITH TIME ZONE DEFAULT now(),

    UNIQUE(persona_id, orchestration_id)
    );

CREATE INDEX idx_persona_instances_persona ON persona_instances(persona_id);
CREATE INDEX idx_persona_instances_orchestration ON persona_instances(orchestration_id);
CREATE INDEX idx_persona_instances_status ON persona_instances(status);

-- ============================================================================
-- 4. KNOWLEDGE BASE (Persona-Specific Knowledge)
-- ============================================================================
CREATE TABLE IF NOT EXISTS persona_knowledge (
                                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    persona_id UUID REFERENCES personas(id),

    -- Knowledge classification
    knowledge_type TEXT NOT NULL, -- 'fact', 'opinion', 'expertise', 'belief', 'procedure'
    domain TEXT, -- 'climate_science', 'literature', etc.

-- The knowledge itself
    content JSONB NOT NULL,
    /*
    For fact:
    {
        "statement": "CO2 levels have risen 50% since pre-industrial times",
        "confidence": 0.95,
        "source": "IPCC AR6 Report",
        "last_verified": "2024-01-15"
    }

    For expertise:
    {
        "skill": "climate_modeling",
        "proficiency": 0.9,
        "sub_skills": ["atmospheric_chemistry", "ocean_circulation"],
        "limitations": ["economic_modeling"]
    }

    For belief:
    {
        "belief": "Technology can help solve climate crisis",
        "strength": 0.7,
        "supporting_evidence": [...],
        "conflicts_with": [...]
    }
    */

    -- Metadata
    embedding VECTOR(1536), -- For semantic search (can be NULL initially)
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

CREATE INDEX idx_persona_knowledge_persona ON persona_knowledge(persona_id);
CREATE INDEX idx_persona_knowledge_type ON persona_knowledge(knowledge_type);
CREATE INDEX idx_persona_knowledge_domain ON persona_knowledge(domain);
-- Vector index for semantic search (when implemented)
-- CREATE INDEX idx_persona_knowledge_embedding ON persona_knowledge USING ivfflat(embedding);

-- ============================================================================
-- 5. TASK EXECUTION LOG (Learning Data)
-- ============================================================================
CREATE TABLE IF NOT EXISTS persona_task_executions (
                                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES persona_instances(id),
    persona_id UUID REFERENCES personas(id),

    -- Task details
    task_type TEXT NOT NULL,
    task_input JSONB NOT NULL,
    task_output JSONB,

    -- Cognitive trace (which components were used, how)
    cognitive_trace JSONB,
    /*
    {
        "perception": {
            "component": "llm_based_v1",
            "classified_as": "complex_writing_task",
            "decomposed_into": ["research", "outline", "write", "review"]
        },
        "knowledge_retrieved": [
            {"source": "semantic_memory", "facts": [...]},
            {"source": "episodic_memory", "relevant_experiences": [...]}
        ],
        "reasoning_steps": [
            {"step": "analyze_requirements", "output": "..."},
            {"step": "plan_structure", "output": "..."}
        ],
        "response_generated": {
            "component": "llm_based_v1",
            "iterations": 2
        }
    }
    */

    -- Outcome
    success BOOLEAN,
    quality_score DECIMAL,
    user_feedback TEXT,

    -- What was learned
    learned_facts JSONB,
    memory_updates JSONB,

    execution_time_ms INTEGER,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

CREATE INDEX idx_task_executions_instance ON persona_task_executions(instance_id);
CREATE INDEX idx_task_executions_persona ON persona_task_executions(persona_id);
CREATE INDEX idx_task_executions_type ON persona_task_executions(task_type);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Get active cognitive components for a persona
CREATE OR REPLACE FUNCTION get_persona_cognitive_system(p_persona_id UUID)
RETURNS TABLE (
    component_type TEXT,
    component_name TEXT,
    implementation JSONB
) AS $$
BEGIN
RETURN QUERY
SELECT
    pcc.component_type,
    pcc.component_name,
    pcc.implementation
FROM persona_cognitive_components pcc
WHERE pcc.persona_id = p_persona_id
  AND pcc.is_active = true
  AND (pcc.is_default = true OR NOT EXISTS (
    SELECT 1 FROM persona_cognitive_components pcc2
    WHERE pcc2.persona_id = p_persona_id
      AND pcc2.component_type = pcc.component_type
      AND pcc2.is_active = true
      AND pcc2.is_default = true
))
ORDER BY pcc.component_type;
END;
$$ LANGUAGE plpgsql;

-- Get or create persona instance
CREATE OR REPLACE FUNCTION get_or_create_persona_instance(
    p_persona_id UUID,
    p_orchestration_id UUID
)
RETURNS UUID AS $$
DECLARE
v_instance_id UUID;
BEGIN
    -- Try to get existing instance
SELECT id INTO v_instance_id
FROM persona_instances
WHERE persona_id = p_persona_id
  AND orchestration_id = p_orchestration_id
  AND status IN ('active', 'hibernating');

-- Create if doesn't exist
IF v_instance_id IS NULL THEN
        INSERT INTO persona_instances (persona_id, orchestration_id)
        VALUES (p_persona_id, p_orchestration_id)
        RETURNING id INTO v_instance_id;
ELSE
        -- Wake up if hibernating
UPDATE persona_instances
SET status = 'active',
    last_active = now()
WHERE id = v_instance_id;
END IF;

RETURN v_instance_id;
END;
$$ LANGUAGE plpgsql;

-- Update persona instance memory after task
CREATE OR REPLACE FUNCTION update_persona_memory(
    p_instance_id UUID,
    p_working_memory JSONB DEFAULT NULL,
    p_episodic_memory JSONB DEFAULT NULL,
    p_semantic_memory JSONB DEFAULT NULL,
    p_emotional_state JSONB DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
UPDATE persona_instances
SET
    working_memory = COALESCE(p_working_memory, working_memory),
    episodic_memory = CASE
                          WHEN p_episodic_memory IS NOT NULL THEN
                              jsonb_set(
                                      episodic_memory,
                                      '{experiences}',
                                      (episodic_memory->'experiences') || p_episodic_memory->'experiences'
                              )
                          ELSE episodic_memory
        END,
    semantic_memory = COALESCE(p_semantic_memory, semantic_memory),
    emotional_state = COALESCE(p_emotional_state, emotional_state),
    last_active = now()
WHERE id = p_instance_id;
END;
$$ LANGUAGE plpgsql;

-- Search persona knowledge base
CREATE OR REPLACE FUNCTION search_persona_knowledge(
    p_persona_id UUID,
    p_query TEXT,
    p_domain TEXT DEFAULT NULL,
    p_limit INTEGER DEFAULT 10
)
RETURNS TABLE (
    knowledge_id UUID,
    knowledge_type TEXT,
    domain TEXT,
    content JSONB,
    relevance_score DECIMAL
) AS $$
BEGIN
    -- Simple text search (Phase 1)
    -- Later: Use vector similarity search
RETURN QUERY
SELECT
    pk.id,
    pk.knowledge_type,
    pk.domain,
    pk.content,
    -- Simple relevance scoring for now
    CASE
        WHEN pk.content::text ILIKE '%' || p_query || '%' THEN 1.0
        ELSE 0.5
        END as relevance_score
FROM persona_knowledge pk
WHERE pk.persona_id = p_persona_id
  AND (p_domain IS NULL OR pk.domain = p_domain)
  AND pk.content::text ILIKE '%' || p_query || '%'
ORDER BY relevance_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Log task execution for learning
CREATE OR REPLACE FUNCTION log_persona_task_execution(
    p_instance_id UUID,
    p_persona_id UUID,
    p_task_type TEXT,
    p_task_input JSONB,
    p_task_output JSONB,
    p_cognitive_trace JSONB,
    p_success BOOLEAN,
    p_quality_score DECIMAL DEFAULT NULL,
    p_learned_facts JSONB DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
v_execution_id UUID;
BEGIN
INSERT INTO persona_task_executions (
    instance_id,
    persona_id,
    task_type,
    task_input,
    task_output,
    cognitive_trace,
    success,
    quality_score,
    learned_facts
)
VALUES (
           p_instance_id,
           p_persona_id,
           p_task_type,
           p_task_input,
           p_task_output,
           p_cognitive_trace,
           p_success,
           p_quality_score,
           p_learned_facts
       )
    RETURNING id INTO v_execution_id;

RETURN v_execution_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE personas IS 'Core personality and identity of each persona - immutable';
COMMENT ON TABLE persona_cognitive_components IS 'Swappable cognitive subsystems (LLM-based → specialized implementations)';
COMMENT ON TABLE persona_instances IS 'Running instances with accumulated memory and state';
COMMENT ON TABLE persona_knowledge IS 'Persona-specific knowledge base (facts, expertise, beliefs)';
COMMENT ON TABLE persona_task_executions IS 'Execution log for learning and improvement';

COMMENT ON COLUMN persona_cognitive_components.component_type IS 'Which cognitive subsystem: perception, memory, reasoning, generation, etc.';
COMMENT ON COLUMN persona_cognitive_components.implementation IS 'How this component is implemented (LLM, vector DB, knowledge graph, custom service)';
COMMENT ON COLUMN persona_instances.working_memory IS 'Current task context and temporary state';
COMMENT ON COLUMN persona_instances.episodic_memory IS 'Past experiences and task history';
COMMENT ON COLUMN persona_instances.semantic_memory IS 'Long-term factual knowledge and learned patterns';