-- ============================================================================
-- PERSONA SYSTEM SCHEMA
-- For managing copywriter personalities and specialized agents
-- ============================================================================

-- ============================================================================
-- PERSONAS: Main persona definitions
-- ============================================================================
CREATE TABLE personas (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name TEXT NOT NULL UNIQUE,
                          description TEXT NOT NULL,

    -- Core configuration (JSONB for flexibility)
                          config JSONB NOT NULL,
    /*
    Example structure:
    {
        "biographical": {
            "background": "...",
            "education": "...",
            "specialization": "..."
        },
        "psychological": {
            "openness": 0.8,
            "conscientiousness": 0.9,
            "extraversion": 0.6,
            "agreeableness": 0.7,
            "neuroticism": 0.3,
            "core_values": ["Results", "Clarity", "Professionalism"]
        },
        "expertise": {
            "domain_1": 0.9,
            "domain_2": 0.7
        },
        "communication": {
            "vocabulary_level": "...",
            "sentence_structure": "...",
            "rhetorical_devices": [...],
            "speech_quirks": [...]
        },
        "voice_parameters": {
            "formality": 0.75,
            "technical_depth": 0.4,
            "sales_pressure": 0.6,
            "data_density": 0.5,
            "emotional_appeal": 0.4
        }
    }
    */

    -- Persona metadata
                          persona_type TEXT, -- 'copywriter', 'technical_writer', 'thought_leader', etc.
                          industry_focus TEXT[], -- industries this persona specializes in
                          content_types TEXT[], -- types of content this persona writes well

    -- Status
                          is_active BOOLEAN DEFAULT true,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                          version INTEGER DEFAULT 1,

    -- Usage tracking
                          usage_count INTEGER DEFAULT 0,
                          last_used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_personas_type ON personas(persona_type);
CREATE INDEX idx_personas_active ON personas(is_active) WHERE is_active = true;
CREATE INDEX idx_personas_industry ON personas USING gin(industry_focus);
CREATE INDEX idx_personas_content_types ON personas USING gin(content_types);

-- ============================================================================
-- SPECIALIZED AGENTS: Sub-agents for different aspects of persona
-- ============================================================================
CREATE TABLE specialized_agents (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    persona_id UUID NOT NULL REFERENCES personas(id) ON DELETE CASCADE,

                                    agent_type TEXT NOT NULL, -- 'knowledge', 'belief', 'cultural', 'psychological', 'style'

    -- Agent configuration (JSONB for flexibility)
                                    config JSONB NOT NULL,
    /*
    Varies by agent_type:

    For 'knowledge' agent:
    {
        "domains": {
            "domain_name": {
                "expertise_level": 0.9,
                "known_topics": [...],
                "knowledge_gaps": [...],
                "citation_style": "...",
                "bias_factors": {...}
            }
        }
    }

    For 'belief' agent:
    {
        "philosophical_outlook": "...",
        "political_leanings": "...",
        "values_hierarchy": [...],
        "ethical_framework": "...",
        "worldview_assumptions": [...]
    }

    For 'cultural' agent:
    {
        "literary_influences": [...],
        "quotations": [...],
        "metaphors": [...],
        "cultural_identity": [...]
    }

    For 'psychological' agent:
    {
        "cognitive_biases": [...],
        "defense_mechanisms": [...],
        "emotional_triggers": [...],
        "social_orientation": "...",
        "decision_making": "..."
    }

    For 'style' agent:
    {
        "vocabulary_level": "...",
        "sentence_complexity": 0.8,
        "formality": 0.7,
        "perspective": "first-person"
    }
    */

    -- Style-specific details (only for style agents)
                                    style_details JSONB,
    /*
    {
        "style_type": "academic",
        "style_subtype": "scientific",
        "word_origins": ["latin", "greek"],
        "preferred_words": [...],
        "avoided_words": [...],
        "rhetorical_devices": [...],
        "special_instructions": [...]
    }
    */

                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_specialized_agents_persona ON specialized_agents(persona_id);
CREATE INDEX idx_specialized_agents_type ON specialized_agents(agent_type);
CREATE UNIQUE INDEX idx_specialized_agents_persona_type ON specialized_agents(persona_id, agent_type);

-- ============================================================================
-- PERSONA ASSIGNMENTS: Track which personas are used where
-- ============================================================================
CREATE TABLE persona_assignments (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Assignment target
                                     orchestration_id UUID NOT NULL,
                                     flow_id UUID REFERENCES site_flows(id) ON DELETE CASCADE,
                                     page_id UUID REFERENCES flow_pages(id) ON DELETE CASCADE,

    -- Can assign to flow stage, page, or section
                                     assignment_level TEXT NOT NULL, -- 'flow_stage', 'page', 'section'
                                     stage_name TEXT, -- if assignment_level = 'flow_stage'

    -- Assigned persona
                                     persona_id UUID NOT NULL REFERENCES personas(id),
                                     persona_name TEXT NOT NULL, -- denormalized for convenience

    -- Assignment metadata
                                     assigned_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                     assigned_by TEXT, -- 'system', 'strategist_agent', 'manual'
                                     assignment_reason TEXT, -- why this persona was chosen

    -- Performance tracking
                                     content_generated INTEGER DEFAULT 0,
                                     avg_quality_score DECIMAL,

                                     UNIQUE(flow_id, stage_name, persona_id) -- one persona per stage
);

CREATE INDEX idx_persona_assignments_orchestration ON persona_assignments(orchestration_id);
CREATE INDEX idx_persona_assignments_flow ON persona_assignments(flow_id);
CREATE INDEX idx_persona_assignments_page ON persona_assignments(page_id);
CREATE INDEX idx_persona_assignments_persona ON persona_assignments(persona_id);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Get persona for a specific page (checks page, then stage, then default)
CREATE OR REPLACE FUNCTION get_persona_for_page(p_page_id UUID)
RETURNS TABLE (
    persona_id UUID,
    persona_name TEXT,
    persona_config JSONB,
    agent_configs JSONB
) AS $$
DECLARE
v_flow_id UUID;
    v_stage TEXT;
    v_persona_id UUID;
    v_persona_name TEXT;
    v_persona_config JSONB;
    v_agents JSONB;
BEGIN
    -- Get flow and stage for this page
SELECT flow_id, stage_in_narrative
INTO v_flow_id, v_stage
FROM flow_pages
WHERE id = p_page_id;

-- Try page-level assignment first
SELECT pa.persona_id, pa.persona_name
INTO v_persona_id, v_persona_name
FROM persona_assignments pa
WHERE pa.page_id = p_page_id
  AND pa.assignment_level = 'page'
    LIMIT 1;

-- If no page assignment, try stage-level
IF v_persona_id IS NULL THEN
SELECT pa.persona_id, pa.persona_name
INTO v_persona_id, v_persona_name
FROM persona_assignments pa
WHERE pa.flow_id = v_flow_id
  AND pa.stage_name = v_stage
  AND pa.assignment_level = 'flow_stage'
    LIMIT 1;
END IF;

    -- If still no assignment, use default (Elena Martinez - general B2B)
    IF v_persona_id IS NULL THEN
SELECT p.id, p.name
INTO v_persona_id, v_persona_name
FROM personas p
WHERE p.name = 'Elena Martinez'
  AND p.is_active = true
    LIMIT 1;
END IF;

    -- Get persona config
SELECT p.config
INTO v_persona_config
FROM personas p
WHERE p.id = v_persona_id;

-- Get all specialized agents for this persona
SELECT jsonb_object_agg(sa.agent_type, sa.config)
INTO v_agents
FROM specialized_agents sa
WHERE sa.persona_id = v_persona_id;

RETURN QUERY
SELECT v_persona_id, v_persona_name, v_persona_config, v_agents;
END;
$$ LANGUAGE plpgsql;

-- Assign persona to flow stage
CREATE OR REPLACE FUNCTION assign_persona_to_stage(
    p_flow_id UUID,
    p_stage_name TEXT,
    p_persona_name TEXT,
    p_reason TEXT DEFAULT 'system_assignment'
)
RETURNS UUID AS $$
DECLARE
v_persona_id UUID;
    v_orchestration_id UUID;
    v_assignment_id UUID;
BEGIN
    -- Get persona ID
SELECT id INTO v_persona_id
FROM personas
WHERE name = p_persona_name
  AND is_active = true;

IF v_persona_id IS NULL THEN
        RAISE EXCEPTION 'Persona % not found or inactive', p_persona_name;
END IF;

    -- Get orchestration ID from flow
SELECT orchestration_id INTO v_orchestration_id
FROM site_flows
WHERE id = p_flow_id;

-- Insert or update assignment
INSERT INTO persona_assignments (
    orchestration_id,
    flow_id,
    assignment_level,
    stage_name,
    persona_id,
    persona_name,
    assigned_by,
    assignment_reason
)
VALUES (
           v_orchestration_id,
           p_flow_id,
           'flow_stage',
           p_stage_name,
           v_persona_id,
           p_persona_name,
           'system',
           p_reason
       )
    ON CONFLICT (flow_id, stage_name, persona_id)
    DO UPDATE SET
    assigned_at = now(),
               assignment_reason = p_reason
               RETURNING id INTO v_assignment_id;

-- Update persona usage
UPDATE personas
SET usage_count = usage_count + 1,
    last_used_at = now()
WHERE id = v_persona_id;

RETURN v_assignment_id;
END;
$$ LANGUAGE plpgsql;

-- Get all personas for a flow (by stage)
CREATE OR REPLACE FUNCTION get_flow_personas(p_flow_id UUID)
RETURNS TABLE (
    stage_name TEXT,
    persona_name TEXT,
    persona_type TEXT,
    voice_parameters JSONB
) AS $$
BEGIN
RETURN QUERY
SELECT
    pa.stage_name,
    p.name,
    p.persona_type,
    p.config->'voice_parameters'
FROM persona_assignments pa
         JOIN personas p ON pa.persona_id = p.id
WHERE pa.flow_id = p_flow_id
  AND pa.assignment_level = 'flow_stage'
ORDER BY pa.stage_name;
END;
$$ LANGUAGE plpgsql;

-- Track persona performance
CREATE OR REPLACE FUNCTION record_persona_usage(
    p_assignment_id UUID,
    p_quality_score DECIMAL DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
UPDATE persona_assignments
SET
    content_generated = content_generated + 1,
    avg_quality_score = CASE
                            WHEN p_quality_score IS NOT NULL THEN
                                COALESCE(
                                        (avg_quality_score * content_generated + p_quality_score) / (content_generated + 1),
                                        p_quality_score
                                )
                            ELSE avg_quality_score
        END
WHERE id = p_assignment_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View: Active copywriter personas with their specializations
CREATE OR REPLACE VIEW v_active_copywriters AS
SELECT
    p.id,
    p.name,
    p.description,
    p.persona_type,
    p.industry_focus,
    p.content_types,
    p.config->'voice_parameters' as voice_parameters,
    p.config->'expertise' as expertise,
    p.usage_count,
    p.last_used_at
FROM personas p
WHERE p.is_active = true
  AND p.persona_type IN ('copywriter', 'technical_writer', 'thought_leader', 'conversion_specialist')
ORDER BY p.usage_count DESC;

-- View: Persona assignments with full context
CREATE OR REPLACE VIEW v_persona_assignments_detail AS
SELECT
    pa.id as assignment_id,
    pa.orchestration_id,
    pa.flow_id,
    sf.flow_name,
    sf.domain,
    pa.stage_name,
    pa.assignment_level,
    p.name as persona_name,
    p.persona_type,
    p.config->'voice_parameters' as voice_parameters,
    pa.assigned_at,
    pa.content_generated,
    pa.avg_quality_score
FROM persona_assignments pa
         JOIN personas p ON pa.persona_id = p.id
         JOIN site_flows sf ON pa.flow_id = sf.id
ORDER BY pa.assigned_at DESC;

-- ============================================================================
-- COMMENTS
-- ============================================================================
COMMENT ON TABLE personas IS 'Main persona definitions for copywriters and specialists';
COMMENT ON TABLE specialized_agents IS 'Sub-agents representing different aspects of a persona (knowledge, beliefs, style, etc)';
COMMENT ON TABLE persona_assignments IS 'Tracks which personas are assigned to which flows/stages/pages';
COMMENT ON COLUMN personas.config IS 'Full persona configuration including biographical, psychological, expertise, and voice parameters';
COMMENT ON COLUMN specialized_agents.agent_type IS 'Type of specialized agent: knowledge, belief, cultural, psychological, or style';
COMMENT ON COLUMN specialized_agents.style_details IS 'Additional style configuration for style agents';
COMMENT ON COLUMN persona_assignments.assignment_level IS 'Level of assignment: flow_stage, page, or section';
COMMENT ON FUNCTION get_persona_for_page IS 'Gets assigned persona for a page, falling back to stage then default';
COMMENT ON FUNCTION assign_persona_to_stage IS 'Assigns a persona to a flow stage';
COMMENT ON FUNCTION get_flow_personas IS 'Gets all persona assignments for a flow';